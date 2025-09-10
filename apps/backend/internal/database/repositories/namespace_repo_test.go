package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamespaceRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewNamespaceRepository(sqlxDB)

	ns := &types.Namespace{
		ID:             uuid.New().String(),
		OrganizationID: "org-123",
		Name:           "test-namespace",
		Description:    "Test namespace",
		IsActive:       true,
		Metadata:       map[string]interface{}{"key": "value"},
	}

	mock.ExpectQuery(`INSERT INTO namespaces`).
		WithArgs(ns.ID, ns.OrganizationID, ns.Name, ns.Description, nil, ns.IsActive, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).
			AddRow(time.Now(), time.Now()))

	err = repo.Create(context.Background(), ns)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNamespaceRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewNamespaceRepository(sqlxDB)

	nsID := uuid.New().String()
	expectedNS := &types.Namespace{
		ID:             nsID,
		OrganizationID: "org-123",
		Name:           "test-namespace",
		Description:    "Test namespace",
		IsActive:       true,
	}

	mock.ExpectQuery(`SELECT .+ FROM namespaces WHERE id = \$1`).
		WithArgs(nsID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description",
			"created_at", "updated_at", "created_by", "is_active", "metadata",
		}).AddRow(
			expectedNS.ID, expectedNS.OrganizationID, expectedNS.Name, expectedNS.Description,
			time.Now(), time.Now(), nil, expectedNS.IsActive, []byte("{}"),
		))

	result, err := repo.GetByID(context.Background(), nsID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	if result != nil {
		assert.Equal(t, expectedNS.ID, result.ID)
		assert.Equal(t, expectedNS.Name, result.Name)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNamespaceRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewNamespaceRepository(sqlxDB)

	orgID := "org-123"

	mock.ExpectQuery(`SELECT .+ FROM namespaces WHERE organization_id = \$1`).
		WithArgs(orgID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description",
			"created_at", "updated_at", "created_by", "is_active", "metadata",
		}).
			AddRow("ns-1", orgID, "namespace-1", "First namespace",
				time.Now(), time.Now(), nil, true, []byte("{}")).
			AddRow("ns-2", orgID, "namespace-2", "Second namespace",
				time.Now(), time.Now(), nil, true, []byte("{}")))

	result, err := repo.List(context.Background(), orgID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "namespace-1", result[0].Name)
	assert.Equal(t, "namespace-2", result[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNamespaceRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewNamespaceRepository(sqlxDB)

	ns := &types.Namespace{
		ID:             uuid.New().String(),
		OrganizationID: "org-123",
		Name:           "updated-namespace",
		Description:    "Updated description",
		IsActive:       false,
		Metadata:       map[string]interface{}{"updated": true},
	}

	mock.ExpectExec(`UPDATE namespaces SET`).
		WithArgs(ns.ID, ns.Name, ns.Description, ns.IsActive, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Update(context.Background(), ns)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNamespaceRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewNamespaceRepository(sqlxDB)

	nsID := uuid.New().String()

	mock.ExpectExec(`DELETE FROM namespaces WHERE id = \$1`).
		WithArgs(nsID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(context.Background(), nsID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNamespaceRepository_AddServer(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewNamespaceRepository(sqlxDB)

	namespaceID := uuid.New().String()
	serverID := uuid.New().String()
	priority := 1

	mock.ExpectExec(`INSERT INTO namespace_server_mappings`).
		WithArgs(sqlmock.AnyArg(), namespaceID, serverID, "ACTIVE", priority).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.AddServer(context.Background(), namespaceID, serverID, priority)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNamespaceRepository_RemoveServer(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewNamespaceRepository(sqlxDB)

	namespaceID := uuid.New().String()
	serverID := uuid.New().String()

	mock.ExpectExec(`DELETE FROM namespace_server_mappings`).
		WithArgs(namespaceID, serverID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.RemoveServer(context.Background(), namespaceID, serverID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNamespaceRepository_GetServers(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewNamespaceRepository(sqlxDB)

	namespaceID := uuid.New().String()

	mock.ExpectQuery(`SELECT .+ FROM namespace_server_mappings`).
		WithArgs(namespaceID).
		WillReturnRows(sqlmock.NewRows([]string{
			"server_id", "server_name", "status", "priority", "created_at",
		}).
			AddRow("srv-1", "server-1", "ACTIVE", 0, time.Now()).
			AddRow("srv-2", "server-2", "INACTIVE", 1, time.Now()))

	servers, err := repo.GetServers(context.Background(), namespaceID)
	assert.NoError(t, err)
	assert.Len(t, servers, 2)
	assert.Equal(t, "server-1", servers[0].ServerName)
	assert.Equal(t, "ACTIVE", servers[0].Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNamespaceRepository_SetToolStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewNamespaceRepository(sqlxDB)

	namespaceID := uuid.New().String()
	serverID := uuid.New().String()
	toolName := "test-tool"
	status := "INACTIVE"

	mock.ExpectExec(`INSERT INTO namespace_tool_mappings`).
		WithArgs(sqlmock.AnyArg(), namespaceID, serverID, toolName, status).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.SetToolStatus(context.Background(), namespaceID, serverID, toolName, status)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
