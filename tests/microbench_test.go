package restkit_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	rest "github.com/RestStore/RestKit"
	routectx "github.com/RestStore/RestKit/internal/context"
	"github.com/RestStore/RestKit/internal/endpoints"
)

func BenchmarkExtractPathParams(b *testing.B) {
	pattern := "/users/{id}/posts/{postId}"
	path := "/users/123/posts/456"

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		params := routectx.ExtractPathParams(pattern, path)
		_ = params
	}
}

func BenchmarkDirectHandlerCall(b *testing.B) {
	type Req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	type Res struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	endpoint := rest.NewEndpoint[Req, Res]().
		WithPath("/users").
		WithMethod(http.MethodPost).
		WithHandler(func(ctx context.Context, req Req) (Res, error) {
			return Res{ID: 1, Name: req.Name, Email: req.Email}, nil
		})

	handler := endpoint.GetHandler()

	reqBody := Req{Name: "John", Email: "john@example.com"}
	jsonBody, _ := json.Marshal(reqBody)

	for b.Loop() {
		req := httptest.NewRequest(
			http.MethodPost,
			"/users",
			bytes.NewReader(jsonBody),
		)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", rec.Code)
		}
	}
}

func BenchmarkDirectHandlerWithPathParam(b *testing.B) {
	type Res struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	endpoint := rest.NewEndpointRes[Res]().
		WithPath("/users/{id}").
		WithMethod(http.MethodGet).
		WithHandler(func(ctx context.Context) (Res, error) {
			return Res{ID: 1, Name: "John", Email: "john@example.com"}, nil
		})

	handler := endpoint.GetHandler()

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
		req.SetPathValue("id", "123")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", rec.Code)
		}
	}
}

func BenchmarkRouteContextCreation(b *testing.B) {
	for b.Loop() {
		rc := routectx.NewRouteContext()
		rc.SetURLParam("id", "123")
		rc.SetURLParam("postId", "456")
		_ = rc.URLParam("id")
	}
}

func BenchmarkNoPathParamsEndpoint(b *testing.B) {
	type Res struct {
		Message string `json:"message"`
	}

	endpoint := rest.NewEndpointRes[Res]().
		WithPath("/ping").
		WithMethod(http.MethodGet).
		WithHandler(func(ctx context.Context) (Res, error) {
			return Res{Message: "pong"}, nil
		})

	handler := endpoint.GetHandler()

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", rec.Code)
		}
	}
}

func BenchmarkEndpointCall(b *testing.B) {
	type Res struct {
		Message string `json:"message"`
	}

	endpoint := &endpoints.EndpointRes[Res]{
		Title:       "Ping",
		Description: "Health check",
		Method:      http.MethodGet,
		Path:        "/ping",
		Handler: func(ctx context.Context) (Res, error) {
			return Res{Message: "pong"}, nil
		},
	}

	endpoint.GetHandler()

	for b.Loop() {
		ctx := context.Background()
		res, err := endpoint.Handler(ctx)
		if err != nil {
			b.Fatal(err)
		}
		_ = res
	}
}
