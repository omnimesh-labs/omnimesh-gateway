# MCP Gateway

[![CI](https://github.com/theognis1002/mcp-gateway/workflows/CI/badge.svg)](https://github.com/theognis1002/mcp-gateway/actions)
[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/theognis1002/mcp-gateway)](https://goreportcard.com/report/github.com/theognis1002/mcp-gateway)
[![codecov](https://codecov.io/gh/theognis1002/mcp-gateway/branch/main/graph/badge.svg)](https://codecov.io/gh/theognis1002/mcp-gateway)
[![Security](https://img.shields.io/badge/Security-Enabled-green.svg)](SECURITY.md)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

A production-ready API gateway for Model Context Protocol (MCP) servers, providing enterprise-grade infrastructure with authentication, logging, rate limiting, server discovery, and multi-protocol transport support.


## 🚀 Quick Start

Get the entire MCP Gateway stack running with a single command:

```bash
# Clone the repository
git clone https://github.com/theognis1002/mcp-gateway.git
cd mcp-gateway

# Option 1: Using Docker Compose directly
docker compose up --build

# Option 2: Using Makefile (automatically detects docker compose vs docker-compose)
make start
```

**Access the application:**
- Backend: `http://localhost:8080`
- Frontend: `http://localhost:3000`
- Admin user: `admin@admin.com` / `qwerty123`


## Architecture

The MCP Gateway is designed with a modular architecture for scalability and maintainability:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│      Users      │    │    AI Agents    │    │     AI Agents   │
│   (Web/Mobile)  │    │    (External)   │    │    (Internal)   │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                   ┌─────────────▼─────────────┐
                   │       MCP Gateway         │
                   │                           │
                   │  ┌─────────────────────┐  │
                   │  │   Security Layer    │  │
                   │  │ • JWT Auth          │  │
                   │  │ • RBAC & Policies   │  │
                   │  │ • API Key Mgmt      │  │
                   │  │ • Rate Limiting     │  │
                   │  └─────────────────────┘  │
                   │                           │
                   │  ┌─────────────────────┐  │
                   │  │     Middleware      │  │
                   │  │ • Content Filtering │  │
                   │  │ • Audit Logging     │  │
                   │  │ • CORS & Headers    │  │
                   │  │ • Request Tracking  │  │
                   │  └─────────────────────┘  │
                   │                           │
                   │  ┌─────────────────────┐  │
                   │  │   Core Services     │  │
                   │  │ • Server Discovery  │  │
                   │  │ • Namespace Manager │  │
                   │  │ • Transport Proxy   │  │
                   │  │ • Virtual Servers   │  │
                   │  │ • Logging & Metrics │  │
                   │  └─────────────────────┘  │
                   └─────────────┬─────────────┘
                                 │
          ┌──────────────────────┼──────────────────────┐
          │                      │                      │
    ┌──────▼──────┐        ┌──────▼──────┐        ┌──────▼──────┐
    │ namespace-1 │        │ namespace-2 │        │ namespace-3 │
    │             │        │             │        │             │
    │┌───────────┐│        │┌───────────┐│        │┌───────────┐│
    ││MCP Server ││        ││MCP Server ││        ││Virtual    ││
    ││     A     ││        ││     C     ││        ││Server A   ││
    │└───────────┘│        │└───────────┘│        │└───────────┘│
    │┌───────────┐│        │┌───────────┐│        │┌───────────┐│
    ││MCP Server ││        ││MCP Server ││        ││Virtual    ││
    ││     B     ││        ││     D     ││        ││Server B   ││
    │└───────────┘│        │└───────────┘│        │└───────────┘│
    └─────────────┘        └─────────────┘        └─────────────┘
```


## Features

### 🔐 **Security & Authentication**
- **Authentication** - Secure authentication (JWT, OAuth2, OIDC) with RBAC
- **API Key Management** - Role-based access control with fine-grained permissions
- **Rate Limiting** - IP-based limiting with Redis backing and memory fallback
- **Content Filtering** - PII detection, regex patterns, and custom filters

### 🏢 **Server & Namespace Management**
- **Dynamic Discovery** - Automatic MCP server discovery and registration
- **Namespaces** - Group servers with isolated namespaces for internal & external usage
- **Health Monitoring** - Server health checks with automated failover and recovery
- **Public Endpoints** - Auto-generated REST APIs for namespace access

### 🔌 **Multi-Protocol Transport**
- **JSON-RPC 2.0** - Standard synchronous RPC over HTTP
- **WebSocket** - Full-duplex bidirectional communication  
- **Server-Sent Events** - Real-time server-to-client streaming
- **Streamable HTTP** - Official MCP protocol implementation
- **STDIO** - Command-line interface bridge

### 🌐 **Service Virtualization**
- **Protocol Support** - REST APIs, GraphQL *(coming soon)*, gRPC *(coming soon)*
- **MCP Integration** - Transform any HTTP service into MCP tools with schema validation
- **Example Integrations** - Internal API docs, microservers, etc.

### 📊 **Logging, Auditing & Metrics**
- **Audit Trails** - Complete request/response logging with security event tracking
- **Performance Metrics** - Real-time monitoring with health checks and alerting
- **External Integration** - AWS CloudWatch, file-based logging, and custom exporters
- **Session Tracking** - Live session management with detailed interaction logs



## 🛠️ Development

### Quick Development Setup

Fast development with backend in Docker and frontend running locally:

```bash
# Terminal 1: Start backend services
make dev

# Terminal 2: Start frontend locally (much faster)
cd apps/frontend
bun install
bun run dev
```

### Essential Commands
```bash
# Development
make dev              # Start backend services (postgres, redis, backend)
make setup            # Complete setup (DB + admin + orgs + namespaces)
make start            # Production-ready local setup with services
make stop             # Stop all services
make clean            # Stop and remove all data
make logs             # View service logs
make help             # Show all available commands

# Database Operations
make migrate          # Run database migrations
make migrate-down     # Rollback migrations
make migrate-status   # Show migration status
make db-shell         # Open PostgreSQL shell

# Testing & Quality
make test             # Run all tests
make lint             # Run linters

# Build & Utilities
make build            # Build containers
make rebuild          # Rebuild and restart containers
make shell            # Open shell in backend container
make bash             # Open bash in backend container
```

### Troubleshooting

**Docker Compose Issues:**
```bash
# The Makefile automatically detects your Docker Compose version
# Check what it's using:
make help  # Will work with either docker-compose or docker compose

# Manual check:
docker compose version    # Modern v2
docker-compose version    # Legacy v1

# If you get "command not found":
# Install Docker Desktop (includes Compose v2) or standalone Compose
```

**Common Issues:**
- **Port conflicts**: Stop other services on ports 8080, 3000, 5432, 6379
- **Permission denied**: Ensure Docker daemon is running
- **Build failures**: Try `make clean` then `make setup`

## 🤝 Contributing

1. Fork the repository
2. Run `make dev` to start backend services
3. Run frontend locally: `cd apps/frontend && bun run dev`
4. Make your changes
5. Run `make test` and `make lint`
6. Submit a Pull Request

Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

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
