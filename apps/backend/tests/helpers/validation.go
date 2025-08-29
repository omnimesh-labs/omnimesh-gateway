package helpers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// APIResponse represents the standard API response structure
type APIResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   APIError               `json:"error,omitempty"`
}

// APIError represents the standard API error structure
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ValidateSuccessResponse validates that a response follows the success format
func ValidateSuccessResponse(t *testing.T, response *httptest.ResponseRecorder, expectedStatus int) map[string]interface{} {
	assert.Equal(t, expectedStatus, response.Code, "Response status should match expected")

	var apiResponse APIResponse
	err := json.Unmarshal(response.Body.Bytes(), &apiResponse)
	require.NoError(t, err, "Response should be valid JSON")

	assert.True(t, apiResponse.Success, "Success response should have success: true")
	assert.NotNil(t, apiResponse.Data, "Success response should contain data field")
	assert.Empty(t, apiResponse.Error.Code, "Success response should not contain error field")

	return apiResponse.Data
}

// ValidateErrorResponse validates that a response follows the error format
func ValidateErrorResponse(t *testing.T, response *httptest.ResponseRecorder, expectedStatus int) APIError {
	assert.Equal(t, expectedStatus, response.Code, "Response status should match expected")

	var apiResponse APIResponse
	err := json.Unmarshal(response.Body.Bytes(), &apiResponse)
	require.NoError(t, err, "Response should be valid JSON")

	assert.False(t, apiResponse.Success, "Error response should have success: false")
	assert.Nil(t, apiResponse.Data, "Error response should not contain data field")
	assert.NotEmpty(t, apiResponse.Error.Code, "Error response should contain error code")
	assert.NotEmpty(t, apiResponse.Error.Message, "Error response should contain error message")

	return apiResponse.Error
}

// ValidateLoginResponse validates a login success response structure
func ValidateLoginResponse(t *testing.T, data map[string]interface{}) {
	// Check required fields
	assert.Contains(t, data, "access_token", "Login response should contain access_token")
	assert.Contains(t, data, "refresh_token", "Login response should contain refresh_token")
	assert.Contains(t, data, "user", "Login response should contain user")
	assert.Contains(t, data, "token_type", "Login response should contain token_type")
	assert.Contains(t, data, "expires_in", "Login response should contain expires_in")

	// Validate token values are not empty
	assert.NotEmpty(t, data["access_token"], "Access token should not be empty")
	assert.NotEmpty(t, data["refresh_token"], "Refresh token should not be empty")
	assert.Equal(t, "Bearer", data["token_type"], "Token type should be Bearer")

	// Validate user object
	user, ok := data["user"].(map[string]interface{})
	require.True(t, ok, "User field should be an object")
	assert.Contains(t, user, "email", "User should contain email")
	assert.Contains(t, user, "id", "User should contain id")
	assert.NotEmpty(t, user["email"], "User email should not be empty")
}

// ValidateRefreshTokenResponse validates a refresh token response structure
func ValidateRefreshTokenResponse(t *testing.T, data map[string]interface{}) {
	// Check required fields
	assert.Contains(t, data, "access_token", "Refresh response should contain access_token")
	assert.Contains(t, data, "refresh_token", "Refresh response should contain refresh_token")
	assert.Contains(t, data, "user", "Refresh response should contain user")

	// Validate token values are not empty
	assert.NotEmpty(t, data["access_token"], "Access token should not be empty")
	assert.NotEmpty(t, data["refresh_token"], "Refresh token should not be empty")

	// Validate user object
	user, ok := data["user"].(map[string]interface{})
	require.True(t, ok, "User field should be an object")
	assert.Contains(t, user, "email", "User should contain email")
	assert.Contains(t, user, "id", "User should contain id")
}

// ValidateUserResponse validates a user response structure
func ValidateUserResponse(t *testing.T, data map[string]interface{}) {
	// User object validation
	assert.Contains(t, data, "email", "User should contain email")
	assert.Contains(t, data, "id", "User should contain id")
	assert.Contains(t, data, "name", "User should contain name")
	assert.Contains(t, data, "organization_id", "User should contain organization_id")
	assert.Contains(t, data, "role", "User should contain role")
	assert.NotEmpty(t, data["email"], "User email should not be empty")
	assert.NotEmpty(t, data["id"], "User id should not be empty")
}

// ValidateAPIKeyResponse validates an API key response structure
func ValidateAPIKeyResponse(t *testing.T, data map[string]interface{}) {
	// Check required fields
	assert.Contains(t, data, "api_key", "API key response should contain api_key")
	assert.Contains(t, data, "key", "API key response should contain key (actual key value)")

	// Validate api_key object
	apiKey, ok := data["api_key"].(map[string]interface{})
	require.True(t, ok, "api_key field should be an object")
	assert.Contains(t, apiKey, "id", "API key should contain id")
	assert.Contains(t, apiKey, "name", "API key should contain name")

	// Validate actual key value
	keyValue, ok := data["key"].(string)
	require.True(t, ok, "key field should be a string")
	assert.NotEmpty(t, keyValue, "Key value should not be empty")
}

// ValidateAuthenticationError validates specific authentication error types
func ValidateAuthenticationError(t *testing.T, apiError APIError, expectedCode string) {
	assert.Equal(t, expectedCode, apiError.Code, "Error code should match expected")
	assert.NotEmpty(t, apiError.Message, "Error message should not be empty")
}

// ValidateResponseConsistency checks that all responses follow the same structure pattern
func ValidateResponseConsistency(t *testing.T, response *httptest.ResponseRecorder, expectSuccess bool) {
	var raw map[string]interface{}
	err := json.Unmarshal(response.Body.Bytes(), &raw)
	require.NoError(t, err, "Response should be valid JSON")

	// All responses should have success field
	assert.Contains(t, raw, "success", "All responses should contain success field")
	success, ok := raw["success"].(bool)
	require.True(t, ok, "Success field should be a boolean")

	if expectSuccess {
		assert.True(t, success, "Success responses should have success: true")
		assert.Contains(t, raw, "data", "Success responses should contain data field")
		assert.NotContains(t, raw, "error", "Success responses should not contain error field")
	} else {
		assert.False(t, success, "Error responses should have success: false")
		assert.Contains(t, raw, "error", "Error responses should contain error field")
		assert.NotContains(t, raw, "data", "Error responses should not contain data field")
	}
}
