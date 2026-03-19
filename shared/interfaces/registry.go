package interfaces

import (
	"context"

	"github.com/sivasankaran83/CapaFabric/shared/models"
)

// CapabilityRegistry stores and retrieves registered capabilities.
type CapabilityRegistry interface {
	// Register adds or updates a capability. Idempotent by capability_id.
	Register(ctx context.Context, cap models.CapabilityMetadata) error

	// Unregister removes a capability by ID.
	Unregister(ctx context.Context, capabilityID string) error

	// Get retrieves a single capability by ID.
	Get(ctx context.Context, capabilityID string) (models.CapabilityMetadata, error)

	// List returns all registered capabilities, optionally filtered by tenantID.
	List(ctx context.Context, tenantID string) ([]models.CapabilityMetadata, error)

	// UpdateStatus updates the health status of a capability.
	UpdateStatus(ctx context.Context, capabilityID string, status models.CapabilityStatus) error

	// UpdateHeartbeat records a heartbeat timestamp for a capability.
	UpdateHeartbeat(ctx context.Context, capabilityID string) error
}
