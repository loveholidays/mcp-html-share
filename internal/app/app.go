package app

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/loveholidays/mcp-html-share/internal/config"
	"github.com/loveholidays/mcp-html-share/internal/health"
	"github.com/loveholidays/mcp-html-share/internal/server"
)

// App coordinates the lifecycle of all application components.
type App struct {
	config       *config.Config
	logger       *slog.Logger
	healthServer *health.Server
	mcpServer    *server.Server
}

// New creates a new application instance.
func New(cfg *config.Config, logger *slog.Logger) *App {
	return &App{
		config:       cfg,
		logger:       logger,
		healthServer: health.New(cfg.HealthPort, logger),
		mcpServer:    server.New(cfg.BucketName, cfg.PublicURL, logger),
	}
}

// Run starts all application components and handles graceful shutdown.
func (a *App) Run() error {
	// Set up signal handling for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Start health server
	a.healthServer.Start()

	// Set up shutdown handler
	go a.handleShutdown(ctx)

	// Start MCP server (this blocks until shutdown)
	a.logger.Info("Starting MCP server", "transport", a.config.Transport)
	if err := a.mcpServer.Start(ctx, a.config.Transport, a.config.HTTPPort); err != nil {
		a.logger.Error("MCP server failed", "error", err)
		return err
	}

	return nil
}

// handleShutdown manages graceful shutdown of all components.
func (a *App) handleShutdown(ctx context.Context) {
	<-ctx.Done()
	a.logger.Info("Shutting down application...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	// Shutdown health server
	if err := a.healthServer.Shutdown(shutdownCtx); err != nil {
		a.logger.Error("Error shutting down health server", "error", err)
	}

	// Shutdown MCP server
	a.mcpServer.Shutdown()

	a.logger.Info("Application shutdown complete")
}
