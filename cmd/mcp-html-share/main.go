package main

import (
	"log/slog"
	"os"

	"github.com/loveholidays/mcp-html-share/internal/app"
	"github.com/loveholidays/mcp-html-share/internal/config"
)

func main() {
	// Load configuration from command line flags
	cfg, err := config.LoadFromFlags()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Set up structured logging
	logger := config.SetupLogger()

	// Create and run the application
	application := app.New(cfg, logger)
	if err := application.Run(); err != nil {
		logger.Error("Application failed", "error", err)
		os.Exit(1)
	}
}
