package middleware

import (
	"context"
	"net/http"
	"time"

	"mcp-gateway/apps/backend/internal/types"

	"github.com/gin-gonic/gin"
)

// Timeout returns a middleware that times out requests
func Timeout() gin.HandlerFunc {
	return TimeoutWithConfig(&TimeoutConfig{
		Timeout: 30 * time.Second,
	})
}

// TimeoutWithConfig returns a middleware that times out requests with custom configuration
func TimeoutWithConfig(config *TimeoutConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), config.Timeout)
		defer cancel()

		// Replace request context
		c.Request = c.Request.WithContext(ctx)

		// Create channel to signal completion
		finished := make(chan struct{})

		// Run the request in a goroutine
		go func() {
			defer func() {
				if err := recover(); err != nil {
					// Handle panic in goroutine
					if config.PanicHandler != nil {
						config.PanicHandler(c, err)
					}
				}
			}()

			c.Next()
			close(finished)
		}()

		// Wait for completion or timeout
		select {
		case <-finished:
			// Request completed normally
			return
		case <-ctx.Done():
			// Request timed out
			if config.TimeoutHandler != nil {
				config.TimeoutHandler(c)
			} else {
				// Default timeout response
				errorResp := &types.ErrorResponse{
					Error:   types.NewTimeoutError("Request timeout"),
					Success: false,
				}
				c.JSON(http.StatusGatewayTimeout, errorResp)
			}
			c.Abort()
			return
		}
	}
}

// TimeoutConfig holds timeout middleware configuration
type TimeoutConfig struct {
	// Timeout duration
	Timeout time.Duration

	// TimeoutHandler handles timeout responses
	TimeoutHandler func(c *gin.Context)

	// PanicHandler handles panics in timeout goroutine
	PanicHandler func(c *gin.Context, err interface{})
}
