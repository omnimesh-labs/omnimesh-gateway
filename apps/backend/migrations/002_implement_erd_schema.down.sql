-- Drop ERD schema implementation

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS log_aggregates CASCADE;
DROP TABLE IF EXISTS rate_limit_usage CASCADE;
DROP TABLE IF EXISTS rate_limits CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS log_index CASCADE;
DROP TABLE IF EXISTS server_stats CASCADE;
DROP TABLE IF EXISTS health_checks CASCADE;
DROP TABLE IF EXISTS mcp_sessions CASCADE;
DROP TABLE IF EXISTS policies CASCADE;
DROP TABLE IF EXISTS api_keys CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS mcp_servers CASCADE;
DROP TABLE IF EXISTS organizations CASCADE;

-- Drop views
DROP VIEW IF EXISTS server_health_summary CASCADE;
DROP VIEW IF EXISTS active_sessions CASCADE;

-- Drop types
DROP TYPE IF EXISTS aggregation_window_enum CASCADE;
DROP TYPE IF EXISTS rate_limit_scope_enum CASCADE;
DROP TYPE IF EXISTS plan_type_enum CASCADE;
DROP TYPE IF EXISTS log_level_enum CASCADE;
DROP TYPE IF EXISTS storage_provider_enum CASCADE;
DROP TYPE IF EXISTS health_status_enum CASCADE;
DROP TYPE IF EXISTS proc_status_enum CASCADE;
DROP TYPE IF EXISTS session_status_enum CASCADE;
DROP TYPE IF EXISTS server_status_enum CASCADE;
DROP TYPE IF EXISTS protocol_enum CASCADE;

-- Drop functions
DROP FUNCTION IF EXISTS set_updated_at() CASCADE;
