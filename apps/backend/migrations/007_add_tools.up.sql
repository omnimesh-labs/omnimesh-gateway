-- Add MCP Tools Table
-- This migration adds support for globally available MCP tools

-- Create tool category enum
CREATE TYPE tool_category_enum AS ENUM ('general', 'data', 'file', 'web', 'system', 'ai', 'dev', 'custom');

-- MCP Tools - Globally available tools
CREATE TABLE mcp_tools (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Tool specification
    function_name VARCHAR(255) NOT NULL, -- The actual function/method name to call
    schema JSONB NOT NULL DEFAULT '{}', -- JSON schema for tool parameters/inputs
    category tool_category_enum DEFAULT 'general',

    -- Implementation details
    implementation_type VARCHAR(50) DEFAULT 'internal', -- 'internal', 'external', 'webhook', 'script'
    endpoint_url VARCHAR(2000), -- For webhook/external tools
    timeout_seconds INTEGER DEFAULT 30,
    max_retries INTEGER DEFAULT 3,

    -- Usage and popularity tracking
    usage_count BIGINT DEFAULT 0,

    -- Access control
    access_permissions JSONB DEFAULT '{"execute": ["*"], "read": ["*"]}',

    -- Status and metadata
    is_active BOOLEAN DEFAULT true,
    is_public BOOLEAN DEFAULT false, -- Whether tool can be used by other orgs
    metadata JSONB DEFAULT '{}',
    tags TEXT[],

    -- Examples and documentation
    examples JSONB DEFAULT '[]', -- Example usage/calls
    documentation TEXT, -- Extended documentation/help text

    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),

    -- Constraints
    UNIQUE(organization_id, name),
    UNIQUE(organization_id, function_name),
    CONSTRAINT valid_function_name CHECK (LENGTH(function_name) > 0),
    CONSTRAINT valid_timeout CHECK (timeout_seconds > 0 AND timeout_seconds <= 600),
    CONSTRAINT valid_max_retries CHECK (max_retries >= 0 AND max_retries <= 10),
    CONSTRAINT valid_usage_count CHECK (usage_count >= 0)
);

-- Add updated_at trigger
CREATE TRIGGER mcp_tools_updated_at
    BEFORE UPDATE ON mcp_tools
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Performance indexes for mcp_tools
CREATE INDEX idx_mcp_tools_org_active ON mcp_tools(organization_id, is_active);
CREATE INDEX idx_mcp_tools_org_category ON mcp_tools(organization_id, category);
CREATE INDEX idx_mcp_tools_function_name ON mcp_tools(organization_id, function_name);
CREATE INDEX idx_mcp_tools_tags ON mcp_tools USING gin(tags);
CREATE INDEX idx_mcp_tools_usage_count ON mcp_tools(usage_count DESC);
CREATE INDEX idx_mcp_tools_public ON mcp_tools(is_public) WHERE is_public = true;
CREATE INDEX idx_mcp_tools_created_at ON mcp_tools(created_at DESC);
CREATE INDEX idx_mcp_tools_name ON mcp_tools(organization_id, name);
