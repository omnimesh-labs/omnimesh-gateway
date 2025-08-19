package models

import (
	"encoding/json"
	"time"

	"mcp-gateway/internal/types"
)

// MCPServerModel handles MCP server database operations
type MCPServerModel struct {
	BaseModel
}

// NewMCPServerModel creates a new MCP server model
func NewMCPServerModel(db Database) *MCPServerModel {
	return &MCPServerModel{
		BaseModel: BaseModel{db: db},
	}
}

// Create inserts a new MCP server into the database
func (m *MCPServerModel) Create(server *types.MCPServer) error {
	query := `
		INSERT INTO mcp_servers (id, organization_id, name, description, url, protocol, version, 
			status, weight, metadata, health_check_url, timeout, max_retries, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	now := time.Now()
	server.CreatedAt = now
	server.UpdatedAt = now

	metadata, _ := json.Marshal(server.Metadata)

	_, err := m.db.Exec(query, server.ID, server.OrganizationID, server.Name, server.Description,
		server.URL, server.Protocol, server.Version, server.Status, server.Weight,
		string(metadata), server.HealthCheckURL, server.Timeout, server.MaxRetries,
		server.IsActive, server.CreatedAt, server.UpdatedAt)

	return err
}

// GetByID retrieves an MCP server by ID
func (m *MCPServerModel) GetByID(id string) (*types.MCPServer, error) {
	query := `
		SELECT id, organization_id, name, description, url, protocol, version, status, weight,
			metadata, health_check_url, timeout, max_retries, is_active, created_at, updated_at
		FROM mcp_servers
		WHERE id = $1
	`

	server := &types.MCPServer{}
	var metadataJSON string
	var timeout int64

	err := m.db.QueryRow(query, id).Scan(
		&server.ID, &server.OrganizationID, &server.Name, &server.Description,
		&server.URL, &server.Protocol, &server.Version, &server.Status, &server.Weight,
		&metadataJSON, &server.HealthCheckURL, &timeout, &server.MaxRetries,
		&server.IsActive, &server.CreatedAt, &server.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse metadata JSON
	if metadataJSON != "" {
		json.Unmarshal([]byte(metadataJSON), &server.Metadata)
	}

	server.Timeout = time.Duration(timeout)

	return server, nil
}

// ListByOrganization lists MCP servers for an organization
func (m *MCPServerModel) ListByOrganization(orgID string, activeOnly bool) ([]*types.MCPServer, error) {
	query := `
		SELECT id, organization_id, name, description, url, protocol, version, status, weight,
			metadata, health_check_url, timeout, max_retries, is_active, created_at, updated_at
		FROM mcp_servers
		WHERE organization_id = $1
	`

	args := []interface{}{orgID}
	if activeOnly {
		query += " AND is_active = true"
	}
	query += " ORDER BY created_at DESC"

	rows, err := m.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []*types.MCPServer
	for rows.Next() {
		server := &types.MCPServer{}
		var metadataJSON string
		var timeout int64

		err := rows.Scan(
			&server.ID, &server.OrganizationID, &server.Name, &server.Description,
			&server.URL, &server.Protocol, &server.Version, &server.Status, &server.Weight,
			&metadataJSON, &server.HealthCheckURL, &timeout, &server.MaxRetries,
			&server.IsActive, &server.CreatedAt, &server.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse metadata JSON
		if metadataJSON != "" {
			json.Unmarshal([]byte(metadataJSON), &server.Metadata)
		}

		server.Timeout = time.Duration(timeout)
		servers = append(servers, server)
	}

	return servers, nil
}

// Update updates an MCP server in the database
func (m *MCPServerModel) Update(server *types.MCPServer) error {
	query := `
		UPDATE mcp_servers
		SET name = $1, description = $2, url = $3, protocol = $4, version = $5,
			weight = $6, metadata = $7, health_check_url = $8, timeout = $9,
			max_retries = $10, updated_at = $11
		WHERE id = $12
	`

	server.UpdatedAt = time.Now()
	metadata, _ := json.Marshal(server.Metadata)

	_, err := m.db.Exec(query, server.Name, server.Description, server.URL,
		server.Protocol, server.Version, server.Weight, string(metadata),
		server.HealthCheckURL, server.Timeout, server.MaxRetries,
		server.UpdatedAt, server.ID)

	return err
}

// UpdateStatus updates the status of an MCP server
func (m *MCPServerModel) UpdateStatus(id, status string) error {
	query := `
		UPDATE mcp_servers
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := m.db.Exec(query, status, time.Now(), id)
	return err
}

// Delete soft deletes an MCP server
func (m *MCPServerModel) Delete(id string) error {
	query := `
		UPDATE mcp_servers
		SET is_active = false, updated_at = $1
		WHERE id = $2
	`

	_, err := m.db.Exec(query, time.Now(), id)
	return err
}

// GetHealthyServers retrieves all healthy servers for an organization
func (m *MCPServerModel) GetHealthyServers(orgID string) ([]*types.MCPServer, error) {
	query := `
		SELECT id, organization_id, name, description, url, protocol, version, status, weight,
			metadata, health_check_url, timeout, max_retries, is_active, created_at, updated_at
		FROM mcp_servers
		WHERE organization_id = $1 AND is_active = true AND status = $2
		ORDER BY weight DESC, created_at DESC
	`

	rows, err := m.db.Query(query, orgID, types.ServerStatusActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []*types.MCPServer
	for rows.Next() {
		server := &types.MCPServer{}
		var metadataJSON string
		var timeout int64

		err := rows.Scan(
			&server.ID, &server.OrganizationID, &server.Name, &server.Description,
			&server.URL, &server.Protocol, &server.Version, &server.Status, &server.Weight,
			&metadataJSON, &server.HealthCheckURL, &timeout, &server.MaxRetries,
			&server.IsActive, &server.CreatedAt, &server.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse metadata JSON
		if metadataJSON != "" {
			json.Unmarshal([]byte(metadataJSON), &server.Metadata)
		}

		server.Timeout = time.Duration(timeout)
		servers = append(servers, server)
	}

	return servers, nil
}

// HealthCheckModel handles health check database operations
type HealthCheckModel struct {
	BaseModel
}

// NewHealthCheckModel creates a new health check model
func NewHealthCheckModel(db Database) *HealthCheckModel {
	return &HealthCheckModel{
		BaseModel: BaseModel{db: db},
	}
}

// Create inserts a new health check record
func (m *HealthCheckModel) Create(check *types.HealthCheck) error {
	query := `
		INSERT INTO health_checks (id, server_id, status, response, latency, error, checked_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := m.db.Exec(query, check.ID, check.ServerID, check.Status,
		check.Response, check.Latency, check.Error, check.CheckedAt)

	return err
}

// GetLatestByServerID retrieves the latest health check for a server
func (m *HealthCheckModel) GetLatestByServerID(serverID string) (*types.HealthCheck, error) {
	query := `
		SELECT id, server_id, status, response, latency, error, checked_at
		FROM health_checks
		WHERE server_id = $1
		ORDER BY checked_at DESC
		LIMIT 1
	`

	check := &types.HealthCheck{}
	err := m.db.QueryRow(query, serverID).Scan(
		&check.ID, &check.ServerID, &check.Status, &check.Response,
		&check.Latency, &check.Error, &check.CheckedAt,
	)

	if err != nil {
		return nil, err
	}

	return check, nil
}

// GetHistoryByServerID retrieves health check history for a server
func (m *HealthCheckModel) GetHistoryByServerID(serverID string, limit int) ([]*types.HealthCheck, error) {
	query := `
		SELECT id, server_id, status, response, latency, error, checked_at
		FROM health_checks
		WHERE server_id = $1
		ORDER BY checked_at DESC
		LIMIT $2
	`

	rows, err := m.db.Query(query, serverID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checks []*types.HealthCheck
	for rows.Next() {
		check := &types.HealthCheck{}
		err := rows.Scan(
			&check.ID, &check.ServerID, &check.Status, &check.Response,
			&check.Latency, &check.Error, &check.CheckedAt,
		)
		if err != nil {
			return nil, err
		}
		checks = append(checks, check)
	}

	return checks, nil
}
