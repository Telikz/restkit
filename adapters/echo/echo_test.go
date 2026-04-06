package restecho

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/reststore/restkit/internal/api"
	ep "github.com/reststore/restkit/internal/endpoints"
	"github.com/reststore/restkit/internal/schema"
)

func TestRegisterRoutes(t *testing.T) {
	t.Run("group endpoints with prefix", func(t *testing.T) {
		e := echo.New()
		apiInstance := api.New()

		endpoint := ep.NewEndpointRes[string]().
			WithMethod("GET").
			WithPath("/users").
			WithHandler(func(ctx context.Context, _ ep.NoRequest) (string, error) {
				return "users", nil
			})

		group := ep.NewGroup("/api/v1").WithEndpoints(endpoint)
		apiInstance.AddGroup(group)
		RegisterRoutes(e, apiInstance)

		req := httptest.NewRequest("GET", "/api/v1/users", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("individual endpoints", func(t *testing.T) {
		e := echo.New()
		apiInstance := api.New()

		endpoint := ep.NewEndpointRes[string]().
			WithMethod("POST").
			WithPath("/login").
			WithHandler(func(ctx context.Context, _ ep.NoRequest) (string, error) {
				return "logged in", nil
			})

		apiInstance.AddEndpoint(endpoint)
		RegisterRoutes(e, apiInstance)

		req := httptest.NewRequest("POST", "/login", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("swagger UI endpoints", func(t *testing.T) {
		e := echo.New()
		apiInstance := api.New().
			WithTitle("Test API").
			WithVersion("1.0.0").
			WithSwaggerUI("/docs")

		RegisterRoutes(e, apiInstance)

		req := httptest.NewRequest("GET", "/docs", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200 for Swagger UI, got %d", rec.Code)
		}
	})
}

func TestExtract(t *testing.T) {
	t.Run("extract with metadata", func(t *testing.T) {
		e := echo.New()
		e.GET("/users", func(c echo.Context) error { return c.String(200, "users") })

		metas := []schema.RouteMeta{
			{
				Method: "GET",
				Path:   "/users",
				Info:   schema.RouteInfo{Summary: "List users"},
			},
		}

		routes, err := Extract(e, metas)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(routes) != 1 {
			t.Errorf("expected 1 route, got %d", len(routes))
		}
		if routes[0].Method != "GET" {
			t.Errorf("expected method GET, got %s", routes[0].Method)
		}
	})

	t.Run("extract with path params", func(t *testing.T) {
		e := echo.New()
		e.GET("/users/:id", func(c echo.Context) error { return c.String(200, "user") })

		metas := []schema.RouteMeta{
			{
				Method: "GET",
				Path:   "/users/:id",
				Info: schema.RouteInfo{
					Summary: "Get user",
					PathParams: []schema.ParamInfo{
						{Name: "id", Type: "integer", Required: true},
					},
				},
			},
		}

		routes, err := Extract(e, metas)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(routes) != 1 {
			t.Fatalf("expected 1 route, got %d", len(routes))
		}
		if len(routes[0].PathParams) != 1 {
			t.Errorf("expected 1 path param, got %d", len(routes[0].PathParams))
		}
	})

	t.Run("skip routes not in metadata", func(t *testing.T) {
		e := echo.New()
		e.GET("/users", func(c echo.Context) error { return nil })
		e.GET("/posts", func(c echo.Context) error { return nil })

		metas := []schema.RouteMeta{
			{Method: "GET", Path: "/users", Info: schema.RouteInfo{Summary: "List users"}},
		}

		routes, err := Extract(e, metas)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(routes) != 1 {
			t.Errorf("expected 1 route, got %d", len(routes))
		}
	})

	t.Run("extract all routes", func(t *testing.T) {
		e := echo.New()
		e.GET("/users", func(c echo.Context) error { return nil })
		e.POST("/users", func(c echo.Context) error { return nil })
		e.GET("/users/:id", func(c echo.Context) error { return nil })

		routes, err := ExtractAll(e)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(routes) < 3 {
			t.Errorf("expected at least 3 routes, got %d", len(routes))
		}
	})

	t.Run("auto path params from pattern", func(t *testing.T) {
		e := echo.New()
		e.GET("/users/:id/posts/:postId", func(c echo.Context) error { return nil })

		routes, err := ExtractAll(e)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		found := false
		for _, r := range routes {
			if r.Method == "GET" && r.Path == "/users/:id/posts/:postId" {
				found = true
				if len(r.PathParams) != 2 {
					t.Errorf("expected 2 path params, got %d", len(r.PathParams))
				}

				paramNames := make(map[string]bool)
				for _, p := range r.PathParams {
					paramNames[p.Name] = true
				}
				if !paramNames["id"] || !paramNames["postId"] {
					t.Error("expected both 'id' and 'postId' params")
				}
				break
			}
		}
		if !found {
			t.Error("expected GET /users/:id/posts/:postId route")
		}
	})
}

func TestMount(t *testing.T) {
	t.Run("mount router", func(t *testing.T) {
		e := echo.New()
		e.GET("/test", func(c echo.Context) error { return c.String(200, "test") })

		apiInstance := api.New()
		err := Mount(apiInstance, "/api", e, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(apiInstance.MountedRouters) != 1 {
			t.Errorf("expected 1 mounted router, got %d", len(apiInstance.MountedRouters))
		}
	})

	t.Run("mount with metadata", func(t *testing.T) {
		e := echo.New()
		e.GET("/external", func(c echo.Context) error { return c.String(200, "external") })

		metas := []schema.RouteMeta{
			{
				Method: "GET",
				Path:   "/external",
				Info:   schema.RouteInfo{Summary: "External route"},
			},
		}

		apiInstance := api.New()
		err := Mount(apiInstance, "/api", e, metas)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(apiInstance.MountedRouters[0].Routes) != 1 {
			t.Errorf("expected 1 route, got %d", len(apiInstance.MountedRouters[0].Routes))
		}
	})
}

func TestRouteMatching(t *testing.T) {
	tests := []struct {
		name        string
		request     *http.Request
		route       schema.MountedRoute
		shouldMatch bool
	}{
		{
			name:        "exact match",
			request:     httptest.NewRequest("GET", "/users", nil),
			route:       schema.MountedRoute{Method: "GET", Path: "/users"},
			shouldMatch: true,
		},
		{
			name:        "method mismatch",
			request:     httptest.NewRequest("POST", "/users", nil),
			route:       schema.MountedRoute{Method: "GET", Path: "/users"},
			shouldMatch: false,
		},
		{
			name:        "path mismatch",
			request:     httptest.NewRequest("GET", "/posts", nil),
			route:       schema.MountedRoute{Method: "GET", Path: "/users"},
			shouldMatch: false,
		},
		{
			name:        "path param match",
			request:     httptest.NewRequest("GET", "/users/123", nil),
			route:       schema.MountedRoute{Method: "GET", Path: "/users/:id"},
			shouldMatch: true,
		},
		{
			name:        "wrong segment count",
			request:     httptest.NewRequest("GET", "/users/123/posts", nil),
			route:       schema.MountedRoute{Method: "GET", Path: "/users/:id"},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesRoute(tt.request, tt.route)
			if result != tt.shouldMatch {
				t.Errorf("matchesRoute() = %v, want %v", result, tt.shouldMatch)
			}
		})
	}
}

func TestMatchPath(t *testing.T) {
	tests := []struct {
		requestPath string
		routePath   string
		want        bool
	}{
		{"/users", "/users", true},
		{"/users/123", "/users/:id", true},
		{"/users/123/posts/456", "/users/:id/posts/:postId", true},
		{"/users", "/posts", false},
		{"/users/123", "/users", false},
		{"/users", "/users/123", false},
	}

	for _, tt := range tests {
		t.Run(tt.requestPath+"_"+tt.routePath, func(t *testing.T) {
			got := matchPath(tt.requestPath, tt.routePath)
			if got != tt.want {
				t.Errorf(
					"matchPath(%q, %q) = %v, want %v",
					tt.requestPath,
					tt.routePath,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestValidationMiddleware(t *testing.T) {
	type TestRequest struct {
		Name  string `json:"name"  validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	t.Run("skips validation for GET", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := validationMiddleware(next, TestRequest{})
		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("skips validation for DELETE", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := validationMiddleware(next, TestRequest{})
		req := httptest.NewRequest("DELETE", "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("validates valid request", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			w.Write(body)
		})

		handler := validationMiddleware(next, TestRequest{})
		body := `{"name":"John","email":"john@example.com"}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler should not be called")
		})

		handler := validationMiddleware(next, TestRequest{})
		body := `invalid json`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("rejects empty body", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler should not be called")
		})

		handler := validationMiddleware(next, TestRequest{})
		req := httptest.NewRequest("POST", "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})
}

func TestExtractParams(t *testing.T) {
	t.Run("uses provided params", func(t *testing.T) {
		provided := []schema.ParamInfo{
			{Name: "id", Type: "integer", Required: true},
		}

		result := extractParams("/users/:id", provided)

		if len(result) != 1 {
			t.Errorf("expected 1 param, got %d", len(result))
		}
		if result[0].Type != "integer" {
			t.Errorf("expected type 'integer', got '%s'", result[0].Type)
		}
	})

	t.Run("extracts from path when empty", func(t *testing.T) {
		result := extractParams("/users/:userId/posts/:postId", []schema.ParamInfo{})

		if len(result) != 2 {
			t.Errorf("expected 2 params, got %d", len(result))
		}

		paramNames := make(map[string]bool)
		for _, p := range result {
			paramNames[p.Name] = true
		}
		if !paramNames["userId"] || !paramNames["postId"] {
			t.Error("expected both userId and postId params")
		}
	})

	t.Run("extracts from path when nil", func(t *testing.T) {
		result := extractParams("/test/:param", nil)

		if len(result) != 1 {
			t.Errorf("expected 1 param, got %d", len(result))
		}
		if result[0].Name != "param" {
			t.Errorf("expected param name 'param', got '%s'", result[0].Name)
		}
	})
}

func TestRouteKey(t *testing.T) {
	tests := []struct {
		method   string
		path     string
		expected string
	}{
		{"GET", "/users", "GET /users"},
		{"POST", "/users", "POST /users"},
		{"GET", "/users/:id", "GET /users/:id"},
		{"DELETE", "/users/123", "DELETE /users/123"},
	}

	for _, tt := range tests {
		result := routeKey(tt.method, tt.path)
		if result != tt.expected {
			t.Errorf("routeKey(%q, %q) = %q, want %q", tt.method, tt.path, result, tt.expected)
		}
	}
}
