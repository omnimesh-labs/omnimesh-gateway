package types

import "time"

// MCPServer represents an MCP server registration
type MCPServer struct {
	ID             string            `json:"id" db:"id"`
	OrganizationID string            `json:"organization_id" db:"organization_id"`
	Name           string            `json:"name" db:"name"`
	Description    string            `json:"description" db:"description"`
	URL            string            `json:"url" db:"url"`
	Protocol       string            `json:"protocol" db:"protocol"` // "http", "websocket", "sse"
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
	URL            string            `json:"url" binding:"required,url"`
	Protocol       string            `json:"protocol" binding:"required"`
	Version        string            `json:"version"`
	Weight         int               `json:"weight"`
	Metadata       map[string]string `json:"metadata"`
	HealthCheckURL string            `json:"health_check_url" binding:"omitempty,url"`
	Timeout        time.Duration     `json:"timeout"`
	MaxRetries     int               `json:"max_retries"`
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
