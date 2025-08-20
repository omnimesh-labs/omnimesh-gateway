package types

// MCPPackage represents a single MCP package in the discovery service
type MCPPackage struct {
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	GitHubURL            string   `json:"githubUrl"`
	GitHubStars          int      `json:"github_stars"`
	PackageRegistry      string   `json:"package_registry"`
	PackageName          string   `json:"package_name"`
	PackageDownloadCount int      `json:"package_download_count"`
	Command              string   `json:"command"`
	Args                 []string `json:"args"`
	Envs                 []string `json:"envs"`
}

// MCPDiscoveryResponse represents the response from the MCP discovery service
type MCPDiscoveryResponse struct {
	Results  map[string]MCPPackage `json:"results"`
	Total    int                   `json:"total"`
	Offset   int                   `json:"offset"`
	PageSize int                   `json:"pageSize"`
	HasMore  bool                  `json:"hasMore"`
}

// MCPDiscoveryRequest represents a search request for MCP packages
type MCPDiscoveryRequest struct {
	Query    string `json:"query" form:"query"` // Optional - empty query returns all packages
	Offset   int    `json:"offset" form:"offset"`
	PageSize int    `json:"pageSize" form:"pageSize"`
}

// MCPDiscoveryListResponse represents our API response format
type MCPDiscoveryListResponse struct {
	Success bool                 `json:"success"`
	Data    MCPDiscoveryResponse `json:"data"`
	Message string               `json:"message,omitempty"`
}
