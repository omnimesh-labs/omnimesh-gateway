package transport

import (
	"fmt"
	"sync"

	"mcp-gateway/apps/backend/internal/types"
)

// BaseTransport provides common functionality for all transport implementations
type BaseTransport struct {
	metadata      map[string]interface{}
	transportType types.TransportType
	sessionID     string
	mu            sync.RWMutex
	connected     bool
}

// NewBaseTransport creates a new base transport instance
func NewBaseTransport(transportType types.TransportType) *BaseTransport {
	return &BaseTransport{
		transportType: transportType,
		connected:     false,
		metadata:      make(map[string]interface{}),
	}
}

// GetTransportType returns the type of this transport
func (b *BaseTransport) GetTransportType() types.TransportType {
	return b.transportType
}

// GetSessionID returns the session ID for this transport
func (b *BaseTransport) GetSessionID() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.sessionID
}

// SetSessionID sets the session ID for this transport
func (b *BaseTransport) SetSessionID(sessionID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.sessionID = sessionID
}

// IsConnected returns whether the transport is currently connected
func (b *BaseTransport) IsConnected() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.connected
}

// setConnected sets the connection status (internal use)
func (b *BaseTransport) setConnected(connected bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.connected = connected
}

// GetMetadata returns the metadata for this transport
func (b *BaseTransport) GetMetadata() map[string]interface{} {
	b.mu.RLock()
	defer b.mu.RUnlock()
	result := make(map[string]interface{})
	for k, v := range b.metadata {
		result[k] = v
	}
	return result
}

// SetMetadata sets metadata for this transport
func (b *BaseTransport) SetMetadata(key string, value interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.metadata[key] = value
}

// ValidateMessage validates that a message conforms to MCP protocol
func ValidateMessage(message interface{}) error {
	if message == nil {
		return fmt.Errorf("message cannot be nil")
	}

	// Check if it's an MCP message
	if mcpMsg, ok := message.(*types.MCPMessage); ok {
		if mcpMsg.ID == "" {
			return fmt.Errorf("MCP message ID is required")
		}
		if mcpMsg.Type == "" {
			return fmt.Errorf("MCP message type is required")
		}
		if mcpMsg.Version == "" {
			return fmt.Errorf("MCP message version is required")
		}
		return nil
	}

	// Check if it's a transport request
	if req, ok := message.(*types.TransportRequest); ok {
		if req.ID == "" {
			return fmt.Errorf("transport request ID is required")
		}
		if req.Method == "" {
			return fmt.Errorf("transport request method is required")
		}
		return nil
	}

	// Check if it's a transport response
	if resp, ok := message.(*types.TransportResponse); ok {
		if resp.ID == "" {
			return fmt.Errorf("transport response ID is required")
		}
		if resp.RequestID == "" {
			return fmt.Errorf("transport response request ID is required")
		}
		return nil
	}

	return fmt.Errorf("unsupported message type: %T", message)
}

// CreateMCPMessage creates a new MCP message with common fields
func CreateMCPMessage(msgType, method, id, version string) *types.MCPMessage {
	return &types.MCPMessage{
		ID:      id,
		Type:    msgType,
		Method:  method,
		Version: version,
		Params:  make(map[string]interface{}),
	}
}

// CreateTransportRequest creates a new transport request
func CreateTransportRequest(id, method, path string, transport types.TransportType) *types.TransportRequest {
	return &types.TransportRequest{
		ID:         id,
		Transport:  transport,
		Method:     method,
		Path:       path,
		Headers:    make(map[string]string),
		Parameters: make(map[string]interface{}),
	}
}

// CreateTransportResponse creates a new transport response
func CreateTransportResponse(id, requestID string, transport types.TransportType, status int) *types.TransportResponse {
	return &types.TransportResponse{
		ID:        id,
		RequestID: requestID,
		Transport: transport,
		Status:    status,
		Headers:   make(map[string]string),
	}
}

// TransportFactory is a factory function type for creating transports
type TransportFactory func(config map[string]interface{}) (types.Transport, error)

// transportRegistry holds registered transport factories
var transportRegistry = make(map[types.TransportType]TransportFactory)
var registryMu sync.RWMutex

// RegisterTransport registers a transport factory for a given transport type
func RegisterTransport(transportType types.TransportType, factory TransportFactory) {
	registryMu.Lock()
	defer registryMu.Unlock()
	transportRegistry[transportType] = factory
}

// CreateTransport creates a transport instance using the registered factory
func CreateTransport(transportType types.TransportType, config map[string]interface{}) (types.Transport, error) {
	registryMu.RLock()
	factory, exists := transportRegistry[transportType]
	registryMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("transport type %s is not registered", transportType)
	}

	return factory(config)
}

// GetRegisteredTransports returns all registered transport types
func GetRegisteredTransports() []types.TransportType {
	registryMu.RLock()
	defer registryMu.RUnlock()

	var transports []types.TransportType
	for transportType := range transportRegistry {
		transports = append(transports, transportType)
	}
	return transports
}

// IsTransportRegistered checks if a transport type is registered
func IsTransportRegistered(transportType types.TransportType) bool {
	registryMu.RLock()
	defer registryMu.RUnlock()
	_, exists := transportRegistry[transportType]
	return exists
}
