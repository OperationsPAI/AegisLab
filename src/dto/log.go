package dto

import (
	"time"

	"aegis/consts"
)

// LogEntry represents a unified log entry used by both OTLP receiver and Loki query
type LogEntry struct {
	Timestamp time.Time       `json:"timestamp"`          // Log timestamp
	Line      string          `json:"line"`               // Log content
	TaskID    string          `json:"task_id"`            // Associated task ID
	JobID     string          `json:"job_id,omitempty"`   // K8s Job name
	TraceID   string          `json:"trace_id,omitempty"` // Trace ID
	Level     consts.LogLevel `json:"level,omitempty"`    // Log level
}

// WSLogMessage represents the WebSocket message format for log streaming
type WSLogMessage struct {
	Type    consts.WSLogType `json:"type"`
	Logs    []LogEntry       `json:"logs,omitempty"`    // Log entries
	Message string           `json:"message,omitempty"` // Error message or end reason
	Total   int              `json:"total,omitempty"`   // Total history log count
}
