package restchi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/reststore/restkit/internal/api"
	routectx "github.com/reststore/restkit/internal/context"
	"github.com/reststore/restkit/internal/schema"
)

// Mount extracts routes from a Chi router and mounts them into a RestKit API.
// Wrapped routes support request validation (via RouteMeta.RequestType) and
// response serialization (same-format only, e.g., JSON→JSON).
//
// Limitation: Cross-format serialization (e.g., JSON→XML) doesn't work for mounted
// Chi routes because the original struct type is lost when JSON is written.
// Use RestKit endpoints directly for cross-format serialization.
func Mount(
	restkitApi *api.Api,
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

	wrappedRouter := wrapWithSerializerAndValidation(router, routes)

	restkitApi.MountRouter(prefix, wrappedRouter, routes)

	return nil
}

// serializingResponseWriter wraps the ResponseWriter to apply API serializers
type serializingResponseWriter struct {
	http.ResponseWriter
	ctx context.Context
}

func (w *serializingResponseWriter) Write(data []byte) (int, error) {
	// Attempt to apply custom serializer from context.
	// Note: This works for same-format conversion (e.g., JSON->JSON with custom formatting).
	// Cross-format (JSON->XML) may fail because unmarshalling JSON to 'any' creates
	// map[string]interface{} which XML serializers cannot handle. In that case,
	// we fall back to writing the original data.
	if v := w.ctx.Value(routectx.SerializerCtxKey); v != nil {
		if serializer, ok := v.(func(http.ResponseWriter, any) error); ok {
			// Try to parse as JSON and re-serialize
			var obj any
			if err := json.Unmarshal(data, &obj); err == nil {
				// Successfully parsed as JSON, use custom serializer
				// Note: This may fail for XML if the data unmarshals to map[string]interface{}
				// In that case, fall back to writing the original JSON
				if err := serializer(w.ResponseWriter, obj); err == nil {
					return len(data), nil
				}
				// Serializer failed (e.g., XML doesn't support maps), fall through to write original
			}
		}
	}
	// No serializer or not valid JSON, write as-is
	return w.ResponseWriter.Write(data)
}

func (w *serializingResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *serializingResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

func wrapWithSerializerAndValidation(router chi.Router, routes []schema.MountedRoute) http.Handler {
	hasValidation := false
	for _, route := range routes {
		if route.RequestType != nil {
			hasValidation = true
			break
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wrap the response writer to support serialization
		sw := &serializingResponseWriter{
			ResponseWriter: w,
			ctx:            r.Context(),
		}

		if !hasValidation {
			router.ServeHTTP(sw, r)
			return
		}

		for _, route := range routes {
			if matchesRoute(r, route) && route.RequestType != nil {
				validationMiddleware(
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						// Wrap again for validation middleware path
						sw2 := &serializingResponseWriter{
							ResponseWriter: w,
							ctx:            r.Context(),
						}
						router.ServeHTTP(sw2, r)
					}),
					route.RequestType,
				).ServeHTTP(sw, r)
				return
			}
		}

		router.ServeHTTP(sw, r)
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

	for i := range routeParts {
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
