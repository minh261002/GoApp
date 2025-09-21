package handler

import (
	"net/http"
	"strconv"

	"go_app/internal/model"
	"go_app/internal/service"
	"go_app/pkg/response"
	"go_app/pkg/validator"

	"github.com/gin-gonic/gin"
)

type OrderTrackingHandler struct {
	orderTrackingService *service.OrderTrackingService
}

func NewOrderTrackingHandler() *OrderTrackingHandler {
	return &OrderTrackingHandler{
		orderTrackingService: service.NewOrderTrackingService(),
	}
}

// ===== ORDER TRACKING ENDPOINTS =====

// CreateOrderTracking creates a new order tracking
// @Summary Create order tracking
// @Description Create tracking for an order
// @Tags order-tracking
// @Accept json
// @Produce json
// @Param tracking body model.OrderTrackingCreateRequest true "Tracking information"
// @Success 201 {object} response.SuccessResponse{data=model.OrderTrackingResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/order-tracking [post]
func (h *OrderTrackingHandler) CreateOrderTracking(c *gin.Context) {
	var req model.OrderTrackingCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if err := validator.ValidateStruct(req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", "User ID not found in context")
		return
	}

	tracking, err := h.orderTrackingService.CreateOrderTracking(&req, userID.(uint))
	if err != nil {
		if err.Error() == "order not found" {
			response.Error(c, http.StatusNotFound, "Order not found", err.Error())
			return
		}
		if err.Error() == "tracking already exists for this order" {
			response.Error(c, http.StatusConflict, "Tracking already exists", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to create order tracking", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Order tracking created successfully", tracking)
}

// GetOrderTrackingByID gets order tracking by ID
// @Summary Get order tracking by ID
// @Description Get order tracking information by ID
// @Tags order-tracking
// @Accept json
// @Produce json
// @Param id path int true "Tracking ID"
// @Success 200 {object} response.SuccessResponse{data=model.OrderTrackingResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/order-tracking/{id} [get]
func (h *OrderTrackingHandler) GetOrderTrackingByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid tracking ID", "Tracking ID must be a valid number")
		return
	}

	tracking, err := h.orderTrackingService.GetOrderTrackingByID(uint(id))
	if err != nil {
		if err.Error() == "order tracking not found" {
			response.Error(c, http.StatusNotFound, "Order tracking not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get order tracking", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Order tracking retrieved successfully", tracking)
}

// GetOrderTrackingByOrderID gets order tracking by order ID
// @Summary Get order tracking by order ID
// @Description Get order tracking information by order ID
// @Tags order-tracking
// @Accept json
// @Produce json
// @Param order_id path int true "Order ID"
// @Success 200 {object} response.SuccessResponse{data=model.OrderTrackingResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/orders/{order_id}/tracking [get]
func (h *OrderTrackingHandler) GetOrderTrackingByOrderID(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid order ID", "Order ID must be a valid number")
		return
	}

	tracking, err := h.orderTrackingService.GetOrderTrackingByOrderID(uint(orderID))
	if err != nil {
		if err.Error() == "order tracking not found" {
			response.Error(c, http.StatusNotFound, "Order tracking not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get order tracking", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Order tracking retrieved successfully", tracking)
}

// GetOrderTrackingByTrackingNumber gets order tracking by tracking number
// @Summary Get order tracking by tracking number
// @Description Get order tracking information by tracking number
// @Tags order-tracking
// @Accept json
// @Produce json
// @Param tracking_number path string true "Tracking Number"
// @Success 200 {object} response.SuccessResponse{data=model.OrderTrackingResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/order-tracking/track/{tracking_number} [get]
func (h *OrderTrackingHandler) GetOrderTrackingByTrackingNumber(c *gin.Context) {
	trackingNumber := c.Param("tracking_number")
	if trackingNumber == "" {
		response.Error(c, http.StatusBadRequest, "Invalid tracking number", "Tracking number is required")
		return
	}

	tracking, err := h.orderTrackingService.GetOrderTrackingByTrackingNumber(trackingNumber)
	if err != nil {
		if err.Error() == "order tracking not found" {
			response.Error(c, http.StatusNotFound, "Order tracking not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get order tracking", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Order tracking retrieved successfully", tracking)
}

// GetAllOrderTrackings gets all order trackings with pagination
// @Summary Get all order trackings
// @Description Get all order trackings with pagination and filters
// @Tags order-tracking
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Filter by status"
// @Param carrier query string false "Filter by carrier"
// @Param is_active query bool false "Filter by active status"
// @Success 200 {object} response.SuccessResponse{data=[]model.OrderTrackingResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/order-tracking [get]
func (h *OrderTrackingHandler) GetAllOrderTrackings(c *gin.Context) {
	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Get filters
	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if carrier := c.Query("carrier"); carrier != "" {
		filters["carrier"] = carrier
	}
	if isActive := c.Query("is_active"); isActive != "" {
		if isActive == "true" {
			filters["is_active"] = true
		} else if isActive == "false" {
			filters["is_active"] = false
		}
	}

	trackings, total, err := h.orderTrackingService.GetAllOrderTrackings(page, limit, filters)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get order trackings", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Order trackings retrieved successfully", gin.H{
		"trackings": trackings,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

// UpdateOrderTracking updates order tracking
// @Summary Update order tracking
// @Description Update order tracking information
// @Tags order-tracking
// @Accept json
// @Produce json
// @Param id path int true "Tracking ID"
// @Param tracking body model.OrderTrackingUpdateRequest true "Tracking update information"
// @Success 200 {object} response.SuccessResponse{data=model.OrderTrackingResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/order-tracking/{id} [put]
func (h *OrderTrackingHandler) UpdateOrderTracking(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid tracking ID", "Tracking ID must be a valid number")
		return
	}

	var req model.OrderTrackingUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if err := validator.ValidateStruct(req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", "User ID not found in context")
		return
	}

	tracking, err := h.orderTrackingService.UpdateOrderTracking(uint(id), &req, userID.(uint))
	if err != nil {
		if err.Error() == "order tracking not found" {
			response.Error(c, http.StatusNotFound, "Order tracking not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update order tracking", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Order tracking updated successfully", tracking)
}

// DeleteOrderTracking deletes order tracking
// @Summary Delete order tracking
// @Description Delete order tracking
// @Tags order-tracking
// @Accept json
// @Produce json
// @Param id path int true "Tracking ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/order-tracking/{id} [delete]
func (h *OrderTrackingHandler) DeleteOrderTracking(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid tracking ID", "Tracking ID must be a valid number")
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", "User ID not found in context")
		return
	}

	err = h.orderTrackingService.DeleteOrderTracking(uint(id), userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete order tracking", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Order tracking deleted successfully", nil)
}

// ===== TRACKING EVENTS ENDPOINTS =====

// GetTrackingEvents gets events for a tracking
// @Summary Get tracking events
// @Description Get tracking events for a specific tracking
// @Tags order-tracking
// @Accept json
// @Produce json
// @Param id path int true "Tracking ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.SuccessResponse{data=[]model.OrderTrackingEventResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/order-tracking/{id}/events [get]
func (h *OrderTrackingHandler) GetTrackingEvents(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid tracking ID", "Tracking ID must be a valid number")
		return
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	events, total, err := h.orderTrackingService.GetTrackingEvents(uint(id), page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get tracking events", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Tracking events retrieved successfully", gin.H{
		"events": events,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

// ===== WEBHOOK ENDPOINTS =====

// ProcessWebhook processes incoming webhook
// @Summary Process tracking webhook
// @Description Process incoming webhook from shipping provider
// @Tags order-tracking
// @Accept json
// @Produce json
// @Param carrier path string true "Carrier Code"
// @Param carrier_code path string true "Carrier Code"
// @Param webhook body model.OrderTrackingWebhookRequest true "Webhook data"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/order-tracking/webhook/{carrier}/{carrier_code} [post]
func (h *OrderTrackingHandler) ProcessWebhook(c *gin.Context) {
	carrier := c.Param("carrier")
	carrierCode := c.Param("carrier_code")

	if carrier == "" || carrierCode == "" {
		response.Error(c, http.StatusBadRequest, "Invalid carrier parameters", "Carrier and carrier code are required")
		return
	}

	var req model.OrderTrackingWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if err := validator.ValidateStruct(req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	err := h.orderTrackingService.ProcessWebhook(carrier, carrierCode, &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to process webhook", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Webhook processed successfully", nil)
}

// ===== SYNC ENDPOINTS =====

// SyncOrderTrackings syncs order trackings
// @Summary Sync order trackings
// @Description Sync order trackings with external providers
// @Tags order-tracking
// @Accept json
// @Produce json
// @Param limit query int false "Maximum number of trackings to sync" default(50)
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/order-tracking/sync [post]
func (h *OrderTrackingHandler) SyncOrderTrackings(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 100 {
		limit = 50
	}

	err := h.orderTrackingService.SyncOrderTrackings(limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to sync order trackings", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Order trackings synced successfully", nil)
}

// ===== STATISTICS ENDPOINTS =====

// GetOrderTrackingStats gets tracking statistics
// @Summary Get tracking statistics
// @Description Get order tracking statistics
// @Tags order-tracking
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=model.OrderTrackingStatsResponse}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/order-tracking/stats [get]
func (h *OrderTrackingHandler) GetOrderTrackingStats(c *gin.Context) {
	stats, err := h.orderTrackingService.GetOrderTrackingStats()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get tracking statistics", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Tracking statistics retrieved successfully", stats)
}
