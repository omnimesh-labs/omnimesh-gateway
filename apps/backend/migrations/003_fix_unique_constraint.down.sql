-- Drop the partial unique index
DROP INDEX IF EXISTS mcp_servers_organization_id_name_active_key;

-- Restore the original unique constraint (this will fail if there are duplicate names)
-- Note: You may need to clean up duplicate names before running this rollback
ALTER TABLE mcp_servers ADD CONSTRAINT mcp_servers_organization_id_name_key
UNIQUE (organization_id, name);
