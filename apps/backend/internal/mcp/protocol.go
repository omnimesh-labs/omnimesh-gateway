package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"mcp-gateway/apps/backend/internal/types"
)

// MCPClient implements the MCP protocol over a transport connection
type MCPClient struct {
	connection Connection
	transport  Transport

	// Request tracking
	pendingRequests sync.Map // map[string]chan *JSONRPCResponse
	requestID       int64

	// Server capabilities
	serverCapabilities map[string]interface{}
	clientInfo         ClientInfo

	// State management
	initialized bool
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc

	// Message handling
	messageHandler func(message []byte)
}

// ClientInfo represents information about the MCP client
type ClientInfo struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Capabilities map[string]interface{} `json:"capabilities"`
}

// JSONRPCRequest represents a JSON-RPC request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      string      `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *types.MCPError `json:"error,omitempty"`
	ID      string      `json:"id"`
}

// InitializeParams represents MCP initialize request parameters
type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
}

// InitializeResult represents MCP initialize response
type InitializeResult struct {
	ProtocolVersion  string                 `json:"protocolVersion"`
	Capabilities     map[string]interface{} `json:"capabilities"`
	ServerInfo       ServerInfo             `json:"serverInfo"`
}

// ServerInfo represents information about the MCP server
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ToolsListResult represents the result of tools/list
type ToolsListResult struct {
	Tools []types.Tool `json:"tools"`
}

// ToolsCallParams represents parameters for tools/call
type ToolsCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// ToolsCallResult represents the result of tools/call
type ToolsCallResult struct {
	Content []ToolCallContent `json:"content"`
	IsError bool             `json:"isError,omitempty"`
}

// ToolCallContent represents content returned by a tool call
type ToolCallContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Data string `json:"data,omitempty"`
}

// NewMCPClient creates a new MCP client with the given transport
func NewMCPClient(transport Transport) *MCPClient {
	ctx, cancel := context.WithCancel(context.Background())

	return &MCPClient{
		transport:          transport,
		pendingRequests:    sync.Map{},
		serverCapabilities: make(map[string]interface{}),
		ctx:                ctx,
		cancel:             cancel,
	}
}

// Connect establishes a connection to the MCP server
func (c *MCPClient) Connect(ctx context.Context, config TransportConfig, clientInfo ClientInfo) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Store client info
	c.clientInfo = clientInfo

	// Create connection
	conn, err := c.transport.Connect(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	c.connection = conn

	// Start message handling
	go c.handleMessages()

	// Initialize the MCP session
	if err := c.initialize(ctx); err != nil {
		c.connection.Close()
		return fmt.Errorf("failed to initialize: %w", err)
	}

	return nil
}

// Initialize sends the initialize request per MCP spec
func (c *MCPClient) initialize(ctx context.Context) error {
	params := InitializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities:    c.clientInfo.Capabilities,
		ClientInfo:      c.clientInfo,
	}

	var result InitializeResult
	if err := c.sendRequest(ctx, "initialize", params, &result); err != nil {
		return err
	}

	// Store server capabilities
	c.serverCapabilities = result.Capabilities
	c.initialized = true

	return nil
}

// ListTools sends a tools/list request and returns the available tools
func (c *MCPClient) ListTools(ctx context.Context) ([]types.Tool, error) {
	c.mu.RLock()
	if !c.initialized {
		c.mu.RUnlock()
		return nil, fmt.Errorf("client not initialized")
	}
	c.mu.RUnlock()

	var result ToolsListResult
	if err := c.sendRequest(ctx, "tools/list", map[string]interface{}{}, &result); err != nil {
		return nil, err
	}

	return result.Tools, nil
}

// CallTool sends a tools/call request to execute a tool
func (c *MCPClient) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (interface{}, error) {
	c.mu.RLock()
	if !c.initialized {
		c.mu.RUnlock()
		return nil, fmt.Errorf("client not initialized")
	}
	c.mu.RUnlock()

	params := ToolsCallParams{
		Name:      name,
		Arguments: arguments,
	}

	var result ToolsCallResult
	if err := c.sendRequest(ctx, "tools/call", params, &result); err != nil {
		return nil, err
	}

	// Check if the tool call resulted in an error
	if result.IsError {
		return nil, fmt.Errorf("tool execution failed: %v", result.Content)
	}

	return result, nil
}

// Close closes the MCP client connection
func (c *MCPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cancel != nil {
		c.cancel()
	}

	if c.connection != nil {
		return c.connection.Close()
	}

	c.initialized = false
	return nil
}

// IsConnected returns whether the client is connected and initialized
func (c *MCPClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.initialized && c.connection != nil && c.connection.IsConnected()
}

// generateRequestID generates a unique request ID
func (c *MCPClient) generateRequestID() string {
	id := atomic.AddInt64(&c.requestID, 1)
	return fmt.Sprintf("req-%d", id)
}

// sendRequest sends a JSON-RPC request and waits for the response
func (c *MCPClient) sendRequest(ctx context.Context, method string, params interface{}, result interface{}) error {
	requestID := c.generateRequestID()

	// Create response channel
	responseChan := make(chan *JSONRPCResponse, 1)
	c.pendingRequests.Store(requestID, responseChan)
	defer c.pendingRequests.Delete(requestID)

	// Create request
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      requestID,
	}

	// Serialize request
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request
	if err := c.connection.Send(ctx, requestBytes); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	// Wait for response
	select {
	case <-ctx.Done():
		return ctx.Err()
	case response := <-responseChan:
		if response.Error != nil {
			return response.Error
		}

		// Deserialize result if provided
		if result != nil && response.Result != nil {
			resultBytes, err := json.Marshal(response.Result)
			if err != nil {
				return fmt.Errorf("failed to marshal result: %w", err)
			}

			if err := json.Unmarshal(resultBytes, result); err != nil {
				return fmt.Errorf("failed to unmarshal result: %w", err)
			}
		}

		return nil
	case <-time.After(30 * time.Second):
		return fmt.Errorf("request timeout")
	}
}

// handleMessages handles incoming messages from the server
func (c *MCPClient) handleMessages() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			message, err := c.connection.Receive(c.ctx)
			if err != nil {
				if c.ctx.Err() == nil {
					// Only log if context wasn't cancelled
					fmt.Printf("Error receiving message: %v\n", err)
				}
				return
			}

			c.handleMessage(message)
		}
	}
}

// handleMessage processes a single incoming message
func (c *MCPClient) handleMessage(messageBytes []byte) {
	var response JSONRPCResponse
	if err := json.Unmarshal(messageBytes, &response); err != nil {
		fmt.Printf("Failed to unmarshal response: %v\n", err)
		return
	}

	// Handle response to pending request
	if response.ID != "" {
		if responseChan, ok := c.pendingRequests.Load(response.ID); ok {
			select {
			case responseChan.(chan *JSONRPCResponse) <- &response:
			case <-time.After(1 * time.Second):
				// Drop response if channel is full
			}
		}
		return
	}

	// Handle notifications or other messages
	if c.messageHandler != nil {
		c.messageHandler(messageBytes)
	}
}

// SetMessageHandler sets a handler for non-response messages
func (c *MCPClient) SetMessageHandler(handler func(message []byte)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messageHandler = handler
}
