package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"mcp-gateway/apps/backend/internal/types"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Redis     RedisConfig     `yaml:"redis"`
	Filters   FiltersConfig   `yaml:"filters"`
	Auth      AuthConfig      `yaml:"auth"`
	Database  DatabaseConfig  `yaml:"database"`
	Server    ServerConfig    `yaml:"server"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	Logging   LoggingConfig   `yaml:"logging"`
	Gateway   GatewayConfig   `yaml:"gateway"`
	Transport TransportConfig `yaml:"transport"`
	Discovery DiscoveryConfig `yaml:"discovery"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string        `yaml:"host" env:"SERVER_HOST"`
	TLS          TLSConfig     `yaml:"tls"`
	Port         int           `yaml:"port" env:"SERVER_PORT"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	CertFile string `yaml:"cert_file" env:"TLS_CERT_FILE"`
	KeyFile  string `yaml:"key_file" env:"TLS_KEY_FILE"`
	Enabled  bool   `yaml:"enabled" env:"TLS_ENABLED"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host         string        `yaml:"host" env:"DB_HOST"`
	User         string        `yaml:"user" env:"DB_USER"`
	Password     string        `yaml:"password" env:"DB_PASSWORD"`
	Database     string        `yaml:"database" env:"DB_NAME"`
	SSLMode      string        `yaml:"ssl_mode" env:"DB_SSL_MODE"`
	Port         int           `yaml:"port" env:"DB_PORT"`
	MaxOpenConns int           `yaml:"max_open_conns"`
	MaxIdleConns int           `yaml:"max_idle_conns"`
	MaxLifetime  time.Duration `yaml:"max_lifetime"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret          string        `yaml:"jwt_secret" env:"JWT_SECRET"`
	AccessTokenExpiry  time.Duration `yaml:"access_token_expiry"`
	RefreshTokenExpiry time.Duration `yaml:"refresh_token_expiry"`
	BCryptCost         int           `yaml:"bcrypt_cost"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Config         map[string]interface{} `yaml:"config"`
	Retention      *RetentionConfig       `yaml:"retention,omitempty"`
	Format         string                 `yaml:"format"`
	Backend        string                 `yaml:"backend" env:"LOG_BACKEND"`
	Environment    string                 `yaml:"environment" env:"ENVIRONMENT"`
	Level          string                 `yaml:"level" env:"LOG_LEVEL"`
	BufferSize     int                    `yaml:"buffer_size"`
	BatchSize      int                    `yaml:"batch_size"`
	FlushInterval  time.Duration          `yaml:"flush_interval"`
	RetentionDays  int                    `yaml:"retention_days"`
	Async          bool                   `yaml:"async"`
	RequestLogging bool                   `yaml:"request_logging"`
	AuditLogging   bool                   `yaml:"audit_logging"`
	MetricsEnabled bool                   `yaml:"metrics_enabled"`
}

// RetentionConfig defines log retention policies
type RetentionConfig struct {
	Policy    string `yaml:"policy"`
	Days      int    `yaml:"days"`
	KeepCount int    `yaml:"keep_count"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Storage         string        `yaml:"storage"`
	DefaultLimit    int           `yaml:"default_limit"`
	DefaultWindow   time.Duration `yaml:"default_window"`
	CleanupInterval time.Duration `yaml:"cleanup_interval"`
	Enabled         bool          `yaml:"enabled"`
	// IP-based rate limiting fields
	IPEnabled           bool     `yaml:"ip_enabled"`
	IPRequestsPerMinute int      `yaml:"ip_requests_per_minute"`
	IPSkipPaths         []string `yaml:"ip_skip_paths"`
	IPCustomHeaders     []string `yaml:"ip_custom_headers"`
}

// DiscoveryConfig holds MCP server discovery configuration
type DiscoveryConfig struct {
	Enabled          bool          `yaml:"enabled"`
	HealthInterval   time.Duration `yaml:"health_interval"`
	FailureThreshold int           `yaml:"failure_threshold"`
	RecoveryTimeout  time.Duration `yaml:"recovery_timeout"`
}

// GatewayConfig holds core gateway configuration
type GatewayConfig struct {
	LoadBalancer   string               `yaml:"load_balancer"`
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`
	ProxyTimeout   time.Duration        `yaml:"proxy_timeout"`
	MaxRetries     int                  `yaml:"max_retries"`
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	Enabled          bool          `yaml:"enabled"`
	FailureThreshold int           `yaml:"failure_threshold"`
	RecoveryTimeout  time.Duration `yaml:"recovery_timeout"`
	HalfOpenRequests int           `yaml:"half_open_requests"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `yaml:"host" env:"REDIS_HOST"`
	Password string `yaml:"password" env:"REDIS_PASSWORD"`
	Port     int    `yaml:"port" env:"REDIS_PORT"`
	Database int    `yaml:"database" env:"REDIS_DB"`
	PoolSize int    `yaml:"pool_size"`
}


// TransportConfig holds transport layer configuration
type TransportConfig struct {
	EnabledTransports  []types.TransportType `yaml:"enabled_transports" env:"TRANSPORT_ENABLED"`
	PathRewrite        PathRewriteConfig     `yaml:"path_rewrite"`
	SSEKeepAlive       time.Duration         `yaml:"sse_keep_alive"`
	WebSocketTimeout   time.Duration         `yaml:"websocket_timeout"`
	SessionTimeout     time.Duration         `yaml:"session_timeout"`
	MaxConnections     int                   `yaml:"max_connections"`
	BufferSize         int                   `yaml:"buffer_size"`
	STDIOTimeout       time.Duration         `yaml:"stdio_timeout"`
	StreamableStateful bool                  `yaml:"streamable_stateful"`
}

// PathRewriteConfig holds path rewriting configuration
type PathRewriteConfig struct {
	LogLevel string                  `yaml:"log_level"`
	Rules    []types.PathRewriteRule `yaml:"rules"`
	Enabled  bool                    `yaml:"enabled"`
}

// FiltersConfig holds content filtering configuration
type FiltersConfig struct {
	DefaultFilters map[string]interface{} `yaml:"default_filters"`
	GlobalSettings FilterGlobalSettings   `yaml:"global_settings"`
	Enabled        bool                   `yaml:"enabled" env:"FILTERS_ENABLED"`
	DatabaseDriven bool                   `yaml:"database_driven"`
}

// FilterGlobalSettings holds global filter settings
type FilterGlobalSettings struct {
	DefaultAction           string `yaml:"default_action"`
	MaxViolationsPerRequest int    `yaml:"max_violations_per_request"`
	LogAllViolations        bool   `yaml:"log_all_violations"`
	BlockOnHighSeverity     bool   `yaml:"block_on_high_severity"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	// Load from file
	config, err := loadFromFile(configPath)
	if err != nil {
		return nil, err
	}

	// Override with environment variables
	if err := overrideWithEnv(config); err != nil {
		return nil, err
	}

	return config, nil
}

// loadFromFile loads configuration from YAML file
func loadFromFile(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Expand environment variables
	expandedData := expandEnvVars(string(data))

	var config Config
	if err := yaml.Unmarshal([]byte(expandedData), &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// expandEnvVars expands environment variables in the format ${VAR:-default}
func expandEnvVars(input string) string {
	// Pattern matches ${VAR:-default} or ${VAR}
	envVarPattern := regexp.MustCompile(`\$\{([^}]+)\}`)

	return envVarPattern.ReplaceAllStringFunc(input, func(match string) string {
		// Remove ${ and }
		varExpr := match[2 : len(match)-1]

		// Check if it has a default value (VAR:-default)
		if strings.Contains(varExpr, ":-") {
			parts := strings.SplitN(varExpr, ":-", 2)
			envVar := parts[0]
			defaultValue := parts[1]

			if value := os.Getenv(envVar); value != "" {
				return value
			}
			return defaultValue
		}

		// Simple variable ${VAR}
		if value := os.Getenv(varExpr); value != "" {
			return value
		}

		// If no environment variable is set and no default, return empty string
		return ""
	})
}

// overrideWithEnv overrides configuration with environment variables
func overrideWithEnv(config *Config) error {
	// For now, just return nil - environment variable override can be implemented later
	return nil
}

// GetDSN returns database connection string
func (d *DatabaseConfig) GetDSN() string {
	// TODO: Implement DSN construction
	return ""
}

// SetDefaults sets default values for transport configuration
func (t *TransportConfig) SetDefaults() {
	if len(t.EnabledTransports) == 0 {
		t.EnabledTransports = []types.TransportType{
			types.TransportTypeHTTP,
			types.TransportTypeSSE,
			types.TransportTypeWebSocket,
			types.TransportTypeStreamable,
		}
	}

	if t.SSEKeepAlive == 0 {
		t.SSEKeepAlive = types.DefaultSSEKeepAlive
	}

	if t.WebSocketTimeout == 0 {
		t.WebSocketTimeout = types.DefaultWebSocketTimeout
	}

	if t.SessionTimeout == 0 {
		t.SessionTimeout = types.DefaultSessionTimeout
	}

	if t.MaxConnections == 0 {
		t.MaxConnections = types.DefaultMaxConnections
	}

	if t.BufferSize == 0 {
		t.BufferSize = types.DefaultBufferSize
	}

	if t.STDIOTimeout == 0 {
		t.STDIOTimeout = types.DefaultSTDIOTimeout
	}

	// Set path rewrite defaults
	if !t.PathRewrite.Enabled {
		t.PathRewrite.Enabled = true
		t.PathRewrite.LogLevel = "info"
	}
}

// Validate validates the transport configuration
func (t *TransportConfig) Validate() error {
	// Validate enabled transports
	for _, transportType := range t.EnabledTransports {
		switch transportType {
		case types.TransportTypeHTTP,
			types.TransportTypeSSE,
			types.TransportTypeWebSocket,
			types.TransportTypeStreamable,
			types.TransportTypeSTDIO:
			// Valid transport types
		default:
			return fmt.Errorf("invalid transport type: %s", transportType)
		}
	}

	// Validate timeouts
	if t.SSEKeepAlive <= 0 {
		return fmt.Errorf("sse_keep_alive must be positive")
	}

	if t.WebSocketTimeout <= 0 {
		return fmt.Errorf("websocket_timeout must be positive")
	}

	if t.SessionTimeout <= 0 {
		return fmt.Errorf("session_timeout must be positive")
	}

	if t.STDIOTimeout <= 0 {
		return fmt.Errorf("stdio_timeout must be positive")
	}

	// Validate connection limits
	if t.MaxConnections <= 0 {
		return fmt.Errorf("max_connections must be positive")
	}

	if t.BufferSize <= 0 {
		return fmt.Errorf("buffer_size must be positive")
	}

	return nil
}

// ToTransportConfig converts to types.TransportConfig
func (t *TransportConfig) ToTransportConfig() *types.TransportConfig {
	return &types.TransportConfig{
		EnabledTransports:  t.EnabledTransports,
		SSEKeepAlive:       t.SSEKeepAlive,
		WebSocketTimeout:   t.WebSocketTimeout,
		SessionTimeout:     t.SessionTimeout,
		MaxConnections:     t.MaxConnections,
		BufferSize:         t.BufferSize,
		StreamableStateful: t.StreamableStateful,
		STDIOTimeout:       t.STDIOTimeout,
	}
}
