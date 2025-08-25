package discovery

import (
	"database/sql"
	"sync"

	"mcp-gateway/apps/backend/internal/types"
)

// Registry manages the server registry
type Registry struct {
	db      *sql.DB
	servers map[string]*types.MCPServer
	stats   map[string]*types.ServerStats
	mu      sync.RWMutex
}

// NewRegistry creates a new server registry
func NewRegistry(db *sql.DB) *Registry {
	return &Registry{
		db:      db,
		servers: make(map[string]*types.MCPServer),
		stats:   make(map[string]*types.ServerStats),
	}
}

// Register adds a server to the registry
func (r *Registry) Register(server *types.MCPServer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// TODO: Implement server registration
	// Store in database
	// Add to in-memory cache
	r.servers[server.ID] = server
	r.initializeStats(server.ID)

	return nil
}

// Unregister removes a server from the registry
func (r *Registry) Unregister(serverID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// TODO: Implement server unregistration
	// Remove from database
	// Remove from in-memory cache
	delete(r.servers, serverID)
	delete(r.stats, serverID)

	return nil
}

// GetServer retrieves a server by ID
func (r *Registry) GetServer(serverID string) (*types.MCPServer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	server, exists := r.servers[serverID]
	return server, exists
}

// GetServers returns all servers for an organization
func (r *Registry) GetServers(orgID string) []*types.MCPServer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var servers []*types.MCPServer
	for _, server := range r.servers {
		if server.OrganizationID == orgID {
			servers = append(servers, server)
		}
	}

	return servers
}

// GetHealthyServers returns healthy servers for an organization
func (r *Registry) GetHealthyServers(orgID string) []*types.MCPServer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var servers []*types.MCPServer
	for _, server := range r.servers {
		if server.OrganizationID == orgID && server.Status == types.ServerStatusActive {
			servers = append(servers, server)
		}
	}

	return servers
}

// UpdateServerStatus updates the status of a server
func (r *Registry) UpdateServerStatus(serverID, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if server, exists := r.servers[serverID]; exists {
		server.Status = status
		// TODO: Update database
		return nil
	}

	return types.NewNotFoundError("Server not found")
}

// GetServerStats returns statistics for a server
func (r *Registry) GetServerStats(serverID string) (*types.ServerStats, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats, exists := r.stats[serverID]
	return stats, exists
}

// UpdateServerStats updates statistics for a server
func (r *Registry) UpdateServerStats(serverID string, stats *types.ServerStats) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.stats[serverID] = stats
}

// IncrementRequests increments the request count for a server
func (r *Registry) IncrementRequests(serverID string, success bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if stats, exists := r.stats[serverID]; exists {
		stats.TotalRequests++
		if success {
			stats.SuccessRequests++
		} else {
			stats.ErrorRequests++
		}
	}
}

// UpdateLatency updates the average latency for a server
func (r *Registry) UpdateLatency(serverID string, latency float64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if stats, exists := r.stats[serverID]; exists {
		// Simple exponential moving average
		alpha := 0.1
		stats.AvgLatency = alpha*latency + (1-alpha)*stats.AvgLatency
	}
}

// LoadFromDatabase loads servers from database into memory
func (r *Registry) LoadFromDatabase() error {
	// TODO: Implement database loading
	// Query all active servers
	// Load into memory cache
	return nil
}

// initializeStats initializes statistics for a server
func (r *Registry) initializeStats(serverID string) {
	r.stats[serverID] = &types.ServerStats{
		ServerID:        serverID,
		TotalRequests:   0,
		SuccessRequests: 0,
		ErrorRequests:   0,
		AvgLatency:      0,
	}
}
