package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// APIResponse is the standard response envelope returned by the AegisLab API.
type APIResponse[T any] struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      T      `json:"data,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
	Errors    []any  `json:"errors,omitempty"`
}

// Pagination contains pagination metadata.
type Pagination struct {
	Page       int `json:"page"`
	Size       int `json:"size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// PaginatedData wraps a list of items together with pagination info.
type PaginatedData[T any] struct {
	Items      []T        `json:"items"`
	Pagination Pagination `json:"pagination"`
}

// APIError is returned when the server responds with a non-2xx status.
type APIError struct {
	StatusCode int
	Code       int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// Client is the core HTTP client for the AegisLab API.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// NewClient creates a new API client.
func NewClient(baseURL, token string, timeout time.Duration) *Client {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// doRequest executes an HTTP request and decodes the JSON response into dest.
func (c *Client) doRequest(method, path string, body any, dest any) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	url := c.BaseURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiResp APIResponse[any]
		if json.Unmarshal(respBody, &apiResp) == nil && apiResp.Message != "" {
			return &APIError{
				StatusCode: resp.StatusCode,
				Code:       apiResp.Code,
				Message:    apiResp.Message,
			}
		}
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	if dest != nil {
		if err := json.Unmarshal(respBody, dest); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// Get sends a GET request.
func (c *Client) Get(path string, dest any) error {
	return c.doRequest(http.MethodGet, path, nil, dest)
}

// Post sends a POST request.
func (c *Client) Post(path string, body any, dest any) error {
	return c.doRequest(http.MethodPost, path, body, dest)
}

// Put sends a PUT request.
func (c *Client) Put(path string, body any, dest any) error {
	return c.doRequest(http.MethodPut, path, body, dest)
}

// Patch sends a PATCH request.
func (c *Client) Patch(path string, body any, dest any) error {
	return c.doRequest(http.MethodPatch, path, body, dest)
}

// Delete sends a DELETE request.
func (c *Client) Delete(path string, dest any) error {
	return c.doRequest(http.MethodDelete, path, nil, dest)
}
