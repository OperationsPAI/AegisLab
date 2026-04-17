package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStderr(fn func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	fn()
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = old
	return string(out)
}

func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old
	return string(out)
}

func TestOutputFormat_Constants(t *testing.T) {
	if FormatJSON != "json" {
		t.Errorf("FormatJSON = %q, want %q", FormatJSON, "json")
	}
}

func TestOutputFormat_Conversion(t *testing.T) {
	if OutputFormat("json") != FormatJSON {
		t.Errorf("OutputFormat(\"json\") != FormatJSON")
	}
	if OutputFormat("table") == FormatJSON {
		t.Errorf("OutputFormat(\"table\") should not equal FormatJSON")
	}
}

func TestPrintInfo_Normal(t *testing.T) {
	Quiet = false
	got := captureStderr(func() {
		PrintInfo("hello info")
	})
	if !strings.Contains(got, "hello info") {
		t.Errorf("PrintInfo output = %q, want it to contain %q", got, "hello info")
	}
}

func TestPrintInfo_Quiet(t *testing.T) {
	Quiet = true
	defer func() { Quiet = false }()

	got := captureStderr(func() {
		PrintInfo("should not appear")
	})
	if got != "" {
		t.Errorf("PrintInfo with Quiet=true produced output: %q", got)
	}
}

func TestPrintError(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		got := captureStderr(func() {
			PrintError(fmt.Errorf("something went wrong"))
		})
		if !strings.Contains(got, "ERROR:") {
			t.Errorf("PrintError output = %q, want prefix containing ERROR:", got)
		}
		if !strings.Contains(got, "something went wrong") {
			t.Errorf("PrintError output = %q, want it to contain %q", got, "something went wrong")
		}
	})

	t.Run("string", func(t *testing.T) {
		got := captureStderr(func() {
			PrintError("plain string error")
		})
		if !strings.Contains(got, "ERROR:") {
			t.Errorf("PrintError output = %q, want prefix containing ERROR:", got)
		}
		if !strings.Contains(got, "plain string error") {
			t.Errorf("PrintError output = %q, want it to contain %q", got, "plain string error")
		}
	})
}

func TestPrintJSON(t *testing.T) {
	data := map[string]string{"key": "value"}
	got := captureStdout(func() {
		PrintJSON(data)
	})

	var parsed map[string]string
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("PrintJSON output is not valid JSON: %v\nOutput: %q", err, got)
	}
	if parsed["key"] != "value" {
		t.Errorf("parsed[\"key\"] = %q, want %q", parsed["key"], "value")
	}
	if !strings.Contains(got, "  ") {
		t.Errorf("PrintJSON output should be indented, got: %q", got)
	}
}

func TestPrintJSON_NilAndEmpty(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		got := captureStdout(func() {
			PrintJSON(nil)
		})
		trimmed := strings.TrimSpace(got)
		if trimmed != "null" {
			t.Errorf("PrintJSON(nil) = %q, want %q", trimmed, "null")
		}
	})

	t.Run("empty map", func(t *testing.T) {
		got := captureStdout(func() {
			PrintJSON(map[string]any{})
		})
		trimmed := strings.TrimSpace(got)
		if trimmed != "{}" {
			t.Errorf("PrintJSON(empty map) = %q, want %q", trimmed, "{}")
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		got := captureStdout(func() {
			PrintJSON([]string{})
		})
		trimmed := strings.TrimSpace(got)
		if trimmed != "[]" {
			t.Errorf("PrintJSON(empty slice) = %q, want %q", trimmed, "[]")
		}
	})
}

func TestPrintTable(t *testing.T) {
	headers := []string{"NAME", "AGE", "CITY"}
	rows := [][]string{
		{"Alice", "30", "NYC"},
		{"Bob", "25", "LA"},
	}
	got := captureStdout(func() {
		PrintTable(headers, rows)
	})

	if !strings.Contains(got, "NAME") {
		t.Errorf("PrintTable output missing header NAME, got: %q", got)
	}
	if !strings.Contains(got, "AGE") {
		t.Errorf("PrintTable output missing header AGE, got: %q", got)
	}
	if !strings.Contains(got, "Alice") {
		t.Errorf("PrintTable output missing row data Alice, got: %q", got)
	}
	if !strings.Contains(got, "Bob") {
		t.Errorf("PrintTable output missing row data Bob, got: %q", got)
	}
}

func TestPrintTable_EmptyRows(t *testing.T) {
	headers := []string{"NAME", "STATUS"}
	got := captureStdout(func() {
		PrintTable(headers, nil)
	})

	if !strings.Contains(got, "NAME") {
		t.Errorf("PrintTable with empty rows missing header NAME, got: %q", got)
	}
	if !strings.Contains(got, "STATUS") {
		t.Errorf("PrintTable with empty rows missing header STATUS, got: %q", got)
	}
}
