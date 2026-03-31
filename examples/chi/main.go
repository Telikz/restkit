package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	rest "github.com/reststore/restkit"
	restchi "github.com/reststore/restkit/adapters/chi"
	ep "github.com/reststore/restkit/internal/endpoints"
)

func main() {
	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	api := rest.NewApi().
		WithVersion("1.0.0").
		WithTitle("Example API").
		WithDescription("An example API using RestKit with Chi").
		AddGroup(userGroup()).
		AddEndpoint(pingEndpoint()).
		WithSwaggerUI("/docs")

	restchi.RegisterRoutes(r, api)

	log.Println("Server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

type CreateUserRequest struct {
	Name  string `json:"name"  validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func userGroup() *ep.Group {
	return ep.NewGroup("/users").
		WithTitle("User Management").
		WithDescription("Endpoints for managing users").
		WithEndpoints(
			createUserEndpoint(),
			getUserEndpoint(),
			listUsersEndpoint(),
		)
}

func createUserEndpoint() *rest.Endpoint[CreateUserRequest, UserResponse] {
	return rest.NewEndpoint[CreateUserRequest, UserResponse]().
		WithPath("/").
		WithMethod("POST").
		WithTitle("Create User").
		WithDescription("Create a new user with the provided name and email").
		WithHandler(func(ctx context.Context, req CreateUserRequest) (UserResponse, error) {
			return UserResponse{ID: 1, Name: req.Name, Email: req.Email}, nil
		})
}

func getUserEndpoint() *rest.Endpoint[rest.NoRequest, UserResponse] {
	return rest.NewEndpointRes[UserResponse]().
		WithPath("/{id}").
		WithMethod("GET").
		WithTitle("Get User").
		WithDescription("Retrieve details for a specific user by ID").
		WithHandler(func(ctx context.Context, _ rest.NoRequest) (UserResponse, error) {
			return UserResponse{
				ID:    1,
				Name:  "John",
				Email: "john@example.com",
			}, nil
		})
}

func listUsersEndpoint() *rest.Endpoint[rest.NoRequest, []UserResponse] {
	return rest.NewEndpointRes[[]UserResponse]().
		WithPath("/").
		WithMethod("GET").
		WithTitle("List Users").
		WithDescription("Retrieve a list of all users").
		WithHandler(func(ctx context.Context, _ rest.NoRequest) ([]UserResponse, error) {
			return []UserResponse{
				{ID: 1, Name: "John", Email: "john@example.com"},
			}, nil
		})
}

func pingEndpoint() *rest.Endpoint[rest.NoRequest, MessageResponse] {
	return rest.NewEndpointRes[MessageResponse]().
		WithPath("/ping").
		WithMethod("GET").
		WithTitle("Ping Endpoint").
		WithHandler(func(ctx context.Context, _ rest.NoRequest) (MessageResponse, error) {
			return MessageResponse{Message: "pong"}, nil
		})
}

type MessageResponse struct {
	Message string `json:"message"`
}
