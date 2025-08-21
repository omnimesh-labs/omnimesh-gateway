package middleware

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"mcp-gateway/apps/backend/internal/types"

	"github.com/gin-gonic/gin"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// PathRewriteMiddleware handles path rewriting for server-specific endpoints
type PathRewriteMiddleware struct {
	rules []types.PathRewriteRule
}

// NewPathRewriteMiddleware creates a new path rewrite middleware
func NewPathRewriteMiddleware() *PathRewriteMiddleware {
	return &PathRewriteMiddleware{
		rules: getDefaultRewriteRules(),
	}
}

// getDefaultRewriteRules returns the default path rewrite rules
func getDefaultRewriteRules() []types.PathRewriteRule {
	return []types.PathRewriteRule{
		{
			Pattern:     `^/servers/([^/]+)/mcp(.*)$`,
			Replacement: "/mcp$2",
			Headers: map[string]string{
				"X-MCP-Server-ID": "$1",
			},
			Context: map[string]string{
				"server_id":     "$1",
				"original_path": "$0",
			},
		},
		{
			Pattern:     `^/servers/([^/]+)/sse(.*)$`,
			Replacement: "/sse$2",
			Headers: map[string]string{
				"X-MCP-Server-ID": "$1",
			},
			Context: map[string]string{
				"server_id":     "$1",
				"original_path": "$0",
			},
		},
		{
			Pattern:     `^/servers/([^/]+)/ws(.*)$`,
			Replacement: "/ws$2",
			Headers: map[string]string{
				"X-MCP-Server-ID": "$1",
			},
			Context: map[string]string{
				"server_id":     "$1",
				"original_path": "$0",
			},
		},
		{
			Pattern:     `^/servers/([^/]+)/rpc(.*)$`,
			Replacement: "/rpc$2",
			Headers: map[string]string{
				"X-MCP-Server-ID": "$1",
			},
			Context: map[string]string{
				"server_id":     "$1",
				"original_path": "$0",
			},
		},
	}
}

// Handler returns the Gin middleware handler
func (p *PathRewriteMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		originalPath := c.Request.URL.Path
		rewritten := false

		// Apply rewrite rules
		for _, rule := range p.rules {
			if newPath, matches := p.applyRule(rule, originalPath); newPath != originalPath {
				// Rewrite the path
				c.Request.URL.Path = newPath
				rewritten = true

				// Add headers from rule
				for key, value := range rule.Headers {
					headerValue := p.substituteMatches(value, matches)
					c.Header(key, headerValue)
				}

				// Add context from rule
				for key, value := range rule.Context {
					contextValue := p.substituteMatches(value, matches)
					c.Set(key, contextValue)
				}

				// Store rewrite information
				c.Set("path_rewritten", true)
				c.Set("original_path", originalPath)
				c.Set("rewritten_path", newPath)
				c.Set("rewrite_rule", rule.Pattern)

				break // Apply only the first matching rule
			}
		}

		if !rewritten {
			c.Set("path_rewritten", false)
		}

		c.Next()
	}
}

// applyRule applies a single rewrite rule to a path
func (p *PathRewriteMiddleware) applyRule(rule types.PathRewriteRule, path string) (string, []string) {
	regex, err := regexp.Compile(rule.Pattern)
	if err != nil {
		return path, nil
	}

	matches := regex.FindStringSubmatch(path)
	if matches == nil {
		return path, nil
	}

	// Perform substitution
	newPath := rule.Replacement
	for i, match := range matches {
		placeholder := "$" + string(rune('0'+i))
		newPath = strings.ReplaceAll(newPath, placeholder, match)
	}

	return newPath, matches
}

// substituteMatches substitutes regex matches in a string
func (p *PathRewriteMiddleware) substituteMatches(value string, matches []string) string {
	result := value
	for i, match := range matches {
		placeholder := "$" + string(rune('0'+i))
		result = strings.ReplaceAll(result, placeholder, match)
	}
	return result
}

// AddRule adds a custom rewrite rule
func (p *PathRewriteMiddleware) AddRule(rule types.PathRewriteRule) {
	p.rules = append(p.rules, rule)
}

// GetRules returns all rewrite rules
func (p *PathRewriteMiddleware) GetRules() []types.PathRewriteRule {
	return p.rules
}

// ServerContextMiddleware extracts server context from rewritten paths
func ServerContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract server ID from context or headers
		var serverID string

		// Check if path was rewritten and server_id was set
		if id, exists := c.Get("server_id"); exists {
			if strID, ok := id.(string); ok {
				serverID = strID
			}
		}

		// Check X-MCP-Server-ID header
		if serverID == "" {
			serverID = c.GetHeader("X-MCP-Server-ID")
		}

		// Extract from URL parameters if present
		if serverID == "" {
			serverID = c.Param("server_id")
		}

		// Create transport context (either with server ID or default for direct transport endpoints)
		transportCtx := &types.TransportContext{
			Request:  c.Request,
			ServerID: serverID, // Will be empty for direct transport endpoints
			Metadata: make(map[string]interface{}),
		}

		// Extract user info from JWT or session
		if userID, exists := c.Get("user_id"); exists {
			if strUserID, ok := userID.(string); ok {
				transportCtx.UserID = strUserID
			}
		}

		if orgID, exists := c.Get("organization_id"); exists {
			if strOrgID, ok := orgID.(string); ok {
				transportCtx.OrganizationID = strOrgID
			}
		}

		// For direct transport endpoints without server context, use default values
		if serverID == "" {
			transportCtx.UserID = "default-user"
			transportCtx.OrganizationID = "default-org"
			transportCtx.ServerID = "default-server"
		}

		// Store transport context
		c.Set("transport_context", transportCtx)

		// Store in request context for downstream handlers
		ctx := context.WithValue(c.Request.Context(), contextKey("transport_context"), transportCtx)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// TransportTypeMiddleware determines the transport type from the request
func TransportTypeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var transportType types.TransportType

		// Determine transport type based on path and headers
		path := c.Request.URL.Path

		switch {
		case strings.HasPrefix(path, "/rpc"):
			transportType = types.TransportTypeHTTP
		case strings.HasPrefix(path, "/sse"):
			transportType = types.TransportTypeSSE
		case strings.HasPrefix(path, "/ws"):
			transportType = types.TransportTypeWebSocket
		case strings.HasPrefix(path, "/stdio"):
			transportType = types.TransportTypeSTDIO
		case strings.HasPrefix(path, "/mcp"):
			// Check Accept header to determine streamable mode
			accept := c.GetHeader("Accept")
			if strings.Contains(accept, "text/event-stream") {
				transportType = types.TransportTypeStreamable
			} else {
				transportType = types.TransportTypeStreamable
			}
		default:
			transportType = types.TransportTypeHTTP // Default fallback
		}

		// Store transport type
		c.Set("transport_type", transportType)

		// Update transport context if it exists
		if ctx, exists := c.Get("transport_context"); exists {
			if transportCtx, ok := ctx.(*types.TransportContext); ok {
				transportCtx.Transport = transportType
			}
		}

		c.Next()
	}
}

// SessionIDMiddleware extracts or generates session IDs for stateful transports
func SessionIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var sessionID string

		// Check various sources for session ID
		sessionID = c.GetHeader("X-Session-ID")
		if sessionID == "" {
			sessionID = c.Query("session_id")
		}
		if sessionID == "" {
			sessionID = c.GetHeader("Authorization") // Could extract from JWT
		}

		// For stateful transports, generate session ID if not present
		transportType, exists := c.Get("transport_type")
		if exists {
			if tt, ok := transportType.(types.TransportType); ok {
				isStateful := tt == types.TransportTypeSSE ||
					tt == types.TransportTypeWebSocket ||
					tt == types.TransportTypeStreamable

				if isStateful && sessionID == "" {
					// Generate new session ID
					sessionID = generateSessionID()
					c.Header("X-Session-ID", sessionID)
				}
			}
		}

		// Store session ID
		if sessionID != "" {
			c.Set("session_id", sessionID)

			// Update transport context if it exists
			if ctx, exists := c.Get("transport_context"); exists {
				if transportCtx, ok := ctx.(*types.TransportContext); ok {
					transportCtx.SessionID = sessionID
				}
			}
		}

		c.Next()
	}
}

// generateSessionID generates a new session ID
func generateSessionID() string {
	// This could be more sophisticated, using UUID, timestamp, etc.
	return "session_" + strings.ReplaceAll(
		strings.ReplaceAll(
			regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(
				regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`).ReplaceAllString(
					strings.ToLower(strings.ReplaceAll(
						"2024-01-01T00:00:00",
						":",
						"",
					)),
					"",
				),
				"",
			),
			"-",
			"",
		),
		"_",
		"",
	) + "_" + generateRandomString(8)
}

// generateRandomString generates a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)] // Simple deterministic for now
	}
	return string(result)
}

// PathRewriteConfig represents configuration for path rewriting
type PathRewriteConfig struct {
	Rules    []types.PathRewriteRule `yaml:"rules" json:"rules"`
	Enabled  bool                    `yaml:"enabled" json:"enabled"`
	LogLevel string                  `yaml:"log_level" json:"log_level"`
}

// ValidateConfig validates the path rewrite configuration
func (c *PathRewriteConfig) ValidateConfig() error {
	for i, rule := range c.Rules {
		// Validate regex pattern
		if _, err := regexp.Compile(rule.Pattern); err != nil {
			return fmt.Errorf("invalid regex pattern in rule %d: %w", i, err)
		}

		// Validate replacement string
		if rule.Replacement == "" {
			return fmt.Errorf("empty replacement in rule %d", i)
		}
	}
	return nil
}

// GetRewriteInfo extracts rewrite information from Gin context
func GetRewriteInfo(c *gin.Context) (bool, string, string) {
	rewritten, _ := c.Get("path_rewritten")
	originalPath, _ := c.Get("original_path")
	rewrittenPath, _ := c.Get("rewritten_path")

	isRewritten, _ := rewritten.(bool)
	origPath, _ := originalPath.(string)
	newPath, _ := rewrittenPath.(string)

	return isRewritten, origPath, newPath
}

// GetTransportContext extracts transport context from Gin context
func GetTransportContext(c *gin.Context) *types.TransportContext {
	if ctx, exists := c.Get("transport_context"); exists {
		if transportCtx, ok := ctx.(*types.TransportContext); ok {
			return transportCtx
		}
	}
	return nil
}

// GetServerID extracts server ID from Gin context
func GetServerID(c *gin.Context) string {
	if ctx := GetTransportContext(c); ctx != nil {
		return ctx.ServerID
	}

	if serverID, exists := c.Get("server_id"); exists {
		if strID, ok := serverID.(string); ok {
			return strID
		}
	}

	return ""
}

// GetSessionID extracts session ID from Gin context
func GetSessionID(c *gin.Context) string {
	if ctx := GetTransportContext(c); ctx != nil {
		return ctx.SessionID
	}

	if sessionID, exists := c.Get("session_id"); exists {
		if strID, ok := sessionID.(string); ok {
			return strID
		}
	}

	return ""
}

// GetTransportType extracts transport type from Gin context
func GetTransportType(c *gin.Context) types.TransportType {
	if ctx := GetTransportContext(c); ctx != nil {
		return ctx.Transport
	}

	if transportType, exists := c.Get("transport_type"); exists {
		if tt, ok := transportType.(types.TransportType); ok {
			return tt
		}
	}

	return types.TransportTypeHTTP // Default fallback
}
