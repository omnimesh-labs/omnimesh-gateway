package models

import (
	"encoding/json"
	"time"

	"mcp-gateway/apps/backend/internal/types"
)

// LogEntryModel handles log entry database operations
type LogEntryModel struct {
	BaseModel
}

// NewLogEntryModel creates a new log entry model
func NewLogEntryModel(db Database) *LogEntryModel {
	return &LogEntryModel{
		BaseModel: BaseModel{db: db},
	}
}

// Create inserts a new log entry into the database
func (m *LogEntryModel) Create(entry *types.LogEntry) error {
	query := `
		INSERT INTO request_logs (id, timestamp, level, type, user_id, organization_id,
			request_id, method, path, status_code, duration, remote_ip, user_agent,
			message, data, error)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	var dataJSON string
	if entry.Data != nil {
		data, _ := json.Marshal(entry.Data)
		dataJSON = string(data)
	}

	_, err := m.db.Exec(query, entry.ID, entry.Timestamp, entry.Level, entry.Type,
		entry.UserID, entry.OrganizationID, entry.RequestID, entry.Method, entry.Path,
		entry.StatusCode, entry.Duration.Nanoseconds(), entry.RemoteIP, entry.UserAgent,
		entry.Message, dataJSON, entry.Error)

	return err
}

// GetByID retrieves a log entry by ID
func (m *LogEntryModel) GetByID(id string) (*types.LogEntry, error) {
	query := `
		SELECT id, timestamp, level, type, user_id, organization_id, request_id,
			method, path, status_code, duration, remote_ip, user_agent, message, data, error
		FROM request_logs
		WHERE id = $1
	`

	entry := &types.LogEntry{}
	var dataJSON string
	var durationNanos int64

	err := m.db.QueryRow(query, id).Scan(
		&entry.ID, &entry.Timestamp, &entry.Level, &entry.Type, &entry.UserID,
		&entry.OrganizationID, &entry.RequestID, &entry.Method, &entry.Path,
		&entry.StatusCode, &durationNanos, &entry.RemoteIP, &entry.UserAgent,
		&entry.Message, &dataJSON, &entry.Error,
	)

	if err != nil {
		return nil, err
	}

	// Parse JSON data and duration
	if dataJSON != "" {
		json.Unmarshal([]byte(dataJSON), &entry.Data)
	}
	entry.Duration = time.Duration(durationNanos)

	return entry, nil
}

// Query searches log entries with filters
func (m *LogEntryModel) Query(query *types.LogQueryRequest) ([]*types.LogEntry, error) {
	sqlQuery := `
		SELECT id, timestamp, level, type, user_id, organization_id, request_id,
			method, path, status_code, duration, remote_ip, user_agent, message, data, error
		FROM request_logs
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	// Add filters
	if !query.StartTime.IsZero() {
		sqlQuery += " AND timestamp >= $" + string(rune(argIndex))
		args = append(args, query.StartTime)
		argIndex++
	}

	if !query.EndTime.IsZero() {
		sqlQuery += " AND timestamp <= $" + string(rune(argIndex))
		args = append(args, query.EndTime)
		argIndex++
	}

	if query.Level != "" {
		sqlQuery += " AND level = $" + string(rune(argIndex))
		args = append(args, query.Level)
		argIndex++
	}

	if query.Type != "" {
		sqlQuery += " AND type = $" + string(rune(argIndex))
		args = append(args, query.Type)
		argIndex++
	}

	if query.UserID != "" {
		sqlQuery += " AND user_id = $" + string(rune(argIndex))
		args = append(args, query.UserID)
		argIndex++
	}

	if query.OrganizationID != "" {
		sqlQuery += " AND organization_id = $" + string(rune(argIndex))
		args = append(args, query.OrganizationID)
		argIndex++
	}

	if query.Method != "" {
		sqlQuery += " AND method = $" + string(rune(argIndex))
		args = append(args, query.Method)
		argIndex++
	}

	if query.Path != "" {
		sqlQuery += " AND path LIKE $" + string(rune(argIndex))
		args = append(args, "%"+query.Path+"%")
		argIndex++
	}

	if query.StatusCode != 0 {
		sqlQuery += " AND status_code = $" + string(rune(argIndex))
		args = append(args, query.StatusCode)
		argIndex++
	}

	if query.Search != "" {
		sqlQuery += " AND (message ILIKE $" + string(rune(argIndex)) + " OR error ILIKE $" + string(rune(argIndex)) + ")"
		args = append(args, "%"+query.Search+"%")
		argIndex++
	}

	// Add ordering and pagination
	sqlQuery += " ORDER BY timestamp DESC"

	if query.Limit > 0 {
		sqlQuery += " LIMIT $" + string(rune(argIndex))
		args = append(args, query.Limit)
		argIndex++
	}

	if query.Offset > 0 {
		sqlQuery += " OFFSET $" + string(rune(argIndex))
		args = append(args, query.Offset)
		argIndex++
	}

	rows, err := m.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*types.LogEntry
	for rows.Next() {
		entry := &types.LogEntry{}
		var dataJSON string
		var durationNanos int64

		err := rows.Scan(
			&entry.ID, &entry.Timestamp, &entry.Level, &entry.Type, &entry.UserID,
			&entry.OrganizationID, &entry.RequestID, &entry.Method, &entry.Path,
			&entry.StatusCode, &durationNanos, &entry.RemoteIP, &entry.UserAgent,
			&entry.Message, &dataJSON, &entry.Error,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON data and duration
		if dataJSON != "" {
			json.Unmarshal([]byte(dataJSON), &entry.Data)
		}
		entry.Duration = time.Duration(durationNanos)

		entries = append(entries, entry)
	}

	return entries, nil
}

// DeleteOldEntries removes log entries older than the specified duration
func (m *LogEntryModel) DeleteOldEntries(olderThan time.Duration) error {
	query := `
		DELETE FROM request_logs
		WHERE timestamp < $1
	`

	cutoff := time.Now().Add(-olderThan)
	_, err := m.db.Exec(query, cutoff)
	return err
}

// AuditLogModel handles audit log database operations
type AuditLogModel struct {
	BaseModel
}

// NewAuditLogModel creates a new audit log model
func NewAuditLogModel(db Database) *AuditLogModel {
	return &AuditLogModel{
		BaseModel: BaseModel{db: db},
	}
}

// Create inserts a new audit log entry into the database
func (m *AuditLogModel) Create(audit *types.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, timestamp, user_id, organization_id, action,
			resource, resource_id, details, remote_ip, user_agent, success, error)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	var detailsJSON string
	if audit.Details != nil {
		details, _ := json.Marshal(audit.Details)
		detailsJSON = string(details)
	}

	_, err := m.db.Exec(query, audit.ID, audit.Timestamp, audit.UserID,
		audit.OrganizationID, audit.Action, audit.Resource, audit.ResourceID,
		detailsJSON, audit.RemoteIP, audit.UserAgent, audit.Success, audit.Error)

	return err
}

// GetByResourceID retrieves audit logs for a specific resource
func (m *AuditLogModel) GetByResourceID(resource, resourceID string, limit, offset int) ([]*types.AuditLog, error) {
	query := `
		SELECT id, timestamp, user_id, organization_id, action, resource,
			resource_id, details, remote_ip, user_agent, success, error
		FROM audit_logs
		WHERE resource = $1 AND resource_id = $2
		ORDER BY timestamp DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := m.db.Query(query, resource, resourceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var audits []*types.AuditLog
	for rows.Next() {
		audit := &types.AuditLog{}
		var detailsJSON string

		err := rows.Scan(
			&audit.ID, &audit.Timestamp, &audit.UserID, &audit.OrganizationID,
			&audit.Action, &audit.Resource, &audit.ResourceID, &detailsJSON,
			&audit.RemoteIP, &audit.UserAgent, &audit.Success, &audit.Error,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON details
		if detailsJSON != "" {
			json.Unmarshal([]byte(detailsJSON), &audit.Details)
		}

		audits = append(audits, audit)
	}

	return audits, nil
}

// GetByOrganization retrieves audit logs for an organization
func (m *AuditLogModel) GetByOrganization(orgID string, startTime, endTime time.Time, limit, offset int) ([]*types.AuditLog, error) {
	query := `
		SELECT id, timestamp, user_id, organization_id, action, resource,
			resource_id, details, remote_ip, user_agent, success, error
		FROM audit_logs
		WHERE organization_id = $1 AND timestamp >= $2 AND timestamp <= $3
		ORDER BY timestamp DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := m.db.Query(query, orgID, startTime, endTime, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var audits []*types.AuditLog
	for rows.Next() {
		audit := &types.AuditLog{}
		var detailsJSON string

		err := rows.Scan(
			&audit.ID, &audit.Timestamp, &audit.UserID, &audit.OrganizationID,
			&audit.Action, &audit.Resource, &audit.ResourceID, &detailsJSON,
			&audit.RemoteIP, &audit.UserAgent, &audit.Success, &audit.Error,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON details
		if detailsJSON != "" {
			json.Unmarshal([]byte(detailsJSON), &audit.Details)
		}

		audits = append(audits, audit)
	}

	return audits, nil
}

// DeleteOldAudits removes audit logs older than the specified duration
func (m *AuditLogModel) DeleteOldAudits(olderThan time.Duration) error {
	query := `
		DELETE FROM audit_logs
		WHERE timestamp < $1
	`

	cutoff := time.Now().Add(-olderThan)
	_, err := m.db.Exec(query, cutoff)
	return err
}
