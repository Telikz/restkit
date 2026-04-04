// Package restchi provides an adapter to register RestKit API endpoints with the Chi router.
package restchi

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/reststore/restkit/internal/api"
	"github.com/reststore/restkit/internal/docs"
)

// RegisterRoutes registers all routes from a RestKit API with Chi router.
func RegisterRoutes(r chi.Router, api *api.Api) {
	registered := make(map[string]bool)

	for _, group := range api.Groups {
		for _, endpoint := range group.GetEndpoints() {
			key := fmt.Sprintf("%s %s", endpoint.GetMethod(), endpoint.GetPath())
			registered[key] = true
			r.MethodFunc(
				endpoint.GetMethod(),
				endpoint.GetPath(),
				endpoint.GetHandler().ServeHTTP,
			)
		}
	}

	for _, endpoint := range api.Endpoints {
		key := fmt.Sprintf("%s %s", endpoint.GetMethod(), endpoint.GetPath())
		if !registered[key] {
			handler := endpoint.GetHandler()
			for i := len(api.Middleware) - 1; i >= 0; i-- {
				handler = api.Middleware[i](handler)
			}
			r.MethodFunc(
				endpoint.GetMethod(),
				endpoint.GetPath(),
				handler.ServeHTTP,
			)
		}
	}

	// Register Swagger UI if enabled
	if api.SwaggerUIEnabled {
		r.Get(
			api.SwaggerUIPath,
			func(w http.ResponseWriter, r *http.Request) {
				docs.ServeSwaggerUI(w, api.SwaggerUIPath)
			},
		)
		r.Get(
			api.SwaggerUIPath+"/openapi.json",
			api.ServeOpenAPI,
		)
	}
}
