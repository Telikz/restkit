package cache

import (
	"net/http"
	"sync"
)

type RouteCache struct {
	exactMatches sync.Map
}

type CachedRoute struct {
	Method  string
	Path    string
	Handler http.Handler
}

func NewRouteCache() *RouteCache {
	return &RouteCache{}
}

func (rc *RouteCache) Get(method, path string) (*CachedRoute, bool) {
	key := method + " " + path
	if val, ok := rc.exactMatches.Load(key); ok {
		return val.(*CachedRoute), true
	}
	return nil, false
}

func (rc *RouteCache) Set(method, path string, handler http.Handler) *CachedRoute {
	route := &CachedRoute{
		Method:  method,
		Path:    path,
		Handler: handler,
	}
	key := method + " " + path
	rc.exactMatches.Store(key, route)
	return route
}
