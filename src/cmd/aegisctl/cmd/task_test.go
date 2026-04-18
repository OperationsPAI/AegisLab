package cmd

import (
	"testing"
)

func TestFormatWait(t *testing.T) {
	cases := []struct {
		name     string
		delta    int64
		expected string
	}{
		{"due_now", 0, "+00:00"},
		{"future_small", 83, "+01:23"},
		{"overdue_small", -5, "-00:05"},
		{"future_large", 3661, "+61:01"},
		{"overdue_large", -125, "-02:05"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := formatWait(tc.delta)
			if got != tc.expected {
				t.Fatalf("formatWait(%d) = %q, want %q", tc.delta, got, tc.expected)
			}
		})
	}
}

func TestExecTimeField(t *testing.T) {
	if got := execTimeField(map[string]any{"execute_time": float64(1234567890)}); got != 1234567890 {
		t.Fatalf("float64 path: got %d", got)
	}
	if got := execTimeField(map[string]any{"execute_time": int64(42)}); got != 42 {
		t.Fatalf("int64 path: got %d", got)
	}
	if got := execTimeField(map[string]any{"execute_time": 7}); got != 7 {
		t.Fatalf("int path: got %d", got)
	}
	if got := execTimeField(map[string]any{}); got != 0 {
		t.Fatalf("missing key: got %d", got)
	}
	if got := execTimeField(map[string]any{"execute_time": nil}); got != 0 {
		t.Fatalf("nil value: got %d", got)
	}
	if got := execTimeField(map[string]any{"execute_time": "nope"}); got != 0 {
		t.Fatalf("unknown type: got %d", got)
	}
}
