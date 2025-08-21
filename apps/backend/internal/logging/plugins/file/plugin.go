package file

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"mcp-gateway/apps/backend/internal/logging"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// FileStorageBackend implements types.StorageBackend for file storage
type FileStorageBackend struct {
	mu          sync.RWMutex
	config      *FileConfig
	currentFile *os.File
	currentSize int64
	initialized bool
}

// FileConfig holds configuration for file storage
type FileConfig struct {
	Path         string `json:"path" yaml:"path"`
	MaxSize      int64  `json:"max_size" yaml:"max_size"`
	MaxFiles     int    `json:"max_files" yaml:"max_files"`
	Rotate       bool   `json:"rotate" yaml:"rotate"`
	Permissions  string `json:"permissions" yaml:"permissions"`
	BufferSize   int    `json:"buffer_size" yaml:"buffer_size"`
	SyncInterval string `json:"sync_interval" yaml:"sync_interval"`
}

// DefaultFileConfig returns default configuration for file storage
func DefaultFileConfig() *FileConfig {
	return &FileConfig{
		Path:         "logs/mcp-gateway.log",
		MaxSize:      100 * 1024 * 1024, // 100MB
		MaxFiles:     10,
		Rotate:       true,
		Permissions:  "0644",
		BufferSize:   8192,
		SyncInterval: "5s",
	}
}

// Initialize sets up the file storage backend
func (f *FileStorageBackend) Initialize(ctx context.Context, config map[string]interface{}) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Parse configuration
	f.config = DefaultFileConfig()
	if err := f.parseConfig(config); err != nil {
		return fmt.Errorf("failed to parse file config: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(f.config.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	if err := f.openLogFile(); err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	f.initialized = true
	return nil
}

// Store saves a log entry to file
func (f *FileStorageBackend) Store(ctx context.Context, entry *logging.LogEntry) error {
	if !f.initialized {
		return fmt.Errorf("file storage not initialized")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// Serialize log entry
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Add newline
	data = append(data, '\n')

	// Check if rotation is needed
	if f.config.Rotate && f.currentSize+int64(len(data)) > f.config.MaxSize {
		if err := f.rotateFile(); err != nil {
			return fmt.Errorf("failed to rotate file: %w", err)
		}
	}

	// Write to file
	n, err := f.currentFile.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write log entry: %w", err)
	}

	f.currentSize += int64(n)
	return nil
}

// StoreBatch saves multiple log entries efficiently
func (f *FileStorageBackend) StoreBatch(ctx context.Context, entries []*logging.LogEntry) error {
	if !f.initialized {
		return fmt.Errorf("file storage not initialized")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	var totalSize int64
	var batch []byte

	// Pre-serialize all entries
	for _, entry := range entries {
		data, err := json.Marshal(entry)
		if err != nil {
			return fmt.Errorf("failed to marshal log entry: %w", err)
		}
		data = append(data, '\n')
		batch = append(batch, data...)
		totalSize += int64(len(data))
	}

	// Check if rotation is needed
	if f.config.Rotate && f.currentSize+totalSize > f.config.MaxSize {
		if err := f.rotateFile(); err != nil {
			return fmt.Errorf("failed to rotate file: %w", err)
		}
	}

	// Write batch
	n, err := f.currentFile.Write(batch)
	if err != nil {
		return fmt.Errorf("failed to write log batch: %w", err)
	}

	f.currentSize += int64(n)
	return nil
}

// Query retrieves log entries based on filters
func (f *FileStorageBackend) Query(ctx context.Context, query *logging.QueryRequest) ([]*logging.LogEntry, error) {
	if !f.initialized {
		return nil, fmt.Errorf("file storage not initialized")
	}

	f.mu.RLock()
	defer f.mu.RUnlock()

	var entries []*logging.LogEntry

	// Get all log files (current + rotated)
	files, err := f.getLogFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get log files: %w", err)
	}

	// Read from files in reverse order (newest first)
	for i := len(files) - 1; i >= 0; i-- {
		fileEntries, err := f.readLogFile(files[i], query)
		if err != nil {
			continue // Skip problematic files
		}
		entries = append(entries, fileEntries...)

		// Check if we have enough entries
		if query.Limit > 0 && len(entries) >= query.Limit+query.Offset {
			break
		}
	}

	// Apply sorting and pagination
	entries = f.applyQueryFilters(entries, query)

	return entries, nil
}

// Close cleanly shuts down the storage backend
func (f *FileStorageBackend) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.currentFile != nil {
		if err := f.currentFile.Sync(); err != nil {
			return fmt.Errorf("failed to sync file: %w", err)
		}
		if err := f.currentFile.Close(); err != nil {
			return fmt.Errorf("failed to close file: %w", err)
		}
		f.currentFile = nil
	}

	f.initialized = false
	return nil
}

// HealthCheck verifies the backend is operational
func (f *FileStorageBackend) HealthCheck(ctx context.Context) error {
	if !f.initialized {
		return fmt.Errorf("file storage not initialized")
	}

	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.currentFile == nil {
		return fmt.Errorf("log file not open")
	}

	if _, err := f.currentFile.Stat(); err != nil {
		return fmt.Errorf("log file stat failed: %w", err)
	}

	return nil
}

// GetCapabilities returns what features this backend supports
func (f *FileStorageBackend) GetCapabilities() logging.BackendCapabilities {
	return logging.BackendCapabilities{
		SupportsQuery:      true,
		SupportsStreaming:  false,
		SupportsRetention:  true,
		SupportsBatchWrite: true,
		SupportsMetrics:    false,
	}
}

// Private helper methods

func (f *FileStorageBackend) parseConfig(config map[string]interface{}) error {
	if path, ok := config["path"].(string); ok {
		f.config.Path = path
	}
	if maxSize, ok := config["max_size"].(float64); ok {
		f.config.MaxSize = int64(maxSize)
	}
	if maxFiles, ok := config["max_files"].(float64); ok {
		f.config.MaxFiles = int(maxFiles)
	}
	if rotate, ok := config["rotate"].(bool); ok {
		f.config.Rotate = rotate
	}
	if permissions, ok := config["permissions"].(string); ok {
		f.config.Permissions = permissions
	}
	if bufferSize, ok := config["buffer_size"].(float64); ok {
		f.config.BufferSize = int(bufferSize)
	}
	if syncInterval, ok := config["sync_interval"].(string); ok {
		f.config.SyncInterval = syncInterval
	}

	return nil
}

func (f *FileStorageBackend) openLogFile() error {
	// Parse permissions
	perm, err := strconv.ParseUint(f.config.Permissions, 8, 32)
	if err != nil {
		perm = 0644
	}

	// Open file for append/create
	file, err := os.OpenFile(f.config.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.FileMode(perm))
	if err != nil {
		return err
	}

	// Get current file size
	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return err
	}

	f.currentFile = file
	f.currentSize = stat.Size()
	return nil
}

func (f *FileStorageBackend) rotateFile() error {
	// Close current file
	if f.currentFile != nil {
		f.currentFile.Close()
	}

	// Move current file to rotated name
	rotatedPath := fmt.Sprintf("%s.1", f.config.Path)
	if err := os.Rename(f.config.Path, rotatedPath); err != nil {
		return err
	}

	// Rotate existing files
	for i := f.config.MaxFiles; i > 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", f.config.Path, i-1)
		newPath := fmt.Sprintf("%s.%d", f.config.Path, i)

		if _, err := os.Stat(oldPath); err == nil {
			os.Rename(oldPath, newPath)
		}
	}

	// Remove oldest file if we exceed max files
	oldestPath := fmt.Sprintf("%s.%d", f.config.Path, f.config.MaxFiles)
	os.Remove(oldestPath)

	return f.openLogFile()
}

func (f *FileStorageBackend) getLogFiles() ([]string, error) {
	var files []string

	// Add current file
	if _, err := os.Stat(f.config.Path); err == nil {
		files = append(files, f.config.Path)
	}

	// Add rotated files
	for i := 1; i <= f.config.MaxFiles; i++ {
		rotatedPath := fmt.Sprintf("%s.%d", f.config.Path, i)
		if _, err := os.Stat(rotatedPath); err == nil {
			files = append(files, rotatedPath)
		}
	}

	return files, nil
}

func (f *FileStorageBackend) readLogFile(path string, query *logging.QueryRequest) ([]*logging.LogEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []*logging.LogEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var entry logging.LogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		if f.matchesQuery(&entry, query) {
			entries = append(entries, &entry)
		}
	}

	return entries, scanner.Err()
}

func (f *FileStorageBackend) matchesQuery(entry *logging.LogEntry, query *logging.QueryRequest) bool {
	if query == nil {
		return true
	}

	// Time range filter
	if query.StartTime != nil && entry.Timestamp.Before(*query.StartTime) {
		return false
	}
	if query.EndTime != nil && entry.Timestamp.After(*query.EndTime) {
		return false
	}

	// Level filter
	if query.Level != "" && entry.Level != query.Level {
		return false
	}

	// String filters
	if query.Logger != "" && entry.Logger != query.Logger {
		return false
	}
	if query.EntityType != "" && entry.EntityType != query.EntityType {
		return false
	}
	if query.EntityID != "" && entry.EntityID != query.EntityID {
		return false
	}
	if query.RequestID != "" && entry.RequestID != query.RequestID {
		return false
	}
	if query.UserID != "" && entry.UserID != query.UserID {
		return false
	}
	if query.OrgID != "" && entry.OrgID != query.OrgID {
		return false
	}

	// Message search
	if query.Message != "" && !strings.Contains(strings.ToLower(entry.Message), strings.ToLower(query.Message)) {
		return false
	}

	return true
}

func (f *FileStorageBackend) applyQueryFilters(entries []*logging.LogEntry, query *logging.QueryRequest) []*logging.LogEntry {
	// Sort by timestamp (newest first by default)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})

	// Apply pagination
	start := query.Offset
	if start > len(entries) {
		return []*logging.LogEntry{}
	}

	end := len(entries)
	if query.Limit > 0 && start+query.Limit < end {
		end = start + query.Limit
	}

	return entries[start:end]
}

// Factory implementation

type FilePluginFactory struct{}

func (f *FilePluginFactory) Create() logging.StorageBackend {
	return &FileStorageBackend{}
}

func (f *FilePluginFactory) GetName() string {
	return "file"
}

func (f *FilePluginFactory) GetDescription() string {
	return "File-based log storage with rotation support"
}

func (f *FilePluginFactory) ValidateConfig(config map[string]interface{}) error {
	if path, ok := config["path"].(string); ok {
		if path == "" {
			return fmt.Errorf("path cannot be empty")
		}
		// Check if directory is writable
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("cannot create log directory: %w", err)
		}
	}

	if maxSize, ok := config["max_size"].(float64); ok {
		if maxSize <= 0 {
			return fmt.Errorf("max_size must be positive")
		}
	}

	if maxFiles, ok := config["max_files"].(float64); ok {
		if maxFiles < 1 {
			return fmt.Errorf("max_files must be at least 1")
		}
	}

	return nil
}

// NewFilePlugin creates a new file plugin factory
func NewFilePlugin() logging.PluginFactory {
	return &FilePluginFactory{}
}
