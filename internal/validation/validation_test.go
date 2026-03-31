package validation

import (
	"context"
	"strings"
	"testing"
)

func TestValidateStruct(t *testing.T) {
	t.Run("valid struct", func(t *testing.T) {
		type ValidUser struct {
			Name  string `validate:"required"`
			Email string `validate:"required,email"`
			Age   int    `validate:"gte=0,lte=150"`
		}

		user := ValidUser{
			Name:  "John Doe",
			Email: "john@example.com",
			Age:   30,
		}

		result := ValidateStruct(context.Background(), user)

		if result.HasErrors() {
			t.Errorf("expected no validation errors, got %v", result.Errors)
		}
	})

	t.Run("missing required field", func(t *testing.T) {
		type RequiredTest struct {
			Name string `validate:"required"`
		}

		obj := RequiredTest{Name: ""}
		result := ValidateStruct(context.Background(), obj)

		if !result.HasErrors() {
			t.Error("expected validation errors for missing required field")
		}

		if result.Status != 422 {
			t.Errorf("expected status 422, got %d", result.Status)
		}

		if result.Code != "validation" {
			t.Errorf("expected code 'validation', got '%s'", result.Code)
		}

		found := false
		for _, err := range result.Errors {
			if err.Field == "name" && strings.Contains(err.Message, "required") {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("expected error for 'name' field, got %v", result.Errors)
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		type EmailTest struct {
			Email string `validate:"email"`
		}

		obj := EmailTest{Email: "not-an-email"}
		result := ValidateStruct(context.Background(), obj)

		if !result.HasErrors() {
			t.Error("expected validation error for invalid email")
		}

		found := false
		for _, err := range result.Errors {
			if err.Field == "email" && strings.Contains(err.Message, "email") {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("expected email validation error, got %v", result.Errors)
		}
	})

	t.Run("min length validation", func(t *testing.T) {
		type MinLength struct {
			Password string `validate:"min=8"`
		}

		obj := MinLength{Password: "short"}
		result := ValidateStruct(context.Background(), obj)

		if !result.HasErrors() {
			t.Error("expected validation error for min length")
		}

		found := false
		for _, err := range result.Errors {
			if err.Field == "password" && strings.Contains(err.Message, "at least") {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("expected min length error, got %v", result.Errors)
		}
	})

	t.Run("max length validation", func(t *testing.T) {
		type MaxLength struct {
			Username string `validate:"max=10"`
		}

		obj := MaxLength{Username: "thisisaverylongusername"}
		result := ValidateStruct(context.Background(), obj)

		if !result.HasErrors() {
			t.Error("expected validation error for max length")
		}

		found := false
		for _, err := range result.Errors {
			if err.Field == "username" && strings.Contains(err.Message, "at most") {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("expected max length error, got %v", result.Errors)
		}
	})

	t.Run("exact length validation", func(t *testing.T) {
		type ExactLength struct {
			Code string `validate:"len=4"`
		}

		obj := ExactLength{Code: "12345"}
		result := ValidateStruct(context.Background(), obj)

		if !result.HasErrors() {
			t.Error("expected validation error for exact length")
		}

		found := false
		for _, err := range result.Errors {
			if err.Field == "code" && strings.Contains(err.Message, "characters long") {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("expected exact length error, got %v", result.Errors)
		}
	})

	t.Run("gte validation", func(t *testing.T) {
		type GteTest struct {
			Score int `validate:"gte=0"`
		}

		obj := GteTest{Score: -5}
		result := ValidateStruct(context.Background(), obj)

		if !result.HasErrors() {
			t.Error("expected validation error for gte")
		}

		found := false
		for _, err := range result.Errors {
			if err.Field == "score" && strings.Contains(err.Message, "greater than or equal") {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("expected gte error, got %v", result.Errors)
		}
	})

	t.Run("lte validation", func(t *testing.T) {
		type LteTest struct {
			Percentage int `validate:"lte=100"`
		}

		obj := LteTest{Percentage: 150}
		result := ValidateStruct(context.Background(), obj)

		if !result.HasErrors() {
			t.Error("expected validation error for lte")
		}

		found := false
		for _, err := range result.Errors {
			if err.Field == "percentage" && strings.Contains(err.Message, "less than or equal") {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("expected lte error, got %v", result.Errors)
		}
	})

	t.Run("gt validation", func(t *testing.T) {
		type GtTest struct {
			Count int `validate:"gt=0"`
		}

		obj := GtTest{Count: 0}
		result := ValidateStruct(context.Background(), obj)

		if !result.HasErrors() {
			t.Error("expected validation error for gt")
		}

		found := false
		for _, err := range result.Errors {
			if err.Field == "count" && strings.Contains(err.Message, "greater than") {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("expected gt error, got %v", result.Errors)
		}
	})

	t.Run("lt validation", func(t *testing.T) {
		type LtTest struct {
			Count int `validate:"lt=10"`
		}

		obj := LtTest{Count: 15}
		result := ValidateStruct(context.Background(), obj)

		if !result.HasErrors() {
			t.Error("expected validation error for lt")
		}

		found := false
		for _, err := range result.Errors {
			if err.Field == "count" && strings.Contains(err.Message, "less than") {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("expected lt error, got %v", result.Errors)
		}
	})

	t.Run("eq validation", func(t *testing.T) {
		type EqTest struct {
			Code string `validate:"eq=ACTIVE"`
		}

		obj := EqTest{Code: "INACTIVE"}
		result := ValidateStruct(context.Background(), obj)

		if !result.HasErrors() {
			t.Error("expected validation error for eq")
		}

		found := false
		for _, err := range result.Errors {
			if err.Field == "code" && strings.Contains(err.Message, "equal to") {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("expected eq error, got %v", result.Errors)
		}
	})

	t.Run("ne validation", func(t *testing.T) {
		type NeTest struct {
			Status string `validate:"ne=DELETED"`
		}

		obj := NeTest{Status: "DELETED"}
		result := ValidateStruct(context.Background(), obj)

		if !result.HasErrors() {
			t.Error("expected validation error for ne")
		}

		found := false
		for _, err := range result.Errors {
			if err.Field == "status" && strings.Contains(err.Message, "not be equal") {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("expected ne error, got %v", result.Errors)
		}
	})

	t.Run("oneof validation", func(t *testing.T) {
		type OneOfTest struct {
			Status string `validate:"oneof=active inactive pending"`
		}

		obj := OneOfTest{Status: "deleted"}
		result := ValidateStruct(context.Background(), obj)

		if !result.HasErrors() {
			t.Error("expected validation error for oneof")
		}

		found := false
		for _, err := range result.Errors {
			if err.Field == "status" && strings.Contains(err.Message, "one of") {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("expected oneof error, got %v", result.Errors)
		}
	})

	t.Run("multiple validation errors", func(t *testing.T) {
		type MultipleErrors struct {
			Name  string `validate:"required"`
			Email string `validate:"required,email"`
			Age   int    `validate:"gte=0"`
		}

		obj := MultipleErrors{
			Name:  "",
			Email: "invalid",
			Age:   -5,
		}
		result := ValidateStruct(context.Background(), obj)

		if len(result.Errors) < 2 {
			t.Errorf("expected multiple errors, got %d: %v", len(result.Errors), result.Errors)
		}
	})

	t.Run("field names are lowercase", func(t *testing.T) {
		type CaseTest struct {
			UserName string `validate:"required"`
		}

		obj := CaseTest{UserName: ""}
		result := ValidateStruct(context.Background(), obj)

		if len(result.Errors) == 0 {
			t.Fatal("expected validation errors")
		}

		// Field name should be lowercase
		if result.Errors[0].Field != "username" {
			t.Errorf("expected lowercase field name 'username', got '%s'", result.Errors[0].Field)
		}
	})

	t.Run("nil struct", func(t *testing.T) {
		// This should not panic
		result := ValidateStruct(context.Background(), nil)

		// For nil input, the behavior depends on the validator implementation
		// The important thing is that it doesn't panic
		_ = result
	})
}

func TestGetErrorMessage(t *testing.T) {
	// Note: We can't directly test getErrorMessage since it's not exported
	// But we can test it indirectly through ValidateStruct

	t.Run("unknown tag generates default message", func(t *testing.T) {
		// Create a struct with a validator tag that doesn't have a custom message
		// This will test the default case in getErrorMessage
		type UnknownTag struct {
			Field string `validate:"startswith=hello"`
		}

		obj := UnknownTag{Field: "world"}
		result := ValidateStruct(context.Background(), obj)

		if !result.HasErrors() {
			t.Fatal("expected validation error")
		}

		// Should have a generic message
		found := false
		for _, err := range result.Errors {
			if strings.Contains(err.Message, "failed validation") {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("expected generic validation message for unknown tag, got %v", result.Errors)
		}
	})
}
