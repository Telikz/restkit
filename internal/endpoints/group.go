package endpoints

import "net/http"

type Group struct {
	Prefix      string
	Title       string
	Description string
	Endpoints   []Route
	Middleware  []func(http.Handler) http.Handler
}

func NewGroup(prefix string) *Group {
	return &Group{
		Prefix:    prefix,
		Endpoints: []Route{},
	}
}

func (g *Group) WithEndpoints(endpoints ...any) *Group {
	for _, ep := range endpoints {
		if e, ok := ep.(Route); ok {
			g.Endpoints = append(g.Endpoints, e)
		}
	}
	return g
}

func (g *Group) WithTitle(title string) *Group {
	g.Title = title
	return g
}

func (g *Group) WithDescription(description string) *Group {
	g.Description = description
	return g
}

func (g *Group) WithMiddleware(
	middleware ...func(http.Handler) http.Handler,
) *Group {
	g.Middleware = append(g.Middleware, middleware...)
	return g
}

func (g *Group) GetEndpoints() []Route {
	endpoints := make([]Route, len(g.Endpoints))
	for i, ep := range g.Endpoints {
		endpoints[i] = &routeInfo{
			route:      ep,
			prefix:     g.Prefix,
			middleware: g.Middleware,
		}
	}
	return endpoints
}
