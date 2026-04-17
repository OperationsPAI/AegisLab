package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"aegis/cmd/aegisctl/output"
)

func newTestServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func healthyResponse() map[string]any {
	return map[string]any{
		"code":    200,
		"message": "success",
		"data": map[string]any{
			"status":  "healthy",
			"version": "1.0.0",
			"uptime":  "1h30m",
			"services": map[string]any{
				"redis":      map[string]any{"status": "healthy", "response_time": "1.2ms"},
				"database":   map[string]any{"status": "healthy", "response_time": "3.5ms"},
				"kubernetes": map[string]any{"status": "healthy", "response_time": "5ms"},
				"buildkit":   map[string]any{"status": "healthy", "response_time": "2ms"},
				"jaeger":     map[string]any{"status": "healthy", "response_time": "1ms"},
			},
		},
	}
}

func unhealthyResponse() map[string]any {
	return map[string]any{
		"code":    200,
		"message": "success",
		"data": map[string]any{
			"status":  "unhealthy",
			"version": "1.0.0",
			"uptime":  "1h30m",
			"services": map[string]any{
				"redis":      map[string]any{"status": "healthy", "response_time": "1.2ms"},
				"database":   map[string]any{"status": "unhealthy", "response_time": "N/A", "error": "connection refused"},
				"kubernetes": map[string]any{"status": "healthy", "response_time": "5ms"},
				"buildkit":   map[string]any{"status": "unhealthy", "response_time": "N/A", "error": "daemon not running"},
				"jaeger":     map[string]any{"status": "healthy", "response_time": "1ms"},
			},
		},
	}
}

func TestStatusHealthy(t *testing.T) {
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/system/health":
			json.NewEncoder(w).Encode(healthyResponse())
		default:
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{"code": 401, "message": "unauthorized"})
		}
	})
	defer ts.Close()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	flagServer = ts.URL
	flagOutput = "table"
	flagToken = ""
	statusCmd.RunE(nil, nil)

	w.Close()
	os.Stdout = old

	var buf [4096]byte
	n, _ := r.Read(buf[:])
	got := string(buf[:n])

	// Should contain green indicators for healthy services
	if !strings.Contains(got, "\u2713") {
		t.Errorf("Expected green check mark for healthy services, got:\n%s", got)
	}
	if !strings.Contains(got, "redis") {
		t.Errorf("Expected 'redis' in output, got:\n%s", got)
	}
}

func TestStatusUnhealthy(t *testing.T) {
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/system/health":
			json.NewEncoder(w).Encode(unhealthyResponse())
		default:
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{"code": 401, "message": "unauthorized"})
		}
	})
	defer ts.Close()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	flagServer = ts.URL
	flagOutput = "table"
	flagToken = ""
	statusCmd.RunE(nil, nil)

	w.Close()
	os.Stdout = old

	var buf [4096]byte
	n, _ := r.Read(buf[:])
	got := string(buf[:n])

	// Should contain red indicators for unhealthy services
	if !strings.Contains(got, "\u2717") {
		t.Errorf("Expected red X for unhealthy services, got:\n%s", got)
	}
	if !strings.Contains(got, "connection refused") {
		t.Errorf("Expected error message in output, got:\n%s", got)
	}
}

func TestStatusServerUnreachable(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	flagServer = "http://127.0.0.1:1" // unreachable port
	flagOutput = "table"
	flagToken = ""
	flagRequestTimeout = 1
	err := statusCmd.RunE(nil, nil)

	w.Close()
	os.Stdout = old

	var buf [4096]byte
	n, _ := r.Read(buf[:])
	got := string(buf[:n])

	// Should not crash
	if err != nil {
		t.Errorf("statusCmd should not return error when server unreachable, got: %v", err)
	}
	// Should show health endpoint failure
	if !strings.Contains(got, "\u2717") {
		t.Errorf("Expected failure indicator when server unreachable, got:\n%s", got)
	}
}

func TestStatusJSON(t *testing.T) {
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/system/health":
			json.NewEncoder(w).Encode(healthyResponse())
		case "/api/v2/auth/profile":
			json.NewEncoder(w).Encode(map[string]any{
				"code": 200, "message": "success",
				"data": map[string]any{"id": 1, "username": "admin"},
			})
		case "/api/v2/tasks":
			json.NewEncoder(w).Encode(map[string]any{
				"code": 200, "message": "success",
				"data": map[string]any{"items": []any{}, "pagination": map[string]any{"page": 1, "size": 100, "total": 0, "total_pages": 0}},
			})
		case "/api/v2/traces":
			json.NewEncoder(w).Encode(map[string]any{
				"code": 200, "message": "success",
				"data": map[string]any{"items": []any{}, "pagination": map[string]any{"page": 1, "size": 10, "total": 0, "total_pages": 0}},
			})
		default:
			w.WriteHeader(404)
		}
	})
	defer ts.Close()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	flagServer = ts.URL
	flagOutput = "json"
	flagToken = "test-token"
	output.Quiet = false
	statusCmd.RunE(nil, nil)

	w.Close()
	os.Stdout = old

	var buf [8192]byte
	n, _ := r.Read(buf[:])
	got := string(buf[:n])

	// Parse and verify JSON output
	var result map[string]any
	if err := json.Unmarshal([]byte(got), &result); err != nil {
		t.Fatalf("Output should be valid JSON: %v\nGot: %s", err, got)
	}

	health, ok := result["health"]
	if !ok {
		t.Fatalf("JSON output should contain 'health' key, got: %v", result)
	}

	healthMap, ok := health.(map[string]any)
	if !ok {
		t.Fatalf("health should be a map, got: %T", health)
	}

	if healthMap["status"] != "healthy" {
		t.Errorf("health.status should be 'healthy', got: %v", healthMap["status"])
	}

	services, ok := healthMap["services"].(map[string]any)
	if !ok {
		t.Fatalf("health.services should be a map, got: %T", healthMap["services"])
	}

	for _, svc := range []string{"redis", "database", "kubernetes"} {
		if _, exists := services[svc]; !exists {
			t.Errorf("health.services should contain %q", svc)
		}
	}
}
