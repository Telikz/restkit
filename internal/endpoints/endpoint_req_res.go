package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"

	routectx "github.com/telikz/restkit/internal/context"
	"github.com/telikz/restkit/internal/errors"
	"github.com/telikz/restkit/internal/middleware"
	"github.com/telikz/restkit/internal/schema"
	"github.com/telikz/restkit/internal/validation"
)

var pathParamRegex = regexp.MustCompile(`\{([^}]+)\}`)

func extractPathParamNames(pattern string) []string {
	var names []string
	matches := pathParamRegex.FindAllStringSubmatch(pattern, -1)
	for _, match := range matches {
		names = append(names, match[1])
	}
	return names
}

type EndpointReqRes[Req any, Res any] struct {
	Title       string
	Description string
	Method      string
	Path        string

	pathParams []string

	Validate   func(ctx context.Context, req Req) ValidationResult
	Handler    func(ctx context.Context, req Req) (Res, error)
	Middleware []func(next http.Handler) http.Handler

	Bind    func(r *http.Request) (Req, error)
	Write   func(w http.ResponseWriter, res Res) error
	OnError func(w http.ResponseWriter, r *http.Request, err error)

	RequestSchema  map[string]any
	ResponseSchema map[string]any
}

func (e *EndpointReqRes[Req, Res]) GetMethod() string {
	return e.Method
}

func (e *EndpointReqRes[Req, Res]) GetPath() string {
	return e.Path
}

func (e *EndpointReqRes[Req, Res]) GetTitle() string {
	return e.Title
}

func (e *EndpointReqRes[Req, Res]) GetDescription() string {
	return e.Description
}

func (e *EndpointReqRes[Req, Res]) GetMiddleware() []func(http.Handler) http.Handler {
	return e.Middleware
}

func (e *EndpointReqRes[Req, Res]) GetRequestSchema() map[string]any {
	return e.RequestSchema
}

func (e *EndpointReqRes[Req, Res]) GetResponseSchema() map[string]any {
	return e.ResponseSchema
}

func (e *EndpointReqRes[Req, Res]) WithTitle(title string) *EndpointReqRes[Req, Res] {
	e.Title = title
	return e
}

func (e *EndpointReqRes[Req, Res]) WithDescription(description string) *EndpointReqRes[Req, Res] {
	e.Description = description
	return e
}

func (e *EndpointReqRes[Req, Res]) WithMethod(method string) *EndpointReqRes[Req, Res] {
	e.Method = method
	return e
}

func (e *EndpointReqRes[Req, Res]) WithPath(path string) *EndpointReqRes[Req, Res] {
	e.Path = path
	return e
}

func (e *EndpointReqRes[Req, Res]) WithHandler(
	handler func(ctx context.Context, req Req) (Res, error),
) *EndpointReqRes[Req, Res] {
	e.Handler = handler
	return e
}

func (e *EndpointReqRes[Req, Res]) WithValidation(
	validate func(ctx context.Context, req Req) ValidationResult,
) *EndpointReqRes[Req, Res] {
	e.Validate = validate
	return e
}

func (e *EndpointReqRes[Req, Res]) WithBind(
	bind func(r *http.Request) (Req, error),
) *EndpointReqRes[Req, Res] {
	e.Bind = bind
	return e
}

func (e *EndpointReqRes[Req, Res]) WithWrite(
	write func(w http.ResponseWriter, res Res) error,
) *EndpointReqRes[Req, Res] {
	e.Write = write
	return e
}

func (e *EndpointReqRes[Req, Res]) WithErrorHandler(
	onError func(w http.ResponseWriter, r *http.Request, err error),
) *EndpointReqRes[Req, Res] {
	e.OnError = onError
	return e
}
func (e *EndpointReqRes[Req, Res]) WithMiddleware(
	mw ...func(next http.Handler) http.Handler,
) *EndpointReqRes[Req, Res] {
	e.Middleware = append(e.Middleware, mw...)
	return e
}

func (e *EndpointReqRes[Req, Res]) WithRequestSchema(
	schema map[string]any) *EndpointReqRes[Req, Res] {
	e.RequestSchema = schema
	return e
}

func (e *EndpointReqRes[Req, Res]) WithResponseSchema(
	schema map[string]any) *EndpointReqRes[Req, Res] {
	e.ResponseSchema = schema
	return e
}

func (e *EndpointReqRes[Req, Res]) GetHandler() http.Handler {
	if e.Handler == nil {
		panic("endpoint handler is nil")
	}
	if e.Bind == nil {
		e.Bind = middleware.JSONBinder[Req]()
	}
	if e.Write == nil {
		e.Write = middleware.JSONWriter[Res]()
	}
	if e.OnError == nil {
		e.OnError = middleware.JSONErrorWriter
	}
	if e.Method == "" {
		e.Method = http.MethodPost
	}
	if e.Validate == nil {
		e.Validate = func(ctx context.Context, req Req) ValidationResult {
			return validation.ValidateStruct(ctx, req)
		}
	}
	if e.RequestSchema == nil {
		e.RequestSchema = schema.SchemaFrom[Req]()
	}
	if e.ResponseSchema == nil {
		e.ResponseSchema = schema.SchemaFrom[Res]()
	}

	if e.pathParams == nil {
		e.pathParams = extractPathParamNames(e.Path)
	}

	var h http.Handler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var ctx context.Context = r.Context()
			if len(e.pathParams) > 0 {
				routeCtx := routectx.NewRouteContext()
				for _, paramName := range e.pathParams {
					value := r.PathValue(paramName)
					routeCtx.SetURLParam(paramName, value)
				}
				ctx = context.WithValue(r.Context(), routectx.RouteCtxKey, routeCtx)
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

func (e *EndpointReqRes[Req, Res]) handleBindErr(
	w http.ResponseWriter, r *http.Request, err error) {
	if e.OnError != nil {
		e.OnError(w, r, err)
		return
	}

	apiErr := errors.NewAPIErrorWithDetails(
		http.StatusBadRequest,
		"bind",
		"Failed to parse request",
		err.Error(),
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.Status)
	json.NewEncoder(w).Encode(apiErr)
}

func (e *EndpointReqRes[Req, Res]) handleValidation(
	w http.ResponseWriter, _ *http.Request, result ValidationResult) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(result.Status)
	json.NewEncoder(w).Encode(map[string]any{
		"status":  result.Status,
		"code":    result.Code,
		"message": result.Message,
		"errors":  result.Errors,
	})
}

func (e *EndpointReqRes[Req, Res]) handleErr(
	w http.ResponseWriter, _ *http.Request, err error) {
	if apiErr, ok := errors.IsAPIError(err); ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(apiErr.Status)
		json.NewEncoder(w).Encode(apiErr)
		return
	}

	apiErr := errors.NewAPIError(
		http.StatusInternalServerError,
		"internal",
		err.Error(),
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.Status)
	json.NewEncoder(w).Encode(apiErr)
}

func (e *EndpointReqRes[Req, Res]) Clone() *EndpointReqRes[Req, Res] {
	cp := *e

	if e.Middleware != nil {
		cp.Middleware = append(
			[]func(http.Handler) http.Handler(nil),
			e.Middleware...,
		)
	}

	return &cp
}
