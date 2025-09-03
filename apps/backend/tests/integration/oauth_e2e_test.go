package integration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
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

	"mcp-gateway/apps/backend/internal/auth"
	"mcp-gateway/apps/backend/internal/server/handlers"
	"mcp-gateway/apps/backend/internal/types"
	"mcp-gateway/apps/backend/tests/helpers"
)

// OAuthE2ETestSuite provides end-to-end testing for OAuth flows
type OAuthE2ETestSuite struct {
	db           *sql.DB
	router       *gin.Engine
	oauthService *auth.OAuthService
	teardown     func()
	testOrgID    string
	testUserID   string
}

// NewOAuthE2ETestSuite creates a new OAuth E2E test suite
func NewOAuthE2ETestSuite(t *testing.T) *OAuthE2ETestSuite {
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
	userID, err := helpers.CreateTestUserWithCredentials(testDB, orgID, "user@oauth.com", "userpassword123")
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

	return &OAuthE2ETestSuite{
		db:           testDB,
		router:       router,
		oauthService: oauthService,
		teardown:     teardown,
		testOrgID:    orgID,
		testUserID:   userID,
	}
}

// Cleanup tears down the test suite
func (suite *OAuthE2ETestSuite) Cleanup() {
	suite.teardown()
}

// TestCompleteAuthorizationCodeFlow tests the complete authorization code flow
func TestCompleteAuthorizationCodeFlow(t *testing.T) {
	suite := NewOAuthE2ETestSuite(t)
	defer suite.Cleanup()

	// Step 1: Register client
	clientReq := &types.ClientRegistrationRequest{
		ClientName:              "E2E Test Client",
		RedirectURIs:            []string{"http://localhost:3000/callback"},
		GrantTypes:              []string{types.GrantTypeAuthorizationCode, types.GrantTypeRefreshToken},
		ResponseTypes:           []string{types.ResponseTypeCode},
		TokenEndpointAuthMethod: types.TokenEndpointAuthClientSecretBasic,
		Scope:                   "read write offline_access",
	}

	client := suite.registerClient(t, clientReq)

	// Step 2: Authorization request (simulate user authorization)
	authorizationCode := suite.performAuthorization(t, client, "read write offline_access", "test-state-123")

	// Step 3: Token exchange
	tokenReq := &types.TokenRequest{
		GrantType:   types.GrantTypeAuthorizationCode,
		Code:        authorizationCode,
		RedirectURI: client.RedirectURIs[0],
		ClientID:    client.ClientID,
		ClientSecret: client.ClientSecret,
	}

	tokens := suite.exchangeCodeForTokens(t, tokenReq)

	// Verify access token works
	suite.verifyAccessToken(t, client, tokens.AccessToken, "read write offline_access")

	// Step 4: Use refresh token to get new access token
	refreshReq := &types.TokenRequest{
		GrantType:    types.GrantTypeRefreshToken,
		RefreshToken: tokens.RefreshToken,
		ClientID:     client.ClientID,
		ClientSecret: client.ClientSecret,
		Scope:        "read", // Request narrower scope
	}

	newTokens := suite.exchangeRefreshToken(t, refreshReq)

	// Verify new access token works with narrower scope
	suite.verifyAccessToken(t, client, newTokens.AccessToken, "read")

	// Step 5: Revoke token
	suite.revokeToken(t, client, newTokens.AccessToken)

	// Verify token no longer works
	suite.verifyTokenRevoked(t, newTokens.AccessToken)
}

// TestAuthorizationCodeFlowWithPKCE tests authorization code flow with PKCE
func TestAuthorizationCodeFlowWithPKCE(t *testing.T) {
	suite := NewOAuthE2ETestSuite(t)
	defer suite.Cleanup()

	// Register public client
	clientReq := &types.ClientRegistrationRequest{
		ClientName:              "PKCE Test Client",
		RedirectURIs:            []string{"http://localhost:3000/callback"},
		GrantTypes:              []string{types.GrantTypeAuthorizationCode, types.GrantTypeRefreshToken},
		ResponseTypes:           []string{types.ResponseTypeCode},
		TokenEndpointAuthMethod: types.TokenEndpointAuthNone, // Public client
		Scope:                   "read write offline_access",
	}

	client := suite.registerClient(t, clientReq)

	testCases := []struct {
		name            string
		codeVerifier    string
		challengeMethod string
		expectSuccess   bool
	}{
		{
			name:            "Valid S256 PKCE",
			codeVerifier:    "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			challengeMethod: types.CodeChallengeMethodS256,
			expectSuccess:   true,
		},
		{
			name:            "Valid Plain PKCE",
			codeVerifier:    "plain-code-verifier-test",
			challengeMethod: types.CodeChallengeMethodPlain,
			expectSuccess:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate code challenge
			var codeChallenge string
			switch tc.challengeMethod {
			case types.CodeChallengeMethodS256:
				hash := sha256.Sum256([]byte(tc.codeVerifier))
				codeChallenge = base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])
			case types.CodeChallengeMethodPlain:
				codeChallenge = tc.codeVerifier
			}

			// Authorization request with PKCE
			authCode := suite.performAuthorizationWithPKCE(t, client, "read", "test-state", codeChallenge, tc.challengeMethod)

			// Token exchange with code verifier
			tokenReq := &types.TokenRequest{
				GrantType:    types.GrantTypeAuthorizationCode,
				Code:         authCode,
				RedirectURI:  client.RedirectURIs[0],
				ClientID:     client.ClientID,
				CodeVerifier: tc.codeVerifier,
			}

			if tc.expectSuccess {
				tokens := suite.exchangeCodeForTokens(t, tokenReq)
				assert.NotEmpty(t, tokens.AccessToken)
				assert.Equal(t, "read", tokens.Scope)
			} else {
				suite.expectTokenExchangeError(t, tokenReq, types.ErrorInvalidGrant)
			}
		})
	}
}

// TestClientCredentialsFlow tests complete client credentials flow
func TestClientCredentialsFlow(t *testing.T) {
	suite := NewOAuthE2ETestSuite(t)
	defer suite.Cleanup()

	// Register client
	clientReq := &types.ClientRegistrationRequest{
		ClientName:              "Client Credentials Test",
		GrantTypes:              []string{types.GrantTypeClientCredentials},
		TokenEndpointAuthMethod: types.TokenEndpointAuthClientSecretBasic,
		Scope:                   "read write delete",
	}

	client := suite.registerClient(t, clientReq)

	testCases := []struct {
		name          string
		requestScope  string
		expectedScope string
		expectSuccess bool
	}{
		{
			name:          "Request full scope",
			requestScope:  "read write delete",
			expectedScope: "read write delete",
			expectSuccess: true,
		},
		{
			name:          "Request partial scope",
			requestScope:  "read",
			expectedScope: "read",
			expectSuccess: true,
		},
		{
			name:          "Request default scope",
			requestScope:  "",
			expectedScope: "read write delete",
			expectSuccess: true,
		},
		{
			name:          "Request invalid scope",
			requestScope:  "admin",
			expectedScope: "",
			expectSuccess: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenReq := &types.TokenRequest{
				GrantType:    types.GrantTypeClientCredentials,
				ClientID:     client.ClientID,
				ClientSecret: client.ClientSecret,
				Scope:        tc.requestScope,
			}

			if tc.expectSuccess {
				tokens := suite.exchangeCodeForTokens(t, tokenReq)
				assert.NotEmpty(t, tokens.AccessToken)
				assert.Equal(t, tc.expectedScope, tokens.Scope)
				assert.Empty(t, tokens.RefreshToken) // No refresh token for client credentials

				// Verify token works
				suite.verifyAccessToken(t, client, tokens.AccessToken, tc.expectedScope)
			} else {
				suite.expectTokenExchangeError(t, tokenReq, types.ErrorInvalidScope)
			}
		})
	}
}

// TestErrorHandling tests various error scenarios
func TestErrorHandling(t *testing.T) {
	suite := NewOAuthE2ETestSuite(t)
	defer suite.Cleanup()

	// Register client for testing
	clientReq := &types.ClientRegistrationRequest{
		ClientName:              "Error Test Client",
		RedirectURIs:            []string{"http://localhost:3000/callback"},
		GrantTypes:              []string{types.GrantTypeAuthorizationCode, types.GrantTypeClientCredentials},
		ResponseTypes:           []string{types.ResponseTypeCode},
		TokenEndpointAuthMethod: types.TokenEndpointAuthClientSecretBasic,
		Scope:                   "read write",
	}

	client := suite.registerClient(t, clientReq)

	testCases := []struct {
		name           string
		tokenRequest   *types.TokenRequest
		expectedError  string
		expectedStatus int
	}{
		{
			name: "Invalid grant type",
			tokenRequest: &types.TokenRequest{
				GrantType:    "invalid_grant",
				ClientID:     client.ClientID,
				ClientSecret: client.ClientSecret,
			},
			expectedError:  types.ErrorUnsupportedGrantType,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid client credentials",
			tokenRequest: &types.TokenRequest{
				GrantType:    types.GrantTypeClientCredentials,
				ClientID:     client.ClientID,
				ClientSecret: "wrong-secret",
			},
			expectedError:  types.ErrorInvalidClient,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing client ID",
			tokenRequest: &types.TokenRequest{
				GrantType:    types.GrantTypeClientCredentials,
				ClientSecret: client.ClientSecret,
			},
			expectedError:  types.ErrorInvalidRequest,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Client not authorized for grant type",
			tokenRequest: &types.TokenRequest{
				GrantType:    types.GrantTypeRefreshToken,
				ClientID:     client.ClientID,
				ClientSecret: client.ClientSecret,
				RefreshToken: "dummy-refresh-token",
			},
			expectedError:  types.ErrorUnauthorizedClient,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid authorization code",
			tokenRequest: &types.TokenRequest{
				GrantType:   types.GrantTypeAuthorizationCode,
				Code:        "invalid-code",
				RedirectURI: client.RedirectURIs[0],
				ClientID:    client.ClientID,
				ClientSecret: client.ClientSecret,
			},
			expectedError:  types.ErrorInvalidGrant,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tc.tokenRequest)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/oauth/token", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			var errorResp types.OAuthError
			err = json.Unmarshal(w.Body.Bytes(), &errorResp)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedError, errorResp.Error)
		})
	}
}

// TestTokenLifecycle tests the complete token lifecycle
func TestTokenLifecycle(t *testing.T) {
	suite := NewOAuthE2ETestSuite(t)
	defer suite.Cleanup()

	// Register client
	client := suite.registerClient(t, &types.ClientRegistrationRequest{
		ClientName:              "Lifecycle Test Client",
		GrantTypes:              []string{types.GrantTypeClientCredentials},
		TokenEndpointAuthMethod: types.TokenEndpointAuthClientSecretBasic,
		Scope:                   "read write",
	})

	// Get token
	tokenReq := &types.TokenRequest{
		GrantType:    types.GrantTypeClientCredentials,
		ClientID:     client.ClientID,
		ClientSecret: client.ClientSecret,
		Scope:        "read",
	}

	tokens := suite.exchangeCodeForTokens(t, tokenReq)

	// 1. Verify token is active via introspection
	introspectionResp := suite.introspectToken(t, client, tokens.AccessToken)
	assert.True(t, introspectionResp.Active)
	assert.Equal(t, "read", introspectionResp.Scope)
	assert.Equal(t, client.ClientID, introspectionResp.ClientID)

	// 2. Verify token can be used for validation
	suite.verifyAccessToken(t, client, tokens.AccessToken, "read")

	// 3. Revoke token
	suite.revokeToken(t, client, tokens.AccessToken)

	// 4. Verify token is no longer active
	introspectionResp = suite.introspectToken(t, client, tokens.AccessToken)
	assert.False(t, introspectionResp.Active)

	// 5. Verify token validation fails
	suite.verifyTokenRevoked(t, tokens.AccessToken)
}

// Helper methods

func (suite *OAuthE2ETestSuite) registerClient(t *testing.T, req *types.ClientRegistrationRequest) *types.ClientRegistrationResponse {
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

func (suite *OAuthE2ETestSuite) performAuthorization(t *testing.T, client *types.ClientRegistrationResponse, scope, state string) string {
	params := url.Values{}
	params.Set("response_type", types.ResponseTypeCode)
	params.Set("client_id", client.ClientID)
	params.Set("redirect_uri", client.RedirectURIs[0])
	params.Set("scope", scope)
	params.Set("state", state)
	params.Set("test_user_id", suite.testUserID) // Bypass authentication for testing

	req := httptest.NewRequest(http.MethodGet, "/oauth/authorize?"+params.Encode(), nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	require.Equal(t, http.StatusFound, w.Code)

	location := w.Header().Get("Location")
	require.NotEmpty(t, location)

	// Parse authorization code from redirect
	redirectURL, err := url.Parse(location)
	require.NoError(t, err)

	code := redirectURL.Query().Get("code")
	require.NotEmpty(t, code)

	// Verify state parameter
	assert.Equal(t, state, redirectURL.Query().Get("state"))

	return code
}

func (suite *OAuthE2ETestSuite) performAuthorizationWithPKCE(t *testing.T, client *types.ClientRegistrationResponse, scope, state, codeChallenge, challengeMethod string) string {
	// Create authorization code with PKCE
	ctx := context.Background()
	code, err := suite.oauthService.CreateAuthorizationCode(
		ctx, client.ClientID, suite.testUserID, client.RedirectURIs[0], scope,
		&codeChallenge, &challengeMethod,
	)
	require.NoError(t, err)

	return code
}

func (suite *OAuthE2ETestSuite) exchangeCodeForTokens(t *testing.T, req *types.TokenRequest) *types.TokenResponse {
	reqBody, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq := httptest.NewRequest(http.MethodPost, "/oauth/token", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, httpReq)
	require.Equal(t, http.StatusOK, w.Code)

	var response types.TokenResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	return &response
}

func (suite *OAuthE2ETestSuite) exchangeRefreshToken(t *testing.T, req *types.TokenRequest) *types.TokenResponse {
	return suite.exchangeCodeForTokens(t, req) // Same logic
}

func (suite *OAuthE2ETestSuite) expectTokenExchangeError(t *testing.T, req *types.TokenRequest, expectedError string) {
	reqBody, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq := httptest.NewRequest(http.MethodPost, "/oauth/token", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResp types.OAuthError
	err = json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, expectedError, errorResp.Error)
}

func (suite *OAuthE2ETestSuite) verifyAccessToken(t *testing.T, client *types.ClientRegistrationResponse, token, expectedScope string) {
	// Use token introspection to verify token
	introspectionResp := suite.introspectToken(t, client, token)
	assert.True(t, introspectionResp.Active)
	assert.Equal(t, expectedScope, introspectionResp.Scope)
	assert.Equal(t, client.ClientID, introspectionResp.ClientID)
}

func (suite *OAuthE2ETestSuite) verifyTokenRevoked(t *testing.T, token string) {
	ctx := context.Background()
	_, err := suite.oauthService.ValidateToken(ctx, token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid or expired token")
}

func (suite *OAuthE2ETestSuite) introspectToken(t *testing.T, client *types.ClientRegistrationResponse, token string) *types.IntrospectionResponse {
	// OAuth 2.0 introspection uses form data, not JSON
	data := url.Values{}
	data.Set("token", token)

	req := httptest.NewRequest(http.MethodPost, "/oauth/introspect", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(client.ClientID, client.ClientSecret)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var response types.IntrospectionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	return &response
}

func (suite *OAuthE2ETestSuite) revokeToken(t *testing.T, client *types.ClientRegistrationResponse, token string) {
	// OAuth 2.0 revocation uses form data, not JSON
	data := url.Values{}
	data.Set("token", token)

	req := httptest.NewRequest(http.MethodPost, "/oauth/revoke", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(client.ClientID, client.ClientSecret)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}
