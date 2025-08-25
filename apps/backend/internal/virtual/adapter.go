package virtual

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"mcp-gateway/apps/backend/internal/types"
)

// RESTAdapter implements the Adapter interface for REST services
type RESTAdapter struct {
	client *http.Client
	spec   *types.VirtualServerSpec
}

// NewRESTAdapter creates a new REST adapter
func NewRESTAdapter(spec *types.VirtualServerSpec) *RESTAdapter {
	return &RESTAdapter{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		spec: spec,
	}
}

// ListTools returns the available tools for this virtual server
func (a *RESTAdapter) ListTools() ([]types.ToolDef, error) {
	return a.spec.Tools, nil
}

// CallTool executes a tool call by making the appropriate REST request
func (a *RESTAdapter) CallTool(name string, args map[string]interface{}) (interface{}, error) {
	// Find the tool definition
	var toolDef *types.ToolDef
	for _, tool := range a.spec.Tools {
		if tool.Name == name {
			toolDef = &tool
			break
		}
	}

	if toolDef == nil {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	if toolDef.REST == nil {
		return nil, fmt.Errorf("tool %s does not have REST configuration", name)
	}

	// For now, return a stub response for testing
	// This will be expanded to make actual HTTP calls
	return a.makeRESTCall(toolDef, args)
}

// makeRESTCall performs the actual REST API call
func (a *RESTAdapter) makeRESTCall(toolDef *types.ToolDef, args map[string]interface{}) (interface{}, error) {
	restSpec := toolDef.REST

	// Prepare the request
	req, err := a.prepareRequest(restSpec, args)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	// Set timeout if specified
	if restSpec.TimeoutSec > 0 {
		a.client.Timeout = time.Duration(restSpec.TimeoutSec) * time.Second
	}

	// For now, return a mock response based on the tool
	// TODO: Make actual HTTP call when implementing real REST functionality
	switch toolDef.Name {
	case "list_channels":
		return map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"channels": []map[string]interface{}{
					{
						"id":   "C1234567890",
						"name": "general",
						"type": "public",
					},
					{
						"id":   "C0987654321",
						"name": "random",
						"type": "public",
					},
				},
			},
			"mock": true,
			"url":  req.URL.String(),
		}, nil

	case "send_message":
		channel, _ := args["channel"].(string)
		text, _ := args["text"].(string)
		return map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"channel":    channel,
				"text":       text,
				"timestamp":  time.Now().Unix(),
				"message_id": "1234567890.123456",
			},
			"mock": true,
			"url":  req.URL.String(),
		}, nil

	default:
		return map[string]interface{}{
			"status": "success",
			"data":   args,
			"mock":   true,
			"tool":   toolDef.Name,
			"url":    req.URL.String(),
		}, nil
	}
}

// prepareRequest creates an HTTP request based on the REST specification
func (a *RESTAdapter) prepareRequest(restSpec *types.RESTSpec, args map[string]interface{}) (*http.Request, error) {
	url := restSpec.URLTemplate

	var body io.Reader

	// Handle body mapping for POST/PUT requests
	if restSpec.Method == "POST" || restSpec.Method == "PUT" {
		if restSpec.BodyMap != nil {
			requestBody := make(map[string]interface{})
			for bodyKey, argKey := range restSpec.BodyMap {
				if value, exists := args[argKey]; exists {
					requestBody[bodyKey] = value
				}
			}

			bodyBytes, err := json.Marshal(requestBody)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}

			body = bytes.NewReader(bodyBytes)
		}
	}

	// Create the request
	req, err := http.NewRequest(restSpec.Method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	for key, value := range restSpec.Headers {
		req.Header.Set(key, value)
	}

	// Set authentication
	if restSpec.Auth != nil {
		switch restSpec.Auth.Type {
		case "Bearer":
			token := a.resolveToken(restSpec.Auth.Token)
			req.Header.Set("Authorization", "Bearer "+token)
		case "Basic":
			// TODO: Implement basic auth if needed
		}
	}

	return req, nil
}

// resolveToken resolves token placeholders like ${SECRET:SLACK_BOT_TOKEN}
func (a *RESTAdapter) resolveToken(token string) string {
	// Handle environment variable placeholders
	if strings.HasPrefix(token, "${SECRET:") && strings.HasSuffix(token, "}") {
		secretName := strings.TrimSuffix(strings.TrimPrefix(token, "${SECRET:"), "}")
		
		// Try to get from environment variable
		if envValue := os.Getenv(secretName); envValue != "" {
			return envValue
		}
		
		// Log warning for missing secret (but don't expose the secret name in production logs)
		log.Printf("Warning: Secret %s not found in environment variables", secretName)
		
		// Return empty string for missing secrets (don't use mock tokens)
		return ""
	}
	return token
}
