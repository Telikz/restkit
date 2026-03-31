package endpoints

import "net/http"

// Group represents a collection of related endpoints with shared configuration
type Group struct {
	Prefix      string
	Title       string
	Description string
	Endpoints   []Endpoint
	Middleware  []func(http.Handler) http.Handler
}

// NewGroup creates a new endpoint group with a path prefix
func NewGroup(prefix string) *Group {
	return &Group{
		Prefix:    prefix,
		Endpoints: []Endpoint{},
	}
}

// WithEndpoints adds endpoints to the group
func (g *Group) WithEndpoints(endpoints ...Endpoint) *Group {
	for _, ep := range endpoints {
		g.Endpoints = append(g.Endpoints, ep)
	}
	return g
}

// WithTitle sets the title of the group
func (g *Group) WithTitle(title string) *Group {
	g.Title = title
	return g
}

// WithDescription sets the description of the group
func (g *Group) WithDescription(description string) *Group {
	g.Description = description
	return g
}

// WithMiddleware adds middleware to all endpoints in the group
func (g *Group) WithMiddleware(
	middleware ...func(http.Handler) http.Handler,
) *Group {
	g.Middleware = append(g.Middleware, middleware...)
	return g
}

// GetEndpoints returns all endpoints with group prefix and middleware applied
func (g *Group) GetEndpoints() []Endpoint {
	endpoints := make([]Endpoint, len(g.Endpoints))
	for i, ep := range g.Endpoints {
		prefixed := g.prefixEndpoint(ep)
		if len(g.Middleware) > 0 {
			endpoints[i] = g.wrapEndpointWithMiddleware(prefixed)
		} else {
			endpoints[i] = prefixed
		}
	}
	return endpoints
}

// prefixEndpoint adds the group prefix to an endpoint's path
func (g *Group) prefixEndpoint(ep Endpoint) Endpoint {
	path := g.Prefix + ep.GetPath()

	return &endpointWrapper{
		definition:   ep,
		pathOverride: path,
	}
}

// wrapEndpointWithMiddleware wraps an endpoint with group middleware
func (g *Group) wrapEndpointWithMiddleware(ep Endpoint) Endpoint {
	if wrapper, ok := ep.(*endpointWrapper); ok {
		allMiddleware := append(wrapper.middlewareOverride, g.Middleware...)
		return &endpointWrapper{
			definition:         wrapper.definition,
			pathOverride:       wrapper.pathOverride,
			middlewareOverride: allMiddleware,
		}
	}

	return &endpointWrapper{
		definition:         ep,
		middlewareOverride: g.Middleware,
	}
}

// endpointWrapper wraps a Endpoint to override certain methods
type endpointWrapper struct {
	definition         Endpoint
	pathOverride       string
	middlewareOverride []func(http.Handler) http.Handler
}

func (w *endpointWrapper) GetMethod() string {
	return w.definition.GetMethod()
}

func (w *endpointWrapper) GetPath() string {
	if w.pathOverride != "" {
		return w.pathOverride
	}
	return w.definition.GetPath()
}

func (w *endpointWrapper) GetTitle() string {
	return w.definition.GetTitle()
}

func (w *endpointWrapper) GetDescription() string {
	return w.definition.GetDescription()
}

func (w *endpointWrapper) GetMiddleware() []func(http.Handler) http.Handler {
	return w.definition.GetMiddleware()
}

func (w *endpointWrapper) GetRequestSchema() map[string]any {
	return w.definition.GetRequestSchema()
}

func (w *endpointWrapper) GetResponseSchema() map[string]any {
	return w.definition.GetResponseSchema()
}

func (w *endpointWrapper) GetHandler() http.Handler {
	handler := w.definition.GetHandler()

	if w.middlewareOverride != nil {
		for i := len(w.middlewareOverride) - 1; i >= 0; i-- {
			handler = w.middlewareOverride[i](handler)
		}
	}

	return handler
}
