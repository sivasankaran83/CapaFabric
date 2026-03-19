package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sivasankaran83/CapaFabric/control-plane/internal/api"
	"github.com/sivasankaran83/CapaFabric/control-plane/internal/config"
	"github.com/sivasankaran83/CapaFabric/control-plane/internal/discovery"
	"github.com/sivasankaran83/CapaFabric/control-plane/internal/registry"
)

const version = "0.1.0"

func main() {
	configPath := flag.String("config", "configs/capafabric.dev.yaml", "path to config YAML file")
	port       := flag.Int("port", 0, "override server port (0 = use config value)")
	dev        := flag.Bool("dev", false, "enable development mode (verbose logging)")
	flag.Parse()

	// ── Structured logger ──
	logLevel := slog.LevelInfo
	if *dev {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})).
		With("service", "capafabric-cp", "version", version)
	slog.SetDefault(logger)

	// ── Config ──
	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("failed to load config", "error", err, "path", *configPath)
		os.Exit(1)
	}
	if *port > 0 {
		cfg.Server.Port = *port
	}
	if err := config.Validate(cfg); err != nil {
		logger.Error("invalid config", "error", err)
		os.Exit(1)
	}

	// ── Registry ──
	reg, err := registry.NewRegistry(cfg.Registry)
	if err != nil {
		logger.Error("failed to initialize registry", "error", err)
		os.Exit(1)
	}

	// ── Discovery ──
	disc := discovery.NewFull(reg)

	// ── HTTP router ──
	handler := api.NewRouter(api.RouterConfig{
		Registry:  reg,
		Discovery: disc,
		Version:   version,
		Logger:    logger,
	})

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// ── Start server ──
	go func() {
		logger.Info("control plane starting",
			"addr", addr,
			"registry", cfg.Registry.Type,
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// ── Graceful shutdown ──
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", "error", err)
	}
	logger.Info("stopped")
}
