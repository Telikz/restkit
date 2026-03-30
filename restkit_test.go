package restkit_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	rest "github.com/telikz/restkit"
)

// Test Types
type TestRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type TestResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type PingResponse struct {
	Message string `json:"message"`
}

// Test NewApi
func TestNewApi(t *testing.T) {
	api := rest.NewApi()
	if api == nil {
		t.Fatal("NewApi() returned nil")
	}
}

// Test NewGroup
func TestNewGroup(t *testing.T) {
	group := rest.NewGroup("/api/v1")
	if group == nil {
		t.Fatal("NewGroup() returned nil")
	}
	if group.Prefix != "/api/v1" {
		t.Errorf("Expected prefix '/api/v1', got '%s'", group.Prefix)
	}
}

// Test NewEndpoint (Req/Res)
func TestNewEndpoint(t *testing.T) {
	endpoint := rest.NewEndpoint[TestRequest, TestResponse]()
	if endpoint == nil {
		t.Fatal("NewEndpoint() returned nil")
	}

	// Check default method is POST
	if endpoint.GetMethod() != http.MethodPost {
		t.Errorf("Expected default method POST, got %s", endpoint.GetMethod())
	}
}

// Test NewEndpointRes
func TestNewEndpointRes(t *testing.T) {
	endpoint := rest.NewEndpointRes[PingResponse]()
	if endpoint == nil {
		t.Fatal("NewEndpointRes() returned nil")
	}

	// Check default method is GET
	if endpoint.GetMethod() != http.MethodGet {
		t.Errorf("Expected default method GET, got %s", endpoint.GetMethod())
	}
}

// Test NewEndpointReq
func TestNewEndpointReq(t *testing.T) {
	endpoint := rest.NewEndpointReq[TestRequest]()
	if endpoint == nil {
		t.Fatal("NewEndpointReq() returned nil")
	}

	// Check default method is DELETE
	if endpoint.GetMethod() != http.MethodDelete {
		t.Errorf("Expected default method DELETE, got %s", endpoint.GetMethod())
	}
}

// Test ExtractPathParams
func TestExtractPathParams(t *testing.T) {
	tests := []struct {
		pattern  string
		path     string
		expected map[string]string
	}{
		{
			pattern:  "/users/{id}",
			path:     "/users/123",
			expected: map[string]string{"id": "123"},
		},
		{
			pattern:  "/users/{id}/posts/{postId}",
			path:     "/users/123/posts/456",
			expected: map[string]string{"id": "123", "postId": "456"},
		},
		{
			pattern:  "/users",
			path:     "/users",
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		params := rest.ExtractPathParams(tt.pattern, tt.path)
		for key, expectedValue := range tt.expected {
			if params[key] != expectedValue {
				t.Errorf("ExtractPathParams(%s, %s): expected %s=%s, got %s",
					tt.pattern, tt.path, key, expectedValue, params[key])
			}
		}
	}
}

// Test URLParam
func TestURLParam(t *testing.T) {
	// Create an endpoint that extracts URL params
	endpoint := rest.NewEndpointRes[map[string]string]().
		WithPath("/users/{id}").
		WithHandler(func(ctx context.Context) (map[string]string, error) {
			id := rest.URLParam(ctx, "id")
			return map[string]string{"id": id}, nil
		})

	// Create API and add endpoint
	api := rest.NewApi().Add(endpoint)
	mux := api.Mux()

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, `"id":"123"`) {
		t.Errorf("Expected response to contain id=123, got: %s", body)
	}
}

// Test RouteCtxFromContext
func TestRouteCtxFromContext(t *testing.T) {
	endpoint := rest.NewEndpointRes[map[string]string]().
		WithPath("/users/{id}/posts/{postId}").
		WithHandler(func(ctx context.Context) (map[string]string, error) {
			routeCtx := rest.RouteCtxFromContext(ctx)
			if routeCtx == nil {
				return nil, errors.New("route context is nil")
			}
			return map[string]string{
				"id":     routeCtx.URLParam("id"),
				"postId": routeCtx.URLParam("postId"),
			}, nil
		})

	api := rest.NewApi().Add(endpoint)
	mux := api.Mux()

	req := httptest.NewRequest(http.MethodGet, "/users/123/posts/456", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, `"id":"123"`) || !strings.Contains(body, `"postId":"456"`) {
		t.Errorf("Expected response to contain both params, got: %s", body)
	}
}

// Test JSONBinder
func TestJSONBinder(t *testing.T) {
	binder := rest.JSONBinder[TestRequest]()
	if binder == nil {
		t.Fatal("JSONBinder() returned nil")
	}

	// Create a request with JSON body
	body := `{"name":"John","email":"john@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	result, err := binder(req)
	if err != nil {
		t.Fatalf("JSONBinder failed: %v", err)
	}

	if result.Name != "John" {
		t.Errorf("Expected name 'John', got '%s'", result.Name)
	}
	if result.Email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got '%s'", result.Email)
	}
}

// Test JSONWriter
func TestJSONWriter(t *testing.T) {
	writer := rest.JSONWriter[TestResponse]()
	if writer == nil {
		t.Fatal("JSONWriter() returned nil")
	}

	rec := httptest.NewRecorder()
	response := TestResponse{ID: 1, Name: "John", Email: "john@example.com"}

	err := writer(rec, response)
	if err != nil {
		t.Fatalf("JSONWriter failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, `"id":1`) {
		t.Errorf("Expected response to contain id, got: %s", body)
	}
}

// Test JSONErrorWriter
func TestJSONErrorWriter(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	testErr := errors.New("test error")

	rest.JSONErrorWriter(rec, req, testErr)

	// JSONErrorWriter returns 400 BadRequest
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}
}

// Test SchemaFrom
func TestSchemaFrom(t *testing.T) {
	schema := rest.SchemaFrom[TestResponse]()
	if schema == nil {
		t.Fatal("SchemaFrom() returned nil")
	}

	// Check that schema has type property
	schemaType, ok := schema["type"]
	if !ok {
		t.Error("Schema missing 'type' property")
	}
	if schemaType != "object" {
		t.Errorf("Expected schema type 'object', got '%v'", schemaType)
	}

	// Check that schema has properties
	properties, ok := schema["properties"]
	if !ok {
		t.Error("Schema missing 'properties'")
	}
	if properties == nil {
		t.Error("Schema properties is nil")
	}
}

// Test PathParamBinder
func TestPathParamBinder(t *testing.T) {
	// Test with int converter
	intBinder := rest.PathParamBinder(rest.StringToInt)
	if intBinder == nil {
		t.Fatal("PathParamBinder(StringToInt) returned nil")
	}

	// Test with string converter (no-op)
	stringBinder := rest.PathParamBinder(rest.StringToString)
	if stringBinder == nil {
		t.Fatal("PathParamBinder(StringToString) returned nil")
	}
}

// Test StringToInt
func TestStringToInt(t *testing.T) {
	val, err := rest.StringToInt("123")
	if err != nil {
		t.Fatalf("StringToInt failed: %v", err)
	}
	if val != 123 {
		t.Errorf("Expected 123, got %d", val)
	}

	// Test invalid input
	_, err = rest.StringToInt("abc")
	if err == nil {
		t.Error("StringToInt should return error for non-numeric string")
	}
}

// Test StringToString
func TestStringToString(t *testing.T) {
	val, err := rest.StringToString("hello")
	if err != nil {
		t.Fatalf("StringToString failed: %v", err)
	}
	if val != "hello" {
		t.Errorf("Expected 'hello', got '%s'", val)
	}
}

// Test LoggingMiddleware
func TestLoggingMiddleware(t *testing.T) {
	middleware := rest.LoggingMiddleware()
	if middleware == nil {
		t.Fatal("LoggingMiddleware() returned nil")
	}
}

// Test CORSMiddleware
func TestCORSMiddleware(t *testing.T) {
	middleware := rest.CORSMiddleware()
	if middleware == nil {
		t.Fatal("CORSMiddleware() returned nil")
	}
}

// Test RecoveryMiddleware
func TestRecoveryMiddleware(t *testing.T) {
	middleware := rest.RecoveryMiddleware()
	if middleware == nil {
		t.Fatal("RecoveryMiddleware() returned nil")
	}
}

// Test Api with Groups
func TestApiWithGroups(t *testing.T) {
	// Create a group
	userGroup := rest.NewGroup("/users").
		WithTitle("Users").
		WithDescription("User management endpoints")

	// Add endpoints to group
	listEndpoint := rest.NewEndpointRes[[]TestResponse]().
		WithPath("/").
		WithMethod(http.MethodGet).
		WithHandler(func(ctx context.Context) ([]TestResponse, error) {
			return []TestResponse{}, nil
		})

	userGroup.WithEndpoints(listEndpoint)

	// Create API and add group
	api := rest.NewApi().
		WithTitle("Test API").
		WithVersion("1.0.0").
		WithDescription("Test API Description").
		AddGroup(userGroup)

	mux := api.Mux()

	// Test group endpoint
	req := httptest.NewRequest(http.MethodGet, "/users/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

// Test Api WithMiddleware
func TestApiWithMiddleware(t *testing.T) {
	called := false
	testMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			next.ServeHTTP(w, r)
		})
	}

	endpoint := rest.NewEndpointRes[PingResponse]().
		WithPath("/ping").
		WithMethod(http.MethodGet).
		WithHandler(func(ctx context.Context) (PingResponse, error) {
			return PingResponse{Message: "pong"}, nil
		})

	api := rest.NewApi().
		WithMiddleware(testMiddleware).
		Add(endpoint)

	mux := api.Mux()

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if !called {
		t.Error("Middleware was not called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

// Test Validation types and functions
func TestValidationTypes(t *testing.T) {
	// Test NewValidation
	validation := rest.NewValidation()
	if validation.HasErrors() {
		t.Error("New validation should not have errors")
	}

	// Test ValidationFailed
	failedValidation := rest.ValidationFailed(400, "error", "validation failed", "field1", "error message")
	if !failedValidation.HasErrors() {
		t.Error("Failed validation should have errors")
	}

	// Test ValidationFailedMulti
	multiValidation := rest.ValidationFailedMulti(
		422, "validation", "multiple errors",
		rest.ValidationError{Field: "field1", Message: "error1"},
		rest.ValidationError{Field: "field2", Message: "error2"},
	)
	if len(multiValidation.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(multiValidation.Errors))
	}
}

// Test ValidateStruct
func TestValidateStruct(t *testing.T) {
	ctx := context.Background()

	// Valid struct
	validReq := TestRequest{Name: "John", Email: "john@example.com"}
	result := rest.ValidateStruct(ctx, validReq)
	if result.HasErrors() {
		t.Errorf("Valid struct should not have errors, got: %v", result.Errors)
	}

	// Invalid struct - missing required fields
	invalidReq := TestRequest{}
	result = rest.ValidateStruct(ctx, invalidReq)
	if !result.HasErrors() {
		t.Error("Invalid struct should have errors")
	}

	// Invalid struct - invalid email
	invalidEmailReq := TestRequest{Name: "John", Email: "invalid-email"}
	result = rest.ValidateStruct(ctx, invalidEmailReq)
	if !result.HasErrors() {
		t.Error("Invalid email should produce validation error")
	}
}

// Test Endpoint WithValidation
func TestEndpointWithValidation(t *testing.T) {
	endpoint := rest.NewEndpointRes[PingResponse]().
		WithPath("/ping").
		WithMethod(http.MethodGet).
		WithValidation(func(ctx context.Context) rest.ValidationResult {
			// Always fail validation for testing
			return rest.ValidationFailed(400, "test", "validation failed", "field", "error")
		}).
		WithHandler(func(ctx context.Context) (PingResponse, error) {
			return PingResponse{Message: "pong"}, nil
		})

	api := rest.NewApi().Add(endpoint)
	mux := api.Mux()

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 (BadRequest), got %d", rec.Code)
	}
}

// Test Endpoint WithMiddleware
func TestEndpointWithMiddleware(t *testing.T) {
	called := false
	endpointMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			next.ServeHTTP(w, r)
		})
	}

	endpoint := rest.NewEndpointRes[PingResponse]().
		WithPath("/ping").
		WithMethod(http.MethodGet).
		WithMiddleware(endpointMiddleware).
		WithHandler(func(ctx context.Context) (PingResponse, error) {
			return PingResponse{Message: "pong"}, nil
		})

	api := rest.NewApi().Add(endpoint)
	mux := api.Mux()

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if !called {
		t.Error("Endpoint middleware was not called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}
