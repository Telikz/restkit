package main

import (
	"context"

	rk "github.com/reststore/restkit"
	"github.com/reststore/restkit/examples/sqlc/db"

	_ "github.com/mattn/go-sqlite3"
	_ "github.com/reststore/restkit/validation/playground"
)

func userEndpoints() *rk.Group {
	return rk.NewGroup("/users").
		WithTitle("User Management").
		WithDescription("Endpoints for managing users").
		WithEndpoints(
			getUserEndpoint(),
			listUsersEndpoint(),
			createUserEndpoint(),
			updateUserEndpoint(),
			deleteUserEndpoint(),
			searchUsersEndpoint(),
		)
}

type CreateUserRequest struct {
	Name  string `json:"name"  validate:"required,min=2,max=32"`
	Email string `json:"email" validate:"required,email"`
}

type UpdateUserRequest struct {
	Name  string `json:"name"  validate:"omitempty,min=2,max=32"`
	Email string `json:"email" validate:"omitempty,email"`
}

func getUserEndpoint() *rk.Endpoint[rk.NoRequest, db.User] {
	return rk.GetEndpoint("/{id}", func(ctx context.Context, q *db.Queries, id int64) (db.User, error) {
		return q.GetUser(ctx, id)
	}).
		WithTitle("Get User").
		WithDescription("Get User Details")
}

func listUsersEndpoint() *rk.Endpoint[rk.NoRequest, []db.User] {
	return rk.ListEndpoint("/", func(ctx context.Context, q *db.Queries, limit, offset int32) ([]db.User, error) {
		return q.ListUsers(ctx, db.ListUsersParams{
			Limit:  int64(limit),
			Offset: int64(offset),
		})
	}).
		WithTitle("List Users").
		WithDescription("List users with pagination")
}

func createUserEndpoint() *rk.Endpoint[CreateUserRequest, db.User] {
	return rk.CreateEndpoint("/", func(ctx context.Context, q *db.Queries, req CreateUserRequest) (db.User, error) {
		return q.CreateUser(ctx, db.CreateUserParams{
			Name:  req.Name,
			Email: req.Email,
		})
	}).
		WithTitle("Create User").
		WithDescription("Create a new user")
}

func updateUserEndpoint() *rk.Endpoint[UpdateUserRequest, rk.NoResponse] {
	return rk.UpdateEndpoint("/{id}", func(ctx context.Context, q *db.Queries, id int64, req UpdateUserRequest) error {
		return q.UpdateUser(ctx, db.UpdateUserParams{
			Name:  req.Name,
			Email: req.Email,
			ID:    id,
		})
	}).
		WithTitle("Update User").
		WithDescription("Update user details")
}

func deleteUserEndpoint() *rk.Endpoint[rk.NoRequest, rk.MessageResponse] {
	return rk.DeleteEndpoint("/{id}", func(ctx context.Context, q *db.Queries, id int64) error {
		return q.DeleteUser(ctx, id)
	}).WithTitle("Delete User").WithDescription("Delete a user")
}

func searchUsersEndpoint() *rk.Endpoint[rk.NoRequest, []db.User] {
	return rk.SearchEndpoint("/search", func(ctx context.Context, q *db.Queries) ([]db.User, error) {
		return q.SearchUsers(ctx, db.SearchUsersParams{
			ID:        nilOrString(rk.URLQueryParam(ctx, "id")),
			Name:      nilOrString(rk.URLQueryParam(ctx, "name")),
			Email:     nilOrString(rk.URLQueryParam(ctx, "email")),
			CreatedAt: nilOrString(rk.URLQueryParam(ctx, "created_at")),
		})
	}).
		WithTitle("Search Users").
		WithDescription("Search users by query parameters").
		WithParameters(
			rk.Parameter{
				Name:        "id",
				Type:        "string",
				Description: "Filter by user ID",
				Location:    rk.ParamLocationQuery,
			},
			rk.Parameter{
				Name:        "name",
				Type:        "string",
				Description: "Filter by name (partial match)",
				Location:    rk.ParamLocationQuery,
			},
			rk.Parameter{
				Name:        "email",
				Type:        "string",
				Description: "Filter by email (partial match)",
				Location:    rk.ParamLocationQuery,
			},
			rk.Parameter{
				Name:        "created_at",
				Type:        "string",
				Description: "Filter by creation date",
				Location:    rk.ParamLocationQuery,
			},
		)
}

func nilOrString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
