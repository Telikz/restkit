package docs

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	ep "github.com/reststore/restkit/internal/endpoints"
	"github.com/reststore/restkit/internal/schema"
)

func TestGenerateOpenAPI(t *testing.T) {
	t.Run("basic spec generation", func(t *testing.T) {
		endpoints := []ep.Route{}
		groups := []*ep.Group{}

		s := &OpenAPISpec{
			Title:       "Test API",
			Description: "Test Description",
			Version:     "1.0.0",
			Endpoints:   endpoints,
			Groups:      groups,
		}

		spec := GenerateOpenAPI(s)

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

		if info["description"] != "Test Description" {
			t.Errorf("expected description 'Test Description', got '%v'", info["description"])
		}

		if info["version"] != "1.0.0" {
			t.Errorf("expected version '1.0.0', got '%v'", info["version"])
		}
	})

	t.Run("with endpoints from groups", func(t *testing.T) {
		endpoint := ep.NewEndpointRes[string]().
			WithMethod("GET").
			WithPath("/ping").
			WithTitle("Ping").
			WithDescription("Health check")

		group := ep.NewGroup("/api/v1").
			WithTitle("API v1").
			WithDescription("Version 1 API").
			WithEndpoints(endpoint)

		groups := []*ep.Group{group}

		s := &OpenAPISpec{
			Title:       "API",
			Description: "",
			Version:     "1.0",
			Endpoints:   []ep.Route{},
			Groups:      groups,
		}

		spec := GenerateOpenAPI(s)

		// Check tags
		tags, ok := spec["tags"].([]map[string]any)
		if !ok {
			t.Fatal("tags should be a slice")
		}

		if len(tags) != 1 {
			t.Errorf("expected 1 tag, got %d", len(tags))
		}

		// Check paths
		paths, ok := spec["paths"].(map[string]any)
		if !ok {
			t.Fatal("paths should be a map")
		}

		// Group prefix now works with typed endpoints!
		if _, ok := paths["/api/v1/ping"]; !ok {
			t.Error("/api/v1/ping should be in paths")
		}
	})

	t.Run("with individual endpoints", func(t *testing.T) {
		endpoint := ep.NewEndpointRes[string]().
			WithMethod("GET").
			WithPath("/health").
			WithTitle("Health")

		endpoints := []ep.Route{endpoint}
		s := &OpenAPISpec{
			Title:       "API",
			Description: "",
			Version:     "1.0",
			Endpoints:   endpoints,
			Groups:      []*ep.Group{},
		}
		spec := GenerateOpenAPI(s)

		paths, ok := spec["paths"].(map[string]any)
		if !ok {
			t.Fatal("paths should be a map")
		}

		if _, ok := paths["/health"]; !ok {
			t.Error("/health should be in paths")
		}
	})

	t.Run("duplicate detection", func(t *testing.T) {
		// Create endpoint for group (will be prefixed to /api/ping)
		groupEndpoint := ep.NewEndpointRes[string]().
			WithMethod("GET").
			WithPath("/ping")

		group := ep.NewGroup("/api").WithEndpoints(groupEndpoint)

		// Create separate endpoint for individual registration (stays at /ping)
		individualEndpoint := ep.NewEndpointRes[string]().
			WithMethod("GET").
			WithPath("/ping")

		// Add both to group and individually - should result in two different paths
		endpoints := []ep.Route{individualEndpoint}
		groups := []*ep.Group{group}

		s := &OpenAPISpec{
			Title:       "API",
			Description: "",
			Version:     "1.0",
			Endpoints:   endpoints,
			Groups:      groups,
		}
		spec := GenerateOpenAPI(s)

		paths, ok := spec["paths"].(map[string]any)
		if !ok {
			t.Fatal("paths should be a map")
		}

		// Group prefix now works with typed endpoints!
		// The endpoint in the group gets prefixed to "/api/ping"
		// The individual endpoint stays at "/ping"
		// So we have two different paths, not duplicates
		pathDataGrouped, ok := paths["/api/ping"].(map[string]any)
		if !ok {
			t.Fatal("/api/ping should exist (grouped endpoint)")
		}

		pathDataIndividual, ok := paths["/ping"].(map[string]any)
		if !ok {
			t.Fatal("/ping should exist (individual endpoint)")
		}

		// Each should only have one operation
		if len(pathDataGrouped) != 1 {
			t.Errorf("expected 1 operation for grouped endpoint, got %d", len(pathDataGrouped))
		}
		if len(pathDataIndividual) != 1 {
			t.Errorf(
				"expected 1 operation for individual endpoint, got %d",
				len(pathDataIndividual),
			)
		}
	})
}

func TestBuildOperation(t *testing.T) {
	t.Run("basic operation", func(t *testing.T) {
		endpoint := ep.NewEndpointRes[string]().
			WithMethod("GET").
			WithPath("/test").
			WithTitle("Test Endpoint").
			WithDescription("A test endpoint")

		op := buildOperation(endpoint, []*ep.Group{})

		if op["summary"] != "Test Endpoint" {
			t.Errorf("expected summary 'Test Endpoint', got '%v'", op["summary"])
		}

		if op["description"] != "A test endpoint" {
			t.Errorf("expected description 'A test endpoint', got '%v'", op["description"])
		}
	})

	t.Run("operation with tags from group", func(t *testing.T) {
		endpoint := ep.NewEndpointRes[string]().
			WithMethod("GET").
			WithPath("/users")

		group := ep.NewGroup("/api").
			WithTitle("User API").
			WithEndpoints(endpoint)

		// Get the endpoint from the group (which has the prefix applied)
		groupEndpoints := group.GetEndpoints()
		if len(groupEndpoints) == 0 {
			t.Fatal("group should have endpoints")
		}

		op := buildOperation(groupEndpoints[0], []*ep.Group{group})

		tags, ok := op["tags"].([]string)
		if !ok {
			t.Fatalf("tags should be a slice of strings, got %T", op["tags"])
		}

		if len(tags) != 1 || tags[0] != "User API" {
			t.Errorf("expected tags ['User API'], got %v", tags)
		}
	})

	t.Run("operation with request schema", func(t *testing.T) {
		type TestReq struct {
			Name string `json:"name"`
		}

		endpoint := ep.NewEndpoint[TestReq, string]().
			WithMethod("POST").
			WithPath("/test")

		// Trigger schema generation by calling GetHandler
		_ = endpoint.GetHandler()

		op := buildOperation(endpoint, []*ep.Group{})

		reqBody, ok := op["requestBody"].(map[string]any)
		if !ok {
			t.Fatal("requestBody should be a map")
		}

		if reqBody["required"] != true {
			t.Error("requestBody should be required")
		}
	})

	t.Run("operation with response schema", func(t *testing.T) {
		type TestRes struct {
			ID int `json:"id"`
		}

		endpoint := ep.NewEndpointRes[TestRes]().
			WithMethod("GET").
			WithPath("/test")

		// Trigger schema generation by calling GetHandler
		_ = endpoint.GetHandler()

		op := buildOperation(endpoint, []*ep.Group{})

		responses, ok := op["responses"].(map[string]any)
		if !ok {
			t.Fatal("responses should be a map")
		}

		if _, ok := responses["200"]; !ok {
			t.Error("should have 200 response")
		}
	})

	t.Run("operation without response schema", func(t *testing.T) {
		endpoint := ep.NewEndpointReq[string]().
			WithMethod("DELETE").
			WithPath("/test")

		op := buildOperation(endpoint, []*ep.Group{})

		responses, ok := op["responses"].(map[string]any)
		if !ok {
			t.Fatal("responses should be a map")
		}

		if _, ok := responses["204"]; !ok {
			t.Error("should have 204 response when no response schema")
		}
	})

	t.Run("operation with path parameters", func(t *testing.T) {
		endpoint := ep.NewEndpointRes[string]().
			WithMethod("GET").
			WithPath("/users/{id}/posts/{postId}")

		op := buildOperation(endpoint, []*ep.Group{})

		parameters, ok := op["parameters"].([]map[string]any)
		if !ok {
			t.Fatal("parameters should be a slice")
		}

		if len(parameters) != 2 {
			t.Errorf("expected 2 parameters, got %d", len(parameters))
		}

		// Check parameter structure
		foundId := false
		foundPostId := false
		for _, param := range parameters {
			name, _ := param["name"].(string)
			if name == "id" {
				foundId = true
				if param["in"] != "path" {
					t.Error("parameter should be 'in: path'")
				}
				if param["required"] != true {
					t.Error("path parameter should be required")
				}
			}
			if name == "postId" {
				foundPostId = true
			}
		}

		if !foundId || !foundPostId {
			t.Error("expected to find both id and postId parameters")
		}
	})

	t.Run("error responses included", func(t *testing.T) {
		endpoint := ep.NewEndpointRes[string]().
			WithMethod("GET").
			WithPath("/test")

		op := buildOperation(endpoint, []*ep.Group{})

		responses, ok := op["responses"].(map[string]any)
		if !ok {
			t.Fatal("responses should be a map")
		}

		// Check standard error responses
		if _, ok := responses["400"]; !ok {
			t.Error("should have 400 error response")
		}
		if _, ok := responses["422"]; !ok {
			t.Error("should have 422 validation error response")
		}
		if _, ok := responses["500"]; !ok {
			t.Error("should have 500 error response")
		}
	})
}

func TestGenerateSchema(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		schema := generateSchema(nil)
		if schema["type"] != "null" {
			t.Errorf("expected type 'null', got '%v'", schema["type"])
		}
	})

	t.Run("struct type", func(t *testing.T) {
		type TestStruct struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		schema := generateSchema(TestStruct{})
		if schema["type"] != "object" {
			t.Errorf("expected type 'object', got '%v'", schema["type"])
		}

		properties, ok := schema["properties"].(map[string]any)
		if !ok {
			t.Fatal("properties should be a map")
		}

		if _, ok := properties["name"]; !ok {
			t.Error("should have 'name' property")
		}
		if _, ok := properties["age"]; !ok {
			t.Error("should have 'age' property")
		}
	})
}

func TestStructToSchema(t *testing.T) {
	t.Run("with openapi tag", func(t *testing.T) {
		type TestStruct struct {
			Description string `json:"description" openapi:"User description"`
		}

		schema := structToSchema(reflect.TypeOf(TestStruct{}))

		properties, ok := schema["properties"].(map[string]any)
		if !ok {
			t.Fatal("properties should be a map")
		}

		descSchema, ok := properties["description"].(map[string]any)
		if !ok {
			t.Fatal("description property should be a map")
		}

		if descSchema["description"] != "User description" {
			t.Errorf("expected description 'User description', got '%v'", descSchema["description"])
		}
	})

	t.Run("with omitempty tag", func(t *testing.T) {
		type TestStruct struct {
			Optional string `json:"optional,omitempty"`
			Required string `json:"required"`
		}

		schema := structToSchema(reflect.TypeOf(TestStruct{}))

		required, ok := schema["required"].([]string)
		if !ok {
			t.Fatal("required should be a slice")
		}

		foundOptional := false
		foundRequired := false
		for _, field := range required {
			if field == "optional" {
				foundOptional = true
			}
			if field == "required" {
				foundRequired = true
			}
		}

		if foundOptional {
			t.Error("optional field should not be in required")
		}
		if !foundRequired {
			t.Error("required field should be in required")
		}
	})

	t.Run("unexported fields ignored", func(t *testing.T) {
		type TestStruct struct {
			Exported   string `json:"exported"`
			unexported string
		}

		schema := structToSchema(reflect.TypeOf(TestStruct{}))

		properties, ok := schema["properties"].(map[string]any)
		if !ok {
			t.Fatal("properties should be a map")
		}

		if _, ok := properties["exported"]; !ok {
			t.Error("should have 'exported' property")
		}
		if _, ok := properties["unexported"]; ok {
			t.Error("should not have 'unexported' property")
		}
	})

	t.Run("json tag with hyphen", func(t *testing.T) {
		type TestStruct struct {
			Ignored string `json:"-"`
			Kept    string `json:"kept"`
		}

		schema := structToSchema(reflect.TypeOf(TestStruct{}))

		properties, ok := schema["properties"].(map[string]any)
		if !ok {
			t.Fatal("properties should be a map")
		}

		if _, ok := properties["Ignored"]; ok {
			t.Error("should not have 'Ignored' property (json:'-')")
		}
		if _, ok := properties["kept"]; !ok {
			t.Error("should have 'kept' property")
		}
	})
}

func TestAddMountedRoutesToSpec(t *testing.T) {
	t.Run("add routes with prefix", func(t *testing.T) {
		spec := map[string]any{
			"paths": map[string]any{},
		}

		routes := []schema.MountedRoute{
			{
				Method:  "GET",
				Path:    "/users",
				Summary: "List users",
			},
		}

		AddMountedRoutesToSpec(spec, "/api/v1", routes)

		paths := spec["paths"].(map[string]any)
		if _, ok := paths["/api/v1/users"]; !ok {
			t.Error("/api/v1/users should be in paths")
		}
	})

	t.Run("handles slash joining", func(t *testing.T) {
		spec := map[string]any{
			"paths": map[string]any{},
		}

		routes := []schema.MountedRoute{
			{
				Method: "GET",
				Path:   "/users",
			},
		}

		// Both ending/starting with slash
		AddMountedRoutesToSpec(spec, "/api/", routes)

		paths := spec["paths"].(map[string]any)
		if _, ok := paths["/api/users"]; !ok {
			t.Error("/api/users should be in paths (double slash handled)")
		}
	})

	t.Run("creates paths if missing", func(t *testing.T) {
		spec := map[string]any{}

		routes := []schema.MountedRoute{
			{
				Method: "GET",
				Path:   "/test",
			},
		}

		AddMountedRoutesToSpec(spec, "", routes)

		if _, ok := spec["paths"]; !ok {
			t.Error("paths should be created if missing")
		}
	})
}

func TestBuildMountedRouteOperation(t *testing.T) {
	t.Run("basic operation", func(t *testing.T) {
		route := schema.MountedRoute{
			Method:      "GET",
			Path:        "/users",
			Summary:     "List users",
			Description: "Get all users",
		}

		op := buildMountedRouteOperation(route)

		if op["summary"] != "List users" {
			t.Errorf("expected summary 'List users', got '%v'", op["summary"])
		}

		if op["description"] != "Get all users" {
			t.Errorf("expected description 'Get all users', got '%v'", op["description"])
		}
	})

	t.Run("with request type", func(t *testing.T) {
		type CreateUserReq struct {
			Name string `json:"name"`
		}

		route := schema.MountedRoute{
			Method:      "POST",
			Path:        "/users",
			RequestType: CreateUserReq{},
		}

		op := buildMountedRouteOperation(route)

		if _, ok := op["requestBody"]; !ok {
			t.Error("should have requestBody with request type")
		}
	})

	t.Run("with response type", func(t *testing.T) {
		type UserRes struct {
			ID int `json:"id"`
		}

		route := schema.MountedRoute{
			Method:       "GET",
			Path:         "/users/{id}",
			ResponseType: UserRes{},
		}

		op := buildMountedRouteOperation(route)

		responses, ok := op["responses"].(map[string]any)
		if !ok {
			t.Fatal("responses should be a map")
		}

		if _, ok := responses["200"]; !ok {
			t.Error("should have 200 response")
		}
	})

	t.Run("with path params", func(t *testing.T) {
		route := schema.MountedRoute{
			Method: "GET",
			Path:   "/users/{id}",
			PathParams: []schema.ParamInfo{
				{
					Name:        "id",
					Type:        "integer",
					Required:    true,
					Description: "User ID",
				},
			},
		}

		op := buildMountedRouteOperation(route)

		parameters, ok := op["parameters"].([]map[string]any)
		if !ok {
			t.Fatal("parameters should be a slice")
		}

		if len(parameters) != 1 {
			t.Errorf("expected 1 parameter, got %d", len(parameters))
		}

		if parameters[0]["name"] != "id" {
			t.Errorf("expected parameter name 'id', got '%v'", parameters[0]["name"])
		}

		if parameters[0]["schema"].(map[string]any)["type"] != "integer" {
			t.Error("parameter should have integer type")
		}
	})

	t.Run("auto-extract path params", func(t *testing.T) {
		route := schema.MountedRoute{
			Method: "GET",
			Path:   "/users/{userId}/posts/{postId}",
			// No PathParams defined - should auto-extract
		}

		op := buildMountedRouteOperation(route)

		parameters, ok := op["parameters"].([]map[string]any)
		if !ok {
			t.Fatal("parameters should be a slice")
		}

		if len(parameters) != 2 {
			t.Errorf("expected 2 auto-extracted parameters, got %d", len(parameters))
		}
	})

	t.Run("without response type", func(t *testing.T) {
		route := schema.MountedRoute{
			Method: "DELETE",
			Path:   "/users/{id}",
		}

		op := buildMountedRouteOperation(route)

		responses, ok := op["responses"].(map[string]any)
		if !ok {
			t.Fatal("responses should be a map")
		}

		if _, ok := responses["204"]; !ok {
			t.Error("should have 204 response when no response type")
		}
	})

	t.Run("default type for param", func(t *testing.T) {
		route := schema.MountedRoute{
			Method: "GET",
			Path:   "/users/{id}",
			PathParams: []schema.ParamInfo{
				{
					Name: "id",
					// Type is empty - should default to string
				},
			},
		}

		op := buildMountedRouteOperation(route)

		parameters := op["parameters"].([]map[string]any)
		schema := parameters[0]["schema"].(map[string]any)

		if schema["type"] != "string" {
			t.Errorf("expected default type 'string', got '%v'", schema["type"])
		}
	})
}

func TestExtractPathParameters(t *testing.T) {
	tests := []struct {
		path     string
		expected []string
	}{
		{"/users/{id}", []string{"id"}},
		{"/users/{userId}/posts/{postId}", []string{"userId", "postId"}},
		{"/health", []string{}},
		{"/users/{id}/profile", []string{"id"}},
		{"/{a}/{b}/{c}", []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := extractPathParameters(tt.path)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d params, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("param %d: expected '%s', got '%s'", i, expected, result[i])
				}
			}
		})
	}
}

func TestServeOpenAPI(t *testing.T) {
	spec := map[string]any{
		"openapi": "3.0.0",
		"info": map[string]any{
			"title":   "Test",
			"version": "1.0",
		},
	}

	rec := httptest.NewRecorder()
	ServeOpenAPI(rec, spec)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}

	body := rec.Body.String()
	if body == "" {
		t.Error("response body should not be empty")
	}

	// Check it's valid JSON
	if !strings.Contains(body, "openapi") {
		t.Error("response should contain 'openapi' key")
	}
}

func TestServeSwaggerUI(t *testing.T) {
	rec := httptest.NewRecorder()
	ServeSwaggerUI(rec, "/swagger")

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "text/html" {
		t.Errorf("expected Content-Type 'text/html', got '%s'", contentType)
	}

	body := rec.Body.String()
	if body == "" {
		t.Error("response body should not be empty")
	}

	// Check for swagger-ui content
	if !strings.Contains(body, "swagger-ui") {
		t.Error("response should contain 'swagger-ui'")
	}

	// Check that path is substituted correctly
	if !strings.Contains(body, "/swagger/openapi.json") {
		t.Error("response should contain the OpenAPI JSON path")
	}
}
