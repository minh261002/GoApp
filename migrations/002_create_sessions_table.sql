-- Create sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    device_id VARCHAR(100) NOT NULL,
    user_agent VARCHAR(500),
    ip_address VARCHAR(45),
    is_active BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_sessions_user_id (user_id),
    INDEX idx_sessions_token (token),
    INDEX idx_sessions_device_id (device_id),
    INDEX idx_sessions_expires_at (expires_at),
    INDEX idx_sessions_deleted_at (deleted_at)
);
