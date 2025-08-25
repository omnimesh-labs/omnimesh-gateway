package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"mcp-gateway/apps/backend/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("RespondWithError sanitizes internal errors", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		// Test with a raw error (should be sanitized)
		rawErr := errors.New("database connection failed: password=secret123 host=internal.db")
		RespondWithError(c, rawErr)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response types.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.NotNil(t, response.Error)

		// Verify the error message is sanitized
		assert.Equal(t, "An internal error occurred. Please try again later.", response.Error.Message)
		assert.Equal(t, types.ErrCodeInternalError, response.Error.Code)

		// Ensure no sensitive information is exposed
		assert.NotContains(t, w.Body.String(), "password")
		assert.NotContains(t, w.Body.String(), "secret123")
		assert.NotContains(t, w.Body.String(), "internal.db")
	})

	t.Run("RespondWithError preserves structured errors", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		// Test with a structured error
		structuredErr := types.NewNotFoundError("User not found")
		RespondWithError(c, structuredErr)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response types.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.NotNil(t, response.Error)
		assert.Equal(t, "User not found", response.Error.Message)
		assert.Equal(t, types.ErrCodeNotFound, response.Error.Code)
	})

	t.Run("RespondWithValidationError returns proper error", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		RespondWithValidationError(c, "Email is required")

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response types.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.NotNil(t, response.Error)
		assert.Equal(t, "Email is required", response.Error.Message)
		assert.Equal(t, types.ErrCodeValidationFailed, response.Error.Code)
	})

	t.Run("RespondWithNotFound returns proper error", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		RespondWithNotFound(c, "Resource")

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response types.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.NotNil(t, response.Error)
		assert.Equal(t, "Resource not found", response.Error.Message)
		assert.Equal(t, types.ErrCodeNotFound, response.Error.Code)
	})

	t.Run("RespondWithUnauthorized returns proper error", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		RespondWithUnauthorized(c, "Invalid token")

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response types.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.NotNil(t, response.Error)
		assert.Equal(t, "Invalid token", response.Error.Message)
		assert.Equal(t, types.ErrCodeUnauthorized, response.Error.Code)
	})

	t.Run("RespondWithServiceUnavailable returns proper error", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		RespondWithServiceUnavailable(c, "")

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var response types.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.NotNil(t, response.Error)
		assert.Equal(t, "Service temporarily unavailable", response.Error.Message)
		assert.Equal(t, types.ErrCodeServiceUnavailable, response.Error.Code)
	})
}
