package interfaces

import (
	"context"

	"github.com/sivasankaran83/CapaFabric/shared/models"
)

// DiscoveryRequest carries the search criteria for capability discovery.
type DiscoveryRequest struct {
	Goal     string   `json:"goal,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	TenantID string   `json:"tenant_id,omitempty"`
	Provider string   `json:"provider,omitempty"` // "openai" | "anthropic" | "raw"
	MaxTools int      `json:"max_tools,omitempty"`
}

// DiscoveryProvider finds capabilities relevant to a discovery request.
type DiscoveryProvider interface {
	// Discover returns capabilities matching the request criteria.
	Discover(ctx context.Context, req DiscoveryRequest) ([]models.CapabilityMetadata, error)
}
