package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	rk "github.com/reststore/restkit"
	restecho "github.com/reststore/restkit/adapters/echo"
	"github.com/reststore/restkit/validators/playground"
)

func main() {
	store := NewUserStore()

	// Old way: manual route + handler registration
	// No swagger docs, no type safety, manual wiring
	echoRouter := echo.New()
	echoRouter.GET("/users", echoListUsers(store))
	echoRouter.GET("/users/:id", echoGetUser(store))
	echoRouter.POST("/users", echoCreateUser(store))

	log.Println("Serving Echo at :8080")
	go http.ListenAndServe(":8080", echoRouter)

	// Create RestKit API
	api := rk.NewApi()
	api.WithSwaggerUI() // Serve Swagger UI at /swagger
	api.WithVersion("1.0.0")
	api.WithTitle("Echo + RestKit Example")

	// Migrate to RestKit with automatic OpenAPI docs
	// Define OpenAPI metadata for the old echo routes
	meta := []rk.RouteMeta{
		{
			Method: "GET",
			Path:   "/users",
			Info: rk.RouteInfo{
				Summary:      "List users",
				ResponseType: []User{}},
		},
		{
			Method: "GET",
			Path:   "/users/:id",
			Info: rk.RouteInfo{
				Summary:      "Get user",
				ResponseType: User{}},
		},
		{
			Method: "POST",
			Path:   "/users",
			Info: rk.RouteInfo{
				Summary:      "Create user",
				RequestType:  CreateUserReq{},
				ResponseType: User{},
			},
		},
	}

	// We get validation for free on the old routes!
	// Using the playground validator we can add struct tags to our request types.
	api.WithValidator(playground.NewValidator())

	// Mount the old echo router at /api/v1 with the provided metadata
	_ = restecho.Mount(api, "/v1", echoRouter, meta)

	api.WithServer("http://localhost:8081", "RestKit API", nil) // Set server URL for OpenAPI docs
	log.Println("Serving RestKit with mounted Echo routes at :8081")
	go http.ListenAndServe(":8081", api.Mux())

	// Later on, we can migrate the routes fully to RestKit
	// avoiding the need to explicitly define the OpenAPI metadata for the old routes.

	api2 := rk.NewApi()
	api2.WithSwaggerUI()
	api2.WithVersion("2.0.0")
	api2.WithTitle("Echo + RestKit Example v2")
	api2.WithValidator(playground.NewValidator())
	_ = restecho.Mount(api2, "/v1", echoRouter, meta)

	api2.AddGroup(rk.NewGroup("/v2/users").
		WithTitle("User Management").
		WithEndpoints(
			createUser(store),
			getUser(store),
			listUsers(store),
		),
	)

	api2.WithServer("http://localhost:8082", "RestKit API", nil) // Set server URL for OpenAPI docs
	log.Println("Serving RestKit API with old and new routes at :8082")
	http.ListenAndServe(":8082", api2.Mux())
}

type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func userToResponse(u User) UserResponse {
	return UserResponse{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
	}
}

func usersToResponse(users []User) []UserResponse {
	resp := make([]UserResponse, len(users))
	for i, u := range users {
		resp[i] = userToResponse(u)
	}
	return resp
}

type CreateUserReq struct {
	Name  string `json:"name"  validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email"`
}

// echoCreateUser defines an Echo HTTP handler for creating a new user.
func echoCreateUser(store *UserStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req CreateUserReq
		if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
			return c.JSON(http.StatusBadRequest,
				echo.Map{"error": "invalid request body"})
		}
		user := store.Create(req.Name, req.Email)
		return c.JSON(http.StatusCreated, userToResponse(user))
	}
}

// createUser defines a RestKit endpoint for creating a new user.
func createUser(store *UserStore) *rk.Endpoint[CreateUserReq, UserResponse] {
	return rk.Post("/",
		func(_ context.Context, req CreateUserReq) (UserResponse, error) {
			user := store.Create(req.Name, req.Email)
			return userToResponse(user), nil
		},
	)
}

// echoGetUser defines an Echo HTTP handler for getting a user by ID.
func echoGetUser(store *UserStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest,
				echo.Map{"error": "invalid user ID"})
		}
		user, err := store.Get(id)
		if err != nil {
			return c.JSON(http.StatusNotFound,
				echo.Map{"error": "user not found"})
		}
		return c.JSON(http.StatusOK, userToResponse(user))
	}
}

// getUser defines a RestKit endpoint for getting a user by ID.
func getUser(store *UserStore) *rk.Endpoint[rk.GetRequest, UserResponse] {
	return rk.Get("/{id}",
		func(_ context.Context, req rk.GetRequest) (UserResponse, error) {
			user, err := store.Get(req.ID)
			if err != nil {
				return UserResponse{}, err
			}
			return userToResponse(user), nil
		},
	)
}

// echoListUsers defines an Echo HTTP handler for listing all users.
func echoListUsers(store *UserStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, usersToResponse(store.List()))
	}
}

// listUsers defines a RestKit endpoint for listing all users.
func listUsers(store *UserStore) *rk.Endpoint[rk.ListRequest, []UserResponse] {
	return rk.List("/",
		func(_ context.Context, req rk.ListRequest) ([]UserResponse, error) {
			return usersToResponse(store.List()), nil
		},
	)
}
