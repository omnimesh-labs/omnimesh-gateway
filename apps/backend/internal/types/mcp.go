package types

import "time"

// MCPMessage represents a Model Context Protocol message
type MCPMessage struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Method  string                 `json:"method,omitempty"`
	Params  map[string]interface{} `json:"params,omitempty"`
	Result  interface{}            `json:"result,omitempty"`
	Error   *MCPError              `json:"error,omitempty"`
	Version string                 `json:"version"`
}

// MCPError represents an MCP protocol error
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCPCapability represents server capabilities
type MCPCapability struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Methods     []string               `json:"methods"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
}

// MCPTool represents an MCP tool
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
	Examples    []MCPToolExample       `json:"examples,omitempty"`
}

// MCPToolExample represents an example usage of an MCP tool
type MCPToolExample struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Input       map[string]interface{} `json:"input"`
	Output      interface{}            `json:"output"`
}

// MCPResource represents an MCP resource
type MCPResource struct {
	URI         string            `json:"uri"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	MimeType    string            `json:"mime_type"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// MCPPrompt represents an MCP prompt template
type MCPPrompt struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Arguments   []MCPPromptArgument `json:"arguments"`
	Template    string              `json:"template"`
	Examples    []MCPPromptExample  `json:"examples,omitempty"`
}

// MCPPromptArgument represents a prompt argument
type MCPPromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Type        string `json:"type"`
}

// MCPPromptExample represents an example prompt usage
type MCPPromptExample struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Arguments   map[string]interface{} `json:"arguments"`
	Output      string                 `json:"output"`
}

// MCPServerInfo represents MCP server information
type MCPServerInfo struct {
	Name         string          `json:"name"`
	Version      string          `json:"version"`
	Description  string          `json:"description"`
	Capabilities []MCPCapability `json:"capabilities"`
	Tools        []MCPTool       `json:"tools,omitempty"`
	Resources    []MCPResource   `json:"resources,omitempty"`
	Prompts      []MCPPrompt     `json:"prompts,omitempty"`
}

// MCPSession represents an active MCP session
type MCPSession struct {
	ID             string                 `json:"id" db:"id"`
	UserID         string                 `json:"user_id" db:"user_id"`
	OrganizationID string                 `json:"organization_id" db:"organization_id"`
	ServerID       string                 `json:"server_id" db:"server_id"`
	Status         string                 `json:"status" db:"status"` // "active", "inactive", "error"
	StartedAt      time.Time              `json:"started_at" db:"started_at"`
	LastActivity   time.Time              `json:"last_activity" db:"last_activity"`
	EndedAt        *time.Time             `json:"ended_at,omitempty" db:"ended_at"`
	Metadata       map[string]interface{} `json:"metadata" db:"metadata"`
}

// MCPRequest represents an MCP request
type MCPRequest struct {
	ID        string                 `json:"id" db:"id"`
	SessionID string                 `json:"session_id" db:"session_id"`
	Method    string                 `json:"method" db:"method"`
	Params    map[string]interface{} `json:"params" db:"params"`
	Timestamp time.Time              `json:"timestamp" db:"timestamp"`
}

// MCPResponse represents an MCP response
type MCPResponse struct {
	ID        string                 `json:"id" db:"id"`
	RequestID string                 `json:"request_id" db:"request_id"`
	Result    map[string]interface{} `json:"result,omitempty" db:"result"`
	Error     *MCPError              `json:"error,omitempty" db:"error"`
	Latency   time.Duration          `json:"latency" db:"latency"`
	Timestamp time.Time              `json:"timestamp" db:"timestamp"`
}

// MCP message types
const (
	MCPMessageTypeRequest      = "request"
	MCPMessageTypeResponse     = "response"
	MCPMessageTypeNotification = "notification"
	MCPMessageTypeError        = "error"
)

// MCP methods
const (
	MCPMethodInitialize    = "initialize"
	MCPMethodListTools     = "tools/list"
	MCPMethodCallTool      = "tools/call"
	MCPMethodListResources = "resources/list"
	MCPMethodReadResource  = "resources/read"
	MCPMethodListPrompts   = "prompts/list"
	MCPMethodGetPrompt     = "prompts/get"
)

// MCP session status
const (
	MCPSessionStatusActive   = "active"
	MCPSessionStatusInactive = "inactive"
	MCPSessionStatusError    = "error"
)

// MCP error codes
const (
	MCPErrorCodeParseError     = -32700
	MCPErrorCodeInvalidRequest = -32600
	MCPErrorCodeMethodNotFound = -32601
	MCPErrorCodeInvalidParams  = -32602
	MCPErrorCodeInternalError  = -32603
	MCPErrorCodeServerError    = -32000
	MCPErrorCodeTimeout        = -32001
	MCPErrorCodeCancelled      = -32002
)

// Standard MCP capabilities
const (
	MCPCapabilityTools     = "tools"
	MCPCapabilityResources = "resources"
	MCPCapabilityPrompts   = "prompts"
	MCPCapabilityLogging   = "logging"
	MCPCapabilitySampling  = "sampling"
)

// LoggingMetrics represents logging service metrics
type LoggingMetrics struct {
	BufferSize      int    `json:"buffer_size"`
	SubscriberCount int    `json:"subscriber_count"`
	CurrentLevel    string `json:"current_level"`
	BackendType     string `json:"backend_type"`
	AsyncMode       bool   `json:"async_mode"`
}
