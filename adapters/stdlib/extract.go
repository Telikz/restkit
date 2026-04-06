package reststdlib

import (
	"errors"
	"net/http"
	"strings"

	"github.com/reststore/restkit/internal/schema"
)

func Extract(mux *http.ServeMux, metas []schema.RouteMeta) ([]schema.MountedRoute, error) {
	metaMap := make(map[string]schema.RouteInfo)
	for _, m := range metas {
		key := m.Method + " " + m.Path
		metaMap[key] = m.Info
	}

	var routes []schema.MountedRoute

	// Note: http.ServeMux doesn't expose registered routes
	// We rely on metadata provided by the user
	for _, m := range metas {
		info := metaMap[m.Method+" "+m.Path]
		routes = append(routes, schema.MountedRoute{
			Method:       m.Method,
			Path:         m.Path,
			Summary:      info.Summary,
			Description:  info.Description,
			RequestType:  info.RequestType,
			ResponseType: info.ResponseType,
			PathParams:   extractPathParams(m.Path),
		})
	}

	return routes, nil
}

func ExtractAll(mux *http.ServeMux) ([]schema.MountedRoute, error) {
	// http.ServeMux doesn't expose its routes - requires manual metadata
	return nil, errors.New("http.ServeMux does not expose registered routes; please provide route metadata via the metas parameter")
}

func extractPathParams(pattern string) []schema.ParamInfo {
	var params []schema.ParamInfo
	parts := strings.SplitSeq(pattern, "/")

	for part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			paramName := strings.TrimPrefix(part, "{")
			paramName = strings.TrimSuffix(paramName, "}")
			paramName = strings.TrimSuffix(paramName, "...")
			params = append(params, schema.ParamInfo{
				Name:     paramName,
				Type:     "string",
				Required: true,
			})
		}
	}

	return params
}
