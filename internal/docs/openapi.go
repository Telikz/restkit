package docs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	ep "github.com/reststore/restkit/internal/endpoints"
	"github.com/reststore/restkit/internal/schema"
)

type OpenAPISpec struct {
	Version     string      `json:"version"`
	Title       string      `json:"title"`
	Summary     string      `json:"summary"`
	Description string      `json:"description"`
	Endpoints   []ep.Route  `json:"endpoints"`
	Groups      []*ep.Group `json:"groups"`
	Servers     []Server    `json:"servers"`
}

type Server struct {
	URL         string
	Description string
	Variables   map[string]struct {
		Default     string
		Description string
	}
}

// GenerateOpenAPI generates an OpenAPI 3.0 specification from endpoints
func GenerateOpenAPI(s *OpenAPISpec) map[string]any {
	paths := make(map[string]any)
	tags := make([]map[string]any, 0)

	for _, group := range s.Groups {
		if group.Title != "" {
			tags = append(tags, map[string]any{
				"name":        group.Title,
				"description": group.Description,
			})
		}
	}

	// Add endpoints from groups
	for _, group := range s.Groups {
		for _, endpoint := range group.GetEndpoints() {
			path := endpoint.GetPath()
			method := strings.ToLower(endpoint.GetMethod())

			if paths[path] == nil {
				paths[path] = make(map[string]any)
			}

			pathOps := paths[path].(map[string]any)
			pathOps[method] = buildOperation(endpoint, s.Groups)
		}
	}

	// Add individual endpoints (avoid duplicates)
	registered := make(map[string]bool)
	for _, group := range s.Groups {
		for _, e := range group.GetEndpoints() {
			key := fmt.Sprintf("%s %s", e.GetMethod(), e.GetPath())
			registered[key] = true
		}
	}

	for _, endpoint := range s.Endpoints {
		key := fmt.Sprintf("%s %s", endpoint.GetMethod(), endpoint.GetPath())
		if !registered[key] {
			path := endpoint.GetPath()
			method := strings.ToLower(endpoint.GetMethod())

			if paths[path] == nil {
				paths[path] = make(map[string]any)
			}

			pathOps := paths[path].(map[string]any)
			pathOps[method] = buildOperation(endpoint, s.Groups)
		}
	}

	servers := make([]map[string]any, 0, len(s.Servers))
	for _, server := range s.Servers {
		servers = append(servers, map[string]any{
			"url":         server.URL,
			"description": server.Description,
			"variables":   server.Variables,
		})
	}

	spec := map[string]any{
		"openapi": "3.0.0",
		"servers": servers,
		"info": map[string]any{
			"title":       s.Title,
			"summary":     s.Summary,
			"description": s.Description,
			"version":     s.Version,
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
	endpoint ep.Route, groups []*ep.Group,
) map[string]any {
	op := map[string]any{
		"title":       endpoint.GetTitle(),
		"summary":     endpoint.GetSummary(),
		"description": endpoint.GetDescription(),
	}

	// Build parameters list
	var paramList []map[string]any

	// Add existing parameters from endpoint
	for _, p := range endpoint.GetParameters() {
		param := map[string]any{
			"name":        p.Name,
			"in":          string(p.Location),
			"description": p.Description,
			"schema": map[string]any{
				"type": p.Type,
			},
		}
		if p.Required {
			param["required"] = true
		}
		paramList = append(paramList, param)
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
	method := endpoint.GetMethod()

	// Only add request body for methods that typically have one
	// GET/DELETE/HEAD should use query/path params, not request body
	if reqSchema != nil && !isEmptyRequestSchema(reqSchema) &&
		method != http.MethodGet && method != http.MethodDelete && method != http.MethodHead {
		op["requestBody"] = map[string]any{
			"required": true,
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": reqSchema,
				},
			},
		}
	}

	// Add path parameters from URL pattern that aren't already in paramList
	urlPathParams := extractPathParameters(endpoint.GetPath())
	existingParams := make(map[string]bool)
	for _, p := range paramList {
		if name, ok := p["name"].(string); ok {
			existingParams[name] = true
		}
	}
	if len(urlPathParams) > 0 {
		for _, param := range urlPathParams {
			if !existingParams[param] {
				paramList = append(paramList, map[string]any{
					"name":        param,
					"in":          "path",
					"required":    true,
					"schema":      map[string]any{"type": "string"},
					"description": param + " parameter",
				})
			}
		}
	}

	if len(paramList) > 0 {
		op["parameters"] = paramList
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

	return op
}

// generateSchema creates a JSON schema from a Go type using reflection
func generateSchema(v any) map[string]any {
	if v == nil {
		return map[string]any{"type": "null"}
	}

	t := reflect.TypeOf(v)
	return schema.TypeToSchema(t)
}

// structToSchema converts a struct type to a JSON schema
func structToSchema(t reflect.Type) map[string]any {
	sc := map[string]any{
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

		fieldSchema := schema.TypeToSchema(field.Type)

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

	sc["properties"] = properties
	if len(required) > 0 {
		sc["required"] = required
	}

	return sc
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
	re := regexp.MustCompile(`\{([^}]+)}`)
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
	_ = json.NewEncoder(w).Encode(spec)
}

func CreateOpenAPIFile(path string, spec map[string]any) (err error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer func() {
		closeErr := file.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")

	if err := enc.Encode(spec); err != nil {
		return err
	}

	return nil
}

func isEmptyRequestSchema(schema map[string]any) bool {
	if schema == nil {
		return true
	}

	if schemaType, ok := schema["type"].(string); ok && schemaType == "object" {
		if properties, ok := schema["properties"].(map[string]any); ok {
			return len(properties) == 0
		}
		return true
	}

	return false
}
