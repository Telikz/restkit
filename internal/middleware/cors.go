package middleware

import (
	"net/http"
	"strconv"
)

// NewCORS creates a CORS middleware with sensible defaults and optional overrides
// defaults:
//   - Methods: GET, POST, PUT, DELETE, OPTIONS, PATCH
//   - Headers: Content-Type, Authorization, Accept, X-Requested-With
//   - Origins: Reflects request origin (safer than wildcard *)
//   - Max-age: 86400 (24 hours)
//   - Credentials: false
func NewCORS(opts ...CORSOption) func(next http.Handler) http.Handler {
	config := &CorsConfig{
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
			"PATCH",
		},
		AllowedHeaders: []string{
			"Content-Type",
			"Authorization",
			"Accept",
			"X-Requested-With",
		},
		MaxAge: 86400,
	}

	for _, opt := range opts {
		opt(config)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			if len(config.AllowedOrigins) == 0 {
				if origin != "" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
				}
			} else {
				for _, allowed := range config.AllowedOrigins {
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
				Set("Access-Control-Allow-Methods", joinStrings(config.AllowedMethods, ", "))

			// Set headers
			w.Header().
				Set("Access-Control-Allow-Headers", joinStrings(config.AllowedHeaders, ", "))

			// Set credentials
			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Set max age
			if config.MaxAge > 0 {
				w.Header().
					Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
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
type CORSOption func(*CorsConfig)

// CorsConfig holds CORS configuration
type CorsConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// WithOrigins sets the allowed origins
func Origins(origins ...string) CORSOption {
	return func(c *CorsConfig) {
		c.AllowedOrigins = origins
	}
}

// WithMethods sets the allowed HTTP methods
func Methods(methods ...string) CORSOption {
	return func(c *CorsConfig) {
		c.AllowedMethods = methods
	}
}

// WithHeaders sets the allowed headers
func Headers(headers ...string) CORSOption {
	return func(c *CorsConfig) {
		c.AllowedHeaders = headers
	}
}

// WithCredentials enables credentials support
func Credentials() CORSOption {
	return func(c *CorsConfig) {
		c.AllowCredentials = true
	}
}

// WithMaxAge sets the max age for preflight cache (in seconds)
func MaxAge(seconds int) CORSOption {
	return func(c *CorsConfig) {
		c.MaxAge = seconds
	}
}
