package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/database/models"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/plugins"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FiltersHandler handles filter-related requests
type FiltersHandler struct {
	db            *sql.DB
	pluginService plugins.PluginService
}

// NewFiltersHandler creates a new filters handler
func NewFiltersHandler(db *sql.DB, pluginService plugins.PluginService) *FiltersHandler {
	return &FiltersHandler{
		db:            db,
		pluginService: pluginService,
	}
}

// CreateFilter handles POST /api/admin/filters
func (h *FiltersHandler) CreateFilter(c *gin.Context) {
	var req CreateFilterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	// Get user context from middleware
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	orgID, exists := c.Get("organization_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization not found"})
		return
	}

	// Validate filter configuration
	factory, err := h.pluginService.GetRegistry().Get(plugins.PluginType(req.Type))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter type", "details": err.Error()})
		return
	}

	if err := factory.ValidateConfig(req.Config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter configuration", "details": err.Error()})
		return
	}

	// Create filter model
	userIDStr := userID.(string)
	filter := models.NewContentFilter(
		orgID.(string),
		req.Name,
		req.Description,
		req.Type,
		req.Enabled,
		req.Priority,
		req.Config,
		&userIDStr,
	)

	// Save to database
	configJSON, err := json.Marshal(filter.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal filter config"})
		return
	}

	query := `
		INSERT INTO content_filters (id, organization_id, name, description, type, enabled, priority, config, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	filter.ID = uuid.New().String()
	err = h.db.QueryRow(query,
		filter.ID, filter.OrganizationID, filter.Name, filter.Description,
		filter.Type, filter.Enabled, filter.Priority, configJSON,
		filter.CreatedBy,
	).Scan(&filter.ID, &filter.CreatedAt, &filter.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create filter", "details": err.Error()})
		return
	}

	// Reload filters for this organization
	if err := h.pluginService.ReloadOrganizationPlugins(c.Request.Context(), orgID.(string)); err != nil {
		// Log error but don't fail the request
		c.Header("X-Warning", "Filter created but failed to reload filters")
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Filter created successfully",
		"filter":  filter,
	})
}

// ListFilters handles GET /api/admin/filters
func (h *FiltersHandler) ListFilters(c *gin.Context) {
	orgID, exists := c.Get("organization_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization not found"})
		return
	}

	// Query parameters for filtering and pagination
	filterType := c.Query("type")
	isActiveStr := c.Query("enabled")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	// Build query with optional filters
	query := `
		SELECT id, organization_id, name, description, type, enabled, priority, config, created_at, updated_at, created_by
		FROM content_filters
		WHERE organization_id = $1
	`
	args := []interface{}{orgID.(string)}
	argIndex := 2

	if filterType != "" {
		query += " AND type = $" + strconv.Itoa(argIndex)
		args = append(args, filterType)
		argIndex++
	}

	if isActiveStr != "" {
		isActive, err := strconv.ParseBool(isActiveStr)
		if err == nil {
			query += " AND enabled = $" + strconv.Itoa(argIndex)
			args = append(args, isActive)
			argIndex++
		}
	}

	query += " ORDER BY priority ASC, created_at DESC LIMIT $" + strconv.Itoa(argIndex) + " OFFSET $" + strconv.Itoa(argIndex+1)
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch filters", "details": err.Error()})
		return
	}
	defer rows.Close()

	var filtersList []models.ContentFilter
	for rows.Next() {
		var filter models.ContentFilter
		var configJSON []byte

		err := rows.Scan(
			&filter.ID, &filter.OrganizationID, &filter.Name, &filter.Description,
			&filter.Type, &filter.Enabled, &filter.Priority, &configJSON,
			&filter.CreatedAt, &filter.UpdatedAt, &filter.CreatedBy,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan filter", "details": err.Error()})
			return
		}

		// Parse JSON config
		if err := json.Unmarshal(configJSON, &filter.Config); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse filter config", "details": err.Error()})
			return
		}

		filtersList = append(filtersList, filter)
	}

	c.JSON(http.StatusOK, gin.H{
		"filters": filtersList,
		"total":   len(filtersList),
		"limit":   limit,
		"offset":  offset,
	})
}

// GetFilter handles GET /api/admin/filters/:id
func (h *FiltersHandler) GetFilter(c *gin.Context) {
	filterID := c.Param("id")
	if filterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filter ID is required"})
		return
	}

	orgID, exists := c.Get("organization_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization not found"})
		return
	}

	query := `
		SELECT id, organization_id, name, description, type, enabled, priority, config, created_at, updated_at, created_by
		FROM content_filters
		WHERE id = $1 AND organization_id = $2
	`

	var filter models.ContentFilter
	var configJSON []byte

	err := h.db.QueryRow(query, filterID, orgID.(string)).Scan(
		&filter.ID, &filter.OrganizationID, &filter.Name, &filter.Description,
		&filter.Type, &filter.Enabled, &filter.Priority, &configJSON,
		&filter.CreatedAt, &filter.UpdatedAt, &filter.CreatedBy,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Filter not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch filter", "details": err.Error()})
		return
	}

	// Parse JSON config
	if err := json.Unmarshal(configJSON, &filter.Config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse filter config", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"filter": filter})
}

// UpdateFilter handles PUT /api/admin/filters/:id
func (h *FiltersHandler) UpdateFilter(c *gin.Context) {
	filterID := c.Param("id")
	if filterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filter ID is required"})
		return
	}

	var req UpdateFilterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	orgID, exists := c.Get("organization_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization not found"})
		return
	}

	// Check if filter exists and belongs to organization
	var existingFilter models.ContentFilter
	checkQuery := "SELECT id, type FROM content_filters WHERE id = $1 AND organization_id = $2"
	err := h.db.QueryRow(checkQuery, filterID, orgID.(string)).Scan(&existingFilter.ID, &existingFilter.Type)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Filter not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check filter", "details": err.Error()})
		return
	}

	// Validate filter configuration if provided
	if req.Config != nil {
		factory, err := h.pluginService.GetRegistry().Get(plugins.PluginType(existingFilter.Type))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter type", "details": err.Error()})
			return
		}

		if err := factory.ValidateConfig(req.Config); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter configuration", "details": err.Error()})
			return
		}
	}

	// Build update query dynamically
	updateFields := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Name != "" {
		updateFields = append(updateFields, "name = $"+strconv.Itoa(argIndex))
		args = append(args, req.Name)
		argIndex++
	}

	if req.Description != "" {
		updateFields = append(updateFields, "description = $"+strconv.Itoa(argIndex))
		args = append(args, req.Description)
		argIndex++
	}

	if req.Priority != 0 {
		updateFields = append(updateFields, "priority = $"+strconv.Itoa(argIndex))
		args = append(args, req.Priority)
		argIndex++
	}

	if req.Config != nil {
		configJSON, err := json.Marshal(req.Config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal filter config"})
			return
		}
		updateFields = append(updateFields, "config = $"+strconv.Itoa(argIndex))
		args = append(args, configJSON)
		argIndex++
	}

	if req.Enabled != nil {
		updateFields = append(updateFields, "enabled = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Enabled)
		argIndex++
	}

	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	// Add updated_at field
	updateFields = append(updateFields, "updated_at = NOW()")

	// Add WHERE conditions
	args = append(args, filterID, orgID.(string))
	whereClause := " WHERE id = $" + strconv.Itoa(argIndex) + " AND organization_id = $" + strconv.Itoa(argIndex+1)

	query := "UPDATE content_filters SET " + updateFields[0]
	for i := 1; i < len(updateFields); i++ {
		query += ", " + updateFields[i]
	}
	query += whereClause

	_, err = h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update filter", "details": err.Error()})
		return
	}

	// Reload filters for this organization
	if err := h.pluginService.ReloadOrganizationPlugins(c.Request.Context(), orgID.(string)); err != nil {
		c.Header("X-Warning", "Filter updated but failed to reload filters")
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Filter updated successfully",
		"filter_id": filterID,
	})
}

// DeleteFilter handles DELETE /api/admin/filters/:id
func (h *FiltersHandler) DeleteFilter(c *gin.Context) {
	filterID := c.Param("id")
	if filterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filter ID is required"})
		return
	}

	orgID, exists := c.Get("organization_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization not found"})
		return
	}

	// Check if filter exists and belongs to organization
	var existingFilter models.ContentFilter
	checkQuery := "SELECT id FROM content_filters WHERE id = $1 AND organization_id = $2"
	err := h.db.QueryRow(checkQuery, filterID, orgID.(string)).Scan(&existingFilter.ID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Filter not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check filter", "details": err.Error()})
		return
	}

	// Delete the filter
	deleteQuery := "DELETE FROM content_filters WHERE id = $1 AND organization_id = $2"
	_, err = h.db.Exec(deleteQuery, filterID, orgID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete filter", "details": err.Error()})
		return
	}

	// Reload filters for this organization
	if err := h.pluginService.ReloadOrganizationPlugins(c.Request.Context(), orgID.(string)); err != nil {
		c.Header("X-Warning", "Filter deleted but failed to reload filters")
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Filter deleted successfully",
		"filter_id": filterID,
	})
}

// GetFilterTypes handles GET /api/admin/filters/types
func (h *FiltersHandler) GetFilterTypes(c *gin.Context) {
	filterTypes, err := h.pluginService.GetRegistry().GetAllInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get filter types", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"filter_types": filterTypes,
	})
}

// GetFilterViolations handles GET /api/admin/filters/violations
func (h *FiltersHandler) GetFilterViolations(c *gin.Context) {
	orgID, exists := c.Get("organization_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization not found"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	violations, err := h.pluginService.GetViolations(c.Request.Context(), orgID.(string), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get violations", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"violations": violations,
		"total":      len(violations),
		"limit":      limit,
		"offset":     offset,
	})
}

// GetFilterMetrics handles GET /api/admin/filters/metrics
func (h *FiltersHandler) GetFilterMetrics(c *gin.Context) {
	metrics, err := h.pluginService.GetMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get metrics", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
	})
}

// Request/Response types
type CreateFilterRequest struct {
	Config      map[string]interface{} `json:"config" binding:"required"`
	Name        string                 `json:"name" binding:"required,min=2"`
	Description string                 `json:"description"`
	Type        string                 `json:"type" binding:"required"`
	Priority    int                    `json:"priority" binding:"min=1,max=1000"`
	Enabled     bool                   `json:"enabled"`
}

type UpdateFilterRequest struct {
	Config      map[string]interface{} `json:"config,omitempty"`
	Enabled     *bool                  `json:"enabled,omitempty"`
	Name        string                 `json:"name,omitempty" binding:"omitempty,min=2"`
	Description string                 `json:"description,omitempty"`
	Priority    int                    `json:"priority,omitempty" binding:"omitempty,min=1,max=1000"`
}
