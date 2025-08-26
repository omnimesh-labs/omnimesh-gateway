# Janex Gateway

[![CI](https://github.com/theognis1002/mcp-gateway/workflows/CI/badge.svg)](https://github.com/theognis1002/mcp-gateway/actions)
[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/theognis1002/mcp-gateway)](https://goreportcard.com/report/github.com/theognis1002/mcp-gateway)
[![codecov](https://codecov.io/gh/theognis1002/mcp-gateway/branch/main/graph/badge.svg)](https://codecov.io/gh/theognis1002/mcp-gateway)
[![Security](https://img.shields.io/badge/Security-Enabled-green.svg)](SECURITY.md)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

A production-ready API gateway for Model Context Protocol (MCP) servers, providing enterprise-grade infrastructure with authentication, logging, rate limiting, server discovery, and multi-protocol transport support.

> **⚡ Enterprise-Ready**: Built for production with comprehensive security, monitoring, and scalability features.

## 🚀 Quick Start

Get the entire Janex Gateway stack running with a single command:

```bash
# Clone the repository
git clone https://github.com/theognis1002/mcp-gateway.git
cd mcp-gateway

# Start everything (PostgreSQL, Redis, Backend, Frontend, Migrations)
make dev
```

This will:
- Start PostgreSQL and Redis databases
- Run database migrations automatically
- Start the backend API server on http://localhost:8080 with hot reload
- Start the frontend dashboard on http://localhost:3000 with hot reload
- Create an admin user: `admin@admin.com` / `qwerty123`

### Available Commands

```bash
make help         # Show all available commands
make dev          # Start in development mode with hot reload
make stop         # Stop all services
make clean        # Stop and remove all data
make test         # Run tests in Docker
make migrate      # Run database migrations
make lint         # Run linters
make logs         # View service logs
```

## Features

- **🔐 Authentication & Authorization** - JWT-based auth with API keys and role-based access control
- **📊 Comprehensive Logging** - Request/response logging, audit trails, performance metrics, and security events
- **⚡ Rate Limiting** - Per-user, per-organization, and per-endpoint rate limiting with multiple algorithms
- **🛡️ IP Rate Limiting** - Redis-backed sliding window or in-memory per-IP rate limiting with smart proxy detection and configurable limits
- **🔍 MCP Server Discovery** - Dynamic server registration, health checking, and load balancing
- **⚙️ Policy Management** - Flexible organization-level policies for access control and routing
- **🌐 Service Virtualization** - Wrap non-MCP services (REST APIs, GraphQL, gRPC) as virtual MCP servers
- **🔌 Multi-Protocol Support** - JSON-RPC 2.0, WebSocket, SSE, and HTTP transports for MCP communication
- **🚀 High Performance** - Built with Go and Gin for maximum throughput and low latency

## Architecture

The Janex Gateway is designed with a modular architecture for scalability and maintainability:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Client    │    │   CLI Client    │    │  Other Client   │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                   ┌─────────────▼─────────────┐
                   │      Janex Gateway        │
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

The Janex Gateway includes a powerful virtualization feature that allows you to wrap non-MCP services as MCP-compatible servers. This enables you to:

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
        "text": "Hello from Janex Gateway!"
      }
    }
  }'
```

See [examples/virtual_servers_example.md](./examples/virtual_servers_example.md) for comprehensive usage examples and testing guide.

## 🛠️ Development

### Docker-Based Development (Recommended)

The entire stack runs in Docker containers with hot reload enabled for both backend and frontend:

```bash
# Start all services (PostgreSQL, Redis, Backend, Frontend)
make dev

# View logs
make logs

# Stop all services
make stop

# Clean everything (removes all data)
make clean
```

### Other Useful Commands

```bash
# Database operations
make migrate          # Run database migrations
make migrate-down     # Rollback migrations
make migrate-status   # Show migration status
make db-shell         # Open PostgreSQL shell

# Testing
make test             # Run all tests
make test-transport   # Transport layer tests
make test-integration # Integration tests
make test-unit        # Unit tests

# Development utilities
make shell            # Open shell in backend container
make bash             # Open bash in backend container
make redis-cli        # Open Redis CLI

# Build operations
make build            # Build containers
make rebuild          # Rebuild and restart containers

# Setup
make setup            # Initial project setup
make setup-admin      # Create admin user
```

### Code Quality

```bash
# Run linting
make lint
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Development Setup

1. Fork the repository
2. Clone your fork and enter the directory
3. Run `make dev` to start the full stack
4. Make your changes
5. Run tests (`make test`)
6. Run linting (`make lint`)
7. Commit your changes (`git commit -m 'Add amazing feature'`)
8. Push to the branch (`git push origin feature/amazing-feature`)
9. Open a Pull Request

### Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/).

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🔒 Security

Security is a top priority. Please review our [Security Policy](SECURITY.md) and report vulnerabilities responsibly.

## 🙏 Acknowledgments

- [Model Context Protocol](https://modelcontextprotocol.io/) for the core specification
- All our [contributors](https://github.com/theognis1002/mcp-gateway/contributors)

---

<div align="center">
  <p>Built with ❤️ for the MCP community</p>
  <p>
    <a href="https://github.com/theognis1002/mcp-gateway">⭐ Star us on GitHub</a> •
    <a href="https://github.com/theognis1002/mcp-gateway/issues">🐛 Report Bug</a> •
    <a href="https://github.com/theognis1002/mcp-gateway/issues">💡 Request Feature</a>
  </p>
</div>
