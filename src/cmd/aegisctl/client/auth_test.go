package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCanonicalAPIKeyString(t *testing.T) {
	got := canonicalAPIKeyString(
		"post",
		"/api/v2/auth/api-key/token",
		"1713333333",
		"abc123",
		"body_hash",
	)

	want := "POST\n/api/v2/auth/api-key/token\n1713333333\nabc123\nbody_hash"
	if got != want {
		t.Fatalf("canonical string mismatch:\nwant: %q\ngot:  %q", want, got)
	}
}

func TestBuildAPIKeyHeaders(t *testing.T) {
	headers, err := buildAPIKeyHeaders(
		"pk_demo",
		"ks_demo",
		time.Unix(1713333333, 0).UTC(),
		"/api/v2/auth/api-key/token",
	)
	if err != nil {
		t.Fatalf("buildAPIKeyHeaders returned error: %v", err)
	}

	if headers["X-Key-Id"] != "pk_demo" {
		t.Fatalf("unexpected key id header: %q", headers["X-Key-Id"])
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

func TestPrepareAPIKeyTokenDebug(t *testing.T) {
	debugInfo, err := PrepareAPIKeyTokenDebug(
		"pk_demo",
		"ks_demo",
		time.Unix(1713333333, 0).UTC(),
		"abc123",
	)
	if err != nil {
		t.Fatalf("PrepareAPIKeyTokenDebug returned error: %v", err)
	}

	if debugInfo.Method != "POST" {
		t.Fatalf("unexpected method: %q", debugInfo.Method)
	}
	if debugInfo.Path != "/api/v2/auth/api-key/token" {
		t.Fatalf("unexpected path: %q", debugInfo.Path)
	}
	if debugInfo.CanonicalString != "POST\n/api/v2/auth/api-key/token\n1713333333\nabc123\ne3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
		t.Fatalf("unexpected canonical string: %q", debugInfo.CanonicalString)
	}
	if debugInfo.BodySHA256 != "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
		t.Fatalf("unexpected body hash: %q", debugInfo.BodySHA256)
	}
	if debugInfo.Headers()["X-Signature"] != debugInfo.Signature {
		t.Fatal("signature header mismatch")
	}
}

func TestPostWithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Key-Id"); got != "pk_demo" {
			t.Fatalf("unexpected X-Key-Id header: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":0,"message":"ok"}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "", 5*time.Second)
	var resp APIResponse[map[string]any]
	if err := c.PostWithHeaders("/api/v2/auth/api-key/token", map[string]string{
		"X-Key-Id": "pk_demo",
	}, &resp); err != nil {
		t.Fatalf("PostWithHeaders returned error: %v", err)
	}
}
