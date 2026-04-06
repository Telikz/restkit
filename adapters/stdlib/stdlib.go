package reststdlib

import (
	"net/http"

	"github.com/reststore/restkit/internal/api"
	"github.com/reststore/restkit/internal/docs"
)

func RegisterRoutes(mux *http.ServeMux, apiInstance *api.Api) {
	registered := make(map[string]bool)

	for _, group := range apiInstance.Groups {
		for _, endpoint := range group.GetEndpoints() {
			key := endpoint.GetMethod() + " " + endpoint.GetPath()
			registered[key] = true
			handler := endpoint.GetHandler()
			mux.Handle(
				endpoint.GetMethod()+" "+endpoint.GetPath(),
				handler,
			)
		}
	}

	for _, endpoint := range apiInstance.Endpoints {
		key := endpoint.GetMethod() + " " + endpoint.GetPath()
		if !registered[key] {
			handler := endpoint.GetHandler()
			for i := len(apiInstance.Middleware) - 1; i >= 0; i-- {
				handler = apiInstance.Middleware[i](handler)
			}
			mux.Handle(
				endpoint.GetMethod()+" "+endpoint.GetPath(),
				handler,
			)
		}
	}

	if apiInstance.SwaggerUIEnabled {
		mux.HandleFunc(
			"GET "+apiInstance.SwaggerUIPath,
			func(w http.ResponseWriter, r *http.Request) {
				docs.ServeSwaggerUI(w, apiInstance.SwaggerUIPath)
			},
		)
		mux.HandleFunc(
			"GET "+apiInstance.SwaggerUIPath+"/openapi.json",
			func(w http.ResponseWriter, r *http.Request) {
				apiInstance.ServeOpenAPI(w, r)
			},
		)
	}
}
