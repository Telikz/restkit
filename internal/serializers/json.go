package serializers

import (
	"encoding/json"
	"net/http"
)

// JSON returns a JSON serializer with optional indentation.
func JSON(indent string) func(w http.ResponseWriter, res any) error {
	return func(w http.ResponseWriter, res any) error {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", indent)
		return encoder.Encode(res)
	}
}

// JSONCompact returns a compact JSON serializer (no whitespace).
func JSONCompact() func(w http.ResponseWriter, res any) error {
	return JSON("")
}

// JSONPretty returns a pretty-printed JSON serializer with 2-space indentation.
func JSONPretty() func(w http.ResponseWriter, res any) error {
	return JSON("  ")
}

// JSONDeserialize returns a standard JSON deserializer.
func JSONDeserialize() func(r *http.Request, req any) error {
	return func(r *http.Request, req any) error {
		return json.NewDecoder(r.Body).Decode(req)
	}
}
