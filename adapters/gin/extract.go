package restgin

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/reststore/restkit/internal/schema"
)

// Extract extracts routes from a Gin router using provided metadata.
func Extract(
	router *gin.Engine,
	metas []schema.RouteMeta,
) ([]schema.MountedRoute, error) {
	metaMap := make(map[string]schema.RouteInfo)

	for _, m := range metas {
		key := routeKey(m.Method, m.Path)
		metaMap[key] = m.Info
	}

	var routes []schema.MountedRoute

	ginRoutes := router.Routes()
	for _, route := range ginRoutes {
		key := routeKey(route.Method, route.Path)
		info, found := metaMap[key]

		if !found {
			continue
		}

		routes = append(routes, schema.MountedRoute{
			Method:       route.Method,
			Path:         route.Path,
			Handler:      nil, // Handler not directly accessible from Gin routes
			Summary:      info.Summary,
			Description:  info.Description,
			RequestType:  info.RequestType,
			ResponseType: info.ResponseType,
			PathParams: extractParams(
				route.Path,
				info.PathParams,
			),
		})
	}

	return routes, nil
}

// ExtractAll extracts all routes from a Gin router without requiring metadata.
func ExtractAll(router *gin.Engine) ([]schema.MountedRoute, error) {
	var routes []schema.MountedRoute

	ginRoutes := router.Routes()
	for _, route := range ginRoutes {
		routes = append(routes, schema.MountedRoute{
			Method:     route.Method,
			Path:       route.Path,
			Handler:    nil,
			PathParams: extractPathParams(route.Path),
		})
	}

	return routes, nil
}

// routeKey creates a unique key for a route using method and path.
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

// extractPathParams extracts path parameters from a Gin path pattern.
// For example, "/users/:id" returns a ParamInfo for "id".
func extractPathParams(pattern string) []schema.ParamInfo {
	var params []schema.ParamInfo
	parts := strings.Split(pattern, "/")

	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			// Handle named parameter like :id
			paramName := strings.TrimPrefix(part, ":")
			// Handle wildcard parameters (e.g., :id*)
			paramName = strings.TrimSuffix(paramName, "*")
			params = append(params, schema.ParamInfo{
				Name:     paramName,
				Type:     "string",
				Required: true,
			})
		}
	}

	return params
}

// extractHandler attempts to extract the http.Handler from a gin.HandlerFunc.
// Note: This is a best-effort function since Gin doesn't expose handlers directly.
func extractHandler(handler gin.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a Gin context from the request
		// This is a simplified version - in practice, you'd need proper Gin context
		fmt.Println("Warning: Handler extraction from Gin is limited")
	})
}
