-- Create order tracking tables

-- Create order_trackings table
CREATE TABLE IF NOT EXISTS order_trackings (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id INT UNSIGNED NOT NULL,
    
    -- Tracking Information
    tracking_number VARCHAR(100) NOT NULL,
    carrier VARCHAR(50) NOT NULL,
    carrier_code VARCHAR(20) NOT NULL,
    
    -- Current Status
    status VARCHAR(50) NOT NULL,
    status_text VARCHAR(255) NOT NULL,
    location VARCHAR(255),
    description TEXT,
    
    -- Estimated Delivery
    estimated_delivery TIMESTAMP NULL,
    actual_delivery TIMESTAMP NULL,
    
    -- Tracking URL
    tracking_url VARCHAR(500),
    
    -- Last Update
    last_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_sync_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Settings
    auto_sync BOOLEAN DEFAULT TRUE,
    notify_user BOOLEAN DEFAULT TRUE,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_order_trackings_order_id (order_id),
    INDEX idx_order_trackings_tracking_number (tracking_number),
    INDEX idx_order_trackings_carrier (carrier),
    INDEX idx_order_trackings_status (status),
    INDEX idx_order_trackings_is_active (is_active),
    INDEX idx_order_trackings_deleted_at (deleted_at),
    
    -- Foreign key
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create order_tracking_events table
CREATE TABLE IF NOT EXISTS order_tracking_events (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_tracking_id INT UNSIGNED NOT NULL,
    
    -- Event Information
    status VARCHAR(50) NOT NULL,
    status_text VARCHAR(255) NOT NULL,
    location VARCHAR(255),
    description TEXT,
    
    -- Event Details
    event_type VARCHAR(50) NOT NULL,
    event_code VARCHAR(20),
    is_important BOOLEAN DEFAULT FALSE,
    
    -- Source
    source VARCHAR(50) NOT NULL,
    source_data TEXT,
    
    -- Timestamps
    event_time TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes
    INDEX idx_order_tracking_events_order_tracking_id (order_tracking_id),
    INDEX idx_order_tracking_events_status (status),
    INDEX idx_order_tracking_events_event_type (event_type),
    INDEX idx_order_tracking_events_event_time (event_time),
    INDEX idx_order_tracking_events_is_important (is_important),
    
    -- Foreign key
    FOREIGN KEY (order_tracking_id) REFERENCES order_trackings(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create order_tracking_webhooks table
CREATE TABLE IF NOT EXISTS order_tracking_webhooks (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    carrier VARCHAR(50) NOT NULL,
    carrier_code VARCHAR(20) NOT NULL,
    
    -- Webhook Configuration
    url VARCHAR(500) NOT NULL,
    secret VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Events to Track
    events TEXT, -- JSON array of events
    retry_count INT DEFAULT 3,
    timeout INT DEFAULT 30, -- seconds
    
    -- Statistics
    success_count INT DEFAULT 0,
    failure_count INT DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_order_tracking_webhooks_carrier (carrier),
    INDEX idx_order_tracking_webhooks_carrier_code (carrier_code),
    INDEX idx_order_tracking_webhooks_is_active (is_active),
    INDEX idx_order_tracking_webhooks_deleted_at (deleted_at),
    
    -- Unique constraint
    UNIQUE KEY unique_carrier_code (carrier, carrier_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create order_tracking_notifications table
CREATE TABLE IF NOT EXISTS order_tracking_notifications (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_tracking_id INT UNSIGNED NOT NULL,
    user_id INT UNSIGNED NOT NULL,
    event_id INT UNSIGNED NOT NULL,
    
    -- Notification Details
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    is_sent BOOLEAN DEFAULT FALSE,
    sent_at TIMESTAMP NULL,
    
    -- Retry Information
    retry_count INT DEFAULT 0,
    max_retries INT DEFAULT 3,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_order_tracking_notifications_order_tracking_id (order_tracking_id),
    INDEX idx_order_tracking_notifications_user_id (user_id),
    INDEX idx_order_tracking_notifications_event_id (event_id),
    INDEX idx_order_tracking_notifications_type (type),
    INDEX idx_order_tracking_notifications_is_sent (is_sent),
    INDEX idx_order_tracking_notifications_deleted_at (deleted_at),
    
    -- Foreign keys
    FOREIGN KEY (order_tracking_id) REFERENCES order_trackings(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (event_id) REFERENCES order_tracking_events(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
