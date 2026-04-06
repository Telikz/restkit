package reststdlib

import (
	"errors"
	"net/http"
	"strings"

	"github.com/reststore/restkit/internal/api"
	"github.com/reststore/restkit/internal/cache"
	"github.com/reststore/restkit/internal/schema"
)

func Mount(restkitApi *api.Api, prefix string, mux *http.ServeMux, metas []schema.RouteMeta) error {
	var routes []schema.MountedRoute
	var err error

	if len(metas) > 0 {
		routes, err = Extract(mux, metas)
	} else {
		routes, err = ExtractAll(mux)
	}

	if err != nil {
		return errors.New("extracting routes from stdlib mux: " + err.Error())
	}

	// Create per-instance cache instead of using global
	mountCache := cache.NewRouteCache()

	for _, route := range routes {
		var handler http.Handler
		if route.RequestType != nil {
			handler = validationMiddleware(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					mux.ServeHTTP(w, r)
				}),
				route.RequestType,
			)
		} else {
			handler = mux
		}
		mountCache.Set(route.Method, route.Path, handler)
	}

	cachedHandler := &cachedMountHandler{
		cache:  mountCache,
		mux:    mux,
		prefix: prefix,
	}

	restkitApi.MountRouter(prefix, cachedHandler, routes)

	return nil
}

type cachedMountHandler struct {
	cache  *cache.RouteCache
	mux    *http.ServeMux
	prefix string
}

func (h *cachedMountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Strip the prefix from the request path for cache lookup
	path := r.URL.Path
	if h.prefix != "" && strings.HasPrefix(path, h.prefix) {
		path = path[len(h.prefix):]
		if path == "" {
			path = "/"
		}
	}

	if route, ok := h.cache.Get(r.Method, path); ok {
		route.Handler.ServeHTTP(w, r)
		return
	}

	h.mux.ServeHTTP(w, r)
}
