package logreceiver

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"aegis/client"
	"aegis/dto"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
)

const (
	// DefaultPort is the default OTLP receiver port
	DefaultPort = 4319

	// DefaultMaxRequestSize is the default maximum request body size (5MB)
	DefaultMaxRequestSize = 5 * 1024 * 1024

	// PubSubChannelPrefix is the Redis Pub/Sub channel prefix for job logs
	PubSubChannelPrefix = "joblogs"
)

// OTLPLogReceiver receives OTLP HTTP log data and publishes to Redis Pub/Sub
type OTLPLogReceiver struct {
	server         *http.Server
	port           int
	maxRequestSize int64
	shutdownCh     chan struct{}

	// Metrics
	receivedTotal  atomic.Int64
	publishedTotal atomic.Int64
	errorsTotal    atomic.Int64
}

// NewOTLPLogReceiver creates a new OTLP log receiver
func NewOTLPLogReceiver(port int, maxRequestSize int64) *OTLPLogReceiver {
	if port == 0 {
		port = DefaultPort
	}
	if maxRequestSize == 0 {
		maxRequestSize = DefaultMaxRequestSize
	}

	return &OTLPLogReceiver{
		port:           port,
		maxRequestSize: maxRequestSize,
		shutdownCh:     make(chan struct{}),
	}
}

// Start starts the OTLP HTTP log receiver
func (r *OTLPLogReceiver) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/logs", r.handleLogs)
	mux.HandleFunc("/health", r.handleHealth)

	r.server = &http.Server{
		Addr:              fmt.Sprintf(":%d", r.port),
		Handler:           mux,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
	}

	// Handle graceful shutdown
	go func() {
		select {
		case <-ctx.Done():
		case <-r.shutdownCh:
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := r.server.Shutdown(shutdownCtx); err != nil {
			logrus.Errorf("OTLP log receiver shutdown error: %v", err)
		}
	}()

	logrus.Infof("OTLP log receiver started on :%d", r.port)

	if err := r.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("OTLP log receiver failed: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the receiver
func (r *OTLPLogReceiver) Shutdown() {
	close(r.shutdownCh)
}

// handleHealth handles health check requests
func (r *OTLPLogReceiver) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":          "healthy",
		"received_total":  r.receivedTotal.Load(),
		"published_total": r.publishedTotal.Load(),
		"errors_total":    r.errorsTotal.Load(),
	})
}

// handleLogs handles OTLP log export requests
func (r *OTLPLogReceiver) handleLogs(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body with size limit
	body, err := io.ReadAll(io.LimitReader(req.Body, r.maxRequestSize))
	if err != nil {
		r.errorsTotal.Add(1)
		logrus.Errorf("OTLP receiver: failed to read request body: %v", err)
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	if len(body) == 0 {
		http.Error(w, "empty request body", http.StatusBadRequest)
		return
	}

	contentEncoding := req.Header.Get("Content-Encoding")
	switch contentEncoding {
	case "gzip":
		gzReader, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			r.errorsTotal.Add(1)
			logrus.Errorf("OTLP receiver: failed to create gzip reader: %v", err)
			http.Error(w, "failed to parse gzip body", http.StatusBadRequest)
			return
		}
		defer gzReader.Close()

		decompressedBody, err := io.ReadAll(gzReader)
		if err != nil {
			r.errorsTotal.Add(1)
			logrus.Errorf("OTLP receiver: failed to decompress gzip body: %v", err)
			http.Error(w, "failed to decompress gzip body", http.StatusBadRequest)
			return
		}
		body = decompressedBody

	default:
		// No compression, proceed with original body
	}

	var exportReq collogspb.ExportLogsServiceRequest

	contentType := req.Header.Get("Content-Type")
	switch contentType {
	case "application/x-protobuf", "application/protobuf":
		if err := proto.Unmarshal(body, &exportReq); err != nil {
			r.errorsTotal.Add(1)
			logrus.Errorf("OTLP receiver: failed to unmarshal protobuf: %v", err)
			http.Error(w, "failed to parse protobuf request", http.StatusBadRequest)
			return
		}

	case "application/json", "":
		// Parse JSON format - support both ExportLogsServiceRequest wrapper and direct ResourceLogs
		if err := r.parseJSONRequest(body, &exportReq); err != nil {
			r.errorsTotal.Add(1)
			logrus.Errorf("OTLP receiver: failed to parse JSON: %v", err)
			http.Error(w, "failed to parse JSON request", http.StatusBadRequest)
			return
		}

	default:
		http.Error(w, fmt.Sprintf("unsupported content type: %s", contentType), http.StatusUnsupportedMediaType)
		return
	}

	// Parse and publish logs
	entries := parseOTLPLogs(&exportReq)
	r.receivedTotal.Add(int64(len(entries)))

	// Publish each entry to Redis Pub/Sub by task_id
	ctx := req.Context()
	for _, entry := range entries {
		if err := r.publishLogEntry(ctx, entry); err != nil {
			r.errorsTotal.Add(1)
			logrus.Errorf("OTLP receiver: failed to publish log entry: %v", err)
		} else {
			r.publishedTotal.Add(1)
		}
	}

	// Return OTLP success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{}`))
}

// parseJSONRequest parses JSON body into ExportLogsServiceRequest.
// Supports the standard OTLP JSON format with "resourceLogs" field.
func (r *OTLPLogReceiver) parseJSONRequest(body []byte, exportReq *collogspb.ExportLogsServiceRequest) error {
	// Try parsing as standard OTLP JSON format
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	resourceLogsRaw, ok := raw["resourceLogs"]
	if !ok {
		return fmt.Errorf("missing 'resourceLogs' field")
	}

	var resourceLogs []json.RawMessage
	if err := json.Unmarshal(resourceLogsRaw, &resourceLogs); err != nil {
		return fmt.Errorf("invalid 'resourceLogs': %w", err)
	}

	for _, rlRaw := range resourceLogs {
		rl, err := parseResourceLog(rlRaw)
		if err != nil {
			logrus.Warnf("OTLP receiver: skipping malformed resource log: %v", err)
			continue
		}
		exportReq.ResourceLogs = append(exportReq.ResourceLogs, rl)
	}

	return nil
}

// publishLogEntry publishes a log entry to Redis Pub/Sub channel keyed by task_id
func (r *OTLPLogReceiver) publishLogEntry(ctx context.Context, entry dto.LogEntry) error {
	channel := fmt.Sprintf("%s:%s", PubSubChannelPrefix, entry.TaskID)
	return client.RedisPublish(ctx, channel, entry)
}

// parseResourceLog parses a single ResourceLog from JSON
func parseResourceLog(data json.RawMessage) (*logspb.ResourceLogs, error) {
	var raw struct {
		Resource struct {
			Attributes []struct {
				Key   string `json:"key"`
				Value struct {
					StringValue *string `json:"stringValue,omitempty"`
					IntValue    *string `json:"intValue,omitempty"`
					BoolValue   *bool   `json:"boolValue,omitempty"`
				} `json:"value"`
			} `json:"attributes"`
		} `json:"resource"`
		ScopeLogs []struct {
			LogRecords []struct {
				TimeUnixNano         string `json:"timeUnixNano"`
				ObservedTimeUnixNano string `json:"observedTimeUnixNano"`
				SeverityText         string `json:"severityText"`
				Body                 struct {
					StringValue *string `json:"stringValue,omitempty"`
				} `json:"body"`
				Attributes []struct {
					Key   string `json:"key"`
					Value struct {
						StringValue *string `json:"stringValue,omitempty"`
					} `json:"value"`
				} `json:"attributes"`
			} `json:"logRecords"`
		} `json:"scopeLogs"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal resource log: %w", err)
	}

	rl := &logspb.ResourceLogs{}

	// Parse resource attributes
	if len(raw.Resource.Attributes) > 0 {
		rl.Resource = &resourcepb.Resource{}
		for _, attr := range raw.Resource.Attributes {
			kv := &commonpb.KeyValue{Key: attr.Key}
			if attr.Value.StringValue != nil {
				kv.Value = &commonpb.AnyValue{
					Value: &commonpb.AnyValue_StringValue{StringValue: *attr.Value.StringValue},
				}
			} else if attr.Value.IntValue != nil {
				// OTLP JSON format uses string representation for int values
				kv.Value = &commonpb.AnyValue{
					Value: &commonpb.AnyValue_StringValue{StringValue: *attr.Value.IntValue},
				}
			}
			rl.Resource.Attributes = append(rl.Resource.Attributes, kv)
		}
	}

	// Parse scope logs
	for _, sl := range raw.ScopeLogs {
		scopeLog := &logspb.ScopeLogs{}
		for _, lr := range sl.LogRecords {
			record := &logspb.LogRecord{}

			// Parse timestamps
			if lr.TimeUnixNano != "" {
				var ts uint64
				_, _ = fmt.Sscanf(lr.TimeUnixNano, "%d", &ts)
				record.TimeUnixNano = ts
			}
			if lr.ObservedTimeUnixNano != "" {
				var ts uint64
				_, _ = fmt.Sscanf(lr.ObservedTimeUnixNano, "%d", &ts)
				record.ObservedTimeUnixNano = ts
			}

			// Parse severity
			record.SeverityText = lr.SeverityText

			// Parse body
			if lr.Body.StringValue != nil {
				record.Body = &commonpb.AnyValue{
					Value: &commonpb.AnyValue_StringValue{StringValue: *lr.Body.StringValue},
				}
			}

			// Parse log-level attributes
			for _, attr := range lr.Attributes {
				kv := &commonpb.KeyValue{Key: attr.Key}
				if attr.Value.StringValue != nil {
					kv.Value = &commonpb.AnyValue{
						Value: &commonpb.AnyValue_StringValue{StringValue: *attr.Value.StringValue},
					}
				}
				record.Attributes = append(record.Attributes, kv)
			}

			scopeLog.LogRecords = append(scopeLog.LogRecords, record)
		}
		rl.ScopeLogs = append(rl.ScopeLogs, scopeLog)
	}

	return rl, nil
}
