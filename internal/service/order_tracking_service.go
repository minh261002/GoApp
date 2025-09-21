package service

import (
	"encoding/json"
	"fmt"
	"go_app/internal/model"
	"go_app/internal/repository"
	"go_app/pkg/logger"
	"time"
)

type OrderTrackingService struct {
	orderTrackingRepo *repository.OrderTrackingRepository
	orderRepo         repository.OrderRepository
	userRepo          repository.UserRepository
}

func NewOrderTrackingService() *OrderTrackingService {
	return &OrderTrackingService{
		orderTrackingRepo: repository.NewOrderTrackingRepository(),
		orderRepo:         repository.NewOrderRepository(),
		userRepo:          repository.NewUserRepository(),
	}
}

// ===== ORDER TRACKING SERVICE =====

// CreateOrderTracking creates a new order tracking
func (s *OrderTrackingService) CreateOrderTracking(req *model.OrderTrackingCreateRequest, userID uint) (*model.OrderTrackingResponse, error) {
	// Check if order exists
	_, err := s.orderRepo.GetOrderByID(req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("order not found")
	}

	// Check if tracking already exists for this order
	existing, err := s.orderTrackingRepo.GetOrderTrackingByOrderID(req.OrderID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("tracking already exists for this order")
	}

	// Set default values
	autoSync := true
	if req.AutoSync != nil {
		autoSync = *req.AutoSync
	}

	notifyUser := true
	if req.NotifyUser != nil {
		notifyUser = *req.NotifyUser
	}

	tracking := &model.OrderTracking{
		OrderID:           req.OrderID,
		TrackingNumber:    req.TrackingNumber,
		Carrier:           req.Carrier,
		CarrierCode:       req.CarrierCode,
		Status:            model.TrackingStatusPending,
		StatusText:        "Order created, awaiting pickup",
		Location:          "",
		Description:       "Order tracking created",
		EstimatedDelivery: req.EstimatedDelivery,
		TrackingURL:       req.TrackingURL,
		LastUpdatedAt:     time.Now(),
		LastSyncAt:        time.Now(),
		AutoSync:          autoSync,
		NotifyUser:        notifyUser,
		IsActive:          true,
	}

	if err := s.orderTrackingRepo.CreateOrderTracking(tracking); err != nil {
		return nil, err
	}

	// Create initial tracking event
	event := &model.OrderTrackingEvent{
		OrderTrackingID: tracking.ID,
		Status:          tracking.Status,
		StatusText:      tracking.StatusText,
		Location:        tracking.Location,
		Description:     tracking.Description,
		EventType:       model.EventTypePickup,
		EventCode:       "CREATED",
		IsImportant:     true,
		Source:          model.SourceManual,
		EventTime:       time.Now(),
	}

	if err := s.orderTrackingRepo.CreateOrderTrackingEvent(event); err != nil {
		logger.Warnf("Failed to create initial tracking event: %v", err)
	}

	return s.convertOrderTrackingToResponse(tracking), nil
}

// GetOrderTrackingByID gets order tracking by ID
func (s *OrderTrackingService) GetOrderTrackingByID(id uint) (*model.OrderTrackingResponse, error) {
	tracking, err := s.orderTrackingRepo.GetOrderTrackingByID(id)
	if err != nil {
		return nil, err
	}

	return s.convertOrderTrackingToResponse(tracking), nil
}

// GetOrderTrackingByOrderID gets order tracking by order ID
func (s *OrderTrackingService) GetOrderTrackingByOrderID(orderID uint) (*model.OrderTrackingResponse, error) {
	tracking, err := s.orderTrackingRepo.GetOrderTrackingByOrderID(orderID)
	if err != nil {
		return nil, err
	}

	return s.convertOrderTrackingToResponse(tracking), nil
}

// GetOrderTrackingByTrackingNumber gets order tracking by tracking number
func (s *OrderTrackingService) GetOrderTrackingByTrackingNumber(trackingNumber string) (*model.OrderTrackingResponse, error) {
	tracking, err := s.orderTrackingRepo.GetOrderTrackingByTrackingNumber(trackingNumber)
	if err != nil {
		return nil, err
	}

	return s.convertOrderTrackingToResponse(tracking), nil
}

// GetAllOrderTrackings gets all order trackings with pagination
func (s *OrderTrackingService) GetAllOrderTrackings(page, limit int, filters map[string]interface{}) ([]model.OrderTrackingResponse, int64, error) {
	trackings, total, err := s.orderTrackingRepo.GetAllOrderTrackings(page, limit, filters)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.OrderTrackingResponse, len(trackings))
	for i, tracking := range trackings {
		responses[i] = *s.convertOrderTrackingToResponse(&tracking)
	}

	return responses, total, nil
}

// UpdateOrderTracking updates order tracking
func (s *OrderTrackingService) UpdateOrderTracking(id uint, req *model.OrderTrackingUpdateRequest, userID uint) (*model.OrderTrackingResponse, error) {
	tracking, err := s.orderTrackingRepo.GetOrderTrackingByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Status != "" {
		tracking.Status = req.Status
	}
	if req.StatusText != "" {
		tracking.StatusText = req.StatusText
	}
	if req.Location != "" {
		tracking.Location = req.Location
	}
	if req.Description != "" {
		tracking.Description = req.Description
	}
	if req.EstimatedDelivery != nil {
		tracking.EstimatedDelivery = req.EstimatedDelivery
	}
	if req.ActualDelivery != nil {
		tracking.ActualDelivery = req.ActualDelivery
	}
	if req.TrackingURL != "" {
		tracking.TrackingURL = req.TrackingURL
	}
	if req.AutoSync != nil {
		tracking.AutoSync = *req.AutoSync
	}
	if req.NotifyUser != nil {
		tracking.NotifyUser = *req.NotifyUser
	}
	if req.IsActive != nil {
		tracking.IsActive = *req.IsActive
	}

	tracking.LastUpdatedAt = time.Now()

	if err := s.orderTrackingRepo.UpdateOrderTracking(tracking); err != nil {
		return nil, err
	}

	return s.convertOrderTrackingToResponse(tracking), nil
}

// DeleteOrderTracking deletes order tracking
func (s *OrderTrackingService) DeleteOrderTracking(id uint, userID uint) error {
	return s.orderTrackingRepo.DeleteOrderTracking(id)
}

// ===== TRACKING EVENTS SERVICE =====

// AddTrackingEvent adds a new tracking event
func (s *OrderTrackingService) AddTrackingEvent(trackingID uint, req *model.OrderTrackingEventResponse) error {
	// Check if tracking exists
	tracking, err := s.orderTrackingRepo.GetOrderTrackingByID(trackingID)
	if err != nil {
		return err
	}

	event := &model.OrderTrackingEvent{
		OrderTrackingID: trackingID,
		Status:          req.Status,
		StatusText:      req.StatusText,
		Location:        req.Location,
		Description:     req.Description,
		EventType:       req.EventType,
		EventCode:       req.EventCode,
		IsImportant:     req.IsImportant,
		Source:          model.SourceManual,
		EventTime:       req.EventTime,
	}

	if err := s.orderTrackingRepo.CreateOrderTrackingEvent(event); err != nil {
		return err
	}

	// Update tracking status
	tracking.Status = req.Status
	tracking.StatusText = req.StatusText
	tracking.Location = req.Location
	tracking.Description = req.Description
	tracking.LastUpdatedAt = time.Now()

	if err := s.orderTrackingRepo.UpdateOrderTracking(tracking); err != nil {
		return err
	}

	// Create notification if needed
	if tracking.NotifyUser {
		if err := s.createTrackingNotification(tracking, event); err != nil {
			logger.Warnf("Failed to create tracking notification: %v", err)
		}
	}

	return nil
}

// GetTrackingEvents gets events for a tracking
func (s *OrderTrackingService) GetTrackingEvents(trackingID uint, page, limit int) ([]model.OrderTrackingEventResponse, int64, error) {
	events, total, err := s.orderTrackingRepo.GetOrderTrackingEvents(trackingID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.OrderTrackingEventResponse, len(events))
	for i, event := range events {
		responses[i] = s.convertOrderTrackingEventToResponse(&event)
	}

	return responses, total, nil
}

// ===== WEBHOOK SERVICE =====

// ProcessWebhook processes incoming webhook
func (s *OrderTrackingService) ProcessWebhook(carrier, carrierCode string, req *model.OrderTrackingWebhookRequest) error {
	// Get webhook configuration
	_, err := s.orderTrackingRepo.GetOrderTrackingWebhookByCarrier(carrier, carrierCode)
	if err != nil {
		return fmt.Errorf("webhook configuration not found")
	}

	// Get tracking by tracking number
	tracking, err := s.orderTrackingRepo.GetOrderTrackingByTrackingNumber(req.TrackingNumber)
	if err != nil {
		return fmt.Errorf("tracking not found")
	}

	// Parse event time
	eventTime, err := time.Parse(time.RFC3339, req.EventTime)
	if err != nil {
		eventTime = time.Now()
	}

	// Create tracking event
	event := &model.OrderTrackingEvent{
		OrderTrackingID: tracking.ID,
		Status:          req.Status,
		StatusText:      req.StatusText,
		Location:        req.Location,
		Description:     req.Description,
		EventType:       s.mapStatusToEventType(req.Status),
		EventCode:       req.Status,
		IsImportant:     s.isImportantEvent(req.Status),
		Source:          model.SourceWebhook,
		SourceData:      s.serializeWebhookData(req),
		EventTime:       eventTime,
	}

	if err := s.orderTrackingRepo.CreateOrderTrackingEvent(event); err != nil {
		return err
	}

	// Update tracking status
	tracking.Status = req.Status
	tracking.StatusText = req.StatusText
	tracking.Location = req.Location
	tracking.Description = req.Description
	tracking.LastUpdatedAt = time.Now()
	tracking.LastSyncAt = time.Now()

	if err := s.orderTrackingRepo.UpdateOrderTracking(tracking); err != nil {
		return err
	}

	// Create notification if needed
	if tracking.NotifyUser {
		if err := s.createTrackingNotification(tracking, event); err != nil {
			logger.Warnf("Failed to create tracking notification: %v", err)
		}
	}

	return nil
}

// ===== SYNC SERVICE =====

// SyncOrderTrackings syncs order trackings with external providers
func (s *OrderTrackingService) SyncOrderTrackings(limit int) error {
	trackings, err := s.orderTrackingRepo.GetOrderTrackingsForSync(limit)
	if err != nil {
		return err
	}

	for _, tracking := range trackings {
		if err := s.syncTrackingWithProvider(&tracking); err != nil {
			logger.Errorf("Failed to sync tracking %d: %v", tracking.ID, err)
			continue
		}
	}

	return nil
}

// ===== STATISTICS SERVICE =====

// GetOrderTrackingStats gets tracking statistics
func (s *OrderTrackingService) GetOrderTrackingStats() (*model.OrderTrackingStatsResponse, error) {
	return s.orderTrackingRepo.GetOrderTrackingStats()
}

// ===== HELPER METHODS =====

// convertOrderTrackingToResponse converts OrderTracking to OrderTrackingResponse
func (s *OrderTrackingService) convertOrderTrackingToResponse(tracking *model.OrderTracking) *model.OrderTrackingResponse {
	response := &model.OrderTrackingResponse{
		ID:                tracking.ID,
		OrderID:           tracking.OrderID,
		TrackingNumber:    tracking.TrackingNumber,
		Carrier:           tracking.Carrier,
		CarrierCode:       tracking.CarrierCode,
		Status:            tracking.Status,
		StatusText:        tracking.StatusText,
		Location:          tracking.Location,
		Description:       tracking.Description,
		EstimatedDelivery: tracking.EstimatedDelivery,
		ActualDelivery:    tracking.ActualDelivery,
		TrackingURL:       tracking.TrackingURL,
		LastUpdatedAt:     tracking.LastUpdatedAt,
		LastSyncAt:        tracking.LastSyncAt,
		AutoSync:          tracking.AutoSync,
		NotifyUser:        tracking.NotifyUser,
		IsActive:          tracking.IsActive,
		CreatedAt:         tracking.CreatedAt,
		UpdatedAt:         tracking.UpdatedAt,
	}

	// Get events for this tracking
	events, _, err := s.orderTrackingRepo.GetOrderTrackingEvents(tracking.ID, 1, 10)
	if err == nil {
		eventResponses := make([]model.OrderTrackingEventResponse, len(events))
		for i, event := range events {
			eventResponses[i] = s.convertOrderTrackingEventToResponse(&event)
		}
		response.Events = eventResponses
	}

	return response
}

// convertOrderTrackingEventToResponse converts OrderTrackingEvent to OrderTrackingEventResponse
func (s *OrderTrackingService) convertOrderTrackingEventToResponse(event *model.OrderTrackingEvent) model.OrderTrackingEventResponse {
	return model.OrderTrackingEventResponse{
		ID:              event.ID,
		OrderTrackingID: event.OrderTrackingID,
		Status:          event.Status,
		StatusText:      event.StatusText,
		Location:        event.Location,
		Description:     event.Description,
		EventType:       event.EventType,
		EventCode:       event.EventCode,
		IsImportant:     event.IsImportant,
		Source:          event.Source,
		EventTime:       event.EventTime,
		CreatedAt:       event.CreatedAt,
	}
}

// mapStatusToEventType maps status to event type
func (s *OrderTrackingService) mapStatusToEventType(status string) string {
	switch status {
	case model.TrackingStatusPickedUp:
		return model.EventTypePickup
	case model.TrackingStatusInTransit:
		return model.EventTypeTransit
	case model.TrackingStatusOutForDelivery:
		return model.EventTypeOutForDelivery
	case model.TrackingStatusDelivered:
		return model.EventTypeDelivered
	case model.TrackingStatusFailed:
		return model.EventTypeFailed
	case model.TrackingStatusReturned:
		return model.EventTypeReturned
	case model.TrackingStatusCancelled:
		return model.EventTypeCancelled
	default:
		return model.EventTypeTransit
	}
}

// isImportantEvent checks if event is important
func (s *OrderTrackingService) isImportantEvent(status string) bool {
	importantStatuses := []string{
		model.TrackingStatusPickedUp,
		model.TrackingStatusOutForDelivery,
		model.TrackingStatusDelivered,
		model.TrackingStatusFailed,
		model.TrackingStatusReturned,
	}

	for _, importantStatus := range importantStatuses {
		if status == importantStatus {
			return true
		}
	}
	return false
}

// serializeWebhookData serializes webhook request data
func (s *OrderTrackingService) serializeWebhookData(req *model.OrderTrackingWebhookRequest) string {
	data, err := json.Marshal(req)
	if err != nil {
		return ""
	}
	return string(data)
}

// createTrackingNotification creates a notification for tracking event
func (s *OrderTrackingService) createTrackingNotification(tracking *model.OrderTracking, event *model.OrderTrackingEvent) error {
	// Get order to get user ID
	order, err := s.orderRepo.GetOrderByID(tracking.OrderID)
	if err != nil {
		return err
	}

	notification := &model.OrderTrackingNotification{
		OrderTrackingID: tracking.ID,
		UserID:          order.UserID,
		EventID:         event.ID,
		Type:            model.NotificationTypeEmail,
		Title:           fmt.Sprintf("Order Update: %s", event.StatusText),
		Message:         fmt.Sprintf("Your order #%s status has been updated to: %s", order.OrderNumber, event.StatusText),
		IsSent:          false,
		RetryCount:      0,
		MaxRetries:      3,
	}

	return s.orderTrackingRepo.CreateOrderTrackingNotification(notification)
}

// syncTrackingWithProvider syncs tracking with external provider
func (s *OrderTrackingService) syncTrackingWithProvider(tracking *model.OrderTracking) error {
	// This would integrate with external shipping providers
	// For now, just update the last sync time
	tracking.LastSyncAt = time.Now()
	return s.orderTrackingRepo.UpdateOrderTracking(tracking)
}
