package plugins

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"mcp-gateway/apps/backend/internal/plugins/shared"
	"mcp-gateway/apps/backend/internal/types"
)

// pluginManager implements PluginManager interface
type pluginManager struct {
	mu      sync.RWMutex
	plugins map[string]shared.Plugin
	stats   *types.FilteringMetrics
}

// NewPluginManager creates a new plugin manager
func NewPluginManager() shared.PluginManager {
	return &pluginManager{
		plugins: make(map[string]Plugin),
		stats: &types.FilteringMetrics{
			TotalRequests:    0,
			TotalBlocked:     0,
			TotalModified:    0,
			FilterStats:      make(map[string]*types.FilterStat),
			ViolationsByType: make(map[string]int64),
			ProcessingTime:   0,
			LastReset:        time.Now(),
		},
	}
}

// AddPlugin adds a plugin to the manager
func (m *pluginManager) AddPlugin(plugin Plugin) error {
	if plugin == nil {
		return fmt.Errorf("plugin cannot be nil")
	}

	name := plugin.GetName()
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[name]; exists {
		return fmt.Errorf("plugin '%s' is already registered", name)
	}

	m.plugins[name] = plugin
	
	// Initialize stats for this plugin
	m.stats.FilterStats[name] = plugin.GetStats()

	return nil
}

// RemovePlugin removes a plugin from the manager
func (m *pluginManager) RemovePlugin(pluginName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[pluginName]; !exists {
		return fmt.Errorf("plugin '%s' not found", pluginName)
	}

	delete(m.plugins, pluginName)
	delete(m.stats.FilterStats, pluginName)

	return nil
}

// GetPlugin returns a plugin by name
func (m *pluginManager) GetPlugin(pluginName string) (Plugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, exists := m.plugins[pluginName]
	if !exists {
		return nil, fmt.Errorf("plugin '%s' not found", pluginName)
	}

	return plugin, nil
}

// ListPlugins returns all registered plugins
func (m *pluginManager) ListPlugins() []Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugins := make([]Plugin, 0, len(m.plugins))
	for _, plugin := range m.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// ApplyPlugins applies all enabled plugins to content
func (m *pluginManager) ApplyPlugins(ctx context.Context, pluginCtx *PluginContext, content *PluginContent) (*PluginResult, *PluginContent, error) {
	results, modifiedContent, err := m.ApplyPluginsInOrder(ctx, pluginCtx, content)
	if err != nil {
		return nil, content, err
	}

	// Merge all results into a single result
	mergedResult := shared.MergePluginResults(results)
	
	return mergedResult, modifiedContent, nil
}

// ApplyPluginsInOrder applies plugins in priority order
func (m *pluginManager) ApplyPluginsInOrder(ctx context.Context, pluginCtx *PluginContext, content *PluginContent) ([]*PluginResult, *PluginContent, error) {
	// Validate inputs
	if err := shared.ValidatePluginContext(pluginCtx); err != nil {
		return nil, content, fmt.Errorf("invalid plugin context: %w", err)
	}

	if err := shared.ValidatePluginContent(content); err != nil {
		return nil, content, fmt.Errorf("invalid plugin content: %w", err)
	}

	startTime := time.Now()
	defer func() {
		m.updateGlobalStats(time.Since(startTime))
	}()

	// Get enabled plugins sorted by priority
	enabledPlugins := m.getEnabledPluginsSorted()
	if len(enabledPlugins) == 0 {
		return []*PluginResult{}, content, nil
	}

	var results []*PluginResult
	currentContent := content

	for _, plugin := range enabledPlugins {
		// Check if this plugin supports the current direction
		capabilities := plugin.GetCapabilities()
		if !m.supportsDirection(capabilities, pluginCtx.Direction) {
			continue
		}

		// Check if this plugin supports the content type
		if !m.supportsContentType(capabilities, pluginCtx.ContentType) {
			continue
		}

		// Apply the plugin with timing
		result, modifiedContent, duration, err := shared.ApplyPluginWithTiming(ctx, plugin, pluginCtx, currentContent)
		if err != nil {
			// Update error stats
			if basePlugin, ok := plugin.(*shared.BasePlugin); ok {
				basePlugin.UpdateStats(false, false, 0, duration, true)
			}
			return results, currentContent, fmt.Errorf("plugin '%s' failed: %w", plugin.GetName(), err)
		}

		// Update plugin stats
		if basePlugin, ok := plugin.(*shared.BasePlugin); ok {
			basePlugin.UpdateStats(result.Blocked, result.Modified, len(result.Violations), duration, false)
		}

		results = append(results, result)

		// If content was modified, update current content for next plugin
		if result.Modified && modifiedContent != nil {
			currentContent = modifiedContent
		}

		// If any plugin blocks, stop processing (short-circuit)
		if result.Blocked && result.Action == PluginActionBlock {
			break
		}
	}

	// Update global stats
	m.updateRequestStats(results)

	return results, currentContent, nil
}

// ApplyPluginsForDirection applies plugins for specific direction
func (m *pluginManager) ApplyPluginsForDirection(ctx context.Context, pluginCtx *PluginContext, content *PluginContent, direction PluginDirection) ([]*PluginResult, *PluginContent, error) {
	// Create a new context with the specified direction
	directionCtx := *pluginCtx
	directionCtx.Direction = direction
	
	return m.ApplyPluginsInOrder(ctx, &directionCtx, content)
}

// EnablePlugin enables a plugin by name
func (m *pluginManager) EnablePlugin(pluginName string) error {
	plugin, err := m.GetPlugin(pluginName)
	if err != nil {
		return err
	}

	if basePlugin, ok := plugin.(*shared.BasePlugin); ok {
		basePlugin.SetEnabled(true)
		return nil
	}

	return fmt.Errorf("plugin '%s' does not support enable/disable operations", pluginName)
}

// DisablePlugin disables a plugin by name
func (m *pluginManager) DisablePlugin(pluginName string) error {
	plugin, err := m.GetPlugin(pluginName)
	if err != nil {
		return err
	}

	if basePlugin, ok := plugin.(*shared.BasePlugin); ok {
		basePlugin.SetEnabled(false)
		return nil
	}

	return fmt.Errorf("plugin '%s' does not support enable/disable operations", pluginName)
}

// SetPluginExecutionMode sets execution mode for a plugin
func (m *pluginManager) SetPluginExecutionMode(pluginName string, mode PluginExecutionMode) error {
	plugin, err := m.GetPlugin(pluginName)
	if err != nil {
		return err
	}

	if basePlugin, ok := plugin.(*shared.BasePlugin); ok {
		return basePlugin.SetExecutionMode(string(mode))
	}

	return fmt.Errorf("plugin '%s' does not support execution mode operations", pluginName)
}

// ReloadConfiguration reloads all plugin configurations
func (m *pluginManager) ReloadConfiguration() error {
	m.mu.RLock()
	plugins := make([]Plugin, 0, len(m.plugins))
	for _, plugin := range m.plugins {
		plugins = append(plugins, plugin)
	}
	m.mu.RUnlock()

	for _, plugin := range plugins {
		if err := plugin.Validate(); err != nil {
			return fmt.Errorf("plugin '%s' configuration validation failed: %w", plugin.GetName(), err)
		}
	}

	return nil
}

// GetStats returns plugin statistics
func (m *pluginManager) GetStats() (*types.FilteringMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a deep copy to prevent race conditions
	statsCopy := &types.FilteringMetrics{
		TotalRequests:    m.stats.TotalRequests,
		TotalBlocked:     m.stats.TotalBlocked,
		TotalModified:    m.stats.TotalModified,
		FilterStats:      make(map[string]*types.FilterStat),
		ViolationsByType: make(map[string]int64),
		ProcessingTime:   m.stats.ProcessingTime,
		LastReset:        m.stats.LastReset,
	}

	for name, stat := range m.stats.FilterStats {
		statCopy := *stat
		statsCopy.FilterStats[name] = &statCopy
	}

	for violationType, count := range m.stats.ViolationsByType {
		statsCopy.ViolationsByType[violationType] = count
	}

	return statsCopy, nil
}

// getEnabledPluginsSorted returns enabled plugins sorted by priority
func (m *pluginManager) getEnabledPluginsSorted() []Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var enabledPlugins []Plugin
	for _, plugin := range m.plugins {
		if plugin.IsEnabled() {
			enabledPlugins = append(enabledPlugins, plugin)
		}
	}

	// Sort by priority (lower number = higher priority)
	sort.Slice(enabledPlugins, func(i, j int) bool {
		return enabledPlugins[i].GetPriority() < enabledPlugins[j].GetPriority()
	})

	return enabledPlugins
}

// supportsDirection checks if plugin supports the given direction
func (m *pluginManager) supportsDirection(capabilities PluginCapabilities, direction PluginDirection) bool {
	switch direction {
	case PluginDirectionInbound:
		return capabilities.SupportsInbound
	case PluginDirectionOutbound:
		return capabilities.SupportsOutbound
	case PluginDirectionPreTool:
		return capabilities.SupportsPreTool
	case PluginDirectionPostTool:
		return capabilities.SupportsPostTool
	default:
		return false
	}
}

// supportsContentType checks if plugin supports the given content type
func (m *pluginManager) supportsContentType(capabilities PluginCapabilities, contentType string) bool {
	// If no specific content types are specified, assume it supports all
	if len(capabilities.SupportedContentTypes) == 0 {
		return true
	}

	for _, supportedType := range capabilities.SupportedContentTypes {
		if supportedType == contentType || supportedType == "*" {
			return true
		}
	}

	return false
}

// updateGlobalStats updates global processing statistics
func (m *pluginManager) updateGlobalStats(processingTime time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats.TotalRequests++
	
	// Update average processing time using running average
	if m.stats.TotalRequests == 1 {
		m.stats.ProcessingTime = processingTime
	} else {
		m.stats.ProcessingTime = m.stats.ProcessingTime + 
			(processingTime - m.stats.ProcessingTime) / time.Duration(m.stats.TotalRequests)
	}
}

// updateRequestStats updates statistics based on plugin results
func (m *pluginManager) updateRequestStats(results []*PluginResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	blocked := false
	modified := false

	for _, result := range results {
		if result.Blocked {
			blocked = true
		}
		if result.Modified {
			modified = true
		}

		// Count violations by type
		for _, violation := range result.Violations {
			m.stats.ViolationsByType[violation.Type]++
		}
	}

	if blocked {
		m.stats.TotalBlocked++
	}
	if modified {
		m.stats.TotalModified++
	}

	// Update individual plugin stats
	for pluginName, plugin := range m.plugins {
		if pluginStat := plugin.GetStats(); pluginStat != nil {
			m.stats.FilterStats[pluginName] = pluginStat
		}
	}
}

// ResetStats resets all statistics
func (m *pluginManager) ResetStats() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats = &types.FilteringMetrics{
		TotalRequests:    0,
		TotalBlocked:     0,
		TotalModified:    0,
		FilterStats:      make(map[string]*types.FilterStat),
		ViolationsByType: make(map[string]int64),
		ProcessingTime:   0,
		LastReset:        time.Now(),
	}

	// Reset individual plugin stats
	for _, plugin := range m.plugins {
		if basePlugin, ok := plugin.(*shared.BasePlugin); ok {
			basePlugin.ResetStats()
		}
	}
}