package restkit

import (
	"context"
	"net/http"

	"github.com/reststore/restkit/core"
)

// API

// Api is the main struct for defining your API, including metadata, endpoints, and middleware.
type Api = core.Api

// NewApi creates a new Api instance.
var NewApi = core.NewApi

// Group represents a collection of endpoints that share a common URL prefix and metadata.
type Group = core.Group

// NewGroup creates a new group of endpoints with a common URL prefix.
var NewGroup = core.NewGroup

// Endpoints

type (
	// Endpoint represents a unified API endpoint with configurable request and response bodies.
	Endpoint[Req any, Res any] = core.Endpoint[Req, Res]

	// NoRequest is a sentinel type for endpoints without a request body.
	NoRequest = core.NoRequest

	// GetRequest is the standard request type for Get endpoints with ID path param.
	GetRequest = core.GetRequest

	// DeleteRequest is the standard request type for Delete endpoints with ID path param.
	DeleteRequest = core.DeleteRequest

	// PaginationRequest provides standard pagination parameters (page, limit).
	PaginationRequest = core.PaginationRequest

	// SearchRequest provides standard search parameters (query, filters).
	SearchRequest = core.SearchRequest

	// ListRequest combines pagination and search parameters for list operations.
	ListRequest = core.ListRequest

	// NoResponse is a sentinel type for endpoints without a response body.
	NoResponse = core.NoResponse

	// MessageResponse is a simple JSON response with a message field.
	MessageResponse = core.MessageResponse

	// Event is an alias for internal/endpoints.Event. See restkit.Event for details.
	Event[T any] = core.Event[T]
)

// Validate is the validation function used by endpoints.
var Validate = core.Validate

// NewEndpoint creates a new endpoint with both request and response bodies.
func NewEndpoint[Req any, Res any]() *Endpoint[Req, Res] {
	return core.NewEndpoint[Req, Res]()
}

// NewEndpointRes creates an endpoint that only returns a response body.
func NewEndpointRes[Res any]() *Endpoint[NoRequest, Res] {
	return core.NewEndpointRes[Res]()
}

// NewEndpointReq creates an endpoint that only accepts a request body.
func NewEndpointReq[Req any]() *Endpoint[Req, NoResponse] {
	return core.NewEndpointReq[Req]()
}

func List[Req any, Res any](
	path string,
	listFn func(ctx context.Context, req Req) ([]Res, error),
) *Endpoint[Req, []Res] {
	return core.List(path, listFn)
}

// ListEndpoint creates an endpoint for listing resources.
func ListEndpoint[Q any, Req any, Res any](
	path string,
	listFn func(ctx context.Context, queries Q, req Req) ([]Res, error),
) *Endpoint[Req, []Res] {
	return core.ListEndpoint(path, listFn)
}

func Get[Req any, Res any](
	path string,
	getFn func(ctx context.Context, req Req) (Res, error),
) *Endpoint[Req, Res] {
	return core.Get(path, getFn)
}

// GetEndpoint creates an endpoint for getting a single resource.
func GetEndpoint[Q any, Req any, Res any](
	path string,
	getFn func(ctx context.Context, queries Q, req Req) (Res, error),
) *Endpoint[Req, Res] {
	return core.GetEndpoint(path, getFn)
}

func Create[Req any, Res any](
	path string,
	createFn func(ctx context.Context, req Req) (Res, error),
) *Endpoint[Req, Res] {
	return core.Create(path, createFn)
}

// CreateEndpoint creates an endpoint for creating resources.
func CreateEndpoint[Q any, Req any, Res any](
	path string,
	createFn func(ctx context.Context, queries Q, req Req) (Res, error),
) *Endpoint[Req, Res] {
	return core.CreateEndpoint(path, createFn)
}

func Update[Req any, Res any](
	path string,
	updateFn func(ctx context.Context, req Req) (Res, error),
) *Endpoint[Req, Res] {
	return core.Update(path, updateFn)
}

// UpdateEndpoint creates an endpoint for updating resources.
func UpdateEndpoint[Q any, Req any, Res any](
	path string,
	updateFn func(ctx context.Context, queries Q, req Req) (Res, error),
) *Endpoint[Req, Res] {
	return core.UpdateEndpoint(path, updateFn)
}

func Delete[Req any](
	path string,
	deleteFn func(ctx context.Context, req Req) error,
) *Endpoint[Req, NoResponse] {
	return core.Delete(path, deleteFn)
}

// DeleteEndpoint creates an endpoint for deleting resources.
func DeleteEndpoint[Q any, Req any](
	path string,
	deleteFn func(ctx context.Context, queries Q, req Req) error,
) *Endpoint[Req, NoResponse] {
	return core.DeleteEndpoint(path, deleteFn)
}

func Search[Req any, Res any](
	path string,
	searchFn func(ctx context.Context, req Req) ([]Res, error),
) *Endpoint[Req, []Res] {
	return core.Search(path, searchFn)
}

// SearchEndpoint creates an endpoint for searching resources.
func SearchEndpoint[Q any, Req any, Res any](
	path string,
	searchFn func(ctx context.Context, queries Q, req Req) ([]Res, error),
) *Endpoint[Req, []Res] {
	return core.SearchEndpoint(path, searchFn)
}

func Stream[Req any, Res any](
	path string,
	streamFn func(ctx context.Context, req Req) (<-chan Event[Res], error),
) *Endpoint[Req, <-chan Event[Res]] {
	return core.Stream(path, streamFn)
}

// Parameters

// Parameter represents an OpenAPI parameter (path or query).
type Parameter = core.Parameter

const (
	// ParamLocationPath indicates a path parameter (e.g., /users/{id}).
	ParamLocationPath = core.ParamLocationPath

	// ParamLocationQuery indicates a query parameter (e.g., ?page=1).
	ParamLocationQuery = core.ParamLocationQuery
)

// Errors

type (
	// APIError represents a standardized API error response with status, code, and message.
	APIError = core.APIError

	// ValidationResult is returned by validation functions with code, message, and list of errors.
	ValidationResult = core.ValidationResult

	// ValidationError represents a single validation error with field and message.
	ValidationError = core.ValidationError
)

var (
	// NewAPIError creates a standardized API error response with status code, error code, and message.
	NewAPIError = core.NewAPIError

	// NewValidation creates an empty validation result to populate with errors.
	NewValidation = core.NewValidation

	// ValidationFailed creates a failed validation result with a single error.
	ValidationFailed = core.ValidationFailed

	// ValidationFailedMulti creates a failed validation result with multiple errors.
	ValidationFailedMulti = core.ValidationFailedMulti
)

// Error codes returned by the API for programmatic error handling.
const (
	ErrCodeInternal      = core.ErrCodeInternal
	ErrCodeConfiguration = core.ErrCodeConfiguration
	ErrCodeValidation    = core.ErrCodeValidation
	ErrCodeBind          = core.ErrCodeBind
	ErrCodeNotFound      = core.ErrCodeNotFound
	ErrCodeUnauthorized  = core.ErrCodeUnauthorized
	ErrCodeForbidden     = core.ErrCodeForbidden
	ErrCodeBadRequest    = core.ErrCodeBadRequest
	ErrCodeMissingParam  = core.ErrCodeMissingParam
)

// Context

var (
	// URLParam retrieves a URL parameter from the request context.
	URLParam = core.URLParam

	// ExtractPathParams extracts parameters from a URL path using a pattern.
	ExtractPathParams = core.ExtractPathParams

	// RouteCtxFromContext extracts the route context from a request context.
	RouteCtxFromContext = core.RouteCtxFromContext

	// URLQueryParam retrieves a URL query parameter from the request context.
	URLQueryParam = core.URLQueryParam

	// WithQueries injects database queries into the context.
	WithQueries = core.WithQueries

	// QueriesFromContext retrieves database queries from the context.
	QueriesFromContext = core.QueriesFromContext

	// MustQueriesFromContext retrieves database queries from the context (panics if not found).
	MustQueriesFromContext = core.MustQueriesFromContext
)

// Serializers and Binders

// Serializers provides standard serialization functions for common formats (JSON, XML).
var Serializers = core.Serializers

// Binders and Writers

// JSONErrorWriter writes error responses as JSON.
var JSONErrorWriter = core.JSONErrorWriter

// JSONBinder creates a bind function for JSON request bodies.
func JSONBinder[Req any]() func(r *http.Request) (Req, error) {
	return core.JSONBinder[Req]()
}

// JSONWriter creates a write function for JSON responses.
func JSONWriter[Res any]() func(w http.ResponseWriter, res Res) error {
	return core.JSONWriter[Res]()
}

// QueryBinder creates a bind function that extracts query parameters.
func QueryBinder[Req any]() func(r *http.Request) (Req, error) {
	return core.QueryBinder[Req]()
}

// MixedBinder creates a bind function that combines path params and JSON body.
func MixedBinder[Req any]() func(r *http.Request) (Req, error) {
	return core.MixedBinder[Req]()
}

// PathParamBinder creates a bind function that extracts the last path segment.
func PathParamBinder[T any](convert func(string) (T, error)) func(r *http.Request) (T, error) {
	return core.PathParamBinder(convert)
}

// Middleware

var (
	// CORSMiddleware creates a CORS middleware with sensible defaults.
	CORSMiddleware = core.NewCORS

	// CORSOptions provides option functions for CORS middleware.
	CORSOptions = core.CORSOptions

	// SecurityHeaderMiddleware adds security headers to all responses.
	SecurityHeaderMiddleware = core.SecurityHeaders

	// SecurityHeadersOptions provides option functions for SecurityHeaders middleware.
	SecurityHeadersOptions = core.SecurityHeadersOptions

	// RequestIDMiddleware injects a unique request ID into each request.
	RequestIDMiddleware = core.RequestID

	// RequestIDOptions provides option functions for RequestID middleware.
	RequestIDOptions = core.RequestIDOptions

	// RequestIDFromContext retrieves the request ID from context.
	RequestIDFromContext = core.RequestIDFromContext

	// DBMiddleware injects database queries into every request context.
	DBMiddleware = core.DBMiddleware

	// TransactionMiddleware wraps requests in a database transaction.
	TransactionMiddleware = core.TransactionMiddleware

	// LoggingMiddleware logs incoming requests with timing.
	LoggingMiddleware = core.LoggingMiddleware

	// RecoveryMiddleware recovers from panics and returns 500 error.
	RecoveryMiddleware = core.RecoveryMiddleware
)

// OpenAPI

// GenerateOpenAPIFile generates OpenAPI spec file at specified location.
var GenerateOpenAPIFile = core.GenerateOpenAPIFile

type (
	// MountedRoute represents a route from an external router mounted into RestKit.
	MountedRoute = core.MountedRoute

	// ParamInfo represents a path parameter definition for OpenAPI documentation.
	ParamInfo = core.ParamInfo

	// RouteInfo contains metadata for a route used by adapters.
	RouteInfo = core.RouteInfo

	// RouteMeta represents a route with its metadata for extraction.
	RouteMeta = core.RouteMeta
)

// SchemaFrom generates a JSON Schema from a Go type using reflection.
func SchemaFrom[T any]() map[string]any {
	return core.SchemaFrom[T]()
}

// Helpers and utilities

var (
	// StringToInt converts a string to int.
	StringToInt = core.StringToInt

	// StringToString is a no-op converter for string path params.
	StringToString = core.StringToString

	// ParseID converts a string ID to int64.
	ParseID = core.ParseID

	// ParseIntID converts a string ID to int.
	ParseIntID = core.ParseIntID

	// ParseUUID converts a string ID to UUID [16]byte.
	ParseUUID = core.ParseUUID
)
