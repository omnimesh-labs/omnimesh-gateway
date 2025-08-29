package middleware

import (
	"context"
	"fmt"
	"mcp-gateway/apps/backend/internal/types"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// EndpointService interface for endpoint operations
type EndpointService interface {
	ResolveEndpoint(ctx context.Context, name string) (*types.EndpointConfig, error)
}

// EndpointLookupMiddleware resolves endpoint by name from URL
func EndpointLookupMiddleware(endpointService EndpointService) gin.HandlerFunc {
	return func(c *gin.Context) {
		endpointName := c.Param("endpoint_name")
		if endpointName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Endpoint name is required"})
			c.Abort()
			return
		}

		// Resolve endpoint
		config, err := endpointService.ResolveEndpoint(c.Request.Context(), endpointName)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Endpoint not found"})
			c.Abort()
			return
		}

		// Store in context
		c.Set("endpoint", config.Endpoint)
		c.Set("namespace", config.Namespace)
		c.Set("endpoint_config", config)

		c.Next()
	}
}

// EndpointAuthService interface for validating API keys
type EndpointAuthService interface {
	ValidateAPIKey(apiKey string) (*types.APIKey, error)
	GetUserByID(userID string) (*types.User, error)
}

// EndpointAuthMiddleware validates access to endpoint based on its auth settings
func EndpointAuthMiddleware(endpointService EndpointService, authService EndpointAuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		endpointVal, exists := c.Get("endpoint")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Endpoint not found in context"})
			c.Abort()
			return
		}

		endpoint := endpointVal.(*types.Endpoint)

		// Check if endpoint is active
		if !endpoint.IsActive {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"details": "Endpoint is not active",
			})
			c.Abort()
			return
		}

		// If public access is enabled, allow without authentication
		if endpoint.EnablePublicAccess {
			c.Next()
			return
		}

		// Check authentication based on endpoint settings
		authenticated := false

		// Try API key authentication if enabled
		if endpoint.EnableAPIKeyAuth {
			if apiKey := extractAPIKey(c, endpoint); apiKey != "" {
				if validatedKey, err := authService.ValidateAPIKey(apiKey); err == nil {
					if u, err := authService.GetUserByID(validatedKey.UserID); err == nil && u.IsActive {
						authenticated = true
						// Set context for downstream handlers
						c.Set("user_id", u.ID)
						c.Set("organization_id", u.OrganizationID)
						c.Set("role", u.Role)
						c.Set("api_key", validatedKey)
					}
				}
			}
		}

		// Try OAuth/JWT authentication if enabled and not already authenticated
		if endpoint.EnableOAuth && !authenticated {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				// For now, we'll implement basic JWT validation
				// This could be extended to support OAuth providers
				token := strings.TrimPrefix(authHeader, "Bearer ")
				if len(token) > 0 {
					// TODO: Implement JWT validation here
					// For now, we'll just accept any Bearer token for OAuth-enabled endpoints
					authenticated = true
				}
			}
		}

		if !authenticated {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"details": "No valid authentication provided",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// extractAPIKey extracts API key from various sources based on endpoint configuration
func extractAPIKey(c *gin.Context, endpoint *types.Endpoint) string {
	// Try X-API-Key header first
	if apiKey := c.GetHeader("X-API-Key"); apiKey != "" {
		return apiKey
	}

	// Try Authorization header with Bearer prefix
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	// Try query parameter if enabled
	if endpoint.UseQueryParamAuth {
		if apiKey := c.Query("api_key"); apiKey != "" {
			return apiKey
		}
	}

	return ""
}

// EndpointRateLimitMiddleware applies rate limiting based on endpoint configuration
func EndpointRateLimitMiddleware() gin.HandlerFunc {
	// Create a map to store limiters per endpoint
	limiters := make(map[string]*limiter.Limiter)

	return func(c *gin.Context) {
		endpointVal, exists := c.Get("endpoint")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Endpoint not found in context"})
			c.Abort()
			return
		}

		endpoint := endpointVal.(*types.Endpoint)

		// Get or create limiter for this endpoint
		limiterKey := endpoint.ID
		lim, exists := limiters[limiterKey]
		if !exists {
			// Create rate limiter with endpoint-specific settings
			rate := limiter.Rate{
				Period: time.Duration(endpoint.RateLimitWindow) * time.Second,
				Limit:  int64(endpoint.RateLimitRequests),
			}
			store := memory.NewStore()
			lim = limiter.New(store, rate)
			limiters[limiterKey] = lim
		}

		// Apply rate limiting based on client IP
		key := fmt.Sprintf("endpoint:%s:%s", endpoint.Name, c.ClientIP())
		context, err := lim.Get(c.Request.Context(), key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limiting error"})
			c.Abort()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", context.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", context.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", context.Reset))

		// Check if rate limit exceeded
		if context.Reached {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": context.Reset,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// EndpointCORSMiddleware applies CORS settings based on endpoint configuration
func EndpointCORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		endpointVal, exists := c.Get("endpoint")
		if !exists {
			// If no endpoint in context, skip CORS handling
			c.Next()
			return
		}

		endpoint := endpointVal.(*types.Endpoint)

		// Get allowed origins
		origin := c.Request.Header.Get("Origin")
		allowed := false

		// Check if origin is allowed
		for _, allowedOrigin := range endpoint.AllowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			// Set CORS headers
			if len(endpoint.AllowedOrigins) > 0 && endpoint.AllowedOrigins[0] == "*" {
				c.Header("Access-Control-Allow-Origin", "*")
			} else {
				c.Header("Access-Control-Allow-Origin", origin)
			}

			c.Header("Access-Control-Allow-Methods", strings.Join(endpoint.AllowedMethods, ", "))
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, mcp-session-id")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400")
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
