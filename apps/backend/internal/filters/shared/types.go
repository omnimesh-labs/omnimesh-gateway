package shared

import (
	"context"
	"time"

	"mcp-gateway/apps/backend/internal/types"
)

// FilterType represents the type of content filter
type FilterType string

const (
	FilterTypePII      FilterType = "pii"
	FilterTypeResource FilterType = "resource"
	FilterTypeDeny     FilterType = "deny"
	FilterTypeRegex    FilterType = "regex"
)

// FilterAction represents the action to take when filter is triggered
type FilterAction string

const (
	FilterActionBlock FilterAction = "block"
	FilterActionWarn  FilterAction = "warn"
	FilterActionAudit FilterAction = "audit"
	FilterActionAllow FilterAction = "allow"
)

// FilterDirection indicates whether content is incoming or outgoing
type FilterDirection string

const (
	FilterDirectionInbound  FilterDirection = "inbound"  // User -> MCP Server
	FilterDirectionOutbound FilterDirection = "outbound" // MCP Server -> User
)

// FilterResult represents the result of applying a filter
type FilterResult struct {
	Blocked     bool                   `json:"blocked"`
	Modified    bool                   `json:"modified"`
	Action      FilterAction           `json:"action"`
	Reason      string                 `json:"reason,omitempty"`
	Violations  []FilterViolation      `json:"violations,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ProcessedAt time.Time              `json:"processed_at"`
}

// FilterViolation represents a specific violation found by a filter
type FilterViolation struct {
	Type        string                 `json:"type"`
	Pattern     string                 `json:"pattern,omitempty"`
	Match       string                 `json:"match,omitempty"`
	Position    int                    `json:"position,omitempty"`
	Severity    string                 `json:"severity,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Replacement string                 `json:"replacement,omitempty"`
}

// FilterContext carries data through the filter chain
type FilterContext struct {
	RequestID      string                 `json:"request_id"`
	OrganizationID string                 `json:"organization_id"`
	UserID         string                 `json:"user_id,omitempty"`
	ServerID       string                 `json:"server_id,omitempty"`
	SessionID      string                 `json:"session_id,omitempty"`
	Transport      types.TransportType    `json:"transport"`
	Direction      FilterDirection        `json:"direction"`
	ContentType    string                 `json:"content_type"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
}

// FilterContent represents content to be filtered
type FilterContent struct {
	Raw     string                 `json:"raw"`
	Parsed  interface{}            `json:"parsed,omitempty"`
	Headers map[string]string      `json:"headers,omitempty"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// FilterCapabilities describes what features a filter supports
type FilterCapabilities struct {
	SupportsInbound       bool     `json:"supports_inbound"`
	SupportsOutbound      bool     `json:"supports_outbound"`
	SupportsModification  bool     `json:"supports_modification"`
	SupportsBlocking      bool     `json:"supports_blocking"`
	SupportedContentTypes []string `json:"supported_content_types"`
	SupportsRealtime      bool     `json:"supports_realtime"`
	SupportsBatch         bool     `json:"supports_batch"`
}

// Helper functions for filter types
func (f FilterType) IsValid() bool {
	switch f {
	case FilterTypePII, FilterTypeResource, FilterTypeDeny, FilterTypeRegex:
		return true
	default:
		return false
	}
}

func (f FilterType) String() string {
	return string(f)
}

// Helper functions for filter actions
func (a FilterAction) IsValid() bool {
	switch a {
	case FilterActionBlock, FilterActionWarn, FilterActionAudit, FilterActionAllow:
		return true
	default:
		return false
	}
}

func (a FilterAction) String() string {
	return string(a)
}

// Helper functions for filter directions
func (d FilterDirection) IsValid() bool {
	switch d {
	case FilterDirectionInbound, FilterDirectionOutbound:
		return true
	default:
		return false
	}
}

func (d FilterDirection) String() string {
	return string(d)
}

// Filter defines the interface for content filters
type Filter interface {
	// GetType returns the filter type
	GetType() FilterType

	// GetName returns the filter name/identifier
	GetName() string

	// GetPriority returns the filter priority (lower = higher priority)
	GetPriority() int

	// IsEnabled returns whether the filter is currently enabled
	IsEnabled() bool

	// Apply applies the filter to the given content
	Apply(ctx context.Context, filterCtx *FilterContext, content *FilterContent) (*FilterResult, *FilterContent, error)

	// Configure updates the filter configuration
	Configure(config map[string]interface{}) error

	// GetConfig returns the current filter configuration
	GetConfig() map[string]interface{}

	// Validate validates the filter configuration
	Validate() error

	// GetCapabilities returns what features this filter supports
	GetCapabilities() FilterCapabilities

	// GetStats returns filter statistics
	GetStats() *types.FilterStat
}

// FilterFactory creates filter instances
type FilterFactory interface {
	// Create returns a new filter instance
	Create(config map[string]interface{}) (Filter, error)

	// GetType returns the filter type this factory creates
	GetType() FilterType

	// GetName returns the factory name
	GetName() string

	// GetDescription returns a human-readable description
	GetDescription() string

	// ValidateConfig validates the configuration for this filter type
	ValidateConfig(config map[string]interface{}) error

	// GetDefaultConfig returns the default configuration
	GetDefaultConfig() map[string]interface{}

	// GetConfigSchema returns the JSON schema for configuration validation
	GetConfigSchema() map[string]interface{}
}
