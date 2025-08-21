package handlers

import (
	"net/http"
	"strconv"
	"time"

	"mcp-gateway/apps/backend/internal/middleware"
	"mcp-gateway/apps/backend/internal/transport"
	"mcp-gateway/apps/backend/internal/types"

	"github.com/gin-gonic/gin"
)

// SSEHandler handles Server-Sent Events transport endpoints
type SSEHandler struct {
	transportManager *transport.Manager
}

// NewSSEHandler creates a new SSE handler
func NewSSEHandler(transportManager *transport.Manager) *SSEHandler {
	return &SSEHandler{
		transportManager: transportManager,
	}
}

// HandleSSE handles SSE connections
func (h *SSEHandler) HandleSSE(c *gin.Context) {
	// Get transport context
	transportCtx := middleware.GetTransportContext(c)
	if transportCtx == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "transport context not found",
		})
		return
	}

	// Create SSE transport connection
	sseTransport, session, err := h.transportManager.CreateConnection(
		c.Request.Context(),
		types.TransportTypeSSE,
		transportCtx.UserID,
		transportCtx.OrganizationID,
		transportCtx.ServerID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create SSE connection: " + err.Error(),
		})
		return
	}

	// Set up SSE connection using interface method if available
	if sseSetup, ok := sseTransport.(interface {
		SetupSSE(http.ResponseWriter, *http.Request) error
	}); ok {
		if err := sseSetup.SetupSSE(c.Writer, c.Request); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to setup SSE: " + err.Error(),
			})
			return
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "SSE setup not supported by this transport implementation",
		})
		return
	}

	// Connect transport
	if err := sseTransport.Connect(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to connect SSE transport: " + err.Error(),
		})
		return
	}

	// Set session ID in transport
	if session != nil {
		sseTransport.SetSessionID(session.ID)
	}

	// Keep connection alive until client disconnects
	<-c.Request.Context().Done()

	// Clean up connection
	h.transportManager.CloseConnection(session.ID)
}

// HandleServerSSE handles server-specific SSE connections
func (h *SSEHandler) HandleServerSSE(c *gin.Context) {
	// This is handled by path rewriting middleware
	// which transforms /servers/{server_id}/sse to /sse
	h.HandleSSE(c)
}

// HandleSSEEvents handles posting events to SSE streams
func (h *SSEHandler) HandleSSEEvents(c *gin.Context) {
	// Get session ID
	sessionID := middleware.GetSessionID(c)
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "session_id is required",
		})
		return
	}

	// Parse event data
	var eventData struct {
		Event string      `json:"event"`
		Data  interface{} `json:"data"`
		ID    string      `json:"id,omitempty"`
		Retry int         `json:"retry,omitempty"`
	}

	if err := c.ShouldBindJSON(&eventData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid event data: " + err.Error(),
		})
		return
	}

	// Get SSE transport connection
	_, err := h.transportManager.GetConnection(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "SSE connection not found: " + err.Error(),
		})
		return
	}

	// Create SSE event
	sseEvent := &types.SSEEvent{
		ID:        eventData.ID,
		Event:     eventData.Event,
		Data:      eventData.Data,
		Retry:     eventData.Retry,
		Timestamp: time.Now(),
	}

	// Send event
	if err := h.transportManager.SendMessage(c.Request.Context(), sessionID, sseEvent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to send SSE event: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Event sent successfully",
		"session_id": sessionID,
		"event_id":   sseEvent.ID,
	})
}

// HandleSSEBroadcast handles broadcasting events to all SSE connections
func (h *SSEHandler) HandleSSEBroadcast(c *gin.Context) {
	// Parse broadcast data
	var broadcastData struct {
		Event    string      `json:"event"`
		Data     interface{} `json:"data"`
		ServerID string      `json:"server_id,omitempty"`
		UserID   string      `json:"user_id,omitempty"`
		OrgID    string      `json:"org_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&broadcastData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid broadcast data: " + err.Error(),
		})
		return
	}

	// Create SSE event
	sseEvent := &types.SSEEvent{
		Event:     broadcastData.Event,
		Data:      broadcastData.Data,
		Timestamp: time.Now(),
	}

	// Broadcast to all SSE connections
	// This is a simplified implementation - in practice you'd filter by server/user/org
	if err := h.transportManager.BroadcastMessage(c.Request.Context(), types.TransportTypeSSE, sseEvent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to broadcast SSE event: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Event broadcasted successfully",
		"event":   broadcastData.Event,
	})
}

// HandleSSEStatus provides status information about SSE connections
func (h *SSEHandler) HandleSSEStatus(c *gin.Context) {
	// Get query parameters
	serverID := c.Query("server_id")
	userID := c.Query("user_id")

	// Get active sessions
	activeSessions := h.transportManager.GetActiveSessions()

	// Filter SSE sessions
	var sseSessions []*types.TransportSession
	for _, session := range activeSessions {
		if session.TransportType != types.TransportTypeSSE {
			continue
		}

		// Apply filters
		if serverID != "" && session.ServerID != serverID {
			continue
		}
		if userID != "" && session.UserID != userID {
			continue
		}

		sseSessions = append(sseSessions, session)
	}

	// Get metrics
	metrics := h.transportManager.GetMetrics()

	response := gin.H{
		"active_connections": len(sseSessions),
		"sessions":           sseSessions,
		"metrics":            metrics,
		"transport_type":     types.TransportTypeSSE,
	}

	c.JSON(http.StatusOK, response)
}

// HandleSSEReplay handles replaying events from a specific point
func (h *SSEHandler) HandleSSEReplay(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "session_id is required",
		})
		return
	}

	// Get last event ID or timestamp
	lastEventID := c.Query("last_event_id")
	sinceParam := c.Query("since")
	limitParam := c.Query("limit")

	// Parse limit
	limit := 50 // Default limit
	if limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Parse since timestamp
	var since *time.Time
	if sinceParam != "" {
		if parsedTime, err := time.Parse(time.RFC3339, sinceParam); err == nil {
			since = &parsedTime
		}
	}

	// Get events for replay
	events, err := h.transportManager.GetSessionEvents(sessionID, since, limit)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Failed to get session events: " + err.Error(),
		})
		return
	}

	// Filter events by last event ID if provided
	if lastEventID != "" {
		var filteredEvents []types.TransportEvent
		foundLastEvent := false

		for _, event := range events {
			if foundLastEvent {
				filteredEvents = append(filteredEvents, event)
			} else if event.ID == lastEventID {
				foundLastEvent = true
			}
		}

		if foundLastEvent {
			events = filteredEvents
		}
	}

	response := gin.H{
		"session_id":    sessionID,
		"events":        events,
		"count":         len(events),
		"last_event_id": lastEventID,
		"since":         since,
		"limit":         limit,
	}

	c.JSON(http.StatusOK, response)
}

// HandleSSEHealth checks SSE transport health
func (h *SSEHandler) HandleSSEHealth(c *gin.Context) {
	// Check if SSE transport is registered and available
	// Don't require an active connection for health check

	// Try to create a test SSE transport to verify it's available
	_, _, err := h.transportManager.CreateConnection(
		c.Request.Context(),
		types.TransportTypeSSE,
		"health-check",
		"health-check",
		"health-check",
	)

	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "SSE transport not available: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"transport": types.TransportTypeSSE,
		"timestamp": time.Now(),
		"capabilities": []string{
			"streaming",
			"server_push",
			"keep_alive",
			"replay",
		},
	})
}
