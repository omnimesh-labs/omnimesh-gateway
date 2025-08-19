package gateway

import (
	"net/http"

	"mcp-gateway/internal/discovery"
	"mcp-gateway/internal/types"
)

// Router handles request routing to MCP servers
type Router struct {
	discovery    *discovery.Service
	loadBalancer LoadBalancer
}

// NewRouter creates a new request router
func NewRouter(discovery *discovery.Service, loadBalancer LoadBalancer) *Router {
	return &Router{
		discovery:    discovery,
		loadBalancer: loadBalancer,
	}
}

// SelectServer selects the best MCP server for a request
func (r *Router) SelectServer(orgID string, req *http.Request) (*types.MCPServer, error) {
	// Get healthy servers for organization
	servers, err := r.discovery.GetHealthyServers(orgID)
	if err != nil {
		return nil, err
	}

	if len(servers) == 0 {
		return nil, types.NewServiceUnavailableError("No healthy servers available")
	}

	// Apply routing rules if any
	filteredServers := r.applyRoutingRules(servers, req)
	if len(filteredServers) == 0 {
		return nil, types.NewServiceUnavailableError("No servers match routing rules")
	}

	// Use load balancer to select server
	return r.loadBalancer.SelectServer(filteredServers)
}

// applyRoutingRules applies routing rules to filter servers
func (r *Router) applyRoutingRules(servers []*types.MCPServer, req *http.Request) []*types.MCPServer {
	// TODO: Implement routing rules
	// Filter servers based on:
	// - Request path patterns
	// - Request headers
	// - Server capabilities
	// - Custom routing policies

	// For now, return all servers
	return servers
}

// GetRoutingRules retrieves routing rules for an organization
func (r *Router) GetRoutingRules(orgID string) ([]*RoutingRule, error) {
	// TODO: Implement routing rule retrieval from database
	return nil, nil
}

// CreateRoutingRule creates a new routing rule
func (r *Router) CreateRoutingRule(orgID string, rule *RoutingRule) error {
	// TODO: Implement routing rule creation
	return nil
}

// UpdateRoutingRule updates an existing routing rule
func (r *Router) UpdateRoutingRule(ruleID string, rule *RoutingRule) error {
	// TODO: Implement routing rule update
	return nil
}

// DeleteRoutingRule deletes a routing rule
func (r *Router) DeleteRoutingRule(ruleID string) error {
	// TODO: Implement routing rule deletion
	return nil
}

// RoutingRule represents a request routing rule
type RoutingRule struct {
	ID             string                 `json:"id"`
	OrganizationID string                 `json:"organization_id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Priority       int                    `json:"priority"`
	Conditions     map[string]interface{} `json:"conditions"`
	Actions        map[string]interface{} `json:"actions"`
	IsActive       bool                   `json:"is_active"`
}

// evaluateRule evaluates if a routing rule applies to a request
func (r *Router) evaluateRule(rule *RoutingRule, req *http.Request) bool {
	// TODO: Implement rule evaluation logic
	// Check conditions against request properties
	// Support pattern matching, header checks, etc.
	return true
}

// applyRuleActions applies routing rule actions
func (r *Router) applyRuleActions(rule *RoutingRule, servers []*types.MCPServer) []*types.MCPServer {
	// TODO: Implement rule action application
	// Actions could include:
	// - Filter by server tags/metadata
	// - Prefer specific servers
	// - Exclude servers
	// - Weight adjustments
	return servers
}
