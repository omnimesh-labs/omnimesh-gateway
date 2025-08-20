package gateway

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"mcp-gateway/apps/backend/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MCPProxy handles MCP server proxy functionality
type MCPProxy struct {
	config        *types.MCPProxyConfig
	sessions      map[string]*types.MCPProxySession
	sessionsMutex sync.RWMutex
	// websocketUpgrader websocket.Upgrader // Removed for now - will implement later with proper websocket library
}

// NewMCPProxy creates a new MCP proxy
func NewMCPProxy(config *types.MCPProxyConfig) *MCPProxy {
	if config == nil {
		config = &types.MCPProxyConfig{
			MaxConcurrentSessions: 100,
			SessionTimeout:        30 * time.Minute,
			ProcessTimeout:        5 * time.Minute,
			BufferSize:            4096,
			EnableLogging:         true,
			LogLevel:              "info",
		}
	}

	return &MCPProxy{
		config:   config,
		sessions: make(map[string]*types.MCPProxySession),
	}
}

// CreateSession creates a new MCP session
func (p *MCPProxy) CreateSession(userID, organizationID string, server *types.MCPServer) (*types.MCPProxySession, error) {
	p.sessionsMutex.Lock()
	defer p.sessionsMutex.Unlock()

	// Check concurrent session limit
	if len(p.sessions) >= p.config.MaxConcurrentSessions {
		return nil, fmt.Errorf("maximum concurrent sessions reached")
	}

	sessionID := uuid.New().String()
	session := &types.MCPProxySession{
		ID:             sessionID,
		UserID:         userID,
		OrganizationID: organizationID,
		ServerID:       server.ID,
		Server:         server,
		Protocol:       server.Protocol,
		Status:         types.SessionStatusInitializing,
		StartedAt:      time.Now(),
		LastActivity:   time.Now(),
		Metadata:       make(map[string]interface{}),
	}

	// Initialize session based on protocol
	var err error
	switch server.Protocol {
	case types.ProtocolStdio:
		err = p.initializeStdioSession(session)
	case types.ProtocolHTTP, types.ProtocolHTTPS:
		err = p.initializeHTTPSession(session)
	default:
		err = fmt.Errorf("unsupported protocol: %s", server.Protocol)
	}

	if err != nil {
		session.Status = types.SessionStatusError
		return nil, err
	}

	session.Status = types.SessionStatusActive
	p.sessions[sessionID] = session

	// Start session cleanup goroutine
	go p.sessionCleanup(sessionID)

	return session, nil
}

// initializeStdioSession starts a stdio-based MCP server process
func (p *MCPProxy) initializeStdioSession(session *types.MCPProxySession) error {
	server := session.Server

	if server.Command == "" {
		return fmt.Errorf("command is required for stdio protocol")
	}

	// Create command
	cmd := exec.Command(server.Command, server.Args...)

	// Set working directory if specified
	if server.WorkingDir != "" {
		cmd.Dir = server.WorkingDir
	}

	// Set environment variables
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, server.Environment...)

	// Create pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdin.Close()
		stdout.Close()
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		stderr.Close()
		return fmt.Errorf("failed to start process: %w", err)
	}

	// Create process info
	process := &types.MCPProcess{
		PID:       cmd.Process.Pid,
		Command:   server.Command,
		Args:      server.Args,
		Status:    types.ProcessStatusRunning,
		StartedAt: time.Now(),
	}

	// Store pipes and process in session
	session.Process = process
	session.StdinPipe = stdin
	session.StdoutPipe = stdout
	session.StderrPipe = stderr

	// Monitor process in background
	go p.monitorProcess(session, cmd)

	return nil
}

// initializeHTTPSession prepares an HTTP-based MCP session
func (p *MCPProxy) initializeHTTPSession(session *types.MCPProxySession) error {
	client := &http.Client{
		Timeout: session.Server.Timeout,
	}

	// Store HTTP client - using interface{} to avoid type issues for now
	var clientInterface interface{} = client
	session.HTTPClient = &clientInterface
	return nil
}

// HandleWebSocket handles WebSocket connections for MCP communication
func (p *MCPProxy) HandleWebSocket(c *gin.Context) {
	// TODO: Implement WebSocket handling when websocket library is available
	c.JSON(http.StatusNotImplemented, gin.H{"error": "WebSocket handling not yet implemented"})
}

// TODO: Implement WebSocket message handling when websocket library is available

// TODO: Implement sendHTTPRequest when needed

// GetSession retrieves a session by ID
func (p *MCPProxy) getSession(sessionID string) *types.MCPProxySession {
	p.sessionsMutex.RLock()
	defer p.sessionsMutex.RUnlock()
	return p.sessions[sessionID]
}

// CloseSession closes and cleans up a session
func (p *MCPProxy) CloseSession(sessionID string) error {
	p.sessionsMutex.Lock()
	defer p.sessionsMutex.Unlock()

	session, exists := p.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	// Close session based on protocol
	switch session.Protocol {
	case types.ProtocolStdio:
		p.closeStdioSession(session)
	case types.ProtocolHTTP, types.ProtocolHTTPS:
		p.closeHTTPSession(session)
	}

	// Update session status
	now := time.Now()
	session.Status = types.SessionStatusClosed
	session.EndedAt = &now

	// Remove from active sessions
	delete(p.sessions, sessionID)

	return nil
}

// closeStdioSession closes a stdio session
func (p *MCPProxy) closeStdioSession(session *types.MCPProxySession) {
	if stdin, ok := session.StdinPipe.(io.WriteCloser); ok {
		stdin.Close()
	}
	if stdout, ok := session.StdoutPipe.(io.ReadCloser); ok {
		stdout.Close()
	}
	if stderr, ok := session.StderrPipe.(io.ReadCloser); ok {
		stderr.Close()
	}

	// Terminate process if still running
	if session.Process != nil && session.Process.Status == types.ProcessStatusRunning {
		// In a real implementation, you'd want to gracefully terminate first
		// then force kill if necessary
		if err := exec.Command("kill", fmt.Sprintf("%d", session.Process.PID)).Run(); err != nil {
			// Log error
		}
	}
}

// closeHTTPSession closes an HTTP session
func (p *MCPProxy) closeHTTPSession(session *types.MCPProxySession) {
	// Nothing special needed for HTTP sessions
}

// monitorProcess monitors a stdio process
func (p *MCPProxy) monitorProcess(session *types.MCPProxySession, cmd *exec.Cmd) {
	err := cmd.Wait()

	now := time.Now()
	session.Process.EndedAt = &now

	if err != nil {
		session.Process.Status = types.ProcessStatusError
		session.Process.Error = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode := exitError.ExitCode()
			session.Process.ExitCode = &exitCode
		}
	} else {
		session.Process.Status = types.ProcessStatusStopped
		exitCode := 0
		session.Process.ExitCode = &exitCode
	}

	// Mark session as closed if process ended unexpectedly
	if session.Status == types.SessionStatusActive {
		session.Status = types.SessionStatusError
		session.EndedAt = &now
	}
}

// sessionCleanup handles session cleanup after timeout
func (p *MCPProxy) sessionCleanup(sessionID string) {
	timer := time.NewTimer(p.config.SessionTimeout)
	defer timer.Stop()

	<-timer.C
	session := p.getSession(sessionID)
	if session != nil && time.Since(session.LastActivity) > p.config.SessionTimeout {
		_ = p.CloseSession(sessionID) // Ignore error for cleanup
	}
}

// ListSessions returns all active sessions for an organization
func (p *MCPProxy) ListSessions(organizationID string) []*types.MCPProxySession {
	p.sessionsMutex.RLock()
	defer p.sessionsMutex.RUnlock()

	var sessions []*types.MCPProxySession
	for _, session := range p.sessions {
		if session.OrganizationID == organizationID {
			sessions = append(sessions, session)
		}
	}

	return sessions
}
