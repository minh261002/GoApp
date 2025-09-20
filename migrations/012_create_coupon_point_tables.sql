-- +migrate Up
CREATE TABLE IF NOT EXISTS coupons (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(20) NOT NULL, -- percentage, fixed, free_shipping, buy_x_get_y
    status VARCHAR(20) DEFAULT 'active', -- active, inactive, expired, used
    
    -- Discount Configuration
    discount_value DECIMAL(10,2) NOT NULL,
    min_order_amount DECIMAL(10,2) DEFAULT 0,
    max_discount_amount DECIMAL(10,2) DEFAULT 0,
    
    -- Usage Configuration
    usage_limit INT DEFAULT 0, -- 0 = unlimited
    usage_count INT DEFAULT 0,
    usage_per_user INT DEFAULT 1,
    
    -- Validity Period
    valid_from TIMESTAMP NOT NULL,
    valid_to TIMESTAMP NOT NULL,
    
    -- Target Configuration
    target_type VARCHAR(20) DEFAULT 'all', -- all, product, category, brand, user
    target_ids TEXT, -- JSON array of target IDs
    
    -- Additional Configuration
    is_stackable BOOLEAN DEFAULT FALSE,
    is_first_time_only BOOLEAN DEFAULT FALSE,
    is_new_user_only BOOLEAN DEFAULT FALSE,
    
    -- Metadata
    created_by INT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE,
    
    INDEX idx_coupons_code (code),
    INDEX idx_coupons_status (status),
    INDEX idx_coupons_type (type),
    INDEX idx_coupons_valid_from (valid_from),
    INDEX idx_coupons_valid_to (valid_to),
    INDEX idx_coupons_created_by (created_by),
    INDEX idx_coupons_deleted_at (deleted_at)
);

-- Create Coupon Usage Table
CREATE TABLE IF NOT EXISTS coupon_usages (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    coupon_id INT UNSIGNED NOT NULL,
    user_id INT UNSIGNED NOT NULL,
    order_id INT UNSIGNED NOT NULL,
    
    -- Usage Details
    discount_amount DECIMAL(10,2) NOT NULL,
    order_amount DECIMAL(10,2) NOT NULL,
    used_at TIMESTAMP NOT NULL,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (coupon_id) REFERENCES coupons(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    
    INDEX idx_coupon_usages_coupon_id (coupon_id),
    INDEX idx_coupon_usages_user_id (user_id),
    INDEX idx_coupon_usages_order_id (order_id),
    INDEX idx_coupon_usages_used_at (used_at),
    INDEX idx_coupon_usages_deleted_at (deleted_at)
);

-- Create Points Table
CREATE TABLE IF NOT EXISTS points (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL UNIQUE,
    
    -- Point Balances
    balance INT DEFAULT 0,
    total_earned INT DEFAULT 0,
    total_redeemed INT DEFAULT 0,
    total_expired INT DEFAULT 0,
    
    -- Point Configuration
    expiry_days INT DEFAULT 365,
    is_active BOOLEAN DEFAULT TRUE,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    INDEX idx_points_user_id (user_id),
    INDEX idx_points_balance (balance),
    INDEX idx_points_is_active (is_active),
    INDEX idx_points_deleted_at (deleted_at)
);

-- Create Point Transactions Table
CREATE TABLE IF NOT EXISTS point_transactions (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    point_id INT UNSIGNED NOT NULL,
    user_id INT UNSIGNED NOT NULL,
    
    -- Transaction Details
    type VARCHAR(20) NOT NULL, -- earn, redeem, expire, refund, adjust
    status VARCHAR(20) DEFAULT 'pending', -- pending, completed, cancelled, expired
    amount INT NOT NULL, -- positive = earn, negative = redeem
    balance INT NOT NULL, -- balance after transaction
    
    -- Reference Information
    reference_type VARCHAR(50), -- order, coupon, manual, etc.
    reference_id INT UNSIGNED,
    order_id INT UNSIGNED NULL,
    
    -- Description
    description TEXT,
    notes TEXT,
    
    -- Expiry
    expires_at TIMESTAMP NULL,
    
    -- Metadata
    created_by INT UNSIGNED NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (point_id) REFERENCES points(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE SET NULL,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
    
    INDEX idx_point_transactions_point_id (point_id),
    INDEX idx_point_transactions_user_id (user_id),
    INDEX idx_point_transactions_type (type),
    INDEX idx_point_transactions_status (status),
    INDEX idx_point_transactions_reference_type (reference_type),
    INDEX idx_point_transactions_reference_id (reference_id),
    INDEX idx_point_transactions_order_id (order_id),
    INDEX idx_point_transactions_expires_at (expires_at),
    INDEX idx_point_transactions_created_by (created_by),
    INDEX idx_point_transactions_deleted_at (deleted_at)
);

-- Add constraints
ALTER TABLE coupons ADD CONSTRAINT chk_coupon_type 
CHECK (type IN ('percentage', 'fixed', 'free_shipping', 'buy_x_get_y'));

ALTER TABLE coupons ADD CONSTRAINT chk_coupon_status 
CHECK (status IN ('active', 'inactive', 'expired', 'used'));

ALTER TABLE coupons ADD CONSTRAINT chk_coupon_discount_value 
CHECK (discount_value > 0);

ALTER TABLE coupons ADD CONSTRAINT chk_coupon_percentage 
CHECK (type != 'percentage' OR discount_value <= 100);

ALTER TABLE coupons ADD CONSTRAINT chk_coupon_usage_limit 
CHECK (usage_limit >= 0);

ALTER TABLE coupons ADD CONSTRAINT chk_coupon_usage_per_user 
CHECK (usage_per_user >= 1);

ALTER TABLE coupons ADD CONSTRAINT chk_coupon_valid_dates 
CHECK (valid_to > valid_from);

ALTER TABLE point_transactions ADD CONSTRAINT chk_point_transaction_type 
CHECK (type IN ('earn', 'redeem', 'expire', 'refund', 'adjust'));

ALTER TABLE point_transactions ADD CONSTRAINT chk_point_transaction_status 
CHECK (status IN ('pending', 'completed', 'cancelled', 'expired'));

ALTER TABLE point_transactions ADD CONSTRAINT chk_point_transaction_amount 
CHECK (amount != 0);

-- Create indexes for better performance
CREATE INDEX idx_coupons_status_valid ON coupons(status, valid_from, valid_to);
CREATE INDEX idx_coupons_type_status ON coupons(type, status);
CREATE INDEX idx_coupons_usage_count ON coupons(usage_count, usage_limit);

CREATE INDEX idx_coupon_usages_coupon_user ON coupon_usages(coupon_id, user_id);
CREATE INDEX idx_coupon_usages_user_used ON coupon_usages(user_id, used_at);

CREATE INDEX idx_points_balance_active ON points(balance, is_active);
CREATE INDEX idx_points_user_active ON points(user_id, is_active);

CREATE INDEX idx_point_transactions_user_type ON point_transactions(user_id, type);
CREATE INDEX idx_point_transactions_user_status ON point_transactions(user_id, status);
CREATE INDEX idx_point_transactions_type_status ON point_transactions(type, status);
CREATE INDEX idx_point_transactions_expires_status ON point_transactions(expires_at, status);

-- Create composite indexes for common queries
CREATE INDEX idx_coupons_active_valid ON coupons(status, valid_from, valid_to) WHERE status = 'active';
CREATE INDEX idx_point_transactions_user_completed ON point_transactions(user_id, status) WHERE status = 'completed';

-- +migrate Down
DROP TABLE IF EXISTS point_transactions;
DROP TABLE IF EXISTS points;
DROP TABLE IF EXISTS coupon_usages;
DROP TABLE IF EXISTS coupons;
