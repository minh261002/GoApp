package repository

import (
	"fmt"
	"go_app/internal/model"
	"go_app/pkg/database"
	"time"

	"gorm.io/gorm"
)

// CouponRepository defines methods for interacting with coupon data
type CouponRepository interface {
	// Basic CRUD
	CreateCoupon(coupon *model.Coupon) error
	GetCouponByID(id uint) (*model.Coupon, error)
	GetCouponByCode(code string) (*model.Coupon, error)
	GetAllCoupons(page, limit int, filters map[string]interface{}) ([]model.Coupon, int64, error)
	UpdateCoupon(coupon *model.Coupon) error
	DeleteCoupon(id uint) error

	// Coupon Management
	GetActiveCoupons(page, limit int) ([]model.Coupon, int64, error)
	GetExpiredCoupons(page, limit int) ([]model.Coupon, int64, error)
	GetCouponsByType(couponType model.CouponType, page, limit int) ([]model.Coupon, int64, error)
	GetCouponsByStatus(status model.CouponStatus, page, limit int) ([]model.Coupon, int64, error)
	SearchCoupons(query string, page, limit int) ([]model.Coupon, int64, error)
	ValidateCoupon(code string, userID uint, orderAmount float64, productIDs []uint) (*model.CouponValidateResponse, error)

	// Coupon Usage
	CreateCouponUsage(usage *model.CouponUsage) error
	GetCouponUsagesByCoupon(couponID uint, page, limit int) ([]model.CouponUsage, int64, error)
	GetCouponUsagesByUser(userID uint, page, limit int) ([]model.CouponUsage, int64, error)
	GetCouponUsagesByOrder(orderID uint) ([]model.CouponUsage, error)
	GetCouponUsageCount(couponID uint) (int64, error)
	GetUserCouponUsageCount(couponID, userID uint) (int64, error)

	// Statistics
	GetCouponStats() (*model.CouponStatsResponse, error)
	GetCouponUsageStats(couponID uint) (map[string]interface{}, error)
}

// PointRepository defines methods for interacting with point data
type PointRepository interface {
	// Basic CRUD
	CreatePoint(point *model.Point) error
	GetPointByID(id uint) (*model.Point, error)
	GetPointByUserID(userID uint) (*model.Point, error)
	UpdatePoint(point *model.Point) error
	DeletePoint(id uint) error

	// Point Transactions
	CreatePointTransaction(transaction *model.PointTransaction) error
	GetPointTransactionByID(id uint) (*model.PointTransaction, error)
	GetPointTransactionsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.PointTransaction, int64, error)
	GetPointTransactionsByPoint(pointID uint, page, limit int) ([]model.PointTransaction, int64, error)
	UpdatePointTransaction(transaction *model.PointTransaction) error
	DeletePointTransaction(id uint) error

	// Point Operations
	EarnPoints(userID uint, amount int, referenceType string, referenceID uint, description string, expiryDays *int) (*model.PointTransaction, error)
	RedeemPoints(userID uint, amount int, referenceType string, referenceID uint, description string) (*model.PointTransaction, error)
	RefundPoints(userID uint, amount int, referenceType string, referenceID uint, description string) (*model.PointTransaction, error)
	AdjustPoints(userID uint, amount int, description string, notes string) (*model.PointTransaction, error)
	ExpirePoints(userID uint, amount int, description string) (*model.PointTransaction, error)

	// Point Queries
	GetUserPointBalance(userID uint) (int, error)
	GetExpiredPoints(userID uint) ([]model.PointTransaction, error)
	GetExpiringPoints(userID uint, days int) ([]model.PointTransaction, error)
	GetPointHistory(userID uint, page, limit int) ([]model.PointTransaction, int64, error)

	// Statistics
	GetPointStats() (*model.PointStatsResponse, error)
	GetUserPointStats(userID uint) (map[string]interface{}, error)
	GetTopEarners(limit int) ([]model.Point, error)
}

// couponRepository implements CouponRepository
type couponRepository struct {
	db *gorm.DB
}

// pointRepository implements PointRepository
type pointRepository struct {
	db *gorm.DB
}

// NewCouponRepository creates a new CouponRepository
func NewCouponRepository() CouponRepository {
	return &couponRepository{
		db: database.DB,
	}
}

// NewPointRepository creates a new PointRepository
func NewPointRepository() PointRepository {
	return &pointRepository{
		db: database.DB,
	}
}

// Coupon Repository Implementation

// CreateCoupon creates a new coupon
func (r *couponRepository) CreateCoupon(coupon *model.Coupon) error {
	return r.db.Create(coupon).Error
}

// GetCouponByID retrieves a coupon by its ID
func (r *couponRepository) GetCouponByID(id uint) (*model.Coupon, error) {
	var coupon model.Coupon
	if err := r.db.Preload("Creator").Preload("Usages.User").Preload("Usages.Order").
		First(&coupon, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &coupon, nil
}

// GetCouponByCode retrieves a coupon by its code
func (r *couponRepository) GetCouponByCode(code string) (*model.Coupon, error) {
	var coupon model.Coupon
	if err := r.db.Preload("Creator").Preload("Usages.User").Preload("Usages.Order").
		Where("code = ?", code).First(&coupon).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &coupon, nil
}

// GetAllCoupons retrieves all coupons with pagination and filters
func (r *couponRepository) GetAllCoupons(page, limit int, filters map[string]interface{}) ([]model.Coupon, int64, error) {
	var coupons []model.Coupon
	var total int64
	db := r.db.Model(&model.Coupon{})

	// Apply filters
	for key, value := range filters {
		switch key {
		case "code":
			db = db.Where("code LIKE ?", fmt.Sprintf("%%%s%%", value.(string)))
		case "name":
			db = db.Where("name LIKE ?", fmt.Sprintf("%%%s%%", value.(string)))
		case "type":
			db = db.Where("type = ?", value)
		case "status":
			db = db.Where("status = ?", value)
		case "created_by":
			db = db.Where("created_by = ?", value)
		case "is_stackable":
			db = db.Where("is_stackable = ?", value)
		case "is_first_time_only":
			db = db.Where("is_first_time_only = ?", value)
		case "is_new_user_only":
			db = db.Where("is_new_user_only = ?", value)
		case "valid_from":
			db = db.Where("valid_from >= ?", value)
		case "valid_to":
			db = db.Where("valid_to <= ?", value)
		case "search":
			searchTerm := fmt.Sprintf("%%%s%%", value.(string))
			db = db.Where("code LIKE ? OR name LIKE ? OR description LIKE ?", searchTerm, searchTerm, searchTerm)
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

	if err := db.Preload("Creator").Preload("Usages.User").Preload("Usages.Order").
		Find(&coupons).Error; err != nil {
		return nil, 0, err
	}

	return coupons, total, nil
}

// UpdateCoupon updates an existing coupon
func (r *couponRepository) UpdateCoupon(coupon *model.Coupon) error {
	return r.db.Save(coupon).Error
}

// DeleteCoupon soft deletes a coupon
func (r *couponRepository) DeleteCoupon(id uint) error {
	return r.db.Delete(&model.Coupon{}, id).Error
}

// GetActiveCoupons retrieves active coupons
func (r *couponRepository) GetActiveCoupons(page, limit int) ([]model.Coupon, int64, error) {
	var coupons []model.Coupon
	var total int64
	now := time.Now()

	db := r.db.Model(&model.Coupon{}).Where("status = ? AND valid_from <= ? AND valid_to >= ?",
		model.CouponStatusActive, now, now)

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

	if err := db.Preload("Creator").Preload("Usages.User").Preload("Usages.Order").
		Find(&coupons).Error; err != nil {
		return nil, 0, err
	}

	return coupons, total, nil
}

// GetExpiredCoupons retrieves expired coupons
func (r *couponRepository) GetExpiredCoupons(page, limit int) ([]model.Coupon, int64, error) {
	var coupons []model.Coupon
	var total int64
	now := time.Now()

	db := r.db.Model(&model.Coupon{}).Where("valid_to < ?", now)

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
	db = db.Order("valid_to DESC")

	if err := db.Preload("Creator").Preload("Usages.User").Preload("Usages.Order").
		Find(&coupons).Error; err != nil {
		return nil, 0, err
	}

	return coupons, total, nil
}

// GetCouponsByType retrieves coupons by type
func (r *couponRepository) GetCouponsByType(couponType model.CouponType, page, limit int) ([]model.Coupon, int64, error) {
	var coupons []model.Coupon
	var total int64
	db := r.db.Model(&model.Coupon{}).Where("type = ?", couponType)

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

	if err := db.Preload("Creator").Preload("Usages.User").Preload("Usages.Order").
		Find(&coupons).Error; err != nil {
		return nil, 0, err
	}

	return coupons, total, nil
}

// GetCouponsByStatus retrieves coupons by status
func (r *couponRepository) GetCouponsByStatus(status model.CouponStatus, page, limit int) ([]model.Coupon, int64, error) {
	var coupons []model.Coupon
	var total int64
	db := r.db.Model(&model.Coupon{}).Where("status = ?", status)

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

	if err := db.Preload("Creator").Preload("Usages.User").Preload("Usages.Order").
		Find(&coupons).Error; err != nil {
		return nil, 0, err
	}

	return coupons, total, nil
}

// SearchCoupons performs full-text search on coupons
func (r *couponRepository) SearchCoupons(query string, page, limit int) ([]model.Coupon, int64, error) {
	var coupons []model.Coupon
	var total int64

	// Use MATCH AGAINST for full-text search
	db := r.db.Model(&model.Coupon{}).
		Where("MATCH(code, name, description) AGAINST(? IN NATURAL LANGUAGE MODE)", query)

	// Count total records
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting by relevance and date
	db = db.Order("MATCH(code, name, description) AGAINST('" + query + "' IN NATURAL LANGUAGE MODE) DESC, created_at DESC")

	if err := db.Preload("Creator").Preload("Usages.User").Preload("Usages.Order").
		Find(&coupons).Error; err != nil {
		return nil, 0, err
	}

	return coupons, total, nil
}

// ValidateCoupon validates a coupon for use
func (r *couponRepository) ValidateCoupon(code string, userID uint, orderAmount float64, productIDs []uint) (*model.CouponValidateResponse, error) {
	coupon, err := r.GetCouponByCode(code)
	if err != nil {
		return &model.CouponValidateResponse{
			Valid:   false,
			Message: "Failed to retrieve coupon",
		}, err
	}

	if coupon == nil {
		return &model.CouponValidateResponse{
			Valid:   false,
			Message: "Coupon not found",
		}, nil
	}

	// Check if coupon is valid
	if !coupon.IsValid() {
		return &model.CouponValidateResponse{
			Valid:   false,
			Message: "Coupon is not valid or has expired",
		}, nil
	}

	// Check if coupon can be used
	if !coupon.CanUse(userID, orderAmount) {
		return &model.CouponValidateResponse{
			Valid:   false,
			Message: "Coupon cannot be used for this order",
		}, nil
	}

	// Check usage per user
	userUsageCount, err := r.GetUserCouponUsageCount(coupon.ID, userID)
	if err != nil {
		return &model.CouponValidateResponse{
			Valid:   false,
			Message: "Failed to check usage count",
		}, err
	}

	if userUsageCount >= int64(coupon.UsagePerUser) {
		return &model.CouponValidateResponse{
			Valid:   false,
			Message: "Coupon usage limit reached for this user",
		}, nil
	}

	// Calculate discount amount
	discountAmount := coupon.CalculateDiscount(orderAmount)

	return &model.CouponValidateResponse{
		Valid:          true,
		DiscountAmount: discountAmount,
		Message:        "Coupon is valid",
		Coupon:         coupon.ToResponse(),
	}, nil
}

// Coupon Usage Methods

// CreateCouponUsage creates a new coupon usage record
func (r *couponRepository) CreateCouponUsage(usage *model.CouponUsage) error {
	return r.db.Create(usage).Error
}

// GetCouponUsagesByCoupon retrieves coupon usages for a specific coupon
func (r *couponRepository) GetCouponUsagesByCoupon(couponID uint, page, limit int) ([]model.CouponUsage, int64, error) {
	var usages []model.CouponUsage
	var total int64
	db := r.db.Model(&model.CouponUsage{}).Where("coupon_id = ?", couponID)

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
	db = db.Order("used_at DESC")

	if err := db.Preload("User").Preload("Order").Preload("Coupon").
		Find(&usages).Error; err != nil {
		return nil, 0, err
	}

	return usages, total, nil
}

// GetCouponUsagesByUser retrieves coupon usages for a specific user
func (r *couponRepository) GetCouponUsagesByUser(userID uint, page, limit int) ([]model.CouponUsage, int64, error) {
	var usages []model.CouponUsage
	var total int64
	db := r.db.Model(&model.CouponUsage{}).Where("user_id = ?", userID)

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
	db = db.Order("used_at DESC")

	if err := db.Preload("User").Preload("Order").Preload("Coupon").
		Find(&usages).Error; err != nil {
		return nil, 0, err
	}

	return usages, total, nil
}

// GetCouponUsagesByOrder retrieves coupon usages for a specific order
func (r *couponRepository) GetCouponUsagesByOrder(orderID uint) ([]model.CouponUsage, error) {
	var usages []model.CouponUsage
	err := r.db.Where("order_id = ?", orderID).
		Preload("User").Preload("Order").Preload("Coupon").
		Order("used_at DESC").
		Find(&usages).Error
	return usages, err
}

// GetCouponUsageCount retrieves usage count for a coupon
func (r *couponRepository) GetCouponUsageCount(couponID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.CouponUsage{}).Where("coupon_id = ?", couponID).Count(&count).Error
	return count, err
}

// GetUserCouponUsageCount retrieves usage count for a user and coupon
func (r *couponRepository) GetUserCouponUsageCount(couponID, userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.CouponUsage{}).Where("coupon_id = ? AND user_id = ?", couponID, userID).Count(&count).Error
	return count, err
}

// Statistics

// GetCouponStats retrieves coupon statistics
func (r *couponRepository) GetCouponStats() (*model.CouponStatsResponse, error) {
	var stats model.CouponStatsResponse
	var count int64

	// Total coupons
	r.db.Model(&model.Coupon{}).Count(&count)
	stats.TotalCoupons = count

	// Active coupons
	r.db.Model(&model.Coupon{}).Where("status = ?", model.CouponStatusActive).Count(&count)
	stats.ActiveCoupons = count

	// Expired coupons
	now := time.Now()
	r.db.Model(&model.Coupon{}).Where("valid_to < ?", now).Count(&count)
	stats.ExpiredCoupons = count

	// Total usages
	r.db.Model(&model.CouponUsage{}).Count(&count)
	stats.TotalUsages = count

	// Total discount
	var totalDiscount float64
	r.db.Model(&model.CouponUsage{}).Select("SUM(discount_amount)").Scan(&totalDiscount)
	stats.TotalDiscount = totalDiscount

	// Most used coupons
	var mostUsedCoupons []model.Coupon
	r.db.Model(&model.Coupon{}).
		Select("coupons.*, COUNT(coupon_usages.id) as usage_count").
		Joins("LEFT JOIN coupon_usages ON coupons.id = coupon_usages.coupon_id").
		Group("coupons.id").
		Order("usage_count DESC").
		Limit(10).
		Preload("Creator").
		Find(&mostUsedCoupons)

	for _, coupon := range mostUsedCoupons {
		stats.MostUsedCoupons = append(stats.MostUsedCoupons, *coupon.ToResponse())
	}

	return &stats, nil
}

// GetCouponUsageStats retrieves usage statistics for a specific coupon
func (r *couponRepository) GetCouponUsageStats(couponID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var count int64

	// Total usages for coupon
	r.db.Model(&model.CouponUsage{}).Where("coupon_id = ?", couponID).Count(&count)
	stats["total_usages"] = count

	// Total discount for coupon
	var totalDiscount float64
	r.db.Model(&model.CouponUsage{}).Where("coupon_id = ?", couponID).Select("SUM(discount_amount)").Scan(&totalDiscount)
	stats["total_discount"] = totalDiscount

	// Average discount per usage
	var avgDiscount float64
	r.db.Model(&model.CouponUsage{}).Where("coupon_id = ?", couponID).Select("AVG(discount_amount)").Scan(&avgDiscount)
	stats["average_discount"] = avgDiscount

	// Unique users
	var uniqueUsers int64
	r.db.Model(&model.CouponUsage{}).Where("coupon_id = ?", couponID).Distinct("user_id").Count(&uniqueUsers)
	stats["unique_users"] = uniqueUsers

	// Recent usages (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	r.db.Model(&model.CouponUsage{}).Where("coupon_id = ? AND used_at >= ?", couponID, thirtyDaysAgo).Count(&count)
	stats["recent_usages"] = count

	return stats, nil
}

// Point Repository Implementation

// CreatePoint creates a new point record
func (r *pointRepository) CreatePoint(point *model.Point) error {
	return r.db.Create(point).Error
}

// GetPointByID retrieves a point record by its ID
func (r *pointRepository) GetPointByID(id uint) (*model.Point, error) {
	var point model.Point
	if err := r.db.Preload("User").Preload("Transactions.Creator").
		First(&point, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &point, nil
}

// GetPointByUserID retrieves a point record by user ID
func (r *pointRepository) GetPointByUserID(userID uint) (*model.Point, error) {
	var point model.Point
	if err := r.db.Preload("User").Preload("Transactions.Creator").
		Where("user_id = ?", userID).First(&point).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &point, nil
}

// UpdatePoint updates an existing point record
func (r *pointRepository) UpdatePoint(point *model.Point) error {
	return r.db.Save(point).Error
}

// DeletePoint soft deletes a point record
func (r *pointRepository) DeletePoint(id uint) error {
	return r.db.Delete(&model.Point{}, id).Error
}

// Point Transaction Methods

// CreatePointTransaction creates a new point transaction
func (r *pointRepository) CreatePointTransaction(transaction *model.PointTransaction) error {
	return r.db.Create(transaction).Error
}

// GetPointTransactionByID retrieves a point transaction by its ID
func (r *pointRepository) GetPointTransactionByID(id uint) (*model.PointTransaction, error) {
	var transaction model.PointTransaction
	if err := r.db.Preload("User").Preload("Order").Preload("Creator").
		First(&transaction, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &transaction, nil
}

// GetPointTransactionsByUser retrieves point transactions for a specific user
func (r *pointRepository) GetPointTransactionsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.PointTransaction, int64, error) {
	var transactions []model.PointTransaction
	var total int64
	db := r.db.Model(&model.PointTransaction{}).Where("user_id = ?", userID)

	// Apply filters
	for key, value := range filters {
		switch key {
		case "type":
			db = db.Where("type = ?", value)
		case "status":
			db = db.Where("status = ?", value)
		case "reference_type":
			db = db.Where("reference_type = ?", value)
		case "reference_id":
			db = db.Where("reference_id = ?", value)
		case "order_id":
			db = db.Where("order_id = ?", value)
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

	if err := db.Preload("User").Preload("Order").Preload("Creator").
		Find(&transactions).Error; err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// GetPointTransactionsByPoint retrieves point transactions for a specific point record
func (r *pointRepository) GetPointTransactionsByPoint(pointID uint, page, limit int) ([]model.PointTransaction, int64, error) {
	var transactions []model.PointTransaction
	var total int64
	db := r.db.Model(&model.PointTransaction{}).Where("point_id = ?", pointID)

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

	if err := db.Preload("User").Preload("Order").Preload("Creator").
		Find(&transactions).Error; err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// UpdatePointTransaction updates an existing point transaction
func (r *pointRepository) UpdatePointTransaction(transaction *model.PointTransaction) error {
	return r.db.Save(transaction).Error
}

// DeletePointTransaction soft deletes a point transaction
func (r *pointRepository) DeletePointTransaction(id uint) error {
	return r.db.Delete(&model.PointTransaction{}, id).Error
}

// Point Operations

// EarnPoints earns points for a user
func (r *pointRepository) EarnPoints(userID uint, amount int, referenceType string, referenceID uint, description string, expiryDays *int) (*model.PointTransaction, error) {
	var transaction *model.PointTransaction
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Get or create point record for user
		var point model.Point
		if err := tx.Where("user_id = ?", userID).First(&point).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create new point record
				point = model.Point{
					UserID:      userID,
					Balance:     0,
					TotalEarned: 0,
					ExpiryDays:  365,
					IsActive:    true,
				}
				if err := tx.Create(&point).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		// Calculate expiry date
		var expiresAt *time.Time
		if expiryDays != nil {
			expiry := time.Now().AddDate(0, 0, *expiryDays)
			expiresAt = &expiry
		} else {
			expiry := time.Now().AddDate(0, 0, point.ExpiryDays)
			expiresAt = &expiry
		}

		// Create transaction
		transaction = &model.PointTransaction{
			PointID:       point.ID,
			UserID:        userID,
			Type:          model.PointTypeEarn,
			Status:        model.PointStatusCompleted,
			Amount:        amount,
			Balance:       point.Balance + amount,
			ReferenceType: referenceType,
			ReferenceID:   referenceID,
			Description:   description,
			ExpiresAt:     expiresAt,
		}

		if err := tx.Create(transaction).Error; err != nil {
			return err
		}

		// Update point balance
		point.Balance += amount
		point.TotalEarned += amount
		if err := tx.Save(&point).Error; err != nil {
			return err
		}

		// Update transaction balance
		transaction.Balance = point.Balance
		if err := tx.Save(transaction).Error; err != nil {
			return err
		}

		return nil
	})
	return transaction, err
}

// RedeemPoints redeems points for a user
func (r *pointRepository) RedeemPoints(userID uint, amount int, referenceType string, referenceID uint, description string) (*model.PointTransaction, error) {
	var transaction *model.PointTransaction
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Get point record for user
		var point model.Point
		if err := tx.Where("user_id = ?", userID).First(&point).Error; err != nil {
			return err
		}

		// Check if user has enough points
		if point.Balance < amount {
			return fmt.Errorf("insufficient points: have %d, need %d", point.Balance, amount)
		}

		// Create transaction
		transaction = &model.PointTransaction{
			PointID:       point.ID,
			UserID:        userID,
			Type:          model.PointTypeRedeem,
			Status:        model.PointStatusCompleted,
			Amount:        -amount, // Negative for redemption
			Balance:       point.Balance - amount,
			ReferenceType: referenceType,
			ReferenceID:   referenceID,
			Description:   description,
		}

		if err := tx.Create(transaction).Error; err != nil {
			return err
		}

		// Update point balance
		point.Balance -= amount
		point.TotalRedeemed += amount
		if err := tx.Save(&point).Error; err != nil {
			return err
		}

		// Update transaction balance
		transaction.Balance = point.Balance
		if err := tx.Save(transaction).Error; err != nil {
			return err
		}

		return nil
	})
	return transaction, err
}

// RefundPoints refunds points to a user
func (r *pointRepository) RefundPoints(userID uint, amount int, referenceType string, referenceID uint, description string) (*model.PointTransaction, error) {
	var transaction *model.PointTransaction
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Get point record for user
		var point model.Point
		if err := tx.Where("user_id = ?", userID).First(&point).Error; err != nil {
			return err
		}

		// Create transaction
		transaction = &model.PointTransaction{
			PointID:       point.ID,
			UserID:        userID,
			Type:          model.PointTypeRefund,
			Status:        model.PointStatusCompleted,
			Amount:        amount,
			Balance:       point.Balance + amount,
			ReferenceType: referenceType,
			ReferenceID:   referenceID,
			Description:   description,
		}

		if err := tx.Create(transaction).Error; err != nil {
			return err
		}

		// Update point balance
		point.Balance += amount
		if err := tx.Save(&point).Error; err != nil {
			return err
		}

		// Update transaction balance
		transaction.Balance = point.Balance
		if err := tx.Save(transaction).Error; err != nil {
			return err
		}

		return nil
	})
	return transaction, err
}

// AdjustPoints adjusts points for a user (admin function)
func (r *pointRepository) AdjustPoints(userID uint, amount int, description string, notes string) (*model.PointTransaction, error) {
	var transaction *model.PointTransaction
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Get point record for user
		var point model.Point
		if err := tx.Where("user_id = ?", userID).First(&point).Error; err != nil {
			return err
		}

		// Check if adjustment would result in negative balance
		if point.Balance+amount < 0 {
			return fmt.Errorf("adjustment would result in negative balance")
		}

		// Create transaction
		transaction = &model.PointTransaction{
			PointID:       point.ID,
			UserID:        userID,
			Type:          model.PointTypeAdjust,
			Status:        model.PointStatusCompleted,
			Amount:        amount,
			Balance:       point.Balance + amount,
			ReferenceType: "manual",
			Description:   description,
			Notes:         notes,
		}

		if err := tx.Create(transaction).Error; err != nil {
			return err
		}

		// Update point balance
		point.Balance += amount
		if amount > 0 {
			point.TotalEarned += amount
		} else {
			point.TotalRedeemed += -amount
		}
		if err := tx.Save(&point).Error; err != nil {
			return err
		}

		// Update transaction balance
		transaction.Balance = point.Balance
		if err := tx.Save(transaction).Error; err != nil {
			return err
		}

		return nil
	})
	return transaction, err
}

// ExpirePoints expires points for a user
func (r *pointRepository) ExpirePoints(userID uint, amount int, description string) (*model.PointTransaction, error) {
	var transaction *model.PointTransaction
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Get point record for user
		var point model.Point
		if err := tx.Where("user_id = ?", userID).First(&point).Error; err != nil {
			return err
		}

		// Ensure we don't expire more than available
		if amount > point.Balance {
			amount = point.Balance
		}

		// Create transaction
		transaction = &model.PointTransaction{
			PointID:       point.ID,
			UserID:        userID,
			Type:          model.PointTypeExpire,
			Status:        model.PointStatusCompleted,
			Amount:        -amount, // Negative for expiry
			Balance:       point.Balance - amount,
			ReferenceType: "expiry",
			Description:   description,
		}

		if err := tx.Create(transaction).Error; err != nil {
			return err
		}

		// Update point balance
		point.Balance -= amount
		point.TotalExpired += amount
		if err := tx.Save(&point).Error; err != nil {
			return err
		}

		// Update transaction balance
		transaction.Balance = point.Balance
		if err := tx.Save(transaction).Error; err != nil {
			return err
		}

		return nil
	})
	return transaction, err
}

// Point Queries

// GetUserPointBalance retrieves point balance for a user
func (r *pointRepository) GetUserPointBalance(userID uint) (int, error) {
	var point model.Point
	if err := r.db.Where("user_id = ?", userID).First(&point).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return point.Balance, nil
}

// GetExpiredPoints retrieves expired points for a user
func (r *pointRepository) GetExpiredPoints(userID uint) ([]model.PointTransaction, error) {
	var transactions []model.PointTransaction
	now := time.Now()
	err := r.db.Where("user_id = ? AND expires_at < ? AND type = ?", userID, now, model.PointTypeEarn).
		Preload("User").Preload("Order").Preload("Creator").
		Order("expires_at ASC").
		Find(&transactions).Error
	return transactions, err
}

// GetExpiringPoints retrieves points expiring within specified days
func (r *pointRepository) GetExpiringPoints(userID uint, days int) ([]model.PointTransaction, error) {
	var transactions []model.PointTransaction
	expiryDate := time.Now().AddDate(0, 0, days)
	err := r.db.Where("user_id = ? AND expires_at BETWEEN ? AND ? AND type = ?",
		userID, time.Now(), expiryDate, model.PointTypeEarn).
		Preload("User").Preload("Order").Preload("Creator").
		Order("expires_at ASC").
		Find(&transactions).Error
	return transactions, err
}

// GetPointHistory retrieves point history for a user
func (r *pointRepository) GetPointHistory(userID uint, page, limit int) ([]model.PointTransaction, int64, error) {
	return r.GetPointTransactionsByUser(userID, page, limit, map[string]interface{}{})
}

// Statistics

// GetPointStats retrieves point statistics
func (r *pointRepository) GetPointStats() (*model.PointStatsResponse, error) {
	var stats model.PointStatsResponse
	var count int64

	// Total users with points
	r.db.Model(&model.Point{}).Count(&count)
	stats.TotalUsers = count

	// Active users (with positive balance)
	r.db.Model(&model.Point{}).Where("balance > 0 AND is_active = ?", true).Count(&count)
	stats.ActiveUsers = count

	// Total points issued
	var totalEarned int64
	r.db.Model(&model.Point{}).Select("SUM(total_earned)").Scan(&totalEarned)
	stats.TotalPointsIssued = totalEarned

	// Total points redeemed
	var totalRedeemed int64
	r.db.Model(&model.Point{}).Select("SUM(total_redeemed)").Scan(&totalRedeemed)
	stats.TotalPointsRedeemed = totalRedeemed

	// Total points expired
	var totalExpired int64
	r.db.Model(&model.Point{}).Select("SUM(total_expired)").Scan(&totalExpired)
	stats.TotalPointsExpired = totalExpired

	// Average balance
	var avgBalance float64
	r.db.Model(&model.Point{}).Select("AVG(balance)").Scan(&avgBalance)
	stats.AverageBalance = avgBalance

	// Top earners
	topEarners, err := r.GetTopEarners(10)
	if err != nil {
		return nil, err
	}
	for _, point := range topEarners {
		stats.TopEarners = append(stats.TopEarners, *point.ToResponse())
	}

	return &stats, nil
}

// GetUserPointStats retrieves point statistics for a specific user
func (r *pointRepository) GetUserPointStats(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var point model.Point
	if err := r.db.Where("user_id = ?", userID).First(&point).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			stats["balance"] = 0
			stats["total_earned"] = 0
			stats["total_redeemed"] = 0
			stats["total_expired"] = 0
			return stats, nil
		}
		return nil, err
	}

	stats["balance"] = point.Balance
	stats["total_earned"] = point.TotalEarned
	stats["total_redeemed"] = point.TotalRedeemed
	stats["total_expired"] = point.TotalExpired
	stats["is_active"] = point.IsActive
	stats["expiry_days"] = point.ExpiryDays

	// Recent transactions count
	var recentCount int64
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	r.db.Model(&model.PointTransaction{}).Where("user_id = ? AND created_at >= ?", userID, thirtyDaysAgo).Count(&recentCount)
	stats["recent_transactions"] = recentCount

	// Expiring points count
	var expiringCount int64
	sevenDaysFromNow := time.Now().AddDate(0, 0, 7)
	r.db.Model(&model.PointTransaction{}).Where("user_id = ? AND expires_at BETWEEN ? AND ? AND type = ?",
		userID, time.Now(), sevenDaysFromNow, model.PointTypeEarn).Count(&expiringCount)
	stats["expiring_points"] = expiringCount

	return stats, nil
}

// GetTopEarners retrieves top earning users
func (r *pointRepository) GetTopEarners(limit int) ([]model.Point, error) {
	var points []model.Point
	err := r.db.Model(&model.Point{}).
		Where("is_active = ?", true).
		Order("total_earned DESC").
		Limit(limit).
		Preload("User").
		Find(&points).Error
	return points, err
}
