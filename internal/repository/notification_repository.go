package repository

import (
	"fmt"
	"go_app/internal/model"
	"go_app/pkg/database"
	"time"

	"gorm.io/gorm"
)

// NotificationRepository defines methods for notification operations
type NotificationRepository interface {
	// Notification CRUD
	CreateNotification(notification *model.Notification) error
	GetNotificationByID(id uint) (*model.Notification, error)
	GetNotificationsByUser(userID uint, req *model.NotificationListRequest) ([]*model.Notification, int64, error)
	UpdateNotification(notification *model.Notification) error
	DeleteNotification(id uint) error
	MarkAsRead(notificationID uint) error
	MarkAsUnread(notificationID uint) error
	MarkAsArchived(notificationID uint) error
	MarkAsUnarchived(notificationID uint) error
	BulkMarkAsRead(notificationIDs []uint) error
	BulkMarkAsArchived(notificationIDs []uint) error

	// Notification Templates
	CreateNotificationTemplate(template *model.NotificationTemplate) error
	GetNotificationTemplateByID(id uint) (*model.NotificationTemplate, error)
	GetNotificationTemplateByName(name string) (*model.NotificationTemplate, error)
	GetNotificationTemplates(req *model.NotificationListRequest) ([]*model.NotificationTemplate, int64, error)
	UpdateNotificationTemplate(template *model.NotificationTemplate) error
	DeleteNotificationTemplate(id uint) error
	GetTemplatesByTypeAndChannel(notificationType model.NotificationType, channel model.NotificationChannel) ([]*model.NotificationTemplate, error)

	// Notification Preferences
	CreateNotificationPreference(preference *model.NotificationPreference) error
	GetNotificationPreferenceByID(id uint) (*model.NotificationPreference, error)
	GetNotificationPreferencesByUser(userID uint) ([]*model.NotificationPreference, error)
	GetNotificationPreferenceByUserTypeChannel(userID uint, notificationType model.NotificationType, channel model.NotificationChannel) (*model.NotificationPreference, error)
	UpdateNotificationPreference(preference *model.NotificationPreference) error
	DeleteNotificationPreference(id uint) error
	DeleteNotificationPreferencesByUser(userID uint) error

	// Notification Queue
	AddToQueue(queueItem *model.NotificationQueue) error
	GetQueueItems(limit int) ([]*model.NotificationQueue, error)
	GetQueueItemsByStatus(status string, limit int) ([]*model.NotificationQueue, error)
	UpdateQueueItem(queueItem *model.NotificationQueue) error
	DeleteQueueItem(id uint) error
	DeleteProcessedQueueItems(olderThan time.Time) error

	// Notification Logs
	CreateNotificationLog(log *model.NotificationLog) error
	GetNotificationLogsByNotification(notificationID uint) ([]*model.NotificationLog, error)
	GetNotificationLogs(req *model.NotificationListRequest) ([]*model.NotificationLog, int64, error)
	DeleteNotificationLogs(olderThan time.Time) error

	// Notification Statistics
	GetNotificationStats(req *model.NotificationStatsRequest) (*model.NotificationStatsResponse, error)
	GetDailyNotificationStats(req *model.NotificationStatsRequest) ([]*model.DailyNotificationStats, error)
	GetTypeNotificationStats(req *model.NotificationStatsRequest) ([]*model.TypeNotificationStats, error)
	GetChannelNotificationStats(req *model.NotificationStatsRequest) ([]*model.ChannelNotificationStats, error)
	UpdateNotificationStats(stats *model.NotificationStats) error

	// Search and Filter
	SearchNotifications(query string, filters map[string]interface{}, page, limit int) ([]*model.Notification, int64, error)
	GetUnreadNotificationCount(userID uint) (int64, error)
	GetNotificationCountByType(userID uint, notificationType model.NotificationType) (int64, error)
	GetNotificationCountByChannel(userID uint, channel model.NotificationChannel) (int64, error)

	// Cleanup
	DeleteExpiredNotifications() error
	DeleteOldNotifications(olderThan time.Time) error
	ArchiveOldNotifications(olderThan time.Time) error
}

// notificationRepository implements NotificationRepository
type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new NotificationRepository
func NewNotificationRepository() NotificationRepository {
	return &notificationRepository{
		db: database.DB,
	}
}

// Notification CRUD

// CreateNotification creates a new notification
func (r *notificationRepository) CreateNotification(notification *model.Notification) error {
	return r.db.Create(notification).Error
}

// GetNotificationByID gets a notification by ID
func (r *notificationRepository) GetNotificationByID(id uint) (*model.Notification, error) {
	var notification model.Notification
	err := r.db.Preload("User").Where("id = ? AND deleted_at IS NULL", id).First(&notification).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// GetNotificationsByUser gets notifications for a user with pagination and filters
func (r *notificationRepository) GetNotificationsByUser(userID uint, req *model.NotificationListRequest) ([]*model.Notification, int64, error) {
	var notifications []*model.Notification
	var total int64

	query := r.db.Model(&model.Notification{}).Where("user_id = ? AND deleted_at IS NULL", userID)

	// Apply filters
	if req.Type != nil {
		query = query.Where("type = ?", *req.Type)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.Channel != nil {
		query = query.Where("channel = ?", *req.Channel)
	}
	if req.Priority != nil {
		query = query.Where("priority = ?", *req.Priority)
	}
	if req.IsRead != nil {
		query = query.Where("is_read = ?", *req.IsRead)
	}
	if req.IsArchived != nil {
		query = query.Where("is_archived = ?", *req.IsArchived)
	}
	if req.StartDate != nil {
		query = query.Where("created_at >= ?", *req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("created_at <= ?", *req.EndDate)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := "created_at"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}
	sortOrder := "desc"
	if req.SortOrder != "" {
		sortOrder = req.SortOrder
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	page := 1
	if req.Page > 0 {
		page = req.Page
	}
	limit := 20
	if req.Limit > 0 {
		limit = req.Limit
	}
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	// Execute query
	err := query.Preload("User").Find(&notifications).Error
	return notifications, total, err
}

// UpdateNotification updates a notification
func (r *notificationRepository) UpdateNotification(notification *model.Notification) error {
	return r.db.Save(notification).Error
}

// DeleteNotification deletes a notification (soft delete)
func (r *notificationRepository) DeleteNotification(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.Notification{}).Error
}

// MarkAsRead marks a notification as read
func (r *notificationRepository) MarkAsRead(notificationID uint) error {
	now := time.Now()
	return r.db.Model(&model.Notification{}).Where("id = ?", notificationID).Updates(map[string]interface{}{
		"is_read": true,
		"read_at": &now,
	}).Error
}

// MarkAsUnread marks a notification as unread
func (r *notificationRepository) MarkAsUnread(notificationID uint) error {
	return r.db.Model(&model.Notification{}).Where("id = ?", notificationID).Updates(map[string]interface{}{
		"is_read": false,
		"read_at": nil,
	}).Error
}

// MarkAsArchived marks a notification as archived
func (r *notificationRepository) MarkAsArchived(notificationID uint) error {
	return r.db.Model(&model.Notification{}).Where("id = ?", notificationID).Update("is_archived", true).Error
}

// MarkAsUnarchived marks a notification as unarchived
func (r *notificationRepository) MarkAsUnarchived(notificationID uint) error {
	return r.db.Model(&model.Notification{}).Where("id = ?", notificationID).Update("is_archived", false).Error
}

// BulkMarkAsRead marks multiple notifications as read
func (r *notificationRepository) BulkMarkAsRead(notificationIDs []uint) error {
	now := time.Now()
	return r.db.Model(&model.Notification{}).Where("id IN ?", notificationIDs).Updates(map[string]interface{}{
		"is_read": true,
		"read_at": &now,
	}).Error
}

// BulkMarkAsArchived marks multiple notifications as archived
func (r *notificationRepository) BulkMarkAsArchived(notificationIDs []uint) error {
	return r.db.Model(&model.Notification{}).Where("id IN ?", notificationIDs).Update("is_archived", true).Error
}

// Notification Templates

// CreateNotificationTemplate creates a new notification template
func (r *notificationRepository) CreateNotificationTemplate(template *model.NotificationTemplate) error {
	return r.db.Create(template).Error
}

// GetNotificationTemplateByID gets a notification template by ID
func (r *notificationRepository) GetNotificationTemplateByID(id uint) (*model.NotificationTemplate, error) {
	var template model.NotificationTemplate
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// GetNotificationTemplateByName gets a notification template by name
func (r *notificationRepository) GetNotificationTemplateByName(name string) (*model.NotificationTemplate, error) {
	var template model.NotificationTemplate
	err := r.db.Where("name = ? AND deleted_at IS NULL", name).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// GetNotificationTemplates gets notification templates with pagination
func (r *notificationRepository) GetNotificationTemplates(req *model.NotificationListRequest) ([]*model.NotificationTemplate, int64, error) {
	var templates []*model.NotificationTemplate
	var total int64

	query := r.db.Model(&model.NotificationTemplate{}).Where("deleted_at IS NULL")

	// Apply filters
	if req.Type != nil {
		query = query.Where("type = ?", *req.Type)
	}
	if req.Channel != nil {
		query = query.Where("channel = ?", *req.Channel)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := "created_at"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}
	sortOrder := "desc"
	if req.SortOrder != "" {
		sortOrder = req.SortOrder
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	page := 1
	if req.Page > 0 {
		page = req.Page
	}
	limit := 20
	if req.Limit > 0 {
		limit = req.Limit
	}
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	// Execute query
	err := query.Find(&templates).Error
	return templates, total, err
}

// UpdateNotificationTemplate updates a notification template
func (r *notificationRepository) UpdateNotificationTemplate(template *model.NotificationTemplate) error {
	return r.db.Save(template).Error
}

// DeleteNotificationTemplate deletes a notification template (soft delete)
func (r *notificationRepository) DeleteNotificationTemplate(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.NotificationTemplate{}).Error
}

// GetTemplatesByTypeAndChannel gets templates by type and channel
func (r *notificationRepository) GetTemplatesByTypeAndChannel(notificationType model.NotificationType, channel model.NotificationChannel) ([]*model.NotificationTemplate, error) {
	var templates []*model.NotificationTemplate
	err := r.db.Where("type = ? AND channel = ? AND is_active = ? AND deleted_at IS NULL", notificationType, channel, true).Find(&templates).Error
	return templates, err
}

// Notification Preferences

// CreateNotificationPreference creates a new notification preference
func (r *notificationRepository) CreateNotificationPreference(preference *model.NotificationPreference) error {
	return r.db.Create(preference).Error
}

// GetNotificationPreferenceByID gets a notification preference by ID
func (r *notificationRepository) GetNotificationPreferenceByID(id uint) (*model.NotificationPreference, error) {
	var preference model.NotificationPreference
	err := r.db.Preload("User").Where("id = ? AND deleted_at IS NULL", id).First(&preference).Error
	if err != nil {
		return nil, err
	}
	return &preference, nil
}

// GetNotificationPreferencesByUser gets notification preferences for a user
func (r *notificationRepository) GetNotificationPreferencesByUser(userID uint) ([]*model.NotificationPreference, error) {
	var preferences []*model.NotificationPreference
	err := r.db.Preload("User").Where("user_id = ? AND deleted_at IS NULL", userID).Find(&preferences).Error
	return preferences, err
}

// GetNotificationPreferenceByUserTypeChannel gets a specific notification preference
func (r *notificationRepository) GetNotificationPreferenceByUserTypeChannel(userID uint, notificationType model.NotificationType, channel model.NotificationChannel) (*model.NotificationPreference, error) {
	var preference model.NotificationPreference
	err := r.db.Preload("User").Where("user_id = ? AND type = ? AND channel = ? AND deleted_at IS NULL", userID, notificationType, channel).First(&preference).Error
	if err != nil {
		return nil, err
	}
	return &preference, nil
}

// UpdateNotificationPreference updates a notification preference
func (r *notificationRepository) UpdateNotificationPreference(preference *model.NotificationPreference) error {
	return r.db.Save(preference).Error
}

// DeleteNotificationPreference deletes a notification preference (soft delete)
func (r *notificationRepository) DeleteNotificationPreference(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.NotificationPreference{}).Error
}

// DeleteNotificationPreferencesByUser deletes all notification preferences for a user
func (r *notificationRepository) DeleteNotificationPreferencesByUser(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.NotificationPreference{}).Error
}

// Notification Queue

// AddToQueue adds a notification to the queue
func (r *notificationRepository) AddToQueue(queueItem *model.NotificationQueue) error {
	return r.db.Create(queueItem).Error
}

// GetQueueItems gets queue items for processing
func (r *notificationRepository) GetQueueItems(limit int) ([]*model.NotificationQueue, error) {
	var queueItems []*model.NotificationQueue
	err := r.db.Preload("Notification").Where("status = ? AND scheduled_at <= ? AND deleted_at IS NULL", "pending", time.Now()).Order("priority DESC, scheduled_at ASC").Limit(limit).Find(&queueItems).Error
	return queueItems, err
}

// GetQueueItemsByStatus gets queue items by status
func (r *notificationRepository) GetQueueItemsByStatus(status string, limit int) ([]*model.NotificationQueue, error) {
	var queueItems []*model.NotificationQueue
	err := r.db.Preload("Notification").Where("status = ? AND deleted_at IS NULL", status).Order("created_at ASC").Limit(limit).Find(&queueItems).Error
	return queueItems, err
}

// UpdateQueueItem updates a queue item
func (r *notificationRepository) UpdateQueueItem(queueItem *model.NotificationQueue) error {
	return r.db.Save(queueItem).Error
}

// DeleteQueueItem deletes a queue item
func (r *notificationRepository) DeleteQueueItem(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.NotificationQueue{}).Error
}

// DeleteProcessedQueueItems deletes processed queue items older than specified time
func (r *notificationRepository) DeleteProcessedQueueItems(olderThan time.Time) error {
	return r.db.Where("status = ? AND processed_at < ?", "completed", olderThan).Delete(&model.NotificationQueue{}).Error
}

// Notification Logs

// CreateNotificationLog creates a new notification log
func (r *notificationRepository) CreateNotificationLog(log *model.NotificationLog) error {
	return r.db.Create(log).Error
}

// GetNotificationLogsByNotification gets logs for a specific notification
func (r *notificationRepository) GetNotificationLogsByNotification(notificationID uint) ([]*model.NotificationLog, error) {
	var logs []*model.NotificationLog
	err := r.db.Preload("Notification").Where("notification_id = ? AND deleted_at IS NULL", notificationID).Order("attempted_at DESC").Find(&logs).Error
	return logs, err
}

// GetNotificationLogs gets notification logs with pagination
func (r *notificationRepository) GetNotificationLogs(req *model.NotificationListRequest) ([]*model.NotificationLog, int64, error) {
	var logs []*model.NotificationLog
	var total int64

	query := r.db.Model(&model.NotificationLog{}).Where("deleted_at IS NULL")

	// Apply filters
	if req.Channel != nil {
		query = query.Where("channel = ?", *req.Channel)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.StartDate != nil {
		query = query.Where("attempted_at >= ?", *req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("attempted_at <= ?", *req.EndDate)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := "attempted_at"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}
	sortOrder := "desc"
	if req.SortOrder != "" {
		sortOrder = req.SortOrder
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	page := 1
	if req.Page > 0 {
		page = req.Page
	}
	limit := 20
	if req.Limit > 0 {
		limit = req.Limit
	}
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	// Execute query
	err := query.Preload("Notification").Find(&logs).Error
	return logs, total, err
}

// DeleteNotificationLogs deletes old notification logs
func (r *notificationRepository) DeleteNotificationLogs(olderThan time.Time) error {
	return r.db.Where("attempted_at < ?", olderThan).Delete(&model.NotificationLog{}).Error
}

// Notification Statistics

// GetNotificationStats gets notification statistics
func (r *notificationRepository) GetNotificationStats(req *model.NotificationStatsRequest) (*model.NotificationStatsResponse, error) {
	var stats model.NotificationStatsResponse

	// Base query
	query := r.db.Model(&model.Notification{}).Where("deleted_at IS NULL")

	// Apply date filters
	if req.StartDate != nil {
		query = query.Where("created_at >= ?", *req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("created_at <= ?", *req.EndDate)
	}
	if req.Type != nil {
		query = query.Where("type = ?", *req.Type)
	}
	if req.Channel != nil {
		query = query.Where("channel = ?", *req.Channel)
	}

	// Get basic stats
	var totalSent, totalDelivered, totalRead, totalFailed int64
	query.Count(&totalSent)
	query.Where("status = ?", model.NotificationStatusDelivered).Count(&totalDelivered)
	query.Where("is_read = ?", true).Count(&totalRead)
	query.Where("status = ?", model.NotificationStatusFailed).Count(&totalFailed)

	stats.TotalSent = totalSent
	stats.TotalDelivered = totalDelivered
	stats.TotalRead = totalRead
	stats.TotalFailed = totalFailed

	// Calculate rates
	if totalSent > 0 {
		stats.DeliveryRate = float64(totalDelivered) / float64(totalSent) * 100
		stats.ReadRate = float64(totalRead) / float64(totalSent) * 100
	}

	// Get daily stats
	dailyStats, err := r.GetDailyNotificationStats(req)
	if err != nil {
		return nil, err
	}
	// Convert to slice of values
	var dailyStatsValues []model.DailyNotificationStats
	for _, stat := range dailyStats {
		dailyStatsValues = append(dailyStatsValues, *stat)
	}
	stats.DailyStats = dailyStatsValues

	// Get type stats
	typeStats, err := r.GetTypeNotificationStats(req)
	if err != nil {
		return nil, err
	}
	// Convert to slice of values
	var typeStatsValues []model.TypeNotificationStats
	for _, stat := range typeStats {
		typeStatsValues = append(typeStatsValues, *stat)
	}
	stats.TypeStats = typeStatsValues

	// Get channel stats
	channelStats, err := r.GetChannelNotificationStats(req)
	if err != nil {
		return nil, err
	}
	// Convert to slice of values
	var channelStatsValues []model.ChannelNotificationStats
	for _, stat := range channelStats {
		channelStatsValues = append(channelStatsValues, *stat)
	}
	stats.ChannelStats = channelStatsValues

	return &stats, nil
}

// GetDailyNotificationStats gets daily notification statistics
func (r *notificationRepository) GetDailyNotificationStats(req *model.NotificationStatsRequest) ([]*model.DailyNotificationStats, error) {
	var stats []*model.DailyNotificationStats

	query := r.db.Model(&model.Notification{}).Select(
		"DATE(created_at) as date",
		"COUNT(*) as total_sent",
		"SUM(CASE WHEN status = 'delivered' THEN 1 ELSE 0 END) as total_delivered",
		"SUM(CASE WHEN is_read = true THEN 1 ELSE 0 END) as total_read",
		"SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as total_failed",
	).Where("deleted_at IS NULL").Group("DATE(created_at)")

	// Apply filters
	if req.StartDate != nil {
		query = query.Where("created_at >= ?", *req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("created_at <= ?", *req.EndDate)
	}
	if req.Type != nil {
		query = query.Where("type = ?", *req.Type)
	}
	if req.Channel != nil {
		query = query.Where("channel = ?", *req.Channel)
	}

	err := query.Order("date DESC").Find(&stats).Error
	return stats, err
}

// GetTypeNotificationStats gets notification statistics by type
func (r *notificationRepository) GetTypeNotificationStats(req *model.NotificationStatsRequest) ([]*model.TypeNotificationStats, error) {
	var stats []*model.TypeNotificationStats

	query := r.db.Model(&model.Notification{}).Select(
		"type",
		"COUNT(*) as total_sent",
		"SUM(CASE WHEN status = 'delivered' THEN 1 ELSE 0 END) as total_delivered",
		"SUM(CASE WHEN is_read = true THEN 1 ELSE 0 END) as total_read",
		"SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as total_failed",
	).Where("deleted_at IS NULL").Group("type")

	// Apply filters
	if req.StartDate != nil {
		query = query.Where("created_at >= ?", *req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("created_at <= ?", *req.EndDate)
	}
	if req.Type != nil {
		query = query.Where("type = ?", *req.Type)
	}
	if req.Channel != nil {
		query = query.Where("channel = ?", *req.Channel)
	}

	err := query.Order("total_sent DESC").Find(&stats).Error
	return stats, err
}

// GetChannelNotificationStats gets notification statistics by channel
func (r *notificationRepository) GetChannelNotificationStats(req *model.NotificationStatsRequest) ([]*model.ChannelNotificationStats, error) {
	var stats []*model.ChannelNotificationStats

	query := r.db.Model(&model.Notification{}).Select(
		"channel",
		"COUNT(*) as total_sent",
		"SUM(CASE WHEN status = 'delivered' THEN 1 ELSE 0 END) as total_delivered",
		"SUM(CASE WHEN is_read = true THEN 1 ELSE 0 END) as total_read",
		"SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as total_failed",
	).Where("deleted_at IS NULL").Group("channel")

	// Apply filters
	if req.StartDate != nil {
		query = query.Where("created_at >= ?", *req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("created_at <= ?", *req.EndDate)
	}
	if req.Type != nil {
		query = query.Where("type = ?", *req.Type)
	}
	if req.Channel != nil {
		query = query.Where("channel = ?", *req.Channel)
	}

	err := query.Order("total_sent DESC").Find(&stats).Error
	return stats, err
}

// UpdateNotificationStats updates notification statistics
func (r *notificationRepository) UpdateNotificationStats(stats *model.NotificationStats) error {
	return r.db.Save(stats).Error
}

// Search and Filter

// SearchNotifications searches notifications
func (r *notificationRepository) SearchNotifications(query string, filters map[string]interface{}, page, limit int) ([]*model.Notification, int64, error) {
	var notifications []*model.Notification
	var total int64

	dbQuery := r.db.Model(&model.Notification{}).Where("deleted_at IS NULL")

	// Apply text search
	if query != "" {
		dbQuery = dbQuery.Where("MATCH(title, message) AGAINST(? IN NATURAL LANGUAGE MODE)", query)
	}

	// Apply filters
	for key, value := range filters {
		switch key {
		case "user_id":
			if userID, ok := value.(uint); ok {
				dbQuery = dbQuery.Where("user_id = ?", userID)
			}
		case "type":
			if notificationType, ok := value.(model.NotificationType); ok {
				dbQuery = dbQuery.Where("type = ?", notificationType)
			}
		case "status":
			if status, ok := value.(model.NotificationStatus); ok {
				dbQuery = dbQuery.Where("status = ?", status)
			}
		case "channel":
			if channel, ok := value.(model.NotificationChannel); ok {
				dbQuery = dbQuery.Where("channel = ?", channel)
			}
		case "priority":
			if priority, ok := value.(model.NotificationPriority); ok {
				dbQuery = dbQuery.Where("priority = ?", priority)
			}
		case "is_read":
			if isRead, ok := value.(bool); ok {
				dbQuery = dbQuery.Where("is_read = ?", isRead)
			}
		case "is_archived":
			if isArchived, ok := value.(bool); ok {
				dbQuery = dbQuery.Where("is_archived = ?", isArchived)
			}
		}
	}

	// Get total count
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * limit
	err := dbQuery.Preload("User").Order("created_at DESC").Offset(offset).Limit(limit).Find(&notifications).Error
	return notifications, total, err
}

// GetUnreadNotificationCount gets unread notification count for a user
func (r *notificationRepository) GetUnreadNotificationCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Notification{}).Where("user_id = ? AND is_read = ? AND deleted_at IS NULL", userID, false).Count(&count).Error
	return count, err
}

// GetNotificationCountByType gets notification count by type for a user
func (r *notificationRepository) GetNotificationCountByType(userID uint, notificationType model.NotificationType) (int64, error) {
	var count int64
	err := r.db.Model(&model.Notification{}).Where("user_id = ? AND type = ? AND deleted_at IS NULL", userID, notificationType).Count(&count).Error
	return count, err
}

// GetNotificationCountByChannel gets notification count by channel for a user
func (r *notificationRepository) GetNotificationCountByChannel(userID uint, channel model.NotificationChannel) (int64, error) {
	var count int64
	err := r.db.Model(&model.Notification{}).Where("user_id = ? AND channel = ? AND deleted_at IS NULL", userID, channel).Count(&count).Error
	return count, err
}

// Cleanup

// DeleteExpiredNotifications deletes expired notifications
func (r *notificationRepository) DeleteExpiredNotifications() error {
	return r.db.Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).Delete(&model.Notification{}).Error
}

// DeleteOldNotifications deletes old notifications
func (r *notificationRepository) DeleteOldNotifications(olderThan time.Time) error {
	return r.db.Where("created_at < ?", olderThan).Delete(&model.Notification{}).Error
}

// ArchiveOldNotifications archives old notifications
func (r *notificationRepository) ArchiveOldNotifications(olderThan time.Time) error {
	return r.db.Model(&model.Notification{}).Where("created_at < ? AND is_archived = ?", olderThan, false).Update("is_archived", true).Error
}
