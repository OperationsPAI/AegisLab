package injection

import (
	"encoding/json"
	"fmt"
	"testing"

	chaos "github.com/OperationsPAI/chaos-experiment/handler"
)

// mockConverter simulates the FriendlySpecToNode conversion for testing ResolveSpecs.
// It returns a Node whose Value equals the length of spec.Type (arbitrary but deterministic).
func mockConverter(spec *FriendlyFaultSpec) (chaos.Node, error) {
	if spec.Type == "" {
		return chaos.Node{}, fmt.Errorf("empty fault type")
	}
	if spec.Type == "UnknownType" {
		return chaos.Node{}, fmt.Errorf("unknown fault type: %s", spec.Type)
	}
	return chaos.Node{
		Value: len(spec.Type),
		Children: map[string]*chaos.Node{
			fmt.Sprintf("%d", len(spec.Type)): {
				Children: map[string]*chaos.Node{
					"0": {Value: 5}, // duration placeholder
				},
			},
		},
	}, nil
}

func TestResolveSpecs_FriendlySpec(t *testing.T) {
	friendly := FriendlyFaultSpec{
		Type:      "CPUStress",
		Namespace: "ts",
		Target:    "some-service",
		Duration:  "5m",
		Params:    map[string]any{"cpu_load": 80},
	}
	raw, err := json.Marshal(friendly)
	if err != nil {
		t.Fatalf("failed to marshal friendly spec: %v", err)
	}

	req := &SubmitInjectionReq{
		Specs: [][]json.RawMessage{{raw}},
	}

	err = req.ResolveSpecs(mockConverter)
	if err != nil {
		t.Fatalf("ResolveSpecs returned error: %v", err)
	}

	if len(req.ResolvedSpecs) != 1 {
		t.Fatalf("expected 1 batch, got %d", len(req.ResolvedSpecs))
	}
	if len(req.ResolvedSpecs[0]) != 1 {
		t.Fatalf("expected 1 node in batch, got %d", len(req.ResolvedSpecs[0]))
	}
	// mockConverter returns Value = len("CPUStress") = 9
	if req.ResolvedSpecs[0][0].Value != 9 {
		t.Errorf("expected node Value=9, got %d", req.ResolvedSpecs[0][0].Value)
	}
}

func TestResolveSpecs_NodeDSLPassthrough(t *testing.T) {
	node := chaos.Node{
		Value: 4,
		Children: map[string]*chaos.Node{
			"4": {
				Children: map[string]*chaos.Node{
					"0": {Value: 5},
					"1": {Value: 0},
					"2": {Value: 0},
					"3": {Value: 80},
					"4": {Value: 2},
				},
			},
		},
	}
	raw, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("failed to marshal node: %v", err)
	}

	req := &SubmitInjectionReq{
		Specs: [][]json.RawMessage{{raw}},
	}

	err = req.ResolveSpecs(mockConverter)
	if err != nil {
		t.Fatalf("ResolveSpecs returned error: %v", err)
	}

	if len(req.ResolvedSpecs) != 1 || len(req.ResolvedSpecs[0]) != 1 {
		t.Fatalf("expected 1x1 nodes, got %dx%d", len(req.ResolvedSpecs), len(req.ResolvedSpecs[0]))
	}
	resolved := req.ResolvedSpecs[0][0]
	if resolved.Value != 4 {
		t.Errorf("expected node Value=4 (CPUStress iota), got %d", resolved.Value)
	}
	if resolved.Children == nil {
		t.Fatal("expected children to be present")
	}
	child4, ok := resolved.Children["4"]
	if !ok {
		t.Fatal("expected child key '4' to exist")
	}
	if child4.Children["3"].Value != 80 {
		t.Errorf("expected cpu_load value=80, got %d", child4.Children["3"].Value)
	}
}

func TestResolveSpecs_MixedArray(t *testing.T) {
	friendly := FriendlyFaultSpec{
		Type:     "MemoryStress",
		Duration: "3m",
	}
	friendlyRaw, _ := json.Marshal(friendly)

	nodeDSL := chaos.Node{
		Value: 0,
		Children: map[string]*chaos.Node{
			"0": {Children: map[string]*chaos.Node{"0": {Value: 1}}},
		},
	}
	nodeRaw, _ := json.Marshal(nodeDSL)

	req := &SubmitInjectionReq{
		Specs: [][]json.RawMessage{{friendlyRaw, nodeRaw}},
	}

	err := req.ResolveSpecs(mockConverter)
	if err != nil {
		t.Fatalf("ResolveSpecs returned error: %v", err)
	}

	if len(req.ResolvedSpecs) != 1 || len(req.ResolvedSpecs[0]) != 2 {
		t.Fatalf("expected 1 batch with 2 specs, got %dx%d", len(req.ResolvedSpecs), len(req.ResolvedSpecs[0]))
	}

	// First is friendly: mockConverter returns Value = len("MemoryStress") = 12
	if req.ResolvedSpecs[0][0].Value != 12 {
		t.Errorf("expected friendly spec Value=12, got %d", req.ResolvedSpecs[0][0].Value)
	}
	// Second is Node DSL passthrough: Value=0
	if req.ResolvedSpecs[0][1].Value != 0 {
		t.Errorf("expected node DSL Value=0, got %d", req.ResolvedSpecs[0][1].Value)
	}
}

func TestResolveSpecs_MultipleBatches(t *testing.T) {
	friendly1 := FriendlyFaultSpec{Type: "CPUStress", Duration: "1m"}
	raw1, _ := json.Marshal(friendly1)

	friendly2 := FriendlyFaultSpec{Type: "PodKill", Duration: "2m"}
	raw2, _ := json.Marshal(friendly2)

	req := &SubmitInjectionReq{
		Specs: [][]json.RawMessage{{raw1}, {raw2}},
	}

	err := req.ResolveSpecs(mockConverter)
	if err != nil {
		t.Fatalf("ResolveSpecs returned error: %v", err)
	}

	if len(req.ResolvedSpecs) != 2 {
		t.Fatalf("expected 2 batches, got %d", len(req.ResolvedSpecs))
	}
	if len(req.ResolvedSpecs[0]) != 1 || len(req.ResolvedSpecs[1]) != 1 {
		t.Fatalf("expected 1 spec per batch, got %d and %d", len(req.ResolvedSpecs[0]), len(req.ResolvedSpecs[1]))
	}
}

func TestResolveSpecs_InvalidJSON(t *testing.T) {
	req := &SubmitInjectionReq{
		Specs: [][]json.RawMessage{{json.RawMessage(`{invalid json`)}},
	}

	err := req.ResolveSpecs(mockConverter)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestResolveSpecs_FriendlyConverterError(t *testing.T) {
	friendly := FriendlyFaultSpec{
		Type:     "UnknownType",
		Duration: "5m",
	}
	raw, _ := json.Marshal(friendly)

	req := &SubmitInjectionReq{
		Specs: [][]json.RawMessage{{raw}},
	}

	err := req.ResolveSpecs(mockConverter)
	if err == nil {
		t.Fatal("expected error from converter for unknown type, got nil")
	}
}

func TestResolveSpecs_EmptySpecs(t *testing.T) {
	req := &SubmitInjectionReq{
		Specs: [][]json.RawMessage{},
	}

	err := req.ResolveSpecs(mockConverter)
	if err != nil {
		t.Fatalf("unexpected error for empty specs: %v", err)
	}
	if len(req.ResolvedSpecs) != 0 {
		t.Errorf("expected 0 batches for empty specs, got %d", len(req.ResolvedSpecs))
	}
}

func TestResolveSpecs_EmptyBatch(t *testing.T) {
	req := &SubmitInjectionReq{
		Specs: [][]json.RawMessage{{}},
	}

	err := req.ResolveSpecs(mockConverter)
	if err != nil {
		t.Fatalf("unexpected error for empty batch: %v", err)
	}
	if len(req.ResolvedSpecs) != 1 {
		t.Fatalf("expected 1 batch, got %d", len(req.ResolvedSpecs))
	}
	if len(req.ResolvedSpecs[0]) != 0 {
		t.Errorf("expected 0 nodes in empty batch, got %d", len(req.ResolvedSpecs[0]))
	}
}

func TestResolveSpecs_FriendlyEmptyType(t *testing.T) {
	// A spec with "type":"" — the string "type" key is present, so it's detected as friendly.
	// The converter should return an error for empty type.
	friendly := FriendlyFaultSpec{
		Type: "",
	}
	raw, _ := json.Marshal(friendly)

	req := &SubmitInjectionReq{
		Specs: [][]json.RawMessage{{raw}},
	}

	err := req.ResolveSpecs(mockConverter)
	if err == nil {
		t.Fatal("expected error for spec with empty type, got nil")
	}
}

func TestResolveSpecs_NodeDSLWithNameField(t *testing.T) {
	// A Node DSL that has a "name" field but no "type" field — should be treated as Node DSL
	raw := json.RawMessage(`{"value":4,"children":{"4":{"children":{"0":{"value":5}}}},"name":"test"}`)

	req := &SubmitInjectionReq{
		Specs: [][]json.RawMessage{{raw}},
	}

	err := req.ResolveSpecs(mockConverter)
	if err != nil {
		t.Fatalf("ResolveSpecs returned error: %v", err)
	}

	if req.ResolvedSpecs[0][0].Value != 4 {
		t.Errorf("expected node Value=4, got %d", req.ResolvedSpecs[0][0].Value)
	}
}
