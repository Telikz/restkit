package docs

import (
	"encoding/json"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	ep "github.com/RestStore/RestKit/internal/endpoints"
	"github.com/RestStore/RestKit/internal/schema"
)

// GenerateOpenAPI generates an OpenAPI 3.0 specification from endpoints
func GenerateOpenAPI(
	title, description, version string,
	endpoints []ep.Endpoint, groups []*ep.Group,
) map[string]any {
	paths := make(map[string]any)
	tags := make([]map[string]any, 0)

	for _, group := range groups {
		if group.Title != "" {
			tags = append(tags, map[string]any{
				"name":        group.Title,
				"description": group.Description,
			})
		}
	}

	// Add endpoints from groups
	for _, group := range groups {
		for _, endpoint := range group.GetEndpoints() {
			path := endpoint.GetPath()
			method := strings.ToLower(endpoint.GetMethod())

			if paths[path] == nil {
				paths[path] = make(map[string]any)
			}

			pathOps := paths[path].(map[string]any)
			pathOps[method] = buildOperation(endpoint, groups)
		}
	}

	// Add individual endpoints (avoid duplicates)
	registered := make(map[string]bool)
	for _, group := range groups {
		for _, ep := range group.GetEndpoints() {
			key := ep.GetMethod() + " " + ep.GetPath()
			registered[key] = true
		}
	}

	for _, endpoint := range endpoints {
		key := endpoint.GetMethod() + " " + endpoint.GetPath()
		if !registered[key] {
			path := endpoint.GetPath()
			method := strings.ToLower(endpoint.GetMethod())

			if paths[path] == nil {
				paths[path] = make(map[string]any)
			}

			pathOps := paths[path].(map[string]any)
			pathOps[method] = buildOperation(endpoint, groups)
		}
	}

	spec := map[string]any{
		"openapi": "3.0.0",
		"info": map[string]any{
			"title":       title,
			"description": description,
			"version":     version,
		},
		"paths": paths,
	}

	if len(tags) > 0 {
		spec["tags"] = tags
	}

	return spec
}

// buildOperation constructs an OpenAPI operation object from an endpoint definition
func buildOperation(
	endpoint ep.Endpoint, groups []*ep.Group,
) map[string]any {
	op := map[string]any{
		"summary":     endpoint.GetTitle(),
		"description": endpoint.GetDescription(),
	}

	for _, group := range groups {
		if group.Title != "" {
			for _, groupEndpoint := range group.GetEndpoints() {
				if groupEndpoint.GetMethod() == endpoint.GetMethod() &&
					groupEndpoint.GetPath() == endpoint.GetPath() {
					op["tags"] = []string{group.Title}
					break
				}
			}
			if op["tags"] != nil {
				break
			}
		}
	}

	reqSchema := endpoint.GetRequestSchema()
	if reqSchema != nil {
		op["requestBody"] = map[string]any{
			"required": true,
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": reqSchema,
				},
			},
		}
	}

	responses := make(map[string]any)

	resSchema := endpoint.GetResponseSchema()
	if resSchema != nil {
		responses["200"] = map[string]any{
			"description": "Success",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": resSchema,
				},
			},
		}
	} else {
		responses["204"] = map[string]any{
			"description": "No Content",
		}
	}

	responses["400"] = map[string]any{
		"description": "Bad Request",
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"status": map[string]any{
							"type":    "integer",
							"example": 400,
						},
						"code": map[string]any{
							"type":    "string",
							"example": "bad_request",
						},
						"message": map[string]any{
							"type":    "string",
							"example": "Failed to parse request",
						},
					},
					"required": []string{"status", "code", "message"},
				},
			},
		},
	}

	responses["422"] = map[string]any{
		"description": "Validation Error",
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"status": map[string]any{
							"type":    "integer",
							"example": 422,
						},
						"code": map[string]any{
							"type":    "string",
							"example": "validation",
						},
						"message": map[string]any{
							"type":    "string",
							"example": "Validation failed",
						},
						"errors": map[string]any{
							"type": "array",
							"items": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"field": map[string]any{
										"type": "string",
									},
									"message": map[string]any{
										"type": "string",
									},
								},
							},
						},
					},
					"required": []string{"status", "code", "message"},
				},
			},
		},
	}

	responses["500"] = map[string]any{
		"description": "Internal Server Error",
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"status": map[string]any{
							"type":    "integer",
							"example": 500,
						},
						"code": map[string]any{
							"type":    "string",
							"example": "internal",
						},
						"message": map[string]any{
							"type":    "string",
							"example": "Internal server error",
						},
					},
					"required": []string{"status", "code", "message"},
				},
			},
		},
	}

	op["responses"] = responses

	pathParams := extractPathParameters(endpoint.GetPath())
	if len(pathParams) > 0 {
		parameters := make([]map[string]any, 0)
		for _, param := range pathParams {
			parameters = append(parameters, map[string]any{
				"name":        param,
				"in":          "path",
				"required":    true,
				"schema":      map[string]any{"type": "string"},
				"description": param + " parameter",
			})
		}
		op["parameters"] = parameters
	}

	return op
}

// generateSchema creates a JSON schema from a Go type using reflection
func generateSchema(v any) map[string]any {
	if v == nil {
		return map[string]any{"type": "null"}
	}

	t := reflect.TypeOf(v)
	return typeToSchema(t)
}

// typeToSchema converts a reflect.Type to a JSON schema
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

// structToSchema converts a struct type to a JSON schema
func structToSchema(t reflect.Type) map[string]any {
	schema := map[string]any{
		"type": "object",
	}

	properties := make(map[string]any)
	required := make([]string, 0)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

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

func AddMountedRoutesToSpec(
	spec map[string]any,
	prefix string,
	routes []schema.MountedRoute,
) {
	paths, ok := spec["paths"].(map[string]any)
	if !ok {
		paths = make(map[string]any)
		spec["paths"] = paths
	}

	for _, route := range routes {
		// Combine prefix with route path, avoiding double slashes
		fullPath := prefix + route.Path
		if strings.HasSuffix(prefix, "/") &&
			strings.HasPrefix(route.Path, "/") {
			fullPath = prefix + route.Path[1:]
		}
		method := strings.ToLower(route.Method)

		if paths[fullPath] == nil {
			paths[fullPath] = make(map[string]any)
		}

		pathOps := paths[fullPath].(map[string]any)
		pathOps[method] = buildMountedRouteOperation(route)
	}
}

// buildMountedRouteOperation creates an OpenAPI operation from a mounted route
func buildMountedRouteOperation(route schema.MountedRoute) map[string]any {
	op := map[string]any{}

	if route.Summary != "" {
		op["summary"] = route.Summary
	}
	if route.Description != "" {
		op["description"] = route.Description
	}

	// Add request body if request type is provided
	if route.RequestType != nil {
		reqSchema := generateSchema(route.RequestType)
		op["requestBody"] = map[string]any{
			"required": true,
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": reqSchema,
				},
			},
		}
	}

	// Add responses
	responses := make(map[string]any)

	if route.ResponseType != nil {
		resSchema := generateSchema(route.ResponseType)
		responses["200"] = map[string]any{
			"description": "Success",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": resSchema,
				},
			},
		}
	} else {
		responses["204"] = map[string]any{
			"description": "No Content",
		}
	}

	// Add standard error responses
	responses["400"] = map[string]any{
		"description": "Bad Request",
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"status": map[string]any{
							"type":    "integer",
							"example": 400,
						},
						"code": map[string]any{
							"type":    "string",
							"example": "bad_request",
						},
						"message": map[string]any{
							"type":    "string",
							"example": "Failed to parse request",
						},
					},
					"required": []string{"status", "code", "message"},
				},
			},
		},
	}

	responses["500"] = map[string]any{
		"description": "Internal Server Error",
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"status": map[string]any{
							"type":    "integer",
							"example": 500,
						},
						"code": map[string]any{
							"type":    "string",
							"example": "internal",
						},
						"message": map[string]any{
							"type":    "string",
							"example": "Internal server error",
						},
					},
					"required": []string{"status", "code", "message"},
				},
			},
		},
	}

	op["responses"] = responses

	// Add path parameters
	if len(route.PathParams) > 0 {
		parameters := make([]map[string]any, 0)
		for _, param := range route.PathParams {
			paramSchema := map[string]any{"type": param.Type}
			if param.Type == "" {
				paramSchema["type"] = "string"
			}
			parameters = append(parameters, map[string]any{
				"name":        param.Name,
				"in":          "path",
				"required":    param.Required,
				"schema":      paramSchema,
				"description": param.Description,
			})
		}
		op["parameters"] = parameters
	} else {
		// Auto-extract path parameters from route path
		pathParams := extractPathParameters(route.Path)
		if len(pathParams) > 0 {
			parameters := make([]map[string]any, 0)
			for _, param := range pathParams {
				parameters = append(parameters, map[string]any{
					"name":        param,
					"in":          "path",
					"required":    true,
					"schema":      map[string]any{"type": "string"},
					"description": param + " parameter",
				})
			}
			op["parameters"] = parameters
		}
	}

	return op
}

// extractPathParameters extracts parameter names from a path pattern like "/users/{id}"
func extractPathParameters(path string) []string {
	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(path, -1)
	params := make([]string, 0)
	for _, match := range matches {
		params = append(params, match[1])
	}
	return params
}

// ServeOpenAPI writes the OpenAPI spec as JSON
func ServeOpenAPI(w http.ResponseWriter, spec map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(spec)
}
