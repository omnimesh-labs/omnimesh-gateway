package discovery

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"mcp-gateway/apps/backend/internal/database/models"
	"mcp-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Service handles MCP server discovery and management
type Service struct {
	db       *sql.DB
	models   *Models
	config   *Config
	registry *Registry
	health   *HealthChecker
	mu       sync.RWMutex
	stopCh   map[uuid.UUID]chan struct{} // For stopping health checks
}

// Models contains all database models used by the discovery service
type Models struct {
	MCPServer   *models.MCPServerModel
	HealthCheck *models.HealthCheckModel
}

// Config holds discovery service configuration
type Config struct {
	Enabled          bool
	HealthInterval   time.Duration
	FailureThreshold int
	RecoveryTimeout  time.Duration
	SingleTenant     bool // If true, use a default organization for single-tenant mode
}

// Default organization UUID for single-tenant mode (matches migration)
var DefaultOrganizationID = uuid.MustParse("00000000-0000-0000-0000-000000000000")

// dbWrapper wraps *sql.DB to implement the Database interface
type dbWrapper struct {
	*sql.DB
}

// NewService creates a new discovery service
func NewService(db *sql.DB, config *Config) *Service {
	// Wrap the database to implement the Database interface
	dbWrap := &dbWrapper{db}

	service := &Service{
		db:     db,
		config: config,
		models: &Models{
			MCPServer:   models.NewMCPServerModel(dbWrap),
			HealthCheck: models.NewHealthCheckModel(dbWrap),
		},
		stopCh: make(map[uuid.UUID]chan struct{}),
	}

	service.registry = NewRegistry(db)
	service.health = NewHealthChecker(service.registry, config)

	return service
}

// RegisterServer registers a new MCP server
func (s *Service) RegisterServer(orgID string, req *types.CreateMCPServerRequest) (*types.MCPServer, error) {
	// Resolve organization ID (handles single-tenant mode)
	orgUUID, err := s.resolveOrganizationID(orgID)
	if err != nil {
		return nil, err
	}

	// Check if server with same name already exists in organization
	existing, err := s.models.MCPServer.GetByName(orgUUID, req.Name)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check for existing server: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("server with name '%s' already exists in organization", req.Name)
	}

	// Convert request to model
	server := &models.MCPServer{
		ID:             uuid.New(),
		OrganizationID: orgUUID,
		Name:           req.Name,
		Description:    sql.NullString{String: req.Description, Valid: req.Description != ""},
		Protocol:       req.Protocol,
		URL:            sql.NullString{String: req.URL, Valid: req.URL != ""},
		Command:        sql.NullString{String: req.Command, Valid: req.Command != ""},
		Args:           pq.StringArray(req.Args),
		Environment:    pq.StringArray(req.Environment),
		WorkingDir:     sql.NullString{String: req.WorkingDir, Valid: req.WorkingDir != ""},
		Version:        sql.NullString{String: req.Version, Valid: req.Version != ""},
		Weight:         req.Weight,
		TimeoutSeconds: int(req.Timeout.Seconds()),
		MaxRetries:     req.MaxRetries,
		Status:         types.ServerStatusInactive, // Start as inactive
		HealthCheckURL: sql.NullString{String: req.HealthCheckURL, Valid: req.HealthCheckURL != ""},
		IsActive:       true,
		Metadata:       convertStringMapToInterface(req.Metadata),
		Tags:           pq.StringArray{}, // Initialize empty tags
	}

	// Set default values if not provided
	if server.Weight == 0 {
		server.Weight = 100
	}
	if server.TimeoutSeconds == 0 {
		server.TimeoutSeconds = 30
	}
	if server.MaxRetries == 0 {
		server.MaxRetries = 3
	}

	// Create server in database
	err = s.models.MCPServer.Create(server)
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	// Start health checking for the server
	go s.startHealthChecking(server.ID)

	// Convert back to types.MCPServer
	return convertModelToTypesMCPServer(server), nil
}

// UnregisterServer removes an MCP server
func (s *Service) UnregisterServer(serverID string) error {
	// Validate server ID
	serverUUID, err := uuid.Parse(serverID)
	if err != nil {
		return fmt.Errorf("invalid server ID: %w", err)
	}

	// Check if server exists
	server, err := s.models.MCPServer.GetByID(serverUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("server not found")
		}
		return fmt.Errorf("failed to get server: %w", err)
	}

	// Stop health checking
	s.stopHealthChecking(serverUUID)

	// Soft delete the server (set is_active = false)
	err = s.models.MCPServer.Delete(serverUUID)
	if err != nil {
		return fmt.Errorf("failed to delete server: %w", err)
	}

	log.Printf("Server %s (%s) unregistered successfully", server.Name, serverUUID)
	return nil
}

// GetServer retrieves a server by ID
func (s *Service) GetServer(serverID string) (*types.MCPServer, error) {
	// Validate server ID
	serverUUID, err := uuid.Parse(serverID)
	if err != nil {
		return nil, fmt.Errorf("invalid server ID: %w", err)
	}

	// Get server from database
	server, err := s.models.MCPServer.GetByID(serverUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("server not found")
		}
		return nil, fmt.Errorf("failed to get server: %w", err)
	}

	// Convert to types.MCPServer
	return convertModelToTypesMCPServer(server), nil
}

// ListServers returns all servers for an organization
func (s *Service) ListServers(orgID string) ([]*types.MCPServer, error) {
	// Resolve organization ID (handles single-tenant mode)
	orgUUID, err := s.resolveOrganizationID(orgID)
	if err != nil {
		return nil, err
	}

	// Get servers from database
	servers, err := s.models.MCPServer.ListByOrganization(orgUUID, true) // Only active servers
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}

	// Convert to types.MCPServer slice
	result := make([]*types.MCPServer, len(servers))
	for i, server := range servers {
		result[i] = convertModelToTypesMCPServer(server)
	}

	return result, nil
}

// GetHealthyServers returns all healthy servers for an organization
func (s *Service) GetHealthyServers(orgID string) ([]*types.MCPServer, error) {
	// Resolve organization ID (handles single-tenant mode)
	orgUUID, err := s.resolveOrganizationID(orgID)
	if err != nil {
		return nil, err
	}

	// Get active servers from database
	servers, err := s.models.MCPServer.GetActiveServers(orgUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active servers: %w", err)
	}

	// Convert to types.MCPServer slice
	result := make([]*types.MCPServer, len(servers))
	for i, server := range servers {
		result[i] = convertModelToTypesMCPServer(server)
	}

	return result, nil
}

// UpdateServer updates server configuration
func (s *Service) UpdateServer(serverID string, req *types.UpdateMCPServerRequest) (*types.MCPServer, error) {
	// Validate server ID
	serverUUID, err := uuid.Parse(serverID)
	if err != nil {
		return nil, fmt.Errorf("invalid server ID: %w", err)
	}

	// Get existing server
	server, err := s.models.MCPServer.GetByID(serverUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("server not found")
		}
		return nil, fmt.Errorf("failed to get server: %w", err)
	}

	// Update fields if provided
	if req.Name != "" {
		// Check if name conflicts with another server in the same organization
		existing, err := s.models.MCPServer.GetByName(server.OrganizationID, req.Name)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to check for existing server: %w", err)
		}
		if existing != nil && existing.ID != server.ID {
			return nil, fmt.Errorf("server with name '%s' already exists in organization", req.Name)
		}
		server.Name = req.Name
	}
	if req.Description != "" {
		server.Description = sql.NullString{String: req.Description, Valid: true}
	}
	if req.URL != "" {
		server.URL = sql.NullString{String: req.URL, Valid: true}
	}
	if req.Protocol != "" {
		server.Protocol = req.Protocol
	}
	if req.Version != "" {
		server.Version = sql.NullString{String: req.Version, Valid: true}
	}
	if req.Weight > 0 {
		server.Weight = req.Weight
	}
	if req.Metadata != nil {
		server.Metadata = convertStringMapToInterface(req.Metadata)
	}
	if req.HealthCheckURL != "" {
		server.HealthCheckURL = sql.NullString{String: req.HealthCheckURL, Valid: true}
	}
	if req.Timeout > 0 {
		server.TimeoutSeconds = int(req.Timeout.Seconds())
	}
	if req.MaxRetries > 0 {
		server.MaxRetries = req.MaxRetries
	}
	if req.IsActive != nil {
		server.IsActive = *req.IsActive
	}
	if req.Command != "" {
		server.Command = sql.NullString{String: req.Command, Valid: true}
	}
	if req.Args != nil {
		server.Args = pq.StringArray(req.Args)
	}
	if req.Environment != nil {
		server.Environment = pq.StringArray(req.Environment)
	}
	if req.WorkingDir != "" {
		server.WorkingDir = sql.NullString{String: req.WorkingDir, Valid: true}
	}

	// Update server in database
	err = s.models.MCPServer.Update(server)
	if err != nil {
		return nil, fmt.Errorf("failed to update server: %w", err)
	}

	// If server was reactivated, restart health checking
	if req.IsActive != nil && *req.IsActive {
		go s.startHealthChecking(server.ID)
	} else if req.IsActive != nil && !*req.IsActive {
		s.stopHealthChecking(server.ID)
	}

	// Convert back to types.MCPServer
	return convertModelToTypesMCPServer(server), nil
}

// GetServerStats returns server statistics
func (s *Service) GetServerStats(serverID string) (*types.LoadBalancerStats, error) {
	// TODO: Implement server statistics retrieval
	return nil, nil
}

// Start starts the discovery service
func (s *Service) Start() error {
	if !s.config.Enabled {
		log.Println("Discovery service is disabled")
		return nil
	}

	log.Println("Starting discovery service...")

	// Load all active servers from database and start health checking
	// We'll load servers for all organizations - in a real implementation
	// you might want to filter by specific organizations
	return nil
}

// Stop stops the discovery service
func (s *Service) Stop() error {
	log.Println("Stopping discovery service...")

	s.mu.Lock()
	defer s.mu.Unlock()

	// Stop all health checking goroutines
	for serverID, stopCh := range s.stopCh {
		close(stopCh)
		log.Printf("Stopped health checking for server %s", serverID)
	}

	// Clear the stop channels map
	s.stopCh = make(map[uuid.UUID]chan struct{})

	log.Println("Discovery service stopped")
	return nil
}

// Helper functions

// resolveOrganizationID resolves the organization ID for single-tenant or multi-tenant mode
func (s *Service) resolveOrganizationID(orgID string) (uuid.UUID, error) {
	// For single-tenant mode, always use the default organization
	if s.config.SingleTenant {
		return DefaultOrganizationID, nil
	}

	// For multi-tenant mode, parse the provided organization ID
	if orgID == "" {
		return uuid.Nil, fmt.Errorf("organization ID is required in multi-tenant mode")
	}

	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	return orgUUID, nil
}

// Helper functions for type conversion

// convertModelToTypesMCPServer converts models.MCPServer to types.MCPServer
func convertModelToTypesMCPServer(server *models.MCPServer) *types.MCPServer {
	result := &types.MCPServer{
		ID:             server.ID.String(),
		OrganizationID: server.OrganizationID.String(),
		Name:           server.Name,
		Protocol:       server.Protocol,
		Version:        server.Version.String,
		Status:         server.Status,
		Weight:         server.Weight,
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

// convertStringMapToInterface converts map[string]string to map[string]interface{}
func convertStringMapToInterface(m map[string]string) map[string]interface{} {
	if m == nil {
		return nil
	}
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = v
	}
	return result
}

// Health checking functions

// startHealthChecking starts health checking for a server
func (s *Service) startHealthChecking(serverID uuid.UUID) {
	if !s.config.Enabled {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Stop existing health check if running
	if stopCh, exists := s.stopCh[serverID]; exists {
		close(stopCh)
	}

	// Start new health check
	stopCh := make(chan struct{})
	s.stopCh[serverID] = stopCh

	go func() {
		ticker := time.NewTicker(s.config.HealthInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.performHealthCheck(serverID)
			case <-stopCh:
				return
			}
		}
	}()
}

// stopHealthChecking stops health checking for a server
func (s *Service) stopHealthChecking(serverID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if stopCh, exists := s.stopCh[serverID]; exists {
		close(stopCh)
		delete(s.stopCh, serverID)
	}
}

// performHealthCheck performs a health check on a server
func (s *Service) performHealthCheck(serverID uuid.UUID) {
	// Get server details
	server, err := s.models.MCPServer.GetByID(serverID)
	if err != nil {
		log.Printf("Failed to get server %s for health check: %v", serverID, err)
		return
	}

	// Skip if server is not active
	if !server.IsActive {
		return
	}

	// Create health check record
	check := &models.HealthCheck{
		ServerID:  serverID,
		CheckedAt: time.Now(),
	}

	// Perform the actual health check based on protocol
	status := s.checkServerHealth(server)
	check.Status = status

	// Update server status if needed
	if status != server.Status {
		err = s.models.MCPServer.UpdateStatus(serverID, status)
		if err != nil {
			log.Printf("Failed to update server %s status: %v", serverID, err)
		}
	}

	// Save health check record
	err = s.models.HealthCheck.Create(check)
	if err != nil {
		log.Printf("Failed to save health check for server %s: %v", serverID, err)
	}
}

// checkServerHealth performs the actual health check logic
func (s *Service) checkServerHealth(server *models.MCPServer) string {
	// TODO: Implement actual health checking logic based on protocol
	// For now, just return healthy status (matching health_status_enum)
	return "healthy"
}
