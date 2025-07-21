package gcs

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

type Service interface {
	UploadHTML(ctx context.Context, content, shortName string) (string, error)
}

type service struct {
	client     *storage.Client
	bucketName string
	publicURL  bool
}

func NewService(bucketName string, publicURL bool) (Service, error) {
	client, err := storage.NewClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &service{
		client:     client,
		bucketName: bucketName,
		publicURL:  publicURL,
	}, nil
}

func (s *service) UploadHTML(ctx context.Context, content, shortName string) (string, error) {
	// Generate unique hash from content
	hash := sha256.Sum256([]byte(content))
	hashStr := fmt.Sprintf("%x", hash)[:8] // Use first 8 chars for brevity

	// Create filename with short name and hash
	filename := fmt.Sprintf("%s-%s.html", sanitizeFilename(shortName), hashStr)

	// Get bucket handle
	bucket := s.client.Bucket(s.bucketName)

	// Create object writer
	obj := bucket.Object(filename)
	writer := obj.NewWriter(ctx)

	// Set content type
	writer.ContentType = "text/html"
	writer.CacheControl = "public, max-age=3600"

	// Write content
	if _, err := io.WriteString(writer, content); err != nil {
		writer.Close()
		return "", fmt.Errorf("failed to write content to GCS: %w", err)
	}

	// Close writer to finalize upload
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close GCS writer: %w", err)
	}

	// Return appropriate URL based on configuration
	if s.publicURL {
		// Return public URL
		url := fmt.Sprintf("https://storage.cloud.google.com/%s/%s", s.bucketName, filename)
		return url, nil
	}

	// Generate signed URL (valid for 7 days)
	opts := &storage.SignedURLOptions{
		Method:  "GET",
		Expires: time.Now().Add(7 * 24 * time.Hour),
	}
	url, err := bucket.SignedURL(filename, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}
	return url, nil
}

// sanitizeFilename removes invalid characters from filename.
func sanitizeFilename(name string) string {
	// Replace spaces and special characters with hyphens
	name = normalizeSpacesAndSpecialChars(name)

	// Remove any characters that aren't alphanumeric or hyphen
	sanitized := filterValidCharacters(name)

	// Ensure we have at least something
	if sanitized == "" {
		sanitized = "html-file"
	}

	// Limit length and convert to lowercase
	return strings.ToLower(limitLength(sanitized, 50))
}

// normalizeSpacesAndSpecialChars replaces spaces and underscores with hyphens.
func normalizeSpacesAndSpecialChars(name string) string {
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	return name
}

// filterValidCharacters removes any characters that aren't alphanumeric or hyphen.
func filterValidCharacters(name string) string {
	var result strings.Builder
	for _, r := range name {
		if isValidCharacter(r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// isValidCharacter checks if a rune is alphanumeric or hyphen.
func isValidCharacter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-'
}

// limitLength truncates a string to the specified maximum length.
func limitLength(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen]
	}
	return s
}
