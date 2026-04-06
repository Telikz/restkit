package main

import (
	"context"
	"time"

	"sqlc/db"

	rk "github.com/reststore/restkit"

	_ "github.com/mattn/go-sqlite3"
)

// UserResponse is a simplified version of db.User for API responses.
// API responses should not expose internal fields, so we define a separate struct.
type UserResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

// GetUserRequest defines the request path for getting a user by ID.
type GetUserRequest struct {
	ID int64 `path:"id"` // Set path:"id" to extract from URL path
}

func getUserEndpoint(q *db.Queries) *rk.Endpoint[GetUserRequest, UserResponse] {
	return rk.Get("/{id}",
		func(ctx context.Context, req GetUserRequest,
		) (UserResponse, error) {
			user, err := q.GetUser(ctx, req.ID)
			if err != nil {
				return UserResponse{}, err
			}
			return toUserResponse(user), nil
		},
	)
}

// ListUsersRequest defines query parameters for listing users with pagination.
type ListUsersRequest struct {
	Limit  int64 `query:"limit"  default:"20"` // Default limit is 20
	Offset int64 `query:"offset" default:"0"`  // Default offset is 0
}

// listUsersEndpoint defines an endpoint for listing users with pagination support.
func listUsersEndpoint(q *db.Queries) *rk.Endpoint[ListUsersRequest, []UserResponse] {
	return rk.List("/",
		func(ctx context.Context, req ListUsersRequest,
		) ([]UserResponse, error) {
			users, err := q.ListUsers(ctx, db.ListUsersParams{
				Limit:  req.Limit,
				Offset: req.Offset,
			})
			if err != nil {
				return nil, err
			}
			return mapUsersToResponse(users), nil
		},
	)
}

// CreateUserRequest defines the request body for creating a new user, through JSON binding.
// Validation tags are used by go-playground/validator to validate incoming requests.
type CreateUserRequest struct {
	Name  string `json:"name"  validate:"required,min=2,max=32"`
	Email string `json:"email" validate:"required,email"`
}

// createUserEndpoint defines an endpoint for creating a new user.
func createUserEndpoint(q *db.Queries) *rk.Endpoint[CreateUserRequest, UserResponse] {
	return rk.Create("/",
		func(ctx context.Context, req CreateUserRequest,
		) (UserResponse, error) {
			user, err := q.CreateUser(ctx, db.CreateUserParams{
				Name:  req.Name,
				Email: req.Email,
			})
			if err != nil {
				return UserResponse{}, err
			}
			return toUserResponse(user), nil
		},
	)
}

// UpdateUserRequest defines the request path and body for updating a user, through JSON binding.
// Validation tags are used by go-playground/validator to validate incoming requests.
type UpdateUserRequest struct {
	ID    int64  `path:"id"`
	Name  string `          json:"name"  validate:"omitempty,min=2,max=32"`
	Email string `          json:"email" validate:"omitempty,email"`
}

// updateUserEndpoint defines an endpoint for updating a user by ID.
func updateUserEndpoint(q *db.Queries) *rk.Endpoint[UpdateUserRequest, UserResponse] {
	return rk.Update("/{id}",
		func(ctx context.Context, req UpdateUserRequest,
		) (UserResponse, error) {
			user, err := q.UpdateUser(ctx, db.UpdateUserParams{
				ID:    req.ID,
				Name:  req.Name,
				Email: req.Email,
			})
			if err != nil {
				return UserResponse{}, err
			}
			return toUserResponse(user), nil
		},
	)
}

// DeleteUserRequest defines the request path for deleting a user.
type DeleteUserRequest struct {
	ID int64 `path:"id"`
}

// deleteUserEndpoint defines an endpoint for deleting a user by ID.
func deleteUserEndpoint(q *db.Queries) *rk.Endpoint[DeleteUserRequest, rk.NoResponse] {
	return rk.Delete("/{id}",
		func(ctx context.Context, req DeleteUserRequest,
		) error {
			return q.DeleteUser(ctx, req.ID)
		},
	)
}

// SearchUsersRequest defines query parameters for searching users with optional filters.
// All fields are pointers to allow distinguishing between "not provided" and "provided with zero value".
type SearchUserRequest struct {
	ID        *string `query:"id"`
	Name      *string `query:"name"`
	Email     *string `query:"email"`
	CreatedAt *string `query:"created_at"`
}

// SearchUsersEndpoint allows searching users with optional filters provided as query parameters.
func searchUsersEndpoint(q *db.Queries) *rk.Endpoint[SearchUserRequest, []UserResponse] {
	return rk.Search("/search",
		func(ctx context.Context, req SearchUserRequest,
		) ([]UserResponse, error) {
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
		},
	)
}

// toUserResponse converts a db.User to a UserResponse,
// formatting the CreatedAt field as an RFC3339 string.
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

// mapUsersToResponse converts a slice of db.User to a slice of UserResponse.
func mapUsersToResponse(users []db.User) []UserResponse {
	result := make([]UserResponse, len(users))
	for i, u := range users {
		result[i] = toUserResponse(u)
	}
	return result
}
