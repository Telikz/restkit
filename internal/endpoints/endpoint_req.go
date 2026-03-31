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

type EndpointReq[Req any] struct {
	Title       string
	Description string
	Method      string
	Path        string

	pathParams []string

	Handler    func(ctx context.Context, req Req) error
	Validate   func(ctx context.Context, req Req) ValidationResult
	Bind       func(r *http.Request) (Req, error)
	Write      func(w http.ResponseWriter, req Req) error
	OnError    func(w http.ResponseWriter, r *http.Request, err error)
	Middleware []func(next http.Handler) http.Handler

	RequestSchema map[string]any
}

func (e *EndpointReq[Req]) GetMethod() string {
	return e.Method
}

func (e *EndpointReq[Req]) GetPath() string {
	return e.Path
}

func (e *EndpointReq[Req]) GetTitle() string {
	return e.Title
}

func (e *EndpointReq[Req]) GetDescription() string {
	return e.Description
}

func (e *EndpointReq[Req]) GetMiddleware() []func(http.Handler) http.Handler {
	return e.Middleware
}

func (e *EndpointReq[Req]) GetRequestSchema() map[string]any {
	return e.RequestSchema
}

func (e *EndpointReq[Req]) GetResponseSchema() map[string]any {
	return nil
}

func (e *EndpointReq[Req]) WithTitle(title string) *EndpointReq[Req] {
	e.Title = title
	return e
}

func (e *EndpointReq[Req]) WithDescription(
	description string,
) *EndpointReq[Req] {
	e.Description = description
	return e
}

func (e *EndpointReq[Req]) WithMethod(method string) *EndpointReq[Req] {
	e.Method = method
	return e
}

func (e *EndpointReq[Req]) WithPath(path string) *EndpointReq[Req] {
	e.Path = path
	return e
}

func (e *EndpointReq[Req]) WithHandler(
	handler func(ctx context.Context, req Req) error,
) *EndpointReq[Req] {
	e.Handler = handler
	return e
}

func (e *EndpointReq[Req]) WithValidation(
	validate func(ctx context.Context, req Req) ValidationResult,
) *EndpointReq[Req] {
	e.Validate = validate
	return e
}

func (e *EndpointReq[Req]) WithBind(
	bind func(r *http.Request) (Req, error),
) *EndpointReq[Req] {
	e.Bind = bind
	return e
}

func (e *EndpointReq[Req]) WithOnError(
	onError func(w http.ResponseWriter, r *http.Request, err error),
) *EndpointReq[Req] {
	e.OnError = onError
	return e
}

func (e *EndpointReq[Req]) WithWrite(
	write func(w http.ResponseWriter, req Req) error,
) *EndpointReq[Req] {
	e.Write = write
	return e
}

func (e *EndpointReq[Req]) WithRequestSchema(
	schema map[string]any,
) *EndpointReq[Req] {
	e.RequestSchema = schema
	return e
}

func (e *EndpointReq[Req]) GetHandler() http.Handler {
	if e.Handler == nil {
		return errorHandler(errors.NewAPIError(
			http.StatusInternalServerError,
			errors.ErrCodeConfiguration,
			errors.ErrMsgHandlerNotSet,
		))
	}
	if e.Bind == nil {
		e.Bind = middleware.JSONBinder[Req]()
	}
	if e.OnError == nil {
		e.OnError = middleware.JSONErrorWriter
	}
	if e.Method == "" {
		e.Method = http.MethodDelete
	}
	if e.Validate == nil {
		e.Validate = func(ctx context.Context, req Req) ValidationResult {
			return validation.ValidateStruct(ctx, req)
		}
	}
	if e.RequestSchema == nil {
		e.RequestSchema = schema.SchemaFrom[Req]()
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

		if err := e.Handler(r.Context(), req); err != nil {
			e.handleErr(w, r, err)
			return
		}

		if e.Write != nil {
			if err := e.Write(w, req); err != nil {
				e.handleErr(w, r, err)
				return
			}
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	})

	for i := len(e.Middleware) - 1; i >= 0; i-- {
		h = e.Middleware[i](h)
	}

	return h
}

func (e *EndpointReq[Req]) handleBindErr(
	w http.ResponseWriter,
	r *http.Request,
	err error,
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
	json.NewEncoder(w).Encode(apiErr)
}

func (e *EndpointReq[Req]) handleValidation(
	w http.ResponseWriter,
	r *http.Request,
	result ValidationResult,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(result.Status)
	// Convert ValidationResult to APIError format
	json.NewEncoder(w).Encode(map[string]any{
		"status":  result.Status,
		"code":    result.Code,
		"message": result.Message,
		"errors":  result.Errors,
	})
}

func (e *EndpointReq[Req]) handleErr(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	if apiErr, ok := errors.IsAPIError(err); ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(apiErr.Status)
		json.NewEncoder(w).Encode(apiErr)
		return
	}

	apiErr := errors.NewAPIError(
		http.StatusInternalServerError,
		errors.ErrCodeInternal,
		err.Error(),
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.Status)
	json.NewEncoder(w).Encode(apiErr)
}

func (e EndpointReq[Req]) Clone() *EndpointReq[Req] {
	cp := e

	if e.Middleware != nil {
		cp.Middleware = append(
			[]func(http.Handler) http.Handler(nil),
			e.Middleware...,
		)
	}

	return &cp
}
