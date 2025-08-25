package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// RateLimit represents the rate_limits table from the ERD
type RateLimit struct {
	CreatedAt         time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at" json:"updated_at"`
	Scope             string         `db:"scope" json:"scope"`
	ScopeID           sql.NullString `db:"scope_id" json:"scope_id,omitempty"`
	RequestsPerMinute int            `db:"requests_per_minute" json:"requests_per_minute"`
	RequestsPerHour   sql.NullInt32  `db:"requests_per_hour" json:"requests_per_hour,omitempty"`
	RequestsPerDay    sql.NullInt32  `db:"requests_per_day" json:"requests_per_day,omitempty"`
	BurstLimit        sql.NullInt32  `db:"burst_limit" json:"burst_limit,omitempty"`
	ID                uuid.UUID      `db:"id" json:"id"`
	OrganizationID    uuid.UUID      `db:"organization_id" json:"organization_id"`
	IsActive          bool           `db:"is_active" json:"is_active"`
}

// RateLimitUsage represents the rate_limit_usage table from the ERD
type RateLimitUsage struct {
	WindowStart    time.Time `db:"window_start" json:"window_start"`
	ExpiresAt      time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
	Identifier     string    `db:"identifier" json:"identifier"`
	RequestCount   int       `db:"request_count" json:"request_count"`
	ID             uuid.UUID `db:"id" json:"id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`
	RateLimitID    uuid.UUID `db:"rate_limit_id" json:"rate_limit_id"`
}

// ServerStats represents the server_stats table from the ERD
type ServerStats struct {
	WindowStart       time.Time `db:"window_start" json:"window_start"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	WindowEnd         time.Time `db:"window_end" json:"window_end"`
	SuccessRequests   int64     `db:"success_requests" json:"success_requests"`
	ActiveSessions    int       `db:"active_sessions" json:"active_sessions"`
	AvgResponseTimeMS float64   `db:"avg_response_time_ms" json:"avg_response_time_ms"`
	MinResponseTimeMS int       `db:"min_response_time_ms" json:"min_response_time_ms"`
	MaxResponseTimeMS int       `db:"max_response_time_ms" json:"max_response_time_ms"`
	ErrorRequests     int64     `db:"error_requests" json:"error_requests"`
	TotalRequests     int64     `db:"total_requests" json:"total_requests"`
	ID                uuid.UUID `db:"id" json:"id"`
	ServerID          uuid.UUID `db:"server_id" json:"server_id"`
}

// RateLimitModel handles rate limit database operations
type RateLimitModel struct {
	db Database
}

// NewRateLimitModel creates a new rate limit model
func NewRateLimitModel(db Database) *RateLimitModel {
	return &RateLimitModel{db: db}
}

// Create inserts a new rate limit rule
func (m *RateLimitModel) Create(limit *RateLimit) error {
	query := `
		INSERT INTO rate_limits (
			id, organization_id, scope, scope_id, requests_per_minute,
			requests_per_hour, requests_per_day, burst_limit, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	if limit.ID == uuid.Nil {
		limit.ID = uuid.New()
	}

	_, err := m.db.Exec(query,
		limit.ID, limit.OrganizationID, limit.Scope, limit.ScopeID,
		limit.RequestsPerMinute, limit.RequestsPerHour, limit.RequestsPerDay,
		limit.BurstLimit, limit.IsActive)
	return err
}

// GetByID retrieves a rate limit by ID
func (m *RateLimitModel) GetByID(id uuid.UUID) (*RateLimit, error) {
	query := `
		SELECT id, organization_id, scope, scope_id, requests_per_minute,
			   requests_per_hour, requests_per_day, burst_limit, is_active,
			   created_at, updated_at
		FROM rate_limits
		WHERE id = $1
	`

	limit := &RateLimit{}
	err := m.db.QueryRow(query, id).Scan(
		&limit.ID, &limit.OrganizationID, &limit.Scope, &limit.ScopeID,
		&limit.RequestsPerMinute, &limit.RequestsPerHour, &limit.RequestsPerDay,
		&limit.BurstLimit, &limit.IsActive, &limit.CreatedAt, &limit.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return limit, nil
}

// GetByScope retrieves rate limits by scope and scope ID
func (m *RateLimitModel) GetByScope(orgID uuid.UUID, scope string, scopeID *string) ([]*RateLimit, error) {
	query := `
		SELECT id, organization_id, scope, scope_id, requests_per_minute,
			   requests_per_hour, requests_per_day, burst_limit, is_active,
			   created_at, updated_at
		FROM rate_limits
		WHERE organization_id = $1 AND scope = $2 AND is_active = true
	`

	args := []interface{}{orgID, scope}

	if scopeID != nil {
		query += " AND scope_id = $3"
		args = append(args, *scopeID)
	} else {
		query += " AND scope_id IS NULL"
	}

	query += " ORDER BY created_at DESC"

	rows, err := m.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var limits []*RateLimit
	for rows.Next() {
		limit := &RateLimit{}
		err := rows.Scan(
			&limit.ID, &limit.OrganizationID, &limit.Scope, &limit.ScopeID,
			&limit.RequestsPerMinute, &limit.RequestsPerHour, &limit.RequestsPerDay,
			&limit.BurstLimit, &limit.IsActive, &limit.CreatedAt, &limit.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		limits = append(limits, limit)
	}

	return limits, nil
}

// ListByOrganization lists rate limits for an organization
func (m *RateLimitModel) ListByOrganization(orgID uuid.UUID, activeOnly bool) ([]*RateLimit, error) {
	query := `
		SELECT id, organization_id, scope, scope_id, requests_per_minute,
			   requests_per_hour, requests_per_day, burst_limit, is_active,
			   created_at, updated_at
		FROM rate_limits
		WHERE organization_id = $1
	`

	args := []interface{}{orgID}
	if activeOnly {
		query += " AND is_active = true"
	}
	query += " ORDER BY scope, created_at DESC"

	rows, err := m.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var limits []*RateLimit
	for rows.Next() {
		limit := &RateLimit{}
		err := rows.Scan(
			&limit.ID, &limit.OrganizationID, &limit.Scope, &limit.ScopeID,
			&limit.RequestsPerMinute, &limit.RequestsPerHour, &limit.RequestsPerDay,
			&limit.BurstLimit, &limit.IsActive, &limit.CreatedAt, &limit.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		limits = append(limits, limit)
	}

	return limits, nil
}

// Update updates a rate limit rule
func (m *RateLimitModel) Update(limit *RateLimit) error {
	query := `
		UPDATE rate_limits
		SET scope = $2, scope_id = $3, requests_per_minute = $4,
			requests_per_hour = $5, requests_per_day = $6, burst_limit = $7,
			is_active = $8
		WHERE id = $1
	`

	_, err := m.db.Exec(query,
		limit.ID, limit.Scope, limit.ScopeID, limit.RequestsPerMinute,
		limit.RequestsPerHour, limit.RequestsPerDay, limit.BurstLimit, limit.IsActive)
	return err
}

// Delete deletes a rate limit rule
func (m *RateLimitModel) Delete(id uuid.UUID) error {
	query := `DELETE FROM rate_limits WHERE id = $1`
	_, err := m.db.Exec(query, id)
	return err
}

// RateLimitUsageModel handles rate limit usage database operations
type RateLimitUsageModel struct {
	db Database
}

// NewRateLimitUsageModel creates a new rate limit usage model
func NewRateLimitUsageModel(db Database) *RateLimitUsageModel {
	return &RateLimitUsageModel{db: db}
}

// IncrementUsage increments the usage count for a rate limit
func (m *RateLimitUsageModel) IncrementUsage(orgID, rateLimitID uuid.UUID, identifier string, windowStart time.Time, expiresAt time.Time) error {
	query := `
		INSERT INTO rate_limit_usage (
			id, organization_id, rate_limit_id, identifier, window_start,
			request_count, expires_at
		) VALUES ($1, $2, $3, $4, $5, 1, $6)
		ON CONFLICT (rate_limit_id, identifier, window_start)
		DO UPDATE SET request_count = rate_limit_usage.request_count + 1
	`

	id := uuid.New()
	_, err := m.db.Exec(query, id, orgID, rateLimitID, identifier, windowStart, expiresAt)
	return err
}

// GetUsage retrieves current usage for a rate limit
func (m *RateLimitUsageModel) GetUsage(rateLimitID uuid.UUID, identifier string, windowStart time.Time) (*RateLimitUsage, error) {
	query := `
		SELECT id, organization_id, rate_limit_id, identifier, window_start,
			   request_count, expires_at, created_at, updated_at
		FROM rate_limit_usage
		WHERE rate_limit_id = $1 AND identifier = $2 AND window_start = $3
	`

	usage := &RateLimitUsage{}
	err := m.db.QueryRow(query, rateLimitID, identifier, windowStart).Scan(
		&usage.ID, &usage.OrganizationID, &usage.RateLimitID, &usage.Identifier,
		&usage.WindowStart, &usage.RequestCount, &usage.ExpiresAt,
		&usage.CreatedAt, &usage.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return usage, nil
}

// GetUsageInRange retrieves usage for a rate limit within a time range
func (m *RateLimitUsageModel) GetUsageInRange(rateLimitID uuid.UUID, identifier string, startTime, endTime time.Time) ([]*RateLimitUsage, error) {
	query := `
		SELECT id, organization_id, rate_limit_id, identifier, window_start,
			   request_count, expires_at, created_at, updated_at
		FROM rate_limit_usage
		WHERE rate_limit_id = $1 AND identifier = $2
			  AND window_start >= $3 AND window_start <= $4
		ORDER BY window_start DESC
	`

	rows, err := m.db.Query(query, rateLimitID, identifier, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usages []*RateLimitUsage
	for rows.Next() {
		usage := &RateLimitUsage{}
		err := rows.Scan(
			&usage.ID, &usage.OrganizationID, &usage.RateLimitID, &usage.Identifier,
			&usage.WindowStart, &usage.RequestCount, &usage.ExpiresAt,
			&usage.CreatedAt, &usage.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		usages = append(usages, usage)
	}

	return usages, nil
}

// CleanupExpiredUsage removes expired usage records
func (m *RateLimitUsageModel) CleanupExpiredUsage() error {
	query := `DELETE FROM rate_limit_usage WHERE expires_at < NOW()`
	_, err := m.db.Exec(query)
	return err
}

// ServerStatsModel handles server stats database operations
type ServerStatsModel struct {
	db Database
}

// NewServerStatsModel creates a new server stats model
func NewServerStatsModel(db Database) *ServerStatsModel {
	return &ServerStatsModel{db: db}
}

// Create inserts a new server stats record
func (m *ServerStatsModel) Create(stats *ServerStats) error {
	query := `
		INSERT INTO server_stats (
			id, server_id, total_requests, success_requests, error_requests,
			active_sessions, avg_response_time_ms, min_response_time_ms,
			max_response_time_ms, window_start, window_end
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	if stats.ID == uuid.Nil {
		stats.ID = uuid.New()
	}

	_, err := m.db.Exec(query,
		stats.ID, stats.ServerID, stats.TotalRequests, stats.SuccessRequests,
		stats.ErrorRequests, stats.ActiveSessions, stats.AvgResponseTimeMS,
		stats.MinResponseTimeMS, stats.MaxResponseTimeMS, stats.WindowStart, stats.WindowEnd)
	return err
}

// GetLatestByServerID retrieves the latest stats for a server
func (m *ServerStatsModel) GetLatestByServerID(serverID uuid.UUID) (*ServerStats, error) {
	query := `
		SELECT id, server_id, total_requests, success_requests, error_requests,
			   active_sessions, avg_response_time_ms, min_response_time_ms,
			   max_response_time_ms, window_start, window_end, created_at, updated_at
		FROM server_stats
		WHERE server_id = $1
		ORDER BY window_end DESC
		LIMIT 1
	`

	stats := &ServerStats{}
	err := m.db.QueryRow(query, serverID).Scan(
		&stats.ID, &stats.ServerID, &stats.TotalRequests, &stats.SuccessRequests,
		&stats.ErrorRequests, &stats.ActiveSessions, &stats.AvgResponseTimeMS,
		&stats.MinResponseTimeMS, &stats.MaxResponseTimeMS, &stats.WindowStart,
		&stats.WindowEnd, &stats.CreatedAt, &stats.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetHistoryByServerID retrieves stats history for a server
func (m *ServerStatsModel) GetHistoryByServerID(serverID uuid.UUID, limit int) ([]*ServerStats, error) {
	query := `
		SELECT id, server_id, total_requests, success_requests, error_requests,
			   active_sessions, avg_response_time_ms, min_response_time_ms,
			   max_response_time_ms, window_start, window_end, created_at, updated_at
		FROM server_stats
		WHERE server_id = $1
		ORDER BY window_end DESC
		LIMIT $2
	`

	rows, err := m.db.Query(query, serverID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statsList []*ServerStats
	for rows.Next() {
		stats := &ServerStats{}
		err := rows.Scan(
			&stats.ID, &stats.ServerID, &stats.TotalRequests, &stats.SuccessRequests,
			&stats.ErrorRequests, &stats.ActiveSessions, &stats.AvgResponseTimeMS,
			&stats.MinResponseTimeMS, &stats.MaxResponseTimeMS, &stats.WindowStart,
			&stats.WindowEnd, &stats.CreatedAt, &stats.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		statsList = append(statsList, stats)
	}

	return statsList, nil
}

// Upsert creates or updates server stats for a given window
func (m *ServerStatsModel) Upsert(stats *ServerStats) error {
	query := `
		INSERT INTO server_stats (
			id, server_id, total_requests, success_requests, error_requests,
			active_sessions, avg_response_time_ms, min_response_time_ms,
			max_response_time_ms, window_start, window_end
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (server_id, window_start)
		DO UPDATE SET
			total_requests = EXCLUDED.total_requests,
			success_requests = EXCLUDED.success_requests,
			error_requests = EXCLUDED.error_requests,
			active_sessions = EXCLUDED.active_sessions,
			avg_response_time_ms = EXCLUDED.avg_response_time_ms,
			min_response_time_ms = EXCLUDED.min_response_time_ms,
			max_response_time_ms = EXCLUDED.max_response_time_ms
	`

	if stats.ID == uuid.Nil {
		stats.ID = uuid.New()
	}

	_, err := m.db.Exec(query,
		stats.ID, stats.ServerID, stats.TotalRequests, stats.SuccessRequests,
		stats.ErrorRequests, stats.ActiveSessions, stats.AvgResponseTimeMS,
		stats.MinResponseTimeMS, stats.MaxResponseTimeMS, stats.WindowStart, stats.WindowEnd)
	return err
}
