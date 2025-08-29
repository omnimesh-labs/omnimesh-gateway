# MCP Gateway Testing Guide

This guide walks through setting up a practical MCP Gateway environment with real servers and endpoints.

## 📋 Prerequisites

Before starting, make sure you have:
- Computer with **Docker** installed ([Download Docker Desktop](https://www.docker.com/products/docker-desktop/))
- **Node.js** version 18+ ([Download Node.js](https://nodejs.org/))
- Basic terminal/command prompt access
- Internet connection for downloading MCP servers

---

## 🚀 Part 1: Starting the MCP Gateway

1. **Clone and Start the Gateway**
   ```bash
   git clone https://github.com/janex-ai/janex-gateway.git
   cd mcp-gateway
   make start
   ```

2. **Wait for Services**
   ```
   ✅ Backend running at: http://localhost:8080
   ✅ Frontend running at: http://localhost:3000
   ✅ Database ready
   ✅ Redis ready
   ```

3. **Login to Dashboard**
   - Open browser to `http://localhost:3000`
   - Login with: **admin@admin.com** / **qwerty123**

---

## 🌐 Part 2: Setting Up Your Namespace

A namespace is your isolated environment for MCP servers and endpoints.

1. **Create a Namespace**
   - In dashboard, go to "Namespaces" → "Create New Namespace"
   ```
   Name: Development Environment
   Slug: dev-env
   Description: Development testing environment for MCP servers
   ```

2. **Create API Key**
   - In your namespace, go to "API Keys" → "Create New Key"
   ```
   Name: Development Key
   Description: Key for testing MCP endpoints
   Expiration: 30 days
   ```
   - Save the API key securely - you'll need it later

---

## 🔧 Part 3: Adding MCP Servers

Let's add some useful MCP servers to your namespace.

1. **Install MCP Servers Globally**
   ```bash
   # Install common MCP servers
   npm install -g @modelcontextprotocol/server-filesystem
   npm install -g @modelcontextprotocol/server-github
   npm install -g @modelcontextprotocol/server-brave-search
   ```

2. **Add Filesystem Server**
   - In your namespace, go to "Servers" → "Add Server"
   ```
   Name: Local Files
   Description: Access to local filesystem
   Protocol: stdio
   Command: npx
   Arguments: ["@modelcontextprotocol/server-filesystem", "/tmp"]
   ```

3. **Add GitHub Server**
   - Create a GitHub token at github.com/settings/tokens
   - Add another server:
   ```
   Name: GitHub Access
   Description: GitHub repository access
   Protocol: stdio
   Command: npx
   Arguments: ["@modelcontextprotocol/server-github"]
   Environment: ["GITHUB_TOKEN=your_token_here"]
   ```

4. **Add Search Server**
   - Get API key from api.search.brave.com
   - Add another server:
   ```
   Name: Web Search
   Description: Internet search capabilities
   Protocol: stdio
   Command: npx
   Arguments: ["@modelcontextprotocol/server-brave-search"]
   Environment: ["BRAVE_API_KEY=your_key_here"]
   ```

---

## 🎯 Part 4: Creating an Endpoint

An endpoint is how AI clients will access your MCP servers.

1. **Create Endpoint**
   - In your namespace, go to "Endpoints" → "Create New Endpoint"
   ```
   Name: Development API
   Description: Main development endpoint
   Base Path: /dev-api
   ```

2. **Configure Servers**
   - In your endpoint settings, enable the servers you added
   - Set rate limits if desired
   - Configure any server-specific settings

3. **Get Endpoint URL**
   - Copy your endpoint URL, it will look like:
   ```
   http://localhost:8080/dev-api
   ```

---

## 🔌 Part 5: Connecting to Your Endpoint

Now let's test connecting to your endpoint using different transport protocols.

### 5.1 Creating an API Key

First, you'll need an API key to authenticate with the endpoints:

```bash
# Login to get JWT token
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@admin.com",
    "password": "qwerty123"
  }'

# Create API key (use the access_token from login response)
curl -X POST http://localhost:8080/api/auth/api-keys \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${BEARER_TOKEN}" \
  -d '{
    "name": "Testing Key",
    "role": "system_admin",
    "expires_at": "2025-12-31T23:59:59Z"
  }'
```

Save the API key from the response - you'll use it in all the examples below.

### 5.2 Transport Endpoints

Your gateway provides multiple transport protocols for MCP communication:

#### A. JSON-RPC over HTTP

```bash
# Basic transport endpoint (no namespace routing)
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/list",
    "params": {},
    "id": "test-1"
  }'

# Batch RPC calls
curl -X POST http://localhost:8080/rpc/batch \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '[
    {"jsonrpc": "2.0", "method": "tools/list", "params": {}, "id": "1"},
    {"jsonrpc": "2.0", "method": "resources/list", "params": {}, "id": "2"}
  ]'

# Server-specific RPC (routes to specific server)
curl -X POST http://localhost:8080/servers/SERVER_ID/rpc \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "read_file",
      "arguments": {"path": "/tmp/test.txt"}
    },
    "id": "server-test-1"
  }'
```

#### B. WebSocket Transport

```bash
# Install wscat for WebSocket testing
npm install -g wscat

# Basic WebSocket connection
wscat -c ws://localhost:8080/ws -H "X-API-Key: YOUR_API_KEY"

# Server-specific WebSocket
wscat -c ws://localhost:8080/servers/SERVER_ID/ws -H "X-API-Key: YOUR_API_KEY"

# Once connected, send MCP messages:
{"jsonrpc": "2.0", "method": "tools/list", "params": {}, "id": "ws-1"}
{"jsonrpc": "2.0", "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test-client", "version": "1.0.0"}}, "id": "init-1"}
```

#### C. Server-Sent Events (SSE)

```bash
# Listen for SSE events
curl -N -H "Accept: text/event-stream" \
  -H "X-API-Key: YOUR_API_KEY" \
  http://localhost:8080/sse

# Send events to SSE stream (in another terminal)
curl -X POST http://localhost:8080/sse/events \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "event": "mcp_request",
    "data": {
      "jsonrpc": "2.0",
      "method": "tools/list",
      "params": {},
      "id": "sse-1"
    }
  }'

# Server-specific SSE
curl -N -H "Accept: text/event-stream" \
  -H "X-API-Key: YOUR_API_KEY" \
  http://localhost:8080/servers/SERVER_ID/sse
```

#### D. MCP Protocol (Streamable HTTP)

```bash
# MCP protocol initialization
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "jsonrpc": "2.0",
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "capabilities": {"tools": {"listChanged": true}},
      "clientInfo": {"name": "test-client", "version": "1.0.0"}
    },
    "id": "init-1"
  }'

# Server-specific MCP
curl -X POST http://localhost:8080/servers/SERVER_ID/mcp \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/list",
    "params": {},
    "id": "mcp-1"
  }'
```

#### E. STDIO Bridge

```bash
# Execute command via STDIO
curl -X POST http://localhost:8080/stdio/execute \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "command": "npx",
    "args": ["@modelcontextprotocol/server-filesystem", "/tmp"],
    "timeout": "30s"
  }'

# Send to STDIO session (use session_id from execute response)
curl -X POST http://localhost:8080/stdio/send \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "X-Session-ID: YOUR_SESSION_ID" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/list",
    "params": {},
    "id": "stdio-1"
  }'
```

### 5.3 Public Namespace Endpoints

For namespace-scoped access (recommended for production):

```bash
# List tools via public endpoint
curl -X GET http://localhost:8080/api/public/endpoints/ENDPOINT_NAME/api/tools \
  -H "X-API-Key: YOUR_API_KEY"

# Execute tool via public endpoint
curl -X POST http://localhost:8080/api/public/endpoints/ENDPOINT_NAME/api/tools/TOOL_NAME \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "arguments": {
      "path": "/tmp/test.txt"
    }
  }'

# MCP protocol via public endpoint
curl -X POST http://localhost:8080/api/public/endpoints/ENDPOINT_NAME/mcp \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/list",
    "params": {},
    "id": "public-1"
  }'

# WebSocket via public endpoint
wscat -c ws://localhost:8080/api/public/endpoints/ENDPOINT_NAME/ws \
  -H "X-API-Key: YOUR_API_KEY"
```

---

## 🤖 Part 6: Testing with Real AI Clients

### Claude Desktop Setup

Configure Claude Desktop to connect to your gateway:

1. **Edit Claude Desktop Config**
   ```bash
   # Mac
   nano ~/Library/Application\ Support/Claude/claude_desktop_config.json
   
   # Windows
   notepad %APPDATA%\Claude\claude_desktop_config.json
   
   # Linux
   nano ~/.config/Claude/claude_desktop_config.json
   ```

2. **Add Gateway Configuration**
   ```json
   {
     "mcpServers": {
       "mcp-gateway": {
         "type": "http",
         "url": "http://localhost:8080/rpc",
         "headers": {
           "X-API-Key": "YOUR_API_KEY"
         }
       },
       "mcp-gateway-namespace": {
         "type": "http", 
         "url": "http://localhost:8080/api/public/endpoints/ENDPOINT_NAME/mcp",
         "headers": {
           "X-API-Key": "YOUR_API_KEY"
         }
       }
     }
   }
   ```

3. **Test with Claude**
   - Restart Claude Desktop
   - Start a new chat
   - Test MCP functionality:
     ```
     Can you list the available tools?
     Can you read a file from /tmp?
     Can you search GitHub for MCP servers?
     ```

### MCP Inspector Integration

You can also test using external MCP Inspector tools:

```bash
# Install MCP Inspector
npm install -g @modelcontextprotocol/inspector

# Connect Inspector to your gateway
mcp-inspector http://localhost:8080/rpc \
  --header "X-API-Key=YOUR_API_KEY"

# Or use namespace-scoped endpoint
mcp-inspector http://localhost:8080/api/public/endpoints/ENDPOINT_NAME/mcp \
  --header "X-API-Key=YOUR_API_KEY"
```

---

## 🔍 Verification Checklist

Test that everything is working correctly:

### ✅ Basic Connectivity
- [ ] Can create API keys via login + auth endpoints
- [ ] Basic RPC endpoint responds: `POST /rpc`
- [ ] WebSocket connects successfully: `ws://localhost:8080/ws`
- [ ] SSE stream works: `GET /sse`

### ✅ MCP Protocol Compliance  
- [ ] Initialize handshake works
- [ ] Tools listing returns results
- [ ] Tool execution succeeds
- [ ] Resources listing works (if configured)
- [ ] Prompts listing works (if configured)

### ✅ Transport Layer Testing
- [ ] JSON-RPC: Single and batch requests work
- [ ] WebSocket: Bidirectional messaging works  
- [ ] SSE: Event streaming and broadcasting works
- [ ] MCP: Protocol-compliant responses
- [ ] STDIO: Command execution and session management

### ✅ Namespace Integration
- [ ] Public endpoints route correctly to namespaces
- [ ] Server-specific routing works
- [ ] Authentication and rate limiting active
- [ ] Tools from multiple servers aggregate properly

### ✅ AI Client Integration
- [ ] Claude Desktop connects and lists tools
- [ ] External MCP Inspector connects successfully
- [ ] Tools execute and return expected results
- [ ] Error handling works gracefully

---

## 🚨 Troubleshooting

**Common Issues:**

1. **Authentication Errors**
   ```bash
   # Check if JWT is valid
   curl -H "Authorization: Bearer YOUR_JWT" http://localhost:8080/api/auth/profile
   
   # Verify API key works
   curl -H "X-API-Key: YOUR_KEY" http://localhost:8080/rpc -X POST -d '{"jsonrpc":"2.0","method":"ping","id":"1"}'
   ```

2. **Transport Connection Issues**
   ```bash
   # Test basic connectivity
   curl -I http://localhost:8080/health
   
   # Check WebSocket upgrade
   curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" http://localhost:8080/ws
   ```

3. **MCP Server Registration**
   ```bash
   # List registered servers
   curl -H "Authorization: Bearer YOUR_JWT" http://localhost:8080/api/gateway/servers
   
   # Check server health
   curl -H "Authorization: Bearer YOUR_JWT" http://localhost:8080/api/gateway/servers/SERVER_ID/stats
   ```

---

## 📚 Next Steps

Once everything is working:

1. **Scale Up**: Add more MCP servers and create additional namespaces
2. **Production Setup**: Configure HTTPS, rate limiting, and monitoring  
3. **Custom Development**: Build your own MCP servers for specific use cases
4. **Integration**: Connect other AI clients and applications
5. **Monitoring**: Set up logging and metrics collection

### Resources
- [MCP Protocol Specification](https://modelcontextprotocol.io/)
- [Official MCP Servers](https://github.com/modelcontextprotocol/servers)  
- [Claude Desktop Documentation](https://claude.ai/docs)
- [Gateway API Reference](../api/README.md)
