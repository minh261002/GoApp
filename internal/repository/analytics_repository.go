package repository

import (
	"fmt"
	"go_app/internal/model"
	"time"

	"gorm.io/gorm"
)

// AnalyticsRepository defines methods for analytics data access
type AnalyticsRepository interface {
	// Reports
	CreateReport(report *model.AnalyticsReport) error
	GetReportByID(id uint) (*model.AnalyticsReport, error)
	GetAllReports(page, limit int, filters map[string]interface{}) ([]model.AnalyticsReport, int64, error)
	GetReportsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.AnalyticsReport, int64, error)
	UpdateReport(report *model.AnalyticsReport) error
	DeleteReport(id uint) error

	// Metrics
	CreateMetric(metric *model.AnalyticsMetric) error
	GetMetricsByReport(reportID uint) ([]model.AnalyticsMetric, error)
	UpdateMetric(metric *model.AnalyticsMetric) error
	DeleteMetric(id uint) error

	// Dashboards
	CreateDashboard(dashboard *model.AnalyticsDashboard) error
	GetDashboardByID(id uint) (*model.AnalyticsDashboard, error)
	GetAllDashboards(page, limit int, filters map[string]interface{}) ([]model.AnalyticsDashboard, int64, error)
	GetDashboardsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.AnalyticsDashboard, int64, error)
	GetPublicDashboards(page, limit int) ([]model.AnalyticsDashboard, int64, error)
	UpdateDashboard(dashboard *model.AnalyticsDashboard) error
	DeleteDashboard(id uint) error

	// Widgets
	CreateWidget(widget *model.AnalyticsWidget) error
	GetWidgetByID(id uint) (*model.AnalyticsWidget, error)
	GetWidgetsByDashboard(dashboardID uint) ([]model.AnalyticsWidget, error)
	UpdateWidget(widget *model.AnalyticsWidget) error
	DeleteWidget(id uint) error

	// Events
	CreateEvent(event *model.AnalyticsEvent) error
	GetEvents(page, limit int, filters map[string]interface{}) ([]model.AnalyticsEvent, int64, error)
	GetEventsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.AnalyticsEvent, int64, error)
	GetEventsByType(eventType string, page, limit int, filters map[string]interface{}) ([]model.AnalyticsEvent, int64, error)
	GetEventsByEntity(entityType string, entityID uint, page, limit int) ([]model.AnalyticsEvent, int64, error)

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

	// Custom Queries
	ExecuteCustomQuery(query string, params []interface{}) ([]map[string]interface{}, error)
	GetReportData(reportID uint) (map[string]interface{}, error)
	GenerateReport(report *model.AnalyticsReport) error
}

// analyticsRepository implements AnalyticsRepository
type analyticsRepository struct {
	db *gorm.DB
}

// NewAnalyticsRepository creates a new AnalyticsRepository
func NewAnalyticsRepository(db *gorm.DB) AnalyticsRepository {
	return &analyticsRepository{db: db}
}

// Reports

// CreateReport creates a new analytics report
func (r *analyticsRepository) CreateReport(report *model.AnalyticsReport) error {
	return r.db.Create(report).Error
}

// GetReportByID retrieves a report by ID
func (r *analyticsRepository) GetReportByID(id uint) (*model.AnalyticsReport, error) {
	var report model.AnalyticsReport
	if err := r.db.Preload("Creator").First(&report, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &report, nil
}

// GetAllReports retrieves all reports with pagination and filters
func (r *analyticsRepository) GetAllReports(page, limit int, filters map[string]interface{}) ([]model.AnalyticsReport, int64, error) {
	var reports []model.AnalyticsReport
	var total int64

	query := r.db.Model(&model.AnalyticsReport{}).Where("deleted_at IS NULL")

	// Apply filters
	if reportType, ok := filters["type"].(string); ok && reportType != "" {
		query = query.Where("type = ?", reportType)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if isPublic, ok := filters["is_public"].(bool); ok {
		query = query.Where("is_public = ?", isPublic)
	}
	if startDate, ok := filters["start_date"].(time.Time); ok {
		query = query.Where("start_date >= ?", startDate)
	}
	if endDate, ok := filters["end_date"].(time.Time); ok {
		query = query.Where("end_date <= ?", endDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and get results
	offset := (page - 1) * limit
	if err := query.Preload("Creator").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// GetReportsByUser retrieves reports by user
func (r *analyticsRepository) GetReportsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.AnalyticsReport, int64, error) {
	var reports []model.AnalyticsReport
	var total int64

	query := r.db.Model(&model.AnalyticsReport{}).Where("deleted_at IS NULL AND created_by = ?", userID)

	// Apply filters
	if reportType, ok := filters["type"].(string); ok && reportType != "" {
		query = query.Where("type = ?", reportType)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and get results
	offset := (page - 1) * limit
	if err := query.Preload("Creator").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// UpdateReport updates a report
func (r *analyticsRepository) UpdateReport(report *model.AnalyticsReport) error {
	return r.db.Save(report).Error
}

// DeleteReport deletes a report
func (r *analyticsRepository) DeleteReport(id uint) error {
	return r.db.Delete(&model.AnalyticsReport{}, id).Error
}

// Metrics

// CreateMetric creates a new analytics metric
func (r *analyticsRepository) CreateMetric(metric *model.AnalyticsMetric) error {
	return r.db.Create(metric).Error
}

// GetMetricsByReport retrieves metrics by report ID
func (r *analyticsRepository) GetMetricsByReport(reportID uint) ([]model.AnalyticsMetric, error) {
	var metrics []model.AnalyticsMetric
	if err := r.db.Where("report_id = ? AND deleted_at IS NULL", reportID).
		Order("created_at ASC").
		Find(&metrics).Error; err != nil {
		return nil, err
	}
	return metrics, nil
}

// UpdateMetric updates a metric
func (r *analyticsRepository) UpdateMetric(metric *model.AnalyticsMetric) error {
	return r.db.Save(metric).Error
}

// DeleteMetric deletes a metric
func (r *analyticsRepository) DeleteMetric(id uint) error {
	return r.db.Delete(&model.AnalyticsMetric{}, id).Error
}

// Dashboards

// CreateDashboard creates a new analytics dashboard
func (r *analyticsRepository) CreateDashboard(dashboard *model.AnalyticsDashboard) error {
	return r.db.Create(dashboard).Error
}

// GetDashboardByID retrieves a dashboard by ID
func (r *analyticsRepository) GetDashboardByID(id uint) (*model.AnalyticsDashboard, error) {
	var dashboard model.AnalyticsDashboard
	if err := r.db.Preload("User").Preload("Widgets").
		First(&dashboard, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &dashboard, nil
}

// GetAllDashboards retrieves all dashboards with pagination and filters
func (r *analyticsRepository) GetAllDashboards(page, limit int, filters map[string]interface{}) ([]model.AnalyticsDashboard, int64, error) {
	var dashboards []model.AnalyticsDashboard
	var total int64

	query := r.db.Model(&model.AnalyticsDashboard{}).Where("deleted_at IS NULL")

	// Apply filters
	if isPublic, ok := filters["is_public"].(bool); ok {
		query = query.Where("is_public = ?", isPublic)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and get results
	offset := (page - 1) * limit
	if err := query.Preload("User").Preload("Widgets").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&dashboards).Error; err != nil {
		return nil, 0, err
	}

	return dashboards, total, nil
}

// GetDashboardsByUser retrieves dashboards by user
func (r *analyticsRepository) GetDashboardsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.AnalyticsDashboard, int64, error) {
	var dashboards []model.AnalyticsDashboard
	var total int64

	query := r.db.Model(&model.AnalyticsDashboard{}).Where("deleted_at IS NULL AND user_id = ?", userID)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and get results
	offset := (page - 1) * limit
	if err := query.Preload("User").Preload("Widgets").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&dashboards).Error; err != nil {
		return nil, 0, err
	}

	return dashboards, total, nil
}

// GetPublicDashboards retrieves public dashboards
func (r *analyticsRepository) GetPublicDashboards(page, limit int) ([]model.AnalyticsDashboard, int64, error) {
	var dashboards []model.AnalyticsDashboard
	var total int64

	query := r.db.Model(&model.AnalyticsDashboard{}).Where("deleted_at IS NULL AND is_public = TRUE")

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and get results
	offset := (page - 1) * limit
	if err := query.Preload("User").Preload("Widgets").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&dashboards).Error; err != nil {
		return nil, 0, err
	}

	return dashboards, total, nil
}

// UpdateDashboard updates a dashboard
func (r *analyticsRepository) UpdateDashboard(dashboard *model.AnalyticsDashboard) error {
	return r.db.Save(dashboard).Error
}

// DeleteDashboard deletes a dashboard
func (r *analyticsRepository) DeleteDashboard(id uint) error {
	return r.db.Delete(&model.AnalyticsDashboard{}, id).Error
}

// Widgets

// CreateWidget creates a new analytics widget
func (r *analyticsRepository) CreateWidget(widget *model.AnalyticsWidget) error {
	return r.db.Create(widget).Error
}

// GetWidgetByID retrieves a widget by ID
func (r *analyticsRepository) GetWidgetByID(id uint) (*model.AnalyticsWidget, error) {
	var widget model.AnalyticsWidget
	if err := r.db.Preload("Dashboard").
		First(&widget, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &widget, nil
}

// GetWidgetsByDashboard retrieves widgets by dashboard ID
func (r *analyticsRepository) GetWidgetsByDashboard(dashboardID uint) ([]model.AnalyticsWidget, error) {
	var widgets []model.AnalyticsWidget
	if err := r.db.Where("dashboard_id = ? AND deleted_at IS NULL", dashboardID).
		Order("created_at ASC").
		Find(&widgets).Error; err != nil {
		return nil, err
	}
	return widgets, nil
}

// UpdateWidget updates a widget
func (r *analyticsRepository) UpdateWidget(widget *model.AnalyticsWidget) error {
	return r.db.Save(widget).Error
}

// DeleteWidget deletes a widget
func (r *analyticsRepository) DeleteWidget(id uint) error {
	return r.db.Delete(&model.AnalyticsWidget{}, id).Error
}

// Events

// CreateEvent creates a new analytics event
func (r *analyticsRepository) CreateEvent(event *model.AnalyticsEvent) error {
	return r.db.Create(event).Error
}

// GetEvents retrieves events with pagination and filters
func (r *analyticsRepository) GetEvents(page, limit int, filters map[string]interface{}) ([]model.AnalyticsEvent, int64, error) {
	var events []model.AnalyticsEvent
	var total int64

	query := r.db.Model(&model.AnalyticsEvent{}).Where("deleted_at IS NULL")

	// Apply filters
	if eventType, ok := filters["event_type"].(string); ok && eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	if entityType, ok := filters["entity_type"].(string); ok && entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}
	if startDate, ok := filters["start_date"].(time.Time); ok {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate, ok := filters["end_date"].(time.Time); ok {
		query = query.Where("created_at <= ?", endDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and get results
	offset := (page - 1) * limit
	if err := query.Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

// GetEventsByUser retrieves events by user
func (r *analyticsRepository) GetEventsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.AnalyticsEvent, int64, error) {
	var events []model.AnalyticsEvent
	var total int64

	query := r.db.Model(&model.AnalyticsEvent{}).Where("deleted_at IS NULL AND user_id = ?", userID)

	// Apply filters
	if eventType, ok := filters["event_type"].(string); ok && eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	if startDate, ok := filters["start_date"].(time.Time); ok {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate, ok := filters["end_date"].(time.Time); ok {
		query = query.Where("created_at <= ?", endDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and get results
	offset := (page - 1) * limit
	if err := query.Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

// GetEventsByType retrieves events by type
func (r *analyticsRepository) GetEventsByType(eventType string, page, limit int, filters map[string]interface{}) ([]model.AnalyticsEvent, int64, error) {
	var events []model.AnalyticsEvent
	var total int64

	query := r.db.Model(&model.AnalyticsEvent{}).Where("deleted_at IS NULL AND event_type = ?", eventType)

	// Apply filters
	if startDate, ok := filters["start_date"].(time.Time); ok {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate, ok := filters["end_date"].(time.Time); ok {
		query = query.Where("created_at <= ?", endDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and get results
	offset := (page - 1) * limit
	if err := query.Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

// GetEventsByEntity retrieves events by entity
func (r *analyticsRepository) GetEventsByEntity(entityType string, entityID uint, page, limit int) ([]model.AnalyticsEvent, int64, error) {
	var events []model.AnalyticsEvent
	var total int64

	query := r.db.Model(&model.AnalyticsEvent{}).Where("deleted_at IS NULL AND entity_type = ? AND entity_id = ?", entityType, entityID)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and get results
	offset := (page - 1) * limit
	if err := query.Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

// Analytics Data

// GetSalesAnalytics retrieves sales analytics data
func (r *analyticsRepository) GetSalesAnalytics(startDate, endDate time.Time, filters map[string]interface{}) (*model.SalesAnalytics, error) {
	var analytics model.SalesAnalytics

	// Base query for orders
	query := r.db.Model(&model.Order{}).Where("deleted_at IS NULL AND created_at BETWEEN ? AND ?", startDate, endDate)

	// Apply filters
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}

	// Total revenue
	var totalRevenue float64
	if err := query.Select("COALESCE(SUM(total_amount), 0)").Scan(&totalRevenue).Error; err != nil {
		return nil, err
	}
	analytics.TotalRevenue = totalRevenue

	// Total orders
	var totalOrders int64
	if err := query.Count(&totalOrders).Error; err != nil {
		return nil, err
	}
	analytics.TotalOrders = totalOrders

	// Average order value
	if totalOrders > 0 {
		analytics.AverageOrderValue = totalRevenue / float64(totalOrders)
	}

	// Top products (simplified - would need order_items join in real implementation)
	// This is a placeholder - actual implementation would join with order_items
	analytics.TopProducts = []model.ProductSalesData{}

	// Revenue by period (simplified)
	analytics.RevenueByPeriod = []model.PeriodData{}

	return &analytics, nil
}

// GetTrafficAnalytics retrieves traffic analytics data
func (r *analyticsRepository) GetTrafficAnalytics(startDate, endDate time.Time, filters map[string]interface{}) (*model.TrafficAnalytics, error) {
	var analytics model.TrafficAnalytics

	// Base query for page views
	query := r.db.Model(&model.AnalyticsEvent{}).Where("deleted_at IS NULL AND event_type = ? AND created_at BETWEEN ? AND ?", model.EventTypePageView, startDate, endDate)

	// Total page views
	var pageViews int64
	if err := query.Count(&pageViews).Error; err != nil {
		return nil, err
	}
	analytics.PageViews = pageViews

	// Unique visitors
	var uniqueVisitors int64
	if err := query.Distinct("user_id").Count(&uniqueVisitors).Error; err != nil {
		return nil, err
	}
	analytics.UniqueVisitors = uniqueVisitors

	// Total visitors (including anonymous)
	var totalVisitors int64
	if err := query.Distinct("COALESCE(user_id, session_id)").Count(&totalVisitors).Error; err != nil {
		return nil, err
	}
	analytics.TotalVisitors = totalVisitors

	// Top pages (simplified)
	analytics.TopPages = []model.PageData{}

	// Traffic sources (simplified)
	analytics.TrafficSources = []model.SourceData{}

	return &analytics, nil
}

// GetUserAnalytics retrieves user analytics data
func (r *analyticsRepository) GetUserAnalytics(startDate, endDate time.Time, filters map[string]interface{}) (*model.UserAnalytics, error) {
	var analytics model.UserAnalytics

	// Base query for users
	query := r.db.Model(&model.User{}).Where("deleted_at IS NULL")

	// Total users
	var totalUsers int64
	if err := query.Count(&totalUsers).Error; err != nil {
		return nil, err
	}
	analytics.TotalUsers = totalUsers

	// New users in period
	var newUsers int64
	if err := query.Where("created_at BETWEEN ? AND ?", startDate, endDate).Count(&newUsers).Error; err != nil {
		return nil, err
	}
	analytics.NewUsers = newUsers

	// Active users (users with events in period)
	var activeUsers int64
	if err := r.db.Model(&model.AnalyticsEvent{}).
		Where("deleted_at IS NULL AND user_id IS NOT NULL AND created_at BETWEEN ? AND ?", startDate, endDate).
		Distinct("user_id").
		Count(&activeUsers).Error; err != nil {
		return nil, err
	}
	analytics.ActiveUsers = activeUsers

	// Top countries (simplified)
	analytics.TopCountries = []model.CountryData{}

	// User segments (simplified)
	analytics.UserSegments = []model.SegmentData{}

	return &analytics, nil
}

// GetInventoryAnalytics retrieves inventory analytics data
func (r *analyticsRepository) GetInventoryAnalytics(startDate, endDate time.Time, filters map[string]interface{}) (*model.InventoryAnalytics, error) {
	var analytics model.InventoryAnalytics

	// Base query for products
	query := r.db.Model(&model.Product{}).Where("deleted_at IS NULL")

	// Total products
	var totalProducts int64
	if err := query.Count(&totalProducts).Error; err != nil {
		return nil, err
	}
	analytics.TotalProducts = totalProducts

	// In stock products
	var inStockProducts int64
	if err := query.Where("stock_quantity > 0").Count(&inStockProducts).Error; err != nil {
		return nil, err
	}
	analytics.InStockProducts = inStockProducts

	// Out of stock products
	var outOfStockProducts int64
	if err := query.Where("stock_quantity = 0").Count(&outOfStockProducts).Error; err != nil {
		return nil, err
	}
	analytics.OutOfStockProducts = outOfStockProducts

	// Low stock products (assuming min_stock_quantity field exists)
	var lowStockProducts int64
	if err := query.Where("stock_quantity > 0 AND stock_quantity <= min_stock_quantity").Count(&lowStockProducts).Error; err != nil {
		return nil, err
	}
	analytics.LowStockProducts = lowStockProducts

	// Top selling products (simplified)
	analytics.TopSellingProducts = []model.ProductInventoryData{}

	// Low stock alerts (simplified)
	analytics.LowStockAlerts = []model.LowStockData{}

	return &analytics, nil
}

// GetProductAnalytics retrieves product analytics data
func (r *analyticsRepository) GetProductAnalytics(startDate, endDate time.Time, filters map[string]interface{}) (map[string]interface{}, error) {
	// Placeholder implementation
	return map[string]interface{}{
		"total_products":  0,
		"active_products": 0,
		"top_products":    []interface{}{},
		"product_views":   0,
		"conversion_rate": 0.0,
	}, nil
}

// GetOrderAnalytics retrieves order analytics data
func (r *analyticsRepository) GetOrderAnalytics(startDate, endDate time.Time, filters map[string]interface{}) (map[string]interface{}, error) {
	// Placeholder implementation
	return map[string]interface{}{
		"total_orders":           0,
		"total_revenue":          0.0,
		"average_order_value":    0.0,
		"order_status_breakdown": map[string]interface{}{},
	}, nil
}

// Event Analytics

// GetEventStats retrieves event statistics
func (r *analyticsRepository) GetEventStats(startDate, endDate time.Time, eventType string) (map[string]interface{}, error) {
	var stats map[string]interface{} = make(map[string]interface{})

	query := r.db.Model(&model.AnalyticsEvent{}).Where("deleted_at IS NULL AND created_at BETWEEN ? AND ?", startDate, endDate)
	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}

	// Total events
	var totalEvents int64
	if err := query.Count(&totalEvents).Error; err != nil {
		return nil, err
	}
	stats["total_events"] = totalEvents

	// Unique users
	var uniqueUsers int64
	if err := query.Distinct("user_id").Count(&uniqueUsers).Error; err != nil {
		return nil, err
	}
	stats["unique_users"] = uniqueUsers

	// Events by type
	var eventsByType []map[string]interface{}
	if err := query.Select("event_type, COUNT(*) as count").
		Group("event_type").
		Scan(&eventsByType).Error; err != nil {
		return nil, err
	}
	stats["events_by_type"] = eventsByType

	return stats, nil
}

// GetTopEvents retrieves top events
func (r *analyticsRepository) GetTopEvents(startDate, endDate time.Time, limit int) ([]map[string]interface{}, error) {
	var topEvents []map[string]interface{}

	if err := r.db.Model(&model.AnalyticsEvent{}).
		Select("event_type, event_name, COUNT(*) as count").
		Where("deleted_at IS NULL AND created_at BETWEEN ? AND ?", startDate, endDate).
		Group("event_type, event_name").
		Order("count DESC").
		Limit(limit).
		Scan(&topEvents).Error; err != nil {
		return nil, err
	}

	return topEvents, nil
}

// GetEventTrends retrieves event trends
func (r *analyticsRepository) GetEventTrends(startDate, endDate time.Time, eventType string) ([]model.PeriodData, error) {
	var trends []model.PeriodData

	query := r.db.Model(&model.AnalyticsEvent{}).
		Select("DATE(created_at) as period, COUNT(*) as count").
		Where("deleted_at IS NULL AND created_at BETWEEN ? AND ?", startDate, endDate)

	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}

	if err := query.Group("DATE(created_at)").
		Order("period ASC").
		Scan(&trends).Error; err != nil {
		return nil, err
	}

	return trends, nil
}

// Custom Queries

// ExecuteCustomQuery executes a custom SQL query
func (r *analyticsRepository) ExecuteCustomQuery(query string, params []interface{}) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	if err := r.db.Raw(query, params...).Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

// GetReportData retrieves report data
func (r *analyticsRepository) GetReportData(reportID uint) (map[string]interface{}, error) {
	report, err := r.GetReportByID(reportID)
	if err != nil {
		return nil, err
	}
	if report == nil {
		return nil, fmt.Errorf("report not found")
	}

	// Parse report data from JSON
	// This would need proper JSON unmarshaling in real implementation
	return map[string]interface{}{
		"report_id": report.ID,
		"name":      report.Name,
		"type":      report.Type,
		"data":      report.Data,
	}, nil
}

// GenerateReport generates a report
func (r *analyticsRepository) GenerateReport(report *model.AnalyticsReport) error {
	// Update report status to processing
	report.Status = model.ReportStatusProcessing
	if err := r.UpdateReport(report); err != nil {
		return err
	}

	// Generate report data based on type
	var data map[string]interface{}

	switch report.Type {
	case model.ReportTypeSales:
		salesData, err := r.GetSalesAnalytics(report.StartDate, report.EndDate, map[string]interface{}{})
		if err != nil {
			return err
		}
		data = map[string]interface{}{"sales": salesData}
	case model.ReportTypeTraffic:
		trafficData, err := r.GetTrafficAnalytics(report.StartDate, report.EndDate, map[string]interface{}{})
		if err != nil {
			return err
		}
		data = map[string]interface{}{"traffic": trafficData}
	case model.ReportTypeUser:
		userData, err := r.GetUserAnalytics(report.StartDate, report.EndDate, map[string]interface{}{})
		if err != nil {
			return err
		}
		data = map[string]interface{}{"users": userData}
	case model.ReportTypeInventory:
		inventoryData, err := r.GetInventoryAnalytics(report.StartDate, report.EndDate, map[string]interface{}{})
		if err != nil {
			return err
		}
		data = map[string]interface{}{"inventory": inventoryData}
	default:
		data = map[string]interface{}{"message": "Report type not implemented"}
	}

	// No error handling needed here since we handle errors in each case

	// Update report with generated data
	report.Data = fmt.Sprintf("%v", data) // Simplified - would need proper JSON marshaling
	report.Status = model.ReportStatusCompleted
	return r.UpdateReport(report)
}
