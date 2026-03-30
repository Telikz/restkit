package restkit

import (
	"context"
	"net/http"

	api "github.com/telikz/restkit/internal"
	rc "github.com/telikz/restkit/internal/context"
	ep "github.com/telikz/restkit/internal/endpoints"
	err "github.com/telikz/restkit/internal/errors"
	mw "github.com/telikz/restkit/internal/middleware"
	sc "github.com/telikz/restkit/internal/schema"
	vd "github.com/telikz/restkit/internal/validation"
)

// Api is the main struct for defining your API,
// including metadata, endpoints, and middleware
type Api = api.Api

// Group represents a collection of endpoints that share a common URL prefix and metadata
type Group = ep.Group

// Endpoint represents an API endpoint with both request and response bodies
type Endpoint[Req any, Res any] = ep.EndpointReqRes[Req, Res]

// EndpointRes represents an API endpoint that only returns a response body without accepting a request body
type EndpointRes[Res any] = ep.EndpointRes[Res]

// EndpointReq represents an API endpoint that only accepts a request body without returning a response body
type EndpointReq[Req any] = ep.EndpointReq[Req]

// MountedRoute represents a route from an external router mounted into RestKit
type MountedRoute = sc.MountedRoute

// ParamInfo represents a path parameter definition for OpenAPI documentation
type ParamInfo = sc.ParamInfo

// ValidationError represents a single validation error with field and message
type ValidationError = err.ValidationError

// ValidationResult is returned by validation functions with code, message, and list of errors
type ValidationResult = err.ValidationResult

// NewValidation creates an empty validation result to populate with errors
var NewValidation = err.NewValidation

// ValidationFailed creates a failed validation result with a single error
var ValidationFailed = err.ValidationFailed

// ValidationFailedMulti creates a failed validation result with multiple errors
var ValidationFailedMulti = err.ValidationFailedMulti

// ValidateStruct validates a struct using go-playground/validator tags
var ValidateStruct = vd.ValidateStruct

// RouteContext contains information about the current route and request
type RouteContext = rc.RouteContext

// NewApi creates a new Api instance
func NewApi() *Api {
	return api.New()
}

// NewGroup creates a new group of endpoints with a common URL prefix
func NewGroup(prefix string) *Group {
	return ep.NewGroup(prefix)
}

// NewEndpoint creates a new endpoint with both request and response bodies
// Sets up sensible defaults: POST method, JSON bind/write, auto-generated schemas
func NewEndpoint[Req any, Res any]() *Endpoint[Req, Res] {
	return ep.NewEndpoint[Req, Res]()
}

// NewEndpointRes creates an endpoint that only returns a response body without accepting a request body
// Sets up sensible defaults: GET method, JSON write, auto-generated response schema
func NewEndpointRes[Res any]() *EndpointRes[Res] {
	return ep.NewEndpointRes[Res]()
}

// NewEndpointReq creates an endpoint that only accepts a request body without returning a response body
// Sets up sensible defaults: DELETE method, JSON bind, auto-generated request schema
func NewEndpointReq[Req any]() *EndpointReq[Req] {
	return ep.NewEndpointReq[Req]()
}

// URLParam retrieves a URL parameter from the request context
func URLParam(ctx context.Context, key string) string {
	return rc.URLParam(ctx, key)
}

// RouteCtxFromContext extracts the route context from a request context
func RouteCtxFromContext(ctx context.Context) *RouteContext {
	return rc.RouteCtxFromContext(ctx)
}

// ExtractPathParams extracts parameters from a URL path using a pattern
// Pattern should be like "/users/{id}/posts/{postId}"
func ExtractPathParams(pattern, path string) map[string]string {
	return rc.ExtractPathParams(pattern, path)
}

// JSONBinder creates a bind function for JSON request bodies
func JSONBinder[Req any]() func(r *http.Request) (Req, error) {
	return mw.JSONBinder[Req]()
}

// SchemaFrom generates a JSON Schema from a Go type using reflection
// Useful for manually setting or overriding endpoint schemas
func SchemaFrom[T any]() map[string]any {
	return sc.SchemaFrom[T]()
}

// PathParamBinder creates a bind function that extracts the last path segment
// and converts it to the specified type
func PathParamBinder[T any](convert func(string) (T, error)) func(r *http.Request) (T, error) {
	return mw.PathParamBinder[T](convert)
}

// JSONWriter creates a write function for JSON responses
func JSONWriter[Res any]() func(w http.ResponseWriter, res Res) error {
	return mw.JSONWriter[Res]()
}

// JSONErrorWriter writes error responses as JSON
func JSONErrorWriter(w http.ResponseWriter, r *http.Request, err error) {
	mw.JSONErrorWriter(w, r, err)
}

// LoggingMiddleware logs incoming requests with timing
func LoggingMiddleware() func(next http.Handler) http.Handler {
	return mw.LoggingMiddleware()
}

// CORSMiddleware adds CORS headers to responses with sensible defaults
// Deprecated: Use NewCORS with options instead for more flexibility
func CORSMiddleware(allowedOrigins ...string) func(next http.Handler) http.Handler {
	return mw.CORSMiddleware(allowedOrigins...)
}

// CORSOption configures CORS middleware behavior
type CORSOption = mw.CORSOption

// NewCORS creates a CORS middleware with sensible defaults and optional overrides
func NewCORS(opts ...CORSOption) func(next http.Handler) http.Handler {
	return mw.NewCORS(opts...)
}

// WithOrigins sets the allowed origins for CORS
func WithOrigins(origins ...string) CORSOption {
	return mw.WithOrigins(origins...)
}

// WithMethods sets the allowed HTTP methods for CORS
func WithMethods(methods ...string) CORSOption {
	return mw.WithMethods(methods...)
}

// WithHeaders sets the allowed headers for CORS
func WithHeaders(headers ...string) CORSOption {
	return mw.WithHeaders(headers...)
}

// WithCredentials enables credentials support for CORS
func WithCredentials() CORSOption {
	return mw.WithCredentials()
}

// WithMaxAge sets the max age for preflight cache (in seconds)
func WithMaxAge(seconds int) CORSOption {
	return mw.WithMaxAge(seconds)
}

// RecoveryMiddleware recovers from panics and returns 500 error
func RecoveryMiddleware() func(next http.Handler) http.Handler {
	return mw.RecoveryMiddleware()
}

// StringToInt converts a string to int
func StringToInt(s string) (int, error) {
	return mw.StringToInt(s)
}

// StringToString is a no-op converter for string path params
func StringToString(s string) (string, error) {
	return mw.StringToString(s)
}

// Error codes returned by the API for programmatic error handling
const (
	// ErrCodeInternal indicates an internal server error
	ErrCodeInternal = err.ErrCodeInternal

	// ErrCodeConfiguration indicates the endpoint is not properly configured
	ErrCodeConfiguration = err.ErrCodeConfiguration

	// ErrCodeValidation indicates validation failed
	ErrCodeValidation = err.ErrCodeValidation

	// ErrCodeBind indicates a request binding/parsing error
	ErrCodeBind = err.ErrCodeBind

	// ErrCodeNotFound indicates a resource was not found
	ErrCodeNotFound = err.ErrCodeNotFound

	// ErrCodeUnauthorized indicates authentication is required
	ErrCodeUnauthorized = err.ErrCodeUnauthorized

	// ErrCodeForbidden indicates access is denied
	ErrCodeForbidden = err.ErrCodeForbidden

	// ErrCodeBadRequest indicates a malformed request
	ErrCodeBadRequest = err.ErrCodeBadRequest

	// ErrCodeMissingParam indicates a missing path parameter
	ErrCodeMissingParam = err.ErrCodeMissingParam
)
