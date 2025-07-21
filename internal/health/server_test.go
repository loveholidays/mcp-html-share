package health

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type HealthServerTestSuite struct {
	suite.Suite
	server *Server
	logger *slog.Logger
}

func TestHealthServerSuite(t *testing.T) {
	suite.Run(t, new(HealthServerTestSuite))
}

func (s *HealthServerTestSuite) SetupSuite() {
	s.logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create a server with a known port for testing
	s.server = New("9091", s.logger)
	s.server.Start()

	// Give the server time to start
	time.Sleep(200 * time.Millisecond)
}

func (s *HealthServerTestSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		s.T().Logf("Error shutting down server: %v", err)
	}
}

func (s *HealthServerTestSuite) TestHealthEndpoint() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:9091/livez", http.NoBody)
	s.NoError(err)
	resp, err := http.DefaultClient.Do(req)
	s.NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	s.NoError(err)
	s.Equal("ok", string(body))
}

func (s *HealthServerTestSuite) TestMetricsEndpoint() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:9091/metrics", http.NoBody)
	s.NoError(err)
	resp, err := http.DefaultClient.Do(req)
	s.NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	s.NoError(err)
	s.Contains(string(body), "# HELP")
}
