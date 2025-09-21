-- Create otps table
CREATE TABLE IF NOT EXISTS otps (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    email VARCHAR(100) NOT NULL,
    code VARCHAR(10) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('password_reset', 'email_verify')),
    is_used BOOLEAN DEFAULT FALSE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_otps_user_id (user_id),
    INDEX idx_otps_email (email),
    INDEX idx_otps_code (code),
    INDEX idx_otps_type (type),
    INDEX idx_otps_expires_at (expires_at),
    INDEX idx_otps_deleted_at (deleted_at)
);
