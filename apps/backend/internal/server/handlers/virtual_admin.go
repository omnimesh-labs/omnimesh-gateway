package handlers

import (
	"net/http"
	"time"

	"mcp-gateway/apps/backend/internal/types"
	"mcp-gateway/apps/backend/internal/virtual"

	"github.com/gin-gonic/gin"
)

// VirtualAdminHandler handles admin endpoints for virtual servers
type VirtualAdminHandler struct {
	virtualService *virtual.Service
}

// NewVirtualAdminHandler creates a new virtual admin handler
func NewVirtualAdminHandler(virtualService *virtual.Service) *VirtualAdminHandler {
	return &VirtualAdminHandler{
		virtualService: virtualService,
	}
}

// CreateVirtualServer creates a new virtual server
func (h *VirtualAdminHandler) CreateVirtualServer(c *gin.Context) {
	var req types.VirtualServerSpec
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if req.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "ID is required",
		})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Name is required",
		})
		return
	}

	if req.AdapterType == "" {
		req.AdapterType = "REST" // Default to REST
	}

	// Set timestamps
	now := time.Now()
	req.CreatedAt = now
	req.UpdatedAt = now

	// Add the virtual server
	if err := h.virtualService.Add(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create virtual server: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    req,
		"message": "Virtual server created successfully",
	})
}

// GetVirtualServer retrieves a specific virtual server
func (h *VirtualAdminHandler) GetVirtualServer(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Server ID is required",
		})
		return
	}

	spec, err := h.virtualService.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Virtual server not found: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    spec,
	})
}

// ListVirtualServers lists all virtual servers
func (h *VirtualAdminHandler) ListVirtualServers(c *gin.Context) {
	specs, err := h.virtualService.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to list virtual servers: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    specs,
		"count":   len(specs),
	})
}

// UpdateVirtualServer updates an existing virtual server
func (h *VirtualAdminHandler) UpdateVirtualServer(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Server ID is required",
		})
		return
	}

	var req types.VirtualServerSpec
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Name is required",
		})
		return
	}

	// Ensure ID matches
	req.ID = id

	// Update the virtual server
	if err := h.virtualService.Update(id, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update virtual server: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    req,
		"message": "Virtual server updated successfully",
	})
}

// DeleteVirtualServer deletes a virtual server
func (h *VirtualAdminHandler) DeleteVirtualServer(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Server ID is required",
		})
		return
	}

	if err := h.virtualService.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete virtual server: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Virtual server deleted successfully",
	})
}

// TestVirtualServerTool tests a specific tool from a virtual server
func (h *VirtualAdminHandler) TestVirtualServerTool(c *gin.Context) {
	id := c.Param("id")
	toolName := c.Param("tool")

	if id == "" || toolName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Server ID and tool name are required",
		})
		return
	}

	// Parse test arguments
	var args map[string]interface{}
	if err := c.ShouldBindJSON(&args); err != nil {
		// If no body provided, use empty args
		args = make(map[string]interface{})
	}

	// Get virtual server spec
	spec, err := h.virtualService.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Virtual server not found: " + err.Error(),
		})
		return
	}

	// Create virtual server instance
	vs := virtual.NewVirtualServer(spec)

	// Call the tool
	result, err := vs.CallTool(toolName, args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Tool call failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"server_id": id,
			"tool_name": toolName,
			"arguments": args,
			"result":    result,
		},
		"message": "Tool test completed successfully",
	})
}

// GetVirtualServerTools lists all tools for a specific virtual server
func (h *VirtualAdminHandler) GetVirtualServerTools(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Server ID is required",
		})
		return
	}

	// Get virtual server spec
	spec, err := h.virtualService.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Virtual server not found: " + err.Error(),
		})
		return
	}

	// Create virtual server instance
	vs := virtual.NewVirtualServer(spec)

	// Get tools list
	result, err := vs.ListTools()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to list tools: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"server_id": id,
			"tools":     result.Tools,
			"count":     len(result.Tools),
		},
	})
}
