-- Rollback Resources and Prompts Tables
-- This migration removes the resources and prompts tables

-- Drop tables (in reverse dependency order)
DROP TABLE IF EXISTS mcp_prompts CASCADE;
DROP TABLE IF EXISTS mcp_resources CASCADE;

-- Drop enums
DROP TYPE IF EXISTS resource_type_enum CASCADE;
DROP TYPE IF EXISTS prompt_category_enum CASCADE;
