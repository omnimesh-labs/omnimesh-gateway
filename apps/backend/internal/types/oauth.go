package types

import (
	"strings"
	"time"
)

// OAuth 2.0 Types and Constants

// OAuth Grant Types
const (
	GrantTypeClientCredentials = "client_credentials"
	GrantTypeAuthorizationCode = "authorization_code"
	GrantTypeRefreshToken      = "refresh_token"
	GrantTypeDeviceCode        = "urn:ietf:params:oauth:grant-type:device_code"
)

// OAuth Response Types
const (
	ResponseTypeCode  = "code"
	ResponseTypeToken = "token"
)

// OAuth Token Types
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// OAuth Client Types
const (
	ClientTypeConfidential = "confidential"
	ClientTypePublic       = "public"
)

// OAuth Token Endpoint Authentication Methods
const (
	TokenEndpointAuthClientSecretBasic = "client_secret_basic"
	TokenEndpointAuthClientSecretPost  = "client_secret_post"
	TokenEndpointAuthNone              = "none"
	TokenEndpointAuthPrivateKeyJWT     = "private_key_jwt"
	TokenEndpointAuthClientSecretJWT   = "client_secret_jwt"
)

// PKCE Code Challenge Methods
const (
	CodeChallengeMethodPlain = "plain"
	CodeChallengeMethodS256  = "S256"
)

// OAuth Scopes
const (
	ScopeRead     = "read"
	ScopeWrite    = "write"
	ScopeDelete   = "delete"
	ScopeAdmin    = "admin"
	ScopeExecute  = "execute"
	ScopeOpenID   = "openid"
	ScopeProfile  = "profile"
	ScopeEmail    = "email"
	ScopeOffline  = "offline_access"
)

// OAuth Error Codes (RFC 6749)
const (
	ErrorInvalidRequest          = "invalid_request"
	ErrorInvalidClient           = "invalid_client"
	ErrorInvalidGrant            = "invalid_grant"
	ErrorUnauthorizedClient      = "unauthorized_client"
	ErrorUnsupportedGrantType    = "unsupported_grant_type"
	ErrorUnsupportedResponseType = "unsupported_response_type"
	ErrorInvalidScope            = "invalid_scope"
	ErrorAccessDenied            = "access_denied"
	ErrorServerError             = "server_error"
	ErrorTemporarilyUnavailable  = "temporarily_unavailable"
)

// OAuth Client represents a registered OAuth 2.0 client
type OAuthClient struct {
	ID                      string    `json:"id" db:"id"`
	ClientID                string    `json:"client_id" db:"client_id"`
	ClientSecretHash        *string   `json:"-" db:"client_secret_hash"` // Hidden from JSON
	ClientName              string    `json:"client_name" db:"client_name"`
	ClientType              string    `json:"client_type" db:"client_type"`
	RedirectURIs            []string  `json:"redirect_uris" db:"redirect_uris"`
	GrantTypes              []string  `json:"grant_types" db:"grant_types"`
	ResponseTypes           []string  `json:"response_types" db:"response_types"`
	Scope                   string    `json:"scope" db:"scope"`
	Contacts                []string  `json:"contacts" db:"contacts"`
	LogoURI                 *string   `json:"logo_uri,omitempty" db:"logo_uri"`
	ClientURI               *string   `json:"client_uri,omitempty" db:"client_uri"`
	PolicyURI               *string   `json:"policy_uri,omitempty" db:"policy_uri"`
	TOSURI                  *string   `json:"tos_uri,omitempty" db:"tos_uri"`
	JWKSURI                 *string   `json:"jwks_uri,omitempty" db:"jwks_uri"`
	TokenEndpointAuthMethod string    `json:"token_endpoint_auth_method" db:"token_endpoint_auth_method"`
	OrganizationID          string    `json:"organization_id" db:"organization_id"`
	IsActive                bool      `json:"is_active" db:"is_active"`
	CreatedAt               time.Time `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time `json:"updated_at" db:"updated_at"`
}

// OAuth Token represents an OAuth access or refresh token
type OAuthToken struct {
	ID              string     `json:"id" db:"id"`
	TokenHash       string     `json:"-" db:"token_hash"` // Hidden from JSON
	TokenType       string     `json:"token_type" db:"token_type"`
	ClientID        string     `json:"client_id" db:"client_id"`
	UserID          *string    `json:"user_id,omitempty" db:"user_id"`
	Scope           string     `json:"scope" db:"scope"`
	ExpiresAt       time.Time  `json:"expires_at" db:"expires_at"`
	RevokedAt       *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
	ParentTokenID   *string    `json:"parent_token_id,omitempty" db:"parent_token_id"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`

	// Joined fields from view
	ClientName     *string `json:"client_name,omitempty" db:"client_name"`
	OrganizationID *string `json:"organization_id,omitempty" db:"organization_id"`
	UserEmail      *string `json:"user_email,omitempty" db:"user_email"`
	UserRole       *string `json:"user_role,omitempty" db:"user_role"`
}

// OAuth Authorization Code for authorization_code grant
type OAuthAuthorizationCode struct {
	ID                    string     `json:"id" db:"id"`
	CodeHash              string     `json:"-" db:"code_hash"` // Hidden from JSON
	ClientID              string     `json:"client_id" db:"client_id"`
	UserID                string     `json:"user_id" db:"user_id"`
	RedirectURI           string     `json:"redirect_uri" db:"redirect_uri"`
	Scope                 string     `json:"scope" db:"scope"`
	CodeChallenge         *string    `json:"code_challenge,omitempty" db:"code_challenge"`
	CodeChallengeMethod   *string    `json:"code_challenge_method,omitempty" db:"code_challenge_method"`
	ExpiresAt             time.Time  `json:"expires_at" db:"expires_at"`
	UsedAt                *time.Time `json:"used_at,omitempty" db:"used_at"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
}

// OAuth User Consent tracking
type OAuthUserConsent struct {
	ID        string     `json:"id" db:"id"`
	UserID    string     `json:"user_id" db:"user_id"`
	ClientID  string     `json:"client_id" db:"client_id"`
	Scope     string     `json:"scope" db:"scope"`
	GrantedAt time.Time  `json:"granted_at" db:"granted_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
}

// OAuth Discovery Responses

// AuthorizationServerMetadata represents OAuth 2.0 Authorization Server Metadata (RFC 8414)
type AuthorizationServerMetadata struct {
	Issuer                                     string   `json:"issuer"`
	AuthorizationEndpoint                      *string  `json:"authorization_endpoint,omitempty"`
	TokenEndpoint                              string   `json:"token_endpoint"`
	JWKSUri                                    *string  `json:"jwks_uri,omitempty"`
	RegistrationEndpoint                       *string  `json:"registration_endpoint,omitempty"`
	ScopesSupported                            []string `json:"scopes_supported,omitempty"`
	ResponseTypesSupported                     []string `json:"response_types_supported"`
	ResponseModesSupported                     []string `json:"response_modes_supported,omitempty"`
	GrantTypesSupported                        []string `json:"grant_types_supported"`
	TokenEndpointAuthMethodsSupported          []string `json:"token_endpoint_auth_methods_supported"`
	TokenEndpointAuthSigningAlgValuesSupported []string `json:"token_endpoint_auth_signing_alg_values_supported,omitempty"`
	ServiceDocumentation                       *string  `json:"service_documentation,omitempty"`
	UILocalesSupported                         []string `json:"ui_locales_supported,omitempty"`
	OpPolicyUri                                *string  `json:"op_policy_uri,omitempty"`
	OpTosUri                                   *string  `json:"op_tos_uri,omitempty"`
	RevocationEndpoint                         *string  `json:"revocation_endpoint,omitempty"`
	RevocationEndpointAuthMethodsSupported     []string `json:"revocation_endpoint_auth_methods_supported,omitempty"`
	IntrospectionEndpoint                      *string  `json:"introspection_endpoint,omitempty"`
	IntrospectionEndpointAuthMethodsSupported  []string `json:"introspection_endpoint_auth_methods_supported,omitempty"`
	CodeChallengeMethodsSupported              []string `json:"code_challenge_methods_supported,omitempty"`
}

// ProtectedResourceMetadata represents OAuth 2.0 Protected Resource Metadata
type ProtectedResourceMetadata struct {
	Resource                                   string   `json:"resource"`
	AuthorizationServers                       []string `json:"authorization_servers"`
	JWKSUri                                    *string  `json:"jwks_uri,omitempty"`
	ScopesSupported                            []string `json:"scopes_supported,omitempty"`
	BearerMethodsSupported                     []string `json:"bearer_methods_supported,omitempty"`
	ResourceDocumentation                      *string  `json:"resource_documentation,omitempty"`
	IntrospectionEndpoint                      *string  `json:"introspection_endpoint,omitempty"`
	IntrospectionEndpointAuthMethodsSupported  []string `json:"introspection_endpoint_auth_methods_supported,omitempty"`
}

// Client Registration Requests/Responses (RFC 7591)

// ClientRegistrationRequest represents a dynamic client registration request
type ClientRegistrationRequest struct {
	RedirectURIs                []string `json:"redirect_uris,omitempty"`
	TokenEndpointAuthMethod     string   `json:"token_endpoint_auth_method,omitempty"`
	GrantTypes                  []string `json:"grant_types,omitempty"`
	ResponseTypes               []string `json:"response_types,omitempty"`
	ClientName                  string   `json:"client_name,omitempty"`
	ClientURI                   string   `json:"client_uri,omitempty"`
	LogoURI                     string   `json:"logo_uri,omitempty"`
	Scope                       string   `json:"scope,omitempty"`
	Contacts                    []string `json:"contacts,omitempty"`
	TOSURI                      string   `json:"tos_uri,omitempty"`
	PolicyURI                   string   `json:"policy_uri,omitempty"`
	JWKSURI                     string   `json:"jwks_uri,omitempty"`
	SoftwareID                  string   `json:"software_id,omitempty"`
	SoftwareVersion             string   `json:"software_version,omitempty"`
	SoftwareStatement           string   `json:"software_statement,omitempty"`
}

// ClientRegistrationResponse represents a successful client registration response
type ClientRegistrationResponse struct {
	ClientID                    string   `json:"client_id"`
	ClientSecret                string   `json:"client_secret,omitempty"`
	ClientIdIssuedAt            int64    `json:"client_id_issued_at,omitempty"`
	ClientSecretExpiresAt       int64    `json:"client_secret_expires_at,omitempty"`
	RedirectURIs                []string `json:"redirect_uris,omitempty"`
	TokenEndpointAuthMethod     string   `json:"token_endpoint_auth_method,omitempty"`
	GrantTypes                  []string `json:"grant_types,omitempty"`
	ResponseTypes               []string `json:"response_types,omitempty"`
	ClientName                  string   `json:"client_name,omitempty"`
	ClientURI                   string   `json:"client_uri,omitempty"`
	LogoURI                     string   `json:"logo_uri,omitempty"`
	Scope                       string   `json:"scope,omitempty"`
	Contacts                    []string `json:"contacts,omitempty"`
	TOSURI                      string   `json:"tos_uri,omitempty"`
	PolicyURI                   string   `json:"policy_uri,omitempty"`
	JWKSURI                     string   `json:"jwks_uri,omitempty"`
}

// Token Request/Response

// TokenRequest represents an OAuth 2.0 token request
type TokenRequest struct {
	GrantType    string `json:"grant_type" form:"grant_type"`
	Code         string `json:"code,omitempty" form:"code"`
	RedirectURI  string `json:"redirect_uri,omitempty" form:"redirect_uri"`
	ClientID     string `json:"client_id,omitempty" form:"client_id"`
	ClientSecret string `json:"client_secret,omitempty" form:"client_secret"`
	Scope        string `json:"scope,omitempty" form:"scope"`
	RefreshToken string `json:"refresh_token,omitempty" form:"refresh_token"`
	CodeVerifier string `json:"code_verifier,omitempty" form:"code_verifier"`
}

// TokenResponse represents an OAuth 2.0 token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	IDToken      string `json:"id_token,omitempty"` // For OpenID Connect
}

// Token Introspection Request/Response (RFC 7662)

// IntrospectionRequest represents a token introspection request
type IntrospectionRequest struct {
	Token         string `json:"token" form:"token"`
	TokenTypeHint string `json:"token_type_hint,omitempty" form:"token_type_hint"`
}

// IntrospectionResponse represents a token introspection response
type IntrospectionResponse struct {
	Active     bool   `json:"active"`
	Scope      string `json:"scope,omitempty"`
	ClientID   string `json:"client_id,omitempty"`
	Username   string `json:"username,omitempty"`
	TokenType  string `json:"token_type,omitempty"`
	Exp        int64  `json:"exp,omitempty"`
	Iat        int64  `json:"iat,omitempty"`
	Nbf        int64  `json:"nbf,omitempty"`
	Sub        string `json:"sub,omitempty"`
	Aud        string `json:"aud,omitempty"`
	Iss        string `json:"iss,omitempty"`
	Jti        string `json:"jti,omitempty"`
}

// Token Revocation Request (RFC 7009)

// RevocationRequest represents a token revocation request
type RevocationRequest struct {
	Token         string `json:"token" form:"token"`
	TokenTypeHint string `json:"token_type_hint,omitempty" form:"token_type_hint"`
}

// OAuth Error Response

// OAuthError represents an OAuth 2.0 error response
type OAuthError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
}

// Authorization Request/Response

// AuthorizationRequest represents an OAuth 2.0 authorization request
type AuthorizationRequest struct {
	ResponseType         string `json:"response_type" form:"response_type"`
	ClientID             string `json:"client_id" form:"client_id"`
	RedirectURI          string `json:"redirect_uri,omitempty" form:"redirect_uri"`
	Scope                string `json:"scope,omitempty" form:"scope"`
	State                string `json:"state,omitempty" form:"state"`
	CodeChallenge        string `json:"code_challenge,omitempty" form:"code_challenge"`
	CodeChallengeMethod  string `json:"code_challenge_method,omitempty" form:"code_challenge_method"`
	ResponseMode         string `json:"response_mode,omitempty" form:"response_mode"`
	Nonce                string `json:"nonce,omitempty" form:"nonce"`
	Display              string `json:"display,omitempty" form:"display"`
	Prompt               string `json:"prompt,omitempty" form:"prompt"`
	MaxAge               int64  `json:"max_age,omitempty" form:"max_age"`
	UILocales            string `json:"ui_locales,omitempty" form:"ui_locales"`
	IDTokenHint          string `json:"id_token_hint,omitempty" form:"id_token_hint"`
	LoginHint            string `json:"login_hint,omitempty" form:"login_hint"`
	ACRValues            string `json:"acr_values,omitempty" form:"acr_values"`
}

// Helper functions for OAuth types

// IsExpired checks if a token is expired
func (t *OAuthToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsRevoked checks if a token is revoked
func (t *OAuthToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

// IsActive checks if a token is active (not expired and not revoked)
func (t *OAuthToken) IsActive() bool {
	return !t.IsExpired() && !t.IsRevoked()
}

// IsExpired checks if an authorization code is expired
func (c *OAuthAuthorizationCode) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// IsUsed checks if an authorization code has been used
func (c *OAuthAuthorizationCode) IsUsed() bool {
	return c.UsedAt != nil
}

// IsValid checks if an authorization code is valid (not expired and not used)
func (c *OAuthAuthorizationCode) IsValid() bool {
	return !c.IsExpired() && !c.IsUsed()
}

// ParseScope splits a scope string into individual scopes
func ParseScope(scope string) []string {
	if scope == "" {
		return []string{}
	}
	
	// Split by space and filter empty strings
	var scopes []string
	for _, s := range strings.Split(scope, " ") {
		if s = strings.TrimSpace(s); s != "" {
			scopes = append(scopes, s)
		}
	}
	return scopes
}

// JoinScope joins individual scopes into a scope string
func JoinScope(scopes []string) string {
	// Filter empty strings and join with space
	var validScopes []string
	for _, scope := range scopes {
		if scope = strings.TrimSpace(scope); scope != "" {
			validScopes = append(validScopes, scope)
		}
	}
	return strings.Join(validScopes, " ")
}

// ValidateScope checks if requested scopes are valid
func ValidateScope(requestedScope string, allowedScopes []string) bool {
	if requestedScope == "" {
		return true
	}
	
	requested := ParseScope(requestedScope)
	allowed := make(map[string]bool)
	for _, scope := range allowedScopes {
		allowed[scope] = true
	}
	
	for _, scope := range requested {
		if !allowed[scope] {
			return false
		}
	}
	return true
}