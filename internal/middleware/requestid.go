package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// requestIDKey is the context key for request ID
type requestIDKey struct{}

// RequestID returns middleware that injects a unique request ID into each request.
func RequestID(opts ...RequestIDOption) func(http.Handler) http.Handler {
	config := &requestIDConfig{
		headerName: "X-Request-ID",
		generator:  generateRequestID,
		propagate:  true,
	}

	for _, opt := range opts {
		opt(config)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := ""

			// Check for upstream request ID if propagation enabled
			if config.propagate {
				requestID = r.Header.Get(config.headerName)
			}

			// Generate new ID if needed
			if requestID == "" {
				requestID = config.generator()
			}

			// Add to response headers
			w.Header().Set(config.headerName, requestID)

			// Add to context
			ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequestIDFromContext retrieves the request ID from context.
// Returns empty string if not found.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}

// RequestIDOption configures request ID middleware
type RequestIDOption func(*requestIDConfig)

type requestIDConfig struct {
	headerName string
	generator  func() string
	propagate  bool
}

// RequestIDHeader sets the header name for request ID (default: X-Request-ID)
func RequestIDHeader(name string) RequestIDOption {
	return func(c *requestIDConfig) {
		c.headerName = name
	}
}

// RequestIDGenerator sets a custom generator function
func RequestIDGenerator(fn func() string) RequestIDOption {
	return func(c *requestIDConfig) {
		c.generator = fn
	}
}

// RequestIDPropagation controls whether to accept upstream request IDs
// When true (default), existing X-Request-ID headers are preserved
// When false, always generate new request IDs
func RequestIDPropagation(allow bool) RequestIDOption {
	return func(c *requestIDConfig) {
		c.propagate = allow
	}
}

// generateRequestID creates a random 16-byte hex string (32 characters).
// Uses crypto/rand for security.
func generateRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback - this should never happen
		return "req-fallback-1234567890abcdef"
	}
	return hex.EncodeToString(b)
}
