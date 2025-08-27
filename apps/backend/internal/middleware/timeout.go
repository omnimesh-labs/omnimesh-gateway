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

		// Use a simple context check instead of goroutines to avoid races
		c.Next()

		// Check if context was cancelled during processing
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				// Request timed out - only write response if headers haven't been written yet
				if !c.Writer.Written() {
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
				}
				c.Abort()
			}
		default:
			// Request completed normally
		}
	}
}

// TimeoutConfig holds timeout middleware configuration
type TimeoutConfig struct {
	TimeoutHandler func(c *gin.Context)
	PanicHandler   func(c *gin.Context, err interface{})
	Timeout        time.Duration
}
