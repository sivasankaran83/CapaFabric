package interfaces

import (
	"context"

	"github.com/sivasankaran83/CapaFabric/shared/models"
)

// PolicyEnforcer evaluates whether an identity may invoke a capability.
type PolicyEnforcer interface {
	// Enforce returns a PolicyDecision for the given identity + capability pair.
	Enforce(ctx context.Context, identity models.AuthIdentity, cap models.CapabilityMetadata) (models.PolicyDecision, error)
}
