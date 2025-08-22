package filters

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	"mcp-gateway/apps/backend/internal/database/models"
	"mcp-gateway/apps/backend/internal/filters/plugins/deny"
	"mcp-gateway/apps/backend/internal/filters/plugins/pii"
	"mcp-gateway/apps/backend/internal/filters/plugins/regex"
	"mcp-gateway/apps/backend/internal/filters/plugins/resource"
	"mcp-gateway/apps/backend/internal/filters/shared"
	"mcp-gateway/apps/backend/internal/types"
)

// filterService implements FilterService interface
type filterService struct {
	db          *sql.DB
	manager     FilterManager
	registry    FilterRegistry
	mu          sync.RWMutex
	orgFilters  map[string][]Filter // Cache of organization filters
	initialized bool
}

// NewFilterService creates a new filter service
func NewFilterService(db *sql.DB) FilterService {
	return &filterService{
		db:         db,
		manager:    NewFilterManager(),
		registry:   GetGlobalRegistry(),
		orgFilters: make(map[string][]Filter),
		initialized: false,
	}
}

// Initialize sets up the filtering service
func (s *filterService) Initialize(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.initialized {
		return nil
	}

	// Register built-in filter factories
	if err := s.registerBuiltinFilters(); err != nil {
		return fmt.Errorf("failed to register builtin filters: %w", err)
	}

	s.initialized = true
	return nil
}

// ProcessContent processes content through all applicable filters
func (s *filterService) ProcessContent(ctx context.Context, filterCtx *FilterContext, content *FilterContent) (*FilterResult, *FilterContent, error) {
	// Load organization-specific filters if not cached
	if err := s.ensureOrganizationFiltersLoaded(ctx, filterCtx.OrganizationID); err != nil {
		return nil, content, fmt.Errorf("failed to load organization filters: %w", err)
	}

	// Create a temporary manager with organization-specific filters
	tempManager := NewFilterManager()
	
	s.mu.RLock()
	orgFilters := s.orgFilters[filterCtx.OrganizationID]
	s.mu.RUnlock()

	for _, filter := range orgFilters {
		if err := tempManager.AddFilter(filter); err != nil {
			return nil, content, fmt.Errorf("failed to add filter to temporary manager: %w", err)
		}
	}

	// Apply filters
	return tempManager.ApplyFilters(ctx, filterCtx, content)
}

// GetManager returns the filter manager
func (s *filterService) GetManager() FilterManager {
	return s.manager
}

// GetRegistry returns the filter registry
func (s *filterService) GetRegistry() FilterRegistry {
	return s.registry
}

// LoadFiltersFromDatabase loads filters from database configuration
func (s *filterService) LoadFiltersFromDatabase(ctx context.Context, organizationID string) error {
	query := `
		SELECT id, organization_id, name, description, type, enabled, priority, config, created_at, updated_at, created_by
		FROM content_filters 
		WHERE organization_id = $1 AND enabled = true
		ORDER BY priority ASC, created_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		return fmt.Errorf("failed to query content filters: %w", err)
	}
	defer rows.Close()

	var filters []Filter
	for rows.Next() {
		var cf models.ContentFilter
		var configJSON []byte

		err := rows.Scan(
			&cf.ID, &cf.OrganizationID, &cf.Name, &cf.Description,
			&cf.Type, &cf.Enabled, &cf.Priority, &configJSON,
			&cf.CreatedAt, &cf.UpdatedAt, &cf.CreatedBy,
		)
		if err != nil {
			return fmt.Errorf("failed to scan content filter: %w", err)
		}

		// Parse JSON config
		if err := json.Unmarshal(configJSON, &cf.Config); err != nil {
			return fmt.Errorf("failed to parse filter config for filter %s: %w", cf.Name, err)
		}

		// Convert to filter instance
		filter, err := s.createFilterFromModel(&cf)
		if err != nil {
			return fmt.Errorf("failed to create filter instance for %s: %w", cf.Name, err)
		}

		filters = append(filters, filter)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating over filter rows: %w", err)
	}

	// Cache the filters for this organization
	s.mu.Lock()
	s.orgFilters[organizationID] = filters
	s.mu.Unlock()

	return nil
}

// SaveFilterToDatabase saves a filter configuration to the database
func (s *filterService) SaveFilterToDatabase(ctx context.Context, organizationID string, filter Filter) error {
	cf := models.NewContentFilter(
		organizationID,
		filter.GetName(),
		"", // Description should be set separately
		string(filter.GetType()),
		filter.IsEnabled(),
		filter.GetPriority(),
		filter.GetConfig(),
		nil,
	)

	configJSON, err := json.Marshal(cf.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal filter config: %w", err)
	}

	query := `
		INSERT INTO content_filters (id, organization_id, name, description, type, enabled, priority, config, created_by, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		ON CONFLICT (organization_id, name) 
		DO UPDATE SET 
			description = EXCLUDED.description,
			type = EXCLUDED.type,
			enabled = EXCLUDED.enabled,
			priority = EXCLUDED.priority,
			config = EXCLUDED.config,
			updated_at = NOW()
	`

	_, err = s.db.ExecContext(ctx, query,
		cf.OrganizationID, cf.Name, cf.Description, cf.Type,
		cf.Enabled, cf.Priority, configJSON, cf.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to save content filter: %w", err)
	}

	// Invalidate cache for this organization
	s.mu.Lock()
	delete(s.orgFilters, organizationID)
	s.mu.Unlock()

	return nil
}

// DeleteFilterFromDatabase deletes a filter configuration from the database
func (s *filterService) DeleteFilterFromDatabase(ctx context.Context, organizationID, filterName string) error {
	query := `DELETE FROM content_filters WHERE organization_id = $1 AND name = $2`

	result, err := s.db.ExecContext(ctx, query, organizationID, filterName)
	if err != nil {
		return fmt.Errorf("failed to delete content filter: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("filter '%s' not found for organization '%s'", filterName, organizationID)
	}

	// Invalidate cache for this organization
	s.mu.Lock()
	delete(s.orgFilters, organizationID)
	s.mu.Unlock()

	return nil
}

// ReloadOrganizationFilters reloads filters for a specific organization
func (s *filterService) ReloadOrganizationFilters(ctx context.Context, organizationID string) error {
	// Clear cache
	s.mu.Lock()
	delete(s.orgFilters, organizationID)
	s.mu.Unlock()

	// Reload from database
	return s.LoadFiltersFromDatabase(ctx, organizationID)
}

// GetOrganizationFilters returns all filters for an organization
func (s *filterService) GetOrganizationFilters(ctx context.Context, organizationID string) ([]Filter, error) {
	if err := s.ensureOrganizationFiltersLoaded(ctx, organizationID); err != nil {
		return nil, err
	}

	s.mu.RLock()
	filters := s.orgFilters[organizationID]
	s.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]Filter, len(filters))
	copy(result, filters)
	return result, nil
}

// HealthCheck verifies the service is operational
func (s *filterService) HealthCheck(ctx context.Context) error {
	// Test database connectivity
	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database connectivity check failed: %w", err)
	}

	// Test registry
	if len(s.registry.List()) == 0 {
		return fmt.Errorf("no filter types registered")
	}

	return nil
}

// Close shuts down the service
func (s *filterService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear caches
	s.orgFilters = make(map[string][]Filter)
	s.initialized = false

	return nil
}

// GetMetrics returns filtering metrics
func (s *filterService) GetMetrics() (*types.FilteringMetrics, error) {
	return s.manager.GetStats()
}

// ensureOrganizationFiltersLoaded ensures filters are loaded for an organization
func (s *filterService) ensureOrganizationFiltersLoaded(ctx context.Context, organizationID string) error {
	s.mu.RLock()
	_, exists := s.orgFilters[organizationID]
	s.mu.RUnlock()

	if !exists {
		return s.LoadFiltersFromDatabase(ctx, organizationID)
	}

	return nil
}

// registerBuiltinFilters registers the built-in filter factories
func (s *filterService) registerBuiltinFilters() error {
	// Register PII filter factory
	piiFactory := &pii.PIIFilterFactory{}
	if err := s.registry.Register(piiFactory); err != nil {
		return fmt.Errorf("failed to register PII filter factory: %w", err)
	}

	// Register Resource filter factory
	resourceFactory := &resource.ResourceFilterFactory{}
	if err := s.registry.Register(resourceFactory); err != nil {
		return fmt.Errorf("failed to register Resource filter factory: %w", err)
	}

	// Register Deny filter factory
	denyFactory := &deny.DenyFilterFactory{}
	if err := s.registry.Register(denyFactory); err != nil {
		return fmt.Errorf("failed to register Deny filter factory: %w", err)
	}

	// Register Regex filter factory
	regexFactory := &regex.RegexFilterFactory{}
	if err := s.registry.Register(regexFactory); err != nil {
		return fmt.Errorf("failed to register Regex filter factory: %w", err)
	}

	return nil
}

// createFilterFromModel converts a database model to a Filter instance
func (s *filterService) createFilterFromModel(cf *models.ContentFilter) (Filter, error) {
	// Convert string type to FilterType
	var filterType FilterType
	switch cf.Type {
	case "pii":
		filterType = FilterTypePII
	case "resource":
		filterType = FilterTypeResource
	case "deny":
		filterType = FilterTypeDeny
	case "regex":
		filterType = FilterTypeRegex
	default:
		return nil, fmt.Errorf("unknown filter type: %s", cf.Type)
	}

	// Get the factory for this filter type
	factory, err := s.registry.Get(filterType)
	if err != nil {
		return nil, err
	}

	// Create the filter with the stored configuration
	filter, err := factory.Create(cf.Config)
	if err != nil {
		return nil, err
	}

	// Set the enabled state and priority if the filter supports it
	if baseFilter, ok := filter.(*shared.BaseFilter); ok {
		baseFilter.SetEnabled(cf.Enabled)
		baseFilter.SetPriority(cf.Priority)
	}

	return filter, nil
}

// LogViolation logs a filter violation to the database
func (s *filterService) LogViolation(ctx context.Context, violation *models.FilterViolation) error {
	query := `
		INSERT INTO filter_violations (
			id, organization_id, filter_id, request_id, session_id, server_id,
			violation_type, action_taken, content_snippet, pattern_matched, severity,
			user_id, remote_ip, user_agent, direction, metadata, created_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW()
		)
	`

	metadataJSON, err := json.Marshal(violation.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal violation metadata: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query,
		violation.OrganizationID, violation.FilterID, violation.RequestID,
		violation.SessionID, violation.ServerID, violation.ViolationType,
		violation.ActionTaken, violation.ContentSnippet, violation.PatternMatched,
		violation.Severity, violation.UserID, violation.RemoteIP,
		violation.UserAgent, violation.Direction, metadataJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to log filter violation: %w", err)
	}

	return nil
}

// GetViolations retrieves filter violations with optional filtering
func (s *filterService) GetViolations(ctx context.Context, organizationID string, limit, offset int) ([]*models.FilterViolation, error) {
	query := `
		SELECT id, organization_id, filter_id, request_id, session_id, server_id,
			   violation_type, action_taken, content_snippet, pattern_matched, severity,
			   user_id, remote_ip, user_agent, direction, metadata, created_at
		FROM filter_violations 
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, query, organizationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query filter violations: %w", err)
	}
	defer rows.Close()

	var violations []*models.FilterViolation
	for rows.Next() {
		var violation models.FilterViolation
		var metadataJSON []byte

		err := rows.Scan(
			&violation.ID, &violation.OrganizationID, &violation.FilterID,
			&violation.RequestID, &violation.SessionID, &violation.ServerID,
			&violation.ViolationType, &violation.ActionTaken, &violation.ContentSnippet,
			&violation.PatternMatched, &violation.Severity, &violation.UserID,
			&violation.RemoteIP, &violation.UserAgent, &violation.Direction,
			&metadataJSON, &violation.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan filter violation: %w", err)
		}

		// Parse JSON metadata
		if err := json.Unmarshal(metadataJSON, &violation.Metadata); err != nil {
			return nil, fmt.Errorf("failed to parse violation metadata: %w", err)
		}

		violations = append(violations, &violation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over violation rows: %w", err)
	}

	return violations, nil
}