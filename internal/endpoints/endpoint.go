package endpoints

import (
	"net/http"

	"github.com/telikz/restkit/internal/errors"
)

// ValidationError is an alias for errors.ValidationError
type ValidationError = errors.ValidationError

// ValidationResult is an alias for errors.ValidationResult
type ValidationResult = errors.ValidationResult

// APIError is an alias for errors.APIError
type APIError = errors.APIError

type Endpoint interface {
	Pattern() string
	HTTPHandler() http.Handler
	GetMethod() string
	GetPath() string
	GetTitle() string
	GetDescription() string
	GetMiddleware() []func(http.Handler) http.Handler
	GetRequestSchema() map[string]any
	GetResponseSchema() map[string]any
}
