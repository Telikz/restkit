package context

import (
	"context"
	"regexp"
)

var RouteCtxKey = &contextKey{"RouteContext"}

type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "api context value " + k.name
}

type RouteContext struct {
	params map[string]string
}

func NewRouteContext() *RouteContext {
	return &RouteContext{}
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

// ExtractPathParams extracts parameters from a URL path using a pattern
// Pattern should be like "/users/{id}/posts/{postId}"
func ExtractPathParams(pattern, path string) map[string]string {
	params := make(map[string]string)

	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(pattern, -1)
	paramNames := make([]string, len(matches))
	for i, match := range matches {
		paramNames[i] = match[1]
	}

	regexPattern := re.ReplaceAllString(pattern, `([^/]+)`)
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
