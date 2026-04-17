package cmd

import "testing"

func TestNewWaitTimeoutError(t *testing.T) {
	got := newWaitTimeoutError(45, "trace", "trace-123", "Running")
	want := "timeout after 45s waiting for trace trace-123 (last state: Running)"

	if got == nil {
		t.Fatal("newWaitTimeoutError() returned nil")
	}
	if got.Error() != want {
		t.Fatalf("newWaitTimeoutError() = %q, want %q", got.Error(), want)
	}
}
