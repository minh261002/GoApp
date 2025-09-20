package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// CouponType defines the type of coupon
type CouponType string

const (
	CouponTypePercentage   CouponType = "percentage"    // Giảm theo phần trăm
	CouponTypeFixed        CouponType = "fixed"         // Giảm số tiền cố định
	CouponTypeFreeShipping CouponType = "free_shipping" // Miễn phí vận chuyển
	CouponTypeBuyXGetY     CouponType = "buy_x_get_y"   // Mua X tặng Y
)

// CouponStatus defines the status of a coupon
type CouponStatus string

const (
	CouponStatusActive   CouponStatus = "active"   // Đang hoạt động
	CouponStatusInactive CouponStatus = "inactive" // Tạm dừng
	CouponStatusExpired  CouponStatus = "expired"  // Hết hạn
	CouponStatusUsed     CouponStatus = "used"     // Đã sử dụng hết
)

// PointType defines the type of point transaction
type PointType string

const (
	PointTypeEarn   PointType = "earn"   // Tích điểm
	PointTypeRedeem PointType = "redeem" // Đổi điểm
	PointTypeExpire PointType = "expire" // Hết hạn điểm
	PointTypeRefund PointType = "refund" // Hoàn điểm
	PointTypeAdjust PointType = "adjust" // Điều chỉnh điểm
)

// PointStatus defines the status of point transaction
type PointStatus string

const (
	PointStatusPending   PointStatus = "pending"   // Chờ xử lý
	PointStatusCompleted PointStatus = "completed" // Hoàn thành
	PointStatusCancelled PointStatus = "cancelled" // Đã hủy
	PointStatusExpired   PointStatus = "expired"   // Đã hết hạn
)

// Coupon represents a discount coupon
type Coupon struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Code        string       `json:"code" gorm:"size:50;uniqueIndex;not null"`
	Name        string       `json:"name" gorm:"size:255;not null"`
	Description string       `json:"description" gorm:"type:text"`
	Type        CouponType   `json:"type" gorm:"size:20;not null"`
	Status      CouponStatus `json:"status" gorm:"size:20;default:'active'"`

	// Discount Configuration
	DiscountValue     float64 `json:"discount_value" gorm:"type:decimal(10,2);not null"`       // Giá trị giảm giá
	MinOrderAmount    float64 `json:"min_order_amount" gorm:"type:decimal(10,2);default:0"`    // Đơn hàng tối thiểu
	MaxDiscountAmount float64 `json:"max_discount_amount" gorm:"type:decimal(10,2);default:0"` // Giảm giá tối đa

	// Usage Configuration
	UsageLimit   int `json:"usage_limit" gorm:"default:0"`    // Giới hạn sử dụng (0 = không giới hạn)
	UsageCount   int `json:"usage_count" gorm:"default:0"`    // Số lần đã sử dụng
	UsagePerUser int `json:"usage_per_user" gorm:"default:1"` // Số lần sử dụng mỗi user

	// Validity Period
	ValidFrom time.Time `json:"valid_from" gorm:"not null"`
	ValidTo   time.Time `json:"valid_to" gorm:"not null"`

	// Target Configuration
	TargetType string `json:"target_type" gorm:"size:20;default:'all'"` // all, product, category, brand, user
	TargetIDs  string `json:"target_ids" gorm:"type:text"`              // JSON array of target IDs

	// Additional Configuration
	IsStackable     bool `json:"is_stackable" gorm:"default:false"`       // Có thể kết hợp với coupon khác
	IsFirstTimeOnly bool `json:"is_first_time_only" gorm:"default:false"` // Chỉ cho lần mua đầu tiên
	IsNewUserOnly   bool `json:"is_new_user_only" gorm:"default:false"`   // Chỉ cho user mới

	// Metadata
	CreatedBy uint           `json:"created_by" gorm:"not null"`
	Creator   *User          `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relations
	Usages []CouponUsage `json:"usages,omitempty" gorm:"foreignKey:CouponID"`
}

// CouponUsage represents the usage of a coupon
type CouponUsage struct {
	ID       uint    `json:"id" gorm:"primaryKey"`
	CouponID uint    `json:"coupon_id" gorm:"not null;index"`
	Coupon   *Coupon `json:"coupon,omitempty" gorm:"foreignKey:CouponID"`
	UserID   uint    `json:"user_id" gorm:"not null;index"`
	User     *User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	OrderID  uint    `json:"order_id" gorm:"not null;index"`
	Order    *Order  `json:"order,omitempty" gorm:"foreignKey:OrderID"`

	// Usage Details
	DiscountAmount float64   `json:"discount_amount" gorm:"type:decimal(10,2);not null"`
	OrderAmount    float64   `json:"order_amount" gorm:"type:decimal(10,2);not null"`
	UsedAt         time.Time `json:"used_at" gorm:"not null"`

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// Point represents user points
type Point struct {
	ID            uint  `json:"id" gorm:"primaryKey"`
	UserID        uint  `json:"user_id" gorm:"not null;uniqueIndex"`
	User          *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Balance       int   `json:"balance" gorm:"default:0"`        // Số điểm hiện tại
	TotalEarned   int   `json:"total_earned" gorm:"default:0"`   // Tổng điểm đã tích
	TotalRedeemed int   `json:"total_redeemed" gorm:"default:0"` // Tổng điểm đã đổi
	TotalExpired  int   `json:"total_expired" gorm:"default:0"`  // Tổng điểm đã hết hạn

	// Point Configuration
	ExpiryDays int  `json:"expiry_days" gorm:"default:365"` // Số ngày hết hạn điểm
	IsActive   bool `json:"is_active" gorm:"default:true"`

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relations
	Transactions []PointTransaction `json:"transactions,omitempty" gorm:"foreignKey:PointID"`
}

// PointTransaction represents a point transaction
type PointTransaction struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	PointID uint   `json:"point_id" gorm:"not null;index"`
	Point   *Point `json:"point,omitempty" gorm:"foreignKey:PointID"`
	UserID  uint   `json:"user_id" gorm:"not null;index"`
	User    *User  `json:"user,omitempty" gorm:"foreignKey:UserID"`

	// Transaction Details
	Type    PointType   `json:"type" gorm:"size:20;not null"`
	Status  PointStatus `json:"status" gorm:"size:20;default:'pending'"`
	Amount  int         `json:"amount" gorm:"not null"`  // Số điểm (dương = tích, âm = đổi)
	Balance int         `json:"balance" gorm:"not null"` // Số dư sau giao dịch

	// Reference Information
	ReferenceType string `json:"reference_type" gorm:"size:50"` // order, coupon, manual, etc.
	ReferenceID   uint   `json:"reference_id"`                  // ID của đơn hàng, coupon, etc.
	OrderID       *uint  `json:"order_id" gorm:"index"`
	Order         *Order `json:"order,omitempty" gorm:"foreignKey:OrderID"`

	// Description
	Description string `json:"description" gorm:"type:text"`
	Notes       string `json:"notes" gorm:"type:text"`

	// Expiry
	ExpiresAt *time.Time `json:"expires_at"`

	// Metadata
	CreatedBy *uint          `json:"created_by"`
	Creator   *User          `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// Request/Response DTOs

// CouponCreateRequest represents the request body for creating a coupon
type CouponCreateRequest struct {
	Code              string     `json:"code" binding:"required,min=3,max=50"`
	Name              string     `json:"name" binding:"required,min=3,max=255"`
	Description       string     `json:"description" binding:"omitempty,max=1000"`
	Type              CouponType `json:"type" binding:"required,oneof=percentage fixed free_shipping buy_x_get_y"`
	DiscountValue     float64    `json:"discount_value" binding:"required,min=0"`
	MinOrderAmount    float64    `json:"min_order_amount" binding:"omitempty,min=0"`
	MaxDiscountAmount float64    `json:"max_discount_amount" binding:"omitempty,min=0"`
	UsageLimit        int        `json:"usage_limit" binding:"omitempty,min=0"`
	UsagePerUser      int        `json:"usage_per_user" binding:"omitempty,min=1"`
	ValidFrom         time.Time  `json:"valid_from" binding:"required"`
	ValidTo           time.Time  `json:"valid_to" binding:"required"`
	TargetType        string     `json:"target_type" binding:"omitempty,oneof=all product category brand user"`
	TargetIDs         []uint     `json:"target_ids" binding:"omitempty"`
	IsStackable       bool       `json:"is_stackable"`
	IsFirstTimeOnly   bool       `json:"is_first_time_only"`
	IsNewUserOnly     bool       `json:"is_new_user_only"`
}

// CouponUpdateRequest represents the request body for updating a coupon
type CouponUpdateRequest struct {
	Name              string       `json:"name" binding:"omitempty,min=3,max=255"`
	Description       string       `json:"description" binding:"omitempty,max=1000"`
	Status            CouponStatus `json:"status" binding:"omitempty,oneof=active inactive expired used"`
	DiscountValue     float64      `json:"discount_value" binding:"omitempty,min=0"`
	MinOrderAmount    float64      `json:"min_order_amount" binding:"omitempty,min=0"`
	MaxDiscountAmount float64      `json:"max_discount_amount" binding:"omitempty,min=0"`
	UsageLimit        int          `json:"usage_limit" binding:"omitempty,min=0"`
	UsagePerUser      int          `json:"usage_per_user" binding:"omitempty,min=1"`
	ValidFrom         *time.Time   `json:"valid_from"`
	ValidTo           *time.Time   `json:"valid_to"`
	TargetType        string       `json:"target_type" binding:"omitempty,oneof=all product category brand user"`
	TargetIDs         []uint       `json:"target_ids" binding:"omitempty"`
	IsStackable       *bool        `json:"is_stackable"`
	IsFirstTimeOnly   *bool        `json:"is_first_time_only"`
	IsNewUserOnly     *bool        `json:"is_new_user_only"`
}

// CouponResponse represents the response body for a coupon
type CouponResponse struct {
	ID                uint                  `json:"id"`
	Code              string                `json:"code"`
	Name              string                `json:"name"`
	Description       string                `json:"description"`
	Type              CouponType            `json:"type"`
	Status            CouponStatus          `json:"status"`
	DiscountValue     float64               `json:"discount_value"`
	MinOrderAmount    float64               `json:"min_order_amount"`
	MaxDiscountAmount float64               `json:"max_discount_amount"`
	UsageLimit        int                   `json:"usage_limit"`
	UsageCount        int                   `json:"usage_count"`
	UsagePerUser      int                   `json:"usage_per_user"`
	ValidFrom         time.Time             `json:"valid_from"`
	ValidTo           time.Time             `json:"valid_to"`
	TargetType        string                `json:"target_type"`
	TargetIDs         []uint                `json:"target_ids"`
	IsStackable       bool                  `json:"is_stackable"`
	IsFirstTimeOnly   bool                  `json:"is_first_time_only"`
	IsNewUserOnly     bool                  `json:"is_new_user_only"`
	CreatedBy         uint                  `json:"created_by"`
	Creator           *User                 `json:"creator,omitempty"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
	Usages            []CouponUsageResponse `json:"usages,omitempty"`
}

// CouponUsageResponse represents the response body for coupon usage
type CouponUsageResponse struct {
	ID             uint      `json:"id"`
	CouponID       uint      `json:"coupon_id"`
	UserID         uint      `json:"user_id"`
	User           *User     `json:"user,omitempty"`
	OrderID        uint      `json:"order_id"`
	Order          *Order    `json:"order,omitempty"`
	DiscountAmount float64   `json:"discount_amount"`
	OrderAmount    float64   `json:"order_amount"`
	UsedAt         time.Time `json:"used_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// PointResponse represents the response body for user points
type PointResponse struct {
	ID            uint      `json:"id"`
	UserID        uint      `json:"user_id"`
	User          *User     `json:"user,omitempty"`
	Balance       int       `json:"balance"`
	TotalEarned   int       `json:"total_earned"`
	TotalRedeemed int       `json:"total_redeemed"`
	TotalExpired  int       `json:"total_expired"`
	ExpiryDays    int       `json:"expiry_days"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PointTransactionResponse represents the response body for point transaction
type PointTransactionResponse struct {
	ID            uint        `json:"id"`
	PointID       uint        `json:"point_id"`
	UserID        uint        `json:"user_id"`
	User          *User       `json:"user,omitempty"`
	Type          PointType   `json:"type"`
	Status        PointStatus `json:"status"`
	Amount        int         `json:"amount"`
	Balance       int         `json:"balance"`
	ReferenceType string      `json:"reference_type"`
	ReferenceID   uint        `json:"reference_id"`
	OrderID       *uint       `json:"order_id"`
	Order         *Order      `json:"order,omitempty"`
	Description   string      `json:"description"`
	Notes         string      `json:"notes"`
	ExpiresAt     *time.Time  `json:"expires_at"`
	CreatedBy     *uint       `json:"created_by"`
	Creator       *User       `json:"creator,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// PointEarnRequest represents the request body for earning points
type PointEarnRequest struct {
	UserID        uint   `json:"user_id" binding:"required"`
	Amount        int    `json:"amount" binding:"required,min=1"`
	ReferenceType string `json:"reference_type" binding:"required"`
	ReferenceID   uint   `json:"reference_id" binding:"required"`
	Description   string `json:"description" binding:"required,min=5,max=255"`
	Notes         string `json:"notes" binding:"omitempty,max=500"`
	ExpiryDays    *int   `json:"expiry_days" binding:"omitempty,min=1,max=3650"`
}

// PointRedeemRequest represents the request body for redeeming points
type PointRedeemRequest struct {
	UserID        uint   `json:"user_id" binding:"required"`
	Amount        int    `json:"amount" binding:"required,min=1"`
	ReferenceType string `json:"reference_type" binding:"required"`
	ReferenceID   uint   `json:"reference_id" binding:"required"`
	Description   string `json:"description" binding:"required,min=5,max=255"`
	Notes         string `json:"notes" binding:"omitempty,max=500"`
}

// PointRefundRequest represents the request body for refunding points
type PointRefundRequest struct {
	UserID        uint   `json:"user_id" binding:"required"`
	Amount        int    `json:"amount" binding:"required,min=1"`
	ReferenceType string `json:"reference_type" binding:"required"`
	ReferenceID   uint   `json:"reference_id" binding:"required"`
	Description   string `json:"description" binding:"required,min=5,max=255"`
	Notes         string `json:"notes" binding:"omitempty,max=500"`
}

// PointAdjustRequest represents the request body for adjusting points (admin function)
type PointAdjustRequest struct {
	UserID      uint   `json:"user_id" binding:"required"`
	Amount      int    `json:"amount" binding:"required"` // Can be positive or negative
	Description string `json:"description" binding:"required,min=5,max=255"`
	Notes       string `json:"notes" binding:"omitempty,max=500"`
}

// PointExpireRequest represents the request body for expiring points (admin function)
type PointExpireRequest struct {
	UserID      uint   `json:"user_id" binding:"required"`
	Amount      int    `json:"amount" binding:"required,min=1"`
	Description string `json:"description" binding:"required,min=5,max=255"`
}

// CouponUseRequest represents the request body for using a coupon
type CouponUseRequest struct {
	Code        string  `json:"code" binding:"required"`
	UserID      uint    `json:"user_id" binding:"required"`
	OrderID     uint    `json:"order_id" binding:"required"`
	OrderAmount float64 `json:"order_amount" binding:"required,min=0"`
	ProductIDs  []uint  `json:"product_ids" binding:"omitempty"`
}

// CouponValidateRequest represents the request body for validating a coupon
type CouponValidateRequest struct {
	Code        string  `json:"code" binding:"required"`
	UserID      uint    `json:"user_id" binding:"required"`
	OrderAmount float64 `json:"order_amount" binding:"required,min=0"`
	ProductIDs  []uint  `json:"product_ids" binding:"omitempty"`
}

// CouponValidateResponse represents the response body for coupon validation
type CouponValidateResponse struct {
	Valid          bool            `json:"valid"`
	DiscountAmount float64         `json:"discount_amount"`
	Message        string          `json:"message"`
	Coupon         *CouponResponse `json:"coupon,omitempty"`
}

// PointStatsResponse represents point statistics
type PointStatsResponse struct {
	TotalUsers          int64           `json:"total_users"`
	ActiveUsers         int64           `json:"active_users"`
	TotalPointsIssued   int64           `json:"total_points_issued"`
	TotalPointsRedeemed int64           `json:"total_points_redeemed"`
	TotalPointsExpired  int64           `json:"total_points_expired"`
	AverageBalance      float64         `json:"average_balance"`
	TopEarners          []PointResponse `json:"top_earners"`
}

// CouponStatsResponse represents coupon statistics
type CouponStatsResponse struct {
	TotalCoupons    int64            `json:"total_coupons"`
	ActiveCoupons   int64            `json:"active_coupons"`
	ExpiredCoupons  int64            `json:"expired_coupons"`
	TotalUsages     int64            `json:"total_usages"`
	TotalDiscount   float64          `json:"total_discount"`
	MostUsedCoupons []CouponResponse `json:"most_used_coupons"`
}

// Helper methods

// IsValid checks if coupon is valid
func (c *Coupon) IsValid() bool {
	now := time.Now()
	return c.Status == CouponStatusActive &&
		now.After(c.ValidFrom) &&
		now.Before(c.ValidTo) &&
		(c.UsageLimit == 0 || c.UsageCount < c.UsageLimit)
}

// IsExpired checks if coupon is expired
func (c *Coupon) IsExpired() bool {
	return time.Now().After(c.ValidTo)
}

// CanUse checks if coupon can be used by a user
func (c *Coupon) CanUse(userID uint, orderAmount float64) bool {
	if !c.IsValid() {
		return false
	}

	if orderAmount < c.MinOrderAmount {
		return false
	}

	return true
}

// CalculateDiscount calculates discount amount
func (c *Coupon) CalculateDiscount(orderAmount float64) float64 {
	var discount float64

	switch c.Type {
	case CouponTypePercentage:
		discount = orderAmount * c.DiscountValue / 100
	case CouponTypeFixed:
		discount = c.DiscountValue
	case CouponTypeFreeShipping:
		discount = 0 // Will be handled separately
	case CouponTypeBuyXGetY:
		discount = 0 // Will be handled separately
	}

	// Apply max discount limit
	if c.MaxDiscountAmount > 0 && discount > c.MaxDiscountAmount {
		discount = c.MaxDiscountAmount
	}

	// Cannot discount more than order amount
	if discount > orderAmount {
		discount = orderAmount
	}

	return discount
}

// ToResponse converts Coupon to CouponResponse
func (c *Coupon) ToResponse() *CouponResponse {
	response := &CouponResponse{
		ID:                c.ID,
		Code:              c.Code,
		Name:              c.Name,
		Description:       c.Description,
		Type:              c.Type,
		Status:            c.Status,
		DiscountValue:     c.DiscountValue,
		MinOrderAmount:    c.MinOrderAmount,
		MaxDiscountAmount: c.MaxDiscountAmount,
		UsageLimit:        c.UsageLimit,
		UsageCount:        c.UsageCount,
		UsagePerUser:      c.UsagePerUser,
		ValidFrom:         c.ValidFrom,
		ValidTo:           c.ValidTo,
		TargetType:        c.TargetType,
		IsStackable:       c.IsStackable,
		IsFirstTimeOnly:   c.IsFirstTimeOnly,
		IsNewUserOnly:     c.IsNewUserOnly,
		CreatedBy:         c.CreatedBy,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}

	// Parse target IDs
	if c.TargetIDs != "" {
		// This would need JSON parsing in real implementation
		response.TargetIDs = []uint{}
	}

	// Add creator information
	if c.Creator != nil {
		response.Creator = c.Creator
	}

	// Add usages
	for _, usage := range c.Usages {
		response.Usages = append(response.Usages, *usage.ToResponse())
	}

	return response
}

// ToResponse converts CouponUsage to CouponUsageResponse
func (cu *CouponUsage) ToResponse() *CouponUsageResponse {
	response := &CouponUsageResponse{
		ID:             cu.ID,
		CouponID:       cu.CouponID,
		UserID:         cu.UserID,
		OrderID:        cu.OrderID,
		DiscountAmount: cu.DiscountAmount,
		OrderAmount:    cu.OrderAmount,
		UsedAt:         cu.UsedAt,
		CreatedAt:      cu.CreatedAt,
		UpdatedAt:      cu.UpdatedAt,
	}

	if cu.User != nil {
		response.User = cu.User
	}
	if cu.Order != nil {
		response.Order = cu.Order
	}

	return response
}

// ToResponse converts Point to PointResponse
func (p *Point) ToResponse() *PointResponse {
	response := &PointResponse{
		ID:            p.ID,
		UserID:        p.UserID,
		Balance:       p.Balance,
		TotalEarned:   p.TotalEarned,
		TotalRedeemed: p.TotalRedeemed,
		TotalExpired:  p.TotalExpired,
		ExpiryDays:    p.ExpiryDays,
		IsActive:      p.IsActive,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}

	if p.User != nil {
		response.User = p.User
	}

	return response
}

// ToResponse converts PointTransaction to PointTransactionResponse
func (pt *PointTransaction) ToResponse() *PointTransactionResponse {
	response := &PointTransactionResponse{
		ID:            pt.ID,
		PointID:       pt.PointID,
		UserID:        pt.UserID,
		Type:          pt.Type,
		Status:        pt.Status,
		Amount:        pt.Amount,
		Balance:       pt.Balance,
		ReferenceType: pt.ReferenceType,
		ReferenceID:   pt.ReferenceID,
		OrderID:       pt.OrderID,
		Description:   pt.Description,
		Notes:         pt.Notes,
		ExpiresAt:     pt.ExpiresAt,
		CreatedBy:     pt.CreatedBy,
		CreatedAt:     pt.CreatedAt,
		UpdatedAt:     pt.UpdatedAt,
	}

	if pt.User != nil {
		response.User = pt.User
	}
	if pt.Order != nil {
		response.Order = pt.Order
	}
	if pt.Creator != nil {
		response.Creator = pt.Creator
	}

	return response
}

// ValidateCoupon validates coupon data
func (c *Coupon) ValidateCoupon() error {
	if c.Code == "" {
		return errors.New("coupon code is required")
	}
	if len(c.Code) < 3 || len(c.Code) > 50 {
		return errors.New("coupon code must be between 3 and 50 characters")
	}
	if c.Name == "" {
		return errors.New("coupon name is required")
	}
	if len(c.Name) < 3 || len(c.Name) > 255 {
		return errors.New("coupon name must be between 3 and 255 characters")
	}
	if c.DiscountValue <= 0 {
		return errors.New("discount value must be greater than 0")
	}
	if c.Type == CouponTypePercentage && c.DiscountValue > 100 {
		return errors.New("percentage discount cannot exceed 100%")
	}
	if c.ValidTo.Before(c.ValidFrom) {
		return errors.New("valid to date must be after valid from date")
	}
	if c.UsageLimit < 0 {
		return errors.New("usage limit cannot be negative")
	}
	if c.UsagePerUser < 1 {
		return errors.New("usage per user must be at least 1")
	}
	return nil
}

// ValidatePointTransaction validates point transaction data
func (pt *PointTransaction) ValidatePointTransaction() error {
	if pt.Amount == 0 {
		return errors.New("amount cannot be zero")
	}
	if pt.Description == "" {
		return errors.New("description is required")
	}
	if len(pt.Description) < 5 || len(pt.Description) > 255 {
		return errors.New("description must be between 5 and 255 characters")
	}
	return nil
}

// GetTypeDisplayName returns display name for coupon type
func (c *Coupon) GetTypeDisplayName() string {
	typeMap := map[CouponType]string{
		CouponTypePercentage:   "Giảm theo phần trăm",
		CouponTypeFixed:        "Giảm số tiền cố định",
		CouponTypeFreeShipping: "Miễn phí vận chuyển",
		CouponTypeBuyXGetY:     "Mua X tặng Y",
	}
	return typeMap[c.Type]
}

// GetStatusDisplayName returns display name for coupon status
func (c *Coupon) GetStatusDisplayName() string {
	statusMap := map[CouponStatus]string{
		CouponStatusActive:   "Đang hoạt động",
		CouponStatusInactive: "Tạm dừng",
		CouponStatusExpired:  "Hết hạn",
		CouponStatusUsed:     "Đã sử dụng hết",
	}
	return statusMap[c.Status]
}

// GetPointTypeDisplayName returns display name for point type
func (pt *PointTransaction) GetPointTypeDisplayName() string {
	typeMap := map[PointType]string{
		PointTypeEarn:   "Tích điểm",
		PointTypeRedeem: "Đổi điểm",
		PointTypeExpire: "Hết hạn điểm",
		PointTypeRefund: "Hoàn điểm",
		PointTypeAdjust: "Điều chỉnh điểm",
	}
	return typeMap[pt.Type]
}

// GetPointStatusDisplayName returns display name for point status
func (pt *PointTransaction) GetPointStatusDisplayName() string {
	statusMap := map[PointStatus]string{
		PointStatusPending:   "Chờ xử lý",
		PointStatusCompleted: "Hoàn thành",
		PointStatusCancelled: "Đã hủy",
		PointStatusExpired:   "Đã hết hạn",
	}
	return statusMap[pt.Status]
}
