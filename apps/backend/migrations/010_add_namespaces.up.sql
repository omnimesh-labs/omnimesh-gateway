-- Migration: Create namespaces and related tables
CREATE TABLE IF NOT EXISTS namespaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    UNIQUE(organization_id, name),
    CONSTRAINT namespaces_name_regex CHECK (name ~ '^[a-zA-Z0-9_-]+$')
);

-- Create namespace server mappings table
CREATE TABLE IF NOT EXISTS namespace_server_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    namespace_id UUID NOT NULL REFERENCES namespaces(id) ON DELETE CASCADE,
    server_id UUID NOT NULL REFERENCES mcp_servers(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE')),
    priority INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(namespace_id, server_id)
);

-- Create namespace tool mappings table
CREATE TABLE IF NOT EXISTS namespace_tool_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    namespace_id UUID NOT NULL REFERENCES namespaces(id) ON DELETE CASCADE,
    server_id UUID NOT NULL REFERENCES mcp_servers(id) ON DELETE CASCADE,
    tool_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(namespace_id, server_id, tool_name)
);

-- Add indexes for performance
CREATE INDEX idx_namespaces_org ON namespaces(organization_id);
CREATE INDEX idx_namespaces_active ON namespaces(is_active);
CREATE INDEX idx_namespace_mappings_namespace ON namespace_server_mappings(namespace_id);
CREATE INDEX idx_namespace_mappings_server ON namespace_server_mappings(server_id);
CREATE INDEX idx_namespace_tool_mappings ON namespace_tool_mappings(namespace_id, status);
CREATE INDEX idx_namespace_tool_mappings_server ON namespace_tool_mappings(server_id);