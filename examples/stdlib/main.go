package main

import (
	"context"
	"log"
	"net/http"

	rk "github.com/reststore/restkit"
	reststdlib "github.com/reststore/restkit/adapters/stdlib"
)

func main() {
	mux := http.NewServeMux()

	api := rk.NewApi().
		WithVersion("1.0.0").
		WithTitle("Stdlib + RestKit Example").
		WithSwaggerUI("/docs")

	api.AddGroup(userGroup())
	api.AddEndpoint(ping())
	reststdlib.RegisterRoutes(mux, api)

	stdlibMux := http.NewServeMux()
	stdlibMux.HandleFunc("GET /native/users", stdlibListUsers)
	stdlibMux.HandleFunc("GET /native/users/{id}", stdlibGetUser)
	stdlibMux.HandleFunc("POST /native/users", stdlibCreateUser)

	meta := []rk.RouteMeta{
		{
			Method: "GET",
			Path:   "/native/users",
			Info:   rk.RouteInfo{Summary: "List users", ResponseType: []User{}},
		},
		{
			Method: "GET",
			Path:   "/native/users/{id}",
			Info:   rk.RouteInfo{Summary: "Get user", ResponseType: User{}},
		},
		{
			Method: "POST",
			Path:   "/native/users",
			Info: rk.RouteInfo{
				Summary:      "Create user",
				RequestType:  CreateUserReq{},
				ResponseType: User{},
			},
		},
	}

	_ = reststdlib.Mount(api, "/api/v1", stdlibMux, meta)

	log.Println("Server on :8080")
	log.Fatal(http.ListenAndServe(":8080", api.Mux()))
}

type CreateUserReq struct {
	Name  string `json:"name"  validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func userGroup() *rk.Group {
	return rk.NewGroup("/users").
		WithEndpoints(
			createUser(),
			getUser(),
			listUsers(),
		)
}

func createUser() *rk.Endpoint[CreateUserReq, User] {
	return rk.NewEndpoint[CreateUserReq, User]().
		WithPath("/").
		WithMethod("POST").
		WithHandler(func(ctx context.Context, req CreateUserReq) (User, error) {
			return User{ID: 1, Name: req.Name, Email: req.Email}, nil
		})
}

func getUser() *rk.Endpoint[rk.NoRequest, User] {
	return rk.NewEndpoint[rk.NoRequest, User]().
		WithPath("/{id}").
		WithMethod("GET").
		WithHandler(func(ctx context.Context, _ rk.NoRequest) (User, error) {
			return User{ID: 1, Name: "John", Email: "john@example.com"}, nil
		})
}

func listUsers() *rk.Endpoint[rk.NoRequest, []User] {
	return rk.NewEndpoint[rk.NoRequest, []User]().
		WithPath("/").
		WithMethod("GET").
		WithHandler(func(ctx context.Context, _ rk.NoRequest) ([]User, error) {
			return []User{{ID: 1, Name: "John", Email: "john@example.com"}}, nil
		})
}

func ping() *rk.Endpoint[rk.NoRequest, Pong] {
	return rk.NewEndpoint[rk.NoRequest, Pong]().
		WithPath("/ping").
		WithMethod("GET").
		WithHandler(func(ctx context.Context, _ rk.NoRequest) (Pong, error) {
			return Pong{Message: "pong"}, nil
		})
}

type Pong struct {
	Message string `json:"message"`
}

func stdlibListUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]`))
}

func stdlibGetUser(w http.ResponseWriter, r *http.Request) {
	_ = r.PathValue("id")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"id":1,"name":"Alice"}`))
}

func stdlibCreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"id":3,"name":"New User"}`))
}
