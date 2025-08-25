package types

// MCPPackage represents a single MCP package in the discovery service
type MCPPackage struct {
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	GitHubURL            string   `json:"githubUrl"`
	PackageRegistry      string   `json:"package_registry"`
	PackageName          string   `json:"package_name"`
	Command              string   `json:"command"`
	Args                 []string `json:"args"`
	Envs                 []string `json:"envs"`
	GitHubStars          int      `json:"github_stars"`
	PackageDownloadCount int      `json:"package_download_count"`
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
	Message string               `json:"message,omitempty"`
	Data    MCPDiscoveryResponse `json:"data"`
	Success bool                 `json:"success"`
}
