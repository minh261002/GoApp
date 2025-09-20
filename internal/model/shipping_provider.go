package model

import (
	"time"

	"gorm.io/gorm"
)

// ShippingProvider represents a shipping provider
type ShippingProvider struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Code        string `json:"code" gorm:"size:50;not null;uniqueIndex"`
	DisplayName string `json:"display_name" gorm:"size:255;not null"`
	Description string `json:"description" gorm:"type:text"`
	Logo        string `json:"logo" gorm:"size:500"`
	Website     string `json:"website" gorm:"size:500"`
	Phone       string `json:"phone" gorm:"size:20"`
	Email       string `json:"email" gorm:"size:255"`

	// Configuration
	Config    string `json:"config" gorm:"type:json"` // Provider-specific config
	IsActive  bool   `json:"is_active" gorm:"default:true"`
	IsDefault bool   `json:"is_default" gorm:"default:false"`
	Priority  int    `json:"priority" gorm:"default:0"` // Higher number = higher priority

	// Features
	SupportsCOD       bool `json:"supports_cod" gorm:"default:false"`
	SupportsTracking  bool `json:"supports_tracking" gorm:"default:false"`
	SupportsInsurance bool `json:"supports_insurance" gorm:"default:false"`
	SupportsFragile   bool `json:"supports_fragile" gorm:"default:false"`

	// Limits
	MinWeight float64 `json:"min_weight" gorm:"type:decimal(8,3);default:0"` // in kg
	MaxWeight float64 `json:"max_weight" gorm:"type:decimal(8,3);default:0"` // in kg
	MinValue  float64 `json:"min_value" gorm:"type:decimal(10,2);default:0"` // in VND
	MaxValue  float64 `json:"max_value" gorm:"type:decimal(10,2);default:0"` // in VND

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// ShippingRate represents shipping rates for different zones
type ShippingRate struct {
	ID         uint              `json:"id" gorm:"primaryKey"`
	ProviderID uint              `json:"provider_id" gorm:"not null;index"`
	Provider   *ShippingProvider `json:"provider,omitempty" gorm:"foreignKey:ProviderID"`

	// Zone Information
	FromProvince string `json:"from_province" gorm:"size:100;not null"`
	FromDistrict string `json:"from_district" gorm:"size:100;not null"`
	ToProvince   string `json:"to_province" gorm:"size:100;not null"`
	ToDistrict   string `json:"to_district" gorm:"size:100;not null"`

	// Weight and Value Ranges
	MinWeight float64 `json:"min_weight" gorm:"type:decimal(8,3);not null"`
	MaxWeight float64 `json:"max_weight" gorm:"type:decimal(8,3);not null"`
	MinValue  float64 `json:"min_value" gorm:"type:decimal(10,2);not null"`
	MaxValue  float64 `json:"max_value" gorm:"type:decimal(10,2);not null"`

	// Pricing
	BaseFee   float64 `json:"base_fee" gorm:"type:decimal(10,2);not null"`       // Base shipping fee
	WeightFee float64 `json:"weight_fee" gorm:"type:decimal(10,2);default:0"`    // Fee per kg
	ValueFee  float64 `json:"value_fee" gorm:"type:decimal(10,2);default:0"`     // Fee per VND
	COD       float64 `json:"cod_fee" gorm:"type:decimal(10,2);default:0"`       // COD fee
	Insurance float64 `json:"insurance_fee" gorm:"type:decimal(10,2);default:0"` // Insurance fee
	Fragile   float64 `json:"fragile_fee" gorm:"type:decimal(10,2);default:0"`   // Fragile fee

	// Delivery Time
	MinDays int `json:"min_days" gorm:"default:1"`
	MaxDays int `json:"max_days" gorm:"default:3"`

	IsActive bool `json:"is_active" gorm:"default:true"`

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// ShippingOrder represents a shipping order with external provider
type ShippingOrder struct {
	ID         uint              `json:"id" gorm:"primaryKey"`
	OrderID    uint              `json:"order_id" gorm:"not null;index"`
	Order      *Order            `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	ProviderID uint              `json:"provider_id" gorm:"not null;index"`
	Provider   *ShippingProvider `json:"provider,omitempty" gorm:"foreignKey:ProviderID"`

	// External Provider Information
	ExternalID   string `json:"external_id" gorm:"size:100;index"` // Provider's order ID
	LabelID      string `json:"label_id" gorm:"size:100;index"`    // Provider's label ID
	TrackingCode string `json:"tracking_code" gorm:"size:100;index"`

	// Shipping Details
	FromName    string `json:"from_name" gorm:"size:255;not null"`
	FromAddress string `json:"from_address" gorm:"type:text;not null"`
	FromPhone   string `json:"from_phone" gorm:"size:20;not null"`
	FromEmail   string `json:"from_email" gorm:"size:255"`

	ToName    string `json:"to_name" gorm:"size:255;not null"`
	ToAddress string `json:"to_address" gorm:"type:text;not null"`
	ToPhone   string `json:"to_phone" gorm:"size:20;not null"`
	ToEmail   string `json:"to_email" gorm:"size:255"`

	// Package Information
	Weight    float64 `json:"weight" gorm:"type:decimal(8,3);not null"`      // in kg
	Value     float64 `json:"value" gorm:"type:decimal(10,2);not null"`      // in VND
	COD       float64 `json:"cod" gorm:"type:decimal(10,2);default:0"`       // COD amount
	Insurance float64 `json:"insurance" gorm:"type:decimal(10,2);default:0"` // Insurance amount

	// Fees
	ShippingFee  float64 `json:"shipping_fee" gorm:"type:decimal(10,2);not null"`
	CODFee       float64 `json:"cod_fee" gorm:"type:decimal(10,2);default:0"`
	InsuranceFee float64 `json:"insurance_fee" gorm:"type:decimal(10,2);default:0"`
	TotalFee     float64 `json:"total_fee" gorm:"type:decimal(10,2);not null"`

	// Status
	Status     string `json:"status" gorm:"size:50;not null;index"`
	StatusText string `json:"status_text" gorm:"size:255"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	ShippedAt   *time.Time `json:"shipped_at"`
	DeliveredAt *time.Time `json:"delivered_at"`
}

// ShippingTracking represents tracking information
type ShippingTracking struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	ShippingOrderID uint           `json:"shipping_order_id" gorm:"not null;index"`
	ShippingOrder   *ShippingOrder `json:"shipping_order,omitempty" gorm:"foreignKey:ShippingOrderID"`

	Status     string `json:"status" gorm:"size:50;not null"`
	StatusText string `json:"status_text" gorm:"size:255;not null"`
	Location   string `json:"location" gorm:"size:255"`
	Note       string `json:"note" gorm:"type:text"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// ShippingProvider Constants
const (
	ProviderCodeGHTK        = "ghtk"
	ProviderCodeGHN         = "ghn"
	ProviderCodeViettelPost = "viettel_post"
	ProviderCodeJT          = "jt"
	ProviderCodeBest        = "best"
)

// Shipping Status Constants
const (
	ShippingOrderStatusPending   = "pending"
	ShippingOrderStatusCreated   = "created"
	ShippingOrderStatusPickedUp  = "picked_up"
	ShippingOrderStatusInTransit = "in_transit"
	ShippingOrderStatusDelivered = "delivered"
	ShippingOrderStatusFailed    = "failed"
	ShippingOrderStatusCancelled = "cancelled"
	ShippingOrderStatusReturned  = "returned"
)

// ShippingOrderRequest represents request to create shipping order
type ShippingOrderRequest struct {
	OrderID    uint     `json:"order_id" validate:"required"`
	ProviderID uint     `json:"provider_id" validate:"required"`
	Weight     float64  `json:"weight" validate:"required,min=0.1"`
	Value      float64  `json:"value" validate:"required,min=0"`
	COD        float64  `json:"cod,omitempty"`
	Insurance  float64  `json:"insurance,omitempty"`
	Notes      string   `json:"notes,omitempty"`
	Tags       []string `json:"tags,omitempty"`
}

// ShippingOrderResponse represents shipping order response
type ShippingOrderResponse struct {
	ID           uint       `json:"id"`
	OrderID      uint       `json:"order_id"`
	ProviderID   uint       `json:"provider_id"`
	ProviderName string     `json:"provider_name"`
	ExternalID   string     `json:"external_id"`
	LabelID      string     `json:"label_id"`
	TrackingCode string     `json:"tracking_code"`
	Status       string     `json:"status"`
	StatusText   string     `json:"status_text"`
	ShippingFee  float64    `json:"shipping_fee"`
	TotalFee     float64    `json:"total_fee"`
	CreatedAt    time.Time  `json:"created_at"`
	ShippedAt    *time.Time `json:"shipped_at,omitempty"`
	DeliveredAt  *time.Time `json:"delivered_at,omitempty"`
}

// CalculateShippingRequest represents request to calculate shipping
type CalculateShippingRequest struct {
	FromProvince string  `json:"from_province" validate:"required"`
	FromDistrict string  `json:"from_district" validate:"required"`
	ToProvince   string  `json:"to_province" validate:"required"`
	ToDistrict   string  `json:"to_district" validate:"required"`
	Weight       float64 `json:"weight" validate:"required,min=0.1"`
	Value        float64 `json:"value" validate:"required,min=0"`
	ProviderID   *uint   `json:"provider_id,omitempty"`
	COD          float64 `json:"cod,omitempty"`
	Insurance    float64 `json:"insurance,omitempty"`
}

// CalculateShippingResponse represents shipping calculation response
type CalculateShippingResponse struct {
	ProviderID   uint    `json:"provider_id"`
	ProviderName string  `json:"provider_name"`
	ProviderCode string  `json:"provider_code"`
	ShippingFee  float64 `json:"shipping_fee"`
	COD          float64 `json:"cod_fee"`
	InsuranceFee float64 `json:"insurance_fee"`
	TotalFee     float64 `json:"total_fee"`
	MinDays      int     `json:"min_days"`
	MaxDays      int     `json:"max_days"`
	IsAvailable  bool    `json:"is_available"`
}

// WebhookData represents webhook data from shipping providers
type WebhookData struct {
	LabelID     string         `json:"label_id"`
	PartnerID   string         `json:"partner_id"`
	Status      string         `json:"status"`
	StatusText  string         `json:"status_text"`
	Created     string         `json:"created"`
	Updated     string         `json:"updated"`
	PickDate    string         `json:"pick_date"`
	DeliverDate string         `json:"deliver_date"`
	Products    []GHTKProduct  `json:"products"`
	Timeline    []GHTKTimeline `json:"timeline"`
}

type GHTKProduct struct {
	Name        string  `json:"name"`
	Weight      int     `json:"weight"`
	Quantity    int     `json:"quantity"`
	ProductCode string  `json:"product_code"`
	Price       float64 `json:"price"`
}

type GHTKTimeline struct {
	Status     string `json:"status"`
	StatusText string `json:"status_text"`
	Time       string `json:"time"`
	Location   string `json:"location"`
	Note       string `json:"note"`
}

// ShippingStats represents shipping statistics
type ShippingStats struct {
	TotalOrders     int64   `json:"total_orders"`
	PendingOrders   int64   `json:"pending_orders"`
	ShippedOrders   int64   `json:"shipped_orders"`
	DeliveredOrders int64   `json:"delivered_orders"`
	FailedOrders    int64   `json:"failed_orders"`
	TotalRevenue    float64 `json:"total_revenue"`
	AverageFee      float64 `json:"average_fee"`
	SuccessRate     float64 `json:"success_rate"`
}
