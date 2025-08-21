package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// MCPServer represents the mcp_servers table from the ERD
type MCPServer struct {
	ID             uuid.UUID              `db:"id" json:"id"`
	OrganizationID uuid.UUID              `db:"organization_id" json:"organization_id"`
	Name           string                 `db:"name" json:"name"`
	Description    sql.NullString         `db:"description" json:"description,omitempty"`
	Protocol       string                 `db:"protocol" json:"protocol"` // protocol_enum
	URL            sql.NullString         `db:"url" json:"url,omitempty"`
	Command        sql.NullString         `db:"command" json:"command,omitempty"`
	Args           pq.StringArray         `db:"args" json:"args,omitempty"`
	Environment    pq.StringArray         `db:"environment" json:"environment,omitempty"`
	WorkingDir     sql.NullString         `db:"working_dir" json:"working_dir,omitempty"`
	Version        sql.NullString         `db:"version" json:"version,omitempty"`
	Weight         int                    `db:"weight" json:"weight"`
	TimeoutSeconds int                    `db:"timeout_seconds" json:"timeout_seconds"`
	MaxRetries     int                    `db:"max_retries" json:"max_retries"`
	Status         string                 `db:"status" json:"status"` // server_status_enum
	HealthCheckURL sql.NullString         `db:"health_check_url" json:"health_check_url,omitempty"`
	IsActive       bool                   `db:"is_active" json:"is_active"`
	Metadata       map[string]interface{} `db:"metadata" json:"metadata,omitempty"`
	Tags           pq.StringArray         `db:"tags" json:"tags,omitempty"`
	CreatedAt      time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time              `db:"updated_at" json:"updated_at"`
}

// MCPServerModel handles MCP server database operations
type MCPServerModel struct {
	db Database
}

// NewMCPServerModel creates a new MCP server model
func NewMCPServerModel(db Database) *MCPServerModel {
	return &MCPServerModel{db: db}
}

// Create inserts a new MCP server
func (m *MCPServerModel) Create(server *MCPServer) error {
	query := `
		INSERT INTO mcp_servers (
			id, organization_id, name, description, protocol, url, command, args, 
			environment, working_dir, version, weight, timeout_seconds, max_retries,
			status, health_check_url, is_active, metadata, tags
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)
	`

	if server.ID == uuid.Nil {
		server.ID = uuid.New()
	}

	// Convert metadata to JSON
	var metadataJSON []byte
	if server.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(server.Metadata)
		if err != nil {
			return err
		}
	}

	_, err := m.db.Exec(query,
		server.ID, server.OrganizationID, server.Name, server.Description,
		server.Protocol, server.URL, server.Command, server.Args,
		server.Environment, server.WorkingDir, server.Version, server.Weight,
		server.TimeoutSeconds, server.MaxRetries, server.Status,
		server.HealthCheckURL, server.IsActive, metadataJSON, server.Tags)
	return err
}

// GetByID retrieves an MCP server by ID
func (m *MCPServerModel) GetByID(id uuid.UUID) (*MCPServer, error) {
	query := `
		SELECT id, organization_id, name, description, protocol, url, command, args,
			   environment, working_dir, version, weight, timeout_seconds, max_retries,
			   status, health_check_url, is_active, metadata, tags, created_at, updated_at
		FROM mcp_servers
		WHERE id = $1
	`

	server := &MCPServer{}
	var metadataJSON []byte

	err := m.db.QueryRow(query, id).Scan(
		&server.ID, &server.OrganizationID, &server.Name, &server.Description,
		&server.Protocol, &server.URL, &server.Command, &server.Args,
		&server.Environment, &server.WorkingDir, &server.Version, &server.Weight,
		&server.TimeoutSeconds, &server.MaxRetries, &server.Status,
		&server.HealthCheckURL, &server.IsActive, &metadataJSON, &server.Tags,
		&server.CreatedAt, &server.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse metadata JSON
	if len(metadataJSON) > 0 {
		err = json.Unmarshal(metadataJSON, &server.Metadata)
		if err != nil {
			return nil, err
		}
	}

	return server, nil
}

// GetByName retrieves an MCP server by name within an organization
func (m *MCPServerModel) GetByName(orgID uuid.UUID, name string) (*MCPServer, error) {
	query := `
		SELECT id, organization_id, name, description, protocol, url, command, args,
			   environment, working_dir, version, weight, timeout_seconds, max_retries,
			   status, health_check_url, is_active, metadata, tags, created_at, updated_at
		FROM mcp_servers
		WHERE organization_id = $1 AND name = $2 AND is_active = true
	`

	server := &MCPServer{}
	var metadataJSON []byte

	err := m.db.QueryRow(query, orgID, name).Scan(
		&server.ID, &server.OrganizationID, &server.Name, &server.Description,
		&server.Protocol, &server.URL, &server.Command, &server.Args,
		&server.Environment, &server.WorkingDir, &server.Version, &server.Weight,
		&server.TimeoutSeconds, &server.MaxRetries, &server.Status,
		&server.HealthCheckURL, &server.IsActive, &metadataJSON, &server.Tags,
		&server.CreatedAt, &server.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse metadata JSON
	if len(metadataJSON) > 0 {
		err = json.Unmarshal(metadataJSON, &server.Metadata)
		if err != nil {
			return nil, err
		}
	}

	return server, nil
}

// ListByOrganization lists MCP servers for an organization
func (m *MCPServerModel) ListByOrganization(orgID uuid.UUID, activeOnly bool) ([]*MCPServer, error) {
	query := `
		SELECT id, organization_id, name, description, protocol, url, command, args,
			   environment, working_dir, version, weight, timeout_seconds, max_retries,
			   status, health_check_url, is_active, metadata, tags, created_at, updated_at
		FROM mcp_servers
		WHERE organization_id = $1
	`

	args := []interface{}{orgID}
	if activeOnly {
		query += " AND is_active = true"
	}
	query += " ORDER BY weight DESC, created_at DESC"

	rows, err := m.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []*MCPServer
	for rows.Next() {
		server := &MCPServer{}
		var metadataJSON []byte

		err := rows.Scan(
			&server.ID, &server.OrganizationID, &server.Name, &server.Description,
			&server.Protocol, &server.URL, &server.Command, &server.Args,
			&server.Environment, &server.WorkingDir, &server.Version, &server.Weight,
			&server.TimeoutSeconds, &server.MaxRetries, &server.Status,
			&server.HealthCheckURL, &server.IsActive, &metadataJSON, &server.Tags,
			&server.CreatedAt, &server.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse metadata JSON
		if len(metadataJSON) > 0 {
			err = json.Unmarshal(metadataJSON, &server.Metadata)
			if err != nil {
				return nil, err
			}
		}

		servers = append(servers, server)
	}

	return servers, nil
}

// GetActiveServers retrieves all active servers for an organization
func (m *MCPServerModel) GetActiveServers(orgID uuid.UUID) ([]*MCPServer, error) {
	query := `
		SELECT id, organization_id, name, description, protocol, url, command, args,
			   environment, working_dir, version, weight, timeout_seconds, max_retries,
			   status, health_check_url, is_active, metadata, tags, created_at, updated_at
		FROM mcp_servers
		WHERE organization_id = $1 AND is_active = true AND status = 'active'
		ORDER BY weight DESC, created_at DESC
	`

	rows, err := m.db.Query(query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []*MCPServer
	for rows.Next() {
		server := &MCPServer{}
		var metadataJSON []byte

		err := rows.Scan(
			&server.ID, &server.OrganizationID, &server.Name, &server.Description,
			&server.Protocol, &server.URL, &server.Command, &server.Args,
			&server.Environment, &server.WorkingDir, &server.Version, &server.Weight,
			&server.TimeoutSeconds, &server.MaxRetries, &server.Status,
			&server.HealthCheckURL, &server.IsActive, &metadataJSON, &server.Tags,
			&server.CreatedAt, &server.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse metadata JSON
		if len(metadataJSON) > 0 {
			err = json.Unmarshal(metadataJSON, &server.Metadata)
			if err != nil {
				return nil, err
			}
		}

		servers = append(servers, server)
	}

	return servers, nil
}

// Update updates an MCP server
func (m *MCPServerModel) Update(server *MCPServer) error {
	query := `
		UPDATE mcp_servers
		SET name = $2, description = $3, protocol = $4, url = $5, command = $6,
			args = $7, environment = $8, working_dir = $9, version = $10,
			weight = $11, timeout_seconds = $12, max_retries = $13,
			health_check_url = $14, metadata = $15, tags = $16
		WHERE id = $1
	`

	// Convert metadata to JSON
	var metadataJSON []byte
	if server.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(server.Metadata)
		if err != nil {
			return err
		}
	}

	_, err := m.db.Exec(query,
		server.ID, server.Name, server.Description, server.Protocol,
		server.URL, server.Command, server.Args, server.Environment,
		server.WorkingDir, server.Version, server.Weight,
		server.TimeoutSeconds, server.MaxRetries, server.HealthCheckURL,
		metadataJSON, server.Tags)
	return err
}

// UpdateStatus updates the status of an MCP server
func (m *MCPServerModel) UpdateStatus(id uuid.UUID, status string) error {
	query := `UPDATE mcp_servers SET status = $2 WHERE id = $1`
	_, err := m.db.Exec(query, id, status)
	return err
}

// Delete soft deletes an MCP server
func (m *MCPServerModel) Delete(id uuid.UUID) error {
	query := `UPDATE mcp_servers SET is_active = false WHERE id = $1`
	_, err := m.db.Exec(query, id)
	return err
}

// HealthCheck represents the health_checks table from the ERD
type HealthCheck struct {
	ID             uuid.UUID      `db:"id" json:"id"`
	ServerID       uuid.UUID      `db:"server_id" json:"server_id"`
	Status         string         `db:"status" json:"status"` // health_status_enum
	ResponseTimeMS sql.NullInt32  `db:"response_time_ms" json:"response_time_ms,omitempty"`
	ResponseBody   sql.NullString `db:"response_body" json:"response_body,omitempty"`
	ErrorMessage   sql.NullString `db:"error_message" json:"error_message,omitempty"`
	CheckedAt      time.Time      `db:"checked_at" json:"checked_at"`
}

// HealthCheckModel handles health check database operations
type HealthCheckModel struct {
	db Database
}

// NewHealthCheckModel creates a new health check model
func NewHealthCheckModel(db Database) *HealthCheckModel {
	return &HealthCheckModel{db: db}
}

// Create inserts a new health check record
func (m *HealthCheckModel) Create(check *HealthCheck) error {
	query := `
		INSERT INTO health_checks (id, server_id, status, response_time_ms, response_body, error_message, checked_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	if check.ID == uuid.Nil {
		check.ID = uuid.New()
	}

	_, err := m.db.Exec(query,
		check.ID, check.ServerID, check.Status, check.ResponseTimeMS,
		check.ResponseBody, check.ErrorMessage, check.CheckedAt)
	return err
}

// GetLatestByServerID retrieves the latest health check for a server
func (m *HealthCheckModel) GetLatestByServerID(serverID uuid.UUID) (*HealthCheck, error) {
	query := `
		SELECT id, server_id, status, response_time_ms, response_body, error_message, checked_at
		FROM health_checks
		WHERE server_id = $1
		ORDER BY checked_at DESC
		LIMIT 1
	`

	check := &HealthCheck{}
	err := m.db.QueryRow(query, serverID).Scan(
		&check.ID, &check.ServerID, &check.Status, &check.ResponseTimeMS,
		&check.ResponseBody, &check.ErrorMessage, &check.CheckedAt,
	)

	if err != nil {
		return nil, err
	}

	return check, nil
}

// GetHistoryByServerID retrieves health check history for a server
func (m *HealthCheckModel) GetHistoryByServerID(serverID uuid.UUID, limit int) ([]*HealthCheck, error) {
	query := `
		SELECT id, server_id, status, response_time_ms, response_body, error_message, checked_at
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

	var checks []*HealthCheck
	for rows.Next() {
		check := &HealthCheck{}
		err := rows.Scan(
			&check.ID, &check.ServerID, &check.Status, &check.ResponseTimeMS,
			&check.ResponseBody, &check.ErrorMessage, &check.CheckedAt,
		)
		if err != nil {
			return nil, err
		}
		checks = append(checks, check)
	}

	return checks, nil
}

// CleanupOldChecks removes health checks older than the specified duration
func (m *HealthCheckModel) CleanupOldChecks(olderThan time.Duration) error {
	query := `DELETE FROM health_checks WHERE checked_at < $1`
	cutoff := time.Now().Add(-olderThan)
	_, err := m.db.Exec(query, cutoff)
	return err
}
