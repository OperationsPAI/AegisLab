package utils

import (
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
)

func CloneMap(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))
	maps.Copy(dst, src)
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
	rawValue, ok := payload[key]
	if !ok {
		return nil, fmt.Errorf(errorMsgTemplate, key)
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
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	for i := range t.NumField() {
		field := t.Field(i)
		// 获取 JSON 标签名，如果没有则用字段名
		tag := field.Tag.Get("json")
		if tag == "" {
			tag = field.Name
		}
		result[tag] = v.Field(i).Interface()
	}

	return result
}
