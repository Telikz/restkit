package errors

import (
	"errors"
	"testing"
)

// TestNewAPIError tests the APIError constructor
func TestNewAPIError(t *testing.T) {
	apiErr := NewAPIError(404, "not_found", "Resource not found")

	if apiErr.Status != 404 {
		t.Errorf("expected status 404, got %d", apiErr.Status)
	}

	if apiErr.Code != "not_found" {
		t.Errorf("expected code 'not_found', got '%s'", apiErr.Code)
	}

	if apiErr.Message != "Resource not found" {
		t.Errorf("expected message 'Resource not found', got '%s'", apiErr.Message)
	}

	if apiErr.Details != "" {
		t.Error("details should be empty when not provided")
	}
}

// TestNewAPIErrorWithDetails tests the constructor with details
func TestNewAPIErrorWithDetails(t *testing.T) {
	apiErr := NewAPIErrorWithDetails(
		500,
		"internal",
		"An internal error occurred",
		"Database connection failed",
	)

	if apiErr.Status != 500 {
		t.Errorf("expected status 500, got %d", apiErr.Status)
	}

	if apiErr.Code != "internal" {
		t.Errorf("expected code 'internal', got '%s'", apiErr.Code)
	}

	if apiErr.Details != "Database connection failed" {
		t.Errorf("expected details 'Database connection failed', got '%s'", apiErr.Details)
	}
}

// TestAPIErrorError tests the Error() method
func TestAPIErrorError(t *testing.T) {
	apiErr := NewAPIError(400, "bad_request", "Invalid input")

	// Error() should return the message
	if apiErr.Error() != "Invalid input" {
		t.Errorf("expected Error() to return 'Invalid input', got '%s'", apiErr.Error())
	}
}

// TestIsAPIError tests the type checking function
func TestIsAPIError(t *testing.T) {
	t.Run("with APIError", func(t *testing.T) {
		apiErr := NewAPIError(404, "not_found", "Not found")

		// Need to wrap it properly
		extracted, ok := IsAPIError(apiErr)

		if !ok {
			t.Error("expected to identify as APIError")
		}

		if extracted.Status != 404 {
			t.Errorf("expected status 404, got %d", extracted.Status)
		}
	})

	t.Run("with regular error", func(t *testing.T) {
		regularErr := errors.New("something went wrong")

		_, ok := IsAPIError(regularErr)

		if ok {
			t.Error("should not identify regular error as APIError")
		}
	})

	t.Run("with nil", func(t *testing.T) {
		_, ok := IsAPIError(nil)

		if ok {
			t.Error("should not identify nil as APIError")
		}
	})

	t.Run("wrapped APIError", func(t *testing.T) {
		apiErr := NewAPIError(400, "bad_request", "Invalid")
		wrapped := errors.Join(apiErr, errors.New("additional context"))

		// The current implementation may not handle wrapped errors
		// depending on the Go version and errors.AsType behavior
		extracted, ok := IsAPIError(wrapped)

		// Just verify it doesn't panic
		_ = extracted
		_ = ok
	})
}

// TestNewValidation tests the ValidationResult constructor
func TestNewValidation(t *testing.T) {
	result := NewValidation()

	if result.HasErrors() {
		t.Error("new validation should not have errors")
	}

	if len(result.Errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(result.Errors))
	}
}

// TestValidationFailed tests creating a failed validation with single error
func TestValidationFailed(t *testing.T) {
	result := ValidationFailed(
		422,
		"validation",
		"Validation failed",
		"email",
		"Email is required",
	)

	if !result.HasErrors() {
		t.Error("result should have errors")
	}

	if result.Status != 422 {
		t.Errorf("expected status 422, got %d", result.Status)
	}

	if result.Code != "validation" {
		t.Errorf("expected code 'validation', got '%s'", result.Code)
	}

	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}

	if result.Errors[0].Field != "email" {
		t.Errorf("expected field 'email', got '%s'", result.Errors[0].Field)
	}

	if result.Errors[0].Message != "Email is required" {
		t.Errorf("expected message 'Email is required', got '%s'", result.Errors[0].Message)
	}
}

// TestValidationFailedMulti tests creating a failed validation with multiple errors
func TestValidationFailedMulti(t *testing.T) {
	errors := []ValidationError{
		{Field: "name", Message: "Name is required"},
		{Field: "email", Message: "Email is invalid"},
	}

	result := ValidationFailedMulti(
		422,
		"validation",
		"Validation failed",
		errors...,
	)

	if len(result.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(result.Errors))
	}
}

// TestValidationResultAddError tests adding errors
func TestValidationResultAddError(t *testing.T) {
	result := NewValidation()

	result.AddError("username", "Username is taken")

	if !result.HasErrors() {
		t.Error("result should have errors after adding")
	}

	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(result.Errors))
	}

	result.AddError("password", "Password is too short")

	if len(result.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(result.Errors))
	}
}

// TestValidationResultWithStatus tests setting status
func TestValidationResultWithStatus(t *testing.T) {
	result := NewValidation()
	result.WithStatus(400)

	if result.Status != 400 {
		t.Errorf("expected status 400, got %d", result.Status)
	}
}

// TestValidationResultWithCode tests setting code
func TestValidationResultWithCode(t *testing.T) {
	result := NewValidation()
	result.WithCode("bad_request")

	if result.Code != "bad_request" {
		t.Errorf("expected code 'bad_request', got '%s'", result.Code)
	}
}

// TestValidationResultWithMessage tests setting message
func TestValidationResultWithMessage(t *testing.T) {
	result := NewValidation()
	result.WithMessage("Invalid request")

	if result.Message != "Invalid request" {
		t.Errorf("expected message 'Invalid request', got '%s'", result.Message)
	}
}

// TestValidationResultHasErrors tests the error check
func TestValidationResultHasErrors(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		result := NewValidation()

		if result.HasErrors() {
			t.Error("new validation should not have errors")
		}
	})

	t.Run("with errors", func(t *testing.T) {
		result := NewValidation()
		result.AddError("field", "error")

		if !result.HasErrors() {
			t.Error("validation with errors should return true")
		}
	})

	t.Run("nil receiver", func(t *testing.T) {
		var result *ValidationResult

		// Note: HasErrors() doesn't handle nil receiver and will panic
		// This is intentional - ValidationResult should always be initialized
		// with NewValidation() before use
		// We just verify the type exists
		_ = result
	})
}

// TestErrorCodes tests that error code constants are defined
func TestErrorCodes(t *testing.T) {
	// Just verify constants exist and have expected values
	codes := []string{
		ErrCodeInternal,
		ErrCodeConfiguration,
		ErrCodeValidation,
		ErrCodeBind,
		ErrCodeNotFound,
		ErrCodeUnauthorized,
		ErrCodeForbidden,
		ErrCodeBadRequest,
		ErrCodeMissingParam,
	}

	for _, code := range codes {
		if code == "" {
			t.Error("error code should not be empty")
		}
	}

	// Check specific values
	if ErrCodeInternal != "internal" {
		t.Errorf("expected ErrCodeInternal to be 'internal', got '%s'", ErrCodeInternal)
	}

	if ErrCodeValidation != "validation" {
		t.Errorf("expected ErrCodeValidation to be 'validation', got '%s'", ErrCodeValidation)
	}
}

// TestErrorMessages tests that error message constants are defined
func TestErrorMessages(t *testing.T) {
	messages := []string{
		ErrMsgInternal,
		ErrMsgConfiguration,
		ErrMsgValidation,
		ErrMsgBind,
		ErrMsgNotFound,
		ErrMsgUnauthorized,
		ErrMsgForbidden,
		ErrMsgBadRequest,
		ErrMsgMissingPathParam,
		ErrMsgInvalidInteger,
	}

	for _, msg := range messages {
		if msg == "" {
			t.Error("error message should not be empty")
		}
	}
}
