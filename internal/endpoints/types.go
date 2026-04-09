package endpoints

import (
	"context"
	"net/http"
	"regexp"

	"github.com/reststore/restkit/internal/errors"
)

// ParamLocation defines where a parameter is located
type ParamLocation string

const (
	ParamLocationPath  ParamLocation = "path"
	ParamLocationQuery ParamLocation = "query"
)

// Parameter defines an OpenAPI parameter (path or query)
type Parameter struct {
	Name        string
	Type        string
	Description string
	Required    bool
	Location    ParamLocation
}

// Route defines the interface for all endpoint types.
type Route interface {
	GetMethod() string
	GetPath() string
	GetTitle() string
	GetSummary() string
	GetDescription() string
	GetMiddleware() []func(http.Handler) http.Handler
	GetRequestSchema() map[string]any
	GetResponseSchema() map[string]any
	GetHandler() http.Handler
	GetParameters() []Parameter
	setPath(path string)
	addMiddleware(mw []func(http.Handler) http.Handler)
}

// NoRequest is a sentinel type indicating an endpoint has no request body.
type NoRequest struct{}

// NoResponse is a sentinel type indicating an endpoint has no response body.
type NoResponse struct{}

// ValidatableRequest is an interface for request types that can validate themselves.
type ValidatableRequest interface {
	Validate(ctx context.Context) ValidationResult
}

// APIError is an alias for errors.APIError
type APIError = errors.APIError

// ValidationError is an alias for errors.ValidationError
type ValidationError = errors.ValidationError

// ValidationResult is an alias for errors.ValidationResult
type ValidationResult = errors.ValidationResult

var pathParamRegex = regexp.MustCompile(`\{([^}]+)}`)

func extractPathParamNames(pattern string) []string {
	var names []string
	matches := pathParamRegex.FindAllStringSubmatch(pattern, -1)
	for _, match := range matches {
		names = append(names, match[1])
	}
	return names
}

func isNoRequest[T any]() bool {
	_, ok := any(*new(T)).(NoRequest)
	return ok
}

func isNoResponse[T any]() bool {
	_, ok := any(*new(T)).(NoResponse)
	return ok
}
