# RestKit - Type-Safe REST API Framework With OpenAPI Support

[![Go Version](https://img.shields.io/github/go-mod/go-version/RestStore/RestKit?style=flat-square&label=go)](https://golang.org/)
[![Tests](https://github.com/RestStore/RestKit/actions/workflows/go.yml/badge.svg)](https://github.com/RestStore/RestKit/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/reststore/restkit)](https://goreportcard.com/report/github.com/reststore/restkit)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/RestStore/RestKit/blob/main/LICENCE)
[![Maintained](https://img.shields.io/badge/Maintained%3F-yes-green.svg)](https://github.com/RestStore/RestKit/graphs/commit-activity)
[![GitHub stars](https://img.shields.io/github/stars/RestStore/RestKit?style=flat-square)](https://github.com/RestStore/RestKit/stargazers)

RestKit is a modern Go framework for building type-safe REST APIs with automatic OpenAPI documentation. It leverages generics to provide compile-time type safety while keeping the API simple and expressive.

## 🎯 Features

- **Type-Safe Endpoints** - Define request/response types with compile-time type checking using Go generics
- **Three Endpoint Patterns** - Full (Request + Response), Response-only, or Request-only endpoints
- **Automatic OpenAPI 3.0** - Generate OpenAPI specs directly from your endpoint definitions
- **Swagger UI Integration** - Built-in interactive API documentation
- **Flexible Middleware** - CORS, logging, panic recovery, plus custom middleware support
- **Request Binding & Validation** - JSON parsing, path parameters, struct validation with customization
- **Response Serialization** - Type-safe JSON responses with proper error handling
- **Endpoint Grouping** - Organize endpoints with common prefixes and shared middleware
- **Router Agnostic** - Built on standard net/http with adapters for Chi and other routers
- **Error Codes** - Standardized error responses with typed error codes

## ⚙️ Installation

RestKit requires **Go 1.26 or higher**. Install it with:

```bash
go get -u github.com/reststore/restkit
```

## ⚡️ Quickstart

Here's a complete example with a type-safe user creation endpoint:

```go
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	rest "github.com/reststore/restkit"
)

// Define your request and response types
type CreateUserReq struct {
	Name  string `json:"name" validate:"required,min=2"`
	Email string `json:"email" validate:"required,email"`
}

type UserRes struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Handler with compile-time type safety
func createUserHandler(ctx context.Context, req CreateUserReq) (UserRes, error) {
	return UserRes{
		ID:    1,
		Name:  req.Name,
		Email: req.Email,
	}, nil
}

func main() {
	// Create API with metadata
	api := rest.NewApi().
		WithVersion("1.0.0").
		WithTitle("User API").
		WithDescription("Manage users")

	// Define endpoint with fluent builder
	createUserEndpoint := rest.NewEndpoint[CreateUserReq, UserRes]().
		WithTitle("Create User").
		WithDescription("Create a new user").
		WithPath("/users").
		WithMethod("POST").
		WithHandler(createUserHandler)

	api.Add(createUserEndpoint)

	// Enable Swagger UI accessible at /swagger
	api.WithSwaggerUI()

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

	log.Println("Server running on http://localhost:8080")
	log.Println("Swagger UI: http://localhost:8080/swagger")
	log.Fatal(server.ListenAndServe())
}
```

Visit `http://localhost:8080/swagger` to see the interactive Swagger UI.

## 📖 Endpoint Types

RestKit provides three endpoint patterns, each with sensible defaults:

### Full Endpoint (Request + Response)

```go
endpoint := rest.NewEndpoint[RequestType, ResponseType]().
	WithPath("/resource").
	WithMethod("POST").
	WithHandler(func(ctx context.Context, req RequestType) (ResponseType, error) {
		// Your logic here
		return res, nil
	})
```

**Defaults:** POST method, JSON bind/write, auto-generated schemas

### Response-Only Endpoint

```go
endpoint := rest.NewEndpointRes[ResponseType]().
	WithPath("/resource/{id}").
	WithHandler(func(ctx context.Context) (ResponseType, error) {
		// Your logic here
		return res, nil
	})
```

**Defaults:** GET method, auto-generated response schema

### Request-Only Endpoint

```go
endpoint := rest.NewEndpointReq[RequestType]().
	WithPath("/resource/{id}").
	WithHandler(func(ctx context.Context, req RequestType) error {
		// Your logic here
		return nil
	})
```

**Defaults:** DELETE method, JSON bind, auto-generated request schema

## 👥 Endpoint Grouping

Organize related endpoints with a common prefix:

```go
userGroup := rest.NewGroup("/api/v1/users").
	WithTitle("Users").
	WithDescription("User management")

// Add endpoints to group
userGroup.WithEndpoints(
	getUser(),
	listUsers(),
	createUser(),
	updateUser(),
	deleteUser(),
)

api.AddGroup(userGroup)

// Routes become:
// GET /api/v1/users/{id}
// GET /api/v1/users
// POST /api/v1/users
// PUT /api/v1/users/{id}
// DELETE /api/v1/users/{id}
```

## 📝 Request Binding

By default, endpoints use JSON binding. Customize with:

```go
// JSON Binding (default)
endpoint.WithBind(rest.JSONBinder[RequestType]())

// Path Parameter Binding
endpoint.WithBind(rest.PathParamBinder[int](rest.StringToInt))

// Custom Binding
endpoint.WithBind(func(r *http.Request) (RequestType, error) {
	// Custom parsing logic
	return req, nil
})
```

## ✍️ Response Writing

Default JSON response writing can be customized:

```go
// JSON Response (default)
endpoint.WithWrite(rest.JSONWriter[ResponseType]())

// Custom Response
endpoint.WithWrite(func(w http.ResponseWriter, res ResponseType) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(res)
})
```

## ✔️ Validation

Add struct tag-based validation:

```go
type CreateUserReq struct {
	Name  string `json:"name" validate:"required,min=2,max=50"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"gte=18,lte=120"`
}

// Validation automatically applied on bind
// Validation errors return proper error responses
```

Validation uses the `go-playground/validator` library. See its docs for all available tags.

## 🔌 Middleware

RestKit includes common middleware and supports custom middleware:

### Built-in Middleware

```go
// HTTP request/response logging
api.WithMiddleware(rest.LoggingMiddleware())

// CORS with default or custom options
api.WithMiddleware(rest.CORSMiddleware())

// Panic recovery (converts panics to 500 responses)
api.WithMiddleware(rest.RecoveryMiddleware())
```

### Custom Middleware

```go
api.WithMiddleware(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Before handler
		log.Printf("Request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		// After handler
	})
})
```

Middleware runs in the order added and is applied globally to all endpoints.

## 📖 OpenAPI & Swagger UI

RestKit automatically generates OpenAPI 3.0 specs from your endpoints:

```go
api.WithSwaggerUI("/docs") // Enable Swagger UI at /docs
// or keep it empty for default /swagger
```

Access at:
- **Swagger UI**: `/docs`
- **OpenAPI JSON**: `/docs/openapi.json`

The spec is generated from:
- Endpoint titles and descriptions
- Auto-generated schemas from request/response types
- HTTP methods and paths
- Endpoint grouping (becomes API tags)

## 🛣️ Path Parameters

Access URL parameters in your handler:

```go
func getUser(ctx context.Context) (UserRes, error) {
	userID := rest.URLParam(ctx, "id")
	// Fetch and return user
}

// Define endpoint
getEndpoint := rest.NewEndpointRes[UserRes]().
	WithPath("/users/{id}").
	WithHandler(getUser)
```

Parameters are extracted from the path pattern and available in the context.

## ⚠️ Error Handling

RestKit provides standardized error responses with typed error codes:

```go
// Default error handler (JSON)
endpoint.WithErrorHandler(rest.JSONErrorWriter)

// Custom error handler
endpoint.WithErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	})
})
```

Error codes include:
- `ErrCodeBadRequest` - Malformed requests
- `ErrCodeValidation` - Validation failures
- `ErrCodeBind` - Request parsing errors
- `ErrCodeNotFound` - Resource not found
- `ErrCodeUnauthorized` - Authentication required
- `ErrCodeForbidden` - Access denied
- `ErrCodeInternal` - Server errors
- `ErrCodeMissingParam` - Missing path parameters
- `ErrCodeConfiguration` - Endpoint misconfiguration

## 🔍 Schema Generation

Schemas are automatically generated from your Go types for OpenAPI documentation:

```go
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Schema auto-generated for OpenAPI spec
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

## 🔀 Router Integration

RestKit is built on Go's standard net/http. Use adapters to integrate with other routers:

### Chi Router Adapter

```go
import restchi "github.com/reststore/restkit/adapters/chi"

// Register RestKit endpoints with Chi
chiRouter := chi.NewRouter()
restchi.RegisterRoutes(chiRouter, api)

// Start server with Chi router
http.ListenAndServe(":8080", chiRouter)
```

### Mount External Router

```go
// Mount external Chi router to RestKit API
restchi.Mount(api, "/", chiRouter,
	[]restkit.RouteMeta{
		{
			Method: "GET",
			Path:   "/users",
			Info: restkit.RouteInfo{
				Summary:      "List users",
				ResponseType: []User{},
			},
		},
	)
```

## 💡 Philosophy

RestKit brings simplicity to Go with compile-time type safety.
Taking inspiration from frameworks like FastEndpoints in .NET, RestKit provides a familiar, fluent API for defining REST endpoints while leveraging Go's strengths.

The framework is designed around:
- **Type Safety** - Catch errors at compile time, not runtime
- **Minimal Boilerplate** - Get productive quickly
- **Standard Library** - Built on Go's net/http, no external runtime dependencies
- **Developer Experience** - Familiar patterns from modern web frameworks
- **Performance** - Fast HTTP handling without reflection in hot paths

## 🚀 Benchmarks

RestKit keeps performance close to raw handlers while giving you type safety and automatic OpenAPI docs.
Here's how it stacks up:

### Real-World Performance

Testing against raw Chi and stdlib, RestKit stays competitive:

| What | RestKit | Raw Chi | Stdlib |
|------|---------|---------|--------|
| Simple GET | 138 µs | 136 µs | 136 µs |
| GET with params | 140 µs | 138 µs | 138 µs |
| POST with JSON | 195 µs | 165 µs | 171 µs |

For simple endpoints, we're within 1-2% of raw handlers. POST requests show a bit more overhead because RestKit is automatically validating and binding your request types, things you'd normally write by hand.

### Handler-Level Performance

At the handler level, the overhead is minimal:

```
Handler call                    6.5 ns
Handler with path params       6.8 µs
Route context creation         474 ns
```

Most of your time goes into HTTP overhead, network I/O, and your actual business logic, not the framework.

### Benchmark Yourself

Want to see the numbers on your machine? Run:

```bash
go test -bench=. -benchmem ./tests
```

Specific comparisons:

```bash
# Just RestKit endpoints
go test -bench=RestKit -benchmem ./tests

# Compare RestKit, Chi, and stdlib side-by-side
go test -bench='RestKit|Chi|Stdlib' -benchmem ./tests
```

## 👀 Examples

See the `examples/` directory for complete working examples:

```bash
# Basic example with groups and versions
go run ./examples/basic

# Chi router integration
go run ./examples/chi

# Mounting external Chi router
go run ./examples/chi_mount
```

Then visit `http://localhost:8080/docs` to explore the API.

## 💻 Development

To contribute to RestKit:

```bash
# Run tests
go test ./...

# Run formatting and linting
go fmt ./... && gofumpt -l -w . && golines -w -m 80 .
```

## 👍 Contributing

Contributions are welcome! Please:

1. Open an issue to discuss your changes
2. Fork the repository
3. Create a feature branch
4. Submit a pull request

## 📄 License

Licensed under the Apache License 2.0. See the [LICENSE](./LICENCE) file for details.

Copyright 2026 Robin Olsen
