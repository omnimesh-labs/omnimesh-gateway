package pii

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/plugins/shared"
)

// PIIFilter implements the PII detection and masking filter
type PIIFilter struct {
	*shared.BasePlugin
	patterns map[string]*PIIPattern
	config   *PIIConfig
}

// PIIPattern represents a compiled PII detection pattern
type PIIPattern struct {
	Pattern     *regexp.Regexp
	Name        string
	Severity    string
	Description string
	Enabled     bool
}

// PIIConfig holds the configuration for the PII filter
type PIIConfig struct {
	Patterns        map[string]bool `json:"patterns"`
	MaskingStrategy string          `json:"masking_strategy"`
	Action          string          `json:"action"`
	CustomPatterns  []CustomPattern `json:"custom_patterns"`
	LogViolations   bool            `json:"log_violations"`
}

// CustomPattern allows users to define custom PII patterns
type CustomPattern struct {
	Name        string `json:"name"`
	Pattern     string `json:"pattern"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

// MaskingStrategy defines how PII should be masked
type MaskingStrategy string

const (
	MaskingStrategyRedact   MaskingStrategy = "redact"
	MaskingStrategyHash     MaskingStrategy = "hash"
	MaskingStrategyPartial  MaskingStrategy = "partial"
	MaskingStrategyTokenize MaskingStrategy = "tokenize"
)

// NewPIIFilter creates a new PII filter instance
func NewPIIFilter(name string, config map[string]interface{}) (*PIIFilter, error) {
	basePlugin := shared.NewBasePlugin(shared.PluginTypePII, name, 10)

	// Set capabilities
	basePlugin.SetCapabilities(shared.PluginCapabilities{
		SupportsInbound:       true,
		SupportsOutbound:      true,
		SupportsModification:  true,
		SupportsBlocking:      true,
		SupportedContentTypes: []string{"*"},
		SupportsRealtime:      true,
		SupportsBatch:         true,
	})

	filter := &PIIFilter{
		BasePlugin: basePlugin,
		patterns:   make(map[string]*PIIPattern),
	}

	if err := filter.Configure(config); err != nil {
		return nil, fmt.Errorf("failed to configure PII filter: %w", err)
	}

	return filter, nil
}

// Apply applies the PII filter to content
func (f *PIIFilter) Apply(ctx context.Context, pluginCtx *shared.PluginContext, content *shared.PluginContent) (*shared.PluginResult, *shared.PluginContent, error) {
	if !f.BasePlugin.IsEnabled() {
		return shared.CreatePluginResult(false, false, shared.PluginActionAllow, "", nil), content, nil
	}

	violations := []shared.FilterViolation{}
	modifiedContent := content.Raw
	contentModified := false

	// Check each enabled pattern
	for _, pattern := range f.patterns {
		if !pattern.Enabled {
			continue
		}

		matches := pattern.Pattern.FindAllStringSubmatch(content.Raw, -1)
		for _, match := range matches {
			if len(match) > 0 {
				// Skip if this looks like a timeout value in JSON
				if f.isTimeoutValue(content.Raw, match[0]) {
					continue
				}

				// Skip if this looks like a duration or timestamp
				if f.isDurationOrTimestamp(match[0]) {
					continue
				}

				// Find the correct index for this specific match
				matchIndex := strings.Index(content.Raw, match[0])
				violation := shared.CreatePluginViolation(
					pattern.Name,
					pattern.Pattern.String(),
					match[0],
					matchIndex,
					pattern.Severity,
				)
				violations = append(violations, violation)

				// Apply masking if configured
				if f.config.MaskingStrategy != "none" {
					maskedValue := f.maskValue(match[0], f.config.MaskingStrategy)
					modifiedContent = strings.ReplaceAll(modifiedContent, match[0], maskedValue)
					contentModified = true
				}
			}
		}
	}

	// Determine action based on violations and configuration
	var action shared.FilterAction
	var blocked bool
	var reason string

	if len(violations) > 0 {
		switch f.config.Action {
		case "block":
			action = shared.PluginActionBlock
			blocked = true
			reason = fmt.Sprintf("PII detected: %d violations found", len(violations))
		case "warn":
			action = shared.FilterActionWarn
			blocked = false
			reason = fmt.Sprintf("PII detected: %d violations found (warning)", len(violations))
		case "audit":
			action = shared.FilterActionAudit
			blocked = false
			reason = fmt.Sprintf("PII detected: %d violations logged for audit", len(violations))
		default:
			action = shared.PluginActionAllow
			blocked = false
		}
	} else {
		action = shared.PluginActionAllow
		blocked = false
	}

	result := shared.CreatePluginResult(blocked, contentModified, action, reason, violations)

	// Return modified content if applicable
	var resultContent *shared.FilterContent
	if contentModified {
		resultContent = shared.CreatePluginContent(modifiedContent, content.Parsed, content.Headers, content.Params)
	} else {
		resultContent = content
	}

	return result, resultContent, nil
}

// Configure updates the filter configuration
func (f *PIIFilter) Configure(config map[string]interface{}) error {
	// Parse configuration
	piiConfig := &PIIConfig{
		Patterns:        make(map[string]bool),
		MaskingStrategy: "redact",
		Action:          "warn",
		LogViolations:   true,
		CustomPatterns:  []CustomPattern{},
	}

	// Load patterns configuration
	if patterns, ok := config["patterns"].(map[string]interface{}); ok {
		for name, enabled := range patterns {
			if enabledBool, ok := enabled.(bool); ok {
				piiConfig.Patterns[name] = enabledBool
			}
		}
	}

	// Load masking strategy
	if strategy, ok := config["masking_strategy"].(string); ok {
		piiConfig.MaskingStrategy = strategy
	}

	// Load action
	if action, ok := config["action"].(string); ok {
		piiConfig.Action = action
	}

	// Load log violations setting
	if logViolations, ok := config["log_violations"].(bool); ok {
		piiConfig.LogViolations = logViolations
	}

	// Load custom patterns
	if customPatterns, ok := config["custom_patterns"].([]interface{}); ok {
		for _, cp := range customPatterns {
			if cpMap, ok := cp.(map[string]interface{}); ok {
				customPattern := CustomPattern{
					Name:        shared.GetConfigValue(cpMap, "name", ""),
					Pattern:     shared.GetConfigValue(cpMap, "pattern", ""),
					Enabled:     shared.GetConfigValue(cpMap, "enabled", true),
					Severity:    shared.GetConfigValue(cpMap, "severity", "medium"),
					Description: shared.GetConfigValue(cpMap, "description", ""),
				}
				if customPattern.Name != "" && customPattern.Pattern != "" {
					piiConfig.CustomPatterns = append(piiConfig.CustomPatterns, customPattern)
				}
			}
		}
	}

	f.config = piiConfig
	f.BasePlugin.SetConfig(config)

	// Compile patterns
	return f.compilePatterns()
}

// compilePatterns compiles all enabled PII detection patterns
func (f *PIIFilter) compilePatterns() error {
	f.patterns = make(map[string]*PIIPattern)

	// Built-in patterns
	builtinPatterns := map[string]string{
		"ssn":         `\b\d{3}-\d{2}-\d{4}\b|\b\d{9}\b`,
		"credit_card": `\b4[0-9]{12}(?:[0-9]{3})?\b|\b5[1-5][0-9]{14}\b|\b3[47][0-9]{13}\b|\b3[0-9]{13}\b|\b6(?:011|5[0-9]{2})[0-9]{12}\b`,
		"email":       `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`,
		"phone":       `\b(?:\+?1[-.\s]?)?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}\b`,
		"aws_keys":    `AKIA[0-9A-Z]{16}|aws_access_key_id\s*=\s*[A-Z0-9]{20}`,
		"ip_address":  `\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`,
		"api_key":     `\b[A-Za-z0-9]{32,}\b`,
	}

	// Compile built-in patterns
	for name, pattern := range builtinPatterns {
		if enabled, exists := f.config.Patterns[name]; exists && enabled {
			compiled, err := regexp.Compile(pattern)
			if err != nil {
				return fmt.Errorf("failed to compile pattern %s: %w", name, err)
			}

			f.patterns[name] = &PIIPattern{
				Name:        name,
				Pattern:     compiled,
				Enabled:     true,
				Severity:    "medium",
				Description: fmt.Sprintf("Built-in %s detection pattern", name),
			}
		}
	}

	// Compile custom patterns
	for _, customPattern := range f.config.CustomPatterns {
		if customPattern.Enabled && customPattern.Pattern != "" {
			compiled, err := regexp.Compile(customPattern.Pattern)
			if err != nil {
				return fmt.Errorf("failed to compile custom pattern %s: %w", customPattern.Name, err)
			}

			f.patterns[customPattern.Name] = &PIIPattern{
				Name:        customPattern.Name,
				Pattern:     compiled,
				Enabled:     true,
				Severity:    customPattern.Severity,
				Description: customPattern.Description,
			}
		}
	}

	return nil
}

// maskValue masks a detected PII value based on the configured strategy
func (f *PIIFilter) maskValue(value, strategy string) string {
	switch MaskingStrategy(strategy) {
	case MaskingStrategyRedact:
		return "[REDACTED]"
	case MaskingStrategyHash:
		return fmt.Sprintf("[HASH:%x]", len(value)) // Simple hash representation
	case MaskingStrategyPartial:
		if len(value) <= 4 {
			return strings.Repeat("*", len(value))
		}
		return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
	case MaskingStrategyTokenize:
		return fmt.Sprintf("[TOKEN:%x]", len(value)) // Token representation
	default:
		return "[MASKED]"
	}
}

// PIIFilterFactory implements FilterFactory for PII filters
type PIIFilterFactory struct{}

// Create creates a new PII filter instance
func (f *PIIFilterFactory) Create(config map[string]interface{}) (shared.Filter, error) {
	name := shared.GetConfigValue(config, "name", "pii-filter")
	return NewPIIFilter(name, config)
}

// GetType returns the filter type
func (f *PIIFilterFactory) GetType() shared.FilterType {
	return shared.FilterTypePII
}

// GetName returns the factory name
func (f *PIIFilterFactory) GetName() string {
	return "PII Filter"
}

// GetDescription returns the factory description
func (f *PIIFilterFactory) GetDescription() string {
	return "Detects and masks personally identifiable information (PII) in content"
}

// GetSupportedExecutionModes returns supported execution modes
func (f *PIIFilterFactory) GetSupportedExecutionModes() []string {
	return []string{
		string(shared.PluginModeEnforcing),
		string(shared.PluginModePermissive),
		string(shared.PluginModeDisabled),
		string(shared.PluginModeAuditOnly),
	}
}

// ValidateConfig validates the configuration for PII filters
func (f *PIIFilterFactory) ValidateConfig(config map[string]interface{}) error {
	// Validate masking strategy
	if strategy, ok := config["masking_strategy"].(string); ok {
		validStrategies := []string{"redact", "hash", "partial", "tokenize", "none"}
		valid := false
		for _, vs := range validStrategies {
			if strategy == vs {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid masking strategy: %s", strategy)
		}
	}

	// Validate action
	if action, ok := config["action"].(string); ok {
		validActions := []string{"block", "warn", "audit", "allow"}
		valid := false
		for _, va := range validActions {
			if action == va {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid action: %s", action)
		}
	}

	// Validate custom patterns
	if customPatterns, ok := config["custom_patterns"].([]interface{}); ok {
		for i, cp := range customPatterns {
			if cpMap, ok := cp.(map[string]interface{}); ok {
				if pattern, exists := cpMap["pattern"].(string); exists && pattern != "" {
					if _, err := regexp.Compile(pattern); err != nil {
						return fmt.Errorf("invalid regex pattern in custom_patterns[%d]: %w", i, err)
					}
				}
			}
		}
	}

	return nil
}

// GetDefaultConfig returns the default configuration for PII filters
func (f *PIIFilterFactory) GetDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"patterns": map[string]bool{
			"ssn":         true,
			"credit_card": true,
			"email":       true,
			"phone":       true,
			"aws_keys":    true,
			"ip_address":  false,
		},
		"masking_strategy": "redact",
		"action":           "warn",
		"log_violations":   true,
		"custom_patterns":  []CustomPattern{},
	}
}

// GetConfigSchema returns the JSON schema for configuration validation
func (f *PIIFilterFactory) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"patterns": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"ssn":         map[string]interface{}{"type": "boolean"},
					"credit_card": map[string]interface{}{"type": "boolean"},
					"email":       map[string]interface{}{"type": "boolean"},
					"phone":       map[string]interface{}{"type": "boolean"},
					"aws_keys":    map[string]interface{}{"type": "boolean"},
					"ip_address":  map[string]interface{}{"type": "boolean"},
				},
			},
			"masking_strategy": map[string]interface{}{
				"type": "string",
				"enum": []string{"redact", "hash", "partial", "tokenize", "none"},
			},
			"action": map[string]interface{}{
				"type": "string",
				"enum": []string{"block", "warn", "audit", "allow"},
			},
			"log_violations": map[string]interface{}{"type": "boolean"},
			"custom_patterns": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name":        map[string]interface{}{"type": "string"},
						"pattern":     map[string]interface{}{"type": "string"},
						"enabled":     map[string]interface{}{"type": "boolean"},
						"severity":    map[string]interface{}{"type": "string"},
						"description": map[string]interface{}{"type": "string"},
					},
					"required": []string{"name", "pattern"},
				},
			},
		},
	}
}

// isTimeoutValue checks if a matched value appears to be a timeout in JSON context
func (f *PIIFilter) isTimeoutValue(content, value string) bool {
	// Look for JSON patterns like "timeout": <value>
	timeoutPatterns := []string{
		`"timeout":\s*` + regexp.QuoteMeta(value),
		`"timeout_seconds":\s*` + regexp.QuoteMeta(value),
		`"timeout_ms":\s*` + regexp.QuoteMeta(value),
		`"timeout_ns":\s*` + regexp.QuoteMeta(value),
		`"duration":\s*` + regexp.QuoteMeta(value),
		`"delay":\s*` + regexp.QuoteMeta(value),
		`"wait_time":\s*` + regexp.QuoteMeta(value),
	}

	for _, pattern := range timeoutPatterns {
		matched, _ := regexp.MatchString(pattern, content)
		if matched {
			return true
		}
	}

	return false
}

// isDurationOrTimestamp checks if a value looks like a duration or timestamp
func (f *PIIFilter) isDurationOrTimestamp(value string) bool {
	// Only match very specific timestamp patterns to avoid false positives
	// Credit cards are 13-19 digits, so we need to be careful

	// Unix timestamps in seconds (exactly 10 digits, starting with 1 for current era)
	if matched, _ := regexp.MatchString(`^1\d{9}$`, value); matched {
		return true
	}

	// Very large durations in nanoseconds (20+ digits only)
	if matched, _ := regexp.MatchString(`^\d{20,}$`, value); matched {
		return true
	}

	return false
}
