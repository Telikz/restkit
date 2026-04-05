package yaml

import (
	"net/http"

	"gopkg.in/yaml.v3"
)

// Serializer returns a YAML serializer.
func Serializer() func(w http.ResponseWriter, res any) error {
	return func(w http.ResponseWriter, res any) error {
		w.Header().Set("Content-Type", "application/yaml")
		w.WriteHeader(http.StatusOK)
		encoder := yaml.NewEncoder(w)
		defer encoder.Close()
		return encoder.Encode(res)
	}
}

// Deserializer returns a YAML deserializer.
func Deserializer() func(r *http.Request, req any) error {
	return func(r *http.Request, req any) error {
		return yaml.NewDecoder(r.Body).Decode(req)
	}
}
