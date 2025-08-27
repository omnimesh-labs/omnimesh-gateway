package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"mcp-gateway/apps/backend/internal/database/repositories"
	"mcp-gateway/apps/backend/internal/types"

	"github.com/jmoiron/sqlx"
)

// NamespaceService handles namespace operations
type NamespaceService struct {
	repo            *repositories.NamespaceRepository
	serverRepo      *repositories.MCPServerRepository
	sessionPool     *NamespaceSessionPool
	endpointService *EndpointService
	toolPrefixCache sync.Map // Cache for prefixed tool names
}

// NewNamespaceService creates a new namespace service
func NewNamespaceService(db *sql.DB, endpointService *EndpointService) *NamespaceService {
	// Wrap the sql.DB with sqlx
	sqlxDB := sqlx.NewDb(db, "postgres")

	return &NamespaceService{
		repo:            repositories.NewNamespaceRepository(sqlxDB),
		endpointService: endpointService,
		serverRepo:  repositories.NewMCPServerRepository(sqlxDB),
		sessionPool: NewNamespaceSessionPool(),
	}
}

// CreateNamespace creates a new namespace
func (s *NamespaceService) CreateNamespace(ctx context.Context, req types.CreateNamespaceRequest) (*types.Namespace, error) {
	// Validate namespace name
	if err := s.validateNamespaceName(req.Name); err != nil {
		return nil, err
	}

	// Check if namespace already exists
	existing, _ := s.repo.GetByName(ctx, req.OrganizationID, req.Name)
	if existing != nil {
		return nil, fmt.Errorf("namespace with name %s already exists", req.Name)
	}

	// Create namespace
	namespace := &types.Namespace{
		OrganizationID: req.OrganizationID,
		Name:           req.Name,
		Description:    req.Description,
		CreatedBy:      req.CreatedBy,
		IsActive:       true,
		Metadata:       req.Metadata,
	}

	if err := s.repo.Create(ctx, namespace); err != nil {
		return nil, fmt.Errorf("failed to create namespace: %w", err)
	}

	// Add servers if provided
	for _, serverID := range req.Servers {
		if err := s.repo.AddServer(ctx, namespace.ID, serverID, 0); err != nil {
			// Log error but don't fail namespace creation
			fmt.Printf("Warning: failed to add server %s to namespace: %v\n", serverID, err)
		}
	}

	return namespace, nil
}

// GetNamespace retrieves a namespace by ID
func (s *NamespaceService) GetNamespace(ctx context.Context, id string) (*types.Namespace, error) {
	namespace, err := s.repo.GetByIDWithServers(ctx, id)
	if err != nil {
		return nil, err
	}
	// Get tools for the namespace
	tools, err := s.AggregateTools(ctx, id)
	if err != nil {
		// Log error but don't fail namespace retrieval
		fmt.Printf("Warning: failed to aggregate tools for namespace %s: %v\n", id, err)
	} else {
		namespace.Tools = tools
	}
	
	// Get endpoint for the namespace if endpoint service is available
	if s.endpointService != nil {
		endpoint, err := s.endpointService.GetEndpointByNamespace(ctx, id)
		if err != nil {
			// Log error but don't fail namespace retrieval
			fmt.Printf("Warning: failed to get endpoint for namespace %s: %v\n", id, err)
		} else {
			namespace.Endpoint = endpoint
		}
	}

	return namespace, nil
}

// ListNamespaces lists all namespaces for an organization
func (s *NamespaceService) ListNamespaces(ctx context.Context, orgID string) ([]*types.Namespace, error) {
	return s.repo.ListWithServerCount(ctx, orgID)
}

// UpdateNamespace updates a namespace
func (s *NamespaceService) UpdateNamespace(ctx context.Context, id string, req types.UpdateNamespaceRequest) (*types.Namespace, error) {
	namespace, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != "" {
		if err := s.validateNamespaceName(req.Name); err != nil {
			return nil, err
		}
		namespace.Name = req.Name
	}

	if req.Description != "" {
		namespace.Description = req.Description
	}

	if req.IsActive != nil {
		namespace.IsActive = *req.IsActive
	}

	if req.Metadata != nil {
		namespace.Metadata = req.Metadata
	}

	if err := s.repo.Update(ctx, namespace); err != nil {
		return nil, err
	}

	// Update server associations if provided
	if req.ServerIDs != nil {
		fmt.Printf("DEBUG: Updating servers for namespace %s with server IDs: %v\n", id, req.ServerIDs)
		// Get current servers
		currentServers, err := s.repo.GetServers(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get current servers: %w", err)
		}
		fmt.Printf("DEBUG: Current servers in namespace: %d\n", len(currentServers))

		// Create maps for easier comparison
		currentServerMap := make(map[string]bool)
		for _, server := range currentServers {
			currentServerMap[server.ServerID] = true
		}

		newServerMap := make(map[string]bool)
		for _, serverID := range req.ServerIDs {
			newServerMap[serverID] = true
		}

		// Remove servers that are no longer in the list
		for _, server := range currentServers {
			if !newServerMap[server.ServerID] {
				if err := s.repo.RemoveServer(ctx, id, server.ServerID); err != nil {
					fmt.Printf("Warning: failed to remove server %s from namespace: %v\n", server.ServerID, err)
				}
			}
		}

		// Add new servers
		for _, serverID := range req.ServerIDs {
			if !currentServerMap[serverID] {
				// Verify server exists before adding
				_, err := s.serverRepo.GetByID(ctx, serverID)
				if err == nil {
					if err := s.repo.AddServer(ctx, id, serverID, 0); err != nil {
						fmt.Printf("ERROR: failed to add server %s to namespace %s: %v\n", serverID, id, err)
					} else {
						fmt.Printf("SUCCESS: Added server %s to namespace %s\n", serverID, id)
					}
				} else {
					fmt.Printf("ERROR: server %s not found: %v, skipping\n", serverID, err)
				}
			} else {
				fmt.Printf("DEBUG: Server %s already in namespace %s, skipping\n", serverID, id)
			}
		}
	}

	// Clear cache for this namespace
	s.clearToolCache(id)

	// Always get the updated namespace with servers
	updatedNamespace, err := s.repo.GetByIDWithServers(ctx, id)
	if err != nil {
		// If we can't get servers, at least return the updated namespace
		return s.repo.GetByID(ctx, id)
	}

	return updatedNamespace, nil
}

// DeleteNamespace deletes a namespace
func (s *NamespaceService) DeleteNamespace(ctx context.Context, id string) error {
	// Clear sessions for this namespace
	s.sessionPool.ClearNamespace(id)

	// Clear cache
	s.clearToolCache(id)

	return s.repo.Delete(ctx, id)
}

// AddServerToNamespace adds a server to a namespace
func (s *NamespaceService) AddServerToNamespace(ctx context.Context, namespaceID string, req types.AddServerToNamespaceRequest) error {
	// Verify server exists
	_, err := s.serverRepo.GetByID(ctx, req.ServerID)
	if err != nil {
		return fmt.Errorf("server not found: %w", err)
	}

	// Add server to namespace
	if err := s.repo.AddServer(ctx, namespaceID, req.ServerID, req.Priority); err != nil {
		return err
	}

	// Clear cache for this namespace
	s.clearToolCache(namespaceID)

	return nil
}

// RemoveServerFromNamespace removes a server from a namespace
func (s *NamespaceService) RemoveServerFromNamespace(ctx context.Context, namespaceID, serverID string) error {
	// Clear sessions for this server in the namespace
	s.sessionPool.ClearServer(namespaceID, serverID)

	// Remove server from namespace
	if err := s.repo.RemoveServer(ctx, namespaceID, serverID); err != nil {
		return fmt.Errorf("failed to remove server %s from namespace %s: %w", serverID, namespaceID, err)
	}

	// Clear cache for this namespace
	s.clearToolCache(namespaceID)

	return nil
}

// UpdateServerStatus updates the status of a server in a namespace
func (s *NamespaceService) UpdateServerStatus(ctx context.Context, namespaceID, serverID string, req types.UpdateServerStatusRequest) error {
	if err := s.repo.UpdateServerStatus(ctx, namespaceID, serverID, req.Status); err != nil {
		return err
	}

	// If setting to inactive, clear sessions
	if req.Status == string(types.NamespaceStatusInactive) {
		s.sessionPool.ClearServer(namespaceID, serverID)
	}

	// Clear cache for this namespace
	s.clearToolCache(namespaceID)

	return nil
}

// AggregateTools aggregates tools from all active servers in a namespace
func (s *NamespaceService) AggregateTools(ctx context.Context, namespaceID string) ([]types.NamespaceTool, error) {
	// Check cache first
	if cached, ok := s.toolPrefixCache.Load(namespaceID); ok {
		return cached.([]types.NamespaceTool), nil
	}

	// Get servers in namespace
	servers, err := s.repo.GetServers(ctx, namespaceID)
	if err != nil {
		return nil, err
	}

	var tools []types.NamespaceTool
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Fetch tools from each active server in parallel
	for _, server := range servers {
		if server.Status != string(types.NamespaceStatusActive) {
			continue
		}

		wg.Add(1)
		go func(srv types.NamespaceServer) {
			defer wg.Done()

			// Get or create session for this server
			session, err := s.sessionPool.GetSession(namespaceID, srv.ServerID)
			if err != nil {
				fmt.Printf("Warning: failed to get session for server %s: %v\n", srv.ServerID, err)
				return
			}

			// Get tools from server
			serverTools, err := s.getServerTools(ctx, session, srv.ServerID)
			if err != nil {
				fmt.Printf("Warning: failed to get tools from server %s: %v\n", srv.ServerID, err)
				return
			}

			// Apply prefixing and add to list
			mu.Lock()
			for _, tool := range serverTools {
				prefixedTool := types.NamespaceTool{
					ServerID:     srv.ServerID,
					ServerName:   srv.ServerName,
					ToolName:     tool.Name,
					PrefixedName: PrefixToolName(srv.ServerName, tool.Name),
					Status:       string(types.NamespaceStatusActive),
					Description:  tool.Description,
				}
				tools = append(tools, prefixedTool)
			}
			mu.Unlock()
		}(server)
	}

	wg.Wait()

	// Apply tool mappings (filter inactive tools)
	toolMappings, err := s.repo.GetTools(ctx, namespaceID)
	if err == nil {
		// Create a map for quick lookup
		inactiveTools := make(map[string]bool)
		for _, mapping := range toolMappings {
			if mapping.Status == string(types.NamespaceStatusInactive) {
				key := fmt.Sprintf("%s:%s", mapping.ServerID, mapping.ToolName)
				inactiveTools[key] = true
			}
		}

		// Filter out inactive tools
		var filteredTools []types.NamespaceTool
		for _, tool := range tools {
			key := fmt.Sprintf("%s:%s", tool.ServerID, tool.ToolName)
			if !inactiveTools[key] {
				filteredTools = append(filteredTools, tool)
			}
		}
		tools = filteredTools
	}

	// Cache the result
	s.toolPrefixCache.Store(namespaceID, tools)

	return tools, nil
}

// ExecuteTool executes a tool in the namespace
func (s *NamespaceService) ExecuteTool(ctx context.Context, namespaceID string, req types.ExecuteNamespaceToolRequest) (*types.NamespaceToolResult, error) {
	// Parse prefixed tool name
	serverName, toolName, err := ParsePrefixedToolName(req.Tool)
	if err != nil {
		return &types.NamespaceToolResult{
			Success: false,
			Error:   fmt.Sprintf("invalid tool name format: %v", err),
		}, nil
	}

	// Find the server by name
	servers, err := s.repo.GetServers(ctx, namespaceID)
	if err != nil {
		return &types.NamespaceToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to get servers: %v", err),
		}, nil
	}

	var targetServer *types.NamespaceServer
	for _, server := range servers {
		if SanitizeServerName(server.ServerName) == serverName {
			targetServer = &server
			break
		}
	}

	if targetServer == nil {
		return &types.NamespaceToolResult{
			Success: false,
			Error:   fmt.Sprintf("server not found for tool %s", req.Tool),
		}, nil
	}

	// Check if server is active
	if targetServer.Status != string(types.NamespaceStatusActive) {
		return &types.NamespaceToolResult{
			Success: false,
			Error:   fmt.Sprintf("server %s is not active", targetServer.ServerName),
		}, nil
	}

	// Get session for the server
	session, err := s.sessionPool.GetSession(namespaceID, targetServer.ServerID)
	if err != nil {
		return &types.NamespaceToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to get session: %v", err),
		}, nil
	}

	// Execute the tool
	result, err := s.executeToolOnServer(ctx, session, toolName, req.Arguments)
	if err != nil {
		return &types.NamespaceToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &types.NamespaceToolResult{
		Success: true,
		Result:  result,
	}, nil
}

// UpdateToolStatus updates the status of a tool in a namespace
func (s *NamespaceService) UpdateToolStatus(ctx context.Context, namespaceID, serverID, toolName string, req types.UpdateToolStatusRequest) error {
	if err := s.repo.SetToolStatus(ctx, namespaceID, serverID, toolName, req.Status); err != nil {
		return err
	}

	// Clear cache for this namespace
	s.clearToolCache(namespaceID)

	return nil
}

// Private helper methods

func (s *NamespaceService) validateNamespaceName(name string) error {
	if len(name) < 3 || len(name) > 50 {
		return fmt.Errorf("namespace name must be between 3 and 50 characters")
	}

	// Name should only contain alphanumeric, underscore, and hyphen
	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '_' || ch == '-') {
			return fmt.Errorf("namespace name can only contain alphanumeric characters, underscores, and hyphens")
		}
	}

	return nil
}

func (s *NamespaceService) clearToolCache(namespaceID string) {
	s.toolPrefixCache.Delete(namespaceID)
}

func (s *NamespaceService) getServerTools(ctx context.Context, session *Session, serverID string) ([]types.Tool, error) {
	// This would normally call the transport layer to get tools from the server
	// For now, returning empty list as transport implementation is pending
	return []types.Tool{}, nil
}

func (s *NamespaceService) executeToolOnServer(ctx context.Context, session *Session, toolName string, args map[string]interface{}) (interface{}, error) {
	// This would normally call the transport layer to execute the tool
	// For now, returning a mock response
	return map[string]interface{}{
		"message": fmt.Sprintf("Executed tool %s", toolName),
		"args":    args,
	}, nil
}

// SanitizeServerName sanitizes a server name for use in tool prefixing
func SanitizeServerName(name string) string {
	// Replace spaces and special characters with underscores
	var result strings.Builder
	for _, ch := range name {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '_' || ch == '-' {
			result.WriteRune(ch)
		} else {
			result.WriteRune('_')
		}
	}
	return result.String()
}

// PrefixToolName creates a prefixed tool name
func PrefixToolName(serverName, toolName string) string {
	return fmt.Sprintf("%s__%s", SanitizeServerName(serverName), toolName)
}

// ParsePrefixedToolName parses a prefixed tool name
func ParsePrefixedToolName(prefixed string) (serverName, toolName string, err error) {
	parts := strings.SplitN(prefixed, "__", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid prefixed tool name format")
	}
	return parts[0], parts[1], nil
}
