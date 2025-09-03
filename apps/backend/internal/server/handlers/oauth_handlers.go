package handlers

import (
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
	orgID := "00000000-0000-0000-0000-000000000000" // Default org
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
	// TODO: Implement JWKS endpoint for JWT signature verification
	// For now, return empty key set
	c.JSON(http.StatusOK, gin.H{
		"keys": []gin.H{},
	})
}

// AuthorizeEndpoint handles GET/POST /oauth/authorize
func (h *OAuthHandler) AuthorizeEndpoint(c *gin.Context) {
	// TODO: Implement authorization endpoint for authorization code flow
	c.JSON(http.StatusNotImplemented, types.OAuthError{
		Error:            types.ErrorServerError,
		ErrorDescription: "Authorization endpoint not yet implemented",
	})
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
