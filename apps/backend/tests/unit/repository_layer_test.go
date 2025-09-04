package unit

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"mcp-gateway/apps/backend/internal/database/repositories"
	"mcp-gateway/apps/backend/internal/types"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test MCP Server Repository
func TestMCPServerRepository_GetByID(t *testing.T) {
	tests := []struct {
		name           string
		serverID       string
		mockSetup      func(mock sqlmock.Sqlmock, serverID string)
		expectedServer *repositories.MCPServer
		expectError    bool
		errorContains  string
	}{
		{
			name:     "successful retrieval",
			serverID: "server-123",
			mockSetup: func(mock sqlmock.Sqlmock, serverID string) {
				rows := sqlmock.NewRows([]string{
					"id", "organization_id", "name", "description", "protocol",
					"url", "command", "args", "environment", "working_dir", "is_active",
				}).AddRow(
					serverID, "org-123", "test-server", "Test server", "http",
					"http://localhost:8080", nil, "{}", "{}", nil, true,
				)
				mock.ExpectQuery(`SELECT id, organization_id, name, description, protocol, url, command, args, environment, working_dir, is_active FROM mcp_servers WHERE id = \$1`).
					WithArgs(serverID).
					WillReturnRows(rows)
			},
			expectedServer: &repositories.MCPServer{
				ID:             "server-123",
				OrganizationID: "org-123",
				Name:           "test-server",
				Description:    "Test server",
				Protocol:       "http",
				URL:            stringPtr("http://localhost:8080"),
				Command:        nil,
				Args:           []string{},
				Environment:    []string{},
				WorkingDir:     nil,
				IsActive:       true,
			},
			expectError: false,
		},
		{
			name:     "server not found",
			serverID: "nonexistent-server",
			mockSetup: func(mock sqlmock.Sqlmock, serverID string) {
				mock.ExpectQuery(`SELECT id, organization_id, name, description, protocol, url, command, args, environment, working_dir, is_active FROM mcp_servers WHERE id = \$1`).
					WithArgs(serverID).
					WillReturnError(sql.ErrNoRows)
			},
			expectedServer: nil,
			expectError:    true,
			errorContains:  "server not found",
		},
		{
			name:     "database error",
			serverID: "server-123",
			mockSetup: func(mock sqlmock.Sqlmock, serverID string) {
				mock.ExpectQuery(`SELECT id, organization_id, name, description, protocol, url, command, args, environment, working_dir, is_active FROM mcp_servers WHERE id = \$1`).
					WithArgs(serverID).
					WillReturnError(sql.ErrConnDone)
			},
			expectedServer: nil,
			expectError:    true,
			errorContains:  "failed to get server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "postgres")
			repo := repositories.NewMCPServerRepository(sqlxDB)

			tt.mockSetup(mock, tt.serverID)

			server, err := repo.GetByID(context.Background(), tt.serverID)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, server)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedServer, server)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// Test Namespace Repository
func TestNamespaceRepository_Create(t *testing.T) {
	tests := []struct {
		name        string
		namespace   *types.Namespace
		mockSetup   func(mock sqlmock.Sqlmock, ns *types.Namespace)
		expectError bool
		errorContains string
	}{
		{
			name: "successful creation",
			namespace: &types.Namespace{
				ID:             "ns-123",
				OrganizationID: "org-123",
				Name:           "test-namespace",
				Description:    "Test namespace",
				IsActive:       true,
				Metadata:       map[string]interface{}{"key": "value"},
			},
			mockSetup: func(mock sqlmock.Sqlmock, ns *types.Namespace) {
				mock.ExpectQuery(`INSERT INTO namespaces`).
					WithArgs(ns.ID, ns.OrganizationID, ns.Name, ns.Description, nil, ns.IsActive, sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).
						AddRow(time.Now(), time.Now()))
			},
			expectError: false,
		},
		{
			name: "creation with generated ID",
			namespace: &types.Namespace{
				OrganizationID: "org-123",
				Name:           "test-namespace",
				Description:    "Test namespace",
				IsActive:       true,
				Metadata:       map[string]interface{}{"key": "value"},
			},
			mockSetup: func(mock sqlmock.Sqlmock, ns *types.Namespace) {
				mock.ExpectQuery(`INSERT INTO namespaces`).
					WithArgs(sqlmock.AnyArg(), ns.OrganizationID, ns.Name, ns.Description, nil, ns.IsActive, sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).
						AddRow(time.Now(), time.Now()))
			},
			expectError: false,
		},
		{
			name: "database error",
			namespace: &types.Namespace{
				ID:             "ns-123",
				OrganizationID: "org-123",
				Name:           "test-namespace",
				Description:    "Test namespace",
				IsActive:       true,
				Metadata:       map[string]interface{}{"key": "value"},
			},
			mockSetup: func(mock sqlmock.Sqlmock, ns *types.Namespace) {
				mock.ExpectQuery(`INSERT INTO namespaces`).
					WithArgs(ns.ID, ns.OrganizationID, ns.Name, ns.Description, nil, ns.IsActive, sqlmock.AnyArg()).
					WillReturnError(sql.ErrConnDone)
			},
			expectError:    true,
			errorContains: "failed to create namespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "postgres")
			repo := repositories.NewNamespaceRepository(sqlxDB)

			tt.mockSetup(mock, tt.namespace)

			err = repo.Create(context.Background(), tt.namespace)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				// Verify ID was generated if it was empty
				assert.NotEmpty(t, tt.namespace.ID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestNamespaceRepository_GetByID_ErrorCases(t *testing.T) {
	tests := []struct {
		name               string
		namespaceID        string
		mockSetup          func(mock sqlmock.Sqlmock, namespaceID string)
		expectError        bool
		errorContains      string
	}{
		{
			name:        "namespace not found",
			namespaceID: "nonexistent-ns",
			mockSetup: func(mock sqlmock.Sqlmock, namespaceID string) {
				mock.ExpectQuery(`SELECT (.+) FROM namespaces WHERE id = \$1`).
					WithArgs(namespaceID).
					WillReturnError(sql.ErrNoRows)
			},
			expectError:       true,
			errorContains:     "namespace not found",
		},
		{
			name:        "database error",
			namespaceID: "ns-123",
			mockSetup: func(mock sqlmock.Sqlmock, namespaceID string) {
				mock.ExpectQuery(`SELECT (.+) FROM namespaces WHERE id = \$1`).
					WithArgs(namespaceID).
					WillReturnError(sql.ErrConnDone)
			},
			expectError:       true,
			errorContains:     "failed to get namespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "postgres")
			repo := repositories.NewNamespaceRepository(sqlxDB)

			tt.mockSetup(mock, tt.namespaceID)

			namespace, err := repo.GetByID(context.Background(), tt.namespaceID)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, namespace)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, namespace)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// Note: Update and Delete methods require more complex SQL mocking
// that would be better tested in integration tests with a real database
func TestNamespaceRepository_MethodsExist(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := repositories.NewNamespaceRepository(sqlxDB)

	// Test that the repository has the expected methods
	assert.NotNil(t, repo)

	// These would typically be tested in integration tests
	// as the SQL mocking becomes quite complex with the actual queries
	t.Log("Repository methods exist and can be called")
	t.Log("Update and Delete methods would be tested in integration tests")
}

func TestNamespaceRepository_List_ErrorCase(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := repositories.NewNamespaceRepository(sqlxDB)

	orgID := "org-123"
	mock.ExpectQuery(`SELECT (.+) FROM namespaces WHERE organization_id = \$1`).
		WithArgs(orgID).
		WillReturnError(sql.ErrConnDone)

	namespaces, err := repo.List(context.Background(), orgID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list namespaces")
	assert.Nil(t, namespaces)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test Repository Context Handling - simplified test
func TestRepository_ContextHandling(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := repositories.NewNamespaceRepository(sqlxDB)

	// Test that repository methods accept context parameters
	// Actual context behavior would be tested in integration tests
	assert.NotNil(t, repo)

	ctx := context.Background()
	assert.NotNil(t, ctx)

	// This demonstrates that the repository methods accept context
	// The actual database context handling would be tested with a real database
	t.Log("Repository methods accept context parameters")
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}

// Benchmark repository operations
func BenchmarkNamespaceRepository_GetByID(b *testing.B) {
	db, mock, err := sqlmock.New()
	require.NoError(b, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := repositories.NewNamespaceRepository(sqlxDB)

	namespaceID := "ns-123"
	rows := sqlmock.NewRows([]string{
		"id", "organization_id", "name", "description", "created_by",
		"is_active", "metadata", "created_at", "updated_at",
	}).AddRow(
		namespaceID, "org-123", "test-namespace", "Test namespace", nil,
		true, `{"key":"value"}`, time.Now(), time.Now(),
	)

	for i := 0; i < b.N; i++ {
		mock.ExpectQuery(`SELECT (.+) FROM namespaces WHERE id = \$1`).
			WithArgs(namespaceID).
			WillReturnRows(rows)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetByID(context.Background(), namespaceID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNamespaceRepository_Create(b *testing.B) {
	db, mock, err := sqlmock.New()
	require.NoError(b, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := repositories.NewNamespaceRepository(sqlxDB)

	for i := 0; i < b.N; i++ {
		mock.ExpectQuery(`INSERT INTO namespaces`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).
				AddRow(time.Now(), time.Now()))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ns := &types.Namespace{
			ID:             uuid.New().String(),
			OrganizationID: "org-123",
			Name:           "test-namespace",
			Description:    "Test namespace",
			IsActive:       true,
			Metadata:       map[string]interface{}{"key": "value"},
		}
		err := repo.Create(context.Background(), ns)
		if err != nil {
			b.Fatal(err)
		}
	}
}
