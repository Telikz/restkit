package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	rk "github.com/reststore/restkit"
)

type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CreateUserRequest defines the request body for creating a new user.
type CreateUserReq struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// createUserEndpoint demonstrates an endpoint that creates a user.
func createUserEndpoint() *rk.Endpoint[CreateUserReq, User] {
	return rk.Post("/users",
		func(ctx context.Context, req CreateUserReq) (User, error) {
			return User{ID: 1, Name: req.Name, Email: req.Email}, nil
		},
	).WithMiddleware(rk.CORSMiddleware())
}

// createPrettyAPI demonstrates using a pretty JSON serializer
// that produces indented JSON output for better readability.
func createPrettyAPI() *rk.Api {
	api := rk.NewApi()
	api.WithSwaggerUI()
	api.WithTitle("Pretty JSON API")
	api.WithSerializer(rk.Serializers.JSONPretty())

	// Add the other APIs as additional servers on different ports
	// for testing different serializers/deserializers with the same endpoints.
	api.WithServer("http://localhost:8080", "Pretty Json Api", nil)
	api.WithServer("http://localhost:8081", "Compact Json Api", nil)
	api.WithServer("http://localhost:8082", "Strict Input Api", nil)
	return api.AddEndpoint(createUserEndpoint())
}

// createCompactAPI demonstrates using a compact JSON serializer
// that produces minified JSON output.
func createCompactAPI() *rk.Api {
	api := rk.NewApi()
	api.WithTitle("Compact JSON API")
	api.WithSerializer(rk.Serializers.JSONCompact())
	return api.AddEndpoint(createUserEndpoint())
}

// createStrictAPI demonstrates a custom deserializer
// that limits input size and returns an error if the input is too large.
func createStrictAPI() *rk.Api {
	api := rk.NewApi()
	api.WithTitle("Strict Input API")
	api.WithDeserializer(customDeserializer)
	return api.AddEndpoint(createUserEndpoint())
}

func main() {
	go createCompactAPI().Serve(":8081") // Start compact API on port 8081
	go createStrictAPI().Serve(":8082")  // Start strict API on port 8082

	fmt.Print("Running swagger at http://localhost:8080/swagger...")

	// Start the pretty API on port 8080, which also serves the OpenAPI docs.
	if err := createPrettyAPI().Serve(":8080"); err != nil {
		panic(err)
	}
}

// customDeserializer is an example of a custom deserializer
// that limits the size of the input JSON.
func customDeserializer(r *http.Request, req any) error {
	const maxSize = 1024
	body := make([]byte, r.ContentLength)
	if r.ContentLength > maxSize {
		return bytes.ErrTooLarge
	}
	n, err := r.Body.Read(body)
	if err != nil && err.Error() != "EOF" {
		return err
	}
	return json.Unmarshal(body[:n], req)
}
