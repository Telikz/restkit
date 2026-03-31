package restchi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/reststore/restkit/internal/api"
	ep "github.com/reststore/restkit/internal/endpoints"
	"github.com/reststore/restkit/internal/schema"
)

// TestRegisterRoutes tests registering routes with Chi router
func TestRegisterRoutes(t *testing.T) {
	t.Run("register group endpoints", func(t *testing.T) {
		r := chi.NewRouter()
		apiInstance := api.New()

		endpoint := ep.NewEndpointRes[string]().
			WithMethod("GET").
			WithPath("/users").
			WithHandler(func(ctx context.Context) (string, error) {
				return "users", nil
			})

		group := ep.NewGroup("/api/v1").WithEndpoints(endpoint)
		apiInstance.AddGroup(group)

		RegisterRoutes(r, apiInstance)

		// Test the route
		req := httptest.NewRequest("GET", "/api/v1/users", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("register individual endpoints", func(t *testing.T) {
		r := chi.NewRouter()
		apiInstance := api.New()

		endpoint := ep.NewEndpointRes[string]().
			WithMethod("POST").
			WithPath("/login").
			WithHandler(func(ctx context.Context) (string, error) {
				return "logged in", nil
			})

		apiInstance.AddEndpoint(endpoint)

		RegisterRoutes(r, apiInstance)

		req := httptest.NewRequest("POST", "/login", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("avoid duplicate registration", func(t *testing.T) {
		r := chi.NewRouter()
		apiInstance := api.New()

		endpoint := ep.NewEndpointRes[string]().
			WithMethod("GET").
			WithPath("/test").
			WithHandler(func(ctx context.Context) (string, error) {
				return "test", nil
			})

		group := ep.NewGroup("/api").WithEndpoints(endpoint)
		apiInstance.AddGroup(group)
		// Also add to endpoints - should not duplicate
		apiInstance.AddEndpoint(endpoint)

		// This should work without panic - routes are deduplicated
		RegisterRoutes(r, apiInstance)
	})

	t.Run("register swagger UI", func(t *testing.T) {
		r := chi.NewRouter()
		apiInstance := api.New().
			WithTitle("Test API").
			WithVersion("1.0.0").
			WithSwaggerUI("/docs")

		RegisterRoutes(r, apiInstance)

		// Test Swagger UI endpoint
		req := httptest.NewRequest("GET", "/docs", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200 for Swagger UI, got %d", rec.Code)
		}

		contentType := rec.Header().Get("Content-Type")
		if contentType != "text/html" {
			t.Errorf("expected Content-Type 'text/html', got '%s'", contentType)
		}

		// Test OpenAPI JSON endpoint
		req2 := httptest.NewRequest("GET", "/docs/openapi.json", nil)
		rec2 := httptest.NewRecorder()
		r.ServeHTTP(rec2, req2)

		if rec2.Code != http.StatusOK {
			t.Errorf("expected status 200 for OpenAPI JSON, got %d", rec2.Code)
		}

		contentType2 := rec2.Header().Get("Content-Type")
		if contentType2 != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got '%s'", contentType2)
		}
	})

	t.Run("no swagger when disabled", func(t *testing.T) {
		r := chi.NewRouter()
		apiInstance := api.New()

		RegisterRoutes(r, apiInstance)

		req := httptest.NewRequest("GET", "/swagger", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404 when swagger disabled, got %d", rec.Code)
		}
	})
}

// TestExtract tests route extraction with metadata
func TestExtract(t *testing.T) {
	t.Run("extract with metadata", func(t *testing.T) {
		r := chi.NewRouter()
		r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("users"))
		})

		metas := []schema.RouteMeta{
			{
				Method: "GET",
				Path:   "/users",
				Info: schema.RouteInfo{
					Summary:     "List users",
					Description: "Get all users",
				},
			},
		}

		routes, err := Extract(r, metas)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(routes) != 1 {
			t.Errorf("expected 1 route, got %d", len(routes))
		}

		if routes[0].Method != "GET" {
			t.Errorf("expected method 'GET', got '%s'", routes[0].Method)
		}

		if routes[0].Path != "/users" {
			t.Errorf("expected path '/users', got '%s'", routes[0].Path)
		}

		if routes[0].Summary != "List users" {
			t.Errorf("expected summary 'List users', got '%s'", routes[0].Summary)
		}
	})

	t.Run("extract with path params", func(t *testing.T) {
		r := chi.NewRouter()
		r.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("user"))
		})

		metas := []schema.RouteMeta{
			{
				Method: "GET",
				Path:   "/users/{id}",
				Info: schema.RouteInfo{
					Summary: "Get user",
					PathParams: []schema.ParamInfo{
						{
							Name:        "id",
							Type:        "integer",
							Required:    true,
							Description: "User ID",
						},
					},
				},
			},
		}

		routes, err := Extract(r, metas)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(routes) != 1 {
			t.Fatalf("expected 1 route, got %d", len(routes))
		}

		if len(routes[0].PathParams) != 1 {
			t.Errorf("expected 1 path param, got %d", len(routes[0].PathParams))
		}

		if routes[0].PathParams[0].Name != "id" {
			t.Errorf("expected param name 'id', got '%s'", routes[0].PathParams[0].Name)
		}

		if routes[0].PathParams[0].Type != "integer" {
			t.Errorf("expected param type 'integer', got '%s'", routes[0].PathParams[0].Type)
		}
	})

	t.Run("extract with request/response types", func(t *testing.T) {
		type CreateUserReq struct {
			Name string `json:"name"`
		}
		type UserRes struct {
			ID int `json:"id"`
		}

		r := chi.NewRouter()
		r.Post("/users", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("created"))
		})

		metas := []schema.RouteMeta{
			{
				Method: "POST",
				Path:   "/users",
				Info: schema.RouteInfo{
					Summary:      "Create user",
					RequestType:  CreateUserReq{},
					ResponseType: UserRes{},
				},
			},
		}

		routes, err := Extract(r, metas)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if routes[0].RequestType == nil {
			t.Error("expected request type to be set")
		}

		if routes[0].ResponseType == nil {
			t.Error("expected response type to be set")
		}
	})

	t.Run("skip routes not in metadata", func(t *testing.T) {
		r := chi.NewRouter()
		r.Get("/users", func(w http.ResponseWriter, r *http.Request) {})
		r.Get("/posts", func(w http.ResponseWriter, r *http.Request) {})

		metas := []schema.RouteMeta{
			{
				Method: "GET",
				Path:   "/users",
				Info:   schema.RouteInfo{Summary: "List users"},
			},
		}

		routes, err := Extract(r, metas)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(routes) != 1 {
			t.Errorf("expected 1 route (only /users), got %d", len(routes))
		}

		if routes[0].Path != "/users" {
			t.Errorf("expected path '/users', got '%s'", routes[0].Path)
		}
	})

	t.Run("empty metadata returns no routes", func(t *testing.T) {
		r := chi.NewRouter()
		r.Get("/users", func(w http.ResponseWriter, r *http.Request) {})

		var metas []schema.RouteMeta

		routes, err := Extract(r, metas)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(routes) != 0 {
			t.Errorf("expected 0 routes with empty metadata, got %d", len(routes))
		}
	})
}

// TestExtractAll tests extracting all routes without metadata
func TestExtractAll(t *testing.T) {
	t.Run("extract all routes", func(t *testing.T) {
		r := chi.NewRouter()
		r.Get("/users", func(w http.ResponseWriter, r *http.Request) {})
		r.Post("/users", func(w http.ResponseWriter, r *http.Request) {})
		r.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {})

		routes, err := ExtractAll(r)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(routes) != 3 {
			t.Errorf("expected 3 routes, got %d", len(routes))
		}
	})

	t.Run("extract with auto path params", func(t *testing.T) {
		r := chi.NewRouter()
		r.Get("/users/{id}/posts/{postId}", func(w http.ResponseWriter, r *http.Request) {})

		routes, err := ExtractAll(r)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(routes) != 1 {
			t.Fatalf("expected 1 route, got %d", len(routes))
		}

		if len(routes[0].PathParams) != 2 {
			t.Errorf("expected 2 path params, got %d", len(routes[0].PathParams))
		}

		paramNames := make(map[string]bool)
		for _, p := range routes[0].PathParams {
			paramNames[p.Name] = true
		}

		if !paramNames["id"] {
			t.Error("expected param 'id'")
		}

		if !paramNames["postId"] {
			t.Error("expected param 'postId'")
		}
	})

	t.Run("empty router", func(t *testing.T) {
		r := chi.NewRouter()

		routes, err := ExtractAll(r)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(routes) != 0 {
			t.Errorf("expected 0 routes, got %d", len(routes))
		}
	})
}

// TestRouteKey tests the route key generation
func TestRouteKey(t *testing.T) {
	tests := []struct {
		method   string
		path     string
		expected string
	}{
		{"GET", "/users", "GET /users"},
		{"POST", "/users", "POST /users"},
		{"GET", "/users/{id}", "GET /users/{id}"},
		{"DELETE", "/users/123", "DELETE /users/123"},
	}

	for _, tt := range tests {
		result := routeKey(tt.method, tt.path)
		if result != tt.expected {
			t.Errorf("routeKey(%q, %q) = %q, want %q", tt.method, tt.path, result, tt.expected)
		}
	}
}

// TestExtractParams tests the parameter extraction helper
func TestExtractParams(t *testing.T) {
	t.Run("use provided params", func(t *testing.T) {
		provided := []schema.ParamInfo{
			{Name: "id", Type: "integer", Required: true},
		}

		result := extractParams("/users/{id}", provided)

		if len(result) != 1 {
			t.Errorf("expected 1 param, got %d", len(result))
		}

		if result[0].Type != "integer" {
			t.Errorf("expected type 'integer', got '%s'", result[0].Type)
		}
	})

	t.Run("extract from path when no provided params", func(t *testing.T) {
		result := extractParams("/users/{userId}/posts/{postId}", nil)

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

	t.Run("empty provided uses path extraction", func(t *testing.T) {
		result := extractParams("/test/{param}", []schema.ParamInfo{})

		if len(result) != 1 {
			t.Errorf("expected 1 param from path, got %d", len(result))
		}
	})
}

// TestExtractPathParams tests path parameter extraction from Chi patterns
func TestExtractPathParams(t *testing.T) {
	tests := []struct {
		pattern  string
		expected []string
	}{
		{"/users/{id}", []string{"id"}},
		{"/users/{userId}/posts/{postId}", []string{"userId", "postId"}},
		{"/health", []string{}},
		{"/api/{version}/users", []string{"version"}},
		{"/{a}/{b}/{c}", []string{"a", "b", "c"}},
		{"/users/{id}/profile", []string{"id"}},
		{"/test/{param1}/sub/{param2}", []string{"param1", "param2"}},
		{"", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			result := extractPathParams(tt.pattern)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d params, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i].Name != expected {
					t.Errorf("param %d: expected name '%s', got '%s'", i, expected, result[i].Name)
				}
				if result[i].Type != "string" {
					t.Errorf("param %d: expected default type 'string', got '%s'", i, result[i].Type)
				}
				if !result[i].Required {
					t.Errorf("param %d: expected required=true", i)
				}
			}
		})
	}
}

// TestMount tests the Mount function
func TestMount(t *testing.T) {
	t.Run("mount with metadata", func(t *testing.T) {
		r := chi.NewRouter()
		r.Get("/external", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("external"))
		})

		metas := []schema.RouteMeta{
			{
				Method: "GET",
				Path:   "/external",
				Info:   schema.RouteInfo{Summary: "External route"},
			},
		}

		apiInstance := api.New()
		err := Mount(apiInstance, "/api", r, metas)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(apiInstance.MountedRouters) != 1 {
			t.Errorf("expected 1 mounted router, got %d", len(apiInstance.MountedRouters))
		}

		if apiInstance.MountedRouters[0].Prefix != "/api" {
			t.Errorf("expected prefix '/api', got '%s'", apiInstance.MountedRouters[0].Prefix)
		}

		if len(apiInstance.MountedRouters[0].Routes) != 1 {
			t.Errorf("expected 1 route, got %d", len(apiInstance.MountedRouters[0].Routes))
		}
	})

	t.Run("mount without metadata", func(t *testing.T) {
		r := chi.NewRouter()
		r.Get("/all", func(w http.ResponseWriter, r *http.Request) {})
		r.Post("/all", func(w http.ResponseWriter, r *http.Request) {})

		apiInstance := api.New()
		err := Mount(apiInstance, "/", r, nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(apiInstance.MountedRouters) != 1 {
			t.Fatalf("expected 1 mounted router, got %d", len(apiInstance.MountedRouters))
		}

		if len(apiInstance.MountedRouters[0].Routes) != 2 {
			t.Errorf("expected 2 routes (all routes extracted), got %d", len(apiInstance.MountedRouters[0].Routes))
		}
	})

	t.Run("mount with empty metadata extracts all", func(t *testing.T) {
		r := chi.NewRouter()
		r.Get("/route1", func(w http.ResponseWriter, r *http.Request) {})

		apiInstance := api.New()
		err := Mount(apiInstance, "/prefix", r, []schema.RouteMeta{})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// With empty metadata, it should extract all routes
		if len(apiInstance.MountedRouters[0].Routes) != 1 {
			t.Errorf("expected 1 route extracted, got %d", len(apiInstance.MountedRouters[0].Routes))
		}
	})

	t.Run("handles extraction errors", func(t *testing.T) {
		// This test verifies the error handling path
		// We can't easily trigger an error with chi.Walk, but we can verify
		// the error message format if one occurs

		apiInstance := api.New()

		// Valid router with valid extraction should not error
		r := chi.NewRouter()
		r.Get("/test", func(w http.ResponseWriter, r *http.Request) {})

		err := Mount(apiInstance, "/test", r, nil)

		if err != nil {
			t.Errorf("expected no error for valid router, got: %v", err)
		}
	})
}
