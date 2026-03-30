package endpoints

import "net/http"

// Group represents a collection of related endpoints with shared configuration
type Group struct {
	prefix      string
	title       string
	description string
	endpoints   []Endpoint
	middleware  []func(http.Handler) http.Handler
}

// NewGroup creates a new endpoint group with a path prefix
func NewGroup(prefix string) *Group {
	return &Group{
		prefix:    prefix,
		endpoints: []Endpoint{},
	}
}

// WithEndpoints adds endpoints to the group
func (g *Group) WithEndpoints(endpoints ...Endpoint) *Group {
	for _, ep := range endpoints {
		prefixedEndpoint := g.prefixEndpoint(ep)
		g.endpoints = append(g.endpoints, prefixedEndpoint)
	}
	return g
}

// WithTitle sets the title of the group
func (g *Group) WithTitle(title string) *Group {
	g.title = title
	return g
}

// WithDescription sets the description of the group
func (g *Group) WithDescription(description string) *Group {
	g.description = description
	return g
}

// WithMiddleware adds middleware to all endpoints in the group
func (g *Group) WithMiddleware(middleware ...func(http.Handler) http.Handler) *Group {
	g.middleware = append(g.middleware, middleware...)
	return g
}

// GetTitle returns the title of the group
func (g *Group) GetTitle() string {
	return g.title
}

// GetDescription returns the description of the group
func (g *Group) GetDescription() string {
	return g.description
}

// GetEndpoints returns all endpoints with group middleware applied
func (g *Group) GetEndpoints() []Endpoint {
	if len(g.middleware) == 0 {
		return g.endpoints
	}

	// Wrap endpoints with group middleware
	wrappedEndpoints := make([]Endpoint, len(g.endpoints))
	for i, ep := range g.endpoints {
		wrappedEndpoints[i] = g.wrapEndpointWithMiddleware(ep)
	}
	return wrappedEndpoints
}

// prefixEndpoint adds the group prefix to an endpoint's path
func (g *Group) prefixEndpoint(ep Endpoint) Endpoint {
	path := g.prefix + ep.GetPath()
	method := ep.GetMethod()

	return &endpointWrapper{
		definition:      ep,
		pathOverride:    path,
		patternOverride: method + " " + path,
	}
}

// wrapEndpointWithMiddleware wraps an endpoint with group middleware
func (g *Group) wrapEndpointWithMiddleware(ep Endpoint) Endpoint {
	if wrapper, ok := ep.(*endpointWrapper); ok {
		allMiddleware := append(wrapper.middlewareOverride, g.middleware...)
		return &endpointWrapper{
			definition:         wrapper.definition,
			pathOverride:       wrapper.pathOverride,
			patternOverride:    wrapper.patternOverride,
			middlewareOverride: allMiddleware,
		}
	}

	return &endpointWrapper{
		definition:         ep,
		middlewareOverride: g.middleware,
	}
}

// endpointWrapper wraps a Endpoint to override certain methods
type endpointWrapper struct {
	definition         Endpoint
	pathOverride       string
	patternOverride    string
	middlewareOverride []func(http.Handler) http.Handler
}

func (w *endpointWrapper) Pattern() string {
	if w.patternOverride != "" {
		return w.patternOverride
	}
	return w.definition.Pattern()
}

func (w *endpointWrapper) HTTPHandler() http.Handler {
	handler := w.definition.HTTPHandler()

	if w.middlewareOverride != nil {
		for i := len(w.middlewareOverride) - 1; i >= 0; i-- {
			handler = w.middlewareOverride[i](handler)
		}
	}

	return handler
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
