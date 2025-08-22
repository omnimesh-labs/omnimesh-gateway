package filters

import (
	"context"

	"mcp-gateway/apps/backend/internal/database/models"
	"mcp-gateway/apps/backend/internal/filters/shared"
	"mcp-gateway/apps/backend/internal/types"
)

// Re-export shared types for backward compatibility
type FilterType = shared.FilterType
type FilterAction = shared.FilterAction
type FilterResult = shared.FilterResult
type FilterViolation = shared.FilterViolation
type FilterContext = shared.FilterContext
type FilterDirection = shared.FilterDirection
type FilterContent = shared.FilterContent
type FilterCapabilities = shared.FilterCapabilities

// Re-export constants
const (
	FilterTypePII      = shared.FilterTypePII
	FilterTypeResource = shared.FilterTypeResource
	FilterTypeDeny     = shared.FilterTypeDeny
	FilterTypeRegex    = shared.FilterTypeRegex
)

const (
	FilterActionBlock = shared.FilterActionBlock
	FilterActionWarn  = shared.FilterActionWarn
	FilterActionAudit = shared.FilterActionAudit
	FilterActionAllow = shared.FilterActionAllow
)

const (
	FilterDirectionInbound  = shared.FilterDirectionInbound
	FilterDirectionOutbound = shared.FilterDirectionOutbound
)

// Filter is re-exported from shared package for backward compatibility
type Filter = shared.Filter


// FilterFactory is re-exported from shared package for backward compatibility
type FilterFactory = shared.FilterFactory

// FilterRegistry manages available filter factories
type FilterRegistry interface {
	// Register adds a filter factory to the registry
	Register(factory FilterFactory) error

	// Get returns a filter factory by type
	Get(filterType FilterType) (FilterFactory, error)

	// List returns all available filter types
	List() []FilterType

	// GetInfo returns information about a filter type
	GetInfo(filterType FilterType) (*FilterInfo, error)

	// GetAllInfo returns information about all registered filters
	GetAllInfo() ([]*FilterInfo, error)
}

// FilterInfo contains metadata about a filter type
type FilterInfo struct {
	Type          FilterType             `json:"type"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Version       string                 `json:"version"`
	Author        string                 `json:"author"`
	ConfigSchema  map[string]any `json:"config_schema,omitempty"`
	Capabilities  FilterCapabilities     `json:"capabilities"`
	DefaultConfig map[string]any `json:"default_config,omitempty"`
}

// FilterManager orchestrates multiple filters
type FilterManager interface {
	// AddFilter adds a filter to the manager
	AddFilter(filter Filter) error

	// RemoveFilter removes a filter from the manager
	RemoveFilter(filterName string) error

	// GetFilter returns a filter by name
	GetFilter(filterName string) (Filter, error)

	// ListFilters returns all registered filters
	ListFilters() []Filter

	// ApplyFilters applies all enabled filters to content
	ApplyFilters(ctx context.Context, filterCtx *FilterContext, content *FilterContent) (*FilterResult, *FilterContent, error)

	// ApplyFiltersInOrder applies filters in priority order
	ApplyFiltersInOrder(ctx context.Context, filterCtx *FilterContext, content *FilterContent) ([]*FilterResult, *FilterContent, error)

	// EnableFilter enables a filter by name
	EnableFilter(filterName string) error

	// DisableFilter disables a filter by name
	DisableFilter(filterName string) error

	// ReloadConfiguration reloads all filter configurations
	ReloadConfiguration() error

	// GetStats returns filtering statistics
	GetStats() (*types.FilteringMetrics, error)
}

// FilterStats contains filtering statistics - using types.FilteringMetrics for consistency
type FilterStats = types.FilteringMetrics

// FilterService defines the main filtering service interface
type FilterService interface {
	// Initialize sets up the filtering service
	Initialize(ctx context.Context) error

	// ProcessContent processes content through all applicable filters
	ProcessContent(ctx context.Context, filterCtx *FilterContext, content *FilterContent) (*FilterResult, *FilterContent, error)

	// GetManager returns the filter manager
	GetManager() FilterManager

	// GetRegistry returns the filter registry
	GetRegistry() FilterRegistry

	// LoadFiltersFromDatabase loads filters from database configuration
	LoadFiltersFromDatabase(ctx context.Context, organizationID string) error

	// SaveFilterToDatabase saves a filter configuration to the database
	SaveFilterToDatabase(ctx context.Context, organizationID string, filter Filter) error

	// DeleteFilterFromDatabase deletes a filter configuration from the database
	DeleteFilterFromDatabase(ctx context.Context, organizationID, filterName string) error

	// ReloadOrganizationFilters reloads filters for a specific organization
	ReloadOrganizationFilters(ctx context.Context, organizationID string) error

	// GetOrganizationFilters returns all filters for an organization
	GetOrganizationFilters(ctx context.Context, organizationID string) ([]Filter, error)

	// HealthCheck verifies the service is operational
	HealthCheck(ctx context.Context) error

	// Close shuts down the service
	Close() error

	// GetMetrics returns filtering metrics
	GetMetrics() (*types.FilteringMetrics, error)

	// GetViolations retrieves filter violations with optional filtering  
	GetViolations(ctx context.Context, organizationID string, limit, offset int) ([]*models.FilterViolation, error)
}

// Helper functions for filter types, actions, and directions are available in the shared package
