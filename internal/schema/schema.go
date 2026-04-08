package schema

import (
	"reflect"
	"strings"
	"time"
)

// SchemaFrom generates a JSON Schema from a Go type using reflection
func SchemaFrom[T any]() map[string]any {
	var zero T
	return TypeToSchema(reflect.TypeOf(zero))
}

func TypeToSchema(t reflect.Type) map[string]any {
	if t == nil {
		return map[string]any{"type": "null"}
	}

	if t.Kind() == reflect.Pointer {
		return TypeToSchema(t.Elem())
	}

	schema := make(map[string]any)

	switch t.Kind() {
	case reflect.Struct:
		return structToSchema(t)
	case reflect.String:
		schema["type"] = "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		schema["type"] = "integer"
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		schema["type"] = "integer"
	case reflect.Float32, reflect.Float64:
		schema["type"] = "number"
	case reflect.Bool:
		schema["type"] = "boolean"
	case reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 && t.Len() == 16 {
			schema["type"] = "string"
			schema["format"] = "uuid"
		} else {
			schema["type"] = "array"
			schema["items"] = TypeToSchema(t.Elem())
		}
	case reflect.Slice:
		schema["type"] = "array"
		schema["items"] = TypeToSchema(t.Elem())
	case reflect.Map:
		schema["type"] = "object"
		schema["additionalProperties"] = TypeToSchema(t.Elem())
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
			jsonTag = strings.Split(jsonTag, ",")[0]
		}

		if jsonTag == "-" {
			continue
		}

		fieldSchema := TypeToSchema(field.Type)

		if desc := field.Tag.Get("openapi"); desc != "" {
			fieldSchema["description"] = desc
		}

		properties[jsonTag] = fieldSchema

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

// ShouldBindFromQuery reports whether a type should use query parameter binding
func ShouldBindFromQuery(t reflect.Type) bool {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return false
	}

	for field := range t.Fields() {
		if field.Tag.Get("query") != "" || field.Tag.Get("path") != "" {
			return true
		}
	}
	return false
}

// ExtractQueryParams extracts query parameter metadata from a struct type
func ExtractQueryParams(t reflect.Type) []QueryParamInfo {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	var params []QueryParamInfo
	for field := range t.Fields() {
		if queryTag := field.Tag.Get("query"); queryTag != "" {
			param := QueryParamInfo{
				Name:     queryTag,
				Type:     goTypeToJSONType(field.Type),
				Required: field.Tag.Get("required") == "true",
			}
			if defaultTag := field.Tag.Get("default"); defaultTag != "" {
				param.Default = defaultTag
			}
			params = append(params, param)
		}
	}
	return params
}

// ExtractPathParams extracts path parameter metadata from a struct type
func ExtractPathParams(t reflect.Type) []PathParamInfo {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	var params []PathParamInfo
	for field := range t.Fields() {
		field := field
		if pathTag := field.Tag.Get("path"); pathTag != "" {
			params = append(params, PathParamInfo{
				Name: pathTag,
				Type: goTypeToJSONType(field.Type),
			})
		}
	}
	return params
}

func goTypeToJSONType(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 && t.Len() == 16 {
			return "string"
		}
		return "array"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	default:
		return "string"
	}
}

// QueryParamInfo contains metadata for a query parameter
type QueryParamInfo struct {
	Name        string
	Type        string
	Required    bool
	Default     string
	Description string
}

// PathParamInfo contains metadata for a path parameter
type PathParamInfo struct {
	Name        string
	Type        string
	Description string
}
