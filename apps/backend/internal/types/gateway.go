package types

import "time"

// MCPServer represents an MCP server registration
type MCPServer struct {
	ID             string            `json:"id" db:"id"`
	OrganizationID string            `json:"organization_id" db:"organization_id"`
	Name           string            `json:"name" db:"name"`
	Description    string            `json:"description" db:"description"`
	URL            string            `json:"url" db:"url"`
	Protocol       string            `json:"protocol" db:"protocol"` // "http", "websocket", "sse", "stdio"
	Version        string            `json:"version" db:"version"`
	Status         string            `json:"status" db:"status"` // "active", "inactive", "unhealthy"
	Metadata       map[string]string `json:"metadata" db:"metadata"`
	HealthCheckURL string            `json:"health_check_url" db:"health_check_url"`
	Timeout        time.Duration     `json:"timeout" db:"timeout"`
	MaxRetries     int               `json:"max_retries" db:"max_retries"`
	IsActive       bool              `json:"is_active" db:"is_active"`
	CreatedAt      time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at" db:"updated_at"`

	// For stdio/command-based servers
	Command     string   `json:"command,omitempty" db:"command"`
	Args        []string `json:"args,omitempty" db:"args"`
	Environment []string `json:"environment,omitempty" db:"environment"`
	WorkingDir  string   `json:"working_dir,omitempty" db:"working_dir"`
}

// HealthCheck represents a health check result
type HealthCheck struct {
	ID        string    `json:"id" db:"id"`
	ServerID  string    `json:"server_id" db:"server_id"`
	Status    string    `json:"status" db:"status"` // "healthy", "unhealthy", "timeout"
	Response  string    `json:"response" db:"response"`
	Latency   int64     `json:"latency" db:"latency"` // in milliseconds
	Error     string    `json:"error,omitempty" db:"error"`
	CheckedAt time.Time `json:"checked_at" db:"checked_at"`
}

// ServerStats represents basic server statistics
type ServerStats struct {
	ServerID        string    `json:"server_id"`
	TotalRequests   int64     `json:"total_requests"`
	SuccessRequests int64     `json:"success_requests"`
	ErrorRequests   int64     `json:"error_requests"`
	AvgLatency      float64   `json:"avg_latency"`
	LastRequest     time.Time `json:"last_request"`
}


// CreateMCPServerRequest represents an MCP server registration request
type CreateMCPServerRequest struct {
	Name           string            `json:"name" binding:"required,min=2"`
	Description    string            `json:"description"`
	URL            string            `json:"url" binding:"omitempty,url"` // Optional for stdio servers
	Protocol       string            `json:"protocol" binding:"required"`
	Version        string            `json:"version"`
	Metadata       map[string]string `json:"metadata"`
	HealthCheckURL string            `json:"health_check_url" binding:"omitempty,url"`
	Timeout        time.Duration     `json:"timeout"`
	MaxRetries     int               `json:"max_retries"`

	// For stdio/command-based servers
	Command     string   `json:"command,omitempty"`     // e.g., "npx"
	Args        []string `json:"args,omitempty"`        // e.g., ["-y", "@executeautomation/playwright-mcp-server"]
	Environment []string `json:"environment,omitempty"` // e.g., ["VAR=value"]
	WorkingDir  string   `json:"working_dir,omitempty"` // Working directory for the command
}

// UpdateMCPServerRequest represents an MCP server update request
type UpdateMCPServerRequest struct {
	Name           string            `json:"name,omitempty" binding:"omitempty,min=2"`
	Description    string            `json:"description,omitempty"`
	URL            string            `json:"url,omitempty" binding:"omitempty,url"`
	Protocol       string            `json:"protocol,omitempty"`
	Version        string            `json:"version,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	HealthCheckURL string            `json:"health_check_url,omitempty" binding:"omitempty,url"`
	Timeout        time.Duration     `json:"timeout,omitempty"`
	MaxRetries     int               `json:"max_retries,omitempty"`
	IsActive       *bool             `json:"is_active,omitempty"`

	// For stdio/command-based servers
	Command     string   `json:"command,omitempty"`
	Args        []string `json:"args,omitempty"`
	Environment []string `json:"environment,omitempty"`
	WorkingDir  string   `json:"working_dir,omitempty"`
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
	PID       int        `json:"pid"`
	Command   string     `json:"command"`
	Args      []string   `json:"args"`
	Status    string     `json:"status"` // "starting", "running", "stopped", "error"
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	ExitCode  *int       `json:"exit_code,omitempty"`
	Error     string     `json:"error,omitempty"`
}

// MCPConfig represents MCP configuration
type MCPConfig struct {
	MaxConcurrentSessions int           `json:"max_concurrent_sessions"`
	SessionTimeout        time.Duration `json:"session_timeout"`
	ProcessTimeout        time.Duration `json:"process_timeout"`
	BufferSize            int           `json:"buffer_size"`
	EnableLogging         bool          `json:"enable_logging"`
	LogLevel              string        `json:"log_level"`
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
	TotalRequests      int64                  `json:"total_requests"`
	TotalBlocked       int64                  `json:"total_blocked"`
	TotalModified      int64                  `json:"total_modified"`
	FilterStats        map[string]*FilterStat `json:"filter_stats"`
	ViolationsByType   map[string]int64       `json:"violations_by_type"`
	ProcessingTime     time.Duration          `json:"processing_time"`
	LastReset          time.Time              `json:"last_reset"`
}

// FilterStat represents statistics for a specific filter
type FilterStat struct {
	Name              string        `json:"name"`
	Type              string        `json:"type"`
	RequestsProcessed int64         `json:"requests_processed"`
	Violations        int64         `json:"violations"`
	Blocks            int64         `json:"blocks"`
	Modifications     int64         `json:"modifications"`
	AverageLatency    time.Duration `json:"average_latency"`
	LastActive        time.Time     `json:"last_active"`
	Errors            int64         `json:"errors"`
}

// Resource represents a globally available MCP resource
type Resource struct {
	ID                string                 `json:"id" db:"id"`
	OrganizationID    string                 `json:"organization_id" db:"organization_id"`
	Name              string                 `json:"name" db:"name"`
	Description       string                 `json:"description,omitempty" db:"description"`
	ResourceType      string                 `json:"resource_type" db:"resource_type"` // "file", "url", "database", "api", "memory", "custom"
	URI               string                 `json:"uri" db:"uri"`
	MimeType          string                 `json:"mime_type,omitempty" db:"mime_type"`
	SizeBytes         *int64                 `json:"size_bytes,omitempty" db:"size_bytes"`
	AccessPermissions map[string]interface{} `json:"access_permissions,omitempty" db:"access_permissions"`
	IsActive          bool                   `json:"is_active" db:"is_active"`
	Metadata          map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	Tags              []string               `json:"tags,omitempty" db:"tags"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
	CreatedBy         *string                `json:"created_by,omitempty" db:"created_by"`
}

// Prompt represents a globally available MCP prompt
type Prompt struct {
	ID             string                 `json:"id" db:"id"`
	OrganizationID string                 `json:"organization_id" db:"organization_id"`
	Name           string                 `json:"name" db:"name"`
	Description    string                 `json:"description,omitempty" db:"description"`
	PromptTemplate string                 `json:"prompt_template" db:"prompt_template"`
	Parameters     []interface{}          `json:"parameters,omitempty" db:"parameters"`
	Category       string                 `json:"category" db:"category"` // "general", "coding", "analysis", "creative", "educational", "business", "custom"
	UsageCount     int64                  `json:"usage_count" db:"usage_count"`
	IsActive       bool                   `json:"is_active" db:"is_active"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	Tags           []string               `json:"tags,omitempty" db:"tags"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
	CreatedBy      *string                `json:"created_by,omitempty" db:"created_by"`
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
	ID                 string                 `json:"id" db:"id"`
	OrganizationID     string                 `json:"organization_id" db:"organization_id"`
	Name               string                 `json:"name" db:"name"`
	Description        string                 `json:"description,omitempty" db:"description"`
	FunctionName       string                 `json:"function_name" db:"function_name"`
	Schema             map[string]interface{} `json:"schema,omitempty" db:"schema"`
	Category           string                 `json:"category" db:"category"` // "general", "data", "file", "web", "system", "ai", "dev", "custom"
	ImplementationType string                 `json:"implementation_type" db:"implementation_type"`
	EndpointURL        string                 `json:"endpoint_url,omitempty" db:"endpoint_url"`
	TimeoutSeconds     int                    `json:"timeout_seconds" db:"timeout_seconds"`
	MaxRetries         int                    `json:"max_retries" db:"max_retries"`
	UsageCount         int64                  `json:"usage_count" db:"usage_count"`
	AccessPermissions  map[string]interface{} `json:"access_permissions,omitempty" db:"access_permissions"`
	IsActive           bool                   `json:"is_active" db:"is_active"`
	IsPublic           bool                   `json:"is_public" db:"is_public"`
	Metadata           map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	Tags               []string               `json:"tags,omitempty" db:"tags"`
	Examples           []interface{}          `json:"examples,omitempty" db:"examples"`
	Documentation      string                 `json:"documentation,omitempty" db:"documentation"`
	CreatedAt          time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at" db:"updated_at"`
	CreatedBy          *string                `json:"created_by,omitempty" db:"created_by"`
}

// CreateGlobalToolRequest represents a tool creation request
type CreateGlobalToolRequest struct {
	Name               string                 `json:"name" binding:"required,min=2"`
	Description        string                 `json:"description"`
	FunctionName       string                 `json:"function_name" binding:"required,min=2"`
	Schema             map[string]interface{} `json:"schema" binding:"required"`
	Category           string                 `json:"category" binding:"required"`
	ImplementationType string                 `json:"implementation_type"`
	EndpointURL        string                 `json:"endpoint_url"`
	TimeoutSeconds     int                    `json:"timeout_seconds"`
	MaxRetries         int                    `json:"max_retries"`
	AccessPermissions  map[string]interface{} `json:"access_permissions"`
	IsPublic           bool                   `json:"is_public"`
	Metadata           map[string]interface{} `json:"metadata"`
	Tags               []string               `json:"tags"`
	Examples           []interface{}          `json:"examples"`
	Documentation      string                 `json:"documentation"`
}

// UpdateGlobalToolRequest represents a tool update request
type UpdateGlobalToolRequest struct {
	Name               string                 `json:"name,omitempty" binding:"omitempty,min=2"`
	Description        string                 `json:"description,omitempty"`
	FunctionName       string                 `json:"function_name,omitempty" binding:"omitempty,min=2"`
	Schema             map[string]interface{} `json:"schema,omitempty"`
	Category           string                 `json:"category,omitempty"`
	ImplementationType string                 `json:"implementation_type,omitempty"`
	EndpointURL        string                 `json:"endpoint_url,omitempty"`
	TimeoutSeconds     *int                   `json:"timeout_seconds,omitempty"`
	MaxRetries         *int                   `json:"max_retries,omitempty"`
	AccessPermissions  map[string]interface{} `json:"access_permissions,omitempty"`
	IsActive           *bool                  `json:"is_active,omitempty"`
	IsPublic           *bool                  `json:"is_public,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	Tags               []string               `json:"tags,omitempty"`
	Examples           []interface{}          `json:"examples,omitempty"`
	Documentation      string                 `json:"documentation,omitempty"`
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

