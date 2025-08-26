-- Remove system_admin role and convert to admin
-- This migration updates any existing system_admin users to admin role

-- Update any existing system_admin users to admin
UPDATE users SET role = 'admin' WHERE role = 'system_admin';

-- Update the check constraint to remove system_admin
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('admin', 'user', 'viewer', 'api_user'));
