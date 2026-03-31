package restchi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/reststore/restkit/internal/api"
	"github.com/reststore/restkit/internal/schema"
)

func Mount(
	a *api.Api,
	prefix string,
	router chi.Router,
	metas []schema.RouteMeta,
) error {
	var routes []schema.MountedRoute
	var err error

	if len(metas) > 0 {
		routes, err = Extract(router, metas)
	} else {
		routes, err = ExtractAll(router)
	}

	if err != nil {
		return errors.New(
			"extracting routes from chi router: " + err.Error(),
		)
	}

	wrappedRouter := wrapWithValidation(router, routes)

	a.MountRouter(prefix, wrappedRouter, routes)

	return nil
}

func wrapWithValidation(router chi.Router, routes []schema.MountedRoute) http.Handler {
	hasValidation := false
	for _, route := range routes {
		if route.RequestType != nil {
			hasValidation = true
			break
		}
	}

	if !hasValidation {
		return router
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, route := range routes {
			if matchesRoute(r, route) && route.RequestType != nil {
				validationMiddleware(
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						router.ServeHTTP(w, r)
					}),
					route.RequestType,
				).ServeHTTP(w, r)
				return
			}
		}

		router.ServeHTTP(w, r)
	})
}

func matchesRoute(r *http.Request, route schema.MountedRoute) bool {
	if !strings.EqualFold(r.Method, route.Method) {
		return false
	}

	return matchPath(r.URL.Path, route.Path)
}

func matchPath(requestPath, routePath string) bool {
	requestParts := strings.Split(requestPath, "/")
	routeParts := strings.Split(routePath, "/")

	if len(requestParts) != len(routeParts) {
		return false
	}

	for i := 0; i < len(routeParts); i++ {
		routePart := routeParts[i]
		requestPart := requestParts[i]

		if strings.HasPrefix(routePart, "{") && strings.HasSuffix(routePart, "}") {
			continue
		}

		if routePart != requestPart {
			return false
		}
	}

	return true
}
