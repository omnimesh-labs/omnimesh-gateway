#!/bin/bash

# Test runner script for MCP Gateway transport tests

set -e

# Change to the directory where this script is located
cd "$(dirname "$0")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
TIMEOUT=${TEST_TIMEOUT:-30s}
VERBOSE=${TEST_VERBOSE:-false}
COVERAGE=${TEST_COVERAGE:-false}

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Parse command line arguments
TEST_TYPE=${1:-all}

case $TEST_TYPE in
    unit)
        log_info "Running unit tests..."
        TEST_PATHS="./unit/..."
        ;;
    transport)
        log_info "Running transport tests..."
        TEST_PATHS="./transport/..."
        ;;
    integration)
        log_info "Running integration tests..."
        TEST_PATHS="./integration/..."
        ;;
    rpc)
        log_info "Running JSON-RPC transport tests..."
        TEST_PATHS="./transport/rpc/..."
        ;;
    sse)
        log_info "Running SSE transport tests..."
        TEST_PATHS="./transport/sse/..."
        ;;
    websocket|ws)
        log_info "Running WebSocket transport tests..."
        TEST_PATHS="./transport/websocket/..."
        ;;
    mcp)
        log_info "Running MCP transport tests..."
        TEST_PATHS="./transport/mcp/..."
        ;;
    stdio)
        log_info "Running STDIO transport tests..."
        TEST_PATHS="./transport/stdio/..."
        ;;
    all)
        log_info "Running all tests..."
        TEST_PATHS="./..."
        ;;
    *)
        log_error "Unknown test type: $TEST_TYPE"
        echo "Usage: $0 [unit|transport|integration|rpc|sse|websocket|mcp|stdio|all]"
        exit 1
        ;;
esac

# Build test flags
TEST_FLAGS="-timeout=$TIMEOUT"

if [ "$VERBOSE" = "true" ]; then
    TEST_FLAGS="$TEST_FLAGS -v"
fi

if [ "$COVERAGE" = "true" ]; then
    TEST_FLAGS="$TEST_FLAGS -cover -coverprofile=coverage.out"
fi

# Check if we need to start the server for integration tests
START_SERVER=false
if [[ "$TEST_TYPE" =~ ^(transport|integration|rpc|sse|websocket|ws|mcp|stdio|all)$ ]]; then
    START_SERVER=true
fi

# Function to start test server
start_test_server() {
    log_info "Starting test server..."
    
    # Check if server is already running
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        log_warning "Server is already running on port 8080"
        return 0
    fi
    
    # Start server in background
    go run ../../cmd/api/main.go > /dev/null 2>&1 &
    SERVER_PID=$!
    
    # Wait for server to be ready
    log_info "Waiting for server to start..."
    for i in {1..30}; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            log_success "Test server started (PID: $SERVER_PID)"
            return 0
        fi
        sleep 1
    done
    
    log_error "Failed to start test server"
    return 1
}

# Function to stop test server
stop_test_server() {
    if [ ! -z "$SERVER_PID" ]; then
        log_info "Stopping test server (PID: $SERVER_PID)..."
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
        log_success "Test server stopped"
    fi
}

# Trap to ensure server is stopped on exit
trap stop_test_server EXIT

# Start server if needed
if [ "$START_SERVER" = "true" ]; then
    start_test_server || exit 1
    
    # Give server a moment to fully initialize
    sleep 2
fi

# Run tests
log_info "Executing tests: go test $TEST_FLAGS $TEST_PATHS"
echo "----------------------------------------"

if go test $TEST_FLAGS $TEST_PATHS; then
    echo "----------------------------------------"
    log_success "All tests passed!"
    
    # Show coverage report if enabled
    if [ "$COVERAGE" = "true" ] && [ -f "coverage.out" ]; then
        echo ""
        log_info "Coverage report:"
        go tool cover -func=coverage.out | tail -1
        
        # Generate HTML coverage report
        go tool cover -html=coverage.out -o coverage.html
        log_info "HTML coverage report generated: coverage.html"
    fi
else
    echo "----------------------------------------"
    log_error "Some tests failed!"
    exit 1
fi
