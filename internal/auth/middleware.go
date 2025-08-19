package auth

import (
	"mcp-gateway/internal/types"

	"github.com/gin-gonic/gin"
)

// Middleware handles authentication and authorization
type Middleware struct {
	jwtManager *JWTManager
	service    *Service
}

// NewMiddleware creates a new auth middleware
func NewMiddleware(jwtManager *JWTManager, service *Service) *Middleware {
	return &Middleware{
		jwtManager: jwtManager,
		service:    service,
	}
}

// RequireAuth middleware that requires valid authentication
func (m *Middleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement authentication middleware
		// Extract token from header
		// Validate token
		// Set user context
		c.Next()
	}
}

// RequireRole middleware that requires specific role
func (m *Middleware) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement role-based authorization
		// Check user role from context
		// Validate against required role
		c.Next()
	}
}

// RequireAPIKey middleware for API key authentication
func (m *Middleware) RequireAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement API key authentication
		// Extract API key from header
		// Validate API key
		// Set user context
		c.Next()
	}
}

// OptionalAuth middleware that allows optional authentication
func (m *Middleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement optional authentication
		// Try to extract and validate token
		// Set user context if valid, continue if not
		c.Next()
	}
}

// extractToken extracts JWT token from request
func (m *Middleware) extractToken(c *gin.Context) string {
	// TODO: Implement token extraction from Authorization header
	return ""
}

// setUserContext sets user information in gin context
func (m *Middleware) setUserContext(c *gin.Context, user *types.User) {
	c.Set("user", user)
	c.Set("user_id", user.ID)
	c.Set("organization_id", user.OrganizationID)
	c.Set("role", user.Role)
}

// respondWithError sends error response
func (m *Middleware) respondWithError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
	c.Abort()
}
