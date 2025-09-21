package model

import (
	"time"

	"gorm.io/gorm"
)

// AnalyticsReport represents a comprehensive analytics report
type AnalyticsReport struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"size:255;not null"`
	Description string `json:"description" gorm:"type:text"`
	Type        string `json:"type" gorm:"size:50;not null"` // sales, traffic, inventory, user, product, order

	// Report configuration
	Period    string    `json:"period" gorm:"size:20;not null"` // daily, weekly, monthly, yearly, custom
	StartDate time.Time `json:"start_date" gorm:"not null"`
	EndDate   time.Time `json:"end_date" gorm:"not null"`
	Filters   string    `json:"filters" gorm:"type:json"` // JSON filters

	// Report data
	Data     string `json:"data" gorm:"type:json"` // JSON report data
	Summary  string `json:"summary" gorm:"type:text"`
	Insights string `json:"insights" gorm:"type:text"`

	// Status
	Status      string `json:"status" gorm:"size:20;default:'pending'"` // pending, processing, completed, failed
	IsScheduled bool   `json:"is_scheduled" gorm:"default:false"`
	IsPublic    bool   `json:"is_public" gorm:"default:false"`

	// Relationships
	CreatedBy uint `json:"created_by" gorm:"not null"`
	Creator   User `json:"creator" gorm:"foreignKey:CreatedBy"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// AnalyticsMetric represents a single analytics metric
type AnalyticsMetric struct {
	ID       uint             `json:"id" gorm:"primaryKey"`
	ReportID uint             `json:"report_id" gorm:"not null;index"`
	Report   *AnalyticsReport `json:"report,omitempty" gorm:"foreignKey:ReportID"`

	// Metric details
	Name        string  `json:"name" gorm:"size:255;not null"`
	Value       float64 `json:"value" gorm:"type:decimal(15,2);not null"`
	Unit        string  `json:"unit" gorm:"size:50"`      // currency, count, percentage, etc.
	Category    string  `json:"category" gorm:"size:100"` // revenue, orders, users, etc.
	SubCategory string  `json:"sub_category" gorm:"size:100"`

	// Comparison data
	PreviousValue float64 `json:"previous_value" gorm:"type:decimal(15,2);default:0"`
	ChangePercent float64 `json:"change_percent" gorm:"type:decimal(5,2);default:0"`
	Trend         string  `json:"trend" gorm:"size:20"` // up, down, stable

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// AnalyticsDashboard represents a dashboard configuration
type AnalyticsDashboard struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"size:255;not null"`
	Description string `json:"description" gorm:"type:text"`
	Layout      string `json:"layout" gorm:"type:json"` // Dashboard layout configuration

	// Access control
	IsPublic bool  `json:"is_public" gorm:"default:false"`
	UserID   *uint `json:"user_id" gorm:"index"`
	User     *User `json:"user,omitempty" gorm:"foreignKey:UserID"`

	// Widgets
	Widgets string `json:"widgets" gorm:"type:json"` // Widget configurations

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// AnalyticsWidget represents a dashboard widget
type AnalyticsWidget struct {
	ID          uint                `json:"id" gorm:"primaryKey"`
	DashboardID uint                `json:"dashboard_id" gorm:"not null;index"`
	Dashboard   *AnalyticsDashboard `json:"dashboard,omitempty" gorm:"foreignKey:DashboardID"`

	// Widget details
	Type        string `json:"type" gorm:"size:50;not null"` // chart, table, metric, kpi
	Title       string `json:"title" gorm:"size:255;not null"`
	Description string `json:"description" gorm:"type:text"`

	// Configuration
	Config   string `json:"config" gorm:"type:json"`   // Widget configuration
	Data     string `json:"data" gorm:"type:json"`     // Widget data
	Position string `json:"position" gorm:"type:json"` // Position and size

	// Status
	IsActive bool `json:"is_active" gorm:"default:true"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// AnalyticsEvent represents an analytics event
type AnalyticsEvent struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	EventType  string `json:"event_type" gorm:"size:50;not null;index"`
	EventName  string `json:"event_name" gorm:"size:255;not null"`
	EntityType string `json:"entity_type" gorm:"size:50;not null"` // order, product, user, etc.
	EntityID   uint   `json:"entity_id" gorm:"not null"`

	// Event data
	Properties string  `json:"properties" gorm:"type:json"` // Event properties
	Value      float64 `json:"value" gorm:"type:decimal(15,2);default:0"`

	// User context
	UserID    *uint  `json:"user_id" gorm:"index"`
	User      *User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	SessionID string `json:"session_id" gorm:"size:255"`

	// Request context
	IPAddress string `json:"ip_address" gorm:"size:45"`
	UserAgent string `json:"user_agent" gorm:"type:text"`
	Referer   string `json:"referer" gorm:"size:500"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// Analytics Constants
const (
	// Report Types
	ReportTypeSales     = "sales"
	ReportTypeTraffic   = "traffic"
	ReportTypeInventory = "inventory"
	ReportTypeUser      = "user"
	ReportTypeProduct   = "product"
	ReportTypeOrder     = "order"
	ReportTypeRevenue   = "revenue"
	ReportTypeMarketing = "marketing"

	// Report Periods
	ReportPeriodDaily   = "daily"
	ReportPeriodWeekly  = "weekly"
	ReportPeriodMonthly = "monthly"
	ReportPeriodYearly  = "yearly"
	ReportPeriodCustom  = "custom"

	// Report Status
	ReportStatusPending    = "pending"
	ReportStatusProcessing = "processing"
	ReportStatusCompleted  = "completed"
	ReportStatusFailed     = "failed"

	// Widget Types
	WidgetTypeChart    = "chart"
	WidgetTypeTable    = "table"
	WidgetTypeMetric   = "metric"
	WidgetTypeKPI      = "kpi"
	WidgetTypeGauge    = "gauge"
	WidgetTypeProgress = "progress"
	WidgetTypeList     = "list"

	// Event Types
	EventTypePageView       = "page_view"
	EventTypeClick          = "click"
	EventTypePurchase       = "purchase"
	EventTypeAddToCart      = "add_to_cart"
	EventTypeRemoveFromCart = "remove_from_cart"
	EventTypeSearch         = "search"
	EventTypeSignUp         = "sign_up"
	EventTypeLogin          = "login"
	EventTypeLogout         = "logout"
	EventTypeEmailOpen      = "email_open"
	EventTypeEmailClick     = "email_click"

	// Entity Types
	EntityTypeOrder    = "order"
	EntityTypeProduct  = "product"
	EntityTypeUser     = "user"
	EntityTypeCategory = "category"
	EntityTypeBrand    = "brand"
	EntityTypePage     = "page"
	EntityTypeEmail    = "email"
	EntityTypeBanner   = "banner"
	EntityTypeSlider   = "slider"
)

// Analytics Report Request/Response Models

// CreateAnalyticsReportRequest represents request to create analytics report
type CreateAnalyticsReportRequest struct {
	Name        string                 `json:"name" validate:"required,min=3,max=255"`
	Description string                 `json:"description"`
	Type        string                 `json:"type" validate:"required,oneof=sales traffic inventory user product order revenue marketing"`
	Period      string                 `json:"period" validate:"required,oneof=daily weekly monthly yearly custom"`
	StartDate   time.Time              `json:"start_date" validate:"required"`
	EndDate     time.Time              `json:"end_date" validate:"required"`
	Filters     map[string]interface{} `json:"filters"`
	IsScheduled bool                   `json:"is_scheduled"`
	IsPublic    bool                   `json:"is_public"`
}

// UpdateAnalyticsReportRequest represents request to update analytics report
type UpdateAnalyticsReportRequest struct {
	Name        string                 `json:"name" validate:"omitempty,min=3,max=255"`
	Description string                 `json:"description"`
	Type        string                 `json:"type" validate:"omitempty,oneof=sales traffic inventory user product order revenue marketing"`
	Period      string                 `json:"period" validate:"omitempty,oneof=daily weekly monthly yearly custom"`
	StartDate   *time.Time             `json:"start_date"`
	EndDate     *time.Time             `json:"end_date"`
	Filters     map[string]interface{} `json:"filters"`
	IsScheduled *bool                  `json:"is_scheduled"`
	IsPublic    *bool                  `json:"is_public"`
}

// AnalyticsReportResponse represents analytics report response
type AnalyticsReportResponse struct {
	ID          uint                   `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Period      string                 `json:"period"`
	StartDate   time.Time              `json:"start_date"`
	EndDate     time.Time              `json:"end_date"`
	Filters     map[string]interface{} `json:"filters"`
	Data        map[string]interface{} `json:"data"`
	Summary     string                 `json:"summary"`
	Insights    string                 `json:"insights"`
	Status      string                 `json:"status"`
	IsScheduled bool                   `json:"is_scheduled"`
	IsPublic    bool                   `json:"is_public"`
	CreatedBy   uint                   `json:"created_by"`
	Creator     User                   `json:"creator"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// AnalyticsMetricResponse represents analytics metric response
type AnalyticsMetricResponse struct {
	ID            uint      `json:"id"`
	ReportID      uint      `json:"report_id"`
	Name          string    `json:"name"`
	Value         float64   `json:"value"`
	Unit          string    `json:"unit"`
	Category      string    `json:"category"`
	SubCategory   string    `json:"sub_category"`
	PreviousValue float64   `json:"previous_value"`
	ChangePercent float64   `json:"change_percent"`
	Trend         string    `json:"trend"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// AnalyticsDashboardResponse represents analytics dashboard response
type AnalyticsDashboardResponse struct {
	ID          uint                      `json:"id"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Layout      map[string]interface{}    `json:"layout"`
	IsPublic    bool                      `json:"is_public"`
	UserID      *uint                     `json:"user_id"`
	User        *User                     `json:"user,omitempty"`
	Widgets     []AnalyticsWidgetResponse `json:"widgets"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
}

// AnalyticsWidgetResponse represents analytics widget response
type AnalyticsWidgetResponse struct {
	ID          uint                   `json:"id"`
	DashboardID uint                   `json:"dashboard_id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	Data        map[string]interface{} `json:"data"`
	Position    map[string]interface{} `json:"position"`
	IsActive    bool                   `json:"is_active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// AnalyticsEventResponse represents analytics event response
type AnalyticsEventResponse struct {
	ID         uint                   `json:"id"`
	EventType  string                 `json:"event_type"`
	EventName  string                 `json:"event_name"`
	EntityType string                 `json:"entity_type"`
	EntityID   uint                   `json:"entity_id"`
	Properties map[string]interface{} `json:"properties"`
	Value      float64                `json:"value"`
	UserID     *uint                  `json:"user_id"`
	User       *User                  `json:"user,omitempty"`
	SessionID  string                 `json:"session_id"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	Referer    string                 `json:"referer"`
	CreatedAt  time.Time              `json:"created_at"`
}

// Analytics Summary Models

// SalesAnalytics represents sales analytics data
type SalesAnalytics struct {
	TotalRevenue      float64             `json:"total_revenue"`
	TotalOrders       int64               `json:"total_orders"`
	AverageOrderValue float64             `json:"average_order_value"`
	ConversionRate    float64             `json:"conversion_rate"`
	RevenueGrowth     float64             `json:"revenue_growth"`
	OrderGrowth       float64             `json:"order_growth"`
	TopProducts       []ProductSalesData  `json:"top_products"`
	TopCategories     []CategorySalesData `json:"top_categories"`
	RevenueByPeriod   []PeriodData        `json:"revenue_by_period"`
	OrdersByPeriod    []PeriodData        `json:"orders_by_period"`
}

// ProductSalesData represents product sales data
type ProductSalesData struct {
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	SKU         string  `json:"sku"`
	Quantity    int64   `json:"quantity"`
	Revenue     float64 `json:"revenue"`
	Orders      int64   `json:"orders"`
}

// CategorySalesData represents category sales data
type CategorySalesData struct {
	CategoryID   uint    `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Revenue      float64 `json:"revenue"`
	Orders       int64   `json:"orders"`
	Products     int64   `json:"products"`
}

// PeriodData represents data for a specific period
type PeriodData struct {
	Period string  `json:"period"`
	Value  float64 `json:"value"`
	Count  int64   `json:"count"`
}

// TrafficAnalytics represents traffic analytics data
type TrafficAnalytics struct {
	TotalVisitors          int64        `json:"total_visitors"`
	UniqueVisitors         int64        `json:"unique_visitors"`
	PageViews              int64        `json:"page_views"`
	BounceRate             float64      `json:"bounce_rate"`
	AverageSessionDuration float64      `json:"average_session_duration"`
	TopPages               []PageData   `json:"top_pages"`
	TrafficSources         []SourceData `json:"traffic_sources"`
	VisitorsByPeriod       []PeriodData `json:"visitors_by_period"`
	PageViewsByPeriod      []PeriodData `json:"page_views_by_period"`
}

// PageData represents page analytics data
type PageData struct {
	Page       string  `json:"page"`
	Views      int64   `json:"views"`
	Visitors   int64   `json:"visitors"`
	BounceRate float64 `json:"bounce_rate"`
}

// SourceData represents traffic source data
type SourceData struct {
	Source    string  `json:"source"`
	Visitors  int64   `json:"visitors"`
	PageViews int64   `json:"page_views"`
	Revenue   float64 `json:"revenue"`
}

// UserAnalytics represents user analytics data
type UserAnalytics struct {
	TotalUsers       int64         `json:"total_users"`
	NewUsers         int64         `json:"new_users"`
	ActiveUsers      int64         `json:"active_users"`
	RetentionRate    float64       `json:"retention_rate"`
	UserGrowth       float64       `json:"user_growth"`
	TopCountries     []CountryData `json:"top_countries"`
	UserSegments     []SegmentData `json:"user_segments"`
	UsersByPeriod    []PeriodData  `json:"users_by_period"`
	ActivityByPeriod []PeriodData  `json:"activity_by_period"`
}

// CountryData represents country analytics data
type CountryData struct {
	Country string  `json:"country"`
	Users   int64   `json:"users"`
	Revenue float64 `json:"revenue"`
	Orders  int64   `json:"orders"`
}

// SegmentData represents user segment data
type SegmentData struct {
	Segment       string  `json:"segment"`
	Users         int64   `json:"users"`
	Revenue       float64 `json:"revenue"`
	Orders        int64   `json:"orders"`
	AvgOrderValue float64 `json:"avg_order_value"`
}

// InventoryAnalytics represents inventory analytics data
type InventoryAnalytics struct {
	TotalProducts       int64                   `json:"total_products"`
	InStockProducts     int64                   `json:"in_stock_products"`
	OutOfStockProducts  int64                   `json:"out_of_stock_products"`
	LowStockProducts    int64                   `json:"low_stock_products"`
	TotalValue          float64                 `json:"total_value"`
	TurnoverRate        float64                 `json:"turnover_rate"`
	TopSellingProducts  []ProductInventoryData  `json:"top_selling_products"`
	LowStockAlerts      []LowStockData          `json:"low_stock_alerts"`
	InventoryByCategory []CategoryInventoryData `json:"inventory_by_category"`
}

// ProductInventoryData represents product inventory data
type ProductInventoryData struct {
	ProductID    uint    `json:"product_id"`
	ProductName  string  `json:"product_name"`
	SKU          string  `json:"sku"`
	CurrentStock int64   `json:"current_stock"`
	MinStock     int64   `json:"min_stock"`
	MaxStock     int64   `json:"max_stock"`
	Value        float64 `json:"value"`
	TurnoverRate float64 `json:"turnover_rate"`
}

// LowStockData represents low stock alert data
type LowStockData struct {
	ProductID    uint   `json:"product_id"`
	ProductName  string `json:"product_name"`
	SKU          string `json:"sku"`
	CurrentStock int64  `json:"current_stock"`
	MinStock     int64  `json:"min_stock"`
	DaysLeft     int64  `json:"days_left"`
}

// CategoryInventoryData represents category inventory data
type CategoryInventoryData struct {
	CategoryID   uint    `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Products     int64   `json:"products"`
	TotalValue   float64 `json:"total_value"`
	AvgValue     float64 `json:"avg_value"`
}

// Additional Request/Response Models

// CreateAnalyticsDashboardRequest represents request to create analytics dashboard
type CreateAnalyticsDashboardRequest struct {
	Name        string                 `json:"name" validate:"required,min=3,max=255"`
	Description string                 `json:"description"`
	Layout      map[string]interface{} `json:"layout"`
	IsPublic    bool                   `json:"is_public"`
}

// UpdateAnalyticsDashboardRequest represents request to update analytics dashboard
type UpdateAnalyticsDashboardRequest struct {
	Name        string                 `json:"name" validate:"omitempty,min=3,max=255"`
	Description string                 `json:"description"`
	Layout      map[string]interface{} `json:"layout"`
	IsPublic    *bool                  `json:"is_public"`
}

// CreateAnalyticsWidgetRequest represents request to create analytics widget
type CreateAnalyticsWidgetRequest struct {
	DashboardID uint                   `json:"dashboard_id" validate:"required"`
	Type        string                 `json:"type" validate:"required,oneof=chart table metric kpi gauge progress list"`
	Title       string                 `json:"title" validate:"required,min=3,max=255"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	Data        map[string]interface{} `json:"data"`
	Position    map[string]interface{} `json:"position"`
}

// UpdateAnalyticsWidgetRequest represents request to update analytics widget
type UpdateAnalyticsWidgetRequest struct {
	Title       string                 `json:"title" validate:"omitempty,min=3,max=255"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	Data        map[string]interface{} `json:"data"`
	Position    map[string]interface{} `json:"position"`
	IsActive    *bool                  `json:"is_active"`
}
