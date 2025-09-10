package handlers

import (
	"context"
	"net/http"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"

	"github.com/gin-gonic/gin"
)

// NamespaceService defines the interface for namespace operations
type NamespaceService interface {
	CreateNamespace(ctx context.Context, req types.CreateNamespaceRequest) (*types.Namespace, error)
	GetNamespace(ctx context.Context, id string) (*types.Namespace, error)
	ListNamespaces(ctx context.Context, orgID string) ([]*types.Namespace, error)
	UpdateNamespace(ctx context.Context, id string, req types.UpdateNamespaceRequest) (*types.Namespace, error)
	DeleteNamespace(ctx context.Context, id string) error
	AddServerToNamespace(ctx context.Context, namespaceID string, req types.AddServerToNamespaceRequest) error
	RemoveServerFromNamespace(ctx context.Context, namespaceID, serverID string) error
	UpdateServerStatus(ctx context.Context, namespaceID, serverID string, req types.UpdateServerStatusRequest) error
	AggregateTools(ctx context.Context, namespaceID string) ([]types.NamespaceTool, error)
	UpdateToolStatus(ctx context.Context, namespaceID, serverID, toolName string, req types.UpdateToolStatusRequest) error
	ExecuteTool(ctx context.Context, namespaceID string, req types.ExecuteNamespaceToolRequest) (*types.NamespaceToolResult, error)
}

// NamespaceHandler handles namespace-related HTTP requests
type NamespaceHandler struct {
	service NamespaceService
}

// NewNamespaceHandler creates a new namespace handler
func NewNamespaceHandler(service NamespaceService) *NamespaceHandler {
	return &NamespaceHandler{
		service: service,
	}
}

// CreateNamespace handles POST /api/namespaces
func (h *NamespaceHandler) CreateNamespace(c *gin.Context) {
	var req types.CreateNamespaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithValidationError(c, "Invalid request format")
		return
	}

	// Get organization ID from context (set by auth middleware)
	orgID, exists := c.Get("organization_id")
	if !exists {
		// Fallback to default organization for now
		orgID = "00000000-0000-0000-0000-000000000001"
	}
	req.OrganizationID = orgID.(string)

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if exists && userID != nil {
		userIDStr := userID.(string)
		req.CreatedBy = &userIDStr
	}

	namespace, err := h.service.CreateNamespace(c.Request.Context(), req)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, namespace)
}

// ListNamespaces handles GET /api/namespaces
func (h *NamespaceHandler) ListNamespaces(c *gin.Context) {
	// Get organization ID from context
	orgID, exists := c.Get("organization_id")
	if !exists {
		// Fallback to default organization for now
		orgID = "00000000-0000-0000-0000-000000000001"
	}

	namespaces, err := h.service.ListNamespaces(c.Request.Context(), orgID.(string))
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"namespaces": namespaces,
		"total":      len(namespaces),
	})
}

// GetNamespace handles GET /api/namespaces/:id
func (h *NamespaceHandler) GetNamespace(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		RespondWithValidationError(c, "namespace ID is required")
		return
	}

	namespace, err := h.service.GetNamespace(c.Request.Context(), id)
	if err != nil {
		RespondWithNotFound(c, "Namespace")
		return
	}

	c.JSON(http.StatusOK, namespace)
}

// UpdateNamespace handles PUT /api/namespaces/:id
func (h *NamespaceHandler) UpdateNamespace(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		RespondWithValidationError(c, "namespace ID is required")
		return
	}

	var req types.UpdateNamespaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithValidationError(c, "Invalid request format")
		return
	}

	namespace, err := h.service.UpdateNamespace(c.Request.Context(), id, req)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, namespace)
}

// DeleteNamespace handles DELETE /api/namespaces/:id
func (h *NamespaceHandler) DeleteNamespace(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		RespondWithValidationError(c, "namespace ID is required")
		return
	}

	if err := h.service.DeleteNamespace(c.Request.Context(), id); err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// AddServerToNamespace handles POST /api/namespaces/:id/servers
func (h *NamespaceHandler) AddServerToNamespace(c *gin.Context) {
	namespaceID := c.Param("id")
	if namespaceID == "" {
		RespondWithValidationError(c, "namespace ID is required")
		return
	}

	var req types.AddServerToNamespaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithValidationError(c, "Invalid request format")
		return
	}

	if err := h.service.AddServerToNamespace(c.Request.Context(), namespaceID, req); err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "server added to namespace"})
}

// RemoveServerFromNamespace handles DELETE /api/namespaces/:id/servers/:server_id
func (h *NamespaceHandler) RemoveServerFromNamespace(c *gin.Context) {
	namespaceID := c.Param("id")
	serverID := c.Param("server_id")

	if namespaceID == "" || serverID == "" {
		RespondWithValidationError(c, "namespace ID and server ID are required")
		return
	}

	if err := h.service.RemoveServerFromNamespace(c.Request.Context(), namespaceID, serverID); err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "server removed from namespace"})
}

// UpdateServerStatus handles PUT /api/namespaces/:id/servers/:server_id/status
func (h *NamespaceHandler) UpdateServerStatus(c *gin.Context) {
	namespaceID := c.Param("id")
	serverID := c.Param("server_id")

	if namespaceID == "" || serverID == "" {
		RespondWithValidationError(c, "namespace ID and server ID are required")
		return
	}

	var req types.UpdateServerStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithValidationError(c, "Invalid request format")
		return
	}

	if err := h.service.UpdateServerStatus(c.Request.Context(), namespaceID, serverID, req); err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "server status updated"})
}

// GetNamespaceTools handles GET /api/namespaces/:id/tools
func (h *NamespaceHandler) GetNamespaceTools(c *gin.Context) {
	namespaceID := c.Param("id")
	if namespaceID == "" {
		RespondWithValidationError(c, "namespace ID is required")
		return
	}

	tools, err := h.service.AggregateTools(c.Request.Context(), namespaceID)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tools": tools,
		"total": len(tools),
	})
}

// UpdateToolStatus handles PUT /api/namespaces/:id/tools/:tool_id/status
func (h *NamespaceHandler) UpdateToolStatus(c *gin.Context) {
	namespaceID := c.Param("id")
	serverID := c.Query("server_id")
	toolName := c.Param("tool_id")

	if namespaceID == "" || serverID == "" || toolName == "" {
		RespondWithValidationError(c, "namespace ID, server ID, and tool name are required")
		return
	}

	var req types.UpdateToolStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithValidationError(c, "Invalid request format")
		return
	}

	if err := h.service.UpdateToolStatus(c.Request.Context(), namespaceID, serverID, toolName, req); err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tool status updated"})
}

// ExecuteNamespaceTool handles POST /api/namespaces/:id/execute
func (h *NamespaceHandler) ExecuteNamespaceTool(c *gin.Context) {
	namespaceID := c.Param("id")
	if namespaceID == "" {
		RespondWithValidationError(c, "namespace ID is required")
		return
	}

	var req types.ExecuteNamespaceToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithValidationError(c, "Invalid request format")
		return
	}

	result, err := h.service.ExecuteTool(c.Request.Context(), namespaceID, req)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
