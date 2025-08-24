package plugins

import (
	"mcp-gateway/apps/backend/internal/plugins/shared"
)

// Re-export all shared types for local use
type (
	Plugin              = shared.Plugin
	PluginFactory       = shared.PluginFactory
	PluginType          = shared.PluginType
	PluginAction        = shared.PluginAction
	PluginResult        = shared.PluginResult
	PluginViolation     = shared.PluginViolation
	PluginContext       = shared.PluginContext
	PluginDirection     = shared.PluginDirection
	PluginContent       = shared.PluginContent
	PluginCapabilities  = shared.PluginCapabilities
	PluginExecutionMode = shared.PluginExecutionMode
	PluginScope         = shared.PluginScope
	PluginInfo          = shared.PluginInfo
	PluginRegistry      = shared.PluginRegistry
	PluginManager       = shared.PluginManager
	PluginService       = shared.PluginService

	// Legacy compatibility
	Filter              = shared.Filter
	FilterFactory       = shared.FilterFactory
	FilterType          = shared.FilterType
	FilterAction        = shared.FilterAction
	FilterResult        = shared.FilterResult
	FilterViolation     = shared.FilterViolation
	FilterContext       = shared.FilterContext
	FilterDirection     = shared.FilterDirection
	FilterContent       = shared.FilterContent
	FilterCapabilities  = shared.FilterCapabilities
	FilterService       = shared.PluginService // Alias FilterService to PluginService
)

// Re-export constants
const (
	PluginTypePII        = shared.PluginTypePII
	PluginTypeResource   = shared.PluginTypeResource
	PluginTypeDeny       = shared.PluginTypeDeny
	PluginTypeRegex      = shared.PluginTypeRegex
	PluginTypeLlamaGuard = shared.PluginTypeLlamaGuard
	PluginTypeOpenAIMod  = shared.PluginTypeOpenAIMod
	PluginTypeCustomLLM  = shared.PluginTypeCustomLLM
)

const (
	PluginActionBlock = shared.PluginActionBlock
	PluginActionWarn  = shared.PluginActionWarn
	PluginActionAudit = shared.PluginActionAudit
	PluginActionAllow = shared.PluginActionAllow
)

const (
	PluginDirectionInbound  = shared.PluginDirectionInbound
	PluginDirectionOutbound = shared.PluginDirectionOutbound
	PluginDirectionPreTool  = shared.PluginDirectionPreTool
	PluginDirectionPostTool = shared.PluginDirectionPostTool
)

const (
	PluginModeEnforcing  = shared.PluginModeEnforcing
	PluginModePermissive = shared.PluginModePermissive
	PluginModeDisabled   = shared.PluginModeDisabled
	PluginModeAuditOnly  = shared.PluginModeAuditOnly
)

// Legacy filter constants for backward compatibility
const (
	FilterTypePII      = shared.FilterTypePII
	FilterTypeResource = shared.FilterTypeResource
	FilterTypeDeny     = shared.FilterTypeDeny
	FilterTypeRegex    = shared.FilterTypeRegex
)

const (
	FilterActionBlock = shared.FilterActionBlock
	FilterActionWarn  = shared.FilterActionWarn
	FilterActionAudit = shared.FilterActionAudit
	FilterActionAllow = shared.FilterActionAllow
)

const (
	FilterDirectionInbound  = shared.FilterDirectionInbound
	FilterDirectionOutbound = shared.FilterDirectionOutbound
)


