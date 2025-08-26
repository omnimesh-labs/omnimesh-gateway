-- Remove key_type column and related changes from api_keys table

-- Drop indexes
DROP INDEX IF EXISTS idx_api_keys_key_type;
DROP INDEX IF EXISTS idx_api_keys_org_type;

-- Remove namespace_id column
ALTER TABLE api_keys DROP COLUMN IF EXISTS namespace_id;

-- Restore user_id NOT NULL constraint (but first ensure no NULL values exist)
-- This assumes all existing records have user_id populated
UPDATE api_keys SET user_id = '00000000-0000-0000-0000-000000000000' WHERE user_id IS NULL;
ALTER TABLE api_keys ALTER COLUMN user_id SET NOT NULL;

-- Remove key_type column
ALTER TABLE api_keys DROP COLUMN IF EXISTS key_type;