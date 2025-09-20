package service

import (
	"errors"
	"go_app/internal/model"
	"go_app/internal/repository"
	"time"
)

// BannerService defines the interface for banner business logic
type BannerService interface {
	// Basic CRUD
	CreateBanner(req *model.BannerCreateRequest, creatorID uint) (*model.BannerResponse, error)
	GetBannerByID(id uint) (*model.BannerResponse, error)
	GetAllBanners(page, limit int, filters map[string]interface{}) ([]model.BannerResponse, int64, error)
	UpdateBanner(id uint, req *model.BannerUpdateRequest) (*model.BannerResponse, error)
	DeleteBanner(id uint) error

	// Banner Management
	GetActiveBanners(page, limit int) ([]model.BannerResponse, int64, error)
	GetBannersByType(bannerType model.BannerType, page, limit int) ([]model.BannerResponse, int64, error)
	GetBannersByPosition(position model.BannerPosition, page, limit int) ([]model.BannerResponse, int64, error)
	GetBannersByStatus(status model.BannerStatus, page, limit int) ([]model.BannerResponse, int64, error)
	SearchBanners(query string, page, limit int) ([]model.BannerResponse, int64, error)
	GetBannersByTargetAudience(audience string, page, limit int) ([]model.BannerResponse, int64, error)
	GetBannersByDeviceType(deviceType string, page, limit int) ([]model.BannerResponse, int64, error)

	// Analytics
	TrackBannerClick(req *model.BannerClickRequest) error
	TrackBannerView(req *model.BannerViewRequest) error
	GetBannerStats() (*model.BannerStatsResponse, error)
	GetBannerAnalytics(bannerID uint, startDate, endDate *time.Time) (map[string]interface{}, error)

	// Scheduling
	GetExpiredBanners() ([]model.BannerResponse, error)
	GetBannersToActivate() ([]model.BannerResponse, error)
	UpdateBannerStatus(id uint, status model.BannerStatus) error
}

// SliderService defines the interface for slider business logic
type SliderService interface {
	// Basic CRUD
	CreateSlider(req *model.SliderCreateRequest, creatorID uint) (*model.SliderResponse, error)
	GetSliderByID(id uint) (*model.SliderResponse, error)
	GetAllSliders(page, limit int, filters map[string]interface{}) ([]model.SliderResponse, int64, error)
	UpdateSlider(id uint, req *model.SliderUpdateRequest) (*model.SliderResponse, error)
	DeleteSlider(id uint) error

	// Slider Management
	GetActiveSliders(page, limit int) ([]model.SliderResponse, int64, error)
	GetSlidersByType(sliderType model.SliderType, page, limit int) ([]model.SliderResponse, int64, error)
	GetSlidersByStatus(status model.SliderStatus, page, limit int) ([]model.SliderResponse, int64, error)
	SearchSliders(query string, page, limit int) ([]model.SliderResponse, int64, error)
	GetSlidersByTargetAudience(audience string, page, limit int) ([]model.SliderResponse, int64, error)
	GetSlidersByDeviceType(deviceType string, page, limit int) ([]model.SliderResponse, int64, error)

	// Slider Items
	CreateSliderItem(req *model.SliderItemCreateRequest) (*model.SliderItemResponse, error)
	GetSliderItemByID(id uint) (*model.SliderItemResponse, error)
	GetSliderItemsBySlider(sliderID uint, page, limit int) ([]model.SliderItemResponse, int64, error)
	UpdateSliderItem(id uint, req *model.SliderItemUpdateRequest) (*model.SliderItemResponse, error)
	DeleteSliderItem(id uint) error
	ReorderSliderItems(sliderID uint, itemOrders map[uint]int) error

	// Analytics
	TrackSliderView(req *model.SliderViewRequest) error
	TrackSliderItemClick(req *model.SliderItemClickRequest) error
	GetSliderStats() (*model.SliderStatsResponse, error)
	GetSliderAnalytics(sliderID uint, startDate, endDate *time.Time) (map[string]interface{}, error)

	// Scheduling
	GetExpiredSliders() ([]model.SliderResponse, error)
	GetSlidersToActivate() ([]model.SliderResponse, error)
	UpdateSliderStatus(id uint, status model.SliderStatus) error
}

// bannerService implements BannerService
type bannerService struct {
	bannerRepo repository.BannerRepository
}

// sliderService implements SliderService
type sliderService struct {
	sliderRepo repository.SliderRepository
}

// NewBannerService creates a new BannerService
func NewBannerService() BannerService {
	return &bannerService{
		bannerRepo: repository.NewBannerRepository(),
	}
}

// NewSliderService creates a new SliderService
func NewSliderService() SliderService {
	return &sliderService{
		sliderRepo: repository.NewSliderRepository(),
	}
}

// Banner Service Implementation

// CreateBanner creates a new banner
func (s *bannerService) CreateBanner(req *model.BannerCreateRequest, creatorID uint) (*model.BannerResponse, error) {
	// Validate content based on type
	if err := s.validateBannerContent(req); err != nil {
		return nil, err
	}

	// Set default values
	if req.Status == "" {
		req.Status = model.BannerStatusDraft
	}
	if req.TargetAudience == "" {
		req.TargetAudience = "all"
	}
	if req.DeviceType == "" {
		req.DeviceType = "all"
	}

	banner := &model.Banner{
		Title:           req.Title,
		Description:     req.Description,
		Type:            req.Type,
		Position:        req.Position,
		Status:          req.Status,
		ImageURL:        req.ImageURL,
		VideoURL:        req.VideoURL,
		TextContent:     req.TextContent,
		ButtonText:      req.ButtonText,
		ButtonURL:       req.ButtonURL,
		Width:           req.Width,
		Height:          req.Height,
		AltText:         req.AltText,
		CSSClass:        req.CSSClass,
		SortOrder:       req.SortOrder,
		TargetAudience:  req.TargetAudience,
		DeviceType:      req.DeviceType,
		Location:        req.Location,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		MetaTitle:       req.MetaTitle,
		MetaDescription: req.MetaDescription,
		MetaKeywords:    req.MetaKeywords,
		CreatedBy:       creatorID,
	}

	if err := s.bannerRepo.CreateBanner(banner); err != nil {
		return nil, err
	}

	return banner.ToResponse(), nil
}

// GetBannerByID retrieves a banner by its ID
func (s *bannerService) GetBannerByID(id uint) (*model.BannerResponse, error) {
	banner, err := s.bannerRepo.GetBannerByID(id)
	if err != nil {
		return nil, err
	}
	if banner == nil {
		return nil, errors.New("banner not found")
	}

	return banner.ToResponse(), nil
}

// GetAllBanners retrieves all banners with pagination and filters
func (s *bannerService) GetAllBanners(page, limit int, filters map[string]interface{}) ([]model.BannerResponse, int64, error) {
	banners, total, err := s.bannerRepo.GetAllBanners(page, limit, filters)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.BannerResponse, len(banners))
	for i, banner := range banners {
		responses[i] = *banner.ToResponse()
	}

	return responses, total, nil
}

// UpdateBanner updates an existing banner
func (s *bannerService) UpdateBanner(id uint, req *model.BannerUpdateRequest) (*model.BannerResponse, error) {
	banner, err := s.bannerRepo.GetBannerByID(id)
	if err != nil {
		return nil, err
	}
	if banner == nil {
		return nil, errors.New("banner not found")
	}

	// Update fields if provided
	if req.Title != nil {
		banner.Title = *req.Title
	}
	if req.Description != nil {
		banner.Description = *req.Description
	}
	if req.Type != nil {
		banner.Type = *req.Type
	}
	if req.Position != nil {
		banner.Position = *req.Position
	}
	if req.Status != nil {
		banner.Status = *req.Status
	}
	if req.ImageURL != nil {
		banner.ImageURL = *req.ImageURL
	}
	if req.VideoURL != nil {
		banner.VideoURL = *req.VideoURL
	}
	if req.TextContent != nil {
		banner.TextContent = *req.TextContent
	}
	if req.ButtonText != nil {
		banner.ButtonText = *req.ButtonText
	}
	if req.ButtonURL != nil {
		banner.ButtonURL = *req.ButtonURL
	}
	if req.Width != nil {
		banner.Width = *req.Width
	}
	if req.Height != nil {
		banner.Height = *req.Height
	}
	if req.AltText != nil {
		banner.AltText = *req.AltText
	}
	if req.CSSClass != nil {
		banner.CSSClass = *req.CSSClass
	}
	if req.SortOrder != nil {
		banner.SortOrder = *req.SortOrder
	}
	if req.TargetAudience != nil {
		banner.TargetAudience = *req.TargetAudience
	}
	if req.DeviceType != nil {
		banner.DeviceType = *req.DeviceType
	}
	if req.Location != nil {
		banner.Location = *req.Location
	}
	if req.StartDate != nil {
		banner.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		banner.EndDate = req.EndDate
	}
	if req.MetaTitle != nil {
		banner.MetaTitle = *req.MetaTitle
	}
	if req.MetaDescription != nil {
		banner.MetaDescription = *req.MetaDescription
	}
	if req.MetaKeywords != nil {
		banner.MetaKeywords = *req.MetaKeywords
	}

	// Validate updated content
	if err := s.validateBannerContentUpdate(banner); err != nil {
		return nil, err
	}

	if err := s.bannerRepo.UpdateBanner(banner); err != nil {
		return nil, err
	}

	return banner.ToResponse(), nil
}

// DeleteBanner soft deletes a banner
func (s *bannerService) DeleteBanner(id uint) error {
	banner, err := s.bannerRepo.GetBannerByID(id)
	if err != nil {
		return err
	}
	if banner == nil {
		return errors.New("banner not found")
	}

	return s.bannerRepo.DeleteBanner(id)
}

// GetActiveBanners retrieves active banners
func (s *bannerService) GetActiveBanners(page, limit int) ([]model.BannerResponse, int64, error) {
	banners, total, err := s.bannerRepo.GetActiveBanners(page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.BannerResponse, len(banners))
	for i, banner := range banners {
		responses[i] = *banner.ToResponse()
	}

	return responses, total, nil
}

// GetBannersByType retrieves banners by type
func (s *bannerService) GetBannersByType(bannerType model.BannerType, page, limit int) ([]model.BannerResponse, int64, error) {
	banners, total, err := s.bannerRepo.GetBannersByType(bannerType, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.BannerResponse, len(banners))
	for i, banner := range banners {
		responses[i] = *banner.ToResponse()
	}

	return responses, total, nil
}

// GetBannersByPosition retrieves banners by position
func (s *bannerService) GetBannersByPosition(position model.BannerPosition, page, limit int) ([]model.BannerResponse, int64, error) {
	banners, total, err := s.bannerRepo.GetBannersByPosition(position, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.BannerResponse, len(banners))
	for i, banner := range banners {
		responses[i] = *banner.ToResponse()
	}

	return responses, total, nil
}

// GetBannersByStatus retrieves banners by status
func (s *bannerService) GetBannersByStatus(status model.BannerStatus, page, limit int) ([]model.BannerResponse, int64, error) {
	banners, total, err := s.bannerRepo.GetBannersByStatus(status, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.BannerResponse, len(banners))
	for i, banner := range banners {
		responses[i] = *banner.ToResponse()
	}

	return responses, total, nil
}

// SearchBanners performs full-text search on banners
func (s *bannerService) SearchBanners(query string, page, limit int) ([]model.BannerResponse, int64, error) {
	banners, total, err := s.bannerRepo.SearchBanners(query, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.BannerResponse, len(banners))
	for i, banner := range banners {
		responses[i] = *banner.ToResponse()
	}

	return responses, total, nil
}

// GetBannersByTargetAudience retrieves banners by target audience
func (s *bannerService) GetBannersByTargetAudience(audience string, page, limit int) ([]model.BannerResponse, int64, error) {
	banners, total, err := s.bannerRepo.GetBannersByTargetAudience(audience, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.BannerResponse, len(banners))
	for i, banner := range banners {
		responses[i] = *banner.ToResponse()
	}

	return responses, total, nil
}

// GetBannersByDeviceType retrieves banners by device type
func (s *bannerService) GetBannersByDeviceType(deviceType string, page, limit int) ([]model.BannerResponse, int64, error) {
	banners, total, err := s.bannerRepo.GetBannersByDeviceType(deviceType, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.BannerResponse, len(banners))
	for i, banner := range banners {
		responses[i] = *banner.ToResponse()
	}

	return responses, total, nil
}

// TrackBannerClick tracks a banner click
func (s *bannerService) TrackBannerClick(req *model.BannerClickRequest) error {
	return s.bannerRepo.TrackBannerClick(req.BannerID, req.UserID, req.IP, req.UserAgent, req.Referrer)
}

// TrackBannerView tracks a banner view
func (s *bannerService) TrackBannerView(req *model.BannerViewRequest) error {
	return s.bannerRepo.TrackBannerView(req.BannerID, req.UserID, req.IP, req.UserAgent, req.Referrer)
}

// GetBannerStats retrieves banner statistics
func (s *bannerService) GetBannerStats() (*model.BannerStatsResponse, error) {
	return s.bannerRepo.GetBannerStats()
}

// GetBannerAnalytics retrieves analytics for a specific banner
func (s *bannerService) GetBannerAnalytics(bannerID uint, startDate, endDate *time.Time) (map[string]interface{}, error) {
	return s.bannerRepo.GetBannerAnalytics(bannerID, startDate, endDate)
}

// GetExpiredBanners retrieves expired banners
func (s *bannerService) GetExpiredBanners() ([]model.BannerResponse, error) {
	banners, err := s.bannerRepo.GetExpiredBanners()
	if err != nil {
		return nil, err
	}

	responses := make([]model.BannerResponse, len(banners))
	for i, banner := range banners {
		responses[i] = *banner.ToResponse()
	}

	return responses, nil
}

// GetBannersToActivate retrieves banners that should be activated
func (s *bannerService) GetBannersToActivate() ([]model.BannerResponse, error) {
	banners, err := s.bannerRepo.GetBannersToActivate()
	if err != nil {
		return nil, err
	}

	responses := make([]model.BannerResponse, len(banners))
	for i, banner := range banners {
		responses[i] = *banner.ToResponse()
	}

	return responses, nil
}

// UpdateBannerStatus updates banner status
func (s *bannerService) UpdateBannerStatus(id uint, status model.BannerStatus) error {
	return s.bannerRepo.UpdateBannerStatus(id, status)
}

// validateBannerContent validates banner content based on type
func (s *bannerService) validateBannerContent(req *model.BannerCreateRequest) error {
	switch req.Type {
	case model.BannerTypeImage:
		if req.ImageURL == "" {
			return errors.New("image URL is required for image banner")
		}
	case model.BannerTypeVideo:
		if req.VideoURL == "" {
			return errors.New("video URL is required for video banner")
		}
	case model.BannerTypeText:
		if req.TextContent == "" {
			return errors.New("text content is required for text banner")
		}
	case model.BannerTypeCarousel:
		// Carousel banners can have multiple content types
		if req.ImageURL == "" && req.VideoURL == "" && req.TextContent == "" {
			return errors.New("at least one content type is required for carousel banner")
		}
	}

	// Validate scheduling
	if req.StartDate != nil && req.EndDate != nil {
		if req.StartDate.After(*req.EndDate) {
			return errors.New("start date cannot be after end date")
		}
	}

	return nil
}

// validateBannerContentUpdate validates updated banner content
func (s *bannerService) validateBannerContentUpdate(banner *model.Banner) error {
	switch banner.Type {
	case model.BannerTypeImage:
		if banner.ImageURL == "" {
			return errors.New("image URL is required for image banner")
		}
	case model.BannerTypeVideo:
		if banner.VideoURL == "" {
			return errors.New("video URL is required for video banner")
		}
	case model.BannerTypeText:
		if banner.TextContent == "" {
			return errors.New("text content is required for text banner")
		}
	case model.BannerTypeCarousel:
		if banner.ImageURL == "" && banner.VideoURL == "" && banner.TextContent == "" {
			return errors.New("at least one content type is required for carousel banner")
		}
	}

	// Validate scheduling
	if banner.StartDate != nil && banner.EndDate != nil {
		if banner.StartDate.After(*banner.EndDate) {
			return errors.New("start date cannot be after end date")
		}
	}

	return nil
}

// Slider Service Implementation

// CreateSlider creates a new slider
func (s *sliderService) CreateSlider(req *model.SliderCreateRequest, creatorID uint) (*model.SliderResponse, error) {
	// Set default values
	if req.Status == "" {
		req.Status = model.SliderStatusDraft
	}
	if req.TargetAudience == "" {
		req.TargetAudience = "all"
	}
	if req.DeviceType == "" {
		req.DeviceType = "all"
	}

	slider := &model.Slider{
		Name:            req.Name,
		Description:     req.Description,
		Type:            req.Type,
		Status:          req.Status,
		Width:           *req.Width,
		Height:          *req.Height,
		AutoPlay:        *req.AutoPlay,
		AutoPlayDelay:   *req.AutoPlayDelay,
		ShowDots:        *req.ShowDots,
		ShowArrows:      *req.ShowArrows,
		InfiniteLoop:    *req.InfiniteLoop,
		FadeEffect:      *req.FadeEffect,
		CSSClass:        req.CSSClass,
		SortOrder:       *req.SortOrder,
		TargetAudience:  req.TargetAudience,
		DeviceType:      req.DeviceType,
		Location:        req.Location,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		MetaTitle:       req.MetaTitle,
		MetaDescription: req.MetaDescription,
		MetaKeywords:    req.MetaKeywords,
		CreatedBy:       creatorID,
	}

	if err := s.sliderRepo.CreateSlider(slider); err != nil {
		return nil, err
	}

	return slider.ToResponse(), nil
}

// GetSliderByID retrieves a slider by its ID
func (s *sliderService) GetSliderByID(id uint) (*model.SliderResponse, error) {
	slider, err := s.sliderRepo.GetSliderByID(id)
	if err != nil {
		return nil, err
	}
	if slider == nil {
		return nil, errors.New("slider not found")
	}

	return slider.ToResponse(), nil
}

// GetAllSliders retrieves all sliders with pagination and filters
func (s *sliderService) GetAllSliders(page, limit int, filters map[string]interface{}) ([]model.SliderResponse, int64, error) {
	sliders, total, err := s.sliderRepo.GetAllSliders(page, limit, filters)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.SliderResponse, len(sliders))
	for i, slider := range sliders {
		responses[i] = *slider.ToResponse()
	}

	return responses, total, nil
}

// UpdateSlider updates an existing slider
func (s *sliderService) UpdateSlider(id uint, req *model.SliderUpdateRequest) (*model.SliderResponse, error) {
	slider, err := s.sliderRepo.GetSliderByID(id)
	if err != nil {
		return nil, err
	}
	if slider == nil {
		return nil, errors.New("slider not found")
	}

	// Update fields if provided
	if req.Name != nil {
		slider.Name = *req.Name
	}
	if req.Description != nil {
		slider.Description = *req.Description
	}
	if req.Type != nil {
		slider.Type = *req.Type
	}
	if req.Status != nil {
		slider.Status = *req.Status
	}
	if req.Width != nil {
		slider.Width = *req.Width
	}
	if req.Height != nil {
		slider.Height = *req.Height
	}
	if req.AutoPlay != nil {
		slider.AutoPlay = *req.AutoPlay
	}
	if req.AutoPlayDelay != nil {
		slider.AutoPlayDelay = *req.AutoPlayDelay
	}
	if req.ShowDots != nil {
		slider.ShowDots = *req.ShowDots
	}
	if req.ShowArrows != nil {
		slider.ShowArrows = *req.ShowArrows
	}
	if req.InfiniteLoop != nil {
		slider.InfiniteLoop = *req.InfiniteLoop
	}
	if req.FadeEffect != nil {
		slider.FadeEffect = *req.FadeEffect
	}
	if req.CSSClass != nil {
		slider.CSSClass = *req.CSSClass
	}
	if req.SortOrder != nil {
		slider.SortOrder = *req.SortOrder
	}
	if req.TargetAudience != nil {
		slider.TargetAudience = *req.TargetAudience
	}
	if req.DeviceType != nil {
		slider.DeviceType = *req.DeviceType
	}
	if req.Location != nil {
		slider.Location = *req.Location
	}
	if req.StartDate != nil {
		slider.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		slider.EndDate = req.EndDate
	}
	if req.MetaTitle != nil {
		slider.MetaTitle = *req.MetaTitle
	}
	if req.MetaDescription != nil {
		slider.MetaDescription = *req.MetaDescription
	}
	if req.MetaKeywords != nil {
		slider.MetaKeywords = *req.MetaKeywords
	}

	// Validate scheduling
	if slider.StartDate != nil && slider.EndDate != nil {
		if slider.StartDate.After(*slider.EndDate) {
			return nil, errors.New("start date cannot be after end date")
		}
	}

	if err := s.sliderRepo.UpdateSlider(slider); err != nil {
		return nil, err
	}

	return slider.ToResponse(), nil
}

// DeleteSlider soft deletes a slider
func (s *sliderService) DeleteSlider(id uint) error {
	slider, err := s.sliderRepo.GetSliderByID(id)
	if err != nil {
		return err
	}
	if slider == nil {
		return errors.New("slider not found")
	}

	return s.sliderRepo.DeleteSlider(id)
}

// GetActiveSliders retrieves active sliders
func (s *sliderService) GetActiveSliders(page, limit int) ([]model.SliderResponse, int64, error) {
	sliders, total, err := s.sliderRepo.GetActiveSliders(page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.SliderResponse, len(sliders))
	for i, slider := range sliders {
		responses[i] = *slider.ToResponse()
	}

	return responses, total, nil
}

// GetSlidersByType retrieves sliders by type
func (s *sliderService) GetSlidersByType(sliderType model.SliderType, page, limit int) ([]model.SliderResponse, int64, error) {
	sliders, total, err := s.sliderRepo.GetSlidersByType(sliderType, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.SliderResponse, len(sliders))
	for i, slider := range sliders {
		responses[i] = *slider.ToResponse()
	}

	return responses, total, nil
}

// GetSlidersByStatus retrieves sliders by status
func (s *sliderService) GetSlidersByStatus(status model.SliderStatus, page, limit int) ([]model.SliderResponse, int64, error) {
	sliders, total, err := s.sliderRepo.GetSlidersByStatus(status, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.SliderResponse, len(sliders))
	for i, slider := range sliders {
		responses[i] = *slider.ToResponse()
	}

	return responses, total, nil
}

// SearchSliders performs full-text search on sliders
func (s *sliderService) SearchSliders(query string, page, limit int) ([]model.SliderResponse, int64, error) {
	sliders, total, err := s.sliderRepo.SearchSliders(query, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.SliderResponse, len(sliders))
	for i, slider := range sliders {
		responses[i] = *slider.ToResponse()
	}

	return responses, total, nil
}

// GetSlidersByTargetAudience retrieves sliders by target audience
func (s *sliderService) GetSlidersByTargetAudience(audience string, page, limit int) ([]model.SliderResponse, int64, error) {
	sliders, total, err := s.sliderRepo.GetSlidersByTargetAudience(audience, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.SliderResponse, len(sliders))
	for i, slider := range sliders {
		responses[i] = *slider.ToResponse()
	}

	return responses, total, nil
}

// GetSlidersByDeviceType retrieves sliders by device type
func (s *sliderService) GetSlidersByDeviceType(deviceType string, page, limit int) ([]model.SliderResponse, int64, error) {
	sliders, total, err := s.sliderRepo.GetSlidersByDeviceType(deviceType, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.SliderResponse, len(sliders))
	for i, slider := range sliders {
		responses[i] = *slider.ToResponse()
	}

	return responses, total, nil
}

// Slider Items

// CreateSliderItem creates a new slider item
func (s *sliderService) CreateSliderItem(req *model.SliderItemCreateRequest) (*model.SliderItemResponse, error) {
	// Validate slider exists
	slider, err := s.sliderRepo.GetSliderByID(req.SliderID)
	if err != nil {
		return nil, err
	}
	if slider == nil {
		return nil, errors.New("slider not found")
	}

	// Validate content based on slider type
	if err := s.validateSliderItemContent(req, slider.Type); err != nil {
		return nil, err
	}

	item := &model.SliderItem{
		SliderID:    req.SliderID,
		Title:       req.Title,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		VideoURL:    req.VideoURL,
		TextContent: req.TextContent,
		ButtonText:  req.ButtonText,
		ButtonURL:   req.ButtonURL,
		Width:       *req.Width,
		Height:      *req.Height,
		AltText:     req.AltText,
		CSSClass:    req.CSSClass,
		SortOrder:   *req.SortOrder,
	}

	if err := s.sliderRepo.CreateSliderItem(item); err != nil {
		return nil, err
	}

	return item.ToResponse(), nil
}

// GetSliderItemByID retrieves a slider item by its ID
func (s *sliderService) GetSliderItemByID(id uint) (*model.SliderItemResponse, error) {
	item, err := s.sliderRepo.GetSliderItemByID(id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("slider item not found")
	}

	return item.ToResponse(), nil
}

// GetSliderItemsBySlider retrieves slider items for a specific slider
func (s *sliderService) GetSliderItemsBySlider(sliderID uint, page, limit int) ([]model.SliderItemResponse, int64, error) {
	items, total, err := s.sliderRepo.GetSliderItemsBySlider(sliderID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.SliderItemResponse, len(items))
	for i, item := range items {
		responses[i] = *item.ToResponse()
	}

	return responses, total, nil
}

// UpdateSliderItem updates an existing slider item
func (s *sliderService) UpdateSliderItem(id uint, req *model.SliderItemUpdateRequest) (*model.SliderItemResponse, error) {
	item, err := s.sliderRepo.GetSliderItemByID(id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("slider item not found")
	}

	// Update fields if provided
	if req.Title != nil {
		item.Title = *req.Title
	}
	if req.Description != nil {
		item.Description = *req.Description
	}
	if req.ImageURL != nil {
		item.ImageURL = *req.ImageURL
	}
	if req.VideoURL != nil {
		item.VideoURL = *req.VideoURL
	}
	if req.TextContent != nil {
		item.TextContent = *req.TextContent
	}
	if req.ButtonText != nil {
		item.ButtonText = *req.ButtonText
	}
	if req.ButtonURL != nil {
		item.ButtonURL = *req.ButtonURL
	}
	if req.Width != nil {
		item.Width = *req.Width
	}
	if req.Height != nil {
		item.Height = *req.Height
	}
	if req.AltText != nil {
		item.AltText = *req.AltText
	}
	if req.CSSClass != nil {
		item.CSSClass = *req.CSSClass
	}
	if req.SortOrder != nil {
		item.SortOrder = *req.SortOrder
	}

	// Get slider to validate content
	slider, err := s.sliderRepo.GetSliderByID(item.SliderID)
	if err != nil {
		return nil, err
	}
	if slider == nil {
		return nil, errors.New("slider not found")
	}

	// Validate updated content
	if err := s.validateSliderItemContentUpdate(item, slider.Type); err != nil {
		return nil, err
	}

	if err := s.sliderRepo.UpdateSliderItem(item); err != nil {
		return nil, err
	}

	return item.ToResponse(), nil
}

// DeleteSliderItem soft deletes a slider item
func (s *sliderService) DeleteSliderItem(id uint) error {
	item, err := s.sliderRepo.GetSliderItemByID(id)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("slider item not found")
	}

	return s.sliderRepo.DeleteSliderItem(id)
}

// ReorderSliderItems reorders slider items
func (s *sliderService) ReorderSliderItems(sliderID uint, itemOrders map[uint]int) error {
	// Validate slider exists
	slider, err := s.sliderRepo.GetSliderByID(sliderID)
	if err != nil {
		return err
	}
	if slider == nil {
		return errors.New("slider not found")
	}

	return s.sliderRepo.ReorderSliderItems(sliderID, itemOrders)
}

// TrackSliderView tracks a slider view
func (s *sliderService) TrackSliderView(req *model.SliderViewRequest) error {
	return s.sliderRepo.TrackSliderView(req.SliderID, req.UserID, req.IP, req.UserAgent, req.Referrer)
}

// TrackSliderItemClick tracks a slider item click
func (s *sliderService) TrackSliderItemClick(req *model.SliderItemClickRequest) error {
	return s.sliderRepo.TrackSliderItemClick(req.ItemID, req.UserID, req.IP, req.UserAgent, req.Referrer)
}

// GetSliderStats retrieves slider statistics
func (s *sliderService) GetSliderStats() (*model.SliderStatsResponse, error) {
	return s.sliderRepo.GetSliderStats()
}

// GetSliderAnalytics retrieves analytics for a specific slider
func (s *sliderService) GetSliderAnalytics(sliderID uint, startDate, endDate *time.Time) (map[string]interface{}, error) {
	return s.sliderRepo.GetSliderAnalytics(sliderID, startDate, endDate)
}

// GetExpiredSliders retrieves expired sliders
func (s *sliderService) GetExpiredSliders() ([]model.SliderResponse, error) {
	sliders, err := s.sliderRepo.GetExpiredSliders()
	if err != nil {
		return nil, err
	}

	responses := make([]model.SliderResponse, len(sliders))
	for i, slider := range sliders {
		responses[i] = *slider.ToResponse()
	}

	return responses, nil
}

// GetSlidersToActivate retrieves sliders that should be activated
func (s *sliderService) GetSlidersToActivate() ([]model.SliderResponse, error) {
	sliders, err := s.sliderRepo.GetSlidersToActivate()
	if err != nil {
		return nil, err
	}

	responses := make([]model.SliderResponse, len(sliders))
	for i, slider := range sliders {
		responses[i] = *slider.ToResponse()
	}

	return responses, nil
}

// UpdateSliderStatus updates slider status
func (s *sliderService) UpdateSliderStatus(id uint, status model.SliderStatus) error {
	return s.sliderRepo.UpdateSliderStatus(id, status)
}

// validateSliderItemContent validates slider item content based on slider type
func (s *sliderService) validateSliderItemContent(req *model.SliderItemCreateRequest, sliderType model.SliderType) error {
	switch sliderType {
	case model.SliderTypeImage:
		if req.ImageURL == "" {
			return errors.New("image URL is required for image slider")
		}
	case model.SliderTypeVideo:
		if req.VideoURL == "" {
			return errors.New("video URL is required for video slider")
		}
	case model.SliderTypeMixed:
		// Mixed sliders can have any content type
		if req.ImageURL == "" && req.VideoURL == "" && req.TextContent == "" {
			return errors.New("at least one content type is required for mixed slider")
		}
	}

	return nil
}

// validateSliderItemContentUpdate validates updated slider item content
func (s *sliderService) validateSliderItemContentUpdate(item *model.SliderItem, sliderType model.SliderType) error {
	switch sliderType {
	case model.SliderTypeImage:
		if item.ImageURL == "" {
			return errors.New("image URL is required for image slider")
		}
	case model.SliderTypeVideo:
		if item.VideoURL == "" {
			return errors.New("video URL is required for video slider")
		}
	case model.SliderTypeMixed:
		if item.ImageURL == "" && item.VideoURL == "" && item.TextContent == "" {
			return errors.New("at least one content type is required for mixed slider")
		}
	}

	return nil
}
