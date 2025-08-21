package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"mcp-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebSocketTransport implements WebSocket transport for real-time bidirectional communication
type WebSocketTransport struct {
	*BaseTransport
	conn         *websocket.Conn
	upgrader     websocket.Upgrader
	messageQueue chan *types.WebSocketMessage
	responseMap  map[string]chan *types.MCPMessage
	config       map[string]interface{}
	mu           sync.RWMutex
	done         chan struct{}
	pingTicker   *time.Ticker
	timeout      time.Duration
	bufferSize   int
}

// NewWebSocketTransport creates a new WebSocket transport instance
func NewWebSocketTransport(config map[string]interface{}) (types.Transport, error) {
	transport := &WebSocketTransport{
		BaseTransport: NewBaseTransport(types.TransportTypeWebSocket),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		messageQueue: make(chan *types.WebSocketMessage, 100),
		responseMap:  make(map[string]chan *types.MCPMessage),
		config:       config,
		done:         make(chan struct{}),
		timeout:      60 * time.Second,
		bufferSize:   1024,
	}

	// Configure from config map
	if timeout, ok := config["websocket_timeout"].(time.Duration); ok {
		transport.timeout = timeout
	}

	if bufferSize, ok := config["buffer_size"].(int); ok {
		transport.bufferSize = bufferSize
		transport.upgrader.ReadBufferSize = bufferSize
		transport.upgrader.WriteBufferSize = bufferSize
		transport.messageQueue = make(chan *types.WebSocketMessage, bufferSize)
	}

	return transport, nil
}

// UpgradeHTTP upgrades an HTTP connection to WebSocket
func (w *WebSocketTransport) UpgradeHTTP(writer http.ResponseWriter, request *http.Request) error {
	conn, err := w.upgrader.Upgrade(writer, request, nil)
	if err != nil {
		return fmt.Errorf("failed to upgrade to WebSocket: %w", err)
	}

	w.mu.Lock()
	w.conn = conn
	w.mu.Unlock()

	// Set connection timeouts
	w.conn.SetReadDeadline(time.Now().Add(w.timeout))
	w.conn.SetWriteDeadline(time.Now().Add(w.timeout))

	// Set up ping handler
	w.conn.SetPingHandler(func(appData string) error {
		w.conn.SetReadDeadline(time.Now().Add(w.timeout))
		return w.conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(10*time.Second))
	})

	// Set up pong handler
	w.conn.SetPongHandler(func(appData string) error {
		w.conn.SetReadDeadline(time.Now().Add(w.timeout))
		return nil
	})

	return nil
}

// Connect establishes WebSocket connection
func (w *WebSocketTransport) Connect(ctx context.Context) error {
	if w.conn == nil {
		return fmt.Errorf("WebSocket connection not established, call UpgradeHTTP first")
	}

	w.setConnected(true)

	// Start message handling goroutines
	go w.readPump()
	go w.writePump()
	go w.pingPump()

	return nil
}

// Disconnect closes WebSocket connection
func (w *WebSocketTransport) Disconnect(ctx context.Context) error {
	w.setConnected(false)

	// Signal done to all goroutines
	close(w.done)

	// Stop ping ticker
	if w.pingTicker != nil {
		w.pingTicker.Stop()
	}

	// Close WebSocket connection
	w.mu.Lock()
	if w.conn != nil {
		w.conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			time.Now().Add(time.Second))
		w.conn.Close()
		w.conn = nil
	}
	w.mu.Unlock()

	// Close channels
	close(w.messageQueue)
	for _, ch := range w.responseMap {
		close(ch)
	}
	w.responseMap = make(map[string]chan *types.MCPMessage)

	return nil
}

// SendMessage sends a message via WebSocket
func (w *WebSocketTransport) SendMessage(ctx context.Context, message interface{}) error {
	if !w.IsConnected() {
		return fmt.Errorf("WebSocket not connected")
	}

	// Convert message to WebSocket format
	wsMessage, err := w.convertToWebSocketMessage(message)
	if err != nil {
		return fmt.Errorf("failed to convert message: %w", err)
	}

	// Send to message queue
	select {
	case w.messageQueue <- wsMessage:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-w.done:
		return fmt.Errorf("WebSocket transport closed")
	}
}

// ReceiveMessage receives a message via WebSocket
func (w *WebSocketTransport) ReceiveMessage(ctx context.Context) (interface{}, error) {
	if !w.IsConnected() {
		return nil, fmt.Errorf("WebSocket not connected")
	}

	// For WebSocket, messages are handled asynchronously by readPump
	// This method can be used to wait for specific responses
	return nil, fmt.Errorf("use async message handling for WebSocket transport")
}

// SendMCPMessage sends an MCP message via WebSocket
func (w *WebSocketTransport) SendMCPMessage(ctx context.Context, mcpMessage *types.MCPMessage) error {
	return w.SendMessage(ctx, mcpMessage)
}

// SendMCPRequest sends an MCP request and waits for response
func (w *WebSocketTransport) SendMCPRequest(ctx context.Context, mcpMessage *types.MCPMessage) (*types.MCPMessage, error) {
	if !w.IsConnected() {
		return nil, fmt.Errorf("WebSocket not connected")
	}

	// Create response channel
	responseChan := make(chan *types.MCPMessage, 1)
	w.mu.Lock()
	w.responseMap[mcpMessage.ID] = responseChan
	w.mu.Unlock()

	// Clean up response channel
	defer func() {
		w.mu.Lock()
		delete(w.responseMap, mcpMessage.ID)
		w.mu.Unlock()
		close(responseChan)
	}()

	// Send message
	if err := w.SendMCPMessage(ctx, mcpMessage); err != nil {
		return nil, err
	}

	// Wait for response
	select {
	case response := <-responseChan:
		return response, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-w.done:
		return nil, fmt.Errorf("WebSocket transport closed")
	case <-time.After(w.timeout):
		return nil, fmt.Errorf("request timeout")
	}
}

// readPump handles incoming WebSocket messages
func (w *WebSocketTransport) readPump() {
	defer func() {
		w.Disconnect(context.Background())
	}()

	for {
		select {
		case <-w.done:
			return
		default:
			// Read message
			w.mu.RLock()
			conn := w.conn
			w.mu.RUnlock()

			if conn == nil {
				return
			}

			conn.SetReadDeadline(time.Now().Add(w.timeout))
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					// Log unexpected close error
				}
				return
			}

			// Handle different message types
			switch messageType {
			case websocket.TextMessage:
				w.handleTextMessage(data)
			case websocket.BinaryMessage:
				w.handleBinaryMessage(data)
			case websocket.CloseMessage:
				return
			}
		}
	}
}

// writePump handles outgoing WebSocket messages
func (w *WebSocketTransport) writePump() {
	defer func() {
		w.Disconnect(context.Background())
	}()

	for {
		select {
		case message, ok := <-w.messageQueue:
			if !ok {
				return
			}

			w.mu.RLock()
			conn := w.conn
			w.mu.RUnlock()

			if conn == nil {
				return
			}

			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			// Send message based on type
			switch message.Type {
			case types.WebSocketMessageTypeText:
				if err := conn.WriteMessage(websocket.TextMessage, w.serializeMessage(message.Data)); err != nil {
					return
				}
			case types.WebSocketMessageTypeBinary:
				if data, ok := message.Data.([]byte); ok {
					if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
						return
					}
				}
			case types.WebSocketMessageTypePing:
				if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second)); err != nil {
					return
				}
			case types.WebSocketMessageTypePong:
				if err := conn.WriteControl(websocket.PongMessage, nil, time.Now().Add(time.Second)); err != nil {
					return
				}
			}

		case <-w.done:
			return
		}
	}
}

// pingPump sends periodic ping messages
func (w *WebSocketTransport) pingPump() {
	w.pingTicker = time.NewTicker(54 * time.Second) // Slightly less than 60s timeout
	defer w.pingTicker.Stop()

	for {
		select {
		case <-w.pingTicker.C:
			pingMessage := &types.WebSocketMessage{
				Type:      types.WebSocketMessageTypePing,
				Data:      nil,
				Timestamp: time.Now(),
			}
			select {
			case w.messageQueue <- pingMessage:
			case <-w.done:
				return
			}
		case <-w.done:
			return
		}
	}
}

// handleTextMessage processes incoming text messages
func (w *WebSocketTransport) handleTextMessage(data []byte) {
	// Try to parse as MCP message
	var mcpMessage types.MCPMessage
	if err := json.Unmarshal(data, &mcpMessage); err != nil {
		// Not an MCP message, handle as generic text
		return
	}

	// Check if this is a response to a pending request
	w.mu.RLock()
	responseChan, exists := w.responseMap[mcpMessage.ID]
	w.mu.RUnlock()

	if exists {
		// Send response to waiting goroutine
		select {
		case responseChan <- &mcpMessage:
		default:
			// Channel full or closed
		}
	}

	// TODO: Handle notifications and other message types
}

// handleBinaryMessage processes incoming binary messages
func (w *WebSocketTransport) handleBinaryMessage(data []byte) {
	// Handle binary data (could be file transfers, etc.)
	// For now, just log that we received binary data
}

// convertToWebSocketMessage converts various message types to WebSocket message format
func (w *WebSocketTransport) convertToWebSocketMessage(message interface{}) (*types.WebSocketMessage, error) {
	switch msg := message.(type) {
	case *types.MCPMessage:
		return &types.WebSocketMessage{
			Type:      types.WebSocketMessageTypeText,
			Data:      msg,
			Timestamp: time.Now(),
		}, nil
	case *types.WebSocketMessage:
		return msg, nil
	case *types.TransportRequest:
		mcpMessage := &types.MCPMessage{
			ID:      msg.ID,
			Type:    types.MCPMessageTypeRequest,
			Method:  msg.Method,
			Version: "2024-11-05",
			Params:  msg.Parameters,
		}
		return &types.WebSocketMessage{
			Type:      types.WebSocketMessageTypeText,
			Data:      mcpMessage,
			Timestamp: time.Now(),
		}, nil
	case string:
		return &types.WebSocketMessage{
			Type:      types.WebSocketMessageTypeText,
			Data:      msg,
			Timestamp: time.Now(),
		}, nil
	case []byte:
		return &types.WebSocketMessage{
			Type:      types.WebSocketMessageTypeBinary,
			Data:      msg,
			Timestamp: time.Now(),
		}, nil
	default:
		return &types.WebSocketMessage{
			Type:      types.WebSocketMessageTypeText,
			Data:      message,
			Timestamp: time.Now(),
		}, nil
	}
}

// serializeMessage serializes message data for transmission
func (w *WebSocketTransport) serializeMessage(data interface{}) []byte {
	switch d := data.(type) {
	case []byte:
		return d
	case string:
		return []byte(d)
	default:
		jsonData, err := json.Marshal(data)
		if err != nil {
			return []byte(fmt.Sprintf(`{"error": "serialization failed: %s"}`, err.Error()))
		}
		return jsonData
	}
}

// Helper methods for MCP operations

// CallTool calls an MCP tool via WebSocket
func (w *WebSocketTransport) CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (*types.MCPMessage, error) {
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

	return w.SendMCPRequest(ctx, mcpMessage)
}

// ListTools lists available MCP tools via WebSocket
func (w *WebSocketTransport) ListTools(ctx context.Context) (*types.MCPMessage, error) {
	mcpMessage := &types.MCPMessage{
		ID:      uuid.New().String(),
		Type:    types.MCPMessageTypeRequest,
		Method:  types.MCPMethodListTools,
		Version: "2024-11-05",
		Params:  make(map[string]interface{}),
	}

	return w.SendMCPRequest(ctx, mcpMessage)
}

// GetConfig returns the transport configuration
func (w *WebSocketTransport) GetConfig() map[string]interface{} {
	return w.config
}

// GetConnection returns the underlying WebSocket connection
func (w *WebSocketTransport) GetConnection() *websocket.Conn {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.conn
}

// IsHealthy checks if the WebSocket connection is healthy
func (w *WebSocketTransport) IsHealthy() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.conn != nil && w.IsConnected()
}

// GetMetrics returns WebSocket transport metrics
func (w *WebSocketTransport) GetMetrics() map[string]interface{} {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return map[string]interface{}{
		"connected":          w.IsConnected(),
		"message_queue_size": len(w.messageQueue),
		"pending_responses":  len(w.responseMap),
		"timeout":            w.timeout,
		"buffer_size":        w.bufferSize,
		"session_id":         w.GetSessionID(),
	}
}

// init registers the WebSocket transport factory
func init() {
	RegisterTransport(types.TransportTypeWebSocket, NewWebSocketTransport)
}
