# MCP Gateway

[![CI](https://github.com/yourusername/mcp-gateway/workflows/CI/badge.svg)](https://github.com/yourusername/mcp-gateway/actions)
[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/mcp-gateway)](https://goreportcard.com/report/github.com/yourusername/mcp-gateway)
[![codecov](https://codecov.io/gh/yourusername/mcp-gateway/branch/main/graph/badge.svg)](https://codecov.io/gh/yourusername/mcp-gateway)
[![Security](https://img.shields.io/badge/Security-Enabled-green.svg)](SECURITY.md)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

A production-ready API gateway for Model Context Protocol (MCP) servers, providing enterprise-grade infrastructure with authentication, logging, rate limiting, server discovery, and multi-protocol transport support.

> **⚡ Enterprise-Ready**: Built for production with comprehensive security, monitoring, and scalability features.

## Features

- **🔐 Authentication & Authorization** - JWT-based auth with API keys and role-based access control
- **📊 Comprehensive Logging** - Request/response logging, audit trails, performance metrics, and security events
- **⚡ Rate Limiting** - Per-user, per-organization, and per-endpoint rate limiting with multiple algorithms
- **🛡️ IP Rate Limiting** - Redis-backed sliding window or in-memory per-IP rate limiting with smart proxy detection
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

See [.claude/architecture.md](./.claude/architecture.md) for detailed architectural documentation.

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

## IP Rate Limiting

The MCP Gateway includes sophisticated IP-based rate limiting to protect against abuse and ensure fair resource usage across clients.

### 🎯 **Key Features**
- **Redis Sliding Window** - Precise rate limiting using Redis sorted sets for distributed deployments
- **Memory Fallback** - Automatic fallback to in-memory token bucket when Redis is unavailable
- **Smart IP Detection** - Extracts real client IPs from X-Real-IP, X-Forwarded-For headers with proxy support
- **Path Exclusions** - Skip rate limiting for health checks, metrics, and other system endpoints
- **Configurable Limits** - Flexible per-minute request limits with burst capacity

### 🛠️ **Configuration**
IP rate limiting is automatically enabled in the middleware chain with sensible defaults:

```yaml
# Development: Redis-backed with 100 req/min default
redis:
  enabled: true
  host: "localhost"
  port: 6379
  password: "password123"

# Automatically applied to all routes except /health, /metrics
```

### 🔧 **Custom Configuration**
```go
// Memory-based rate limiting (single instance)
router.Use(middleware.IPRateLimitWithMemory(60)) // 60 requests per minute

// Redis-based rate limiting (distributed)
router.Use(middleware.IPRateLimitWithRedis(100, "localhost:6379", "password", 0))

// Advanced configuration
config := &middleware.IPRateLimitConfig{
    RequestsPerMin: 120,
    RedisEnabled:   true,
    SkipPaths:     []string{"/health", "/metrics", "/debug"},
    CustomHeaders: []string{"X-Real-IP", "CF-Connecting-IP"},
}
router.Use(middleware.IPRateLimit(config))
```

### 📊 **Rate Limit Response**
When rate limits are exceeded, clients receive a structured JSON response:
```json
{
  "error": "rate_limit_exceeded",
  "message": "Too many requests from this IP address. Please try again later.",
  "code": "RATE_LIMIT_EXCEEDED"
}
```

### 🚀 **Performance**
- **Redis Mode**: Supports thousands of concurrent clients with precise sliding window tracking
- **Memory Mode**: High-performance in-memory storage with automatic cleanup for single-instance deployments
- **Smart Fallback**: Automatic Redis → Memory fallback ensures service continuity

## 🚀 Quick Start

### Prerequisites

- **Go 1.25+** - [Installation Guide](https://golang.org/doc/install)
- **PostgreSQL 12+** - Database backend
- **Redis (optional)** - For distributed rate limiting and caching
- **Docker & Docker Compose** - For development environment

### 1. Clone and Setup

```bash
# Clone the repository
git clone https://github.com/yourusername/mcp-gateway.git
cd mcp-gateway

# Copy environment configuration
cp .env.example .env
# Edit .env with your database credentials and JWT secret

# Install pre-commit hooks (recommended)
make setup-precommit
```

### 2. Start Development Environment

```bash
# Start PostgreSQL and Redis with Docker
make docker-run

# Run database migrations
make migrate

# Create admin user (email: admin@admin.com, password: qwerty123)
make setup-admin
```

### 3. Run the Application

```bash
# Start the backend server
make run

# Or with live reload for development
make watch
```

The API will be available at `http://localhost:8080`

### 4. Frontend Dashboard (Optional)

```bash
cd apps/frontend
bun install
bun run dev
```

The dashboard will be available at `http://localhost:3000`

## 🛠️ Development

### Make Commands

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

### Code Quality

```bash
# Run linting
make lint

# Run linting with fixes
make lint-fix

# Run security checks
make security

# Run pre-commit on all files
make precommit-all
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Development Setup

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Run linting (`make lint`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/).

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🔒 Security

Security is a top priority. Please review our [Security Policy](SECURITY.md) and report vulnerabilities responsibly.

## 🙏 Acknowledgments

- [Model Context Protocol](https://modelcontextprotocol.io/) for the core specification
- [Gin Framework](https://gin-gonic.com/) for the HTTP framework
- All our [contributors](https://github.com/yourusername/mcp-gateway/contributors)

---

<div align="center">
  <p>Built with ❤️ for the MCP community</p>
  <p>
    <a href="https://github.com/yourusername/mcp-gateway">⭐ Star us on GitHub</a> •
    <a href="https://github.com/yourusername/mcp-gateway/issues">🐛 Report Bug</a> •
    <a href="https://github.com/yourusername/mcp-gateway/issues">💡 Request Feature</a>
  </p>
</div>
