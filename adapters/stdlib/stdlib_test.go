package reststdlib

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	rk "github.com/reststore/restkit"
)

// Test mounting stdlib routes with a prefix
func TestMountWithPrefix(t *testing.T) {
	stdlibMux := http.NewServeMux()
	stdlibMux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "User ID: %s", r.PathValue("id"))
	})

	api := rk.NewApi()
	meta := []rk.RouteMeta{
		{
			Method: "GET",
			Path:   "/users/{id}",
			Info: rk.RouteInfo{
				Summary: "Get user",
				ResponseType: struct {
					ID string `json:"id"`
				}{},
			},
		},
	}

	err := Mount(api, "/v1", stdlibMux, meta)
	if err != nil {
		t.Fatalf("Mount failed: %v", err)
	}

	handler := api.Mux()
	req := httptest.NewRequest("GET", "/v1/users/123", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
	expected := "User ID: 123"
	if rr.Body.String() != expected {
		t.Errorf("Expected body %q, got %q", expected, rr.Body.String())
	}
}

// Test mounting stdlib routes without a prefix
func TestMountWithoutPrefix(t *testing.T) {
	stdlibMux := http.NewServeMux()
	stdlibMux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "User ID: %s", r.PathValue("id"))
	})

	api := rk.NewApi()
	meta := []rk.RouteMeta{
		{
			Method: "GET",
			Path:   "/users/{id}",
			Info: rk.RouteInfo{
				Summary: "Get user",
				ResponseType: struct {
					ID string `json:"id"`
				}{},
			},
		},
	}

	err := Mount(api, "", stdlibMux, meta)
	if err != nil {
		t.Fatalf("Mount failed: %v", err)
	}

	handler := api.Mux()
	req := httptest.NewRequest("GET", "/users/123", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
	expected := "User ID: 123"
	if rr.Body.String() != expected {
		t.Errorf("Expected body %q, got %q", expected, rr.Body.String())
	}
}

// Test mounting stdlib routes with a nested prefix
func TestMountWithNestedPrefix(t *testing.T) {
	stdlibMux := http.NewServeMux()
	stdlibMux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "User ID: %s", r.PathValue("id"))
	})

	api := rk.NewApi()
	meta := []rk.RouteMeta{
		{
			Method: "GET",
			Path:   "/users/{id}",
			Info: rk.RouteInfo{
				Summary: "Get user",
				ResponseType: struct {
					ID string `json:"id"`
				}{},
			},
		},
	}

	err := Mount(api, "/api/v1", stdlibMux, meta)
	if err != nil {
		t.Fatalf("Mount failed: %v", err)
	}

	handler := api.Mux()
	req := httptest.NewRequest("GET", "/api/v1/users/123", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
	expected := "User ID: 123"
	if rr.Body.String() != expected {
		t.Errorf("Expected body %q, got %q", expected, rr.Body.String())
	}
}

// Test mounting with native endpoints (mimics the example setup)
func TestMountWithNativeEndpoints(t *testing.T) {
	stdlibMux := http.NewServeMux()
	stdlibMux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"id":%s,"source":"stdlib"}`, r.PathValue("id"))
	})
	stdlibMux.HandleFunc("GET /users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"id":1,"source":"stdlib"}]`)
	})

	api := rk.NewApi()
	meta := []rk.RouteMeta{
		{
			Method: "GET",
			Path:   "/users/{id}",
			Info: rk.RouteInfo{
				Summary: "Get user (stdlib)",
				ResponseType: struct {
					ID     int    `json:"id"`
					Source string `json:"source"`
				}{},
			},
		},
		{
			Method: "GET",
			Path:   "/users",
			Info: rk.RouteInfo{
				Summary:      "List users (stdlib)",
				ResponseType: []struct{}{},
			},
		},
	}

	err := Mount(api, "/v1", stdlibMux, meta)
	if err != nil {
		t.Fatalf("Mount failed: %v", err)
	}

	api.AddGroup(rk.NewGroup("/v2/users").
		WithTitle("User Management (native)").
		WithEndpoints(
			rk.Get("/{id}", func(ctx context.Context, req rk.GetRequest) (map[string]any, error) {
				return map[string]any{
					"id":     req.ID,
					"source": "native",
				}, nil
			}),
		),
	)

	handler := api.Mux()

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "v1 mounted route - get user",
			method:     "GET",
			path:       "/v1/users/123",
			wantStatus: http.StatusOK,
			wantBody:   `{"id":123,"source":"stdlib"}`,
		},
		{
			name:       "v1 mounted route - list users",
			method:     "GET",
			path:       "/v1/users",
			wantStatus: http.StatusOK,
			wantBody:   `[{"id":1,"source":"stdlib"}]`,
		},
		{
			name:       "v2 native endpoint - get user",
			method:     "GET",
			path:       "/v2/users/456",
			wantStatus: http.StatusOK,
			wantBody:   "{\"id\":456,\"source\":\"native\"}\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("Status = %d, want %d", rr.Code, tc.wantStatus)
			}
			if rr.Body.String() != tc.wantBody {
				t.Errorf("Body = %q, want %q", rr.Body.String(), tc.wantBody)
			}
		})
	}
}

// Test with a shared stdlibMux (like the example does)
func TestMountWithSharedMux(t *testing.T) {
	stdlibMux := http.NewServeMux()
	stdlibMux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"id":%s,"source":"shared"}`, r.PathValue("id"))
	})

	// Simulate first server at :8080 (pure stdlib)
	ts1 := httptest.NewServer(stdlibMux)
	defer ts1.Close()

	// Verify first server works
	resp, err := http.Get(ts1.URL + "/users/123")
	if err != nil {
		t.Fatalf("Failed to contact first server: %v", err)
	}
	resp.Body.Close()

	// Create RestKit API using the SAME stdlibMux
	api := rk.NewApi()
	meta := []rk.RouteMeta{
		{
			Method: "GET",
			Path:   "/users/{id}",
			Info: rk.RouteInfo{
				Summary: "Get user",
				ResponseType: struct {
					ID int `json:"id"`
				}{},
			},
		},
	}

	err = Mount(api, "/v1", stdlibMux, meta)
	if err != nil {
		t.Fatalf("Mount failed: %v", err)
	}

	api.AddGroup(rk.NewGroup("/v2/users").
		WithEndpoints(
			rk.Get("/{id}", func(ctx context.Context, req rk.GetRequest) (map[string]any, error) {
				return map[string]any{"id": req.ID, "source": "native"}, nil
			}),
		),
	)

	ts2 := httptest.NewServer(api.Mux())
	defer ts2.Close()

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "v1 mounted route with shared mux",
			path:       "/v1/users/123",
			wantStatus: http.StatusOK,
			wantBody:   `{"id":123,"source":"shared"}`,
		},
		{
			name:       "v2 native endpoint",
			path:       "/v2/users/456",
			wantStatus: http.StatusOK,
			wantBody:   `{"id":456,"source":"native"}` + "\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(ts2.URL + tc.path)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.wantStatus {
				t.Errorf("Status = %d, want %d", resp.StatusCode, tc.wantStatus)
			}

			body := make([]byte, 1024)
			n, _ := resp.Body.Read(body)
			gotBody := string(body[:n])

			if gotBody != tc.wantBody {
				t.Errorf("Body = %q, want %q", gotBody, tc.wantBody)
			}
		})
	}
}
