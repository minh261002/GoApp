-- Create banners table
CREATE TABLE banners (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    type ENUM('image', 'video', 'carousel', 'text') NOT NULL,
    position ENUM('header', 'footer', 'sidebar', 'main', 'popup', 'mobile', 'desktop', 'category', 'product') NOT NULL,
    status ENUM('active', 'inactive', 'draft', 'expired') NOT NULL DEFAULT 'draft',
    
    -- Content
    image_url VARCHAR(500),
    video_url VARCHAR(500),
    text_content TEXT,
    button_text VARCHAR(100),
    button_url VARCHAR(500),
    
    -- Display settings
    width INT DEFAULT 0,
    height INT DEFAULT 0,
    alt_text VARCHAR(255),
    css_class VARCHAR(255),
    sort_order INT DEFAULT 0,
    
    -- Targeting
    target_audience VARCHAR(255),
    device_type VARCHAR(50),
    location VARCHAR(255),
    
    -- Scheduling
    start_date DATETIME NULL,
    end_date DATETIME NULL,
    
    -- Analytics
    click_count BIGINT DEFAULT 0,
    view_count BIGINT DEFAULT 0,
    impression_count BIGINT DEFAULT 0,
    
    -- SEO
    meta_title VARCHAR(255),
    meta_description TEXT,
    meta_keywords VARCHAR(500),
    
    -- Relationships
    created_by BIGINT UNSIGNED NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_banners_type (type),
    INDEX idx_banners_position (position),
    INDEX idx_banners_status (status),
    INDEX idx_banners_created_by (created_by),
    INDEX idx_banners_sort_order (sort_order),
    INDEX idx_banners_start_date (start_date),
    INDEX idx_banners_end_date (end_date),
    INDEX idx_banners_deleted_at (deleted_at),
    INDEX idx_banners_target_audience (target_audience),
    INDEX idx_banners_device_type (device_type),
    INDEX idx_banners_location (location),
    
    -- Foreign keys
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create sliders table
CREATE TABLE sliders (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type ENUM('image', 'video', 'mixed') NOT NULL,
    status ENUM('active', 'inactive', 'draft') NOT NULL DEFAULT 'draft',
    
    -- Display settings
    width INT DEFAULT 0,
    height INT DEFAULT 0,
    auto_play BOOLEAN DEFAULT TRUE,
    auto_play_delay INT DEFAULT 5000,
    show_dots BOOLEAN DEFAULT TRUE,
    show_arrows BOOLEAN DEFAULT TRUE,
    infinite_loop BOOLEAN DEFAULT TRUE,
    fade_effect BOOLEAN DEFAULT FALSE,
    css_class VARCHAR(255),
    sort_order INT DEFAULT 0,
    
    -- Targeting
    target_audience VARCHAR(255),
    device_type VARCHAR(50),
    location VARCHAR(255),
    
    -- Scheduling
    start_date DATETIME NULL,
    end_date DATETIME NULL,
    
    -- Analytics
    view_count BIGINT DEFAULT 0,
    impression_count BIGINT DEFAULT 0,
    
    -- SEO
    meta_title VARCHAR(255),
    meta_description TEXT,
    meta_keywords VARCHAR(500),
    
    -- Relationships
    created_by BIGINT UNSIGNED NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_sliders_type (type),
    INDEX idx_sliders_status (status),
    INDEX idx_sliders_created_by (created_by),
    INDEX idx_sliders_sort_order (sort_order),
    INDEX idx_sliders_start_date (start_date),
    INDEX idx_sliders_end_date (end_date),
    INDEX idx_sliders_deleted_at (deleted_at),
    INDEX idx_sliders_target_audience (target_audience),
    INDEX idx_sliders_device_type (device_type),
    INDEX idx_sliders_location (location),
    
    -- Foreign keys
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create slider_items table
CREATE TABLE slider_items (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    slider_id BIGINT UNSIGNED NOT NULL,
    title VARCHAR(255),
    description TEXT,
    
    -- Content
    image_url VARCHAR(500),
    video_url VARCHAR(500),
    text_content TEXT,
    button_text VARCHAR(100),
    button_url VARCHAR(500),
    
    -- Display settings
    width INT DEFAULT 0,
    height INT DEFAULT 0,
    alt_text VARCHAR(255),
    css_class VARCHAR(255),
    sort_order INT DEFAULT 0,
    
    -- Analytics
    click_count BIGINT DEFAULT 0,
    view_count BIGINT DEFAULT 0,
    impression_count BIGINT DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_slider_items_slider_id (slider_id),
    INDEX idx_slider_items_sort_order (sort_order),
    INDEX idx_slider_items_deleted_at (deleted_at),
    
    -- Foreign keys
    FOREIGN KEY (slider_id) REFERENCES sliders(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create banner_clicks table for tracking banner clicks
CREATE TABLE banner_clicks (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    banner_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    referrer VARCHAR(500),
    clicked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes
    INDEX idx_banner_clicks_banner_id (banner_id),
    INDEX idx_banner_clicks_user_id (user_id),
    INDEX idx_banner_clicks_clicked_at (clicked_at),
    
    -- Foreign keys
    FOREIGN KEY (banner_id) REFERENCES banners(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create banner_views table for tracking banner views
CREATE TABLE banner_views (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    banner_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    referrer VARCHAR(500),
    viewed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes
    INDEX idx_banner_views_banner_id (banner_id),
    INDEX idx_banner_views_user_id (user_id),
    INDEX idx_banner_views_viewed_at (viewed_at),
    
    -- Foreign keys
    FOREIGN KEY (banner_id) REFERENCES banners(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create slider_views table for tracking slider views
CREATE TABLE slider_views (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    slider_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    referrer VARCHAR(500),
    viewed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes
    INDEX idx_slider_views_slider_id (slider_id),
    INDEX idx_slider_views_user_id (user_id),
    INDEX idx_slider_views_viewed_at (viewed_at),
    
    -- Foreign keys
    FOREIGN KEY (slider_id) REFERENCES sliders(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create slider_item_clicks table for tracking slider item clicks
CREATE TABLE slider_item_clicks (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    item_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    referrer VARCHAR(500),
    clicked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes
    INDEX idx_slider_item_clicks_item_id (item_id),
    INDEX idx_slider_item_clicks_user_id (user_id),
    INDEX idx_slider_item_clicks_clicked_at (clicked_at),
    
    -- Foreign keys
    FOREIGN KEY (item_id) REFERENCES slider_items(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert sample data
INSERT INTO banners (title, description, type, position, status, image_url, button_text, button_url, width, height, sort_order, created_by) VALUES
('Welcome Banner', 'Welcome to our store', 'image', 'main', 'active', '/images/banner1.jpg', 'Shop Now', '/products', 1200, 400, 1, 1),
('Sale Banner', '50% Off Sale', 'image', 'header', 'active', '/images/banner2.jpg', 'Get Deal', '/sale', 800, 200, 2, 1),
('Newsletter Banner', 'Subscribe to our newsletter', 'text', 'popup', 'active', '', 'Subscribe', '/newsletter', 0, 0, 3, 1);

INSERT INTO sliders (name, description, type, status, width, height, auto_play, auto_play_delay, show_dots, show_arrows, infinite_loop, sort_order, created_by) VALUES
('Homepage Slider', 'Main slider for homepage', 'image', 'active', 1200, 500, TRUE, 5000, TRUE, TRUE, TRUE, 1, 1),
('Product Slider', 'Product showcase slider', 'mixed', 'active', 800, 400, TRUE, 3000, TRUE, TRUE, TRUE, 2, 1);

INSERT INTO slider_items (slider_id, title, description, image_url, button_text, button_url, width, height, sort_order) VALUES
(1, 'New Collection', 'Discover our latest collection', '/images/slider1.jpg', 'Shop Now', '/collections/new', 1200, 500, 1),
(1, 'Summer Sale', 'Up to 70% off summer items', '/images/slider2.jpg', 'Get Deal', '/sale', 1200, 500, 2),
(1, 'Free Shipping', 'Free shipping on orders over $100', '/images/slider3.jpg', 'Learn More', '/shipping', 1200, 500, 3),
(2, 'Featured Product 1', 'Amazing product description', '/images/product1.jpg', 'View Product', '/products/1', 400, 400, 1),
(2, 'Featured Product 2', 'Another great product', '/images/product2.jpg', 'View Product', '/products/2', 400, 400, 2);
