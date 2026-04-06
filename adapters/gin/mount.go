package restgin

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/reststore/restkit/internal/api"
	"github.com/reststore/restkit/internal/cache"
	"github.com/reststore/restkit/internal/schema"
)

var mountCache = cache.NewRouteCache()

func Mount(
	restkitApi *api.Api,
	prefix string,
	router *gin.Engine,
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
		return errors.New("extracting routes from gin router: " + err.Error())
	}

	for _, route := range routes {
		var handler http.Handler
		if route.RequestType != nil {
			handler = validationMiddleware(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					router.ServeHTTP(w, r)
				}),
				route.RequestType,
			)
		} else {
			handler = router
		}
		mountCache.Set(route.Method, prefix+route.Path, handler)
	}

	cachedHandler := &cachedMountHandler{
		cache:  mountCache,
		router: router,
		prefix: prefix,
	}

	restkitApi.MountRouter(prefix, cachedHandler, routes)

	return nil
}

type cachedMountHandler struct {
	cache  *cache.RouteCache
	router *gin.Engine
	prefix string
}

func (h *cachedMountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if route, ok := h.cache.Get(r.Method, r.URL.Path); ok {
		route.Handler.ServeHTTP(w, r)
		return
	}

	h.router.ServeHTTP(w, r)
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

	for i := range routeParts {
		if strings.HasPrefix(routeParts[i], ":") {
			continue
		}
		if routeParts[i] != requestParts[i] {
			return false
		}
	}

	return true
}
