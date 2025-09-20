-- Create wishlists table
CREATE TABLE wishlists (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status ENUM('active', 'inactive', 'private', 'public') NOT NULL DEFAULT 'active',
    is_default BOOLEAN DEFAULT FALSE,
    is_public BOOLEAN DEFAULT FALSE,
    sort_order INT DEFAULT 0,
    
    -- Analytics
    view_count BIGINT DEFAULT 0,
    share_count BIGINT DEFAULT 0,
    item_count BIGINT DEFAULT 0,
    
    -- SEO
    slug VARCHAR(255) UNIQUE,
    meta_title VARCHAR(255),
    meta_description TEXT,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_wishlists_user_id (user_id),
    INDEX idx_wishlists_status (status),
    INDEX idx_wishlists_is_public (is_public),
    INDEX idx_wishlists_is_default (is_default),
    INDEX idx_wishlists_slug (slug),
    INDEX idx_wishlists_deleted_at (deleted_at),
    
    -- Foreign keys
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create wishlist_items table
CREATE TABLE wishlist_items (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    wishlist_id BIGINT UNSIGNED NOT NULL,
    product_id BIGINT UNSIGNED NOT NULL,
    status ENUM('active', 'inactive', 'purchased', 'removed') NOT NULL DEFAULT 'active',
    quantity INT DEFAULT 1,
    notes TEXT,
    priority INT DEFAULT 0,
    sort_order INT DEFAULT 0,
    
    -- Price tracking
    added_price DECIMAL(10,2) DEFAULT 0.00,
    current_price DECIMAL(10,2) DEFAULT 0.00,
    price_change DECIMAL(10,2) DEFAULT 0.00,
    price_change_percent DECIMAL(5,2) DEFAULT 0.00,
    
    -- Notifications
    notify_on_price_drop BOOLEAN DEFAULT TRUE,
    notify_on_stock BOOLEAN DEFAULT TRUE,
    notify_on_sale BOOLEAN DEFAULT TRUE,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_wishlist_items_wishlist_id (wishlist_id),
    INDEX idx_wishlist_items_product_id (product_id),
    INDEX idx_wishlist_items_status (status),
    INDEX idx_wishlist_items_priority (priority),
    INDEX idx_wishlist_items_deleted_at (deleted_at),
    UNIQUE KEY unique_wishlist_product (wishlist_id, product_id),
    
    -- Foreign keys
    FOREIGN KEY (wishlist_id) REFERENCES wishlists(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create favorites table
CREATE TABLE favorites (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    product_id BIGINT UNSIGNED NOT NULL,
    notes TEXT,
    priority INT DEFAULT 0,
    
    -- Notifications
    notify_on_price_drop BOOLEAN DEFAULT TRUE,
    notify_on_stock BOOLEAN DEFAULT TRUE,
    notify_on_sale BOOLEAN DEFAULT TRUE,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_favorites_user_id (user_id),
    INDEX idx_favorites_product_id (product_id),
    INDEX idx_favorites_priority (priority),
    INDEX idx_favorites_deleted_at (deleted_at),
    UNIQUE KEY unique_user_product (user_id, product_id),
    
    -- Foreign keys
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create wishlist_shares table
CREATE TABLE wishlist_shares (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    wishlist_id BIGINT UNSIGNED NOT NULL,
    shared_by BIGINT UNSIGNED NOT NULL,
    shared_with BIGINT UNSIGNED NOT NULL,
    token VARCHAR(255) UNIQUE,
    is_active BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP NULL,
    
    -- Permissions
    can_view BOOLEAN DEFAULT TRUE,
    can_edit BOOLEAN DEFAULT FALSE,
    can_delete BOOLEAN DEFAULT FALSE,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_wishlist_shares_wishlist_id (wishlist_id),
    INDEX idx_wishlist_shares_shared_by (shared_by),
    INDEX idx_wishlist_shares_shared_with (shared_with),
    INDEX idx_wishlist_shares_token (token),
    INDEX idx_wishlist_shares_is_active (is_active),
    INDEX idx_wishlist_shares_expires_at (expires_at),
    INDEX idx_wishlist_shares_deleted_at (deleted_at),
    
    -- Foreign keys
    FOREIGN KEY (wishlist_id) REFERENCES wishlists(id) ON DELETE CASCADE,
    FOREIGN KEY (shared_by) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (shared_with) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create wishlist_views table for tracking wishlist views
CREATE TABLE wishlist_views (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    wishlist_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    referrer VARCHAR(500),
    viewed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes
    INDEX idx_wishlist_views_wishlist_id (wishlist_id),
    INDEX idx_wishlist_views_user_id (user_id),
    INDEX idx_wishlist_views_viewed_at (viewed_at),
    
    -- Foreign keys
    FOREIGN KEY (wishlist_id) REFERENCES wishlists(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create wishlist_item_views table for tracking wishlist item views
CREATE TABLE wishlist_item_views (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    item_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    referrer VARCHAR(500),
    viewed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes
    INDEX idx_wishlist_item_views_item_id (item_id),
    INDEX idx_wishlist_item_views_user_id (user_id),
    INDEX idx_wishlist_item_views_viewed_at (viewed_at),
    
    -- Foreign keys
    FOREIGN KEY (item_id) REFERENCES wishlist_items(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create wishlist_item_clicks table for tracking wishlist item clicks
CREATE TABLE wishlist_item_clicks (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    item_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    referrer VARCHAR(500),
    clicked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes
    INDEX idx_wishlist_item_clicks_item_id (item_id),
    INDEX idx_wishlist_item_clicks_user_id (user_id),
    INDEX idx_wishlist_item_clicks_clicked_at (clicked_at),
    
    -- Foreign keys
    FOREIGN KEY (item_id) REFERENCES wishlist_items(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert sample data
INSERT INTO wishlists (user_id, name, description, status, is_default, is_public, slug, meta_title, meta_description) VALUES
(1, 'My Wishlist', 'My personal wishlist', 'active', TRUE, FALSE, 'my-wishlist', 'My Wishlist', 'My personal wishlist'),
(1, 'Gift Ideas', 'Gift ideas for family and friends', 'active', FALSE, TRUE, 'gift-ideas', 'Gift Ideas', 'Gift ideas for family and friends'),
(2, 'Tech Wishlist', 'Technology products I want', 'active', TRUE, FALSE, 'tech-wishlist', 'Tech Wishlist', 'Technology products I want');

INSERT INTO wishlist_items (wishlist_id, product_id, status, quantity, notes, priority, added_price, current_price) VALUES
(1, 1, 'active', 1, 'Great product!', 1, 99.99, 99.99),
(1, 2, 'active', 2, 'Need this for work', 0, 149.99, 149.99),
(2, 3, 'active', 1, 'Perfect gift for mom', 2, 79.99, 79.99),
(3, 4, 'active', 1, 'Latest tech', 1, 299.99, 299.99);

INSERT INTO favorites (user_id, product_id, notes, priority) VALUES
(1, 1, 'My favorite product', 1),
(1, 3, 'Love this one', 0),
(2, 2, 'Must have', 2),
(2, 4, 'Awesome tech', 1);
