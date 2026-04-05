package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	rk "github.com/reststore/restkit"
)

type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateUserReq struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func createPrettyAPI() *rk.Api {
	api := rk.NewApi()
	api.WithTitle("Pretty JSON API")
	api.WithSerializer(rk.Serializers.JSONPretty())

	api.AddEndpoint(rk.NewEndpoint[rk.GetRequest, User]().
		WithPath("/users/{id}").
		WithMethod(http.MethodGet).
		WithHandler(func(ctx context.Context, req rk.GetRequest) (User, error) {
			return User{ID: req.ID, Name: "John Doe", Email: "john@example.com"}, nil
		}))

	return api
}

func createCompactAPI() *rk.Api {
	api := rk.NewApi()
	api.WithTitle("Compact JSON API")
	api.WithSerializer(rk.Serializers.JSONCompact())

	api.AddEndpoint(rk.NewEndpoint[rk.GetRequest, User]().
		WithPath("/users/{id}").
		WithMethod(http.MethodGet).
		WithHandler(func(ctx context.Context, req rk.GetRequest) (User, error) {
			return User{ID: req.ID, Name: "Jane Doe", Email: "jane@example.com"}, nil
		}))

	return api
}

func createStrictAPI() *rk.Api {
	api := rk.NewApi()
	api.WithTitle("Strict Input API")
	api.WithDeserializer(func(r *http.Request, req any) error {
		const maxSize = 10 * 1024
		body := make([]byte, r.ContentLength)
		if r.ContentLength > maxSize {
			return bytes.ErrTooLarge
		}
		n, err := r.Body.Read(body)
		if err != nil && err.Error() != "EOF" {
			return err
		}
		return json.Unmarshal(body[:n], req)
	})

	api.AddEndpoint(rk.NewEndpoint[CreateUserReq, User]().
		WithPath("/users").
		WithMethod(http.MethodPost).
		WithHandler(func(ctx context.Context, req CreateUserReq) (User, error) {
			return User{ID: 1, Name: req.Name, Email: req.Email}, nil
		}))

	return api
}

func main() {
	go createPrettyAPI().Serve(":8080")
	go createCompactAPI().Serve(":8081")
	go createStrictAPI().Serve(":8082")
	select {}
}
