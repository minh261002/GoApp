package model

import (
	"time"

	"gorm.io/gorm"
)

// OrderStatus defines the status of an order
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"    // Chờ xử lý
	OrderStatusConfirmed  OrderStatus = "confirmed"  // Đã xác nhận
	OrderStatusProcessing OrderStatus = "processing" // Đang xử lý
	OrderStatusShipped    OrderStatus = "shipped"    // Đã giao hàng
	OrderStatusDelivered  OrderStatus = "delivered"  // Đã giao thành công
	OrderStatusCancelled  OrderStatus = "cancelled"  // Đã hủy
	OrderStatusReturned   OrderStatus = "returned"   // Đã trả hàng
	OrderStatusRefunded   OrderStatus = "refunded"   // Đã hoàn tiền
)

// PaymentStatus defines the payment status
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"   // Chờ thanh toán
	PaymentStatusPaid      PaymentStatus = "paid"      // Đã thanh toán
	PaymentStatusFailed    PaymentStatus = "failed"    // Thanh toán thất bại
	PaymentStatusRefunded  PaymentStatus = "refunded"  // Đã hoàn tiền
	PaymentStatusCancelled PaymentStatus = "cancelled" // Hủy thanh toán
)

// PaymentMethod defines the payment method
type PaymentMethod string

const (
	PaymentMethodCOD    PaymentMethod = "cod"    // Cash on Delivery
	PaymentMethodVietQR PaymentMethod = "vietqr" // VietQR (PayOS)
)

// ShippingStatus defines the shipping status
type ShippingStatus string

const (
	ShippingStatusPending   ShippingStatus = "pending"    // Chờ giao hàng
	ShippingStatusPickedUp  ShippingStatus = "picked_up"  // Đã lấy hàng
	ShippingStatusInTransit ShippingStatus = "in_transit" // Đang vận chuyển
	ShippingStatusDelivered ShippingStatus = "delivered"  // Đã giao hàng
	ShippingStatusFailed    ShippingStatus = "failed"     // Giao hàng thất bại
	ShippingStatusReturned  ShippingStatus = "returned"   // Trả hàng
)

// Order represents an order in the system
type Order struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	OrderNumber    string         `json:"order_number" gorm:"uniqueIndex;size:50;not null"` // Mã đơn hàng
	UserID         uint           `json:"user_id" gorm:"not null;index"`
	User           *User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Status         OrderStatus    `json:"status" gorm:"size:20;default:pending;index"`
	PaymentStatus  PaymentStatus  `json:"payment_status" gorm:"size:20;default:pending;index"`
	ShippingStatus ShippingStatus `json:"shipping_status" gorm:"size:20;default:pending;index"`

	// Customer Information
	CustomerName  string `json:"customer_name" gorm:"size:255;not null"`
	CustomerEmail string `json:"customer_email" gorm:"size:255;not null"`
	CustomerPhone string `json:"customer_phone" gorm:"size:20;not null"`

	// Address Information
	ShippingAddress string `json:"shipping_address" gorm:"type:text;not null"`
	BillingAddress  string `json:"billing_address" gorm:"type:text"`

	// Address References (optional - for linking to saved addresses)
	ShippingAddressID  *uint    `json:"shipping_address_id" gorm:"index"`
	BillingAddressID   *uint    `json:"billing_address_id" gorm:"index"`
	ShippingAddressRef *Address `json:"shipping_address_ref,omitempty" gorm:"foreignKey:ShippingAddressID"`
	BillingAddressRef  *Address `json:"billing_address_ref,omitempty" gorm:"foreignKey:BillingAddressID"`

	// Pricing Information
	SubTotal       float64 `json:"sub_total" gorm:"type:decimal(10,2);not null"`        // Tổng tiền hàng
	TaxAmount      float64 `json:"tax_amount" gorm:"type:decimal(10,2);default:0"`      // Thuế
	ShippingCost   float64 `json:"shipping_cost" gorm:"type:decimal(10,2);default:0"`   // Phí vận chuyển
	DiscountAmount float64 `json:"discount_amount" gorm:"type:decimal(10,2);default:0"` // Giảm giá
	TotalAmount    float64 `json:"total_amount" gorm:"type:decimal(10,2);not null"`     // Tổng cộng

	// Payment Information
	PaymentMethod    PaymentMethod `json:"payment_method" gorm:"size:20;not null"`
	PaymentReference string        `json:"payment_reference" gorm:"size:100"` // Mã tham chiếu thanh toán
	PaidAt           *time.Time    `json:"paid_at"`                           // Thời gian thanh toán

	// Shipping Information
	ShippingMethod string     `json:"shipping_method" gorm:"size:100"` // Phương thức vận chuyển
	TrackingNumber string     `json:"tracking_number" gorm:"size:100"` // Mã vận đơn
	ShippedAt      *time.Time `json:"shipped_at"`                      // Thời gian giao hàng
	DeliveredAt    *time.Time `json:"delivered_at"`                    // Thời gian nhận hàng

	// Additional Information
	Notes      string `json:"notes" gorm:"type:text"`       // Ghi chú
	AdminNotes string `json:"admin_notes" gorm:"type:text"` // Ghi chú admin
	Tags       string `json:"tags" gorm:"size:500"`         // Tags phân loại

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relations
	OrderItems      []OrderItem       `json:"order_items,omitempty" gorm:"foreignKey:OrderID"`
	Payments        []Payment         `json:"payments,omitempty" gorm:"foreignKey:OrderID"`
	ShippingHistory []ShippingHistory `json:"shipping_history,omitempty" gorm:"foreignKey:OrderID"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID               uint            `json:"id" gorm:"primaryKey"`
	OrderID          uint            `json:"order_id" gorm:"not null;index"`
	Order            *Order          `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	ProductID        uint            `json:"product_id" gorm:"not null;index"`
	Product          *Product        `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	ProductVariantID *uint           `json:"product_variant_id" gorm:"index"`
	ProductVariant   *ProductVariant `json:"product_variant,omitempty" gorm:"foreignKey:ProductVariantID"`

	// Product Information (snapshot at time of order)
	ProductName  string `json:"product_name" gorm:"size:255;not null"`
	ProductSKU   string `json:"product_sku" gorm:"size:100;not null"`
	ProductImage string `json:"product_image" gorm:"size:500"`
	VariantName  string `json:"variant_name" gorm:"size:255"` // e.g., "Size: L, Color: Red"

	// Pricing Information
	UnitPrice  float64 `json:"unit_price" gorm:"type:decimal(10,2);not null"`  // Giá đơn vị
	Quantity   int     `json:"quantity" gorm:"not null"`                       // Số lượng
	TotalPrice float64 `json:"total_price" gorm:"type:decimal(10,2);not null"` // Tổng tiền

	// Additional Information
	Weight     float64 `json:"weight" gorm:"type:decimal(8,2);default:0"` // Trọng lượng (kg)
	Dimensions string  `json:"dimensions" gorm:"size:100"`                // Kích thước (LxWxH)
	Notes      string  `json:"notes" gorm:"type:text"`                    // Ghi chú

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// Cart represents a shopping cart
type Cart struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	UserID    uint   `json:"user_id" gorm:"not null;index"`
	User      *User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	SessionID string `json:"session_id" gorm:"size:100;index"` // For guest users

	// Cart Information
	ItemsCount     int     `json:"items_count" gorm:"default:0"`                        // Tổng số sản phẩm
	ItemsQuantity  int     `json:"items_quantity" gorm:"default:0"`                     // Tổng số lượng
	SubTotal       float64 `json:"sub_total" gorm:"type:decimal(10,2);default:0"`       // Tổng tiền hàng
	TaxAmount      float64 `json:"tax_amount" gorm:"type:decimal(10,2);default:0"`      // Thuế
	ShippingCost   float64 `json:"shipping_cost" gorm:"type:decimal(10,2);default:0"`   // Phí vận chuyển
	DiscountAmount float64 `json:"discount_amount" gorm:"type:decimal(10,2);default:0"` // Giảm giá
	TotalAmount    float64 `json:"total_amount" gorm:"type:decimal(10,2);default:0"`    // Tổng cộng

	// Additional Information
	ShippingAddress string `json:"shipping_address" gorm:"type:text"` // Địa chỉ giao hàng
	BillingAddress  string `json:"billing_address" gorm:"type:text"`  // Địa chỉ thanh toán
	Notes           string `json:"notes" gorm:"type:text"`            // Ghi chú

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relations
	CartItems []CartItem `json:"cart_items,omitempty" gorm:"foreignKey:CartID"`
}

// CartItem represents an item in a cart
type CartItem struct {
	ID               uint            `json:"id" gorm:"primaryKey"`
	CartID           uint            `json:"cart_id" gorm:"not null;index"`
	Cart             *Cart           `json:"cart,omitempty" gorm:"foreignKey:CartID"`
	ProductID        uint            `json:"product_id" gorm:"not null;index"`
	Product          *Product        `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	ProductVariantID *uint           `json:"product_variant_id" gorm:"index"`
	ProductVariant   *ProductVariant `json:"product_variant,omitempty" gorm:"foreignKey:ProductVariantID"`

	Quantity   int     `json:"quantity" gorm:"not null"`                       // Số lượng
	UnitPrice  float64 `json:"unit_price" gorm:"type:decimal(10,2);not null"`  // Giá đơn vị
	TotalPrice float64 `json:"total_price" gorm:"type:decimal(10,2);not null"` // Tổng tiền

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// Payment represents a payment for an order
type Payment struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	OrderID uint   `json:"order_id" gorm:"not null;index"`
	Order   *Order `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	UserID  uint   `json:"user_id" gorm:"not null;index"`
	User    *User  `json:"user,omitempty" gorm:"foreignKey:UserID"`

	// Payment Information
	PaymentMethod PaymentMethod `json:"payment_method" gorm:"size:20;not null"`
	Status        PaymentStatus `json:"status" gorm:"size:20;default:pending;index"`
	Amount        float64       `json:"amount" gorm:"type:decimal(10,2);not null"` // Số tiền
	Currency      string        `json:"currency" gorm:"size:3;default:VND"`        // Đơn vị tiền tệ

	// Transaction Information
	TransactionID   string `json:"transaction_id" gorm:"size:100;uniqueIndex"` // Mã giao dịch
	ReferenceID     string `json:"reference_id" gorm:"size:100"`               // Mã tham chiếu
	GatewayResponse string `json:"gateway_response" gorm:"type:text"`          // Phản hồi từ gateway

	// Additional Information
	Description string `json:"description" gorm:"size:500"` // Mô tả
	Notes       string `json:"notes" gorm:"type:text"`      // Ghi chú

	// Timestamps
	ProcessedAt *time.Time     `json:"processed_at"` // Thời gian xử lý
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// ShippingHistory represents shipping status history
type ShippingHistory struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	OrderID uint   `json:"order_id" gorm:"not null;index"`
	Order   *Order `json:"order,omitempty" gorm:"foreignKey:OrderID"`

	// Status Information
	Status      ShippingStatus `json:"status" gorm:"size:20;not null;index"`
	Description string         `json:"description" gorm:"size:500"` // Mô tả trạng thái
	Location    string         `json:"location" gorm:"size:255"`    // Vị trí hiện tại
	Notes       string         `json:"notes" gorm:"type:text"`      // Ghi chú

	// Additional Information
	UpdatedBy     uint  `json:"updated_by" gorm:"index"` // Người cập nhật
	UpdatedByUser *User `json:"updated_by_user,omitempty" gorm:"foreignKey:UpdatedBy"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// Payment Link Response
type PaymentLinkResponse struct {
	PaymentURL    string        `json:"payment_url"`
	QRCode        string        `json:"qr_code,omitempty"`
	OrderCode     int           `json:"order_code,omitempty"`
	Amount        float64       `json:"amount"`
	AccountNumber string        `json:"account_number,omitempty"`
	AccountName   string        `json:"account_name,omitempty"`
	ExpiresAt     time.Time     `json:"expires_at"`
	PaymentMethod PaymentMethod `json:"payment_method"`
}

// Payment Info Response
type PaymentInfoResponse struct {
	OrderCode     int           `json:"order_code"`
	Amount        float64       `json:"amount"`
	Status        PaymentStatus `json:"status"`
	TransactionID string        `json:"transaction_id"`
	Reference     string        `json:"reference"`
	AccountNumber string        `json:"account_number,omitempty"`
	Description   string        `json:"description"`
	PaymentMethod PaymentMethod `json:"payment_method"`
}

// Payment Webhook Response
type PaymentWebhookResponse struct {
	OrderCode     int           `json:"order_code"`
	Amount        float64       `json:"amount"`
	Status        PaymentStatus `json:"status"`
	TransactionID string        `json:"transaction_id"`
	Reference     string        `json:"reference"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	RawData       []byte        `json:"raw_data,omitempty"`
}

// Payment Method Info
type PaymentMethodInfo struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Description string  `json:"description"`
	IsActive    bool    `json:"is_active"`
	IsOnline    bool    `json:"is_online"`
	FeeType     string  `json:"fee_type"` // fixed, percentage, none
	FeeValue    float64 `json:"fee_value"`
	MinAmount   float64 `json:"min_amount"`
	MaxAmount   float64 `json:"max_amount"`
	Currency    string  `json:"currency"`
}

// Request/Response structs

// OrderCreateRequest represents the request body for creating an order
type OrderCreateRequest struct {
	// User Information (optional - for admin creating orders for other users)
	UserID *uint `json:"user_id"` // If not provided, will use the authenticated user's ID

	// Customer Information
	CustomerName  string `json:"customer_name" binding:"required,min=2,max=255"`
	CustomerEmail string `json:"customer_email" binding:"required,email"`
	CustomerPhone string `json:"customer_phone" binding:"required,min=10,max=20"`

	// Address Information
	ShippingAddress string `json:"shipping_address" binding:"required,min=10"`
	BillingAddress  string `json:"billing_address"`

	// Address References (optional - for linking to saved addresses)
	ShippingAddressID *uint `json:"shipping_address_id"` // ID of saved address
	BillingAddressID  *uint `json:"billing_address_id"`  // ID of saved address

	// Payment Information
	PaymentMethod PaymentMethod `json:"payment_method" binding:"required,oneof=cash bank card wallet cod"`

	// Shipping Information
	ShippingMethod string `json:"shipping_method" binding:"required,min=2,max=100"`

	// Additional Information
	Notes string `json:"notes"`

	// Cart Items (if creating from cart)
	CartID *uint `json:"cart_id"`
}

// OrderUpdateRequest represents the request body for updating an order
type OrderUpdateRequest struct {
	Status         *OrderStatus    `json:"status" binding:"omitempty,oneof=pending confirmed processing shipped delivered cancelled returned refunded"`
	PaymentStatus  *PaymentStatus  `json:"payment_status" binding:"omitempty,oneof=pending paid failed refunded cancelled"`
	ShippingStatus *ShippingStatus `json:"shipping_status" binding:"omitempty,oneof=pending picked_up in_transit delivered failed returned"`

	// Customer Information
	CustomerName  string `json:"customer_name" binding:"omitempty,min=2,max=255"`
	CustomerEmail string `json:"customer_email" binding:"omitempty,email"`
	CustomerPhone string `json:"customer_phone" binding:"omitempty,min=10,max=20"`

	// Address Information
	ShippingAddress string `json:"shipping_address" binding:"omitempty,min=10"`
	BillingAddress  string `json:"billing_address"`

	// Shipping Information
	TrackingNumber string `json:"tracking_number" binding:"omitempty,max=100"`
	ShippingMethod string `json:"shipping_method" binding:"omitempty,min=2,max=100"`

	// Additional Information
	Notes      string `json:"notes"`
	AdminNotes string `json:"admin_notes"`
	Tags       string `json:"tags" binding:"omitempty,max=500"`
}

// OrderItemCreateRequest represents the request body for creating an order item
type OrderItemCreateRequest struct {
	ProductID        uint    `json:"product_id" binding:"required"`
	ProductVariantID *uint   `json:"product_variant_id"`
	Quantity         int     `json:"quantity" binding:"required,min=1"`
	UnitPrice        float64 `json:"unit_price" binding:"required,min=0"`
	Notes            string  `json:"notes"`
}

// CartCreateRequest represents the request body for creating a cart
type CartCreateRequest struct {
	SessionID       string `json:"session_id"` // For guest users
	ShippingAddress string `json:"shipping_address"`
	BillingAddress  string `json:"billing_address"`
	Notes           string `json:"notes"`
}

// CartItemCreateRequest represents the request body for adding item to cart
type CartItemCreateRequest struct {
	ProductID        uint  `json:"product_id" binding:"required"`
	ProductVariantID *uint `json:"product_variant_id"`
	Quantity         int   `json:"quantity" binding:"required,min=1"`
}

// PaymentCreateRequest represents the request body for creating a payment
type PaymentCreateRequest struct {
	OrderID       uint          `json:"order_id" binding:"required"`
	PaymentMethod PaymentMethod `json:"payment_method" binding:"required,oneof=cash bank card wallet cod"`
	Amount        float64       `json:"amount" binding:"required,min=0"`
	Currency      string        `json:"currency" binding:"omitempty,len=3"`
	Description   string        `json:"description"`
	Notes         string        `json:"notes"`
}

// OrderResponse represents the response body for an order
type OrderResponse struct {
	ID             uint           `json:"id"`
	OrderNumber    string         `json:"order_number"`
	UserID         uint           `json:"user_id"`
	UserName       string         `json:"user_name,omitempty"`
	Status         OrderStatus    `json:"status"`
	PaymentStatus  PaymentStatus  `json:"payment_status"`
	ShippingStatus ShippingStatus `json:"shipping_status"`

	// Customer Information
	CustomerName  string `json:"customer_name"`
	CustomerEmail string `json:"customer_email"`
	CustomerPhone string `json:"customer_phone"`

	// Address Information
	ShippingAddress string `json:"shipping_address"`
	BillingAddress  string `json:"billing_address"`

	// Pricing Information
	SubTotal       float64 `json:"sub_total"`
	TaxAmount      float64 `json:"tax_amount"`
	ShippingCost   float64 `json:"shipping_cost"`
	DiscountAmount float64 `json:"discount_amount"`
	TotalAmount    float64 `json:"total_amount"`

	// Payment Information
	PaymentMethod    PaymentMethod `json:"payment_method"`
	PaymentReference string        `json:"payment_reference"`
	PaidAt           *time.Time    `json:"paid_at"`

	// Shipping Information
	ShippingMethod string     `json:"shipping_method"`
	TrackingNumber string     `json:"tracking_number"`
	ShippedAt      *time.Time `json:"shipped_at"`
	DeliveredAt    *time.Time `json:"delivered_at"`

	// Additional Information
	Notes      string `json:"notes"`
	AdminNotes string `json:"admin_notes"`
	Tags       string `json:"tags"`

	// Relations
	OrderItems      []OrderItemResponse       `json:"order_items,omitempty"`
	Payments        []PaymentResponse         `json:"payments,omitempty"`
	ShippingHistory []ShippingHistoryResponse `json:"shipping_history,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OrderItemResponse represents the response body for an order item
type OrderItemResponse struct {
	ID               uint      `json:"id"`
	OrderID          uint      `json:"order_id"`
	ProductID        uint      `json:"product_id"`
	ProductName      string    `json:"product_name"`
	ProductSKU       string    `json:"product_sku"`
	ProductImage     string    `json:"product_image"`
	ProductVariantID *uint     `json:"product_variant_id"`
	VariantName      string    `json:"variant_name"`
	UnitPrice        float64   `json:"unit_price"`
	Quantity         int       `json:"quantity"`
	TotalPrice       float64   `json:"total_price"`
	Weight           float64   `json:"weight"`
	Dimensions       string    `json:"dimensions"`
	Notes            string    `json:"notes"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// CartResponse represents the response body for a cart
type CartResponse struct {
	ID              uint               `json:"id"`
	UserID          uint               `json:"user_id"`
	SessionID       string             `json:"session_id"`
	ItemsCount      int                `json:"items_count"`
	ItemsQuantity   int                `json:"items_quantity"`
	SubTotal        float64            `json:"sub_total"`
	TaxAmount       float64            `json:"tax_amount"`
	ShippingCost    float64            `json:"shipping_cost"`
	DiscountAmount  float64            `json:"discount_amount"`
	TotalAmount     float64            `json:"total_amount"`
	ShippingAddress string             `json:"shipping_address"`
	BillingAddress  string             `json:"billing_address"`
	Notes           string             `json:"notes"`
	CartItems       []CartItemResponse `json:"cart_items,omitempty"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

// CartItemResponse represents the response body for a cart item
type CartItemResponse struct {
	ID               uint      `json:"id"`
	CartID           uint      `json:"cart_id"`
	ProductID        uint      `json:"product_id"`
	ProductName      string    `json:"product_name,omitempty"`
	ProductSKU       string    `json:"product_sku,omitempty"`
	ProductImage     string    `json:"product_image,omitempty"`
	ProductVariantID *uint     `json:"product_variant_id"`
	VariantName      string    `json:"variant_name,omitempty"`
	Quantity         int       `json:"quantity"`
	UnitPrice        float64   `json:"unit_price"`
	TotalPrice       float64   `json:"total_price"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// PaymentResponse represents the response body for a payment
type PaymentResponse struct {
	ID            uint          `json:"id"`
	OrderID       uint          `json:"order_id"`
	UserID        uint          `json:"user_id"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	Status        PaymentStatus `json:"status"`
	Amount        float64       `json:"amount"`
	Currency      string        `json:"currency"`
	TransactionID string        `json:"transaction_id"`
	ReferenceID   string        `json:"reference_id"`
	Description   string        `json:"description"`
	Notes         string        `json:"notes"`
	ProcessedAt   *time.Time    `json:"processed_at"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// ShippingHistoryResponse represents the response body for shipping history
type ShippingHistoryResponse struct {
	ID            uint           `json:"id"`
	OrderID       uint           `json:"order_id"`
	Status        ShippingStatus `json:"status"`
	Description   string         `json:"description"`
	Location      string         `json:"location"`
	Notes         string         `json:"notes"`
	UpdatedBy     uint           `json:"updated_by"`
	UpdatedByName string         `json:"updated_by_name,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}

// Cart Update Request
type CartUpdateRequest struct {
	ShippingAddress *string `json:"shipping_address,omitempty"`
	BillingAddress  *string `json:"billing_address,omitempty"`
	Notes           *string `json:"notes,omitempty"`
}

// OrderStatsResponse represents order statistics
type OrderStatsResponse struct {
	TotalOrders       int64   `json:"total_orders"`
	PendingOrders     int64   `json:"pending_orders"`
	ConfirmedOrders   int64   `json:"confirmed_orders"`
	ShippedOrders     int64   `json:"shipped_orders"`
	DeliveredOrders   int64   `json:"delivered_orders"`
	CancelledOrders   int64   `json:"cancelled_orders"`
	TotalRevenue      float64 `json:"total_revenue"`
	AverageOrderValue float64 `json:"average_order_value"`
	ConversionRate    float64 `json:"conversion_rate"`
}

// Helper methods

// IsCompleted checks if order is completed
func (o *Order) IsCompleted() bool {
	return o.Status == OrderStatusDelivered
}

// IsCancelled checks if order is cancelled
func (o *Order) IsCancelled() bool {
	return o.Status == OrderStatusCancelled || o.Status == OrderStatusReturned || o.Status == OrderStatusRefunded
}

// CanBeCancelled checks if order can be cancelled
func (o *Order) CanBeCancelled() bool {
	return o.Status == OrderStatusPending || o.Status == OrderStatusConfirmed
}

// CanBeShipped checks if order can be shipped
func (o *Order) CanBeShipped() bool {
	return o.Status == OrderStatusConfirmed || o.Status == OrderStatusProcessing
}

// CanBeDelivered checks if order can be delivered
func (o *Order) CanBeDelivered() bool {
	return o.Status == OrderStatusShipped
}

// IsPaid checks if order is paid
func (o *Order) IsPaid() bool {
	return o.PaymentStatus == PaymentStatusPaid
}

// GetStatusDisplayName returns display name for order status
func (o *Order) GetStatusDisplayName() string {
	statusMap := map[OrderStatus]string{
		OrderStatusPending:    "Chờ xử lý",
		OrderStatusConfirmed:  "Đã xác nhận",
		OrderStatusProcessing: "Đang xử lý",
		OrderStatusShipped:    "Đã giao hàng",
		OrderStatusDelivered:  "Đã giao thành công",
		OrderStatusCancelled:  "Đã hủy",
		OrderStatusReturned:   "Đã trả hàng",
		OrderStatusRefunded:   "Đã hoàn tiền",
	}
	return statusMap[o.Status]
}

// GetPaymentStatusDisplayName returns display name for payment status
func (o *Order) GetPaymentStatusDisplayName() string {
	statusMap := map[PaymentStatus]string{
		PaymentStatusPending:   "Chờ thanh toán",
		PaymentStatusPaid:      "Đã thanh toán",
		PaymentStatusFailed:    "Thanh toán thất bại",
		PaymentStatusRefunded:  "Đã hoàn tiền",
		PaymentStatusCancelled: "Hủy thanh toán",
	}
	return statusMap[o.PaymentStatus]
}

// GetShippingStatusDisplayName returns display name for shipping status
func (o *Order) GetShippingStatusDisplayName() string {
	statusMap := map[ShippingStatus]string{
		ShippingStatusPending:   "Chờ giao hàng",
		ShippingStatusPickedUp:  "Đã lấy hàng",
		ShippingStatusInTransit: "Đang vận chuyển",
		ShippingStatusDelivered: "Đã giao hàng",
		ShippingStatusFailed:    "Giao hàng thất bại",
		ShippingStatusReturned:  "Trả hàng",
	}
	return statusMap[o.ShippingStatus]
}

// CalculateTotal calculates total amount for order
func (o *Order) CalculateTotal() {
	o.TotalAmount = o.SubTotal + o.TaxAmount + o.ShippingCost - o.DiscountAmount
}

// CalculateTotal calculates total price for order item
func (oi *OrderItem) CalculateTotal() {
	oi.TotalPrice = oi.UnitPrice * float64(oi.Quantity)
}

// CalculateTotal calculates total amount for cart
func (c *Cart) CalculateTotal() {
	c.TotalAmount = c.SubTotal + c.TaxAmount + c.ShippingCost - c.DiscountAmount
}

// CalculateTotal calculates total price for cart item
func (ci *CartItem) CalculateTotal() {
	ci.TotalPrice = ci.UnitPrice * float64(ci.Quantity)
}
