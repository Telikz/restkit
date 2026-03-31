package schema

import (
	"reflect"
	"testing"
	"time"
)

// TestSchemaFrom tests the generic schema generator
func TestSchemaFrom(t *testing.T) {
	t.Run("string type", func(t *testing.T) {
		schema := SchemaFrom[string]()
		if schema["type"] != "string" {
			t.Errorf("expected type 'string', got '%v'", schema["type"])
		}
	})

	t.Run("int type", func(t *testing.T) {
		schema := SchemaFrom[int]()
		if schema["type"] != "integer" {
			t.Errorf("expected type 'integer', got '%v'", schema["type"])
		}
	})

	t.Run("int64 type", func(t *testing.T) {
		schema := SchemaFrom[int64]()
		if schema["type"] != "integer" {
			t.Errorf("expected type 'integer', got '%v'", schema["type"])
		}
	})

	t.Run("uint type", func(t *testing.T) {
		schema := SchemaFrom[uint]()
		if schema["type"] != "integer" {
			t.Errorf("expected type 'integer', got '%v'", schema["type"])
		}
	})

	t.Run("float64 type", func(t *testing.T) {
		schema := SchemaFrom[float64]()
		if schema["type"] != "number" {
			t.Errorf("expected type 'number', got '%v'", schema["type"])
		}
	})

	t.Run("float32 type", func(t *testing.T) {
		schema := SchemaFrom[float32]()
		if schema["type"] != "number" {
			t.Errorf("expected type 'number', got '%v'", schema["type"])
		}
	})

	t.Run("bool type", func(t *testing.T) {
		schema := SchemaFrom[bool]()
		if schema["type"] != "boolean" {
			t.Errorf("expected type 'boolean', got '%v'", schema["type"])
		}
	})
}

// TestTypeToSchema tests the type to schema conversion
func TestTypeToSchema(t *testing.T) {
	t.Run("nil type", func(t *testing.T) {
		schema := TypeToSchema(nil)
		if schema["type"] != "null" {
			t.Errorf("expected type 'null', got '%v'", schema["type"])
		}
	})

	t.Run("pointer type", func(t *testing.T) {
		schema := TypeToSchema(reflect.TypeOf((*string)(nil)))
		if schema["type"] != "string" {
			t.Errorf("expected pointer to resolve to 'string', got '%v'", schema["type"])
		}
	})

	t.Run("slice type", func(t *testing.T) {
		type StringSlice []string
		schema := TypeToSchema(reflect.TypeOf(StringSlice{}))

		if schema["type"] != "array" {
			t.Errorf("expected type 'array', got '%v'", schema["type"])
		}

		items, ok := schema["items"].(map[string]any)
		if !ok {
			t.Fatal("items should be a map")
		}

		if items["type"] != "string" {
			t.Errorf("expected items type 'string', got '%v'", items["type"])
		}
	})

	t.Run("array type", func(t *testing.T) {
		schema := TypeToSchema(reflect.TypeOf([3]int{}))

		if schema["type"] != "array" {
			t.Errorf("expected type 'array', got '%v'", schema["type"])
		}

		items, ok := schema["items"].(map[string]any)
		if !ok {
			t.Fatal("items should be a map")
		}

		if items["type"] != "integer" {
			t.Errorf("expected items type 'integer', got '%v'", items["type"])
		}
	})

	t.Run("map type", func(t *testing.T) {
		schema := TypeToSchema(reflect.TypeOf(map[string]int{}))

		if schema["type"] != "object" {
			t.Errorf("expected type 'object', got '%v'", schema["type"])
		}

		additionalProps, ok := schema["additionalProperties"].(map[string]any)
		if !ok {
			t.Fatal("additionalProperties should be a map")
		}

		if additionalProps["type"] != "integer" {
			t.Errorf("expected additionalProperties type 'integer', got '%v'", additionalProps["type"])
		}
	})

	t.Run("time.Time type", func(t *testing.T) {
		schema := TypeToSchema(reflect.TypeOf(time.Time{}))

		if schema["type"] != "string" {
			t.Errorf("expected type 'string' for time.Time, got '%v'", schema["type"])
		}

		format, ok := schema["format"].(string)
		if !ok || format != "date-time" {
			t.Errorf("expected format 'date-time', got '%v'", schema["format"])
		}
	})

	t.Run("complex struct", func(t *testing.T) {
		type Address struct {
			Street string `json:"street"`
			City   string `json:"city"`
		}

		type Person struct {
			Name    string  `json:"name"`
			Age     int     `json:"age"`
			Address Address `json:"address"`
		}

		schema := TypeToSchema(reflect.TypeOf(Person{}))

		if schema["type"] != "object" {
			t.Errorf("expected type 'object', got '%v'", schema["type"])
		}

		properties, ok := schema["properties"].(map[string]any)
		if !ok {
			t.Fatal("properties should be a map")
		}

		if _, ok := properties["name"]; !ok {
			t.Error("should have 'name' property")
		}

		if _, ok := properties["age"]; !ok {
			t.Error("should have 'age' property")
		}

		if _, ok := properties["address"]; !ok {
			t.Error("should have 'address' property")
		}
	})
}

// TestStructToSchema tests the struct schema conversion
func TestStructToSchema(t *testing.T) {
	t.Run("basic struct", func(t *testing.T) {
		type Simple struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		schema := structToSchema(reflect.TypeOf(Simple{}))

		if schema["type"] != "object" {
			t.Errorf("expected type 'object', got '%v'", schema["type"])
		}

		properties, ok := schema["properties"].(map[string]any)
		if !ok {
			t.Fatal("properties should be a map")
		}

		if len(properties) != 2 {
			t.Errorf("expected 2 properties, got %d", len(properties))
		}

		nameSchema, ok := properties["name"].(map[string]any)
		if !ok {
			t.Fatal("name should be a map")
		}

		if nameSchema["type"] != "string" {
			t.Errorf("expected name type 'string', got '%v'", nameSchema["type"])
		}

		ageSchema, ok := properties["age"].(map[string]any)
		if !ok {
			t.Fatal("age should be a map")
		}

		if ageSchema["type"] != "integer" {
			t.Errorf("expected age type 'integer', got '%v'", ageSchema["type"])
		}
	})

	t.Run("with required fields", func(t *testing.T) {
		type RequiredTest struct {
			Name     string `json:"name"`
			Optional string `json:"optional,omitempty"`
		}

		schema := structToSchema(reflect.TypeOf(RequiredTest{}))

		required, ok := schema["required"].([]string)
		if !ok {
			t.Fatal("required should be a slice")
		}

		foundName := false
		foundOptional := false
		for _, field := range required {
			if field == "name" {
				foundName = true
			}
			if field == "optional" {
				foundOptional = true
			}
		}

		if !foundName {
			t.Error("name should be in required")
		}

		if foundOptional {
			t.Error("optional should not be in required (has omitempty)")
		}
	})

	t.Run("with json tag options", func(t *testing.T) {
		type TagTest struct {
			FieldName string `json:"custom_name"`
			Skipped   string `json:"-"`
			LowerCase string // no tag, should use lowercase field name
		}

		schema := structToSchema(reflect.TypeOf(TagTest{}))

		properties, ok := schema["properties"].(map[string]any)
		if !ok {
			t.Fatal("properties should be a map")
		}

		if _, ok := properties["custom_name"]; !ok {
			t.Error("should have 'custom_name' property (from json tag)")
		}

		if _, ok := properties["Skipped"]; ok {
			t.Error("should not have 'Skipped' property (json:'-')")
		}

		if _, ok := properties["lowercase"]; !ok {
			t.Error("should have 'lowercase' property (auto-generated from field name)")
		}
	})

	t.Run("with openapi description", func(t *testing.T) {
		type Described struct {
			Name string `json:"name" openapi:"The user's full name"`
		}

		schema := structToSchema(reflect.TypeOf(Described{}))

		properties, ok := schema["properties"].(map[string]any)
		if !ok {
			t.Fatal("properties should be a map")
		}

		nameSchema, ok := properties["name"].(map[string]any)
		if !ok {
			t.Fatal("name should be a map")
		}

		if nameSchema["description"] != "The user's full name" {
			t.Errorf("expected description 'The user's full name', got '%v'", nameSchema["description"])
		}
	})

	t.Run("unexported fields ignored", func(t *testing.T) {
		type Mixed struct {
			Exported   string
			unexported string
			Another    string
		}

		schema := structToSchema(reflect.TypeOf(Mixed{}))

		properties, ok := schema["properties"].(map[string]any)
		if !ok {
			t.Fatal("properties should be a map")
		}

		// Should only have exported fields
		if len(properties) != 2 {
			t.Errorf("expected 2 properties (exported only), got %d", len(properties))
		}

		if _, ok := properties["unexported"]; ok {
			t.Error("should not have unexported field")
		}
	})

	t.Run("nested struct", func(t *testing.T) {
		type Address struct {
			Street string `json:"street"`
			City   string `json:"city"`
		}

		type Person struct {
			Name    string  `json:"name"`
			Address Address `json:"address"`
		}

		schema := structToSchema(reflect.TypeOf(Person{}))

		properties, ok := schema["properties"].(map[string]any)
		if !ok {
			t.Fatal("properties should be a map")
		}

		addressSchema, ok := properties["address"].(map[string]any)
		if !ok {
			t.Fatal("address should be a map")
		}

		if addressSchema["type"] != "object" {
			t.Errorf("expected nested address type 'object', got '%v'", addressSchema["type"])
		}

		addressProps, ok := addressSchema["properties"].(map[string]any)
		if !ok {
			t.Fatal("address properties should be a map")
		}

		if _, ok := addressProps["street"]; !ok {
			t.Error("address should have 'street' property")
		}

		if _, ok := addressProps["city"]; !ok {
			t.Error("address should have 'city' property")
		}
	})

	t.Run("slice field", func(t *testing.T) {
		type WithSlice struct {
			Tags []string `json:"tags"`
		}

		schema := structToSchema(reflect.TypeOf(WithSlice{}))

		properties, ok := schema["properties"].(map[string]any)
		if !ok {
			t.Fatal("properties should be a map")
		}

		tagsSchema, ok := properties["tags"].(map[string]any)
		if !ok {
			t.Fatal("tags should be a map")
		}

		if tagsSchema["type"] != "array" {
			t.Errorf("expected tags type 'array', got '%v'", tagsSchema["type"])
		}
	})

	t.Run("pointer field", func(t *testing.T) {
		type WithPointer struct {
			Optional *string `json:"optional,omitempty"`
		}

		schema := structToSchema(reflect.TypeOf(WithPointer{}))

		properties, ok := schema["properties"].(map[string]any)
		if !ok {
			t.Fatal("properties should be a map")
		}

		optionalSchema, ok := properties["optional"].(map[string]any)
		if !ok {
			t.Fatal("optional should be a map")
		}

		// Pointer should resolve to underlying type
		if optionalSchema["type"] != "string" {
			t.Errorf("expected optional type 'string', got '%v'", optionalSchema["type"])
		}
	})
}

// TestSchemaFromComplexTypes tests complex type schemas
func TestSchemaFromComplexTypes(t *testing.T) {
	t.Run("slice of strings", func(t *testing.T) {
		schema := SchemaFrom[[]string]()

		if schema["type"] != "array" {
			t.Errorf("expected type 'array', got '%v'", schema["type"])
		}

		items, ok := schema["items"].(map[string]any)
		if !ok {
			t.Fatal("items should be a map")
		}

		if items["type"] != "string" {
			t.Errorf("expected items type 'string', got '%v'", items["type"])
		}
	})

	t.Run("map of string to int", func(t *testing.T) {
		schema := SchemaFrom[map[string]int]()

		if schema["type"] != "object" {
			t.Errorf("expected type 'object', got '%v'", schema["type"])
		}

		additionalProps, ok := schema["additionalProperties"].(map[string]any)
		if !ok {
			t.Fatal("additionalProperties should be a map")
		}

		if additionalProps["type"] != "integer" {
			t.Errorf("expected additionalProperties type 'integer', got '%v'", additionalProps["type"])
		}
	})

	t.Run("pointer to struct", func(t *testing.T) {
		type User struct {
			Name string `json:"name"`
		}

		schema := SchemaFrom[*User]()

		if schema["type"] != "object" {
			t.Errorf("expected type 'object' for pointer to struct, got '%v'", schema["type"])
		}
	})
}
