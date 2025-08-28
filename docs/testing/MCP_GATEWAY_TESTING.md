# MCP Gateway Testing Guide

A comprehensive guide for non-technical users to test the Janex MCP Gateway with real MCP clients and servers.

## 📋 Prerequisites

Before starting, make sure you have:
- Computer with **Docker** installed ([Download Docker Desktop](https://www.docker.com/products/docker-desktop/))
- **Node.js** version 18+ ([Download Node.js](https://nodejs.org/))
- Basic terminal/command prompt access
- Internet connection for downloading MCP servers

---

## 🚀 Part 1: Starting the MCP Gateway

### Step 1: Get the Gateway Running

1. **Open Terminal/Command Prompt**
   - Windows: Press `Win + R`, type `cmd`, press Enter
   - Mac: Press `Cmd + Space`, type `terminal`, press Enter
   - Linux: Press `Ctrl + Alt + T`

2. **Clone and Start the Gateway**
   ```bash
   # Clone the repository
   git clone https://github.com/janex-ai/janex-gateway.git
   cd mcp-gateway
   
   # Start everything with one command
   make start
   ```

3. **Wait for Services to Start**
   Look for these success messages:
   ```
   ✅ Backend running at: http://localhost:8080
   ✅ Frontend running at: http://localhost:3000
   ✅ Database ready
   ✅ Redis ready
   ```

4. **Verify the Gateway is Working**
   - Open browser to `http://localhost:3000`
   - Login with: **admin@admin.com** / **qwerty123**
   - You should see the dashboard

> **❌ If something goes wrong:** Run `make stop` then `make clean` then try `make start` again

---

## 🔍 Part 2: Setting Up MCP Inspector (Your Test Client)

MCP Inspector is the easiest way to test MCP servers. It's like a web browser but for MCP.

### Step 2.1: Install MCP Inspector

```bash
# Install globally
npm install -g @modelcontextprotocol/inspector

# Verify it's working
npx @modelcontextprotocol/inspector --help
```

### Step 2.2: Test MCP Inspector with a Simple Server

Let's start with a basic filesystem MCP server:

1. **Install the Filesystem MCP Server**
   ```bash
   npm install -g @modelcontextprotocol/server-filesystem
   ```

2. **Start MCP Inspector with the Filesystem Server**
   ```bash
   npx @modelcontextprotocol/inspector npx @modelcontextprotocol/server-filesystem /tmp
   ```

3. **Open the Inspector Interface**
   - Look for: `Inspector available at http://localhost:6274`
   - Open that URL in your browser
   - You should see the MCP Inspector interface

4. **Test Basic MCP Operations**
   - Click "Connect" (should show ✅ Connected)
   - Click "List Tools" - should show filesystem tools like `read_file`, `write_file`
   - Click "List Resources" - should show files in `/tmp` directory

> **✅ Success indicator:** You see tools and resources listed in the Inspector interface

---

## 🏢 Part 3: Creating Namespaces and Servers in the Gateway

Now let's add MCP servers to your gateway so they can be accessed by AI clients.

### Step 3.1: Create Your First Namespace

1. **Open the Gateway Dashboard**
   - Go to `http://localhost:3000`
   - Login: **admin@admin.com** / **qwerty123**

2. **Navigate to Namespaces**
   - Click "Namespaces" in the sidebar
   - Click "Create New Namespace"

3. **Fill in Namespace Details**
   ```
   Name: My Test Namespace
   Slug: test-namespace
   Description: Testing MCP servers and protocols
   ```

4. **Click "Create Namespace"**

### Step 3.2: Add an MCP Server to Your Namespace

1. **Go to MCP Servers Section**
   - From the namespace page, click "Add Server" or navigate to "Servers"

2. **Create a Filesystem Server**
   ```
   Server Name: Filesystem Server
   Description: Access local files
   Protocol: stdio
   Command: npx
   Arguments: ["@modelcontextprotocol/server-filesystem", "/tmp"]
   Namespace: test-namespace
   ```

3. **Click "Create Server"**

4. **Verify Server Status**
   - Server status should show "Active" with a green ✅
   - If red ❌, check that the filesystem server is installed

---

## 🧪 Part 4: Testing All Transport Protocols

The gateway supports multiple ways to communicate with MCP servers. Let's test each one.

### Step 4.1: Test HTTP JSON-RPC (Easiest)

This is like making API calls to your MCP servers.

1. **Test Basic Connectivity**
   ```bash
   curl -X POST http://localhost:8080/rpc \
     -H "Content-Type: application/json" \
     -d '{
       "jsonrpc": "2.0",
       "method": "tools/list",
       "params": {},
       "id": "test-1"
     }'
   ```

2. **Expected Response:**
   ```json
   {
     "jsonrpc": "2.0",
     "id": "test-1",
     "result": {
       "tools": [
         {"name": "read_file", "description": "Read a file..."},
         {"name": "write_file", "description": "Write to a file..."}
       ]
     }
   }
   ```

3. **Test Calling a Tool**
   ```bash
   curl -X POST http://localhost:8080/rpc \
     -H "Content-Type: application/json" \
     -d '{
       "jsonrpc": "2.0",
       "method": "tools/call",
       "params": {
         "name": "read_file",
         "arguments": {"path": "/tmp"}
       },
       "id": "test-2"
     }'
   ```

> **✅ Success:** You get JSON responses with tool lists and file contents

### Step 4.2: Test WebSocket (Real-time)

WebSocket allows real-time bidirectional communication.

1. **Install a WebSocket Client**
   ```bash
   npm install -g websocat
   ```

2. **Connect to Gateway WebSocket**
   ```bash
   websocat ws://localhost:8080/ws
   ```

3. **Send a Message** (type this in the WebSocket connection):
   ```json
   {"jsonrpc": "2.0", "method": "tools/list", "params": {}, "id": "ws-test"}
   ```

4. **You Should See Response:**
   ```json
   {"jsonrpc": "2.0", "id": "ws-test", "result": {"tools": [...]}}
   ```

> **✅ Success:** You can send messages and get real-time responses

### Step 4.3: Test Server-Sent Events (SSE - Streaming)

SSE is great for streaming responses from MCP servers.

1. **Start SSE Connection**
   ```bash
   curl -N -H "Accept: text/event-stream" http://localhost:8080/sse
   ```

2. **In Another Terminal, Send Event**
   ```bash
   curl -X POST http://localhost:8080/sse/events \
     -H "Content-Type: application/json" \
     -d '{
       "event": "message",
       "data": {
         "type": "mcp_request",
         "method": "tools/list"
       }
     }'
   ```

3. **You Should See Streaming Response:**
   ```
   event: message
   data: {"tools": [...]}
   ```

### Step 4.4: Test STDIO (Command Line Interface)

STDIO is perfect for command-line MCP servers.

1. **Execute Direct Command**
   ```bash
   curl -X POST http://localhost:8080/stdio/execute \
     -H "Content-Type: application/json" \
     -d '{
       "command": "npx",
       "args": ["@modelcontextprotocol/server-filesystem", "/tmp"],
       "timeout": "30s"
     }'
   ```

2. **Send MCP Message to Process**
   ```bash
   curl -X POST http://localhost:8080/stdio/send \
     -H "Content-Type: application/json" \
     -H "X-Session-ID: your-session-id" \
     -d '{
       "id": "init-1",
       "type": "request", 
       "method": "initialize",
       "params": {
         "protocolVersion": "2024-11-05",
         "capabilities": {},
         "clientInfo": {"name": "gateway-test", "version": "1.0"}
       }
     }'
   ```

---

## 🤖 Part 5: Testing with Claude Desktop (Real AI Client)

Now let's connect a real AI client to your gateway.

### Step 5.1: Install Claude Desktop

1. **Download Claude Desktop**
   - Go to [claude.ai/download](https://claude.ai/download)
   - Download for your operating system
   - Install and create account

### Step 5.2: Configure Claude to Use Your Gateway

1. **Find Claude's Config File**
   - Mac: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - Windows: `%APPDATA%\Claude\claude_desktop_config.json`
   - Linux: `~/.config/Claude/claude_desktop_config.json`

2. **Add Gateway Server to Config**
   ```json
   {
     "mcpServers": {
       "gateway-filesystem": {
         "command": "curl",
         "args": [
           "-X", "POST",
           "http://localhost:8080/rpc",
           "-H", "Content-Type: application/json",
           "-d", "{\"jsonrpc\":\"2.0\",\"method\":\"tools/list\",\"params\":{},\"id\":\"claude-1\"}"
         ]
       }
     }
   }
   ```

3. **Restart Claude Desktop**

4. **Test in Claude**
   - Open new chat
   - Ask: "What files are in /tmp?"
   - Claude should use your gateway to list files

> **✅ Success:** Claude can read files through your MCP Gateway

---

## 🌐 Part 6: Testing Popular MCP Servers

Let's test with some popular real-world MCP servers.

### Step 6.1: GitHub MCP Server

1. **Install GitHub MCP Server**
   ```bash
   npm install -g @modelcontextprotocol/server-github
   ```

2. **Get GitHub Token**
   - Go to [github.com/settings/tokens](https://github.com/settings/tokens)
   - Generate new classic token
   - Select scopes: `repo`, `user`

3. **Add GitHub Server to Gateway**
   - In gateway dashboard, create new server:
   ```
   Name: GitHub Server
   Protocol: stdio
   Command: npx
   Arguments: ["@modelcontextprotocol/server-github"]
   Environment: ["GITHUB_PERSONAL_ACCESS_TOKEN=your_token_here"]
   ```

4. **Test GitHub Server**
   ```bash
   curl -X POST http://localhost:8080/rpc \
     -H "Content-Type: application/json" \
     -d '{
       "jsonrpc": "2.0",
       "method": "tools/call",
       "params": {
         "name": "search_repositories",
         "arguments": {"query": "mcp", "limit": 5}
       },
       "id": "github-test"
     }'
   ```

### Step 6.2: Brave Search MCP Server

1. **Install Brave Search Server**
   ```bash
   npm install -g @modelcontextprotocol/server-brave-search
   ```

2. **Get Brave API Key**
   - Go to [api.search.brave.com](https://api.search.brave.com)
   - Sign up and get API key

3. **Add to Gateway**
   ```
   Name: Brave Search
   Protocol: stdio  
   Command: npx
   Arguments: ["@modelcontextprotocol/server-brave-search"]
   Environment: ["BRAVE_API_KEY=your_api_key"]
   ```

4. **Test Web Search**
   ```bash
   curl -X POST http://localhost:8080/rpc \
     -H "Content-Type: application/json" \
     -d '{
       "jsonrpc": "2.0",
       "method": "tools/call", 
       "params": {
         "name": "brave_web_search",
         "arguments": {"query": "MCP protocol"}
       },
       "id": "search-test"
     }'
   ```

---

## 🎛️ Part 7: Testing Virtual Servers (REST API Wrapper)

Virtual servers let you wrap regular REST APIs as MCP servers.

### Step 7.1: Create a Virtual Server

1. **In Gateway Dashboard, Go to Virtual Servers**

2. **Create New Virtual Server**
   ```json
   {
     "name": "JSONPlaceholder API",
     "description": "Test REST API as MCP",
     "adapter_type": "REST",
     "tools": [
       {
         "name": "get_posts",
         "description": "Get blog posts",
         "inputSchema": {
           "type": "object",
           "properties": {
             "userId": {"type": "number", "description": "User ID filter"}
           }
         },
         "method": "GET",
         "url": "https://jsonplaceholder.typicode.com/posts",
         "response_type": "json"
       }
     ]
   }
   ```

3. **Test Virtual Server**
   ```bash
   curl -X POST http://localhost:8080/mcp/rpc \
     -H "Content-Type: application/json" \
     -d '{
       "jsonrpc": "2.0",
       "method": "tools/call",
       "params": {
         "name": "get_posts",
         "arguments": {"userId": 1}
       },
       "id": "virtual-test"
     }'
   ```

> **✅ Success:** You get blog posts from the REST API via MCP protocol

---

## 🔧 Part 8: Advanced Testing with MCP Inspector

### Step 8.1: Test Multiple Servers

1. **Start Inspector with Multiple Servers**
   ```bash
   npx @modelcontextprotocol/inspector \
     npx @modelcontextprotocol/server-filesystem /tmp \
     --server "npx @modelcontextprotocol/server-github" \
     --server "curl -X POST http://localhost:8080/rpc"
   ```

2. **Switch Between Servers in Inspector UI**
   - Use server dropdown in Inspector
   - Each server shows different tools and resources

### Step 8.2: Test Error Handling

1. **Try Invalid Tool Call**
   ```bash
   curl -X POST http://localhost:8080/rpc \
     -H "Content-Type: application/json" \
     -d '{
       "jsonrpc": "2.0",
       "method": "tools/call",
       "params": {
         "name": "nonexistent_tool",
         "arguments": {}
       },
       "id": "error-test"
     }'
   ```

2. **Expected Error Response:**
   ```json
   {
     "jsonrpc": "2.0",
     "id": "error-test",
     "error": {
       "code": -32601,
       "message": "Method not found"
     }
   }
   ```

---

## 🩺 Troubleshooting

### Common Issues and Solutions

#### ❌ "Connection refused" errors
- **Solution:** Make sure gateway is running with `make start`
- Check if ports 8080 and 3000 are free
- Wait for all services to fully start (30-60 seconds)

#### ❌ "Server not found" errors
- **Solution:** Verify server is created and active in dashboard
- Check server logs in gateway UI
- Ensure MCP server package is installed globally

#### ❌ MCP Inspector won't connect
- **Solution:** 
  ```bash
  # Kill any running inspector processes
  pkill -f inspector
  
  # Clear npm cache
  npm cache clean --force
  
  # Reinstall inspector
  npm install -g @modelcontextprotocol/inspector
  ```

#### ❌ Authentication errors with external APIs
- **Solution:** Double-check API keys in environment variables
- Make sure API keys have correct permissions
- Check API key isn't expired

#### ❌ WebSocket connections fail
- **Solution:** Check if your firewall blocks WebSocket connections
- Try testing with different WebSocket client
- Verify WebSocket endpoint is correct

### Getting Help

1. **Check Gateway Logs**
   ```bash
   make logs
   ```

2. **Check Service Status**
   ```bash
   curl http://localhost:8080/health
   ```

3. **Reset Everything**
   ```bash
   make clean
   make start
   ```

---

## 🎉 Success Checklist

By the end of this guide, you should be able to:

- [ ] ✅ Start the MCP Gateway with `make start`
- [ ] ✅ Access the dashboard at `http://localhost:3000`
- [ ] ✅ Create namespaces and MCP servers via the UI
- [ ] ✅ Use MCP Inspector to test servers visually
- [ ] ✅ Make HTTP JSON-RPC calls with curl
- [ ] ✅ Connect via WebSocket for real-time communication
- [ ] ✅ Use SSE for streaming responses
- [ ] ✅ Test STDIO transport for command-line servers
- [ ] ✅ Connect Claude Desktop as a real AI client
- [ ] ✅ Install and test popular MCP servers (GitHub, Brave Search, Filesystem)
- [ ] ✅ Create virtual servers to wrap REST APIs
- [ ] ✅ Handle errors gracefully and troubleshoot issues

## 🚀 Next Steps

Now that you've tested the MCP Gateway, consider:

1. **Production Deployment:** Deploy to cloud services
2. **Custom MCP Servers:** Build your own MCP servers
3. **Advanced Configurations:** Set up authentication, rate limiting
4. **Integration:** Connect to your existing APIs and databases
5. **Monitoring:** Set up logging and metrics for production use

---

## 📚 Additional Resources

- [Official MCP Documentation](https://modelcontextprotocol.io/)
- [MCP Server Examples](https://github.com/modelcontextprotocol/servers)
- [Janex Gateway Documentation](./README.md)
- [Claude Desktop Configuration](https://claude.ai/docs)
- [MCP Inspector GitHub](https://github.com/modelcontextprotocol/inspector)

Happy testing! 🎊