package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"mcp-gateway/apps/backend/internal/types"
)

// NamespaceRepository handles namespace database operations
type NamespaceRepository struct {
	db *sqlx.DB
}

// NewNamespaceRepository creates a new namespace repository
func NewNamespaceRepository(db *sqlx.DB) *NamespaceRepository {
	return &NamespaceRepository{db: db}
}

// Create creates a new namespace
func (r *NamespaceRepository) Create(ctx context.Context, ns *types.Namespace) error {
	if ns.ID == "" {
		ns.ID = uuid.New().String()
	}

	metadataJSON, err := json.Marshal(ns.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO namespaces (
			id, organization_id, name, description, 
			created_by, is_active, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING created_at, updated_at`

	err = r.db.QueryRowContext(
		ctx, query,
		ns.ID, ns.OrganizationID, ns.Name, ns.Description,
		ns.CreatedBy, ns.IsActive, metadataJSON,
	).Scan(&ns.CreatedAt, &ns.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	return nil
}

// GetByID retrieves a namespace by ID
func (r *NamespaceRepository) GetByID(ctx context.Context, id string) (*types.Namespace, error) {
	ns := &types.Namespace{}
	var metadataJSON []byte

	query := `
		SELECT 
			id, organization_id, name, description, 
			created_at, updated_at, created_by, is_active, metadata
		FROM namespaces 
		WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ns.ID, &ns.OrganizationID, &ns.Name, &ns.Description,
		&ns.CreatedAt, &ns.UpdatedAt, &ns.CreatedBy, &ns.IsActive, &metadataJSON,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("namespace not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &ns.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return ns, nil
}

// GetByName retrieves a namespace by organization ID and name
func (r *NamespaceRepository) GetByName(ctx context.Context, orgID, name string) (*types.Namespace, error) {
	ns := &types.Namespace{}
	var metadataJSON []byte

	query := `
		SELECT 
			id, organization_id, name, description, 
			created_at, updated_at, created_by, is_active, metadata
		FROM namespaces 
		WHERE organization_id = $1 AND name = $2`

	err := r.db.QueryRowContext(ctx, query, orgID, name).Scan(
		&ns.ID, &ns.OrganizationID, &ns.Name, &ns.Description,
		&ns.CreatedAt, &ns.UpdatedAt, &ns.CreatedBy, &ns.IsActive, &metadataJSON,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("namespace not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &ns.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return ns, nil
}

// List retrieves all namespaces for an organization
func (r *NamespaceRepository) List(ctx context.Context, orgID string) ([]*types.Namespace, error) {
	query := `
		SELECT 
			id, organization_id, name, description, 
			created_at, updated_at, created_by, is_active, metadata
		FROM namespaces 
		WHERE organization_id = $1
		ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}
	defer rows.Close()

	var namespaces []*types.Namespace
	for rows.Next() {
		ns := &types.Namespace{}
		var metadataJSON []byte

		err := rows.Scan(
			&ns.ID, &ns.OrganizationID, &ns.Name, &ns.Description,
			&ns.CreatedAt, &ns.UpdatedAt, &ns.CreatedBy, &ns.IsActive, &metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan namespace: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &ns.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		namespaces = append(namespaces, ns)
	}

	return namespaces, nil
}

// ListWithServerCount retrieves all namespaces for an organization with server counts
func (r *NamespaceRepository) ListWithServerCount(ctx context.Context, orgID string) ([]*types.Namespace, error) {
	query := `
		SELECT 
			n.id, n.organization_id, n.name, n.description, 
			n.created_at, n.updated_at, n.created_by, n.is_active, n.metadata,
			COUNT(DISTINCT nsm.server_id) as server_count
		FROM namespaces n
		LEFT JOIN namespace_server_mappings nsm ON n.id = nsm.namespace_id
		WHERE n.organization_id = $1
		GROUP BY n.id, n.organization_id, n.name, n.description, 
				 n.created_at, n.updated_at, n.created_by, n.is_active, n.metadata
		ORDER BY n.name`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces with server count: %w", err)
	}
	defer rows.Close()

	var namespaces []*types.Namespace
	for rows.Next() {
		ns := &types.Namespace{}
		var metadataJSON []byte

		err := rows.Scan(
			&ns.ID, &ns.OrganizationID, &ns.Name, &ns.Description,
			&ns.CreatedAt, &ns.UpdatedAt, &ns.CreatedBy, &ns.IsActive, &metadataJSON,
			&ns.ServerCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan namespace: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &ns.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		namespaces = append(namespaces, ns)
	}

	return namespaces, nil
}

// Update updates a namespace
func (r *NamespaceRepository) Update(ctx context.Context, ns *types.Namespace) error {
	metadataJSON, err := json.Marshal(ns.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE namespaces 
		SET name = $2, description = $3, is_active = $4, 
		    metadata = $5, updated_at = NOW()
		WHERE id = $1`

	result, err := r.db.ExecContext(
		ctx, query,
		ns.ID, ns.Name, ns.Description, ns.IsActive, metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to update namespace: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("namespace not found")
	}

	return nil
}

// Delete deletes a namespace
func (r *NamespaceRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM namespaces WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("namespace not found")
	}

	return nil
}

// AddServer adds a server to a namespace
func (r *NamespaceRepository) AddServer(ctx context.Context, namespaceID, serverID string, priority int) error {
	query := `
		INSERT INTO namespace_server_mappings (
			id, namespace_id, server_id, status, priority
		) VALUES (
			$1, $2, $3, $4, $5
		) ON CONFLICT (namespace_id, server_id) DO UPDATE
		SET priority = $5, status = $4`

	_, err := r.db.ExecContext(
		ctx, query,
		uuid.New().String(), namespaceID, serverID, "ACTIVE", priority,
	)

	if err != nil {
		return fmt.Errorf("failed to add server to namespace: %w", err)
	}

	return nil
}

// RemoveServer removes a server from a namespace
func (r *NamespaceRepository) RemoveServer(ctx context.Context, namespaceID, serverID string) error {
	query := `DELETE FROM namespace_server_mappings WHERE namespace_id = $1 AND server_id = $2`

	result, err := r.db.ExecContext(ctx, query, namespaceID, serverID)
	if err != nil {
		return fmt.Errorf("failed to remove server from namespace: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	// Don't treat "no rows affected" as an error - the server might not be in the namespace
	if rowsAffected == 0 {
		// Server wasn't in the namespace, but that's okay
		return nil
	}

	return nil
}

// UpdateServerStatus updates the status of a server in a namespace
func (r *NamespaceRepository) UpdateServerStatus(ctx context.Context, namespaceID, serverID, status string) error {
	query := `
		UPDATE namespace_server_mappings 
		SET status = $3 
		WHERE namespace_id = $1 AND server_id = $2`

	result, err := r.db.ExecContext(ctx, query, namespaceID, serverID, status)
	if err != nil {
		return fmt.Errorf("failed to update server status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("server not found in namespace")
	}

	return nil
}

// GetServers retrieves all servers in a namespace
func (r *NamespaceRepository) GetServers(ctx context.Context, namespaceID string) ([]types.NamespaceServer, error) {
	query := `
		SELECT 
			nsm.server_id, ms.name as server_name, nsm.status, 
			nsm.priority, nsm.created_at
		FROM namespace_server_mappings nsm
		JOIN mcp_servers ms ON nsm.server_id = ms.id
		WHERE nsm.namespace_id = $1
		ORDER BY nsm.priority, ms.name`

	rows, err := r.db.QueryContext(ctx, query, namespaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace servers: %w", err)
	}
	defer rows.Close()

	var servers []types.NamespaceServer
	for rows.Next() {
		var server types.NamespaceServer
		err := rows.Scan(
			&server.ServerID, &server.ServerName, &server.Status,
			&server.Priority, &server.JoinedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan server: %w", err)
		}
		servers = append(servers, server)
	}

	return servers, nil
}

// SetToolStatus sets the status of a tool in a namespace
func (r *NamespaceRepository) SetToolStatus(ctx context.Context, namespaceID, serverID, toolName, status string) error {
	query := `
		INSERT INTO namespace_tool_mappings (
			id, namespace_id, server_id, tool_name, status
		) VALUES (
			$1, $2, $3, $4, $5
		) ON CONFLICT (namespace_id, server_id, tool_name) DO UPDATE
		SET status = $5`

	_, err := r.db.ExecContext(
		ctx, query,
		uuid.New().String(), namespaceID, serverID, toolName, status,
	)

	if err != nil {
		return fmt.Errorf("failed to set tool status: %w", err)
	}

	return nil
}

// GetTools retrieves all tools in a namespace
func (r *NamespaceRepository) GetTools(ctx context.Context, namespaceID string) ([]types.NamespaceTool, error) {
	query := `
		SELECT 
			ntm.server_id, ms.name as server_name, ntm.tool_name, ntm.status
		FROM namespace_tool_mappings ntm
		JOIN mcp_servers ms ON ntm.server_id = ms.id
		WHERE ntm.namespace_id = $1
		ORDER BY ms.name, ntm.tool_name`

	rows, err := r.db.QueryContext(ctx, query, namespaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace tools: %w", err)
	}
	defer rows.Close()

	var tools []types.NamespaceTool
	for rows.Next() {
		var tool types.NamespaceTool
		err := rows.Scan(
			&tool.ServerID, &tool.ServerName, &tool.ToolName, &tool.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tool: %w", err)
		}
		tools = append(tools, tool)
	}

	return tools, nil
}

// GetByIDWithServers retrieves a namespace with its servers
func (r *NamespaceRepository) GetByIDWithServers(ctx context.Context, id string) (*types.Namespace, error) {
	ns, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	servers, err := r.GetServers(ctx, id)
	if err != nil {
		return nil, err
	}

	ns.Servers = servers
	return ns, nil
}