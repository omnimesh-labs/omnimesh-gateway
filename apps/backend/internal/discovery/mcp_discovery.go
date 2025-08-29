package discovery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"mcp-gateway/apps/backend/internal/types"
)

// MCPDiscoveryService handles external MCP package discovery
type MCPDiscoveryService struct {
	httpClient *http.Client
	baseURL    string
}

// NewMCPDiscoveryService creates a new MCP discovery service
func NewMCPDiscoveryService(baseURL string) *MCPDiscoveryService {
	return &MCPDiscoveryService{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchPackages searches for MCP packages using the external discovery service
func (s *MCPDiscoveryService) SearchPackages(req *types.MCPDiscoveryRequest) (*types.MCPDiscoveryResponse, error) {
	// Check if base URL is configured
	if s.baseURL == "" {
		return &types.MCPDiscoveryResponse{
			Results:  make(map[string]types.MCPPackage),
			Total:    0,
			Offset:   req.Offset,
			PageSize: req.PageSize,
			HasMore:  false,
		}, nil
	}

	// Build URL with query parameters
	searchURL, err := url.Parse(s.baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	params := url.Values{}
	params.Add("query", req.Query)
	if req.Offset > 0 {
		params.Add("offset", fmt.Sprintf("%d", req.Offset))
	}
	if req.PageSize > 0 {
		params.Add("pageSize", fmt.Sprintf("%d", req.PageSize))
	}
	searchURL.RawQuery = params.Encode()

	// Make HTTP request
	resp, err := s.httpClient.Get(searchURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Parse response
	var discoveryResp types.MCPDiscoveryResponse
	if err := json.NewDecoder(resp.Body).Decode(&discoveryResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &discoveryResp, nil
}

// ListAllPackages returns all packages without a specific query
func (s *MCPDiscoveryService) ListAllPackages(offset, pageSize int) (*types.MCPDiscoveryResponse, error) {
	req := &types.MCPDiscoveryRequest{
		Query:    "", // Empty query to get all packages
		Offset:   offset,
		PageSize: pageSize,
	}

	return s.SearchPackages(req)
}
