package transport

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"mcp-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
)

// SessionManager manages transport sessions for stateful transports
type SessionManager struct {
	sessions map[string]*types.TransportSession
	events   map[string][]types.TransportEvent
	mu       sync.RWMutex
	config   *types.TransportConfig
	cleanup  chan struct{}
	done     chan struct{}
}

// NewSessionManager creates a new session manager
func NewSessionManager(config *types.TransportConfig) *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*types.TransportSession),
		events:   make(map[string][]types.TransportEvent),
		config:   config,
		cleanup:  make(chan struct{}, 1),
		done:     make(chan struct{}),
	}

	// Start cleanup goroutine
	go sm.cleanupLoop()

	return sm
}

// CreateSession creates a new transport session
func (sm *SessionManager) CreateSession(ctx context.Context, userID, orgID, serverID string, transportType types.TransportType) (*types.TransportSession, error) {
	sessionID := uuid.New().String()
	now := time.Now()

	session := &types.TransportSession{
		ID:             sessionID,
		UserID:         userID,
		OrganizationID: orgID,
		ServerID:       serverID,
		TransportType:  transportType,
		Status:         types.TransportSessionStatusActive,
		CreatedAt:      now,
		LastActivity:   now,
		ExpiresAt:      now.Add(sm.config.SessionTimeout),
		Metadata:       make(map[string]interface{}),
		EventStore:     make([]types.TransportEvent, 0),
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.sessions[sessionID] = session
	sm.events[sessionID] = make([]types.TransportEvent, 0)

	// Add creation event
	sm.addEventLocked(sessionID, types.TransportEventTypeConnect, map[string]interface{}{
		"transport_type":  transportType,
		"user_id":         userID,
		"organization_id": orgID,
		"server_id":       serverID,
	})

	return session, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(sessionID string) (*types.TransportSession, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session %s has expired", sessionID)
	}

	// Return a copy to avoid concurrent modification
	sessionCopy := *session
	sessionCopy.EventStore = sm.events[sessionID]

	return &sessionCopy, nil
}

// UpdateSession updates an existing session
func (sm *SessionManager) UpdateSession(sessionID string, updates map[string]interface{}) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	// Update last activity
	session.LastActivity = time.Now()

	// Apply updates
	for key, value := range updates {
		switch key {
		case "status":
			if status, ok := value.(string); ok {
				session.Status = status
			}
		case "server_id":
			if serverID, ok := value.(string); ok {
				session.ServerID = serverID
			}
		case "expires_at":
			if expiresAt, ok := value.(time.Time); ok {
				session.ExpiresAt = expiresAt
			}
		default:
			// Store in metadata
			session.Metadata[key] = value
		}
	}

	return nil
}

// CloseSession closes a session and cleans up resources
func (sm *SessionManager) CloseSession(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	session.Status = types.TransportSessionStatusClosed
	session.LastActivity = time.Now()

	// Add disconnect event
	sm.addEventLocked(sessionID, types.TransportEventTypeDisconnect, map[string]interface{}{
		"reason": "manual_close",
	})

	// Remove from active sessions after a delay to allow for final event processing
	go func() {
		time.Sleep(5 * time.Second)
		sm.mu.Lock()
		defer sm.mu.Unlock()
		delete(sm.sessions, sessionID)
		delete(sm.events, sessionID)
	}()

	return nil
}

// AddEvent adds an event to a session's event store
func (sm *SessionManager) AddEvent(sessionID string, eventType string, data map[string]interface{}) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	return sm.addEventLocked(sessionID, eventType, data)
}

// addEventLocked adds an event without acquiring the lock (internal use)
func (sm *SessionManager) addEventLocked(sessionID string, eventType string, data map[string]interface{}) error {
	_, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	event := types.TransportEvent{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
	}

	sm.events[sessionID] = append(sm.events[sessionID], event)

	// Update session last activity
	if session := sm.sessions[sessionID]; session != nil {
		session.LastActivity = event.Timestamp
	}

	return nil
}

// GetEvents retrieves events for a session
func (sm *SessionManager) GetEvents(sessionID string, since *time.Time, limit int) ([]types.TransportEvent, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	events, exists := sm.events[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	var result []types.TransportEvent
	for _, event := range events {
		if since != nil && event.Timestamp.Before(*since) {
			continue
		}
		result = append(result, event)
		if limit > 0 && len(result) >= limit {
			break
		}
	}

	return result, nil
}

// GetActiveSessions returns all active sessions
func (sm *SessionManager) GetActiveSessions() []*types.TransportSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var sessions []*types.TransportSession
	for _, session := range sm.sessions {
		if session.Status == types.TransportSessionStatusActive && time.Now().Before(session.ExpiresAt) {
			sessionCopy := *session
			sessions = append(sessions, &sessionCopy)
		}
	}

	return sessions
}

// GetSessionsByUser returns all sessions for a specific user
func (sm *SessionManager) GetSessionsByUser(userID string) []*types.TransportSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var sessions []*types.TransportSession
	for _, session := range sm.sessions {
		if session.UserID == userID {
			sessionCopy := *session
			sessions = append(sessions, &sessionCopy)
		}
	}

	return sessions
}

// GetSessionsByTransport returns all sessions for a specific transport type
func (sm *SessionManager) GetSessionsByTransport(transportType types.TransportType) []*types.TransportSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var sessions []*types.TransportSession
	for _, session := range sm.sessions {
		if session.TransportType == transportType {
			sessionCopy := *session
			sessions = append(sessions, &sessionCopy)
		}
	}

	return sessions
}

// TouchSession updates the last activity time for a session
func (sm *SessionManager) TouchSession(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	session.LastActivity = time.Now()
	// Extend expiration time if needed
	minExpiry := time.Now().Add(sm.config.SessionTimeout / 4) // Keep at least 25% of timeout remaining
	if session.ExpiresAt.Before(minExpiry) {
		session.ExpiresAt = time.Now().Add(sm.config.SessionTimeout)
	}

	return nil
}

// cleanupLoop runs periodically to clean up expired sessions
func (sm *SessionManager) cleanupLoop() {
	ticker := time.NewTicker(time.Minute) // Check every minute
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.cleanupExpiredSessions()
		case <-sm.cleanup:
			sm.cleanupExpiredSessions()
		case <-sm.done:
			return
		}
	}
}

// cleanupExpiredSessions removes expired sessions
func (sm *SessionManager) cleanupExpiredSessions() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	var expiredSessions []string

	for sessionID, session := range sm.sessions {
		if now.After(session.ExpiresAt) ||
			(session.Status == types.TransportSessionStatusClosed &&
				now.Sub(session.LastActivity) > 5*time.Minute) {
			expiredSessions = append(expiredSessions, sessionID)
		}
	}

	for _, sessionID := range expiredSessions {
		// Add expiration event before cleanup
		sm.addEventLocked(sessionID, types.TransportEventTypeDisconnect, map[string]interface{}{
			"reason": "expired",
		})
		delete(sm.sessions, sessionID)
		delete(sm.events, sessionID)
	}
}

// Shutdown gracefully shuts down the session manager
func (sm *SessionManager) Shutdown(ctx context.Context) error {
	close(sm.done)

	// Close all active sessions
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for sessionID := range sm.sessions {
		sm.addEventLocked(sessionID, types.TransportEventTypeDisconnect, map[string]interface{}{
			"reason": "shutdown",
		})
	}

	return nil
}

// GetMetrics returns session manager metrics
func (sm *SessionManager) GetMetrics() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	metrics := map[string]interface{}{
		"total_sessions":  len(sm.sessions),
		"active_sessions": 0,
		"by_transport":    make(map[string]int),
		"by_status":       make(map[string]int),
	}

	for _, session := range sm.sessions {
		if session.Status == types.TransportSessionStatusActive && time.Now().Before(session.ExpiresAt) {
			metrics["active_sessions"] = metrics["active_sessions"].(int) + 1
		}

		// Count by transport
		transportCount := metrics["by_transport"].(map[string]int)
		transportCount[string(session.TransportType)]++

		// Count by status
		statusCount := metrics["by_status"].(map[string]int)
		statusCount[session.Status]++
	}

	return metrics
}

// Implement database storage helpers for session persistence

// MetadataValue is a helper type for database storage
type MetadataValue map[string]interface{}

// Value implements the driver.Valuer interface for metadata
func (m MetadataValue) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for metadata
func (m *MetadataValue) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into MetadataValue", value)
	}

	return json.Unmarshal(bytes, m)
}
