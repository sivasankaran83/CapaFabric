package handlers

import (
	"log/slog"
	"net/http"

	"github.com/sivasankaran83/CapaFabric/control-plane/internal/api/dto"
	capaerrors "github.com/sivasankaran83/CapaFabric/shared/errors"
	"github.com/sivasankaran83/CapaFabric/shared/interfaces"
	"github.com/sivasankaran83/CapaFabric/shared/models"
)

// HealthHandler handles GET /api/v1/health and GET /api/v1/health/capabilities.
type HealthHandler struct {
	registry interfaces.CapabilityRegistry
	version  string
	logger   *slog.Logger
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(registry interfaces.CapabilityRegistry, version string, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		registry: registry,
		version:  version,
		logger:   logger.With("handler", "health"),
	}
}

// Health handles GET /api/v1/health.
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) error {
	return writeJSON(w, http.StatusOK, dto.HealthResponse{
		Status:  "ok",
		Version: h.version,
		Checks:  map[string]string{"registry": "ok"},
	})
}

// Capabilities handles GET /api/v1/health/capabilities.
func (h *HealthHandler) Capabilities(w http.ResponseWriter, r *http.Request) error {
	caps, err := h.registry.List(r.Context(), "")
	if err != nil {
		return capaerrors.Internal("registry.List failed", err)
	}

	counts := map[string]int{
		"healthy":   0,
		"unhealthy": 0,
		"unknown":   0,
		"total":     len(caps),
	}
	for _, cap := range caps {
		switch cap.Status {
		case models.CapabilityStatusHealthy:
			counts["healthy"]++
		case models.CapabilityStatusUnhealthy:
			counts["unhealthy"]++
		default:
			counts["unknown"]++
		}
	}

	return writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"counts": counts,
	})
}
