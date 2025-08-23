package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"mcp-gateway/apps/backend/internal/database/models"
	"mcp-gateway/apps/backend/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ResourceHandler handles MCP resource endpoints
type ResourceHandler struct {
	resourceModel *models.MCPResourceModel
}

// NewResourceHandler creates a new resource handler
func NewResourceHandler(resourceModel *models.MCPResourceModel) *ResourceHandler {
	return &ResourceHandler{
		resourceModel: resourceModel,
	}
}

// ListResources lists all resources for an organization
func (h *ResourceHandler) ListResources(c *gin.Context) {
	orgID, exists := c.Get("organization_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error:   types.NewUnauthorizedError("Organization ID not found"),
			Success: false,
		})
		return
	}

	orgUUID, err := uuid.Parse(orgID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Invalid organization ID format"),
			Success: false,
		})
		return
	}

	// Parse query parameters
	activeOnly := c.Query("active") != "false"
	resourceType := c.Query("type")
	searchTerm := c.Query("search")

	var resources []*models.MCPResource

	if searchTerm != "" {
		limit := 50
		offset := 0
		if l := c.Query("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
				limit = parsed
			}
		}
		if o := c.Query("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
				offset = parsed
			}
		}
		
		resources, err = h.resourceModel.SearchResources(orgUUID, searchTerm, limit, offset)
	} else if resourceType != "" {
		resources, err = h.resourceModel.ListByType(orgUUID, resourceType, activeOnly)
	} else {
		resources, err = h.resourceModel.ListByOrganization(orgUUID, activeOnly)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Failed to retrieve resources"),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resources,
		"count":   len(resources),
	})
}

// CreateResource creates a new MCP resource
func (h *ResourceHandler) CreateResource(c *gin.Context) {
	orgID, exists := c.Get("organization_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error:   types.NewUnauthorizedError("Organization ID not found"),
			Success: false,
		})
		return
	}

	userID, userExists := c.Get("user_id")
	
	var req types.CreateResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError(err.Error()),
			Success: false,
		})
		return
	}

	orgUUID, err := uuid.Parse(orgID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Invalid organization ID format"),
			Success: false,
		})
		return
	}

	// Validate resource type
	validResourceTypes := []string{
		types.ResourceTypeFile,
		types.ResourceTypeURL,
		types.ResourceTypeDatabase,
		types.ResourceTypeAPI,
		types.ResourceTypeMemory,
		types.ResourceTypeCustom,
	}
	isValidType := false
	for _, validType := range validResourceTypes {
		if req.ResourceType == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Invalid resource type"),
			Success: false,
		})
		return
	}

	// Check if resource name already exists
	_, err = h.resourceModel.GetByName(orgUUID, req.Name)
	if err == nil {
		c.JSON(http.StatusConflict, types.ErrorResponse{
			Error:   types.NewValidationError("Resource with this name already exists"),
			Success: false,
		})
		return
	}

	// Create resource model
	resource := &models.MCPResource{
		OrganizationID:    orgUUID,
		Name:              req.Name,
		ResourceType:      req.ResourceType,
		URI:               req.URI,
		IsActive:          true,
		Metadata:          req.Metadata,
		Tags:              req.Tags,
		AccessPermissions: req.AccessPermissions,
	}

	if req.Description != "" {
		resource.Description = sql.NullString{String: req.Description, Valid: true}
	}
	if req.MimeType != "" {
		resource.MimeType = sql.NullString{String: req.MimeType, Valid: true}
	}
	if req.SizeBytes != nil {
		resource.SizeBytes = sql.NullInt64{Int64: *req.SizeBytes, Valid: true}
	}
	if userExists {
		if userUUID, err := uuid.Parse(userID.(string)); err == nil {
			resource.CreatedBy = uuid.NullUUID{UUID: userUUID, Valid: true}
		}
	}

	// Create resource
	err = h.resourceModel.Create(resource)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Failed to create resource"),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    resource,
	})
}

// GetResource retrieves a specific MCP resource
func (h *ResourceHandler) GetResource(c *gin.Context) {
	resourceID := c.Param("id")
	if resourceID == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Resource ID is required"),
			Success: false,
		})
		return
	}

	resourceUUID, err := uuid.Parse(resourceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Invalid resource ID format"),
			Success: false,
		})
		return
	}

	resource, err := h.resourceModel.GetByID(resourceUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, types.ErrorResponse{
				Error:   types.NewNotFoundError("Resource not found"),
				Success: false,
			})
		} else {
			c.JSON(http.StatusInternalServerError, types.ErrorResponse{
				Error:   types.NewInternalError("Failed to retrieve resource"),
				Success: false,
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resource,
	})
}

// UpdateResource updates an existing MCP resource
func (h *ResourceHandler) UpdateResource(c *gin.Context) {
	resourceID := c.Param("id")
	if resourceID == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Resource ID is required"),
			Success: false,
		})
		return
	}

	resourceUUID, err := uuid.Parse(resourceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Invalid resource ID format"),
			Success: false,
		})
		return
	}

	var req types.UpdateResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError(err.Error()),
			Success: false,
		})
		return
	}

	// Get existing resource
	resource, err := h.resourceModel.GetByID(resourceUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, types.ErrorResponse{
				Error:   types.NewNotFoundError("Resource not found"),
				Success: false,
			})
		} else {
			c.JSON(http.StatusInternalServerError, types.ErrorResponse{
				Error:   types.NewInternalError("Failed to retrieve resource"),
				Success: false,
			})
		}
		return
	}

	// Update fields if provided
	if req.Name != "" {
		resource.Name = req.Name
	}
	if req.Description != "" {
		resource.Description = sql.NullString{String: req.Description, Valid: true}
	}
	if req.ResourceType != "" {
		// Validate resource type
		validResourceTypes := []string{
			types.ResourceTypeFile,
			types.ResourceTypeURL,
			types.ResourceTypeDatabase,
			types.ResourceTypeAPI,
			types.ResourceTypeMemory,
			types.ResourceTypeCustom,
		}
		isValidType := false
		for _, validType := range validResourceTypes {
			if req.ResourceType == validType {
				isValidType = true
				break
			}
		}
		if !isValidType {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Error:   types.NewValidationError("Invalid resource type"),
				Success: false,
			})
			return
		}
		resource.ResourceType = req.ResourceType
	}
	if req.URI != "" {
		resource.URI = req.URI
	}
	if req.MimeType != "" {
		resource.MimeType = sql.NullString{String: req.MimeType, Valid: true}
	}
	if req.SizeBytes != nil {
		resource.SizeBytes = sql.NullInt64{Int64: *req.SizeBytes, Valid: true}
	}
	if req.AccessPermissions != nil {
		resource.AccessPermissions = req.AccessPermissions
	}
	if req.IsActive != nil {
		resource.IsActive = *req.IsActive
	}
	if req.Metadata != nil {
		resource.Metadata = req.Metadata
	}
	if req.Tags != nil {
		resource.Tags = req.Tags
	}

	err = h.resourceModel.Update(resource)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Failed to update resource"),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resource,
	})
}

// DeleteResource soft deletes an MCP resource
func (h *ResourceHandler) DeleteResource(c *gin.Context) {
	resourceID := c.Param("id")
	if resourceID == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Resource ID is required"),
			Success: false,
		})
		return
	}

	resourceUUID, err := uuid.Parse(resourceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Invalid resource ID format"),
			Success: false,
		})
		return
	}

	// Check if resource exists
	_, err = h.resourceModel.GetByID(resourceUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, types.ErrorResponse{
				Error:   types.NewNotFoundError("Resource not found"),
				Success: false,
			})
		} else {
			c.JSON(http.StatusInternalServerError, types.ErrorResponse{
				Error:   types.NewInternalError("Failed to retrieve resource"),
				Success: false,
			})
		}
		return
	}

	err = h.resourceModel.Delete(resourceUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Failed to delete resource"),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Resource deleted successfully",
	})
}