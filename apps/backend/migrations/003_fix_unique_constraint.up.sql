-- Drop the existing unique constraint
ALTER TABLE mcp_servers DROP CONSTRAINT IF EXISTS mcp_servers_organization_id_name_key;

-- Create a partial unique index that only applies to active servers
CREATE UNIQUE INDEX mcp_servers_organization_id_name_active_key
ON mcp_servers (organization_id, name)
WHERE is_active = true;
