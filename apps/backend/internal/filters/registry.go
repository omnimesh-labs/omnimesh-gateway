package filters

import (
	"fmt"
	"sync"
)

// filterRegistry implements FilterRegistry interface
type filterRegistry struct {
	mu        sync.RWMutex
	factories map[FilterType]FilterFactory
}

// NewFilterRegistry creates a new filter registry
func NewFilterRegistry() FilterRegistry {
	return &filterRegistry{
		factories: make(map[FilterType]FilterFactory),
	}
}

// Register adds a filter factory to the registry
func (r *filterRegistry) Register(factory FilterFactory) error {
	if factory == nil {
		return fmt.Errorf("filter factory cannot be nil")
	}

	filterType := factory.GetType()
	if !filterType.IsValid() {
		return fmt.Errorf("invalid filter type: %s", filterType)
	}

	name := factory.GetName()
	if name == "" {
		return fmt.Errorf("filter factory name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[filterType]; exists {
		return fmt.Errorf("filter type '%s' is already registered", filterType)
	}

	r.factories[filterType] = factory
	return nil
}

// Get returns a filter factory by type
func (r *filterRegistry) Get(filterType FilterType) (FilterFactory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, exists := r.factories[filterType]
	if !exists {
		return nil, fmt.Errorf("filter type '%s' not found", filterType)
	}

	return factory, nil
}

// List returns all available filter types
func (r *filterRegistry) List() []FilterType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]FilterType, 0, len(r.factories))
	for filterType := range r.factories {
		types = append(types, filterType)
	}

	return types
}

// GetInfo returns information about a filter type
func (r *filterRegistry) GetInfo(filterType FilterType) (*FilterInfo, error) {
	factory, err := r.Get(filterType)
	if err != nil {
		return nil, err
	}

	// Create a temporary filter instance to get capabilities
	tempFilter, err := factory.Create(factory.GetDefaultConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary filter for capabilities: %w", err)
	}

	capabilities := tempFilter.GetCapabilities()

	return &FilterInfo{
		Type:          factory.GetType(),
		Name:          factory.GetName(),
		Description:   factory.GetDescription(),
		Version:       "1.0.0", // Could be enhanced to get from factory
		Author:        "MCP Gateway Team",
		ConfigSchema:  factory.GetConfigSchema(),
		Capabilities:  capabilities,
		DefaultConfig: factory.GetDefaultConfig(),
	}, nil
}

// GetAllInfo returns information about all registered filters
func (r *filterRegistry) GetAllInfo() ([]*FilterInfo, error) {
	r.mu.RLock()
	types := make([]FilterType, 0, len(r.factories))
	for filterType := range r.factories {
		types = append(types, filterType)
	}
	r.mu.RUnlock()

	var infos []*FilterInfo
	for _, filterType := range types {
		info, err := r.GetInfo(filterType)
		if err != nil {
			return nil, fmt.Errorf("failed to get info for filter type %s: %w", filterType, err)
		}
		infos = append(infos, info)
	}

	return infos, nil
}

// Global registry instance
var globalRegistry = NewFilterRegistry()

// GetGlobalRegistry returns the global filter registry
func GetGlobalRegistry() FilterRegistry {
	return globalRegistry
}

// RegisterFilter is a convenience function to register a filter globally
func RegisterFilter(factory FilterFactory) error {
	return globalRegistry.Register(factory)
}

// GetFilter is a convenience function to get a filter from the global registry
func GetFilter(filterType FilterType) (FilterFactory, error) {
	return globalRegistry.Get(filterType)
}

// ListFilters is a convenience function to list all filters
func ListFilters() []FilterType {
	return globalRegistry.List()
}

// GetFilterInfo is a convenience function to get filter info
func GetFilterInfo(filterType FilterType) (*FilterInfo, error) {
	return globalRegistry.GetInfo(filterType)
}

// GetAllFilterInfo is a convenience function to get all filter info
func GetAllFilterInfo() ([]*FilterInfo, error) {
	return globalRegistry.GetAllInfo()
}