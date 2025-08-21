package logging

import (
	"fmt"
	"sync"
)

// pluginRegistry implements PluginRegistry interface
type pluginRegistry struct {
	mu        sync.RWMutex
	factories map[string]PluginFactory
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry() PluginRegistry {
	return &pluginRegistry{
		factories: make(map[string]PluginFactory),
	}
}

// Register adds a plugin factory to the registry
func (r *pluginRegistry) Register(factory PluginFactory) error {
	if factory == nil {
		return fmt.Errorf("plugin factory cannot be nil")
	}

	name := factory.GetName()
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("plugin '%s' is already registered", name)
	}

	r.factories[name] = factory
	return nil
}

// Get returns a plugin factory by name
func (r *pluginRegistry) Get(name string) (PluginFactory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, exists := r.factories[name]
	if !exists {
		return nil, fmt.Errorf("plugin '%s' not found", name)
	}

	return factory, nil
}

// List returns all available plugin names
func (r *pluginRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}

	return names
}

// GetInfo returns information about a plugin
func (r *pluginRegistry) GetInfo(name string) (*PluginInfo, error) {
	factory, err := r.Get(name)
	if err != nil {
		return nil, err
	}

	// Create a backend instance to get capabilities
	backend := factory.Create()
	capabilities := backend.GetCapabilities()

	return &PluginInfo{
		Name:         factory.GetName(),
		Description:  factory.GetDescription(),
		Version:      "1.0.0", // Could be enhanced to get from plugin
		Author:       "MCP Gateway Team",
		Capabilities: capabilities,
	}, nil
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
func GetPlugin(name string) (PluginFactory, error) {
	return globalRegistry.Get(name)
}

// ListPlugins is a convenience function to list all plugins
func ListPlugins() []string {
	return globalRegistry.List()
}
