package api

import (
	"context"
	"net/http"

	routectx "github.com/reststore/restkit/internal/context"
	"github.com/reststore/restkit/internal/docs"
	ep "github.com/reststore/restkit/internal/endpoints"
	errs "github.com/reststore/restkit/internal/errors"
	"github.com/reststore/restkit/internal/schema"
	"github.com/reststore/restkit/internal/validation"
)

// Api is the main struct for defining your API
type Api struct {
	Version string

	Title          string
	Summary        string
	Description    string
	TermsOfService string

	Contact struct {
		Name  string
		URL   string
		Email string
	}

	License struct {
		Name string
		URL  string
	}

	Servers []docs.Server

	Endpoints      []ep.Route
	Groups         []*ep.Group
	Middleware     []func(http.Handler) http.Handler
	MountedRouters []schema.MountedRouter

	SwaggerUIEnabled bool
	SwaggerUIPath    string

	Validator    func(ctx context.Context, s any) errs.ValidationResult
	Serializer   func(w http.ResponseWriter, res any) error
	Deserializer func(r *http.Request, req any) error

	tlsCertFile string
	tlsKeyFile  string
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

func (api *Api) WithSummary(summary string) *Api {
	api.Summary = summary
	return api
}

func (api *Api) WithDescription(description string) *Api {
	api.Description = description
	return api
}

func (api *Api) WithTermsOfService(tos string) *Api {
	api.TermsOfService = tos
	return api
}

func (api *Api) WithContact(name, url, email string) *Api {
	api.Contact.Name = name
	api.Contact.URL = url
	api.Contact.Email = email
	return api
}

func (api *Api) WithLicense(name, url string) *Api {
	api.License.Name = name
	api.License.URL = url
	return api
}

// WithServer adds a server to the API
func (api *Api) WithServer(url, description string, variables map[string]struct{ Default, Description string }) *Api {
	api.Servers = append(api.Servers, docs.Server{
		URL:         url,
		Description: description,
		Variables:   variables,
	})
	return api
}

// WithServers adds one or more servers to the API
func (api *Api) WithServers(servers ...docs.Server) *Api {
	api.Servers = append(api.Servers, servers...)
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

func (api *Api) WithValidator(
	validator func(ctx context.Context, s any) errs.ValidationResult,
) *Api {
	api.Validator = validator
	validation.DefaultValidator = validator
	return api
}

func (api *Api) WithSerializer(
	serializer func(w http.ResponseWriter, res any) error,
) *Api {
	api.Serializer = serializer
	return api
}

func (api *Api) WithDeserializer(
	deserializer func(r *http.Request, req any) error,
) *Api {
	api.Deserializer = deserializer
	return api
}

func (api *Api) WithTLS(certFile, keyFile string) *Api {
	api.tlsCertFile = certFile
	api.tlsKeyFile = keyFile
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

	// Create API configuration injector middleware
	configMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Inject validator if set
			if api.Validator != nil {
				ctx = context.WithValue(ctx, routectx.ValidatorCtxKey, api.Validator)
			}

			// Inject serializer if set
			if api.Serializer != nil {
				ctx = context.WithValue(ctx, routectx.SerializerCtxKey, api.Serializer)
			}

			// Inject deserializer if set
			if api.Deserializer != nil {
				ctx = context.WithValue(ctx, routectx.DeserializerCtxKey, api.Deserializer)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	// Register RestKit endpoints
	for _, endpoint := range api.Endpoints {
		handler := endpoint.GetHandler()

		// Wrap with config middleware first (so it's available in handler)
		handler = configMiddleware(handler)

		for i := len(api.Middleware) - 1; i >= 0; i-- {
			handler = api.Middleware[i](handler)
		}

		mux.Handle(
			endpoint.GetMethod()+" "+endpoint.GetPath(),
			handler,
		)
	}

	// Register mounted routers (wrap with config middleware)
	for _, mounted := range api.MountedRouters {
		router := configMiddleware(mounted.Router)
		if mounted.Prefix == "" || mounted.Prefix == "/" {
			mux.Handle("/", router)
		} else {
			mux.Handle(mounted.Prefix+"/", http.StripPrefix(mounted.Prefix, router))
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

func (api *Api) ServeTLS(addr, certFile, keyFile string) error {
	cert := certFile
	key := keyFile
	if cert == "" {
		cert = api.tlsCertFile
	}
	if key == "" {
		key = api.tlsKeyFile
	}
	if cert == "" || key == "" {
		return errs.NewAPIError(
			http.StatusInternalServerError,
			errs.ErrCodeConfiguration,
			"TLS certificate and key files must be provided via WithTLS() or ServeTLS()",
		)
	}
	return http.ListenAndServeTLS(addr, cert, key, api.Mux())
}

// GenerateOpenAPI generates an OpenAPI spec including both RestKit endpoints and mounted routes
func (api *Api) GenerateOpenAPI() map[string]any {
	s := &docs.OpenAPISpec{
		Version:     api.Version,
		Title:       api.Title,
		Summary:     api.Summary,
		Description: api.Description,
		Endpoints:   api.Endpoints,
		Groups:      api.Groups,
		Servers:     api.Servers,
	}

	if len(s.Servers) == 0 {
		s.Servers = append(s.Servers, docs.Server{URL: "http://localhost:8080", Description: "Dev Server"})
	}

	spec := docs.GenerateOpenAPI(s)

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
