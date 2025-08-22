package filters

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"mcp-gateway/apps/backend/internal/types"
)

// filterManager implements FilterManager interface
type filterManager struct {
	mu      sync.RWMutex
	filters map[string]Filter
	stats   *types.FilteringMetrics
}

// NewFilterManager creates a new filter manager
func NewFilterManager() FilterManager {
	return &filterManager{
		filters: make(map[string]Filter),
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

// AddFilter adds a filter to the manager
func (m *filterManager) AddFilter(filter Filter) error {
	if filter == nil {
		return fmt.Errorf("filter cannot be nil")
	}

	name := filter.GetName()
	if name == "" {
		return fmt.Errorf("filter name cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.filters[name]; exists {
		return fmt.Errorf("filter '%s' is already registered", name)
	}

	m.filters[name] = filter
	
	// Initialize stats for this filter
	m.stats.FilterStats[name] = filter.GetStats()

	return nil
}

// RemoveFilter removes a filter from the manager
func (m *filterManager) RemoveFilter(filterName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.filters[filterName]; !exists {
		return fmt.Errorf("filter '%s' not found", filterName)
	}

	delete(m.filters, filterName)
	delete(m.stats.FilterStats, filterName)

	return nil
}

// GetFilter returns a filter by name
func (m *filterManager) GetFilter(filterName string) (Filter, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	filter, exists := m.filters[filterName]
	if !exists {
		return nil, fmt.Errorf("filter '%s' not found", filterName)
	}

	return filter, nil
}

// ListFilters returns all registered filters
func (m *filterManager) ListFilters() []Filter {
	m.mu.RLock()
	defer m.mu.RUnlock()

	filters := make([]Filter, 0, len(m.filters))
	for _, filter := range m.filters {
		filters = append(filters, filter)
	}

	return filters
}

// ApplyFilters applies all enabled filters to content
func (m *filterManager) ApplyFilters(ctx context.Context, filterCtx *FilterContext, content *FilterContent) (*FilterResult, *FilterContent, error) {
	results, modifiedContent, err := m.ApplyFiltersInOrder(ctx, filterCtx, content)
	if err != nil {
		return nil, content, err
	}

	// Merge all results into a single result
	mergedResult := MergeFilterResults(results)
	
	return mergedResult, modifiedContent, nil
}

// ApplyFiltersInOrder applies filters in priority order
func (m *filterManager) ApplyFiltersInOrder(ctx context.Context, filterCtx *FilterContext, content *FilterContent) ([]*FilterResult, *FilterContent, error) {
	// Validate inputs
	if err := ValidateFilterContext(filterCtx); err != nil {
		return nil, content, fmt.Errorf("invalid filter context: %w", err)
	}

	if err := ValidateFilterContent(content); err != nil {
		return nil, content, fmt.Errorf("invalid filter content: %w", err)
	}

	startTime := time.Now()
	defer func() {
		m.updateGlobalStats(time.Since(startTime))
	}()

	// Get enabled filters sorted by priority
	enabledFilters := m.getEnabledFiltersSorted()
	if len(enabledFilters) == 0 {
		return []*FilterResult{}, content, nil
	}

	var results []*FilterResult
	currentContent := content

	for _, filter := range enabledFilters {
		// Check if this filter supports the current direction
		capabilities := filter.GetCapabilities()
		if !m.supportsDirection(capabilities, filterCtx.Direction) {
			continue
		}

		// Check if this filter supports the content type
		if !m.supportsContentType(capabilities, filterCtx.ContentType) {
			continue
		}

		// Apply the filter with timing
		result, modifiedContent, duration, err := ApplyFilterWithTiming(ctx, filter, filterCtx, currentContent)
		if err != nil {
			// Update error stats
			if baseFilter, ok := filter.(*BaseFilter); ok {
				baseFilter.UpdateStats(false, false, 0, duration, true)
			}
			return results, currentContent, fmt.Errorf("filter '%s' failed: %w", filter.GetName(), err)
		}

		// Update filter stats
		if baseFilter, ok := filter.(*BaseFilter); ok {
			baseFilter.UpdateStats(result.Blocked, result.Modified, len(result.Violations), duration, false)
		}

		results = append(results, result)

		// If content was modified, update current content for next filter
		if result.Modified && modifiedContent != nil {
			currentContent = modifiedContent
		}

		// If any filter blocks, stop processing (short-circuit)
		if result.Blocked && result.Action == FilterActionBlock {
			break
		}
	}

	// Update global stats
	m.updateRequestStats(results)

	return results, currentContent, nil
}

// EnableFilter enables a filter by name
func (m *filterManager) EnableFilter(filterName string) error {
	filter, err := m.GetFilter(filterName)
	if err != nil {
		return err
	}

	if baseFilter, ok := filter.(*BaseFilter); ok {
		baseFilter.SetEnabled(true)
		return nil
	}

	return fmt.Errorf("filter '%s' does not support enable/disable operations", filterName)
}

// DisableFilter disables a filter by name
func (m *filterManager) DisableFilter(filterName string) error {
	filter, err := m.GetFilter(filterName)
	if err != nil {
		return err
	}

	if baseFilter, ok := filter.(*BaseFilter); ok {
		baseFilter.SetEnabled(false)
		return nil
	}

	return fmt.Errorf("filter '%s' does not support enable/disable operations", filterName)
}

// ReloadConfiguration reloads all filter configurations
func (m *filterManager) ReloadConfiguration() error {
	m.mu.RLock()
	filters := make([]Filter, 0, len(m.filters))
	for _, filter := range m.filters {
		filters = append(filters, filter)
	}
	m.mu.RUnlock()

	for _, filter := range filters {
		if err := filter.Validate(); err != nil {
			return fmt.Errorf("filter '%s' configuration validation failed: %w", filter.GetName(), err)
		}
	}

	return nil
}

// GetStats returns filtering statistics
func (m *filterManager) GetStats() (*types.FilteringMetrics, error) {
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

// getEnabledFiltersSorted returns enabled filters sorted by priority
func (m *filterManager) getEnabledFiltersSorted() []Filter {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var enabledFilters []Filter
	for _, filter := range m.filters {
		if filter.IsEnabled() {
			enabledFilters = append(enabledFilters, filter)
		}
	}

	// Sort by priority (lower number = higher priority)
	sort.Slice(enabledFilters, func(i, j int) bool {
		return enabledFilters[i].GetPriority() < enabledFilters[j].GetPriority()
	})

	return enabledFilters
}

// supportsDirection checks if filter supports the given direction
func (m *filterManager) supportsDirection(capabilities FilterCapabilities, direction FilterDirection) bool {
	switch direction {
	case FilterDirectionInbound:
		return capabilities.SupportsInbound
	case FilterDirectionOutbound:
		return capabilities.SupportsOutbound
	default:
		return false
	}
}

// supportsContentType checks if filter supports the given content type
func (m *filterManager) supportsContentType(capabilities FilterCapabilities, contentType string) bool {
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
func (m *filterManager) updateGlobalStats(processingTime time.Duration) {
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

// updateRequestStats updates statistics based on filter results
func (m *filterManager) updateRequestStats(results []*FilterResult) {
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

	// Update individual filter stats
	for filterName, filter := range m.filters {
		if filterStat := filter.GetStats(); filterStat != nil {
			m.stats.FilterStats[filterName] = filterStat
		}
	}
}

// ResetStats resets all statistics
func (m *filterManager) ResetStats() {
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

	// Reset individual filter stats
	for _, filter := range m.filters {
		if baseFilter, ok := filter.(*BaseFilter); ok {
			baseFilter.ResetStats()
		}
	}
}