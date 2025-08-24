package plugins

import (
	"fmt"
	"sync"
)

// pluginRegistry implements PluginRegistry interface
type pluginRegistry struct {
	mu        sync.RWMutex
	factories map[PluginType]PluginFactory
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry() PluginRegistry {
	return &pluginRegistry{
		factories: make(map[PluginType]PluginFactory),
	}
}

// Register adds a plugin factory to the registry
func (r *pluginRegistry) Register(factory PluginFactory) error {
	if factory == nil {
		return fmt.Errorf("plugin factory cannot be nil")
	}

	pluginType := factory.GetType()
	if !pluginType.IsValid() {
		return fmt.Errorf("invalid plugin type: %s", pluginType)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[pluginType]; exists {
		return fmt.Errorf("plugin type '%s' is already registered", pluginType)
	}

	r.factories[pluginType] = factory
	return nil
}

// Get returns a plugin factory by type
func (r *pluginRegistry) Get(pluginType PluginType) (PluginFactory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, exists := r.factories[pluginType]
	if !exists {
		return nil, fmt.Errorf("plugin type '%s' not found", pluginType)
	}

	return factory, nil
}

// List returns all available plugin types
func (r *pluginRegistry) List() []PluginType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]PluginType, 0, len(r.factories))
	for pluginType := range r.factories {
		types = append(types, pluginType)
	}

	return types
}

// GetInfo returns information about a plugin type
func (r *pluginRegistry) GetInfo(pluginType PluginType) (*PluginInfo, error) {
	factory, err := r.Get(pluginType)
	if err != nil {
		return nil, err
	}

	// Create a plugin instance to get capabilities
	defaultConfig := factory.GetDefaultConfig()
	plugin, err := factory.Create(defaultConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin instance for info: %w", err)
	}

	capabilities := plugin.GetCapabilities()

	// Determine category based on plugin type
	category := "content"
	if pluginType.IsAIPlugin() {
		category = "ai"
	}

	return &PluginInfo{
		Type:           pluginType,
		Name:           factory.GetName(),
		Description:    factory.GetDescription(),
		Version:        "1.0.0", // Could be enhanced to get from plugin
		Author:         "MCP Gateway Team",
		ConfigSchema:   factory.GetConfigSchema(),
		Capabilities:   capabilities,
		DefaultConfig:  defaultConfig,
		SupportedModes: factory.GetSupportedExecutionModes(),
		RequiresAPI:    capabilities.RequiresExternalAPI,
		Category:       category,
	}, nil
}

// GetAllInfo returns information about all registered plugins
func (r *pluginRegistry) GetAllInfo() ([]*PluginInfo, error) {
	r.mu.RLock()
	types := make([]PluginType, 0, len(r.factories))
	for pluginType := range r.factories {
		types = append(types, pluginType)
	}
	r.mu.RUnlock()

	infos := make([]*PluginInfo, 0, len(types))
	for _, pluginType := range types {
		info, err := r.GetInfo(pluginType)
		if err != nil {
			return nil, fmt.Errorf("failed to get info for plugin %s: %w", pluginType, err)
		}
		infos = append(infos, info)
	}

	return infos, nil
}

// GetByCategory returns plugins of a specific category
func (r *pluginRegistry) GetByCategory(category string) []PluginFactory {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var factories []PluginFactory
	for pluginType, factory := range r.factories {
		pluginCategory := "content"
		if pluginType.IsAIPlugin() {
			pluginCategory = "ai"
		}

		if pluginCategory == category {
			factories = append(factories, factory)
		}
	}

	return factories
}

// Global registry instance
var globalRegistry = NewPluginRegistry()

// GetGlobalRegistry returns the global plugin registry
func GetGlobalRegistry() PluginRegistry {
	return globalRegistry
}

// RegisterPlugin is a convenience function to register a plugin globally
func RegisterPlugin(factory PluginFactory) error {
	return globalRegistry.Register(factory)
}

// GetPlugin is a convenience function to get a plugin from the global registry
func GetPlugin(pluginType PluginType) (PluginFactory, error) {
	return globalRegistry.Get(pluginType)
}

// ListPlugins is a convenience function to list all plugins
func ListPlugins() []PluginType {
	return globalRegistry.List()
}
