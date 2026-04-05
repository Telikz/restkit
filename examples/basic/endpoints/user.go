package endpoints

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	rk "github.com/reststore/restkit"
)

// User represents a user in the system
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateUserRequest represents the request body for creating a new user
type CreateUserRequest struct {
	Name  string `json:"name"  validate:"required,min=2,max=32"`
	Email string `json:"email" validate:"required,email"`
}

// UpdateUserRequest for partial updates
type UpdateUserRequest struct {
	Name  string `json:"name"  validate:"omitempty,min=2,max=32"`
	Email string `json:"email" validate:"omitempty,email"`
}

// MessageResponse is a generic response for operations that return a message
type MessageResponse struct {
	Message string `json:"message"`
}

func GetUser() *rk.Endpoint[rk.NoRequest, User] {
	return rk.NewEndpoint[rk.NoRequest, User]().
		WithMethod(http.MethodGet).
		WithPath("/users/{id}").
		WithTitle("Get User").
		WithDescription("Get a user by ID").
		WithHandler(func(ctx context.Context, _ rk.NoRequest) (User, error) {
			return getUserHandler(ctx)
		})
}

func ListUsers() *rk.Endpoint[rk.NoRequest, []User] {
	return rk.NewEndpoint[rk.NoRequest, []User]().
		WithMethod(http.MethodGet).
		WithPath("/users").
		WithTitle("List Users").
		WithDescription("Get a list of all users").
		WithHandler(func(ctx context.Context, _ rk.NoRequest) ([]User, error) {
			return listUsersHandler(ctx)
		})
}

func CreateUser() *rk.Endpoint[CreateUserRequest, *User] {
	return rk.NewEndpoint[CreateUserRequest, *User]().
		WithMethod(http.MethodPost).
		WithPath("/users").
		WithTitle("Create User").
		WithDescription("Create a new user").
		WithHandler(createUserHandler)
}

func UpdateUser() *rk.Endpoint[UpdateUserRequest, rk.NoResponse] {
	return rk.NewEndpoint[UpdateUserRequest, rk.NoResponse]().
		WithMethod(http.MethodPatch).
		WithPath("/users/{id}").
		WithTitle("Update User").
		WithDescription("Update a user by ID").
		WithHandler(func(ctx context.Context, req UpdateUserRequest) (rk.NoResponse, error) {
			return rk.NoResponse{}, updateUserHandler(ctx, req)
		})
}

func DeleteUser() *rk.Endpoint[rk.NoRequest, MessageResponse] {
	return rk.NewEndpoint[rk.NoRequest, MessageResponse]().
		WithMethod(http.MethodDelete).
		WithPath("/users/{id}").
		WithTitle("Delete User").
		WithHandler(func(ctx context.Context, _ rk.NoRequest) (MessageResponse, error) {
			return deleteUserHandler(ctx)
		}).
		WithDescription("Delete a user by ID")
}

var nextID = 3

var users = map[int]User{
	1: {
		ID:        1,
		Name:      "Alice",
		Email:     "alice@example.com",
		CreatedAt: time.Now(),
	},
	2: {ID: 2, Name: "Bob", Email: "bob@example.com", CreatedAt: time.Now()},
}

func getUserHandler(ctx context.Context) (User, error) {
	idStr := rk.URLParam(ctx, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return User{}, errors.New("invalid id")
	}

	user, exists := users[id]
	if !exists {
		return User{}, errors.New("user not found")
	}
	return user, nil
}

func listUsersHandler(ctx context.Context) ([]User, error) {
	userList := make([]User, 0, len(users))
	for _, u := range users {
		userList = append(userList, u)
	}
	return userList, nil
}

func createUserHandler(
	ctx context.Context,
	req CreateUserRequest,
) (*User, error) {
	user := User{
		ID:        nextID,
		Name:      req.Name,
		Email:     req.Email,
		CreatedAt: time.Now(),
	}
	nextID++
	users[user.ID] = user
	return &user, nil
}

func updateUserHandler(ctx context.Context, req UpdateUserRequest) error {
	idStr := rk.URLParam(ctx, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return errors.New("invalid id")
	}

	existing, exists := users[id]
	if !exists {
		return errors.New("user not found")
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Email != "" {
		existing.Email = req.Email
	}

	users[id] = existing
	return nil
}

func deleteUserHandler(ctx context.Context) (MessageResponse, error) {
	idStr := rk.URLParam(ctx, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return MessageResponse{}, errors.New("invalid id")
	}

	if _, exists := users[id]; !exists {
		return MessageResponse{}, errors.New("user not found")
	}
	delete(users, id)
	return MessageResponse{Message: "user deleted successfully"}, nil
}
