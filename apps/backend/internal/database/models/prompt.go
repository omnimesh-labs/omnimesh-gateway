package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// MCPPrompt represents the mcp_prompts table
type MCPPrompt struct {
	ID             uuid.UUID              `db:"id" json:"id"`
	OrganizationID uuid.UUID              `db:"organization_id" json:"organization_id"`
	Name           string                 `db:"name" json:"name"`
	Description    sql.NullString         `db:"description" json:"description,omitempty"`
	PromptTemplate string                 `db:"prompt_template" json:"prompt_template"`
	Parameters     []interface{}          `db:"parameters" json:"parameters,omitempty"`
	Category       string                 `db:"category" json:"category"` // prompt_category_enum
	UsageCount     int64                  `db:"usage_count" json:"usage_count"`
	IsActive       bool                   `db:"is_active" json:"is_active"`
	Metadata       map[string]interface{} `db:"metadata" json:"metadata,omitempty"`
	Tags           pq.StringArray         `db:"tags" json:"tags,omitempty"`
	CreatedAt      time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time              `db:"updated_at" json:"updated_at"`
	CreatedBy      uuid.NullUUID          `db:"created_by" json:"created_by,omitempty"`
}

// MCPPromptModel handles MCP prompt database operations
type MCPPromptModel struct {
	db Database
}

// NewMCPPromptModel creates a new MCP prompt model
func NewMCPPromptModel(db Database) *MCPPromptModel {
	return &MCPPromptModel{db: db}
}

// Create inserts a new MCP prompt
func (m *MCPPromptModel) Create(prompt *MCPPrompt) error {
	query := `
		INSERT INTO mcp_prompts (
			id, organization_id, name, description, prompt_template, parameters, 
			category, usage_count, is_active, metadata, tags, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`

	if prompt.ID == uuid.Nil {
		prompt.ID = uuid.New()
	}

	// Convert metadata to JSON
	var metadataJSON []byte
	if prompt.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(prompt.Metadata)
		if err != nil {
			return err
		}
	}

	// Convert parameters to JSON
	var parametersJSON []byte
	if prompt.Parameters != nil {
		var err error
		parametersJSON, err = json.Marshal(prompt.Parameters)
		if err != nil {
			return err
		}
	}

	_, err := m.db.Exec(query,
		prompt.ID, prompt.OrganizationID, prompt.Name, prompt.Description,
		prompt.PromptTemplate, parametersJSON, prompt.Category, prompt.UsageCount,
		prompt.IsActive, metadataJSON, prompt.Tags, prompt.CreatedBy)
	return err
}

// GetByID retrieves an MCP prompt by ID
func (m *MCPPromptModel) GetByID(id uuid.UUID) (*MCPPrompt, error) {
	query := `
		SELECT id, organization_id, name, description, prompt_template, parameters,
			   category, usage_count, is_active, metadata, tags, 
			   created_at, updated_at, created_by
		FROM mcp_prompts
		WHERE id = $1
	`

	prompt := &MCPPrompt{}
	var metadataJSON []byte
	var parametersJSON []byte

	err := m.db.QueryRow(query, id).Scan(
		&prompt.ID, &prompt.OrganizationID, &prompt.Name, &prompt.Description,
		&prompt.PromptTemplate, &parametersJSON, &prompt.Category, &prompt.UsageCount,
		&prompt.IsActive, &metadataJSON, &prompt.Tags,
		&prompt.CreatedAt, &prompt.UpdatedAt, &prompt.CreatedBy,
	)

	if err != nil {
		return nil, err
	}

	// Parse metadata JSON
	if len(metadataJSON) > 0 {
		err = json.Unmarshal(metadataJSON, &prompt.Metadata)
		if err != nil {
			return nil, err
		}
	}

	// Parse parameters JSON
	if len(parametersJSON) > 0 {
		err = json.Unmarshal(parametersJSON, &prompt.Parameters)
		if err != nil {
			return nil, err
		}
	}

	return prompt, nil
}

// GetByName retrieves an MCP prompt by name within an organization
func (m *MCPPromptModel) GetByName(orgID uuid.UUID, name string) (*MCPPrompt, error) {
	query := `
		SELECT id, organization_id, name, description, prompt_template, parameters,
			   category, usage_count, is_active, metadata, tags, 
			   created_at, updated_at, created_by
		FROM mcp_prompts
		WHERE organization_id = $1 AND name = $2 AND is_active = true
	`

	prompt := &MCPPrompt{}
	var metadataJSON []byte
	var parametersJSON []byte

	err := m.db.QueryRow(query, orgID, name).Scan(
		&prompt.ID, &prompt.OrganizationID, &prompt.Name, &prompt.Description,
		&prompt.PromptTemplate, &parametersJSON, &prompt.Category, &prompt.UsageCount,
		&prompt.IsActive, &metadataJSON, &prompt.Tags,
		&prompt.CreatedAt, &prompt.UpdatedAt, &prompt.CreatedBy,
	)

	if err != nil {
		return nil, err
	}

	// Parse metadata JSON
	if len(metadataJSON) > 0 {
		err = json.Unmarshal(metadataJSON, &prompt.Metadata)
		if err != nil {
			return nil, err
		}
	}

	// Parse parameters JSON
	if len(parametersJSON) > 0 {
		err = json.Unmarshal(parametersJSON, &prompt.Parameters)
		if err != nil {
			return nil, err
		}
	}

	return prompt, nil
}

// ListByOrganization lists MCP prompts for an organization
func (m *MCPPromptModel) ListByOrganization(orgID uuid.UUID, activeOnly bool) ([]*MCPPrompt, error) {
	query := `
		SELECT id, organization_id, name, description, prompt_template, parameters,
			   category, usage_count, is_active, metadata, tags, 
			   created_at, updated_at, created_by
		FROM mcp_prompts
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

	var prompts []*MCPPrompt
	for rows.Next() {
		prompt := &MCPPrompt{}
		var metadataJSON []byte
		var parametersJSON []byte

		err := rows.Scan(
			&prompt.ID, &prompt.OrganizationID, &prompt.Name, &prompt.Description,
			&prompt.PromptTemplate, &parametersJSON, &prompt.Category, &prompt.UsageCount,
			&prompt.IsActive, &metadataJSON, &prompt.Tags,
			&prompt.CreatedAt, &prompt.UpdatedAt, &prompt.CreatedBy,
		)
		if err != nil {
			return nil, err
		}

		// Parse metadata JSON
		if len(metadataJSON) > 0 {
			err = json.Unmarshal(metadataJSON, &prompt.Metadata)
			if err != nil {
				return nil, err
			}
		}

		// Parse parameters JSON
		if len(parametersJSON) > 0 {
			err = json.Unmarshal(parametersJSON, &prompt.Parameters)
			if err != nil {
				return nil, err
			}
		}

		prompts = append(prompts, prompt)
	}

	return prompts, nil
}

// ListByCategory lists MCP prompts by category for an organization
func (m *MCPPromptModel) ListByCategory(orgID uuid.UUID, category string, activeOnly bool) ([]*MCPPrompt, error) {
	query := `
		SELECT id, organization_id, name, description, prompt_template, parameters,
			   category, usage_count, is_active, metadata, tags, 
			   created_at, updated_at, created_by
		FROM mcp_prompts
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

	var prompts []*MCPPrompt
	for rows.Next() {
		prompt := &MCPPrompt{}
		var metadataJSON []byte
		var parametersJSON []byte

		err := rows.Scan(
			&prompt.ID, &prompt.OrganizationID, &prompt.Name, &prompt.Description,
			&prompt.PromptTemplate, &parametersJSON, &prompt.Category, &prompt.UsageCount,
			&prompt.IsActive, &metadataJSON, &prompt.Tags,
			&prompt.CreatedAt, &prompt.UpdatedAt, &prompt.CreatedBy,
		)
		if err != nil {
			return nil, err
		}

		// Parse metadata JSON
		if len(metadataJSON) > 0 {
			err = json.Unmarshal(metadataJSON, &prompt.Metadata)
			if err != nil {
				return nil, err
			}
		}

		// Parse parameters JSON
		if len(parametersJSON) > 0 {
			err = json.Unmarshal(parametersJSON, &prompt.Parameters)
			if err != nil {
				return nil, err
			}
		}

		prompts = append(prompts, prompt)
	}

	return prompts, nil
}

// GetPopularPrompts gets the most popular prompts for an organization
func (m *MCPPromptModel) GetPopularPrompts(orgID uuid.UUID, limit int) ([]*MCPPrompt, error) {
	query := `
		SELECT id, organization_id, name, description, prompt_template, parameters,
			   category, usage_count, is_active, metadata, tags, 
			   created_at, updated_at, created_by
		FROM mcp_prompts
		WHERE organization_id = $1 AND is_active = true
		ORDER BY usage_count DESC, created_at DESC
		LIMIT $2
	`

	rows, err := m.db.Query(query, orgID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prompts []*MCPPrompt
	for rows.Next() {
		prompt := &MCPPrompt{}
		var metadataJSON []byte
		var parametersJSON []byte

		err := rows.Scan(
			&prompt.ID, &prompt.OrganizationID, &prompt.Name, &prompt.Description,
			&prompt.PromptTemplate, &parametersJSON, &prompt.Category, &prompt.UsageCount,
			&prompt.IsActive, &metadataJSON, &prompt.Tags,
			&prompt.CreatedAt, &prompt.UpdatedAt, &prompt.CreatedBy,
		)
		if err != nil {
			return nil, err
		}

		// Parse metadata JSON
		if len(metadataJSON) > 0 {
			err = json.Unmarshal(metadataJSON, &prompt.Metadata)
			if err != nil {
				return nil, err
			}
		}

		// Parse parameters JSON
		if len(parametersJSON) > 0 {
			err = json.Unmarshal(parametersJSON, &prompt.Parameters)
			if err != nil {
				return nil, err
			}
		}

		prompts = append(prompts, prompt)
	}

	return prompts, nil
}

// Update updates an MCP prompt
func (m *MCPPromptModel) Update(prompt *MCPPrompt) error {
	query := `
		UPDATE mcp_prompts
		SET name = $2, description = $3, prompt_template = $4, parameters = $5,
			category = $6, metadata = $7, tags = $8
		WHERE id = $1
	`

	// Convert metadata to JSON
	var metadataJSON []byte
	if prompt.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(prompt.Metadata)
		if err != nil {
			return err
		}
	}

	// Convert parameters to JSON
	var parametersJSON []byte
	if prompt.Parameters != nil {
		var err error
		parametersJSON, err = json.Marshal(prompt.Parameters)
		if err != nil {
			return err
		}
	}

	_, err := m.db.Exec(query,
		prompt.ID, prompt.Name, prompt.Description, prompt.PromptTemplate,
		parametersJSON, prompt.Category, metadataJSON, prompt.Tags)
	return err
}

// IncrementUsageCount increments the usage count for a prompt
func (m *MCPPromptModel) IncrementUsageCount(id uuid.UUID) error {
	query := `UPDATE mcp_prompts SET usage_count = usage_count + 1 WHERE id = $1`
	_, err := m.db.Exec(query, id)
	return err
}

// Delete soft deletes an MCP prompt
func (m *MCPPromptModel) Delete(id uuid.UUID) error {
	query := `UPDATE mcp_prompts SET is_active = false WHERE id = $1`
	_, err := m.db.Exec(query, id)
	return err
}

// SearchPrompts searches prompts by name, description, or tags
func (m *MCPPromptModel) SearchPrompts(orgID uuid.UUID, searchTerm string, limit int, offset int) ([]*MCPPrompt, error) {
	query := `
		SELECT id, organization_id, name, description, prompt_template, parameters,
			   category, usage_count, is_active, metadata, tags, 
			   created_at, updated_at, created_by
		FROM mcp_prompts
		WHERE organization_id = $1 AND is_active = true
		AND (
			name ILIKE $2 OR 
			description ILIKE $2 OR 
			prompt_template ILIKE $2 OR
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

	var prompts []*MCPPrompt
	for rows.Next() {
		prompt := &MCPPrompt{}
		var metadataJSON []byte
		var parametersJSON []byte

		err := rows.Scan(
			&prompt.ID, &prompt.OrganizationID, &prompt.Name, &prompt.Description,
			&prompt.PromptTemplate, &parametersJSON, &prompt.Category, &prompt.UsageCount,
			&prompt.IsActive, &metadataJSON, &prompt.Tags,
			&prompt.CreatedAt, &prompt.UpdatedAt, &prompt.CreatedBy,
		)
		if err != nil {
			return nil, err
		}

		// Parse metadata JSON
		if len(metadataJSON) > 0 {
			err = json.Unmarshal(metadataJSON, &prompt.Metadata)
			if err != nil {
				return nil, err
			}
		}

		// Parse parameters JSON
		if len(parametersJSON) > 0 {
			err = json.Unmarshal(parametersJSON, &prompt.Parameters)
			if err != nil {
				return nil, err
			}
		}

		prompts = append(prompts, prompt)
	}

	return prompts, nil
}