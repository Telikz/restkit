package restchi

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/telikz/restkit/internal/schema"
)

// Extract extracts routes from a Chi router using provided metadata.
func Extract(
	router chi.Router,
	metas []schema.RouteMeta,
) ([]schema.MountedRoute, error) {
	metaMap := make(map[string]schema.RouteInfo)

	for _, m := range metas {
		key := routeKey(m.Method, m.Path)
		metaMap[key] = m.Info
	}

	var routes []schema.MountedRoute

	err := chi.Walk(
		router,
		func(method string, route string, handler http.Handler,
			middlewares ...func(http.Handler) http.Handler) error {
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
				PathParams: extractParams(
					route,
					info.PathParams,
				),
			})
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"walking chi router: %w",
			err,
		)
	}

	return routes, nil
}

// ExtractAll extracts all routes from a Chi router without requiring metadata.
func ExtractAll(router chi.Router) ([]schema.MountedRoute, error) {
	var routes []schema.MountedRoute

	err := chi.Walk(
		router,
		func(method string, route string, handler http.Handler,
			middlewares ...func(http.Handler) http.Handler,
		) error {
			routes = append(routes, schema.MountedRoute{
				Method:     method,
				Path:       route,
				Handler:    handler,
				PathParams: extractPathParams(route),
			})
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"walking chi router: %w",
			err,
		)
	}

	return routes, nil
}

// routeKey creates a unique key for a route using method
// and path.
func routeKey(method, path string) string {
	return method + " " + path
}

// extractParams returns provided params if available,
// otherwise extracts from path pattern.
func extractParams(
	path string,
	provided []schema.ParamInfo,
) []schema.ParamInfo {
	if len(provided) > 0 {
		return provided
	}
	return extractPathParams(path)
}

// extractPathParams extracts path parameters from a Chi path pattern.
//
//	For example, "/users/{id}" returns a ParamInfo for "id".
func extractPathParams(pattern string) []schema.ParamInfo {
	var params []schema.ParamInfo
	start := -1
	for i, c := range pattern {
		if c == '{' {
			start = i + 1
		} else if c == '}' && start != -1 {
			params = append(params, schema.ParamInfo{
				Name:     pattern[start:i],
				Type:     "string",
				Required: true,
			})
			start = -1
		}
	}
	return params
}
