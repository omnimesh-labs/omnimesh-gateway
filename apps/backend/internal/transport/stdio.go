package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"mcp-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
)

// STDIOTransport implements STDIO transport for command-line MCP servers
type STDIOTransport struct {
	stdout       io.ReadCloser
	stderr       io.ReadCloser
	stdin        io.WriteCloser
	responseMap  map[string]chan *types.MCPMessage
	cmd          *exec.Cmd
	messageQueue chan *types.MCPMessage
	*BaseTransport
	config       map[string]interface{}
	done         chan struct{}
	command      string
	workingDir   string
	args         []string
	env          []string
	timeout      time.Duration
	mu           sync.RWMutex
	cleanupOnce  sync.Once // Ensure cleanup only runs once
}

// NewSTDIOTransport creates a new STDIO transport instance
func NewSTDIOTransport(config map[string]interface{}) (types.Transport, error) {
	transport := &STDIOTransport{
		BaseTransport: NewBaseTransport(types.TransportTypeSTDIO),
		messageQueue:  make(chan *types.MCPMessage, 100),
		responseMap:   make(map[string]chan *types.MCPMessage),
		config:        config,
		done:          make(chan struct{}),
		timeout:       30 * time.Second,
		env:           os.Environ(), // Start with current environment
	}

	// Configure from config map
	if command, ok := config["command"].(string); ok {
		transport.command = command
	} else {
		return nil, fmt.Errorf("command is required for STDIO transport")
	}

	if args, ok := config["args"].([]string); ok {
		transport.args = args
	}

	if timeout, ok := config["stdio_timeout"].(time.Duration); ok {
		transport.timeout = timeout
	}

	if workingDir, ok := config["working_dir"].(string); ok {
		transport.workingDir = workingDir
	}

	if env, ok := config["env"].(map[string]string); ok {
		for key, value := range env {
			transport.env = append(transport.env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	return transport, nil
}

// Connect establishes STDIO connection by starting the command
func (s *STDIOTransport) Connect(ctx context.Context) error {
	if s.cmd != nil {
		return fmt.Errorf("STDIO transport already connected")
	}

	// Create command
	s.cmd = exec.CommandContext(ctx, s.command, s.args...)
	s.cmd.Env = s.env

	if s.workingDir != "" {
		s.cmd.Dir = s.workingDir
	}

	// Set up pipes
	var err error
	s.stdin, err = s.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	s.stdout, err = s.cmd.StdoutPipe()
	if err != nil {
		s.stdin.Close()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	s.stderr, err = s.cmd.StderrPipe()
	if err != nil {
		s.stdin.Close()
		s.stdout.Close()
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := s.cmd.Start(); err != nil {
		s.cleanup()
		return fmt.Errorf("failed to start command: %w", err)
	}

	s.setConnected(true)

	// Start I/O goroutines
	go s.readLoop()
	go s.writeLoop()
	go s.errorLoop()
	go s.processMonitor()

	return nil
}

// Disconnect closes STDIO connection and terminates the command
func (s *STDIOTransport) Disconnect(ctx context.Context) error {
	s.setConnected(false)

	// Signal done to all goroutines safely
	select {
	case <-s.done:
		// Channel already closed
	default:
		close(s.done)
	}

	// Close stdin to signal the process to exit gracefully
	if s.stdin != nil {
		s.stdin.Close()
	}

	// Wait for process to exit with timeout
	if s.cmd != nil && s.cmd.Process != nil {
		done := make(chan error, 1)
		go func() {
			done <- s.cmd.Wait()
		}()

		select {
		case <-done:
			// Process exited normally
		case <-time.After(5 * time.Second):
			// Force kill if it doesn't exit gracefully
			s.cmd.Process.Kill()
			<-done // Wait for the kill to complete
		}
	}

	s.cleanup()
	return nil
}

// SendMessage sends a message via STDIO
func (s *STDIOTransport) SendMessage(ctx context.Context, message interface{}) error {
	if !s.IsConnected() {
		return fmt.Errorf("STDIO transport not connected")
	}

	// Convert message to MCP format
	mcpMessage, err := s.convertToMCPMessage(message)
	if err != nil {
		return fmt.Errorf("failed to convert message: %w", err)
	}

	// Send to message queue
	select {
	case s.messageQueue <- mcpMessage:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-s.done:
		return fmt.Errorf("STDIO transport closed")
	}
}

// ReceiveMessage receives a message via STDIO
func (s *STDIOTransport) ReceiveMessage(ctx context.Context) (interface{}, error) {
	if !s.IsConnected() {
		return nil, fmt.Errorf("STDIO transport not connected")
	}

	// For STDIO, messages are handled asynchronously by readLoop
	// This method can be used to wait for specific responses
	return nil, fmt.Errorf("use SendRequest for synchronous STDIO communication")
}

// SendRequest sends a request and waits for response
func (s *STDIOTransport) SendRequest(ctx context.Context, mcpMessage *types.MCPMessage) (*types.MCPMessage, error) {
	if !s.IsConnected() {
		return nil, fmt.Errorf("STDIO transport not connected")
	}

	// Create response channel
	responseChan := make(chan *types.MCPMessage, 1)
	s.mu.Lock()
	s.responseMap[mcpMessage.ID] = responseChan
	s.mu.Unlock()

	// Clean up response channel
	defer func() {
		s.mu.Lock()
		delete(s.responseMap, mcpMessage.ID)
		s.mu.Unlock()
		close(responseChan)
	}()

	// Send message
	if err := s.SendMessage(ctx, mcpMessage); err != nil {
		return nil, err
	}

	// Wait for response
	select {
	case response := <-responseChan:
		return response, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-s.done:
		return nil, fmt.Errorf("STDIO transport closed")
	case <-time.After(s.timeout):
		return nil, fmt.Errorf("request timeout")
	}
}

// readLoop handles reading from stdout
func (s *STDIOTransport) readLoop() {
	defer func() {
		if s.stdout != nil {
			s.stdout.Close()
		}
	}()

	scanner := bufio.NewScanner(s.stdout)
	for scanner.Scan() {
		select {
		case <-s.done:
			return
		default:
			line := scanner.Text()
			s.handleOutputLine(line)
		}
	}

	if err := scanner.Err(); err != nil {
		// Log scanning error
	}
}

// writeLoop handles writing to stdin
func (s *STDIOTransport) writeLoop() {
	defer func() {
		if s.stdin != nil {
			s.stdin.Close()
		}
	}()

	for {
		select {
		case message, ok := <-s.messageQueue:
			if !ok {
				return
			}

			// Serialize message to JSON
			jsonData, err := json.Marshal(message)
			if err != nil {
				continue
			}

			// Write to stdin with newline
			if _, err := s.stdin.Write(append(jsonData, '\n')); err != nil {
				return
			}

		case <-s.done:
			return
		}
	}
}

// errorLoop handles reading from stderr
func (s *STDIOTransport) errorLoop() {
	defer func() {
		if s.stderr != nil {
			s.stderr.Close()
		}
	}()

	scanner := bufio.NewScanner(s.stderr)
	for scanner.Scan() {
		select {
		case <-s.done:
			return
		default:
			line := scanner.Text()
			// Log stderr output or handle errors
			s.handleErrorLine(line)
		}
	}
}

// processMonitor monitors the subprocess
func (s *STDIOTransport) processMonitor() {
	if s.cmd == nil {
		return
	}

	// Wait for process to exit
	err := s.cmd.Wait()

	// If we're not already disconnecting, handle unexpected exit
	if s.IsConnected() {
		s.setConnected(false)
		// Log process exit with error if any
		if err != nil {
			// Handle process exit error
		}
	}
}

// handleOutputLine processes a line from stdout
func (s *STDIOTransport) handleOutputLine(line string) {
	// Try to parse as JSON MCP message
	var mcpMessage types.MCPMessage
	if err := json.Unmarshal([]byte(line), &mcpMessage); err != nil {
		// Not a valid JSON message, might be plain text output
		return
	}

	// Check if this is a response to a pending request
	s.mu.RLock()
	responseChan, exists := s.responseMap[mcpMessage.ID]
	s.mu.RUnlock()

	if exists {
		// Send response to waiting goroutine
		select {
		case responseChan <- &mcpMessage:
		default:
			// Channel full or closed
		}
		return
	}

	// Handle notifications and other message types
	s.handleNotification(&mcpMessage)
}

// handleErrorLine processes a line from stderr
func (s *STDIOTransport) handleErrorLine(line string) {
	// Log or handle stderr output
	// This could be error messages, debug output, etc.
}

// handleNotification handles notification messages
func (s *STDIOTransport) handleNotification(message *types.MCPMessage) {
	// Handle notifications from the MCP server
	// These are messages that don't require a response
}

// convertToMCPMessage converts various message types to MCP message format
func (s *STDIOTransport) convertToMCPMessage(message interface{}) (*types.MCPMessage, error) {
	switch msg := message.(type) {
	case *types.MCPMessage:
		return msg, nil
	case *types.TransportRequest:
		mcpMessage := &types.MCPMessage{
			ID:      msg.ID,
			Type:    types.MCPMessageTypeRequest,
			Method:  msg.Method,
			Version: "2024-11-05",
			Params:  msg.Parameters,
		}
		return mcpMessage, nil
	case *types.STDIOCommand:
		// Convert STDIO command to MCP message
		mcpMessage := &types.MCPMessage{
			ID:      uuid.New().String(),
			Type:    types.MCPMessageTypeRequest,
			Method:  "stdio/execute",
			Version: "2024-11-05",
			Params: map[string]interface{}{
				"command": msg.Command,
				"args":    msg.Args,
				"env":     msg.Env,
				"dir":     msg.Dir,
				"timeout": msg.Timeout,
			},
		}
		return mcpMessage, nil
	case map[string]interface{}:
		// Try to parse as MCP message
		jsonData, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}
		var mcpMessage types.MCPMessage
		if err := json.Unmarshal(jsonData, &mcpMessage); err != nil {
			return nil, err
		}
		return &mcpMessage, nil
	default:
		return nil, fmt.Errorf("unsupported message type: %T", message)
	}
}

// cleanup cleans up resources safely
func (s *STDIOTransport) cleanup() {
	s.cleanupOnce.Do(func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		// Close file handles
		if s.stdin != nil {
			s.stdin.Close()
			s.stdin = nil
		}
		if s.stdout != nil {
			s.stdout.Close()
			s.stdout = nil
		}
		if s.stderr != nil {
			s.stderr.Close()
			s.stderr = nil
		}

		// Close channels safely
		if s.messageQueue != nil {
			// Use select to avoid panic if channel is already closed
			select {
			case <-s.messageQueue:
				// Channel already closed by someone else
			default:
				close(s.messageQueue)
			}
			s.messageQueue = nil
		}

		// Close all response channels
		for id, ch := range s.responseMap {
			select {
			case <-ch:
				// Channel already closed
			default:
				close(ch)
			}
			delete(s.responseMap, id)
		}

		s.cmd = nil
	})
}

// Helper methods for MCP operations

// CallTool calls an MCP tool via STDIO
func (s *STDIOTransport) CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (*types.MCPMessage, error) {
	mcpMessage := &types.MCPMessage{
		ID:      uuid.New().String(),
		Type:    types.MCPMessageTypeRequest,
		Method:  types.MCPMethodCallTool,
		Version: "2024-11-05",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": arguments,
		},
	}

	return s.SendRequest(ctx, mcpMessage)
}

// ListTools lists available MCP tools via STDIO
func (s *STDIOTransport) ListTools(ctx context.Context) (*types.MCPMessage, error) {
	mcpMessage := &types.MCPMessage{
		ID:      uuid.New().String(),
		Type:    types.MCPMessageTypeRequest,
		Method:  types.MCPMethodListTools,
		Version: "2024-11-05",
		Params:  make(map[string]interface{}),
	}

	return s.SendRequest(ctx, mcpMessage)
}

// ExecuteCommand executes a system command via STDIO bridge
func (s *STDIOTransport) ExecuteCommand(ctx context.Context, command *types.STDIOCommand) (*types.MCPMessage, error) {
	mcpMessage, err := s.convertToMCPMessage(command)
	if err != nil {
		return nil, err
	}
	return s.SendRequest(ctx, mcpMessage)
}

// Configuration and utility methods

// GetCommand returns the command being executed
func (s *STDIOTransport) GetCommand() string {
	return s.command
}

// GetArgs returns the command arguments
func (s *STDIOTransport) GetArgs() []string {
	return s.args
}

// GetPID returns the process ID if running
func (s *STDIOTransport) GetPID() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.cmd != nil && s.cmd.Process != nil {
		return s.cmd.Process.Pid
	}
	return 0
}

// IsProcessRunning checks if the subprocess is still running
func (s *STDIOTransport) IsProcessRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.cmd == nil || s.cmd.Process == nil {
		return false
	}

	// Check if process is still running
	err := s.cmd.Process.Signal(os.Signal(nil))
	return err == nil
}

// GetConfig returns the transport configuration
func (s *STDIOTransport) GetConfig() map[string]interface{} {
	return s.config
}

// GetMetrics returns STDIO transport metrics
func (s *STDIOTransport) GetMetrics() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"connected":          s.IsConnected(),
		"process_running":    s.IsProcessRunning(),
		"pid":                s.GetPID(),
		"command":            s.command,
		"args":               s.args,
		"working_dir":        s.workingDir,
		"message_queue_size": len(s.messageQueue),
		"pending_responses":  len(s.responseMap),
		"timeout":            s.timeout,
		"session_id":         s.GetSessionID(),
	}
}

// RestartProcess restarts the subprocess
func (s *STDIOTransport) RestartProcess(ctx context.Context) error {
	// Disconnect first
	if err := s.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect: %w", err)
	}

	// Wait a moment for cleanup
	time.Sleep(100 * time.Millisecond)

	// Reconnect
	return s.Connect(ctx)
}

// init registers the STDIO transport factory
func init() {
	RegisterTransport(types.TransportTypeSTDIO, NewSTDIOTransport)
}
