package repositories

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"mcp-gateway/apps/backend/internal/types"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// NullableJSONB handles nullable JSONB columns
type NullableJSONB struct {
	Valid bool
	Data  map[string]interface{}
}

// Scan implements the sql.Scanner interface
func (n *NullableJSONB) Scan(value interface{}) error {
	if value == nil {
		n.Valid = false
		n.Data = make(map[string]interface{})
		return nil
	}

	n.Valid = true
	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			n.Data = make(map[string]interface{})
			return nil
		}
		return json.Unmarshal(v, &n.Data)
	case string:
		if v == "" {
			n.Data = make(map[string]interface{})
			return nil
		}
		return json.Unmarshal([]byte(v), &n.Data)
	default:
		return fmt.Errorf("cannot scan type %T into NullableJSONB", value)
	}
}

// Value implements the driver.Valuer interface
func (n NullableJSONB) Value() (driver.Value, error) {
	if !n.Valid || n.Data == nil || len(n.Data) == 0 {
		return nil, nil
	}
	return json.Marshal(n.Data)
}

// EndpointRepository handles endpoint database operations
type EndpointRepository struct {
	db *sqlx.DB
}

// NewEndpointRepository creates a new endpoint repository
func NewEndpointRepository(db *sqlx.DB) *EndpointRepository {
	return &EndpointRepository{db: db}
}

// Create creates a new endpoint
func (r *EndpointRepository) Create(ctx context.Context, endpoint *types.Endpoint) error {
	query := `
		INSERT INTO endpoints (
			organization_id, namespace_id, name, description,
			enable_api_key_auth, enable_oauth, enable_public_access, use_query_param_auth,
			rate_limit_requests, rate_limit_window,
			allowed_origins, allowed_methods,
			created_by, is_active, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		) RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		endpoint.OrganizationID, endpoint.NamespaceID, endpoint.Name, endpoint.Description,
		endpoint.EnableAPIKeyAuth, endpoint.EnableOAuth, endpoint.EnablePublicAccess, endpoint.UseQueryParamAuth,
		endpoint.RateLimitRequests, endpoint.RateLimitWindow,
		pq.Array(endpoint.AllowedOrigins), pq.Array(endpoint.AllowedMethods),
		endpoint.CreatedBy, endpoint.IsActive, endpoint.Metadata,
	).Scan(&endpoint.ID, &endpoint.CreatedAt, &endpoint.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create endpoint: %w", err)
	}

	return nil
}

// GetByID retrieves an endpoint by ID
func (r *EndpointRepository) GetByID(ctx context.Context, id string) (*types.Endpoint, error) {
	endpoint := &types.Endpoint{}

	query := `
		SELECT
			id, organization_id, namespace_id, name, description,
			enable_api_key_auth, enable_oauth, enable_public_access, use_query_param_auth,
			rate_limit_requests, rate_limit_window,
			allowed_origins, allowed_methods,
			created_at, updated_at, created_by, is_active, metadata
		FROM endpoints
		WHERE id = $1`

	var metadata NullableJSONB
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&endpoint.ID, &endpoint.OrganizationID, &endpoint.NamespaceID, &endpoint.Name, &endpoint.Description,
		&endpoint.EnableAPIKeyAuth, &endpoint.EnableOAuth, &endpoint.EnablePublicAccess, &endpoint.UseQueryParamAuth,
		&endpoint.RateLimitRequests, &endpoint.RateLimitWindow,
		pq.Array(&endpoint.AllowedOrigins), pq.Array(&endpoint.AllowedMethods),
		&endpoint.CreatedAt, &endpoint.UpdatedAt, &endpoint.CreatedBy, &endpoint.IsActive, &metadata,
	)
	endpoint.Metadata = metadata.Data

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("endpoint not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoint: %w", err)
	}

	return endpoint, nil
}

// GetByName retrieves an endpoint by name
func (r *EndpointRepository) GetByName(ctx context.Context, name string) (*types.Endpoint, error) {
	endpoint := &types.Endpoint{}

	query := `
		SELECT
			id, organization_id, namespace_id, name, description,
			enable_api_key_auth, enable_oauth, enable_public_access, use_query_param_auth,
			rate_limit_requests, rate_limit_window,
			allowed_origins, allowed_methods,
			created_at, updated_at, created_by, is_active, metadata
		FROM endpoints
		WHERE name = $1 AND is_active = true`

	var metadata NullableJSONB
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&endpoint.ID, &endpoint.OrganizationID, &endpoint.NamespaceID, &endpoint.Name, &endpoint.Description,
		&endpoint.EnableAPIKeyAuth, &endpoint.EnableOAuth, &endpoint.EnablePublicAccess, &endpoint.UseQueryParamAuth,
		&endpoint.RateLimitRequests, &endpoint.RateLimitWindow,
		pq.Array(&endpoint.AllowedOrigins), pq.Array(&endpoint.AllowedMethods),
		&endpoint.CreatedAt, &endpoint.UpdatedAt, &endpoint.CreatedBy, &endpoint.IsActive, &metadata,
	)
	endpoint.Metadata = metadata.Data

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("endpoint not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoint: %w", err)
	}

	return endpoint, nil
}

// GetByNameWithNamespace retrieves an endpoint by name with its namespace
func (r *EndpointRepository) GetByNameWithNamespace(ctx context.Context, name string) (*types.Endpoint, error) {
	endpoint, err := r.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}

	// Get namespace details
	namespaceRepo := NewNamespaceRepository(r.db)
	namespace, err := namespaceRepo.GetByIDWithServers(ctx, endpoint.NamespaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace for endpoint: %w", err)
	}

	endpoint.Namespace = namespace
	return endpoint, nil
}

// List retrieves all endpoints for an organization
func (r *EndpointRepository) List(ctx context.Context, orgID string) ([]*types.Endpoint, error) {
	query := `
		SELECT
			id, organization_id, namespace_id, name, description,
			enable_api_key_auth, enable_oauth, enable_public_access, use_query_param_auth,
			rate_limit_requests, rate_limit_window,
			allowed_origins, allowed_methods,
			created_at, updated_at, created_by, is_active, metadata
		FROM endpoints
		WHERE organization_id = $1
		ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list endpoints: %w", err)
	}
	defer rows.Close()

	var endpoints []*types.Endpoint
	for rows.Next() {
		endpoint := &types.Endpoint{}
		var metadata NullableJSONB
		err := rows.Scan(
			&endpoint.ID, &endpoint.OrganizationID, &endpoint.NamespaceID, &endpoint.Name, &endpoint.Description,
			&endpoint.EnableAPIKeyAuth, &endpoint.EnableOAuth, &endpoint.EnablePublicAccess, &endpoint.UseQueryParamAuth,
			&endpoint.RateLimitRequests, &endpoint.RateLimitWindow,
			pq.Array(&endpoint.AllowedOrigins), pq.Array(&endpoint.AllowedMethods),
			&endpoint.CreatedAt, &endpoint.UpdatedAt, &endpoint.CreatedBy, &endpoint.IsActive, &metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan endpoint: %w", err)
		}
		endpoint.Metadata = metadata.Data
		endpoints = append(endpoints, endpoint)
	}

	return endpoints, nil
}

// ListPublic retrieves all public endpoints
func (r *EndpointRepository) ListPublic(ctx context.Context) ([]*types.Endpoint, error) {
	query := `
		SELECT
			id, organization_id, namespace_id, name, description,
			enable_api_key_auth, enable_oauth, enable_public_access, use_query_param_auth,
			rate_limit_requests, rate_limit_window,
			allowed_origins, allowed_methods,
			created_at, updated_at, created_by, is_active, metadata
		FROM endpoints
		WHERE is_active = true AND enable_public_access = true
		ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list public endpoints: %w", err)
	}
	defer rows.Close()

	var endpoints []*types.Endpoint
	for rows.Next() {
		endpoint := &types.Endpoint{}
		var metadata NullableJSONB
		err := rows.Scan(
			&endpoint.ID, &endpoint.OrganizationID, &endpoint.NamespaceID, &endpoint.Name, &endpoint.Description,
			&endpoint.EnableAPIKeyAuth, &endpoint.EnableOAuth, &endpoint.EnablePublicAccess, &endpoint.UseQueryParamAuth,
			&endpoint.RateLimitRequests, &endpoint.RateLimitWindow,
			pq.Array(&endpoint.AllowedOrigins), pq.Array(&endpoint.AllowedMethods),
			&endpoint.CreatedAt, &endpoint.UpdatedAt, &endpoint.CreatedBy, &endpoint.IsActive, &metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan endpoint: %w", err)
		}
		endpoint.Metadata = metadata.Data
		endpoints = append(endpoints, endpoint)
	}

	return endpoints, nil
}

// Update updates an endpoint
func (r *EndpointRepository) Update(ctx context.Context, endpoint *types.Endpoint) error {
	query := `
		UPDATE endpoints SET
			description = $2,
			enable_api_key_auth = $3,
			enable_oauth = $4,
			enable_public_access = $5,
			use_query_param_auth = $6,
			rate_limit_requests = $7,
			rate_limit_window = $8,
			allowed_origins = $9,
			allowed_methods = $10,
			is_active = $11,
			metadata = $12,
			updated_at = NOW()
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		endpoint.ID, endpoint.Description,
		endpoint.EnableAPIKeyAuth, endpoint.EnableOAuth, endpoint.EnablePublicAccess, endpoint.UseQueryParamAuth,
		endpoint.RateLimitRequests, endpoint.RateLimitWindow,
		pq.Array(endpoint.AllowedOrigins), pq.Array(endpoint.AllowedMethods),
		endpoint.IsActive, endpoint.Metadata,
	)

	if err != nil {
		return fmt.Errorf("failed to update endpoint: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("endpoint not found")
	}

	return nil
}

// Delete deletes an endpoint
func (r *EndpointRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM endpoints WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete endpoint: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("endpoint not found")
	}

	return nil
}

// ValidateNameUniqueness checks if an endpoint name is unique
func (r *EndpointRepository) ValidateNameUniqueness(ctx context.Context, name string) error {
	var count int
	query := `SELECT COUNT(*) FROM endpoints WHERE name = $1`

	err := r.db.QueryRowContext(ctx, query, name).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check name uniqueness: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("endpoint name '%s' already exists", name)
	}

	return nil
}
