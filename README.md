# RestKit- Type-Safe REST API Framework for Go

RestKit is a modern Go framework for building type-safe REST APIs with automatic OpenAPI documentation. It leverages Go generics to provide compile-time type safety while keeping the API simple and expressive.

## Features

- **Type-Safe Endpoints** - Define request/response types and get compile-time type checking
- **Generic Endpoint Builders** - Support for endpoints with request+response, response-only, or request-only
- **Automatic OpenAPI Generation** - Generate OpenAPI 3.0 specs from your endpoint definitions
- **Swagger UI Integration** - Built-in Swagger UI for exploring your API
- **Flexible Middleware** - CORS, logging, panic recovery, and custom middleware support
- **Request Binding & Validation** - JSON unmarshaling, path parameters, custom validators
- **Response Serialization** - JSON writing with proper error handling

## Installation

```bash
go get github.com/telikz/restkit
```

## Quick Start

### Basic API Setup

```go
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	rest "github.com/telikz/restkit"
)

func main() {
	// Create API instance
	api := rest.NewAPI()
	api.WithVersion("1.0.0")
	api.WithTitle("My API")
	api.WithDescription("A simple example API")

	// Add endpoints
	api.Add(createUserEndpoint())

	// Enable Swagger UI
	api.WithSwaggerUI(true).WithSwaggerUIPath("/docs")

	// Add middleware
	api.WithMiddleware(rest.LoggingMiddleware())
	api.WithMiddleware(rest.CORSMiddleware())

	// Start server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      api.Mux(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Println("Server running on :8080")
	log.Println("Swagger UI: http://localhost:8080/docs")
	log.Fatal(server.ListenAndServe())
}

// Request and Response types
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func createUserEndpoint() rest.Endpoint[CreateUserRequest, UserResponse] {
	return *rest.NewEndpoint[CreateUserRequest, UserResponse]().
		WithTitle("Create User").
		WithDescription("Create a new user").
		WithPath("/users").
		WithMethod("POST").
		WithHandler(func(ctx context.Context, req CreateUserRequest) (UserResponse, error) {
			// Your business logic here
			return UserResponse{
				ID:    1,
				Name:  req.Name,
				Email: req.Email,
			}, nil
		}).
		WithBind(rest.JSONBinder[CreateUserRequest]()).
		WithWrite(rest.JSONWriter[UserResponse]()).
		WithErrorHandler(rest.JSONErrorWriter).
		WithRequestSchema(rest.SchemaFrom[CreateUserRequest]()).
		WithResponseSchema(rest.SchemaFrom[UserResponse]())
}
```

## Endpoint Types

rest supports three endpoint variants:

### Full Endpoint (Request + Response)

```go
endpoint := rest.NewEndpoint[RequestType, ResponseType]()
// Default: POST method, auto-generated schemas
```

### Response-Only Endpoint

```go
endpoint := rest.NewEndpointRes[ResponseType]()
// Default: GET method
```

### Request-Only Endpoint

```go
endpoint := rest.NewEndpointReq[RequestType]()
// Default: DELETE method
```

## Grouping Endpoints

Group related endpoints with a common prefix:

```go
userGroup := rest.NewGroup("/api/v1/users").
	WithTitle("Users").
	WithDescription("User management endpoints").
	WithEndpoints(
		getUser(),
		listUsers(),
		createUser(),
		updateUser(),
		deleteUser(),
	)

api.AddGroup(userGroup)

// API routes become:
// GET /api/v1/users/{id}
// GET /api/v1/users
// POST /api/v1/users
// PUT /api/v1/users/{id}
// DELETE /api/v1/users/{id}
```

## Request Binding

### JSON Binding (Default)

```go
endpoint.WithBind(rest.JSONBinder[RequestType]())
```

### Path Parameter Binding

```go
endpoint.WithBind(rest.PathParamBinder[int](rest.StringToInt))
```

### Custom Binding

```go
endpoint.WithBind(func(r *http.Request) (RequestType, error) {
	// Custom logic
	return req, nil
})
```

## Response Writing

### JSON Response (Default)

```go
endpoint.WithWrite(rest.JSONWriter[ResponseType]())
```

### Custom Response

```go
endpoint.WithWrite(func(w http.ResponseWriter, res ResponseType) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(res)
})
```

## Validation

Add request validation:

```go
endpoint.WithValidation(func(req RequestType) error {
	if req.Name == "" {
		return errors.New("name is required")
	}
	return nil
})
```

## Middleware

### Built-in Middleware

```go
// Logging
api.WithMiddleware(rest.LoggingMiddleware())

// CORS
api.WithMiddleware(rest.CORSMiddleware())

// Panic Recovery
api.WithMiddleware(rest.RecoveryMiddleware())
```

### Custom Middleware

```go
api.WithMiddleware(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Before
		next.ServeHTTP(w, r)
		// After
	})
})
```

## OpenAPI & Swagger UI

rest automatically generates OpenAPI 3.0 specifications from your endpoints:

```go
api.WithSwaggerUI(true).WithSwaggerUIPath("/docs")
```

- **OpenAPI JSON** available at: `/docs/openapi.json`
- **Swagger UI** available at: `/docs`

The OpenAPI spec is generated from:

- Endpoint titles and descriptions
- Request/response schemas (auto-generated or manual)
- HTTP methods and paths
- Endpoint groups and tags

## Path Parameters

Access URL parameters in your handler:

```go
func getUserHandler(ctx context.Context, req GetUserRequest) (UserResponse, error) {
	userID := rest.URLParam(ctx, "id")
	// Use userID
	return user, nil
}
```

## Error Handling

### Default Error Handler

```go
endpoint.WithErrorHandler(rest.JSONErrorWriter)
```

### Custom Error Handler

```go
endpoint.WithErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	})
})
```

## Schema Generation

Schemas are auto-generated from your types:

```go
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Auto-generated schema:
endpoint.WithResponseSchema(rest.SchemaFrom[User]())

// Or override manually:
endpoint.WithResponseSchema(map[string]any{
	"type": "object",
	"properties": map[string]any{
		"id":    map[string]any{"type": "integer"},
		"name":  map[string]any{"type": "string"},
		"email": map[string]any{"type": "string"},
	},
	"required": []string{"id", "name", "email"},
})
```

## Examples

See the `example/` directory for a complete working example with multiple endpoints.

Run the example:

```bash
go run ./example
```

Then visit:

- API: http://localhost:8080
- Swagger UI: http://localhost:8080/docs
- OpenAPI JSON: http://localhost:8080/docs/openapi.json

## License
