package handlers

import (
	"fmt"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// HandleEndpointSSE handles SSE connections for endpoints
func HandleEndpointSSE(namespaceService NamespaceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespaceVal, exists := c.Get("namespace")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Namespace not found in context"})
			return
		}
		namespace := namespaceVal.(*types.Namespace)

		// Set SSE headers
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")

		// Create SSE connection
		flusher, ok := c.Writer.(http.Flusher)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
			return
		}

		// Get or create session ID
		sessionID := c.GetHeader("mcp-session-id")
		if sessionID == "" {
			sessionID = uuid.New().String()
			c.Writer.Header().Set("mcp-session-id", sessionID)
		}

		// Send initial connection event
		fmt.Fprintf(c.Writer, "event: connected\n")
		fmt.Fprintf(c.Writer, "data: {\"session_id\":\"%s\",\"namespace_id\":\"%s\"}\n\n", sessionID, namespace.ID)
		flusher.Flush()

		// Keep connection alive with periodic pings
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		// Handle client disconnect
		clientDisconnect := c.Request.Context().Done()

		for {
			select {
			case <-ticker.C:
				// Send ping event
				fmt.Fprintf(c.Writer, "event: ping\n")
				fmt.Fprintf(c.Writer, "data: {\"timestamp\":%d}\n\n", time.Now().Unix())
				flusher.Flush()

			case <-clientDisconnect:
				// Client disconnected
				return
			}
		}
	}
}

// HandleEndpointSSEMessage handles messages sent to SSE endpoints
func HandleEndpointSSEMessage(namespaceService NamespaceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespaceVal, exists := c.Get("namespace")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Namespace not found in context"})
			return
		}
		namespace := namespaceVal.(*types.Namespace)

		// Get session ID
		sessionID := c.GetHeader("mcp-session-id")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID required"})
			return
		}

		// Parse request body
		var message map[string]interface{}
		if err := c.ShouldBindJSON(&message); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message format"})
			return
		}

		// Process message based on type
		messageType, ok := message["type"].(string)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Message type required"})
			return
		}

		// Route message to appropriate handler
		switch messageType {
		case "tool_call":
			// Handle tool execution
			toolName, _ := message["tool"].(string)
			args, _ := message["arguments"].(map[string]interface{})

			result, err := namespaceService.ExecuteTool(c.Request.Context(), namespace.ID, types.ExecuteNamespaceToolRequest{
				Tool:      toolName,
				Arguments: args,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, result)

		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown message type"})
		}
	}
}

// HandleEndpointHTTP handles HTTP/MCP protocol requests for endpoints
func HandleEndpointHTTP(namespaceService NamespaceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespaceVal, exists := c.Get("namespace")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Namespace not found in context"})
			return
		}
		namespace := namespaceVal.(*types.Namespace)

		// Handle different HTTP methods
		switch c.Request.Method {
		case "HEAD":
			// Handle HEAD requests for OAuth discovery
			c.Header("Content-Type", "application/json")
			c.Status(http.StatusOK)
			return

		case "GET":
			// Handle GET requests - return server capabilities
			c.JSON(http.StatusOK, gin.H{
				"jsonrpc": "2.0",
				"result": gin.H{
					"capabilities": gin.H{
						"tools": gin.H{
							"listChanged": true,
						},
						"resources": gin.H{
							"subscribe":   true,
							"listChanged": true,
						},
					},
					"serverInfo": gin.H{
						"name":    "Omnimesh Gateway",
						"version": "1.0.0",
					},
					"protocolVersion": "2024-11-05",
				},
				"id": nil,
			})
			return

		case "POST":
			// Handle POST requests with JSON-RPC payload
			break

		case "OPTIONS":
			// Handle preflight requests
			c.Header("Access-Control-Allow-Methods", "GET, POST, HEAD, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
			c.Status(http.StatusNoContent)
			return

		default:
			c.JSON(http.StatusMethodNotAllowed, gin.H{
				"jsonrpc": "2.0",
				"error": gin.H{
					"code":    -32601,
					"message": "Method not allowed",
				},
				"id": nil,
			})
			return
		}

		// Parse MCP request for POST requests
		var request map[string]interface{}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"jsonrpc": "2.0",
				"error": map[string]interface{}{
					"code":    -32700,
					"message": "Parse error",
				},
			})
			return
		}

		// Extract method and params
		method, _ := request["method"].(string)
		params, _ := request["params"].(map[string]interface{})
		id := request["id"]

		// Route based on method
		switch method {
		case "initialize":
			// Handle MCP initialize request
			c.JSON(http.StatusOK, gin.H{
				"jsonrpc": "2.0",
				"result": gin.H{
					"capabilities": gin.H{
						"tools": gin.H{
							"listChanged": true,
						},
						"resources": gin.H{
							"subscribe":   true,
							"listChanged": true,
						},
						"prompts": gin.H{
							"listChanged": true,
						},
						"logging": gin.H{},
						"roots": gin.H{
							"listChanged": true,
						},
					},
					"serverInfo": gin.H{
						"name":    "Omnimesh Gateway",
						"version": "1.0.0",
					},
					"protocolVersion": "2024-11-05",
				},
				"id": id,
			})
			return

		case "notifications/initialized":
			// Handle MCP initialized notification (no response expected)
			c.Status(http.StatusOK)
			return

		case "tools/list":
			// Get tools from namespace (real MCP servers)
			namespacedTools, err := namespaceService.AggregateTools(c.Request.Context(), namespace.ID)
			if err != nil {
				// Log error but continue - we'll try to get virtual server tools
				fmt.Printf("Warning: failed to get namespaced tools: %v\n", err)
			}

			// Get tools from virtual servers (temporary hardcode approach for demo)
			virtualTools := []map[string]interface{}{
				{
					"name":        "get_figma_files",
					"description": "Get list of Figma files in a team",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"team_id": map[string]interface{}{
								"type":        "string",
								"description": "The Figma team ID",
							},
						},
						"required": []string{"team_id"},
					},
				},
				{
					"name":        "get_figma_file",
					"description": "Get a specific Figma file details",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"file_key": map[string]interface{}{
								"type":        "string",
								"description": "The Figma file key",
							},
						},
						"required": []string{"file_key"},
					},
				},
				{
					"name":        "list_repositories",
					"description": "List repositories for an organization or user",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"owner": map[string]interface{}{
								"type":        "string",
								"description": "The organization or user name",
							},
							"type": map[string]interface{}{
								"type":        "string",
								"enum":        []string{"all", "public", "private"},
								"default":     "public",
								"description": "Repository visibility filter",
							},
						},
						"required": []string{"owner"},
					},
				},
				{
					"name":        "get_current_time",
					"description": "Get the current time in various formats",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"timezone": map[string]interface{}{
								"type":        "string",
								"default":     "UTC",
								"description": "Timezone for the time (e.g., UTC, America/New_York)",
							},
							"format": map[string]interface{}{
								"type":        "string",
								"default":     "iso",
								"enum":        []string{"iso", "unix", "human"},
								"description": "Output format",
							},
						},
					},
				},
			}

			// Combine namespace tools and virtual tools
			allTools := virtualTools
			for _, tool := range namespacedTools {
				allTools = append(allTools, map[string]interface{}{
					"name":        tool.PrefixedName,
					"description": fmt.Sprintf("%s (from %s)", tool.ToolName, tool.ServerName),
				})
			}

			c.JSON(http.StatusOK, gin.H{
				"jsonrpc": "2.0",
				"result": map[string]interface{}{
					"tools": allTools,
				},
				"id": id,
			})

		case "tools/call":
			// Execute tool
			toolName, _ := params["name"].(string)
			arguments, _ := params["arguments"].(map[string]interface{})

			result, err := namespaceService.ExecuteTool(c.Request.Context(), namespace.ID, types.ExecuteNamespaceToolRequest{
				Tool:      toolName,
				Arguments: arguments,
			})
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"jsonrpc": "2.0",
					"error": map[string]interface{}{
						"code":    -32603,
						"message": "Tool execution failed",
						"data":    err.Error(),
					},
					"id": id,
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"jsonrpc": "2.0",
				"result":  result,
				"id":      id,
			})

		default:
			c.JSON(http.StatusOK, gin.H{
				"jsonrpc": "2.0",
				"error": map[string]interface{}{
					"code":    -32601,
					"message": "Method not found",
				},
				"id": id,
			})
		}
	}
}

// HandleEndpointWebSocket handles WebSocket connections for endpoints
func HandleEndpointWebSocket(namespaceService NamespaceService) gin.HandlerFunc {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// Check origin based on endpoint CORS settings
			return true // TODO: Implement proper CORS check
		},
	}

	return func(c *gin.Context) {
		namespaceVal, exists := c.Get("namespace")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Namespace not found in context"})
			return
		}
		namespace := namespaceVal.(*types.Namespace)

		// Upgrade to WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "WebSocket upgrade failed"})
			return
		}
		defer conn.Close()

		// Send welcome message
		welcomeMsg := map[string]interface{}{
			"type":         "welcome",
			"namespace_id": namespace.ID,
			"session_id":   uuid.New().String(),
		}
		conn.WriteJSON(welcomeMsg)

		// Handle messages
		for {
			var message map[string]interface{}
			if err := conn.ReadJSON(&message); err != nil {
				break
			}

			// Process message based on type
			messageType, _ := message["type"].(string)

			switch messageType {
			case "ping":
				// Respond with pong
				conn.WriteJSON(map[string]interface{}{
					"type":      "pong",
					"timestamp": time.Now().Unix(),
				})

			case "tool_call":
				// Execute tool
				toolName, _ := message["tool"].(string)
				args, _ := message["arguments"].(map[string]interface{})

				result, err := namespaceService.ExecuteTool(c.Request.Context(), namespace.ID, types.ExecuteNamespaceToolRequest{
					Tool:      toolName,
					Arguments: args,
				})

				if err != nil {
					conn.WriteJSON(map[string]interface{}{
						"type":  "error",
						"error": err.Error(),
					})
				} else {
					conn.WriteJSON(map[string]interface{}{
						"type":   "tool_result",
						"result": result,
					})
				}

			default:
				conn.WriteJSON(map[string]interface{}{
					"type":  "error",
					"error": "Unknown message type",
				})
			}
		}
	}
}

// HandleEndpointToolExecution handles REST-style tool execution
func HandleEndpointToolExecution(namespaceService NamespaceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespaceVal, exists := c.Get("namespace")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Namespace not found in context"})
			return
		}
		namespace := namespaceVal.(*types.Namespace)

		toolName := c.Param("tool_name")
		if toolName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tool name required"})
			return
		}

		// Parse request body
		var args map[string]interface{}
		if err := c.ShouldBindJSON(&args); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Execute tool through namespace
		result, err := namespaceService.ExecuteTool(c.Request.Context(), namespace.ID, types.ExecuteNamespaceToolRequest{
			Tool:      toolName,
			Arguments: args,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Tool execution failed",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// HandleEndpointHealth handles health check for endpoints
func HandleEndpointHealth() gin.HandlerFunc {
	return func(c *gin.Context) {
		endpointVal, exists := c.Get("endpoint")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Endpoint not found in context"})
			return
		}
		endpoint := endpointVal.(*types.Endpoint)

		c.JSON(http.StatusOK, gin.H{
			"status":   "healthy",
			"endpoint": endpoint.Name,
			"active":   endpoint.IsActive,
		})
	}
}
