package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityConfig holds security headers configuration
type SecurityConfig struct {
	CustomHeaders             map[string]string
	ContentSecurityPolicy     string
	XFrameOptions             string
	XContentTypeOptions       string
	ReferrerPolicy            string
	PermissionsPolicy         string
	StrictTransportSecurity   string
	XXSSProtection            string
	CrossOriginEmbedderPolicy string
	CrossOriginOpenerPolicy   string
	CrossOriginResourcePolicy string
}

// DefaultSecurityConfig returns default security headers configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		ContentSecurityPolicy:     "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' ws: wss:; frame-ancestors 'none';",
		XFrameOptions:             "DENY",
		XContentTypeOptions:       "nosniff",
		ReferrerPolicy:            "strict-origin-when-cross-origin",
		PermissionsPolicy:         "geolocation=(), microphone=(), camera=()",
		StrictTransportSecurity:   "max-age=31536000; includeSubDomains; preload",
		XXSSProtection:            "1; mode=block",
		CrossOriginEmbedderPolicy: "require-corp",
		CrossOriginOpenerPolicy:   "same-origin",
		CrossOriginResourcePolicy: "cross-origin",
		CustomHeaders:             make(map[string]string),
	}
}

// DevelopmentSecurityConfig returns security headers configuration suitable for development
func DevelopmentSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		ContentSecurityPolicy: "default-src 'self' 'unsafe-inline' 'unsafe-eval'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https: http:; font-src 'self' data:; connect-src 'self' ws: wss: http: https:; frame-ancestors 'self';",
		XFrameOptions:         "SAMEORIGIN",
		XContentTypeOptions:   "nosniff",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		PermissionsPolicy:     "geolocation=(), microphone=(), camera=()",
		// Don't enforce HSTS in development
		StrictTransportSecurity: "",
		XXSSProtection:          "1; mode=block",
		// Relaxed CORS policies for development
		CrossOriginEmbedderPolicy: "",
		CrossOriginOpenerPolicy:   "",
		CrossOriginResourcePolicy: "cross-origin",
		CustomHeaders:             make(map[string]string),
	}
}

// SecurityHeaders returns a security headers middleware with default configuration
func SecurityHeaders() gin.HandlerFunc {
	return SecurityHeadersWithConfig(DefaultSecurityConfig())
}

// SecurityHeadersWithConfig returns a security headers middleware with custom configuration
func SecurityHeadersWithConfig(config *SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set Content Security Policy
		if config.ContentSecurityPolicy != "" {
			c.Header("Content-Security-Policy", config.ContentSecurityPolicy)
		}

		// Set X-Frame-Options
		if config.XFrameOptions != "" {
			c.Header("X-Frame-Options", config.XFrameOptions)
		}

		// Set X-Content-Type-Options
		if config.XContentTypeOptions != "" {
			c.Header("X-Content-Type-Options", config.XContentTypeOptions)
		}

		// Set Referrer-Policy
		if config.ReferrerPolicy != "" {
			c.Header("Referrer-Policy", config.ReferrerPolicy)
		}

		// Set Permissions-Policy
		if config.PermissionsPolicy != "" {
			c.Header("Permissions-Policy", config.PermissionsPolicy)
		}

		// Set Strict-Transport-Security (only for HTTPS)
		if config.StrictTransportSecurity != "" && c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", config.StrictTransportSecurity)
		}

		// Set X-XSS-Protection
		if config.XXSSProtection != "" {
			c.Header("X-XSS-Protection", config.XXSSProtection)
		}

		// Set Cross-Origin-Embedder-Policy
		if config.CrossOriginEmbedderPolicy != "" {
			c.Header("Cross-Origin-Embedder-Policy", config.CrossOriginEmbedderPolicy)
		}

		// Set Cross-Origin-Opener-Policy
		if config.CrossOriginOpenerPolicy != "" {
			c.Header("Cross-Origin-Opener-Policy", config.CrossOriginOpenerPolicy)
		}

		// Set Cross-Origin-Resource-Policy
		if config.CrossOriginResourcePolicy != "" {
			c.Header("Cross-Origin-Resource-Policy", config.CrossOriginResourcePolicy)
		}

		// Set custom headers
		for key, value := range config.CustomHeaders {
			if value != "" {
				c.Header(key, value)
			}
		}

		c.Next()
	}
}

// SecureHeaders is an alias for SecurityHeaders for backward compatibility
func SecureHeaders() gin.HandlerFunc {
	return SecurityHeaders()
}

// SecureHeadersWithConfig is an alias for SecurityHeadersWithConfig for backward compatibility
func SecureHeadersWithConfig(config *SecurityConfig) gin.HandlerFunc {
	return SecurityHeadersWithConfig(config)
}
