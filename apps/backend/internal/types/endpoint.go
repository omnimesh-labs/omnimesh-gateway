package types

import (
	"time"
)

// Endpoint represents a public-facing URL that maps to a namespace
type Endpoint struct {
	ID                 string                 `json:"id" db:"id"`
	OrganizationID     string                 `json:"organization_id" db:"organization_id"`
	NamespaceID        string                 `json:"namespace_id" db:"namespace_id"`
	Name               string                 `json:"name" db:"name"`
	Description        string                 `json:"description" db:"description"`

	// Authentication
	EnableAPIKeyAuth   bool                   `json:"enable_api_key_auth" db:"enable_api_key_auth"`
	EnableOAuth        bool                   `json:"enable_oauth" db:"enable_oauth"`
	EnablePublicAccess bool                   `json:"enable_public_access" db:"enable_public_access"`
	UseQueryParamAuth  bool                   `json:"use_query_param_auth" db:"use_query_param_auth"`

	// Rate limiting
	RateLimitRequests  int                    `json:"rate_limit_requests" db:"rate_limit_requests"`
	RateLimitWindow    int                    `json:"rate_limit_window" db:"rate_limit_window"`

	// CORS
	AllowedOrigins     []string               `json:"allowed_origins" db:"allowed_origins"`
	AllowedMethods     []string               `json:"allowed_methods" db:"allowed_methods"`

	CreatedAt          time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at" db:"updated_at"`
	CreatedBy          *string                `json:"created_by" db:"created_by"`
	IsActive           bool                   `json:"is_active" db:"is_active"`
	Metadata           map[string]interface{} `json:"metadata" db:"metadata"`

	// Computed fields
	Namespace          *Namespace             `json:"namespace,omitempty"`
	URLs               *EndpointURLs          `json:"urls,omitempty"`
}

// EndpointURLs represents the available URLs for an endpoint
type EndpointURLs struct {
	SSE           string `json:"sse"`
	HTTP          string `json:"http"`
	WebSocket     string `json:"websocket"`
	OpenAPI       string `json:"openapi"`
	Documentation string `json:"documentation"`
}

// CreateEndpointRequest represents the request to create an endpoint
type CreateEndpointRequest struct {
	NamespaceID        string                 `json:"namespace_id" binding:"required"`
	Name               string                 `json:"name" binding:"required"`
	Description        string                 `json:"description"`
	EnableAPIKeyAuth   bool                   `json:"enable_api_key_auth"`
	EnableOAuth        bool                   `json:"enable_oauth"`
	EnablePublicAccess bool                   `json:"enable_public_access"`
	UseQueryParamAuth  bool                   `json:"use_query_param_auth"`
	RateLimitRequests  int                    `json:"rate_limit_requests"`
	RateLimitWindow    int                    `json:"rate_limit_window"`
	AllowedOrigins     []string               `json:"allowed_origins"`
	AllowedMethods     []string               `json:"allowed_methods"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// UpdateEndpointRequest represents the request to update an endpoint
type UpdateEndpointRequest struct {
	Description        string                 `json:"description,omitempty"`
	EnableAPIKeyAuth   *bool                  `json:"enable_api_key_auth,omitempty"`
	EnableOAuth        *bool                  `json:"enable_oauth,omitempty"`
	EnablePublicAccess *bool                  `json:"enable_public_access,omitempty"`
	UseQueryParamAuth  *bool                  `json:"use_query_param_auth,omitempty"`
	RateLimitRequests  *int                   `json:"rate_limit_requests,omitempty"`
	RateLimitWindow    *int                   `json:"rate_limit_window,omitempty"`
	AllowedOrigins     []string               `json:"allowed_origins,omitempty"`
	AllowedMethods     []string               `json:"allowed_methods,omitempty"`
	IsActive           *bool                  `json:"is_active,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// EndpointConfig represents the configuration for an endpoint (used in middleware)
type EndpointConfig struct {
	Endpoint  *Endpoint
	Namespace *Namespace
}
