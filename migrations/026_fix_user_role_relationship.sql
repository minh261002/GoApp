-- Fix user-role relationship by updating role_id values
-- This migration should run after roles table is created

-- Update existing users to have default role_id for 'user' role
-- Get the role_id for 'user' role dynamically
UPDATE users 
SET role_id = (
    SELECT id FROM roles 
    WHERE name = 'user' AND is_active = TRUE 
    LIMIT 1
)
WHERE role_id IS NULL;
