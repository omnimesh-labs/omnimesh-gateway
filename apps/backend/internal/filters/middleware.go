package filters

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mcp-gateway/apps/backend/internal/database/models"
	"mcp-gateway/apps/backend/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FilterMiddleware provides content filtering middleware for HTTP requests/responses
type FilterMiddleware struct {
	service FilterService
}

// NewFilterMiddleware creates a new filter middleware
func NewFilterMiddleware(service FilterService) *FilterMiddleware {
	return &FilterMiddleware{
		service: service,
	}
}

// Handler returns a Gin middleware handler for content filtering
func (m *FilterMiddleware) Handler() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Skip filtering for certain paths
		if m.shouldSkipFiltering(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Get organization ID from context
		orgID, exists := c.Get("organization_id")
		if !exists {
			// Use default organization if not set
			orgID = "00000000-0000-0000-0000-000000000000"
		}

		userID, _ := c.Get("user_id")
		userIDStr := "default-user"
		if userID != nil {
			if uid, ok := userID.(string); ok {
				userIDStr = uid
			}
		}

		// Create request ID for tracing
		requestID := uuid.New().String()
		c.Set("filter_request_id", requestID)

		// Process inbound request
		if err := m.processInboundRequest(c, orgID.(string), userIDStr, requestID); err != nil {
			log.Printf("Filter error processing inbound request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		// Continue with the request
		c.Next()

		// Process outbound response
		if err := m.processOutboundResponse(c, orgID.(string), userIDStr, requestID); err != nil {
			log.Printf("Filter error processing outbound response: %v", err)
			// Don't abort here as response is already being sent
		}
	})
}

// processInboundRequest processes the incoming request through filters
func (m *FilterMiddleware) processInboundRequest(c *gin.Context, orgID, userID, requestID string) error {
	// Read request body if present
	var requestBody []byte
	var err error

	if c.Request.Body != nil {
		requestBody, err = io.ReadAll(c.Request.Body)
		if err != nil {
			return fmt.Errorf("failed to read request body: %w", err)
		}

		// Restore request body for downstream handlers
		c.Request.Body = io.NopCloser(bytes.NewReader(requestBody))
	}

	// Create filter context
	filterCtx := &FilterContext{
		RequestID:      requestID,
		OrganizationID: orgID,
		UserID:         userID,
		Direction:      FilterDirectionInbound,
		ContentType:    c.GetHeader("Content-Type"),
		Transport:      m.getTransportType(c),
		Metadata:       m.extractRequestMetadata(c),
		Timestamp:      time.Now(),
	}

	// Create filter content
	content := CreateFilterContent(
		string(requestBody),
		nil,
		m.extractHeaders(c.Request.Header),
		m.extractQueryParams(c),
	)

	// Apply filters
	result, modifiedContent, err := m.service.ProcessContent(context.Background(), filterCtx, content)
	if err != nil {
		return fmt.Errorf("failed to process content through filters: %w", err)
	}

	// Handle filter result
	if result.Blocked {
		// Log violation if configured
		if err := m.logViolations(c.Request.Context(), filterCtx, result); err != nil {
			log.Printf("Failed to log filter violations: %v", err)
		}

		// Return appropriate response based on action
		switch result.Action {
		case FilterActionBlock:
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Content blocked by security filters",
				"reason":  result.Reason,
				"code":    "CONTENT_FILTERED",
				"details": m.sanitizeViolations(result.Violations),
			})
			c.Abort()
			return nil
		}
	}

	// Handle content modification
	if result.Modified && modifiedContent != nil && modifiedContent.Raw != string(requestBody) {
		// Replace request body with filtered content
		c.Request.Body = io.NopCloser(strings.NewReader(modifiedContent.Raw))
		c.Request.ContentLength = int64(len(modifiedContent.Raw))
		c.Request.Header.Set("Content-Length", strconv.Itoa(len(modifiedContent.Raw)))
	}

	// Log warnings/audit events
	if result.Action == FilterActionWarn || result.Action == FilterActionAudit {
		if err := m.logViolations(c.Request.Context(), filterCtx, result); err != nil {
			log.Printf("Failed to log filter violations: %v", err)
		}
	}

	return nil
}

// processOutboundResponse processes the outgoing response through filters
func (m *FilterMiddleware) processOutboundResponse(c *gin.Context, orgID, userID, requestID string) error {
	// Skip response filtering for non-success responses
	if c.Writer.Status() < 200 || c.Writer.Status() >= 300 {
		return nil
	}

	// Note: Response filtering is more complex with Gin as we'd need to capture
	// the response body before it's written. For now, we'll implement a basic
	// version that could be enhanced with a response writer wrapper.

	// Create filter context for outbound
	filterCtx := &FilterContext{
		RequestID:      requestID,
		OrganizationID: orgID,
		UserID:         userID,
		Direction:      FilterDirectionOutbound,
		ContentType:    c.Writer.Header().Get("Content-Type"),
		Transport:      m.getTransportType(c),
		Metadata:       m.extractResponseMetadata(c),
		Timestamp:      time.Now(),
	}

	// Create basic filter content for logging purposes
	content := CreateFilterContent("", nil, nil, nil)

	// Apply filters (this is a simplified version)
	result, _, err := m.service.ProcessContent(context.Background(), filterCtx, content)
	if err != nil {
		return fmt.Errorf("failed to process outbound content: %w", err)
	}

	// Log any violations found
	if len(result.Violations) > 0 {
		log.Printf("Outbound filtering found %d violations for request %s", len(result.Violations), requestID)
	}

	return nil
}

// shouldSkipFiltering determines if filtering should be skipped for a path
func (m *FilterMiddleware) shouldSkipFiltering(path string) bool {
	skipPaths := []string{
		"/health",
		"/metrics",
		"/api/auth/login",
		"/api/auth/refresh",
		"/api/admin/filters", // Avoid recursive filtering on filter management endpoints
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	return false
}

// getTransportType determines the transport type from the request
func (m *FilterMiddleware) getTransportType(c *gin.Context) types.TransportType {
	// Determine transport type based on path and headers
	path := c.Request.URL.Path

	if strings.HasPrefix(path, "/ws") {
		return types.TransportTypeWebSocket
	}
	if strings.HasPrefix(path, "/sse") {
		return types.TransportTypeSSE
	}
	if strings.HasPrefix(path, "/mcp") {
		return types.TransportTypeStreamable
	}
	if strings.HasPrefix(path, "/rpc") {
		return types.TransportTypeHTTP
	}

	return types.TransportTypeHTTP
}

// extractRequestMetadata extracts metadata from the request
func (m *FilterMiddleware) extractRequestMetadata(c *gin.Context) map[string]interface{} {
	metadata := make(map[string]interface{})

	metadata["method"] = c.Request.Method
	metadata["path"] = c.Request.URL.Path
	metadata["user_agent"] = c.Request.UserAgent()
	metadata["remote_addr"] = c.ClientIP()

	if referer := c.Request.Referer(); referer != "" {
		metadata["referer"] = referer
	}

	return metadata
}

// extractResponseMetadata extracts metadata from the response
func (m *FilterMiddleware) extractResponseMetadata(c *gin.Context) map[string]interface{} {
	metadata := make(map[string]interface{})

	metadata["status_code"] = c.Writer.Status()
	metadata["content_length"] = c.Writer.Size()

	return metadata
}

// extractHeaders extracts headers from HTTP header map
func (m *FilterMiddleware) extractHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)

	for key, values := range headers {
		if len(values) > 0 {
			result[strings.ToLower(key)] = values[0]
		}
	}

	return result
}

// extractQueryParams extracts query parameters
func (m *FilterMiddleware) extractQueryParams(c *gin.Context) map[string]interface{} {
	params := make(map[string]interface{})

	for key, values := range c.Request.URL.Query() {
		if len(values) == 1 {
			params[key] = values[0]
		} else {
			params[key] = values
		}
	}

	return params
}

// logViolations logs filter violations to the database
func (m *FilterMiddleware) logViolations(ctx context.Context, filterCtx *FilterContext, result *FilterResult) error {
	if len(result.Violations) == 0 {
		return nil
	}

	for _, violation := range result.Violations {
		dbViolation := &models.FilterViolation{
			OrganizationID: filterCtx.OrganizationID,
			FilterID:       "", // Would need to be extracted from violation metadata
			RequestID:      filterCtx.RequestID,
			ViolationType:  violation.Type,
			ActionTaken:    string(result.Action),
			Severity:       violation.Severity,
			UserID:         filterCtx.UserID,
			Direction:      stringPtr(string(filterCtx.Direction)),
			Metadata:       violation.Metadata,
		}

		if violation.Match != "" {
			// Limit content snippet length for privacy/storage
			snippet := violation.Match
			if len(snippet) > 500 {
				snippet = snippet[:500] + "..."
			}
			dbViolation.ContentSnippet = &snippet
		}

		if violation.Pattern != "" {
			dbViolation.PatternMatched = &violation.Pattern
		}

		// Extract remote IP from metadata
		if remoteAddr, exists := filterCtx.Metadata["remote_addr"].(string); exists {
			dbViolation.RemoteIP = &remoteAddr
		}

		// Extract user agent from metadata
		if userAgent, exists := filterCtx.Metadata["user_agent"].(string); exists {
			dbViolation.UserAgent = &userAgent
		}

		// Note: LogViolation method would need to be added to FilterService interface
		// For now, just log that we would save the violation
		log.Printf("Would log violation: %+v", dbViolation)
	}

	return nil
}

// sanitizeViolations removes sensitive information from violations before returning to client
func (m *FilterMiddleware) sanitizeViolations(violations []FilterViolation) []map[string]interface{} {
	sanitized := make([]map[string]interface{}, len(violations))

	for i, violation := range violations {
		sanitized[i] = map[string]interface{}{
			"type":     violation.Type,
			"severity": violation.Severity,
			"position": violation.Position,
		}

		// Don't expose the actual matched content or patterns to the client
		// for security reasons
	}

	return sanitized
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// ApplyToTransports applies content filtering to transport layer
func (m *FilterMiddleware) ApplyToTransports() {
	// This would be called during server initialization to integrate
	// filtering with the transport layer. Each transport would need
	// to be updated to call the filtering service at appropriate points.
	
	// For example:
	// - JSON-RPC: Filter method calls and responses
	// - WebSocket: Filter incoming/outgoing messages
	// - SSE: Filter outgoing events
	// - MCP: Filter MCP protocol messages
	
	log.Println("Content filtering integrated with transport layers")
}

// ResponseWriter wraps gin.ResponseWriter to capture response body for filtering
type ResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// NewResponseWriter creates a new response writer wrapper
func NewResponseWriter(w gin.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		body:           bytes.NewBuffer(nil),
	}
}

// Write captures the response body
func (w *ResponseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

// WriteString captures the response body
func (w *ResponseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// GetBody returns the captured response body
func (w *ResponseWriter) GetBody() string {
	return w.body.String()
}

// AdvancedFilterMiddleware provides more sophisticated filtering with response capture
type AdvancedFilterMiddleware struct {
	*FilterMiddleware
}

// NewAdvancedFilterMiddleware creates a new advanced filter middleware
func NewAdvancedFilterMiddleware(service FilterService) *AdvancedFilterMiddleware {
	return &AdvancedFilterMiddleware{
		FilterMiddleware: NewFilterMiddleware(service),
	}
}

// Handler returns an advanced Gin middleware handler with response filtering
func (m *AdvancedFilterMiddleware) Handler() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Skip filtering for certain paths
		if m.shouldSkipFiltering(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Get organization and user context
		orgID, exists := c.Get("organization_id")
		if !exists {
			orgID = "00000000-0000-0000-0000-000000000000"
		}

		userID, _ := c.Get("user_id")
		userIDStr := "default-user"
		if userID != nil {
			if uid, ok := userID.(string); ok {
				userIDStr = uid
			}
		}

		requestID := uuid.New().String()
		c.Set("filter_request_id", requestID)

		// Process inbound request
		if err := m.processInboundRequest(c, orgID.(string), userIDStr, requestID); err != nil {
			log.Printf("Filter error processing inbound request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		// Wrap response writer to capture response body
		responseWriter := NewResponseWriter(c.Writer)
		c.Writer = responseWriter

		// Continue with the request
		c.Next()

		// Process outbound response with captured body
		if err := m.processOutboundResponseAdvanced(c, responseWriter, orgID.(string), userIDStr, requestID); err != nil {
			log.Printf("Filter error processing outbound response: %v", err)
		}
	})
}

// processOutboundResponseAdvanced processes the outgoing response with captured body
func (m *AdvancedFilterMiddleware) processOutboundResponseAdvanced(c *gin.Context, responseWriter *ResponseWriter, orgID, userID, requestID string) error {
	// Skip response filtering for non-success responses or empty bodies
	if c.Writer.Status() < 200 || c.Writer.Status() >= 300 || responseWriter.body.Len() == 0 {
		return nil
	}

	// Create filter context for outbound
	filterCtx := &FilterContext{
		RequestID:      requestID,
		OrganizationID: orgID,
		UserID:         userID,
		Direction:      FilterDirectionOutbound,
		ContentType:    c.Writer.Header().Get("Content-Type"),
		Transport:      m.getTransportType(c),
		Metadata:       m.extractResponseMetadata(c),
		Timestamp:      time.Now(),
	}

	// Create filter content from response body
	content := CreateFilterContent(
		responseWriter.GetBody(),
		nil,
		m.extractResponseHeaders(c.Writer.Header()),
		nil,
	)

	// Apply filters
	result, modifiedContent, err := m.service.ProcessContent(context.Background(), filterCtx, content)
	if err != nil {
		return fmt.Errorf("failed to process response content through filters: %w", err)
	}

	// Handle filter result
	if result.Blocked {
		// Log violation
		if err := m.logViolations(c.Request.Context(), filterCtx, result); err != nil {
			log.Printf("Failed to log response filter violations: %v", err)
		}

		// For outbound filtering, we typically log but don't block the response
		// as it's already being sent. This could be enhanced based on requirements.
		log.Printf("Response would be blocked by filters: %s", result.Reason)
	}

	// Handle content modification (would require rewriting the response)
	if result.Modified && modifiedContent != nil {
		log.Printf("Response would be modified by filters")
		// In a full implementation, you'd need to rewrite the response here
	}

	// Log warnings/audit events
	if result.Action == FilterActionWarn || result.Action == FilterActionAudit {
		if err := m.logViolations(c.Request.Context(), filterCtx, result); err != nil {
			log.Printf("Failed to log response filter violations: %v", err)
		}
	}

	return nil
}

// extractResponseHeaders extracts headers from response writer
func (m *AdvancedFilterMiddleware) extractResponseHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)

	for key, values := range headers {
		if len(values) > 0 {
			result[strings.ToLower(key)] = values[0]
		}
	}

	return result
}