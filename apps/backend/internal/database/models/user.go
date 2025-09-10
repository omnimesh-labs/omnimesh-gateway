package models

import (
	"time"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"
)

// UserModel handles user database operations
type UserModel struct {
	BaseModel
}

// NewUserModel creates a new user model
func NewUserModel(db Database) *UserModel {
	return &UserModel{
		BaseModel: BaseModel{db: db},
	}
}

// Create inserts a new user into the database
func (m *UserModel) Create(user *types.User) error {
	query := `
		INSERT INTO users (id, email, name, password_hash, organization_id, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := m.db.Exec(query, user.ID, user.Email, user.Name, user.PasswordHash,
		user.OrganizationID, user.Role, user.IsActive, user.CreatedAt, user.UpdatedAt)

	return err
}

// GetByID retrieves a user by ID
func (m *UserModel) GetByID(id string) (*types.User, error) {
	query := `
		SELECT id, email, name, password_hash, organization_id, role, is_active, created_at, updated_at
		FROM users
		WHERE id = $1 AND is_active = true
	`

	user := &types.User{}
	err := m.db.QueryRow(query, id).Scan(
		&user.ID, &user.Email, &user.Name, &user.PasswordHash,
		&user.OrganizationID, &user.Role, &user.IsActive,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (m *UserModel) GetByEmail(email string) (*types.User, error) {
	query := `
		SELECT id, email, name, password_hash, organization_id, role, is_active, created_at, updated_at
		FROM users
		WHERE email = $1 AND is_active = true
	`

	user := &types.User{}
	err := m.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.Name, &user.PasswordHash,
		&user.OrganizationID, &user.Role, &user.IsActive,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// Update updates a user in the database
func (m *UserModel) Update(user *types.User) error {
	query := `
		UPDATE users
		SET name = $1, role = $2, updated_at = $3
		WHERE id = $4
	`

	user.UpdatedAt = time.Now()

	_, err := m.db.Exec(query, user.Name, user.Role, user.UpdatedAt, user.ID)
	return err
}

// Delete soft deletes a user
func (m *UserModel) Delete(id string) error {
	query := `
		UPDATE users
		SET is_active = false, updated_at = $1
		WHERE id = $2
	`

	_, err := m.db.Exec(query, time.Now(), id)
	return err
}

// ListByOrganization lists users for an organization
func (m *UserModel) ListByOrganization(orgID string, limit, offset int) ([]*types.User, error) {
	query := `
		SELECT id, email, name, password_hash, organization_id, role, is_active, created_at, updated_at
		FROM users
		WHERE organization_id = $1 AND is_active = true
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := m.db.Query(query, orgID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*types.User
	for rows.Next() {
		user := &types.User{}
		err := rows.Scan(
			&user.ID, &user.Email, &user.Name, &user.PasswordHash,
			&user.OrganizationID, &user.Role, &user.IsActive,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// Count returns the total number of users for an organization
func (m *UserModel) Count(orgID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM users
		WHERE organization_id = $1 AND is_active = true
	`

	var count int
	err := m.db.QueryRow(query, orgID).Scan(&count)
	return count, err
}
