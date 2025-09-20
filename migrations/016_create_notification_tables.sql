-- +migrate Up

-- Create notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NULL,
    type VARCHAR(50) NOT NULL,
    priority VARCHAR(20) NOT NULL DEFAULT 'normal',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    channel VARCHAR(20) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    data JSON,
    action_url VARCHAR(500),
    image_url VARCHAR(500),
    is_read BOOLEAN DEFAULT FALSE,
    is_archived BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP NULL,
    sent_at TIMESTAMP NULL,
    delivered_at TIMESTAMP NULL,
    failed_at TIMESTAMP NULL,
    retry_count INT DEFAULT 0,
    error_msg TEXT,
    expires_at TIMESTAMP NULL,
    scheduled_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_notifications_user_id (user_id),
    INDEX idx_notifications_type (type),
    INDEX idx_notifications_status (status),
    INDEX idx_notifications_channel (channel),
    INDEX idx_notifications_priority (priority),
    INDEX idx_notifications_is_read (is_read),
    INDEX idx_notifications_is_archived (is_archived),
    INDEX idx_notifications_scheduled_at (scheduled_at),
    INDEX idx_notifications_created_at (created_at),
    INDEX idx_notifications_deleted_at (deleted_at),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create notification_templates table
CREATE TABLE IF NOT EXISTS notification_templates (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    type VARCHAR(50) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    variables JSON,
    is_active BOOLEAN DEFAULT TRUE,
    is_system BOOLEAN DEFAULT FALSE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_notification_templates_type (type),
    INDEX idx_notification_templates_channel (channel),
    INDEX idx_notification_templates_is_active (is_active),
    INDEX idx_notification_templates_is_system (is_system),
    INDEX idx_notification_templates_deleted_at (deleted_at)
);

-- Create notification_preferences table
CREATE TABLE IF NOT EXISTS notification_preferences (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL,
    type VARCHAR(50) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    is_enabled BOOLEAN DEFAULT TRUE,
    frequency VARCHAR(20) DEFAULT 'immediate',
    quiet_hours VARCHAR(20),
    timezone VARCHAR(50) DEFAULT 'UTC',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    UNIQUE KEY unique_user_type_channel (user_id, type, channel),
    INDEX idx_notification_preferences_user_id (user_id),
    INDEX idx_notification_preferences_type (type),
    INDEX idx_notification_preferences_channel (channel),
    INDEX idx_notification_preferences_is_enabled (is_enabled),
    INDEX idx_notification_preferences_deleted_at (deleted_at),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create notification_logs table
CREATE TABLE IF NOT EXISTS notification_logs (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    notification_id INT UNSIGNED NOT NULL,
    channel VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    provider VARCHAR(100),
    provider_id VARCHAR(255),
    response TEXT,
    error_msg TEXT,
    attempted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    delivered_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_notification_logs_notification_id (notification_id),
    INDEX idx_notification_logs_channel (channel),
    INDEX idx_notification_logs_status (status),
    INDEX idx_notification_logs_provider (provider),
    INDEX idx_notification_logs_attempted_at (attempted_at),
    INDEX idx_notification_logs_deleted_at (deleted_at),
    
    FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE
);

-- Create notification_queue table
CREATE TABLE IF NOT EXISTS notification_queue (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    notification_id INT UNSIGNED NOT NULL,
    priority VARCHAR(20) NOT NULL DEFAULT 'normal',
    channel VARCHAR(20) NOT NULL,
    scheduled_at TIMESTAMP NOT NULL,
    processed_at TIMESTAMP NULL,
    retry_count INT DEFAULT 0,
    max_retries INT DEFAULT 3,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_notification_queue_notification_id (notification_id),
    INDEX idx_notification_queue_priority (priority),
    INDEX idx_notification_queue_channel (channel),
    INDEX idx_notification_queue_scheduled_at (scheduled_at),
    INDEX idx_notification_queue_status (status),
    INDEX idx_notification_queue_deleted_at (deleted_at),
    
    FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE
);

-- Create notification_stats table
CREATE TABLE IF NOT EXISTS notification_stats (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    date DATE NOT NULL,
    type VARCHAR(50) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    total_sent BIGINT DEFAULT 0,
    total_delivered BIGINT DEFAULT 0,
    total_read BIGINT DEFAULT 0,
    total_failed BIGINT DEFAULT 0,
    delivery_rate DECIMAL(5,2) DEFAULT 0.00,
    read_rate DECIMAL(5,2) DEFAULT 0.00,
    average_delivery_time BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    UNIQUE KEY unique_date_type_channel (date, type, channel),
    INDEX idx_notification_stats_date (date),
    INDEX idx_notification_stats_type (type),
    INDEX idx_notification_stats_channel (channel),
    INDEX idx_notification_stats_deleted_at (deleted_at)
);

-- Insert default notification templates
INSERT INTO notification_templates (name, type, channel, subject, body, variables, is_active, is_system, description) VALUES
-- Order notifications
('order_placed', 'order', 'email', 'Order Confirmation - {{order_number}}', 'Thank you for your order! Your order #{{order_number}} has been placed successfully.', '["order_number", "customer_name", "total_amount", "order_date"]', TRUE, TRUE, 'Email template for order confirmation'),
('order_placed', 'order', 'in_app', 'Order Placed', 'Your order #{{order_number}} has been placed successfully.', '["order_number", "total_amount"]', TRUE, TRUE, 'In-app notification for order confirmation'),
('order_shipped', 'shipping', 'email', 'Order Shipped - {{order_number}}', 'Great news! Your order #{{order_number}} has been shipped and is on its way to you.', '["order_number", "tracking_number", "estimated_delivery"]', TRUE, TRUE, 'Email template for order shipped'),
('order_delivered', 'shipping', 'email', 'Order Delivered - {{order_number}}', 'Your order #{{order_number}} has been delivered successfully.', '["order_number", "delivery_date"]', TRUE, TRUE, 'Email template for order delivered'),
('order_cancelled', 'order', 'email', 'Order Cancelled - {{order_number}}', 'Your order #{{order_number}} has been cancelled.', '["order_number", "cancellation_reason"]', TRUE, TRUE, 'Email template for order cancellation'),

-- Payment notifications
('payment_success', 'payment', 'email', 'Payment Successful - {{order_number}}', 'Your payment for order #{{order_number}} has been processed successfully.', '["order_number", "amount", "payment_method"]', TRUE, TRUE, 'Email template for successful payment'),
('payment_failed', 'payment', 'email', 'Payment Failed - {{order_number}}', 'Your payment for order #{{order_number}} has failed. Please try again.', '["order_number", "amount", "payment_method", "error_message"]', TRUE, TRUE, 'Email template for failed payment'),

-- Product notifications
('product_back_in_stock', 'product', 'email', 'Product Back in Stock - {{product_name}}', 'Good news! {{product_name}} is back in stock and available for purchase.', '["product_name", "product_url", "current_price"]', TRUE, TRUE, 'Email template for product back in stock'),
('price_drop', 'product', 'email', 'Price Drop Alert - {{product_name}}', 'The price of {{product_name}} has dropped! Check it out now.', '["product_name", "old_price", "new_price", "product_url"]', TRUE, TRUE, 'Email template for price drop notification'),

-- Promotion notifications
('promotion_available', 'promotion', 'email', 'Special Offer - {{promotion_title}}', 'Don\'t miss out on this special offer: {{promotion_title}}', '["promotion_title", "discount_percentage", "valid_until", "promotion_url"]', TRUE, TRUE, 'Email template for promotion notification'),

-- System notifications
('system_maintenance', 'system', 'email', 'Scheduled Maintenance - {{maintenance_date}}', 'We will be performing scheduled maintenance on {{maintenance_date}}. The system may be temporarily unavailable.', '["maintenance_date", "maintenance_duration", "affected_services"]', TRUE, TRUE, 'Email template for system maintenance notification'),
('security_alert', 'security', 'email', 'Security Alert - {{alert_type}}', 'We detected unusual activity on your account. Please review and take necessary action.', '["alert_type", "detected_at", "recommended_action"]', TRUE, TRUE, 'Email template for security alert'),

-- Review notifications
('review_request', 'review', 'email', 'How was your experience? - {{product_name}}', 'We\'d love to hear about your experience with {{product_name}}. Please leave a review.', '["product_name", "product_url", "review_url"]', TRUE, TRUE, 'Email template for review request'),

-- Wishlist notifications
('wishlist_item_on_sale', 'wishlist', 'email', 'Wishlist Item on Sale - {{product_name}}', 'An item from your wishlist is now on sale! {{product_name}}', '["product_name", "old_price", "new_price", "product_url"]', TRUE, TRUE, 'Email template for wishlist item on sale'),

-- Inventory notifications
('low_stock_alert', 'inventory', 'email', 'Low Stock Alert - {{product_name}}', 'Product {{product_name}} is running low on stock. Current quantity: {{current_quantity}}', '["product_name", "current_quantity", "minimum_quantity", "product_url"]', TRUE, TRUE, 'Email template for low stock alert'),

-- Coupon notifications
('coupon_expiring', 'coupon', 'email', 'Coupon Expiring Soon - {{coupon_code}}', 'Your coupon {{coupon_code}} expires on {{expiry_date}}. Use it before it\'s too late!', '["coupon_code", "discount_amount", "expiry_date", "coupon_url"]', TRUE, TRUE, 'Email template for expiring coupon'),

-- Point notifications
('points_earned', 'point', 'email', 'Points Earned - {{points_amount}}', 'You\'ve earned {{points_amount}} points! Your current balance is {{total_points}}.', '["points_amount", "total_points", "earned_from", "points_url"]', TRUE, TRUE, 'Email template for points earned'),
('points_expiring', 'point', 'email', 'Points Expiring Soon - {{points_amount}}', 'You have {{points_amount}} points expiring on {{expiry_date}}. Use them before they expire!', '["points_amount", "expiry_date", "points_url"]', TRUE, TRUE, 'Email template for expiring points');

-- Insert default notification preferences for existing users
INSERT INTO notification_preferences (user_id, type, channel, is_enabled, frequency, timezone)
SELECT 
    u.id,
    'order',
    'email',
    TRUE,
    'immediate',
    'UTC'
FROM users u
WHERE u.role = 'user';

INSERT INTO notification_preferences (user_id, type, channel, is_enabled, frequency, timezone)
SELECT 
    u.id,
    'order',
    'in_app',
    TRUE,
    'immediate',
    'UTC'
FROM users u
WHERE u.role = 'user';

INSERT INTO notification_preferences (user_id, type, channel, is_enabled, frequency, timezone)
SELECT 
    u.id,
    'payment',
    'email',
    TRUE,
    'immediate',
    'UTC'
FROM users u
WHERE u.role = 'user';

INSERT INTO notification_preferences (user_id, type, channel, is_enabled, frequency, timezone)
SELECT 
    u.id,
    'shipping',
    'email',
    TRUE,
    'immediate',
    'UTC'
FROM users u
WHERE u.role = 'user';

INSERT INTO notification_preferences (user_id, type, channel, is_enabled, frequency, timezone)
SELECT 
    u.id,
    'product',
    'email',
    TRUE,
    'daily',
    'UTC'
FROM users u
WHERE u.role = 'user';

INSERT INTO notification_preferences (user_id, type, channel, is_enabled, frequency, timezone)
SELECT 
    u.id,
    'promotion',
    'email',
    TRUE,
    'immediate',
    'UTC'
FROM users u
WHERE u.role = 'user';

INSERT INTO notification_preferences (user_id, type, channel, is_enabled, frequency, timezone)
SELECT 
    u.id,
    'system',
    'email',
    TRUE,
    'immediate',
    'UTC'
FROM users u
WHERE u.role = 'user';

INSERT INTO notification_preferences (user_id, type, channel, is_enabled, frequency, timezone)
SELECT 
    u.id,
    'security',
    'email',
    TRUE,
    'immediate',
    'UTC'
FROM users u
WHERE u.role = 'user';

-- +migrate Down
DROP TABLE IF EXISTS notification_stats;
DROP TABLE IF EXISTS notification_queue;
DROP TABLE IF EXISTS notification_logs;
DROP TABLE IF EXISTS notification_preferences;
DROP TABLE IF EXISTS notification_templates;
DROP TABLE IF EXISTS notifications;
