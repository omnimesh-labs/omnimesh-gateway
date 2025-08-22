package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mcp-gateway/apps/backend/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockService is a mock implementation of the auth service
type MockService struct {
	mock.Mock
}

func (m *MockService) GetUserByID(userID string) (*types.User, error) {
	args := m.Called(userID)
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockService) ValidateAPIKey(apiKey string) (*types.APIKey, error) {
	args := m.Called(apiKey)
	return args.Get(0).(*types.APIKey), args.Error(1)
}

func setupTestMiddleware() (*Middleware, *MockService, *JWTManager) {
	config := &Config{
		JWTSecret:          "test-secret-key-for-testing",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
	}

	jwtManager := NewJWTManager(config.JWTSecret, config.AccessTokenExpiry, config.RefreshTokenExpiry)
	mockService := &MockService{}
	middleware := NewMiddlewareWithInterface(jwtManager, mockService)

	return middleware, mockService, jwtManager
}

func createTestUser() *types.User {
	return &types.User{
		ID:             uuid.New().String(),
		Email:          "test@example.com",
		Name:           "Test User",
		OrganizationID: uuid.New().String(),
		Role:           types.RoleUser,
		IsActive:       true,
	}
}

func TestMiddleware_RequireAuth_Success(t *testing.T) {
	middleware, mockService, jwtManager := setupTestMiddleware()
	user := createTestUser()

	// Generate valid access token
	token, err := jwtManager.GenerateAccessToken(user)
	assert.NoError(t, err)

	// Mock service call
	mockService.On("GetUserByID", user.ID).Return(user, nil)

	// Setup gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	// Create a test handler to verify context is set
	handlerCalled := false
	testHandler := func(c *gin.Context) {
		handlerCalled = true
		// Verify user context is set
		contextUser, exists := c.Get("user")
		assert.True(t, exists)
		assert.Equal(t, user.ID, contextUser.(*types.User).ID)

		userID, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, user.ID, userID)

		orgID, exists := c.Get("organization_id")
		assert.True(t, exists)
		assert.Equal(t, user.OrganizationID, orgID)

		role, exists := c.Get("role")
		assert.True(t, exists)
		assert.Equal(t, user.Role, role)

		c.Status(http.StatusOK)
	}

	// Execute middleware and handler
	middleware.RequireAuth()(c)
	if !c.IsAborted() {
		testHandler(c)
	}

	assert.False(t, c.IsAborted())
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestMiddleware_RequireAuth_MissingToken(t *testing.T) {
	middleware, _, _ := setupTestMiddleware()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	middleware.RequireAuth()(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMiddleware_RequireAuth_InvalidToken(t *testing.T) {
	middleware, _, _ := setupTestMiddleware()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid.jwt.token")

	middleware.RequireAuth()(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMiddleware_RequireAuth_RefreshTokenRejected(t *testing.T) {
	middleware, _, jwtManager := setupTestMiddleware()
	user := createTestUser()

	// Generate refresh token (should be rejected for API access)
	token, err := jwtManager.GenerateRefreshToken(user)
	assert.NoError(t, err)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	middleware.RequireAuth()(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMiddleware_RequireAuth_InactiveUser(t *testing.T) {
	middleware, mockService, jwtManager := setupTestMiddleware()
	user := createTestUser()
	user.IsActive = false // Set user as inactive

	token, err := jwtManager.GenerateAccessToken(user)
	assert.NoError(t, err)

	mockService.On("GetUserByID", user.ID).Return(user, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	middleware.RequireAuth()(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}

func TestMiddleware_RequireRole_Success(t *testing.T) {
	middleware, _, _ := setupTestMiddleware()
	user := createTestUser()
	user.Role = types.RoleAdmin

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// Set user context (normally done by RequireAuth)
	c.Set("role", user.Role)

	handlerCalled := false
	testHandler := func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusOK)
	}

	middleware.RequireRole(types.RoleUser)(c)
	if !c.IsAborted() {
		testHandler(c)
	}

	assert.False(t, c.IsAborted())
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_RequireRole_InsufficientPermissions(t *testing.T) {
	middleware, _, _ := setupTestMiddleware()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// Set user context with insufficient role
	c.Set("role", types.RoleViewer)

	middleware.RequireRole(types.RoleAdmin)(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestMiddleware_RequireAPIKey_Success(t *testing.T) {
	middleware, mockService, _ := setupTestMiddleware()
	user := createTestUser()
	apiKey := &types.APIKey{
		ID:     uuid.New().String(),
		UserID: user.ID,
		Name:   "Test API Key",
	}

	mockService.On("ValidateAPIKey", "test-api-key").Return(apiKey, nil)
	mockService.On("GetUserByID", user.ID).Return(user, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("X-API-Key", "test-api-key")

	handlerCalled := false
	testHandler := func(c *gin.Context) {
		handlerCalled = true
		// Verify API key context is set
		contextAPIKey, exists := c.Get("api_key")
		assert.True(t, exists)
		assert.Equal(t, apiKey.ID, contextAPIKey.(*types.APIKey).ID)
		c.Status(http.StatusOK)
	}

	middleware.RequireAPIKey()(c)
	if !c.IsAborted() {
		testHandler(c)
	}

	assert.False(t, c.IsAborted())
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestMiddleware_OptionalAuth_WithValidToken(t *testing.T) {
	middleware, mockService, jwtManager := setupTestMiddleware()
	user := createTestUser()

	token, err := jwtManager.GenerateAccessToken(user)
	assert.NoError(t, err)

	mockService.On("GetUserByID", user.ID).Return(user, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	handlerCalled := false
	testHandler := func(c *gin.Context) {
		handlerCalled = true
		// Verify user context is set
		contextUser, exists := c.Get("user")
		assert.True(t, exists)
		assert.Equal(t, user.ID, contextUser.(*types.User).ID)
		c.Status(http.StatusOK)
	}

	middleware.OptionalAuth()(c)
	if !c.IsAborted() {
		testHandler(c)
	}

	assert.False(t, c.IsAborted())
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestMiddleware_OptionalAuth_WithoutToken(t *testing.T) {
	middleware, _, _ := setupTestMiddleware()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	handlerCalled := false
	testHandler := func(c *gin.Context) {
		handlerCalled = true
		// Verify no user context is set
		_, exists := c.Get("user")
		assert.False(t, exists)
		c.Status(http.StatusOK)
	}

	middleware.OptionalAuth()(c)
	if !c.IsAborted() {
		testHandler(c)
	}

	assert.False(t, c.IsAborted())
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_ExtractToken_ValidBearer(t *testing.T) {
	middleware, _, _ := setupTestMiddleware()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer test.jwt.token")

	token := middleware.extractToken(c)
	assert.Equal(t, "test.jwt.token", token)
}

func TestMiddleware_ExtractToken_InvalidFormat(t *testing.T) {
	middleware, _, _ := setupTestMiddleware()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Basic dGVzdDp0ZXN0")

	token := middleware.extractToken(c)
	assert.Empty(t, token)
}

func TestMiddleware_HasRequiredRole_Hierarchy(t *testing.T) {
	middleware, _, _ := setupTestMiddleware()

	// Test role hierarchy
	assert.True(t, middleware.hasRequiredRole(types.RoleSystemAdmin, types.RoleAdmin))
	assert.True(t, middleware.hasRequiredRole(types.RoleAdmin, types.RoleUser))
	assert.True(t, middleware.hasRequiredRole(types.RoleUser, types.RoleViewer))
	assert.True(t, middleware.hasRequiredRole(types.RoleAdmin, types.RoleViewer))

	// Test same level
	assert.True(t, middleware.hasRequiredRole(types.RoleUser, types.RoleUser))
	assert.True(t, middleware.hasRequiredRole(types.RoleViewer, types.RoleAPIUser))

	// Test insufficient permissions
	assert.False(t, middleware.hasRequiredRole(types.RoleViewer, types.RoleUser))
	assert.False(t, middleware.hasRequiredRole(types.RoleUser, types.RoleAdmin))
	assert.False(t, middleware.hasRequiredRole(types.RoleAdmin, types.RoleSystemAdmin))
}

func TestMiddleware_RequirePermission_Success(t *testing.T) {
	middleware, _, _ := setupTestMiddleware()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// Set user context with admin role
	c.Set("role", types.RoleAdmin)

	handlerCalled := false
	testHandler := func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusOK)
	}

	// Test permission that admin should have
	middleware.RequirePermission(types.PermissionUserManage)(c)
	if !c.IsAborted() {
		testHandler(c)
	}

	assert.False(t, c.IsAborted())
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_RequirePermission_Forbidden(t *testing.T) {
	middleware, _, _ := setupTestMiddleware()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// Set user context with viewer role
	c.Set("role", types.RoleViewer)

	// Test permission that viewer should NOT have
	middleware.RequirePermission(types.PermissionUserManage)(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestMiddleware_RequireResourceAccess_Success(t *testing.T) {
	middleware, _, _ := setupTestMiddleware()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// Set user context with user role
	c.Set("role", types.RoleUser)

	handlerCalled := false
	testHandler := func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusOK)
	}

	// Test resource access that user should have (server read)
	middleware.RequireResourceAccess("server", "read")(c)
	if !c.IsAborted() {
		testHandler(c)
	}

	assert.False(t, c.IsAborted())
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_RequireResourceAccess_Forbidden(t *testing.T) {
	middleware, _, _ := setupTestMiddleware()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", "/test", nil)

	// Set user context with user role
	c.Set("role", types.RoleUser)

	// Test resource access that user should NOT have (server delete)
	middleware.RequireResourceAccess("server", "delete")(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestMiddleware_RequireSystemAdmin_Success(t *testing.T) {
	middleware, _, _ := setupTestMiddleware()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// Set user context with system admin role
	c.Set("role", types.RoleSystemAdmin)

	handlerCalled := false
	testHandler := func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusOK)
	}

	middleware.RequireSystemAdmin()(c)
	if !c.IsAborted() {
		testHandler(c)
	}

	assert.False(t, c.IsAborted())
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_RequireSystemAdmin_Forbidden(t *testing.T) {
	middleware, _, _ := setupTestMiddleware()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// Set user context with regular admin role (not system admin)
	c.Set("role", types.RoleAdmin)

	middleware.RequireSystemAdmin()(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestMiddleware_RequireOrganizationAccess_Success(t *testing.T) {
	middleware, _, _ := setupTestMiddleware()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test?organization_id=org123", nil)

	// Set user context with matching organization
	c.Set("organization_id", "org123")
	c.Set("role", types.RoleUser)

	handlerCalled := false
	testHandler := func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusOK)
	}

	middleware.RequireOrganizationAccess()(c)
	if !c.IsAborted() {
		testHandler(c)
	}

	assert.False(t, c.IsAborted())
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_RequireOrganizationAccess_SystemAdminBypass(t *testing.T) {
	middleware, _, _ := setupTestMiddleware()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test?organization_id=org456", nil)

	// Set user context with different organization but system admin role
	c.Set("organization_id", "org123")
	c.Set("role", types.RoleSystemAdmin)

	handlerCalled := false
	testHandler := func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusOK)
	}

	middleware.RequireOrganizationAccess()(c)
	if !c.IsAborted() {
		testHandler(c)
	}

	// System admin should bypass organization check
	assert.False(t, c.IsAborted())
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_RequireOrganizationAccess_Forbidden(t *testing.T) {
	middleware, _, _ := setupTestMiddleware()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test?organization_id=org456", nil)

	// Set user context with different organization
	c.Set("organization_id", "org123")
	c.Set("role", types.RoleUser)

	middleware.RequireOrganizationAccess()(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestMiddleware_RequireAnyPermission_Success(t *testing.T) {
	middleware, _, _ := setupTestMiddleware()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// Set user context with user role
	c.Set("role", types.RoleUser)

	handlerCalled := false
	testHandler := func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusOK)
	}

	// Test permissions where user should have at least one (read)
	permissions := []string{types.PermissionRead, types.PermissionUserManage}
	middleware.RequireAnyPermission(permissions)(c)
	if !c.IsAborted() {
		testHandler(c)
	}

	assert.False(t, c.IsAborted())
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_RequireAnyPermission_Forbidden(t *testing.T) {
	middleware, _, _ := setupTestMiddleware()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// Set user context with viewer role
	c.Set("role", types.RoleViewer)

	// Test permissions that viewer should NOT have
	permissions := []string{types.PermissionUserManage, types.PermissionSystemManage}
	middleware.RequireAnyPermission(permissions)(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusForbidden, w.Code)
}