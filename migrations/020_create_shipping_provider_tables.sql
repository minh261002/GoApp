-- Create shipping_providers table
CREATE TABLE shipping_providers (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    code VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    logo VARCHAR(500),
    website VARCHAR(500),
    phone VARCHAR(20),
    email VARCHAR(255),
    config JSON,
    is_active BOOLEAN DEFAULT TRUE,
    is_default BOOLEAN DEFAULT FALSE,
    priority INT DEFAULT 0,
    supports_cod BOOLEAN DEFAULT FALSE,
    supports_tracking BOOLEAN DEFAULT FALSE,
    supports_insurance BOOLEAN DEFAULT FALSE,
    supports_fragile BOOLEAN DEFAULT FALSE,
    min_weight DECIMAL(8,3) DEFAULT 0,
    max_weight DECIMAL(8,3) DEFAULT 0,
    min_value DECIMAL(10,2) DEFAULT 0,
    max_value DECIMAL(10,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_code (code),
    INDEX idx_is_active (is_active),
    INDEX idx_priority (priority),
    INDEX idx_deleted_at (deleted_at)
);

-- Create shipping_rates table
CREATE TABLE shipping_rates (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    provider_id BIGINT UNSIGNED NOT NULL,
    from_province VARCHAR(100) NOT NULL,
    from_district VARCHAR(100) NOT NULL,
    to_province VARCHAR(100) NOT NULL,
    to_district VARCHAR(100) NOT NULL,
    min_weight DECIMAL(8,3) NOT NULL,
    max_weight DECIMAL(8,3) NOT NULL,
    min_value DECIMAL(10,2) NOT NULL,
    max_value DECIMAL(10,2) NOT NULL,
    base_fee DECIMAL(10,2) NOT NULL,
    weight_fee DECIMAL(10,2) DEFAULT 0,
    value_fee DECIMAL(10,2) DEFAULT 0,
    cod_fee DECIMAL(10,2) DEFAULT 0,
    insurance_fee DECIMAL(10,2) DEFAULT 0,
    fragile_fee DECIMAL(10,2) DEFAULT 0,
    min_days INT DEFAULT 1,
    max_days INT DEFAULT 3,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (provider_id) REFERENCES shipping_providers(id) ON DELETE CASCADE,
    INDEX idx_provider_id (provider_id),
    INDEX idx_from_location (from_province, from_district),
    INDEX idx_to_location (to_province, to_district),
    INDEX idx_weight_range (min_weight, max_weight),
    INDEX idx_value_range (min_value, max_value),
    INDEX idx_is_active (is_active),
    INDEX idx_deleted_at (deleted_at)
);

-- Create shipping_orders table
CREATE TABLE shipping_orders (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id BIGINT UNSIGNED NOT NULL,
    provider_id BIGINT UNSIGNED NOT NULL,
    external_id VARCHAR(100),
    label_id VARCHAR(100),
    tracking_code VARCHAR(100),
    from_name VARCHAR(255) NOT NULL,
    from_address TEXT NOT NULL,
    from_phone VARCHAR(20) NOT NULL,
    from_email VARCHAR(255),
    to_name VARCHAR(255) NOT NULL,
    to_address TEXT NOT NULL,
    to_phone VARCHAR(20) NOT NULL,
    to_email VARCHAR(255),
    weight DECIMAL(8,3) NOT NULL,
    value DECIMAL(10,2) NOT NULL,
    cod DECIMAL(10,2) DEFAULT 0,
    insurance DECIMAL(10,2) DEFAULT 0,
    shipping_fee DECIMAL(10,2) NOT NULL,
    cod_fee DECIMAL(10,2) DEFAULT 0,
    insurance_fee DECIMAL(10,2) DEFAULT 0,
    total_fee DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) NOT NULL,
    status_text VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    shipped_at TIMESTAMP NULL,
    delivered_at TIMESTAMP NULL,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    FOREIGN KEY (provider_id) REFERENCES shipping_providers(id) ON DELETE CASCADE,
    INDEX idx_order_id (order_id),
    INDEX idx_provider_id (provider_id),
    INDEX idx_external_id (external_id),
    INDEX idx_label_id (label_id),
    INDEX idx_tracking_code (tracking_code),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);

-- Create shipping_tracking table
CREATE TABLE shipping_tracking (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    shipping_order_id BIGINT UNSIGNED NOT NULL,
    status VARCHAR(50) NOT NULL,
    status_text VARCHAR(255) NOT NULL,
    location VARCHAR(255),
    note TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (shipping_order_id) REFERENCES shipping_orders(id) ON DELETE CASCADE,
    INDEX idx_shipping_order_id (shipping_order_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);

-- Insert default shipping providers
INSERT INTO shipping_providers (name, code, display_name, description, website, phone, email, config, is_active, is_default, priority, supports_cod, supports_tracking, supports_insurance, supports_fragile, min_weight, max_weight, min_value, max_value) VALUES
('Giao Hàng Tiết Kiệm', 'ghtk', 'Giao Hàng Tiết Kiệm', 'Dịch vụ giao hàng nhanh và tiết kiệm', 'https://ghtk.vn', '1900 1234', 'support@ghtk.vn', '{"base_url": "https://services.ghtk.vn", "test_url": "https://dev.ghtk.vn", "timeout": 30}', TRUE, TRUE, 100, TRUE, TRUE, TRUE, TRUE, 0.1, 30.0, 0, 50000000),
('Giao Hàng Nhanh', 'ghn', 'Giao Hàng Nhanh', 'Dịch vụ giao hàng nhanh 24/7', 'https://ghn.vn', '1900 1235', 'support@ghn.vn', '{"base_url": "https://api.ghn.vn", "timeout": 30}', TRUE, FALSE, 90, TRUE, TRUE, TRUE, FALSE, 0.1, 25.0, 0, 30000000),
('Viettel Post', 'viettel_post', 'Viettel Post', 'Dịch vụ bưu chính Viettel', 'https://viettelpost.vn', '1900 1236', 'support@viettelpost.vn', '{"base_url": "https://api.viettelpost.vn", "timeout": 30}', TRUE, FALSE, 80, TRUE, TRUE, TRUE, FALSE, 0.1, 20.0, 0, 20000000),
('J&T Express', 'jt', 'J&T Express', 'Dịch vụ giao hàng quốc tế', 'https://jtexpress.vn', '1900 1237', 'support@jtexpress.vn', '{"base_url": "https://api.jtexpress.vn", "timeout": 30}', TRUE, FALSE, 70, TRUE, TRUE, FALSE, FALSE, 0.1, 15.0, 0, 10000000),
('Best Express', 'best', 'Best Express', 'Dịch vụ giao hàng tốt nhất', 'https://bestexpress.vn', '1900 1238', 'support@bestexpress.vn', '{"base_url": "https://api.bestexpress.vn", "timeout": 30}', TRUE, FALSE, 60, TRUE, TRUE, TRUE, TRUE, 0.1, 10.0, 0, 5000000);

-- Insert default shipping rates for GHTK
INSERT INTO shipping_rates (provider_id, from_province, from_district, to_province, to_district, min_weight, max_weight, min_value, max_value, base_fee, weight_fee, value_fee, cod_fee, insurance_fee, fragile_fee, min_days, max_days, is_active) VALUES
-- Hà Nội to TP.HCM
(1, 'Hà Nội', 'Quận Ba Đình', 'TP. Hồ Chí Minh', 'Quận 1', 0.1, 0.5, 0, 1000000, 22000, 0, 0, 0, 0, 0, 1, 2, TRUE),
(1, 'Hà Nội', 'Quận Ba Đình', 'TP. Hồ Chí Minh', 'Quận 1', 0.5, 1.0, 0, 1000000, 25000, 0, 0, 0, 0, 0, 1, 2, TRUE),
(1, 'Hà Nội', 'Quận Ba Đình', 'TP. Hồ Chí Minh', 'Quận 1', 1.0, 2.0, 0, 1000000, 30000, 0, 0, 0, 0, 0, 1, 2, TRUE),
(1, 'Hà Nội', 'Quận Ba Đình', 'TP. Hồ Chí Minh', 'Quận 1', 2.0, 5.0, 0, 1000000, 35000, 0, 0, 0, 0, 0, 1, 2, TRUE),
(1, 'Hà Nội', 'Quận Ba Đình', 'TP. Hồ Chí Minh', 'Quận 1', 5.0, 10.0, 0, 1000000, 40000, 0, 0, 0, 0, 0, 1, 2, TRUE),

-- TP.HCM to Hà Nội
(1, 'TP. Hồ Chí Minh', 'Quận 1', 'Hà Nội', 'Quận Ba Đình', 0.1, 0.5, 0, 1000000, 22000, 0, 0, 0, 0, 0, 1, 2, TRUE),
(1, 'TP. Hồ Chí Minh', 'Quận 1', 'Hà Nội', 'Quận Ba Đình', 0.5, 1.0, 0, 1000000, 25000, 0, 0, 0, 0, 0, 1, 2, TRUE),
(1, 'TP. Hồ Chí Minh', 'Quận 1', 'Hà Nội', 'Quận Ba Đình', 1.0, 2.0, 0, 1000000, 30000, 0, 0, 0, 0, 0, 1, 2, TRUE),
(1, 'TP. Hồ Chí Minh', 'Quận 1', 'Hà Nội', 'Quận Ba Đình', 2.0, 5.0, 0, 1000000, 35000, 0, 0, 0, 0, 0, 1, 2, TRUE),
(1, 'TP. Hồ Chí Minh', 'Quận 1', 'Hà Nội', 'Quận Ba Đình', 5.0, 10.0, 0, 1000000, 40000, 0, 0, 0, 0, 0, 1, 2, TRUE),

-- Nội thành Hà Nội
(1, 'Hà Nội', 'Quận Ba Đình', 'Hà Nội', 'Quận Cầu Giấy', 0.1, 0.5, 0, 1000000, 15000, 0, 0, 0, 0, 0, 1, 1, TRUE),
(1, 'Hà Nội', 'Quận Ba Đình', 'Hà Nội', 'Quận Cầu Giấy', 0.5, 1.0, 0, 1000000, 18000, 0, 0, 0, 0, 0, 1, 1, TRUE),
(1, 'Hà Nội', 'Quận Ba Đình', 'Hà Nội', 'Quận Cầu Giấy', 1.0, 2.0, 0, 1000000, 20000, 0, 0, 0, 0, 0, 1, 1, TRUE),

-- Nội thành TP.HCM
(1, 'TP. Hồ Chí Minh', 'Quận 1', 'TP. Hồ Chí Minh', 'Quận 3', 0.1, 0.5, 0, 1000000, 15000, 0, 0, 0, 0, 0, 1, 1, TRUE),
(1, 'TP. Hồ Chí Minh', 'Quận 1', 'TP. Hồ Chí Minh', 'Quận 3', 0.5, 1.0, 0, 1000000, 18000, 0, 0, 0, 0, 0, 1, 1, TRUE),
(1, 'TP. Hồ Chí Minh', 'Quận 1', 'TP. Hồ Chí Minh', 'Quận 3', 1.0, 2.0, 0, 1000000, 20000, 0, 0, 0, 0, 0, 1, 1, TRUE);
