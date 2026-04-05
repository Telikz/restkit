package endpoints

import (
	"context"
	"net/http"

	rk "github.com/reststore/restkit"
)

type PingResponse struct {
	Message string `json:"message"`
}

func Ping() *rk.Endpoint[rk.NoRequest, PingResponse] {
	return rk.NewEndpoint[rk.NoRequest, PingResponse]().
		WithPath("/ping").
		WithMethod(http.MethodGet).
		WithTitle("Ping Endpoint").
		WithDescription("A simple endpoint to test connectivity").
		WithHandler(func(ctx context.Context, _ rk.NoRequest) (PingResponse, error) {
			return PingResponse{Message: "pong"}, nil
		})
}
