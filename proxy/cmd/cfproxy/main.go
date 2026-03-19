package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	proxyapi "github.com/sivasankaran83/CapaFabric/proxy/internal/api"
	"github.com/sivasankaran83/CapaFabric/proxy/internal/cache"
	"github.com/sivasankaran83/CapaFabric/proxy/internal/config"
	proxymanifest "github.com/sivasankaran83/CapaFabric/proxy/internal/manifest"
	"github.com/sivasankaran83/CapaFabric/proxy/internal/registration"
	"github.com/sivasankaran83/CapaFabric/shared/models"
)

const version = "0.1.0"

func main() {
	configPath   := flag.String("config", "", "path to proxy config YAML")
	modeFlag     := flag.String("mode", "", "override mode: capability | agent")
	manifestFlag := flag.String("manifest", "", "path to capability manifest YAML (capability mode)")
	portFlag     := flag.Int("port", 0, "override listen port")
	devFlag      := flag.Bool("dev", false, "verbose logging")
	flag.Parse()

	// ── Logging ──
	level := slog.LevelInfo
	if *devFlag {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})).
		With("service", "cfproxy", "version", version)
	slog.SetDefault(logger)

	// ── Config ──
	var cfg *config.ProxyConfig
	var err error
	if *configPath != "" {
		cfg, err = config.Load(*configPath)
		if err != nil {
			logger.Error("failed to load config", "error", err)
			os.Exit(1)
		}
	} else {
		d := config.Defaults()
		cfg = &d
	}

	// CLI flags override YAML values.
	if *modeFlag != "" {
		cfg.Mode = config.ProxyMode(*modeFlag)
	}
	if *manifestFlag != "" {
		cfg.Manifest = *manifestFlag
	}
	if *portFlag > 0 {
		cfg.Port = *portFlag
	}

	if err := config.Validate(cfg); err != nil {
		logger.Error("invalid config", "error", err)
		os.Exit(1)
	}

	// Root context cancelled on shutdown signal.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var wg sync.WaitGroup
	var handler http.Handler

	switch cfg.Mode {

	// ──────────────────────────────────────────────────────────────
	// CAPABILITY MODE
	// Reads manifest → registers with CP → routes /invoke to app
	// ──────────────────────────────────────────────────────────────
	case config.ModeCapability:
		proxyURL := fmt.Sprintf("http://localhost:%d", cfg.Port)

		caps, err := proxymanifest.LoadCapabilities(cfg.Manifest, proxyURL)
		if err != nil {
			logger.Error("loading manifest", "error", err, "manifest", cfg.Manifest)
			os.Exit(1)
		}
		logger.Info("loaded capabilities from manifest",
			"count", len(caps),
			"manifest", cfg.Manifest,
		)

		appInfo, _ := proxymanifest.ManifestApp(cfg.Manifest)
		basePath := ""
		healthPath := "/health"
		if appInfo != nil {
			basePath = appInfo.BasePath
			if appInfo.HealthPath != "" {
				healthPath = appInfo.HealthPath
			}
		}
		appBaseURL := fmt.Sprintf("http://localhost:%d%s", cfg.App.Port, basePath)

		store := proxyapi.NewStaticStore(caps)
		handler = proxyapi.NewRouter(proxyapi.RouterConfig{
			Store:         store,
			AppBaseURL:    appBaseURL,
			AppPort:       cfg.App.Port,
			AppHealthPath: healthPath,
			Version:       version,
			Logger:        logger,
		})

		// Register with CP (retries with backoff until success or ctx cancelled).
		reg := registration.NewRegistrar(cfg.ControlPlane.URL, caps)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := reg.Register(ctx); err != nil {
				logger.Error("capability registration failed", "error", err)
			}
		}()

		// Periodic heartbeat.
		agentID := agentIDFromCaps(caps)
		hb := registration.NewHeartbeatSender(
			cfg.ControlPlane.URL,
			agentID,
			capabilityIDs(caps),
			30*time.Second,
		)
		wg.Add(1)
		go func() {
			defer wg.Done()
			hb.Start(ctx)
		}()

	// ──────────────────────────────────────────────────────────────
	// AGENT MODE
	// Caches CP config → serves discover + invoke for the agent
	// ──────────────────────────────────────────────────────────────
	case config.ModeAgent:
		ttl := time.Duration(cfg.ControlPlane.CacheTTLSeconds) * time.Second
		if ttl == 0 {
			ttl = 30 * time.Second
		}

		cfgCache := cache.NewConfigCache(ttl)
		refresher := cache.NewRefresher(cfg.ControlPlane.URL, cfgCache, ttl)
		wg.Add(1)
		go func() {
			defer wg.Done()
			refresher.Start(ctx)
		}()

		handler = proxyapi.NewRouter(proxyapi.RouterConfig{
			Store:   cfgCache,
			Version: version,
			Logger:  logger,
		})
	}

	// ── HTTP server ──
	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("cfproxy starting",
			"mode", cfg.Mode,
			"addr", addr,
			"cp_url", cfg.ControlPlane.URL,
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal.
	<-ctx.Done()

	logger.Info("shutting down cfproxy...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)

	wg.Wait()
	logger.Info("cfproxy stopped")
}

func agentIDFromCaps(caps []models.CapabilityMetadata) string {
	if len(caps) > 0 {
		return caps[0].AgentID
	}
	return "unknown"
}

func capabilityIDs(caps []models.CapabilityMetadata) []string {
	ids := make([]string, len(caps))
	for i, c := range caps {
		ids[i] = c.CapabilityID
	}
	return ids
}
