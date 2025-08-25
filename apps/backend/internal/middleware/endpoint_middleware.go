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
	ValidateAccess(ctx context.Context, endpoint *types.Endpoint, req *http.Request) error
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

// EndpointAuthMiddleware validates access to endpoint based on its auth settings
func EndpointAuthMiddleware(endpointService EndpointService) gin.HandlerFunc {
	return func(c *gin.Context) {
		endpointVal, exists := c.Get("endpoint")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Endpoint not found in context"})
			c.Abort()
			return
		}

		endpoint := endpointVal.(*types.Endpoint)

		// Check authentication based on endpoint settings
		if err := endpointService.ValidateAccess(c.Request.Context(), endpoint, c.Request); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
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
				"error": "Rate limit exceeded",
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
