package llamaguard

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"mcp-gateway/apps/backend/internal/plugins/shared"
)

// LlamaGuardPlugin implements the LlamaGuard AI safety filter
type LlamaGuardPlugin struct {
	*shared.BasePlugin
	config *LlamaGuardConfig
	client *http.Client
}

// LlamaGuardConfig holds the configuration for the LlamaGuard plugin
type LlamaGuardConfig struct {
	APIEndpoint     string   `json:"api_endpoint"`
	APIKey          string   `json:"api_key"`
	Model           string   `json:"model"`
	Categories      []string `json:"categories"`
	Action          string   `json:"action"`
	Threshold       float64  `json:"threshold"`
	TimeoutSeconds  int      `json:"timeout_seconds"`
	MaxRetries      int      `json:"max_retries"`
	LogViolations   bool     `json:"log_violations"`
	CustomPrompt    string   `json:"custom_prompt"`
	EnableStreaming bool     `json:"enable_streaming"`
}

// LlamaGuardRequest represents a request to LlamaGuard API
type LlamaGuardRequest struct {
	Model       string                 `json:"model"`
	Messages    []LlamaGuardMessage    `json:"messages"`
	MaxTokens   int                    `json:"max_tokens"`
	Temperature float64                `json:"temperature"`
	Stream      bool                   `json:"stream"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// LlamaGuardMessage represents a message in the request
type LlamaGuardMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LlamaGuardResponse represents a response from LlamaGuard API
type LlamaGuardResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// LlamaGuardViolation represents a specific violation found by LlamaGuard
type LlamaGuardViolation struct {
	Category   string  `json:"category"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
	Severity   string  `json:"severity"`
}

// Default LlamaGuard categories
var DefaultCategories = []string{
	"S1: Violent Crimes",
	"S2: Non-Violent Crimes", 
	"S3: Sex Crimes",
	"S4: Child Exploitation",
	"S5: Defamation",
	"S6: Specialized Advice",
	"S7: Privacy",
	"S8: Intellectual Property",
	"S9: Indiscriminate Weapons",
	"S10: Hate",
	"S11: Self-Harm",
	"S12: Sexual Content",
	"S13: Elections",
}

// NewLlamaGuardPlugin creates a new LlamaGuard plugin instance
func NewLlamaGuardPlugin(name string, config map[string]interface{}) (*LlamaGuardPlugin, error) {
	basePlugin := shared.NewBasePlugin(shared.PluginTypeLlamaGuard, name, 5) // Higher priority for safety
	
	// Set capabilities for AI middleware
	basePlugin.SetCapabilities(shared.PluginCapabilities{
		SupportsInbound:       true,
		SupportsOutbound:      true,
		SupportsPreTool:       true,
		SupportsPostTool:      true,
		SupportsModification:  false, // LlamaGuard blocks, doesn't modify
		SupportsBlocking:      true,
		SupportedContentTypes: []string{"text/plain", "application/json", "*"},
		SupportsRealtime:      true,
		SupportsBatch:         true,
		RequiresExternalAPI:   true,
		SupportsStreaming:     true,
		SupportsTokenization:  true,
		SupportedLanguages:    []string{"en", "es", "fr", "de", "it", "pt", "ru", "ja", "ko", "zh"},
	})

	plugin := &LlamaGuardPlugin{
		BasePlugin: basePlugin,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	if err := plugin.Configure(config); err != nil {
		return nil, fmt.Errorf("failed to configure LlamaGuard plugin: %w", err)
	}

	return plugin, nil
}

// Apply applies the LlamaGuard filter to content
func (p *LlamaGuardPlugin) Apply(ctx context.Context, pluginCtx *shared.PluginContext, content *shared.PluginContent) (*shared.PluginResult, *shared.PluginContent, error) {
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

	// Call LlamaGuard API
	violations, err := p.checkContent(ctx, content.Raw, pluginCtx)
	if err != nil {
		// In permissive mode, log error but allow request
		if mode == shared.PluginModePermissive {
			return shared.CreatePluginResult(false, false, shared.PluginActionAllow, fmt.Sprintf("LlamaGuard error (permissive): %v", err), nil), content, nil
		}
		return nil, content, fmt.Errorf("LlamaGuard API call failed: %w", err)
	}

	// Determine action based on violations and configuration
	var action shared.PluginAction
	var blocked bool
	var reason string

	if len(violations) > 0 {
		highestConfidence := 0.0
		var primaryViolation LlamaGuardViolation
		
		for _, violation := range violations {
			if violation.Confidence > highestConfidence {
				highestConfidence = violation.Confidence
				primaryViolation = violation
			}
		}

		// Check if violation meets threshold
		if highestConfidence >= p.config.Threshold {
			switch mode {
			case shared.PluginModeEnforcing:
				switch p.config.Action {
				case "block":
					action = shared.PluginActionBlock
					blocked = true
					reason = fmt.Sprintf("LlamaGuard detected unsafe content: %s (confidence: %.2f)", primaryViolation.Category, primaryViolation.Confidence)
				case "warn":
					action = shared.PluginActionWarn
					blocked = false
					reason = fmt.Sprintf("LlamaGuard warning: %s (confidence: %.2f)", primaryViolation.Category, primaryViolation.Confidence)
				default:
					action = shared.PluginActionAudit
					blocked = false
					reason = fmt.Sprintf("LlamaGuard audit: %s (confidence: %.2f)", primaryViolation.Category, primaryViolation.Confidence)
				}
			case shared.PluginModePermissive:
				action = shared.PluginActionAudit
				blocked = false
				reason = fmt.Sprintf("LlamaGuard detected unsafe content (permissive): %s (confidence: %.2f)", primaryViolation.Category, primaryViolation.Confidence)
			case shared.PluginModeAuditOnly:
				action = shared.PluginActionAudit
				blocked = false
				reason = fmt.Sprintf("LlamaGuard audit only: %s (confidence: %.2f)", primaryViolation.Category, primaryViolation.Confidence)
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
				"llamaguard_reason": violation.Reason,
				"model":            p.config.Model,
				"api_endpoint":     p.config.APIEndpoint,
			},
		}
	}

	result := shared.CreatePluginResult(blocked, false, action, reason, pluginViolations)
	result.PluginName = p.GetName()
	result.PluginType = p.GetType()

	return result, content, nil
}

// Configure updates the plugin configuration
func (p *LlamaGuardPlugin) Configure(config map[string]interface{}) error {
	// Parse configuration with defaults
	llamaConfig := &LlamaGuardConfig{
		APIEndpoint:     shared.GetConfigValue(config, "api_endpoint", "https://api.together.xyz/v1/chat/completions"),
		APIKey:          shared.GetConfigValue(config, "api_key", ""),
		Model:           shared.GetConfigValue(config, "model", "meta-llama/Llama-Guard-2-8b"),
		Categories:      shared.GetConfigStringSlice(config, "categories", DefaultCategories),
		Action:          shared.GetConfigValue(config, "action", "block"),
		Threshold:       shared.GetConfigValue(config, "threshold", 0.8),
		TimeoutSeconds:  shared.GetConfigValue(config, "timeout_seconds", 30),
		MaxRetries:      shared.GetConfigValue(config, "max_retries", 3),
		LogViolations:   shared.GetConfigValue(config, "log_violations", true),
		CustomPrompt:    shared.GetConfigValue(config, "custom_prompt", ""),
		EnableStreaming: shared.GetConfigValue(config, "enable_streaming", false),
	}

	// Validate required configuration
	if llamaConfig.APIKey == "" {
		return fmt.Errorf("api_key is required for LlamaGuard plugin")
	}

	if llamaConfig.APIEndpoint == "" {
		return fmt.Errorf("api_endpoint is required for LlamaGuard plugin")
	}

	p.config = llamaConfig
	p.BasePlugin.SetConfig(config)

	// Update HTTP client timeout
	p.client.Timeout = time.Duration(llamaConfig.TimeoutSeconds) * time.Second

	return p.Validate()
}

// checkContent calls the LlamaGuard API to check content safety
func (p *LlamaGuardPlugin) checkContent(ctx context.Context, content string, pluginCtx *shared.PluginContext) ([]LlamaGuardViolation, error) {
	// Prepare the prompt for LlamaGuard
	prompt := p.buildPrompt(content)
	
	// Prepare the API request
	request := LlamaGuardRequest{
		Model: p.config.Model,
		Messages: []LlamaGuardMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   150,
		Temperature: 0.0,
		Stream:      p.config.EnableStreaming,
		Metadata: map[string]interface{}{
			"request_id":      pluginCtx.RequestID,
			"organization_id": pluginCtx.OrganizationID,
			"user_id":        pluginCtx.UserID,
		},
	}

	// Make API call with retries
	var response *LlamaGuardResponse
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
		return nil, fmt.Errorf("LlamaGuard API call failed after %d attempts: %w", p.config.MaxRetries, err)
	}

	// Parse the response
	return p.parseResponse(response)
}

// buildPrompt creates the LlamaGuard prompt
func (p *LlamaGuardPlugin) buildPrompt(content string) string {
	if p.config.CustomPrompt != "" {
		return strings.ReplaceAll(p.config.CustomPrompt, "{CONTENT}", content)
	}

	// Default LlamaGuard prompt format
	categories := strings.Join(p.config.Categories, "\n")
	
	return fmt.Sprintf(`Task: Check if there is unsafe content in 'User' messages in conversations according to our safety policy with the below categories.

<BEGIN UNSAFE CONTENT CATEGORIES>
%s
<END UNSAFE CONTENT CATEGORIES>

<BEGIN CONVERSATION>
User: %s
<END CONVERSATION>

Provide your safety assessment for the LAST user message:
- First line must contain 'safe' or 'unsafe'
- If unsafe, provide the violated category
- Explain your reasoning`, categories, content)
}

// callAPI makes the actual API call to LlamaGuard
func (p *LlamaGuardPlugin) callAPI(ctx context.Context, request LlamaGuardRequest) (*LlamaGuardResponse, error) {
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
	httpReq.Header.Set("User-Agent", "MCP-Gateway-LlamaGuard/1.0")

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
	var response LlamaGuardResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// parseResponse parses the LlamaGuard API response
func (p *LlamaGuardPlugin) parseResponse(response *LlamaGuardResponse) ([]LlamaGuardViolation, error) {
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no choices in LlamaGuard response")
	}

	content := response.Choices[0].Message.Content
	lines := strings.Split(strings.TrimSpace(content), "\n")
	
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty response from LlamaGuard")
	}

	// Parse first line for safety assessment
	firstLine := strings.ToLower(strings.TrimSpace(lines[0]))
	
	if strings.Contains(firstLine, "safe") && !strings.Contains(firstLine, "unsafe") {
		// Content is safe
		return []LlamaGuardViolation{}, nil
	}

	// Content is unsafe, parse violations
	var violations []LlamaGuardViolation
	
	// Extract category from response
	category := "Unknown"
	reason := content
	
	// Look for category indicators in the response
	for _, cat := range p.config.Categories {
		if strings.Contains(strings.ToUpper(content), strings.ToUpper(cat)) {
			category = cat
			break
		}
	}
	
	// Calculate confidence based on response certainty
	confidence := 0.9 // Default high confidence for explicit "unsafe" responses
	if strings.Contains(strings.ToLower(content), "potentially") || strings.Contains(strings.ToLower(content), "might") {
		confidence = 0.7
	}
	
	// Determine severity based on category
	severity := "medium"
	if strings.Contains(strings.ToUpper(category), "VIOLENCE") || strings.Contains(strings.ToUpper(category), "HARM") {
		severity = "high"
	} else if strings.Contains(strings.ToUpper(category), "PRIVACY") || strings.Contains(strings.ToUpper(category), "INTELLECTUAL") {
		severity = "low"
	}

	violations = append(violations, LlamaGuardViolation{
		Category:   category,
		Confidence: confidence,
		Reason:     reason,
		Severity:   severity,
	})

	return violations, nil
}

// LlamaGuardPluginFactory implements PluginFactory for LlamaGuard plugins
type LlamaGuardPluginFactory struct{}

// Create creates a new LlamaGuard plugin instance
func (f *LlamaGuardPluginFactory) Create(config map[string]interface{}) (shared.Plugin, error) {
	name := shared.GetConfigValue(config, "name", "llamaguard-plugin")
	return NewLlamaGuardPlugin(name, config)
}

// GetType returns the plugin type
func (f *LlamaGuardPluginFactory) GetType() shared.PluginType {
	return shared.PluginTypeLlamaGuard
}

// GetName returns the factory name
func (f *LlamaGuardPluginFactory) GetName() string {
	return "LlamaGuard AI Safety Plugin"
}

// GetDescription returns the factory description
func (f *LlamaGuardPluginFactory) GetDescription() string {
	return "AI-powered content safety filter using Meta's LlamaGuard model for detecting harmful content across multiple categories"
}

// ValidateConfig validates the configuration for LlamaGuard plugins
func (f *LlamaGuardPluginFactory) ValidateConfig(config map[string]interface{}) error {
	// Check required fields
	if apiKey, ok := config["api_key"].(string); !ok || apiKey == "" {
		return fmt.Errorf("api_key is required")
	}

	if endpoint, ok := config["api_endpoint"].(string); !ok || endpoint == "" {
		return fmt.Errorf("api_endpoint is required")
	}

	// Validate threshold
	if threshold, ok := config["threshold"]; ok {
		if thresholdFloat, ok := threshold.(float64); ok {
			if thresholdFloat < 0.0 || thresholdFloat > 1.0 {
				return fmt.Errorf("threshold must be between 0.0 and 1.0")
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

// GetDefaultConfig returns the default configuration for LlamaGuard plugins
func (f *LlamaGuardPluginFactory) GetDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"api_endpoint":     "https://api.together.xyz/v1/chat/completions",
		"model":            "meta-llama/Llama-Guard-2-8b",
		"categories":       DefaultCategories,
		"action":           "block",
		"threshold":        0.8,
		"timeout_seconds":  30,
		"max_retries":      3,
		"log_violations":   true,
		"enable_streaming": false,
	}
}

// GetConfigSchema returns the JSON schema for configuration validation
func (f *LlamaGuardPluginFactory) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"required": []string{"api_key", "api_endpoint"},
		"properties": map[string]interface{}{
			"api_key": map[string]interface{}{
				"type": "string",
				"description": "API key for LlamaGuard service",
			},
			"api_endpoint": map[string]interface{}{
				"type": "string", 
				"description": "LlamaGuard API endpoint URL",
			},
			"model": map[string]interface{}{
				"type": "string",
				"description": "LlamaGuard model to use",
			},
			"categories": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{"type": "string"},
				"description": "Safety categories to check",
			},
			"action": map[string]interface{}{
				"type": "string",
				"enum": []string{"block", "warn", "audit", "allow"},
				"description": "Action to take when violations are found",
			},
			"threshold": map[string]interface{}{
				"type": "number",
				"minimum": 0.0,
				"maximum": 1.0,
				"description": "Confidence threshold for triggering actions",
			},
			"timeout_seconds": map[string]interface{}{
				"type": "integer",
				"minimum": 1,
				"maximum": 300,
				"description": "API request timeout in seconds",
			},
			"max_retries": map[string]interface{}{
				"type": "integer",
				"minimum": 0,
				"maximum": 10,
				"description": "Maximum number of API retries",
			},
			"log_violations": map[string]interface{}{
				"type": "boolean",
				"description": "Whether to log violations to audit trail",
			},
			"custom_prompt": map[string]interface{}{
				"type": "string",
				"description": "Custom prompt template (use {CONTENT} for content placeholder)",
			},
			"enable_streaming": map[string]interface{}{
				"type": "boolean",
				"description": "Enable streaming responses from LlamaGuard",
			},
		},
	}
}

// GetSupportedExecutionModes returns supported execution modes
func (f *LlamaGuardPluginFactory) GetSupportedExecutionModes() []string {
	return []string{
		string(shared.PluginModeEnforcing),
		string(shared.PluginModePermissive),
		string(shared.PluginModeDisabled),
		string(shared.PluginModeAuditOnly),
	}
}