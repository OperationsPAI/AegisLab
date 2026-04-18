package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	aegisclient "aegis/cmd/aegisctl/client"

	chaoscli "github.com/OperationsPAI/chaos-experiment/pkg/chaoscli"
)

// TestBackendSubmitterPostsExpectedPayload verifies that aegisctl's chaos
// subcommand sends the FaultSpec YAML shape directly to the project injection
// endpoint, without the old /translate round-trip.
func TestBackendSubmitterPostsExpectedPayload(t *testing.T) {
	var (
		gotInjectPath string
		gotInjectBody map[string]any
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/injections/translate":
			t.Fatalf("unexpected call to deprecated /translate endpoint")
		case "/api/v2/projects/42/injections/inject":
			gotInjectPath = r.URL.Path
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST inject, got %s", r.Method)
			}
			if err := json.NewDecoder(r.Body).Decode(&gotInjectBody); err != nil {
				t.Fatalf("Decode inject body: %v", err)
			}
			_ = json.NewEncoder(w).Encode(aegisclient.APIResponse[any]{
				Code:    0,
				Message: "ok",
				Data:    map[string]any{"id": 1},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	submitter := &backendSubmitter{
		client: aegisclient.NewClient(srv.URL, "token", time.Second),
		resolveProject: func() (int, error) {
			return 42, nil
		},
		defaults: &chaosDefaults{
			pedestal:    "",
			benchmark:   "clickhouse",
			interval:    60,
			preDuration: 30,
		},
	}

	spec := chaoscli.Spec{
		Type:      "NetworkDelay",
		Namespace: "ts",
		Target:    "0",
		Duration:  "30s",
		Params: map[string]any{
			"target_service": "checkout",
			"latency":        100,
		},
	}

	if err := submitter.Submit(context.Background(), spec); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	if gotInjectPath != "/api/v2/projects/42/injections/inject" {
		t.Fatalf("expected project injection endpoint, got %q", gotInjectPath)
	}

	pedestal, _ := gotInjectBody["pedestal"].(map[string]any)
	benchmark, _ := gotInjectBody["benchmark"].(map[string]any)
	if pedestal["name"] != "ts" || benchmark["name"] != "clickhouse" {
		t.Fatalf("unexpected defaults: %+v", gotInjectBody)
	}

	// specs should be the FaultSpec YAML shape: [][]FaultSpec, each with type/namespace/target/duration/params.
	specs, _ := gotInjectBody["specs"].([]any)
	if len(specs) != 1 {
		t.Fatalf("expected one fault spec batch, got %+v", gotInjectBody["specs"])
	}
	firstBatch, _ := specs[0].([]any)
	if len(firstBatch) != 1 {
		t.Fatalf("expected one fault spec in first batch, got %+v", firstBatch)
	}
	fs, _ := firstBatch[0].(map[string]any)
	if fs["type"] != "NetworkDelay" || fs["namespace"] != "ts" || fs["target"] != "0" || fs["duration"] != "30s" {
		t.Fatalf("unexpected fault spec payload: %+v", fs)
	}
	params, _ := fs["params"].(map[string]any)
	if params["target_service"] != "checkout" || params["latency"] != float64(100) {
		t.Fatalf("unexpected fault spec params: %+v", params)
	}
}
