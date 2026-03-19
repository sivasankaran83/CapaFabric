package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/sivasankaran83/CapaFabric/control-plane/internal/api/dto"
	capaerrors "github.com/sivasankaran83/CapaFabric/shared/errors"
	"github.com/sivasankaran83/CapaFabric/shared/interfaces"
)

// HeartbeatHandler handles POST /api/v1/heartbeat.
type HeartbeatHandler struct {
	registry interfaces.CapabilityRegistry
	logger   *slog.Logger
}

// NewHeartbeatHandler creates a new HeartbeatHandler.
func NewHeartbeatHandler(registry interfaces.CapabilityRegistry, logger *slog.Logger) *HeartbeatHandler {
	return &HeartbeatHandler{
		registry: registry,
		logger:   logger.With("handler", "heartbeat"),
	}
}

func (h *HeartbeatHandler) Heartbeat(w http.ResponseWriter, r *http.Request) error {
	var req dto.HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return capaerrors.Wrap(capaerrors.ErrValidation, "invalid request body", err)
	}

	var failed []string
	for _, id := range req.CapabilityIDs {
		if err := h.registry.UpdateHeartbeat(r.Context(), id); err != nil {
			h.logger.Warn("heartbeat update failed", "capability_id", id, "error", err)
			failed = append(failed, id)
		}
	}

	if len(failed) > 0 {
		return writeJSON(w, http.StatusMultiStatus, map[string]any{
			"status": "partial",
			"failed": failed,
		})
	}

	h.logger.Debug("heartbeat received",
		"agent_id", req.AgentID,
		"capability_count", len(req.CapabilityIDs),
	)
	return writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
