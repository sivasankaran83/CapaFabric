package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	capaerrors "github.com/sivasankaran83/CapaFabric/shared/errors"
	"github.com/sivasankaran83/CapaFabric/shared/models"
)

// CapabilitySource is the minimal interface the discover handler needs.
type CapabilitySource interface {
	Capabilities() []models.CapabilityMetadata
}

// DiscoverHandler handles POST /api/v1/discover.
type DiscoverHandler struct {
	source CapabilitySource
	logger *slog.Logger
}

// NewDiscoverHandler creates a DiscoverHandler backed by source.
func NewDiscoverHandler(source CapabilitySource, logger *slog.Logger) *DiscoverHandler {
	return &DiscoverHandler{
		source: source,
		logger: logger.With("handler", "discover"),
	}
}

func (h *DiscoverHandler) Discover(w http.ResponseWriter, r *http.Request) error {
	var req struct {
		Goal     string   `json:"goal"`
		Tags     []string `json:"tags"`
		TenantID string   `json:"tenant_id"`
		Provider string   `json:"provider"`
		MaxTools int      `json:"max_tools"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return capaerrors.Wrap(capaerrors.ErrValidation, "invalid request body", err)
	}

	all := h.source.Capabilities()
	var filtered []models.CapabilityMetadata
	for _, cap := range all {
		if cap.Status == models.CapabilityStatusUnhealthy {
			continue
		}
		if req.TenantID != "" && cap.TenantID != "" && cap.TenantID != req.TenantID {
			continue
		}
		if len(req.Tags) > 0 && !hasAnyTag(cap.Tags, req.Tags) {
			continue
		}
		filtered = append(filtered, cap)
	}

	if req.MaxTools > 0 && len(filtered) > req.MaxTools {
		filtered = filtered[:req.MaxTools]
	}
	if filtered == nil {
		filtered = []models.CapabilityMetadata{}
	}

	h.logger.Debug("discover", "count", len(filtered), "goal", req.Goal)
	return writeJSON(w, http.StatusOK, map[string]any{
		"capabilities": filtered,
		"count":        len(filtered),
	})
}

func hasAnyTag(capTags, filter []string) bool {
	set := make(map[string]struct{}, len(capTags))
	for _, t := range capTags {
		set[strings.ToLower(t)] = struct{}{}
	}
	for _, f := range filter {
		if _, ok := set[strings.ToLower(f)]; ok {
			return true
		}
	}
	return false
}
