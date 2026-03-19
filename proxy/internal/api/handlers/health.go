package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	capaerrors "github.com/sivasankaran83/CapaFabric/shared/errors"
)

// HealthHandler serves GET /api/v1/health and GET /api/v1/health/app.
type HealthHandler struct {
	version       string
	appPort       int
	appHealthPath string
	logger        *slog.Logger
}

// NewHealthHandler creates a HealthHandler.
// appPort == 0 signals agent mode (no backing app to probe).
func NewHealthHandler(version string, appPort int, appHealthPath string, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		version:       version,
		appPort:       appPort,
		appHealthPath: appHealthPath,
		logger:        logger.With("handler", "health"),
	}
}

// Health handles GET /api/v1/health — always 200 if the proxy is up.
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) error {
	return writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": h.version,
	})
}

// App handles GET /api/v1/health/app — probes the backing application.
func (h *HealthHandler) App(w http.ResponseWriter, r *http.Request) error {
	if h.appPort == 0 {
		return writeJSON(w, http.StatusOK, map[string]string{
			"status": "not_applicable",
			"reason": "agent mode — no backing application",
		})
	}

	healthURL := fmt.Sprintf("http://localhost:%d%s", h.appPort, h.appHealthPath)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(healthURL)
	if err != nil {
		return capaerrors.Wrap(capaerrors.ErrCapabilityUnhealthy,
			"backing application unreachable", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return capaerrors.New(capaerrors.ErrCapabilityUnhealthy,
			fmt.Sprintf("backing application returned HTTP %d", resp.StatusCode))
	}

	return writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}
