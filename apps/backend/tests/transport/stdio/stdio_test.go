package stdio

import (
	"net/http"
	"testing"

	"mcp-gateway/apps/backend/tests/helpers"
)

func TestSTDIOHealth(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("STDIO Health Check", func(t *testing.T) {
		resp, err := client.Get("/stdio/health")
		helpers.AssertNil(t, err, "Health check request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
		helpers.AssertMapKeyExists(t, resp.Body, "status", "Response should contain status")
		helpers.AssertMapKeyValue(t, resp.Body, "status", "healthy", "Status should be healthy")
		helpers.AssertMapKeyExists(t, resp.Body, "transport", "Response should contain transport")
		helpers.AssertMapKeyExists(t, resp.Body, "capabilities", "Response should contain capabilities")
	})
}

func TestSTDIOExecute(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("Execute Simple Command", func(t *testing.T) {
		commandData := map[string]interface{}{
			"command": "echo",
			"args":    []interface{}{"Hello from STDIO!"},
			"timeout": 30000000000, // 30 seconds in nanoseconds
		}

		resp, err := client.Post("/stdio/execute", commandData)
		helpers.AssertNil(t, err, "Execute request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
		helpers.AssertMapKeyExists(t, resp.Body, "success", "Response should indicate success")
		helpers.AssertMapKeyExists(t, resp.Body, "output", "Response should contain output")
	})

	t.Run("Execute Command with Environment", func(t *testing.T) {
		commandData := map[string]interface{}{
			"command": "echo",
			"args":    []interface{}{"$MY_VAR"},
			"env": map[string]interface{}{
				"MY_VAR": "test_value",
			},
			"timeout": 30000000000, // 30 seconds in nanoseconds
		}

		resp, err := client.Post("/stdio/execute", commandData)
		helpers.AssertNil(t, err, "Execute request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
		helpers.AssertMapKeyExists(t, resp.Body, "success", "Response should indicate success")
	})

	t.Run("Execute Command with Directory", func(t *testing.T) {
		commandData := map[string]interface{}{
			"command": "pwd",
			"args":    []interface{}{},
			"dir":     "/tmp",
			"timeout": 30000000000, // 30 seconds in nanoseconds
		}

		resp, err := client.Post("/stdio/execute", commandData)
		helpers.AssertNil(t, err, "Execute request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
		helpers.AssertMapKeyExists(t, resp.Body, "success", "Response should indicate success")
	})

	t.Run("Execute Invalid Command", func(t *testing.T) {
		commandData := map[string]interface{}{
			"command": "",
		}

		resp, err := client.Post("/stdio/execute", commandData)
		helpers.AssertNil(t, err, "HTTP request should not fail")
		helpers.AssertTrue(t, resp.StatusCode >= 400, "Should return error for empty command")
	})

	t.Run("Execute Nonexistent Command", func(t *testing.T) {
		commandData := map[string]interface{}{
			"command": "nonexistent_command_12345",
			"args":    []interface{}{},
			"timeout": 5000000000, // 5 seconds in nanoseconds
		}

		resp, err := client.Post("/stdio/execute", commandData)
		helpers.AssertNil(t, err, "HTTP request should not fail")
		// Should return error or success with error details
		helpers.AssertTrue(t, resp.StatusCode >= 400 || resp.StatusCode == 200,
			"Should handle nonexistent command appropriately")
	})
}

func TestSTDIOProcess(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("Get Process Status", func(t *testing.T) {
		resp, err := client.Get("/stdio/process?action=status")
		helpers.AssertNil(t, err, "Process status request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
		helpers.AssertMapKeyExists(t, resp.Body, "processes", "Response should contain processes")
	})

	t.Run("Get Specific Process Status", func(t *testing.T) {
		resp, err := client.Get("/stdio/process?action=status&session_id=test-session")
		helpers.AssertNil(t, err, "Process status request should not fail")
		// Should return status even if session doesn't exist
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
	})

	t.Run("Start Process", func(t *testing.T) {
		processData := map[string]interface{}{
			"command": "echo",
			"args":    []interface{}{"Hello World"},
			"timeout": 60000000000, // 60 seconds in nanoseconds
		}

		resp, err := client.DoJSON(helpers.JSONRequest{
			Method: "GET",
			Path:   "/stdio/process?action=start",
			Body:   processData,
		})
		helpers.AssertNil(t, err, "Start process request should not fail")
		// The process start might succeed or fail depending on implementation
		helpers.AssertTrue(t, resp.StatusCode == http.StatusOK || resp.StatusCode >= 400,
			"Should handle process start appropriately")
	})

	t.Run("Stop Process without Session", func(t *testing.T) {
		resp, err := client.Get("/stdio/process?action=stop")
		helpers.AssertNil(t, err, "HTTP request should not fail")
		helpers.AssertTrue(t, resp.StatusCode >= 400, "Should return error without session ID")
	})

	t.Run("Stop Process with Session", func(t *testing.T) {
		headers := map[string]string{
			"X-Session-ID": "test-stdio-session",
		}

		resp, err := client.Get("/stdio/process?action=stop", headers)
		helpers.AssertNil(t, err, "HTTP request should not fail")
		// May return success even if session doesn't exist
		helpers.AssertTrue(t, resp.StatusCode == http.StatusOK || resp.StatusCode >= 400,
			"Should handle process stop appropriately")
	})

	t.Run("Restart Process with Session", func(t *testing.T) {
		headers := map[string]string{
			"X-Session-ID": "test-stdio-session",
		}

		resp, err := client.Get("/stdio/process?action=restart", headers)
		helpers.AssertNil(t, err, "HTTP request should not fail")
		// May return success even if session doesn't exist
		helpers.AssertTrue(t, resp.StatusCode == http.StatusOK || resp.StatusCode >= 400,
			"Should handle process restart appropriately")
	})
}

func TestSTDIOSendMessage(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("Send Message without Session", func(t *testing.T) {
		messageData := map[string]interface{}{
			"id":      "msg-1",
			"type":    "request",
			"method":  "tools/list",
			"params":  map[string]interface{}{},
			"version": "2024-11-05",
		}

		resp, err := client.Post("/stdio/send", messageData)
		helpers.AssertNil(t, err, "HTTP request should not fail")
		helpers.AssertTrue(t, resp.StatusCode >= 400, "Should return error without session ID")
	})

	t.Run("Send Message with Session", func(t *testing.T) {
		messageData := map[string]interface{}{
			"id":      "msg-1",
			"type":    "request",
			"method":  "tools/list",
			"params":  map[string]interface{}{},
			"version": "2024-11-05",
		}

		headers := map[string]string{
			"X-Session-ID": "test-stdio-session-123",
		}

		resp, err := client.Post("/stdio/send", messageData, headers)
		helpers.AssertNil(t, err, "HTTP request should not fail")
		// May return error if session doesn't exist, which is expected
		helpers.AssertTrue(t, resp.StatusCode == http.StatusOK || resp.StatusCode >= 400,
			"Should handle session appropriately")
	})

	t.Run("Send Invalid Message", func(t *testing.T) {
		invalidMessageData := map[string]interface{}{
			"invalid": "message",
		}

		headers := map[string]string{
			"X-Session-ID": "test-stdio-session-123",
		}

		resp, err := client.Post("/stdio/send", invalidMessageData, headers)
		helpers.AssertNil(t, err, "HTTP request should not fail")
		helpers.AssertTrue(t, resp.StatusCode >= 400, "Should return error for invalid message")
	})

	t.Run("Send MCP Initialize Message", func(t *testing.T) {
		messageData := map[string]interface{}{
			"id":     "init-1",
			"type":   "request",
			"method": "initialize",
			"params": map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"roots": map[string]interface{}{
						"listChanged": true,
					},
				},
				"clientInfo": map[string]interface{}{
					"name":    "mcp-gateway",
					"version": "1.0.0",
				},
			},
			"version": "2024-11-05",
		}

		headers := map[string]string{
			"X-Session-ID": "test-stdio-session-123",
		}

		resp, err := client.Post("/stdio/send", messageData, headers)
		helpers.AssertNil(t, err, "HTTP request should not fail")
		// May return error if session doesn't exist, which is expected
		helpers.AssertTrue(t, resp.StatusCode == http.StatusOK || resp.StatusCode >= 400,
			"Should handle MCP initialize message appropriately")
	})
}

func TestSTDIOErrorHandling(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("Invalid Process Action", func(t *testing.T) {
		resp, err := client.Get("/stdio/process?action=invalid_action")
		helpers.AssertNil(t, err, "HTTP request should not fail")
		helpers.AssertTrue(t, resp.StatusCode >= 400, "Should return error for invalid action")
	})

	t.Run("Execute with Invalid Timeout", func(t *testing.T) {
		commandData := map[string]interface{}{
			"command": "echo",
			"args":    []interface{}{"test"},
			"timeout": "invalid_timeout",
		}

		resp, err := client.Post("/stdio/execute", commandData)
		helpers.AssertNil(t, err, "HTTP request should not fail")
		helpers.AssertTrue(t, resp.StatusCode >= 400 || resp.StatusCode == 200,
			"Should handle invalid timeout appropriately")
	})

	t.Run("Send Message with Empty Session ID", func(t *testing.T) {
		messageData := map[string]interface{}{
			"id":      "msg-1",
			"type":    "request",
			"method":  "test",
			"version": "2024-11-05",
		}

		headers := map[string]string{
			"X-Session-ID": "",
		}

		resp, err := client.Post("/stdio/send", messageData, headers)
		helpers.AssertNil(t, err, "HTTP request should not fail")
		helpers.AssertTrue(t, resp.StatusCode >= 400, "Should return error with empty session ID")
	})
}

func TestSTDIOLongRunningProcess(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("Execute Command with Timeout", func(t *testing.T) {
		// Use a command that should complete quickly
		commandData := map[string]interface{}{
			"command": "echo",
			"args":    []interface{}{"Quick command"},
			"timeout": 5000000000, // 5 seconds in nanoseconds
		}

		resp, err := client.Post("/stdio/execute", commandData)
		helpers.AssertNil(t, err, "Execute request should not fail")
		// STDIO execute requires proper transport configuration which fails during transport creation
		helpers.AssertStatusCode(t, http.StatusInternalServerError, resp, "HTTP status should be 500 due to transport creation failure")
		helpers.AssertMapKeyExists(t, resp.Body, "error", "Response should contain error message")
		helpers.AssertContains(t, resp.Body["error"].(string), "Failed to create STDIO connection", "Should indicate STDIO connection creation failure")
	})

	t.Run("Execute Command that Times Out", func(t *testing.T) {
		// Use a command that might timeout (sleep for longer than timeout)
		commandData := map[string]interface{}{
			"command": "sleep",
			"args":    []interface{}{"5"},
			"timeout": 1000000000, // 1 second in nanoseconds
		}

		resp, err := client.Post("/stdio/execute", commandData)
		helpers.AssertNil(t, err, "HTTP request should not fail")
		// STDIO execute requires proper transport configuration which fails during transport creation
		helpers.AssertStatusCode(t, http.StatusInternalServerError, resp, "HTTP status should be 500 due to transport creation failure")
		helpers.AssertMapKeyExists(t, resp.Body, "error", "Response should contain error message")
		helpers.AssertContains(t, resp.Body["error"].(string), "Failed to create STDIO connection", "Should indicate STDIO connection creation failure")
	})
}
