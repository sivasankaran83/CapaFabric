package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/sivasankaran83/CapaFabric/control-plane/internal/api/dto"
	capaerrors "github.com/sivasankaran83/CapaFabric/shared/errors"
	"github.com/sivasankaran83/CapaFabric/shared/interfaces"
	"github.com/sivasankaran83/CapaFabric/shared/models"
)

// DiscoverHandler handles POST /api/v1/discover.
type DiscoverHandler struct {
	discovery interfaces.DiscoveryProvider
	logger    *slog.Logger
}

// NewDiscoverHandler creates a new DiscoverHandler.
func NewDiscoverHandler(discovery interfaces.DiscoveryProvider, logger *slog.Logger) *DiscoverHandler {
	return &DiscoverHandler{
		discovery: discovery,
		logger:    logger.With("handler", "discover"),
	}
}

func (h *DiscoverHandler) Discover(w http.ResponseWriter, r *http.Request) error {
	var req dto.DiscoverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return capaerrors.Wrap(capaerrors.ErrValidation, "invalid request body", err)
	}

	caps, err := h.discovery.Discover(r.Context(), interfaces.DiscoveryRequest{
		Goal:     req.Goal,
		Tags:     req.Tags,
		TenantID: req.TenantID,
		Provider: req.Provider,
		MaxTools: req.MaxTools,
	})
	if err != nil {
		return capaerrors.Internal("discovery failed", err)
	}
	if caps == nil {
		caps = []models.CapabilityMetadata{}
	}

	h.logger.Debug("discovery completed", "count", len(caps), "goal", req.Goal)
	return writeJSON(w, http.StatusOK, dto.DiscoverResponse{
		Capabilities: caps,
		Count:        len(caps),
	})
}
