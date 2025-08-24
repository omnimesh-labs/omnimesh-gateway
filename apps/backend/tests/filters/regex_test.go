package plugins

import (
	"context"
	"testing"

	"mcp-gateway/apps/backend/internal/plugins/content_filters/regex"
	"mcp-gateway/apps/backend/internal/plugins/shared"
)

func TestRegexFilter_NewRegexFilter(t *testing.T) {
	config := map[string]interface{}{
		"rules": []interface{}{
			map[string]interface{}{
				"name":        "remove-spaces",
				"pattern":     `\s+`,
				"replacement": " ",
				"enabled":     true,
				"action":      "replace",
			},
		},
		"action": "warn",
	}

	filter, err := regex.NewRegexFilter("test-regex", config)
	if err != nil {
		t.Fatalf("Failed to create Regex filter: %v", err)
	}

	if filter.GetName() != "test-regex" {
		t.Errorf("Expected name 'test-regex', got '%s'", filter.GetName())
	}

	if filter.GetType() != shared.PluginTypeRegex {
		t.Errorf("Expected type '%s', got '%s'", shared.PluginTypeRegex, filter.GetType())
	}

	if !filter.IsEnabled() {
		t.Errorf("Expected filter to be enabled by default")
	}
}

func TestRegexFilter_ApplyReplaceRule(t *testing.T) {
	config := map[string]interface{}{
		"rules": []interface{}{
			map[string]interface{}{
				"name":        "normalize-spaces",
				"pattern":     `\s{2,}`,
				"replacement": " ",
				"enabled":     true,
				"action":      "replace",
				"severity":    "low",
				"category":    "formatting",
			},
		},
		"action": "allow",
	}

	filter, err := regex.NewRegexFilter("test-regex", config)
	if err != nil {
		t.Fatalf("Failed to create Regex filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
	content := shared.CreatePluginContent("This  has    extra     spaces", nil, nil, nil)

	result, modifiedContent, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if result.Blocked {
		t.Errorf("Expected content not to be blocked with 'allow' action")
	}

	if result.Action != shared.FilterActionAllow {
		t.Errorf("Expected action to be 'allow', got '%s'", result.Action)
	}

	if len(result.Violations) == 0 {
		t.Errorf("Expected violations to be found")
	}

	if !result.Modified {
		t.Errorf("Expected content to be modified")
	}

	expected := "This has extra spaces"
	if modifiedContent.Raw != expected {
		t.Errorf("Expected '%s', got '%s'", expected, modifiedContent.Raw)
	}

	// Check violation details
	violation := result.Violations[0]
	if violation.Type != "regex_match" {
		t.Errorf("Expected violation type 'regex_match', got '%s'", violation.Type)
	}

	if violation.Severity != "low" {
		t.Errorf("Expected severity 'low', got '%s'", violation.Severity)
	}
}

func TestRegexFilter_ApplyBlockRule(t *testing.T) {
	config := map[string]interface{}{
		"rules": []interface{}{
			map[string]interface{}{
				"name":        "block-profanity",
				"pattern":     `(?i)\b(badword|curse)\b`,
				"replacement": "",
				"enabled":     true,
				"action":      "block",
				"severity":    "high",
			},
		},
		"action": "allow",
	}

	filter, err := regex.NewRegexFilter("test-regex", config)
	if err != nil {
		t.Fatalf("Failed to create Regex filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
	content := shared.CreatePluginContent("This contains a badword in it", nil, nil, nil)

	result, _, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if !result.Blocked {
		t.Errorf("Expected content to be blocked due to rule-level block action")
	}

	if result.Action != shared.FilterActionBlock {
		t.Errorf("Expected action to be 'block', got '%s'", result.Action)
	}

	if len(result.Violations) == 0 {
		t.Errorf("Expected violations to be found")
	}

	// Check violation metadata
	violation := result.Violations[0]
	if violation.Metadata["action"] != "block" {
		t.Errorf("Expected violation action to be 'block', got '%s'", violation.Metadata["action"])
	}

	if violation.Metadata["rule_name"] != "block-profanity" {
		t.Errorf("Expected violation rule_name to be 'block-profanity', got '%s'", violation.Metadata["rule_name"])
	}
}

func TestRegexFilter_ApplyMultipleRules(t *testing.T) {
	config := map[string]interface{}{
		"rules": []interface{}{
			map[string]interface{}{
				"name":        "normalize-spaces",
				"pattern":     `\s{2,}`,
				"replacement": " ",
				"enabled":     true,
				"action":      "replace",
			},
			map[string]interface{}{
				"name":        "remove-tabs",
				"pattern":     `\t`,
				"replacement": "    ",
				"enabled":     true,
				"action":      "replace",
			},
		},
		"action": "allow",
	}

	filter, err := regex.NewRegexFilter("test-regex", config)
	if err != nil {
		t.Fatalf("Failed to create Regex filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
	content := shared.CreatePluginContent("This  has\textra     spaces", nil, nil, nil)

	result, modifiedContent, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if result.Blocked {
		t.Errorf("Expected content not to be blocked")
	}

	if !result.Modified {
		t.Errorf("Expected content to be modified")
	}

	expected := "This has    extra spaces"
	if modifiedContent.Raw != expected {
		t.Errorf("Expected '%s', got '%s'", expected, modifiedContent.Raw)
	}

	// Should have violations for both rules
	if len(result.Violations) < 2 {
		t.Errorf("Expected at least 2 violations, got %d", len(result.Violations))
	}
}

func TestRegexFilter_ApplyWarnAction(t *testing.T) {
	config := map[string]interface{}{
		"rules": []interface{}{
			map[string]interface{}{
				"name":        "detect-numbers",
				"pattern":     `\d+`,
				"replacement": "",
				"enabled":     true,
				"action":      "warn",
			},
		},
		"action": "warn",
	}

	filter, err := regex.NewRegexFilter("test-regex", config)
	if err != nil {
		t.Fatalf("Failed to create Regex filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
	content := shared.CreatePluginContent("I have 123 items", nil, nil, nil)

	result, modifiedContent, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if result.Blocked {
		t.Errorf("Expected content not to be blocked with 'warn' action")
	}

	if result.Action != shared.FilterActionWarn {
		t.Errorf("Expected action to be 'warn', got '%s'", result.Action)
	}

	if len(result.Violations) == 0 {
		t.Errorf("Expected violations to be found")
	}

	// Content should not be modified for 'warn' action
	if result.Modified {
		t.Errorf("Expected content not to be modified with 'warn' action")
	}

	if modifiedContent.Raw != content.Raw {
		t.Errorf("Expected content to remain unchanged")
	}
}

func TestRegexFilter_ApplyDisabledRule(t *testing.T) {
	config := map[string]interface{}{
		"rules": []interface{}{
			map[string]interface{}{
				"name":        "disabled-rule",
				"pattern":     `\d+`,
				"replacement": "XXX",
				"enabled":     false,
				"action":      "replace",
			},
			map[string]interface{}{
				"name":        "enabled-rule",
				"pattern":     `[A-Z]+`,
				"replacement": "YYY",
				"enabled":     true,
				"action":      "replace",
			},
		},
		"action": "allow",
	}

	filter, err := regex.NewRegexFilter("test-regex", config)
	if err != nil {
		t.Fatalf("Failed to create Regex filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
	content := shared.CreatePluginContent("Test 123 HELLO", nil, nil, nil)

	result, modifiedContent, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if result.Blocked {
		t.Errorf("Expected content not to be blocked")
	}

	if !result.Modified {
		t.Errorf("Expected content to be modified")
	}

	// Should only apply enabled rule
	expected := "Test 123 YYY"
	if modifiedContent.Raw != expected {
		t.Errorf("Expected '%s', got '%s'", expected, modifiedContent.Raw)
	}

	// Should only have one violation (from enabled rule)
	if len(result.Violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(result.Violations))
	}
}

func TestRegexFilter_ApplyNoMatches(t *testing.T) {
	config := map[string]interface{}{
		"rules": []interface{}{
			map[string]interface{}{
				"name":        "find-digits",
				"pattern":     `\d+`,
				"replacement": "XXX",
				"enabled":     true,
				"action":      "replace",
			},
		},
		"action": "block",
	}

	filter, err := regex.NewRegexFilter("test-regex", config)
	if err != nil {
		t.Fatalf("Failed to create Regex filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
	content := shared.CreatePluginContent("This has no numbers", nil, nil, nil)

	result, modifiedContent, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if result.Blocked {
		t.Errorf("Expected content not to be blocked when no matches")
	}

	if result.Action != shared.FilterActionAllow {
		t.Errorf("Expected action to be 'allow', got '%s'", result.Action)
	}

	if len(result.Violations) != 0 {
		t.Errorf("Expected no violations, got %d", len(result.Violations))
	}

	if result.Modified {
		t.Errorf("Expected content not to be modified")
	}

	if modifiedContent.Raw != content.Raw {
		t.Errorf("Expected content to remain unchanged")
	}
}

func TestRegexFilter_ApplyDisabled(t *testing.T) {
	config := map[string]interface{}{
		"rules": []interface{}{
			map[string]interface{}{
				"name":        "find-digits",
				"pattern":     `\d+`,
				"replacement": "XXX",
				"enabled":     true,
				"action":      "block",
			},
		},
		"action": "block",
	}

	filter, err := regex.NewRegexFilter("test-regex", config)
	if err != nil {
		t.Fatalf("Failed to create Regex filter: %v", err)
	}

	// Disable the filter
	filter.SetEnabled(false)

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
	content := shared.CreatePluginContent("This has 123 numbers", nil, nil, nil)

	result, _, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if result.Blocked {
		t.Errorf("Expected content not to be blocked when filter is disabled")
	}

	if result.Action != shared.FilterActionAllow {
		t.Errorf("Expected action to be 'allow', got '%s'", result.Action)
	}

	if len(result.Violations) != 0 {
		t.Errorf("Expected no violations when filter is disabled, got %d", len(result.Violations))
	}

	if result.Modified {
		t.Errorf("Expected content not to be modified when filter is disabled")
	}
}

func TestRegexFilterFactory_ValidateConfig(t *testing.T) {
	factory := &regex.RegexFilterFactory{}

	tests := []struct {
		name        string
		config      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"action": "warn",
				"rules": []interface{}{
					map[string]interface{}{
						"name":    "test-rule",
						"pattern": `\d+`,
						"action":  "replace",
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid action",
			config: map[string]interface{}{
				"action": "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid regex pattern",
			config: map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						"name":    "bad-rule",
						"pattern": "[invalid",
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid rule action",
			config: map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						"name":    "test-rule",
						"pattern": `\d+`,
						"action":  "invalid",
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := factory.ValidateConfig(tt.config)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestRegexFilterFactory_GetDefaultConfig(t *testing.T) {
	factory := &regex.RegexFilterFactory{}
	config := factory.GetDefaultConfig()

	if config == nil {
		t.Errorf("Expected default config to be non-nil")
	}

	if action, ok := config["action"].(string); !ok || action != "warn" {
		t.Errorf("Expected default action to be 'warn', got '%v'", config["action"])
	}

	if rules, ok := config["rules"]; !ok {
		t.Errorf("Expected default config to have 'rules' key")
	} else if rulesSlice, ok := rules.([]regex.Rule); !ok || len(rulesSlice) == 0 {
		t.Errorf("Expected default rules to be non-empty slice")
	}
}

func TestRegexFilter_GetCapabilities(t *testing.T) {
	config := map[string]interface{}{
		"rules": []interface{}{},
		"action": "allow",
	}

	filter, err := regex.NewRegexFilter("test-regex", config)
	if err != nil {
		t.Fatalf("Failed to create Regex filter: %v", err)
	}

	capabilities := filter.GetCapabilities()

	if !capabilities.SupportsInbound {
		t.Errorf("Expected filter to support inbound direction")
	}

	if !capabilities.SupportsOutbound {
		t.Errorf("Expected filter to support outbound direction")
	}

	if !capabilities.SupportsModification {
		t.Errorf("Expected filter to support modification")
	}

	if !capabilities.SupportsBlocking {
		t.Errorf("Expected filter to support blocking")
	}

	if !capabilities.SupportsRealtime {
		t.Errorf("Expected filter to support realtime processing")
	}

	if !capabilities.SupportsBatch {
		t.Errorf("Expected filter to support batch processing")
	}
}

func TestRegexFilter_Configure(t *testing.T) {
	filter, err := regex.NewRegexFilter("test-regex", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to create Regex filter: %v", err)
	}

	newConfig := map[string]interface{}{
		"rules": []interface{}{
			map[string]interface{}{
				"name":        "new-rule",
				"pattern":     `test`,
				"replacement": "TEST",
				"enabled":     true,
				"action":      "replace",
			},
		},
		"action": "warn",
	}

	err = filter.Configure(newConfig)
	if err != nil {
		t.Fatalf("Failed to configure filter: %v", err)
	}

	config := filter.GetConfig()
	if action, ok := config["action"].(string); !ok || action != "warn" {
		t.Errorf("Expected configured action to be 'warn', got '%v'", config["action"])
	}
}