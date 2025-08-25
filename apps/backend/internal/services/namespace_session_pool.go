package services

import (
	"sync"
	
	"github.com/google/uuid"
)

// Session represents a namespace session (placeholder for now)
type Session struct {
	ID       string
	ServerID string
	Status   string
}

// NamespaceSessionPool manages sessions for namespace servers
type NamespaceSessionPool struct {
	sessions map[string]map[string]*Session // namespace -> server -> session
	mu       sync.RWMutex
}

// NewNamespaceSessionPool creates a new namespace session pool
func NewNamespaceSessionPool() *NamespaceSessionPool {
	return &NamespaceSessionPool{
		sessions: make(map[string]map[string]*Session),
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

	// Create new session
	// This would normally create a real transport session
	// For now, creating a placeholder
	session := &Session{
		ID:       uuid.New().String(),
		ServerID: serverID,
		Status:   "active",
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
				// session.Close() // Would be called when transport is implemented
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
				// session.Close() // Would be called when transport is implemented
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