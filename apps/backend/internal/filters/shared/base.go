package shared

import (
	"context"
	"fmt"
	"sync"
	"time"

	"mcp-gateway/apps/backend/internal/types"
)

// BaseFilter provides common functionality for all filter implementations
type BaseFilter struct {
	filterType   FilterType
	name         string
	priority     int
	enabled      bool
	config       map[string]interface{}
	capabilities FilterCapabilities
	mu           sync.RWMutex
	stats        *types.FilterStat
}

// NewBaseFilter creates a new base filter instance
func NewBaseFilter(filterType FilterType, name string, priority int) *BaseFilter {
	return &BaseFilter{
		filterType: filterType,
		name:       name,
		priority:   priority,
		enabled:    true,
		config:     make(map[string]interface{}),
		stats: &types.FilterStat{
			Name:              name,
			Type:              string(filterType),
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

// GetType returns the filter type
func (b *BaseFilter) GetType() FilterType {
	return b.filterType
}

// GetName returns the filter name
func (b *BaseFilter) GetName() string {
	return b.name
}

// GetPriority returns the filter priority
func (b *BaseFilter) GetPriority() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.priority
}

// SetPriority sets the filter priority
func (b *BaseFilter) SetPriority(priority int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.priority = priority
}

// IsEnabled returns whether the filter is enabled
func (b *BaseFilter) IsEnabled() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.enabled
}

// SetEnabled sets the filter enabled state
func (b *BaseFilter) SetEnabled(enabled bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.enabled = enabled
}

// GetConfig returns the current filter configuration
func (b *BaseFilter) GetConfig() map[string]interface{} {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range b.config {
		result[k] = v
	}
	return result
}

// SetConfig sets the filter configuration
func (b *BaseFilter) SetConfig(config map[string]interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.config == nil {
		b.config = make(map[string]interface{})
	}

	for k, v := range config {
		b.config[k] = v
	}
}

// GetCapabilities returns the filter capabilities
func (b *BaseFilter) GetCapabilities() FilterCapabilities {
	return b.capabilities
}

// SetCapabilities sets the filter capabilities
func (b *BaseFilter) SetCapabilities(capabilities FilterCapabilities) {
	b.capabilities = capabilities
}

// GetStats returns the filter statistics
func (b *BaseFilter) GetStats() *types.FilterStat {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Return a copy to prevent race conditions
	statsCopy := *b.stats
	return &statsCopy
}

// UpdateStats updates filter statistics
func (b *BaseFilter) UpdateStats(blocked, modified bool, violations int, latency time.Duration, hasError bool) {
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

// ResetStats resets filter statistics
func (b *BaseFilter) ResetStats() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.stats = &types.FilterStat{
		Name:              b.name,
		Type:              string(b.filterType),
		RequestsProcessed: 0,
		Violations:        0,
		Blocks:            0,
		Modifications:     0,
		AverageLatency:    0,
		LastActive:        time.Now(),
		Errors:            0,
	}
}

// Apply is a placeholder implementation - actual filters should override this
func (b *BaseFilter) Apply(ctx context.Context, filterCtx *FilterContext, content *FilterContent) (*FilterResult, *FilterContent, error) {
	return CreateFilterResult(false, false, FilterActionAllow, "base filter - no implementation", nil), content, nil
}

// Configure updates the filter configuration
func (b *BaseFilter) Configure(config map[string]interface{}) error {
	b.SetConfig(config)
	return b.Validate()
}

// Validate provides basic configuration validation
func (b *BaseFilter) Validate() error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.config == nil {
		return fmt.Errorf("filter configuration is nil")
	}

	// Check required fields that should exist in all filters
	if _, exists := b.config["enabled"]; !exists {
		b.config["enabled"] = true
	}

	return nil
}

// CreateFilterResult creates a standard filter result
func CreateFilterResult(blocked, modified bool, action FilterAction, reason string, violations []FilterViolation) *FilterResult {
	return &FilterResult{
		Blocked:     blocked,
		Modified:    modified,
		Action:      action,
		Reason:      reason,
		Violations:  violations,
		Metadata:    make(map[string]interface{}),
		ProcessedAt: time.Now(),
	}
}

// CreateFilterViolation creates a standard filter violation
func CreateFilterViolation(violationType, pattern, match string, position int, severity string) FilterViolation {
	return FilterViolation{
		Type:     violationType,
		Pattern:  pattern,
		Match:    match,
		Position: position,
		Severity: severity,
		Metadata: make(map[string]interface{}),
	}
}

// CreateFilterContext creates a filter context from request data
func CreateFilterContext(requestID, orgID, userID, serverID, sessionID string, direction FilterDirection, contentType string) *FilterContext {
	return &FilterContext{
		RequestID:      requestID,
		OrganizationID: orgID,
		UserID:         userID,
		ServerID:       serverID,
		SessionID:      sessionID,
		Direction:      direction,
		ContentType:    contentType,
		Metadata:       make(map[string]interface{}),
		Timestamp:      time.Now(),
	}
}

// CreateFilterContent creates filter content from raw data
func CreateFilterContent(raw string, parsed interface{}, headers map[string]string, params map[string]interface{}) *FilterContent {
	content := &FilterContent{
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
		return CreateFilterResult(false, false, FilterActionAllow, "", nil)
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
