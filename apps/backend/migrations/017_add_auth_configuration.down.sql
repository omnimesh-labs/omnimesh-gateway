-- Rollback auth configuration tables

-- Drop triggers first
DROP TRIGGER IF EXISTS auth_configurations_updated_at ON auth_configurations;
DROP TRIGGER IF EXISTS session_configurations_updated_at ON session_configurations;
DROP TRIGGER IF EXISTS security_policies_updated_at ON security_policies;

-- Drop indexes
DROP INDEX IF EXISTS idx_auth_configurations_org_id;
DROP INDEX IF EXISTS idx_session_configurations_org_id;
DROP INDEX IF EXISTS idx_security_policies_org_id;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS security_policies CASCADE;
DROP TABLE IF EXISTS session_configurations CASCADE;
DROP TABLE IF EXISTS auth_configurations CASCADE;