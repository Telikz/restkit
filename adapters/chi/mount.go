package restchi

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/telikz/restkit"
	"github.com/telikz/restkit/internal/schema"
)

type Info struct {
	Summary      string
	Description  string
	RequestType  any
	ResponseType any
	PathParams   []restkit.ParamInfo
}

func Meta(method, path string, info Info) routeMeta {
	return routeMeta{
		method: method,
		path:   path,
		info:   info,
	}
}

func Metas(metas ...routeMeta) []routeMeta {
	return metas
}

type routeMeta struct {
	method string
	path   string
	info   Info
}

func Extract(router chi.Router, metas []routeMeta) ([]schema.MountedRoute, error) {
	metaMap := make(map[string]Info)
	for _, m := range metas {
		key := routeKey(m.method, m.path)
		metaMap[key] = m.info
	}

	var routes []schema.MountedRoute

	err := chi.Walk(router, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		key := routeKey(method, route)
		info, found := metaMap[key]

		if !found {
			return nil
		}

		routes = append(routes, schema.MountedRoute{
			Method:       method,
			Path:         route,
			Handler:      handler,
			Summary:      info.Summary,
			Description:  info.Description,
			RequestType:  info.RequestType,
			ResponseType: info.ResponseType,
			PathParams:   extractParams(route, info.PathParams),
		})
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walking chi router: %w", err)
	}

	return routes, nil
}

func routeKey(method, path string) string {
	return method + " " + path
}

func extractParams(path string, provided []restkit.ParamInfo) []restkit.ParamInfo {
	if len(provided) > 0 {
		return provided
	}
	return extractPathParams(path)
}

func extractPathParams(pattern string) []restkit.ParamInfo {
	var params []restkit.ParamInfo
	start := -1
	for i, c := range pattern {
		if c == '{' {
			start = i + 1
		} else if c == '}' && start != -1 {
			params = append(params, restkit.ParamInfo{
				Name:     pattern[start:i],
				Type:     "string",
				Required: true,
			})
			start = -1
		}
	}
	return params
}
