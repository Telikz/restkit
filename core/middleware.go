package core

import (
	"net/http"

	mw "github.com/reststore/restkit/internal/middleware"
	"github.com/reststore/restkit/internal/serializers"
)

// Serializers provides standard serialization functions for common formats.
var Serializers = struct {
	JSON            func(indent string) func(w http.ResponseWriter, res any) error
	JSONCompact     func() func(w http.ResponseWriter, res any) error
	JSONPretty      func() func(w http.ResponseWriter, res any) error
	JSONDeserialize func() func(r *http.Request, req any) error
	XML             func() func(w http.ResponseWriter, res any) error
	XMLDeserialize  func() func(r *http.Request, req any) error
}{
	JSON:            serializers.JSON,
	JSONCompact:     serializers.JSONCompact,
	JSONPretty:      serializers.JSONPretty,
	JSONDeserialize: serializers.JSONDeserialize,
	XML:             serializers.XML,
	XMLDeserialize:  serializers.XMLDeserialize,
}

// Binders

// JSONBinder creates a bind function for JSON request bodies.
func JSONBinder[Req any]() func(r *http.Request) (Req, error) {
	return mw.JSONBinder[Req]()
}

// QueryBinder creates a bind function that extracts query parameters.
func QueryBinder[Req any]() func(r *http.Request) (Req, error) {
	return mw.QueryBinder[Req]()
}

// MixedBinder creates a bind function that combines path params and JSON body.
func MixedBinder[Req any]() func(r *http.Request) (Req, error) {
	return mw.MixedBinder[Req]()
}

// PathParamBinder creates a bind function that extracts the last path segment.
func PathParamBinder[T any](convert func(string) (T, error)) func(r *http.Request) (T, error) {
	return mw.PathParamBinder(convert)
}

// JSONWriter creates a write function for JSON responses.
func JSONWriter[Res any]() func(w http.ResponseWriter, res Res) error {
	return mw.JSONWriter[Res]()
}

// JSONErrorWriter writes error responses as JSON.
var JSONErrorWriter = mw.JSONErrorWriter

// Middleware

// LoggingMiddleware logs incoming requests with timing.
var LoggingMiddleware = mw.LoggingMiddleware

// RecoveryMiddleware recovers from panics and returns 500 error.
var RecoveryMiddleware = mw.RecoveryMiddleware

// SecurityHeaders returns middleware that adds security headers to all responses.
func SecurityHeaders(opts ...SecurityHeadersOption) func(http.Handler) http.Handler {
	return mw.SecurityHeaders(opts...)
}

// RequestID returns middleware that injects a unique request ID into each request.
func RequestID(opts ...RequestIDOption) func(http.Handler) http.Handler {
	return mw.RequestID(opts...)
}

// RequestIDFromContext retrieves the request ID from context.
var RequestIDFromContext = mw.RequestIDFromContext

// NewCORS creates a CORS middleware with sensible defaults and optional overrides.
func NewCORS(opts ...CORSOption) func(next http.Handler) http.Handler {
	return mw.NewCORS(opts...)
}

// DBMiddleware injects database queries into every request context.
var DBMiddleware = mw.DBMiddleware

// TransactionMiddleware wraps requests in a database transaction.
var TransactionMiddleware = mw.TransactionMiddleware

// Middleware Options

// CORSOption configures CORS middleware behavior.
type CORSOption = mw.CORSOption

// CORSOptions provides option functions for CORS middleware.
var CORSOptions = struct {
	Origins     func(...string) CORSOption
	Methods     func(...string) CORSOption
	Headers     func(...string) CORSOption
	Credentials func() CORSOption
	MaxAge      func(int) CORSOption
}{
	Origins:     mw.Origins,
	Methods:     mw.Methods,
	Headers:     mw.Headers,
	Credentials: mw.Credentials,
	MaxAge:      mw.MaxAge,
}

// SecurityHeadersOption configures security headers middleware.
type SecurityHeadersOption = mw.SecurityHeadersOption

// SecurityHeadersOptions provides option functions for SecurityHeaders middleware.
var SecurityHeadersOptions = struct {
	ContentTypeOptions func(string) SecurityHeadersOption
	FrameOptions       func(string) SecurityHeadersOption
	XSSProtection      func(string) SecurityHeadersOption
	ReferrerPolicy     func(string) SecurityHeadersOption
	CSP                func(string) SecurityHeadersOption
	HSTS               func(string) SecurityHeadersOption
}{
	ContentTypeOptions: mw.ContentTypeOptions,
	FrameOptions:       mw.FrameOptions,
	XSSProtection:      mw.XSSProtection,
	ReferrerPolicy:     mw.ReferrerPolicy,
	CSP:                mw.CSP,
	HSTS:               mw.HSTS,
}

// RequestIDOption configures request ID middleware.
type RequestIDOption = mw.RequestIDOption

// RequestIDOptions provides option functions for RequestID middleware.
var RequestIDOptions = struct {
	Header    func(string) RequestIDOption
	Generator func(func() string) RequestIDOption
	Propagate func(bool) RequestIDOption
}{
	Header:    mw.RequestIDHeader,
	Generator: func(fn func() string) RequestIDOption { return mw.RequestIDGenerator(fn) },
	Propagate: mw.RequestIDPropagation,
}
