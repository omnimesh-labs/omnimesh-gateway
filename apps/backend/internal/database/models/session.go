package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// MCPSession represents the mcp_sessions table from the ERD
type MCPSession struct {
	ID              uuid.UUID              `db:"id" json:"id"`
	OrganizationID  uuid.UUID              `db:"organization_id" json:"organization_id"`
	ServerID        uuid.UUID              `db:"server_id" json:"server_id"`
	Status          string                 `db:"status" json:"status"`     // session_status_enum
	Protocol        string                 `db:"protocol" json:"protocol"` // protocol_enum
	ClientID        sql.NullString         `db:"client_id" json:"client_id,omitempty"`
	ConnectionID    sql.NullString         `db:"connection_id" json:"connection_id,omitempty"`
	ProcessPID      sql.NullInt32          `db:"process_pid" json:"process_pid,omitempty"`
	ProcessStatus   sql.NullString         `db:"process_status" json:"process_status,omitempty"` // proc_status_enum
	ProcessExitCode sql.NullInt32          `db:"process_exit_code" json:"process_exit_code,omitempty"`
	ProcessError    sql.NullString         `db:"process_error" json:"process_error,omitempty"`
	StartedAt       time.Time              `db:"started_at" json:"started_at"`
	LastActivity    time.Time              `db:"last_activity" json:"last_activity"`
	EndedAt         sql.NullTime           `db:"ended_at" json:"ended_at,omitempty"`
	Metadata        map[string]interface{} `db:"metadata" json:"metadata,omitempty"`
	UserID          string                 `db:"user_id" json:"user_id"`
	CreatedAt       time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time              `db:"updated_at" json:"updated_at"`
}

// MCPSessionModel handles MCP session database operations
type MCPSessionModel struct {
	db Database
}

// NewMCPSessionModel creates a new MCP session model
func NewMCPSessionModel(db Database) *MCPSessionModel {
	return &MCPSessionModel{db: db}
}

// Create inserts a new MCP session
func (m *MCPSessionModel) Create(session *MCPSession) error {
	query := `
		INSERT INTO mcp_sessions (
			id, organization_id, server_id, status, protocol, client_id, connection_id,
			process_pid, process_status, process_exit_code, process_error,
			started_at, last_activity, ended_at, metadata, user_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
	`

	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}

	// Convert metadata to JSON
	var metadataJSON []byte
	if session.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(session.Metadata)
		if err != nil {
			return err
		}
	}

	_, err := m.db.Exec(query,
		session.ID, session.OrganizationID, session.ServerID, session.Status,
		session.Protocol, session.ClientID, session.ConnectionID,
		session.ProcessPID, session.ProcessStatus, session.ProcessExitCode,
		session.ProcessError, session.StartedAt, session.LastActivity,
		session.EndedAt, metadataJSON, session.UserID)
	return err
}

// GetByID retrieves an MCP session by ID
func (m *MCPSessionModel) GetByID(id uuid.UUID) (*MCPSession, error) {
	query := `
		SELECT id, organization_id, server_id, status, protocol, client_id, connection_id,
			   process_pid, process_status, process_exit_code, process_error,
			   started_at, last_activity, ended_at, metadata, user_id, created_at, updated_at
		FROM mcp_sessions
		WHERE id = $1
	`

	session := &MCPSession{}
	var metadataJSON []byte

	err := m.db.QueryRow(query, id).Scan(
		&session.ID, &session.OrganizationID, &session.ServerID, &session.Status,
		&session.Protocol, &session.ClientID, &session.ConnectionID,
		&session.ProcessPID, &session.ProcessStatus, &session.ProcessExitCode,
		&session.ProcessError, &session.StartedAt, &session.LastActivity,
		&session.EndedAt, &metadataJSON, &session.UserID,
		&session.CreatedAt, &session.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse metadata JSON
	if len(metadataJSON) > 0 {
		err = json.Unmarshal(metadataJSON, &session.Metadata)
		if err != nil {
			return nil, err
		}
	}

	return session, nil
}

// GetByClientID retrieves an MCP session by client ID
func (m *MCPSessionModel) GetByClientID(clientID string) (*MCPSession, error) {
	query := `
		SELECT id, organization_id, server_id, status, protocol, client_id, connection_id,
			   process_pid, process_status, process_exit_code, process_error,
			   started_at, last_activity, ended_at, metadata, user_id, created_at, updated_at
		FROM mcp_sessions
		WHERE client_id = $1
	`

	session := &MCPSession{}
	var metadataJSON []byte

	err := m.db.QueryRow(query, clientID).Scan(
		&session.ID, &session.OrganizationID, &session.ServerID, &session.Status,
		&session.Protocol, &session.ClientID, &session.ConnectionID,
		&session.ProcessPID, &session.ProcessStatus, &session.ProcessExitCode,
		&session.ProcessError, &session.StartedAt, &session.LastActivity,
		&session.EndedAt, &metadataJSON, &session.UserID,
		&session.CreatedAt, &session.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse metadata JSON
	if len(metadataJSON) > 0 {
		err = json.Unmarshal(metadataJSON, &session.Metadata)
		if err != nil {
			return nil, err
		}
	}

	return session, nil
}

// ListByServer lists sessions for a server
func (m *MCPSessionModel) ListByServer(serverID uuid.UUID, activeOnly bool) ([]*MCPSession, error) {
	query := `
		SELECT id, organization_id, server_id, status, protocol, client_id, connection_id,
			   process_pid, process_status, process_exit_code, process_error,
			   started_at, last_activity, ended_at, metadata, user_id, created_at, updated_at
		FROM mcp_sessions
		WHERE server_id = $1
	`

	args := []interface{}{serverID}
	if activeOnly {
		query += " AND status = 'active'"
	}
	query += " ORDER BY started_at DESC"

	rows, err := m.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*MCPSession
	for rows.Next() {
		session := &MCPSession{}
		var metadataJSON []byte

		err := rows.Scan(
			&session.ID, &session.OrganizationID, &session.ServerID, &session.Status,
			&session.Protocol, &session.ClientID, &session.ConnectionID,
			&session.ProcessPID, &session.ProcessStatus, &session.ProcessExitCode,
			&session.ProcessError, &session.StartedAt, &session.LastActivity,
			&session.EndedAt, &metadataJSON, &session.UserID,
			&session.CreatedAt, &session.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse metadata JSON
		if len(metadataJSON) > 0 {
			err = json.Unmarshal(metadataJSON, &session.Metadata)
			if err != nil {
				return nil, err
			}
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// ListByOrganization lists sessions for an organization
func (m *MCPSessionModel) ListByOrganization(orgID uuid.UUID, activeOnly bool) ([]*MCPSession, error) {
	query := `
		SELECT id, organization_id, server_id, status, protocol, client_id, connection_id,
			   process_pid, process_status, process_exit_code, process_error,
			   started_at, last_activity, ended_at, metadata, user_id, created_at, updated_at
		FROM mcp_sessions
		WHERE organization_id = $1
	`

	args := []interface{}{orgID}
	if activeOnly {
		query += " AND status = 'active'"
	}
	query += " ORDER BY started_at DESC"

	rows, err := m.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*MCPSession
	for rows.Next() {
		session := &MCPSession{}
		var metadataJSON []byte

		err := rows.Scan(
			&session.ID, &session.OrganizationID, &session.ServerID, &session.Status,
			&session.Protocol, &session.ClientID, &session.ConnectionID,
			&session.ProcessPID, &session.ProcessStatus, &session.ProcessExitCode,
			&session.ProcessError, &session.StartedAt, &session.LastActivity,
			&session.EndedAt, &metadataJSON, &session.UserID,
			&session.CreatedAt, &session.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse metadata JSON
		if len(metadataJSON) > 0 {
			err = json.Unmarshal(metadataJSON, &session.Metadata)
			if err != nil {
				return nil, err
			}
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// Update updates an MCP session
func (m *MCPSessionModel) Update(session *MCPSession) error {
	query := `
		UPDATE mcp_sessions
		SET status = $2, process_pid = $3, process_status = $4, 
			process_exit_code = $5, process_error = $6, last_activity = $7,
			ended_at = $8, metadata = $9
		WHERE id = $1
	`

	// Convert metadata to JSON
	var metadataJSON []byte
	if session.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(session.Metadata)
		if err != nil {
			return err
		}
	}

	_, err := m.db.Exec(query,
		session.ID, session.Status, session.ProcessPID, session.ProcessStatus,
		session.ProcessExitCode, session.ProcessError, session.LastActivity,
		session.EndedAt, metadataJSON)
	return err
}

// UpdateActivity updates the last activity time for a session
func (m *MCPSessionModel) UpdateActivity(id uuid.UUID) error {
	query := `UPDATE mcp_sessions SET last_activity = NOW() WHERE id = $1`
	_, err := m.db.Exec(query, id)
	return err
}

// EndSession marks a session as closed
func (m *MCPSessionModel) EndSession(id uuid.UUID, exitCode *int32, processError *string) error {
	query := `
		UPDATE mcp_sessions 
		SET status = 'closed', ended_at = NOW(), 
			process_exit_code = $2, process_error = $3
		WHERE id = $1
	`
	_, err := m.db.Exec(query, id, exitCode, processError)
	return err
}

// CleanupOldSessions removes closed sessions older than the specified duration
func (m *MCPSessionModel) CleanupOldSessions(olderThan time.Duration) error {
	query := `
		DELETE FROM mcp_sessions
		WHERE status = 'closed' AND ended_at < $1
	`

	cutoff := time.Now().Add(-olderThan)
	_, err := m.db.Exec(query, cutoff)
	return err
}
