package utils

import "time"

func BoolPtr(b bool) *bool {
	return &b
}

func GetIntValue(ptr *int, defaultValue int) int {
	if ptr == nil {
		return defaultValue
	}

	return *ptr
}

func GetStringValue(ptr *string, defaultValue string) string {
	if ptr == nil {
		return defaultValue
	}

	return *ptr
}

func GetTimeValue(ptr *time.Time, defaultValue time.Time) time.Time {
	if ptr == nil {
		return defaultValue
	}

	return *ptr
}
