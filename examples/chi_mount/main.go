package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/RestStore/RestKit"
	restchi "github.com/RestStore/RestKit/adapters/chi"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	router := chi.NewRouter()
	router.Get("/users", listUsers)
	router.Get("/users/{id}", getUser)
	router.Get("/users/{id}/posts", listUserPosts)
	router.Post("/users/{id}/posts", createUserPost)

	api := restkit.NewApi().
		WithTitle("My API").WithSwaggerUI()

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
	}

	restchi.Mount(api, "/", router, meta)
	http.ListenAndServe(":8080", api.Mux())
}

type Post struct {
	ID      int    `json:"id"`
	UserID  int    `json:"user_id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type CreatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func listUsers(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode([]User{{ID: 1, Name: "Alice"}})
}

func getUser(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(User{ID: 1, Name: "Alice"})
}

func listUserPosts(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode([]Post{
		{ID: 1, UserID: 1, Title: "Hello", Content: "World"},
	})
}

func createUserPost(w http.ResponseWriter, r *http.Request) {
	var req CreatePostRequest
	json.NewDecoder(r.Body).Decode(&req)
	json.NewEncoder(w).
		Encode(Post{ID: 2, UserID: 1, Title: req.Title, Content: req.Content})
}
