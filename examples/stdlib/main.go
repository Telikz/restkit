package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	rk "github.com/reststore/restkit"
	reststdlib "github.com/reststore/restkit/adapters/stdlib"
	"github.com/reststore/restkit/validators/playground"
)

func main() {
	store := NewUserStore()

	// Old way: manual route + handler registration
	// No swagger docs, no type safety, manual wiring
	stdlibMux := http.NewServeMux()
	stdlibMux.HandleFunc("GET /users", stdlibListUsers(store))
	stdlibMux.HandleFunc("GET /users/{id}", stdlibGetUser(store))
	stdlibMux.HandleFunc("POST /users", stdlibCreateUser(store))

	log.Println("Serving stdlib at :8080")
	go http.ListenAndServe(":8080", stdlibMux)

	// Create RestKit API
	api := rk.NewApi()
	api.WithSwaggerUI() // Serve Swagger UI at /swagger
	api.WithVersion("1.0.0")
	api.WithTitle("Stdlib + RestKit Example")

	// Migrate to RestKit with automatic OpenAPI docs
	// Define OpenAPI metadata for the old stdlib routes
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

	// Mount the old stdlib mux at /api/v1 with the provided metadata
	_ = reststdlib.Mount(api, "/api/v1", stdlibMux, meta)

	api.WithServer("http://localhost:8081", "Dev Server for old routes", nil)
	log.Println("Serving RestKit with mounted stdlib routes at :8081")
	go http.ListenAndServe(":8081", api.Mux())

	// Later on, we can migrate the routes fully to RestKit
	// avoiding the need to explicitly define the OpenAPI metadata for the old routes.

	api2 := rk.NewApi()
	api2.WithSwaggerUI()
	api2.WithVersion("2.0.0")
	api2.WithTitle("Stdlib + RestKit Example v2")
	api2.WithValidator(playground.NewValidator())
	_ = reststdlib.Mount(api2, "/v1", stdlibMux, meta)

	api2.AddGroup(rk.NewGroup("/v2/users").
		WithTitle("User Management").
		WithEndpoints(
			getUserEndpoint(store),
			listUsersEndpoint(store),
			createUserEndpoint(store),
		),
	)

	api2.WithServer("http://localhost:8082", "Dev Server for old and new routes", nil)
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

// stdlibCreateUser defines a standard library HTTP handler for creating a new user.
func stdlibCreateUser(store *UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateUserReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		user := store.Create(req.Name, req.Email)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err := json.NewEncoder(w).Encode(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// createUser defines a RestKit endpoint for creating a new user.
func createUserEndpoint(store *UserStore) *rk.Endpoint[CreateUserReq, UserResponse] {
	return rk.Create("/",
		func(_ context.Context, req CreateUserReq) (UserResponse, error) {
			user := store.Create(req.Name, req.Email)
			return userToResponse(user), nil
		},
	)
}

// stdlibGetUser defines a standard library HTTP handler for getting a user by ID.
func stdlibGetUser(store *UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
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
func getUserEndpoint(store *UserStore) *rk.Endpoint[rk.GetRequest, UserResponse] {
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

// stdlibListUsers defines a standard library HTTP handler for listing all users.
func stdlibListUsers(store *UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(usersToResponse(store.List()))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// listUsers defines a RestKit endpoint for listing all users.
func listUsersEndpoint(store *UserStore) *rk.Endpoint[rk.ListRequest, []UserResponse] {
	return rk.List("/",
		func(_ context.Context, req rk.ListRequest,
		) ([]UserResponse, error) {
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
	for i, u := range users {
		res[i] = userToResponse(u)
	}
	return res
}
