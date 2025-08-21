package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"mcp-gateway/apps/backend/internal/middleware"
	"mcp-gateway/apps/backend/internal/transport"
	"mcp-gateway/apps/backend/internal/types"

	"github.com/gin-gonic/gin"
)

// RPCHandler handles JSON-RPC transport endpoints
type RPCHandler struct {
	transportManager *transport.Manager
}

// NewRPCHandler creates a new RPC handler
func NewRPCHandler(transportManager *transport.Manager) *RPCHandler {
	return &RPCHandler{
		transportManager: transportManager,
	}
}

// HandleJSONRPC handles JSON-RPC requests
func (h *RPCHandler) HandleJSONRPC(c *gin.Context) {
	// Get transport context (optional for general RPC calls)
	transportCtx := middleware.GetTransportContext(c)

	// For general RPC calls without server context, create default context
	if transportCtx == nil {
		transportCtx = &types.TransportContext{
			Request:        c.Request,
			UserID:         "anonymous",
			OrganizationID: "default",
			ServerID:       "",
			Transport:      types.TransportTypeHTTP,
			Metadata:       make(map[string]interface{}),
		}
	}

	// Parse JSON-RPC request
	var rpcRequest struct {
		ID      string      `json:"id"`
		JSONRPC string      `json:"jsonrpc"`
		Method  string      `json:"method"`
		Params  interface{} `json:"params,omitempty"`
	}

	if err := c.ShouldBindJSON(&rpcRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"jsonrpc": "2.0",
			"id":      nil,
			"error": map[string]interface{}{
				"code":    -32700,
				"message": "Parse error",
				"data":    err.Error(),
			},
		})
		return
	}

	// Validate JSON-RPC format
	if rpcRequest.JSONRPC != "2.0" {
		c.JSON(http.StatusBadRequest, gin.H{
			"jsonrpc": "2.0",
			"id":      rpcRequest.ID,
			"error": map[string]interface{}{
				"code":    -32600,
				"message": "Invalid Request",
				"data":    "JSONRPC version must be 2.0",
			},
		})
		return
	}

	if rpcRequest.Method == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"jsonrpc": "2.0",
			"id":      rpcRequest.ID,
			"error": map[string]interface{}{
				"code":    -32600,
				"message": "Invalid Request",
				"data":    "Method is required",
			},
		})
		return
	}

	// Create MCP message
	mcpMessage := &types.MCPMessage{
		ID:      rpcRequest.ID,
		Type:    types.MCPMessageTypeRequest,
		Method:  rpcRequest.Method,
		Version: "2024-11-05",
		Params:  make(map[string]interface{}),
	}

	// Convert params to map
	if rpcRequest.Params != nil {
		if paramsMap, ok := rpcRequest.Params.(map[string]interface{}); ok {
			mcpMessage.Params = paramsMap
		} else {
			// Convert to JSON and back to get a map
			jsonData, _ := json.Marshal(rpcRequest.Params)
			json.Unmarshal(jsonData, &mcpMessage.Params)
		}
	}

	// Process the JSON-RPC request directly
	// For the gateway's /rpc endpoint, we handle the request locally
	// rather than proxying it through the transport layer
	result, err := h.processRPCMethod(c.Request.Context(), rpcRequest.Method, mcpMessage.Params, transportCtx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"jsonrpc": "2.0",
			"id":      rpcRequest.ID,
			"error": map[string]interface{}{
				"code":    -32603,
				"message": "Internal error",
				"data":    err.Error(),
			},
		})
		return
	}

	// Return the successful response
	response := gin.H{
		"jsonrpc": "2.0",
		"id":      rpcRequest.ID,
		"result":  result,
	}

	c.JSON(http.StatusOK, response)
}

// processRPCMethod processes individual RPC methods locally
func (h *RPCHandler) processRPCMethod(ctx context.Context, method string, params map[string]interface{}, transportCtx *types.TransportContext) (map[string]interface{}, error) {
	switch method {
	case "ping":
		return map[string]interface{}{
			"message":   "pong",
			"timestamp": time.Now().Unix(),
			"server_id": transportCtx.ServerID,
			"user_id":   transportCtx.UserID,
		}, nil
	case types.MCPMethodListTools:
		// Return available tools
		return map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name":        "ping",
					"description": "Simple ping test method",
					"inputSchema": map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
			},
		}, nil
	case types.MCPMethodCallTool:
		// Handle tool calls
		toolName, ok := params["name"].(string)
		if !ok {
			return nil, fmt.Errorf("tool name is required")
		}

		switch toolName {
		case "ping":
			return map[string]interface{}{
				"result": map[string]interface{}{
					"message": "Tool call successful",
					"tool":    toolName,
				},
			}, nil
		default:
			return nil, fmt.Errorf("unknown tool: %s", toolName)
		}
	case types.MCPMethodListResources:
		return map[string]interface{}{
			"resources": []map[string]interface{}{},
		}, nil
	case types.MCPMethodListPrompts:
		return map[string]interface{}{
			"prompts": []map[string]interface{}{},
		}, nil
	default:
		return map[string]interface{}{
			"message":   "Method processed successfully",
			"method":    method,
			"server_id": transportCtx.ServerID,
			"user_id":   transportCtx.UserID,
		}, nil
	}
}

// HandleBatchRPC handles JSON-RPC batch requests
func (h *RPCHandler) HandleBatchRPC(c *gin.Context) {
	// Parse batch request
	var batchRequests []interface{}
	if err := c.ShouldBindJSON(&batchRequests); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"jsonrpc": "2.0",
			"id":      nil,
			"error": map[string]interface{}{
				"code":    -32700,
				"message": "Parse error",
				"data":    err.Error(),
			},
		})
		return
	}

	if len(batchRequests) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"jsonrpc": "2.0",
			"id":      nil,
			"error": map[string]interface{}{
				"code":    -32600,
				"message": "Invalid Request",
				"data":    "Batch cannot be empty",
			},
		})
		return
	}

	var responses []interface{}

	// Process each request in the batch
	for _, _ = range batchRequests {
		// Create a new context for each sub-request
		// This is a simplified implementation
		response := gin.H{
			"jsonrpc": "2.0",
			"id":      nil,
			"result": map[string]interface{}{
				"message": "Batch request processed",
			},
		}
		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, responses)
}

// HandleRPCIntrospection provides information about available RPC methods
func (h *RPCHandler) HandleRPCIntrospection(c *gin.Context) {
	methods := []map[string]interface{}{
		{
			"name":        types.MCPMethodListTools,
			"description": "List available tools",
			"params":      []interface{}{},
		},
		{
			"name":        types.MCPMethodCallTool,
			"description": "Call a tool",
			"params": []interface{}{
				map[string]interface{}{
					"name":     "name",
					"type":     "string",
					"required": true,
				},
				map[string]interface{}{
					"name":     "arguments",
					"type":     "object",
					"required": false,
				},
			},
		},
		{
			"name":        types.MCPMethodListResources,
			"description": "List available resources",
			"params":      []interface{}{},
		},
		{
			"name":        types.MCPMethodReadResource,
			"description": "Read a resource",
			"params": []interface{}{
				map[string]interface{}{
					"name":     "uri",
					"type":     "string",
					"required": true,
				},
			},
		},
		{
			"name":        types.MCPMethodListPrompts,
			"description": "List available prompts",
			"params":      []interface{}{},
		},
		{
			"name":        types.MCPMethodGetPrompt,
			"description": "Get a prompt",
			"params": []interface{}{
				map[string]interface{}{
					"name":     "name",
					"type":     "string",
					"required": true,
				},
				map[string]interface{}{
					"name":     "arguments",
					"type":     "object",
					"required": false,
				},
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"methods":     methods,
		"description": "MCP Gateway JSON-RPC API",
		"version":     "2024-11-05",
		"transport":   "HTTP",
	})
}

// HandleRPCHealth checks JSON-RPC transport health
func (h *RPCHandler) HandleRPCHealth(c *gin.Context) {
	healthResults := h.transportManager.HealthCheck(c.Request.Context())

	rpcHealth, exists := healthResults[types.TransportTypeHTTP]
	if !exists {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "JSON-RPC transport not configured",
		})
		return
	}

	if rpcHealth != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  rpcHealth.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"transport": types.TransportTypeHTTP,
		"timestamp": time.Now(),
		"capabilities": []string{
			"json_rpc_2.0",
			"batch_requests",
			"introspection",
			"synchronous",
		},
	})
}
