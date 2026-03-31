package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/reststore/restkit"
	restchi "github.com/reststore/restkit/adapters/chi"
	_ "github.com/reststore/restkit/validation/playground"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Post struct {
	ID      int    `json:"id"`
	UserID  int    `json:"user_id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type CreatePostRequest struct {
	Title   string `json:"title"   validate:"required,min=3"`
	Content string `json:"content" validate:"required,min=10"`
}

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (r CreateUserRequest) Validate(ctx context.Context) restkit.ValidationResult {
	validation := restkit.NewValidation()

	if r.Name == "" {
		validation.AddError("name", "Name is required")
	}
	if len(r.Name) < 2 {
		validation.AddError("name", "Name must be at least 2 characters")
	}
	if r.Email == "" {
		validation.AddError("email", "Email is required")
	}
	if !isValidEmail(r.Email) {
		validation.AddError("email", "Invalid email format")
	}

	if validation.HasErrors() {
		validation.Status = 422
		validation.Code = "validation_failed"
		validation.Message = "Validation failed"
	}

	return validation
}

func isValidEmail(email string) bool {
	return len(email) > 3 && contains(email, "@")
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func main() {
	router := chi.NewRouter()

	router.Get("/users", listUsers)
	router.Get("/users/{id}", getUser)
	router.Get("/users/{id}/posts", listUserPosts)
	router.Post("/users/{id}/posts", createUserPost)
	router.Post("/users", createUser)

	api := restkit.NewApi().
		WithTitle("Chi Integration Example").
		WithDescription("Demonstrates automatic validation for existing Chi APIs").
		WithSwaggerUI()

	meta := []restkit.RouteMeta{
		{
			Method: "GET",
			Path:   "/users",
			Info: restkit.RouteInfo{
				Summary:      "List users",
				ResponseType: []User{},
			},
		},
		{
			Method: "GET",
			Path:   "/users/{id}",
			Info: restkit.RouteInfo{
				Summary:      "Get user",
				ResponseType: User{},
			},
		},
		{
			Method: "GET",
			Path:   "/users/{id}/posts",
			Info: restkit.RouteInfo{
				Summary:      "List user's posts",
				ResponseType: []Post{},
			},
		},
		{
			Method: "POST",
			Path:   "/users/{id}/posts",
			Info: restkit.RouteInfo{
				Summary:      "Create post for user",
				RequestType:  CreatePostRequest{},
				ResponseType: Post{},
			},
		},
		{
			Method: "POST",
			Path:   "/users",
			Info: restkit.RouteInfo{
				Summary:      "Create user",
				RequestType:  CreateUserRequest{},
				ResponseType: User{},
			},
		},
	}

	_ = restchi.Mount(api, "/api", router, meta)

	if err := http.ListenAndServe(":8080", api.Mux()); err != nil {
		panic(err)
	}
}

func listUsers(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode([]User{{ID: 1, Name: "Alice"}})
}

func getUser(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode(User{ID: 1, Name: "Alice"})
}

func listUserPosts(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode([]Post{
		{ID: 1, UserID: 1, Title: "Hello", Content: "World"},
	})
}

// Standard Chi handler - works with automatic validation
func createUserPost(w http.ResponseWriter, r *http.Request) {
	var req CreatePostRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	_ = json.NewEncoder(w).Encode(Post{
		ID:      2,
		UserID:  1,
		Title:   req.Title,
		Content: req.Content,
	})
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	_ = json.NewEncoder(w).Encode(User{
		ID:   2,
		Name: req.Name,
	})
}
