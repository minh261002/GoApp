package handler

import (
	"net/http"
	"time"

	"go_app/internal/model"
	"go_app/internal/service"
	"go_app/pkg/logger"
	"go_app/pkg/response"

	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	eventService service.EventService
}

func NewEventHandler(eventService service.EventService) *EventHandler {
	return &EventHandler{
		eventService: eventService,
	}
}

// TriggerOrderCreated triggers order created event
// @Summary Trigger order created event
// @Description Manually trigger order created event for testing
// @Tags events
// @Accept json
// @Produce json
// @Param order body model.Order true "Order data"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/events/order/created [post]
func (h *EventHandler) TriggerOrderCreated(c *gin.Context) {
	var order model.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err := h.eventService.OnOrderCreated(&order)
	if err != nil {
		logger.Errorf("Failed to trigger order created event: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to trigger event", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Order created event triggered successfully", nil)
}

// TriggerOrderStatusUpdated triggers order status updated event
// @Summary Trigger order status updated event
// @Description Manually trigger order status updated event for testing
// @Tags events
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Event data with order, oldStatus, newStatus"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/events/order/status-updated [post]
func (h *EventHandler) TriggerOrderStatusUpdated(c *gin.Context) {
	var req struct {
		Order     model.Order       `json:"order"`
		OldStatus model.OrderStatus `json:"old_status"`
		NewStatus model.OrderStatus `json:"new_status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err := h.eventService.OnOrderStatusUpdated(&req.Order, req.OldStatus, req.NewStatus)
	if err != nil {
		logger.Errorf("Failed to trigger order status updated event: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to trigger event", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Order status updated event triggered successfully", nil)
}

// TriggerOrderShipped triggers order shipped event
// @Summary Trigger order shipped event
// @Description Manually trigger order shipped event for testing
// @Tags events
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Event data with order and trackingNumber"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/events/order/shipped [post]
func (h *EventHandler) TriggerOrderShipped(c *gin.Context) {
	var req struct {
		Order          model.Order `json:"order"`
		TrackingNumber string      `json:"tracking_number"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err := h.eventService.OnOrderShipped(&req.Order, req.TrackingNumber)
	if err != nil {
		logger.Errorf("Failed to trigger order shipped event: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to trigger event", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Order shipped event triggered successfully", nil)
}

// TriggerOrderDelivered triggers order delivered event
// @Summary Trigger order delivered event
// @Description Manually trigger order delivered event for testing
// @Tags events
// @Accept json
// @Produce json
// @Param order body model.Order true "Order data"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/events/order/delivered [post]
func (h *EventHandler) TriggerOrderDelivered(c *gin.Context) {
	var order model.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err := h.eventService.OnOrderDelivered(&order)
	if err != nil {
		logger.Errorf("Failed to trigger order delivered event: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to trigger event", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Order delivered event triggered successfully", nil)
}

// TriggerOrderCancelled triggers order cancelled event
// @Summary Trigger order cancelled event
// @Description Manually trigger order cancelled event for testing
// @Tags events
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Event data with order and reason"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/events/order/cancelled [post]
func (h *EventHandler) TriggerOrderCancelled(c *gin.Context) {
	var req struct {
		Order  model.Order `json:"order"`
		Reason string      `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err := h.eventService.OnOrderCancelled(&req.Order, req.Reason)
	if err != nil {
		logger.Errorf("Failed to trigger order cancelled event: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to trigger event", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Order cancelled event triggered successfully", nil)
}

// TriggerPaymentSuccess triggers payment success event
// @Summary Trigger payment success event
// @Description Manually trigger payment success event for testing
// @Tags events
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Event data with order and payment"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/events/payment/success [post]
func (h *EventHandler) TriggerPaymentSuccess(c *gin.Context) {
	var req struct {
		Order   model.Order   `json:"order"`
		Payment model.Payment `json:"payment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err := h.eventService.OnPaymentSuccess(&req.Order, &req.Payment)
	if err != nil {
		logger.Errorf("Failed to trigger payment success event: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to trigger event", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Payment success event triggered successfully", nil)
}

// TriggerPaymentFailed triggers payment failed event
// @Summary Trigger payment failed event
// @Description Manually trigger payment failed event for testing
// @Tags events
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Event data with order, payment and error message"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/events/payment/failed [post]
func (h *EventHandler) TriggerPaymentFailed(c *gin.Context) {
	var req struct {
		Order    model.Order   `json:"order"`
		Payment  model.Payment `json:"payment"`
		ErrorMsg string        `json:"error_msg"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err := h.eventService.OnPaymentFailed(&req.Order, &req.Payment, req.ErrorMsg)
	if err != nil {
		logger.Errorf("Failed to trigger payment failed event: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to trigger event", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Payment failed event triggered successfully", nil)
}

// TriggerProductBackInStock triggers product back in stock event
// @Summary Trigger product back in stock event
// @Description Manually trigger product back in stock event for testing
// @Tags events
// @Accept json
// @Produce json
// @Param product body model.Product true "Product data"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/events/product/back-in-stock [post]
func (h *EventHandler) TriggerProductBackInStock(c *gin.Context) {
	var product model.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err := h.eventService.OnProductBackInStock(&product)
	if err != nil {
		logger.Errorf("Failed to trigger product back in stock event: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to trigger event", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Product back in stock event triggered successfully", nil)
}

// TriggerPriceDrop triggers price drop event
// @Summary Trigger price drop event
// @Description Manually trigger price drop event for testing
// @Tags events
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Event data with product, oldPrice and newPrice"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/events/product/price-drop [post]
func (h *EventHandler) TriggerPriceDrop(c *gin.Context) {
	var req struct {
		Product  model.Product `json:"product"`
		OldPrice float64       `json:"old_price"`
		NewPrice float64       `json:"new_price"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err := h.eventService.OnPriceDrop(&req.Product, req.OldPrice, req.NewPrice)
	if err != nil {
		logger.Errorf("Failed to trigger price drop event: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to trigger event", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Price drop event triggered successfully", nil)
}

// TriggerReviewCreated triggers review created event
// @Summary Trigger review created event
// @Description Manually trigger review created event for testing
// @Tags events
// @Accept json
// @Produce json
// @Param review body model.Review true "Review data"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/events/review/created [post]
func (h *EventHandler) TriggerReviewCreated(c *gin.Context) {
	var review model.Review
	if err := c.ShouldBindJSON(&review); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err := h.eventService.OnReviewCreated(&review)
	if err != nil {
		logger.Errorf("Failed to trigger review created event: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to trigger event", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Review created event triggered successfully", nil)
}

// TriggerReviewApproved triggers review approved event
// @Summary Trigger review approved event
// @Description Manually trigger review approved event for testing
// @Tags events
// @Accept json
// @Produce json
// @Param review body model.Review true "Review data"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/events/review/approved [post]
func (h *EventHandler) TriggerReviewApproved(c *gin.Context) {
	var review model.Review
	if err := c.ShouldBindJSON(&review); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err := h.eventService.OnReviewApproved(&review)
	if err != nil {
		logger.Errorf("Failed to trigger review approved event: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to trigger event", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Review approved event triggered successfully", nil)
}

// TriggerLowStockAlert triggers low stock alert event
// @Summary Trigger low stock alert event
// @Description Manually trigger low stock alert event for testing
// @Tags events
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Event data with product, currentStock and minStock"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/events/inventory/low-stock [post]
func (h *EventHandler) TriggerLowStockAlert(c *gin.Context) {
	var req struct {
		Product      model.Product `json:"product"`
		CurrentStock int           `json:"current_stock"`
		MinStock     int           `json:"min_stock"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err := h.eventService.OnLowStockAlert(&req.Product, req.CurrentStock, req.MinStock)
	if err != nil {
		logger.Errorf("Failed to trigger low stock alert event: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to trigger event", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Low stock alert event triggered successfully", nil)
}

// TriggerCouponExpiring triggers coupon expiring event
// @Summary Trigger coupon expiring event
// @Description Manually trigger coupon expiring event for testing
// @Tags events
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Event data with coupon and daysLeft"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/events/coupon/expiring [post]
func (h *EventHandler) TriggerCouponExpiring(c *gin.Context) {
	var req struct {
		Coupon   model.Coupon `json:"coupon"`
		DaysLeft int          `json:"days_left"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err := h.eventService.OnCouponExpiring(&req.Coupon, req.DaysLeft)
	if err != nil {
		logger.Errorf("Failed to trigger coupon expiring event: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to trigger event", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Coupon expiring event triggered successfully", nil)
}

// TriggerPointsEarned triggers points earned event
// @Summary Trigger points earned event
// @Description Manually trigger points earned event for testing
// @Tags events
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Event data with userID, points and source"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/events/points/earned [post]
func (h *EventHandler) TriggerPointsEarned(c *gin.Context) {
	var req struct {
		UserID uint   `json:"user_id"`
		Points int64  `json:"points"`
		Source string `json:"source"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err := h.eventService.OnPointsEarned(req.UserID, req.Points, req.Source)
	if err != nil {
		logger.Errorf("Failed to trigger points earned event: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to trigger event", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Points earned event triggered successfully", nil)
}

// TriggerPointsExpiring triggers points expiring event
// @Summary Trigger points expiring event
// @Description Manually trigger points expiring event for testing
// @Tags events
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Event data with userID, points and expiryDate"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/events/points/expiring [post]
func (h *EventHandler) TriggerPointsExpiring(c *gin.Context) {
	var req struct {
		UserID     uint      `json:"user_id"`
		Points     int64     `json:"points"`
		ExpiryDate time.Time `json:"expiry_date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err := h.eventService.OnPointsExpiring(req.UserID, req.Points, req.ExpiryDate)
	if err != nil {
		logger.Errorf("Failed to trigger points expiring event: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to trigger event", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Points expiring event triggered successfully", nil)
}
