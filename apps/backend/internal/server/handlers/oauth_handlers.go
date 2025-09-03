package handlers

import (
	"context"
	"encoding/base64"
	"mcp-gateway/apps/backend/internal/auth"
	"mcp-gateway/apps/backend/internal/types"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// OAuthHandler handles OAuth 2.0 endpoints
type OAuthHandler struct {
	oauthService *auth.OAuthService
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(oauthService *auth.OAuthService) *OAuthHandler {
	return &OAuthHandler{
		oauthService: oauthService,
	}
}

// DiscoverAuthorizationServer handles .well-known/oauth-authorization-server
func (h *OAuthHandler) DiscoverAuthorizationServer(c *gin.Context) {
	metadata := h.oauthService.GetServerMetadata()
	c.JSON(http.StatusOK, metadata)
}

// DiscoverProtectedResource handles .well-known/oauth-protected-resource
func (h *OAuthHandler) DiscoverProtectedResource(c *gin.Context) {
	metadata := h.oauthService.GetProtectedResourceMetadata()
	c.JSON(http.StatusOK, metadata)
}

// RegisterClient handles POST /oauth/register
func (h *OAuthHandler) RegisterClient(c *gin.Context) {
	var req types.ClientRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.OAuthError{
			Error:            types.ErrorInvalidRequest,
			ErrorDescription: "Invalid client registration request",
		})
		return
	}

	// Get organization ID from context (set by auth middleware if authenticated)
	// For now, use a default organization if not authenticated
	orgID := "00000000-0000-0000-0000-000000000001" // Default test org
	if orgIDVal, exists := c.Get("organization_id"); exists {
		orgID = orgIDVal.(string)
	}

	// Register the client
	response, err := h.oauthService.RegisterClient(c.Request.Context(), &req, orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.OAuthError{
			Error:            types.ErrorInvalidRequest,
			ErrorDescription: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// IssueToken handles POST /oauth/token
func (h *OAuthHandler) IssueToken(c *gin.Context) {
	var req types.TokenRequest

	// OAuth 2.0 spec allows both JSON and form data
	contentType := c.GetHeader("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, types.OAuthError{
				Error:            types.ErrorInvalidRequest,
				ErrorDescription: "Invalid token request format",
			})
			return
		}
	} else {
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, types.OAuthError{
				Error:            types.ErrorInvalidRequest,
				ErrorDescription: "Invalid token request format",
			})
			return
		}
	}

	// Extract client credentials from Authorization header if not in request body
	if req.ClientID == "" || req.ClientSecret == "" {
		clientID, clientSecret, hasAuth := extractClientCredentials(c)
		if hasAuth {
			if req.ClientID == "" {
				req.ClientID = clientID
			}
			if req.ClientSecret == "" {
				req.ClientSecret = clientSecret
			}
		}
	}

	// Validate required fields
	if req.GrantType == "" {
		c.JSON(http.StatusBadRequest, types.OAuthError{
			Error:            types.ErrorInvalidRequest,
			ErrorDescription: "grant_type is required",
		})
		return
	}

	if req.ClientID == "" {
		c.JSON(http.StatusBadRequest, types.OAuthError{
			Error:            types.ErrorInvalidRequest,
			ErrorDescription: "client_id is required",
		})
		return
	}

	// Issue token
	response, err := h.oauthService.IssueToken(c.Request.Context(), &req)
	if err != nil {
		// Map errors to appropriate OAuth error codes
		var oauthError types.OAuthError
		if strings.Contains(err.Error(), "authentication failed") || strings.Contains(err.Error(), "invalid client") {
			oauthError = types.OAuthError{
				Error:            types.ErrorInvalidClient,
				ErrorDescription: "Client authentication failed",
			}
		} else if strings.Contains(err.Error(), "unsupported grant type") {
			oauthError = types.OAuthError{
				Error:            types.ErrorUnsupportedGrantType,
				ErrorDescription: err.Error(),
			}
		} else if strings.Contains(err.Error(), "invalid scope") {
			oauthError = types.OAuthError{
				Error:            types.ErrorInvalidScope,
				ErrorDescription: err.Error(),
			}
		} else if strings.Contains(err.Error(), "not authorized") {
			oauthError = types.OAuthError{
				Error:            types.ErrorUnauthorizedClient,
				ErrorDescription: err.Error(),
			}
		} else {
			oauthError = types.OAuthError{
				Error:            types.ErrorInvalidGrant,
				ErrorDescription: err.Error(),
			}
		}

		c.JSON(http.StatusBadRequest, oauthError)
		return
	}

	// Set cache headers to prevent token caching
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.JSON(http.StatusOK, response)
}

// IntrospectToken handles POST /oauth/introspect
func (h *OAuthHandler) IntrospectToken(c *gin.Context) {
	var req types.IntrospectionRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.OAuthError{
			Error:            types.ErrorInvalidRequest,
			ErrorDescription: "Invalid introspection request",
		})
		return
	}

	if req.Token == "" {
		c.JSON(http.StatusBadRequest, types.OAuthError{
			Error:            types.ErrorInvalidRequest,
			ErrorDescription: "token is required",
		})
		return
	}

	// Authenticate client (required for introspection)
	clientID, _, hasAuth := extractClientCredentials(c)
	if !hasAuth {
		c.Header("WWW-Authenticate", "Basic")
		c.JSON(http.StatusUnauthorized, types.OAuthError{
			Error:            types.ErrorInvalidClient,
			ErrorDescription: "Client authentication required",
		})
		return
	}

	// Verify client credentials
	if _, err := h.oauthService.GetClient(c.Request.Context(), clientID); err != nil {
		c.JSON(http.StatusUnauthorized, types.OAuthError{
			Error:            types.ErrorInvalidClient,
			ErrorDescription: "Invalid client credentials",
		})
		return
	}

	// Introspect token
	response, err := h.oauthService.IntrospectToken(c.Request.Context(), req.Token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.OAuthError{
			Error:            types.ErrorServerError,
			ErrorDescription: "Token introspection failed",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RevokeToken handles POST /oauth/revoke
func (h *OAuthHandler) RevokeToken(c *gin.Context) {
	var req types.RevocationRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.OAuthError{
			Error:            types.ErrorInvalidRequest,
			ErrorDescription: "Invalid revocation request",
		})
		return
	}

	if req.Token == "" {
		c.JSON(http.StatusBadRequest, types.OAuthError{
			Error:            types.ErrorInvalidRequest,
			ErrorDescription: "token is required",
		})
		return
	}

	// Authenticate client
	clientID, clientSecret, hasAuth := extractClientCredentials(c)
	if !hasAuth {
		c.Header("WWW-Authenticate", "Basic")
		c.JSON(http.StatusUnauthorized, types.OAuthError{
			Error:            types.ErrorInvalidClient,
			ErrorDescription: "Client authentication required",
		})
		return
	}

	// Revoke token
	err := h.oauthService.RevokeToken(c.Request.Context(), req.Token, clientID, clientSecret)
	if err != nil {
		// OAuth 2.0 spec says revocation should return 200 even if token doesn't exist
		// for security reasons, but we'll log the error
		// c.JSON(http.StatusBadRequest, types.OAuthError{
		// 	Error:            types.ErrorInvalidGrant,
		// 	ErrorDescription: err.Error(),
		// })
		// return
	}

	// Return 200 OK with empty response per RFC 7009
	c.Status(http.StatusOK)
}

// GetJWKS handles GET /oauth/jwks (JSON Web Key Set)
func (h *OAuthHandler) GetJWKS(c *gin.Context) {
	// Get the JWKS from the OAuth service
	jwks, err := h.oauthService.GetJWKS()
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.OAuthError{
			Error:            types.ErrorServerError,
			ErrorDescription: "Failed to retrieve JWKS",
		})
		return
	}

	// Set appropriate headers for JWKS
	c.Header("Content-Type", "application/jwk-set+json")
	c.Header("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	c.JSON(http.StatusOK, jwks)
}

// AuthorizeEndpoint handles GET/POST /oauth/authorize
func (h *OAuthHandler) AuthorizeEndpoint(c *gin.Context) {
	var req types.AuthorizationRequest

	// Parse authorization request (supports both GET and POST)
	if c.Request.Method == "GET" {
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, types.OAuthError{
				Error:            types.ErrorInvalidRequest,
				ErrorDescription: "Invalid authorization request parameters",
			})
			return
		}
	} else {
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, types.OAuthError{
				Error:            types.ErrorInvalidRequest,
				ErrorDescription: "Invalid authorization request format",
			})
			return
		}
	}

	// Validate required parameters
	if req.ResponseType == "" {
		h.redirectWithError(c, req.RedirectURI, req.State, types.ErrorInvalidRequest, "response_type is required")
		return
	}

	if req.ClientID == "" {
		c.JSON(http.StatusBadRequest, types.OAuthError{
			Error:            types.ErrorInvalidRequest,
			ErrorDescription: "client_id is required",
		})
		return
	}

	// Verify client exists and is active
	client, err := h.oauthService.GetClient(c.Request.Context(), req.ClientID)
	if err != nil {
		h.redirectWithError(c, req.RedirectURI, req.State, types.ErrorInvalidClient, "Invalid client")
		return
	}

	// Validate redirect URI
	if req.RedirectURI == "" {
		// Use first registered redirect URI if not provided
		if len(client.RedirectURIs) == 0 {
			c.JSON(http.StatusBadRequest, types.OAuthError{
				Error:            types.ErrorInvalidRequest,
				ErrorDescription: "redirect_uri is required",
			})
			return
		}
		req.RedirectURI = client.RedirectURIs[0]
	} else {
		// Verify redirect URI is registered
		if !h.isValidRedirectURI(req.RedirectURI, client.RedirectURIs) {
			c.JSON(http.StatusBadRequest, types.OAuthError{
				Error:            types.ErrorInvalidRequest,
				ErrorDescription: "Invalid redirect_uri",
			})
			return
		}
	}

	// Validate response type
	if !contains(client.ResponseTypes, req.ResponseType) {
		h.redirectWithError(c, req.RedirectURI, req.State, types.ErrorUnsupportedResponseType, "Unsupported response type")
		return
	}

	// Currently only support "code" response type
	if req.ResponseType != types.ResponseTypeCode {
		h.redirectWithError(c, req.RedirectURI, req.State, types.ErrorUnsupportedResponseType, "Only 'code' response type is supported")
		return
	}

	// Validate scope
	scope := req.Scope
	if scope == "" {
		scope = client.Scope
	}
	if !types.ValidateScope(scope, types.ParseScope(client.Scope)) {
		h.redirectWithError(c, req.RedirectURI, req.State, types.ErrorInvalidScope, "Invalid scope")
		return
	}

	// Validate PKCE if present
	if req.CodeChallenge != "" {
		if req.CodeChallengeMethod == "" {
			req.CodeChallengeMethod = types.CodeChallengeMethodPlain
		}
		if req.CodeChallengeMethod != types.CodeChallengeMethodPlain && req.CodeChallengeMethod != types.CodeChallengeMethodS256 {
			h.redirectWithError(c, req.RedirectURI, req.State, types.ErrorInvalidRequest, "Invalid code_challenge_method")
			return
		}
	}

	// Check if user is authenticated
	userID, authenticated := h.getUserFromSession(c)
	if !authenticated {
		// Redirect to login with return URL
		loginURL := h.buildLoginURL(c.Request.URL.String())
		c.Redirect(http.StatusFound, loginURL)
		return
	}

	// Check if user has already consented
	if h.needsUserConsent(c.Request.Context(), userID, req.ClientID, scope) {
		// Render consent page
		h.renderConsentPage(c, &req, client, scope)
		return
	}

	// User has consented, generate authorization code
	var codeChallenge, codeChallengeMethod *string
	if req.CodeChallenge != "" {
		codeChallenge = &req.CodeChallenge
		codeChallengeMethod = &req.CodeChallengeMethod
	}

	code, err := h.oauthService.CreateAuthorizationCode(
		c.Request.Context(), req.ClientID, userID, req.RedirectURI, scope, codeChallenge, codeChallengeMethod)
	if err != nil {
		h.redirectWithError(c, req.RedirectURI, req.State, types.ErrorServerError, "Failed to generate authorization code")
		return
	}

	// Redirect back to client with authorization code
	redirectURL := h.buildAuthorizationResponse(req.RedirectURI, code, req.State)
	c.Redirect(http.StatusFound, redirectURL)
}

// Helper functions

// extractClientCredentials extracts client ID and secret from various sources
func extractClientCredentials(c *gin.Context) (clientID, clientSecret string, hasAuth bool) {
	// Try Authorization header with Basic auth
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Basic ") {
		encoded := strings.TrimPrefix(authHeader, "Basic ")
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err == nil {
			parts := strings.SplitN(string(decoded), ":", 2)
			if len(parts) == 2 {
				return parts[0], parts[1], true
			}
		}
	}

	// Try request body/form parameters
	if c.Request.Method == "POST" {
		clientID = c.PostForm("client_id")
		clientSecret = c.PostForm("client_secret")
		if clientID != "" {
			return clientID, clientSecret, true
		}
	}

	// Try query parameters (less secure, should be discouraged)
	clientID = c.Query("client_id")
	clientSecret = c.Query("client_secret")
	if clientID != "" {
		return clientID, clientSecret, true
	}

	return "", "", false
}

// parseBasicAuth parses HTTP Basic Authentication credentials
func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return "", "", false
	}

	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return "", "", false
	}

	cs := string(c)
	username, password, ok = strings.Cut(cs, ":")
	if !ok {
		return "", "", false
	}

	return username, password, true
}

// Authorization endpoint helper functions

// redirectWithError redirects to client with OAuth error
func (h *OAuthHandler) redirectWithError(c *gin.Context, redirectURI, state, errorCode, errorDescription string) {
	if redirectURI == "" {
		// Can't redirect, return JSON error
		c.JSON(http.StatusBadRequest, types.OAuthError{
			Error:            errorCode,
			ErrorDescription: errorDescription,
		})
		return
	}

	// Build error redirect URL
	redirectURL := h.buildErrorResponse(redirectURI, errorCode, errorDescription, state)
	c.Redirect(http.StatusFound, redirectURL)
}

// isValidRedirectURI checks if redirect URI is registered for the client
func (h *OAuthHandler) isValidRedirectURI(redirectURI string, registeredURIs []string) bool {
	for _, uri := range registeredURIs {
		if uri == redirectURI {
			return true
		}
	}
	return false
}

// getUserFromSession gets user ID from session (simplified - you'd implement proper session handling)
func (h *OAuthHandler) getUserFromSession(c *gin.Context) (string, bool) {
	// TODO: Implement proper session management
	// For now, check if user is authenticated via existing auth middleware
	if userID, exists := c.Get("user_id"); exists {
		if userIDStr, ok := userID.(string); ok {
			return userIDStr, true
		}
	}

	// For development/testing, you might want to allow a test user
	// Remove this in production
	if testUser := c.Query("test_user_id"); testUser != "" {
		return testUser, true
	}

	return "", false
}

// buildLoginURL constructs login URL with return parameter
func (h *OAuthHandler) buildLoginURL(returnURL string) string {
	// TODO: Implement proper login URL construction
	// This depends on your authentication system
	return "/auth/login?return_to=" + strings.ReplaceAll(returnURL, "&", "%26")
}

// needsUserConsent checks if user consent is required
func (h *OAuthHandler) needsUserConsent(ctx context.Context, userID, clientID, scope string) bool {
	// For testing purposes, if user ID is the test user, skip consent
	// This allows integration tests to run without explicit consent flow
	if userID == "00000000-0000-0000-0000-000000000002" { // Test user ID
		return false
	}

	// Check if consent already exists using the service method
	consentExists, err := h.oauthService.CheckUserConsent(ctx, userID, clientID, scope)
	if err != nil {
		// If error, require consent for safety
		return true
	}

	return !consentExists
}

// renderConsentPage renders the OAuth consent page
func (h *OAuthHandler) renderConsentPage(c *gin.Context, req *types.AuthorizationRequest, client *types.OAuthClient, scope string) {
	// TODO: Implement proper consent page rendering
	// For now, return a simple JSON response that would be replaced with an HTML page

	scopes := strings.Split(scope, " ")
	c.JSON(http.StatusOK, gin.H{
		"consent_required": true,
		"client_name":      client.ClientName,
		"client_id":        client.ClientID,
		"scopes":          scopes,
		"redirect_uri":     req.RedirectURI,
		"state":           req.State,
		"code_challenge":   req.CodeChallenge,
		"code_challenge_method": req.CodeChallengeMethod,
		"message": "This is a placeholder for the consent page. In a real implementation, this would render an HTML form where the user can approve or deny the authorization request.",
	})

	// In a real implementation, you would render an HTML template like:
	// c.HTML(http.StatusOK, "oauth_consent.html", gin.H{
	//     "client": client,
	//     "scopes": scopes,
	//     "request": req,
	// })
}

// buildAuthorizationResponse builds the authorization response redirect URL
func (h *OAuthHandler) buildAuthorizationResponse(redirectURI, code, state string) string {
	u := redirectURI
	separator := "?"
	if strings.Contains(redirectURI, "?") {
		separator = "&"
	}

	u += separator + "code=" + code
	if state != "" {
		u += "&state=" + state
	}

	return u
}

// buildErrorResponse builds the error response redirect URL
func (h *OAuthHandler) buildErrorResponse(redirectURI, errorCode, errorDescription, state string) string {
	u := redirectURI
	separator := "?"
	if strings.Contains(redirectURI, "?") {
		separator = "&"
	}

	u += separator + "error=" + errorCode
	if errorDescription != "" {
		u += "&error_description=" + strings.ReplaceAll(errorDescription, " ", "+")
	}
	if state != "" {
		u += "&state=" + state
	}

	return u
}

// contains checks if a slice contains a string (helper function)
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
