package regex

import (
	"context"
	"fmt"
	"regexp"

	"mcp-gateway/apps/backend/internal/plugins/shared"
)

// RegexFilter implements pattern-based search and replace functionality
type RegexFilter struct {
	*shared.BasePlugin
	config *RegexConfig
	rules  []*RegexRule
}

// RegexConfig holds the configuration for the Regex filter
type RegexConfig struct {
	Rules         []Rule `json:"rules"`
	Action        string `json:"action"`
	LogViolations bool   `json:"log_violations"`
	LogMatches    bool   `json:"log_matches"`
}

// Rule represents a single regex rule
type Rule struct {
	Name        string `json:"name"`
	Pattern     string `json:"pattern"`
	Replacement string `json:"replacement"`
	Enabled     bool   `json:"enabled"`
	Severity    string `json:"severity"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Action      string `json:"action"` // "replace", "block", "warn", "audit"
}

// RegexRule represents a compiled regex rule
type RegexRule struct {
	Name        string
	Pattern     *regexp.Regexp
	Replacement string
	Enabled     bool
	Severity    string
	Category    string
	Description string
	Action      string
}

// NewRegexFilter creates a new Regex filter instance
func NewRegexFilter(name string, config map[string]interface{}) (*RegexFilter, error) {
	basePlugin := shared.NewBasePlugin(shared.PluginTypeRegex, name, 40)

	// Set capabilities
	basePlugin.SetCapabilities(shared.PluginCapabilities{
		SupportsInbound:       true,
		SupportsOutbound:      true,
		SupportsModification:  true, // Regex filter can modify content
		SupportsBlocking:      true,
		SupportedContentTypes: []string{"*"},
		SupportsRealtime:      true,
		SupportsBatch:         true,
	})

	filter := &RegexFilter{
		BasePlugin: basePlugin,
	}

	if err := filter.Configure(config); err != nil {
		return nil, fmt.Errorf("failed to configure Regex filter: %w", err)
	}

	return filter, nil
}

// Apply applies the Regex filter to content
func (f *RegexFilter) Apply(ctx context.Context, pluginCtx *shared.PluginContext, content *shared.PluginContent) (*shared.PluginResult, *shared.PluginContent, error) {
	if !f.BasePlugin.IsEnabled() {
		return shared.CreatePluginResult(false, false, shared.PluginActionAllow, "", nil), content, nil
	}

	violations := []shared.FilterViolation{}
	modifiedContent := content.Raw
	contentModified := false
	shouldBlock := false

	// Apply each enabled rule
	for _, rule := range f.rules {
		if !rule.Enabled {
			continue
		}

		matches := rule.Pattern.FindAllStringSubmatch(modifiedContent, -1)
		matchIndices := rule.Pattern.FindAllStringSubmatchIndex(modifiedContent, -1)

		for i, match := range matches {
			if len(match) > 0 {
				// Create violation record
				var position int
				if i < len(matchIndices) && len(matchIndices[i]) >= 2 {
					position = matchIndices[i][0]
				}

				violation := shared.CreatePluginViolation(
					"regex_match",
					rule.Pattern.String(),
					match[0],
					position,
					rule.Severity,
				)
				violation.Metadata["rule_name"] = rule.Name
				violation.Metadata["category"] = rule.Category
				violation.Metadata["action"] = rule.Action

				if rule.Replacement != "" {
					violation.Replacement = rule.Replacement
				}

				violations = append(violations, violation)

				// Handle rule-specific actions
				switch rule.Action {
				case "replace":
					if rule.Replacement != "" {
						modifiedContent = rule.Pattern.ReplaceAllString(modifiedContent, rule.Replacement)
						contentModified = true
					}
				case "block":
					shouldBlock = true
				case "warn", "audit":
					// These are handled at the filter level
				}
			}
		}
	}

	// Determine overall action based on violations and configuration
	var action shared.FilterAction
	var blocked bool
	var reason string

	if shouldBlock {
		// If any rule requires blocking, block regardless of filter action
		action = shared.PluginActionBlock
		blocked = true
		reason = fmt.Sprintf("Content blocked by regex rules: %d violations found", len(violations))
	} else if len(violations) > 0 {
		switch f.config.Action {
		case "block":
			action = shared.PluginActionBlock
			blocked = true
			reason = fmt.Sprintf("Content blocked: %d regex violations found", len(violations))
		case "warn":
			action = shared.FilterActionWarn
			blocked = false
			reason = fmt.Sprintf("Content warning: %d regex matches found", len(violations))
		case "audit":
			action = shared.FilterActionAudit
			blocked = false
			reason = fmt.Sprintf("Content audit: %d regex matches logged", len(violations))
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
	var resultContent *shared.PluginContent
	if contentModified {
		resultContent = shared.CreatePluginContent(modifiedContent, content.Parsed, content.Headers, content.Params)
	} else {
		resultContent = content
	}

	return result, resultContent, nil
}

// Configure updates the filter configuration
func (f *RegexFilter) Configure(config map[string]interface{}) error {
	// Parse configuration
	regexConfig := &RegexConfig{
		Rules:         []Rule{},
		Action:        "warn",
		LogViolations: true,
		LogMatches:    false,
	}

	// Load rules
	if rules, ok := config["rules"].([]interface{}); ok {
		for _, rule := range rules {
			if ruleMap, ok := rule.(map[string]interface{}); ok {
				regexRule := Rule{
					Name:        shared.GetConfigValue(ruleMap, "name", ""),
					Pattern:     shared.GetConfigValue(ruleMap, "pattern", ""),
					Replacement: shared.GetConfigValue(ruleMap, "replacement", ""),
					Enabled:     shared.GetConfigValue(ruleMap, "enabled", true),
					Severity:    shared.GetConfigValue(ruleMap, "severity", "medium"),
					Category:    shared.GetConfigValue(ruleMap, "category", "custom"),
					Description: shared.GetConfigValue(ruleMap, "description", ""),
					Action:      shared.GetConfigValue(ruleMap, "action", "replace"),
				}
				if regexRule.Name != "" && regexRule.Pattern != "" {
					regexConfig.Rules = append(regexConfig.Rules, regexRule)
				}
			}
		}
	}

	// Load action
	regexConfig.Action = shared.GetConfigValue(config, "action", "warn")

	// Load log violations setting
	regexConfig.LogViolations = shared.GetConfigValue(config, "log_violations", true)

	// Load log matches setting
	regexConfig.LogMatches = shared.GetConfigValue(config, "log_matches", false)

	f.config = regexConfig
	f.BasePlugin.SetConfig(config)

	// Compile regex patterns
	return f.compileRules()
}

// compileRules compiles all regex patterns
func (f *RegexFilter) compileRules() error {
	f.rules = []*RegexRule{}

	for _, rule := range f.config.Rules {
		if !rule.Enabled || rule.Pattern == "" {
			continue
		}

		compiled, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return fmt.Errorf("failed to compile regex pattern for rule '%s': %w", rule.Name, err)
		}

		compiledRule := &RegexRule{
			Name:        rule.Name,
			Pattern:     compiled,
			Replacement: rule.Replacement,
			Enabled:     rule.Enabled,
			Severity:    rule.Severity,
			Category:    rule.Category,
			Description: rule.Description,
			Action:      rule.Action,
		}

		f.rules = append(f.rules, compiledRule)
	}

	return nil
}

// RegexFilterFactory implements FilterFactory for Regex filters
type RegexFilterFactory struct{}

// Create creates a new Regex filter instance
func (f *RegexFilterFactory) Create(config map[string]interface{}) (shared.Filter, error) {
	name := shared.GetConfigValue(config, "name", "regex-filter")
	return NewRegexFilter(name, config)
}

// GetType returns the filter type
func (f *RegexFilterFactory) GetType() shared.FilterType {
	return shared.FilterTypeRegex
}

// GetName returns the factory name
func (f *RegexFilterFactory) GetName() string {
	return "Regex Filter"
}

// GetDescription returns the factory description
func (f *RegexFilterFactory) GetDescription() string {
	return "Pattern-based content filtering with search and replace functionality"
}

// GetSupportedExecutionModes returns supported execution modes
func (f *RegexFilterFactory) GetSupportedExecutionModes() []string {
	return []string{
		string(shared.PluginModeEnforcing),
		string(shared.PluginModePermissive),
		string(shared.PluginModeDisabled),
		string(shared.PluginModeAuditOnly),
	}
}

// ValidateConfig validates the configuration for Regex filters
func (f *RegexFilterFactory) ValidateConfig(config map[string]interface{}) error {
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

	// Validate rules
	if rules, ok := config["rules"].([]interface{}); ok {
		for i, rule := range rules {
			if ruleMap, ok := rule.(map[string]interface{}); ok {
				// Validate pattern
				if pattern, exists := ruleMap["pattern"].(string); exists && pattern != "" {
					if _, err := regexp.Compile(pattern); err != nil {
						return fmt.Errorf("invalid regex pattern in rules[%d]: %w", i, err)
					}
				}

				// Validate rule action
				if ruleAction, exists := ruleMap["action"].(string); exists {
					validRuleActions := []string{"replace", "block", "warn", "audit"}
					valid := false
					for _, vra := range validRuleActions {
						if ruleAction == vra {
							valid = true
							break
						}
					}
					if !valid {
						return fmt.Errorf("invalid rule action in rules[%d]: %s", i, ruleAction)
					}
				}
			}
		}
	}

	return nil
}

// GetDefaultConfig returns the default configuration for Regex filters
func (f *RegexFilterFactory) GetDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"rules": []Rule{
			{
				Name:        "remove-extra-spaces",
				Pattern:     `\s{2,}`,
				Replacement: " ",
				Enabled:     false,
				Severity:    "low",
				Category:    "formatting",
				Description: "Remove extra spaces from content",
				Action:      "replace",
			},
		},
		"action":         "warn",
		"log_violations": true,
		"log_matches":    false,
	}
}

// GetConfigSchema returns the JSON schema for configuration validation
func (f *RegexFilterFactory) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"rules": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name":        map[string]interface{}{"type": "string"},
						"pattern":     map[string]interface{}{"type": "string"},
						"replacement": map[string]interface{}{"type": "string"},
						"enabled":     map[string]interface{}{"type": "boolean"},
						"severity":    map[string]interface{}{"type": "string"},
						"category":    map[string]interface{}{"type": "string"},
						"description": map[string]interface{}{"type": "string"},
						"action": map[string]interface{}{
							"type": "string",
							"enum": []string{"replace", "block", "warn", "audit"},
						},
					},
					"required": []string{"name", "pattern"},
				},
			},
			"action": map[string]interface{}{
				"type": "string",
				"enum": []string{"block", "warn", "audit", "allow"},
			},
			"log_violations": map[string]interface{}{"type": "boolean"},
			"log_matches":    map[string]interface{}{"type": "boolean"},
		},
	}
}
