package restkit_test

import (
	"testing"
	"time"

	rk "github.com/reststore/restkit"
)

func TestStringPtr(t *testing.T) {
	t.Run("returns pointer for non-empty string", func(t *testing.T) {
		result := rk.StringPtr("hello")
		if result == nil {
			t.Fatal("expected non-nil pointer")
		}
		if *result != "hello" {
			t.Errorf("expected 'hello', got %q", *result)
		}
	})

	t.Run("returns nil for empty string", func(t *testing.T) {
		result := rk.StringPtr("")
		if result != nil {
			t.Errorf("expected nil, got %v", *result)
		}
	})
}

func TestIntPtr(t *testing.T) {
	t.Run("returns pointer for valid int", func(t *testing.T) {
		result := rk.IntPtr("42")
		if result == nil {
			t.Fatal("expected non-nil pointer")
		}
		if *result != 42 {
			t.Errorf("expected 42, got %d", *result)
		}
	})

	t.Run("returns nil for empty string", func(t *testing.T) {
		result := rk.IntPtr("")
		if result != nil {
			t.Errorf("expected nil, got %v", *result)
		}
	})

	t.Run("returns nil for invalid string", func(t *testing.T) {
		result := rk.IntPtr("not-a-number")
		if result != nil {
			t.Errorf("expected nil, got %v", *result)
		}
	})
}

func TestInt64Ptr(t *testing.T) {
	t.Run("returns pointer for valid int64", func(t *testing.T) {
		result := rk.Int64Ptr("123456789")
		if result == nil {
			t.Fatal("expected non-nil pointer")
		}
		if *result != 123456789 {
			t.Errorf("expected 123456789, got %d", *result)
		}
	})

	t.Run("returns nil for empty string", func(t *testing.T) {
		result := rk.Int64Ptr("")
		if result != nil {
			t.Errorf("expected nil, got %v", *result)
		}
	})
}

func TestInt32Ptr(t *testing.T) {
	t.Run("returns pointer for valid int32", func(t *testing.T) {
		result := rk.Int32Ptr("100")
		if result == nil {
			t.Fatal("expected non-nil pointer")
		}
		if *result != 100 {
			t.Errorf("expected 100, got %d", *result)
		}
	})

	t.Run("returns nil for empty string", func(t *testing.T) {
		result := rk.Int32Ptr("")
		if result != nil {
			t.Errorf("expected nil, got %v", *result)
		}
	})
}

func TestBoolPtr(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"1", true},
		{"True", true},
		{"false", false},
		{"0", false},
		{"False", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := rk.BoolPtr(tt.input)
			if result == nil {
				t.Fatal("expected non-nil pointer")
			}
			if *result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, *result)
			}
		})
	}

	t.Run("returns nil for empty string", func(t *testing.T) {
		result := rk.BoolPtr("")
		if result != nil {
			t.Errorf("expected nil, got %v", *result)
		}
	})

	t.Run("returns nil for invalid string", func(t *testing.T) {
		result := rk.BoolPtr("maybe")
		if result != nil {
			t.Errorf("expected nil, got %v", *result)
		}
	})
}

func TestFloat64Ptr(t *testing.T) {
	t.Run("returns pointer for valid float", func(t *testing.T) {
		result := rk.Float64Ptr("3.14159")
		if result == nil {
			t.Fatal("expected non-nil pointer")
		}
		if *result != 3.14159 {
			t.Errorf("expected 3.14159, got %f", *result)
		}
	})

	t.Run("returns nil for empty string", func(t *testing.T) {
		result := rk.Float64Ptr("")
		if result != nil {
			t.Errorf("expected nil, got %v", *result)
		}
	})
}

func TestFloat32Ptr(t *testing.T) {
	t.Run("returns pointer for valid float", func(t *testing.T) {
		result := rk.Float32Ptr("2.5")
		if result == nil {
			t.Fatal("expected non-nil pointer")
		}
		if *result != 2.5 {
			t.Errorf("expected 2.5, got %f", *result)
		}
	})

	t.Run("returns nil for empty string", func(t *testing.T) {
		result := rk.Float32Ptr("")
		if result != nil {
			t.Errorf("expected nil, got %v", *result)
		}
	})
}

func TestTimePtr(t *testing.T) {
	t.Run("returns pointer for valid time", func(t *testing.T) {
		result := rk.TimePtr("2024-01-15", "2006-01-02")
		if result == nil {
			t.Fatal("expected non-nil pointer")
		}
		expected := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		if !result.Equal(expected) {
			t.Errorf("expected %v, got %v", expected, *result)
		}
	})

	t.Run("returns nil for empty string", func(t *testing.T) {
		result := rk.TimePtr("", "2006-01-02")
		if result != nil {
			t.Errorf("expected nil, got %v", *result)
		}
	})

	t.Run("returns nil for invalid time", func(t *testing.T) {
		result := rk.TimePtr("not-a-date", "2006-01-02")
		if result != nil {
			t.Errorf("expected nil, got %v", *result)
		}
	})
}

func TestTimePtrRFC3339(t *testing.T) {
	t.Run("parses RFC3339", func(t *testing.T) {
		result := rk.TimePtrRFC3339("2024-01-15T10:30:00Z")
		if result == nil {
			t.Fatal("expected non-nil pointer")
		}
		if result.Year() != 2024 || result.Month() != 1 || result.Day() != 15 {
			t.Errorf("expected 2024-01-15, got %v", *result)
		}
	})
}

func TestTimePtrDate(t *testing.T) {
	t.Run("parses date", func(t *testing.T) {
		result := rk.TimePtrDate("2024-01-15")
		if result == nil {
			t.Fatal("expected non-nil pointer")
		}
		if result.Year() != 2024 || result.Month() != 1 || result.Day() != 15 {
			t.Errorf("expected 2024-01-15, got %v", *result)
		}
	})
}
