# MCP Gateway

A production-ready API gateway for Model Context Protocol (MCP) servers, providing organization-level policies, authentication, logging, rate limiting, and server discovery.

## Features

- **🔐 Authentication & Authorization** - JWT-based auth with API keys and role-based access control
- **📊 Comprehensive Logging** - Request/response logging, audit trails, performance metrics, and security events
- **⚡ Rate Limiting** - Per-user, per-organization, and per-endpoint rate limiting with multiple algorithms
- **🔍 MCP Server Discovery** - Dynamic server registration, health checking, and load balancing
- **⚙️ Policy Management** - Flexible organization-level policies for access control and routing
- **🌐 Service Virtualization** - Wrap non-MCP services (REST APIs, GraphQL, gRPC) as virtual MCP servers
- **🔌 Multi-Protocol Support** - JSON-RPC 2.0, WebSocket, SSE, and HTTP transports for MCP communication
- **🚀 High Performance** - Built with Go and Gin for maximum throughput and low latency

## Architecture

The MCP Gateway is designed with a modular architecture for scalability and maintainability:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Client    │    │   CLI Client    │    │  Other Client   │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                   ┌─────────────▼─────────────┐
                   │      MCP Gateway          │
                   │  ┌─────────────────────┐  │
                   │  │   Middleware Chain  │  │
                   │  │ • Auth & AuthZ      │  │
                   │  │ • Rate Limiting     │  │
                   │  │ • Logging & Audit   │  │
                   │  │ • CORS & Security   │  │
                   │  └─────────────────────┘  │
                   │  ┌─────────────────────┐  │
                   │  │   Core Services     │  │
                   │  │ • Server Discovery  │  │
                   │  │ • Load Balancing    │  │
                   │  │ • Policy Engine     │  │
                   │  │ • Virtual Servers   │  │
                   │  │ • Config Management │  │
                   │  └─────────────────────┘  │
                   └─────────────┬─────────────┘
                                 │
          ┌──────────────────────┼──────────────────────┐
          │                      │                      │
┌─────────▼───────┐    ┌─────────▼───────┐    ┌─────────▼───────┐
│  MCP Server A   │    │  Virtual MCP    │    │  MCP Server C   │
│  (AI Assistant) │    │  (REST/GraphQL) │    │  (Data Access)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                       ┌─────────▼───────┐
                       │   External API  │
                       │ (Slack, GitHub, │
                       │  Stripe, etc.)  │
                       └─────────────────┘
```

See [ARCHITECTURE.md](./ARCHITECTURE.md) for detailed architectural documentation.

## Virtual Server Feature

The MCP Gateway includes a powerful virtualization feature that allows you to wrap non-MCP services as MCP-compatible servers. This enables you to:

### 🎯 **Unify API Access**
- Present REST APIs, GraphQL endpoints, and other services through a consistent MCP interface
- Transform any HTTP-based service into an MCP server with tools, prompts, and resources
- Centralize access control and rate limiting across all services

### 🛠️ **Supported Protocols**
- **REST APIs** - Convert HTTP endpoints into MCP tools with automatic parameter mapping
- **GraphQL** - Expose GraphQL queries and mutations as MCP tools *(coming soon)*
- **gRPC** - Bridge gRPC services to MCP protocol *(coming soon)*
- **SOAP** - Legacy SOAP web services support *(coming soon)*

### 📋 **Key Features**
- **JSON-RPC 2.0 Interface** - Standard MCP protocol support via `POST /mcp/rpc`
- **Admin REST API** - Full CRUD operations for virtual server management
- **Database Persistence** - Virtual servers stored in PostgreSQL with in-memory caching
- **Mock Responses** - Built-in testing with mock data for development
- **Error Handling** - Proper JSON-RPC error code mapping (-32601, -32602, -32000)
- **Tool Configuration** - Flexible tool definitions with schema validation

### 🚀 **Example Use Cases**
- **Slack Integration** - Expose Slack's REST API as MCP tools for sending messages and listing channels
- **GitHub Operations** - Wrap GitHub API for repository management, issue creation, and more
- **Database Access** - Convert database queries into MCP tools with proper authentication
- **Third-party SaaS** - Integrate any REST-based service (Stripe, Twilio, etc.) into your MCP workflow

### 📖 **Quick Example**
```bash
# Create a virtual Slack server
curl -X POST http://localhost:8080/api/admin/virtual-servers \
  -H "Content-Type: application/json" \
  -d '{
    "id": "slack_server",
    "name": "Slack API",
    "description": "Slack REST API as MCP server",
    "adapterType": "REST",
    "tools": [...] 
  }'

# Use the virtual server via MCP JSON-RPC
curl -X POST http://localhost:8080/mcp/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": "1",
    "method": "tools/call",
    "params": {
      "server_id": "slack_server",
      "name": "send_message",
      "arguments": {
        "channel": "#general",
        "text": "Hello from MCP Gateway!"
      }
    }
  }'
```

See [examples/virtual_servers_example.md](./examples/virtual_servers_example.md) for comprehensive usage examples and testing guide.

## Quick Start

## MakeFile

Run build make command with tests
```bash
make all
```

Build the application
```bash
make build
```

Run the application
```bash
make run
```
Create DB container
```bash
make docker-run
```

Shutdown DB Container
```bash
make docker-down
```

DB Integrations Test:
```bash
make itest
```

Live reload the application:
```bash
make watch
```

Run the test suite:
```bash
make test
```

Clean up binary from the last build:
```bash
make clean
```
