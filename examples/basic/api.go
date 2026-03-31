package main

import (
	"log"
	"net/http"
	"time"

	rest "github.com/reststore/restkit"
	"github.com/reststore/restkit/examples/basic/endpoints"
	_ "github.com/reststore/restkit/validation/playground" // Opt-in validation
)

func main() {
	a := rest.NewApi()
	a.WithVersion("1.1")
	a.WithTitle("User API")
	a.WithDescription("RESTful API for managing users")

	// Add global middleware (applies to all endpoints)
	a.WithMiddleware(rest.NewCORS())
	a.WithMiddleware(rest.LoggingMiddleware())

	// Add individual endpoints
	a.AddEndpoint(endpoints.Ping())

	// Example of grouping endpoints under a common prefix
	a.AddGroup(rest.NewGroup("/api/v1").
		WithTitle("User Management v1").
		WithDescription("All user-related endpoints").
		WithEndpoints(
			endpoints.GetUser(),
			endpoints.ListUsers(),
			endpoints.CreateUser(),
			endpoints.UpdateUser(),
			endpoints.DeleteUser(),
		),
	)

	// Example of a second group with a different prefix
	a.AddGroup(rest.NewGroup("/api/v2").
		WithTitle("User Management v2").
		WithDescription("All user-related endpoints").
		WithEndpoints(
			endpoints.GetUser().WithPath("/get-user"),
			endpoints.ListUsers().WithPath("/list-users"),
			endpoints.CreateUser().WithPath("/create-user"),
			endpoints.UpdateUser().WithPath("/update-user"),
			endpoints.DeleteUser().WithPath("/delete-user"),
		),
	)

	a.WithSwaggerUI("/docs")

	// Configure and HTTP server with timeouts
	server := http.Server{
		Addr:         ":8080",
		Handler:      a.Mux(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start the server
	log.Println("Starting API server on :8080")
	log.Println("Swagger UI available at http://localhost:8080/docs")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
