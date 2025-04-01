package utils

import (
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
