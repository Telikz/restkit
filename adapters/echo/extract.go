package restecho

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/reststore/restkit/internal/schema"
)

func Extract(router *echo.Echo, metas []schema.RouteMeta) ([]schema.MountedRoute, error) {
	metaMap := make(map[string]schema.RouteInfo)
	for _, m := range metas {
		key := routeKey(m.Method, m.Path)
		metaMap[key] = m.Info
	}

	var routes []schema.MountedRoute

	for _, route := range router.Routes() {
		key := routeKey(route.Method, route.Path)
		info, found := metaMap[key]
		if !found {
			continue
		}

		routes = append(routes, schema.MountedRoute{
			Method:       route.Method,
			Path:         route.Path,
			Summary:      info.Summary,
			Description:  info.Description,
			RequestType:  info.RequestType,
			ResponseType: info.ResponseType,
			PathParams:   extractParams(route.Path, info.PathParams),
		})
	}

	return routes, nil
}

func ExtractAll(router *echo.Echo) ([]schema.MountedRoute, error) {
	var routes []schema.MountedRoute

	for _, route := range router.Routes() {
		routes = append(routes, schema.MountedRoute{
			Method:     route.Method,
			Path:       route.Path,
			PathParams: extractPathParams(route.Path),
		})
	}

	return routes, nil
}

func routeKey(method, path string) string {
	return method + " " + path
}

func extractParams(path string, provided []schema.ParamInfo) []schema.ParamInfo {
	if len(provided) > 0 {
		return provided
	}
	return extractPathParams(path)
}

func extractPathParams(pattern string) []schema.ParamInfo {
	var params []schema.ParamInfo
	parts := strings.Split(pattern, "/")

	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			paramName := strings.TrimPrefix(part, ":")
			params = append(params, schema.ParamInfo{
				Name:     paramName,
				Type:     "string",
				Required: true,
			})
		}
	}

	return params
}
