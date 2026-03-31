// Package restchi provides an adapter to register RestKit API endpoints with the Chi router.
package restchi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/reststore/restkit/internal/api"
	"github.com/reststore/restkit/internal/docs"
	"github.com/reststore/restkit/internal/endpoints"
)

// RegisterRoutes registers all routes from a RestKit API with Chi router.
func RegisterRoutes(r chi.Router, api *api.Api) {
	for _, group := range api.Groups {
		for _, endpoint := range group.GetEndpoints() {
			r.MethodFunc(
				endpoint.GetMethod(),
				endpoint.GetPath(),
				endpoint.GetHandler().ServeHTTP,
			)
		}
	}

	registered := make(map[endpoints.Endpoint]bool)
	for _, group := range api.Groups {
		for _, ep := range group.GetEndpoints() {
			registered[ep] = true
		}
	}

	for _, endpoint := range api.Endpoints {
		if !registered[endpoint] {
			r.MethodFunc(
				endpoint.GetMethod(),
				endpoint.GetPath(),
				endpoint.GetHandler().ServeHTTP,
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
