# MCP Gateway Architecture

## Overview
This MCP (Model Context Protocol) Gateway provides organization-level policies, authentication, logging, and rate limiting for MCP server access.

## Core Features
1. **Authentication & Authorization** - JWT-based auth with org-level policies
2. **Logging & Audit** - Comprehensive request/response logging and audit trails
3. **Rate Limiting** - Per-user, per-org, and per-endpoint rate limiting
4. **MCP Server Discovery** - Dynamic discovery and health checking of MCP servers
5. **Gateway Configuration** - Flexible policy management and configuration
6. **Proxy & Load Balancing** - Intelligent routing to MCP servers

## Directory Structure

```
mcp-gateway/
├── apps/
│   ├── backend/              # Go API backend
│   │   ├── cmd/
│   │   │   ├── api/          # API server entrypoint
│   │   │   ├── migrate/      # Database migration tool
│   │   │   └── worker/       # Background worker for health checks, cleanup
│   │   ├── internal/
│   │   │   ├── auth/         # Authentication & Authorization
│   │   │   │   ├── jwt.go    # JWT token management
│   │   │   │   ├── middleware.go # Auth middleware
│   │   │   │   ├── policies.go # Policy engine
│   │   │   │   └── service.go # Auth service
│   │   │   ├── config/       # Configuration management
│   │   │   │   ├── config.go # Config structs and loading
│   │   │   │   ├── policy.go # Policy configuration
│   │   │   │   └── validation.go # Config validation
│   │   │   ├── database/     # Database layer
│   │   │   │   ├── database.go # Database connection
│   │   │   │   └── models/   # Database models
│   │   │   ├── discovery/    # MCP Server Discovery
│   │   │   │   ├── health.go # Health checking
│   │   │   │   ├── registry.go # Server registry
│   │   │   │   ├── mcp_discovery.go # MCP discovery service
│   │   │   │   └── service.go # Discovery service
│   │   │   ├── gateway/      # Core gateway functionality
│   │   │   │   ├── proxy.go  # HTTP proxy logic
│   │   │   │   ├── router.go # Request routing
│   │   │   │   └── loadbalancer.go # Load balancing
│   │   │   ├── logging/      # Logging & Audit
│   │   │   │   ├── audit.go  # Audit logging
│   │   │   │   ├── middleware.go # Request logging middleware
│   │   │   │   └── service.go # Logging service
│   │   │   ├── middleware/   # HTTP Middleware
│   │   │   │   ├── chain.go  # Middleware chain builder
│   │   │   │   ├── cors.go   # CORS middleware
│   │   │   │   ├── recovery.go # Panic recovery
│   │   │   │   ├── timeout.go # Request timeout
│   │   │   │   └── path_rewrite.go # Path rewriting for server-specific endpoints
│   │   │   ├── ratelimit/    # Rate Limiting
│   │   │   │   ├── limiter.go # Rate limiter implementation
│   │   │   │   ├── middleware.go # Rate limiting middleware
│   │   │   │   ├── storage.go # Rate limit storage (Redis/Memory)
│   │   │   │   ├── service.go # Rate limiting service
│   │   │   │   └── policies.go # Rate limiting policies
│   │   │   ├── server/       # HTTP Server
│   │   │   │   ├── handlers/ # HTTP handlers
│   │   │   │   │   ├── auth.go # Auth endpoints
│   │   │   │   │   ├── gateway.go # Gateway endpoints
│   │   │   │   │   ├── health.go # Health check endpoints
│   │   │   │   │   ├── admin.go # Admin endpoints
│   │   │   │   │   ├── mcp_discovery.go # MCP discovery endpoints
│   │   │   │   │   ├── transport_rpc.go # JSON-RPC transport handler
│   │   │   │   │   ├── transport_sse.go # SSE transport handler
│   │   │   │   │   ├── transport_ws.go # WebSocket transport handler
│   │   │   │   │   ├── transport_mcp.go # Streamable HTTP transport handler
│   │   │   │   │   └── transport_stdio.go # STDIO transport handler
│   │   │   │   ├── routes.go # Route definitions
│   │   │   │   └── server.go # Server setup
│   │   │   ├── transport/    # Transport Layer (NEW)
│   │   │   │   ├── base.go   # Transport interface and base implementation
│   │   │   │   ├── manager.go # Transport manager/multiplexer
│   │   │   │   ├── session.go # Session management for stateful transports
│   │   │   │   ├── jsonrpc.go # JSON-RPC over HTTP transport
│   │   │   │   ├── sse.go     # Server-Sent Events transport
│   │   │   │   ├── websocket.go # WebSocket transport
│   │   │   │   ├── streamable.go # Streamable HTTP transport
│   │   │   │   └── stdio.go  # STDIO transport bridge
│   │   │   └── types/        # Shared types and interfaces
│   │   │       ├── auth.go   # Auth-related types
│   │   │       ├── config.go # Configuration types
│   │   │       ├── discovery.go # Discovery types
│   │   │       ├── errors.go # Custom error types
│   │   │       ├── gateway.go # Gateway types
│   │   │       ├── mcp.go    # MCP protocol types
│   │   │       └── transport.go # Transport-related types
│   │   ├── migrations/       # Database migrations
│   │   └── configs/          # Configuration files
│   │       ├── development.yaml
│   │       ├── production.yaml
│   │       └── test.yaml
│   └── frontend/             # Next.js frontend dashboard
│       ├── src/
│       │   └── app/          # Next.js App Router
│       ├── package.json
│       ├── next.config.js
│       └── tsconfig.json
├── pkg/                      # Shared Go packages
│   ├── client/               # MCP client library
│   ├── protocol/             # MCP protocol definitions
│   └── utils/                # Utility functions
├── docs/                     # Documentation
│   ├── api/                  # API documentation
│   └── deployment/           # Deployment guides
├── scripts/                  # Build and deployment scripts
├── go.mod                    # Go module definition
├── go.sum                    # Go dependencies
├── Makefile                  # Build commands
├── docker-compose.yml        # Docker services
└── .air.toml                 # Hot reload configuration
```

## Key Components

### 1. Authentication & Authorization (`apps/backend/internal/auth/`)
- JWT-based authentication
- Organization-level policies
- Role-based access control (RBAC)
- API key management

### 2. Logging & Audit (`apps/backend/internal/logging/`)
- Request/response logging
- Audit trail for policy changes
- Performance metrics
- Error tracking

### 3. Rate Limiting (`apps/backend/internal/ratelimit/`)
- Per-user rate limits
- Per-organization rate limits
- Per-endpoint rate limits
- Sliding window implementation

### 4. MCP Server Discovery (`apps/backend/internal/discovery/`)
- Service registry
- Health checking
- Auto-discovery mechanisms
- Load balancing strategies

### 5. Gateway Configuration (`apps/backend/internal/config/`)
- Policy management
- Dynamic configuration updates
- Environment-specific configs
- Validation and defaults

### 6. Gateway Core (`apps/backend/internal/gateway/`)
- HTTP proxy functionality
- Request routing
- Load balancing
- Circuit breaker patterns

### 7. Transport Layer (`apps/backend/internal/transport/`)
- Multi-protocol MCP transport support
- Session management for stateful transports
- Transport manager/multiplexer
- Protocol-specific implementations:
  - **JSON-RPC over HTTP**: Standard synchronous RPC
  - **Server-Sent Events (SSE)**: Real-time server-to-client streaming
  - **WebSocket**: Full-duplex bidirectional communication
  - **Streamable HTTP**: Official MCP protocol with session support
  - **STDIO**: Command-line interface bridge

### 8. Frontend Dashboard (`apps/frontend/`)
- Next.js TypeScript application
- MCP Gateway management interface
- Real-time monitoring dashboard
- Configuration management UI

## Database Schema Design

### Core Tables
- `organizations` - Organization metadata
- `users` - User accounts and roles
- `api_keys` - API key management
- `mcp_servers` - Registered MCP servers
- `policies` - Organization policies
- `rate_limits` - Rate limiting configurations
- `audit_logs` - Audit trail
- `request_logs` - Request/response logs

## API Endpoints

### Authentication
- `POST /auth/login` - User login
- `POST /auth/refresh` - Token refresh
- `POST /auth/logout` - User logout
- `POST /auth/api-keys` - Generate API key

### Gateway Management
- `GET /gateway/servers` - List available MCP servers
- `POST /gateway/servers` - Register new MCP server
- `DELETE /gateway/servers/{id}` - Unregister MCP server

### Policy Management
- `GET /policies` - List organization policies
- `POST /policies` - Create policy
- `PUT /policies/{id}` - Update policy
- `DELETE /policies/{id}` - Delete policy

### Proxy Endpoints
- `/*` - Proxy requests to MCP servers (with auth/rate limiting)

### Transport Endpoints
#### JSON-RPC over HTTP
- `POST /rpc` - JSON-RPC requests
- `POST /rpc/batch` - Batch JSON-RPC requests
- `GET /rpc/introspection` - Available RPC methods

#### Server-Sent Events (SSE)
- `GET /sse` - SSE connection
- `POST /sse/events` - Send SSE events
- `POST /sse/broadcast` - Broadcast to all SSE clients
- `GET /sse/status` - SSE connection status
- `GET /sse/replay/{session_id}` - Replay events from specific point

#### WebSocket
- `GET /ws` - WebSocket connection upgrade
- `POST /ws/send` - Send message to WebSocket
- `POST /ws/broadcast` - Broadcast to all WebSocket clients
- `GET /ws/status` - WebSocket connection status
- `POST /ws/ping` - Send ping to WebSocket
- `DELETE /ws/close` - Close WebSocket connection

#### Streamable HTTP (MCP Protocol)
- `GET|POST /mcp` - Streamable HTTP endpoints
- `GET /mcp/capabilities` - MCP capabilities
- `GET /mcp/status` - MCP connection status

#### STDIO Bridge
- `POST /stdio/execute` - Execute command via STDIO
- `GET|POST /stdio/process` - Manage STDIO processes
- `POST /stdio/send` - Send message to STDIO process

#### Server-Specific Endpoints
- `POST /servers/{server_id}/rpc` - Server-specific JSON-RPC
- `GET /servers/{server_id}/sse` - Server-specific SSE
- `GET /servers/{server_id}/ws` - Server-specific WebSocket
- `GET|POST /servers/{server_id}/mcp` - Server-specific MCP

### Admin & Monitoring
- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics
- `GET /admin/logs` - Audit logs
- `GET /admin/stats` - Usage statistics

## Configuration Management

### Environment Variables
- Database connection settings
- JWT secrets
- Rate limiting defaults
- Discovery service settings

### YAML Configuration Files
- Policy definitions
- Server configurations
- Feature flags
- Environment-specific overrides
- Transport layer settings

### Transport Configuration
- Enabled transport protocols
- Session management settings
- Connection limits and timeouts
- Path rewriting rules
- Protocol-specific configurations

## Security Considerations
- JWT token validation
- Request sanitization
- Rate limiting bypass protection
- Audit logging for all admin actions
- Secure header handling
- TLS termination

## Scalability Features
- Horizontal scaling support
- Redis for distributed rate limiting
- Database connection pooling
- Caching strategies
- Health check mechanisms

## Monitoring & Observability
- Prometheus metrics
- Structured logging
- Distributed tracing
- Performance profiling
- Error rate monitoring
