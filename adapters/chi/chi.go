// Package restchi provides an adapter to register RestKit API endpoints with the Chi router.
package restchi

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/reststore/restkit/internal/api"
	routectx "github.com/reststore/restkit/internal/context"
	"github.com/reststore/restkit/internal/docs"
)

// RegisterRoutes registers all routes from a RestKit API with Chi router.
func RegisterRoutes(r chi.Router, apiInstance *api.Api) {
	registered := make(map[string]bool)

	configMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Inject validator if set
			if apiInstance.Validator != nil {
				ctx = context.WithValue(ctx, routectx.ValidatorCtxKey, apiInstance.Validator)
			}

			// Inject serializer if set
			if apiInstance.Serializer != nil {
				ctx = context.WithValue(ctx, routectx.SerializerCtxKey, apiInstance.Serializer)
			}

			// Inject deserializer if set
			if apiInstance.Deserializer != nil {
				ctx = context.WithValue(ctx, routectx.DeserializerCtxKey, apiInstance.Deserializer)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	for _, group := range apiInstance.Groups {
		for _, endpoint := range group.GetEndpoints() {
			key := fmt.Sprintf("%s %s", endpoint.GetMethod(), endpoint.GetPath())
			registered[key] = true
			handler := endpoint.GetHandler()
			handler = configMiddleware(handler)
			r.MethodFunc(
				endpoint.GetMethod(),
				endpoint.GetPath(),
				handler.ServeHTTP,
			)
		}
	}

	for _, endpoint := range apiInstance.Endpoints {
		key := fmt.Sprintf("%s %s", endpoint.GetMethod(), endpoint.GetPath())
		if !registered[key] {
			handler := endpoint.GetHandler()
			handler = configMiddleware(handler)
			for i := len(apiInstance.Middleware) - 1; i >= 0; i-- {
				handler = apiInstance.Middleware[i](handler)
			}
			r.MethodFunc(
				endpoint.GetMethod(),
				endpoint.GetPath(),
				handler.ServeHTTP,
			)
		}
	}

	// Register Swagger UI if enabled
	if apiInstance.SwaggerUIEnabled {
		r.Get(
			apiInstance.SwaggerUIPath,
			func(w http.ResponseWriter, r *http.Request) {
				docs.ServeSwaggerUI(w, apiInstance.SwaggerUIPath)
			},
		)
		r.Get(
			apiInstance.SwaggerUIPath+"/openapi.json",
			apiInstance.ServeOpenAPI,
		)
	}
}
