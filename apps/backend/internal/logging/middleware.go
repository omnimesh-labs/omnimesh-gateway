package logging

import (
	"bytes"
	"io"
	"time"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Middleware provides request logging middleware
type Middleware struct {
	service *Service
}

// NewMiddleware creates a new logging middleware
func NewMiddleware(service *Service) *Middleware {
	return &Middleware{
		service: service,
	}
}

// RequestLogger logs HTTP requests and responses
func (m *Middleware) RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip logging for health checks and metrics endpoints
		if m.shouldSkipLogging(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Generate request ID
		requestID := uuid.New().String()
		c.Set("request_id", requestID)

		// Capture request details
		startTime := time.Now()

		// Capture request body if needed
		var requestBody []byte
		if c.Request.Body != nil && m.shouldLogBody(c.Request.Method) {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Create response writer to capture response
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer([]byte{}),
		}
		c.Writer = writer

		// Capture user context before processing (to avoid race conditions)
		var userID, orgID string
		if uid, exists := c.Get("user_id"); exists {
			if u, ok := uid.(string); ok {
				userID = u
			}
		}
		if oid, exists := c.Get("organization_id"); exists {
			if o, ok := oid.(string); ok {
				orgID = o
			}
		}

		// Process request
		c.Next()

		// Calculate duration and capture errors after processing
		duration := time.Since(startTime)
		var errorStr string
		if len(c.Errors) > 0 {
			errorStr = c.Errors.String()
		}

		// Create log entry
		entry := &types.LogEntry{
			ID:             requestID,
			Timestamp:      startTime,
			Level:          m.getLogLevel(writer.status),
			Type:           types.LogTypeRequest,
			RequestID:      requestID,
			Method:         c.Request.Method,
			Path:           c.Request.URL.Path,
			StatusCode:     writer.status,
			Duration:       duration,
			RemoteIP:       c.ClientIP(),
			UserAgent:      c.Request.UserAgent(),
			Message:        "HTTP Request",
			UserID:         userID,
			OrganizationID: orgID,
		}

		// Add additional data
		entry.Data = map[string]interface{}{
			"query_params":  c.Request.URL.RawQuery,
			"response_size": writer.size,
		}

		if len(requestBody) > 0 && len(requestBody) < 1024 {
			entry.Data["request_body"] = string(requestBody)
		}

		if writer.body.Len() > 0 && writer.body.Len() < 1024 {
			entry.Data["response_body"] = writer.body.String()
		}

		if errorStr != "" {
			entry.Error = errorStr
		}

		logEntry := &LogEntry{
			ID:         entry.ID,
			Timestamp:  entry.Timestamp,
			Level:      LogLevel(entry.Level),
			Message:    entry.Message,
			Logger:     "request",
			RequestID:  entry.RequestID,
			UserID:     entry.UserID,
			OrgID:      entry.OrganizationID,
			StatusCode: entry.StatusCode,
			Data:       entry.Data,
		}

		if err := m.service.Log(c.Request.Context(), logEntry); err != nil {
			// TODO: Handle logging error (maybe fallback to file logging)
		}
	}
}

// AuditLogger logs administrative actions
func (m *Middleware) AuditLogger(action, resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Get user context
		userID, _ := c.Get("user_id")
		orgID, _ := c.Get("organization_id")

		// Get resource ID from URL parameter
		resourceID := c.Param("id")
		if resourceID == "" {
			resourceID = c.Param("server_id")
		}

		// Handle nil values for single-tenant setup
		userIDStr := "system"
		if userID != nil {
			if uid, ok := userID.(string); ok {
				userIDStr = uid
			}
		}

		orgIDStr := "00000000-0000-0000-0000-000000000000"
		if orgID != nil {
			if oid, ok := orgID.(string); ok {
				orgIDStr = oid
			}
		}

		// Create audit entry
		audit := &types.AuditLog{
			Timestamp:      time.Now(),
			UserID:         userIDStr,
			OrganizationID: orgIDStr,
			Action:         action,
			Resource:       resource,
			ResourceID:     resourceID,
			RemoteIP:       c.ClientIP(),
			UserAgent:      c.Request.UserAgent(),
			Success:        c.Writer.Status() < 400,
		}

		if c.Writer.Status() >= 400 {
			audit.Error = "Request failed"
		}

		// Add request details
		audit.Details = map[string]interface{}{
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"status_code": c.Writer.Status(),
		}

		// Log the audit event
		if err := m.service.LogAudit(c.Request.Context(), audit); err != nil {
			// TODO: Handle audit logging error
		}
	}
}

// MetricsCollector collects performance metrics
func (m *Middleware) MetricsCollector() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		duration := time.Since(startTime)

		// Collect metrics
		orgID, _ := c.Get("organization_id")

		// Request duration metric
		durationMetric := &types.Metric{
			Timestamp: startTime,
			Name:      "http_request_duration",
			Type:      types.MetricTypeHistogram,
			Value:     float64(duration.Nanoseconds()) / 1e6, // Convert to milliseconds
			Tags: map[string]string{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
				"status": string(rune(c.Writer.Status())),
			},
		}

		if orgID != nil {
			durationMetric.OrganizationID = orgID.(string)
		}

		// Request count metric
		countMetric := &types.Metric{
			Timestamp: startTime,
			Name:      "http_requests_total",
			Type:      types.MetricTypeCounter,
			Value:     1,
			Tags: map[string]string{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
				"status": string(rune(c.Writer.Status())),
			},
		}

		if orgID != nil {
			countMetric.OrganizationID = orgID.(string)
		}

		// Log metrics
		m.service.LogMetric(c.Request.Context(), durationMetric)
		m.service.LogMetric(c.Request.Context(), countMetric)
	}
}

// responseWriter captures response data
type responseWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
	size   int
}

func (w *responseWriter) Write(data []byte) (int, error) {
	size, err := w.ResponseWriter.Write(data)
	w.size += size
	if w.body != nil {
		w.body.Write(data)
	}
	return size, err
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// shouldSkipLogging determines if a path should be skipped for logging
func (m *Middleware) shouldSkipLogging(path string) bool {
	skipPaths := []string{
		"/health",
		"/metrics",
		"/favicon.ico",
	}

	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}

	return false
}

// shouldLogBody determines if request body should be logged
func (m *Middleware) shouldLogBody(method string) bool {
	return method == "POST" || method == "PUT" || method == "PATCH"
}

// getLogLevel determines log level based on status code
func (m *Middleware) getLogLevel(statusCode int) string {
	switch {
	case statusCode >= 500:
		return types.LogLevelError
	case statusCode >= 400:
		return types.LogLevelWarn
	default:
		return types.LogLevelInfo
	}
}
