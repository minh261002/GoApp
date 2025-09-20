-- +migrate Up
CREATE TABLE IF NOT EXISTS permissions (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    is_system BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_permissions_resource (resource),
    INDEX idx_permissions_action (action),
    INDEX idx_permissions_is_active (is_active),
    INDEX idx_permissions_deleted_at (deleted_at),
    UNIQUE KEY unique_permission_resource_action (resource, action)
);

CREATE TABLE IF NOT EXISTS roles (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_roles_is_active (is_active),
    INDEX idx_roles_deleted_at (deleted_at)
);

CREATE TABLE IF NOT EXISTS role_permissions (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    role_id INT UNSIGNED NOT NULL,
    permission_id INT UNSIGNED NOT NULL,
    granted_by INT UNSIGNED NOT NULL,
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE,
    FOREIGN KEY (granted_by) REFERENCES users(id) ON DELETE CASCADE,
    
    UNIQUE KEY unique_role_permission (role_id, permission_id),
    INDEX idx_role_permissions_role_id (role_id),
    INDEX idx_role_permissions_permission_id (permission_id),
    INDEX idx_role_permissions_granted_by (granted_by),
    INDEX idx_role_permissions_deleted_at (deleted_at)
);

CREATE TABLE IF NOT EXISTS user_permissions (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL,
    permission_id INT UNSIGNED NOT NULL,
    is_granted BOOLEAN DEFAULT TRUE,
    granted_by INT UNSIGNED NOT NULL,
    reason TEXT,
    expires_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE,
    FOREIGN KEY (granted_by) REFERENCES users(id) ON DELETE CASCADE,
    
    UNIQUE KEY unique_user_permission (user_id, permission_id),
    INDEX idx_user_permissions_user_id (user_id),
    INDEX idx_user_permissions_permission_id (permission_id),
    INDEX idx_user_permissions_granted_by (granted_by),
    INDEX idx_user_permissions_expires_at (expires_at),
    INDEX idx_user_permissions_deleted_at (deleted_at)
);

CREATE TABLE IF NOT EXISTS permission_logs (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL,
    target_user_id INT UNSIGNED NULL,
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id INT UNSIGNED NOT NULL,
    details TEXT,
    ip_address VARCHAR(45),
    user_agent VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (target_user_id) REFERENCES users(id) ON DELETE SET NULL,
    
    INDEX idx_permission_logs_user_id (user_id),
    INDEX idx_permission_logs_target_user_id (target_user_id),
    INDEX idx_permission_logs_action (action),
    INDEX idx_permission_logs_resource_type (resource_type),
    INDEX idx_permission_logs_created_at (created_at)
);

-- Insert default system roles
INSERT INTO roles (name, display_name, description, is_system, is_active) VALUES
('super_admin', 'Super Administrator', 'Full system access with all permissions', TRUE, TRUE),
('admin', 'Administrator', 'Administrative access to most system features', TRUE, TRUE),
('moderator', 'Moderator', 'Moderation access to content and users', TRUE, TRUE),
('user', 'User', 'Basic user access', TRUE, TRUE),
('guest', 'Guest', 'Limited read-only access', TRUE, TRUE);

-- Insert default system permissions
INSERT INTO permissions (name, display_name, description, resource, action, is_system, is_active) VALUES
-- User permissions
('user.read', 'Read Users', 'View user information', 'user', 'read', TRUE, TRUE),
('user.write', 'Create/Update Users', 'Create and update user information', 'user', 'write', TRUE, TRUE),
('user.delete', 'Delete Users', 'Delete user accounts', 'user', 'delete', TRUE, TRUE),
('user.manage', 'Manage Users', 'Full user management including roles and permissions', 'user', 'manage', TRUE, TRUE),

-- Brand permissions
('brand.read', 'Read Brands', 'View brand information', 'brand', 'read', TRUE, TRUE),
('brand.write', 'Create/Update Brands', 'Create and update brand information', 'brand', 'write', TRUE, TRUE),
('brand.delete', 'Delete Brands', 'Delete brand records', 'brand', 'delete', TRUE, TRUE),
('brand.manage', 'Manage Brands', 'Full brand management', 'brand', 'manage', TRUE, TRUE),

-- Category permissions
('category.read', 'Read Categories', 'View category information', 'category', 'read', TRUE, TRUE),
('category.write', 'Create/Update Categories', 'Create and update category information', 'category', 'write', TRUE, TRUE),
('category.delete', 'Delete Categories', 'Delete category records', 'category', 'delete', TRUE, TRUE),
('category.manage', 'Manage Categories', 'Full category management', 'category', 'manage', TRUE, TRUE),

-- Product permissions
('product.read', 'Read Products', 'View product information', 'product', 'read', TRUE, TRUE),
('product.write', 'Create/Update Products', 'Create and update product information', 'product', 'write', TRUE, TRUE),
('product.delete', 'Delete Products', 'Delete product records', 'product', 'delete', TRUE, TRUE),
('product.manage', 'Manage Products', 'Full product management', 'product', 'manage', TRUE, TRUE),

-- Inventory permissions
('inventory.read', 'Read Inventory', 'View inventory information', 'inventory', 'read', TRUE, TRUE),
('inventory.write', 'Create/Update Inventory', 'Create and update inventory records', 'inventory', 'write', TRUE, TRUE),
('inventory.delete', 'Delete Inventory', 'Delete inventory records', 'inventory', 'delete', TRUE, TRUE),
('inventory.manage', 'Manage Inventory', 'Full inventory management', 'inventory', 'manage', TRUE, TRUE),

-- Upload permissions
('upload.read', 'Read Uploads', 'View uploaded files', 'upload', 'read', TRUE, TRUE),
('upload.write', 'Upload Files', 'Upload new files', 'upload', 'write', TRUE, TRUE),
('upload.delete', 'Delete Uploads', 'Delete uploaded files', 'upload', 'delete', TRUE, TRUE),
('upload.manage', 'Manage Uploads', 'Full upload management', 'upload', 'manage', TRUE, TRUE),

-- Order permissions
('order.read', 'Read Orders', 'View order information', 'order', 'read', TRUE, TRUE),
('order.write', 'Create/Update Orders', 'Create and update orders', 'order', 'write', TRUE, TRUE),
('order.delete', 'Delete Orders', 'Delete order records', 'order', 'delete', TRUE, TRUE),
('order.manage', 'Manage Orders', 'Full order management', 'order', 'manage', TRUE, TRUE),

-- Customer permissions
('customer.read', 'Read Customers', 'View customer information', 'customer', 'read', TRUE, TRUE),
('customer.write', 'Create/Update Customers', 'Create and update customer information', 'customer', 'write', TRUE, TRUE),
('customer.delete', 'Delete Customers', 'Delete customer records', 'customer', 'delete', TRUE, TRUE),
('customer.manage', 'Manage Customers', 'Full customer management', 'customer', 'manage', TRUE, TRUE),

-- Report permissions
('report.read', 'Read Reports', 'View reports and analytics', 'report', 'read', TRUE, TRUE),
('report.write', 'Create Reports', 'Create custom reports', 'report', 'write', TRUE, TRUE),
('report.manage', 'Manage Reports', 'Full report management', 'report', 'manage', TRUE, TRUE),

-- System permissions
('system.read', 'Read System', 'View system information', 'system', 'read', TRUE, TRUE),
('system.write', 'Configure System', 'Configure system settings', 'system', 'write', TRUE, TRUE),
('system.manage', 'Manage System', 'Full system management', 'system', 'manage', TRUE, TRUE),
('system.admin', 'System Administration', 'Complete system administration access', 'system', 'admin', TRUE, TRUE);

-- Assign permissions to roles
-- Super Admin gets all permissions
INSERT INTO role_permissions (role_id, permission_id, granted_by)
SELECT r.id, p.id, 1
FROM roles r, permissions p
WHERE r.name = 'super_admin';

-- Admin gets most permissions except system.admin
INSERT INTO role_permissions (role_id, permission_id, granted_by)
SELECT r.id, p.id, 1
FROM roles r, permissions p
WHERE r.name = 'admin' AND p.action != 'admin';

-- Moderator gets read/write permissions for content
INSERT INTO role_permissions (role_id, permission_id, granted_by)
SELECT r.id, p.id, 1
FROM roles r, permissions p
WHERE r.name = 'moderator' 
AND p.action IN ('read', 'write')
AND p.resource IN ('user', 'brand', 'category', 'product', 'inventory', 'upload', 'order', 'customer');

-- User gets read permissions for most resources
INSERT INTO role_permissions (role_id, permission_id, granted_by)
SELECT r.id, p.id, 1
FROM roles r, permissions p
WHERE r.name = 'user' 
AND p.action = 'read'
AND p.resource IN ('brand', 'category', 'product', 'inventory', 'upload');

-- Guest gets read permissions for public content
INSERT INTO role_permissions (role_id, permission_id, granted_by)
SELECT r.id, p.id, 1
FROM roles r, permissions p
WHERE r.name = 'guest' 
AND p.action = 'read'
AND p.resource IN ('brand', 'category', 'product', 'inventory');

-- +migrate Down
DROP TABLE IF EXISTS permission_logs;
DROP TABLE IF EXISTS user_permissions;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
