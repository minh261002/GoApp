package handler

import (
	"net/http"
	"strconv"
	"time"

	"go_app/internal/model"
	"go_app/internal/service"
	"go_app/pkg/logger"
	"go_app/pkg/response"

	"github.com/gin-gonic/gin"
)

// AnalyticsHandler handles analytics-related HTTP requests
type AnalyticsHandler struct {
	analyticsService service.AnalyticsService
}

// NewAnalyticsHandler creates a new AnalyticsHandler
func NewAnalyticsHandler(analyticsService service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// Reports

// CreateReport creates a new analytics report
// @Summary Create analytics report
// @Description Create a new analytics report with specified parameters
// @Tags analytics
// @Accept json
// @Produce json
// @Param request body model.CreateAnalyticsReportRequest true "Report creation request"
// @Success 201 {object} response.Response{data=model.AnalyticsReportResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/analytics/reports [post]
func (h *AnalyticsHandler) CreateReport(c *gin.Context) {
	var req model.CreateAnalyticsReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	report, err := h.analyticsService.CreateReport(&req, userID.(uint))
	if err != nil {
		logger.Errorf("Failed to create analytics report: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create report", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Report created successfully", report)
}

// GetReportByID retrieves a report by ID
// @Summary Get analytics report
// @Description Get analytics report by ID
// @Tags analytics
// @Produce json
// @Param id path int true "Report ID"
// @Success 200 {object} response.Response{data=model.AnalyticsReportResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/analytics/reports/{id} [get]
func (h *AnalyticsHandler) GetReportByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid report ID", err.Error())
		return
	}

	report, err := h.analyticsService.GetReportByID(uint(id))
	if err != nil {
		logger.Errorf("Failed to get report by ID %d: %v", id, err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve report", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Report retrieved successfully", report)
}

// GetAllReports retrieves all reports with pagination and filters
// @Summary Get all analytics reports
// @Description Get all analytics reports with pagination and filters
// @Tags analytics
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param type query string false "Report type filter"
// @Param status query string false "Report status filter"
// @Param is_public query bool false "Public reports filter"
// @Param start_date query string false "Start date filter (YYYY-MM-DD)"
// @Param end_date query string false "End date filter (YYYY-MM-DD)"
// @Success 200 {object} response.Response{data=[]model.AnalyticsReportResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/analytics/reports [get]
func (h *AnalyticsHandler) GetAllReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Build filters
	filters := make(map[string]interface{})
	if reportType := c.Query("type"); reportType != "" {
		filters["type"] = reportType
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if isPublic := c.Query("is_public"); isPublic != "" {
		if isPublic == "true" {
			filters["is_public"] = true
		} else if isPublic == "false" {
			filters["is_public"] = false
		}
	}
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filters["start_date"] = startDate
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filters["end_date"] = endDate
		}
	}

	reports, total, err := h.analyticsService.GetAllReports(page, limit, filters)
	if err != nil {
		logger.Errorf("Failed to get all reports: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve reports", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Reports retrieved successfully", map[string]interface{}{
		"reports": reports,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// GetReportsByUser retrieves reports by user
// @Summary Get user analytics reports
// @Description Get analytics reports created by the authenticated user
// @Tags analytics
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param type query string false "Report type filter"
// @Param status query string false "Report status filter"
// @Success 200 {object} response.Response{data=[]model.AnalyticsReportResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/analytics/reports/my [get]
func (h *AnalyticsHandler) GetReportsByUser(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Build filters
	filters := make(map[string]interface{})
	if reportType := c.Query("type"); reportType != "" {
		filters["type"] = reportType
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	reports, total, err := h.analyticsService.GetReportsByUser(userID.(uint), page, limit, filters)
	if err != nil {
		logger.Errorf("Failed to get user reports: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve user reports", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "User reports retrieved successfully", map[string]interface{}{
		"reports": reports,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// UpdateReport updates a report
// @Summary Update analytics report
// @Description Update an existing analytics report
// @Tags analytics
// @Accept json
// @Produce json
// @Param id path int true "Report ID"
// @Param request body model.UpdateAnalyticsReportRequest true "Report update request"
// @Success 200 {object} response.Response{data=model.AnalyticsReportResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/analytics/reports/{id} [put]
func (h *AnalyticsHandler) UpdateReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid report ID", err.Error())
		return
	}

	var req model.UpdateAnalyticsReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	report, err := h.analyticsService.UpdateReport(uint(id), &req, userID.(uint))
	if err != nil {
		logger.Errorf("Failed to update report %d: %v", id, err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update report", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Report updated successfully", report)
}

// DeleteReport deletes a report
// @Summary Delete analytics report
// @Description Delete an analytics report
// @Tags analytics
// @Param id path int true "Report ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/analytics/reports/{id} [delete]
func (h *AnalyticsHandler) DeleteReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid report ID", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	err = h.analyticsService.DeleteReport(uint(id), userID.(uint))
	if err != nil {
		logger.Errorf("Failed to delete report %d: %v", id, err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete report", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Report deleted successfully", nil)
}

// GenerateReport generates a report
// @Summary Generate analytics report
// @Description Generate analytics report data
// @Tags analytics
// @Param id path int true "Report ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/analytics/reports/{id}/generate [post]
func (h *AnalyticsHandler) GenerateReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid report ID", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	err = h.analyticsService.GenerateReport(uint(id), userID.(uint))
	if err != nil {
		logger.Errorf("Failed to generate report %d: %v", id, err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate report", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Report generation started", nil)
}

// Dashboards

// CreateDashboard creates a new analytics dashboard
// @Summary Create analytics dashboard
// @Description Create a new analytics dashboard
// @Tags analytics
// @Accept json
// @Produce json
// @Param request body model.CreateAnalyticsDashboardRequest true "Dashboard creation request"
// @Success 201 {object} response.Response{data=model.AnalyticsDashboardResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/analytics/dashboards [post]
func (h *AnalyticsHandler) CreateDashboard(c *gin.Context) {
	var req model.CreateAnalyticsDashboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	dashboard, err := h.analyticsService.CreateDashboard(&req, userID.(uint))
	if err != nil {
		logger.Errorf("Failed to create analytics dashboard: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create dashboard", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Dashboard created successfully", dashboard)
}

// GetDashboardByID retrieves a dashboard by ID
// @Summary Get analytics dashboard
// @Description Get analytics dashboard by ID
// @Tags analytics
// @Produce json
// @Param id path int true "Dashboard ID"
// @Success 200 {object} response.Response{data=model.AnalyticsDashboardResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/analytics/dashboards/{id} [get]
func (h *AnalyticsHandler) GetDashboardByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid dashboard ID", err.Error())
		return
	}

	dashboard, err := h.analyticsService.GetDashboardByID(uint(id))
	if err != nil {
		logger.Errorf("Failed to get dashboard by ID %d: %v", id, err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve dashboard", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Dashboard retrieved successfully", dashboard)
}

// GetAllDashboards retrieves all dashboards with pagination and filters
// @Summary Get all analytics dashboards
// @Description Get all analytics dashboards with pagination and filters
// @Tags analytics
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param is_public query bool false "Public dashboards filter"
// @Success 200 {object} response.Response{data=[]model.AnalyticsDashboardResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/analytics/dashboards [get]
func (h *AnalyticsHandler) GetAllDashboards(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Build filters
	filters := make(map[string]interface{})
	if isPublic := c.Query("is_public"); isPublic != "" {
		if isPublic == "true" {
			filters["is_public"] = true
		} else if isPublic == "false" {
			filters["is_public"] = false
		}
	}

	dashboards, total, err := h.analyticsService.GetAllDashboards(page, limit, filters)
	if err != nil {
		logger.Errorf("Failed to get all dashboards: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve dashboards", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Dashboards retrieved successfully", map[string]interface{}{
		"dashboards": dashboards,
		"total":      total,
		"page":       page,
		"limit":      limit,
	})
}

// GetPublicDashboards retrieves public dashboards
// @Summary Get public analytics dashboards
// @Description Get public analytics dashboards
// @Tags analytics
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.Response{data=[]model.AnalyticsDashboardResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/analytics/dashboards/public [get]
func (h *AnalyticsHandler) GetPublicDashboards(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	dashboards, total, err := h.analyticsService.GetPublicDashboards(page, limit)
	if err != nil {
		logger.Errorf("Failed to get public dashboards: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve public dashboards", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Public dashboards retrieved successfully", map[string]interface{}{
		"dashboards": dashboards,
		"total":      total,
		"page":       page,
		"limit":      limit,
	})
}

// Analytics Data

// GetSalesAnalytics retrieves sales analytics data
// @Summary Get sales analytics
// @Description Get sales analytics data for the specified period
// @Tags analytics
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Param status query string false "Order status filter"
// @Success 200 {object} response.Response{data=model.SalesAnalytics}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/analytics/sales [get]
func (h *AnalyticsHandler) GetSalesAnalytics(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Missing required parameters", "start_date and end_date are required")
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid start date format", "start_date must be in YYYY-MM-DD format")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid end date format", "end_date must be in YYYY-MM-DD format")
		return
	}

	// Build filters
	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	analytics, err := h.analyticsService.GetSalesAnalytics(startDate, endDate, filters)
	if err != nil {
		logger.Errorf("Failed to get sales analytics: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve sales analytics", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Sales analytics retrieved successfully", analytics)
}

// GetTrafficAnalytics retrieves traffic analytics data
// @Summary Get traffic analytics
// @Description Get traffic analytics data for the specified period
// @Tags analytics
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} response.Response{data=model.TrafficAnalytics}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/analytics/traffic [get]
func (h *AnalyticsHandler) GetTrafficAnalytics(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Missing required parameters", "start_date and end_date are required")
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid start date format", "start_date must be in YYYY-MM-DD format")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid end date format", "end_date must be in YYYY-MM-DD format")
		return
	}

	analytics, err := h.analyticsService.GetTrafficAnalytics(startDate, endDate, map[string]interface{}{})
	if err != nil {
		logger.Errorf("Failed to get traffic analytics: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve traffic analytics", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Traffic analytics retrieved successfully", analytics)
}

// GetUserAnalytics retrieves user analytics data
// @Summary Get user analytics
// @Description Get user analytics data for the specified period
// @Tags analytics
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} response.Response{data=model.UserAnalytics}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/analytics/users [get]
func (h *AnalyticsHandler) GetUserAnalytics(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Missing required parameters", "start_date and end_date are required")
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid start date format", "start_date must be in YYYY-MM-DD format")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid end date format", "end_date must be in YYYY-MM-DD format")
		return
	}

	analytics, err := h.analyticsService.GetUserAnalytics(startDate, endDate, map[string]interface{}{})
	if err != nil {
		logger.Errorf("Failed to get user analytics: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve user analytics", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "User analytics retrieved successfully", analytics)
}

// GetInventoryAnalytics retrieves inventory analytics data
// @Summary Get inventory analytics
// @Description Get inventory analytics data for the specified period
// @Tags analytics
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} response.Response{data=model.InventoryAnalytics}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/analytics/inventory [get]
func (h *AnalyticsHandler) GetInventoryAnalytics(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Missing required parameters", "start_date and end_date are required")
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid start date format", "start_date must be in YYYY-MM-DD format")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid end date format", "end_date must be in YYYY-MM-DD format")
		return
	}

	analytics, err := h.analyticsService.GetInventoryAnalytics(startDate, endDate, map[string]interface{}{})
	if err != nil {
		logger.Errorf("Failed to get inventory analytics: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve inventory analytics", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Inventory analytics retrieved successfully", analytics)
}

// GetAnalyticsSummary retrieves a comprehensive analytics summary
// @Summary Get analytics summary
// @Description Get a comprehensive analytics summary for the specified period
// @Tags analytics
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/analytics/summary [get]
func (h *AnalyticsHandler) GetAnalyticsSummary(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Missing required parameters", "start_date and end_date are required")
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid start date format", "start_date must be in YYYY-MM-DD format")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid end date format", "end_date must be in YYYY-MM-DD format")
		return
	}

	summary, err := h.analyticsService.GetAnalyticsSummary(startDate, endDate)
	if err != nil {
		logger.Errorf("Failed to get analytics summary: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve analytics summary", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Analytics summary retrieved successfully", summary)
}

// TrackEvent tracks an analytics event
// @Summary Track analytics event
// @Description Track an analytics event
// @Tags analytics
// @Accept json
// @Produce json
// @Param event body model.AnalyticsEvent true "Event data"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/analytics/events [post]
func (h *AnalyticsHandler) TrackEvent(c *gin.Context) {
	var event model.AnalyticsEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Set request context
	event.IPAddress = c.ClientIP()
	event.UserAgent = c.GetHeader("User-Agent")
	event.Referer = c.GetHeader("Referer")

	// Get user ID from context if available
	if userID, exists := c.Get("user_id"); exists {
		userIDUint := userID.(uint)
		event.UserID = &userIDUint
	}

	err := h.analyticsService.TrackEvent(&event)
	if err != nil {
		logger.Errorf("Failed to track analytics event: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to track event", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Event tracked successfully", nil)
}
