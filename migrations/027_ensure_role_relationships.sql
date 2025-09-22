-- Ensure all users have proper role_id relationships
-- This migration ensures no users have NULL role_id

-- First, ensure we have the default 'user' role
INSERT IGNORE INTO roles (name, display_name, description, is_system, is_active) 
VALUES ('user', 'User', 'Basic user access', TRUE, TRUE);

-- Update any users with NULL role_id to have the 'user' role
UPDATE users 
SET role_id = (
    SELECT id FROM roles 
    WHERE name = 'user' AND is_active = TRUE 
    LIMIT 1
)
WHERE role_id IS NULL;

-- Add constraint to ensure role_id is never NULL
ALTER TABLE users MODIFY COLUMN role_id BIGINT UNSIGNED NOT NULL;

-- Add foreign key constraint if it doesn't exist
-- (This might already exist from the original migration)
-- ALTER TABLE users ADD CONSTRAINT fk_users_role_id FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE RESTRICT ON UPDATE CASCADE;
