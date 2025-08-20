package models

import (
	"database/sql"
)

// Database interface defines database operations
type Database interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	Begin() (*sql.Tx, error)
}

// BaseModel provides common database functionality
type BaseModel struct {
	db Database
}

// Transaction wraps a function in a database transaction
func (m *BaseModel) Transaction(fn func(*sql.Tx) error) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Models holds all model instances
type Models struct {
	User         *UserModel
	Organization *OrganizationModel
	APIKey       *APIKeyModel
	Policy       *PolicyModel
	MCPServer    *MCPServerModel
	HealthCheck  *HealthCheckModel
	// LogEntry      *LogEntryModel
	// AuditLog      *AuditLogModel
	RateLimitRule *RateLimitRuleModel
}

// NewModels creates a new Models instance
func NewModels(db Database) *Models {
	return &Models{
		User:         NewUserModel(db),
		Organization: NewOrganizationModel(db),
		APIKey:       NewAPIKeyModel(db),
		Policy:       NewPolicyModel(db),
		MCPServer:    NewMCPServerModel(db),
		HealthCheck:  NewHealthCheckModel(db),
		// LogEntry:     NewLogEntryModel(db),
		// AuditLog:     NewAuditLogModel(db),
		RateLimitRule: NewRateLimitRuleModel(db),
	}
}
