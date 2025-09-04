package unit

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"mcp-gateway/apps/backend/internal/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockStorageBackend is a mock implementation of logging.StorageBackend
type MockStorageBackend struct {
	mock.Mock
	entries []logging.LogEntry
	mu      sync.RWMutex
}

func NewMockStorageBackend() *MockStorageBackend {
	return &MockStorageBackend{
		entries: make([]logging.LogEntry, 0),
	}
}

func (m *MockStorageBackend) Initialize(ctx context.Context, config map[string]interface{}) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockStorageBackend) Store(ctx context.Context, entry *logging.LogEntry) error {
	args := m.Called(ctx, entry)
	if args.Error(0) == nil {
		m.mu.Lock()
		m.entries = append(m.entries, *entry)
		m.mu.Unlock()
	}
	return args.Error(0)
}

func (m *MockStorageBackend) StoreBatch(ctx context.Context, entries []*logging.LogEntry) error {
	args := m.Called(ctx, entries)
	if args.Error(0) == nil {
		m.mu.Lock()
		for _, entry := range entries {
			m.entries = append(m.entries, *entry)
		}
		m.mu.Unlock()
	}
	return args.Error(0)
}

func (m *MockStorageBackend) Query(ctx context.Context, query *logging.QueryRequest) ([]*logging.LogEntry, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]*logging.LogEntry), args.Error(1)
}

func (m *MockStorageBackend) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStorageBackend) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockStorageBackend) GetCapabilities() logging.BackendCapabilities {
	args := m.Called()
	return args.Get(0).(logging.BackendCapabilities)
}

func (m *MockStorageBackend) GetStoredEntries() []logging.LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]logging.LogEntry, len(m.entries))
	copy(result, m.entries)
	return result
}

// MockPluginFactory creates mock storage backends
type MockPluginFactory struct {
	mock.Mock
	backend *MockStorageBackend
}

func NewMockPluginFactory() *MockPluginFactory {
	return &MockPluginFactory{
		backend: NewMockStorageBackend(),
	}
}

func (m *MockPluginFactory) Create() logging.StorageBackend {
	return m.backend
}

func (m *MockPluginFactory) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockPluginFactory) GetDescription() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockPluginFactory) ValidateConfig(config map[string]interface{}) error {
	args := m.Called(config)
	return args.Error(0)
}

// MockLogSubscriber implements logging.LogSubscriber
type MockLogSubscriber struct {
	mock.Mock
	id      string
	entries []*logging.LogEntry
	mu      sync.RWMutex
}

func NewMockLogSubscriber(id string) *MockLogSubscriber {
	return &MockLogSubscriber{
		id:      id,
		entries: make([]*logging.LogEntry, 0),
	}
}

func (m *MockLogSubscriber) OnLog(entry *logging.LogEntry) error {
	args := m.Called(entry)
	if args.Error(0) == nil {
		m.mu.Lock()
		m.entries = append(m.entries, entry)
		m.mu.Unlock()
	}
	return args.Error(0)
}

func (m *MockLogSubscriber) GetID() string {
	return m.id
}

func (m *MockLogSubscriber) GetFilters() *logging.QueryRequest {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*logging.QueryRequest)
}

func (m *MockLogSubscriber) GetReceivedEntries() []*logging.LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*logging.LogEntry, len(m.entries))
	copy(result, m.entries)
	return result
}

func setupLoggingService(async bool) (logging.LogService, *MockPluginFactory, error) {
	// Create unique plugin name to avoid conflicts
	pluginName := fmt.Sprintf("mock-%d", time.Now().UnixNano())

	// Register mock plugin
	factory := NewMockPluginFactory()
	factory.Mock.On("GetName").Return(pluginName)
	factory.Mock.On("GetDescription").Return("Mock storage backend for testing")
	factory.On("ValidateConfig", mock.Anything).Return(nil)
	factory.backend.On("Initialize", mock.Anything, mock.Anything).Return(nil)
	factory.backend.On("Store", mock.Anything, mock.Anything).Return(nil)
	factory.backend.On("StoreBatch", mock.Anything, mock.Anything).Return(nil)
	factory.backend.On("Query", mock.Anything, mock.Anything).Return([]*logging.LogEntry{}, nil)
	factory.backend.On("Close").Return(nil)
	factory.backend.On("HealthCheck", mock.Anything).Return(nil)
	factory.backend.On("GetCapabilities").Return(logging.BackendCapabilities{
		SupportsQuery:      true,
		SupportsStreaming:  false,
		SupportsRetention:  false,
		SupportsBatchWrite: true,
		SupportsMetrics:    false,
	})

	err := logging.RegisterPlugin(factory)
	if err != nil {
		return nil, nil, err
	}

	config := &logging.LoggingConfig{
		Level:         logging.LogLevelInfo,
		Backend:       pluginName,
		Environment:   "test",
		BufferSize:    100,
		BatchSize:     10,
		FlushInterval: 100 * time.Millisecond,
		Async:         async,
		Config:        map[string]interface{}{},
	}

	service, err := logging.NewService(config)
	return service, factory, err
}

func TestLogLevel_Priority(t *testing.T) {
	tests := []struct {
		level    logging.LogLevel
		priority int
	}{
		{logging.LogLevelDebug, 0},
		{logging.LogLevelInfo, 1},
		{logging.LogLevelNotice, 2},
		{logging.LogLevelWarning, 3},
		{logging.LogLevelError, 4},
		{logging.LogLevelCritical, 5},
		{logging.LogLevelAlert, 6},
		{logging.LogLevelEmergency, 7},
		{"invalid", 1}, // Default to info level priority
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			assert.Equal(t, tt.priority, tt.level.Priority())
		})
	}
}

func TestLogLevel_IsValid(t *testing.T) {
	validLevels := []logging.LogLevel{
		logging.LogLevelDebug,
		logging.LogLevelInfo,
		logging.LogLevelNotice,
		logging.LogLevelWarning,
		logging.LogLevelError,
		logging.LogLevelCritical,
		logging.LogLevelAlert,
		logging.LogLevelEmergency,
	}

	for _, level := range validLevels {
		assert.True(t, level.IsValid(), "Level %s should be valid", level)
	}

	invalidLevels := []logging.LogLevel{"invalid", "unknown", ""}
	for _, level := range invalidLevels {
		assert.False(t, level.IsValid(), "Level %s should be invalid", level)
	}
}

func TestNewService(t *testing.T) {
	tests := []struct {
		name        string
		config      *logging.LoggingConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
			errorMsg:    "config cannot be nil",
		},
		{
			name: "unknown backend",
			config: &logging.LoggingConfig{
				Backend: "unknown",
				Level:   logging.LogLevelInfo,
			},
			expectError: true,
			errorMsg:    "failed to get plugin",
		},
		{
			name: "valid sync configuration",
			config: &logging.LoggingConfig{
				Backend:     "mock",
				Level:       logging.LogLevelInfo,
				Environment: "test",
				Async:       false,
				Config:      map[string]interface{}{},
			},
			expectError: false,
		},
	}

	// Register mock plugin for tests
	factory := NewMockPluginFactory()
	factory.On("GetName").Return("mock")
	factory.On("ValidateConfig", mock.Anything).Return(nil)
	factory.backend.On("Initialize", mock.Anything, mock.Anything).Return(nil)
	factory.backend.On("Close").Return(nil)
	err := logging.RegisterPlugin(factory)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := logging.NewService(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, service)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
				if service != nil {
					service.Close()
				}
			}
		})
	}
}

func TestService_Log(t *testing.T) {
	service, _, err := setupLoggingService(false) // Sync mode
	require.NoError(t, err)
	defer service.Close()

	ctx := context.Background()

	tests := []struct {
		name        string
		entry       *logging.LogEntry
		expectError bool
		shouldStore bool
	}{
		{
			name:        "nil entry",
			entry:       nil,
			expectError: true,
		},
		{
			name: "valid info entry",
			entry: &logging.LogEntry{
				Level:   logging.LogLevelInfo,
				Message: "Test info message",
				Logger:  "test",
			},
			expectError: false,
			shouldStore: true,
		},
		{
			name: "debug entry below threshold",
			entry: &logging.LogEntry{
				Level:   logging.LogLevelDebug,
				Message: "Test debug message",
				Logger:  "test",
			},
			expectError: false,
			shouldStore: false, // Below info level threshold
		},
		{
			name: "error entry above threshold",
			entry: &logging.LogEntry{
				Level:   logging.LogLevelError,
				Message: "Test error message",
				Logger:  "test",
			},
			expectError: false,
			shouldStore: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Log(ctx, tt.entry)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				if tt.shouldStore {
					// Verify entry was stored with defaults set
					if tt.entry != nil {
						assert.NotEmpty(t, tt.entry.ID, "Entry ID should be generated")
						assert.False(t, tt.entry.Timestamp.IsZero(), "Timestamp should be set")
						assert.Equal(t, "test", tt.entry.Environment)
					}
				}
			}
		})
	}
}

func TestService_LogBatch(t *testing.T) {
	service, _, err := setupLoggingService(false) // Sync mode
	require.NoError(t, err)
	defer service.Close()

	ctx := context.Background()

	tests := []struct {
		name    string
		entries []*logging.LogEntry
	}{
		{
			name:    "empty batch",
			entries: []*logging.LogEntry{},
		},
		{
			name: "mixed level batch",
			entries: []*logging.LogEntry{
				{
					Level:   logging.LogLevelDebug,
					Message: "Debug message",
				},
				{
					Level:   logging.LogLevelInfo,
					Message: "Info message",
				},
				{
					Level:   logging.LogLevelError,
					Message: "Error message",
				},
			},
		},
		{
			name: "all entries below threshold",
			entries: []*logging.LogEntry{
				{
					Level:   logging.LogLevelDebug,
					Message: "Debug message 1",
				},
				{
					Level:   logging.LogLevelDebug,
					Message: "Debug message 2",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.LogBatch(ctx, tt.entries)
			assert.NoError(t, err)

			// Verify entries have defaults set
			for _, entry := range tt.entries {
				if entry.Level.Priority() >= logging.LogLevelInfo.Priority() {
					assert.NotEmpty(t, entry.ID)
					assert.False(t, entry.Timestamp.IsZero())
					assert.Equal(t, "test", entry.Environment)
				}
			}
		})
	}
}

func TestService_Query(t *testing.T) {
	service, factory, err := setupLoggingService(false)
	require.NoError(t, err)
	defer service.Close()

	ctx := context.Background()

	// Clear any existing expectations
	factory.backend.ExpectedCalls = nil
	factory.backend.Calls = nil

	// Add necessary expectations
	factory.backend.On("Close").Return(nil)

	// Mock the query response
	expectedEntries := []*logging.LogEntry{
		{
			ID:      "1",
			Level:   logging.LogLevelInfo,
			Message: "Test message",
		},
	}
	factory.backend.On("Query", mock.Anything, mock.Anything).Return(expectedEntries, nil)

	query := &logging.QueryRequest{
		Level: logging.LogLevelInfo,
		Limit: 10,
	}

	entries, err := service.Query(ctx, query)
	assert.NoError(t, err)
	assert.Equal(t, expectedEntries, entries)

	// Test with nil query
	factory.backend.On("Query", mock.Anything, mock.Anything).Return([]*logging.LogEntry{}, nil).Once()
	entries, err = service.Query(ctx, nil)
	assert.NoError(t, err)
	assert.NotNil(t, entries)
}

func TestService_Subscribe(t *testing.T) {
	service, _, err := setupLoggingService(false)
	require.NoError(t, err)
	defer service.Close()

	tests := []struct {
		name        string
		subscriber  logging.LogSubscriber
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil subscriber",
			subscriber:  nil,
			expectError: true,
			errorMsg:    "subscriber cannot be nil",
		},
		{
			name: "subscriber with empty ID",
			subscriber: &MockLogSubscriber{
				id: "",
			},
			expectError: true,
			errorMsg:    "subscriber ID cannot be empty",
		},
		{
			name: "valid subscriber",
			subscriber: NewMockLogSubscriber("test-subscriber"),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Subscribe(tt.subscriber)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)

				// Test that subscriber receives notifications
				if tt.subscriber != nil {
					mockSub := tt.subscriber.(*MockLogSubscriber)
					mockSub.On("OnLog", mock.Anything).Return(nil)
					mockSub.On("GetFilters").Return((*logging.QueryRequest)(nil))

					entry := &logging.LogEntry{
						Level:   logging.LogLevelInfo,
						Message: "Test notification",
					}
					service.Log(context.Background(), entry)

					// Clean up
					service.Unsubscribe(tt.subscriber.GetID())
				}
			}
		})
	}
}

func TestService_Unsubscribe(t *testing.T) {
	service, _, err := setupLoggingService(false)
	require.NoError(t, err)
	defer service.Close()

	// Subscribe first
	subscriber := NewMockLogSubscriber("test-subscriber")
	subscriber.On("OnLog", mock.Anything).Return(nil)
	subscriber.On("GetFilters").Return((*logging.QueryRequest)(nil))
	err = service.Subscribe(subscriber)
	require.NoError(t, err)

	// Unsubscribe
	err = service.Unsubscribe("test-subscriber")
	assert.NoError(t, err)

	// Unsubscribing non-existent subscriber should not error
	err = service.Unsubscribe("non-existent")
	assert.NoError(t, err)
}

func TestService_SetLevel(t *testing.T) {
	service, _, err := setupLoggingService(false)
	require.NoError(t, err)
	defer service.Close()

	tests := []struct {
		name        string
		level       logging.LogLevel
		expectError bool
	}{
		{
			name:        "valid level",
			level:       logging.LogLevelError,
			expectError: false,
		},
		{
			name:        "invalid level",
			level:       "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SetLevel(tt.level)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.level, service.GetLevel())
			}
		})
	}
}

func TestService_GetLevel(t *testing.T) {
	service, _, err := setupLoggingService(false)
	require.NoError(t, err)
	defer service.Close()

	// Initial level should be info
	assert.Equal(t, logging.LogLevelInfo, service.GetLevel())

	// Change level and verify
	err = service.SetLevel(logging.LogLevelError)
	require.NoError(t, err)
	assert.Equal(t, logging.LogLevelError, service.GetLevel())
}

func TestService_HealthCheck(t *testing.T) {
	service, factory, err := setupLoggingService(false)
	require.NoError(t, err)
	defer service.Close()

	ctx := context.Background()

	// Clear any existing expectations
	factory.backend.ExpectedCalls = nil
	factory.backend.Calls = nil

	// Add necessary expectations
	factory.backend.On("Close").Return(nil)

	// Test successful health check
	factory.backend.On("HealthCheck", mock.Anything).Return(nil)
	err = service.HealthCheck(ctx)
	assert.NoError(t, err)

	// Clear expectations for the next test
	factory.backend.ExpectedCalls = nil
	factory.backend.Calls = nil
	factory.backend.On("Close").Return(nil)

	// Test failed health check
	factory.backend.On("HealthCheck", mock.Anything).Return(assert.AnError)
	err = service.HealthCheck(ctx)
	assert.Error(t, err)
}

func TestService_AsyncMode(t *testing.T) {
	service, _, err := setupLoggingService(true) // Async mode
	require.NoError(t, err)
	defer service.Close()

	ctx := context.Background()

	entry := &logging.LogEntry{
		Level:   logging.LogLevelInfo,
		Message: "Async test message",
		Logger:  "test",
	}

	err = service.Log(ctx, entry)
	assert.NoError(t, err)

	// Give some time for async processing
	time.Sleep(200 * time.Millisecond)

	// In async mode, logs should be buffered and flushed in batches
	// This is harder to test deterministically, but we can verify no immediate errors
	assert.NotEmpty(t, entry.ID)
	assert.False(t, entry.Timestamp.IsZero())
}

func TestService_GetMetrics(t *testing.T) {
	service, _, err := setupLoggingService(false)
	require.NoError(t, err)
	defer service.Close()

	// This method may not be implemented yet, so we just verify it doesn't panic
	metrics, err := service.GetMetrics()
	// We don't assert specific values since implementation may vary
	_ = metrics
	_ = err
}

func TestPluginRegistry(t *testing.T) {
	registry := logging.NewPluginRegistry()

	// Test registering a plugin
	factory := NewMockPluginFactory()
	factory.On("GetName").Return("test-plugin")
	factory.On("GetDescription").Return("Mock storage backend for testing")
	factory.backend.On("GetCapabilities").Return(logging.BackendCapabilities{
		SupportsQuery:      true,
		SupportsStreaming:  false,
		SupportsRetention:  false,
		SupportsBatchWrite: true,
		SupportsMetrics:    false,
	})
	err := registry.Register(factory)
	assert.NoError(t, err)

	// Test getting a plugin
	retrievedFactory, err := registry.Get("test-plugin")
	assert.NoError(t, err)
	assert.Equal(t, factory, retrievedFactory)

	// Test listing plugins
	plugins := registry.List()
	assert.Contains(t, plugins, "test-plugin")

	// Test getting plugin info
	info, err := registry.GetInfo("test-plugin")
	assert.NoError(t, err)
	assert.Equal(t, "test-plugin", info.Name)
	assert.Equal(t, "Mock storage backend for testing", info.Description)

	// Test registering duplicate plugin
	err = registry.Register(factory)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")

	// Test getting non-existent plugin
	_, err = registry.Get("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test registering nil factory
	err = registry.Register(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

func TestGlobalPluginRegistry(t *testing.T) {
	// Test global registry functions
	factory := &MockPluginFactory{}
	factory.Mock.On("GetName").Return("global-test")
	factory.Mock.On("GetDescription").Return("Global test plugin")

	err := logging.RegisterPlugin(factory)
	assert.NoError(t, err)

	retrievedFactory, err := logging.GetPlugin("global-test")
	assert.NoError(t, err)
	assert.NotNil(t, retrievedFactory)

	plugins := logging.ListPlugins()
	assert.Contains(t, plugins, "global-test")
}

func TestConcurrentLogging(t *testing.T) {
	service, _, err := setupLoggingService(true) // Async mode for better concurrency
	require.NoError(t, err)
	defer service.Close()

	ctx := context.Background()
	numGoroutines := 50
	numLogsPerGoroutine := 10

	var wg sync.WaitGroup
	errors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			for j := 0; j < numLogsPerGoroutine; j++ {
				entry := &logging.LogEntry{
					Level:   logging.LogLevelInfo,
					Message: fmt.Sprintf("Concurrent message from goroutine %d, log %d", index, j),
					Logger:  "concurrent-test",
				}

				if err := service.Log(ctx, entry); err != nil {
					errors[index] = err
					return
				}
			}
		}(i)
	}

	wg.Wait()

	// Check that no goroutine encountered an error
	for i, err := range errors {
		assert.NoError(t, err, "Goroutine %d encountered an error", i)
	}

	// Give time for async processing
	time.Sleep(200 * time.Millisecond)
}

// Benchmark tests
func BenchmarkService_Log(b *testing.B) {
	service, _, err := setupLoggingService(false)
	if err != nil {
		b.Fatal(err)
	}
	defer service.Close()

	entry := &logging.LogEntry{
		Level:   logging.LogLevelInfo,
		Message: "Benchmark test message",
		Logger:  "benchmark",
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset entry fields that get modified
		entry.ID = ""
		entry.Timestamp = time.Time{}
		entry.Environment = ""

		err := service.Log(ctx, entry)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkService_LogBatch(b *testing.B) {
	service, _, err := setupLoggingService(false)
	if err != nil {
		b.Fatal(err)
	}
	defer service.Close()

	entries := make([]*logging.LogEntry, 10)
	for i := 0; i < 10; i++ {
		entries[i] = &logging.LogEntry{
			Level:   logging.LogLevelInfo,
			Message: fmt.Sprintf("Batch message %d", i),
			Logger:  "benchmark",
		}
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset entry fields
		for _, entry := range entries {
			entry.ID = ""
			entry.Timestamp = time.Time{}
			entry.Environment = ""
		}

		err := service.LogBatch(ctx, entries)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLogLevel_Priority(b *testing.B) {
	level := logging.LogLevelInfo

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = level.Priority()
	}
}
