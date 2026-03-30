package api

import (
	"net/http"

	"github.com/telikz/restkit/internal/docs"
	"github.com/telikz/restkit/internal/endpoints"
	ep "github.com/telikz/restkit/internal/endpoints"
)

// Api is the main struct for defining your API
type Api struct {
	Version     string
	Title       string
	Description string

	Endpoints  []ep.Endpoint
	Groups     []*ep.Group
	Middleware []func(http.Handler) http.Handler

	SwaggerUIEnabled bool
	SwaggerUIPath    string
}

func New() *Api {
	return &Api{}
}

func (api *Api) WithVersion(version string) *Api {
	api.Version = version
	return api
}

func (api *Api) WithTitle(title string) *Api {
	api.Title = title
	return api
}

func (api *Api) WithDescription(description string) *Api {
	api.Description = description
	return api
}

func (api *Api) Add(eps ...endpoints.Endpoint) *Api {
	api.Endpoints = append(api.Endpoints, eps...)
	return api
}

func (api *Api) AddGroup(group *endpoints.Group) *Api {
	api.Groups = append(api.Groups, group)
	api.Endpoints = append(api.Endpoints, group.GetEndpoints()...)
	return api
}

func (api *Api) WithSwaggerUI(enabled bool) *Api {
	api.SwaggerUIEnabled = enabled
	if api.SwaggerUIPath == "" {
		api.SwaggerUIPath = "/swagger"
	}
	return api
}

func (api *Api) WithSwaggerUIPath(path string) *Api {
	api.SwaggerUIPath = path
	return api
}

func (api *Api) WithMiddleware(middleware ...func(http.Handler) http.Handler) *Api {
	api.Middleware = append(api.Middleware, middleware...)
	return api
}

func (api *Api) Mux() http.Handler {
	mux := http.NewServeMux()

	for _, endpoint := range api.Endpoints {
		handler := endpoint.GetHandler()

		for i := len(api.Middleware) - 1; i >= 0; i-- {
			handler = api.Middleware[i](handler)
		}

		mux.Handle(endpoint.GetMethod()+" "+endpoint.GetPath(), handler)
	}

	if api.SwaggerUIEnabled {
		mux.HandleFunc("GET "+api.SwaggerUIPath, api.serveSwaggerUI)
		mux.HandleFunc("GET "+api.SwaggerUIPath+"/openapi.json", api.ServeOpenAPI)
	}

	return mux
}

func (api *Api) Serve(addr string) error {
	return http.ListenAndServe(addr, api.Mux())
}

// GenerateOpenAPI generates an OpenAPI spec by delegating to docs package
func (api *Api) GenerateOpenAPI() map[string]any {
	return docs.GenerateOpenAPI(api.Title, api.Description, api.Version, api.Endpoints, api.Groups)
}

// ServeOpenAPI serves the OpenAPI spec as JSON
func (api *Api) ServeOpenAPI(w http.ResponseWriter, r *http.Request) {
	spec := api.GenerateOpenAPI()
	docs.ServeOpenAPI(w, spec)
}

// serveSwaggerUI serves the Swagger UI HTML
func (api *Api) serveSwaggerUI(w http.ResponseWriter, r *http.Request) {
	docs.ServeSwaggerUI(w, api.SwaggerUIPath)
}
