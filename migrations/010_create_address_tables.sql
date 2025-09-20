-- +migrate Up
CREATE TABLE IF NOT EXISTS addresses (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL,
    
    -- Address Information
    type VARCHAR(20) DEFAULT 'home',
    is_default BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Contact Information
    full_name VARCHAR(255) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    email VARCHAR(255),
    
    -- Address Details
    address_line1 VARCHAR(255) NOT NULL,
    address_line2 VARCHAR(255),
    ward VARCHAR(100) NOT NULL,
    district VARCHAR(100) NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100),
    country VARCHAR(100) DEFAULT 'Vietnam',
    postal_code VARCHAR(20),
    
    -- Geographic Information
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    
    -- Additional Information
    landmark VARCHAR(255),
    instructions TEXT,
    notes TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    INDEX idx_addresses_user_id (user_id),
    INDEX idx_addresses_type (type),
    INDEX idx_addresses_is_default (is_default),
    INDEX idx_addresses_is_active (is_active),
    INDEX idx_addresses_city (city),
    INDEX idx_addresses_district (district),
    INDEX idx_addresses_deleted_at (deleted_at)
);

-- Add constraints
ALTER TABLE addresses ADD CONSTRAINT chk_address_type 
CHECK (type IN ('home', 'office', 'billing', 'shipping', 'other'));

ALTER TABLE addresses ADD CONSTRAINT chk_latitude 
CHECK (latitude IS NULL OR (latitude >= -90 AND latitude <= 90));

ALTER TABLE addresses ADD CONSTRAINT chk_longitude 
CHECK (longitude IS NULL OR (longitude >= -180 AND longitude <= 180));

-- Create index for geographic queries
CREATE INDEX idx_addresses_coordinates ON addresses(latitude, longitude);

-- Create index for full-text search
CREATE FULLTEXT INDEX idx_addresses_fulltext ON addresses(full_name, address_line1, ward, district, city);

-- +migrate Down
DROP TABLE IF EXISTS addresses;
