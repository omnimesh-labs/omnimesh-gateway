-- Initial schema for MCP Gateway

-- Organizations table
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(255) UNIQUE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'user', 'viewer', 'service')),
    is_active BOOLEAN DEFAULT true,
    email_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- API Keys table
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    permissions JSONB DEFAULT '[]',
    is_active BOOLEAN DEFAULT true,
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- MCP Servers table
CREATE TABLE mcp_servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    url VARCHAR(512) NOT NULL,
    description TEXT,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    version VARCHAR(50),
    protocol VARCHAR(20) NOT NULL DEFAULT 'http' CHECK (protocol IN ('http', 'websocket', 'grpc')),
    status VARCHAR(20) NOT NULL DEFAULT 'unknown' CHECK (status IN ('healthy', 'unhealthy', 'maintenance', 'unknown')),
    health_endpoint VARCHAR(512),
    tags JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    weight INTEGER DEFAULT 100 CHECK (weight >= 1 AND weight <= 100),
    max_concurrency INTEGER DEFAULT 100,
    timeout_ms INTEGER DEFAULT 30000,
    retry_policy JSONB,
    is_active BOOLEAN DEFAULT true,
    last_health_check TIMESTAMP WITH TIME ZONE,
    failure_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Policies table
CREATE TABLE policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    policy_type VARCHAR(50) NOT NULL CHECK (policy_type IN ('access', 'rate_limit', 'routing', 'logging', 'security', 'transform')),
    rules JSONB NOT NULL DEFAULT '[]',
    priority INTEGER DEFAULT 100 CHECK (priority >= 1 AND priority <= 100),
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id)
);

-- Rate Limit Policies table
CREATE TABLE rate_limit_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    server_id UUID REFERENCES mcp_servers(id) ON DELETE CASCADE,
    endpoint VARCHAR(512),
    requests_per_minute INTEGER,
    requests_per_hour INTEGER,
    requests_per_day INTEGER,
    burst_limit INTEGER,
    window_size_ms BIGINT NOT NULL DEFAULT 60000,
    is_active BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 100,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Request Logs table
CREATE TABLE request_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id VARCHAR(255) NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    server_id UUID REFERENCES mcp_servers(id) ON DELETE SET NULL,
    method VARCHAR(10) NOT NULL,
    path TEXT NOT NULL,
    status_code INTEGER NOT NULL,
    request_size BIGINT DEFAULT 0,
    response_size BIGINT DEFAULT 0,
    duration_ms BIGINT NOT NULL,
    client_ip INET,
    user_agent TEXT,
    error TEXT,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Audit Logs table
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    resource_id UUID,
    details JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Performance Logs table
CREATE TABLE performance_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id VARCHAR(255) NOT NULL,
    server_id UUID REFERENCES mcp_servers(id) ON DELETE SET NULL,
    endpoint VARCHAR(512) NOT NULL,
    method VARCHAR(10) NOT NULL,
    duration_ms BIGINT NOT NULL,
    request_size BIGINT DEFAULT 0,
    response_size BIGINT DEFAULT 0,
    memory_usage BIGINT DEFAULT 0,
    cpu_usage REAL DEFAULT 0,
    database_time_ms BIGINT DEFAULT 0,
    external_time_ms BIGINT DEFAULT 0,
    cache_hit BOOLEAN DEFAULT false,
    metadata JSONB DEFAULT '{}',
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Error Logs table
CREATE TABLE error_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id VARCHAR(255),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    server_id UUID REFERENCES mcp_servers(id) ON DELETE SET NULL,
    error_type VARCHAR(100) NOT NULL,
    error_code VARCHAR(50) NOT NULL,
    error_message TEXT NOT NULL,
    stack_trace TEXT,
    context JSONB DEFAULT '{}',
    severity VARCHAR(20) NOT NULL DEFAULT 'medium' CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    resolved BOOLEAN DEFAULT false,
    resolved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    resolved_at TIMESTAMP WITH TIME ZONE,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Security Logs table
CREATE TABLE security_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    event_type VARCHAR(100) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'medium' CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    ip_address INET NOT NULL,
    user_agent TEXT,
    details JSONB DEFAULT '{}',
    action VARCHAR(50) NOT NULL CHECK (action IN ('blocked', 'allowed', 'flagged')),
    rule_id UUID,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- MCP Sessions table (for tracking active sessions)
CREATE TABLE mcp_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id VARCHAR(255) UNIQUE NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    server_id UUID NOT NULL REFERENCES mcp_servers(id) ON DELETE CASCADE,
    client_info JSONB NOT NULL DEFAULT '{}',
    start_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    message_count BIGINT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}'
);

-- Indexes for better performance

-- Users indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_organization_id ON users(organization_id);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_is_active ON users(is_active);

-- API Keys indexes
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_organization_id ON api_keys(organization_id);
CREATE INDEX idx_api_keys_is_active ON api_keys(is_active);
CREATE INDEX idx_api_keys_expires_at ON api_keys(expires_at);

-- MCP Servers indexes
CREATE INDEX idx_mcp_servers_organization_id ON mcp_servers(organization_id);
CREATE INDEX idx_mcp_servers_status ON mcp_servers(status);
CREATE INDEX idx_mcp_servers_protocol ON mcp_servers(protocol);
CREATE INDEX idx_mcp_servers_is_active ON mcp_servers(is_active);
CREATE INDEX idx_mcp_servers_tags ON mcp_servers USING gin(tags);

-- Policies indexes
CREATE INDEX idx_policies_organization_id ON policies(organization_id);
CREATE INDEX idx_policies_policy_type ON policies(policy_type);
CREATE INDEX idx_policies_is_active ON policies(is_active);
CREATE INDEX idx_policies_priority ON policies(priority);

-- Rate Limit Policies indexes
CREATE INDEX idx_rate_limit_policies_organization_id ON rate_limit_policies(organization_id);
CREATE INDEX idx_rate_limit_policies_user_id ON rate_limit_policies(user_id);
CREATE INDEX idx_rate_limit_policies_server_id ON rate_limit_policies(server_id);
CREATE INDEX idx_rate_limit_policies_is_active ON rate_limit_policies(is_active);

-- Request Logs indexes (for analytics and monitoring)
CREATE INDEX idx_request_logs_user_id ON request_logs(user_id);
CREATE INDEX idx_request_logs_organization_id ON request_logs(organization_id);
CREATE INDEX idx_request_logs_server_id ON request_logs(server_id);
CREATE INDEX idx_request_logs_timestamp ON request_logs(timestamp);
CREATE INDEX idx_request_logs_status_code ON request_logs(status_code);
CREATE INDEX idx_request_logs_method ON request_logs(method);

-- Audit Logs indexes
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_organization_id ON audit_logs(organization_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp);

-- Performance Logs indexes
CREATE INDEX idx_performance_logs_server_id ON performance_logs(server_id);
CREATE INDEX idx_performance_logs_endpoint ON performance_logs(endpoint);
CREATE INDEX idx_performance_logs_timestamp ON performance_logs(timestamp);
CREATE INDEX idx_performance_logs_duration_ms ON performance_logs(duration_ms);

-- Error Logs indexes
CREATE INDEX idx_error_logs_user_id ON error_logs(user_id);
CREATE INDEX idx_error_logs_organization_id ON error_logs(organization_id);
CREATE INDEX idx_error_logs_server_id ON error_logs(server_id);
CREATE INDEX idx_error_logs_error_type ON error_logs(error_type);
CREATE INDEX idx_error_logs_severity ON error_logs(severity);
CREATE INDEX idx_error_logs_resolved ON error_logs(resolved);
CREATE INDEX idx_error_logs_timestamp ON error_logs(timestamp);

-- Security Logs indexes
CREATE INDEX idx_security_logs_user_id ON security_logs(user_id);
CREATE INDEX idx_security_logs_organization_id ON security_logs(organization_id);
CREATE INDEX idx_security_logs_event_type ON security_logs(event_type);
CREATE INDEX idx_security_logs_severity ON security_logs(severity);
CREATE INDEX idx_security_logs_ip_address ON security_logs(ip_address);
CREATE INDEX idx_security_logs_timestamp ON security_logs(timestamp);

-- MCP Sessions indexes
CREATE INDEX idx_mcp_sessions_user_id ON mcp_sessions(user_id);
CREATE INDEX idx_mcp_sessions_organization_id ON mcp_sessions(organization_id);
CREATE INDEX idx_mcp_sessions_server_id ON mcp_sessions(server_id);
CREATE INDEX idx_mcp_sessions_is_active ON mcp_sessions(is_active);
CREATE INDEX idx_mcp_sessions_last_activity ON mcp_sessions(last_activity);

-- Updated at triggers
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON organizations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_mcp_servers_updated_at BEFORE UPDATE ON mcp_servers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_policies_updated_at BEFORE UPDATE ON policies FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_rate_limit_policies_updated_at BEFORE UPDATE ON rate_limit_policies FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
