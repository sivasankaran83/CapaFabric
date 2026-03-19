package interfaces

import (
	"context"

	"github.com/sivasankaran83/CapaFabric/shared/models"
)

// CapabilityInvoker orchestrates the full invocation pipeline for a capability.
type CapabilityInvoker interface {
	// Invoke runs auth → policy → transport → audit for a single capability call.
	Invoke(ctx context.Context, capabilityID string, ictx models.InvocationContext) (models.InvocationResult, error)
}
