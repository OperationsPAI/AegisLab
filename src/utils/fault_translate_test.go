package utils

import (
	"sort"
	"testing"

	chaos "github.com/OperationsPAI/chaos-experiment/handler"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// TestBuildReverseTypeMap
// ---------------------------------------------------------------------------

func TestBuildReverseTypeMap(t *testing.T) {
	reverseMap := BuildReverseTypeMap(chaos.ChaosTypeMap)

	t.Run("map is non-empty", func(t *testing.T) {
		assert.NotEmpty(t, reverseMap, "reverse map should contain entries")
	})

	t.Run("CPUStress maps to correct ChaosType", func(t *testing.T) {
		ct, ok := reverseMap["CPUStress"]
		assert.True(t, ok, "CPUStress should be present in reverse map")
		assert.Equal(t, chaos.CPUStress, ct, "CPUStress should map to chaos.CPUStress (4)")
	})

	t.Run("PodKill maps to correct ChaosType", func(t *testing.T) {
		ct, ok := reverseMap["PodKill"]
		assert.True(t, ok, "PodKill should be present in reverse map")
		assert.Equal(t, chaos.PodKill, ct, "PodKill should map to chaos.PodKill (0)")
	})

	t.Run("all ChaosTypeMap entries have reverse entries", func(t *testing.T) {
		for ct, name := range chaos.ChaosTypeMap {
			got, ok := reverseMap[name]
			assert.True(t, ok, "reverse map should contain key %q", name)
			assert.Equal(t, ct, got, "reverse map[%q] should equal %d", name, ct)
		}
	})

	t.Run("map size matches ChaosTypeMap", func(t *testing.T) {
		assert.Equal(t, len(chaos.ChaosTypeMap), len(reverseMap),
			"reverse map should have same number of entries as ChaosTypeMap")
	})
}

// helper to avoid importing strings just for tests
func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

// ---------------------------------------------------------------------------
// TestBuildSystemIndexMap
// ---------------------------------------------------------------------------

func TestBuildSystemIndexMap(t *testing.T) {
	allSystems := chaos.GetAllSystemTypes()
	systemMap := BuildSystemIndexMap(allSystems)

	t.Run("map is non-empty", func(t *testing.T) {
		assert.NotEmpty(t, systemMap, "system index map should not be empty")
	})

	t.Run("all system types are present", func(t *testing.T) {
		for _, sys := range allSystems {
			_, ok := systemMap[sys.String()]
			assert.True(t, ok, "system %q should be present in the map", sys.String())
		}
	})

	t.Run("indices are 0-based and contiguous", func(t *testing.T) {
		n := len(systemMap)
		seen := make(map[int]bool, n)
		for name, idx := range systemMap {
			assert.GreaterOrEqual(t, idx, 0, "index for %q should be >= 0", name)
			assert.Less(t, idx, n, "index for %q should be < %d", name, n)
			assert.False(t, seen[idx], "index %d should not be duplicated", idx)
			seen[idx] = true
		}
		// All indices 0..n-1 should be present
		for i := 0; i < n; i++ {
			assert.True(t, seen[i], "index %d should be assigned", i)
		}
	})

	t.Run("sorted alphabetically", func(t *testing.T) {
		// Collect names sorted by their index
		type entry struct {
			name string
			idx  int
		}
		entries := make([]entry, 0, len(systemMap))
		for name, idx := range systemMap {
			entries = append(entries, entry{name, idx})
		}
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].idx < entries[j].idx
		})

		// The names in index order should be alphabetically sorted
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.name
		}
		assert.True(t, sort.StringsAreSorted(names),
			"system names ordered by index should be alphabetically sorted, got: %v", names)
	})

	t.Run("map size matches input", func(t *testing.T) {
		// Deduplicate input to get expected size
		unique := make(map[string]struct{})
		for _, sys := range allSystems {
			unique[sys.String()] = struct{}{}
		}
		assert.Equal(t, len(unique), len(systemMap),
			"map size should match number of unique system types")
	})
}

// ---------------------------------------------------------------------------
// TestExtractFieldDescriptions
// ---------------------------------------------------------------------------

func TestExtractFieldDescriptions(t *testing.T) {
	descriptions := ExtractFieldDescriptions(chaos.SpecMap)

	t.Run("result is non-empty", func(t *testing.T) {
		assert.NotEmpty(t, descriptions, "field descriptions should not be empty")
	})

	t.Run("CPUStress has expected fields", func(t *testing.T) {
		cpuFields, ok := descriptions["CPUStress"]
		assert.True(t, ok, "CPUStress should have field descriptions")
		assert.NotEmpty(t, cpuFields, "CPUStress should have at least one field")

		// Build a name lookup for convenience
		fieldByName := make(map[string]FieldDescription, len(cpuFields))
		for _, f := range cpuFields {
			fieldByName[f.Name] = f
		}

		expectedFields := []string{"Duration", "Namespace", "CPULoad", "CPUWorker"}
		for _, name := range expectedFields {
			_, exists := fieldByName[name]
			assert.True(t, exists, "CPUStress should contain field %q", name)
		}
	})

	t.Run("CPULoad range is parsed correctly", func(t *testing.T) {
		cpuFields, ok := descriptions["CPUStress"]
		if !ok {
			t.Skip("CPUStress not found in descriptions")
		}

		var cpuLoadField *FieldDescription
		for i, f := range cpuFields {
			if f.Name == "CPULoad" {
				cpuLoadField = &cpuFields[i]
				break
			}
		}

		if assert.NotNil(t, cpuLoadField, "CPULoad field should exist") {
			assert.Equal(t, 1, cpuLoadField.RangeMin, "CPULoad range min should be 1")
			assert.Equal(t, 100, cpuLoadField.RangeMax, "CPULoad range max should be 100")
		}
	})

	t.Run("dynamic flag is set correctly", func(t *testing.T) {
		cpuFields, ok := descriptions["CPUStress"]
		if !ok {
			t.Skip("CPUStress not found in descriptions")
		}

		fieldByName := make(map[string]FieldDescription, len(cpuFields))
		for _, f := range cpuFields {
			fieldByName[f.Name] = f
		}

		if ns, ok := fieldByName["Namespace"]; ok {
			assert.True(t, ns.IsDynamic, "Namespace should be dynamic")
		}
		if dur, ok := fieldByName["Duration"]; ok {
			assert.False(t, dur.IsDynamic, "Duration should not be dynamic")
		}
	})

	t.Run("descriptions are populated", func(t *testing.T) {
		// At least some fields across all types should have non-empty descriptions
		hasDescription := false
		for _, fields := range descriptions {
			for _, f := range fields {
				if f.Description != "" {
					hasDescription = true
					break
				}
			}
			if hasDescription {
				break
			}
		}
		assert.True(t, hasDescription, "at least some fields should have non-empty descriptions")
	})

	t.Run("all spec map entries produce descriptions", func(t *testing.T) {
		for ct := range chaos.SpecMap {
			name := chaos.GetChaosTypeName(ct)
			if name == "" {
				continue // skip unregistered types
			}
			_, ok := descriptions[name]
			assert.True(t, ok, "spec map entry %q (type %d) should have field descriptions", name, ct)
		}
	})

	t.Run("field indices are non-negative", func(t *testing.T) {
		for name, fields := range descriptions {
			for _, f := range fields {
				assert.GreaterOrEqual(t, f.Index, 0,
					"field %q in %q should have non-negative index", f.Name, name)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// TestTranslateFaultSpec
// ---------------------------------------------------------------------------

func TestTranslateFaultSpec(t *testing.T) {
	reverseTypeMap := BuildReverseTypeMap(chaos.ChaosTypeMap)
	systemMap := BuildSystemIndexMap(chaos.GetAllSystemTypes())

	t.Run("valid CPUStress spec translates successfully", func(t *testing.T) {
		spec := FaultSpecInput{
			Type:      "CPUStress",
			Namespace: pickFirstSystem(systemMap),
			Target:    "frontend",
			Duration:  "5m",
			Extra: map[string]any{
				"CPULoad":   80,
				"CPUWorker": 2,
			},
		}

		node, warnings, err := TranslateFaultSpec(spec, reverseTypeMap, systemMap)
		assert.NoError(t, err, "translation should succeed")
		assert.NotNil(t, node, "node should not be nil")
		assert.Equal(t, "CPUStress", node.Name)
		assert.Equal(t, int(chaos.CPUStress), node.Value)
		// Warnings may or may not be present; just ensure no error
		_ = warnings
	})

	t.Run("unknown fault type returns error", func(t *testing.T) {
		spec := FaultSpecInput{
			Type:     "NonExistentFaultType",
			Duration: "5m",
		}

		node, _, err := TranslateFaultSpec(spec, reverseTypeMap, systemMap)
		assert.Error(t, err, "unknown fault type should return error")
		assert.Nil(t, node, "node should be nil on error")
		assert.Contains(t, err.Error(), "unknown fault type")
	})

	t.Run("case-insensitive fault type match produces warning", func(t *testing.T) {
		spec := FaultSpecInput{
			Type:     "cpustress", // lowercase
			Duration: "5m",
		}

		node, warnings, err := TranslateFaultSpec(spec, reverseTypeMap, systemMap)
		assert.NoError(t, err, "case-insensitive match should succeed")
		assert.NotNil(t, node)
		assert.NotEmpty(t, warnings, "should produce a case-insensitive warning")
	})

	t.Run("duration parsing", func(t *testing.T) {
		tests := []struct {
			name     string
			duration string
		}{
			{"seconds", "60s"},
			{"minutes", "5m"},
			{"hours", "1h"},
			{"plain integer", "10"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				spec := FaultSpecInput{
					Type:     "CPUStress",
					Duration: tt.duration,
				}
				node, _, err := TranslateFaultSpec(spec, reverseTypeMap, systemMap)
				assert.NoError(t, err)
				assert.NotNil(t, node)
			})
		}
	})
}

// ---------------------------------------------------------------------------
// TestParseDurationToMinutes
// ---------------------------------------------------------------------------

func TestParseDurationToMinutes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{name: "empty string", input: "", expected: 0},
		{name: "plain integer", input: "10", expected: 10},
		{name: "seconds", input: "120s", expected: 2},
		{name: "minutes", input: "5m", expected: 5},
		{name: "hours", input: "2h", expected: 120},
		{name: "seconds less than a minute", input: "30s", expected: 0},
		{name: "uppercase S", input: "120S", expected: 2},
		{name: "uppercase M", input: "5M", expected: 5},
		{name: "uppercase H", input: "1H", expected: 60},
		{name: "invalid format", input: "abc", expected: 0},
		{name: "whitespace", input: "  5m  ", expected: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDurationToMinutes(tt.input)
			assert.Equal(t, tt.expected, result, "parseDurationToMinutes(%q)", tt.input)
		})
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// pickFirstSystem returns the first system name from the map (alphabetically first).
func pickFirstSystem(systemMap map[string]int) string {
	for name, idx := range systemMap {
		if idx == 0 {
			return name
		}
	}
	return ""
}
