package cache

import (
	"sync"
	"time"

	"github.com/sivasankaran83/CapaFabric/shared/models"
)

// ConfigSnapshot is the cached view of Control Plane state used by the proxy
// when operating in agent mode.
type ConfigSnapshot struct {
	Capabilities []models.CapabilityMetadata
	FetchedAt    time.Time
}

// ConfigCache holds a TTL-refreshed snapshot of CP configuration.
// When the CP is unreachable, the proxy serves from the last good snapshot.
type ConfigCache struct {
	mu       sync.RWMutex
	snapshot *ConfigSnapshot
	ttl      time.Duration
}

// NewConfigCache creates an empty cache with the given TTL.
func NewConfigCache(ttl time.Duration) *ConfigCache {
	return &ConfigCache{ttl: ttl}
}

// Set stores a new snapshot, replacing any previous value.
func (c *ConfigCache) Set(snap ConfigSnapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()
	snap.FetchedAt = time.Now()
	c.snapshot = &snap
}

// Get returns the current snapshot and whether it is still within TTL.
// Returns (nil, false) if the cache has never been populated.
func (c *ConfigCache) Get() (*ConfigSnapshot, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.snapshot == nil {
		return nil, false
	}
	fresh := time.Since(c.snapshot.FetchedAt) < c.ttl
	return c.snapshot, fresh
}

// Capabilities returns the cached capability list regardless of TTL staleness.
// This is the fallback path when the CP is unreachable.
func (c *ConfigCache) Capabilities() []models.CapabilityMetadata {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.snapshot == nil {
		return nil
	}
	return c.snapshot.Capabilities
}
