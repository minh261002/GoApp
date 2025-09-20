package model

import (
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeOrder     NotificationType = "order"
	NotificationTypePayment   NotificationType = "payment"
	NotificationTypeShipping  NotificationType = "shipping"
	NotificationTypeProduct   NotificationType = "product"
	NotificationTypePromotion NotificationType = "promotion"
	NotificationTypeSystem    NotificationType = "system"
	NotificationTypeSecurity  NotificationType = "security"
	NotificationTypeReview    NotificationType = "review"
	NotificationTypeWishlist  NotificationType = "wishlist"
	NotificationTypeInventory NotificationType = "inventory"
	NotificationTypeCoupon    NotificationType = "coupon"
	NotificationTypePoint     NotificationType = "point"
	NotificationTypeGeneral   NotificationType = "general"
)

// NotificationPriority represents the priority of notification
type NotificationPriority string

const (
	NotificationPriorityLow    NotificationPriority = "low"
	NotificationPriorityNormal NotificationPriority = "normal"
	NotificationPriorityHigh   NotificationPriority = "high"
	NotificationPriorityUrgent NotificationPriority = "urgent"
)

// NotificationStatus represents the status of notification
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusDelivered NotificationStatus = "delivered"
	NotificationStatusRead      NotificationStatus = "read"
	NotificationStatusFailed    NotificationStatus = "failed"
	NotificationStatusCancelled NotificationStatus = "cancelled"
)

// NotificationChannel represents the delivery channel
type NotificationChannel string

const (
	NotificationChannelEmail   NotificationChannel = "email"
	NotificationChannelSMS     NotificationChannel = "sms"
	NotificationChannelPush    NotificationChannel = "push"
	NotificationChannelInApp   NotificationChannel = "in_app"
	NotificationChannelWebhook NotificationChannel = "webhook"
)

// Notification represents a notification
type Notification struct {
	ID          uint                 `json:"id" gorm:"primaryKey"`
	UserID      *uint                `json:"user_id" gorm:"index"`
	Type        NotificationType     `json:"type" gorm:"size:50;not null;index"`
	Priority    NotificationPriority `json:"priority" gorm:"size:20;not null;default:'normal'"`
	Status      NotificationStatus   `json:"status" gorm:"size:20;not null;default:'pending'"`
	Channel     NotificationChannel  `json:"channel" gorm:"size:20;not null"`
	Title       string               `json:"title" gorm:"size:255;not null"`
	Message     string               `json:"message" gorm:"type:text;not null"`
	Data        string               `json:"data" gorm:"type:json"` // Additional data for the notification
	ActionURL   string               `json:"action_url" gorm:"size:500"`
	ImageURL    string               `json:"image_url" gorm:"size:500"`
	IsRead      bool                 `json:"is_read" gorm:"default:false"`
	IsArchived  bool                 `json:"is_archived" gorm:"default:false"`
	ReadAt      *time.Time           `json:"read_at"`
	SentAt      *time.Time           `json:"sent_at"`
	DeliveredAt *time.Time           `json:"delivered_at"`
	FailedAt    *time.Time           `json:"failed_at"`
	RetryCount  int                  `json:"retry_count" gorm:"default:0"`
	ErrorMsg    string               `json:"error_msg" gorm:"type:text"`
	ExpiresAt   *time.Time           `json:"expires_at"`
	ScheduledAt *time.Time           `json:"scheduled_at"` // For scheduled notifications

	// Relationships
	User *User `json:"user" gorm:"foreignKey:UserID"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

// NotificationTemplate represents a notification template
type NotificationTemplate struct {
	ID          uint                `json:"id" gorm:"primaryKey"`
	Name        string              `json:"name" gorm:"size:100;not null;unique"`
	Type        NotificationType    `json:"type" gorm:"size:50;not null"`
	Channel     NotificationChannel `json:"channel" gorm:"size:20;not null"`
	Subject     string              `json:"subject" gorm:"size:255;not null"`
	Body        string              `json:"body" gorm:"type:text;not null"`
	Variables   string              `json:"variables" gorm:"type:json"` // Available variables for template
	IsActive    bool                `json:"is_active" gorm:"default:true"`
	IsSystem    bool                `json:"is_system" gorm:"default:false"`
	Description string              `json:"description" gorm:"type:text"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

// NotificationPreference represents user notification preferences
type NotificationPreference struct {
	ID         uint                `json:"id" gorm:"primaryKey"`
	UserID     uint                `json:"user_id" gorm:"not null;index"`
	Type       NotificationType    `json:"type" gorm:"size:50;not null"`
	Channel    NotificationChannel `json:"channel" gorm:"size:20;not null"`
	IsEnabled  bool                `json:"is_enabled" gorm:"default:true"`
	Frequency  string              `json:"frequency" gorm:"size:20;default:'immediate'"` // immediate, daily, weekly, monthly
	QuietHours string              `json:"quiet_hours" gorm:"size:20"`                   // e.g., "22:00-08:00"
	Timezone   string              `json:"timezone" gorm:"size:50;default:'UTC'"`

	// Relationships
	User *User `json:"user" gorm:"foreignKey:UserID"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`

	// Unique constraint
	// UNIQUE KEY unique_user_type_channel (user_id, type, channel)
}

// NotificationLog represents notification delivery logs
type NotificationLog struct {
	ID             uint                `json:"id" gorm:"primaryKey"`
	NotificationID uint                `json:"notification_id" gorm:"not null;index"`
	Channel        NotificationChannel `json:"channel" gorm:"size:20;not null"`
	Status         NotificationStatus  `json:"status" gorm:"size:20;not null"`
	Provider       string              `json:"provider" gorm:"size:100"`    // Email provider, SMS provider, etc.
	ProviderID     string              `json:"provider_id" gorm:"size:255"` // External provider message ID
	Response       string              `json:"response" gorm:"type:text"`   // Provider response
	ErrorMsg       string              `json:"error_msg" gorm:"type:text"`
	AttemptedAt    time.Time           `json:"attempted_at"`
	DeliveredAt    *time.Time          `json:"delivered_at"`

	// Relationships
	Notification *Notification `json:"notification" gorm:"foreignKey:NotificationID"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

// NotificationQueue represents queued notifications for processing
type NotificationQueue struct {
	ID             uint                 `json:"id" gorm:"primaryKey"`
	NotificationID uint                 `json:"notification_id" gorm:"not null;index"`
	Priority       NotificationPriority `json:"priority" gorm:"size:20;not null;default:'normal'"`
	Channel        NotificationChannel  `json:"channel" gorm:"size:20;not null"`
	ScheduledAt    time.Time            `json:"scheduled_at" gorm:"not null;index"`
	ProcessedAt    *time.Time           `json:"processed_at"`
	RetryCount     int                  `json:"retry_count" gorm:"default:0"`
	MaxRetries     int                  `json:"max_retries" gorm:"default:3"`
	Status         string               `json:"status" gorm:"size:20;default:'pending'"` // pending, processing, completed, failed

	// Relationships
	Notification *Notification `json:"notification" gorm:"foreignKey:NotificationID"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

// NotificationStats represents notification statistics
type NotificationStats struct {
	ID                  uint                `json:"id" gorm:"primaryKey"`
	Date                time.Time           `json:"date" gorm:"not null;index"`
	Type                NotificationType    `json:"type" gorm:"size:50;not null"`
	Channel             NotificationChannel `json:"channel" gorm:"size:20;not null"`
	TotalSent           int64               `json:"total_sent" gorm:"default:0"`
	TotalDelivered      int64               `json:"total_delivered" gorm:"default:0"`
	TotalRead           int64               `json:"total_read" gorm:"default:0"`
	TotalFailed         int64               `json:"total_failed" gorm:"default:0"`
	DeliveryRate        float64             `json:"delivery_rate" gorm:"type:decimal(5,2);default:0"`
	ReadRate            float64             `json:"read_rate" gorm:"type:decimal(5,2);default:0"`
	AverageDeliveryTime int64               `json:"average_delivery_time" gorm:"default:0"` // in seconds

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`

	// Unique constraint
	// UNIQUE KEY unique_date_type_channel (date, type, channel)
}

// Request/Response DTOs

// CreateNotificationRequest represents a request to create a notification
type CreateNotificationRequest struct {
	UserID      *uint                  `json:"user_id" binding:"omitempty"`
	Type        NotificationType       `json:"type" binding:"required,oneof=order payment shipping product promotion system security review wishlist inventory coupon point general"`
	Priority    NotificationPriority   `json:"priority" binding:"omitempty,oneof=low normal high urgent"`
	Channel     NotificationChannel    `json:"channel" binding:"required,oneof=email sms push in_app webhook"`
	Title       string                 `json:"title" binding:"required,min=1,max=255"`
	Message     string                 `json:"message" binding:"required,min=1"`
	Data        map[string]interface{} `json:"data" binding:"omitempty"`
	ActionURL   string                 `json:"action_url" binding:"omitempty,max=500"`
	ImageURL    string                 `json:"image_url" binding:"omitempty,max=500"`
	ExpiresAt   *time.Time             `json:"expires_at" binding:"omitempty"`
	ScheduledAt *time.Time             `json:"scheduled_at" binding:"omitempty"`
}

// UpdateNotificationRequest represents a request to update a notification
type UpdateNotificationRequest struct {
	Status     *NotificationStatus `json:"status" binding:"omitempty,oneof=pending sent delivered read failed cancelled"`
	IsRead     *bool               `json:"is_read" binding:"omitempty"`
	IsArchived *bool               `json:"is_archived" binding:"omitempty"`
	RetryCount *int                `json:"retry_count" binding:"omitempty,min=0"`
	ErrorMsg   *string             `json:"error_msg" binding:"omitempty"`
}

// NotificationListRequest represents a request to list notifications
type NotificationListRequest struct {
	UserID     *uint                 `json:"user_id" binding:"omitempty"`
	Type       *NotificationType     `json:"type" binding:"omitempty"`
	Status     *NotificationStatus   `json:"status" binding:"omitempty"`
	Channel    *NotificationChannel  `json:"channel" binding:"omitempty"`
	Priority   *NotificationPriority `json:"priority" binding:"omitempty"`
	IsRead     *bool                 `json:"is_read" binding:"omitempty"`
	IsArchived *bool                 `json:"is_archived" binding:"omitempty"`
	StartDate  *time.Time            `json:"start_date" binding:"omitempty"`
	EndDate    *time.Time            `json:"end_date" binding:"omitempty"`
	Page       int                   `json:"page" binding:"omitempty,min=1"`
	Limit      int                   `json:"limit" binding:"omitempty,min=1,max=100"`
	SortBy     string                `json:"sort_by" binding:"omitempty,oneof=created_at updated_at priority scheduled_at"`
	SortOrder  string                `json:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// NotificationStatsRequest represents a request for notification statistics
type NotificationStatsRequest struct {
	StartDate *time.Time           `json:"start_date" binding:"omitempty"`
	EndDate   *time.Time           `json:"end_date" binding:"omitempty"`
	Type      *NotificationType    `json:"type" binding:"omitempty"`
	Channel   *NotificationChannel `json:"channel" binding:"omitempty"`
	GroupBy   string               `json:"group_by" binding:"omitempty,oneof=day week month"`
}

// NotificationPreferenceRequest represents a request to update notification preferences
type NotificationPreferenceRequest struct {
	Type       NotificationType    `json:"type" binding:"required"`
	Channel    NotificationChannel `json:"channel" binding:"required"`
	IsEnabled  bool                `json:"is_enabled"`
	Frequency  string              `json:"frequency" binding:"omitempty,oneof=immediate daily weekly monthly"`
	QuietHours string              `json:"quiet_hours" binding:"omitempty"`
	Timezone   string              `json:"timezone" binding:"omitempty"`
}

// Response DTOs

// NotificationResponse represents a notification response
type NotificationResponse struct {
	ID          uint                   `json:"id"`
	UserID      *uint                  `json:"user_id"`
	Type        NotificationType       `json:"type"`
	Priority    NotificationPriority   `json:"priority"`
	Status      NotificationStatus     `json:"status"`
	Channel     NotificationChannel    `json:"channel"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data"`
	ActionURL   string                 `json:"action_url"`
	ImageURL    string                 `json:"image_url"`
	IsRead      bool                   `json:"is_read"`
	IsArchived  bool                   `json:"is_archived"`
	ReadAt      *time.Time             `json:"read_at"`
	SentAt      *time.Time             `json:"sent_at"`
	DeliveredAt *time.Time             `json:"delivered_at"`
	FailedAt    *time.Time             `json:"failed_at"`
	RetryCount  int                    `json:"retry_count"`
	ErrorMsg    string                 `json:"error_msg"`
	ExpiresAt   *time.Time             `json:"expires_at"`
	ScheduledAt *time.Time             `json:"scheduled_at"`
	User        *User                  `json:"user,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// NotificationTemplateResponse represents a notification template response
type NotificationTemplateResponse struct {
	ID          uint                `json:"id"`
	Name        string              `json:"name"`
	Type        NotificationType    `json:"type"`
	Channel     NotificationChannel `json:"channel"`
	Subject     string              `json:"subject"`
	Body        string              `json:"body"`
	Variables   []string            `json:"variables"`
	IsActive    bool                `json:"is_active"`
	IsSystem    bool                `json:"is_system"`
	Description string              `json:"description"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// NotificationPreferenceResponse represents a notification preference response
type NotificationPreferenceResponse struct {
	ID         uint                `json:"id"`
	UserID     uint                `json:"user_id"`
	Type       NotificationType    `json:"type"`
	Channel    NotificationChannel `json:"channel"`
	IsEnabled  bool                `json:"is_enabled"`
	Frequency  string              `json:"frequency"`
	QuietHours string              `json:"quiet_hours"`
	Timezone   string              `json:"timezone"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
}

// NotificationStatsResponse represents notification statistics response
type NotificationStatsResponse struct {
	TotalSent           int64                      `json:"total_sent"`
	TotalDelivered      int64                      `json:"total_delivered"`
	TotalRead           int64                      `json:"total_read"`
	TotalFailed         int64                      `json:"total_failed"`
	DeliveryRate        float64                    `json:"delivery_rate"`
	ReadRate            float64                    `json:"read_rate"`
	AverageDeliveryTime int64                      `json:"average_delivery_time"`
	DailyStats          []DailyNotificationStats   `json:"daily_stats"`
	TypeStats           []TypeNotificationStats    `json:"type_stats"`
	ChannelStats        []ChannelNotificationStats `json:"channel_stats"`
}

// DailyNotificationStats represents daily notification statistics
type DailyNotificationStats struct {
	Date           time.Time `json:"date"`
	TotalSent      int64     `json:"total_sent"`
	TotalDelivered int64     `json:"total_delivered"`
	TotalRead      int64     `json:"total_read"`
	TotalFailed    int64     `json:"total_failed"`
	DeliveryRate   float64   `json:"delivery_rate"`
	ReadRate       float64   `json:"read_rate"`
}

// TypeNotificationStats represents notification statistics by type
type TypeNotificationStats struct {
	Type           NotificationType `json:"type"`
	TotalSent      int64            `json:"total_sent"`
	TotalDelivered int64            `json:"total_delivered"`
	TotalRead      int64            `json:"total_read"`
	TotalFailed    int64            `json:"total_failed"`
	DeliveryRate   float64          `json:"delivery_rate"`
	ReadRate       float64          `json:"read_rate"`
}

// ChannelNotificationStats represents notification statistics by channel
type ChannelNotificationStats struct {
	Channel        NotificationChannel `json:"channel"`
	TotalSent      int64               `json:"total_sent"`
	TotalDelivered int64               `json:"total_delivered"`
	TotalRead      int64               `json:"total_read"`
	TotalFailed    int64               `json:"total_failed"`
	DeliveryRate   float64             `json:"delivery_rate"`
	ReadRate       float64             `json:"read_rate"`
}

// Methods

// ToResponse converts Notification to NotificationResponse
func (n *Notification) ToResponse() *NotificationResponse {
	var data map[string]interface{}
	if n.Data != "" {
		// TODO: Parse JSON string to map
		data = make(map[string]interface{})
	}

	return &NotificationResponse{
		ID:          n.ID,
		UserID:      n.UserID,
		Type:        n.Type,
		Priority:    n.Priority,
		Status:      n.Status,
		Channel:     n.Channel,
		Title:       n.Title,
		Message:     n.Message,
		Data:        data,
		ActionURL:   n.ActionURL,
		ImageURL:    n.ImageURL,
		IsRead:      n.IsRead,
		IsArchived:  n.IsArchived,
		ReadAt:      n.ReadAt,
		SentAt:      n.SentAt,
		DeliveredAt: n.DeliveredAt,
		FailedAt:    n.FailedAt,
		RetryCount:  n.RetryCount,
		ErrorMsg:    n.ErrorMsg,
		ExpiresAt:   n.ExpiresAt,
		ScheduledAt: n.ScheduledAt,
		User:        n.User,
		CreatedAt:   n.CreatedAt,
		UpdatedAt:   n.UpdatedAt,
	}
}

// ToResponse converts NotificationTemplate to NotificationTemplateResponse
func (nt *NotificationTemplate) ToResponse() *NotificationTemplateResponse {
	var variables []string
	if nt.Variables != "" {
		// TODO: Parse JSON string to []string
		variables = []string{}
	}

	return &NotificationTemplateResponse{
		ID:          nt.ID,
		Name:        nt.Name,
		Type:        nt.Type,
		Channel:     nt.Channel,
		Subject:     nt.Subject,
		Body:        nt.Body,
		Variables:   variables,
		IsActive:    nt.IsActive,
		IsSystem:    nt.IsSystem,
		Description: nt.Description,
		CreatedAt:   nt.CreatedAt,
		UpdatedAt:   nt.UpdatedAt,
	}
}

// ToResponse converts NotificationPreference to NotificationPreferenceResponse
func (np *NotificationPreference) ToResponse() *NotificationPreferenceResponse {
	return &NotificationPreferenceResponse{
		ID:         np.ID,
		UserID:     np.UserID,
		Type:       np.Type,
		Channel:    np.Channel,
		IsEnabled:  np.IsEnabled,
		Frequency:  np.Frequency,
		QuietHours: np.QuietHours,
		Timezone:   np.Timezone,
		CreatedAt:  np.CreatedAt,
		UpdatedAt:  np.UpdatedAt,
	}
}
