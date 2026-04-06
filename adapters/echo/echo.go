package restecho

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/reststore/restkit/internal/api"
	"github.com/reststore/restkit/internal/docs"
)

func RegisterRoutes(e *echo.Echo, apiInstance *api.Api) {
	registered := make(map[string]bool)

	for _, group := range apiInstance.Groups {
		for _, endpoint := range group.GetEndpoints() {
			key := fmt.Sprintf("%s %s", endpoint.GetMethod(), endpoint.GetPath())
			registered[key] = true
			handler := endpoint.GetHandler()
			e.Add(
				endpoint.GetMethod(),
				endpoint.GetPath(),
				echo.WrapHandler(handler),
			)
		}
	}

	for _, endpoint := range apiInstance.Endpoints {
		key := fmt.Sprintf("%s %s", endpoint.GetMethod(), endpoint.GetPath())
		if !registered[key] {
			handler := endpoint.GetHandler()
			for i := len(apiInstance.Middleware) - 1; i >= 0; i-- {
				handler = apiInstance.Middleware[i](handler)
			}
			e.Add(
				endpoint.GetMethod(),
				endpoint.GetPath(),
				echo.WrapHandler(handler),
			)
		}
	}

	if apiInstance.SwaggerUIEnabled {
		e.GET(apiInstance.SwaggerUIPath, func(c echo.Context) error {
			docs.ServeSwaggerUI(c.Response().Writer, apiInstance.SwaggerUIPath)
			return nil
		})
		e.GET(apiInstance.SwaggerUIPath+"/openapi.json", func(c echo.Context) error {
			apiInstance.ServeOpenAPI(c.Response().Writer, c.Request())
			return nil
		})
	}
}
