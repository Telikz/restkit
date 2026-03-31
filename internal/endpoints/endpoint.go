package endpoints

import (
	"encoding/json"
	"net/http"

	"github.com/reststore/restkit/internal/errors"
)

// APIError is an alias for errors.APIError
type APIError = errors.APIError

// ValidationError is an alias for errors.ValidationError
type ValidationError = errors.ValidationError

// ValidationResult is an alias for errors.ValidationResult
type ValidationResult = errors.ValidationResult

// errorHandler creates a handler that always returns a specific error
func errorHandler(apiErr errors.APIError) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(apiErr.Status)
		json.NewEncoder(w).Encode(apiErr)
	})
}

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
