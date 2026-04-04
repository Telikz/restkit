package restkit

import (
	"strconv"
	"time"
)

// StringPtr converts a string to *string.
// Returns nil if the input is empty.
func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// IntPtr converts a string to *int.
// Returns nil if the input is empty or invalid.
func IntPtr(s string) *int {
	if s == "" {
		return nil
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &val
}

// Int64Ptr converts a string to *int64.
// Returns nil if the input is empty or invalid.
func Int64Ptr(s string) *int64 {
	if s == "" {
		return nil
	}
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil
	}
	return &val
}

// Int32Ptr converts a string to *int32.
// Returns nil if the input is empty or invalid.
func Int32Ptr(s string) *int32 {
	if s == "" {
		return nil
	}
	val, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return nil
	}
	v := int32(val)
	return &v
}

// BoolPtr converts a string to *bool.
// Returns nil if the input is empty or invalid.
func BoolPtr(s string) *bool {
	if s == "" {
		return nil
	}
	val, err := strconv.ParseBool(s)
	if err != nil {
		return nil
	}
	return &val
}

// Float64Ptr converts a string to *float64.
// Returns nil if the input is empty or invalid.
func Float64Ptr(s string) *float64 {
	if s == "" {
		return nil
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &val
}

// Float32Ptr converts a string to *float32.
// Returns nil if the input is empty or invalid.
func Float32Ptr(s string) *float32 {
	if s == "" {
		return nil
	}
	val, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return nil
	}
	v := float32(val)
	return &v
}

// TimePtr parses a string using the given layout and returns *time.Time.
// Returns nil if the input is empty or invalid.
// Common layouts: time.RFC3339, "2006-01-02", "2006-01-02 15:04:05"
func TimePtr(s, layout string) *time.Time {
	if s == "" {
		return nil
	}
	val, err := time.Parse(layout, s)
	if err != nil {
		return nil
	}
	return &val
}

// TimePtrRFC3339 parses a string as RFC3339 and returns *time.Time.
// Returns nil if the input is empty or invalid.
func TimePtrRFC3339(s string) *time.Time {
	return TimePtr(s, time.RFC3339)
}

// TimePtrDate parses a string as date (YYYY-MM-DD) and returns *time.Time.
// Returns nil if the input is empty or invalid.
func TimePtrDate(s string) *time.Time {
	return TimePtr(s, "2006-01-02")
}
