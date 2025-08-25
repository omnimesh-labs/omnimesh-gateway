-- Rollback MCP Tools Table
-- This migration removes the tools table

-- Drop table
DROP TABLE IF EXISTS mcp_tools CASCADE;

-- Drop enum
DROP TYPE IF EXISTS tool_category_enum CASCADE;
