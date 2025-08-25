package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"mcp-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// A2AAgentModel handles A2A agent database operations
type A2AAgentModel struct {
	db Database
}

// NewA2AAgentModel creates a new A2A agent model
func NewA2AAgentModel(db Database) *A2AAgentModel {
	return &A2AAgentModel{db: db}
}

// Create inserts a new A2A agent
func (m *A2AAgentModel) Create(agent *types.A2AAgent) error {
	query := `
		INSERT INTO a2a_agents (
			id, organization_id, name, description, endpoint_url, agent_type,
			protocol_version, capabilities, config, auth_type, auth_value,
			is_active, tags, metadata, health_status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING created_at, updated_at`

	capabilitiesJSON, err := json.Marshal(agent.CapabilitiesData)
	if err != nil {
		return fmt.Errorf("failed to marshal capabilities: %w", err)
	}
	agent.Capabilities = capabilitiesJSON

	configJSON, err := json.Marshal(agent.ConfigData)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	agent.Config = configJSON

	metadataJSON, err := json.Marshal(agent.MetadataData)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	agent.Metadata = metadataJSON

	return m.db.QueryRow(
		query,
		agent.ID,
		agent.OrganizationID,
		agent.Name,
		agent.Description,
		agent.EndpointURL,
		agent.AgentType,
		agent.ProtocolVersion,
		capabilitiesJSON,
		configJSON,
		agent.AuthType,
		agent.AuthValue,
		agent.IsActive,
		pq.StringArray(agent.Tags),
		metadataJSON,
		agent.HealthStatus,
	).Scan(&agent.CreatedAt, &agent.UpdatedAt)
}

// GetByID retrieves an A2A agent by ID
func (m *A2AAgentModel) GetByID(id uuid.UUID) (*types.A2AAgent, error) {
	query := `
		SELECT id, organization_id, name, description, endpoint_url, agent_type,
		       protocol_version, capabilities, config, auth_type, auth_value,
		       is_active, tags, metadata, last_health_check, health_status,
		       health_error, created_at, updated_at
		FROM a2a_agents
		WHERE id = $1`

	agent := &types.A2AAgent{}
	var capabilitiesJSON, configJSON, metadataJSON json.RawMessage
	var authValue, healthError sql.NullString

	err := m.db.QueryRow(query, id).Scan(
		&agent.ID,
		&agent.OrganizationID,
		&agent.Name,
		&agent.Description,
		&agent.EndpointURL,
		&agent.AgentType,
		&agent.ProtocolVersion,
		&capabilitiesJSON,
		&configJSON,
		&agent.AuthType,
		&authValue,
		&agent.IsActive,
		(*pq.StringArray)(&agent.Tags),
		&metadataJSON,
		&agent.LastHealthCheck,
		&agent.HealthStatus,
		&healthError,
		&agent.CreatedAt,
		&agent.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("agent not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	// Set raw JSON fields
	agent.Capabilities = capabilitiesJSON
	agent.Config = configJSON
	agent.Metadata = metadataJSON
	if authValue.Valid {
		agent.AuthValue = authValue.String
	}
	if healthError.Valid {
		agent.HealthError = healthError.String
	}

	// Unmarshal JSON data
	if err := json.Unmarshal(capabilitiesJSON, &agent.CapabilitiesData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal capabilities: %w", err)
	}

	if err := json.Unmarshal(configJSON, &agent.ConfigData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &agent.MetadataData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return agent, nil
}

// GetByName retrieves an A2A agent by organization ID and name
func (m *A2AAgentModel) GetByName(orgID uuid.UUID, name string) (*types.A2AAgent, error) {
	query := `
		SELECT id, organization_id, name, description, endpoint_url, agent_type,
		       protocol_version, capabilities, config, auth_type, auth_value,
		       is_active, tags, metadata, last_health_check, health_status,
		       health_error, created_at, updated_at
		FROM a2a_agents
		WHERE organization_id = $1 AND name = $2`

	agent := &types.A2AAgent{}
	var capabilitiesJSON, configJSON, metadataJSON json.RawMessage
	var authValue, healthError sql.NullString

	err := m.db.QueryRow(query, orgID, name).Scan(
		&agent.ID,
		&agent.OrganizationID,
		&agent.Name,
		&agent.Description,
		&agent.EndpointURL,
		&agent.AgentType,
		&agent.ProtocolVersion,
		&capabilitiesJSON,
		&configJSON,
		&agent.AuthType,
		&authValue,
		&agent.IsActive,
		(*pq.StringArray)(&agent.Tags),
		&metadataJSON,
		&agent.LastHealthCheck,
		&agent.HealthStatus,
		&healthError,
		&agent.CreatedAt,
		&agent.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("agent not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	// Set raw JSON fields
	agent.Capabilities = capabilitiesJSON
	agent.Config = configJSON
	agent.Metadata = metadataJSON
	if authValue.Valid {
		agent.AuthValue = authValue.String
	}
	if healthError.Valid {
		agent.HealthError = healthError.String
	}

	// Unmarshal JSON data
	if err := json.Unmarshal(capabilitiesJSON, &agent.CapabilitiesData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal capabilities: %w", err)
	}

	if err := json.Unmarshal(configJSON, &agent.ConfigData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &agent.MetadataData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return agent, nil
}

// List retrieves all A2A agents for an organization with optional filtering
func (m *A2AAgentModel) List(orgID uuid.UUID, filters map[string]interface{}) ([]*types.A2AAgent, error) {
	query := `
		SELECT id, organization_id, name, description, endpoint_url, agent_type,
		       protocol_version, capabilities, config, auth_type, auth_value,
		       is_active, tags, metadata, last_health_check, health_status,
		       health_error, created_at, updated_at
		FROM a2a_agents
		WHERE organization_id = $1`

	args := []interface{}{orgID}
	argIndex := 2

	// Apply filters
	if agentType, ok := filters["agent_type"].(string); ok && agentType != "" {
		query += fmt.Sprintf(" AND agent_type = $%d", argIndex)
		args = append(args, agentType)
		argIndex++
	}

	if isActive, ok := filters["is_active"].(bool); ok {
		query += fmt.Sprintf(" AND is_active = $%d", argIndex)
		args = append(args, isActive)
		argIndex++
	}

	if healthStatus, ok := filters["health_status"].(string); ok && healthStatus != "" {
		query += fmt.Sprintf(" AND health_status = $%d", argIndex)
		args = append(args, healthStatus)
		argIndex++
	}

	if tags, ok := filters["tags"].([]string); ok && len(tags) > 0 {
		query += fmt.Sprintf(" AND tags && $%d", argIndex)
		args = append(args, pq.StringArray(tags))
		argIndex++
	}

	query += " ORDER BY name ASC"

	rows, err := m.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	defer rows.Close()

	var agents []*types.A2AAgent

	for rows.Next() {
		agent := &types.A2AAgent{}
		var capabilitiesJSON, configJSON, metadataJSON json.RawMessage
		var authValue, healthError sql.NullString

		err := rows.Scan(
			&agent.ID,
			&agent.OrganizationID,
			&agent.Name,
			&agent.Description,
			&agent.EndpointURL,
			&agent.AgentType,
			&agent.ProtocolVersion,
			&capabilitiesJSON,
			&configJSON,
			&agent.AuthType,
			&authValue,
			&agent.IsActive,
			(*pq.StringArray)(&agent.Tags),
			&metadataJSON,
			&agent.LastHealthCheck,
			&agent.HealthStatus,
			&healthError,
			&agent.CreatedAt,
			&agent.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}

		// Set raw JSON fields
		agent.Capabilities = capabilitiesJSON
		agent.Config = configJSON
		agent.Metadata = metadataJSON
		if authValue.Valid {
			agent.AuthValue = authValue.String
		}
		if healthError.Valid {
			agent.HealthError = healthError.String
		}

		// Unmarshal JSON data
		if err := json.Unmarshal(capabilitiesJSON, &agent.CapabilitiesData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal capabilities: %w", err)
		}

		if err := json.Unmarshal(configJSON, &agent.ConfigData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &agent.MetadataData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		agents = append(agents, agent)
	}

	return agents, rows.Err()
}

// Update updates an existing A2A agent
func (m *A2AAgentModel) Update(agent *types.A2AAgent) error {
	query := `
		UPDATE a2a_agents
		SET name = $2, description = $3, endpoint_url = $4, agent_type = $5,
		    protocol_version = $6, capabilities = $7, config = $8, auth_type = $9,
		    auth_value = $10, is_active = $11, tags = $12, metadata = $13
		WHERE id = $1
		RETURNING updated_at`

	capabilitiesJSON, err := json.Marshal(agent.CapabilitiesData)
	if err != nil {
		return fmt.Errorf("failed to marshal capabilities: %w", err)
	}
	agent.Capabilities = capabilitiesJSON

	configJSON, err := json.Marshal(agent.ConfigData)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	agent.Config = configJSON

	metadataJSON, err := json.Marshal(agent.MetadataData)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	agent.Metadata = metadataJSON

	return m.db.QueryRow(
		query,
		agent.ID,
		agent.Name,
		agent.Description,
		agent.EndpointURL,
		agent.AgentType,
		agent.ProtocolVersion,
		capabilitiesJSON,
		configJSON,
		agent.AuthType,
		agent.AuthValue,
		agent.IsActive,
		pq.StringArray(agent.Tags),
		metadataJSON,
	).Scan(&agent.UpdatedAt)
}

// Delete removes an A2A agent
func (m *A2AAgentModel) Delete(id uuid.UUID) error {
	query := `DELETE FROM a2a_agents WHERE id = $1`
	result, err := m.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("agent not found: %s", id)
	}

	return nil
}

// Toggle enables or disables an A2A agent
func (m *A2AAgentModel) Toggle(id uuid.UUID, active bool) error {
	query := `
		UPDATE a2a_agents
		SET is_active = $2
		WHERE id = $1
		RETURNING updated_at`

	var updatedAt time.Time
	err := m.db.QueryRow(query, id, active).Scan(&updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("agent not found: %s", id)
		}
		return fmt.Errorf("failed to toggle agent: %w", err)
	}

	return nil
}

// UpdateHealth updates the health status of an A2A agent
func (m *A2AAgentModel) UpdateHealth(id uuid.UUID, status types.A2AHealthStatus, message string) error {
	query := `
		UPDATE a2a_agents
		SET health_status = $2, health_error = $3, last_health_check = NOW()
		WHERE id = $1`

	result, err := m.db.Exec(query, id, status, message)
	if err != nil {
		return fmt.Errorf("failed to update health: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("agent not found: %s", id)
	}

	return nil
}

// ListActive retrieves all active A2A agents for an organization
func (m *A2AAgentModel) ListActive(orgID uuid.UUID) ([]*types.A2AAgent, error) {
	filters := map[string]interface{}{
		"is_active": true,
	}
	return m.List(orgID, filters)
}

// ListByType retrieves all A2A agents of a specific type for an organization
func (m *A2AAgentModel) ListByType(orgID uuid.UUID, agentType types.AgentType) ([]*types.A2AAgent, error) {
	filters := map[string]interface{}{
		"agent_type": string(agentType),
	}
	return m.List(orgID, filters)
}

// A2AAgentToolModel handles A2A agent tool database operations
type A2AAgentToolModel struct {
	db Database
}

// NewA2AAgentToolModel creates a new A2A agent tool model
func NewA2AAgentToolModel(db Database) *A2AAgentToolModel {
	return &A2AAgentToolModel{db: db}
}

// Create inserts a new A2A agent tool mapping
func (m *A2AAgentToolModel) Create(tool *types.A2AAgentTool) error {
	query := `
		INSERT INTO a2a_agent_tools (
			id, agent_id, virtual_server_id, tool_name, tool_config, is_active
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`

	toolConfigJSON, err := json.Marshal(tool.ToolConfigData)
	if err != nil {
		return fmt.Errorf("failed to marshal tool config: %w", err)
	}
	tool.ToolConfig = toolConfigJSON

	return m.db.QueryRow(
		query,
		tool.ID,
		tool.AgentID,
		tool.VirtualServerID,
		tool.ToolName,
		toolConfigJSON,
		tool.IsActive,
	).Scan(&tool.CreatedAt, &tool.UpdatedAt)
}

// GetByAgentAndVirtualServer retrieves agent tools by agent and virtual server
func (m *A2AAgentToolModel) GetByAgentAndVirtualServer(agentID, virtualServerID uuid.UUID) ([]*types.A2AAgentTool, error) {
	query := `
		SELECT id, agent_id, virtual_server_id, tool_name, tool_config, is_active, created_at, updated_at
		FROM a2a_agent_tools
		WHERE agent_id = $1 AND virtual_server_id = $2
		ORDER BY tool_name ASC`

	rows, err := m.db.Query(query, agentID, virtualServerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list agent tools: %w", err)
	}
	defer rows.Close()

	var tools []*types.A2AAgentTool

	for rows.Next() {
		tool := &types.A2AAgentTool{}
		var toolConfigJSON json.RawMessage

		err := rows.Scan(
			&tool.ID,
			&tool.AgentID,
			&tool.VirtualServerID,
			&tool.ToolName,
			&toolConfigJSON,
			&tool.IsActive,
			&tool.CreatedAt,
			&tool.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan agent tool: %w", err)
		}

		tool.ToolConfig = toolConfigJSON
		if err := json.Unmarshal(toolConfigJSON, &tool.ToolConfigData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tool config: %w", err)
		}

		tools = append(tools, tool)
	}

	return tools, rows.Err()
}

// Delete removes an A2A agent tool mapping
func (m *A2AAgentToolModel) Delete(agentID, virtualServerID uuid.UUID, toolName string) error {
	query := `DELETE FROM a2a_agent_tools WHERE agent_id = $1 AND virtual_server_id = $2 AND tool_name = $3`
	result, err := m.db.Exec(query, agentID, virtualServerID, toolName)
	if err != nil {
		return fmt.Errorf("failed to delete agent tool: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("agent tool mapping not found")
	}

	return nil
}

// DeleteByAgent removes all tool mappings for an agent
func (m *A2AAgentToolModel) DeleteByAgent(agentID uuid.UUID) error {
	query := `DELETE FROM a2a_agent_tools WHERE agent_id = $1`
	_, err := m.db.Exec(query, agentID)
	if err != nil {
		return fmt.Errorf("failed to delete agent tools: %w", err)
	}
	return nil
}
