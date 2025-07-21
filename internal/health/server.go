package health

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server provides health and metrics endpoints.
type Server struct {
	server *http.Server
	logger *slog.Logger
}

// New creates a new health server.
func New(port string, logger *slog.Logger) *Server {
	mux := setupRoutes()

	return &Server{
		server: &http.Server{
			Addr:              ":" + port,
			Handler:           mux,
			ReadHeaderTimeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// Start begins serving health and metrics endpoints.
func (s *Server) Start() {
	go func() {
		s.logger.Info("Starting health server", "port", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Health server failed", "error", err)
		}
	}()
}

// Shutdown gracefully stops the health server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down health server")
	return s.server.Shutdown(ctx)
}

// setupRoutes configures the health and metrics endpoints.
func setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/livez", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil {
			// Log error but don't fail the health check
			return
		}
	})

	mux.Handle("/metrics", promhttp.Handler())

	return mux
}
