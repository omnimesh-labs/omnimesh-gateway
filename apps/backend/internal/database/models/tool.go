package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// MCPTool represents the mcp_tools table
type MCPTool struct {
	UpdatedAt           time.Time              `db:"updated_at" json:"updated_at"`
	CreatedAt           time.Time              `db:"created_at" json:"created_at"`
	AccessPermissions   map[string]interface{} `db:"access_permissions" json:"access_permissions,omitempty"`
	CreatedByUUID       *uuid.UUID             `db:"-" json:"created_by,omitempty"`
	DocumentationString *string                `db:"-" json:"documentation,omitempty"`
	EndpointURLString   *string                `db:"-" json:"endpoint_url,omitempty"`
	DescriptionString   *string                `db:"-" json:"description,omitempty"`
	Schema              map[string]interface{} `db:"schema" json:"schema,omitempty"`
	Metadata            map[string]interface{} `db:"metadata" json:"metadata,omitempty"`
	Category            string                 `db:"category" json:"category"`
	ImplementationType  string                 `db:"implementation_type" json:"implementation_type"`
	Name                string                 `db:"name" json:"name"`
	FunctionName        string                 `db:"function_name" json:"function_name"`
	Documentation       sql.NullString         `db:"documentation" json:"-"`
	EndpointURL         sql.NullString         `db:"endpoint_url" json:"-"`
	Tags                pq.StringArray         `db:"tags" json:"tags,omitempty"`
	Examples            []interface{}          `db:"examples" json:"examples,omitempty"`
	Description         sql.NullString         `db:"description" json:"-"`
	MaxRetries          int                    `db:"max_retries" json:"max_retries"`
	TimeoutSeconds      int                    `db:"timeout_seconds" json:"timeout_seconds"`
	UsageCount          int64                  `db:"usage_count" json:"usage_count"`
	CreatedBy           uuid.NullUUID          `db:"created_by" json:"-"`
	ID                  uuid.UUID              `db:"id" json:"id"`
	OrganizationID      uuid.UUID              `db:"organization_id" json:"organization_id"`
	IsPublic            bool                   `db:"is_public" json:"is_public"`
	IsActive            bool                   `db:"is_active" json:"is_active"`
}

// MCPToolModel handles MCP tool database operations
type MCPToolModel struct {
	db Database
}

// NewMCPToolModel creates a new MCP tool model
func NewMCPToolModel(db Database) *MCPToolModel {
	return &MCPToolModel{db: db}
}

// Create inserts a new MCP tool
func (m *MCPToolModel) Create(tool *MCPTool) error {
	query := `
		INSERT INTO mcp_tools (
			id, organization_id, name, description, function_name, schema, category,
			implementation_type, endpoint_url, timeout_seconds, max_retries, usage_count,
			access_permissions, is_active, is_public, metadata, tags, examples,
			documentation, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		)
	`

	if tool.ID == uuid.Nil {
		tool.ID = uuid.New()
	}

	// Convert JSON fields to bytes
	var schemaJSON []byte
	if tool.Schema != nil {
		var err error
		schemaJSON, err = json.Marshal(tool.Schema)
		if err != nil {
			return err
		}
	}

	var metadataJSON []byte
	if tool.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(tool.Metadata)
		if err != nil {
			return err
		}
	}

	var accessPermissionsJSON []byte
	if tool.AccessPermissions != nil {
		var err error
		accessPermissionsJSON, err = json.Marshal(tool.AccessPermissions)
		if err != nil {
			return err
		}
	}

	var examplesJSON []byte
	if tool.Examples != nil {
		var err error
		examplesJSON, err = json.Marshal(tool.Examples)
		if err != nil {
			return err
		}
	}

	_, err := m.db.Exec(query,
		tool.ID, tool.OrganizationID, tool.Name, tool.Description, tool.FunctionName,
		schemaJSON, tool.Category, tool.ImplementationType, tool.EndpointURL,
		tool.TimeoutSeconds, tool.MaxRetries, tool.UsageCount, accessPermissionsJSON,
		tool.IsActive, tool.IsPublic, metadataJSON, tool.Tags, examplesJSON,
		tool.Documentation, tool.CreatedBy)
	return err
}

// GetByID retrieves an MCP tool by ID
func (m *MCPToolModel) GetByID(id uuid.UUID) (*MCPTool, error) {
	query := `
		SELECT id, organization_id, name, description, function_name, schema, category,
			   implementation_type, endpoint_url, timeout_seconds, max_retries, usage_count,
			   access_permissions, is_active, is_public, metadata, tags, examples,
			   documentation, created_at, updated_at, created_by
		FROM mcp_tools
		WHERE id = $1
	`

	tool := &MCPTool{}
	var schemaJSON, metadataJSON, accessPermissionsJSON, examplesJSON []byte

	err := m.db.QueryRow(query, id).Scan(
		&tool.ID, &tool.OrganizationID, &tool.Name, &tool.Description, &tool.FunctionName,
		&schemaJSON, &tool.Category, &tool.ImplementationType, &tool.EndpointURL,
		&tool.TimeoutSeconds, &tool.MaxRetries, &tool.UsageCount, &accessPermissionsJSON,
		&tool.IsActive, &tool.IsPublic, &metadataJSON, &tool.Tags, &examplesJSON,
		&tool.Documentation, &tool.CreatedAt, &tool.UpdatedAt, &tool.CreatedBy,
	)

	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	if len(schemaJSON) > 0 {
		err = json.Unmarshal(schemaJSON, &tool.Schema)
		if err != nil {
			return nil, err
		}
	}

	if len(metadataJSON) > 0 {
		err = json.Unmarshal(metadataJSON, &tool.Metadata)
		if err != nil {
			return nil, err
		}
	}

	if len(accessPermissionsJSON) > 0 {
		err = json.Unmarshal(accessPermissionsJSON, &tool.AccessPermissions)
		if err != nil {
			return nil, err
		}
	}

	if len(examplesJSON) > 0 {
		err = json.Unmarshal(examplesJSON, &tool.Examples)
		if err != nil {
			return nil, err
		}
	}

	// Convert SQL null types to JSON-friendly pointers
	convertToolNullTypes(tool)

	return tool, nil
}

// convertToolNullTypes converts SQL null types to JSON-friendly pointer types
func convertToolNullTypes(tool *MCPTool) {
	if tool.Description.Valid {
		tool.DescriptionString = &tool.Description.String
	}
	if tool.EndpointURL.Valid {
		tool.EndpointURLString = &tool.EndpointURL.String
	}
	if tool.Documentation.Valid {
		tool.DocumentationString = &tool.Documentation.String
	}
	if tool.CreatedBy.Valid {
		tool.CreatedByUUID = &tool.CreatedBy.UUID
	}
}

// GetByName retrieves an MCP tool by name within an organization
func (m *MCPToolModel) GetByName(orgID uuid.UUID, name string) (*MCPTool, error) {
	query := `
		SELECT id, organization_id, name, description, function_name, schema, category,
			   implementation_type, endpoint_url, timeout_seconds, max_retries, usage_count,
			   access_permissions, is_active, is_public, metadata, tags, examples,
			   documentation, created_at, updated_at, created_by
		FROM mcp_tools
		WHERE organization_id = $1 AND name = $2 AND is_active = true
	`

	tool := &MCPTool{}
	var schemaJSON, metadataJSON, accessPermissionsJSON, examplesJSON []byte

	err := m.db.QueryRow(query, orgID, name).Scan(
		&tool.ID, &tool.OrganizationID, &tool.Name, &tool.Description, &tool.FunctionName,
		&schemaJSON, &tool.Category, &tool.ImplementationType, &tool.EndpointURL,
		&tool.TimeoutSeconds, &tool.MaxRetries, &tool.UsageCount, &accessPermissionsJSON,
		&tool.IsActive, &tool.IsPublic, &metadataJSON, &tool.Tags, &examplesJSON,
		&tool.Documentation, &tool.CreatedAt, &tool.UpdatedAt, &tool.CreatedBy,
	)

	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	if len(schemaJSON) > 0 {
		err = json.Unmarshal(schemaJSON, &tool.Schema)
		if err != nil {
			return nil, err
		}
	}

	if len(metadataJSON) > 0 {
		err = json.Unmarshal(metadataJSON, &tool.Metadata)
		if err != nil {
			return nil, err
		}
	}

	if len(accessPermissionsJSON) > 0 {
		err = json.Unmarshal(accessPermissionsJSON, &tool.AccessPermissions)
		if err != nil {
			return nil, err
		}
	}

	if len(examplesJSON) > 0 {
		err = json.Unmarshal(examplesJSON, &tool.Examples)
		if err != nil {
			return nil, err
		}
	}

	// Convert SQL null types to JSON-friendly pointers
	convertToolNullTypes(tool)

	return tool, nil
}

// GetByFunctionName retrieves an MCP tool by function name within an organization
func (m *MCPToolModel) GetByFunctionName(orgID uuid.UUID, functionName string) (*MCPTool, error) {
	query := `
		SELECT id, organization_id, name, description, function_name, schema, category,
			   implementation_type, endpoint_url, timeout_seconds, max_retries, usage_count,
			   access_permissions, is_active, is_public, metadata, tags, examples,
			   documentation, created_at, updated_at, created_by
		FROM mcp_tools
		WHERE organization_id = $1 AND function_name = $2 AND is_active = true
	`

	tool := &MCPTool{}
	var schemaJSON, metadataJSON, accessPermissionsJSON, examplesJSON []byte

	err := m.db.QueryRow(query, orgID, functionName).Scan(
		&tool.ID, &tool.OrganizationID, &tool.Name, &tool.Description, &tool.FunctionName,
		&schemaJSON, &tool.Category, &tool.ImplementationType, &tool.EndpointURL,
		&tool.TimeoutSeconds, &tool.MaxRetries, &tool.UsageCount, &accessPermissionsJSON,
		&tool.IsActive, &tool.IsPublic, &metadataJSON, &tool.Tags, &examplesJSON,
		&tool.Documentation, &tool.CreatedAt, &tool.UpdatedAt, &tool.CreatedBy,
	)

	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	if len(schemaJSON) > 0 {
		err = json.Unmarshal(schemaJSON, &tool.Schema)
		if err != nil {
			return nil, err
		}
	}

	if len(metadataJSON) > 0 {
		err = json.Unmarshal(metadataJSON, &tool.Metadata)
		if err != nil {
			return nil, err
		}
	}

	if len(accessPermissionsJSON) > 0 {
		err = json.Unmarshal(accessPermissionsJSON, &tool.AccessPermissions)
		if err != nil {
			return nil, err
		}
	}

	if len(examplesJSON) > 0 {
		err = json.Unmarshal(examplesJSON, &tool.Examples)
		if err != nil {
			return nil, err
		}
	}

	// Convert SQL null types to JSON-friendly pointers
	convertToolNullTypes(tool)

	return tool, nil
}

// ListByOrganization lists MCP tools for an organization
func (m *MCPToolModel) ListByOrganization(orgID uuid.UUID, activeOnly bool) ([]*MCPTool, error) {
	query := `
		SELECT id, organization_id, name, description, function_name, schema, category,
			   implementation_type, endpoint_url, timeout_seconds, max_retries, usage_count,
			   access_permissions, is_active, is_public, metadata, tags, examples,
			   documentation, created_at, updated_at, created_by
		FROM mcp_tools
		WHERE organization_id = $1
	`

	args := []interface{}{orgID}
	if activeOnly {
		query += " AND is_active = true"
	}
	query += " ORDER BY usage_count DESC, created_at DESC"

	rows, err := m.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tools []*MCPTool
	for rows.Next() {
		tool := &MCPTool{}
		var schemaJSON, metadataJSON, accessPermissionsJSON, examplesJSON []byte

		err := rows.Scan(
			&tool.ID, &tool.OrganizationID, &tool.Name, &tool.Description, &tool.FunctionName,
			&schemaJSON, &tool.Category, &tool.ImplementationType, &tool.EndpointURL,
			&tool.TimeoutSeconds, &tool.MaxRetries, &tool.UsageCount, &accessPermissionsJSON,
			&tool.IsActive, &tool.IsPublic, &metadataJSON, &tool.Tags, &examplesJSON,
			&tool.Documentation, &tool.CreatedAt, &tool.UpdatedAt, &tool.CreatedBy,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		if len(schemaJSON) > 0 {
			err = json.Unmarshal(schemaJSON, &tool.Schema)
			if err != nil {
				return nil, err
			}
		}

		if len(metadataJSON) > 0 {
			err = json.Unmarshal(metadataJSON, &tool.Metadata)
			if err != nil {
				return nil, err
			}
		}

		if len(accessPermissionsJSON) > 0 {
			err = json.Unmarshal(accessPermissionsJSON, &tool.AccessPermissions)
			if err != nil {
				return nil, err
			}
		}

		if len(examplesJSON) > 0 {
			err = json.Unmarshal(examplesJSON, &tool.Examples)
			if err != nil {
				return nil, err
			}
		}

		tools = append(tools, tool)
	}

	return tools, nil
}

// ListByCategory lists MCP tools by category for an organization
func (m *MCPToolModel) ListByCategory(orgID uuid.UUID, category string, activeOnly bool) ([]*MCPTool, error) {
	query := `
		SELECT id, organization_id, name, description, function_name, schema, category,
			   implementation_type, endpoint_url, timeout_seconds, max_retries, usage_count,
			   access_permissions, is_active, is_public, metadata, tags, examples,
			   documentation, created_at, updated_at, created_by
		FROM mcp_tools
		WHERE organization_id = $1 AND category = $2
	`

	args := []interface{}{orgID, category}
	if activeOnly {
		query += " AND is_active = true"
	}
	query += " ORDER BY usage_count DESC, created_at DESC"

	rows, err := m.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return m.parseToolRows(rows)
}

// GetPopularTools gets the most popular tools for an organization
func (m *MCPToolModel) GetPopularTools(orgID uuid.UUID, limit int) ([]*MCPTool, error) {
	query := `
		SELECT id, organization_id, name, description, function_name, schema, category,
			   implementation_type, endpoint_url, timeout_seconds, max_retries, usage_count,
			   access_permissions, is_active, is_public, metadata, tags, examples,
			   documentation, created_at, updated_at, created_by
		FROM mcp_tools
		WHERE organization_id = $1 AND is_active = true
		ORDER BY usage_count DESC, created_at DESC
		LIMIT $2
	`

	rows, err := m.db.Query(query, orgID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return m.parseToolRows(rows)
}

// ListPublicTools lists all public tools (available to all organizations)
func (m *MCPToolModel) ListPublicTools(limit int, offset int) ([]*MCPTool, error) {
	query := `
		SELECT id, organization_id, name, description, function_name, schema, category,
			   implementation_type, endpoint_url, timeout_seconds, max_retries, usage_count,
			   access_permissions, is_active, is_public, metadata, tags, examples,
			   documentation, created_at, updated_at, created_by
		FROM mcp_tools
		WHERE is_public = true AND is_active = true
		ORDER BY usage_count DESC, created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := m.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return m.parseToolRows(rows)
}

// Update updates an MCP tool
func (m *MCPToolModel) Update(tool *MCPTool) error {
	query := `
		UPDATE mcp_tools
		SET name = $2, description = $3, function_name = $4, schema = $5, category = $6,
			implementation_type = $7, endpoint_url = $8, timeout_seconds = $9, max_retries = $10,
			access_permissions = $11, is_active = $12, is_public = $13, metadata = $14, tags = $15,
			examples = $16, documentation = $17
		WHERE id = $1
	`

	// Convert JSON fields to bytes
	var schemaJSON []byte
	if tool.Schema != nil {
		var err error
		schemaJSON, err = json.Marshal(tool.Schema)
		if err != nil {
			return err
		}
	}

	var metadataJSON []byte
	if tool.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(tool.Metadata)
		if err != nil {
			return err
		}
	}

	var accessPermissionsJSON []byte
	if tool.AccessPermissions != nil {
		var err error
		accessPermissionsJSON, err = json.Marshal(tool.AccessPermissions)
		if err != nil {
			return err
		}
	}

	var examplesJSON []byte
	if tool.Examples != nil {
		var err error
		examplesJSON, err = json.Marshal(tool.Examples)
		if err != nil {
			return err
		}
	}

	_, err := m.db.Exec(query,
		tool.ID, tool.Name, tool.Description, tool.FunctionName, schemaJSON, tool.Category,
		tool.ImplementationType, tool.EndpointURL, tool.TimeoutSeconds, tool.MaxRetries,
		accessPermissionsJSON, tool.IsActive, tool.IsPublic, metadataJSON, tool.Tags, examplesJSON,
		tool.Documentation)
	return err
}

// IncrementUsageCount increments the usage count for a tool
func (m *MCPToolModel) IncrementUsageCount(id uuid.UUID) error {
	query := `UPDATE mcp_tools SET usage_count = usage_count + 1 WHERE id = $1`
	_, err := m.db.Exec(query, id)
	return err
}

// Delete soft deletes an MCP tool
func (m *MCPToolModel) Delete(id uuid.UUID) error {
	query := `UPDATE mcp_tools SET is_active = false WHERE id = $1`
	_, err := m.db.Exec(query, id)
	return err
}

// SearchTools searches tools by name, description, function name, or tags
func (m *MCPToolModel) SearchTools(orgID uuid.UUID, searchTerm string, limit int, offset int) ([]*MCPTool, error) {
	query := `
		SELECT id, organization_id, name, description, function_name, schema, category,
			   implementation_type, endpoint_url, timeout_seconds, max_retries, usage_count,
			   access_permissions, is_active, is_public, metadata, tags, examples,
			   documentation, created_at, updated_at, created_by
		FROM mcp_tools
		WHERE organization_id = $1 AND is_active = true
		AND (
			name ILIKE $2 OR
			description ILIKE $2 OR
			function_name ILIKE $2 OR
			documentation ILIKE $2 OR
			$3 = ANY(tags)
		)
		ORDER BY usage_count DESC, created_at DESC
		LIMIT $4 OFFSET $5
	`

	searchPattern := "%" + searchTerm + "%"
	rows, err := m.db.Query(query, orgID, searchPattern, searchTerm, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return m.parseToolRows(rows)
}

// parseToolRows is a helper function to parse rows into MCPTool structs
func (m *MCPToolModel) parseToolRows(rows *sql.Rows) ([]*MCPTool, error) {
	var tools []*MCPTool
	for rows.Next() {
		tool := &MCPTool{}
		var schemaJSON, metadataJSON, accessPermissionsJSON, examplesJSON []byte

		err := rows.Scan(
			&tool.ID, &tool.OrganizationID, &tool.Name, &tool.Description, &tool.FunctionName,
			&schemaJSON, &tool.Category, &tool.ImplementationType, &tool.EndpointURL,
			&tool.TimeoutSeconds, &tool.MaxRetries, &tool.UsageCount, &accessPermissionsJSON,
			&tool.IsActive, &tool.IsPublic, &metadataJSON, &tool.Tags, &examplesJSON,
			&tool.Documentation, &tool.CreatedAt, &tool.UpdatedAt, &tool.CreatedBy,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		if len(schemaJSON) > 0 {
			err = json.Unmarshal(schemaJSON, &tool.Schema)
			if err != nil {
				return nil, err
			}
		}

		if len(metadataJSON) > 0 {
			err = json.Unmarshal(metadataJSON, &tool.Metadata)
			if err != nil {
				return nil, err
			}
		}

		if len(accessPermissionsJSON) > 0 {
			err = json.Unmarshal(accessPermissionsJSON, &tool.AccessPermissions)
			if err != nil {
				return nil, err
			}
		}

		if len(examplesJSON) > 0 {
			err = json.Unmarshal(examplesJSON, &tool.Examples)
			if err != nil {
				return nil, err
			}
		}

		tools = append(tools, tool)
	}

	// Convert SQL null types to JSON-friendly pointers
	for _, tool := range tools {
		convertToolNullTypes(tool)
	}

	return tools, nil
}
