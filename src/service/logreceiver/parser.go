package logreceiver

import (
	"encoding/json"
	"fmt"
	"time"

	"aegis/consts"
	"aegis/dto"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
)

// parseOTLPLogs parses an OTLP ExportLogsServiceRequest and extracts LogEntry items.
// It walks the three-level structure: ResourceLogs -> ScopeLogs -> LogRecords.
func parseOTLPLogs(req *collogspb.ExportLogsServiceRequest) []dto.LogEntry {
	var entries []dto.LogEntry

	for _, rl := range req.GetResourceLogs() {
		// Extract resource-level attributes (task_id, trace_id, job_id, etc.)
		resAttrs := kvListToMap(rl.GetResource().GetAttributes())

		for _, sl := range rl.GetScopeLogs() {
			for _, lr := range sl.GetLogRecords() {
				entry := parseLogRecord(lr, resAttrs)
				if entry.TaskID != "" {
					entries = append(entries, entry)
				}
			}
		}
	}

	return entries
}

// parseLogRecord converts a single OTLP LogRecord into a LogEntry.
func parseLogRecord(lr *logspb.LogRecord, resAttrs map[string]string) dto.LogEntry {
	// Merge resource attributes with log-level attributes (log-level takes precedence)
	logAttrs := kvListToMap(lr.GetAttributes())

	entry := dto.LogEntry{
		Timestamp: parseTimestamp(lr.GetTimeUnixNano(), lr.GetObservedTimeUnixNano()),
		Line:      extractBody(lr),
		TaskID:    firstNonEmpty(logAttrs["task_id"], resAttrs["task_id"]),
		TraceID:   firstNonEmpty(logAttrs["trace_id"], resAttrs["trace_id"]),
		JobID:     firstNonEmpty(logAttrs["job_id"], resAttrs["job_id"]),
		Level:     mapSeverity(lr.GetSeverityText(), lr.GetSeverityNumber()),
	}

	return entry
}

// parseTimestamp converts OTLP nanosecond timestamps to time.Time.
// Falls back to ObservedTimeUnixNano if TimeUnixNano is zero.
func parseTimestamp(timeNano, observedNano uint64) time.Time {
	ns := timeNano
	if ns == 0 {
		ns = observedNano
	}
	if ns == 0 {
		return time.Now()
	}
	return time.Unix(0, int64(ns))
}

// extractBody extracts the log line from the Body field.
func extractBody(lr *logspb.LogRecord) string {
	body := lr.GetBody()
	if body == nil {
		return ""
	}

	switch v := body.GetValue().(type) {
	case *commonpb.AnyValue_StringValue:
		return v.StringValue
	case *commonpb.AnyValue_IntValue:
		return fmt.Sprintf("%d", v.IntValue)
	case *commonpb.AnyValue_DoubleValue:
		return fmt.Sprintf("%f", v.DoubleValue)
	case *commonpb.AnyValue_BoolValue:
		return fmt.Sprintf("%t", v.BoolValue)
	case *commonpb.AnyValue_BytesValue:
		return string(v.BytesValue)
	default:
		// For complex types, marshal to JSON
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Sprintf("%v", body)
		}
		return string(data)
	}
}

// kvListToMap converts OTLP KeyValue list to a simple string map.
// Only string values are extracted; other types are converted to string representation.
func kvListToMap(attrs []*commonpb.KeyValue) map[string]string {
	m := make(map[string]string, len(attrs))
	for _, kv := range attrs {
		if kv == nil || kv.GetValue() == nil {
			continue
		}
		switch v := kv.GetValue().GetValue().(type) {
		case *commonpb.AnyValue_StringValue:
			m[kv.GetKey()] = v.StringValue
		case *commonpb.AnyValue_IntValue:
			m[kv.GetKey()] = fmt.Sprintf("%d", v.IntValue)
		case *commonpb.AnyValue_BoolValue:
			m[kv.GetKey()] = fmt.Sprintf("%t", v.BoolValue)
		case *commonpb.AnyValue_DoubleValue:
			m[kv.GetKey()] = fmt.Sprintf("%f", v.DoubleValue)
		}
	}
	return m
}

// mapSeverity maps OTLP severity to a simple level string.
func mapSeverity(severityText string, severityNumber logspb.SeverityNumber) consts.LogLevel {
	if severityText != "" {
		return consts.LogLevel(severityText)
	}

	switch {
	case severityNumber >= logspb.SeverityNumber_SEVERITY_NUMBER_ERROR:
		return consts.LogLevelError
	case severityNumber >= logspb.SeverityNumber_SEVERITY_NUMBER_WARN:
		return consts.LogLevelWarn
	case severityNumber >= logspb.SeverityNumber_SEVERITY_NUMBER_INFO:
		return consts.LogLevelInfo
	case severityNumber >= logspb.SeverityNumber_SEVERITY_NUMBER_DEBUG:
		return consts.LogLevelDebug
	default:
		return consts.LogLevelInfo
	}
}

// firstNonEmpty returns the first non-empty string from the given values.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
