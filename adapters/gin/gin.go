package restgin

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/reststore/restkit/internal/api"
	routectx "github.com/reststore/restkit/internal/context"
	"github.com/reststore/restkit/internal/docs"
)

func RegisterRoutes(r *gin.Engine, apiInstance *api.Api) {
	registered := make(map[string]bool)

	configMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			if apiInstance.Validator != nil {
				ctx = context.WithValue(ctx, routectx.ValidatorCtxKey, apiInstance.Validator)
			}
			if apiInstance.Serializer != nil {
				ctx = context.WithValue(ctx, routectx.SerializerCtxKey, apiInstance.Serializer)
			}
			if apiInstance.Deserializer != nil {
				ctx = context.WithValue(ctx, routectx.DeserializerCtxKey, apiInstance.Deserializer)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	for _, group := range apiInstance.Groups {
		for _, endpoint := range group.GetEndpoints() {
			key := fmt.Sprintf("%s %s", endpoint.GetMethod(), endpoint.GetPath())
			registered[key] = true
			handler := endpoint.GetHandler()
			handler = configMiddleware(handler)
			r.Handle(
				endpoint.GetMethod(),
				endpoint.GetPath(),
				adaptHandler(handler),
			)
		}
	}

	for _, endpoint := range apiInstance.Endpoints {
		key := fmt.Sprintf("%s %s", endpoint.GetMethod(), endpoint.GetPath())
		if !registered[key] {
			handler := endpoint.GetHandler()
			handler = configMiddleware(handler)
			for i := len(apiInstance.Middleware) - 1; i >= 0; i-- {
				handler = apiInstance.Middleware[i](handler)
			}
			r.Handle(
				endpoint.GetMethod(),
				endpoint.GetPath(),
				adaptHandler(handler),
			)
		}
	}

	if apiInstance.SwaggerUIEnabled {
		r.GET(
			apiInstance.SwaggerUIPath,
			func(c *gin.Context) {
				docs.ServeSwaggerUI(c.Writer, apiInstance.SwaggerUIPath)
			},
		)
		r.GET(
			apiInstance.SwaggerUIPath+"/openapi.json",
			func(c *gin.Context) {
				apiInstance.ServeOpenAPI(c.Writer, c.Request)
			},
		)
	}
}

func adaptHandler(h http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
