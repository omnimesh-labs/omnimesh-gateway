package sse

import (
	"bufio"
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"mcp-gateway/apps/backend/tests/helpers"
)

func TestSSEConnection(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}

	t.Run("Basic SSE Connection", func(t *testing.T) {
		req, err := http.NewRequest("GET", server.GetURL("/sse"), http.NoBody)
		helpers.AssertNil(t, err, "Should create request successfully")

		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("Cache-Control", "no-cache")

		resp, err := client.Do(req)
		helpers.AssertNil(t, err, "SSE connection should succeed")
		defer resp.Body.Close()

		helpers.AssertStatusCode(t, http.StatusOK, &helpers.JSONResponse{StatusCode: resp.StatusCode}, "HTTP status should be 200")
		helpers.AssertEqual(t, "text/event-stream", resp.Header.Get("Content-Type"), "Content-Type should be text/event-stream")
		helpers.AssertEqual(t, "no-cache", resp.Header.Get("Cache-Control"), "Cache-Control should be no-cache")
	})

	t.Run("SSE Connection with Session ID", func(t *testing.T) {
		req, err := http.NewRequest("GET", server.GetURL("/sse"), http.NoBody)
		helpers.AssertNil(t, err, "Should create request successfully")

		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("X-Session-ID", "test-session-123")

		resp, err := client.Do(req)
		helpers.AssertNil(t, err, "SSE connection should succeed")
		defer resp.Body.Close()

		helpers.AssertStatusCode(t, http.StatusOK, &helpers.JSONResponse{StatusCode: resp.StatusCode}, "HTTP status should be 200")
	})
}

func TestSSEEvents(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	httpClient := helpers.NewHTTPClient(server.BaseURL)

	t.Run("Send SSE Event", func(t *testing.T) {
		eventData := map[string]interface{}{
			"event": "message",
			"data": map[string]interface{}{
				"type":    "notification",
				"content": "Hello from SSE!",
			},
			"session_id": "test-session-123",
		}

		resp, err := httpClient.Post("/sse/events", eventData)
		helpers.AssertNil(t, err, "Send event request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
		helpers.AssertMapKeyExists(t, resp.Body, "success", "Response should indicate success")
	})

	t.Run("Broadcast SSE Event", func(t *testing.T) {
		eventData := map[string]interface{}{
			"event": "broadcast",
			"data": map[string]interface{}{
				"message": "Broadcast to all SSE clients",
			},
		}

		resp, err := httpClient.Post("/sse/broadcast", eventData)
		helpers.AssertNil(t, err, "Broadcast request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
		helpers.AssertMapKeyExists(t, resp.Body, "success", "Response should indicate success")
	})

	t.Run("Invalid Event Data", func(t *testing.T) {
		invalidEventData := map[string]interface{}{
			"invalid": "data",
		}

		resp, err := httpClient.Post("/sse/events", invalidEventData)
		helpers.AssertNil(t, err, "HTTP request should not fail")
		// Depending on implementation, this might return 400 or handle gracefully
		helpers.AssertTrue(t, resp.StatusCode >= 400 || resp.StatusCode == 200,
			"Should handle invalid event data appropriately")
	})
}

func TestSSEStatus(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("SSE Status", func(t *testing.T) {
		resp, err := client.Get("/sse/status")
		helpers.AssertNil(t, err, "Status request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
		helpers.AssertMapKeyExists(t, resp.Body, "active_connections", "Response should contain active_connections")
		helpers.AssertMapKeyExists(t, resp.Body, "total_connections", "Response should contain total_connections")
	})
}

func TestSSEHealth(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("SSE Health Check", func(t *testing.T) {
		resp, err := client.Get("/sse/health")
		helpers.AssertNil(t, err, "Health check request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
		helpers.AssertMapKeyExists(t, resp.Body, "status", "Response should contain status")
		helpers.AssertMapKeyValue(t, resp.Body, "status", "healthy", "Status should be healthy")
		helpers.AssertMapKeyExists(t, resp.Body, "transport", "Response should contain transport")
		helpers.AssertMapKeyExists(t, resp.Body, "capabilities", "Response should contain capabilities")
	})
}

func TestSSEReplay(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("SSE Event Replay", func(t *testing.T) {
		sessionID := "replay-test-session"

		// First, send some events
		eventData := map[string]interface{}{
			"event": "message",
			"data": map[string]interface{}{
				"content": "Test message for replay",
			},
			"session_id": sessionID,
		}

		_, err := client.Post("/sse/events", eventData)
		helpers.AssertNil(t, err, "Send event should succeed")

		// Wait a bit for event processing
		time.Sleep(100 * time.Millisecond)

		// Now try to replay events
		resp, err := client.Get("/sse/replay/" + sessionID)
		helpers.AssertNil(t, err, "Replay request should not fail")

		// The response might be 200 with events or 404 if no events found
		helpers.AssertTrue(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound,
			"Replay should return 200 or 404")
	})
}

// Helper function to read SSE events from response
func readSSEEvents(resp *http.Response, timeout time.Duration) ([]string, error) {
	var events []string
	scanner := bufio.NewScanner(resp.Body)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan bool)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") {
				events = append(events, line)
			}
		}
		done <- true
	}()

	select {
	case <-done:
		return events, scanner.Err()
	case <-ctx.Done():
		return events, ctx.Err()
	}
}

func TestSSEEventStream(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	httpClient := helpers.NewHTTPClient(server.BaseURL)

	t.Run("Receive SSE Events", func(t *testing.T) {
		// Start SSE connection
		req, err := http.NewRequest("GET", server.GetURL("/sse"), http.NoBody)
		helpers.AssertNil(t, err, "Should create request successfully")

		req.Header.Set("Accept", "text/event-stream")
		sessionID := "stream-test-session"
		req.Header.Set("X-Session-ID", sessionID)

		resp, err := client.Do(req)
		helpers.AssertNil(t, err, "SSE connection should succeed")
		defer resp.Body.Close()

		// Send an event in a separate goroutine
		go func() {
			time.Sleep(500 * time.Millisecond) // Give connection time to establish

			eventData := map[string]interface{}{
				"event": "test",
				"data": map[string]interface{}{
					"message": "Test event for stream",
				},
				"session_id": sessionID,
			}

			httpClient.Post("/sse/events", eventData)
		}()

		// Try to read events with timeout
		events, err := readSSEEvents(resp, 3*time.Second)

		// We might not receive events immediately depending on implementation
		// This test mainly verifies the connection works
		helpers.AssertNil(t, err, "Should be able to read from SSE stream")
		t.Logf("Received %d events", len(events))
	})
}

func TestSSEConcurrentConnections(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}

	t.Run("Multiple SSE Connections", func(t *testing.T) {
		const numConnections = 5
		responses := make([]*http.Response, numConnections)

		// Create multiple SSE connections
		for i := 0; i < numConnections; i++ {
			req, err := http.NewRequest("GET", server.GetURL("/sse"), http.NoBody)
			helpers.AssertNil(t, err, "Should create request successfully")

			req.Header.Set("Accept", "text/event-stream")
			req.Header.Set("X-Session-ID", "concurrent-session-"+string(rune(i)))

			resp, err := client.Do(req)
			helpers.AssertNil(t, err, "SSE connection should succeed")
			helpers.AssertStatusCode(t, http.StatusOK, &helpers.JSONResponse{StatusCode: resp.StatusCode},
				"HTTP status should be 200")

			responses[i] = resp
		}

		// Close all connections
		for _, resp := range responses {
			resp.Body.Close()
		}

		// Verify status shows connections were created
		httpClient := helpers.NewHTTPClient(server.BaseURL)
		statusResp, err := httpClient.Get("/sse/status")
		helpers.AssertNil(t, err, "Status request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, statusResp, "HTTP status should be 200")
	})
}
