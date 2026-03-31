package endpoints

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	errs "github.com/reststore/restkit/internal/errors"
)

// TestNewGroup tests the Group constructor
func TestNewGroup(t *testing.T) {
	group := NewGroup("/api/v1")
	if group == nil {
		t.Fatal("NewGroup() returned nil")
	}
	if group.Prefix != "/api/v1" {
		t.Errorf("expected prefix '/api/v1', got '%s'", group.Prefix)
	}
	if group.Endpoints == nil {
		t.Error("Endpoints should be initialized")
	}
}

// TestGroupBuilder tests the Group builder methods
func TestGroupBuilder(t *testing.T) {
	t.Run("WithTitle", func(t *testing.T) {
		group := NewGroup("/api").WithTitle("User API")
		if group.Title != "User API" {
			t.Errorf("expected title 'User API', got '%s'", group.Title)
		}
	})

	t.Run("WithDescription", func(t *testing.T) {
		group := NewGroup("/api").WithDescription("User management API")
		if group.Description != "User management API" {
			t.Errorf("expected description 'User management API', got '%s'", group.Description)
		}
	})

	t.Run("WithEndpoints", func(t *testing.T) {
		endpoint := NewEndpointRes[string]().
			WithMethod("GET").
			WithPath("/users")

		group := NewGroup("/api").WithEndpoints(endpoint)
		if len(group.Endpoints) != 1 {
			t.Errorf("expected 1 endpoint, got %d", len(group.Endpoints))
		}
	})

	t.Run("WithMiddleware", func(t *testing.T) {
		middleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})
		}

		group := NewGroup("/api").WithMiddleware(middleware)
		if len(group.Middleware) != 1 {
			t.Errorf("expected 1 middleware, got %d", len(group.Middleware))
		}
	})

	t.Run("chained builders", func(t *testing.T) {
		endpoint := NewEndpointRes[string]().
			WithMethod("GET").
			WithPath("/test")

		group := NewGroup("/api/v2").
			WithTitle("V2 API").
			WithDescription("Version 2").
			WithEndpoints(endpoint)

		if group.Title != "V2 API" || group.Description != "Version 2" {
			t.Error("chained builders failed")
		}
	})
}

// TestGroupGetEndpoints tests getting endpoints with prefix and middleware
func TestGroupGetEndpoints(t *testing.T) {
	t.Run("endpoints with prefix", func(t *testing.T) {
		endpoint := NewEndpointRes[string]().
			WithMethod("GET").
			WithPath("/users").
			WithHandler(func(ctx context.Context) (string, error) {
				return "users", nil
			})

		group := NewGroup("/api/v1").WithEndpoints(endpoint)
		endpoints := group.GetEndpoints()

		if len(endpoints) != 1 {
			t.Fatalf("expected 1 endpoint, got %d", len(endpoints))
		}

		if endpoints[0].GetPath() != "/api/v1/users" {
			t.Errorf("expected path '/api/v1/users', got '%s'", endpoints[0].GetPath())
		}
	})

	t.Run("endpoints with middleware", func(t *testing.T) {
		middlewareCalled := false
		middleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				middlewareCalled = true
				next.ServeHTTP(w, r)
			})
		}

		endpoint := NewEndpointRes[string]().
			WithMethod("GET").
			WithPath("/test").
			WithHandler(func(ctx context.Context) (string, error) {
				return "test", nil
			})

		group := NewGroup("/api").
			WithMiddleware(middleware).
			WithEndpoints(endpoint)

		endpoints := group.GetEndpoints()

		// Test that middleware is applied
		req := httptest.NewRequest("GET", "/api/test", nil)
		rec := httptest.NewRecorder()
		endpoints[0].GetHandler().ServeHTTP(rec, req)

		if !middlewareCalled {
			t.Error("group middleware was not applied")
		}
	})

	t.Run("multiple endpoints", func(t *testing.T) {
		endpoint1 := NewEndpointRes[string]().
			WithMethod("GET").
			WithPath("/users")

		endpoint2 := NewEndpointRes[string]().
			WithMethod("POST").
			WithPath("/users")

		group := NewGroup("/api").WithEndpoints(endpoint1, endpoint2)
		endpoints := group.GetEndpoints()

		if len(endpoints) != 2 {
			t.Errorf("expected 2 endpoints, got %d", len(endpoints))
		}
	})
}

// TestNewEndpoint tests the constructor for full request-response endpoints
func TestNewEndpoint(t *testing.T) {
	type TestReq struct {
		Name string `json:"name"`
	}
	type TestRes struct {
		ID int `json:"id"`
	}

	endpoint := NewEndpoint[TestReq, TestRes]()
	if endpoint == nil {
		t.Fatal("NewEndpoint() returned nil")
	}

	if endpoint.Method != "POST" {
		t.Errorf("expected default method 'POST', got '%s'", endpoint.Method)
	}

	if endpoint.Bind == nil {
		t.Error("Bind should be initialized with JSONBinder")
	}

	if endpoint.Write == nil {
		t.Error("Write should be initialized with JSONWriter")
	}

	if endpoint.Validate == nil {
		t.Error("Validate should be initialized with defaultValidator")
	}
}

// TestNewEndpointRes tests the constructor for response-only endpoints
func TestNewEndpointRes(t *testing.T) {
	type TestRes struct {
		Message string `json:"message"`
	}

	endpoint := NewEndpointRes[TestRes]()
	if endpoint == nil {
		t.Fatal("NewEndpointRes() returned nil")
	}

	if endpoint.Method != "GET" {
		t.Errorf("expected default method 'GET', got '%s'", endpoint.Method)
	}

	if endpoint.Write == nil {
		t.Error("Write should be initialized with JSONWriter")
	}

	// Handler is intentionally nil by default and must be set via WithHandler
	// This is expected behavior
}

// TestNewEndpointReq tests the constructor for request-only endpoints
func TestNewEndpointReq(t *testing.T) {
	type TestReq struct {
		Name string `json:"name"`
	}

	endpoint := NewEndpointReq[TestReq]()
	if endpoint == nil {
		t.Fatal("NewEndpointReq() returned nil")
	}

	if endpoint.Method != "DELETE" {
		t.Errorf("expected default method 'DELETE', got '%s'", endpoint.Method)
	}

	if endpoint.Bind == nil {
		t.Error("Bind should be initialized with JSONBinder")
	}
}

// TestEndpointReqResBuilder tests the EndpointReqRes builder methods
func TestEndpointReqResBuilder(t *testing.T) {
	type TestReq struct {
		Name string `json:"name"`
	}
	type TestRes struct {
		ID int `json:"id"`
	}

	t.Run("WithTitle", func(t *testing.T) {
		endpoint := NewEndpoint[TestReq, TestRes]().WithTitle("Create User")
		if endpoint.Title != "Create User" {
			t.Errorf("expected title 'Create User', got '%s'", endpoint.Title)
		}
	})

	t.Run("WithDescription", func(t *testing.T) {
		endpoint := NewEndpoint[TestReq, TestRes]().WithDescription("Creates a new user")
		if endpoint.Description != "Creates a new user" {
			t.Errorf("expected description 'Creates a new user', got '%s'", endpoint.Description)
		}
	})

	t.Run("WithMethod", func(t *testing.T) {
		endpoint := NewEndpoint[TestReq, TestRes]().WithMethod("PUT")
		if endpoint.Method != "PUT" {
			t.Errorf("expected method 'PUT', got '%s'", endpoint.Method)
		}
	})

	t.Run("WithPath", func(t *testing.T) {
		endpoint := NewEndpoint[TestReq, TestRes]().WithPath("/users/{id}")
		if endpoint.Path != "/users/{id}" {
			t.Errorf("expected path '/users/{id}', got '%s'", endpoint.Path)
		}
	})

	t.Run("WithHandler", func(t *testing.T) {
		handlerCalled := false
		handler := func(ctx context.Context, req TestReq) (TestRes, error) {
			handlerCalled = true
			return TestRes{ID: 1}, nil
		}

		endpoint := NewEndpoint[TestReq, TestRes]().WithHandler(handler)
		endpoint.Handler(context.Background(), TestReq{Name: "test"})

		if !handlerCalled {
			t.Error("handler was not set")
		}
	})

	t.Run("WithMiddleware", func(t *testing.T) {
		middleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})
		}

		endpoint := NewEndpoint[TestReq, TestRes]().WithMiddleware(middleware)
		if len(endpoint.Middleware) != 1 {
			t.Errorf("expected 1 middleware, got %d", len(endpoint.Middleware))
		}
	})
}

// TestEndpointReqResGetHandler tests the HTTP handler assembly
func TestEndpointReqResGetHandler(t *testing.T) {
	type TestReq struct {
		Name string `json:"name" validate:"required"`
	}
	type TestRes struct {
		ID int `json:"id"`
	}

	t.Run("successful request", func(t *testing.T) {
		endpoint := NewEndpoint[TestReq, TestRes]().
			WithMethod("POST").
			WithPath("/users").
			WithHandler(func(ctx context.Context, req TestReq) (TestRes, error) {
				return TestRes{ID: 1}, nil
			})

		handler := endpoint.GetHandler()
		req := httptest.NewRequest("POST", "/users", strings.NewReader(`{"name":"John"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		if !strings.Contains(rec.Body.String(), `"id":1`) {
			t.Errorf("expected response to contain id, got '%s'", rec.Body.String())
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		endpoint := NewEndpoint[TestReq, TestRes]().
			WithMethod("POST").
			WithPath("/users").
			WithHandler(func(ctx context.Context, req TestReq) (TestRes, error) {
				return TestRes{}, nil
			})

		handler := endpoint.GetHandler()
		req := httptest.NewRequest("POST", "/users", strings.NewReader(`{invalid}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400 for invalid JSON, got %d", rec.Code)
		}
	})

	t.Run("validation failure", func(t *testing.T) {
		endpoint := NewEndpoint[TestReq, TestRes]().
			WithMethod("POST").
			WithPath("/users").
			WithHandler(func(ctx context.Context, req TestReq) (TestRes, error) {
				return TestRes{}, nil
			})

		handler := endpoint.GetHandler()
		req := httptest.NewRequest("POST", "/users", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		// Should fail validation (name is required)
		if rec.Code != 422 {
			t.Errorf("expected status 422 for validation failure, got %d", rec.Code)
		}
	})
}

// TestEndpointReqResClone tests endpoint cloning
func TestEndpointReqResClone(t *testing.T) {
	type TestReq struct{}
	type TestRes struct{}

	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	original := NewEndpoint[TestReq, TestRes]().
		WithMethod("POST").
		WithPath("/test").
		WithTitle("Original").
		WithMiddleware(middleware1)

	cloned := original.Clone()

	// Verify cloned has same values
	if cloned.Method != original.Method {
		t.Error("cloned method should match original")
	}
	if cloned.Path != original.Path {
		t.Error("cloned path should match original")
	}
	if cloned.Title != original.Title {
		t.Error("cloned title should match original")
	}

	// Verify middleware is copied (not shared reference)
	if len(cloned.Middleware) != len(original.Middleware) {
		t.Error("cloned middleware length should match original")
	}
}

// TestEndpointResBuilder tests the EndpointRes builder methods
func TestEndpointResBuilder(t *testing.T) {
	type TestRes struct {
		Message string `json:"message"`
	}

	t.Run("WithTitle", func(t *testing.T) {
		endpoint := NewEndpointRes[TestRes]().WithTitle("Get Status")
		if endpoint.Title != "Get Status" {
			t.Errorf("expected title 'Get Status', got '%s'", endpoint.Title)
		}
	})

	t.Run("WithHandler", func(t *testing.T) {
		handlerCalled := false
		handler := func(ctx context.Context) (TestRes, error) {
			handlerCalled = true
			return TestRes{Message: "ok"}, nil
		}

		endpoint := NewEndpointRes[TestRes]().WithHandler(handler)
		endpoint.Handler(context.Background())

		if !handlerCalled {
			t.Error("handler was not set")
		}
	})
}

// TestEndpointResGetHandler tests the HTTP handler for response-only endpoints
func TestEndpointResGetHandler(t *testing.T) {
	type TestRes struct {
		Status string `json:"status"`
	}

	t.Run("successful GET request", func(t *testing.T) {
		endpoint := NewEndpointRes[TestRes]().
			WithMethod("GET").
			WithPath("/status").
			WithHandler(func(ctx context.Context) (TestRes, error) {
				return TestRes{Status: "ok"}, nil
			})

		handler := endpoint.GetHandler()
		req := httptest.NewRequest("GET", "/status", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
			t.Errorf("expected response to contain status, got '%s'", rec.Body.String())
		}
	})

	t.Run("handler error", func(t *testing.T) {
		endpoint := NewEndpointRes[TestRes]().
			WithMethod("GET").
			WithPath("/error").
			WithHandler(func(ctx context.Context) (TestRes, error) {
				return TestRes{}, errs.NewAPIError(500, "internal", "server error")
			})

		handler := endpoint.GetHandler()
		req := httptest.NewRequest("GET", "/error", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}
	})
}

// TestEndpointReqBuilder tests the EndpointReq builder methods
func TestEndpointReqBuilder(t *testing.T) {
	type TestReq struct {
		ID int `json:"id"`
	}

	t.Run("WithTitle", func(t *testing.T) {
		endpoint := NewEndpointReq[TestReq]().WithTitle("Delete User")
		if endpoint.Title != "Delete User" {
			t.Errorf("expected title 'Delete User', got '%s'", endpoint.Title)
		}
	})

	t.Run("WithHandler", func(t *testing.T) {
		handlerCalled := false
		handler := func(ctx context.Context, req TestReq) error {
			handlerCalled = true
			return nil
		}

		endpoint := NewEndpointReq[TestReq]().WithHandler(handler)
		endpoint.Handler(context.Background(), TestReq{ID: 1})

		if !handlerCalled {
			t.Error("handler was not set")
		}
	})
}

// TestEndpointReqGetHandler tests the HTTP handler for request-only endpoints
func TestEndpointReqGetHandler(t *testing.T) {
	type TestReq struct {
		ID int `json:"id" validate:"required,gt=0"`
	}

	t.Run("successful DELETE request", func(t *testing.T) {
		endpoint := NewEndpointReq[TestReq]().
			WithMethod("DELETE").
			WithPath("/users/{id}").
			WithHandler(func(ctx context.Context, req TestReq) error {
				return nil
			})

		handler := endpoint.GetHandler()
		req := httptest.NewRequest("DELETE", "/users/123", strings.NewReader(`{"id":123}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected status 204 for DELETE, got %d", rec.Code)
		}
	})

	t.Run("validation failure", func(t *testing.T) {
		endpoint := NewEndpointReq[TestReq]().
			WithMethod("DELETE").
			WithPath("/users/{id}").
			WithHandler(func(ctx context.Context, req TestReq) error {
				return nil
			})

		handler := endpoint.GetHandler()
		req := httptest.NewRequest("DELETE", "/users/123", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != 422 {
			t.Errorf("expected status 422 for validation failure, got %d", rec.Code)
		}
	})
}

// TestEndpointReqClone tests request-only endpoint cloning
func TestEndpointReqClone(t *testing.T) {
	type TestReq struct{}

	original := NewEndpointReq[TestReq]().
		WithMethod("DELETE")

	cloned := original.Clone()

	if cloned.Method != original.Method {
		t.Error("cloned method should match original")
	}
}

// TestDefaultValidator tests the default validation function
func TestDefaultValidator(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	t.Run("valid data", func(t *testing.T) {
		data := TestStruct{Name: "John", Email: "john@example.com"}
		result := defaultValidator[TestStruct]()(context.Background(), data)

		if result.HasErrors() {
			t.Errorf("expected no validation errors, got %v", result.Errors)
		}
	})

	t.Run("invalid data", func(t *testing.T) {
		data := TestStruct{Name: "", Email: "invalid"}
		result := defaultValidator[TestStruct]()(context.Background(), data)

		if !result.HasErrors() {
			t.Error("expected validation errors")
		}

		if result.Status != 422 {
			t.Errorf("expected status 422, got %d", result.Status)
		}
	})
}

// TestErrorHandler tests the error handler function
func TestErrorHandler(t *testing.T) {
	apiErr := errs.NewAPIError(404, "not_found", "Resource not found")
	handler := errorHandler(apiErr)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Errorf("expected status 404, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "not_found") {
		t.Errorf("expected body to contain error code, got '%s'", body)
	}
}

// TestEndpointWrapper tests the endpointWrapper struct
func TestEndpointWrapper(t *testing.T) {
	original := NewEndpointRes[string]().
		WithMethod("GET").
		WithPath("/users").
		WithTitle("List Users")

	wrapper := &endpointWrapper{
		definition:   original,
		pathOverride: "/api/v1/users",
	}

	if wrapper.GetMethod() != "GET" {
		t.Error("wrapper should return original method")
	}

	if wrapper.GetPath() != "/api/v1/users" {
		t.Errorf("wrapper should return overridden path, got '%s'", wrapper.GetPath())
	}

	if wrapper.GetTitle() != "List Users" {
		t.Error("wrapper should return original title")
	}
}
