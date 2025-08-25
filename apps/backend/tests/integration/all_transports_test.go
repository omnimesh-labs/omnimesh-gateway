package integration

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"mcp-gateway/apps/backend/tests/helpers"
)

func TestAllTransportsHealthCheck(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	transports := []string{"rpc", "sse", "ws", "mcp", "stdio"}

	t.Run("All Transport Health Checks", func(t *testing.T) {
		for _, transport := range transports {
			t.Run(transport+" Health", func(t *testing.T) {
				resp, err := client.Get("/" + transport + "/health")
				helpers.AssertNil(t, err, "Health check for %s should not fail", transport)
				helpers.AssertStatusCode(t, http.StatusOK, resp, "Health check for %s should return 200", transport)
				helpers.AssertMapKeyExists(t, resp.Body, "status", "Response should contain status for %s", transport)
				helpers.AssertMapKeyValue(t, resp.Body, "status", "healthy", "Status should be healthy for %s", transport)
			})
		}
	})
}

func TestConcurrentTransportRequests(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("Concurrent Mixed Transport Requests", func(t *testing.T) {
		const numRequests = 30
		var wg sync.WaitGroup
		results := make([]error, numRequests)

		requests := []struct {
			fn   func() error
			name string
		}{
			{
				name: "RPC Ping",
				fn: func() error {
					_, _, err := client.DoJSONRPC("/rpc", "ping", map[string]interface{}{}, "")
					return err
				},
			},
			{
				name: "SSE Status",
				fn: func() error {
					_, err := client.Get("/sse/status")
					return err
				},
			},
			{
				name: "WebSocket Status",
				fn: func() error {
					_, err := client.Get("/ws/status")
					return err
				},
			},
			{
				name: "MCP Capabilities",
				fn: func() error {
					_, err := client.Get("/mcp/capabilities")
					return err
				},
			},
			{
				name: "STDIO Health",
				fn: func() error {
					_, err := client.Get("/stdio/health")
					return err
				},
			},
		}

		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				req := requests[index%len(requests)]
				results[index] = req.fn()
			}(i)
		}

		wg.Wait()

		// Check results
		successCount := 0
		for i, err := range results {
			if err == nil {
				successCount++
			} else {
				t.Errorf("Request %d failed: %v", i, err)
			}
		}

		helpers.AssertEqual(t, numRequests, successCount, "All concurrent mixed transport requests should succeed")
	})
}

func TestTransportInteroperability(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("Cross-Transport Session Management", func(t *testing.T) {
		sessionID := "test-integration-session-123"

		// Try to use the same session ID across different transports
		headers := map[string]string{
			"X-Session-ID": sessionID,
		}

		// Test SSE with session
		sseResp, err := client.Get("/sse/status", headers)
		helpers.AssertNil(t, err, "SSE status with session should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, sseResp, "SSE status should return 200")

		// Test WebSocket with session
		wsResp, err := client.Get("/ws/status", headers)
		helpers.AssertNil(t, err, "WebSocket status with session should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, wsResp, "WebSocket status should return 200")

		// Test MCP with session
		mcpResp, err := client.Get("/mcp/status", headers)
		helpers.AssertNil(t, err, "MCP status with session should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, mcpResp, "MCP status should return 200")
	})
}

func TestTransportLoadTesting(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("High Load JSON-RPC Requests", func(t *testing.T) {
		const numRequests = 100
		const concurrency = 10

		start := time.Now()

		var wg sync.WaitGroup
		semaphore := make(chan struct{}, concurrency)
		results := make([]error, numRequests)

		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				_, _, err := client.DoJSONRPC("/rpc", "ping", map[string]interface{}{}, "")
				results[index] = err
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)

		// Check results
		successCount := 0
		for _, err := range results {
			if err == nil {
				successCount++
			}
		}

		t.Logf("Completed %d requests in %v (%.2f req/sec)",
			numRequests, duration, float64(numRequests)/duration.Seconds())

		// Assert that most requests succeeded (high load tests should still be reliable)
		minSuccessRate := 0.95 // 95% success rate minimum
		actualSuccessRate := float64(successCount) / float64(numRequests)

		helpers.AssertTrue(t, actualSuccessRate >= minSuccessRate,
			"Success rate should be at least %.0f%%, got %.0f%%",
			minSuccessRate*100, actualSuccessRate*100)
	})
}

func TestTransportResponseTimes(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	transports := []struct {
		fn   func() error
		name string
	}{
		{
			name: "JSON-RPC",
			fn: func() error {
				_, _, err := client.DoJSONRPC("/rpc", "ping", map[string]interface{}{}, "")
				return err
			},
		},
		{
			name: "SSE",
			fn: func() error {
				_, err := client.Get("/sse/status")
				return err
			},
		},
		{
			name: "WebSocket",
			fn: func() error {
				_, err := client.Get("/ws/status")
				return err
			},
		},
		{
			name: "MCP",
			fn: func() error {
				_, err := client.Get("/mcp/capabilities")
				return err
			},
		},
		{
			name: "STDIO",
			fn: func() error {
				_, err := client.Get("/stdio/health")
				return err
			},
		},
	}

	t.Run("Transport Response Time Benchmarks", func(t *testing.T) {
		const numSamples = 20

		for _, transport := range transports {
			t.Run(transport.name+" Response Time", func(t *testing.T) {
				durations := make([]time.Duration, numSamples)
				successCount := 0

				for i := 0; i < numSamples; i++ {
					start := time.Now()
					err := transport.fn()
					duration := time.Since(start)

					durations[i] = duration
					if err == nil {
						successCount++
					}
				}

				// Calculate average response time
				var total time.Duration
				for _, d := range durations {
					total += d
				}
				average := total / time.Duration(numSamples)

				t.Logf("%s - Average response time: %v (success rate: %d/%d)",
					transport.name, average, successCount, numSamples)

				// Assert reasonable response time (under 500ms for CI/CD compatibility)
				// This is a sanity check, not a performance benchmark
				maxResponseTime := 500 * time.Millisecond
				helpers.AssertTrue(t, average < maxResponseTime,
					"%s average response time should be under %v for basic operations, got %v",
					transport.name, maxResponseTime, average)

				// Assert reasonable success rate
				successRate := float64(successCount) / float64(numSamples)
				helpers.AssertTrue(t, successRate >= 0.95,
					"%s success rate should be at least 95%%, got %.0f%%",
					transport.name, successRate*100)
			})
		}
	})
}

func TestSystemHealthOverview(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("Overall System Health", func(t *testing.T) {
		resp, err := client.Get("/health")
		helpers.AssertNil(t, err, "System health check should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "System health should return 200")
		helpers.AssertMapKeyExists(t, resp.Body, "status", "Response should contain overall status")
	})
}
