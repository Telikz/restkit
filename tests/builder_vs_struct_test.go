package restkit_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	rest "github.com/telikz/restkit"
	"github.com/telikz/restkit/internal/endpoints"
)

func BenchmarkBuilderAPI(b *testing.B) {
	for b.Loop() {
		ep := rest.NewEndpointRes[struct {
			Message string `json:"message"`
		}]().
			WithPath("/ping").
			WithMethod(http.MethodGet).
			WithHandler(func(ctx context.Context) (struct {
				Message string `json:"message"`
			}, error) {
				return struct {
					Message string `json:"message"`
				}{Message: "pong"}, nil
			})

		_ = ep.GetHandler()
	}
}

func BenchmarkStructDirect(b *testing.B) {
	for b.Loop() {
		ep := &endpoints.EndpointRes[struct {
			Message string `json:"message"`
		}]{
			Path:   "/ping",
			Method: http.MethodGet,
			Handler: func(ctx context.Context) (struct {
				Message string `json:"message"`
			}, error) {
				return struct {
					Message string `json:"message"`
				}{Message: "pong"}, nil
			},
		}

		_ = ep.GetHandler()
	}
}

func BenchmarkBuilderAPIWithDefaults(b *testing.B) {
	for b.Loop() {
		ep := rest.NewEndpointRes[struct {
			Message string `json:"message"`
		}]().
			WithPath("/ping").
			WithMethod(http.MethodGet).
			WithTitle("Ping Endpoint").
			WithDescription("Health check endpoint").
			WithHandler(func(ctx context.Context) (struct {
				Message string `json:"message"`
			}, error) {
				return struct {
					Message string `json:"message"`
				}{Message: "pong"}, nil
			})

		_ = ep.GetHandler()
	}
}

func BenchmarkStructDirectWithDefaults(b *testing.B) {
	for b.Loop() {
		ep := &endpoints.EndpointRes[struct {
			Message string `json:"message"`
		}]{
			Title:       "Ping Endpoint",
			Description: "Health check endpoint",
			Path:        "/ping",
			Method:      http.MethodGet,
			Handler: func(ctx context.Context) (struct {
				Message string `json:"message"`
			}, error) {
				return struct {
					Message string `json:"message"`
				}{Message: "pong"}, nil
			},
		}

		_ = ep.GetHandler()
	}
}

func BenchmarkBuilderAPIRequest(b *testing.B) {
	type Req struct {
		Name string `json:"name"`
	}
	type Res struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	for b.Loop() {
		ep := rest.NewEndpoint[Req, Res]().
			WithPath("/users").
			WithMethod(http.MethodPost).
			WithHandler(func(ctx context.Context, req Req) (Res, error) {
				return Res{ID: 1, Name: req.Name}, nil
			})

		_ = ep.GetHandler()
	}
}

func BenchmarkStructDirectRequest(b *testing.B) {
	type Req struct {
		Name string `json:"name"`
	}
	type Res struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	for b.Loop() {
		ep := &endpoints.EndpointReqRes[Req, Res]{
			Path:   "/users",
			Method: http.MethodPost,
			Handler: func(ctx context.Context, req Req) (Res, error) {
				return Res{ID: 1, Name: req.Name}, nil
			},
		}

		_ = ep.GetHandler()
	}
}

func BenchmarkBuilderFullRequest(b *testing.B) {
	type Res struct {
		Message string `json:"message"`
	}

	ep := rest.NewEndpointRes[Res]().
		WithPath("/ping").
		WithMethod(http.MethodGet).
		WithHandler(func(ctx context.Context) (Res, error) {
			return Res{Message: "pong"}, nil
		})

	handler := ep.GetHandler()

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatal("failed")
		}
	}
}

func BenchmarkStructFullRequest(b *testing.B) {
	type Res struct {
		Message string `json:"message"`
	}

	ep := &endpoints.EndpointRes[Res]{
		Path:   "/ping",
		Method: http.MethodGet,
		Handler: func(ctx context.Context) (Res, error) {
			return Res{Message: "pong"}, nil
		},
	}

	handler := ep.GetHandler()

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatal("failed")
		}
	}
}
