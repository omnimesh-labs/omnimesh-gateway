package handlers

import (
	"net/http"

	"mcp-gateway/apps/backend/internal/discovery"
	"mcp-gateway/apps/backend/internal/gateway"
	"mcp-gateway/apps/backend/internal/types"

	"github.com/gin-gonic/gin"
)

// GatewayHandler handles gateway management endpoints
type GatewayHandler struct {
	discoveryService *discovery.Service
	proxy            *gateway.Proxy
	mcpProxy         *gateway.MCPProxy
}

// NewGatewayHandler creates a new gateway handler
func NewGatewayHandler(discoveryService *discovery.Service, proxy *gateway.Proxy, mcpProxy *gateway.MCPProxy) *GatewayHandler {
	return &GatewayHandler{
		discoveryService: discoveryService,
		proxy:            proxy,
		mcpProxy:         mcpProxy,
	}
}

// ListServers lists all MCP servers
func (h *GatewayHandler) ListServers(c *gin.Context) {
	// TODO: Implement server listing logic
	servers, err := h.discoveryService.ListServers("default-org")
	if err != nil {
		c.JSON(types.GetStatusCode(err), types.ErrorResponse{
			Error:   err.(*types.Error),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    servers,
	})
}

// RegisterServer registers a new MCP server
func (h *GatewayHandler) RegisterServer(c *gin.Context) {
	var req types.CreateMCPServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError(err.Error()),
			Success: false,
		})
		return
	}

	// TODO: Implement server registration logic
	server, err := h.discoveryService.RegisterServer("default-org", &req)
	if err != nil {
		c.JSON(types.GetStatusCode(err), types.ErrorResponse{
			Error:   err.(*types.Error),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    server,
	})
}

// GetServer retrieves a specific MCP server
func (h *GatewayHandler) GetServer(c *gin.Context) {
	serverID := c.Param("id")
	if serverID == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Server ID is required"),
			Success: false,
		})
		return
	}

	// TODO: Implement server retrieval logic
	server, err := h.discoveryService.GetServer(serverID)
	if err != nil {
		c.JSON(types.GetStatusCode(err), types.ErrorResponse{
			Error:   err.(*types.Error),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    server,
	})
}

// UpdateServer updates an existing MCP server
func (h *GatewayHandler) UpdateServer(c *gin.Context) {
	serverID := c.Param("id")
	if serverID == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Server ID is required"),
			Success: false,
		})
		return
	}

	var req types.UpdateMCPServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError(err.Error()),
			Success: false,
		})
		return
	}

	// TODO: Implement server update logic
	server, err := h.discoveryService.UpdateServer(serverID, &req)
	if err != nil {
		c.JSON(types.GetStatusCode(err), types.ErrorResponse{
			Error:   err.(*types.Error),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    server,
	})
}

// UnregisterServer removes an MCP server
func (h *GatewayHandler) UnregisterServer(c *gin.Context) {
	serverID := c.Param("id")
	if serverID == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Server ID is required"),
			Success: false,
		})
		return
	}

	// TODO: Implement server unregistration logic
	err := h.discoveryService.UnregisterServer(serverID)
	if err != nil {
		c.JSON(types.GetStatusCode(err), types.ErrorResponse{
			Error:   err.(*types.Error),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Server unregistered successfully",
	})
}

// GetServerStats returns statistics for a server
func (h *GatewayHandler) GetServerStats(c *gin.Context) {
	serverID := c.Param("id")
	if serverID == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Server ID is required"),
			Success: false,
		})
		return
	}

	// TODO: Implement server statistics retrieval
	stats, err := h.discoveryService.GetServerStats(serverID)
	if err != nil {
		c.JSON(types.GetStatusCode(err), types.ErrorResponse{
			Error:   err.(*types.Error),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// ProxyRequest proxies requests to MCP servers
func (h *GatewayHandler) ProxyRequest(c *gin.Context) {
	h.proxy.ProxyRequest(c)
}

// CreateMCPSession creates a new MCP session
func (h *GatewayHandler) CreateMCPSession(c *gin.Context) {
	var req struct {
		ServerID string `json:"server_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError(err.Error()),
			Success: false,
		})
		return
	}

	// Get server details
	server, err := h.discoveryService.GetServer(req.ServerID)
	if err != nil {
		c.JSON(types.GetStatusCode(err), types.ErrorResponse{
			Error:   err.(*types.Error),
			Success: false,
		})
		return
	}

	// Create session with default values for single-tenant
	session, err := h.mcpProxy.CreateSession("default-user", "default-org", server)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError(err.Error()),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    session,
	})
}

// HandleMCPWebSocket handles MCP WebSocket connections
func (h *GatewayHandler) HandleMCPWebSocket(c *gin.Context) {
	h.mcpProxy.HandleWebSocket(c)
}

// ListMCPSessions lists all MCP sessions
func (h *GatewayHandler) ListMCPSessions(c *gin.Context) {
	sessions := h.mcpProxy.ListSessions("default-org")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    sessions,
	})
}

// CloseMCPSession closes an MCP session
func (h *GatewayHandler) CloseMCPSession(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Session ID is required"),
			Success: false,
		})
		return
	}

	err := h.mcpProxy.CloseSession(sessionID)
	if err != nil {
		c.JSON(types.GetStatusCode(err), types.ErrorResponse{
			Error:   err.(*types.Error),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Session closed successfully",
	})
}
