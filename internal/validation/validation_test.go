package validation

import (
	"context"
	"testing"
)

func TestValidate(t *testing.T) {
	t.Run("returns empty result", func(t *testing.T) {
		type TestStruct struct {
			Name  string `validate:"required"`
			Email string `validate:"required,email"`
		}
		obj := TestStruct{Name: "John", Email: "john@example.com"}
		result := Validate(context.Background(), obj)
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
		result := Validate(context.Background(), obj)
		if result.HasErrors() {
			t.Errorf("expected no errors for no-op, got %v", result.Errors)
		}
	})

	t.Run("handles nil input", func(t *testing.T) {
		result := Validate(context.Background(), nil)
		if result.HasErrors() {
			t.Errorf("expected no errors for nil, got %v", result.Errors)
		}
	})
}
