-- +migrate Up
CREATE TABLE IF NOT EXISTS reviews (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    product_id BIGINT UNSIGNED NOT NULL,
    order_id INT UNSIGNED NULL,
    
    -- Review Content
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    rating INT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    type VARCHAR(20) DEFAULT 'product',
    
    -- Review Details
    is_verified BOOLEAN DEFAULT FALSE,
    is_helpful INT DEFAULT 0,
    is_not_helpful INT DEFAULT 0,
    is_anonymous BOOLEAN DEFAULT FALSE,
    
    -- Moderation
    moderated_by INT UNSIGNED NULL,
    moderated_at TIMESTAMP NULL,
    moderation_note TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE SET NULL,
    FOREIGN KEY (moderated_by) REFERENCES users(id) ON DELETE SET NULL,
    
    INDEX idx_reviews_user_id (user_id),
    INDEX idx_reviews_product_id (product_id),
    INDEX idx_reviews_order_id (order_id),
    INDEX idx_reviews_status (status),
    INDEX idx_reviews_type (type),
    INDEX idx_reviews_rating (rating),
    INDEX idx_reviews_is_verified (is_verified),
    INDEX idx_reviews_created_at (created_at),
    INDEX idx_reviews_deleted_at (deleted_at)
);

-- Create Review Images Table
CREATE TABLE IF NOT EXISTS review_images (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    review_id BIGINT UNSIGNED NOT NULL,
    image_url VARCHAR(500) NOT NULL,
    image_path VARCHAR(500) NOT NULL,
    alt_text VARCHAR(255),
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (review_id) REFERENCES reviews(id) ON DELETE CASCADE,
    
    INDEX idx_review_images_review_id (review_id),
    INDEX idx_review_images_sort_order (sort_order),
    INDEX idx_review_images_deleted_at (deleted_at)
);

-- Create Review Helpful Votes Table
CREATE TABLE IF NOT EXISTS review_helpful_votes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    review_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    is_helpful BOOLEAN NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (review_id) REFERENCES reviews(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    UNIQUE KEY unique_user_review_vote (user_id, review_id),
    INDEX idx_review_helpful_votes_review_id (review_id),
    INDEX idx_review_helpful_votes_user_id (user_id),
    INDEX idx_review_helpful_votes_is_helpful (is_helpful),
    INDEX idx_review_helpful_votes_deleted_at (deleted_at)
);

-- Add constraints
ALTER TABLE reviews ADD CONSTRAINT chk_review_rating 
CHECK (rating >= 1 AND rating <= 5);

ALTER TABLE reviews ADD CONSTRAINT chk_review_status 
CHECK (status IN ('pending', 'approved', 'rejected', 'hidden'));

ALTER TABLE reviews ADD CONSTRAINT chk_review_type 
CHECK (type IN ('product', 'service', 'order', 'overall'));

-- Create indexes for better performance
CREATE INDEX idx_reviews_product_status ON reviews(product_id, status);
CREATE INDEX idx_reviews_user_status ON reviews(user_id, status);
CREATE INDEX idx_reviews_rating_status ON reviews(rating, status);
CREATE INDEX idx_reviews_created_status ON reviews(created_at, status);

-- Create full-text search index
CREATE FULLTEXT INDEX idx_reviews_fulltext ON reviews(title, content);

-- Create composite indexes for common queries
CREATE INDEX idx_reviews_product_rating ON reviews(product_id, rating, status);
CREATE INDEX idx_reviews_user_created ON reviews(user_id, created_at DESC);

-- +migrate Down
DROP TABLE IF EXISTS review_helpful_votes;
DROP TABLE IF EXISTS review_images;
DROP TABLE IF EXISTS reviews;
