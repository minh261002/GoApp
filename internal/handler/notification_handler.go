package handler

import (
	"go_app/internal/model"
	"go_app/internal/service"
	"go_app/pkg/response"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// NotificationHandler handles notification-related HTTP requests
type NotificationHandler struct {
	notificationService service.NotificationService
}

// NewNotificationHandler creates a new NotificationHandler
func NewNotificationHandler(notificationService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// CreateNotification creates a new notification
// @Summary Create notification
// @Description Create a new notification
// @Tags notifications
// @Accept json
// @Produce json
// @Param notification body model.CreateNotificationRequest true "Notification data"
// @Success 201 {object} response.SuccessResponse{data=model.NotificationResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications [post]
// @Security BearerAuth
func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	var req model.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	notification, err := h.notificationService.CreateNotification(&req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create notification", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Notification created successfully", notification)
}

// GetNotificationByID gets a notification by ID
// @Summary Get notification by ID
// @Description Get a notification by its ID
// @Tags notifications
// @Produce json
// @Param id path int true "Notification ID"
// @Success 200 {object} response.SuccessResponse{data=model.NotificationResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/{id} [get]
// @Security BearerAuth
func (h *NotificationHandler) GetNotificationByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid notification ID", err.Error())
		return
	}

	notification, err := h.notificationService.GetNotificationByID(uint(id))
	if err != nil {
		if err.Error() == "notification not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Notification not found", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get notification", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Notification retrieved successfully", notification)
}

// GetNotificationsByUser gets notifications for a user
// @Summary Get user notifications
// @Description Get notifications for the authenticated user
// @Tags notifications
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param type query string false "Notification type"
// @Param status query string false "Notification status"
// @Param channel query string false "Notification channel"
// @Param priority query string false "Notification priority"
// @Param is_read query bool false "Is read"
// @Param is_archived query bool false "Is archived"
// @Param start_date query string false "Start date (RFC3339)"
// @Param end_date query string false "End date (RFC3339)"
// @Param sort_by query string false "Sort by field" default(created_at)
// @Param sort_order query string false "Sort order" default(desc)
// @Success 200 {object} response.SuccessResponse{data=[]model.NotificationResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications [get]
// @Security BearerAuth
func (h *NotificationHandler) GetNotificationsByUser(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "User ID not found in context")
		return
	}

	// Parse query parameters
	req := &model.NotificationListRequest{
		Page:      1,
		Limit:     20,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			req.Limit = limit
		}
	}

	if typeStr := c.Query("type"); typeStr != "" {
		notificationType := model.NotificationType(typeStr)
		req.Type = &notificationType
	}

	if statusStr := c.Query("status"); statusStr != "" {
		status := model.NotificationStatus(statusStr)
		req.Status = &status
	}

	if channelStr := c.Query("channel"); channelStr != "" {
		channel := model.NotificationChannel(channelStr)
		req.Channel = &channel
	}

	if priorityStr := c.Query("priority"); priorityStr != "" {
		priority := model.NotificationPriority(priorityStr)
		req.Priority = &priority
	}

	if isReadStr := c.Query("is_read"); isReadStr != "" {
		if isRead, err := strconv.ParseBool(isReadStr); err == nil {
			req.IsRead = &isRead
		}
	}

	if isArchivedStr := c.Query("is_archived"); isArchivedStr != "" {
		if isArchived, err := strconv.ParseBool(isArchivedStr); err == nil {
			req.IsArchived = &isArchived
		}
	}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			req.StartDate = &startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			req.EndDate = &endDate
		}
	}

	if sortBy := c.Query("sort_by"); sortBy != "" {
		req.SortBy = sortBy
	}

	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		req.SortOrder = sortOrder
	}

	notifications, pagination, err := h.notificationService.GetNotificationsByUser(userID.(uint), req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get notifications", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Notifications retrieved successfully", notifications, pagination.Page, pagination.Limit, pagination.Total)
}

// UpdateNotification updates a notification
// @Summary Update notification
// @Description Update a notification
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path int true "Notification ID"
// @Param notification body model.UpdateNotificationRequest true "Notification update data"
// @Success 200 {object} response.SuccessResponse{data=model.NotificationResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/{id} [put]
// @Security BearerAuth
func (h *NotificationHandler) UpdateNotification(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid notification ID", err.Error())
		return
	}

	var req model.UpdateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	notification, err := h.notificationService.UpdateNotification(uint(id), &req)
	if err != nil {
		if err.Error() == "notification not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Notification not found", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update notification", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Notification updated successfully", notification)
}

// DeleteNotification deletes a notification
// @Summary Delete notification
// @Description Delete a notification
// @Tags notifications
// @Produce json
// @Param id path int true "Notification ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/{id} [delete]
// @Security BearerAuth
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid notification ID", err.Error())
		return
	}

	err = h.notificationService.DeleteNotification(uint(id))
	if err != nil {
		if err.Error() == "notification not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Notification not found", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete notification", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Notification deleted successfully", nil)
}

// MarkAsRead marks a notification as read
// @Summary Mark notification as read
// @Description Mark a notification as read
// @Tags notifications
// @Produce json
// @Param id path int true "Notification ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/{id}/read [post]
// @Security BearerAuth
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid notification ID", err.Error())
		return
	}

	err = h.notificationService.MarkAsRead(uint(id))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to mark notification as read", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Notification marked as read", nil)
}

// MarkAsUnread marks a notification as unread
// @Summary Mark notification as unread
// @Description Mark a notification as unread
// @Tags notifications
// @Produce json
// @Param id path int true "Notification ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/{id}/unread [post]
// @Security BearerAuth
func (h *NotificationHandler) MarkAsUnread(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid notification ID", err.Error())
		return
	}

	err = h.notificationService.MarkAsUnread(uint(id))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to mark notification as unread", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Notification marked as unread", nil)
}

// MarkAsArchived marks a notification as archived
// @Summary Mark notification as archived
// @Description Mark a notification as archived
// @Tags notifications
// @Produce json
// @Param id path int true "Notification ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/{id}/archive [post]
// @Security BearerAuth
func (h *NotificationHandler) MarkAsArchived(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid notification ID", err.Error())
		return
	}

	err = h.notificationService.MarkAsArchived(uint(id))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to mark notification as archived", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Notification marked as archived", nil)
}

// MarkAsUnarchived marks a notification as unarchived
// @Summary Mark notification as unarchived
// @Description Mark a notification as unarchived
// @Tags notifications
// @Produce json
// @Param id path int true "Notification ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/{id}/unarchive [post]
// @Security BearerAuth
func (h *NotificationHandler) MarkAsUnarchived(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid notification ID", err.Error())
		return
	}

	err = h.notificationService.MarkAsUnarchived(uint(id))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to mark notification as unarchived", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Notification marked as unarchived", nil)
}

// BulkMarkAsRead marks multiple notifications as read
// @Summary Bulk mark notifications as read
// @Description Mark multiple notifications as read
// @Tags notifications
// @Accept json
// @Produce json
// @Param notification_ids body []uint true "Notification IDs"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/bulk/read [post]
// @Security BearerAuth
func (h *NotificationHandler) BulkMarkAsRead(c *gin.Context) {
	var req struct {
		NotificationIDs []uint `json:"notification_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	err := h.notificationService.BulkMarkAsRead(req.NotificationIDs)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to bulk mark notifications as read", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Notifications marked as read", nil)
}

// BulkMarkAsArchived marks multiple notifications as archived
// @Summary Bulk mark notifications as archived
// @Description Mark multiple notifications as archived
// @Tags notifications
// @Accept json
// @Produce json
// @Param notification_ids body []uint true "Notification IDs"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/bulk/archive [post]
// @Security BearerAuth
func (h *NotificationHandler) BulkMarkAsArchived(c *gin.Context) {
	var req struct {
		NotificationIDs []uint `json:"notification_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	err := h.notificationService.BulkMarkAsArchived(req.NotificationIDs)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to bulk mark notifications as archived", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Notifications marked as archived", nil)
}

// GetUnreadNotificationCount gets unread notification count for a user
// @Summary Get unread notification count
// @Description Get unread notification count for the authenticated user
// @Tags notifications
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=map[string]int64}
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/unread-count [get]
// @Security BearerAuth
func (h *NotificationHandler) GetUnreadNotificationCount(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "User ID not found in context")
		return
	}

	count, err := h.notificationService.GetUnreadNotificationCount(userID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get unread notification count", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Unread notification count retrieved successfully", map[string]int64{
		"unread_count": count,
	})
}

// GetNotificationStats gets notification statistics
// @Summary Get notification statistics
// @Description Get notification statistics
// @Tags notifications
// @Produce json
// @Param start_date query string false "Start date (RFC3339)"
// @Param end_date query string false "End date (RFC3339)"
// @Param type query string false "Notification type"
// @Param channel query string false "Notification channel"
// @Param group_by query string false "Group by" Enums(day, week, month)
// @Success 200 {object} response.SuccessResponse{data=model.NotificationStatsResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/stats [get]
// @Security BearerAuth
func (h *NotificationHandler) GetNotificationStats(c *gin.Context) {
	req := &model.NotificationStatsRequest{}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			req.StartDate = &startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			req.EndDate = &endDate
		}
	}

	if typeStr := c.Query("type"); typeStr != "" {
		notificationType := model.NotificationType(typeStr)
		req.Type = &notificationType
	}

	if channelStr := c.Query("channel"); channelStr != "" {
		channel := model.NotificationChannel(channelStr)
		req.Channel = &channel
	}

	if groupBy := c.Query("group_by"); groupBy != "" {
		req.GroupBy = groupBy
	}

	stats, err := h.notificationService.GetNotificationStats(req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get notification statistics", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Notification statistics retrieved successfully", stats)
}

// SearchNotifications searches notifications
// @Summary Search notifications
// @Description Search notifications with filters
// @Tags notifications
// @Produce json
// @Param q query string false "Search query"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param user_id query int false "User ID"
// @Param type query string false "Notification type"
// @Param status query string false "Notification status"
// @Param channel query string false "Notification channel"
// @Param priority query string false "Notification priority"
// @Param is_read query bool false "Is read"
// @Param is_archived query bool false "Is archived"
// @Success 200 {object} response.SuccessResponse{data=[]model.NotificationResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/search [get]
// @Security BearerAuth
func (h *NotificationHandler) SearchNotifications(c *gin.Context) {
	query := c.Query("q")
	page := 1
	limit := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Build filters
	filters := make(map[string]interface{})

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			filters["user_id"] = uint(userID)
		}
	}

	if typeStr := c.Query("type"); typeStr != "" {
		filters["type"] = model.NotificationType(typeStr)
	}

	if statusStr := c.Query("status"); statusStr != "" {
		filters["status"] = model.NotificationStatus(statusStr)
	}

	if channelStr := c.Query("channel"); channelStr != "" {
		filters["channel"] = model.NotificationChannel(channelStr)
	}

	if priorityStr := c.Query("priority"); priorityStr != "" {
		filters["priority"] = model.NotificationPriority(priorityStr)
	}

	if isReadStr := c.Query("is_read"); isReadStr != "" {
		if isRead, err := strconv.ParseBool(isReadStr); err == nil {
			filters["is_read"] = isRead
		}
	}

	if isArchivedStr := c.Query("is_archived"); isArchivedStr != "" {
		if isArchived, err := strconv.ParseBool(isArchivedStr); err == nil {
			filters["is_archived"] = isArchived
		}
	}

	notifications, pagination, err := h.notificationService.SearchNotifications(query, filters, page, limit)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to search notifications", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Notifications search completed", notifications, pagination.Page, pagination.Limit, pagination.Total)
}

// Notification Templates

// CreateNotificationTemplate creates a new notification template
// @Summary Create notification template
// @Description Create a new notification template
// @Tags notification-templates
// @Accept json
// @Produce json
// @Param template body model.CreateNotificationRequest true "Template data"
// @Success 201 {object} response.SuccessResponse{data=model.NotificationTemplateResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notification-templates [post]
// @Security BearerAuth
func (h *NotificationHandler) CreateNotificationTemplate(c *gin.Context) {
	var req model.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	template, err := h.notificationService.CreateNotificationTemplate(&req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create notification template", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Notification template created successfully", template)
}

// GetNotificationTemplateByID gets a notification template by ID
// @Summary Get notification template by ID
// @Description Get a notification template by its ID
// @Tags notification-templates
// @Produce json
// @Param id path int true "Template ID"
// @Success 200 {object} response.SuccessResponse{data=model.NotificationTemplateResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notification-templates/{id} [get]
// @Security BearerAuth
func (h *NotificationHandler) GetNotificationTemplateByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid template ID", err.Error())
		return
	}

	template, err := h.notificationService.GetNotificationTemplateByID(uint(id))
	if err != nil {
		if err.Error() == "notification template not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Notification template not found", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get notification template", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Notification template retrieved successfully", template)
}

// GetNotificationTemplates gets notification templates
// @Summary Get notification templates
// @Description Get notification templates with pagination
// @Tags notification-templates
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param type query string false "Template type"
// @Param channel query string false "Template channel"
// @Param sort_by query string false "Sort by field" default(created_at)
// @Param sort_order query string false "Sort order" default(desc)
// @Success 200 {object} response.SuccessResponse{data=[]model.NotificationTemplateResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notification-templates [get]
// @Security BearerAuth
func (h *NotificationHandler) GetNotificationTemplates(c *gin.Context) {
	req := &model.NotificationListRequest{
		Page:      1,
		Limit:     20,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			req.Limit = limit
		}
	}

	if typeStr := c.Query("type"); typeStr != "" {
		notificationType := model.NotificationType(typeStr)
		req.Type = &notificationType
	}

	if channelStr := c.Query("channel"); channelStr != "" {
		channel := model.NotificationChannel(channelStr)
		req.Channel = &channel
	}

	if sortBy := c.Query("sort_by"); sortBy != "" {
		req.SortBy = sortBy
	}

	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		req.SortOrder = sortOrder
	}

	templates, pagination, err := h.notificationService.GetNotificationTemplates(req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get notification templates", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Notification templates retrieved successfully", templates, pagination.Page, pagination.Limit, pagination.Total)
}

// UpdateNotificationTemplate updates a notification template
// @Summary Update notification template
// @Description Update a notification template
// @Tags notification-templates
// @Accept json
// @Produce json
// @Param id path int true "Template ID"
// @Param template body model.CreateNotificationRequest true "Template update data"
// @Success 200 {object} response.SuccessResponse{data=model.NotificationTemplateResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notification-templates/{id} [put]
// @Security BearerAuth
func (h *NotificationHandler) UpdateNotificationTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid template ID", err.Error())
		return
	}

	var req model.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	template, err := h.notificationService.UpdateNotificationTemplate(uint(id), &req)
	if err != nil {
		if err.Error() == "notification template not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Notification template not found", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update notification template", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Notification template updated successfully", template)
}

// DeleteNotificationTemplate deletes a notification template
// @Summary Delete notification template
// @Description Delete a notification template
// @Tags notification-templates
// @Produce json
// @Param id path int true "Template ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notification-templates/{id} [delete]
// @Security BearerAuth
func (h *NotificationHandler) DeleteNotificationTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid template ID", err.Error())
		return
	}

	err = h.notificationService.DeleteNotificationTemplate(uint(id))
	if err != nil {
		if err.Error() == "notification template not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Notification template not found", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete notification template", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Notification template deleted successfully", nil)
}

// Notification Preferences

// CreateNotificationPreference creates a new notification preference
// @Summary Create notification preference
// @Description Create a new notification preference for the authenticated user
// @Tags notification-preferences
// @Accept json
// @Produce json
// @Param preference body model.NotificationPreferenceRequest true "Preference data"
// @Success 201 {object} response.SuccessResponse{data=model.NotificationPreferenceResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notification-preferences [post]
// @Security BearerAuth
func (h *NotificationHandler) CreateNotificationPreference(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "User ID not found in context")
		return
	}

	var req model.NotificationPreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	preference, err := h.notificationService.CreateNotificationPreference(userID.(uint), &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create notification preference", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Notification preference created successfully", preference)
}

// GetNotificationPreferencesByUser gets notification preferences for a user
// @Summary Get user notification preferences
// @Description Get notification preferences for the authenticated user
// @Tags notification-preferences
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=[]model.NotificationPreferenceResponse}
// @Failure 500 {object} response.ErrorResponse
// @Router /notification-preferences [get]
// @Security BearerAuth
func (h *NotificationHandler) GetNotificationPreferencesByUser(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "User ID not found in context")
		return
	}

	preferences, err := h.notificationService.GetNotificationPreferencesByUser(userID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get notification preferences", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Notification preferences retrieved successfully", preferences)
}

// UpdateNotificationPreference updates a notification preference
// @Summary Update notification preference
// @Description Update a notification preference
// @Tags notification-preferences
// @Accept json
// @Produce json
// @Param id path int true "Preference ID"
// @Param preference body model.NotificationPreferenceRequest true "Preference update data"
// @Success 200 {object} response.SuccessResponse{data=model.NotificationPreferenceResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notification-preferences/{id} [put]
// @Security BearerAuth
func (h *NotificationHandler) UpdateNotificationPreference(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid preference ID", err.Error())
		return
	}

	var req model.NotificationPreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	preference, err := h.notificationService.UpdateNotificationPreference(uint(id), &req)
	if err != nil {
		if err.Error() == "notification preference not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Notification preference not found", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update notification preference", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Notification preference updated successfully", preference)
}

// DeleteNotificationPreference deletes a notification preference
// @Summary Delete notification preference
// @Description Delete a notification preference
// @Tags notification-preferences
// @Produce json
// @Param id path int true "Preference ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notification-preferences/{id} [delete]
// @Security BearerAuth
func (h *NotificationHandler) DeleteNotificationPreference(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid preference ID", err.Error())
		return
	}

	err = h.notificationService.DeleteNotificationPreference(uint(id))
	if err != nil {
		if err.Error() == "notification preference not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Notification preference not found", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete notification preference", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Notification preference deleted successfully", nil)
}
