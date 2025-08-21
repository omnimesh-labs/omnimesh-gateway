package server

import (
	"context"
	"mcp-gateway/apps/backend/internal/discovery"
	"mcp-gateway/apps/backend/internal/gateway"
	"mcp-gateway/apps/backend/internal/logging"
	"mcp-gateway/apps/backend/internal/middleware"
	"mcp-gateway/apps/backend/internal/server/handlers"
	"mcp-gateway/apps/backend/internal/transport"
	"mcp-gateway/apps/backend/internal/types"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.New()
	r.Use(gin.Recovery())

	// Initialize logging middleware
	loggingMiddleware := logging.NewMiddleware(s.logging.(*logging.Service))
	r.Use(loggingMiddleware.RequestLogger())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		AllowCredentials: true,
	}))

	r.GET("/", s.HelloWorldHandler)

	r.GET("/health", s.healthHandler)

	// Initialize middleware
	pathRewriteMiddleware := middleware.NewPathRewriteMiddleware()

	// Apply transport middleware for all transport routes
	transportGroup := r.Group("/")
	transportGroup.Use(pathRewriteMiddleware.Handler())
	transportGroup.Use(middleware.ServerContextMiddleware())
	transportGroup.Use(middleware.TransportTypeMiddleware())
	transportGroup.Use(middleware.SessionIDMiddleware())

	// Initialize services
	mcpDiscoveryService := discovery.NewMCPDiscoveryService(s.cfg.MCPDiscovery.BaseURL)

	// Initialize discovery service
	discoveryConfig := &discovery.Config{
		Enabled:          true,
		HealthInterval:   30 * time.Second,
		FailureThreshold: 3,
		RecoveryTimeout:  5 * time.Minute,
		SingleTenant:     true,
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

	legacyProxy := gateway.NewProxy(nil, nil)

	// Initialize transport manager
	transportConfig := s.cfg.Transport.ToTransportConfig()
	transportManager := transport.NewManager(transportConfig)
	if err := transportManager.Initialize(context.TODO()); err != nil {
		// Log error but continue - transport layer is optional
	}

	// Initialize handlers
	mcpDiscoveryHandler := handlers.NewMCPDiscoveryHandler(mcpDiscoveryService)
	gatewayHandler := handlers.NewGatewayHandler(discoveryService, legacyProxy, mcpProxy)

	// Initialize transport handlers
	rpcHandler := handlers.NewRPCHandler(transportManager)
	sseHandler := handlers.NewSSEHandler(transportManager)
	wsHandler := handlers.NewWebSocketHandler(transportManager)
	mcpHandler := handlers.NewMCPHandler(transportManager)
	stdioHandler := handlers.NewSTDIOHandler(transportManager)

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
			gateway.POST("/servers", loggingMiddleware.AuditLogger("register", "server"), gatewayHandler.RegisterServer)
			gateway.GET("/servers/:id", gatewayHandler.GetServer)
			gateway.PUT("/servers/:id", loggingMiddleware.AuditLogger("update", "server"), gatewayHandler.UpdateServer)
			gateway.DELETE("/servers/:id", loggingMiddleware.AuditLogger("unregister", "server"), gatewayHandler.UnregisterServer)
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

	// Transport endpoints (with middleware applied)
	// JSON-RPC over HTTP
	transportGroup.POST("/rpc", rpcHandler.HandleJSONRPC)
	transportGroup.POST("/rpc/batch", rpcHandler.HandleBatchRPC)
	transportGroup.GET("/rpc/introspection", rpcHandler.HandleRPCIntrospection)
	transportGroup.GET("/rpc/health", rpcHandler.HandleRPCHealth)

	// Server-Sent Events
	transportGroup.GET("/sse", sseHandler.HandleSSE)
	transportGroup.POST("/sse/events", sseHandler.HandleSSEEvents)
	transportGroup.POST("/sse/broadcast", sseHandler.HandleSSEBroadcast)
	transportGroup.GET("/sse/status", sseHandler.HandleSSEStatus)
	transportGroup.GET("/sse/replay/:session_id", sseHandler.HandleSSEReplay)
	transportGroup.GET("/sse/health", sseHandler.HandleSSEHealth)

	// WebSocket
	transportGroup.GET("/ws", wsHandler.HandleWebSocket)
	transportGroup.POST("/ws/send", wsHandler.HandleWebSocketSend)
	transportGroup.POST("/ws/broadcast", wsHandler.HandleWebSocketBroadcast)
	transportGroup.GET("/ws/status", wsHandler.HandleWebSocketStatus)
	transportGroup.POST("/ws/ping", wsHandler.HandleWebSocketPing)
	transportGroup.DELETE("/ws/close", wsHandler.HandleWebSocketClose)
	transportGroup.GET("/ws/health", wsHandler.HandleWebSocketHealth)
	transportGroup.GET("/ws/metrics", wsHandler.HandleWebSocketMetrics)

	// Streamable HTTP (MCP Protocol)
	transportGroup.Any("/mcp", mcpHandler.HandleStreamableHTTP)
	transportGroup.GET("/mcp/capabilities", mcpHandler.HandleMCPCapabilities)
	transportGroup.GET("/mcp/status", mcpHandler.HandleMCPStatus)
	transportGroup.GET("/mcp/health", mcpHandler.HandleMCPHealth)

	// STDIO Bridge
	transportGroup.POST("/stdio/execute", stdioHandler.HandleSTDIOExecute)
	transportGroup.Any("/stdio/process", stdioHandler.HandleSTDIOProcess)
	transportGroup.POST("/stdio/send", stdioHandler.HandleSTDIOSend)
	transportGroup.GET("/stdio/health", stdioHandler.HandleSTDIOHealth)

	// Server-specific transport endpoints (will be rewritten by middleware)
	// These routes are handled by the path rewriting middleware
	transportGroup.POST("/servers/:server_id/rpc", rpcHandler.HandleJSONRPC)
	transportGroup.GET("/servers/:server_id/sse", sseHandler.HandleServerSSE)
	transportGroup.GET("/servers/:server_id/ws", wsHandler.HandleServerWebSocket)
	transportGroup.Any("/servers/:server_id/mcp", mcpHandler.HandleServerStreamableHTTP)

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
