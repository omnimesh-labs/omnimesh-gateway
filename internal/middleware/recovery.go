package middleware

import (
	"io"
	"net/http"
	"runtime/debug"

	"mcp-gateway/internal/types"

	"github.com/gin-gonic/gin"
)

// Recovery returns a middleware that recovers from panics
func Recovery() gin.HandlerFunc {
	return RecoveryWithWriter(gin.DefaultWriter)
}

// RecoveryWithWriter returns a middleware that recovers from panics and writes to writer
func RecoveryWithWriter(writer io.Writer) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic

				// TODO: Use proper logging service
				writer.Write([]byte("Panic recovered: " + string(debug.Stack())))

				// Create error response
				errorResp := &types.ErrorResponse{
					Error:   types.NewInternalError("Internal server error"),
					Success: false,
				}

				c.JSON(http.StatusInternalServerError, errorResp)
				c.Abort()
			}
		}()

		c.Next()
	}
}

// RecoveryWithConfig returns a middleware that recovers from panics with custom configuration
func RecoveryWithConfig(config *RecoveryConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic if handler is provided
				if config.LogHandler != nil {
					config.LogHandler(c, err, debug.Stack())
				}

				// Call custom recovery handler if provided
				if config.RecoveryHandler != nil {
					config.RecoveryHandler(c, err)
					return
				}

				// Default error response
				errorResp := &types.ErrorResponse{
					Error:   types.NewInternalError("Internal server error"),
					Success: false,
				}

				c.JSON(http.StatusInternalServerError, errorResp)
				c.Abort()
			}
		}()

		c.Next()
	}
}

// RecoveryConfig holds recovery middleware configuration
type RecoveryConfig struct {
	// LogHandler handles panic logging
	LogHandler func(c *gin.Context, err interface{}, stack []byte)

	// RecoveryHandler handles the recovery process
	RecoveryHandler func(c *gin.Context, err interface{})
}
