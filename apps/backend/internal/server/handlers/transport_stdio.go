package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"mcp-gateway/apps/backend/internal/middleware"
	"mcp-gateway/apps/backend/internal/transport"
	"mcp-gateway/apps/backend/internal/types"

	"github.com/gin-gonic/gin"
)

// STDIOHandler handles STDIO transport endpoints
type STDIOHandler struct {
	transportManager *transport.Manager
}

// NewSTDIOHandler creates a new STDIO handler
func NewSTDIOHandler(transportManager *transport.Manager) *STDIOHandler {
	return &STDIOHandler{
		transportManager: transportManager,
	}
}

// HandleSTDIOExecute handles STDIO command execution
func (h *STDIOHandler) HandleSTDIOExecute(c *gin.Context) {
	// Get transport context
	transportCtx := middleware.GetTransportContext(c)
	if transportCtx == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "transport context not found",
		})
		return
	}

	// Parse STDIO command
	var stdioCmd types.STDIOCommand
	if err := c.ShouldBindJSON(&stdioCmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid STDIO command: " + err.Error(),
		})
		return
	}

	// Validate command
	if stdioCmd.Command == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "command is required",
		})
		return
	}

	// Create STDIO transport with command configuration
	config := map[string]interface{}{
		"command":     stdioCmd.Command,
		"args":        stdioCmd.Args,
		"env":         stdioCmd.Env,
		"working_dir": stdioCmd.Dir,
		"timeout":     stdioCmd.Timeout,
	}

	// Create STDIO transport connection
	stdioTransport, session, err := h.transportManager.CreateConnectionWithConfig(
		c.Request.Context(),
		types.TransportTypeSTDIO,
		transportCtx.UserID,
		transportCtx.OrganizationID,
		transportCtx.ServerID,
		config,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create STDIO connection: " + err.Error(),
		})
		return
	}

	// Connect transport (starts the subprocess)
	if err := stdioTransport.Connect(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start STDIO process: " + err.Error(),
		})
		return
	}

	// Set session ID if available
	if session != nil {
		stdioTransport.SetSessionID(session.ID)
	}

	// Execute the command using interface method if available
	var response interface{}
	if executor, ok := stdioTransport.(interface {
		ExecuteCommand(context.Context, *types.STDIOCommand) (interface{}, error)
	}); ok {
		response, err = executor.ExecuteCommand(c.Request.Context(), &stdioCmd)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to execute command: " + err.Error(),
			})
			return
		}
	} else {
		// Fallback to sending the command as a message
		if err := stdioTransport.SendMessage(c.Request.Context(), &stdioCmd); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to send command: " + err.Error(),
			})
			return
		}
		response = map[string]interface{}{"status": "command_sent"}
	}

	// Clean up connection after execution
	defer func() {
		if session != nil {
			h.transportManager.CloseConnection(session.ID)
		}
	}()

	// Get PID using interface method if available
	var pid int
	if pidGetter, ok := stdioTransport.(interface{ GetPID() int }); ok {
		pid = pidGetter.GetPID()
	}

	// Return response
	var sessionID string
	if session != nil {
		sessionID = session.ID
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"command":    stdioCmd.Command,
		"args":       stdioCmd.Args,
		"pid":        pid,
		"response":   response,
		"session_id": sessionID,
		"timestamp":  time.Now(),
	})
}

// HandleSTDIOProcess handles managing long-running STDIO processes
func (h *STDIOHandler) HandleSTDIOProcess(c *gin.Context) {
	action := c.Query("action") // start, stop, restart, status
	if action == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "action parameter is required (start, stop, restart, status)",
		})
		return
	}

	switch action {
	case "start":
		h.handleSTDIOStart(c)
	case "stop":
		h.handleSTDIOStop(c)
	case "restart":
		h.handleSTDIORestart(c)
	case "status":
		h.handleSTDIOStatus(c)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid action. Must be: start, stop, restart, or status",
		})
	}
}

// handleSTDIOStart starts a STDIO process
func (h *STDIOHandler) handleSTDIOStart(c *gin.Context) {
	// Get transport context
	transportCtx := middleware.GetTransportContext(c)
	if transportCtx == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "transport context not found",
		})
		return
	}

	// Parse process configuration
	var processConfig struct {
		Command    string            `json:"command" binding:"required"`
		Args       []string          `json:"args"`
		Env        map[string]string `json:"env"`
		WorkingDir string            `json:"working_dir"`
		Timeout    time.Duration     `json:"timeout"`
	}

	if err := c.ShouldBindJSON(&processConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid process configuration: " + err.Error(),
		})
		return
	}

	// Create STDIO transport with process configuration
	config := map[string]interface{}{
		"command":     processConfig.Command,
		"args":        processConfig.Args,
		"env":         processConfig.Env,
		"working_dir": processConfig.WorkingDir,
		"timeout":     processConfig.Timeout,
	}

	// Update config with defaults
	if config["timeout"] == nil {
		config["timeout"] = 30 * time.Second
	}

	stdioTransport, session, err := h.transportManager.CreateConnectionWithConfig(
		c.Request.Context(),
		types.TransportTypeSTDIO,
		transportCtx.UserID,
		transportCtx.OrganizationID,
		transportCtx.ServerID,
		config,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create STDIO transport: " + err.Error(),
		})
		return
	}

	// Start the process
	if err := stdioTransport.Connect(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start process: " + err.Error(),
		})
		return
	}

	// Get PID using interface method if available
	var pid int
	if pidGetter, ok := stdioTransport.(interface{ GetPID() int }); ok {
		pid = pidGetter.GetPID()
	}

	var sessionID string
	if session != nil {
		sessionID = session.ID
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "started",
		"session_id": sessionID,
		"pid":        pid,
		"command":    processConfig.Command,
		"args":       processConfig.Args,
		"timestamp":  time.Now(),
	})
}

// handleSTDIOStop stops a STDIO process
func (h *STDIOHandler) handleSTDIOStop(c *gin.Context) {
	sessionID := middleware.GetSessionID(c)
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "session_id is required",
		})
		return
	}

	// Get transport
	transport, err := h.transportManager.GetConnection(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "STDIO process not found: " + err.Error(),
		})
		return
	}

	// Get PID using interface method if available
	var pid int
	if pidGetter, ok := transport.(interface{ GetPID() int }); ok {
		pid = pidGetter.GetPID()
	}

	// Stop the process
	if err := h.transportManager.CloseConnection(sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to stop process: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "stopped",
		"session_id": sessionID,
		"pid":        pid,
		"timestamp":  time.Now(),
	})
}

// handleSTDIORestart restarts a STDIO process
func (h *STDIOHandler) handleSTDIORestart(c *gin.Context) {
	sessionID := middleware.GetSessionID(c)
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "session_id is required",
		})
		return
	}

	// Get transport
	transport, err := h.transportManager.GetConnection(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "STDIO process not found: " + err.Error(),
		})
		return
	}

	// Restart the process using interface method if available
	if restarter, ok := transport.(interface{ RestartProcess(context.Context) error }); ok {
		if err := restarter.RestartProcess(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to restart process: " + err.Error(),
			})
			return
		}
	} else {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error": "Restart not supported for this transport type",
		})
		return
	}

	// Get PID after restart
	var pid int
	if pidGetter, ok := transport.(interface{ GetPID() int }); ok {
		pid = pidGetter.GetPID()
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "restarted",
		"session_id": sessionID,
		"pid":        pid,
		"timestamp":  time.Now(),
	})
}

// handleSTDIOStatus gets status of a STDIO process
func (h *STDIOHandler) handleSTDIOStatus(c *gin.Context) {
	sessionID := c.Query("session_id")

	if sessionID != "" {
		// Get status for specific session
		transport, err := h.transportManager.GetConnection(sessionID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "STDIO process not found: " + err.Error(),
			})
			return
		}

		// Get metrics using interface method if available
		var metrics map[string]interface{}
		if metricsGetter, ok := transport.(interface{ GetMetrics() map[string]interface{} }); ok {
			metrics = metricsGetter.GetMetrics()
		} else {
			metrics = map[string]interface{}{
				"transport_type": types.TransportTypeSTDIO,
				"session_id":     sessionID,
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"session_id": sessionID,
			"status":     "running",
			"metrics":    metrics,
			"timestamp":  time.Now(),
		})
		return
	}

	// Get status for all STDIO processes
	activeSessions := h.transportManager.GetActiveSessions()

	var stdioSessions []*types.TransportSession
	for _, session := range activeSessions {
		if session.TransportType == types.TransportTypeSTDIO {
			stdioSessions = append(stdioSessions, session)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"transport_type":   types.TransportTypeSTDIO,
		"active_processes": len(stdioSessions),
		"sessions":         stdioSessions,
		"timestamp":        time.Now(),
	})
}

// HandleSTDIOSend handles sending messages to STDIO processes
func (h *STDIOHandler) HandleSTDIOSend(c *gin.Context) {
	sessionID := middleware.GetSessionID(c)
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "session_id is required",
		})
		return
	}

	// Parse MCP message
	var mcpMessage types.MCPMessage
	if err := c.ShouldBindJSON(&mcpMessage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid MCP message: " + err.Error(),
		})
		return
	}

	// Send message to STDIO process
	if err := h.transportManager.SendMessage(c.Request.Context(), sessionID, &mcpMessage); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to send message: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "sent",
		"session_id": sessionID,
		"message_id": mcpMessage.ID,
		"method":     mcpMessage.Method,
		"timestamp":  time.Now(),
	})
}

// HandleSTDIOHealth checks STDIO transport health
func (h *STDIOHandler) HandleSTDIOHealth(c *gin.Context) {
	// Check if STDIO transport is registered and available
	// Don't require a command configuration for health check

	// Try to create a test STDIO transport with a dummy command to verify it's available
	_, _, err := h.transportManager.CreateConnection(
		c.Request.Context(),
		types.TransportTypeSTDIO,
		"health-check",
		"health-check",
		"health-check",
	)

	// For STDIO, we expect this to fail with "command is required" which means the transport is available
	// but needs proper configuration. Any other error means the transport isn't available.
	if err != nil && !strings.Contains(err.Error(), "command is required") {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "STDIO transport not available: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"transport": types.TransportTypeSTDIO,
		"timestamp": time.Now(),
		"capabilities": []string{
			"subprocess_management",
			"command_execution",
			"process_monitoring",
			"environment_variables",
			"working_directory",
			"timeout_support",
		},
	})
}
