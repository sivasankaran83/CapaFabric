package api

import (
	"log/slog"
	"net/http"

	"github.com/sivasankaran83/CapaFabric/control-plane/internal/api/handlers"
	"github.com/sivasankaran83/CapaFabric/shared/interfaces"
)

// RouterConfig holds dependencies needed to build the control plane HTTP router.
type RouterConfig struct {
	Registry  interfaces.CapabilityRegistry
	Discovery interfaces.DiscoveryProvider
	Version   string
	Logger    *slog.Logger
}

// NewRouter builds and returns the control plane HTTP handler.
func NewRouter(cfg RouterConfig) http.Handler {
	log := cfg.Logger

	caps    := handlers.NewCapabilitiesHandler(cfg.Registry, log)
	disc    := handlers.NewDiscoverHandler(cfg.Discovery, log)
	inv     := handlers.NewInvokeHandler(cfg.Registry, log)
	hb      := handlers.NewHeartbeatHandler(cfg.Registry, log)
	health  := handlers.NewHealthHandler(cfg.Registry, cfg.Version, log)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/capabilities/register",    Wrap(caps.Register,    log))
	mux.HandleFunc("GET /api/v1/capabilities",              Wrap(caps.List,        log))
	mux.HandleFunc("DELETE /api/v1/capabilities/{id}",      Wrap(caps.Unregister,  log))
	mux.HandleFunc("POST /api/v1/discover",                 Wrap(disc.Discover,    log))
	mux.HandleFunc("POST /api/v1/invoke/{id}",              Wrap(inv.Invoke,       log))
	mux.HandleFunc("POST /api/v1/heartbeat",                Wrap(hb.Heartbeat,     log))
	mux.HandleFunc("GET /api/v1/health",                    Wrap(health.Health,    log))
	mux.HandleFunc("GET /api/v1/health/capabilities",       Wrap(health.Capabilities, log))

	return CORS(Recovery(log)(Logger(log)(RequestID(mux))))
}
