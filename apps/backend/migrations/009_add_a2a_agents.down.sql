-- Remove A2A (Agent-to-Agent) integration support
-- This rollback migration removes all A2A agent related tables and types

-- Remove foreign key column from audit_logs
DROP INDEX IF EXISTS idx_audit_logs_a2a_agent;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS a2a_agent_id;

-- Drop A2A agent tools table
DROP TRIGGER IF EXISTS a2a_agent_tools_updated_at ON a2a_agent_tools;
DROP INDEX IF EXISTS idx_a2a_agent_tools_agent;
DROP INDEX IF EXISTS idx_a2a_agent_tools_virtual_server;
DROP INDEX IF EXISTS idx_a2a_agent_tools_active;
DROP TABLE IF EXISTS a2a_agent_tools;

-- Drop A2A agents table
DROP TRIGGER IF EXISTS a2a_agents_updated_at ON a2a_agents;
DROP INDEX IF EXISTS idx_a2a_agents_org_active;
DROP INDEX IF EXISTS idx_a2a_agents_agent_type;
DROP INDEX IF EXISTS idx_a2a_agents_name;
DROP INDEX IF EXISTS idx_a2a_agents_health;
DROP INDEX IF EXISTS idx_a2a_agents_tags;
DROP TABLE IF EXISTS a2a_agents;

-- Drop custom types
DROP TYPE IF EXISTS auth_type_enum;
DROP TYPE IF EXISTS agent_type_enum;