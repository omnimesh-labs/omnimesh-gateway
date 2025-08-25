package types

import (
	"database/sql/driver"
	"time"
)

// Namespace represents a logical grouping of MCP servers
type Namespace struct {
	ID             string                 `json:"id" db:"id"`
	OrganizationID string                 `json:"organization_id" db:"organization_id"`
	Name           string                 `json:"name" db:"name"`
	Description    string                 `json:"description" db:"description"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
	CreatedBy      *string                `json:"created_by" db:"created_by"`
	IsActive       bool                   `json:"is_active" db:"is_active"`
	Metadata       map[string]interface{} `json:"metadata" db:"metadata"`
	Servers        []NamespaceServer      `json:"servers,omitempty"`
	Tools          []NamespaceTool        `json:"tools,omitempty"`
	ServerCount    int                    `json:"server_count,omitempty" db:"-"`
}

// NamespaceServer represents a server within a namespace
type NamespaceServer struct {
	ServerID   string    `json:"server_id" db:"server_id"`
	ServerName string    `json:"server_name" db:"server_name"`
	Status     string    `json:"status" db:"status"`
	Priority   int       `json:"priority" db:"priority"`
	JoinedAt   time.Time `json:"joined_at" db:"created_at"`
}

// NamespaceTool represents a tool exposed by a server in a namespace
type NamespaceTool struct {
	ServerID     string `json:"server_id" db:"server_id"`
	ServerName   string `json:"server_name,omitempty"`
	ToolName     string `json:"tool_name" db:"tool_name"`
	PrefixedName string `json:"prefixed_name"`
	Status       string `json:"status" db:"status"`
	Description  string `json:"description,omitempty"`
}

// NamespaceServerMapping represents the mapping between namespace and server
type NamespaceServerMapping struct {
	ID          string    `json:"id" db:"id"`
	NamespaceID string    `json:"namespace_id" db:"namespace_id"`
	ServerID    string    `json:"server_id" db:"server_id"`
	Status      string    `json:"status" db:"status"`
	Priority    int       `json:"priority" db:"priority"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// NamespaceToolMapping represents the mapping between namespace and tool
type NamespaceToolMapping struct {
	ID          string    `json:"id" db:"id"`
	NamespaceID string    `json:"namespace_id" db:"namespace_id"`
	ServerID    string    `json:"server_id" db:"server_id"`
	ToolName    string    `json:"tool_name" db:"tool_name"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// NamespaceStatus represents the status of a namespace entity
type NamespaceStatus string

const (
	NamespaceStatusActive   NamespaceStatus = "ACTIVE"
	NamespaceStatusInactive NamespaceStatus = "INACTIVE"
)

// Value implements driver.Valuer interface
func (s NamespaceStatus) Value() (driver.Value, error) {
	return string(s), nil
}

// CreateNamespaceRequest represents the request to create a namespace
type CreateNamespaceRequest struct {
	Name           string                 `json:"name" binding:"required"`
	Description    string                 `json:"description"`
	OrganizationID string                 `json:"organization_id"`
	CreatedBy      *string                `json:"created_by,omitempty"`
	Servers        []string               `json:"servers"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// UpdateNamespaceRequest represents the request to update a namespace
type UpdateNamespaceRequest struct {
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	IsActive    *bool                  `json:"is_active,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ServerIDs   []string               `json:"server_ids,omitempty"`
}

// AddServerToNamespaceRequest represents the request to add a server to namespace
type AddServerToNamespaceRequest struct {
	ServerID string `json:"server_id" binding:"required"`
	Priority int    `json:"priority"`
}

// UpdateServerStatusRequest represents the request to update server status in namespace
type UpdateServerStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=ACTIVE INACTIVE"`
}

// UpdateToolStatusRequest represents the request to update tool status in namespace
type UpdateToolStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=ACTIVE INACTIVE"`
}

// ExecuteNamespaceToolRequest represents the request to execute a tool in namespace
type ExecuteNamespaceToolRequest struct {
	Tool      string                 `json:"tool" binding:"required"`
	Arguments map[string]interface{} `json:"arguments"`
}

// NamespaceToolResult represents the result of a tool execution
type NamespaceToolResult struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
}
