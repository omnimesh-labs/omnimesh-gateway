package auth

import (
	"net/http"
	"strings"

	"mcp-gateway/apps/backend/internal/types"

	"github.com/gin-gonic/gin"
)

// ServiceInterface defines the methods needed by the middleware
type ServiceInterface interface {
	GetUserByID(userID string) (*types.User, error)
	ValidateAPIKey(apiKey string) (*types.APIKey, error)
}

// Ensure Service implements ServiceInterface
var _ ServiceInterface = (*Service)(nil)

// Middleware handles authentication and authorization
type Middleware struct {
	jwtManager *JWTManager
	service    ServiceInterface
}

// NewMiddleware creates a new auth middleware
func NewMiddleware(jwtManager *JWTManager, service *Service) *Middleware {
	return &Middleware{
		jwtManager: jwtManager,
		service:    service,
	}
}

// NewMiddlewareWithInterface creates a new auth middleware with interface
func NewMiddlewareWithInterface(jwtManager *JWTManager, service ServiceInterface) *Middleware {
	return &Middleware{
		jwtManager: jwtManager,
		service:    service,
	}
}

// RequireAuth middleware that requires valid authentication
func (m *Middleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		token := m.extractToken(c)
		if token == "" {
			m.respondWithError(c, http.StatusUnauthorized, "Authorization header required")
			return
		}

		// Validate token
		claims, err := m.jwtManager.ValidateToken(token)
		if err != nil {
			m.respondWithError(c, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Ensure it's an access token
		if claims.TokenType != "access" {
			m.respondWithError(c, http.StatusUnauthorized, "Invalid token type")
			return
		}

		// Get user information
		user, err := m.service.GetUserByID(claims.UserID)
		if err != nil {
			m.respondWithError(c, http.StatusUnauthorized, "User not found")
			return
		}

		// Check if user account is still active
		if !user.IsActive {
			m.respondWithError(c, http.StatusUnauthorized, "User account is inactive")
			return
		}

		// Set user context
		m.setUserContext(c, user)
		c.Next()
	}
}

// RequireRole middleware that requires specific role
func (m *Middleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context (set by RequireAuth middleware)
		userRole, exists := c.Get("role")
		if !exists {
			m.respondWithError(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		// Check role hierarchy: system_admin > admin > user > viewer
		if !m.hasRequiredRole(userRole.(string), requiredRole) {
			m.respondWithError(c, http.StatusForbidden, "Insufficient permissions")
			return
		}

		c.Next()
	}
}

// RequireAPIKey middleware for API key authentication
func (m *Middleware) RequireAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract API key from X-API-Key header
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			m.respondWithError(c, http.StatusUnauthorized, "API key required")
			return
		}

		// Validate API key (placeholder - implementation in service)
		validatedKey, err := m.service.ValidateAPIKey(apiKey)
		if err != nil {
			m.respondWithError(c, http.StatusUnauthorized, "Invalid API key")
			return
		}

		// Get user associated with API key
		user, err := m.service.GetUserByID(validatedKey.UserID)
		if err != nil || !user.IsActive {
			m.respondWithError(c, http.StatusUnauthorized, "API key user not found or inactive")
			return
		}

		// Set user context and API key info
		m.setUserContext(c, user)
		c.Set("api_key", validatedKey)
		c.Next()
	}
}

// OptionalAuth middleware that allows optional authentication
func (m *Middleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to extract token from Authorization header
		token := m.extractToken(c)
		if token == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		// Try to validate token
		claims, err := m.jwtManager.ValidateToken(token)
		if err != nil {
			// Invalid token, continue without authentication
			c.Next()
			return
		}

		// Ensure it's an access token
		if claims.TokenType != "access" {
			// Wrong token type, continue without authentication
			c.Next()
			return
		}

		// Get user information
		user, err := m.service.GetUserByID(claims.UserID)
		if err != nil || !user.IsActive {
			// User not found or inactive, continue without authentication
			c.Next()
			return
		}

		// Set user context if everything is valid
		m.setUserContext(c, user)
		c.Next()
	}
}

// extractToken extracts JWT token from request
func (m *Middleware) extractToken(c *gin.Context) string {
	// Get Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check if it starts with "Bearer "
	const bearerSchema = "Bearer "
	if len(authHeader) > len(bearerSchema) && strings.HasPrefix(authHeader, bearerSchema) {
		return authHeader[len(bearerSchema):]
	}

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
	var errorCode string
	switch status {
	case http.StatusUnauthorized:
		errorCode = types.ErrCodeUnauthorized
	case http.StatusForbidden:
		errorCode = types.ErrCodeInsufficientRights
	default:
		errorCode = types.ErrCodeInternalError
	}

	c.JSON(status, types.ErrorResponse{
		Error:   types.NewError(errorCode, message, status),
		Success: false,
	})
	c.Abort()
}

// hasRequiredRole checks if user role meets the required role level
func (m *Middleware) hasRequiredRole(userRole, requiredRole string) bool {
	roleHierarchy := map[string]int{
		types.RoleSystemAdmin: 4,
		types.RoleAdmin:       3,
		types.RoleUser:        2,
		types.RoleViewer:      1,
		types.RoleAPIUser:     1, // Same level as viewer
	}

	userLevel := roleHierarchy[userRole]
	requiredLevel := roleHierarchy[requiredRole]

	return userLevel >= requiredLevel
}
