package model

import (
	"time"

	"gorm.io/gorm"
)

// OrderTracking represents real-time order tracking
type OrderTracking struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	OrderID uint   `json:"order_id" gorm:"not null;index"`
	Order   *Order `json:"order,omitempty" gorm:"foreignKey:OrderID"`

	// Tracking Information
	TrackingNumber string `json:"tracking_number" gorm:"size:100;not null;index"`
	Carrier        string `json:"carrier" gorm:"size:50;not null"` // GHTK, GHN, ViettelPost, etc.
	CarrierCode    string `json:"carrier_code" gorm:"size:20;not null"`

	// Current Status
	Status      string `json:"status" gorm:"size:50;not null;index"`
	StatusText  string `json:"status_text" gorm:"size:255;not null"`
	Location    string `json:"location" gorm:"size:255"`
	Description string `json:"description" gorm:"type:text"`

	// Estimated Delivery
	EstimatedDelivery *time.Time `json:"estimated_delivery"`
	ActualDelivery    *time.Time `json:"actual_delivery"`

	// Tracking URL
	TrackingURL string `json:"tracking_url" gorm:"size:500"`

	// Last Update
	LastUpdatedAt time.Time `json:"last_updated_at"`
	LastSyncAt    time.Time `json:"last_sync_at"`

	// Settings
	AutoSync   bool `json:"auto_sync" gorm:"default:true"`
	NotifyUser bool `json:"notify_user" gorm:"default:true"`
	IsActive   bool `json:"is_active" gorm:"default:true"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// OrderTrackingEvent represents a tracking event
type OrderTrackingEvent struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	OrderTrackingID uint           `json:"order_tracking_id" gorm:"not null;index"`
	OrderTracking   *OrderTracking `json:"order_tracking,omitempty" gorm:"foreignKey:OrderTrackingID"`

	// Event Information
	Status      string `json:"status" gorm:"size:50;not null"`
	StatusText  string `json:"status_text" gorm:"size:255;not null"`
	Location    string `json:"location" gorm:"size:255"`
	Description string `json:"description" gorm:"type:text"`

	// Event Details
	EventType   string `json:"event_type" gorm:"size:50;not null"` // pickup, transit, delivered, etc.
	EventCode   string `json:"event_code" gorm:"size:20"`
	IsImportant bool   `json:"is_important" gorm:"default:false"`

	// Source
	Source     string `json:"source" gorm:"size:50;not null"` // api, webhook, manual
	SourceData string `json:"source_data" gorm:"type:text"`   // Raw data from source

	// Timestamps
	EventTime time.Time `json:"event_time" gorm:"not null;index"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// OrderTrackingWebhook represents webhook configuration
type OrderTrackingWebhook struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Carrier     string `json:"carrier" gorm:"size:50;not null;index"`
	CarrierCode string `json:"carrier_code" gorm:"size:20;not null"`

	// Webhook Configuration
	URL      string `json:"url" gorm:"size:500;not null"`
	Secret   string `json:"secret" gorm:"size:255"`
	IsActive bool   `json:"is_active" gorm:"default:true"`

	// Events to Track
	Events     string `json:"events" gorm:"type:text"` // JSON array of events
	RetryCount int    `json:"retry_count" gorm:"default:3"`
	Timeout    int    `json:"timeout" gorm:"default:30"` // seconds

	// Statistics
	SuccessCount int `json:"success_count" gorm:"default:0"`
	FailureCount int `json:"failure_count" gorm:"default:0"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// OrderTrackingNotification represents tracking notifications
type OrderTrackingNotification struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	OrderTrackingID uint           `json:"order_tracking_id" gorm:"not null;index"`
	OrderTracking   *OrderTracking `json:"order_tracking,omitempty" gorm:"foreignKey:OrderTrackingID"`

	// Notification Information
	UserID  uint                `json:"user_id" gorm:"not null;index"`
	User    *User               `json:"user,omitempty" gorm:"foreignKey:UserID"`
	EventID uint                `json:"event_id" gorm:"not null;index"`
	Event   *OrderTrackingEvent `json:"event,omitempty" gorm:"foreignKey:EventID"`

	// Notification Details
	Type    string     `json:"type" gorm:"size:50;not null"` // email, sms, push, webhook
	Title   string     `json:"title" gorm:"size:255;not null"`
	Message string     `json:"message" gorm:"type:text;not null"`
	IsSent  bool       `json:"is_sent" gorm:"default:false"`
	SentAt  *time.Time `json:"sent_at"`

	// Retry Information
	RetryCount int `json:"retry_count" gorm:"default:0"`
	MaxRetries int `json:"max_retries" gorm:"default:3"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// ===== REQUEST/RESPONSE MODELS =====

// OrderTrackingCreateRequest represents request to create order tracking
type OrderTrackingCreateRequest struct {
	OrderID           uint       `json:"order_id" binding:"required"`
	TrackingNumber    string     `json:"tracking_number" binding:"required,min=1,max=100"`
	Carrier           string     `json:"carrier" binding:"required,min=1,max=50"`
	CarrierCode       string     `json:"carrier_code" binding:"required,min=1,max=20"`
	EstimatedDelivery *time.Time `json:"estimated_delivery"`
	TrackingURL       string     `json:"tracking_url" binding:"omitempty,url"`
	AutoSync          *bool      `json:"auto_sync"`
	NotifyUser        *bool      `json:"notify_user"`
}

// OrderTrackingUpdateRequest represents request to update order tracking
type OrderTrackingUpdateRequest struct {
	Status            string     `json:"status" binding:"omitempty,min=1,max=50"`
	StatusText        string     `json:"status_text" binding:"omitempty,min=1,max=255"`
	Location          string     `json:"location" binding:"omitempty,max=255"`
	Description       string     `json:"description"`
	EstimatedDelivery *time.Time `json:"estimated_delivery"`
	ActualDelivery    *time.Time `json:"actual_delivery"`
	TrackingURL       string     `json:"tracking_url" binding:"omitempty,url"`
	AutoSync          *bool      `json:"auto_sync"`
	NotifyUser        *bool      `json:"notify_user"`
	IsActive          *bool      `json:"is_active"`
}

// OrderTrackingResponse represents response for order tracking
type OrderTrackingResponse struct {
	ID                uint       `json:"id"`
	OrderID           uint       `json:"order_id"`
	TrackingNumber    string     `json:"tracking_number"`
	Carrier           string     `json:"carrier"`
	CarrierCode       string     `json:"carrier_code"`
	Status            string     `json:"status"`
	StatusText        string     `json:"status_text"`
	Location          string     `json:"location"`
	Description       string     `json:"description"`
	EstimatedDelivery *time.Time `json:"estimated_delivery"`
	ActualDelivery    *time.Time `json:"actual_delivery"`
	TrackingURL       string     `json:"tracking_url"`
	LastUpdatedAt     time.Time  `json:"last_updated_at"`
	LastSyncAt        time.Time  `json:"last_sync_at"`
	AutoSync          bool       `json:"auto_sync"`
	NotifyUser        bool       `json:"notify_user"`
	IsActive          bool       `json:"is_active"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`

	// Relations
	Events []OrderTrackingEventResponse `json:"events,omitempty"`
}

// OrderTrackingEventResponse represents response for tracking event
type OrderTrackingEventResponse struct {
	ID              uint      `json:"id"`
	OrderTrackingID uint      `json:"order_tracking_id"`
	Status          string    `json:"status"`
	StatusText      string    `json:"status_text"`
	Location        string    `json:"location"`
	Description     string    `json:"description"`
	EventType       string    `json:"event_type"`
	EventCode       string    `json:"event_code"`
	IsImportant     bool      `json:"is_important"`
	Source          string    `json:"source"`
	EventTime       time.Time `json:"event_time"`
	CreatedAt       time.Time `json:"created_at"`
}

// OrderTrackingWebhookRequest represents webhook request
type OrderTrackingWebhookRequest struct {
	OrderID        uint   `json:"order_id" binding:"required"`
	TrackingNumber string `json:"tracking_number" binding:"required"`
	Status         string `json:"status" binding:"required"`
	StatusText     string `json:"status_text" binding:"required"`
	Location       string `json:"location"`
	Description    string `json:"description"`
	EventTime      string `json:"event_time" binding:"required"`
	Source         string `json:"source" binding:"required"`
}

// OrderTrackingSyncRequest represents sync request
type OrderTrackingSyncRequest struct {
	OrderTrackingIDs []uint `json:"order_tracking_ids" binding:"required,min=1"`
	ForceSync        bool   `json:"force_sync"`
}

// OrderTrackingStatsResponse represents tracking statistics
type OrderTrackingStatsResponse struct {
	TotalTrackings      int64   `json:"total_trackings"`
	ActiveTrackings     int64   `json:"active_trackings"`
	DeliveredOrders     int64   `json:"delivered_orders"`
	InTransitOrders     int64   `json:"in_transit_orders"`
	PendingOrders       int64   `json:"pending_orders"`
	FailedDeliveries    int64   `json:"failed_deliveries"`
	AverageDeliveryTime float64 `json:"average_delivery_time"` // in hours
}

// ===== CONSTANTS =====

// Tracking Status Constants
const (
	TrackingStatusPending        = "pending"
	TrackingStatusPickedUp       = "picked_up"
	TrackingStatusInTransit      = "in_transit"
	TrackingStatusOutForDelivery = "out_for_delivery"
	TrackingStatusDelivered      = "delivered"
	TrackingStatusFailed         = "failed"
	TrackingStatusReturned       = "returned"
	TrackingStatusCancelled      = "cancelled"
)

// Event Type Constants
const (
	EventTypePickup         = "pickup"
	EventTypeTransit        = "transit"
	EventTypeOutForDelivery = "out_for_delivery"
	EventTypeDelivered      = "delivered"
	EventTypeFailed         = "failed"
	EventTypeReturned       = "returned"
	EventTypeCancelled      = "cancelled"
)

// Notification Type Constants
const (
	NotificationTypeEmail   = "email"
	NotificationTypeSMS     = "sms"
	NotificationTypePush    = "push"
	NotificationTypeWebhook = "webhook"
)

// Source Constants
const (
	SourceAPI     = "api"
	SourceWebhook = "webhook"
	SourceManual  = "manual"
	SourceSync    = "sync"
)
