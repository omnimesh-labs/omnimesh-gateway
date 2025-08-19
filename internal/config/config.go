package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	Auth      AuthConfig      `yaml:"auth"`
	Logging   LoggingConfig   `yaml:"logging"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	Discovery DiscoveryConfig `yaml:"discovery"`
	Gateway   GatewayConfig   `yaml:"gateway"`
	Redis     RedisConfig     `yaml:"redis"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string        `yaml:"host" env:"SERVER_HOST"`
	Port         int           `yaml:"port" env:"SERVER_PORT"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
	TLS          TLSConfig     `yaml:"tls"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled" env:"TLS_ENABLED"`
	CertFile string `yaml:"cert_file" env:"TLS_CERT_FILE"`
	KeyFile  string `yaml:"key_file" env:"TLS_KEY_FILE"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host         string        `yaml:"host" env:"DB_HOST"`
	Port         int           `yaml:"port" env:"DB_PORT"`
	User         string        `yaml:"user" env:"DB_USER"`
	Password     string        `yaml:"password" env:"DB_PASSWORD"`
	Database     string        `yaml:"database" env:"DB_NAME"`
	SSLMode      string        `yaml:"ssl_mode" env:"DB_SSL_MODE"`
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
	Level          string `yaml:"level" env:"LOG_LEVEL"`
	Format         string `yaml:"format"`
	RequestLogging bool   `yaml:"request_logging"`
	AuditLogging   bool   `yaml:"audit_logging"`
	MetricsEnabled bool   `yaml:"metrics_enabled"`
	RetentionDays  int    `yaml:"retention_days"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled         bool          `yaml:"enabled"`
	DefaultLimit    int           `yaml:"default_limit"`
	DefaultWindow   time.Duration `yaml:"default_window"`
	Storage         string        `yaml:"storage"` // "memory" or "redis"
	CleanupInterval time.Duration `yaml:"cleanup_interval"`
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
	ProxyTimeout   time.Duration        `yaml:"proxy_timeout"`
	MaxRetries     int                  `yaml:"max_retries"`
	LoadBalancer   string               `yaml:"load_balancer"` // "round_robin", "least_conn", "weighted"
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`
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
	Port     int    `yaml:"port" env:"REDIS_PORT"`
	Password string `yaml:"password" env:"REDIS_PASSWORD"`
	Database int    `yaml:"database" env:"REDIS_DB"`
	PoolSize int    `yaml:"pool_size"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	// TODO: Implement configuration loading
	// Read YAML file
	// Override with environment variables
	// Validate configuration
	return nil, nil
}

// loadFromFile loads configuration from YAML file
func loadFromFile(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// overrideWithEnv overrides configuration with environment variables
func overrideWithEnv(config *Config) error {
	// TODO: Implement environment variable override
	return nil
}

// GetDSN returns database connection string
func (d *DatabaseConfig) GetDSN() string {
	// TODO: Implement DSN construction
	return ""
}
