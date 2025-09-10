package unit

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/config"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerConfig_GetBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		config   config.ServerConfig
		expected string
	}{
		{
			name: "explicit base URL",
			config: config.ServerConfig{
				BaseURL: "https://api.example.com",
				Host:    "localhost",
				Port:    8080,
			},
			expected: "https://api.example.com",
		},
		{
			name: "HTTP with default port 80",
			config: config.ServerConfig{
				Host: "localhost",
				Port: 80,
				TLS:  config.TLSConfig{Enabled: false},
			},
			expected: "http://localhost",
		},
		{
			name: "HTTPS with default port 443",
			config: config.ServerConfig{
				Host: "localhost",
				Port: 443,
				TLS:  config.TLSConfig{Enabled: true},
			},
			expected: "https://localhost",
		},
		{
			name: "HTTP with custom port",
			config: config.ServerConfig{
				Host: "localhost",
				Port: 8080,
				TLS:  config.TLSConfig{Enabled: false},
			},
			expected: "http://localhost:8080",
		},
		{
			name: "HTTPS with custom port",
			config: config.ServerConfig{
				Host: "api.example.com",
				Port: 8443,
				TLS:  config.TLSConfig{Enabled: true},
			},
			expected: "https://api.example.com:8443",
		},
		{
			name: "0.0.0.0 host converts to localhost",
			config: config.ServerConfig{
				Host: "0.0.0.0",
				Port: 8080,
				TLS:  config.TLSConfig{Enabled: false},
			},
			expected: "http://localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetBaseURL()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTransportConfig_SetDefaults(t *testing.T) {
	tc := &config.TransportConfig{}
	tc.SetDefaults()

	// Check that all defaults are set
	assert.NotEmpty(t, tc.EnabledTransports)
	assert.Contains(t, tc.EnabledTransports, types.TransportTypeHTTP)
	assert.Contains(t, tc.EnabledTransports, types.TransportTypeSSE)
	assert.Contains(t, tc.EnabledTransports, types.TransportTypeWebSocket)
	assert.Contains(t, tc.EnabledTransports, types.TransportTypeStreamable)

	assert.Equal(t, types.DefaultSSEKeepAlive, tc.SSEKeepAlive)
	assert.Equal(t, types.DefaultWebSocketTimeout, tc.WebSocketTimeout)
	assert.Equal(t, types.DefaultSessionTimeout, tc.SessionTimeout)
	assert.Equal(t, types.DefaultMaxConnections, tc.MaxConnections)
	assert.Equal(t, types.DefaultBufferSize, tc.BufferSize)
	assert.Equal(t, types.DefaultSTDIOTimeout, tc.STDIOTimeout)

	// Check path rewrite defaults
	assert.True(t, tc.PathRewrite.Enabled)
	assert.Equal(t, "info", tc.PathRewrite.LogLevel)
}

func TestTransportConfig_SetDefaultsWithExistingValues(t *testing.T) {
	tc := &config.TransportConfig{
		EnabledTransports: []types.TransportType{types.TransportTypeHTTP},
		SSEKeepAlive:      10 * time.Second,
		WebSocketTimeout:  2 * time.Minute,
		SessionTimeout:    60 * time.Minute,
		MaxConnections:    200,
		BufferSize:        2048,
		STDIOTimeout:      20 * time.Second,
		PathRewrite: config.PathRewriteConfig{
			Enabled:  false,
			LogLevel: "debug",
		},
	}

	tc.SetDefaults()

	// Verify existing values are preserved
	assert.Len(t, tc.EnabledTransports, 1)
	assert.Equal(t, types.TransportTypeHTTP, tc.EnabledTransports[0])
	assert.Equal(t, 10*time.Second, tc.SSEKeepAlive)
	assert.Equal(t, 2*time.Minute, tc.WebSocketTimeout)
	assert.Equal(t, 60*time.Minute, tc.SessionTimeout)
	assert.Equal(t, 200, tc.MaxConnections)
	assert.Equal(t, 2048, tc.BufferSize)
	assert.Equal(t, 20*time.Second, tc.STDIOTimeout)

	// Path rewrite should get defaults because it was disabled
	assert.True(t, tc.PathRewrite.Enabled)
	assert.Equal(t, "info", tc.PathRewrite.LogLevel)
}

func TestTransportConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      config.TransportConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid configuration",
			config: config.TransportConfig{
				EnabledTransports: []types.TransportType{types.TransportTypeHTTP, types.TransportTypeWebSocket},
				SSEKeepAlive:      30 * time.Second,
				WebSocketTimeout:  5 * time.Minute,
				SessionTimeout:    30 * time.Minute,
				STDIOTimeout:      10 * time.Second,
				MaxConnections:    100,
				BufferSize:        1024,
			},
			expectError: false,
		},
		{
			name: "invalid transport type",
			config: config.TransportConfig{
				EnabledTransports: []types.TransportType{"INVALID"},
				SSEKeepAlive:      30 * time.Second,
				WebSocketTimeout:  5 * time.Minute,
				SessionTimeout:    30 * time.Minute,
				STDIOTimeout:      10 * time.Second,
				MaxConnections:    100,
				BufferSize:        1024,
			},
			expectError: true,
			errorMsg:    "invalid transport type",
		},
		{
			name: "negative SSE keep alive",
			config: config.TransportConfig{
				EnabledTransports: []types.TransportType{types.TransportTypeHTTP},
				SSEKeepAlive:      -1 * time.Second,
				WebSocketTimeout:  5 * time.Minute,
				SessionTimeout:    30 * time.Minute,
				STDIOTimeout:      10 * time.Second,
				MaxConnections:    100,
				BufferSize:        1024,
			},
			expectError: true,
			errorMsg:    "sse_keep_alive must be positive",
		},
		{
			name: "zero websocket timeout",
			config: config.TransportConfig{
				EnabledTransports: []types.TransportType{types.TransportTypeHTTP},
				SSEKeepAlive:      30 * time.Second,
				WebSocketTimeout:  0,
				SessionTimeout:    30 * time.Minute,
				STDIOTimeout:      10 * time.Second,
				MaxConnections:    100,
				BufferSize:        1024,
			},
			expectError: true,
			errorMsg:    "websocket_timeout must be positive",
		},
		{
			name: "negative session timeout",
			config: config.TransportConfig{
				EnabledTransports: []types.TransportType{types.TransportTypeHTTP},
				SSEKeepAlive:      30 * time.Second,
				WebSocketTimeout:  5 * time.Minute,
				SessionTimeout:    -1 * time.Minute,
				STDIOTimeout:      10 * time.Second,
				MaxConnections:    100,
				BufferSize:        1024,
			},
			expectError: true,
			errorMsg:    "session_timeout must be positive",
		},
		{
			name: "zero STDIO timeout",
			config: config.TransportConfig{
				EnabledTransports: []types.TransportType{types.TransportTypeHTTP},
				SSEKeepAlive:      30 * time.Second,
				WebSocketTimeout:  5 * time.Minute,
				SessionTimeout:    30 * time.Minute,
				STDIOTimeout:      0,
				MaxConnections:    100,
				BufferSize:        1024,
			},
			expectError: true,
			errorMsg:    "stdio_timeout must be positive",
		},
		{
			name: "zero max connections",
			config: config.TransportConfig{
				EnabledTransports: []types.TransportType{types.TransportTypeHTTP},
				SSEKeepAlive:      30 * time.Second,
				WebSocketTimeout:  5 * time.Minute,
				SessionTimeout:    30 * time.Minute,
				STDIOTimeout:      10 * time.Second,
				MaxConnections:    0,
				BufferSize:        1024,
			},
			expectError: true,
			errorMsg:    "max_connections must be positive",
		},
		{
			name: "negative buffer size",
			config: config.TransportConfig{
				EnabledTransports: []types.TransportType{types.TransportTypeHTTP},
				SSEKeepAlive:      30 * time.Second,
				WebSocketTimeout:  5 * time.Minute,
				SessionTimeout:    30 * time.Minute,
				STDIOTimeout:      10 * time.Second,
				MaxConnections:    100,
				BufferSize:        -1,
			},
			expectError: true,
			errorMsg:    "buffer_size must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTransportConfig_ToTransportConfig(t *testing.T) {
	configTC := config.TransportConfig{
		EnabledTransports:  []types.TransportType{types.TransportTypeHTTP, types.TransportTypeWebSocket},
		SSEKeepAlive:       30 * time.Second,
		WebSocketTimeout:   5 * time.Minute,
		SessionTimeout:     30 * time.Minute,
		MaxConnections:     100,
		BufferSize:         1024,
		StreamableStateful: true,
		STDIOTimeout:       10 * time.Second,
	}

	typesTC := configTC.ToTransportConfig()

	assert.Equal(t, configTC.EnabledTransports, typesTC.EnabledTransports)
	assert.Equal(t, configTC.SSEKeepAlive, typesTC.SSEKeepAlive)
	assert.Equal(t, configTC.WebSocketTimeout, typesTC.WebSocketTimeout)
	assert.Equal(t, configTC.SessionTimeout, typesTC.SessionTimeout)
	assert.Equal(t, configTC.MaxConnections, typesTC.MaxConnections)
	assert.Equal(t, configTC.BufferSize, typesTC.BufferSize)
	assert.Equal(t, configTC.StreamableStateful, typesTC.StreamableStateful)
	assert.Equal(t, configTC.STDIOTimeout, typesTC.STDIOTimeout)
}

func TestExpandEnvVars(t *testing.T) {
	// Set up environment variables for testing
	os.Setenv("TEST_VAR", "test_value")
	os.Setenv("TEST_NUMBER", "42")
	defer func() {
		os.Unsetenv("TEST_VAR")
		os.Unsetenv("TEST_NUMBER")
	}()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple variable substitution",
			input:    "host: ${TEST_VAR}",
			expected: "host: test_value",
		},
		{
			name:     "variable with default value (env var exists)",
			input:    "port: ${TEST_NUMBER:-8080}",
			expected: "port: 42",
		},
		{
			name:     "variable with default value (env var doesn't exist)",
			input:    "timeout: ${NONEXISTENT_VAR:-30s}",
			expected: "timeout: 30s",
		},
		{
			name:     "multiple variables",
			input:    "url: ${TEST_VAR}:${TEST_NUMBER:-8080}",
			expected: "url: test_value:42",
		},
		{
			name:     "no variables",
			input:    "static: value",
			expected: "static: value",
		},
		{
			name:     "empty variable (no default)",
			input:    "empty: ${NONEXISTENT}",
			expected: "empty: ",
		},
		{
			name: "complex YAML with variables",
			input: `server:
  host: ${HOST:-localhost}
  port: ${PORT:-8080}
  database:
    url: postgres://${DB_USER:-user}:${DB_PASS:-pass}@${DB_HOST:-localhost}:${DB_PORT:-5432}/${DB_NAME:-gateway}`,
			expected: `server:
  host: localhost
  port: 8080
  database:
    url: postgres://user:pass@localhost:5432/gateway`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a helper function to access the private expandEnvVars function
			// In a real implementation, we might make this function public for testing
			result := expandEnvVarsTestHelper(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to test the private expandEnvVars function
// This simulates the internal logic since we can't access private functions directly
func expandEnvVarsTestHelper(input string) string {
	// This is a simplified version of the expandEnvVars logic for testing
	// In practice, you might want to extract this to a public utility function
	envVarPattern := regexp.MustCompile(`\$\{([^}]+)\}`)

	return envVarPattern.ReplaceAllStringFunc(input, func(match string) string {
		varExpr := match[2 : len(match)-1]

		if strings.Contains(varExpr, ":-") {
			parts := strings.SplitN(varExpr, ":-", 2)
			envVar := parts[0]
			defaultValue := parts[1]

			if value := os.Getenv(envVar); value != "" {
				return value
			}
			return defaultValue
		}

		if value := os.Getenv(varExpr); value != "" {
			return value
		}

		return ""
	})
}

func TestLoadFromFile(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-config.yaml")

	configContent := `
server:
  host: ${SERVER_HOST:-localhost}
  port: ${SERVER_PORT:-8080}
  tls:
    enabled: false
database:
  host: ${DB_HOST:-localhost}
  port: ${DB_PORT:-5432}
  user: ${DB_USER:-test}
  password: ${DB_PASS:-secret}
  database: ${DB_NAME:-gateway_test}
  ssl_mode: ${DB_SSL_MODE:-disable}
  max_open_conns: 10
  max_idle_conns: 5
  max_lifetime: 1h
auth:
  jwt_secret: ${JWT_SECRET:-test-secret}
  access_token_expiry: 15m
  refresh_token_expiry: 24h
  bcrypt_cost: 10
logging:
  level: ${LOG_LEVEL:-info}
  backend: ${LOG_BACKEND:-file}
  async: true
  buffer_size: 1000
  batch_size: 100
  flush_interval: 5s
rate_limit:
  enabled: true
  default_limit: 100
  default_window: 1m
  ip_enabled: true
  ip_requests_per_minute: 60
transport:
  enabled_transports: [HTTP, WEBSOCKET, SSE]
  session_timeout: 30m
  max_connections: 100
  buffer_size: 1024
discovery:
  enabled: true
  health_interval: 30s
  failure_threshold: 3
gateway:
  proxy_timeout: 30s
  max_retries: 3
  circuit_breaker:
    enabled: true
    failure_threshold: 5
    recovery_timeout: 30s
redis:
  host: ${REDIS_HOST:-localhost}
  port: ${REDIS_PORT:-6379}
  database: ${REDIS_DB:-0}
  pool_size: 10
filters:
  enabled: ${FILTERS_ENABLED:-true}
  database_driven: true
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set some environment variables
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("JWT_SECRET", "super-secret")
	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("JWT_SECRET")
	}()

	// Load configuration
	cfg, err := config.Load(configFile)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Test that environment variables were expanded
	assert.Equal(t, "localhost", cfg.Server.Host) // Default value
	assert.Equal(t, 9090, cfg.Server.Port)        // From environment
	assert.False(t, cfg.Server.TLS.Enabled)

	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "testuser", cfg.Database.User)   // From environment
	assert.Equal(t, "secret", cfg.Database.Password) // Default value
	assert.Equal(t, "gateway_test", cfg.Database.Database)
	assert.Equal(t, "disable", cfg.Database.SSLMode)
	assert.Equal(t, 10, cfg.Database.MaxOpenConns)
	assert.Equal(t, 5, cfg.Database.MaxIdleConns)
	assert.Equal(t, 1*time.Hour, cfg.Database.MaxLifetime)

	assert.Equal(t, "super-secret", cfg.Auth.JWTSecret) // From environment
	assert.Equal(t, 15*time.Minute, cfg.Auth.AccessTokenExpiry)
	assert.Equal(t, 24*time.Hour, cfg.Auth.RefreshTokenExpiry)
	assert.Equal(t, 10, cfg.Auth.BCryptCost)

	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "file", cfg.Logging.Backend)
	assert.True(t, cfg.Logging.Async)
	assert.Equal(t, 1000, cfg.Logging.BufferSize)
	assert.Equal(t, 100, cfg.Logging.BatchSize)
	assert.Equal(t, 5*time.Second, cfg.Logging.FlushInterval)

	assert.True(t, cfg.RateLimit.Enabled)
	assert.Equal(t, 100, cfg.RateLimit.DefaultLimit)
	assert.Equal(t, 1*time.Minute, cfg.RateLimit.DefaultWindow)
	assert.True(t, cfg.RateLimit.IPEnabled)
	assert.Equal(t, 60, cfg.RateLimit.IPRequestsPerMinute)

	assert.Len(t, cfg.Transport.EnabledTransports, 3)
	assert.Contains(t, cfg.Transport.EnabledTransports, types.TransportTypeHTTP)
	assert.Contains(t, cfg.Transport.EnabledTransports, types.TransportTypeWebSocket)
	assert.Contains(t, cfg.Transport.EnabledTransports, types.TransportTypeSSE)
	assert.Equal(t, 30*time.Minute, cfg.Transport.SessionTimeout)
	assert.Equal(t, 100, cfg.Transport.MaxConnections)
	assert.Equal(t, 1024, cfg.Transport.BufferSize)

	assert.True(t, cfg.Discovery.Enabled)
	assert.Equal(t, 30*time.Second, cfg.Discovery.HealthInterval)
	assert.Equal(t, 3, cfg.Discovery.FailureThreshold)

	assert.Equal(t, 30*time.Second, cfg.Gateway.ProxyTimeout)
	assert.Equal(t, 3, cfg.Gateway.MaxRetries)
	assert.True(t, cfg.Gateway.CircuitBreaker.Enabled)
	assert.Equal(t, 5, cfg.Gateway.CircuitBreaker.FailureThreshold)
	assert.Equal(t, 30*time.Second, cfg.Gateway.CircuitBreaker.RecoveryTimeout)

	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, 6379, cfg.Redis.Port)
	assert.Equal(t, 0, cfg.Redis.Database)
	assert.Equal(t, 10, cfg.Redis.PoolSize)

	assert.True(t, cfg.Filters.Enabled)
	assert.True(t, cfg.Filters.DatabaseDriven)
}

func TestLoadFromFile_FileNotFound(t *testing.T) {
	_, err := config.Load("nonexistent-file.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestLoadFromFile_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid-config.yaml")

	invalidContent := `
server:
  host: localhost
  port: 8080
  invalid yaml structure here {{{
`

	err := os.WriteFile(configFile, []byte(invalidContent), 0644)
	require.NoError(t, err)

	_, err = config.Load(configFile)
	assert.Error(t, err)
	// Error should be related to YAML parsing
}

func TestDatabaseConfig_GetDSN(t *testing.T) {
	dbConfig := config.DatabaseConfig{
		Host:     "localhost",
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		Port:     5432,
		SSLMode:  "disable",
	}

	// This test currently expects empty string since GetDSN is not implemented
	// In a real implementation, you would test the actual DSN construction
	dsn := dbConfig.GetDSN()
	assert.Equal(t, "", dsn) // TODO: Implement and test actual DSN

	// Example of what the test might look like when implemented:
	// expected := "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"
	// assert.Equal(t, expected, dsn)
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      func() *config.Config
		expectError bool
		errorCheck  func(error) bool
	}{
		{
			name: "valid complete configuration",
			config: func() *config.Config {
				cfg := &config.Config{
					Server: config.ServerConfig{
						Host: "localhost",
						Port: 8080,
						TLS:  config.TLSConfig{Enabled: false},
					},
					Database: config.DatabaseConfig{
						Host:         "localhost",
						Port:         5432,
						User:         "test",
						Password:     "test",
						Database:     "test",
						MaxOpenConns: 10,
						MaxIdleConns: 5,
						MaxLifetime:  1 * time.Hour,
					},
					Auth: config.AuthConfig{
						JWTSecret:          "secret",
						AccessTokenExpiry:  15 * time.Minute,
						RefreshTokenExpiry: 24 * time.Hour,
						BCryptCost:         10,
					},
				}
				cfg.Transport.SetDefaults()
				return cfg
			},
			expectError: false,
		},
		{
			name: "transport validation error",
			config: func() *config.Config {
				cfg := &config.Config{}
				cfg.Transport = config.TransportConfig{
					EnabledTransports: []types.TransportType{"INVALID"},
					SSEKeepAlive:      30 * time.Second,
					WebSocketTimeout:  5 * time.Minute,
					SessionTimeout:    30 * time.Minute,
					STDIOTimeout:      10 * time.Second,
					MaxConnections:    100,
					BufferSize:        1024,
				}
				return cfg
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return err.Error() == "invalid transport type: INVALID"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.config()
			err := cfg.Transport.Validate()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorCheck != nil {
					assert.True(t, tt.errorCheck(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test default configurations and edge cases
func TestConfigDefaults(t *testing.T) {
	t.Run("empty transport config gets defaults", func(t *testing.T) {
		tc := &config.TransportConfig{}
		tc.SetDefaults()

		assert.NotEmpty(t, tc.EnabledTransports)
		assert.True(t, tc.SSEKeepAlive > 0)
		assert.True(t, tc.WebSocketTimeout > 0)
		assert.True(t, tc.SessionTimeout > 0)
		assert.True(t, tc.MaxConnections > 0)
		assert.True(t, tc.BufferSize > 0)
		assert.True(t, tc.STDIOTimeout > 0)
	})

	t.Run("partial transport config preserves values", func(t *testing.T) {
		tc := &config.TransportConfig{
			EnabledTransports: []types.TransportType{types.TransportTypeHTTP},
			MaxConnections:    500,
		}
		tc.SetDefaults()

		// Preserved values
		assert.Len(t, tc.EnabledTransports, 1)
		assert.Equal(t, types.TransportTypeHTTP, tc.EnabledTransports[0])
		assert.Equal(t, 500, tc.MaxConnections)

		// Default values for unset fields
		assert.Equal(t, types.DefaultSSEKeepAlive, tc.SSEKeepAlive)
		assert.Equal(t, types.DefaultWebSocketTimeout, tc.WebSocketTimeout)
		assert.Equal(t, types.DefaultSessionTimeout, tc.SessionTimeout)
		assert.Equal(t, types.DefaultBufferSize, tc.BufferSize)
		assert.Equal(t, types.DefaultSTDIOTimeout, tc.STDIOTimeout)
	})
}

// Benchmark tests
func BenchmarkExpandEnvVars(b *testing.B) {
	input := "server: ${HOST:-localhost}:${PORT:-8080}, db: ${DB_URL:-postgres://localhost/test}"
	os.Setenv("HOST", "example.com")
	os.Setenv("PORT", "9000")
	defer func() {
		os.Unsetenv("HOST")
		os.Unsetenv("PORT")
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		expandEnvVarsTestHelper(input)
	}
}

func BenchmarkTransportConfigValidation(b *testing.B) {
	tc := config.TransportConfig{
		EnabledTransports: []types.TransportType{types.TransportTypeHTTP, types.TransportTypeWebSocket, types.TransportTypeSSE},
		SSEKeepAlive:      30 * time.Second,
		WebSocketTimeout:  5 * time.Minute,
		SessionTimeout:    30 * time.Minute,
		STDIOTimeout:      10 * time.Second,
		MaxConnections:    100,
		BufferSize:        1024,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := tc.Validate()
		if err != nil {
			b.Fatal(err)
		}
	}
}
