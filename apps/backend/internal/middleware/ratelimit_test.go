package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestIPRateLimitWithMemory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestsPerMin int
		numRequests    int
		expectedStatus int
		description    string
	}{
		{
			name:           "under_limit",
			requestsPerMin: 10,
			numRequests:    5,
			expectedStatus: http.StatusOK,
			description:    "Should allow requests under the limit",
		},
		{
			name:           "over_limit",
			requestsPerMin: 2,
			numRequests:    5,
			expectedStatus: http.StatusTooManyRequests,
			description:    "Should block requests over the limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(IPRateLimitWithMemory(tt.requestsPerMin))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			var lastStatus int
			for i := 0; i < tt.numRequests; i++ {
				req, _ := http.NewRequest("GET", "/test", nil)
				req.RemoteAddr = "127.0.0.1:12345" // Consistent IP
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				lastStatus = w.Code

				// Small delay to prevent race conditions
				time.Sleep(10 * time.Millisecond)
			}

			if tt.expectedStatus == http.StatusTooManyRequests {
				assert.Equal(t, tt.expectedStatus, lastStatus, tt.description)
			} else {
				assert.Equal(t, tt.expectedStatus, lastStatus, tt.description)
			}
		})
	}
}

func TestIPRateLimitSkipPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &IPRateLimitConfig{
		Enabled:        true,
		RequestsPerMin: 1, // Very low limit
		RedisEnabled:   false,
		SkipPaths:      []string{"/health", "/metrics"},
	}

	router := gin.New()
	router.Use(IPRateLimit(config))
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Health endpoint should never be rate limited
	for i := 0; i < 10; i++ {
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "Health endpoint should not be rate limited")
	}

	// API endpoint should be rate limited after 1 request
	req1, _ := http.NewRequest("GET", "/api/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code, "First API request should succeed")

	req2, _ := http.NewRequest("GET", "/api/test", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code, "Second API request should be rate limited")
}

func TestGetClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &IPRateLimitConfig{
		CustomHeaders: []string{"X-Real-IP", "X-Forwarded-For"},
	}

	tests := []struct {
		name        string
		headers     map[string]string
		remoteAddr  string
		expectedIP  string
		description string
	}{
		{
			name:        "x_real_ip_header",
			headers:     map[string]string{"X-Real-IP": "192.168.1.100"},
			remoteAddr:  "127.0.0.1:12345",
			expectedIP:  "192.168.1.100",
			description: "Should use X-Real-IP header when present",
		},
		{
			name:        "x_forwarded_for_header",
			headers:     map[string]string{"X-Forwarded-For": "10.0.0.5, 192.168.1.1"},
			remoteAddr:  "127.0.0.1:12345",
			expectedIP:  "10.0.0.5",
			description: "Should use first IP from X-Forwarded-For header",
		},
		{
			name:        "fallback_to_remote_addr",
			headers:     map[string]string{},
			remoteAddr:  "203.0.113.42:54321",
			expectedIP:  "203.0.113.42",
			description: "Should fallback to remote address when no headers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/test", func(c *gin.Context) {
				ip := getClientIP(c, config)
				c.JSON(http.StatusOK, gin.H{"ip": ip})
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedIP, tt.description)
		})
	}
}

func TestExtractFirstIP(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"192.168.1.1", "192.168.1.1"},
		{"10.0.0.1, 192.168.1.1", "10.0.0.1"},
		{"203.0.113.42, 198.51.100.1, 127.0.0.1", "203.0.113.42"},
		{"  192.168.1.5  , 10.0.0.1", "  192.168.1.5  "},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractFirstIP(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShouldSkipPath(t *testing.T) {
	skipPaths := []string{"/health", "/metrics", "/debug/vars"}

	tests := []struct {
		path     string
		expected bool
	}{
		{"/health", true},
		{"/metrics", true},
		{"/debug/vars", true},
		{"/api/test", false},
		{"/health/detailed", false}, // Exact match only
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := shouldSkipPath(tt.path, skipPaths)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func BenchmarkIPRateLimit(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(IPRateLimitWithMemory(1000)) // High limit for benchmarking
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}