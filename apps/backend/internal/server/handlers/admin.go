package handlers

import (
	"net/http"
	"strconv"
	"time"

	"mcp-gateway/apps/backend/internal/auth"
	"mcp-gateway/apps/backend/internal/config"
	"mcp-gateway/apps/backend/internal/logging"
	"mcp-gateway/apps/backend/internal/ratelimit"
	"mcp-gateway/apps/backend/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminHandler handles administrative endpoints
type AdminHandler struct {
	authService      *auth.Service
	loggingService   *logging.Service
	rateLimitService *ratelimit.Service
	configService    *config.Service
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(authService *auth.Service, loggingService *logging.Service, rateLimitService *ratelimit.Service, configService *config.Service) *AdminHandler {
	return &AdminHandler{
		authService:      authService,
		loggingService:   loggingService,
		rateLimitService: rateLimitService,
		configService:    configService,
	}
}

// ListUsers lists all users in the organization
func (h *AdminHandler) ListUsers(c *gin.Context) {
	// TODO: Implement user listing with pagination
	// orgID, _ := c.Get("organization_id")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    []interface{}{},
		"pagination": gin.H{
			"page":  1,
			"limit": 10,
			"total": 0,
		},
	})
}

// CreateUser creates a new user
func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req types.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError(err.Error()),
			Success: false,
		})
		return
	}

	// Set organization ID from context
	if orgID, exists := c.Get("organization_id"); exists {
		req.OrganizationID = orgID.(string)
	}

	// TODO: Implement user creation logic
	user, err := h.authService.CreateUser(&req)
	if err != nil {
		c.JSON(types.GetStatusCode(err), types.ErrorResponse{
			Error:   err.(*types.Error),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    user,
	})
}

// GetUser retrieves a specific user
func (h *AdminHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("User ID is required"),
			Success: false,
		})
		return
	}

	// TODO: Implement user retrieval logic
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		c.JSON(types.GetStatusCode(err), types.ErrorResponse{
			Error:   err.(*types.Error),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
	})
}

// UpdateUser updates a user
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("User ID is required"),
			Success: false,
		})
		return
	}

	var req types.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError(err.Error()),
			Success: false,
		})
		return
	}

	// TODO: Implement user update logic
	user, err := h.authService.UpdateUser(userID, &req)
	if err != nil {
		c.JSON(types.GetStatusCode(err), types.ErrorResponse{
			Error:   err.(*types.Error),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
	})
}

// DeleteUser deletes a user
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("User ID is required"),
			Success: false,
		})
		return
	}

	// TODO: Implement user deletion logic
	err := h.authService.DeleteUser(userID)
	if err != nil {
		c.JSON(types.GetStatusCode(err), types.ErrorResponse{
			Error:   err.(*types.Error),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User deleted successfully",
	})
}

// GetLogs retrieves system logs
func (h *AdminHandler) GetLogs(c *gin.Context) {
	// Parse query parameters
	query := &types.LogQueryRequest{}

	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			query.StartTime = t
		}
	}

	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			query.EndTime = t
		}
	}

	query.Level = c.Query("level")
	query.Type = c.Query("type")
	query.UserID = c.Query("user_id")
	query.Method = c.Query("method")
	query.Path = c.Query("path")
	query.Search = c.Query("search")

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			query.Limit = l
		}
	} else {
		query.Limit = 100
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			query.Offset = o
		}
	}

	// Set organization ID from context
	if orgID, exists := c.Get("organization_id"); exists {
		query.OrganizationID = orgID.(string)
	}

	// Convert types.LogQueryRequest to logging.QueryRequest
	logQuery := &logging.QueryRequest{
		StartTime: &query.StartTime,
		EndTime:   &query.EndTime,
		Level:     logging.LogLevel(query.Level),
		UserID:    query.UserID,
		OrgID:     query.OrganizationID,
		Message:   query.Search,
		Limit:     query.Limit,
		Offset:    query.Offset,
	}

	// Add method and path to filters if provided
	if query.Method != "" || query.Path != "" {
		logQuery.Filters = make(map[string]interface{})
		if query.Method != "" {
			logQuery.Filters["method"] = query.Method
		}
		if query.Path != "" {
			logQuery.Filters["path"] = query.Path
		}
	}

	// TODO: Implement log retrieval logic
	logs, err := h.loggingService.Query(c.Request.Context(), logQuery)
	if err != nil {
		c.JSON(types.GetStatusCode(err), types.ErrorResponse{
			Error:   err.(*types.Error),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
	})
}

// GetAuditLogs retrieves audit logs
func (h *AdminHandler) GetAuditLogs(c *gin.Context) {
	// Parse query parameters
	limit := 100
	offset := 0
	
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}
	
	resourceType := c.Query("resource_type")
	action := c.Query("action")
	actorID := c.Query("actor_id")
	
	// TODO: Set organization ID from context
	// orgID, _ := c.Get("organization_id")
	
	// TODO: Implement audit log retrieval logic
	// For now, return empty result with proper structure
	auditLogs := []gin.H{}
	
	// Filter logic would be implemented here
	_ = resourceType
	_ = action
	_ = actorID
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    auditLogs,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"total":  0,
		},
	})
}

// GetStats returns system statistics
func (h *AdminHandler) GetStats(c *gin.Context) {
	orgID, _ := c.Get("organization_id")

	// TODO: Aggregate statistics from various services
	stats := gin.H{
		"users": gin.H{
			"total":  0,
			"active": 0,
		},
		"servers": gin.H{
			"total":   0,
			"healthy": 0,
		},
		"requests": gin.H{
			"total":      0,
			"successful": 0,
			"failed":     0,
		},
		"rate_limits": gin.H{
			"total_limits": 0,
			"blocked":      0,
		},
	}

	if orgID != nil {
		// Get organization-specific stats
		// rateLimitStats, _ := h.rateLimitService.GetStats(orgID.(string))
		// logStats, _ := h.loggingService.GetLogStats(orgID.(string), time.Now().Add(-24*time.Hour), time.Now())
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetMetrics returns Prometheus-style metrics
func (h *AdminHandler) GetMetrics(c *gin.Context) {
	// TODO: Implement Prometheus metrics export
	// This should return metrics in Prometheus text format
	c.Header("Content-Type", "text/plain; version=0.0.4")
	c.String(http.StatusOK, `# HELP mcp_gateway_requests_total Total number of requests
# TYPE mcp_gateway_requests_total counter
mcp_gateway_requests_total{method="GET",status="200"} 100
mcp_gateway_requests_total{method="POST",status="201"} 50

# HELP mcp_gateway_request_duration_seconds Request duration in seconds
# TYPE mcp_gateway_request_duration_seconds histogram
mcp_gateway_request_duration_seconds_bucket{le="0.1"} 80
mcp_gateway_request_duration_seconds_bucket{le="0.5"} 120
mcp_gateway_request_duration_seconds_bucket{le="1.0"} 140
mcp_gateway_request_duration_seconds_bucket{le="+Inf"} 150
mcp_gateway_request_duration_seconds_sum 45.2
mcp_gateway_request_duration_seconds_count 150
`)
}

// ExportConfiguration exports configuration entities based on the request
func (h *AdminHandler) ExportConfiguration(c *gin.Context) {
	var req types.ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError(err.Error()),
			Success: false,
		})
		return
	}

	orgIDStr, exists := c.Get("organization_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error:   types.NewUnauthorizedError("Organization ID not found"),
			Success: false,
		})
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error:   types.NewUnauthorizedError("User ID not found"),
			Success: false,
		})
		return
	}

	// Parse UUIDs
	orgID, err := uuid.Parse(orgIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Invalid organization ID"),
			Success: false,
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Invalid user ID"),
			Success: false,
		})
		return
	}

	// Export configuration using the config service
	export, err := h.configService.ExportConfiguration(c.Request.Context(), orgID, userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Failed to export configuration: " + err.Error()),
			Success: false,
		})
		return
	}

	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", `attachment; filename="mcp-gateway-config-export.json"`)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    export,
	})
}

// ImportConfiguration imports configuration from an uploaded file
func (h *AdminHandler) ImportConfiguration(c *gin.Context) {
	var req types.ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError(err.Error()),
			Success: false,
		})
		return
	}

	orgIDStr, exists := c.Get("organization_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error:   types.NewUnauthorizedError("Organization ID not found"),
			Success: false,
		})
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error:   types.NewUnauthorizedError("User ID not found"),
			Success: false,
		})
		return
	}

	// Parse UUIDs
	orgID, err := uuid.Parse(orgIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Invalid organization ID"),
			Success: false,
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Invalid user ID"),
			Success: false,
		})
		return
	}

	// Import configuration using the config service
	result, err := h.configService.ImportConfiguration(c.Request.Context(), orgID, userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Failed to import configuration: " + err.Error()),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// ValidateImport validates an import file without performing the import
func (h *AdminHandler) ValidateImport(c *gin.Context) {
	var req types.ValidateImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError(err.Error()),
			Success: false,
		})
		return
	}

	orgIDStr, exists := c.Get("organization_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error:   types.NewUnauthorizedError("Organization ID not found"),
			Success: false,
		})
		return
	}

	// Parse UUID
	orgID, err := uuid.Parse(orgIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Invalid organization ID"),
			Success: false,
		})
		return
	}

	// Validate import data using the config service
	validation, err := h.configService.ValidateImport(c.Request.Context(), orgID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Failed to validate import: " + err.Error()),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    validation,
	})
}

// GetImportHistory returns the history of import operations
func (h *AdminHandler) GetImportHistory(c *gin.Context) {
	// Parse query parameters
	query := &types.ImportHistoryQuery{}
	
	if status := c.Query("status"); status != "" {
		query.Status = types.ImportStatus(status)
	}
	
	query.EntityType = c.Query("entity_type")
	query.ImportedBy = c.Query("imported_by")
	
	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			query.StartDate = &t
		}
	}
	
	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			query.EndDate = &t
		}
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 1000 {
			query.Limit = l
		}
	}
	if query.Limit == 0 {
		query.Limit = 50
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			query.Offset = o
		}
	}

	orgIDStr, exists := c.Get("organization_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error:   types.NewUnauthorizedError("Organization ID not found"),
			Success: false,
		})
		return
	}

	// Parse UUID
	orgID, err := uuid.Parse(orgIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Invalid organization ID"),
			Success: false,
		})
		return
	}

	// Get import history using the config service
	history, total, err := h.configService.GetImportHistory(c.Request.Context(), orgID, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Failed to get import history: " + err.Error()),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    history,
		"pagination": gin.H{
			"limit":  query.Limit,
			"offset": query.Offset,
			"total":  total,
		},
	})
}
