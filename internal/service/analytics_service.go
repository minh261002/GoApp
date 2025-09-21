package service

import (
	"fmt"
	"go_app/internal/model"
	"go_app/internal/repository"
	"go_app/pkg/logger"
	"time"
)

// AnalyticsService defines methods for analytics business logic
type AnalyticsService interface {
	// Reports
	CreateReport(req *model.CreateAnalyticsReportRequest, userID uint) (*model.AnalyticsReportResponse, error)
	GetReportByID(id uint) (*model.AnalyticsReportResponse, error)
	GetAllReports(page, limit int, filters map[string]interface{}) ([]model.AnalyticsReportResponse, int64, error)
	GetReportsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.AnalyticsReportResponse, int64, error)
	UpdateReport(id uint, req *model.UpdateAnalyticsReportRequest, userID uint) (*model.AnalyticsReportResponse, error)
	DeleteReport(id uint, userID uint) error
	GenerateReport(id uint, userID uint) error

	// Dashboards
	CreateDashboard(req *model.CreateAnalyticsDashboardRequest, userID uint) (*model.AnalyticsDashboardResponse, error)
	GetDashboardByID(id uint) (*model.AnalyticsDashboardResponse, error)
	GetAllDashboards(page, limit int, filters map[string]interface{}) ([]model.AnalyticsDashboardResponse, int64, error)
	GetDashboardsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.AnalyticsDashboardResponse, int64, error)
	GetPublicDashboards(page, limit int) ([]model.AnalyticsDashboardResponse, int64, error)
	UpdateDashboard(id uint, req *model.UpdateAnalyticsDashboardRequest, userID uint) (*model.AnalyticsDashboardResponse, error)
	DeleteDashboard(id uint, userID uint) error

	// Widgets
	CreateWidget(req *model.CreateAnalyticsWidgetRequest, userID uint) (*model.AnalyticsWidgetResponse, error)
	GetWidgetByID(id uint) (*model.AnalyticsWidgetResponse, error)
	GetWidgetsByDashboard(dashboardID uint) ([]model.AnalyticsWidgetResponse, error)
	UpdateWidget(id uint, req *model.UpdateAnalyticsWidgetRequest, userID uint) (*model.AnalyticsWidgetResponse, error)
	DeleteWidget(id uint, userID uint) error

	// Events
	TrackEvent(event *model.AnalyticsEvent) error
	GetEvents(page, limit int, filters map[string]interface{}) ([]model.AnalyticsEventResponse, int64, error)
	GetEventsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.AnalyticsEventResponse, int64, error)
	GetEventsByType(eventType string, page, limit int, filters map[string]interface{}) ([]model.AnalyticsEventResponse, int64, error)
	GetEventsByEntity(entityType string, entityID uint, page, limit int) ([]model.AnalyticsEventResponse, int64, error)

	// Analytics Data
	GetSalesAnalytics(startDate, endDate time.Time, filters map[string]interface{}) (*model.SalesAnalytics, error)
	GetTrafficAnalytics(startDate, endDate time.Time, filters map[string]interface{}) (*model.TrafficAnalytics, error)
	GetUserAnalytics(startDate, endDate time.Time, filters map[string]interface{}) (*model.UserAnalytics, error)
	GetInventoryAnalytics(startDate, endDate time.Time, filters map[string]interface{}) (*model.InventoryAnalytics, error)
	GetProductAnalytics(startDate, endDate time.Time, filters map[string]interface{}) (map[string]interface{}, error)
	GetOrderAnalytics(startDate, endDate time.Time, filters map[string]interface{}) (map[string]interface{}, error)

	// Event Analytics
	GetEventStats(startDate, endDate time.Time, eventType string) (map[string]interface{}, error)
	GetTopEvents(startDate, endDate time.Time, limit int) ([]map[string]interface{}, error)
	GetEventTrends(startDate, endDate time.Time, eventType string) ([]model.PeriodData, error)

	// Custom Analytics
	ExecuteCustomQuery(query string, params []interface{}) ([]map[string]interface{}, error)
	GetAnalyticsSummary(startDate, endDate time.Time) (map[string]interface{}, error)
}

// analyticsService implements AnalyticsService
type analyticsService struct {
	analyticsRepo repository.AnalyticsRepository
	userRepo      repository.UserRepository
}

// NewAnalyticsService creates a new AnalyticsService
func NewAnalyticsService(analyticsRepo repository.AnalyticsRepository, userRepo repository.UserRepository) AnalyticsService {
	return &analyticsService{
		analyticsRepo: analyticsRepo,
		userRepo:      userRepo,
	}
}

// Reports

// CreateReport creates a new analytics report
func (s *analyticsService) CreateReport(req *model.CreateAnalyticsReportRequest, userID uint) (*model.AnalyticsReportResponse, error) {
	// Validate dates
	if req.EndDate.Before(req.StartDate) {
		return nil, fmt.Errorf("end date must be after start date")
	}

	// Create report
	report := &model.AnalyticsReport{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Period:      req.Period,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Filters:     fmt.Sprintf("%v", req.Filters), // Simplified JSON conversion
		Status:      model.ReportStatusPending,
		IsScheduled: req.IsScheduled,
		IsPublic:    req.IsPublic,
		CreatedBy:   userID,
	}

	if err := s.analyticsRepo.CreateReport(report); err != nil {
		logger.Errorf("Failed to create analytics report: %v", err)
		return nil, fmt.Errorf("failed to create report")
	}

	// Get created report with relations
	createdReport, err := s.analyticsRepo.GetReportByID(report.ID)
	if err != nil {
		logger.Errorf("Failed to get created report: %v", err)
		return nil, fmt.Errorf("failed to retrieve created report")
	}

	return s.toReportResponse(createdReport), nil
}

// GetReportByID retrieves a report by ID
func (s *analyticsService) GetReportByID(id uint) (*model.AnalyticsReportResponse, error) {
	report, err := s.analyticsRepo.GetReportByID(id)
	if err != nil {
		logger.Errorf("Failed to get report by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve report")
	}
	if report == nil {
		return nil, fmt.Errorf("report not found")
	}

	return s.toReportResponse(report), nil
}

// GetAllReports retrieves all reports with pagination and filters
func (s *analyticsService) GetAllReports(page, limit int, filters map[string]interface{}) ([]model.AnalyticsReportResponse, int64, error) {
	reports, total, err := s.analyticsRepo.GetAllReports(page, limit, filters)
	if err != nil {
		logger.Errorf("Failed to get all reports: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve reports")
	}

	var responses []model.AnalyticsReportResponse
	for _, report := range reports {
		responses = append(responses, *s.toReportResponse(&report))
	}

	return responses, total, nil
}

// GetReportsByUser retrieves reports by user
func (s *analyticsService) GetReportsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.AnalyticsReportResponse, int64, error) {
	reports, total, err := s.analyticsRepo.GetReportsByUser(userID, page, limit, filters)
	if err != nil {
		logger.Errorf("Failed to get reports by user %d: %v", userID, err)
		return nil, 0, fmt.Errorf("failed to retrieve user reports")
	}

	var responses []model.AnalyticsReportResponse
	for _, report := range reports {
		responses = append(responses, *s.toReportResponse(&report))
	}

	return responses, total, nil
}

// UpdateReport updates a report
func (s *analyticsService) UpdateReport(id uint, req *model.UpdateAnalyticsReportRequest, userID uint) (*model.AnalyticsReportResponse, error) {
	report, err := s.analyticsRepo.GetReportByID(id)
	if err != nil {
		logger.Errorf("Failed to get report by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve report")
	}
	if report == nil {
		return nil, fmt.Errorf("report not found")
	}

	// Check ownership
	if report.CreatedBy != userID {
		return nil, fmt.Errorf("unauthorized to update this report")
	}

	// Update fields
	if req.Name != "" {
		report.Name = req.Name
	}
	if req.Description != "" {
		report.Description = req.Description
	}
	if req.Type != "" {
		report.Type = req.Type
	}
	if req.Period != "" {
		report.Period = req.Period
	}
	if req.StartDate != nil {
		report.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		report.EndDate = *req.EndDate
	}
	if req.Filters != nil {
		report.Filters = fmt.Sprintf("%v", req.Filters)
	}
	if req.IsScheduled != nil {
		report.IsScheduled = *req.IsScheduled
	}
	if req.IsPublic != nil {
		report.IsPublic = *req.IsPublic
	}

	if err := s.analyticsRepo.UpdateReport(report); err != nil {
		logger.Errorf("Failed to update report %d: %v", id, err)
		return nil, fmt.Errorf("failed to update report")
	}

	return s.toReportResponse(report), nil
}

// DeleteReport deletes a report
func (s *analyticsService) DeleteReport(id uint, userID uint) error {
	report, err := s.analyticsRepo.GetReportByID(id)
	if err != nil {
		logger.Errorf("Failed to get report by ID %d: %v", id, err)
		return fmt.Errorf("failed to retrieve report")
	}
	if report == nil {
		return fmt.Errorf("report not found")
	}

	// Check ownership
	if report.CreatedBy != userID {
		return fmt.Errorf("unauthorized to delete this report")
	}

	if err := s.analyticsRepo.DeleteReport(id); err != nil {
		logger.Errorf("Failed to delete report %d: %v", id, err)
		return fmt.Errorf("failed to delete report")
	}

	return nil
}

// GenerateReport generates a report
func (s *analyticsService) GenerateReport(id uint, userID uint) error {
	report, err := s.analyticsRepo.GetReportByID(id)
	if err != nil {
		logger.Errorf("Failed to get report by ID %d: %v", id, err)
		return fmt.Errorf("failed to retrieve report")
	}
	if report == nil {
		return fmt.Errorf("report not found")
	}

	// Check ownership
	if report.CreatedBy != userID {
		return fmt.Errorf("unauthorized to generate this report")
	}

	if err := s.analyticsRepo.GenerateReport(report); err != nil {
		logger.Errorf("Failed to generate report %d: %v", id, err)
		return fmt.Errorf("failed to generate report")
	}

	return nil
}

// Dashboards

// CreateDashboard creates a new analytics dashboard
func (s *analyticsService) CreateDashboard(req *model.CreateAnalyticsDashboardRequest, userID uint) (*model.AnalyticsDashboardResponse, error) {
	dashboard := &model.AnalyticsDashboard{
		Name:        req.Name,
		Description: req.Description,
		Layout:      fmt.Sprintf("%v", req.Layout), // Simplified JSON conversion
		IsPublic:    req.IsPublic,
		UserID:      &userID,
	}

	if err := s.analyticsRepo.CreateDashboard(dashboard); err != nil {
		logger.Errorf("Failed to create analytics dashboard: %v", err)
		return nil, fmt.Errorf("failed to create dashboard")
	}

	// Get created dashboard with relations
	createdDashboard, err := s.analyticsRepo.GetDashboardByID(dashboard.ID)
	if err != nil {
		logger.Errorf("Failed to get created dashboard: %v", err)
		return nil, fmt.Errorf("failed to retrieve created dashboard")
	}

	return s.toDashboardResponse(createdDashboard), nil
}

// GetDashboardByID retrieves a dashboard by ID
func (s *analyticsService) GetDashboardByID(id uint) (*model.AnalyticsDashboardResponse, error) {
	dashboard, err := s.analyticsRepo.GetDashboardByID(id)
	if err != nil {
		logger.Errorf("Failed to get dashboard by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve dashboard")
	}
	if dashboard == nil {
		return nil, fmt.Errorf("dashboard not found")
	}

	return s.toDashboardResponse(dashboard), nil
}

// GetAllDashboards retrieves all dashboards with pagination and filters
func (s *analyticsService) GetAllDashboards(page, limit int, filters map[string]interface{}) ([]model.AnalyticsDashboardResponse, int64, error) {
	dashboards, total, err := s.analyticsRepo.GetAllDashboards(page, limit, filters)
	if err != nil {
		logger.Errorf("Failed to get all dashboards: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve dashboards")
	}

	var responses []model.AnalyticsDashboardResponse
	for _, dashboard := range dashboards {
		responses = append(responses, *s.toDashboardResponse(&dashboard))
	}

	return responses, total, nil
}

// GetDashboardsByUser retrieves dashboards by user
func (s *analyticsService) GetDashboardsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.AnalyticsDashboardResponse, int64, error) {
	dashboards, total, err := s.analyticsRepo.GetDashboardsByUser(userID, page, limit, filters)
	if err != nil {
		logger.Errorf("Failed to get dashboards by user %d: %v", userID, err)
		return nil, 0, fmt.Errorf("failed to retrieve user dashboards")
	}

	var responses []model.AnalyticsDashboardResponse
	for _, dashboard := range dashboards {
		responses = append(responses, *s.toDashboardResponse(&dashboard))
	}

	return responses, total, nil
}

// GetPublicDashboards retrieves public dashboards
func (s *analyticsService) GetPublicDashboards(page, limit int) ([]model.AnalyticsDashboardResponse, int64, error) {
	dashboards, total, err := s.analyticsRepo.GetPublicDashboards(page, limit)
	if err != nil {
		logger.Errorf("Failed to get public dashboards: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve public dashboards")
	}

	var responses []model.AnalyticsDashboardResponse
	for _, dashboard := range dashboards {
		responses = append(responses, *s.toDashboardResponse(&dashboard))
	}

	return responses, total, nil
}

// UpdateDashboard updates a dashboard
func (s *analyticsService) UpdateDashboard(id uint, req *model.UpdateAnalyticsDashboardRequest, userID uint) (*model.AnalyticsDashboardResponse, error) {
	dashboard, err := s.analyticsRepo.GetDashboardByID(id)
	if err != nil {
		logger.Errorf("Failed to get dashboard by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve dashboard")
	}
	if dashboard == nil {
		return nil, fmt.Errorf("dashboard not found")
	}

	// Check ownership
	if dashboard.UserID == nil || *dashboard.UserID != userID {
		return nil, fmt.Errorf("unauthorized to update this dashboard")
	}

	// Update fields
	if req.Name != "" {
		dashboard.Name = req.Name
	}
	if req.Description != "" {
		dashboard.Description = req.Description
	}
	if req.Layout != nil {
		dashboard.Layout = fmt.Sprintf("%v", req.Layout)
	}
	if req.IsPublic != nil {
		dashboard.IsPublic = *req.IsPublic
	}

	if err := s.analyticsRepo.UpdateDashboard(dashboard); err != nil {
		logger.Errorf("Failed to update dashboard %d: %v", id, err)
		return nil, fmt.Errorf("failed to update dashboard")
	}

	return s.toDashboardResponse(dashboard), nil
}

// DeleteDashboard deletes a dashboard
func (s *analyticsService) DeleteDashboard(id uint, userID uint) error {
	dashboard, err := s.analyticsRepo.GetDashboardByID(id)
	if err != nil {
		logger.Errorf("Failed to get dashboard by ID %d: %v", id, err)
		return fmt.Errorf("failed to retrieve dashboard")
	}
	if dashboard == nil {
		return fmt.Errorf("dashboard not found")
	}

	// Check ownership
	if dashboard.UserID == nil || *dashboard.UserID != userID {
		return fmt.Errorf("unauthorized to delete this dashboard")
	}

	if err := s.analyticsRepo.DeleteDashboard(id); err != nil {
		logger.Errorf("Failed to delete dashboard %d: %v", id, err)
		return fmt.Errorf("failed to delete dashboard")
	}

	return nil
}

// Widgets

// CreateWidget creates a new analytics widget
func (s *analyticsService) CreateWidget(req *model.CreateAnalyticsWidgetRequest, userID uint) (*model.AnalyticsWidgetResponse, error) {
	// Verify dashboard ownership
	dashboard, err := s.analyticsRepo.GetDashboardByID(req.DashboardID)
	if err != nil {
		logger.Errorf("Failed to get dashboard by ID %d: %v", req.DashboardID, err)
		return nil, fmt.Errorf("failed to retrieve dashboard")
	}
	if dashboard == nil {
		return nil, fmt.Errorf("dashboard not found")
	}

	// Check ownership
	if dashboard.UserID == nil || *dashboard.UserID != userID {
		return nil, fmt.Errorf("unauthorized to create widget for this dashboard")
	}

	widget := &model.AnalyticsWidget{
		DashboardID: req.DashboardID,
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		Config:      fmt.Sprintf("%v", req.Config),   // Simplified JSON conversion
		Data:        fmt.Sprintf("%v", req.Data),     // Simplified JSON conversion
		Position:    fmt.Sprintf("%v", req.Position), // Simplified JSON conversion
		IsActive:    true,
	}

	if err := s.analyticsRepo.CreateWidget(widget); err != nil {
		logger.Errorf("Failed to create analytics widget: %v", err)
		return nil, fmt.Errorf("failed to create widget")
	}

	// Get created widget with relations
	createdWidget, err := s.analyticsRepo.GetWidgetByID(widget.ID)
	if err != nil {
		logger.Errorf("Failed to get created widget: %v", err)
		return nil, fmt.Errorf("failed to retrieve created widget")
	}

	return s.toWidgetResponse(createdWidget), nil
}

// GetWidgetByID retrieves a widget by ID
func (s *analyticsService) GetWidgetByID(id uint) (*model.AnalyticsWidgetResponse, error) {
	widget, err := s.analyticsRepo.GetWidgetByID(id)
	if err != nil {
		logger.Errorf("Failed to get widget by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve widget")
	}
	if widget == nil {
		return nil, fmt.Errorf("widget not found")
	}

	return s.toWidgetResponse(widget), nil
}

// GetWidgetsByDashboard retrieves widgets by dashboard ID
func (s *analyticsService) GetWidgetsByDashboard(dashboardID uint) ([]model.AnalyticsWidgetResponse, error) {
	widgets, err := s.analyticsRepo.GetWidgetsByDashboard(dashboardID)
	if err != nil {
		logger.Errorf("Failed to get widgets by dashboard %d: %v", dashboardID, err)
		return nil, fmt.Errorf("failed to retrieve dashboard widgets")
	}

	var responses []model.AnalyticsWidgetResponse
	for _, widget := range widgets {
		responses = append(responses, *s.toWidgetResponse(&widget))
	}

	return responses, nil
}

// UpdateWidget updates a widget
func (s *analyticsService) UpdateWidget(id uint, req *model.UpdateAnalyticsWidgetRequest, userID uint) (*model.AnalyticsWidgetResponse, error) {
	widget, err := s.analyticsRepo.GetWidgetByID(id)
	if err != nil {
		logger.Errorf("Failed to get widget by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve widget")
	}
	if widget == nil {
		return nil, fmt.Errorf("widget not found")
	}

	// Verify dashboard ownership
	dashboard, err := s.analyticsRepo.GetDashboardByID(widget.DashboardID)
	if err != nil {
		logger.Errorf("Failed to get dashboard by ID %d: %v", widget.DashboardID, err)
		return nil, fmt.Errorf("failed to retrieve dashboard")
	}
	if dashboard == nil {
		return nil, fmt.Errorf("dashboard not found")
	}

	// Check ownership
	if dashboard.UserID == nil || *dashboard.UserID != userID {
		return nil, fmt.Errorf("unauthorized to update this widget")
	}

	// Update fields
	if req.Title != "" {
		widget.Title = req.Title
	}
	if req.Description != "" {
		widget.Description = req.Description
	}
	if req.Config != nil {
		widget.Config = fmt.Sprintf("%v", req.Config)
	}
	if req.Data != nil {
		widget.Data = fmt.Sprintf("%v", req.Data)
	}
	if req.Position != nil {
		widget.Position = fmt.Sprintf("%v", req.Position)
	}
	if req.IsActive != nil {
		widget.IsActive = *req.IsActive
	}

	if err := s.analyticsRepo.UpdateWidget(widget); err != nil {
		logger.Errorf("Failed to update widget %d: %v", id, err)
		return nil, fmt.Errorf("failed to update widget")
	}

	return s.toWidgetResponse(widget), nil
}

// DeleteWidget deletes a widget
func (s *analyticsService) DeleteWidget(id uint, userID uint) error {
	widget, err := s.analyticsRepo.GetWidgetByID(id)
	if err != nil {
		logger.Errorf("Failed to get widget by ID %d: %v", id, err)
		return fmt.Errorf("failed to retrieve widget")
	}
	if widget == nil {
		return fmt.Errorf("widget not found")
	}

	// Verify dashboard ownership
	dashboard, err := s.analyticsRepo.GetDashboardByID(widget.DashboardID)
	if err != nil {
		logger.Errorf("Failed to get dashboard by ID %d: %v", widget.DashboardID, err)
		return fmt.Errorf("failed to retrieve dashboard")
	}
	if dashboard == nil {
		return fmt.Errorf("dashboard not found")
	}

	// Check ownership
	if dashboard.UserID == nil || *dashboard.UserID != userID {
		return fmt.Errorf("unauthorized to delete this widget")
	}

	if err := s.analyticsRepo.DeleteWidget(id); err != nil {
		logger.Errorf("Failed to delete widget %d: %v", id, err)
		return fmt.Errorf("failed to delete widget")
	}

	return nil
}

// Events

// TrackEvent tracks an analytics event
func (s *analyticsService) TrackEvent(event *model.AnalyticsEvent) error {
	if err := s.analyticsRepo.CreateEvent(event); err != nil {
		logger.Errorf("Failed to track analytics event: %v", err)
		return fmt.Errorf("failed to track event")
	}
	return nil
}

// GetEvents retrieves events with pagination and filters
func (s *analyticsService) GetEvents(page, limit int, filters map[string]interface{}) ([]model.AnalyticsEventResponse, int64, error) {
	events, total, err := s.analyticsRepo.GetEvents(page, limit, filters)
	if err != nil {
		logger.Errorf("Failed to get events: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve events")
	}

	var responses []model.AnalyticsEventResponse
	for _, event := range events {
		responses = append(responses, *s.toEventResponse(&event))
	}

	return responses, total, nil
}

// GetEventsByUser retrieves events by user
func (s *analyticsService) GetEventsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.AnalyticsEventResponse, int64, error) {
	events, total, err := s.analyticsRepo.GetEventsByUser(userID, page, limit, filters)
	if err != nil {
		logger.Errorf("Failed to get events by user %d: %v", userID, err)
		return nil, 0, fmt.Errorf("failed to retrieve user events")
	}

	var responses []model.AnalyticsEventResponse
	for _, event := range events {
		responses = append(responses, *s.toEventResponse(&event))
	}

	return responses, total, nil
}

// GetEventsByType retrieves events by type
func (s *analyticsService) GetEventsByType(eventType string, page, limit int, filters map[string]interface{}) ([]model.AnalyticsEventResponse, int64, error) {
	events, total, err := s.analyticsRepo.GetEventsByType(eventType, page, limit, filters)
	if err != nil {
		logger.Errorf("Failed to get events by type %s: %v", eventType, err)
		return nil, 0, fmt.Errorf("failed to retrieve events by type")
	}

	var responses []model.AnalyticsEventResponse
	for _, event := range events {
		responses = append(responses, *s.toEventResponse(&event))
	}

	return responses, total, nil
}

// GetEventsByEntity retrieves events by entity
func (s *analyticsService) GetEventsByEntity(entityType string, entityID uint, page, limit int) ([]model.AnalyticsEventResponse, int64, error) {
	events, total, err := s.analyticsRepo.GetEventsByEntity(entityType, entityID, page, limit)
	if err != nil {
		logger.Errorf("Failed to get events by entity %s %d: %v", entityType, entityID, err)
		return nil, 0, fmt.Errorf("failed to retrieve events by entity")
	}

	var responses []model.AnalyticsEventResponse
	for _, event := range events {
		responses = append(responses, *s.toEventResponse(&event))
	}

	return responses, total, nil
}

// Analytics Data

// GetSalesAnalytics retrieves sales analytics data
func (s *analyticsService) GetSalesAnalytics(startDate, endDate time.Time, filters map[string]interface{}) (*model.SalesAnalytics, error) {
	analytics, err := s.analyticsRepo.GetSalesAnalytics(startDate, endDate, filters)
	if err != nil {
		logger.Errorf("Failed to get sales analytics: %v", err)
		return nil, fmt.Errorf("failed to retrieve sales analytics")
	}
	return analytics, nil
}

// GetTrafficAnalytics retrieves traffic analytics data
func (s *analyticsService) GetTrafficAnalytics(startDate, endDate time.Time, filters map[string]interface{}) (*model.TrafficAnalytics, error) {
	analytics, err := s.analyticsRepo.GetTrafficAnalytics(startDate, endDate, filters)
	if err != nil {
		logger.Errorf("Failed to get traffic analytics: %v", err)
		return nil, fmt.Errorf("failed to retrieve traffic analytics")
	}
	return analytics, nil
}

// GetUserAnalytics retrieves user analytics data
func (s *analyticsService) GetUserAnalytics(startDate, endDate time.Time, filters map[string]interface{}) (*model.UserAnalytics, error) {
	analytics, err := s.analyticsRepo.GetUserAnalytics(startDate, endDate, filters)
	if err != nil {
		logger.Errorf("Failed to get user analytics: %v", err)
		return nil, fmt.Errorf("failed to retrieve user analytics")
	}
	return analytics, nil
}

// GetInventoryAnalytics retrieves inventory analytics data
func (s *analyticsService) GetInventoryAnalytics(startDate, endDate time.Time, filters map[string]interface{}) (*model.InventoryAnalytics, error) {
	analytics, err := s.analyticsRepo.GetInventoryAnalytics(startDate, endDate, filters)
	if err != nil {
		logger.Errorf("Failed to get inventory analytics: %v", err)
		return nil, fmt.Errorf("failed to retrieve inventory analytics")
	}
	return analytics, nil
}

// GetProductAnalytics retrieves product analytics data
func (s *analyticsService) GetProductAnalytics(startDate, endDate time.Time, filters map[string]interface{}) (map[string]interface{}, error) {
	analytics, err := s.analyticsRepo.GetProductAnalytics(startDate, endDate, filters)
	if err != nil {
		logger.Errorf("Failed to get product analytics: %v", err)
		return nil, fmt.Errorf("failed to retrieve product analytics")
	}
	return analytics, nil
}

// GetOrderAnalytics retrieves order analytics data
func (s *analyticsService) GetOrderAnalytics(startDate, endDate time.Time, filters map[string]interface{}) (map[string]interface{}, error) {
	analytics, err := s.analyticsRepo.GetOrderAnalytics(startDate, endDate, filters)
	if err != nil {
		logger.Errorf("Failed to get order analytics: %v", err)
		return nil, fmt.Errorf("failed to retrieve order analytics")
	}
	return analytics, nil
}

// Event Analytics

// GetEventStats retrieves event statistics
func (s *analyticsService) GetEventStats(startDate, endDate time.Time, eventType string) (map[string]interface{}, error) {
	stats, err := s.analyticsRepo.GetEventStats(startDate, endDate, eventType)
	if err != nil {
		logger.Errorf("Failed to get event stats: %v", err)
		return nil, fmt.Errorf("failed to retrieve event statistics")
	}
	return stats, nil
}

// GetTopEvents retrieves top events
func (s *analyticsService) GetTopEvents(startDate, endDate time.Time, limit int) ([]map[string]interface{}, error) {
	events, err := s.analyticsRepo.GetTopEvents(startDate, endDate, limit)
	if err != nil {
		logger.Errorf("Failed to get top events: %v", err)
		return nil, fmt.Errorf("failed to retrieve top events")
	}
	return events, nil
}

// GetEventTrends retrieves event trends
func (s *analyticsService) GetEventTrends(startDate, endDate time.Time, eventType string) ([]model.PeriodData, error) {
	trends, err := s.analyticsRepo.GetEventTrends(startDate, endDate, eventType)
	if err != nil {
		logger.Errorf("Failed to get event trends: %v", err)
		return nil, fmt.Errorf("failed to retrieve event trends")
	}
	return trends, nil
}

// Custom Analytics

// ExecuteCustomQuery executes a custom SQL query
func (s *analyticsService) ExecuteCustomQuery(query string, params []interface{}) ([]map[string]interface{}, error) {
	results, err := s.analyticsRepo.ExecuteCustomQuery(query, params)
	if err != nil {
		logger.Errorf("Failed to execute custom query: %v", err)
		return nil, fmt.Errorf("failed to execute custom query")
	}
	return results, nil
}

// GetAnalyticsSummary retrieves a comprehensive analytics summary
func (s *analyticsService) GetAnalyticsSummary(startDate, endDate time.Time) (map[string]interface{}, error) {
	summary := make(map[string]interface{})

	// Get sales analytics
	sales, err := s.GetSalesAnalytics(startDate, endDate, map[string]interface{}{})
	if err != nil {
		logger.Errorf("Failed to get sales analytics for summary: %v", err)
	} else {
		summary["sales"] = sales
	}

	// Get traffic analytics
	traffic, err := s.GetTrafficAnalytics(startDate, endDate, map[string]interface{}{})
	if err != nil {
		logger.Errorf("Failed to get traffic analytics for summary: %v", err)
	} else {
		summary["traffic"] = traffic
	}

	// Get user analytics
	users, err := s.GetUserAnalytics(startDate, endDate, map[string]interface{}{})
	if err != nil {
		logger.Errorf("Failed to get user analytics for summary: %v", err)
	} else {
		summary["users"] = users
	}

	// Get inventory analytics
	inventory, err := s.GetInventoryAnalytics(startDate, endDate, map[string]interface{}{})
	if err != nil {
		logger.Errorf("Failed to get inventory analytics for summary: %v", err)
	} else {
		summary["inventory"] = inventory
	}

	// Get event stats
	eventStats, err := s.GetEventStats(startDate, endDate, "")
	if err != nil {
		logger.Errorf("Failed to get event stats for summary: %v", err)
	} else {
		summary["events"] = eventStats
	}

	return summary, nil
}

// Helper methods for response conversion

// toReportResponse converts AnalyticsReport to AnalyticsReportResponse
func (s *analyticsService) toReportResponse(report *model.AnalyticsReport) *model.AnalyticsReportResponse {
	// Parse JSON fields (simplified - would need proper JSON unmarshaling)
	var filters map[string]interface{}
	var data map[string]interface{}

	// Convert Creator to User (simplified)
	creator := report.Creator

	return &model.AnalyticsReportResponse{
		ID:          report.ID,
		Name:        report.Name,
		Description: report.Description,
		Type:        report.Type,
		Period:      report.Period,
		StartDate:   report.StartDate,
		EndDate:     report.EndDate,
		Filters:     filters,
		Data:        data,
		Summary:     report.Summary,
		Insights:    report.Insights,
		Status:      report.Status,
		IsScheduled: report.IsScheduled,
		IsPublic:    report.IsPublic,
		CreatedBy:   report.CreatedBy,
		Creator:     creator,
		CreatedAt:   report.CreatedAt,
		UpdatedAt:   report.UpdatedAt,
	}
}

// toDashboardResponse converts AnalyticsDashboard to AnalyticsDashboardResponse
func (s *analyticsService) toDashboardResponse(dashboard *model.AnalyticsDashboard) *model.AnalyticsDashboardResponse {
	// Parse JSON fields (simplified)
	var layout map[string]interface{}
	var widgets []model.AnalyticsWidgetResponse

	// Convert User to User (simplified)
	var user *model.User
	if dashboard.User != nil && dashboard.User.ID != 0 {
		user = dashboard.User
	}

	// Convert widgets (simplified - would need proper JSON unmarshaling)
	// For now, return empty slice

	return &model.AnalyticsDashboardResponse{
		ID:          dashboard.ID,
		Name:        dashboard.Name,
		Description: dashboard.Description,
		Layout:      layout,
		IsPublic:    dashboard.IsPublic,
		UserID:      dashboard.UserID,
		User:        user,
		Widgets:     widgets,
		CreatedAt:   dashboard.CreatedAt,
		UpdatedAt:   dashboard.UpdatedAt,
	}
}

// toWidgetResponse converts AnalyticsWidget to AnalyticsWidgetResponse
func (s *analyticsService) toWidgetResponse(widget *model.AnalyticsWidget) *model.AnalyticsWidgetResponse {
	// Parse JSON fields (simplified)
	var config map[string]interface{}
	var data map[string]interface{}
	var position map[string]interface{}

	return &model.AnalyticsWidgetResponse{
		ID:          widget.ID,
		DashboardID: widget.DashboardID,
		Type:        widget.Type,
		Title:       widget.Title,
		Description: widget.Description,
		Config:      config,
		Data:        data,
		Position:    position,
		IsActive:    widget.IsActive,
		CreatedAt:   widget.CreatedAt,
		UpdatedAt:   widget.UpdatedAt,
	}
}

// toEventResponse converts AnalyticsEvent to AnalyticsEventResponse
func (s *analyticsService) toEventResponse(event *model.AnalyticsEvent) *model.AnalyticsEventResponse {
	// Parse JSON fields (simplified)
	var properties map[string]interface{}

	// Convert User to User (simplified)
	var user *model.User
	if event.User != nil && event.User.ID != 0 {
		user = event.User
	}

	return &model.AnalyticsEventResponse{
		ID:         event.ID,
		EventType:  event.EventType,
		EventName:  event.EventName,
		EntityType: event.EntityType,
		EntityID:   event.EntityID,
		Properties: properties,
		Value:      event.Value,
		UserID:     event.UserID,
		User:       user,
		SessionID:  event.SessionID,
		IPAddress:  event.IPAddress,
		UserAgent:  event.UserAgent,
		Referer:    event.Referer,
		CreatedAt:  event.CreatedAt,
	}
}
