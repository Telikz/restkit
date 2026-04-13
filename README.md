# RestKit

[![Go Version](https://img.shields.io/github/go-mod/go-version/RestStore/RestKit?style=flat-square&label=go)](https://golang.org/)
[![Tests](https://github.com/RestStore/RestKit/actions/workflows/go.yml/badge.svg)](https://github.com/RestStore/RestKit/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/reststore/restkit)](https://goreportcard.com/report/github.com/reststore/restkit)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/RestStore/RestKit/blob/main/LICENCE)

**Type-safe REST APIs with automatic OpenAPI documentation**

RestKit brings compile-time type safety to REST APIs using Go generics, while automatically generating OpenAPI 3.0 specs from your code. Write handlers with typed requests and responses—no reflection in hot paths, no manual schema writing.

## Why RestKit?

- **Type safety without boilerplate** - Generic endpoints catch errors at compile time
- **Auto-generated OpenAPI** - Swagger UI and schemas from your Go types
- **Minimal overhead** - Microsecond-level framework cost ([benchmarks](#performance))
- **Database integration** - Built-in sqlc support with context injection and auto-transactions
- **Flexible middleware** - Apply at global, group, or endpoint level
- **Progressive enhancement** - Mount existing routers, migrate incrementally
- **Modern protocols** - SSE, WebSocket, gRPC, HTTP/3 support built-in
- **Router agnostic** - Adapters for Chi, Echo, Gin, or use stdlib

## Install

```bash
go get github.com/reststore/restkit
```

Requires Go 1.26+

## Quickstart

```go
package main

import (
	"context"
	"log"
	"net/http"

	rk "github.com/reststore/restkit"
)

type CreateUserReq struct {
	Name  string `json:"name" validate:"required,min=2"`
	Email string `json:"email" validate:"required,email"`
}

type UserRes struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	api := rk.NewApi()
	api.WithSwaggerUI()
	api.WithVersion("1.0.0")
	api.WithTitle("User API")

	// Type-safe endpoint - compiler ensures handler matches types
	api.Add(rk.Post("/users",
		func(ctx context.Context, req CreateUserReq) (UserRes, error) {
			// Auto-parsed JSON, validated, type-safe
			return UserRes{ID: 1, Name: req.Name, Email: req.Email}, nil
		}),
	)

	log.Println("Server: http://localhost:8080")
	log.Println("Swagger: http://localhost:8080/swagger")
	http.ListenAndServe(":8080", api.Mux())
}
```

That's it. You now have:
- Type-safe request/response handling
- Automatic OpenAPI 3.0 spec at `/swagger/openapi.json`
- Interactive Swagger UI at `/swagger`
- Runtime validation (add `api.WithValidator(playground.NewValidator())`)

## Core Concepts

### Endpoint Helpers

RestKit provides CRUD helpers with smart defaults:

```go
// GET /users/{id} - path param auto-extracted
rk.Get("/users/{id}",
	func(ctx context.Context, req rk.GetRequest) (User, error) {
		return db.GetUser(ctx, req.ID) // req.ID from path
	},
)

// GET /users?limit=20&offset=0 - query params with defaults
type ListReq struct {
	Limit  int32 `query:"limit" default:"20"`
	Offset int32 `query:"offset" default:"0"`
}
rk.List("/users",
	func(ctx context.Context, req ListReq) ([]User, error) {
		return db.ListUsers(ctx, req.Limit, req.Offset)
	},
)

// POST /users - JSON body
rk.Post("/users",
	func(ctx context.Context, req CreateUserReq) (User, error) {
		return db.CreateUser(ctx, req)
	},
)

// PATCH /users/{id} - path + JSON body
type UpdateReq struct {
	ID   int64  `path:"id"`
	Name string `json:"name"`
}
rk.Patch("/users/{id}",
	func(ctx context.Context, req UpdateReq) error {
		return db.UpdateUser(ctx, req.ID, req.Name)
	},
)

// DELETE /users/{id}
rk.Delete("/users/{id}",
	func(ctx context.Context, req rk.DeleteRequest) error {
		return db.DeleteUser(ctx, req.ID)
	},
)
```
All helpers use `NewEndpoint[Req, Res]()` under the hood with method-specific defaults.

### Struct Tags

Bind URL params, query strings, and JSON bodies using struct tags:

```go
type UpdateUserReq struct {
	ID       int64   `path:"id"`                      // from /users/{id}
	Name     string  `json:"name"`                    // from JSON body
	Active   *bool   `query:"active"`                 // optional query param
	PageSize int     `query:"page_size" default:"20"` // with default
}
```

Works seamlessly with sqlc - use `*string`, `*int64` for nullable types.

### Grouping & Organization

```go
users := rk.NewGroup("/api/v1/users").
	WithTitle("Users").
	WithEndpoints(
		rk.List("/", listUsers),
		rk.Post("/", createUser),
		rk.Get("/{id}", getUser),
		rk.Patch("/{id}", updateUser),
		rk.Delete("/{id}", deleteUser),
	)

api.AddGroup(users)
// Routes: GET/POST /api/v1/users, GET/PATCH/DELETE /api/v1/users/{id}
```

### Validation

Opt-in validation via go-playground/validator:

```go
import "github.com/reststore/restkit/validators/playground"

api.WithValidator(playground.NewValidator())

type CreateUserReq struct {
	Name  string `json:"name" validate:"required,min=2,max=50"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"gte=18,lte=120"`
}
```

### Middleware

Apply middleware at three levels:

```go
// Global - applies to all endpoints (supports chaining)
api.WithMiddleware(
	rk.LoggingMiddleware(),
	rk.RecoveryMiddleware(),
	rk.CORSMiddleware(),
)

// Group - applies to all endpoints in the group
users := rk.NewGroup("/users").
	WithMiddleware(authMiddleware, rateLimitMiddleware).
	WithEndpoints(...)

// Endpoint - applies to a single endpoint
rk.Post("/users", createUser).
	WithMiddleware(cacheMiddleware, validationMiddleware)
```

**Built-in middleware:**

```go
// CORS with configurable options
api.WithMiddleware(rk.CORSMiddleware(
	rk.CORSOptions.Origins("https://example.com"),
	rk.CORSOptions.Methods("GET", "POST"),
	rk.CORSOptions.Credentials(),
))

// Security headers (CSP, HSTS, X-Frame-Options, etc.)
api.WithMiddleware(rk.SecurityHeaderMiddleware(
	rk.SecurityHeadersOptions.CSP("default-src 'self'"),
	rk.SecurityHeadersOptions.HSTS("max-age=31536000"),
))

// Request ID injection and propagation
api.WithMiddleware(rk.RequestIDMiddleware(
	rk.RequestIDOptions.Header("X-Request-ID"),
))

// Access request ID in handlers
func handler(ctx context.Context, req Req) (Res, error) {
	requestID := rk.RequestIDFromContext(ctx)
	// ...
}

// Custom middleware
api.WithMiddleware(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// before
		next.ServeHTTP(w, r)
		// after
	})
})
```

### Database Integration

**Inject database queries into context** - works seamlessly with sqlc:

```go
import "your-project/db"

queries := db.New(database)

// Inject queries globally
api.WithMiddleware(rk.DBMiddleware(queries))

// Access in handlers
func getUser(ctx context.Context, req rk.GetRequest) (User, error) {
	q := rk.Queries(ctx).(*db.Queries)
	return q.GetUser(ctx, req.ID)
}
```

**Automatic transactions** - commits on success (2xx), rolls back on error:

```go
api.WithMiddleware(rk.TransactionMiddleware(
	database,
	db.New,        // creates queries from *sql.DB
	db.WithTx,     // wraps queries with transaction
))

// Now every request runs in a transaction
func createUser(ctx context.Context, req CreateUserReq) (User, error) {
	q := rk.Queries(ctx).(*db.Queries)
	// Automatically committed if no error, rolled back otherwise
	return q.CreateUser(ctx, db.CreateUserParams{
		Name:  req.Name,
		Email: req.Email,
	})
}
```

See `examples/sqlc` for a complete working example.

## Advanced Features

### Server-Sent Events (SSE)

Stream real-time data to clients:

```go
rk.Stream("/events/{id}",
 func(ctx context.Context, req EventReq) (<-chan rk.Event[Data], error) {
	stream := make(chan rk.Event[Data])
	go func() {
		defer close(stream)
		for i := range 10 {
			stream <- rk.Event[Data]{
				ID:    fmt.Sprintf("%d", i),
				Event: "message",
				Data:  Data{Message: fmt.Sprintf("Event %d", i)},
			}
			time.Sleep(time.Second)
		}
	}()
	return stream, nil
})
```

### WebSocket Support

Real-time bidirectional communication:

```go
import "github.com/reststore/restkit/extra/websocket"

websocket.New("/ws/{room}",
 func(ctx context.Context, req WsReq, conn *websocket.Conn) error {
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil { return nil }
		conn.WriteMessage(msgType, []byte("Echo: " + string(msg)))
	}
})
```

### gRPC Gateway

Expose gRPC services as REST endpoints:

```go
import rkgrpc "github.com/reststore/restkit/extra/grpc"

rkgrpc.GRPC("/hello", grpcClient,
 func(ctx context.Context, c pb.GreeterClient, req *pb.HelloRequest,
 ) (*pb.HelloReply, error) {
	return c.SayHello(ctx, req)
})
```

### HTTP/3 Support

Run HTTP/2 and HTTP/3 simultaneously:

```go
import "github.com/reststore/restkit/extra/http3"

http3.Serve(api, ":8080", ":8081", "cert.pem", "key.pem")
// HTTP/2 on :8080 (TCP), HTTP/3 on :8081 (UDP)
```

### Custom Serializers

YAML, pretty JSON, or custom formats:

```go
// YAML responses
import rkyml "github.com/reststore/restkit/serializers/yaml"
api.WithSerializer(rkyml.Serializer())
api.WithDeserializer(rkyml.Deserializer())

// Pretty JSON (indented)
api.WithSerializer(rk.Serializers.JSONPretty())

// Custom
api.WithSerializer(func(w http.ResponseWriter, data any) error {
	// your serialization logic
})
```

## OpenAPI & Swagger

Automatic OpenAPI 3.0 generation from your Go types:

```go
api.WithSwaggerUI()  // default: /swagger

// Multiple server URLs
api.WithServer("https://api.prod.com", "Production", nil)
api.WithServer("https://api.staging.com", "Staging", nil)

// Export spec to file
rk.GenerateOpenAPIFile("docs/openapi.json", api.GenerateOpenAPI())
```

Access Swagger UI at `/swagger`. Schemas auto-generated from struct types. Groups become OpenAPI tags.

## Router Adapters

### Use Your Existing Router

Built on stdlib `net/http`, with adapters for popular routers:

```go
// Chi
import restchi "github.com/reststore/restkit/adapters/chi"
router := chi.NewRouter()
restchi.RegisterRoutes(router, api)
http.ListenAndServe(":8080", router)

// Echo
import restecho "github.com/reststore/restkit/adapters/echo"
e := echo.New()
restecho.RegisterRoutes(e, api)
e.Start(":8080")

// Gin
import restgin "github.com/reststore/restkit/adapters/gin"
router := gin.Default()
restgin.RegisterRoutes(router, api)
router.Run(":8080")
```

Or use stdlib directly: `http.ListenAndServe(":8080", api.Mux())`

### Progressive Enhancement

**Mount existing routers into RestKit** - add type safety and OpenAPI docs without rewriting:

```go
// Your existing Chi router with legacy endpoints
legacyRouter := chi.NewRouter()
legacyRouter.Get("/users", oldHandler)
legacyRouter.Post("/users", oldCreateHandler)

// Create RestKit API
api := rk.NewApi().WithSwaggerUI()

// Mount the legacy router with metadata for OpenAPI docs
restchi.Mount(api, "/api/v1", legacyRouter, []rk.RouteMeta{
	{
		Method: "GET",
		Path:   "/users",
		Info:   rk.RouteInfo{
			Summary: "List users",
			ResponseType: []User{}},
	},
	{
		Method: "POST",
		Path:   "/users",
		Info:   rk.RouteInfo{
			Summary:      "Create user",
			RequestType:  CreateUserReq{},
			ResponseType: User{},
		},
	},
})

// Add new RestKit endpoints alongside legacy routes
api.AddGroup(rk.NewGroup("/api/v2/users").WithEndpoints(
	rk.Get("/{id}", getUser),
	rk.Post("/", createUser),
))

// Both legacy and RestKit routes work, all in Swagger
http.ListenAndServe(":8080", api.Mux())
```

**Benefits:**
- ✅ Add OpenAPI docs to existing routes without changes
- ✅ Enable validation on legacy endpoints (via `RequestType` metadata)
- ✅ Enhance route-by-route at your own pace
- ✅ Run legacy and modern endpoints side-by-side

## Performance

Minimal overhead - type safety without sacrificing speed:

```
Simple GET request:     1.5 µs   (1605 B/op, 17 allocs/op)
GET with path params:   2.1 µs   (2552 B/op, 21 allocs/op)
POST with JSON:         3.3 µs   (6359 B/op, 24 allocs/op)
Handler call overhead:  2.2 ns   (0 allocs)
```

The overhead is negligible compared to network I/O and database queries.
You get type safety, auto-validation, and OpenAPI generation with performance close to hand-written handlers.

Run benchmarks yourself:
```bash
go test -bench=. -benchmem ./tests
```

## Examples

Complete working examples in `examples/`:

```bash
go run ./examples/basic       # groups, validation, CRUD
go run ./examples/sqlc        # sqlc integration with RestKit
go run ./examples/stream      # Server-Sent Events (SSE)
go run ./examples/websocket   # WebSocket endpoints
go run ./examples/grpc        # gRPC gateway
go run ./examples/http3       # HTTP/3 support
go run ./examples/yaml        # YAML serialization
go run ./examples/serializer  # custom serializers/deserializers
go run ./examples/chi         # Chi router adapter
go run ./examples/echo        # Echo router adapter
go run ./examples/gin         # Gin router adapter
go run ./examples/stdlib      # Standard library router
```

Visit `http://localhost:8080/swagger` to explore the API.

## Development

```bash
just fmt     # format (gofumpt + golines, required before commit)
just tidy    # tidy all modules
go test ./...  # run tests
```

## Contributing

1. Open an issue to discuss changes
2. Fork and create a feature branch
3. Run `just fmt` before committing
4. Submit a pull request

## 📄 License

Licensed under the Apache License 2.0. See the [LICENSE](./LICENCE) file for details.