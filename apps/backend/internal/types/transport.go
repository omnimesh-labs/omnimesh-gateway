package types

import (
	"context"
	"net/http"
	"time"
)

// TransportType represents the type of transport protocol
type TransportType string

const (
	TransportTypeHTTP       TransportType = "HTTP"
	TransportTypeSSE        TransportType = "SSE"
	TransportTypeWebSocket  TransportType = "WEBSOCKET"
	TransportTypeStreamable TransportType = "STREAMABLE"
	TransportTypeSTDIO      TransportType = "STDIO"
)

// Transport interface defines the contract for all transport implementations
type Transport interface {
	// Connect establishes a connection for this transport
	Connect(ctx context.Context) error

	// Disconnect closes the connection for this transport
	Disconnect(ctx context.Context) error

	// SendMessage sends a message via this transport
	SendMessage(ctx context.Context, message interface{}) error

	// ReceiveMessage receives a message via this transport
	ReceiveMessage(ctx context.Context) (interface{}, error)

	// IsConnected returns whether the transport is currently connected
	IsConnected() bool

	// GetTransportType returns the type of this transport
	GetTransportType() TransportType

	// GetSessionID returns the session ID for stateful transports
	GetSessionID() string

	// SetSessionID sets the session ID for stateful transports
	SetSessionID(sessionID string)
}

// TransportSession represents a session for stateful transports
type TransportSession struct {
	ID             string                 `json:"id" db:"id"`
	UserID         string                 `json:"user_id" db:"user_id"`
	OrganizationID string                 `json:"organization_id" db:"organization_id"`
	ServerID       string                 `json:"server_id,omitempty" db:"server_id"`
	TransportType  TransportType          `json:"transport_type" db:"transport_type"`
	Status         string                 `json:"status" db:"status"` // "active", "inactive", "closed"
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	LastActivity   time.Time              `json:"last_activity" db:"last_activity"`
	ExpiresAt      time.Time              `json:"expires_at" db:"expires_at"`
	Metadata       map[string]interface{} `json:"metadata" db:"metadata"`
	EventStore     []TransportEvent       `json:"event_store,omitempty" db:"-"`
}

// TransportEvent represents an event in a transport session
type TransportEvent struct {
	ID        string                 `json:"id"`
	SessionID string                 `json:"session_id"`
	Type      string                 `json:"type"` // "message", "connect", "disconnect", "error"
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// TransportConfig represents configuration for transport layer
type TransportConfig struct {
	EnabledTransports []TransportType `yaml:"enabled_transports" json:"enabled_transports"`
	SSEKeepAlive      time.Duration   `yaml:"sse_keep_alive" json:"sse_keep_alive"`
	WebSocketTimeout  time.Duration   `yaml:"websocket_timeout" json:"websocket_timeout"`
	SessionTimeout    time.Duration   `yaml:"session_timeout" json:"session_timeout"`
	MaxConnections    int             `yaml:"max_connections" json:"max_connections"`
	BufferSize        int             `yaml:"buffer_size" json:"buffer_size"`

	// Streamable HTTP specific settings
	StreamableStateful bool `yaml:"streamable_stateful" json:"streamable_stateful"`

	// STDIO specific settings
	STDIOTimeout time.Duration `yaml:"stdio_timeout" json:"stdio_timeout"`
}

// TransportRequest represents a request through any transport
type TransportRequest struct {
	ID         string                 `json:"id"`
	SessionID  string                 `json:"session_id,omitempty"`
	Transport  TransportType          `json:"transport"`
	Method     string                 `json:"method"`
	Path       string                 `json:"path"`
	Headers    map[string]string      `json:"headers,omitempty"`
	Body       interface{}            `json:"body,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// TransportResponse represents a response through any transport
type TransportResponse struct {
	ID        string            `json:"id"`
	RequestID string            `json:"request_id"`
	SessionID string            `json:"session_id,omitempty"`
	Transport TransportType     `json:"transport"`
	Status    int               `json:"status"`
	Headers   map[string]string `json:"headers,omitempty"`
	Body      interface{}       `json:"body,omitempty"`
	Error     *MCPError         `json:"error,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Latency   time.Duration     `json:"latency"`
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type      string      `json:"type"` // "text", "binary", "ping", "pong", "close"
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// SSEEvent represents a Server-Sent Event
type SSEEvent struct {
	ID        string      `json:"id,omitempty"`
	Event     string      `json:"event,omitempty"`
	Data      interface{} `json:"data"`
	Retry     int         `json:"retry,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// StreamableHTTPRequest represents a request for Streamable HTTP transport
type StreamableHTTPRequest struct {
	Method     string                 `json:"method"`
	Path       string                 `json:"path"`
	Headers    map[string]string      `json:"headers,omitempty"`
	Body       interface{}            `json:"body,omitempty"`
	Stateful   bool                   `json:"stateful"`
	SessionID  string                 `json:"session_id,omitempty"`
	StreamMode string                 `json:"stream_mode"` // "json", "sse"
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// STDIOCommand represents a command for STDIO transport
type STDIOCommand struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
	Dir     string            `json:"dir,omitempty"`
	Timeout time.Duration     `json:"timeout,omitempty"`
}

// PathRewriteRule represents a path rewriting rule
type PathRewriteRule struct {
	Pattern     string            `json:"pattern"`
	Replacement string            `json:"replacement"`
	Headers     map[string]string `json:"headers,omitempty"`
	Context     map[string]string `json:"context,omitempty"`
}

// TransportContext holds context information for transport operations
type TransportContext struct {
	Request        *http.Request
	SessionID      string
	UserID         string
	OrganizationID string
	ServerID       string
	Transport      TransportType
	Metadata       map[string]interface{}
}

// Transport session status constants
const (
	TransportSessionStatusActive   = "active"
	TransportSessionStatusInactive = "inactive"
	TransportSessionStatusClosed   = "closed"
	TransportSessionStatusError    = "error"
)

// Transport event type constants
const (
	TransportEventTypeMessage    = "message"
	TransportEventTypeConnect    = "connect"
	TransportEventTypeDisconnect = "disconnect"
	TransportEventTypeError      = "error"
	TransportEventTypePing       = "ping"
	TransportEventTypePong       = "pong"
)

// Streamable HTTP mode constants
const (
	StreamableModeJSON = "json"
	StreamableModeSSE  = "sse"
)

// WebSocket message type constants
const (
	WebSocketMessageTypeText   = "text"
	WebSocketMessageTypeBinary = "binary"
	WebSocketMessageTypePing   = "ping"
	WebSocketMessageTypePong   = "pong"
	WebSocketMessageTypeClose  = "close"
)

// Default transport configuration values
const (
	DefaultSSEKeepAlive     = 30 * time.Second
	DefaultWebSocketTimeout = 60 * time.Second
	DefaultSessionTimeout   = 24 * time.Hour
	DefaultMaxConnections   = 1000
	DefaultBufferSize       = 1024
	DefaultSTDIOTimeout     = 30 * time.Second
)
