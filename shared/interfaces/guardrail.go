package interfaces

import (
	"context"

	"github.com/sivasankaran83/CapaFabric/shared/models"
)

// GuardrailProvider checks content against one or more safety rules.
type GuardrailProvider interface {
	// CheckInbound evaluates content before it reaches the capability.
	CheckInbound(ctx context.Context, content string, ictx models.InvocationContext) ([]models.GuardrailResult, error)

	// CheckOutbound evaluates capability output before it is returned to the caller.
	CheckOutbound(ctx context.Context, content string, ictx models.InvocationContext) ([]models.GuardrailResult, error)
}
