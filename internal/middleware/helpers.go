package middleware

import (
	"encoding/json"
	stderrors "errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/telikz/restkit/internal/errors"
)

// JSONBinder creates a bind function for JSON request bodies
func JSONBinder[Req any]() func(r *http.Request) (Req, error) {
	return func(r *http.Request) (Req, error) {
		var req Req
		err := json.NewDecoder(r.Body).Decode(&req)
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		return json.NewEncoder(w).Encode(res)
	}
}

// JSONErrorWriter writes error responses as JSON in APIError format
func JSONErrorWriter(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("Error handling %s %s: %v", r.Method, r.URL.Path, err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
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
					json.NewEncoder(w).Encode(map[string]string{
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
