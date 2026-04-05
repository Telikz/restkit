package middleware

import "net/http"

// SecurityHeaders returns middleware that adds security headers to all responses.
//
// Default headers:
//   - X-Content-Type-Options: nosniff
//   - X-Frame-Options: DENY
//   - X-XSS-Protection: 1; mode=block
//   - Referrer-Policy: strict-origin-when-cross-origin
//   - Content-Security-Policy: default-src 'self'
//   - Strict-Transport-Security: max-age=31536000; includeSubDomains
func SecurityHeaders(opts ...SecurityHeadersOption) func(http.Handler) http.Handler {
	config := &securityHeadersConfig{
		contentTypeOptions: "nosniff",
		frameOptions:       "DENY",
		xssProtection:      "1; mode=block",
		referrerPolicy:     "strict-origin-when-cross-origin",
		csp:                "default-src 'self'",
		hsts:               "max-age=31536000; includeSubDomains",
	}

	for _, opt := range opts {
		opt(config)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", config.contentTypeOptions)
			w.Header().Set("X-Frame-Options", config.frameOptions)
			w.Header().Set("X-XSS-Protection", config.xssProtection)
			w.Header().Set("Referrer-Policy", config.referrerPolicy)
			w.Header().Set("Content-Security-Policy", config.csp)
			w.Header().Set("Strict-Transport-Security", config.hsts)

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeadersOption configures security headers middleware
type SecurityHeadersOption func(*securityHeadersConfig)

type securityHeadersConfig struct {
	contentTypeOptions string
	frameOptions       string
	xssProtection      string
	referrerPolicy     string
	csp                string
	hsts               string
}

// ContentTypeOptions sets X-Content-Type-Options header
func ContentTypeOptions(value string) SecurityHeadersOption {
	return func(c *securityHeadersConfig) {
		c.contentTypeOptions = value
	}
}

// FrameOptions sets X-Frame-Options header
func FrameOptions(value string) SecurityHeadersOption {
	return func(c *securityHeadersConfig) {
		c.frameOptions = value
	}
}

// XSSProtection sets X-XSS-Protection header
func XSSProtection(value string) SecurityHeadersOption {
	return func(c *securityHeadersConfig) {
		c.xssProtection = value
	}
}

// ReferrerPolicy sets Referrer-Policy header
func ReferrerPolicy(value string) SecurityHeadersOption {
	return func(c *securityHeadersConfig) {
		c.referrerPolicy = value
	}
}

// CSP sets Content-Security-Policy header
func CSP(value string) SecurityHeadersOption {
	return func(c *securityHeadersConfig) {
		c.csp = value
	}
}

// HSTS sets Strict-Transport-Security header
func HSTS(value string) SecurityHeadersOption {
	return func(c *securityHeadersConfig) {
		c.hsts = value
	}
}
