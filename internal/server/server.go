package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/loveholidays/mcp-html-share/internal/gcs"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	uploadsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "html_uploads_total",
			Help: "Total number of HTML uploads",
		},
		[]string{"status"},
	)

	uploadDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "html_upload_duration_seconds",
			Help: "Duration of HTML uploads",
		},
		[]string{"status"},
	)
)

type Server struct {
	gcsService gcs.Service
	logger     *slog.Logger
	server     *server.MCPServer
}

func New(bucketName string, publicURL bool, logger *slog.Logger) *Server {
	gcsService, err := gcs.NewService(bucketName, publicURL)
	if err != nil {
		logger.Error("Failed to create GCS service", "error", err)
		os.Exit(1)
	}

	return &Server{
		gcsService: gcsService,
		logger:     logger,
	}
}

func (s *Server) Start(ctx context.Context, transport, httpPort string) error {
	s.setupMCPServer()

	switch transport {
	case "stdio":
		return s.startStdioServer(ctx)
	case "http":
		return s.startHTTPServer(ctx, httpPort)
	default:
		return fmt.Errorf("unsupported transport: %s", transport)
	}
}

// setupMCPServer initializes the MCP server and registers tools.
func (s *Server) setupMCPServer() {
	s.server = server.NewMCPServer("mcp-html-share", "1.0.0")
	s.server.AddTool(s.createShareHTMLTool(), s.handleShareHTML)
}

// createShareHTMLTool creates the share-html tool configuration.
func (s *Server) createShareHTMLTool() mcp.Tool {
	return mcp.Tool{
		Name: "share-html",
		Description: `Upload HTML content to GCS and return a public URL. 
Note: When using external JavaScript libraries via CDN:
- Prefer well-established CDN providers (cdnjs, jsdelivr, unpkg)
- Some libraries may not work properly when loaded via CDN (e.g., Recharts requires specific module loading)
- Recommended libraries that work reliably via CDN: Chart.js, D3.js, Plotly, jQuery, Bootstrap
- Always use the full UMD or standalone builds for libraries
- Test library availability by checking if the library object exists in window scope
- For React-based visualizations, consider using vanilla JavaScript or libraries designed for CDN usage`,
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"html_content": map[string]any{
					"type":        "string",
					"description": "The HTML content to upload",
				},
				"short_name": map[string]any{
					"type":        "string",
					"description": "A short name describing the content",
				},
			},
			Required: []string{"html_content", "short_name"},
		},
	}
}

// startStdioServer starts the server with stdio transport.
func (s *Server) startStdioServer(ctx context.Context) error {
	s.logger.Info("Starting MCP server with stdio transport")

	errChan := make(chan error, 1)
	go func() {
		errChan <- server.ServeStdio(s.server)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		s.logger.Info("Context cancelled, shutting down stdio server")
		return ctx.Err()
	}
}

// startHTTPServer starts the server with HTTP transport.
func (s *Server) startHTTPServer(ctx context.Context, httpPort string) error {
	s.logger.Info("Starting MCP server with HTTP transport", "port", httpPort)

	httpServer := server.NewStreamableHTTPServer(s.server)
	httpSrv := &http.Server{
		Addr:              ":" + httpPort,
		Handler:           httpServer,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errChan := make(chan error, 1)
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return s.shutdownHTTPServer(ctx, httpSrv)
	}
}

// shutdownHTTPServer gracefully shuts down the HTTP server.
func (s *Server) shutdownHTTPServer(ctx context.Context, httpSrv *http.Server) error {
	s.logger.Info("Context cancelled, shutting down HTTP server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(shutdownCtx); err != nil { //nolint:contextcheck // We need a fresh context for shutdown since original is cancelled
		s.logger.Error("Error during HTTP server shutdown", "error", err)
	}

	return ctx.Err()
}

func (s *Server) Shutdown() {
	s.logger.Info("MCP server shutting down")
}

func (s *Server) handleShareHTML(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	timer := prometheus.NewTimer(uploadDuration.WithLabelValues("unknown"))
	defer timer.ObserveDuration()

	// Extract parameters
	htmlContent, err := request.RequireString("html_content")
	if err != nil {
		uploadsTotal.WithLabelValues("error").Inc()
		return nil, fmt.Errorf("html_content parameter is required: %w", err)
	}

	shortName, err := request.RequireString("short_name")
	if err != nil {
		uploadsTotal.WithLabelValues("error").Inc()
		return nil, fmt.Errorf("short_name parameter is required: %w", err)
	}

	// Validate inputs
	if htmlContent == "" {
		uploadsTotal.WithLabelValues("error").Inc()
		return nil, fmt.Errorf("html_content cannot be empty")
	}

	if shortName == "" {
		uploadsTotal.WithLabelValues("error").Inc()
		return nil, fmt.Errorf("short_name cannot be empty")
	}

	s.logger.Info("Uploading HTML content", "short_name", shortName, "content_length", len(htmlContent))

	// Upload to GCS
	url, err := s.gcsService.UploadHTML(ctx, htmlContent, shortName)
	if err != nil {
		uploadsTotal.WithLabelValues("error").Inc()
		s.logger.Error("Failed to upload HTML content", "error", err, "short_name", shortName)
		return nil, fmt.Errorf("failed to upload HTML content: %w", err)
	}

	uploadsTotal.WithLabelValues("success").Inc()
	s.logger.Info("Successfully uploaded HTML content", "url", url, "short_name", shortName)

	// Return the URL
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(fmt.Sprintf("HTML content uploaded successfully! You can access it at: %s", url)),
		},
	}, nil
}
