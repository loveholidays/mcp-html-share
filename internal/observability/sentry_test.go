package observability_test

import (
	"errors"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/loveholidays/mcp-html-share/internal/observability"
	"github.com/stretchr/testify/suite"
)

type SentryTestSuite struct {
	suite.Suite
}

func TestSentrySuite(t *testing.T) {
	suite.Run(t, new(SentryTestSuite))
}

func (s *SentryTestSuite) TestInitWithoutDSNDisablesReporting() {
	reporter := observability.Init("", "test", nil)
	reporter.CaptureError(errors.New("secret body"))
	reporter.Close()
}

func (s *SentryTestSuite) TestScrubEventRemovesSensitiveFields() {
	event := &sentry.Event{
		Message:     "secret body",
		Environment: "production",
		Request:     &sentry.Request{URL: "https://example.test/private?token=secret", Data: "secret body"},
		User:        sentry.User{Email: "person@example.test", IPAddress: "192.0.2.1"},
		Contexts:    map[string]sentry.Context{"local": {"path": "/private"}},
		Breadcrumbs: []*sentry.Breadcrumb{{Message: "secret"}},
		Tags:        map[string]string{"path": "/private"},
		Exception:   []sentry.Exception{{Value: "secret body"}},
	}

	scrubbed := observability.ScrubEvent(event, nil)
	s.Equal(event.EventID, scrubbed.EventID)
	s.Equal(event.Level, scrubbed.Level)
	s.Equal("production", scrubbed.Environment)
	s.Equal("application error", scrubbed.Message)
	s.Nil(scrubbed.Request)
	s.True(scrubbed.User.IsEmpty())
	s.Nil(scrubbed.Contexts)
	s.Nil(scrubbed.Breadcrumbs)
	s.Nil(scrubbed.Tags)
	s.Empty(scrubbed.Exception)
}
