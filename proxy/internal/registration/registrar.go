package registration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/sivasankaran83/CapaFabric/shared/models"
)

// Registrar registers capabilities with the Control Plane and unregisters them
// on graceful shutdown.
type Registrar struct {
	cpURL      string
	client     *http.Client
	caps       []models.CapabilityMetadata
}

// NewRegistrar creates a Registrar that will POST to cpURL/api/v1/capabilities/register.
func NewRegistrar(cpURL string, caps []models.CapabilityMetadata) *Registrar {
	return &Registrar{
		cpURL: cpURL,
		caps:  caps,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Register attempts to register all capabilities with the Control Plane.
// Retries with exponential backoff until ctx is cancelled.
func (r *Registrar) Register(ctx context.Context) error {
	backoff := 1 * time.Second
	const maxBackoff = 30 * time.Second

	for {
		err := r.registerAll(ctx)
		if err == nil {
			slog.Info("registered capabilities with control plane",
				"count", len(r.caps),
				"cp_url", r.cpURL,
			)
			return nil
		}

		slog.Warn("registration failed, retrying",
			"error", err,
			"retry_in", backoff,
		)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}

func (r *Registrar) registerAll(ctx context.Context) error {
	url := r.cpURL + "/api/v1/capabilities/register"
	for _, cap := range r.caps {
		if err := r.registerOne(ctx, url, cap); err != nil {
			return fmt.Errorf("registering %s: %w", cap.CapabilityID, err)
		}
	}
	return nil
}

func (r *Registrar) registerOne(ctx context.Context, url string, cap models.CapabilityMetadata) error {
	body, err := json.Marshal(cap)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("control plane returned HTTP %d", resp.StatusCode)
	}
	return nil
}

// Unregister removes all capabilities from the Control Plane.
// Best-effort — logs errors but does not return them.
func (r *Registrar) Unregister(ctx context.Context) {
	client := &http.Client{Timeout: 5 * time.Second}
	for _, cap := range r.caps {
		url := fmt.Sprintf("%s/api/v1/capabilities/%s", r.cpURL, cap.CapabilityID)
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
		if err != nil {
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			slog.Warn("failed to unregister capability", "id", cap.CapabilityID, "error", err)
			continue
		}
		resp.Body.Close()
		slog.Info("unregistered capability", "id", cap.CapabilityID)
	}
}

// CapabilityIDs returns the IDs of all capabilities managed by this registrar.
func (r *Registrar) CapabilityIDs() []string {
	ids := make([]string, len(r.caps))
	for i, c := range r.caps {
		ids[i] = c.CapabilityID
	}
	return ids
}
