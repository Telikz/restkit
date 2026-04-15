package endpoints

import (
	"context"
	"time"

	"github.com/google/uuid"
	rk "github.com/reststore/restkit"
)

// User represents a user in the system
type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// GetUserRequest defines the request path for getting a user by ID.
type GetUserRequest struct {
	ID uuid.UUID `path:"id"`
}

// GetUser defines an endpoint for getting a user by ID.
func GetUser() *rk.Endpoint[GetUserRequest, User] {
	return rk.Get("/users/{id}",
		func(ctx context.Context, req GetUserRequest) (User, error) {
			user, exists := users[req.ID]
			if !exists {
				return User{},
					rk.ErrNotFound.WithMessage("User not found")
			}
			return user, nil
		},
	)
}

// ListUsers defines an endpoint for listing all users.
func ListUsers() *rk.Endpoint[rk.NoRequest, []User] {
	return rk.List("/users",
		func(ctx context.Context, _ rk.NoRequest) ([]User, error) {
			userList := make([]User, 0, len(users))
			for _, u := range users {
				userList = append(userList, u)
			}
			return userList, nil
		},
	)
}

// CreateUserRequest represents the request body for creating a new user
type CreateUserRequest struct {
	Name  string `json:"name"  validate:"required,min=2,max=32"`
	Email string `json:"email" validate:"required,email"`
}

// CreateUser defines an endpoint for creating a new user.
func CreateUser() *rk.Endpoint[CreateUserRequest, User] {
	return rk.Post("/users",
		func(ctx context.Context, req CreateUserRequest) (User, error) {
			user := User{
				ID:        uuid.New(),
				Name:      req.Name,
				Email:     req.Email,
				CreatedAt: time.Now(),
			}
			nextID++
			users[user.ID] = user
			return user, nil
		},
	)
}

// UpdateUserRequest for partial updates
type UpdateUserRequest struct {
	ID    uuid.UUID `path:"id"`
	Name  string    `json:"name"  validate:"omitempty,min=2,max=32"`
	Email string    `json:"email" validate:"omitempty,email"`
}

// UpdateUserResponse represents the response body for updating a user
type UpdateUserResponse struct {
	Message string `json:"message"`
	User    User   `json:"user"`
}

// UpdateUser defines an endpoint for updating a user by ID.
func UpdateUser() *rk.Endpoint[UpdateUserRequest, UpdateUserResponse] {
	return rk.Patch("/users/{id}",
		func(ctx context.Context, req UpdateUserRequest) (UpdateUserResponse, error) {
			existing, exists := users[req.ID]
			if !exists {
				return UpdateUserResponse{},
					rk.ErrNotFound.WithMessage("User not found")
			}
			return UpdateUserResponse{
				Message: "User updated successfully",
				User:    rk.UpdateFields(User{}, existing, req),
			}, nil
		},
	)
}

// DeleteUser defines an endpoint for deleting a user by ID.
func DeleteUser() *rk.Endpoint[GetUserRequest, rk.NoResponse] {
	return rk.Delete("/users/{id}",
		func(ctx context.Context, req GetUserRequest) (rk.NoResponse, error) {
			if _, exists := users[req.ID]; !exists {
				return rk.NoResponse{},
					rk.ErrNotFound.WithMessage("User not found")
			}
			delete(users, req.ID)
			return rk.NoResponse{}, nil
		},
	)
}

// In-memory user store for demonstration purposes

var (
	nextID    = 3
	AliceUUID = uuid.MustParse("019d6c96-65c7-7422-876e-7dff72c62556")
	BobUUID   = uuid.MustParse("12345678-1234-1234-1234-123456789abc")
)

var users = map[uuid.UUID]User{
	AliceUUID: {
		ID:        AliceUUID,
		Name:      "Alice",
		Email:     "alice@example.com",
		CreatedAt: time.Now(),
	},
	BobUUID: {
		ID:        BobUUID,
		Name:      "Bob",
		Email:     "bob@example.com",
		CreatedAt: time.Now(),
	},
}
