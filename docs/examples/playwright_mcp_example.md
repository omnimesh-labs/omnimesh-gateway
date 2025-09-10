# Omnimesh Gateway - Playwright Server Example

This example shows how to use the Omnimesh Gateway to connect to the Playwright MCP server.

## Step 1: Register the Playwright MCP Server

Make a POST request to register the Playwright MCP server:

```bash
curl -X POST http://localhost:8080/api/gateway/servers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Playwright",
    "description": "Automate web browsers for testing, scraping, and visual analysis.",
    "protocol": "stdio",
    "command": "npx",
    "args": ["-y", "@executeautomation/playwright-mcp-server"],
    "environment": [],
    "version": "1.0.0",
    "weight": 100,
    "timeout": 300000000000,
    "max_retries": 3
  }'
```

Expected response:
```json
{
  "success": true,
  "data": {
    "id": "server-uuid-here",
    "name": "Playwright",
    "protocol": "stdio",
    "command": "npx",
    "args": ["-y", "@executeautomation/playwright-mcp-server"],
    "status": "active",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

## Step 2: Create an MCP Session

Create a session to start the Playwright MCP server process:

```bash
curl -X POST http://localhost:8080/api/gateway/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "server_id": "server-uuid-here"
  }'
```

Expected response:
```json
{
  "success": true,
  "data": {
    "id": "session-uuid-here",
    "server_id": "server-uuid-here",
    "protocol": "stdio",
    "status": "active",
    "started_at": "2024-01-01T00:00:00Z",
    "process": {
      "pid": 12345,
      "command": "npx",
      "args": ["-y", "@executeautomation/playwright-mcp-server"],
      "status": "running",
      "started_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

## Step 3: Communicate with the MCP Server

Once WebSocket support is implemented, you can connect to:
```
ws://localhost:8080/api/gateway/ws?session_id=session-uuid-here
```

For now, you can send MCP messages through HTTP POST requests to a future endpoint like:
```bash
curl -X POST http://localhost:8080/api/gateway/sessions/session-uuid-here/messages \
  -H "Content-Type: application/json" \
  -d '{
    "id": "msg-1",
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
        "name": "omnimesh-gateway",
        "version": "1.0.0"
      }
    },
    "version": "2024-11-05"
  }'
```

## Step 4: List Active Sessions

Check all active sessions:

```bash
curl -X GET http://localhost:8080/api/gateway/sessions
```

## Step 5: Close the Session

When done, close the session:

```bash
curl -X DELETE http://localhost:8080/api/gateway/sessions/session-uuid-here
```

## Available MCP Methods

Once connected, you can use standard MCP methods like:

- `initialize` - Initialize the MCP session
- `tools/list` - List available tools
- `tools/call` - Call a specific tool
- `resources/list` - List available resources
- `resources/read` - Read a specific resource

## Architecture

```
Client → Omnimesh Gateway → Playwright MCP Server (via stdio)
```

The Omnimesh Gateway:
1. Manages the lifecycle of MCP server processes
2. Provides WebSocket/HTTP interface for clients
3. Handles authentication and rate limiting
4. Logs all interactions for audit purposes
5. Can load balance across multiple instances

## Next Steps

- Add WebSocket support for real-time communication
- Implement proper error handling and recovery
- Add authentication and authorization
- Add metrics and monitoring
- Support for HTTP-based MCP servers
- Session persistence and recovery
