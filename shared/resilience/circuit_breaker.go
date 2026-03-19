package resilience

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	capaerrors "github.com/sivasankaran83/CapaFabric/shared/errors"
)

// State represents a circuit breaker state machine value.
type State int

const (
	StateClosed   State = iota // normal operation
	StateOpen                  // failing — reject immediately
	StateHalfOpen              // testing — one probe request allowed
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker protects a downstream dependency using the standard
// closed → open → half-open state machine.
type CircuitBreaker struct {
	mu               sync.Mutex
	state            State
	failures         int
	successes        int
	failureThreshold int
	successThreshold int
	timeout          time.Duration
	lastFailure      time.Time
	name             string
	logger           *slog.Logger
}

// NewCircuitBreaker creates a CircuitBreaker that opens after failThreshold
// consecutive failures and recovers after timeout.
func NewCircuitBreaker(name string, failThreshold int, timeout time.Duration, logger *slog.Logger) *CircuitBreaker {
	return &CircuitBreaker{
		name:             name,
		failureThreshold: failThreshold,
		successThreshold: 2,
		timeout:          timeout,
		logger:           logger,
	}
}

// Do executes fn if the circuit is closed or half-open.
// Returns ErrCapabilityUnhealthy immediately if the circuit is open.
func (cb *CircuitBreaker) Do(ctx context.Context, fn func(context.Context) error) error {
	if !cb.allow() {
		return capaerrors.New(capaerrors.ErrCapabilityUnhealthy,
			fmt.Sprintf("circuit breaker open for %q", cb.name))
	}
	err := fn(ctx)
	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
	return err
}

// State returns the current circuit state (for metrics/health endpoints).
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

func (cb *CircuitBreaker) allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case StateClosed:
		return true
	case StateHalfOpen:
		return true
	case StateOpen:
		if time.Since(cb.lastFailure) >= cb.timeout {
			cb.state = StateHalfOpen
			cb.logger.Info("circuit half-open", "name", cb.name)
			return true
		}
		return false
	default:
		return false
	}
}

func (cb *CircuitBreaker) onFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.successes = 0
	cb.lastFailure = time.Now()
	if cb.state == StateHalfOpen || cb.failures >= cb.failureThreshold {
		cb.state = StateOpen
		cb.logger.Warn("circuit opened", "name", cb.name, "failures", cb.failures)
	}
}

func (cb *CircuitBreaker) onSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.successes++
	cb.failures = 0
	if cb.state == StateHalfOpen && cb.successes >= cb.successThreshold {
		cb.state = StateClosed
		cb.logger.Info("circuit closed", "name", cb.name)
	}
}
