package context

import (
	"context"
	"testing"
)

func TestNewRouteContext(t *testing.T) {
	rc := NewRouteContext()
	if rc == nil {
		t.Fatal("NewRouteContext() returned nil")
	}

	if rc.params == nil {
		t.Error("params map should be initialized")
	}
}

func TestRouteContextPool(t *testing.T) {
	// Get a context from pool
	rc1 := NewRouteContext()
	rc1.SetURLParam("key", "value")

	// Put it back to pool (happens automatically when released)
	// Get another context - it might be the same one from the pool
	rc2 := NewRouteContext()

	// rc2 should have empty params (cleared when put back)
	if val := rc2.URLParam("key"); val != "" {
		t.Errorf("pooled context should have cleared params, got '%s'", val)
	}
}

func TestRouteContextURLParam(t *testing.T) {
	t.Run("get existing param", func(t *testing.T) {
		rc := NewRouteContext()
		rc.SetURLParam("id", "123")

		val := rc.URLParam("id")
		if val != "123" {
			t.Errorf("expected '123', got '%s'", val)
		}
	})

	t.Run("get non-existent param", func(t *testing.T) {
		rc := NewRouteContext()

		val := rc.URLParam("nonexistent")
		if val != "" {
			t.Errorf("expected empty string for non-existent param, got '%s'", val)
		}
	})

	t.Run("nil context", func(t *testing.T) {
		var rc *RouteContext

		val := rc.URLParam("id")
		if val != "" {
			t.Errorf("expected empty string for nil context, got '%s'", val)
		}
	})

	t.Run("nil params map", func(t *testing.T) {
		rc := &RouteContext{params: nil}

		val := rc.URLParam("id")
		if val != "" {
			t.Errorf("expected empty string for nil params, got '%s'", val)
		}
	})
}

func TestRouteContextSetURLParam(t *testing.T) {
	t.Run("set new param", func(t *testing.T) {
		rc := NewRouteContext()
		rc.SetURLParam("user", "john")

		if rc.URLParam("user") != "john" {
			t.Error("failed to set URL param")
		}
	})

	t.Run("set param with nil map", func(t *testing.T) {
		rc := &RouteContext{params: nil}
		rc.SetURLParam("key", "value")

		if rc.URLParam("key") != "value" {
			t.Error("failed to set URL param with nil map")
		}
	})

	t.Run("overwrite existing param", func(t *testing.T) {
		rc := NewRouteContext()
		rc.SetURLParam("id", "123")
		rc.SetURLParam("id", "456")

		if rc.URLParam("id") != "456" {
			t.Error("failed to overwrite URL param")
		}
	})
}

func TestURLParamFromContext(t *testing.T) {
	t.Run("extract from context with RouteContext", func(t *testing.T) {
		rc := NewRouteContext()
		rc.SetURLParam("id", "789")

		ctx := context.WithValue(context.Background(), RouteCtxKey, rc)
		val := URLParam(ctx, "id")

		if val != "789" {
			t.Errorf("expected '789', got '%s'", val)
		}
	})

	t.Run("extract from context without RouteContext", func(t *testing.T) {
		ctx := context.Background()
		val := URLParam(ctx, "id")

		if val != "" {
			t.Errorf("expected empty string, got '%s'", val)
		}
	})

	t.Run("extract with wrong type in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), RouteCtxKey, "not a RouteContext")
		val := URLParam(ctx, "id")

		if val != "" {
			t.Errorf("expected empty string for wrong type, got '%s'", val)
		}
	})
}

func TestRouteCtxFromContext(t *testing.T) {
	t.Run("extract valid RouteContext", func(t *testing.T) {
		rc := NewRouteContext()
		rc.SetURLParam("test", "value")

		ctx := context.WithValue(context.Background(), RouteCtxKey, rc)
		extracted := RouteCtxFromContext(ctx)

		if extracted == nil {
			t.Fatal("expected to extract RouteContext")
		}

		if extracted.URLParam("test") != "value" {
			t.Error("extracted RouteContext has wrong values")
		}
	})

	t.Run("extract from empty context", func(t *testing.T) {
		ctx := context.Background()
		extracted := RouteCtxFromContext(ctx)

		if extracted != nil {
			t.Error("expected nil for empty context")
		}
	})

	t.Run("extract with wrong type", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), RouteCtxKey, 12345)
		extracted := RouteCtxFromContext(ctx)

		if extracted != nil {
			t.Error("expected nil for wrong type")
		}
	})
}

func TestExtractPathParams(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		path     string
		expected map[string]string
	}{
		{
			name:     "single parameter",
			pattern:  "/users/{id}",
			path:     "/users/123",
			expected: map[string]string{"id": "123"},
		},
		{
			name:     "multiple parameters",
			pattern:  "/users/{userId}/posts/{postId}",
			path:     "/users/42/posts/99",
			expected: map[string]string{"userId": "42", "postId": "99"},
		},
		{
			name:     "no parameters",
			pattern:  "/health",
			path:     "/health",
			expected: map[string]string{},
		},
		{
			name:     "no match - wrong path",
			pattern:  "/users/{id}",
			path:     "/products/123",
			expected: map[string]string{},
		},
		{
			name:     "no match - too many segments",
			pattern:  "/users/{id}",
			path:     "/users/123/extra",
			expected: map[string]string{},
		},
		{
			name:     "parameter with hyphen in name",
			pattern:  "/users/{user-id}",
			path:     "/users/abc-123",
			expected: map[string]string{"user-id": "abc-123"},
		},
		{
			name:     "nested path with parameter at end",
			pattern:  "/api/v1/users/{id}",
			path:     "/api/v1/users/456",
			expected: map[string]string{"id": "456"},
		},
		{
			name:     "complex nested parameters",
			pattern:  "/api/{version}/users/{userId}/orders/{orderId}",
			path:     "/api/v2/users/100/orders/500",
			expected: map[string]string{"version": "v2", "userId": "100", "orderId": "500"},
		},
		{
			name:     "regex injection attempt",
			pattern:  "/test/{param}",
			path:     "/test/abc[def]ghi",
			expected: map[string]string{"param": "abc[def]ghi"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractPathParams(tt.pattern, tt.path)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d params, got %d", len(tt.expected), len(result))
				return
			}

			for key, expectedVal := range tt.expected {
				if result[key] != expectedVal {
					t.Errorf("param '%s': expected '%s', got '%s'", key, expectedVal, result[key])
				}
			}
		})
	}
}

func TestExtractPathParamsInvalidPattern(t *testing.T) {
	// Most invalid patterns are handled by escaping, but let's test edge cases

	// This pattern should work even with special characters
	pattern := "/test/{param}"
	path := "/test/value"
	result := ExtractPathParams(pattern, path)

	if result["param"] != "value" {
		t.Errorf("expected param 'value', got '%s'", result["param"])
	}
}

func TestContextKey(t *testing.T) {
	if RouteCtxKey == nil {
		t.Error("RouteCtxKey should not be nil")
	}

	if RouteCtxKey.String() != "api context value RouteContext" {
		t.Errorf("unexpected context key string: %s", RouteCtxKey.String())
	}
}

func TestWithQueries(t *testing.T) {
	queries := map[string]string{"db": "test"}
	ctx := WithQueries(context.Background(), queries)

	retrieved := Queries(ctx)
	if retrieved == nil {
		t.Fatal("expected to retrieve queries from context")
	}

	retrievedMap, ok := retrieved.(map[string]string)
	if !ok {
		t.Fatal("retrieved value is not the expected type")
	}

	if retrievedMap["db"] != "test" {
		t.Errorf("expected db='test', got '%s'", retrievedMap["db"])
	}
}

func TestQueriesFromContext(t *testing.T) {
	t.Run("retrieve existing queries", func(t *testing.T) {
		queries := "test-queries"
		ctx := WithQueries(context.Background(), queries)

		result := Queries(ctx)
		if result != queries {
			t.Errorf("expected '%v', got '%v'", queries, result)
		}
	})

	t.Run("retrieve from empty context", func(t *testing.T) {
		ctx := context.Background()
		result := Queries(ctx)
		if result != nil {
			t.Error("expected nil for empty context")
		}
	})
}

func TestMustQueries(t *testing.T) {
	t.Run("retrieve existing queries", func(t *testing.T) {
		queries := "test-queries"
		ctx := WithQueries(context.Background(), queries)

		result, err := MustQueries(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != queries {
			t.Errorf("expected '%v', got '%v'", queries, result)
		}
	})

	t.Run("error on missing queries", func(t *testing.T) {
		ctx := context.Background()

		_, err := MustQueries(ctx)
		if err == nil {
			t.Error("expected error for missing queries")
		}
	})
}

func TestURLQueryParam(t *testing.T) {
	t.Run("get existing query param", func(t *testing.T) {
		rc := NewRouteContext()
		rc.SetURLQueryParam("page", "1")

		val := rc.URLQueryParam("page")
		if val != "1" {
			t.Errorf("expected '1', got '%s'", val)
		}
	})

	t.Run("get non-existent query param", func(t *testing.T) {
		rc := NewRouteContext()

		val := rc.URLQueryParam("nonexistent")
		if val != "" {
			t.Errorf("expected empty string, got '%s'", val)
		}
	})

	t.Run("nil context", func(t *testing.T) {
		var rc *RouteContext

		val := rc.URLQueryParam("page")
		if val != "" {
			t.Errorf("expected empty string for nil context, got '%s'", val)
		}
	})

	t.Run("nil queryParams map", func(t *testing.T) {
		rc := &RouteContext{queryParams: nil}

		val := rc.URLQueryParam("page")
		if val != "" {
			t.Errorf("expected empty string for nil map, got '%s'", val)
		}
	})
}

func TestSetURLQueryParam(t *testing.T) {
	t.Run("set new query param", func(t *testing.T) {
		rc := NewRouteContext()
		rc.SetURLQueryParam("limit", "10")

		if rc.URLQueryParam("limit") != "10" {
			t.Error("failed to set query param")
		}
	})

	t.Run("set param with nil map", func(t *testing.T) {
		rc := &RouteContext{queryParams: nil}
		rc.SetURLQueryParam("offset", "20")

		if rc.URLQueryParam("offset") != "20" {
			t.Error("failed to set query param with nil map")
		}
	})

	t.Run("overwrite existing param", func(t *testing.T) {
		rc := NewRouteContext()
		rc.SetURLQueryParam("sort", "asc")
		rc.SetURLQueryParam("sort", "desc")

		if rc.URLQueryParam("sort") != "desc" {
			t.Error("failed to overwrite query param")
		}
	})
}

func TestURLQueryParamFromContext(t *testing.T) {
	t.Run("extract query param from context", func(t *testing.T) {
		rc := NewRouteContext()
		rc.SetURLQueryParam("search", "golang")

		ctx := context.WithValue(context.Background(), RouteCtxKey, rc)
		val := URLQueryParam(ctx, "search")

		if val != "golang" {
			t.Errorf("expected 'golang', got '%s'", val)
		}
	})

	t.Run("extract from context without RouteContext", func(t *testing.T) {
		ctx := context.Background()
		val := URLQueryParam(ctx, "search")

		if val != "" {
			t.Errorf("expected empty string, got '%s'", val)
		}
	})
}
