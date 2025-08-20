package server

import (
	"mcp-gateway/apps/backend/internal/discovery"
	"mcp-gateway/apps/backend/internal/server/handlers"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Add your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // Enable cookies/auth
	}))

	r.GET("/", s.HelloWorldHandler)

	r.GET("/health", s.healthHandler)

	// Initialize MCP Discovery Service and Handler
	mcpDiscoveryService := discovery.NewMCPDiscoveryService(s.cfg.MCPDiscovery.BaseURL)
	mcpDiscoveryHandler := handlers.NewMCPDiscoveryHandler(mcpDiscoveryService)

	// MCP Discovery API routes
	api := r.Group("/api")
	{
		mcp := api.Group("/mcp")
		{
			// Search for MCP packages
			mcp.GET("/search", mcpDiscoveryHandler.SearchPackages)

			// List all MCP packages
			mcp.GET("/packages", mcpDiscoveryHandler.ListPackages)

			// Get specific package details
			mcp.GET("/packages/:packageName", mcpDiscoveryHandler.GetPackageDetails)
		}
	}

	return r
}

func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
