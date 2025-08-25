package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mcp-gateway/apps/backend/internal/server/handlers"
	"mcp-gateway/apps/backend/internal/services"
	"mcp-gateway/apps/backend/internal/types"
	"mcp-gateway/apps/backend/tests/helpers"
)

// IntegrationTestSuite provides a test setup for integration tests
type IntegrationTestSuite struct {
	db              *sql.DB
	Router          *gin.Engine
	AuthToken       string
	TestNamespaceID string
	teardown        func()
}

// NewIntegrationTestSuite creates a new integration test suite
func NewIntegrationTestSuite(t *testing.T) *IntegrationTestSuite {
	// Setup test database
	testDB, teardown, err := helpers.SetupTestDatabase(t)
	require.NoError(t, err)

	// Run migrations
	err = helpers.RunMigrations(testDB)
	require.NoError(t, err)

	// Clean database
	helpers.CleanDatabase(t, testDB)

	// Create test organization
	orgID, err := helpers.CreateTestOrganization(testDB)
	require.NoError(t, err)

	// Create test user
	userID, err := helpers.CreateTestUser(testDB, orgID)
	require.NoError(t, err)

	// Initialize services
	namespaceService := services.NewNamespaceService(testDB)
	endpointService := services.NewEndpointService(testDB, "http://localhost:8080")

	// Initialize handlers
	namespaceHandler := handlers.NewNamespaceHandler(namespaceService)
	endpointHandler := handlers.NewEndpointHandler(endpointService)

	// Setup test router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add middleware to set organization ID and user ID for tests
	router.Use(func(c *gin.Context) {
		c.Set("organization_id", orgID)
		c.Set("user_id", userID)
		c.Next()
	})

	// Register routes
	api := router.Group("/api")

	// Namespace routes
	namespaces := api.Group("/namespaces")
	{
		namespaces.POST("", namespaceHandler.CreateNamespace)
		namespaces.GET("", namespaceHandler.ListNamespaces)
		namespaces.GET("/:id", namespaceHandler.GetNamespace)
		namespaces.PUT("/:id", namespaceHandler.UpdateNamespace)
		namespaces.DELETE("/:id", namespaceHandler.DeleteNamespace)
	}

	// Endpoint routes
	endpoints := api.Group("/endpoints")
	{
		endpoints.GET("", endpointHandler.ListEndpoints)
		endpoints.POST("", endpointHandler.CreateEndpoint)
		endpoints.GET("/:id", endpointHandler.GetEndpoint)
		endpoints.PUT("/:id", endpointHandler.UpdateEndpoint)
		endpoints.DELETE("/:id", endpointHandler.DeleteEndpoint)
	}

	// Public endpoint routes
	public := api.Group("/public")
	{
		public.GET("/endpoints", endpointHandler.ListEndpoints)
		public.GET("/endpoints/:endpoint_name/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "healthy"})
		})
	}

	return &IntegrationTestSuite{
		db:       testDB,
		Router:   router,
		AuthToken: "test-token", // Mock token for tests
		teardown: teardown,
	}
}

// Cleanup cleans up the test suite
func (s *IntegrationTestSuite) Cleanup() {
	if s.teardown != nil {
		s.teardown()
	}
}

func TestEndpointIntegration(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	suite := NewIntegrationTestSuite(t)
	defer suite.Cleanup()

	router := suite.Router
	authToken := suite.AuthToken

	t.Run("Create Namespace First", func(t *testing.T) {
		// Create a namespace for testing
		createNamespaceReq := types.CreateNamespaceRequest{
			Name:        "test-namespace",
			Description: "Test namespace for endpoints",
		}

		body, _ := json.Marshal(createNamespaceReq)
		req := httptest.NewRequest("POST", "/api/namespaces", bytes.NewBuffer(body))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var namespace types.Namespace
		err := json.Unmarshal(w.Body.Bytes(), &namespace)
		require.NoError(t, err)
		suite.TestNamespaceID = namespace.ID
	})

	t.Run("Create Endpoint", func(t *testing.T) {
		createEndpointReq := types.CreateEndpointRequest{
			NamespaceID:        suite.TestNamespaceID,
			Name:               "test-endpoint",
			Description:        "Test endpoint",
			EnableAPIKeyAuth:   true,
			EnablePublicAccess: false,
			RateLimitRequests:  100,
			RateLimitWindow:    60,
			AllowedOrigins:     []string{"*"},
			AllowedMethods:     []string{"GET", "POST"},
		}

		body, _ := json.Marshal(createEndpointReq)
		req := httptest.NewRequest("POST", "/api/endpoints", bytes.NewBuffer(body))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var endpoint types.Endpoint
		err := json.Unmarshal(w.Body.Bytes(), &endpoint)
		require.NoError(t, err)

		assert.Equal(t, "test-endpoint", endpoint.Name)
		assert.NotEmpty(t, endpoint.ID)
		assert.NotNil(t, endpoint.URLs)
		assert.Contains(t, endpoint.URLs.SSE, "/api/public/endpoints/test-endpoint/sse")
	})

	t.Run("List Endpoints", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/endpoints", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		endpoints, ok := response["endpoints"].([]interface{})
		assert.True(t, ok)
		assert.GreaterOrEqual(t, len(endpoints), 1)
	})

	t.Run("Access Public Endpoint Without Auth", func(t *testing.T) {
		// Create a public endpoint
		createEndpointReq := types.CreateEndpointRequest{
			NamespaceID:        suite.TestNamespaceID,
			Name:               "public-endpoint",
			Description:        "Public test endpoint",
			EnablePublicAccess: true,
			RateLimitRequests:  10,
			RateLimitWindow:    60,
		}

		body, _ := json.Marshal(createEndpointReq)
		req := httptest.NewRequest("POST", "/api/endpoints", bytes.NewBuffer(body))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		// Now try to access the public endpoint without auth
		req = httptest.NewRequest("GET", "/api/public/endpoints/public-endpoint/health", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should NOT be accessible without auth (requires authentication)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("List Public Endpoints", func(t *testing.T) {
		// This endpoint now requires authentication
		req := httptest.NewRequest("GET", "/api/public/endpoints", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		endpoints, ok := response["endpoints"].([]interface{})
		assert.True(t, ok)
		assert.GreaterOrEqual(t, len(endpoints), 1)
	})
}
