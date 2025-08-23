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

// ConfigImport represents the config_imports table
type ConfigImport struct {
	ID               uuid.UUID                  `db:"id" json:"id"`
	OrganizationID   uuid.UUID                  `db:"organization_id" json:"organization_id"`
	Filename         string                     `db:"filename" json:"filename"`
	EntityTypes      pq.StringArray             `db:"entity_types" json:"entity_types"`
	Status           types.ImportStatus         `db:"status" json:"status"`
	ConflictStrategy types.ConflictStrategy     `db:"conflict_strategy" json:"conflict_strategy"`
	DryRun           bool                       `db:"dry_run" json:"dry_run"`
	Summary          map[string]interface{}     `db:"summary" json:"summary"`
	ErrorCount       int                        `db:"error_count" json:"error_count"`
	WarningCount     int                        `db:"warning_count" json:"warning_count"`
	Duration         *time.Duration             `db:"duration" json:"duration,omitempty"`
	ImportedBy       uuid.UUID                  `db:"imported_by" json:"imported_by"`
	ImportedByName   sql.NullString             `db:"imported_by_name" json:"imported_by_name,omitempty"`
	ImportedByEmail  sql.NullString             `db:"imported_by_email" json:"imported_by_email,omitempty"`
	Metadata         map[string]interface{}     `db:"metadata" json:"metadata"`
	DetailsFilePath  sql.NullString             `db:"details_file_path" json:"details_file_path,omitempty"`
	CreatedAt        time.Time                  `db:"created_at" json:"created_at"`
	CompletedAt      *time.Time                 `db:"completed_at" json:"completed_at,omitempty"`
}

// ConfigExport represents the config_exports table
type ConfigExport struct {
	ID               uuid.UUID                  `db:"id" json:"id"`
	OrganizationID   uuid.UUID                  `db:"organization_id" json:"organization_id"`
	EntityTypes      pq.StringArray             `db:"entity_types" json:"entity_types"`
	Filters          map[string]interface{}     `db:"filters" json:"filters"`
	TotalEntities    int                        `db:"total_entities" json:"total_entities"`
	FileSizeBytes    *int64                     `db:"file_size_bytes" json:"file_size_bytes,omitempty"`
	Filename         sql.NullString             `db:"filename" json:"filename,omitempty"`
	FilePath         sql.NullString             `db:"file_path" json:"file_path,omitempty"`
	ExpiresAt        *time.Time                 `db:"expires_at" json:"expires_at,omitempty"`
	ExportedBy       uuid.UUID                  `db:"exported_by" json:"exported_by"`
	ExportedByName   sql.NullString             `db:"exported_by_name" json:"exported_by_name,omitempty"`
	ExportedByEmail  sql.NullString             `db:"exported_by_email" json:"exported_by_email,omitempty"`
	Metadata         map[string]interface{}     `db:"metadata" json:"metadata"`
	CreatedAt        time.Time                  `db:"created_at" json:"created_at"`
}

// ConfigImportModel handles config import database operations
type ConfigImportModel struct {
	db Database
}

// NewConfigImportModel creates a new config import model
func NewConfigImportModel(db Database) *ConfigImportModel {
	return &ConfigImportModel{db: db}
}

// Create inserts a new config import record
func (m *ConfigImportModel) Create(importRecord *ConfigImport) error {
	query := `
		INSERT INTO config_imports (
			id, organization_id, filename, entity_types, status, conflict_strategy,
			dry_run, summary, error_count, warning_count, duration, imported_by,
			imported_by_name, imported_by_email, metadata, details_file_path
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
		RETURNING created_at, completed_at
	`

	if importRecord.ID == uuid.Nil {
		importRecord.ID = uuid.New()
	}

	summaryJSON, err := json.Marshal(importRecord.Summary)
	if err != nil {
		return err
	}

	metadataJSON, err := json.Marshal(importRecord.Metadata)
	if err != nil {
		return err
	}

	return m.db.QueryRow(query,
		importRecord.ID, importRecord.OrganizationID, importRecord.Filename,
		importRecord.EntityTypes, importRecord.Status, importRecord.ConflictStrategy,
		importRecord.DryRun, summaryJSON, importRecord.ErrorCount, importRecord.WarningCount,
		importRecord.Duration, importRecord.ImportedBy, importRecord.ImportedByName,
		importRecord.ImportedByEmail, metadataJSON, importRecord.DetailsFilePath,
	).Scan(&importRecord.CreatedAt, &importRecord.CompletedAt)
}

// GetByID retrieves a config import by ID
func (m *ConfigImportModel) GetByID(id uuid.UUID) (*ConfigImport, error) {
	query := `
		SELECT id, organization_id, filename, entity_types, status, conflict_strategy,
			   dry_run, summary, error_count, warning_count, duration, imported_by,
			   imported_by_name, imported_by_email, metadata, details_file_path,
			   created_at, completed_at
		FROM config_imports
		WHERE id = $1
	`

	importRecord := &ConfigImport{}
	var summaryJSON, metadataJSON []byte

	err := m.db.QueryRow(query, id).Scan(
		&importRecord.ID, &importRecord.OrganizationID, &importRecord.Filename,
		&importRecord.EntityTypes, &importRecord.Status, &importRecord.ConflictStrategy,
		&importRecord.DryRun, &summaryJSON, &importRecord.ErrorCount, &importRecord.WarningCount,
		&importRecord.Duration, &importRecord.ImportedBy, &importRecord.ImportedByName,
		&importRecord.ImportedByEmail, &metadataJSON, &importRecord.DetailsFilePath,
		&importRecord.CreatedAt, &importRecord.CompletedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(summaryJSON, &importRecord.Summary); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(metadataJSON, &importRecord.Metadata); err != nil {
		return nil, err
	}

	return importRecord, nil
}

// Update updates a config import record
func (m *ConfigImportModel) Update(importRecord *ConfigImport) error {
	query := `
		UPDATE config_imports
		SET status = $2, summary = $3, error_count = $4, warning_count = $5,
			duration = $6, completed_at = $7, metadata = $8
		WHERE id = $1
	`

	summaryJSON, err := json.Marshal(importRecord.Summary)
	if err != nil {
		return err
	}

	metadataJSON, err := json.Marshal(importRecord.Metadata)
	if err != nil {
		return err
	}

	_, err = m.db.Exec(query,
		importRecord.ID, importRecord.Status, summaryJSON, importRecord.ErrorCount,
		importRecord.WarningCount, importRecord.Duration, importRecord.CompletedAt,
		metadataJSON,
	)

	return err
}

// ListByOrganization lists config imports for an organization with filters and pagination
func (m *ConfigImportModel) ListByOrganization(orgID uuid.UUID, query *types.ImportHistoryQuery) ([]*ConfigImport, int, error) {
	baseQuery := `
		SELECT id, organization_id, filename, entity_types, status, conflict_strategy,
			   dry_run, summary, error_count, warning_count, duration, imported_by,
			   imported_by_name, imported_by_email, metadata, details_file_path,
			   created_at, completed_at
		FROM config_imports
		WHERE organization_id = $1
	`

	countQuery := `SELECT COUNT(*) FROM config_imports WHERE organization_id = $1`
	
	var conditions []string
	var args []interface{}
	argCount := 1
	args = append(args, orgID)

	// Add filters
	if query.Status != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("status = $%d", argCount))
		args = append(args, query.Status)
	}

	if query.EntityType != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("$%d = ANY(entity_types)", argCount))
		args = append(args, query.EntityType)
	}

	if query.ImportedBy != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("imported_by_name ILIKE $%d", argCount))
		args = append(args, "%"+query.ImportedBy+"%")
	}

	if query.StartDate != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argCount))
		args = append(args, *query.StartDate)
	}

	if query.EndDate != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argCount))
		args = append(args, *query.EndDate)
	}

	if len(conditions) > 0 {
		whereClause := " AND " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			whereClause += " AND " + conditions[i]
		}
		baseQuery += whereClause
		countQuery += whereClause
	}

	// Get total count
	var total int
	if err := m.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Add pagination
	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT %d OFFSET %d", query.Limit, query.Offset)

	rows, err := m.db.Query(baseQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var imports []*ConfigImport
	for rows.Next() {
		importRecord := &ConfigImport{}
		var summaryJSON, metadataJSON []byte

		err := rows.Scan(
			&importRecord.ID, &importRecord.OrganizationID, &importRecord.Filename,
			&importRecord.EntityTypes, &importRecord.Status, &importRecord.ConflictStrategy,
			&importRecord.DryRun, &summaryJSON, &importRecord.ErrorCount, &importRecord.WarningCount,
			&importRecord.Duration, &importRecord.ImportedBy, &importRecord.ImportedByName,
			&importRecord.ImportedByEmail, &metadataJSON, &importRecord.DetailsFilePath,
			&importRecord.CreatedAt, &importRecord.CompletedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if err := json.Unmarshal(summaryJSON, &importRecord.Summary); err != nil {
			return nil, 0, err
		}

		if err := json.Unmarshal(metadataJSON, &importRecord.Metadata); err != nil {
			return nil, 0, err
		}

		imports = append(imports, importRecord)
	}

	return imports, total, nil
}

// Delete removes a config import record
func (m *ConfigImportModel) Delete(id uuid.UUID) error {
	query := `DELETE FROM config_imports WHERE id = $1`
	_, err := m.db.Exec(query, id)
	return err
}

// ConfigExportModel handles config export database operations
type ConfigExportModel struct {
	db Database
}

// NewConfigExportModel creates a new config export model
func NewConfigExportModel(db Database) *ConfigExportModel {
	return &ConfigExportModel{db: db}
}

// Create inserts a new config export record
func (m *ConfigExportModel) Create(exportRecord *ConfigExport) error {
	query := `
		INSERT INTO config_exports (
			id, organization_id, entity_types, filters, total_entities, 
			file_size_bytes, filename, file_path, expires_at, exported_by,
			exported_by_name, exported_by_email, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
		RETURNING created_at
	`

	if exportRecord.ID == uuid.Nil {
		exportRecord.ID = uuid.New()
	}

	filtersJSON, err := json.Marshal(exportRecord.Filters)
	if err != nil {
		return err
	}

	metadataJSON, err := json.Marshal(exportRecord.Metadata)
	if err != nil {
		return err
	}

	return m.db.QueryRow(query,
		exportRecord.ID, exportRecord.OrganizationID, exportRecord.EntityTypes,
		filtersJSON, exportRecord.TotalEntities, exportRecord.FileSizeBytes,
		exportRecord.Filename, exportRecord.FilePath, exportRecord.ExpiresAt,
		exportRecord.ExportedBy, exportRecord.ExportedByName, exportRecord.ExportedByEmail,
		metadataJSON,
	).Scan(&exportRecord.CreatedAt)
}

// GetByID retrieves a config export by ID
func (m *ConfigExportModel) GetByID(id uuid.UUID) (*ConfigExport, error) {
	query := `
		SELECT id, organization_id, entity_types, filters, total_entities,
			   file_size_bytes, filename, file_path, expires_at, exported_by,
			   exported_by_name, exported_by_email, metadata, created_at
		FROM config_exports
		WHERE id = $1
	`

	exportRecord := &ConfigExport{}
	var filtersJSON, metadataJSON []byte

	err := m.db.QueryRow(query, id).Scan(
		&exportRecord.ID, &exportRecord.OrganizationID, &exportRecord.EntityTypes,
		&filtersJSON, &exportRecord.TotalEntities, &exportRecord.FileSizeBytes,
		&exportRecord.Filename, &exportRecord.FilePath, &exportRecord.ExpiresAt,
		&exportRecord.ExportedBy, &exportRecord.ExportedByName, &exportRecord.ExportedByEmail,
		&metadataJSON, &exportRecord.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(filtersJSON, &exportRecord.Filters); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(metadataJSON, &exportRecord.Metadata); err != nil {
		return nil, err
	}

	return exportRecord, nil
}

// ListByOrganization lists config exports for an organization
func (m *ConfigExportModel) ListByOrganization(orgID uuid.UUID, limit, offset int) ([]*ConfigExport, error) {
	query := `
		SELECT id, organization_id, entity_types, filters, total_entities,
			   file_size_bytes, filename, file_path, expires_at, exported_by,
			   exported_by_name, exported_by_email, metadata, created_at
		FROM config_exports
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := m.db.Query(query, orgID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exports []*ConfigExport
	for rows.Next() {
		exportRecord := &ConfigExport{}
		var filtersJSON, metadataJSON []byte

		err := rows.Scan(
			&exportRecord.ID, &exportRecord.OrganizationID, &exportRecord.EntityTypes,
			&filtersJSON, &exportRecord.TotalEntities, &exportRecord.FileSizeBytes,
			&exportRecord.Filename, &exportRecord.FilePath, &exportRecord.ExpiresAt,
			&exportRecord.ExportedBy, &exportRecord.ExportedByName, &exportRecord.ExportedByEmail,
			&metadataJSON, &exportRecord.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(filtersJSON, &exportRecord.Filters); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metadataJSON, &exportRecord.Metadata); err != nil {
			return nil, err
		}

		exports = append(exports, exportRecord)
	}

	return exports, nil
}

// Delete removes a config export record
func (m *ConfigExportModel) Delete(id uuid.UUID) error {
	query := `DELETE FROM config_exports WHERE id = $1`
	_, err := m.db.Exec(query, id)
	return err
}