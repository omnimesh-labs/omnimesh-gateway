# Omnimesh AI Gateway Test Suite

Comprehensive test suite for the Omnimesh AI Gateway transport layer APIs.

## Test Structure

```
tests/
├── helpers/              # Test helper utilities
│   ├── server.go         # Test server management
│   ├── http.go          # HTTP client helpers
│   └── assertions.go    # Test assertion helpers
├── transport/           # Transport-specific tests
│   ├── rpc/            # JSON-RPC transport tests
│   ├── sse/            # Server-Sent Events transport tests
│   ├── websocket/      # WebSocket transport tests
│   ├── mcp/            # MCP protocol transport tests
│   └── stdio/          # STDIO transport tests
├── integration/         # Integration tests
│   └── all_transports_test.go
├── unit/               # Unit tests (placeholder)
├── run_tests.sh       # Test runner script
└── README.md          # This file
```

## Running Tests

### Using Makefile Commands

```bash
# Run all tests
make test

# Run specific transport tests
make test-rpc         # JSON-RPC transport tests
make test-sse         # SSE transport tests
make test-websocket   # WebSocket transport tests
make test-mcp         # MCP transport tests
make test-stdio       # STDIO transport tests

# Run test categories
make test-transport   # All transport tests
make test-integration # Integration tests
make test-unit        # Unit tests

# Run with coverage
make test-coverage

# Run with verbose output
make test-verbose

# Run all transport tests sequentially
make test-all-transports
```

### Using Test Runner Script

```bash
# Run all tests
./tests/run_tests.sh all

# Run specific test categories
./tests/run_tests.sh transport
./tests/run_tests.sh integration
./tests/run_tests.sh unit

# Run specific transport tests
./tests/run_tests.sh rpc
./tests/run_tests.sh sse
./tests/run_tests.sh websocket  # or 'ws'
./tests/run_tests.sh mcp
./tests/run_tests.sh stdio
```

### Environment Variables

```bash
# Enable verbose output
TEST_VERBOSE=true ./tests/run_tests.sh all

# Enable coverage reporting
TEST_COVERAGE=true ./tests/run_tests.sh all

# Set test timeout (default: 30s)
TEST_TIMEOUT=60s ./tests/run_tests.sh all
```

## Test Coverage

### JSON-RPC Transport Tests
- ✅ Basic JSON-RPC requests (ping, tools/list, tools/call)
- ✅ Error handling (invalid JSON, missing fields)
- ✅ Concurrent request handling
- ✅ Introspection endpoint
- ✅ Health checks
- ✅ Performance testing

### SSE Transport Tests
- ✅ SSE connection establishment
- ✅ Event sending and broadcasting
- ✅ Status monitoring
- ✅ Health checks
- ✅ Event replay functionality
- ✅ Concurrent connections

### WebSocket Transport Tests
- ✅ HTTP endpoint testing (status, metrics)
- ✅ Message sending and broadcasting
- ✅ Ping/pong functionality
- ✅ Connection management
- ✅ Health checks
- ✅ Error handling

### MCP Transport Tests
- ✅ MCP protocol capabilities
- ✅ JSON and SSE modes
- ✅ Status and health monitoring
- ✅ Protocol method testing
- ✅ Stateful/stateless modes
- ✅ Error handling

### STDIO Transport Tests
- ✅ Command execution
- ✅ Process management (start, stop, restart)
- ✅ Message sending to processes
- ✅ Health checks
- ✅ Error handling
- ✅ Timeout handling

### Integration Tests
- ✅ Cross-transport health checks
- ✅ Concurrent mixed transport requests
- ✅ Transport interoperability
- ✅ Load testing
- ✅ Response time benchmarks
- ✅ System health overview

## Test Features

### Automatic Test Server Management
- Tests automatically start and stop a test server instance
- Server runs on a random available port to avoid conflicts
- Waits for server readiness before running tests
- Graceful shutdown after tests complete

### Comprehensive Error Testing
- Invalid request formats
- Missing required fields
- Network timeouts
- Server errors
- Session management errors

### Performance Testing
- Concurrent request handling
- Response time measurements
- Load testing with high request volumes
- Success rate monitoring under load

### Cross-Transport Testing
- Session management across transports
- Mixed concurrent requests
- Transport interoperability verification

## Helper Utilities

### Test Server (`helpers/server.go`)
- Creates isolated test server instances
- Handles port allocation and server lifecycle
- Provides health check waiting functionality

### HTTP Client (`helpers/http.go`)
- Simplified HTTP request handling
- JSON-RPC request helpers
- Response parsing and error handling

### Assertions (`helpers/assertions.go`)
- Common test assertions
- HTTP status code validation
- JSON-RPC response validation
- Map key/value assertions

## Adding New Tests

1. **Create test file** in appropriate directory (`tests/transport/<transport>/`)
2. **Use helper functions** from `tests/helpers/`
3. **Follow naming convention**: `Test<Feature><Scenario>`
4. **Include error cases** and edge conditions
5. **Add performance tests** for critical paths
6. **Update Makefile** if adding new test categories

### Example Test Structure

```go
func TestNewFeature(t *testing.T) {
    server, err := helpers.NewTestServer()
    if err != nil {
        t.Fatalf("Failed to create test server: %v", err)
    }
    defer server.Stop()

    if err := server.Start(); err != nil {
        t.Fatalf("Failed to start test server: %v", err)
    }

    client := helpers.NewHTTPClient(server.BaseURL)

    t.Run("Basic Functionality", func(t *testing.T) {
        // Test implementation
    })

    t.Run("Error Handling", func(t *testing.T) {
        // Error case testing
    })
}
```

## Continuous Integration

The test suite is designed to run in CI environments:

- Tests are self-contained and don't require external dependencies
- Automatic server management eliminates setup complexity
- Configurable timeouts and retry logic for flaky environments
- Coverage reporting for code quality metrics

## Troubleshooting

### Common Issues

1. **Port conflicts**: Tests use random ports, but if issues persist, check for hanging processes
2. **Server startup timeouts**: Increase `TEST_TIMEOUT` environment variable
3. **Test failures under load**: Some tests may be sensitive to system performance
4. **Coverage reports**: Ensure `go tool cover` is available in your PATH

### Debug Mode

```bash
# Run with verbose output to see detailed test execution
TEST_VERBOSE=true ./tests/run_tests.sh all

# Run specific failing tests
./tests/run_tests.sh rpc -v
```
