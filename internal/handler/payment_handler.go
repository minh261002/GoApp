package handler

import (
	"net/http"
	"strconv"

	"go_app/internal/model"
	"go_app/internal/service"
	"go_app/pkg/response"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	orderService          service.OrderService
	paymentGatewayService service.PaymentGatewayService
}

func NewPaymentHandler(orderService service.OrderService, paymentGatewayService service.PaymentGatewayService) *PaymentHandler {
	return &PaymentHandler{
		orderService:          orderService,
		paymentGatewayService: paymentGatewayService,
	}
}

// CreatePaymentLink creates a payment link for an order
// @Summary Create payment link
// @Description Create a payment link for an order using specified payment method
// @Tags payments
// @Accept json
// @Produce json
// @Param order_id path int true "Order ID"
// @Param payment_method query string true "Payment method" Enums(vietqr, cod)
// @Success 200 {object} response.Response{data=model.PaymentLinkResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/orders/{order_id}/payment/link [post]
func (h *PaymentHandler) CreatePaymentLink(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID", err.Error())
		return
	}

	paymentMethodStr := c.Query("payment_method")
	if paymentMethodStr == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Payment method is required", nil)
		return
	}

	paymentMethod := model.PaymentMethod(paymentMethodStr)
	if !isValidPaymentMethod(paymentMethod) {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid payment method", nil)
		return
	}

	// Get order details
	order, err := h.orderService.GetOrderByID(uint(orderID))
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "Order not found", err.Error())
		return
	}

	// Convert OrderResponse to Order model (simplified)
	orderModel := &model.Order{
		ID:          order.ID,
		OrderNumber: order.OrderNumber,
		TotalAmount: order.TotalAmount,
		OrderItems:  convertOrderItemsToModel(order.OrderItems),
	}

	// Create payment link
	paymentLink, err := h.paymentGatewayService.CreatePaymentLink(orderModel, paymentMethod)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create payment link", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Payment link created successfully", paymentLink)
}

// ProcessPayment processes a payment
// @Summary Process payment
// @Description Process a payment using order code and payment method
// @Tags payments
// @Accept json
// @Produce json
// @Param order_code path int true "Order Code"
// @Param payment_method query string true "Payment method" Enums(vietqr, cod)
// @Success 200 {object} response.Response{data=model.PaymentInfoResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/payments/process/{order_code} [get]
func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	orderCode, err := strconv.Atoi(c.Param("order_code"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid order code", err.Error())
		return
	}

	paymentMethodStr := c.Query("payment_method")
	if paymentMethodStr == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Payment method is required", nil)
		return
	}

	paymentMethod := model.PaymentMethod(paymentMethodStr)
	if !isValidPaymentMethod(paymentMethod) {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid payment method", nil)
		return
	}

	// Process payment
	paymentInfo, err := h.paymentGatewayService.ProcessPayment(orderCode, paymentMethod)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to process payment", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Payment processed successfully", paymentInfo)
}

// CancelPayment cancels a payment
// @Summary Cancel payment
// @Description Cancel a payment using order code and payment method
// @Tags payments
// @Accept json
// @Produce json
// @Param order_code path int true "Order Code"
// @Param payment_method query string true "Payment method" Enums(vietqr, cod)
// @Param reason body map[string]string true "Cancellation reason"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/payments/cancel/{order_code} [post]
func (h *PaymentHandler) CancelPayment(c *gin.Context) {
	orderCode, err := strconv.Atoi(c.Param("order_code"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid order code", err.Error())
		return
	}

	paymentMethodStr := c.Query("payment_method")
	if paymentMethodStr == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Payment method is required", nil)
		return
	}

	paymentMethod := model.PaymentMethod(paymentMethodStr)
	if !isValidPaymentMethod(paymentMethod) {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid payment method", nil)
		return
	}

	var req struct {
		Reason string `json:"reason" validate:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Cancel payment
	err = h.paymentGatewayService.CancelPayment(orderCode, paymentMethod, req.Reason)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to cancel payment", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Payment cancelled successfully", nil)
}

// HandleWebhook handles payment webhooks
// @Summary Handle payment webhook
// @Description Handle webhook notifications from payment gateways
// @Tags payments
// @Accept json
// @Produce json
// @Param payment_method path string true "Payment method" Enums(vietqr, cod)
// @Param webhook body map[string]interface{} true "Webhook data"
// @Success 200 {object} response.Response{data=model.PaymentWebhookResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/payments/webhook/{payment_method} [post]
func (h *PaymentHandler) HandleWebhook(c *gin.Context) {
	paymentMethodStr := c.Param("payment_method")
	paymentMethod := model.PaymentMethod(paymentMethodStr)
	if !isValidPaymentMethod(paymentMethod) {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid payment method", nil)
		return
	}

	// Read webhook data
	webhookData, err := c.GetRawData()
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Failed to read webhook data", err.Error())
		return
	}

	// Get signature from header
	signature := c.GetHeader("X-PayOS-Signature")
	if signature == "" {
		signature = c.GetHeader("X-Signature")
	}

	// Verify webhook signature
	if !h.paymentGatewayService.VerifyWebhook(paymentMethod, signature, webhookData) {
		response.ErrorResponse(c, http.StatusUnauthorized, "Invalid webhook signature", nil)
		return
	}

	// Handle webhook
	webhookResponse, err := h.paymentGatewayService.HandleWebhook(paymentMethod, webhookData)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to handle webhook", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Webhook handled successfully", webhookResponse)
}

// GetPaymentMethods gets available payment methods
// @Summary Get payment methods
// @Description Get list of available payment methods
// @Tags payments
// @Produce json
// @Success 200 {object} response.Response{data=[]model.PaymentMethodInfo}
// @Router /api/v1/payments/methods [get]
func (h *PaymentHandler) GetPaymentMethods(c *gin.Context) {
	paymentMethods := []model.PaymentMethodInfo{
		{
			Code:        string(model.PaymentMethodVietQR),
			Name:        "VietQR",
			DisplayName: "VietQR (PayOS)",
			Description: "Thanh toán qua mã QR VietQR",
			IsActive:    true,
			IsOnline:    true,
			FeeType:     "percentage",
			FeeValue:    0.5, // 0.5%
			MinAmount:   1000,
			MaxAmount:   50000000,
			Currency:    "VND",
		},
		{
			Code:        string(model.PaymentMethodCOD),
			Name:        "Cash on Delivery",
			DisplayName: "Thanh toán khi nhận hàng",
			Description: "Thanh toán bằng tiền mặt khi nhận hàng",
			IsActive:    true,
			IsOnline:    false,
			FeeType:     "none",
			FeeValue:    0,
			MinAmount:   1000,
			MaxAmount:   50000000,
			Currency:    "VND",
		},
	}

	response.SuccessResponse(c, http.StatusOK, "Payment methods retrieved successfully", paymentMethods)
}

// Helper functions

func isValidPaymentMethod(method model.PaymentMethod) bool {
	validMethods := []model.PaymentMethod{
		model.PaymentMethodVietQR,
		model.PaymentMethodCOD,
	}

	for _, validMethod := range validMethods {
		if method == validMethod {
			return true
		}
	}
	return false
}

func convertOrderItemsToModel(items []model.OrderItemResponse) []model.OrderItem {
	var orderItems []model.OrderItem
	for _, item := range items {
		orderItems = append(orderItems, model.OrderItem{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TotalPrice:  item.TotalPrice,
		})
	}
	return orderItems
}
