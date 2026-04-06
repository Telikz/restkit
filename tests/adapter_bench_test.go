package restkit_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	"github.com/labstack/echo/v4"
	rest "github.com/reststore/restkit"
	restchi "github.com/reststore/restkit/adapters/chi"
	restecho "github.com/reststore/restkit/adapters/echo"
	restgin "github.com/reststore/restkit/adapters/gin"
)

type BenchRes struct {
	Message string `json:"message"`
}

func setupChiAdapter() (*chi.Mux, *rest.Api) {
	r := chi.NewRouter()
	api := rest.NewApi().
		WithTitle("Bench API").
		WithVersion("1.0.0").
		WithSwaggerUI("/docs")

	endpoint := rest.NewEndpointRes[BenchRes]().
		WithPath("/ping").
		WithMethod(http.MethodGet).
		WithHandler(func(ctx context.Context, _ rest.NoRequest) (BenchRes, error) {
			return BenchRes{Message: "pong"}, nil
		})

	api.AddEndpoint(endpoint)
	restchi.RegisterRoutes(r, api)
	return r, api
}

func setupGinAdapter() (*gin.Engine, *rest.Api) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := rest.NewApi().
		WithTitle("Bench API").
		WithVersion("1.0.0").
		WithSwaggerUI("/docs")

	endpoint := rest.NewEndpointRes[BenchRes]().
		WithPath("/ping").
		WithMethod(http.MethodGet).
		WithHandler(func(ctx context.Context, _ rest.NoRequest) (BenchRes, error) {
			return BenchRes{Message: "pong"}, nil
		})

	api.AddEndpoint(endpoint)
	restgin.RegisterRoutes(r, api)
	return r, api
}

func setupEchoAdapter() (*echo.Echo, *rest.Api) {
	e := echo.New()
	api := rest.NewApi().
		WithTitle("Bench API").
		WithVersion("1.0.0").
		WithSwaggerUI("/docs")

	endpoint := rest.NewEndpointRes[BenchRes]().
		WithPath("/ping").
		WithMethod(http.MethodGet).
		WithHandler(func(ctx context.Context, _ rest.NoRequest) (BenchRes, error) {
			return BenchRes{Message: "pong"}, nil
		})

	api.AddEndpoint(endpoint)
	restecho.RegisterRoutes(e, api)
	return e, api
}

func BenchmarkChiAdapter_SimpleRequest(b *testing.B) {
	r, _ := setupChiAdapter()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", rec.Code)
		}
	}
}

func BenchmarkGinAdapter_SimpleRequest(b *testing.B) {
	r, _ := setupGinAdapter()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", rec.Code)
		}
	}
}

func BenchmarkEchoAdapter_SimpleRequest(b *testing.B) {
	e, _ := setupEchoAdapter()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", rec.Code)
		}
	}
}

func BenchmarkChiAdapter_PathParams(b *testing.B) {
	r := chi.NewRouter()
	api := rest.NewApi()

	endpoint := rest.NewEndpointRes[BenchRes]().
		WithPath("/users/{id}/posts/{postId}").
		WithMethod(http.MethodGet).
		WithHandler(func(ctx context.Context, _ rest.NoRequest) (BenchRes, error) {
			return BenchRes{Message: "found"}, nil
		})

	api.AddEndpoint(endpoint)
	restchi.RegisterRoutes(r, api)

	req := httptest.NewRequest(http.MethodGet, "/users/123/posts/456", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", rec.Code)
		}
	}
}

func BenchmarkGinAdapter_PathParams(b *testing.B) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := rest.NewApi()

	endpoint := rest.NewEndpointRes[BenchRes]().
		WithPath("/users/:id/posts/:postId").
		WithMethod(http.MethodGet).
		WithHandler(func(ctx context.Context, _ rest.NoRequest) (BenchRes, error) {
			return BenchRes{Message: "found"}, nil
		})

	api.AddEndpoint(endpoint)
	restgin.RegisterRoutes(r, api)

	req := httptest.NewRequest(http.MethodGet, "/users/123/posts/456", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", rec.Code)
		}
	}
}

func BenchmarkEchoAdapter_PathParams(b *testing.B) {
	e := echo.New()
	api := rest.NewApi()

	endpoint := rest.NewEndpointRes[BenchRes]().
		WithPath("/users/:id/posts/:postId").
		WithMethod(http.MethodGet).
		WithHandler(func(ctx context.Context, _ rest.NoRequest) (BenchRes, error) {
			return BenchRes{Message: "found"}, nil
		})

	api.AddEndpoint(endpoint)
	restecho.RegisterRoutes(e, api)

	req := httptest.NewRequest(http.MethodGet, "/users/123/posts/456", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", rec.Code)
		}
	}
}

func BenchmarkChiAdapter_MountPattern(b *testing.B) {
	r := chi.NewRouter()
	r.Get("/native/users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"message":"users"}`))
	})

	api := rest.NewApi()
	metas := []rest.RouteMeta{
		{Method: "GET", Path: "/native/users", Info: rest.RouteInfo{Summary: "List"}},
	}
	restchi.Mount(api, "/api", r, metas)

	req := httptest.NewRequest(http.MethodGet, "/api/native/users", nil)
	mux := api.Mux()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", rec.Code)
		}
	}
}

func BenchmarkGinAdapter_MountPattern(b *testing.B) {
	r := gin.New()
	r.GET("/native/users", func(c *gin.Context) {
		c.JSON(200, BenchRes{Message: "users"})
	})

	api := rest.NewApi()
	metas := []rest.RouteMeta{
		{Method: "GET", Path: "/native/users", Info: rest.RouteInfo{Summary: "List"}},
	}
	restgin.Mount(api, "/api", r, metas)

	req := httptest.NewRequest(http.MethodGet, "/api/native/users", nil)
	mux := api.Mux()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", rec.Code)
		}
	}
}

func BenchmarkEchoAdapter_MountPattern(b *testing.B) {
	e := echo.New()
	e.GET("/native/users", func(c echo.Context) error {
		return c.JSON(200, BenchRes{Message: "users"})
	})

	api := rest.NewApi()
	metas := []rest.RouteMeta{
		{Method: "GET", Path: "/native/users", Info: rest.RouteInfo{Summary: "List"}},
	}
	restecho.Mount(api, "/api", e, metas)

	req := httptest.NewRequest(http.MethodGet, "/api/native/users", nil)
	mux := api.Mux()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", rec.Code)
		}
	}
}

func BenchmarkChiAdapter_SwaggerUI(b *testing.B) {
	r, _ := setupChiAdapter()
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", rec.Code)
		}
	}
}

func BenchmarkGinAdapter_SwaggerUI(b *testing.B) {
	r, _ := setupGinAdapter()
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", rec.Code)
		}
	}
}

func BenchmarkEchoAdapter_SwaggerUI(b *testing.B) {
	e, _ := setupEchoAdapter()
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", rec.Code)
		}
	}
}

func BenchmarkChiAdapter_MultipleEndpoints(b *testing.B) {
	r := chi.NewRouter()
	api := rest.NewApi().
		WithTitle("Bench API").
		WithVersion("1.0.0")

	for i := 0; i < 10; i++ {
		endpoint := rest.NewEndpointRes[BenchRes]().
			WithPath("/endpoint/" + string(rune('a'+i))).
			WithMethod(http.MethodGet).
			WithHandler(func(ctx context.Context, _ rest.NoRequest) (BenchRes, error) {
				return BenchRes{Message: "ok"}, nil
			})
		api.AddEndpoint(endpoint)
	}

	restchi.RegisterRoutes(r, api)

	req := httptest.NewRequest(http.MethodGet, "/endpoint/a", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", rec.Code)
		}
	}
}

func BenchmarkGinAdapter_MultipleEndpoints(b *testing.B) {
	r := gin.New()
	api := rest.NewApi().
		WithTitle("Bench API").
		WithVersion("1.0.0")

	for i := 0; i < 10; i++ {
		endpoint := rest.NewEndpointRes[BenchRes]().
			WithPath("/endpoint/" + string(rune('a'+i))).
			WithMethod(http.MethodGet).
			WithHandler(func(ctx context.Context, _ rest.NoRequest) (BenchRes, error) {
				return BenchRes{Message: "ok"}, nil
			})
		api.AddEndpoint(endpoint)
	}

	restgin.RegisterRoutes(r, api)

	req := httptest.NewRequest(http.MethodGet, "/endpoint/a", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", rec.Code)
		}
	}
}

func BenchmarkEchoAdapter_MultipleEndpoints(b *testing.B) {
	e := echo.New()
	api := rest.NewApi().
		WithTitle("Bench API").
		WithVersion("1.0.0")

	for i := 0; i < 10; i++ {
		endpoint := rest.NewEndpointRes[BenchRes]().
			WithPath("/endpoint/" + string(rune('a'+i))).
			WithMethod(http.MethodGet).
			WithHandler(func(ctx context.Context, _ rest.NoRequest) (BenchRes, error) {
				return BenchRes{Message: "ok"}, nil
			})
		api.AddEndpoint(endpoint)
	}

	restecho.RegisterRoutes(e, api)

	req := httptest.NewRequest(http.MethodGet, "/endpoint/a", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", rec.Code)
		}
	}
}
