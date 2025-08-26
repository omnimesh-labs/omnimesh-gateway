package inspector

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SessionStatus represents the status of an inspector session
type SessionStatus string

const (
	SessionStatusInitializing SessionStatus = "initializing"
	SessionStatusConnected    SessionStatus = "connected"
	SessionStatusDisconnected SessionStatus = "disconnected"
	SessionStatusError        SessionStatus = "error"
)

// String returns the string representation of SessionStatus
func (s SessionStatus) String() string {
	return string(s)
}

// InspectorSession represents an active inspector session
type InspectorSession struct {
	ID           string                 `json:"id"`
	ServerID     string                 `json:"server_id"`
	UserID       string                 `json:"user_id"`
	OrgID        string                 `json:"org_id"`
	NamespaceID  string                 `json:"namespace_id"`
	Status       SessionStatus          `json:"status"`
	Capabilities map[string]interface{} `json:"capabilities"`
	CreatedAt    time.Time              `json:"created_at"`
	LastActivity time.Time              `json:"last_activity"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// InspectorRequest represents a request to execute on an MCP server
type InspectorRequest struct {
	ID        string                 `json:"id"`
	SessionID string                 `json:"session_id"`
	Method    string                 `json:"method"`
	Params    map[string]interface{} `json:"params"`
	Timestamp time.Time              `json:"timestamp"`
}

// InspectorResponse represents a response from an MCP server
type InspectorResponse struct {
	ID        string        `json:"id"`
	RequestID string        `json:"request_id"`
	Result    interface{}   `json:"result,omitempty"`
	Error     *MCPError     `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
}

// MCPError represents an MCP protocol error
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// String returns the string representation of MCPError
func (e *MCPError) String() string {
	return fmt.Sprintf("MCPError{Code: %d, Message: %s}", e.Code, e.Message)
}

// InspectorEvent represents an event in the inspector
type InspectorEvent struct {
	ID        string      `json:"id"`
	SessionID string      `json:"session_id"`
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// Tool represents an MCP tool
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// Resource represents an MCP resource
type Resource struct {
	URI         string                 `json:"uri"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	MimeType    string                 `json:"mimeType,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Prompt represents an MCP prompt
type Prompt struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Arguments   []PromptArgument       `json:"arguments,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PromptArgument represents an argument for a prompt
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// ListToolsResult represents the result of listing tools
type ListToolsResult struct {
	Tools      []Tool  `json:"tools"`
	NextCursor *string `json:"nextCursor,omitempty"`
}

// CallToolResult represents the result of calling a tool
type CallToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

// ToolContent represents content returned from a tool
type ToolContent struct {
	Type     string      `json:"type"`
	Text     string      `json:"text,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	MimeType string      `json:"mimeType,omitempty"`
}

// ListResourcesResult represents the result of listing resources
type ListResourcesResult struct {
	Resources  []Resource `json:"resources"`
	NextCursor *string    `json:"nextCursor,omitempty"`
}

// ReadResourceResult represents the result of reading a resource
type ReadResourceResult struct {
	Contents []ResourceContent `json:"contents"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ResourceContent represents content from a resource
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"`
}

// ListPromptsResult represents the result of listing prompts
type ListPromptsResult struct {
	Prompts    []Prompt `json:"prompts"`
	NextCursor *string  `json:"nextCursor,omitempty"`
}

// GetPromptResult represents the result of getting a prompt
type GetPromptResult struct {
	Messages []PromptMessage `json:"messages"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// PromptMessage represents a message in a prompt
type PromptMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// PingResult represents the result of a ping
type PingResult struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// ServerCapabilities represents MCP server capabilities
type ServerCapabilities struct {
	Tools      *ToolsCapability      `json:"tools,omitempty"`
	Resources  *ResourcesCapability  `json:"resources,omitempty"`
	Prompts    *PromptsCapability    `json:"prompts,omitempty"`
	Logging    *LoggingCapability    `json:"logging,omitempty"`
	Sampling   *SamplingCapability   `json:"sampling,omitempty"`
	Roots      *RootsCapability      `json:"roots,omitempty"`
}

// ToolsCapability represents tools capability
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapability represents resources capability
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// PromptsCapability represents prompts capability
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// LoggingCapability represents logging capability
type LoggingCapability struct {
	Level string `json:"level,omitempty"`
}

// SamplingCapability represents sampling capability
type SamplingCapability struct{}

// RootsCapability represents roots capability
type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// CreateSessionRequest represents a request to create an inspector session
type CreateSessionRequest struct {
	ServerID    string                 `json:"server_id" binding:"required"`
	NamespaceID string                 `json:"namespace_id"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ExecuteRequestBody represents the body of an execute request
type ExecuteRequestBody struct {
	Method string                 `json:"method" binding:"required"`
	Params map[string]interface{} `json:"params"`
}

// NewInspectorSession creates a new inspector session
func NewInspectorSession(serverID, userID, orgID, namespaceID string) *InspectorSession {
	return &InspectorSession{
		ID:           uuid.New().String(),
		ServerID:     serverID,
		UserID:       userID,
		OrgID:        orgID,
		NamespaceID:  namespaceID,
		Status:       SessionStatusInitializing,
		Capabilities: make(map[string]interface{}),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		Metadata:     make(map[string]interface{}),
	}
}

// MarshalJSON custom marshaller for InspectorResponse to handle duration
func (r InspectorResponse) MarshalJSON() ([]byte, error) {
	type Alias InspectorResponse
	return json.Marshal(&struct {
		Duration int64 `json:"duration"` // Duration in milliseconds
		*Alias
	}{
		Duration: r.Duration.Milliseconds(),
		Alias:    (*Alias)(&r),
	})
}