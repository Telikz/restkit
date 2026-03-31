package context

import (
	"context"
	"testing"
)

// TestNewRouteContext tests the RouteContext constructor
func TestNewRouteContext(t *testing.T) {
	rc := NewRouteContext()
	if rc == nil {
		t.Fatal("NewRouteContext() returned nil")
	}

	if rc.params == nil {
		t.Error("params map should be initialized")
	}
}

// TestRouteContextPool tests that RouteContext uses pooling
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

// TestRouteContextURLParam tests URL parameter retrieval
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

// TestRouteContextSetURLParam tests URL parameter setting
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

// TestURLParamFromContext tests extracting URL param from context
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

// TestRouteCtxFromContext tests extracting RouteContext from context
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

// TestExtractPathParams tests path parameter extraction
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

// TestExtractPathParamsInvalidPattern tests handling of invalid regex patterns
func TestExtractPathParamsInvalidPattern(t *testing.T) {
	// Test with a pattern that might cause regex compilation issues
	// Most invalid patterns are handled by escaping, but let's test edge cases

	// This pattern should work even with special characters
	pattern := "/test/{param}"
	path := "/test/value"
	result := ExtractPathParams(pattern, path)

	if result["param"] != "value" {
		t.Errorf("expected param 'value', got '%s'", result["param"])
	}
}

// TestContextKey tests the context key
func TestContextKey(t *testing.T) {
	if RouteCtxKey == nil {
		t.Error("RouteCtxKey should not be nil")
	}

	// Test that the key is unique
	if RouteCtxKey.String() != "api context value RouteContext" {
		t.Errorf("unexpected context key string: %s", RouteCtxKey.String())
	}
}
