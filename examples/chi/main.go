package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/reststore/restkit"
	restchi "github.com/reststore/restkit/adapters/chi"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateUserReq struct {
	Name  string `json:"name"  validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

func main() {
	r := chi.NewRouter()

	api := restkit.NewApi().
		WithVersion("1.0.0").
		WithTitle("Chi + RestKit Example").
		WithSwaggerUI("/docs")

	api.AddGroup(userGroup())
	api.AddEndpoint(ping())
	restchi.RegisterRoutes(r, api)

	chiRouter := chi.NewRouter()
	chiRouter.Get("/native/users", chiListUsers)
	chiRouter.Get("/native/users/{id}", chiGetUser)
	chiRouter.Post("/native/users", chiCreateUser)

	meta := []restkit.RouteMeta{
		{
			Method: "GET",
			Path:   "/native/users",
			Info: restkit.RouteInfo{
				Summary:      "List users",
				ResponseType: []User{},
			},
		},
		{
			Method: "GET",
			Path:   "/native/users/{id}",
			Info: restkit.RouteInfo{
				Summary:      "Get user",
				ResponseType: User{},
			},
		},
		{
			Method: "POST",
			Path:   "/native/users",
			Info: restkit.RouteInfo{
				Summary:      "Create user",
				RequestType:  CreateUserReq{},
				ResponseType: User{},
			},
		},
	}

	_ = restchi.Mount(api, "/api/v1", chiRouter, meta)

	log.Println("Server on :8080")
	log.Fatal(http.ListenAndServe(":8080", api.Mux()))
}

func userGroup() *restkit.Group {
	return restkit.NewGroup("/users").
		WithEndpoints(
			createUser(),
			getUser(),
			listUsers(),
		)
}

func createUser() *restkit.Endpoint[CreateUserReq, User] {
	return restkit.NewEndpoint[CreateUserReq, User]().
		WithPath("/").
		WithMethod("POST").
		WithHandler(func(ctx context.Context, req CreateUserReq) (User, error) {
			return User{ID: 1, Name: req.Name, Email: req.Email}, nil
		})
}

func getUser() *restkit.Endpoint[restkit.NoRequest, User] {
	return restkit.NewEndpoint[restkit.NoRequest, User]().
		WithPath("/{id}").
		WithMethod("GET").
		WithHandler(func(ctx context.Context, _ restkit.NoRequest) (User, error) {
			return User{ID: 1, Name: "John", Email: "john@example.com"}, nil
		})
}

func listUsers() *restkit.Endpoint[restkit.NoRequest, []User] {
	return restkit.NewEndpoint[restkit.NoRequest, []User]().
		WithPath("/").
		WithMethod("GET").
		WithHandler(func(ctx context.Context, _ restkit.NoRequest) ([]User, error) {
			return []User{{ID: 1, Name: "John", Email: "john@example.com"}}, nil
		})
}

func ping() *restkit.Endpoint[restkit.NoRequest, Pong] {
	return restkit.NewEndpoint[restkit.NoRequest, Pong]().
		WithPath("/ping").
		WithMethod("GET").
		WithHandler(func(ctx context.Context, _ restkit.NoRequest) (Pong, error) {
			return Pong{Message: "pong"}, nil
		})
}

type Pong struct {
	Message string `json:"message"`
}

func chiListUsers(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode([]User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
	})
}

func chiGetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_ = json.NewEncoder(w).Encode(User{ID: 1, Name: "Alice (" + id + ")"})
}

func chiCreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserReq
	_ = json.NewDecoder(r.Body).Decode(&req)
	_ = json.NewEncoder(w).Encode(User{ID: 3, Name: req.Name, Email: req.Email})
}
