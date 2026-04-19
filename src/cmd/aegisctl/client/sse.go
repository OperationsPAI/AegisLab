package client

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// SSEEvent represents a single Server-Sent Event.
type SSEEvent struct {
	ID    string
	Event string
	Data  string
}

// SSEReader reads Server-Sent Events from an HTTP endpoint.
type SSEReader struct {
	url    string
	token  string
	lastID string
	client *http.Client
}

// NewSSEReader creates a new SSE reader for the given endpoint.
func NewSSEReader(baseURL, path, token string) *SSEReader {
	return &SSEReader{
		url:   baseURL + path,
		token: token,
		client: &http.Client{
			Timeout: 0, // No timeout for streaming connections.
		},
	}
}

// Stream opens an SSE connection and returns channels for events and errors.
// The caller should cancel the context to stop streaming.
func (r *SSEReader) Stream(ctx context.Context) (<-chan SSEEvent, <-chan error) {
	events := make(chan SSEEvent, 64)
	errs := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errs)

		for {
			err := r.readStream(ctx, events)
			if err != nil {
				// If context was cancelled, exit cleanly.
				if ctx.Err() != nil {
					return
				}
				// Reconnect after a short delay.
				select {
				case <-ctx.Done():
					return
				case <-time.After(2 * time.Second):
					continue
				}
			}
			// Stream ended without error (server closed); try to reconnect.
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Second):
				continue
			}
		}
	}()

	return events, errs
}

func (r *SSEReader) readStream(ctx context.Context, events chan<- SSEEvent) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.url, nil)
	if err != nil {
		return fmt.Errorf("create SSE request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	if r.token != "" {
		req.Header.Set("Authorization", "Bearer "+r.token)
	}
	if r.lastID != "" {
		req.Header.Set("Last-Event-ID", r.lastID)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("SSE connect: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SSE server returned status %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	var current SSEEvent

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			// Empty line = event dispatch.
			if current.Data != "" || current.Event != "" {
				if current.ID != "" {
					r.lastID = current.ID
				}
				select {
				case events <- current:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			current = SSEEvent{}
			continue
		}

		if strings.HasPrefix(line, ":") {
			// Comment line, skip.
			continue
		}

		if strings.HasPrefix(line, "data:") {
			value := strings.TrimPrefix(line, "data:")
			value = strings.TrimLeft(value, " ")
			if current.Data != "" {
				current.Data += "\n" + value
			} else {
				current.Data = value
			}
		} else if strings.HasPrefix(line, "id:") {
			current.ID = strings.TrimLeft(strings.TrimPrefix(line, "id:"), " ")
		} else if strings.HasPrefix(line, "event:") {
			current.Event = strings.TrimLeft(strings.TrimPrefix(line, "event:"), " ")
		}
	}

	return scanner.Err()
}
