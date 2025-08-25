package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"mcp-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
)

// SSETransport implements Server-Sent Events transport for real-time streaming
type SSETransport struct {
	writer  http.ResponseWriter
	flusher http.Flusher
	*BaseTransport
	request         *http.Request
	eventQueue      chan *types.SSEEvent
	config          map[string]interface{}
	done            chan struct{}
	keepAliveTicker *time.Ticker
	lastEventID     string
	keepAlive       time.Duration
	bufferSize      int
	mu              sync.RWMutex
}

// NewSSETransport creates a new SSE transport instance
func NewSSETransport(config map[string]interface{}) (types.Transport, error) {
	transport := &SSETransport{
		BaseTransport: NewBaseTransport(types.TransportTypeSSE),
		eventQueue:    make(chan *types.SSEEvent, 100),
		config:        config,
		done:          make(chan struct{}),
		keepAlive:     30 * time.Second,
		bufferSize:    100,
	}

	// Configure from config map
	if keepAlive, ok := config["sse_keep_alive"].(time.Duration); ok {
		transport.keepAlive = keepAlive
	}

	if bufferSize, ok := config["buffer_size"].(int); ok {
		transport.bufferSize = bufferSize
		transport.eventQueue = make(chan *types.SSEEvent, bufferSize)
	}

	return transport, nil
}

// SetupSSE sets up SSE connection with HTTP response writer
func (s *SSETransport) SetupSSE(writer http.ResponseWriter, request *http.Request) error {
	// Check if the response writer supports flushing
	flusher, ok := writer.(http.Flusher)
	if !ok {
		return fmt.Errorf("response writer does not support flushing")
	}

	s.mu.Lock()
	s.writer = writer
	s.request = request
	s.flusher = flusher
	s.mu.Unlock()

	// Set SSE headers
	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// Get last event ID from request headers if present
	if lastEventID := request.Header.Get("Last-Event-ID"); lastEventID != "" {
		s.lastEventID = lastEventID
	}

	return nil
}

// Connect establishes SSE connection
func (s *SSETransport) Connect(ctx context.Context) error {
	if s.writer == nil {
		return fmt.Errorf("SSE not set up, call SetupSSE first")
	}

	s.setConnected(true)

	// Send connection established event
	connectEvent := &types.SSEEvent{
		ID:        uuid.New().String(),
		Event:     "connected",
		Data:      map[string]interface{}{"status": "connected", "session_id": s.GetSessionID()},
		Timestamp: time.Now(),
	}

	// Start event streaming
	go s.eventLoop()

	// Send initial connection event
	return s.SendEvent(ctx, connectEvent)
}

// Disconnect closes SSE connection
func (s *SSETransport) Disconnect(ctx context.Context) error {
	s.setConnected(false)

	// Send disconnect event
	disconnectEvent := &types.SSEEvent{
		ID:        uuid.New().String(),
		Event:     "disconnected",
		Data:      map[string]interface{}{"status": "disconnected"},
		Timestamp: time.Now(),
	}

	s.SendEvent(ctx, disconnectEvent)

	// Signal done to event loop
	close(s.done)

	// Stop keep-alive ticker
	if s.keepAliveTicker != nil {
		s.keepAliveTicker.Stop()
	}

	// Clear writer and flusher to prevent race conditions
	s.mu.Lock()
	s.writer = nil
	s.flusher = nil
	s.mu.Unlock()

	// Close event queue
	close(s.eventQueue)

	return nil
}

// SendMessage sends a message via SSE
func (s *SSETransport) SendMessage(ctx context.Context, message interface{}) error {
	if !s.IsConnected() {
		return fmt.Errorf("SSE not connected")
	}

	// Convert message to SSE event
	event, err := s.convertToSSEEvent(message)
	if err != nil {
		return fmt.Errorf("failed to convert message to SSE event: %w", err)
	}

	return s.SendEvent(ctx, event)
}

// ReceiveMessage receives a message via SSE (not applicable for SSE, which is unidirectional)
func (s *SSETransport) ReceiveMessage(ctx context.Context) (interface{}, error) {
	return nil, fmt.Errorf("SSE is unidirectional, use SendMessage for server-to-client communication")
}

// SendEvent sends an SSE event
func (s *SSETransport) SendEvent(ctx context.Context, event *types.SSEEvent) error {
	if !s.IsConnected() {
		return fmt.Errorf("SSE not connected")
	}

	// Add timestamp if not set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Add ID if not set
	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	// Send to event queue
	select {
	case s.eventQueue <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-s.done:
		return fmt.Errorf("SSE transport closed")
	}
}

// SendMCPEvent sends an MCP message as SSE event
func (s *SSETransport) SendMCPEvent(ctx context.Context, mcpMessage *types.MCPMessage) error {
	event := &types.SSEEvent{
		ID:        mcpMessage.ID,
		Event:     "mcp_message",
		Data:      mcpMessage,
		Timestamp: time.Now(),
	}

	return s.SendEvent(ctx, event)
}

// BroadcastEvent sends an event to all connected SSE clients (this would be handled at the manager level)
func (s *SSETransport) BroadcastEvent(ctx context.Context, event *types.SSEEvent) error {
	return s.SendEvent(ctx, event)
}

// eventLoop handles the SSE event streaming
func (s *SSETransport) eventLoop() {
	// Set up keep-alive ticker
	s.keepAliveTicker = time.NewTicker(s.keepAlive)
	defer s.keepAliveTicker.Stop()

	for {
		select {
		case event, ok := <-s.eventQueue:
			if !ok {
				return
			}
			s.writeEvent(event)

		case <-s.keepAliveTicker.C:
			// Send keep-alive comment
			s.writeComment("keep-alive")

		case <-s.done:
			return

		case <-s.request.Context().Done():
			// Client disconnected
			return
		}
	}
}

// writeEvent writes an SSE event to the response stream
func (s *SSETransport) writeEvent(event *types.SSEEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.writer == nil || s.flusher == nil {
		return
	}

	// Add event to history for replay
	sseEventHistory.addEvent(event)

	// Write event ID
	if event.ID != "" {
		fmt.Fprintf(s.writer, "id: %s\n", event.ID)
		s.lastEventID = event.ID
	}

	// Write event type
	if event.Event != "" {
		fmt.Fprintf(s.writer, "event: %s\n", event.Event)
	}

	// Write retry time
	if event.Retry > 0 {
		fmt.Fprintf(s.writer, "retry: %d\n", event.Retry)
	}

	// Write data
	dataStr := s.serializeEventData(event.Data)
	for _, line := range strings.Split(dataStr, "\n") {
		fmt.Fprintf(s.writer, "data: %s\n", line)
	}

	// End event with empty line
	fmt.Fprintf(s.writer, "\n")

	// Flush the response
	s.flusher.Flush()
}

// writeComment writes an SSE comment (for keep-alive)
func (s *SSETransport) writeComment(comment string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.writer == nil || s.flusher == nil {
		return
	}

	fmt.Fprintf(s.writer, ": %s\n\n", comment)
	s.flusher.Flush()
}

// serializeEventData serializes event data for SSE transmission
func (s *SSETransport) serializeEventData(data interface{}) string {
	switch d := data.(type) {
	case string:
		return d
	case []byte:
		return string(d)
	default:
		jsonData, err := json.Marshal(data)
		if err != nil {
			return fmt.Sprintf(`{"error": "serialization failed: %s"}`, err.Error())
		}
		return string(jsonData)
	}
}

// convertToSSEEvent converts various message types to SSE event format
func (s *SSETransport) convertToSSEEvent(message interface{}) (*types.SSEEvent, error) {
	switch msg := message.(type) {
	case *types.SSEEvent:
		return msg, nil
	case *types.MCPMessage:
		return &types.SSEEvent{
			ID:        msg.ID,
			Event:     "mcp_message",
			Data:      msg,
			Timestamp: time.Now(),
		}, nil
	case *types.TransportRequest:
		return &types.SSEEvent{
			ID:        msg.ID,
			Event:     "transport_request",
			Data:      msg,
			Timestamp: time.Now(),
		}, nil
	case *types.TransportResponse:
		return &types.SSEEvent{
			ID:        msg.ID,
			Event:     "transport_response",
			Data:      msg,
			Timestamp: time.Now(),
		}, nil
	case map[string]interface{}:
		return &types.SSEEvent{
			ID:        uuid.New().String(),
			Event:     "data",
			Data:      msg,
			Timestamp: time.Now(),
		}, nil
	case string:
		return &types.SSEEvent{
			ID:        uuid.New().String(),
			Event:     "message",
			Data:      msg,
			Timestamp: time.Now(),
		}, nil
	default:
		return &types.SSEEvent{
			ID:        uuid.New().String(),
			Event:     "data",
			Data:      message,
			Timestamp: time.Now(),
		}, nil
	}
}

// Helper methods for common SSE operations

// SendNotification sends a notification event
func (s *SSETransport) SendNotification(ctx context.Context, notification string, data interface{}) error {
	event := &types.SSEEvent{
		ID:    uuid.New().String(),
		Event: "notification",
		Data: map[string]interface{}{
			"message": notification,
			"data":    data,
		},
		Timestamp: time.Now(),
	}

	return s.SendEvent(ctx, event)
}

// SendError sends an error event
func (s *SSETransport) SendError(ctx context.Context, err error) error {
	event := &types.SSEEvent{
		ID:    uuid.New().String(),
		Event: "error",
		Data: map[string]interface{}{
			"error": err.Error(),
			"type":  "transport_error",
		},
		Timestamp: time.Now(),
	}

	return s.SendEvent(ctx, event)
}

// SendProgress sends a progress update event
func (s *SSETransport) SendProgress(ctx context.Context, progress int, message string) error {
	event := &types.SSEEvent{
		ID:    uuid.New().String(),
		Event: "progress",
		Data: map[string]interface{}{
			"progress": progress,
			"message":  message,
		},
		Timestamp: time.Now(),
	}

	return s.SendEvent(ctx, event)
}

// SendHeartbeat sends a heartbeat event
func (s *SSETransport) SendHeartbeat(ctx context.Context) error {
	event := &types.SSEEvent{
		ID:    uuid.New().String(),
		Event: "heartbeat",
		Data: map[string]interface{}{
			"timestamp":  time.Now().Unix(),
			"session_id": s.GetSessionID(),
		},
		Timestamp: time.Now(),
	}

	return s.SendEvent(ctx, event)
}

// GetLastEventID returns the last event ID sent
func (s *SSETransport) GetLastEventID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastEventID
}

// GetConfig returns the transport configuration
func (s *SSETransport) GetConfig() map[string]interface{} {
	return s.config
}

// IsHealthy checks if the SSE connection is healthy
func (s *SSETransport) IsHealthy() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.writer != nil && s.IsConnected()
}

// GetMetrics returns SSE transport metrics
func (s *SSETransport) GetMetrics() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"connected":        s.IsConnected(),
		"event_queue_size": len(s.eventQueue),
		"keep_alive":       s.keepAlive,
		"buffer_size":      s.bufferSize,
		"last_event_id":    s.lastEventID,
		"session_id":       s.GetSessionID(),
	}
}

// eventHistory stores events for replay functionality
type eventHistory struct {
	lastID  string
	events  []*types.SSEEvent
	maxSize int
	mu      sync.RWMutex
}

// newEventHistory creates a new event history store
func newEventHistory(maxSize int) *eventHistory {
	return &eventHistory{
		events:  make([]*types.SSEEvent, 0),
		maxSize: maxSize,
	}
}

// addEvent adds an event to the history
func (eh *eventHistory) addEvent(event *types.SSEEvent) {
	eh.mu.Lock()
	defer eh.mu.Unlock()

	eh.events = append(eh.events, event)
	eh.lastID = event.ID

	// Keep only the most recent events
	if len(eh.events) > eh.maxSize {
		eh.events = eh.events[len(eh.events)-eh.maxSize:]
	}
}

// getEventsSince returns events since a specific event ID
func (eh *eventHistory) getEventsSince(since string, limit int) ([]*types.SSEEvent, error) {
	eh.mu.RLock()
	defer eh.mu.RUnlock()

	if since == "" {
		// Return latest events if no since ID provided
		start := 0
		if limit > 0 && len(eh.events) > limit {
			start = len(eh.events) - limit
		}
		return eh.events[start:], nil
	}

	// Find the event with the since ID
	var sinceIndex = -1
	for i, event := range eh.events {
		if event.ID == since {
			sinceIndex = i
			break
		}
	}

	if sinceIndex == -1 {
		return nil, fmt.Errorf("event ID %s not found in history", since)
	}

	// Return events after the since ID
	start := sinceIndex + 1
	end := len(eh.events)

	if limit > 0 && (end-start) > limit {
		end = start + limit
	}

	if start >= len(eh.events) {
		return []*types.SSEEvent{}, nil
	}

	return eh.events[start:end], nil
}

// Add event history to SSETransport
var sseEventHistory = newEventHistory(1000) // Store last 1000 events

// GetEventHistory returns recent events for replay (if supported)
func (s *SSETransport) GetEventHistory(since string, limit int) ([]*types.SSEEvent, error) {
	return sseEventHistory.getEventsSince(since, limit)
}

// SupportsReplay indicates whether this transport supports event replay
func (s *SSETransport) SupportsReplay() bool {
	return true // Now implemented with in-memory event storage
}

// init registers the SSE transport factory
func init() {
	RegisterTransport(types.TransportTypeSSE, NewSSETransport)
}
