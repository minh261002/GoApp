-- Fix user-role relationship by updating role_id values
-- This migration should run after roles table is created

-- Update existing users to have default role_id = 1 (assuming 'user' role has id = 1)
-- Note: This assumes the 'user' role exists with id = 1
UPDATE users SET role_id = 1 WHERE role_id IS NULL;
