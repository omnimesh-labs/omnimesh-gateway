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
	rbac       *RBAC
}

// NewMiddleware creates a new auth middleware
func NewMiddleware(jwtManager *JWTManager, service *Service) *Middleware {
	return &Middleware{
		jwtManager: jwtManager,
		service:    service,
		rbac:       NewRBAC(),
	}
}

// NewMiddlewareWithInterface creates a new auth middleware with interface
func NewMiddlewareWithInterface(jwtManager *JWTManager, service ServiceInterface) *Middleware {
	return &Middleware{
		jwtManager: jwtManager,
		service:    service,
		rbac:       NewRBAC(),
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
	return m.rbac.HasRequiredRole(userRole, requiredRole)
}

// RequirePermission middleware that requires specific permission
func (m *Middleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			m.respondWithError(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		if !m.rbac.HasPermission(userRole.(string), permission) {
			m.respondWithError(c, http.StatusForbidden, "Insufficient permissions")
			return
		}

		c.Next()
	}
}

// RequireAnyPermission middleware that requires any of the specified permissions
func (m *Middleware) RequireAnyPermission(permissions []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			m.respondWithError(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		if !m.rbac.HasAnyPermission(userRole.(string), permissions) {
			m.respondWithError(c, http.StatusForbidden, "Insufficient permissions")
			return
		}

		c.Next()
	}
}

// RequireAllPermissions middleware that requires all of the specified permissions
func (m *Middleware) RequireAllPermissions(permissions []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			m.respondWithError(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		if !m.rbac.HasAllPermissions(userRole.(string), permissions) {
			m.respondWithError(c, http.StatusForbidden, "Insufficient permissions")
			return
		}

		c.Next()
	}
}

// RequireResourceAccess middleware for resource-based access control
func (m *Middleware) RequireResourceAccess(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			m.respondWithError(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		if !m.rbac.CanAccessResource(userRole.(string), resource, action) {
			m.respondWithError(c, http.StatusForbidden, "Insufficient permissions for this resource")
			return
		}

		c.Next()
	}
}

// RequireOrganizationAccess middleware for organization-level access control
func (m *Middleware) RequireOrganizationAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		userOrgID, exists := c.Get("organization_id")
		if !exists {
			m.respondWithError(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		// Get organization ID from request (path parameter, query, or body)
		requestOrgID := m.extractOrganizationID(c)
		
		// System admins can access any organization
		userRole, _ := c.Get("role")
		if m.rbac.IsSystemAdmin(userRole.(string)) {
			c.Next()
			return
		}

		// For other roles, organization IDs must match
		if requestOrgID != "" && requestOrgID != userOrgID.(string) {
			m.respondWithError(c, http.StatusForbidden, "Access denied: organization mismatch")
			return
		}

		c.Next()
	}
}

// RequireSystemAdmin middleware that requires system admin role
func (m *Middleware) RequireSystemAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			m.respondWithError(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		if !m.rbac.IsSystemAdmin(userRole.(string)) {
			m.respondWithError(c, http.StatusForbidden, "System admin access required")
			return
		}

		c.Next()
	}
}

// RequireAdmin middleware that requires admin role or higher
func (m *Middleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			m.respondWithError(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		if !m.rbac.IsAdmin(userRole.(string)) {
			m.respondWithError(c, http.StatusForbidden, "Admin access required")
			return
		}

		c.Next()
	}
}

// RequireUser middleware that requires user role or higher
func (m *Middleware) RequireUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			m.respondWithError(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		if !m.rbac.IsUser(userRole.(string)) {
			m.respondWithError(c, http.StatusForbidden, "User access required")
			return
		}

		c.Next()
	}
}

// extractOrganizationID extracts organization ID from request context
func (m *Middleware) extractOrganizationID(c *gin.Context) string {
	// Try path parameter first
	if orgID := c.Param("organization_id"); orgID != "" {
		return orgID
	}
	
	// Try query parameter
	if orgID := c.Query("organization_id"); orgID != "" {
		return orgID
	}
	
	// Try header
	if orgID := c.GetHeader("X-Organization-ID"); orgID != "" {
		return orgID
	}
	
	// Could also parse from request body if needed
	return ""
}
