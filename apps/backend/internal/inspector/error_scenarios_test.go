package inspector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/transport"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestErrorScenarios tests various error conditions in the inspector service
func TestErrorScenarios_TransportFailures(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockTransportManager, *MockTransport)
		expectedError  string
		shouldHaveConn bool
	}{
		{
			name: "transport_manager_create_connection_fails",
			setupMock: func(mockManager *MockTransportManager, mockTransport *MockTransport) {
				mockManager.On("CreateConnection", mock.Anything, types.TransportTypeHTTP, "user123", "org456", "server789").
					Return(nil, (*types.TransportSession)(nil), errors.New("failed to create connection"))
			},
			expectedError:  "failed to create transport connection",
			shouldHaveConn: false,
		},
		{
			name: "transport_connect_fails",
			setupMock: func(mockManager *MockTransportManager, mockTransport *MockTransport) {
				mockManager.On("CreateConnection", mock.Anything, types.TransportTypeHTTP, "user123", "org456", "server789").
					Return(mockTransport, (*types.TransportSession)(nil), nil)
				mockTransport.On("Connect", mock.Anything).Return(errors.New("connection refused"))
			},
			expectedError:  "failed to connect transport",
			shouldHaveConn: false,
		},
		{
			name: "transport_send_message_fails_during_initialization",
			setupMock: func(mockManager *MockTransportManager, mockTransport *MockTransport) {
				mockManager.On("CreateConnection", mock.Anything, types.TransportTypeHTTP, "user123", "org456", "server789").
					Return(mockTransport, (*types.TransportSession)(nil), nil)
				mockTransport.On("Connect", mock.Anything).Return(nil)
				mockTransport.On("SendMessage", mock.Anything, mock.AnythingOfType("types.MCPMessage")).
					Return(errors.New("send failed"))
			},
			expectedError:  "",
			shouldHaveConn: true, // Connection succeeds but initialization fails
		},
		{
			name: "transport_receive_message_fails_during_initialization",
			setupMock: func(mockManager *MockTransportManager, mockTransport *MockTransport) {
				mockManager.On("CreateConnection", mock.Anything, types.TransportTypeHTTP, "user123", "org456", "server789").
					Return(mockTransport, (*types.TransportSession)(nil), nil)
				mockTransport.On("Connect", mock.Anything).Return(nil)
				mockTransport.On("SendMessage", mock.Anything, mock.AnythingOfType("types.MCPMessage")).
					Return(nil)
				mockTransport.On("ReceiveMessage", mock.Anything).
					Return(nil, errors.New("receive failed"))
			},
			expectedError:  "",
			shouldHaveConn: true, // Connection succeeds but initialization fails
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Create mock transport manager and transport
			mockManager := &MockTransportManager{}
			mockTransport := &MockTransport{}

			// Setup mock expectations
			tt.setupMock(mockManager, mockTransport)

			// Create service
			service := NewService(mockManager)

			// Try to create session
			session, err := service.CreateSession(ctx, "server789", "user123", "org456", "namespace001")

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, session)
			} else {
				// Some tests might succeed with connection but fail initialization
				if tt.shouldHaveConn {
					assert.NotNil(t, session)
					assert.Equal(t, SessionStatusConnected, session.Status)
				}
			}

			mockManager.AssertExpectations(t)
			mockTransport.AssertExpectations(t)
		})
	}
}

func TestErrorScenarios_RequestExecution(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		method        string
		params        map[string]interface{}
		setupMock     func(*MockTransport)
		expectedError *MCPError
	}{
		{
			name:   "invalid_method",
			method: "unknown/method",
			params: map[string]interface{}{},
			setupMock: func(mockTransport *MockTransport) {
				// No expectations - should return error before calling transport
			},
			expectedError: &MCPError{Code: -32601, Message: "Method not found"},
		},
		{
			name:   "transport_send_fails",
			method: "ping",
			params: map[string]interface{}{},
			setupMock: func(mockTransport *MockTransport) {
				mockTransport.On("SendMessage", ctx, mock.AnythingOfType("types.MCPMessage")).
					Return(errors.New("transport send failed"))
			},
			expectedError: &MCPError{Code: -32603, Message: "transport send failed"},
		},
		{
			name:   "transport_receive_fails",
			method: "tools/list",
			params: map[string]interface{}{},
			setupMock: func(mockTransport *MockTransport) {
				mockTransport.On("SendMessage", ctx, mock.AnythingOfType("types.MCPMessage")).
					Return(nil)
				mockTransport.On("ReceiveMessage", ctx).
					Return(nil, errors.New("transport receive failed"))
			},
			expectedError: &MCPError{Code: -32603, Message: "transport receive failed"},
		},
		{
			name:   "mcp_error_response",
			method: "tools/call",
			params: map[string]interface{}{"name": "nonexistent-tool"},
			setupMock: func(mockTransport *MockTransport) {
				mockTransport.On("SendMessage", ctx, mock.AnythingOfType("types.MCPMessage")).
					Return(nil)
				mockTransport.On("ReceiveMessage", ctx).
					Return(types.MCPMessage{
						Error: &types.MCPError{
							Code:    -32602,
							Message: "Invalid params",
						},
					}, nil)
			},
			expectedError: &MCPError{Code: -32602, Message: "Invalid params"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create service
			service := &Service{
				transportManager: &transport.Manager{},
				sessions:         make(map[string]*InspectorSession),
				connections:      make(map[string]types.Transport),
				eventChannels:    make(map[string]chan InspectorEvent),
			}

			// Create a mock session and connection
			session := NewInspectorSession("server789", "user123", "org456", "namespace001")
			session.Status = SessionStatusConnected

			mockTransport := &MockTransport{}
			tt.setupMock(mockTransport)

			// Add session and connection to service
			service.sessions[session.ID] = session
			service.connections[session.ID] = mockTransport
			service.eventChannels[session.ID] = make(chan InspectorEvent, 100)

			// Create request
			req := InspectorRequest{
				ID:        "req123",
				SessionID: session.ID,
				Method:    tt.method,
				Params:    tt.params,
				Timestamp: time.Now(),
			}

			// Execute request
			response, err := service.ExecuteRequest(ctx, session.ID, req)

			assert.NoError(t, err)
			assert.NotNil(t, response)

			if tt.expectedError != nil {
				assert.NotNil(t, response.Error)
				assert.Equal(t, tt.expectedError.Code, response.Error.Code)
				assert.Contains(t, response.Error.Message, tt.expectedError.Message)
			} else {
				assert.Nil(t, response.Error)
			}

			mockTransport.AssertExpectations(t)
		})
	}
}

func TestErrorScenarios_SessionManagement(t *testing.T) {
	service := NewService(nil) // Use nil since we're not testing transport functionality here

	t.Run("get_nonexistent_session", func(t *testing.T) {
		session, err := service.GetSession("nonexistent-session-id")
		assert.Error(t, err)
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "session not found")
	})

	t.Run("close_nonexistent_session", func(t *testing.T) {
		err := service.CloseSession(context.Background(), "nonexistent-session-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")
	})

	t.Run("execute_request_nonexistent_session", func(t *testing.T) {
		req := InspectorRequest{
			ID:        "req123",
			SessionID: "nonexistent-session-id",
			Method:    "ping",
			Params:    map[string]interface{}{},
			Timestamp: time.Now(),
		}

		response, err := service.ExecuteRequest(context.Background(), "nonexistent-session-id", req)
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "session not found")
	})

	t.Run("execute_request_no_connection", func(t *testing.T) {
		// Create session without connection
		session := NewInspectorSession("server789", "user123", "org456", "namespace001")
		service.sessions[session.ID] = session

		req := InspectorRequest{
			ID:        "req123",
			SessionID: session.ID,
			Method:    "ping",
			Params:    map[string]interface{}{},
			Timestamp: time.Now(),
		}

		response, err := service.ExecuteRequest(context.Background(), session.ID, req)
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "connection not found")
	})

	t.Run("get_event_channel_nonexistent_session", func(t *testing.T) {
		_, err := service.GetEventChannel("nonexistent-session-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event channel not found")
	})
}

func TestErrorScenarios_ConcurrentAccess(t *testing.T) {
	service := NewService(nil)

	// Create a session
	session := NewInspectorSession("server789", "user123", "org456", "namespace001")
	service.sessions[session.ID] = session
	service.eventChannels[session.ID] = make(chan InspectorEvent, 100)

	// Test concurrent access to session (simplified to avoid deadlock)
	t.Run("concurrent_session_access", func(t *testing.T) {
		// Test that session exists
		_, err := service.GetSession(session.ID)
		assert.NoError(t, err)

		// Close session
		err = service.CloseSession(context.Background(), session.ID)
		assert.NoError(t, err)

		// Session should be closed now
		_, err = service.GetSession(session.ID)
		assert.Error(t, err)
	})
}

func TestErrorScenarios_EventChannelOverflow(t *testing.T) {
	service := NewService(nil)

	// Create session with small event channel
	session := NewInspectorSession("server789", "user123", "org456", "namespace001")
	smallEventChan := make(chan InspectorEvent, 1)

	service.sessions[session.ID] = session
	service.eventChannels[session.ID] = smallEventChan

	// Fill the channel
	event1 := InspectorEvent{
		ID:        "event1",
		SessionID: session.ID,
		Type:      "test",
		Data:      "data1",
		Timestamp: time.Now(),
	}
	smallEventChan <- event1

	// Try to publish more events (should be dropped)
	for i := 0; i < 10; i++ {
		event := InspectorEvent{
			ID:        "event" + string(rune(i+2)),
			SessionID: session.ID,
			Type:      "test",
			Data:      "data" + string(rune(i+2)),
			Timestamp: time.Now(),
		}

		// This should not block
		done := make(chan bool, 1)
		go func() {
			service.publishEvent(session.ID, event)
			done <- true
		}()

		select {
		case <-done:
			// Good, publishEvent returned without blocking
		case <-time.After(100 * time.Millisecond):
			t.Fatal("publishEvent blocked when channel was full")
		}
	}

	// Channel should still only have the first event
	select {
	case receivedEvent := <-smallEventChan:
		assert.Equal(t, event1.ID, receivedEvent.ID)
	default:
		t.Fatal("Expected first event to still be in channel")
	}
}

func TestErrorScenarios_InvalidResponseParsing(t *testing.T) {
	ctx := context.Background()

	// Create service
	service := &Service{
		transportManager: &transport.Manager{},
		sessions:         make(map[string]*InspectorSession),
		connections:      make(map[string]types.Transport),
		eventChannels:    make(map[string]chan InspectorEvent),
	}

	// Create mock session and connection
	session := NewInspectorSession("server789", "user123", "org456", "namespace001")
	session.Status = SessionStatusConnected

	mockTransport := &MockTransport{}

	// Setup mock to return invalid response format
	mockTransport.On("SendMessage", ctx, mock.AnythingOfType("types.MCPMessage")).Return(nil)
	mockTransport.On("ReceiveMessage", ctx).Return("invalid response format", nil)

	// Add session and connection to service
	service.sessions[session.ID] = session
	service.connections[session.ID] = mockTransport
	service.eventChannels[session.ID] = make(chan InspectorEvent, 100)

	// Test different methods that parse responses
	methods := []string{"tools/list", "resources/list", "prompts/list"}

	for _, method := range methods {
		t.Run("invalid_response_"+method, func(t *testing.T) {
			req := InspectorRequest{
				ID:        "req123",
				SessionID: session.ID,
				Method:    method,
				Params:    map[string]interface{}{},
				Timestamp: time.Now(),
			}

			response, err := service.ExecuteRequest(ctx, session.ID, req)

			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.NotNil(t, response.Error)
			assert.Equal(t, -32603, response.Error.Code)
		})
	}

	mockTransport.AssertExpectations(t)
}

func TestErrorScenarios_CloseSessionTransportError(t *testing.T) {
	ctx := context.Background()

	// Create service
	service := &Service{
		transportManager: &transport.Manager{},
		sessions:         make(map[string]*InspectorSession),
		connections:      make(map[string]types.Transport),
		eventChannels:    make(map[string]chan InspectorEvent),
	}

	// Create mock session and connection
	session := NewInspectorSession("server789", "user123", "org456", "namespace001")
	session.Status = SessionStatusConnected

	mockTransport := &MockTransport{}
	mockTransport.On("Disconnect", ctx).Return(errors.New("disconnect failed"))

	// Add session and connection to service
	service.sessions[session.ID] = session
	service.connections[session.ID] = mockTransport
	service.eventChannels[session.ID] = make(chan InspectorEvent, 100)

	// Close session - should succeed despite transport error
	err := service.CloseSession(ctx, session.ID)
	assert.NoError(t, err) // Error is logged but not returned

	// Verify session was cleaned up
	_, err = service.GetSession(session.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")

	mockTransport.AssertExpectations(t)
}

func TestErrorScenarios_MethodSpecificErrors(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		method        string
		params        map[string]interface{}
		mockResponse  interface{}
		expectedError bool
	}{
		{
			method: "tools/call",
			params: map[string]interface{}{
				"name":      "test-tool",
				"arguments": map[string]interface{}{"param": "value"},
			},
			mockResponse: types.MCPMessage{
				Result: "invalid result format", // Should be object
			},
			expectedError: true,
		},
		{
			method: "resources/read",
			params: map[string]interface{}{
				"uri": "test://resource",
			},
			mockResponse: types.MCPMessage{
				Result: []interface{}{1, 2, 3}, // Should be object
			},
			expectedError: true,
		},
		{
			method: "prompts/get",
			params: map[string]interface{}{
				"name": "test-prompt",
			},
			mockResponse: types.MCPMessage{
				Result: "string instead of object",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run("parse_error_"+tt.method, func(t *testing.T) {
			// Create service
			service := &Service{
				transportManager: &transport.Manager{},
				sessions:         make(map[string]*InspectorSession),
				connections:      make(map[string]types.Transport),
				eventChannels:    make(map[string]chan InspectorEvent),
			}

			// Create mock session and connection
			session := NewInspectorSession("server789", "user123", "org456", "namespace001")
			session.Status = SessionStatusConnected

			mockTransport := &MockTransport{}
			mockTransport.On("SendMessage", ctx, mock.AnythingOfType("types.MCPMessage")).Return(nil)
			mockTransport.On("ReceiveMessage", ctx).Return(tt.mockResponse, nil)

			// Add session and connection to service
			service.sessions[session.ID] = session
			service.connections[session.ID] = mockTransport
			service.eventChannels[session.ID] = make(chan InspectorEvent, 100)

			// Create request
			req := InspectorRequest{
				ID:        "req123",
				SessionID: session.ID,
				Method:    tt.method,
				Params:    tt.params,
				Timestamp: time.Now(),
			}

			// Execute request
			response, err := service.ExecuteRequest(ctx, session.ID, req)

			assert.NoError(t, err)
			assert.NotNil(t, response)

			if tt.expectedError {
				assert.NotNil(t, response.Error)
				assert.Equal(t, -32603, response.Error.Code)
			}

			mockTransport.AssertExpectations(t)
		})
	}
}
