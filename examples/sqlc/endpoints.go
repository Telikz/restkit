package main

import (
	"context"
	"time"

	"sqlc/db"

	rk "github.com/reststore/restkit"

	_ "github.com/mattn/go-sqlite3"
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

// Response DTO - never return raw db models directly
type UserResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

// CreateUserRequest body validated via struct tags
type CreateUserRequest struct {
	Name  string `json:"name"  validate:"required,min=2,max=32"`
	Email string `json:"email" validate:"required,email"`
}

// UpdateUserRequest combines path param (ID) with body fields
type UpdateUserRequest struct {
	ID    int64  `path:"id"`
	Name  string `          json:"name"  validate:"omitempty,min=2,max=32"`
	Email string `          json:"email" validate:"omitempty,email"`
}

// SearchUsersRequest uses pointer fields for optional filters
type SearchUsersRequest struct {
	ID        *string `query:"id"`
	Name      *string `query:"name"`
	Email     *string `query:"email"`
	CreatedAt *string `query:"created_at"`
}

// ListUsersRequest for pagination
type ListUsersRequest struct {
	Limit  int32 `query:"limit"  default:"20"`
	Offset int32 `query:"offset" default:"0"`
}

func getUserEndpoint() *rk.Endpoint[rk.GetRequest, UserResponse] {
	return rk.GetEndpoint("/{id}", func(ctx context.Context, q *db.Queries, req rk.GetRequest) (UserResponse, error) {
		user, err := q.GetUser(ctx, req.ID)
		if err != nil {
			return UserResponse{}, err
		}
		return toUserResponse(user), nil
	}).
		WithTitle("Get User").
		WithDescription("Get user by ID")
}

func listUsersEndpoint() *rk.Endpoint[ListUsersRequest, []UserResponse] {
	return rk.ListEndpoint("/", func(ctx context.Context, q *db.Queries, req ListUsersRequest) ([]UserResponse, error) {
		users, err := q.ListUsers(ctx, db.ListUsersParams{
			Limit:  int64(req.Limit),
			Offset: int64(req.Offset),
		})
		if err != nil {
			return nil, err
		}
		return mapUsersToResponse(users), nil
	}).
		WithTitle("List Users").
		WithDescription("List users with pagination")
}

func createUserEndpoint() *rk.Endpoint[CreateUserRequest, UserResponse] {
	return rk.CreateEndpoint("/", func(ctx context.Context, q *db.Queries, req CreateUserRequest) (UserResponse, error) {
		user, err := q.CreateUser(ctx, db.CreateUserParams{
			Name:  req.Name,
			Email: req.Email,
		})
		if err != nil {
			return UserResponse{}, err
		}
		return toUserResponse(user), nil
	}).
		WithTitle("Create User").
		WithDescription("Create a new user")
}

func updateUserEndpoint() *rk.Endpoint[UpdateUserRequest, rk.NoResponse] {
	return rk.UpdateEndpoint("/{id}", func(ctx context.Context, q *db.Queries, req UpdateUserRequest) error {
		return q.UpdateUser(ctx, db.UpdateUserParams{
			ID:    req.ID,
			Name:  req.Name,
			Email: req.Email,
		})
	}).
		WithTitle("Update User").
		WithDescription("Update user details")
}

func deleteUserEndpoint() *rk.Endpoint[rk.DeleteRequest, rk.MessageResponse] {
	return rk.DeleteEndpoint("/{id}", func(ctx context.Context, q *db.Queries, req rk.DeleteRequest) error {
		return q.DeleteUser(ctx, req.ID)
	}).
		WithTitle("Delete User").
		WithDescription("Delete a user")
}

func searchUsersEndpoint() *rk.Endpoint[SearchUsersRequest, []UserResponse] {
	return rk.SearchEndpoint("/search", func(ctx context.Context, q *db.Queries, req SearchUsersRequest) ([]UserResponse, error) {
		users, err := q.SearchUsers(ctx, db.SearchUsersParams{
			ID:        req.ID,
			Name:      req.Name,
			Email:     req.Email,
			CreatedAt: req.CreatedAt,
		})
		if err != nil {
			return nil, err
		}
		return mapUsersToResponse(users), nil
	}).
		WithTitle("Search Users").
		WithDescription("Search users by query parameters")
}

func toUserResponse(u db.User) UserResponse {
	createdAt := ""
	if u.CreatedAt.Valid {
		createdAt = u.CreatedAt.Time.Format(time.RFC3339)
	}
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: createdAt,
	}
}

func mapUsersToResponse(users []db.User) []UserResponse {
	result := make([]UserResponse, len(users))
	for i, u := range users {
		result[i] = toUserResponse(u)
	}
	return result
}
