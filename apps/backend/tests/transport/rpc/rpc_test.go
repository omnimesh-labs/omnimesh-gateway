package rpc

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"mcp-gateway/apps/backend/tests/helpers"
)

func TestJSONRPCBasicRequests(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("Ping Request", func(t *testing.T) {
		rpcResp, httpResp, err := client.DoJSONRPC("/rpc", "ping", map[string]interface{}{}, "test-ping")

		helpers.AssertNil(t, err, "JSON-RPC request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, httpResp, "HTTP status should be 200")
		helpers.AssertJSONRPCSuccess(t, rpcResp, "JSON-RPC response should be successful")
		helpers.AssertMapKeyExists(t, rpcResp.Result, "message", "Response should contain message")
		helpers.AssertMapKeyValue(t, rpcResp.Result, "message", "pong", "Message should be 'pong'")
	})

	t.Run("List Tools Request", func(t *testing.T) {
		rpcResp, httpResp, err := client.DoJSONRPC("/rpc", "tools/list", map[string]interface{}{}, "test-list-tools")

		helpers.AssertNil(t, err, "JSON-RPC request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, httpResp, "HTTP status should be 200")
		helpers.AssertJSONRPCSuccess(t, rpcResp, "JSON-RPC response should be successful")
		helpers.AssertMapKeyExists(t, rpcResp.Result, "tools", "Response should contain tools")
	})

	t.Run("Call Tool Request", func(t *testing.T) {
		params := map[string]interface{}{
			"name":      "ping",
			"arguments": map[string]interface{}{},
		}
		rpcResp, httpResp, err := client.DoJSONRPC("/rpc", "tools/call", params, "test-call-tool")

		helpers.AssertNil(t, err, "JSON-RPC request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, httpResp, "HTTP status should be 200")
		helpers.AssertJSONRPCSuccess(t, rpcResp, "JSON-RPC response should be successful")
		helpers.AssertMapKeyExists(t, rpcResp.Result, "result", "Response should contain result")
	})

	t.Run("Unknown Method", func(t *testing.T) {
		rpcResp, httpResp, err := client.DoJSONRPC("/rpc", "unknown/method", map[string]interface{}{}, "test-unknown")

		helpers.AssertNil(t, err, "JSON-RPC request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, httpResp, "HTTP status should be 200")
		helpers.AssertJSONRPCSuccess(t, rpcResp, "JSON-RPC response should be successful for unknown methods")
		helpers.AssertMapKeyExists(t, rpcResp.Result, "method", "Response should contain method")
		helpers.AssertMapKeyValue(t, rpcResp.Result, "method", "unknown/method", "Method should match request")
	})
}

func TestJSONRPCErrorHandling(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("Invalid JSON", func(t *testing.T) {
		resp, err := client.DoJSON(helpers.JSONRequest{
			Method: "POST",
			Path:   "/rpc",
			Body:   "invalid json",
		})

		helpers.AssertNil(t, err, "HTTP request should not fail")
		helpers.AssertStatusCode(t, http.StatusBadRequest, resp, "Should return 400 for invalid JSON")
	})

	t.Run("Missing JSONRPC Version", func(t *testing.T) {
		invalidReq := map[string]interface{}{
			"method": "ping",
			"id":     "test",
		}
		resp, err := client.Post("/rpc", invalidReq)

		helpers.AssertNil(t, err, "HTTP request should not fail")
		helpers.AssertStatusCode(t, http.StatusBadRequest, resp, "Should return 400 for missing jsonrpc version")
	})

	t.Run("Missing Method", func(t *testing.T) {
		invalidReq := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      "test",
		}
		resp, err := client.Post("/rpc", invalidReq)

		helpers.AssertNil(t, err, "HTTP request should not fail")
		helpers.AssertStatusCode(t, http.StatusBadRequest, resp, "Should return 400 for missing method")
	})

	t.Run("Invalid Tool Call", func(t *testing.T) {
		params := map[string]interface{}{
			"name": "nonexistent_tool",
		}
		_, httpResp, err := client.DoJSONRPC("/rpc", "tools/call", params, "test-invalid-tool")

		helpers.AssertNil(t, err, "JSON-RPC request should not fail")
		helpers.AssertStatusCode(t, http.StatusInternalServerError, httpResp, "Should return 500 for invalid tool")
	})
}

func TestJSONRPCConcurrency(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("Concurrent Ping Requests", func(t *testing.T) {
		const numRequests = 20
		var wg sync.WaitGroup
		results := make([]error, numRequests)

		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				rpcResp, httpResp, err := client.DoJSONRPC("/rpc", "ping", map[string]interface{}{},
					"concurrent-test-"+string(rune(index)))

				if err != nil {
					results[index] = err
					return
				}

				if httpResp.StatusCode != http.StatusOK {
					results[index] = err
					return
				}

				if rpcResp.Error != nil {
					results[index] = err
					return
				}
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

		helpers.AssertEqual(t, numRequests, successCount, "All concurrent requests should succeed")
	})

	t.Run("Mixed Concurrent Requests", func(t *testing.T) {
		const numRequests = 15
		var wg sync.WaitGroup
		results := make([]error, numRequests)

		methods := []string{"ping", "tools/list", "resources/list", "prompts/list"}

		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				method := methods[index%len(methods)]
				rpcResp, httpResp, err := client.DoJSONRPC("/rpc", method, map[string]interface{}{},
					"mixed-test-"+string(rune(index)))

				if err != nil {
					results[index] = err
					return
				}

				if httpResp.StatusCode != http.StatusOK {
					results[index] = err
					return
				}

				if rpcResp.Error != nil {
					results[index] = err
					return
				}
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

		helpers.AssertEqual(t, numRequests, successCount, "All mixed concurrent requests should succeed")
	})
}

func TestJSONRPCIntrospection(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("RPC Introspection", func(t *testing.T) {
		resp, err := client.Get("/rpc/introspection")

		helpers.AssertNil(t, err, "Introspection request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
		helpers.AssertMapKeyExists(t, resp.Body, "methods", "Response should contain methods")
		helpers.AssertMapKeyExists(t, resp.Body, "description", "Response should contain description")
		helpers.AssertMapKeyExists(t, resp.Body, "version", "Response should contain version")
		helpers.AssertMapKeyExists(t, resp.Body, "transport", "Response should contain transport")
	})
}

func TestJSONRPCHealth(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("RPC Health Check", func(t *testing.T) {
		resp, err := client.Get("/rpc/health")

		helpers.AssertNil(t, err, "Health check request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, resp, "HTTP status should be 200")
		helpers.AssertMapKeyExists(t, resp.Body, "status", "Response should contain status")
		helpers.AssertMapKeyValue(t, resp.Body, "status", "healthy", "Status should be healthy")
		helpers.AssertMapKeyExists(t, resp.Body, "transport", "Response should contain transport")
		helpers.AssertMapKeyExists(t, resp.Body, "capabilities", "Response should contain capabilities")
	})
}

func TestJSONRPCResponseTime(t *testing.T) {
	server, err := helpers.NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	client := helpers.NewHTTPClient(server.BaseURL)

	t.Run("Response Time Performance", func(t *testing.T) {
		const numRequests = 10
		durations := make([]time.Duration, numRequests)

		for i := 0; i < numRequests; i++ {
			start := time.Now()

			rpcResp, httpResp, err := client.DoJSONRPC("/rpc", "ping", map[string]interface{}{},
				"perf-test-"+string(rune(i)))

			duration := time.Since(start)
			durations[i] = duration

			helpers.AssertNil(t, err, "Request should not fail")
			helpers.AssertStatusCode(t, http.StatusOK, httpResp, "HTTP status should be 200")
			helpers.AssertJSONRPCSuccess(t, rpcResp, "JSON-RPC response should be successful")
		}

		// Calculate average response time
		var total time.Duration
		for _, d := range durations {
			total += d
		}
		average := total / time.Duration(numRequests)

		t.Logf("Average response time: %v", average)

		// Assert that average response time is reasonable (under 100ms)
		helpers.AssertTrue(t, average < 100*time.Millisecond,
			"Average response time should be under 100ms, got %v", average)
	})
}
