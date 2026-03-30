package endpoints

import (
	"context"
	"net/http"

	rest "github.com/telikz/restkit"
)

type PingResponse struct {
	Message string `json:"message"`
}

func Ping() *rest.EndpointRes[PingResponse] {
	return rest.NewEndpointRes[PingResponse]().
		WithPath("/ping").
		WithMethod(http.MethodGet).
		WithTitle("Ping Endpoint").
		WithDescription("A simple endpoint to test connectivity").
		WithHandler(func(ctx context.Context) (PingResponse, error) {
			return PingResponse{Message: "pong"}, nil
		})
}
