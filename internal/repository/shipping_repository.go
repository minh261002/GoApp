package repository

import (
	"go_app/internal/model"

	"gorm.io/gorm"
)

type ShippingRepository interface {
	// Shipping Providers
	CreateShippingProvider(provider *model.ShippingProvider) error
	GetShippingProviderByID(id uint) (*model.ShippingProvider, error)
	GetShippingProviderByCode(code string) (*model.ShippingProvider, error)
	GetAllShippingProviders() ([]model.ShippingProvider, error)
	GetActiveShippingProviders() ([]model.ShippingProvider, error)
	UpdateShippingProvider(provider *model.ShippingProvider) error
	DeleteShippingProvider(id uint) error

	// Shipping Rates
	CreateShippingRate(rate *model.ShippingRate) error
	GetShippingRateByID(id uint) (*model.ShippingRate, error)
	GetShippingRatesByProvider(providerID uint) ([]model.ShippingRate, error)
	GetShippingRatesForCalculation(req *model.CalculateShippingRequest) ([]model.ShippingRate, error)
	UpdateShippingRate(rate *model.ShippingRate) error
	DeleteShippingRate(id uint) error

	// Shipping Orders
	CreateShippingOrder(order *model.ShippingOrder) error
	GetShippingOrderByID(id uint) (*model.ShippingOrder, error)
	GetShippingOrderByOrderID(orderID uint) (*model.ShippingOrder, error)
	GetShippingOrderByExternalID(externalID string) (*model.ShippingOrder, error)
	GetShippingOrderByLabelID(labelID string) (*model.ShippingOrder, error)
	GetShippingOrderByTrackingCode(trackingCode string) (*model.ShippingOrder, error)
	UpdateShippingOrder(order *model.ShippingOrder) error
	DeleteShippingOrder(id uint) error
	GetShippingOrders(page, limit int, filters map[string]interface{}) ([]model.ShippingOrder, int64, error)

	// Shipping Tracking
	CreateShippingTracking(tracking *model.ShippingTracking) error
	GetShippingTrackingByOrderID(orderID uint) ([]model.ShippingTracking, error)
	GetShippingTrackingByShippingOrderID(shippingOrderID uint) ([]model.ShippingTracking, error)

	// Statistics
	GetShippingStats() (*model.ShippingStats, error)
	GetShippingStatsByProvider(providerID uint) (*model.ShippingStats, error)
}

type shippingRepository struct {
	db *gorm.DB
}

func NewShippingRepository(db *gorm.DB) ShippingRepository {
	return &shippingRepository{db: db}
}

// Shipping Providers
func (r *shippingRepository) CreateShippingProvider(provider *model.ShippingProvider) error {
	return r.db.Create(provider).Error
}

func (r *shippingRepository) GetShippingProviderByID(id uint) (*model.ShippingProvider, error) {
	var provider model.ShippingProvider
	err := r.db.Where("id = ?", id).First(&provider).Error
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *shippingRepository) GetShippingProviderByCode(code string) (*model.ShippingProvider, error) {
	var provider model.ShippingProvider
	err := r.db.Where("code = ? AND deleted_at IS NULL", code).First(&provider).Error
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *shippingRepository) GetAllShippingProviders() ([]model.ShippingProvider, error) {
	var providers []model.ShippingProvider
	err := r.db.Where("deleted_at IS NULL").Order("priority DESC, created_at ASC").Find(&providers).Error
	return providers, err
}

func (r *shippingRepository) GetActiveShippingProviders() ([]model.ShippingProvider, error) {
	var providers []model.ShippingProvider
	err := r.db.Where("is_active = ? AND deleted_at IS NULL", true).Order("priority DESC, created_at ASC").Find(&providers).Error
	return providers, err
}

func (r *shippingRepository) UpdateShippingProvider(provider *model.ShippingProvider) error {
	return r.db.Save(provider).Error
}

func (r *shippingRepository) DeleteShippingProvider(id uint) error {
	return r.db.Delete(&model.ShippingProvider{}, id).Error
}

// Shipping Rates
func (r *shippingRepository) CreateShippingRate(rate *model.ShippingRate) error {
	return r.db.Create(rate).Error
}

func (r *shippingRepository) GetShippingRateByID(id uint) (*model.ShippingRate, error) {
	var rate model.ShippingRate
	err := r.db.Preload("Provider").Where("id = ?", id).First(&rate).Error
	if err != nil {
		return nil, err
	}
	return &rate, nil
}

func (r *shippingRepository) GetShippingRatesByProvider(providerID uint) ([]model.ShippingRate, error) {
	var rates []model.ShippingRate
	err := r.db.Where("provider_id = ? AND is_active = ? AND deleted_at IS NULL", providerID, true).Find(&rates).Error
	return rates, err
}

func (r *shippingRepository) GetShippingRatesForCalculation(req *model.CalculateShippingRequest) ([]model.ShippingRate, error) {
	var rates []model.ShippingRate

	query := r.db.Where(`
		from_province = ? AND from_district = ? AND 
		to_province = ? AND to_district = ? AND
		? BETWEEN min_weight AND max_weight AND
		? BETWEEN min_value AND max_value AND
		is_active = ? AND deleted_at IS NULL`,
		req.FromProvince, req.FromDistrict,
		req.ToProvince, req.ToDistrict,
		req.Weight, req.Value, true)

	if req.ProviderID != nil {
		query = query.Where("provider_id = ?", *req.ProviderID)
	}

	err := query.Preload("Provider").Find(&rates).Error
	return rates, err
}

func (r *shippingRepository) UpdateShippingRate(rate *model.ShippingRate) error {
	return r.db.Save(rate).Error
}

func (r *shippingRepository) DeleteShippingRate(id uint) error {
	return r.db.Delete(&model.ShippingRate{}, id).Error
}

// Shipping Orders
func (r *shippingRepository) CreateShippingOrder(order *model.ShippingOrder) error {
	return r.db.Create(order).Error
}

func (r *shippingRepository) GetShippingOrderByID(id uint) (*model.ShippingOrder, error) {
	var order model.ShippingOrder
	err := r.db.Preload("Order").Preload("Provider").Where("id = ?", id).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *shippingRepository) GetShippingOrderByOrderID(orderID uint) (*model.ShippingOrder, error) {
	var order model.ShippingOrder
	err := r.db.Preload("Order").Preload("Provider").Where("order_id = ?", orderID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *shippingRepository) GetShippingOrderByExternalID(externalID string) (*model.ShippingOrder, error) {
	var order model.ShippingOrder
	err := r.db.Preload("Order").Preload("Provider").Where("external_id = ?", externalID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *shippingRepository) GetShippingOrderByLabelID(labelID string) (*model.ShippingOrder, error) {
	var order model.ShippingOrder
	err := r.db.Preload("Order").Preload("Provider").Where("label_id = ?", labelID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *shippingRepository) GetShippingOrderByTrackingCode(trackingCode string) (*model.ShippingOrder, error) {
	var order model.ShippingOrder
	err := r.db.Preload("Order").Preload("Provider").Where("tracking_code = ?", trackingCode).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *shippingRepository) UpdateShippingOrder(order *model.ShippingOrder) error {
	return r.db.Save(order).Error
}

func (r *shippingRepository) DeleteShippingOrder(id uint) error {
	return r.db.Delete(&model.ShippingOrder{}, id).Error
}

func (r *shippingRepository) GetShippingOrders(page, limit int, filters map[string]interface{}) ([]model.ShippingOrder, int64, error) {
	var orders []model.ShippingOrder
	var total int64

	query := r.db.Model(&model.ShippingOrder{}).Preload("Order").Preload("Provider")

	// Apply filters
	if providerID, ok := filters["provider_id"]; ok {
		query = query.Where("provider_id = ?", providerID)
	}
	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}
	if orderID, ok := filters["order_id"]; ok {
		query = query.Where("order_id = ?", orderID)
	}
	if fromDate, ok := filters["from_date"]; ok {
		query = query.Where("created_at >= ?", fromDate)
	}
	if toDate, ok := filters["to_date"]; ok {
		query = query.Where("created_at <= ?", toDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&orders).Error
	return orders, total, err
}

// Shipping Tracking
func (r *shippingRepository) CreateShippingTracking(tracking *model.ShippingTracking) error {
	return r.db.Create(tracking).Error
}

func (r *shippingRepository) GetShippingTrackingByOrderID(orderID uint) ([]model.ShippingTracking, error) {
	var tracking []model.ShippingTracking
	err := r.db.Joins("JOIN shipping_orders ON shipping_tracking.shipping_order_id = shipping_orders.id").
		Where("shipping_orders.order_id = ?", orderID).
		Order("shipping_tracking.created_at ASC").
		Find(&tracking).Error
	return tracking, err
}

func (r *shippingRepository) GetShippingTrackingByShippingOrderID(shippingOrderID uint) ([]model.ShippingTracking, error) {
	var tracking []model.ShippingTracking
	err := r.db.Where("shipping_order_id = ?", shippingOrderID).
		Order("created_at ASC").
		Find(&tracking).Error
	return tracking, err
}

// Statistics
func (r *shippingRepository) GetShippingStats() (*model.ShippingStats, error) {
	stats := &model.ShippingStats{}

	// Count orders by status
	r.db.Model(&model.ShippingOrder{}).Count(&stats.TotalOrders)
	r.db.Model(&model.ShippingOrder{}).Where("status = ?", model.ShippingOrderStatusPending).Count(&stats.PendingOrders)
	r.db.Model(&model.ShippingOrder{}).Where("status = ?", model.ShippingOrderStatusInTransit).Count(&stats.ShippedOrders)
	r.db.Model(&model.ShippingOrder{}).Where("status = ?", model.ShippingOrderStatusDelivered).Count(&stats.DeliveredOrders)
	r.db.Model(&model.ShippingOrder{}).Where("status = ?", model.ShippingOrderStatusFailed).Count(&stats.FailedOrders)

	// Calculate revenue
	var totalRevenue float64
	r.db.Model(&model.ShippingOrder{}).Select("COALESCE(SUM(total_fee), 0)").Scan(&totalRevenue)
	stats.TotalRevenue = totalRevenue

	// Calculate average fee
	if stats.TotalOrders > 0 {
		stats.AverageFee = totalRevenue / float64(stats.TotalOrders)
	}

	// Calculate success rate
	successfulOrders := stats.DeliveredOrders
	totalProcessedOrders := stats.PendingOrders + stats.ShippedOrders + stats.DeliveredOrders + stats.FailedOrders
	if totalProcessedOrders > 0 {
		stats.SuccessRate = float64(successfulOrders) / float64(totalProcessedOrders) * 100
	}

	return stats, nil
}

func (r *shippingRepository) GetShippingStatsByProvider(providerID uint) (*model.ShippingStats, error) {
	stats := &model.ShippingStats{}

	// Count orders by status for specific provider
	r.db.Model(&model.ShippingOrder{}).Where("provider_id = ?", providerID).Count(&stats.TotalOrders)
	r.db.Model(&model.ShippingOrder{}).Where("provider_id = ? AND status = ?", providerID, model.ShippingOrderStatusPending).Count(&stats.PendingOrders)
	r.db.Model(&model.ShippingOrder{}).Where("provider_id = ? AND status = ?", providerID, model.ShippingOrderStatusInTransit).Count(&stats.ShippedOrders)
	r.db.Model(&model.ShippingOrder{}).Where("provider_id = ? AND status = ?", providerID, model.ShippingOrderStatusDelivered).Count(&stats.DeliveredOrders)
	r.db.Model(&model.ShippingOrder{}).Where("provider_id = ? AND status = ?", providerID, model.ShippingOrderStatusFailed).Count(&stats.FailedOrders)

	// Calculate revenue for specific provider
	var totalRevenue float64
	r.db.Model(&model.ShippingOrder{}).Where("provider_id = ?", providerID).Select("COALESCE(SUM(total_fee), 0)").Scan(&totalRevenue)
	stats.TotalRevenue = totalRevenue

	// Calculate average fee
	if stats.TotalOrders > 0 {
		stats.AverageFee = totalRevenue / float64(stats.TotalOrders)
	}

	// Calculate success rate
	successfulOrders := stats.DeliveredOrders
	totalProcessedOrders := stats.PendingOrders + stats.ShippedOrders + stats.DeliveredOrders + stats.FailedOrders
	if totalProcessedOrders > 0 {
		stats.SuccessRate = float64(successfulOrders) / float64(totalProcessedOrders) * 100
	}

	return stats, nil
}
