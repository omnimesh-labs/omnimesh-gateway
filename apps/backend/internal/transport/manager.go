package transport

import (
	"context"
	"fmt"
	"sync"
	"time"

	"mcp-gateway/apps/backend/internal/types"
)

// Manager coordinates all transport implementations and sessions
type Manager struct {
	config         *types.TransportConfig
	sessionManager *SessionManager
	transports     map[types.TransportType]types.Transport
	connections    map[string]types.Transport // sessionID -> transport
	mu             sync.RWMutex
	metrics        *TransportMetrics
}

// TransportMetrics holds metrics for transport operations
type TransportMetrics struct {
	ConnectionsTotal  map[types.TransportType]int64 `json:"connections_total"`
	ActiveConnections map[types.TransportType]int   `json:"active_connections"`
	MessagesTotal     map[types.TransportType]int64 `json:"messages_total"`
	ErrorsTotal       map[types.TransportType]int64 `json:"errors_total"`
	ResponseTimeAvg   map[types.TransportType]int64 `json:"response_time_avg_ms"`
	LastActivity      time.Time                     `json:"last_activity"`
	mu                sync.RWMutex                  `json:"-"`
}

// NewManager creates a new transport manager
func NewManager(config *types.TransportConfig) *Manager {
	return &Manager{
		config:         config,
		sessionManager: NewSessionManager(config),
		transports:     make(map[types.TransportType]types.Transport),
		connections:    make(map[string]types.Transport),
		metrics:        NewTransportMetrics(),
	}
}

// NewTransportMetrics creates a new metrics instance
func NewTransportMetrics() *TransportMetrics {
	return &TransportMetrics{
		ConnectionsTotal:  make(map[types.TransportType]int64),
		ActiveConnections: make(map[types.TransportType]int),
		MessagesTotal:     make(map[types.TransportType]int64),
		ErrorsTotal:       make(map[types.TransportType]int64),
		ResponseTimeAvg:   make(map[types.TransportType]int64),
		LastActivity:      time.Now(),
	}
}

// Initialize initializes the transport manager with enabled transports
func (m *Manager) Initialize(ctx context.Context) error {
	for _, transportType := range m.config.EnabledTransports {
		if !IsTransportRegistered(transportType) {
			return fmt.Errorf("transport type %s is not registered", transportType)
		}
	}
	return nil
}

// CreateConnection creates a new connection for the specified transport type
func (m *Manager) CreateConnection(ctx context.Context, transportType types.TransportType, userID, orgID, serverID string) (types.Transport, *types.TransportSession, error) {
	return m.CreateConnectionWithConfig(ctx, transportType, userID, orgID, serverID, nil)
}

// CreateConnectionWithConfig creates a new connection with custom configuration
func (m *Manager) CreateConnectionWithConfig(ctx context.Context, transportType types.TransportType, userID, orgID, serverID string, customConfig map[string]interface{}) (types.Transport, *types.TransportSession, error) {
	// Check if transport type is enabled
	if !m.isTransportEnabled(transportType) {
		return nil, nil, fmt.Errorf("transport type %s is not enabled", transportType)
	}

	// Create session for stateful transports
	var session *types.TransportSession
	var err error

	if m.isStatefulTransport(transportType) {
		session, err = m.sessionManager.CreateSession(ctx, userID, orgID, serverID, transportType)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create session: %w", err)
		}
	}

	// Create transport instance
	config := map[string]interface{}{
		"type":                transportType,
		"session_timeout":     m.config.SessionTimeout,
		"max_connections":     m.config.MaxConnections,
		"buffer_size":         m.config.BufferSize,
		"sse_keep_alive":      m.config.SSEKeepAlive,
		"websocket_timeout":   m.config.WebSocketTimeout,
		"streamable_stateful": m.config.StreamableStateful,
		"stdio_timeout":       m.config.STDIOTimeout,
	}

	// Merge custom configuration
	for key, value := range customConfig {
		config[key] = value
	}

	transport, err := CreateTransport(transportType, config)
	if err != nil {
		if session != nil {
			m.sessionManager.CloseSession(session.ID)
		}
		return nil, nil, fmt.Errorf("failed to create transport: %w", err)
	}

	// Set session ID for stateful transports
	if session != nil {
		transport.SetSessionID(session.ID)
	}

	// Store connection
	if session != nil {
		m.mu.Lock()
		m.connections[session.ID] = transport
		m.mu.Unlock()
	}

	// Update metrics
	m.metrics.mu.Lock()
	m.metrics.ConnectionsTotal[transportType]++
	m.metrics.ActiveConnections[transportType]++
	m.metrics.LastActivity = time.Now()
	m.metrics.mu.Unlock()

	return transport, session, nil
}

// GetConnection retrieves an existing connection by session ID
func (m *Manager) GetConnection(sessionID string) (types.Transport, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	transport, exists := m.connections[sessionID]
	if !exists {
		return nil, fmt.Errorf("connection not found for session %s", sessionID)
	}

	return transport, nil
}

// CloseConnection closes a connection and cleans up resources
func (m *Manager) CloseConnection(sessionID string) error {
	m.mu.Lock()
	transport, exists := m.connections[sessionID]
	if exists {
		delete(m.connections, sessionID)
	}
	m.mu.Unlock()

	if !exists {
		return fmt.Errorf("connection not found for session %s", sessionID)
	}

	// Disconnect transport
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := transport.Disconnect(ctx); err != nil {
		// Log error but continue cleanup
	}

	// Close session
	if err := m.sessionManager.CloseSession(sessionID); err != nil {
		return fmt.Errorf("failed to close session: %w", err)
	}

	// Update metrics
	transportType := transport.GetTransportType()
	m.metrics.mu.Lock()
	if m.metrics.ActiveConnections[transportType] > 0 {
		m.metrics.ActiveConnections[transportType]--
	}
	m.metrics.mu.Unlock()

	return nil
}

// SendMessage sends a message through the specified transport
func (m *Manager) SendMessage(ctx context.Context, sessionID string, message interface{}) error {
	transport, err := m.GetConnection(sessionID)
	if err != nil {
		return err
	}

	// Validate message
	if err := ValidateMessage(message); err != nil {
		m.metrics.mu.Lock()
		m.metrics.ErrorsTotal[transport.GetTransportType()]++
		m.metrics.mu.Unlock()
		return fmt.Errorf("message validation failed: %w", err)
	}

	// Send message
	start := time.Now()
	err = transport.SendMessage(ctx, message)
	duration := time.Since(start)

	// Update metrics
	transportType := transport.GetTransportType()
	m.metrics.mu.Lock()
	m.metrics.MessagesTotal[transportType]++
	if err != nil {
		m.metrics.ErrorsTotal[transportType]++
	} else {
		// Update average response time (simple moving average)
		if m.metrics.ResponseTimeAvg[transportType] == 0 {
			m.metrics.ResponseTimeAvg[transportType] = duration.Milliseconds()
		} else {
			m.metrics.ResponseTimeAvg[transportType] = (m.metrics.ResponseTimeAvg[transportType] + duration.Milliseconds()) / 2
		}
	}
	m.metrics.LastActivity = time.Now()
	m.metrics.mu.Unlock()

	// Touch session to keep it active
	if sessionID != "" {
		m.sessionManager.TouchSession(sessionID)

		// Add message event to session
		eventData := map[string]interface{}{
			"direction": "outbound",
			"type":      fmt.Sprintf("%T", message),
			"duration":  duration.Milliseconds(),
		}
		if err != nil {
			eventData["error"] = err.Error()
		}
		m.sessionManager.AddEvent(sessionID, types.TransportEventTypeMessage, eventData)
	}

	return err
}

// ReceiveMessage receives a message from the specified transport
func (m *Manager) ReceiveMessage(ctx context.Context, sessionID string) (interface{}, error) {
	transport, err := m.GetConnection(sessionID)
	if err != nil {
		return nil, err
	}

	// Receive message
	start := time.Now()
	message, err := transport.ReceiveMessage(ctx)
	duration := time.Since(start)

	// Update metrics
	transportType := transport.GetTransportType()
	m.metrics.mu.Lock()
	m.metrics.MessagesTotal[transportType]++
	if err != nil {
		m.metrics.ErrorsTotal[transportType]++
	}
	m.metrics.LastActivity = time.Now()
	m.metrics.mu.Unlock()

	// Touch session and add event
	if sessionID != "" {
		m.sessionManager.TouchSession(sessionID)

		eventData := map[string]interface{}{
			"direction": "inbound",
			"duration":  duration.Milliseconds(),
		}
		if err != nil {
			eventData["error"] = err.Error()
		} else {
			eventData["type"] = fmt.Sprintf("%T", message)
		}
		m.sessionManager.AddEvent(sessionID, types.TransportEventTypeMessage, eventData)
	}

	return message, err
}

// BroadcastMessage sends a message to all connections of a specific transport type
func (m *Manager) BroadcastMessage(ctx context.Context, transportType types.TransportType, message interface{}) error {
	m.mu.RLock()
	var connections []types.Transport
	for _, transport := range m.connections {
		if transport.GetTransportType() == transportType {
			connections = append(connections, transport)
		}
	}
	m.mu.RUnlock()

	var errors []error
	for _, transport := range connections {
		if err := transport.SendMessage(ctx, message); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("broadcast failed for %d connections: %v", len(errors), errors)
	}

	return nil
}

// GetSession returns session information
func (m *Manager) GetSession(sessionID string) (*types.TransportSession, error) {
	return m.sessionManager.GetSession(sessionID)
}

// GetActiveSessions returns all active sessions
func (m *Manager) GetActiveSessions() []*types.TransportSession {
	return m.sessionManager.GetActiveSessions()
}

// GetSessionsByUser returns sessions for a specific user
func (m *Manager) GetSessionsByUser(userID string) []*types.TransportSession {
	return m.sessionManager.GetSessionsByUser(userID)
}

// GetSessionEvents returns events for a session
func (m *Manager) GetSessionEvents(sessionID string, since *time.Time, limit int) ([]types.TransportEvent, error) {
	return m.sessionManager.GetEvents(sessionID, since, limit)
}

// GetMetrics returns transport manager metrics
func (m *Manager) GetMetrics() map[string]interface{} {
	m.metrics.mu.RLock()
	defer m.metrics.mu.RUnlock()

	sessionMetrics := m.sessionManager.GetMetrics()

	return map[string]interface{}{
		"connections_total":  m.metrics.ConnectionsTotal,
		"active_connections": m.metrics.ActiveConnections,
		"messages_total":     m.metrics.MessagesTotal,
		"errors_total":       m.metrics.ErrorsTotal,
		"response_time_avg":  m.metrics.ResponseTimeAvg,
		"last_activity":      m.metrics.LastActivity,
		"sessions":           sessionMetrics,
		"enabled_transports": m.config.EnabledTransports,
		"max_connections":    m.config.MaxConnections,
	}
}

// HealthCheck performs health checks on all active transports
func (m *Manager) HealthCheck(ctx context.Context) map[types.TransportType]error {
	results := make(map[types.TransportType]error)

	for _, transportType := range m.config.EnabledTransports {
		// Create a test transport instance
		config := map[string]interface{}{
			"type": transportType,
		}

		transport, err := CreateTransport(transportType, config)
		if err != nil {
			results[transportType] = fmt.Errorf("failed to create transport: %w", err)
			continue
		}

		// Test connection
		if err := transport.Connect(ctx); err != nil {
			results[transportType] = fmt.Errorf("connection failed: %w", err)
			continue
		}

		// Disconnect test transport
		transport.Disconnect(ctx)
		results[transportType] = nil
	}

	return results
}

// Shutdown gracefully shuts down the transport manager
func (m *Manager) Shutdown(ctx context.Context) error {
	// Close all connections
	m.mu.Lock()
	connections := make(map[string]types.Transport)
	for k, v := range m.connections {
		connections[k] = v
	}
	m.mu.Unlock()

	for sessionID, transport := range connections {
		if err := transport.Disconnect(ctx); err != nil {
			// Log error but continue
		}
		delete(m.connections, sessionID)
	}

	// Shutdown session manager
	return m.sessionManager.Shutdown(ctx)
}

// Helper methods

// isTransportEnabled checks if a transport type is enabled in config
func (m *Manager) isTransportEnabled(transportType types.TransportType) bool {
	for _, enabled := range m.config.EnabledTransports {
		if enabled == transportType {
			return true
		}
	}
	return false
}

// isStatefulTransport checks if a transport type requires session management
func (m *Manager) isStatefulTransport(transportType types.TransportType) bool {
	switch transportType {
	case types.TransportTypeSSE, types.TransportTypeWebSocket, types.TransportTypeStreamable, types.TransportTypeSTDIO:
		return true
	case types.TransportTypeHTTP:
		return false
	default:
		return false
	}
}

// GetSupportedTransports returns all supported transport types
func (m *Manager) GetSupportedTransports() []types.TransportType {
	return []types.TransportType{
		types.TransportTypeHTTP,
		types.TransportTypeSSE,
		types.TransportTypeWebSocket,
		types.TransportTypeStreamable,
		types.TransportTypeSTDIO,
	}
}

// GetEnabledTransports returns enabled transport types
func (m *Manager) GetEnabledTransports() []types.TransportType {
	return m.config.EnabledTransports
}
