package handler

import (
	"net/http"
	"strconv"

	"go_app/internal/model"
	"go_app/internal/service"
	"go_app/pkg/logger"
	"go_app/pkg/response"

	"github.com/gin-gonic/gin"
)

type ShippingHandler struct {
	shippingService service.ShippingService
}

func NewShippingHandler(shippingService service.ShippingService) *ShippingHandler {
	return &ShippingHandler{
		shippingService: shippingService,
	}
}

// Shipping Providers
// CreateShippingProvider creates a new shipping provider
// @Summary Create shipping provider
// @Description Create a new shipping provider
// @Tags shipping-providers
// @Accept json
// @Produce json
// @Param provider body model.ShippingProvider true "Shipping provider data"
// @Success 201 {object} response.Response{data=model.ShippingProvider}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/shipping/providers [post]
func (h *ShippingHandler) CreateShippingProvider(c *gin.Context) {
	var req model.ShippingProvider
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	provider, err := h.shippingService.CreateShippingProvider(&req)
	if err != nil {
		logger.Errorf("Failed to create shipping provider: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create shipping provider", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Shipping provider created successfully", provider)
}

// GetShippingProviderByID gets a shipping provider by ID
// @Summary Get shipping provider by ID
// @Description Get a shipping provider by its ID
// @Tags shipping-providers
// @Produce json
// @Param id path int true "Provider ID"
// @Success 200 {object} response.Response{data=model.ShippingProvider}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/shipping/providers/{id} [get]
func (h *ShippingHandler) GetShippingProviderByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid provider ID", err.Error())
		return
	}

	provider, err := h.shippingService.GetShippingProviderByID(uint(id))
	if err != nil {
		if err.Error() == "shipping provider not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Shipping provider not found", nil)
			return
		}
		logger.Errorf("Failed to get shipping provider: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get shipping provider", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Shipping provider retrieved successfully", provider)
}

// GetAllShippingProviders gets all shipping providers
// @Summary Get all shipping providers
// @Description Get all shipping providers
// @Tags shipping-providers
// @Produce json
// @Success 200 {object} response.Response{data=[]model.ShippingProvider}
// @Failure 500 {object} response.Response
// @Router /api/v1/shipping/providers [get]
func (h *ShippingHandler) GetAllShippingProviders(c *gin.Context) {
	providers, err := h.shippingService.GetAllShippingProviders()
	if err != nil {
		logger.Errorf("Failed to get all shipping providers: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get shipping providers", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Shipping providers retrieved successfully", providers)
}

// GetActiveShippingProviders gets active shipping providers
// @Summary Get active shipping providers
// @Description Get active shipping providers
// @Tags shipping-providers
// @Produce json
// @Success 200 {object} response.Response{data=[]model.ShippingProvider}
// @Failure 500 {object} response.Response
// @Router /api/v1/shipping/providers/active [get]
func (h *ShippingHandler) GetActiveShippingProviders(c *gin.Context) {
	providers, err := h.shippingService.GetActiveShippingProviders()
	if err != nil {
		logger.Errorf("Failed to get active shipping providers: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get active shipping providers", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Active shipping providers retrieved successfully", providers)
}

// UpdateShippingProvider updates a shipping provider
// @Summary Update shipping provider
// @Description Update an existing shipping provider
// @Tags shipping-providers
// @Accept json
// @Produce json
// @Param id path int true "Provider ID"
// @Param provider body model.ShippingProvider true "Shipping provider data"
// @Success 200 {object} response.Response{data=model.ShippingProvider}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/shipping/providers/{id} [put]
func (h *ShippingHandler) UpdateShippingProvider(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid provider ID", err.Error())
		return
	}

	var req model.ShippingProvider
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	provider, err := h.shippingService.UpdateShippingProvider(uint(id), &req)
	if err != nil {
		if err.Error() == "shipping provider not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Shipping provider not found", nil)
			return
		}
		logger.Errorf("Failed to update shipping provider: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update shipping provider", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Shipping provider updated successfully", provider)
}

// DeleteShippingProvider deletes a shipping provider
// @Summary Delete shipping provider
// @Description Delete a shipping provider by ID
// @Tags shipping-providers
// @Produce json
// @Param id path int true "Provider ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/shipping/providers/{id} [delete]
func (h *ShippingHandler) DeleteShippingProvider(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid provider ID", err.Error())
		return
	}

	err = h.shippingService.DeleteShippingProvider(uint(id))
	if err != nil {
		if err.Error() == "shipping provider not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Shipping provider not found", nil)
			return
		}
		logger.Errorf("Failed to delete shipping provider: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete shipping provider", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Shipping provider deleted successfully", nil)
}

// Shipping Calculation
// CalculateShipping calculates shipping fees
// @Summary Calculate shipping fees
// @Description Calculate shipping fees for given parameters
// @Tags shipping
// @Accept json
// @Produce json
// @Param request body model.CalculateShippingRequest true "Shipping calculation request"
// @Success 200 {object} response.Response{data=[]model.CalculateShippingResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/shipping/calculate [post]
func (h *ShippingHandler) CalculateShipping(c *gin.Context) {
	var req model.CalculateShippingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	responses, err := h.shippingService.CalculateShipping(&req)
	if err != nil {
		logger.Errorf("Failed to calculate shipping: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to calculate shipping", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Shipping calculation completed", responses)
}

// CalculateShippingWithGHTK calculates shipping fees using GHTK
// @Summary Calculate shipping fees with GHTK
// @Description Calculate shipping fees using GHTK API
// @Tags shipping
// @Accept json
// @Produce json
// @Param request body model.CalculateShippingRequest true "Shipping calculation request"
// @Success 200 {object} response.Response{data=model.CalculateShippingResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/shipping/calculate/ghtk [post]
func (h *ShippingHandler) CalculateShippingWithGHTK(c *gin.Context) {
	var req model.CalculateShippingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	calcResp, err := h.shippingService.CalculateShippingWithGHTK(&req)
	if err != nil {
		logger.Errorf("Failed to calculate GHTK shipping: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to calculate GHTK shipping", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "GHTK shipping calculation completed", calcResp)
}

// Shipping Orders
// CreateShippingOrder creates a new shipping order
// @Summary Create shipping order
// @Description Create a new shipping order
// @Tags shipping-orders
// @Accept json
// @Produce json
// @Param order body model.ShippingOrderRequest true "Shipping order data"
// @Success 201 {object} response.Response{data=model.ShippingOrderResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/shipping/orders [post]
func (h *ShippingHandler) CreateShippingOrder(c *gin.Context) {
	var req model.ShippingOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	order, err := h.shippingService.CreateShippingOrder(&req)
	if err != nil {
		logger.Errorf("Failed to create shipping order: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create shipping order", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Shipping order created successfully", order)
}

// GetShippingOrderByID gets a shipping order by ID
// @Summary Get shipping order by ID
// @Description Get a shipping order by its ID
// @Tags shipping-orders
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} response.Response{data=model.ShippingOrderResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/shipping/orders/{id} [get]
func (h *ShippingHandler) GetShippingOrderByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID", err.Error())
		return
	}

	order, err := h.shippingService.GetShippingOrderByID(uint(id))
	if err != nil {
		if err.Error() == "shipping order not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Shipping order not found", nil)
			return
		}
		logger.Errorf("Failed to get shipping order: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get shipping order", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Shipping order retrieved successfully", order)
}

// GetShippingOrderByOrderID gets a shipping order by order ID
// @Summary Get shipping order by order ID
// @Description Get a shipping order by order ID
// @Tags shipping-orders
// @Produce json
// @Param order_id path int true "Order ID"
// @Success 200 {object} response.Response{data=model.ShippingOrderResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/shipping/orders/order/{order_id} [get]
func (h *ShippingHandler) GetShippingOrderByOrderID(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID", err.Error())
		return
	}

	order, err := h.shippingService.GetShippingOrderByOrderID(uint(orderID))
	if err != nil {
		if err.Error() == "shipping order not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Shipping order not found", nil)
			return
		}
		logger.Errorf("Failed to get shipping order: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get shipping order", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Shipping order retrieved successfully", order)
}

// GetShippingOrderByTrackingCode gets a shipping order by tracking code
// @Summary Get shipping order by tracking code
// @Description Get a shipping order by tracking code
// @Tags shipping-orders
// @Produce json
// @Param tracking_code path string true "Tracking Code"
// @Success 200 {object} response.Response{data=model.ShippingOrderResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/shipping/orders/tracking/{tracking_code} [get]
func (h *ShippingHandler) GetShippingOrderByTrackingCode(c *gin.Context) {
	trackingCode := c.Param("tracking_code")
	if trackingCode == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Tracking code is required", nil)
		return
	}

	order, err := h.shippingService.GetShippingOrderByTrackingCode(trackingCode)
	if err != nil {
		if err.Error() == "shipping order not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Shipping order not found", nil)
			return
		}
		logger.Errorf("Failed to get shipping order: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get shipping order", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Shipping order retrieved successfully", order)
}

// CancelShippingOrder cancels a shipping order
// @Summary Cancel shipping order
// @Description Cancel a shipping order
// @Tags shipping-orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param reason body map[string]string true "Cancellation reason"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/shipping/orders/{id}/cancel [post]
func (h *ShippingHandler) CancelShippingOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID", err.Error())
		return
	}

	var req struct {
		Reason string `json:"reason" validate:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err = h.shippingService.CancelShippingOrder(uint(id), req.Reason)
	if err != nil {
		if err.Error() == "shipping order not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Shipping order not found", nil)
			return
		}
		logger.Errorf("Failed to cancel shipping order: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to cancel shipping order", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Shipping order cancelled successfully", nil)
}

// GetShippingOrders gets shipping orders with pagination
// @Summary Get shipping orders
// @Description Get shipping orders with pagination and filters
// @Tags shipping-orders
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param provider_id query int false "Filter by provider ID"
// @Param status query string false "Filter by status"
// @Param order_id query int false "Filter by order ID"
// @Success 200 {object} response.Response{data=[]model.ShippingOrderResponse}
// @Failure 500 {object} response.Response
// @Router /api/v1/shipping/orders [get]
func (h *ShippingHandler) GetShippingOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	filters := make(map[string]interface{})
	if providerID := c.Query("provider_id"); providerID != "" {
		if id, err := strconv.ParseUint(providerID, 10, 32); err == nil {
			filters["provider_id"] = id
		}
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if orderID := c.Query("order_id"); orderID != "" {
		if id, err := strconv.ParseUint(orderID, 10, 32); err == nil {
			filters["order_id"] = id
		}
	}

	orders, total, err := h.shippingService.GetShippingOrders(page, limit, filters)
	if err != nil {
		logger.Errorf("Failed to get shipping orders: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get shipping orders", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Shipping orders retrieved successfully", gin.H{
		"orders": orders,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

// Tracking
// GetShippingTracking gets shipping tracking information
// @Summary Get shipping tracking
// @Description Get shipping tracking information for an order
// @Tags shipping-tracking
// @Produce json
// @Param order_id path int true "Order ID"
// @Success 200 {object} response.Response{data=[]model.ShippingTracking}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/shipping/tracking/{order_id} [get]
func (h *ShippingHandler) GetShippingTracking(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID", err.Error())
		return
	}

	tracking, err := h.shippingService.GetShippingTracking(uint(orderID))
	if err != nil {
		logger.Errorf("Failed to get shipping tracking: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get shipping tracking", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Shipping tracking retrieved successfully", tracking)
}

// Webhook
// HandleShippingWebhook handles shipping webhooks
// @Summary Handle shipping webhook
// @Description Handle webhook notifications from shipping providers
// @Tags shipping-webhooks
// @Accept json
// @Produce json
// @Param provider path string true "Provider name"
// @Param webhook body model.WebhookData true "Webhook data"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/shipping/webhook/{provider} [post]
func (h *ShippingHandler) HandleShippingWebhook(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Provider is required", nil)
		return
	}

	var webhookData model.WebhookData
	if err := c.ShouldBindJSON(&webhookData); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid webhook data", err.Error())
		return
	}

	err := h.shippingService.UpdateShippingStatusFromWebhook(&webhookData)
	if err != nil {
		logger.Errorf("Failed to handle shipping webhook: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to handle webhook", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Webhook processed successfully", nil)
}

// Statistics
// GetShippingStats gets shipping statistics
// @Summary Get shipping stats
// @Description Get overall shipping statistics
// @Tags shipping-stats
// @Produce json
// @Success 200 {object} response.Response{data=model.ShippingStats}
// @Failure 500 {object} response.Response
// @Router /api/v1/shipping/stats [get]
func (h *ShippingHandler) GetShippingStats(c *gin.Context) {
	stats, err := h.shippingService.GetShippingStats()
	if err != nil {
		logger.Errorf("Failed to get shipping stats: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get shipping stats", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Shipping stats retrieved successfully", stats)
}

// GetShippingStatsByProvider gets shipping statistics b pyrovider
// @Summary Get shipping stats by provider
// @Description Get shipping statistics for a specific provider
// @Tags shipping-stats
// @Produce json
// @Param provider_id path int true "Provider ID"
// @Success 200 {object} response.Response{data=model.ShippingStats}
// @Failure 500 {object} response.Response
// @Router /api/v1/shipping/stats/provider/{provider_id} [get]
func (h *ShippingHandler) GetShippingStatsByProvider(c *gin.Context) {
	providerIDStr := c.Param("provider_id")
	providerID, err := strconv.ParseUint(providerIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid provider ID", err.Error())
		return
	}

	stats, err := h.shippingService.GetShippingStatsByProvider(uint(providerID))
	if err != nil {
		logger.Errorf("Failed to get shipping stats by provider: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get shipping stats", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Shipping stats retrieved successfully", stats)
}
