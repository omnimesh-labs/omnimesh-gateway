package types

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// VirtualServerSpec defines a virtual MCP server configuration
type VirtualServerSpec struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	AdapterType string    `json:"adapterType" db:"adapter_type"`
	Tools       []ToolDef `json:"tools" db:"tools"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

// ToolDef defines a tool that can be called through MCP
type ToolDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
	REST        *RESTSpec              `json:"REST,omitempty"`
}

// RESTSpec defines how to make REST API calls for a tool
type RESTSpec struct {
	Method      string            `json:"method"`
	URLTemplate string            `json:"URLTemplate"`
	Headers     map[string]string `json:"headers,omitempty"`
	Auth        *AuthSpec         `json:"auth,omitempty"`
	BodyMap     map[string]string `json:"bodyMap,omitempty"`
	TimeoutSec  int               `json:"timeoutSec,omitempty"`
}

// AuthSpec defines authentication configuration for REST calls
type AuthSpec struct {
	Type  string `json:"type"` // Bearer, Basic, etc.
	Token string `json:"token"`
}

// Server interface defines MCP-facing operations
type Server interface {
	Initialize(params InitializeParams) (*InitializeResult, error)
	ListTools() (*ListToolsResult, error)
	CallTool(name string, args map[string]interface{}) (*CallToolResult, error)
}

// Adapter interface defines backend-facing operations for different protocols
type Adapter interface {
	ListTools() ([]ToolDef, error)
	CallTool(name string, args map[string]interface{}) (interface{}, error)
}

// Virtual MCP JSON-RPC request/response structures
type VirtualMCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type VirtualMCPResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      interface{}      `json:"id"`
	Result  interface{}      `json:"result,omitempty"`
	Error   *VirtualMCPError `json:"error,omitempty"`
}

type VirtualMCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCP method parameters and results
type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
	ServerID        string                 `json:"server_id,omitempty"` // Custom param to select virtual server
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      ServerInfo             `json:"serverInfo"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ListToolsParams struct {
	ServerID string `json:"server_id,omitempty"` // Custom param to select virtual server
}

type ListToolsResult struct {
	Tools []Tool `json:"tools"`
}

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
	ServerID  string                 `json:"server_id,omitempty"` // Custom param to select virtual server
}

type CallToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type ToolContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}

// VirtualServerRegistry defines the interface for managing virtual servers
type VirtualServerRegistry interface {
	Add(spec *VirtualServerSpec) error
	Get(id string) (*VirtualServerSpec, error)
	List() ([]*VirtualServerSpec, error)
	Delete(id string) error
	Update(id string, spec *VirtualServerSpec) error
}

// JSON-RPC error codes
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
	ServerError    = -32000
)

// Error messages
var ErrorMessages = map[int]string{
	ParseError:     "Parse error",
	InvalidRequest: "Invalid Request",
	MethodNotFound: "Method not found",
	InvalidParams:  "Invalid params",
	InternalError:  "Internal error",
	ServerError:    "Server error",
}

// Database model for virtual servers
type VirtualServer struct {
	ID             uuid.UUID              `db:"id" json:"id"`
	OrganizationID uuid.UUID              `db:"organization_id" json:"organization_id"`
	Name           string                 `db:"name" json:"name"`
	Description    string                 `db:"description" json:"description"`
	AdapterType    string                 `db:"adapter_type" json:"adapter_type"`
	Tools          json.RawMessage        `db:"tools" json:"-"` // Raw JSON in DB
	ToolsData      []ToolDef              `db:"-" json:"tools"` // Parsed tools for Go
	IsActive       bool                   `db:"is_active" json:"is_active"`
	Metadata       map[string]interface{} `db:"metadata" json:"metadata,omitempty"`
	CreatedAt      time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time              `db:"updated_at" json:"updated_at"`
}
