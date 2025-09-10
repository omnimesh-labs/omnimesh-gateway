package handlers

import (
	"net/http"
	"strings"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/discovery"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"

	"github.com/gin-gonic/gin"
)

// GatewayHandler handles gateway management endpoints
type GatewayHandler struct {
	discoveryService *discovery.Service
}

// NewGatewayHandler creates a new gateway handler
func NewGatewayHandler(discoveryService *discovery.Service) *GatewayHandler {
	return &GatewayHandler{
		discoveryService: discoveryService,
	}
}

// convertToTypesError converts a standard Go error to a types.Error
func convertToTypesError(err error) *types.Error {
	if err == nil {
		return nil
	}

	// If it's already a types.Error, return it
	if typesErr, ok := err.(*types.Error); ok {
		return typesErr
	}

	// Convert based on error message patterns
	errMsg := err.Error()
	errMsgLower := strings.ToLower(errMsg)

	switch {
	case strings.Contains(errMsgLower, "not found"):
		return types.NewNotFoundError(errMsg)
	case strings.Contains(errMsgLower, "already exists"):
		return types.NewAlreadyExistsError(errMsg)
	case strings.Contains(errMsgLower, "invalid"):
		return types.NewValidationError(errMsg)
	case strings.Contains(errMsgLower, "organization"):
		return types.NewValidationError(errMsg)
	case strings.Contains(errMsgLower, "server not found") || strings.Contains(errMsgLower, "server does not exist"):
		return types.NewServerNotFoundError(errMsg)
	case strings.Contains(errMsgLower, "stdio") || strings.Contains(errMsgLower, "communication") || strings.Contains(errMsgLower, "failed to receive") || strings.Contains(errMsgLower, "protocol"):
		return types.NewBadGatewayError(errMsg)
	default:
		return types.NewInternalError(errMsg)
	}
}

// ListServers lists all MCP servers
func (h *GatewayHandler) ListServers(c *gin.Context) {
	servers, err := h.discoveryService.ListServers("default-org")
	if err != nil {
		typesErr := convertToTypesError(err)
		c.JSON(types.GetStatusCode(typesErr), types.ErrorResponse{
			Error:   typesErr,
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

	server, err := h.discoveryService.RegisterServer("default-org", &req)
	if err != nil {
		typesErr := convertToTypesError(err)
		c.JSON(types.GetStatusCode(typesErr), types.ErrorResponse{
			Error:   typesErr,
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

	server, err := h.discoveryService.GetServer(serverID)
	if err != nil {
		typesErr := convertToTypesError(err)
		c.JSON(types.GetStatusCode(typesErr), types.ErrorResponse{
			Error:   typesErr,
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

	server, err := h.discoveryService.UpdateServer(serverID, &req)
	if err != nil {
		typesErr := convertToTypesError(err)
		c.JSON(types.GetStatusCode(typesErr), types.ErrorResponse{
			Error:   typesErr,
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

	err := h.discoveryService.UnregisterServer(serverID)
	if err != nil {
		typesErr := convertToTypesError(err)
		c.JSON(types.GetStatusCode(typesErr), types.ErrorResponse{
			Error:   typesErr,
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

	stats, err := h.discoveryService.GetServerStats(serverID)
	if err != nil {
		typesErr := convertToTypesError(err)
		c.JSON(types.GetStatusCode(typesErr), types.ErrorResponse{
			Error:   typesErr,
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// DiscoverServerTools manually triggers tool discovery for a specific server
func (h *GatewayHandler) DiscoverServerTools(c *gin.Context) {
	serverID := c.Param("id")
	if serverID == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Server ID is required"),
			Success: false,
		})
		return
	}

	// Trigger tool discovery for the server
	err := h.discoveryService.DiscoverServerTools(serverID)
	if err != nil {
		typesErr := convertToTypesError(err)
		c.JSON(types.GetStatusCode(typesErr), types.ErrorResponse{
			Error:   typesErr,
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tool discovery initiated successfully",
	})
}

// ProxyRequest is deprecated - proxy functionality removed
func (h *GatewayHandler) ProxyRequest(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, types.ErrorResponse{
		Error:   types.NewNotImplementedError("Proxy functionality has been removed"),
		Success: false,
	})
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

	// MCP session functionality removed - return not implemented
	c.JSON(http.StatusNotImplemented, types.ErrorResponse{
		Error:   types.NewInternalError("MCP session functionality has been removed"),
		Success: false,
	})
}

// HandleMCPWebSocket handles MCP WebSocket connections
func (h *GatewayHandler) HandleMCPWebSocket(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, types.ErrorResponse{
		Error:   types.NewInternalError("MCP WebSocket functionality has been removed"),
		Success: false,
	})
}

// ListMCPSessions lists all MCP sessions
func (h *GatewayHandler) ListMCPSessions(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, types.ErrorResponse{
		Error:   types.NewInternalError("MCP session functionality has been removed"),
		Success: false,
	})
}

// CloseMCPSession closes an MCP session
func (h *GatewayHandler) CloseMCPSession(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, types.ErrorResponse{
		Error:   types.NewInternalError("MCP session functionality has been removed"),
		Success: false,
	})
}
