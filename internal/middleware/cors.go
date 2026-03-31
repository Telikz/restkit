package middleware

import (
	"net/http"
	"strconv"
)

// NewCORS creates a CORS middleware with sensible defaults and optional overrides
// Sensible defaults:
//   - Methods: GET, POST, PUT, DELETE, OPTIONS, PATCH
//   - Headers: Content-Type, Authorization, Accept, X-Requested-With
//   - Origins: Reflects request origin (safer than wildcard *)
//   - Max-age: 86400 (24 hours)
//   - Credentials: false
func NewCORS(opts ...CORSOption) func(next http.Handler) http.Handler {
	config := &corsConfig{
		allowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
			"PATCH",
		},
		allowedHeaders: []string{
			"Content-Type",
			"Authorization",
			"Accept",
			"X-Requested-With",
		},
		maxAge: 86400,
	}

	for _, opt := range opts {
		opt(config)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			if len(config.allowedOrigins) == 0 {
				if origin != "" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
				}
			} else {
				for _, allowed := range config.allowedOrigins {
					if allowed == "*" || allowed == origin {
						w.Header().Set("Access-Control-Allow-Origin", allowed)
						if allowed != "*" {
							w.Header().Set("Vary", "Origin")
						}
						break
					}
				}
			}

			// Set methods
			w.Header().
				Set("Access-Control-Allow-Methods", joinStrings(config.allowedMethods, ", "))

			// Set headers
			w.Header().
				Set("Access-Control-Allow-Headers", joinStrings(config.allowedHeaders, ", "))

			// Set credentials
			if config.allowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Set max age
			if config.maxAge > 0 {
				w.Header().
					Set("Access-Control-Max-Age", strconv.Itoa(config.maxAge))
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// CORSOption configures CORS middleware behavior
type CORSOption func(*corsConfig)

// corsConfig holds CORS configuration
type corsConfig struct {
	allowedOrigins   []string
	allowedMethods   []string
	allowedHeaders   []string
	allowCredentials bool
	maxAge           int
}

// WithOrigins sets the allowed origins
func WithOrigins(origins ...string) CORSOption {
	return func(c *corsConfig) {
		c.allowedOrigins = origins
	}
}

// WithMethods sets the allowed HTTP methods
func WithMethods(methods ...string) CORSOption {
	return func(c *corsConfig) {
		c.allowedMethods = methods
	}
}

// WithHeaders sets the allowed headers
func WithHeaders(headers ...string) CORSOption {
	return func(c *corsConfig) {
		c.allowedHeaders = headers
	}
}

// WithCredentials enables credentials support
func WithCredentials() CORSOption {
	return func(c *corsConfig) {
		c.allowCredentials = true
	}
}

// WithMaxAge sets the max age for preflight cache (in seconds)
func WithMaxAge(seconds int) CORSOption {
	return func(c *corsConfig) {
		c.maxAge = seconds
	}
}
