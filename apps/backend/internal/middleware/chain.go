package middleware

import (
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/config"

	"github.com/gin-gonic/gin"
)

// Chain represents a middleware chain
type Chain struct {
	middlewares []gin.HandlerFunc
}

// NewChain creates a new middleware chain
func NewChain() *Chain {
	return &Chain{
		middlewares: make([]gin.HandlerFunc, 0),
	}
}

// Use adds middleware to the chain
func (c *Chain) Use(middleware gin.HandlerFunc) *Chain {
	c.middlewares = append(c.middlewares, middleware)
	return c
}

// Build returns the middleware chain as a slice
func (c *Chain) Build() []gin.HandlerFunc {
	return c.middlewares
}

// Apply applies the middleware chain to a gin router group
func (c *Chain) Apply(group *gin.RouterGroup) {
	for _, middleware := range c.middlewares {
		group.Use(middleware)
	}
}

// DefaultChain creates a default middleware chain for the gateway
func DefaultChain() *Chain {
	return NewChain().
		Use(Recovery()).
		Use(SecurityHeaders()).
		Use(IPRateLimitWithMemory(100)). // 100 requests per minute by default
		Use(Timeout())
}

// DefaultChainFromConfig creates a default middleware chain from app config
func DefaultChainFromConfig(cfg *config.Config) *Chain {
	return NewChain().
		Use(Recovery()).
		Use(SecurityHeaders()).
		Use(IPRateLimitFromAppConfig(&cfg.RateLimit, &cfg.Redis)).
		Use(Timeout())
}

// DefaultChainWithConfig creates a default middleware chain with custom security config
func DefaultChainWithConfig(securityConfig *SecurityConfig) *Chain {
	return NewChain().
		Use(Recovery()).
		Use(SecurityHeadersWithConfig(securityConfig)).
		Use(IPRateLimitWithMemory(100)). // 100 requests per minute by default
		Use(Timeout())
}

// DefaultChainWithAppConfig creates a default middleware chain with app and security config
func DefaultChainWithAppConfig(cfg *config.Config, securityConfig *SecurityConfig) *Chain {
	return NewChain().
		Use(Recovery()).
		Use(SecurityHeadersWithConfig(securityConfig)).
		Use(IPRateLimitFromAppConfig(&cfg.RateLimit, &cfg.Redis)).
		Use(Timeout())
}

// DefaultChainWithRateLimit creates a default middleware chain with Redis rate limiting
func DefaultChainWithRateLimit(requestsPerMin int, redisAddr, redisPassword string, redisDB int) *Chain {
	return NewChain().
		Use(Recovery()).
		Use(SecurityHeaders()).
		Use(IPRateLimitWithRedis(requestsPerMin, redisAddr, redisPassword, redisDB)).
		Use(Timeout())
}

// AuthenticatedChain creates a middleware chain for authenticated routes
func AuthenticatedChain() *Chain {
	return DefaultChain()
	// Auth middleware will be added by the auth service
}

// AdminChain creates a middleware chain for admin routes
func AdminChain() *Chain {
	return AuthenticatedChain()
	// Role-based middleware will be added by the auth service
}
