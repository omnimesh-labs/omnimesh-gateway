package types

import "time"

// MCPServer represents an MCP server registration
type MCPServer struct {
	UpdatedAt      time.Time         `json:"updated_at" db:"updated_at"`
	CreatedAt      time.Time         `json:"created_at" db:"created_at"`
	Metadata       map[string]string `json:"metadata" db:"metadata"`
	URL            string            `json:"url" db:"url"`
	Name           string            `json:"name" db:"name"`
	Protocol       string            `json:"protocol" db:"protocol"`
	Version        string            `json:"version" db:"version"`
	Status         string            `json:"status" db:"status"`
	Description    string            `json:"description" db:"description"`
	HealthCheckURL string            `json:"health_check_url" db:"health_check_url"`
	WorkingDir     string            `json:"working_dir,omitempty" db:"working_dir"`
	Command        string            `json:"command,omitempty" db:"command"`
	OrganizationID string            `json:"organization_id" db:"organization_id"`
	ID             string            `json:"id" db:"id"`
	Args           []string          `json:"args,omitempty" db:"args"`
	Environment    []string          `json:"environment,omitempty" db:"environment"`
	MaxRetries     int               `json:"max_retries" db:"max_retries"`
	Timeout        time.Duration     `json:"timeout" db:"timeout"`
	IsActive       bool              `json:"is_active" db:"is_active"`
}

// HealthCheck represents a health check result
type HealthCheck struct {
	CheckedAt time.Time `json:"checked_at" db:"checked_at"`
	ID        string    `json:"id" db:"id"`
	ServerID  string    `json:"server_id" db:"server_id"`
	Status    string    `json:"status" db:"status"`
	Response  string    `json:"response" db:"response"`
	Error     string    `json:"error,omitempty" db:"error"`
	Latency   int64     `json:"latency" db:"latency"`
}

// ServerStats represents basic server statistics
type ServerStats struct {
	LastRequest     time.Time `json:"last_request"`
	ServerID        string    `json:"server_id"`
	TotalRequests   int64     `json:"total_requests"`
	SuccessRequests int64     `json:"success_requests"`
	ErrorRequests   int64     `json:"error_requests"`
	AvgLatency      float64   `json:"avg_latency"`
}

// CreateMCPServerRequest represents an MCP server registration request
type CreateMCPServerRequest struct {
	Metadata       map[string]string `json:"metadata"`
	HealthCheckURL string            `json:"health_check_url" binding:"omitempty,url"`
	URL            string            `json:"url" binding:"omitempty,url"`
	Protocol       string            `json:"protocol" binding:"required"`
	Version        string            `json:"version"`
	Description    string            `json:"description"`
	Name           string            `json:"name" binding:"required,min=2"`
	Command        string            `json:"command,omitempty"`
	WorkingDir     string            `json:"working_dir,omitempty"`
	Args           []string          `json:"args,omitempty"`
	Environment    []string          `json:"environment,omitempty"`
	Timeout        time.Duration     `json:"timeout"`
	MaxRetries     int               `json:"max_retries"`
}

// UpdateMCPServerRequest represents an MCP server update request
type UpdateMCPServerRequest struct {
	Metadata       map[string]string `json:"metadata,omitempty"`
	IsActive       *bool             `json:"is_active,omitempty"`
	Protocol       string            `json:"protocol,omitempty"`
	Name           string            `json:"name,omitempty" binding:"omitempty,min=2"`
	Version        string            `json:"version,omitempty"`
	URL            string            `json:"url,omitempty" binding:"omitempty,url"`
	HealthCheckURL string            `json:"health_check_url,omitempty" binding:"omitempty,url"`
	Description    string            `json:"description,omitempty"`
	Command        string            `json:"command,omitempty"`
	WorkingDir     string            `json:"working_dir,omitempty"`
	Args           []string          `json:"args,omitempty"`
	Environment    []string          `json:"environment,omitempty"`
	Timeout        time.Duration     `json:"timeout,omitempty"`
	MaxRetries     int               `json:"max_retries,omitempty"`
}

// ServerStatus constants
const (
	ServerStatusActive      = "active"
	ServerStatusInactive    = "inactive"
	ServerStatusUnhealthy   = "unhealthy"
	ServerStatusMaintenance = "maintenance"
)

// Protocol constants
const (
	ProtocolHTTP      = "http"
	ProtocolHTTPS     = "https"
	ProtocolWebSocket = "websocket"
	ProtocolSSE       = "sse"
	ProtocolStdio     = "stdio"
)

// Health check status constants
const (
	HealthStatusHealthy   = "healthy"
	HealthStatusUnhealthy = "unhealthy"
	HealthStatusTimeout   = "timeout"
	HealthStatusError     = "error"
)

// MCPProcess represents a running MCP server process
type MCPProcess struct {
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	ExitCode  *int       `json:"exit_code,omitempty"`
	Command   string     `json:"command"`
	Status    string     `json:"status"`
	Error     string     `json:"error,omitempty"`
	Args      []string   `json:"args"`
	PID       int        `json:"pid"`
}

// MCPConfig represents MCP configuration
type MCPConfig struct {
	LogLevel              string        `json:"log_level"`
	MaxConcurrentSessions int           `json:"max_concurrent_sessions"`
	SessionTimeout        time.Duration `json:"session_timeout"`
	ProcessTimeout        time.Duration `json:"process_timeout"`
	BufferSize            int           `json:"buffer_size"`
	EnableLogging         bool          `json:"enable_logging"`
}

// MCP session status constants
const (
	SessionStatusInitializing = "initializing"
	SessionStatusActive       = "active"
	SessionStatusClosed       = "closed"
	SessionStatusError        = "error"
)

// MCP process status constants
const (
	ProcessStatusStarting = "starting"
	ProcessStatusRunning  = "running"
	ProcessStatusStopped  = "stopped"
	ProcessStatusError    = "error"
)

// FilteringMetrics represents filtering system metrics
type FilteringMetrics struct {
	LastReset        time.Time              `json:"last_reset"`
	FilterStats      map[string]*FilterStat `json:"filter_stats"`
	ViolationsByType map[string]int64       `json:"violations_by_type"`
	TotalRequests    int64                  `json:"total_requests"`
	TotalBlocked     int64                  `json:"total_blocked"`
	TotalModified    int64                  `json:"total_modified"`
	ProcessingTime   time.Duration          `json:"processing_time"`
}

// FilterStat represents statistics for a specific filter
type FilterStat struct {
	LastActive        time.Time     `json:"last_active"`
	Name              string        `json:"name"`
	Type              string        `json:"type"`
	RequestsProcessed int64         `json:"requests_processed"`
	Violations        int64         `json:"violations"`
	Blocks            int64         `json:"blocks"`
	Modifications     int64         `json:"modifications"`
	AverageLatency    time.Duration `json:"average_latency"`
	Errors            int64         `json:"errors"`
}

// Resource represents a globally available MCP resource
type Resource struct {
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	SizeBytes         *int64                 `json:"size_bytes,omitempty" db:"size_bytes"`
	CreatedBy         *string                `json:"created_by,omitempty" db:"created_by"`
	Metadata          map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	AccessPermissions map[string]interface{} `json:"access_permissions,omitempty" db:"access_permissions"`
	MimeType          string                 `json:"mime_type,omitempty" db:"mime_type"`
	ID                string                 `json:"id" db:"id"`
	URI               string                 `json:"uri" db:"uri"`
	ResourceType      string                 `json:"resource_type" db:"resource_type"`
	Description       string                 `json:"description,omitempty" db:"description"`
	Name              string                 `json:"name" db:"name"`
	OrganizationID    string                 `json:"organization_id" db:"organization_id"`
	Tags              []string               `json:"tags,omitempty" db:"tags"`
	IsActive          bool                   `json:"is_active" db:"is_active"`
}

// Prompt represents a globally available MCP prompt
type Prompt struct {
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedBy      *string                `json:"created_by,omitempty" db:"created_by"`
	Description    string                 `json:"description,omitempty" db:"description"`
	Category       string                 `json:"category" db:"category"`
	PromptTemplate string                 `json:"prompt_template" db:"prompt_template"`
	ID             string                 `json:"id" db:"id"`
	Name           string                 `json:"name" db:"name"`
	OrganizationID string                 `json:"organization_id" db:"organization_id"`
	Parameters     []interface{}          `json:"parameters,omitempty" db:"parameters"`
	Tags           []string               `json:"tags,omitempty" db:"tags"`
	UsageCount     int64                  `json:"usage_count" db:"usage_count"`
	IsActive       bool                   `json:"is_active" db:"is_active"`
}

// CreateResourceRequest represents a resource creation request
type CreateResourceRequest struct {
	Name              string                 `json:"name" binding:"required,min=2"`
	Description       string                 `json:"description"`
	ResourceType      string                 `json:"resource_type" binding:"required"`
	URI               string                 `json:"uri" binding:"required"`
	MimeType          string                 `json:"mime_type"`
	SizeBytes         *int64                 `json:"size_bytes"`
	AccessPermissions map[string]interface{} `json:"access_permissions"`
	Metadata          map[string]interface{} `json:"metadata"`
	Tags              []string               `json:"tags"`
}

// UpdateResourceRequest represents an MCP resource update request
type UpdateResourceRequest struct {
	Name              string                 `json:"name,omitempty" binding:"omitempty,min=2"`
	Description       string                 `json:"description,omitempty"`
	ResourceType      string                 `json:"resource_type,omitempty"`
	URI               string                 `json:"uri,omitempty"`
	MimeType          string                 `json:"mime_type,omitempty"`
	SizeBytes         *int64                 `json:"size_bytes,omitempty"`
	AccessPermissions map[string]interface{} `json:"access_permissions,omitempty"`
	IsActive          *bool                  `json:"is_active,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	Tags              []string               `json:"tags,omitempty"`
}

// CreatePromptRequest represents an MCP prompt creation request
type CreatePromptRequest struct {
	Name           string                 `json:"name" binding:"required,min=2"`
	Description    string                 `json:"description"`
	PromptTemplate string                 `json:"prompt_template" binding:"required"`
	Parameters     []interface{}          `json:"parameters"`
	Category       string                 `json:"category" binding:"required"`
	Metadata       map[string]interface{} `json:"metadata"`
	Tags           []string               `json:"tags"`
}

// UpdatePromptRequest represents an MCP prompt update request
type UpdatePromptRequest struct {
	Name           string                 `json:"name,omitempty" binding:"omitempty,min=2"`
	Description    string                 `json:"description,omitempty"`
	PromptTemplate string                 `json:"prompt_template,omitempty"`
	Parameters     []interface{}          `json:"parameters,omitempty"`
	Category       string                 `json:"category,omitempty"`
	IsActive       *bool                  `json:"is_active,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
}

// Resource type constants
const (
	ResourceTypeFile     = "file"
	ResourceTypeURL      = "url"
	ResourceTypeDatabase = "database"
	ResourceTypeAPI      = "api"
	ResourceTypeMemory   = "memory"
	ResourceTypeCustom   = "custom"
)

// Prompt category constants
const (
	PromptCategoryGeneral     = "general"
	PromptCategoryCoding      = "coding"
	PromptCategoryAnalysis    = "analysis"
	PromptCategoryCreative    = "creative"
	PromptCategoryEducational = "educational"
	PromptCategoryBusiness    = "business"
	PromptCategoryCustom      = "custom"
)

// GlobalTool represents a globally available MCP tool
type GlobalTool struct {
	CreatedAt          time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at" db:"updated_at"`
	Metadata           map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	AccessPermissions  map[string]interface{} `json:"access_permissions,omitempty" db:"access_permissions"`
	CreatedBy          *string                `json:"created_by,omitempty" db:"created_by"`
	Schema             map[string]interface{} `json:"schema,omitempty" db:"schema"`
	ID                 string                 `json:"id" db:"id"`
	EndpointURL        string                 `json:"endpoint_url,omitempty" db:"endpoint_url"`
	FunctionName       string                 `json:"function_name" db:"function_name"`
	ImplementationType string                 `json:"implementation_type" db:"implementation_type"`
	OrganizationID     string                 `json:"organization_id" db:"organization_id"`
	Category           string                 `json:"category" db:"category"`
	Name               string                 `json:"name" db:"name"`
	Documentation      string                 `json:"documentation,omitempty" db:"documentation"`
	Description        string                 `json:"description,omitempty" db:"description"`
	Examples           []interface{}          `json:"examples,omitempty" db:"examples"`
	Tags               []string               `json:"tags,omitempty" db:"tags"`
	MaxRetries         int                    `json:"max_retries" db:"max_retries"`
	UsageCount         int64                  `json:"usage_count" db:"usage_count"`
	TimeoutSeconds     int                    `json:"timeout_seconds" db:"timeout_seconds"`
	IsPublic           bool                   `json:"is_public" db:"is_public"`
	IsActive           bool                   `json:"is_active" db:"is_active"`
}

// CreateGlobalToolRequest represents a tool creation request
type CreateGlobalToolRequest struct {
	AccessPermissions  map[string]interface{} `json:"access_permissions"`
	Schema             map[string]interface{} `json:"schema" binding:"required"`
	Metadata           map[string]interface{} `json:"metadata"`
	FunctionName       string                 `json:"function_name" binding:"required,min=2"`
	Category           string                 `json:"category" binding:"required"`
	ImplementationType string                 `json:"implementation_type"`
	Description        string                 `json:"description"`
	Documentation      string                 `json:"documentation"`
	EndpointURL        string                 `json:"endpoint_url"`
	Name               string                 `json:"name" binding:"required,min=2"`
	Tags               []string               `json:"tags"`
	Examples           []interface{}          `json:"examples"`
	TimeoutSeconds     int                    `json:"timeout_seconds"`
	MaxRetries         int                    `json:"max_retries"`
	IsPublic           bool                   `json:"is_public"`
}

// UpdateGlobalToolRequest represents a tool update request
type UpdateGlobalToolRequest struct {
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	IsActive           *bool                  `json:"is_active,omitempty"`
	IsPublic           *bool                  `json:"is_public,omitempty"`
	Schema             map[string]interface{} `json:"schema,omitempty"`
	TimeoutSeconds     *int                   `json:"timeout_seconds,omitempty"`
	MaxRetries         *int                   `json:"max_retries,omitempty"`
	AccessPermissions  map[string]interface{} `json:"access_permissions,omitempty"`
	Description        string                 `json:"description,omitempty"`
	ImplementationType string                 `json:"implementation_type,omitempty"`
	Category           string                 `json:"category,omitempty"`
	EndpointURL        string                 `json:"endpoint_url,omitempty"`
	FunctionName       string                 `json:"function_name,omitempty" binding:"omitempty,min=2"`
	Name               string                 `json:"name,omitempty" binding:"omitempty,min=2"`
	Documentation      string                 `json:"documentation,omitempty"`
	Tags               []string               `json:"tags,omitempty"`
	Examples           []interface{}          `json:"examples,omitempty"`
}

// Tool category constants
const (
	ToolCategoryGeneral = "general"
	ToolCategoryData    = "data"
	ToolCategoryFile    = "file"
	ToolCategoryWeb     = "web"
	ToolCategorySystem  = "system"
	ToolCategoryAI      = "ai"
	ToolCategoryDev     = "dev"
	ToolCategoryCustom  = "custom"
)

// Tool implementation type constants
const (
	ToolImplementationInternal = "internal"
	ToolImplementationExternal = "external"
	ToolImplementationWebhook  = "webhook"
	ToolImplementationScript   = "script"
)
