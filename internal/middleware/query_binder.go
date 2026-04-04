package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
)

// QueryBinder binds query and path parameters to a struct using tags.
// Supports: `query:"name"`, `query:"name" default:"20"`, `path:"id"`.
// Pointer fields are optional (nil when value is empty).
func QueryBinder[Req any]() func(r *http.Request) (Req, error) {
	return func(r *http.Request) (Req, error) {
		var req Req
		reqVal := reflect.ValueOf(&req).Elem()
		reqType := reqVal.Type()

		for i := 0; i < reqVal.NumField(); i++ {
			field := reqVal.Field(i)
			fieldType := reqType.Field(i)

			if pathTag := fieldType.Tag.Get("path"); pathTag != "" {
				if err := bindPathParam(r, field, pathTag); err != nil {
					return req, err
				}
				continue
			}

			if queryTag := fieldType.Tag.Get("query"); queryTag != "" {
				if err := bindQueryParam(r, field, fieldType, queryTag); err != nil {
					return req, err
				}
			}
		}

		return req, nil
	}
}

// MixedBinder binds path params from URL and JSON body to a struct.
// Use for Update endpoints that need both `path:"id"` and `json:"field"` tags.
func MixedBinder[Req any]() func(r *http.Request) (Req, error) {
	return func(r *http.Request) (Req, error) {
		var req Req
		reqVal := reflect.ValueOf(&req).Elem()
		reqType := reqVal.Type()

		// First: bind path parameters
		for i := 0; i < reqVal.NumField(); i++ {
			field := reqVal.Field(i)
			fieldType := reqType.Field(i)

			if pathTag := fieldType.Tag.Get("path"); pathTag != "" {
				if err := bindPathParam(r, field, pathTag); err != nil {
					return req, err
				}
			}
		}

		// Second: decode JSON body (only if there's a body)
		if r.Body != nil && r.ContentLength > 0 {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				return req, fmt.Errorf("failed to decode JSON body: %w", err)
			}
		}

		return req, nil
	}
}

func bindPathParam(r *http.Request, field reflect.Value, paramName string) error {
	value := r.PathValue(paramName)
	if value == "" {
		return fmt.Errorf("path parameter %q not found", paramName)
	}
	return setFieldValue(field, value)
}

func bindQueryParam(
	r *http.Request,
	field reflect.Value,
	fieldType reflect.StructField,
	paramName string,
) error {
	value := r.URL.Query().Get(paramName)

	if value == "" {
		if defaultTag := fieldType.Tag.Get("default"); defaultTag != "" {
			value = defaultTag
		}
	}

	if field.Kind() == reflect.Ptr {
		if value == "" {
			return nil
		}
		elemType := field.Type().Elem()
		elem := reflect.New(elemType)
		if err := setFieldValue(elem.Elem(), value); err != nil {
			return fmt.Errorf("query parameter %q: %w", paramName, err)
		}
		field.Set(elem)
		return nil
	}

	if value == "" {
		if fieldType.Tag.Get("required") == "true" {
			return fmt.Errorf("required query parameter %q is missing", paramName)
		}
		return nil
	}

	if err := setFieldValue(field, value); err != nil {
		return fmt.Errorf("query parameter %q: %w", paramName, err)
	}

	return nil
}

func setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer value %q", value)
		}
		field.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unsigned integer value %q", value)
		}
		field.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value %q", value)
		}
		field.SetFloat(floatVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value %q", value)
		}
		field.SetBool(boolVal)
	default:
		return fmt.Errorf("unsupported field type %v", field.Kind())
	}
	return nil
}

// QueryParamInfo contains metadata for OpenAPI documentation.
type QueryParamInfo struct {
	Name        string
	Type        string
	Required    bool
	Default     string
	Description string
}

// PathParamInfo contains metadata for OpenAPI documentation.
type PathParamInfo struct {
	Name        string
	Type        string
	Description string
}

// ExtractQueryParams extracts query param metadata from a struct type.
func ExtractQueryParams[Req any]() []QueryParamInfo {
	var req Req
	reqType := reflect.TypeOf(req)

	var params []QueryParamInfo
	for i := 0; i < reqType.NumField(); i++ {
		field := reqType.Field(i)
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

// ExtractPathParams extracts path param metadata from a struct type.
func ExtractPathParams[Req any]() []PathParamInfo {
	var req Req
	reqType := reflect.TypeOf(req)

	var params []PathParamInfo
	for i := 0; i < reqType.NumField(); i++ {
		field := reqType.Field(i)
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

// HasQueryTag reports whether the type has fields with `query` tags.
func HasQueryTag[Req any]() bool {
	var req Req
	reqType := reflect.TypeOf(req)

	for i := 0; i < reqType.NumField(); i++ {
		if reqType.Field(i).Tag.Get("query") != "" {
			return true
		}
	}
	return false
}

// HasPathTag reports whether the type has fields with `path` tags.
func HasPathTag[Req any]() bool {
	var req Req
	reqType := reflect.TypeOf(req)

	for i := 0; i < reqType.NumField(); i++ {
		if reqType.Field(i).Tag.Get("path") != "" {
			return true
		}
	}
	return false
}
