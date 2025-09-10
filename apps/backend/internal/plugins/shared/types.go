package shared

import (
	"context"
	"time"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"
)

// PluginType represents the type of content plugin
type PluginType string

// Legacy alias for backward compatibility
type FilterType = PluginType

const (
	// Content filtering plugins
	PluginTypePII      PluginType = "pii"
	PluginTypeResource PluginType = "resource"
	PluginTypeDeny     PluginType = "deny"
	PluginTypeRegex    PluginType = "regex"

	// AI Middleware plugins
	PluginTypeLlamaGuard PluginType = "llamaguard"
	PluginTypeOpenAIMod  PluginType = "openai_moderation"
	PluginTypeCustomLLM  PluginType = "custom_llm"

	// Legacy aliases for backward compatibility
	FilterTypePII      = PluginTypePII
	FilterTypeResource = PluginTypeResource
	FilterTypeDeny     = PluginTypeDeny
	FilterTypeRegex    = PluginTypeRegex
)

// PluginAction represents the action to take when plugin is triggered
type PluginAction string

// Legacy alias for backward compatibility
type FilterAction = PluginAction

const (
	PluginActionBlock PluginAction = "block"
	PluginActionWarn  PluginAction = "warn"
	PluginActionAudit PluginAction = "audit"
	PluginActionAllow PluginAction = "allow"

	// Legacy aliases for backward compatibility
	FilterActionBlock = PluginActionBlock
	FilterActionWarn  = PluginActionWarn
	FilterActionAudit = PluginActionAudit
	FilterActionAllow = PluginActionAllow
)

// PluginDirection indicates whether content is incoming or outgoing
type PluginDirection string

// Legacy alias for backward compatibility
type FilterDirection = PluginDirection

const (
	PluginDirectionInbound  PluginDirection = "inbound"   // User -> MCP Server
	PluginDirectionOutbound PluginDirection = "outbound"  // MCP Server -> User
	PluginDirectionPreTool  PluginDirection = "pre_tool"  // Before tool execution
	PluginDirectionPostTool PluginDirection = "post_tool" // After tool execution

	// Legacy aliases for backward compatibility
	FilterDirectionInbound  = PluginDirectionInbound
	FilterDirectionOutbound = PluginDirectionOutbound
)

// PluginResult represents the result of applying a plugin
type PluginResult struct {
	ProcessedAt time.Time              `json:"processed_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Action      PluginAction           `json:"action"`
	Reason      string                 `json:"reason,omitempty"`
	PluginName  string                 `json:"plugin_name,omitempty"`
	PluginType  PluginType             `json:"plugin_type,omitempty"`
	Violations  []PluginViolation      `json:"violations,omitempty"`
	Blocked     bool                   `json:"blocked"`
	Modified    bool                   `json:"modified"`
}

// Legacy alias for backward compatibility
type FilterResult = PluginResult

// PluginViolation represents a specific violation found by a plugin
type PluginViolation struct {
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Type        string                 `json:"type"`
	Pattern     string                 `json:"pattern,omitempty"`
	Match       string                 `json:"match,omitempty"`
	Severity    string                 `json:"severity,omitempty"`
	Replacement string                 `json:"replacement,omitempty"`
	Category    string                 `json:"category,omitempty"`
	Position    int                    `json:"position,omitempty"`
	Confidence  float64                `json:"confidence,omitempty"`
}

// Legacy alias for backward compatibility
type FilterViolation = PluginViolation

// PluginContext carries data through the plugin chain
type PluginContext struct {
	RequestID      string                 `json:"request_id"`
	OrganizationID string                 `json:"organization_id"`
	UserID         string                 `json:"user_id,omitempty"`
	ServerID       string                 `json:"server_id,omitempty"`
	SessionID      string                 `json:"session_id,omitempty"`
	Transport      types.TransportType    `json:"transport"`
	Direction      PluginDirection        `json:"direction"`
	ContentType    string                 `json:"content_type"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`

	// Enhanced context for AI middleware
	ExecutionMode string `json:"execution_mode,omitempty"` // enforcing, permissive, disabled
	ToolName      string `json:"tool_name,omitempty"`      // Current tool being executed
	VirtualServer string `json:"virtual_server,omitempty"` // Virtual server ID
}

// Legacy alias for backward compatibility
type FilterContext = PluginContext

// PluginContent represents content to be processed by plugins
type PluginContent struct {
	Parsed     interface{}            `json:"parsed,omitempty"`
	Headers    map[string]string      `json:"headers,omitempty"`
	Params     map[string]interface{} `json:"params,omitempty"`
	Raw        string                 `json:"raw"`
	Language   string                 `json:"language,omitempty"`
	Tokens     []string               `json:"tokens,omitempty"`
	Embeddings []float64              `json:"embeddings,omitempty"`
}

// Legacy alias for backward compatibility
type FilterContent = PluginContent

// PluginCapabilities describes what features a plugin supports
type PluginCapabilities struct {
	SupportedContentTypes []string `json:"supported_content_types"`
	SupportedLanguages    []string `json:"supported_languages,omitempty"`
	SupportsPostTool      bool     `json:"supports_post_tool"`
	SupportsInbound       bool     `json:"supports_inbound"`
	SupportsModification  bool     `json:"supports_modification"`
	SupportsBlocking      bool     `json:"supports_blocking"`
	SupportsPreTool       bool     `json:"supports_pre_tool"`
	SupportsRealtime      bool     `json:"supports_realtime"`
	SupportsBatch         bool     `json:"supports_batch"`
	RequiresExternalAPI   bool     `json:"requires_external_api"`
	SupportsStreaming     bool     `json:"supports_streaming"`
	SupportsTokenization  bool     `json:"supports_tokenization"`
	SupportsOutbound      bool     `json:"supports_outbound"`
}

// Legacy alias for backward compatibility
type FilterCapabilities = PluginCapabilities

// Helper functions for plugin types
func (p PluginType) IsValid() bool {
	switch p {
	case PluginTypePII, PluginTypeResource, PluginTypeDeny, PluginTypeRegex,
		PluginTypeLlamaGuard, PluginTypeOpenAIMod, PluginTypeCustomLLM:
		return true
	default:
		return false
	}
}

func (p PluginType) String() string {
	return string(p)
}

// IsAIPlugin returns true if this is an AI middleware plugin
func (p PluginType) IsAIPlugin() bool {
	switch p {
	case PluginTypeLlamaGuard, PluginTypeOpenAIMod, PluginTypeCustomLLM:
		return true
	default:
		return false
	}
}

// Helper functions for plugin actions
func (a PluginAction) IsValid() bool {
	switch a {
	case PluginActionBlock, PluginActionWarn, PluginActionAudit, PluginActionAllow:
		return true
	default:
		return false
	}
}

func (a PluginAction) String() string {
	return string(a)
}

// Helper functions for plugin directions
func (d PluginDirection) IsValid() bool {
	switch d {
	case PluginDirectionInbound, PluginDirectionOutbound,
		PluginDirectionPreTool, PluginDirectionPostTool:
		return true
	default:
		return false
	}
}

func (d PluginDirection) String() string {
	return string(d)
}

// Plugin defines the interface for content plugins
type Plugin interface {
	// GetType returns the plugin type
	GetType() PluginType

	// GetName returns the plugin name/identifier
	GetName() string

	// GetPriority returns the plugin priority (lower = higher priority)
	GetPriority() int

	// IsEnabled returns whether the plugin is currently enabled
	IsEnabled() bool

	// Apply applies the plugin to the given content
	Apply(ctx context.Context, pluginCtx *PluginContext, content *PluginContent) (*PluginResult, *PluginContent, error)

	// Configure updates the plugin configuration
	Configure(config map[string]interface{}) error

	// GetConfig returns the current plugin configuration
	GetConfig() map[string]interface{}

	// Validate validates the plugin configuration
	Validate() error

	// GetCapabilities returns what features this plugin supports
	GetCapabilities() PluginCapabilities

	// GetStats returns plugin statistics
	GetStats() *types.FilterStat

	// GetExecutionMode returns the current execution mode
	GetExecutionMode() string

	// SetExecutionMode sets the execution mode (enforcing, permissive, disabled)
	SetExecutionMode(mode string) error
}

// Legacy alias for backward compatibility
type Filter = Plugin

// PluginFactory creates plugin instances
type PluginFactory interface {
	// Create returns a new plugin instance
	Create(config map[string]interface{}) (Plugin, error)

	// GetType returns the plugin type this factory creates
	GetType() PluginType

	// GetName returns the factory name
	GetName() string

	// GetDescription returns a human-readable description
	GetDescription() string

	// ValidateConfig validates the configuration for this plugin type
	ValidateConfig(config map[string]interface{}) error

	// GetDefaultConfig returns the default configuration
	GetDefaultConfig() map[string]interface{}

	// GetConfigSchema returns the JSON schema for configuration validation
	GetConfigSchema() map[string]interface{}

	// GetSupportedExecutionModes returns supported execution modes
	GetSupportedExecutionModes() []string
}

// Legacy alias for backward compatibility
type FilterFactory = PluginFactory

// PluginExecutionMode defines how plugins are executed
type PluginExecutionMode string

const (
	PluginModeEnforcing  PluginExecutionMode = "enforcing"  // Block on violations
	PluginModePermissive PluginExecutionMode = "permissive" // Log but allow violations
	PluginModeDisabled   PluginExecutionMode = "disabled"   // Plugin is disabled
	PluginModeAuditOnly  PluginExecutionMode = "audit_only" // Only log for audit
)

// Helper functions for execution modes

// PluginService defines the main plugin service interface
type PluginService interface {
	// Initialize sets up the plugin service
	Initialize(ctx context.Context) error

	// ProcessContent processes content through all applicable plugins
	ProcessContent(ctx context.Context, pluginCtx *PluginContext, content *PluginContent) (*PluginResult, *PluginContent, error)

	// ProcessContentWithDirection processes content for specific direction
	ProcessContentWithDirection(ctx context.Context, pluginCtx *PluginContext, content *PluginContent, direction PluginDirection) (*PluginResult, *PluginContent, error)

	// GetManager returns the plugin manager
	GetManager() PluginManager

	// GetRegistry returns the plugin registry
	GetRegistry() PluginRegistry

	// LoadPluginsFromDatabase loads plugins from database configuration
	LoadPluginsFromDatabase(ctx context.Context, organizationID string) error

	// SavePluginToDatabase saves a plugin configuration to the database
	SavePluginToDatabase(ctx context.Context, organizationID string, plugin Plugin) error

	// DeletePluginFromDatabase deletes a plugin configuration from the database
	DeletePluginFromDatabase(ctx context.Context, organizationID, pluginName string) error

	// ReloadOrganizationPlugins reloads plugins for a specific organization
	ReloadOrganizationPlugins(ctx context.Context, organizationID string) error

	// GetOrganizationPlugins returns all plugins for an organization
	GetOrganizationPlugins(ctx context.Context, organizationID string) ([]Plugin, error)

	// ExportPluginConfig exports plugin configuration as JSON/YAML
	ExportPluginConfig(ctx context.Context, organizationID string, format string) ([]byte, error)

	// ImportPluginConfig imports plugin configuration from JSON/YAML
	ImportPluginConfig(ctx context.Context, organizationID string, configData []byte, format string) error

	// HealthCheck verifies the service is operational
	HealthCheck(ctx context.Context) error

	// Close shuts down the service
	Close() error

	// GetMetrics returns plugin metrics
	GetMetrics() (*types.FilteringMetrics, error)

	// GetViolations retrieves plugin violations with optional filtering
	GetViolations(ctx context.Context, organizationID string, limit, offset int) ([]interface{}, error)
}

// PluginRegistry manages available plugin factories
type PluginRegistry interface {
	// Register adds a plugin factory to the registry
	Register(factory PluginFactory) error

	// Get returns a plugin factory by type
	Get(pluginType PluginType) (PluginFactory, error)

	// List returns all available plugin types
	List() []PluginType

	// GetInfo returns information about a plugin type
	GetInfo(pluginType PluginType) (*PluginInfo, error)

	// GetAllInfo returns information about all registered plugins
	GetAllInfo() ([]*PluginInfo, error)

	// GetByCategory returns plugins of a specific category (AI, content filter, etc.)
	GetByCategory(category string) []PluginFactory
}

// PluginInfo contains metadata about a plugin type
type PluginInfo struct {
	ConfigSchema   map[string]any     `json:"config_schema,omitempty"`
	DefaultConfig  map[string]any     `json:"default_config,omitempty"`
	Type           PluginType         `json:"type"`
	Name           string             `json:"name"`
	Description    string             `json:"description"`
	Version        string             `json:"version"`
	Author         string             `json:"author"`
	Category       string             `json:"category,omitempty"`
	SupportedModes []string           `json:"supported_modes,omitempty"`
	Capabilities   PluginCapabilities `json:"capabilities"`
	RequiresAPI    bool               `json:"requires_api,omitempty"`
}

// PluginManager orchestrates multiple plugins with enhanced lifecycle support
type PluginManager interface {
	// AddPlugin adds a plugin to the manager
	AddPlugin(plugin Plugin) error

	// RemovePlugin removes a plugin from the manager
	RemovePlugin(pluginName string) error

	// GetPlugin returns a plugin by name
	GetPlugin(pluginName string) (Plugin, error)

	// ListPlugins returns all registered plugins
	ListPlugins() []Plugin

	// ApplyPlugins applies all enabled plugins to content
	ApplyPlugins(ctx context.Context, pluginCtx *PluginContext, content *PluginContent) (*PluginResult, *PluginContent, error)

	// ApplyPluginsInOrder applies plugins in priority order
	ApplyPluginsInOrder(ctx context.Context, pluginCtx *PluginContext, content *PluginContent) ([]*PluginResult, *PluginContent, error)

	// ApplyPluginsForDirection applies plugins for specific direction (inbound, outbound, pre-tool, post-tool)
	ApplyPluginsForDirection(ctx context.Context, pluginCtx *PluginContext, content *PluginContent, direction PluginDirection) ([]*PluginResult, *PluginContent, error)

	// EnablePlugin enables a plugin by name
	EnablePlugin(pluginName string) error

	// DisablePlugin disables a plugin by name
	DisablePlugin(pluginName string) error

	// SetPluginExecutionMode sets execution mode for a plugin
	SetPluginExecutionMode(pluginName string, mode PluginExecutionMode) error

	// ReloadConfiguration reloads all plugin configurations
	ReloadConfiguration() error

	// GetStats returns plugin statistics
	GetStats() (*types.FilteringMetrics, error)
}

func (m PluginExecutionMode) IsValid() bool {
	switch m {
	case PluginModeEnforcing, PluginModePermissive, PluginModeDisabled, PluginModeAuditOnly:
		return true
	default:
		return false
	}
}

func (m PluginExecutionMode) String() string {
	return string(m)
}

// PluginScope defines the scope where plugins apply
type PluginScope string

const (
	PluginScopeGlobal        PluginScope = "global"         // Apply to all requests
	PluginScopeUser          PluginScope = "user"           // Per-user configuration
	PluginScopeVirtualServer PluginScope = "virtual_server" // Per virtual server
	PluginScopeTool          PluginScope = "tool"           // Per tool
	PluginScopeTenant        PluginScope = "tenant"         // Per tenant/organization
)

// Helper functions for plugin scopes
func (s PluginScope) IsValid() bool {
	switch s {
	case PluginScopeGlobal, PluginScopeUser, PluginScopeVirtualServer, PluginScopeTool, PluginScopeTenant:
		return true
	default:
		return false
	}
}

func (s PluginScope) String() string {
	return string(s)
}
