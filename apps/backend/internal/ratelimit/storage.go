package ratelimit

import (
	"encoding/json"
	"sync"
	"time"
)

// Storage interface defines rate limit storage operations
type Storage interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte, expiration time.Duration) error
	Increment(key string, expiration time.Duration) (int64, error)
	Delete(key string) error
	Cleanup() error
}

// MemoryStorage implements in-memory rate limit storage
type MemoryStorage struct {
	mu   sync.RWMutex
	data map[string]*storageItem
	stop chan struct{}
	done chan struct{}
}

type storageItem struct {
	Value      []byte
	Expiration time.Time
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage(cleanupInterval time.Duration) *MemoryStorage {
	storage := &MemoryStorage{
		data: make(map[string]*storageItem),
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}

	// Start cleanup goroutine
	go storage.cleanupLoop(cleanupInterval)

	return storage
}

// Get retrieves a value from memory storage
func (m *MemoryStorage) Get(key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.data[key]
	if !exists || (item.Expiration.Before(time.Now()) && !item.Expiration.IsZero()) {
		return nil, nil
	}

	return item.Value, nil
}

// Set stores a value in memory storage
func (m *MemoryStorage) Set(key string, value []byte, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var exp time.Time
	if expiration > 0 {
		exp = time.Now().Add(expiration)
	}

	m.data[key] = &storageItem{
		Value:      value,
		Expiration: exp,
	}

	return nil
}

// Increment increments a counter in memory storage
func (m *MemoryStorage) Increment(key string, expiration time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var count int64 = 1
	item, exists := m.data[key]

	if exists && (item.Expiration.IsZero() || item.Expiration.After(time.Now())) {
		// Unmarshal existing count
		if err := json.Unmarshal(item.Value, &count); err == nil {
			count++
		}
	}

	// Marshal new count
	value, _ := json.Marshal(count)

	var exp time.Time
	if expiration > 0 {
		exp = time.Now().Add(expiration)
	}

	m.data[key] = &storageItem{
		Value:      value,
		Expiration: exp,
	}

	return count, nil
}

// Delete removes a value from memory storage
func (m *MemoryStorage) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, key)
	return nil
}

// Cleanup removes expired items from memory storage
func (m *MemoryStorage) Cleanup() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for key, item := range m.data {
		if !item.Expiration.IsZero() && item.Expiration.Before(now) {
			delete(m.data, key)
		}
	}

	return nil
}

// Close shuts down the memory storage
func (m *MemoryStorage) Close() error {
	close(m.stop)
	<-m.done
	return nil
}

// cleanupLoop runs periodic cleanup
func (m *MemoryStorage) cleanupLoop(interval time.Duration) {
	defer close(m.done)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.Cleanup()
		case <-m.stop:
			return
		}
	}
}

// RedisStorage implements Redis-based rate limit storage
type RedisStorage struct {
	// TODO: Implement Redis storage
	// Use go-redis client
	// Implement all Storage interface methods
}

// NewRedisStorage creates a new Redis storage
func NewRedisStorage(addr, password string, db int) *RedisStorage {
	// TODO: Implement Redis storage initialization
	return &RedisStorage{}
}

// Get retrieves a value from Redis storage
func (r *RedisStorage) Get(key string) ([]byte, error) {
	// TODO: Implement Redis Get
	return nil, nil
}

// Set stores a value in Redis storage
func (r *RedisStorage) Set(key string, value []byte, expiration time.Duration) error {
	// TODO: Implement Redis Set
	return nil
}

// Increment increments a counter in Redis storage
func (r *RedisStorage) Increment(key string, expiration time.Duration) (int64, error) {
	// TODO: Implement Redis Increment
	return 0, nil
}

// Delete removes a value from Redis storage
func (r *RedisStorage) Delete(key string) error {
	// TODO: Implement Redis Delete
	return nil
}

// Cleanup removes expired items from Redis storage
func (r *RedisStorage) Cleanup() error {
	// TODO: Implement Redis Cleanup (may not be needed as Redis handles expiration)
	return nil
}
