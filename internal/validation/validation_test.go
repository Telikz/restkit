package validation

import (
	"context"
	"testing"
)

func TestValidateStruct(t *testing.T) {
	t.Run("returns empty result", func(t *testing.T) {
		type TestStruct struct {
			Name  string `validate:"required"`
			Email string `validate:"required,email"`
		}
		obj := TestStruct{Name: "John", Email: "john@example.com"}
		result := ValidateStruct(context.Background(), obj)
		if result.HasErrors() {
			t.Errorf("expected no errors, got %v", result.Errors)
		}
	})

	t.Run("returns empty for invalid struct", func(t *testing.T) {
		type TestStruct struct {
			Name  string `validate:"required"`
			Email string `validate:"required,email"`
		}
		obj := TestStruct{}
		result := ValidateStruct(context.Background(), obj)
		if result.HasErrors() {
			t.Errorf("expected no errors for no-op, got %v", result.Errors)
		}
	})

	t.Run("handles nil input", func(t *testing.T) {
		result := ValidateStruct(context.Background(), nil)
		if result.HasErrors() {
			t.Errorf("expected no errors for nil, got %v", result.Errors)
		}
	})
}
