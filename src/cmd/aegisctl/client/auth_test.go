package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCanonicalAccessKeyString(t *testing.T) {
	got := canonicalAccessKeyString(
		"post",
		"/api/v2/auth/access-key/token",
		"ak_demo",
		"1713333333",
		"abc123",
	)

	want := "POST\n/api/v2/auth/access-key/token\nak_demo\n1713333333\nabc123"
	if got != want {
		t.Fatalf("canonical string mismatch:\nwant: %q\ngot:  %q", want, got)
	}
}

func TestBuildAccessKeyHeaders(t *testing.T) {
	headers, err := buildAccessKeyHeaders(
		"ak_demo",
		"sk_demo",
		time.Unix(1713333333, 0).UTC(),
		"/api/v2/auth/access-key/token",
	)
	if err != nil {
		t.Fatalf("buildAccessKeyHeaders returned error: %v", err)
	}

	if headers["X-Access-Key"] != "ak_demo" {
		t.Fatalf("unexpected access key header: %q", headers["X-Access-Key"])
	}
	if headers["X-Timestamp"] != "1713333333" {
		t.Fatalf("unexpected timestamp header: %q", headers["X-Timestamp"])
	}
	if headers["X-Nonce"] == "" {
		t.Fatal("expected nonce header to be set")
	}
	if len(headers["X-Signature"]) != 64 {
		t.Fatalf("unexpected signature length: %d", len(headers["X-Signature"]))
	}
}

func TestPrepareAccessKeyTokenDebug(t *testing.T) {
	debugInfo, err := PrepareAccessKeyTokenDebug(
		"ak_demo",
		"sk_demo",
		time.Unix(1713333333, 0).UTC(),
		"abc123",
	)
	if err != nil {
		t.Fatalf("PrepareAccessKeyTokenDebug returned error: %v", err)
	}

	if debugInfo.Method != "POST" {
		t.Fatalf("unexpected method: %q", debugInfo.Method)
	}
	if debugInfo.Path != "/api/v2/auth/access-key/token" {
		t.Fatalf("unexpected path: %q", debugInfo.Path)
	}
	if debugInfo.CanonicalString != "POST\n/api/v2/auth/access-key/token\nak_demo\n1713333333\nabc123" {
		t.Fatalf("unexpected canonical string: %q", debugInfo.CanonicalString)
	}
	if debugInfo.Headers()["X-Signature"] != debugInfo.Signature {
		t.Fatal("signature header mismatch")
	}
}

func TestPostWithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Access-Key"); got != "ak_demo" {
			t.Fatalf("unexpected X-Access-Key header: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":0,"message":"ok"}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "", 5*time.Second)
	var resp APIResponse[map[string]any]
	if err := c.PostWithHeaders("/api/v2/auth/access-key/token", map[string]string{
		"X-Access-Key": "ak_demo",
	}, &resp); err != nil {
		t.Fatalf("PostWithHeaders returned error: %v", err)
	}
}
