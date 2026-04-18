package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
)

// WSReader reads text messages from a WebSocket endpoint.
type WSReader struct {
	url   string
	token string
}

// NewWSReader creates a new WebSocket reader for the given endpoint.
func NewWSReader(baseURL, path, token string) *WSReader {
	// Convert http(s):// to ws(s)://.
	wsURL := baseURL
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)
	wsURL = strings.Replace(wsURL, "http://", "ws://", 1)

	// Append token as query parameter.
	fullURL := wsURL + path
	if token != "" {
		u, err := url.Parse(fullURL)
		if err == nil {
			q := u.Query()
			q.Set("token", token)
			u.RawQuery = q.Encode()
			fullURL = u.String()
		}
	}

	return &WSReader{
		url:   fullURL,
		token: token,
	}
}

// Stream opens a WebSocket connection and returns channels for messages and errors.
// The caller should cancel the context to stop streaming.
func (r *WSReader) Stream(ctx context.Context) (<-chan string, <-chan error) {
	messages := make(chan string, 64)
	errs := make(chan error, 1)

	go func() {
		defer close(messages)
		defer close(errs)

		conn, _, err := websocket.DefaultDialer.DialContext(ctx, r.url, nil)
		if err != nil {
			errs <- fmt.Errorf("websocket connect: %w", err)
			return
		}
		defer func() {
			_ = conn.Close()
		}()

		// Close the connection when context is cancelled.
		go func() {
			<-ctx.Done()
			_ = conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			_ = conn.Close()
		}()

		for {
			msgType, data, err := conn.ReadMessage()
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					return
				}
				errs <- fmt.Errorf("websocket read: %w", err)
				return
			}

			if msgType == websocket.TextMessage {
				select {
				case messages <- string(data):
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return messages, errs
}
