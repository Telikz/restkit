package endpoints

import (
	"net/http"

	"github.com/telikz/restkit/internal/errors"
)

// APIError is an alias for errors.APIError
type APIError = errors.APIError

// ValidationError is an alias for errors.ValidationError
type ValidationError = errors.ValidationError

// ValidationResult is an alias for errors.ValidationResult
type ValidationResult = errors.ValidationResult

// Endpoint defines the interface for all endpoint types.
// This interface is router-agnostic and is used internally by RestKit to manage endpoints.
type Endpoint interface {
	GetMethod() string
	GetPath() string
	GetTitle() string
	GetDescription() string
	GetMiddleware() []func(http.Handler) http.Handler
	GetRequestSchema() map[string]any
	GetResponseSchema() map[string]any
	GetHandler() http.Handler
}
