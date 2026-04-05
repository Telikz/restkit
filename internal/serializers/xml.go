package serializers

import (
	"encoding/xml"
	"net/http"
)

// XML returns an XML serializer.
func XML() func(w http.ResponseWriter, res any) error {
	return func(w http.ResponseWriter, res any) error {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		return xml.NewEncoder(w).Encode(res)
	}
}

// XMLDeserialize returns an XML deserializer.
func XMLDeserialize() func(r *http.Request, req any) error {
	return func(r *http.Request, req any) error {
		return xml.NewDecoder(r.Body).Decode(req)
	}
}
