package repository

import (
	"fmt"
	"go_app/internal/model"
	"go_app/pkg/database"
	"time"

	"gorm.io/gorm"
)

// BannerRepository defines methods for interacting with banner data
type BannerRepository interface {
	// Basic CRUD
	CreateBanner(banner *model.Banner) error
	GetBannerByID(id uint) (*model.Banner, error)
	GetAllBanners(page, limit int, filters map[string]interface{}) ([]model.Banner, int64, error)
	UpdateBanner(banner *model.Banner) error
	DeleteBanner(id uint) error

	// Banner Management
	GetActiveBanners(page, limit int) ([]model.Banner, int64, error)
	GetBannersByType(bannerType model.BannerType, page, limit int) ([]model.Banner, int64, error)
	GetBannersByPosition(position model.BannerPosition, page, limit int) ([]model.Banner, int64, error)
	GetBannersByStatus(status model.BannerStatus, page, limit int) ([]model.Banner, int64, error)
	SearchBanners(query string, page, limit int) ([]model.Banner, int64, error)
	GetBannersByTargetAudience(audience string, page, limit int) ([]model.Banner, int64, error)
	GetBannersByDeviceType(deviceType string, page, limit int) ([]model.Banner, int64, error)

	// Analytics
	TrackBannerClick(bannerID uint, userID *uint, ip, userAgent, referrer string) error
	TrackBannerView(bannerID uint, userID *uint, ip, userAgent, referrer string) error
	GetBannerStats() (*model.BannerStatsResponse, error)
	GetBannerAnalytics(bannerID uint, startDate, endDate *time.Time) (map[string]interface{}, error)

	// Scheduling
	GetExpiredBanners() ([]model.Banner, error)
	GetBannersToActivate() ([]model.Banner, error)
	UpdateBannerStatus(id uint, status model.BannerStatus) error
}

// SliderRepository defines methods for interacting with slider data
type SliderRepository interface {
	// Basic CRUD
	CreateSlider(slider *model.Slider) error
	GetSliderByID(id uint) (*model.Slider, error)
	GetAllSliders(page, limit int, filters map[string]interface{}) ([]model.Slider, int64, error)
	UpdateSlider(slider *model.Slider) error
	DeleteSlider(id uint) error

	// Slider Management
	GetActiveSliders(page, limit int) ([]model.Slider, int64, error)
	GetSlidersByType(sliderType model.SliderType, page, limit int) ([]model.Slider, int64, error)
	GetSlidersByStatus(status model.SliderStatus, page, limit int) ([]model.Slider, int64, error)
	SearchSliders(query string, page, limit int) ([]model.Slider, int64, error)
	GetSlidersByTargetAudience(audience string, page, limit int) ([]model.Slider, int64, error)
	GetSlidersByDeviceType(deviceType string, page, limit int) ([]model.Slider, int64, error)

	// Slider Items
	CreateSliderItem(item *model.SliderItem) error
	GetSliderItemByID(id uint) (*model.SliderItem, error)
	GetSliderItemsBySlider(sliderID uint, page, limit int) ([]model.SliderItem, int64, error)
	UpdateSliderItem(item *model.SliderItem) error
	DeleteSliderItem(id uint) error
	ReorderSliderItems(sliderID uint, itemOrders map[uint]int) error

	// Analytics
	TrackSliderView(sliderID uint, userID *uint, ip, userAgent, referrer string) error
	TrackSliderItemClick(itemID uint, userID *uint, ip, userAgent, referrer string) error
	GetSliderStats() (*model.SliderStatsResponse, error)
	GetSliderAnalytics(sliderID uint, startDate, endDate *time.Time) (map[string]interface{}, error)

	// Scheduling
	GetExpiredSliders() ([]model.Slider, error)
	GetSlidersToActivate() ([]model.Slider, error)
	UpdateSliderStatus(id uint, status model.SliderStatus) error
}

// bannerRepository implements BannerRepository
type bannerRepository struct {
	db *gorm.DB
}

// sliderRepository implements SliderRepository
type sliderRepository struct {
	db *gorm.DB
}

// NewBannerRepository creates a new BannerRepository
func NewBannerRepository() BannerRepository {
	return &bannerRepository{
		db: database.DB,
	}
}

// NewSliderRepository creates a new SliderRepository
func NewSliderRepository() SliderRepository {
	return &sliderRepository{
		db: database.DB,
	}
}

// Banner Repository Implementation

// CreateBanner creates a new banner
func (r *bannerRepository) CreateBanner(banner *model.Banner) error {
	return r.db.Create(banner).Error
}

// GetBannerByID retrieves a banner by its ID
func (r *bannerRepository) GetBannerByID(id uint) (*model.Banner, error) {
	var banner model.Banner
	if err := r.db.Preload("Creator").
		First(&banner, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &banner, nil
}

// GetAllBanners retrieves all banners with pagination and filters
func (r *bannerRepository) GetAllBanners(page, limit int, filters map[string]interface{}) ([]model.Banner, int64, error) {
	var banners []model.Banner
	var total int64
	db := r.db.Model(&model.Banner{})

	// Apply filters
	for key, value := range filters {
		switch key {
		case "type":
			db = db.Where("type = ?", value)
		case "position":
			db = db.Where("position = ?", value)
		case "status":
			db = db.Where("status = ?", value)
		case "target_audience":
			db = db.Where("target_audience = ?", value)
		case "device_type":
			db = db.Where("device_type = ?", value)
		case "location":
			db = db.Where("location = ?", value)
		case "created_by":
			db = db.Where("created_by = ?", value)
		case "start_date":
			db = db.Where("start_date >= ?", value)
		case "end_date":
			db = db.Where("end_date <= ?", value)
		case "search":
			searchTerm := fmt.Sprintf("%%%s%%", value.(string))
			db = db.Where("title LIKE ? OR description LIKE ? OR alt_text LIKE ?", searchTerm, searchTerm, searchTerm)
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
	db = db.Order("sort_order ASC, created_at DESC")

	if err := db.Preload("Creator").Find(&banners).Error; err != nil {
		return nil, 0, err
	}

	return banners, total, nil
}

// UpdateBanner updates an existing banner
func (r *bannerRepository) UpdateBanner(banner *model.Banner) error {
	return r.db.Save(banner).Error
}

// DeleteBanner soft deletes a banner
func (r *bannerRepository) DeleteBanner(id uint) error {
	return r.db.Delete(&model.Banner{}, id).Error
}

// GetActiveBanners retrieves active banners
func (r *bannerRepository) GetActiveBanners(page, limit int) ([]model.Banner, int64, error) {
	var banners []model.Banner
	var total int64
	now := time.Now()

	db := r.db.Model(&model.Banner{}).Where("status = ? AND (start_date IS NULL OR start_date <= ?) AND (end_date IS NULL OR end_date >= ?)",
		model.BannerStatusActive, now, now)

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
	db = db.Order("sort_order ASC, created_at DESC")

	if err := db.Preload("Creator").Find(&banners).Error; err != nil {
		return nil, 0, err
	}

	return banners, total, nil
}

// GetBannersByType retrieves banners by type
func (r *bannerRepository) GetBannersByType(bannerType model.BannerType, page, limit int) ([]model.Banner, int64, error) {
	var banners []model.Banner
	var total int64
	db := r.db.Model(&model.Banner{}).Where("type = ?", bannerType)

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
	db = db.Order("sort_order ASC, created_at DESC")

	if err := db.Preload("Creator").Find(&banners).Error; err != nil {
		return nil, 0, err
	}

	return banners, total, nil
}

// GetBannersByPosition retrieves banners by position
func (r *bannerRepository) GetBannersByPosition(position model.BannerPosition, page, limit int) ([]model.Banner, int64, error) {
	var banners []model.Banner
	var total int64
	db := r.db.Model(&model.Banner{}).Where("position = ?", position)

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
	db = db.Order("sort_order ASC, created_at DESC")

	if err := db.Preload("Creator").Find(&banners).Error; err != nil {
		return nil, 0, err
	}

	return banners, total, nil
}

// GetBannersByStatus retrieves banners by status
func (r *bannerRepository) GetBannersByStatus(status model.BannerStatus, page, limit int) ([]model.Banner, int64, error) {
	var banners []model.Banner
	var total int64
	db := r.db.Model(&model.Banner{}).Where("status = ?", status)

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
	db = db.Order("sort_order ASC, created_at DESC")

	if err := db.Preload("Creator").Find(&banners).Error; err != nil {
		return nil, 0, err
	}

	return banners, total, nil
}

// SearchBanners performs full-text search on banners
func (r *bannerRepository) SearchBanners(query string, page, limit int) ([]model.Banner, int64, error) {
	var banners []model.Banner
	var total int64

	// Use MATCH AGAINST for full-text search
	db := r.db.Model(&model.Banner{}).
		Where("MATCH(title, description, alt_text) AGAINST(? IN NATURAL LANGUAGE MODE)", query)

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
	db = db.Order("MATCH(title, description, alt_text) AGAINST('" + query + "' IN NATURAL LANGUAGE MODE) DESC, created_at DESC")

	if err := db.Preload("Creator").Find(&banners).Error; err != nil {
		return nil, 0, err
	}

	return banners, total, nil
}

// GetBannersByTargetAudience retrieves banners by target audience
func (r *bannerRepository) GetBannersByTargetAudience(audience string, page, limit int) ([]model.Banner, int64, error) {
	var banners []model.Banner
	var total int64
	db := r.db.Model(&model.Banner{}).Where("target_audience = ? OR target_audience = 'all'", audience)

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
	db = db.Order("sort_order ASC, created_at DESC")

	if err := db.Preload("Creator").Find(&banners).Error; err != nil {
		return nil, 0, err
	}

	return banners, total, nil
}

// GetBannersByDeviceType retrieves banners by device type
func (r *bannerRepository) GetBannersByDeviceType(deviceType string, page, limit int) ([]model.Banner, int64, error) {
	var banners []model.Banner
	var total int64
	db := r.db.Model(&model.Banner{}).Where("device_type = ? OR device_type = 'all'", deviceType)

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
	db = db.Order("sort_order ASC, created_at DESC")

	if err := db.Preload("Creator").Find(&banners).Error; err != nil {
		return nil, 0, err
	}

	return banners, total, nil
}

// TrackBannerClick tracks a banner click
func (r *bannerRepository) TrackBannerClick(bannerID uint, userID *uint, ip, userAgent, referrer string) error {
	// Create click record
	click := map[string]interface{}{
		"banner_id":  bannerID,
		"user_id":    userID,
		"ip_address": ip,
		"user_agent": userAgent,
		"referrer":   referrer,
		"clicked_at": time.Now(),
	}

	if err := r.db.Table("banner_clicks").Create(click).Error; err != nil {
		return err
	}

	// Update banner click count
	return r.db.Model(&model.Banner{}).Where("id = ?", bannerID).
		Update("click_count", gorm.Expr("click_count + 1")).Error
}

// TrackBannerView tracks a banner view
func (r *bannerRepository) TrackBannerView(bannerID uint, userID *uint, ip, userAgent, referrer string) error {
	// Create view record
	view := map[string]interface{}{
		"banner_id":  bannerID,
		"user_id":    userID,
		"ip_address": ip,
		"user_agent": userAgent,
		"referrer":   referrer,
		"viewed_at":  time.Now(),
	}

	if err := r.db.Table("banner_views").Create(view).Error; err != nil {
		return err
	}

	// Update banner view count
	return r.db.Model(&model.Banner{}).Where("id = ?", bannerID).
		Update("view_count", gorm.Expr("view_count + 1")).Error
}

// GetBannerStats retrieves banner statistics
func (r *bannerRepository) GetBannerStats() (*model.BannerStatsResponse, error) {
	var stats model.BannerStatsResponse
	var count int64

	// Total banners
	r.db.Model(&model.Banner{}).Count(&count)
	stats.TotalBanners = count

	// Active banners
	r.db.Model(&model.Banner{}).Where("status = ?", model.BannerStatusActive).Count(&count)
	stats.ActiveBanners = count

	// Inactive banners
	r.db.Model(&model.Banner{}).Where("status = ?", model.BannerStatusInactive).Count(&count)
	stats.InactiveBanners = count

	// Draft banners
	r.db.Model(&model.Banner{}).Where("status = ?", model.BannerStatusDraft).Count(&count)
	stats.DraftBanners = count

	// Expired banners
	now := time.Now()
	r.db.Model(&model.Banner{}).Where("end_date < ?", now).Count(&count)
	stats.ExpiredBanners = count

	// Total clicks
	var totalClicks int64
	r.db.Model(&model.Banner{}).Select("SUM(click_count)").Scan(&totalClicks)
	stats.TotalClicks = totalClicks

	// Total views
	var totalViews int64
	r.db.Model(&model.Banner{}).Select("SUM(view_count)").Scan(&totalViews)
	stats.TotalViews = totalViews

	// Total impressions
	var totalImpressions int64
	r.db.Model(&model.Banner{}).Select("SUM(impression_count)").Scan(&totalImpressions)
	stats.TotalImpressions = totalImpressions

	// Top banners by clicks
	var topBanners []model.Banner
	r.db.Model(&model.Banner{}).
		Order("click_count DESC").
		Limit(10).
		Preload("Creator").
		Find(&topBanners)

	for _, banner := range topBanners {
		stats.TopBanners = append(stats.TopBanners, *banner.ToResponse())
	}

	return &stats, nil
}

// GetBannerAnalytics retrieves analytics for a specific banner
func (r *bannerRepository) GetBannerAnalytics(bannerID uint, startDate, endDate *time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Base query
	query := r.db.Table("banner_clicks").Where("banner_id = ?", bannerID)
	if startDate != nil {
		query = query.Where("clicked_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("clicked_at <= ?", *endDate)
	}

	// Click count
	var clickCount int64
	query.Count(&clickCount)
	stats["clicks"] = clickCount

	// View count
	viewQuery := r.db.Table("banner_views").Where("banner_id = ?", bannerID)
	if startDate != nil {
		viewQuery = viewQuery.Where("viewed_at >= ?", *startDate)
	}
	if endDate != nil {
		viewQuery = viewQuery.Where("viewed_at <= ?", *endDate)
	}

	var viewCount int64
	viewQuery.Count(&viewCount)
	stats["views"] = viewCount

	// Click-through rate
	if viewCount > 0 {
		stats["ctr"] = float64(clickCount) / float64(viewCount) * 100
	} else {
		stats["ctr"] = 0.0
	}

	// Unique users
	var uniqueUsers int64
	query.Distinct("user_id").Count(&uniqueUsers)
	stats["unique_users"] = uniqueUsers

	return stats, nil
}

// GetExpiredBanners retrieves expired banners
func (r *bannerRepository) GetExpiredBanners() ([]model.Banner, error) {
	var banners []model.Banner
	now := time.Now()
	err := r.db.Where("end_date < ? AND status != ?", now, model.BannerStatusExpired).
		Preload("Creator").Find(&banners).Error
	return banners, err
}

// GetBannersToActivate retrieves banners that should be activated
func (r *bannerRepository) GetBannersToActivate() ([]model.Banner, error) {
	var banners []model.Banner
	now := time.Now()
	err := r.db.Where("start_date <= ? AND status = ?", now, model.BannerStatusDraft).
		Preload("Creator").Find(&banners).Error
	return banners, err
}

// UpdateBannerStatus updates banner status
func (r *bannerRepository) UpdateBannerStatus(id uint, status model.BannerStatus) error {
	return r.db.Model(&model.Banner{}).Where("id = ?", id).Update("status", status).Error
}

// Slider Repository Implementation

// CreateSlider creates a new slider
func (r *sliderRepository) CreateSlider(slider *model.Slider) error {
	return r.db.Create(slider).Error
}

// GetSliderByID retrieves a slider by its ID
func (r *sliderRepository) GetSliderByID(id uint) (*model.Slider, error) {
	var slider model.Slider
	if err := r.db.Preload("Creator").Preload("Items").
		First(&slider, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &slider, nil
}

// GetAllSliders retrieves all sliders with pagination and filters
func (r *sliderRepository) GetAllSliders(page, limit int, filters map[string]interface{}) ([]model.Slider, int64, error) {
	var sliders []model.Slider
	var total int64
	db := r.db.Model(&model.Slider{})

	// Apply filters
	for key, value := range filters {
		switch key {
		case "type":
			db = db.Where("type = ?", value)
		case "status":
			db = db.Where("status = ?", value)
		case "target_audience":
			db = db.Where("target_audience = ?", value)
		case "device_type":
			db = db.Where("device_type = ?", value)
		case "location":
			db = db.Where("location = ?", value)
		case "created_by":
			db = db.Where("created_by = ?", value)
		case "start_date":
			db = db.Where("start_date >= ?", value)
		case "end_date":
			db = db.Where("end_date <= ?", value)
		case "search":
			searchTerm := fmt.Sprintf("%%%s%%", value.(string))
			db = db.Where("name LIKE ? OR description LIKE ?", searchTerm, searchTerm)
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
	db = db.Order("sort_order ASC, created_at DESC")

	if err := db.Preload("Creator").Preload("Items").
		Find(&sliders).Error; err != nil {
		return nil, 0, err
	}

	return sliders, total, nil
}

// UpdateSlider updates an existing slider
func (r *sliderRepository) UpdateSlider(slider *model.Slider) error {
	return r.db.Save(slider).Error
}

// DeleteSlider soft deletes a slider
func (r *sliderRepository) DeleteSlider(id uint) error {
	return r.db.Delete(&model.Slider{}, id).Error
}

// GetActiveSliders retrieves active sliders
func (r *sliderRepository) GetActiveSliders(page, limit int) ([]model.Slider, int64, error) {
	var sliders []model.Slider
	var total int64
	now := time.Now()

	db := r.db.Model(&model.Slider{}).Where("status = ? AND (start_date IS NULL OR start_date <= ?) AND (end_date IS NULL OR end_date >= ?)",
		model.SliderStatusActive, now, now)

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
	db = db.Order("sort_order ASC, created_at DESC")

	if err := db.Preload("Creator").Preload("Items").
		Find(&sliders).Error; err != nil {
		return nil, 0, err
	}

	return sliders, total, nil
}

// GetSlidersByType retrieves sliders by type
func (r *sliderRepository) GetSlidersByType(sliderType model.SliderType, page, limit int) ([]model.Slider, int64, error) {
	var sliders []model.Slider
	var total int64
	db := r.db.Model(&model.Slider{}).Where("type = ?", sliderType)

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
	db = db.Order("sort_order ASC, created_at DESC")

	if err := db.Preload("Creator").Preload("Items").
		Find(&sliders).Error; err != nil {
		return nil, 0, err
	}

	return sliders, total, nil
}

// GetSlidersByStatus retrieves sliders by status
func (r *sliderRepository) GetSlidersByStatus(status model.SliderStatus, page, limit int) ([]model.Slider, int64, error) {
	var sliders []model.Slider
	var total int64
	db := r.db.Model(&model.Slider{}).Where("status = ?", status)

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
	db = db.Order("sort_order ASC, created_at DESC")

	if err := db.Preload("Creator").Preload("Items").
		Find(&sliders).Error; err != nil {
		return nil, 0, err
	}

	return sliders, total, nil
}

// SearchSliders performs full-text search on sliders
func (r *sliderRepository) SearchSliders(query string, page, limit int) ([]model.Slider, int64, error) {
	var sliders []model.Slider
	var total int64

	// Use MATCH AGAINST for full-text search
	db := r.db.Model(&model.Slider{}).
		Where("MATCH(name, description) AGAINST(? IN NATURAL LANGUAGE MODE)", query)

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
	db = db.Order("MATCH(name, description) AGAINST('" + query + "' IN NATURAL LANGUAGE MODE) DESC, created_at DESC")

	if err := db.Preload("Creator").Preload("Items").
		Find(&sliders).Error; err != nil {
		return nil, 0, err
	}

	return sliders, total, nil
}

// GetSlidersByTargetAudience retrieves sliders by target audience
func (r *sliderRepository) GetSlidersByTargetAudience(audience string, page, limit int) ([]model.Slider, int64, error) {
	var sliders []model.Slider
	var total int64
	db := r.db.Model(&model.Slider{}).Where("target_audience = ? OR target_audience = 'all'", audience)

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
	db = db.Order("sort_order ASC, created_at DESC")

	if err := db.Preload("Creator").Preload("Items").
		Find(&sliders).Error; err != nil {
		return nil, 0, err
	}

	return sliders, total, nil
}

// GetSlidersByDeviceType retrieves sliders by device type
func (r *sliderRepository) GetSlidersByDeviceType(deviceType string, page, limit int) ([]model.Slider, int64, error) {
	var sliders []model.Slider
	var total int64
	db := r.db.Model(&model.Slider{}).Where("device_type = ? OR device_type = 'all'", deviceType)

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
	db = db.Order("sort_order ASC, created_at DESC")

	if err := db.Preload("Creator").Preload("Items").
		Find(&sliders).Error; err != nil {
		return nil, 0, err
	}

	return sliders, total, nil
}

// Slider Items

// CreateSliderItem creates a new slider item
func (r *sliderRepository) CreateSliderItem(item *model.SliderItem) error {
	return r.db.Create(item).Error
}

// GetSliderItemByID retrieves a slider item by its ID
func (r *sliderRepository) GetSliderItemByID(id uint) (*model.SliderItem, error) {
	var item model.SliderItem
	if err := r.db.Preload("Slider").
		First(&item, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

// GetSliderItemsBySlider retrieves slider items for a specific slider
func (r *sliderRepository) GetSliderItemsBySlider(sliderID uint, page, limit int) ([]model.SliderItem, int64, error) {
	var items []model.SliderItem
	var total int64
	db := r.db.Model(&model.SliderItem{}).Where("slider_id = ?", sliderID)

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
	db = db.Order("sort_order ASC, created_at ASC")

	if err := db.Preload("Slider").Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// UpdateSliderItem updates an existing slider item
func (r *sliderRepository) UpdateSliderItem(item *model.SliderItem) error {
	return r.db.Save(item).Error
}

// DeleteSliderItem soft deletes a slider item
func (r *sliderRepository) DeleteSliderItem(id uint) error {
	return r.db.Delete(&model.SliderItem{}, id).Error
}

// ReorderSliderItems reorders slider items
func (r *sliderRepository) ReorderSliderItems(sliderID uint, itemOrders map[uint]int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for itemID, order := range itemOrders {
			if err := tx.Model(&model.SliderItem{}).
				Where("id = ? AND slider_id = ?", itemID, sliderID).
				Update("sort_order", order).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// TrackSliderView tracks a slider view
func (r *sliderRepository) TrackSliderView(sliderID uint, userID *uint, ip, userAgent, referrer string) error {
	// Create view record
	view := map[string]interface{}{
		"slider_id":  sliderID,
		"user_id":    userID,
		"ip_address": ip,
		"user_agent": userAgent,
		"referrer":   referrer,
		"viewed_at":  time.Now(),
	}

	if err := r.db.Table("slider_views").Create(view).Error; err != nil {
		return err
	}

	// Update slider view count
	return r.db.Model(&model.Slider{}).Where("id = ?", sliderID).
		Update("view_count", gorm.Expr("view_count + 1")).Error
}

// TrackSliderItemClick tracks a slider item click
func (r *sliderRepository) TrackSliderItemClick(itemID uint, userID *uint, ip, userAgent, referrer string) error {
	// Create click record
	click := map[string]interface{}{
		"item_id":    itemID,
		"user_id":    userID,
		"ip_address": ip,
		"user_agent": userAgent,
		"referrer":   referrer,
		"clicked_at": time.Now(),
	}

	if err := r.db.Table("slider_item_clicks").Create(click).Error; err != nil {
		return err
	}

	// Update slider item click count
	return r.db.Model(&model.SliderItem{}).Where("id = ?", itemID).
		Update("click_count", gorm.Expr("click_count + 1")).Error
}

// GetSliderStats retrieves slider statistics
func (r *sliderRepository) GetSliderStats() (*model.SliderStatsResponse, error) {
	var stats model.SliderStatsResponse
	var count int64

	// Total sliders
	r.db.Model(&model.Slider{}).Count(&count)
	stats.TotalSliders = count

	// Active sliders
	r.db.Model(&model.Slider{}).Where("status = ?", model.SliderStatusActive).Count(&count)
	stats.ActiveSliders = count

	// Inactive sliders
	r.db.Model(&model.Slider{}).Where("status = ?", model.SliderStatusInactive).Count(&count)
	stats.InactiveSliders = count

	// Draft sliders
	r.db.Model(&model.Slider{}).Where("status = ?", model.SliderStatusDraft).Count(&count)
	stats.DraftSliders = count

	// Total views
	var totalViews int64
	r.db.Model(&model.Slider{}).Select("SUM(view_count)").Scan(&totalViews)
	stats.TotalViews = totalViews

	// Total impressions
	var totalImpressions int64
	r.db.Model(&model.Slider{}).Select("SUM(impression_count)").Scan(&totalImpressions)
	stats.TotalImpressions = totalImpressions

	// Top sliders by views
	var topSliders []model.Slider
	r.db.Model(&model.Slider{}).
		Order("view_count DESC").
		Limit(10).
		Preload("Creator").
		Find(&topSliders)

	for _, slider := range topSliders {
		stats.TopSliders = append(stats.TopSliders, *slider.ToResponse())
	}

	return &stats, nil
}

// GetSliderAnalytics retrieves analytics for a specific slider
func (r *sliderRepository) GetSliderAnalytics(sliderID uint, startDate, endDate *time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// View count
	viewQuery := r.db.Table("slider_views").Where("slider_id = ?", sliderID)
	if startDate != nil {
		viewQuery = viewQuery.Where("viewed_at >= ?", *startDate)
	}
	if endDate != nil {
		viewQuery = viewQuery.Where("viewed_at <= ?", *endDate)
	}

	var viewCount int64
	viewQuery.Count(&viewCount)
	stats["views"] = viewCount

	// Click count for all items in slider
	var clickCount int64
	clickQuery := r.db.Table("slider_item_clicks").
		Joins("JOIN slider_items ON slider_item_clicks.item_id = slider_items.id").
		Where("slider_items.slider_id = ?", sliderID)
	if startDate != nil {
		clickQuery = clickQuery.Where("slider_item_clicks.clicked_at >= ?", *startDate)
	}
	if endDate != nil {
		clickQuery = clickQuery.Where("slider_item_clicks.clicked_at <= ?", *endDate)
	}

	clickQuery.Count(&clickCount)
	stats["clicks"] = clickCount

	// Click-through rate
	if viewCount > 0 {
		stats["ctr"] = float64(clickCount) / float64(viewCount) * 100
	} else {
		stats["ctr"] = 0.0
	}

	// Unique users
	var uniqueUsers int64
	viewQuery.Distinct("user_id").Count(&uniqueUsers)
	stats["unique_users"] = uniqueUsers

	return stats, nil
}

// GetExpiredSliders retrieves expired sliders
func (r *sliderRepository) GetExpiredSliders() ([]model.Slider, error) {
	var sliders []model.Slider
	now := time.Now()
	err := r.db.Where("end_date < ? AND status != ?", now, model.SliderStatusInactive).
		Preload("Creator").Preload("Items").Find(&sliders).Error
	return sliders, err
}

// GetSlidersToActivate retrieves sliders that should be activated
func (r *sliderRepository) GetSlidersToActivate() ([]model.Slider, error) {
	var sliders []model.Slider
	now := time.Now()
	err := r.db.Where("start_date <= ? AND status = ?", now, model.SliderStatusDraft).
		Preload("Creator").Preload("Items").Find(&sliders).Error
	return sliders, err
}

// UpdateSliderStatus updates slider status
func (r *sliderRepository) UpdateSliderStatus(id uint, status model.SliderStatus) error {
	return r.db.Model(&model.Slider{}).Where("id = ?", id).Update("status", status).Error
}
