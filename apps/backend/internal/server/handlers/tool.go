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

// ToolWithServerInfo extends MCPTool with server information
type ToolWithServerInfo struct {
	*models.MCPTool
	ServerName     *string `json:"server_name,omitempty"`
	ServerProtocol *string `json:"server_protocol,omitempty"`
	ServerStatus   *string `json:"server_status,omitempty"`
}

// ToolHandler handles MCP tool endpoints
type ToolHandler struct {
	toolModel   *models.MCPToolModel
	serverModel *models.MCPServerModel
}

// NewToolHandler creates a new tool handler
func NewToolHandler(toolModel *models.MCPToolModel, serverModel *models.MCPServerModel) *ToolHandler {
	return &ToolHandler{
		toolModel:   toolModel,
		serverModel: serverModel,
	}
}

// ListTools lists all tools for an organization
func (h *ToolHandler) ListTools(c *gin.Context) {
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

	activeOnly := c.Query("active") == "true"
	category := c.Query("category")
	searchTerm := c.Query("search")
	popular := c.Query("popular") == "true"
	includePublic := c.Query("include_public") == "true"

	var tools []*models.MCPTool

	if popular {
		limit := 10
		if l := c.Query("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
				limit = parsed
			}
		}
		tools, err = h.toolModel.GetPopularTools(orgUUID, limit)
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

		tools, err = h.toolModel.SearchTools(orgUUID, searchTerm, limit, offset)
	} else if category != "" {
		tools, err = h.toolModel.ListByCategory(orgUUID, category, activeOnly)
	} else {
		tools, err = h.toolModel.ListByOrganization(orgUUID, activeOnly)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Failed to retrieve tools"),
			Success: false,
		})
		return
	}

	// Include public tools if requested
	if includePublic {
		publicTools, err := h.toolModel.ListPublicTools(50, 0)
		if err == nil {
			tools = append(tools, publicTools...)
		}
	}

	// Enrich tools with server information
	enrichedTools, err := h.enrichToolsWithServerInfo(tools)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Failed to enrich tools with server information"),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    enrichedTools,
		"count":   len(enrichedTools),
	})
}

// CreateTool creates a new MCP tool
func (h *ToolHandler) CreateTool(c *gin.Context) {
	orgID, exists := c.Get("organization_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error:   types.NewUnauthorizedError("Organization ID not found"),
			Success: false,
		})
		return
	}

	userID, userExists := c.Get("user_id")

	var req types.CreateGlobalToolRequest
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
		types.ToolCategoryGeneral,
		types.ToolCategoryData,
		types.ToolCategoryFile,
		types.ToolCategoryWeb,
		types.ToolCategorySystem,
		types.ToolCategoryAI,
		types.ToolCategoryDev,
		types.ToolCategoryCustom,
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
			Error:   types.NewValidationError("Invalid tool category"),
			Success: false,
		})
		return
	}

	// Validate implementation type
	validImplementationTypes := []string{
		types.ToolImplementationInternal,
		types.ToolImplementationExternal,
		types.ToolImplementationWebhook,
		types.ToolImplementationScript,
	}
	if req.ImplementationType == "" {
		req.ImplementationType = types.ToolImplementationInternal
	}
	isValidImplementationType := false
	for _, validType := range validImplementationTypes {
		if req.ImplementationType == validType {
			isValidImplementationType = true
			break
		}
	}
	if !isValidImplementationType {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Invalid implementation type"),
			Success: false,
		})
		return
	}

	// Check if tool name already exists
	_, err = h.toolModel.GetByName(orgUUID, req.Name)
	if err == nil {
		c.JSON(http.StatusConflict, types.ErrorResponse{
			Error:   types.NewValidationError("Tool with this name already exists"),
			Success: false,
		})
		return
	}

	// Check if function name already exists
	_, err = h.toolModel.GetByFunctionName(orgUUID, req.FunctionName)
	if err == nil {
		c.JSON(http.StatusConflict, types.ErrorResponse{
			Error:   types.NewValidationError("Tool with this function name already exists"),
			Success: false,
		})
		return
	}

	// Set defaults
	if req.TimeoutSeconds <= 0 {
		req.TimeoutSeconds = 30
	}
	if req.MaxRetries < 0 {
		req.MaxRetries = 3
	}

	// Create tool model
	tool := &models.MCPTool{
		OrganizationID:     orgUUID,
		Name:               req.Name,
		FunctionName:       req.FunctionName,
		Schema:             req.Schema,
		Category:           req.Category,
		ImplementationType: req.ImplementationType,
		TimeoutSeconds:     req.TimeoutSeconds,
		MaxRetries:         req.MaxRetries,
		UsageCount:         0,
		AccessPermissions:  req.AccessPermissions,
		IsActive:           true,
		IsPublic:           req.IsPublic,
		Metadata:           req.Metadata,
		Tags:               req.Tags,
		Examples:           req.Examples,
	}

	if req.Description != "" {
		tool.Description = sql.NullString{String: req.Description, Valid: true}
	}
	if req.EndpointURL != "" {
		tool.EndpointURL = sql.NullString{String: req.EndpointURL, Valid: true}
	}
	if req.Documentation != "" {
		tool.Documentation = sql.NullString{String: req.Documentation, Valid: true}
	}
	if userExists {
		if userUUID, err := uuid.Parse(userID.(string)); err == nil {
			tool.CreatedBy = uuid.NullUUID{UUID: userUUID, Valid: true}
		}
	}

	// Create tool
	err = h.toolModel.Create(tool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Failed to create tool"),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    tool,
	})
}

// GetTool retrieves a specific MCP tool
func (h *ToolHandler) GetTool(c *gin.Context) {
	toolID := c.Param("id")
	if toolID == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Tool ID is required"),
			Success: false,
		})
		return
	}

	toolUUID, err := uuid.Parse(toolID)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Invalid tool ID format"),
			Success: false,
		})
		return
	}

	tool, err := h.toolModel.GetByID(toolUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, types.ErrorResponse{
				Error:   types.NewNotFoundError("Tool not found"),
				Success: false,
			})
		} else {
			c.JSON(http.StatusInternalServerError, types.ErrorResponse{
				Error:   types.NewInternalError("Failed to retrieve tool"),
				Success: false,
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tool,
	})
}

// UpdateTool updates an existing MCP tool
func (h *ToolHandler) UpdateTool(c *gin.Context) {
	toolID := c.Param("id")
	if toolID == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Tool ID is required"),
			Success: false,
		})
		return
	}

	toolUUID, err := uuid.Parse(toolID)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Invalid tool ID format"),
			Success: false,
		})
		return
	}

	var req types.UpdateGlobalToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError(err.Error()),
			Success: false,
		})
		return
	}

	// Get existing tool
	tool, err := h.toolModel.GetByID(toolUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, types.ErrorResponse{
				Error:   types.NewNotFoundError("Tool not found"),
				Success: false,
			})
		} else {
			c.JSON(http.StatusInternalServerError, types.ErrorResponse{
				Error:   types.NewInternalError("Failed to retrieve tool"),
				Success: false,
			})
		}
		return
	}

	// Update fields if provided
	if req.Name != "" {
		tool.Name = req.Name
	}
	if req.Description != "" {
		tool.Description = sql.NullString{String: req.Description, Valid: true}
	}
	if req.FunctionName != "" {
		tool.FunctionName = req.FunctionName
	}
	if req.Schema != nil {
		tool.Schema = req.Schema
	}
	if req.Category != "" {
		// Validate category
		validCategories := []string{
			types.ToolCategoryGeneral,
			types.ToolCategoryData,
			types.ToolCategoryFile,
			types.ToolCategoryWeb,
			types.ToolCategorySystem,
			types.ToolCategoryAI,
			types.ToolCategoryDev,
			types.ToolCategoryCustom,
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
				Error:   types.NewValidationError("Invalid tool category"),
				Success: false,
			})
			return
		}
		tool.Category = req.Category
	}
	if req.ImplementationType != "" {
		// Validate implementation type
		validImplementationTypes := []string{
			types.ToolImplementationInternal,
			types.ToolImplementationExternal,
			types.ToolImplementationWebhook,
			types.ToolImplementationScript,
		}
		isValidImplementationType := false
		for _, validType := range validImplementationTypes {
			if req.ImplementationType == validType {
				isValidImplementationType = true
				break
			}
		}
		if !isValidImplementationType {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Error:   types.NewValidationError("Invalid implementation type"),
				Success: false,
			})
			return
		}
		tool.ImplementationType = req.ImplementationType
	}
	if req.EndpointURL != "" {
		tool.EndpointURL = sql.NullString{String: req.EndpointURL, Valid: true}
	}
	if req.TimeoutSeconds != nil {
		if *req.TimeoutSeconds <= 0 || *req.TimeoutSeconds > 600 {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Error:   types.NewValidationError("Timeout must be between 1 and 600 seconds"),
				Success: false,
			})
			return
		}
		tool.TimeoutSeconds = *req.TimeoutSeconds
	}
	if req.MaxRetries != nil {
		if *req.MaxRetries < 0 || *req.MaxRetries > 10 {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Error:   types.NewValidationError("Max retries must be between 0 and 10"),
				Success: false,
			})
			return
		}
		tool.MaxRetries = *req.MaxRetries
	}
	if req.AccessPermissions != nil {
		tool.AccessPermissions = req.AccessPermissions
	}
	if req.IsActive != nil {
		tool.IsActive = *req.IsActive
	}
	if req.IsPublic != nil {
		tool.IsPublic = *req.IsPublic
	}
	if req.Metadata != nil {
		tool.Metadata = req.Metadata
	}
	if req.Tags != nil {
		tool.Tags = req.Tags
	}
	if req.Examples != nil {
		tool.Examples = req.Examples
	}
	if req.Documentation != "" {
		tool.Documentation = sql.NullString{String: req.Documentation, Valid: true}
	}

	err = h.toolModel.Update(tool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Failed to update tool"),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tool,
	})
}

// DeleteTool soft deletes an MCP tool
func (h *ToolHandler) DeleteTool(c *gin.Context) {
	toolID := c.Param("id")
	if toolID == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Tool ID is required"),
			Success: false,
		})
		return
	}

	toolUUID, err := uuid.Parse(toolID)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Invalid tool ID format"),
			Success: false,
		})
		return
	}

	// Check if tool exists
	_, err = h.toolModel.GetByID(toolUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, types.ErrorResponse{
				Error:   types.NewNotFoundError("Tool not found"),
				Success: false,
			})
		} else {
			c.JSON(http.StatusInternalServerError, types.ErrorResponse{
				Error:   types.NewInternalError("Failed to retrieve tool"),
				Success: false,
			})
		}
		return
	}

	err = h.toolModel.Delete(toolUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Failed to delete tool"),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tool deleted successfully",
	})
}

// ExecuteTool executes a tool (increments usage count and returns tool info)
func (h *ToolHandler) ExecuteTool(c *gin.Context) {
	toolID := c.Param("id")
	if toolID == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Tool ID is required"),
			Success: false,
		})
		return
	}

	toolUUID, err := uuid.Parse(toolID)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Invalid tool ID format"),
			Success: false,
		})
		return
	}

	tool, err := h.toolModel.GetByID(toolUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, types.ErrorResponse{
				Error:   types.NewNotFoundError("Tool not found"),
				Success: false,
			})
		} else {
			c.JSON(http.StatusInternalServerError, types.ErrorResponse{
				Error:   types.NewInternalError("Failed to retrieve tool"),
				Success: false,
			})
		}
		return
	}

	// Increment usage count
	err = h.toolModel.IncrementUsageCount(toolUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Failed to update usage count"),
			Success: false,
		})
		return
	}

	// Return updated tool (increment locally for response)
	tool.UsageCount++
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tool,
		"message": "Tool execution recorded",
	})
}

// GetToolByFunction retrieves a tool by its function name
func (h *ToolHandler) GetToolByFunction(c *gin.Context) {
	functionName := c.Param("function_name")
	if functionName == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Function name is required"),
			Success: false,
		})
		return
	}

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

	tool, err := h.toolModel.GetByFunctionName(orgUUID, functionName)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, types.ErrorResponse{
				Error:   types.NewNotFoundError("Tool not found"),
				Success: false,
			})
		} else {
			c.JSON(http.StatusInternalServerError, types.ErrorResponse{
				Error:   types.NewInternalError("Failed to retrieve tool"),
				Success: false,
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tool,
	})
}

// ListPublicTools lists all public tools available to any organization
func (h *ToolHandler) ListPublicTools(c *gin.Context) {
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

	tools, err := h.toolModel.ListPublicTools(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Failed to retrieve public tools"),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tools,
		"count":   len(tools),
	})
}

// enrichToolsWithServerInfo enriches tools with server information
func (h *ToolHandler) enrichToolsWithServerInfo(tools []*models.MCPTool) ([]*ToolWithServerInfo, error) {
	enrichedTools := make([]*ToolWithServerInfo, len(tools))

	for i, tool := range tools {
		enriched := &ToolWithServerInfo{MCPTool: tool}

		// If tool has a server ID, get the server information
		if tool.ServerID.Valid {
			server, err := h.serverModel.GetByID(tool.ServerID.UUID)
			if err == nil {
				enriched.ServerName = &server.Name
				enriched.ServerProtocol = &server.Protocol
				enriched.ServerStatus = &server.Status
			}
			// If error occurs, we just skip adding server info but don't fail the whole request
		}

		enrichedTools[i] = enriched
	}

	return enrichedTools, nil
}
