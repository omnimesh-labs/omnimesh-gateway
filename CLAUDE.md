# Janex Gateway - Claude Development Guide

## Project Overview

The **Janex Gateway** is a production-ready API gateway for Model Context Protocol (MCP) servers, providing enterprise-grade infrastructure with authentication, logging, rate limiting, server discovery, and multi-protocol transport support.

### Core Purpose
- **Enterprise Infrastructure**: JWT authentication, RBAC, and flexible policies
- **Multi-Protocol Support**: JSON-RPC, WebSocket, SSE, HTTP, and STDIO transports
- **Service Virtualization**: Wrap REST/GraphQL/gRPC services as virtual MCP servers
- **Production Ready**: Comprehensive logging, IP rate limiting with Redis/memory backends, and health monitoring
- **Namespace Management**: Group MCP servers into logical namespaces within an organization
- **Developer Tools**: MCP Inspector for real-time debugging and testing

### Key Features
- 🔐 **Authentication & Authorization** - JWT-based auth with API keys, RBAC, and OAuth2
- 📊 **Comprehensive Logging** - Request/response logging, audit trails, performance metrics
- ⚡ **Rate Limiting** - IP-based rate limiting with sliding window algorithms
- 🛡️ **Smart Rate Limiting** - Redis-backed with memory fallback and proxy detection
- 🔍 **MCP Server Discovery** - Dynamic registration and health checking
- 🌐 **Service Virtualization** - Wrap non-MCP services as virtual MCP servers
- 🔌 **Multi-Protocol Support** - JSON-RPC, WebSocket, SSE, HTTP, and STDIO transports
- 🚀 **High Performance** - Built with Go and Gin for maximum throughput and low latency
- 🏢 **Namespace Management** - Group and organize MCP servers into logical namespaces
- 🔍 **MCP Inspector** - Real-time debugging and testing interface
- 🤝 **Agent-to-Agent (A2A)** - Agent-to-agent authentication and communication
- 🔧 **Plugin System** - Extensible middleware with content filters and AI integrations
- 📡 **Endpoint Management** - Dynamic REST endpoint creation with OpenAPI generation
- 🛠️ **Content Filtering** - PII detection, regex filtering, and resource protection

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
│   │   │   │   ├── rbac.go   # Role-based access control
│   │   │   │   ├── cache.go  # JWT blacklist cache
│   │   │   │   ├── cleanup.go # Token cleanup service
│   │   │   │   └── service.go # Auth service
│   │   │   ├── a2a/          # Agent-to-Agent Auth
│   │   │   │   ├── service.go # A2A service
│   │   │   │   ├── client.go # A2A client
│   │   │   │   └── adapter.go # A2A adapters
│   │   │   ├── config/       # Configuration management
│   │   │   │   ├── config.go # Config structs and loading
│   │   │   │   ├── policy.go # Policy configuration
│   │   │   │   └── validation.go # Config validation
│   │   │   ├── database/     # Database layer
│   │   │   │   ├── database.go # Database connection
│   │   │   │   ├── models/   # Database models
│   │   │   │   │   ├── base.go # Base model
│   │   │   │   │   ├── user.go # User model
│   │   │   │   │   ├── organization.go # Organization model
│   │   │   │   │   ├── namespace.go # Namespace model
│   │   │   │   │   ├── server.go # MCP server model
│   │   │   │   │   ├── session.go # Session model
│   │   │   │   │   ├── virtual_server.go # Virtual server model
│   │   │   │   │   ├── a2a.go   # Agent-to-Agent models
│   │   │   │   │   ├── config.go # Configuration models
│   │   │   │   │   ├── content_filter.go # Content filter models
│   │   │   │   │   ├── logging.go # Logging models
│   │   │   │   │   └── ratelimit.go # Rate limiting models
│   │   │   │   └── repositories/ # Database repositories
│   │   │   │       └── namespace_repo.go # Namespace repository
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
│   │   │   ├── plugins/      # Plugin System
│   │   │   │   ├── manager.go # Plugin manager
│   │   │   │   ├── registry.go # Plugin registry
│   │   │   │   ├── interfaces.go # Plugin interfaces
│   │   │   │   ├── service.go # Plugin service
│   │   │   │   ├── middleware.go # Plugin middleware
│   │   │   │   ├── content_filters/ # Content filtering plugins
│   │   │   │   │   ├── pii/    # PII detection
│   │   │   │   │   ├── regex/  # Regex filtering
│   │   │   │   │   ├── deny/   # Deny lists
│   │   │   │   │   └── resource/ # Resource protection
│   │   │   │   ├── ai_middleware/ # AI integration plugins
│   │   │   │   │   ├── openai_mod/ # OpenAI moderation
│   │   │   │   │   └── llamaguard/ # Llama Guard
│   │   │   │   └── shared/   # Shared plugin utilities
│   │   │   ├── inspector/    # MCP Inspector
│   │   │   │   ├── service.go # Inspector service
│   │   │   │   └── types.go  # Inspector types
│   │   │   ├── services/     # Business Services
│   │   │   │   ├── namespace_service.go # Namespace management
│   │   │   │   ├── namespace_session_pool.go # Session pooling
│   │   │   │   ├── endpoint_service.go # Endpoint management
│   │   │   │   └── openapi_generator.go # OpenAPI generation
│   │   │   ├── server/       # HTTP Server
│   │   │   │   ├── handlers/ # HTTP handlers
│   │   │   │   │   ├── auth.go # Auth endpoints
│   │   │   │   │   ├── oauth_handlers.go # OAuth2 endpoints
│   │   │   │   │   ├── gateway.go # Gateway endpoints
│   │   │   │   │   ├── health.go # Health check endpoints
│   │   │   │   │   ├── admin.go # Admin endpoints
│   │   │   │   │   ├── mcp_discovery.go # MCP discovery endpoints
│   │   │   │   │   ├── namespace_handlers.go # Namespace management
│   │   │   │   │   ├── inspector_handlers.go # MCP Inspector endpoints
│   │   │   │   │   ├── endpoint_*.go # Endpoint management
│   │   │   │   │   ├── filters.go # Content filter endpoints
│   │   │   │   │   ├── resource.go # Resource endpoints
│   │   │   │   │   ├── tool.go # Tool endpoints
│   │   │   │   │   ├── prompt.go # Prompt endpoints
│   │   │   │   │   ├── a2a.go # Agent-to-Agent endpoints
│   │   │   │   │   ├── policy.go # Policy endpoints
│   │   │   │   │   ├── transport_*.go # Transport handlers
│   │   │   │   │   ├── virtual_admin.go # Virtual server admin
│   │   │   │   │   ├── virtual_mcp.go # Virtual server MCP
│   │   │   │   │   └── common.go # Common handler utilities
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
│   │   │   │   ├── a2a.go    # Agent-to-Agent types
│   │   │   │   ├── config.go # Configuration types
│   │   │   │   ├── configuration.go # Runtime config types
│   │   │   │   ├── discovery.go # Discovery types
│   │   │   │   ├── endpoint.go # Endpoint types
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

### Authentication & OAuth2
- `POST /auth/login` - Username/password login
- `POST /auth/refresh` - Token refresh
- `POST /auth/logout` - User logout
- `POST /auth/api-keys` - Create API keys
- `GET /auth/profile` - Get user profile
- `PUT /auth/profile` - Update user profile

#### OAuth2 Endpoints
- `GET /auth/oauth/authorize` - OAuth2 authorization endpoint
- `POST /auth/oauth/token` - OAuth2 token exchange
- `GET /auth/oauth/userinfo` - OAuth2 user information
- `POST /auth/oauth/revoke` - Token revocation
- `GET /auth/oauth/jwks` - JSON Web Key Set

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
- `GET /api/namespaces` - List all namespaces
- `GET /api/namespaces/{id}` - Get namespace details
- `POST /api/namespaces` - Create new namespace
- `PUT /api/namespaces/{id}` - Update namespace
- `DELETE /api/namespaces/{id}` - Delete namespace
- `GET /api/namespaces/{id}/servers` - List servers in namespace
- `GET /api/namespaces/{id}/sessions` - List sessions in namespace

### Virtual Server Management
- `GET /api/admin/virtual-servers` - List virtual servers
- `POST /api/admin/virtual-servers` - Create virtual server
- `PUT /api/admin/virtual-servers/{id}` - Update virtual server
- `DELETE /api/admin/virtual-servers/{id}` - Delete virtual server
- `POST /mcp/rpc` - Virtual MCP JSON-RPC interface

### Content Filtering & Policies
- `GET /api/admin/filters` - List content filters
- `POST /api/admin/filters` - Create content filter
- `PUT /api/admin/filters/{id}` - Update content filter
- `DELETE /api/admin/filters/{id}` - Delete content filter
- `POST /api/admin/filters/{id}/test` - Test content filter

### Agent-to-Agent (A2A)
- `POST /api/a2a/register` - Register A2A client
- `GET /api/a2a/clients` - List A2A clients
- `PUT /api/a2a/clients/{id}` - Update A2A client
- `DELETE /api/a2a/clients/{id}` - Delete A2A client
- `POST /api/a2a/token` - Get A2A access token

### MCP Inspector
- `GET /api/inspector/sessions` - List active sessions
- `GET /api/inspector/sessions/{id}` - Get session details
- `POST /api/inspector/sessions/{id}/message` - Send message to session
- `GET /api/inspector/sessions/{id}/logs` - Get session logs

### Endpoint Management
- `GET /api/admin/endpoints` - List custom endpoints
- `POST /api/admin/endpoints` - Create custom endpoint
- `PUT /api/admin/endpoints/{id}` - Update custom endpoint
- `DELETE /api/admin/endpoints/{id}` - Delete custom endpoint
- `GET /api/admin/endpoints/{id}/openapi` - Get OpenAPI specification

### Resources, Tools & Prompts
- `GET /api/resources` - List resources
- `POST /api/resources` - Create resource
- `GET /api/tools` - List tools
- `POST /api/tools` - Create tool
- `GET /api/prompts` - List prompts
- `POST /api/prompts` - Create prompt

### Admin & Monitoring
- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics
- `GET /admin/logs` - Audit logs
- `GET /admin/stats` - Usage statistics
- `GET /admin/policies` - List policies
- `POST /admin/policies` - Create policy

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
make setup                  # Complete setup (DB + admin + orgs + namespaces)
make start                  # Production-ready local setup with services
make setup-reset            # Reset database (WARNING: deletes all data)
```

### Configuration
- **Environment Variables**: Database connection, JWT secrets, rate limiting defaults
- **YAML Configuration**: Policy definitions, server configurations, feature flags
- **Development Config**: `apps/backend/configs/development.yaml`
- **Production Config**: `apps/backend/configs/production.yaml`

### Testing Strategy

Our testing approach ensures high code quality and reliability across all components:

#### Test Architecture
- **Testcontainers**: Real PostgreSQL and Redis instances for integration tests
- **Docker Environment**: Isolated test environment with `DOCKER_ENV=1`
- **Coverage Reporting**: HTML coverage reports with `make test-coverage`
- **Parallel Execution**: Tests run concurrently for faster feedback

#### Test Categories

##### 1. Unit Tests (`apps/backend/tests/unit/`)
- **Authentication**: JWT validation, token blacklist, RBAC
- **Services**: Namespace management, endpoint services, OAuth
- **Filters**: Content filtering plugins (PII, regex, deny, resource)
- **Repositories**: Database layer operations
- **Utilities**: A2A client/service, helpers

##### 2. Integration Tests (`apps/backend/tests/integration/`)
- **Authentication Flow**: Complete login/logout/refresh cycles
- **Namespace Management**: CRUD operations with database
- **OAuth End-to-End**: Authorization code flow testing
- **Endpoint Management**: Dynamic endpoint creation/deletion
- **Multi-Transport**: All transport protocols working together

##### 3. Transport Tests (`apps/backend/tests/transport/`)
- **JSON-RPC** (`rpc/`): Request/response validation, batch operations
- **WebSocket** (`websocket/`): Connection lifecycle, message handling
- **Server-Sent Events** (`sse/`): Event streaming, client management
- **MCP Protocol** (`mcp/`): Streamable HTTP, capabilities exchange
- **STDIO Bridge** (`stdio/`): Process management, I/O handling

##### 4. Component Tests (within modules)
- **Middleware Tests**: Security headers, rate limiting, CORS
- **Handler Tests**: HTTP endpoint behavior, error handling
- **Service Tests**: Business logic validation
- **Repository Tests**: Database operations and constraints

#### Test Data Management
- **Fixtures**: Consistent test data across test suites
- **Cleanup**: Automatic cleanup after each test
- **Isolation**: Tests don't interfere with each other
- **Seeding**: Proper database seeding for integration tests

#### Mock Strategy
- **External Services**: HTTP clients, third-party APIs
- **Database**: Optional mocking for pure unit tests
- **Transport**: Mock transport layers for service testing
- **Authentication**: Mock JWT validation for handler tests

## Implementation Status

### ✅ Complete Features

#### Core Infrastructure
- ✅ Architecture and scaffolding
- ✅ Database schema with comprehensive migrations
- ✅ Configuration system with validation
- ✅ Type definitions and interfaces
- ✅ Middleware chain with composable patterns
- ✅ Transport layer implementations (JSON-RPC, WebSocket, SSE, MCP, STDIO)
- ✅ Virtual server framework
- ✅ Comprehensive testing infrastructure

#### Authentication & Authorization
- ✅ JWT token management with blacklist cache
- ✅ Role-based access control (RBAC)
- ✅ OAuth2 authorization code flow
- ✅ API key management (user and A2A keys)
- ✅ Agent-to-Agent (A2A) authentication system

#### Advanced Features
- ✅ IP-based rate limiting with Redis sliding window and memory fallback
- ✅ Namespace management with session pooling
- ✅ Plugin system with content filtering
- ✅ MCP Inspector for real-time debugging
- ✅ Dynamic endpoint management with OpenAPI generation
- ✅ Content filtering plugins (PII, regex, deny lists, resource protection)
- ✅ AI middleware integration (OpenAI moderation, Llama Guard)

#### Testing & Quality
- ✅ Extensive unit test coverage
- ✅ Integration tests with testcontainers
- ✅ Transport-specific test suites
- ✅ End-to-end authentication flows
- ✅ Error handling and edge case testing

### 🔄 Areas for Enhancement

While the core system is fully functional, these areas can be expanded:

#### Performance & Scalability
- **Metrics Collection**: Enhanced Prometheus metrics and custom dashboards
- **Caching Layer**: Redis-based response caching for frequently accessed data
- **Connection Pooling**: Advanced database connection management
- **Load Balancing**: Multi-instance deployment support

#### Advanced Features
- **GraphQL Support**: Virtual server adapters for GraphQL endpoints
- **gRPC Bridge**: Protocol translation for gRPC services
- **WebHook Management**: Outbound webhook system for event notifications
- **Batch Operations**: Bulk operations for administrative tasks

#### Monitoring & Observability
- **Distributed Tracing**: OpenTelemetry integration
- **Advanced Logging**: Structured logging with correlation IDs
- **Health Dashboards**: Real-time system health visualization
- **Alert Management**: Automated alerting for system issues

#### Developer Experience
- **SDK Generation**: Auto-generated client SDKs
- **API Documentation**: Enhanced interactive API documentation
- **Development Tools**: CLI tools for gateway management
- **Testing Utilities**: Enhanced testing helpers and fixtures

### Key Dependencies

#### Backend (Go)
- **Framework**: Gin HTTP framework
- **Database**: PostgreSQL with extensions (uuid-ossp, pgcrypto, citext)
- **Cache**: Redis for rate limiting and session management
- **Authentication**: JWT-go, OAuth2, bcrypt for password hashing
- **Rate Limiting**: ulule/limiter with Redis and memory backends
- **Testing**: Testcontainers, testify, Go testing framework
- **Transport**: Gorilla WebSocket, Server-Sent Events
- **Configuration**: Viper, YAML configuration management

#### Frontend (Next.js)
- **Framework**: Next.js 14 with App Router
- **Language**: TypeScript for type safety
- **UI**: React 18, modern component patterns
- **Package Manager**: Bun (preferred over npm)

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
4. **Complete Setup**: `make setup` (creates admin user, orgs, namespaces)
5. **Build & Run**: `make build && make run`
6. **Development**: `make watch` for live reload
7. **Testing**: `make test` for full test suite

### Quick Start for Authentication
```bash
# Start database and apply migrations
make docker-run
make migrate

# Complete setup (creates admin user, orgs, namespaces)
make setup

# Start the backend server
make run

# In another terminal, start frontend
cd apps/frontend && bun run dev
```

**Admin Credentials**: `admin@admin.com` / `qwerty123`

The codebase provides a comprehensive foundation for a production-ready Janex Gateway with enterprise features and multi-protocol support.

## DEVELOPMENT GUIDELINES
- ALWAYS use `bun` instead of `npm`
- Do NOT expose unhandled exception errors in the API response body - its a security vulnerability!
- NO estimates. Don't include effort estimates or "phases"
