package transport

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/sivasankaran83/CapaFabric/shared/interfaces"
	"github.com/sivasankaran83/CapaFabric/shared/models"
	"github.com/sivasankaran83/CapaFabric/shared/resilience"
)

// Manager creates transport adapters and maintains per-capability circuit breakers.
// A single Manager instance should be shared across requests so that circuit
// breaker state accumulates correctly.
type Manager struct {
	breakers sync.Map // key: capabilityID → *resilience.CircuitBreaker
	logger   *slog.Logger
}

// NewManager returns a Manager with its own logger component.
func NewManager(logger *slog.Logger) *Manager {
	return &Manager{
		logger: logger.With("component", "transport-manager"),
	}
}

// Adapter returns a TransportAdapter for the given capability.
// The adapter is backed by a per-capability circuit breaker held in the Manager.
func (m *Manager) Adapter(cap models.CapabilityMetadata, appBaseURL string) (interfaces.TransportAdapter, error) {
	kind := cap.Transport.Kind
	if kind == "" {
		kind = models.TransportHTTP
	}
	switch kind {
	case models.TransportHTTP:
		cb := m.breakerFor(cap.CapabilityID)
		return newHTTP(appBaseURL, cb, m.logger), nil
	default:
		return nil, fmt.Errorf("unsupported transport kind %q for capability %s (supported: http)",
			kind, cap.CapabilityID)
	}
}

// breakerFor returns (or lazily creates) the circuit breaker for capabilityID.
func (m *Manager) breakerFor(capabilityID string) *resilience.CircuitBreaker {
	if v, ok := m.breakers.Load(capabilityID); ok {
		return v.(*resilience.CircuitBreaker)
	}
	cb := resilience.NewCircuitBreaker(
		capabilityID,
		5,              // open after 5 consecutive failures
		30*time.Second, // attempt recovery after 30 s
		m.logger,
	)
	actual, _ := m.breakers.LoadOrStore(capabilityID, cb)
	return actual.(*resilience.CircuitBreaker)
}

// NewTransportAdapter is a convenience wrapper for callers that have no Manager.
// It creates a one-shot adapter without a shared circuit breaker (suitable for tests).
func NewTransportAdapter(cap models.CapabilityMetadata, appBaseURL string) (interfaces.TransportAdapter, error) {
	kind := cap.Transport.Kind
	if kind == "" {
		kind = models.TransportHTTP
	}
	switch kind {
	case models.TransportHTTP:
		return newHTTP(appBaseURL, nil, slog.Default()), nil
	default:
		return nil, fmt.Errorf("unsupported transport kind %q for capability %s (supported: http)",
			kind, cap.CapabilityID)
	}
}
