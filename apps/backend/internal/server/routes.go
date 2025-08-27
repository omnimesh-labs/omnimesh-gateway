package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"mcp-gateway/apps/backend/internal/a2a"
	"mcp-gateway/apps/backend/internal/auth"
	"mcp-gateway/apps/backend/internal/config"
	"mcp-gateway/apps/backend/internal/database/models"
	"mcp-gateway/apps/backend/internal/discovery"
	"mcp-gateway/apps/backend/internal/inspector"
	"mcp-gateway/apps/backend/internal/logging"
	"mcp-gateway/apps/backend/internal/middleware"
	"mcp-gateway/apps/backend/internal/plugins"
	"mcp-gateway/apps/backend/internal/server/handlers"
	"mcp-gateway/apps/backend/internal/services"
	"mcp-gateway/apps/backend/internal/transport"
	"mcp-gateway/apps/backend/internal/types"
	"mcp-gateway/apps/backend/internal/virtual"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.New()

	// Initialize logging middleware
	loggingMiddleware := logging.NewMiddleware(s.logging.(*logging.Service))

	// Initialize plugin service for content filtering
	pluginService := plugins.NewPluginService(s.db.GetDB())
	if err := pluginService.Initialize(context.TODO()); err != nil {
		// Log error but continue - content filtering is optional for basic functionality
	}
	contentFilterMiddleware := plugins.NewFilterMiddleware(pluginService)

	// Configure security headers based on environment
	var securityConfig *middleware.SecurityConfig
	if s.cfg.Logging.Environment == "development" {
		securityConfig = middleware.DevelopmentSecurityConfig()
	} else {
		securityConfig = middleware.DefaultSecurityConfig()
	}

	// Apply CORS middleware globally to all routes
	var corsConfig cors.Config
	if s.cfg.Logging.Environment == "development" {
		// Development CORS - allow common development origins
		corsConfig = cors.Config{
			AllowOrigins: []string{
				"http://localhost:3000",
				"http://localhost:3001",
				"http://localhost:5173",
				"http://localhost:8080",
				"http://127.0.0.1:3000",
				"http://127.0.0.1:8080",
				"http://backend:8080",
				"http://frontend:3000",
			},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
			AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-Requested-With", "X-API-Key"},
			AllowCredentials: true,
			ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		}
	} else {
		// Strict CORS for production
		corsConfig = cors.Config{
			AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173", "http://localhost:8080", "http://backend:8080"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
			AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
			AllowCredentials: true,
		}
	}
	r.Use(cors.New(corsConfig))

	// Apply default middleware chain to root router
	defaultChain := middleware.DefaultChainWithConfig(securityConfig)
	defaultChain.Use(loggingMiddleware.RequestLogger())
	// Only apply content filtering if not in test environment
	if os.Getenv("SKIP_CONTENT_FILTERING") != "true" {
		defaultChain.Use(contentFilterMiddleware.Handler())
	}
	rootGroup := &r.RouterGroup
	defaultChain.Apply(rootGroup)

	r.GET("/", s.HelloWorldHandler)

	r.GET("/health", s.healthHandler)

	// Initialize middleware
	pathRewriteMiddleware := middleware.NewPathRewriteMiddleware()

	// Create transport middleware chain
	transportChain := middleware.NewChain().
		Use(pathRewriteMiddleware.Handler()).
		Use(middleware.ServerContextMiddleware()).
		Use(middleware.TransportTypeMiddleware()).
		Use(middleware.SessionIDMiddleware())

	// Apply transport middleware for all transport routes
	transportGroup := r.Group("/")
	transportChain.Apply(transportGroup)

	// Initialize services
	mcpDiscoveryService := discovery.NewMCPDiscoveryService(s.cfg.MCPDiscovery.BaseURL)

	// Initialize endpoint service
	baseURL := "http://localhost:8080" // TODO: Get from config
	endpointService := services.NewEndpointService(s.db.GetDB(), baseURL)

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

	// Initialize A2A services
	a2aService := a2a.NewService(s.db.GetDB())
	a2aClient := a2a.NewClient(30*time.Second, 3)
	a2aAdapter := a2a.NewAdapter(a2aService, a2aClient)

	// Initialize namespace service
	namespaceService := services.NewNamespaceService(s.db.GetDB(), endpointService)

	// Initialize inspector service
	inspectorService := inspector.NewService(transportManager)

	// Initialize handlers
	mcpDiscoveryHandler := handlers.NewMCPDiscoveryHandler(mcpDiscoveryService)
	gatewayHandler := handlers.NewGatewayHandler(discoveryService)
	virtualAdminHandler := handlers.NewVirtualAdminHandler(virtualService)
	virtualMCPHandler := handlers.NewVirtualMCPHandler(virtualService)
	a2aHandler := handlers.NewA2AHandler(a2aService, a2aClient, a2aAdapter)
	namespaceHandler := handlers.NewNamespaceHandler(namespaceService)
	inspectorHandler := handlers.NewInspectorHandler(inspectorService)

	// Initialize config service
	configService := config.NewService(s.db.GetDB())

	// Initialize admin handler (for logging and system management)
	adminHandler := handlers.NewAdminHandler(nil, s.logging.(*logging.Service), configService)

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
		// Try to get JWT secret from environment variable
		if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
			authConfig.JWTSecret = jwtSecret
		} else {
			log.Fatal("JWT_SECRET environment variable is required. Please set a secure secret.")
		}
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
	rpcHandler := handlers.NewRPCHandler(transportManager, discoveryService)
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
			authenticatedChain := middleware.AuthenticatedChain().Use(authMiddleware.RequireAuth())
			protected := auth.Group("/")
			authenticatedChain.Apply(protected)
			{
				protected.POST("/logout", authHandler.Logout)
				protected.GET("/profile", authHandler.GetProfile)
				protected.PUT("/profile", authHandler.UpdateProfile)
				protected.POST("/api-keys", authHandler.CreateAPIKey)
				protected.GET("/api-keys", authHandler.ListAPIKeys)
				protected.DELETE("/api-keys/:id", authHandler.DeleteAPIKey)
			}
		}

		// MCP Discovery routes (require authentication and read permission)
		mcpChain := middleware.AuthenticatedChain().
			Use(authMiddleware.RequireAuth()).
			Use(authMiddleware.RequirePermission(types.PermissionRead))
		mcp := api.Group("/mcp")
		mcpChain.Apply(mcp)
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
		gatewayChain := middleware.AuthenticatedChain().
			Use(authMiddleware.RequireAuth()).
			Use(authMiddleware.RequireOrganizationAccess())
		gateway := api.Group("/gateway")
		gatewayChain.Apply(gateway)
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

		// Namespace management routes (protected)
		namespaceChain := middleware.AuthenticatedChain().
			Use(authMiddleware.RequireAuth()).
			Use(authMiddleware.RequireOrganizationAccess())
		namespaces := api.Group("/namespaces")
		namespaceChain.Apply(namespaces)
		{
			// Namespace CRUD operations
			namespaces.GET("",
				authMiddleware.RequireResourceAccess("namespace", "read"),
				namespaceHandler.ListNamespaces)
			namespaces.POST("",
				authMiddleware.RequireResourceAccess("namespace", "write"),
				loggingMiddleware.AuditLogger("create", "namespace"),
				namespaceHandler.CreateNamespace)
			namespaces.GET("/:id",
				authMiddleware.RequireResourceAccess("namespace", "read"),
				namespaceHandler.GetNamespace)
			namespaces.PUT("/:id",
				authMiddleware.RequireResourceAccess("namespace", "write"),
				loggingMiddleware.AuditLogger("update", "namespace"),
				namespaceHandler.UpdateNamespace)
			namespaces.DELETE("/:id",
				authMiddleware.RequireResourceAccess("namespace", "delete"),
				loggingMiddleware.AuditLogger("delete", "namespace"),
				namespaceHandler.DeleteNamespace)

			// Server mappings
			namespaces.POST("/:id/servers",
				authMiddleware.RequireResourceAccess("namespace", "write"),
				loggingMiddleware.AuditLogger("add-server", "namespace"),
				namespaceHandler.AddServerToNamespace)
			namespaces.DELETE("/:id/servers/:server_id",
				authMiddleware.RequireResourceAccess("namespace", "write"),
				loggingMiddleware.AuditLogger("remove-server", "namespace"),
				namespaceHandler.RemoveServerFromNamespace)
			namespaces.PUT("/:id/servers/:server_id/status",
				authMiddleware.RequireResourceAccess("namespace", "write"),
				loggingMiddleware.AuditLogger("update-server-status", "namespace"),
				namespaceHandler.UpdateServerStatus)

			// Tool management
			namespaces.GET("/:id/tools",
				authMiddleware.RequireResourceAccess("namespace", "read"),
				namespaceHandler.GetNamespaceTools)
			namespaces.PUT("/:id/tools/:tool_id/status",
				authMiddleware.RequireResourceAccess("namespace", "write"),
				loggingMiddleware.AuditLogger("update-tool-status", "namespace"),
				namespaceHandler.UpdateToolStatus)

			// MCP operations
			namespaces.POST("/:id/execute",
				authMiddleware.RequireResourceAccess("namespace", "execute"),
				loggingMiddleware.AuditLogger("execute-tool", "namespace"),
				namespaceHandler.ExecuteNamespaceTool)
		}

		// Inspector routes (protected)
		inspectorChain := middleware.AuthenticatedChain().
			Use(authMiddleware.RequireAuth()).
			Use(authMiddleware.RequireOrganizationAccess())
		inspectorGroup := api.Group("/inspector")
		inspectorChain.Apply(inspectorGroup)
		{
			// Session management
			inspectorGroup.POST("/sessions",
				authMiddleware.RequireResourceAccess("inspector", "write"),
				loggingMiddleware.AuditLogger("create", "inspector-session"),
				inspectorHandler.CreateSession)
			inspectorGroup.GET("/sessions/:id",
				authMiddleware.RequireResourceAccess("inspector", "read"),
				inspectorHandler.GetSession)
			inspectorGroup.DELETE("/sessions/:id",
				authMiddleware.RequireResourceAccess("inspector", "delete"),
				loggingMiddleware.AuditLogger("close", "inspector-session"),
				inspectorHandler.CloseSession)

			// Request execution
			inspectorGroup.POST("/sessions/:id/request",
				authMiddleware.RequireResourceAccess("inspector", "execute"),
				inspectorHandler.ExecuteRequest)

			// Event streaming
			inspectorGroup.GET("/sessions/:id/events",
				authMiddleware.RequireResourceAccess("inspector", "read"),
				inspectorHandler.StreamEvents) // SSE endpoint
			inspectorGroup.GET("/sessions/:id/ws",
				authMiddleware.RequireResourceAccess("inspector", "read"),
				inspectorHandler.HandleWebSocket) // WebSocket endpoint

			// Server capabilities
			inspectorGroup.GET("/servers/:id/capabilities",
				authMiddleware.RequireResourceAccess("inspector", "read"),
				inspectorHandler.GetServerCapabilities)
		}

		// A2A (Agent-to-Agent) management routes (protected)
		a2aChain := middleware.AuthenticatedChain().
			Use(authMiddleware.RequireAuth()).
			Use(authMiddleware.RequireOrganizationAccess())
		a2aGroup := api.Group("/a2a")
		a2aChain.Apply(a2aGroup)
		{
			// Agent management - requires appropriate permissions
			a2aGroup.GET("",
				authMiddleware.RequireResourceAccess("a2a_agent", "read"),
				a2aHandler.ListAgents)
			a2aGroup.POST("",
				authMiddleware.RequireResourceAccess("a2a_agent", "write"),
				loggingMiddleware.AuditLogger("create", "a2a-agent"),
				a2aHandler.RegisterAgent)
			a2aGroup.GET("/:id",
				authMiddleware.RequireResourceAccess("a2a_agent", "read"),
				a2aHandler.GetAgent)
			a2aGroup.PUT("/:id",
				authMiddleware.RequireResourceAccess("a2a_agent", "write"),
				loggingMiddleware.AuditLogger("update", "a2a-agent"),
				a2aHandler.UpdateAgent)
			a2aGroup.DELETE("/:id",
				authMiddleware.RequireResourceAccess("a2a_agent", "delete"),
				loggingMiddleware.AuditLogger("delete", "a2a-agent"),
				a2aHandler.DeleteAgent)

			// Agent status management
			a2aGroup.POST("/:id/toggle",
				authMiddleware.RequireResourceAccess("a2a_agent", "write"),
				loggingMiddleware.AuditLogger("toggle", "a2a-agent"),
				a2aHandler.ToggleAgent)

			// Agent health checking
			a2aGroup.GET("/:id/health",
				authMiddleware.RequireResourceAccess("a2a_agent", "read"),
				a2aHandler.HealthCheckAgent)

			// Agent testing
			a2aGroup.POST("/:id/test",
				authMiddleware.RequireResourceAccess("a2a_agent", "execute"),
				loggingMiddleware.AuditLogger("test", "a2a-agent"),
				a2aHandler.TestAgent)

			// Agent tools
			a2aGroup.GET("/:id/tools",
				authMiddleware.RequireResourceAccess("a2a_agent", "read"),
				a2aHandler.GetAgentTools)

			// Agent invocation by ID
			a2aGroup.POST("/:id/invoke",
				authMiddleware.RequireResourceAccess("a2a_agent", "execute"),
				loggingMiddleware.AuditLogger("invoke", "a2a-agent"),
				a2aHandler.InvokeAgent)

			// Agent chat by ID
			a2aGroup.POST("/:id/chat",
				authMiddleware.RequireResourceAccess("a2a_agent", "execute"),
				loggingMiddleware.AuditLogger("chat", "a2a-agent"),
				a2aHandler.ChatWithAgent)

			// A2A statistics
			a2aGroup.GET("/stats",
				authMiddleware.RequireResourceAccess("a2a_agent", "read"),
				a2aHandler.GetAgentStats)
		}

		// Endpoint management routes (protected)
		endpointHandler := handlers.NewEndpointHandler(endpointService)
		endpoints := api.Group("/endpoints")
		endpoints.Use(authMiddleware.RequireAuth()).
			Use(authMiddleware.RequireOrganizationAccess())
		{
			endpoints.GET("",
				authMiddleware.RequireResourceAccess("endpoint", "read"),
				endpointHandler.ListEndpoints)
			endpoints.POST("",
				authMiddleware.RequireResourceAccess("endpoint", "write"),
				loggingMiddleware.AuditLogger("create", "endpoint"),
				endpointHandler.CreateEndpoint)
			endpoints.GET("/:id",
				authMiddleware.RequireResourceAccess("endpoint", "read"),
				endpointHandler.GetEndpoint)
			endpoints.PUT("/:id",
				authMiddleware.RequireResourceAccess("endpoint", "write"),
				loggingMiddleware.AuditLogger("update", "endpoint"),
				endpointHandler.UpdateEndpoint)
			endpoints.DELETE("/:id",
				authMiddleware.RequireResourceAccess("endpoint", "delete"),
				loggingMiddleware.AuditLogger("delete", "endpoint"),
				endpointHandler.DeleteEndpoint)
			endpoints.POST("/:id/regenerate-keys",
				authMiddleware.RequireResourceAccess("endpoint", "write"),
				loggingMiddleware.AuditLogger("regenerate-keys", "endpoint"),
				endpointHandler.RegenerateEndpointKeys)
		}

		// Admin routes for virtual servers and system management (protected)
		adminChain := middleware.AdminChain().
			Use(authMiddleware.RequireAuth()).
			Use(authMiddleware.RequireOrganizationAccess())
		admin := api.Group("/admin")
		adminChain.Apply(admin)
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

			// Configuration management - requires admin access
			config := admin.Group("/config")
			{
				config.POST("/export",
					authMiddleware.RequireAdmin(),
					authMiddleware.RequirePermission(types.PermissionRead),
					loggingMiddleware.AuditLogger("export", "configuration"),
					adminHandler.ExportConfiguration)
				config.POST("/import",
					authMiddleware.RequireAdmin(),
					authMiddleware.RequirePermission(types.PermissionWrite),
					loggingMiddleware.AuditLogger("import", "configuration"),
					adminHandler.ImportConfiguration)
				config.POST("/validate-import",
					authMiddleware.RequireAdmin(),
					authMiddleware.RequirePermission(types.PermissionRead),
					adminHandler.ValidateImport)
				config.GET("/import-history",
					authMiddleware.RequireAdmin(),
					authMiddleware.RequirePermission(types.PermissionRead),
					adminHandler.GetImportHistory)
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

	// Public-facing endpoint routes (requires authentication)
	// These are public in the sense that they're accessible via custom URLs,
	// but still require proper authentication based on endpoint configuration
	// Note: reusing the endpointHandler created above in the protected routes
	publicEndpoints := api.Group("/public")
	publicEndpoints.Use(authMiddleware.RequireAuth())
	{
		// List all available endpoints for the authenticated user's organization
		// Get the handler from the protected routes section above
		endpointHandlerForPublic := handlers.NewEndpointHandler(endpointService)
		publicEndpoints.GET("/endpoints", endpointHandlerForPublic.ListEndpoints)

		// Endpoint-specific routes with custom URL paths
		endpoint := publicEndpoints.Group("/endpoints/:endpoint_name")
		endpoint.Use(
			middleware.EndpointLookupMiddleware(endpointService),
			middleware.EndpointAuthMiddleware(endpointService),
			middleware.EndpointRateLimitMiddleware(),
			middleware.EndpointCORSMiddleware(),
		)
		{
			// SSE transport
			endpoint.GET("/sse", handlers.HandleEndpointSSE(namespaceService))
			endpoint.POST("/message", handlers.HandleEndpointSSEMessage(namespaceService))

			// HTTP transport (MCP protocol)
			endpoint.Any("/mcp", handlers.HandleEndpointHTTP(namespaceService))

			// WebSocket transport
			endpoint.GET("/ws", handlers.HandleEndpointWebSocket(namespaceService))

			// OpenAPI/REST interface
			endpoint.GET("/api/openapi.json", handlers.HandleEndpointOpenAPI(endpointService, namespaceService, baseURL))
			endpoint.GET("/api/docs", handlers.HandleEndpointOpenAPI(endpointService, namespaceService, baseURL))
			endpoint.GET("/api/tools", handlers.HandleEndpointToolsList(namespaceService))
			endpoint.POST("/api/tools/:tool_name", handlers.HandleEndpointToolExecution(namespaceService))

			// Health check
			endpoint.GET("/health", handlers.HandleEndpointHealth())
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
