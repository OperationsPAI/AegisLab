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

func TestBackendSubmitterPostsExpectedPayload(t *testing.T) {
	var (
		gotPath string
		gotBody map[string]any
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		_ = json.NewEncoder(w).Encode(aegisclient.APIResponse[any]{
			Code:    0,
			Message: "ok",
			Data:    map[string]any{"id": 1},
		})
	}))
	defer srv.Close()

	submitter := &backendSubmitter{
		client: aegisclient.NewClient(srv.URL, "token", time.Second),
		resolveProject: func() (int, error) {
			return 42, nil
		},
		defaults: &chaosDefaults{
			pedestal:    "ts",
			benchmark:   "clickhouse",
			interval:    60,
			preDuration: 30,
		},
	}

	spec := chaoscli.Spec{
		Type:      "NetworkDelay",
		Namespace: "ts",
		Target:    "frontend",
		Duration:  "30s",
		Params: map[string]any{
			"target_service": "checkout",
			"latency":        100,
		},
	}

	if err := submitter.Submit(context.Background(), spec); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	if gotPath != "/api/v2/projects/42/injections/inject" {
		t.Fatalf("expected project injection endpoint, got %q", gotPath)
	}
	pedestal, _ := gotBody["pedestal"].(map[string]any)
	benchmark, _ := gotBody["benchmark"].(map[string]any)
	if pedestal["name"] != "ts" || benchmark["name"] != "clickhouse" {
		t.Fatalf("unexpected defaults: %+v", gotBody)
	}
	specs, _ := gotBody["specs"].([]any)
	if len(specs) != 1 {
		t.Fatalf("expected one fault spec batch, got %+v", gotBody["specs"])
	}
	firstBatch, _ := specs[0].([]any)
	if len(firstBatch) != 1 {
		t.Fatalf("expected one node in first batch, got %+v", firstBatch)
	}
	node, _ := firstBatch[0].(map[string]any)
	if node["value"] == nil {
		t.Fatalf("expected translated node payload, got %+v", node)
	}
}
