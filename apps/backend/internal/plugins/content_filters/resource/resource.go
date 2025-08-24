package resource

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"mcp-gateway/apps/backend/internal/plugins/shared"
)

// ResourceFilter implements URI validation, protocol filtering, domain blocking, and content filtering
type ResourceFilter struct {
	*shared.BasePlugin
	config *ResourceConfig
}

// ResourceConfig holds the configuration for the Resource filter
type ResourceConfig struct {
	AllowedProtocols     []string `json:"allowed_protocols"`
	BlockedDomains       []string `json:"blocked_domains"`
	AllowedDomains       []string `json:"allowed_domains"`
	MaxContentSize       int64    `json:"max_content_size"`
	AllowedContentTypes  []string `json:"allowed_content_types"`
	BlockedContentTypes  []string `json:"blocked_content_types"`
	AllowPrivateNetworks bool     `json:"allow_private_networks"`
	AllowLocalhost       bool     `json:"allow_localhost"`
	Action               string   `json:"action"`
	LogViolations        bool     `json:"log_violations"`
}

// NewResourceFilter creates a new Resource filter instance
func NewResourceFilter(name string, config map[string]interface{}) (*ResourceFilter, error) {
	basePlugin := shared.NewBasePlugin(shared.PluginTypeResource, name, 20)

	// Set capabilities
	basePlugin.SetCapabilities(shared.PluginCapabilities{
		SupportsInbound:       true,
		SupportsOutbound:      true,
		SupportsModification:  false, // Resource filter blocks but doesn't modify
		SupportsBlocking:      true,
		SupportedContentTypes: []string{"*"},
		SupportsRealtime:      true,
		SupportsBatch:         false,
	})

	filter := &ResourceFilter{
		BasePlugin: basePlugin,
	}

	if err := filter.Configure(config); err != nil {
		return nil, fmt.Errorf("failed to configure Resource filter: %w", err)
	}

	return filter, nil
}

// Apply applies the Resource filter to content
func (f *ResourceFilter) Apply(ctx context.Context, pluginCtx *shared.PluginContext, content *shared.PluginContent) (*shared.PluginResult, *shared.PluginContent, error) {
	if !f.BasePlugin.IsEnabled() {
		return shared.CreatePluginResult(false, false, shared.PluginActionAllow, "", nil), content, nil
	}

	violations := []shared.PluginViolation{}

	// Extract URLs from content
	urls := f.extractURLs(content.Raw)

	// Check each URL
	for _, urlStr := range urls {
		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			violation := shared.CreatePluginViolation(
				"invalid_url",
				"",
				urlStr,
				0,
				"medium",
			)
			violation.Metadata["error"] = err.Error()
			violations = append(violations, violation)
			continue
		}

		// Check protocol
		if violation := f.checkProtocol(parsedURL); violation != nil {
			violations = append(violations, *violation)
		}

		// Check domain
		if violation := f.checkDomain(parsedURL); violation != nil {
			violations = append(violations, *violation)
		}

		// Check for private networks
		if violation := f.checkPrivateNetworks(parsedURL); violation != nil {
			violations = append(violations, *violation)
		}
	}

	// Check content size
	if violation := f.checkContentSize(content); violation != nil {
		violations = append(violations, *violation)
	}

	// Check content type
	if violation := f.checkContentType(pluginCtx, content); violation != nil {
		violations = append(violations, *violation)
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
			reason = fmt.Sprintf("Resource access denied: %d violations found", len(violations))
		case "warn":
			action = shared.FilterActionWarn
			blocked = false
			reason = fmt.Sprintf("Resource violations detected: %d issues found (warning)", len(violations))
		case "audit":
			action = shared.FilterActionAudit
			blocked = false
			reason = fmt.Sprintf("Resource violations logged: %d issues found for audit", len(violations))
		default:
			action = shared.PluginActionAllow
			blocked = false
		}
	} else {
		action = shared.PluginActionAllow
		blocked = false
	}

	result := shared.CreatePluginResult(blocked, false, action, reason, violations)

	return result, content, nil
}

// Configure updates the filter configuration
func (f *ResourceFilter) Configure(config map[string]interface{}) error {
	// Parse configuration
	resourceConfig := &ResourceConfig{
		AllowedProtocols:     []string{"https", "http"},
		BlockedDomains:       []string{},
		AllowedDomains:       []string{},
		MaxContentSize:       10485760, // 10MB default
		AllowedContentTypes:  []string{},
		BlockedContentTypes:  []string{},
		AllowPrivateNetworks: false,
		AllowLocalhost:       false,
		Action:               "block",
		LogViolations:        true,
	}

	// Load allowed protocols
	if protocols, ok := config["allowed_protocols"].([]interface{}); ok {
		resourceConfig.AllowedProtocols = make([]string, len(protocols))
		for i, p := range protocols {
			if str, ok := p.(string); ok {
				resourceConfig.AllowedProtocols[i] = strings.ToLower(str)
			}
		}
	} else if protocols := shared.GetConfigStringSlice(config, "allowed_protocols", nil); protocols != nil {
		resourceConfig.AllowedProtocols = protocols
	}

	// Load blocked domains
	resourceConfig.BlockedDomains = shared.GetConfigStringSlice(config, "blocked_domains", []string{})

	// Load allowed domains
	resourceConfig.AllowedDomains = shared.GetConfigStringSlice(config, "allowed_domains", []string{})

	// Load max content size
	resourceConfig.MaxContentSize = int64(shared.GetConfigValue(config, "max_content_size", 10485760))

	// Load allowed content types
	resourceConfig.AllowedContentTypes = shared.GetConfigStringSlice(config, "allowed_content_types", []string{})

	// Load blocked content types
	resourceConfig.BlockedContentTypes = shared.GetConfigStringSlice(config, "blocked_content_types", []string{})

	// Load private networks setting
	resourceConfig.AllowPrivateNetworks = shared.GetConfigValue(config, "allow_private_networks", false)

	// Load localhost setting
	resourceConfig.AllowLocalhost = shared.GetConfigValue(config, "allow_localhost", false)

	// Load action
	resourceConfig.Action = shared.GetConfigValue(config, "action", "block")

	// Load log violations setting
	resourceConfig.LogViolations = shared.GetConfigValue(config, "log_violations", true)

	f.config = resourceConfig
	f.BasePlugin.SetConfig(config)

	return f.BasePlugin.Validate()
}

// extractURLs extracts URLs from content using a simple regex
func (f *ResourceFilter) extractURLs(content string) []string {
	// Simple URL regex - can be improved for better accuracy
	urlRegex := regexp.MustCompile(`https?://[^\s<>"{}|\\^` + "`" + `\[\]]+`)
	matches := urlRegex.FindAllString(content, -1)

	var urls []string
	for _, match := range matches {
		urls = append(urls, strings.TrimRight(match, ".,;:!?)]}"))
	}

	return urls
}

// checkProtocol checks if the URL protocol is allowed
func (f *ResourceFilter) checkProtocol(parsedURL *url.URL) *shared.FilterViolation {
	protocol := strings.ToLower(parsedURL.Scheme)

	for _, allowed := range f.config.AllowedProtocols {
		if protocol == allowed {
			return nil
		}
	}

	violation := shared.CreatePluginViolation(
		"blocked_protocol",
		"",
		parsedURL.String(),
		0,
		"high",
	)
	violation.Metadata["protocol"] = protocol
	violation.Metadata["allowed_protocols"] = f.config.AllowedProtocols

	return &violation
}

// checkDomain checks if the domain is allowed or blocked
func (f *ResourceFilter) checkDomain(parsedURL *url.URL) *shared.FilterViolation {
	hostname := strings.ToLower(parsedURL.Hostname())

	// Check allowed domains (whitelist takes precedence)
	if len(f.config.AllowedDomains) > 0 {
		for _, allowed := range f.config.AllowedDomains {
			if f.matchesDomain(hostname, allowed) {
				return nil
			}
		}

		violation := shared.CreatePluginViolation(
			"domain_not_allowed",
			"",
			parsedURL.String(),
			0,
			"high",
		)
		violation.Metadata["domain"] = hostname
		violation.Metadata["allowed_domains"] = f.config.AllowedDomains

		return &violation
	}

	// Check blocked domains
	for _, blocked := range f.config.BlockedDomains {
		if f.matchesDomain(hostname, blocked) {
			violation := shared.CreatePluginViolation(
				"blocked_domain",
				"",
				parsedURL.String(),
				0,
				"high",
			)
			violation.Metadata["domain"] = hostname
			violation.Metadata["blocked_domain"] = blocked

			return &violation
		}
	}

	return nil
}

// checkPrivateNetworks checks for private network and localhost access
func (f *ResourceFilter) checkPrivateNetworks(parsedURL *url.URL) *shared.FilterViolation {
	hostname := strings.ToLower(parsedURL.Hostname())

	// Check localhost
	if !f.config.AllowLocalhost {
		localhostPatterns := []string{
			"localhost",
			"127.0.0.1",
			"::1",
			"0.0.0.0",
		}

		for _, pattern := range localhostPatterns {
			if hostname == pattern {
				violation := shared.CreatePluginViolation(
					"localhost_access",
					"",
					parsedURL.String(),
					0,
					"medium",
				)
				violation.Metadata["hostname"] = hostname

				return &violation
			}
		}
	}

	// Check private networks
	if !f.config.AllowPrivateNetworks {
		privatePatterns := []regexp.Regexp{
			*regexp.MustCompile(`^10\.`),                         // 10.0.0.0/8
			*regexp.MustCompile(`^192\.168\.`),                   // 192.168.0.0/16
			*regexp.MustCompile(`^172\.(1[6-9]|2[0-9]|3[01])\.`), // 172.16.0.0/12
		}

		for _, pattern := range privatePatterns {
			if pattern.MatchString(hostname) {
				violation := shared.CreatePluginViolation(
					"private_network_access",
					pattern.String(),
					parsedURL.String(),
					0,
					"medium",
				)
				violation.Metadata["hostname"] = hostname

				return &violation
			}
		}
	}

	return nil
}

// checkContentSize checks if content size exceeds limits
func (f *ResourceFilter) checkContentSize(content *shared.FilterContent) *shared.FilterViolation {
	if f.config.MaxContentSize <= 0 {
		return nil
	}

	contentSize := int64(len(content.Raw))
	if contentSize > f.config.MaxContentSize {
		violation := shared.CreatePluginViolation(
			"content_size_exceeded",
			"",
			"",
			0,
			"medium",
		)
		violation.Metadata["content_size"] = contentSize
		violation.Metadata["max_size"] = f.config.MaxContentSize

		return &violation
	}

	return nil
}

// checkContentType checks if content type is allowed or blocked
func (f *ResourceFilter) checkContentType(pluginCtx *shared.FilterContext, content *shared.FilterContent) *shared.FilterViolation {
	contentType := pluginCtx.ContentType
	if contentType == "" {
		// Try to get from headers
		if ct, exists := content.Headers["content-type"]; exists {
			contentType = ct
		}
	}

	if contentType == "" {
		return nil
	}

	// Normalize content type (remove parameters)
	if idx := strings.Index(contentType, ";"); idx > 0 {
		contentType = contentType[:idx]
	}
	contentType = strings.TrimSpace(strings.ToLower(contentType))

	// Check blocked content types
	for _, blocked := range f.config.BlockedContentTypes {
		if contentType == strings.ToLower(blocked) {
			violation := shared.CreatePluginViolation(
				"blocked_content_type",
				"",
				"",
				0,
				"medium",
			)
			violation.Metadata["content_type"] = contentType
			violation.Metadata["blocked_type"] = blocked

			return &violation
		}
	}

	// Check allowed content types (if specified)
	if len(f.config.AllowedContentTypes) > 0 {
		for _, allowed := range f.config.AllowedContentTypes {
			if contentType == strings.ToLower(allowed) {
				return nil
			}
		}

		violation := shared.CreatePluginViolation(
			"content_type_not_allowed",
			"",
			"",
			0,
			"medium",
		)
		violation.Metadata["content_type"] = contentType
		violation.Metadata["allowed_types"] = f.config.AllowedContentTypes

		return &violation
	}

	return nil
}

// matchesDomain checks if hostname matches domain pattern (supports wildcards)
func (f *ResourceFilter) matchesDomain(hostname, pattern string) bool {
	pattern = strings.ToLower(pattern)

	// Exact match
	if hostname == pattern {
		return true
	}

	// Wildcard subdomain match (*.example.com)
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:] // Remove *
		return strings.HasSuffix(hostname, suffix)
	}

	return false
}

// ResourceFilterFactory implements FilterFactory for Resource filters
type ResourceFilterFactory struct{}

// Create creates a new Resource filter instance
func (f *ResourceFilterFactory) Create(config map[string]interface{}) (shared.Filter, error) {
	name := shared.GetConfigValue(config, "name", "resource-filter")
	return NewResourceFilter(name, config)
}

// GetType returns the filter type
func (f *ResourceFilterFactory) GetType() shared.FilterType {
	return shared.FilterTypeResource
}

// GetName returns the factory name
func (f *ResourceFilterFactory) GetName() string {
	return "Resource Filter"
}

// GetDescription returns the factory description
func (f *ResourceFilterFactory) GetDescription() string {
	return "Validates URI access, filters protocols and domains, enforces content size limits"
}

// GetSupportedExecutionModes returns supported execution modes
func (f *ResourceFilterFactory) GetSupportedExecutionModes() []string {
	return []string{
		string(shared.PluginModeEnforcing),
		string(shared.PluginModePermissive),
		string(shared.PluginModeDisabled),
		string(shared.PluginModeAuditOnly),
	}
}

// ValidateConfig validates the configuration for Resource filters
func (f *ResourceFilterFactory) ValidateConfig(config map[string]interface{}) error {
	// Validate max content size
	if size, ok := config["max_content_size"].(float64); ok {
		if size < 0 {
			return fmt.Errorf("max_content_size must be non-negative")
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

	return nil
}

// GetDefaultConfig returns the default configuration for Resource filters
func (f *ResourceFilterFactory) GetDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"allowed_protocols":      []string{"https", "http"},
		"blocked_domains":        []string{"localhost", "127.0.0.1", "0.0.0.0"},
		"allowed_domains":        []string{},
		"max_content_size":       10485760, // 10MB
		"allowed_content_types":  []string{"application/json", "text/plain", "text/html"},
		"blocked_content_types":  []string{},
		"allow_private_networks": false,
		"allow_localhost":        false,
		"action":                 "block",
		"log_violations":         true,
	}
}

// GetConfigSchema returns the JSON schema for configuration validation
func (f *ResourceFilterFactory) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"allowed_protocols": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "string"},
			},
			"blocked_domains": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "string"},
			},
			"allowed_domains": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "string"},
			},
			"max_content_size": map[string]interface{}{
				"type":    "number",
				"minimum": 0,
			},
			"allowed_content_types": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "string"},
			},
			"blocked_content_types": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "string"},
			},
			"allow_private_networks": map[string]interface{}{"type": "boolean"},
			"allow_localhost":        map[string]interface{}{"type": "boolean"},
			"action": map[string]interface{}{
				"type": "string",
				"enum": []string{"block", "warn", "audit", "allow"},
			},
			"log_violations": map[string]interface{}{"type": "boolean"},
		},
	}
}
