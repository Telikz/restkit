package schema

import (
	"net/http"
)

type MountedRouter struct {
	Prefix string
	Router http.Handler
	Routes []MountedRoute
}

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

type ParamInfo struct {
	Name        string
	Type        string
	Required    bool
	Description string
}
