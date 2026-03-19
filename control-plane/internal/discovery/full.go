package discovery

import (
	"context"
	"strings"

	"github.com/sivasankaran83/CapaFabric/shared/interfaces"
	"github.com/sivasankaran83/CapaFabric/shared/models"
)

// FullDiscoveryProvider returns all healthy capabilities, with optional tag filtering.
type FullDiscoveryProvider struct {
	registry interfaces.CapabilityRegistry
}

// NewFull creates a FullDiscoveryProvider backed by the given registry.
func NewFull(registry interfaces.CapabilityRegistry) *FullDiscoveryProvider {
	return &FullDiscoveryProvider{registry: registry}
}

func (p *FullDiscoveryProvider) Discover(ctx context.Context, req interfaces.DiscoveryRequest) ([]models.CapabilityMetadata, error) {
	all, err := p.registry.List(ctx, req.TenantID)
	if err != nil {
		return nil, err
	}

	var out []models.CapabilityMetadata
	for _, cap := range all {
		if cap.Status == models.CapabilityStatusUnhealthy {
			continue
		}
		if len(req.Tags) > 0 && !hasAnyTag(cap.Tags, req.Tags) {
			continue
		}
		out = append(out, cap)
	}

	if req.MaxTools > 0 && len(out) > req.MaxTools {
		out = out[:req.MaxTools]
	}
	return out, nil
}

// hasAnyTag returns true if capTags contains at least one tag from filter.
func hasAnyTag(capTags, filter []string) bool {
	set := make(map[string]struct{}, len(capTags))
	for _, t := range capTags {
		set[strings.ToLower(t)] = struct{}{}
	}
	for _, f := range filter {
		if _, ok := set[strings.ToLower(f)]; ok {
			return true
		}
	}
	return false
}
