package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	rk "github.com/reststore/restkit"
	restchi "github.com/reststore/restkit/adapters/chi"
	"github.com/reststore/restkit/validators/playground"
)

func main() {
	store := NewUserStore()

	// Old way: manual route + handler registration
	// No swagger docs, no type safety, manual wiring
	chiRouter := chi.NewRouter()
	chiRouter.Get("/users", chiListUsers(store))
	chiRouter.Get("/users/{id}", chiGetUser(store))
	chiRouter.Post("/users", chiCreateUser(store))

	log.Println("Serving Chi at :8080")
	go http.ListenAndServe(":8080", chiRouter)

	// Create RestKit API
	api := rk.NewApi()
	api.WithSwaggerUI() // Serve Swagger UI at /swagger
	api.WithVersion("1.0.0")
	api.WithTitle("Chi + RestKit Example")

	// Migrate to RestKit with automatic OpenAPI docs
	// Define OpenAPI metadata for the old chi routes
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
			Path:   "/users/{id}",
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

	// Mount the old chi router at /api/v1 with the provided metadata
	_ = restchi.Mount(api, "/api/v1", chiRouter, meta)

	api.WithServers("http://localhost:8081") // Set server URL for OpenAPI docs
	log.Println("Serving RestKit with mounted Chi routes at :8081")
	go http.ListenAndServe(":8081", api.Mux())

	// Later on, we can migrate the routes fully to RestKit
	// avoiding the need to explicitly define the OpenAPI metadata for the old routes.

	api2 := rk.NewApi()
	api2.WithSwaggerUI()
	api2.WithVersion("2.0.0")
	api2.WithTitle("Chi + RestKit Example v2")
	api2.WithValidator(playground.NewValidator())
	_ = restchi.Mount(api2, "/v1", chiRouter, meta)

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

// chiCreateUser defines a Chi HTTP handler for creating a new user.
func chiCreateUser(store *UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateUserReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		user := store.Create(req.Name, req.Email)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err := json.NewEncoder(w).Encode(userToResponse(user))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// createUser defines a RestKit endpoint for creating a new user.
func createUser(store *UserStore) *rk.Endpoint[CreateUserReq, UserResponse] {
	return rk.Create("/",
		func(_ context.Context, req CreateUserReq) (UserResponse, error) {
			return userToResponse(store.Create(req.Name, req.Email)), nil
		},
	)
}

// chiGetUser defines a Chi HTTP handler for getting a user by ID.
func chiGetUser(store *UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid user ID", http.StatusBadRequest)
			return
		}
		user, err := store.Get(id)
		if err != nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(userToResponse(user))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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

// chiListUsers defines a Chi HTTP handler for listing all users.
func chiListUsers(store *UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := usersToResponse(store.List())
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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

func userToResponse(u User) UserResponse {
	return UserResponse{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
	}
}

func usersToResponse(users []User) []UserResponse {
	responses := make([]UserResponse, len(users))
	for i, user := range users {
		responses[i] = userToResponse(user)
	}
	return responses
}
