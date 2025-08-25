package plugins

import (
	"context"
	"testing"

	"mcp-gateway/apps/backend/internal/plugins/content_filters/deny"
	"mcp-gateway/apps/backend/internal/plugins/shared"
)

func TestDenyFilter_NewDenyFilter(t *testing.T) {
	config := map[string]interface{}{
		"blocked_words": []string{"password", "secret"},
		"action":        "warn",
	}

	filter, err := deny.NewDenyFilter("test-deny", config)
	if err != nil {
		t.Fatalf("Failed to create Deny filter: %v", err)
	}

	if filter.GetName() != "test-deny" {
		t.Errorf("Expected name 'test-deny', got '%s'", filter.GetName())
	}

	if filter.GetType() != shared.PluginTypeDeny {
		t.Errorf("Expected type '%s', got '%s'", shared.PluginTypeDeny, filter.GetType())
	}

	if !filter.IsEnabled() {
		t.Errorf("Expected filter to be enabled by default")
	}
}

func TestDenyFilter_ApplyBlockedWord(t *testing.T) {
	config := map[string]interface{}{
		"blocked_words":  []string{"password", "secret", "token"},
		"action":         "block",
		"case_sensitive": false,
	}

	filter, err := deny.NewDenyFilter("test-deny", config)
	if err != nil {
		t.Fatalf("Failed to create Deny filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
	content := shared.CreatePluginContent("Please enter your password here", nil, nil, nil)

	result, _, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if !result.Blocked {
		t.Errorf("Expected content to be blocked due to blocked word")
	}

	if result.Action != shared.FilterActionBlock {
		t.Errorf("Expected action to be 'block', got '%s'", result.Action)
	}

	if len(result.Violations) == 0 {
		t.Errorf("Expected violations to be found")
	}

	// Check violation details
	violation := result.Violations[0]
	if violation.Type != "blocked_word" {
		t.Errorf("Expected violation type 'blocked_word', got '%s'", violation.Type)
	}

	if violation.Metadata["category"] != "word" {
		t.Errorf("Expected category 'word', got '%s'", violation.Metadata["category"])
	}

	if violation.Metadata["word"] != "password" {
		t.Errorf("Expected word 'password', got '%s'", violation.Metadata["word"])
	}
}

func TestDenyFilter_ApplyBlockedPhrase(t *testing.T) {
	config := map[string]interface{}{
		"blocked_phrases": []string{"credit card number", "social security"},
		"action":          "warn",
		"case_sensitive":  false,
	}

	filter, err := deny.NewDenyFilter("test-deny", config)
	if err != nil {
		t.Fatalf("Failed to create Deny filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
	content := shared.CreatePluginContent("Please provide your credit card number", nil, nil, nil)

	result, _, err := filter.Apply(ctx, filterCtx, content)
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

	// Check violation details
	violation := result.Violations[0]
	if violation.Type != "blocked_word" {
		t.Errorf("Expected violation type 'blocked_word', got '%s'", violation.Type)
	}

	if violation.Metadata["category"] != "phrase" {
		t.Errorf("Expected category 'phrase', got '%s'", violation.Metadata["category"])
	}
}

func TestDenyFilter_ApplyCaseSensitive(t *testing.T) {
	config := map[string]interface{}{
		"blocked_words":  []string{"Password"},
		"action":         "block",
		"case_sensitive": true,
	}

	filter, err := deny.NewDenyFilter("test-deny", config)
	if err != nil {
		t.Fatalf("Failed to create Deny filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")

	// Test exact case match - should block
	content1 := shared.CreatePluginContent("Enter your Password", nil, nil, nil)
	result1, _, err := filter.Apply(ctx, filterCtx, content1)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if !result1.Blocked {
		t.Errorf("Expected content to be blocked with exact case match")
	}

	// Test different case - should not block
	content2 := shared.CreatePluginContent("Enter your password", nil, nil, nil)
	result2, _, err := filter.Apply(ctx, filterCtx, content2)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if result2.Blocked {
		t.Errorf("Expected content not to be blocked with different case")
	}

	if result2.Action != shared.FilterActionAllow {
		t.Errorf("Expected action to be 'allow', got '%s'", result2.Action)
	}
}

func TestDenyFilter_ApplyWholeWordsOnly(t *testing.T) {
	config := map[string]interface{}{
		"blocked_words":    []string{"pass"},
		"action":           "block",
		"case_sensitive":   false,
		"whole_words_only": true,
	}

	filter, err := deny.NewDenyFilter("test-deny", config)
	if err != nil {
		t.Fatalf("Failed to create Deny filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")

	// Test whole word match - should block
	content1 := shared.CreatePluginContent("Enter the pass code", nil, nil, nil)
	result1, _, err := filter.Apply(ctx, filterCtx, content1)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if !result1.Blocked {
		t.Errorf("Expected content to be blocked with whole word match")
	}

	// Test partial word match - should not block
	content2 := shared.CreatePluginContent("Enter the password", nil, nil, nil)
	result2, _, err := filter.Apply(ctx, filterCtx, content2)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if result2.Blocked {
		t.Errorf("Expected content not to be blocked with partial word match")
	}

	if result2.Action != shared.FilterActionAllow {
		t.Errorf("Expected action to be 'allow', got '%s'", result2.Action)
	}
}

func TestDenyFilter_ApplyBlockedPattern(t *testing.T) {
	config := map[string]interface{}{
		"blocked_patterns": []string{`\b\d{3}-\d{2}-\d{4}\b`}, // SSN pattern
		"action":           "audit",
	}

	filter, err := deny.NewDenyFilter("test-deny", config)
	if err != nil {
		t.Fatalf("Failed to create Deny filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
	content := shared.CreatePluginContent("My SSN is 123-45-6789", nil, nil, nil)

	result, _, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if result.Blocked {
		t.Errorf("Expected content not to be blocked with 'audit' action")
	}

	if result.Action != shared.FilterActionAudit {
		t.Errorf("Expected action to be 'audit', got '%s'", result.Action)
	}

	if len(result.Violations) == 0 {
		t.Errorf("Expected violations to be found")
	}

	// Check violation details
	violation := result.Violations[0]
	if violation.Type != "blocked_pattern" {
		t.Errorf("Expected violation type 'blocked_pattern', got '%s'", violation.Type)
	}

	if violation.Pattern == "" {
		t.Errorf("Expected pattern to be set in violation")
	}
}

func TestDenyFilter_ApplyCustomRules(t *testing.T) {
	config := map[string]interface{}{
		"custom_rules": []interface{}{
			map[string]interface{}{
				"name":        "api-key-detection",
				"pattern":     `api[_-]?key[:\s=]+[a-zA-Z0-9]{20,}`,
				"enabled":     true,
				"severity":    "high",
				"category":    "security",
				"description": "Detect API keys in content",
			},
		},
		"action": "block",
	}

	filter, err := deny.NewDenyFilter("test-deny", config)
	if err != nil {
		t.Fatalf("Failed to create Deny filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
	content := shared.CreatePluginContent("api_key=abcdef1234567890123456789", nil, nil, nil)

	result, _, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if !result.Blocked {
		t.Errorf("Expected content to be blocked due to custom rule")
	}

	if result.Action != shared.FilterActionBlock {
		t.Errorf("Expected action to be 'block', got '%s'", result.Action)
	}

	if len(result.Violations) == 0 {
		t.Errorf("Expected violations to be found")
	}

	// Check violation details
	violation := result.Violations[0]
	if violation.Type != "blocked_pattern" {
		t.Errorf("Expected violation type 'blocked_pattern', got '%s'", violation.Type)
	}
}

func TestDenyFilter_ApplyMultipleMatches(t *testing.T) {
	config := map[string]interface{}{
		"blocked_words":  []string{"password", "secret"},
		"action":         "warn",
		"case_sensitive": false,
	}

	filter, err := deny.NewDenyFilter("test-deny", config)
	if err != nil {
		t.Fatalf("Failed to create Deny filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
	content := shared.CreatePluginContent("Enter password and secret key", nil, nil, nil)

	result, _, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if result.Blocked {
		t.Errorf("Expected content not to be blocked with 'warn' action")
	}

	if result.Action != shared.FilterActionWarn {
		t.Errorf("Expected action to be 'warn', got '%s'", result.Action)
	}

	// Should have violations for both words
	if len(result.Violations) < 2 {
		t.Errorf("Expected at least 2 violations, got %d", len(result.Violations))
	}
}

func TestDenyFilter_ApplyRepeatedWord(t *testing.T) {
	config := map[string]interface{}{
		"blocked_words":  []string{"test"},
		"action":         "warn",
		"case_sensitive": false,
	}

	filter, err := deny.NewDenyFilter("test-deny", config)
	if err != nil {
		t.Fatalf("Failed to create Deny filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
	content := shared.CreatePluginContent("This is a test and another test", nil, nil, nil)

	result, _, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if result.Blocked {
		t.Errorf("Expected content not to be blocked with 'warn' action")
	}

	if result.Action != shared.FilterActionWarn {
		t.Errorf("Expected action to be 'warn', got '%s'", result.Action)
	}

	// Should have violations for both occurrences of "test"
	if len(result.Violations) != 2 {
		t.Errorf("Expected 2 violations for repeated word, got %d", len(result.Violations))
	}
}

func TestDenyFilter_ApplyNoMatches(t *testing.T) {
	config := map[string]interface{}{
		"blocked_words": []string{"password", "secret"},
		"action":        "block",
	}

	filter, err := deny.NewDenyFilter("test-deny", config)
	if err != nil {
		t.Fatalf("Failed to create Deny filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
	content := shared.CreatePluginContent("This is clean content with no blocked words", nil, nil, nil)

	result, _, err := filter.Apply(ctx, filterCtx, content)
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
}

func TestDenyFilter_ApplyDisabledCustomRule(t *testing.T) {
	config := map[string]interface{}{
		"custom_rules": []interface{}{
			map[string]interface{}{
				"name":        "disabled-rule",
				"pattern":     `test`,
				"enabled":     false,
				"severity":    "medium",
				"category":    "custom",
				"description": "Disabled test rule",
			},
			map[string]interface{}{
				"name":        "enabled-rule",
				"pattern":     `sample`,
				"enabled":     true,
				"severity":    "low",
				"category":    "custom",
				"description": "Enabled test rule",
			},
		},
		"action": "warn",
	}

	filter, err := deny.NewDenyFilter("test-deny", config)
	if err != nil {
		t.Fatalf("Failed to create Deny filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
	content := shared.CreatePluginContent("This is a test sample", nil, nil, nil)

	result, _, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if result.Action != shared.FilterActionWarn {
		t.Errorf("Expected action to be 'warn', got '%s'", result.Action)
	}

	// Should only have one violation (from enabled rule)
	if len(result.Violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(result.Violations))
	}

	if len(result.Violations) > 0 {
		violation := result.Violations[0]
		if violation.Match != "sample" {
			t.Errorf("Expected match 'sample', got '%s'", violation.Match)
		}
	}
}

func TestDenyFilter_ApplyDisabled(t *testing.T) {
	config := map[string]interface{}{
		"blocked_words": []string{"password"},
		"action":        "block",
	}

	filter, err := deny.NewDenyFilter("test-deny", config)
	if err != nil {
		t.Fatalf("Failed to create Deny filter: %v", err)
	}

	// Disable the filter
	filter.SetEnabled(false)

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
	content := shared.CreatePluginContent("Enter your password", nil, nil, nil)

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
}

func TestDenyFilterFactory_ValidateConfig(t *testing.T) {
	factory := &deny.DenyFilterFactory{}

	tests := []struct {
		config      map[string]interface{}
		name        string
		expectError bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"blocked_words": []string{"test"},
				"action":        "warn",
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
			name: "invalid blocked pattern regex",
			config: map[string]interface{}{
				"blocked_patterns": []interface{}{"[invalid"},
			},
			expectError: true,
		},
		{
			name: "invalid custom rule regex",
			config: map[string]interface{}{
				"custom_rules": []interface{}{
					map[string]interface{}{
						"name":    "bad-rule",
						"pattern": "[invalid",
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

func TestDenyFilterFactory_GetDefaultConfig(t *testing.T) {
	factory := &deny.DenyFilterFactory{}
	config := factory.GetDefaultConfig()

	if config == nil {
		t.Errorf("Expected default config to be non-nil")
	}

	if action, ok := config["action"].(string); !ok || action != "warn" {
		t.Errorf("Expected default action to be 'warn', got '%v'", config["action"])
	}

	if blockedWords, ok := config["blocked_words"].([]string); !ok || len(blockedWords) == 0 {
		t.Errorf("Expected default blocked_words to be non-empty slice")
	}
}

func TestDenyFilter_GetCapabilities(t *testing.T) {
	config := map[string]interface{}{
		"blocked_words": []string{"test"},
		"action":        "warn",
	}

	filter, err := deny.NewDenyFilter("test-deny", config)
	if err != nil {
		t.Fatalf("Failed to create Deny filter: %v", err)
	}

	capabilities := filter.GetCapabilities()

	if !capabilities.SupportsInbound {
		t.Errorf("Expected filter to support inbound direction")
	}

	if !capabilities.SupportsOutbound {
		t.Errorf("Expected filter to support outbound direction")
	}

	if capabilities.SupportsModification {
		t.Errorf("Expected filter not to support modification")
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

func TestDenyFilter_Configure(t *testing.T) {
	filter, err := deny.NewDenyFilter("test-deny", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to create Deny filter: %v", err)
	}

	newConfig := map[string]interface{}{
		"blocked_words":    []string{"new", "config"},
		"action":           "audit",
		"case_sensitive":   true,
		"whole_words_only": true,
	}

	err = filter.Configure(newConfig)
	if err != nil {
		t.Fatalf("Failed to configure filter: %v", err)
	}

	config := filter.GetConfig()
	if action, ok := config["action"].(string); !ok || action != "audit" {
		t.Errorf("Expected configured action to be 'audit', got '%v'", config["action"])
	}

	if caseSensitive, ok := config["case_sensitive"].(bool); !ok || !caseSensitive {
		t.Errorf("Expected case_sensitive to be true, got '%v'", config["case_sensitive"])
	}
}
