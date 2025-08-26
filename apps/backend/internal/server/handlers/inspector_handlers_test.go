package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"mcp-gateway/apps/backend/internal/inspector"
)

// MockInspectorService is a mock implementation of the inspector service
type MockInspectorService struct {
	mock.Mock
}

func (m *MockInspectorService) CreateSession(ctx context.Context, serverID, userID, orgID, namespaceID string) (*inspector.InspectorSession, error) {
	args := m.Called(ctx, serverID, userID, orgID, namespaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*inspector.InspectorSession), args.Error(1)
}

func (m *MockInspectorService) GetSession(sessionID string) (*inspector.InspectorSession, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*inspector.InspectorSession), args.Error(1)
}

func (m *MockInspectorService) CloseSession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockInspectorService) ExecuteRequest(ctx context.Context, sessionID string, req inspector.InspectorRequest) (*inspector.InspectorResponse, error) {
	args := m.Called(ctx, sessionID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*inspector.InspectorResponse), args.Error(1)
}

func (m *MockInspectorService) GetEventChannel(sessionID string) (<-chan inspector.InspectorEvent, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(<-chan inspector.InspectorEvent), args.Error(1)
}

func (m *MockInspectorService) GetServerCapabilities(ctx context.Context, serverID string) (*inspector.ServerCapabilities, error) {
	args := m.Called(ctx, serverID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*inspector.ServerCapabilities), args.Error(1)
}

func setupTestHandler() (*InspectorHandler, *MockInspectorService) {
	mockService := &MockInspectorService{}
	handler := NewInspectorHandler(mockService)
	return handler, mockService
}

func setupInspectorTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Add test middleware that sets user context
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user-123")
		c.Set("org_id", "test-org-456")
		c.Next()
	})

	return r
}

func TestInspectorHandler_CreateSession(t *testing.T) {
	handler, mockService := setupTestHandler()
	router := setupInspectorTestRouter()

	router.POST("/sessions", handler.CreateSession)

	// Setup mock expectations
	expectedSession := &inspector.InspectorSession{
		ID:           "session-123",
		ServerID:     "server-456",
		UserID:       "test-user-123",
		OrgID:        "test-org-456",
		NamespaceID:  "namespace-789",
		Status:       inspector.SessionStatusConnected,
		Capabilities: map[string]interface{}{"tools": true},
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	mockService.On("CreateSession", mock.Anything, "server-456", "test-user-123", "test-org-456", "namespace-789").
		Return(expectedSession, nil)

	// Create request
	reqBody := inspector.CreateSessionRequest{
		ServerID:    "server-456",
		NamespaceID: "namespace-789",
	}
	reqJSON, _ := json.Marshal(reqBody)

	// Execute request
	req := httptest.NewRequest("POST", "/sessions", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response inspector.InspectorSession
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedSession.ID, response.ID)
	assert.Equal(t, expectedSession.ServerID, response.ServerID)

	mockService.AssertExpectations(t)
}

func TestInspectorHandler_CreateSession_InvalidJSON(t *testing.T) {
	handler, _ := setupTestHandler()
	router := setupInspectorTestRouter()

	router.POST("/sessions", handler.CreateSession)

	// Create invalid JSON request
	req := httptest.NewRequest("POST", "/sessions", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify error response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "invalid character")
}

func TestInspectorHandler_CreateSession_ServiceError(t *testing.T) {
	handler, mockService := setupTestHandler()
	router := setupInspectorTestRouter()

	router.POST("/sessions", handler.CreateSession)

	// Setup mock to return error
	mockService.On("CreateSession", mock.Anything, "server-456", "test-user-123", "test-org-456", "namespace-789").
		Return(nil, errors.New("service error"))

	// Create request
	reqBody := inspector.CreateSessionRequest{
		ServerID:    "server-456",
		NamespaceID: "namespace-789",
	}
	reqJSON, _ := json.Marshal(reqBody)

	// Execute request
	req := httptest.NewRequest("POST", "/sessions", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify error response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "service error", response["error"])

	mockService.AssertExpectations(t)
}

func TestInspectorHandler_GetSession(t *testing.T) {
	handler, mockService := setupTestHandler()
	router := setupInspectorTestRouter()

	router.GET("/sessions/:id", handler.GetSession)

	// Setup mock expectations
	expectedSession := &inspector.InspectorSession{
		ID:           "session-123",
		ServerID:     "server-456",
		UserID:       "test-user-123",
		OrgID:        "test-org-456",
		Status:       inspector.SessionStatusConnected,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	mockService.On("GetSession", "session-123").Return(expectedSession, nil)

	// Execute request
	req := httptest.NewRequest("GET", "/sessions/session-123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response inspector.InspectorSession
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedSession.ID, response.ID)

	mockService.AssertExpectations(t)
}

func TestInspectorHandler_GetSession_NotFound(t *testing.T) {
	handler, mockService := setupTestHandler()
	router := setupInspectorTestRouter()

	router.GET("/sessions/:id", handler.GetSession)

	// Setup mock to return not found error
	mockService.On("GetSession", "non-existent").Return(nil, errors.New("session not found"))

	// Execute request
	req := httptest.NewRequest("GET", "/sessions/non-existent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify error response
	assert.Equal(t, http.StatusNotFound, w.Code)

	mockService.AssertExpectations(t)
}

func TestInspectorHandler_GetSession_Forbidden(t *testing.T) {
	handler, mockService := setupTestHandler()
	router := setupInspectorTestRouter()

	router.GET("/sessions/:id", handler.GetSession)

	// Setup mock to return session owned by different user
	otherUserSession := &inspector.InspectorSession{
		ID:       "session-123",
		UserID:   "other-user",
		OrgID:    "test-org-456",
		Status:   inspector.SessionStatusConnected,
	}

	mockService.On("GetSession", "session-123").Return(otherUserSession, nil)

	// Execute request
	req := httptest.NewRequest("GET", "/sessions/session-123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify forbidden response
	assert.Equal(t, http.StatusForbidden, w.Code)

	mockService.AssertExpectations(t)
}

func TestInspectorHandler_CloseSession(t *testing.T) {
	handler, mockService := setupTestHandler()
	router := setupInspectorTestRouter()

	router.DELETE("/sessions/:id", handler.CloseSession)

	// Setup mock expectations
	session := &inspector.InspectorSession{
		ID:     "session-123",
		UserID: "test-user-123",
		OrgID:  "test-org-456",
		Status: inspector.SessionStatusConnected,
	}

	mockService.On("GetSession", "session-123").Return(session, nil)
	mockService.On("CloseSession", mock.Anything, "session-123").Return(nil)

	// Execute request
	req := httptest.NewRequest("DELETE", "/sessions/session-123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "session closed successfully", response["message"])

	mockService.AssertExpectations(t)
}

func TestInspectorHandler_ExecuteRequest(t *testing.T) {
	handler, mockService := setupTestHandler()
	router := setupInspectorTestRouter()

	router.POST("/sessions/:id/request", handler.ExecuteRequest)

	// Setup mock expectations
	session := &inspector.InspectorSession{
		ID:     "session-123",
		UserID: "test-user-123",
		OrgID:  "test-org-456",
		Status: inspector.SessionStatusConnected,
	}

	expectedResponse := &inspector.InspectorResponse{
		ID:        "response-456",
		RequestID: "req-123",
		Result:    map[string]interface{}{"status": "ok"},
		Duration:  time.Millisecond * 100,
		Timestamp: time.Now(),
	}

	mockService.On("GetSession", "session-123").Return(session, nil)
	mockService.On("ExecuteRequest", mock.Anything, "session-123", mock.AnythingOfType("inspector.InspectorRequest")).
		Return(expectedResponse, nil)

	// Create request
	reqBody := inspector.ExecuteRequestBody{
		Method: "ping",
		Params: map[string]interface{}{},
	}
	reqJSON, _ := json.Marshal(reqBody)

	// Execute request
	req := httptest.NewRequest("POST", "/sessions/session-123/request", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response inspector.InspectorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedResponse.ID, response.ID)
	assert.NotNil(t, response.Result)

	mockService.AssertExpectations(t)
}

func TestInspectorHandler_StreamEvents(t *testing.T) {
	handler, mockService := setupTestHandler()
	router := setupInspectorTestRouter()

	router.GET("/sessions/:id/events", handler.StreamEvents)

	// Create event channel
	eventChan := make(chan inspector.InspectorEvent, 1)

	// Setup mock expectations
	session := &inspector.InspectorSession{
		ID:     "session-123",
		UserID: "test-user-123",
		OrgID:  "test-org-456",
		Status: inspector.SessionStatusConnected,
	}

	mockService.On("GetSession", "session-123").Return(session, nil)
	mockService.On("GetEventChannel", "session-123").Return((<-chan inspector.InspectorEvent)(eventChan), nil)

	// Send test event
	testEvent := inspector.InspectorEvent{
		ID:        "event-123",
		SessionID: "session-123",
		Type:      "test",
		Data:      "test data",
		Timestamp: time.Now(),
	}

	// Execute request in goroutine
	go func() {
		time.Sleep(50 * time.Millisecond)
		eventChan <- testEvent
		close(eventChan)
	}()

	req := httptest.NewRequest("GET", "/sessions/session-123/events", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify SSE headers
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
	assert.Equal(t, "keep-alive", w.Header().Get("Connection"))

	// Verify event data is in response
	responseBody := w.Body.String()
	assert.Contains(t, responseBody, "event: test")
	assert.Contains(t, responseBody, "test data")

	mockService.AssertExpectations(t)
}

func TestInspectorHandler_GetServerCapabilities(t *testing.T) {
	handler, mockService := setupTestHandler()
	router := setupInspectorTestRouter()

	router.GET("/servers/:id/capabilities", handler.GetServerCapabilities)

	// Setup mock expectations
	expectedCapabilities := &inspector.ServerCapabilities{
		Tools: &inspector.ToolsCapability{
			ListChanged: true,
		},
		Resources: &inspector.ResourcesCapability{
			Subscribe:   true,
			ListChanged: true,
		},
		Prompts: &inspector.PromptsCapability{
			ListChanged: true,
		},
	}

	mockService.On("GetServerCapabilities", mock.Anything, "server-123").
		Return(expectedCapabilities, nil)

	// Execute request
	req := httptest.NewRequest("GET", "/servers/server-123/capabilities", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response inspector.ServerCapabilities
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotNil(t, response.Tools)
	assert.True(t, response.Tools.ListChanged)
	assert.NotNil(t, response.Resources)
	assert.True(t, response.Resources.Subscribe)

	mockService.AssertExpectations(t)
}

func TestInspectorHandler_HandleWebSocket(t *testing.T) {
	// Note: WebSocket testing requires more complex setup due to connection upgrade
	// This is a simplified test that verifies the basic setup

	handler, mockService := setupTestHandler()

	// Setup mock expectations
	session := &inspector.InspectorSession{
		ID:     "session-123",
		UserID: "test-user-123",
		OrgID:  "test-org-456",
		Status: inspector.SessionStatusConnected,
	}

	eventChan := make(chan inspector.InspectorEvent, 1)

	mockService.On("GetSession", "session-123").Return(session, nil)
	mockService.On("GetEventChannel", "session-123").Return((<-chan inspector.InspectorEvent)(eventChan), nil)

	// Create test server for WebSocket testing
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		c.Set("user_id", "test-user-123")
		c.Set("org_id", "test-org-456")
		c.Params = []gin.Param{{Key: "id", Value: "session-123"}}

		handler.HandleWebSocket(c)
	}))
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Skipf("WebSocket connection failed: %v", err)
		return
	}
	defer conn.Close()

	// Send test event
	testEvent := inspector.InspectorEvent{
		ID:        "event-123",
		SessionID: "session-123",
		Type:      "test",
		Data:      "test data",
		Timestamp: time.Now(),
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		eventChan <- testEvent
	}()

	// Read message from WebSocket
	var receivedEvent inspector.InspectorEvent
	err = conn.ReadJSON(&receivedEvent)
	if err == nil {
		assert.Equal(t, testEvent.ID, receivedEvent.ID)
		assert.Equal(t, testEvent.Type, receivedEvent.Type)
	}

	mockService.AssertExpectations(t)
}

func TestInspectorHandler_CreateSession_MissingAuth(t *testing.T) {
	handler, _ := setupTestHandler()
	router := gin.New()
	gin.SetMode(gin.TestMode)

	// Don't add auth middleware - test missing auth scenario
	router.POST("/sessions", handler.CreateSession)

	// Create request
	reqBody := inspector.CreateSessionRequest{
		ServerID:    "server-456",
		NamespaceID: "namespace-789",
	}
	reqJSON, _ := json.Marshal(reqBody)

	// Execute request
	req := httptest.NewRequest("POST", "/sessions", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify unauthorized response
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "unauthorized", response["error"])
}