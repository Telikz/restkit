package endpoints

import "net/http"

// routeInfo wraps a Route and applies group-level prefix and middleware.
type routeInfo struct {
	route      Route
	prefix     string
	middleware []func(http.Handler) http.Handler
}

func (r *routeInfo) GetMethod() string {
	return r.route.GetMethod()
}

func (r *routeInfo) GetPath() string {
	if r.route.GetPath() == "/" {
		return r.prefix
	}
	return r.prefix + r.route.GetPath()
}

func (r *routeInfo) GetTitle() string {
	return r.route.GetTitle()
}

func (r *routeInfo) GetDescription() string {
	return r.route.GetDescription()
}

func (r *routeInfo) GetMiddleware() []func(http.Handler) http.Handler {
	combined := make([]func(http.Handler) http.Handler, 0,
		len(r.route.GetMiddleware())+len(r.middleware))
	combined = append(combined, r.route.GetMiddleware()...)
	combined = append(combined, r.middleware...)
	return combined
}

func (r *routeInfo) GetRequestSchema() map[string]any {
	return r.route.GetRequestSchema()
}

func (r *routeInfo) GetResponseSchema() map[string]any {
	return r.route.GetResponseSchema()
}

func (r *routeInfo) GetHandler() http.Handler {
	handler := r.route.GetHandler()
	for i := len(r.middleware) - 1; i >= 0; i-- {
		handler = r.middleware[i](handler)
	}
	return handler
}

func (r *routeInfo) setPath(path string) {}

func (r *routeInfo) addMiddleware(mw []func(http.Handler) http.Handler) {
	r.middleware = append(r.middleware, mw...)
}
