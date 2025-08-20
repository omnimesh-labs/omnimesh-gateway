package models

import (
	"time"

	"github.com/google/uuid"
)

// Organization represents the organizations table from the ERD
type Organization struct {
	ID               uuid.UUID `db:"id" json:"id"`
	Name             string    `db:"name" json:"name"`
	Slug             string    `db:"slug" json:"slug"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
	IsActive         bool      `db:"is_active" json:"is_active"`
	PlanType         string    `db:"plan_type" json:"plan_type"`
	MaxServers       int       `db:"max_servers" json:"max_servers"`
	MaxSessions      int       `db:"max_sessions" json:"max_sessions"`
	LogRetentionDays int       `db:"log_retention_days" json:"log_retention_days"`
}

// OrganizationModel handles organization database operations
type OrganizationModel struct {
	db Database
}

// NewOrganizationModel creates a new organization model
func NewOrganizationModel(db Database) *OrganizationModel {
	return &OrganizationModel{db: db}
}

// Create inserts a new organization
func (m *OrganizationModel) Create(org *Organization) error {
	query := `
		INSERT INTO organizations (id, name, slug, is_active, plan_type, max_servers, max_sessions, log_retention_days)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	if org.ID == uuid.Nil {
		org.ID = uuid.New()
	}

	_, err := m.db.Exec(query,
		org.ID, org.Name, org.Slug, org.IsActive,
		org.PlanType, org.MaxServers, org.MaxSessions, org.LogRetentionDays)
	return err
}

// GetByID retrieves an organization by ID
func (m *OrganizationModel) GetByID(id uuid.UUID) (*Organization, error) {
	query := `
		SELECT id, name, slug, created_at, updated_at, is_active, 
			   plan_type, max_servers, max_sessions, log_retention_days
		FROM organizations
		WHERE id = $1
	`

	org := &Organization{}
	err := m.db.QueryRow(query, id).Scan(
		&org.ID, &org.Name, &org.Slug, &org.CreatedAt, &org.UpdatedAt,
		&org.IsActive, &org.PlanType, &org.MaxServers, &org.MaxSessions, &org.LogRetentionDays,
	)

	if err != nil {
		return nil, err
	}

	return org, nil
}

// GetBySlug retrieves an organization by slug
func (m *OrganizationModel) GetBySlug(slug string) (*Organization, error) {
	query := `
		SELECT id, name, slug, created_at, updated_at, is_active, 
			   plan_type, max_servers, max_sessions, log_retention_days
		FROM organizations
		WHERE slug = $1
	`

	org := &Organization{}
	err := m.db.QueryRow(query, slug).Scan(
		&org.ID, &org.Name, &org.Slug, &org.CreatedAt, &org.UpdatedAt,
		&org.IsActive, &org.PlanType, &org.MaxServers, &org.MaxSessions, &org.LogRetentionDays,
	)

	if err != nil {
		return nil, err
	}

	return org, nil
}

// GetDefault retrieves the default organization
func (m *OrganizationModel) GetDefault() (*Organization, error) {
	defaultID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	return m.GetByID(defaultID)
}

// Update updates an organization
func (m *OrganizationModel) Update(org *Organization) error {
	query := `
		UPDATE organizations
		SET name = $2, slug = $3, is_active = $4, plan_type = $5,
			max_servers = $6, max_sessions = $7, log_retention_days = $8
		WHERE id = $1
	`

	_, err := m.db.Exec(query,
		org.ID, org.Name, org.Slug, org.IsActive,
		org.PlanType, org.MaxServers, org.MaxSessions, org.LogRetentionDays)
	return err
}

// List lists all organizations
func (m *OrganizationModel) List(limit, offset int) ([]*Organization, error) {
	query := `
		SELECT id, name, slug, created_at, updated_at, is_active, 
			   plan_type, max_servers, max_sessions, log_retention_days
		FROM organizations
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := m.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*Organization
	for rows.Next() {
		org := &Organization{}
		err := rows.Scan(
			&org.ID, &org.Name, &org.Slug, &org.CreatedAt, &org.UpdatedAt,
			&org.IsActive, &org.PlanType, &org.MaxServers, &org.MaxSessions, &org.LogRetentionDays,
		)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}

	return orgs, nil
}
