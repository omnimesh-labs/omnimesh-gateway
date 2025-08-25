package openai_mod

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"mcp-gateway/apps/backend/internal/plugins/shared"
)

// OpenAIModerationPlugin implements the OpenAI Moderation API filter
type OpenAIModerationPlugin struct {
	*shared.BasePlugin
	config *OpenAIModerationConfig
	client *http.Client
}

// OpenAIModerationConfig holds the configuration for the OpenAI Moderation plugin
type OpenAIModerationConfig struct {
	CategoryThresholds map[string]float64 `json:"category_thresholds"`
	APIEndpoint        string             `json:"api_endpoint"`
	APIKey             string             `json:"api_key"`
	Model              string             `json:"model"`
	Action             string             `json:"action"`
	TimeoutSeconds     int                `json:"timeout_seconds"`
	MaxRetries         int                `json:"max_retries"`
	Threshold          float64            `json:"threshold"`
	LogViolations      bool               `json:"log_violations"`
}

// OpenAIModerationRequest represents a request to OpenAI Moderation API
type OpenAIModerationRequest struct {
	Input string `json:"input"`
	Model string `json:"model,omitempty"`
}

// OpenAIModerationResponse represents a response from OpenAI Moderation API
type OpenAIModerationResponse struct {
	ID      string                   `json:"id"`
	Model   string                   `json:"model"`
	Results []OpenAIModerationResult `json:"results"`
}

// OpenAIModerationResult represents a single moderation result
type OpenAIModerationResult struct {
	CategoryScores ModerationCategoryScores `json:"category_scores"`
	Categories     ModerationCategories     `json:"categories"`
	Flagged        bool                     `json:"flagged"`
}

// ModerationCategories represents flagged categories
type ModerationCategories struct {
	Sexual                bool `json:"sexual"`
	Hate                  bool `json:"hate"`
	Harassment            bool `json:"harassment"`
	SelfHarm              bool `json:"self-harm"`
	SexualMinors          bool `json:"sexual/minors"`
	HateThreatening       bool `json:"hate/threatening"`
	ViolenceGraphic       bool `json:"violence/graphic"`
	SelfHarmIntent        bool `json:"self-harm/intent"`
	SelfHarmInstructions  bool `json:"self-harm/instructions"`
	HarassmentThreatening bool `json:"harassment/threatening"`
	Violence              bool `json:"violence"`
}

// ModerationCategoryScores represents confidence scores for categories
type ModerationCategoryScores struct {
	Sexual                float64 `json:"sexual"`
	Hate                  float64 `json:"hate"`
	Harassment            float64 `json:"harassment"`
	SelfHarm              float64 `json:"self-harm"`
	SexualMinors          float64 `json:"sexual/minors"`
	HateThreatening       float64 `json:"hate/threatening"`
	ViolenceGraphic       float64 `json:"violence/graphic"`
	SelfHarmIntent        float64 `json:"self-harm/intent"`
	SelfHarmInstructions  float64 `json:"self-harm/instructions"`
	HarassmentThreatening float64 `json:"harassment/threatening"`
	Violence              float64 `json:"violence"`
}

// OpenAIModerationViolation represents a specific violation found by OpenAI Moderation
type OpenAIModerationViolation struct {
	Category   string  `json:"category"`
	Severity   string  `json:"severity"`
	Confidence float64 `json:"confidence"`
	Flagged    bool    `json:"flagged"`
}

// NewOpenAIModerationPlugin creates a new OpenAI Moderation plugin instance
func NewOpenAIModerationPlugin(name string, config map[string]interface{}) (*OpenAIModerationPlugin, error) {
	basePlugin := shared.NewBasePlugin(shared.PluginTypeOpenAIMod, name, 5) // Higher priority for safety

	// Set capabilities for AI middleware
	basePlugin.SetCapabilities(shared.PluginCapabilities{
		SupportsInbound:       true,
		SupportsOutbound:      true,
		SupportsPreTool:       true,
		SupportsPostTool:      true,
		SupportsModification:  false, // OpenAI Moderation blocks, doesn't modify
		SupportsBlocking:      true,
		SupportedContentTypes: []string{"text/plain", "application/json", "*"},
		SupportsRealtime:      true,
		SupportsBatch:         true,
		RequiresExternalAPI:   true,
		SupportsStreaming:     false,
		SupportsTokenization:  false,
		SupportedLanguages:    []string{"en"}, // OpenAI Moderation is primarily English
	})

	plugin := &OpenAIModerationPlugin{
		BasePlugin: basePlugin,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	if err := plugin.Configure(config); err != nil {
		return nil, fmt.Errorf("failed to configure OpenAI Moderation plugin: %w", err)
	}

	return plugin, nil
}

// Apply applies the OpenAI Moderation filter to content
func (p *OpenAIModerationPlugin) Apply(ctx context.Context, pluginCtx *shared.PluginContext, content *shared.PluginContent) (*shared.PluginResult, *shared.PluginContent, error) {
	if !p.BasePlugin.IsEnabled() {
		return shared.CreatePluginResult(false, false, shared.PluginActionAllow, "", nil), content, nil
	}

	// Check execution mode
	mode := shared.PluginExecutionMode(p.GetExecutionMode())
	switch mode {
	case shared.PluginModeDisabled:
		return shared.CreatePluginResult(false, false, shared.PluginActionAllow, "plugin disabled", nil), content, nil
	case shared.PluginModeAuditOnly:
		// Continue with processing but only log, don't block
	}

	// Skip empty content
	if content.Raw == "" {
		return shared.CreatePluginResult(false, false, shared.PluginActionAllow, "", nil), content, nil
	}

	// Call OpenAI Moderation API
	violations, err := p.checkContent(ctx, content.Raw, pluginCtx)
	if err != nil {
		// In permissive mode, log error but allow request
		if mode == shared.PluginModePermissive {
			return shared.CreatePluginResult(false, false, shared.PluginActionAllow, fmt.Sprintf("OpenAI Moderation error (permissive): %v", err), nil), content, nil
		}
		return nil, content, fmt.Errorf("OpenAI Moderation API call failed: %w", err)
	}

	// Determine action based on violations and configuration
	var action shared.PluginAction
	var blocked bool
	var reason string

	if len(violations) > 0 {
		highestConfidence := 0.0
		var primaryViolation OpenAIModerationViolation

		for _, violation := range violations {
			if violation.Confidence > highestConfidence {
				highestConfidence = violation.Confidence
				primaryViolation = violation
			}
		}

		// Check if violation meets threshold
		shouldBlock := primaryViolation.Flagged
		if customThreshold, exists := p.config.CategoryThresholds[primaryViolation.Category]; exists {
			shouldBlock = primaryViolation.Confidence >= customThreshold
		} else {
			shouldBlock = primaryViolation.Confidence >= p.config.Threshold
		}

		if shouldBlock {
			switch mode {
			case shared.PluginModeEnforcing:
				switch p.config.Action {
				case "block":
					action = shared.PluginActionBlock
					blocked = true
					reason = fmt.Sprintf("OpenAI Moderation detected unsafe content: %s (confidence: %.3f)", primaryViolation.Category, primaryViolation.Confidence)
				case "warn":
					action = shared.PluginActionWarn
					blocked = false
					reason = fmt.Sprintf("OpenAI Moderation warning: %s (confidence: %.3f)", primaryViolation.Category, primaryViolation.Confidence)
				default:
					action = shared.PluginActionAudit
					blocked = false
					reason = fmt.Sprintf("OpenAI Moderation audit: %s (confidence: %.3f)", primaryViolation.Category, primaryViolation.Confidence)
				}
			case shared.PluginModePermissive:
				action = shared.PluginActionAudit
				blocked = false
				reason = fmt.Sprintf("OpenAI Moderation detected unsafe content (permissive): %s (confidence: %.3f)", primaryViolation.Category, primaryViolation.Confidence)
			case shared.PluginModeAuditOnly:
				action = shared.PluginActionAudit
				blocked = false
				reason = fmt.Sprintf("OpenAI Moderation audit only: %s (confidence: %.3f)", primaryViolation.Category, primaryViolation.Confidence)
			}
		}
	} else {
		action = shared.PluginActionAllow
		blocked = false
	}

	// Convert violations to plugin violations
	pluginViolations := make([]shared.PluginViolation, len(violations))
	for i, violation := range violations {
		pluginViolations[i] = shared.PluginViolation{
			Type:       violation.Category,
			Match:      content.Raw,
			Severity:   violation.Severity,
			Confidence: violation.Confidence,
			Category:   violation.Category,
			Metadata: map[string]interface{}{
				"flagged":      violation.Flagged,
				"model":        p.config.Model,
				"api_endpoint": p.config.APIEndpoint,
			},
		}
	}

	result := shared.CreatePluginResult(blocked, false, action, reason, pluginViolations)
	result.PluginName = p.GetName()
	result.PluginType = p.GetType()

	return result, content, nil
}

// Configure updates the plugin configuration
func (p *OpenAIModerationPlugin) Configure(config map[string]interface{}) error {
	// Parse configuration with defaults
	moderationConfig := &OpenAIModerationConfig{
		APIEndpoint:        shared.GetConfigValue(config, "api_endpoint", "https://api.openai.com/v1/moderations"),
		APIKey:             shared.GetConfigValue(config, "api_key", ""),
		Model:              shared.GetConfigValue(config, "model", "text-moderation-latest"),
		Action:             shared.GetConfigValue(config, "action", "block"),
		TimeoutSeconds:     shared.GetConfigValue(config, "timeout_seconds", 30),
		MaxRetries:         shared.GetConfigValue(config, "max_retries", 3),
		LogViolations:      shared.GetConfigValue(config, "log_violations", true),
		Threshold:          shared.GetConfigValue(config, "threshold", 0.5),
		CategoryThresholds: make(map[string]float64),
	}

	// Parse category thresholds if provided
	if categoryThresholds, ok := config["category_thresholds"].(map[string]interface{}); ok {
		for category, threshold := range categoryThresholds {
			if thresholdFloat, ok := threshold.(float64); ok {
				moderationConfig.CategoryThresholds[category] = thresholdFloat
			}
		}
	}

	// Validate required configuration
	if moderationConfig.APIKey == "" {
		return fmt.Errorf("api_key is required for OpenAI Moderation plugin")
	}

	p.config = moderationConfig
	p.BasePlugin.SetConfig(config)

	// Update HTTP client timeout
	p.client.Timeout = time.Duration(moderationConfig.TimeoutSeconds) * time.Second

	return p.Validate()
}

// checkContent calls the OpenAI Moderation API to check content safety
func (p *OpenAIModerationPlugin) checkContent(ctx context.Context, content string, pluginCtx *shared.PluginContext) ([]OpenAIModerationViolation, error) {
	// Prepare the API request
	request := OpenAIModerationRequest{
		Input: content,
		Model: p.config.Model,
	}

	// Make API call with retries
	var response *OpenAIModerationResponse
	var err error

	for attempt := 0; attempt < p.config.MaxRetries; attempt++ {
		response, err = p.callAPI(ctx, request)
		if err == nil {
			break
		}

		if attempt < p.config.MaxRetries-1 {
			// Exponential backoff
			backoff := time.Duration(1<<attempt) * time.Second
			time.Sleep(backoff)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("OpenAI Moderation API call failed after %d attempts: %w", p.config.MaxRetries, err)
	}

	// Parse the response
	return p.parseResponse(response)
}

// callAPI makes the actual API call to OpenAI Moderation
func (p *OpenAIModerationPlugin) callAPI(ctx context.Context, request OpenAIModerationRequest) (*OpenAIModerationResponse, error) {
	// Marshal request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.APIEndpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.config.APIKey))
	httpReq.Header.Set("User-Agent", "MCP-Gateway-OpenAI-Moderation/1.0")

	// Make the request
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response
	var response OpenAIModerationResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// parseResponse parses the OpenAI Moderation API response
func (p *OpenAIModerationPlugin) parseResponse(response *OpenAIModerationResponse) ([]OpenAIModerationViolation, error) {
	if len(response.Results) == 0 {
		return nil, fmt.Errorf("no results in OpenAI Moderation response")
	}

	result := response.Results[0]
	var violations []OpenAIModerationViolation

	// Check each category
	categories := map[string]struct {
		flagged bool
		score   float64
	}{
		"sexual":                 {result.Categories.Sexual, result.CategoryScores.Sexual},
		"hate":                   {result.Categories.Hate, result.CategoryScores.Hate},
		"harassment":             {result.Categories.Harassment, result.CategoryScores.Harassment},
		"self-harm":              {result.Categories.SelfHarm, result.CategoryScores.SelfHarm},
		"sexual/minors":          {result.Categories.SexualMinors, result.CategoryScores.SexualMinors},
		"hate/threatening":       {result.Categories.HateThreatening, result.CategoryScores.HateThreatening},
		"violence/graphic":       {result.Categories.ViolenceGraphic, result.CategoryScores.ViolenceGraphic},
		"self-harm/intent":       {result.Categories.SelfHarmIntent, result.CategoryScores.SelfHarmIntent},
		"self-harm/instructions": {result.Categories.SelfHarmInstructions, result.CategoryScores.SelfHarmInstructions},
		"harassment/threatening": {result.Categories.HarassmentThreatening, result.CategoryScores.HarassmentThreatening},
		"violence":               {result.Categories.Violence, result.CategoryScores.Violence},
	}

	for category, data := range categories {
		// Check if category is flagged or meets custom threshold
		shouldFlag := data.flagged
		if customThreshold, exists := p.config.CategoryThresholds[category]; exists {
			shouldFlag = data.score >= customThreshold
		} else {
			shouldFlag = data.score >= p.config.Threshold || data.flagged
		}

		if shouldFlag {
			// Determine severity based on category
			severity := "medium"
			if category == "sexual/minors" || category == "self-harm/intent" || category == "violence/graphic" {
				severity = "high"
			} else if category == "hate/threatening" || category == "harassment/threatening" {
				severity = "high"
			} else if category == "sexual" && data.score > 0.8 {
				severity = "high"
			}

			violations = append(violations, OpenAIModerationViolation{
				Category:   category,
				Confidence: data.score,
				Flagged:    data.flagged,
				Severity:   severity,
			})
		}
	}

	return violations, nil
}

// OpenAIModerationPluginFactory implements PluginFactory for OpenAI Moderation plugins
type OpenAIModerationPluginFactory struct{}

// Create creates a new OpenAI Moderation plugin instance
func (f *OpenAIModerationPluginFactory) Create(config map[string]interface{}) (shared.Plugin, error) {
	name := shared.GetConfigValue(config, "name", "openai-moderation-plugin")
	return NewOpenAIModerationPlugin(name, config)
}

// GetType returns the plugin type
func (f *OpenAIModerationPluginFactory) GetType() shared.PluginType {
	return shared.PluginTypeOpenAIMod
}

// GetName returns the factory name
func (f *OpenAIModerationPluginFactory) GetName() string {
	return "OpenAI Moderation Plugin"
}

// GetDescription returns the factory description
func (f *OpenAIModerationPluginFactory) GetDescription() string {
	return "OpenAI Moderation API integration for detecting harmful content across multiple categories with fine-grained control"
}

// ValidateConfig validates the configuration for OpenAI Moderation plugins
func (f *OpenAIModerationPluginFactory) ValidateConfig(config map[string]interface{}) error {
	// Check required fields
	if apiKey, ok := config["api_key"].(string); !ok || apiKey == "" {
		return fmt.Errorf("api_key is required")
	}

	// Validate threshold
	if threshold, ok := config["threshold"]; ok {
		if thresholdFloat, ok := threshold.(float64); ok {
			if thresholdFloat < 0.0 || thresholdFloat > 1.0 {
				return fmt.Errorf("threshold must be between 0.0 and 1.0")
			}
		}
	}

	// Validate category thresholds
	if categoryThresholds, ok := config["category_thresholds"].(map[string]interface{}); ok {
		for category, threshold := range categoryThresholds {
			if thresholdFloat, ok := threshold.(float64); ok {
				if thresholdFloat < 0.0 || thresholdFloat > 1.0 {
					return fmt.Errorf("category threshold for %s must be between 0.0 and 1.0", category)
				}
			} else {
				return fmt.Errorf("category threshold for %s must be a number", category)
			}
		}
	}

	// Validate action
	if action, ok := config["action"].(string); ok {
		validActions := []string{"block", "warn", "audit", "allow"}
		valid := false
		for _, validAction := range validActions {
			if action == validAction {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid action: %s, must be one of: %v", action, validActions)
		}
	}

	return nil
}

// GetDefaultConfig returns the default configuration for OpenAI Moderation plugins
func (f *OpenAIModerationPluginFactory) GetDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"api_endpoint":    "https://api.openai.com/v1/moderations",
		"model":           "text-moderation-latest",
		"action":          "block",
		"threshold":       0.5,
		"timeout_seconds": 30,
		"max_retries":     3,
		"log_violations":  true,
		"category_thresholds": map[string]float64{
			"sexual/minors":          0.1, // Very low threshold for CSAM
			"self-harm/intent":       0.3, // Lower threshold for self-harm
			"hate/threatening":       0.4, // Lower threshold for threatening content
			"harassment/threatening": 0.4, // Lower threshold for threatening content
		},
	}
}

// GetConfigSchema returns the JSON schema for configuration validation
func (f *OpenAIModerationPluginFactory) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":     "object",
		"required": []string{"api_key"},
		"properties": map[string]interface{}{
			"api_key": map[string]interface{}{
				"type":        "string",
				"description": "OpenAI API key",
			},
			"api_endpoint": map[string]interface{}{
				"type":        "string",
				"description": "OpenAI Moderation API endpoint URL",
			},
			"model": map[string]interface{}{
				"type":        "string",
				"description": "OpenAI moderation model to use",
			},
			"action": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"block", "warn", "audit", "allow"},
				"description": "Action to take when violations are found",
			},
			"threshold": map[string]interface{}{
				"type":        "number",
				"minimum":     0.0,
				"maximum":     1.0,
				"description": "Default confidence threshold for triggering actions",
			},
			"category_thresholds": map[string]interface{}{
				"type":        "object",
				"description": "Per-category confidence thresholds",
				"additionalProperties": map[string]interface{}{
					"type":    "number",
					"minimum": 0.0,
					"maximum": 1.0,
				},
			},
			"timeout_seconds": map[string]interface{}{
				"type":        "integer",
				"minimum":     1,
				"maximum":     300,
				"description": "API request timeout in seconds",
			},
			"max_retries": map[string]interface{}{
				"type":        "integer",
				"minimum":     0,
				"maximum":     10,
				"description": "Maximum number of API retries",
			},
			"log_violations": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to log violations to audit trail",
			},
		},
	}
}

// GetSupportedExecutionModes returns supported execution modes
func (f *OpenAIModerationPluginFactory) GetSupportedExecutionModes() []string {
	return []string{
		string(shared.PluginModeEnforcing),
		string(shared.PluginModePermissive),
		string(shared.PluginModeDisabled),
		string(shared.PluginModeAuditOnly),
	}
}
