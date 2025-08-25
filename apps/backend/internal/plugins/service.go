package plugins

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	"mcp-gateway/apps/backend/internal/database/models"
	"mcp-gateway/apps/backend/internal/plugins/ai_middleware/llamaguard"
	"mcp-gateway/apps/backend/internal/plugins/ai_middleware/openai_mod"
	"mcp-gateway/apps/backend/internal/plugins/content_filters/deny"
	"mcp-gateway/apps/backend/internal/plugins/content_filters/pii"
	"mcp-gateway/apps/backend/internal/plugins/content_filters/regex"
	"mcp-gateway/apps/backend/internal/plugins/content_filters/resource"
	"mcp-gateway/apps/backend/internal/plugins/shared"
	"mcp-gateway/apps/backend/internal/types"
)

// pluginService implements PluginService interface
type pluginService struct {
	manager     shared.PluginManager
	registry    shared.PluginRegistry
	db          *sql.DB
	orgPlugins  map[string][]Plugin
	mu          sync.RWMutex
	initialized bool
}

// NewPluginService creates a new plugin service
func NewPluginService(db *sql.DB) PluginService {
	return &pluginService{
		db:          db,
		manager:     NewPluginManager(),
		registry:    GetGlobalRegistry(),
		orgPlugins:  make(map[string][]Plugin),
		initialized: false,
	}
}

// Initialize sets up the plugin service
func (s *pluginService) Initialize(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.initialized {
		return nil
	}

	// Register built-in plugin factories
	if err := s.registerBuiltinFilters(); err != nil {
		return fmt.Errorf("failed to register builtin plugins: %w", err)
	}

	s.initialized = true
	return nil
}

// ProcessContent processes content through all applicable plugins
func (s *pluginService) ProcessContent(ctx context.Context, pluginCtx *PluginContext, content *PluginContent) (*PluginResult, *PluginContent, error) {
	// Load organization-specific plugins if not cached
	if err := s.ensureOrganizationFiltersLoaded(ctx, pluginCtx.OrganizationID); err != nil {
		return nil, content, fmt.Errorf("failed to load organization plugins: %w", err)
	}

	// Create a temporary manager with organization-specific plugins
	tempManager := NewPluginManager()

	s.mu.RLock()
	orgPlugins := s.orgPlugins[pluginCtx.OrganizationID]
	s.mu.RUnlock()

	for _, plugin := range orgPlugins {
		if err := tempManager.AddPlugin(plugin); err != nil {
			return nil, content, fmt.Errorf("failed to add plugin to temporary manager: %w", err)
		}
	}

	// Apply plugins
	return tempManager.ApplyPlugins(ctx, pluginCtx, content)
}

// GetManager returns the plugin manager
func (s *pluginService) GetManager() PluginManager {
	return s.manager
}

// GetRegistry returns the plugin registry
func (s *pluginService) GetRegistry() PluginRegistry {
	return s.registry
}

// LoadPluginsFromDatabase loads plugins from database configuration
func (s *pluginService) LoadPluginsFromDatabase(ctx context.Context, organizationID string) error {
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

	var filters []Plugin
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

	// Cache the plugins for this organization
	s.mu.Lock()
	s.orgPlugins[organizationID] = filters
	s.mu.Unlock()

	return nil
}

// SavePluginToDatabase saves a plugin configuration to the database
func (s *pluginService) SavePluginToDatabase(ctx context.Context, organizationID string, plugin Plugin) error {
	cf := models.NewContentFilter(
		organizationID,
		plugin.GetName(),
		"", // Description should be set separately
		string(plugin.GetType()),
		plugin.IsEnabled(),
		plugin.GetPriority(),
		plugin.GetConfig(),
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
	delete(s.orgPlugins, organizationID)
	s.mu.Unlock()

	return nil
}

// DeletePluginFromDatabase deletes a plugin configuration from the database
func (s *pluginService) DeletePluginFromDatabase(ctx context.Context, organizationID, pluginName string) error {
	query := `DELETE FROM content_filters WHERE organization_id = $1 AND name = $2`

	result, err := s.db.ExecContext(ctx, query, organizationID, pluginName)
	if err != nil {
		return fmt.Errorf("failed to delete content filter: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("filter '%s' not found for organization '%s'", pluginName, organizationID)
	}

	// Invalidate cache for this organization
	s.mu.Lock()
	delete(s.orgPlugins, organizationID)
	s.mu.Unlock()

	return nil
}

// ReloadOrganizationPlugins reloads plugins for a specific organization
func (s *pluginService) ReloadOrganizationPlugins(ctx context.Context, organizationID string) error {
	// Clear cache
	s.mu.Lock()
	delete(s.orgPlugins, organizationID)
	s.mu.Unlock()

	// Reload from database
	return s.LoadPluginsFromDatabase(ctx, organizationID)
}

// GetOrganizationPlugins returns all plugins for an organization
func (s *pluginService) GetOrganizationPlugins(ctx context.Context, organizationID string) ([]Plugin, error) {
	if err := s.ensureOrganizationFiltersLoaded(ctx, organizationID); err != nil {
		return nil, err
	}

	s.mu.RLock()
	plugins := s.orgPlugins[organizationID]
	s.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]Plugin, len(plugins))
	copy(result, plugins)
	return result, nil
}

// HealthCheck verifies the service is operational
func (s *pluginService) HealthCheck(ctx context.Context) error {
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
func (s *pluginService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear caches
	s.orgPlugins = make(map[string][]Plugin)
	s.initialized = false

	return nil
}

// GetMetrics returns filtering metrics
func (s *pluginService) GetMetrics() (*types.FilteringMetrics, error) {
	return s.manager.GetStats()
}

// ensureOrganizationFiltersLoaded ensures filters are loaded for an organization
func (s *pluginService) ensureOrganizationFiltersLoaded(ctx context.Context, organizationID string) error {
	s.mu.RLock()
	_, exists := s.orgPlugins[organizationID]
	s.mu.RUnlock()

	if !exists {
		return s.LoadPluginsFromDatabase(ctx, organizationID)
	}

	return nil
}

// registerBuiltinFilters registers the built-in filter factories
func (s *pluginService) registerBuiltinFilters() error {
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

	// Register LlamaGuard AI middleware factory
	llamaGuardFactory := &llamaguard.LlamaGuardPluginFactory{}
	if err := s.registry.Register(llamaGuardFactory); err != nil {
		return fmt.Errorf("failed to register LlamaGuard plugin factory: %w", err)
	}

	// Register OpenAI Moderation AI middleware factory
	openaiModFactory := &openai_mod.OpenAIModerationPluginFactory{}
	if err := s.registry.Register(openaiModFactory); err != nil {
		return fmt.Errorf("failed to register OpenAI Moderation plugin factory: %w", err)
	}

	return nil
}

// createFilterFromModel converts a database model to a Filter instance
func (s *pluginService) createFilterFromModel(cf *models.ContentFilter) (Plugin, error) {
	// Convert string type to PluginType (updated from FilterType)
	var pluginType shared.PluginType
	switch cf.Type {
	case "pii":
		pluginType = shared.PluginTypePII
	case "resource":
		pluginType = shared.PluginTypeResource
	case "deny":
		pluginType = shared.PluginTypeDeny
	case "regex":
		pluginType = shared.PluginTypeRegex
	case "llamaguard":
		pluginType = shared.PluginTypeLlamaGuard
	case "openai_moderation":
		pluginType = shared.PluginTypeOpenAIMod
	default:
		return nil, fmt.Errorf("unknown plugin type: %s", cf.Type)
	}

	// Get the factory for this plugin type
	factory, err := s.registry.Get(pluginType)
	if err != nil {
		return nil, err
	}

	// Create the filter with the stored configuration
	filter, err := factory.Create(cf.Config)
	if err != nil {
		return nil, err
	}

	// Set the enabled state and priority if the filter supports it
	if basePlugin, ok := filter.(*shared.BasePlugin); ok {
		basePlugin.SetEnabled(cf.Enabled)
		basePlugin.SetPriority(cf.Priority)
	}

	return filter, nil
}

// LogViolation logs a filter violation to the database
func (s *pluginService) LogViolation(ctx context.Context, violation *models.FilterViolation) error {
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
func (s *pluginService) GetViolations(ctx context.Context, organizationID string, limit, offset int) ([]interface{}, error) {
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

	// Convert to []interface{}
	result := make([]interface{}, len(violations))
	for i, v := range violations {
		result[i] = v
	}
	return result, nil
}

// ProcessContentWithDirection processes content for specific direction
func (s *pluginService) ProcessContentWithDirection(ctx context.Context, pluginCtx *PluginContext, content *PluginContent, direction PluginDirection) (*PluginResult, *PluginContent, error) {
	// Create a new context with the specified direction
	directionCtx := *pluginCtx
	directionCtx.Direction = direction

	return s.ProcessContent(ctx, &directionCtx, content)
}

// ExportPluginConfig exports plugin configuration as JSON/YAML
func (s *pluginService) ExportPluginConfig(ctx context.Context, organizationID string, format string) ([]byte, error) {
	plugins, err := s.GetOrganizationPlugins(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization plugins: %w", err)
	}

	// Create export structure
	export := make(map[string]interface{})
	export["version"] = "1.0"
	export["organization_id"] = organizationID
	export["plugins"] = plugins

	switch format {
	case "json":
		return json.Marshal(export)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// ImportPluginConfig imports plugin configuration from JSON/YAML
func (s *pluginService) ImportPluginConfig(ctx context.Context, organizationID string, configData []byte, format string) error {
	// Parse based on format
	var config map[string]interface{}
	switch format {
	case "json":
		if err := json.Unmarshal(configData, &config); err != nil {
			return fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		return fmt.Errorf("unsupported import format: %s", format)
	}

	pluginsData, ok := config["plugins"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid plugins data in config")
	}

	// For now, return an error indicating this functionality needs implementation
	_ = pluginsData
	return fmt.Errorf("plugin import functionality not yet implemented")
}
