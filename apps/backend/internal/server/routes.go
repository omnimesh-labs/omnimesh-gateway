package server

import (
	"mcp-gateway/apps/backend/internal/discovery"
	"mcp-gateway/apps/backend/internal/gateway"
	"mcp-gateway/apps/backend/internal/server/handlers"
	"mcp-gateway/apps/backend/internal/types"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"}, // Add frontend URLs
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		AllowCredentials: true, // Enable cookies/auth
	}))

	r.GET("/", s.HelloWorldHandler)

	r.GET("/health", s.healthHandler)

	// Initialize services
	mcpDiscoveryService := discovery.NewMCPDiscoveryService(s.cfg.MCPDiscovery.BaseURL)

	// Initialize discovery service
	discoveryConfig := &discovery.Config{
		Enabled:          true,
		HealthInterval:   30 * time.Second,
		FailureThreshold: 3,
		RecoveryTimeout:  5 * time.Minute,
		SingleTenant:     true, // Enable single-tenant mode
	}
	discoveryService := discovery.NewService(s.db.GetDB(), discoveryConfig)

	// Initialize proxy services
	proxyConfig := &types.MCPProxyConfig{
		MaxConcurrentSessions: 100,
		SessionTimeout:        30 * time.Minute,
		ProcessTimeout:        5 * time.Minute,
		BufferSize:            4096,
		EnableLogging:         true,
		LogLevel:              "info",
	}
	mcpProxy := gateway.NewMCPProxy(proxyConfig)

	// You'll need to implement the legacy Proxy as well
	legacyProxy := gateway.NewProxy(nil, nil) // Placeholder - implement as needed

	// Initialize handlers
	mcpDiscoveryHandler := handlers.NewMCPDiscoveryHandler(mcpDiscoveryService)
	gatewayHandler := handlers.NewGatewayHandler(discoveryService, legacyProxy, mcpProxy)

	// API routes
	api := r.Group("/api")
	{
		// MCP Discovery routes
		mcp := api.Group("/mcp")
		{
			// Search for MCP packages
			mcp.GET("/search", mcpDiscoveryHandler.SearchPackages)

			// List all MCP packages
			mcp.GET("/packages", mcpDiscoveryHandler.ListPackages)

			// Get specific package details
			mcp.GET("/packages/:packageName", mcpDiscoveryHandler.GetPackageDetails)
		}

		// Gateway management routes
		gateway := api.Group("/gateway")
		{
			// Server management
			gateway.GET("/servers", gatewayHandler.ListServers)
			gateway.POST("/servers", gatewayHandler.RegisterServer)
			gateway.GET("/servers/:id", gatewayHandler.GetServer)
			gateway.PUT("/servers/:id", gatewayHandler.UpdateServer)
			gateway.DELETE("/servers/:id", gatewayHandler.UnregisterServer)
			gateway.GET("/servers/:id/stats", gatewayHandler.GetServerStats)

			// MCP session management
			gateway.POST("/sessions", gatewayHandler.CreateMCPSession)
			gateway.GET("/sessions", gatewayHandler.ListMCPSessions)
			gateway.DELETE("/sessions/:session_id", gatewayHandler.CloseMCPSession)

			// WebSocket endpoint for MCP communication
			gateway.GET("/ws", gatewayHandler.HandleMCPWebSocket)

			// Legacy proxy endpoint
			gateway.Any("/proxy/*path", gatewayHandler.ProxyRequest)
		}
	}

	return r
}

func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "all quiet on the western front"

	c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
