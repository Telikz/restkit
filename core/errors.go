package core

import (
	err "github.com/reststore/restkit/internal/errors"
	vd "github.com/reststore/restkit/internal/validation"
)

// ValidationError is an alias for internal/errors.ValidationError. See restkit.ValidationError for details.
type ValidationError = err.ValidationError

// ValidationResult is an alias for internal/errors.ValidationResult. See restkit.ValidationResult for details.
type ValidationResult = err.ValidationResult

// APIError is an alias for internal/errors.APIError. See restkit.APIError for details.
type APIError = err.APIError

// NewAPIError creates a standardized API error response.
var NewAPIError = err.NewAPIError

// NewValidation creates an empty validation result to populate with errors.
var NewValidation = err.NewValidation

// ValidationFailed creates a failed validation result with a single error.
var ValidationFailed = err.ValidationFailed

// ValidationFailedMulti creates a failed validation result with multiple errors.
var ValidationFailedMulti = err.ValidationFailedMulti

// Validate is the validation function used by endpoints.
var Validate = vd.Validate

const (
	ErrCodeInternal      = err.ErrCodeInternal
	ErrCodeConfiguration = err.ErrCodeConfiguration
	ErrCodeValidation    = err.ErrCodeValidation
	ErrCodeBind          = err.ErrCodeBind
	ErrCodeNotFound      = err.ErrCodeNotFound
	ErrCodeUnauthorized  = err.ErrCodeUnauthorized
	ErrCodeForbidden     = err.ErrCodeForbidden
	ErrCodeBadRequest    = err.ErrCodeBadRequest
	ErrCodeMissingParam  = err.ErrCodeMissingParam
)
