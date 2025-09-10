# Omnimesh Gateway Architecture

## Overview
The Omnimesh AI Gateway provides organization-level policies, authentication, logging, and rate limiting for MCP server access.

## Core Features
1. **Authentication & Authorization** - JWT-based auth with RBAC and flexible policies
2. **Logging & Audit** - Comprehensive request/response logging and audit trails
3. **Rate Limiting** - IP-based rate limiting with Redis backing and memory fallback
4. **MCP Server Discovery** - Dynamic discovery and health checking of MCP servers
5. **Gateway Configuration** - Flexible policy management and configuration
6. **Namespace Management** - Organize MCP servers into logical namespaces
7. **MCP Inspector** - Real-time debugging and testing interface

## Directory Structure

```
omnimesh-gateway/
├── apps/
│   ├── backend/              # Go API backend
│   │   ├── cmd/
│   │   │   ├── api/          # API server entrypoint (main.go)
│   │   │   ├── migrate/      # Database migration tool
│   │   │   └── worker/       # Background worker for health checks
│   │   ├── internal/         # Core business logic modules
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
│   │   │   │   └── service.go # Gateway service logic
│   │   │   ├── logging/      # Logging & Audit
│   │   │   │   ├── audit.go  # Audit logging
│   │   │   │   ├── interfaces.go # Logging interfaces
│   │   │   │   ├── middleware.go # Request logging middleware
│   │   │   │   ├── service.go # Logging service
│   │   │   │   ├── registry.go # Logging plugin registry
│   │   │   │   └── plugins/  # Logging plugins (AWS, File)
│   │   │   ├── middleware/   # HTTP Middleware
│   │   │   │   ├── chain.go  # Middleware chain builder
│   │   │   │   ├── cors.go   # CORS middleware
│   │   │   │   ├── recovery.go # Panic recovery
│   │   │   │   ├── timeout.go # Request timeout
│   │   │   │   ├── path_rewrite.go # Path rewriting
│   │   │   │   ├── security.go # Security headers middleware
│   │   │   │   ├── iratelimit.go # IP-based rate limiting with Redis/Memory backends
│   │   │   ├── server/       # HTTP Server
│   │   │   │   ├── handlers/ # HTTP handlers
│   │   │   │   │   ├── auth.go # Auth endpoints
│   │   │   │   │   ├── gateway.go # Gateway endpoints
│   │   │   │   │   ├── health.go # Health check endpoints
│   │   │   │   │   ├── admin.go # Admin endpoints
│   │   │   │   │   ├── mcp_discovery.go # MCP discovery endpoints
│   │   │   │   │   ├── transport_*.go # Transport handlers
│   │   │   │   │   ├── virtual_admin.go # Virtual server admin
│   │   │   │   │   └── virtual_mcp.go # Virtual server MCP
│   │   │   │   ├── routes.go # Route definitions
│   │   │   │   └── server.go # Server setup
│   │   │   ├── transport/    # Transport Layer
│   │   │   │   ├── base.go   # Transport interface
│   │   │   │   ├── manager.go # Transport manager/multiplexer
│   │   │   │   ├── session.go # Session management
│   │   │   │   ├── jsonrpc.go # JSON-RPC over HTTP
│   │   │   │   ├── sse.go     # Server-Sent Events
│   │   │   │   ├── websocket.go # WebSocket
│   │   │   │   ├── streamable.go # Streamable HTTP
│   │   │   │   └── stdio.go  # STDIO transport bridge
│   │   │   ├── types/        # Shared types and interfaces
│   │   │   │   ├── auth.go   # Auth-related types
│   │   │   │   ├── config.go # Configuration types
│   │   │   │   ├── discovery.go # Discovery types
│   │   │   │   ├── errors.go # Custom error types
│   │   │   │   ├── gateway.go # Gateway types
│   │   │   │   ├── mcp.go    # MCP protocol types
│   │   │   │   ├── transport.go # Transport types
│   │   │   │   └── virtual.go # Virtual server types
│   │   │   └── virtual/      # Service Virtualization
│   │   │       ├── adapter.go # Virtual server adapters
│   │   │       ├── server.go # Virtual server implementation
│   │   │       └── service.go # Virtual server service
│   │   ├── migrations/       # Database migrations
│   │   ├── configs/          # Configuration files
│   │   └── tests/            # Test suites
│   │       ├── helpers/      # Test helpers
│   │       ├── integration/  # Integration tests
│   │       ├── transport/    # Transport-specific tests
│   │       └── unit/         # Unit tests
│   └── frontend/             # Next.js dashboard
│       ├── src/
│       │   ├── app/          # Next.js App Router
│       │   ├── components/   # React components
│       │   └── lib/          # Frontend utilities
│       ├── package.json
│       └── tsconfig.json
├── pkg/                      # Shared Go packages
│   ├── client/               # MCP client library
│   ├── protocol/             # MCP protocol definitions
│   └── utils/                # Utility functions
├── docs/                     # Documentation
├── examples/                 # Usage examples
├── scripts/                  # Build and deployment scripts
├── go.mod                    # Go module definition
├── Makefile                  # Build commands
└── docker-compose.yml        # Docker services
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
- Omnimesh Gateway management interface
- Real-time monitoring dashboard
- Configuration management UI

## Database Schema Design

### Core Tables
- `namespaces` - Logical grouping of MCP servers
- `organizations` - Organization metadata  
- `users` - User accounts and roles
- `api_keys` - API key management
- `mcp_servers` - Registered MCP servers (namespace-scoped)
- `policies` - Organization policies
- `rate_limits` - Rate limiting configurations
- `audit_logs` - Audit trail
- `request_logs` - Request/response logs
- `virtual_servers` - Service virtualization configurations (namespace-scoped)
- `mcp_sessions` - Active MCP sessions (namespace-scoped)

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

### Namespace Management
- `GET /api/namespaces` - List all namespaces
- `GET /api/namespaces/{id}` - Get namespace details
- `POST /api/namespaces` - Create new namespace
- `PUT /api/namespaces/{id}` - Update namespace
- `DELETE /api/namespaces/{id}` - Delete namespace
- `GET /api/namespaces/{id}/servers` - List servers in namespace
- `GET /api/namespaces/{id}/sessions` - List sessions in namespace

### MCP Inspector
- `POST /api/inspector/sessions` - Create inspector session
- `GET /api/inspector/sessions/{id}` - Get inspector session details
- `DELETE /api/inspector/sessions/{id}` - Close inspector session
- `POST /api/inspector/sessions/{id}/request` - Execute request in session
- `GET /api/inspector/sessions/{id}/events` - Stream events (SSE)
- `GET /api/inspector/sessions/{id}/ws` - WebSocket connection
- `GET /api/inspector/servers/{id}/capabilities` - Get server capabilities

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
