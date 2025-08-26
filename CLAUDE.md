# MCP Gateway - Claude Development Guide

## Project Overview

The **MCP Gateway** is a production-ready API gateway for Model Context Protocol (MCP) servers, providing enterprise-grade infrastructure with authentication, logging, rate limiting, server discovery, and multi-protocol transport support.

### Core Purpose
- **Enterprise Infrastructure**: Organization-level policies, JWT authentication, RBAC, and A2A authentication
- **Multi-Protocol Support**: JSON-RPC, WebSocket, SSE, HTTP, and STDIO transports
- **Service Virtualization**: Wrap REST/GraphQL/gRPC services as virtual MCP servers
- **Production Ready**: Comprehensive logging, IP rate limiting with Redis/memory backends, and health monitoring
- **Multi-tenancy**: Namespace-based isolation for resources and configurations

### Key Features
- 🔐 **Authentication & Authorization** - JWT-based auth with org-level policies, API keys, and A2A authentication
- 📊 **Comprehensive Logging** - Request/response logging, audit trails, performance metrics
- ⚡ **Rate Limiting** - Multi-level rate limiting (user, org, endpoint) with sliding window algorithms
- 🛡️ **IP Rate Limiting** - Redis-backed sliding window or in-memory per-IP rate limiting with smart proxy detection
- 🔍 **MCP Server Discovery** - Dynamic registration and health checking
- 🌐 **Service Virtualization** - Wrap non-MCP services as virtual MCP servers
- 🔌 **Multi-Protocol Support** - JSON-RPC, WebSocket, SSE, HTTP, and STDIO transports
- 🚀 **High Performance** - Built with Go and Gin for maximum throughput and low latency
- 🏢 **Multi-tenant Namespaces** - Resource isolation and management through namespace scoping

## Architecture

### Technology Stack
- **Backend**: Go 1.25 with Gin framework
- **Database**: PostgreSQL with comprehensive migration system
- **Cache**: Redis (optional, falls back to in-memory when disabled)
- **Frontend**: Next.js 14 TypeScript dashboard
- **Testing**: Extensive test suites for all transport layers

### Architectural Decisions
- **Rate Limiting**: Uses industry-standard `ulule/limiter` library instead of custom implementations
- **Redis Integration**: Sliding window rate limiting for distributed deployments with memory fallback
- **Middleware Pattern**: Composable middleware chain for cross-cutting concerns
- **Clean Architecture**: Separation of concerns with internal packages for different domains

### Project Structure
```
mcp-gateway/
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

## Database Schema

### Architecture Pattern
- **Control-plane (Postgres)**: Small, fast queries for operations and metadata
- **Data-plane (Object Storage)**: Bulk log data stored in S3/GCS/CloudWatch/Loki with pointers in Postgres
- **Query Path**: UI queries Postgres index → fetch full logs from object storage on demand

### Core Tables

#### Namespaces
```sql
CREATE TABLE namespaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug CITEXT UNIQUE NOT NULL,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(organization_id, slug)
);
```

#### Organizations
```sql
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug CITEXT UNIQUE NOT NULL,
    plan_type plan_type_enum DEFAULT 'free',
    max_servers INTEGER DEFAULT 10,
    max_sessions INTEGER DEFAULT 100,
    log_retention_days INTEGER DEFAULT 7,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT true
);
```

#### MCP Servers
```sql
CREATE TABLE mcp_servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    namespace_id UUID NOT NULL REFERENCES namespaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    protocol protocol_enum NOT NULL,
    url VARCHAR(500),
    command VARCHAR(255),
    args TEXT[],
    environment TEXT[],
    working_dir VARCHAR(500),
    version VARCHAR(50),
    timeout_seconds INTEGER DEFAULT 300,
    max_retries INTEGER DEFAULT 3,
    status server_status_enum DEFAULT 'active',
    health_check_url VARCHAR(500),
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    tags TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(namespace_id, name)
);
```

#### Virtual Servers
```sql
CREATE TABLE virtual_servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    namespace_id UUID NOT NULL REFERENCES namespaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    adapter_type adapter_type_enum NOT NULL DEFAULT 'REST',
    tools JSONB NOT NULL DEFAULT '[]',
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(namespace_id, name)
);
```

#### Sessions
```sql
CREATE TABLE mcp_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    namespace_id UUID NOT NULL REFERENCES namespaces(id) ON DELETE CASCADE,
    server_id UUID NOT NULL REFERENCES mcp_servers(id) ON DELETE CASCADE,
    status session_status_enum DEFAULT 'initializing',
    protocol protocol_enum NOT NULL,
    client_id UUID,
    connection_id UUID,
    process_pid INTEGER,
    process_status proc_status_enum,
    process_exit_code INTEGER,
    process_error TEXT,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    user_id VARCHAR(255) DEFAULT 'default-user',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### Users & Authentication
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email CITEXT UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'user', 'viewer', 'api_user')),
    is_active BOOLEAN DEFAULT true,
    email_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### API Keys
```sql
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,  -- Nullable for A2A keys
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    namespace_id UUID REFERENCES namespaces(id) ON DELETE CASCADE,  -- Optional namespace scope
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    prefix VARCHAR(20) NOT NULL,
    key_type VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (key_type IN ('user', 'a2a')),
    permissions TEXT[] DEFAULT ARRAY[]::TEXT[],
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### Virtual Servers
```sql
CREATE TABLE virtual_servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    namespace_id UUID NOT NULL REFERENCES namespaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    adapter_type adapter_type_enum NOT NULL DEFAULT 'REST',
    tools JSONB NOT NULL DEFAULT '[]',
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(namespace_id, name)
);
```

#### Logging & Audit
- **log_index**: Fast queries with pointers to detailed logs in object storage
- **audit_logs**: High-level audit trail for administrative actions
- **log_aggregates**: Hourly/daily rollups for fast dashboard queries

#### Rate Limiting
- **rate_limits**: Configure rate limiting rules
- **rate_limit_usage**: Track rate limit usage (Redis-backed in production)

#### Health & Monitoring
- **health_checks**: Track server health status over time
- **server_stats**: Aggregate statistics for monitoring

## API Endpoints

### Authentication
- `POST /auth/login` - Username/password login
- `POST /auth/refresh` - Token refresh
- `POST /auth/logout` - User logout
- `POST /auth/api-keys` - Create API keys
- `GET /auth/profile` - Get user profile
- `PUT /auth/profile` - Update user profile
- `POST /auth/a2a/token` - App-to-app authentication token exchange

### Gateway Management
- `GET /gateway/servers` - List available MCP servers
- `POST /gateway/servers` - Register new MCP server
- `DELETE /gateway/servers/{id}` - Unregister MCP server

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

#### WebSocket
- `GET /ws` - WebSocket connection upgrade
- `POST /ws/send` - Send message to WebSocket
- `POST /ws/broadcast` - Broadcast to all WebSocket clients
- `GET /ws/status` - WebSocket connection status

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
- `GET /api/admin/namespaces` - List all namespaces
- `GET /api/admin/namespaces/{id}` - Get namespace details
- `POST /api/admin/namespaces` - Create new namespace
- `PUT /api/admin/namespaces/{id}` - Update namespace
- `DELETE /api/admin/namespaces/{id}` - Delete namespace
- `GET /api/admin/namespaces/{id}/servers` - List servers in namespace
- `GET /api/admin/namespaces/{id}/sessions` - List sessions in namespace

### Virtual Server Management
- `GET /api/admin/virtual-servers` - List virtual servers
- `POST /api/admin/virtual-servers` - Create virtual server
- `PUT /api/admin/virtual-servers/{id}` - Update virtual server
- `DELETE /api/admin/virtual-servers/{id}` - Delete virtual server
- `POST /mcp/rpc` - Virtual MCP JSON-RPC interface

### Admin & Monitoring
- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics
- `GET /admin/logs` - Audit logs
- `GET /admin/stats` - Usage statistics

## Service Virtualization

The standout feature that allows wrapping non-MCP services as MCP-compatible servers:

### Supported Protocols
- **REST APIs** - Convert HTTP endpoints into MCP tools with automatic parameter mapping
- **GraphQL** - Expose GraphQL queries and mutations as MCP tools *(coming soon)*
- **gRPC** - Bridge gRPC services to MCP protocol *(coming soon)*
- **SOAP** - Legacy SOAP web services support *(coming soon)*

### Key Features
- **JSON-RPC 2.0 Interface** - Standard MCP protocol support via `POST /mcp/rpc`
- **Admin REST API** - Full CRUD operations for virtual server management
- **Database Persistence** - Virtual servers stored in PostgreSQL with in-memory caching
- **Mock Responses** - Built-in testing with mock data for development
- **Error Handling** - Proper JSON-RPC error code mapping (-32601, -32602, -32000)
- **Tool Configuration** - Flexible tool definitions with schema validation

### Example Use Cases
- **Slack Integration** - Expose Slack's REST API as MCP tools for sending messages
- **GitHub Operations** - Wrap GitHub API for repository management, issue creation
- **Database Access** - Convert database queries into MCP tools with proper authentication
- **Third-party SaaS** - Integrate any REST-based service (Stripe, Twilio, etc.)

## Development

### Build Commands (Makefile)
```bash
# Build and test
make all                    # Build + test
make build                  # Build application
make run                    # Run application

# Database
make docker-run             # Start DB container
make docker-down            # Stop DB container
make migrate                # Run migrations
make migrate-down           # Rollback migrations

# Testing
make test                   # Run all tests
make test-transport         # Transport layer tests
make test-integration       # Integration tests
make test-unit              # Unit tests
make test-coverage          # Test with coverage

# Specific transport tests
make test-rpc               # JSON-RPC tests
make test-sse               # SSE tests
make test-websocket         # WebSocket tests
make test-mcp               # MCP tests
make test-stdio             # STDIO tests

# Development
make watch                  # Live reload with air
make clean                  # Clean build artifacts

# Local Setup
make setup                  # Interactive setup menu
make setup-admin            # Create admin user (admin@admin.com)
make setup-reset            # Reset database (WARNING: deletes all data)
```

### Configuration
- **Environment Variables**: Database connection, JWT secrets, rate limiting defaults
- **YAML Configuration**: Policy definitions, server configurations, feature flags
- **Development Config**: `apps/backend/configs/development.yaml`
- **Production Config**: `apps/backend/configs/production.yaml`

### Testing
- **Unit Tests**: Individual functions and methods
- **Integration Tests**: Service interactions with database
- **Transport Tests**: All transport layer implementations
- **End-to-End Tests**: Complete request flows

## Implementation Status

### ✅ Complete
- Architecture and scaffolding
- Database schema with migrations
- Configuration system
- Type definitions
- Middleware chain
- Transport layer interfaces
- Virtual server framework
- Testing infrastructure
- IP-based rate limiting with Redis sliding window and memory fallback

### 🔄 Implementation Needed
Following the **IMPLEMENTATION_GUIDE.md**, the main areas needing business logic:

#### Phase 1: Core Foundation
1. Database connection & migration tool implementation
2. Configuration loading completion
3. Basic HTTP server setup

#### Phase 2: Authentication & Authorization
1. JWT implementation (token generation/validation)
2. User management (CRUD operations, password hashing)
3. API key management
4. Policy engine completion

#### Phase 3: Logging & Monitoring
1. Request logging with database persistence
2. Audit trail implementation
3. Performance metrics tracking

#### Phase 4: Rate Limiting ✅ IP Rate Limiting Complete
1. ✅ IP-based rate limiting with Redis sliding window and memory fallback (using ulule/limiter)
2. TODO: User/organization-level rate limiting (requires auth implementation)
3. TODO: Per-endpoint rate limiting policies
4. TODO: Rate limiting management API

#### Phase 5: MCP Server Discovery
1. Server registration and management
2. Health monitoring implementation

#### Phase 6: API Endpoints
1. Authentication endpoints completion
2. Gateway management endpoints
3. Admin interface implementation

### Key Dependencies
- **Database**: PostgreSQL with extensions (uuid-ossp, pgcrypto, citext)
- **Go Modules**: Gin, JWT, PostgreSQL drivers, Redis client, ulule/limiter (rate limiting), testcontainers
- **Frontend**: Next.js 14, React 18, TypeScript

### Security Considerations
- JWT token validation and rotation
- Password security with bcrypt
- SQL injection prevention with parameterized queries
- Rate limiting for DoS protection
- Input validation and sanitization
- HTTPS enforcement in production

## Getting Started

1. **Prerequisites**: Go 1.25+, PostgreSQL, Docker (optional)
2. **Environment Setup**: Create `.env` file with database credentials
3. **Database Setup**: `make docker-run` and `make migrate`
4. **Create Admin User**: `make setup-admin` (creates admin@admin.com / qwerty123)
5. **Build & Run**: `make build && make run`
6. **Development**: `make watch` for live reload
7. **Testing**: `make test` for full test suite

### Quick Start for Authentication
```bash
# Start database and apply migrations
make docker-run
make migrate

# Create admin user for testing
make setup-admin

# Start the backend server
make run

# In another terminal, start frontend
cd apps/frontend && bun run dev
```

**Admin Credentials**: `admin@admin.com` / `qwerty123`

The codebase provides a comprehensive foundation for a production-ready MCP Gateway with enterprise features and multi-protocol support.

## DEVELOPMENT GUIDELINES
- ALWAYS use `bun` instead of `npm`
- Do NOT expose unhandled exception errors in the API response body - its a security vulnerability!
- NO estimates. Don't include effort estimates or "phases"
