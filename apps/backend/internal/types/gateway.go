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

