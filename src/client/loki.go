package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"aegis/config"
	"aegis/dto"

	"github.com/sirupsen/logrus"
)

// LokiClient wraps the Loki HTTP API for querying historical logs
type LokiClient struct {
	address    string
	httpClient *http.Client
}

// QueryOpts defines options for Loki log queries
type QueryOpts struct {
	Start     time.Time // Query start time (default: 1 hour ago)
	End       time.Time // Query end time (default: now)
	Limit     int       // Max entries to return (default: 5000)
	Direction string    // "forward" (chronological) or "backward"
}

// lokiQueryRangeResponse represents the Loki query_range API response
type lokiQueryRangeResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Stream map[string]string `json:"stream"`
			Values [][]string        `json:"values"` // [[nanosecond_timestamp, log_line], ...]
		} `json:"result"`
	} `json:"data"`
}

// NewLokiClient creates a new Loki client using configuration
func NewLokiClient() *LokiClient {
	address := config.GetString("loki.address")
	timeout := config.GetString("loki.timeout")
	timeoutDuration := 10 * time.Second
	if timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			timeoutDuration = d
		}
	}

	return &LokiClient{
		address: address,
		httpClient: &http.Client{
			Timeout: timeoutDuration,
		},
	}
}

// QueryJobLogs queries historical job logs from Loki by task_id
func (c *LokiClient) QueryJobLogs(ctx context.Context, taskID string, opts QueryOpts) ([]dto.LogEntry, error) {
	if taskID == "" {
		return nil, fmt.Errorf("taskID is required")
	}

	// Apply defaults
	if opts.Start.IsZero() {
		opts.Start = time.Now().Add(-1 * time.Hour)
	}
	if opts.End.IsZero() {
		opts.End = time.Now()
	}
	if opts.Limit <= 0 {
		maxEntries := config.GetInt("loki.max_entries")
		if maxEntries > 0 {
			opts.Limit = maxEntries
		} else {
			opts.Limit = 5000
		}
	}
	if opts.Direction == "" {
		opts.Direction = "forward"
	}

	// Build LogQL query
	// Use Structured Metadata filter since task_id is stored as structured metadata in Loki
	logQL := fmt.Sprintf(`{app="rcabench"} | task_id=%q`, taskID)

	// Build request URL
	params := url.Values{}
	params.Set("query", logQL)
	params.Set("start", strconv.FormatInt(opts.Start.UnixNano(), 10))
	params.Set("end", strconv.FormatInt(opts.End.UnixNano(), 10))
	params.Set("limit", strconv.Itoa(opts.Limit))
	params.Set("direction", opts.Direction)

	reqURL := fmt.Sprintf("%s/loki/api/v1/query_range?%s", c.address, params.Encode())

	logrus.Infof("Loki query: url=%s", reqURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Loki request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("loki query failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("loki query returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Loki response: %w", err)
	}

	var lokiResp lokiQueryRangeResponse
	if err := json.Unmarshal(body, &lokiResp); err != nil {
		return nil, fmt.Errorf("failed to parse Loki response: %w", err)
	}

	if lokiResp.Status != "success" {
		return nil, fmt.Errorf("loki query status: %s", lokiResp.Status)
	}

	if len(lokiResp.Data.Result) == 0 {
		logrus.Warnf("Loki returned 0 streams for task %s, raw response: %s", taskID, string(body))
	}

	// Convert Loki results to LogEntry
	var entries []dto.LogEntry
	for _, result := range lokiResp.Data.Result {
		for _, value := range result.Values {
			if len(value) < 2 {
				continue
			}

			// Parse nanosecond timestamp
			nsec, err := strconv.ParseInt(value[0], 10, 64)
			if err != nil {
				logrus.Warnf("Loki: invalid timestamp %s: %v", value[0], err)
				continue
			}

			entry := dto.LogEntry{
				Timestamp: time.Unix(0, nsec),
				Line:      value[1],
				TaskID:    taskID,
				// Extract additional metadata from stream labels if available
				TraceID: result.Stream["trace_id"],
				JobID:   result.Stream["job_id"],
			}

			entries = append(entries, entry)
		}
	}

	logrus.Infof("Loki: queried %d log entries for task %s", len(entries), taskID)
	return entries, nil
}
