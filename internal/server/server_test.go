package server

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/suite"
)

// mockGCSService implements the gcs.Service interface for testing.
type mockGCSService struct {
	uploadFunc func(ctx context.Context, content, shortName string) (string, error)
}

func (m *mockGCSService) UploadHTML(ctx context.Context, content, shortName string) (string, error) {
	if m.uploadFunc != nil {
		return m.uploadFunc(ctx, content, shortName)
	}
	return "https://storage.cloud.google.com/test-bucket/test-file-12345678.html", nil
}

type ServerTestSuite struct {
	suite.Suite
	logger *slog.Logger
}

func TestServerSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

func (s *ServerTestSuite) SetupTest() {
	s.logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func (s *ServerTestSuite) createRequest(args map[string]any) mcp.CallToolRequest {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args
	return req
}

// createServerWithMock creates a server with a mock GCS service.
func (s *ServerTestSuite) createServerWithMock(uploadFunc func(ctx context.Context, content, shortName string) (string, error)) *Server {
	mockGCS := &mockGCSService{
		uploadFunc: uploadFunc,
	}

	return &Server{
		gcsService: mockGCS,
		logger:     s.logger,
	}
}

// assertSuccessfulResponse asserts that the response is successful and contains the expected URL.
func (s *ServerTestSuite) assertSuccessfulResponse(result *mcp.CallToolResult, err error, expectedURL string) {
	s.NoError(err)
	s.NotNil(result)
	s.Len(result.Content, 1)

	textContent, ok := mcp.AsTextContent(result.Content[0])
	s.True(ok, "Expected text content")
	s.Contains(textContent.Text, expectedURL)
}

// assertErrorResponse asserts that the response contains an error.
func (s *ServerTestSuite) assertErrorResponse(result *mcp.CallToolResult, err error) {
	s.Error(err)
	s.Nil(result)
}

func (s *ServerTestSuite) TestHandleShareHTMLWithSuccessfulUpload() {
	server := s.createServerWithMock(func(_ context.Context, _, _ string) (string, error) {
		return "https://storage.cloud.google.com/test-bucket/test-page-12345678.html", nil
	})

	request := s.createRequest(map[string]any{
		"html_content": "<html><body><h1>Test</h1></body></html>",
		"short_name":   "test-page",
	})

	result, err := server.handleShareHTML(context.Background(), request)
	s.assertSuccessfulResponse(result, err, "https://storage.cloud.google.com/test-bucket/test-page-12345678.html")
}

func (s *ServerTestSuite) TestHandleShareHTMLWithMissingHTMLContent() {
	server := s.createServerWithMock(nil)

	request := s.createRequest(map[string]any{
		"short_name": "test-page",
	})

	result, err := server.handleShareHTML(context.Background(), request)
	s.assertErrorResponse(result, err)
}

func (s *ServerTestSuite) TestHandleShareHTMLWithMissingShortName() {
	server := s.createServerWithMock(nil)

	request := s.createRequest(map[string]any{
		"html_content": "<html><body><h1>Test</h1></body></html>",
	})

	result, err := server.handleShareHTML(context.Background(), request)
	s.assertErrorResponse(result, err)
}

func (s *ServerTestSuite) TestHandleShareHTMLWithEmptyHTMLContent() {
	server := s.createServerWithMock(nil)

	request := s.createRequest(map[string]any{
		"html_content": "",
		"short_name":   "test-page",
	})

	result, err := server.handleShareHTML(context.Background(), request)
	s.assertErrorResponse(result, err)
}

func (s *ServerTestSuite) TestHandleShareHTMLWithEmptyShortName() {
	server := s.createServerWithMock(nil)

	request := s.createRequest(map[string]any{
		"html_content": "<html><body><h1>Test</h1></body></html>",
		"short_name":   "",
	})

	result, err := server.handleShareHTML(context.Background(), request)
	s.assertErrorResponse(result, err)
}

func (s *ServerTestSuite) TestHandleShareHTMLWithInvalidHTMLContentType() {
	server := s.createServerWithMock(nil)

	request := s.createRequest(map[string]any{
		"html_content": 123,
		"short_name":   "test-page",
	})

	result, err := server.handleShareHTML(context.Background(), request)
	s.assertErrorResponse(result, err)
}

func (s *ServerTestSuite) TestHandleShareHTMLWithInvalidShortNameType() {
	server := s.createServerWithMock(nil)

	request := s.createRequest(map[string]any{
		"html_content": "<html><body><h1>Test</h1></body></html>",
		"short_name":   123,
	})

	result, err := server.handleShareHTML(context.Background(), request)
	s.assertErrorResponse(result, err)
}
