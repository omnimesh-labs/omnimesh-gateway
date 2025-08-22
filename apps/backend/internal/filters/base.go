package filters

import (
	"mcp-gateway/apps/backend/internal/filters/shared"
)

// Re-export shared BaseFilter and helper functions for backward compatibility
type BaseFilter = shared.BaseFilter

// Re-export constructor and helper functions
var NewBaseFilter = shared.NewBaseFilter
var CreateFilterResult = shared.CreateFilterResult
var CreateFilterViolation = shared.CreateFilterViolation
var CreateFilterContext = shared.CreateFilterContext
var CreateFilterContent = shared.CreateFilterContent
var GetConfigStringSlice = shared.GetConfigStringSlice
var MergeFilterResults = shared.MergeFilterResults
var ApplyFilterWithTiming = shared.ApplyFilterWithTiming
var ValidateFilterContent = shared.ValidateFilterContent
var ValidateFilterContext = shared.ValidateFilterContext

// GetConfigValue is a generic function that needs to be called directly from shared package
// Use shared.GetConfigValue[T]() for generic type-safe configuration access