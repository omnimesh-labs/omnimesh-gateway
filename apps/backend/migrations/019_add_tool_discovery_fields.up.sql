-- Migration: Add tool discovery fields to mcp_tools table
ALTER TABLE mcp_tools
ADD COLUMN server_id UUID REFERENCES mcp_servers(id) ON DELETE SET NULL,
ADD COLUMN source_type VARCHAR(20) DEFAULT 'manual' CHECK (source_type IN ('manual', 'discovered')),
ADD COLUMN last_discovered_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN discovery_metadata JSONB DEFAULT '{}';

-- Create index for server-based tool queries
CREATE INDEX idx_mcp_tools_server ON mcp_tools(server_id);
CREATE INDEX idx_mcp_tools_source ON mcp_tools(source_type);

-- Update existing tools to have manual source type
UPDATE mcp_tools SET source_type = 'manual' WHERE source_type IS NULL;
