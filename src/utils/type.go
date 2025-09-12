package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func ConvertToType[T any](value any) (T, error) {
	var zero T

	if result, ok := value.(T); ok {
		return result, nil
	}

	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return zero, fmt.Errorf("failed to marshal to JSON: %v", err)
	}

	var result T
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return zero, fmt.Errorf("failed to unmarshal from JSON: %v", err)
	}

	return result, nil
}

func GetTypeName(obj any) string {
	objType := reflect.TypeOf(obj)
	if objType != nil && objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}

	// Handle pointer type
	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}

	// Handle anonymous type or invalid type
	objName := "item"
	if objType != nil {
		objName = objType.Name()
	}

	return objName
}
func Must(err error) {
	if err != nil {
		panic(err)
	}
}
