package models

import (
	"time"

	"mcp-gateway/apps/backend/internal/types"
)

// OrganizationModel handles organization database operations
type OrganizationModel struct {
	BaseModel
}

// NewOrganizationModel creates a new organization model
func NewOrganizationModel(db Database) *OrganizationModel {
	return &OrganizationModel{
		BaseModel: BaseModel{db: db},
	}
}

// Create inserts a new organization into the database
func (m *OrganizationModel) Create(org *types.Organization) error {
	query := `
		INSERT INTO organizations (id, name, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	now := time.Now()
	org.CreatedAt = now
	org.UpdatedAt = now

	_, err := m.db.Exec(query, org.ID, org.Name, org.Description, org.IsActive, org.CreatedAt, org.UpdatedAt)
	return err
}

// GetByID retrieves an organization by ID
func (m *OrganizationModel) GetByID(id string) (*types.Organization, error) {
	query := `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM organizations
		WHERE id = $1 AND is_active = true
	`

	org := &types.Organization{}
	err := m.db.QueryRow(query, id).Scan(
		&org.ID, &org.Name, &org.Description, &org.IsActive, &org.CreatedAt, &org.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return org, nil
}

// Update updates an organization in the database
func (m *OrganizationModel) Update(org *types.Organization) error {
	query := `
		UPDATE organizations
		SET name = $1, description = $2, updated_at = $3
		WHERE id = $4
	`

	org.UpdatedAt = time.Now()

	_, err := m.db.Exec(query, org.Name, org.Description, org.UpdatedAt, org.ID)
	return err
}

// Delete soft deletes an organization
func (m *OrganizationModel) Delete(id string) error {
	query := `
		UPDATE organizations
		SET is_active = false, updated_at = $1
		WHERE id = $2
	`

	_, err := m.db.Exec(query, time.Now(), id)
	return err
}

// List lists all active organizations
func (m *OrganizationModel) List(limit, offset int) ([]*types.Organization, error) {
	query := `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM organizations
		WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := m.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*types.Organization
	for rows.Next() {
		org := &types.Organization{}
		err := rows.Scan(
			&org.ID, &org.Name, &org.Description, &org.IsActive, &org.CreatedAt, &org.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}

	return orgs, nil
}

// APIKeyModel handles API key database operations
type APIKeyModel struct {
	BaseModel
}

// NewAPIKeyModel creates a new API key model
func NewAPIKeyModel(db Database) *APIKeyModel {
	return &APIKeyModel{
		BaseModel: BaseModel{db: db},
	}
}

// Create inserts a new API key into the database
func (m *APIKeyModel) Create(apiKey *types.APIKey) error {
	query := `
		INSERT INTO api_keys (id, user_id, organization_id, name, key_hash, prefix, permissions,
			expires_at, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	now := time.Now()
	apiKey.CreatedAt = now
	apiKey.UpdatedAt = now

	// Convert permissions slice to PostgreSQL array format
	// This is a simplified version - in production, you might want to use a proper JSON column
	permissionsJSON := ""
	if len(apiKey.Permissions) > 0 {
		// Convert to JSON string for storage
		// TODO: Use proper JSON marshaling
	}

	_, err := m.db.Exec(query, apiKey.ID, apiKey.UserID, apiKey.OrganizationID,
		apiKey.Name, apiKey.KeyHash, apiKey.Prefix, permissionsJSON,
		apiKey.ExpiresAt, apiKey.IsActive, apiKey.CreatedAt, apiKey.UpdatedAt)

	return err
}

// GetByHash retrieves an API key by its hash
func (m *APIKeyModel) GetByHash(keyHash string) (*types.APIKey, error) {
	query := `
		SELECT id, user_id, organization_id, name, key_hash, prefix, permissions,
			expires_at, last_used_at, is_active, created_at, updated_at
		FROM api_keys
		WHERE key_hash = $1 AND is_active = true
	`

	apiKey := &types.APIKey{}
	var permissionsJSON string

	err := m.db.QueryRow(query, keyHash).Scan(
		&apiKey.ID, &apiKey.UserID, &apiKey.OrganizationID, &apiKey.Name,
		&apiKey.KeyHash, &apiKey.Prefix, &permissionsJSON, &apiKey.ExpiresAt,
		&apiKey.LastUsedAt, &apiKey.IsActive, &apiKey.CreatedAt, &apiKey.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse permissions JSON
	// TODO: Implement proper JSON unmarshaling

	return apiKey, nil
}

// ListByUser lists API keys for a user
func (m *APIKeyModel) ListByUser(userID string) ([]*types.APIKey, error) {
	query := `
		SELECT id, user_id, organization_id, name, key_hash, prefix, permissions,
			expires_at, last_used_at, is_active, created_at, updated_at
		FROM api_keys
		WHERE user_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`

	rows, err := m.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apiKeys []*types.APIKey
	for rows.Next() {
		apiKey := &types.APIKey{}
		var permissionsJSON string

		err := rows.Scan(
			&apiKey.ID, &apiKey.UserID, &apiKey.OrganizationID, &apiKey.Name,
			&apiKey.KeyHash, &apiKey.Prefix, &permissionsJSON, &apiKey.ExpiresAt,
			&apiKey.LastUsedAt, &apiKey.IsActive, &apiKey.CreatedAt, &apiKey.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse permissions JSON
		// TODO: Implement proper JSON unmarshaling

		apiKeys = append(apiKeys, apiKey)
	}

	return apiKeys, nil
}

// UpdateLastUsed updates the last used timestamp for an API key
func (m *APIKeyModel) UpdateLastUsed(id string) error {
	query := `
		UPDATE api_keys
		SET last_used_at = $1, updated_at = $2
		WHERE id = $3
	`

	now := time.Now()
	_, err := m.db.Exec(query, now, now, id)
	return err
}

// Revoke revokes an API key
func (m *APIKeyModel) Revoke(id string) error {
	query := `
		UPDATE api_keys
		SET is_active = false, updated_at = $1
		WHERE id = $2
	`

	_, err := m.db.Exec(query, time.Now(), id)
	return err
}
