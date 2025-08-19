package handlers

import (
	"net/http"
	"strconv"

	"mcp-gateway/internal/discovery"
	"mcp-gateway/internal/types"

	"github.com/gin-gonic/gin"
)

// MCPDiscoveryHandler handles MCP package discovery endpoints
type MCPDiscoveryHandler struct {
	discoveryService *discovery.MCPDiscoveryService
}

// NewMCPDiscoveryHandler creates a new MCP discovery handler
func NewMCPDiscoveryHandler(discoveryService *discovery.MCPDiscoveryService) *MCPDiscoveryHandler {
	return &MCPDiscoveryHandler{
		discoveryService: discoveryService,
	}
}

// SearchPackages handles GET /api/mcp/search?query=<term>
// If no query is provided, it returns all available packages
func (h *MCPDiscoveryHandler) SearchPackages(c *gin.Context) {
	query := c.Query("query") // Optional - can be empty

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	pageSize := 20 // Default page size
	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	req := &types.MCPDiscoveryRequest{
		Query:    query,
		Offset:   offset,
		PageSize: pageSize,
	}

	result, err := h.discoveryService.SearchPackages(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Failed to search MCP packages: " + err.Error()),
			Success: false,
		})
		return
	}

	response := types.MCPDiscoveryListResponse{
		Success: true,
		Data:    *result,
		Message: "MCP packages retrieved successfully",
	}

	c.JSON(http.StatusOK, response)
}

// ListPackages handles GET /api/mcp/packages - lists all available packages
func (h *MCPDiscoveryHandler) ListPackages(c *gin.Context) {
	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	pageSize := 20 // Default page size
	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	result, err := h.discoveryService.ListAllPackages(offset, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Failed to list MCP packages: " + err.Error()),
			Success: false,
		})
		return
	}

	response := types.MCPDiscoveryListResponse{
		Success: true,
		Data:    *result,
		Message: "MCP packages listed successfully",
	}

	c.JSON(http.StatusOK, response)
}

// GetPackageDetails handles GET /api/mcp/packages/:packageName
func (h *MCPDiscoveryHandler) GetPackageDetails(c *gin.Context) {
	packageName := c.Param("packageName")
	if packageName == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("package name is required"),
			Success: false,
		})
		return
	}

	// Search for the specific package
	req := &types.MCPDiscoveryRequest{
		Query:    packageName,
		Offset:   0,
		PageSize: 50, // Get more results to find exact match
	}

	result, err := h.discoveryService.SearchPackages(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Failed to get package details: " + err.Error()),
			Success: false,
		})
		return
	}

	// Look for exact package name match
	for key, pkg := range result.Results {
		if key == packageName || pkg.PackageName == packageName {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    pkg,
				"message": "Package details retrieved successfully",
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, types.ErrorResponse{
		Error:   types.NewNotFoundError("Package not found: " + packageName),
		Success: false,
	})
}
