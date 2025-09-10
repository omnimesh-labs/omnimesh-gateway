package inspector

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
)

// TransportManager interface for creating transport connections
type TransportManager interface {
	CreateConnection(ctx context.Context, transportType types.TransportType, userID, orgID, serverID string) (types.Transport, *types.TransportSession, error)
}

// InspectorService interface for inspector operations
type InspectorService interface {
	CreateSession(ctx context.Context, serverID, userID, orgID, namespaceID string) (*InspectorSession, error)
	GetSession(sessionID string) (*InspectorSession, error)
	CloseSession(ctx context.Context, sessionID string) error
	ExecuteRequest(ctx context.Context, sessionID string, req InspectorRequest) (*InspectorResponse, error)
	GetEventChannel(sessionID string) (<-chan InspectorEvent, error)
	GetServerCapabilities(ctx context.Context, serverID string) (*ServerCapabilities, error)
}

// Service manages inspector sessions and MCP connections
type Service struct {
	transportManager TransportManager
	sessions         map[string]*InspectorSession
	connections      map[string]types.Transport
	eventChannels    map[string]chan InspectorEvent
	mu               sync.RWMutex
}

// NewService creates a new inspector service
func NewService(transportManager TransportManager) *Service {
	return &Service{
		transportManager: transportManager,
		sessions:         make(map[string]*InspectorSession),
		connections:      make(map[string]types.Transport),
		eventChannels:    make(map[string]chan InspectorEvent),
	}
}

// CreateSession creates a new inspector session
func (s *Service) CreateSession(ctx context.Context, serverID, userID, orgID, namespaceID string) (*InspectorSession, error) {
	session := NewInspectorSession(serverID, userID, orgID, namespaceID)

	// Create transport connection - use HTTP transport for JSON-RPC
	transport, _, err := s.transportManager.CreateConnection(ctx, types.TransportTypeHTTP, userID, orgID, serverID)
	if err != nil {
		session.Status = SessionStatusError
		return nil, fmt.Errorf("failed to create transport connection: %w", err)
	}

	// Initialize the connection
	if err := transport.Connect(ctx); err != nil {
		session.Status = SessionStatusError
		return nil, fmt.Errorf("failed to connect transport: %w", err)
	}

	// Get server capabilities
	capabilities, err := s.getServerCapabilities(ctx, transport)
	if err != nil {
		// Non-fatal: server might not support capabilities
		capabilities = make(map[string]interface{})
	}

	session.Capabilities = capabilities
	session.Status = SessionStatusConnected

	// Store session and connection
	s.mu.Lock()
	s.sessions[session.ID] = session
	s.connections[session.ID] = transport
	s.eventChannels[session.ID] = make(chan InspectorEvent, 100)
	s.mu.Unlock()

	// Publish connection event
	s.publishEvent(session.ID, InspectorEvent{
		ID:        uuid.New().String(),
		SessionID: session.ID,
		Type:      "session_created",
		Data:      session,
		Timestamp: time.Now(),
	})

	return session, nil
}

// GetSession retrieves a session by ID
func (s *Service) GetSession(sessionID string) (*InspectorSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session, nil
}

// CloseSession closes an inspector session
func (s *Service) CloseSession(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Close transport connection
	if conn, ok := s.connections[sessionID]; ok {
		if err := conn.Disconnect(ctx); err != nil {
			// Log error but continue cleanup
			fmt.Printf("Error closing transport connection: %v\n", err)
		}
		delete(s.connections, sessionID)
	}

	// Update session status
	session.Status = SessionStatusDisconnected

	// Publish disconnection event before closing channel
	if ch, ok := s.eventChannels[sessionID]; ok {
		event := InspectorEvent{
			ID:        uuid.New().String(),
			SessionID: sessionID,
			Type:      "session_closed",
			Data:      session,
			Timestamp: time.Now(),
		}

		// Try to send event before closing channel
		select {
		case ch <- event:
			// Event sent
		default:
			// Channel full, skip event
		}

		close(ch)
		delete(s.eventChannels, sessionID)
	}

	// Remove session
	delete(s.sessions, sessionID)

	return nil
}

// ExecuteRequest executes an MCP request on a session
func (s *Service) ExecuteRequest(ctx context.Context, sessionID string, req InspectorRequest) (*InspectorResponse, error) {
	start := time.Now()

	// Get session and connection
	s.mu.RLock()
	session, sessionExists := s.sessions[sessionID]
	conn, connExists := s.connections[sessionID]
	s.mu.RUnlock()

	if !sessionExists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	if !connExists {
		return nil, fmt.Errorf("connection not found for session: %s", sessionID)
	}

	// Update last activity
	session.LastActivity = time.Now()

	// Execute based on method
	var result interface{}
	var mcpErr *MCPError

	switch req.Method {
	case "tools/list":
		result, mcpErr = s.listTools(ctx, conn, req.Params)
	case "tools/call":
		result, mcpErr = s.callTool(ctx, conn, req.Params)
	case "resources/list":
		result, mcpErr = s.listResources(ctx, conn, req.Params)
	case "resources/read":
		result, mcpErr = s.readResource(ctx, conn, req.Params)
	case "prompts/list":
		result, mcpErr = s.listPrompts(ctx, conn, req.Params)
	case "prompts/get":
		result, mcpErr = s.getPrompt(ctx, conn, req.Params)
	case "ping":
		result, mcpErr = s.ping(ctx, conn)
	case "completion/complete":
		result, mcpErr = s.complete(ctx, conn, req.Params)
	case "initialize":
		result, mcpErr = s.initialize(ctx, conn, req.Params)
	default:
		mcpErr = &MCPError{Code: -32601, Message: "Method not found"}
	}

	response := &InspectorResponse{
		ID:        uuid.New().String(),
		RequestID: req.ID,
		Result:    result,
		Error:     mcpErr,
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	}

	// Publish response event
	s.publishEvent(sessionID, InspectorEvent{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Type:      "response",
		Data:      response,
		Timestamp: time.Now(),
	})

	return response, nil
}

// GetEventChannel returns the event channel for a session
func (s *Service) GetEventChannel(sessionID string) (<-chan InspectorEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ch, exists := s.eventChannels[sessionID]
	if !exists {
		return nil, fmt.Errorf("event channel not found for session: %s", sessionID)
	}

	return ch, nil
}

// GetServerCapabilities retrieves capabilities for a server
func (s *Service) GetServerCapabilities(ctx context.Context, serverID string) (*ServerCapabilities, error) {
	// This would typically query the server directly or from cache
	// For now, return a standard set of capabilities
	return &ServerCapabilities{
		Tools:     &ToolsCapability{ListChanged: true},
		Resources: &ResourcesCapability{Subscribe: true, ListChanged: true},
		Prompts:   &PromptsCapability{ListChanged: true},
		Logging:   &LoggingCapability{Level: "info"},
		Sampling:  &SamplingCapability{},
		Roots:     &RootsCapability{ListChanged: true},
	}, nil
}

// Private helper methods

func (s *Service) getServerCapabilities(ctx context.Context, transport types.Transport) (map[string]interface{}, error) {
	// Send initialize request to get capabilities
	request := types.MCPMessage{
		ID:     uuid.New().String(),
		Type:   types.MCPMessageTypeRequest,
		Method: "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "1.0.0",
			"clientInfo": map[string]interface{}{
				"name":    "omnimesh-gateway-inspector",
				"version": "1.0.0",
			},
		},
		Version: "2.0",
	}

	if err := transport.SendMessage(ctx, request); err != nil {
		return nil, err
	}

	response, err := transport.ReceiveMessage(ctx)
	if err != nil {
		return nil, err
	}

	// Parse capabilities from response
	if msg, ok := response.(types.MCPMessage); ok {
		if msg.Result != nil {
			if resultMap, ok := msg.Result.(map[string]interface{}); ok {
				if capabilities, ok := resultMap["capabilities"].(map[string]interface{}); ok {
					return capabilities, nil
				}
			}
		}
	}

	return make(map[string]interface{}), nil
}

// sendAndReceive is a helper to send MCP requests and receive responses
func (s *Service) sendAndReceive(ctx context.Context, transport types.Transport, method string, params map[string]interface{}) (*types.MCPMessage, error) {
	request := types.MCPMessage{
		ID:      uuid.New().String(),
		Type:    types.MCPMessageTypeRequest,
		Method:  method,
		Params:  params,
		Version: "2.0",
	}

	if err := transport.SendMessage(ctx, request); err != nil {
		return nil, err
	}

	response, err := transport.ReceiveMessage(ctx)
	if err != nil {
		return nil, err
	}

	if msg, ok := response.(types.MCPMessage); ok {
		return &msg, nil
	}

	// Try to convert the response to MCPMessage
	if data, err := json.Marshal(response); err == nil {
		var msg types.MCPMessage
		if err := json.Unmarshal(data, &msg); err == nil {
			return &msg, nil
		}
	}

	return nil, fmt.Errorf("invalid response type")
}

func (s *Service) listTools(ctx context.Context, transport types.Transport, params map[string]interface{}) (*ListToolsResult, *MCPError) {
	response, err := s.sendAndReceive(ctx, transport, "tools/list", params)
	if err != nil {
		return nil, &MCPError{Code: -32603, Message: err.Error()}
	}

	if response.Error != nil {
		return nil, &MCPError{Code: response.Error.Code, Message: response.Error.Message}
	}

	// Parse response
	var result ListToolsResult
	if data, err := json.Marshal(response.Result); err == nil {
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, &MCPError{Code: -32603, Message: "Failed to parse tools list"}
		}
	}

	return &result, nil
}

func (s *Service) callTool(ctx context.Context, transport types.Transport, params map[string]interface{}) (*CallToolResult, *MCPError) {
	response, err := s.sendAndReceive(ctx, transport, "tools/call", params)
	if err != nil {
		return nil, &MCPError{Code: -32603, Message: err.Error()}
	}

	if response.Error != nil {
		return nil, &MCPError{Code: response.Error.Code, Message: response.Error.Message}
	}

	// Parse response
	var result CallToolResult
	if data, err := json.Marshal(response.Result); err == nil {
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, &MCPError{Code: -32603, Message: "Failed to parse tool result"}
		}
	}

	return &result, nil
}

func (s *Service) listResources(ctx context.Context, transport types.Transport, params map[string]interface{}) (*ListResourcesResult, *MCPError) {
	response, err := s.sendAndReceive(ctx, transport, "resources/list", params)
	if err != nil {
		return nil, &MCPError{Code: -32603, Message: err.Error()}
	}

	if response.Error != nil {
		return nil, &MCPError{Code: response.Error.Code, Message: response.Error.Message}
	}

	// Parse response
	var result ListResourcesResult
	if data, err := json.Marshal(response.Result); err == nil {
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, &MCPError{Code: -32603, Message: "Failed to parse resources list"}
		}
	}

	return &result, nil
}

func (s *Service) readResource(ctx context.Context, transport types.Transport, params map[string]interface{}) (*ReadResourceResult, *MCPError) {
	response, err := s.sendAndReceive(ctx, transport, "resources/read", params)
	if err != nil {
		return nil, &MCPError{Code: -32603, Message: err.Error()}
	}

	if response.Error != nil {
		return nil, &MCPError{Code: response.Error.Code, Message: response.Error.Message}
	}

	// Parse response
	var result ReadResourceResult
	if data, err := json.Marshal(response.Result); err == nil {
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, &MCPError{Code: -32603, Message: "Failed to parse resource content"}
		}
	}

	return &result, nil
}

func (s *Service) listPrompts(ctx context.Context, transport types.Transport, params map[string]interface{}) (*ListPromptsResult, *MCPError) {
	response, err := s.sendAndReceive(ctx, transport, "prompts/list", params)
	if err != nil {
		return nil, &MCPError{Code: -32603, Message: err.Error()}
	}

	if response.Error != nil {
		return nil, &MCPError{Code: response.Error.Code, Message: response.Error.Message}
	}

	// Parse response
	var result ListPromptsResult
	if data, err := json.Marshal(response.Result); err == nil {
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, &MCPError{Code: -32603, Message: "Failed to parse prompts list"}
		}
	}

	return &result, nil
}

func (s *Service) getPrompt(ctx context.Context, transport types.Transport, params map[string]interface{}) (*GetPromptResult, *MCPError) {
	response, err := s.sendAndReceive(ctx, transport, "prompts/get", params)
	if err != nil {
		return nil, &MCPError{Code: -32603, Message: err.Error()}
	}

	if response.Error != nil {
		return nil, &MCPError{Code: response.Error.Code, Message: response.Error.Message}
	}

	// Parse response
	var result GetPromptResult
	if data, err := json.Marshal(response.Result); err == nil {
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, &MCPError{Code: -32603, Message: "Failed to parse prompt"}
		}
	}

	return &result, nil
}

func (s *Service) ping(ctx context.Context, transport types.Transport) (*PingResult, *MCPError) {
	response, err := s.sendAndReceive(ctx, transport, "ping", map[string]interface{}{})
	if err != nil {
		return nil, &MCPError{Code: -32603, Message: err.Error()}
	}

	if response.Error != nil {
		return nil, &MCPError{Code: response.Error.Code, Message: response.Error.Message}
	}

	return &PingResult{
		Status:    "ok",
		Timestamp: time.Now(),
	}, nil
}

func (s *Service) complete(ctx context.Context, transport types.Transport, params map[string]interface{}) (interface{}, *MCPError) {
	response, err := s.sendAndReceive(ctx, transport, "completion/complete", params)
	if err != nil {
		return nil, &MCPError{Code: -32603, Message: err.Error()}
	}

	if response.Error != nil {
		return nil, &MCPError{Code: response.Error.Code, Message: response.Error.Message}
	}

	return response.Result, nil
}

func (s *Service) initialize(ctx context.Context, transport types.Transport, params map[string]interface{}) (interface{}, *MCPError) {
	response, err := s.sendAndReceive(ctx, transport, "initialize", params)
	if err != nil {
		return nil, &MCPError{Code: -32603, Message: err.Error()}
	}

	if response.Error != nil {
		return nil, &MCPError{Code: response.Error.Code, Message: response.Error.Message}
	}

	return response.Result, nil
}

func (s *Service) publishEvent(sessionID string, event InspectorEvent) {
	s.mu.RLock()
	ch, exists := s.eventChannels[sessionID]
	s.mu.RUnlock()

	if exists {
		select {
		case ch <- event:
			// Event sent
		default:
			// Channel full, drop event (could log this)
		}
	}
}
