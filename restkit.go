package restkit

import (
	"context"
	"net/http"

	"github.com/reststore/restkit/internal/api"
	rc "github.com/reststore/restkit/internal/context"
	"github.com/reststore/restkit/internal/docs"
	ep "github.com/reststore/restkit/internal/endpoints"
	err "github.com/reststore/restkit/internal/errors"
	mw "github.com/reststore/restkit/internal/middleware"
	sc "github.com/reststore/restkit/internal/schema"
	vd "github.com/reststore/restkit/internal/validation"
)

// Api is the main struct for defining your API, including metadata, endpoints, and middleware
type Api = api.Api

// Group represents a collection of endpoints that share a common URL prefix and metadata
type Group = ep.Group

// Endpoint represents a unified API endpoint with configurable request and response bodies.
// Use NoRequest or NoResponse as type parameters when you don't need a request body or response body.
type Endpoint[Req any, Res any] = ep.Endpoint[Req, Res]

// NoRequest is a sentinel type for endpoints without a request body.
type NoRequest = ep.NoRequest

// NoResponse is a sentinel type for endpoints without a response body.
type NoResponse = ep.NoResponse

// MountedRoute represents a route from an external router mounted into RestKit
type MountedRoute = sc.MountedRoute

// ParamInfo represents a path parameter definition for OpenAPI documentation
type ParamInfo = sc.ParamInfo

// Parameter represents an OpenAPI parameter (path or query)
type Parameter = ep.Parameter

// ParamLocation defines where a parameter is located
type ParamLocation = ep.ParamLocation

// ParamLocationPath indicates a path parameter
const ParamLocationPath = ep.ParamLocationPath

// ParamLocationQuery indicates a query parameter
const ParamLocationQuery = ep.ParamLocationQuery

// RouteInfo contains metadata for a route used by adapters
type RouteInfo = sc.RouteInfo

// RouteMeta represents a route with its metadata for extraction
type RouteMeta = sc.RouteMeta

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

// ValidateStruct is a no-op by default. Import playground to enable:
//
//	import _ "github.com/reststore/restkit/validation/playground"
var ValidateStruct = vd.ValidateStruct

// GenerateOpenAPIFile generates openApi spec file at specified location
var GenerateOpenAPIFile = docs.CreateOpenAPIFile

// RouteContext contains information about the current route and request
type RouteContext = rc.RouteContext

// WithQueries injects database queries into the context.
var WithQueries = rc.WithQueries

// QueriesFromContext retrieves database queries from the context.
var QueriesFromContext = rc.QueriesFromContext

// MustQueriesFromContext retrieves database queries from the context.
// Panics if not found.
var MustQueriesFromContext = rc.MustQueriesFromContext

// NewApi creates a new Api instance
func NewApi() *Api {
	return api.New()
}

// NewGroup creates a new group of endpoints with a common URL prefix
func NewGroup(prefix string) *Group {
	return ep.NewGroup(prefix)
}

// NewEndpoint creates a new endpoint with both request and response bodies.
// Sets up sensible defaults: POST method (or GET for NoRequest, DELETE for NoResponse),
// JSON bind/write (when applicable), auto-generated schemas.
func NewEndpoint[Req any, Res any]() *Endpoint[Req, Res] {
	return ep.NewEndpoint[Req, Res]()
}

// NewEndpointRes creates an endpoint that only returns a response body without accepting a request body.
// Sets up sensible defaults: GET method, JSON write, auto-generated response schema.
// This is equivalent to NewEndpoint[NoRequest, Res]().
func NewEndpointRes[Res any]() *Endpoint[NoRequest, Res] {
	return ep.NewEndpointRes[Res]()
}

// NewEndpointReq creates an endpoint that only accepts a request body without returning a response body.
// Sets up sensible defaults: DELETE method, JSON bind, auto-generated request schema.
// This is equivalent to NewEndpoint[Req, NoResponse]().
func NewEndpointReq[Req any]() *Endpoint[Req, NoResponse] {
	return ep.NewEndpointReq[Req]()
}

// URLParam retrieves a URL parameter from the request context
func URLParam(ctx context.Context, key string) string {
	return rc.URLParam(ctx, key)
}

// URLQueryParam retrieves a URL query parameter from the request context
func URLQueryParam(ctx context.Context, key string) string {
	return rc.URLQueryParam(ctx, key)
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
func PathParamBinder[T any](
	convert func(string) (T, error),
) func(r *http.Request) (T, error) {
	return mw.PathParamBinder(convert)
}

// JSONWriter creates a write function for JSON responses
func JSONWriter[Res any]() func(w http.ResponseWriter, res Res) error {
	return mw.JSONWriter[Res]()
}

// JSONErrorWriter writes error responses as JSON
func JSONErrorWriter(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	mw.JSONErrorWriter(w, r, err)
}

// LoggingMiddleware logs incoming requests with timing
func LoggingMiddleware() func(next http.Handler) http.Handler {
	return mw.LoggingMiddleware()
}

// CORSOption configures CORS middleware behavior
type CORSOption = mw.CORSOption

// NewCORS creates a CORS middleware with sensible defaults and optional overrides
func NewCORS(
	opts ...CORSOption,
) func(next http.Handler) http.Handler {
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
var RecoveryMiddleware = mw.RecoveryMiddleware

// DBMiddleware injects database queries into every request context.
var DBMiddleware = mw.DBMiddleware

// TransactionMiddleware wraps requests in a database transaction.
var TransactionMiddleware = mw.TransactionMiddleware

// StringToInt converts a string to int
var StringToInt = mw.StringToInt

// StringToString is a no-op converter for string path params
var StringToString = mw.StringToString

// ParseID converts a string ID to int64.
var ParseID = ep.ParseID

// ParseIntID converts a string ID to int.
var ParseIntID = ep.ParseIntID

// MessageResponse is a simple response with a message.
type MessageResponse = ep.MessageResponse

// PaginationParams contains common pagination parameters.
type PaginationParams = ep.PaginationParams

// ListParams contains pagination and sorting parameters for list endpoints.
type ListParams = ep.ListParams

// SearchParams contains pagination and search query parameters.
type SearchParams = ep.SearchParams

func ListEndpoint[Q any, T any](
	path string,
	listFn func(ctx context.Context, queries Q, limit, offset int32) ([]T, error),
) *Endpoint[NoRequest, []T] {
	return ep.ListEndpoint(path, listFn)
}

func ListPaginatedEndpoint[Q any, T any](
	path string,
	listFn func(ctx context.Context, queries Q, params ListParams) ([]T, error),
) *Endpoint[NoRequest, []T] {
	return ep.ListPaginatedEndpoint(path, listFn)
}

func GetEndpoint[Q any, T any](
	path string,
	getFn func(ctx context.Context, queries Q, id int64) (T, error),
) *Endpoint[NoRequest, T] {
	return ep.GetEndpoint(path, getFn)
}

func CreateEndpoint[Q any, Req any, Res any](
	path string,
	createFn func(ctx context.Context, queries Q, req Req) (Res, error),
) *Endpoint[Req, Res] {
	return ep.CreateEndpoint(path, createFn)
}

func UpdateEndpoint[Q any, Req any](
	path string,
	updateFn func(ctx context.Context, queries Q, id int64, req Req) error,
) *Endpoint[Req, NoResponse] {
	return ep.UpdateEndpoint(path, updateFn)
}

func DeleteEndpoint[Q any](
	path string,
	deleteFn func(ctx context.Context, queries Q, id int64) error,
) *Endpoint[NoRequest, MessageResponse] {
	return ep.DeleteEndpoint(path, deleteFn)
}

func SearchEndpoint[Q any, T any](
	path string,
	searchFn func(ctx context.Context, queries Q) ([]T, error),
) *Endpoint[NoRequest, []T] {
	return ep.SearchEndpoint(path, searchFn)
}

func SearchPaginatedEndpoint[Q any, T any](
	path string,
	searchFn func(ctx context.Context, queries Q, params SearchParams) ([]T, error),
) *Endpoint[NoRequest, []T] {
	return ep.SearchPaginatedEndpoint(path, searchFn)
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
