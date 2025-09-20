-- +migrate Up
CREATE TABLE IF NOT EXISTS orders (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_number VARCHAR(50) NOT NULL UNIQUE,
    user_id INT UNSIGNED NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    payment_status VARCHAR(20) DEFAULT 'pending',
    shipping_status VARCHAR(20) DEFAULT 'pending',
    
    -- Customer Information
    customer_name VARCHAR(255) NOT NULL,
    customer_email VARCHAR(255) NOT NULL,
    customer_phone VARCHAR(20) NOT NULL,
    
    -- Address Information
    shipping_address TEXT NOT NULL,
    billing_address TEXT,
    
    -- Pricing Information
    sub_total DECIMAL(10,2) NOT NULL,
    tax_amount DECIMAL(10,2) DEFAULT 0,
    shipping_cost DECIMAL(10,2) DEFAULT 0,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    total_amount DECIMAL(10,2) NOT NULL,
    
    -- Payment Information
    payment_method VARCHAR(20) NOT NULL,
    payment_reference VARCHAR(100),
    paid_at TIMESTAMP NULL,
    
    -- Shipping Information
    shipping_method VARCHAR(100),
    tracking_number VARCHAR(100),
    shipped_at TIMESTAMP NULL,
    delivered_at TIMESTAMP NULL,
    
    -- Additional Information
    notes TEXT,
    admin_notes TEXT,
    tags VARCHAR(500),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    INDEX idx_orders_user_id (user_id),
    INDEX idx_orders_status (status),
    INDEX idx_orders_payment_status (payment_status),
    INDEX idx_orders_shipping_status (shipping_status),
    INDEX idx_orders_order_number (order_number),
    INDEX idx_orders_customer_email (customer_email),
    INDEX idx_orders_created_at (created_at),
    INDEX idx_orders_deleted_at (deleted_at)
);

CREATE TABLE IF NOT EXISTS order_items (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id INT UNSIGNED NOT NULL,
    product_id INT UNSIGNED NOT NULL,
    product_variant_id INT UNSIGNED NULL,
    
    -- Product Information (snapshot at time of order)
    product_name VARCHAR(255) NOT NULL,
    product_sku VARCHAR(100) NOT NULL,
    product_image VARCHAR(500),
    variant_name VARCHAR(255),
    
    -- Pricing Information
    unit_price DECIMAL(10,2) NOT NULL,
    quantity INT NOT NULL,
    total_price DECIMAL(10,2) NOT NULL,
    
    -- Additional Information
    weight DECIMAL(8,2) DEFAULT 0,
    dimensions VARCHAR(100),
    notes TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    FOREIGN KEY (product_variant_id) REFERENCES product_variants(id) ON DELETE SET NULL,
    
    INDEX idx_order_items_order_id (order_id),
    INDEX idx_order_items_product_id (product_id),
    INDEX idx_order_items_product_variant_id (product_variant_id),
    INDEX idx_order_items_deleted_at (deleted_at)
);

CREATE TABLE IF NOT EXISTS carts (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL,
    session_id VARCHAR(100),
    
    -- Cart Information
    items_count INT DEFAULT 0,
    items_quantity INT DEFAULT 0,
    sub_total DECIMAL(10,2) DEFAULT 0,
    tax_amount DECIMAL(10,2) DEFAULT 0,
    shipping_cost DECIMAL(10,2) DEFAULT 0,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    total_amount DECIMAL(10,2) DEFAULT 0,
    
    -- Additional Information
    shipping_address TEXT,
    billing_address TEXT,
    notes TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    INDEX idx_carts_user_id (user_id),
    INDEX idx_carts_session_id (session_id),
    INDEX idx_carts_deleted_at (deleted_at)
);

CREATE TABLE IF NOT EXISTS cart_items (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    cart_id INT UNSIGNED NOT NULL,
    product_id INT UNSIGNED NOT NULL,
    product_variant_id INT UNSIGNED NULL,
    
    quantity INT NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    total_price DECIMAL(10,2) NOT NULL,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (cart_id) REFERENCES carts(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    FOREIGN KEY (product_variant_id) REFERENCES product_variants(id) ON DELETE SET NULL,
    
    INDEX idx_cart_items_cart_id (cart_id),
    INDEX idx_cart_items_product_id (product_id),
    INDEX idx_cart_items_product_variant_id (product_variant_id),
    INDEX idx_cart_items_deleted_at (deleted_at)
);

CREATE TABLE IF NOT EXISTS payments (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id INT UNSIGNED NOT NULL,
    user_id INT UNSIGNED NOT NULL,
    
    -- Payment Information
    payment_method VARCHAR(20) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'VND',
    
    -- Transaction Information
    transaction_id VARCHAR(100) UNIQUE,
    reference_id VARCHAR(100),
    gateway_response TEXT,
    
    -- Additional Information
    description VARCHAR(500),
    notes TEXT,
    
    -- Timestamps
    processed_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    INDEX idx_payments_order_id (order_id),
    INDEX idx_payments_user_id (user_id),
    INDEX idx_payments_status (status),
    INDEX idx_payments_transaction_id (transaction_id),
    INDEX idx_payments_deleted_at (deleted_at)
);

CREATE TABLE IF NOT EXISTS shipping_history (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id INT UNSIGNED NOT NULL,
    
    -- Status Information
    status VARCHAR(20) NOT NULL,
    description VARCHAR(500),
    location VARCHAR(255),
    notes TEXT,
    
    -- Additional Information
    updated_by INT UNSIGNED,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL,
    
    INDEX idx_shipping_history_order_id (order_id),
    INDEX idx_shipping_history_status (status),
    INDEX idx_shipping_history_updated_by (updated_by),
    INDEX idx_shipping_history_created_at (created_at)
);

-- Insert default order statuses and payment methods
-- These are handled by the application, but we can add some constraints
ALTER TABLE orders ADD CONSTRAINT chk_order_status 
CHECK (status IN ('pending', 'confirmed', 'processing', 'shipped', 'delivered', 'cancelled', 'returned', 'refunded'));

ALTER TABLE orders ADD CONSTRAINT chk_payment_status 
CHECK (payment_status IN ('pending', 'paid', 'failed', 'refunded', 'cancelled'));

ALTER TABLE orders ADD CONSTRAINT chk_shipping_status 
CHECK (shipping_status IN ('pending', 'picked_up', 'in_transit', 'delivered', 'failed', 'returned'));

ALTER TABLE orders ADD CONSTRAINT chk_payment_method 
CHECK (payment_method IN ('cash', 'bank', 'card', 'wallet', 'cod'));

ALTER TABLE payments ADD CONSTRAINT chk_payment_method_payment 
CHECK (payment_method IN ('cash', 'bank', 'card', 'wallet', 'cod'));

ALTER TABLE payments ADD CONSTRAINT chk_payment_status_payment 
CHECK (status IN ('pending', 'paid', 'failed', 'refunded', 'cancelled'));

ALTER TABLE shipping_history ADD CONSTRAINT chk_shipping_status_history 
CHECK (status IN ('pending', 'picked_up', 'in_transit', 'delivered', 'failed', 'returned'));

-- +migrate Down
DROP TABLE IF EXISTS shipping_history;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS cart_items;
DROP TABLE IF EXISTS carts;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
