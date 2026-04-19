package loki

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

type Client struct {
	address    string
	httpClient *http.Client
}

type QueryOpts struct {
	Start     time.Time
	End       time.Time
	Limit     int
	Direction string
}

type queryRangeResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Stream map[string]string `json:"stream"`
			Values [][]string        `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

func NewClient() *Client {
	address := config.GetString("loki.address")
	timeout := config.GetString("loki.timeout")
	timeoutDuration := 10 * time.Second
	if timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			timeoutDuration = d
		}
	}

	return &Client{
		address: address,
		httpClient: &http.Client{
			Timeout: timeoutDuration,
		},
	}
}

func (c *Client) QueryJobLogs(ctx context.Context, taskID string, opts QueryOpts) ([]dto.LogEntry, error) {
	if taskID == "" {
		return nil, fmt.Errorf("taskID is required")
	}

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

	logQL := fmt.Sprintf(`{app="rcabench"} | task_id=%q`, taskID)

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

	var lokiResp queryRangeResponse
	if err := json.Unmarshal(body, &lokiResp); err != nil {
		return nil, fmt.Errorf("failed to parse Loki response: %w", err)
	}
	if lokiResp.Status != "success" {
		return nil, fmt.Errorf("loki query status: %s", lokiResp.Status)
	}
	if len(lokiResp.Data.Result) == 0 {
		logrus.Warnf("Loki returned 0 streams for task %s, raw response: %s", taskID, string(body))
	}

	var entries []dto.LogEntry
	for _, result := range lokiResp.Data.Result {
		for _, value := range result.Values {
			if len(value) < 2 {
				continue
			}
			nsec, err := strconv.ParseInt(value[0], 10, 64)
			if err != nil {
				logrus.Warnf("Loki: invalid timestamp %s: %v", value[0], err)
				continue
			}

			entries = append(entries, dto.LogEntry{
				Timestamp: time.Unix(0, nsec),
				Line:      value[1],
				TaskID:    taskID,
				TraceID:   result.Stream["trace_id"],
				JobID:     result.Stream["job_id"],
			})
		}
	}

	logrus.Infof("Loki: queried %d log entries for task %s", len(entries), taskID)
	return entries, nil
}
