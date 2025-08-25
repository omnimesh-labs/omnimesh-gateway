package models

import (
	"database/sql"
	"encoding/json"
	"net"
	"time"

	"github.com/google/uuid"
)

// LogIndex represents the log_index table from the ERD
type LogIndex struct {
	CreatedAt       time.Time      `db:"created_at" json:"created_at"`
	StartedAt       time.Time      `db:"started_at" json:"started_at"`
	ServerID        *uuid.UUID     `db:"server_id" json:"server_id,omitempty"`
	SessionID       *uuid.UUID     `db:"session_id" json:"session_id,omitempty"`
	ConnectionID    *uuid.UUID     `db:"connection_id" json:"connection_id,omitempty"`
	ClientID        *uuid.UUID     `db:"client_id" json:"client_id,omitempty"`
	RemoteIP        *net.IP        `db:"remote_ip" json:"remote_ip,omitempty"`
	UserID          string         `db:"user_id" json:"user_id"`
	Level           string         `db:"level" json:"level"`
	StorageProvider string         `db:"storage_provider" json:"storage_provider"`
	ObjectURI       string         `db:"object_uri" json:"object_uri"`
	RPCMethod       sql.NullString `db:"rpc_method" json:"rpc_method,omitempty"`
	ByteOffset      sql.NullInt64  `db:"byte_offset" json:"byte_offset,omitempty"`
	StatusCode      sql.NullInt32  `db:"status_code" json:"status_code,omitempty"`
	DurationMS      sql.NullInt32  `db:"duration_ms" json:"duration_ms,omitempty"`
	ID              uuid.UUID      `db:"id" json:"id"`
	OrganizationID  uuid.UUID      `db:"organization_id" json:"organization_id"`
	ErrorFlag       bool           `db:"error_flag" json:"error_flag"`
}

// AuditLog represents the audit_logs table from the ERD
type AuditLog struct {
	CreatedAt      time.Time              `db:"created_at" json:"created_at"`
	ResourceID     *uuid.UUID             `db:"resource_id" json:"resource_id,omitempty"`
	ActorIP        *net.IP                `db:"actor_ip" json:"actor_ip,omitempty"`
	OldValues      map[string]interface{} `db:"old_values" json:"old_values,omitempty"`
	NewValues      map[string]interface{} `db:"new_values" json:"new_values,omitempty"`
	Metadata       map[string]interface{} `db:"metadata" json:"metadata,omitempty"`
	Action         string                 `db:"action" json:"action"`
	ResourceType   string                 `db:"resource_type" json:"resource_type"`
	ActorID        string                 `db:"actor_id" json:"actor_id"`
	ID             uuid.UUID              `db:"id" json:"id"`
	OrganizationID uuid.UUID              `db:"organization_id" json:"organization_id"`
}

// LogAggregate represents the log_aggregates table from the ERD
type LogAggregate struct {
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at" json:"updated_at"`
	WindowStart     time.Time       `db:"window_start" json:"window_start"`
	ServerID        *uuid.UUID      `db:"server_id" json:"server_id,omitempty"`
	WindowType      string          `db:"window_type" json:"window_type"`
	RPCMethod       sql.NullString  `db:"rpc_method" json:"rpc_method,omitempty"`
	AvgDurationMS   sql.NullFloat64 `db:"avg_duration_ms" json:"avg_duration_ms,omitempty"`
	SuccessRequests int64           `db:"success_requests" json:"success_requests"`
	ErrorRequests   int64           `db:"error_requests" json:"error_requests"`
	TotalBytesIn    int64           `db:"total_bytes_in" json:"total_bytes_in"`
	TotalBytesOut   int64           `db:"total_bytes_out" json:"total_bytes_out"`
	TotalRequests   int64           `db:"total_requests" json:"total_requests"`
	P50DurationMS   sql.NullInt32   `db:"p50_duration_ms" json:"p50_duration_ms,omitempty"`
	P95DurationMS   sql.NullInt32   `db:"p95_duration_ms" json:"p95_duration_ms,omitempty"`
	P99DurationMS   sql.NullInt32   `db:"p99_duration_ms" json:"p99_duration_ms,omitempty"`
	ID              uuid.UUID       `db:"id" json:"id"`
	OrganizationID  uuid.UUID       `db:"organization_id" json:"organization_id"`
}

// LogIndexModel handles log index database operations
type LogIndexModel struct {
	db Database
}

// NewLogIndexModel creates a new log index model
func NewLogIndexModel(db Database) *LogIndexModel {
	return &LogIndexModel{db: db}
}

// Create inserts a new log index entry
func (m *LogIndexModel) Create(logEntry *LogIndex) error {
	query := `
		INSERT INTO log_index (
			id, organization_id, server_id, session_id, rpc_method, level,
			started_at, duration_ms, status_code, error_flag, storage_provider,
			object_uri, byte_offset, user_id, remote_ip, client_id, connection_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	if logEntry.ID == uuid.Nil {
		logEntry.ID = uuid.New()
	}

	_, err := m.db.Exec(query,
		logEntry.ID, logEntry.OrganizationID, logEntry.ServerID, logEntry.SessionID,
		logEntry.RPCMethod, logEntry.Level, logEntry.StartedAt, logEntry.DurationMS,
		logEntry.StatusCode, logEntry.ErrorFlag, logEntry.StorageProvider,
		logEntry.ObjectURI, logEntry.ByteOffset, logEntry.UserID, logEntry.RemoteIP,
		logEntry.ClientID, logEntry.ConnectionID)
	return err
}

// GetByID retrieves a log index entry by ID
func (m *LogIndexModel) GetByID(id uuid.UUID) (*LogIndex, error) {
	query := `
		SELECT id, organization_id, server_id, session_id, rpc_method, level,
			   started_at, duration_ms, status_code, error_flag, storage_provider,
			   object_uri, byte_offset, user_id, remote_ip, client_id, connection_id, created_at
		FROM log_index
		WHERE id = $1
	`

	logEntry := &LogIndex{}
	err := m.db.QueryRow(query, id).Scan(
		&logEntry.ID, &logEntry.OrganizationID, &logEntry.ServerID, &logEntry.SessionID,
		&logEntry.RPCMethod, &logEntry.Level, &logEntry.StartedAt, &logEntry.DurationMS,
		&logEntry.StatusCode, &logEntry.ErrorFlag, &logEntry.StorageProvider,
		&logEntry.ObjectURI, &logEntry.ByteOffset, &logEntry.UserID, &logEntry.RemoteIP,
		&logEntry.ClientID, &logEntry.ConnectionID, &logEntry.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return logEntry, nil
}

// ListByOrganization lists log entries for an organization
func (m *LogIndexModel) ListByOrganization(orgID uuid.UUID, limit, offset int) ([]*LogIndex, error) {
	query := `
		SELECT id, organization_id, server_id, session_id, rpc_method, level,
			   started_at, duration_ms, status_code, error_flag, storage_provider,
			   object_uri, byte_offset, user_id, remote_ip, client_id, connection_id, created_at
		FROM log_index
		WHERE organization_id = $1
		ORDER BY started_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := m.db.Query(query, orgID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*LogIndex
	for rows.Next() {
		logEntry := &LogIndex{}
		err := rows.Scan(
			&logEntry.ID, &logEntry.OrganizationID, &logEntry.ServerID, &logEntry.SessionID,
			&logEntry.RPCMethod, &logEntry.Level, &logEntry.StartedAt, &logEntry.DurationMS,
			&logEntry.StatusCode, &logEntry.ErrorFlag, &logEntry.StorageProvider,
			&logEntry.ObjectURI, &logEntry.ByteOffset, &logEntry.UserID, &logEntry.RemoteIP,
			&logEntry.ClientID, &logEntry.ConnectionID, &logEntry.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, logEntry)
	}

	return logs, nil
}

// ListErrors lists error log entries for an organization
func (m *LogIndexModel) ListErrors(orgID uuid.UUID, limit, offset int) ([]*LogIndex, error) {
	query := `
		SELECT id, organization_id, server_id, session_id, rpc_method, level,
			   started_at, duration_ms, status_code, error_flag, storage_provider,
			   object_uri, byte_offset, user_id, remote_ip, client_id, connection_id, created_at
		FROM log_index
		WHERE organization_id = $1 AND error_flag = true
		ORDER BY started_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := m.db.Query(query, orgID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*LogIndex
	for rows.Next() {
		logEntry := &LogIndex{}
		err := rows.Scan(
			&logEntry.ID, &logEntry.OrganizationID, &logEntry.ServerID, &logEntry.SessionID,
			&logEntry.RPCMethod, &logEntry.Level, &logEntry.StartedAt, &logEntry.DurationMS,
			&logEntry.StatusCode, &logEntry.ErrorFlag, &logEntry.StorageProvider,
			&logEntry.ObjectURI, &logEntry.ByteOffset, &logEntry.UserID, &logEntry.RemoteIP,
			&logEntry.ClientID, &logEntry.ConnectionID, &logEntry.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, logEntry)
	}

	return logs, nil
}

// CleanupOldLogs removes log index entries older than the retention policy
func (m *LogIndexModel) CleanupOldLogs(orgID uuid.UUID, retentionDays int) error {
	query := `
		DELETE FROM log_index
		WHERE organization_id = $1 AND created_at < NOW() - INTERVAL '%d days'
	`

	_, err := m.db.Exec(query, orgID, retentionDays)
	return err
}

// AuditLogModel handles audit log database operations
type AuditLogModel struct {
	db Database
}

// NewAuditLogModel creates a new audit log model
func NewAuditLogModel(db Database) *AuditLogModel {
	return &AuditLogModel{db: db}
}

// Create inserts a new audit log entry
func (m *AuditLogModel) Create(audit *AuditLog) error {
	query := `
		INSERT INTO audit_logs (
			id, organization_id, action, resource_type, resource_id,
			actor_id, actor_ip, old_values, new_values, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	if audit.ID == uuid.Nil {
		audit.ID = uuid.New()
	}

	// Convert maps to JSON
	var oldValuesJSON, newValuesJSON, metadataJSON []byte
	var err error

	if audit.OldValues != nil {
		oldValuesJSON, err = json.Marshal(audit.OldValues)
		if err != nil {
			return err
		}
	}

	if audit.NewValues != nil {
		newValuesJSON, err = json.Marshal(audit.NewValues)
		if err != nil {
			return err
		}
	}

	if audit.Metadata != nil {
		metadataJSON, err = json.Marshal(audit.Metadata)
		if err != nil {
			return err
		}
	}

	_, err = m.db.Exec(query,
		audit.ID, audit.OrganizationID, audit.Action, audit.ResourceType,
		audit.ResourceID, audit.ActorID, audit.ActorIP, oldValuesJSON,
		newValuesJSON, metadataJSON)
	return err
}

// ListByOrganization lists audit logs for an organization
func (m *AuditLogModel) ListByOrganization(orgID uuid.UUID, limit, offset int) ([]*AuditLog, error) {
	query := `
		SELECT id, organization_id, action, resource_type, resource_id,
			   actor_id, actor_ip, old_values, new_values, metadata, created_at
		FROM audit_logs
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := m.db.Query(query, orgID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var audits []*AuditLog
	for rows.Next() {
		audit := &AuditLog{}
		var oldValuesJSON, newValuesJSON, metadataJSON []byte

		err := rows.Scan(
			&audit.ID, &audit.OrganizationID, &audit.Action, &audit.ResourceType,
			&audit.ResourceID, &audit.ActorID, &audit.ActorIP, &oldValuesJSON,
			&newValuesJSON, &metadataJSON, &audit.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		if len(oldValuesJSON) > 0 {
			json.Unmarshal(oldValuesJSON, &audit.OldValues)
		}
		if len(newValuesJSON) > 0 {
			json.Unmarshal(newValuesJSON, &audit.NewValues)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &audit.Metadata)
		}

		audits = append(audits, audit)
	}

	return audits, nil
}

// ListByResourceType lists audit logs for a specific resource type
func (m *AuditLogModel) ListByResourceType(orgID uuid.UUID, resourceType string, resourceID *uuid.UUID, limit, offset int) ([]*AuditLog, error) {
	query := `
		SELECT id, organization_id, action, resource_type, resource_id,
			   actor_id, actor_ip, old_values, new_values, metadata, created_at
		FROM audit_logs
		WHERE organization_id = $1 AND resource_type = $2
	`

	args := []interface{}{orgID, resourceType}
	if resourceID != nil {
		query += " AND resource_id = $3"
		args = append(args, *resourceID)
		query += " ORDER BY created_at DESC LIMIT $4 OFFSET $5"
		args = append(args, limit, offset)
	} else {
		query += " ORDER BY created_at DESC LIMIT $3 OFFSET $4"
		args = append(args, limit, offset)
	}

	rows, err := m.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var audits []*AuditLog
	for rows.Next() {
		audit := &AuditLog{}
		var oldValuesJSON, newValuesJSON, metadataJSON []byte

		err := rows.Scan(
			&audit.ID, &audit.OrganizationID, &audit.Action, &audit.ResourceType,
			&audit.ResourceID, &audit.ActorID, &audit.ActorIP, &oldValuesJSON,
			&newValuesJSON, &metadataJSON, &audit.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		if len(oldValuesJSON) > 0 {
			json.Unmarshal(oldValuesJSON, &audit.OldValues)
		}
		if len(newValuesJSON) > 0 {
			json.Unmarshal(newValuesJSON, &audit.NewValues)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &audit.Metadata)
		}

		audits = append(audits, audit)
	}

	return audits, nil
}

// LogAggregateModel handles log aggregate database operations
type LogAggregateModel struct {
	db Database
}

// NewLogAggregateModel creates a new log aggregate model
func NewLogAggregateModel(db Database) *LogAggregateModel {
	return &LogAggregateModel{db: db}
}

// Upsert creates or updates a log aggregate entry
func (m *LogAggregateModel) Upsert(agg *LogAggregate) error {
	query := `
		INSERT INTO log_aggregates (
			id, organization_id, server_id, window_type, window_start, rpc_method,
			total_requests, success_requests, error_requests, p50_duration_ms,
			p95_duration_ms, p99_duration_ms, avg_duration_ms, total_bytes_in, total_bytes_out
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (organization_id, server_id, window_type, window_start, rpc_method)
		DO UPDATE SET
			total_requests = EXCLUDED.total_requests,
			success_requests = EXCLUDED.success_requests,
			error_requests = EXCLUDED.error_requests,
			p50_duration_ms = EXCLUDED.p50_duration_ms,
			p95_duration_ms = EXCLUDED.p95_duration_ms,
			p99_duration_ms = EXCLUDED.p99_duration_ms,
			avg_duration_ms = EXCLUDED.avg_duration_ms,
			total_bytes_in = EXCLUDED.total_bytes_in,
			total_bytes_out = EXCLUDED.total_bytes_out
	`

	if agg.ID == uuid.Nil {
		agg.ID = uuid.New()
	}

	_, err := m.db.Exec(query,
		agg.ID, agg.OrganizationID, agg.ServerID, agg.WindowType,
		agg.WindowStart, agg.RPCMethod, agg.TotalRequests,
		agg.SuccessRequests, agg.ErrorRequests, agg.P50DurationMS,
		agg.P95DurationMS, agg.P99DurationMS, agg.AvgDurationMS,
		agg.TotalBytesIn, agg.TotalBytesOut)
	return err
}

// GetDashboardData retrieves aggregated data for dashboard
func (m *LogAggregateModel) GetDashboardData(orgID uuid.UUID, windowType string, startTime, endTime time.Time) ([]*LogAggregate, error) {
	query := `
		SELECT id, organization_id, server_id, window_type, window_start, rpc_method,
			   total_requests, success_requests, error_requests, p50_duration_ms,
			   p95_duration_ms, p99_duration_ms, avg_duration_ms, total_bytes_in,
			   total_bytes_out, created_at, updated_at
		FROM log_aggregates
		WHERE organization_id = $1 AND window_type = $2
			  AND window_start >= $3 AND window_start <= $4
			  AND rpc_method IS NULL
		ORDER BY window_start DESC
	`

	rows, err := m.db.Query(query, orgID, windowType, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var aggregates []*LogAggregate
	for rows.Next() {
		agg := &LogAggregate{}
		err := rows.Scan(
			&agg.ID, &agg.OrganizationID, &agg.ServerID, &agg.WindowType,
			&agg.WindowStart, &agg.RPCMethod, &agg.TotalRequests,
			&agg.SuccessRequests, &agg.ErrorRequests, &agg.P50DurationMS,
			&agg.P95DurationMS, &agg.P99DurationMS, &agg.AvgDurationMS,
			&agg.TotalBytesIn, &agg.TotalBytesOut, &agg.CreatedAt, &agg.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		aggregates = append(aggregates, agg)
	}

	return aggregates, nil
}
