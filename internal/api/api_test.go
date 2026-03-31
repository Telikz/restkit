package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ep "github.com/reststore/restkit/internal/endpoints"
	"github.com/reststore/restkit/internal/schema"
)

// TestNew tests the Api constructor
func TestNew(t *testing.T) {
	api := New()
	if api == nil {
		t.Fatal("New() returned nil")
	}
	// Note: New() returns a zero-valued struct where slices are nil
	// This is fine - nil slices work correctly with append
}

// TestApiBuilder tests the builder methods
func TestApiBuilder(t *testing.T) {
	t.Run("WithVersion", func(t *testing.T) {
		api := New().WithVersion("1.0.0")
		if api.Version != "1.0.0" {
			t.Errorf("expected version '1.0.0', got '%s'", api.Version)
		}
	})

	t.Run("WithTitle", func(t *testing.T) {
		api := New().WithTitle("Test API")
		if api.Title != "Test API" {
			t.Errorf("expected title 'Test API', got '%s'", api.Title)
		}
	})

	t.Run("WithDescription", func(t *testing.T) {
		api := New().WithDescription("A test API")
		if api.Description != "A test API" {
			t.Errorf("expected description 'A test API', got '%s'", api.Description)
		}
	})

	t.Run("Chained builders", func(t *testing.T) {
		api := New().
			WithVersion("2.0.0").
			WithTitle("Chained API").
			WithDescription("Testing chains")
		if api.Version != "2.0.0" || api.Title != "Chained API" {
			t.Error("builder chaining failed")
		}
	})
}

// TestAddEndpoint tests endpoint registration
func TestAddEndpoint(t *testing.T) {
	api := New()

	endpoint := ep.NewEndpointRes[string]().
		WithMethod("GET").
		WithPath("/test").
		WithHandler(func(ctx context.Context) (string, error) {
			return "test", nil
		})

	api.AddEndpoint(endpoint)

	if len(api.Endpoints) != 1 {
		t.Errorf("expected 1 endpoint, got %d", len(api.Endpoints))
	}

	if api.Endpoints[0].GetPath() != "/test" {
		t.Errorf("expected path '/test', got '%s'", api.Endpoints[0].GetPath())
	}
}

// TestAddGroup tests group registration
func TestAddGroup(t *testing.T) {
	api := New()

	endpoint := ep.NewEndpointRes[string]().
		WithMethod("GET").
		WithPath("/users").
		WithHandler(func(ctx context.Context) (string, error) {
			return "users", nil
		})

	group := ep.NewGroup("/api/v1").WithEndpoints(endpoint)

	api.AddGroup(group)

	if len(api.Groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(api.Groups))
	}

	if len(api.Endpoints) != 1 {
		t.Errorf("expected endpoints from group to be added, got %d", len(api.Endpoints))
	}

	if api.Endpoints[0].GetPath() != "/api/v1/users" {
		t.Errorf("expected prefixed path '/api/v1/users', got '%s'", api.Endpoints[0].GetPath())
	}
}

// TestWithSwaggerUI tests swagger UI configuration
func TestWithSwaggerUI(t *testing.T) {
	t.Run("default path", func(t *testing.T) {
		api := New().WithSwaggerUI()
		if !api.SwaggerUIEnabled {
			t.Error("SwaggerUI should be enabled")
		}
		if api.SwaggerUIPath != "/swagger" {
			t.Errorf("expected default path '/swagger', got '%s'", api.SwaggerUIPath)
		}
	})

	t.Run("custom path", func(t *testing.T) {
		api := New().WithSwaggerUI("/docs")
		if api.SwaggerUIPath != "/docs" {
			t.Errorf("expected custom path '/docs', got '%s'", api.SwaggerUIPath)
		}
	})

	t.Run("empty path uses default", func(t *testing.T) {
		api := New().WithSwaggerUI("")
		if api.SwaggerUIPath != "/swagger" {
			t.Errorf("expected default path for empty string, got '%s'", api.SwaggerUIPath)
		}
	})
}

// TestWithSwaggerUIPath tests the deprecated path setter
func TestWithSwaggerUIPath(t *testing.T) {
	api := New().WithSwaggerUIPath("/api-docs")
	if api.SwaggerUIPath != "/api-docs" {
		t.Errorf("expected path '/api-docs', got '%s'", api.SwaggerUIPath)
	}
}

// TestWithMiddleware tests middleware registration
func TestWithMiddleware(t *testing.T) {
	api := New()

	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	api.WithMiddleware(middleware1, middleware2)

	if len(api.Middleware) != 2 {
		t.Errorf("expected 2 middleware, got %d", len(api.Middleware))
	}
}

// TestMountRouter tests router mounting
func TestMountRouter(t *testing.T) {
	api := New()

	mountedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	routes := []schema.MountedRoute{
		{
			Method:  "GET",
			Path:    "/external",
			Handler: mountedHandler,
		},
	}

	api.MountRouter("/external-api", mountedHandler, routes)

	if len(api.MountedRouters) != 1 {
		t.Errorf("expected 1 mounted router, got %d", len(api.MountedRouters))
	}

	if api.MountedRouters[0].Prefix != "/external-api" {
		t.Errorf("expected prefix '/external-api', got '%s'", api.MountedRouters[0].Prefix)
	}
}

// TestMux tests the HTTP mux assembly
func TestMux(t *testing.T) {
	t.Run("basic endpoint registration", func(t *testing.T) {
		api := New()

		endpoint := ep.NewEndpointRes[string]().
			WithMethod("GET").
			WithPath("/test").
			WithHandler(func(ctx context.Context) (string, error) {
				return "test", nil
			})

		api.AddEndpoint(endpoint)
		mux := api.Mux()

		if mux == nil {
			t.Fatal("Mux() returned nil")
		}

		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		if !strings.Contains(rec.Body.String(), `"test"`) {
			t.Errorf("expected body to contain '\"test\"', got '%s'", rec.Body.String())
		}
	})

	t.Run("global middleware application", func(t *testing.T) {
		api := New()

		middlewareCalled := false
		middleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				middlewareCalled = true
				next.ServeHTTP(w, r)
			})
		}

		endpoint := ep.NewEndpointRes[string]().
			WithMethod("GET").
			WithPath("/middleware-test").
			WithHandler(func(ctx context.Context) (string, error) {
				return "ok", nil
			})

		api.WithMiddleware(middleware).AddEndpoint(endpoint)
		mux := api.Mux()

		req := httptest.NewRequest("GET", "/middleware-test", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if !middlewareCalled {
			t.Error("global middleware was not called")
		}
	})

	t.Run("mounted router with prefix", func(t *testing.T) {
		api := New()

		mountedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("mounted"))
		})

		routes := []schema.MountedRoute{}
		api.MountRouter("/api", mountedHandler, routes)

		mux := api.Mux()
		req := httptest.NewRequest("GET", "/api/", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("swagger UI enabled", func(t *testing.T) {
		api := New().
			WithTitle("Test API").
			WithDescription("Test").
			WithVersion("1.0.0").
			WithSwaggerUI()

		mux := api.Mux()

		// Test swagger UI HTML endpoint
		req := httptest.NewRequest("GET", "/swagger", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected swagger UI status 200, got %d", rec.Code)
		}

		contentType := rec.Header().Get("Content-Type")
		if contentType != "text/html" {
			t.Errorf("expected Content-Type 'text/html', got '%s'", contentType)
		}

		// Test OpenAPI JSON endpoint
		req2 := httptest.NewRequest("GET", "/swagger/openapi.json", nil)
		rec2 := httptest.NewRecorder()
		mux.ServeHTTP(rec2, req2)

		if rec2.Code != http.StatusOK {
			t.Errorf("expected OpenAPI status 200, got %d", rec2.Code)
		}

		contentType2 := rec2.Header().Get("Content-Type")
		if contentType2 != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got '%s'", contentType2)
		}
	})
}

// TestGenerateOpenAPI tests the OpenAPI spec generation
func TestGenerateOpenAPI(t *testing.T) {
	api := New().
		WithTitle("Test API").
		WithDescription("A test API").
		WithVersion("1.0.0")

	endpoint := ep.NewEndpointRes[string]().
		WithMethod("GET").
		WithPath("/ping").
		WithTitle("Ping").
		WithDescription("Health check").
		WithHandler(func(ctx context.Context) (string, error) {
			return "pong", nil
		})

	api.AddEndpoint(endpoint)

	spec := api.GenerateOpenAPI()

	// Check spec structure
	if spec["openapi"] != "3.0.0" {
		t.Errorf("expected openapi version '3.0.0', got '%v'", spec["openapi"])
	}

	info, ok := spec["info"].(map[string]any)
	if !ok {
		t.Fatal("info should be a map")
	}

	if info["title"] != "Test API" {
		t.Errorf("expected title 'Test API', got '%v'", info["title"])
	}

	// Check paths
	paths, ok := spec["paths"].(map[string]any)
	if !ok {
		t.Fatal("paths should be a map")
	}

	if _, ok := paths["/ping"]; !ok {
		t.Error("/ping path should be in spec")
	}
}

// TestServeOpenAPI tests the OpenAPI HTTP handler
func TestServeOpenAPI(t *testing.T) {
	api := New().
		WithTitle("Test API").
		WithVersion("1.0.0")

	req := httptest.NewRequest("GET", "/openapi.json", nil)
	rec := httptest.NewRecorder()

	api.ServeOpenAPI(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}

	// Check that body contains valid JSON
	if rec.Body.Len() == 0 {
		t.Error("response body should not be empty")
	}
}

// TestServeOpenAPIWithMountedRoutes tests that mounted routes appear in spec
func TestGenerateOpenAPIWithMountedRoutes(t *testing.T) {
	api := New().
		WithTitle("Test API").
		WithVersion("1.0.0")

	mountedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	routes := []schema.MountedRoute{
		{
			Method:      "GET",
			Path:        "/external",
			Handler:     mountedHandler,
			Summary:     "External Route",
			Description: "From mounted router",
		},
	}

	api.MountRouter("/api", mountedHandler, routes)

	spec := api.GenerateOpenAPI()

	paths, ok := spec["paths"].(map[string]any)
	if !ok {
		t.Fatal("paths should be a map")
	}

	// Check that mounted route appears in paths
	if _, ok := paths["/api/external"]; !ok {
		t.Error("mounted route /api/external should be in spec")
	}
}

// TestServeSwaggerUI tests the swagger UI HTTP handler
func TestServeSwaggerUI(t *testing.T) {
	api := New().WithSwaggerUI("/docs")

	req := httptest.NewRequest("GET", "/docs", nil)
	rec := httptest.NewRecorder()

	api.serveSwaggerUI(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "text/html" {
		t.Errorf("expected Content-Type 'text/html', got '%s'", contentType)
	}

	body := rec.Body.String()
	if !contains(body, "swagger-ui") && !contains(body, "Swagger UI") {
		t.Error("response body should contain swagger UI content")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsInternal(s, substr)))
}

func containsInternal(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
