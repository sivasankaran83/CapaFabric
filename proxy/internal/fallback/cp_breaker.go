// Package fallback provides resilience primitives for proxy→control-plane connectivity.
// When the control-plane is unreachable the proxy serves capabilities from its stale
// cache so that agent invocations continue to work in a degraded-but-functional state.
package fallback

import (
	"log/slog"
	"time"

	"github.com/sivasankaran83/CapaFabric/shared/resilience"
)

const (
	// cpFailThreshold is the number of consecutive CP failures before the
	// circuit opens and the proxy stops attempting live refreshes.
	cpFailThreshold = 3

	// cpRecoveryTimeout is how long the circuit stays open before a probe
	// request is allowed through to test CP liveness.
	cpRecoveryTimeout = 60 * time.Second
)

// NewCPBreaker returns a CircuitBreaker configured for control-plane connectivity.
// Use it to wrap every outbound call from the proxy to the CP (register, heartbeat,
// config refresh). When the circuit is open, callers should serve from the stale
// ConfigCache instead of retrying the CP.
func NewCPBreaker(logger *slog.Logger) *resilience.CircuitBreaker {
	return resilience.NewCircuitBreaker(
		"control-plane",
		cpFailThreshold,
		cpRecoveryTimeout,
		logger,
	)
}
