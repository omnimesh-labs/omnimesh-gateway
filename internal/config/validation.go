package config

import (
	"errors"
	"fmt"
	"time"
)

// Validate validates the entire configuration
func (c *Config) Validate() error {
	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("server config: %w", err)
	}

	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("database config: %w", err)
	}

	if err := c.Auth.Validate(); err != nil {
		return fmt.Errorf("auth config: %w", err)
	}

	if err := c.Logging.Validate(); err != nil {
		return fmt.Errorf("logging config: %w", err)
	}

	if err := c.RateLimit.Validate(); err != nil {
		return fmt.Errorf("rate limit config: %w", err)
	}

	if err := c.Discovery.Validate(); err != nil {
		return fmt.Errorf("discovery config: %w", err)
	}

	if err := c.Gateway.Validate(); err != nil {
		return fmt.Errorf("gateway config: %w", err)
	}

	return nil
}

// Validate validates server configuration
func (s *ServerConfig) Validate() error {
	if s.Port <= 0 || s.Port > 65535 {
		return errors.New("invalid port number")
	}

	if s.ReadTimeout < 0 {
		return errors.New("read timeout cannot be negative")
	}

	if s.WriteTimeout < 0 {
		return errors.New("write timeout cannot be negative")
	}

	if s.IdleTimeout < 0 {
		return errors.New("idle timeout cannot be negative")
	}

	return s.TLS.Validate()
}

// Validate validates TLS configuration
func (t *TLSConfig) Validate() error {
	if !t.Enabled {
		return nil
	}

	if t.CertFile == "" {
		return errors.New("TLS cert file is required when TLS is enabled")
	}

	if t.KeyFile == "" {
		return errors.New("TLS key file is required when TLS is enabled")
	}

	return nil
}

// Validate validates database configuration
func (d *DatabaseConfig) Validate() error {
	if d.Host == "" {
		return errors.New("database host is required")
	}

	if d.Port <= 0 || d.Port > 65535 {
		return errors.New("invalid database port")
	}

	if d.User == "" {
		return errors.New("database user is required")
	}

	if d.Database == "" {
		return errors.New("database name is required")
	}

	if d.MaxOpenConns < 0 {
		return errors.New("max open connections cannot be negative")
	}

	if d.MaxIdleConns < 0 {
		return errors.New("max idle connections cannot be negative")
	}

	if d.MaxLifetime < 0 {
		return errors.New("max connection lifetime cannot be negative")
	}

	return nil
}

// Validate validates auth configuration
func (a *AuthConfig) Validate() error {
	if a.JWTSecret == "" {
		return errors.New("JWT secret is required")
	}

	if len(a.JWTSecret) < 32 {
		return errors.New("JWT secret must be at least 32 characters")
	}

	if a.AccessTokenExpiry <= 0 {
		return errors.New("access token expiry must be positive")
	}

	if a.RefreshTokenExpiry <= 0 {
		return errors.New("refresh token expiry must be positive")
	}

	if a.BCryptCost < 4 || a.BCryptCost > 31 {
		return errors.New("bcrypt cost must be between 4 and 31")
	}

	return nil
}

// Validate validates logging configuration
func (l *LoggingConfig) Validate() error {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}

	if !validLevels[l.Level] {
		return errors.New("invalid log level")
	}

	validFormats := map[string]bool{
		"json": true,
		"text": true,
	}

	if !validFormats[l.Format] {
		return errors.New("invalid log format")
	}

	if l.RetentionDays < 0 {
		return errors.New("retention days cannot be negative")
	}

	return nil
}

// Validate validates rate limit configuration
func (r *RateLimitConfig) Validate() error {
	if r.DefaultLimit < 0 {
		return errors.New("default limit cannot be negative")
	}

	if r.DefaultWindow <= 0 {
		return errors.New("default window must be positive")
	}

	validStorage := map[string]bool{
		"memory": true,
		"redis":  true,
	}

	if !validStorage[r.Storage] {
		return errors.New("invalid rate limit storage type")
	}

	if r.CleanupInterval <= 0 {
		return errors.New("cleanup interval must be positive")
	}

	return nil
}

// Validate validates discovery configuration
func (d *DiscoveryConfig) Validate() error {
	if d.HealthInterval <= 0 {
		return errors.New("health interval must be positive")
	}

	if d.FailureThreshold <= 0 {
		return errors.New("failure threshold must be positive")
	}

	if d.RecoveryTimeout <= 0 {
		return errors.New("recovery timeout must be positive")
	}

	return nil
}

// Validate validates gateway configuration
func (g *GatewayConfig) Validate() error {
	if g.ProxyTimeout <= 0 {
		return errors.New("proxy timeout must be positive")
	}

	if g.MaxRetries < 0 {
		return errors.New("max retries cannot be negative")
	}

	validBalancers := map[string]bool{
		"round_robin": true,
		"least_conn":  true,
		"weighted":    true,
	}

	if !validBalancers[g.LoadBalancer] {
		return errors.New("invalid load balancer type")
	}

	return g.CircuitBreaker.Validate()
}

// Validate validates circuit breaker configuration
func (c *CircuitBreakerConfig) Validate() error {
	if !c.Enabled {
		return nil
	}

	if c.FailureThreshold <= 0 {
		return errors.New("failure threshold must be positive")
	}

	if c.RecoveryTimeout <= 0 {
		return errors.New("recovery timeout must be positive")
	}

	if c.HalfOpenRequests <= 0 {
		return errors.New("half open requests must be positive")
	}

	return nil
}

// SetDefaults sets default values for configuration
func (c *Config) SetDefaults() {
	// Server defaults
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = 30 * time.Second
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = 30 * time.Second
	}
	if c.Server.IdleTimeout == 0 {
		c.Server.IdleTimeout = 120 * time.Second
	}

	// Database defaults
	if c.Database.Port == 0 {
		c.Database.Port = 5432
	}
	if c.Database.SSLMode == "" {
		c.Database.SSLMode = "require"
	}
	if c.Database.MaxOpenConns == 0 {
		c.Database.MaxOpenConns = 25
	}
	if c.Database.MaxIdleConns == 0 {
		c.Database.MaxIdleConns = 5
	}
	if c.Database.MaxLifetime == 0 {
		c.Database.MaxLifetime = 5 * time.Minute
	}

	// Auth defaults
	if c.Auth.AccessTokenExpiry == 0 {
		c.Auth.AccessTokenExpiry = 15 * time.Minute
	}
	if c.Auth.RefreshTokenExpiry == 0 {
		c.Auth.RefreshTokenExpiry = 24 * time.Hour
	}
	if c.Auth.BCryptCost == 0 {
		c.Auth.BCryptCost = 12
	}

	// Logging defaults
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "json"
	}
	if c.Logging.RetentionDays == 0 {
		c.Logging.RetentionDays = 30
	}

	// Rate limit defaults
	if c.RateLimit.DefaultLimit == 0 {
		c.RateLimit.DefaultLimit = 1000
	}
	if c.RateLimit.DefaultWindow == 0 {
		c.RateLimit.DefaultWindow = time.Hour
	}
	if c.RateLimit.Storage == "" {
		c.RateLimit.Storage = "memory"
	}
	if c.RateLimit.CleanupInterval == 0 {
		c.RateLimit.CleanupInterval = 5 * time.Minute
	}

	// Discovery defaults
	if c.Discovery.HealthInterval == 0 {
		c.Discovery.HealthInterval = 30 * time.Second
	}
	if c.Discovery.FailureThreshold == 0 {
		c.Discovery.FailureThreshold = 3
	}
	if c.Discovery.RecoveryTimeout == 0 {
		c.Discovery.RecoveryTimeout = 60 * time.Second
	}

	// Gateway defaults
	if c.Gateway.ProxyTimeout == 0 {
		c.Gateway.ProxyTimeout = 30 * time.Second
	}
	if c.Gateway.MaxRetries == 0 {
		c.Gateway.MaxRetries = 3
	}
	if c.Gateway.LoadBalancer == "" {
		c.Gateway.LoadBalancer = "round_robin"
	}
	if c.Gateway.CircuitBreaker.FailureThreshold == 0 {
		c.Gateway.CircuitBreaker.FailureThreshold = 5
	}
	if c.Gateway.CircuitBreaker.RecoveryTimeout == 0 {
		c.Gateway.CircuitBreaker.RecoveryTimeout = 60 * time.Second
	}
	if c.Gateway.CircuitBreaker.HalfOpenRequests == 0 {
		c.Gateway.CircuitBreaker.HalfOpenRequests = 3
	}

	// Redis defaults
	if c.Redis.Port == 0 {
		c.Redis.Port = 6379
	}
	if c.Redis.PoolSize == 0 {
		c.Redis.PoolSize = 10
	}
}
