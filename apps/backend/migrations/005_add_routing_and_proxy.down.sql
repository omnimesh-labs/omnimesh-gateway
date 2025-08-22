-- Rollback routing and proxy functionality including content filters

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS request_routing_log CASCADE;
DROP TABLE IF EXISTS proxy_routes CASCADE;
DROP TABLE IF EXISTS filter_violations CASCADE;
DROP TABLE IF EXISTS content_filters CASCADE;

-- Drop enums
DROP TYPE IF EXISTS filter_action_enum CASCADE;
DROP TYPE IF EXISTS filter_type_enum CASCADE;