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

	api := &rest.Api{
		Version:     "1.0.0",
		Title:       "Example API",
		Description: "An example API using RestKit with Chi",
		Groups:      []*ep.Group{userGroup()},
		Endpoints:   []ep.Endpoint{pingEndpoint()},
	}
	api.WithSwaggerUI("/docs")

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
	return &ep.Group{
		Prefix:      "/users",
		Title:       "User Management",
		Description: "Endpoints for managing users",
		Endpoints: []ep.Endpoint{
			createUserEndpoint(),
			getUserEndpoint(),
			listUsersEndpoint(),
		},
	}
}

func createUserEndpoint() *rest.Endpoint[CreateUserRequest, UserResponse] {
	return &rest.Endpoint[CreateUserRequest, UserResponse]{
		Path:        "",
		Method:      "POST",
		Title:       "Create User",
		Description: "Create a new user with the provided name and email",
		Handler: func(ctx context.Context, req CreateUserRequest) (UserResponse, error) {
			return UserResponse{ID: 1, Name: req.Name, Email: req.Email}, nil
		},
	}
}

func getUserEndpoint() *rest.EndpointRes[UserResponse] {
	return &rest.EndpointRes[UserResponse]{
		Path:        "/{id}",
		Method:      "GET",
		Title:       "Get User",
		Description: "Retrieve details for a specific user by ID",
		Handler: func(ctx context.Context) (UserResponse, error) {
			return UserResponse{
				ID:    1,
				Name:  "John",
				Email: "john@example.com",
			}, nil
		},
	}
}

func listUsersEndpoint() *rest.EndpointRes[[]UserResponse] {
	return &rest.EndpointRes[[]UserResponse]{
		Path:        "",
		Method:      "GET",
		Title:       "List Users",
		Description: "Retrieve a list of all users",
		Handler: func(ctx context.Context) ([]UserResponse, error) {
			return []UserResponse{
				{ID: 1, Name: "John", Email: "john@example.com"},
			}, nil
		},
	}
}

func pingEndpoint() *rest.EndpointRes[MessageResponse] {
	return &rest.EndpointRes[MessageResponse]{
		Path:   "/ping",
		Method: "GET",
		Title:  "Ping Endpoint",
		Handler: func(ctx context.Context) (MessageResponse, error) {
			return MessageResponse{Message: "pong"}, nil
		},
	}
}

type MessageResponse struct {
	Message string `json:"message"`
}
