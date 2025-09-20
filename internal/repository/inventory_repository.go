package repository

import (
	"go_app/internal/model"
	"go_app/pkg/database"
	"time"

	"gorm.io/gorm"
)

// InventoryRepository defines methods for interacting with inventory data
type InventoryRepository interface {
	// Inventory Movements
	CreateMovement(movement *model.InventoryMovement) error
	GetMovementByID(id uint) (*model.InventoryMovement, error)
	GetMovements(page, limit int, filters map[string]interface{}) ([]model.InventoryMovement, int64, error)
	UpdateMovement(movement *model.InventoryMovement) error
	DeleteMovement(id uint) error
	ApproveMovement(id uint, approvedBy uint) error
	CompleteMovement(id uint) error
	GetMovementsByProduct(productID uint, variantID *uint) ([]model.InventoryMovement, error)
	GetMovementsByReference(reference string) ([]model.InventoryMovement, error)

	// Stock Levels
	CreateStockLevel(stockLevel *model.StockLevel) error
	GetStockLevelByID(id uint) (*model.StockLevel, error)
	GetStockLevelByProduct(productID uint, variantID *uint) (*model.StockLevel, error)
	GetAllStockLevels(page, limit int, filters map[string]interface{}) ([]model.StockLevel, int64, error)
	UpdateStockLevel(stockLevel *model.StockLevel) error
	DeleteStockLevel(id uint) error
	UpdateStockQuantity(productID uint, variantID *uint, quantity int) error
	ReserveStock(productID uint, variantID *uint, quantity int) error
	ReleaseStock(productID uint, variantID *uint, quantity int) error
	GetLowStockProducts(threshold int) ([]model.StockLevel, error)
	GetOutOfStockProducts() ([]model.StockLevel, error)

	// Inventory Adjustments
	CreateAdjustment(adjustment *model.InventoryAdjustment) error
	GetAdjustmentByID(id uint) (*model.InventoryAdjustment, error)
	GetAdjustments(page, limit int, filters map[string]interface{}) ([]model.InventoryAdjustment, int64, error)
	UpdateAdjustment(adjustment *model.InventoryAdjustment) error
	DeleteAdjustment(id uint) error
	GetAdjustmentsByProduct(productID uint, variantID *uint) ([]model.InventoryAdjustment, error)

	// Statistics
	GetInventoryStats() (*model.InventoryStatsResponse, error)
	GetLowStockAlerts() ([]model.LowStockAlert, error)
	GetStockValue() (float64, error)
	GetMovementStats(startDate, endDate time.Time) (map[string]interface{}, error)
}

// inventoryRepository implements InventoryRepository
type inventoryRepository struct {
	db *gorm.DB
}

// NewInventoryRepository creates a new InventoryRepository
func NewInventoryRepository() InventoryRepository {
	return &inventoryRepository{
		db: database.DB,
	}
}

// Inventory Movements

// CreateMovement creates a new inventory movement
func (r *inventoryRepository) CreateMovement(movement *model.InventoryMovement) error {
	return r.db.Create(movement).Error
}

// GetMovementByID retrieves an inventory movement by its ID
func (r *inventoryRepository) GetMovementByID(id uint) (*model.InventoryMovement, error) {
	var movement model.InventoryMovement
	if err := r.db.Preload("Product").Preload("Variant").Preload("CreatedByUser").Preload("ApprovedByUser").First(&movement, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &movement, nil
}

// GetMovements retrieves inventory movements with pagination and filters
func (r *inventoryRepository) GetMovements(page, limit int, filters map[string]interface{}) ([]model.InventoryMovement, int64, error) {
	var movements []model.InventoryMovement
	var total int64
	db := r.db.Model(&model.InventoryMovement{}).Preload("Product").Preload("Variant").Preload("CreatedByUser").Preload("ApprovedByUser")

	// Apply filters
	for key, value := range filters {
		switch key {
		case "product_id":
			db = db.Where("product_id = ?", value)
		case "variant_id":
			db = db.Where("variant_id = ?", value)
		case "type":
			db = db.Where("type = ?", value)
		case "status":
			db = db.Where("status = ?", value)
		case "created_by":
			db = db.Where("created_by = ?", value)
		case "reference":
			db = db.Where("reference LIKE ?", "%"+value.(string)+"%")
		case "date_from":
			db = db.Where("created_at >= ?", value)
		case "date_to":
			db = db.Where("created_at <= ?", value)
		}
	}

	// Count total records
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting
	db = db.Order("created_at DESC")

	if err := db.Find(&movements).Error; err != nil {
		return nil, 0, err
	}

	return movements, total, nil
}

// UpdateMovement updates an existing inventory movement
func (r *inventoryRepository) UpdateMovement(movement *model.InventoryMovement) error {
	return r.db.Save(movement).Error
}

// DeleteMovement soft deletes an inventory movement
func (r *inventoryRepository) DeleteMovement(id uint) error {
	return r.db.Delete(&model.InventoryMovement{}, id).Error
}

// ApproveMovement approves an inventory movement
func (r *inventoryRepository) ApproveMovement(id uint, approvedBy uint) error {
	now := time.Now()
	return r.db.Model(&model.InventoryMovement{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      model.MovementStatusApproved,
		"approved_by": approvedBy,
		"approved_at": &now,
	}).Error
}

// CompleteMovement completes an inventory movement
func (r *inventoryRepository) CompleteMovement(id uint) error {
	now := time.Now()
	return r.db.Model(&model.InventoryMovement{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       model.MovementStatusCompleted,
		"completed_at": &now,
	}).Error
}

// GetMovementsByProduct retrieves movements for a specific product/variant
func (r *inventoryRepository) GetMovementsByProduct(productID uint, variantID *uint) ([]model.InventoryMovement, error) {
	var movements []model.InventoryMovement
	db := r.db.Where("product_id = ?", productID)
	if variantID != nil {
		db = db.Where("variant_id = ?", *variantID)
	}
	err := db.Preload("Product").Preload("Variant").Order("created_at DESC").Find(&movements).Error
	return movements, err
}

// GetMovementsByReference retrieves movements by reference
func (r *inventoryRepository) GetMovementsByReference(reference string) ([]model.InventoryMovement, error) {
	var movements []model.InventoryMovement
	err := r.db.Where("reference = ?", reference).Preload("Product").Preload("Variant").Order("created_at DESC").Find(&movements).Error
	return movements, err
}

// Stock Levels

// CreateStockLevel creates a new stock level
func (r *inventoryRepository) CreateStockLevel(stockLevel *model.StockLevel) error {
	return r.db.Create(stockLevel).Error
}

// GetStockLevelByID retrieves a stock level by its ID
func (r *inventoryRepository) GetStockLevelByID(id uint) (*model.StockLevel, error) {
	var stockLevel model.StockLevel
	if err := r.db.Preload("Product").Preload("Variant").First(&stockLevel, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &stockLevel, nil
}

// GetStockLevelByProduct retrieves stock level for a specific product/variant
func (r *inventoryRepository) GetStockLevelByProduct(productID uint, variantID *uint) (*model.StockLevel, error) {
	var stockLevel model.StockLevel
	db := r.db.Where("product_id = ?", productID)
	if variantID != nil {
		db = db.Where("variant_id = ?", *variantID)
	} else {
		db = db.Where("variant_id IS NULL")
	}

	if err := db.Preload("Product").Preload("Variant").First(&stockLevel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &stockLevel, nil
}

// GetAllStockLevels retrieves all stock levels with pagination and filters
func (r *inventoryRepository) GetAllStockLevels(page, limit int, filters map[string]interface{}) ([]model.StockLevel, int64, error) {
	var stockLevels []model.StockLevel
	var total int64
	db := r.db.Model(&model.StockLevel{}).Preload("Product").Preload("Variant")

	// Apply filters
	for key, value := range filters {
		switch key {
		case "product_id":
			db = db.Where("product_id = ?", value)
		case "variant_id":
			db = db.Where("variant_id = ?", value)
		case "low_stock":
			if value.(bool) {
				db = db.Where("available_quantity <= min_stock_level")
			}
		case "out_of_stock":
			if value.(bool) {
				db = db.Where("available_quantity = 0")
			}
		}
	}

	// Count total records
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting
	db = db.Order("available_quantity ASC")

	if err := db.Find(&stockLevels).Error; err != nil {
		return nil, 0, err
	}

	return stockLevels, total, nil
}

// UpdateStockLevel updates an existing stock level
func (r *inventoryRepository) UpdateStockLevel(stockLevel *model.StockLevel) error {
	return r.db.Save(stockLevel).Error
}

// DeleteStockLevel soft deletes a stock level
func (r *inventoryRepository) DeleteStockLevel(id uint) error {
	return r.db.Delete(&model.StockLevel{}, id).Error
}

// UpdateStockQuantity updates stock quantity for a product/variant
func (r *inventoryRepository) UpdateStockQuantity(productID uint, variantID *uint, quantity int) error {
	now := time.Now()
	updates := map[string]interface{}{
		"available_quantity": quantity,
		"total_quantity":     quantity,
		"last_movement_at":   &now,
	}

	db := r.db.Model(&model.StockLevel{}).Where("product_id = ?", productID)
	if variantID != nil {
		db = db.Where("variant_id = ?", *variantID)
	} else {
		db = db.Where("variant_id IS NULL")
	}

	return db.Updates(updates).Error
}

// ReserveStock reserves stock for a product/variant
func (r *inventoryRepository) ReserveStock(productID uint, variantID *uint, quantity int) error {
	db := r.db.Model(&model.StockLevel{}).Where("product_id = ?", productID)
	if variantID != nil {
		db = db.Where("variant_id = ?", *variantID)
	} else {
		db = db.Where("variant_id IS NULL")
	}

	return db.UpdateColumns(map[string]interface{}{
		"available_quantity": gorm.Expr("available_quantity - ?", quantity),
		"reserved_quantity":  gorm.Expr("reserved_quantity + ?", quantity),
	}).Error
}

// ReleaseStock releases reserved stock for a product/variant
func (r *inventoryRepository) ReleaseStock(productID uint, variantID *uint, quantity int) error {
	db := r.db.Model(&model.StockLevel{}).Where("product_id = ?", productID)
	if variantID != nil {
		db = db.Where("variant_id = ?", *variantID)
	} else {
		db = db.Where("variant_id IS NULL")
	}

	return db.UpdateColumns(map[string]interface{}{
		"available_quantity": gorm.Expr("available_quantity + ?", quantity),
		"reserved_quantity":  gorm.Expr("reserved_quantity - ?", quantity),
	}).Error
}

// GetLowStockProducts retrieves products with low stock
func (r *inventoryRepository) GetLowStockProducts(threshold int) ([]model.StockLevel, error) {
	var stockLevels []model.StockLevel
	err := r.db.Where("available_quantity <= ?", threshold).Preload("Product").Preload("Variant").Find(&stockLevels).Error
	return stockLevels, err
}

// GetOutOfStockProducts retrieves out of stock products
func (r *inventoryRepository) GetOutOfStockProducts() ([]model.StockLevel, error) {
	var stockLevels []model.StockLevel
	err := r.db.Where("available_quantity = 0").Preload("Product").Preload("Variant").Find(&stockLevels).Error
	return stockLevels, err
}

// Inventory Adjustments

// CreateAdjustment creates a new inventory adjustment
func (r *inventoryRepository) CreateAdjustment(adjustment *model.InventoryAdjustment) error {
	return r.db.Create(adjustment).Error
}

// GetAdjustmentByID retrieves an inventory adjustment by its ID
func (r *inventoryRepository) GetAdjustmentByID(id uint) (*model.InventoryAdjustment, error) {
	var adjustment model.InventoryAdjustment
	if err := r.db.Preload("Product").Preload("Variant").Preload("CreatedByUser").First(&adjustment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &adjustment, nil
}

// GetAdjustments retrieves inventory adjustments with pagination and filters
func (r *inventoryRepository) GetAdjustments(page, limit int, filters map[string]interface{}) ([]model.InventoryAdjustment, int64, error) {
	var adjustments []model.InventoryAdjustment
	var total int64
	db := r.db.Model(&model.InventoryAdjustment{}).Preload("Product").Preload("Variant").Preload("CreatedByUser")

	// Apply filters
	for key, value := range filters {
		switch key {
		case "product_id":
			db = db.Where("product_id = ?", value)
		case "variant_id":
			db = db.Where("variant_id = ?", value)
		case "created_by":
			db = db.Where("created_by = ?", value)
		case "date_from":
			db = db.Where("created_at >= ?", value)
		case "date_to":
			db = db.Where("created_at <= ?", value)
		}
	}

	// Count total records
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting
	db = db.Order("created_at DESC")

	if err := db.Find(&adjustments).Error; err != nil {
		return nil, 0, err
	}

	return adjustments, total, nil
}

// UpdateAdjustment updates an existing inventory adjustment
func (r *inventoryRepository) UpdateAdjustment(adjustment *model.InventoryAdjustment) error {
	return r.db.Save(adjustment).Error
}

// DeleteAdjustment soft deletes an inventory adjustment
func (r *inventoryRepository) DeleteAdjustment(id uint) error {
	return r.db.Delete(&model.InventoryAdjustment{}, id).Error
}

// GetAdjustmentsByProduct retrieves adjustments for a specific product/variant
func (r *inventoryRepository) GetAdjustmentsByProduct(productID uint, variantID *uint) ([]model.InventoryAdjustment, error) {
	var adjustments []model.InventoryAdjustment
	db := r.db.Where("product_id = ?", productID)
	if variantID != nil {
		db = db.Where("variant_id = ?", *variantID)
	}
	err := db.Preload("Product").Preload("Variant").Order("created_at DESC").Find(&adjustments).Error
	return adjustments, err
}

// Statistics

// GetInventoryStats retrieves inventory statistics
func (r *inventoryRepository) GetInventoryStats() (*model.InventoryStatsResponse, error) {
	var stats model.InventoryStatsResponse
	var count int64

	// Total products
	r.db.Model(&model.Product{}).Count(&count)
	stats.TotalProducts = count

	// In stock products
	r.db.Model(&model.StockLevel{}).Where("available_quantity > 0").Count(&count)
	stats.InStockProducts = count

	// Out of stock products
	r.db.Model(&model.StockLevel{}).Where("available_quantity = 0").Count(&count)
	stats.OutOfStockProducts = count

	// Low stock products
	r.db.Model(&model.StockLevel{}).Where("available_quantity <= min_stock_level AND available_quantity > 0").Count(&count)
	stats.LowStockProducts = count

	// Total movements
	r.db.Model(&model.InventoryMovement{}).Count(&count)
	stats.TotalMovements = count

	// Pending movements
	r.db.Model(&model.InventoryMovement{}).Where("status = ?", model.MovementStatusPending).Count(&count)
	stats.PendingMovements = count

	// Completed movements
	r.db.Model(&model.InventoryMovement{}).Where("status = ?", model.MovementStatusCompleted).Count(&count)
	stats.CompletedMovements = count

	// Total value (simplified calculation)
	var totalValue float64
	r.db.Model(&model.StockLevel{}).Select("SUM(available_quantity * 0)").Scan(&totalValue) // This would need actual cost calculation
	stats.TotalValue = totalValue

	return &stats, nil
}

// GetLowStockAlerts retrieves low stock alerts
func (r *inventoryRepository) GetLowStockAlerts() ([]model.LowStockAlert, error) {
	var alerts []model.LowStockAlert

	err := r.db.Table("stock_levels sl").
		Select(`
			sl.product_id,
			p.name as product_name,
			sl.variant_id,
			pv.name as variant_name,
			sl.available_quantity,
			sl.min_stock_level,
			sl.reorder_point
		`).
		Joins("LEFT JOIN products p ON sl.product_id = p.id").
		Joins("LEFT JOIN product_variants pv ON sl.variant_id = pv.id").
		Where("sl.available_quantity <= sl.min_stock_level AND sl.available_quantity > 0").
		Find(&alerts).Error

	return alerts, err
}

// GetStockValue calculates total stock value
func (r *inventoryRepository) GetStockValue() (float64, error) {
	var totalValue float64
	err := r.db.Model(&model.StockLevel{}).Select("SUM(available_quantity * 0)").Scan(&totalValue).Error // This would need actual cost calculation
	return totalValue, err
}

// GetMovementStats retrieves movement statistics for a date range
func (r *inventoryRepository) GetMovementStats(startDate, endDate time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total movements in period
	var totalMovements int64
	r.db.Model(&model.InventoryMovement{}).Where("created_at BETWEEN ? AND ?", startDate, endDate).Count(&totalMovements)
	stats["total_movements"] = totalMovements

	// Inbound movements
	var inboundMovements int64
	r.db.Model(&model.InventoryMovement{}).Where("type = ? AND created_at BETWEEN ? AND ?", model.MovementTypeInbound, startDate, endDate).Count(&inboundMovements)
	stats["inbound_movements"] = inboundMovements

	// Outbound movements
	var outboundMovements int64
	r.db.Model(&model.InventoryMovement{}).Where("type = ? AND created_at BETWEEN ? AND ?", model.MovementTypeOutbound, startDate, endDate).Count(&outboundMovements)
	stats["outbound_movements"] = outboundMovements

	// Total inbound quantity
	var inboundQuantity int64
	r.db.Model(&model.InventoryMovement{}).Where("type = ? AND created_at BETWEEN ? AND ?", model.MovementTypeInbound, startDate, endDate).Select("SUM(quantity)").Scan(&inboundQuantity)
	stats["inbound_quantity"] = inboundQuantity

	// Total outbound quantity
	var outboundQuantity int64
	r.db.Model(&model.InventoryMovement{}).Where("type = ? AND created_at BETWEEN ? AND ?", model.MovementTypeOutbound, startDate, endDate).Select("SUM(quantity)").Scan(&outboundQuantity)
	stats["outbound_quantity"] = outboundQuantity

	return stats, nil
}
