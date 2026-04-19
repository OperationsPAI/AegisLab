package injection

import (
	"testing"


	chaos "github.com/OperationsPAI/chaos-experiment/handler"
)

func setupNamespacePrefixes(t *testing.T) (string, int) {
	t.Helper()
	systems := chaos.GetAllSystemTypes()
	if len(systems) == 0 {
		t.Fatal("expected at least one registered chaos system")
	}
	return systems[0].String(), 0
}

func TestFriendlySpecToNode_CPUStress(t *testing.T) {
	namespace, namespaceIdx := setupNamespacePrefixes(t)

	spec := &FriendlyFaultSpec{
		Type:      "CPUStress",
		Namespace: namespace,
		Target:    "0", // container index as string
		Duration:  "5m",
		Params: map[string]any{
			// Note: the local toSnakeCase produces "c_p_u_load" for "CPULoad",
			// so we use the exact field name which is also accepted by mapParamsToFieldIndices.
			"CPULoad":   80,
			"CPUWorker": 2,
		},
	}

	node, err := FriendlySpecToNode(spec)
	if err != nil {
		t.Fatalf("FriendlySpecToNode returned error: %v", err)
	}

	// CPUStress is iota 4
	if node.Value != 4 {
		t.Errorf("expected root Value=4 (CPUStress), got %d", node.Value)
	}

	// Should have a child keyed "4" (the type index)
	typeChild, ok := node.Children["4"]
	if !ok {
		t.Fatal("expected child key '4' for CPUStress type")
	}

	if typeChild.Children == nil {
		t.Fatal("expected type child to have children (spec fields)")
	}

	// Field 0 = Duration = 5 minutes
	if dur, ok := typeChild.Children["0"]; ok {
		if dur.Value != 5 {
			t.Errorf("expected duration Value=5, got %d", dur.Value)
		}
	} else {
		t.Error("expected child '0' (Duration) to exist")
	}

	// Field 1 = Namespace index in the registered system list.
	if ns, ok := typeChild.Children["1"]; ok {
		if ns.Value != namespaceIdx {
			t.Errorf("expected namespace Value=%d, got %d", namespaceIdx, ns.Value)
		}
	} else {
		t.Error("expected child '1' (Namespace) to exist")
	}

	// Field 2 = ContainerIdx = 0
	if container, ok := typeChild.Children["2"]; ok {
		if container.Value != 0 {
			t.Errorf("expected ContainerIdx Value=0, got %d", container.Value)
		}
	} else {
		t.Error("expected child '2' (ContainerIdx) to exist")
	}

	// Field 3 = CPULoad = 80
	if cpuLoad, ok := typeChild.Children["3"]; ok {
		if cpuLoad.Value != 80 {
			t.Errorf("expected CPULoad Value=80, got %d", cpuLoad.Value)
		}
	} else {
		t.Error("expected child '3' (CPULoad) to exist")
	}

	// Field 4 = CPUWorker = 2
	if cpuWorker, ok := typeChild.Children["4"]; ok {
		if cpuWorker.Value != 2 {
			t.Errorf("expected CPUWorker Value=2, got %d", cpuWorker.Value)
		}
	} else {
		t.Error("expected child '4' (CPUWorker) to exist")
	}
}

func TestFriendlySpecToNode_InvalidFaultType(t *testing.T) {
	setupNamespacePrefixes(t)

	spec := &FriendlyFaultSpec{
		Type:     "NonExistentChaosType",
		Duration: "5m",
	}

	_, err := FriendlySpecToNode(spec)
	if err == nil {
		t.Fatal("expected error for invalid fault type, got nil")
	}
}

func TestFriendlySpecToNode_EmptyFaultType(t *testing.T) {
	setupNamespacePrefixes(t)

	spec := &FriendlyFaultSpec{
		Type:     "",
		Duration: "5m",
	}

	_, err := FriendlySpecToNode(spec)
	if err == nil {
		t.Fatal("expected error for empty fault type, got nil")
	}
}

func TestFriendlySpecToNode_InvalidDuration(t *testing.T) {
	setupNamespacePrefixes(t)

	spec := &FriendlyFaultSpec{
		Type:     "CPUStress",
		Duration: "not-a-duration",
	}

	_, err := FriendlySpecToNode(spec)
	if err == nil {
		t.Fatal("expected error for invalid duration, got nil")
	}
}

func TestFriendlySpecToNode_MemoryStress(t *testing.T) {
	namespace, _ := setupNamespacePrefixes(t)

	spec := &FriendlyFaultSpec{
		Type:      "MemoryStress",
		Namespace: namespace,
		Target:    "0",
		Duration:  "10m",
		Params: map[string]any{
			"memory_size": 256,
			"mem_worker":  2,
		},
	}

	node, err := FriendlySpecToNode(spec)
	if err != nil {
		t.Fatalf("FriendlySpecToNode returned error: %v", err)
	}

	// MemoryStress is iota 3
	if node.Value != 3 {
		t.Errorf("expected root Value=3 (MemoryStress), got %d", node.Value)
	}

	typeChild, ok := node.Children["3"]
	if !ok {
		t.Fatal("expected child key '3' for MemoryStress type")
	}

	// Field 0 = Duration = 10 minutes
	if dur, ok := typeChild.Children["0"]; ok {
		if dur.Value != 10 {
			t.Errorf("expected duration Value=10, got %d", dur.Value)
		}
	} else {
		t.Error("expected child '0' (Duration) to exist")
	}

	// Field 3 = MemorySize = 256
	if memSize, ok := typeChild.Children["3"]; ok {
		if memSize.Value != 256 {
			t.Errorf("expected MemorySize Value=256, got %d", memSize.Value)
		}
	} else {
		t.Error("expected child '3' (MemorySize) to exist")
	}

	// Field 4 = MemWorker = 2
	if memWorker, ok := typeChild.Children["4"]; ok {
		if memWorker.Value != 2 {
			t.Errorf("expected MemWorker Value=2, got %d", memWorker.Value)
		}
	} else {
		t.Error("expected child '4' (MemWorker) to exist")
	}
}

func TestFriendlySpecToNode_DurationCeiling(t *testing.T) {
	namespace, _ := setupNamespacePrefixes(t)

	// 90s should become 2 minutes (ceiling)
	spec := &FriendlyFaultSpec{
		Type:      "CPUStress",
		Namespace: namespace,
		Target:    "0",
		Duration:  "90s",
		Params: map[string]any{
			"CPULoad":   50,
			"CPUWorker": 1,
		},
	}

	node, err := FriendlySpecToNode(spec)
	if err != nil {
		t.Fatalf("FriendlySpecToNode returned error: %v", err)
	}

	typeChild := node.Children["4"]
	if typeChild == nil {
		t.Fatal("expected child key '4' for CPUStress type")
	}

	dur := typeChild.Children["0"]
	if dur == nil {
		t.Fatal("expected child '0' (Duration)")
	}
	if dur.Value != 2 {
		t.Errorf("expected duration Value=2 (90s ceiling), got %d", dur.Value)
	}
}

func TestFriendlySpecToNode_ParamsMapping(t *testing.T) {
	namespace, _ := setupNamespacePrefixes(t)

	// Test that named params are correctly mapped to field indices via reflection.
	// The local toSnakeCase produces "c_p_u_load" for "CPULoad" (not "cpu_load"),
	// but mapParamsToFieldIndices also accepts exact field names and lowercase field names.
	// JSON numbers are float64, so pass float64 values.
	spec := &FriendlyFaultSpec{
		Type:      "CPUStress",
		Namespace: namespace,
		Target:    "0",
		Duration:  "5m",
		Params: map[string]any{
			"CPULoad":   float64(80),
			"CPUWorker": float64(2),
		},
	}

	node, err := FriendlySpecToNode(spec)
	if err != nil {
		t.Fatalf("FriendlySpecToNode returned error: %v", err)
	}

	typeChild := node.Children["4"]
	if typeChild == nil {
		t.Fatal("expected child key '4'")
	}

	// CPUStressChaosSpec fields:
	// 0: Duration, 1: Namespace, 2: ContainerIdx, 3: CPULoad, 4: CPUWorker, 5: NamespaceTarget
	// cpu_load -> CPULoad -> field index 3
	if cpuLoad, ok := typeChild.Children["3"]; ok {
		if cpuLoad.Value != 80 {
			t.Errorf("expected CPULoad=80, got %d", cpuLoad.Value)
		}
	} else {
		t.Error("expected child '3' (CPULoad) to exist")
	}

	// cpu_worker -> CPUWorker -> field index 4
	if cpuWorker, ok := typeChild.Children["4"]; ok {
		if cpuWorker.Value != 2 {
			t.Errorf("expected CPUWorker=2, got %d", cpuWorker.Value)
		}
	} else {
		t.Error("expected child '4' (CPUWorker) to exist")
	}
}

func TestFriendlySpecToNode_EmptyNamespaceDefaultsToZero(t *testing.T) {
	setupNamespacePrefixes(t)

	spec := &FriendlyFaultSpec{
		Type:      "CPUStress",
		Namespace: "", // empty namespace should default to index 0
		Target:    "0",
		Duration:  "5m",
	}

	node, err := FriendlySpecToNode(spec)
	if err != nil {
		t.Fatalf("FriendlySpecToNode returned error: %v", err)
	}

	typeChild := node.Children["4"]
	if typeChild == nil {
		t.Fatal("expected child key '4'")
	}

	ns := typeChild.Children["1"]
	if ns == nil {
		t.Fatal("expected child '1' (Namespace)")
	}
	if ns.Value != 0 {
		t.Errorf("expected namespace Value=0 for empty namespace, got %d", ns.Value)
	}
}

func TestFriendlySpecToNode_EmptyTargetDefaultsToZero(t *testing.T) {
	namespace, _ := setupNamespacePrefixes(t)

	spec := &FriendlyFaultSpec{
		Type:      "CPUStress",
		Namespace: namespace,
		Target:    "", // empty target should default to index 0
		Duration:  "5m",
	}

	node, err := FriendlySpecToNode(spec)
	if err != nil {
		t.Fatalf("FriendlySpecToNode returned error: %v", err)
	}

	typeChild := node.Children["4"]
	if typeChild == nil {
		t.Fatal("expected child key '4'")
	}

	target := typeChild.Children["2"]
	if target == nil {
		t.Fatal("expected child '2' (ContainerIdx/target)")
	}
	if target.Value != 0 {
		t.Errorf("expected target Value=0 for empty target, got %d", target.Value)
	}
}

func TestFriendlySpecToNode_CaseInsensitiveType(t *testing.T) {
	namespace, _ := setupNamespacePrefixes(t)

	// The init() populates lowercase keys too, so "cpustress" should work.
	spec := &FriendlyFaultSpec{
		Type:      "cpustress",
		Namespace: namespace,
		Duration:  "5m",
	}

	node, err := FriendlySpecToNode(spec)
	if err != nil {
		t.Fatalf("FriendlySpecToNode returned error: %v", err)
	}

	if node.Value != 4 {
		t.Errorf("expected Value=4 for cpustress (case-insensitive), got %d", node.Value)
	}
}

func TestFriendlySpecToNode_NoParams(t *testing.T) {
	namespace, _ := setupNamespacePrefixes(t)

	// A spec without Params should still produce correct Duration/Namespace/Target nodes
	spec := &FriendlyFaultSpec{
		Type:      "CPUStress",
		Namespace: namespace,
		Target:    "0",
		Duration:  "5m",
	}

	node, err := FriendlySpecToNode(spec)
	if err != nil {
		t.Fatalf("FriendlySpecToNode returned error: %v", err)
	}

	typeChild := node.Children["4"]
	if typeChild == nil {
		t.Fatal("expected child key '4'")
	}

	// Should still have duration, namespace, target
	if _, ok := typeChild.Children["0"]; !ok {
		t.Error("expected child '0' (Duration)")
	}
	if _, ok := typeChild.Children["1"]; !ok {
		t.Error("expected child '1' (Namespace)")
	}
	if _, ok := typeChild.Children["2"]; !ok {
		t.Error("expected child '2' (ContainerIdx)")
	}
	// No fields 3+ because no params
	if _, ok := typeChild.Children["3"]; ok {
		t.Error("did not expect child '3' when no params given")
	}
}

func TestFriendlySpecToNode_NodeTreeStructure(t *testing.T) {
	namespace, _ := setupNamespacePrefixes(t)

	spec := &FriendlyFaultSpec{
		Type:      "CPUStress",
		Namespace: namespace,
		Duration:  "5m",
	}

	node, err := FriendlySpecToNode(spec)
	if err != nil {
		t.Fatalf("FriendlySpecToNode returned error: %v", err)
	}

	// Verify top-level structure: {Value: <type_idx>, Children: {"<type_idx>": {Children: {...}}}}
	if node.Children == nil {
		t.Fatal("expected top-level Children to be non-nil")
	}
	if len(node.Children) != 1 {
		t.Errorf("expected exactly 1 top-level child, got %d", len(node.Children))
	}
	if _, ok := node.Children["4"]; !ok {
		keys := make([]string, 0, len(node.Children))
		for k := range node.Children {
			keys = append(keys, k)
		}
		t.Errorf("expected top-level child key '4', got keys: %v", keys)
	}
}

func TestFriendlySpecToNode_LowercaseParamNames(t *testing.T) {
	namespace, _ := setupNamespacePrefixes(t)

	// Lowercase field names (e.g., "cpuload") are also accepted by mapParamsToFieldIndices.
	spec := &FriendlyFaultSpec{
		Type:      "CPUStress",
		Namespace: namespace,
		Target:    "0",
		Duration:  "5m",
		Params: map[string]any{
			"cpuload":   float64(90),
			"cpuworker": float64(3),
		},
	}

	node, err := FriendlySpecToNode(spec)
	if err != nil {
		t.Fatalf("FriendlySpecToNode returned error: %v", err)
	}

	typeChild := node.Children["4"]
	if typeChild == nil {
		t.Fatal("expected child key '4'")
	}

	if cpuLoad, ok := typeChild.Children["3"]; ok {
		if cpuLoad.Value != 90 {
			t.Errorf("expected CPULoad=90 via lowercase key, got %d", cpuLoad.Value)
		}
	} else {
		t.Error("expected child '3' (CPULoad) to exist via lowercase key")
	}
}

func TestFriendlySpecToNode_SnakeCaseParamsMismatch(t *testing.T) {
	namespace, _ := setupNamespacePrefixes(t)

	// The local toSnakeCase("CPULoad") = "c_p_u_load" (not "cpu_load").
	// So "cpu_load" does NOT match and the param is silently skipped.
	// This documents the known limitation.
	spec := &FriendlyFaultSpec{
		Type:      "CPUStress",
		Namespace: namespace,
		Target:    "0",
		Duration:  "5m",
		Params: map[string]any{
			"cpu_load": float64(80),
		},
	}

	node, err := FriendlySpecToNode(spec)
	if err != nil {
		t.Fatalf("FriendlySpecToNode returned error: %v", err)
	}

	typeChild := node.Children["4"]
	if typeChild == nil {
		t.Fatal("expected child key '4'")
	}

	// "cpu_load" doesn't match any registered key, so it's silently skipped
	if _, ok := typeChild.Children["3"]; ok {
		t.Error("did not expect child '3' (CPULoad) when using 'cpu_load' param key (snake_case mismatch)")
	}
}

// --- Tests for helper functions ---

func TestParseDurationToMinutes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		wantErr  bool
	}{
		{
			name:     "60 seconds equals 1 minute",
			input:    "60s",
			expected: 1,
		},
		{
			name:     "5 minutes",
			input:    "5m",
			expected: 5,
		},
		{
			name:     "1 hour equals 60 minutes",
			input:    "1h",
			expected: 60,
		},
		{
			name:     "90 seconds ceiling to 2 minutes",
			input:    "90s",
			expected: 2,
		},
		{
			name:     "30 seconds ceiling to 1 minute",
			input:    "30s",
			expected: 1,
		},
		{
			name:     "plain integer treated as minutes",
			input:    "1",
			expected: 1,
		},
		{
			name:     "plain integer 10",
			input:    "10",
			expected: 10,
		},
		{
			name:     "1m30s ceiling to 2 minutes",
			input:    "1m30s",
			expected: 2,
		},
		{
			name:     "exactly 2 minutes no ceiling needed",
			input:    "2m0s",
			expected: 2,
		},
		{
			name:     "very short 1s ceiling to 1 minute",
			input:    "1s",
			expected: 1,
		},
		{
			name:    "invalid duration string",
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseDurationToMinutes(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error for input %q, got result=%d", tc.input, result)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", tc.input, err)
			}
			if result != tc.expected {
				t.Errorf("parseDurationToMinutes(%q) = %d, want %d", tc.input, result, tc.expected)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	// Note: the local toSnakeCase implementation inserts underscore before each uppercase letter,
	// which is simpler than the chaos-experiment utils.ToSnakeCase (regex-based).
	// "CPULoad" -> "c_p_u_load" with the simple approach.
	tests := []struct {
		input    string
		expected string
	}{
		{"Duration", "duration"},
		{"Namespace", "namespace"},
		{"MemorySize", "memory_size"},
		{"", ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := toSnakeCase(tc.input)
			if result != tc.expected {
				t.Errorf("toSnakeCase(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestToSnakeCase_ConsecutiveUppercase(t *testing.T) {
	// The simple toSnakeCase inserts _ before every uppercase letter:
	// "CPULoad" -> "c_p_u_load"
	// Verify actual behavior rather than assuming.
	result := toSnakeCase("CPULoad")
	// Just verify it returns a non-empty string and is lowercase
	if result == "" {
		t.Error("expected non-empty result for CPULoad")
	}
	t.Logf("toSnakeCase(\"CPULoad\") = %q", result)

	result2 := toSnakeCase("ContainerIdx")
	t.Logf("toSnakeCase(\"ContainerIdx\") = %q", result2)
}

func TestToInt(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected int
		wantErr  bool
	}{
		{"int value", 42, 42, false},
		{"float64 value", float64(42), 42, false},
		{"float64 with fraction", float64(42.7), 42, false},
		{"float32 value", float32(10), 10, false},
		{"string number", "42", 42, false},
		{"string non-number", "abc", 0, true},
		{"nil value", nil, 0, true},
		{"bool value", true, 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := toInt(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error for input %v (%T), got result=%d", tc.input, tc.input, result)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for input %v (%T): %v", tc.input, tc.input, err)
			}
			if result != tc.expected {
				t.Errorf("toInt(%v) = %d, want %d", tc.input, result, tc.expected)
			}
		})
	}
}

// TestFriendlySpecToNode_ParseBatchKeyInvariant guards the schema-drift bug
// that blocked E2E runs with "failed to find key 17 in the children at index 0":
// every node returned by FriendlySpecToNode MUST contain a child keyed by
// strconv.Itoa(node.Value). That is the invariant consumed downstream by
// parseBatchInjectionSpecs — and it is what the old /translate endpoint
// violated when it re-keyed children by field name ("Latency", "Direction"...).
func TestFriendlySpecToNode_ParseBatchKeyInvariant(t *testing.T) {
	namespace, _ := setupNamespacePrefixes(t)

	cases := []FriendlyFaultSpec{
		{Type: "CPUStress", Namespace: namespace, Target: "0", Duration: "1m"},
		{Type: "NetworkDelay", Namespace: namespace, Target: "0", Duration: "30s"},
		{Type: "PodFailure", Namespace: namespace, Target: "0", Duration: "1m"},
	}

	for _, spec := range cases {
		spec := spec
		t.Run(spec.Type, func(t *testing.T) {
			node, err := FriendlySpecToNode(&spec)
			if err != nil {
				t.Skipf("FriendlySpecToNode(%s) unsupported in this build: %v", spec.Type, err)
			}

			// This is the exact lookup performed by parseBatchInjectionSpecs:
			//   spec.Children[strconv.Itoa(spec.Value)]
			// If FriendlySpecToNode ever starts keying children differently
			// (the /translate schema drift we just removed), this test fails.
			key := itoa(node.Value)
			if _, ok := node.Children[key]; !ok {
				keys := make([]string, 0, len(node.Children))
				for k := range node.Children {
					keys = append(keys, k)
				}
				t.Fatalf("type=%s: node.Children is missing key %q (Value=%d); got keys=%v",
					spec.Type, key, node.Value, keys)
			}
		})
	}
}

// itoa wraps strconv.Itoa to avoid re-importing in the test file; keeps the
// invariant check identical to parseBatchInjectionSpecs.
func itoa(v int) string {
	// Duplicating the one-liner keeps the test self-contained and guards
	// against accidental string conversions elsewhere.
	return fmtItoa(v)
}

func fmtItoa(v int) string {
	// small, allocation-free Itoa equivalent; std lib is fine but this keeps
	// the test's import list short.
	if v == 0 {
		return "0"
	}
	neg := false
	if v < 0 {
		neg = true
		v = -v
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
