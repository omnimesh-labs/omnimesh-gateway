package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/auth"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/server/handlers"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/tests/helpers"
)

// AuthIntegrationTestSuite provides test setup for auth integration tests
type AuthIntegrationTestSuite struct {
	db         *sql.DB
	router     *gin.Engine
	authSvc    *auth.Service
	teardown   func()
	testOrgID  string
	testUserID string
}

// NewAuthIntegrationTestSuite creates a new auth integration test suite
func NewAuthIntegrationTestSuite(t *testing.T) *AuthIntegrationTestSuite {
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

	// Create test user with known credentials
	userID, err := helpers.CreateTestUserWithCredentials(testDB, orgID, "test@example.com", "testpassword123")
	require.NoError(t, err)

	// Setup auth service
	authConfig := &auth.Config{
		JWTSecret:          "test-secret-key-for-integration-tests",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		BCryptCost:         10, // Lower cost for faster tests
	}
	authSvc := auth.NewService(testDB, authConfig)

	// Setup JWT manager and middleware
	jwtManager := auth.NewJWTManager(authConfig.JWTSecret, authConfig.AccessTokenExpiry, authConfig.RefreshTokenExpiry)
	authMiddleware := auth.NewMiddleware(jwtManager, authSvc)

	// Setup router with auth handlers
	gin.SetMode(gin.TestMode)
	router := gin.New()
	authHandler := handlers.NewAuthHandler(authSvc)

	// Setup auth routes
	authRoutes := router.Group("/api/auth")
	{
		authRoutes.POST("/login", authHandler.Login)
		authRoutes.POST("/refresh", authHandler.RefreshToken)
		authRoutes.POST("/logout", authHandler.Logout)

		// Protected routes requiring authentication
		protected := authRoutes.Group("", authMiddleware.RequireAuth())
		{
			protected.GET("/profile", authHandler.GetProfile)
			protected.PUT("/profile", authHandler.UpdateProfile)
			protected.POST("/api-keys", authHandler.CreateAPIKey)
		}
	}

	return &AuthIntegrationTestSuite{
		db:         testDB,
		router:     router,
		authSvc:    authSvc,
		teardown:   teardown,
		testOrgID:  orgID,
		testUserID: userID,
	}
}

// Cleanup tears down the test suite
func (suite *AuthIntegrationTestSuite) Cleanup() {
	suite.teardown()
}

// TestLoginEndpointResponseStructure tests that login returns properly wrapped response
func TestLoginEndpointResponseStructure(t *testing.T) {
	suite := NewAuthIntegrationTestSuite(t)
	defer suite.Cleanup()

	// Create login request
	loginReq := types.LoginRequest{
		Email:    "test@example.com",
		Password: "testpassword123",
	}

	reqBody, err := json.Marshal(loginReq)
	require.NoError(t, err)

	// Make login request
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Validate successful login response
	data := helpers.ValidateSuccessResponse(t, w, http.StatusOK)
	helpers.ValidateLoginResponse(t, data)

	// Additional validation for specific test data
	user := data["user"].(map[string]interface{})
	assert.Equal(t, "test@example.com", user["email"], "User email should match")
}

// TestLoginEndpointErrorResponseStructure tests error response structure
func TestLoginEndpointErrorResponseStructure(t *testing.T) {
	suite := NewAuthIntegrationTestSuite(t)
	defer suite.Cleanup()

	// Create login request with invalid credentials
	loginReq := types.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	reqBody, err := json.Marshal(loginReq)
	require.NoError(t, err)

	// Make login request
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Validate error response structure
	apiError := helpers.ValidateErrorResponse(t, w, http.StatusUnauthorized)
	helpers.ValidateAuthenticationError(t, apiError, "UNAUTHORIZED")
}

// TestRefreshTokenResponseStructure tests refresh token endpoint response
func TestRefreshTokenResponseStructure(t *testing.T) {
	suite := NewAuthIntegrationTestSuite(t)
	defer suite.Cleanup()

	// First, login to get tokens
	loginTokens := suite.performLogin(t, "test@example.com", "testpassword123")

	// Create refresh request
	refreshReq := types.RefreshTokenRequest{
		RefreshToken: loginTokens["refresh_token"].(string),
	}

	reqBody, err := json.Marshal(refreshReq)
	require.NoError(t, err)

	// Make refresh request
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Validate successful refresh token response
	data := helpers.ValidateSuccessResponse(t, w, http.StatusOK)
	helpers.ValidateRefreshTokenResponse(t, data)
}

// TestGetProfileResponseStructure tests profile endpoint response
func TestGetProfileResponseStructure(t *testing.T) {
	suite := NewAuthIntegrationTestSuite(t)
	defer suite.Cleanup()

	// Login and get access token
	tokens := suite.performLogin(t, "test@example.com", "testpassword123")
	accessToken := tokens["access_token"].(string)

	// Make profile request
	req := httptest.NewRequest(http.MethodGet, "/api/auth/profile", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Validate successful profile response
	data := helpers.ValidateSuccessResponse(t, w, http.StatusOK)
	helpers.ValidateUserResponse(t, data)

	// Additional validation for specific test data
	assert.Equal(t, "test@example.com", data["email"], "User email should match")
}

// TestUpdateProfileResponseStructure tests profile update endpoint response
func TestUpdateProfileResponseStructure(t *testing.T) {
	suite := NewAuthIntegrationTestSuite(t)
	defer suite.Cleanup()

	// Login and get access token
	tokens := suite.performLogin(t, "test@example.com", "testpassword123")
	accessToken := tokens["access_token"].(string)

	// Create update request
	updateReq := types.UpdateUserRequest{
		Name: "Updated Test User",
	}

	reqBody, err := json.Marshal(updateReq)
	require.NoError(t, err)

	// Make update request
	req := httptest.NewRequest(http.MethodPut, "/api/auth/profile", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Validate successful profile update response
	data := helpers.ValidateSuccessResponse(t, w, http.StatusOK)
	helpers.ValidateUserResponse(t, data)

	// Additional validation for specific test data
	assert.Equal(t, "Updated Test User", data["name"], "User name should be updated")
}

// TestCreateAPIKeyResponseStructure tests API key creation endpoint response
func TestCreateAPIKeyResponseStructure(t *testing.T) {
	suite := NewAuthIntegrationTestSuite(t)
	defer suite.Cleanup()

	// Login as admin user and get access token
	tokens := suite.performLogin(t, "test@example.com", "testpassword123")
	accessToken := tokens["access_token"].(string)

	// Create API key request
	createReq := types.CreateAPIKeyRequest{
		Name: "Test API Key",
		Role: "user",
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	// Make create API key request
	req := httptest.NewRequest(http.MethodPost, "/api/auth/api-keys", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Validate successful API key creation response (201 Created or 200 OK)
	expectedStatus := http.StatusCreated
	if w.Code == http.StatusOK {
		expectedStatus = http.StatusOK
	}

	data := helpers.ValidateSuccessResponse(t, w, expectedStatus)
	helpers.ValidateAPIKeyResponse(t, data)
}

// TestResponseConsistencyAcrossEndpoints validates that all auth endpoints use consistent response wrapping
func TestResponseConsistencyAcrossEndpoints(t *testing.T) {
	suite := NewAuthIntegrationTestSuite(t)
	defer suite.Cleanup()

	testCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		headers        map[string]string
		expectedStatus int
		expectData     bool
	}{
		{
			name:           "Login Success",
			method:         http.MethodPost,
			path:           "/api/auth/login",
			body:           types.LoginRequest{Email: "test@example.com", Password: "testpassword123"},
			expectedStatus: http.StatusOK,
			expectData:     true,
		},
		{
			name:           "Login Error",
			method:         http.MethodPost,
			path:           "/api/auth/login",
			body:           types.LoginRequest{Email: "test@example.com", Password: "wrongpassword"},
			expectedStatus: http.StatusUnauthorized,
			expectData:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var reqBody []byte
			var err error

			if tc.body != nil {
				reqBody, err = json.Marshal(tc.body)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(tc.method, tc.path, bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Set additional headers if provided
			for key, value := range tc.headers {
				req.Header.Set(key, value)
			}

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			// Validate response consistency
			helpers.ValidateResponseConsistency(t, w, tc.expectData)
		})
	}
}

// Helper method to perform login and return tokens
func (suite *AuthIntegrationTestSuite) performLogin(t *testing.T, email, password string) map[string]interface{} {
	loginReq := types.LoginRequest{
		Email:    email,
		Password: password,
	}

	reqBody, err := json.Marshal(loginReq)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "Login should succeed")

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data, ok := response["data"].(map[string]interface{})
	require.True(t, ok, "Response should contain data")

	return data
}
