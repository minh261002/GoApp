package model

import (
	"time"

	"gorm.io/gorm"
)

// RateLimitRule represents a rate limit rule configuration
type RateLimitRule struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Description string `json:"description" gorm:"type:text"`

	// Rate limit settings
	Requests   int    `json:"requests" gorm:"not null"`            // Number of requests allowed
	Window     int    `json:"window" gorm:"not null"`              // Time window in seconds
	WindowType string `json:"window_type" gorm:"size:20;not null"` // second, minute, hour, day

	// Target settings
	TargetType  string `json:"target_type" gorm:"size:20;not null"` // ip, user, api_key, endpoint
	TargetValue string `json:"target_value" gorm:"size:255"`        // Specific target value (optional)

	// Scope settings
	Scope      string `json:"scope" gorm:"size:50;not null"` // global, endpoint, method, path
	ScopeValue string `json:"scope_value" gorm:"size:255"`   // Specific scope value (optional)

	// Status
	IsActive bool `json:"is_active" gorm:"default:true"`
	Priority int  `json:"priority" gorm:"default:0"` // Higher number = higher priority

	// Error response
	ErrorCode    int    `json:"error_code" gorm:"default:429"`
	ErrorMessage string `json:"error_message" gorm:"type:text"`

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// RateLimitLog represents rate limit violation logs
type RateLimitLog struct {
	ID     uint           `json:"id" gorm:"primaryKey"`
	RuleID uint           `json:"rule_id" gorm:"not null;index"`
	Rule   *RateLimitRule `json:"rule,omitempty" gorm:"foreignKey:RuleID"`

	// Client information
	ClientIP string `json:"client_ip" gorm:"size:45;not null"`
	UserID   *uint  `json:"user_id" gorm:"index"`
	User     *User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	APIKey   string `json:"api_key" gorm:"size:255"`

	// Request information
	Method    string `json:"method" gorm:"size:10;not null"`
	Path      string `json:"path" gorm:"size:500;not null"`
	UserAgent string `json:"user_agent" gorm:"size:500"`
	Referer   string `json:"referer" gorm:"size:500"`

	// Rate limit information
	Limit     int       `json:"limit" gorm:"not null"`
	Current   int       `json:"current" gorm:"not null"`
	Remaining int       `json:"remaining" gorm:"not null"`
	ResetTime time.Time `json:"reset_time" gorm:"not null"`

	// Violation details
	ViolationType string `json:"violation_type" gorm:"size:50;not null"` // exceeded, blocked, bypassed
	IsBlocked     bool   `json:"is_blocked" gorm:"default:false"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// RateLimitStats represents rate limit statistics
type RateLimitStats struct {
	ID     uint           `json:"id" gorm:"primaryKey"`
	RuleID uint           `json:"rule_id" gorm:"not null;index"`
	Rule   *RateLimitRule `json:"rule,omitempty" gorm:"foreignKey:RuleID"`

	// Time period
	Period      string    `json:"period" gorm:"size:20;not null"` // hour, day, week, month
	PeriodStart time.Time `json:"period_start" gorm:"not null"`
	PeriodEnd   time.Time `json:"period_end" gorm:"not null"`

	// Statistics
	TotalRequests   int64   `json:"total_requests" gorm:"not null"`
	BlockedRequests int64   `json:"blocked_requests" gorm:"not null"`
	UniqueClients   int64   `json:"unique_clients" gorm:"not null"`
	AverageRequests float64 `json:"average_requests" gorm:"type:decimal(10,2)"`
	PeakRequests    int64   `json:"peak_requests" gorm:"not null"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// RateLimitWhitelist represents whitelisted clients/IPs
type RateLimitWhitelist struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Type        string `json:"type" gorm:"size:20;not null"` // ip, user, api_key, range
	Value       string `json:"value" gorm:"size:255;not null"`
	Description string `json:"description" gorm:"type:text"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// RateLimitBlacklist represents blacklisted clients/IPs
type RateLimitBlacklist struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	Type      string     `json:"type" gorm:"size:20;not null"` // ip, user, api_key, range
	Value     string     `json:"value" gorm:"size:255;not null"`
	Reason    string     `json:"reason" gorm:"type:text"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
	ExpiresAt *time.Time `json:"expires_at"`

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// RateLimitConfig represents rate limit configuration
type RateLimitConfig struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Description string `json:"description" gorm:"type:text"`

	// Global settings
	IsEnabled   bool `json:"is_enabled" gorm:"default:true"`
	DefaultRule uint `json:"default_rule" gorm:"not null"`

	// Redis settings
	RedisHost     string `json:"redis_host" gorm:"size:255;not null"`
	RedisPort     int    `json:"redis_port" gorm:"not null"`
	RedisDB       int    `json:"redis_db" gorm:"default:0"`
	RedisPassword string `json:"redis_password" gorm:"size:255"`

	// Cleanup settings
	LogRetentionDays   int `json:"log_retention_days" gorm:"default:30"`
	StatsRetentionDays int `json:"stats_retention_days" gorm:"default:90"`

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// Rate Limit Constants
const (
	// Target Types
	TargetTypeIP       = "ip"
	TargetTypeUser     = "user"
	TargetTypeAPIKey   = "api_key"
	TargetTypeEndpoint = "endpoint"

	// Scope Types
	ScopeGlobal   = "global"
	ScopeEndpoint = "endpoint"
	ScopeMethod   = "method"
	ScopePath     = "path"

	// Window Types
	WindowTypeSecond = "second"
	WindowTypeMinute = "minute"
	WindowTypeHour   = "hour"
	WindowTypeDay    = "day"

	// Violation Types
	ViolationTypeExceeded = "exceeded"
	ViolationTypeBlocked  = "blocked"
	ViolationTypeBypassed = "bypassed"

	// Whitelist/Blacklist Types
	ListTypeIP     = "ip"
	ListTypeUser   = "user"
	ListTypeAPIKey = "api_key"
	ListTypeRange  = "range"
)

// Request/Response Models

// RateLimitRuleRequest represents request to create/update rate limit rule
type RateLimitRuleRequest struct {
	Name         string `json:"name" validate:"required,min=3,max=100"`
	Description  string `json:"description"`
	Requests     int    `json:"requests" validate:"required,min=1"`
	Window       int    `json:"window" validate:"required,min=1"`
	WindowType   string `json:"window_type" validate:"required,oneof=second minute hour day"`
	TargetType   string `json:"target_type" validate:"required,oneof=ip user api_key endpoint"`
	TargetValue  string `json:"target_value"`
	Scope        string `json:"scope" validate:"required,oneof=global endpoint method path"`
	ScopeValue   string `json:"scope_value"`
	IsActive     bool   `json:"is_active"`
	Priority     int    `json:"priority"`
	ErrorCode    int    `json:"error_code" validate:"min=400,max=599"`
	ErrorMessage string `json:"error_message"`
}

// RateLimitRuleResponse represents rate limit rule response
type RateLimitRuleResponse struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Requests     int       `json:"requests"`
	Window       int       `json:"window"`
	WindowType   string    `json:"window_type"`
	TargetType   string    `json:"target_type"`
	TargetValue  string    `json:"target_value"`
	Scope        string    `json:"scope"`
	ScopeValue   string    `json:"scope_value"`
	IsActive     bool      `json:"is_active"`
	Priority     int       `json:"priority"`
	ErrorCode    int       `json:"error_code"`
	ErrorMessage string    `json:"error_message"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RateLimitLogResponse represents rate limit log response
type RateLimitLogResponse struct {
	ID            uint      `json:"id"`
	RuleID        uint      `json:"rule_id"`
	RuleName      string    `json:"rule_name"`
	ClientIP      string    `json:"client_ip"`
	UserID        *uint     `json:"user_id"`
	APIKey        string    `json:"api_key"`
	Method        string    `json:"method"`
	Path          string    `json:"path"`
	UserAgent     string    `json:"user_agent"`
	Referer       string    `json:"referer"`
	Limit         int       `json:"limit"`
	Current       int       `json:"current"`
	Remaining     int       `json:"remaining"`
	ResetTime     time.Time `json:"reset_time"`
	ViolationType string    `json:"violation_type"`
	IsBlocked     bool      `json:"is_blocked"`
	CreatedAt     time.Time `json:"created_at"`
}

// RateLimitStatsResponse represents rate limit statistics response
type RateLimitStatsResponse struct {
	ID              uint      `json:"id"`
	RuleID          uint      `json:"rule_id"`
	RuleName        string    `json:"rule_name"`
	Period          string    `json:"period"`
	PeriodStart     time.Time `json:"period_start"`
	PeriodEnd       time.Time `json:"period_end"`
	TotalRequests   int64     `json:"total_requests"`
	BlockedRequests int64     `json:"blocked_requests"`
	UniqueClients   int64     `json:"unique_clients"`
	AverageRequests float64   `json:"average_requests"`
	PeakRequests    int64     `json:"peak_requests"`
	CreatedAt       time.Time `json:"created_at"`
}

// RateLimitWhitelistRequest represents request to add to whitelist
type RateLimitWhitelistRequest struct {
	Type        string `json:"type" validate:"required,oneof=ip user api_key range"`
	Value       string `json:"value" validate:"required,min=1,max=255"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

// RateLimitBlacklistRequest represents request to add to blacklist
type RateLimitBlacklistRequest struct {
	Type      string     `json:"type" validate:"required,oneof=ip user api_key range"`
	Value     string     `json:"value" validate:"required,min=1,max=255"`
	Reason    string     `json:"reason"`
	IsActive  bool       `json:"is_active"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// RateLimitConfigRequest represents request to update rate limit config
type RateLimitConfigRequest struct {
	Name               string `json:"name" validate:"required,min=3,max=100"`
	Description        string `json:"description"`
	IsEnabled          bool   `json:"is_enabled"`
	DefaultRule        uint   `json:"default_rule" validate:"required"`
	RedisHost          string `json:"redis_host" validate:"required"`
	RedisPort          int    `json:"redis_port" validate:"required,min=1,max=65535"`
	RedisDB            int    `json:"redis_db" validate:"min=0,max=15"`
	RedisPassword      string `json:"redis_password"`
	LogRetentionDays   int    `json:"log_retention_days" validate:"min=1,max=365"`
	StatsRetentionDays int    `json:"stats_retention_days" validate:"min=1,max=365"`
}
