package virtual

import (
	"fmt"

	"mcp-gateway/apps/backend/internal/types"
)

// VirtualServer implements the Server interface for MCP protocol
type VirtualServer struct {
	spec    *types.VirtualServerSpec
	adapter types.Adapter
}

// NewVirtualServer creates a new virtual server instance
func NewVirtualServer(spec *types.VirtualServerSpec) *VirtualServer {
	var adapter types.Adapter

	// Create appropriate adapter based on adapter type
	switch spec.AdapterType {
	case "REST":
		adapter = NewRESTAdapter(spec)
	default:
		// For now, default to REST adapter
		adapter = NewRESTAdapter(spec)
	}

	return &VirtualServer{
		spec:    spec,
		adapter: adapter,
	}
}

// Initialize handles the MCP initialize method
func (vs *VirtualServer) Initialize(params types.InitializeParams) (*types.InitializeResult, error) {
	// Validate protocol version
	if params.ProtocolVersion == "" {
		return nil, fmt.Errorf("protocol version is required")
	}

	// Return initialize result
	result := &types.InitializeResult{
		ProtocolVersion: "2024-11-05", // MCP protocol version
		Capabilities: map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": false,
			},
		},
		ServerInfo: types.ServerInfo{
			Name:    vs.spec.Name,
			Version: "1.0.0",
		},
	}

	return result, nil
}

// ListTools returns the available tools for this virtual server
func (vs *VirtualServer) ListTools() (*types.ListToolsResult, error) {
	toolDefs, err := vs.adapter.ListTools()
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	// Convert tool definitions to MCP tool format
	var tools []types.Tool
	for _, toolDef := range toolDefs {
		tool := types.Tool{
			Name:        toolDef.Name,
			Description: toolDef.Description,
			InputSchema: toolDef.InputSchema,
		}
		tools = append(tools, tool)
	}

	result := &types.ListToolsResult{
		Tools: tools,
	}

	return result, nil
}

// CallTool executes a tool call
func (vs *VirtualServer) CallTool(name string, args map[string]interface{}) (*types.CallToolResult, error) {
	// Delegate to the adapter
	response, err := vs.adapter.CallTool(name, args)
	if err != nil {
		return &types.CallToolResult{
			Content: []types.ToolContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Error calling tool %s: %v", name, err),
				},
			},
			IsError: true,
		}, nil
	}

	// Convert response to MCP format
	content := vs.formatResponse(response)

	result := &types.CallToolResult{
		Content: content,
		IsError: false,
	}

	return result, nil
}

// formatResponse converts adapter response to MCP tool content format
func (vs *VirtualServer) formatResponse(response interface{}) []types.ToolContent {
	// Convert response to JSON string for now
	switch resp := response.(type) {
	case string:
		return []types.ToolContent{
			{
				Type: "text",
				Text: resp,
			},
		}
	case map[string]interface{}, []interface{}:
		// Convert to pretty JSON
		jsonStr := fmt.Sprintf("%+v", resp)
		return []types.ToolContent{
			{
				Type: "text",
				Text: jsonStr,
			},
		}
	default:
		return []types.ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("%v", response),
			},
		}
	}
}
