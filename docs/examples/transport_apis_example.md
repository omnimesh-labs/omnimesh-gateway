# MCP Gateway - Transport APIs Testing Guide

This guide provides curl examples to test all the transport layer APIs in the MCP Gateway.

## Prerequisites

1. Start the MCP Gateway server:
```bash
make run
```

2. The server should be running on `http://localhost:8080`

## 1. JSON-RPC over HTTP Transport

### Basic JSON-RPC Request
```bash
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/list",
    "params": {},
    "id": "req-1"
  }'
```

### Batch JSON-RPC Request
```bash
curl -X POST http://localhost:8080/rpc/batch \
  -H "Content-Type: application/json" \
  -d '[
    {
      "jsonrpc": "2.0",
      "method": "tools/list",
      "params": {},
      "id": "req-1"
    },
    {
      "jsonrpc": "2.0",
      "method": "resources/list",
      "params": {},
      "id": "req-2"
    }
  ]'
```

### RPC Introspection
```bash
curl -X GET http://localhost:8080/rpc/introspection
```

### Server-Specific RPC (requires server registration)
```bash
curl -X POST http://localhost:8080/servers/server-uuid-here/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/list",
    "params": {},
    "id": "req-1"
  }'
```

## 2. Server-Sent Events (SSE) Transport

### Connect to SSE Stream
```bash
curl -N -H "Accept: text/event-stream" \
  http://localhost:8080/sse
```

### Send SSE Event
```bash
curl -X POST http://localhost:8080/sse/events \
  -H "Content-Type: application/json" \
  -d '{
    "event": "message",
    "data": {
      "type": "notification",
      "content": "Hello from SSE!"
    },
    "session_id": "sse-session-123"
  }'
```

### Broadcast SSE Event
```bash
curl -X POST http://localhost:8080/sse/broadcast \
  -H "Content-Type: application/json" \
  -d '{
    "event": "broadcast",
    "data": {
      "message": "Broadcast to all SSE clients"
    }
  }'
```

### Get SSE Status
```bash
curl -X GET http://localhost:8080/sse/status
```

### Replay SSE Events
```bash
curl -X GET http://localhost:8080/sse/replay/sse-session-123
```

### SSE Health Check
```bash
curl -X GET http://localhost:8080/sse/health
```

### Server-Specific SSE
```bash
curl -N -H "Accept: text/event-stream" \
  http://localhost:8080/servers/server-uuid-here/sse
```

## 3. WebSocket Transport

Note: WebSocket connections require a WebSocket client. Here are equivalent curl commands for the HTTP endpoints:

### WebSocket Status
```bash
curl -X GET http://localhost:8080/ws/status
```

### Send WebSocket Message
```bash
curl -X POST http://localhost:8080/ws/send \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: ws-session-123" \
  -d '{
    "type": "text",
    "data": {
      "message": "Hello WebSocket!"
    }
  }'
```

### Broadcast WebSocket Message
```bash
curl -X POST http://localhost:8080/ws/broadcast \
  -H "Content-Type: application/json" \
  -d '{
    "type": "text",
    "data": {
      "message": "Broadcast to all WebSocket clients"
    }
  }'
```

### WebSocket Ping
```bash
curl -X POST http://localhost:8080/ws/ping \
  -H "X-Session-ID: ws-session-123"
```

### Close WebSocket Connection
```bash
curl -X DELETE http://localhost:8080/ws/close \
  -H "X-Session-ID: ws-session-123"
```

### WebSocket Health Check
```bash
curl -X GET http://localhost:8080/ws/health
```

### WebSocket Metrics
```bash
curl -X GET http://localhost:8080/ws/metrics
```

### WebSocket Metrics for Specific Session
```bash
curl -X GET "http://localhost:8080/ws/metrics?session_id=ws-session-123"
```

### Connect to WebSocket (using websocat or similar tool)
```bash
# Install websocat: cargo install websocat
websocat ws://localhost:8080/ws
```

### Server-Specific WebSocket
```bash
websocat ws://localhost:8080/servers/server-uuid-here/ws
```

## 4. Streamable HTTP (MCP Protocol) Transport

### MCP GET Request (JSON Mode)
```bash
curl -X GET http://localhost:8080/mcp \
  -H "Accept: application/json"
```

### MCP GET Request (SSE Mode)
```bash
curl -N -X GET http://localhost:8080/mcp \
  -H "Accept: text/event-stream"
```

### MCP POST Request (JSON Mode)
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{
    "method": "POST",
    "path": "/tools/call",
    "headers": {
      "Content-Type": "application/json"
    },
    "body": {
      "name": "example_tool",
      "arguments": {
        "param1": "value1"
      }
    },
    "stateful": true,
    "stream_mode": "json"
  }'
```

### MCP POST Request (SSE Mode)
```bash
curl -N -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d '{
    "method": "POST",
    "path": "/tools/call",
    "headers": {
      "Content-Type": "application/json"
    },
    "body": {
      "name": "example_tool",
      "arguments": {
        "param1": "value1"
      }
    },
    "stateful": true,
    "stream_mode": "sse"
  }'
```

### MCP Capabilities
```bash
curl -X GET http://localhost:8080/mcp/capabilities
```

### MCP Status
```bash
curl -X GET http://localhost:8080/mcp/status
```

### MCP Status with Filters
```bash
curl -X GET "http://localhost:8080/mcp/status?server_id=server-123&user_id=user-456"
```

### MCP Health Check
```bash
curl -X GET http://localhost:8080/mcp/health
```

### Server-Specific MCP
```bash
curl -X GET http://localhost:8080/servers/server-uuid-here/mcp \
  -H "Accept: application/json"
```

## 5. STDIO Transport

### Execute STDIO Command
```bash
curl -X POST http://localhost:8080/stdio/execute \
  -H "Content-Type: application/json" \
  -d '{
    "command": "echo",
    "args": ["Hello from STDIO!"],
    "env": {
      "MY_VAR": "test_value"
    },
    "dir": "/tmp",
    "timeout": "30s"
  }'
```

### Start STDIO Process
```bash
curl -X GET "http://localhost:8080/stdio/process?action=start" \
  -H "Content-Type: application/json" \
  -d '{
    "command": "node",
    "args": ["-e", "setInterval(() => console.log(new Date().toISOString()), 1000)"],
    "timeout": "60s"
  }'
```

### Get STDIO Process Status
```bash
curl -X GET "http://localhost:8080/stdio/process?action=status"
```

### Get Specific STDIO Process Status
```bash
curl -X GET "http://localhost:8080/stdio/process?action=status&session_id=stdio-session-123"
```

### Stop STDIO Process
```bash
curl -X GET "http://localhost:8080/stdio/process?action=stop" \
  -H "X-Session-ID: stdio-session-123"
```

### Restart STDIO Process
```bash
curl -X GET "http://localhost:8080/stdio/process?action=restart" \
  -H "X-Session-ID: stdio-session-123"
```

### Send Message to STDIO Process
```bash
curl -X POST http://localhost:8080/stdio/send \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: stdio-session-123" \
  -d '{
    "id": "msg-1",
    "type": "request",
    "method": "tools/list",
    "params": {},
    "version": "2024-11-05"
  }'
```

### STDIO Health Check
```bash
curl -X GET http://localhost:8080/stdio/health
```

## 6. Testing with Real MCP Server

### Register a Test MCP Server
```bash
curl -X POST http://localhost:8080/api/gateway/servers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Echo Server",
    "description": "Simple echo MCP server for testing",
    "protocol": "stdio",
    "command": "node",
    "args": ["-e", "process.stdin.pipe(process.stdout)"],
    "environment": [],
    "version": "1.0.0",
    "weight": 100,
    "timeout": 300000000000,
    "max_retries": 3
  }'
```

### Create MCP Session
```bash
curl -X POST http://localhost:8080/api/gateway/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "server_id": "server-uuid-from-registration"
  }'
```

### Test STDIO with Registered Server
```bash
curl -X POST http://localhost:8080/stdio/send \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: session-uuid-from-creation" \
  -d '{
    "id": "test-1",
    "type": "request",
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "capabilities": {
        "roots": {
          "listChanged": true
        }
      },
      "clientInfo": {
        "name": "mcp-gateway",
        "version": "1.0.0"
      }
    },
    "version": "2024-11-05"
  }'
```

## 7. Health Checks for All Transports

### Check All Transport Health
```bash
curl -X GET http://localhost:8080/rpc/health
curl -X GET http://localhost:8080/sse/health
curl -X GET http://localhost:8080/ws/health
curl -X GET http://localhost:8080/mcp/health
curl -X GET http://localhost:8080/stdio/health
```

## 8. Error Testing

### Invalid JSON-RPC Request
```bash
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "invalid": "request"
  }'
```

### Missing Session ID
```bash
curl -X POST http://localhost:8080/ws/send \
  -H "Content-Type: application/json" \
  -d '{
    "type": "text",
    "data": "This should fail without session ID"
  }'
```

### Invalid STDIO Command
```bash
curl -X POST http://localhost:8080/stdio/execute \
  -H "Content-Type: application/json" \
  -d '{
    "command": ""
  }'
```

## 9. Load Testing Examples

### Concurrent JSON-RPC Requests
```bash
for i in {1..10}; do
  curl -X POST http://localhost:8080/rpc \
    -H "Content-Type: application/json" \
    -d "{
      \"jsonrpc\": \"2.0\",
      \"method\": \"ping\",
      \"params\": {},
      \"id\": \"req-$i\"
    }" &
done
wait
```

### Multiple SSE Connections
```bash
for i in {1..5}; do
  curl -N -H "Accept: text/event-stream" \
    http://localhost:8080/sse &
done
```

## Tips for Testing

1. **Use session IDs**: Many endpoints require session IDs. You can generate them or use the ones returned from session creation.

2. **Monitor logs**: Check the gateway logs to see how requests are processed.

3. **Test error cases**: Try invalid inputs to ensure proper error handling.

4. **Use proper headers**: Some endpoints require specific headers like `Accept: text/event-stream` for SSE.

5. **WebSocket tools**: For WebSocket testing, use tools like:
   - `websocat` (Rust): `cargo install websocat`
   - `wscat` (Node.js): `npm install -g wscat`
   - Browser DevTools WebSocket interface

6. **Server registration**: For server-specific endpoints, you need to register servers first through the gateway API.

## 10. Quick Smoke Tests

These are simplified smoke tests to quickly verify all transports are working:

### Basic Server Health
```bash
curl -s http://localhost:8080/health | jq .
```

### JSON-RPC Smoke Test
```bash
# Test ping method
curl -s -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{"id":"test-1","jsonrpc":"2.0","method":"ping","params":{}}' | jq .

# Test tools/list method  
curl -s -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{"id":"test-2","jsonrpc":"2.0","method":"tools/list","params":{}}' | jq .

# Test health and introspection
curl -s http://localhost:8080/rpc/health | jq .
curl -s http://localhost:8080/rpc/introspection | jq .
```

### SSE Smoke Test
```bash
# Test health and status
curl -s http://localhost:8080/sse/health | jq .
curl -s http://localhost:8080/sse/status | jq .
```

### WebSocket Smoke Test  
```bash
# Test health and status
curl -s http://localhost:8080/ws/health | jq .
curl -s http://localhost:8080/ws/status | jq .
```

### MCP Streamable HTTP Smoke Test
```bash
# Test health, capabilities, and status
curl -s http://localhost:8080/mcp/health | jq .
curl -s http://localhost:8080/mcp/capabilities | jq .  
curl -s http://localhost:8080/mcp/status | jq .

# Test GET request (JSON mode)
curl -s http://localhost:8080/mcp | jq .
```

### STDIO Smoke Test
```bash
# Test health
curl -s http://localhost:8080/stdio/health | jq .
```

### Server-Specific Endpoint Smoke Test
```bash
# Test server-specific JSON-RPC (uses path rewriting middleware)
curl -s -X POST http://localhost:8080/servers/test-server/rpc \
  -H "Content-Type: application/json" \
  -d '{"id":"test-proxy","jsonrpc":"2.0","method":"ping","params":{}}' | jq .
```

### Virtual MCP Server Smoke Test
```bash
# Test virtual MCP JSON-RPC endpoint
curl -s -X POST http://localhost:8080/mcp/rpc \
  -H "Content-Type: application/json" \
  -d '{"id":"virtual-test","jsonrpc":"2.0","method":"tools/list","params":{}}' | jq .

# List virtual servers
curl -s http://localhost:8080/api/admin/virtual-servers | jq .
```

### All Transport Health Check
```bash
echo "=== Transport Health Status ==="
echo "JSON-RPC:" && curl -s http://localhost:8080/rpc/health | jq '.status'
echo "SSE:" && curl -s http://localhost:8080/sse/health | jq '.status'  
echo "WebSocket:" && curl -s http://localhost:8080/ws/health | jq '.status'
echo "MCP:" && curl -s http://localhost:8080/mcp/health | jq '.status'
echo "STDIO:" && curl -s http://localhost:8080/stdio/health | jq '.status'
```

## Monitoring and Debugging

### Check Gateway Health
```bash
curl -X GET http://localhost:8080/health
```

### Monitor Active Sessions
```bash
curl -X GET http://localhost:8080/api/gateway/sessions
```

### View Server List
```bash
curl -X GET http://localhost:8080/api/gateway/servers
```
