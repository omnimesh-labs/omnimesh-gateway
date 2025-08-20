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
	Weight         int               `json:"weight" db:"weight"`
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

// ProxyRequest represents a proxied request
type ProxyRequest struct {
	ID             string            `json:"id"`
	UserID         string            `json:"user_id"`
	OrganizationID string            `json:"organization_id"`
	ServerID       string            `json:"server_id"`
	Method         string            `json:"method"`
	Path           string            `json:"path"`
	Headers        map[string]string `json:"headers"`
	Body           []byte            `json:"body,omitempty"`
	RemoteIP       string            `json:"remote_ip"`
	UserAgent      string            `json:"user_agent"`
	StartTime      time.Time         `json:"start_time"`
}

// ProxyResponse represents a proxied response
type ProxyResponse struct {
	RequestID  string            `json:"request_id"`
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body,omitempty"`
	Latency    time.Duration     `json:"latency"`
	Error      string            `json:"error,omitempty"`
	EndTime    time.Time         `json:"end_time"`
}

// LoadBalancerStats represents load balancer statistics
type LoadBalancerStats struct {
	ServerID        string    `json:"server_id"`
	ActiveRequests  int       `json:"active_requests"`
	TotalRequests   int64     `json:"total_requests"`
	SuccessRequests int64     `json:"success_requests"`
	ErrorRequests   int64     `json:"error_requests"`
	AvgLatency      float64   `json:"avg_latency"`
	LastRequest     time.Time `json:"last_request"`
}

// CircuitBreakerState represents circuit breaker state
type CircuitBreakerState struct {
	ServerID      string    `json:"server_id"`
	State         string    `json:"state"` // "closed", "open", "half_open"
	FailureCount  int       `json:"failure_count"`
	SuccessCount  int       `json:"success_count"`
	LastFailure   time.Time `json:"last_failure"`
	NextRetryTime time.Time `json:"next_retry_time"`
}

// CreateMCPServerRequest represents an MCP server registration request
type CreateMCPServerRequest struct {
	Name           string            `json:"name" binding:"required,min=2"`
	Description    string            `json:"description"`
	URL            string            `json:"url" binding:"omitempty,url"` // Optional for stdio servers
	Protocol       string            `json:"protocol" binding:"required"`
	Version        string            `json:"version"`
	Weight         int               `json:"weight"`
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
	Weight         int               `json:"weight,omitempty"`
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

// Circuit breaker state constants
const (
	CircuitBreakerClosed   = "closed"
	CircuitBreakerOpen     = "open"
	CircuitBreakerHalfOpen = "half_open"
)

// Load balancer algorithm constants
const (
	LoadBalancerRoundRobin = "round_robin"
	LoadBalancerLeastConn  = "least_conn"
	LoadBalancerWeighted   = "weighted"
	LoadBalancerRandom     = "random"
)

// MCPProxySession represents an active MCP proxy session
type MCPProxySession struct {
	ID             string                 `json:"id"`
	UserID         string                 `json:"user_id"`
	OrganizationID string                 `json:"organization_id"`
	ServerID       string                 `json:"server_id"`
	Server         *MCPServer             `json:"server"`
	Protocol       string                 `json:"protocol"`
	Status         string                 `json:"status"` // "initializing", "active", "closed", "error"
	StartedAt      time.Time              `json:"started_at"`
	LastActivity   time.Time              `json:"last_activity"`
	EndedAt        *time.Time             `json:"ended_at,omitempty"`
	Metadata       map[string]interface{} `json:"metadata"`

	// For stdio sessions
	Process    *MCPProcess `json:"process,omitempty"`
	StdinPipe  interface{} `json:"-"` // io.WriteCloser
	StdoutPipe interface{} `json:"-"` // io.ReadCloser
	StderrPipe interface{} `json:"-"` // io.ReadCloser

	// For HTTP sessions
	HTTPClient *interface{} `json:"-"` // *http.Client
}

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

// MCPProxyConfig represents proxy configuration
type MCPProxyConfig struct {
	MaxConcurrentSessions int           `json:"max_concurrent_sessions"`
	SessionTimeout        time.Duration `json:"session_timeout"`
	ProcessTimeout        time.Duration `json:"process_timeout"`
	BufferSize            int           `json:"buffer_size"`
	EnableLogging         bool          `json:"enable_logging"`
	LogLevel              string        `json:"log_level"`
}

// MCP proxy session status constants
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
