package endpoints

import (
	"context"
	"encoding/json"
	"net/http"

	routectx "github.com/reststore/restkit/internal/context"
	"github.com/reststore/restkit/internal/errors"
	"github.com/reststore/restkit/internal/middleware"
	"github.com/reststore/restkit/internal/schema"
	"github.com/reststore/restkit/internal/validation"
)

// Endpoint represents a unified API endpoint that can handle both request and response bodies,
// or either one individually using NoRequest or NoResponse as type parameters.
type Endpoint[Req any, Res any] struct {
	Title       string
	Summary     string
	Description string
	Method      string
	Path        string
	Scheme      string // "http", "https", "ws", "wss"

	pathParams []string

	Validate   func(ctx context.Context, req Req) ValidationResult
	Handler    func(ctx context.Context, req Req) (Res, error)
	Middleware []func(next http.Handler) http.Handler

	Bind    func(r *http.Request) (Req, error)
	Write   func(w http.ResponseWriter, res Res) error
	OnError func(w http.ResponseWriter, r *http.Request, err error)

	RequestSchema  map[string]any
	ResponseSchema map[string]any
	Parameters     []Parameter
}

// GetMethod returns the HTTP method for the endpoint.
func (e *Endpoint[Req, Res]) GetMethod() string {
	return e.Method
}

// GetPath returns the URL path for the endpoint.
func (e *Endpoint[Req, Res]) GetPath() string {
	return e.Path
}

// GetTitle returns the title of the endpoint.
func (e *Endpoint[Req, Res]) GetTitle() string {
	return e.Title
}

func (e *Endpoint[Req, Res]) GetSummary() string {
	return e.Summary
}

// GetDescription returns the description of the endpoint.
func (e *Endpoint[Req, Res]) GetDescription() string {
	return e.Description
}

// GetMiddleware returns the middleware chain for the endpoint.
func (e *Endpoint[Req, Res]) GetMiddleware() []func(http.Handler) http.Handler {
	return e.Middleware
}

// GetRequestSchema returns the JSON schema for the request body.
func (e *Endpoint[Req, Res]) GetRequestSchema() map[string]any {
	return e.RequestSchema
}

// GetResponseSchema returns the JSON schema for the response body.
func (e *Endpoint[Req, Res]) GetResponseSchema() map[string]any {
	return e.ResponseSchema
}

// GetParameters returns the OpenAPI parameters for this endpoint.
func (e *Endpoint[Req, Res]) GetParameters() []Parameter {
	if len(e.Parameters) > 0 {
		return e.Parameters
	}
	// Build path parameters from path pattern
	var params []Parameter
	for _, name := range e.pathParams {
		params = append(params, Parameter{
			Name:     name,
			Type:     "string",
			Required: true,
			Location: ParamLocationPath,
		})
	}
	return params
}

// GetScheme returns the protocol scheme for this endpoint (http, https, ws, wss).
func (e *Endpoint[Req, Res]) GetScheme() string {
	if e.Scheme != "" {
		return e.Scheme
	}
	return "http"
}

// WithParameters sets query/path parameters for OpenAPI documentation.
func (e *Endpoint[Req, Res]) WithParameters(params ...Parameter) *Endpoint[Req, Res] {
	e.Parameters = params
	return e
}

// WithTitle sets the title of the endpoint.
func (e *Endpoint[Req, Res]) WithTitle(title string) *Endpoint[Req, Res] {
	e.Title = title
	return e
}

// WithSummary sets the summary of the endpoint.
func (e *Endpoint[Req, Res]) WithSummary(summary string) *Endpoint[Req, Res] {
	e.Summary = summary
	return e
}

// WithDescription sets the description of the endpoint.
func (e *Endpoint[Req, Res]) WithDescription(description string) *Endpoint[Req, Res] {
	e.Description = description
	return e
}

// WithMethod sets the HTTP method for the endpoint.
func (e *Endpoint[Req, Res]) WithMethod(method string) *Endpoint[Req, Res] {
	e.Method = method
	return e
}

// WithPath sets the URL path for the endpoint.
func (e *Endpoint[Req, Res]) WithPath(path string) *Endpoint[Req, Res] {
	e.Path = path
	return e
}

// WithScheme sets the protocol scheme for the endpoint (http, https, ws, wss).
func (e *Endpoint[Req, Res]) WithScheme(scheme string) *Endpoint[Req, Res] {
	e.Scheme = scheme
	return e
}

// WithHandler sets the handler function for the endpoint.
func (e *Endpoint[Req, Res]) WithHandler(
	handler func(ctx context.Context, req Req) (Res, error),
) *Endpoint[Req, Res] {
	e.Handler = handler
	return e
}

// WithValidation sets a custom validation function for the endpoint.
func (e *Endpoint[Req, Res]) WithValidation(
	validate func(ctx context.Context, req Req) ValidationResult,
) *Endpoint[Req, Res] {
	e.Validate = validate
	return e
}

// WithBind sets a custom bind function for the endpoint.
func (e *Endpoint[Req, Res]) WithBind(bind func(r *http.Request) (Req, error)) *Endpoint[Req, Res] {
	e.Bind = bind
	return e
}

// WithWrite sets a custom write function for the endpoint.
func (e *Endpoint[Req, Res]) WithWrite(
	write func(w http.ResponseWriter, res Res) error,
) *Endpoint[Req, Res] {
	e.Write = write
	return e
}

// WithErrorHandler sets a custom error handler for the endpoint.
func (e *Endpoint[Req, Res]) WithErrorHandler(
	onError func(w http.ResponseWriter, r *http.Request, err error),
) *Endpoint[Req, Res] {
	e.OnError = onError
	return e
}

// WithMiddleware adds middleware to the endpoint.
func (e *Endpoint[Req, Res]) WithMiddleware(
	mw ...func(next http.Handler) http.Handler,
) *Endpoint[Req, Res] {
	e.Middleware = append(e.Middleware, mw...)
	return e
}

// WithRequestSchema sets a custom request schema for the endpoint.
func (e *Endpoint[Req, Res]) WithRequestSchema(schema map[string]any) *Endpoint[Req, Res] {
	e.RequestSchema = schema
	return e
}

// WithResponseSchema sets a custom response schema for the endpoint.
func (e *Endpoint[Req, Res]) WithResponseSchema(schema map[string]any) *Endpoint[Req, Res] {
	e.ResponseSchema = schema
	return e
}

// errorHandler creates a handler that always returns a specific error
func errorHandler(apiErr errors.APIError) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(apiErr.Status)
		_ = json.NewEncoder(w).Encode(apiErr)
	})
}

// GetHandler returns an HTTP handler for the endpoint.
func (e *Endpoint[Req, Res]) GetHandler() http.Handler {
	// Generate schemas early so they're available even if Handler is nil
	if e.RequestSchema == nil {
		e.RequestSchema = schema.SchemaFrom[Req]()
	}

	if e.ResponseSchema == nil {
		e.ResponseSchema = schema.SchemaFrom[Res]()
	}

	if e.Handler == nil {
		return errorHandler(errors.NewAPIError(
			http.StatusInternalServerError,
			errors.ErrCodeConfiguration,
			errors.ErrMsgHandlerNotSet,
		))
	}

	if e.Bind == nil {
		if isNoRequest[Req]() {
			e.Bind = func(r *http.Request) (Req, error) {
				var zero Req
				return zero, nil
			}
		} else if middleware.HasPathTag[Req]() || middleware.HasQueryTag[Req]() {
			e.Bind = middleware.QueryBinder[Req]()
		} else {
			e.Bind = middleware.JSONBinder[Req]()
		}
	}

	if e.Validate == nil {
		e.Validate = func(ctx context.Context, req Req) ValidationResult {
			if isNoRequest[Req]() {
				return ValidationResult{}
			}

			if v, ok := any(req).(ValidatableRequest); ok {
				return v.Validate(ctx)
			}

			return validation.Validate(ctx, req)
		}
	}

	if e.Write == nil {
		if isNoResponse[Res]() {
			e.Write = func(w http.ResponseWriter, _ Res) error {
				w.WriteHeader(http.StatusNoContent)
				return nil
			}
		} else {
			e.Write = middleware.JSONWriter[Res]()
		}
	}

	if e.OnError == nil {
		e.OnError = middleware.JSONErrorWriter
	}

	if e.pathParams == nil {
		e.pathParams = extractPathParamNames(e.Path)
	}

	var h http.Handler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var ctx context.Context = r.Context()

			routeCtx := routectx.NewRouteContext()

			// Extract path parameters
			if len(e.pathParams) > 0 {
				for _, paramName := range e.pathParams {
					value := r.PathValue(paramName)
					routeCtx.SetURLParam(paramName, value)
				}
			}

			// Extract query parameters
			if r.URL != nil && len(r.URL.Query()) > 0 {
				for key, values := range r.URL.Query() {
					if len(values) > 0 {
						routeCtx.SetURLQueryParam(key, values[0])
					}
				}
			}

			// Only add to context if we have any params
			if len(e.pathParams) > 0 || (r.URL != nil && len(r.URL.Query()) > 0) {
				ctx = context.WithValue(
					r.Context(),
					routectx.RouteCtxKey,
					routeCtx,
				)
				r = r.WithContext(ctx)
			}

			req, err := e.Bind(r)
			if err != nil {
				e.handleBindErr(w, r, err)
				return
			}

			if e.Validate != nil {
				result := e.Validate(ctx, req)
				if result.HasErrors() {
					e.handleValidation(w, r, result)
					return
				}
			}

			res, err := e.Handler(r.Context(), req)
			if err != nil {
				e.handleErr(w, r, err)
				return
			}

			if v := ctx.Value(routectx.SerializerCtxKey); v != nil {
				if serializer, ok := v.(func(http.ResponseWriter, any) error); ok {
					if err := serializer(w, res); err != nil {
						e.handleErr(w, r, err)
						return
					}
					return
				}
			}

			if err := e.Write(w, res); err != nil {
				e.handleErr(w, r, err)
				return
			}
		})

	for i := len(e.Middleware) - 1; i >= 0; i-- {
		h = e.Middleware[i](h)
	}

	return h
}

func (e *Endpoint[Req, Res]) handleBindErr(
	w http.ResponseWriter, r *http.Request, err error,
) {
	if e.OnError != nil {
		e.OnError(w, r, err)
		return
	}

	apiErr := errors.NewAPIErrorWithDetails(
		http.StatusBadRequest,
		errors.ErrCodeBind,
		errors.ErrMsgBind,
		err.Error(),
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.Status)
	_ = json.NewEncoder(w).Encode(apiErr)
}

func (e *Endpoint[Req, Res]) handleValidation(
	w http.ResponseWriter, _ *http.Request, result ValidationResult,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(result.Status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  result.Status,
		"code":    result.Code,
		"message": result.Message,
		"errors":  result.Errors,
	})
}

func (e *Endpoint[Req, Res]) handleErr(
	w http.ResponseWriter, _ *http.Request, err error,
) {
	if apiErr, ok := errors.IsAPIError(err); ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(apiErr.Status)
		_ = json.NewEncoder(w).Encode(apiErr)
		return
	}

	apiErr := errors.NewAPIError(
		http.StatusInternalServerError,
		errors.ErrCodeInternal,
		err.Error(),
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.Status)
	_ = json.NewEncoder(w).Encode(apiErr)
}

// Clone creates a copy of the endpoint with the same configuration.
func (e *Endpoint[Req, Res]) Clone() *Endpoint[Req, Res] {
	cp := *e

	if e.Middleware != nil {
		cp.Middleware = append(
			[]func(http.Handler) http.Handler(nil),
			e.Middleware...,
		)
	}

	return &cp
}

// setPath sets the path of the endpoint (unexported method for interface)
func (e *Endpoint[Req, Res]) setPath(path string) {
	e.Path = path
}

// addMiddleware adds middleware to the endpoint (unexported method for interface)
func (e *Endpoint[Req, Res]) addMiddleware(mw []func(http.Handler) http.Handler) {
	e.Middleware = append(e.Middleware, mw...)
}
