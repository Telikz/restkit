package api

import (
	"net/http"

	"github.com/telikz/restkit/internal/docs"
	"github.com/telikz/restkit/internal/endpoints"
	ep "github.com/telikz/restkit/internal/endpoints"
)

// API is the main struct for defining your API
type API struct {
	Version     string
	Title       string
	Description string

	Endpoints  []ep.Endpoint
	Groups     []*ep.Group
	Middleware []func(http.Handler) http.Handler

	SwaggerUIEnabled bool
	SwaggerUIPath    string
}

func New() *API {
	return &API{}
}

func (api *API) WithVersion(version string) *API {
	api.Version = version
	return api
}

func (api *API) WithTitle(title string) *API {
	api.Title = title
	return api
}

func (api *API) WithDescription(description string) *API {
	api.Description = description
	return api
}

func (api *API) Add(eps ...endpoints.Endpoint) *API {
	api.Endpoints = append(api.Endpoints, eps...)
	return api
}

func (api *API) AddGroup(group *endpoints.Group) *API {
	api.Groups = append(api.Groups, group)
	api.Endpoints = append(api.Endpoints, group.GetEndpoints()...)
	return api
}

func (api *API) WithSwaggerUI(enabled bool) *API {
	api.SwaggerUIEnabled = enabled
	if api.SwaggerUIPath == "" {
		api.SwaggerUIPath = "/swagger"
	}
	return api
}

func (api *API) WithSwaggerUIPath(path string) *API {
	api.SwaggerUIPath = path
	return api
}

func (api *API) WithMiddleware(middleware ...func(http.Handler) http.Handler) *API {
	api.Middleware = append(api.Middleware, middleware...)
	return api
}

func (api *API) Mux() http.Handler {
	mux := http.NewServeMux()

	for _, endpoint := range api.Endpoints {
		handler := endpoint.HTTPHandler()

		for i := len(api.Middleware) - 1; i >= 0; i-- {
			handler = api.Middleware[i](handler)
		}

		mux.Handle(endpoint.Pattern(), handler)
	}

	if api.SwaggerUIEnabled {
		mux.HandleFunc("GET "+api.SwaggerUIPath, api.serveSwaggerUI)
		mux.HandleFunc("GET "+api.SwaggerUIPath+"/openapi.json", api.ServeOpenAPI)
	}

	return mux
}

func (api *API) Serve(addr string) error {
	return http.ListenAndServe(addr, api.Mux())
}

// GenerateOpenAPI generates an OpenAPI spec by delegating to docs package
func (api *API) GenerateOpenAPI() map[string]any {
	return docs.GenerateOpenAPI(api.Title, api.Description, api.Version, api.Endpoints, api.Groups)
}

// ServeOpenAPI serves the OpenAPI spec as JSON
func (api *API) ServeOpenAPI(w http.ResponseWriter, r *http.Request) {
	spec := api.GenerateOpenAPI()
	docs.ServeOpenAPI(w, spec)
}

// serveSwaggerUI serves the Swagger UI HTML
func (api *API) serveSwaggerUI(w http.ResponseWriter, r *http.Request) {
	docs.ServeSwaggerUI(w, api.SwaggerUIPath)
}
