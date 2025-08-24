package a2a

import (
	"database/sql"
	"fmt"
	"sync"

	"mcp-gateway/apps/backend/internal/database/models"
	"mcp-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
)

// Service manages A2A agents and provides business logic
type Service struct {
	db             *sql.DB
	agentModel     *models.A2AAgentModel
	agentToolModel *models.A2AAgentToolModel
	cache          *sync.Map // In-memory cache for performance
	mu             sync.RWMutex
}

// dbWrapper wraps *sql.DB to implement the Database interface
type dbWrapper struct {
	*sql.DB
}

// NewService creates a new A2A service
func NewService(db *sql.DB) *Service {
	dbWrap := &dbWrapper{db}
	return &Service{
		db:             db,
		agentModel:     models.NewA2AAgentModel(dbWrap),
		agentToolModel: models.NewA2AAgentToolModel(dbWrap),
		cache:          &sync.Map{},
	}
}

// Create creates a new A2A agent
func (s *Service) Create(spec *types.A2AAgentSpec) (*types.A2AAgent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate ID if not provided
	var agentID uuid.UUID
	var err error
	if spec.ID == "" {
		agentID = uuid.New()
	} else {
		agentID, err = uuid.Parse(spec.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid agent ID: %w", err)
		}
	}

	// Set defaults based on agent type
	capabilities := spec.Capabilities
	if capabilities == nil {
		capabilities = types.DefaultAgentCapabilities[spec.AgentType]
	}

	config := spec.Config
	if config == nil {
		config = types.DefaultAgentConfigs[spec.AgentType]
	}

	// Convert spec to database model
	agent := &types.A2AAgent{
		ID:               agentID,
		OrganizationID:   uuid.MustParse("00000000-0000-0000-0000-000000000000"), // Default org for single-tenant
		Name:             spec.Name,
		Description:      spec.Description,
		EndpointURL:      spec.EndpointURL,
		AgentType:        spec.AgentType,
		ProtocolVersion:  spec.ProtocolVersion,
		CapabilitiesData: capabilities,
		ConfigData:       config,
		AuthType:         spec.AuthType,
		AuthValue:        spec.AuthValue, // This will be encrypted by the model
		IsActive:         spec.IsActive,
		Tags:             spec.Tags,
		MetadataData:     spec.Metadata,
		HealthStatus:     types.A2AHealthStatusUnknown,
	}

	// Set defaults
	if agent.ProtocolVersion == "" {
		agent.ProtocolVersion = "1.0"
	}
	if agent.AgentType == "" {
		agent.AgentType = types.AgentTypeCustom
	}
	if agent.AuthType == "" {
		agent.AuthType = types.AuthTypeNone
	}
	if agent.Tags == nil {
		agent.Tags = []string{}
	}
	if agent.CapabilitiesData == nil {
		agent.CapabilitiesData = make(map[string]interface{})
	}
	if agent.ConfigData == nil {
		agent.ConfigData = make(map[string]interface{})
	}
	if agent.MetadataData == nil {
		agent.MetadataData = make(map[string]interface{})
	}

	// Persist to database
	if err := s.agentModel.Create(agent); err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// Update cache
	cacheKey := fmt.Sprintf("agent:%s", agent.ID.String())
	s.cache.Store(cacheKey, agent)

	return agent, nil
}

// Get retrieves an A2A agent by ID
func (s *Service) Get(id uuid.UUID) (*types.A2AAgent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check cache first
	cacheKey := fmt.Sprintf("agent:%s", id.String())
	if cached, exists := s.cache.Load(cacheKey); exists {
		if agent, ok := cached.(*types.A2AAgent); ok {
			return agent, nil
		}
	}

	// Load from database
	agent, err := s.agentModel.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update cache
	s.cache.Store(cacheKey, agent)

	return agent, nil
}

// GetByName retrieves an A2A agent by organization ID and name
func (s *Service) GetByName(orgID uuid.UUID, name string) (*types.A2AAgent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// For single-tenant mode, use default org
	if orgID == uuid.Nil {
		orgID = uuid.MustParse("00000000-0000-0000-0000-000000000000")
	}

	// Check cache by name
	cacheKey := fmt.Sprintf("agent_name:%s:%s", orgID.String(), name)
	if cached, exists := s.cache.Load(cacheKey); exists {
		if agent, ok := cached.(*types.A2AAgent); ok {
			return agent, nil
		}
	}

	// Load from database
	agent, err := s.agentModel.GetByName(orgID, name)
	if err != nil {
		return nil, err
	}

	// Update cache
	s.cache.Store(cacheKey, agent)
	s.cache.Store(fmt.Sprintf("agent:%s", agent.ID.String()), agent)

	return agent, nil
}

// List retrieves all A2A agents with optional filtering
func (s *Service) List(orgID uuid.UUID, filters map[string]interface{}) ([]*types.A2AAgent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// For single-tenant mode, use default org
	if orgID == uuid.Nil {
		orgID = uuid.MustParse("00000000-0000-0000-0000-000000000000")
	}

	// Load from database
	agents, err := s.agentModel.List(orgID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}

	// Update cache for each agent
	for _, agent := range agents {
		cacheKey := fmt.Sprintf("agent:%s", agent.ID.String())
		s.cache.Store(cacheKey, agent)
		nameKey := fmt.Sprintf("agent_name:%s:%s", agent.OrganizationID.String(), agent.Name)
		s.cache.Store(nameKey, agent)
	}

	return agents, nil
}

// Update updates an existing A2A agent
func (s *Service) Update(id uuid.UUID, spec *types.A2AAgentSpec) (*types.A2AAgent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get existing agent
	existing, err := s.agentModel.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing agent: %w", err)
	}

	// Update fields from spec
	if spec.Name != "" {
		existing.Name = spec.Name
	}
	if spec.Description != "" {
		existing.Description = spec.Description
	}
	if spec.EndpointURL != "" {
		existing.EndpointURL = spec.EndpointURL
	}
	if spec.AgentType != "" {
		existing.AgentType = spec.AgentType
	}
	if spec.ProtocolVersion != "" {
		existing.ProtocolVersion = spec.ProtocolVersion
	}
	if spec.Capabilities != nil {
		existing.CapabilitiesData = spec.Capabilities
	}
	if spec.Config != nil {
		existing.ConfigData = spec.Config
	}
	if spec.AuthType != "" {
		existing.AuthType = spec.AuthType
	}
	if spec.AuthValue != "" {
		existing.AuthValue = spec.AuthValue
	}
	if spec.Tags != nil {
		existing.Tags = spec.Tags
	}
	if spec.Metadata != nil {
		existing.MetadataData = spec.Metadata
	}

	// Always update active status from spec
	existing.IsActive = spec.IsActive

	// Persist to database
	if err := s.agentModel.Update(existing); err != nil {
		return nil, fmt.Errorf("failed to update agent: %w", err)
	}

	// Clear cache for this agent
	s.clearAgentCache(existing)

	return existing, nil
}

// Delete removes an A2A agent
func (s *Service) Delete(id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get agent first to clear cache
	agent, err := s.agentModel.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get agent for deletion: %w", err)
	}

	// Delete associated tools first
	if err := s.agentToolModel.DeleteByAgent(id); err != nil {
		return fmt.Errorf("failed to delete agent tools: %w", err)
	}

	// Delete from database
	if err := s.agentModel.Delete(id); err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}

	// Clear cache
	s.clearAgentCache(agent)

	return nil
}

// Toggle enables or disables an A2A agent
func (s *Service) Toggle(id uuid.UUID, active bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Toggle in database
	if err := s.agentModel.Toggle(id, active); err != nil {
		return fmt.Errorf("failed to toggle agent: %w", err)
	}

	// Clear cache to force reload
	cacheKey := fmt.Sprintf("agent:%s", id.String())
	s.cache.Delete(cacheKey)

	return nil
}

// UpdateHealth updates the health status of an A2A agent
func (s *Service) UpdateHealth(id uuid.UUID, status types.A2AHealthStatus, message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Update health in database
	if err := s.agentModel.UpdateHealth(id, status, message); err != nil {
		return fmt.Errorf("failed to update agent health: %w", err)
	}

	// Clear cache to force reload
	cacheKey := fmt.Sprintf("agent:%s", id.String())
	s.cache.Delete(cacheKey)

	return nil
}

// ListActive retrieves all active A2A agents for an organization
func (s *Service) ListActive(orgID uuid.UUID) ([]*types.A2AAgent, error) {
	filters := map[string]interface{}{
		"is_active": true,
	}
	return s.List(orgID, filters)
}

// ListByType retrieves all A2A agents of a specific type
func (s *Service) ListByType(orgID uuid.UUID, agentType types.AgentType) ([]*types.A2AAgent, error) {
	filters := map[string]interface{}{
		"agent_type": string(agentType),
	}
	return s.List(orgID, filters)
}

// RegisterTool registers a tool from an A2A agent to a virtual server
func (s *Service) RegisterTool(agentID, virtualServerID uuid.UUID, toolName string, config map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if config == nil {
		config = make(map[string]interface{})
	}

	tool := &types.A2AAgentTool{
		ID:               uuid.New(),
		AgentID:          agentID,
		VirtualServerID:  virtualServerID,
		ToolName:         toolName,
		ToolConfigData:   config,
		IsActive:         true,
	}

	return s.agentToolModel.Create(tool)
}

// UnregisterTool removes a tool mapping from an A2A agent to a virtual server
func (s *Service) UnregisterTool(agentID, virtualServerID uuid.UUID, toolName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.agentToolModel.Delete(agentID, virtualServerID, toolName)
}

// GetAgentTools retrieves all tools registered for an agent and virtual server
func (s *Service) GetAgentTools(agentID, virtualServerID uuid.UUID) ([]*types.A2AAgentTool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.agentToolModel.GetByAgentAndVirtualServer(agentID, virtualServerID)
}

// ClearCache clears the in-memory cache
func (s *Service) ClearCache() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache = &sync.Map{}
}

// clearAgentCache clears cache entries for a specific agent
func (s *Service) clearAgentCache(agent *types.A2AAgent) {
	cacheKey := fmt.Sprintf("agent:%s", agent.ID.String())
	s.cache.Delete(cacheKey)

	nameKey := fmt.Sprintf("agent_name:%s:%s", agent.OrganizationID.String(), agent.Name)
	s.cache.Delete(nameKey)
}

// LoadCache preloads the cache with all A2A agents
func (s *Service) LoadCache() error {
	// Use default org for single-tenant mode
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	agents, err := s.List(orgID, nil)
	if err != nil {
		return fmt.Errorf("failed to load cache: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, agent := range agents {
		cacheKey := fmt.Sprintf("agent:%s", agent.ID.String())
		s.cache.Store(cacheKey, agent)

		nameKey := fmt.Sprintf("agent_name:%s:%s", agent.OrganizationID.String(), agent.Name)
		s.cache.Store(nameKey, agent)
	}

	return nil
}

// Stats returns statistics about A2A agents
func (s *Service) Stats(orgID uuid.UUID) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// For single-tenant mode, use default org
	if orgID == uuid.Nil {
		orgID = uuid.MustParse("00000000-0000-0000-0000-000000000000")
	}

	stats := make(map[string]interface{})

	// Total agents
	allAgents, err := s.List(orgID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get total agents: %w", err)
	}
	stats["total"] = len(allAgents)

	// Active agents
	activeAgents, err := s.ListActive(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active agents: %w", err)
	}
	stats["active"] = len(activeAgents)

	// Agents by type
	agentTypes := make(map[string]int)
	healthStatuses := make(map[string]int)

	for _, agent := range allAgents {
		agentTypes[string(agent.AgentType)]++
		healthStatuses[string(agent.HealthStatus)]++
	}

	stats["by_type"] = agentTypes
	stats["by_health"] = healthStatuses

	return stats, nil
}