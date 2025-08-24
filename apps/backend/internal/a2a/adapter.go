package a2a

import (
	"fmt"

	"mcp-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
)

// Adapter implements the types.A2AAdapter interface for integrating A2A agents with virtual servers
type Adapter struct {
	service *Service
	client  *Client
}

// NewAdapter creates a new A2A adapter
func NewAdapter(service *Service, client *Client) *Adapter {
	return &Adapter{
		service: service,
		client:  client,
	}
}

// ListTools returns available tools from an A2A agent that can be exposed through virtual servers
func (a *Adapter) ListTools(agent *types.A2AAgent) ([]types.ToolDef, error) {
	// Check if agent supports tools capability
	if capabilities, ok := agent.CapabilitiesData[types.CapabilityTools].(bool); !ok || !capabilities {
		return []types.ToolDef{}, nil
	}

	// For different agent types, we define standard tools they can provide
	var tools []types.ToolDef

	switch agent.AgentType {
	case types.AgentTypeOpenAI:
		tools = a.getOpenAITools(agent)
	case types.AgentTypeAnthropic:
		tools = a.getAnthropicTools(agent)
	case types.AgentTypeCustom, types.AgentTypeGeneric:
		tools = a.getCustomTools(agent)
	default:
		return nil, fmt.Errorf("unsupported agent type: %s", agent.AgentType)
	}

	return tools, nil
}

// CallTool executes a tool call through an A2A agent
func (a *Adapter) CallTool(agent *types.A2AAgent, name string, args map[string]interface{}) (interface{}, error) {
	// Validate that the agent supports tools
	if capabilities, ok := agent.CapabilitiesData[types.CapabilityTools].(bool); !ok || !capabilities {
		return nil, fmt.Errorf("agent %s does not support tools", agent.Name)
	}

	// Validate that the agent is active
	if !agent.IsActive {
		return nil, fmt.Errorf("agent %s is not active", agent.Name)
	}

	// Prepare the tool call based on agent type and tool name
	switch name {
	case "chat":
		return a.callChatTool(agent, args)
	case "analyze":
		return a.callAnalyzeTool(agent, args)
	case "summarize":
		return a.callSummarizeTool(agent, args)
	case "translate":
		return a.callTranslateTool(agent, args)
	case "code_review":
		return a.callCodeReviewTool(agent, args)
	case "question_answer":
		return a.callQuestionAnswerTool(agent, args)
	default:
		// For custom tools, try to call directly
		return a.callCustomTool(agent, name, args)
	}
}

// RegisterTool registers a tool from an A2A agent to a virtual server
func (a *Adapter) RegisterTool(agentID uuid.UUID, virtualServerID uuid.UUID, toolName string, config map[string]interface{}) error {
	// Validate that the agent exists and supports tools
	agent, err := a.service.Get(agentID)
	if err != nil {
		return fmt.Errorf("failed to get agent: %w", err)
	}

	if capabilities, ok := agent.CapabilitiesData[types.CapabilityTools].(bool); !ok || !capabilities {
		return fmt.Errorf("agent %s does not support tools", agent.Name)
	}

	// Register the tool mapping
	return a.service.RegisterTool(agentID, virtualServerID, toolName, config)
}

// UnregisterTool removes a tool mapping from an A2A agent to a virtual server
func (a *Adapter) UnregisterTool(agentID uuid.UUID, virtualServerID uuid.UUID, toolName string) error {
	return a.service.UnregisterTool(agentID, virtualServerID, toolName)
}

// getOpenAITools returns standard tools available for OpenAI agents
func (a *Adapter) getOpenAITools(agent *types.A2AAgent) []types.ToolDef {
	tools := []types.ToolDef{
		{
			Name:        "chat",
			Description: "Chat with OpenAI GPT models for general conversation and assistance",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"message": map[string]interface{}{
						"type":        "string",
						"description": "The message to send to the AI agent",
					},
					"context": map[string]interface{}{
						"type":        "string",
						"description": "Optional context or system prompt",
					},
					"max_tokens": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of tokens in the response",
						"default":     1000,
					},
				},
				"required": []string{"message"},
			},
		},
		{
			Name:        "analyze",
			Description: "Analyze text, data, or documents using OpenAI capabilities",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"content": map[string]interface{}{
						"type":        "string",
						"description": "The content to analyze",
					},
					"analysis_type": map[string]interface{}{
						"type":        "string",
						"description": "Type of analysis (sentiment, summary, topics, etc.)",
						"default":     "general",
					},
				},
				"required": []string{"content"},
			},
		},
		{
			Name:        "code_review",
			Description: "Review code and provide feedback using OpenAI",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"code": map[string]interface{}{
						"type":        "string",
						"description": "The code to review",
					},
					"language": map[string]interface{}{
						"type":        "string",
						"description": "Programming language of the code",
					},
					"focus": map[string]interface{}{
						"type":        "string",
						"description": "What to focus on (bugs, performance, style, etc.)",
						"default":     "general",
					},
				},
				"required": []string{"code"},
			},
		},
	}

	// Add image analysis if the agent supports it
	if capabilities, ok := agent.CapabilitiesData[types.CapabilityImages].(bool); ok && capabilities {
		tools = append(tools, types.ToolDef{
			Name:        "analyze_image",
			Description: "Analyze and describe images using OpenAI vision capabilities",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"image_url": map[string]interface{}{
						"type":        "string",
						"description": "URL of the image to analyze",
					},
					"prompt": map[string]interface{}{
						"type":        "string",
						"description": "What to analyze about the image",
						"default":     "Describe what you see in this image",
					},
				},
				"required": []string{"image_url"},
			},
		})
	}

	return tools
}

// getAnthropicTools returns standard tools available for Anthropic agents
func (a *Adapter) getAnthropicTools(agent *types.A2AAgent) []types.ToolDef {
	return []types.ToolDef{
		{
			Name:        "chat",
			Description: "Chat with Anthropic Claude for general conversation and assistance",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"message": map[string]interface{}{
						"type":        "string",
						"description": "The message to send to Claude",
					},
					"context": map[string]interface{}{
						"type":        "string",
						"description": "Optional context or system prompt",
					},
					"max_tokens": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of tokens in the response",
						"default":     1000,
					},
				},
				"required": []string{"message"},
			},
		},
		{
			Name:        "analyze",
			Description: "Analyze text, data, or documents using Claude's analytical capabilities",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"content": map[string]interface{}{
						"type":        "string",
						"description": "The content to analyze",
					},
					"analysis_type": map[string]interface{}{
						"type":        "string",
						"description": "Type of analysis to perform",
						"default":     "comprehensive",
					},
				},
				"required": []string{"content"},
			},
		},
		{
			Name:        "summarize",
			Description: "Summarize long text using Claude's comprehension abilities",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{
						"type":        "string",
						"description": "The text to summarize",
					},
					"length": map[string]interface{}{
						"type":        "string",
						"description": "Desired summary length (short, medium, long)",
						"default":     "medium",
					},
				},
				"required": []string{"text"},
			},
		},
	}
}

// getCustomTools returns standard tools available for custom/generic agents
func (a *Adapter) getCustomTools(agent *types.A2AAgent) []types.ToolDef {
	return []types.ToolDef{
		{
			Name:        "chat",
			Description: "Chat with the custom AI agent",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"message": map[string]interface{}{
						"type":        "string",
						"description": "The message to send to the agent",
					},
					"context": map[string]interface{}{
						"type":        "string",
						"description": "Optional context",
					},
				},
				"required": []string{"message"},
			},
		},
		{
			Name:        "question_answer",
			Description: "Ask questions and get answers from the agent",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"question": map[string]interface{}{
						"type":        "string",
						"description": "The question to ask",
					},
					"context": map[string]interface{}{
						"type":        "string",
						"description": "Optional context for the question",
					},
				},
				"required": []string{"question"},
			},
		},
	}
}

// callChatTool handles chat tool calls
func (a *Adapter) callChatTool(agent *types.A2AAgent, args map[string]interface{}) (interface{}, error) {
	message, ok := args["message"].(string)
	if !ok {
		return nil, fmt.Errorf("message parameter is required and must be a string")
	}

	// Build chat messages
	messages := []types.A2AChatMessage{
		{
			Role:    "user",
			Content: message,
		},
	}

	// Add context as system message if provided
	if context, ok := args["context"].(string); ok && context != "" {
		messages = append([]types.A2AChatMessage{
			{
				Role:    "system",
				Content: context,
			},
		}, messages...)
	}

	// Prepare chat request
	chatRequest := &types.A2AChatRequest{
		Messages: messages,
	}

	// Set max_tokens if provided
	if maxTokens, ok := args["max_tokens"].(float64); ok {
		chatRequest.MaxTokens = int(maxTokens)
	}

	// Make the chat request
	response, err := a.client.Chat(agent, chatRequest)
	if err != nil {
		return nil, fmt.Errorf("chat request failed: %w", err)
	}

	// Return the response content
	if response.Message != nil {
		return map[string]interface{}{
			"content":       response.Message.Content,
			"finish_reason": response.FinishReason,
			"usage":         response.Usage,
		}, nil
	}

	return response, nil
}

// callAnalyzeTool handles analyze tool calls
func (a *Adapter) callAnalyzeTool(agent *types.A2AAgent, args map[string]interface{}) (interface{}, error) {
	content, ok := args["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content parameter is required and must be a string")
	}

	analysisType := "general"
	if at, ok := args["analysis_type"].(string); ok {
		analysisType = at
	}

	// Build analysis prompt
	prompt := fmt.Sprintf("Please analyze the following content with focus on %s:\n\n%s", analysisType, content)

	return a.callChatTool(agent, map[string]interface{}{
		"message": prompt,
	})
}

// callSummarizeTool handles summarize tool calls
func (a *Adapter) callSummarizeTool(agent *types.A2AAgent, args map[string]interface{}) (interface{}, error) {
	text, ok := args["text"].(string)
	if !ok {
		return nil, fmt.Errorf("text parameter is required and must be a string")
	}

	length := "medium"
	if l, ok := args["length"].(string); ok {
		length = l
	}

	// Build summarization prompt
	prompt := fmt.Sprintf("Please provide a %s summary of the following text:\n\n%s", length, text)

	return a.callChatTool(agent, map[string]interface{}{
		"message": prompt,
	})
}

// callTranslateTool handles translate tool calls
func (a *Adapter) callTranslateTool(agent *types.A2AAgent, args map[string]interface{}) (interface{}, error) {
	text, ok := args["text"].(string)
	if !ok {
		return nil, fmt.Errorf("text parameter is required and must be a string")
	}

	targetLang, ok := args["target_language"].(string)
	if !ok {
		return nil, fmt.Errorf("target_language parameter is required and must be a string")
	}

	sourceLang := "auto-detect"
	if sl, ok := args["source_language"].(string); ok {
		sourceLang = sl
	}

	// Build translation prompt
	prompt := fmt.Sprintf("Please translate the following text from %s to %s:\n\n%s", sourceLang, targetLang, text)

	return a.callChatTool(agent, map[string]interface{}{
		"message": prompt,
	})
}

// callCodeReviewTool handles code review tool calls
func (a *Adapter) callCodeReviewTool(agent *types.A2AAgent, args map[string]interface{}) (interface{}, error) {
	code, ok := args["code"].(string)
	if !ok {
		return nil, fmt.Errorf("code parameter is required and must be a string")
	}

	language := "unknown"
	if lang, ok := args["language"].(string); ok {
		language = lang
	}

	focus := "general"
	if f, ok := args["focus"].(string); ok {
		focus = f
	}

	// Build code review prompt
	prompt := fmt.Sprintf("Please review the following %s code with focus on %s:\n\n```%s\n%s\n```", language, focus, language, code)

	return a.callChatTool(agent, map[string]interface{}{
		"message": prompt,
	})
}

// callQuestionAnswerTool handles question-answer tool calls
func (a *Adapter) callQuestionAnswerTool(agent *types.A2AAgent, args map[string]interface{}) (interface{}, error) {
	question, ok := args["question"].(string)
	if !ok {
		return nil, fmt.Errorf("question parameter is required and must be a string")
	}

	// Use the question directly as the message
	chatArgs := map[string]interface{}{
		"message": question,
	}

	// Add context if provided
	if context, ok := args["context"].(string); ok && context != "" {
		chatArgs["context"] = context
	}

	return a.callChatTool(agent, chatArgs)
}

// callCustomTool handles custom tool calls by delegating to the agent
func (a *Adapter) callCustomTool(agent *types.A2AAgent, toolName string, args map[string]interface{}) (interface{}, error) {
	// Prepare a custom A2A request
	request := &types.A2ARequest{
		InteractionType: types.InteractionTypeTool,
		Parameters: map[string]interface{}{
			"tool_name": toolName,
			"arguments": args,
		},
		ProtocolVersion: agent.ProtocolVersion,
		AgentID:         agent.ID.String(),
	}

	// Make the invocation request
	response, err := a.client.Invoke(agent, request)
	if err != nil {
		return nil, fmt.Errorf("custom tool invocation failed: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("custom tool failed: %s", response.Error)
	}

	return response.Data, nil
}