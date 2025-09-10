package rpc

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/tests/helpers"
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
		// Unknown methods should return an error in JSON-RPC format
		helpers.AssertNotNil(t, rpcResp.Error, "JSON-RPC response should contain error for unknown method")
		if rpcResp.Error != nil {
			helpers.AssertEqual(t, -32601, rpcResp.Error.Code, "Error code should be -32601 (Method not found)")
		}
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
		rpcResp, httpResp, err := client.DoJSONRPC("/rpc", "tools/call", params, "test-invalid-tool")

		helpers.AssertNil(t, err, "JSON-RPC request should not fail")
		helpers.AssertStatusCode(t, http.StatusOK, httpResp, "HTTP status should be 200 (errors are in JSON-RPC response)")
		// Tool not found should return a JSON-RPC error
		helpers.AssertNotNil(t, rpcResp.Error, "Should return JSON-RPC error for invalid tool")
		if rpcResp.Error != nil {
			// -32602 for invalid params or -32000 for server error
			helpers.AssertTrue(t, rpcResp.Error.Code == -32602 || rpcResp.Error.Code == -32000,
				"Error code should be -32602 (Invalid params) or -32000 (Server error), got %d", rpcResp.Error.Code)
		}
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
					fmt.Sprintf("concurrent-test-%d", index))

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
					fmt.Sprintf("mixed-test-%d", index))

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
				fmt.Sprintf("perf-test-%d", i))

			duration := time.Since(start)
			durations[i] = duration

			helpers.AssertNil(t, err, "Request should not fail")
			helpers.AssertStatusCode(t, http.StatusOK, httpResp, "HTTP status should be 200")
			helpers.AssertJSONRPCSuccess(t, rpcResp, "JSON-RPC response should be successful")
		}

		var total time.Duration
		for _, d := range durations {
			total += d
		}
		average := total / time.Duration(numRequests)

		t.Logf("Average response time: %v", average)

		// Assert that average response time is reasonable (under 500ms for CI/CD compatibility)
		// This is a sanity check, not a performance benchmark
		maxResponseTime := 500 * time.Millisecond
		helpers.AssertTrue(t, average < maxResponseTime,
			"Average response time should be under %v for basic operations, got %v", maxResponseTime, average)
	})
}
