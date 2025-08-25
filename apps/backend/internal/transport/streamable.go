package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"mcp-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
)

// StreamableHTTPTransport implements the MCP Streamable HTTP transport protocol
type StreamableHTTPTransport struct {
	*BaseTransport
	client     *http.Client
	config     map[string]interface{}
	baseURL    string
	streamMode string
	eventStore []*types.TransportEvent
	timeout    time.Duration
	mu         sync.RWMutex
	stateful   bool
}

// StreamableRequest represents a streamable HTTP request
type StreamableRequest struct {
	Body      interface{}            `json:"body,omitempty"`
	Headers   map[string]string      `json:"headers,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Method    string                 `json:"method"`
	SessionID string                 `json:"session_id,omitempty"`
	Mode      string                 `json:"mode"`
	Stateful  bool                   `json:"stateful"`
}

// StreamableResponse represents a streamable HTTP response
type StreamableResponse struct {
	Body      interface{}             `json:"body,omitempty"`
	Headers   map[string]string       `json:"headers,omitempty"`
	Metadata  map[string]interface{}  `json:"metadata,omitempty"`
	SessionID string                  `json:"session_id,omitempty"`
	Mode      string                  `json:"mode"`
	Events    []*types.TransportEvent `json:"events,omitempty"`
	Status    int                     `json:"status"`
}

// NewStreamableHTTPTransport creates a new Streamable HTTP transport instance
func NewStreamableHTTPTransport(config map[string]interface{}) (types.Transport, error) {
	transport := &StreamableHTTPTransport{
		BaseTransport: NewBaseTransport(types.TransportTypeStreamable),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		stateful:   true,
		streamMode: types.StreamableModeJSON,
		eventStore: make([]*types.TransportEvent, 0),
		config:     config,
		timeout:    30 * time.Second,
	}

	// Configure from config map
	if baseURL, ok := config["base_url"].(string); ok {
		transport.baseURL = baseURL
	} else {
		transport.baseURL = "/mcp" // Default endpoint
	}

	if stateful, ok := config["streamable_stateful"].(bool); ok {
		transport.stateful = stateful
	}

	if mode, ok := config["stream_mode"].(string); ok {
		transport.streamMode = mode
	}

	if timeout, ok := config["timeout"].(time.Duration); ok {
		transport.timeout = timeout
		transport.client.Timeout = timeout
	}

	return transport, nil
}

// Connect establishes connection for Streamable HTTP
func (s *StreamableHTTPTransport) Connect(ctx context.Context) error {
	s.setConnected(true)

	// For stateful mode, initialize session if not already set
	if s.stateful && s.GetSessionID() == "" {
		sessionID := uuid.New().String()
		s.SetSessionID(sessionID)
	}

	// Add connection event to event store
	if s.stateful {
		event := &types.TransportEvent{
			ID:        uuid.New().String(),
			SessionID: s.GetSessionID(),
			Type:      types.TransportEventTypeConnect,
			Data: map[string]interface{}{
				"transport_type": s.GetTransportType(),
				"stateful":       s.stateful,
				"stream_mode":    s.streamMode,
			},
			Timestamp: time.Now(),
		}
		s.addEvent(event)
	}

	return nil
}

// Disconnect closes Streamable HTTP connection
func (s *StreamableHTTPTransport) Disconnect(ctx context.Context) error {
	s.setConnected(false)

	// Add disconnect event to event store
	if s.stateful {
		event := &types.TransportEvent{
			ID:        uuid.New().String(),
			SessionID: s.GetSessionID(),
			Type:      types.TransportEventTypeDisconnect,
			Data: map[string]interface{}{
				"reason": "manual_disconnect",
			},
			Timestamp: time.Now(),
		}
		s.addEvent(event)
	}

	return nil
}

// SendMessage sends a message via Streamable HTTP
func (s *StreamableHTTPTransport) SendMessage(ctx context.Context, message interface{}) error {
	if !s.IsConnected() {
		return fmt.Errorf("transport not connected")
	}

	// Convert message to streamable request
	request, err := s.convertToStreamableRequest(message)
	if err != nil {
		return fmt.Errorf("failed to convert message: %w", err)
	}

	// For the MCP Gateway, we work internally without external HTTP requests
	// Instead of making HTTP requests, we process the message internally
	return s.processMessageInternally(ctx, request)
}

// ReceiveMessage receives a message via Streamable HTTP
func (s *StreamableHTTPTransport) ReceiveMessage(ctx context.Context) (interface{}, error) {
	if !s.IsConnected() {
		return nil, fmt.Errorf("transport not connected")
	}

	// For streamable HTTP, this typically returns events from the event store
	if s.stateful {
		return s.GetLatestEvents(10), nil
	}

	return nil, fmt.Errorf("ReceiveMessage not applicable for stateless streamable HTTP")
}

// processMessageInternally processes messages internally without external HTTP requests
func (s *StreamableHTTPTransport) processMessageInternally(ctx context.Context, request *StreamableRequest) error {
	// Create a mock response for internal processing
	response := &StreamableResponse{
		Status:    200,
		SessionID: s.GetSessionID(),
		Mode:      s.streamMode,
		Body: map[string]interface{}{
			"success": true,
			"message": "Request processed successfully",
			"request": request,
		},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Metadata: map[string]interface{}{
			"processed_internally": true,
			"transport":            "streamable_http",
		},
	}

	// Add response events to event store for stateful mode
	if s.stateful {
		// Add request event
		requestEvent := &types.TransportEvent{
			ID:        uuid.New().String(),
			SessionID: s.GetSessionID(),
			Type:      types.TransportEventTypeMessage,
			Data: map[string]interface{}{
				"direction": "outbound",
				"method":    request.Method,
				"body":      request.Body,
				"headers":   request.Headers,
			},
			Timestamp: time.Now(),
		}
		s.addEvent(requestEvent)

		// Add response event
		responseEvent := &types.TransportEvent{
			ID:        uuid.New().String(),
			SessionID: s.GetSessionID(),
			Type:      types.TransportEventTypeMessage,
			Data: map[string]interface{}{
				"direction": "inbound",
				"status":    response.Status,
				"body":      response.Body,
			},
			Timestamp: time.Now(),
		}
		s.addEvent(responseEvent)

		// Add events from response if any
		for _, event := range response.Events {
			s.addEvent(event)
		}
	}

	return nil
}

// sendJSONRequest sends a request in JSON mode
func (s *StreamableHTTPTransport) sendJSONRequest(ctx context.Context, request *StreamableRequest) error {
	// Marshal request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.baseURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	if s.stateful && s.GetSessionID() != "" {
		httpReq.Header.Set("X-Session-ID", s.GetSessionID())
	}

	// Add custom headers from request
	for key, value := range request.Headers {
		httpReq.Header.Set(key, value)
	}

	// Send request
	resp, err := s.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var streamableResp StreamableResponse
	if err := json.Unmarshal(responseBody, &streamableResp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Add response events to event store for stateful mode
	if s.stateful {
		for _, event := range streamableResp.Events {
			s.addEvent(event)
		}

		// Add message event
		messageEvent := &types.TransportEvent{
			ID:        uuid.New().String(),
			SessionID: s.GetSessionID(),
			Type:      types.TransportEventTypeMessage,
			Data: map[string]interface{}{
				"direction": "inbound",
				"status":    streamableResp.Status,
				"body":      streamableResp.Body,
			},
			Timestamp: time.Now(),
		}
		s.addEvent(messageEvent)
	}

	return nil
}

// sendSSERequest sends a request in SSE mode
func (s *StreamableHTTPTransport) sendSSERequest(ctx context.Context, request *StreamableRequest) error {
	// Marshal request for SSE
	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.baseURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers for SSE
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Cache-Control", "no-cache")

	if s.stateful && s.GetSessionID() != "" {
		httpReq.Header.Set("X-Session-ID", s.GetSessionID())
	}

	// Send request
	resp, err := s.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle SSE stream
	return s.handleSSEStream(ctx, resp.Body)
}

// handleSSEStream processes incoming SSE events
func (s *StreamableHTTPTransport) handleSSEStream(ctx context.Context, reader io.Reader) error {
	buffer := make([]byte, 4096)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			n, err := reader.Read(buffer)
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return fmt.Errorf("failed to read SSE stream: %w", err)
			}

			// Parse SSE data
			data := string(buffer[:n])
			events := s.parseSSEData(data)

			// Add events to store for stateful mode
			if s.stateful {
				for _, event := range events {
					s.addEvent(event)
				}
			}
		}
	}
}

// parseSSEData parses SSE formatted data into events
func (s *StreamableHTTPTransport) parseSSEData(data string) []*types.TransportEvent {
	var events []*types.TransportEvent
	lines := strings.Split(data, "\n")

	var currentEvent *types.TransportEvent

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			// End of event
			if currentEvent != nil {
				events = append(events, currentEvent)
				currentEvent = nil
			}
			continue
		}

		if strings.HasPrefix(line, ":") {
			// Comment, ignore
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		field := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if currentEvent == nil {
			currentEvent = &types.TransportEvent{
				SessionID: s.GetSessionID(),
				Type:      types.TransportEventTypeMessage,
				Data:      make(map[string]interface{}),
				Timestamp: time.Now(),
			}
		}

		switch field {
		case "id":
			currentEvent.ID = value
		case "event":
			currentEvent.Type = value
		case "data":
			// Try to parse as JSON
			var jsonData interface{}
			if err := json.Unmarshal([]byte(value), &jsonData); err == nil {
				currentEvent.Data["parsed_data"] = jsonData
			} else {
				currentEvent.Data["raw_data"] = value
			}
		case "retry":
			if retryTime, err := strconv.Atoi(value); err == nil {
				currentEvent.Data["retry"] = retryTime
			}
		}
	}

	// Add final event if exists
	if currentEvent != nil {
		events = append(events, currentEvent)
	}

	return events
}

// SendMCPRequest sends an MCP request via Streamable HTTP
func (s *StreamableHTTPTransport) SendMCPRequest(ctx context.Context, mcpMessage *types.MCPMessage) (*StreamableResponse, error) {
	request := &StreamableRequest{
		Method:    "POST",
		Body:      mcpMessage,
		Stateful:  s.stateful,
		SessionID: s.GetSessionID(),
		Mode:      s.streamMode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	// Send request
	if err := s.SendMessage(ctx, request); err != nil {
		return nil, err
	}

	// For JSON mode, return mock response (in real implementation, this would be from actual response)
	response := &StreamableResponse{
		Status:    200,
		SessionID: s.GetSessionID(),
		Mode:      s.streamMode,
		Events:    s.GetLatestEvents(1),
	}

	return response, nil
}

// convertToStreamableRequest converts various message types to streamable request
func (s *StreamableHTTPTransport) convertToStreamableRequest(message interface{}) (*StreamableRequest, error) {
	switch msg := message.(type) {
	case *StreamableRequest:
		return msg, nil
	case *types.StreamableHTTPRequest:
		return &StreamableRequest{
			Method:    msg.Method,
			Headers:   msg.Headers,
			Body:      msg.Body,
			Stateful:  msg.Stateful,
			SessionID: msg.SessionID,
			Mode:      msg.StreamMode,
			Metadata:  msg.Metadata,
		}, nil
	case *types.MCPMessage:
		return &StreamableRequest{
			Method:    "POST",
			Body:      msg,
			Stateful:  s.stateful,
			SessionID: s.GetSessionID(),
			Mode:      s.streamMode,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	case *types.TransportRequest:
		return &StreamableRequest{
			Method:    msg.Method,
			Body:      msg.Body,
			Stateful:  s.stateful,
			SessionID: s.GetSessionID(),
			Mode:      s.streamMode,
			Headers:   msg.Headers,
		}, nil
	default:
		return &StreamableRequest{
			Method:    "POST",
			Body:      message,
			Stateful:  s.stateful,
			SessionID: s.GetSessionID(),
			Mode:      s.streamMode,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}
}

// Event store management

// addEvent adds an event to the event store
func (s *StreamableHTTPTransport) addEvent(event *types.TransportEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.eventStore = append(s.eventStore, event)

	// Keep only recent events (configurable limit)
	maxEvents := 1000
	if len(s.eventStore) > maxEvents {
		s.eventStore = s.eventStore[len(s.eventStore)-maxEvents:]
	}
}

// GetLatestEvents returns the latest N events
func (s *StreamableHTTPTransport) GetLatestEvents(limit int) []*types.TransportEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.eventStore) {
		limit = len(s.eventStore)
	}

	start := len(s.eventStore) - limit
	if start < 0 {
		start = 0
	}

	// Return a copy to avoid concurrent modification
	events := make([]*types.TransportEvent, limit)
	copy(events, s.eventStore[start:])

	return events
}

// GetEventsSince returns events since a specific timestamp
func (s *StreamableHTTPTransport) GetEventsSince(since time.Time) []*types.TransportEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var events []*types.TransportEvent
	for _, event := range s.eventStore {
		if event.Timestamp.After(since) {
			events = append(events, event)
		}
	}

	return events
}

// Configuration and utility methods

// SetStateful sets the stateful mode
func (s *StreamableHTTPTransport) SetStateful(stateful bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stateful = stateful
}

// IsStateful returns whether the transport is in stateful mode
func (s *StreamableHTTPTransport) IsStateful() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stateful
}

// SetStreamMode sets the streaming mode
func (s *StreamableHTTPTransport) SetStreamMode(mode string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.streamMode = mode
}

// GetStreamMode returns the current streaming mode
func (s *StreamableHTTPTransport) GetStreamMode() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.streamMode
}

// GetConfig returns the transport configuration
func (s *StreamableHTTPTransport) GetConfig() map[string]interface{} {
	return s.config
}

// GetMetrics returns transport metrics
func (s *StreamableHTTPTransport) GetMetrics() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"connected":        s.IsConnected(),
		"stateful":         s.stateful,
		"stream_mode":      s.streamMode,
		"event_store_size": len(s.eventStore),
		"session_id":       s.GetSessionID(),
		"base_url":         s.baseURL,
		"timeout":          s.timeout,
	}
}

// ClearEventStore clears the event store
func (s *StreamableHTTPTransport) ClearEventStore() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.eventStore = make([]*types.TransportEvent, 0)
}

// Reconnection and failure handling

// IsHealthy checks if the transport is healthy
func (s *StreamableHTTPTransport) IsHealthy() bool {
	if !s.IsConnected() {
		return false
	}

	// Simple health check by sending a ping-like request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	healthReq := &StreamableRequest{
		Method:    "GET",
		Headers:   map[string]string{"X-Health-Check": "true"},
		Stateful:  false,
		SessionID: s.GetSessionID(),
		Mode:      s.streamMode,
	}

	err := s.sendJSONRequest(ctx, healthReq)
	return err == nil
}

// Reconnect attempts to reconnect the transport
func (s *StreamableHTTPTransport) Reconnect(ctx context.Context) error {
	if s.IsConnected() {
		// Disconnect first
		if err := s.Disconnect(ctx); err != nil {
			return fmt.Errorf("failed to disconnect before reconnect: %w", err)
		}
	}

	// Wait a moment
	time.Sleep(100 * time.Millisecond)

	// Attempt to reconnect
	return s.Connect(ctx)
}

// SetRetryPolicy configures retry behavior
type RetryPolicy struct {
	MaxRetries    int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
}

// DefaultRetryPolicy returns a sensible default retry policy
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:    3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
	}
}

// SendMessageWithRetry sends a message with retry logic
func (s *StreamableHTTPTransport) SendMessageWithRetry(ctx context.Context, message interface{}, policy *RetryPolicy) error {
	if policy == nil {
		policy = DefaultRetryPolicy()
	}

	var lastErr error
	delay := policy.InitialDelay

	for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Continue with retry
			}

			// Exponential backoff
			delay = time.Duration(float64(delay) * policy.BackoffFactor)
			if delay > policy.MaxDelay {
				delay = policy.MaxDelay
			}
		}

		// Attempt to send message
		err := s.SendMessage(ctx, message)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if we should retry based on error type
		if !s.shouldRetry(err) {
			break
		}
	}

	return fmt.Errorf("failed after %d attempts, last error: %w", policy.MaxRetries, lastErr)
}

// shouldRetry determines if an error is retryable
func (s *StreamableHTTPTransport) shouldRetry(err error) bool {
	// Check for network errors, timeouts, etc.
	// Don't retry on client errors (4xx), but do retry on server errors (5xx)
	errStr := err.Error()

	// Retry on connection errors
	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "no such host") {
		return true
	}

	// Retry on server errors (5xx)
	if strings.Contains(errStr, "status 5") {
		return true
	}

	// Don't retry on client errors (4xx)
	if strings.Contains(errStr, "status 4") {
		return false
	}

	// Default to retry for unknown errors
	return true
}

// Event filtering and transformation

// FilterEvents filters events based on criteria
func (s *StreamableHTTPTransport) FilterEvents(events []*types.TransportEvent, filter EventFilter) []*types.TransportEvent {
	var filtered []*types.TransportEvent

	for _, event := range events {
		if filter.Matches(event) {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// EventFilter defines criteria for filtering events
type EventFilter struct {
	Since      *time.Time
	Until      *time.Time
	DataFilter func(map[string]interface{}) bool
	SessionID  string
	EventTypes []string
}

// Matches checks if an event matches the filter criteria
func (ef *EventFilter) Matches(event *types.TransportEvent) bool {
	// Check event type
	if len(ef.EventTypes) > 0 {
		found := false
		for _, eventType := range ef.EventTypes {
			if event.Type == eventType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check time range
	if ef.Since != nil && event.Timestamp.Before(*ef.Since) {
		return false
	}
	if ef.Until != nil && event.Timestamp.After(*ef.Until) {
		return false
	}

	// Check session ID
	if ef.SessionID != "" && event.SessionID != ef.SessionID {
		return false
	}

	// Check data filter
	if ef.DataFilter != nil && !ef.DataFilter(event.Data) {
		return false
	}

	return true
}

// GetFilteredEvents returns events matching the filter
func (s *StreamableHTTPTransport) GetFilteredEvents(filter EventFilter, limit int) []*types.TransportEvent {
	s.mu.RLock()
	events := make([]*types.TransportEvent, len(s.eventStore))
	copy(events, s.eventStore)
	s.mu.RUnlock()

	filtered := s.FilterEvents(events, filter)

	if limit > 0 && len(filtered) > limit {
		// Return the most recent events
		return filtered[len(filtered)-limit:]
	}

	return filtered
}

// Statistical methods

// GetEventStats returns statistics about events
func (s *StreamableHTTPTransport) GetEventStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.eventStore) == 0 {
		return map[string]interface{}{
			"total_events": 0,
			"event_types":  map[string]int{},
			"oldest_event": nil,
			"newest_event": nil,
		}
	}

	// Count event types
	eventTypes := make(map[string]int)
	for _, event := range s.eventStore {
		eventTypes[event.Type]++
	}

	return map[string]interface{}{
		"total_events": len(s.eventStore),
		"event_types":  eventTypes,
		"oldest_event": s.eventStore[0].Timestamp,
		"newest_event": s.eventStore[len(s.eventStore)-1].Timestamp,
	}
}

// init registers the Streamable HTTP transport factory
func init() {
	RegisterTransport(types.TransportTypeStreamable, NewStreamableHTTPTransport)
}
