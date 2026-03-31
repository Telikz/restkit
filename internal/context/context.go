package context

import (
	"context"
	"regexp"
	"sync"
)

var RouteCtxKey = &contextKey{"RouteContext"}

var pathParamRegex = regexp.MustCompile(`\{([^}]+)}`)

// routeContextPool provides a pool for RouteContext to reduce allocations
type routeContextPool struct {
	pool sync.Pool
}

var rcPool = &routeContextPool{
	pool: sync.Pool{
		New: func() any {
			return &RouteContext{
				params: make(map[string]string),
			}
		},
	},
}

// Get acquires a RouteContext from the pool
func (p *routeContextPool) Get() *RouteContext {
	rc := p.pool.Get().(*RouteContext)
	for k := range rc.params {
		delete(rc.params, k)
	}
	return rc
}

// Put returns a RouteContext to the pool
func (p *routeContextPool) Put(rc *RouteContext) {
	if rc != nil {
		p.pool.Put(rc)
	}
}

type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "api context value " + k.name
}

type RouteContext struct {
	params map[string]string
}

// NewRouteContext creates a new RouteContext using the pool for efficiency
func NewRouteContext() *RouteContext {
	return rcPool.Get()
}

// URLParam retrieves a URL parameter by name
func (rc *RouteContext) URLParam(key string) string {
	if rc == nil || rc.params == nil {
		return ""
	}
	return rc.params[key]
}

// SetURLParam sets a URL parameter
func (rc *RouteContext) SetURLParam(key, value string) {
	if rc.params == nil {
		rc.params = make(map[string]string)
	}
	rc.params[key] = value
}

// URLParam extracts a URL parameter from the request context
func URLParam(ctx context.Context, key string) string {
	if rc := RouteCtxFromContext(ctx); rc != nil {
		return rc.URLParam(key)
	}
	return ""
}

// RouteCtxFromContext extracts the route context from a request context
func RouteCtxFromContext(ctx context.Context) *RouteContext {
	val, _ := ctx.Value(RouteCtxKey).(*RouteContext)
	return val
}

// ExtractPathParams extracts parameters from a URL path using a
// Pattern should be like "/users/{id}/posts/{postId}"
func ExtractPathParams(pattern, path string) map[string]string {
	params := make(map[string]string)

	matches := pathParamRegex.FindAllStringSubmatch(pattern, -1)
	paramNames := make([]string, len(matches))
	for i, match := range matches {
		paramNames[i] = match[1]
	}

	regexPattern := pathParamRegex.ReplaceAllString(pattern, `([^/]+)`)
	regexPattern = "^" + regexPattern + "$"

	pathRegex, err := regexp.Compile(regexPattern)
	if err != nil {
		return params
	}

	values := pathRegex.FindStringSubmatch(path)
	if values == nil {
		return params
	}

	for i, name := range paramNames {
		if i+1 < len(values) {
			params[name] = values[i+1]
		}
	}

	return params
}
