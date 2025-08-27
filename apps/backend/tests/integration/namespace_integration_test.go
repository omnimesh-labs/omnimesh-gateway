package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"mcp-gateway/apps/backend/internal/server/handlers"
	"mcp-gateway/apps/backend/internal/services"
	"mcp-gateway/apps/backend/internal/types"
	"mcp-gateway/apps/backend/tests/helpers"
)

// NamespaceIntegrationTestSuite tests the namespace feature end-to-end
type NamespaceIntegrationTestSuite struct {
	suite.Suite
	db       *sql.DB
	service  *services.NamespaceService
	handler  *handlers.NamespaceHandler
	router   *gin.Engine
	teardown func()
}

// SetupSuite runs once before all tests
func (suite *NamespaceIntegrationTestSuite) SetupSuite() {
	// Use test database helper
	testDB, teardown, err := helpers.SetupTestDatabase(suite.T())
	require.NoError(suite.T(), err)

	suite.db = testDB
	suite.teardown = teardown

	// Run migrations
	err = helpers.RunMigrations(testDB)
	require.NoError(suite.T(), err)

	// Initialize service and handler
	endpointService := services.NewEndpointService(testDB, "http://localhost:8080")
	suite.service = services.NewNamespaceService(testDB, endpointService)
	suite.handler = handlers.NewNamespaceHandler(suite.service)

	// Setup test router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	// Add middleware to set organization ID for tests
	suite.router.Use(func(c *gin.Context) {
		c.Set("organization_id", "00000000-0000-0000-0000-000000000001")
		c.Next()
	})

	// Register routes
	api := suite.router.Group("/api")
	namespaces := api.Group("/namespaces")
	{
		namespaces.POST("", suite.handler.CreateNamespace)
		namespaces.GET("", suite.handler.ListNamespaces)
		namespaces.GET("/:id", suite.handler.GetNamespace)
		namespaces.PUT("/:id", suite.handler.UpdateNamespace)
		namespaces.DELETE("/:id", suite.handler.DeleteNamespace)
		namespaces.POST("/:id/servers", suite.handler.AddServerToNamespace)
		namespaces.DELETE("/:id/servers/:server_id", suite.handler.RemoveServerFromNamespace)
		namespaces.PUT("/:id/servers/:server_id/status", suite.handler.UpdateServerStatus)
		namespaces.GET("/:id/tools", suite.handler.GetNamespaceTools)
		namespaces.PUT("/:id/tools/:server_id/:tool_name/status", suite.handler.UpdateToolStatus)
		namespaces.POST("/:id/execute", suite.handler.ExecuteNamespaceTool)
	}
}

// TearDownSuite runs once after all tests
func (suite *NamespaceIntegrationTestSuite) TearDownSuite() {
	if suite.teardown != nil {
		suite.teardown()
	}
}

// SetupTest runs before each test
func (suite *NamespaceIntegrationTestSuite) SetupTest() {
	// Clean up database before each test
	helpers.CleanDatabase(suite.T(), suite.db)

	// Create test organization
	_, err := helpers.CreateTestOrganization(suite.db)
	require.NoError(suite.T(), err)
}

// TestCreateNamespace tests creating a new namespace
func (suite *NamespaceIntegrationTestSuite) TestCreateNamespace() {
	req := types.CreateNamespaceRequest{
		Name:        "test-namespace",
		Description: "Test namespace description",
	}

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/api/namespaces", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	suite.router.ServeHTTP(w, httpReq)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response types.Namespace
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test-namespace", response.Name)
	assert.Equal(suite.T(), "Test namespace description", response.Description)
	assert.NotEmpty(suite.T(), response.ID)
	assert.True(suite.T(), response.IsActive)
}

// TestListNamespaces tests listing namespaces
func (suite *NamespaceIntegrationTestSuite) TestListNamespaces() {
	// Create test namespaces
	ctx := context.Background()
	orgID := "00000000-0000-0000-0000-000000000001"

	_, err := suite.service.CreateNamespace(ctx, types.CreateNamespaceRequest{
		OrganizationID: orgID,
		Name:           "namespace-1",
		Description:    "First namespace",
	})
	require.NoError(suite.T(), err)

	_, err = suite.service.CreateNamespace(ctx, types.CreateNamespaceRequest{
		OrganizationID: orgID,
		Name:           "namespace-2",
		Description:    "Second namespace",
	})
	require.NoError(suite.T(), err)

	// List namespaces
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/namespaces", nil)

	suite.router.ServeHTTP(w, httpReq)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(2), response["total"])

	namespaces := response["namespaces"].([]interface{})
	assert.Len(suite.T(), namespaces, 2)
}

// TestGetNamespace tests retrieving a single namespace
func (suite *NamespaceIntegrationTestSuite) TestGetNamespace() {
	// Create a namespace
	ctx := context.Background()
	namespace, err := suite.service.CreateNamespace(ctx, types.CreateNamespaceRequest{
		OrganizationID: "00000000-0000-0000-0000-000000000001",
		Name:           "test-namespace",
		Description:    "Test namespace",
	})
	require.NoError(suite.T(), err)

	// Get the namespace
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/namespaces/"+namespace.ID, nil)

	suite.router.ServeHTTP(w, httpReq)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response types.Namespace
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), namespace.ID, response.ID)
	assert.Equal(suite.T(), "test-namespace", response.Name)
}

// TestUpdateNamespace tests updating a namespace
func (suite *NamespaceIntegrationTestSuite) TestUpdateNamespace() {
	// Create a namespace
	ctx := context.Background()
	namespace, err := suite.service.CreateNamespace(ctx, types.CreateNamespaceRequest{
		OrganizationID: "00000000-0000-0000-0000-000000000001",
		Name:           "old-name",
		Description:    "Old description",
	})
	require.NoError(suite.T(), err)

	// Update the namespace
	isActive := false
	updateReq := types.UpdateNamespaceRequest{
		Name:        "new-name",
		Description: "New description",
		IsActive:    &isActive,
	}

	body, _ := json.Marshal(updateReq)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("PUT", "/api/namespaces/"+namespace.ID, bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	suite.router.ServeHTTP(w, httpReq)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response types.Namespace
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "new-name", response.Name)
	assert.Equal(suite.T(), "New description", response.Description)
	assert.False(suite.T(), response.IsActive)
}

// TestDeleteNamespace tests deleting a namespace
func (suite *NamespaceIntegrationTestSuite) TestDeleteNamespace() {
	// Create a namespace
	ctx := context.Background()
	namespace, err := suite.service.CreateNamespace(ctx, types.CreateNamespaceRequest{
		OrganizationID: "00000000-0000-0000-0000-000000000001",
		Name:           "to-delete",
		Description:    "Will be deleted",
	})
	require.NoError(suite.T(), err)

	// Delete the namespace
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("DELETE", "/api/namespaces/"+namespace.ID, nil)

	suite.router.ServeHTTP(w, httpReq)

	assert.Equal(suite.T(), http.StatusNoContent, w.Code)

	// Verify namespace is deleted
	_, err = suite.service.GetNamespace(ctx, namespace.ID)
	assert.Error(suite.T(), err)
}

// TestNamespaceWithServers tests namespace operations with servers
func (suite *NamespaceIntegrationTestSuite) TestNamespaceWithServers() {
	// Create a namespace
	ctx := context.Background()
	namespace, err := suite.service.CreateNamespace(ctx, types.CreateNamespaceRequest{
		OrganizationID: "00000000-0000-0000-0000-000000000001",
		Name:           "namespace-with-servers",
		Description:    "Namespace with servers",
	})
	require.NoError(suite.T(), err)

	// Create an MCP server in the database first
	serverID := "11111111-1111-1111-1111-111111111111"
	_, err = suite.db.Exec(`
		INSERT INTO mcp_servers (id, organization_id, name, description, protocol, url, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, serverID, "00000000-0000-0000-0000-000000000001", "Test Server", "Test server description", "http", "http://localhost:8080", true)
	require.NoError(suite.T(), err)

	// Add server to namespace
	addReq := types.AddServerToNamespaceRequest{
		ServerID: serverID,
		Priority: 1,
	}

	body, _ := json.Marshal(addReq)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/api/namespaces/"+namespace.ID+"/servers", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	suite.router.ServeHTTP(w, httpReq)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Get namespace and verify server is included
	w = httptest.NewRecorder()
	httpReq, _ = http.NewRequest("GET", "/api/namespaces/"+namespace.ID, nil)

	suite.router.ServeHTTP(w, httpReq)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var nsResponse types.Namespace
	err = json.Unmarshal(w.Body.Bytes(), &nsResponse)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), nsResponse.Servers, 1)
	assert.Equal(suite.T(), serverID, nsResponse.Servers[0].ServerID)

	// Remove server from namespace
	w = httptest.NewRecorder()
	httpReq, _ = http.NewRequest("DELETE", "/api/namespaces/"+namespace.ID+"/servers/"+serverID, nil)

	suite.router.ServeHTTP(w, httpReq)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Verify server is removed
	w = httptest.NewRecorder()
	httpReq, _ = http.NewRequest("GET", "/api/namespaces/"+namespace.ID, nil)

	suite.router.ServeHTTP(w, httpReq)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var nsResponseAfterDelete types.Namespace
	err = json.Unmarshal(w.Body.Bytes(), &nsResponseAfterDelete)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), nsResponseAfterDelete.Servers, 0)
}

// TestExecuteNamespaceTool tests executing a tool through a namespace
func (suite *NamespaceIntegrationTestSuite) TestExecuteNamespaceTool() {
	// Create a namespace
	ctx := context.Background()
	namespace, err := suite.service.CreateNamespace(ctx, types.CreateNamespaceRequest{
		OrganizationID: "00000000-0000-0000-0000-000000000001",
		Name:           "namespace-with-tools",
		Description:    "Namespace with tools",
	})
	require.NoError(suite.T(), err)

	// Execute a tool (will return mock response for now)
	execReq := types.ExecuteNamespaceToolRequest{
		Tool:      "server1__tool1",
		Arguments: map[string]interface{}{"arg": "value"},
	}

	body, _ := json.Marshal(execReq)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/api/namespaces/"+namespace.ID+"/execute", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	suite.router.ServeHTTP(w, httpReq)

	// The actual execution will fail since we don't have real servers
	// but we're testing the endpoint exists and handles the request
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response types.NamespaceToolResult
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success) // Expected to fail since no real server
}

// TestConcurrentNamespaceOperations tests concurrent operations on namespaces
func (suite *NamespaceIntegrationTestSuite) TestConcurrentNamespaceOperations() {
	// Create multiple namespaces concurrently
	ctx := context.Background()
	numNamespaces := 10
	done := make(chan bool, numNamespaces)

	for i := 0; i < numNamespaces; i++ {
		go func(index int) {
			_, err := suite.service.CreateNamespace(ctx, types.CreateNamespaceRequest{
				OrganizationID: "00000000-0000-0000-0000-000000000001",
				Name:           fmt.Sprintf("concurrent-namespace-%d", index),
				Description:    fmt.Sprintf("Concurrent namespace %d", index),
			})
			assert.NoError(suite.T(), err)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numNamespaces; i++ {
		<-done
	}

	// Verify all namespaces were created
	namespaces, err := suite.service.ListNamespaces(ctx, "00000000-0000-0000-0000-000000000001")
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), namespaces, numNamespaces)
}

// TestNamespaceNameValidation tests namespace name validation
func (suite *NamespaceIntegrationTestSuite) TestNamespaceNameValidation() {
	testCases := []struct {
		name        string
		nsName      string
		expectError bool
	}{
		{"valid name", "valid-namespace-123", false},
		{"valid with underscores", "valid_namespace", false},
		{"too short", "ab", true},
		{"too long", "this-is-a-very-long-namespace-name-that-exceeds-fifty-characters-limit", true},
		{"with spaces", "invalid namespace", true},
		{"with special chars", "invalid@namespace!", true},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			req := types.CreateNamespaceRequest{
				Name:        tc.nsName,
				Description: "Test",
			}

			body, _ := json.Marshal(req)
			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("POST", "/api/namespaces", bytes.NewBuffer(body))
			httpReq.Header.Set("Content-Type", "application/json")

			suite.router.ServeHTTP(w, httpReq)

			if tc.expectError {
				assert.NotEqual(t, http.StatusCreated, w.Code)
			} else {
				assert.Equal(t, http.StatusCreated, w.Code)
			}
		})
	}
}

// TestRunSuite runs the test suite
func TestNamespaceIntegrationSuite(t *testing.T) {
	suite.Run(t, new(NamespaceIntegrationTestSuite))
}
