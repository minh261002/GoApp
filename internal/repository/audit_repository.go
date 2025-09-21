package repository

import (
	"fmt"
	"go_app/internal/model"
	"strings"
	"time"

	"gorm.io/gorm"
)

// AuditRepository defines methods for audit log data access
type AuditRepository interface {
	// Audit Logs
	CreateAuditLog(log *model.AuditLog) error
	GetAuditLogByID(id uint) (*model.AuditLog, error)
	GetAllAuditLogs(page, limit int, filters map[string]interface{}) ([]model.AuditLog, int64, error)
	GetAuditLogsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.AuditLog, int64, error)
	GetAuditLogsByResource(resource string, resourceID uint, page, limit int) ([]model.AuditLog, int64, error)
	GetAuditLogsByAction(action string, page, limit int, filters map[string]interface{}) ([]model.AuditLog, int64, error)
	SearchAuditLogs(req *model.AuditSearchRequest) ([]model.AuditLog, int64, error)
	UpdateAuditLog(log *model.AuditLog) error
	DeleteAuditLog(id uint) error
	DeleteOldLogs(retentionDays int) error

	// Audit Config
	CreateAuditConfig(config *model.AuditLogConfig) error
	GetAuditConfigByID(id uint) (*model.AuditLogConfig, error)
	GetAuditConfigByName(name string) (*model.AuditLogConfig, error)
	GetAllAuditConfigs(page, limit int, filters map[string]interface{}) ([]model.AuditLogConfig, int64, error)
	UpdateAuditConfig(config *model.AuditLogConfig) error
	DeleteAuditConfig(id uint) error

	// Audit Summaries
	CreateAuditSummary(summary *model.AuditLogSummary) error
	GetAuditSummaryByID(id uint) (*model.AuditLogSummary, error)
	GetAuditSummaries(startDate, endDate time.Time, filters map[string]interface{}) ([]model.AuditLogSummary, error)
	UpdateAuditSummary(summary *model.AuditLogSummary) error
	DeleteAuditSummary(id uint) error
	GenerateDailySummaries(date time.Time) error

	// Statistics
	GetAuditStats(startDate, endDate time.Time, filters map[string]interface{}) (*model.AuditStats, error)
	GetTopActions(startDate, endDate time.Time, limit int) ([]model.ActionStats, error)
	GetTopResources(startDate, endDate time.Time, limit int) ([]model.ResourceStats, error)
	GetTopUsers(startDate, endDate time.Time, limit int) ([]model.UserStats, error)
	GetDailyStats(startDate, endDate time.Time) ([]model.DailyStats, error)
	GetHourlyStats(startDate, endDate time.Time) ([]model.HourlyStats, error)
	GetRecentActivity(limit int) ([]model.AuditLog, error)

	// Export
	ExportAuditLogs(req *model.AuditExportRequest) (*model.AuditExportResponse, error)
	GetAuditLogsForExport(filters map[string]interface{}) ([]model.AuditLog, error)

	// Cleanup
	CleanupOldLogs(retentionDays int) error
	CleanupOldSummaries(retentionDays int) error
	OptimizeAuditTables() error
}

// auditRepository implements AuditRepository
type auditRepository struct {
	db *gorm.DB
}

// NewAuditRepository creates a new AuditRepository
func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{db: db}
}

// Audit Logs

// CreateAuditLog creates a new audit log entry
func (r *auditRepository) CreateAuditLog(log *model.AuditLog) error {
	return r.db.Create(log).Error
}

// GetAuditLogByID retrieves an audit log by ID
func (r *auditRepository) GetAuditLogByID(id uint) (*model.AuditLog, error) {
	var log model.AuditLog
	if err := r.db.Preload("User").Preload("TargetUser").
		First(&log, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &log, nil
}

// GetAllAuditLogs retrieves all audit logs with pagination and filters
func (r *auditRepository) GetAllAuditLogs(page, limit int, filters map[string]interface{}) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64

	query := r.db.Model(&model.AuditLog{}).Where("deleted_at IS NULL")

	// Apply filters
	if userID, ok := filters["user_id"].(uint); ok && userID != 0 {
		query = query.Where("user_id = ?", userID)
	}
	if action, ok := filters["action"].(string); ok && action != "" {
		query = query.Where("action = ?", action)
	}
	if resource, ok := filters["resource"].(string); ok && resource != "" {
		query = query.Where("resource = ?", resource)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if severity, ok := filters["severity"].(string); ok && severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if startDate, ok := filters["start_date"].(time.Time); ok {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate, ok := filters["end_date"].(time.Time); ok {
		query = query.Where("created_at <= ?", endDate)
	}
	if ipAddress, ok := filters["ip_address"].(string); ok && ipAddress != "" {
		query = query.Where("ip_address = ?", ipAddress)
	}
	if sessionID, ok := filters["session_id"].(string); ok && sessionID != "" {
		query = query.Where("session_id = ?", sessionID)
	}
	if tags, ok := filters["tags"].([]string); ok && len(tags) > 0 {
		for _, tag := range tags {
			query = query.Where("tags LIKE ?", "%"+tag+"%")
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and get results
	offset := (page - 1) * limit
	if err := query.Preload("User").Preload("TargetUser").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetAuditLogsByUser retrieves audit logs by user
func (r *auditRepository) GetAuditLogsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64

	query := r.db.Model(&model.AuditLog{}).Where("deleted_at IS NULL AND user_id = ?", userID)

	// Apply filters
	if action, ok := filters["action"].(string); ok && action != "" {
		query = query.Where("action = ?", action)
	}
	if resource, ok := filters["resource"].(string); ok && resource != "" {
		query = query.Where("resource = ?", resource)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
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
	if err := query.Preload("User").Preload("TargetUser").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetAuditLogsByResource retrieves audit logs by resource
func (r *auditRepository) GetAuditLogsByResource(resource string, resourceID uint, page, limit int) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64

	query := r.db.Model(&model.AuditLog{}).Where("deleted_at IS NULL AND resource = ? AND resource_id = ?", resource, resourceID)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and get results
	offset := (page - 1) * limit
	if err := query.Preload("User").Preload("TargetUser").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetAuditLogsByAction retrieves audit logs by action
func (r *auditRepository) GetAuditLogsByAction(action string, page, limit int, filters map[string]interface{}) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64

	query := r.db.Model(&model.AuditLog{}).Where("deleted_at IS NULL AND action = ?", action)

	// Apply filters
	if userID, ok := filters["user_id"].(uint); ok && userID != 0 {
		query = query.Where("user_id = ?", userID)
	}
	if resource, ok := filters["resource"].(string); ok && resource != "" {
		query = query.Where("resource = ?", resource)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
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
	if err := query.Preload("User").Preload("TargetUser").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// SearchAuditLogs performs advanced search on audit logs
func (r *auditRepository) SearchAuditLogs(req *model.AuditSearchRequest) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64

	query := r.db.Model(&model.AuditLog{}).Where("deleted_at IS NULL")

	// Apply search filters
	if req.Query != "" {
		searchTerm := "%" + req.Query + "%"
		query = query.Where("(message LIKE ? OR resource_name LIKE ? OR action LIKE ? OR resource LIKE ?)",
			searchTerm, searchTerm, searchTerm, searchTerm)
	}
	if req.UserID != nil {
		query = query.Where("user_id = ?", *req.UserID)
	}
	if req.Action != "" {
		query = query.Where("action = ?", req.Action)
	}
	if req.Resource != "" {
		query = query.Where("resource = ?", req.Resource)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Severity != "" {
		query = query.Where("severity = ?", req.Severity)
	}
	if req.StartDate != nil {
		query = query.Where("created_at >= ?", *req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("created_at <= ?", *req.EndDate)
	}
	if req.IPAddress != "" {
		query = query.Where("ip_address = ?", req.IPAddress)
	}
	if req.SessionID != "" {
		query = query.Where("session_id = ?", req.SessionID)
	}
	if len(req.Tags) > 0 {
		for _, tag := range req.Tags {
			query = query.Where("tags LIKE ?", "%"+tag+"%")
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder := req.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, strings.ToUpper(sortOrder)))

	// Apply pagination and get results
	offset := (req.Page - 1) * req.Limit
	if err := query.Preload("User").Preload("TargetUser").
		Offset(offset).
		Limit(req.Limit).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// UpdateAuditLog updates an audit log
func (r *auditRepository) UpdateAuditLog(log *model.AuditLog) error {
	return r.db.Save(log).Error
}

// DeleteAuditLog deletes an audit log
func (r *auditRepository) DeleteAuditLog(id uint) error {
	return r.db.Delete(&model.AuditLog{}, id).Error
}

// DeleteOldLogs deletes old audit logs based on retention policy
func (r *auditRepository) DeleteOldLogs(retentionDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	return r.db.Where("created_at < ?", cutoffDate).Delete(&model.AuditLog{}).Error
}

// Audit Config

// CreateAuditConfig creates a new audit log configuration
func (r *auditRepository) CreateAuditConfig(config *model.AuditLogConfig) error {
	return r.db.Create(config).Error
}

// GetAuditConfigByID retrieves an audit config by ID
func (r *auditRepository) GetAuditConfigByID(id uint) (*model.AuditLogConfig, error) {
	var config model.AuditLogConfig
	if err := r.db.First(&config, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// GetAuditConfigByName retrieves an audit config by name
func (r *auditRepository) GetAuditConfigByName(name string) (*model.AuditLogConfig, error) {
	var config model.AuditLogConfig
	if err := r.db.Where("name = ? AND deleted_at IS NULL", name).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// GetAllAuditConfigs retrieves all audit configs with pagination and filters
func (r *auditRepository) GetAllAuditConfigs(page, limit int, filters map[string]interface{}) ([]model.AuditLogConfig, int64, error) {
	var configs []model.AuditLogConfig
	var total int64

	query := r.db.Model(&model.AuditLogConfig{}).Where("deleted_at IS NULL")

	// Apply filters
	if isEnabled, ok := filters["is_enabled"].(bool); ok {
		query = query.Where("is_enabled = ?", isEnabled)
	}
	if logLevel, ok := filters["log_level"].(string); ok && logLevel != "" {
		query = query.Where("log_level = ?", logLevel)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and get results
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&configs).Error; err != nil {
		return nil, 0, err
	}

	return configs, total, nil
}

// UpdateAuditConfig updates an audit config
func (r *auditRepository) UpdateAuditConfig(config *model.AuditLogConfig) error {
	return r.db.Save(config).Error
}

// DeleteAuditConfig deletes an audit config
func (r *auditRepository) DeleteAuditConfig(id uint) error {
	return r.db.Delete(&model.AuditLogConfig{}, id).Error
}

// Audit Summaries

// CreateAuditSummary creates a new audit log summary
func (r *auditRepository) CreateAuditSummary(summary *model.AuditLogSummary) error {
	return r.db.Create(summary).Error
}

// GetAuditSummaryByID retrieves an audit summary by ID
func (r *auditRepository) GetAuditSummaryByID(id uint) (*model.AuditLogSummary, error) {
	var summary model.AuditLogSummary
	if err := r.db.First(&summary, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &summary, nil
}

// GetAuditSummaries retrieves audit summaries for a date range
func (r *auditRepository) GetAuditSummaries(startDate, endDate time.Time, filters map[string]interface{}) ([]model.AuditLogSummary, error) {
	var summaries []model.AuditLogSummary

	query := r.db.Model(&model.AuditLogSummary{}).Where("deleted_at IS NULL AND date BETWEEN ? AND ?", startDate, endDate)

	// Apply filters
	if resource, ok := filters["resource"].(string); ok && resource != "" {
		query = query.Where("resource = ?", resource)
	}
	if action, ok := filters["action"].(string); ok && action != "" {
		query = query.Where("action = ?", action)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("date DESC, resource, action").Find(&summaries).Error; err != nil {
		return nil, err
	}

	return summaries, nil
}

// UpdateAuditSummary updates an audit summary
func (r *auditRepository) UpdateAuditSummary(summary *model.AuditLogSummary) error {
	return r.db.Save(summary).Error
}

// DeleteAuditSummary deletes an audit summary
func (r *auditRepository) DeleteAuditSummary(id uint) error {
	return r.db.Delete(&model.AuditLogSummary{}, id).Error
}

// GenerateDailySummaries generates daily summaries for a specific date
func (r *auditRepository) GenerateDailySummaries(date time.Time) error {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Get unique combinations of resource, action, status for the day
	var combinations []struct {
		Resource string
		Action   string
		Status   string
	}

	if err := r.db.Model(&model.AuditLog{}).
		Select("DISTINCT resource, action, status").
		Where("created_at BETWEEN ? AND ? AND deleted_at IS NULL", startOfDay, endOfDay).
		Scan(&combinations).Error; err != nil {
		return err
	}

	// Generate summary for each combination
	for _, combo := range combinations {
		var summary model.AuditLogSummary

		// Check if summary already exists
		if err := r.db.Where("date = ? AND resource = ? AND action = ? AND status = ?",
			startOfDay, combo.Resource, combo.Action, combo.Status).
			First(&summary).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}
			// Create new summary
			summary = model.AuditLogSummary{
				Date:      startOfDay,
				Resource:  combo.Resource,
				Action:    combo.Action,
				Status:    combo.Status,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}

		// Calculate statistics
		var stats struct {
			TotalCount   int64
			SuccessCount int64
			FailureCount int64
			ErrorCount   int64
			UniqueUsers  int64
		}

		// Total count
		if err := r.db.Model(&model.AuditLog{}).
			Where("created_at BETWEEN ? AND ? AND resource = ? AND action = ? AND status = ? AND deleted_at IS NULL",
				startOfDay, endOfDay, combo.Resource, combo.Action, combo.Status).
			Count(&stats.TotalCount).Error; err != nil {
			return err
		}

		// Success count
		if err := r.db.Model(&model.AuditLog{}).
			Where("created_at BETWEEN ? AND ? AND resource = ? AND action = ? AND status = ? AND deleted_at IS NULL",
				startOfDay, endOfDay, combo.Resource, combo.Action, model.StatusSuccess).
			Count(&stats.SuccessCount).Error; err != nil {
			return err
		}

		// Failure count
		if err := r.db.Model(&model.AuditLog{}).
			Where("created_at BETWEEN ? AND ? AND resource = ? AND action = ? AND status = ? AND deleted_at IS NULL",
				startOfDay, endOfDay, combo.Resource, combo.Action, model.StatusFailure).
			Count(&stats.FailureCount).Error; err != nil {
			return err
		}

		// Error count
		if err := r.db.Model(&model.AuditLog{}).
			Where("created_at BETWEEN ? AND ? AND resource = ? AND action = ? AND status = ? AND deleted_at IS NULL",
				startOfDay, endOfDay, combo.Resource, combo.Action, model.StatusError).
			Count(&stats.ErrorCount).Error; err != nil {
			return err
		}

		// Unique users
		if err := r.db.Model(&model.AuditLog{}).
			Where("created_at BETWEEN ? AND ? AND resource = ? AND action = ? AND deleted_at IS NULL",
				startOfDay, endOfDay, combo.Resource, combo.Action).
			Distinct("user_id").
			Count(&stats.UniqueUsers).Error; err != nil {
			return err
		}

		// Update summary
		summary.TotalCount = stats.TotalCount
		summary.SuccessCount = stats.SuccessCount
		summary.FailureCount = stats.FailureCount
		summary.ErrorCount = stats.ErrorCount
		summary.UniqueUsers = stats.UniqueUsers
		summary.UpdatedAt = time.Now()

		// Save summary
		if err := r.db.Save(&summary).Error; err != nil {
			return err
		}
	}

	return nil
}

// Statistics

// GetAuditStats retrieves comprehensive audit statistics
func (r *auditRepository) GetAuditStats(startDate, endDate time.Time, filters map[string]interface{}) (*model.AuditStats, error) {
	stats := &model.AuditStats{}

	query := r.db.Model(&model.AuditLog{}).Where("deleted_at IS NULL AND created_at BETWEEN ? AND ?", startDate, endDate)

	// Apply filters
	if userID, ok := filters["user_id"].(uint); ok && userID != 0 {
		query = query.Where("user_id = ?", userID)
	}
	if action, ok := filters["action"].(string); ok && action != "" {
		query = query.Where("action = ?", action)
	}
	if resource, ok := filters["resource"].(string); ok && resource != "" {
		query = query.Where("resource = ?", resource)
	}

	// Total logs
	if err := query.Count(&stats.TotalLogs).Error; err != nil {
		return nil, err
	}

	// Success logs
	if err := query.Where("status = ?", model.StatusSuccess).Count(&stats.SuccessLogs).Error; err != nil {
		return nil, err
	}

	// Failure logs
	if err := query.Where("status = ?", model.StatusFailure).Count(&stats.FailureLogs).Error; err != nil {
		return nil, err
	}

	// Error logs
	if err := query.Where("status = ?", model.StatusError).Count(&stats.ErrorLogs).Error; err != nil {
		return nil, err
	}

	// Unique users
	if err := query.Distinct("user_id").Count(&stats.UniqueUsers).Error; err != nil {
		return nil, err
	}

	// Get top actions
	topActions, err := r.GetTopActions(startDate, endDate, 10)
	if err != nil {
		return nil, err
	}
	stats.TopActions = topActions

	// Get top resources
	topResources, err := r.GetTopResources(startDate, endDate, 10)
	if err != nil {
		return nil, err
	}
	stats.TopResources = topResources

	// Get top users
	topUsers, err := r.GetTopUsers(startDate, endDate, 10)
	if err != nil {
		return nil, err
	}
	stats.TopUsers = topUsers

	// Get recent activity
	recentActivity, err := r.GetRecentActivity(20)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	var recentActivityResponse []model.AuditLogResponse
	for _, log := range recentActivity {
		recentActivityResponse = append(recentActivityResponse, *r.toAuditLogResponse(&log))
	}
	stats.RecentActivity = recentActivityResponse

	// Get daily stats
	dailyStats, err := r.GetDailyStats(startDate, endDate)
	if err != nil {
		return nil, err
	}
	stats.DailyStats = dailyStats

	// Get hourly stats
	hourlyStats, err := r.GetHourlyStats(startDate, endDate)
	if err != nil {
		return nil, err
	}
	stats.HourlyStats = hourlyStats

	return stats, nil
}

// GetTopActions retrieves top actions by count
func (r *auditRepository) GetTopActions(startDate, endDate time.Time, limit int) ([]model.ActionStats, error) {
	var actions []model.ActionStats

	if err := r.db.Model(&model.AuditLog{}).
		Select("action, COUNT(*) as count, (SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) * 100.0 / COUNT(*)) as success_rate").
		Where("deleted_at IS NULL AND created_at BETWEEN ? AND ?", startDate, endDate).
		Group("action").
		Order("count DESC").
		Limit(limit).
		Scan(&actions).Error; err != nil {
		return nil, err
	}

	return actions, nil
}

// GetTopResources retrieves top resources by count
func (r *auditRepository) GetTopResources(startDate, endDate time.Time, limit int) ([]model.ResourceStats, error) {
	var resources []model.ResourceStats

	if err := r.db.Model(&model.AuditLog{}).
		Select("resource, COUNT(*) as count, (SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) * 100.0 / COUNT(*)) as success_rate").
		Where("deleted_at IS NULL AND created_at BETWEEN ? AND ?", startDate, endDate).
		Group("resource").
		Order("count DESC").
		Limit(limit).
		Scan(&resources).Error; err != nil {
		return nil, err
	}

	return resources, nil
}

// GetTopUsers retrieves top users by activity count
func (r *auditRepository) GetTopUsers(startDate, endDate time.Time, limit int) ([]model.UserStats, error) {
	var users []model.UserStats

	if err := r.db.Table("audit_logs").
		Select("audit_logs.user_id, users.username, users.email, COUNT(*) as action_count, MAX(audit_logs.created_at) as last_activity").
		Joins("LEFT JOIN users ON audit_logs.user_id = users.id").
		Where("audit_logs.deleted_at IS NULL AND audit_logs.created_at BETWEEN ? AND ? AND audit_logs.user_id IS NOT NULL", startDate, endDate).
		Group("audit_logs.user_id, users.username, users.email").
		Order("action_count DESC").
		Limit(limit).
		Scan(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

// GetDailyStats retrieves daily statistics
func (r *auditRepository) GetDailyStats(startDate, endDate time.Time) ([]model.DailyStats, error) {
	var stats []model.DailyStats

	if err := r.db.Model(&model.AuditLog{}).
		Select("DATE(created_at) as date, COUNT(*) as total_logs, "+
			"SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success_logs, "+
			"SUM(CASE WHEN status = 'failure' THEN 1 ELSE 0 END) as failure_logs, "+
			"SUM(CASE WHEN status = 'error' THEN 1 ELSE 0 END) as error_logs, "+
			"COUNT(DISTINCT user_id) as unique_users").
		Where("deleted_at IS NULL AND created_at BETWEEN ? AND ?", startDate, endDate).
		Group("DATE(created_at)").
		Order("date DESC").
		Scan(&stats).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// GetHourlyStats retrieves hourly statistics
func (r *auditRepository) GetHourlyStats(startDate, endDate time.Time) ([]model.HourlyStats, error) {
	var stats []model.HourlyStats

	if err := r.db.Model(&model.AuditLog{}).
		Select("HOUR(created_at) as hour, COUNT(*) as total_logs, "+
			"SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success_logs, "+
			"SUM(CASE WHEN status = 'failure' THEN 1 ELSE 0 END) as failure_logs, "+
			"SUM(CASE WHEN status = 'error' THEN 1 ELSE 0 END) as error_logs, "+
			"COUNT(DISTINCT user_id) as unique_users").
		Where("deleted_at IS NULL AND created_at BETWEEN ? AND ?", startDate, endDate).
		Group("HOUR(created_at)").
		Order("hour ASC").
		Scan(&stats).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// GetRecentActivity retrieves recent audit log activity
func (r *auditRepository) GetRecentActivity(limit int) ([]model.AuditLog, error) {
	var logs []model.AuditLog

	if err := r.db.Where("deleted_at IS NULL").
		Preload("User").Preload("TargetUser").
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

// Export

// ExportAuditLogs exports audit logs in specified format
func (r *auditRepository) ExportAuditLogs(req *model.AuditExportRequest) (*model.AuditExportResponse, error) {
	// This is a placeholder implementation
	// In a real system, you would generate the export file and return download URL
	exportID := fmt.Sprintf("audit_export_%d", time.Now().Unix())

	response := &model.AuditExportResponse{
		ExportID:    exportID,
		Format:      req.Format,
		Status:      "processing",
		DownloadURL: fmt.Sprintf("/api/v1/audit/exports/%s/download", exportID),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		CreatedAt:   time.Now(),
	}

	return response, nil
}

// GetAuditLogsForExport retrieves audit logs for export
func (r *auditRepository) GetAuditLogsForExport(filters map[string]interface{}) ([]model.AuditLog, error) {
	var logs []model.AuditLog

	query := r.db.Model(&model.AuditLog{}).Where("deleted_at IS NULL")

	// Apply filters
	if userID, ok := filters["user_id"].(uint); ok && userID != 0 {
		query = query.Where("user_id = ?", userID)
	}
	if action, ok := filters["action"].(string); ok && action != "" {
		query = query.Where("action = ?", action)
	}
	if resource, ok := filters["resource"].(string); ok && resource != "" {
		query = query.Where("resource = ?", resource)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if startDate, ok := filters["start_date"].(time.Time); ok {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate, ok := filters["end_date"].(time.Time); ok {
		query = query.Where("created_at <= ?", endDate)
	}

	if err := query.Preload("User").Preload("TargetUser").
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

// Cleanup

// CleanupOldLogs cleans up old audit logs
func (r *auditRepository) CleanupOldLogs(retentionDays int) error {
	return r.DeleteOldLogs(retentionDays)
}

// CleanupOldSummaries cleans up old audit summaries
func (r *auditRepository) CleanupOldSummaries(retentionDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	return r.db.Where("date < ?", cutoffDate).Delete(&model.AuditLogSummary{}).Error
}

// OptimizeAuditTables optimizes audit log tables
func (r *auditRepository) OptimizeAuditTables() error {
	// This is a placeholder implementation
	// In a real system, you would run OPTIMIZE TABLE commands
	return nil
}

// Helper methods

// toAuditLogResponse converts AuditLog to AuditLogResponse
func (r *auditRepository) toAuditLogResponse(log *model.AuditLog) *model.AuditLogResponse {
	// Parse JSON fields (simplified - would need proper JSON unmarshaling)
	var oldValues map[string]interface{}
	var newValues map[string]interface{}
	var changes map[string]interface{}
	var metadata map[string]interface{}
	var tags []string

	// Convert User to User (simplified)
	var user *model.User
	if log.User != nil && log.User.ID != 0 {
		user = log.User
	}

	var targetUser *model.User
	if log.TargetUser != nil && log.TargetUser.ID != 0 {
		targetUser = log.TargetUser
	}

	return &model.AuditLogResponse{
		ID:           log.ID,
		UserID:       log.UserID,
		User:         user,
		Action:       log.Action,
		Resource:     log.Resource,
		ResourceID:   log.ResourceID,
		ResourceName: log.ResourceName,
		Operation:    log.Operation,
		Status:       log.Status,
		Message:      log.Message,
		IPAddress:    log.IPAddress,
		UserAgent:    log.UserAgent,
		Referer:      log.Referer,
		SessionID:    log.SessionID,
		OldValues:    oldValues,
		NewValues:    newValues,
		Changes:      changes,
		Metadata:     metadata,
		Tags:         tags,
		Severity:     log.Severity,
		TargetUserID: log.TargetUserID,
		TargetUser:   targetUser,
		CreatedAt:    log.CreatedAt,
		UpdatedAt:    log.UpdatedAt,
	}
}
