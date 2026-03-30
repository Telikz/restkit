package docs

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/telikz/restkit/internal/endpoints"
)

// GenerateOpenAPI generates an OpenAPI 3.0 specification from endpoints
func GenerateOpenAPI(title, description, version string, endpoints []endpoints.Endpoint, groups []*endpoints.Group) map[string]any {
	paths := make(map[string]any)
	tags := make([]map[string]any, 0)

	for _, group := range groups {
		if group.GetTitle() != "" {
			tags = append(tags, map[string]any{
				"name":        group.GetTitle(),
				"description": group.GetDescription(),
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
func buildOperation(endpoint endpoints.Endpoint, groups []*endpoints.Group) map[string]any {
	op := map[string]any{
		"summary":     endpoint.GetTitle(),
		"description": endpoint.GetDescription(),
	}

	for _, group := range groups {
		if group.GetTitle() != "" {
			for _, groupEndpoint := range group.GetEndpoints() {
				if groupEndpoint.GetMethod() == endpoint.GetMethod() && groupEndpoint.GetPath() == endpoint.GetPath() {
					op["tags"] = []string{group.GetTitle()}
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
