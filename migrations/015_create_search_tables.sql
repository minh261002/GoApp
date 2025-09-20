-- Create search_queries table
CREATE TABLE search_queries (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NULL,
    query VARCHAR(500) NOT NULL,
    search_type ENUM('product', 'category', 'brand', 'user', 'wishlist', 'review', 'order', 'all') NOT NULL,
    filters JSON,
    sort_by VARCHAR(100),
    sort_order VARCHAR(10) DEFAULT 'desc',
    page INT DEFAULT 1,
    limit_count INT DEFAULT 10,
    results INT DEFAULT 0,
    duration BIGINT DEFAULT 0,
    ip_address VARCHAR(45),
    user_agent TEXT,
    referrer VARCHAR(500),
    status ENUM('active', 'inactive', 'pending', 'expired') DEFAULT 'active',
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_search_queries_user_id (user_id),
    INDEX idx_search_queries_query (query(255)),
    INDEX idx_search_queries_search_type (search_type),
    INDEX idx_search_queries_status (status),
    INDEX idx_search_queries_created_at (created_at),
    INDEX idx_search_queries_deleted_at (deleted_at),
    
    -- Foreign keys
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create search_filters table
CREATE TABLE search_filters (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    label VARCHAR(255) NOT NULL,
    description TEXT,
    type ENUM('range', 'select', 'multi_select', 'boolean', 'date', 'text') NOT NULL,
    field VARCHAR(100) NOT NULL,
    options JSON,
    min_value DECIMAL(10,2),
    max_value DECIMAL(10,2),
    default_value VARCHAR(255),
    is_required BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INT DEFAULT 0,
    search_type ENUM('product', 'category', 'brand', 'user', 'wishlist', 'review', 'order', 'all') NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_search_filters_name (name),
    INDEX idx_search_filters_type (type),
    INDEX idx_search_filters_field (field),
    INDEX idx_search_filters_search_type (search_type),
    INDEX idx_search_filters_is_active (is_active),
    INDEX idx_search_filters_sort_order (sort_order),
    INDEX idx_search_filters_deleted_at (deleted_at),
    
    -- Unique constraint
    UNIQUE KEY unique_filter_name_type (name, search_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create search_suggestions table
CREATE TABLE search_suggestions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    query VARCHAR(255) NOT NULL,
    suggestions JSON,
    search_type ENUM('product', 'category', 'brand', 'user', 'wishlist', 'review', 'order', 'all') NOT NULL,
    count BIGINT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_search_suggestions_query (query),
    INDEX idx_search_suggestions_search_type (search_type),
    INDEX idx_search_suggestions_count (count),
    INDEX idx_search_suggestions_is_active (is_active),
    INDEX idx_search_suggestions_deleted_at (deleted_at),
    
    -- Unique constraint
    UNIQUE KEY unique_query_type (query, search_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create search_history table
CREATE TABLE search_history (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    query VARCHAR(500) NOT NULL,
    search_type ENUM('product', 'category', 'brand', 'user', 'wishlist', 'review', 'order', 'all') NOT NULL,
    results INT DEFAULT 0,
    ip_address VARCHAR(45),
    user_agent TEXT,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_search_history_user_id (user_id),
    INDEX idx_search_history_query (query(255)),
    INDEX idx_search_history_search_type (search_type),
    INDEX idx_search_history_created_at (created_at),
    INDEX idx_search_history_deleted_at (deleted_at),
    
    -- Foreign keys
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create search_analytics table
CREATE TABLE search_analytics (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    date DATE NOT NULL,
    search_type ENUM('product', 'category', 'brand', 'user', 'wishlist', 'review', 'order', 'all') NOT NULL,
    total_searches BIGINT DEFAULT 0,
    unique_queries BIGINT DEFAULT 0,
    total_results BIGINT DEFAULT 0,
    zero_results BIGINT DEFAULT 0,
    avg_results DECIMAL(10,2) DEFAULT 0.00,
    avg_duration DECIMAL(10,2) DEFAULT 0.00,
    top_queries JSON,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_search_analytics_date (date),
    INDEX idx_search_analytics_search_type (search_type),
    INDEX idx_search_analytics_deleted_at (deleted_at),
    
    -- Unique constraint
    UNIQUE KEY unique_date_type (date, search_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create search_index table
CREATE TABLE search_index (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,
    entity_id BIGINT UNSIGNED NOT NULL,
    title VARCHAR(500) NOT NULL,
    content TEXT,
    keywords TEXT,
    tags JSON,
    metadata JSON,
    weight DECIMAL(5,2) DEFAULT 1.00,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_search_index_entity_type (entity_type),
    INDEX idx_search_index_entity_id (entity_id),
    INDEX idx_search_index_title (title(255)),
    INDEX idx_search_index_weight (weight),
    INDEX idx_search_index_is_active (is_active),
    INDEX idx_search_index_deleted_at (deleted_at),
    
    -- Full-text search index
    FULLTEXT KEY ft_search_content (title, content, keywords),
    
    -- Unique constraint
    UNIQUE KEY unique_entity (entity_type, entity_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create search_facets table for caching facets
CREATE TABLE search_facets (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    search_type ENUM('product', 'category', 'brand', 'user', 'wishlist', 'review', 'order', 'all') NOT NULL,
    filter_name VARCHAR(100) NOT NULL,
    facet_data JSON,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_search_facets_search_type (search_type),
    INDEX idx_search_facets_filter_name (filter_name),
    INDEX idx_search_facets_last_updated (last_updated),
    INDEX idx_search_facets_deleted_at (deleted_at),
    
    -- Unique constraint
    UNIQUE KEY unique_type_filter (search_type, filter_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default search filters for products
INSERT INTO search_filters (name, label, description, type, field, search_type, sort_order, is_active) VALUES
-- Product filters
('price', 'Price Range', 'Filter by price range', 'range', 'regular_price', 'product', 1, TRUE),
('brand', 'Brand', 'Filter by brand', 'multi_select', 'brand_id', 'product', 2, TRUE),
('category', 'Category', 'Filter by category', 'multi_select', 'category_id', 'product', 3, TRUE),
('status', 'Status', 'Filter by product status', 'select', 'status', 'product', 4, TRUE),
('in_stock', 'In Stock', 'Filter by stock availability', 'boolean', 'stock_quantity', 'product', 5, TRUE),
('rating', 'Rating', 'Filter by average rating', 'range', 'average_rating', 'product', 6, TRUE),
('created_date', 'Created Date', 'Filter by creation date', 'date', 'created_at', 'product', 7, TRUE),
('sale', 'On Sale', 'Filter by sale status', 'boolean', 'sale_price', 'product', 8, TRUE),

-- Category filters
('parent', 'Parent Category', 'Filter by parent category', 'select', 'parent_id', 'category', 1, TRUE),
('level', 'Category Level', 'Filter by category level', 'range', 'level', 'category', 2, TRUE),
('is_active', 'Active Status', 'Filter by active status', 'boolean', 'is_active', 'category', 3, TRUE),

-- Brand filters
('country', 'Country', 'Filter by brand country', 'select', 'country', 'brand', 1, TRUE),
('is_active', 'Active Status', 'Filter by active status', 'boolean', 'is_active', 'brand', 2, TRUE),

-- User filters
('role', 'User Role', 'Filter by user role', 'select', 'role', 'user', 1, TRUE),
('status', 'User Status', 'Filter by user status', 'select', 'status', 'user', 2, TRUE),
('created_date', 'Registration Date', 'Filter by registration date', 'date', 'created_at', 'user', 3, TRUE),

-- Wishlist filters
('is_public', 'Public Status', 'Filter by public status', 'boolean', 'is_public', 'wishlist', 1, TRUE),
('is_default', 'Default Status', 'Filter by default status', 'boolean', 'is_default', 'wishlist', 2, TRUE),
('created_date', 'Created Date', 'Filter by creation date', 'date', 'created_at', 'wishlist', 3, TRUE),

-- Review filters
('rating', 'Rating', 'Filter by rating', 'range', 'rating', 'review', 1, TRUE),
('status', 'Review Status', 'Filter by review status', 'select', 'status', 'review', 2, TRUE),
('verified', 'Verified Reviews', 'Filter by verification status', 'boolean', 'is_verified', 'review', 3, TRUE),
('created_date', 'Review Date', 'Filter by review date', 'date', 'created_at', 'review', 4, TRUE),

-- Order filters
('status', 'Order Status', 'Filter by order status', 'select', 'status', 'order', 1, TRUE),
('payment_status', 'Payment Status', 'Filter by payment status', 'select', 'payment_status', 'order', 2, TRUE),
('created_date', 'Order Date', 'Filter by order date', 'date', 'created_at', 'order', 3, TRUE),
('total_amount', 'Total Amount', 'Filter by total amount', 'range', 'total_amount', 'order', 4, TRUE);

-- Insert sample search suggestions
INSERT INTO search_suggestions (query, suggestions, search_type, count, is_active) VALUES
-- Product suggestions
('laptop', '["laptop gaming", "laptop dell", "laptop asus", "laptop hp", "laptop macbook"]', 'product', 150, TRUE),
('phone', '["phone samsung", "phone iphone", "phone xiaomi", "phone oppo", "phone vivo"]', 'product', 200, TRUE),
('shirt', '["shirt cotton", "shirt polo", "shirt t-shirt", "shirt long sleeve", "shirt short sleeve"]', 'product', 100, TRUE),
('shoes', '["shoes nike", "shoes adidas", "shoes running", "shoes casual", "shoes formal"]', 'product', 120, TRUE),
('watch', '["watch smart", "watch apple", "watch samsung", "watch luxury", "watch sport"]', 'product', 80, TRUE),

-- Category suggestions
('electronics', '["electronics computer", "electronics phone", "electronics camera", "electronics audio"]', 'category', 50, TRUE),
('clothing', '["clothing men", "clothing women", "clothing kids", "clothing accessories"]', 'category', 60, TRUE),
('home', '["home furniture", "home decor", "home kitchen", "home garden"]', 'category', 40, TRUE),

-- Brand suggestions
('apple', '["apple iphone", "apple macbook", "apple ipad", "apple watch", "apple airpods"]', 'brand', 100, TRUE),
('samsung', '["samsung galaxy", "samsung tv", "samsung refrigerator", "samsung washing machine"]', 'brand', 90, TRUE),
('nike', '["nike shoes", "nike clothing", "nike accessories", "nike sportswear"]', 'brand', 70, TRUE);

-- Insert sample search index entries
INSERT INTO search_index (entity_type, entity_id, title, content, keywords, tags, metadata, weight, is_active) VALUES
-- Product entries
('product', 1, 'MacBook Pro 16-inch', 'Apple MacBook Pro with M2 chip, 16GB RAM, 512GB SSD', 'macbook, apple, laptop, computer, m2, pro', '["laptop", "apple", "computer", "macbook"]', '{"brand": "Apple", "category": "Electronics", "price": 1999.99}', 1.0, TRUE),
('product', 2, 'iPhone 15 Pro', 'Latest iPhone with A17 Pro chip, 48MP camera, titanium design', 'iphone, apple, phone, smartphone, camera', '["phone", "apple", "smartphone", "camera"]', '{"brand": "Apple", "category": "Electronics", "price": 999.99}', 1.0, TRUE),
('product', 3, 'Nike Air Max 270', 'Comfortable running shoes with Air Max technology', 'nike, shoes, running, air max, sneakers', '["shoes", "nike", "running", "sneakers"]', '{"brand": "Nike", "category": "Footwear", "price": 150.00}', 1.0, TRUE),

-- Category entries
('category', 1, 'Electronics', 'Electronic devices and accessories', 'electronics, devices, gadgets, technology', '["electronics", "technology", "gadgets"]', '{"parent_id": null, "level": 1}', 1.0, TRUE),
('category', 2, 'Laptops', 'Portable computers and laptops', 'laptops, computers, portable, notebook', '["laptops", "computers", "portable"]', '{"parent_id": 1, "level": 2}', 1.0, TRUE),
('category', 3, 'Smartphones', 'Mobile phones and accessories', 'smartphones, phones, mobile, cell', '["smartphones", "phones", "mobile"]', '{"parent_id": 1, "level": 2}', 1.0, TRUE),

-- Brand entries
('brand', 1, 'Apple', 'Technology company known for innovative products', 'apple, technology, innovation, design', '["technology", "innovation", "design"]', '{"country": "USA", "founded": 1976}', 1.0, TRUE),
('brand', 2, 'Samsung', 'Korean electronics and technology company', 'samsung, korean, electronics, technology', '["electronics", "korean", "technology"]', '{"country": "South Korea", "founded": 1938}', 1.0, TRUE),
('brand', 3, 'Nike', 'American sportswear and footwear company', 'nike, sportswear, footwear, sports', '["sportswear", "footwear", "sports"]', '{"country": "USA", "founded": 1964}', 1.0, TRUE);

-- Create indexes for better performance
CREATE INDEX idx_search_queries_composite ON search_queries (search_type, status, created_at);
CREATE INDEX idx_search_history_composite ON search_history (user_id, search_type, created_at);
CREATE INDEX idx_search_analytics_composite ON search_analytics (date, search_type);
CREATE INDEX idx_search_index_composite ON search_index (entity_type, is_active, weight);
CREATE INDEX idx_search_suggestions_composite ON search_suggestions (search_type, is_active, count DESC);
