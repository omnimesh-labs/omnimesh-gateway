package logging

import (
	"context"
	"fmt"
	"sync"
	"time"

	"mcp-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
)

// Service implements LogService interface with plugin-based storage
type Service struct {
	backend     StorageBackend
	config      *LoggingConfig
	subscribers map[string]LogSubscriber
	stopCh      chan struct{}
	level       LogLevel
	buffer      []*LogEntry
	wg          sync.WaitGroup
	mu          sync.RWMutex
	bufferMu    sync.Mutex
}

// NewService creates a new logging service with plugin-based storage
func NewService(config *LoggingConfig) (LogService, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Get plugin factory
	factory, err := GetPlugin(config.Backend)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin '%s': %w", config.Backend, err)
	}

	// Validate configuration
	if err := factory.ValidateConfig(config.Config); err != nil {
		return nil, fmt.Errorf("invalid plugin config: %w", err)
	}

	// Create backend instance
	backend := factory.Create()

	// Initialize service
	s := &Service{
		config:      config,
		backend:     backend,
		subscribers: make(map[string]LogSubscriber),
		level:       config.Level,
		buffer:      make([]*LogEntry, 0, config.BufferSize),
		stopCh:      make(chan struct{}),
	}

	// Initialize backend
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := backend.Initialize(ctx, config.Config); err != nil {
		return nil, fmt.Errorf("failed to initialize backend: %w", err)
	}

	// Start background flush if async mode is enabled
	if config.Async {
		s.startBackgroundFlush()
	}

	return s, nil
}

// Log writes a log entry
func (s *Service) Log(ctx context.Context, entry *LogEntry) error {
	if entry == nil {
		return fmt.Errorf("log entry cannot be nil")
	}

	// Check log level
	if entry.Level.Priority() < s.level.Priority() {
		return nil // Skip logs below current level
	}

	// Set defaults
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	if entry.Environment == "" {
		entry.Environment = s.config.Environment
	}

	// Notify subscribers
	s.notifySubscribers(entry)

	// Store the log
	if s.config.Async {
		return s.bufferLog(entry)
	}

	return s.backend.Store(ctx, entry)
}

// LogBatch writes multiple log entries
func (s *Service) LogBatch(ctx context.Context, entries []*LogEntry) error {
	if len(entries) == 0 {
		return nil
	}

	// Filter by log level and set defaults
	var filteredEntries []*LogEntry
	for _, entry := range entries {
		if entry.Level.Priority() >= s.level.Priority() {
			if entry.ID == "" {
				entry.ID = uuid.New().String()
			}
			if entry.Timestamp.IsZero() {
				entry.Timestamp = time.Now()
			}
			if entry.Environment == "" {
				entry.Environment = s.config.Environment
			}
			filteredEntries = append(filteredEntries, entry)
			s.notifySubscribers(entry)
		}
	}

	if len(filteredEntries) == 0 {
		return nil
	}

	// Store the logs
	if s.config.Async {
		return s.bufferLogs(filteredEntries)
	}

	return s.backend.StoreBatch(ctx, filteredEntries)
}

// Query searches for log entries
func (s *Service) Query(ctx context.Context, query *QueryRequest) ([]*LogEntry, error) {
	if query == nil {
		query = &QueryRequest{}
	}

	return s.backend.Query(ctx, query)
}

// Subscribe adds a real-time log subscriber
func (s *Service) Subscribe(subscriber LogSubscriber) error {
	if subscriber == nil {
		return fmt.Errorf("subscriber cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	id := subscriber.GetID()
	if id == "" {
		return fmt.Errorf("subscriber ID cannot be empty")
	}

	s.subscribers[id] = subscriber
	return nil
}

// Unsubscribe removes a log subscriber
func (s *Service) Unsubscribe(subscriberID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.subscribers, subscriberID)
	return nil
}

// SetLevel changes the minimum log level
func (s *Service) SetLevel(level LogLevel) error {
	if !level.IsValid() {
		return fmt.Errorf("invalid log level: %s", level)
	}

	s.mu.Lock()
	s.level = level
	s.mu.Unlock()

	// Log the level change
	entry := &LogEntry{
		Level:   LogLevelInfo,
		Message: fmt.Sprintf("Log level changed to %s", level),
		Logger:  "logging.service",
		Data:    map[string]interface{}{"new_level": string(level)},
	}

	return s.Log(context.Background(), entry)
}

// GetLevel returns the current log level
func (s *Service) GetLevel() LogLevel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.level
}

// HealthCheck verifies the service is operational
func (s *Service) HealthCheck(ctx context.Context) error {
	if s.backend == nil {
		return fmt.Errorf("backend not initialized")
	}

	return s.backend.HealthCheck(ctx)
}

// Close shuts down the service
func (s *Service) Close() error {
	// Stop background flush
	close(s.stopCh)
	s.wg.Wait()

	// Flush remaining buffer
	if len(s.buffer) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		s.flushBuffer(ctx)
	}

	// Close backend
	if s.backend != nil {
		return s.backend.Close()
	}

	return nil
}

// GetMetrics returns logging metrics
func (s *Service) GetMetrics() (*types.LoggingMetrics, error) {
	s.mu.RLock()
	bufferSize := len(s.buffer)
	subscriberCount := len(s.subscribers)
	s.mu.RUnlock()

	return &types.LoggingMetrics{
		BufferSize:      bufferSize,
		SubscriberCount: subscriberCount,
		CurrentLevel:    string(s.level),
		BackendType:     s.config.Backend,
		AsyncMode:       s.config.Async,
	}, nil
}

// Private helper methods

func (s *Service) notifySubscribers(entry *LogEntry) {
	s.mu.RLock()
	subscribers := make([]LogSubscriber, 0, len(s.subscribers))
	for _, sub := range s.subscribers {
		subscribers = append(subscribers, sub)
	}
	s.mu.RUnlock()

	// Notify subscribers in goroutines to avoid blocking
	for _, sub := range subscribers {
		go func(subscriber LogSubscriber) {
			if s.matchesFilter(entry, subscriber.GetFilters()) {
				subscriber.OnLog(entry)
			}
		}(sub)
	}
}

func (s *Service) matchesFilter(entry *LogEntry, filter *QueryRequest) bool {
	if filter == nil {
		return true
	}

	// Check level filter
	if filter.Level != "" && entry.Level != filter.Level {
		return false
	}

	// Check other filters
	if filter.Logger != "" && entry.Logger != filter.Logger {
		return false
	}
	if filter.EntityType != "" && entry.EntityType != filter.EntityType {
		return false
	}
	if filter.UserID != "" && entry.UserID != filter.UserID {
		return false
	}
	if filter.OrgID != "" && entry.OrgID != filter.OrgID {
		return false
	}

	return true
}

func (s *Service) bufferLog(entry *LogEntry) error {
	s.bufferMu.Lock()
	defer s.bufferMu.Unlock()

	s.buffer = append(s.buffer, entry)

	// Check if buffer is full
	if len(s.buffer) >= s.config.BatchSize {
		return s.flushBufferLocked(context.Background())
	}

	return nil
}

func (s *Service) bufferLogs(entries []*LogEntry) error {
	s.bufferMu.Lock()
	defer s.bufferMu.Unlock()

	s.buffer = append(s.buffer, entries...)

	// Check if buffer needs flushing
	if len(s.buffer) >= s.config.BatchSize {
		return s.flushBufferLocked(context.Background())
	}

	return nil
}

func (s *Service) startBackgroundFlush() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.config.FlushInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				s.flushBuffer(ctx)
				cancel()
			case <-s.stopCh:
				return
			}
		}
	}()
}

func (s *Service) flushBuffer(ctx context.Context) error {
	s.bufferMu.Lock()
	defer s.bufferMu.Unlock()
	return s.flushBufferLocked(ctx)
}

func (s *Service) flushBufferLocked(ctx context.Context) error {
	if len(s.buffer) == 0 {
		return nil
	}

	batch := make([]*LogEntry, len(s.buffer))
	copy(batch, s.buffer)
	s.buffer = s.buffer[:0] // Clear buffer

	return s.backend.StoreBatch(ctx, batch)
}

// LogRequest logs an HTTP request
func (s *Service) LogRequest(ctx context.Context, req *types.LogEntry) error {
	return s.Log(ctx, &LogEntry{
		Level:     LogLevelInfo,
		Message:   fmt.Sprintf("%s %s", req.Method, req.Path),
		Data:      req.Data,
		Timestamp: time.Now(),
		Logger:    "request",
	})
}

// LogAudit logs an audit event
func (s *Service) LogAudit(ctx context.Context, event *types.AuditLog) error {
	return s.Log(ctx, &LogEntry{
		Level:      LogLevelInfo,
		Message:    event.Action,
		Data:       event.Details,
		Timestamp:  time.Now(),
		Logger:     "audit",
		UserID:     event.UserID,
		OrgID:      event.OrganizationID,
		EntityType: event.Resource,
		EntityID:   event.ResourceID,
	})
}

// LogMetric logs a metric event
func (s *Service) LogMetric(ctx context.Context, metric *types.Metric) error {
	return s.Log(ctx, &LogEntry{
		Level:     LogLevelInfo,
		Message:   metric.Name,
		Data:      metric.Metadata,
		Timestamp: time.Now(),
		Logger:    "metric",
		OrgID:     metric.OrganizationID,
	})
}
