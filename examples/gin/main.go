package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	rk "github.com/reststore/restkit"
	restgin "github.com/reststore/restkit/adapters/gin"
	"github.com/reststore/restkit/validators/playground"
)

func main() {
	store := NewUserStore()

	// Old way: manual route + handler registration
	// No swagger docs, no type safety, manual wiring
	ginRouter := gin.New()
	ginRouter.GET("/users", ginListUsers(store))
	ginRouter.GET("/users/:id", ginGetUser(store))
	ginRouter.POST("/users", ginCreateUser(store))

	log.Println("Serving Gin at :8080")
	go http.ListenAndServe(":8080", ginRouter)

	// Create RestKit API
	api := rk.NewApi()
	api.WithSwaggerUI() // Serve Swagger UI at /swagger
	api.WithVersion("1.0.0")
	api.WithTitle("Gin + RestKit Example")

	// Migrate to RestKit with automatic OpenAPI docs
	// Define OpenAPI metadata for the old gin routes
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

	// Mount the old gin router at /api/v1 with the provided metadata
	_ = restgin.Mount(api, "/api/v1", ginRouter, meta)

	api.WithServers("http://localhost:8081") // Set server URL for OpenAPI docs
	log.Println("Serving RestKit with mounted Gin routes at :8081")
	go http.ListenAndServe(":8081", api.Mux())

	// Later on, we can migrate the routes fully to RestKit
	// avoiding the need to explicitly define the OpenAPI metadata for the old routes.

	api2 := rk.NewApi()
	api2.WithSwaggerUI()
	api2.WithVersion("2.0.0")
	api2.WithTitle("Gin + RestKit Example v2")
	api2.WithValidator(playground.NewValidator())
	_ = restgin.Mount(api2, "/v1", ginRouter, meta)

	api2.AddGroup(rk.NewGroup("/v2/users").
		WithTitle("User Management").
		WithEndpoints(
			createUser(store),
			getUser(store),
			listUsers(store),
		),
	)

	api2.WithServers("http://localhost:8082") // Set server URL for OpenAPI docs
	log.Println("Serving RestKit API with old and new routes at :8082")
	http.ListenAndServe(":8082", api2.Mux())
}

type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateUserReq struct {
	Name  string `json:"name"  validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email"`
}

// ginCreateUser defines a Gin HTTP handler for creating a new user.
func ginCreateUser(store *UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateUserReq
		if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
			c.JSON(http.StatusBadRequest,
				gin.H{"error": "invalid request body"})
			return
		}
		user := store.Create(req.Name, req.Email)
		c.JSON(http.StatusCreated, userToResponse(user))
	}
}

// createUser defines a RestKit endpoint for creating a new user.
func createUser(store *UserStore) *rk.Endpoint[CreateUserReq, UserResponse] {
	return rk.Create("/",
		func(_ context.Context, req CreateUserReq) (UserResponse, error) {
			user := store.Create(req.Name, req.Email)
			return userToResponse(user), nil
		},
	)
}

// ginGetUser defines a Gin HTTP handler for getting a user by ID.
func ginGetUser(store *UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest,
				gin.H{"error": "invalid user ID"})
			return
		}
		user, err := store.Get(id)
		if err != nil {
			c.JSON(http.StatusNotFound,
				gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusOK, userToResponse(user))
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

// ginListUsers defines a Gin HTTP handler for listing all users.
func ginListUsers(store *UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, usersToResponse(store.List()))
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

func userToResponse(user User) UserResponse {
	return UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
}

func usersToResponse(users []User) []UserResponse {
	res := make([]UserResponse, len(users))
	for i, user := range users {
		res[i] = userToResponse(user)
	}
	return res
}
