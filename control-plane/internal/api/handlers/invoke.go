package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/sivasankaran83/CapaFabric/control-plane/internal/api/dto"
	capaerrors "github.com/sivasankaran83/CapaFabric/shared/errors"
	"github.com/sivasankaran83/CapaFabric/shared/interfaces"
	"github.com/sivasankaran83/CapaFabric/shared/models"
)

// InvokeHandler handles POST /api/v1/invoke/{id}.
// In Week 1 the CP does not route traffic itself — it returns the proxy routing info.
// Full invocation pipeline lives in the proxy (Week 2+).
type InvokeHandler struct {
	registry interfaces.CapabilityRegistry
	logger   *slog.Logger
}

// NewInvokeHandler creates a new InvokeHandler.
func NewInvokeHandler(registry interfaces.CapabilityRegistry, logger *slog.Logger) *InvokeHandler {
	return &InvokeHandler{
		registry: registry,
		logger:   logger.With("handler", "invoke"),
	}
}

func (h *InvokeHandler) Invoke(w http.ResponseWriter, r *http.Request) error {
	capabilityID := r.PathValue("id")
	if capabilityID == "" {
		return capaerrors.New(capaerrors.ErrValidation, "capability_id is required")
	}

	var req dto.InvokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return capaerrors.Wrap(capaerrors.ErrValidation, "invalid request body", err)
	}

	cap, err := h.registry.Get(r.Context(), capabilityID)
	if err != nil {
		return err // already AppError
	}

	if cap.Status == models.CapabilityStatusUnhealthy {
		return capaerrors.New(capaerrors.ErrCapabilityUnhealthy,
			"capability is currently unhealthy, try again later")
	}

	// CP returns routing info; the proxy executes the actual invocation.
	result := models.InvocationResult{
		RequestID:    uuid.NewString(),
		CapabilityID: capabilityID,
		Success:      true,
		Result:       map[string]any{"proxy_url": cap.ProxyURL, "capability": cap},
		DurationMS:   time.Since(time.Now()).Milliseconds(),
	}

	h.logger.Debug("invoke routed", "capability_id", capabilityID, "proxy_url", cap.ProxyURL)
	return writeJSON(w, http.StatusOK, dto.InvokeResponse{InvocationResult: result})
}
