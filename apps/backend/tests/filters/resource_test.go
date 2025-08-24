package plugins

import (
	"context"
	"testing"

	"mcp-gateway/apps/backend/internal/plugins/content_filters/resource"
	"mcp-gateway/apps/backend/internal/plugins/shared"
)

func TestResourceFilter_NewResourceFilter(t *testing.T) {
	config := map[string]interface{}{
		"allowed_protocols": []string{"https", "http"},
		"blocked_domains":   []string{"malicious.com"},
		"action":            "block",
	}

	filter, err := resource.NewResourceFilter("test-resource", config)
	if err != nil {
		t.Fatalf("Failed to create Resource filter: %v", err)
	}

	if filter.GetName() != "test-resource" {
		t.Errorf("Expected name 'test-resource', got '%s'", filter.GetName())
	}

	if filter.GetType() != shared.PluginTypeResource {
		t.Errorf("Expected type '%s', got '%s'", shared.PluginTypeResource, filter.GetType())
	}

	if !filter.IsEnabled() {
		t.Errorf("Expected filter to be enabled by default")
	}
}

func TestResourceFilter_ApplyBlockedProtocol(t *testing.T) {
	config := map[string]interface{}{
		"allowed_protocols": []string{"https"},
		"action":            "block",
	}

	filter, err := resource.NewResourceFilter("test-resource", config)
	if err != nil {
		t.Fatalf("Failed to create Resource filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
	content := shared.CreatePluginContent("Please visit http://example.com for more info", nil, nil, nil)

	result, _, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if !result.Blocked {
		t.Errorf("Expected content to be blocked due to disallowed protocol")
	}

	if result.Action != shared.FilterActionBlock {
		t.Errorf("Expected action to be 'block', got '%s'", result.Action)
	}

	if len(result.Violations) == 0 {
		t.Errorf("Expected violations to be found")
	}

	// Check violation details
	violation := result.Violations[0]
	if violation.Type != "blocked_protocol" {
		t.Errorf("Expected violation type 'blocked_protocol', got '%s'", violation.Type)
	}

	if violation.Metadata["protocol"] != "http" {
		t.Errorf("Expected protocol 'http', got '%s'", violation.Metadata["protocol"])
	}
}

func TestResourceFilter_ApplyAllowedProtocol(t *testing.T) {
	config := map[string]interface{}{
		"allowed_protocols": []string{"https", "http"},
		"action":            "block",
	}

	filter, err := resource.NewResourceFilter("test-resource", config)
	if err != nil {
		t.Fatalf("Failed to create Resource filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
	content := shared.CreatePluginContent("Please visit https://example.com for more info", nil, nil, nil)

	result, _, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	// Should not be blocked if protocol is allowed and no other violations
	if result.Blocked {
		t.Errorf("Expected content not to be blocked with allowed protocol")
	}

	if result.Action != shared.FilterActionAllow {
		t.Errorf("Expected action to be 'allow', got '%s'", result.Action)
	}
}

func TestResourceFilter_ApplyBlockedDomain(t *testing.T) {
	config := map[string]interface{}{
		"allowed_protocols": []string{"https", "http"},
		"blocked_domains":   []string{"malicious.com", "blocked.net"},
		"action":            "block",
	}

	filter, err := resource.NewResourceFilter("test-resource", config)
	if err != nil {
		t.Fatalf("Failed to create Resource filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
	content := shared.CreatePluginContent("Don't visit https://malicious.com/bad", nil, nil, nil)

	result, _, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if !result.Blocked {
		t.Errorf("Expected content to be blocked due to blocked domain")
	}

	if result.Action != shared.FilterActionBlock {
		t.Errorf("Expected action to be 'block', got '%s'", result.Action)
	}

	if len(result.Violations) == 0 {
		t.Errorf("Expected violations to be found")
	}

	// Check violation details
	violation := result.Violations[0]
	if violation.Type != "blocked_domain" {
		t.Errorf("Expected violation type 'blocked_domain', got '%s'", violation.Type)
	}

	if violation.Metadata["domain"] != "malicious.com" {
		t.Errorf("Expected domain 'malicious.com', got '%s'", violation.Metadata["domain"])
	}
}

func TestResourceFilter_ApplyAllowedDomainWhitelist(t *testing.T) {
	config := map[string]interface{}{
		"allowed_protocols": []string{"https", "http"},
		"allowed_domains":   []string{"safe.com", "trusted.org"},
		"action":            "block",
	}

	filter, err := resource.NewResourceFilter("test-resource", config)
	if err != nil {
		t.Fatalf("Failed to create Resource filter: %v", err)
	}

	// Test allowed domain
	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
	content := shared.CreatePluginContent("Visit https://safe.com for help", nil, nil, nil)

	result, _, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if result.Blocked {
		t.Errorf("Expected content not to be blocked with whitelisted domain")
	}

	// Test non-allowed domain
	content2 := shared.CreatePluginContent("Visit https://unknown.com for help", nil, nil, nil)
	result2, _, err2 := filter.Apply(ctx, filterCtx, content2)
	if err2 != nil {
		t.Fatalf("Filter apply failed: %v", err2)
	}

	if !result2.Blocked {
		t.Errorf("Expected content to be blocked with non-whitelisted domain")
	}

	if len(result2.Violations) == 0 {
		t.Errorf("Expected violations to be found")
	}

	violation := result2.Violations[0]
	if violation.Type != "domain_not_allowed" {
		t.Errorf("Expected violation type 'domain_not_allowed', got '%s'", violation.Type)
	}
}

func TestResourceFilter_ApplyWildcardDomain(t *testing.T) {
	config := map[string]interface{}{
		"allowed_protocols": []string{"https"},
		"allowed_domains":   []string{"*.example.com"},
		"action":            "block",
	}

	filter, err := resource.NewResourceFilter("test-resource", config)
	if err != nil {
		t.Fatalf("Failed to create Resource filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")

	// Test subdomain match
	content := shared.CreatePluginContent("Visit https://api.example.com/data", nil, nil, nil)
	result, _, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if result.Blocked {
		t.Errorf("Expected content not to be blocked with wildcard subdomain match")
	}

	// Test exact domain match  
	content2 := shared.CreatePluginContent("Visit https://example.com/data", nil, nil, nil)
	result2, _, err2 := filter.Apply(ctx, filterCtx, content2)
	if err2 != nil {
		t.Fatalf("Filter apply failed: %v", err2)
	}

	if result2.Blocked {
		t.Errorf("Expected content not to be blocked with wildcard exact match")
	}
}

func TestResourceFilter_ApplyLocalhostBlocked(t *testing.T) {
	config := map[string]interface{}{
		"allowed_protocols":      []string{"http", "https"},
		"allow_localhost":        false,
		"allow_private_networks": false,
		"action":                 "block",
	}

	filter, err := resource.NewResourceFilter("test-resource", config)
	if err != nil {
		t.Fatalf("Failed to create Resource filter: %v", err)
	}

	testCases := []string{
		"http://localhost:8080/api",
		"https://127.0.0.1:3000/data",
		"http://::1/test",
		"http://0.0.0.0:5000/endpoint",
	}

	for _, testURL := range testCases {
		ctx := context.Background()
		filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
		content := shared.CreatePluginContent("Connect to "+testURL, nil, nil, nil)

		result, _, err := filter.Apply(ctx, filterCtx, content)
		if err != nil {
			t.Fatalf("Filter apply failed for %s: %v", testURL, err)
		}

		if !result.Blocked {
			t.Errorf("Expected %s to be blocked", testURL)
		}

		if len(result.Violations) == 0 {
			t.Errorf("Expected violations for %s", testURL)
		}

		violation := result.Violations[0]
		if violation.Type != "localhost_access" {
			t.Errorf("Expected violation type 'localhost_access' for %s, got '%s'", testURL, violation.Type)
		}
	}
}

func TestResourceFilter_ApplyPrivateNetworkBlocked(t *testing.T) {
	config := map[string]interface{}{
		"allowed_protocols":      []string{"http"},
		"allow_private_networks": false,
		"action":                 "block",
	}

	filter, err := resource.NewResourceFilter("test-resource", config)
	if err != nil {
		t.Fatalf("Failed to create Resource filter: %v", err)
	}

	testCases := []string{
		"http://10.0.0.1/api",
		"http://192.168.1.100/data",
		"http://172.16.0.1/test",
		"http://172.31.255.255/endpoint",
	}

	for _, testURL := range testCases {
		ctx := context.Background()
		filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
		content := shared.CreatePluginContent("Connect to "+testURL, nil, nil, nil)

		result, _, err := filter.Apply(ctx, filterCtx, content)
		if err != nil {
			t.Fatalf("Filter apply failed for %s: %v", testURL, err)
		}

		if !result.Blocked {
			t.Errorf("Expected %s to be blocked", testURL)
		}

		if len(result.Violations) == 0 {
			t.Errorf("Expected violations for %s", testURL)
		}

		violation := result.Violations[0]
		if violation.Type != "private_network_access" {
			t.Errorf("Expected violation type 'private_network_access' for %s, got '%s'", testURL, violation.Type)
		}
	}
}

func TestResourceFilter_ApplyContentSizeExceeded(t *testing.T) {
	config := map[string]interface{}{
		"max_content_size": 100,
		"action":           "block",
	}

	filter, err := resource.NewResourceFilter("test-resource", config)
	if err != nil {
		t.Fatalf("Failed to create Resource filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
	
	// Create content larger than the limit
	largeContent := make([]byte, 150)
	for i := range largeContent {
		largeContent[i] = 'A'
	}
	content := shared.CreatePluginContent(string(largeContent), nil, nil, nil)

	result, _, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if !result.Blocked {
		t.Errorf("Expected content to be blocked due to size limit")
	}

	if len(result.Violations) == 0 {
		t.Errorf("Expected violations to be found")
	}

	violation := result.Violations[0]
	if violation.Type != "content_size_exceeded" {
		t.Errorf("Expected violation type 'content_size_exceeded', got '%s'", violation.Type)
	}

	if violation.Metadata["content_size"].(int64) != 150 {
		t.Errorf("Expected content_size 150, got %v", violation.Metadata["content_size"])
	}
}

func TestResourceFilter_ApplyBlockedContentType(t *testing.T) {
	config := map[string]interface{}{
		"blocked_content_types": []string{"application/octet-stream", "text/html"},
		"action":                "block",
	}

	filter, err := resource.NewResourceFilter("test-resource", config)
	if err != nil {
		t.Fatalf("Failed to create Resource filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/html")
	content := shared.CreatePluginContent("<html><body>Test</body></html>", nil, nil, nil)

	result, _, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if !result.Blocked {
		t.Errorf("Expected content to be blocked due to blocked content type")
	}

	if len(result.Violations) == 0 {
		t.Errorf("Expected violations to be found")
	}

	violation := result.Violations[0]
	if violation.Type != "blocked_content_type" {
		t.Errorf("Expected violation type 'blocked_content_type', got '%s'", violation.Type)
	}

	if violation.Metadata["content_type"] != "text/html" {
		t.Errorf("Expected content_type 'text/html', got '%s'", violation.Metadata["content_type"])
	}
}

func TestResourceFilter_ApplyAllowedContentType(t *testing.T) {
	config := map[string]interface{}{
		"allowed_content_types": []string{"application/json", "text/plain"},
		"action":                "block",
	}

	filter, err := resource.NewResourceFilter("test-resource", config)
	if err != nil {
		t.Fatalf("Failed to create Resource filter: %v", err)
	}

	ctx := context.Background()

	// Test allowed content type
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "application/json")
	content := shared.CreatePluginContent(`{"key": "value"}`, nil, nil, nil)

	result, _, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if result.Blocked {
		t.Errorf("Expected content not to be blocked with allowed content type")
	}

	// Test disallowed content type
	filterCtx2 := shared.CreatePluginContext("req-2", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/html")
	content2 := shared.CreatePluginContent("<html></html>", nil, nil, nil)

	result2, _, err2 := filter.Apply(ctx, filterCtx2, content2)
	if err2 != nil {
		t.Fatalf("Filter apply failed: %v", err2)
	}

	if !result2.Blocked {
		t.Errorf("Expected content to be blocked with disallowed content type")
	}

	if len(result2.Violations) == 0 {
		t.Errorf("Expected violations to be found")
	}

	violation := result2.Violations[0]
	if violation.Type != "content_type_not_allowed" {
		t.Errorf("Expected violation type 'content_type_not_allowed', got '%s'", violation.Type)
	}
}

func TestResourceFilter_ApplyInvalidURL(t *testing.T) {
	config := map[string]interface{}{
		"action": "warn",
	}

	filter, err := resource.NewResourceFilter("test-resource", config)
	if err != nil {
		t.Fatalf("Failed to create Resource filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
	content := shared.CreatePluginContent("Visit ://invalid-url for more info", nil, nil, nil)

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
		t.Errorf("Expected violations to be found for invalid URL")
	}

	violation := result.Violations[0]
	if violation.Type != "invalid_url" {
		t.Errorf("Expected violation type 'invalid_url', got '%s'", violation.Type)
	}
}

func TestResourceFilter_ApplyNoUrls(t *testing.T) {
	config := map[string]interface{}{
		"blocked_domains": []string{"malicious.com"},
		"action":          "block",
	}

	filter, err := resource.NewResourceFilter("test-resource", config)
	if err != nil {
		t.Fatalf("Failed to create Resource filter: %v", err)
	}

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
	content := shared.CreatePluginContent("This is just plain text with no URLs", nil, nil, nil)

	result, _, err := filter.Apply(ctx, filterCtx, content)
	if err != nil {
		t.Fatalf("Filter apply failed: %v", err)
	}

	if result.Blocked {
		t.Errorf("Expected content not to be blocked when no URLs present")
	}

	if result.Action != shared.FilterActionAllow {
		t.Errorf("Expected action to be 'allow', got '%s'", result.Action)
	}

	if len(result.Violations) != 0 {
		t.Errorf("Expected no violations when no URLs present, got %d", len(result.Violations))
	}
}

func TestResourceFilter_ApplyDisabled(t *testing.T) {
	config := map[string]interface{}{
		"blocked_domains": []string{"malicious.com"},
		"action":          "block",
	}

	filter, err := resource.NewResourceFilter("test-resource", config)
	if err != nil {
		t.Fatalf("Failed to create Resource filter: %v", err)
	}

	// Disable the filter
	filter.SetEnabled(false)

	ctx := context.Background()
	filterCtx := shared.CreatePluginContext("req-1", "org-1", "user-1", "", "", shared.PluginDirectionInbound, "text/plain")
	content := shared.CreatePluginContent("Visit https://malicious.com/bad", nil, nil, nil)

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

func TestResourceFilterFactory_ValidateConfig(t *testing.T) {
	factory := &resource.ResourceFilterFactory{}

	tests := []struct {
		name        string
		config      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"allowed_protocols": []string{"https"},
				"max_content_size":  1024,
				"action":            "block",
			},
			expectError: false,
		},
		{
			name: "negative content size",
			config: map[string]interface{}{
				"max_content_size": -100,
			},
			expectError: true,
		},
		{
			name: "invalid action",
			config: map[string]interface{}{
				"action": "invalid",
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

func TestResourceFilter_GetCapabilities(t *testing.T) {
	config := map[string]interface{}{
		"action": "block",
	}

	filter, err := resource.NewResourceFilter("test-resource", config)
	if err != nil {
		t.Fatalf("Failed to create Resource filter: %v", err)
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

	if capabilities.SupportsBatch {
		t.Errorf("Expected filter not to support batch processing")
	}
}