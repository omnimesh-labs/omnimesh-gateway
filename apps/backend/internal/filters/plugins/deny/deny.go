package deny

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"mcp-gateway/apps/backend/internal/filters/shared"
)

// DenyFilter implements simple word/phrase blocking with violation reporting
type DenyFilter struct {
	*shared.BaseFilter
	config        *DenyConfig
	blockedWords  []*BlockedWord
	compiledRegex []*regexp.Regexp
}

// DenyConfig holds the configuration for the Deny filter
type DenyConfig struct {
	BlockedWords    []string     `json:"blocked_words"`
	BlockedPhrases  []string     `json:"blocked_phrases"`
	BlockedPatterns []string     `json:"blocked_patterns"`
	CaseSensitive   bool         `json:"case_sensitive"`
	WholeWordsOnly  bool         `json:"whole_words_only"`
	Action          string       `json:"action"`
	LogViolations   bool         `json:"log_violations"`
	CustomRules     []CustomRule `json:"custom_rules"`
}

// BlockedWord represents a blocked word with metadata
type BlockedWord struct {
	Word     string
	Original string
	Severity string
	Category string
}

// CustomRule allows for more complex blocking rules
type CustomRule struct {
	Name        string `json:"name"`
	Pattern     string `json:"pattern"`
	Enabled     bool   `json:"enabled"`
	Severity    string `json:"severity"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

// NewDenyFilter creates a new Deny filter instance
func NewDenyFilter(name string, config map[string]interface{}) (*DenyFilter, error) {
	baseFilter := shared.NewBaseFilter(shared.FilterTypeDeny, name, 30)

	// Set capabilities
	baseFilter.SetCapabilities(shared.FilterCapabilities{
		SupportsInbound:       true,
		SupportsOutbound:      true,
		SupportsModification:  false, // Deny filter blocks but doesn't modify
		SupportsBlocking:      true,
		SupportedContentTypes: []string{"*"},
		SupportsRealtime:      true,
		SupportsBatch:         true,
	})

	filter := &DenyFilter{
		BaseFilter: baseFilter,
	}

	if err := filter.Configure(config); err != nil {
		return nil, fmt.Errorf("failed to configure Deny filter: %w", err)
	}

	return filter, nil
}

// Apply applies the Deny filter to content
func (f *DenyFilter) Apply(ctx context.Context, filterCtx *shared.FilterContext, content *shared.FilterContent) (*shared.FilterResult, *shared.FilterContent, error) {
	if !f.BaseFilter.IsEnabled() {
		return shared.CreateFilterResult(false, false, shared.FilterActionAllow, "", nil), content, nil
	}

	violations := []shared.FilterViolation{}
	searchText := content.Raw

	// Convert to lowercase for case-insensitive search
	if !f.config.CaseSensitive {
		searchText = strings.ToLower(searchText)
	}

	// Check blocked words
	for _, blockedWord := range f.blockedWords {
		violations = append(violations, f.checkWord(content.Raw, searchText, blockedWord)...)
	}

	// Check custom pattern rules
	for _, regex := range f.compiledRegex {
		matches := regex.FindAllStringSubmatch(searchText, -1)
		for _, match := range matches {
			if len(match) > 0 {
				violation := shared.CreateFilterViolation(
					"blocked_pattern",
					regex.String(),
					match[0],
					regex.FindStringIndex(searchText)[0],
					"medium",
				)
				violation.Metadata["pattern"] = regex.String()
				violations = append(violations, violation)
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
			action = shared.FilterActionBlock
			blocked = true
			reason = fmt.Sprintf("Content blocked: %d prohibited items found", len(violations))
		case "warn":
			action = shared.FilterActionWarn
			blocked = false
			reason = fmt.Sprintf("Content warning: %d prohibited items found", len(violations))
		case "audit":
			action = shared.FilterActionAudit
			blocked = false
			reason = fmt.Sprintf("Content audit: %d prohibited items logged", len(violations))
		default:
			action = shared.FilterActionAllow
			blocked = false
		}
	} else {
		action = shared.FilterActionAllow
		blocked = false
	}

	result := shared.CreateFilterResult(blocked, false, action, reason, violations)

	return result, content, nil
}

// checkWord checks for blocked words in the content
func (f *DenyFilter) checkWord(originalContent, searchText string, blockedWord *BlockedWord) []shared.FilterViolation {
	var violations []shared.FilterViolation
	searchWord := blockedWord.Word

	if f.config.WholeWordsOnly {
		// Use word boundary regex for whole words only
		pattern := `\b` + regexp.QuoteMeta(searchWord) + `\b`
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return violations // Skip invalid patterns
		}

		matches := regex.FindAllStringIndex(searchText, -1)
		for _, match := range matches {
			violation := shared.CreateFilterViolation(
				"blocked_word",
				pattern,
				originalContent[match[0]:match[1]],
				match[0],
				blockedWord.Severity,
			)
			violation.Metadata["category"] = blockedWord.Category
			violation.Metadata["word"] = blockedWord.Original
			violations = append(violations, violation)
		}
	} else {
		// Simple substring search
		startIndex := 0
		for {
			index := strings.Index(searchText[startIndex:], searchWord)
			if index == -1 {
				break
			}

			absoluteIndex := startIndex + index
			actualMatch := originalContent[absoluteIndex : absoluteIndex+len(searchWord)]

			violation := shared.CreateFilterViolation(
				"blocked_word",
				searchWord,
				actualMatch,
				absoluteIndex,
				blockedWord.Severity,
			)
			violation.Metadata["category"] = blockedWord.Category
			violation.Metadata["word"] = blockedWord.Original
			violations = append(violations, violation)

			startIndex = absoluteIndex + 1
		}
	}

	return violations
}

// Configure updates the filter configuration
func (f *DenyFilter) Configure(config map[string]interface{}) error {
	// Parse configuration
	denyConfig := &DenyConfig{
		BlockedWords:    []string{},
		BlockedPhrases:  []string{},
		BlockedPatterns: []string{},
		CaseSensitive:   false,
		WholeWordsOnly:  false,
		Action:          "warn",
		LogViolations:   true,
		CustomRules:     []CustomRule{},
	}

	// Load blocked words
	denyConfig.BlockedWords = shared.GetConfigStringSlice(config, "blocked_words", []string{})

	// Load blocked phrases
	denyConfig.BlockedPhrases = shared.GetConfigStringSlice(config, "blocked_phrases", []string{})

	// Load blocked patterns
	denyConfig.BlockedPatterns = shared.GetConfigStringSlice(config, "blocked_patterns", []string{})

	// Load case sensitivity setting
	denyConfig.CaseSensitive = shared.GetConfigValue(config, "case_sensitive", false)

	// Load whole words only setting
	denyConfig.WholeWordsOnly = shared.GetConfigValue(config, "whole_words_only", false)

	// Load action
	denyConfig.Action = shared.GetConfigValue(config, "action", "warn")

	// Load log violations setting
	denyConfig.LogViolations = shared.GetConfigValue(config, "log_violations", true)

	// Load custom rules
	if customRules, ok := config["custom_rules"].([]interface{}); ok {
		for _, rule := range customRules {
			if ruleMap, ok := rule.(map[string]interface{}); ok {
				customRule := CustomRule{
					Name:        shared.GetConfigValue(ruleMap, "name", ""),
					Pattern:     shared.GetConfigValue(ruleMap, "pattern", ""),
					Enabled:     shared.GetConfigValue(ruleMap, "enabled", true),
					Severity:    shared.GetConfigValue(ruleMap, "severity", "medium"),
					Category:    shared.GetConfigValue(ruleMap, "category", "custom"),
					Description: shared.GetConfigValue(ruleMap, "description", ""),
				}
				if customRule.Name != "" && customRule.Pattern != "" && customRule.Enabled {
					denyConfig.CustomRules = append(denyConfig.CustomRules, customRule)
				}
			}
		}
	}

	f.config = denyConfig
	f.BaseFilter.SetConfig(config)

	// Compile blocked words and patterns
	return f.compileBlockedItems()
}

// compileBlockedItems compiles blocked words, phrases, and patterns
func (f *DenyFilter) compileBlockedItems() error {
	f.blockedWords = []*BlockedWord{}
	f.compiledRegex = []*regexp.Regexp{}

	// Process blocked words
	for _, word := range f.config.BlockedWords {
		if word != "" {
			blockedWord := &BlockedWord{
				Word:     f.normalizeWord(word),
				Original: word,
				Severity: "medium",
				Category: "word",
			}
			f.blockedWords = append(f.blockedWords, blockedWord)
		}
	}

	// Process blocked phrases
	for _, phrase := range f.config.BlockedPhrases {
		if phrase != "" {
			blockedWord := &BlockedWord{
				Word:     f.normalizeWord(phrase),
				Original: phrase,
				Severity: "medium",
				Category: "phrase",
			}
			f.blockedWords = append(f.blockedWords, blockedWord)
		}
	}

	// Compile blocked patterns
	for _, pattern := range f.config.BlockedPatterns {
		if pattern != "" {
			compiled, err := regexp.Compile(pattern)
			if err != nil {
				return fmt.Errorf("failed to compile blocked pattern '%s': %w", pattern, err)
			}
			f.compiledRegex = append(f.compiledRegex, compiled)
		}
	}

	// Compile custom rules
	for _, rule := range f.config.CustomRules {
		if rule.Enabled && rule.Pattern != "" {
			compiled, err := regexp.Compile(rule.Pattern)
			if err != nil {
				return fmt.Errorf("failed to compile custom rule pattern '%s': %w", rule.Pattern, err)
			}
			f.compiledRegex = append(f.compiledRegex, compiled)
		}
	}

	return nil
}

// normalizeWord normalizes a word based on case sensitivity setting
func (f *DenyFilter) normalizeWord(word string) string {
	if !f.config.CaseSensitive {
		return strings.ToLower(word)
	}
	return word
}

// DenyFilterFactory implements FilterFactory for Deny filters
type DenyFilterFactory struct{}

// Create creates a new Deny filter instance
func (f *DenyFilterFactory) Create(config map[string]interface{}) (shared.Filter, error) {
	name := shared.GetConfigValue(config, "name", "deny-filter")
	return NewDenyFilter(name, config)
}

// GetType returns the filter type
func (f *DenyFilterFactory) GetType() shared.FilterType {
	return shared.FilterTypeDeny
}

// GetName returns the factory name
func (f *DenyFilterFactory) GetName() string {
	return "Deny Filter"
}

// GetDescription returns the factory description
func (f *DenyFilterFactory) GetDescription() string {
	return "Blocks content containing prohibited words, phrases, or patterns"
}

// ValidateConfig validates the configuration for Deny filters
func (f *DenyFilterFactory) ValidateConfig(config map[string]interface{}) error {
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

	// Validate blocked patterns
	if patterns, ok := config["blocked_patterns"].([]interface{}); ok {
		for i, pattern := range patterns {
			if patternStr, ok := pattern.(string); ok && patternStr != "" {
				if _, err := regexp.Compile(patternStr); err != nil {
					return fmt.Errorf("invalid regex pattern in blocked_patterns[%d]: %w", i, err)
				}
			}
		}
	}

	// Validate custom rules patterns
	if customRules, ok := config["custom_rules"].([]interface{}); ok {
		for i, rule := range customRules {
			if ruleMap, ok := rule.(map[string]interface{}); ok {
				if pattern, exists := ruleMap["pattern"].(string); exists && pattern != "" {
					if _, err := regexp.Compile(pattern); err != nil {
						return fmt.Errorf("invalid regex pattern in custom_rules[%d]: %w", i, err)
					}
				}
			}
		}
	}

	return nil
}

// GetDefaultConfig returns the default configuration for Deny filters
func (f *DenyFilterFactory) GetDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"blocked_words":    []string{"password", "secret", "token"},
		"blocked_phrases":  []string{},
		"blocked_patterns": []string{},
		"case_sensitive":   false,
		"whole_words_only": false,
		"action":           "warn",
		"log_violations":   true,
		"custom_rules":     []CustomRule{},
	}
}

// GetConfigSchema returns the JSON schema for configuration validation
func (f *DenyFilterFactory) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"blocked_words": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "string"},
			},
			"blocked_phrases": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "string"},
			},
			"blocked_patterns": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "string"},
			},
			"case_sensitive":   map[string]interface{}{"type": "boolean"},
			"whole_words_only": map[string]interface{}{"type": "boolean"},
			"action": map[string]interface{}{
				"type": "string",
				"enum": []string{"block", "warn", "audit", "allow"},
			},
			"log_violations": map[string]interface{}{"type": "boolean"},
			"custom_rules": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name":        map[string]interface{}{"type": "string"},
						"pattern":     map[string]interface{}{"type": "string"},
						"enabled":     map[string]interface{}{"type": "boolean"},
						"severity":    map[string]interface{}{"type": "string"},
						"category":    map[string]interface{}{"type": "string"},
						"description": map[string]interface{}{"type": "string"},
					},
					"required": []string{"name", "pattern"},
				},
			},
		},
	}
}
