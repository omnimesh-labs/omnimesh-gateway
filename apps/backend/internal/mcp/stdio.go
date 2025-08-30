package mcp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

// StdioTransport implements the Transport interface for stdio-based MCP servers
type StdioTransport struct{}

// Type returns the transport type identifier
func (st *StdioTransport) Type() string {
	return "stdio"
}

// Connect establishes a stdio connection to an MCP server
func (st *StdioTransport) Connect(ctx context.Context, config TransportConfig) (Connection, error) {
	if config.Command == "" {
		return nil, fmt.Errorf("command is required for stdio transport")
	}

	// Create the command
	cmd := exec.CommandContext(ctx, config.Command, config.Args...)

	// Set environment variables
	if len(config.Environment) > 0 {
		env := os.Environ()
		for key, value := range config.Environment {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = env
	}

	// Set working directory
	if config.WorkingDir != "" {
		cmd.Dir = config.WorkingDir
	}

	// Create pipes for stdin/stdout communication
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdin.Close()
		stdout.Close()
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		stderr.Close()
		return nil, fmt.Errorf("failed to start process: %w", err)
	}

	conn := &StdioConnection{
		cmd:       cmd,
		stdin:     stdin,
		stdout:    stdout,
		stderr:    stderr,
		connected: true,
		mu:        &sync.RWMutex{},
		messages:  make(chan []byte, 100),
		errors:    make(chan error, 10),
	}

	// Start reading from stdout in a separate goroutine
	go conn.readLoop(ctx)

	// Start reading from stderr for debugging
	go conn.readStderr(ctx)

	return conn, nil
}

// StdioConnection represents a stdio connection to an MCP server
type StdioConnection struct {
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	stderr    io.ReadCloser
	connected bool
	mu        *sync.RWMutex
	messages  chan []byte
	errors    chan error
}

// Send sends a message to the MCP server via stdin
func (sc *StdioConnection) Send(ctx context.Context, message []byte) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.connected {
		return fmt.Errorf("connection is closed")
	}

	// Add newline if not present (required for JSON-RPC over stdio)
	if len(message) > 0 && message[len(message)-1] != '\n' {
		message = append(message, '\n')
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err := sc.stdin.Write(message)
		return err
	}
}

// Receive receives a message from the MCP server via stdout
func (sc *StdioConnection) Receive(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-sc.errors:
		return nil, err
	case message := <-sc.messages:
		return message, nil
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("timeout waiting for message")
	}
}

// Close closes the stdio connection
func (sc *StdioConnection) Close() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if !sc.connected {
		return nil
	}

	sc.connected = false

	// Close pipes
	if sc.stdin != nil {
		sc.stdin.Close()
	}
	if sc.stdout != nil {
		sc.stdout.Close()
	}
	if sc.stderr != nil {
		sc.stderr.Close()
	}

	// Terminate the process
	if sc.cmd != nil && sc.cmd.Process != nil {
		sc.cmd.Process.Kill()
		sc.cmd.Wait()
	}

	// Close channels
	close(sc.messages)
	close(sc.errors)

	return nil
}

// IsConnected returns whether the connection is active
func (sc *StdioConnection) IsConnected() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.connected
}

// readLoop reads messages from stdout and sends them to the messages channel
func (sc *StdioConnection) readLoop(ctx context.Context) {
	defer func() {
		sc.mu.Lock()
		sc.connected = false
		sc.mu.Unlock()
	}()

	scanner := bufio.NewScanner(sc.stdout)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		case sc.messages <- scanner.Bytes():
		case <-time.After(5 * time.Second):
			// Drop message if channel is full and we can't send within timeout
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		select {
		case sc.errors <- fmt.Errorf("stdout read error: %w", err):
		default:
		}
	}
}

// readStderr reads from stderr for debugging purposes
func (sc *StdioConnection) readStderr(ctx context.Context) {
	scanner := bufio.NewScanner(sc.stderr)
	for scanner.Scan() {
		// Log stderr messages for debugging
		// In a production system, you might want to send these to a logger
		_ = scanner.Text()
	}
}
