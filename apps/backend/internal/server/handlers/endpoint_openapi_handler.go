package handlers

import (
	"context"
	"fmt"
	"mcp-gateway/apps/backend/internal/services"
	"mcp-gateway/apps/backend/internal/types"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// OpenAPIService defines the interface for OpenAPI operations
type OpenAPIService interface {
	GenerateSpec(endpoint *types.Endpoint, namespace *types.Namespace, tools []types.NamespaceTool) *services.OpenAPISpec
}

// HandleEndpointOpenAPI handles OpenAPI spec generation for endpoints
func HandleEndpointOpenAPI(endpointService EndpointService, namespaceService NamespaceService, baseURL string) gin.HandlerFunc {
	generator := services.NewOpenAPIGenerator(baseURL)

	return func(c *gin.Context) {
		endpointName := c.Param("endpoint_name")
		if endpointName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Endpoint name required"})
			return
		}

		// Get endpoint config
		config, err := endpointService.ResolveEndpoint(c.Request.Context(), endpointName)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Endpoint not found"})
			return
		}

		// Get tools for the namespace
		tools, err := namespaceService.AggregateTools(c.Request.Context(), config.Namespace.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tools"})
			return
		}

		// Generate OpenAPI spec
		spec := generator.GenerateSpec(config.Endpoint, config.Namespace, tools)

		// Return based on path
		if strings.HasSuffix(c.Request.URL.Path, "/openapi.json") {
			c.JSON(http.StatusOK, spec)
		} else if strings.HasSuffix(c.Request.URL.Path, "/docs") {
			// Serve Swagger UI HTML
			swaggerHTML := generateSwaggerHTML(endpointName, config.Endpoint.Name)
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerHTML))
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		}
	}
}

// HandleEndpointToolsList handles GET /metamcp/:endpoint_name/api/tools
func HandleEndpointToolsList(namespaceService NamespaceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespaceVal, exists := c.Get("namespace")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Namespace not found in context"})
			return
		}
		namespace := namespaceVal.(*types.Namespace)

		// Get tools for the namespace
		tools, err := namespaceService.AggregateTools(c.Request.Context(), namespace.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tools"})
			return
		}

		// Format tools for API response
		toolList := make([]map[string]interface{}, len(tools))
		for i, tool := range tools {
			toolList[i] = map[string]interface{}{
				"name":        tool.ToolName,
				"description": tool.Description,
				"server":      tool.ServerName,
				"status":      tool.Status,
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"tools": toolList,
			"total": len(toolList),
		})
	}
}

func generateSwaggerHTML(endpointName, title string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>%s - API Documentation</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui.css">
    <style>
        body {
            margin: 0;
            padding: 0;
        }
        #swagger-ui {
            max-width: 1200px;
            margin: 0 auto;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: "/api/public/endpoints/%s/api/openapi.json",
                dom_id: '#swagger-ui',
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                layout: "StandaloneLayout",
                deepLinking: true,
                persistAuthorization: true
            });
        };
    </script>
</body>
</html>
`, title, endpointName)
}

// EndpointOpenAPIService wrapper to implement the interface
type EndpointOpenAPIService struct {
	endpointService  EndpointService
	namespaceService NamespaceService
}

// NewEndpointOpenAPIService creates a new OpenAPI service
func NewEndpointOpenAPIService(endpointService EndpointService, namespaceService NamespaceService) *EndpointOpenAPIService {
	return &EndpointOpenAPIService{
		endpointService:  endpointService,
		namespaceService: namespaceService,
	}
}

// GenerateOpenAPISpec generates an OpenAPI specification for an endpoint
func (s *EndpointOpenAPIService) GenerateOpenAPISpec(ctx context.Context, endpointName string) (*services.OpenAPISpec, error) {
	// Get endpoint config
	config, err := s.endpointService.ResolveEndpoint(ctx, endpointName)
	if err != nil {
		return nil, fmt.Errorf("endpoint not found: %w", err)
	}

	// Get tools for the namespace
	tools, err := s.namespaceService.AggregateTools(ctx, config.Namespace.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tools: %w", err)
	}

	// Generate OpenAPI spec
	generator := services.NewOpenAPIGenerator("http://localhost:8080") // TODO: Get from config
	spec := generator.GenerateSpec(config.Endpoint, config.Namespace, tools)

	return spec, nil
}
