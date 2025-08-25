package discovery

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"mcp-gateway/apps/backend/internal/database/models"
	"mcp-gateway/apps/backend/internal/types"
)

// dbWrapper wraps *sql.DB to implement the Database interface
type dbWrapper struct {
	*sql.DB
}

// Registry manages the server registry
type Registry struct {
	db          *sql.DB
	serverModel *models.MCPServerModel
	servers     map[string]*types.MCPServer
	stats       map[string]*types.ServerStats
	mu          sync.RWMutex
}

// NewRegistry creates a new server registry
func NewRegistry(db *sql.DB) *Registry {
	// Wrap the database to implement the Database interface
	dbWrap := &dbWrapper{db}

	return &Registry{
		db:          db,
		serverModel: models.NewMCPServerModel(dbWrap),
		servers:     make(map[string]*types.MCPServer),
		stats:       make(map[string]*types.ServerStats),
	}
}

// Register adds a server to the registry
func (r *Registry) Register(server *types.MCPServer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Convert types.MCPServer to models.MCPServer for database storage
	serverUUID, err := uuid.Parse(server.ID)
	if err != nil {
		return fmt.Errorf("invalid server ID: %w", err)
	}

	orgUUID, err := uuid.Parse(server.OrganizationID)
	if err != nil {
		return fmt.Errorf("invalid organization ID: %w", err)
	}

	modelServer := &models.MCPServer{
		ID:             serverUUID,
		OrganizationID: orgUUID,
		Name:           server.Name,
		Description:    sql.NullString{String: server.Description, Valid: server.Description != ""},
		Protocol:       server.Protocol,
		URL:            sql.NullString{String: server.URL, Valid: server.URL != ""},
		Command:        sql.NullString{String: server.Command, Valid: server.Command != ""},
		Args:           pq.StringArray(server.Args),
		Environment:    pq.StringArray(server.Environment),
		WorkingDir:     sql.NullString{String: server.WorkingDir, Valid: server.WorkingDir != ""},
		Version:        sql.NullString{String: server.Version, Valid: server.Version != ""},
		TimeoutSeconds: int(server.Timeout.Seconds()),
		MaxRetries:     server.MaxRetries,
		Status:         server.Status,
		HealthCheckURL: sql.NullString{String: server.HealthCheckURL, Valid: server.HealthCheckURL != ""},
		IsActive:       server.IsActive,
		Metadata:       convertStringMapToInterface(server.Metadata),
		Tags:           pq.StringArray{},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Store in database
	err = r.serverModel.Create(modelServer)
	if err != nil {
		return fmt.Errorf("failed to store server in database: %w", err)
	}

	// Add to in-memory cache
	r.servers[server.ID] = server
	r.initializeStats(server.ID)

	log.Printf("Server %s (%s) registered successfully", server.Name, server.ID)
	return nil
}

// Unregister removes a server from the registry
func (r *Registry) Unregister(serverID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	serverUUID, err := uuid.Parse(serverID)
	if err != nil {
		return fmt.Errorf("invalid server ID: %w", err)
	}

	// Remove from database (soft delete)
	err = r.serverModel.Delete(serverUUID)
	if err != nil {
		return fmt.Errorf("failed to remove server from database: %w", err)
	}

	// Remove from in-memory cache
	delete(r.servers, serverID)
	delete(r.stats, serverID)

	log.Printf("Server %s unregistered successfully", serverID)
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

	serverUUID, err := uuid.Parse(serverID)
	if err != nil {
		return fmt.Errorf("invalid server ID: %w", err)
	}

	// Update in database
	err = r.serverModel.UpdateStatus(serverUUID, status)
	if err != nil {
		log.Printf("Failed to update server status in database: %v", err)
		// Continue with in-memory update even if database update fails
	}

	// Update in-memory cache
	if server, exists := r.servers[serverID]; exists {
		server.Status = status
		server.UpdatedAt = time.Now()
		return nil
	}

	return types.NewNotFoundError("Server not found in cache")
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
	r.mu.Lock()
	defer r.mu.Unlock()

	// Clear existing cache
	r.servers = make(map[string]*types.MCPServer)
	r.stats = make(map[string]*types.ServerStats)

	// Query all active servers from database
	rows, err := r.db.Query(`
		SELECT id, organization_id, name, description, protocol, url, command, args,
		       environment, working_dir, version, timeout_seconds, max_retries, status,
		       health_check_url, is_active, metadata, tags, created_at, updated_at
		FROM mcp_servers
		WHERE is_active = true
		ORDER BY name
	`)
	if err != nil {
		return fmt.Errorf("failed to query servers from database: %w", err)
	}
	defer rows.Close()

	loadedCount := 0
	for rows.Next() {
		var server models.MCPServer
		err := rows.Scan(
			&server.ID, &server.OrganizationID, &server.Name, &server.Description,
			&server.Protocol, &server.URL, &server.Command, &server.Args,
			&server.Environment, &server.WorkingDir, &server.Version,
			&server.TimeoutSeconds, &server.MaxRetries, &server.Status,
			&server.HealthCheckURL, &server.IsActive, &server.Metadata,
			&server.Tags, &server.CreatedAt, &server.UpdatedAt,
		)
		if err != nil {
			log.Printf("Failed to scan server row: %v", err)
			continue
		}

		// Convert models.MCPServer to types.MCPServer
		typesServer := r.convertModelToTypes(&server)

		// Add to in-memory cache
		r.servers[typesServer.ID] = typesServer
		r.initializeStats(typesServer.ID)

		loadedCount++
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating over server rows: %w", err)
	}

	log.Printf("Loaded %d active servers from database into registry cache", loadedCount)
	return nil
}

// getAllServers returns all servers in the registry
func (r *Registry) getAllServers() []*types.MCPServer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	servers := make([]*types.MCPServer, 0, len(r.servers))
	for _, server := range r.servers {
		servers = append(servers, server)
	}

	return servers
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

// convertModelToTypes converts models.MCPServer to types.MCPServer
func (r *Registry) convertModelToTypes(server *models.MCPServer) *types.MCPServer {
	result := &types.MCPServer{
		ID:             server.ID.String(),
		OrganizationID: server.OrganizationID.String(),
		Name:           server.Name,
		Protocol:       server.Protocol,
		Version:        server.Version.String,
		Status:         server.Status,
		Timeout:        time.Duration(server.TimeoutSeconds) * time.Second,
		MaxRetries:     server.MaxRetries,
		IsActive:       server.IsActive,
		CreatedAt:      server.CreatedAt,
		UpdatedAt:      server.UpdatedAt,
	}

	if server.Description.Valid {
		result.Description = server.Description.String
	}
	if server.URL.Valid {
		result.URL = server.URL.String
	}
	if server.Command.Valid {
		result.Command = server.Command.String
	}
	if server.WorkingDir.Valid {
		result.WorkingDir = server.WorkingDir.String
	}
	if server.HealthCheckURL.Valid {
		result.HealthCheckURL = server.HealthCheckURL.String
	}

	// Convert arrays
	result.Args = []string(server.Args)
	result.Environment = []string(server.Environment)

	// Convert metadata from map[string]interface{} to map[string]string
	if server.Metadata != nil {
		result.Metadata = make(map[string]string)
		for k, v := range server.Metadata {
			if str, ok := v.(string); ok {
				result.Metadata[k] = str
			}
		}
	}

	return result
}
