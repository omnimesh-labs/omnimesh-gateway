-- Drop Configuration Import History Tables
-- This migration removes configuration import/export tracking

-- Drop tables
DROP TABLE IF EXISTS config_exports CASCADE;
DROP TABLE IF EXISTS config_imports CASCADE;

-- Drop enums
DROP TYPE IF EXISTS conflict_strategy_enum;
DROP TYPE IF EXISTS import_status_enum;