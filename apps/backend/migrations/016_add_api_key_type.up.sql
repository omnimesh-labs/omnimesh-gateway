-- Add key_type column to api_keys table
-- This column is needed to distinguish between user API keys and app-to-app authentication keys

ALTER TABLE api_keys
ADD COLUMN key_type VARCHAR(20) NOT NULL DEFAULT 'user'
CHECK (key_type IN ('user', 'a2a'));

-- Update the user_id constraint to allow NULL for A2A keys
ALTER TABLE api_keys
ALTER COLUMN user_id DROP NOT NULL;

-- Add namespace_id column for optional namespace scope
ALTER TABLE api_keys
ADD COLUMN namespace_id UUID REFERENCES namespaces(id) ON DELETE CASCADE;

-- Add index for key_type
CREATE INDEX idx_api_keys_key_type ON api_keys(key_type);

-- Add composite index for organization and key type
CREATE INDEX idx_api_keys_org_type ON api_keys(organization_id, key_type);