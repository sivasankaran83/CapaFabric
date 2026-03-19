package interfaces

import (
	"context"

	"github.com/sivasankaran83/CapaFabric/shared/models"
)

// Span represents an active trace span.
type Span interface {
	// End closes the span, optionally recording an error.
	End(err error)

	// SetAttribute records a key-value attribute on the span.
	SetAttribute(key string, value any)

	// Context returns a context that carries the span for propagation.
	Context() context.Context
}

// InvocationTracer creates and manages OTEL-compatible spans.
type InvocationTracer interface {
	// Start begins a new span for a capability invocation.
	Start(ctx context.Context, ictx models.InvocationContext) (Span, context.Context)
}
