package registry

import (
	"context"
	"sync"
	"time"

	capaerrors "github.com/sivasankaran83/CapaFabric/shared/errors"
	"github.com/sivasankaran83/CapaFabric/shared/models"
)

// InMemoryCapabilityRegistry is a thread-safe in-process capability registry.
// Suitable for development and single-instance deployments.
type InMemoryCapabilityRegistry struct {
	mu           sync.RWMutex
	capabilities map[string]models.CapabilityMetadata
}

// NewInMemory returns an empty InMemoryCapabilityRegistry.
func NewInMemory() *InMemoryCapabilityRegistry {
	return &InMemoryCapabilityRegistry{
		capabilities: make(map[string]models.CapabilityMetadata),
	}
}

func (r *InMemoryCapabilityRegistry) Register(_ context.Context, cap models.CapabilityMetadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if cap.RegisteredAt.IsZero() {
		cap.RegisteredAt = time.Now().UTC()
	}
	if cap.Status == "" {
		cap.Status = models.CapabilityStatusUnknown
	}
	r.capabilities[cap.CapabilityID] = cap
	return nil
}

func (r *InMemoryCapabilityRegistry) Unregister(_ context.Context, capabilityID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.capabilities[capabilityID]; !ok {
		return capaerrors.NotFound("capability", capabilityID)
	}
	delete(r.capabilities, capabilityID)
	return nil
}

func (r *InMemoryCapabilityRegistry) Get(_ context.Context, capabilityID string) (models.CapabilityMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cap, ok := r.capabilities[capabilityID]
	if !ok {
		return models.CapabilityMetadata{}, capaerrors.NotFound("capability", capabilityID)
	}
	return cap, nil
}

func (r *InMemoryCapabilityRegistry) List(_ context.Context, tenantID string) ([]models.CapabilityMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]models.CapabilityMetadata, 0, len(r.capabilities))
	for _, cap := range r.capabilities {
		if tenantID == "" || cap.TenantID == tenantID {
			out = append(out, cap)
		}
	}
	return out, nil
}

func (r *InMemoryCapabilityRegistry) UpdateStatus(_ context.Context, capabilityID string, status models.CapabilityStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cap, ok := r.capabilities[capabilityID]
	if !ok {
		return capaerrors.NotFound("capability", capabilityID)
	}
	cap.Status = status
	r.capabilities[capabilityID] = cap
	return nil
}

func (r *InMemoryCapabilityRegistry) UpdateHeartbeat(_ context.Context, capabilityID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cap, ok := r.capabilities[capabilityID]
	if !ok {
		return capaerrors.NotFound("capability", capabilityID)
	}
	now := time.Now().UTC()
	cap.LastHeartbeat = &now
	cap.Status = models.CapabilityStatusHealthy
	r.capabilities[capabilityID] = cap
	return nil
}
