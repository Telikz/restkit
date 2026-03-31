package endpoints

import (
	"context"
	"net/http"

	"github.com/RestStore/RestKit/internal/middleware"
	"github.com/RestStore/RestKit/internal/schema"
	"github.com/RestStore/RestKit/internal/validation"
)

// NewEndpoint creates a new Endpoint with auto-generated schemas and sensible defaults
// Default method is POST, default error handler uses JSONErrorWriter, default write uses JSONWriter
// Default validation uses go-playground/validator with struct tags
func NewEndpoint[Req any, Res any]() *EndpointReqRes[Req, Res] {
	return &EndpointReqRes[Req, Res]{
		Title:       "Not set",
		Description: "Not set",

		Path:    "/not-set",
		Method:  http.MethodPost,
		Handler: nil,

		Bind:     middleware.JSONBinder[Req](),
		Write:    middleware.JSONWriter[Res](),
		OnError:  middleware.JSONErrorWriter,
		Validate: defaultValidator[Req](),

		RequestSchema:  schema.SchemaFrom[Req](),
		ResponseSchema: schema.SchemaFrom[Res](),
	}
}

// NewEndpointRes creates a new EndpointRes with auto-generated response schema and defaults
// Default method is GET, default error handler uses JSONErrorWriter, default write uses JSONWriter
func NewEndpointRes[Res any]() *EndpointRes[Res] {
	return &EndpointRes[Res]{
		Title:       "Not set",
		Description: "Not set",

		Path:    "/not-set",
		Method:  http.MethodGet,
		Handler: nil,

		Write:   middleware.JSONWriter[Res](),
		OnError: middleware.JSONErrorWriter,

		ResponseSchema: schema.SchemaFrom[Res](),
	}
}

// NewEndpointReq creates a new EndpointReq with auto-generated request schema and defaults
// Default method is DELETE, default error handler uses JSONErrorWriter, default bind uses JSONBinder
// Default validation uses go-playground/validator with struct tags
func NewEndpointReq[Req any]() *EndpointReq[Req] {
	return &EndpointReq[Req]{
		Title:       "Not set",
		Description: "Not set",

		Path:    "/not-set",
		Method:  http.MethodDelete,
		Handler: nil,

		Bind:     middleware.JSONBinder[Req](),
		OnError:  middleware.JSONErrorWriter,
		Validate: defaultValidator[Req](),

		RequestSchema: schema.SchemaFrom[Req](),
	}
}

// defaultValidator returns a validation function that uses go-playground/validator
func defaultValidator[Req any]() func(context.Context, Req) ValidationResult {
	return func(ctx context.Context, req Req) ValidationResult {
		return validation.ValidateStruct(ctx, req)
	}
}
