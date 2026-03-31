package restchi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"

	"github.com/reststore/restkit/internal/endpoints"
	"github.com/reststore/restkit/internal/errors"
	"github.com/reststore/restkit/internal/validation"
)

func validationMiddleware(next http.Handler, reqType any) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodDelete {
			next.ServeHTTP(w, r)
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(errors.NewAPIError(
				http.StatusBadRequest,
				errors.ErrCodeBind,
				"Failed to read request body",
			))
			return
		}
		r.Body.Close()

		reqValue := reflect.New(reflect.TypeOf(reqType)).Interface()
		if err := json.Unmarshal(bodyBytes, reqValue); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(errors.NewAPIError(
				http.StatusBadRequest,
				errors.ErrCodeBind,
				"Failed to parse request body",
			))
			return
		}

		reqElem := reflect.ValueOf(reqValue).Elem().Interface()

		if v, ok := reqElem.(endpoints.ValidatableRequest); ok {
			result := v.Validate(r.Context())
			if result.HasErrors() {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(result.Status)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"status":  result.Status,
					"code":    result.Code,
					"message": result.Message,
					"errors":  result.Errors,
				})
				return
			}
		} else {
			result := validation.ValidateStruct(r.Context(), reqElem)
			if result.HasErrors() {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(result.Status)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"status":  result.Status,
					"code":    result.Code,
					"message": result.Message,
					"errors":  result.Errors,
				})
				return
			}
		}

		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		r.ContentLength = int64(len(bodyBytes))

		next.ServeHTTP(w, r)
	})
}
