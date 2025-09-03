-- Rollback OAuth 2.0 Support Migration

-- Drop view
DROP VIEW IF EXISTS active_oauth_tokens;

-- Remove OAuth columns from existing tables
ALTER TABLE api_keys DROP COLUMN IF EXISTS oauth_scopes;
ALTER TABLE api_keys DROP COLUMN IF EXISTS oauth_client_id;

ALTER TABLE endpoints DROP COLUMN IF EXISTS require_oauth;
ALTER TABLE endpoints DROP COLUMN IF EXISTS oauth_scopes;

ALTER TABLE users DROP COLUMN IF EXISTS oauth_scopes;

ALTER TABLE organizations DROP COLUMN IF EXISTS oauth_default_scope;
ALTER TABLE organizations DROP COLUMN IF EXISTS oauth_require_consent;
ALTER TABLE organizations DROP COLUMN IF EXISTS oauth_enabled;

-- Drop OAuth-specific tables (in reverse order due to foreign keys)
DROP TABLE IF EXISTS oauth_user_consents;
DROP TABLE IF EXISTS oauth_authorization_codes;
DROP TABLE IF EXISTS oauth_tokens;
DROP TABLE IF EXISTS oauth_clients;