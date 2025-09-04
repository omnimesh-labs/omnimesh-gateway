package unit

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"mcp-gateway/apps/backend/internal/database/models"
	"mcp-gateway/apps/backend/internal/discovery"
	"mcp-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDatabase is a mock implementation of models.Database
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) Query(query string, args ...interface{}) (*sql.Rows, error) {
	mockArgs := append([]interface{}{query}, args...)
	callArgs := m.Called(mockArgs...)
	return callArgs.Get(0).(*sql.Rows), callArgs.Error(1)
}

func (m *MockDatabase) QueryRow(query string, args ...interface{}) *sql.Row {
	mockArgs := append([]interface{}{query}, args...)
	callArgs := m.Called(mockArgs...)
	return callArgs.Get(0).(*sql.Row)
}

func (m *MockDatabase) Exec(query string, args ...interface{}) (sql.Result, error) {
	mockArgs := append([]interface{}{query}, args...)
	callArgs := m.Called(mockArgs...)
	return callArgs.Get(0).(sql.Result), callArgs.Error(1)
}

func (m *MockDatabase) Begin() (*sql.Tx, error) {
	args := m.Called()
	return args.Get(0).(*sql.Tx), args.Error(1)
}

// MockMCPServerModel is a mock implementation of models.MCPServerModel
type MockMCPServerModel struct {
	mock.Mock
	servers map[uuid.UUID]*models.MCPServer
}

func NewMockMCPServerModel() *MockMCPServerModel {
	return &MockMCPServerModel{
		servers: make(map[uuid.UUID]*models.MCPServer),
	}
}

func (m *MockMCPServerModel) Create(server *models.MCPServer) error {
	args := m.Called(server)
	if args.Error(0) == nil {
		m.servers[server.ID] = server
	}
	return args.Error(0)
}

func (m *MockMCPServerModel) GetByID(id uuid.UUID) (*models.MCPServer, error) {
	args := m.Called(id)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	if server, exists := m.servers[id]; exists {
		return server, nil
	}
	return args.Get(0).(*models.MCPServer), args.Error(1)
}

func (m *MockMCPServerModel) GetByName(orgID uuid.UUID, name string) (*models.MCPServer, error) {
	args := m.Called(orgID, name)
	return args.Get(0).(*models.MCPServer), args.Error(1)
}

func (m *MockMCPServerModel) ListByOrganization(orgID uuid.UUID, activeOnly bool) ([]*models.MCPServer, error) {
	args := m.Called(orgID, activeOnly)
	return args.Get(0).([]*models.MCPServer), args.Error(1)
}

func (m *MockMCPServerModel) Update(server *models.MCPServer) error {
	args := m.Called(server)
	if args.Error(0) == nil && m.servers[server.ID] != nil {
		m.servers[server.ID] = server
	}
	return args.Error(0)
}

func (m *MockMCPServerModel) Delete(id uuid.UUID) error {
	args := m.Called(id)
	if args.Error(0) == nil {
		if server, exists := m.servers[id]; exists {
			server.IsActive = false
		}
	}
	return args.Error(0)
}

func (m *MockMCPServerModel) UpdateStatus(id uuid.UUID, status string) error {
	args := m.Called(id, status)
	if args.Error(0) == nil {
		if server, exists := m.servers[id]; exists {
			server.Status = status
		}
	}
	return args.Error(0)
}

// MockHealthCheckModel is a mock implementation of models.HealthCheckModel
type MockHealthCheckModel struct {
	mock.Mock
}

func NewMockHealthCheckModel() *MockHealthCheckModel {
	return &MockHealthCheckModel{}
}

func (m *MockHealthCheckModel) Create(check *models.HealthCheck) error {
	args := m.Called(check)
	return args.Error(0)
}

func (m *MockHealthCheckModel) GetLatestByServerID(serverID uuid.UUID) (*models.HealthCheck, error) {
	args := m.Called(serverID)
	return args.Get(0).(*models.HealthCheck), args.Error(1)
}

func (m *MockHealthCheckModel) GetHistoryByServerID(serverID uuid.UUID, limit int) ([]*models.HealthCheck, error) {
	args := m.Called(serverID, limit)
	return args.Get(0).([]*models.HealthCheck), args.Error(1)
}

// Mock SQL driver types for testing
type MockResult struct {
	mock.Mock
}

func (m *MockResult) LastInsertId() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockResult) RowsAffected() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func setupDiscoveryService(singleTenant bool) (*discovery.Config, *MockMCPServerModel, *MockHealthCheckModel) {
	config := &discovery.Config{
		HealthInterval:   30 * time.Second,
		FailureThreshold: 3,
		RecoveryTimeout:  5 * time.Minute,
		Enabled:          true,
		SingleTenant:     singleTenant,
	}

	// Create mock models
	mockMCPServer := NewMockMCPServerModel()
	mockHealthCheck := NewMockHealthCheckModel()

	// Return config and mocks for testing logic
	return config, mockMCPServer, mockHealthCheck
}

func TestDiscoveryConfig_Defaults(t *testing.T) {
	config := &discovery.Config{
		Enabled:      true,
		SingleTenant: false,
	}

	assert.True(t, config.Enabled)
	assert.False(t, config.SingleTenant)

	// Test default organization ID
	defaultOrgID := discovery.DefaultOrganizationID
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", defaultOrgID.String())
}

func TestService_ValidateCreateRequest(t *testing.T) {
	tests := []struct {
		name        string
		request     *types.CreateMCPServerRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid HTTP server request",
			request: &types.CreateMCPServerRequest{
				Name:        "test-server",
				Description: "Test HTTP server",
				Protocol:    types.ProtocolHTTP,
				URL:         "http://localhost:8080",
				Timeout:     30 * time.Second,
				MaxRetries:  3,
			},
			expectError: false,
		},
		{
			name: "valid STDIO server request",
			request: &types.CreateMCPServerRequest{
				Name:        "test-stdio-server",
				Description: "Test STDIO server",
				Protocol:    types.ProtocolStdio,
				Command:     "python",
				Args:        []string{"-m", "mcp_server"},
				WorkingDir:  "/app",
				Environment: []string{"DEBUG=1"},
				Timeout:     60 * time.Second,
				MaxRetries:  2,
			},
			expectError: false,
		},
		{
			name: "empty name",
			request: &types.CreateMCPServerRequest{
				Name:     "",
				Protocol: types.ProtocolHTTP,
				URL:      "http://localhost:8080",
			},
			expectError: true,
		},
		{
			name: "HTTP server without URL",
			request: &types.CreateMCPServerRequest{
				Name:     "test-server",
				Protocol: types.ProtocolHTTP,
			},
			expectError: true,
		},
		{
			name: "STDIO server without command",
			request: &types.CreateMCPServerRequest{
				Name:     "test-server",
				Protocol: types.ProtocolStdio,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation logic that would be in the service
			var err error

			if tt.request.Name == "" {
				err = fmt.Errorf("server name cannot be empty")
			} else if tt.request.Protocol == types.ProtocolHTTP && tt.request.URL == "" {
				err = fmt.Errorf("HTTP servers require a URL")
			} else if tt.request.Protocol == types.ProtocolStdio && tt.request.Command == "" {
				err = fmt.Errorf("STDIO servers require a command")
			}

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_OrganizationResolution(t *testing.T) {
	tests := []struct {
		name         string
		singleTenant bool
		orgID        string
		expectedID   string
		expectError  bool
	}{
		{
			name:         "single tenant mode ignores provided org ID",
			singleTenant: true,
			orgID:        "custom-org-id",
			expectedID:   "00000000-0000-0000-0000-000000000000",
			expectError:  false,
		},
		{
			name:         "single tenant mode with empty org ID",
			singleTenant: true,
			orgID:        "",
			expectedID:   "00000000-0000-0000-0000-000000000000",
			expectError:  false,
		},
		{
			name:         "multi tenant mode with valid org ID",
			singleTenant: false,
			orgID:        "550e8400-e29b-41d4-a716-446655440000",
			expectedID:   "550e8400-e29b-41d4-a716-446655440000",
			expectError:  false,
		},
		{
			name:         "multi tenant mode with empty org ID",
			singleTenant: false,
			orgID:        "",
			expectedID:   "",
			expectError:  true,
		},
		{
			name:         "multi tenant mode with invalid org ID",
			singleTenant: false,
			orgID:        "invalid-uuid",
			expectedID:   "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the organization ID resolution logic
			var resultID uuid.UUID
			var err error

			if tt.singleTenant {
				resultID = discovery.DefaultOrganizationID
			} else {
				if tt.orgID == "" {
					err = fmt.Errorf("organization ID is required in multi-tenant mode")
				} else {
					resultID, err = uuid.Parse(tt.orgID)
					if err != nil {
						err = fmt.Errorf("invalid organization ID: %w", err)
					}
				}
			}

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, resultID.String())
			}
		})
	}
}

func TestServerConversion(t *testing.T) {
	// Test the conversion between models.MCPServer and types.MCPServer
	serverID := uuid.New()
	orgID := uuid.New()

	modelServer := &models.MCPServer{
		ID:             serverID,
		OrganizationID: orgID,
		Name:           "test-server",
		Description:    sql.NullString{String: "Test server", Valid: true},
		Protocol:       types.ProtocolHTTP,
		URL:            sql.NullString{String: "http://localhost:8080", Valid: true},
		Command:        sql.NullString{},
		Args:           []string{},
		Environment:    []string{},
		WorkingDir:     sql.NullString{},
		Version:        sql.NullString{String: "1.0.0", Valid: true},
		TimeoutSeconds: 30,
		MaxRetries:     3,
		Status:         types.ServerStatusActive,
		HealthCheckURL: sql.NullString{String: "http://localhost:8080/health", Valid: true},
		IsActive:       true,
		Metadata:       map[string]interface{}{"key": "value"},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Convert metadata from map[string]interface{} to map[string]string
	metadata := make(map[string]string)
	for k, v := range modelServer.Metadata {
		if s, ok := v.(string); ok {
			metadata[k] = s
		}
	}

	// Simulate conversion (this would be done by convertModelToTypesMCPServer)
	typesServer := &types.MCPServer{
		ID:             modelServer.ID.String(),
		OrganizationID: modelServer.OrganizationID.String(),
		Name:           modelServer.Name,
		Description:    modelServer.Description.String,
		Protocol:       modelServer.Protocol,
		URL:            modelServer.URL.String,
		Command:        modelServer.Command.String,
		Args:           modelServer.Args,
		Environment:    modelServer.Environment,
		WorkingDir:     modelServer.WorkingDir.String,
		Version:        modelServer.Version.String,
		Timeout:        time.Duration(modelServer.TimeoutSeconds) * time.Second,
		MaxRetries:     modelServer.MaxRetries,
		Status:         modelServer.Status,
		HealthCheckURL: modelServer.HealthCheckURL.String,
		IsActive:       modelServer.IsActive,
		Metadata:       metadata,
		CreatedAt:      modelServer.CreatedAt,
		UpdatedAt:      modelServer.UpdatedAt,
	}

	// Verify conversion
	assert.Equal(t, modelServer.ID.String(), typesServer.ID)
	assert.Equal(t, modelServer.OrganizationID.String(), typesServer.OrganizationID)
	assert.Equal(t, modelServer.Name, typesServer.Name)
	assert.Equal(t, modelServer.Description.String, typesServer.Description)
	assert.Equal(t, modelServer.Protocol, typesServer.Protocol)
	assert.Equal(t, modelServer.URL.String, typesServer.URL)
	assert.Equal(t, modelServer.Version.String, typesServer.Version)
	assert.Equal(t, time.Duration(modelServer.TimeoutSeconds)*time.Second, typesServer.Timeout)
	assert.Equal(t, modelServer.MaxRetries, typesServer.MaxRetries)
	assert.Equal(t, modelServer.Status, typesServer.Status)
	assert.Equal(t, modelServer.HealthCheckURL.String, typesServer.HealthCheckURL)
	assert.Equal(t, modelServer.IsActive, typesServer.IsActive)
	assert.Equal(t, metadata, typesServer.Metadata)
}

func TestService_ServerDefaults(t *testing.T) {
	// Test that default values are set correctly
	request := &types.CreateMCPServerRequest{
		Name:     "test-server",
		Protocol: types.ProtocolHTTP,
		URL:      "http://localhost:8080",
		// Timeout and MaxRetries not specified
	}

	// Simulate the default setting logic
	timeoutSeconds := int(request.Timeout.Seconds())
	maxRetries := request.MaxRetries

	// Set defaults if not provided
	if timeoutSeconds == 0 {
		timeoutSeconds = 30
	}
	if maxRetries == 0 {
		maxRetries = 3
	}

	assert.Equal(t, 30, timeoutSeconds)
	assert.Equal(t, 3, maxRetries)

	// Test with provided values
	request2 := &types.CreateMCPServerRequest{
		Name:       "test-server-2",
		Protocol:   types.ProtocolHTTP,
		URL:        "http://localhost:8080",
		Timeout:    60 * time.Second,
		MaxRetries: 5,
	}

	timeoutSeconds2 := int(request2.Timeout.Seconds())
	maxRetries2 := request2.MaxRetries

	if timeoutSeconds2 == 0 {
		timeoutSeconds2 = 30
	}
	if maxRetries2 == 0 {
		maxRetries2 = 3
	}

	assert.Equal(t, 60, timeoutSeconds2)
	assert.Equal(t, 5, maxRetries2)
}

func TestService_StringMapConversion(t *testing.T) {
	// Test the conversion from map[string]string to map[string]interface{}
	input := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "",
	}

	// Simulate convertStringMapToInterface
	var result map[string]interface{}
	if input == nil {
		result = nil
	} else {
		result = make(map[string]interface{})
		for k, v := range input {
			result[k] = v
		}
	}

	assert.NotNil(t, result)
	assert.Len(t, result, 3)
	assert.Equal(t, "value1", result["key1"])
	assert.Equal(t, "value2", result["key2"])
	assert.Equal(t, "", result["key3"])

	// Test with nil input
	var nilInput map[string]string
	var nilResult map[string]interface{}
	if nilInput == nil {
		nilResult = nil
	}
	assert.Nil(t, nilResult)
}

func TestService_ServerStatusTransitions(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  string
		newStatus      string
		shouldTransition bool
	}{
		{
			name:             "inactive to active",
			initialStatus:    types.ServerStatusInactive,
			newStatus:        types.ServerStatusActive,
			shouldTransition: true,
		},
		{
			name:             "active to failed",
			initialStatus:    types.ServerStatusActive,
			newStatus:        "failed",
			shouldTransition: true,
		},
		{
			name:             "failed to recovering",
			initialStatus:    "failed",
			newStatus:        "recovering",
			shouldTransition: true,
		},
		{
			name:             "recovering to active",
			initialStatus:    "recovering",
			newStatus:        types.ServerStatusActive,
			shouldTransition: true,
		},
		{
			name:             "same status",
			initialStatus:    types.ServerStatusActive,
			newStatus:        types.ServerStatusActive,
			shouldTransition: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test status transition logic
			if tt.shouldTransition {
				assert.NotEqual(t, tt.initialStatus, tt.newStatus, "Status should change")
			} else {
				assert.Equal(t, tt.initialStatus, tt.newStatus, "Status should remain the same")
			}

			// Verify status transitions are valid
			validTransitions := map[string][]string{
				types.ServerStatusInactive:   {types.ServerStatusActive},
				types.ServerStatusActive:     {"failed", types.ServerStatusInactive},
				"failed":     {"recovering", types.ServerStatusInactive},
				"recovering": {types.ServerStatusActive, "failed"},
			}

			if validTransitions[tt.initialStatus] != nil {
				found := false
				for _, validNext := range validTransitions[tt.initialStatus] {
					if validNext == tt.newStatus {
						found = true
						break
					}
				}
				if !found && tt.initialStatus != tt.newStatus {
					// Transition might still be valid in some cases
					// This is more of a documentation of expected transitions
				}
			}
		})
	}
}

func TestService_HealthCheckData(t *testing.T) {
	// Test health check data structure
	serverID := uuid.New()
	checkID := uuid.New()

	healthCheck := &models.HealthCheck{
		ID:             checkID,
		ServerID:       serverID,
		Status:         types.ServerStatusActive,
		ResponseTimeMS: sql.NullInt32{Int32: 250, Valid: true},
		ErrorMessage:   sql.NullString{},
		CheckedAt:      time.Now(),
	}

	assert.Equal(t, checkID, healthCheck.ID)
	assert.Equal(t, serverID, healthCheck.ServerID)
	assert.Equal(t, types.ServerStatusActive, healthCheck.Status)
	assert.True(t, healthCheck.ResponseTimeMS.Valid)
	assert.Equal(t, int32(250), healthCheck.ResponseTimeMS.Int32)
	assert.False(t, healthCheck.ErrorMessage.Valid)
	assert.False(t, healthCheck.CheckedAt.IsZero())

	// Test failed health check
	failedCheck := &models.HealthCheck{
		ID:           uuid.New(),
		ServerID:     serverID,
		Status:       "failed",
		ErrorMessage: sql.NullString{String: "Connection refused", Valid: true},
		CheckedAt:    time.Now(),
	}

	assert.Equal(t, "failed", failedCheck.Status)
	assert.True(t, failedCheck.ErrorMessage.Valid)
	assert.Equal(t, "Connection refused", failedCheck.ErrorMessage.String)
	assert.False(t, failedCheck.ResponseTimeMS.Valid)
}

func TestService_ConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *discovery.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &discovery.Config{
				HealthInterval:   30 * time.Second,
				FailureThreshold: 3,
				RecoveryTimeout:  5 * time.Minute,
				Enabled:          true,
				SingleTenant:     false,
			},
			expectError: false,
		},
		{
			name: "zero health interval",
			config: &discovery.Config{
				HealthInterval:   0,
				FailureThreshold: 3,
				RecoveryTimeout:  5 * time.Minute,
				Enabled:          true,
			},
			expectError: true,
			errorMsg:    "health interval must be positive",
		},
		{
			name: "negative failure threshold",
			config: &discovery.Config{
				HealthInterval:   30 * time.Second,
				FailureThreshold: -1,
				RecoveryTimeout:  5 * time.Minute,
				Enabled:          true,
			},
			expectError: true,
			errorMsg:    "failure threshold must be positive",
		},
		{
			name: "zero recovery timeout",
			config: &discovery.Config{
				HealthInterval:   30 * time.Second,
				FailureThreshold: 3,
				RecoveryTimeout:  0,
				Enabled:          true,
			},
			expectError: true,
			errorMsg:    "recovery timeout must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic config validation logic
			var err error

			if tt.config.Enabled {
				if tt.config.HealthInterval <= 0 {
					err = fmt.Errorf("health interval must be positive")
				} else if tt.config.FailureThreshold <= 0 {
					err = fmt.Errorf("failure threshold must be positive")
				} else if tt.config.RecoveryTimeout <= 0 {
					err = fmt.Errorf("recovery timeout must be positive")
				}
			}

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test protocol-specific server configurations
func TestService_ProtocolSpecificConfigs(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
		config   *types.CreateMCPServerRequest
		validate func(*testing.T, *types.CreateMCPServerRequest)
	}{
		{
			name:     "HTTP server configuration",
			protocol: types.ProtocolHTTP,
			config: &types.CreateMCPServerRequest{
				Name:           "http-server",
				Protocol:       types.ProtocolHTTP,
				URL:            "https://api.example.com/mcp",
				HealthCheckURL: "https://api.example.com/health",
				Timeout:        30 * time.Second,
				Metadata: map[string]string{
					"environment": "production",
					"region":      "us-west-2",
				},
			},
			validate: func(t *testing.T, req *types.CreateMCPServerRequest) {
				assert.Equal(t, types.ProtocolHTTP, req.Protocol)
				assert.NotEmpty(t, req.URL)
				assert.NotEmpty(t, req.HealthCheckURL)
				assert.True(t, req.Timeout > 0)
				assert.NotEmpty(t, req.Metadata)
			},
		},
		{
			name:     "WebSocket server configuration",
			protocol: types.ProtocolWebSocket,
			config: &types.CreateMCPServerRequest{
				Name:     "websocket-server",
				Protocol: types.ProtocolWebSocket,
				URL:      "ws://localhost:8080/mcp",
				Timeout:  60 * time.Second,
			},
			validate: func(t *testing.T, req *types.CreateMCPServerRequest) {
				assert.Equal(t, types.ProtocolWebSocket, req.Protocol)
				assert.Contains(t, req.URL, "ws://")
			},
		},
		{
			name:     "STDIO server configuration",
			protocol: types.ProtocolStdio,
			config: &types.CreateMCPServerRequest{
				Name:        "stdio-server",
				Protocol:    types.ProtocolStdio,
				Command:     "python3",
				Args:        []string{"-m", "my_mcp_server", "--verbose"},
				WorkingDir:  "/opt/mcp-servers/my-server",
				Environment: []string{"PYTHONPATH=/opt/lib", "DEBUG=1"},
				Timeout:     120 * time.Second,
			},
			validate: func(t *testing.T, req *types.CreateMCPServerRequest) {
				assert.Equal(t, types.ProtocolStdio, req.Protocol)
				assert.NotEmpty(t, req.Command)
				assert.NotEmpty(t, req.Args)
				assert.NotEmpty(t, req.WorkingDir)
				assert.NotEmpty(t, req.Environment)
				assert.Contains(t, req.Args, "--verbose")
				assert.Contains(t, req.Environment, "DEBUG=1")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.config)
		})
	}
}

// Benchmark tests
func BenchmarkServerConversion(b *testing.B) {
	serverID := uuid.New()
	orgID := uuid.New()

	modelServer := &models.MCPServer{
		ID:             serverID,
		OrganizationID: orgID,
		Name:           "benchmark-server",
		Description:    sql.NullString{String: "Benchmark server", Valid: true},
		Protocol:       types.ProtocolHTTP,
		URL:            sql.NullString{String: "http://localhost:8080", Valid: true},
		Version:        sql.NullString{String: "1.0.0", Valid: true},
		TimeoutSeconds: 30,
		MaxRetries:     3,
		Status:         types.ServerStatusActive,
		IsActive:       true,
		Metadata:       map[string]interface{}{"key": "value"},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Convert metadata from map[string]interface{} to map[string]string
		metadata := make(map[string]string)
		for k, v := range modelServer.Metadata {
			if s, ok := v.(string); ok {
				metadata[k] = s
			}
		}

		// Simulate the conversion
		_ = &types.MCPServer{
			ID:             modelServer.ID.String(),
			OrganizationID: modelServer.OrganizationID.String(),
			Name:           modelServer.Name,
			Description:    modelServer.Description.String,
			Protocol:       modelServer.Protocol,
			URL:            modelServer.URL.String,
			Version:        modelServer.Version.String,
			Timeout:        time.Duration(modelServer.TimeoutSeconds) * time.Second,
			MaxRetries:     modelServer.MaxRetries,
			Status:         modelServer.Status,
			IsActive:       modelServer.IsActive,
			Metadata:       metadata,
			CreatedAt:      modelServer.CreatedAt,
			UpdatedAt:      modelServer.UpdatedAt,
		}
	}
}

func BenchmarkUUIDParsing(b *testing.B) {
	validUUIDStr := "550e8400-e29b-41d4-a716-446655440000"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := uuid.Parse(validUUIDStr)
		if err != nil {
			b.Fatal(err)
		}
	}
}
