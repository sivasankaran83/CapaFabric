package interfaces

import (
	"context"

	"github.com/sivasankaran83/CapaFabric/shared/models"
)

// TransportAdapter executes the physical call to an application endpoint.
type TransportAdapter interface {
	// Invoke calls the capability's backing application and returns the result.
	Invoke(ctx context.Context, cap models.CapabilityMetadata, ictx models.InvocationContext) (models.InvocationResult, error)
}
