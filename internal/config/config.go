package config

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
)

// Config holds all application configuration.
type Config struct {
	BucketName string
	Transport  string
	HTTPPort   string
	HealthPort string
	PublicURL  bool
}

// LoadFromFlags parses command line flags and returns configuration.
func LoadFromFlags() (*Config, error) {
	bucketName := flag.String("bucket", "", "GCS bucket name for HTML uploads (required)")
	transport := flag.String("transport", "stdio", "Transport mode: stdio or http")
	httpPort := flag.String("http-port", "8080", "HTTP port for MCP server")
	healthPort := flag.String("health-port", "9090", "Health and metrics port")
	publicURL := flag.Bool("public-url", true, "Return public URLs (true) or signed URLs (false)")

	flag.Parse()

	config := &Config{
		BucketName: *bucketName,
		Transport:  *transport,
		HTTPPort:   *httpPort,
		HealthPort: *healthPort,
		PublicURL:  *publicURL,
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate checks that all required configuration is present.
func (c *Config) Validate() error {
	if c.BucketName == "" {
		return fmt.Errorf("bucket flag is required")
	}

	if c.Transport != "stdio" && c.Transport != "http" {
		return fmt.Errorf("transport must be either 'stdio' or 'http'")
	}

	return nil
}

// SetupLogger configures the default structured logger.
func SetupLogger() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
	return logger
}
