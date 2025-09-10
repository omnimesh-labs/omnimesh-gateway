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

// MockTransport is a mock implementation of the Transport interface
type MockTransport struct {
	mock.Mock
}

func (m *MockTransport) Connect(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTransport) Disconnect(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTransport) SendMessage(ctx context.Context, message interface{}) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockTransport) ReceiveMessage(ctx context.Context) (interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0), args.Error(1)
}

func (m *MockTransport) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockTransport) GetTransportType() types.TransportType {
	args := m.Called()
	return args.Get(0).(types.TransportType)
}

func (m *MockTransport) GetSessionID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockTransport) SetSessionID(sessionID string) {
	m.Called(sessionID)
}

// MockTransportManager is a mock implementation of the TransportManager
type MockTransportManager struct {
	mock.Mock
}

func (m *MockTransportManager) CreateConnection(ctx context.Context, transportType types.TransportType, userID, orgID, serverID string) (types.Transport, *types.TransportSession, error) {
	args := m.Called(ctx, transportType, userID, orgID, serverID)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(types.Transport), args.Get(1).(*types.TransportSession), args.Error(2)
}

func TestService_CreateSession(t *testing.T) {
	ctx := context.Background()

	// Create mock transport manager
	mockManager := &MockTransportManager{}
	mockTransport := &MockTransport{}

	// Setup expectations
	mockManager.On("CreateConnection", ctx, types.TransportTypeHTTP, "user123", "org456", "server789").
		Return(mockTransport, (*types.TransportSession)(nil), nil)

	mockTransport.On("Connect", ctx).Return(nil)
	mockTransport.On("SendMessage", ctx, mock.AnythingOfType("types.MCPMessage")).Return(nil)
	mockTransport.On("ReceiveMessage", ctx).Return(types.MCPMessage{
		Result: map[string]interface{}{
			"capabilities": map[string]interface{}{
				"tools":     true,
				"resources": true,
			},
		},
	}, nil)

	// For now, we'll test the session creation logic
	session := NewInspectorSession("server789", "user123", "org456", "namespace001")

	assert.NotNil(t, session)
	assert.Equal(t, "server789", session.ServerID)
	assert.Equal(t, "user123", session.UserID)
	assert.Equal(t, "org456", session.OrgID)
	assert.Equal(t, "namespace001", session.NamespaceID)
	assert.Equal(t, SessionStatusInitializing, session.Status)
	assert.NotEmpty(t, session.ID)
}

func TestService_ExecuteRequest(t *testing.T) {
	ctx := context.Background()

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

	// Add session and connection to service
	service.sessions[session.ID] = session
	service.connections[session.ID] = mockTransport
	service.eventChannels[session.ID] = make(chan InspectorEvent, 100)

	// Setup mock expectations for ping
	mockTransport.On("SendMessage", ctx, mock.AnythingOfType("types.MCPMessage")).Return(nil)
	mockTransport.On("ReceiveMessage", ctx).Return(types.MCPMessage{
		Result: map[string]interface{}{"status": "ok"},
	}, nil)

	// Create request
	req := InspectorRequest{
		ID:        "req123",
		SessionID: session.ID,
		Method:    "ping",
		Params:    map[string]interface{}{},
		Timestamp: time.Now(),
	}

	// Execute request
	response, err := service.ExecuteRequest(ctx, session.ID, req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Nil(t, response.Error)
	assert.Equal(t, "req123", response.RequestID)

	mockTransport.AssertExpectations(t)
}

func TestService_CloseSession(t *testing.T) {
	ctx := context.Background()

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
	mockTransport.On("Disconnect", ctx).Return(nil)

	// Add session and connection to service
	service.sessions[session.ID] = session
	service.connections[session.ID] = mockTransport
	service.eventChannels[session.ID] = make(chan InspectorEvent, 100)

	// Close session
	err := service.CloseSession(ctx, session.ID)

	assert.NoError(t, err)
	assert.Empty(t, service.sessions)
	assert.Empty(t, service.connections)
	assert.Empty(t, service.eventChannels)

	mockTransport.AssertExpectations(t)
}

func TestService_GetServerCapabilities(t *testing.T) {
	ctx := context.Background()

	// Create service
	service := NewService(nil)

	// Get capabilities
	capabilities, err := service.GetServerCapabilities(ctx, "server123")

	assert.NoError(t, err)
	assert.NotNil(t, capabilities)
	assert.NotNil(t, capabilities.Tools)
	assert.NotNil(t, capabilities.Resources)
	assert.NotNil(t, capabilities.Prompts)
	assert.NotNil(t, capabilities.Logging)
	assert.NotNil(t, capabilities.Sampling)
	assert.NotNil(t, capabilities.Roots)
}

func TestService_CreateSession_TransportError(t *testing.T) {
	ctx := context.Background()

	// Create mock transport manager
	mockManager := &MockTransportManager{}

	// Setup expectations for transport error
	mockManager.On("CreateConnection", ctx, types.TransportTypeHTTP, "user123", "org456", "server789").
		Return(nil, nil, errors.New("transport connection failed"))

	// Create service
	service := NewService(mockManager)

	// Try to create session - should fail
	session, err := service.CreateSession(ctx, "server789", "user123", "org456", "namespace001")

	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "failed to create transport connection")

	mockManager.AssertExpectations(t)
}

func TestService_CreateSession_ConnectError(t *testing.T) {
	ctx := context.Background()

	// Create mock transport manager and transport
	mockManager := &MockTransportManager{}
	mockTransport := &MockTransport{}

	// Setup expectations
	mockManager.On("CreateConnection", ctx, types.TransportTypeHTTP, "user123", "org456", "server789").
		Return(mockTransport, (*types.TransportSession)(nil), nil)
	mockTransport.On("Connect", ctx).Return(errors.New("connection failed"))

	// Create service
	service := NewService(mockManager)

	// Try to create session - should fail
	session, err := service.CreateSession(ctx, "server789", "user123", "org456", "namespace001")

	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "failed to connect transport")

	mockManager.AssertExpectations(t)
	mockTransport.AssertExpectations(t)
}

func TestService_GetSession_NotFound(t *testing.T) {
	// Create service
	service := NewService(nil)

	// Try to get non-existent session
	session, err := service.GetSession("non-existent-id")

	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "session not found")
}

func TestService_CloseSession_NotFound(t *testing.T) {
	ctx := context.Background()

	// Create service
	service := NewService(nil)

	// Try to close non-existent session
	err := service.CloseSession(ctx, "non-existent-id")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestService_ExecuteRequest_SessionNotFound(t *testing.T) {
	ctx := context.Background()

	// Create service
	service := NewService(nil)

	// Create request
	req := InspectorRequest{
		ID:        "req123",
		SessionID: "non-existent-session",
		Method:    "ping",
		Params:    map[string]interface{}{},
		Timestamp: time.Now(),
	}

	// Execute request - should fail
	response, err := service.ExecuteRequest(ctx, "non-existent-session", req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "session not found")
}

func TestService_ExecuteRequest_ConnectionNotFound(t *testing.T) {
	ctx := context.Background()

	// Create service
	service := NewService(nil)

	// Create a session without connection
	session := NewInspectorSession("server789", "user123", "org456", "namespace001")
	service.sessions[session.ID] = session

	// Create request
	req := InspectorRequest{
		ID:        "req123",
		SessionID: session.ID,
		Method:    "ping",
		Params:    map[string]interface{}{},
		Timestamp: time.Now(),
	}

	// Execute request - should fail
	response, err := service.ExecuteRequest(ctx, session.ID, req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "connection not found")
}

func TestService_ExecuteRequest_InvalidMethod(t *testing.T) {
	ctx := context.Background()

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

	// Add session and connection to service
	service.sessions[session.ID] = session
	service.connections[session.ID] = mockTransport
	service.eventChannels[session.ID] = make(chan InspectorEvent, 100)

	// Create request with invalid method
	req := InspectorRequest{
		ID:        "req123",
		SessionID: session.ID,
		Method:    "invalid/method",
		Params:    map[string]interface{}{},
		Timestamp: time.Now(),
	}

	// Execute request - should return error response
	response, err := service.ExecuteRequest(ctx, session.ID, req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotNil(t, response.Error)
	assert.Equal(t, -32601, response.Error.Code)
	assert.Equal(t, "Method not found", response.Error.Message)
}

func TestService_GetEventChannel_NotFound(t *testing.T) {
	// Create service
	service := NewService(nil)

	// Try to get event channel for non-existent session
	_, err := service.GetEventChannel("non-existent-id")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event channel not found")
}

func TestService_ListTools_TransportError(t *testing.T) {
	ctx := context.Background()

	// Create service
	service := &Service{
		transportManager: &transport.Manager{},
		sessions:         make(map[string]*InspectorSession),
		connections:      make(map[string]types.Transport),
		eventChannels:    make(map[string]chan InspectorEvent),
	}

	// Create mock transport that returns error
	mockTransport := &MockTransport{}
	mockTransport.On("SendMessage", ctx, mock.AnythingOfType("types.MCPMessage")).Return(errors.New("transport error"))

	// Test listTools method
	result, mcpErr := service.listTools(ctx, mockTransport, map[string]interface{}{})

	assert.Nil(t, result)
	assert.NotNil(t, mcpErr)
	assert.Equal(t, -32603, mcpErr.Code)
	assert.Contains(t, mcpErr.Message, "transport error")

	mockTransport.AssertExpectations(t)
}

func TestService_EventPublishing(t *testing.T) {
	// Create service
	service := NewService(nil)

	// Create session and event channel
	session := NewInspectorSession("server789", "user123", "org456", "namespace001")
	eventChan := make(chan InspectorEvent, 100)

	service.sessions[session.ID] = session
	service.eventChannels[session.ID] = eventChan

	// Publish event
	event := InspectorEvent{
		ID:        "event123",
		SessionID: session.ID,
		Type:      "test_event",
		Data:      "test data",
		Timestamp: time.Now(),
	}

	service.publishEvent(session.ID, event)

	// Check event was received
	select {
	case receivedEvent := <-eventChan:
		assert.Equal(t, event.ID, receivedEvent.ID)
		assert.Equal(t, event.Type, receivedEvent.Type)
		assert.Equal(t, event.Data, receivedEvent.Data)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected event was not received")
	}
}

func TestService_EventPublishing_ChannelFull(t *testing.T) {
	// Create service
	service := NewService(nil)

	// Create session and small event channel
	session := NewInspectorSession("server789", "user123", "org456", "namespace001")
	eventChan := make(chan InspectorEvent, 1)

	service.sessions[session.ID] = session
	service.eventChannels[session.ID] = eventChan

	// Fill the channel
	event1 := InspectorEvent{ID: "event1", SessionID: session.ID, Type: "test", Data: "data1", Timestamp: time.Now()}
	eventChan <- event1

	// Try to publish another event (should be dropped due to full channel)
	event2 := InspectorEvent{ID: "event2", SessionID: session.ID, Type: "test", Data: "data2", Timestamp: time.Now()}

	// This should not block (event should be dropped)
	done := make(chan bool, 1)
	go func() {
		service.publishEvent(session.ID, event2)
		done <- true
	}()

	select {
	case <-done:
		// Good, publishEvent returned without blocking
	case <-time.After(100 * time.Millisecond):
		t.Fatal("publishEvent blocked when channel was full")
	}
}

func TestInspectorSession_Creation(t *testing.T) {
	session := NewInspectorSession("server123", "user456", "org789", "namespace001")

	assert.NotEmpty(t, session.ID)
	assert.Equal(t, "server123", session.ServerID)
	assert.Equal(t, "user456", session.UserID)
	assert.Equal(t, "org789", session.OrgID)
	assert.Equal(t, "namespace001", session.NamespaceID)
	assert.Equal(t, SessionStatusInitializing, session.Status)
	assert.NotZero(t, session.CreatedAt)
	assert.NotZero(t, session.LastActivity)
	assert.NotNil(t, session.Capabilities)
}

func TestMCPError_String(t *testing.T) {
	err := &MCPError{
		Code:    -32601,
		Message: "Method not found",
		Data:    "additional data",
	}

	str := err.String()
	assert.Contains(t, str, "-32601")
	assert.Contains(t, str, "Method not found")
}

func TestSessionStatus_String(t *testing.T) {
	assert.Equal(t, "initializing", SessionStatusInitializing.String())
	assert.Equal(t, "connected", SessionStatusConnected.String())
	assert.Equal(t, "disconnected", SessionStatusDisconnected.String())
	assert.Equal(t, "error", SessionStatusError.String())
}
