package types

import (
	"time"
)

// ConfigurationExport represents a complete configuration export
type ConfigurationExport struct {
	Servers        []interface{}  `json:"servers,omitempty"`
	VirtualServers []interface{}  `json:"virtual_servers,omitempty"`
	Tools          []interface{}  `json:"tools,omitempty"`
	Prompts        []interface{}  `json:"prompts,omitempty"`
	Resources      []interface{}  `json:"resources,omitempty"`
	Policies       []interface{}  `json:"policies,omitempty"`
	RateLimits     []interface{}  `json:"rate_limits,omitempty"`
	Metadata       ExportMetadata `json:"metadata"`
}

// ExportMetadata contains information about the export
type ExportMetadata struct {
	Timestamp     time.Time         `json:"timestamp"`
	Tags          map[string]string `json:"tags,omitempty"`
	ExportID      string            `json:"export_id"`
	Version       string            `json:"version"`
	Gateway       string            `json:"gateway"`
	Organization  string            `json:"organization"`
	ExportedBy    string            `json:"exported_by"`
	EntityTypes   []string          `json:"entity_types"`
	Filters       ExportFilters     `json:"filters"`
	TotalEntities int               `json:"total_entities"`
}

// ExportFilters contains filtering options applied during export
type ExportFilters struct {
	DateRange           *DateRange `json:"date_range,omitempty"`
	Tags                []string   `json:"tags,omitempty"`
	IncludeInactive     bool       `json:"include_inactive"`
	IncludeDependencies bool       `json:"include_dependencies"`
}

// DateRange represents a time range filter
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ExportRequest represents a configuration export request
type ExportRequest struct {
	EntityTypes         []string      `json:"entity_types" binding:"required,min=1"`
	Tags                []string      `json:"tags,omitempty"`
	Filters             ExportFilters `json:"filters,omitempty"`
	IncludeInactive     bool          `json:"include_inactive"`
	IncludeDependencies bool          `json:"include_dependencies"`
}

// ImportRequest represents a configuration import request
type ImportRequest struct {
	ConflictStrategy ConflictStrategy    `json:"conflict_strategy"`
	RekeySecret      string              `json:"rekey_secret,omitempty"`
	Options          ImportOptions       `json:"options,omitempty"`
	ConfigData       ConfigurationExport `json:"config_data"`
	DryRun           bool                `json:"dry_run"`
}

// ConflictStrategy defines how to handle conflicts during import
type ConflictStrategy string

const (
	ConflictStrategyUpdate ConflictStrategy = "update" // Update existing items
	ConflictStrategySkip   ConflictStrategy = "skip"   // Skip conflicting items
	ConflictStrategyRename ConflictStrategy = "rename" // Rename conflicting items
	ConflictStrategyFail   ConflictStrategy = "fail"   // Fail on conflicts
)

// ImportOptions contains additional import configuration
type ImportOptions struct {
	IgnoreEntityTypes  []string `json:"ignore_entity_types,omitempty"`
	PreserveIDs        bool     `json:"preserve_ids"`
	UpdateTimestamps   bool     `json:"update_timestamps"`
	ValidateReferences bool     `json:"validate_references"`
}

// ImportResult contains the results of an import operation
type ImportResult struct {
	StartedAt   time.Time          `json:"started_at"`
	CompletedAt *time.Time         `json:"completed_at,omitempty"`
	Duration    *time.Duration     `json:"duration,omitempty"`
	ImportID    string             `json:"import_id"`
	Status      ImportStatus       `json:"status"`
	Details     []ImportItemResult `json:"details,omitempty"`
	Errors      []ImportError      `json:"errors,omitempty"`
	Warnings    []ImportWarning    `json:"warnings,omitempty"`
	Summary     ImportSummary      `json:"summary"`
}

// ImportStatus represents the status of an import operation
type ImportStatus string

const (
	ImportStatusPending    ImportStatus = "pending"
	ImportStatusRunning    ImportStatus = "running"
	ImportStatusCompleted  ImportStatus = "completed"
	ImportStatusFailed     ImportStatus = "failed"
	ImportStatusPartial    ImportStatus = "partial"
	ImportStatusValidating ImportStatus = "validating"
)

// ImportSummary provides high-level statistics about an import
type ImportSummary struct {
	EntityCounts   map[string]ImportEntityCount `json:"entity_counts"`
	TotalItems     int                          `json:"total_items"`
	ProcessedItems int                          `json:"processed_items"`
	CreatedItems   int                          `json:"created_items"`
	UpdatedItems   int                          `json:"updated_items"`
	SkippedItems   int                          `json:"skipped_items"`
	FailedItems    int                          `json:"failed_items"`
}

// ImportEntityCount tracks import statistics per entity type
type ImportEntityCount struct {
	Total   int `json:"total"`
	Created int `json:"created"`
	Updated int `json:"updated"`
	Skipped int `json:"skipped"`
	Failed  int `json:"failed"`
}

// ImportItemResult contains the result of importing a single item
type ImportItemResult struct {
	EntityType string       `json:"entity_type"`
	EntityID   string       `json:"entity_id,omitempty"`
	EntityName string       `json:"entity_name"`
	Action     ImportAction `json:"action"`
	Status     string       `json:"status"`
	Message    string       `json:"message,omitempty"`
	Error      string       `json:"error,omitempty"`
	OldID      string       `json:"old_id,omitempty"`
	NewID      string       `json:"new_id,omitempty"`
}

// ImportAction represents the action taken during import
type ImportAction string

const (
	ImportActionCreate ImportAction = "create"
	ImportActionUpdate ImportAction = "update"
	ImportActionSkip   ImportAction = "skip"
	ImportActionRename ImportAction = "rename"
)

// ImportError represents an error that occurred during import
type ImportError struct {
	Details    map[string]interface{} `json:"details,omitempty"`
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	EntityType string                 `json:"entity_type,omitempty"`
	EntityName string                 `json:"entity_name,omitempty"`
}

// ImportWarning represents a warning that occurred during import
type ImportWarning struct {
	Details    map[string]interface{} `json:"details,omitempty"`
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	EntityType string                 `json:"entity_type,omitempty"`
	EntityName string                 `json:"entity_name,omitempty"`
}

// ImportHistory represents a historical import operation
type ImportHistory struct {
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	Duration         *time.Duration         `json:"duration" db:"duration"`
	Metadata         map[string]interface{} `json:"metadata" db:"metadata"`
	CompletedAt      *time.Time             `json:"completed_at" db:"completed_at"`
	Filename         string                 `json:"filename" db:"filename"`
	OrganizationID   string                 `json:"organization_id" db:"organization_id"`
	Status           ImportStatus           `json:"status" db:"status"`
	ConflictStrategy ConflictStrategy       `json:"conflict_strategy" db:"conflict_strategy"`
	ID               string                 `json:"id" db:"id"`
	ImportedBy       string                 `json:"imported_by" db:"imported_by"`
	EntityTypes      []string               `json:"entity_types" db:"entity_types"`
	Summary          ImportSummary          `json:"summary" db:"summary"`
	WarningCount     int                    `json:"warning_count" db:"warning_count"`
	ErrorCount       int                    `json:"error_count" db:"error_count"`
	DryRun           bool                   `json:"dry_run" db:"dry_run"`
}

// CreateImportHistoryRequest represents a request to create an import history record
type CreateImportHistoryRequest struct {
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	Filename         string                 `json:"filename" binding:"required"`
	ConflictStrategy ConflictStrategy       `json:"conflict_strategy" binding:"required"`
	EntityTypes      []string               `json:"entity_types" binding:"required"`
	DryRun           bool                   `json:"dry_run"`
}

// ConfigurationEntity represents a common interface for exportable/importable entities
type ConfigurationEntity interface {
	GetID() string
	GetEntityType() string
	GetName() string
	GetOrganizationID() string
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	IsActive() bool
	GetTags() []string
}

// ValidateImportRequest represents a request to validate an import file
type ValidateImportRequest struct {
	ConfigData ConfigurationExport `json:"config_data" binding:"required"`
}

// ValidationResult represents the result of import validation
type ValidationResult struct {
	EntityCounts       map[string]int      `json:"entity_counts"`
	Errors             []ValidationError   `json:"errors,omitempty"`
	Warnings           []ValidationWarning `json:"warnings,omitempty"`
	Conflicts          []ConflictItem      `json:"conflicts,omitempty"`
	Dependencies       []DependencyItem    `json:"dependencies,omitempty"`
	CompatibilityCheck CompatibilityResult `json:"compatibility_check"`
	Valid              bool                `json:"valid"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Details    map[string]interface{} `json:"details,omitempty"`
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	EntityType string                 `json:"entity_type,omitempty"`
	EntityName string                 `json:"entity_name,omitempty"`
	Field      string                 `json:"field,omitempty"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Details    map[string]interface{} `json:"details,omitempty"`
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	EntityType string                 `json:"entity_type,omitempty"`
	EntityName string                 `json:"entity_name,omitempty"`
}

// ConflictItem represents a potential conflict during import
type ConflictItem struct {
	ExistingValue interface{}            `json:"existing_value"`
	ImportValue   interface{}            `json:"import_value"`
	Details       map[string]interface{} `json:"details,omitempty"`
	EntityType    string                 `json:"entity_type"`
	EntityName    string                 `json:"entity_name"`
	ConflictType  string                 `json:"conflict_type"`
	Suggestion    string                 `json:"suggestion,omitempty"`
}

// DependencyItem represents a dependency relationship
type DependencyItem struct {
	EntityType    string `json:"entity_type"`
	EntityName    string `json:"entity_name"`
	DependsOnType string `json:"depends_on_type"`
	DependsOnName string `json:"depends_on_name"`
	DependsOnID   string `json:"depends_on_id"`
	Required      bool   `json:"required"`
	Missing       bool   `json:"missing"`
}

// CompatibilityResult represents compatibility check results
type CompatibilityResult struct {
	Version          string   `json:"version"`
	RequiredVersion  string   `json:"required_version,omitempty"`
	MissingFeatures  []string `json:"missing_features,omitempty"`
	UnsupportedTypes []string `json:"unsupported_types,omitempty"`
	Compatible       bool     `json:"compatible"`
}

// ImportHistoryQuery represents query parameters for import history
type ImportHistoryQuery struct {
	StartDate  *time.Time   `json:"start_date,omitempty" form:"start_date"`
	EndDate    *time.Time   `json:"end_date,omitempty" form:"end_date"`
	Status     ImportStatus `json:"status,omitempty" form:"status"`
	EntityType string       `json:"entity_type,omitempty" form:"entity_type"`
	ImportedBy string       `json:"imported_by,omitempty" form:"imported_by"`
	Limit      int          `json:"limit" form:"limit"`
	Offset     int          `json:"offset" form:"offset"`
}

// Entity type constants for configuration management
const (
	EntityTypeServer        = "server"
	EntityTypeVirtualServer = "virtual_server"
	EntityTypeTool          = "tool"
	EntityTypePrompt        = "prompt"
	EntityTypeResource      = "resource"
	EntityTypePolicy        = "policy"
	EntityTypeRateLimit     = "rate_limit"
	EntityTypeUser          = "user"
	EntityTypeAPIKey        = "api_key"
	EntityTypeRoot          = "root"
)

// All supported entity types
var SupportedEntityTypes = []string{
	EntityTypeServer,
	EntityTypeVirtualServer,
	EntityTypeTool,
	EntityTypePrompt,
	EntityTypeResource,
	EntityTypePolicy,
	EntityTypeRateLimit,
}
