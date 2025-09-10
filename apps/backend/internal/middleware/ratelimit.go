package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	ginmiddleware "github.com/ulule/limiter/v3/drivers/middleware/gin"
	memorystore "github.com/ulule/limiter/v3/drivers/store/memory"
	redisstore "github.com/ulule/limiter/v3/drivers/store/redis"
)

// IPRateLimitConfig holds the configuration for IP rate limiting
type IPRateLimitConfig struct {
	RedisAddr      string   `yaml:"redis_addr"`
	RedisPassword  string   `yaml:"redis_password"`
	TrustedProxies []string `yaml:"trusted_proxies"`
	SkipPaths      []string `yaml:"skip_paths"`
	CustomHeaders  []string `yaml:"custom_headers"`
	RequestsPerMin int      `yaml:"requests_per_minute"`
	BurstSize      int      `yaml:"burst_size"`
	RedisDB        int      `yaml:"redis_db"`
	Enabled        bool     `yaml:"enabled"`
	RedisEnabled   bool     `yaml:"redis_enabled"`
}

// DefaultIPRateLimitConfig returns default configuration
func DefaultIPRateLimitConfig() *IPRateLimitConfig {
	return &IPRateLimitConfig{
		Enabled:        true,
		RequestsPerMin: 100,
		BurstSize:      20,
		RedisEnabled:   false,
		RedisAddr:      "localhost:6379",
		RedisDB:        0,
		TrustedProxies: []string{"127.0.0.1", "::1"},
		SkipPaths:      []string{"/health", "/metrics"},
		CustomHeaders:  []string{"X-Real-IP", "X-Forwarded-For"},
	}
}

// IPRateLimitFromAppConfig creates IP-based rate limiting middleware from app config
func IPRateLimitFromAppConfig(cfg *config.RateLimitConfig, redisCfg *config.RedisConfig) gin.HandlerFunc {
	if !cfg.IPEnabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Build IPRateLimitConfig from app config
	ipConfig := &IPRateLimitConfig{
		Enabled:        cfg.IPEnabled,
		RequestsPerMin: cfg.IPRequestsPerMinute,
		SkipPaths:      cfg.IPSkipPaths,
		CustomHeaders:  cfg.IPCustomHeaders,
	}

	// Determine storage backend
	if cfg.Storage == "redis" && redisCfg != nil {
		ipConfig.RedisEnabled = true
		ipConfig.RedisAddr = fmt.Sprintf("%s:%d", redisCfg.Host, redisCfg.Port)
		ipConfig.RedisPassword = redisCfg.Password
		ipConfig.RedisDB = redisCfg.Database
	} else {
		ipConfig.RedisEnabled = false
	}

	return IPRateLimit(ipConfig)
}

// IPRateLimit creates IP-based rate limiting middleware
func IPRateLimit(config *IPRateLimitConfig) gin.HandlerFunc {
	if !config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Create the appropriate store
	var store limiter.Store
	var err error

	if config.RedisEnabled {
		// Use Redis store for distributed rate limiting
		redisClient := redis.NewClient(&redis.Options{
			Addr:     config.RedisAddr,
			Password: config.RedisPassword,
			DB:       config.RedisDB,
		})

		store, err = redisstore.NewStore(redisClient)
		if err != nil {
			// Fallback to memory store if Redis fails
			store = memorystore.NewStore()
		}
	} else {
		// Use memory store for single-instance rate limiting
		store = memorystore.NewStore()
	}

	// Create rate limiter with sliding window
	rateStr := fmt.Sprintf("%d-M", config.RequestsPerMin) // e.g., "100-M" for 100 requests per minute
	rate, err := limiter.NewRateFromFormatted(rateStr)
	if err != nil {
		// Fallback to default rate
		rate = limiter.Rate{
			Period: time.Minute,
			Limit:  int64(config.RequestsPerMin),
		}
	}

	limiterInstance := limiter.New(store, rate)

	// Create middleware with custom key getter
	mw := ginmiddleware.NewMiddleware(limiterInstance, ginmiddleware.WithKeyGetter(func(c *gin.Context) string {
		return getClientIP(c, config)
	}))

	return func(c *gin.Context) {
		// Skip rate limiting for certain paths
		if shouldSkipPath(c.Request.URL.Path, config.SkipPaths) {
			c.Next()
			return
		}

		// Apply rate limiting
		mw(c)
	}
}

// getClientIP extracts the real client IP considering proxies and custom headers
func getClientIP(c *gin.Context, config *IPRateLimitConfig) string {
	// First, try custom headers in order
	for _, header := range config.CustomHeaders {
		if ip := c.GetHeader(header); ip != "" {
			// Take first IP if comma-separated list
			if len(ip) > 0 {
				return extractFirstIP(ip)
			}
		}
	}

	// Fallback to Gin's ClientIP which handles X-Forwarded-For, X-Real-IP, etc.
	return c.ClientIP()
}

// extractFirstIP extracts the first IP from a comma-separated list
func extractFirstIP(ipList string) string {
	for i, char := range ipList {
		if char == ',' {
			return ipList[:i]
		}
	}
	return ipList
}

// shouldSkipPath checks if the path should skip rate limiting
func shouldSkipPath(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// IPRateLimitWithRedis creates IP rate limiting middleware with Redis backend
func IPRateLimitWithRedis(requestsPerMin int, redisAddr, redisPassword string, redisDB int) gin.HandlerFunc {
	config := &IPRateLimitConfig{
		Enabled:        true,
		RequestsPerMin: requestsPerMin,
		RedisEnabled:   true,
		RedisAddr:      redisAddr,
		RedisPassword:  redisPassword,
		RedisDB:        redisDB,
		TrustedProxies: []string{"127.0.0.1", "::1"},
		SkipPaths:      []string{"/health", "/metrics"},
		CustomHeaders:  []string{"X-Real-IP", "X-Forwarded-For"},
	}
	return IPRateLimit(config)
}

// IPRateLimitWithMemory creates IP rate limiting middleware with memory backend
func IPRateLimitWithMemory(requestsPerMin int) gin.HandlerFunc {
	config := &IPRateLimitConfig{
		Enabled:        true,
		RequestsPerMin: requestsPerMin,
		RedisEnabled:   false,
		TrustedProxies: []string{"127.0.0.1", "::1"},
		SkipPaths:      []string{"/health", "/metrics"},
		CustomHeaders:  []string{"X-Real-IP", "X-Forwarded-For"},
	}
	return IPRateLimit(config)
}

// CustomRateLimitResponse provides a custom response for rate limit exceeded
func CustomRateLimitResponse() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Writer.Status() == http.StatusTooManyRequests {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests from this IP address. Please try again later.",
				"code":    "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
