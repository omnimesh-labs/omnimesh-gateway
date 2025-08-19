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
├── cmd/
│   ├── api/                    # API server entrypoint
│   ├── migrate/               # Database migration tool
│   └── worker/                # Background worker for health checks, cleanup
├── internal/
│   ├── auth/                  # Authentication & Authorization
│   │   ├── jwt.go            # JWT token management
│   │   ├── middleware.go     # Auth middleware
│   │   ├── policies.go       # Policy engine
│   │   └── service.go        # Auth service
│   ├── config/               # Configuration management
│   │   ├── config.go         # Config structs and loading
│   │   ├── policy.go         # Policy configuration
│   │   └── validation.go     # Config validation
│   ├── database/             # Database layer (existing)
│   │   ├── migrations/       # SQL migrations
│   │   └── models/           # Database models
│   ├── discovery/            # MCP Server Discovery
│   │   ├── health.go         # Health checking
│   │   ├── registry.go       # Server registry
│   │   └── service.go        # Discovery service
│   ├── gateway/              # Core gateway functionality
│   │   ├── proxy.go          # HTTP proxy logic
│   │   ├── router.go         # Request routing
│   │   └── loadbalancer.go   # Load balancing
│   ├── logging/              # Logging & Audit
│   │   ├── audit.go          # Audit logging
│   │   ├── middleware.go     # Request logging middleware
│   │   └── service.go        # Logging service
│   ├── middleware/           # HTTP Middleware
│   │   ├── chain.go          # Middleware chain builder
│   │   ├── cors.go           # CORS middleware
│   │   ├── recovery.go       # Panic recovery
│   │   └── timeout.go        # Request timeout
│   ├── ratelimit/            # Rate Limiting
│   │   ├── limiter.go        # Rate limiter implementation
│   │   ├── middleware.go     # Rate limiting middleware
│   │   ├── storage.go        # Rate limit storage (Redis/Memory)
│   │   └── policies.go       # Rate limiting policies
│   ├── server/               # HTTP Server (existing, enhanced)
│   │   ├── handlers/         # HTTP handlers
│   │   │   ├── auth.go       # Auth endpoints
│   │   │   ├── gateway.go    # Gateway endpoints
│   │   │   ├── health.go     # Health check endpoints
│   │   │   └── admin.go      # Admin endpoints
│   │   ├── routes.go         # Route definitions
│   │   └── server.go         # Server setup
│   └── types/                # Shared types and interfaces
│       ├── auth.go           # Auth-related types
│       ├── config.go         # Configuration types
│       ├── errors.go         # Custom error types
│       ├── gateway.go        # Gateway types
│       └── mcp.go            # MCP protocol types
├── pkg/                      # Public packages
│   ├── client/               # MCP client library
│   ├── protocol/             # MCP protocol definitions
│   └── utils/                # Utility functions
├── migrations/               # Database migrations
├── configs/                  # Configuration files
│   ├── development.yaml
│   ├── production.yaml
│   └── test.yaml
├── scripts/                  # Build and deployment scripts
└── docs/                     # Documentation
    ├── api/                  # API documentation
    └── deployment/           # Deployment guides
```

## Key Components

### 1. Authentication & Authorization (`internal/auth/`)
- JWT-based authentication
- Organization-level policies
- Role-based access control (RBAC)
- API key management

### 2. Logging & Audit (`internal/logging/`)
- Request/response logging
- Audit trail for policy changes
- Performance metrics
- Error tracking

### 3. Rate Limiting (`internal/ratelimit/`)
- Per-user rate limits
- Per-organization rate limits
- Per-endpoint rate limits
- Sliding window implementation

### 4. MCP Server Discovery (`internal/discovery/`)
- Service registry
- Health checking
- Auto-discovery mechanisms
- Load balancing strategies

### 5. Gateway Configuration (`internal/config/`)
- Policy management
- Dynamic configuration updates
- Environment-specific configs
- Validation and defaults

### 6. Gateway Core (`internal/gateway/`)
- HTTP proxy functionality
- Request routing
- Load balancing
- Circuit breaker patterns

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
