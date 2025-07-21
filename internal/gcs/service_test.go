package gcs

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type GCSServiceTestSuite struct {
	suite.Suite
}

func TestGCSServiceSuite(t *testing.T) {
	suite.Run(t, new(GCSServiceTestSuite))
}

func (s *GCSServiceTestSuite) TestSanitizeFilenameWithSimpleName() {
	result := sanitizeFilename("hello-world")
	s.Equal("hello-world", result)
	s.LessOrEqual(len(result), 50, "filename should not exceed 50 characters")
}

func (s *GCSServiceTestSuite) TestSanitizeFilenameWithSpaces() {
	result := sanitizeFilename("hello world")
	s.Equal("hello-world", result)
	s.LessOrEqual(len(result), 50, "filename should not exceed 50 characters")
}

func (s *GCSServiceTestSuite) TestSanitizeFilenameWithUnderscores() {
	result := sanitizeFilename("hello_world")
	s.Equal("hello-world", result)
	s.LessOrEqual(len(result), 50, "filename should not exceed 50 characters")
}

func (s *GCSServiceTestSuite) TestSanitizeFilenameWithSpecialCharacters() {
	result := sanitizeFilename("hello@world!#$%")
	s.Equal("helloworld", result)
	s.LessOrEqual(len(result), 50, "filename should not exceed 50 characters")
}

func (s *GCSServiceTestSuite) TestSanitizeFilenameWithMixedCase() {
	result := sanitizeFilename("Hello-World")
	s.Equal("hello-world", result)
	s.LessOrEqual(len(result), 50, "filename should not exceed 50 characters")
}

func (s *GCSServiceTestSuite) TestSanitizeFilenameWithEmptyString() {
	result := sanitizeFilename("")
	s.Equal("html-file", result)
	s.LessOrEqual(len(result), 50, "filename should not exceed 50 characters")
}

func (s *GCSServiceTestSuite) TestSanitizeFilenameWithOnlySpecialCharacters() {
	result := sanitizeFilename("@#$%^&*()")
	s.Equal("html-file", result)
	s.LessOrEqual(len(result), 50, "filename should not exceed 50 characters")
}

func (s *GCSServiceTestSuite) TestSanitizeFilenameWithLongString() {
	result := sanitizeFilename("this-is-a-very-long-filename-that-should-be-truncated-to-fifty-characters-max")
	s.Equal("this-is-a-very-long-filename-that-should-be-trunca", result)
	s.LessOrEqual(len(result), 50, "filename should not exceed 50 characters")
}

func (s *GCSServiceTestSuite) TestSanitizeFilenameWithAlphanumericAndHyphens() {
	result := sanitizeFilename("test-123-abc")
	s.Equal("test-123-abc", result)
	s.LessOrEqual(len(result), 50, "filename should not exceed 50 characters")
}
