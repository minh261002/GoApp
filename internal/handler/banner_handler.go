package handler

import (
	"net/http"
	"strconv"
	"time"

	"go_app/internal/model"
	"go_app/internal/service"
	"go_app/pkg/response"

	"github.com/gin-gonic/gin"
)

// BannerHandler handles banner-related HTTP requests
type BannerHandler struct {
	bannerService service.BannerService
}

// SliderHandler handles slider-related HTTP requests
type SliderHandler struct {
	sliderService service.SliderService
}

// NewBannerHandler creates a new BannerHandler
func NewBannerHandler() *BannerHandler {
	return &BannerHandler{
		bannerService: service.NewBannerService(),
	}
}

// NewSliderHandler creates a new SliderHandler
func NewSliderHandler() *SliderHandler {
	return &SliderHandler{
		sliderService: service.NewSliderService(),
	}
}

// Banner Handlers

// CreateBanner creates a new banner
func (h *BannerHandler) CreateBanner(c *gin.Context) {
	var req model.BannerCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	// Get creator ID from context
	creatorID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", "")
		return
	}

	banner, err := h.bannerService.CreateBanner(&req, creatorID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create banner", err.Error())
		return
	}

	response.Success(c, *banner, "Banner created successfully")
}

// GetBannerByID retrieves a banner by its ID
func (h *BannerHandler) GetBannerByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid banner ID", err.Error())
		return
	}

	banner, err := h.bannerService.GetBannerByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve banner", err.Error())
		return
	}

	response.Success(c, *banner, "Banner retrieved successfully")
}

// GetAllBanners retrieves all banners with pagination and filters
func (h *BannerHandler) GetAllBanners(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Build filters from query parameters
	filters := make(map[string]interface{})
	if typeParam := c.Query("type"); typeParam != "" {
		filters["type"] = typeParam
	}
	if position := c.Query("position"); position != "" {
		filters["position"] = position
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if audience := c.Query("target_audience"); audience != "" {
		filters["target_audience"] = audience
	}
	if deviceType := c.Query("device_type"); deviceType != "" {
		filters["device_type"] = deviceType
	}
	if location := c.Query("location"); location != "" {
		filters["location"] = location
	}
	if createdBy := c.Query("created_by"); createdBy != "" {
		if id, err := strconv.ParseUint(createdBy, 10, 32); err == nil {
			filters["created_by"] = uint(id)
		}
	}
	if startDate := c.Query("start_date"); startDate != "" {
		if date, err := time.Parse("2006-01-02", startDate); err == nil {
			filters["start_date"] = date
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if date, err := time.Parse("2006-01-02", endDate); err == nil {
			filters["end_date"] = date
		}
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	banners, total, err := h.bannerService.GetAllBanners(page, limit, filters)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve banners", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Banners retrieved successfully", banners, page, limit, total)
}

// UpdateBanner updates an existing banner
func (h *BannerHandler) UpdateBanner(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid banner ID", err.Error())
		return
	}

	var req model.BannerUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	banner, err := h.bannerService.UpdateBanner(uint(id), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update banner", err.Error())
		return
	}

	response.Success(c, *banner, "Banner updated successfully")
}

// DeleteBanner soft deletes a banner
func (h *BannerHandler) DeleteBanner(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid banner ID", err.Error())
		return
	}

	if err := h.bannerService.DeleteBanner(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete banner", err.Error())
		return
	}

	response.Success(c, nil, "Banner deleted successfully")
}

// GetActiveBanners retrieves active banners (public)
func (h *BannerHandler) GetActiveBanners(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	banners, total, err := h.bannerService.GetActiveBanners(page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve active banners", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Active banners retrieved successfully", banners, page, limit, total)
}

// GetBannersByType retrieves banners by type (public)
func (h *BannerHandler) GetBannersByType(c *gin.Context) {
	bannerType := c.Param("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	banners, total, err := h.bannerService.GetBannersByType(model.BannerType(bannerType), page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve banners by type", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Banners by type retrieved successfully", banners, page, limit, total)
}

// GetBannersByPosition retrieves banners by position (public)
func (h *BannerHandler) GetBannersByPosition(c *gin.Context) {
	position := c.Param("position")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	banners, total, err := h.bannerService.GetBannersByPosition(model.BannerPosition(position), page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve banners by position", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Banners by position retrieved successfully", banners, page, limit, total)
}

// GetBannersByStatus retrieves banners by status
func (h *BannerHandler) GetBannersByStatus(c *gin.Context) {
	status := c.Param("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	banners, total, err := h.bannerService.GetBannersByStatus(model.BannerStatus(status), page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve banners by status", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Banners by status retrieved successfully", banners, page, limit, total)
}

// SearchBanners performs full-text search on banners (public)
func (h *BannerHandler) SearchBanners(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.Error(c, http.StatusBadRequest, "Search query is required", "")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	banners, total, err := h.bannerService.SearchBanners(query, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to search banners", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Banner search completed successfully", banners, page, limit, total)
}

// GetBannersByTargetAudience retrieves banners by target audience (public)
func (h *BannerHandler) GetBannersByTargetAudience(c *gin.Context) {
	audience := c.Param("audience")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	banners, total, err := h.bannerService.GetBannersByTargetAudience(audience, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve banners by target audience", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Banners by target audience retrieved successfully", banners, page, limit, total)
}

// GetBannersByDeviceType retrieves banners by device type (public)
func (h *BannerHandler) GetBannersByDeviceType(c *gin.Context) {
	deviceType := c.Param("device_type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	banners, total, err := h.bannerService.GetBannersByDeviceType(deviceType, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve banners by device type", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Banners by device type retrieved successfully", banners, page, limit, total)
}

// TrackBannerClick tracks a banner click (public)
func (h *BannerHandler) TrackBannerClick(c *gin.Context) {
	var req model.BannerClickRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	// Get user ID from context if available
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = &[]uint{userID.(uint)}[0]
	}

	// Get IP and User-Agent from request
	req.IP = c.ClientIP()
	req.UserAgent = c.GetHeader("User-Agent")
	req.Referrer = c.GetHeader("Referer")

	if err := h.bannerService.TrackBannerClick(&req); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to track banner click", err.Error())
		return
	}

	response.Success(c, nil, "Banner click tracked successfully")
}

// TrackBannerView tracks a banner view (public)
func (h *BannerHandler) TrackBannerView(c *gin.Context) {
	var req model.BannerViewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	// Get user ID from context if available
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = &[]uint{userID.(uint)}[0]
	}

	// Get IP and User-Agent from request
	req.IP = c.ClientIP()
	req.UserAgent = c.GetHeader("User-Agent")
	req.Referrer = c.GetHeader("Referer")

	if err := h.bannerService.TrackBannerView(&req); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to track banner view", err.Error())
		return
	}

	response.Success(c, nil, "Banner view tracked successfully")
}

// GetBannerStats retrieves banner statistics
func (h *BannerHandler) GetBannerStats(c *gin.Context) {
	stats, err := h.bannerService.GetBannerStats()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve banner statistics", err.Error())
		return
	}

	response.Success(c, *stats, "Banner statistics retrieved successfully")
}

// GetBannerAnalytics retrieves analytics for a specific banner
func (h *BannerHandler) GetBannerAnalytics(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid banner ID", err.Error())
		return
	}

	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if date, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &date
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if date, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &date
		}
	}

	analytics, err := h.bannerService.GetBannerAnalytics(uint(id), startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve banner analytics", err.Error())
		return
	}

	response.Success(c, analytics, "Banner analytics retrieved successfully")
}

// GetExpiredBanners retrieves expired banners
func (h *BannerHandler) GetExpiredBanners(c *gin.Context) {
	banners, err := h.bannerService.GetExpiredBanners()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve expired banners", err.Error())
		return
	}

	response.Success(c, banners, "Expired banners retrieved successfully")
}

// GetBannersToActivate retrieves banners that should be activated
func (h *BannerHandler) GetBannersToActivate(c *gin.Context) {
	banners, err := h.bannerService.GetBannersToActivate()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve banners to activate", err.Error())
		return
	}

	response.Success(c, banners, "Banners to activate retrieved successfully")
}

// UpdateBannerStatus updates banner status
func (h *BannerHandler) UpdateBannerStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid banner ID", err.Error())
		return
	}

	var req struct {
		Status model.BannerStatus `json:"status" binding:"required,oneof=active inactive draft expired"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	if err := h.bannerService.UpdateBannerStatus(uint(id), req.Status); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update banner status", err.Error())
		return
	}

	response.Success(c, nil, "Banner status updated successfully")
}

// Slider Handlers

// CreateSlider creates a new slider
func (h *SliderHandler) CreateSlider(c *gin.Context) {
	var req model.SliderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	// Get creator ID from context
	creatorID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", "")
		return
	}

	slider, err := h.sliderService.CreateSlider(&req, creatorID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create slider", err.Error())
		return
	}

	response.Success(c, *slider, "Slider created successfully")
}

// GetSliderByID retrieves a slider by its ID
func (h *SliderHandler) GetSliderByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid slider ID", err.Error())
		return
	}

	slider, err := h.sliderService.GetSliderByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve slider", err.Error())
		return
	}

	response.Success(c, *slider, "Slider retrieved successfully")
}

// GetAllSliders retrieves all sliders with pagination and filters
func (h *SliderHandler) GetAllSliders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Build filters from query parameters
	filters := make(map[string]interface{})
	if typeParam := c.Query("type"); typeParam != "" {
		filters["type"] = typeParam
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if audience := c.Query("target_audience"); audience != "" {
		filters["target_audience"] = audience
	}
	if deviceType := c.Query("device_type"); deviceType != "" {
		filters["device_type"] = deviceType
	}
	if location := c.Query("location"); location != "" {
		filters["location"] = location
	}
	if createdBy := c.Query("created_by"); createdBy != "" {
		if id, err := strconv.ParseUint(createdBy, 10, 32); err == nil {
			filters["created_by"] = uint(id)
		}
	}
	if startDate := c.Query("start_date"); startDate != "" {
		if date, err := time.Parse("2006-01-02", startDate); err == nil {
			filters["start_date"] = date
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if date, err := time.Parse("2006-01-02", endDate); err == nil {
			filters["end_date"] = date
		}
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	sliders, total, err := h.sliderService.GetAllSliders(page, limit, filters)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve sliders", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Sliders retrieved successfully", sliders, page, limit, total)
}

// UpdateSlider updates an existing slider
func (h *SliderHandler) UpdateSlider(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid slider ID", err.Error())
		return
	}

	var req model.SliderUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	slider, err := h.sliderService.UpdateSlider(uint(id), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update slider", err.Error())
		return
	}

	response.Success(c, *slider, "Slider updated successfully")
}

// DeleteSlider soft deletes a slider
func (h *SliderHandler) DeleteSlider(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid slider ID", err.Error())
		return
	}

	if err := h.sliderService.DeleteSlider(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete slider", err.Error())
		return
	}

	response.Success(c, nil, "Slider deleted successfully")
}

// GetActiveSliders retrieves active sliders (public)
func (h *SliderHandler) GetActiveSliders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	sliders, total, err := h.sliderService.GetActiveSliders(page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve active sliders", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Active sliders retrieved successfully", sliders, page, limit, total)
}

// GetSlidersByType retrieves sliders by type (public)
func (h *SliderHandler) GetSlidersByType(c *gin.Context) {
	sliderType := c.Param("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	sliders, total, err := h.sliderService.GetSlidersByType(model.SliderType(sliderType), page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve sliders by type", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Sliders by type retrieved successfully", sliders, page, limit, total)
}

// GetSlidersByStatus retrieves sliders by status
func (h *SliderHandler) GetSlidersByStatus(c *gin.Context) {
	status := c.Param("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	sliders, total, err := h.sliderService.GetSlidersByStatus(model.SliderStatus(status), page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve sliders by status", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Sliders by status retrieved successfully", sliders, page, limit, total)
}

// SearchSliders performs full-text search on sliders (public)
func (h *SliderHandler) SearchSliders(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.Error(c, http.StatusBadRequest, "Search query is required", "")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	sliders, total, err := h.sliderService.SearchSliders(query, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to search sliders", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Slider search completed successfully", sliders, page, limit, total)
}

// GetSlidersByTargetAudience retrieves sliders by target audience (public)
func (h *SliderHandler) GetSlidersByTargetAudience(c *gin.Context) {
	audience := c.Param("audience")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	sliders, total, err := h.sliderService.GetSlidersByTargetAudience(audience, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve sliders by target audience", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Sliders by target audience retrieved successfully", sliders, page, limit, total)
}

// GetSlidersByDeviceType retrieves sliders by device type (public)
func (h *SliderHandler) GetSlidersByDeviceType(c *gin.Context) {
	deviceType := c.Param("device_type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	sliders, total, err := h.sliderService.GetSlidersByDeviceType(deviceType, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve sliders by device type", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Sliders by device type retrieved successfully", sliders, page, limit, total)
}

// Slider Items

// CreateSliderItem creates a new slider item
func (h *SliderHandler) CreateSliderItem(c *gin.Context) {
	var req model.SliderItemCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	item, err := h.sliderService.CreateSliderItem(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create slider item", err.Error())
		return
	}

	response.Success(c, *item, "Slider item created successfully")
}

// GetSliderItemByID retrieves a slider item by its ID
func (h *SliderHandler) GetSliderItemByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid slider item ID", err.Error())
		return
	}

	item, err := h.sliderService.GetSliderItemByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve slider item", err.Error())
		return
	}

	response.Success(c, *item, "Slider item retrieved successfully")
}

// GetSliderItemsBySlider retrieves slider items for a specific slider
func (h *SliderHandler) GetSliderItemsBySlider(c *gin.Context) {
	sliderIDStr := c.Param("slider_id")
	sliderID, err := strconv.ParseUint(sliderIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid slider ID", err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	items, total, err := h.sliderService.GetSliderItemsBySlider(uint(sliderID), page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve slider items", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Slider items retrieved successfully", items, page, limit, total)
}

// UpdateSliderItem updates an existing slider item
func (h *SliderHandler) UpdateSliderItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid slider item ID", err.Error())
		return
	}

	var req model.SliderItemUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	item, err := h.sliderService.UpdateSliderItem(uint(id), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update slider item", err.Error())
		return
	}

	response.Success(c, *item, "Slider item updated successfully")
}

// DeleteSliderItem soft deletes a slider item
func (h *SliderHandler) DeleteSliderItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid slider item ID", err.Error())
		return
	}

	if err := h.sliderService.DeleteSliderItem(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete slider item", err.Error())
		return
	}

	response.Success(c, nil, "Slider item deleted successfully")
}

// ReorderSliderItems reorders slider items
func (h *SliderHandler) ReorderSliderItems(c *gin.Context) {
	sliderIDStr := c.Param("slider_id")
	sliderID, err := strconv.ParseUint(sliderIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid slider ID", err.Error())
		return
	}

	var req struct {
		ItemOrders map[uint]int `json:"item_orders" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	if err := h.sliderService.ReorderSliderItems(uint(sliderID), req.ItemOrders); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to reorder slider items", err.Error())
		return
	}

	response.Success(c, nil, "Slider items reordered successfully")
}

// TrackSliderView tracks a slider view (public)
func (h *SliderHandler) TrackSliderView(c *gin.Context) {
	var req model.SliderViewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	// Get user ID from context if available
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = &[]uint{userID.(uint)}[0]
	}

	// Get IP and User-Agent from request
	req.IP = c.ClientIP()
	req.UserAgent = c.GetHeader("User-Agent")
	req.Referrer = c.GetHeader("Referer")

	if err := h.sliderService.TrackSliderView(&req); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to track slider view", err.Error())
		return
	}

	response.Success(c, nil, "Slider view tracked successfully")
}

// TrackSliderItemClick tracks a slider item click (public)
func (h *SliderHandler) TrackSliderItemClick(c *gin.Context) {
	var req model.SliderItemClickRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	// Get user ID from context if available
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = &[]uint{userID.(uint)}[0]
	}

	// Get IP and User-Agent from request
	req.IP = c.ClientIP()
	req.UserAgent = c.GetHeader("User-Agent")
	req.Referrer = c.GetHeader("Referer")

	if err := h.sliderService.TrackSliderItemClick(&req); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to track slider item click", err.Error())
		return
	}

	response.Success(c, nil, "Slider item click tracked successfully")
}

// GetSliderStats retrieves slider statistics
func (h *SliderHandler) GetSliderStats(c *gin.Context) {
	stats, err := h.sliderService.GetSliderStats()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve slider statistics", err.Error())
		return
	}

	response.Success(c, *stats, "Slider statistics retrieved successfully")
}

// GetSliderAnalytics retrieves analytics for a specific slider
func (h *SliderHandler) GetSliderAnalytics(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid slider ID", err.Error())
		return
	}

	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if date, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &date
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if date, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &date
		}
	}

	analytics, err := h.sliderService.GetSliderAnalytics(uint(id), startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve slider analytics", err.Error())
		return
	}

	response.Success(c, analytics, "Slider analytics retrieved successfully")
}

// GetExpiredSliders retrieves expired sliders
func (h *SliderHandler) GetExpiredSliders(c *gin.Context) {
	sliders, err := h.sliderService.GetExpiredSliders()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve expired sliders", err.Error())
		return
	}

	response.Success(c, sliders, "Expired sliders retrieved successfully")
}

// GetSlidersToActivate retrieves sliders that should be activated
func (h *SliderHandler) GetSlidersToActivate(c *gin.Context) {
	sliders, err := h.sliderService.GetSlidersToActivate()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve sliders to activate", err.Error())
		return
	}

	response.Success(c, sliders, "Sliders to activate retrieved successfully")
}

// UpdateSliderStatus updates slider status
func (h *SliderHandler) UpdateSliderStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid slider ID", err.Error())
		return
	}

	var req struct {
		Status model.SliderStatus `json:"status" binding:"required,oneof=active inactive draft"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	if err := h.sliderService.UpdateSliderStatus(uint(id), req.Status); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update slider status", err.Error())
		return
	}

	response.Success(c, nil, "Slider status updated successfully")
}
