package handler

import (
	"net/http"
	"strconv"

	"go_app/internal/model"
	"go_app/internal/repository"
	"go_app/internal/service"
	"go_app/pkg/response"

	"github.com/gin-gonic/gin"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
	orderService service.OrderService
	eventService service.EventService
}

// NewOrderHandler creates a new OrderHandler
func NewOrderHandler() *OrderHandler {
	// Initialize services
	orderService := service.NewOrderService()
	notificationService := service.NewNotificationService(repository.NewNotificationRepository(), repository.NewUserRepository())
	eventService := service.NewEventService(notificationService, nil, nil)

	return &OrderHandler{
		orderService: orderService,
		eventService: eventService,
	}
}

// Orders

// CreateOrder creates a new order
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req model.OrderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	order, err := h.orderService.CreateOrder(&req, userID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create order", err.Error())
		return
	}

	// Send order created notification
	// TODO: Convert OrderResponse to Order model for notification
	// For now, just log the event

	response.SuccessResponse(c, http.StatusCreated, "Order created successfully", order)
}

// CreateOrderForUser creates an order for a specific user (Admin only)
func (h *OrderHandler) CreateOrderForUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	var req model.OrderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Set the target user ID
	userIDUint := uint(userID)
	req.UserID = &userIDUint

	// Get current user ID from context
	currentUserID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	order, err := h.orderService.CreateOrder(&req, currentUserID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create order for user", err.Error())
		return
	}

	// Send order created notification
	// TODO: Convert OrderResponse to Order model for notification
	// For now, just log the event

	response.SuccessResponse(c, http.StatusCreated, "Order created successfully for user", order)
}

// GetOrderByID retrieves an order by its ID
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID", err.Error())
		return
	}

	order, err := h.orderService.GetOrderByID(uint(id))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve order", err.Error())
		return
	}

	if order == nil {
		response.ErrorResponse(c, http.StatusNotFound, "Order not found", "order not found")
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Order retrieved successfully", order)
}

// GetOrderByOrderNumber retrieves an order by its order number
func (h *OrderHandler) GetOrderByOrderNumber(c *gin.Context) {
	orderNumber := c.Param("order_number")
	if orderNumber == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Order number is required", "order_number parameter is required")
		return
	}

	order, err := h.orderService.GetOrderByOrderNumber(orderNumber)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve order", err.Error())
		return
	}

	if order == nil {
		response.ErrorResponse(c, http.StatusNotFound, "Order not found", "order not found")
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Order retrieved successfully", order)
}

// GetAllOrders retrieves all orders with pagination and filters
func (h *OrderHandler) GetAllOrders(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Parse filters
	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if paymentStatus := c.Query("payment_status"); paymentStatus != "" {
		filters["payment_status"] = paymentStatus
	}
	if shippingStatus := c.Query("shipping_status"); shippingStatus != "" {
		filters["shipping_status"] = shippingStatus
	}
	if paymentMethod := c.Query("payment_method"); paymentMethod != "" {
		filters["payment_method"] = paymentMethod
	}
	if userID := c.Query("user_id"); userID != "" {
		if id, err := strconv.ParseUint(userID, 10, 32); err == nil {
			filters["user_id"] = uint(id)
		}
	}
	if customerEmail := c.Query("customer_email"); customerEmail != "" {
		filters["customer_email"] = customerEmail
	}
	if customerPhone := c.Query("customer_phone"); customerPhone != "" {
		filters["customer_phone"] = customerPhone
	}
	if orderNumber := c.Query("order_number"); orderNumber != "" {
		filters["order_number"] = orderNumber
	}
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		filters["date_from"] = dateFrom
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		filters["date_to"] = dateTo
	}
	if minAmount := c.Query("min_amount"); minAmount != "" {
		if amount, err := strconv.ParseFloat(minAmount, 64); err == nil {
			filters["min_amount"] = amount
		}
	}
	if maxAmount := c.Query("max_amount"); maxAmount != "" {
		if amount, err := strconv.ParseFloat(maxAmount, 64); err == nil {
			filters["max_amount"] = amount
		}
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	orders, total, err := h.orderService.GetAllOrders(page, limit, filters)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve orders", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Orders retrieved successfully", orders, page, limit, total)
}

// GetOrdersByUser retrieves orders for a specific user
func (h *OrderHandler) GetOrdersByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Parse filters
	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if paymentStatus := c.Query("payment_status"); paymentStatus != "" {
		filters["payment_status"] = paymentStatus
	}
	if shippingStatus := c.Query("shipping_status"); shippingStatus != "" {
		filters["shipping_status"] = shippingStatus
	}

	orders, total, err := h.orderService.GetOrdersByUser(uint(userID), page, limit, filters)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve user orders", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "User orders retrieved successfully", orders, page, limit, total)
}

// UpdateOrder updates an existing order
func (h *OrderHandler) UpdateOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID", err.Error())
		return
	}

	var req model.OrderUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	// Get old order to compare status changes
	oldOrder, err := h.orderService.GetOrderByID(uint(id))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve order", err.Error())
		return
	}

	order, err := h.orderService.UpdateOrder(uint(id), &req, userID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update order", err.Error())
		return
	}

	// Send notification if status changed
	// TODO: Convert OrderResponse to Order model for notification
	// For now, just log the event
	if req.Status != nil && *req.Status != oldOrder.Status {
		// Log status change
	}

	response.SuccessResponse(c, http.StatusOK, "Order updated successfully", order)
}

// DeleteOrder deletes an order
func (h *OrderHandler) DeleteOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	if err := h.orderService.DeleteOrder(uint(id), userID.(uint)); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete order", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Order deleted successfully", nil)
}

// CancelOrder cancels an order
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID", err.Error())
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	if err := h.orderService.CancelOrder(uint(id), userID.(uint), req.Reason); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to cancel order", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Order cancelled successfully", nil)
}

// ConfirmOrder confirms an order
func (h *OrderHandler) ConfirmOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	if err := h.orderService.ConfirmOrder(uint(id), userID.(uint)); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to confirm order", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Order confirmed successfully", nil)
}

// ShipOrder ships an order
func (h *OrderHandler) ShipOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID", err.Error())
		return
	}

	var req struct {
		TrackingNumber string `json:"tracking_number" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	if err := h.orderService.ShipOrder(uint(id), userID.(uint), req.TrackingNumber); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to ship order", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Order shipped successfully", nil)
}

// DeliverOrder delivers an order
func (h *OrderHandler) DeliverOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	if err := h.orderService.DeliverOrder(uint(id), userID.(uint)); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to deliver order", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Order delivered successfully", nil)
}

// GetOrderItems retrieves order items for an order
func (h *OrderHandler) GetOrderItems(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID", err.Error())
		return
	}

	items, err := h.orderService.GetOrderItems(uint(orderID))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve order items", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Order items retrieved successfully", items)
}

// Cart Management

// CreateCart creates a new cart
func (h *OrderHandler) CreateCart(c *gin.Context) {
	var req model.CartCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	cart, err := h.orderService.CreateCart(&req, userID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create cart", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Cart created successfully", cart)
}

// GetCart retrieves a cart
func (h *OrderHandler) GetCart(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	// Get session ID from query parameter (for guest users)
	sessionID := c.Query("session_id")

	cart, err := h.orderService.GetCart(userID.(uint), sessionID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve cart", err.Error())
		return
	}

	if cart == nil {
		response.ErrorResponse(c, http.StatusNotFound, "Cart not found", "cart not found")
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Cart retrieved successfully", cart)
}

// UpdateCart updates a cart
func (h *OrderHandler) UpdateCart(c *gin.Context) {
	cartIDStr := c.Param("cart_id")
	cartID, err := strconv.ParseUint(cartIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid cart ID", err.Error())
		return
	}

	var req model.CartCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	cart, err := h.orderService.UpdateCart(uint(cartID), &req, userID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update cart", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Cart updated successfully", cart)
}

// DeleteCart deletes a cart
func (h *OrderHandler) DeleteCart(c *gin.Context) {
	cartIDStr := c.Param("cart_id")
	cartID, err := strconv.ParseUint(cartIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid cart ID", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	if err := h.orderService.DeleteCart(uint(cartID), userID.(uint)); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete cart", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Cart deleted successfully", nil)
}

// ClearCart clears a cart
func (h *OrderHandler) ClearCart(c *gin.Context) {
	cartIDStr := c.Param("cart_id")
	cartID, err := strconv.ParseUint(cartIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid cart ID", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	if err := h.orderService.ClearCart(uint(cartID), userID.(uint)); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to clear cart", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Cart cleared successfully", nil)
}

// AddToCart adds an item to cart
func (h *OrderHandler) AddToCart(c *gin.Context) {
	cartIDStr := c.Param("cart_id")
	cartID, err := strconv.ParseUint(cartIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid cart ID", err.Error())
		return
	}

	var req model.CartItemCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	item, err := h.orderService.AddToCart(uint(cartID), &req, userID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to add item to cart", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Item added to cart successfully", item)
}

// UpdateCartItem updates a cart item
func (h *OrderHandler) UpdateCartItem(c *gin.Context) {
	cartIDStr := c.Param("cart_id")
	cartID, err := strconv.ParseUint(cartIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid cart ID", err.Error())
		return
	}

	itemIDStr := c.Param("item_id")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid item ID", err.Error())
		return
	}

	var req model.CartItemCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	item, err := h.orderService.UpdateCartItem(uint(cartID), uint(itemID), &req, userID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update cart item", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Cart item updated successfully", item)
}

// RemoveFromCart removes an item from cart
func (h *OrderHandler) RemoveFromCart(c *gin.Context) {
	cartIDStr := c.Param("cart_id")
	cartID, err := strconv.ParseUint(cartIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid cart ID", err.Error())
		return
	}

	itemIDStr := c.Param("item_id")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid item ID", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	if err := h.orderService.RemoveFromCart(uint(cartID), uint(itemID), userID.(uint)); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove item from cart", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Item removed from cart successfully", nil)
}

// GetCartItems retrieves cart items
func (h *OrderHandler) GetCartItems(c *gin.Context) {
	cartIDStr := c.Param("cart_id")
	cartID, err := strconv.ParseUint(cartIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid cart ID", err.Error())
		return
	}

	items, err := h.orderService.GetCartItems(uint(cartID))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve cart items", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Cart items retrieved successfully", items)
}

// ConvertCartToOrder converts a cart to an order
func (h *OrderHandler) ConvertCartToOrder(c *gin.Context) {
	cartIDStr := c.Param("cart_id")
	cartID, err := strconv.ParseUint(cartIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid cart ID", err.Error())
		return
	}

	var req model.OrderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Set cart ID
	cartIDUint := uint(cartID)
	req.CartID = &cartIDUint

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	order, err := h.orderService.ConvertCartToOrder(uint(cartID), &req, userID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to convert cart to order", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Cart converted to order successfully", order)
}

// Statistics

// GetOrderStats retrieves order statistics
func (h *OrderHandler) GetOrderStats(c *gin.Context) {
	stats, err := h.orderService.GetOrderStats()
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve order statistics", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Order statistics retrieved successfully", stats)
}

// GetOrderStatsByUser retrieves order statistics for a specific user
func (h *OrderHandler) GetOrderStatsByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	stats, err := h.orderService.GetOrderStatsByUser(uint(userID))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve user order statistics", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "User order statistics retrieved successfully", stats)
}

// GetRevenueStats retrieves revenue statistics
func (h *OrderHandler) GetRevenueStats(c *gin.Context) {
	stats, err := h.orderService.GetRevenueStats()
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve revenue statistics", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Revenue statistics retrieved successfully", stats)
}
