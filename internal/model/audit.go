package model

import (
	"time"

	"gorm.io/gorm"
)

// AuditLog represents a comprehensive audit log entry
type AuditLog struct {
	ID     uint  `json:"id" gorm:"primaryKey"`
	UserID *uint `json:"user_id" gorm:"index"`
	User   *User `json:"user,omitempty" gorm:"foreignKey:UserID"`

	// Action details
	Action       string `json:"action" gorm:"size:100;not null;index"`   // create, read, update, delete, login, logout, etc.
	Resource     string `json:"resource" gorm:"size:100;not null;index"` // user, order, product, etc.
	ResourceID   *uint  `json:"resource_id" gorm:"index"`                // ID of the affected resource
	ResourceName string `json:"resource_name" gorm:"size:255"`           // Name/title of the resource

	// Operation details
	Operation string `json:"operation" gorm:"size:50;not null"`    // CRUD operation
	Status    string `json:"status" gorm:"size:20;not null;index"` // success, failure, error
	Message   string `json:"message" gorm:"type:text"`             // Human-readable message

	// Request context
	IPAddress string `json:"ip_address" gorm:"size:45;index"`
	UserAgent string `json:"user_agent" gorm:"type:text"`
	Referer   string `json:"referer" gorm:"size:500"`
	SessionID string `json:"session_id" gorm:"size:255;index"`

	// Data changes
	OldValues string `json:"old_values" gorm:"type:json"` // Previous values (for updates)
	NewValues string `json:"new_values" gorm:"type:json"` // New values (for creates/updates)
	Changes   string `json:"changes" gorm:"type:json"`    // Diff of changes

	// Additional context
	Metadata string `json:"metadata" gorm:"type:json"`              // Additional context data
	Tags     string `json:"tags" gorm:"size:500"`                   // Comma-separated tags
	Severity string `json:"severity" gorm:"size:20;default:'info'"` // info, warning, error, critical

	// Target user (for actions affecting other users)
	TargetUserID *uint `json:"target_user_id" gorm:"index"`
	TargetUser   *User `json:"target_user,omitempty" gorm:"foreignKey:TargetUserID"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// AuditLogConfig represents audit log configuration
type AuditLogConfig struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"size:255;not null;uniqueIndex"`
	Description string `json:"description" gorm:"type:text"`

	// Configuration
	IsEnabled    bool   `json:"is_enabled" gorm:"default:true"`
	LogLevel     string `json:"log_level" gorm:"size:20;default:'info'"` // info, warning, error, critical
	Resources    string `json:"resources" gorm:"type:json"`              // Resources to audit
	Actions      string `json:"actions" gorm:"type:json"`                // Actions to audit
	ExcludeUsers string `json:"exclude_users" gorm:"type:json"`          // Users to exclude from auditing

	// Retention settings
	RetentionDays int `json:"retention_days" gorm:"default:90"`    // Days to keep logs
	MaxLogSize    int `json:"max_log_size" gorm:"default:1000000"` // Maximum number of logs

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// AuditLogSummary represents audit log summary statistics
type AuditLogSummary struct {
	ID       uint      `json:"id" gorm:"primaryKey"`
	Date     time.Time `json:"date" gorm:"type:date;not null;index"`
	Resource string    `json:"resource" gorm:"size:100;not null;index"`
	Action   string    `json:"action" gorm:"size:100;not null;index"`
	Status   string    `json:"status" gorm:"size:20;not null;index"`

	// Statistics
	TotalCount   int64 `json:"total_count" gorm:"default:0"`
	SuccessCount int64 `json:"success_count" gorm:"default:0"`
	FailureCount int64 `json:"failure_count" gorm:"default:0"`
	ErrorCount   int64 `json:"error_count" gorm:"default:0"`

	// Unique users
	UniqueUsers int64 `json:"unique_users" gorm:"default:0"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// Audit Constants
const (
	// Actions
	ActionCreate           = "create"
	ActionRead             = "read"
	ActionUpdate           = "update"
	ActionDelete           = "delete"
	ActionLogin            = "login"
	ActionLogout           = "logout"
	ActionRegister         = "register"
	ActionPasswordChange   = "password_change"
	ActionPasswordReset    = "password_reset"
	ActionEmailVerify      = "email_verify"
	ActionPermissionGrant  = "permission_grant"
	ActionPermissionRevoke = "permission_revoke"
	ActionRoleAssign       = "role_assign"
	ActionRoleRemove       = "role_remove"
	ActionFileUpload       = "file_upload"
	ActionFileDelete       = "file_delete"
	ActionOrderCreate      = "order_create"
	ActionOrderUpdate      = "order_update"
	ActionOrderCancel      = "order_cancel"
	ActionPaymentProcess   = "payment_process"
	ActionInventoryUpdate  = "inventory_update"
	ActionProductCreate    = "product_create"
	ActionProductUpdate    = "product_update"
	ActionProductDelete    = "product_delete"
	ActionCategoryCreate   = "category_create"
	ActionCategoryUpdate   = "category_update"
	ActionCategoryDelete   = "category_delete"
	ActionBrandCreate      = "brand_create"
	ActionBrandUpdate      = "brand_update"
	ActionBrandDelete      = "brand_delete"
	ActionReviewCreate     = "review_create"
	ActionReviewUpdate     = "review_update"
	ActionReviewDelete     = "review_delete"
	ActionCouponCreate     = "coupon_create"
	ActionCouponUpdate     = "coupon_update"
	ActionCouponDelete     = "coupon_delete"
	ActionNotificationSend = "notification_send"
	ActionReportGenerate   = "report_generate"
	ActionDashboardCreate  = "dashboard_create"
	ActionDashboardUpdate  = "dashboard_update"
	ActionDashboardDelete  = "dashboard_delete"
	ActionSearch           = "search"
	ActionExport           = "export"
	ActionImport           = "import"
	ActionBackup           = "backup"
	ActionRestore          = "restore"
	ActionSystemConfig     = "system_config"

	// Resources
	ResourceUser         = "user"
	ResourceOrder        = "order"
	ResourceProduct      = "product"
	ResourceCategory     = "category"
	ResourceBrand        = "brand"
	ResourceInventory    = "inventory"
	ResourceUpload       = "upload"
	ResourcePermission   = "permission"
	ResourceRole         = "role"
	ResourceNotification = "notification"
	ResourceReport       = "report"
	ResourceDashboard    = "dashboard"
	ResourceAnalytics    = "analytics"
	ResourceSearch       = "search"
	ResourceSystem       = "system"
	ResourceAudit        = "audit"
	ResourceEmail        = "email"
	ResourcePayment      = "payment"
	ResourceShipping     = "shipping"
	ResourceReview       = "review"
	ResourceCoupon       = "coupon"
	ResourcePoint        = "point"
	ResourceBanner       = "banner"
	ResourceSlider       = "slider"
	ResourceWishlist     = "wishlist"
	ResourceAddress      = "address"
	ResourceRateLimit    = "rate_limit"
	ResourceEvent        = "event"

	// Operations
	OperationCreate   = "CREATE"
	OperationRead     = "READ"
	OperationUpdate   = "UPDATE"
	OperationDelete   = "DELETE"
	OperationLogin    = "LOGIN"
	OperationLogout   = "LOGOUT"
	OperationGrant    = "GRANT"
	OperationRevoke   = "REVOKE"
	OperationAssign   = "ASSIGN"
	OperationRemove   = "REMOVE"
	OperationUpload   = "UPLOAD"
	OperationDownload = "DOWNLOAD"
	OperationExport   = "EXPORT"
	OperationImport   = "IMPORT"
	OperationBackup   = "BACKUP"
	OperationRestore  = "RESTORE"
	OperationConfig   = "CONFIG"

	// Status
	StatusSuccess = "success"
	StatusFailure = "failure"
	StatusError   = "error"
	StatusWarning = "warning"

	// Severity
	SeverityInfo     = "info"
	SeverityWarning  = "warning"
	SeverityError    = "error"
	SeverityCritical = "critical"

	// Log Levels
	LogLevelInfo     = "info"
	LogLevelWarning  = "warning"
	LogLevelError    = "error"
	LogLevelCritical = "critical"
)

// Request/Response Models

// CreateAuditLogRequest represents request to create audit log
type CreateAuditLogRequest struct {
	UserID       *uint                  `json:"user_id"`
	Action       string                 `json:"action" validate:"required,min=2,max=100"`
	Resource     string                 `json:"resource" validate:"required,min=2,max=100"`
	ResourceID   *uint                  `json:"resource_id"`
	ResourceName string                 `json:"resource_name"`
	Operation    string                 `json:"operation" validate:"required,min=2,max=50"`
	Status       string                 `json:"status" validate:"required,oneof=success failure error warning"`
	Message      string                 `json:"message"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	Referer      string                 `json:"referer"`
	SessionID    string                 `json:"session_id"`
	OldValues    map[string]interface{} `json:"old_values"`
	NewValues    map[string]interface{} `json:"new_values"`
	Changes      map[string]interface{} `json:"changes"`
	Metadata     map[string]interface{} `json:"metadata"`
	Tags         []string               `json:"tags"`
	Severity     string                 `json:"severity" validate:"omitempty,oneof=info warning error critical"`
	TargetUserID *uint                  `json:"target_user_id"`
}

// UpdateAuditLogRequest represents request to update audit log
type UpdateAuditLogRequest struct {
	Message  string                 `json:"message"`
	Status   string                 `json:"status" validate:"omitempty,oneof=success failure error warning"`
	Severity string                 `json:"severity" validate:"omitempty,oneof=info warning error critical"`
	Metadata map[string]interface{} `json:"metadata"`
	Tags     []string               `json:"tags"`
}

// AuditLogResponse represents audit log response
type AuditLogResponse struct {
	ID           uint                   `json:"id"`
	UserID       *uint                  `json:"user_id"`
	User         *User                  `json:"user,omitempty"`
	Action       string                 `json:"action"`
	Resource     string                 `json:"resource"`
	ResourceID   *uint                  `json:"resource_id"`
	ResourceName string                 `json:"resource_name"`
	Operation    string                 `json:"operation"`
	Status       string                 `json:"status"`
	Message      string                 `json:"message"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	Referer      string                 `json:"referer"`
	SessionID    string                 `json:"session_id"`
	OldValues    map[string]interface{} `json:"old_values"`
	NewValues    map[string]interface{} `json:"new_values"`
	Changes      map[string]interface{} `json:"changes"`
	Metadata     map[string]interface{} `json:"metadata"`
	Tags         []string               `json:"tags"`
	Severity     string                 `json:"severity"`
	TargetUserID *uint                  `json:"target_user_id"`
	TargetUser   *User                  `json:"target_user,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// AuditLogConfigResponse represents audit log config response
type AuditLogConfigResponse struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	IsEnabled     bool      `json:"is_enabled"`
	LogLevel      string    `json:"log_level"`
	Resources     []string  `json:"resources"`
	Actions       []string  `json:"actions"`
	ExcludeUsers  []uint    `json:"exclude_users"`
	RetentionDays int       `json:"retention_days"`
	MaxLogSize    int       `json:"max_log_size"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// AuditLogSummaryResponse represents audit log summary response
type AuditLogSummaryResponse struct {
	ID           uint      `json:"id"`
	Date         time.Time `json:"date"`
	Resource     string    `json:"resource"`
	Action       string    `json:"action"`
	Status       string    `json:"status"`
	TotalCount   int64     `json:"total_count"`
	SuccessCount int64     `json:"success_count"`
	FailureCount int64     `json:"failure_count"`
	ErrorCount   int64     `json:"error_count"`
	UniqueUsers  int64     `json:"unique_users"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Audit Statistics Models

// AuditStats represents audit statistics
type AuditStats struct {
	TotalLogs      int64              `json:"total_logs"`
	SuccessLogs    int64              `json:"success_logs"`
	FailureLogs    int64              `json:"failure_logs"`
	ErrorLogs      int64              `json:"error_logs"`
	UniqueUsers    int64              `json:"unique_users"`
	TopActions     []ActionStats      `json:"top_actions"`
	TopResources   []ResourceStats    `json:"top_resources"`
	TopUsers       []UserStats        `json:"top_users"`
	RecentActivity []AuditLogResponse `json:"recent_activity"`
	DailyStats     []DailyStats       `json:"daily_stats"`
	HourlyStats    []HourlyStats      `json:"hourly_stats"`
}

// ActionStats represents action statistics
type ActionStats struct {
	Action      string  `json:"action"`
	Count       int64   `json:"count"`
	SuccessRate float64 `json:"success_rate"`
}

// ResourceStats represents resource statistics
type ResourceStats struct {
	Resource    string  `json:"resource"`
	Count       int64   `json:"count"`
	SuccessRate float64 `json:"success_rate"`
}

// UserStats represents user statistics
type UserStats struct {
	UserID       uint      `json:"user_id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	ActionCount  int64     `json:"action_count"`
	LastActivity time.Time `json:"last_activity"`
}

// DailyStats represents daily statistics
type DailyStats struct {
	Date        time.Time `json:"date"`
	TotalLogs   int64     `json:"total_logs"`
	SuccessLogs int64     `json:"success_logs"`
	FailureLogs int64     `json:"failure_logs"`
	ErrorLogs   int64     `json:"error_logs"`
	UniqueUsers int64     `json:"unique_users"`
}

// HourlyStats represents hourly statistics
type HourlyStats struct {
	Hour        int   `json:"hour"`
	TotalLogs   int64 `json:"total_logs"`
	SuccessLogs int64 `json:"success_logs"`
	FailureLogs int64 `json:"failure_logs"`
	ErrorLogs   int64 `json:"error_logs"`
	UniqueUsers int64 `json:"unique_users"`
}

// Audit Search Models

// AuditSearchRequest represents audit log search request
type AuditSearchRequest struct {
	Query     string                 `json:"query"`
	UserID    *uint                  `json:"user_id"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Status    string                 `json:"status"`
	Severity  string                 `json:"severity"`
	StartDate *time.Time             `json:"start_date"`
	EndDate   *time.Time             `json:"end_date"`
	IPAddress string                 `json:"ip_address"`
	SessionID string                 `json:"session_id"`
	Tags      []string               `json:"tags"`
	Metadata  map[string]interface{} `json:"metadata"`
	Page      int                    `json:"page" validate:"min=1"`
	Limit     int                    `json:"limit" validate:"min=1,max=100"`
	SortBy    string                 `json:"sort_by" validate:"omitempty,oneof=created_at action resource status severity"`
	SortOrder string                 `json:"sort_order" validate:"omitempty,oneof=asc desc"`
}

// AuditSearchResponse represents audit log search response
type AuditSearchResponse struct {
	Logs       []AuditLogResponse `json:"logs"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"total_pages"`
	HasNext    bool               `json:"has_next"`
	HasPrev    bool               `json:"has_prev"`
}

// Audit Export Models

// AuditExportRequest represents audit log export request
type AuditExportRequest struct {
	Format      string     `json:"format" validate:"required,oneof=csv json xlsx pdf"`
	UserID      *uint      `json:"user_id"`
	Action      string     `json:"action"`
	Resource    string     `json:"resource"`
	Status      string     `json:"status"`
	Severity    string     `json:"severity"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	Fields      []string   `json:"fields"`
	IncludeData bool       `json:"include_data"`
}

// AuditExportResponse represents audit log export response
type AuditExportResponse struct {
	ExportID    string    `json:"export_id"`
	Format      string    `json:"format"`
	Status      string    `json:"status"`
	DownloadURL string    `json:"download_url"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// Additional Request/Response Models

// CreateAuditConfigRequest represents request to create audit config
type CreateAuditConfigRequest struct {
	Name          string   `json:"name" validate:"required,min=3,max=255"`
	Description   string   `json:"description"`
	IsEnabled     bool     `json:"is_enabled"`
	LogLevel      string   `json:"log_level" validate:"omitempty,oneof=info warning error critical"`
	Resources     []string `json:"resources"`
	Actions       []string `json:"actions"`
	ExcludeUsers  []uint   `json:"exclude_users"`
	RetentionDays int      `json:"retention_days" validate:"min=1,max=3650"`
	MaxLogSize    int      `json:"max_log_size" validate:"min=1000"`
}

// UpdateAuditConfigRequest represents request to update audit config
type UpdateAuditConfigRequest struct {
	Name          string   `json:"name" validate:"omitempty,min=3,max=255"`
	Description   string   `json:"description"`
	IsEnabled     *bool    `json:"is_enabled"`
	LogLevel      string   `json:"log_level" validate:"omitempty,oneof=info warning error critical"`
	Resources     []string `json:"resources"`
	Actions       []string `json:"actions"`
	ExcludeUsers  []uint   `json:"exclude_users"`
	RetentionDays int      `json:"retention_days" validate:"omitempty,min=1,max=3650"`
	MaxLogSize    int      `json:"max_log_size" validate:"omitempty,min=1000"`
}
