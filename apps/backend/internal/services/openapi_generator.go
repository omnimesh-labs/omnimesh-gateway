package services

import (
	"fmt"
	"mcp-gateway/apps/backend/internal/types"
	"strings"
)

// OpenAPIGenerator generates OpenAPI specifications for endpoints
type OpenAPIGenerator struct {
	baseURL string
}

// NewOpenAPIGenerator creates a new OpenAPI generator
func NewOpenAPIGenerator(baseURL string) *OpenAPIGenerator {
	return &OpenAPIGenerator{
		baseURL: baseURL,
	}
}

// OpenAPISpec represents an OpenAPI 3.0 specification
type OpenAPISpec struct {
	OpenAPI    string                 `json:"openapi"`
	Info       Info                   `json:"info"`
	Servers    []Server               `json:"servers"`
	Paths      map[string]PathItem    `json:"paths"`
	Components Components             `json:"components"`
	Security   []SecurityRequirement  `json:"security,omitempty"`
}

// Info represents the info section of OpenAPI spec
type Info struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

// Server represents a server in OpenAPI spec
type Server struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// PathItem represents a path in OpenAPI spec
type PathItem struct {
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
}

// Operation represents an operation in OpenAPI spec
type Operation struct {
	Summary     string                 `json:"summary,omitempty"`
	Description string                 `json:"description,omitempty"`
	OperationID string                 `json:"operationId"`
	Parameters  []Parameter            `json:"parameters,omitempty"`
	RequestBody *RequestBody           `json:"requestBody,omitempty"`
	Responses   map[string]Response    `json:"responses"`
	Security    []SecurityRequirement  `json:"security,omitempty"`
}

// Parameter represents a parameter in OpenAPI spec
type Parameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"`
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"required"`
	Schema      Schema      `json:"schema"`
}

// RequestBody represents a request body in OpenAPI spec
type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Required    bool                 `json:"required"`
	Content     map[string]MediaType `json:"content"`
}

// Response represents a response in OpenAPI spec
type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// MediaType represents a media type in OpenAPI spec
type MediaType struct {
	Schema Schema `json:"schema"`
}

// Schema represents a schema in OpenAPI spec
type Schema struct {
	Type       string                 `json:"type,omitempty"`
	Format     string                 `json:"format,omitempty"`
	Properties map[string]Schema      `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
	Items      *Schema                `json:"items,omitempty"`
	Ref        string                 `json:"$ref,omitempty"`
	Example    interface{}            `json:"example,omitempty"`
}

// Components represents the components section of OpenAPI spec
type Components struct {
	Schemas         map[string]Schema         `json:"schemas,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

// SecurityScheme represents a security scheme in OpenAPI spec
type SecurityScheme struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Name        string `json:"name,omitempty"`
	In          string `json:"in,omitempty"`
	Scheme      string `json:"scheme,omitempty"`
	Flows       *Flows `json:"flows,omitempty"`
}

// Flows represents OAuth flows in OpenAPI spec
type Flows struct {
	AuthorizationCode *Flow `json:"authorizationCode,omitempty"`
	Implicit          *Flow `json:"implicit,omitempty"`
	ClientCredentials *Flow `json:"clientCredentials,omitempty"`
}

// Flow represents an OAuth flow in OpenAPI spec
type Flow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}

// SecurityRequirement represents a security requirement in OpenAPI spec
type SecurityRequirement map[string][]string

// GenerateSpec generates an OpenAPI specification for an endpoint
func (g *OpenAPIGenerator) GenerateSpec(endpoint *types.Endpoint, namespace *types.Namespace, tools []types.NamespaceTool) *OpenAPISpec {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: Info{
			Title:       fmt.Sprintf("%s - MCP Gateway", endpoint.Name),
			Description: g.generateDescription(endpoint, namespace),
			Version:     "1.0.0",
		},
		Servers: []Server{
			{
				URL:         fmt.Sprintf("%s/api/public/endpoints/%s/api", g.baseURL, endpoint.Name),
				Description: "MCP Gateway Endpoint API",
			},
		},
		Paths:      make(map[string]PathItem),
		Components: g.generateComponents(endpoint),
	}

	// Add security requirements if not public
	if !endpoint.EnablePublicAccess {
		spec.Security = g.generateSecurityRequirements(endpoint)
	}

	// Generate paths for each tool
	for _, tool := range tools {
		path := fmt.Sprintf("/tools/%s", tool.ToolName)
		spec.Paths[path] = g.generateToolPath(tool, endpoint)
	}

	// Add tools listing endpoint
	spec.Paths["/tools"] = g.generateToolsListPath(endpoint)

	return spec
}

func (g *OpenAPIGenerator) generateDescription(endpoint *types.Endpoint, namespace *types.Namespace) string {
	desc := endpoint.Description
	if desc == "" {
		desc = fmt.Sprintf("API endpoint for namespace: %s", namespace.Name)
	}

	if namespace.Description != "" {
		desc += fmt.Sprintf("\n\n%s", namespace.Description)
	}

	return desc
}

func (g *OpenAPIGenerator) generateComponents(endpoint *types.Endpoint) Components {
	components := Components{
		Schemas:         make(map[string]Schema),
		SecuritySchemes: make(map[string]SecurityScheme),
	}

	// Add common schemas
	components.Schemas["ToolRequest"] = Schema{
		Type: "object",
		Properties: map[string]Schema{
			"arguments": {
				Type: "object",
			},
		},
	}

	components.Schemas["ToolResponse"] = Schema{
		Type: "object",
		Properties: map[string]Schema{
			"success": {
				Type: "boolean",
			},
			"result": {
				Type: "object",
			},
			"error": {
				Type: "string",
			},
		},
	}

	components.Schemas["Error"] = Schema{
		Type: "object",
		Properties: map[string]Schema{
			"error": {
				Type: "string",
			},
			"details": {
				Type: "string",
			},
		},
	}

	// Add security schemes based on endpoint configuration
	if endpoint.EnableAPIKeyAuth {
		components.SecuritySchemes["api_key"] = SecurityScheme{
			Type:        "apiKey",
			Description: "API Key authentication",
			Name:        "Authorization",
			In:          "header",
		}

		if endpoint.UseQueryParamAuth {
			components.SecuritySchemes["api_key_query"] = SecurityScheme{
				Type:        "apiKey",
				Description: "API Key authentication via query parameter",
				Name:        "api_key",
				In:          "query",
			}
		}
	}

	if endpoint.EnableOAuth {
		components.SecuritySchemes["oauth2"] = SecurityScheme{
			Type:        "oauth2",
			Description: "OAuth 2.0 authentication",
			Flows: &Flows{
				AuthorizationCode: &Flow{
					AuthorizationURL: fmt.Sprintf("%s/oauth/authorize", g.baseURL),
					TokenURL:         fmt.Sprintf("%s/oauth/token", g.baseURL),
					Scopes: map[string]string{
						"read":  "Read access to tools",
						"write": "Execute tools",
					},
				},
			},
		}
	}

	return components
}

func (g *OpenAPIGenerator) generateSecurityRequirements(endpoint *types.Endpoint) []SecurityRequirement {
	var requirements []SecurityRequirement

	if endpoint.EnableAPIKeyAuth {
		requirements = append(requirements, SecurityRequirement{
			"api_key": {},
		})

		if endpoint.UseQueryParamAuth {
			requirements = append(requirements, SecurityRequirement{
				"api_key_query": {},
			})
		}
	}

	if endpoint.EnableOAuth {
		requirements = append(requirements, SecurityRequirement{
			"oauth2": {"read", "write"},
		})
	}

	return requirements
}

func (g *OpenAPIGenerator) generateToolPath(tool types.NamespaceTool, endpoint *types.Endpoint) PathItem {
	operation := &Operation{
		Summary:     fmt.Sprintf("Execute %s", tool.ToolName),
		Description: tool.Description,
		OperationID: g.sanitizeOperationID(tool.ToolName),
		RequestBody: &RequestBody{
			Description: "Tool execution parameters",
			Required:    true,
			Content: map[string]MediaType{
				"application/json": {
					Schema: Schema{
						Ref: "#/components/schemas/ToolRequest",
					},
				},
			},
		},
		Responses: map[string]Response{
			"200": {
				Description: "Successful tool execution",
				Content: map[string]MediaType{
					"application/json": {
						Schema: Schema{
							Ref: "#/components/schemas/ToolResponse",
						},
					},
				},
			},
			"400": {
				Description: "Bad request",
				Content: map[string]MediaType{
					"application/json": {
						Schema: Schema{
							Ref: "#/components/schemas/Error",
						},
					},
				},
			},
			"401": {
				Description: "Unauthorized",
				Content: map[string]MediaType{
					"application/json": {
						Schema: Schema{
							Ref: "#/components/schemas/Error",
						},
					},
				},
			},
			"500": {
				Description: "Internal server error",
				Content: map[string]MediaType{
					"application/json": {
						Schema: Schema{
							Ref: "#/components/schemas/Error",
						},
					},
				},
			},
		},
	}

	// Add security if not public
	if !endpoint.EnablePublicAccess {
		operation.Security = g.generateSecurityRequirements(endpoint)
	}

	return PathItem{
		Post: operation,
	}
}

func (g *OpenAPIGenerator) generateToolsListPath(endpoint *types.Endpoint) PathItem {
	operation := &Operation{
		Summary:     "List available tools",
		Description: "Get a list of all available tools in this namespace",
		OperationID: "listTools",
		Responses: map[string]Response{
			"200": {
				Description: "List of available tools",
				Content: map[string]MediaType{
					"application/json": {
						Schema: Schema{
							Type: "object",
							Properties: map[string]Schema{
								"tools": {
									Type: "array",
									Items: &Schema{
										Type: "object",
										Properties: map[string]Schema{
											"name": {
												Type: "string",
											},
											"description": {
												Type: "string",
											},
											"server": {
												Type: "string",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"401": {
				Description: "Unauthorized",
				Content: map[string]MediaType{
					"application/json": {
						Schema: Schema{
							Ref: "#/components/schemas/Error",
						},
					},
				},
			},
		},
	}

	// Add security if not public
	if !endpoint.EnablePublicAccess {
		operation.Security = g.generateSecurityRequirements(endpoint)
	}

	return PathItem{
		Get: operation,
	}
}

func (g *OpenAPIGenerator) sanitizeOperationID(toolName string) string {
	// Replace non-alphanumeric characters with underscores
	var result strings.Builder
	for _, ch := range toolName {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
		   (ch >= '0' && ch <= '9') {
			result.WriteRune(ch)
		} else {
			result.WriteRune('_')
		}
	}
	return result.String()
}
