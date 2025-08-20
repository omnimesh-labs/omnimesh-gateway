# MCP Gateway Database ERD

## Overview
This document outlines the database schema for the MCP Gateway. The design separates control-plane data (kept in Postgres) from data-plane logs (stored in object storage/log systems) for scalability and cost efficiency.

## Architecture Pattern
- **Control-plane (Postgres)**: Small, fast queries for operations and metadata
- **Data-plane (Object Storage)**: Bulk log data stored in S3/GCS/CloudWatch/Loki with pointers in Postgres
- **Query Path**: UI queries Postgres index → fetch full logs from object storage on demand

## Core Features Supported
1. **MCP Server Management** - Register and manage MCP servers
2. **Session Management** - Track active MCP sessions and processes  
3. **Log Index & Aggregates** - Fast queries with pointers to detailed logs in object storage
4. **Health Monitoring** - Track server health and performance
5. **Rate Limiting** - Per-server and per-user rate limiting
6. **Multi-tenant Ready** - Organization structure with tenant isolation

## Database Setup

### Required Extensions
```sql
-- Required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";    -- For UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";     -- For gen_random_uuid (secure UUIDs)
CREATE EXTENSION IF NOT EXISTS "citext";       -- Case-insensitive text
```

## Database Enums

```sql
-- Enums for type safety and performance
CREATE TYPE protocol_enum AS ENUM ('stdio','http','https','websocket','sse');
CREATE TYPE server_status_enum AS ENUM ('active','inactive','unhealthy','maintenance');
CREATE TYPE session_status_enum AS ENUM ('initializing','active','closed','error');
CREATE TYPE proc_status_enum AS ENUM ('starting','running','stopped','error');
CREATE TYPE health_status_enum AS ENUM ('healthy','unhealthy','timeout','error');
CREATE TYPE storage_provider_enum AS ENUM ('s3','gcs','azure','cloudwatch','loki','elastic','clickhouse');
CREATE TYPE log_level_enum AS ENUM ('trace','debug','info','warn','error','fatal');
CREATE TYPE plan_type_enum AS ENUM ('free','pro','enterprise');
CREATE TYPE rate_limit_scope_enum AS ENUM ('global', 'server', 'user', 'ip');
CREATE TYPE aggregation_window_enum AS ENUM ('hourly', 'daily');

-- Generic trigger for updated_at
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END; $$ LANGUAGE plpgsql;
```

## Database Tables

### Core Tables

#### `organizations`
Primary tenant isolation table for multi-tenant support.
```sql
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug CITEXT UNIQUE NOT NULL,  -- Case-insensitive unique slugs
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT true,
    
    -- Commercial fields
    plan_type plan_type_enum DEFAULT 'free',
    max_servers INTEGER DEFAULT 10,
    max_sessions INTEGER DEFAULT 100,
    log_retention_days INTEGER DEFAULT 7,
    
    CONSTRAINT valid_log_retention CHECK (log_retention_days > 0 AND log_retention_days <= 3650)
);

CREATE TRIGGER organizations_updated_at 
    BEFORE UPDATE ON organizations 
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

#### `mcp_servers`
Registry of available MCP servers.
```sql
CREATE TABLE mcp_servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    protocol protocol_enum NOT NULL,
    
    -- Connection details
    url VARCHAR(500), -- For HTTP-based servers
    command VARCHAR(255), -- For stdio servers
    args TEXT[], -- Command arguments as array, NOT JSONB
    environment TEXT[], -- Environment variables as array (KEY=VALUE format)
    -- NOTE: Store secret refs like 'vault:secret/db/password' not raw values
    working_dir VARCHAR(500),
    
    -- Configuration
    version VARCHAR(50),
    weight INTEGER DEFAULT 100,
    timeout_seconds INTEGER DEFAULT 300,
    max_retries INTEGER DEFAULT 3,
    
    -- Status
    status server_status_enum DEFAULT 'active',
    health_check_url VARCHAR(500),
    is_active BOOLEAN DEFAULT true,
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    tags TEXT[], -- String array for simple tags, NOT JSONB
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(organization_id, name),
    CONSTRAINT valid_timeout CHECK (timeout_seconds > 0 AND timeout_seconds <= 3600),
    CONSTRAINT valid_weight CHECK (weight >= 0 AND weight <= 1000)
);

CREATE TRIGGER mcp_servers_updated_at 
    BEFORE UPDATE ON mcp_servers 
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

#### `mcp_sessions`
Active MCP proxy sessions.
```sql
CREATE TABLE mcp_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    server_id UUID NOT NULL REFERENCES mcp_servers(id) ON DELETE CASCADE,
    
    -- Session details
    status session_status_enum DEFAULT 'initializing',
    protocol protocol_enum NOT NULL,
    
    -- Observability IDs for correlation across proxy <-> DB <-> logs
    client_id UUID, -- Client/request identifier
    connection_id UUID, -- WebSocket/HTTP connection identifier
    
    -- Process info (for stdio sessions)
    process_pid INTEGER,
    process_status proc_status_enum,
    process_exit_code INTEGER,
    process_error TEXT,
    
    -- Timing
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE,
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- User tracking
    user_id VARCHAR(255) DEFAULT 'default-user',
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT valid_session_timing CHECK (
        ended_at IS NULL OR ended_at >= started_at
    )
);

CREATE TRIGGER mcp_sessions_updated_at 
    BEFORE UPDATE ON mcp_sessions 
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

### Health & Monitoring Tables

#### `health_checks`
Track server health status over time (consider monthly partitioning for scale).
```sql
CREATE TABLE health_checks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id UUID NOT NULL REFERENCES mcp_servers(id) ON DELETE CASCADE,
    
    status health_status_enum NOT NULL,
    response_time_ms INTEGER,
    response_body TEXT,
    error_message TEXT,
    
    checked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
) PARTITION BY RANGE (checked_at);

-- Create monthly partitions (example for current month)
CREATE TABLE health_checks_y2024m01 PARTITION OF health_checks
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

CREATE INDEX idx_health_checks_server_time ON health_checks(server_id, checked_at DESC);
```

#### `server_stats`
Aggregate statistics for load balancing and monitoring.
```sql
CREATE TABLE server_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id UUID NOT NULL REFERENCES mcp_servers(id) ON DELETE CASCADE,
    
    -- Counters
    total_requests BIGINT DEFAULT 0,
    success_requests BIGINT DEFAULT 0,
    error_requests BIGINT DEFAULT 0,
    active_sessions INTEGER DEFAULT 0,
    
    -- Performance
    avg_response_time_ms FLOAT DEFAULT 0,
    min_response_time_ms INTEGER DEFAULT 0,
    max_response_time_ms INTEGER DEFAULT 0,
    
    -- Time window
    window_start TIMESTAMP WITH TIME ZONE NOT NULL,
    window_end TIMESTAMP WITH TIME ZONE NOT NULL,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(server_id, window_start),
    CONSTRAINT valid_window_order CHECK (window_end > window_start)
);

CREATE TRIGGER server_stats_updated_at 
    BEFORE UPDATE ON server_stats 
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

### Logging & Audit Tables

#### `log_index`
Index table with pointers to detailed logs stored in object storage.
```sql
CREATE TABLE log_index (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    server_id UUID REFERENCES mcp_servers(id) ON DELETE SET NULL,
    session_id UUID REFERENCES mcp_sessions(id) ON DELETE SET NULL,
    
    -- Request classification
    rpc_method VARCHAR(100), -- e.g., 'tools/call', 'resources/read'
    level log_level_enum NOT NULL DEFAULT 'info',
    
    -- Timing and status
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    duration_ms INTEGER,
    status_code INTEGER,
    error_flag BOOLEAN DEFAULT false,
    
    -- Object storage pointers
    storage_provider storage_provider_enum NOT NULL,
    object_uri TEXT NOT NULL, -- e.g., s3://bucket/org/2025-01-20/server123/logs.jsonl.gz
    byte_offset BIGINT, -- Optional: position within file for packed logs
    
    -- Client context (kept minimal)
    user_id VARCHAR(255) DEFAULT 'default-user',
    remote_ip INET,
    
    -- Observability correlation
    client_id UUID, -- Match with mcp_sessions.client_id
    connection_id UUID, -- Match with mcp_sessions.connection_id
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for fast queries
CREATE INDEX idx_log_index_org_time ON log_index(organization_id, started_at DESC);
CREATE INDEX idx_log_index_server_time ON log_index(server_id, started_at DESC) WHERE server_id IS NOT NULL;
CREATE INDEX idx_log_index_session ON log_index(session_id, started_at DESC) WHERE session_id IS NOT NULL;
CREATE INDEX idx_log_index_method_time ON log_index(rpc_method, started_at DESC);
CREATE INDEX idx_log_index_errors ON log_index(organization_id, started_at DESC) WHERE error_flag = true;
CREATE INDEX idx_log_index_org_level_time ON log_index(organization_id, level, started_at DESC); -- Quick error/info drills
```

#### `audit_logs`
High-level audit trail for administrative actions (kept in Postgres).
```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Action details
    action VARCHAR(100) NOT NULL, -- 'server.created', 'session.started', etc.
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID,
    
    -- Actor
    actor_id VARCHAR(255) DEFAULT 'system',
    actor_ip INET,
    
    -- Changes
    old_values JSONB,
    new_values JSONB,
    
    -- Context
    metadata JSONB DEFAULT '{}',
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_org_time ON audit_logs(organization_id, created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id, created_at DESC);
```

### Rate Limiting Tables

#### `rate_limits`
Configure rate limiting rules.
```sql
CREATE TABLE rate_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Scope
    scope rate_limit_scope_enum NOT NULL,
    scope_id VARCHAR(255), -- server_id, user_id, or IP address
    
    -- Limits
    requests_per_minute INTEGER NOT NULL,
    requests_per_hour INTEGER,
    requests_per_day INTEGER,
    
    -- Burst settings
    burst_limit INTEGER,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(organization_id, scope, scope_id),
    CONSTRAINT valid_rate_limits CHECK (
        requests_per_minute > 0 AND 
        (requests_per_hour IS NULL OR requests_per_hour >= requests_per_minute) AND
        (requests_per_day IS NULL OR requests_per_day >= COALESCE(requests_per_hour, requests_per_minute))
    )
);

CREATE TRIGGER rate_limits_updated_at 
    BEFORE UPDATE ON rate_limits 
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

#### `rate_limit_usage`
Track rate limit usage (can be Redis-backed in production).
```sql
CREATE TABLE rate_limit_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    rate_limit_id UUID NOT NULL REFERENCES rate_limits(id) ON DELETE CASCADE,
    
    -- Tracking
    identifier VARCHAR(255) NOT NULL, -- user_id, IP, etc.
    window_start TIMESTAMP WITH TIME ZONE NOT NULL,
    request_count INTEGER DEFAULT 0,
    
    -- TTL for cleanup
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(rate_limit_id, identifier, window_start)
);

CREATE TRIGGER rate_limit_usage_updated_at 
    BEFORE UPDATE ON rate_limit_usage 
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

### Aggregates Tables

#### `log_aggregates`
Hourly/daily rollups for fast dashboard queries.
```sql
CREATE TYPE aggregation_window_enum AS ENUM ('hourly', 'daily');

CREATE TABLE log_aggregates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    server_id UUID REFERENCES mcp_servers(id) ON DELETE SET NULL,
    
    -- Aggregation details
    window_type aggregation_window_enum NOT NULL,
    window_start TIMESTAMP WITH TIME ZONE NOT NULL,
    rpc_method VARCHAR(100), -- NULL for all methods
    
    -- Metrics
    total_requests BIGINT DEFAULT 0,
    success_requests BIGINT DEFAULT 0,
    error_requests BIGINT DEFAULT 0,
    
    -- Performance percentiles (in milliseconds)
    p50_duration_ms INTEGER,
    p95_duration_ms INTEGER,
    p99_duration_ms INTEGER,
    avg_duration_ms FLOAT,
    
    -- Volume stats
    total_bytes_in BIGINT DEFAULT 0,
    total_bytes_out BIGINT DEFAULT 0,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(organization_id, server_id, window_type, window_start, rpc_method)
);

CREATE TRIGGER log_aggregates_updated_at 
    BEFORE UPDATE ON log_aggregates 
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_log_aggregates_org_window ON log_aggregates(organization_id, window_type, window_start DESC);
CREATE INDEX idx_log_aggregates_server_window ON log_aggregates(server_id, window_type, window_start DESC) WHERE server_id IS NOT NULL;
```

## Indexes

```sql
-- Performance indexes
CREATE INDEX idx_mcp_servers_org_active ON mcp_servers(organization_id, is_active);
CREATE INDEX idx_mcp_servers_org_status ON mcp_servers(organization_id, status); -- Added per request
CREATE INDEX idx_mcp_sessions_server_status ON mcp_sessions(server_id, status);
CREATE INDEX idx_mcp_sessions_org_active ON mcp_sessions(organization_id, status) WHERE status = 'active';

-- Cleanup indexes
CREATE INDEX idx_health_checks_cleanup ON health_checks(checked_at) WHERE checked_at < NOW() - INTERVAL '30 days';
CREATE INDEX idx_log_index_cleanup ON log_index(created_at) WHERE created_at < NOW() - INTERVAL '7 days';
CREATE INDEX idx_rate_limit_usage_cleanup ON rate_limit_usage(expires_at) WHERE expires_at < NOW();

-- Composite indexes for common queries
CREATE INDEX idx_log_index_org_method_time ON log_index(organization_id, rpc_method, started_at DESC);
CREATE INDEX idx_log_aggregates_dashboard ON log_aggregates(organization_id, window_type, window_start DESC) WHERE rpc_method IS NULL;

-- Observability correlation indexes
CREATE INDEX idx_mcp_sessions_client_id ON mcp_sessions(client_id) WHERE client_id IS NOT NULL;
CREATE INDEX idx_log_index_client_id ON log_index(client_id) WHERE client_id IS NOT NULL;
```

## Row Level Security (RLS)

Multi-tenant isolation policies (disabled in OSS, enabled in commercial).

```sql
-- Enable RLS on all tenant-isolated tables (disabled by default for OSS)
-- ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE mcp_servers ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE mcp_sessions ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE log_index ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE rate_limits ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE rate_limit_usage ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE log_aggregates ENABLE ROW LEVEL SECURITY;

-- RLS policies for tenant isolation
-- Usage: SET app.current_org = 'org-uuid'; before queries

-- Organizations: users can only see their own org
-- CREATE POLICY org_isolation ON organizations
--     USING (id = current_setting('app.current_org')::uuid);

-- All other tables: filter by organization_id
-- CREATE POLICY org_isolation ON mcp_servers
--     USING (organization_id = current_setting('app.current_org')::uuid);

-- CREATE POLICY org_isolation ON mcp_sessions
--     USING (organization_id = current_setting('app.current_org')::uuid);

-- CREATE POLICY org_isolation ON log_index
--     USING (organization_id = current_setting('app.current_org')::uuid);

-- CREATE POLICY org_isolation ON audit_logs
--     USING (organization_id = current_setting('app.current_org')::uuid);

-- CREATE POLICY org_isolation ON rate_limits
--     USING (organization_id = current_setting('app.current_org')::uuid);

-- CREATE POLICY org_isolation ON rate_limit_usage
--     USING (organization_id = current_setting('app.current_org')::uuid);

-- CREATE POLICY org_isolation ON log_aggregates
--     USING (organization_id = current_setting('app.current_org')::uuid);
```

## Views

### Active Sessions View
```sql
CREATE VIEW active_sessions AS
SELECT 
    s.id,
    s.organization_id,
    s.server_id,
    srv.name as server_name,
    srv.protocol,
    s.status,
    s.started_at,
    s.last_activity,
    EXTRACT(EPOCH FROM (NOW() - s.started_at)) as duration_seconds
FROM mcp_sessions s
JOIN mcp_servers srv ON s.server_id = srv.id
WHERE s.status = 'active';
```

### Server Health Summary
```sql
CREATE VIEW server_health_summary AS
SELECT 
    s.id,
    s.organization_id,
    s.name,
    s.status,
    hc.status as last_health_status,
    hc.response_time_ms as last_response_time,
    hc.checked_at as last_health_check,
    COALESCE(ss.active_sessions, 0) as active_sessions,
    COALESCE(ss.total_requests, 0) as total_requests
FROM mcp_servers s
LEFT JOIN LATERAL (
    SELECT status, response_time_ms, checked_at
    FROM health_checks
    WHERE server_id = s.id
    ORDER BY checked_at DESC
    LIMIT 1
) hc ON true
LEFT JOIN LATERAL (
    SELECT active_sessions, total_requests
    FROM server_stats
    WHERE server_id = s.id
    ORDER BY window_end DESC
    LIMIT 1
) ss ON true;
```

## Log Retention & Cleanup

### Automated Cleanup Jobs
```sql
-- Example cleanup job (run daily via cron/scheduler)
-- Prune log_index records based on organization retention policy
DELETE FROM log_index 
WHERE created_at < NOW() - INTERVAL '1 day' * (
    SELECT log_retention_days 
    FROM organizations 
    WHERE id = log_index.organization_id
);

-- Cleanup old health checks (keep 30 days max)
DELETE FROM health_checks 
WHERE checked_at < NOW() - INTERVAL '30 days';

-- Cleanup expired rate limit usage
DELETE FROM rate_limit_usage 
WHERE expires_at < NOW();
```

### S3/CloudWatch Lifecycle Management
- **Free tier**: 7-day retention with S3 lifecycle rules
- **Pro tier**: 30-day retention with transition to IA after 7 days
- **Enterprise**: 1-year retention with tiered storage (Standard → IA → Glacier)
- **Customer buckets**: Managed by customer's own lifecycle policies

### Time Sync Requirements
- **Critical**: Ensure NTP/chrony is configured on all database servers
- **Recommended**: Use UTC timezone for all timestamps
- **Monitoring**: Alert on clock drift > 1 second between replicas

## Migration Strategy

### Phase 1: Single Tenant (Current)
- Create all tables with organization_id
- Insert default organization record  
- All operations use 'default-org'
- RLS policies commented out (OSS mode)

### Phase 2: Multi-Tenant Ready (Commercial)
- Add proper organization management
- Add user authentication with JWT containing org claim
- Enable RLS policies: `ALTER TABLE ... ENABLE ROW LEVEL SECURITY`
- Implement tenant isolation middleware: `SET app.current_org = ?`
- Add organization-level quotas and limits

## Multi-Tenant Considerations

### Tenant Isolation Strategies
1. **Path Prefixing**: `s3://logs/org/<org_id>/date=YYYY-MM-DD/server_id/logs.jsonl.gz`
2. **Per-Tenant Retention**: Lifecycle rules on prefixes (free=7d, pro=30d, enterprise=1y)
3. **Customer-Owned Buckets** (Enterprise): Write to customer's bucket with STS/assumed role
4. **PII Safety**: Redaction at emit time, add `redaction_version` to log_index

### What to Keep in Postgres vs Object Storage
**Keep in Postgres (Control Plane)**:
- Audit events and admin actions
- Log index with pointers to detailed logs
- Aggregated rollups (hourly/daily metrics)
- Health checks and server stats
- Rate limiting metadata

**Store in Object Storage (Data Plane)**:
- Full request/response bodies
- Headers and detailed metadata
- Large JSON payloads
- Binary data and attachments

### Query Patterns
1. **Dashboard Queries**: Fast queries against Postgres aggregates
2. **Detail Views**: Fetch specific log by `object_uri` from storage
3. **Analytics**: Point Athena/BigQuery at object storage for ad-hoc queries
4. **Search**: Use log aggregation tools (Loki, Elastic) for full-text search

## Default Data

```sql
-- Insert default organization for single-tenant mode
INSERT INTO organizations (id, name, slug, plan_type, max_servers, max_sessions, log_retention_days)
VALUES (
    '00000000-0000-0000-0000-000000000000',
    'Default Organization',
    'default-org',
    'enterprise',
    1000,
    10000,
    365
);

-- Default rate limits
INSERT INTO rate_limits (organization_id, scope, requests_per_minute, requests_per_hour, requests_per_day)
VALUES (
    '00000000-0000-0000-0000-000000000000',
    'global',
    1000,
    10000,
    100000
);
```
