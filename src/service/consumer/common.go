package consumer

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

const (
	minDelayMinutes  = 1
	maxDelayMinutes  = 5
	customTimeFormat = "20060102_150405"
)

// handleExecutionError is a helper function to handle errors consistently across all executors
// It logs the error, adds span events, records error to span, and returns a wrapped error
func handleExecutionError(span trace.Span, logEntry *logrus.Entry, message string, err error) error {
	logEntry.Errorf("%s: %v", message, err)
	span.AddEvent(message)
	span.RecordError(err)
	return fmt.Errorf("%s: %w", message, err)
}
