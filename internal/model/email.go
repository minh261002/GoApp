package model

import (
	"time"
)

// EmailTemplate represents an email template
type EmailTemplate struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	Name      string     `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Subject   string     `json:"subject" gorm:"size:255;not null"`
	Body      string     `json:"body" gorm:"type:text;not null"`
	BodyHTML  string     `json:"body_html" gorm:"type:text"`
	Type      string     `json:"type" gorm:"size:50;not null;index"` // order_confirmation, order_shipped, etc.
	Language  string     `json:"language" gorm:"size:10;default:'vi'"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
	Variables string     `json:"variables" gorm:"type:json"` // Available variables for this template
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

// EmailQueue represents an email in the queue
type EmailQueue struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	To           string         `json:"to" gorm:"size:255;not null"`
	CC           string         `json:"cc" gorm:"size:500"`
	BCC          string         `json:"bcc" gorm:"size:500"`
	Subject      string         `json:"subject" gorm:"size:255;not null"`
	Body         string         `json:"body" gorm:"type:text;not null"`
	BodyHTML     string         `json:"body_html" gorm:"type:text"`
	TemplateID   *uint          `json:"template_id"`
	Template     *EmailTemplate `json:"template,omitempty" gorm:"foreignKey:TemplateID"`
	Priority     int            `json:"priority" gorm:"default:0"`               // 0=normal, 1=high, 2=urgent
	Status       string         `json:"status" gorm:"size:20;default:'pending'"` // pending, sending, sent, failed
	Attempts     int            `json:"attempts" gorm:"default:0"`
	MaxAttempts  int            `json:"max_attempts" gorm:"default:3"`
	ErrorMessage string         `json:"error_message" gorm:"type:text"`
	ScheduledAt  *time.Time     `json:"scheduled_at"`
	SentAt       *time.Time     `json:"sent_at"`
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

// EmailLog represents email sending logs
type EmailLog struct {
	ID           uint        `json:"id" gorm:"primaryKey"`
	EmailQueueID uint        `json:"email_queue_id"`
	EmailQueue   *EmailQueue `json:"email_queue,omitempty" gorm:"foreignKey:EmailQueueID"`
	To           string      `json:"to" gorm:"size:255;not null"`
	Subject      string      `json:"subject" gorm:"size:255;not null"`
	Status       string      `json:"status" gorm:"size:20;not null"` // sent, failed, bounced
	ErrorMessage string      `json:"error_message" gorm:"type:text"`
	Provider     string      `json:"provider" gorm:"size:50"`     // smtp, sendgrid, etc.
	ProviderID   string      `json:"provider_id" gorm:"size:100"` // External provider message ID
	SentAt       time.Time   `json:"sent_at"`
	CreatedAt    time.Time   `json:"created_at" gorm:"autoCreateTime"`
}

// EmailConfig represents email configuration
type EmailConfig struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Provider  string    `json:"provider" gorm:"size:50;not null"` // smtp, sendgrid, etc.
	Host      string    `json:"host" gorm:"size:255;not null"`
	Port      int       `json:"port" gorm:"not null"`
	Username  string    `json:"username" gorm:"size:255;not null"`
	Password  string    `json:"password" gorm:"size:255;not null"`
	From      string    `json:"from" gorm:"size:255;not null"`
	FromName  string    `json:"from_name" gorm:"size:255"`
	SSL       bool      `json:"ssl" gorm:"default:true"`
	TLS       bool      `json:"tls" gorm:"default:true"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// EmailRequest represents a request to send an email
type EmailRequest struct {
	To           []string               `json:"to" validate:"required,min=1"`
	CC           []string               `json:"cc,omitempty"`
	BCC          []string               `json:"bcc,omitempty"`
	Subject      string                 `json:"subject" validate:"required"`
	Body         string                 `json:"body,omitempty"`
	BodyHTML     string                 `json:"body_html,omitempty"`
	TemplateName string                 `json:"template_name,omitempty"`
	Variables    map[string]interface{} `json:"variables,omitempty"`
	Priority     int                    `json:"priority,omitempty"`
	ScheduledAt  *time.Time             `json:"scheduled_at,omitempty"`
}

// EmailResponse represents the response after sending an email
type EmailResponse struct {
	ID        uint       `json:"id"`
	Status    string     `json:"status"`
	Message   string     `json:"message"`
	SentAt    *time.Time `json:"sent_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// EmailStats represents email statistics
type EmailStats struct {
	TotalSent    int64      `json:"total_sent"`
	TotalFailed  int64      `json:"total_failed"`
	TotalPending int64      `json:"total_pending"`
	SuccessRate  float64    `json:"success_rate"`
	AverageTime  float64    `json:"average_time"` // in seconds
	LastSentAt   *time.Time `json:"last_sent_at,omitempty"`
}

// EmailTemplateType represents different types of email templates
const (
	EmailTemplateTypeOrderConfirmation = "order_confirmation"
	EmailTemplateTypeOrderShipped      = "order_shipped"
	EmailTemplateTypeOrderDelivered    = "order_delivered"
	EmailTemplateTypeOrderCancelled    = "order_cancelled"
	EmailTemplateTypePaymentReceived   = "payment_received"
	EmailTemplateTypePasswordReset     = "password_reset"
	EmailTemplateTypeWelcome           = "welcome"
	EmailTemplateTypeNewsletter        = "newsletter"
	EmailTemplateTypePromotion         = "promotion"
	EmailTemplateTypeSystemAlert       = "system_alert"
)

// EmailStatus represents different email statuses
const (
	EmailStatusPending = "pending"
	EmailStatusSending = "sending"
	EmailStatusSent    = "sent"
	EmailStatusFailed  = "failed"
	EmailStatusBounced = "bounced"
)

// EmailPriority represents different email priorities
const (
	EmailPriorityLow    = 0
	EmailPriorityNormal = 1
	EmailPriorityHigh   = 2
	EmailPriorityUrgent = 3
)
