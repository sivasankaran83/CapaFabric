package registration

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

// HeartbeatSender periodically POSTs heartbeats for registered capabilities
// to the Control Plane so the health monitor knows they are alive.
type HeartbeatSender struct {
	cpURL         string
	agentID       string
	capabilityIDs []string
	interval      time.Duration
	client        *http.Client
}

// NewHeartbeatSender creates a HeartbeatSender that sends beats on interval.
func NewHeartbeatSender(cpURL, agentID string, capabilityIDs []string, interval time.Duration) *HeartbeatSender {
	return &HeartbeatSender{
		cpURL:         cpURL,
		agentID:       agentID,
		capabilityIDs: capabilityIDs,
		interval:      interval,
		client:        &http.Client{Timeout: 5 * time.Second},
	}
}

// Start runs the heartbeat loop until ctx is cancelled.
func (h *HeartbeatSender) Start(ctx context.Context) {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	// Send an immediate heartbeat on startup.
	h.send(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.send(ctx)
		}
	}
}

func (h *HeartbeatSender) send(ctx context.Context) {
	payload := map[string]any{
		"agent_id":       h.agentID,
		"capability_ids": h.capabilityIDs,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		h.cpURL+"/api/v1/heartbeat", bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		slog.Warn("heartbeat failed", "error", err)
		return
	}
	resp.Body.Close()

	if resp.StatusCode >= 400 {
		slog.Warn("heartbeat rejected", "status", resp.StatusCode)
	}
}
