package schema

import (
	"reflect"
	"strings"
	"time"
)

// SchemaFrom generates a JSON Schema from a Go type using reflection
func SchemaFrom[T any]() map[string]any {
	var zero T
	return typeToSchema(reflect.TypeOf(zero))
}

func typeToSchema(t reflect.Type) map[string]any {
	if t == nil {
		return map[string]any{"type": "null"}
	}

	if t.Kind() == reflect.Pointer {
		return typeToSchema(t.Elem())
	}

	schema := make(map[string]any)

	switch t.Kind() {
	case reflect.Struct:
		return structToSchema(t)
	case reflect.String:
		schema["type"] = "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		schema["type"] = "integer"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema["type"] = "integer"
	case reflect.Float32, reflect.Float64:
		schema["type"] = "number"
	case reflect.Bool:
		schema["type"] = "boolean"
	case reflect.Slice, reflect.Array:
		schema["type"] = "array"
		schema["items"] = typeToSchema(t.Elem())
	case reflect.Map:
		schema["type"] = "object"
		schema["additionalProperties"] = typeToSchema(t.Elem())
	default:
		schema["type"] = "string"
	}

	return schema
}

func structToSchema(t reflect.Type) map[string]any {
	if t == reflect.TypeFor[time.Time]() {
		return map[string]any{
			"type":   "string",
			"format": "date-time",
		}
	}

	schema := map[string]any{
		"type": "object",
	}

	properties := make(map[string]any)
	required := make([]string, 0)

	for field := range t.Fields() {

		if !field.IsExported() {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = strings.ToLower(field.Name)
		} else {
			// Handle tags like "json:field_name,omitempty"
			jsonTag = strings.Split(jsonTag, ",")[0]
		}

		if jsonTag == "-" {
			continue
		}

		fieldSchema := typeToSchema(field.Type)

		// Add description from openapi tag if present
		if desc := field.Tag.Get("openapi"); desc != "" {
			fieldSchema["description"] = desc
		}

		properties[jsonTag] = fieldSchema

		// Check if field is required
		if !strings.Contains(field.Tag.Get("json"), "omitempty") {
			required = append(required, jsonTag)
		}
	}

	schema["properties"] = properties
	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}
