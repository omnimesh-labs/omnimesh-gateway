package types

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AgentType represents the type of AI agent
type AgentType string

const (
	AgentTypeCustom    AgentType = "custom"
	AgentTypeOpenAI    AgentType = "openai"
	AgentTypeAnthropic AgentType = "anthropic"
	AgentTypeGeneric   AgentType = "generic"
)

// AuthType represents the authentication method for an agent
type AuthType string

const (
	AuthTypeNone   AuthType = "none"
	AuthTypeAPIKey AuthType = "api_key"
	AuthTypeBearer AuthType = "bearer"
	AuthTypeOAuth  AuthType = "oauth"
)

// A2AHealthStatus represents the health status of an agent
type A2AHealthStatus string

const (
	A2AHealthStatusUnknown   A2AHealthStatus = "unknown"
	A2AHealthStatusHealthy   A2AHealthStatus = "healthy"
	A2AHealthStatusUnhealthy A2AHealthStatus = "unhealthy"
)

// A2AAgent represents an external AI agent in the database
type A2AAgent struct {
	UpdatedAt        time.Time              `db:"updated_at" json:"updated_at"`
	CreatedAt        time.Time              `db:"created_at" json:"created_at"`
	CapabilitiesData map[string]interface{} `db:"-" json:"capabilities"`
	LastHealthCheck  *time.Time             `db:"last_health_check" json:"last_health_check,omitempty"`
	MetadataData     map[string]interface{} `db:"-" json:"metadata"`
	ConfigData       map[string]interface{} `db:"-" json:"config"`
	AgentType        AgentType              `db:"agent_type" json:"agent_type"`
	HealthError      string                 `db:"health_error" json:"health_error,omitempty"`
	ProtocolVersion  string                 `db:"protocol_version" json:"protocol_version"`
	AuthType         AuthType               `db:"auth_type" json:"auth_type"`
	AuthValue        string                 `db:"auth_value" json:"-"`
	Name             string                 `db:"name" json:"name"`
	EndpointURL      string                 `db:"endpoint_url" json:"endpoint_url"`
	Description      string                 `db:"description" json:"description"`
	HealthStatus     A2AHealthStatus        `db:"health_status" json:"health_status"`
	Config           json.RawMessage        `db:"config" json:"-"`
	Capabilities     json.RawMessage        `db:"capabilities" json:"-"`
	Tags             []string               `db:"tags" json:"tags"`
	Metadata         json.RawMessage        `db:"metadata" json:"-"`
	ID               uuid.UUID              `db:"id" json:"id"`
	OrganizationID   uuid.UUID              `db:"organization_id" json:"organization_id"`
	IsActive         bool                   `db:"is_active" json:"is_active"`
}

// A2AAgentSpec represents an A2A agent for API operations
type A2AAgentSpec struct {
	UpdatedAt       time.Time              `json:"updated_at,omitempty"`
	CreatedAt       time.Time              `json:"created_at,omitempty"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	LastHealthCheck *time.Time             `json:"last_health_check,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
	Config          map[string]interface{} `json:"config"`
	AuthType        AuthType               `json:"auth_type"`
	ProtocolVersion string                 `json:"protocol_version"`
	ID              string                 `json:"id,omitempty"`
	AuthValue       string                 `json:"auth_value,omitempty"`
	AgentType       AgentType              `json:"agent_type"`
	EndpointURL     string                 `json:"endpoint_url" binding:"required,url"`
	HealthStatus    A2AHealthStatus        `json:"health_status,omitempty"`
	HealthError     string                 `json:"health_error,omitempty"`
	Description     string                 `json:"description"`
	Name            string                 `json:"name" binding:"required"`
	Tags            []string               `json:"tags"`
	IsActive        bool                   `json:"is_active"`
}

// A2AAgentTool represents a tool exposed by an A2A agent through a virtual server
type A2AAgentTool struct {
	CreatedAt       time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time              `db:"updated_at" json:"updated_at"`
	ToolConfigData  map[string]interface{} `db:"-" json:"tool_config"`
	ToolName        string                 `db:"tool_name" json:"tool_name"`
	ToolConfig      json.RawMessage        `db:"tool_config" json:"-"`
	ID              uuid.UUID              `db:"id" json:"id"`
	AgentID         uuid.UUID              `db:"agent_id" json:"agent_id"`
	VirtualServerID uuid.UUID              `db:"virtual_server_id" json:"virtual_server_id"`
	IsActive        bool                   `db:"is_active" json:"is_active"`
}

// A2ARequest represents a request to an external A2A agent
type A2ARequest struct {
	Parameters      map[string]interface{} `json:"parameters"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	InteractionType string                 `json:"interaction_type"`
	ProtocolVersion string                 `json:"protocol_version"`
	AgentID         string                 `json:"agent_id,omitempty"`
	SessionID       string                 `json:"session_id,omitempty"`
}

// A2AResponse represents a response from an external A2A agent
type A2AResponse struct {
	Data            interface{}            `json:"data,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Usage           *A2AUsage              `json:"usage,omitempty"`
	Error           string                 `json:"error,omitempty"`
	ProtocolVersion string                 `json:"protocol_version"`
	ErrorCode       int                    `json:"error_code,omitempty"`
	Success         bool                   `json:"success"`
}

// A2AUsage represents usage statistics from an A2A agent call
type A2AUsage struct {
	InputTokens  int `json:"input_tokens,omitempty"`
	OutputTokens int `json:"output_tokens,omitempty"`
	TotalTokens  int `json:"total_tokens,omitempty"`
	Duration     int `json:"duration_ms,omitempty"`
}

// A2AChatMessage represents a chat message for A2A agents that support conversation
type A2AChatMessage struct {
	Role    string `json:"role"` // system, user, assistant
	Content string `json:"content"`
}

// A2AChatRequest represents a chat request to an A2A agent
type A2AChatRequest struct {
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Messages    []A2AChatMessage       `json:"messages"`
	Tools       []interface{}          `json:"tools,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
}

// A2AChatResponse represents a chat response from an A2A agent
type A2AChatResponse struct {
	Message      *A2AChatMessage `json:"message,omitempty"`
	FinishReason string          `json:"finish_reason,omitempty"`
	Usage        *A2AUsage       `json:"usage,omitempty"`
	ToolCalls    []interface{}   `json:"tool_calls,omitempty"`
}

// A2AHealthCheck represents a health check request/response
type A2AHealthCheck struct {
	Timestamp    time.Time `json:"timestamp"`
	Status       string    `json:"status"`
	Message      string    `json:"message,omitempty"`
	ResponseTime int       `json:"response_time_ms,omitempty"`
	AgentID      uuid.UUID `json:"agent_id"`
}

// A2AAgentRegistry defines the interface for managing A2A agents
type A2AAgentRegistry interface {
	Create(spec *A2AAgentSpec) (*A2AAgent, error)
	Get(id uuid.UUID) (*A2AAgent, error)
	GetByName(orgID uuid.UUID, name string) (*A2AAgent, error)
	List(orgID uuid.UUID, filters map[string]interface{}) ([]*A2AAgent, error)
	Update(id uuid.UUID, spec *A2AAgentSpec) (*A2AAgent, error)
	Delete(id uuid.UUID) error
	Toggle(id uuid.UUID, active bool) error
	UpdateHealth(id uuid.UUID, status A2AHealthStatus, message string) error
}

// A2AClient defines the interface for communicating with external A2A agents
type A2AClient interface {
	Chat(agent *A2AAgent, request *A2AChatRequest) (*A2AChatResponse, error)
	Invoke(agent *A2AAgent, request *A2ARequest) (*A2AResponse, error)
	HealthCheck(agent *A2AAgent) (*A2AHealthCheck, error)
}

// A2AAdapter defines the interface for integrating A2A agents with virtual servers
type A2AAdapter interface {
	ListTools(agent *A2AAgent) ([]ToolDef, error)
	CallTool(agent *A2AAgent, name string, args map[string]interface{}) (interface{}, error)
	RegisterTool(agentID uuid.UUID, virtualServerID uuid.UUID, toolName string, config map[string]interface{}) error
	UnregisterTool(agentID uuid.UUID, virtualServerID uuid.UUID, toolName string) error
}

// Standard capability keys for A2A agents
const (
	CapabilityChat            = "chat"
	CapabilityTools           = "tools"
	CapabilityStreaming       = "streaming"
	CapabilityFunctionCalling = "function_calling"
	CapabilityDocuments       = "documents"
	CapabilityImages          = "images"
	CapabilityAudio           = "audio"
)

// Standard interaction types for A2A requests
const (
	InteractionTypeQuery    = "query"
	InteractionTypeChat     = "chat"
	InteractionTypeTool     = "tool"
	InteractionTypeHealth   = "health"
	InteractionTypeMetadata = "metadata"
)

// Default configurations for different agent types
var DefaultAgentConfigs = map[AgentType]map[string]interface{}{
	AgentTypeOpenAI: {
		"model":             "gpt-4",
		"max_tokens":        2000,
		"temperature":       0.7,
		"top_p":             1.0,
		"frequency_penalty": 0.0,
		"presence_penalty":  0.0,
	},
	AgentTypeAnthropic: {
		"model":       "claude-3-sonnet-20240229",
		"max_tokens":  4000,
		"temperature": 0.7,
	},
	AgentTypeCustom: {
		"timeout": 30,
		"retries": 3,
	},
	AgentTypeGeneric: {
		"timeout":     30,
		"retries":     3,
		"max_tokens":  1000,
		"temperature": 0.7,
	},
}

// Default capabilities for different agent types
var DefaultAgentCapabilities = map[AgentType]map[string]interface{}{
	AgentTypeOpenAI: {
		CapabilityChat:            true,
		CapabilityTools:           true,
		CapabilityStreaming:       true,
		CapabilityFunctionCalling: true,
		CapabilityImages:          true,
	},
	AgentTypeAnthropic: {
		CapabilityChat:            true,
		CapabilityTools:           true,
		CapabilityStreaming:       false,
		CapabilityFunctionCalling: true,
		CapabilityDocuments:       true,
	},
	AgentTypeCustom: {
		CapabilityChat:      true,
		CapabilityTools:     false,
		CapabilityStreaming: false,
	},
	AgentTypeGeneric: {
		CapabilityChat: true,
	},
}
