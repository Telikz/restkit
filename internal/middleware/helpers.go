package middleware

import (
	"encoding/json"
	stderrors "errors"
	"log"
	"net/http"
	"strconv"
	"time"

	routectx "github.com/reststore/restkit/internal/context"
	"github.com/reststore/restkit/internal/errors"
)

// DefaultSerializer is the global default serializer function.
// Can be overridden per-API or per-request via context.
var DefaultSerializer func(w http.ResponseWriter, res any) error

// DefaultDeserializer is the global default deserializer function.
// Can be overridden per-API or per-request via context.
var DefaultDeserializer func(r *http.Request, req any) error

// serializeWithFallback attempts to use context serializer, then global, then default JSON
func serializeWithFallback(w http.ResponseWriter, res any, r *http.Request) error {
	if r != nil {
		if v := r.Context().Value(routectx.SerializerCtxKey); v != nil {
			if serializer, ok := v.(func(http.ResponseWriter, any) error); ok {
				return serializer(w, res)
			}
		}
	}

	if DefaultSerializer != nil {
		return DefaultSerializer(w, res)
	}

	// Default JSON serialization
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(res)
}

// deserializeWithFallback attempts to use context deserializer, then global, then default JSON
func deserializeWithFallback(r *http.Request, req any) error {
	// Check context first (per-API deserializer)
	if v := r.Context().Value(routectx.DeserializerCtxKey); v != nil {
		if deserializer, ok := v.(func(*http.Request, any) error); ok {
			return deserializer(r, req)
		}
	}

	// Fall back to global deserializer
	if DefaultDeserializer != nil {
		return DefaultDeserializer(r, req)
	}

	// Default JSON deserialization
	return json.NewDecoder(r.Body).Decode(req)
}

// JSONBinder creates a bind function for JSON request bodies
func JSONBinder[Req any]() func(r *http.Request) (Req, error) {
	return func(r *http.Request) (Req, error) {
		var req Req
		err := deserializeWithFallback(r, &req)
		return req, err
	}
}

// PathParamBinder creates a bind function that extracts
// the last path segment and converts it to the specified type
func PathParamBinder[T any](
	convert func(string) (T, error),
) func(r *http.Request) (T, error) {
	return func(r *http.Request) (T, error) {
		path := r.URL.Path
		var paramStr string

		for i := len(path) - 1; i >= 0; i-- {
			if path[i] == '/' {
				paramStr = path[i+1:]
				break
			}
		}

		if paramStr == "" {
			var zero T
			return zero,
				stderrors.New(errors.ErrMsgMissingPathParam)
		}

		return convert(paramStr)
	}
}

// StringToInt converts a string to int
func StringToInt(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, stderrors.New(errors.ErrMsgInvalidInteger)
	}
	return n, nil
}

// StringToString is a no-op converter for string path params
func StringToString(s string) (string, error) {
	return s, nil
}

// JSONWriter creates a write function for JSON responses
func JSONWriter[Res any]() func(w http.ResponseWriter, res Res) error {
	return func(w http.ResponseWriter, res Res) error {
		return serializeWithFallback(w, res, nil)
	}
}

// JSONWriterWithRequest creates a write function for JSON responses with request context access
func JSONWriterWithRequest[Res any]() func(w http.ResponseWriter, res Res, r *http.Request) error {
	return func(w http.ResponseWriter, res Res, r *http.Request) error {
		return serializeWithFallback(w, res, r)
	}
}

// JSONErrorWriter writes error responses as JSON in APIError format
func JSONErrorWriter(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("Error handling %s %s: %v", r.Method, r.URL.Path, err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  http.StatusBadRequest,
		"code":    errors.ErrCodeBadRequest,
		"message": err.Error(),
	})
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// RecoveryMiddleware recovers from panics and returns 500 error
func RecoveryMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("Panic in %s %s: %v", r.Method, r.URL.Path, err)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_ = json.NewEncoder(w).Encode(map[string]string{
						"error": "internal server error",
					})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware logs incoming requests with timing
func LoggingMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			log.Printf("[%s] %s completed in %v",
				r.Method, r.URL.Path, time.Since(start))
		})
	}
}
