package observability

import (
	"log/slog"
	"reflect"
	"time"

	"github.com/getsentry/sentry-go"
)

type Sentry struct {
	enabled bool
}

func Init(dsn, environment string, logger *slog.Logger) *Sentry {
	if dsn == "" {
		return &Sentry{}
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      environment,
		BeforeSend:       ScrubEvent,
		AttachStacktrace: false,
		SendDefaultPII:   false,
		EnableTracing:    false,
	})
	if err != nil {
		if logger != nil {
			logger.Error("Failed to initialize Sentry", "error", err)
		}
		return &Sentry{}
	}

	return &Sentry{enabled: true}
}

func (s *Sentry) CaptureError(err error) {
	if !s.enabled || err == nil {
		return
	}

	typeName := "error"
	if errorType := reflect.TypeOf(err); errorType != nil {
		typeName = errorType.String()
	}
	sentry.CaptureEvent(&sentry.Event{
		Level: sentry.LevelError,
		Exception: []sentry.Exception{{
			Type:  typeName,
			Value: "operation failed",
		}},
	})
}

func (s *Sentry) Close() {
	if s.enabled {
		sentry.Flush(2 * time.Second)
	}
}

func ScrubEvent(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
	eventID, level := event.EventID, event.Level
	return &sentry.Event{EventID: eventID, Level: level, Message: "application error"}
}
