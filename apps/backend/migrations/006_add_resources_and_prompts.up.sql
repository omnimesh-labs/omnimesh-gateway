-- Add Resources and Prompts Tables
-- This migration adds support for globally available MCP resources and prompts

-- Create resource type enum
CREATE TYPE resource_type_enum AS ENUM ('file', 'url', 'database', 'api', 'memory', 'custom');

-- Create prompt category enum
CREATE TYPE prompt_category_enum AS ENUM ('general', 'coding', 'analysis', 'creative', 'educational', 'business', 'custom');

-- MCP Resources - Globally available resources
CREATE TABLE mcp_resources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    resource_type resource_type_enum NOT NULL DEFAULT 'custom',

    -- Resource location and access
    uri VARCHAR(2000) NOT NULL, -- Resource URI/path/endpoint
    mime_type VARCHAR(255),     -- MIME type for files
    size_bytes BIGINT,          -- Size in bytes (for files)

    -- Access control
    access_permissions JSONB DEFAULT '{"read": ["*"], "write": ["admin"]}',

    -- Status and metadata
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    tags TEXT[],

    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),

    -- Constraints
    UNIQUE(organization_id, name),
    CONSTRAINT valid_uri_length CHECK (LENGTH(uri) > 0),
    CONSTRAINT valid_size_bytes CHECK (size_bytes IS NULL OR size_bytes >= 0)
);

-- MCP Prompts - Globally available prompts/templates
CREATE TABLE mcp_prompts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Prompt content and structure
    prompt_template TEXT NOT NULL, -- The actual prompt content
    parameters JSONB DEFAULT '[]', -- Expected parameters: [{"name": "input", "type": "string", "required": true}]
    category prompt_category_enum DEFAULT 'general',

    -- Usage and popularity tracking
    usage_count BIGINT DEFAULT 0,

    -- Status and metadata
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    tags TEXT[],

    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),

    -- Constraints
    UNIQUE(organization_id, name),
    CONSTRAINT valid_prompt_template CHECK (LENGTH(prompt_template) > 0),
    CONSTRAINT valid_usage_count CHECK (usage_count >= 0)
);

-- Add updated_at triggers
CREATE TRIGGER mcp_resources_updated_at
    BEFORE UPDATE ON mcp_resources
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER mcp_prompts_updated_at
    BEFORE UPDATE ON mcp_prompts
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Performance indexes for mcp_resources
CREATE INDEX idx_mcp_resources_org_active ON mcp_resources(organization_id, is_active);
CREATE INDEX idx_mcp_resources_org_type ON mcp_resources(organization_id, resource_type);
CREATE INDEX idx_mcp_resources_tags ON mcp_resources USING gin(tags);
CREATE INDEX idx_mcp_resources_created_at ON mcp_resources(created_at DESC);
CREATE INDEX idx_mcp_resources_name ON mcp_resources(organization_id, name);

-- Performance indexes for mcp_prompts
CREATE INDEX idx_mcp_prompts_org_active ON mcp_prompts(organization_id, is_active);
CREATE INDEX idx_mcp_prompts_org_category ON mcp_prompts(organization_id, category);
CREATE INDEX idx_mcp_prompts_tags ON mcp_prompts USING gin(tags);
CREATE INDEX idx_mcp_prompts_usage_count ON mcp_prompts(usage_count DESC);
CREATE INDEX idx_mcp_prompts_created_at ON mcp_prompts(created_at DESC);
CREATE INDEX idx_mcp_prompts_name ON mcp_prompts(organization_id, name);
