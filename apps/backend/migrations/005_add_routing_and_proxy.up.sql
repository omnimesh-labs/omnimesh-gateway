-- Add routing and proxy functionality including content filters
-- This migration adds tables for request routing, proxying, and content filtering

-- Drop existing types if they exist
DROP TYPE IF EXISTS filter_type_enum CASCADE;
DROP TYPE IF EXISTS filter_action_enum CASCADE;

-- Create enums for content filtering
CREATE TYPE filter_type_enum AS ENUM ('pii', 'resource', 'deny', 'regex');
CREATE TYPE filter_action_enum AS ENUM ('block', 'warn', 'audit', 'allow');

-- Content Filters - Database-driven filter configuration
CREATE TABLE content_filters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type filter_type_enum NOT NULL,
    enabled BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 100 CHECK (priority >= 1 AND priority <= 1000),
    config JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    
    UNIQUE(organization_id, name)
);

CREATE TRIGGER content_filters_updated_at 
    BEFORE UPDATE ON content_filters 
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Filter Violations - Track filter violations for audit and reporting
CREATE TABLE filter_violations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    filter_id UUID NOT NULL REFERENCES content_filters(id) ON DELETE CASCADE,
    request_id VARCHAR(255) NOT NULL,
    session_id UUID REFERENCES mcp_sessions(id) ON DELETE SET NULL,
    server_id UUID REFERENCES mcp_servers(id) ON DELETE SET NULL,
    
    -- Violation details
    violation_type VARCHAR(100) NOT NULL,
    action_taken filter_action_enum NOT NULL,
    content_snippet TEXT, -- Partial content that triggered the violation (limited length for privacy)
    pattern_matched VARCHAR(500),
    severity VARCHAR(20) DEFAULT 'medium' CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    
    -- Context
    user_id VARCHAR(255) DEFAULT 'default-user',
    remote_ip INET,
    user_agent TEXT,
    direction VARCHAR(20) CHECK (direction IN ('inbound', 'outbound')),
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Proxy Routes - Define routing rules for proxying requests to different backends
CREATE TABLE proxy_routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Route matching
    path_pattern VARCHAR(500) NOT NULL, -- e.g., "/api/v1/servers/{server_id}/*"
    method_pattern VARCHAR(20) DEFAULT '*', -- HTTP method or * for all
    host_pattern VARCHAR(255) DEFAULT '*', -- Host header pattern
    
    -- Target configuration
    target_type VARCHAR(50) NOT NULL DEFAULT 'mcp_server' CHECK (target_type IN ('mcp_server', 'http_backend', 'virtual_server')),
    target_config JSONB NOT NULL DEFAULT '{}',
    
    -- Route behavior
    enabled BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 100 CHECK (priority >= 1 AND priority <= 1000),
    timeout_seconds INTEGER DEFAULT 30 CHECK (timeout_seconds > 0),
    max_retries INTEGER DEFAULT 3 CHECK (max_retries >= 0),
    
    -- Load balancing (for multiple targets)
    load_balancer_type VARCHAR(50) DEFAULT 'round_robin' CHECK (load_balancer_type IN ('round_robin', 'least_conn', 'weighted', 'random')),
    health_check_enabled BOOLEAN DEFAULT true,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(organization_id, name)
);

CREATE TRIGGER proxy_routes_updated_at 
    BEFORE UPDATE ON proxy_routes 
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Request Routing Log - Track routing decisions for debugging and analytics
CREATE TABLE request_routing_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    request_id VARCHAR(255) NOT NULL,
    
    -- Request details
    method VARCHAR(20) NOT NULL,
    path VARCHAR(500) NOT NULL,
    host VARCHAR(255),
    user_agent TEXT,
    remote_ip INET,
    
    -- Routing decision
    matched_route_id UUID REFERENCES proxy_routes(id) ON DELETE SET NULL,
    target_server_id UUID REFERENCES mcp_servers(id) ON DELETE SET NULL,
    routing_decision VARCHAR(100) NOT NULL, -- 'routed', 'blocked', 'no_match', 'error'
    
    -- Performance
    route_resolution_time_ms INTEGER,
    total_request_time_ms INTEGER,
    
    -- Status
    status_code INTEGER,
    error_message TEXT,
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Performance indexes

-- Content Filters indexes
CREATE INDEX idx_content_filters_org_enabled ON content_filters(organization_id, enabled);
CREATE INDEX idx_content_filters_org_type ON content_filters(organization_id, type);
CREATE INDEX idx_content_filters_priority ON content_filters(enabled, priority) WHERE enabled = true;

-- Filter Violations indexes (for reporting and analytics)
CREATE INDEX idx_filter_violations_org_time ON filter_violations(organization_id, created_at DESC);
CREATE INDEX idx_filter_violations_filter_time ON filter_violations(filter_id, created_at DESC);
CREATE INDEX idx_filter_violations_request ON filter_violations(request_id);
CREATE INDEX idx_filter_violations_severity ON filter_violations(severity, created_at DESC);
CREATE INDEX idx_filter_violations_type ON filter_violations(violation_type, created_at DESC);
CREATE INDEX idx_filter_violations_cleanup ON filter_violations(created_at); -- For cleanup jobs

-- Proxy Routes indexes
CREATE INDEX idx_proxy_routes_org_enabled ON proxy_routes(organization_id, enabled);
CREATE INDEX idx_proxy_routes_priority ON proxy_routes(enabled, priority) WHERE enabled = true;
CREATE INDEX idx_proxy_routes_target_type ON proxy_routes(target_type, enabled);

-- Request Routing Log indexes
CREATE INDEX idx_request_routing_log_org_time ON request_routing_log(organization_id, created_at DESC);
CREATE INDEX idx_request_routing_log_request_id ON request_routing_log(request_id);
CREATE INDEX idx_request_routing_log_route_time ON request_routing_log(matched_route_id, created_at DESC) WHERE matched_route_id IS NOT NULL;
CREATE INDEX idx_request_routing_log_decision ON request_routing_log(routing_decision, created_at DESC);
CREATE INDEX idx_request_routing_log_cleanup ON request_routing_log(created_at); -- For cleanup jobs

-- Insert default content filters for the default organization
INSERT INTO content_filters (organization_id, name, description, type, enabled, priority, config)
VALUES 
    (
        '00000000-0000-0000-0000-000000000000',
        'default-pii-filter',
        'Default PII detection filter',
        'pii',
        true,
        10,
        '{
            "patterns": {
                "ssn": true,
                "credit_card": true,
                "email": true,
                "phone": true,
                "aws_keys": true,
                "ip_address": false
            },
            "masking_strategy": "redact",
            "action": "warn",
            "log_violations": true
        }'
    ),
    (
        '00000000-0000-0000-0000-000000000000',
        'default-resource-filter',
        'Default resource access filter',
        'resource',
        true,
        20,
        '{
            "allowed_protocols": ["https", "http"],
            "blocked_domains": ["localhost", "127.0.0.1", "0.0.0.0"],
            "max_content_size": 10485760,
            "allowed_content_types": ["application/json", "text/plain", "text/html"],
            "action": "block",
            "log_violations": true
        }'
    ),
    (
        '00000000-0000-0000-0000-000000000000',
        'default-deny-filter',
        'Default content blocking filter',
        'deny',
        false,
        30,
        '{
            "blocked_words": ["password", "secret", "token"],
            "case_sensitive": false,
            "action": "warn",
            "log_violations": true
        }'
    )
ON CONFLICT (organization_id, name) DO NOTHING;

-- Insert default proxy route for MCP servers
INSERT INTO proxy_routes (organization_id, name, description, path_pattern, method_pattern, target_type, target_config, enabled, priority)
VALUES (
    '00000000-0000-0000-0000-000000000000',
    'default-mcp-route',
    'Default route for MCP server requests',
    '/mcp/*',
    '*',
    'mcp_server',
    '{
        "load_balancing": "round_robin",
        "health_check": true,
        "failover_enabled": true
    }',
    true,
    100
) ON CONFLICT (organization_id, name) DO NOTHING;