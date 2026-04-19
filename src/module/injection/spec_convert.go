package injection

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

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
// compatible with parseBatchInjectionSpecs.
func FriendlySpecToNode(spec *FriendlyFaultSpec) (chaos.Node, error) {
	if spec.Type == "" {
		return chaos.Node{}, fmt.Errorf("fault type is required")
	}

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

	durationMinutes, err := parseDurationToMinutes(spec.Duration)
	if err != nil {
		return chaos.Node{}, fmt.Errorf("invalid duration %q: %w", spec.Duration, err)
	}

	namespaceIdx, err := resolveNamespaceIndex(spec.Namespace)
	if err != nil {
		return chaos.Node{}, fmt.Errorf("failed to resolve namespace %q: %w", spec.Namespace, err)
	}

	targetIdx, err := resolveTargetIndex(spec.Target)
	if err != nil {
		return chaos.Node{}, fmt.Errorf("failed to resolve target %q: %w", spec.Target, err)
	}

	// Field 0 = Duration, Field 1 = Namespace, Field 2 = ContainerIdx/AppIdx/etc.
	specChildren := map[string]*chaos.Node{
		"0": {Value: durationMinutes},
		"1": {Value: namespaceIdx},
		"2": {Value: targetIdx},
	}

	if len(spec.Params) > 0 {
		specType := getSpecType(typeIdx)
		if specType != nil {
			if err := mapParamsToFieldIndices(spec.Params, specType, specChildren); err != nil {
				return chaos.Node{}, fmt.Errorf("failed to map params: %w", err)
			}
		}
	}

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

// parseDurationToMinutes converts "60s" / "5m" / "1h" / plain integer to minutes.
func parseDurationToMinutes(duration string) (int, error) {
	if duration == "" {
		return 0, fmt.Errorf("duration is required")
	}

	d, err := time.ParseDuration(duration)
	if err == nil {
		minutes := int(math.Ceil(d.Minutes()))
		if minutes < 1 {
			minutes = 1
		}
		return minutes, nil
	}

	if mins, err2 := strconv.Atoi(duration); err2 == nil && mins > 0 {
		return mins, nil
	}

	return 0, fmt.Errorf("cannot parse duration %q: expected Go duration (e.g., \"60s\", \"5m\") or integer minutes", duration)
}

// resolveNamespaceIndex accepts a namespace field from FriendlyFaultSpec.
// Under chaos-experiment v1.0.1+, namespace resolution moved to per-system
// registrations (GetNamespaceByIndex), and the old package-level
// chaos.NamespacePrefixs slice no longer exists. The backend's downstream
// pipeline owns name→index resolution, so here we accept numeric strings
// directly and fall back to 0 for names (best-effort, matching
// resolveTargetIndex behavior).
func resolveNamespaceIndex(namespace string) (int, error) {
	if namespace == "" {
		return 0, nil
	}
	if idx, err := strconv.Atoi(namespace); err == nil {
		return idx, nil
	}
	return 0, nil
}

// resolveTargetIndex turns numeric strings into indices; non-numeric names default to 0.
// Full name→index resolution requires K8s state (internal to chaos-experiment).
func resolveTargetIndex(target string) (int, error) {
	if target == "" {
		return 0, nil
	}

	if idx, err := strconv.Atoi(target); err == nil {
		return idx, nil
	}

	return 0, nil
}

// getSpecType returns the zero-value spec struct for a given ChaosType index.
func getSpecType(typeIdx int) any {
	ct := chaos.ChaosType(typeIdx)
	if spec, ok := chaos.SpecMap[ct]; ok {
		return spec
	}
	return nil
}

// mapParamsToFieldIndices maps param names to spec struct field indices (3+).
// Fields 0-2 are already populated (Duration, Namespace, Target).
func mapParamsToFieldIndices(params map[string]any, specType any, children map[string]*chaos.Node) error {
	rt := reflect.TypeOf(specType)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	nameToIdx := make(map[string]int, rt.NumField())
	for i := 3; i < rt.NumField(); i++ {
		field := rt.Field(i)
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
			continue
		}

		intVal, err := toInt(val)
		if err != nil {
			return fmt.Errorf("param %q: %w", key, err)
		}

		children[strconv.Itoa(idx)] = &chaos.Node{Value: intVal}
	}

	return nil
}

// toSnakeCase converts CamelCase → snake_case ("CPULoad" → "c_p_u_load").
// Acceptable for key lookup (we also try the lowercase and exact forms).
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(r + ('a' - 'A'))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// toInt converts numeric types to int.
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
