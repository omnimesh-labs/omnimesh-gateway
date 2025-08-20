package ratelimit

import (
	"net/http"
	"strconv"
	"time"

	"mcp-gateway/apps/backend/internal/types"

	"github.com/gin-gonic/gin"
)

// Middleware provides rate limiting middleware
type Middleware struct {
	service *Service
}

// NewMiddleware creates a new rate limiting middleware
func NewMiddleware(service *Service) *Middleware {
	return &Middleware{
		service: service,
	}
}

// RateLimit applies rate limiting to requests
func (m *Middleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract rate limiting context
		ctx := m.extractRateLimitContext(c)

		// Check rate limits
		allowed, usage, err := m.service.CheckRateLimit(ctx)
		if err != nil {
			// Log error but don't block request
			c.Next()
			return
		}

		// Add rate limit headers
		m.addRateLimitHeaders(c, usage)

		if !allowed {
			// Rate limit exceeded
			errorResp := &types.ErrorResponse{
				Error:   types.NewRateLimitExceededError("Rate limit exceeded"),
				Success: false,
			}
			c.JSON(http.StatusTooManyRequests, errorResp)
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserRateLimit applies per-user rate limiting
func (m *Middleware) UserRateLimit(limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		key := "user:" + userID.(string)
		allowed, usage, err := m.service.CheckLimit(key, limit, window)
		if err != nil {
			c.Next()
			return
		}

		m.addRateLimitHeaders(c, usage)

		if !allowed {
			errorResp := &types.ErrorResponse{
				Error:   types.NewRateLimitExceededError("User rate limit exceeded"),
				Success: false,
			}
			c.JSON(http.StatusTooManyRequests, errorResp)
			c.Abort()
			return
		}

		c.Next()
	}
}

// OrganizationRateLimit applies per-organization rate limiting
func (m *Middleware) OrganizationRateLimit(limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID, exists := c.Get("organization_id")
		if !exists {
			c.Next()
			return
		}

		key := "org:" + orgID.(string)
		allowed, usage, err := m.service.CheckLimit(key, limit, window)
		if err != nil {
			c.Next()
			return
		}

		m.addRateLimitHeaders(c, usage)

		if !allowed {
			errorResp := &types.ErrorResponse{
				Error:   types.NewRateLimitExceededError("Organization rate limit exceeded"),
				Success: false,
			}
			c.JSON(http.StatusTooManyRequests, errorResp)
			c.Abort()
			return
		}

		c.Next()
	}
}

// EndpointRateLimit applies per-endpoint rate limiting
func (m *Middleware) EndpointRateLimit(limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "endpoint:" + c.Request.Method + ":" + c.FullPath()
		allowed, usage, err := m.service.CheckLimit(key, limit, window)
		if err != nil {
			c.Next()
			return
		}

		m.addRateLimitHeaders(c, usage)

		if !allowed {
			errorResp := &types.ErrorResponse{
				Error:   types.NewRateLimitExceededError("Endpoint rate limit exceeded"),
				Success: false,
			}
			c.JSON(http.StatusTooManyRequests, errorResp)
			c.Abort()
			return
		}

		c.Next()
	}
}

// IPRateLimit applies per-IP rate limiting
func (m *Middleware) IPRateLimit(limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "ip:" + c.ClientIP()
		allowed, usage, err := m.service.CheckLimit(key, limit, window)
		if err != nil {
			c.Next()
			return
		}

		m.addRateLimitHeaders(c, usage)

		if !allowed {
			errorResp := &types.ErrorResponse{
				Error:   types.NewRateLimitExceededError("IP rate limit exceeded"),
				Success: false,
			}
			c.JSON(http.StatusTooManyRequests, errorResp)
			c.Abort()
			return
		}

		c.Next()
	}
}

// extractRateLimitContext extracts rate limiting context from request
func (m *Middleware) extractRateLimitContext(c *gin.Context) *RateLimitContext {
	ctx := &RateLimitContext{
		Method:    c.Request.Method,
		Path:      c.FullPath(),
		RemoteIP:  c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}

	if userID, exists := c.Get("user_id"); exists {
		ctx.UserID = userID.(string)
	}

	if orgID, exists := c.Get("organization_id"); exists {
		ctx.OrganizationID = orgID.(string)
	}

	if role, exists := c.Get("role"); exists {
		ctx.Role = role.(string)
	}

	return ctx
}

// addRateLimitHeaders adds rate limit information to response headers
func (m *Middleware) addRateLimitHeaders(c *gin.Context, usage *Usage) {
	if usage == nil {
		return
	}

	c.Header("X-RateLimit-Limit", strconv.Itoa(usage.Limit))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(usage.Remaining))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(usage.ResetTime.Unix(), 10))
}

// RateLimitContext contains context for rate limiting decisions
type RateLimitContext struct {
	UserID         string
	OrganizationID string
	Role           string
	Method         string
	Path           string
	RemoteIP       string
	UserAgent      string
}
