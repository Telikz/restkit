package restkit_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	rest "github.com/telikz/restkit"
	restchi "github.com/telikz/restkit/adapters/chi"
	ep "github.com/telikz/restkit/internal/endpoints"
)

type CreateUserRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// BenchmarkRestKitPing tests GET /ping using RestKit handler
func BenchmarkRestKitPing(b *testing.B) {
	router := setupRestKitRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	for b.Loop() {
		resp, err := http.Get(server.URL + "/ping")
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkRawChiPing tests GET /ping using raw Chi handler
func BenchmarkRawChiPing(b *testing.B) {
	router := setupRawChiRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	for b.Loop() {
		resp, err := http.Get(server.URL + "/ping")
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkStdlibPing tests GET /ping using stdlib ServeMux
func BenchmarkStdlibPing(b *testing.B) {
	mux := setupStdlibMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	for b.Loop() {
		resp, err := http.Get(server.URL + "/ping")
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkRestKitGetUser tests GET /users/{id} using RestKit with path parameter
func BenchmarkRestKitGetUser(b *testing.B) {
	router := setupRestKitRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	for b.Loop() {
		resp, err := http.Get(server.URL + "/users/1")
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkRawChiGetUser tests GET /users/{id} using raw Chi with chi.URLParam
func BenchmarkRawChiGetUser(b *testing.B) {
	router := setupRawChiRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	for b.Loop() {
		resp, err := http.Get(server.URL + "/users/1")
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkStdlibGetUser tests GET /users/{id} using stdlib ServeMux with r.PathValue
func BenchmarkStdlibGetUser(b *testing.B) {
	mux := setupStdlibMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	for b.Loop() {
		resp, err := http.Get(server.URL + "/users/1")
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkRestKitCreateUser tests POST /users with JSON body binding using RestKit
func BenchmarkRestKitCreateUser(b *testing.B) {
	router := setupRestKitRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	reqBody := CreateUserRequest{
		Name:  "John Doe",
		Email: "john@example.com",
	}
	jsonBody, _ := json.Marshal(reqBody)

	for b.Loop() {
		resp, err := http.Post(
			server.URL+"/users",
			"application/json",
			bytes.NewReader(jsonBody),
		)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkRawChiCreateUser tests POST /users with JSON decoding using raw Chi
func BenchmarkRawChiCreateUser(b *testing.B) {
	router := setupRawChiRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	reqBody := CreateUserRequest{
		Name:  "John Doe",
		Email: "john@example.com",
	}
	jsonBody, _ := json.Marshal(reqBody)

	for b.Loop() {
		resp, err := http.Post(
			server.URL+"/users",
			"application/json",
			bytes.NewReader(jsonBody),
		)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkStdlibCreateUser tests POST /users with JSON decoding using stdlib
func BenchmarkStdlibCreateUser(b *testing.B) {
	mux := setupStdlibMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	reqBody := CreateUserRequest{
		Name:  "John Doe",
		Email: "john@example.com",
	}
	jsonBody, _ := json.Marshal(reqBody)

	for b.Loop() {
		resp, err := http.Post(
			server.URL+"/users",
			"application/json",
			bytes.NewReader(jsonBody),
		)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func setupRestKitRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(chiMiddleware.Recoverer)

	api := &rest.Api{
		Version:     "1.0.0",
		Title:       "Example API",
		Description: "An example API using RestKit with Chi",
		Groups:      []*ep.Group{userGroup()},
		Endpoints:   []ep.Endpoint{pingEndpoint()},
	}

	restchi.RegisterRoutes(r, api)
	return r
}

func setupRawChiRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(chiMiddleware.Recoverer)

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "pong"})
	})

	r.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(UserResponse{
			ID:    1,
			Name:  "John",
			Email: "john@example.com",
		})
		_ = id
	})

	r.Post("/users", func(w http.ResponseWriter, r *http.Request) {
		var req CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(UserResponse{
			ID:    1,
			Name:  req.Name,
			Email: req.Email,
		})
	})

	r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]UserResponse{
			{ID: 1, Name: "John", Email: "john@example.com"},
		})
	})

	return r
}

func setupStdlibMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "pong"})
	})

	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(UserResponse{
			ID:    1,
			Name:  "John",
			Email: "john@example.com",
		})
		_ = id
	})

	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		var req CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(UserResponse{
			ID:    1,
			Name:  req.Name,
			Email: req.Email,
		})
	})

	mux.HandleFunc("GET /users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]UserResponse{
			{ID: 1, Name: "John", Email: "john@example.com"},
		})
	})

	return mux
}

func userGroup() *ep.Group {
	return &ep.Group{
		Prefix:      "/users",
		Title:       "User Management",
		Description: "Endpoints for managing users",
		Endpoints:   []ep.Endpoint{createUserEndpoint(), getUserEndpoint(), listUsersEndpoint()},
	}
}

func createUserEndpoint() *rest.Endpoint[CreateUserRequest, UserResponse] {
	return rest.NewEndpoint[CreateUserRequest, UserResponse]().
		WithPath("").
		WithMethod(http.MethodPost).
		WithHandler(func(ctx context.Context, req CreateUserRequest) (UserResponse, error) {
			return UserResponse{ID: 1, Name: req.Name, Email: req.Email}, nil
		})
}

func getUserEndpoint() *rest.EndpointRes[UserResponse] {
	return rest.NewEndpointRes[UserResponse]().
		WithPath("/{id}").
		WithMethod(http.MethodGet).
		WithHandler(func(ctx context.Context) (UserResponse, error) {
			return UserResponse{ID: 1, Name: "John", Email: "john@example.com"}, nil
		})
}

func listUsersEndpoint() *rest.EndpointRes[[]UserResponse] {
	return rest.NewEndpointRes[[]UserResponse]().
		WithPath("").
		WithMethod(http.MethodGet).
		WithHandler(func(ctx context.Context) ([]UserResponse, error) {
			return []UserResponse{{ID: 1, Name: "John", Email: "john@example.com"}}, nil
		})
}

func pingEndpoint() *rest.EndpointRes[map[string]string] {
	return rest.NewEndpointRes[map[string]string]().
		WithPath("/ping").
		WithMethod(http.MethodGet).
		WithHandler(func(ctx context.Context) (map[string]string, error) {
			return map[string]string{"message": "pong"}, nil
		})
}
