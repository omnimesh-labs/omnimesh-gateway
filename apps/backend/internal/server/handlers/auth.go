package handlers

import (
	"net/http"

	"mcp-gateway/apps/backend/internal/auth"
	"mcp-gateway/apps/backend/internal/types"

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

	// TODO: Implement login logic
	response, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(types.GetStatusCode(err), types.ErrorResponse{
			Error:   err.(*types.Error),
			Success: false,
		})
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

	// TODO: Implement token refresh logic
	response, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(types.GetStatusCode(err), types.ErrorResponse{
			Error:   err.(*types.Error),
			Success: false,
		})
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

	// TODO: Implement logout logic
	err := h.authService.Logout(token)
	if err != nil {
		c.JSON(types.GetStatusCode(err), types.ErrorResponse{
			Error:   err.(*types.Error),
			Success: false,
		})
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

	// TODO: Implement API key creation logic
	apiKey, err := h.authService.CreateAPIKey(userID.(string), &req)
	if err != nil {
		c.JSON(types.GetStatusCode(err), types.ErrorResponse{
			Error:   err.(*types.Error),
			Success: false,
		})
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

	// TODO: Implement profile retrieval logic
	user, err := h.authService.GetUserByID(userID.(string))
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

	// TODO: Implement profile update logic
	user, err := h.authService.UpdateUser(userID.(string), &req)
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
