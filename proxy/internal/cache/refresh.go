package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/sivasankaran83/CapaFabric/shared/models"
)

// Refresher polls the Control Plane and updates the ConfigCache on each tick.
// On CP reconnect, it refreshes immediately. On disconnect, it logs a warning
// and the proxy continues serving from the stale cache.
type Refresher struct {
	cpURL    string
	cache    *ConfigCache
	interval time.Duration
	client   *http.Client
}

// NewRefresher creates a Refresher that polls cpURL every interval.
func NewRefresher(cpURL string, cache *ConfigCache, interval time.Duration) *Refresher {
	return &Refresher{
		cpURL:    cpURL,
		cache:    cache,
		interval: interval,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

// Start runs the polling loop until ctx is cancelled.
func (r *Refresher) Start(ctx context.Context) {
	// Attempt an immediate fetch before the first tick.
	if err := r.fetch(ctx); err != nil {
		slog.Warn("initial config fetch failed — will retry", "error", err)
	}

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.fetch(ctx); err != nil {
				slog.Warn("config refresh failed — serving from cache", "error", err)
			}
		}
	}
}

func (r *Refresher) fetch(ctx context.Context) error {
	url := r.cpURL + "/api/v1/capabilities"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("control plane returned HTTP %d", resp.StatusCode)
	}

	var envelope struct {
		Capabilities []models.CapabilityMetadata `json:"capabilities"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return fmt.Errorf("decoding capability list: %w", err)
	}

	r.cache.Set(ConfigSnapshot{Capabilities: envelope.Capabilities})
	slog.Debug("config cache refreshed", "capabilities", len(envelope.Capabilities))
	return nil
}
