package unit

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"

	"mcp-gateway/apps/backend/internal/database/models"
	"mcp-gateway/apps/backend/internal/types"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPToolModel_Create(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	defer mockDB.db.Close()

	model := models.NewMCPToolModel(mockDB)

	orgID := uuid.New()
	userID := uuid.New()

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Message to echo",
			},
		},
		"required": []string{"message"},
	}

	examples := []interface{}{
		map[string]interface{}{
			"input":  map[string]interface{}{"message": "hello"},
			"output": map[string]interface{}{"result": "hello"},
		},
	}

	tool := &models.MCPTool{
		OrganizationID:     orgID,
		Name:               "Test Tool",
		FunctionName:       "test_function",
		Schema:             schema,
		Category:           types.ToolCategoryGeneral,
		ImplementationType: types.ToolImplementationInternal,
		TimeoutSeconds:     30,
		MaxRetries:         3,
		UsageCount:         0,
		IsActive:           true,
		IsPublic:           false,
		Metadata:           map[string]interface{}{"version": "1.0"},
		Tags:               pq.StringArray{"test", "function"},
		Examples:           examples,
		Description:        sql.NullString{String: "Test tool description", Valid: true},
		EndpointURL:        sql.NullString{String: "", Valid: false},
		Documentation:      sql.NullString{String: "Test documentation", Valid: true},
		CreatedBy:          uuid.NullUUID{UUID: userID, Valid: true},
	}

	// Expect the INSERT query
	mock.ExpectExec(`INSERT INTO mcp_tools`).
		WithArgs(sqlmock.AnyArg(), orgID, "Test Tool", "Test tool description", "test_function",
			sqlmock.AnyArg(), types.ToolCategoryGeneral, types.ToolImplementationInternal,
			sqlmock.AnyArg(), 30, 3, int64(0), sqlmock.AnyArg(), true, false,
			sqlmock.AnyArg(), pq.StringArray{"test", "function"}, sqlmock.AnyArg(),
			"Test documentation", userID, sqlmock.AnyArg(), "manual", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := model.Create(tool)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, tool.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMCPToolModel_GetByID(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	defer mockDB.db.Close()

	model := models.NewMCPToolModel(mockDB)

	toolID := uuid.New()
	orgID := uuid.New()
	userID := uuid.New()

	schema := map[string]interface{}{"type": "object"}
	schemaJSON, _ := json.Marshal(schema)

	metadata := map[string]interface{}{"version": "1.0"}
	metadataJSON, _ := json.Marshal(metadata)

	accessPermissions := map[string]interface{}{"execute": []interface{}{"*"}}
	accessJSON, _ := json.Marshal(accessPermissions)

	examples := []interface{}{map[string]interface{}{"test": true}}
	examplesJSON, _ := json.Marshal(examples)

	// Expect the SELECT query
	mock.ExpectQuery(`SELECT (.+) FROM mcp_tools WHERE id = \$1`).
		WithArgs(toolID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description", "function_name", "schema", "category",
			"implementation_type", "endpoint_url", "timeout_seconds", "max_retries", "usage_count",
			"access_permissions", "is_active", "is_public", "metadata", "tags", "examples",
			"documentation", "created_at", "updated_at", "created_by", "server_id", "source_type",
			"last_discovered_at", "discovery_metadata",
		}).AddRow(
			toolID, orgID, "Test Tool", "Test description", "test_function", schemaJSON,
			types.ToolCategoryGeneral, types.ToolImplementationInternal, "",
			30, 3, int64(5), accessJSON, true, false, metadataJSON,
			pq.StringArray{"test", "function"}, examplesJSON, "Test docs",
			time.Now(), time.Now(), userID, nil, "manual", nil, []byte("{}"),
		))

	tool, err := model.GetByID(toolID)
	require.NoError(t, err)
	require.NotNil(t, tool)

	assert.Equal(t, toolID, tool.ID)
	assert.Equal(t, orgID, tool.OrganizationID)
	assert.Equal(t, "Test Tool", tool.Name)
	assert.Equal(t, "Test description", tool.Description.String)
	assert.Equal(t, "test_function", tool.FunctionName)
	assert.Equal(t, types.ToolCategoryGeneral, tool.Category)
	assert.Equal(t, types.ToolImplementationInternal, tool.ImplementationType)
	assert.Equal(t, 30, tool.TimeoutSeconds)
	assert.Equal(t, 3, tool.MaxRetries)
	assert.Equal(t, int64(5), tool.UsageCount)
	assert.True(t, tool.IsActive)
	assert.False(t, tool.IsPublic)
	assert.Equal(t, schema, tool.Schema)
	assert.Equal(t, metadata, tool.Metadata)
	assert.Equal(t, accessPermissions, tool.AccessPermissions)
	assert.Equal(t, examples, tool.Examples)
	assert.Equal(t, []string{"test", "function"}, []string(tool.Tags))

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMCPToolModel_GetByFunctionName(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	defer mockDB.db.Close()

	model := models.NewMCPToolModel(mockDB)

	toolID := uuid.New()
	orgID := uuid.New()

	schema := map[string]interface{}{"type": "object"}
	schemaJSON, _ := json.Marshal(schema)

	// Expect the SELECT query with function name
	mock.ExpectQuery(`SELECT (.+) FROM mcp_tools WHERE organization_id = \$1 AND function_name = \$2 AND is_active = true`).
		WithArgs(orgID, "echo_function").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description", "function_name", "schema", "category",
			"implementation_type", "endpoint_url", "timeout_seconds", "max_retries", "usage_count",
			"access_permissions", "is_active", "is_public", "metadata", "tags", "examples",
			"documentation", "created_at", "updated_at", "created_by", "server_id", "source_type",
			"last_discovered_at", "discovery_metadata",
		}).AddRow(
			toolID, orgID, "Echo Tool", "Description", "echo_function", schemaJSON,
			types.ToolCategoryGeneral, types.ToolImplementationInternal, "",
			30, 3, int64(10), []byte("{}"), true, false, []byte("{}"),
			pq.StringArray{"echo"}, []byte("[]"), "Docs",
			time.Now(), time.Now(), nil, nil, "manual", nil, []byte("{}"),
		))

	tool, err := model.GetByFunctionName(orgID, "echo_function")
	require.NoError(t, err)
	require.NotNil(t, tool)

	assert.Equal(t, toolID, tool.ID)
	assert.Equal(t, "Echo Tool", tool.Name)
	assert.Equal(t, "echo_function", tool.FunctionName)
	assert.Equal(t, int64(10), tool.UsageCount)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMCPToolModel_ListByOrganization(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	defer mockDB.db.Close()

	model := models.NewMCPToolModel(mockDB)

	orgID := uuid.New()
	toolID1 := uuid.New()
	toolID2 := uuid.New()

	schema := map[string]interface{}{"type": "object"}
	schemaJSON, _ := json.Marshal(schema)

	// Expect the SELECT query (ordered by usage_count DESC)
	mock.ExpectQuery(`SELECT (.+) FROM mcp_tools WHERE organization_id = \$1 AND is_active = true ORDER BY usage_count DESC, created_at DESC`).
		WithArgs(orgID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description", "function_name", "schema", "category",
			"implementation_type", "endpoint_url", "timeout_seconds", "max_retries", "usage_count",
			"access_permissions", "is_active", "is_public", "metadata", "tags", "examples",
			"documentation", "created_at", "updated_at", "created_by", "server_id", "source_type",
			"last_discovered_at", "discovery_metadata",
		}).AddRow(
			toolID1, orgID, "Popular Tool", "Description 1", "popular_func", schemaJSON,
			types.ToolCategoryData, types.ToolImplementationInternal, "",
			30, 3, int64(15), []byte("{}"), true, false, []byte("{}"),
			pq.StringArray{"data"}, []byte("[]"), "Docs 1",
			time.Now(), time.Now(), nil, nil, "manual", nil, []byte("{}"),
		).AddRow(
			toolID2, orgID, "New Tool", "Description 2", "new_func", schemaJSON,
			types.ToolCategoryGeneral, types.ToolImplementationExternal, "https://api.example.com",
			60, 2, int64(3), []byte("{}"), true, true, []byte("{}"),
			pq.StringArray{"general"}, []byte("[]"), "Docs 2",
			time.Now(), time.Now(), nil, nil, "manual", nil, []byte("{}"),
		))

	tools, err := model.ListByOrganization(orgID, true)
	require.NoError(t, err)
	require.Len(t, tools, 2)

	// Should be ordered by usage count (descending)
	assert.Equal(t, toolID1, tools[0].ID)
	assert.Equal(t, "Popular Tool", tools[0].Name)
	assert.Equal(t, int64(15), tools[0].UsageCount)
	assert.Equal(t, types.ToolCategoryData, tools[0].Category)

	assert.Equal(t, toolID2, tools[1].ID)
	assert.Equal(t, "New Tool", tools[1].Name)
	assert.Equal(t, int64(3), tools[1].UsageCount)
	assert.Equal(t, types.ToolImplementationExternal, tools[1].ImplementationType)
	assert.True(t, tools[1].IsPublic)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMCPToolModel_ListByCategory(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	defer mockDB.db.Close()

	model := models.NewMCPToolModel(mockDB)

	orgID := uuid.New()
	toolID := uuid.New()

	schema := map[string]interface{}{"type": "object"}
	schemaJSON, _ := json.Marshal(schema)

	// Expect the SELECT query with category filter
	mock.ExpectQuery(`SELECT (.+) FROM mcp_tools WHERE organization_id = \$1 AND category = \$2 AND is_active = true ORDER BY usage_count DESC, created_at DESC`).
		WithArgs(orgID, types.ToolCategoryDev).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description", "function_name", "schema", "category",
			"implementation_type", "endpoint_url", "timeout_seconds", "max_retries", "usage_count",
			"access_permissions", "is_active", "is_public", "metadata", "tags", "examples",
			"documentation", "created_at", "updated_at", "created_by", "server_id", "source_type",
			"last_discovered_at", "discovery_metadata",
		}).AddRow(
			toolID, orgID, "Dev Tool", "A development tool", "dev_func", schemaJSON,
			types.ToolCategoryDev, types.ToolImplementationScript, "/scripts/dev.sh",
			45, 1, int64(25), []byte("{}"), true, false, []byte("{}"),
			pq.StringArray{"dev", "script"}, []byte("[]"), "Dev docs",
			time.Now(), time.Now(), nil, nil, "manual", nil, []byte("{}"),
		))

	tools, err := model.ListByCategory(orgID, types.ToolCategoryDev, true)
	require.NoError(t, err)
	require.Len(t, tools, 1)

	assert.Equal(t, toolID, tools[0].ID)
	assert.Equal(t, "Dev Tool", tools[0].Name)
	assert.Equal(t, types.ToolCategoryDev, tools[0].Category)
	assert.Equal(t, types.ToolImplementationScript, tools[0].ImplementationType)
	assert.Equal(t, int64(25), tools[0].UsageCount)
	assert.Equal(t, 45, tools[0].TimeoutSeconds)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMCPToolModel_ListPublicTools(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	defer mockDB.db.Close()

	model := models.NewMCPToolModel(mockDB)

	toolID := uuid.New()
	orgID := uuid.New()

	schema := map[string]interface{}{"type": "object"}
	schemaJSON, _ := json.Marshal(schema)

	// Expect the SELECT query for public tools
	mock.ExpectQuery(`SELECT (.+) FROM mcp_tools WHERE is_public = true AND is_active = true ORDER BY usage_count DESC, created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(50, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description", "function_name", "schema", "category",
			"implementation_type", "endpoint_url", "timeout_seconds", "max_retries", "usage_count",
			"access_permissions", "is_active", "is_public", "metadata", "tags", "examples",
			"documentation", "created_at", "updated_at", "created_by", "server_id", "source_type",
			"last_discovered_at", "discovery_metadata",
		}).AddRow(
			toolID, orgID, "Public Tool", "A public tool", "public_func", schemaJSON,
			types.ToolCategoryGeneral, types.ToolImplementationInternal, "",
			30, 3, int64(100), []byte("{}"), true, true, []byte("{}"),
			pq.StringArray{"public"}, []byte("[]"), "Public docs",
			time.Now(), time.Now(), nil, nil, "manual", nil, []byte("{}"),
		))

	tools, err := model.ListPublicTools(50, 0)
	require.NoError(t, err)
	require.Len(t, tools, 1)

	assert.Equal(t, toolID, tools[0].ID)
	assert.Equal(t, "Public Tool", tools[0].Name)
	assert.True(t, tools[0].IsPublic)
	assert.Equal(t, int64(100), tools[0].UsageCount)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMCPToolModel_IncrementUsageCount(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	defer mockDB.db.Close()

	model := models.NewMCPToolModel(mockDB)

	toolID := uuid.New()

	// Expect the UPDATE query to increment usage count
	mock.ExpectExec(`UPDATE mcp_tools SET usage_count = usage_count \+ 1 WHERE id = \$1`).
		WithArgs(toolID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := model.IncrementUsageCount(toolID)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMCPToolModel_Update(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	defer mockDB.db.Close()

	model := models.NewMCPToolModel(mockDB)

	toolID := uuid.New()
	orgID := uuid.New()

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"updated": map[string]interface{}{
				"type": "boolean",
			},
		},
	}

	tool := &models.MCPTool{
		ID:                 toolID,
		OrganizationID:     orgID,
		Name:               "Updated Tool",
		FunctionName:       "updated_function",
		Schema:             schema,
		Category:           types.ToolCategoryAI,
		ImplementationType: types.ToolImplementationWebhook,
		TimeoutSeconds:     60,
		MaxRetries:         5,
		IsPublic:           true,
		Metadata:           map[string]interface{}{"version": "2.0"},
		Tags:               pq.StringArray{"updated", "ai"},
		Examples:           []interface{}{map[string]interface{}{"updated": true}},
		EndpointURL:        sql.NullString{String: "https://webhook.example.com", Valid: true},
		Documentation:      sql.NullString{String: "Updated documentation", Valid: true},
	}

	// Expect the UPDATE query
	mock.ExpectExec(`UPDATE mcp_tools SET (.+) WHERE id = \$1`).
		WithArgs(toolID, "Updated Tool", driver.Value(nil), "updated_function",
			sqlmock.AnyArg(), types.ToolCategoryAI, types.ToolImplementationWebhook,
			"https://webhook.example.com", 60, 5, sqlmock.AnyArg(), false, true,
			sqlmock.AnyArg(), pq.StringArray{"updated", "ai"}, sqlmock.AnyArg(),
			"Updated documentation", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := model.Update(tool)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMCPToolModel_Delete(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	defer mockDB.db.Close()

	model := models.NewMCPToolModel(mockDB)

	toolID := uuid.New()

	// Expect the UPDATE query (soft delete)
	mock.ExpectExec(`UPDATE mcp_tools SET is_active = false WHERE id = \$1`).
		WithArgs(toolID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := model.Delete(toolID)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMCPToolModel_SearchTools(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	defer mockDB.db.Close()

	model := models.NewMCPToolModel(mockDB)

	orgID := uuid.New()
	toolID := uuid.New()

	schema := map[string]interface{}{"type": "object"}
	schemaJSON, _ := json.Marshal(schema)

	// Expect the search query
	mock.ExpectQuery(`SELECT (.+) FROM mcp_tools WHERE organization_id = \$1 AND is_active = true AND \((.+)\) ORDER BY usage_count DESC, created_at DESC LIMIT \$4 OFFSET \$5`).
		WithArgs(orgID, "%search%", "search", 10, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description", "function_name", "schema", "category",
			"implementation_type", "endpoint_url", "timeout_seconds", "max_retries", "usage_count",
			"access_permissions", "is_active", "is_public", "metadata", "tags", "examples",
			"documentation", "created_at", "updated_at", "created_by", "server_id", "source_type",
			"last_discovered_at", "discovery_metadata",
		}).AddRow(
			toolID, orgID, "Search Tool", "Tool for search", "search_func", schemaJSON,
			types.ToolCategoryGeneral, types.ToolImplementationInternal, "",
			30, 3, int64(7), []byte("{}"), true, false, []byte("{}"),
			pq.StringArray{"search", "utility"}, []byte("[]"), "Search docs",
			time.Now(), time.Now(), nil, nil, "manual", nil, []byte("{}"),
		))

	tools, err := model.SearchTools(orgID, "search", 10, 0)
	require.NoError(t, err)
	require.Len(t, tools, 1)

	assert.Equal(t, toolID, tools[0].ID)
	assert.Equal(t, "Search Tool", tools[0].Name)
	assert.Equal(t, "search_func", tools[0].FunctionName)
	assert.Contains(t, []string(tools[0].Tags), "search")

	assert.NoError(t, mock.ExpectationsWereMet())
}
