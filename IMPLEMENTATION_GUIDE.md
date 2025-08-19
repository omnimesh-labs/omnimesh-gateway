# MCP Gateway Implementation Guide

This guide provides a roadmap for implementing the actual business logic in the scaffolded MCP Gateway architecture.

## Current Status

✅ **Architecture Complete** - All modules, interfaces, and structures are defined
✅ **Database Schema** - Complete PostgreSQL schema with all tables and indexes
✅ **Configuration System** - YAML-based config with environment variable overrides
✅ **Type Definitions** - Comprehensive type system for all components
✅ **Middleware Chain** - Flexible middleware architecture for request processing

## Implementation Priority

### Phase 1: Core Foundation (Week 1-2)

1. **Database Connection & Migrations**
   - Implement database migration tool in `cmd/migrate/`
   - Complete `internal/database/database.go` with connection pooling
   - Add migration runner functionality

2. **Configuration Loading**
   - Complete `internal/config/config.go` implementation
   - Add configuration validation
   - Environment variable parsing

3. **Basic HTTP Server**
   - Update `internal/server/server.go` with proper initialization
   - Implement health check endpoints
   - Basic routing structure

### Phase 2: Authentication & Authorization (Week 2-3)

1. **JWT Implementation**
   - Complete `internal/auth/jwt.go` token generation/validation
   - Implement refresh token logic
   - Token blacklisting for security

2. **User Management**
   - Complete `internal/auth/service.go` user CRUD operations
   - Password hashing (bcrypt)
   - User authentication flow

3. **API Key Management**
   - API key generation and validation
   - Permission mapping
   - Key expiration handling

4. **Policy Engine**
   - Complete `internal/auth/policies.go` policy evaluation
   - Rule condition parsing
   - Policy caching for performance

### Phase 3: Logging & Monitoring (Week 3-4)

1. **Request Logging**
   - Complete `internal/logging/service.go` database persistence
   - Async logging for performance
   - Log rotation and cleanup

2. **Audit Trail**
   - Administrative action logging
   - Change tracking
   - Compliance reporting

3. **Performance Metrics**
   - Response time tracking
   - Error rate monitoring
   - Resource usage metrics

### Phase 4: Rate Limiting (Week 4-5)

1. **Storage Backends**
   - Complete Redis integration in `internal/ratelimit/storage.go`
   - Memory storage implementation
   - Fallback mechanisms

2. **Rate Limiting Algorithms**
   - Complete sliding window algorithm in `internal/ratelimit/limiter.go`
   - Token bucket implementation
   - Fixed window implementation

3. **Policy Management**
   - Rate limit policy CRUD
   - Dynamic policy updates
   - Per-user/org/endpoint limits

### Phase 5: MCP Server Discovery (Week 5-6)

1. **Server Registration**
   - Complete `internal/discovery/service.go` server management
   - Health check implementation
   - Server metadata tracking

2. **Health Monitoring**
   - Complete `internal/discovery/health.go` health checking
   - Failure detection and recovery
   - Circuit breaker pattern

3. **Load Balancing**
   - Round-robin implementation
   - Least connections algorithm
   - Weighted distribution

### Phase 6: Gateway Core (Week 6-7)

1. **HTTP Proxy**
   - Create `internal/gateway/proxy.go` for request forwarding
   - Request/response transformation
   - Error handling and retries

2. **Routing Engine**
   - Create `internal/gateway/router.go` for intelligent routing
   - Path-based routing
   - Header-based routing

3. **Circuit Breaker**
   - Implement circuit breaker pattern
   - Fallback mechanisms
   - Auto-recovery logic

### Phase 7: API Endpoints (Week 7-8)

1. **Authentication Endpoints**
   - Complete `internal/server/handlers/auth.go`
   - Login/logout/refresh endpoints
   - API key management endpoints

2. **Gateway Management**
   - Complete `internal/server/handlers/gateway.go`
   - Server registration/management
   - Policy management endpoints

3. **Admin Interface**
   - Complete `internal/server/handlers/admin.go`
   - User management
   - Analytics and reporting

## Key Implementation Details

### Database Queries

Each service needs SQL queries for CRUD operations. Example for users:

```sql
-- Get user by email
SELECT id, email, name, organization_id, role, is_active, created_at, updated_at 
FROM users 
WHERE email = $1 AND is_active = true

-- Create user
INSERT INTO users (email, name, password_hash, organization_id, role) 
VALUES ($1, $2, $3, $4, $5) 
RETURNING id, created_at

-- Update user
UPDATE users 
SET name = $1, role = $2, updated_at = NOW() 
WHERE id = $3
```

### Error Handling

Use the structured error types defined in `internal/types/errors.go`:

```go
// Example error handling
if user == nil {
    return nil, types.NewNotFoundError("User not found")
}

if !user.IsActive {
    return nil, types.NewUnauthorizedError("User account is disabled")
}
```

### Configuration Usage

Access configuration through the service constructors:

```go
authService := auth.NewService(db, &auth.Config{
    JWTSecret:          config.Auth.JWTSecret,
    AccessTokenExpiry:  config.Auth.AccessTokenExpiry,
    RefreshTokenExpiry: config.Auth.RefreshTokenExpiry,
})
```

### Testing Strategy

1. **Unit Tests** - Test individual functions and methods
2. **Integration Tests** - Test service interactions with database
3. **End-to-End Tests** - Test complete request flows
4. **Load Tests** - Test performance under load

### Security Considerations

1. **Password Security** - Use bcrypt with appropriate cost
2. **JWT Security** - Rotate secrets, use secure algorithms
3. **SQL Injection** - Use parameterized queries
4. **Rate Limiting** - Prevent abuse and DoS attacks
5. **Input Validation** - Validate all user inputs
6. **HTTPS Only** - Force TLS in production

### Performance Optimizations

1. **Database Indexes** - Already defined in migration
2. **Connection Pooling** - Configure appropriate pool sizes
3. **Caching** - Cache policies and user permissions
4. **Async Logging** - Don't block requests for logging
5. **Load Balancing** - Distribute traffic efficiently

## Directory Structure Reference

```
mcp-gateway/
├── cmd/
│   ├── api/main.go           # ✅ Server entrypoint
│   ├── migrate/              # 🔄 Database migrations tool
│   └── worker/               # 🔄 Background workers
├── internal/
│   ├── auth/                 # ✅ Authentication & authorization
│   ├── config/               # ✅ Configuration management
│   ├── database/             # 🔄 Database layer (needs queries)
│   ├── discovery/            # ✅ MCP server discovery
│   ├── gateway/              # 🔄 Core gateway (needs proxy logic)
│   ├── logging/              # ✅ Logging & audit
│   ├── middleware/           # ✅ HTTP middleware
│   ├── ratelimit/            # ✅ Rate limiting
│   ├── server/               # 🔄 HTTP server (needs handlers)
│   └── types/                # ✅ Shared types
├── migrations/               # ✅ Database schema
├── configs/                  # ✅ Configuration files
└── docs/                     # 📚 Documentation
```

Legend:
- ✅ Structure complete
- 🔄 Needs implementation
- 📚 Documentation

## Next Steps

1. **Choose your starting phase** based on priorities
2. **Set up development environment** with PostgreSQL
3. **Run database migrations** to create schema
4. **Implement one module at a time** following the interfaces
5. **Write tests** for each implemented component
6. **Integrate components** as you build them

The scaffolded architecture provides a solid foundation - now it's time to bring it to life with actual business logic!
