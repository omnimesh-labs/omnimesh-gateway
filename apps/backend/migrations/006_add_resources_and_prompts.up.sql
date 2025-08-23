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

-- Add some default resources and prompts for the default organization
INSERT INTO mcp_resources (organization_id, name, description, resource_type, uri, mime_type, created_by, metadata, tags) VALUES 
(
    '00000000-0000-0000-0000-000000000000',
    'System Documentation',
    'Default system documentation resource',
    'url',
    'https://docs.example.com/mcp-gateway',
    'text/html',
    NULL,
    '{"version": "1.0", "public": true}',
    ARRAY['documentation', 'system', 'help']
),
(
    '00000000-0000-0000-0000-000000000000', 
    'API Reference',
    'MCP Gateway API reference documentation',
    'url',
    'https://api.example.com/docs',
    'application/json',
    NULL,
    '{"version": "1.0", "swagger": true}',
    ARRAY['api', 'reference', 'documentation']
) ON CONFLICT (organization_id, name) DO NOTHING;

INSERT INTO mcp_prompts (organization_id, name, description, prompt_template, parameters, category, created_by, metadata, tags) VALUES 
(
    '00000000-0000-0000-0000-000000000000',
    'Code Review',
    'Template for conducting thorough code reviews',
    'Please review the following code for:\n1. Correctness and logic\n2. Performance considerations\n3. Security issues\n4. Code style and maintainability\n\nCode:\n{{code}}\n\nPlease provide specific feedback and suggestions for improvement.',
    '[{"name": "code", "type": "string", "required": true, "description": "Code to be reviewed"}]',
    'coding',
    NULL,
    '{"version": "1.0", "language_agnostic": true}',
    ARRAY['code-review', 'development', 'quality']
),
(
    '00000000-0000-0000-0000-000000000000',
    'Documentation Generation',
    'Generate comprehensive documentation from code',
    'Generate clear, comprehensive documentation for the following code. Include:\n1. Purpose and functionality\n2. Parameters and return values\n3. Usage examples\n4. Any important notes or limitations\n\nCode:\n{{code}}\n\nFormat: {{format}}\n\nPlease provide well-structured documentation.',
    '[{"name": "code", "type": "string", "required": true, "description": "Code to document"}, {"name": "format", "type": "string", "required": false, "description": "Documentation format (markdown, rst, etc.)", "default": "markdown"}]',
    'coding',
    NULL,
    '{"version": "1.0", "supports_multiple_formats": true}',
    ARRAY['documentation', 'code-generation', 'development']
),
(
    '00000000-0000-0000-0000-000000000000',
    'Data Analysis',
    'Template for structured data analysis tasks',
    'Analyze the following data and provide insights:\n\nData: {{data}}\n\nAnalysis Requirements:\n{{requirements}}\n\nPlease provide:\n1. Summary statistics\n2. Key patterns and trends\n3. Actionable insights\n4. Visualizations if applicable',
    '[{"name": "data", "type": "string", "required": true, "description": "Data to analyze"}, {"name": "requirements", "type": "string", "required": false, "description": "Specific analysis requirements", "default": "General analysis"}]',
    'analysis',
    NULL,
    '{"version": "1.0", "supports_visualizations": true}',
    ARRAY['data-analysis', 'statistics', 'insights']
) ON CONFLICT (organization_id, name) DO NOTHING;