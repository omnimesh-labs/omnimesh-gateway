package unit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/a2a"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestA2AClient_Chat_OpenAI(t *testing.T) {
	// Mock OpenAI API response
	mockResponse := map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "Hello! How can I help you today?",
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     10,
			"completion_tokens": 15,
			"total_tokens":      25,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer test-api-key")

		// Verify request method
		assert.Equal(t, "POST", r.Method)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create agent
	agent := &types.A2AAgent{
		ID:          uuid.New(),
		Name:        "OpenAI Test Agent",
		EndpointURL: server.URL,
		AgentType:   types.AgentTypeOpenAI,
		AuthType:    types.AuthTypeAPIKey,
		AuthValue:   "test-api-key",
		ConfigData: map[string]interface{}{
			"model":       "gpt-4",
			"max_tokens":  1000,
			"temperature": 0.7,
		},
		CapabilitiesData: map[string]interface{}{
			types.CapabilityChat: true,
		},
		IsActive: true,
	}

	// Create client and make request
	client := a2a.NewClient(30*time.Second, 3)
	request := &types.A2AChatRequest{
		Messages: []types.A2AChatMessage{
			{
				Role:    "user",
				Content: "Hello!",
			},
		},
		MaxTokens:   1000,
		Temperature: 0.7,
	}

	response, err := client.Chat(agent, request)
	require.NoError(t, err)
	require.NotNil(t, response)

	assert.Equal(t, "assistant", response.Message.Role)
	assert.Equal(t, "Hello! How can I help you today?", response.Message.Content)
	assert.Equal(t, "stop", response.FinishReason)
	assert.Equal(t, 10, response.Usage.InputTokens)
	assert.Equal(t, 15, response.Usage.OutputTokens)
	assert.Equal(t, 25, response.Usage.TotalTokens)
}

func TestA2AClient_Chat_Anthropic(t *testing.T) {
	// Mock Anthropic API response
	mockResponse := map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": "Hello! I'm Claude, nice to meet you!",
			},
		},
		"stop_reason": "end_turn",
		"usage": map[string]interface{}{
			"input_tokens":  8,
			"output_tokens": 12,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-api-key", r.Header.Get("x-api-key"))
		assert.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create agent
	agent := &types.A2AAgent{
		ID:          uuid.New(),
		Name:        "Anthropic Test Agent",
		EndpointURL: server.URL,
		AgentType:   types.AgentTypeAnthropic,
		AuthType:    types.AuthTypeAPIKey,
		AuthValue:   "test-api-key",
		ConfigData: map[string]interface{}{
			"model":       "claude-3-sonnet-20240229",
			"max_tokens":  4000,
			"temperature": 0.7,
		},
		CapabilitiesData: map[string]interface{}{
			types.CapabilityChat: true,
		},
		IsActive: true,
	}

	// Create client and make request
	client := a2a.NewClient(30*time.Second, 3)
	request := &types.A2AChatRequest{
		Messages: []types.A2AChatMessage{
			{
				Role:    "user",
				Content: "Hello!",
			},
		},
	}

	response, err := client.Chat(agent, request)
	require.NoError(t, err)
	require.NotNil(t, response)

	assert.Equal(t, "assistant", response.Message.Role)
	assert.Equal(t, "Hello! I'm Claude, nice to meet you!", response.Message.Content)
	assert.Equal(t, "end_turn", response.FinishReason)
	assert.Equal(t, 8, response.Usage.InputTokens)
	assert.Equal(t, 12, response.Usage.OutputTokens)
	assert.Equal(t, 20, response.Usage.TotalTokens)
}

func TestA2AClient_Chat_CustomAgent(t *testing.T) {
	// Mock custom agent API response
	mockResponse := types.A2AResponse{
		Success: true,
		Data: map[string]interface{}{
			"message": "Custom agent response",
		},
		ProtocolVersion: "1.0",
		Usage: &types.A2AUsage{
			TotalTokens: 50,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer custom-token", r.Header.Get("Authorization"))

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create agent
	agent := &types.A2AAgent{
		ID:              uuid.New(),
		Name:            "Custom Test Agent",
		EndpointURL:     server.URL,
		AgentType:       types.AgentTypeCustom,
		AuthType:        types.AuthTypeBearer,
		AuthValue:       "custom-token",
		ProtocolVersion: "1.0",
		CapabilitiesData: map[string]interface{}{
			types.CapabilityChat: true,
		},
		IsActive: true,
	}

	// Create client and make request
	client := a2a.NewClient(30*time.Second, 3)
	request := &types.A2AChatRequest{
		Messages: []types.A2AChatMessage{
			{
				Role:    "user",
				Content: "Hello custom agent!",
			},
		},
	}

	response, err := client.Chat(agent, request)
	require.NoError(t, err)
	require.NotNil(t, response)

	assert.Equal(t, "assistant", response.Message.Role)
	assert.Equal(t, "Custom agent response", response.Message.Content)
	assert.Equal(t, 50, response.Usage.TotalTokens)
}

func TestA2AClient_Invoke(t *testing.T) {
	// Mock generic invocation response
	mockResponse := types.A2AResponse{
		Success:         true,
		Data:            "Generic invocation successful",
		ProtocolVersion: "1.0",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var req types.A2ARequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Verify request structure
		assert.Equal(t, types.InteractionTypeQuery, req.InteractionType)
		assert.Equal(t, "1.0", req.ProtocolVersion)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create agent
	agent := &types.A2AAgent{
		ID:               uuid.New(),
		Name:             "Generic Test Agent",
		EndpointURL:      server.URL,
		AgentType:        types.AgentTypeGeneric,
		AuthType:         types.AuthTypeNone,
		ProtocolVersion:  "1.0",
		CapabilitiesData: map[string]interface{}{},
		IsActive:         true,
	}

	// Create client and make request
	client := a2a.NewClient(30*time.Second, 3)
	request := &types.A2ARequest{
		InteractionType: types.InteractionTypeQuery,
		Parameters: map[string]interface{}{
			"action": "test",
		},
		ProtocolVersion: "1.0",
	}

	response, err := client.Invoke(agent, request)
	require.NoError(t, err)
	require.NotNil(t, response)

	assert.True(t, response.Success)
	assert.Equal(t, "Generic invocation successful", response.Data)
	assert.Equal(t, "1.0", response.ProtocolVersion)
}

func TestA2AClient_HealthCheck(t *testing.T) {
	// Test successful health check
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	agent := &types.A2AAgent{
		ID:          uuid.New(),
		Name:        "Health Check Test Agent",
		EndpointURL: server.URL,
		AgentType:   types.AgentTypeGeneric,
		AuthType:    types.AuthTypeNone,
		IsActive:    true,
	}

	client := a2a.NewClient(10*time.Second, 1)
	healthCheck, err := client.HealthCheck(agent)
	require.NoError(t, err)
	require.NotNil(t, healthCheck)

	assert.Equal(t, agent.ID, healthCheck.AgentID)
	assert.Equal(t, "healthy", healthCheck.Status)
	assert.Equal(t, "Agent is responding", healthCheck.Message)
	assert.Greater(t, healthCheck.ResponseTime, 0)
	assert.False(t, healthCheck.Timestamp.IsZero())
}

func TestA2AClient_HealthCheck_Failure(t *testing.T) {
	// Test failed health check (server not responding)
	agent := &types.A2AAgent{
		ID:          uuid.New(),
		Name:        "Failed Health Check Agent",
		EndpointURL: "http://localhost:99999/nonexistent", // Invalid endpoint
		AgentType:   types.AgentTypeGeneric,
		AuthType:    types.AuthTypeNone,
		IsActive:    true,
	}

	client := a2a.NewClient(1*time.Second, 1) // Short timeout for quick test
	healthCheck, err := client.HealthCheck(agent)
	require.NoError(t, err) // HealthCheck doesn't return error, it captures it in the result
	require.NotNil(t, healthCheck)

	assert.Equal(t, agent.ID, healthCheck.AgentID)
	assert.Equal(t, "unhealthy", healthCheck.Status)
	assert.NotEmpty(t, healthCheck.Message)
	assert.Greater(t, healthCheck.ResponseTime, 0)
}

func TestA2AClient_AuthenticationTypes(t *testing.T) {
	tests := []struct {
		expectedAuth func(*http.Request) bool
		name         string
		agentType    types.AgentType
		authType     types.AuthType
		authValue    string
	}{
		{
			name:      "OpenAI Bearer Token",
			agentType: types.AgentTypeOpenAI,
			authType:  types.AuthTypeAPIKey,
			authValue: "sk-test123",
			expectedAuth: func(r *http.Request) bool {
				return r.Header.Get("Authorization") == "Bearer sk-test123"
			},
		},
		{
			name:      "Anthropic API Key",
			agentType: types.AgentTypeAnthropic,
			authType:  types.AuthTypeAPIKey,
			authValue: "ant-test123",
			expectedAuth: func(r *http.Request) bool {
				return r.Header.Get("x-api-key") == "ant-test123" &&
					r.Header.Get("anthropic-version") == "2023-06-01"
			},
		},
		{
			name:      "Custom Bearer Token",
			agentType: types.AgentTypeCustom,
			authType:  types.AuthTypeBearer,
			authValue: "bearer-token-123",
			expectedAuth: func(r *http.Request) bool {
				return r.Header.Get("Authorization") == "Bearer bearer-token-123"
			},
		},
		{
			name:      "Generic API Key",
			agentType: types.AgentTypeGeneric,
			authType:  types.AuthTypeAPIKey,
			authValue: "generic-key-123",
			expectedAuth: func(r *http.Request) bool {
				return r.Header.Get("X-API-Key") == "generic-key-123"
			},
		},
		{
			name:      "No Authentication",
			agentType: types.AgentTypeGeneric,
			authType:  types.AuthTypeNone,
			authValue: "",
			expectedAuth: func(r *http.Request) bool {
				return r.Header.Get("Authorization") == "" && r.Header.Get("X-API-Key") == ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify authentication headers
				assert.True(t, tt.expectedAuth(r), "Authentication headers not as expected")

				// Return minimal success response
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status": "ok"}`))
			}))
			defer server.Close()

			agent := &types.A2AAgent{
				ID:          uuid.New(),
				Name:        "Auth Test Agent",
				EndpointURL: server.URL,
				AgentType:   tt.agentType,
				AuthType:    tt.authType,
				AuthValue:   tt.authValue,
				IsActive:    true,
			}

			client := a2a.NewClient(5*time.Second, 1)
			request := &types.A2ARequest{
				InteractionType: types.InteractionTypeHealth,
				Parameters:      map[string]interface{}{},
				ProtocolVersion: "1.0",
			}

			// This will test the authentication headers
			_, err := client.Invoke(agent, request)
			require.NoError(t, err)
		})
	}
}

func TestA2AClient_Retry_Logic(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount < 3 {
			// Fail the first two requests
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Server error"))
			return
		}
		// Success on the third request
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "data": "success after retries"}`))
	}))
	defer server.Close()

	agent := &types.A2AAgent{
		ID:          uuid.New(),
		Name:        "Retry Test Agent",
		EndpointURL: server.URL,
		AgentType:   types.AgentTypeGeneric,
		AuthType:    types.AuthTypeNone,
		IsActive:    true,
	}

	client := a2a.NewClient(5*time.Second, 3) // Allow 3 retries
	request := &types.A2ARequest{
		InteractionType: types.InteractionTypeQuery,
		Parameters:      map[string]interface{}{},
		ProtocolVersion: "1.0",
	}

	response, err := client.Invoke(agent, request)
	require.NoError(t, err)
	require.NotNil(t, response)

	assert.True(t, response.Success)
	assert.Equal(t, "success after retries", response.Data)
	assert.Equal(t, 3, requestCount, "Should have made exactly 3 requests")
}
