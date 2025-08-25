package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		config          *SecurityConfig
		expectedHeaders map[string]string
		name            string
	}{
		{
			name:   "Default security headers",
			config: DefaultSecurityConfig(),
			expectedHeaders: map[string]string{
				"Content-Security-Policy":      "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' ws: wss:; frame-ancestors 'none';",
				"X-Frame-Options":              "DENY",
				"X-Content-Type-Options":       "nosniff",
				"Referrer-Policy":              "strict-origin-when-cross-origin",
				"Permissions-Policy":           "geolocation=(), microphone=(), camera=()",
				"X-XSS-Protection":             "1; mode=block",
				"Cross-Origin-Embedder-Policy": "require-corp",
				"Cross-Origin-Opener-Policy":   "same-origin",
				"Cross-Origin-Resource-Policy": "cross-origin",
			},
		},
		{
			name:   "Development security headers",
			config: DevelopmentSecurityConfig(),
			expectedHeaders: map[string]string{
				"Content-Security-Policy":      "default-src 'self' 'unsafe-inline' 'unsafe-eval'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https: http:; font-src 'self' data:; connect-src 'self' ws: wss: http: https:; frame-ancestors 'self';",
				"X-Frame-Options":              "SAMEORIGIN",
				"X-Content-Type-Options":       "nosniff",
				"Referrer-Policy":              "strict-origin-when-cross-origin",
				"Permissions-Policy":           "geolocation=(), microphone=(), camera=()",
				"X-XSS-Protection":             "1; mode=block",
				"Cross-Origin-Resource-Policy": "cross-origin",
			},
		},
		{
			name: "Custom security headers",
			config: &SecurityConfig{
				ContentSecurityPolicy: "default-src 'none'",
				XFrameOptions:         "SAMEORIGIN",
				CustomHeaders: map[string]string{
					"X-Custom-Header": "custom-value",
				},
			},
			expectedHeaders: map[string]string{
				"Content-Security-Policy": "default-src 'none'",
				"X-Frame-Options":         "SAMEORIGIN",
				"X-Custom-Header":         "custom-value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new Gin router
			r := gin.New()
			r.Use(SecurityHeadersWithConfig(tt.config))
			r.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "test"})
			})

			// Create a test request
			req, _ := http.NewRequest("GET", "/test", http.NoBody)
			w := httptest.NewRecorder()

			// Perform the request
			r.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, http.StatusOK, w.Code)

			// Check expected headers
			for header, expectedValue := range tt.expectedHeaders {
				actualValue := w.Header().Get(header)
				assert.Equal(t, expectedValue, actualValue, "Header %s should match", header)
			}
		})
	}
}

func TestSecurityHeadersHTTPS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a new Gin router with HSTS enabled
	r := gin.New()
	config := DefaultSecurityConfig()
	r.Use(SecurityHeadersWithConfig(config))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	t.Run("HSTS header not set for HTTP", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", http.NoBody)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		// HSTS should not be set for HTTP requests
		hsts := w.Header().Get("Strict-Transport-Security")
		assert.Empty(t, hsts, "HSTS header should not be set for HTTP requests")
	})
}

func TestSecurityHeadersWithEmptyConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a config with empty values
	config := &SecurityConfig{
		ContentSecurityPolicy: "",
		XFrameOptions:         "",
		CustomHeaders:         map[string]string{},
	}

	r := gin.New()
	r.Use(SecurityHeadersWithConfig(config))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Check that empty headers are not set
	assert.Empty(t, w.Header().Get("Content-Security-Policy"))
	assert.Empty(t, w.Header().Get("X-Frame-Options"))
}

func TestDefaultSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(SecurityHeaders()) // Use default config
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Check that default headers are set
	assert.NotEmpty(t, w.Header().Get("Content-Security-Policy"))
	assert.NotEmpty(t, w.Header().Get("X-Frame-Options"))
	assert.NotEmpty(t, w.Header().Get("X-Content-Type-Options"))
}
