package model

import (
	"time"
)

// BannerType represents the type of banner
type BannerType string

const (
	BannerTypeImage    BannerType = "image"
	BannerTypeVideo    BannerType = "video"
	BannerTypeCarousel BannerType = "carousel"
	BannerTypeText     BannerType = "text"
)

// BannerPosition represents where the banner is displayed
type BannerPosition string

const (
	BannerPositionHeader   BannerPosition = "header"
	BannerPositionFooter   BannerPosition = "footer"
	BannerPositionSidebar  BannerPosition = "sidebar"
	BannerPositionMain     BannerPosition = "main"
	BannerPositionPopup    BannerPosition = "popup"
	BannerPositionMobile   BannerPosition = "mobile"
	BannerPositionDesktop  BannerPosition = "desktop"
	BannerPositionCategory BannerPosition = "category"
	BannerPositionProduct  BannerPosition = "product"
)

// BannerStatus represents the status of a banner
type BannerStatus string

const (
	BannerStatusActive   BannerStatus = "active"
	BannerStatusInactive BannerStatus = "inactive"
	BannerStatusDraft    BannerStatus = "draft"
	BannerStatusExpired  BannerStatus = "expired"
)

// Banner represents a banner in the system
type Banner struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Title       string         `json:"title" gorm:"size:255;not null"`
	Description string         `json:"description" gorm:"type:text"`
	Type        BannerType     `json:"type" gorm:"size:50;not null"`
	Position    BannerPosition `json:"position" gorm:"size:50;not null"`
	Status      BannerStatus   `json:"status" gorm:"size:50;not null;default:'draft'"`

	// Content
	ImageURL    string `json:"image_url" gorm:"size:500"`
	VideoURL    string `json:"video_url" gorm:"size:500"`
	TextContent string `json:"text_content" gorm:"type:text"`
	ButtonText  string `json:"button_text" gorm:"size:100"`
	ButtonURL   string `json:"button_url" gorm:"size:500"`

	// Display settings
	Width     int    `json:"width" gorm:"default:0"`
	Height    int    `json:"height" gorm:"default:0"`
	AltText   string `json:"alt_text" gorm:"size:255"`
	CSSClass  string `json:"css_class" gorm:"size:255"`
	SortOrder int    `json:"sort_order" gorm:"default:0"`

	// Targeting
	TargetAudience string `json:"target_audience" gorm:"size:255"` // all, new_users, returning_users, vip
	DeviceType     string `json:"device_type" gorm:"size:50"`      // all, mobile, desktop, tablet
	Location       string `json:"location" gorm:"size:255"`        // all, specific countries/cities

	// Scheduling
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`

	// Analytics
	ClickCount      int64 `json:"click_count" gorm:"default:0"`
	ViewCount       int64 `json:"view_count" gorm:"default:0"`
	ImpressionCount int64 `json:"impression_count" gorm:"default:0"`

	// SEO
	MetaTitle       string `json:"meta_title" gorm:"size:255"`
	MetaDescription string `json:"meta_description" gorm:"type:text"`
	MetaKeywords    string `json:"meta_keywords" gorm:"size:500"`

	// Relationships
	CreatedBy uint `json:"created_by" gorm:"not null"`
	Creator   User `json:"creator" gorm:"foreignKey:CreatedBy"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

// SliderType represents the type of slider
type SliderType string

const (
	SliderTypeImage SliderType = "image"
	SliderTypeVideo SliderType = "video"
	SliderTypeMixed SliderType = "mixed"
)

// SliderStatus represents the status of a slider
type SliderStatus string

const (
	SliderStatusActive   SliderStatus = "active"
	SliderStatusInactive SliderStatus = "inactive"
	SliderStatusDraft    SliderStatus = "draft"
)

// Slider represents a slider in the system
type Slider struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Name        string       `json:"name" gorm:"size:255;not null"`
	Description string       `json:"description" gorm:"type:text"`
	Type        SliderType   `json:"type" gorm:"size:50;not null"`
	Status      SliderStatus `json:"status" gorm:"size:50;not null;default:'draft'"`

	// Display settings
	Width         int    `json:"width" gorm:"default:0"`
	Height        int    `json:"height" gorm:"default:0"`
	AutoPlay      bool   `json:"auto_play" gorm:"default:true"`
	AutoPlayDelay int    `json:"auto_play_delay" gorm:"default:5000"` // milliseconds
	ShowDots      bool   `json:"show_dots" gorm:"default:true"`
	ShowArrows    bool   `json:"show_arrows" gorm:"default:true"`
	InfiniteLoop  bool   `json:"infinite_loop" gorm:"default:true"`
	FadeEffect    bool   `json:"fade_effect" gorm:"default:false"`
	CSSClass      string `json:"css_class" gorm:"size:255"`
	SortOrder     int    `json:"sort_order" gorm:"default:0"`

	// Targeting
	TargetAudience string `json:"target_audience" gorm:"size:255"`
	DeviceType     string `json:"device_type" gorm:"size:50"`
	Location       string `json:"location" gorm:"size:255"`

	// Scheduling
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`

	// Analytics
	ViewCount       int64 `json:"view_count" gorm:"default:0"`
	ImpressionCount int64 `json:"impression_count" gorm:"default:0"`

	// SEO
	MetaTitle       string `json:"meta_title" gorm:"size:255"`
	MetaDescription string `json:"meta_description" gorm:"type:text"`
	MetaKeywords    string `json:"meta_keywords" gorm:"size:500"`

	// Relationships
	CreatedBy uint         `json:"created_by" gorm:"not null"`
	Creator   User         `json:"creator" gorm:"foreignKey:CreatedBy"`
	Items     []SliderItem `json:"items" gorm:"foreignKey:SliderID"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

// SliderItem represents an item in a slider
type SliderItem struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	SliderID    uint   `json:"slider_id" gorm:"not null"`
	Title       string `json:"title" gorm:"size:255"`
	Description string `json:"description" gorm:"type:text"`

	// Content
	ImageURL    string `json:"image_url" gorm:"size:500"`
	VideoURL    string `json:"video_url" gorm:"size:500"`
	TextContent string `json:"text_content" gorm:"type:text"`
	ButtonText  string `json:"button_text" gorm:"size:100"`
	ButtonURL   string `json:"button_url" gorm:"size:500"`

	// Display settings
	Width     int    `json:"width" gorm:"default:0"`
	Height    int    `json:"height" gorm:"default:0"`
	AltText   string `json:"alt_text" gorm:"size:255"`
	CSSClass  string `json:"css_class" gorm:"size:255"`
	SortOrder int    `json:"sort_order" gorm:"default:0"`

	// Analytics
	ClickCount      int64 `json:"click_count" gorm:"default:0"`
	ViewCount       int64 `json:"view_count" gorm:"default:0"`
	ImpressionCount int64 `json:"impression_count" gorm:"default:0"`

	// Relationships
	Slider Slider `json:"slider" gorm:"foreignKey:SliderID"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

// Request/Response DTOs

// BannerCreateRequest represents the request to create a banner
type BannerCreateRequest struct {
	Title       string         `json:"title" binding:"required,min=1,max=255"`
	Description string         `json:"description" binding:"omitempty,max=1000"`
	Type        BannerType     `json:"type" binding:"required,oneof=image video carousel text"`
	Position    BannerPosition `json:"position" binding:"required,oneof=header footer sidebar main popup mobile desktop category product"`
	Status      BannerStatus   `json:"status" binding:"omitempty,oneof=active inactive draft expired"`

	// Content
	ImageURL    string `json:"image_url" binding:"omitempty,url"`
	VideoURL    string `json:"video_url" binding:"omitempty,url"`
	TextContent string `json:"text_content" binding:"omitempty,max=2000"`
	ButtonText  string `json:"button_text" binding:"omitempty,max=100"`
	ButtonURL   string `json:"button_url" binding:"omitempty,url"`

	// Display settings
	Width     int    `json:"width" binding:"omitempty,min=0,max=2000"`
	Height    int    `json:"height" binding:"omitempty,min=0,max=2000"`
	AltText   string `json:"alt_text" binding:"omitempty,max=255"`
	CSSClass  string `json:"css_class" binding:"omitempty,max=255"`
	SortOrder int    `json:"sort_order" binding:"omitempty,min=0"`

	// Targeting
	TargetAudience string `json:"target_audience" binding:"omitempty,oneof=all new_users returning_users vip"`
	DeviceType     string `json:"device_type" binding:"omitempty,oneof=all mobile desktop tablet"`
	Location       string `json:"location" binding:"omitempty,max=255"`

	// Scheduling
	StartDate *time.Time `json:"start_date" binding:"omitempty"`
	EndDate   *time.Time `json:"end_date" binding:"omitempty"`

	// SEO
	MetaTitle       string `json:"meta_title" binding:"omitempty,max=255"`
	MetaDescription string `json:"meta_description" binding:"omitempty,max=500"`
	MetaKeywords    string `json:"meta_keywords" binding:"omitempty,max=500"`
}

// BannerUpdateRequest represents the request to update a banner
type BannerUpdateRequest struct {
	Title       *string         `json:"title" binding:"omitempty,min=1,max=255"`
	Description *string         `json:"description" binding:"omitempty,max=1000"`
	Type        *BannerType     `json:"type" binding:"omitempty,oneof=image video carousel text"`
	Position    *BannerPosition `json:"position" binding:"omitempty,oneof=header footer sidebar main popup mobile desktop category product"`
	Status      *BannerStatus   `json:"status" binding:"omitempty,oneof=active inactive draft expired"`

	// Content
	ImageURL    *string `json:"image_url" binding:"omitempty,url"`
	VideoURL    *string `json:"video_url" binding:"omitempty,url"`
	TextContent *string `json:"text_content" binding:"omitempty,max=2000"`
	ButtonText  *string `json:"button_text" binding:"omitempty,max=100"`
	ButtonURL   *string `json:"button_url" binding:"omitempty,url"`

	// Display settings
	Width     *int    `json:"width" binding:"omitempty,min=0,max=2000"`
	Height    *int    `json:"height" binding:"omitempty,min=0,max=2000"`
	AltText   *string `json:"alt_text" binding:"omitempty,max=255"`
	CSSClass  *string `json:"css_class" binding:"omitempty,max=255"`
	SortOrder *int    `json:"sort_order" binding:"omitempty,min=0"`

	// Targeting
	TargetAudience *string `json:"target_audience" binding:"omitempty,oneof=all new_users returning_users vip"`
	DeviceType     *string `json:"device_type" binding:"omitempty,oneof=all mobile desktop tablet"`
	Location       *string `json:"location" binding:"omitempty,max=255"`

	// Scheduling
	StartDate *time.Time `json:"start_date" binding:"omitempty"`
	EndDate   *time.Time `json:"end_date" binding:"omitempty"`

	// SEO
	MetaTitle       *string `json:"meta_title" binding:"omitempty,max=255"`
	MetaDescription *string `json:"meta_description" binding:"omitempty,max=500"`
	MetaKeywords    *string `json:"meta_keywords" binding:"omitempty,max=500"`
}

// BannerResponse represents the response for a banner
type BannerResponse struct {
	ID          uint           `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Type        BannerType     `json:"type"`
	Position    BannerPosition `json:"position"`
	Status      BannerStatus   `json:"status"`

	// Content
	ImageURL    string `json:"image_url"`
	VideoURL    string `json:"video_url"`
	TextContent string `json:"text_content"`
	ButtonText  string `json:"button_text"`
	ButtonURL   string `json:"button_url"`

	// Display settings
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	AltText   string `json:"alt_text"`
	CSSClass  string `json:"css_class"`
	SortOrder int    `json:"sort_order"`

	// Targeting
	TargetAudience string `json:"target_audience"`
	DeviceType     string `json:"device_type"`
	Location       string `json:"location"`

	// Scheduling
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`

	// Analytics
	ClickCount      int64 `json:"click_count"`
	ViewCount       int64 `json:"view_count"`
	ImpressionCount int64 `json:"impression_count"`

	// SEO
	MetaTitle       string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`
	MetaKeywords    string `json:"meta_keywords"`

	// Relationships
	CreatedBy uint `json:"created_by"`
	Creator   User `json:"creator"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

// SliderCreateRequest represents the request to create a slider
type SliderCreateRequest struct {
	Name        string       `json:"name" binding:"required,min=1,max=255"`
	Description string       `json:"description" binding:"omitempty,max=1000"`
	Type        SliderType   `json:"type" binding:"required,oneof=image video mixed"`
	Status      SliderStatus `json:"status" binding:"omitempty,oneof=active inactive draft"`

	// Display settings
	Width         *int   `json:"width" binding:"omitempty,min=0,max=2000"`
	Height        *int   `json:"height" binding:"omitempty,min=0,max=2000"`
	AutoPlay      *bool  `json:"auto_play" binding:"omitempty"`
	AutoPlayDelay *int   `json:"auto_play_delay" binding:"omitempty,min=1000,max=30000"`
	ShowDots      *bool  `json:"show_dots" binding:"omitempty"`
	ShowArrows    *bool  `json:"show_arrows" binding:"omitempty"`
	InfiniteLoop  *bool  `json:"infinite_loop" binding:"omitempty"`
	FadeEffect    *bool  `json:"fade_effect" binding:"omitempty"`
	CSSClass      string `json:"css_class" binding:"omitempty,max=255"`
	SortOrder     *int   `json:"sort_order" binding:"omitempty,min=0"`

	// Targeting
	TargetAudience string `json:"target_audience" binding:"omitempty,oneof=all new_users returning_users vip"`
	DeviceType     string `json:"device_type" binding:"omitempty,oneof=all mobile desktop tablet"`
	Location       string `json:"location" binding:"omitempty,max=255"`

	// Scheduling
	StartDate *time.Time `json:"start_date" binding:"omitempty"`
	EndDate   *time.Time `json:"end_date" binding:"omitempty"`

	// SEO
	MetaTitle       string `json:"meta_title" binding:"omitempty,max=255"`
	MetaDescription string `json:"meta_description" binding:"omitempty,max=500"`
	MetaKeywords    string `json:"meta_keywords" binding:"omitempty,max=500"`
}

// SliderUpdateRequest represents the request to update a slider
type SliderUpdateRequest struct {
	Name        *string       `json:"name" binding:"omitempty,min=1,max=255"`
	Description *string       `json:"description" binding:"omitempty,max=1000"`
	Type        *SliderType   `json:"type" binding:"omitempty,oneof=image video mixed"`
	Status      *SliderStatus `json:"status" binding:"omitempty,oneof=active inactive draft"`

	// Display settings
	Width         *int    `json:"width" binding:"omitempty,min=0,max=2000"`
	Height        *int    `json:"height" binding:"omitempty,min=0,max=2000"`
	AutoPlay      *bool   `json:"auto_play" binding:"omitempty"`
	AutoPlayDelay *int    `json:"auto_play_delay" binding:"omitempty,min=1000,max=30000"`
	ShowDots      *bool   `json:"show_dots" binding:"omitempty"`
	ShowArrows    *bool   `json:"show_arrows" binding:"omitempty"`
	InfiniteLoop  *bool   `json:"infinite_loop" binding:"omitempty"`
	FadeEffect    *bool   `json:"fade_effect" binding:"omitempty"`
	CSSClass      *string `json:"css_class" binding:"omitempty,max=255"`
	SortOrder     *int    `json:"sort_order" binding:"omitempty,min=0"`

	// Targeting
	TargetAudience *string `json:"target_audience" binding:"omitempty,oneof=all new_users returning_users vip"`
	DeviceType     *string `json:"device_type" binding:"omitempty,oneof=all mobile desktop tablet"`
	Location       *string `json:"location" binding:"omitempty,max=255"`

	// Scheduling
	StartDate *time.Time `json:"start_date" binding:"omitempty"`
	EndDate   *time.Time `json:"end_date" binding:"omitempty"`

	// SEO
	MetaTitle       *string `json:"meta_title" binding:"omitempty,max=255"`
	MetaDescription *string `json:"meta_description" binding:"omitempty,max=500"`
	MetaKeywords    *string `json:"meta_keywords" binding:"omitempty,max=500"`
}

// SliderResponse represents the response for a slider
type SliderResponse struct {
	ID          uint         `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Type        SliderType   `json:"type"`
	Status      SliderStatus `json:"status"`

	// Display settings
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	AutoPlay      bool   `json:"auto_play"`
	AutoPlayDelay int    `json:"auto_play_delay"`
	ShowDots      bool   `json:"show_dots"`
	ShowArrows    bool   `json:"show_arrows"`
	InfiniteLoop  bool   `json:"infinite_loop"`
	FadeEffect    bool   `json:"fade_effect"`
	CSSClass      string `json:"css_class"`
	SortOrder     int    `json:"sort_order"`

	// Targeting
	TargetAudience string `json:"target_audience"`
	DeviceType     string `json:"device_type"`
	Location       string `json:"location"`

	// Scheduling
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`

	// Analytics
	ViewCount       int64 `json:"view_count"`
	ImpressionCount int64 `json:"impression_count"`

	// SEO
	MetaTitle       string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`
	MetaKeywords    string `json:"meta_keywords"`

	// Relationships
	CreatedBy uint                 `json:"created_by"`
	Creator   User                 `json:"creator"`
	Items     []SliderItemResponse `json:"items"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

// SliderItemCreateRequest represents the request to create a slider item
type SliderItemCreateRequest struct {
	SliderID    uint   `json:"slider_id" binding:"required"`
	Title       string `json:"title" binding:"omitempty,max=255"`
	Description string `json:"description" binding:"omitempty,max=1000"`

	// Content
	ImageURL    string `json:"image_url" binding:"omitempty,url"`
	VideoURL    string `json:"video_url" binding:"omitempty,url"`
	TextContent string `json:"text_content" binding:"omitempty,max=2000"`
	ButtonText  string `json:"button_text" binding:"omitempty,max=100"`
	ButtonURL   string `json:"button_url" binding:"omitempty,url"`

	// Display settings
	Width     *int   `json:"width" binding:"omitempty,min=0,max=2000"`
	Height    *int   `json:"height" binding:"omitempty,min=0,max=2000"`
	AltText   string `json:"alt_text" binding:"omitempty,max=255"`
	CSSClass  string `json:"css_class" binding:"omitempty,max=255"`
	SortOrder *int   `json:"sort_order" binding:"omitempty,min=0"`
}

// SliderItemUpdateRequest represents the request to update a slider item
type SliderItemUpdateRequest struct {
	Title       *string `json:"title" binding:"omitempty,max=255"`
	Description *string `json:"description" binding:"omitempty,max=1000"`

	// Content
	ImageURL    *string `json:"image_url" binding:"omitempty,url"`
	VideoURL    *string `json:"video_url" binding:"omitempty,url"`
	TextContent *string `json:"text_content" binding:"omitempty,max=2000"`
	ButtonText  *string `json:"button_text" binding:"omitempty,max=100"`
	ButtonURL   *string `json:"button_url" binding:"omitempty,url"`

	// Display settings
	Width     *int    `json:"width" binding:"omitempty,min=0,max=2000"`
	Height    *int    `json:"height" binding:"omitempty,min=0,max=2000"`
	AltText   *string `json:"alt_text" binding:"omitempty,max=255"`
	CSSClass  *string `json:"css_class" binding:"omitempty,max=255"`
	SortOrder *int    `json:"sort_order" binding:"omitempty,min=0"`
}

// SliderItemResponse represents the response for a slider item
type SliderItemResponse struct {
	ID          uint   `json:"id"`
	SliderID    uint   `json:"slider_id"`
	Title       string `json:"title"`
	Description string `json:"description"`

	// Content
	ImageURL    string `json:"image_url"`
	VideoURL    string `json:"video_url"`
	TextContent string `json:"text_content"`
	ButtonText  string `json:"button_text"`
	ButtonURL   string `json:"button_url"`

	// Display settings
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	AltText   string `json:"alt_text"`
	CSSClass  string `json:"css_class"`
	SortOrder int    `json:"sort_order"`

	// Analytics
	ClickCount      int64 `json:"click_count"`
	ViewCount       int64 `json:"view_count"`
	ImpressionCount int64 `json:"impression_count"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

// BannerStatsResponse represents banner statistics
type BannerStatsResponse struct {
	TotalBanners     int64            `json:"total_banners"`
	ActiveBanners    int64            `json:"active_banners"`
	InactiveBanners  int64            `json:"inactive_banners"`
	DraftBanners     int64            `json:"draft_banners"`
	ExpiredBanners   int64            `json:"expired_banners"`
	TotalClicks      int64            `json:"total_clicks"`
	TotalViews       int64            `json:"total_views"`
	TotalImpressions int64            `json:"total_impressions"`
	TopBanners       []BannerResponse `json:"top_banners"`
}

// SliderStatsResponse represents slider statistics
type SliderStatsResponse struct {
	TotalSliders     int64            `json:"total_sliders"`
	ActiveSliders    int64            `json:"active_sliders"`
	InactiveSliders  int64            `json:"inactive_sliders"`
	DraftSliders     int64            `json:"draft_sliders"`
	TotalViews       int64            `json:"total_views"`
	TotalImpressions int64            `json:"total_impressions"`
	TopSliders       []SliderResponse `json:"top_sliders"`
}

// BannerClickRequest represents a banner click tracking request
type BannerClickRequest struct {
	BannerID  uint   `json:"banner_id" binding:"required"`
	UserID    *uint  `json:"user_id" binding:"omitempty"`
	IP        string `json:"ip" binding:"omitempty"`
	UserAgent string `json:"user_agent" binding:"omitempty"`
	Referrer  string `json:"referrer" binding:"omitempty"`
}

// BannerViewRequest represents a banner view tracking request
type BannerViewRequest struct {
	BannerID  uint   `json:"banner_id" binding:"required"`
	UserID    *uint  `json:"user_id" binding:"omitempty"`
	IP        string `json:"ip" binding:"omitempty"`
	UserAgent string `json:"user_agent" binding:"omitempty"`
	Referrer  string `json:"referrer" binding:"omitempty"`
}

// SliderViewRequest represents a slider view tracking request
type SliderViewRequest struct {
	SliderID  uint   `json:"slider_id" binding:"required"`
	UserID    *uint  `json:"user_id" binding:"omitempty"`
	IP        string `json:"ip" binding:"omitempty"`
	UserAgent string `json:"user_agent" binding:"omitempty"`
	Referrer  string `json:"referrer" binding:"omitempty"`
}

// SliderItemClickRequest represents a slider item click tracking request
type SliderItemClickRequest struct {
	ItemID    uint   `json:"item_id" binding:"required"`
	UserID    *uint  `json:"user_id" binding:"omitempty"`
	IP        string `json:"ip" binding:"omitempty"`
	UserAgent string `json:"user_agent" binding:"omitempty"`
	Referrer  string `json:"referrer" binding:"omitempty"`
}

// BannerFilterRequest represents banner filtering options
type BannerFilterRequest struct {
	Type           *BannerType     `json:"type" form:"type"`
	Position       *BannerPosition `json:"position" form:"position"`
	Status         *BannerStatus   `json:"status" form:"status"`
	TargetAudience *string         `json:"target_audience" form:"target_audience"`
	DeviceType     *string         `json:"device_type" form:"device_type"`
	Location       *string         `json:"location" form:"location"`
	CreatedBy      *uint           `json:"created_by" form:"created_by"`
	StartDate      *time.Time      `json:"start_date" form:"start_date"`
	EndDate        *time.Time      `json:"end_date" form:"end_date"`
	Search         *string         `json:"search" form:"search"`
}

// SliderFilterRequest represents slider filtering options
type SliderFilterRequest struct {
	Type           *SliderType   `json:"type" form:"type"`
	Status         *SliderStatus `json:"status" form:"status"`
	TargetAudience *string       `json:"target_audience" form:"target_audience"`
	DeviceType     *string       `json:"device_type" form:"device_type"`
	Location       *string       `json:"location" form:"location"`
	CreatedBy      *uint         `json:"created_by" form:"created_by"`
	StartDate      *time.Time    `json:"start_date" form:"start_date"`
	EndDate        *time.Time    `json:"end_date" form:"end_date"`
	Search         *string       `json:"search" form:"search"`
}

// Helper methods

// IsActive checks if a banner is currently active
func (b *Banner) IsActive() bool {
	now := time.Now()
	return b.Status == BannerStatusActive &&
		(b.StartDate == nil || now.After(*b.StartDate)) &&
		(b.EndDate == nil || now.Before(*b.EndDate))
}

// IsExpired checks if a banner has expired
func (b *Banner) IsExpired() bool {
	now := time.Now()
	return b.EndDate != nil && now.After(*b.EndDate)
}

// IsActive checks if a slider is currently active
func (s *Slider) IsActive() bool {
	now := time.Now()
	return s.Status == SliderStatusActive &&
		(s.StartDate == nil || now.After(*s.StartDate)) &&
		(s.EndDate == nil || now.Before(*s.EndDate))
}

// IsExpired checks if a slider has expired
func (s *Slider) IsExpired() bool {
	now := time.Now()
	return s.EndDate != nil && now.After(*s.EndDate)
}

// ToResponse converts Banner to BannerResponse
func (b *Banner) ToResponse() *BannerResponse {
	return &BannerResponse{
		ID:              b.ID,
		Title:           b.Title,
		Description:     b.Description,
		Type:            b.Type,
		Position:        b.Position,
		Status:          b.Status,
		ImageURL:        b.ImageURL,
		VideoURL:        b.VideoURL,
		TextContent:     b.TextContent,
		ButtonText:      b.ButtonText,
		ButtonURL:       b.ButtonURL,
		Width:           b.Width,
		Height:          b.Height,
		AltText:         b.AltText,
		CSSClass:        b.CSSClass,
		SortOrder:       b.SortOrder,
		TargetAudience:  b.TargetAudience,
		DeviceType:      b.DeviceType,
		Location:        b.Location,
		StartDate:       b.StartDate,
		EndDate:         b.EndDate,
		ClickCount:      b.ClickCount,
		ViewCount:       b.ViewCount,
		ImpressionCount: b.ImpressionCount,
		MetaTitle:       b.MetaTitle,
		MetaDescription: b.MetaDescription,
		MetaKeywords:    b.MetaKeywords,
		CreatedBy:       b.CreatedBy,
		Creator:         b.Creator,
		CreatedAt:       b.CreatedAt,
		UpdatedAt:       b.UpdatedAt,
		DeletedAt:       b.DeletedAt,
	}
}

// ToResponse converts Slider to SliderResponse
func (s *Slider) ToResponse() *SliderResponse {
	items := make([]SliderItemResponse, len(s.Items))
	for i, item := range s.Items {
		items[i] = *item.ToResponse()
	}

	return &SliderResponse{
		ID:              s.ID,
		Name:            s.Name,
		Description:     s.Description,
		Type:            s.Type,
		Status:          s.Status,
		Width:           s.Width,
		Height:          s.Height,
		AutoPlay:        s.AutoPlay,
		AutoPlayDelay:   s.AutoPlayDelay,
		ShowDots:        s.ShowDots,
		ShowArrows:      s.ShowArrows,
		InfiniteLoop:    s.InfiniteLoop,
		FadeEffect:      s.FadeEffect,
		CSSClass:        s.CSSClass,
		SortOrder:       s.SortOrder,
		TargetAudience:  s.TargetAudience,
		DeviceType:      s.DeviceType,
		Location:        s.Location,
		StartDate:       s.StartDate,
		EndDate:         s.EndDate,
		ViewCount:       s.ViewCount,
		ImpressionCount: s.ImpressionCount,
		MetaTitle:       s.MetaTitle,
		MetaDescription: s.MetaDescription,
		MetaKeywords:    s.MetaKeywords,
		CreatedBy:       s.CreatedBy,
		Creator:         s.Creator,
		Items:           items,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
		DeletedAt:       s.DeletedAt,
	}
}

// ToResponse converts SliderItem to SliderItemResponse
func (si *SliderItem) ToResponse() *SliderItemResponse {
	return &SliderItemResponse{
		ID:              si.ID,
		SliderID:        si.SliderID,
		Title:           si.Title,
		Description:     si.Description,
		ImageURL:        si.ImageURL,
		VideoURL:        si.VideoURL,
		TextContent:     si.TextContent,
		ButtonText:      si.ButtonText,
		ButtonURL:       si.ButtonURL,
		Width:           si.Width,
		Height:          si.Height,
		AltText:         si.AltText,
		CSSClass:        si.CSSClass,
		SortOrder:       si.SortOrder,
		ClickCount:      si.ClickCount,
		ViewCount:       si.ViewCount,
		ImpressionCount: si.ImpressionCount,
		CreatedAt:       si.CreatedAt,
		UpdatedAt:       si.UpdatedAt,
		DeletedAt:       si.DeletedAt,
	}
}
