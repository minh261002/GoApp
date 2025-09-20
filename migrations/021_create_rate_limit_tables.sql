-- Create rate_limit_rules table
CREATE TABLE rate_limit_rules (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    requests INT NOT NULL,
    window INT NOT NULL,
    window_type VARCHAR(20) NOT NULL,
    target_type VARCHAR(20) NOT NULL,
    target_value VARCHAR(255),
    scope VARCHAR(50) NOT NULL,
    scope_value VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    priority INT DEFAULT 0,
    error_code INT DEFAULT 429,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_name (name),
    INDEX idx_target_type (target_type),
    INDEX idx_scope (scope),
    INDEX idx_is_active (is_active),
    INDEX idx_priority (priority),
    INDEX idx_deleted_at (deleted_at)
);

-- Create rate_limit_logs table
CREATE TABLE rate_limit_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    rule_id BIGINT UNSIGNED NOT NULL,
    client_ip VARCHAR(45) NOT NULL,
    user_id BIGINT UNSIGNED,
    api_key VARCHAR(255),
    method VARCHAR(10) NOT NULL,
    path VARCHAR(500) NOT NULL,
    user_agent VARCHAR(500),
    referer VARCHAR(500),
    limit_count INT NOT NULL,
    current_count INT NOT NULL,
    remaining_count INT NOT NULL,
    reset_time TIMESTAMP NOT NULL,
    violation_type VARCHAR(50) NOT NULL,
    is_blocked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (rule_id) REFERENCES rate_limit_rules(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_rule_id (rule_id),
    INDEX idx_client_ip (client_ip),
    INDEX idx_user_id (user_id),
    INDEX idx_api_key (api_key),
    INDEX idx_violation_type (violation_type),
    INDEX idx_is_blocked (is_blocked),
    INDEX idx_created_at (created_at)
);

-- Create rate_limit_stats table
CREATE TABLE rate_limit_stats (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    rule_id BIGINT UNSIGNED NOT NULL,
    period VARCHAR(20) NOT NULL,
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,
    total_requests BIGINT NOT NULL,
    blocked_requests BIGINT NOT NULL,
    unique_clients BIGINT NOT NULL,
    average_requests DECIMAL(10,2),
    peak_requests BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (rule_id) REFERENCES rate_limit_rules(id) ON DELETE CASCADE,
    INDEX idx_rule_id (rule_id),
    INDEX idx_period (period),
    INDEX idx_period_start (period_start),
    INDEX idx_period_end (period_end),
    INDEX idx_created_at (created_at)
);

-- Create rate_limit_whitelist table
CREATE TABLE rate_limit_whitelist (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    type VARCHAR(20) NOT NULL,
    value VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_type (type),
    INDEX idx_value (value),
    INDEX idx_is_active (is_active),
    INDEX idx_deleted_at (deleted_at)
);

-- Create rate_limit_blacklist table
CREATE TABLE rate_limit_blacklist (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    type VARCHAR(20) NOT NULL,
    value VARCHAR(255) NOT NULL,
    reason TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_type (type),
    INDEX idx_value (value),
    INDEX idx_is_active (is_active),
    INDEX idx_expires_at (expires_at),
    INDEX idx_deleted_at (deleted_at)
);

-- Create rate_limit_configs table
CREATE TABLE rate_limit_configs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    is_enabled BOOLEAN DEFAULT TRUE,
    default_rule BIGINT UNSIGNED NOT NULL,
    redis_host VARCHAR(255) NOT NULL,
    redis_port INT NOT NULL,
    redis_db INT DEFAULT 0,
    redis_password VARCHAR(255),
    log_retention_days INT DEFAULT 30,
    stats_retention_days INT DEFAULT 90,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (default_rule) REFERENCES rate_limit_rules(id) ON DELETE RESTRICT,
    INDEX idx_name (name),
    INDEX idx_is_enabled (is_enabled),
    INDEX idx_deleted_at (deleted_at)
);

-- Insert default rate limit rules
INSERT INTO rate_limit_rules (name, description, requests, window, window_type, target_type, scope, is_active, priority, error_code, error_message) VALUES
-- Global rate limits
('global_ip_limit', 'Global IP rate limit - 100 requests per hour', 100, 1, 'hour', 'ip', 'global', TRUE, 100, 429, 'Too many requests from this IP address'),
('global_user_limit', 'Global user rate limit - 1000 requests per hour', 1000, 1, 'hour', 'user', 'global', TRUE, 90, 429, 'Too many requests from this user'),
('global_api_key_limit', 'Global API key rate limit - 5000 requests per hour', 5000, 1, 'hour', 'api_key', 'global', TRUE, 80, 429, 'Too many requests with this API key'),

-- Endpoint-specific rate limits
('auth_login_limit', 'Login endpoint rate limit - 5 requests per minute', 5, 1, 'minute', 'ip', 'endpoint', TRUE, 200, 429, 'Too many login attempts'),
('auth_register_limit', 'Register endpoint rate limit - 3 requests per minute', 3, 1, 'minute', 'ip', 'endpoint', TRUE, 190, 429, 'Too many registration attempts'),
('password_reset_limit', 'Password reset rate limit - 3 requests per hour', 3, 1, 'hour', 'ip', 'endpoint', TRUE, 180, 429, 'Too many password reset attempts'),
('payment_limit', 'Payment endpoint rate limit - 10 requests per minute', 10, 1, 'minute', 'user', 'endpoint', TRUE, 170, 429, 'Too many payment requests'),
('upload_limit', 'Upload endpoint rate limit - 20 requests per hour', 20, 1, 'hour', 'user', 'endpoint', TRUE, 160, 429, 'Too many upload requests'),

-- Method-specific rate limits
('post_limit', 'POST method rate limit - 50 requests per hour', 50, 1, 'hour', 'ip', 'method', TRUE, 150, 429, 'Too many POST requests'),
('put_limit', 'PUT method rate limit - 30 requests per hour', 30, 1, 'hour', 'ip', 'method', TRUE, 140, 429, 'Too many PUT requests'),
('delete_limit', 'DELETE method rate limit - 20 requests per hour', 20, 1, 'hour', 'ip', 'method', TRUE, 130, 429, 'Too many DELETE requests'),

-- Tiered rate limits
('admin_limit', 'Admin user rate limit - 10000 requests per hour', 10000, 1, 'hour', 'user', 'global', TRUE, 300, 429, 'Admin rate limit exceeded'),
('premium_limit', 'Premium user rate limit - 2000 requests per hour', 2000, 1, 'hour', 'user', 'global', TRUE, 250, 429, 'Premium user rate limit exceeded'),
('guest_limit', 'Guest user rate limit - 10 requests per hour', 10, 1, 'hour', 'ip', 'global', TRUE, 50, 429, 'Guest rate limit exceeded');

-- Insert default rate limit config
INSERT INTO rate_limit_configs (name, description, is_enabled, default_rule, redis_host, redis_port, redis_db, log_retention_days, stats_retention_days) VALUES
('default_config', 'Default rate limiting configuration', TRUE, 1, 'localhost', 6379, 0, 30, 90);

-- Insert some whitelist entries
INSERT INTO rate_limit_whitelist (type, value, description, is_active) VALUES
('ip', '127.0.0.1', 'Localhost - always allow', TRUE),
('ip', '::1', 'Localhost IPv6 - always allow', TRUE),
('ip', '192.168.0.0/16', 'Private network range', TRUE),
('ip', '10.0.0.0/8', 'Private network range', TRUE),
('ip', '172.16.0.0/12', 'Private network range', TRUE);

-- Insert some blacklist entries (examples)
INSERT INTO rate_limit_blacklist (type, value, reason, is_active, expires_at) VALUES
('ip', '192.168.1.100', 'Suspicious activity detected', TRUE, DATE_ADD(NOW(), INTERVAL 24 HOUR)),
('ip', '10.0.0.50', 'Multiple failed login attempts', TRUE, DATE_ADD(NOW(), INTERVAL 1 HOUR));
