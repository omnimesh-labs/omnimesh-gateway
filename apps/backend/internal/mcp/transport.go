package mcp

import (
	"context"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"
)

// Transport defines the interface for different MCP transport mechanisms
type Transport interface {
	Connect(ctx context.Context, config TransportConfig) (Connection, error)
	Type() string
}

// Connection represents an active connection to an MCP server
type Connection interface {
	Send(ctx context.Context, message []byte) error
	Receive(ctx context.Context) ([]byte, error)
	Close() error
	IsConnected() bool
}

// TransportConfig holds configuration for different transport types
type TransportConfig struct {
	Type        string            `json:"type"`        // "stdio", "http", "websocket"
	Command     string            `json:"command"`     // For stdio: command to execute
	Args        []string          `json:"args"`        // For stdio: command arguments
	URL         string            `json:"url"`         // For http/ws: connection URL
	Headers     map[string]string `json:"headers"`     // For http/ws: custom headers
	Environment map[string]string `json:"environment"` // For stdio: environment variables
	WorkingDir  string            `json:"working_dir"` // For stdio: working directory
}

// TransportManager manages different transport types
type TransportManager struct {
	transports map[string]Transport
}

// NewTransportManager creates a new transport manager
func NewTransportManager() *TransportManager {
	tm := &TransportManager{
		transports: make(map[string]Transport),
	}

	// Register built-in transports
	tm.RegisterTransport(&StdioTransport{})

	return tm
}

// RegisterTransport registers a new transport type
func (tm *TransportManager) RegisterTransport(transport Transport) {
	tm.transports[transport.Type()] = transport
}

// CreateConnection creates a connection using the specified transport
func (tm *TransportManager) CreateConnection(ctx context.Context, config TransportConfig) (Connection, error) {
	transport, exists := tm.transports[config.Type]
	if !exists {
		return nil, &types.MCPError{
			Code:    -32601,
			Message: "unsupported transport type",
			Data:    map[string]interface{}{"type": config.Type},
		}
	}

	return transport.Connect(ctx, config)
}

// GetSupportedTransports returns list of supported transport types
func (tm *TransportManager) GetSupportedTransports() []string {
	var types []string
	for transportType := range tm.transports {
		types = append(types, transportType)
	}
	return types
}
