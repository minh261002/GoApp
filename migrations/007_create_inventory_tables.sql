-- +migrate Up
CREATE TABLE IF NOT EXISTS inventory_movements (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    product_id INT UNSIGNED NOT NULL,
    variant_id INT UNSIGNED NULL,
    type VARCHAR(50) NOT NULL, -- inbound, outbound, adjustment, transfer, return
    status VARCHAR(50) DEFAULT 'pending', -- pending, approved, completed, cancelled
    quantity INT NOT NULL, -- Positive for inbound, negative for outbound
    unit_cost DECIMAL(10,2) DEFAULT 0.00,
    total_cost DECIMAL(10,2) DEFAULT 0.00,
    reference VARCHAR(255), -- PO number, SO number, etc.
    reference_type VARCHAR(50), -- purchase_order, sales_order, etc.
    notes TEXT,
    created_by INT UNSIGNED NOT NULL,
    approved_by INT UNSIGNED NULL,
    approved_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    FOREIGN KEY (variant_id) REFERENCES product_variants(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL,
    
    INDEX idx_inventory_movements_product_id (product_id),
    INDEX idx_inventory_movements_variant_id (variant_id),
    INDEX idx_inventory_movements_type (type),
    INDEX idx_inventory_movements_status (status),
    INDEX idx_inventory_movements_created_by (created_by),
    INDEX idx_inventory_movements_approved_by (approved_by),
    INDEX idx_inventory_movements_reference (reference),
    INDEX idx_inventory_movements_created_at (created_at),
    INDEX idx_inventory_movements_deleted_at (deleted_at)
);

CREATE TABLE IF NOT EXISTS stock_levels (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    product_id INT UNSIGNED NOT NULL,
    variant_id INT UNSIGNED NULL,
    available_quantity INT DEFAULT 0, -- Số lượng có sẵn
    reserved_quantity INT DEFAULT 0,  -- Số lượng đã đặt
    incoming_quantity INT DEFAULT 0,  -- Số lượng sắp về
    total_quantity INT DEFAULT 0,     -- Tổng số lượng
    min_stock_level INT DEFAULT 0,    -- Mức tồn kho tối thiểu
    max_stock_level INT DEFAULT 0,    -- Mức tồn kho tối đa
    reorder_point INT DEFAULT 0,      -- Điểm đặt hàng lại
    last_movement_at TIMESTAMP NULL,  -- Lần di chuyển cuối
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    FOREIGN KEY (variant_id) REFERENCES product_variants(id) ON DELETE CASCADE,
    
    UNIQUE KEY unique_stock_level (product_id, variant_id),
    INDEX idx_stock_levels_product_id (product_id),
    INDEX idx_stock_levels_variant_id (variant_id),
    INDEX idx_stock_levels_available_quantity (available_quantity),
    INDEX idx_stock_levels_min_stock_level (min_stock_level),
    INDEX idx_stock_levels_deleted_at (deleted_at)
);

CREATE TABLE IF NOT EXISTS inventory_adjustments (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    product_id INT UNSIGNED NOT NULL,
    variant_id INT UNSIGNED NULL,
    reason VARCHAR(255) NOT NULL, -- Lý do điều chỉnh
    quantity_before INT NOT NULL, -- Số lượng trước
    quantity_after INT NOT NULL,  -- Số lượng sau
    quantity_diff INT NOT NULL,   -- Chênh lệch
    notes TEXT,
    created_by INT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    FOREIGN KEY (variant_id) REFERENCES product_variants(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE,
    
    INDEX idx_inventory_adjustments_product_id (product_id),
    INDEX idx_inventory_adjustments_variant_id (variant_id),
    INDEX idx_inventory_adjustments_created_by (created_by),
    INDEX idx_inventory_adjustments_created_at (created_at),
    INDEX idx_inventory_adjustments_deleted_at (deleted_at)
);

-- +migrate Down
DROP TABLE IF EXISTS inventory_adjustments;
DROP TABLE IF EXISTS stock_levels;
DROP TABLE IF EXISTS inventory_movements;
