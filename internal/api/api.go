package api

import (
	"net/http"

	"github.com/reststore/restkit/internal/docs"
	ep "github.com/reststore/restkit/internal/endpoints"
	"github.com/reststore/restkit/internal/schema"
)

// Api is the main struct for defining your API
type Api struct {
	Version     string
	Title       string
	Description string

	Endpoints      []ep.Route
	Groups         []*ep.Group
	Middleware     []func(http.Handler) http.Handler
	MountedRouters []schema.MountedRouter

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

// AddEndpoint adds one or more endpoints to the API
func (api *Api) AddEndpoint(eps ...any) *Api {
	for _, e := range eps {
		if endpoint, ok := e.(ep.Route); ok {
			api.Endpoints = append(api.Endpoints, endpoint)
		}
	}
	return api
}

func (api *Api) AddGroup(group *ep.Group) *Api {
	api.Groups = append(api.Groups, group)
	for _, e := range group.GetEndpoints() {
		api.Endpoints = append(api.Endpoints, e)
	}
	return api
}

func (api *Api) WithSwaggerUI(path ...string) *Api {
	api.SwaggerUIEnabled = true
	if len(path) > 0 && path[0] != "" {
		api.SwaggerUIPath = path[0]
	} else if api.SwaggerUIPath == "" {
		api.SwaggerUIPath = "/swagger"
	}
	return api
}

// WithSwaggerUIPath sets the Swagger UI path (deprecated: use WithSwaggerUI(path) instead)
func (api *Api) WithSwaggerUIPath(path string) *Api {
	api.SwaggerUIPath = path
	return api
}

func (api *Api) WithMiddleware(
	middleware ...func(http.Handler) http.Handler,
) *Api {
	api.Middleware = append(api.Middleware, middleware...)
	return api
}

// MountRouter mounts an external router (e.g., Chi, Gin) into the RestKit API
// with route definitions for OpenAPI documentation. The prefix is prepended to all routes.
func (api *Api) MountRouter(
	prefix string,
	router http.Handler,
	routes []schema.MountedRoute,
) *Api {
	api.MountedRouters = append(
		api.MountedRouters,
		schema.MountedRouter{
			Prefix: prefix,
			Router: router,
			Routes: routes,
		},
	)
	return api
}

func (api *Api) Mux() http.Handler {
	mux := http.NewServeMux()

	// Register RestKit endpoints
	for _, endpoint := range api.Endpoints {
		handler := endpoint.GetHandler()

		for i := len(api.Middleware) - 1; i >= 0; i-- {
			handler = api.Middleware[i](handler)
		}

		mux.Handle(
			endpoint.GetMethod()+" "+endpoint.GetPath(),
			handler,
		)
	}

	// Register mounted routers
	for _, mounted := range api.MountedRouters {
		if mounted.Prefix == "" || mounted.Prefix == "/" {
			mux.Handle("/", mounted.Router)
		} else {
			mux.Handle(mounted.Prefix+"/", http.StripPrefix(mounted.Prefix, mounted.Router))
		}
	}

	if api.SwaggerUIEnabled {
		mux.HandleFunc("GET "+api.SwaggerUIPath, api.serveSwaggerUI)
		mux.HandleFunc("GET "+api.SwaggerUIPath+"/openapi.json",
			api.ServeOpenAPI,
		)
	}

	return mux
}

func (api *Api) Serve(addr string) error {
	return http.ListenAndServe(addr, api.Mux())
}

// GenerateOpenAPI generates an OpenAPI spec including both RestKit endpoints and mounted routes
func (api *Api) GenerateOpenAPI() map[string]any {
	spec := docs.GenerateOpenAPI(
		api.Title,
		api.Description,
		api.Version,
		api.Endpoints,
		api.Groups,
	)

	// Add mounted router routes to the OpenAPI spec
	for _, mounted := range api.MountedRouters {
		docs.AddMountedRoutesToSpec(
			spec,
			mounted.Prefix,
			mounted.Routes,
		)
	}

	return spec
}

// ServeOpenAPI serves the OpenAPI spec as JSON
func (api *Api) ServeOpenAPI(w http.ResponseWriter, _ *http.Request) {
	spec := api.GenerateOpenAPI()
	docs.ServeOpenAPI(w, spec)
}

// serveSwaggerUI serves the Swagger UI HTML
func (api *Api) serveSwaggerUI(
	w http.ResponseWriter,
	_ *http.Request,
) {
	docs.ServeSwaggerUI(w, api.SwaggerUIPath)
}
