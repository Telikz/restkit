package cache

import (
	"net/http"
	"strings"
	"sync"
)

type RouteCache struct {
	exactMatches sync.Map
	// patternMatches stores routes with path parameters like /users/{id}
	patternMatches []patternRoute
	mu             sync.RWMutex
}

type patternRoute struct {
	Method  string
	Path    string // The pattern, e.g., /users/{id}
	Handler http.Handler
	parts   []string // Pre-split path parts for efficient matching
}

type CachedRoute struct {
	Method  string
	Path    string
	Handler http.Handler
}

func NewRouteCache() *RouteCache {
	return &RouteCache{
		patternMatches: make([]patternRoute, 0),
	}
}

func (rc *RouteCache) Get(method, path string) (*CachedRoute, bool) {
	// First try exact match
	key := method + " " + path
	if val, ok := rc.exactMatches.Load(key); ok {
		return val.(*CachedRoute), true
	}

	// Then try pattern matching for parameterized routes
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	requestParts := strings.Split(path, "/")
	for _, pr := range rc.patternMatches {
		if pr.Method != method {
			continue
		}
		if matchPathPattern(requestParts, pr.parts) {
			return &CachedRoute{
				Method:  pr.Method,
				Path:    pr.Path,
				Handler: pr.Handler,
			}, true
		}
	}

	return nil, false
}

func (rc *RouteCache) Set(method, path string, handler http.Handler) *CachedRoute {
	route := &CachedRoute{
		Method:  method,
		Path:    path,
		Handler: handler,
	}

	// Check if path has parameters like {id}
	if strings.Contains(path, "{") && strings.Contains(path, "}") {
		rc.mu.Lock()
		rc.patternMatches = append(rc.patternMatches, patternRoute{
			Method:  method,
			Path:    path,
			Handler: handler,
			parts:   strings.Split(path, "/"),
		})
		rc.mu.Unlock()
	} else {
		key := method + " " + path
		rc.exactMatches.Store(key, route)
	}

	return route
}

// matchPathPattern matches request path parts against a pattern
// Pattern parts with {param} match any non-empty value in that position
func matchPathPattern(requestParts, patternParts []string) bool {
	if len(requestParts) != len(patternParts) {
		return false
	}

	for i, patternPart := range patternParts {
		// If it's a parameter placeholder like {id}, it matches any non-empty value
		if strings.HasPrefix(patternPart, "{") && strings.HasSuffix(patternPart, "}") {
			if requestParts[i] == "" {
				return false
			}
			continue
		}
		// Otherwise, must match exactly
		if patternPart != requestParts[i] {
			return false
		}
	}

	return true
}
