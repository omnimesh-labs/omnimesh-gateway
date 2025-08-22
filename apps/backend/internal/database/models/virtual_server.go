package models

import (
	"encoding/json"

	"mcp-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
)

// VirtualServerModel handles virtual server database operations
type VirtualServerModel struct {
	db Database
}

// NewVirtualServerModel creates a new virtual server model
func NewVirtualServerModel(db Database) *VirtualServerModel {
	return &VirtualServerModel{db: db}
}

// Create inserts a new virtual server
func (m *VirtualServerModel) Create(vs *types.VirtualServer) error {
	query := `
		INSERT INTO virtual_servers (
			id, organization_id, name, description, adapter_type, tools, is_active, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at`

	toolsJSON, err := json.Marshal(vs.ToolsData)
	if err != nil {
		return err
	}
	vs.Tools = toolsJSON

	metadataJSON, err := json.Marshal(vs.Metadata)
	if err != nil {
		return err
	}

	return m.db.QueryRow(
		query,
		vs.ID,
		vs.OrganizationID,
		vs.Name,
		vs.Description,
		vs.AdapterType,
		toolsJSON,
		vs.IsActive,
		metadataJSON,
	).Scan(&vs.CreatedAt, &vs.UpdatedAt)
}

// GetByID retrieves a virtual server by ID
func (m *VirtualServerModel) GetByID(id uuid.UUID) (*types.VirtualServer, error) {
	query := `
		SELECT id, organization_id, name, description, adapter_type, tools, is_active, metadata, created_at, updated_at
		FROM virtual_servers
		WHERE id = $1`

	vs := &types.VirtualServer{}
	var toolsJSON, metadataJSON json.RawMessage

	err := m.db.QueryRow(query, id).Scan(
		&vs.ID,
		&vs.OrganizationID,
		&vs.Name,
		&vs.Description,
		&vs.AdapterType,
		&toolsJSON,
		&vs.IsActive,
		&metadataJSON,
		&vs.CreatedAt,
		&vs.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	vs.Tools = toolsJSON
	if err := json.Unmarshal(toolsJSON, &vs.ToolsData); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(metadataJSON, &vs.Metadata); err != nil {
		return nil, err
	}

	return vs, nil
}

// GetByName retrieves a virtual server by organization ID and name
func (m *VirtualServerModel) GetByName(orgID uuid.UUID, name string) (*types.VirtualServer, error) {
	query := `
		SELECT id, organization_id, name, description, adapter_type, tools, is_active, metadata, created_at, updated_at
		FROM virtual_servers
		WHERE organization_id = $1 AND name = $2`

	vs := &types.VirtualServer{}
	var toolsJSON, metadataJSON json.RawMessage

	err := m.db.QueryRow(query, orgID, name).Scan(
		&vs.ID,
		&vs.OrganizationID,
		&vs.Name,
		&vs.Description,
		&vs.AdapterType,
		&toolsJSON,
		&vs.IsActive,
		&metadataJSON,
		&vs.CreatedAt,
		&vs.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	vs.Tools = toolsJSON
	if err := json.Unmarshal(toolsJSON, &vs.ToolsData); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(metadataJSON, &vs.Metadata); err != nil {
		return nil, err
	}

	return vs, nil
}

// List retrieves all virtual servers for an organization
func (m *VirtualServerModel) List(orgID uuid.UUID) ([]*types.VirtualServer, error) {
	query := `
		SELECT id, organization_id, name, description, adapter_type, tools, is_active, metadata, created_at, updated_at
		FROM virtual_servers
		WHERE organization_id = $1
		ORDER BY name ASC`

	rows, err := m.db.Query(query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []*types.VirtualServer

	for rows.Next() {
		vs := &types.VirtualServer{}
		var toolsJSON, metadataJSON json.RawMessage

		err := rows.Scan(
			&vs.ID,
			&vs.OrganizationID,
			&vs.Name,
			&vs.Description,
			&vs.AdapterType,
			&toolsJSON,
			&vs.IsActive,
			&metadataJSON,
			&vs.CreatedAt,
			&vs.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(toolsJSON, &vs.Tools); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metadataJSON, &vs.Metadata); err != nil {
			return nil, err
		}

		servers = append(servers, vs)
	}

	return servers, rows.Err()
}

// Update updates an existing virtual server
func (m *VirtualServerModel) Update(vs *types.VirtualServer) error {
	query := `
		UPDATE virtual_servers
		SET name = $2, description = $3, adapter_type = $4, tools = $5, is_active = $6, metadata = $7
		WHERE id = $1
		RETURNING updated_at`

	toolsJSON, err := json.Marshal(vs.ToolsData)
	if err != nil {
		return err
	}
	vs.Tools = toolsJSON

	metadataJSON, err := json.Marshal(vs.Metadata)
	if err != nil {
		return err
	}

	return m.db.QueryRow(
		query,
		vs.ID,
		vs.Name,
		vs.Description,
		vs.AdapterType,
		toolsJSON,
		vs.IsActive,
		metadataJSON,
	).Scan(&vs.UpdatedAt)
}

// Delete removes a virtual server
func (m *VirtualServerModel) Delete(id uuid.UUID) error {
	query := `DELETE FROM virtual_servers WHERE id = $1`
	_, err := m.db.Exec(query, id)
	return err
}

// ListActive retrieves all active virtual servers for an organization
func (m *VirtualServerModel) ListActive(orgID uuid.UUID) ([]*types.VirtualServer, error) {
	query := `
		SELECT id, organization_id, name, description, adapter_type, tools, is_active, metadata, created_at, updated_at
		FROM virtual_servers
		WHERE organization_id = $1 AND is_active = true
		ORDER BY name ASC`

	rows, err := m.db.Query(query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []*types.VirtualServer

	for rows.Next() {
		vs := &types.VirtualServer{}
		var toolsJSON, metadataJSON json.RawMessage

		err := rows.Scan(
			&vs.ID,
			&vs.OrganizationID,
			&vs.Name,
			&vs.Description,
			&vs.AdapterType,
			&toolsJSON,
			&vs.IsActive,
			&metadataJSON,
			&vs.CreatedAt,
			&vs.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(toolsJSON, &vs.Tools); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metadataJSON, &vs.Metadata); err != nil {
			return nil, err
		}

		servers = append(servers, vs)
	}

	return servers, rows.Err()
}
