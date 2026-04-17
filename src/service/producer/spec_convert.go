package producer

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"aegis/dto"

	chaos "github.com/OperationsPAI/chaos-experiment/handler"
)

// chaosTypeNameToIndex maps human-readable fault type names to their ChaosType (= InjectionConf field index).
// Populated once at init from chaos.ChaosTypeMap.
var chaosTypeNameToIndex map[string]int

func init() {
	chaosTypeNameToIndex = make(map[string]int, len(chaos.ChaosTypeMap)*2)
	for ct, name := range chaos.ChaosTypeMap {
		chaosTypeNameToIndex[name] = int(ct)
		chaosTypeNameToIndex[strings.ToLower(name)] = int(ct)
	}
}

// FriendlySpecToNode converts a human-readable FriendlyFaultSpec into a chaos.Node tree
// that is compatible with the existing parseBatchInjectionSpecs pipeline.
func FriendlySpecToNode(spec *dto.FriendlyFaultSpec) (chaos.Node, error) {
	if spec.Type == "" {
		return chaos.Node{}, fmt.Errorf("fault type is required")
	}

	// Resolve fault type name to ChaosType index
	typeIdx, ok := chaosTypeNameToIndex[spec.Type]
	if !ok {
		typeIdx, ok = chaosTypeNameToIndex[strings.ToLower(spec.Type)]
		if !ok {
			available := make([]string, 0, len(chaos.ChaosTypeMap))
			for _, name := range chaos.ChaosTypeMap {
				available = append(available, name)
			}
			return chaos.Node{}, fmt.Errorf("unknown fault type %q, available: %v", spec.Type, available)
		}
	}

	// Parse duration to minutes (the chaos library uses integer minutes)
	durationMinutes, err := parseDurationToMinutes(spec.Duration)
	if err != nil {
		return chaos.Node{}, fmt.Errorf("invalid duration %q: %w", spec.Duration, err)
	}

	// Resolve namespace string to its index in chaos.NamespacePrefixs
	namespaceIdx, err := resolveNamespaceIndex(spec.Namespace)
	if err != nil {
		return chaos.Node{}, fmt.Errorf("failed to resolve namespace %q: %w", spec.Namespace, err)
	}

	// Resolve target to a numeric index.
	// The target field maps to the 3rd field (index 2) of the spec struct,
	// which is ContainerIdx, AppIdx, etc. depending on the fault type.
	// Target must be a numeric string or empty (defaults to 0).
	targetIdx, err := resolveTargetIndex(spec.Target)
	if err != nil {
		return chaos.Node{}, fmt.Errorf("failed to resolve target %q: %w", spec.Target, err)
	}

	// Build the inner children map — these map to spec struct field indices.
	// Field 0 = Duration, Field 1 = Namespace, Field 2 = ContainerIdx/AppIdx/etc.
	specChildren := map[string]*chaos.Node{
		"0": {Value: durationMinutes}, // Duration
		"1": {Value: namespaceIdx},    // Namespace
		"2": {Value: targetIdx},       // ContainerIdx / AppIdx / etc.
	}

	// Map additional params to their corresponding field indices (fields 3+)
	if len(spec.Params) > 0 {
		specType := getSpecType(typeIdx)
		if specType != nil {
			if err := mapParamsToFieldIndices(spec.Params, specType, specChildren); err != nil {
				return chaos.Node{}, fmt.Errorf("failed to map params: %w", err)
			}
		}
	}

	// Build the chaos.Node tree structure expected by parseBatchInjectionSpecs:
	//   {Value: <type_idx>, Children: {"<type_idx>": {Children: {"0": ..., "1": ..., "2": ...}}}}
	typeIdxStr := strconv.Itoa(typeIdx)
	node := chaos.Node{
		Value: typeIdx,
		Children: map[string]*chaos.Node{
			typeIdxStr: {
				Children: specChildren,
			},
		},
	}

	return node, nil
}

// parseDurationToMinutes converts a duration string (e.g., "60s", "5m", "1h") to integer minutes.
// Also accepts plain integer strings interpreted as minutes.
func parseDurationToMinutes(duration string) (int, error) {
	if duration == "" {
		return 0, fmt.Errorf("duration is required")
	}

	// Try Go duration format first (e.g., "60s", "5m")
	d, err := time.ParseDuration(duration)
	if err == nil {
		minutes := int(math.Ceil(d.Minutes()))
		if minutes < 1 {
			minutes = 1
		}
		return minutes, nil
	}

	// Fall back to plain integer (interpreted as minutes)
	if mins, err2 := strconv.Atoi(duration); err2 == nil && mins > 0 {
		return mins, nil
	}

	return 0, fmt.Errorf("cannot parse duration %q: expected Go duration (e.g., \"60s\", \"5m\") or integer minutes", duration)
}

// resolveNamespaceIndex maps a namespace prefix string to its index in chaos.NamespacePrefixs.
func resolveNamespaceIndex(namespace string) (int, error) {
	if namespace == "" {
		return 0, nil
	}

	// Exact match only
	for idx, prefix := range chaos.NamespacePrefixs {
		if prefix == namespace {
			return idx, nil
		}
	}

	return 0, fmt.Errorf("namespace %q not found in registered prefixes: %v", namespace, chaos.NamespacePrefixs)
}

// resolveTargetIndex resolves the target field to a numeric index.
// If target is empty, defaults to 0. If numeric, parses directly.
// Non-numeric non-empty targets return an error.
func resolveTargetIndex(target string) (int, error) {
	if target == "" {
		return 0, nil
	}

	// Try numeric index first
	if idx, err := strconv.Atoi(target); err == nil {
		return idx, nil
	}

	// Non-numeric, non-empty target is an error — users must use numeric indices.
	return 0, fmt.Errorf("target %q is not a valid numeric index; use 'aegisctl inject metadata' to look up numeric indices", target)
}

// getSpecType returns the zero-value spec struct for a given ChaosType index.
func getSpecType(typeIdx int) any {
	ct := chaos.ChaosType(typeIdx)
	if spec, ok := chaos.SpecMap[ct]; ok {
		return spec
	}
	return nil
}

// mapParamsToFieldIndices maps user-provided param names to spec struct field indices.
// Fields 0-2 are already populated (Duration, Namespace, ContainerIdx/AppIdx).
// This handles fields 3+ (e.g., CPULoad, CPUWorker for CPUStressChaosSpec).
func mapParamsToFieldIndices(params map[string]any, specType any, children map[string]*chaos.Node) error {
	rt := reflect.TypeOf(specType)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	// Build name → field index map for fields 3+
	nameToIdx := make(map[string]int, rt.NumField())
	for i := 3; i < rt.NumField(); i++ {
		field := rt.Field(i)
		// Skip the NamespaceTarget field — it's set internally
		if field.Name == "NamespaceTarget" {
			continue
		}
		nameToIdx[field.Name] = i
		nameToIdx[strings.ToLower(field.Name)] = i
		nameToIdx[toSnakeCase(field.Name)] = i
	}

	for key, val := range params {
		idx, ok := nameToIdx[key]
		if !ok {
			idx, ok = nameToIdx[strings.ToLower(key)]
		}
		if !ok {
			available := make([]string, 0, len(nameToIdx))
			seen := make(map[int]bool)
			for name, fieldIdx := range nameToIdx {
				if !seen[fieldIdx] {
					seen[fieldIdx] = true
					available = append(available, name)
				}
			}
			return fmt.Errorf("unknown param %q, available fields: %v", key, available)
		}

		intVal, err := toInt(val)
		if err != nil {
			return fmt.Errorf("param %q: %w", key, err)
		}

		children[strconv.Itoa(idx)] = &chaos.Node{Value: intVal}
	}

	return nil
}

// toSnakeCase converts CamelCase to snake_case, handling consecutive uppercase runs.
// e.g., "CPULoad" → "cpu_load", "CPUWorker" → "cpu_worker", "MemorySize" → "memory_size".
func toSnakeCase(s string) string {
	var result strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				prev := runes[i-1]
				// Insert underscore on lowercase→uppercase transition,
				// or when this uppercase letter is followed by a lowercase letter
				// (end of an uppercase run, e.g., the 'L' in "CPULoad").
				if prev >= 'a' && prev <= 'z' {
					result.WriteByte('_')
				} else if prev >= 'A' && prev <= 'Z' && i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z' {
					result.WriteByte('_')
				}
			}
			result.WriteRune(r + ('a' - 'A'))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// toInt converts various numeric types to int.
func toInt(v any) (int, error) {
	switch val := v.(type) {
	case int:
		return val, nil
	case float64:
		return int(val), nil
	case float32:
		return int(val), nil
	case string:
		return strconv.Atoi(val)
	case json.Number:
		n, err := val.Int64()
		return int(n), err
	default:
		return 0, fmt.Errorf("cannot convert %T to int", v)
	}
}
