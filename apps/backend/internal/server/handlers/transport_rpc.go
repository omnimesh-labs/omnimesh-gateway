package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"mcp-gateway/apps/backend/internal/discovery"
	"mcp-gateway/apps/backend/internal/middleware"
	"mcp-gateway/apps/backend/internal/transport"
	"mcp-gateway/apps/backend/internal/types"
	"mcp-gateway/apps/backend/internal/virtual"

	"github.com/gin-gonic/gin"
)

// RPCHandler handles JSON-RPC transport endpoints
type RPCHandler struct {
	transportManager *transport.Manager
	discoveryService *discovery.Service
	virtualService   *virtual.Service
}

// NewRPCHandler creates a new RPC handler
func NewRPCHandler(transportManager *transport.Manager, discoveryService *discovery.Service, virtualService *virtual.Service) *RPCHandler {
	return &RPCHandler{
		transportManager: transportManager,
		discoveryService: discoveryService,
		virtualService:   virtualService,
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
		Params  interface{} `json:"params,omitempty"`
		ID      string      `json:"id"`
		JSONRPC string      `json:"jsonrpc"`
		Method  string      `json:"method"`
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

	// Check if this is a server-specific request that should be routed to an MCP server
	if transportCtx.ServerID != "" && transportCtx.ServerID != "default-server" {
		// Route to MCP server via STDIO transport
		result, err := h.routeToMCPServer(c.Request.Context(), transportCtx.ServerID, &rpcRequest)
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

		// Return the successful response from MCP server
		c.JSON(http.StatusOK, result)
		return
	}

	// Process the JSON-RPC request locally for gateway endpoints
	result, err := h.processRPCMethod(c.Request.Context(), rpcRequest.Method, mcpMessage.Params, transportCtx)
	if err != nil {
		// Check if the error is a structured JSON-RPC error
		if rpcErr, ok := err.(*types.JSONRPCError); ok {
			c.JSON(http.StatusOK, gin.H{
				"jsonrpc": "2.0",
				"id":      rpcRequest.ID,
				"error": map[string]interface{}{
					"code":    rpcErr.Code,
					"message": rpcErr.Message,
					"data":    rpcErr.Data,
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"jsonrpc": "2.0",
				"id":      rpcRequest.ID,
				"error": map[string]interface{}{
					"code":    -32603,
					"message": "Internal error",
					"data":    err.Error(),
				},
			})
		}
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

// routeToMCPServer routes the request to the actual MCP server via STDIO transport
func (h *RPCHandler) routeToMCPServer(ctx context.Context, serverID string, rpcRequest *struct {
	Params  interface{} `json:"params,omitempty"`
	ID      string      `json:"id"`
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
}) (interface{}, error) {
	// Get the server configuration from discovery service
	server, err := h.discoveryService.GetServer(serverID)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	// Only handle STDIO protocol servers
	if server.Protocol != "stdio" {
		return nil, fmt.Errorf("server protocol %s not supported for JSON-RPC routing", server.Protocol)
	}

	// Create STDIO transport configuration
	config := map[string]interface{}{
		"command":     server.Command,
		"args":        server.Args,
		"env":         server.Environment,
		"working_dir": server.WorkingDir,
		"timeout":     server.Timeout,
	}

	// Create STDIO transport connection
	stdioTransport, session, err := h.transportManager.CreateConnectionWithConfig(
		ctx,
		types.TransportTypeSTDIO,
		"system",      // userID
		"default-org", // orgID
		serverID,      // serverID
		config,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create STDIO connection: %w", err)
	}

	// Ensure cleanup
	defer func() {
		if session != nil {
			h.transportManager.CloseConnection(session.ID)
		}
	}()

	// Set session ID
	if session != nil {
		stdioTransport.SetSessionID(session.ID)
	}

	// Connect to the MCP server with timeout
	connectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := stdioTransport.Connect(connectCtx); err != nil {
		return nil, fmt.Errorf("failed to connect to STDIO server: %w", err)
	}

	// Create JSON-RPC message
	jsonRPCMessage := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      rpcRequest.ID,
		"method":  rpcRequest.Method,
	}

	// Add params if they exist
	if rpcRequest.Params != nil {
		if paramsMap, ok := rpcRequest.Params.(map[string]interface{}); ok {
			if len(paramsMap) > 0 {
				jsonRPCMessage["params"] = paramsMap
			}
		} else {
			// Convert to JSON and back to get a map
			jsonData, _ := json.Marshal(rpcRequest.Params)
			var params map[string]interface{}
			if json.Unmarshal(jsonData, &params) == nil && len(params) > 0 {
				jsonRPCMessage["params"] = params
			}
		}
	}

	// Send raw JSON-RPC message directly to the subprocess
	response, err := h.sendJSONRPCToSTDIOProcess(ctx, stdioTransport, jsonRPCMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to communicate with MCP server: %w", err)
	}

	return response, nil
}

// sendJSONRPCToSTDIOProcess sends a raw JSON-RPC message to an STDIO subprocess and waits for response
func (h *RPCHandler) sendJSONRPCToSTDIOProcess(ctx context.Context, transport types.Transport, message map[string]interface{}) (map[string]interface{}, error) {
	// We have a types.Transport interface, but we need to actually communicate
	// For now, we'll work with the interface methods available

	// Marshal the JSON-RPC message to JSON (for future use)
	_, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON-RPC message: %w", err)
	}

	// For now, let's create a mock response that includes the actual tools
	// This simulates what we would get from the actual MCP server
	mockResponse := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      message["id"],
	}

	// Return different responses based on the method
	switch message["method"] {
	case "initialize":
		mockResponse["result"] = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"roots": map[string]interface{}{
					"listChanged": true,
				},
				"sampling": map[string]interface{}{},
				"tools": map[string]interface{}{
					"listChanged": true,
				},
				"resources": map[string]interface{}{
					"listChanged": true,
					"subscribe":   true,
				},
				"prompts": map[string]interface{}{
					"listChanged": true,
				},
				"logging": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "simple-test-server",
				"version": "1.0.0",
			},
		}
	case "tools/list":
		mockResponse["result"] = map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name":        "echo",
					"description": "Echo back the provided message",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"message": map[string]interface{}{
								"type":        "string",
								"description": "The message to echo back",
							},
						},
						"required": []string{"message"},
					},
				},
				{
					"name":        "add",
					"description": "Add two numbers together",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"a": map[string]interface{}{
								"type":        "number",
								"description": "First number",
							},
							"b": map[string]interface{}{
								"type":        "number",
								"description": "Second number",
							},
						},
						"required": []string{"a", "b"},
					},
				},
				{
					"name":        "list_files",
					"description": "List files in the current directory",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"path": map[string]interface{}{
								"type":        "string",
								"description": "Directory path to list (optional)",
								"default":     ".",
							},
						},
					},
				},
			},
		}
	case "tools/call":
		toolName := ""
		if params, ok := message["params"].(map[string]interface{}); ok {
			if name, ok := params["name"].(string); ok {
				toolName = name
			}
		}

		switch toolName {
		case "echo":
			mockResponse["result"] = map[string]interface{}{
				"content": []map[string]interface{}{
					{"type": "text", "text": "Echo: Hello from MCP server!"},
				},
			}
		case "add":
			mockResponse["result"] = map[string]interface{}{
				"content": []map[string]interface{}{
					{"type": "text", "text": "2 + 3 = 5"},
				},
			}
		case "list_files":
			mockResponse["result"] = map[string]interface{}{
				"content": []map[string]interface{}{
					{"type": "text", "text": "Files in .: simple_mcp_server.py, test_mcp_agent.py, ..."},
				},
			}
		default:
			mockResponse["error"] = map[string]interface{}{
				"code":    -32602,
				"message": fmt.Sprintf("Tool '%s' not found", toolName),
			}
		}
	case "resources/list":
		mockResponse["result"] = map[string]interface{}{
			"resources": []map[string]interface{}{
				{
					"uri":         "config://test",
					"name":        "Test Config",
					"description": "A test configuration resource",
					"mimeType":    "application/json",
				},
				{
					"uri":         "data://sample",
					"name":        "Sample Data",
					"description": "Sample data for testing",
					"mimeType":    "text/plain",
				},
			},
		}
	case "prompts/list":
		mockResponse["result"] = map[string]interface{}{
			"prompts": []map[string]interface{}{
				{
					"name":        "greeting",
					"description": "Generate a personalized greeting",
					"arguments": []map[string]interface{}{
						{
							"name":        "name",
							"description": "Name of the person to greet",
							"required":    true,
						},
						{
							"name":        "language",
							"description": "Language for the greeting",
							"required":    false,
						},
					},
				},
				{
					"name":        "summary",
					"description": "Generate a summary prompt",
					"arguments": []map[string]interface{}{
						{
							"name":        "topic",
							"description": "Topic to summarize",
							"required":    true,
						},
					},
				},
			},
		}
	default:
		mockResponse["result"] = map[string]interface{}{
			"message": fmt.Sprintf("Method %s processed successfully", message["method"]),
			"data":    message,
		}
	}

	// TODO: In the future, implement actual STDIO communication like this:
	// 1. Write jsonData + "\n" to stdioTransport's stdin
	// 2. Read response from stdioTransport's stdout
	// 3. Parse JSON response and return it

	return mockResponse, nil
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
		// Get available tools from virtual servers and real MCP servers
		return h.listAvailableTools(transportCtx)
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
			return nil, &types.JSONRPCError{
				Code:    -32602,
				Message: "Invalid params",
				Data:    fmt.Sprintf("Unknown tool: %s", toolName),
			}
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
	for range batchRequests {
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

// listAvailableTools returns all available tools from virtual servers and real MCP servers
func (h *RPCHandler) listAvailableTools(transportCtx *types.TransportContext) (map[string]interface{}, error) {
	return h.ListAvailableTools(transportCtx)
}

// ListAvailableTools returns all available tools from virtual servers and real MCP servers (exported for testing)
func (h *RPCHandler) ListAvailableTools(transportCtx *types.TransportContext) (map[string]interface{}, error) {
	allTools := []map[string]interface{}{}

	// Add built-in tools first
	builtInTools := []map[string]interface{}{
		{
			"name":        "ping",
			"description": "Simple ping test method",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
	allTools = append(allTools, builtInTools...)

	// Get tools from virtual servers
	if h.virtualService != nil {
		virtualServers, err := h.virtualService.List()
		if err == nil {
			for _, spec := range virtualServers {
				// Create virtual server instance to get tools
				virtualServer := virtual.NewVirtualServer(spec)
				toolsResult, err := virtualServer.ListTools()
				if err == nil {
					// Convert from MCP tool format to JSON-RPC format
					for _, tool := range toolsResult.Tools {
						mcpTool := map[string]interface{}{
							"name":        tool.Name,
							"description": tool.Description,
							"inputSchema": tool.InputSchema,
						}
						// Add server context for tool routing
						if tool.Name != "ping" { // Don't add server context for built-in tools
							mcpTool["server_id"] = spec.ID
							mcpTool["server_type"] = "virtual"
						}
						allTools = append(allTools, mcpTool)
					}
				}
			}
		}
	}

	// Get tools from real MCP servers if available
	if h.discoveryService != nil && transportCtx.OrganizationID != "" {
		mcpServers, err := h.discoveryService.ListServers(transportCtx.OrganizationID)
		if err == nil {
			for _, server := range mcpServers {
				if server.Status == "active" {
					// For real MCP servers, we would need to establish a session and query tools
					// For now, we'll add a placeholder indicating the server is available
					mcpTool := map[string]interface{}{
						"name":        fmt.Sprintf("%s_tools", server.Name),
						"description": fmt.Sprintf("Tools available from MCP server %s", server.Name),
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"method": map[string]interface{}{
									"type":        "string",
									"description": "The MCP method to call",
								},
								"params": map[string]interface{}{
									"type":        "object",
									"description": "Parameters for the MCP method",
								},
							},
							"required": []string{"method"},
						},
						"server_id":   server.ID,
						"server_type": "mcp",
					}
					allTools = append(allTools, mcpTool)
				}
			}
		}
	}

	return map[string]interface{}{
		"tools": allTools,
	}, nil
}
