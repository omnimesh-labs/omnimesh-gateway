# MCP Gateway

A production-ready API gateway for Model Context Protocol (MCP) servers, providing organization-level policies, authentication, logging, rate limiting, and server discovery.

## Features

- **🔐 Authentication & Authorization** - JWT-based auth with API keys and role-based access control
- **📊 Comprehensive Logging** - Request/response logging, audit trails, performance metrics, and security events
- **⚡ Rate Limiting** - Per-user, per-organization, and per-endpoint rate limiting with multiple algorithms
- **🔍 MCP Server Discovery** - Dynamic server registration, health checking, and load balancing
- **⚙️ Policy Management** - Flexible organization-level policies for access control and routing
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
                   │  │ • Config Management │  │
                   │  └─────────────────────┘  │
                   └─────────────┬─────────────┘
                                 │
          ┌──────────────────────┼──────────────────────┐
          │                      │                      │
┌─────────▼───────┐    ┌─────────▼───────┐    ┌─────────▼───────┐
│  MCP Server A   │    │  MCP Server B   │    │  MCP Server C   │
│  (AI Assistant) │    │  (Code Tools)   │    │  (Data Access)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

See [ARCHITECTURE.md](./ARCHITECTURE.md) for detailed architectural documentation.

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
