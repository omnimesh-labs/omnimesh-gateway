package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"mcp-gateway/apps/backend/internal/types"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// OAuthService handles OAuth 2.0 operations
type OAuthService struct {
	db        *sqlx.DB
	jwtSecret string
	issuer    string
	config    *OAuthConfig
}

// OAuthConfig holds OAuth 2.0 configuration
type OAuthConfig struct {
	Issuer                    string        `yaml:"issuer"`
	AuthorizationEndpoint     string        `yaml:"authorization_endpoint"`
	TokenEndpoint             string        `yaml:"token_endpoint"`
	RegistrationEndpoint      string        `yaml:"registration_endpoint"`
	IntrospectionEndpoint     string        `yaml:"introspection_endpoint"`
	RevocationEndpoint        string        `yaml:"revocation_endpoint"`
	JWKSUri                   string        `yaml:"jwks_uri"`
	SupportedGrantTypes       []string      `yaml:"supported_grant_types"`
	SupportedResponseTypes    []string      `yaml:"supported_response_types"`
	SupportedScopes           []string      `yaml:"supported_scopes"`
	TokenExpiry               time.Duration `yaml:"token_expiry"`
	RefreshTokenExpiry        time.Duration `yaml:"refresh_token_expiry"`
	AuthCodeExpiry            time.Duration `yaml:"auth_code_expiry"`
	EnableDynamicRegistration bool          `yaml:"enable_dynamic_registration"`
	RequireClientAuth         bool          `yaml:"require_client_authentication"`
	AllowPublicClients        bool          `yaml:"allow_public_clients"`
}

// DefaultOAuthConfig returns default OAuth configuration
func DefaultOAuthConfig() *OAuthConfig {
	return &OAuthConfig{
		Issuer:                    "http://localhost:8080",
		AuthorizationEndpoint:     "/oauth/authorize",
		TokenEndpoint:             "/oauth/token",
		RegistrationEndpoint:      "/oauth/register",
		IntrospectionEndpoint:     "/oauth/introspect",
		RevocationEndpoint:        "/oauth/revoke",
		JWKSUri:                   "/oauth/jwks",
		SupportedGrantTypes:       []string{types.GrantTypeClientCredentials, types.GrantTypeAuthorizationCode, types.GrantTypeRefreshToken},
		SupportedResponseTypes:    []string{types.ResponseTypeCode, types.ResponseTypeToken},
		SupportedScopes:           []string{types.ScopeRead, types.ScopeWrite, types.ScopeDelete, types.ScopeExecute, types.ScopeAdmin},
		TokenExpiry:               time.Hour,
		RefreshTokenExpiry:        24 * time.Hour * 30, // 30 days
		AuthCodeExpiry:            10 * time.Minute,
		EnableDynamicRegistration: true,
		RequireClientAuth:         true,
		AllowPublicClients:        true,
	}
}

// NewOAuthService creates a new OAuth service
func NewOAuthService(db *sqlx.DB, jwtSecret string, issuer string, config *OAuthConfig) *OAuthService {
	if config == nil {
		config = DefaultOAuthConfig()
	}
	if issuer != "" {
		config.Issuer = issuer
	}

	return &OAuthService{
		db:        db,
		jwtSecret: jwtSecret,
		issuer:    config.Issuer,
		config:    config,
	}
}

// GetServerMetadata returns OAuth 2.0 Authorization Server Metadata
func (s *OAuthService) GetServerMetadata() *types.AuthorizationServerMetadata {
	baseURL := strings.TrimSuffix(s.issuer, "/")

	return &types.AuthorizationServerMetadata{
		Issuer:                                    s.issuer,
		AuthorizationEndpoint:                     stringPtr(baseURL + s.config.AuthorizationEndpoint),
		TokenEndpoint:                             baseURL + s.config.TokenEndpoint,
		JWKSUri:                                   stringPtr(baseURL + s.config.JWKSUri),
		RegistrationEndpoint:                      stringPtr(baseURL + s.config.RegistrationEndpoint),
		ScopesSupported:                           s.config.SupportedScopes,
		ResponseTypesSupported:                    s.config.SupportedResponseTypes,
		GrantTypesSupported:                       s.config.SupportedGrantTypes,
		TokenEndpointAuthMethodsSupported:         []string{types.TokenEndpointAuthClientSecretBasic, types.TokenEndpointAuthClientSecretPost, types.TokenEndpointAuthNone},
		RevocationEndpoint:                        stringPtr(baseURL + s.config.RevocationEndpoint),
		RevocationEndpointAuthMethodsSupported:    []string{types.TokenEndpointAuthClientSecretBasic, types.TokenEndpointAuthClientSecretPost},
		IntrospectionEndpoint:                     stringPtr(baseURL + s.config.IntrospectionEndpoint),
		IntrospectionEndpointAuthMethodsSupported: []string{types.TokenEndpointAuthClientSecretBasic, types.TokenEndpointAuthClientSecretPost},
		CodeChallengeMethodsSupported:             []string{types.CodeChallengeMethodPlain, types.CodeChallengeMethodS256},
	}
}

// GetProtectedResourceMetadata returns OAuth 2.0 Protected Resource Metadata
func (s *OAuthService) GetProtectedResourceMetadata() *types.ProtectedResourceMetadata {
	baseURL := strings.TrimSuffix(s.issuer, "/")

	return &types.ProtectedResourceMetadata{
		Resource:               s.issuer,
		AuthorizationServers:   []string{s.issuer},
		ScopesSupported:        s.config.SupportedScopes,
		BearerMethodsSupported: []string{"header", "body", "query"},
		IntrospectionEndpoint:  stringPtr(baseURL + s.config.IntrospectionEndpoint),
		IntrospectionEndpointAuthMethodsSupported: []string{types.TokenEndpointAuthClientSecretBasic, types.TokenEndpointAuthClientSecretPost},
	}
}

// RegisterClient handles dynamic client registration
func (s *OAuthService) RegisterClient(ctx context.Context, req *types.ClientRegistrationRequest, orgID string) (*types.ClientRegistrationResponse, error) {
	if !s.config.EnableDynamicRegistration {
		return nil, fmt.Errorf("dynamic client registration is disabled")
	}

	// Generate client ID
	clientID := generateClientID()

	// Generate client secret for confidential clients
	var clientSecret string
	var clientSecretHash *string

	// Default to confidential client if not specified
	tokenEndpointAuthMethod := req.TokenEndpointAuthMethod
	if tokenEndpointAuthMethod == "" {
		tokenEndpointAuthMethod = types.TokenEndpointAuthClientSecretBasic
	}

	// Generate secret for clients that need it
	if tokenEndpointAuthMethod != types.TokenEndpointAuthNone {
		clientSecret = generateClientSecret()
		hash, err := bcrypt.GenerateFromPassword([]byte(clientSecret), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash client secret: %w", err)
		}
		hashStr := string(hash)
		clientSecretHash = &hashStr
	}

	// Set defaults
	grantTypes := req.GrantTypes
	if len(grantTypes) == 0 {
		grantTypes = []string{types.GrantTypeClientCredentials}
	}

	responseTypes := req.ResponseTypes
	if len(responseTypes) == 0 {
		responseTypes = []string{types.ResponseTypeToken}
	}

	scope := req.Scope
	if scope == "" {
		scope = types.ScopeRead
	}

	// Create client record
	client := &types.OAuthClient{
		ID:                      uuid.New().String(),
		ClientID:                clientID,
		ClientSecretHash:        clientSecretHash,
		ClientName:              req.ClientName,
		ClientType:              types.ClientTypeConfidential,
		RedirectURIs:            req.RedirectURIs,
		GrantTypes:              grantTypes,
		ResponseTypes:           responseTypes,
		Scope:                   scope,
		Contacts:                req.Contacts,
		LogoURI:                 stringPtrIfNotEmpty(req.LogoURI),
		ClientURI:               stringPtrIfNotEmpty(req.ClientURI),
		PolicyURI:               stringPtrIfNotEmpty(req.PolicyURI),
		TOSURI:                  stringPtrIfNotEmpty(req.TOSURI),
		JWKSURI:                 stringPtrIfNotEmpty(req.JWKSURI),
		TokenEndpointAuthMethod: tokenEndpointAuthMethod,
		OrganizationID:          orgID,
		IsActive:                true,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}

	// Set client type based on auth method
	if tokenEndpointAuthMethod == types.TokenEndpointAuthNone {
		client.ClientType = types.ClientTypePublic
	}

	// Insert into database
	query := `
		INSERT INTO oauth_clients (
			id, client_id, client_secret_hash, client_name, client_type,
			redirect_uris, grant_types, response_types, scope, contacts,
			logo_uri, client_uri, policy_uri, tos_uri, jwks_uri,
			token_endpoint_auth_method, organization_id, is_active,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		)`

	_, err := s.db.ExecContext(ctx, query,
		client.ID, client.ClientID, client.ClientSecretHash, client.ClientName, client.ClientType,
		pq.Array(client.RedirectURIs), pq.Array(client.GrantTypes), pq.Array(client.ResponseTypes), client.Scope, pq.Array(client.Contacts),
		client.LogoURI, client.ClientURI, client.PolicyURI, client.TOSURI, client.JWKSURI,
		client.TokenEndpointAuthMethod, client.OrganizationID, client.IsActive,
		client.CreatedAt, client.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert client: %w", err)
	}

	// Build response
	response := &types.ClientRegistrationResponse{
		ClientID:                client.ClientID,
		ClientIdIssuedAt:        client.CreatedAt.Unix(),
		ClientSecretExpiresAt:   0, // Never expires
		RedirectURIs:            client.RedirectURIs,
		TokenEndpointAuthMethod: client.TokenEndpointAuthMethod,
		GrantTypes:              client.GrantTypes,
		ResponseTypes:           client.ResponseTypes,
		ClientName:              client.ClientName,
		Scope:                   client.Scope,
		Contacts:                client.Contacts,
	}

	if client.ClientURI != nil {
		response.ClientURI = *client.ClientURI
	}
	if client.LogoURI != nil {
		response.LogoURI = *client.LogoURI
	}
	if client.PolicyURI != nil {
		response.PolicyURI = *client.PolicyURI
	}
	if client.TOSURI != nil {
		response.TOSURI = *client.TOSURI
	}
	if client.JWKSURI != nil {
		response.JWKSURI = *client.JWKSURI
	}

	// Include client secret if generated
	if clientSecret != "" {
		response.ClientSecret = clientSecret
	}

	return response, nil
}

// IssueToken issues an access token based on the grant type
func (s *OAuthService) IssueToken(ctx context.Context, req *types.TokenRequest) (*types.TokenResponse, error) {
	switch req.GrantType {
	case types.GrantTypeClientCredentials:
		return s.handleClientCredentialsGrant(ctx, req)
	case types.GrantTypeAuthorizationCode:
		return s.handleAuthorizationCodeGrant(ctx, req)
	case types.GrantTypeRefreshToken:
		return s.handleRefreshTokenGrant(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported grant type: %s", req.GrantType)
	}
}

// handleClientCredentialsGrant handles client_credentials grant
func (s *OAuthService) handleClientCredentialsGrant(ctx context.Context, req *types.TokenRequest) (*types.TokenResponse, error) {
	// Authenticate client
	client, err := s.authenticateClient(ctx, req.ClientID, req.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("client authentication failed: %w", err)
	}

	// Check if client is authorized for this grant type
	if !contains(client.GrantTypes, types.GrantTypeClientCredentials) {
		return nil, fmt.Errorf("client not authorized for client_credentials grant")
	}

	// Validate scope
	scope := req.Scope
	if scope == "" {
		scope = client.Scope
	}
	if !types.ValidateScope(scope, types.ParseScope(client.Scope)) {
		return nil, fmt.Errorf("invalid scope")
	}

	// Generate access token
	expiresAt := time.Now().Add(s.config.TokenExpiry)
	accessToken, err := s.generateAccessToken(client.ClientID, "", scope, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Store token in database
	tokenHash := hashToken(accessToken)
	tokenRecord := &types.OAuthToken{
		ID:        uuid.New().String(),
		TokenHash: tokenHash,
		TokenType: types.TokenTypeAccess,
		ClientID:  client.ClientID,
		UserID:    nil, // No user for client credentials
		Scope:     scope,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	query := `
		INSERT INTO oauth_tokens (
			id, token_hash, token_type, client_id, user_id, scope, expires_at, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)`

	_, err = s.db.ExecContext(ctx, query,
		tokenRecord.ID, tokenRecord.TokenHash, tokenRecord.TokenType, tokenRecord.ClientID,
		tokenRecord.UserID, tokenRecord.Scope, tokenRecord.ExpiresAt, tokenRecord.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to store token: %w", err)
	}

	return &types.TokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.config.TokenExpiry.Seconds()),
		Scope:       scope,
	}, nil
}

// handleAuthorizationCodeGrant handles authorization_code grant
func (s *OAuthService) handleAuthorizationCodeGrant(ctx context.Context, req *types.TokenRequest) (*types.TokenResponse, error) {
	// TODO: Implement authorization code grant
	return nil, fmt.Errorf("authorization_code grant not yet implemented")
}

// handleRefreshTokenGrant handles refresh_token grant
func (s *OAuthService) handleRefreshTokenGrant(ctx context.Context, req *types.TokenRequest) (*types.TokenResponse, error) {
	// TODO: Implement refresh token grant
	return nil, fmt.Errorf("refresh_token grant not yet implemented")
}

// IntrospectToken introspects an OAuth token
func (s *OAuthService) IntrospectToken(ctx context.Context, token string) (*types.IntrospectionResponse, error) {
	tokenHash := hashToken(token)

	var tokenRecord types.OAuthToken
	query := `
		SELECT t.id, t.token_hash, t.token_type, t.client_id, t.user_id, t.scope, 
			   t.expires_at, t.revoked_at, t.parent_token_id, t.created_at,
			   c.client_name, c.organization_id, u.email as user_email, u.role as user_role
		FROM oauth_tokens t
		JOIN oauth_clients c ON t.client_id = c.client_id
		LEFT JOIN users u ON t.user_id = u.id
		WHERE t.token_hash = $1 AND t.revoked_at IS NULL`

	err := s.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&tokenRecord.ID, &tokenRecord.TokenHash, &tokenRecord.TokenType, &tokenRecord.ClientID,
		&tokenRecord.UserID, &tokenRecord.Scope, &tokenRecord.ExpiresAt, &tokenRecord.RevokedAt,
		&tokenRecord.ParentTokenID, &tokenRecord.CreatedAt, &tokenRecord.ClientName,
		&tokenRecord.OrganizationID, &tokenRecord.UserEmail, &tokenRecord.UserRole)
	if err != nil {
		if err == sql.ErrNoRows {
			return &types.IntrospectionResponse{Active: false}, nil
		}
		return nil, fmt.Errorf("failed to query token: %w", err)
	}

	// Check if token is active
	if !tokenRecord.IsActive() {
		return &types.IntrospectionResponse{Active: false}, nil
	}

	response := &types.IntrospectionResponse{
		Active:    true,
		Scope:     tokenRecord.Scope,
		ClientID:  tokenRecord.ClientID,
		TokenType: "Bearer",
		Exp:       tokenRecord.ExpiresAt.Unix(),
		Iat:       tokenRecord.CreatedAt.Unix(),
		Iss:       s.issuer,
	}

	if tokenRecord.UserID != nil {
		response.Sub = *tokenRecord.UserID
		if tokenRecord.UserEmail != nil {
			response.Username = *tokenRecord.UserEmail
		}
	}

	return response, nil
}

// RevokeToken revokes an OAuth token
func (s *OAuthService) RevokeToken(ctx context.Context, token string, clientID string, clientSecret string) error {
	// Authenticate client
	if _, err := s.authenticateClient(ctx, clientID, clientSecret); err != nil {
		return fmt.Errorf("client authentication failed: %w", err)
	}

	tokenHash := hashToken(token)

	// Revoke the token
	query := `
		UPDATE oauth_tokens 
		SET revoked_at = NOW() 
		WHERE token_hash = $1 AND client_id = $2 AND revoked_at IS NULL`

	result, err := s.db.ExecContext(ctx, query, tokenHash, clientID)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("token not found or already revoked")
	}

	return nil
}

// ValidateToken validates a Bearer token and returns token info
func (s *OAuthService) ValidateToken(ctx context.Context, bearerToken string) (*types.OAuthToken, error) {
	// Remove "Bearer " prefix if present
	token := strings.TrimPrefix(bearerToken, "Bearer ")
	token = strings.TrimSpace(token)

	tokenHash := hashToken(token)

	var tokenRecord types.OAuthToken
	query := `
		SELECT t.id, t.token_hash, t.token_type, t.client_id, t.user_id, t.scope, 
			   t.expires_at, t.revoked_at, t.parent_token_id, t.created_at,
			   c.client_name, c.organization_id, u.email as user_email, u.role as user_role
		FROM oauth_tokens t
		JOIN oauth_clients c ON t.client_id = c.client_id
		LEFT JOIN users u ON t.user_id = u.id
		WHERE t.token_hash = $1 AND t.revoked_at IS NULL AND t.expires_at > NOW()`

	err := s.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&tokenRecord.ID, &tokenRecord.TokenHash, &tokenRecord.TokenType, &tokenRecord.ClientID,
		&tokenRecord.UserID, &tokenRecord.Scope, &tokenRecord.ExpiresAt, &tokenRecord.RevokedAt,
		&tokenRecord.ParentTokenID, &tokenRecord.CreatedAt, &tokenRecord.ClientName,
		&tokenRecord.OrganizationID, &tokenRecord.UserEmail, &tokenRecord.UserRole)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid or expired token")
		}
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	return &tokenRecord, nil
}

// GetClient retrieves an OAuth client by client ID
func (s *OAuthService) GetClient(ctx context.Context, clientID string) (*types.OAuthClient, error) {
	var client types.OAuthClient
	query := `
		SELECT id, client_id, client_secret_hash, client_name, client_type,
			   redirect_uris, grant_types, response_types, scope, contacts,
			   logo_uri, client_uri, policy_uri, tos_uri, jwks_uri,
			   token_endpoint_auth_method, organization_id, is_active,
			   created_at, updated_at
		FROM oauth_clients 
		WHERE client_id = $1 AND is_active = true`

	err := s.db.QueryRowContext(ctx, query, clientID).Scan(
		&client.ID, &client.ClientID, &client.ClientSecretHash, &client.ClientName, &client.ClientType,
		pq.Array(&client.RedirectURIs), pq.Array(&client.GrantTypes), pq.Array(&client.ResponseTypes), &client.Scope, pq.Array(&client.Contacts),
		&client.LogoURI, &client.ClientURI, &client.PolicyURI, &client.TOSURI, &client.JWKSURI,
		&client.TokenEndpointAuthMethod, &client.OrganizationID, &client.IsActive,
		&client.CreatedAt, &client.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("client not found")
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	return &client, nil
}

// Helper functions

// authenticateClient authenticates a client using client credentials
func (s *OAuthService) authenticateClient(ctx context.Context, clientID, clientSecret string) (*types.OAuthClient, error) {
	client, err := s.GetClient(ctx, clientID)
	if err != nil {
		return nil, err
	}

	// Public clients don't need a secret
	if client.TokenEndpointAuthMethod == types.TokenEndpointAuthNone {
		if clientSecret != "" {
			return nil, fmt.Errorf("public client should not provide client_secret")
		}
		return client, nil
	}

	// Confidential clients need a secret
	if client.ClientSecretHash == nil {
		return nil, fmt.Errorf("client secret hash not found")
	}

	if clientSecret == "" {
		return nil, fmt.Errorf("client_secret required for confidential client")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*client.ClientSecretHash), []byte(clientSecret)); err != nil {
		return nil, fmt.Errorf("invalid client credentials")
	}

	return client, nil
}

// generateAccessToken creates a JWT access token
func (s *OAuthService) generateAccessToken(clientID, userID, scope string, expiresAt time.Time) (string, error) {
	claims := jwt.MapClaims{
		"iss":       s.issuer,
		"aud":       s.issuer,
		"sub":       clientID,
		"client_id": clientID,
		"scope":     scope,
		"iat":       time.Now().Unix(),
		"exp":       expiresAt.Unix(),
		"token_use": "access",
	}

	if userID != "" {
		claims["user_id"] = userID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return "client_" + generateRandomString(16)
}

// generateClientSecret generates a secure client secret
func generateClientSecret() string {
	return "secret_" + generateRandomString(32)
}

// generateRandomString generates a random string of specified length
func generateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to UUID if random generation fails
		return strings.ReplaceAll(uuid.New().String(), "-", "")
	}
	return hex.EncodeToString(bytes)[:length]
}

// hashToken creates a SHA256 hash of the token for database storage
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}

// stringPtr returns a pointer to the string if not empty
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// stringPtrIfNotEmpty returns a pointer to string if not empty, nil otherwise
func stringPtrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
