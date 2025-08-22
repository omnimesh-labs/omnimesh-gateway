-- MCP Gateway Database ERD Implementation
-- This migration implements the complete ERD schema for the MCP Gateway

-- Required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";    -- For UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";     -- For gen_random_uuid (secure UUIDs)
CREATE EXTENSION IF NOT EXISTS "citext";       -- Case-insensitive text

-- Drop existing tables if they conflict (in reverse dependency order)
DROP TABLE IF EXISTS mcp_sessions CASCADE;
DROP TABLE IF EXISTS security_logs CASCADE;
DROP TABLE IF EXISTS error_logs CASCADE;
DROP TABLE IF EXISTS performance_logs CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS request_logs CASCADE;
DROP TABLE IF EXISTS rate_limit_policies CASCADE;
DROP TABLE IF EXISTS policies CASCADE;
DROP TABLE IF EXISTS mcp_servers CASCADE;
DROP TABLE IF EXISTS api_keys CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS organizations CASCADE;

-- Drop existing types if they exist
DROP TYPE IF EXISTS protocol_enum CASCADE;
DROP TYPE IF EXISTS server_status_enum CASCADE;
DROP TYPE IF EXISTS session_status_enum CASCADE;
DROP TYPE IF EXISTS proc_status_enum CASCADE;
DROP TYPE IF EXISTS health_status_enum CASCADE;
DROP TYPE IF EXISTS storage_provider_enum CASCADE;
DROP TYPE IF EXISTS log_level_enum CASCADE;
DROP TYPE IF EXISTS plan_type_enum CASCADE;
DROP TYPE IF EXISTS rate_limit_scope_enum CASCADE;
DROP TYPE IF EXISTS aggregation_window_enum CASCADE;

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

-- Core Tables

-- Organizations - Primary tenant isolation table for multi-tenant support
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

-- MCP Servers - Registry of available MCP servers
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
    CONSTRAINT valid_timeout CHECK (timeout_seconds > 0 AND timeout_seconds <= 3600)
);

CREATE TRIGGER mcp_servers_updated_at 
    BEFORE UPDATE ON mcp_servers 
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- MCP Sessions - Active MCP sessions
CREATE TABLE mcp_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    server_id UUID NOT NULL REFERENCES mcp_servers(id) ON DELETE CASCADE,
    
    -- Session details
    status session_status_enum DEFAULT 'initializing',
    protocol protocol_enum NOT NULL,
    
    -- Observability IDs for correlation across sessions <-> DB <-> logs
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

-- Health & Monitoring Tables

-- Health Checks - Track server health status over time (partitioned for scale)
CREATE TABLE health_checks (
    id UUID DEFAULT gen_random_uuid(),
    server_id UUID NOT NULL REFERENCES mcp_servers(id) ON DELETE CASCADE,
    
    status health_status_enum NOT NULL,
    response_time_ms INTEGER,
    response_body TEXT,
    error_message TEXT,
    
    checked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Primary key must include partition key for partitioned tables
    PRIMARY KEY (id, checked_at)
) PARTITION BY RANGE (checked_at);

-- Create current month partition (you'll need to create new partitions monthly)
CREATE TABLE health_checks_current PARTITION OF health_checks
    FOR VALUES FROM (date_trunc('month', NOW())) TO (date_trunc('month', NOW() + INTERVAL '1 month'));

-- Server Stats - Aggregate statistics for monitoring
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

-- Logging & Audit Tables

-- Log Index - Index table with pointers to detailed logs stored in object storage
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

-- Audit Logs - High-level audit trail for administrative actions (kept in Postgres)
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

-- Rate Limiting Tables

-- Rate Limits - Configure rate limiting rules
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

-- Rate Limit Usage - Track rate limit usage (can be Redis-backed in production)
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

-- Aggregates Tables

-- Log Aggregates - Hourly/daily rollups for fast dashboard queries
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

-- Performance Indexes

-- Organizations indexes
CREATE INDEX idx_organizations_slug ON organizations(slug);
CREATE INDEX idx_organizations_active ON organizations(is_active);

-- MCP Servers indexes
CREATE INDEX idx_mcp_servers_org_active ON mcp_servers(organization_id, is_active);
CREATE INDEX idx_mcp_servers_org_status ON mcp_servers(organization_id, status);
CREATE INDEX idx_mcp_servers_protocol ON mcp_servers(protocol);
CREATE INDEX idx_mcp_servers_tags ON mcp_servers USING gin(tags);

-- MCP Sessions indexes
CREATE INDEX idx_mcp_sessions_server_status ON mcp_sessions(server_id, status);
CREATE INDEX idx_mcp_sessions_org_active ON mcp_sessions(organization_id, status) WHERE status = 'active';
CREATE INDEX idx_mcp_sessions_client_id ON mcp_sessions(client_id) WHERE client_id IS NOT NULL;

-- Health Checks indexes
CREATE INDEX idx_health_checks_server_time ON health_checks(server_id, checked_at DESC);
CREATE INDEX idx_health_checks_cleanup ON health_checks(checked_at);

-- Server Stats indexes
CREATE INDEX idx_server_stats_server_window ON server_stats(server_id, window_start DESC);

-- Log Index indexes for fast queries
CREATE INDEX idx_log_index_org_time ON log_index(organization_id, started_at DESC);
CREATE INDEX idx_log_index_server_time ON log_index(server_id, started_at DESC) WHERE server_id IS NOT NULL;
CREATE INDEX idx_log_index_session ON log_index(session_id, started_at DESC) WHERE session_id IS NOT NULL;
CREATE INDEX idx_log_index_method_time ON log_index(rpc_method, started_at DESC);
CREATE INDEX idx_log_index_errors ON log_index(organization_id, started_at DESC) WHERE error_flag = true;
CREATE INDEX idx_log_index_org_level_time ON log_index(organization_id, level, started_at DESC);
CREATE INDEX idx_log_index_cleanup ON log_index(created_at);
CREATE INDEX idx_log_index_client_id ON log_index(client_id) WHERE client_id IS NOT NULL;

-- Audit Logs indexes
CREATE INDEX idx_audit_logs_org_time ON audit_logs(organization_id, created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id, created_at DESC);

-- Rate Limits indexes
CREATE INDEX idx_rate_limits_org_scope ON rate_limits(organization_id, scope, scope_id);
CREATE INDEX idx_rate_limits_active ON rate_limits(is_active);

-- Rate Limit Usage indexes
CREATE INDEX idx_rate_limit_usage_cleanup ON rate_limit_usage(expires_at);
CREATE INDEX idx_rate_limit_usage_window ON rate_limit_usage(rate_limit_id, window_start);

-- Log Aggregates indexes
CREATE INDEX idx_log_aggregates_org_window ON log_aggregates(organization_id, window_type, window_start DESC);
CREATE INDEX idx_log_aggregates_server_window ON log_aggregates(server_id, window_type, window_start DESC) WHERE server_id IS NOT NULL;
CREATE INDEX idx_log_aggregates_dashboard ON log_aggregates(organization_id, window_type, window_start DESC) WHERE rpc_method IS NULL;

-- Composite indexes for common queries
CREATE INDEX idx_log_index_org_method_time ON log_index(organization_id, rpc_method, started_at DESC);

-- Views

-- Active Sessions View
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

-- Server Health Summary
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

-- Authentication Tables (Users, API Keys, Policies)

-- Users table - Authentication and authorization
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email CITEXT UNIQUE NOT NULL, -- Case-insensitive email
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'user', 'viewer', 'api_user', 'system_admin')),
    is_active BOOLEAN DEFAULT true,
    email_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TRIGGER users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- API Keys table
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    prefix VARCHAR(20) NOT NULL, -- For key identification (e.g., "mcp_sk_")
    permissions TEXT[] DEFAULT ARRAY[]::TEXT[], -- Array of permission strings
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TRIGGER api_keys_updated_at 
    BEFORE UPDATE ON api_keys 
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Policies table - Access control policies
CREATE TABLE policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL CHECK (type IN ('access', 'rate_limit', 'security')),
    priority INTEGER DEFAULT 100 CHECK (priority >= 1 AND priority <= 1000),
    conditions JSONB NOT NULL DEFAULT '{}',
    actions JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    
    UNIQUE(organization_id, name)
);

CREATE TRIGGER policies_updated_at 
    BEFORE UPDATE ON policies 
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Users indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_org_active ON users(organization_id, is_active);
CREATE INDEX idx_users_role ON users(role);

-- API Keys indexes
CREATE INDEX idx_api_keys_user_org ON api_keys(user_id, organization_id);
CREATE INDEX idx_api_keys_active ON api_keys(is_active);
CREATE INDEX idx_api_keys_expires ON api_keys(expires_at);
CREATE INDEX idx_api_keys_prefix ON api_keys(prefix);

-- Policies indexes
CREATE INDEX idx_policies_org_active ON policies(organization_id, is_active);
CREATE INDEX idx_policies_type_priority ON policies(type, priority);

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
) ON CONFLICT (id) DO NOTHING;

-- Default rate limits
INSERT INTO rate_limits (organization_id, scope, requests_per_minute, requests_per_hour, requests_per_day)
VALUES (
    '00000000-0000-0000-0000-000000000000',
    'global',
    1000,
    10000,
    100000
) ON CONFLICT (organization_id, scope, scope_id) DO NOTHING;
