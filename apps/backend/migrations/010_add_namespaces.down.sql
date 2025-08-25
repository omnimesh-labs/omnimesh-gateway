-- Drop indexes
DROP INDEX IF EXISTS idx_namespace_tool_mappings_server;
DROP INDEX IF EXISTS idx_namespace_tool_mappings;
DROP INDEX IF EXISTS idx_namespace_mappings_server;
DROP INDEX IF EXISTS idx_namespace_mappings_namespace;
DROP INDEX IF EXISTS idx_namespaces_active;
DROP INDEX IF EXISTS idx_namespaces_org;

-- Drop tables
DROP TABLE IF EXISTS namespace_tool_mappings;
DROP TABLE IF EXISTS namespace_server_mappings;
DROP TABLE IF EXISTS namespaces;