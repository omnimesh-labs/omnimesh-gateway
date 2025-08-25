package auth

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenCache defines the interface for JWT token caching/blacklisting
type TokenCache interface {
	// Set adds a token to the blacklist with expiration
	Set(ctx context.Context, token string, expiration time.Duration) error

	// IsBlacklisted checks if a token is blacklisted
	IsBlacklisted(ctx context.Context, token string) (bool, error)

	// Cleanup removes expired tokens (for memory cache)
	Cleanup(ctx context.Context) error

	// Close closes the cache connection
	Close() error
}

// RedisTokenCache implements TokenCache using Redis
type RedisTokenCache struct {
	client *redis.Client
	prefix string
}

// NewRedisTokenCache creates a new Redis-backed token cache
func NewRedisTokenCache(addr, password string, db int) (*RedisTokenCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisTokenCache{
		client: client,
		prefix: "jwt_blacklist:",
	}, nil
}

// Set adds a token to the blacklist with expiration
func (r *RedisTokenCache) Set(ctx context.Context, token string, expiration time.Duration) error {
	key := r.prefix + token
	return r.client.Set(ctx, key, "revoked", expiration).Err()
}

// IsBlacklisted checks if a token is blacklisted
func (r *RedisTokenCache) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	key := r.prefix + token
	result := r.client.Get(ctx, key)

	if result.Err() == redis.Nil {
		return false, nil // Key doesn't exist, not blacklisted
	}

	if result.Err() != nil {
		return false, result.Err()
	}

	return true, nil // Key exists, token is blacklisted
}

// Cleanup is a no-op for Redis (TTL handles expiration automatically)
func (r *RedisTokenCache) Cleanup(ctx context.Context) error {
	return nil // Redis handles TTL automatically
}

// Close closes the Redis connection
func (r *RedisTokenCache) Close() error {
	return r.client.Close()
}

// MemoryTokenCache implements TokenCache using in-memory storage
type MemoryTokenCache struct {
	tokens map[string]time.Time
	mu     sync.RWMutex
}

// NewMemoryTokenCache creates a new memory-backed token cache
func NewMemoryTokenCache() *MemoryTokenCache {
	return &MemoryTokenCache{
		tokens: make(map[string]time.Time),
	}
}

// Set adds a token to the blacklist with expiration
func (m *MemoryTokenCache) Set(ctx context.Context, token string, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	expiryTime := time.Now().Add(expiration)
	m.tokens[token] = expiryTime

	return nil
}

// IsBlacklisted checks if a token is blacklisted
func (m *MemoryTokenCache) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	expiryTime, exists := m.tokens[token]
	if !exists {
		return false, nil
	}

	// Check if token has expired
	if time.Now().After(expiryTime) {
		// Clean up expired token (we can do this safely since we have a lock)
		m.mu.RUnlock()
		m.mu.Lock()
		delete(m.tokens, token)
		m.mu.Unlock()
		m.mu.RLock()
		return false, nil
	}

	return true, nil
}

// Cleanup removes expired tokens from memory
func (m *MemoryTokenCache) Cleanup(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for token, expiryTime := range m.tokens {
		if now.After(expiryTime) {
			delete(m.tokens, token)
		}
	}

	return nil
}

// Close is a no-op for memory cache
func (m *MemoryTokenCache) Close() error {
	return nil
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	RedisAddr     string `yaml:"redis_addr"`
	RedisPassword string `yaml:"redis_password"`
	RedisDB       int    `yaml:"redis_db"`
	UseRedis      bool   `yaml:"use_redis"`
}

// NewTokenCache creates a new token cache based on configuration
func NewTokenCache(config CacheConfig) (TokenCache, error) {
	if config.UseRedis {
		return NewRedisTokenCache(config.RedisAddr, config.RedisPassword, config.RedisDB)
	}

	return NewMemoryTokenCache(), nil
}
