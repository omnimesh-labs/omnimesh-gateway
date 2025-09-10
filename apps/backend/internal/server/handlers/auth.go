package handlers

import (
	"net/http"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/auth"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *auth.Service
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req types.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError(err.Error()),
			Success: false,
		})
		return
	}

	response, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req types.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError(err.Error()),
			Success: false,
		})
		return
	}

	response, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Extract token from Authorization header
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("Authorization header required"),
			Success: false,
		})
		return
	}

	// Remove "Bearer " prefix
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	err := h.authService.Logout(token)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}

// CreateAPIKey handles API key creation
func (h *AuthHandler) CreateAPIKey(c *gin.Context) {
	var req types.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError(err.Error()),
			Success: false,
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error:   types.NewUnauthorizedError("User not authenticated"),
			Success: false,
		})
		return
	}

	// Get user role from context
	userRole, exists := c.Get("role")
	if !exists || userRole != types.RoleAdmin {
		c.JSON(http.StatusForbidden, types.ErrorResponse{
			Error:   types.NewForbiddenError("Only admins can create API keys"),
			Success: false,
		})
		return
	}

	apiKey, err := h.authService.CreateAPIKey(userID.(string), &req)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    apiKey,
	})
}

// GetProfile returns the current user's profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error:   types.NewUnauthorizedError("User not authenticated"),
			Success: false,
		})
		return
	}

	user, err := h.authService.GetUserByID(userID.(string))
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
	})
}

// UpdateProfile updates the current user's profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req types.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError(err.Error()),
			Success: false,
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error:   types.NewUnauthorizedError("User not authenticated"),
			Success: false,
		})
		return
	}

	user, err := h.authService.UpdateUser(userID.(string), &req)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
	})
}

// ListAPIKeys returns all API keys (admin only)
func (h *AuthHandler) ListAPIKeys(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error:   types.NewUnauthorizedError("User not authenticated"),
			Success: false,
		})
		return
	}

	// Get user role from context
	userRole, exists := c.Get("role")
	if !exists || userRole != types.RoleAdmin {
		c.JSON(http.StatusForbidden, types.ErrorResponse{
			Error:   types.NewForbiddenError("Only admins can list API keys"),
			Success: false,
		})
		return
	}

	// Get organization ID from context
	orgID, exists := c.Get("organization_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Organization ID not found in context"),
			Success: false,
		})
		return
	}

	// List all API keys for the organization (admins can see all)
	keys, err := h.authService.ListAllAPIKeys(orgID.(string))
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    keys,
	})
}

// DeleteAPIKey deletes an API key
func (h *AuthHandler) DeleteAPIKey(c *gin.Context) {
	keyID := c.Param("id")
	if keyID == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   types.NewValidationError("API key ID required"),
			Success: false,
		})
		return
	}

	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error:   types.NewUnauthorizedError("User not authenticated"),
			Success: false,
		})
		return
	}

	// Get user role from context
	userRole, exists := c.Get("role")
	if !exists || userRole != types.RoleAdmin {
		c.JSON(http.StatusForbidden, types.ErrorResponse{
			Error:   types.NewForbiddenError("Only admins can delete API keys"),
			Success: false,
		})
		return
	}

	// Get organization ID from context
	orgID, exists := c.Get("organization_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   types.NewInternalError("Organization ID not found in context"),
			Success: false,
		})
		return
	}

	// Admin can delete any API key in their organization
	err := h.authService.DeleteAPIKeyByAdmin(orgID.(string), keyID)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "API key deleted successfully",
	})
}
