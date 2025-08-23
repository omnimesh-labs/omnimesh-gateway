package types

import "time"

// User represents a user in the system
type User struct {
	ID             string    `json:"id" db:"id"`
	Email          string    `json:"email" db:"email"`
	Name           string    `json:"name" db:"name"`
	PasswordHash   string    `json:"-" db:"password_hash"`
	OrganizationID string    `json:"organization_id" db:"organization_id"`
	Role           string    `json:"role" db:"role"`
	IsActive       bool      `json:"is_active" db:"is_active"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Organization represents an organization
type Organization struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// APIKey represents an API key
type APIKey struct {
	ID             string     `json:"id" db:"id"`
	UserID         string     `json:"user_id" db:"user_id"`
	OrganizationID string     `json:"organization_id" db:"organization_id"`
	Name           string     `json:"name" db:"name"`
	KeyHash        string     `json:"-" db:"key_hash"`
	Prefix         string     `json:"prefix" db:"prefix"`
	Permissions    []string   `json:"permissions" db:"permissions"`
	ExpiresAt      *time.Time `json:"expires_at" db:"expires_at"`
	LastUsedAt     *time.Time `json:"last_used_at" db:"last_used_at"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// Policy represents an access control policy
type Policy struct {
	ID             string                 `json:"id" db:"id"`
	OrganizationID string                 `json:"organization_id" db:"organization_id"`
	Name           string                 `json:"name" db:"name"`
	Description    string                 `json:"description" db:"description"`
	Type           string                 `json:"type" db:"type"` // "access", "rate_limit", "routing"
	Priority       int                    `json:"priority" db:"priority"`
	Conditions     map[string]interface{} `json:"conditions" db:"conditions"`
	Actions        map[string]interface{} `json:"actions" db:"actions"`
	IsActive       bool                   `json:"is_active" db:"is_active"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// TokenResponse represents a token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// CreateUserRequest represents a user creation request
type CreateUserRequest struct {
	Email          string `json:"email" binding:"required,email"`
	Name           string `json:"name" binding:"required,min=2"`
	Password       string `json:"password" binding:"required,min=8"`
	OrganizationID string `json:"organization_id" binding:"required"`
	Role           string `json:"role" binding:"required"`
}

// UpdateUserRequest represents a user update request
type UpdateUserRequest struct {
	Name string `json:"name,omitempty" binding:"omitempty,min=2"`
	Role string `json:"role,omitempty"`
}

// CreateAPIKeyRequest represents an API key creation request
type CreateAPIKeyRequest struct {
	Name        string     `json:"name" binding:"required,min=2"`
	Permissions []string   `json:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// CreateOrganizationRequest represents an organization creation request
type CreateOrganizationRequest struct {
	Name        string `json:"name" binding:"required,min=2"`
	Description string `json:"description"`
}

// UpdateOrganizationRequest represents an organization update request
type UpdateOrganizationRequest struct {
	Name        string `json:"name,omitempty" binding:"omitempty,min=2"`
	Description string `json:"description,omitempty"`
}

// CreatePolicyRequest represents a policy creation request
type CreatePolicyRequest struct {
	Name        string                 `json:"name" binding:"required,min=2"`
	Description string                 `json:"description"`
	Type        string                 `json:"type" binding:"required"`
	Priority    int                    `json:"priority"`
	Conditions  map[string]interface{} `json:"conditions" binding:"required"`
	Actions     map[string]interface{} `json:"actions" binding:"required"`
}

// UpdatePolicyRequest represents a policy update request
type UpdatePolicyRequest struct {
	Name        string                 `json:"name,omitempty" binding:"omitempty,min=2"`
	Description string                 `json:"description,omitempty"`
	Priority    int                    `json:"priority,omitempty"`
	Conditions  map[string]interface{} `json:"conditions,omitempty"`
	Actions     map[string]interface{} `json:"actions,omitempty"`
	IsActive    *bool                  `json:"is_active,omitempty"`
}

// UserRole constants
const (
	RoleAdmin       = "admin"
	RoleUser        = "user"
	RoleViewer      = "viewer"
	RoleAPIUser     = "api_user"
	RoleSystemAdmin = "system_admin"
)

// Permission constants
const (
	// Basic permissions
	PermissionRead       = "read"
	PermissionWrite      = "write"
	PermissionDelete     = "delete"
	PermissionAdmin      = "admin"
	
	// API access permissions
	PermissionAPIAccess  = "api_access"
	PermissionAPIKeyManage = "api_key_manage"
	
	// User management permissions
	PermissionUserRead    = "user_read"
	PermissionUserWrite   = "user_write"
	PermissionUserDelete  = "user_delete"
	PermissionUserManage  = "user_manage"
	
	// Server management permissions
	PermissionServerRead   = "server_read"
	PermissionServerWrite  = "server_write" 
	PermissionServerDelete = "server_delete"
	PermissionServerManage = "server_manage"
	
	// Session management permissions
	PermissionSessionRead   = "session_read"
	PermissionSessionWrite  = "session_write"
	PermissionSessionDelete = "session_delete"
	PermissionSessionManage = "session_manage"
	
	// Virtual server permissions
	PermissionVirtualServerRead   = "virtual_server_read"
	PermissionVirtualServerWrite  = "virtual_server_write"
	PermissionVirtualServerDelete = "virtual_server_delete"
	PermissionVirtualServerManage = "virtual_server_manage"
	
	// Resource permissions
	PermissionResourceRead   = "resource_read"
	PermissionResourceWrite  = "resource_write"
	PermissionResourceDelete = "resource_delete"
	PermissionResourceManage = "resource_manage"
	
	// Prompt permissions
	PermissionPromptRead   = "prompt_read"
	PermissionPromptWrite  = "prompt_write"
	PermissionPromptDelete = "prompt_delete"
	PermissionPromptManage = "prompt_manage"
	
	// Tool permissions
	PermissionToolRead    = "tool_read"
	PermissionToolWrite   = "tool_write"
	PermissionToolDelete  = "tool_delete"
	PermissionToolManage  = "tool_manage"
	PermissionToolExecute = "tool_execute"
	
	// Audit and logging permissions
	PermissionAuditRead      = "audit_read"
	PermissionLogsRead       = "logs_read"
	PermissionMetricsRead    = "metrics_read"
	PermissionSystemManage   = "system_manage"
	
	// Organization permissions
	PermissionOrgRead    = "org_read"
	PermissionOrgWrite   = "org_write"
	PermissionOrgDelete  = "org_delete"
	PermissionOrgManage  = "org_manage"
)

// PolicyType constants
const (
	PolicyTypeAccess    = "access"
	PolicyTypeRateLimit = "rate_limit"
	PolicyTypeRouting   = "routing"
	PolicyTypeSecurity  = "security"
)
