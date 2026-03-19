# CapaFabric — Claude Code Guide

## Repository layout

```
CapaFabric/
├── shared/          # Go module: shared/models, errors, interfaces, resilience, manifest
├── control-plane/   # Go module: centralized registry, policy, health, config distribution
├── proxy/           # Go module: per-pod sidecar; capability or agent mode
├── sdk/             # Language SDKs (.NET, Go, TypeScript, Python, Java)
├── examples/        # End-to-end examples (agent-dotnet, capability-dotnet-oracle, ...)
├── docs/            # Architecture docs and ADRs (docs/adr/)
├── contracts/       # OpenAPI + JSON Schema contracts (language-neutral)
├── config/          # LiteLLM and shared infra configs
└── Makefile         # All common tasks (build, test, lint, run, docker)
```

Go workspace: `go.work` links `shared`, `control-plane`, and `proxy` as a single workspace.

Module paths:
- `github.com/sivasankaran83/CapaFabric/shared`
- `github.com/sivasankaran83/CapaFabric/control-plane`
- `github.com/sivasankaran83/CapaFabric/proxy`

Both `control-plane/go.mod` and `proxy/go.mod` carry a `replace` directive for `shared`:
```
replace github.com/sivasankaran83/CapaFabric/shared => ../shared
```
This is required for `go mod tidy` — the workspace alone is not enough.

## Build and run

```bash
# Build both binaries (no UI)
make build

# Run control plane (dev config, in-memory everything)
make run-cp

# Run proxy — agent mode
make run-proxy-goal

# Run proxy — capability mode (set MANIFEST=path/to/manifest.yaml)
make run-proxy-cap MANIFEST=examples/capability-dotnet-oracle/manifest.yaml

# Test all modules
make test

# Lint (go vet) all modules
make lint
```

Individual module builds:
```bash
cd control-plane && go build ./...
cd proxy        && go build ./...
cd shared       && go build ./...
```

## Go conventions (must follow)

### Error handling
- All handlers implement `AppHandler`: `func(w http.ResponseWriter, r *http.Request) error`
- Handlers **never** call `w.WriteHeader` or `w.Write` on error paths — they return an error
- The `Wrap` middleware (in `api/middleware.go` for both CP and proxy) is the **single** place that translates `AppError` → HTTP status
- Use typed `ErrorCode` constants from `shared/errors` (`ErrNotFound`, `ErrValidation`, etc.)
- `capaerrors.MapToHTTPStatus(code)` maps codes to HTTP status — do not hardcode status codes in handlers
- Wrap errors with context: `capaerrors.Wrap(capaerrors.ErrTransportError, "message", cause)`
- Use `capaerrors.NotFound("entity", "id")` for 404s

### Logging
- Use `log/slog` throughout — no `fmt.Println` or `log.Printf`
- Inject `*slog.Logger` via constructor; narrow with `.With("handler", "name")` per component
- Entry points (`main.go`) set `.With("service", "...", "version", "...")` on the root logger

### Concurrency
- All long-running goroutines receive a `context.Context` and must exit when it is cancelled
- `main.go` creates a root context via `signal.NotifyContext` (SIGINT/SIGTERM)
- Track goroutines with `sync.WaitGroup`; wait before exiting

### Resilience
- `shared/resilience.CircuitBreaker` — use for all outbound calls to external services
- `shared/resilience.Retry` with `DefaultRetryConfig()` — wraps transient calls; business errors (`ErrNotFound`, `ErrValidation`, etc.) are not retried
- The `transport.Manager` in `proxy/internal/transport` maintains per-capability circuit breakers; use `Manager.Adapter()` — do not call `NewTransportAdapter` directly from handlers

### Interfaces and DI
- All cross-package dependencies go through interfaces defined in `shared/interfaces/`
- Constructor injection only — no DI container, no globals
- Factory pattern: `NewXxx(config, logger) → Interface`

## Key ports

| Component            | Port |
|----------------------|------|
| Control Plane HTTP   | 8080 |
| Proxy (agent mode)   | 3500 |
| Proxy (capability)   | 3501 |
| App (backing service)| 8081 |
| LiteLLM              | 4000 |

## API surface

All routes are prefixed `/api/v1/`.

**Control Plane**
- `POST /api/v1/capabilities/register`
- `DELETE /api/v1/capabilities/{id}`
- `GET  /api/v1/capabilities`
- `POST /api/v1/discover`
- `POST /api/v1/invoke/{id}`
- `POST /api/v1/heartbeat`
- `GET  /api/v1/health`
- `GET  /api/v1/health/capabilities`

**Proxy**
- `POST /api/v1/discover`
- `POST /api/v1/invoke/{id}`
- `GET  /api/v1/health`
- `GET  /api/v1/health/app`

## Architecture decisions

ADRs live in `docs/adr/`. Key ones:

| ADR | Decision |
|-----|----------|
| 001 | Capability/Goal vocabulary — not "tool/task" |
| 002 | Control Plane + Thin Proxy split |
| 003 | Language-neutral contracts (OpenAPI, Protobuf, JSON Schema) |
| 004 | Pluggable cognitive provider |
| 005 | YAML manifest as primary integration contract |
| 006 | Zero hardcoded workflow — LLM reasons over capabilities |
| 007 | Governance opt-in via YAML (`auth`, `guardrails`, `policy`) |
| 008 | Agents are capabilities (multi-agent via composition) |
| 009 | Agent base class pattern for SDK |
| 010 | Loop protection via call-chain headers |

## Implementation phases

| Week | Scope |
|------|-------|
| 1 | `shared` models/interfaces/errors/manifest, CP registry + API |
| 2 | Proxy config/manifest/transport/registration/cache/API |
| 3 | .NET SDK + dotnet examples |
| 4 | CP policy (auth + enforcer) + config distribution + proxy cache |
| 5 | CP + proxy guardrails + pipeline + LLM provider |
| 6 | CP + proxy context, OTEL tracing, audit log |
| 7 | CP state/Redis, health monitor, load balancer, proxy fallback |
| 8 | CP admin UI, multi-agent example, docker-compose, CI |

## Common mistakes to avoid

- Do not use `github.com/psiog/capafabric` — the module path is `github.com/sivasankaran83/CapaFabric`
- Do not call `capaerrors.NotFound(id)` — the signature is `NotFound(entity, id string)`
- Do not skip the `replace` directive when running `go mod tidy`
- Do not add string error codes (`capaerrors.ErrCodeXxx`) where typed `ErrorCode` constants (`capaerrors.ErrXxx`) are expected
- Do not create a new `transport.Manager` per request — it must be shared so circuit breaker state accumulates
