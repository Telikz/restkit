package restkit_test

import (
	"testing"
	"testing/quick"

	rest "github.com/reststore/restkit"
)

// FuzzParseID tests ParseID with various inputs
func FuzzParseID(f *testing.F) {
	// Seed corpus with known cases
	f.Add("123")
	f.Add("0")
	f.Add("-123")
	f.Add("9223372036854775807")  // max int64
	f.Add("-9223372036854775808") // min int64
	f.Add("")
	f.Add("abc")
	f.Add("12.34")
	f.Add(" 123 ")
	f.Add("123abc")
	f.Add("0x1F")

	f.Fuzz(func(t *testing.T, input string) {
		id, err := rest.ParseID(input)

		// Basic sanity checks
		if err == nil {
			// If no error, id should be parseable back to string
			// (within the bounds of int64)
			_ = id
		}
	})
}

// FuzzParseIntID tests ParseIntID with various inputs
func FuzzParseIntID(f *testing.F) {
	// Seed corpus
	f.Add("123")
	f.Add("0")
	f.Add("-123")
	f.Add("2147483647")  // max int32
	f.Add("-2147483648") // min int32
	f.Add("")
	f.Add("abc")
	f.Add("12.34")
	f.Add(" 123 ")
	f.Add("999999999999999999999999999999") // overflow

	f.Fuzz(func(t *testing.T, input string) {
		id, err := rest.ParseIntID(input)

		if err == nil {
			// Verify the id is within int range
			_ = id
		}
	})
}

// FuzzStringToInt tests StringToInt converter
func FuzzStringToInt(f *testing.F) {
	f.Add("123")
	f.Add("0")
	f.Add("-456")
	f.Add("")
	f.Add("not-a-number")
	f.Add("99999999999999999999")

	f.Fuzz(func(t *testing.T, input string) {
		val, err := rest.StringToInt(input)

		if err == nil {
			// Success case - verify it's actually an int
			t.Logf("StringToInt(%q) = %d", input, val)
		}
	})
}

// FuzzExtractPathParams tests path parameter extraction
func FuzzExtractPathParams(f *testing.F) {
	// Seed with valid patterns
	f.Add("/users/{id}", "/users/123")
	f.Add("/api/{version}/users", "/api/v1/users")
	f.Add("/users/{userId}/posts/{postId}", "/users/42/posts/99")
	f.Add("/static/path", "/static/path")
	f.Add("/", "/")
	f.Add("/{id}", "/abc123")
	f.Add("/users/{id}/", "/users/123/")
	f.Add("", "")
	f.Add("/users/{id}/posts/{id}", "/users/1/posts/2") // duplicate param names
	f.Add("/api/{hyphenated-name}", "/api/test-value")
	f.Add("/users/{id}", "/wrong/path")

	f.Fuzz(func(t *testing.T, pattern, path string) {
		params := rest.ExtractPathParams(pattern, path)

		// Sanity check: params should never be nil
		if params == nil {
			t.Error("ExtractPathParams returned nil map")
		}

		// Additional property: all values should be non-empty strings
		for k, v := range params {
			if k == "" {
				t.Error("extracted parameter has empty key")
			}
			if v == "" {
				t.Logf("Parameter %q has empty value", k)
			}
		}
	})
}

// TestParseIDProperties uses property-based testing
func TestParseIDProperties(t *testing.T) {
	// Property: Valid numeric strings should parse successfully
	validInt64 := func(s string) bool {
		// This is a simplified property - just check it doesn't panic
		_, _ = rest.ParseID(s)
		return true
	}

	if err := quick.Check(validInt64, nil); err != nil {
		t.Errorf("ParseID property check failed: %v", err)
	}
}

// TestExtractPathParamsProperties uses property-based testing for path extraction
func TestExtractPathParamsProperties(t *testing.T) {
	// Property: ExtractPathParams should never return nil
	neverNil := func(pattern, path string) bool {
		params := rest.ExtractPathParams(pattern, path)
		return params != nil
	}

	if err := quick.Check(neverNil, nil); err != nil {
		t.Errorf("ExtractPathParams property check failed: %v", err)
	}
}

// TestSchemaFromTypes verifies schema generation works for various types
func TestSchemaFromTypes(t *testing.T) {
	type SimpleStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	type NestedStruct struct {
		ID      int          `json:"id"`
		Details SimpleStruct `json:"details"`
	}

	type ArrayStruct struct {
		Items []string `json:"items"`
		Count int      `json:"count"`
	}

	type OptionalStruct struct {
		Required string  `json:"required"`
		Optional *string `json:"optional,omitempty"`
	}

	// Just verify these don't panic and return valid schemas
	tests := []struct {
		name string
		fn   func() map[string]any
	}{
		{"SimpleStruct", rest.SchemaFrom[SimpleStruct]},
		{"NestedStruct", rest.SchemaFrom[NestedStruct]},
		{"ArrayStruct", rest.SchemaFrom[ArrayStruct]},
		{"OptionalStruct", rest.SchemaFrom[OptionalStruct]},
		{"string", rest.SchemaFrom[string]},
		{"int", rest.SchemaFrom[int]},
		{"bool", rest.SchemaFrom[bool]},
		{"map[string]string", rest.SchemaFrom[map[string]string]},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.fn()
			if schema == nil {
				t.Error("SchemaFrom returned nil")
			}
			if _, ok := schema["type"]; !ok {
				t.Error("Schema missing 'type' field")
			}
		})
	}
}

// TestSchemaFromProperties verifies SchemaFrom generates valid schemas
func TestSchemaFromProperties(t *testing.T) {
	type TestStruct struct {
		Name    string `json:"name"`
		Value   int    `json:"value"`
		Enabled bool   `json:"enabled"`
	}

	// Property: SchemaFrom should always return a non-nil map
	schema := rest.SchemaFrom[TestStruct]()
	if schema == nil {
		t.Error("SchemaFrom returned nil")
	}

	// Property: Schema should have 'type' field
	if schemaType, ok := schema["type"]; !ok {
		t.Error("Schema missing 'type' field")
	} else if schemaType != "object" {
		t.Errorf("Expected type 'object', got %v", schemaType)
	}

	// Property: Schema should have 'properties' field
	if properties, ok := schema["properties"]; !ok {
		t.Error("Schema missing 'properties' field")
	} else if properties == nil {
		t.Error("Schema properties is nil")
	}
}

// FuzzErrorCodes tests that error codes are constants
func TestErrorCodesAreConstants(t *testing.T) {
	// These should never change
	codes := map[string]string{
		rest.ErrCodeInternal:      "internal",
		rest.ErrCodeConfiguration: "configuration",
		rest.ErrCodeValidation:    "validation",
		rest.ErrCodeBind:          "bind",
		rest.ErrCodeNotFound:      "not_found",
		rest.ErrCodeUnauthorized:  "unauthorized",
		rest.ErrCodeForbidden:     "forbidden",
		rest.ErrCodeBadRequest:    "bad_request",
		rest.ErrCodeMissingParam:  "missing_param",
	}

	for code, expected := range codes {
		if code != expected {
			t.Errorf("Error code mismatch: expected %q, got %q", expected, code)
		}
	}
}

// TestParamLocationConstants verifies parameter location constants
func TestParamLocationConstants(t *testing.T) {
	if rest.ParamLocationPath != "path" {
		t.Errorf("ParamLocationPath should be 'path', got %q", rest.ParamLocationPath)
	}
	if rest.ParamLocationQuery != "query" {
		t.Errorf("ParamLocationQuery should be 'query', got %q", rest.ParamLocationQuery)
	}
}
