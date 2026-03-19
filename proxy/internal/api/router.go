package api

import (
	"log/slog"
	"net/http"

	"github.com/sivasankaran83/CapaFabric/proxy/internal/api/handlers"
	"github.com/sivasankaran83/CapaFabric/shared/models"
)

// CapabilityStore is the source of truth for capabilities the proxy knows about.
// In capability mode: a static slice from the manifest.
// In agent mode: the live ConfigCache.
type CapabilityStore interface {
	Capabilities() []models.CapabilityMetadata
}

// RouterConfig holds the dependencies needed to build the proxy HTTP router.
type RouterConfig struct {
	Store         CapabilityStore
	AppBaseURL    string // non-empty in capability mode; empty in agent mode
	AppPort       int    // for health/app probing; 0 in agent mode
	AppHealthPath string
	Version       string
	Logger        *slog.Logger
}

// NewRouter builds and returns the proxy HTTP handler.
func NewRouter(cfg RouterConfig) http.Handler {
	log := cfg.Logger

	disc := handlers.NewDiscoverHandler(cfg.Store, log)
	inv := handlers.NewInvokeHandler(cfg.Store, cfg.AppBaseURL, log)
	health := handlers.NewHealthHandler(cfg.Version, cfg.AppPort, cfg.AppHealthPath, log)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/discover", Wrap(disc.Discover, log))
	mux.HandleFunc("POST /api/v1/invoke/{id}", Wrap(inv.Invoke, log))
	mux.HandleFunc("GET /api/v1/health", Wrap(health.Health, log))
	mux.HandleFunc("GET /api/v1/health/app", Wrap(health.App, log))

	return Recovery(log)(Logger(log)(RequestID(mux)))
}
