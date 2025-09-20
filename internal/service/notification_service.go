package service

import (
	"encoding/json"
	"fmt"
	"go_app/internal/model"
	"go_app/internal/repository"
	"go_app/pkg/logger"
	"time"

	"gorm.io/gorm"
)

// PaginationResponse represents pagination information
type PaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// NotificationService defines methods for notification business logic
type NotificationService interface {
	// Notification CRUD
	CreateNotification(req *model.CreateNotificationRequest) (*model.NotificationResponse, error)
	GetNotificationByID(id uint) (*model.NotificationResponse, error)
	GetNotificationsByUser(userID uint, req *model.NotificationListRequest) ([]*model.NotificationResponse, *PaginationResponse, error)
	UpdateNotification(id uint, req *model.UpdateNotificationRequest) (*model.NotificationResponse, error)
	DeleteNotification(id uint) error
	MarkAsRead(notificationID uint) error
	MarkAsUnread(notificationID uint) error
	MarkAsArchived(notificationID uint) error
	MarkAsUnarchived(notificationID uint) error
	BulkMarkAsRead(notificationIDs []uint) error
	BulkMarkAsArchived(notificationIDs []uint) error

	// Notification Templates
	CreateNotificationTemplate(req *model.CreateNotificationRequest) (*model.NotificationTemplateResponse, error)
	GetNotificationTemplateByID(id uint) (*model.NotificationTemplateResponse, error)
	GetNotificationTemplateByName(name string) (*model.NotificationTemplateResponse, error)
	GetNotificationTemplates(req *model.NotificationListRequest) ([]*model.NotificationTemplateResponse, *PaginationResponse, error)
	UpdateNotificationTemplate(id uint, req *model.CreateNotificationRequest) (*model.NotificationTemplateResponse, error)
	DeleteNotificationTemplate(id uint) error

	// Notification Preferences
	CreateNotificationPreference(userID uint, req *model.NotificationPreferenceRequest) (*model.NotificationPreferenceResponse, error)
	GetNotificationPreferenceByID(id uint) (*model.NotificationPreferenceResponse, error)
	GetNotificationPreferencesByUser(userID uint) ([]*model.NotificationPreferenceResponse, error)
	UpdateNotificationPreference(id uint, req *model.NotificationPreferenceRequest) (*model.NotificationPreferenceResponse, error)
	DeleteNotificationPreference(id uint) error

	// Notification Statistics
	GetNotificationStats(req *model.NotificationStatsRequest) (*model.NotificationStatsResponse, error)
	GetUnreadNotificationCount(userID uint) (int64, error)
	GetNotificationCountByType(userID uint, notificationType model.NotificationType) (int64, error)
	GetNotificationCountByChannel(userID uint, channel model.NotificationChannel) (int64, error)

	// Search and Filter
	SearchNotifications(query string, filters map[string]interface{}, page, limit int) ([]*model.NotificationResponse, *PaginationResponse, error)

	// Notification Processing
	ProcessNotificationQueue() error
	SendNotification(notification *model.Notification) error
	ScheduleNotification(notification *model.Notification, scheduledAt time.Time) error

	// Cleanup
	DeleteExpiredNotifications() error
	DeleteOldNotifications(olderThan time.Time) error
	ArchiveOldNotifications(olderThan time.Time) error
}

// notificationService implements NotificationService
type notificationService struct {
	notificationRepo repository.NotificationRepository
	userRepo         repository.UserRepository
}

// NewNotificationService creates a new NotificationService
func NewNotificationService(notificationRepo repository.NotificationRepository, userRepo repository.UserRepository) NotificationService {
	return &notificationService{
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
	}
}

// Notification CRUD

// CreateNotification creates a new notification
func (s *notificationService) CreateNotification(req *model.CreateNotificationRequest) (*model.NotificationResponse, error) {
	// Convert data to JSON string
	var dataJSON string
	if req.Data != nil {
		dataBytes, err := json.Marshal(req.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data: %w", err)
		}
		dataJSON = string(dataBytes)
	}

	// Set default priority if not provided
	priority := req.Priority
	if priority == "" {
		priority = model.NotificationPriorityNormal
	}

	// Create notification
	notification := &model.Notification{
		UserID:      req.UserID,
		Type:        req.Type,
		Priority:    priority,
		Status:      model.NotificationStatusPending,
		Channel:     req.Channel,
		Title:       req.Title,
		Message:     req.Message,
		Data:        dataJSON,
		ActionURL:   req.ActionURL,
		ImageURL:    req.ImageURL,
		ExpiresAt:   req.ExpiresAt,
		ScheduledAt: req.ScheduledAt,
	}

	// Save to database
	if err := s.notificationRepo.CreateNotification(notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// If scheduled, add to queue
	if req.ScheduledAt != nil {
		queueItem := &model.NotificationQueue{
			NotificationID: notification.ID,
			Priority:       notification.Priority,
			Channel:        notification.Channel,
			ScheduledAt:    *req.ScheduledAt,
			Status:         "pending",
		}
		if err := s.notificationRepo.AddToQueue(queueItem); err != nil {
			logger.Warnf("Failed to add notification to queue: %v", err)
		}
	} else {
		// Send immediately
		go s.SendNotification(notification)
	}

	return notification.ToResponse(), nil
}

// GetNotificationByID gets a notification by ID
func (s *notificationService) GetNotificationByID(id uint) (*model.NotificationResponse, error) {
	notification, err := s.notificationRepo.GetNotificationByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("notification not found")
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}
	return notification.ToResponse(), nil
}

// GetNotificationsByUser gets notifications for a user
func (s *notificationService) GetNotificationsByUser(userID uint, req *model.NotificationListRequest) ([]*model.NotificationResponse, *PaginationResponse, error) {
	notifications, total, err := s.notificationRepo.GetNotificationsByUser(userID, req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	// Convert to response
	var responses []*model.NotificationResponse
	for _, notification := range notifications {
		responses = append(responses, notification.ToResponse())
	}

	// Calculate pagination
	page := 1
	if req.Page > 0 {
		page = req.Page
	}
	limit := 20
	if req.Limit > 0 {
		limit = req.Limit
	}
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	pagination := &PaginationResponse{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	return responses, pagination, nil
}

// UpdateNotification updates a notification
func (s *notificationService) UpdateNotification(id uint, req *model.UpdateNotificationRequest) (*model.NotificationResponse, error) {
	notification, err := s.notificationRepo.GetNotificationByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("notification not found")
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	// Update fields
	if req.Status != nil {
		notification.Status = *req.Status
	}
	if req.IsRead != nil {
		notification.IsRead = *req.IsRead
		if *req.IsRead && notification.ReadAt == nil {
			now := time.Now()
			notification.ReadAt = &now
		} else if !*req.IsRead {
			notification.ReadAt = nil
		}
	}
	if req.IsArchived != nil {
		notification.IsArchived = *req.IsArchived
	}
	if req.RetryCount != nil {
		notification.RetryCount = *req.RetryCount
	}
	if req.ErrorMsg != nil {
		notification.ErrorMsg = *req.ErrorMsg
	}

	// Save changes
	if err := s.notificationRepo.UpdateNotification(notification); err != nil {
		return nil, fmt.Errorf("failed to update notification: %w", err)
	}

	return notification.ToResponse(), nil
}

// DeleteNotification deletes a notification
func (s *notificationService) DeleteNotification(id uint) error {
	// Check if notification exists
	_, err := s.notificationRepo.GetNotificationByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("notification not found")
		}
		return fmt.Errorf("failed to get notification: %w", err)
	}

	// Delete notification
	if err := s.notificationRepo.DeleteNotification(id); err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	return nil
}

// MarkAsRead marks a notification as read
func (s *notificationService) MarkAsRead(notificationID uint) error {
	if err := s.notificationRepo.MarkAsRead(notificationID); err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}
	return nil
}

// MarkAsUnread marks a notification as unread
func (s *notificationService) MarkAsUnread(notificationID uint) error {
	if err := s.notificationRepo.MarkAsUnread(notificationID); err != nil {
		return fmt.Errorf("failed to mark notification as unread: %w", err)
	}
	return nil
}

// MarkAsArchived marks a notification as archived
func (s *notificationService) MarkAsArchived(notificationID uint) error {
	if err := s.notificationRepo.MarkAsArchived(notificationID); err != nil {
		return fmt.Errorf("failed to mark notification as archived: %w", err)
	}
	return nil
}

// MarkAsUnarchived marks a notification as unarchived
func (s *notificationService) MarkAsUnarchived(notificationID uint) error {
	if err := s.notificationRepo.MarkAsUnarchived(notificationID); err != nil {
		return fmt.Errorf("failed to mark notification as unarchived: %w", err)
	}
	return nil
}

// BulkMarkAsRead marks multiple notifications as read
func (s *notificationService) BulkMarkAsRead(notificationIDs []uint) error {
	if err := s.notificationRepo.BulkMarkAsRead(notificationIDs); err != nil {
		return fmt.Errorf("failed to bulk mark notifications as read: %w", err)
	}
	return nil
}

// BulkMarkAsArchived marks multiple notifications as archived
func (s *notificationService) BulkMarkAsArchived(notificationIDs []uint) error {
	if err := s.notificationRepo.BulkMarkAsArchived(notificationIDs); err != nil {
		return fmt.Errorf("failed to bulk mark notifications as archived: %w", err)
	}
	return nil
}

// Notification Templates

// CreateNotificationTemplate creates a new notification template
func (s *notificationService) CreateNotificationTemplate(req *model.CreateNotificationRequest) (*model.NotificationTemplateResponse, error) {
	// Convert data to JSON string
	var variablesJSON string
	if req.Data != nil {
		variablesBytes, err := json.Marshal(req.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal variables: %w", err)
		}
		variablesJSON = string(variablesBytes)
	}

	// Create template
	template := &model.NotificationTemplate{
		Name:        req.Title, // Using title as template name
		Type:        req.Type,
		Channel:     req.Channel,
		Subject:     req.Title,
		Body:        req.Message,
		Variables:   variablesJSON,
		IsActive:    true,
		IsSystem:    false,
		Description: req.ActionURL, // Using action URL as description
	}

	// Save to database
	if err := s.notificationRepo.CreateNotificationTemplate(template); err != nil {
		return nil, fmt.Errorf("failed to create notification template: %w", err)
	}

	return template.ToResponse(), nil
}

// GetNotificationTemplateByID gets a notification template by ID
func (s *notificationService) GetNotificationTemplateByID(id uint) (*model.NotificationTemplateResponse, error) {
	template, err := s.notificationRepo.GetNotificationTemplateByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("notification template not found")
		}
		return nil, fmt.Errorf("failed to get notification template: %w", err)
	}
	return template.ToResponse(), nil
}

// GetNotificationTemplateByName gets a notification template by name
func (s *notificationService) GetNotificationTemplateByName(name string) (*model.NotificationTemplateResponse, error) {
	template, err := s.notificationRepo.GetNotificationTemplateByName(name)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("notification template not found")
		}
		return nil, fmt.Errorf("failed to get notification template: %w", err)
	}
	return template.ToResponse(), nil
}

// GetNotificationTemplates gets notification templates
func (s *notificationService) GetNotificationTemplates(req *model.NotificationListRequest) ([]*model.NotificationTemplateResponse, *PaginationResponse, error) {
	templates, total, err := s.notificationRepo.GetNotificationTemplates(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get notification templates: %w", err)
	}

	// Convert to response
	var responses []*model.NotificationTemplateResponse
	for _, template := range templates {
		responses = append(responses, template.ToResponse())
	}

	// Calculate pagination
	page := 1
	if req.Page > 0 {
		page = req.Page
	}
	limit := 20
	if req.Limit > 0 {
		limit = req.Limit
	}
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	pagination := &PaginationResponse{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	return responses, pagination, nil
}

// UpdateNotificationTemplate updates a notification template
func (s *notificationService) UpdateNotificationTemplate(id uint, req *model.CreateNotificationRequest) (*model.NotificationTemplateResponse, error) {
	template, err := s.notificationRepo.GetNotificationTemplateByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("notification template not found")
		}
		return nil, fmt.Errorf("failed to get notification template: %w", err)
	}

	// Convert data to JSON string
	var variablesJSON string
	if req.Data != nil {
		variablesBytes, err := json.Marshal(req.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal variables: %w", err)
		}
		variablesJSON = string(variablesBytes)
	}

	// Update fields
	template.Name = req.Title
	template.Type = req.Type
	template.Channel = req.Channel
	template.Subject = req.Title
	template.Body = req.Message
	template.Variables = variablesJSON
	template.Description = req.ActionURL

	// Save changes
	if err := s.notificationRepo.UpdateNotificationTemplate(template); err != nil {
		return nil, fmt.Errorf("failed to update notification template: %w", err)
	}

	return template.ToResponse(), nil
}

// DeleteNotificationTemplate deletes a notification template
func (s *notificationService) DeleteNotificationTemplate(id uint) error {
	// Check if template exists
	_, err := s.notificationRepo.GetNotificationTemplateByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("notification template not found")
		}
		return fmt.Errorf("failed to get notification template: %w", err)
	}

	// Delete template
	if err := s.notificationRepo.DeleteNotificationTemplate(id); err != nil {
		return fmt.Errorf("failed to delete notification template: %w", err)
	}

	return nil
}

// Notification Preferences

// CreateNotificationPreference creates a new notification preference
func (s *notificationService) CreateNotificationPreference(userID uint, req *model.NotificationPreferenceRequest) (*model.NotificationPreferenceResponse, error) {
	// Check if preference already exists
	existing, err := s.notificationRepo.GetNotificationPreferenceByUserTypeChannel(userID, req.Type, req.Channel)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("notification preference already exists")
	}

	// Create preference
	preference := &model.NotificationPreference{
		UserID:     userID,
		Type:       req.Type,
		Channel:    req.Channel,
		IsEnabled:  req.IsEnabled,
		Frequency:  req.Frequency,
		QuietHours: req.QuietHours,
		Timezone:   req.Timezone,
	}

	// Save to database
	if err := s.notificationRepo.CreateNotificationPreference(preference); err != nil {
		return nil, fmt.Errorf("failed to create notification preference: %w", err)
	}

	return preference.ToResponse(), nil
}

// GetNotificationPreferenceByID gets a notification preference by ID
func (s *notificationService) GetNotificationPreferenceByID(id uint) (*model.NotificationPreferenceResponse, error) {
	preference, err := s.notificationRepo.GetNotificationPreferenceByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("notification preference not found")
		}
		return nil, fmt.Errorf("failed to get notification preference: %w", err)
	}
	return preference.ToResponse(), nil
}

// GetNotificationPreferencesByUser gets notification preferences for a user
func (s *notificationService) GetNotificationPreferencesByUser(userID uint) ([]*model.NotificationPreferenceResponse, error) {
	preferences, err := s.notificationRepo.GetNotificationPreferencesByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification preferences: %w", err)
	}

	// Convert to response
	var responses []*model.NotificationPreferenceResponse
	for _, preference := range preferences {
		responses = append(responses, preference.ToResponse())
	}

	return responses, nil
}

// UpdateNotificationPreference updates a notification preference
func (s *notificationService) UpdateNotificationPreference(id uint, req *model.NotificationPreferenceRequest) (*model.NotificationPreferenceResponse, error) {
	preference, err := s.notificationRepo.GetNotificationPreferenceByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("notification preference not found")
		}
		return nil, fmt.Errorf("failed to get notification preference: %w", err)
	}

	// Update fields
	preference.Type = req.Type
	preference.Channel = req.Channel
	preference.IsEnabled = req.IsEnabled
	preference.Frequency = req.Frequency
	preference.QuietHours = req.QuietHours
	preference.Timezone = req.Timezone

	// Save changes
	if err := s.notificationRepo.UpdateNotificationPreference(preference); err != nil {
		return nil, fmt.Errorf("failed to update notification preference: %w", err)
	}

	return preference.ToResponse(), nil
}

// DeleteNotificationPreference deletes a notification preference
func (s *notificationService) DeleteNotificationPreference(id uint) error {
	// Check if preference exists
	_, err := s.notificationRepo.GetNotificationPreferenceByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("notification preference not found")
		}
		return fmt.Errorf("failed to get notification preference: %w", err)
	}

	// Delete preference
	if err := s.notificationRepo.DeleteNotificationPreference(id); err != nil {
		return fmt.Errorf("failed to delete notification preference: %w", err)
	}

	return nil
}

// Notification Statistics

// GetNotificationStats gets notification statistics
func (s *notificationService) GetNotificationStats(req *model.NotificationStatsRequest) (*model.NotificationStatsResponse, error) {
	stats, err := s.notificationRepo.GetNotificationStats(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification statistics: %w", err)
	}
	return stats, nil
}

// GetUnreadNotificationCount gets unread notification count for a user
func (s *notificationService) GetUnreadNotificationCount(userID uint) (int64, error) {
	count, err := s.notificationRepo.GetUnreadNotificationCount(userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread notification count: %w", err)
	}
	return count, nil
}

// GetNotificationCountByType gets notification count by type for a user
func (s *notificationService) GetNotificationCountByType(userID uint, notificationType model.NotificationType) (int64, error) {
	count, err := s.notificationRepo.GetNotificationCountByType(userID, notificationType)
	if err != nil {
		return 0, fmt.Errorf("failed to get notification count by type: %w", err)
	}
	return count, nil
}

// GetNotificationCountByChannel gets notification count by channel for a user
func (s *notificationService) GetNotificationCountByChannel(userID uint, channel model.NotificationChannel) (int64, error) {
	count, err := s.notificationRepo.GetNotificationCountByChannel(userID, channel)
	if err != nil {
		return 0, fmt.Errorf("failed to get notification count by channel: %w", err)
	}
	return count, nil
}

// Search and Filter

// SearchNotifications searches notifications
func (s *notificationService) SearchNotifications(query string, filters map[string]interface{}, page, limit int) ([]*model.NotificationResponse, *PaginationResponse, error) {
	notifications, total, err := s.notificationRepo.SearchNotifications(query, filters, page, limit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to search notifications: %w", err)
	}

	// Convert to response
	var responses []*model.NotificationResponse
	for _, notification := range notifications {
		responses = append(responses, notification.ToResponse())
	}

	// Calculate pagination
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	pagination := &PaginationResponse{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	return responses, pagination, nil
}

// Notification Processing

// ProcessNotificationQueue processes queued notifications
func (s *notificationService) ProcessNotificationQueue() error {
	// Get pending queue items
	queueItems, err := s.notificationRepo.GetQueueItems(100)
	if err != nil {
		return fmt.Errorf("failed to get queue items: %w", err)
	}

	for _, queueItem := range queueItems {
		// Update status to processing
		queueItem.Status = "processing"
		queueItem.ProcessedAt = &time.Time{}
		*queueItem.ProcessedAt = time.Now()
		s.notificationRepo.UpdateQueueItem(queueItem)

		// Get notification
		notification, err := s.notificationRepo.GetNotificationByID(queueItem.NotificationID)
		if err != nil {
			logger.Errorf("Failed to get notification %d: %v", queueItem.NotificationID, err)
			continue
		}

		// Send notification
		if err := s.SendNotification(notification); err != nil {
			logger.Errorf("Failed to send notification %d: %v", notification.ID, err)
			queueItem.RetryCount++
			if queueItem.RetryCount >= queueItem.MaxRetries {
				queueItem.Status = "failed"
			} else {
				queueItem.Status = "pending"
			}
		} else {
			queueItem.Status = "completed"
		}

		// Update queue item
		s.notificationRepo.UpdateQueueItem(queueItem)
	}

	return nil
}

// SendNotification sends a notification
func (s *notificationService) SendNotification(notification *model.Notification) error {
	// Check if user has preferences for this type and channel
	if notification.UserID != nil {
		preference, err := s.notificationRepo.GetNotificationPreferenceByUserTypeChannel(*notification.UserID, notification.Type, notification.Channel)
		if err == nil && preference != nil && !preference.IsEnabled {
			// User has disabled this type of notification
			notification.Status = model.NotificationStatusCancelled
			s.notificationRepo.UpdateNotification(notification)
			return nil
		}
	}

	// Update status to sent
	now := time.Now()
	notification.Status = model.NotificationStatusSent
	notification.SentAt = &now
	s.notificationRepo.UpdateNotification(notification)

	// Create log entry
	log := &model.NotificationLog{
		NotificationID: notification.ID,
		Channel:        notification.Channel,
		Status:         model.NotificationStatusSent,
		AttemptedAt:    now,
	}

	// TODO: Implement actual sending logic based on channel
	// For now, just simulate success
	notification.Status = model.NotificationStatusDelivered
	notification.DeliveredAt = &now
	s.notificationRepo.UpdateNotification(notification)

	log.Status = model.NotificationStatusDelivered
	log.DeliveredAt = &now
	s.notificationRepo.CreateNotificationLog(log)

	return nil
}

// ScheduleNotification schedules a notification
func (s *notificationService) ScheduleNotification(notification *model.Notification, scheduledAt time.Time) error {
	// Add to queue
	queueItem := &model.NotificationQueue{
		NotificationID: notification.ID,
		Priority:       notification.Priority,
		Channel:        notification.Channel,
		ScheduledAt:    scheduledAt,
		Status:         "pending",
	}

	return s.notificationRepo.AddToQueue(queueItem)
}

// Cleanup

// DeleteExpiredNotifications deletes expired notifications
func (s *notificationService) DeleteExpiredNotifications() error {
	if err := s.notificationRepo.DeleteExpiredNotifications(); err != nil {
		return fmt.Errorf("failed to delete expired notifications: %w", err)
	}
	return nil
}

// DeleteOldNotifications deletes old notifications
func (s *notificationService) DeleteOldNotifications(olderThan time.Time) error {
	if err := s.notificationRepo.DeleteOldNotifications(olderThan); err != nil {
		return fmt.Errorf("failed to delete old notifications: %w", err)
	}
	return nil
}

// ArchiveOldNotifications archives old notifications
func (s *notificationService) ArchiveOldNotifications(olderThan time.Time) error {
	if err := s.notificationRepo.ArchiveOldNotifications(olderThan); err != nil {
		return fmt.Errorf("failed to archive old notifications: %w", err)
	}
	return nil
}
