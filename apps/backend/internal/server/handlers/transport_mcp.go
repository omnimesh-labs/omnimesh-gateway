package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/middleware"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/transport"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"

	"github.com/gin-gonic/gin"
)

// StreamableTransport interface defines methods needed for streamable HTTP transport
type StreamableTransport interface {
	types.Transport
	GetStreamMode() string
	GetEventsSince(time.Time) []*types.TransportEvent
	GetLatestEvents(int) []*types.TransportEvent
	IsStateful() bool
}

// MCPHandler handles Streamable HTTP transport endpoints
type MCPHandler struct {
	transportManager *transport.Manager
}

// NewMCPHandler creates a new MCP handler
func NewMCPHandler(transportManager *transport.Manager) *MCPHandler {
	return &MCPHandler{
		transportManager: transportManager,
	}
}

// HandleStreamableHTTP handles Streamable HTTP requests
func (h *MCPHandler) HandleStreamableHTTP(c *gin.Context) {
	// Get transport context
	transportCtx := middleware.GetTransportContext(c)
	if transportCtx == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "transport context not found",
		})
		return
	}

	// Determine stream mode based on Accept header
	accept := c.GetHeader("Accept")
	var streamMode string
	if strings.Contains(accept, "text/event-stream") {
		streamMode = types.StreamableModeSSE
	} else {
		streamMode = types.StreamableModeJSON
	}

	// Create streamable transport connection
	streamableTransport, session, err := h.transportManager.CreateConnection(
		c.Request.Context(),
		types.TransportTypeStreamable,
		transportCtx.UserID,
		transportCtx.OrganizationID,
		transportCtx.ServerID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create streamable connection: " + err.Error(),
		})
		return
	}

	// Configure transport using interface methods if available
	if streamModeSetter, ok := streamableTransport.(interface{ SetStreamMode(string) }); ok {
		streamModeSetter.SetStreamMode(streamMode)
	}
	if session != nil {
		streamableTransport.SetSessionID(session.ID)
		if statefulSetter, ok := streamableTransport.(interface{ SetStateful(bool) }); ok {
			statefulSetter.SetStateful(true)
		}
	}

	// Connect transport
	if err := streamableTransport.Connect(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to connect streamable transport: " + err.Error(),
		})
		return
	}

	// Cast to StreamableTransport interface
	streamableImpl, ok := streamableTransport.(StreamableTransport)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Transport does not implement required streamable methods",
		})
		return
	}

	// Handle based on HTTP method and stream mode
	switch c.Request.Method {
	case "GET":
		h.handleStreamableGET(c, streamableImpl, session)
	case "POST":
		h.handleStreamablePOST(c, streamableImpl, session)
	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error": "Method not allowed",
		})
	}
}

// handleStreamableGET handles GET requests for streamable HTTP
func (h *MCPHandler) handleStreamableGET(c *gin.Context, transport StreamableTransport, session *types.TransportSession) {
	// Determine stream mode based on Accept header
	accept := c.GetHeader("Accept")
	var streamMode string
	if strings.Contains(accept, "text/event-stream") {
		streamMode = types.StreamableModeSSE
	} else {
		streamMode = types.StreamableModeJSON
	}

	switch streamMode {
	case types.StreamableModeSSE:
		h.handleStreamableSSEGET(c, transport, session)
	case types.StreamableModeJSON:
		h.handleStreamableJSONGET(c, transport, session)
	}
}

// handleStreamableSSEGET handles SSE mode GET requests
func (h *MCPHandler) handleStreamableSSEGET(c *gin.Context, transport StreamableTransport, session *types.TransportSession) {
	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	if session != nil {
		c.Header("X-Session-ID", session.ID)
	}

	// Send initial connection event
	var sessionID string
	if session != nil {
		sessionID = session.ID
	}
	c.SSEvent("connected", gin.H{
		"status":     "connected",
		"session_id": sessionID,
		"timestamp":  time.Now(),
	})

	// For GET in SSE mode, we stream events from the session's event store
	if session != nil {
		// Get recent events
		events := transport.GetEventsSince(time.Now().Add(-1 * time.Hour)) // Last hour

		for _, event := range events {
			c.SSEvent(event.Type, gin.H{
				"id":        event.ID,
				"data":      event.Data,
				"timestamp": event.Timestamp,
			})
		}
	}

	// Keep connection alive until client disconnects
	<-c.Request.Context().Done()

	// Clean up
	if session != nil {
		h.transportManager.CloseConnection(session.ID)
	}
}

// handleStreamableJSONGET handles JSON mode GET requests
func (h *MCPHandler) handleStreamableJSONGET(c *gin.Context, transport StreamableTransport, session *types.TransportSession) {
	var response gin.H

	if session != nil && transport.IsStateful() {
		// Return session information and recent events
		events := transport.GetEventsSince(time.Now().Add(-1 * time.Hour))

		response = gin.H{
			"status":     "connected",
			"session_id": session.ID,
			"stateful":   true,
			"mode":       transport.GetStreamMode(),
			"events":     events,
			"timestamp":  time.Now(),
		}
	} else {
		// Return connection information
		response = gin.H{
			"status":    "ready",
			"stateful":  false,
			"mode":      transport.GetStreamMode(),
			"timestamp": time.Now(),
		}
	}

	c.JSON(http.StatusOK, response)
}

// handleStreamablePOST handles POST requests for streamable HTTP
func (h *MCPHandler) handleStreamablePOST(c *gin.Context, transport StreamableTransport, session *types.TransportSession) {
	// Parse streamable request
	var streamableReq types.StreamableHTTPRequest
	if err := c.ShouldBindJSON(&streamableReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid streamable request: " + err.Error(),
		})
		return
	}

	// Send message through transport
	if err := h.transportManager.SendMessage(c.Request.Context(), session.ID, &streamableReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to send message: " + err.Error(),
		})
		return
	}

	// Handle response based on Accept header stream mode
	accept := c.GetHeader("Accept")
	var responseStreamMode string
	if strings.Contains(accept, "text/event-stream") {
		responseStreamMode = types.StreamableModeSSE
	} else {
		responseStreamMode = types.StreamableModeJSON
	}

	if responseStreamMode == types.StreamableModeSSE {
		h.handleStreamableSSEResponse(c, transport, session)
	} else {
		h.handleStreamableJSONResponse(c, transport, session, &streamableReq)
	}
}

// handleStreamableSSEResponse handles SSE responses
func (h *MCPHandler) handleStreamableSSEResponse(c *gin.Context, transport StreamableTransport, session *types.TransportSession) {
	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	if session != nil {
		c.Header("X-Session-ID", session.ID)
	}

	// Send response as SSE event
	var sessionID string
	if session != nil {
		sessionID = session.ID
	}
	c.SSEvent("response", gin.H{
		"status":     "processed",
		"session_id": sessionID,
		"timestamp":  time.Now(),
	})

	// Stream any additional events
	events := transport.GetLatestEvents(5)
	for _, event := range events {
		c.SSEvent(event.Type, gin.H{
			"id":        event.ID,
			"data":      event.Data,
			"timestamp": event.Timestamp,
		})
	}
}

// handleStreamableJSONResponse handles JSON responses
func (h *MCPHandler) handleStreamableJSONResponse(c *gin.Context, transport StreamableTransport, session *types.TransportSession, req *types.StreamableHTTPRequest) {
	var response gin.H

	if session != nil && transport.IsStateful() {
		// Get recent events for stateful response
		events := transport.GetLatestEvents(10)

		response = gin.H{
			"status":     "processed",
			"session_id": session.ID,
			"stateful":   true,
			"mode":       transport.GetStreamMode(),
			"events":     events,
			"timestamp":  time.Now(),
		}
	} else {
		// Stateless response
		response = gin.H{
			"status":    "processed",
			"stateful":  false,
			"mode":      transport.GetStreamMode(),
			"timestamp": time.Now(),
		}
	}

	c.JSON(http.StatusOK, response)
}

// HandleServerStreamableHTTP handles server-specific streamable HTTP
func (h *MCPHandler) HandleServerStreamableHTTP(c *gin.Context) {
	// This is handled by path rewriting middleware
	// which transforms /servers/{server_id}/mcp to /mcp
	h.HandleStreamableHTTP(c)
}

// HandleMCPCapabilities returns MCP capabilities
func (h *MCPHandler) HandleMCPCapabilities(c *gin.Context) {
	capabilities := gin.H{
		"capabilities": gin.H{
			"roots": gin.H{
				"listChanged": true,
			},
			"sampling": gin.H{},
			"tools": gin.H{
				"listChanged": true,
			},
			"resources": gin.H{
				"listChanged": true,
				"subscribe":   true,
			},
			"prompts": gin.H{
				"listChanged": true,
			},
			"logging": gin.H{},
		},
		"protocol_version": "2024-11-05",
		"server_info": gin.H{
			"name":    "Omnimesh Gateway",
			"version": "1.0.0",
		},
		"transports": []string{
			string(types.TransportTypeHTTP),
			string(types.TransportTypeSSE),
			string(types.TransportTypeWebSocket),
			string(types.TransportTypeStreamable),
			string(types.TransportTypeSTDIO),
		},
		"modes": []string{
			types.StreamableModeJSON,
			types.StreamableModeSSE,
		},
	}

	c.JSON(http.StatusOK, capabilities)
}

// HandleMCPStatus provides status information about streamable connections
func (h *MCPHandler) HandleMCPStatus(c *gin.Context) {
	// Get query parameters
	serverID := c.Query("server_id")
	userID := c.Query("user_id")
	_ = c.Query("mode") // mode filtering would require session metadata

	// Get active sessions
	activeSessions := h.transportManager.GetActiveSessions()

	// Filter streamable sessions
	var streamableSessions []*types.TransportSession
	for _, session := range activeSessions {
		if session.TransportType != types.TransportTypeStreamable {
			continue
		}

		// Apply filters
		if serverID != "" && session.ServerID != serverID {
			continue
		}
		if userID != "" && session.UserID != userID {
			continue
		}
		// Note: mode filtering would require session metadata

		streamableSessions = append(streamableSessions, session)
	}

	// Get metrics
	metrics := h.transportManager.GetMetrics()

	response := gin.H{
		"active_sessions": len(streamableSessions),
		"total_sessions":  len(streamableSessions), // For compatibility with test expectations
		"transport_mode":  types.TransportTypeStreamable,
		"sessions":        streamableSessions,
		"metrics":         metrics,
		"supported_modes": []string{types.StreamableModeJSON, types.StreamableModeSSE},
	}

	c.JSON(http.StatusOK, response)
}

// HandleMCPHealth checks streamable HTTP transport health
func (h *MCPHandler) HandleMCPHealth(c *gin.Context) {
	healthResults := h.transportManager.HealthCheck(c.Request.Context())

	streamableHealth, exists := healthResults[types.TransportTypeStreamable]
	if !exists {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "Streamable HTTP transport not configured",
		})
		return
	}

	if streamableHealth != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  streamableHealth.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"transport": types.TransportTypeStreamable,
		"timestamp": time.Now(),
		"capabilities": []string{
			"stateful",
			"stateless",
			"json_mode",
			"sse_mode",
			"event_store",
			"session_management",
		},
	})
}
