package handlers

import (
	"log"
	"net/http"

	"mcp-gateway/apps/backend/internal/types"

	"github.com/gin-gonic/gin"
)

// RespondWithError logs the actual error and returns a sanitized error response
func RespondWithError(c *gin.Context, err error) {
	// Determine the type of error and respond appropriately
	if typedErr, ok := err.(*types.Error); ok {
		// Use the structured error directly
		c.JSON(typedErr.Status, types.ErrorResponse{
			Error:   typedErr,
			Success: false,
		})
		return
	}

	// For untyped errors, log them and return a generic message
	log.Printf("[ERROR] Internal error at %s %s: %v", c.Request.Method, c.Request.URL.Path, err)

	c.JSON(http.StatusInternalServerError, types.ErrorResponse{
		Error:   types.NewInternalError("An internal error occurred. Please try again later."),
		Success: false,
	})
}

// RespondWithValidationError returns a validation error response
func RespondWithValidationError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, types.ErrorResponse{
		Error:   types.NewValidationError(message),
		Success: false,
	})
}

// RespondWithNotFound returns a not found error response
func RespondWithNotFound(c *gin.Context, resource string) {
	c.JSON(http.StatusNotFound, types.ErrorResponse{
		Error:   types.NewNotFoundError(resource + " not found"),
		Success: false,
	})
}

// RespondWithUnauthorized returns an unauthorized error response
func RespondWithUnauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "Unauthorized access"
	}
	c.JSON(http.StatusUnauthorized, types.ErrorResponse{
		Error:   types.NewUnauthorizedError(message),
		Success: false,
	})
}

// RespondWithForbidden returns a forbidden error response
func RespondWithForbidden(c *gin.Context) {
	c.JSON(http.StatusForbidden, types.ErrorResponse{
		Error:   types.NewInsufficientRightsError(),
		Success: false,
	})
}

// RespondWithConflict returns a conflict error response
func RespondWithConflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, types.ErrorResponse{
		Error:   types.NewConflictError(message),
		Success: false,
	})
}

// RespondWithServiceUnavailable returns a service unavailable error response
func RespondWithServiceUnavailable(c *gin.Context, message string) {
	if message == "" {
		message = "Service temporarily unavailable"
	}
	c.JSON(http.StatusServiceUnavailable, types.ErrorResponse{
		Error:   types.NewServiceUnavailableError(message),
		Success: false,
	})
}

// RespondWithSuccess returns a successful response
func RespondWithSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// RespondWithCreated returns a created response
func RespondWithCreated(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    data,
	})
}
