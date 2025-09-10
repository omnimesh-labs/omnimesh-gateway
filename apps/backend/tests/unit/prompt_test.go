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

func TestMCPPromptModel_Create(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	defer mockDB.db.Close()

	model := models.NewMCPPromptModel(mockDB)

	orgID := uuid.New()
	userID := uuid.New()

	parameters := []interface{}{
		map[string]interface{}{
			"name":        "input",
			"type":        "string",
			"required":    true,
			"description": "Input text",
		},
	}

	prompt := &models.MCPPrompt{
		OrganizationID: orgID,
		Name:           "Test Prompt",
		PromptTemplate: "Please analyze: {{input}}",
		Parameters:     parameters,
		Category:       types.PromptCategoryGeneral,
		UsageCount:     0,
		IsActive:       true,
		Metadata:       map[string]interface{}{"version": "1.0"},
		Tags:           pq.StringArray{"test", "analysis"},
		Description:    sql.NullString{String: "Test prompt description", Valid: true},
		CreatedBy:      uuid.NullUUID{UUID: userID, Valid: true},
	}

	// Expect the INSERT query
	mock.ExpectExec(`INSERT INTO mcp_prompts`).
		WithArgs(sqlmock.AnyArg(), orgID, "Test Prompt", "Test prompt description",
			"Please analyze: {{input}}", sqlmock.AnyArg(), types.PromptCategoryGeneral,
			int64(0), true, sqlmock.AnyArg(), pq.StringArray{"test", "analysis"}, userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := model.Create(prompt)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, prompt.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMCPPromptModel_GetByID(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	defer mockDB.db.Close()

	model := models.NewMCPPromptModel(mockDB)

	promptID := uuid.New()
	orgID := uuid.New()
	userID := uuid.New()

	metadata := map[string]interface{}{"version": "1.0"}
	metadataJSON, _ := json.Marshal(metadata)

	parameters := []interface{}{
		map[string]interface{}{
			"name":        "input",
			"type":        "string",
			"required":    true,
			"description": "Input text",
		},
	}
	parametersJSON, _ := json.Marshal(parameters)

	// Expect the SELECT query
	mock.ExpectQuery(`SELECT (.+) FROM mcp_prompts WHERE id = \$1`).
		WithArgs(promptID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description", "prompt_template", "parameters",
			"category", "usage_count", "is_active", "metadata", "tags",
			"created_at", "updated_at", "created_by",
		}).AddRow(
			promptID, orgID, "Test Prompt", "Test description", "Please analyze: {{input}}",
			parametersJSON, types.PromptCategoryGeneral, int64(5), true, metadataJSON,
			pq.StringArray{"test", "analysis"}, time.Now(), time.Now(), userID,
		))

	prompt, err := model.GetByID(promptID)
	require.NoError(t, err)
	require.NotNil(t, prompt)

	assert.Equal(t, promptID, prompt.ID)
	assert.Equal(t, orgID, prompt.OrganizationID)
	assert.Equal(t, "Test Prompt", prompt.Name)
	assert.Equal(t, "Test description", prompt.Description.String)
	assert.Equal(t, "Please analyze: {{input}}", prompt.PromptTemplate)
	assert.Equal(t, types.PromptCategoryGeneral, prompt.Category)
	assert.Equal(t, int64(5), prompt.UsageCount)
	assert.True(t, prompt.IsActive)
	assert.Equal(t, metadata, prompt.Metadata)
	assert.Equal(t, parameters, prompt.Parameters)
	assert.Equal(t, []string{"test", "analysis"}, []string(prompt.Tags))

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMCPPromptModel_ListByOrganization(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	defer mockDB.db.Close()

	model := models.NewMCPPromptModel(mockDB)

	orgID := uuid.New()
	promptID1 := uuid.New()
	promptID2 := uuid.New()

	metadata := map[string]interface{}{"version": "1.0"}
	metadataJSON, _ := json.Marshal(metadata)

	parameters := []interface{}{}
	parametersJSON, _ := json.Marshal(parameters)

	// Expect the SELECT query (ordered by usage_count DESC)
	mock.ExpectQuery(`SELECT (.+) FROM mcp_prompts WHERE organization_id = \$1 AND is_active = true ORDER BY usage_count DESC, created_at DESC`).
		WithArgs(orgID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description", "prompt_template", "parameters",
			"category", "usage_count", "is_active", "metadata", "tags",
			"created_at", "updated_at", "created_by",
		}).AddRow(
			promptID1, orgID, "Popular Prompt", "Description 1", "Template 1",
			parametersJSON, types.PromptCategoryCoding, int64(10), true, metadataJSON,
			pq.StringArray{"coding"}, time.Now(), time.Now(), nil,
		).AddRow(
			promptID2, orgID, "New Prompt", "Description 2", "Template 2",
			parametersJSON, types.PromptCategoryGeneral, int64(2), true, metadataJSON,
			pq.StringArray{"general"}, time.Now(), time.Now(), nil,
		))

	prompts, err := model.ListByOrganization(orgID, true)
	require.NoError(t, err)
	require.Len(t, prompts, 2)

	// Should be ordered by usage count (descending)
	assert.Equal(t, promptID1, prompts[0].ID)
	assert.Equal(t, "Popular Prompt", prompts[0].Name)
	assert.Equal(t, int64(10), prompts[0].UsageCount)

	assert.Equal(t, promptID2, prompts[1].ID)
	assert.Equal(t, "New Prompt", prompts[1].Name)
	assert.Equal(t, int64(2), prompts[1].UsageCount)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMCPPromptModel_ListByCategory(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	defer mockDB.db.Close()

	model := models.NewMCPPromptModel(mockDB)

	orgID := uuid.New()
	promptID := uuid.New()

	metadata := map[string]interface{}{"version": "1.0"}
	metadataJSON, _ := json.Marshal(metadata)

	parameters := []interface{}{}
	parametersJSON, _ := json.Marshal(parameters)

	// Expect the SELECT query with category filter
	mock.ExpectQuery(`SELECT (.+) FROM mcp_prompts WHERE organization_id = \$1 AND category = \$2 AND is_active = true ORDER BY usage_count DESC, created_at DESC`).
		WithArgs(orgID, types.PromptCategoryCoding).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description", "prompt_template", "parameters",
			"category", "usage_count", "is_active", "metadata", "tags",
			"created_at", "updated_at", "created_by",
		}).AddRow(
			promptID, orgID, "Coding Prompt", "A coding prompt", "Write code for: {{task}}",
			parametersJSON, types.PromptCategoryCoding, int64(15), true, metadataJSON,
			pq.StringArray{"coding", "development"}, time.Now(), time.Now(), nil,
		))

	prompts, err := model.ListByCategory(orgID, types.PromptCategoryCoding, true)
	require.NoError(t, err)
	require.Len(t, prompts, 1)

	assert.Equal(t, promptID, prompts[0].ID)
	assert.Equal(t, "Coding Prompt", prompts[0].Name)
	assert.Equal(t, types.PromptCategoryCoding, prompts[0].Category)
	assert.Equal(t, int64(15), prompts[0].UsageCount)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMCPPromptModel_IncrementUsageCount(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	defer mockDB.db.Close()

	model := models.NewMCPPromptModel(mockDB)

	promptID := uuid.New()

	// Expect the UPDATE query to increment usage count
	mock.ExpectExec(`UPDATE mcp_prompts SET usage_count = usage_count \+ 1 WHERE id = \$1`).
		WithArgs(promptID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := model.IncrementUsageCount(promptID)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMCPPromptModel_Update(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	defer mockDB.db.Close()

	model := models.NewMCPPromptModel(mockDB)

	promptID := uuid.New()
	orgID := uuid.New()

	parameters := []interface{}{
		map[string]interface{}{
			"name":        "task",
			"type":        "string",
			"required":    true,
			"description": "The task to complete",
		},
	}

	prompt := &models.MCPPrompt{
		ID:             promptID,
		OrganizationID: orgID,
		Name:           "Updated Prompt",
		PromptTemplate: "Complete the task: {{task}}",
		Parameters:     parameters,
		Category:       types.PromptCategoryCoding,
		Metadata:       map[string]interface{}{"version": "2.0"},
		Tags:           pq.StringArray{"updated", "coding"},
	}

	// Expect the UPDATE query
	mock.ExpectExec(`UPDATE mcp_prompts SET (.+) WHERE id = \$1`).
		WithArgs(promptID, "Updated Prompt", driver.Value(nil), "Complete the task: {{task}}",
			sqlmock.AnyArg(), types.PromptCategoryCoding, sqlmock.AnyArg(),
			pq.StringArray{"updated", "coding"}, false).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := model.Update(prompt)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMCPPromptModel_Delete(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	defer mockDB.db.Close()

	model := models.NewMCPPromptModel(mockDB)

	promptID := uuid.New()

	// Expect the UPDATE query (soft delete)
	mock.ExpectExec(`UPDATE mcp_prompts SET is_active = false WHERE id = \$1`).
		WithArgs(promptID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := model.Delete(promptID)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}
