package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
)

// JSONRPCTransport implements JSON-RPC over HTTP transport
type JSONRPCTransport struct {
	*BaseTransport
	client       *http.Client
	requestQueue chan *types.MCPMessage
	responseMap  map[string]chan *types.MCPMessage
	config       map[string]interface{}
	endpoint     string
	timeout      time.Duration
}

// JSONRPCRequest represents a JSON-RPC request
type JSONRPCRequest struct {
	Params  interface{} `json:"params,omitempty"`
	ID      string      `json:"id"`
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
}

// JSONRPCResponse represents a JSON-RPC response
type JSONRPCResponse struct {
	Result  map[string]interface{} `json:"result,omitempty"`
	Error   *JSONRPCError          `json:"error,omitempty"`
	ID      string                 `json:"id"`
	JSONRPC string                 `json:"jsonrpc"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message"`
	Code    int         `json:"code"`
}

// NewJSONRPCTransport creates a new JSON-RPC transport instance
func NewJSONRPCTransport(config map[string]interface{}) (types.Transport, error) {
	transport := &JSONRPCTransport{
		BaseTransport: NewBaseTransport(types.TransportTypeHTTP),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		timeout:      30 * time.Second,
		requestQueue: make(chan *types.MCPMessage, 100),
		responseMap:  make(map[string]chan *types.MCPMessage),
		config:       config,
	}

	// Configure from config map
	if endpoint, ok := config["endpoint"].(string); ok {
		transport.endpoint = endpoint
	} else {
		transport.endpoint = "/rpc" // Default endpoint
	}

	if timeout, ok := config["timeout"].(time.Duration); ok {
		transport.timeout = timeout
		transport.client.Timeout = timeout
	}

	return transport, nil
}

// Connect establishes connection (for HTTP, this is a no-op)
func (j *JSONRPCTransport) Connect(ctx context.Context) error {
	j.setConnected(true)
	return nil
}

// Disconnect closes connection (for HTTP, this is a no-op)
func (j *JSONRPCTransport) Disconnect(ctx context.Context) error {
	j.setConnected(false)
	close(j.requestQueue)

	// Close all pending response channels
	for _, ch := range j.responseMap {
		close(ch)
	}
	j.responseMap = make(map[string]chan *types.MCPMessage)

	return nil
}

// SendMessage sends a message via JSON-RPC
func (j *JSONRPCTransport) SendMessage(ctx context.Context, message interface{}) error {
	if !j.IsConnected() {
		return fmt.Errorf("transport not connected")
	}

	// Convert message to MCP format
	mcpMessage, err := j.convertToMCPMessage(message)
	if err != nil {
		return fmt.Errorf("failed to convert message: %w", err)
	}

	// Create JSON-RPC request
	rpcRequest := &JSONRPCRequest{
		ID:      mcpMessage.ID,
		JSONRPC: "2.0",
		Method:  mcpMessage.Method,
		Params:  mcpMessage.Params,
	}

	// Marshal request
	requestBody, err := json.Marshal(rpcRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", j.endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add session ID if available
	if sessionID := j.GetSessionID(); sessionID != "" {
		req.Header.Set("X-Session-ID", sessionID)
	}

	resp, err := j.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// For synchronous JSON-RPC, we don't need to handle the response here
	// as it will be handled in ReceiveMessage or by the caller
	return nil
}

// ReceiveMessage receives a message via JSON-RPC (used for responses)
func (j *JSONRPCTransport) ReceiveMessage(ctx context.Context) (interface{}, error) {
	if !j.IsConnected() {
		return nil, fmt.Errorf("transport not connected")
	}

	// For HTTP/JSON-RPC, this is typically called as part of SendMessage
	// In a real implementation, you might want to handle this differently
	// For now, we'll return a placeholder
	return nil, fmt.Errorf("ReceiveMessage not implemented for synchronous JSON-RPC")
}

// SendRequest sends a JSON-RPC request and waits for response
func (j *JSONRPCTransport) SendRequest(ctx context.Context, method string, params interface{}) (*JSONRPCResponse, error) {
	if !j.IsConnected() {
		return nil, fmt.Errorf("transport not connected")
	}

	// Create request
	requestID := uuid.New().String()
	rpcRequest := &JSONRPCRequest{
		ID:      requestID,
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	// Marshal request
	requestBody, err := json.Marshal(rpcRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", j.endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add session ID if available
	if sessionID := j.GetSessionID(); sessionID != "" {
		req.Header.Set("X-Session-ID", sessionID)
	}

	resp, err := j.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse JSON-RPC response
	var rpcResponse JSONRPCResponse
	if err := json.Unmarshal(responseBody, &rpcResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &rpcResponse, nil
}

// SendMCPRequest sends an MCP request via JSON-RPC
func (j *JSONRPCTransport) SendMCPRequest(ctx context.Context, mcpMessage *types.MCPMessage) (*types.MCPMessage, error) {
	// Send JSON-RPC request
	rpcResponse, err := j.SendRequest(ctx, mcpMessage.Method, mcpMessage.Params)
	if err != nil {
		return nil, err
	}

	// Convert JSON-RPC response to MCP message
	mcpResponse := &types.MCPMessage{
		ID:      rpcResponse.ID,
		Type:    types.MCPMessageTypeResponse,
		Version: mcpMessage.Version,
	}

	if rpcResponse.Error != nil {
		mcpResponse.Error = &types.MCPError{
			Code:    rpcResponse.Error.Code,
			Message: rpcResponse.Error.Message,
			Data:    rpcResponse.Error.Data,
		}
	} else {
		mcpResponse.Result = rpcResponse.Result
	}

	return mcpResponse, nil
}

// CallTool calls an MCP tool via JSON-RPC
func (j *JSONRPCTransport) CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (*types.MCPMessage, error) {
	mcpMessage := &types.MCPMessage{
		ID:      uuid.New().String(),
		Type:    types.MCPMessageTypeRequest,
		Method:  types.MCPMethodCallTool,
		Version: "2024-11-05",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": arguments,
		},
	}

	return j.SendMCPRequest(ctx, mcpMessage)
}

// ListTools lists available MCP tools via JSON-RPC
func (j *JSONRPCTransport) ListTools(ctx context.Context) (*types.MCPMessage, error) {
	mcpMessage := &types.MCPMessage{
		ID:      uuid.New().String(),
		Type:    types.MCPMessageTypeRequest,
		Method:  types.MCPMethodListTools,
		Version: "2024-11-05",
		Params:  make(map[string]interface{}),
	}

	return j.SendMCPRequest(ctx, mcpMessage)
}

// ListResources lists available MCP resources via JSON-RPC
func (j *JSONRPCTransport) ListResources(ctx context.Context) (*types.MCPMessage, error) {
	mcpMessage := &types.MCPMessage{
		ID:      uuid.New().String(),
		Type:    types.MCPMessageTypeRequest,
		Method:  types.MCPMethodListResources,
		Version: "2024-11-05",
		Params:  make(map[string]interface{}),
	}

	return j.SendMCPRequest(ctx, mcpMessage)
}

// ReadResource reads an MCP resource via JSON-RPC
func (j *JSONRPCTransport) ReadResource(ctx context.Context, uri string) (*types.MCPMessage, error) {
	mcpMessage := &types.MCPMessage{
		ID:      uuid.New().String(),
		Type:    types.MCPMessageTypeRequest,
		Method:  types.MCPMethodReadResource,
		Version: "2024-11-05",
		Params: map[string]interface{}{
			"uri": uri,
		},
	}

	return j.SendMCPRequest(ctx, mcpMessage)
}

// ListPrompts lists available MCP prompts via JSON-RPC
func (j *JSONRPCTransport) ListPrompts(ctx context.Context) (*types.MCPMessage, error) {
	mcpMessage := &types.MCPMessage{
		ID:      uuid.New().String(),
		Type:    types.MCPMessageTypeRequest,
		Method:  types.MCPMethodListPrompts,
		Version: "2024-11-05",
		Params:  make(map[string]interface{}),
	}

	return j.SendMCPRequest(ctx, mcpMessage)
}

// GetPrompt gets an MCP prompt via JSON-RPC
func (j *JSONRPCTransport) GetPrompt(ctx context.Context, name string, arguments map[string]interface{}) (*types.MCPMessage, error) {
	mcpMessage := &types.MCPMessage{
		ID:      uuid.New().String(),
		Type:    types.MCPMessageTypeRequest,
		Method:  types.MCPMethodGetPrompt,
		Version: "2024-11-05",
		Params: map[string]interface{}{
			"name":      name,
			"arguments": arguments,
		},
	}

	return j.SendMCPRequest(ctx, mcpMessage)
}

// convertToMCPMessage converts various message types to MCP message format
func (j *JSONRPCTransport) convertToMCPMessage(message interface{}) (*types.MCPMessage, error) {
	switch msg := message.(type) {
	case *types.MCPMessage:
		return msg, nil
	case *types.TransportRequest:
		mcpMessage := &types.MCPMessage{
			ID:      msg.ID,
			Type:    types.MCPMessageTypeRequest,
			Method:  msg.Method,
			Version: "2024-11-05",
			Params:  msg.Parameters,
		}
		return mcpMessage, nil
	case map[string]interface{}:
		// Try to parse as MCP message
		jsonData, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}
		var mcpMessage types.MCPMessage
		if err := json.Unmarshal(jsonData, &mcpMessage); err != nil {
			return nil, err
		}
		return &mcpMessage, nil
	default:
		return nil, fmt.Errorf("unsupported message type: %T", message)
	}
}

// GetConfig returns the transport configuration
func (j *JSONRPCTransport) GetConfig() map[string]interface{} {
	return j.config
}

// SetEndpoint sets the JSON-RPC endpoint
func (j *JSONRPCTransport) SetEndpoint(endpoint string) {
	j.endpoint = endpoint
}

// GetEndpoint returns the JSON-RPC endpoint
func (j *JSONRPCTransport) GetEndpoint() string {
	return j.endpoint
}

// SendBatchRequest sends multiple JSON-RPC requests in a single batch
func (j *JSONRPCTransport) SendBatchRequest(ctx context.Context, requests []*JSONRPCRequest) ([]*JSONRPCResponse, error) {
	if !j.IsConnected() {
		return nil, fmt.Errorf("transport not connected")
	}

	if len(requests) == 0 {
		return []*JSONRPCResponse{}, nil
	}

	// Marshal batch request
	requestBody, err := json.Marshal(requests)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch request: %w", err)
	}

	// Send HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", j.endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add session ID if available
	if sessionID := j.GetSessionID(); sessionID != "" {
		req.Header.Set("X-Session-ID", sessionID)
	}

	resp, err := j.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse JSON-RPC batch response
	var rpcResponses []*JSONRPCResponse
	if err := json.Unmarshal(responseBody, &rpcResponses); err != nil {
		return nil, fmt.Errorf("failed to unmarshal batch response: %w", err)
	}

	return rpcResponses, nil
}

// SendMCPBatchRequest sends multiple MCP requests in a JSON-RPC batch
func (j *JSONRPCTransport) SendMCPBatchRequest(ctx context.Context, mcpMessages []*types.MCPMessage) ([]*types.MCPMessage, error) {
	// Convert MCP messages to JSON-RPC requests
	var rpcRequests []*JSONRPCRequest
	for _, mcpMsg := range mcpMessages {
		rpcReq := &JSONRPCRequest{
			ID:      mcpMsg.ID,
			JSONRPC: "2.0",
			Method:  mcpMsg.Method,
			Params:  mcpMsg.Params,
		}
		rpcRequests = append(rpcRequests, rpcReq)
	}

	// Send batch request
	rpcResponses, err := j.SendBatchRequest(ctx, rpcRequests)
	if err != nil {
		return nil, err
	}

	// Convert JSON-RPC responses to MCP messages
	var mcpResponses []*types.MCPMessage
	for i, rpcResp := range rpcResponses {
		var version string
		if i < len(mcpMessages) {
			version = mcpMessages[i].Version
		} else {
			version = "2024-11-05"
		}

		mcpResponse := &types.MCPMessage{
			ID:      rpcResp.ID,
			Type:    types.MCPMessageTypeResponse,
			Version: version,
		}

		if rpcResp.Error != nil {
			mcpResponse.Error = &types.MCPError{
				Code:    rpcResp.Error.Code,
				Message: rpcResp.Error.Message,
				Data:    rpcResp.Error.Data,
			}
		} else {
			mcpResponse.Result = rpcResp.Result
		}

		mcpResponses = append(mcpResponses, mcpResponse)
	}

	return mcpResponses, nil
}

// ValidateJSONRPCRequest validates a JSON-RPC request
func ValidateJSONRPCRequest(req *JSONRPCRequest) error {
	if req.JSONRPC != "2.0" {
		return fmt.Errorf("invalid JSON-RPC version: %s, expected 2.0", req.JSONRPC)
	}
	if req.ID == "" {
		return fmt.Errorf("missing request ID")
	}
	if req.Method == "" {
		return fmt.Errorf("missing method")
	}
	return nil
}

// ValidateJSONRPCResponse validates a JSON-RPC response
func ValidateJSONRPCResponse(resp *JSONRPCResponse) error {
	if resp.JSONRPC != "2.0" {
		return fmt.Errorf("invalid JSON-RPC version: %s, expected 2.0", resp.JSONRPC)
	}
	if resp.ID == "" {
		return fmt.Errorf("missing response ID")
	}
	if resp.Result == nil && resp.Error == nil {
		return fmt.Errorf("response must have either result or error")
	}
	if resp.Result != nil && resp.Error != nil {
		return fmt.Errorf("response cannot have both result and error")
	}
	return nil
}

// GetConnectionMetrics returns detailed metrics about the JSON-RPC transport
func (j *JSONRPCTransport) GetConnectionMetrics() map[string]interface{} {
	return map[string]interface{}{
		"transport_type":    types.TransportTypeHTTP,
		"endpoint":          j.endpoint,
		"timeout":           j.timeout,
		"connected":         j.IsConnected(),
		"session_id":        j.GetSessionID(),
		"pending_requests":  len(j.requestQueue),
		"pending_responses": len(j.responseMap),
		"client_timeout":    j.client.Timeout,
	}
}

// init registers the JSON-RPC transport factory
func init() {
	RegisterTransport(types.TransportTypeHTTP, NewJSONRPCTransport)
}
