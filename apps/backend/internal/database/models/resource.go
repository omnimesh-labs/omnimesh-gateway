package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// MCPResource represents the mcp_resources table
type MCPResource struct {
	CreatedAt         time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time              `db:"updated_at" json:"updated_at"`
	AccessPermissions map[string]interface{} `db:"access_permissions" json:"access_permissions,omitempty"`
	CreatedByUUID     *uuid.UUID             `db:"-" json:"created_by,omitempty"`
	SizeBytesInt64    *int64                 `db:"-" json:"size_bytes,omitempty"`
	MimeTypeString    *string                `db:"-" json:"mime_type,omitempty"`
	DescriptionString *string                `db:"-" json:"description,omitempty"`
	Metadata          map[string]interface{} `db:"metadata" json:"metadata,omitempty"`
	Name              string                 `db:"name" json:"name"`
	ResourceType      string                 `db:"resource_type" json:"resource_type"`
	URI               string                 `db:"uri" json:"uri"`
	Tags              pq.StringArray         `db:"tags" json:"tags,omitempty"`
	Description       sql.NullString         `db:"description" json:"-"`
	MimeType          sql.NullString         `db:"mime_type" json:"-"`
	SizeBytes         sql.NullInt64          `db:"size_bytes" json:"-"`
	CreatedBy         uuid.NullUUID          `db:"created_by" json:"-"`
	OrganizationID    uuid.UUID              `db:"organization_id" json:"organization_id"`
	ID                uuid.UUID              `db:"id" json:"id"`
	IsActive          bool                   `db:"is_active" json:"is_active"`
}

// MCPResourceModel handles MCP resource database operations
type MCPResourceModel struct {
	db Database
}

// NewMCPResourceModel creates a new MCP resource model
func NewMCPResourceModel(db Database) *MCPResourceModel {
	return &MCPResourceModel{db: db}
}

// Create inserts a new MCP resource
func (m *MCPResourceModel) Create(resource *MCPResource) error {
	query := `
		INSERT INTO mcp_resources (
			id, organization_id, name, description, resource_type, uri, mime_type,
			size_bytes, access_permissions, is_active, metadata, tags, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	if resource.ID == uuid.Nil {
		resource.ID = uuid.New()
	}

	// Convert metadata to JSON
	var metadataJSON []byte
	if resource.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(resource.Metadata)
		if err != nil {
			return err
		}
	}

	// Convert access_permissions to JSON
	var accessPermissionsJSON []byte
	if resource.AccessPermissions != nil {
		var err error
		accessPermissionsJSON, err = json.Marshal(resource.AccessPermissions)
		if err != nil {
			return err
		}
	}

	_, err := m.db.Exec(query,
		resource.ID, resource.OrganizationID, resource.Name, resource.Description,
		resource.ResourceType, resource.URI, resource.MimeType, resource.SizeBytes,
		accessPermissionsJSON, resource.IsActive, metadataJSON, resource.Tags,
		resource.CreatedBy)
	return err
}

// GetByID retrieves an MCP resource by ID
func (m *MCPResourceModel) GetByID(id uuid.UUID) (*MCPResource, error) {
	query := `
		SELECT id, organization_id, name, description, resource_type, uri, mime_type,
			   size_bytes, access_permissions, is_active, metadata, tags,
			   created_at, updated_at, created_by
		FROM mcp_resources
		WHERE id = $1
	`

	resource := &MCPResource{}
	var metadataJSON []byte
	var accessPermissionsJSON []byte

	err := m.db.QueryRow(query, id).Scan(
		&resource.ID, &resource.OrganizationID, &resource.Name, &resource.Description,
		&resource.ResourceType, &resource.URI, &resource.MimeType, &resource.SizeBytes,
		&accessPermissionsJSON, &resource.IsActive, &metadataJSON, &resource.Tags,
		&resource.CreatedAt, &resource.UpdatedAt, &resource.CreatedBy,
	)

	if err != nil {
		return nil, err
	}

	// Parse metadata JSON
	if len(metadataJSON) > 0 {
		err = json.Unmarshal(metadataJSON, &resource.Metadata)
		if err != nil {
			return nil, err
		}
	}

	// Parse access_permissions JSON
	if len(accessPermissionsJSON) > 0 {
		err = json.Unmarshal(accessPermissionsJSON, &resource.AccessPermissions)
		if err != nil {
			return nil, err
		}
	}

	// Convert SQL null types to JSON-friendly pointers
	convertResourceNullTypes(resource)

	return resource, nil
}

// GetByName retrieves an MCP resource by name within an organization
func (m *MCPResourceModel) GetByName(orgID uuid.UUID, name string) (*MCPResource, error) {
	query := `
		SELECT id, organization_id, name, description, resource_type, uri, mime_type,
			   size_bytes, access_permissions, is_active, metadata, tags,
			   created_at, updated_at, created_by
		FROM mcp_resources
		WHERE organization_id = $1 AND name = $2 AND is_active = true
	`

	resource := &MCPResource{}
	var metadataJSON []byte
	var accessPermissionsJSON []byte

	err := m.db.QueryRow(query, orgID, name).Scan(
		&resource.ID, &resource.OrganizationID, &resource.Name, &resource.Description,
		&resource.ResourceType, &resource.URI, &resource.MimeType, &resource.SizeBytes,
		&accessPermissionsJSON, &resource.IsActive, &metadataJSON, &resource.Tags,
		&resource.CreatedAt, &resource.UpdatedAt, &resource.CreatedBy,
	)

	if err != nil {
		return nil, err
	}

	// Parse metadata JSON
	if len(metadataJSON) > 0 {
		err = json.Unmarshal(metadataJSON, &resource.Metadata)
		if err != nil {
			return nil, err
		}
	}

	// Parse access_permissions JSON
	if len(accessPermissionsJSON) > 0 {
		err = json.Unmarshal(accessPermissionsJSON, &resource.AccessPermissions)
		if err != nil {
			return nil, err
		}
	}

	// Convert SQL null types to JSON-friendly pointers
	convertResourceNullTypes(resource)

	return resource, nil
}

// ListByOrganization lists MCP resources for an organization
func (m *MCPResourceModel) ListByOrganization(orgID uuid.UUID, activeOnly bool) ([]*MCPResource, error) {
	query := `
		SELECT id, organization_id, name, description, resource_type, uri, mime_type,
			   size_bytes, access_permissions, is_active, metadata, tags,
			   created_at, updated_at, created_by
		FROM mcp_resources
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

	var resources []*MCPResource
	for rows.Next() {
		resource := &MCPResource{}
		var metadataJSON []byte
		var accessPermissionsJSON []byte

		err := rows.Scan(
			&resource.ID, &resource.OrganizationID, &resource.Name, &resource.Description,
			&resource.ResourceType, &resource.URI, &resource.MimeType, &resource.SizeBytes,
			&accessPermissionsJSON, &resource.IsActive, &metadataJSON, &resource.Tags,
			&resource.CreatedAt, &resource.UpdatedAt, &resource.CreatedBy,
		)
		if err != nil {
			return nil, err
		}

		// Parse metadata JSON
		if len(metadataJSON) > 0 {
			err = json.Unmarshal(metadataJSON, &resource.Metadata)
			if err != nil {
				return nil, err
			}
		}

		// Parse access_permissions JSON
		if len(accessPermissionsJSON) > 0 {
			err = json.Unmarshal(accessPermissionsJSON, &resource.AccessPermissions)
			if err != nil {
				return nil, err
			}
		}

		resources = append(resources, resource)
	}

	// Convert SQL null types to JSON-friendly pointers
	for _, resource := range resources {
		convertResourceNullTypes(resource)
	}

	return resources, nil
}

// ListByType lists MCP resources by type for an organization
func (m *MCPResourceModel) ListByType(orgID uuid.UUID, resourceType string, activeOnly bool) ([]*MCPResource, error) {
	query := `
		SELECT id, organization_id, name, description, resource_type, uri, mime_type,
			   size_bytes, access_permissions, is_active, metadata, tags,
			   created_at, updated_at, created_by
		FROM mcp_resources
		WHERE organization_id = $1 AND resource_type = $2
	`

	args := []interface{}{orgID, resourceType}
	if activeOnly {
		query += " AND is_active = true"
	}
	query += " ORDER BY created_at DESC"

	rows, err := m.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []*MCPResource
	for rows.Next() {
		resource := &MCPResource{}
		var metadataJSON []byte
		var accessPermissionsJSON []byte

		err := rows.Scan(
			&resource.ID, &resource.OrganizationID, &resource.Name, &resource.Description,
			&resource.ResourceType, &resource.URI, &resource.MimeType, &resource.SizeBytes,
			&accessPermissionsJSON, &resource.IsActive, &metadataJSON, &resource.Tags,
			&resource.CreatedAt, &resource.UpdatedAt, &resource.CreatedBy,
		)
		if err != nil {
			return nil, err
		}

		// Parse metadata JSON
		if len(metadataJSON) > 0 {
			err = json.Unmarshal(metadataJSON, &resource.Metadata)
			if err != nil {
				return nil, err
			}
		}

		// Parse access_permissions JSON
		if len(accessPermissionsJSON) > 0 {
			err = json.Unmarshal(accessPermissionsJSON, &resource.AccessPermissions)
			if err != nil {
				return nil, err
			}
		}

		resources = append(resources, resource)
	}

	// Convert SQL null types to JSON-friendly pointers
	for _, resource := range resources {
		convertResourceNullTypes(resource)
	}

	return resources, nil
}

// Update updates an MCP resource
func (m *MCPResourceModel) Update(resource *MCPResource) error {
	query := `
		UPDATE mcp_resources
		SET name = $2, description = $3, resource_type = $4, uri = $5, mime_type = $6,
			size_bytes = $7, access_permissions = $8, is_active = $9, metadata = $10, tags = $11, updated_at = NOW()
		WHERE id = $1
	`

	// Convert metadata to JSON
	var metadataJSON []byte
	if resource.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(resource.Metadata)
		if err != nil {
			return err
		}
	}

	// Convert access_permissions to JSON
	var accessPermissionsJSON []byte
	if resource.AccessPermissions != nil {
		var err error
		accessPermissionsJSON, err = json.Marshal(resource.AccessPermissions)
		if err != nil {
			return err
		}
	}

	_, err := m.db.Exec(query,
		resource.ID, resource.Name, resource.Description, resource.ResourceType,
		resource.URI, resource.MimeType, resource.SizeBytes, accessPermissionsJSON,
		resource.IsActive, metadataJSON, resource.Tags)
	return err
}

// Delete soft deletes an MCP resource
func (m *MCPResourceModel) Delete(id uuid.UUID) error {
	query := `UPDATE mcp_resources SET is_active = false WHERE id = $1`
	_, err := m.db.Exec(query, id)
	return err
}

// SearchResources searches resources by name, description, or tags
func (m *MCPResourceModel) SearchResources(orgID uuid.UUID, searchTerm string, limit int, offset int) ([]*MCPResource, error) {
	query := `
		SELECT id, organization_id, name, description, resource_type, uri, mime_type,
			   size_bytes, access_permissions, is_active, metadata, tags,
			   created_at, updated_at, created_by
		FROM mcp_resources
		WHERE organization_id = $1 AND is_active = true
		AND (
			name ILIKE $2 OR
			description ILIKE $2 OR
			$3 = ANY(tags)
		)
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`

	searchPattern := "%" + searchTerm + "%"
	rows, err := m.db.Query(query, orgID, searchPattern, searchTerm, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []*MCPResource
	for rows.Next() {
		resource := &MCPResource{}
		var metadataJSON []byte
		var accessPermissionsJSON []byte

		err := rows.Scan(
			&resource.ID, &resource.OrganizationID, &resource.Name, &resource.Description,
			&resource.ResourceType, &resource.URI, &resource.MimeType, &resource.SizeBytes,
			&accessPermissionsJSON, &resource.IsActive, &metadataJSON, &resource.Tags,
			&resource.CreatedAt, &resource.UpdatedAt, &resource.CreatedBy,
		)
		if err != nil {
			return nil, err
		}

		// Parse metadata JSON
		if len(metadataJSON) > 0 {
			err = json.Unmarshal(metadataJSON, &resource.Metadata)
			if err != nil {
				return nil, err
			}
		}

		// Parse access_permissions JSON
		if len(accessPermissionsJSON) > 0 {
			err = json.Unmarshal(accessPermissionsJSON, &resource.AccessPermissions)
			if err != nil {
				return nil, err
			}
		}

		resources = append(resources, resource)
	}

	// Convert SQL null types to JSON-friendly pointers
	for _, resource := range resources {
		convertResourceNullTypes(resource)
	}

	return resources, nil
}

// convertResourceNullTypes converts SQL null types to JSON-friendly pointer types
func convertResourceNullTypes(resource *MCPResource) {
	if resource.Description.Valid {
		resource.DescriptionString = &resource.Description.String
	}
	if resource.MimeType.Valid {
		resource.MimeTypeString = &resource.MimeType.String
	}
	if resource.SizeBytes.Valid {
		resource.SizeBytesInt64 = &resource.SizeBytes.Int64
	}
	if resource.CreatedBy.Valid {
		resource.CreatedByUUID = &resource.CreatedBy.UUID
	}
}
