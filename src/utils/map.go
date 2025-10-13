package utils

import (
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"strings"
)

func CloneMap(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))
	maps.Copy(dst, src)
	return dst
}

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

func GetMapField(m map[string]any, keys ...string) (string, bool) {
	current := m
	for i, key := range keys {
		val, exists := current[key]
		if !exists {
			return "", false
		}

		if i == len(keys)-1 {
			strVal, ok := val.(string)
			return strVal, ok
		}

		nextMap, ok := val.(map[string]any)
		if !ok {
			return "", false
		}
		current = nextMap
	}
	return "", false
}

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
