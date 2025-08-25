package types

import (
	"context"
	"time"
)

// RateLimitRule represents a rate limiting rule
type RateLimitRule struct {
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
	Conditions     map[string]interface{} `json:"conditions" db:"conditions"`
	OrganizationID string                 `json:"organization_id" db:"organization_id"`
	Name           string                 `json:"name" db:"name"`
	Description    string                 `json:"description" db:"description"`
	Type           string                 `json:"type" db:"type"`
	ID             string                 `json:"id" db:"id"`
	Algorithm      string                 `json:"algorithm" db:"algorithm"`
	Window         time.Duration          `json:"window" db:"window"`
	Priority       int                    `json:"priority" db:"priority"`
	Limit          int                    `json:"limit" db:"limit"`
	IsActive       bool                   `json:"is_active" db:"is_active"`
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp      time.Time              `json:"timestamp" db:"timestamp"`
	Data           map[string]interface{} `json:"data,omitempty" db:"data"`
	Path           string                 `json:"path,omitempty" db:"path"`
	UserAgent      string                 `json:"user_agent,omitempty" db:"user_agent"`
	UserID         string                 `json:"user_id,omitempty" db:"user_id"`
	OrganizationID string                 `json:"organization_id,omitempty" db:"organization_id"`
	RequestID      string                 `json:"request_id,omitempty" db:"request_id"`
	Method         string                 `json:"method,omitempty" db:"method"`
	ID             string                 `json:"id" db:"id"`
	Error          string                 `json:"error,omitempty" db:"error"`
	Level          string                 `json:"level" db:"level"`
	RemoteIP       string                 `json:"remote_ip,omitempty" db:"remote_ip"`
	Type           string                 `json:"type" db:"type"`
	Message        string                 `json:"message" db:"message"`
	Duration       time.Duration          `json:"duration,omitempty" db:"duration"`
	StatusCode     int                    `json:"status_code,omitempty" db:"status_code"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	Timestamp      time.Time              `json:"timestamp" db:"timestamp"`
	Details        map[string]interface{} `json:"details" db:"details"`
	ID             string                 `json:"id" db:"id"`
	UserID         string                 `json:"user_id" db:"user_id"`
	OrganizationID string                 `json:"organization_id" db:"organization_id"`
	Action         string                 `json:"action" db:"action"`
	Resource       string                 `json:"resource" db:"resource"`
	ResourceID     string                 `json:"resource_id" db:"resource_id"`
	RemoteIP       string                 `json:"remote_ip" db:"remote_ip"`
	UserAgent      string                 `json:"user_agent" db:"user_agent"`
	Error          string                 `json:"error,omitempty" db:"error"`
	Success        bool                   `json:"success" db:"success"`
}

// Metric represents a performance metric
type Metric struct {
	Timestamp      time.Time              `json:"timestamp" db:"timestamp"`
	Tags           map[string]string      `json:"tags,omitempty" db:"tags"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	ID             string                 `json:"id" db:"id"`
	OrganizationID string                 `json:"organization_id,omitempty" db:"organization_id"`
	ServerID       string                 `json:"server_id,omitempty" db:"server_id"`
	Name           string                 `json:"name" db:"name"`
	Type           string                 `json:"type" db:"type"`
	Value          float64                `json:"value" db:"value"`
}

// CreateRateLimitRuleRequest represents a rate limit rule creation request
type CreateRateLimitRuleRequest struct {
	Conditions  map[string]interface{} `json:"conditions"`
	Name        string                 `json:"name" binding:"required,min=2"`
	Description string                 `json:"description"`
	Type        string                 `json:"type" binding:"required"`
	Algorithm   string                 `json:"algorithm" binding:"required"`
	Limit       int                    `json:"limit" binding:"required,min=1"`
	Window      time.Duration          `json:"window" binding:"required"`
	Priority    int                    `json:"priority"`
}

// UpdateRateLimitRuleRequest represents a rate limit rule update request
type UpdateRateLimitRuleRequest struct {
	Conditions  map[string]interface{} `json:"conditions,omitempty"`
	IsActive    *bool                  `json:"is_active,omitempty"`
	Name        string                 `json:"name,omitempty" binding:"omitempty,min=2"`
	Description string                 `json:"description,omitempty"`
	Algorithm   string                 `json:"algorithm,omitempty"`
	Limit       int                    `json:"limit,omitempty" binding:"omitempty,min=1"`
	Window      time.Duration          `json:"window,omitempty"`
	Priority    int                    `json:"priority,omitempty"`
}

// LogQueryRequest represents a log query request
type LogQueryRequest struct {
	StartTime      time.Time `json:"start_time" form:"start_time"`
	EndTime        time.Time `json:"end_time" form:"end_time"`
	Level          string    `json:"level,omitempty" form:"level"`
	Type           string    `json:"type,omitempty" form:"type"`
	UserID         string    `json:"user_id,omitempty" form:"user_id"`
	OrganizationID string    `json:"organization_id,omitempty" form:"organization_id"`
	Method         string    `json:"method,omitempty" form:"method"`
	Path           string    `json:"path,omitempty" form:"path"`
	Search         string    `json:"search,omitempty" form:"search"`
	StatusCode     int       `json:"status_code,omitempty" form:"status_code"`
	Limit          int       `json:"limit" form:"limit"`
	Offset         int       `json:"offset" form:"offset"`
}

// QueryRequest represents a log query with filters (for logging service)
type QueryRequest struct {
	StartTime  *time.Time             `json:"start_time,omitempty"`
	EndTime    *time.Time             `json:"end_time,omitempty"`
	Filters    map[string]interface{} `json:"filters,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	EntityType string                 `json:"entity_type,omitempty"`
	EntityID   string                 `json:"entity_id,omitempty"`
	Logger     string                 `json:"logger,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
	OrgID      string                 `json:"org_id,omitempty"`
	Message    string                 `json:"message,omitempty"`
	OrderBy    string                 `json:"order_by,omitempty"`
	Level      string                 `json:"level,omitempty"`
	Limit      int                    `json:"limit,omitempty"`
	Offset     int                    `json:"offset,omitempty"`
}

// StorageBackend defines the interface for log storage backends
type StorageBackend interface {
	// Initialize sets up the storage backend with configuration
	Initialize(ctx context.Context, config map[string]interface{}) error

	// Store saves a log entry to the backend
	Store(ctx context.Context, entry *LogEntry) error

	// StoreBatch saves multiple log entries efficiently
	StoreBatch(ctx context.Context, entries []*LogEntry) error

	// Query retrieves log entries based on filters
	Query(ctx context.Context, query *QueryRequest) ([]*LogEntry, error)

	// Close cleanly shuts down the storage backend
	Close() error

	// HealthCheck verifies the backend is operational
	HealthCheck(ctx context.Context) error

	// GetCapabilities returns what features this backend supports
	GetCapabilities() BackendCapabilities
}

// BackendCapabilities describes what features a storage backend supports
type BackendCapabilities struct {
	SupportsQuery      bool `json:"supports_query"`
	SupportsStreaming  bool `json:"supports_streaming"`
	SupportsRetention  bool `json:"supports_retention"`
	SupportsBatchWrite bool `json:"supports_batch_write"`
	SupportsMetrics    bool `json:"supports_metrics"`
}

// Log level constants
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
	LogLevelFatal = "fatal"
)

// Log type constants
const (
	LogTypeRequest = "request"
	LogTypeAudit   = "audit"
	LogTypeSystem  = "system"
	LogTypeError   = "error"
	LogTypeMetric  = "metric"
)

// Rate limit type constants
const (
	RateLimitTypeUser         = "user"
	RateLimitTypeOrganization = "organization"
	RateLimitTypeEndpoint     = "endpoint"
	RateLimitTypeGlobal       = "global"
	RateLimitTypeAPIKey       = "api_key"
)

// Rate limit algorithm constants
const (
	RateLimitAlgorithmSlidingWindow = "sliding_window"
	RateLimitAlgorithmFixedWindow   = "fixed_window"
	RateLimitAlgorithmTokenBucket   = "token_bucket"
	RateLimitAlgorithmLeakyBucket   = "leaky_bucket"
)

// Metric type constants
const (
	MetricTypeCounter   = "counter"
	MetricTypeGauge     = "gauge"
	MetricTypeHistogram = "histogram"
	MetricTypeSummary   = "summary"
)

// Audit action constants
const (
	AuditActionCreate = "create"
	AuditActionRead   = "read"
	AuditActionUpdate = "update"
	AuditActionDelete = "delete"
	AuditActionLogin  = "login"
	AuditActionLogout = "logout"
)
