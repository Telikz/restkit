package cache

import (
	"net/http"
	"testing"
)

func TestRouteCache_ExactMatch(t *testing.T) {
	cache := NewRouteCache()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	cache.Set("GET", "/users", handler)

	route, found := cache.Get("GET", "/users")
	if !found {
		t.Error("Expected to find exact match route")
	}
	if route.Path != "/users" {
		t.Errorf("Expected path /users, got %s", route.Path)
	}
}

func TestRouteCache_PatternMatch(t *testing.T) {
	cache := NewRouteCache()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	// Set a parameterized route
	cache.Set("GET", "/users/{id}", handler)

	// Should match actual request paths
	testCases := []struct {
		path    string
		found   bool
		pattern string
	}{
		{"/users/123", true, "/users/{id}"},
		{"/users/abc", true, "/users/{id}"},
		{"/users/", false, ""},
		{"/users/123/extra", false, ""},
		{"/products/123", false, ""},
	}

	for _, tc := range testCases {
		route, found := cache.Get("GET", tc.path)
		if found != tc.found {
			t.Errorf("Path %s: expected found=%v, got=%v", tc.path, tc.found, found)
			continue
		}
		if found && route.Path != tc.pattern {
			t.Errorf("Path %s: expected pattern %s, got %s", tc.path, tc.pattern, route.Path)
		}
	}
}

func TestRouteCache_MultipleMethods(t *testing.T) {
	cache := NewRouteCache()
	getHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	postHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	cache.Set("GET", "/users/{id}", getHandler)
	cache.Set("POST", "/users", postHandler)

	// GET /users/123 should find the GET handler
	if _, found := cache.Get("GET", "/users/123"); !found {
		t.Error("Expected to find GET /users/{id} for /users/123")
	}

	// POST /users should find the POST handler
	if _, found := cache.Get("POST", "/users"); !found {
		t.Error("Expected to find POST /users")
	}

	// GET /users (exact) should not match any pattern
	if _, found := cache.Get("GET", "/users"); found {
		t.Error("Should not find GET /users (no exact match or pattern)")
	}
}

func TestRouteCache_MultiplePatternRoutes(t *testing.T) {
	cache := NewRouteCache()
	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	cache.Set("GET", "/users/{id}", handler1)
	cache.Set("GET", "/users/{id}/posts/{postId}", handler2)

	// Should match the more specific pattern
	route, found := cache.Get("GET", "/users/123/posts/456")
	if !found {
		t.Error("Expected to find pattern for /users/123/posts/456")
	}
	if route.Path != "/users/{id}/posts/{postId}" {
		t.Errorf("Expected /users/{id}/posts/{postId}, got %s", route.Path)
	}
}

func TestRouteCache_PerInstanceIsolation(t *testing.T) {
	// This tests that each cache instance is independent
	cache1 := NewRouteCache()
	cache2 := NewRouteCache()

	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	cache1.Set("GET", "/route", handler1)
	cache2.Set("GET", "/route", handler2)

	// Both should find their own routes
	if _, found := cache1.Get("GET", "/route"); !found {
		t.Error("cache1 should find /route")
	}
	if _, found := cache2.Get("GET", "/route"); !found {
		t.Error("cache2 should find /route")
	}
}

func TestMatchPathPattern(t *testing.T) {
	testCases := []struct {
		requestParts []string
		patternParts []string
		shouldMatch  bool
	}{
		// Basic matches
		{[]string{"", "users", "123"}, []string{"", "users", "{id}"}, true},
		{[]string{"", "users", "abc"}, []string{"", "users", "{id}"}, true},

		// Mismatches - empty param value
		{[]string{"", "users", ""}, []string{"", "users", "{id}"}, false},

		// Mismatches
		{[]string{"", "users", "123", "extra"}, []string{"", "users", "{id}"}, false},
		{[]string{"", "users"}, []string{"", "users", "{id}"}, false},
		{[]string{"", "products", "123"}, []string{"", "users", "{id}"}, false},

		// Multiple parameters
		{[]string{"", "users", "123", "posts", "456"}, []string{"", "users", "{userId}", "posts", "{postId}"}, true},
		{[]string{"", "users", "123", "posts"}, []string{"", "users", "{userId}", "posts", "{postId}"}, false},

		// Static segments mixed with params
		{[]string{"", "api", "v1", "users", "123"}, []string{"", "api", "v1", "users", "{id}"}, true},
		{[]string{"", "api", "v2", "users", "123"}, []string{"", "api", "v1", "users", "{id}"}, false},

		// Empty patterns
		{[]string{""}, []string{""}, true},
		{[]string{"", "users"}, []string{""}, false},
	}

	for _, tc := range testCases {
		result := matchPathPattern(tc.requestParts, tc.patternParts)
		if result != tc.shouldMatch {
			t.Errorf("Request %v vs Pattern %v: expected %v, got %v",
				tc.requestParts, tc.patternParts, tc.shouldMatch, result)
		}
	}
}
