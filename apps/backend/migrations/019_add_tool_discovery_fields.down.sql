-- Rollback: Remove tool discovery fields from mcp_tools table
DROP INDEX IF EXISTS idx_mcp_tools_server;
DROP INDEX IF EXISTS idx_mcp_tools_source;

ALTER TABLE mcp_tools
DROP COLUMN IF EXISTS server_id,
DROP COLUMN IF EXISTS source_type,
DROP COLUMN IF EXISTS last_discovered_at,
DROP COLUMN IF EXISTS discovery_metadata;
