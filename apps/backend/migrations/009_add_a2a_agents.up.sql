-- Add A2A (Agent-to-Agent) integration support
-- This migration adds support for managing external AI agents and exposing them through the Omnimesh Gateway

-- Add agent type enum for different AI agent types
CREATE TYPE agent_type_enum AS ENUM ('custom', 'openai', 'anthropic', 'generic');

-- Add auth type enum for different authentication methods
CREATE TYPE auth_type_enum AS ENUM ('api_key', 'bearer', 'oauth', 'none');

-- A2A Agents - Configuration for external AI agents
CREATE TABLE a2a_agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Basic information
    name VARCHAR(255) NOT NULL,
    description TEXT,
    endpoint_url VARCHAR(500) NOT NULL,
    agent_type agent_type_enum NOT NULL DEFAULT 'custom',
    protocol_version VARCHAR(50) DEFAULT '1.0',

    -- Capabilities stored as JSONB for flexibility
    capabilities JSONB NOT NULL DEFAULT '{"chat": true, "tools": false, "streaming": false}',

    -- Configuration parameters (max_tokens, temperature, etc.)
    config JSONB DEFAULT '{}',

    -- Authentication configuration
    auth_type auth_type_enum NOT NULL DEFAULT 'none',
    auth_value TEXT, -- Encrypted credentials stored here

    -- Status and metadata
    is_active BOOLEAN DEFAULT true,
    tags TEXT[] DEFAULT ARRAY[]::TEXT[],
    metadata JSONB DEFAULT '{}',

    -- Health tracking
    last_health_check TIMESTAMP WITH TIME ZONE,
    health_status VARCHAR(50) DEFAULT 'unknown', -- unknown, healthy, unhealthy
    health_error TEXT,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Constraints
    UNIQUE(organization_id, name),
    CONSTRAINT valid_capabilities_format CHECK (
        jsonb_typeof(capabilities) = 'object'
    ),
    CONSTRAINT valid_config_format CHECK (
        jsonb_typeof(config) = 'object'
    ),
    CONSTRAINT valid_endpoint_url CHECK (
        endpoint_url ~ '^https?://.+'
    )
);

-- Trigger for updated_at
CREATE TRIGGER a2a_agents_updated_at
    BEFORE UPDATE ON a2a_agents
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Performance indexes
CREATE INDEX idx_a2a_agents_org_active ON a2a_agents(organization_id, is_active);
CREATE INDEX idx_a2a_agents_agent_type ON a2a_agents(agent_type);
CREATE INDEX idx_a2a_agents_name ON a2a_agents(organization_id, name);
CREATE INDEX idx_a2a_agents_health ON a2a_agents(health_status, last_health_check);
CREATE INDEX idx_a2a_agents_tags ON a2a_agents USING gin(tags);

-- A2A Agent Tools - Link agents to virtual server tools
CREATE TABLE a2a_agent_tools (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES a2a_agents(id) ON DELETE CASCADE,
    virtual_server_id UUID NOT NULL REFERENCES virtual_servers(id) ON DELETE CASCADE,
    tool_name VARCHAR(255) NOT NULL,
    tool_config JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(agent_id, virtual_server_id, tool_name)
);

-- Trigger for updated_at
CREATE TRIGGER a2a_agent_tools_updated_at
    BEFORE UPDATE ON a2a_agent_tools
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Performance indexes for agent tools
CREATE INDEX idx_a2a_agent_tools_agent ON a2a_agent_tools(agent_id);
CREATE INDEX idx_a2a_agent_tools_virtual_server ON a2a_agent_tools(virtual_server_id);
CREATE INDEX idx_a2a_agent_tools_active ON a2a_agent_tools(is_active);

-- Add A2A agent support to audit logs
ALTER TABLE audit_logs ADD COLUMN a2a_agent_id UUID REFERENCES a2a_agents(id) ON DELETE SET NULL;
CREATE INDEX idx_audit_logs_a2a_agent ON audit_logs(a2a_agent_id, created_at DESC) WHERE a2a_agent_id IS NOT NULL;

-- Insert example OpenAI agent for testing
INSERT INTO a2a_agents (
    organization_id,
    name,
    description,
    endpoint_url,
    agent_type,
    protocol_version,
    capabilities,
    config,
    auth_type,
    metadata
) VALUES (
    '00000000-0000-0000-0000-000000000000',
    'OpenAI GPT-4',
    'OpenAI GPT-4 agent for general AI assistance',
    'https://api.openai.com/v1/chat/completions',
    'openai',
    '1.0',
    '{
        "chat": true,
        "tools": true,
        "streaming": true,
        "function_calling": true
    }'::jsonb,
    '{
        "model": "gpt-4",
        "max_tokens": 2000,
        "temperature": 0.7,
        "top_p": 1.0,
        "frequency_penalty": 0.0,
        "presence_penalty": 0.0
    }'::jsonb,
    'api_key',
    '{"example": true, "provider": "openai"}'::jsonb
) ON CONFLICT (organization_id, name) DO NOTHING;

-- Insert example Anthropic agent for testing
INSERT INTO a2a_agents (
    organization_id,
    name,
    description,
    endpoint_url,
    agent_type,
    protocol_version,
    capabilities,
    config,
    auth_type,
    metadata
) VALUES (
    '00000000-0000-0000-0000-000000000000',
    'Anthropic Claude',
    'Anthropic Claude agent for AI assistance',
    'https://api.anthropic.com/v1/messages',
    'anthropic',
    '1.0',
    '{
        "chat": true,
        "tools": true,
        "streaming": false,
        "function_calling": true
    }'::jsonb,
    '{
        "model": "claude-3-sonnet-20240229",
        "max_tokens": 4000,
        "temperature": 0.7
    }'::jsonb,
    'api_key',
    '{"example": true, "provider": "anthropic"}'::jsonb
) ON CONFLICT (organization_id, name) DO NOTHING;

-- Insert example custom agent for testing
INSERT INTO a2a_agents (
    organization_id,
    name,
    description,
    endpoint_url,
    agent_type,
    protocol_version,
    capabilities,
    config,
    auth_type,
    metadata
) VALUES (
    '00000000-0000-0000-0000-000000000000',
    'Custom Assistant',
    'Custom AI assistant agent',
    'https://api.example.com/v1/agent',
    'custom',
    '1.0',
    '{
        "chat": true,
        "tools": false,
        "streaming": true
    }'::jsonb,
    '{
        "timeout": 30,
        "retries": 3
    }'::jsonb,
    'bearer',
    '{"example": true, "provider": "custom"}'::jsonb
) ON CONFLICT (organization_id, name) DO NOTHING;
