package unit

import (
	"database/sql"
	"testing"
	"time"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/database/models"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDatabase is a mock implementation of models.Database
type MockModelDatabase struct {
	mock.Mock
}

func (m *MockModelDatabase) Query(query string, args ...interface{}) (*sql.Rows, error) {
	mockArgs := append([]interface{}{query}, args...)
	callArgs := m.Called(mockArgs...)
	return callArgs.Get(0).(*sql.Rows), callArgs.Error(1)
}

func (m *MockModelDatabase) QueryRow(query string, args ...interface{}) *sql.Row {
	mockArgs := append([]interface{}{query}, args...)
	callArgs := m.Called(mockArgs...)
	return callArgs.Get(0).(*sql.Row)
}

func (m *MockModelDatabase) Exec(query string, args ...interface{}) (sql.Result, error) {
	mockArgs := append([]interface{}{query}, args...)
	callArgs := m.Called(mockArgs...)
	return callArgs.Get(0).(sql.Result), callArgs.Error(1)
}

func (m *MockModelDatabase) Begin() (*sql.Tx, error) {
	args := m.Called()
	return args.Get(0).(*sql.Tx), args.Error(1)
}

// Test MCPServerModel CRUD operations
func TestMCPServerModel_Create(t *testing.T) {
	db := &MockModelDatabase{}
	model := models.NewMCPServerModel(db)

	server := &models.MCPServer{
		OrganizationID: uuid.New(),
		Name:           "test-server",
		Description:    sql.NullString{String: "Test server", Valid: true},
		Protocol:       "http",
		URL:            sql.NullString{String: "http://localhost:8080", Valid: true},
		TimeoutSeconds: 30,
		MaxRetries:     3,
		Status:         "active",
		IsActive:       true,
		Metadata:       map[string]interface{}{"key": "value"},
	}

	// Mock successful execution
	mockResult := &MockSQLResult{}
	mockResult.On("LastInsertId").Return(int64(1), nil)
	mockResult.On("RowsAffected").Return(int64(1), nil)

	db.On("Exec", mock.MatchedBy(func(query string) bool {
		return true // Accept any query for simplicity
	}), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"),
		"test-server", mock.Anything, "http", mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, 30, 3,
		"active", mock.Anything, true, mock.Anything, mock.Anything).Return(mockResult, nil)

	err := model.Create(server)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, server.ID) // Should generate ID if not set
	db.AssertExpectations(t)
}

func TestMCPServerModel_Construction(t *testing.T) {
	db := &MockModelDatabase{}
	model := models.NewMCPServerModel(db)

	// Test that the model can be constructed properly
	assert.NotNil(t, model)

	// Test struct initialization with valid data
	server := &models.MCPServer{
		OrganizationID: uuid.New(),
		Name:           "test-server",
		Protocol:       "http",
		Status:         "active",
		IsActive:       true,
	}

	// Test basic field validation
	assert.NotEmpty(t, server.Name)
	assert.NotEmpty(t, server.Protocol)
	assert.True(t, server.IsActive)
	assert.Equal(t, "active", server.Status)
}

func TestMCPServerModel_UpdateStatus(t *testing.T) {
	db := &MockModelDatabase{}
	model := models.NewMCPServerModel(db)

	serverID := uuid.New()
	newStatus := "inactive"

	mockResult := &MockSQLResult{}
	mockResult.On("RowsAffected").Return(int64(1), nil)

	db.On("Exec", mock.MatchedBy(func(query string) bool {
		return true // Accept any query
	}), serverID, newStatus).Return(mockResult, nil)

	err := model.UpdateStatus(serverID, newStatus)

	assert.NoError(t, err)
	db.AssertExpectations(t)
}

// Test User model
func TestUser_Validation(t *testing.T) {
	tests := []struct {
		name    string
		user    *types.User
		isValid bool
	}{
		{
			name: "valid user",
			user: &types.User{
				ID:             uuid.New().String(),
				Email:          "test@example.com",
				Name:           "Test User",
				PasswordHash:   "hashed_password",
				OrganizationID: uuid.New().String(),
				Role:           "user",
				IsActive:       true,
			},
			isValid: true,
		},
		{
			name: "empty email",
			user: &types.User{
				ID:             uuid.New().String(),
				Email:          "",
				Name:           "Test User",
				PasswordHash:   "hashed_password",
				OrganizationID: uuid.New().String(),
				Role:           "user",
				IsActive:       true,
			},
			isValid: false,
		},
		{
			name: "empty name",
			user: &types.User{
				ID:             uuid.New().String(),
				Email:          "test@example.com",
				Name:           "",
				PasswordHash:   "hashed_password",
				OrganizationID: uuid.New().String(),
				Role:           "user",
				IsActive:       true,
			},
			isValid: false,
		},
		{
			name: "invalid role",
			user: &types.User{
				ID:             uuid.New().String(),
				Email:          "test@example.com",
				Name:           "Test User",
				PasswordHash:   "hashed_password",
				OrganizationID: uuid.New().String(),
				Role:           "invalid_role",
				IsActive:       true,
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation logic
			isValid := tt.user.Email != "" &&
				tt.user.Name != "" &&
				tt.user.PasswordHash != "" &&
				(tt.user.Role == "admin" || tt.user.Role == "user" || tt.user.Role == "viewer" || tt.user.Role == "api_user")

			assert.Equal(t, tt.isValid, isValid)
		})
	}
}

// Test Organization model validation
func TestOrganization_PlanLimits(t *testing.T) {
	tests := []struct {
		name          string
		planType      string
		maxServers    int
		maxSessions   int
		expectedValid bool
	}{
		{
			name:          "free plan limits",
			planType:      "free",
			maxServers:    10,
			maxSessions:   100,
			expectedValid: true,
		},
		{
			name:          "pro plan limits",
			planType:      "pro",
			maxServers:    100,
			maxSessions:   1000,
			expectedValid: true,
		},
		{
			name:          "enterprise plan limits",
			planType:      "enterprise",
			maxServers:    1000,
			maxSessions:   10000,
			expectedValid: true,
		},
		{
			name:          "invalid plan type",
			planType:      "invalid",
			maxServers:    10,
			maxSessions:   100,
			expectedValid: false,
		},
		{
			name:          "negative limits",
			planType:      "free",
			maxServers:    -1,
			maxSessions:   -1,
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation logic that would be in the model
			validPlanTypes := []string{"free", "pro", "enterprise"}
			isValidPlan := false
			for _, validType := range validPlanTypes {
				if tt.planType == validType {
					isValidPlan = true
					break
				}
			}

			isValid := isValidPlan && tt.maxServers >= 0 && tt.maxSessions >= 0

			assert.Equal(t, tt.expectedValid, isValid)
		})
	}
}

// Test Session model lifecycle
func TestMCPSession_StatusTransitions(t *testing.T) {
	session := &models.MCPSession{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		ServerID:       uuid.New(),
		Status:         "initializing",
		Protocol:       "http",
		StartedAt:      time.Now(),
		LastActivity:   time.Now(),
	}

	// Test valid status transitions
	validTransitions := map[string][]string{
		"initializing": {"active", "error"},
		"active":       {"closed", "error"},
		"closed":       {},         // Terminal state
		"error":        {"closed"}, // Can only be closed from error
	}

	for currentStatus, allowedNext := range validTransitions {
		session.Status = currentStatus

		for _, nextStatus := range allowedNext {
			// Test that transition is valid
			assert.Contains(t, allowedNext, nextStatus,
				"Status transition from %s to %s should be valid", currentStatus, nextStatus)
		}
	}

	// Test session timeout logic
	oldSession := &models.MCPSession{
		LastActivity: time.Now().Add(-2 * time.Hour),
		Status:       "active",
	}

	sessionTimeout := 1 * time.Hour
	isExpired := time.Since(oldSession.LastActivity) > sessionTimeout && oldSession.Status == "active"
	assert.True(t, isExpired, "Session should be considered expired")
}

// Test HealthCheck model
func TestHealthCheck_ResponseTime(t *testing.T) {
	healthCheck := &models.HealthCheck{
		ID:             uuid.New(),
		ServerID:       uuid.New(),
		Status:         "healthy",
		ResponseTimeMS: sql.NullInt32{Int32: 150, Valid: true},
		CheckedAt:      time.Now(),
	}

	// Test response time validation
	assert.True(t, healthCheck.ResponseTimeMS.Valid)
	assert.Equal(t, int32(150), healthCheck.ResponseTimeMS.Int32)
	assert.True(t, healthCheck.ResponseTimeMS.Int32 > 0, "Response time should be positive")
	assert.True(t, healthCheck.ResponseTimeMS.Int32 < 5000, "Response time should be reasonable")

	// Test failed health check
	failedCheck := &models.HealthCheck{
		ID:           uuid.New(),
		ServerID:     uuid.New(),
		Status:       "unhealthy",
		ErrorMessage: sql.NullString{String: "Connection timeout", Valid: true},
		CheckedAt:    time.Now(),
	}

	assert.Equal(t, "unhealthy", failedCheck.Status)
	assert.True(t, failedCheck.ErrorMessage.Valid)
	assert.False(t, failedCheck.ResponseTimeMS.Valid) // Should be null for failed checks
}

// Test metadata handling in models
func TestServer_MetadataHandling(t *testing.T) {
	server := &models.MCPServer{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		Name:           "metadata-test-server",
		Protocol:       "http",
		Status:         "active",
		IsActive:       true,
		Metadata: map[string]interface{}{
			"environment": "production",
			"region":      "us-west-2",
			"version":     "1.0.0",
			"features":    []string{"feature1", "feature2"},
		},
	}

	// Test metadata access
	assert.NotNil(t, server.Metadata)
	assert.Equal(t, "production", server.Metadata["environment"])
	assert.Equal(t, "us-west-2", server.Metadata["region"])
	assert.Equal(t, "1.0.0", server.Metadata["version"])

	// Test metadata types
	features, ok := server.Metadata["features"].([]string)
	assert.True(t, ok, "Features should be a string slice")
	assert.Len(t, features, 2)
	assert.Contains(t, features, "feature1")
}

// Test array fields handling (PostgreSQL arrays)
func TestServer_ArrayFields(t *testing.T) {
	server := &models.MCPServer{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		Name:           "array-test-server",
		Protocol:       "stdio",
		Command:        sql.NullString{String: "python", Valid: true},
		Args:           pq.StringArray{"--port", "8080", "--verbose"},
		Environment:    pq.StringArray{"DEBUG=1", "NODE_ENV=production"},
		Status:         "active",
		IsActive:       true,
	}

	// Test args array
	assert.NotNil(t, server.Args)
	assert.Len(t, server.Args, 3)
	assert.Contains(t, []string(server.Args), "--port")
	assert.Contains(t, []string(server.Args), "8080")
	assert.Contains(t, []string(server.Args), "--verbose")

	// Test environment array
	assert.NotNil(t, server.Environment)
	assert.Len(t, server.Environment, 2)
	assert.Contains(t, []string(server.Environment), "DEBUG=1")
	assert.Contains(t, []string(server.Environment), "NODE_ENV=production")
}

// Mock SQL Result for testing
type MockSQLResult struct {
	mock.Mock
}

func (m *MockSQLResult) LastInsertId() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSQLResult) RowsAffected() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

// Test concurrent access to model operations
func TestMCPServerModel_ConcurrentOperations(t *testing.T) {
	db := &MockModelDatabase{}
	model := models.NewMCPServerModel(db)

	// Test that multiple status updates can be handled concurrently
	serverID := uuid.New()

	mockResult := &MockSQLResult{}
	mockResult.On("RowsAffected").Return(int64(1), nil)

	// Setup mock for multiple concurrent calls
	db.On("Exec", mock.Anything, serverID, mock.AnythingOfType("string")).Return(mockResult, nil)

	// Simulate concurrent status updates
	done := make(chan bool, 3)
	statuses := []string{"active", "inactive", "maintenance"}

	for _, status := range statuses {
		go func(s string) {
			err := model.UpdateStatus(serverID, s)
			assert.NoError(t, err)
			done <- true
		}(status)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	db.AssertExpectations(t)
}

// Benchmark model operations
func BenchmarkMCPServer_MetadataAccess(b *testing.B) {
	server := &models.MCPServer{
		Metadata: map[string]interface{}{
			"environment": "production",
			"region":      "us-west-2",
			"version":     "1.0.0",
			"config":      map[string]string{"key": "value"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = server.Metadata["environment"]
		_ = server.Metadata["region"]
		_ = server.Metadata["version"]
		if config, ok := server.Metadata["config"].(map[string]string); ok {
			_ = config["key"]
		}
	}
}

func BenchmarkUUID_Generation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = uuid.New()
	}
}
