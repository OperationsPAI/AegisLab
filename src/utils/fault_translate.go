package utils

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	chaos "github.com/OperationsPAI/chaos-experiment/handler"
)

// FieldDescription describes a single field in a fault spec struct.
type FieldDescription struct {
	Index       int    `json:"index"`
	Name        string `json:"name"`
	RangeMin    int    `json:"range_min"`
	RangeMax    int    `json:"range_max"`
	IsDynamic   bool   `json:"is_dynamic"`
	Description string `json:"description"`
}

// BuildSystemIndexMap builds a sorted name-to-index map from system types.
func BuildSystemIndexMap(systems []chaos.SystemType) map[string]int {
	result := make(map[string]int, len(systems))
	sorted := make([]string, 0, len(systems))
	for _, s := range systems {
		sorted = append(sorted, s.String())
	}
	sort.Strings(sorted)
	for i, name := range sorted {
		result[name] = i
	}
	return result
}

// BuildReverseTypeMap inverts ChaosTypeMap to map name -> ChaosType.
func BuildReverseTypeMap(typeMap map[chaos.ChaosType]string) map[string]chaos.ChaosType {
	result := make(map[string]chaos.ChaosType, len(typeMap))
	for k, v := range typeMap {
		result[v] = k
		// Also add lowercase variant for case-insensitive lookup.
		result[strings.ToLower(v)] = k
	}
	return result
}

// ExtractFieldDescriptions reflects on spec structs to extract field metadata.
func ExtractFieldDescriptions(specMap map[chaos.ChaosType]any) map[string][]FieldDescription {
	result := make(map[string][]FieldDescription, len(specMap))
	for ct, spec := range specMap {
		name := chaos.GetChaosTypeName(ct)
		if name == "" {
			continue
		}
		fields := extractFieldsFromStruct(spec)
		result[name] = fields
	}
	return result
}

// extractFieldsFromStruct extracts FieldDescription from a struct's reflect tags.
func extractFieldsFromStruct(v any) []FieldDescription {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	fields := make([]FieldDescription, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		fd := FieldDescription{
			Index: i,
			Name:  f.Name,
		}

		// Parse range tag: "min-max"
		if rangeTag := f.Tag.Get("range"); rangeTag != "" {
			parts := strings.SplitN(rangeTag, "-", 2)
			if len(parts) == 2 {
				fd.RangeMin, _ = strconv.Atoi(parts[0])
				fd.RangeMax, _ = strconv.Atoi(parts[1])
			}
		}

		// Parse dynamic tag
		if dynTag := f.Tag.Get("dynamic"); dynTag == "true" {
			fd.IsDynamic = true
		}

		// Parse description tag
		fd.Description = f.Tag.Get("description")

		fields = append(fields, fd)
	}
	return fields
}

// FaultSpecInput represents a human-readable fault specification for translation.
type FaultSpecInput struct {
	Type      string         `json:"type"`
	Namespace string         `json:"namespace"`
	Target    string         `json:"target"`
	Duration  string         `json:"duration"`
	Extra     map[string]any `json:"extra,omitempty"`
}

// TranslateFaultSpec translates a single FaultSpecInput to a chaos.Node.
// Returns the node, any warnings, and an error if the translation fails.
func TranslateFaultSpec(spec FaultSpecInput, reverseTypeMap map[string]chaos.ChaosType, systemIndexMap map[string]int) (*chaos.Node, []string, error) {
	var warnings []string

	// Resolve fault type name to ChaosType
	ct, ok := reverseTypeMap[spec.Type]
	if !ok {
		ct, ok = reverseTypeMap[strings.ToLower(spec.Type)]
		if !ok {
			return nil, warnings, fmt.Errorf("unknown fault type: %q", spec.Type)
		}
		warnings = append(warnings, fmt.Sprintf("fault type %q matched case-insensitively", spec.Type))
	}

	// Get the spec struct for this fault type to know the fields
	specStruct, exists := chaos.SpecMap[ct]
	if !exists {
		return nil, warnings, fmt.Errorf("no spec struct registered for fault type: %q", spec.Type)
	}

	fields := extractFieldsFromStruct(specStruct)

	// Build the root node
	root := &chaos.Node{
		Name:  spec.Type,
		Value: int(ct),
		Children: make(map[string]*chaos.Node),
	}

	// Set field values from the spec input
	for _, fd := range fields {
		child := &chaos.Node{
			Name:        fd.Name,
			Range:       []int{fd.RangeMin, fd.RangeMax},
			Description: fd.Description,
		}

		// Try to populate from extra map
		if spec.Extra != nil {
			if val, ok := spec.Extra[fd.Name]; ok {
				switch v := val.(type) {
				case float64:
					child.Value = int(v)
				case int:
					child.Value = v
				case string:
					n, err := strconv.Atoi(v)
					if err == nil {
						child.Value = n
					} else {
						warnings = append(warnings, fmt.Sprintf("field %q has non-integer value %q, defaulting to 0", fd.Name, v))
					}
				}
			}
		}

		// Handle well-known fields from top-level spec properties
		switch fd.Name {
		case "Duration":
			if spec.Duration != "" {
				dur := parseDurationToMinutes(spec.Duration)
				if dur > 0 {
					child.Value = dur
				} else {
					warnings = append(warnings, fmt.Sprintf("could not parse duration %q as minutes", spec.Duration))
				}
			}
		case "Namespace":
			if spec.Namespace != "" {
				if idx, ok := systemIndexMap[spec.Namespace]; ok {
					child.Value = idx
				} else {
					warnings = append(warnings, fmt.Sprintf("unknown namespace %q, using 0", spec.Namespace))
				}
			}
		}

		root.Children[fd.Name] = child
	}

	return root, warnings, nil
}

// parseDurationToMinutes parses a duration string like "60s", "5m", "1h" to minutes.
func parseDurationToMinutes(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	// Try plain integer (assume minutes)
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}

	unit := s[len(s)-1]
	numStr := s[:len(s)-1]
	n, err := strconv.Atoi(numStr)
	if err != nil {
		return 0
	}

	switch unit {
	case 's', 'S':
		return n / 60
	case 'm', 'M':
		return n
	case 'h', 'H':
		return n * 60
	default:
		return 0
	}
}
