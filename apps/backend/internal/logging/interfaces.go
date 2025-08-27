package logging

import (
	"context"
	"time"

	"mcp-gateway/apps/backend/internal/types"
)

// LogLevel represents the severity of a log entry
type LogLevel string

const (
	LogLevelDebug     LogLevel = "debug"
	LogLevelInfo      LogLevel = "info"
	LogLevelNotice    LogLevel = "notice"
	LogLevelWarning   LogLevel = "warning"
	LogLevelError     LogLevel = "error"
	LogLevelCritical  LogLevel = "critical"
	LogLevelAlert     LogLevel = "alert"
	LogLevelEmergency LogLevel = "emergency"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Level       LogLevel               `json:"level"`
	Message     string                 `json:"message"`
	Logger      string                 `json:"logger,omitempty"`
	EntityType  string                 `json:"entity_type,omitempty"`
	EntityID    string                 `json:"entity_id,omitempty"`
	EntityName  string                 `json:"entity_name,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	OrgID       string                 `json:"org_id,omitempty"`
	StatusCode  int                    `json:"status_code,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Source      string                 `json:"source,omitempty"`
	Environment string                 `json:"environment,omitempty"`
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

// QueryRequest represents a log query with filters
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
	Level      LogLevel               `json:"level,omitempty"`
	Limit      int                    `json:"limit,omitempty"`
	Offset     int                    `json:"offset,omitempty"`
}

// LogSubscriber receives log events in real-time
type LogSubscriber interface {
	// OnLog is called when a new log entry is available
	OnLog(entry *LogEntry) error

	// GetID returns a unique identifier for this subscriber
	GetID() string

	// GetFilters returns the filters this subscriber is interested in
	GetFilters() *QueryRequest
}

// PluginFactory creates storage backend instances
type PluginFactory interface {
	// Create returns a new storage backend instance
	Create() StorageBackend

	// GetName returns the plugin name (e.g., "file", "aws-cloudwatch")
	GetName() string

	// GetDescription returns a human-readable description
	GetDescription() string

	// ValidateConfig validates the configuration for this plugin
	ValidateConfig(config map[string]interface{}) error
}

// PluginRegistry manages available storage plugins
type PluginRegistry interface {
	// Register adds a plugin factory to the registry
	Register(factory PluginFactory) error

	// Get returns a plugin factory by name
	Get(name string) (PluginFactory, error)

	// List returns all available plugin names
	List() []string

	// GetInfo returns information about a plugin
	GetInfo(name string) (*PluginInfo, error)
}

// PluginInfo contains metadata about a plugin
type PluginInfo struct {
	ConfigSchema map[string]interface{} `json:"config_schema,omitempty"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Version      string                 `json:"version"`
	Author       string                 `json:"author"`
	Capabilities BackendCapabilities    `json:"capabilities"`
}

// LoggingConfig represents the logging service configuration
type LoggingConfig struct {
	Config        map[string]interface{} `json:"config" yaml:"config"`
	Retention     *RetentionConfig       `json:"retention,omitempty" yaml:"retention,omitempty"`
	Level         LogLevel               `json:"level" yaml:"level"`
	Environment   string                 `json:"environment" yaml:"environment"`
	Backend       string                 `json:"backend" yaml:"backend"`
	BufferSize    int                    `json:"buffer_size" yaml:"buffer_size"`
	BatchSize     int                    `json:"batch_size" yaml:"batch_size"`
	FlushInterval time.Duration          `json:"flush_interval" yaml:"flush_interval"`
	Async         bool                   `json:"async" yaml:"async"`
}

// RetentionConfig defines log retention policies
type RetentionConfig struct {
	Policy    string `json:"policy" yaml:"policy"`
	Days      int    `json:"days" yaml:"days"`
	KeepCount int    `json:"keep_count" yaml:"keep_count"`
}

// LogService defines the main logging service interface
type LogService interface {
	// Log writes a log entry
	Log(ctx context.Context, entry *LogEntry) error

	// LogBatch writes multiple log entries
	LogBatch(ctx context.Context, entries []*LogEntry) error

	// Query searches for log entries
	Query(ctx context.Context, query *QueryRequest) ([]*LogEntry, error)

	// Subscribe adds a real-time log subscriber
	Subscribe(subscriber LogSubscriber) error

	// Unsubscribe removes a log subscriber
	Unsubscribe(subscriberID string) error

	// SetLevel changes the minimum log level
	SetLevel(level LogLevel) error

	// GetLevel returns the current log level
	GetLevel() LogLevel

	// HealthCheck verifies the service is operational
	HealthCheck(ctx context.Context) error

	// Close shuts down the service
	Close() error

	// GetMetrics returns logging metrics
	GetMetrics() (*types.LoggingMetrics, error)
}

// Helper functions for log levels
func (l LogLevel) IsValid() bool {
	switch l {
	case LogLevelDebug, LogLevelInfo, LogLevelNotice, LogLevelWarning,
		LogLevelError, LogLevelCritical, LogLevelAlert, LogLevelEmergency:
		return true
	default:
		return false
	}
}

func (l LogLevel) Priority() int {
	switch l {
	case LogLevelDebug:
		return 0
	case LogLevelInfo:
		return 1
	case LogLevelNotice:
		return 2
	case LogLevelWarning:
		return 3
	case LogLevelError:
		return 4
	case LogLevelCritical:
		return 5
	case LogLevelAlert:
		return 6
	case LogLevelEmergency:
		return 7
	default:
		return 1
	}
}

func (l LogLevel) String() string {
	return string(l)
}
