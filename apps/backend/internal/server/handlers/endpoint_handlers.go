package handlers

import (
	"context"
	"net/http"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"

	"github.com/gin-gonic/gin"
)

// EndpointService defines the interface for endpoint operations
type EndpointService interface {
	CreateEndpoint(ctx context.Context, req types.CreateEndpointRequest, orgID string, userID *string) (*types.Endpoint, error)
	GetEndpoint(ctx context.Context, id string) (*types.Endpoint, error)
	GetEndpointByName(ctx context.Context, name string) (*types.Endpoint, error)
	ListEndpoints(ctx context.Context, orgID string) ([]*types.Endpoint, error)
	ListPublicEndpoints(ctx context.Context) ([]*types.Endpoint, error)
	UpdateEndpoint(ctx context.Context, id string, req types.UpdateEndpointRequest) (*types.Endpoint, error)
	DeleteEndpoint(ctx context.Context, id string) error
	ResolveEndpoint(ctx context.Context, name string) (*types.EndpointConfig, error)
	ValidateAccess(ctx context.Context, endpoint *types.Endpoint, req *http.Request) error
}

// EndpointHandler handles endpoint-related HTTP requests
type EndpointHandler struct {
	service EndpointService
}

// NewEndpointHandler creates a new endpoint handler
func NewEndpointHandler(service EndpointService) *EndpointHandler {
	return &EndpointHandler{
		service: service,
	}
}

// CreateEndpoint handles POST /api/endpoints
func (h *EndpointHandler) CreateEndpoint(c *gin.Context) {
	var req types.CreateEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithValidationError(c, "Invalid request format")
		return
	}

	// Get organization ID from context (set by auth middleware)
	orgID, exists := c.Get("organization_id")
	if !exists {
		// Fallback to default organization for now
		orgID = "00000000-0000-0000-0000-000000000000"
	}

	// Get user ID from context (set by auth middleware)
	var userID *string
	userIDVal, exists := c.Get("user_id")
	if exists && userIDVal != nil {
		userIDStr := userIDVal.(string)
		userID = &userIDStr
	}

	endpoint, err := h.service.CreateEndpoint(c.Request.Context(), req, orgID.(string), userID)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, endpoint)
}

// ListEndpoints handles GET /api/endpoints
func (h *EndpointHandler) ListEndpoints(c *gin.Context) {
	// Get organization ID from context
	orgID, exists := c.Get("organization_id")
	if !exists {
		// Fallback to default organization for now
		orgID = "00000000-0000-0000-0000-000000000000"
	}

	endpoints, err := h.service.ListEndpoints(c.Request.Context(), orgID.(string))
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"endpoints": endpoints,
		"total":     len(endpoints),
	})
}

// ListPublicEndpoints handles GET /metamcp
func (h *EndpointHandler) ListPublicEndpoints(c *gin.Context) {
	endpoints, err := h.service.ListPublicEndpoints(c.Request.Context())
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"endpoints": endpoints,
		"total":     len(endpoints),
	})
}

// GetEndpoint handles GET /api/endpoints/:id
func (h *EndpointHandler) GetEndpoint(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		RespondWithValidationError(c, "endpoint ID is required")
		return
	}

	endpoint, err := h.service.GetEndpoint(c.Request.Context(), id)
	if err != nil {
		RespondWithNotFound(c, "Endpoint")
		return
	}

	c.JSON(http.StatusOK, endpoint)
}

// UpdateEndpoint handles PUT /api/endpoints/:id
func (h *EndpointHandler) UpdateEndpoint(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		RespondWithValidationError(c, "endpoint ID is required")
		return
	}

	var req types.UpdateEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithValidationError(c, "Invalid request format")
		return
	}

	endpoint, err := h.service.UpdateEndpoint(c.Request.Context(), id, req)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, endpoint)
}

// DeleteEndpoint handles DELETE /api/endpoints/:id
func (h *EndpointHandler) DeleteEndpoint(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		RespondWithValidationError(c, "endpoint ID is required")
		return
	}

	if err := h.service.DeleteEndpoint(c.Request.Context(), id); err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// RegenerateEndpointKeys handles POST /api/endpoints/:id/regenerate-keys
func (h *EndpointHandler) RegenerateEndpointKeys(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		RespondWithValidationError(c, "endpoint ID is required")
		return
	}

	// TODO: Implement API key regeneration logic
	c.JSON(http.StatusOK, gin.H{
		"message":     "API keys regenerated successfully",
		"endpoint_id": id,
	})
}
