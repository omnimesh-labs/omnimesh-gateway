package websocket

import (
	"net/http"
	"testing"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/tests/helpers"
)

func TestWebSocketStatus(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("WebSocket Status", func(t *testing.T) {
		resp, err := client.Get("/ws/status")
		helpers.AssertNil(t, err, "Status request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
		helpers.AssertMapKeyExists(t, resp.Body, "active_connections", "Response should contain active_connections")
		helpers.AssertMapKeyExists(t, resp.Body, "total_connections", "Response should contain total_connections")
	})
}

func TestWebSocketHealth(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("WebSocket Health Check", func(t *testing.T) {
		resp, err := client.Get("/ws/health")
		helpers.AssertNil(t, err, "Health check request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
		helpers.AssertMapKeyExists(t, resp.Body, "status", "Response should contain status")
		helpers.AssertMapKeyValue(t, resp.Body, "status", "healthy", "Status should be healthy")
		helpers.AssertMapKeyExists(t, resp.Body, "transport", "Response should contain transport")
		helpers.AssertMapKeyExists(t, resp.Body, "capabilities", "Response should contain capabilities")
	})
}

func TestWebSocketMetrics(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("WebSocket Metrics", func(t *testing.T) {
		resp, err := client.Get("/ws/metrics")
		helpers.AssertNil(t, err, "Metrics request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
		helpers.AssertMapKeyExists(t, resp.Body, "connections", "Response should contain connections metrics")
		helpers.AssertMapKeyExists(t, resp.Body, "messages", "Response should contain messages metrics")
	})

	t.Run("WebSocket Metrics with Session ID", func(t *testing.T) {
		resp, err := client.Get("/ws/metrics?session_id=test-session")
		helpers.AssertNil(t, err, "Metrics request should not fail")
		// Should return 200 and empty metrics if session doesn't exist
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
		helpers.AssertMapKeyExists(t, resp.Body, "connections", "Response should contain connections metrics")
		helpers.AssertMapKeyExists(t, resp.Body, "messages", "Response should contain messages metrics")
	})
}

func TestWebSocketSendMessage(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("Send WebSocket Message without Session", func(t *testing.T) {
		messageData := map[string]interface{}{
			"type": "text",
			"data": map[string]interface{}{
				"message": "Hello WebSocket!",
			},
		}

		resp, err := client.Post("/ws/send", messageData)
		helpers.AssertNil(t, err, "HTTP request should not fail")
		// Should return 400 Bad Request since no session ID is provided
		helpers.AssertStatusCode(t, http.StatusBadRequest, resp, "Should return 400 without session ID")
	})

	t.Run("Send WebSocket Message with Session", func(t *testing.T) {
		messageData := map[string]interface{}{
			"type": "text",
			"data": map[string]interface{}{
				"message": "Hello WebSocket!",
			},
		}

		headers := map[string]string{
			"X-Session-ID": "test-ws-session-123",
		}

		resp, err := client.Post("/ws/send", messageData, headers)
		helpers.AssertNil(t, err, "HTTP request should not fail")
		// Should return 404 if session doesn't exist, 200 if it does
		helpers.AssertTrue(t, resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusOK,
			"Should return 404 for non-existent session or 200 for existing session, got %d", resp.StatusCode)
	})

	t.Run("Broadcast WebSocket Message", func(t *testing.T) {
		messageData := map[string]interface{}{
			"type": "text",
			"data": map[string]interface{}{
				"message": "Broadcast to all WebSocket clients",
			},
		}

		resp, err := client.Post("/ws/broadcast", messageData)
		helpers.AssertNil(t, err, "Broadcast request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
		helpers.AssertMapKeyExists(t, resp.Body, "success", "Response should contain success field")
		helpers.AssertMapKeyValue(t, resp.Body, "success", true, "Broadcast should succeed")
	})
}

func TestWebSocketPing(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("WebSocket Ping without Session", func(t *testing.T) {
		resp, err := client.Post("/ws/ping", nil)
		helpers.AssertNil(t, err, "HTTP request should not fail")

		// Should return 400 Bad Request since no session ID is provided
		helpers.AssertStatusCode(t, http.StatusBadRequest, resp, "Should return 400 without session ID")
	})

	t.Run("WebSocket Ping with Session", func(t *testing.T) {
		headers := map[string]string{
			"X-Session-ID": "test-ws-session-123",
		}

		resp, err := client.Post("/ws/ping", nil, headers)
		helpers.AssertNil(t, err, "HTTP request should not fail")
		// Should return 404 if session doesn't exist, 200 if it does
		helpers.AssertTrue(t, resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusOK,
			"Should return 404 for non-existent session or 200 for existing session, got %d", resp.StatusCode)
	})
}

func TestWebSocketClose(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("Close WebSocket Connection without Session", func(t *testing.T) {
		resp, err := client.Delete("/ws/close")
		helpers.AssertNil(t, err, "HTTP request should not fail")
		// Should return 400 Bad Request since no session ID is provided
		helpers.AssertStatusCode(t, http.StatusBadRequest, resp, "Should return 400 without session ID")
	})

	t.Run("Close WebSocket Connection with Session", func(t *testing.T) {
		headers := map[string]string{
			"X-Session-ID": "test-ws-session-123",
		}

		resp, err := client.Delete("/ws/close", headers)
		helpers.AssertNil(t, err, "HTTP request should not fail")
		// Should return 200 OK (idempotent - closing non-existent session is a no-op)
		helpers.AssertStatusCode(t, http.StatusOK, resp, "Should return 200 (idempotent operation)")
	})
}

func TestWebSocketErrorHandling(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("Invalid Message Data", func(t *testing.T) {
		invalidMessageData := map[string]interface{}{
			"invalid": "data",
		}

		headers := map[string]string{
			"X-Session-ID": "test-ws-session-123",
		}

		resp, err := client.Post("/ws/send", invalidMessageData, headers)
		helpers.AssertNil(t, err, "HTTP request should not fail")
		// Should return 400 Bad Request for invalid message format
		helpers.AssertStatusCode(t, http.StatusBadRequest, resp, "Should return 400 for invalid message data")
	})

	t.Run("Empty Session ID", func(t *testing.T) {
		messageData := map[string]interface{}{
			"type": "text",
			"data": map[string]interface{}{
				"message": "Test message",
			},
		}

		headers := map[string]string{
			"X-Session-ID": "",
		}

		resp, err := client.Post("/ws/send", messageData, headers)
		helpers.AssertNil(t, err, "HTTP request should not fail")
		helpers.AssertStatusCode(t, http.StatusBadRequest, resp, "Should return 400 with empty session ID")
	})
}
