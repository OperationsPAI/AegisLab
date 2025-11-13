package utils

import (
	"encoding/json"
	"fmt"
	maps0 "maps"
	"reflect"
	"strings"
)

// DeepMergeClone creates a new map that is the deep merge of multiple maps.
// Later maps take precedence over earlier maps for conflicting keys
func DeepMergeClone(maps ...map[string]any) map[string]any {
	result := make(map[string]any)
	for _, m := range maps {
		if m != nil {
			result = deepMerge(result, m)
		}
	}
	return result
}

func deepMerge(dst, src map[string]any) map[string]any {
	if dst == nil {
		dst = make(map[string]any)
	}

	for key, srcVal := range src {
		if dstVal, exists := dst[key]; exists {
			// If both values are maps, merge them recursively
			if dstMap, dstOk := dstVal.(map[string]any); dstOk {
				if srcMap, srcOk := srcVal.(map[string]any); srcOk {
					dst[key] = deepMerge(dstMap, srcMap)
					continue
				}
			}
		}
		dst[key] = srcVal
	}

	return dst
}

// GetPointerIntFromMap retrieves an integer value from a map by key and returns a pointer to it.
func GetPointerIntFromMap(payload map[string]any, key string) (*int, error) {
	val, ok := payload[key]
	if !ok {
		return nil, nil
	}

	if val == nil {
		return nil, nil
	}

	if intVal, ok := val.(int); ok {
		return &intVal, nil
	}

	if floatVal, ok := val.(float64); ok {
		if floatVal == float64(int(floatVal)) {
			intVal := int(floatVal)
			return &intVal, nil
		}
	}

	if strVal, ok := val.(string); ok {
		var intVal int
		_, err := fmt.Sscanf(strVal, "%d", &intVal)
		if err == nil {
			return &intVal, nil
		}
	}

	return nil, fmt.Errorf("value for key '%s' is not a valid integer type (got %T)", key, val)
}

// MapToStruct maps a nested map (or the entire map if key is empty) to a struct of type T
func MapToStruct[T any](payload map[string]any, key, errorMsgTemplate string) (*T, error) {
	var rawValue any
	if key == "" {
		rawValue = payload
	} else {
		var ok bool
		if rawValue, ok = payload[key]; !ok {
			return nil, fmt.Errorf(errorMsgTemplate, key)
		}
	}

	innerMap, ok := rawValue.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s: expected map[string]any, got %T", fmt.Sprintf(errorMsgTemplate, key), rawValue)
	}

	jsonData, err := json.Marshal(innerMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal intermediate map for key '%s': %w", key, err)
	}

	var result T
	if err := json.Unmarshal(jsonData, &result); err != nil {
		typeName := reflect.TypeOf(result).Name()
		if typeName == "" {
			typeName = reflect.TypeOf(result).String()
		}
		return nil, fmt.Errorf("failed to unmarshal JSON for key '%s' into type %s: %w", key, typeName, err)
	}

	return &result, nil
}

// MergeSimpleMaps creates a new map that is the shallow merge of multiple maps.
func MergeSimpleMaps[K comparable, V any](maps ...map[K]V) map[K]V {
	result := make(map[K]V)
	for _, m := range maps {
		maps0.Copy(result, m)
	}
	return result
}

// StructToMap converts a struct to a map[string]any using reflection
func StructToMap(obj any) map[string]any {
	result := make(map[string]any)

	v := reflect.ValueOf(obj)
	t := reflect.TypeOf(obj)

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return result
		}

		v = v.Elem()
		t = t.Elem()
	}

	if v.Kind() != reflect.Struct {
		return result
	}

	for i := range t.NumField() {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if !fieldValue.CanInterface() {
			continue
		}

		tag := field.Tag.Get("json")
		if tag == "" {
			tag = field.Name
		}

		if commaIdx := strings.Index(tag, ","); commaIdx != -1 {
			tag = tag[:commaIdx]
		}

		if tag == "-" {
			continue
		}

		result[tag] = fieldValue.Interface()
	}

	return result
}
