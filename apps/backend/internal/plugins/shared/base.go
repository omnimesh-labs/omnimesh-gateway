package shared

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"
)

// BasePlugin provides common functionality for all plugin implementations
type BasePlugin struct {
	config        map[string]interface{}
	stats         *types.FilterStat
	pluginType    PluginType
	name          string
	executionMode PluginExecutionMode
	capabilities  PluginCapabilities
	priority      int
	mu            sync.RWMutex
	enabled       bool
}

// NewBasePlugin creates a new base plugin instance
func NewBasePlugin(pluginType PluginType, name string, priority int) *BasePlugin {
	return &BasePlugin{
		pluginType:    pluginType,
		name:          name,
		priority:      priority,
		enabled:       true,
		executionMode: PluginModeEnforcing,
		config:        make(map[string]interface{}),
		stats: &types.FilterStat{
			Name:              name,
			Type:              string(pluginType),
			RequestsProcessed: 0,
			Violations:        0,
			Blocks:            0,
			Modifications:     0,
			AverageLatency:    0,
			LastActive:        time.Now(),
			Errors:            0,
		},
	}
}

// GetType returns the plugin type
func (b *BasePlugin) GetType() PluginType {
	return b.pluginType
}

// GetName returns the plugin name
func (b *BasePlugin) GetName() string {
	return b.name
}

// GetPriority returns the plugin priority
func (b *BasePlugin) GetPriority() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.priority
}

// SetPriority sets the plugin priority
func (b *BasePlugin) SetPriority(priority int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.priority = priority
}

// IsEnabled returns whether the plugin is enabled
func (b *BasePlugin) IsEnabled() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.enabled && b.executionMode != PluginModeDisabled
}

// SetEnabled sets the plugin enabled state
func (b *BasePlugin) SetEnabled(enabled bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.enabled = enabled
}

// GetConfig returns the current plugin configuration
func (b *BasePlugin) GetConfig() map[string]interface{} {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range b.config {
		result[k] = v
	}
	return result
}

// SetConfig sets the plugin configuration
func (b *BasePlugin) SetConfig(config map[string]interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.config == nil {
		b.config = make(map[string]interface{})
	}

	for k, v := range config {
		b.config[k] = v
	}
}

// GetCapabilities returns the plugin capabilities
func (b *BasePlugin) GetCapabilities() PluginCapabilities {
	return b.capabilities
}

// SetCapabilities sets the plugin capabilities
func (b *BasePlugin) SetCapabilities(capabilities PluginCapabilities) {
	b.capabilities = capabilities
}

// GetStats returns the plugin statistics
func (b *BasePlugin) GetStats() *types.FilterStat {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Return a copy to prevent race conditions
	statsCopy := *b.stats
	return &statsCopy
}

// GetExecutionMode returns the current execution mode
func (b *BasePlugin) GetExecutionMode() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return string(b.executionMode)
}

// SetExecutionMode sets the execution mode
func (b *BasePlugin) SetExecutionMode(mode string) error {
	executionMode := PluginExecutionMode(mode)
	if !executionMode.IsValid() {
		return fmt.Errorf("invalid execution mode: %s", mode)
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	b.executionMode = executionMode
	return nil
}

// UpdateStats updates plugin statistics
func (b *BasePlugin) UpdateStats(blocked, modified bool, violations int, latency time.Duration, hasError bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.stats.RequestsProcessed++
	b.stats.Violations += int64(violations)
	b.stats.LastActive = time.Now()

	if blocked {
		b.stats.Blocks++
	}

	if modified {
		b.stats.Modifications++
	}

	if hasError {
		b.stats.Errors++
	}

	// Update average latency using running average
	if b.stats.RequestsProcessed == 1 {
		b.stats.AverageLatency = latency
	} else {
		// Simple running average: new_avg = old_avg + (new_value - old_avg) / count
		b.stats.AverageLatency = b.stats.AverageLatency +
			(latency-b.stats.AverageLatency)/time.Duration(b.stats.RequestsProcessed)
	}
}

// ResetStats resets plugin statistics
func (b *BasePlugin) ResetStats() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.stats = &types.FilterStat{
		Name:              b.name,
		Type:              string(b.pluginType),
		RequestsProcessed: 0,
		Violations:        0,
		Blocks:            0,
		Modifications:     0,
		AverageLatency:    0,
		LastActive:        time.Now(),
		Errors:            0,
	}
}

// Apply is a placeholder implementation - actual plugins should override this
func (b *BasePlugin) Apply(ctx context.Context, pluginCtx *PluginContext, content *PluginContent) (*PluginResult, *PluginContent, error) {
	return CreatePluginResult(false, false, PluginActionAllow, "base plugin - no implementation", nil), content, nil
}

// Configure updates the plugin configuration
func (b *BasePlugin) Configure(config map[string]interface{}) error {
	b.SetConfig(config)
	return b.Validate()
}

// Validate provides basic configuration validation
func (b *BasePlugin) Validate() error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.config == nil {
		return fmt.Errorf("plugin configuration is nil")
	}

	// Check required fields that should exist in all plugins
	if _, exists := b.config["enabled"]; !exists {
		b.config["enabled"] = true
	}

	// Set execution mode if provided in config
	if mode, exists := b.config["execution_mode"]; exists {
		if modeStr, ok := mode.(string); ok {
			if err := b.SetExecutionMode(modeStr); err != nil {
				return fmt.Errorf("invalid execution mode in config: %w", err)
			}
		}
	}

	return nil
}

// CreatePluginResult creates a standard plugin result
func CreatePluginResult(blocked, modified bool, action PluginAction, reason string, violations []PluginViolation) *PluginResult {
	return &PluginResult{
		Blocked:     blocked,
		Modified:    modified,
		Action:      action,
		Reason:      reason,
		Violations:  violations,
		Metadata:    make(map[string]interface{}),
		ProcessedAt: time.Now(),
	}
}

// CreatePluginViolation creates a standard plugin violation
func CreatePluginViolation(violationType, pattern, match string, position int, severity string) PluginViolation {
	return PluginViolation{
		Type:     violationType,
		Pattern:  pattern,
		Match:    match,
		Position: position,
		Severity: severity,
		Metadata: make(map[string]interface{}),
	}
}

// CreatePluginContext creates a plugin context from request data
func CreatePluginContext(requestID, orgID, userID, serverID, sessionID string, direction PluginDirection, contentType string) *PluginContext {
	return &PluginContext{
		RequestID:      requestID,
		OrganizationID: orgID,
		UserID:         userID,
		ServerID:       serverID,
		SessionID:      sessionID,
		Direction:      direction,
		ContentType:    contentType,
		Metadata:       make(map[string]interface{}),
		Timestamp:      time.Now(),
		ExecutionMode:  string(PluginModeEnforcing),
	}
}

// CreatePluginContent creates plugin content from raw data
func CreatePluginContent(raw string, parsed interface{}, headers map[string]string, params map[string]interface{}) *PluginContent {
	content := &PluginContent{
		Raw:    raw,
		Parsed: parsed,
	}

	if headers != nil {
		content.Headers = headers
	} else {
		content.Headers = make(map[string]string)
	}

	if params != nil {
		content.Params = params
	} else {
		content.Params = make(map[string]interface{})
	}

	return content
}

// GetConfigValue safely gets a configuration value with type checking
func GetConfigValue[T any](config map[string]interface{}, key string, defaultValue T) T {
	if value, exists := config[key]; exists {
		if typedValue, ok := value.(T); ok {
			return typedValue
		}
	}
	return defaultValue
}

// GetConfigStringSlice safely gets a string slice from configuration
func GetConfigStringSlice(config map[string]interface{}, key string, defaultValue []string) []string {
	if value, exists := config[key]; exists {
		switch v := value.(type) {
		case []string:
			return v
		case []interface{}:
			result := make([]string, len(v))
			for i, item := range v {
				if str, ok := item.(string); ok {
					result[i] = str
				}
			}
			return result
		}
	}
	return defaultValue
}

// MergeFilterResults combines multiple filter results into one
func MergeFilterResults(results []*FilterResult) *FilterResult {
	if len(results) == 0 {
		return CreatePluginResult(false, false, PluginActionAllow, "", nil)
	}

	if len(results) == 1 {
		return results[0]
	}

	merged := &FilterResult{
		Blocked:     false,
		Modified:    false,
		Action:      FilterActionAllow,
		Violations:  make([]FilterViolation, 0),
		Metadata:    make(map[string]interface{}),
		ProcessedAt: time.Now(),
	}

	var reasons []string

	for _, result := range results {
		// If any filter blocks, the overall result is blocked
		if result.Blocked {
			merged.Blocked = true
			merged.Action = FilterActionBlock
		}

		// If any filter modifies, the overall result is modified
		if result.Modified {
			merged.Modified = true
		}

		// Collect all violations
		merged.Violations = append(merged.Violations, result.Violations...)

		// Collect reasons
		if result.Reason != "" {
			reasons = append(reasons, result.Reason)
		}

		// Merge metadata
		for k, v := range result.Metadata {
			merged.Metadata[k] = v
		}
	}

	// Set combined reason
	if len(reasons) > 0 {
		merged.Reason = fmt.Sprintf("Multiple filter violations: %v", reasons)
	}

	return merged
}

// ApplyFilterWithTiming applies a filter and measures timing
func ApplyFilterWithTiming(ctx context.Context, filter Filter, filterCtx *FilterContext, content *FilterContent) (*FilterResult, *FilterContent, time.Duration, error) {
	startTime := time.Now()

	result, modifiedContent, err := filter.Apply(ctx, filterCtx, content)

	duration := time.Since(startTime)

	return result, modifiedContent, duration, err
}

// ValidateFilterContent validates filter content structure
func ValidateFilterContent(content *FilterContent) error {
	if content == nil {
		return fmt.Errorf("filter content cannot be nil")
	}

	if content.Raw == "" && content.Parsed == nil {
		return fmt.Errorf("filter content must have either raw or parsed data")
	}

	return nil
}

// ValidateFilterContext validates filter context structure
func ValidateFilterContext(ctx *FilterContext) error {
	if ctx == nil {
		return fmt.Errorf("filter context cannot be nil")
	}

	if ctx.RequestID == "" {
		return fmt.Errorf("filter context must have a request ID")
	}

	if ctx.OrganizationID == "" {
		return fmt.Errorf("filter context must have an organization ID")
	}

	if !ctx.Direction.IsValid() {
		return fmt.Errorf("filter context has invalid direction: %s", ctx.Direction)
	}

	return nil
}

// MergePluginResults combines multiple plugin results into one
func MergePluginResults(results []*PluginResult) *PluginResult {
	if len(results) == 0 {
		return CreatePluginResult(false, false, PluginActionAllow, "", nil)
	}

	if len(results) == 1 {
		return results[0]
	}

	merged := &PluginResult{
		Blocked:     false,
		Modified:    false,
		Action:      PluginActionAllow,
		Violations:  make([]PluginViolation, 0),
		Metadata:    make(map[string]interface{}),
		ProcessedAt: time.Now(),
	}

	var reasons []string

	for _, result := range results {
		// If any plugin blocks, the overall result is blocked
		if result.Blocked {
			merged.Blocked = true
			merged.Action = PluginActionBlock
		}

		// If any plugin modifies, the overall result is modified
		if result.Modified {
			merged.Modified = true
		}

		// Collect all violations
		merged.Violations = append(merged.Violations, result.Violations...)

		// Collect reasons
		if result.Reason != "" {
			reasons = append(reasons, result.Reason)
		}

		// Merge metadata
		for k, v := range result.Metadata {
			merged.Metadata[k] = v
		}
	}

	// Set combined reason
	if len(reasons) > 0 {
		merged.Reason = fmt.Sprintf("Multiple plugin violations: %v", reasons)
	}

	return merged
}

// ApplyPluginWithTiming applies a plugin and measures timing
func ApplyPluginWithTiming(ctx context.Context, plugin Plugin, pluginCtx *PluginContext, content *PluginContent) (*PluginResult, *PluginContent, time.Duration, error) {
	startTime := time.Now()

	result, modifiedContent, err := plugin.Apply(ctx, pluginCtx, content)

	duration := time.Since(startTime)

	return result, modifiedContent, duration, err
}

// ValidatePluginContent validates plugin content structure
func ValidatePluginContent(content *PluginContent) error {
	if content == nil {
		return fmt.Errorf("plugin content cannot be nil")
	}

	if content.Raw == "" && content.Parsed == nil {
		return fmt.Errorf("plugin content must have either raw or parsed data")
	}

	return nil
}

// ValidatePluginContext validates plugin context structure
func ValidatePluginContext(ctx *PluginContext) error {
	if ctx == nil {
		return fmt.Errorf("plugin context cannot be nil")
	}

	if ctx.RequestID == "" {
		return fmt.Errorf("plugin context must have a request ID")
	}

	if ctx.OrganizationID == "" {
		return fmt.Errorf("plugin context must have an organization ID")
	}

	if !ctx.Direction.IsValid() {
		return fmt.Errorf("plugin context has invalid direction: %s", ctx.Direction)
	}

	return nil
}
