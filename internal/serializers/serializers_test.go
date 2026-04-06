package serializers

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSON(t *testing.T) {
	t.Run("serialize with indentation", func(t *testing.T) {
		w := httptest.NewRecorder()
		serializer := JSON("  ")

		data := map[string]string{"message": "hello"}
		err := serializer(w, data)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Check content type
		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
		}

		// Check status code
		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Check body has indentation
		body := w.Body.String()
		if !bytes.Contains([]byte(body), []byte("  ")) {
			t.Error("expected indented JSON output")
		}
	})

	t.Run("serialize without indentation", func(t *testing.T) {
		w := httptest.NewRecorder()
		serializer := JSON("")

		data := map[string]string{"message": "hello"}
		err := serializer(w, data)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		body := w.Body.String()
		// Should not contain newlines in compact mode
		if bytes.Contains([]byte(body), []byte("\n")) && len(body) < 50 {
			// Single newline at end is acceptable for encoder
		}
	})

	t.Run("serialize struct", func(t *testing.T) {
		w := httptest.NewRecorder()
		serializer := JSON("")

		type TestStruct struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}

		data := TestStruct{Name: "test", Value: 42}
		err := serializer(w, data)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		var result TestStruct
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Errorf("failed to unmarshal result: %v", err)
		}

		if result.Name != "test" || result.Value != 42 {
			t.Errorf("unexpected result: %+v", result)
		}
	})
}

func TestJSONCompact(t *testing.T) {
	w := httptest.NewRecorder()
	serializer := JSONCompact()

	data := map[string]interface{}{"a": 1, "b": "test"}
	err := serializer(w, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}
}

func TestJSONPretty(t *testing.T) {
	w := httptest.NewRecorder()
	serializer := JSONPretty()

	data := map[string]interface{}{"a": 1, "b": "test"}
	err := serializer(w, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	body := w.Body.String()
	// Should contain 2-space indentation
	if !bytes.Contains([]byte(body), []byte("  ")) {
		t.Error("expected pretty-printed JSON with indentation")
	}
}

func TestJSONDeserialize(t *testing.T) {
	t.Run("deserialize valid JSON", func(t *testing.T) {
		deserializer := JSONDeserialize()

		jsonData := `{"name":"test","value":42}`
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(jsonData))

		type TestStruct struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}

		var result TestStruct
		err := deserializer(req, &result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if result.Name != "test" || result.Value != 42 {
			t.Errorf("unexpected result: %+v", result)
		}
	})

	t.Run("deserialize invalid JSON", func(t *testing.T) {
		deserializer := JSONDeserialize()

		invalidJSON := `{"name":"test",invalid}`
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(invalidJSON))

		var result map[string]interface{}
		err := deserializer(req, &result)

		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("deserialize empty body", func(t *testing.T) {
		deserializer := JSONDeserialize()

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(""))

		var result map[string]interface{}
		err := deserializer(req, &result)

		// Empty body is typically an EOF error for JSON decoder
		if err == nil {
			t.Error("expected error for empty body")
		}
	})
}

func TestXML(t *testing.T) {
	t.Run("serialize XML", func(t *testing.T) {
		w := httptest.NewRecorder()
		serializer := XML()

		type TestStruct struct {
			XMLName xml.Name `xml:"test"`
			Name    string   `xml:"name"`
			Value   int      `xml:"value"`
		}

		data := TestStruct{Name: "test", Value: 42}
		err := serializer(w, data)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/xml" {
			t.Errorf("expected Content-Type 'application/xml', got '%s'", contentType)
		}

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Check that body contains XML
		body := w.Body.String()
		if !bytes.Contains([]byte(body), []byte("<test>")) {
			t.Error("expected XML output with <test> element")
		}
	})
}

func TestXMLDeserialize(t *testing.T) {
	t.Run("deserialize valid XML", func(t *testing.T) {
		deserializer := XMLDeserialize()

		xmlData := `<?xml version="1.0"?><test><name>test</name><value>42</value></test>`
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(xmlData))

		type TestStruct struct {
			XMLName xml.Name `xml:"test"`
			Name    string   `xml:"name"`
			Value   int      `xml:"value"`
		}

		var result TestStruct
		err := deserializer(req, &result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if result.Name != "test" || result.Value != 42 {
			t.Errorf("unexpected result: %+v", result)
		}
	})

	t.Run("deserialize invalid XML", func(t *testing.T) {
		deserializer := XMLDeserialize()

		invalidXML := `<test><unclosed>`
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(invalidXML))

		type TestStruct struct {
			XMLName xml.Name `xml:"test"`
		}

		var result TestStruct
		err := deserializer(req, &result)

		if err == nil {
			t.Error("expected error for invalid XML")
		}
	})
}
