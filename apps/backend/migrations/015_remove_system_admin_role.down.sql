-- Rollback: Re-add system_admin role option
-- Note: This doesn't restore any previously converted system_admin users

-- Update the check constraint to re-add system_admin
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('admin', 'user', 'viewer', 'api_user', 'system_admin'));
