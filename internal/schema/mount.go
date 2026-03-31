package schema

import (
	"net/http"
)

// MountedRouter represents a router that has been mounted to the API with its prefix and extracted routes.
type MountedRouter struct {
	Prefix string
	Router http.Handler
	Routes []MountedRoute
}

// RouteMeta represents a route with its metadata for extraction.
type MountedRoute struct {
	Method       string
	Path         string
	Handler      http.Handler
	Middlewares  []func(http.Handler) http.Handler
	Summary      string
	Description  string
	RequestType  any
	ResponseType any
	PathParams   []ParamInfo
}

// ParamInfo represents information about a path parameter for documentation purposes.
type ParamInfo struct {
	Name        string
	Type        string
	Required    bool
	Description string
}

// RouteInfo contains metadata for a route that can be used by adapters.
type RouteInfo struct {
	Summary      string
	Description  string
	RequestType  any
	ResponseType any
	PathParams   []ParamInfo
}

// RouteMeta represents a route with its metadata for extraction.
type RouteMeta struct {
	Method string
	Path   string
	Info   RouteInfo
}
