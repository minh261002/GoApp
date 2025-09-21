-- Create products table
CREATE TABLE IF NOT EXISTS products (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(300) NOT NULL UNIQUE,
    description TEXT,
    short_description VARCHAR(500),
    sku VARCHAR(100) UNIQUE,
    type ENUM('simple', 'variable') DEFAULT 'simple',
    status ENUM('draft', 'active', 'inactive', 'archived') DEFAULT 'draft',
    
    -- Pricing
    regular_price DECIMAL(10,2) DEFAULT 0.00,
    sale_price DECIMAL(10,2) NULL,
    cost_price DECIMAL(10,2) NULL,
    
    -- Inventory
    manage_stock BOOLEAN DEFAULT TRUE,
    stock_quantity INT DEFAULT 0,
    low_stock_threshold INT DEFAULT 5,
    stock_status VARCHAR(20) DEFAULT 'instock',
    
    -- Dimensions & Weight
    weight DECIMAL(8,2) NULL,
    length DECIMAL(8,2) NULL,
    width DECIMAL(8,2) NULL,
    height DECIMAL(8,2) NULL,
    
    -- Media
    images TEXT, -- JSON array of image URLs
    featured_image VARCHAR(500),
    
    -- SEO
    meta_title VARCHAR(255),
    meta_description VARCHAR(500),
    meta_keywords VARCHAR(500),
    
    -- Relationships
    brand_id INT UNSIGNED NULL,
    category_id INT UNSIGNED NULL,
    
    -- Settings
    is_featured BOOLEAN DEFAULT FALSE,
    is_digital BOOLEAN DEFAULT FALSE,
    requires_shipping BOOLEAN DEFAULT TRUE,
    is_downloadable BOOLEAN DEFAULT FALSE,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_products_name (name),
    INDEX idx_products_slug (slug),
    INDEX idx_products_sku (sku),
    INDEX idx_products_type (type),
    INDEX idx_products_status (status),
    INDEX idx_products_brand_id (brand_id),
    INDEX idx_products_category_id (category_id),
    INDEX idx_products_is_featured (is_featured),
    INDEX idx_products_stock_status (stock_status),
    INDEX idx_products_deleted_at (deleted_at),
    
    -- Foreign keys
    FOREIGN KEY (brand_id) REFERENCES brands(id) ON DELETE SET NULL,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create product_variants table
CREATE TABLE IF NOT EXISTS product_variants (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    product_id BIGINT UNSIGNED NOT NULL,
    
    -- Variant Info
    name VARCHAR(255) NOT NULL,
    sku VARCHAR(100) UNIQUE,
    description TEXT,
    
    -- Pricing
    regular_price DECIMAL(10,2) NOT NULL,
    sale_price DECIMAL(10,2) NULL,
    cost_price DECIMAL(10,2) NULL,
    
    -- Inventory
    stock_quantity INT DEFAULT 0,
    stock_status VARCHAR(20) DEFAULT 'instock',
    manage_stock BOOLEAN DEFAULT TRUE,
    
    -- Media
    image VARCHAR(500),
    
    -- Attributes (JSON format: {"size": "L", "color": "Red"})
    attributes TEXT,
    
    -- Settings
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INT DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_product_variants_product_id (product_id),
    INDEX idx_product_variants_sku (sku),
    INDEX idx_product_variants_is_active (is_active),
    INDEX idx_product_variants_sort_order (sort_order),
    INDEX idx_product_variants_deleted_at (deleted_at),
    
    -- Foreign key
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create product_attributes table
CREATE TABLE IF NOT EXISTS product_attributes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    product_id BIGINT UNSIGNED NOT NULL,
    
    -- Attribute Info
    name VARCHAR(100) NOT NULL,
    value VARCHAR(255) NOT NULL,
    slug VARCHAR(120) NOT NULL,
    
    -- Settings
    is_visible BOOLEAN DEFAULT TRUE,
    is_variation BOOLEAN DEFAULT FALSE,
    sort_order INT DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_product_attributes_product_id (product_id),
    INDEX idx_product_attributes_name (name),
    INDEX idx_product_attributes_slug (slug),
    INDEX idx_product_attributes_is_visible (is_visible),
    INDEX idx_product_attributes_is_variation (is_variation),
    INDEX idx_product_attributes_sort_order (sort_order),
    INDEX idx_product_attributes_deleted_at (deleted_at),
    
    -- Foreign key
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert some sample products
INSERT INTO products (name, slug, description, short_description, sku, type, status, regular_price, stock_quantity, brand_id, category_id, is_featured) VALUES
('iPhone 15 Pro', 'iphone-15-pro', 'Latest iPhone with advanced features', 'Premium smartphone', 'IPH15PRO-001', 'simple', 'active', 999.00, 50, 1, 10, TRUE),
('Samsung Galaxy S24', 'samsung-galaxy-s24', 'Flagship Android smartphone', 'High-end Android phone', 'SGS24-001', 'simple', 'active', 899.00, 30, 2, 11, TRUE),
('Nike Air Max 270', 'nike-air-max-270', 'Comfortable running shoes', 'Athletic footwear', 'NAM270-001', 'variable', 'active', 150.00, 0, 3, 19, FALSE),
('Adidas Ultraboost 22', 'adidas-ultraboost-22', 'Premium running shoes', 'Performance running', 'AUB22-001', 'variable', 'active', 180.00, 0, 4, 19, FALSE);

-- Insert sample product variants for Nike Air Max 270
INSERT INTO product_variants (product_id, name, sku, regular_price, stock_quantity, attributes, is_active, sort_order) VALUES
(3, 'Nike Air Max 270 - Black/White - Size 8', 'NAM270-001-BW-8', 150.00, 10, '{"color": "Black/White", "size": "8"}', TRUE, 1),
(3, 'Nike Air Max 270 - Black/White - Size 9', 'NAM270-001-BW-9', 150.00, 15, '{"color": "Black/White", "size": "9"}', TRUE, 2),
(3, 'Nike Air Max 270 - Black/White - Size 10', 'NAM270-001-BW-10', 150.00, 12, '{"color": "Black/White", "size": "10"}', TRUE, 3),
(3, 'Nike Air Max 270 - White/Black - Size 8', 'NAM270-001-WB-8', 150.00, 8, '{"color": "White/Black", "size": "8"}', TRUE, 4),
(3, 'Nike Air Max 270 - White/Black - Size 9', 'NAM270-001-WB-9', 150.00, 20, '{"color": "White/Black", "size": "9"}', TRUE, 5),
(3, 'Nike Air Max 270 - White/Black - Size 10', 'NAM270-001-WB-10', 150.00, 18, '{"color": "White/Black", "size": "10"}', TRUE, 6);

-- Insert sample product variants for Adidas Ultraboost 22
INSERT INTO product_variants (product_id, name, sku, regular_price, stock_quantity, attributes, is_active, sort_order) VALUES
(4, 'Adidas Ultraboost 22 - Core Black - Size 8', 'AUB22-001-CB-8', 180.00, 5, '{"color": "Core Black", "size": "8"}', TRUE, 1),
(4, 'Adidas Ultraboost 22 - Core Black - Size 9', 'AUB22-001-CB-9', 180.00, 8, '{"color": "Core Black", "size": "9"}', TRUE, 2),
(4, 'Adidas Ultraboost 22 - Core Black - Size 10', 'AUB22-001-CB-10', 180.00, 6, '{"color": "Core Black", "size": "10"}', TRUE, 3),
(4, 'Adidas Ultraboost 22 - Cloud White - Size 8', 'AUB22-001-CW-8', 180.00, 3, '{"color": "Cloud White", "size": "8"}', TRUE, 4),
(4, 'Adidas Ultraboost 22 - Cloud White - Size 9', 'AUB22-001-CW-9', 180.00, 7, '{"color": "Cloud White", "size": "9"}', TRUE, 5),
(4, 'Adidas Ultraboost 22 - Cloud White - Size 10', 'AUB22-001-CW-10', 180.00, 4, '{"color": "Cloud White", "size": "10"}', TRUE, 6);

-- Insert sample product attributes
INSERT INTO product_attributes (product_id, name, value, slug, is_visible, is_variation, sort_order) VALUES
-- iPhone 15 Pro attributes
(1, 'Color', 'Space Black', 'color', TRUE, FALSE, 1),
(1, 'Storage', '256GB', 'storage', TRUE, FALSE, 2),
(1, 'Screen Size', '6.1 inch', 'screen-size', TRUE, FALSE, 3),
(1, 'Operating System', 'iOS 17', 'operating-system', TRUE, FALSE, 4),

-- Samsung Galaxy S24 attributes
(2, 'Color', 'Titanium Gray', 'color', TRUE, FALSE, 1),
(2, 'Storage', '512GB', 'storage', TRUE, FALSE, 2),
(2, 'Screen Size', '6.2 inch', 'screen-size', TRUE, FALSE, 3),
(2, 'Operating System', 'Android 14', 'operating-system', TRUE, FALSE, 4),

-- Nike Air Max 270 attributes (variation attributes)
(3, 'Color', 'Black/White', 'color', TRUE, TRUE, 1),
(3, 'Size', '8,9,10', 'size', TRUE, TRUE, 2),
(3, 'Material', 'Mesh and Synthetic', 'material', TRUE, FALSE, 3),
(3, 'Gender', 'Unisex', 'gender', TRUE, FALSE, 4),

-- Adidas Ultraboost 22 attributes (variation attributes)
(4, 'Color', 'Core Black,Cloud White', 'color', TRUE, TRUE, 1),
(4, 'Size', '8,9,10', 'size', TRUE, TRUE, 2),
(4, 'Material', 'Primeknit+', 'material', TRUE, FALSE, 3),
(4, 'Gender', 'Unisex', 'gender', TRUE, FALSE, 4);
