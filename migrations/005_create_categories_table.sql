-- Create categories table with hierarchical structure
CREATE TABLE IF NOT EXISTS categories (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(120) NOT NULL UNIQUE,
    description TEXT,
    image VARCHAR(255),
    icon VARCHAR(100),
    parent_id INT UNSIGNED NULL,
    level INT DEFAULT 0,
    path VARCHAR(500) DEFAULT '',
    sort_order INT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    is_leaf BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes for performance
    INDEX idx_categories_name (name),
    INDEX idx_categories_slug (slug),
    INDEX idx_categories_parent_id (parent_id),
    INDEX idx_categories_level (level),
    INDEX idx_categories_path (path),
    INDEX idx_categories_is_active (is_active),
    INDEX idx_categories_sort_order (sort_order),
    INDEX idx_categories_deleted_at (deleted_at),
    
    -- Foreign key constraint
    FOREIGN KEY (parent_id) REFERENCES categories(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert some sample categories for testing
INSERT INTO categories (name, slug, description, level, path, sort_order, is_active, is_leaf) VALUES
('Electronics', 'electronics', 'Electronic devices and gadgets', 0, '1', 1, TRUE, FALSE),
('Clothing', 'clothing', 'Fashion and apparel', 0, '2', 2, TRUE, FALSE),
('Books', 'books', 'Books and publications', 0, '3', 3, TRUE, FALSE),
('Home & Garden', 'home-garden', 'Home and garden products', 0, '4', 4, TRUE, FALSE),
('Sports', 'sports', 'Sports and fitness equipment', 0, '5', 5, TRUE, FALSE);

-- Insert subcategories for Electronics
INSERT INTO categories (name, slug, description, parent_id, level, path, sort_order, is_active, is_leaf) VALUES
('Smartphones', 'smartphones', 'Mobile phones and accessories', 1, 1, '1/6', 1, TRUE, FALSE),
('Laptops', 'laptops', 'Laptop computers and accessories', 1, 1, '1/7', 2, TRUE, FALSE),
('Tablets', 'tablets', 'Tablet computers and accessories', 1, 1, '1/8', 3, TRUE, FALSE),
('Audio', 'audio', 'Audio equipment and accessories', 1, 1, '1/9', 4, TRUE, FALSE);

-- Insert subcategories for Smartphones
INSERT INTO categories (name, slug, description, parent_id, level, path, sort_order, is_active, is_leaf) VALUES
('iPhone', 'iphone', 'Apple iPhone smartphones', 6, 2, '1/6/10', 1, TRUE, TRUE),
('Samsung Galaxy', 'samsung-galaxy', 'Samsung Galaxy smartphones', 6, 2, '1/6/11', 2, TRUE, TRUE),
('Google Pixel', 'google-pixel', 'Google Pixel smartphones', 6, 2, '1/6/12', 3, TRUE, TRUE),
('OnePlus', 'oneplus', 'OnePlus smartphones', 6, 2, '1/6/13', 4, TRUE, TRUE);

-- Insert subcategories for Clothing
INSERT INTO categories (name, slug, description, parent_id, level, path, sort_order, is_active, is_leaf) VALUES
('Men\'s Clothing', 'mens-clothing', 'Men\'s fashion and apparel', 2, 1, '2/14', 1, TRUE, FALSE),
('Women\'s Clothing', 'womens-clothing', 'Women\'s fashion and apparel', 2, 1, '2/15', 2, TRUE, FALSE),
('Kids\' Clothing', 'kids-clothing', 'Children\'s fashion and apparel', 2, 1, '2/16', 3, TRUE, FALSE);

-- Insert subcategories for Men's Clothing
INSERT INTO categories (name, slug, description, parent_id, level, path, sort_order, is_active, is_leaf) VALUES
('Shirts', 'mens-shirts', 'Men\'s shirts and tops', 14, 2, '2/14/17', 1, TRUE, TRUE),
('Pants', 'mens-pants', 'Men\'s pants and trousers', 14, 2, '2/14/18', 2, TRUE, TRUE),
('Shoes', 'mens-shoes', 'Men\'s footwear', 14, 2, '2/14/19', 3, TRUE, TRUE),
('Accessories', 'mens-accessories', 'Men\'s accessories', 14, 2, '2/14/20', 4, TRUE, TRUE);
