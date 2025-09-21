-- Create advanced cart features tables

-- Update cart_items table to add advanced features
ALTER TABLE cart_items 
ADD COLUMN is_saved_for_later BOOLEAN DEFAULT FALSE,
ADD COLUMN notes TEXT,
ADD COLUMN priority INT DEFAULT 0;

-- Create cart_shares table
CREATE TABLE IF NOT EXISTS cart_shares (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    cart_id INT UNSIGNED NOT NULL,
    shared_by INT UNSIGNED NOT NULL,
    
    -- Share Information
    token VARCHAR(255) NOT NULL UNIQUE,
    is_active BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP NOT NULL,
    max_uses INT DEFAULT 0, -- 0 = unlimited
    used_count INT DEFAULT 0,
    
    -- Permissions
    can_view BOOLEAN DEFAULT TRUE,
    can_edit BOOLEAN DEFAULT FALSE,
    can_delete BOOLEAN DEFAULT FALSE,
    
    -- Access Control
    password_protected BOOLEAN DEFAULT FALSE,
    password VARCHAR(255), -- Hashed password
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_cart_shares_cart_id (cart_id),
    INDEX idx_cart_shares_shared_by (shared_by),
    INDEX idx_cart_shares_token (token),
    INDEX idx_cart_shares_is_active (is_active),
    INDEX idx_cart_shares_expires_at (expires_at),
    INDEX idx_cart_shares_deleted_at (deleted_at),
    
    -- Foreign keys
    FOREIGN KEY (cart_id) REFERENCES carts(id) ON DELETE CASCADE,
    FOREIGN KEY (shared_by) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create saved_for_later table
CREATE TABLE IF NOT EXISTS saved_for_later (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL,
    product_id INT UNSIGNED NOT NULL,
    product_variant_id INT UNSIGNED NULL,
    
    -- Item Information
    quantity INT NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    total_price DECIMAL(10,2) NOT NULL,
    
    -- Additional Information
    notes TEXT,
    priority INT DEFAULT 0,
    remind_at TIMESTAMP NULL,
    
    -- Notifications
    notify_on_price_drop BOOLEAN DEFAULT TRUE,
    notify_on_stock BOOLEAN DEFAULT TRUE,
    notify_on_sale BOOLEAN DEFAULT TRUE,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_saved_for_later_user_id (user_id),
    INDEX idx_saved_for_later_product_id (product_id),
    INDEX idx_saved_for_later_product_variant_id (product_variant_id),
    INDEX idx_saved_for_later_priority (priority),
    INDEX idx_saved_for_later_remind_at (remind_at),
    INDEX idx_saved_for_later_deleted_at (deleted_at),
    
    -- Foreign keys
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    FOREIGN KEY (product_variant_id) REFERENCES product_variants(id) ON DELETE CASCADE,
    
    -- Unique constraint
    UNIQUE KEY unique_user_product_variant (user_id, product_id, product_variant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
