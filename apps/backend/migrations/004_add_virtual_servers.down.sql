-- Rollback virtual servers migration

-- Remove virtual server reference from audit logs
DROP INDEX IF EXISTS idx_audit_logs_virtual_server;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS virtual_server_id;

-- Drop virtual servers table and related objects
DROP TABLE IF EXISTS virtual_servers CASCADE;
DROP TYPE IF EXISTS adapter_type_enum CASCADE;
