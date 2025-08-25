package helpers

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"mcp-gateway/apps/backend/internal/config"
	"mcp-gateway/apps/backend/internal/database"
	"mcp-gateway/apps/backend/internal/logging"
	"mcp-gateway/apps/backend/internal/logging/plugins/file"
	"mcp-gateway/apps/backend/internal/server"
	"mcp-gateway/apps/backend/internal/types"
)

var (
	pluginRegistered sync.Once
)

// TestServer represents a test server instance
type TestServer struct {
	Server   *http.Server
	shutdown chan struct{}
	Address  string
	BaseURL  string
	Port     int
}

// NewTestServer creates a new test server instance
func NewTestServer() (*TestServer, error) {
	// Register the file logging plugin once globally
	pluginRegistered.Do(func() {
		logging.RegisterPlugin(file.NewFilePlugin())
	})

	// Reset database instance to ensure fresh connection with test env vars
	database.ResetForTests()

	// Set environment variables to prevent server from re-registering plugins and configure database
	os.Setenv("SKIP_PLUGIN_REGISTRATION", "true")
	os.Setenv("SKIP_CONTENT_FILTERING", "true")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_DATABASE", "postgres")
	os.Setenv("DB_USERNAME", "postgres")
	os.Setenv("DB_PASSWORD", "changeme123")
	os.Setenv("DB_SCHEMA", "public")

	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, fmt.Errorf("failed to find available port: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Create test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: port,
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "changeme123",
			Database: "postgres",
			SSLMode:  "disable",
		},
		Auth: config.AuthConfig{
			JWTSecret:          "test-secret",
			AccessTokenExpiry:  time.Hour,
			RefreshTokenExpiry: 24 * time.Hour,
			BCryptCost:         10,
		},
		Logging: config.LoggingConfig{
			Level:         "info",
			Backend:       "file",
			Environment:   "test",
			BufferSize:    100,
			BatchSize:     10,
			FlushInterval: time.Second,
			Async:         false,
			Config: map[string]interface{}{
				"file_path": "/tmp/test.log",
			},
		},
		Transport: config.TransportConfig{
			EnabledTransports:  []types.TransportType{types.TransportTypeHTTP, types.TransportTypeSSE, types.TransportTypeWebSocket, types.TransportTypeStreamable, types.TransportTypeSTDIO},
			MaxConnections:     100,
			SessionTimeout:     time.Hour,
			BufferSize:         1024,
			SSEKeepAlive:       30 * time.Second,
			WebSocketTimeout:   60 * time.Second,
			StreamableStateful: false,
			STDIOTimeout:       30 * time.Second,
		},
	}

	address := fmt.Sprintf("localhost:%d", port)
	baseURL := fmt.Sprintf("http://%s", address)

	// Create server with specific address
	srv := server.NewServer(cfg)
	srv.Addr = address

	return &TestServer{
		Server:   srv,
		Address:  address,
		Port:     port,
		BaseURL:  baseURL,
		shutdown: make(chan struct{}),
	}, nil
}

// Start starts the test server
func (ts *TestServer) Start() error {
	go func() {
		if err := ts.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("Failed to start test server: %v", err))
		}
	}()

	// Wait for server to be ready
	return ts.WaitForReady(10 * time.Second)
}

// Stop stops the test server
func (ts *TestServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	close(ts.shutdown)
	return ts.Server.Shutdown(ctx)
}

// WaitForReady waits for the server to be ready
func (ts *TestServer) WaitForReady(timeout time.Duration) error {
	client := &http.Client{Timeout: 1 * time.Second}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := client.Get(ts.BaseURL + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("server did not become ready within %v", timeout)
}

// GetURL returns the full URL for a given path
func (ts *TestServer) GetURL(path string) string {
	if path[0] != '/' {
		path = "/" + path
	}
	return ts.BaseURL + path
}
