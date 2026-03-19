# CapaFabric — Go Best Practices for Control Plane & Proxy

> Production-grade Go patterns. Covers: DI, middleware, error handling,
> circuit breaker, retry, structured logging, goroutines, repository pattern.

---

## 1. Dependency Injection (No Framework)

Constructor injection with interfaces. All wired in main.go.

```go
// Interface (small, focused)
type CapabilityRegistry interface {
    Register(ctx context.Context, cap CapabilityMetadata) error
    Unregister(ctx context.Context, capabilityID string) error
    GetByID(ctx context.Context, capabilityID string) (*CapabilityMetadata, error)
    GetAll(ctx context.Context) ([]CapabilityMetadata, error)
    Search(ctx context.Context, query string, limit int) ([]CapabilityMetadata, error)
    MarkUnhealthy(ctx context.Context, capabilityID string) error
    EvictStale(ctx context.Context, ttl time.Duration) (int, error)
    Close() error
}

// Handler accepts interface (never concrete type)
type CapabilityHandler struct {
    registry CapabilityRegistry
    logger   *slog.Logger
}

func NewCapabilityHandler(reg CapabilityRegistry, log *slog.Logger) *CapabilityHandler {
    return &CapabilityHandler{registry: reg, logger: log.With("handler", "capabilities")}
}

// main.go — composition root (ONLY place concrete types appear)
func main() {
    cfg := config.Load()
    logger := setupLogger(cfg)

    var reg registry.CapabilityRegistry
    switch cfg.Registry.Type {
    case "inmemory": reg = registry.NewInMemory()
    case "redis":    reg = registry.NewRedis(cfg.Registry.URL)
    case "postgres": reg = registry.NewPostgres(cfg.Registry.URL)
    case "sqlite":   reg = registry.NewSQLite(cfg.Registry.URL)
    }

    capHandler := handlers.NewCapabilityHandler(reg, logger)
    router := api.NewRouter(capHandler, logger)
    // ... start server
}
```

---

## 2. Centralized Error Handling

### Domain errors (not HTTP errors)

```go
// shared/errors/errors.go
type AppError struct {
    Code    ErrorCode `json:"code"`
    Message string    `json:"message"`
    Detail  string    `json:"detail,omitempty"`
    Cause   error     `json:"-"`
}

func (e *AppError) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }
func (e *AppError) Unwrap() error { return e.Cause }

type ErrorCode string
const (
    ErrNotFound          ErrorCode = "NOT_FOUND"
    ErrAlreadyExists     ErrorCode = "ALREADY_EXISTS"
    ErrValidation        ErrorCode = "VALIDATION_ERROR"
    ErrAuthentication    ErrorCode = "AUTHENTICATION_FAILED"
    ErrAuthorization     ErrorCode = "AUTHORIZATION_DENIED"
    ErrRateLimited       ErrorCode = "RATE_LIMITED"
    ErrCircularChain     ErrorCode = "CIRCULAR_CHAIN"
    ErrMaxDepthExceeded  ErrorCode = "MAX_DEPTH_EXCEEDED"
    ErrGuardrailBlocked  ErrorCode = "GUARDRAIL_BLOCKED"
    ErrCapabilityUnhealthy ErrorCode = "CAPABILITY_UNHEALTHY"
    ErrInternal          ErrorCode = "INTERNAL_ERROR"
)

func NotFound(entity, id string) *AppError {
    return &AppError{Code: ErrNotFound, Message: fmt.Sprintf("%s '%s' not found", entity, id)}
}
func Forbidden(reason string) *AppError {
    return &AppError{Code: ErrAuthorization, Message: reason}
}
func CircularChain(chain []string) *AppError {
    return &AppError{Code: ErrCircularChain, Message: "circular agent chain", Detail: strings.Join(chain, " -> ")}
}
func Internal(msg string, cause error) *AppError {
    return &AppError{Code: ErrInternal, Message: msg, Cause: cause}
}
```

### Error-to-HTTP middleware (SINGLE place for translation)

```go
// Handlers return errors — never write HTTP status directly
type AppHandler func(w http.ResponseWriter, r *http.Request) error

func Wrap(h AppHandler, logger *slog.Logger) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := h(w, r); err != nil {
            var appErr *cferrors.AppError
            if !errors.As(err, &appErr) {
                appErr = cferrors.Internal("unexpected error", err)
            }
            status := mapCodeToHTTPStatus(appErr.Code)
            logger.Error("request error", "code", appErr.Code, "status", status, "path", r.URL.Path)
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(status)
            json.NewEncoder(w).Encode(appErr)
        }
    }
}

func mapCodeToHTTPStatus(code cferrors.ErrorCode) int {
    switch code {
    case cferrors.ErrNotFound:            return 404
    case cferrors.ErrAlreadyExists:       return 409
    case cferrors.ErrValidation:          return 400
    case cferrors.ErrAuthentication:      return 401
    case cferrors.ErrAuthorization:       return 403
    case cferrors.ErrRateLimited:         return 429
    case cferrors.ErrCircularChain:       return 409
    case cferrors.ErrCapabilityUnhealthy: return 502
    default:                              return 500
    }
}

// Handler example — returns error, never writes status
func (h *CapabilityHandler) GetByID(w http.ResponseWriter, r *http.Request) error {
    id := chi.URLParam(r, "capabilityID")
    cap, err := h.registry.GetByID(r.Context(), id)
    if err != nil { return err }  // registry returns AppError
    return json.NewEncoder(w).Encode(cap)
}
```

---

## 3. Middleware Stack

```go
func NewRouter(capHandler *handlers.CapabilityHandler, logger *slog.Logger) http.Handler {
    r := chi.NewRouter()
    r.Use(middleware.RequestID)
    r.Use(middleware.Recoverer(logger))
    r.Use(middleware.StructuredLogger(logger))
    r.Use(middleware.Timeout(30 * time.Second))

    r.Post("/api/v1/capabilities/register", Wrap(capHandler.Register, logger))
    r.Get("/api/v1/capabilities/{capabilityID}", Wrap(capHandler.GetByID, logger))
    r.Post("/api/v1/discover", Wrap(capHandler.Discover, logger))
    r.Post("/api/v1/invoke/{capabilityID}", Wrap(capHandler.Invoke, logger))
    r.Get("/api/v1/health", Wrap(capHandler.Health, logger))
    return r
}

func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if rec := recover(); rec != nil {
                    logger.Error("panic", "panic", rec, "stack", string(debug.Stack()))
                    w.WriteHeader(500)
                }
            }()
            next.ServeHTTP(w, r)
        })
    }
}
```

---

## 4. Circuit Breaker

```go
type State int
const (
    Closed   State = iota  // normal
    Open                    // failing — reject immediately
    HalfOpen               // testing — one request allowed
)

type CircuitBreaker struct {
    mu               sync.RWMutex
    state            State
    failures         int
    successes        int
    failureThreshold int
    successThreshold int
    timeout          time.Duration
    lastFailure      time.Time
    logger           *slog.Logger
}

func NewCircuitBreaker(failThreshold int, timeout time.Duration, logger *slog.Logger) *CircuitBreaker {
    return &CircuitBreaker{failureThreshold: failThreshold, successThreshold: 2, timeout: timeout, logger: logger}
}

func (cb *CircuitBreaker) Do(ctx context.Context, fn func(ctx context.Context) error) error {
    if !cb.allowRequest() {
        return cferrors.New(cferrors.ErrCapabilityUnhealthy, "circuit breaker open")
    }
    err := fn(ctx)
    if err != nil { cb.recordFailure() } else { cb.recordSuccess() }
    return err
}

func (cb *CircuitBreaker) allowRequest() bool {
    cb.mu.RLock()
    defer cb.mu.RUnlock()
    if cb.state == Closed { return true }
    if cb.state == Open && time.Since(cb.lastFailure) > cb.timeout {
        cb.mu.RUnlock(); cb.mu.Lock()
        cb.state = HalfOpen
        cb.mu.Unlock(); cb.mu.RLock()
        return true
    }
    return cb.state == HalfOpen
}

func (cb *CircuitBreaker) recordFailure() {
    cb.mu.Lock(); defer cb.mu.Unlock()
    cb.failures++; cb.successes = 0; cb.lastFailure = time.Now()
    if cb.failures >= cb.failureThreshold {
        cb.state = Open
        cb.logger.Warn("circuit opened", "failures", cb.failures)
    }
}

func (cb *CircuitBreaker) recordSuccess() {
    cb.mu.Lock(); defer cb.mu.Unlock()
    cb.successes++; cb.failures = 0
    if cb.state == HalfOpen && cb.successes >= cb.successThreshold {
        cb.state = Closed
        cb.logger.Info("circuit closed")
    }
}
```

---

## 5. Retry with Exponential Backoff

```go
type RetryConfig struct {
    MaxAttempts  int
    InitialDelay time.Duration
    MaxDelay     time.Duration
    Multiplier   float64
    Retryable    func(error) bool
}

func DefaultRetry() RetryConfig {
    return RetryConfig{
        MaxAttempts: 3, InitialDelay: 100 * time.Millisecond,
        MaxDelay: 5 * time.Second, Multiplier: 2.0,
        Retryable: func(err error) bool {
            var appErr *cferrors.AppError
            if errors.As(err, &appErr) {
                switch appErr.Code {
                case cferrors.ErrNotFound, cferrors.ErrValidation,
                     cferrors.ErrAuthentication, cferrors.ErrAuthorization:
                    return false  // don't retry business errors
                }
            }
            return true  // retry transient errors
        },
    }
}

func Retry(ctx context.Context, cfg RetryConfig, fn func(ctx context.Context) error) error {
    delay := cfg.InitialDelay
    var lastErr error
    for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
        lastErr = fn(ctx)
        if lastErr == nil { return nil }
        if !cfg.Retryable(lastErr) { return lastErr }
        if attempt == cfg.MaxAttempts { break }
        jitter := time.Duration(rand.Int63n(int64(delay / 4)))
        select {
        case <-ctx.Done(): return ctx.Err()
        case <-time.After(delay + jitter):
        }
        delay = time.Duration(float64(delay) * cfg.Multiplier)
        if delay > cfg.MaxDelay { delay = cfg.MaxDelay }
    }
    return fmt.Errorf("max retries (%d): %w", cfg.MaxAttempts, lastErr)
}
```

---

## 6. Structured Logging (slog)

```go
func setupLogger(cfg *config.Config) *slog.Logger {
    var handler slog.Handler
    if cfg.Logging.Format == "json" {
        handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: parseLevel(cfg.Logging.Level)})
    } else {
        handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: parseLevel(cfg.Logging.Level)})
    }
    return slog.New(handler).With("service", cfg.Server.ServiceName)
}

// Usage: always structured
// logger.Info("capability registered", "capability_id", cap.ID, "source", cap.Source)
// logger.Error("invoke failed", "err", err, "capability_id", id, "duration_ms", d)
// logger.Warn("circuit opened", "endpoint", ep, "failures", n)
```

---

## 7. Goroutines for Background Jobs

```go
type Monitor struct {
    registry CapabilityRegistry
    logger   *slog.Logger
    cfg      HealthConfig
    cancel   context.CancelFunc
    wg       sync.WaitGroup
}

func (m *Monitor) Start(ctx context.Context) {
    ctx, m.cancel = context.WithCancel(ctx)
    m.wg.Add(1)
    go func() {
        defer m.wg.Done()
        ticker := time.NewTicker(m.cfg.CheckInterval)
        defer ticker.Stop()
        for {
            select {
            case <-ctx.Done(): return
            case <-ticker.C: m.checkAll(ctx)
            }
        }
    }()
}

func (m *Monitor) checkAll(ctx context.Context) {
    caps, _ := m.registry.GetAll(ctx)
    sem := make(chan struct{}, 10)  // bounded concurrency
    var wg sync.WaitGroup
    for _, cap := range caps {
        if cap.HealthEndpoint == "" { continue }
        wg.Add(1)
        sem <- struct{}{}
        go func(c CapabilityMetadata) {
            defer wg.Done()
            defer func() { <-sem }()
            // check health, mark unhealthy if needed
        }(cap)
    }
    wg.Wait()
}

func (m *Monitor) Stop() { m.cancel(); m.wg.Wait() }

// Graceful shutdown in main.go
func main() {
    // ... setup ...
    monitor.Start(context.Background())
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    monitor.Stop()
    server.Shutdown(context.Background())
}
```

---

## 8. Repository Pattern (Database-Agnostic)

```
CapabilityRegistry interface
    |
    +-- InMemory     (sync.RWMutex + map)       -- Level 0-3
    +-- SQLite       (database/sql + sqlite3)    -- Level 4 (embedded, 0 RAM)
    +-- Redis        (go-redis/redis)            -- Level 5 (production)
    +-- Postgres     (jackc/pgx)                 -- Level 5 (production)

StateStore interface
    |
    +-- InMemory     (sync.RWMutex + map)
    +-- SQLite       (database/sql)
    +-- Redis        (go-redis/redis)
    +-- Postgres     (jackc/pgx)

All implement the SAME interface. Swap via YAML config.
Zero code changes in handlers, middleware, or business logic.
```

---

## 9. Transport with Circuit Breaker + Retry

```go
type HTTPTransport struct {
    client   *http.Client
    breakers map[string]*CircuitBreaker
    mu       sync.RWMutex
    retry    RetryConfig
    logger   *slog.Logger
}

func (t *HTTPTransport) Invoke(ctx context.Context, meta *CapabilityMetadata, args json.RawMessage) (json.RawMessage, error) {
    breaker := t.getOrCreateBreaker(meta.Transport.Endpoint)
    var result json.RawMessage
    err := Retry(ctx, t.retry, func(ctx context.Context) error {
        return breaker.Do(ctx, func(ctx context.Context) error {
            var err error
            result, err = t.doHTTP(ctx, meta, args)
            return err
        })
    })
    return result, err
}
```

---

## Summary

```
Pattern              | Go Idiom                          | Location
---------------------+-----------------------------------+---------------------------
DI                   | Constructor injection in main.go  | cmd/capafabric/main.go
Errors               | Values, not exceptions            | shared/errors/
Error→HTTP           | Wrap middleware                    | internal/api/middleware/
Middleware           | Chi middleware chain               | internal/api/router.go
Circuit Breaker      | State machine + sync.RWMutex      | shared/resilience/
Retry                | Context-aware + jitter            | shared/resilience/
Logging              | slog with .With()                 | Throughout
Background Jobs      | ticker + select + WaitGroup       | internal/health/
Graceful Shutdown    | signal.Notify + WaitGroup         | main.go
Repository           | Interface + N implementations     | internal/registry/
DB Agnostic          | Switch in main.go                 | cmd/capafabric/main.go
```
