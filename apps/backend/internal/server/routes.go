package server

import (
	"context"
	"mcp-gateway/apps/backend/internal/auth"
	"mcp-gateway/apps/backend/internal/database/models"
	"mcp-gateway/apps/backend/internal/discovery"
	"mcp-gateway/apps/backend/internal/logging"
	"mcp-gateway/apps/backend/internal/middleware"
	"mcp-gateway/apps/backend/internal/server/handlers"
	"mcp-gateway/apps/backend/internal/transport"
	"mcp-gateway/apps/backend/internal/types"
	"mcp-gateway/apps/backend/internal/virtual"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.New()
	r.Use(gin.Recovery())

	// Add security headers middleware early in the chain
	var securityConfig *middleware.SecurityConfig
	if s.cfg.Logging.Environment == "development" {
		securityConfig = middleware.DevelopmentSecurityConfig()
	} else {
		securityConfig = middleware.DefaultSecurityConfig()
	}
	r.Use(middleware.SecurityHeadersWithConfig(securityConfig))

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

	// Initialize transport manager
	transportConfig := s.cfg.Transport.ToTransportConfig()
	transportManager := transport.NewManager(transportConfig)
	if err := transportManager.Initialize(context.TODO()); err != nil {
		// Log error but continue - transport layer is optional
	}

	// Initialize virtual server service
	virtualService := virtual.NewService(s.db.GetDB())

	// Initialize handlers
	mcpDiscoveryHandler := handlers.NewMCPDiscoveryHandler(mcpDiscoveryService)
	gatewayHandler := handlers.NewGatewayHandler(discoveryService)
	virtualAdminHandler := handlers.NewVirtualAdminHandler(virtualService)
	virtualMCPHandler := handlers.NewVirtualMCPHandler(virtualService)

	// Initialize admin handler (for logging and system management)
	adminHandler := handlers.NewAdminHandler(nil, s.logging.(*logging.Service), nil, nil)

	// Initialize policy handler
	policyHandler := handlers.NewPolicyHandler(s.db.GetDB())

	// Initialize resource, prompt, and tool models and handlers
	resourceModel := models.NewMCPResourceModel(s.db.GetDB())
	promptModel := models.NewMCPPromptModel(s.db.GetDB())
	toolModel := models.NewMCPToolModel(s.db.GetDB())
	resourceHandler := handlers.NewResourceHandler(resourceModel)
	promptHandler := handlers.NewPromptHandler(promptModel)
	toolHandler := handlers.NewToolHandler(toolModel)

	// Initialize authentication service
	authConfig := &auth.Config{
		JWTSecret:          s.cfg.Auth.JWTSecret,
		AccessTokenExpiry:  s.cfg.Auth.AccessTokenExpiry,
		RefreshTokenExpiry: s.cfg.Auth.RefreshTokenExpiry,
		BCryptCost:         s.cfg.Auth.BCryptCost,
	}

	// Set defaults if not configured
	if authConfig.JWTSecret == "" {
		authConfig.JWTSecret = "development-secret-change-in-production" // TODO: Get from env
	}
	if authConfig.AccessTokenExpiry == 0 {
		authConfig.AccessTokenExpiry = 15 * time.Minute
	}
	if authConfig.RefreshTokenExpiry == 0 {
		authConfig.RefreshTokenExpiry = 24 * time.Hour
	}
	if authConfig.BCryptCost == 0 {
		authConfig.BCryptCost = 12
	}

	authService := auth.NewService(s.db.GetDB(), authConfig)
	authHandler := handlers.NewAuthHandler(authService)

	// Initialize auth middleware
	authMiddleware := auth.NewMiddleware(authService.GetJWTManager(), authService)

	// Initialize transport handlers
	rpcHandler := handlers.NewRPCHandler(transportManager)
	sseHandler := handlers.NewSSEHandler(transportManager)
	wsHandler := handlers.NewWebSocketHandler(transportManager)
	mcpHandler := handlers.NewMCPHandler(transportManager)
	stdioHandler := handlers.NewSTDIOHandler(transportManager)

	// Virtual MCP JSON-RPC endpoint
	r.POST("/mcp/rpc", virtualMCPHandler.HandleMCPRPC)

	// API routes
	api := r.Group("/api")
	{
		// Authentication routes
		auth := api.Group("/auth")
		{
			// Public routes (no auth required)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)

			// Protected routes (auth required)
			protected := auth.Group("/")
			protected.Use(authMiddleware.RequireAuth())
			{
				protected.POST("/logout", authHandler.Logout)
				protected.GET("/profile", authHandler.GetProfile)
				protected.PUT("/profile", authHandler.UpdateProfile)
				protected.POST("/api-keys", authHandler.CreateAPIKey)
			}
		}

		// MCP Discovery routes (require authentication and read permission)
		mcp := api.Group("/mcp")
		mcp.Use(authMiddleware.RequireAuth())
		mcp.Use(authMiddleware.RequirePermission(types.PermissionRead))
		{
			// Search for MCP packages
			mcp.GET("/search", mcpDiscoveryHandler.SearchPackages)

			// List all MCP packages
			mcp.GET("/packages", mcpDiscoveryHandler.ListPackages)

			// Get specific package details
			mcp.GET("/packages/:packageName", mcpDiscoveryHandler.GetPackageDetails)

			// Public tools available to all organizations
			mcp.GET("/tools/public", toolHandler.ListPublicTools)
		}

		// Gateway management routes (protected)
		gateway := api.Group("/gateway")
		gateway.Use(authMiddleware.RequireAuth())
		gateway.Use(authMiddleware.RequireOrganizationAccess())
		{
			// Server management - requires appropriate permissions
			gateway.GET("/servers",
				authMiddleware.RequireResourceAccess("server", "read"),
				gatewayHandler.ListServers)
			gateway.POST("/servers",
				authMiddleware.RequireResourceAccess("server", "write"),
				loggingMiddleware.AuditLogger("register", "server"),
				gatewayHandler.RegisterServer)
			gateway.GET("/servers/:id",
				authMiddleware.RequireResourceAccess("server", "read"),
				gatewayHandler.GetServer)
			gateway.PUT("/servers/:id",
				authMiddleware.RequireResourceAccess("server", "write"),
				loggingMiddleware.AuditLogger("update", "server"),
				gatewayHandler.UpdateServer)
			gateway.DELETE("/servers/:id",
				authMiddleware.RequireResourceAccess("server", "delete"),
				loggingMiddleware.AuditLogger("unregister", "server"),
				gatewayHandler.UnregisterServer)
			gateway.GET("/servers/:id/stats",
				authMiddleware.RequireResourceAccess("server", "read"),
				gatewayHandler.GetServerStats)

			// MCP session management - requires session permissions
			gateway.POST("/sessions",
				authMiddleware.RequireResourceAccess("session", "write"),
				gatewayHandler.CreateMCPSession)
			gateway.GET("/sessions",
				authMiddleware.RequireResourceAccess("session", "read"),
				gatewayHandler.ListMCPSessions)
			gateway.DELETE("/sessions/:session_id",
				authMiddleware.RequireResourceAccess("session", "delete"),
				gatewayHandler.CloseMCPSession)

			// WebSocket endpoint for MCP communication
			gateway.GET("/ws",
				authMiddleware.RequireResourceAccess("session", "write"),
				gatewayHandler.HandleMCPWebSocket)

			// Resource management - requires resource permissions
			gateway.GET("/resources",
				authMiddleware.RequireResourceAccess("resource", "read"),
				resourceHandler.ListResources)
			gateway.POST("/resources",
				authMiddleware.RequireResourceAccess("resource", "write"),
				loggingMiddleware.AuditLogger("create", "resource"),
				resourceHandler.CreateResource)
			gateway.GET("/resources/:id",
				authMiddleware.RequireResourceAccess("resource", "read"),
				resourceHandler.GetResource)
			gateway.PUT("/resources/:id",
				authMiddleware.RequireResourceAccess("resource", "write"),
				loggingMiddleware.AuditLogger("update", "resource"),
				resourceHandler.UpdateResource)
			gateway.DELETE("/resources/:id",
				authMiddleware.RequireResourceAccess("resource", "delete"),
				loggingMiddleware.AuditLogger("delete", "resource"),
				resourceHandler.DeleteResource)

			// Prompt management - requires prompt permissions
			gateway.GET("/prompts",
				authMiddleware.RequireResourceAccess("prompt", "read"),
				promptHandler.ListPrompts)
			gateway.POST("/prompts",
				authMiddleware.RequireResourceAccess("prompt", "write"),
				loggingMiddleware.AuditLogger("create", "prompt"),
				promptHandler.CreatePrompt)
			gateway.GET("/prompts/:id",
				authMiddleware.RequireResourceAccess("prompt", "read"),
				promptHandler.GetPrompt)
			gateway.PUT("/prompts/:id",
				authMiddleware.RequireResourceAccess("prompt", "write"),
				loggingMiddleware.AuditLogger("update", "prompt"),
				promptHandler.UpdatePrompt)
			gateway.DELETE("/prompts/:id",
				authMiddleware.RequireResourceAccess("prompt", "delete"),
				loggingMiddleware.AuditLogger("delete", "prompt"),
				promptHandler.DeletePrompt)
			gateway.POST("/prompts/:id/use",
				authMiddleware.RequireResourceAccess("prompt", "read"),
				promptHandler.UsePrompt)

			// Tool management - requires tool permissions
			gateway.GET("/tools",
				authMiddleware.RequireResourceAccess("tool", "read"),
				toolHandler.ListTools)
			gateway.POST("/tools",
				authMiddleware.RequireResourceAccess("tool", "write"),
				loggingMiddleware.AuditLogger("create", "tool"),
				toolHandler.CreateTool)
			gateway.GET("/tools/:id",
				authMiddleware.RequireResourceAccess("tool", "read"),
				toolHandler.GetTool)
			gateway.PUT("/tools/:id",
				authMiddleware.RequireResourceAccess("tool", "write"),
				loggingMiddleware.AuditLogger("update", "tool"),
				toolHandler.UpdateTool)
			gateway.DELETE("/tools/:id",
				authMiddleware.RequireResourceAccess("tool", "delete"),
				loggingMiddleware.AuditLogger("delete", "tool"),
				toolHandler.DeleteTool)
			gateway.POST("/tools/:id/execute",
				authMiddleware.RequireResourceAccess("tool", "execute"),
				toolHandler.ExecuteTool)

			// Tool function lookup
			gateway.GET("/tools/function/:function_name",
				authMiddleware.RequireResourceAccess("tool", "read"),
				toolHandler.GetToolByFunction)

		}

		// Admin routes for virtual servers and system management (protected)
		admin := api.Group("/admin")
		admin.Use(authMiddleware.RequireAuth())
		admin.Use(authMiddleware.RequireOrganizationAccess())
		{
			// Logging and audit routes - require admin access
			admin.GET("/logs",
				authMiddleware.RequireAdmin(),
				authMiddleware.RequirePermission(types.PermissionLogsRead),
				adminHandler.GetLogs)
			admin.GET("/audit",
				authMiddleware.RequireAdmin(),
				authMiddleware.RequirePermission(types.PermissionAuditRead),
				adminHandler.GetAuditLogs)
			admin.GET("/stats",
				authMiddleware.RequireAdmin(),
				authMiddleware.RequirePermission(types.PermissionMetricsRead),
				adminHandler.GetStats)
			admin.GET("/metrics",
				authMiddleware.RequireAdmin(),
				authMiddleware.RequirePermission(types.PermissionMetricsRead),
				adminHandler.GetMetrics)

			// Virtual server management - role-based access
			virtual := admin.Group("/virtual-servers")
			{
				virtual.POST("",
					authMiddleware.RequireResourceAccess("virtual_server", "write"),
					loggingMiddleware.AuditLogger("create", "virtual-server"),
					virtualAdminHandler.CreateVirtualServer)
				virtual.GET("",
					authMiddleware.RequireResourceAccess("virtual_server", "read"),
					virtualAdminHandler.ListVirtualServers)
				virtual.GET("/:id",
					authMiddleware.RequireResourceAccess("virtual_server", "read"),
					virtualAdminHandler.GetVirtualServer)
				virtual.PUT("/:id",
					authMiddleware.RequireResourceAccess("virtual_server", "write"),
					loggingMiddleware.AuditLogger("update", "virtual-server"),
					virtualAdminHandler.UpdateVirtualServer)
				virtual.DELETE("/:id",
					authMiddleware.RequireResourceAccess("virtual_server", "delete"),
					loggingMiddleware.AuditLogger("delete", "virtual-server"),
					virtualAdminHandler.DeleteVirtualServer)
				virtual.GET("/:id/tools",
					authMiddleware.RequireResourceAccess("virtual_server", "read"),
					virtualAdminHandler.GetVirtualServerTools)
				virtual.POST("/:id/tools/:tool/test",
					authMiddleware.RequireResourceAccess("virtual_server", "write"),
					virtualAdminHandler.TestVirtualServerTool)
			}

			// Policy management - requires admin access and policy permissions
			policies := admin.Group("/policies")
			{
				policies.GET("",
					authMiddleware.RequireAdmin(),
					authMiddleware.RequirePermission(types.PermissionRead),
					policyHandler.ListPolicies)
				policies.POST("",
					authMiddleware.RequireAdmin(),
					authMiddleware.RequirePermission(types.PermissionWrite),
					loggingMiddleware.AuditLogger("create", "policy"),
					policyHandler.CreatePolicy)
				policies.GET("/:id",
					authMiddleware.RequireAdmin(),
					authMiddleware.RequirePermission(types.PermissionRead),
					policyHandler.GetPolicy)
				policies.PUT("/:id",
					authMiddleware.RequireAdmin(),
					authMiddleware.RequirePermission(types.PermissionWrite),
					loggingMiddleware.AuditLogger("update", "policy"),
					policyHandler.UpdatePolicy)
				policies.DELETE("/:id",
					authMiddleware.RequireAdmin(),
					authMiddleware.RequirePermission(types.PermissionDelete),
					loggingMiddleware.AuditLogger("delete", "policy"),
					policyHandler.DeletePolicy)
			}

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
