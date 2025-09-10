package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/database/models"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PromptHandler handles MCP prompt endpoints
type PromptHandler struct {
	promptModel *models.MCPPromptModel
}

// NewPromptHandler creates a new prompt handler
func NewPromptHandler(promptModel *models.MCPPromptModel) *PromptHandler {
	return &PromptHandler{
		promptModel: promptModel,
	}
}

// ListPrompts lists all prompts for an organization
func (h *PromptHandler) ListPrompts(c *gin.Context) {
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

	// Parse query parameters - show all prompts by default so UI can toggle active/inactive
	activeOnly := c.Query("active") == "true"
	category := c.Query("category")
	searchTerm := c.Query("search")
	popular := c.Query("popular") == "true"

	var prompts []*models.MCPPrompt

	if popular {
		limit := 10
		if l := c.Query("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
				limit = parsed
			}
		}
		prompts, err = h.promptModel.GetPopularPrompts(orgUUID, activeOnly, limit)
	} else if searchTerm != "" {
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

		prompts, err = h.promptModel.SearchPrompts(orgUUID, searchTerm, activeOnly, limit, offset)
	} else if category != "" {
		prompts, err = h.promptModel.ListByCategory(orgUUID, category, activeOnly)
	} else {
		prompts, err = h.promptModel.ListByOrganization(orgUUID, activeOnly)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Failed to retrieve prompts"),
			Success: false,
		})
		return
	}

	// Ensure data is always an array, never null
	if prompts == nil {
		prompts = []*models.MCPPrompt{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    prompts,
		"count":   len(prompts),
	})
}

// CreatePrompt creates a new MCP prompt
func (h *PromptHandler) CreatePrompt(c *gin.Context) {
	orgID, exists := c.Get("organization_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error:   types.NewUnauthorizedError("Organization ID not found"),
			Success: false,
		})
		return
	}

	userID, userExists := c.Get("user_id")

	var req types.CreatePromptRequest
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

	// Validate category
	validCategories := []string{
		types.PromptCategoryGeneral,
		types.PromptCategoryCoding,
		types.PromptCategoryAnalysis,
		types.PromptCategoryCreative,
		types.PromptCategoryEducational,
		types.PromptCategoryBusiness,
		types.PromptCategoryCustom,
	}
	isValidCategory := false
	for _, validCategory := range validCategories {
		if req.Category == validCategory {
			isValidCategory = true
			break
		}
	}
	if !isValidCategory {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Invalid prompt category"),
			Success: false,
		})
		return
	}

	// Check if prompt name already exists
	_, err = h.promptModel.GetByName(orgUUID, req.Name)
	if err == nil {
		c.JSON(http.StatusConflict, types.ErrorResponse{
			Error:   types.NewValidationError("Prompt with this name already exists"),
			Success: false,
		})
		return
	}

	// Create prompt model
	prompt := &models.MCPPrompt{
		OrganizationID: orgUUID,
		Name:           req.Name,
		PromptTemplate: req.PromptTemplate,
		Category:       req.Category,
		UsageCount:     0,
		IsActive:       true,
		Metadata:       req.Metadata,
		Tags:           req.Tags,
		Parameters:     req.Parameters,
	}

	if req.Description != "" {
		prompt.Description = sql.NullString{String: req.Description, Valid: true}
	}
	if userExists {
		if userUUID, err := uuid.Parse(userID.(string)); err == nil {
			prompt.CreatedBy = uuid.NullUUID{UUID: userUUID, Valid: true}
		}
	}

	// Create prompt
	err = h.promptModel.Create(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Failed to create prompt"),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    prompt,
	})
}

// GetPrompt retrieves a specific MCP prompt
func (h *PromptHandler) GetPrompt(c *gin.Context) {
	promptID := c.Param("id")
	if promptID == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Prompt ID is required"),
			Success: false,
		})
		return
	}

	promptUUID, err := uuid.Parse(promptID)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Invalid prompt ID format"),
			Success: false,
		})
		return
	}

	prompt, err := h.promptModel.GetByID(promptUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, types.ErrorResponse{
				Error:   types.NewNotFoundError("Prompt not found"),
				Success: false,
			})
		} else {
			c.JSON(http.StatusInternalServerError, types.ErrorResponse{
				Error:   types.NewInternalError("Failed to retrieve prompt"),
				Success: false,
			})
		}
		return
	}

	// Increment usage count for analytics
	_ = h.promptModel.IncrementUsageCount(promptUUID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    prompt,
	})
}

// UpdatePrompt updates an existing MCP prompt
func (h *PromptHandler) UpdatePrompt(c *gin.Context) {
	promptID := c.Param("id")
	if promptID == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Prompt ID is required"),
			Success: false,
		})
		return
	}

	promptUUID, err := uuid.Parse(promptID)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Invalid prompt ID format"),
			Success: false,
		})
		return
	}

	var req types.UpdatePromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError(err.Error()),
			Success: false,
		})
		return
	}

	// Get existing prompt
	prompt, err := h.promptModel.GetByID(promptUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, types.ErrorResponse{
				Error:   types.NewNotFoundError("Prompt not found"),
				Success: false,
			})
		} else {
			c.JSON(http.StatusInternalServerError, types.ErrorResponse{
				Error:   types.NewInternalError("Failed to retrieve prompt"),
				Success: false,
			})
		}
		return
	}

	// Update fields if provided
	if req.Name != "" {
		prompt.Name = req.Name
	}
	if req.Description != "" {
		prompt.Description = sql.NullString{String: req.Description, Valid: true}
	}
	if req.PromptTemplate != "" {
		prompt.PromptTemplate = req.PromptTemplate
	}
	if req.Category != "" {
		// Validate category
		validCategories := []string{
			types.PromptCategoryGeneral,
			types.PromptCategoryCoding,
			types.PromptCategoryAnalysis,
			types.PromptCategoryCreative,
			types.PromptCategoryEducational,
			types.PromptCategoryBusiness,
			types.PromptCategoryCustom,
		}
		isValidCategory := false
		for _, validCategory := range validCategories {
			if req.Category == validCategory {
				isValidCategory = true
				break
			}
		}
		if !isValidCategory {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Error:   types.NewValidationError("Invalid prompt category"),
				Success: false,
			})
			return
		}
		prompt.Category = req.Category
	}
	if req.Parameters != nil {
		prompt.Parameters = req.Parameters
	}
	if req.IsActive != nil {
		prompt.IsActive = *req.IsActive
	}
	if req.Metadata != nil {
		prompt.Metadata = req.Metadata
	}
	if req.Tags != nil {
		prompt.Tags = req.Tags
	}

	err = h.promptModel.Update(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Failed to update prompt"),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    prompt,
	})
}

// DeletePrompt soft deletes an MCP prompt
func (h *PromptHandler) DeletePrompt(c *gin.Context) {
	promptID := c.Param("id")
	if promptID == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Prompt ID is required"),
			Success: false,
		})
		return
	}

	promptUUID, err := uuid.Parse(promptID)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Invalid prompt ID format"),
			Success: false,
		})
		return
	}

	// Check if prompt exists
	_, err = h.promptModel.GetByID(promptUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, types.ErrorResponse{
				Error:   types.NewNotFoundError("Prompt not found"),
				Success: false,
			})
		} else {
			c.JSON(http.StatusInternalServerError, types.ErrorResponse{
				Error:   types.NewInternalError("Failed to retrieve prompt"),
				Success: false,
			})
		}
		return
	}

	err = h.promptModel.Delete(promptUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Failed to delete prompt"),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Prompt deleted successfully",
	})
}

// UsePrompt increments usage count and returns the prompt (for analytics)
func (h *PromptHandler) UsePrompt(c *gin.Context) {
	promptID := c.Param("id")
	if promptID == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Prompt ID is required"),
			Success: false,
		})
		return
	}

	promptUUID, err := uuid.Parse(promptID)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Invalid prompt ID format"),
			Success: false,
		})
		return
	}

	prompt, err := h.promptModel.GetByID(promptUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, types.ErrorResponse{
				Error:   types.NewNotFoundError("Prompt not found"),
				Success: false,
			})
		} else {
			c.JSON(http.StatusInternalServerError, types.ErrorResponse{
				Error:   types.NewInternalError("Failed to retrieve prompt"),
				Success: false,
			})
		}
		return
	}

	// Increment usage count
	err = h.promptModel.IncrementUsageCount(promptUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Failed to update usage count"),
			Success: false,
		})
		return
	}

	// Return updated prompt
	prompt.UsageCount++
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    prompt,
		"message": "Prompt usage recorded",
	})
}
