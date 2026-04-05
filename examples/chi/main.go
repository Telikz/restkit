package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	rk "github.com/reststore/restkit"
	restchi "github.com/reststore/restkit/adapters/chi"
)

func main() {
	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	api := rk.NewApi().
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

func userGroup() *rk.Group {
	return rk.NewGroup("/users").
		WithTitle("User Management").
		WithDescription("Endpoints for managing users").
		WithEndpoints(
			createUserEndpoint(),
			getUserEndpoint(),
			listUsersEndpoint(),
		)
}

func createUserEndpoint() *rk.Endpoint[CreateUserRequest, UserResponse] {
	return rk.NewEndpoint[CreateUserRequest, UserResponse]().
		WithPath("/").
		WithMethod("POST").
		WithTitle("Create User").
		WithDescription("Create a new user with the provided name and email").
		WithHandler(func(ctx context.Context, req CreateUserRequest) (UserResponse, error) {
			return UserResponse{ID: 1, Name: req.Name, Email: req.Email}, nil
		})
}

func getUserEndpoint() *rk.Endpoint[rk.NoRequest, UserResponse] {
	return rk.NewEndpoint[rk.NoRequest, UserResponse]().
		WithPath("/{id}").
		WithMethod("GET").
		WithTitle("Get User").
		WithDescription("Retrieve details for a specific user by ID").
		WithHandler(func(ctx context.Context, _ rk.NoRequest) (UserResponse, error) {
			return UserResponse{
				ID:    1,
				Name:  "John",
				Email: "john@example.com",
			}, nil
		})
}

func listUsersEndpoint() *rk.Endpoint[rk.NoRequest, []UserResponse] {
	return rk.NewEndpoint[rk.NoRequest, []UserResponse]().
		WithPath("/").
		WithMethod("GET").
		WithTitle("List Users").
		WithDescription("Retrieve a list of all users").
		WithHandler(func(ctx context.Context, _ rk.NoRequest) ([]UserResponse, error) {
			return []UserResponse{
				{ID: 1, Name: "John", Email: "john@example.com"},
			}, nil
		})
}

func pingEndpoint() *rk.Endpoint[rk.NoRequest, MessageResponse] {
	return rk.NewEndpoint[rk.NoRequest, MessageResponse]().
		WithPath("/ping").
		WithMethod("GET").
		WithTitle("Ping Endpoint").
		WithHandler(func(ctx context.Context, _ rk.NoRequest) (MessageResponse, error) {
			return MessageResponse{Message: "pong"}, nil
		})
}

type MessageResponse struct {
	Message string `json:"message"`
}
