package unit

import (
	"testing"
	"time"

	"mcp-gateway/apps/backend/internal/a2a"
	"mcp-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockA2AAgentModel implements a simple in-memory mock for testing
type mockA2AAgentModel struct {
	agents map[uuid.UUID]*types.A2AAgent
}

func newMockA2AAgentModel() *mockA2AAgentModel {
	return &mockA2AAgentModel{
		agents: make(map[uuid.UUID]*types.A2AAgent),
	}
}

func (m *mockA2AAgentModel) Create(agent *types.A2AAgent) error {
	agent.CreatedAt = time.Now()
	agent.UpdatedAt = agent.CreatedAt
	m.agents[agent.ID] = agent
	return nil
}

func (m *mockA2AAgentModel) GetByID(id uuid.UUID) (*types.A2AAgent, error) {
	if agent, exists := m.agents[id]; exists {
		return agent, nil
	}
	return nil, ErrNotFound
}

func (m *mockA2AAgentModel) GetByName(orgID uuid.UUID, name string) (*types.A2AAgent, error) {
	for _, agent := range m.agents {
		if agent.OrganizationID == orgID && agent.Name == name {
			return agent, nil
		}
	}
	return nil, ErrNotFound
}

func (m *mockA2AAgentModel) List(orgID uuid.UUID, filters map[string]interface{}) ([]*types.A2AAgent, error) {
	var result []*types.A2AAgent

	for _, agent := range m.agents {
		if agent.OrganizationID != orgID {
			continue
		}

		// Apply filters
		match := true
		if agentType, ok := filters["agent_type"].(string); ok && agentType != "" {
			if string(agent.AgentType) != agentType {
				match = false
			}
		}
		if isActive, ok := filters["is_active"].(bool); ok {
			if agent.IsActive != isActive {
				match = false
			}
		}

		if match {
			result = append(result, agent)
		}
	}

	return result, nil
}

func (m *mockA2AAgentModel) Update(agent *types.A2AAgent) error {
	if _, exists := m.agents[agent.ID]; !exists {
		return ErrNotFound
	}
	agent.UpdatedAt = time.Now()
	m.agents[agent.ID] = agent
	return nil
}

func (m *mockA2AAgentModel) Delete(id uuid.UUID) error {
	if _, exists := m.agents[id]; !exists {
		return ErrNotFound
	}
	delete(m.agents, id)
	return nil
}

func (m *mockA2AAgentModel) Toggle(id uuid.UUID, active bool) error {
	agent, exists := m.agents[id]
	if !exists {
		return ErrNotFound
	}
	agent.IsActive = active
	agent.UpdatedAt = time.Now()
	return nil
}

func (m *mockA2AAgentModel) UpdateHealth(id uuid.UUID, status types.A2AHealthStatus, message string) error {
	agent, exists := m.agents[id]
	if !exists {
		return ErrNotFound
	}
	agent.HealthStatus = status
	agent.HealthError = message
	now := time.Now()
	agent.LastHealthCheck = &now
	return nil
}

func TestA2AService_Create(t *testing.T) {
	service := a2a.NewService(nil)

	spec := &types.A2AAgentSpec{
		Name:        "Test Agent",
		Description: "A test AI agent",
		EndpointURL: "https://api.example.com/agent",
		AgentType:   types.AgentTypeCustom,
		AuthType:    types.AuthTypeAPIKey,
		AuthValue:   "test-api-key",
		IsActive:    true,
		Tags:        []string{"test", "ai"},
	}

	agent, err := service.Create(spec)
	require.NoError(t, err)
	require.NotNil(t, agent)

	assert.Equal(t, spec.Name, agent.Name)
	assert.Equal(t, spec.Description, agent.Description)
	assert.Equal(t, spec.EndpointURL, agent.EndpointURL)
	assert.Equal(t, spec.AgentType, agent.AgentType)
	assert.Equal(t, spec.AuthType, agent.AuthType)
	assert.Equal(t, spec.IsActive, agent.IsActive)
	assert.Equal(t, spec.Tags, agent.Tags)
	assert.NotEqual(t, uuid.Nil, agent.ID)
	assert.Equal(t, types.A2AHealthStatusUnknown, agent.HealthStatus)
}

func TestA2AService_Get(t *testing.T) {
	service := a2a.NewService(nil)

	// Create an agent first
	spec := &types.A2AAgentSpec{
		Name:        "Test Agent",
		Description: "A test AI agent",
		EndpointURL: "https://api.example.com/agent",
		AgentType:   types.AgentTypeOpenAI,
		AuthType:    types.AuthTypeAPIKey,
		IsActive:    true,
	}

	created, err := service.Create(spec)
	require.NoError(t, err)

	// Get the agent
	retrieved, err := service.Get(created.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, created.Name, retrieved.Name)
	assert.Equal(t, created.EndpointURL, retrieved.EndpointURL)
}

func TestA2AService_Update(t *testing.T) {
	service := a2a.NewService(nil)

	// Create an agent first
	spec := &types.A2AAgentSpec{
		Name:        "Test Agent",
		Description: "A test AI agent",
		EndpointURL: "https://api.example.com/agent",
		AgentType:   types.AgentTypeCustom,
		AuthType:    types.AuthTypeNone,
		IsActive:    true,
	}

	created, err := service.Create(spec)
	require.NoError(t, err)

	// Update the agent
	updateSpec := &types.A2AAgentSpec{
		Name:        "Updated Test Agent",
		Description: "An updated test AI agent",
		EndpointURL: "https://api.updated.com/agent",
		AgentType:   types.AgentTypeOpenAI,
		AuthType:    types.AuthTypeAPIKey,
		IsActive:    false,
		Tags:        []string{"updated", "test"},
	}

	updated, err := service.Update(created.ID, updateSpec)
	require.NoError(t, err)
	require.NotNil(t, updated)

	assert.Equal(t, updateSpec.Name, updated.Name)
	assert.Equal(t, updateSpec.Description, updated.Description)
	assert.Equal(t, updateSpec.EndpointURL, updated.EndpointURL)
	assert.Equal(t, updateSpec.AgentType, updated.AgentType)
	assert.Equal(t, updateSpec.AuthType, updated.AuthType)
	assert.Equal(t, updateSpec.IsActive, updated.IsActive)
	assert.Equal(t, updateSpec.Tags, updated.Tags)
}

func TestA2AService_Delete(t *testing.T) {
	service := a2a.NewService(nil)

	// Create an agent first
	spec := &types.A2AAgentSpec{
		Name:        "Test Agent",
		Description: "A test AI agent",
		EndpointURL: "https://api.example.com/agent",
		AgentType:   types.AgentTypeCustom,
		IsActive:    true,
	}

	created, err := service.Create(spec)
	require.NoError(t, err)

	// Delete the agent
	err = service.Delete(created.ID)
	require.NoError(t, err)

	// Verify it's gone
	_, err = service.Get(created.ID)
	require.Error(t, err)
}

func TestA2AService_Toggle(t *testing.T) {
	service := a2a.NewService(nil)

	// Create an active agent
	spec := &types.A2AAgentSpec{
		Name:        "Test Agent",
		EndpointURL: "https://api.example.com/agent",
		IsActive:    true,
	}

	created, err := service.Create(spec)
	require.NoError(t, err)
	assert.True(t, created.IsActive)

	// Toggle to inactive
	err = service.Toggle(created.ID, false)
	require.NoError(t, err)

	// Verify the change
	retrieved, err := service.Get(created.ID)
	require.NoError(t, err)
	assert.False(t, retrieved.IsActive)

	// Toggle back to active
	err = service.Toggle(created.ID, true)
	require.NoError(t, err)

	retrieved, err = service.Get(created.ID)
	require.NoError(t, err)
	assert.True(t, retrieved.IsActive)
}

func TestA2AService_List(t *testing.T) {
	service := a2a.NewService(nil)
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	// Create multiple agents
	agents := []*types.A2AAgentSpec{
		{
			Name:        "OpenAI Agent",
			EndpointURL: "https://api.openai.com/v1/chat/completions",
			AgentType:   types.AgentTypeOpenAI,
			IsActive:    true,
		},
		{
			Name:        "Anthropic Agent",
			EndpointURL: "https://api.anthropic.com/v1/messages",
			AgentType:   types.AgentTypeAnthropic,
			IsActive:    true,
		},
		{
			Name:        "Custom Agent",
			EndpointURL: "https://api.custom.com/agent",
			AgentType:   types.AgentTypeCustom,
			IsActive:    false,
		},
	}

	for _, spec := range agents {
		_, err := service.Create(spec)
		require.NoError(t, err)
	}

	// List all agents
	allAgents, err := service.List(orgID, nil)
	require.NoError(t, err)
	assert.Len(t, allAgents, 3)

	// List only active agents
	activeAgents, err := service.List(orgID, map[string]interface{}{
		"is_active": true,
	})
	require.NoError(t, err)
	assert.Len(t, activeAgents, 2)

	// List by agent type
	openaiAgents, err := service.List(orgID, map[string]interface{}{
		"agent_type": string(types.AgentTypeOpenAI),
	})
	require.NoError(t, err)
	assert.Len(t, openaiAgents, 1)
	assert.Equal(t, "OpenAI Agent", openaiAgents[0].Name)
}

func TestA2AService_UpdateHealth(t *testing.T) {
	service := a2a.NewService(nil)

	// Create an agent
	spec := &types.A2AAgentSpec{
		Name:        "Test Agent",
		EndpointURL: "https://api.example.com/agent",
		IsActive:    true,
	}

	created, err := service.Create(spec)
	require.NoError(t, err)
	assert.Equal(t, types.A2AHealthStatusUnknown, created.HealthStatus)
	assert.Nil(t, created.LastHealthCheck)

	// Update health to healthy
	err = service.UpdateHealth(created.ID, types.A2AHealthStatusHealthy, "Agent is responding")
	require.NoError(t, err)

	// Verify the update
	retrieved, err := service.Get(created.ID)
	require.NoError(t, err)
	assert.Equal(t, types.A2AHealthStatusHealthy, retrieved.HealthStatus)
	assert.Equal(t, "Agent is responding", retrieved.HealthError)
	assert.NotNil(t, retrieved.LastHealthCheck)

	// Update health to unhealthy
	err = service.UpdateHealth(created.ID, types.A2AHealthStatusUnhealthy, "Connection timeout")
	require.NoError(t, err)

	retrieved, err = service.Get(created.ID)
	require.NoError(t, err)
	assert.Equal(t, types.A2AHealthStatusUnhealthy, retrieved.HealthStatus)
	assert.Equal(t, "Connection timeout", retrieved.HealthError)
}

func TestA2AService_Stats(t *testing.T) {
	service := a2a.NewService(nil)
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	// Create agents with different types and statuses
	agents := []*types.A2AAgentSpec{
		{Name: "OpenAI 1", EndpointURL: "https://api.openai.com", AgentType: types.AgentTypeOpenAI, IsActive: true},
		{Name: "OpenAI 2", EndpointURL: "https://api.openai.com", AgentType: types.AgentTypeOpenAI, IsActive: false},
		{Name: "Anthropic 1", EndpointURL: "https://api.anthropic.com", AgentType: types.AgentTypeAnthropic, IsActive: true},
		{Name: "Custom 1", EndpointURL: "https://api.custom.com", AgentType: types.AgentTypeCustom, IsActive: true},
	}

	for _, spec := range agents {
		_, err := service.Create(spec)
		require.NoError(t, err)
	}

	// Get stats
	stats, err := service.Stats(orgID)
	require.NoError(t, err)
	require.NotNil(t, stats)

	assert.Equal(t, 4, stats["total"])
	assert.Equal(t, 3, stats["active"])

	byType, ok := stats["by_type"].(map[string]int)
	require.True(t, ok)
	assert.Equal(t, 2, byType["openai"])
	assert.Equal(t, 1, byType["anthropic"])
	assert.Equal(t, 1, byType["custom"])

	byHealth, ok := stats["by_health"].(map[string]int)
	require.True(t, ok)
	assert.Equal(t, 4, byHealth["unknown"]) // All agents start as unknown health
}

// Define a common error for tests
var ErrNotFound = assert.AnError
