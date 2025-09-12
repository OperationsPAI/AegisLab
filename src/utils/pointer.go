package utils

import "time"

func BoolPtr(b bool) *bool {
	return &b
}

func IntPtr(i int) *int {
	return &i
}

func GetBoolValue(ptr *bool, defaultValue bool) bool {
	if ptr == nil {
		return defaultValue
	}

	return *ptr
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

func TimePtr(t time.Time) *time.Time {
	return &t
}

func GetTimePtr(ptr *time.Time, defaultValue time.Time) *time.Time {
	if ptr == nil {
		return &defaultValue
	}

	return ptr
}
