package model

import (
	"time"

	"gorm.io/gorm"
)

// InventoryMovementType defines the type of inventory movement
type InventoryMovementType string

const (
	MovementTypeInbound    InventoryMovementType = "inbound"    // Nhập hàng
	MovementTypeOutbound   InventoryMovementType = "outbound"   // Xuất hàng
	MovementTypeAdjustment InventoryMovementType = "adjustment" // Điều chỉnh
	MovementTypeTransfer   InventoryMovementType = "transfer"   // Chuyển kho
	MovementTypeReturn     InventoryMovementType = "return"     // Trả hàng
)

// InventoryMovementStatus defines the status of inventory movement
type InventoryMovementStatus string

const (
	MovementStatusPending   InventoryMovementStatus = "pending"   // Chờ xử lý
	MovementStatusApproved  InventoryMovementStatus = "approved"  // Đã duyệt
	MovementStatusCompleted InventoryMovementStatus = "completed" // Hoàn thành
	MovementStatusCancelled InventoryMovementStatus = "cancelled" // Đã hủy
)

// InventoryMovement represents an inventory movement record
type InventoryMovement struct {
	ID             uint                    `json:"id" gorm:"primaryKey"`
	ProductID      uint                    `json:"product_id" gorm:"not null;index"`
	Product        *Product                `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	VariantID      *uint                   `json:"variant_id" gorm:"index"`
	Variant        *ProductVariant         `json:"variant,omitempty" gorm:"foreignKey:VariantID"`
	Type           InventoryMovementType   `json:"type" gorm:"type:varchar(50);not null"`
	Status         InventoryMovementStatus `json:"status" gorm:"type:varchar(50);default:'pending'"`
	Quantity       int                     `json:"quantity" gorm:"not null"` // Số lượng (dương cho inbound, âm cho outbound)
	UnitCost       float64                 `json:"unit_cost" gorm:"type:decimal(10,2);default:0.00"`
	TotalCost      float64                 `json:"total_cost" gorm:"type:decimal(10,2);default:0.00"`
	Reference      string                  `json:"reference" gorm:"type:varchar(255)"`     // Số tham chiếu (PO, SO, etc.)
	ReferenceType  string                  `json:"reference_type" gorm:"type:varchar(50)"` // purchase_order, sales_order, etc.
	Notes          string                  `json:"notes" gorm:"type:text"`
	CreatedBy      uint                    `json:"created_by" gorm:"not null;index"`
	CreatedByUser  *User                   `json:"created_by_user,omitempty" gorm:"foreignKey:CreatedBy"`
	ApprovedBy     *uint                   `json:"approved_by" gorm:"index"`
	ApprovedByUser *User                   `json:"approved_by_user,omitempty" gorm:"foreignKey:ApprovedBy"`
	ApprovedAt     *time.Time              `json:"approved_at"`
	CompletedAt    *time.Time              `json:"completed_at"`
	CreatedAt      time.Time               `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time               `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt          `json:"deleted_at" gorm:"index"`
}

// StockLevel represents current stock level for a product/variant
type StockLevel struct {
	ID                uint            `json:"id" gorm:"primaryKey"`
	ProductID         uint            `json:"product_id" gorm:"not null;index"`
	Product           *Product        `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	VariantID         *uint           `json:"variant_id" gorm:"index"`
	Variant           *ProductVariant `json:"variant,omitempty" gorm:"foreignKey:VariantID"`
	AvailableQuantity int             `json:"available_quantity" gorm:"default:0"` // Số lượng có sẵn
	ReservedQuantity  int             `json:"reserved_quantity" gorm:"default:0"`  // Số lượng đã đặt
	IncomingQuantity  int             `json:"incoming_quantity" gorm:"default:0"`  // Số lượng sắp về
	TotalQuantity     int             `json:"total_quantity" gorm:"default:0"`     // Tổng số lượng
	MinStockLevel     int             `json:"min_stock_level" gorm:"default:0"`    // Mức tồn kho tối thiểu
	MaxStockLevel     int             `json:"max_stock_level" gorm:"default:0"`    // Mức tồn kho tối đa
	ReorderPoint      int             `json:"reorder_point" gorm:"default:0"`      // Điểm đặt hàng lại
	LastMovementAt    *time.Time      `json:"last_movement_at"`                    // Lần di chuyển cuối
	CreatedAt         time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt         gorm.DeletedAt  `json:"deleted_at" gorm:"index"`
}

// InventoryAdjustment represents a stock adjustment
type InventoryAdjustment struct {
	ID             uint            `json:"id" gorm:"primaryKey"`
	ProductID      uint            `json:"product_id" gorm:"not null;index"`
	Product        *Product        `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	VariantID      *uint           `json:"variant_id" gorm:"index"`
	Variant        *ProductVariant `json:"variant,omitempty" gorm:"foreignKey:VariantID"`
	Reason         string          `json:"reason" gorm:"type:varchar(255);not null"` // Lý do điều chỉnh
	QuantityBefore int             `json:"quantity_before" gorm:"not null"`          // Số lượng trước
	QuantityAfter  int             `json:"quantity_after" gorm:"not null"`           // Số lượng sau
	QuantityDiff   int             `json:"quantity_diff" gorm:"not null"`            // Chênh lệch
	Notes          string          `json:"notes" gorm:"type:text"`
	CreatedBy      uint            `json:"created_by" gorm:"not null;index"`
	CreatedByUser  *User           `json:"created_by_user,omitempty" gorm:"foreignKey:CreatedBy"`
	CreatedAt      time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt  `json:"deleted_at" gorm:"index"`
}

// Request/Response structs

// InventoryMovementCreateRequest represents the request body for creating an inventory movement
type InventoryMovementCreateRequest struct {
	ProductID     uint                  `json:"product_id" binding:"required"`
	VariantID     *uint                 `json:"variant_id"`
	Type          InventoryMovementType `json:"type" binding:"required,oneof=inbound outbound adjustment transfer return"`
	Quantity      int                   `json:"quantity" binding:"required"`
	UnitCost      float64               `json:"unit_cost" binding:"gte=0"`
	Reference     string                `json:"reference"`
	ReferenceType string                `json:"reference_type"`
	Notes         string                `json:"notes"`
}

// InventoryMovementUpdateRequest represents the request body for updating an inventory movement
type InventoryMovementUpdateRequest struct {
	Status InventoryMovementStatus `json:"status" binding:"required,oneof=pending approved completed cancelled"`
	Notes  string                  `json:"notes"`
}

// InventoryAdjustmentCreateRequest represents the request body for creating an inventory adjustment
type InventoryAdjustmentCreateRequest struct {
	ProductID     uint   `json:"product_id" binding:"required"`
	VariantID     *uint  `json:"variant_id"`
	Reason        string `json:"reason" binding:"required,min=3,max=255"`
	QuantityAfter int    `json:"quantity_after" binding:"required"`
	Notes         string `json:"notes"`
}

// StockLevelUpdateRequest represents the request body for updating stock levels
type StockLevelUpdateRequest struct {
	MinStockLevel int `json:"min_stock_level" binding:"gte=0"`
	MaxStockLevel int `json:"max_stock_level" binding:"gte=0"`
	ReorderPoint  int `json:"reorder_point" binding:"gte=0"`
}

// InventoryMovementResponse represents the response body for an inventory movement
type InventoryMovementResponse struct {
	ID             uint                    `json:"id"`
	ProductID      uint                    `json:"product_id"`
	ProductName    string                  `json:"product_name,omitempty"`
	VariantID      *uint                   `json:"variant_id"`
	VariantName    string                  `json:"variant_name,omitempty"`
	Type           InventoryMovementType   `json:"type"`
	Status         InventoryMovementStatus `json:"status"`
	Quantity       int                     `json:"quantity"`
	UnitCost       float64                 `json:"unit_cost"`
	TotalCost      float64                 `json:"total_cost"`
	Reference      string                  `json:"reference"`
	ReferenceType  string                  `json:"reference_type"`
	Notes          string                  `json:"notes"`
	CreatedBy      uint                    `json:"created_by"`
	CreatedByName  string                  `json:"created_by_name,omitempty"`
	ApprovedBy     *uint                   `json:"approved_by"`
	ApprovedByName string                  `json:"approved_by_name,omitempty"`
	ApprovedAt     *time.Time              `json:"approved_at"`
	CompletedAt    *time.Time              `json:"completed_at"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
}

// StockLevelResponse represents the response body for stock level
type StockLevelResponse struct {
	ID                uint       `json:"id"`
	ProductID         uint       `json:"product_id"`
	ProductName       string     `json:"product_name,omitempty"`
	VariantID         *uint      `json:"variant_id"`
	VariantName       string     `json:"variant_name,omitempty"`
	AvailableQuantity int        `json:"available_quantity"`
	ReservedQuantity  int        `json:"reserved_quantity"`
	IncomingQuantity  int        `json:"incoming_quantity"`
	TotalQuantity     int        `json:"total_quantity"`
	MinStockLevel     int        `json:"min_stock_level"`
	MaxStockLevel     int        `json:"max_stock_level"`
	ReorderPoint      int        `json:"reorder_point"`
	LastMovementAt    *time.Time `json:"last_movement_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// InventoryAdjustmentResponse represents the response body for an inventory adjustment
type InventoryAdjustmentResponse struct {
	ID             uint      `json:"id"`
	ProductID      uint      `json:"product_id"`
	ProductName    string    `json:"product_name,omitempty"`
	VariantID      *uint     `json:"variant_id"`
	VariantName    string    `json:"variant_name,omitempty"`
	Reason         string    `json:"reason"`
	QuantityBefore int       `json:"quantity_before"`
	QuantityAfter  int       `json:"quantity_after"`
	QuantityDiff   int       `json:"quantity_diff"`
	Notes          string    `json:"notes"`
	CreatedBy      uint      `json:"created_by"`
	CreatedByName  string    `json:"created_by_name,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// InventoryStatsResponse represents inventory statistics
type InventoryStatsResponse struct {
	TotalProducts      int64   `json:"total_products"`
	InStockProducts    int64   `json:"in_stock_products"`
	OutOfStockProducts int64   `json:"out_of_stock_products"`
	LowStockProducts   int64   `json:"low_stock_products"`
	TotalValue         float64 `json:"total_value"`
	TotalMovements     int64   `json:"total_movements"`
	PendingMovements   int64   `json:"pending_movements"`
	CompletedMovements int64   `json:"completed_movements"`
}

// LowStockAlert represents a low stock alert
type LowStockAlert struct {
	ProductID         uint   `json:"product_id"`
	ProductName       string `json:"product_name"`
	VariantID         *uint  `json:"variant_id"`
	VariantName       string `json:"variant_name,omitempty"`
	CurrentQuantity   int    `json:"current_quantity"`
	MinStockLevel     int    `json:"min_stock_level"`
	ReorderPoint      int    `json:"reorder_point"`
	DaysUntilStockout int    `json:"days_until_stockout,omitempty"`
}
