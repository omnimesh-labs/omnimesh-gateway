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
	// ERD-based models
	Organization   *OrganizationModel
	MCPServer      *MCPServerModel
	MCPSession     *MCPSessionModel
	HealthCheck    *HealthCheckModel
	ServerStats    *ServerStatsModel
	LogIndex       *LogIndexModel
	AuditLog       *AuditLogModel
	LogAggregate   *LogAggregateModel
	RateLimit      *RateLimitModel
	RateLimitUsage *RateLimitUsageModel

	// Legacy models (deprecated - will be removed in future versions)
	User *UserModel
}

// NewModels creates a new Models instance
func NewModels(db Database) *Models {
	return &Models{
		// ERD-based models
		Organization:   NewOrganizationModel(db),
		MCPServer:      NewMCPServerModel(db),
		MCPSession:     NewMCPSessionModel(db),
		HealthCheck:    NewHealthCheckModel(db),
		ServerStats:    NewServerStatsModel(db),
		LogIndex:       NewLogIndexModel(db),
		AuditLog:       NewAuditLogModel(db),
		LogAggregate:   NewLogAggregateModel(db),
		RateLimit:      NewRateLimitModel(db),
		RateLimitUsage: NewRateLimitUsageModel(db),

		// Legacy models (deprecated)
		User: NewUserModel(db),
	}
}
