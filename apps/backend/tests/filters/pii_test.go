package plugins

import (
	"context"
	"testing"

	"mcp-gateway/apps/backend/internal/plugins/content_filters/pii"
	"mcp-gateway/apps/backend/internal/plugins/shared"
)

func TestPIIFilter_NewPIIFilter(t *testing.T) {
	config := map[string]interface{}{
		"patterns": map[string]interface{}{
			"ssn":   true,
			"email": true,
		},
		"masking_strategy": "redact",
		"action":           "warn",
	}

	filter, err := pii.NewPIIFilter("test-pii", config)
	if err != nil {
		t.Fatalf("Failed to create PII filter: %v", err)
	}

	if filter.GetName() != "test-pii" {
		t.Errorf("Expected name 'test-pii', got '%s'", filter.GetName())
	}

	if filter.GetType() != shared.FilterTypePII {
		t.Errorf("Expected type '%s', got '%s'", shared.FilterTypePII, filter.GetType())
	}

	if !filter.IsEnabled() {
		t.Errorf("Expected filter to be enabled by default")
	}
}

func TestPIIFilter_ApplySSNDetection(t *testing.T) {
	config := map[string]interface{}{
		"patterns": map[string]interface{}{
			"ssn": true,
		},
		"masking_strategy": "redact",
		"action":           "block",
	}

	filter, err := pii.NewPIIFilter("test-pii", config)
	if err != nil {
		t.Fatalf("Failed to create PII filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreateFilterContext("req-1", "org-1", "user-1", "", "", shared.FilterDirectionInbound, "text/plain")
	content := shared.CreateFilterContent("My SSN is 123-45-6789", nil, nil, nil)

	result, modifiedContent, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if !result.Blocked {
		t.Errorf("Expected content to be blocked due to SSN")
	}

	if result.Action != shared.FilterActionBlock {
		t.Errorf("Expected action to be 'block', got '%s'", result.Action)
	}

	if len(result.Violations) == 0 {
		t.Errorf("Expected violations to be found")
	}

	// Check that content was modified (masked)
	if !result.Modified {
		t.Errorf("Expected content to be modified")
	}

	if modifiedContent.Raw == content.Raw {
		t.Errorf("Expected content to be masked")
	}

	if modifiedContent.Raw != "My SSN is [REDACTED]" {
		t.Errorf("Expected 'My SSN is [REDACTED]', got '%s'", modifiedContent.Raw)
	}
}

func TestPIIFilter_ApplyEmailDetection(t *testing.T) {
	config := map[string]interface{}{
		"patterns": map[string]interface{}{
			"email": true,
		},
		"masking_strategy": "partial",
		"action":           "warn",
	}

	filter, err := pii.NewPIIFilter("test-pii", config)
	if err != nil {
		t.Fatalf("Failed to create PII filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreateFilterContext("req-1", "org-1", "user-1", "", "", shared.FilterDirectionInbound, "text/plain")
	content := shared.CreateFilterContent("Contact me at john.doe@example.com", nil, nil, nil)

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

	if !result.Modified {
		t.Errorf("Expected content to be modified")
	}

	// Check partial masking (jo****@example.com format)
	if !contains(modifiedContent.Raw, "jo") || !contains(modifiedContent.Raw, "om") {
		t.Errorf("Expected partial masking, got '%s'", modifiedContent.Raw)
	}
}

func TestPIIFilter_ApplyCreditCardDetection(t *testing.T) {
	config := map[string]interface{}{
		"patterns": map[string]interface{}{
			"credit_card": true,
		},
		"masking_strategy": "hash",
		"action":           "audit",
	}

	filter, err := pii.NewPIIFilter("test-pii", config)
	if err != nil {
		t.Fatalf("Failed to create PII filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreateFilterContext("req-1", "org-1", "user-1", "", "", shared.FilterDirectionInbound, "text/plain")
	content := shared.CreateFilterContent("My credit card is 4111111111111111", nil, nil, nil)

	result, modifiedContent, err := filter.Apply(ctx, filterCtx, content)
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

	if !result.Modified {
		t.Errorf("Expected content to be modified")
	}

	// Check hash masking
	if !contains(modifiedContent.Raw, "[HASH:") {
		t.Errorf("Expected hash masking, got '%s'", modifiedContent.Raw)
	}
}

func TestPIIFilter_ApplyCustomPattern(t *testing.T) {
	config := map[string]interface{}{
		"patterns": map[string]interface{}{},
		"custom_patterns": []interface{}{
			map[string]interface{}{
				"name":        "api_key",
				"pattern":     `api_key_[a-zA-Z0-9]{32}`,
				"enabled":     true,
				"severity":    "high",
				"description": "API key detection",
			},
		},
		"masking_strategy": "tokenize",
		"action":           "block",
	}

	filter, err := pii.NewPIIFilter("test-pii", config)
	if err != nil {
		t.Fatalf("Failed to create PII filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreateFilterContext("req-1", "org-1", "user-1", "", "", shared.FilterDirectionInbound, "text/plain")
	content := shared.CreateFilterContent("Here is my key: api_key_abcdef1234567890abcdef1234567890", nil, nil, nil)

	result, modifiedContent, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if !result.Blocked {
		t.Errorf("Expected content to be blocked due to custom API key pattern")
	}

	if result.Action != shared.FilterActionBlock {
		t.Errorf("Expected action to be 'block', got '%s'", result.Action)
	}

	if len(result.Violations) == 0 {
		t.Errorf("Expected violations to be found")
	}

	// Check violation details
	violation := result.Violations[0]
	if violation.Severity != "high" {
		t.Errorf("Expected severity 'high', got '%s'", violation.Severity)
	}

	if !result.Modified {
		t.Errorf("Expected content to be modified")
	}

	// Check tokenize masking
	if !contains(modifiedContent.Raw, "[TOKEN:") {
		t.Errorf("Expected tokenize masking, got '%s'", modifiedContent.Raw)
	}
}

func TestPIIFilter_ApplyNoViolations(t *testing.T) {
	config := map[string]interface{}{
		"patterns": map[string]interface{}{
			"ssn":   true,
			"email": true,
		},
		"masking_strategy": "redact",
		"action":           "block",
	}

	filter, err := pii.NewPIIFilter("test-pii", config)
	if err != nil {
		t.Fatalf("Failed to create PII filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreateFilterContext("req-1", "org-1", "user-1", "", "", shared.FilterDirectionInbound, "text/plain")
	content := shared.CreateFilterContent("This is clean content with no PII", nil, nil, nil)

	result, modifiedContent, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if result.Blocked {
		t.Errorf("Expected content not to be blocked")
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

func TestPIIFilter_ApplyDisabled(t *testing.T) {
	config := map[string]interface{}{
		"patterns": map[string]interface{}{
			"ssn": true,
		},
		"action": "block",
	}

	filter, err := pii.NewPIIFilter("test-pii", config)
	if err != nil {
		t.Fatalf("Failed to create PII filter: %v", err)
	}

	// Disable the filter
	filter.SetEnabled(false)

	ctx := context.Background()
	filterCtx := shared.CreateFilterContext("req-1", "org-1", "user-1", "", "", shared.FilterDirectionInbound, "text/plain")
	content := shared.CreateFilterContent("My SSN is 123-45-6789", nil, nil, nil)

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

func TestPIIFilterFactory_ValidateConfig(t *testing.T) {
	factory := &pii.PIIFilterFactory{}

	tests := []struct {
		name        string
		config      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"masking_strategy": "redact",
				"action":           "warn",
			},
			expectError: false,
		},
		{
			name: "invalid masking strategy",
			config: map[string]interface{}{
				"masking_strategy": "invalid",
				"action":           "warn",
			},
			expectError: true,
		},
		{
			name: "invalid action",
			config: map[string]interface{}{
				"masking_strategy": "redact",
				"action":           "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid custom pattern regex",
			config: map[string]interface{}{
				"custom_patterns": []interface{}{
					map[string]interface{}{
						"name":    "bad_pattern",
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

func TestPIIFilter_GetStats(t *testing.T) {
	config := map[string]interface{}{
		"patterns": map[string]interface{}{
			"ssn": true,
		},
		"action": "block",
	}

	filter, err := pii.NewPIIFilter("test-pii", config)
	if err != nil {
		t.Fatalf("Failed to create PII filter: %v", err)
	}

	stats := filter.GetStats()
	if stats == nil {
		t.Errorf("Expected stats to be non-nil")
	}

	if stats.Name != "test-pii" {
		t.Errorf("Expected stats name 'test-pii', got '%s'", stats.Name)
	}

	if stats.Type != string(shared.FilterTypePII) {
		t.Errorf("Expected stats type '%s', got '%s'", shared.FilterTypePII, stats.Type)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr || 
	       len(s) > len(substr) && s[len(s)-len(substr):] == substr ||
	       len(s) > len(substr) && s[1:len(s)-1] != substr && 
	       s != substr && (s[0:2] == substr[0:2] || s[len(s)-2:] == substr[len(substr)-2:])
}