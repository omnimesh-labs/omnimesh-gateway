package unit

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/database/models"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDB implements models.Database interface
type mockDB struct {
	db   *sql.DB
	mock sqlmock.Sqlmock
}

func (m *mockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return m.db.Query(query, args...)
}

func (m *mockDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return m.db.QueryRow(query, args...)
}

func (m *mockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return m.db.Exec(query, args...)
}

func (m *mockDB) Begin() (*sql.Tx, error) {
	return m.db.Begin()
}

func setupMockDB(t *testing.T) (*mockDB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	return &mockDB{db: db, mock: mock}, mock
}

func TestMCPResourceModel_Create(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	defer mockDB.db.Close()

	model := models.NewMCPResourceModel(mockDB)

	orgID := uuid.New()
	userID := uuid.New()

	resource := &models.MCPResource{
		OrganizationID: orgID,
		Name:           "Test Resource",
		ResourceType:   types.ResourceTypeFile,
		URI:            "/path/to/resource.txt",
		IsActive:       true,
		Metadata:       map[string]interface{}{"version": "1.0"},
		Tags:           pq.StringArray{"test", "file"},
		Description:    sql.NullString{String: "Test description", Valid: true},
		MimeType:       sql.NullString{String: "text/plain", Valid: true},
		SizeBytes:      sql.NullInt64{Int64: 1024, Valid: true},
		CreatedBy:      uuid.NullUUID{UUID: userID, Valid: true},
	}

	// Expect the INSERT query
	mock.ExpectExec(`INSERT INTO mcp_resources`).
		WithArgs(sqlmock.AnyArg(), orgID, "Test Resource", "Test description", types.ResourceTypeFile,
			"/path/to/resource.txt", "text/plain", int64(1024), sqlmock.AnyArg(), true,
			sqlmock.AnyArg(), pq.StringArray{"test", "file"}, userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := model.Create(resource)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, resource.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMCPResourceModel_GetByID(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	defer mockDB.db.Close()

	model := models.NewMCPResourceModel(mockDB)

	resourceID := uuid.New()
	orgID := uuid.New()
	userID := uuid.New()

	metadata := map[string]interface{}{"version": "1.0"}
	metadataJSON, _ := json.Marshal(metadata)

	accessPermissions := map[string]interface{}{"read": []interface{}{"*"}}
	accessJSON, _ := json.Marshal(accessPermissions)

	// Expect the SELECT query
	mock.ExpectQuery(`SELECT (.+) FROM mcp_resources WHERE id = \$1`).
		WithArgs(resourceID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description", "resource_type", "uri", "mime_type",
			"size_bytes", "access_permissions", "is_active", "metadata", "tags",
			"created_at", "updated_at", "created_by",
		}).AddRow(
			resourceID, orgID, "Test Resource", "Test description", types.ResourceTypeFile,
			"/path/to/resource.txt", "text/plain", int64(1024), accessJSON, true, metadataJSON,
			pq.StringArray{"test", "file"}, time.Now(), time.Now(), userID,
		))

	resource, err := model.GetByID(resourceID)
	require.NoError(t, err)
	require.NotNil(t, resource)

	assert.Equal(t, resourceID, resource.ID)
	assert.Equal(t, orgID, resource.OrganizationID)
	assert.Equal(t, "Test Resource", resource.Name)
	assert.Equal(t, "Test description", resource.Description.String)
	assert.Equal(t, types.ResourceTypeFile, resource.ResourceType)
	assert.Equal(t, "/path/to/resource.txt", resource.URI)
	assert.Equal(t, "text/plain", resource.MimeType.String)
	assert.Equal(t, int64(1024), resource.SizeBytes.Int64)
	assert.True(t, resource.IsActive)
	assert.Equal(t, metadata, resource.Metadata)
	assert.Equal(t, accessPermissions, resource.AccessPermissions)
	assert.Equal(t, []string{"test", "file"}, []string(resource.Tags))

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMCPResourceModel_ListByOrganization(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	defer mockDB.db.Close()

	model := models.NewMCPResourceModel(mockDB)

	orgID := uuid.New()
	resourceID1 := uuid.New()
	resourceID2 := uuid.New()

	metadata := map[string]interface{}{"version": "1.0"}
	metadataJSON, _ := json.Marshal(metadata)

	// Expect the SELECT query
	mock.ExpectQuery(`SELECT (.+) FROM mcp_resources WHERE organization_id = \$1 AND is_active = true ORDER BY created_at DESC`).
		WithArgs(orgID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description", "resource_type", "uri", "mime_type",
			"size_bytes", "access_permissions", "is_active", "metadata", "tags",
			"created_at", "updated_at", "created_by",
		}).AddRow(
			resourceID1, orgID, "Resource 1", "Description 1", types.ResourceTypeFile,
			"/path/1.txt", "text/plain", nil, []byte("{}"), true, metadataJSON,
			pq.StringArray{"test"}, time.Now(), time.Now(), nil,
		).AddRow(
			resourceID2, orgID, "Resource 2", "Description 2", types.ResourceTypeURL,
			"https://example.com", "text/html", nil, []byte("{}"), true, metadataJSON,
			pq.StringArray{"web"}, time.Now(), time.Now(), nil,
		))

	resources, err := model.ListByOrganization(orgID, true)
	require.NoError(t, err)
	require.Len(t, resources, 2)

	assert.Equal(t, resourceID1, resources[0].ID)
	assert.Equal(t, "Resource 1", resources[0].Name)
	assert.Equal(t, types.ResourceTypeFile, resources[0].ResourceType)

	assert.Equal(t, resourceID2, resources[1].ID)
	assert.Equal(t, "Resource 2", resources[1].Name)
	assert.Equal(t, types.ResourceTypeURL, resources[1].ResourceType)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMCPResourceModel_Update(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	defer mockDB.db.Close()

	model := models.NewMCPResourceModel(mockDB)

	resourceID := uuid.New()
	orgID := uuid.New()

	resource := &models.MCPResource{
		ID:             resourceID,
		OrganizationID: orgID,
		Name:           "Updated Resource",
		ResourceType:   types.ResourceTypeAPI,
		URI:            "https://api.example.com",
		IsActive:       true,
		Metadata:       map[string]interface{}{"version": "2.0"},
		Tags:           pq.StringArray{"api", "updated"},
	}

	// Expect the UPDATE query
	mock.ExpectExec(`UPDATE mcp_resources SET (.+) WHERE id = \$1`).
		WithArgs(resourceID, "Updated Resource", driver.Value(nil), types.ResourceTypeAPI,
			"https://api.example.com", driver.Value(nil), driver.Value(nil), sqlmock.AnyArg(),
			true, sqlmock.AnyArg(), pq.StringArray{"api", "updated"}).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := model.Update(resource)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMCPResourceModel_Delete(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	defer mockDB.db.Close()

	model := models.NewMCPResourceModel(mockDB)

	resourceID := uuid.New()

	// Expect the UPDATE query (soft delete)
	mock.ExpectExec(`UPDATE mcp_resources SET is_active = false WHERE id = \$1`).
		WithArgs(resourceID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := model.Delete(resourceID)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}
