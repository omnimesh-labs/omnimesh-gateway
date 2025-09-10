package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/auth"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/server/handlers"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/tests/helpers"
)

// OAuthIntegrationTestSuite provides test setup for OAuth integration tests
type OAuthIntegrationTestSuite struct {
	db           *sql.DB
	router       *gin.Engine
	oauthService *auth.OAuthService
	teardown     func()
	testOrgID    string
	testUserID   string
}

// NewOAuthIntegrationTestSuite creates a new OAuth integration test suite
func NewOAuthIntegrationTestSuite(t *testing.T) *OAuthIntegrationTestSuite {
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
	userID, err := helpers.CreateTestUserWithCredentials(testDB, orgID, "test@oauth.com", "testpassword123")
	require.NoError(t, err)

	// Setup OAuth service
	sqlxDB := sqlx.NewDb(testDB, "postgres")
	oauthConfig := auth.DefaultOAuthConfig()
	oauthConfig.Issuer = "http://localhost:8080"
	oauthService := auth.NewOAuthService(sqlxDB, "test-jwt-secret", "http://localhost:8080", oauthConfig)

	// Setup router with OAuth handlers
	gin.SetMode(gin.TestMode)
	router := gin.New()
	oauthHandler := handlers.NewOAuthHandler(oauthService)

	// Setup OAuth discovery routes
	router.GET("/.well-known/oauth-authorization-server", oauthHandler.DiscoverAuthorizationServer)
	router.GET("/.well-known/oauth-protected-resource", oauthHandler.DiscoverProtectedResource)

	// Setup OAuth routes
	oauth := router.Group("/oauth")
	{
		oauth.POST("/register", oauthHandler.RegisterClient)
		oauth.POST("/token", oauthHandler.IssueToken)
		oauth.POST("/introspect", oauthHandler.IntrospectToken)
		oauth.POST("/revoke", oauthHandler.RevokeToken)
		oauth.GET("/jwks", oauthHandler.GetJWKS)
		oauth.GET("/authorize", oauthHandler.AuthorizeEndpoint)
		oauth.POST("/authorize", oauthHandler.AuthorizeEndpoint)
	}

	return &OAuthIntegrationTestSuite{
		db:           testDB,
		router:       router,
		oauthService: oauthService,
		teardown:     teardown,
		testOrgID:    orgID,
		testUserID:   userID,
	}
}

// Cleanup tears down the test suite
func (suite *OAuthIntegrationTestSuite) Cleanup() {
	suite.teardown()
}

// TestOAuthDiscoveryEndpoints tests OAuth discovery endpoints
func TestOAuthDiscoveryEndpoints(t *testing.T) {
	suite := NewOAuthIntegrationTestSuite(t)
	defer suite.Cleanup()

	t.Run("Authorization Server Discovery", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-authorization-server", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var metadata types.AuthorizationServerMetadata
		err := json.Unmarshal(w.Body.Bytes(), &metadata)
		require.NoError(t, err)

		assert.Equal(t, "http://localhost:8080", metadata.Issuer)
		assert.Equal(t, "http://localhost:8080/oauth/token", metadata.TokenEndpoint)
		assert.NotNil(t, metadata.AuthorizationEndpoint)
		assert.Equal(t, "http://localhost:8080/oauth/authorize", *metadata.AuthorizationEndpoint)
		assert.Contains(t, metadata.GrantTypesSupported, types.GrantTypeClientCredentials)
		assert.Contains(t, metadata.GrantTypesSupported, types.GrantTypeAuthorizationCode)
	})

	t.Run("Protected Resource Discovery", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-protected-resource", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var metadata types.ProtectedResourceMetadata
		err := json.Unmarshal(w.Body.Bytes(), &metadata)
		require.NoError(t, err)

		assert.Equal(t, "http://localhost:8080", metadata.Resource)
		assert.Contains(t, metadata.AuthorizationServers, "http://localhost:8080")
	})
}

// TestClientRegistrationEndpoint tests client registration
func TestClientRegistrationEndpoint(t *testing.T) {
	suite := NewOAuthIntegrationTestSuite(t)
	defer suite.Cleanup()

	testCases := []struct {
		name           string
		request        types.ClientRegistrationRequest
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name: "Valid Client Registration",
			request: types.ClientRegistrationRequest{
				ClientName:              "Test Client",
				RedirectURIs:            []string{"http://localhost:3000/callback"},
				GrantTypes:              []string{types.GrantTypeAuthorizationCode, types.GrantTypeRefreshToken},
				ResponseTypes:           []string{types.ResponseTypeCode},
				TokenEndpointAuthMethod: types.TokenEndpointAuthClientSecretBasic,
				Scope:                   "read write",
			},
			expectedStatus: http.StatusCreated,
			expectSuccess:  true,
		},
		{
			name: "Public Client Registration",
			request: types.ClientRegistrationRequest{
				ClientName:              "Public Client",
				RedirectURIs:            []string{"http://localhost:3000/callback"},
				GrantTypes:              []string{types.GrantTypeAuthorizationCode},
				ResponseTypes:           []string{types.ResponseTypeCode},
				TokenEndpointAuthMethod: types.TokenEndpointAuthNone,
				Scope:                   "read",
			},
			expectedStatus: http.StatusCreated,
			expectSuccess:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tc.request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/oauth/register", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectSuccess {
				var response types.ClientRegistrationResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.NotEmpty(t, response.ClientID)
				assert.True(t, strings.HasPrefix(response.ClientID, "client_"))
				assert.Equal(t, tc.request.ClientName, response.ClientName)
				assert.Equal(t, tc.request.RedirectURIs, response.RedirectURIs)

				if tc.request.TokenEndpointAuthMethod != types.TokenEndpointAuthNone {
					assert.NotEmpty(t, response.ClientSecret)
				}
			}
		})
	}
}

// TestTokenEndpoint tests token endpoint with different grant types
func TestTokenEndpoint(t *testing.T) {
	suite := NewOAuthIntegrationTestSuite(t)
	defer suite.Cleanup()

	// Register a client first
	client := suite.registerTestClient(t, &types.ClientRegistrationRequest{
		ClientName:              "Token Test Client",
		GrantTypes:              []string{types.GrantTypeClientCredentials, types.GrantTypeAuthorizationCode, types.GrantTypeRefreshToken},
		TokenEndpointAuthMethod: types.TokenEndpointAuthClientSecretBasic,
		RedirectURIs:            []string{"http://localhost:3000/callback"},
		Scope:                   "read write offline_access",
	})

	t.Run("Client Credentials Grant", func(t *testing.T) {
		tokenReq := types.TokenRequest{
			GrantType:    types.GrantTypeClientCredentials,
			ClientID:     client.ClientID,
			ClientSecret: client.ClientSecret,
			Scope:        "read",
		}

		reqBody, err := json.Marshal(tokenReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/oauth/token", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response types.TokenResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.AccessToken)
		assert.Equal(t, "Bearer", response.TokenType)
		assert.Greater(t, response.ExpiresIn, int64(0))
		assert.Equal(t, "read", response.Scope)
		assert.Empty(t, response.RefreshToken) // No refresh token for client credentials

		// Verify cache headers
		assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))
		assert.Equal(t, "no-cache", w.Header().Get("Pragma"))
	})

	t.Run("Client Credentials with Basic Auth", func(t *testing.T) {
		tokenReq := types.TokenRequest{
			GrantType: types.GrantTypeClientCredentials,
			Scope:     "read",
		}

		reqBody, err := json.Marshal(tokenReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/oauth/token", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth(client.ClientID, client.ClientSecret)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Authorization Code Grant", func(t *testing.T) {
		// Create authorization code first
		ctx := context.Background()
		code, err := suite.oauthService.CreateAuthorizationCode(
			ctx, client.ClientID, suite.testUserID, client.RedirectURIs[0],
			"read write offline_access", nil, nil,
		)
		require.NoError(t, err)

		tokenReq := types.TokenRequest{
			GrantType:    types.GrantTypeAuthorizationCode,
			Code:         code,
			RedirectURI:  client.RedirectURIs[0],
			ClientID:     client.ClientID,
			ClientSecret: client.ClientSecret,
		}

		reqBody, err := json.Marshal(tokenReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/oauth/token", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response types.TokenResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.AccessToken)
		assert.Equal(t, "Bearer", response.TokenType)
		assert.Greater(t, response.ExpiresIn, int64(0))
		assert.Equal(t, "read write offline_access", response.Scope)
		assert.NotEmpty(t, response.RefreshToken) // Should have refresh token
	})

	t.Run("Invalid Client Credentials", func(t *testing.T) {
		tokenReq := types.TokenRequest{
			GrantType:    types.GrantTypeClientCredentials,
			ClientID:     client.ClientID,
			ClientSecret: "wrong-secret",
			Scope:        "read",
		}

		reqBody, err := json.Marshal(tokenReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/oauth/token", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResp types.OAuthError
		err = json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Equal(t, types.ErrorInvalidClient, errorResp.Error)
	})

	t.Run("Missing Grant Type", func(t *testing.T) {
		tokenReq := types.TokenRequest{
			ClientID:     client.ClientID,
			ClientSecret: client.ClientSecret,
			Scope:        "read",
		}

		reqBody, err := json.Marshal(tokenReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/oauth/token", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResp types.OAuthError
		err = json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Equal(t, types.ErrorInvalidRequest, errorResp.Error)
	})
}

// TestJWKSEndpoint tests JWKS endpoint
func TestJWKSEndpoint(t *testing.T) {
	suite := NewOAuthIntegrationTestSuite(t)
	defer suite.Cleanup()

	req := httptest.NewRequest(http.MethodGet, "/oauth/jwks", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/jwk-set+json", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Cache-Control"), "public")

	var jwks map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &jwks)
	require.NoError(t, err)

	keys, ok := jwks["keys"].([]interface{})
	require.True(t, ok)
	require.Len(t, keys, 1)

	key := keys[0].(map[string]interface{})
	assert.Equal(t, "oct", key["kty"])
	assert.Equal(t, "sig", key["use"])
	assert.Equal(t, "HS256", key["alg"])
}

// TestTokenIntrospectionEndpoint tests token introspection
func TestTokenIntrospectionEndpoint(t *testing.T) {
	suite := NewOAuthIntegrationTestSuite(t)
	defer suite.Cleanup()

	// Register client and get token
	client := suite.registerTestClient(t, &types.ClientRegistrationRequest{
		ClientName:              "Introspection Test Client",
		GrantTypes:              []string{types.GrantTypeClientCredentials},
		TokenEndpointAuthMethod: types.TokenEndpointAuthClientSecretBasic,
		Scope:                   "read write",
	})

	token := suite.getClientCredentialsToken(t, client, "read")

	t.Run("Valid Token Introspection", func(t *testing.T) {
		// OAuth 2.0 introspection uses form data, not JSON
		data := url.Values{}
		data.Set("token", token)

		req := httptest.NewRequest(http.MethodPost, "/oauth/introspect", strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth(client.ClientID, client.ClientSecret)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response types.IntrospectionResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Active)
		assert.Equal(t, client.ClientID, response.ClientID)
		assert.Equal(t, "Bearer", response.TokenType)
		assert.Equal(t, "read", response.Scope)
	})

	t.Run("Invalid Token Introspection", func(t *testing.T) {
		// OAuth 2.0 introspection uses form data, not JSON
		data := url.Values{}
		data.Set("token", "invalid-token")

		req := httptest.NewRequest(http.MethodPost, "/oauth/introspect", strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth(client.ClientID, client.ClientSecret)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response types.IntrospectionResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.False(t, response.Active)
	})

	t.Run("No Client Authentication", func(t *testing.T) {
		// OAuth 2.0 introspection uses form data, not JSON
		data := url.Values{}
		data.Set("token", token)

		req := httptest.NewRequest(http.MethodPost, "/oauth/introspect", strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Equal(t, "Basic", w.Header().Get("WWW-Authenticate"))
	})
}

// TestTokenRevocationEndpoint tests token revocation
func TestTokenRevocationEndpoint(t *testing.T) {
	suite := NewOAuthIntegrationTestSuite(t)
	defer suite.Cleanup()

	// Register client and get token
	client := suite.registerTestClient(t, &types.ClientRegistrationRequest{
		ClientName:              "Revocation Test Client",
		GrantTypes:              []string{types.GrantTypeClientCredentials},
		TokenEndpointAuthMethod: types.TokenEndpointAuthClientSecretBasic,
		Scope:                   "read write",
	})

	token := suite.getClientCredentialsToken(t, client, "read")

	t.Run("Valid Token Revocation", func(t *testing.T) {
		// OAuth 2.0 revocation uses form data, not JSON
		data := url.Values{}
		data.Set("token", token)

		req := httptest.NewRequest(http.MethodPost, "/oauth/revoke", strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth(client.ClientID, client.ClientSecret)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Body.String()) // Should be empty response per RFC 7009
	})

	t.Run("Already Revoked Token", func(t *testing.T) {
		// Should still return 200 OK per OAuth spec
		// OAuth 2.0 revocation uses form data, not JSON
		data := url.Values{}
		data.Set("token", token)

		req := httptest.NewRequest(http.MethodPost, "/oauth/revoke", strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth(client.ClientID, client.ClientSecret)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestAuthorizeEndpoint tests authorization endpoint
func TestAuthorizeEndpoint(t *testing.T) {
	suite := NewOAuthIntegrationTestSuite(t)
	defer suite.Cleanup()

	// Register client
	client := suite.registerTestClient(t, &types.ClientRegistrationRequest{
		ClientName:              "Authorization Test Client",
		RedirectURIs:            []string{"http://localhost:3000/callback"},
		GrantTypes:              []string{types.GrantTypeAuthorizationCode},
		ResponseTypes:           []string{types.ResponseTypeCode},
		TokenEndpointAuthMethod: types.TokenEndpointAuthClientSecretBasic,
		Scope:                   "read write",
	})

	t.Run("Valid Authorization Request", func(t *testing.T) {
		params := url.Values{}
		params.Set("response_type", types.ResponseTypeCode)
		params.Set("client_id", client.ClientID)
		params.Set("redirect_uri", client.RedirectURIs[0])
		params.Set("scope", "read")
		params.Set("state", "test-state")
		params.Set("test_user_id", suite.testUserID) // Bypass authentication for testing

		req := httptest.NewRequest(http.MethodGet, "/oauth/authorize?"+params.Encode(), nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		// Should redirect with authorization code
		assert.Equal(t, http.StatusFound, w.Code)

		location := w.Header().Get("Location")
		assert.NotEmpty(t, location)
		assert.Contains(t, location, "code=")
		assert.Contains(t, location, "state=test-state")
	})

	t.Run("Missing Response Type", func(t *testing.T) {
		params := url.Values{}
		params.Set("client_id", client.ClientID)
		params.Set("redirect_uri", client.RedirectURIs[0])

		req := httptest.NewRequest(http.MethodGet, "/oauth/authorize?"+params.Encode(), nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		// Should redirect with error
		assert.Equal(t, http.StatusFound, w.Code)
		location := w.Header().Get("Location")
		assert.Contains(t, location, "error=invalid_request")
	})

	t.Run("Invalid Client ID", func(t *testing.T) {
		params := url.Values{}
		params.Set("response_type", types.ResponseTypeCode)
		params.Set("client_id", "invalid-client")
		params.Set("redirect_uri", "http://example.com/callback")

		req := httptest.NewRequest(http.MethodGet, "/oauth/authorize?"+params.Encode(), nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		// Should redirect with error
		assert.Equal(t, http.StatusFound, w.Code)
		location := w.Header().Get("Location")
		assert.Contains(t, location, "error=invalid_client")
	})

	t.Run("Missing Client ID", func(t *testing.T) {
		params := url.Values{}
		params.Set("response_type", types.ResponseTypeCode)

		req := httptest.NewRequest(http.MethodGet, "/oauth/authorize?"+params.Encode(), nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		// Should return JSON error (can't redirect without client_id)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResp types.OAuthError
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Equal(t, types.ErrorInvalidRequest, errorResp.Error)
	})
}

// Helper methods

func (suite *OAuthIntegrationTestSuite) registerTestClient(t *testing.T, req *types.ClientRegistrationRequest) *types.ClientRegistrationResponse {
	reqBody, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq := httptest.NewRequest(http.MethodPost, "/oauth/register", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, httpReq)

	require.Equal(t, http.StatusCreated, w.Code)

	var response types.ClientRegistrationResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	return &response
}

func (suite *OAuthIntegrationTestSuite) getClientCredentialsToken(t *testing.T, client *types.ClientRegistrationResponse, scope string) string {
	tokenReq := types.TokenRequest{
		GrantType:    types.GrantTypeClientCredentials,
		ClientID:     client.ClientID,
		ClientSecret: client.ClientSecret,
		Scope:        scope,
	}

	reqBody, err := json.Marshal(tokenReq)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/oauth/token", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response types.TokenResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	return response.AccessToken
}
