-- Create audit_logs table
CREATE TABLE audit_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NULL,
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    resource_id BIGINT UNSIGNED NULL,
    resource_name VARCHAR(255),
    operation VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    message TEXT,
    ip_address VARCHAR(45),
    user_agent TEXT,
    referer VARCHAR(500),
    session_id VARCHAR(255),
    old_values JSON,
    new_values JSON,
    changes JSON,
    metadata JSON,
    tags VARCHAR(500),
    severity VARCHAR(20) DEFAULT 'info',
    target_user_id BIGINT UNSIGNED NULL,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_audit_logs_user_id (user_id),
    INDEX idx_audit_logs_action (action),
    INDEX idx_audit_logs_resource (resource),
    INDEX idx_audit_logs_resource_id (resource_id),
    INDEX idx_audit_logs_status (status),
    INDEX idx_audit_logs_severity (severity),
    INDEX idx_audit_logs_created_at (created_at),
    INDEX idx_audit_logs_ip_address (ip_address),
    INDEX idx_audit_logs_session_id (session_id),
    INDEX idx_audit_logs_target_user_id (target_user_id),
    INDEX idx_audit_logs_deleted_at (deleted_at),
    
    -- Composite indexes
    INDEX idx_audit_logs_user_action (user_id, action),
    INDEX idx_audit_logs_resource_action (resource, action),
    INDEX idx_audit_logs_status_created (status, created_at),
    INDEX idx_audit_logs_user_created (user_id, created_at),
    
    -- Foreign keys
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (target_user_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create audit_log_configs table
CREATE TABLE audit_log_configs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    is_enabled BOOLEAN DEFAULT TRUE,
    log_level VARCHAR(20) DEFAULT 'info',
    resources JSON,
    actions JSON,
    exclude_users JSON,
    retention_days INT DEFAULT 90,
    max_log_size INT DEFAULT 1000000,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_audit_log_configs_name (name),
    INDEX idx_audit_log_configs_is_enabled (is_enabled),
    INDEX idx_audit_log_configs_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create audit_log_summaries table
CREATE TABLE audit_log_summaries (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    date DATE NOT NULL,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,
    total_count BIGINT DEFAULT 0,
    success_count BIGINT DEFAULT 0,
    failure_count BIGINT DEFAULT 0,
    error_count BIGINT DEFAULT 0,
    unique_users BIGINT DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_audit_log_summaries_date (date),
    INDEX idx_audit_log_summaries_resource (resource),
    INDEX idx_audit_log_summaries_action (action),
    INDEX idx_audit_log_summaries_status (status),
    INDEX idx_audit_log_summaries_deleted_at (deleted_at),
    
    -- Composite indexes
    INDEX idx_audit_log_summaries_date_resource (date, resource),
    INDEX idx_audit_log_summaries_date_action (date, action),
    INDEX idx_audit_log_summaries_resource_action (resource, action),
    
    -- Unique constraint
    UNIQUE KEY unique_date_resource_action_status (date, resource, action, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default audit log configuration
INSERT INTO audit_log_configs (name, description, is_enabled, log_level, resources, actions, exclude_users, retention_days, max_log_size, created_at, updated_at) VALUES
('default', 'Default audit log configuration', TRUE, 'info', 
 '["user", "order", "product", "category", "brand", "inventory", "permission", "role", "system"]',
 '["create", "read", "update", "delete", "login", "logout", "register", "password_change", "permission_grant", "permission_revoke"]',
 '[]',
 90, 1000000, NOW(), NOW());

-- Insert default audit log summaries for the past 30 days
INSERT INTO audit_log_summaries (date, resource, action, status, total_count, success_count, failure_count, error_count, unique_users, created_at, updated_at)
SELECT 
    DATE_SUB(CURDATE(), INTERVAL n DAY) as date,
    'system' as resource,
    'login' as action,
    'success' as status,
    0 as total_count,
    0 as success_count,
    0 as failure_count,
    0 as error_count,
    0 as unique_users,
    NOW() as created_at,
    NOW() as updated_at
FROM (
    SELECT 0 as n UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION
    SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9 UNION
    SELECT 10 UNION SELECT 11 UNION SELECT 12 UNION SELECT 13 UNION SELECT 14 UNION
    SELECT 15 UNION SELECT 16 UNION SELECT 17 UNION SELECT 18 UNION SELECT 19 UNION
    SELECT 20 UNION SELECT 21 UNION SELECT 22 UNION SELECT 23 UNION SELECT 24 UNION
    SELECT 25 UNION SELECT 26 UNION SELECT 27 UNION SELECT 28 UNION SELECT 29
) as numbers;
