package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/sivasankaran83/CapaFabric/control-plane/internal/api/dto"
	capaerrors "github.com/sivasankaran83/CapaFabric/shared/errors"
	"github.com/sivasankaran83/CapaFabric/shared/interfaces"
	"github.com/sivasankaran83/CapaFabric/shared/models"
)

// CapabilitiesHandler handles capability registration and listing.
type CapabilitiesHandler struct {
	registry interfaces.CapabilityRegistry
	logger   *slog.Logger
}

// NewCapabilitiesHandler creates a new CapabilitiesHandler.
func NewCapabilitiesHandler(registry interfaces.CapabilityRegistry, logger *slog.Logger) *CapabilitiesHandler {
	return &CapabilitiesHandler{
		registry: registry,
		logger:   logger.With("handler", "capabilities"),
	}
}

// Register handles POST /api/v1/capabilities/register.
func (h *CapabilitiesHandler) Register(w http.ResponseWriter, r *http.Request) error {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return capaerrors.Wrap(capaerrors.ErrValidation, "invalid request body", err)
	}
	if req.CapabilityID == "" || req.AgentID == "" || req.Description == "" {
		return capaerrors.New(capaerrors.ErrValidation,
			"capability_id, agent_id, and description are required")
	}

	cap := models.CapabilityMetadata{
		CapabilityID:     req.CapabilityID,
		Name:             req.Name,
		Description:      req.Description,
		Tags:             req.Tags,
		Language:         req.Language,
		AgentID:          req.AgentID,
		TenantID:         req.TenantID,
		Source:           models.CapabilitySourceManifest,
		Idempotent:       req.Idempotent,
		SideEffects:      req.SideEffects,
		RequiresApproval: req.RequiresApproval,
		Endpoint:         req.Endpoint,
		Security:         req.Security,
		Transport:        req.Transport,
		ProxyURL:         req.ProxyURL,
		AppPort:          req.AppPort,
		Status:           models.CapabilityStatusUnknown,
		RegisteredAt:     time.Now().UTC(),
	}

	if err := h.registry.Register(r.Context(), cap); err != nil {
		return capaerrors.Internal("registry.Register failed", err)
	}

	h.logger.Info("capability registered", "capability_id", cap.CapabilityID, "agent_id", cap.AgentID)
	return writeJSON(w, http.StatusCreated, dto.RegisterResponse{
		CapabilityID: cap.CapabilityID,
		Status:       "registered",
	})
}

// Unregister handles DELETE /api/v1/capabilities/{id}.
func (h *CapabilitiesHandler) Unregister(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	if id == "" {
		return capaerrors.New(capaerrors.ErrValidation, "capability id is required")
	}
	if err := h.registry.Unregister(r.Context(), id); err != nil {
		return err // registry already returns AppError
	}
	h.logger.Info("capability unregistered", "capability_id", id)
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// List handles GET /api/v1/capabilities.
func (h *CapabilitiesHandler) List(w http.ResponseWriter, r *http.Request) error {
	tenantID := r.URL.Query().Get("tenant_id")
	caps, err := h.registry.List(r.Context(), tenantID)
	if err != nil {
		return capaerrors.Internal("registry.List failed", err)
	}
	if caps == nil {
		caps = []models.CapabilityMetadata{}
	}
	return writeJSON(w, http.StatusOK, dto.CapabilityListResponse{
		Capabilities: caps,
		Count:        len(caps),
	})
}
