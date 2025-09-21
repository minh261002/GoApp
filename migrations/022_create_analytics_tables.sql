-- Create analytics_reports table
CREATE TABLE analytics_reports (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type ENUM('sales', 'traffic', 'inventory', 'user', 'product', 'order', 'revenue', 'marketing') NOT NULL,
    period ENUM('daily', 'weekly', 'monthly', 'yearly', 'custom') NOT NULL,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    filters JSON,
    data JSON,
    summary TEXT,
    insights TEXT,
    status ENUM('pending', 'processing', 'completed', 'failed') DEFAULT 'pending',
    is_scheduled BOOLEAN DEFAULT FALSE,
    is_public BOOLEAN DEFAULT FALSE,
    created_by BIGINT UNSIGNED NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_analytics_reports_type (type),
    INDEX idx_analytics_reports_period (period),
    INDEX idx_analytics_reports_status (status),
    INDEX idx_analytics_reports_created_by (created_by),
    INDEX idx_analytics_reports_deleted_at (deleted_at),
    INDEX idx_analytics_reports_dates (start_date, end_date),
    
    -- Foreign keys
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create analytics_metrics table
CREATE TABLE analytics_metrics (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    report_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(255) NOT NULL,
    value DECIMAL(15,2) NOT NULL,
    unit VARCHAR(50),
    category VARCHAR(100),
    sub_category VARCHAR(100),
    previous_value DECIMAL(15,2) DEFAULT 0.00,
    change_percent DECIMAL(5,2) DEFAULT 0.00,
    trend ENUM('up', 'down', 'stable') DEFAULT 'stable',
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_analytics_metrics_report_id (report_id),
    INDEX idx_analytics_metrics_category (category),
    INDEX idx_analytics_metrics_deleted_at (deleted_at),
    
    -- Foreign keys
    FOREIGN KEY (report_id) REFERENCES analytics_reports(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create analytics_dashboards table
CREATE TABLE analytics_dashboards (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    layout JSON,
    is_public BOOLEAN DEFAULT FALSE,
    user_id BIGINT UNSIGNED NULL,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_analytics_dashboards_user_id (user_id),
    INDEX idx_analytics_dashboards_is_public (is_public),
    INDEX idx_analytics_dashboards_deleted_at (deleted_at),
    
    -- Foreign keys
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create analytics_widgets table
CREATE TABLE analytics_widgets (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    dashboard_id BIGINT UNSIGNED NOT NULL,
    type ENUM('chart', 'table', 'metric', 'kpi', 'gauge', 'progress', 'list') NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    config JSON,
    data JSON,
    position JSON,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_analytics_widgets_dashboard_id (dashboard_id),
    INDEX idx_analytics_widgets_type (type),
    INDEX idx_analytics_widgets_is_active (is_active),
    INDEX idx_analytics_widgets_deleted_at (deleted_at),
    
    -- Foreign keys
    FOREIGN KEY (dashboard_id) REFERENCES analytics_dashboards(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create analytics_events table
CREATE TABLE analytics_events (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    event_type ENUM('page_view', 'click', 'purchase', 'add_to_cart', 'remove_from_cart', 'search', 'sign_up', 'login', 'logout', 'email_open', 'email_click') NOT NULL,
    event_name VARCHAR(255) NOT NULL,
    entity_type ENUM('order', 'product', 'user', 'category', 'brand', 'page', 'email', 'banner', 'slider') NOT NULL,
    entity_id BIGINT UNSIGNED NOT NULL,
    properties JSON,
    value DECIMAL(15,2) DEFAULT 0.00,
    user_id BIGINT UNSIGNED NULL,
    session_id VARCHAR(255),
    ip_address VARCHAR(45),
    user_agent TEXT,
    referer VARCHAR(500),
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_analytics_events_event_type (event_type),
    INDEX idx_analytics_events_entity_type (entity_type),
    INDEX idx_analytics_events_entity_id (entity_id),
    INDEX idx_analytics_events_user_id (user_id),
    INDEX idx_analytics_events_session_id (session_id),
    INDEX idx_analytics_events_created_at (created_at),
    INDEX idx_analytics_events_deleted_at (deleted_at),
    
    -- Foreign keys
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default analytics dashboard
INSERT INTO analytics_dashboards (name, description, layout, is_public, created_at, updated_at) VALUES
('Default Dashboard', 'Default analytics dashboard with key metrics', '{"columns": 3, "rows": 4}', TRUE, NOW(), NOW());

-- Insert default widgets for the dashboard
INSERT INTO analytics_widgets (dashboard_id, type, title, description, config, data, position, is_active, created_at, updated_at) VALUES
(1, 'kpi', 'Total Revenue', 'Total revenue for the selected period', '{"format": "currency", "color": "green"}', '{"value": 0}', '{"x": 0, "y": 0, "w": 1, "h": 1}', TRUE, NOW(), NOW()),
(1, 'kpi', 'Total Orders', 'Total number of orders', '{"format": "number", "color": "blue"}', '{"value": 0}', '{"x": 1, "y": 0, "w": 1, "h": 1}', TRUE, NOW(), NOW()),
(1, 'kpi', 'Average Order Value', 'Average value per order', '{"format": "currency", "color": "purple"}', '{"value": 0}', '{"x": 2, "y": 0, "w": 1, "h": 1}', TRUE, NOW(), NOW()),
(1, 'chart', 'Revenue Trend', 'Revenue trend over time', '{"type": "line", "xAxis": "date", "yAxis": "revenue"}', '{"data": []}', '{"x": 0, "y": 1, "w": 2, "h": 2}', TRUE, NOW(), NOW()),
(1, 'chart', 'Top Products', 'Best selling products', '{"type": "bar", "xAxis": "product", "yAxis": "sales"}', '{"data": []}', '{"x": 2, "y": 1, "w": 1, "h": 2}', TRUE, NOW(), NOW()),
(1, 'table', 'Recent Orders', 'Most recent orders', '{"columns": ["order_number", "customer", "amount", "status"]}', '{"data": []}', '{"x": 0, "y": 3, "w": 3, "h": 1}', TRUE, NOW(), NOW());
