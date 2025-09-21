package repository

import (
	"fmt"
	"go_app/internal/model"
	"go_app/pkg/database"
	"time"

	"gorm.io/gorm"
)

type OrderTrackingRepository struct {
	db *gorm.DB
}

func NewOrderTrackingRepository() *OrderTrackingRepository {
	return &OrderTrackingRepository{
		db: database.GetDB(),
	}
}

// ===== ORDER TRACKING REPOSITORY =====

// CreateOrderTracking creates a new order tracking
func (r *OrderTrackingRepository) CreateOrderTracking(tracking *model.OrderTracking) error {
	if err := r.db.Create(tracking).Error; err != nil {
		return fmt.Errorf("failed to create order tracking: %w", err)
	}
	return nil
}

// GetOrderTrackingByID gets order tracking by ID
func (r *OrderTrackingRepository) GetOrderTrackingByID(id uint) (*model.OrderTracking, error) {
	var tracking model.OrderTracking
	if err := r.db.Preload("Order").Where("id = ?", id).First(&tracking).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("order tracking not found")
		}
		return nil, fmt.Errorf("failed to get order tracking: %w", err)
	}
	return &tracking, nil
}

// GetOrderTrackingByOrderID gets order tracking by order ID
func (r *OrderTrackingRepository) GetOrderTrackingByOrderID(orderID uint) (*model.OrderTracking, error) {
	var tracking model.OrderTracking
	if err := r.db.Preload("Order").Where("order_id = ? AND is_active = ?", orderID, true).First(&tracking).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("order tracking not found")
		}
		return nil, fmt.Errorf("failed to get order tracking: %w", err)
	}
	return &tracking, nil
}

// GetOrderTrackingByTrackingNumber gets order tracking by tracking number
func (r *OrderTrackingRepository) GetOrderTrackingByTrackingNumber(trackingNumber string) (*model.OrderTracking, error) {
	var tracking model.OrderTracking
	if err := r.db.Preload("Order").Where("tracking_number = ? AND is_active = ?", trackingNumber, true).First(&tracking).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("order tracking not found")
		}
		return nil, fmt.Errorf("failed to get order tracking: %w", err)
	}
	return &tracking, nil
}

// GetAllOrderTrackings gets all order trackings with pagination
func (r *OrderTrackingRepository) GetAllOrderTrackings(page, limit int, filters map[string]interface{}) ([]model.OrderTracking, int64, error) {
	var trackings []model.OrderTracking
	var total int64

	query := r.db.Model(&model.OrderTracking{})

	// Apply filters
	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}
	if carrier, ok := filters["carrier"]; ok {
		query = query.Where("carrier = ?", carrier)
	}
	if isActive, ok := filters["is_active"]; ok {
		query = query.Where("is_active = ?", isActive)
	}
	if autoSync, ok := filters["auto_sync"]; ok {
		query = query.Where("auto_sync = ?", autoSync)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count order trackings: %w", err)
	}

	// Get trackings with pagination
	offset := (page - 1) * limit
	if err := query.Preload("Order").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&trackings).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get order trackings: %w", err)
	}

	return trackings, total, nil
}

// UpdateOrderTracking updates order tracking
func (r *OrderTrackingRepository) UpdateOrderTracking(tracking *model.OrderTracking) error {
	if err := r.db.Save(tracking).Error; err != nil {
		return fmt.Errorf("failed to update order tracking: %w", err)
	}
	return nil
}

// DeleteOrderTracking deletes order tracking
func (r *OrderTrackingRepository) DeleteOrderTracking(id uint) error {
	if err := r.db.Delete(&model.OrderTracking{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete order tracking: %w", err)
	}
	return nil
}

// GetOrderTrackingsForSync gets trackings that need to be synced
func (r *OrderTrackingRepository) GetOrderTrackingsForSync(limit int) ([]model.OrderTracking, error) {
	var trackings []model.OrderTracking

	// Get trackings that are active, have auto_sync enabled, and haven't been synced recently
	cutoffTime := time.Now().Add(-1 * time.Hour) // Sync at least once per hour

	if err := r.db.Preload("Order").
		Where("is_active = ? AND auto_sync = ? AND (last_sync_at < ? OR last_sync_at IS NULL)", true, true, cutoffTime).
		Order("last_sync_at ASC").
		Limit(limit).
		Find(&trackings).Error; err != nil {
		return nil, fmt.Errorf("failed to get trackings for sync: %w", err)
	}

	return trackings, nil
}

// ===== ORDER TRACKING EVENTS REPOSITORY =====

// CreateOrderTrackingEvent creates a new tracking event
func (r *OrderTrackingRepository) CreateOrderTrackingEvent(event *model.OrderTrackingEvent) error {
	if err := r.db.Create(event).Error; err != nil {
		return fmt.Errorf("failed to create tracking event: %w", err)
	}
	return nil
}

// GetOrderTrackingEvents gets events for a tracking
func (r *OrderTrackingRepository) GetOrderTrackingEvents(trackingID uint, page, limit int) ([]model.OrderTrackingEvent, int64, error) {
	var events []model.OrderTrackingEvent
	var total int64

	// Count total
	if err := r.db.Model(&model.OrderTrackingEvent{}).Where("order_tracking_id = ?", trackingID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count tracking events: %w", err)
	}

	// Get events with pagination
	offset := (page - 1) * limit
	if err := r.db.Where("order_tracking_id = ?", trackingID).
		Order("event_time DESC").
		Offset(offset).Limit(limit).
		Find(&events).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get tracking events: %w", err)
	}

	return events, total, nil
}

// GetLatestOrderTrackingEvent gets the latest event for a tracking
func (r *OrderTrackingRepository) GetLatestOrderTrackingEvent(trackingID uint) (*model.OrderTrackingEvent, error) {
	var event model.OrderTrackingEvent
	if err := r.db.Where("order_tracking_id = ?", trackingID).
		Order("event_time DESC").
		First(&event).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no tracking events found")
		}
		return nil, fmt.Errorf("failed to get latest tracking event: %w", err)
	}
	return &event, nil
}

// ===== ORDER TRACKING WEBHOOKS REPOSITORY =====

// CreateOrderTrackingWebhook creates a new webhook configuration
func (r *OrderTrackingRepository) CreateOrderTrackingWebhook(webhook *model.OrderTrackingWebhook) error {
	if err := r.db.Create(webhook).Error; err != nil {
		return fmt.Errorf("failed to create tracking webhook: %w", err)
	}
	return nil
}

// GetOrderTrackingWebhookByCarrier gets webhook by carrier
func (r *OrderTrackingRepository) GetOrderTrackingWebhookByCarrier(carrier, carrierCode string) (*model.OrderTrackingWebhook, error) {
	var webhook model.OrderTrackingWebhook
	if err := r.db.Where("carrier = ? AND carrier_code = ? AND is_active = ?", carrier, carrierCode, true).First(&webhook).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("webhook not found")
		}
		return nil, fmt.Errorf("failed to get tracking webhook: %w", err)
	}
	return &webhook, nil
}

// GetAllOrderTrackingWebhooks gets all webhook configurations
func (r *OrderTrackingRepository) GetAllOrderTrackingWebhooks() ([]model.OrderTrackingWebhook, error) {
	var webhooks []model.OrderTrackingWebhook
	if err := r.db.Where("is_active = ?", true).Find(&webhooks).Error; err != nil {
		return nil, fmt.Errorf("failed to get tracking webhooks: %w", err)
	}
	return webhooks, nil
}

// UpdateOrderTrackingWebhook updates webhook configuration
func (r *OrderTrackingRepository) UpdateOrderTrackingWebhook(webhook *model.OrderTrackingWebhook) error {
	if err := r.db.Save(webhook).Error; err != nil {
		return fmt.Errorf("failed to update tracking webhook: %w", err)
	}
	return nil
}

// ===== ORDER TRACKING NOTIFICATIONS REPOSITORY =====

// CreateOrderTrackingNotification creates a new notification
func (r *OrderTrackingRepository) CreateOrderTrackingNotification(notification *model.OrderTrackingNotification) error {
	if err := r.db.Create(notification).Error; err != nil {
		return fmt.Errorf("failed to create tracking notification: %w", err)
	}
	return nil
}

// GetOrderTrackingNotifications gets notifications for a user
func (r *OrderTrackingRepository) GetOrderTrackingNotifications(userID uint, page, limit int) ([]model.OrderTrackingNotification, int64, error) {
	var notifications []model.OrderTrackingNotification
	var total int64

	// Count total
	if err := r.db.Model(&model.OrderTrackingNotification{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count tracking notifications: %w", err)
	}

	// Get notifications with pagination
	offset := (page - 1) * limit
	if err := r.db.Preload("OrderTracking").Preload("Event").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&notifications).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get tracking notifications: %w", err)
	}

	return notifications, total, nil
}

// GetPendingOrderTrackingNotifications gets pending notifications
func (r *OrderTrackingRepository) GetPendingOrderTrackingNotifications(limit int) ([]model.OrderTrackingNotification, error) {
	var notifications []model.OrderTrackingNotification

	if err := r.db.Preload("OrderTracking").Preload("Event").
		Where("is_sent = ? AND retry_count < max_retries", false).
		Order("created_at ASC").
		Limit(limit).
		Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("failed to get pending notifications: %w", err)
	}

	return notifications, nil
}

// UpdateOrderTrackingNotification updates notification
func (r *OrderTrackingRepository) UpdateOrderTrackingNotification(notification *model.OrderTrackingNotification) error {
	if err := r.db.Save(notification).Error; err != nil {
		return fmt.Errorf("failed to update tracking notification: %w", err)
	}
	return nil
}

// ===== STATISTICS =====

// GetOrderTrackingStats gets tracking statistics
func (r *OrderTrackingRepository) GetOrderTrackingStats() (*model.OrderTrackingStatsResponse, error) {
	stats := &model.OrderTrackingStatsResponse{}

	// Total trackings
	if err := r.db.Model(&model.OrderTracking{}).Count(&stats.TotalTrackings).Error; err != nil {
		return nil, fmt.Errorf("failed to count total trackings: %w", err)
	}

	// Active trackings
	if err := r.db.Model(&model.OrderTracking{}).Where("is_active = ?", true).Count(&stats.ActiveTrackings).Error; err != nil {
		return nil, fmt.Errorf("failed to count active trackings: %w", err)
	}

	// Delivered orders
	if err := r.db.Model(&model.OrderTracking{}).Where("status = ?", model.TrackingStatusDelivered).Count(&stats.DeliveredOrders).Error; err != nil {
		return nil, fmt.Errorf("failed to count delivered orders: %w", err)
	}

	// In transit orders
	if err := r.db.Model(&model.OrderTracking{}).Where("status = ?", model.TrackingStatusInTransit).Count(&stats.InTransitOrders).Error; err != nil {
		return nil, fmt.Errorf("failed to count in transit orders: %w", err)
	}

	// Pending orders
	if err := r.db.Model(&model.OrderTracking{}).Where("status = ?", model.TrackingStatusPending).Count(&stats.PendingOrders).Error; err != nil {
		return nil, fmt.Errorf("failed to count pending orders: %w", err)
	}

	// Failed deliveries
	if err := r.db.Model(&model.OrderTracking{}).Where("status = ?", model.TrackingStatusFailed).Count(&stats.FailedDeliveries).Error; err != nil {
		return nil, fmt.Errorf("failed to count failed deliveries: %w", err)
	}

	// Calculate average delivery time
	var avgDeliveryTime float64
	if err := r.db.Raw(`
		SELECT AVG(TIMESTAMPDIFF(HOUR, created_at, actual_delivery)) 
		FROM order_trackings 
		WHERE status = ? AND actual_delivery IS NOT NULL
	`, model.TrackingStatusDelivered).Scan(&avgDeliveryTime).Error; err != nil {
		// If no delivered orders, set to 0
		avgDeliveryTime = 0
	}
	stats.AverageDeliveryTime = avgDeliveryTime

	return stats, nil
}
