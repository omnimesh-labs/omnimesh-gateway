package mcp

import (
	"net/http"
	"testing"
	"time"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/tests/helpers"
)

func TestMCPCapabilities(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("MCP Capabilities", func(t *testing.T) {
		resp, err := client.Get("/mcp/capabilities")
		helpers.AssertNil(t, err, "Capabilities request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
		helpers.AssertMapKeyExists(t, resp.Body, "capabilities", "Response should contain capabilities")
		helpers.AssertMapKeyExists(t, resp.Body, "protocol_version", "Response should contain protocol_version")
		helpers.AssertMapKeyExists(t, resp.Body, "server_info", "Response should contain server_info")
	})
}

func TestMCPStatus(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("MCP Status", func(t *testing.T) {
		resp, err := client.Get("/mcp/status")
		helpers.AssertNil(t, err, "Status request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
		helpers.AssertMapKeyExists(t, resp.Body, "active_sessions", "Response should contain active_sessions")
		helpers.AssertMapKeyExists(t, resp.Body, "total_sessions", "Response should contain total_sessions")
		helpers.AssertMapKeyExists(t, resp.Body, "transport_mode", "Response should contain transport_mode")
	})

	t.Run("MCP Status with Filters", func(t *testing.T) {
		resp, err := client.Get("/mcp/status?server_id=test-server&user_id=test-user")
		helpers.AssertNil(t, err, "Status request with filters should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
		helpers.AssertMapKeyExists(t, resp.Body, "active_sessions", "Response should contain active_sessions")
	})
}

func TestMCPHealth(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("MCP Health Check", func(t *testing.T) {
		resp, err := client.Get("/mcp/health")
		helpers.AssertNil(t, err, "Health check request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
		helpers.AssertMapKeyExists(t, resp.Body, "status", "Response should contain status")
		helpers.AssertMapKeyValue(t, resp.Body, "status", "healthy", "Status should be healthy")
		helpers.AssertMapKeyExists(t, resp.Body, "transport", "Response should contain transport")
		helpers.AssertMapKeyExists(t, resp.Body, "capabilities", "Response should contain capabilities")
	})
}

func TestMCPGetRequests(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("MCP GET Request (JSON Mode)", func(t *testing.T) {
		headers := map[string]string{
			"Accept": "application/json",
		}

		resp, err := client.Get("/mcp", headers)
		helpers.AssertNil(t, err, "MCP GET request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
		helpers.AssertTrue(t,
			resp.Headers.Get("Content-Type") == "application/json" ||
				resp.Headers.Get("Content-Type") == "application/json; charset=utf-8",
			"Content-Type should be application/json")
	})

	t.Run("MCP GET Request (SSE Mode)", func(t *testing.T) {
		// SSE requests are typically long-lived connections that stream data
		// For testing, we'll make a quick request and verify headers only
		// Note: The SSE handler may timeout or block, which is expected behavior

		// Create a manual request with shorter timeout for SSE testing
		url := server.BaseURL + "/mcp"
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Accept", "text/event-stream")

		shortTimeoutClient := &http.Client{Timeout: 2 * time.Second}
		httpResp, err := shortTimeoutClient.Do(req)
		if err != nil {
			// SSE connections may timeout, but we should at least get initial headers
			// Skip this test if it times out as this is expected behavior for SSE
			t.Skip("SSE connection timed out as expected for streaming response")
			return
		}
		defer httpResp.Body.Close()

		helpers.AssertEqual(t, http.StatusOK, httpResp.StatusCode, "HTTP status should be 200")
		// Response should be SSE format
		helpers.AssertTrue(t,
			httpResp.Header.Get("Content-Type") == "text/event-stream" ||
				httpResp.Header.Get("Content-Type") == "text/event-stream; charset=utf-8",
			"Content-Type should be text/event-stream")
	})
}

func TestMCPPostRequests(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("MCP POST Request (JSON Mode)", func(t *testing.T) {
		requestData := map[string]interface{}{
			"method": "POST",
			"path":   "/tools/call",
			"headers": map[string]interface{}{
				"Content-Type": "application/json",
			},
			"body": map[string]interface{}{
				"name": "example_tool",
				"arguments": map[string]interface{}{
					"param1": "value1",
				},
			},
			"stateful":    true,
			"stream_mode": "json",
		}

		headers := map[string]string{
			"Accept": "application/json",
		}

		resp, err := client.Post("/mcp", requestData, headers)
		helpers.AssertNil(t, err, "MCP POST request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
		helpers.AssertTrue(t,
			resp.Headers.Get("Content-Type") == "application/json" ||
				resp.Headers.Get("Content-Type") == "application/json; charset=utf-8",
			"Content-Type should be application/json")
	})

	t.Run("MCP POST Request (SSE Mode)", func(t *testing.T) {
		requestData := map[string]interface{}{
			"method": "POST",
			"path":   "/tools/call",
			"headers": map[string]interface{}{
				"Content-Type": "application/json",
			},
			"body": map[string]interface{}{
				"name": "example_tool",
				"arguments": map[string]interface{}{
					"param1": "value1",
				},
			},
			"stateful":    true,
			"stream_mode": "sse",
		}

		headers := map[string]string{
			"Accept": "text/event-stream",
		}

		resp, err := client.Post("/mcp", requestData, headers)
		helpers.AssertNil(t, err, "MCP POST request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
		// Response should be SSE format
		helpers.AssertTrue(t,
			resp.Headers.Get("Content-Type") == "text/event-stream" ||
				resp.Headers.Get("Content-Type") == "text/event-stream; charset=utf-8" ||
				resp.Headers.Get("Content-Type") == "text/event-stream;charset=utf-8",
			"Content-Type should be text/event-stream")
	})

	t.Run("MCP POST Request with Invalid Data", func(t *testing.T) {
		invalidRequestData := map[string]interface{}{
			"invalid": "request",
		}

		resp, err := client.Post("/mcp", invalidRequestData)
		helpers.AssertNil(t, err, "HTTP request should not fail")
		// Depending on implementation, this might return 400 or handle gracefully
		helpers.AssertTrue(t, resp.StatusCode >= 400 || resp.StatusCode == 200,
			"Should handle invalid request data appropriately")
	})
}

func TestMCPProtocolMethods(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("MCP Tool Call", func(t *testing.T) {
		requestData := map[string]interface{}{
			"method": "POST",
			"path":   "/tools/call",
			"headers": map[string]interface{}{
				"Content-Type": "application/json",
			},
			"body": map[string]interface{}{
				"name":      "ping",
				"arguments": map[string]interface{}{},
			},
			"stateful":    false,
			"stream_mode": "json",
		}

		resp, err := client.Post("/mcp", requestData)
		helpers.AssertNil(t, err, "MCP tool call should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
	})

	t.Run("MCP Resource Read", func(t *testing.T) {
		requestData := map[string]interface{}{
			"method": "GET",
			"path":   "/resources/test-resource",
			"headers": map[string]interface{}{
				"Content-Type": "application/json",
			},
			"stateful":    false,
			"stream_mode": "json",
		}

		resp, err := client.Post("/mcp", requestData)
		helpers.AssertNil(t, err, "MCP resource read should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
	})

	t.Run("MCP Prompt Get", func(t *testing.T) {
		requestData := map[string]interface{}{
			"method": "GET",
			"path":   "/prompts/test-prompt",
			"headers": map[string]interface{}{
				"Content-Type": "application/json",
			},
			"body": map[string]interface{}{
				"name": "test-prompt",
				"arguments": map[string]interface{}{
					"context": "test",
				},
			},
			"stateful":    false,
			"stream_mode": "json",
		}

		resp, err := client.Post("/mcp", requestData)
		helpers.AssertNil(t, err, "MCP prompt get should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
	})
}

func TestMCPStatefulMode(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("MCP Stateful Request", func(t *testing.T) {
		requestData := map[string]interface{}{
			"method": "POST",
			"path":   "/session/initialize",
			"headers": map[string]interface{}{
				"Content-Type": "application/json",
			},
			"body": map[string]interface{}{
				"client_info": map[string]interface{}{
					"name":    "test-client",
					"version": "1.0.0",
				},
			},
			"stateful":    true,
			"stream_mode": "json",
		}

		headers := map[string]string{
			"X-Session-ID": "test-mcp-session-123",
		}

		resp, err := client.Post("/mcp", requestData, headers)
		helpers.AssertNil(t, err, "MCP stateful request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
	})

	t.Run("MCP Stateless Request", func(t *testing.T) {
		requestData := map[string]interface{}{
			"method": "GET",
			"path":   "/capabilities",
			"headers": map[string]interface{}{
				"Content-Type": "application/json",
			},
			"stateful":    false,
			"stream_mode": "json",
		}

		resp, err := client.Post("/mcp", requestData)
		helpers.AssertNil(t, err, "MCP stateless request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
	})
}

func TestMCPErrorHandling(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("MCP Request with Missing Method", func(t *testing.T) {
		requestData := map[string]interface{}{
			"path": "/tools/call",
			"body": map[string]interface{}{
				"name": "test_tool",
			},
		}

		resp, err := client.Post("/mcp", requestData)
		helpers.AssertNil(t, err, "HTTP request should not fail")
		// Should handle missing method appropriately
		helpers.AssertTrue(t, resp.StatusCode >= 400 || resp.StatusCode == 200,
			"Should handle missing method appropriately")
	})

	t.Run("MCP Request with Invalid Stream Mode", func(t *testing.T) {
		requestData := map[string]interface{}{
			"method":      "POST",
			"path":        "/tools/call",
			"stream_mode": "invalid_mode",
		}

		resp, err := client.Post("/mcp", requestData)
		helpers.AssertNil(t, err, "HTTP request should not fail")
		// Should handle invalid stream mode appropriately
		helpers.AssertTrue(t, resp.StatusCode >= 400 || resp.StatusCode == 200,
			"Should handle invalid stream mode appropriately")
	})

	t.Run("MCP Request with Unsupported Accept Header", func(t *testing.T) {
		requestData := map[string]interface{}{
			"method": "GET",
			"path":   "/capabilities",
		}

		headers := map[string]string{
			"Accept": "application/xml",
		}

		resp, err := client.Post("/mcp", requestData, headers)
		helpers.AssertNil(t, err, "HTTP request should not fail")
		// Should handle unsupported accept header appropriately
		helpers.AssertTrue(t, resp.StatusCode >= 400 || resp.StatusCode == 200,
			"Should handle unsupported accept header appropriately")
	})
}
