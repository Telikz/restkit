package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJSONBinder(t *testing.T) {
	type TestRequest struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	t.Run("successful bind", func(t *testing.T) {
		binder := JSONBinder[TestRequest]()

		body := `{"name":"John","email":"john@example.com"}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		result, err := binder(req)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if result.Name != "John" {
			t.Errorf("expected name 'John', got '%s'", result.Name)
		}

		if result.Email != "john@example.com" {
			t.Errorf("expected email 'john@example.com', got '%s'", result.Email)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		binder := JSONBinder[TestRequest]()

		body := `{"name": "invalid`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(body))

		_, err := binder(req)

		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("empty body", func(t *testing.T) {
		binder := JSONBinder[TestRequest]()

		req := httptest.NewRequest("POST", "/test", strings.NewReader(""))

		_, err := binder(req)

		if err == nil {
			t.Error("expected error for empty body")
		}
	})

	t.Run("type mismatch", func(t *testing.T) {
		type Mismatch struct {
			Count int `json:"count"`
		}

		binder := JSONBinder[Mismatch]()

		body := `{"count": "not a number"}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(body))

		_, err := binder(req)

		if err == nil {
			t.Error("expected error for type mismatch")
		}
	})
}

func TestPathParamBinder(t *testing.T) {
	t.Run("extract string param", func(t *testing.T) {
		binder := PathParamBinder(StringToString)

		req := httptest.NewRequest("GET", "/users/123", nil)

		result, err := binder(req)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if result != "123" {
			t.Errorf("expected '123', got '%s'", result)
		}
	})

	t.Run("extract int param", func(t *testing.T) {
		binder := PathParamBinder(StringToInt)

		req := httptest.NewRequest("GET", "/users/456", nil)

		result, err := binder(req)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if result != 456 {
			t.Errorf("expected 456, got %d", result)
		}
	})

	t.Run("invalid int", func(t *testing.T) {
		binder := PathParamBinder(StringToInt)

		req := httptest.NewRequest("GET", "/users/abc", nil)

		_, err := binder(req)

		if err == nil {
			t.Error("expected error for invalid integer")
		}
	})

	t.Run("empty path", func(t *testing.T) {
		binder := PathParamBinder(StringToString)

		req := httptest.NewRequest("GET", "/", nil)

		_, err := binder(req)

		if err == nil {
			t.Error("expected error for empty path")
		}
	})

	t.Run("path without slash", func(t *testing.T) {
		binder := PathParamBinder(StringToString)

		req := httptest.NewRequest("GET", "/", nil)

		_, err := binder(req)

		if err == nil {
			t.Error("expected error for path with only slash")
		}
	})
}

func TestStringToInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		wantErr  bool
	}{
		{"123", 123, false},
		{"0", 0, false},
		{"-42", -42, false},
		{"abc", 0, true},
		{"", 0, true},
		{"12.34", 0, true},
		{" 123 ", 0, true}, // strconv.Atoi doesn't trim spaces
	}

	for _, tt := range tests {
		result, err := StringToInt(tt.input)

		if tt.wantErr {
			if err == nil {
				t.Errorf("StringToInt(%q) expected error", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("StringToInt(%q) unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("StringToInt(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		}
	}
}

func TestStringToString(t *testing.T) {
	tests := []string{"hello", "world", "123", "", "with spaces"}

	for _, input := range tests {
		result, err := StringToString(input)

		if err != nil {
			t.Errorf("StringToString(%q) unexpected error: %v", input, err)
		}

		if result != input {
			t.Errorf("StringToString(%q) = %q", input, result)
		}
	}
}

func TestJSONWriter(t *testing.T) {
	type TestResponse struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	t.Run("successful write", func(t *testing.T) {
		writer := JSONWriter[TestResponse]()

		rec := httptest.NewRecorder()
		response := TestResponse{ID: 1, Name: "Test"}

		err := writer(rec, response)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		contentType := rec.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
		}

		body := rec.Body.String()
		if !strings.Contains(body, `"id":1`) {
			t.Errorf("expected body to contain id, got '%s'", body)
		}
	})

	t.Run("nil write", func(t *testing.T) {
		writer := JSONWriter[*TestResponse]()

		rec := httptest.NewRecorder()
		var response *TestResponse

		err := writer(rec, response)

		if err != nil {
			t.Errorf("expected no error for nil, got %v", err)
		}
	})
}

func TestJSONErrorWriter(t *testing.T) {
	t.Run("writes error response", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		testErr := errors.New("something went wrong")

		JSONErrorWriter(rec, req, testErr)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}

		contentType := rec.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "something went wrong") {
			t.Errorf("expected body to contain error message, got '%s'", body)
		}

		if !strings.Contains(body, "bad_request") {
			t.Errorf("expected body to contain error code, got '%s'", body)
		}
	})
}

func TestJoinStrings(t *testing.T) {
	tests := []struct {
		strs     []string
		sep      string
		expected string
	}{
		{[]string{"a", "b", "c"}, ",", "a,b,c"},
		{[]string{"x"}, "-", "x"},
		{[]string{}, ",", ""},
		{[]string{"hello", "world"}, " ", "hello world"},
		{[]string{"GET", "POST", "PUT"}, ", ", "GET, POST, PUT"},
	}

	for _, tt := range tests {
		result := joinStrings(tt.strs, tt.sep)
		if result != tt.expected {
			t.Errorf("joinStrings(%v, %q) = %q, want %q", tt.strs, tt.sep, result, tt.expected)
		}
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	t.Run("recovers from panic", func(t *testing.T) {
		middleware := RecoveryMiddleware()

		panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("something went terribly wrong")
		})

		handler := middleware(panicHandler)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}

		contentType := rec.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "internal server error") {
			t.Errorf("expected body to contain error message, got '%s'", body)
		}
	})

	t.Run("no panic - normal flow", func(t *testing.T) {
		middleware := RecoveryMiddleware()

		normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("success"))
		})

		handler := middleware(normalHandler)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		if rec.Body.String() != "success" {
			t.Errorf("expected body 'success', got '%s'", rec.Body.String())
		}
	})
}

func TestLoggingMiddleware(t *testing.T) {
	middleware := LoggingMiddleware()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(testHandler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/users", nil)

	// Just verify it doesn't panic and passes through
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestNewCORS(t *testing.T) {
	t.Run("default CORS headers", func(t *testing.T) {
		middleware := NewCORS()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://example.com")

		handler.ServeHTTP(rec, req)

		// Check that CORS headers are set
		allowOrigin := rec.Header().Get("Access-Control-Allow-Origin")
		if allowOrigin == "" {
			t.Error("expected Access-Control-Allow-Origin header")
		}

		allowMethods := rec.Header().Get("Access-Control-Allow-Methods")
		if allowMethods == "" {
			t.Error("expected Access-Control-Allow-Methods header")
		}

		allowHeaders := rec.Header().Get("Access-Control-Allow-Headers")
		if allowHeaders == "" {
			t.Error("expected Access-Control-Allow-Headers header")
		}

		maxAge := rec.Header().Get("Access-Control-Max-Age")
		if maxAge == "" {
			t.Error("expected Access-Control-Max-Age header")
		}
	})

	t.Run("preflight request", func(t *testing.T) {
		middleware := NewCORS()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("handler should not be called for preflight")
		}))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://example.com")

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200 for preflight, got %d", rec.Code)
		}
	})

	t.Run("with specific origins", func(t *testing.T) {
		middleware := NewCORS(WithOrigins("http://localhost:3000"))

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")

		handler.ServeHTTP(rec, req)

		allowOrigin := rec.Header().Get("Access-Control-Allow-Origin")
		if allowOrigin != "http://localhost:3000" {
			t.Errorf("expected origin 'http://localhost:3000', got '%s'", allowOrigin)
		}
	})

	t.Run("with wildcard origin", func(t *testing.T) {
		middleware := NewCORS(WithOrigins("*"))

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)

		handler.ServeHTTP(rec, req)

		allowOrigin := rec.Header().Get("Access-Control-Allow-Origin")
		if allowOrigin != "*" {
			t.Errorf("expected wildcard origin, got '%s'", allowOrigin)
		}

		// Vary header should not be set for wildcard
		vary := rec.Header().Get("Vary")
		if vary == "Origin" {
			t.Error("Vary header should not be set for wildcard origin")
		}
	})

	t.Run("with custom methods", func(t *testing.T) {
		middleware := NewCORS(WithMethods("GET", "POST"))

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)

		handler.ServeHTTP(rec, req)

		allowMethods := rec.Header().Get("Access-Control-Allow-Methods")
		if !strings.Contains(allowMethods, "GET") || !strings.Contains(allowMethods, "POST") {
			t.Errorf("expected methods to contain GET and POST, got '%s'", allowMethods)
		}
	})

	t.Run("with custom headers", func(t *testing.T) {
		middleware := NewCORS(WithHeaders("X-Custom-Header"))

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)

		handler.ServeHTTP(rec, req)

		allowHeaders := rec.Header().Get("Access-Control-Allow-Headers")
		if !strings.Contains(allowHeaders, "X-Custom-Header") {
			t.Errorf("expected headers to contain X-Custom-Header, got '%s'", allowHeaders)
		}
	})

	t.Run("with credentials", func(t *testing.T) {
		middleware := NewCORS(WithCredentials())

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)

		handler.ServeHTTP(rec, req)

		allowCredentials := rec.Header().Get("Access-Control-Allow-Credentials")
		if allowCredentials != "true" {
			t.Errorf("expected credentials 'true', got '%s'", allowCredentials)
		}
	})

	t.Run("with custom max age", func(t *testing.T) {
		middleware := NewCORS(WithMaxAge(3600))

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)

		handler.ServeHTTP(rec, req)

		maxAge := rec.Header().Get("Access-Control-Max-Age")
		if maxAge != "3600" {
			t.Errorf("expected max-age '3600', got '%s'", maxAge)
		}
	})
}
