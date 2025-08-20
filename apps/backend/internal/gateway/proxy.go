package gateway

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"mcp-gateway/apps/backend/internal/types"

	"github.com/gin-gonic/gin"
)

// Proxy handles HTTP proxy functionality
type Proxy struct {
	router    *Router
	transport *http.Transport
	config    *ProxyConfig
}

// ProxyConfig holds proxy configuration
type ProxyConfig struct {
	Timeout    time.Duration
	MaxRetries int
}

// NewProxy creates a new HTTP proxy
func NewProxy(router *Router, config *ProxyConfig) *Proxy {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	return &Proxy{
		router:    router,
		transport: transport,
		config:    config,
	}
}

// ProxyRequest handles proxying requests to MCP servers
func (p *Proxy) ProxyRequest(c *gin.Context) {
	// Get organization ID from context
	orgID, exists := c.Get("organization_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization not found"})
		return
	}

	// Select target server
	server, err := p.router.SelectServer(orgID.(string), c.Request)
	if err != nil {
		p.handleProxyError(c, err)
		return
	}

	// Create proxy request
	proxyReq := &types.ProxyRequest{
		UserID:         p.getUserID(c),
		OrganizationID: orgID.(string),
		ServerID:       server.ID,
		Method:         c.Request.Method,
		Path:           c.Request.URL.Path,
		Headers:        p.extractHeaders(c.Request),
		RemoteIP:       c.ClientIP(),
		UserAgent:      c.Request.UserAgent(),
		StartTime:      time.Now(),
	}

	// Proxy the request
	response := p.proxyToServer(c, server, proxyReq)

	// Handle response
	p.handleProxyResponse(c, response)
}

// proxyToServer performs the actual proxy operation
func (p *Proxy) proxyToServer(c *gin.Context, server *types.MCPServer, proxyReq *types.ProxyRequest) *types.ProxyResponse {
	startTime := time.Now()

	response := &types.ProxyResponse{
		RequestID: proxyReq.ID,
		EndTime:   time.Now(),
	}

	// Parse target URL
	targetURL, err := url.Parse(server.URL)
	if err != nil {
		response.Error = "Invalid server URL"
		response.StatusCode = http.StatusInternalServerError
		return response
	}

	// Create reverse proxy
	reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
	reverseProxy.Transport = p.transport

	// Modify request
	reverseProxy.Director = func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.Host = targetURL.Host

		// Add headers
		req.Header.Set("X-Forwarded-For", proxyReq.RemoteIP)
		req.Header.Set("X-Forwarded-Proto", req.URL.Scheme)
		req.Header.Set("X-Request-ID", proxyReq.ID)
	}

	// Create response writer to capture response
	writer := &proxyResponseWriter{
		ResponseWriter: c.Writer,
		response:       response,
	}

	// Add timeout context
	ctx, cancel := context.WithTimeout(c.Request.Context(), p.config.Timeout)
	defer cancel()
	c.Request = c.Request.WithContext(ctx)

	// Perform proxy
	reverseProxy.ServeHTTP(writer, c.Request)

	response.Latency = time.Since(startTime)
	return response
}

// proxyResponseWriter captures proxy response data
type proxyResponseWriter struct {
	gin.ResponseWriter
	response *types.ProxyResponse
	body     bytes.Buffer
}

func (w *proxyResponseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *proxyResponseWriter) WriteHeader(statusCode int) {
	w.response.StatusCode = statusCode
	w.response.Headers = make(map[string]string)

	for key, values := range w.Header() {
		if len(values) > 0 {
			w.response.Headers[key] = values[0]
		}
	}

	w.ResponseWriter.WriteHeader(statusCode)
}

// handleProxyResponse handles the proxy response
func (p *Proxy) handleProxyResponse(c *gin.Context, response *types.ProxyResponse) {
	// TODO: Log the response
	// TODO: Update server statistics
	// TODO: Handle circuit breaker logic

	if response.Error != "" {
		c.JSON(response.StatusCode, gin.H{"error": response.Error})
		return
	}

	// Response is already written by the reverse proxy
}

// handleProxyError handles proxy errors
func (p *Proxy) handleProxyError(c *gin.Context, err error) {
	var statusCode int
	var message string

	switch e := err.(type) {
	case *types.Error:
		statusCode = e.Status
		message = e.Message
	default:
		statusCode = http.StatusInternalServerError
		message = "Internal server error"
	}

	errorResp := &types.ErrorResponse{
		Error:   types.NewError("PROXY_ERROR", message, statusCode),
		Success: false,
	}

	c.JSON(statusCode, errorResp)
}

// getUserID extracts user ID from context
func (p *Proxy) getUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return userID.(string)
	}
	return ""
}

// extractHeaders extracts request headers
func (p *Proxy) extractHeaders(req *http.Request) map[string]string {
	headers := make(map[string]string)

	for key, values := range req.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return headers
}

// readRequestBody reads and returns request body
func (p *Proxy) readRequestBody(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return nil, nil
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	// Restore body for subsequent reads
	req.Body = io.NopCloser(bytes.NewBuffer(body))

	return body, nil
}
