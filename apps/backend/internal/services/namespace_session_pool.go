package services

import (
	"sync"
	"time"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/mcp"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
)

// Session represents a namespace session with real MCP connection
type Session struct {
	ID           string
	ServerID     string
	NamespaceID  string
	Status       string
	Connection   *mcp.MCPClient
	LastUsed     time.Time
	Tools        []types.Tool
	Capabilities map[string]interface{}
	mu           sync.RWMutex
}

// Close closes the session and cleans up resources
func (s *Session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Status = "closed"
	if s.Connection != nil {
		return s.Connection.Close()
	}
	return nil
}

// IsConnected returns whether the session has an active connection
func (s *Session) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Connection != nil && s.Connection.IsConnected()
}

// UpdateLastUsed updates the last used timestamp
func (s *Session) UpdateLastUsed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastUsed = time.Now()
}

// NamespaceSessionPool manages sessions for namespace servers
type NamespaceSessionPool struct {
	sessions         map[string]map[string]*Session // namespace -> server -> session
	transportManager *mcp.TransportManager
	mu               sync.RWMutex
}

// NewNamespaceSessionPool creates a new namespace session pool
func NewNamespaceSessionPool() *NamespaceSessionPool {
	return &NamespaceSessionPool{
		sessions:         make(map[string]map[string]*Session),
		transportManager: mcp.NewTransportManager(),
	}
}

// GetSession gets or creates a session for a server in a namespace
func (p *NamespaceSessionPool) GetSession(namespaceID, serverID string) (*Session, error) {
	// Try to get existing session
	p.mu.RLock()
	if nsMap, ok := p.sessions[namespaceID]; ok {
		if session, ok := nsMap[serverID]; ok {
			p.mu.RUnlock()
			return session, nil
		}
	}
	p.mu.RUnlock()

	// Create new session
	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if nsMap, ok := p.sessions[namespaceID]; ok {
		if session, ok := nsMap[serverID]; ok {
			return session, nil
		}
	}

	// Create namespace map if it doesn't exist
	if _, ok := p.sessions[namespaceID]; !ok {
		p.sessions[namespaceID] = make(map[string]*Session)
	}

	// Create new session placeholder (real MCP connection will be established by namespace service)
	session := &Session{
		ID:           uuid.New().String(),
		ServerID:     serverID,
		NamespaceID:  namespaceID,
		Status:       "created",
		LastUsed:     time.Now(),
		Tools:        make([]types.Tool, 0),
		Capabilities: make(map[string]interface{}),
	}

	p.sessions[namespaceID][serverID] = session
	return session, nil
}

// RemoveSession removes a session from the pool
func (p *NamespaceSessionPool) RemoveSession(namespaceID, serverID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if nsMap, ok := p.sessions[namespaceID]; ok {
		if session, ok := nsMap[serverID]; ok {
			// Close the session if needed
			if session != nil {
				session.Close()
			}
			delete(nsMap, serverID)
		}

		// Clean up empty namespace map
		if len(nsMap) == 0 {
			delete(p.sessions, namespaceID)
		}
	}
}

// ClearNamespace removes all sessions for a namespace
func (p *NamespaceSessionPool) ClearNamespace(namespaceID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if nsMap, ok := p.sessions[namespaceID]; ok {
		// Close all sessions
		for _, session := range nsMap {
			if session != nil {
				session.Close()
			}
		}
		delete(p.sessions, namespaceID)
	}
}

// ClearServer removes a specific server session from a namespace
func (p *NamespaceSessionPool) ClearServer(namespaceID, serverID string) {
	p.RemoveSession(namespaceID, serverID)
}

// GetAllSessions returns all sessions in the pool (for debugging)
func (p *NamespaceSessionPool) GetAllSessions() map[string]map[string]*Session {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Create a copy to avoid race conditions
	result := make(map[string]map[string]*Session)
	for nsID, nsMap := range p.sessions {
		result[nsID] = make(map[string]*Session)
		for srvID, session := range nsMap {
			result[nsID][srvID] = session
		}
	}
	return result
}

// GetNamespaceSessions returns all sessions for a namespace
func (p *NamespaceSessionPool) GetNamespaceSessions(namespaceID string) map[string]*Session {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if nsMap, ok := p.sessions[namespaceID]; ok {
		// Create a copy to avoid race conditions
		result := make(map[string]*Session)
		for srvID, session := range nsMap {
			result[srvID] = session
		}
		return result
	}
	return nil
}
