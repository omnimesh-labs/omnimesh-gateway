package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"mcp-gateway/apps/backend/internal/inspector"
)

// InspectorHandler handles inspector-related HTTP requests
type InspectorHandler struct {
	service inspector.InspectorService
}

// NewInspectorHandler creates a new inspector handler
func NewInspectorHandler(service inspector.InspectorService) *InspectorHandler {
	return &InspectorHandler{
		service: service,
	}
}

// CreateSession creates a new inspector session
func (h *InspectorHandler) CreateSession(c *gin.Context) {
	var req inspector.CreateSessionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user info from context (set by auth middleware)
	userID := c.GetString("user_id")
	orgID := c.GetString("org_id")

	if userID == "" || orgID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Create session
	session, err := h.service.CreateSession(c.Request.Context(), req.ServerID, userID, orgID, req.NamespaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, session)
}

// GetSession retrieves a session by ID
func (h *InspectorHandler) GetSession(c *gin.Context) {
	sessionID := c.Param("id")

	session, err := h.service.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Check if user owns this session
	userID := c.GetString("user_id")
	if session.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// CloseSession closes an inspector session
func (h *InspectorHandler) CloseSession(c *gin.Context) {
	sessionID := c.Param("id")

	// Verify session ownership
	session, err := h.service.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	if session.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	// Close session
	if err := h.service.CloseSession(c.Request.Context(), sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "session closed successfully"})
}

// ExecuteRequest executes an MCP request on a session
func (h *InspectorHandler) ExecuteRequest(c *gin.Context) {
	sessionID := c.Param("id")

	// Verify session ownership
	session, err := h.service.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	if session.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	// Parse request body
	var reqBody inspector.ExecuteRequestBody
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create inspector request
	req := inspector.InspectorRequest{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Method:    reqBody.Method,
		Params:    reqBody.Params,
		Timestamp: time.Now(),
	}

	// Execute request
	response, err := h.service.ExecuteRequest(c.Request.Context(), sessionID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// StreamEvents streams events for a session using Server-Sent Events
func (h *InspectorHandler) StreamEvents(c *gin.Context) {
	sessionID := c.Param("id")

	// Verify session ownership
	session, err := h.service.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	if session.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Get event channel
	events, err := h.service.GetEventChannel(sessionID)
	if err != nil {
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", err.Error())
		c.Writer.Flush()
		return
	}

	// Create context for handling client disconnect
	ctx := c.Request.Context()

	// Send events
	for {
		select {
		case event, ok := <-events:
			if !ok {
				// Channel closed, session ended
				fmt.Fprintf(c.Writer, "event: close\ndata: session closed\n\n")
				c.Writer.Flush()
				return
			}

			// Marshal event to JSON
			data, err := json.Marshal(event)
			if err != nil {
				fmt.Fprintf(c.Writer, "event: error\ndata: failed to marshal event\n\n")
			} else {
				fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event.Type, data)
			}

			// Flush the data immediately
			if f, ok := c.Writer.(http.Flusher); ok {
				f.Flush()
			}

		case <-ctx.Done():
			// Client disconnected
			return

		case <-time.After(30 * time.Second):
			// Send heartbeat to keep connection alive
			fmt.Fprintf(c.Writer, ": heartbeat\n\n")
			if f, ok := c.Writer.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for now (configure properly in production)
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// HandleWebSocket handles WebSocket connections for inspector sessions
func (h *InspectorHandler) HandleWebSocket(c *gin.Context) {
	sessionID := c.Param("id")

	// Verify session ownership
	session, err := h.service.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	if session.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// Get event channel
	events, err := h.service.GetEventChannel(sessionID)
	if err != nil {
		conn.WriteJSON(gin.H{"error": err.Error()})
		return
	}

	// Create context for handling disconnect
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// Handle incoming messages (requests from client)
	go func() {
		defer cancel()
		for {
			var reqBody inspector.ExecuteRequestBody
			if err := conn.ReadJSON(&reqBody); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					fmt.Printf("WebSocket error: %v\n", err)
				}
				break
			}

			// Create and execute request
			req := inspector.InspectorRequest{
				ID:        uuid.New().String(),
				SessionID: sessionID,
				Method:    reqBody.Method,
				Params:    reqBody.Params,
				Timestamp: time.Now(),
			}

			// Execute request in goroutine to not block reading
			go func(req inspector.InspectorRequest) {
				response, err := h.service.ExecuteRequest(ctx, sessionID, req)
				if err != nil {
					conn.WriteJSON(gin.H{"error": err.Error()})
				} else {
					conn.WriteJSON(response)
				}
			}(req)
		}
	}()

	// Send events to client
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case event, ok := <-events:
			if !ok {
				// Channel closed
				conn.WriteJSON(gin.H{"type": "close", "message": "session closed"})
				return
			}

			if err := conn.WriteJSON(event); err != nil {
				return
			}

		case <-ticker.C:
			// Send ping to keep connection alive
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-ctx.Done():
			// Context cancelled (client disconnected or error)
			return
		}
	}
}

// GetServerCapabilities returns the capabilities of an MCP server
func (h *InspectorHandler) GetServerCapabilities(c *gin.Context) {
	serverID := c.Param("id")

	capabilities, err := h.service.GetServerCapabilities(c.Request.Context(), serverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, capabilities)
}