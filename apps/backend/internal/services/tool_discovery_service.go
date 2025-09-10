package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/database/models"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/transport"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
)

// ToolDiscoveryService handles discovery of tools from MCP servers
type ToolDiscoveryService struct {
	toolModel        *models.MCPToolModel
	serverRepo       ServerRepository
	transportManager *transport.Manager
}

// ServerRepository interface for server operations
type ServerRepository interface {
	GetByID(ctx context.Context, id string) (*models.MCPServer, error)
}

// NewToolDiscoveryService creates a new tool discovery service
func NewToolDiscoveryService(toolModel *models.MCPToolModel, serverRepo ServerRepository, transportManager *transport.Manager) *ToolDiscoveryService {
	return &ToolDiscoveryService{
		toolModel:        toolModel,
		serverRepo:       serverRepo,
		transportManager: transportManager,
	}
}

// DiscoverServerTools discovers and stores tools from an MCP server using namespace service integration
func (s *ToolDiscoveryService) DiscoverServerTools(ctx context.Context, serverID uuid.UUID, organizationID uuid.UUID) error {
	// Get server configuration
	server, err := s.serverRepo.GetByID(ctx, serverID.String())
	if err != nil {
		return fmt.Errorf("failed to get server config: %w", err)
	}

	// Attempt real MCP tool discovery
	tools, discoveryErr := s.discoverRealMCPTools(ctx, server)

	// If real discovery fails, log error and continue with empty tools list
	if discoveryErr != nil {
		log.Printf("Real MCP discovery failed for server %s: %v. No tools will be registered.", server.Name, discoveryErr)
		return discoveryErr // Return error instead of continuing with empty tools
	}

	// Store discovered tools
	now := time.Now()
	for _, tool := range tools {
		mcpTool := &models.MCPTool{
			OrganizationID:     organizationID,
			ServerID:           uuid.NullUUID{UUID: serverID, Valid: true},
			Name:               tool.Name,
			FunctionName:       tool.Name, // MCP tools use name as function name
			Category:           s.categorizeToolByName(tool.Name),
			ImplementationType: "discovered",
			SourceType:         "discovered",
			TimeoutSeconds:     30,
			MaxRetries:         3,
			UsageCount:         0,
			IsActive:           true,
			IsPublic:           false,
			Schema:             tool.InputSchema,
			LastDiscoveredAt:   &now,
			DiscoveryMetadata: map[string]interface{}{
				"server_id":     serverID.String(),
				"server_name":   server.Name,
				"discovered_at": now.Format(time.RFC3339),
				"protocol":      server.Protocol,
			},
		}

		// Set description if available
		if tool.Description != "" {
			mcpTool.Description.String = tool.Description
			mcpTool.Description.Valid = true
		}

		// Upsert the tool (create or update if exists)
		if err := s.toolModel.UpsertDiscoveredTool(mcpTool); err != nil {
			log.Printf("Warning: failed to upsert discovered tool %s: %v", tool.Name, err)
		}
	}

	log.Printf("Successfully discovered %d tools from server %s", len(tools), server.Name)
	return nil
}

// RefreshServerTools refreshes tools for a specific server
func (s *ToolDiscoveryService) RefreshServerTools(ctx context.Context, serverID uuid.UUID, organizationID uuid.UUID) error {
	// Delete existing discovered tools for this server
	if err := s.toolModel.DeleteDiscoveredTools(serverID); err != nil {
		log.Printf("Warning: failed to delete existing discovered tools for server %s: %v", serverID, err)
	}

	// Discover tools again
	return s.DiscoverServerTools(ctx, serverID, organizationID)
}

// GetDiscoveredToolsForServer gets all discovered tools for a server
func (s *ToolDiscoveryService) GetDiscoveredToolsForServer(serverID uuid.UUID) ([]*models.MCPTool, error) {
	return s.toolModel.GetByServerID(serverID)
}

// discoverRealMCPTools attempts to discover tools from a real MCP server using transport layer
func (s *ToolDiscoveryService) discoverRealMCPTools(ctx context.Context, server *models.MCPServer) ([]types.MCPTool, error) {
	// Set a shorter timeout for tool discovery to prevent long hangs on non-MCP servers
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var tools []types.MCPTool

	// Check if transport manager is available
	if s.transportManager == nil {
		return nil, fmt.Errorf("transport manager not available for real MCP discovery")
	}

	// Determine appropriate transport type based on server protocol
	transportType := s.getTransportTypeForServer(server)
	if transportType == "" {
		return nil, fmt.Errorf("unsupported protocol %s for server %s", server.Protocol, server.Name)
	}

	// Create transport configuration for this server
	config := s.buildTransportConfig(server)

	// Create transport connection
	transport, session, err := s.transportManager.CreateConnectionWithConfig(
		ctx,
		transportType,
		"system", // system user for tool discovery
		server.OrganizationID.String(),
		server.ID.String(),
		config,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MCP server")
	}

	// Ensure we clean up the connection
	defer func() {
		if session != nil {
			if err := s.transportManager.CloseConnection(session.ID); err != nil {
				log.Printf("Warning: failed to close transport connection %s: %v", session.ID, err)
			}
		}
	}()

	// Connect to the server
	if err := transport.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to establish connection to MCP server")
	}

	// Quick check: For STDIO transport, verify the command exists and is executable
	if transportType == types.TransportTypeSTDIO {
		if server.Command.Valid && server.Command.String != "" {
			// Basic validation that the command exists
			log.Printf("Attempting tool discovery on STDIO server with command: %s", server.Command.String)
		} else {
			return nil, fmt.Errorf("STDIO server missing command configuration")
		}
	}

	// Send a generic MCP list tools request
	request := &types.MCPMessage{
		ID:      uuid.New().String(),
		Type:    types.MCPMessageTypeRequest,
		Method:  types.MCPMethodListTools,
		Version: "2024-11-05",
		Params:  map[string]interface{}{},
	}

	var mcpResponse *types.MCPMessage

	// For STDIO transport, use synchronous request-response pattern
	if transportType == types.TransportTypeSTDIO {
		// Check if the transport has a SendRequest method using reflection/interface
		type syncTransport interface {
			SendRequest(ctx context.Context, request *types.MCPMessage) (*types.MCPMessage, error)
		}

		if syncTrans, ok := transport.(syncTransport); ok {
			mcpResponse, err = syncTrans.SendRequest(ctx, request)
			if err != nil {
				// Provide helpful error messages based on the error type
				if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "context deadline exceeded") {
					return nil, fmt.Errorf("server did not respond to MCP protocol request - this may not be an MCP-compatible server")
				}
				// Sanitize other error messages for API response
				return nil, fmt.Errorf("failed to communicate with MCP server: connection error")
			}
		} else {
			// Fallback to async pattern (though this will likely fail for STDIO)
			if err := transport.SendMessage(ctx, request); err != nil {
				return nil, fmt.Errorf("failed to communicate with MCP server: send error")
			}

			response, err := transport.ReceiveMessage(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to communicate with MCP server: receive error")
			}

			if resp, ok := response.(*types.MCPMessage); ok {
				mcpResponse = resp
			} else {
				return nil, fmt.Errorf("received unexpected response from MCP server")
			}
		}
	} else {
		// For other transport types, use async message sending
		if err := transport.SendMessage(ctx, request); err != nil {
			return nil, fmt.Errorf("failed to communicate with MCP server: send error")
		}

		response, err := transport.ReceiveMessage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to communicate with MCP server: receive error")
		}

		// Type check the response
		if resp, ok := response.(*types.MCPMessage); ok {
			mcpResponse = resp
		} else {
			return nil, fmt.Errorf("received unexpected response from MCP server")
		}
	}

	// Parse the response to extract tools
	tools, err = s.parseMCPToolsResponse(mcpResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to process MCP server response")
	}

	log.Printf("Successfully discovered %d real tools from server %s", len(tools), server.Name)
	return tools, nil
}

// getTransportTypeForServer determines the appropriate transport type for a server
func (s *ToolDiscoveryService) getTransportTypeForServer(server *models.MCPServer) types.TransportType {
	switch strings.ToLower(server.Protocol) {
	case "stdio":
		return types.TransportTypeSTDIO
	case "http", "https":
		return types.TransportTypeHTTP
	case "websocket", "ws", "wss":
		return types.TransportTypeWebSocket
	case "sse":
		return types.TransportTypeSSE
	default:
		return ""
	}
}

// buildTransportConfig creates transport configuration for a server
func (s *ToolDiscoveryService) buildTransportConfig(server *models.MCPServer) map[string]interface{} {
	config := make(map[string]interface{})

	switch strings.ToLower(server.Protocol) {
	case "stdio":
		if server.Command.Valid {
			config["command"] = server.Command.String
		}
		if len(server.Args) > 0 {
			config["args"] = server.Args
		}
		if server.WorkingDir.Valid {
			config["working_dir"] = server.WorkingDir.String
		}
		if len(server.Environment) > 0 {
			// Convert environment array to map
			envMap := make(map[string]string)
			for _, env := range server.Environment {
				parts := strings.SplitN(env, "=", 2)
				if len(parts) == 2 {
					envMap[parts[0]] = parts[1]
				}
			}
			config["env"] = envMap
		}
	case "http", "https":
		if server.URL.Valid {
			config["url"] = server.URL.String
		}
	case "websocket", "ws", "wss":
		if server.URL.Valid {
			config["url"] = server.URL.String
		}
	case "sse":
		if server.URL.Valid {
			config["url"] = server.URL.String
		}
	}

	// Add common configuration
	config["timeout"] = time.Duration(server.TimeoutSeconds) * time.Second

	return config
}

// parseMCPToolsResponse parses an MCP tools/list response into tool definitions
func (s *ToolDiscoveryService) parseMCPToolsResponse(response *types.MCPMessage) ([]types.MCPTool, error) {
	var tools []types.MCPTool

	if response.Error != nil {
		return nil, fmt.Errorf("MCP error: %s", response.Error.Message)
	}

	// The result should contain a "tools" array
	if response.Result == nil {
		return tools, nil // No tools available
	}

	resultMap, ok := response.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result format: %T", response.Result)
	}

	toolsArray, exists := resultMap["tools"]
	if !exists {
		return tools, nil // No tools in response
	}

	toolsList, ok := toolsArray.([]interface{})
	if !ok {
		return nil, fmt.Errorf("tools field is not an array: %T", toolsArray)
	}

	// Parse each tool
	for _, toolInterface := range toolsList {
		toolMap, ok := toolInterface.(map[string]interface{})
		if !ok {
			log.Printf("Warning: skipping non-object tool: %T", toolInterface)
			continue
		}

		tool := types.MCPTool{}

		// Extract name (required)
		if name, exists := toolMap["name"]; exists {
			if nameStr, ok := name.(string); ok {
				tool.Name = nameStr
			}
		}

		// Extract description
		if desc, exists := toolMap["description"]; exists {
			if descStr, ok := desc.(string); ok {
				tool.Description = descStr
			}
		}

		// Extract input schema
		if schema, exists := toolMap["inputSchema"]; exists {
			if schemaMap, ok := schema.(map[string]interface{}); ok {
				tool.InputSchema = schemaMap
			}
		}

		// Skip tools without a name
		if tool.Name == "" {
			log.Printf("Warning: skipping tool without name")
			continue
		}

		tools = append(tools, tool)
	}

	return tools, nil
}

// categorizeToolByName attempts to categorize a tool based on its name
func (s *ToolDiscoveryService) categorizeToolByName(name string) string {
	nameLower := strings.ToLower(name)

	// File operations
	if strings.Contains(nameLower, "file") || strings.Contains(nameLower, "read") ||
		strings.Contains(nameLower, "write") || strings.Contains(nameLower, "directory") {
		return types.ToolCategoryFile
	}

	// Web/HTTP operations
	if strings.Contains(nameLower, "http") || strings.Contains(nameLower, "web") ||
		strings.Contains(nameLower, "fetch") || strings.Contains(nameLower, "request") {
		return types.ToolCategoryWeb
	}

	// Database operations
	if strings.Contains(nameLower, "db") || strings.Contains(nameLower, "database") ||
		strings.Contains(nameLower, "sql") || strings.Contains(nameLower, "query") {
		return types.ToolCategoryData
	}

	// System operations
	if strings.Contains(nameLower, "system") || strings.Contains(nameLower, "exec") ||
		strings.Contains(nameLower, "command") || strings.Contains(nameLower, "shell") {
		return types.ToolCategorySystem
	}

	// AI/ML operations
	if strings.Contains(nameLower, "ai") || strings.Contains(nameLower, "ml") ||
		strings.Contains(nameLower, "model") || strings.Contains(nameLower, "predict") {
		return types.ToolCategoryAI
	}

	// Default to general
	return types.ToolCategoryGeneral
}
