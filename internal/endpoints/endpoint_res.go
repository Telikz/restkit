package endpoints

import (
	"context"
	"encoding/json"
	"net/http"

	routectx "github.com/reststore/restkit/internal/context"
	"github.com/reststore/restkit/internal/errors"
	"github.com/reststore/restkit/internal/middleware"
	"github.com/reststore/restkit/internal/schema"
)

type EndpointRes[Res any] struct {
	Title       string
	Description string
	Method      string
	Path        string

	pathParams []string

	Handler    func(ctx context.Context) (Res, error)
	Validate   func(ctx context.Context) ValidationResult
	Write      func(w http.ResponseWriter, res Res) error
	OnError    func(w http.ResponseWriter, r *http.Request, err error)
	Middleware []func(next http.Handler) http.Handler

	ResponseSchema map[string]any
}

func (e *EndpointRes[Res]) GetMethod() string {
	return e.Method
}

func (e *EndpointRes[Res]) GetPath() string {
	return e.Path
}

func (e *EndpointRes[Res]) GetTitle() string {
	return e.Title
}

func (e *EndpointRes[Res]) GetDescription() string {
	return e.Description
}

func (e *EndpointRes[Res]) GetMiddleware() []func(http.Handler) http.Handler {
	return e.Middleware
}

func (e *EndpointRes[Res]) GetResponseSchema() map[string]any {
	return e.ResponseSchema
}

func (e *EndpointRes[Res]) GetRequestSchema() map[string]any {
	return nil
}

func (e *EndpointRes[Res]) WithTitle(title string) *EndpointRes[Res] {
	e.Title = title
	return e
}

func (e *EndpointRes[Res]) WithDescription(
	description string,
) *EndpointRes[Res] {
	e.Description = description
	return e
}

func (e *EndpointRes[Res]) WithMethod(method string) *EndpointRes[Res] {
	e.Method = method
	return e
}

func (e *EndpointRes[Res]) WithPath(path string) *EndpointRes[Res] {
	e.Path = path
	return e
}

func (e *EndpointRes[Res]) WithHandler(
	handler func(ctx context.Context) (Res, error),
) *EndpointRes[Res] {
	e.Handler = handler
	return e
}

func (e *EndpointRes[Res]) WithWrite(
	write func(w http.ResponseWriter, res Res) error,
) *EndpointRes[Res] {
	e.Write = write
	return e
}

func (e *EndpointRes[Res]) WithValidation(
	validate func(ctx context.Context) ValidationResult,
) *EndpointRes[Res] {
	e.Validate = validate
	return e
}

func (e *EndpointRes[Res]) WithMiddleware(
	middleware ...func(next http.Handler) http.Handler,
) *EndpointRes[Res] {
	e.Middleware = append(e.Middleware, middleware...)
	return e
}

func (e *EndpointRes[Res]) WithResponseSchema(
	schema map[string]any,
) *EndpointRes[Res] {
	e.ResponseSchema = schema
	return e
}

func (e *EndpointRes[Res]) GetHandler() http.Handler {
	if e.Handler == nil {
		return errorHandler(errors.NewAPIError(
			http.StatusInternalServerError,
			errors.ErrCodeConfiguration,
			errors.ErrMsgHandlerNotSet,
		))
	}
	if e.Write == nil {
		e.Write = middleware.JSONWriter[Res]()
	}
	if e.OnError == nil {
		e.OnError = middleware.JSONErrorWriter
	}
	if e.Method == "" {
		e.Method = http.MethodGet
	}
	if e.ResponseSchema == nil {
		e.ResponseSchema = schema.SchemaFrom[Res]()
	}

	if e.pathParams == nil {
		e.pathParams = extractPathParamNames(e.Path)
	}

	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		if e.Validate != nil {
			result := e.Validate(ctx)
			if result.HasErrors() {
				e.handleValidation(w, r, result)
				return
			}
		}

		res, err := e.Handler(r.Context())
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

func (e *EndpointRes[Res]) handleValidation(
	w http.ResponseWriter,
	_ *http.Request,
	result ValidationResult,
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

func (e *EndpointRes[Res]) handleErr(
	w http.ResponseWriter,
	_ *http.Request,
	err error,
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
