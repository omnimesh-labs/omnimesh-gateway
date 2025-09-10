package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/middleware"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/transport"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"

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

	// Read raw request body first
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read request body: " + err.Error(),
		})
		return
	}

	// Parse STDIO command manually to handle timeout field properly
	var rawCmd map[string]interface{}
	if err := json.Unmarshal(body, &rawCmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid STDIO command: " + err.Error(),
		})
		return
	}

	// Convert to STDIOCommand struct
	stdioCmd := types.STDIOCommand{}

	if cmd, ok := rawCmd["command"].(string); ok {
		stdioCmd.Command = cmd
	}

	if dir, ok := rawCmd["dir"].(string); ok {
		stdioCmd.Dir = dir
	}

	// Handle args - can be []interface{} or []string
	if args, ok := rawCmd["args"].([]interface{}); ok {
		for _, arg := range args {
			if argStr, ok := arg.(string); ok {
				stdioCmd.Args = append(stdioCmd.Args, argStr)
			}
		}
	}

	// Handle environment variables
	if envMap, ok := rawCmd["env"].(map[string]interface{}); ok {
		stdioCmd.Env = make(map[string]string)
		for key, value := range envMap {
			if valueStr, ok := value.(string); ok {
				stdioCmd.Env[key] = valueStr
			}
		}
	}

	// Handle timeout - can be integer (nanoseconds) or string duration
	if timeoutRaw, exists := rawCmd["timeout"]; exists {
		switch t := timeoutRaw.(type) {
		case float64:
			stdioCmd.Timeout = time.Duration(int64(t)) * time.Nanosecond
		case int64:
			stdioCmd.Timeout = time.Duration(t) * time.Nanosecond
		case string:
			if parsed, err := time.ParseDuration(t); err == nil {
				stdioCmd.Timeout = parsed
			}
		}
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

	// For simple command execution, execute directly without expecting MCP protocol
	var response interface{}

	// Execute the command directly using os/exec for simple shell commands
	cmdCtx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, stdioCmd.Command, stdioCmd.Args...)

	// Set environment variables if provided
	if stdioCmd.Env != nil {
		for key, value := range stdioCmd.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Set working directory if provided
	if stdioCmd.Dir != "" {
		cmd.Dir = stdioCmd.Dir
	}

	// Execute the command and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Command failed, but we still want to return the output
		response = map[string]interface{}{
			"output": string(output),
			"error":  err.Error(),
			"status": "failed",
		}
	} else {
		// Command succeeded
		response = map[string]interface{}{
			"output": string(output),
			"status": "success",
		}
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

	// Return response with expected structure
	var sessionID string
	if session != nil {
		sessionID = session.ID
	}

	// Extract output from response
	var outputData interface{}
	if response != nil {
		if respMap, ok := response.(map[string]interface{}); ok {
			if respMap["output"] != nil {
				outputData = respMap["output"]
			} else if respMap["result"] != nil {
				outputData = respMap["result"]
			} else {
				outputData = response
			}
		} else {
			outputData = response
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"output":     outputData,
		"command":    stdioCmd.Command,
		"args":       stdioCmd.Args,
		"pid":        pid,
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
		Env        map[string]string `json:"env"`
		Command    string            `json:"command" binding:"required"`
		WorkingDir string            `json:"working_dir"`
		Args       []string          `json:"args"`
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
			// Session doesn't exist, return status with "not_found" status instead of error
			c.JSON(http.StatusOK, gin.H{
				"session_id": sessionID,
				"status":     "not_found",
				"message":    "STDIO process not found",
				"timestamp":  time.Now(),
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
		"processes":        stdioSessions,
		"sessions":         stdioSessions, // TODO: safe to remove this?
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
