package service

import (
	"errors"
	"fmt"
	"go_app/internal/model"
	"go_app/internal/repository"
	"go_app/pkg/logger"
	"math/rand"
	"time"
)

// OrderService defines methods for order business logic
type OrderService interface {
	// Orders
	CreateOrder(req *model.OrderCreateRequest, userID uint) (*model.OrderResponse, error)
	GetOrderByID(id uint) (*model.OrderResponse, error)
	GetOrderByOrderNumber(orderNumber string) (*model.OrderResponse, error)
	GetAllOrders(page, limit int, filters map[string]interface{}) ([]model.OrderResponse, int64, error)
	GetOrdersByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.OrderResponse, int64, error)
	UpdateOrder(id uint, req *model.OrderUpdateRequest, userID uint) (*model.OrderResponse, error)
	DeleteOrder(id uint, userID uint) error
	CancelOrder(id uint, userID uint, reason string) error
	ConfirmOrder(id uint, userID uint) error
	ShipOrder(id uint, userID uint, trackingNumber string) error
	DeliverOrder(id uint, userID uint) error

	// Order Items
	AddOrderItem(orderID uint, req *model.OrderItemCreateRequest, userID uint) (*model.OrderItemResponse, error)
	UpdateOrderItem(orderID, itemID uint, req *model.OrderItemCreateRequest, userID uint) (*model.OrderItemResponse, error)
	RemoveOrderItem(orderID, itemID uint, userID uint) error
	GetOrderItems(orderID uint) ([]model.OrderItemResponse, error)

	// Cart
	CreateCart(req *model.CartCreateRequest, userID uint) (*model.CartResponse, error)
	GetCart(userID uint, sessionID string) (*model.CartResponse, error)
	UpdateCart(cartID uint, req *model.CartUpdateRequest, userID uint) (*model.CartResponse, error)
	DeleteCart(cartID uint, userID uint) error
	ClearCart(cartID uint, userID uint) error

	// Cart Items
	AddToCart(cartID uint, req *model.CartItemCreateRequest, userID uint) (*model.CartItemResponse, error)
	UpdateCartItem(cartID, itemID uint, req *model.CartItemCreateRequest, userID uint) (*model.CartItemResponse, error)
	RemoveFromCart(cartID, itemID uint, userID uint) error
	GetCartItems(cartID uint) ([]model.CartItemResponse, error)
	SyncCartWithUser(cartID, userID uint) error
	GetCartStats() (map[string]interface{}, error)

	// Payments
	CreatePayment(req *model.PaymentCreateRequest, userID uint) (*model.PaymentResponse, error)
	ProcessPayment(paymentID uint, userID uint) (*model.PaymentResponse, error)
	RefundPayment(paymentID uint, userID uint, reason string) (*model.PaymentResponse, error)
	GetPaymentsByOrder(orderID uint) ([]model.PaymentResponse, error)

	// Shipping
	UpdateShippingStatus(orderID uint, status model.ShippingStatus, userID uint, description, location, notes string) error
	GetShippingHistory(orderID uint) ([]model.ShippingHistoryResponse, error)

	// Statistics
	GetOrderStats() (*model.OrderStatsResponse, error)
	GetOrderStatsByUser(userID uint) (map[string]interface{}, error)
	GetRevenueStats() (map[string]interface{}, error)

	// Utility
	ConvertCartToOrder(cartID uint, req *model.OrderCreateRequest, userID uint) (*model.OrderResponse, error)
	GenerateOrderNumber() string
	CalculateOrderTotal(order *model.Order) error
	ValidateOrder(order *model.Order) error
}

// orderService implements OrderService
type orderService struct {
	orderRepo     repository.OrderRepository
	productRepo   *repository.ProductRepository
	inventoryRepo repository.InventoryRepository
	userRepo      repository.UserRepository
	eventService  EventService
}

// NewOrderService creates a new OrderService
func NewOrderService() OrderService {
	return &orderService{
		orderRepo:     repository.NewOrderRepository(),
		productRepo:   repository.NewProductRepository(),
		inventoryRepo: repository.NewInventoryRepository(),
		userRepo:      repository.NewUserRepository(),
		eventService:  nil, // Will be set by dependency injection
	}
}

// NewOrderServiceWithEvent creates a new OrderService with EventService
func NewOrderServiceWithEvent(eventService EventService) OrderService {
	return &orderService{
		orderRepo:     repository.NewOrderRepository(),
		productRepo:   repository.NewProductRepository(),
		inventoryRepo: repository.NewInventoryRepository(),
		userRepo:      repository.NewUserRepository(),
		eventService:  eventService,
	}
}

// Orders

// CreateOrder creates a new order
func (s *orderService) CreateOrder(req *model.OrderCreateRequest, userID uint) (*model.OrderResponse, error) {
	// Determine the target user ID for the order
	targetUserID := userID
	if req.UserID != nil {
		// Check if current user has permission to create orders for other users
		currentUser, err := s.userRepo.GetByID(userID)
		if err != nil {
			logger.Errorf("Error getting current user by ID %d: %v", userID, err)
			return nil, fmt.Errorf("failed to retrieve current user")
		}
		if currentUser == nil {
			return nil, errors.New("current user not found")
		}

		// Only admin users can create orders for other users
		if currentUser.Role != "admin" && currentUser.Role != "super_admin" {
			return nil, errors.New("only admin users can create orders for other users")
		}

		targetUserID = *req.UserID
	}

	// Get target user information
	targetUser, err := s.userRepo.GetByID(targetUserID)
	if err != nil {
		logger.Errorf("Error getting target user by ID %d: %v", targetUserID, err)
		return nil, fmt.Errorf("failed to retrieve target user")
	}
	if targetUser == nil {
		return nil, errors.New("target user not found")
	}

	// Generate order number
	orderNumber := s.GenerateOrderNumber()

	// Create order
	order := &model.Order{
		OrderNumber:     orderNumber,
		UserID:          targetUserID,
		Status:          model.OrderStatusPending,
		PaymentStatus:   model.PaymentStatusPending,
		ShippingStatus:  model.ShippingStatusPending,
		CustomerName:    req.CustomerName,
		CustomerEmail:   req.CustomerEmail,
		CustomerPhone:   req.CustomerPhone,
		ShippingAddress: req.ShippingAddress,
		BillingAddress:  req.BillingAddress,
		PaymentMethod:   req.PaymentMethod,
		ShippingMethod:  req.ShippingMethod,
		Notes:           req.Notes,
		SubTotal:        0,
		TaxAmount:       0,
		ShippingCost:    0,
		DiscountAmount:  0,
		TotalAmount:     0,
	}

	// If creating from cart, copy items
	if req.CartID != nil {
		cart, err := s.orderRepo.GetCartByID(*req.CartID)
		if err != nil {
			logger.Errorf("Error getting cart by ID %d: %v", *req.CartID, err)
			return nil, fmt.Errorf("failed to retrieve cart")
		}
		if cart == nil {
			return nil, errors.New("cart not found")
		}

		// Copy cart items to order
		cartItems, err := s.orderRepo.GetCartItemsByCart(*req.CartID)
		if err != nil {
			logger.Errorf("Error getting cart items: %v", err)
			return nil, fmt.Errorf("failed to retrieve cart items")
		}

		order.SubTotal = cart.SubTotal
		order.TaxAmount = cart.TaxAmount
		order.ShippingCost = cart.ShippingCost
		order.DiscountAmount = cart.DiscountAmount
		order.CalculateTotal()

		// Create order in database
		if err := s.orderRepo.CreateOrder(order); err != nil {
			logger.Errorf("Error creating order: %v", err)
			return nil, fmt.Errorf("failed to create order")
		}

		// Create order items
		for _, cartItem := range cartItems {
			orderItem := &model.OrderItem{
				OrderID:          order.ID,
				ProductID:        cartItem.ProductID,
				ProductVariantID: cartItem.ProductVariantID,
				ProductName:      cartItem.Product.Name,
				ProductSKU:       cartItem.Product.SKU,
				ProductImage:     "", // Will be set from product images
				VariantName:      s.getVariantName(cartItem.ProductVariant),
				UnitPrice:        cartItem.UnitPrice,
				Quantity:         cartItem.Quantity,
				TotalPrice:       cartItem.TotalPrice,
			}
			orderItem.CalculateTotal()

			if err := s.orderRepo.CreateOrderItem(orderItem); err != nil {
				logger.Errorf("Error creating order item: %v", err)
				return nil, fmt.Errorf("failed to create order item")
			}
		}

		// Clear cart after order creation
		if err := s.orderRepo.ClearCart(*req.CartID); err != nil {
			logger.Warnf("Failed to clear cart after order creation: %v", err)
		}
	} else {
		// Create empty order
		if err := s.orderRepo.CreateOrder(order); err != nil {
			logger.Errorf("Error creating order: %v", err)
			return nil, fmt.Errorf("failed to create order")
		}
	}

	// Validate order
	if err := s.ValidateOrder(order); err != nil {
		logger.Errorf("Order validation failed: %v", err)
		return nil, err
	}

	// Get created order with relations
	createdOrder, err := s.orderRepo.GetOrderByID(order.ID)
	if err != nil {
		logger.Errorf("Error getting created order: %v", err)
		return nil, fmt.Errorf("failed to retrieve created order")
	}

	// Trigger order created event
	if s.eventService != nil {
		if err := s.eventService.OnOrderCreated(createdOrder); err != nil {
			logger.Errorf("Failed to trigger order created event: %v", err)
			// Don't return error, just log it
		}
	}

	return s.toOrderResponse(createdOrder), nil
}

// GetOrderByID retrieves an order by its ID
func (s *orderService) GetOrderByID(id uint) (*model.OrderResponse, error) {
	order, err := s.orderRepo.GetOrderByID(id)
	if err != nil {
		logger.Errorf("Error getting order by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve order")
	}
	if order == nil {
		return nil, errors.New("order not found")
	}
	return s.toOrderResponse(order), nil
}

// GetOrderByOrderNumber retrieves an order by its order number
func (s *orderService) GetOrderByOrderNumber(orderNumber string) (*model.OrderResponse, error) {
	order, err := s.orderRepo.GetOrderByOrderNumber(orderNumber)
	if err != nil {
		logger.Errorf("Error getting order by order number %s: %v", orderNumber, err)
		return nil, fmt.Errorf("failed to retrieve order")
	}
	if order == nil {
		return nil, errors.New("order not found")
	}
	return s.toOrderResponse(order), nil
}

// GetAllOrders retrieves all orders with pagination and filters
func (s *orderService) GetAllOrders(page, limit int, filters map[string]interface{}) ([]model.OrderResponse, int64, error) {
	orders, total, err := s.orderRepo.GetAllOrders(page, limit, filters)
	if err != nil {
		logger.Errorf("Error getting orders: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve orders")
	}

	var responses []model.OrderResponse
	for _, order := range orders {
		responses = append(responses, *s.toOrderResponse(&order))
	}
	return responses, total, nil
}

// GetOrdersByUser retrieves orders for a specific user
func (s *orderService) GetOrdersByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.OrderResponse, int64, error) {
	orders, total, err := s.orderRepo.GetOrdersByUser(userID, page, limit, filters)
	if err != nil {
		logger.Errorf("Error getting user orders: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve user orders")
	}

	var responses []model.OrderResponse
	for _, order := range orders {
		responses = append(responses, *s.toOrderResponse(&order))
	}
	return responses, total, nil
}

// UpdateOrder updates an existing order
func (s *orderService) UpdateOrder(id uint, req *model.OrderUpdateRequest, userID uint) (*model.OrderResponse, error) {
	order, err := s.orderRepo.GetOrderByID(id)
	if err != nil {
		logger.Errorf("Error getting order by ID %d for update: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve order")
	}
	if order == nil {
		return nil, errors.New("order not found")
	}

	// Check if order can be updated
	if order.IsCompleted() || order.IsCancelled() {
		return nil, errors.New("cannot update completed or cancelled order")
	}

	// Store old status for event trigger
	oldStatus := order.Status

	// Update fields
	if req.Status != nil {
		order.Status = *req.Status
	}
	if req.PaymentStatus != nil {
		order.PaymentStatus = *req.PaymentStatus
	}
	if req.ShippingStatus != nil {
		order.ShippingStatus = *req.ShippingStatus
	}
	if req.CustomerName != "" {
		order.CustomerName = req.CustomerName
	}
	if req.CustomerEmail != "" {
		order.CustomerEmail = req.CustomerEmail
	}
	if req.CustomerPhone != "" {
		order.CustomerPhone = req.CustomerPhone
	}
	if req.ShippingAddress != "" {
		order.ShippingAddress = req.ShippingAddress
	}
	if req.BillingAddress != "" {
		order.BillingAddress = req.BillingAddress
	}
	if req.TrackingNumber != "" {
		order.TrackingNumber = req.TrackingNumber
	}
	if req.ShippingMethod != "" {
		order.ShippingMethod = req.ShippingMethod
	}
	if req.Notes != "" {
		order.Notes = req.Notes
	}
	if req.AdminNotes != "" {
		order.AdminNotes = req.AdminNotes
	}
	if req.Tags != "" {
		order.Tags = req.Tags
	}

	// Update timestamps based on status changes
	if req.Status != nil {
		switch *req.Status {
		case model.OrderStatusConfirmed:
			// Order confirmed
		case model.OrderStatusShipped:
			now := time.Now()
			order.ShippedAt = &now
		case model.OrderStatusDelivered:
			now := time.Now()
			order.DeliveredAt = &now
		}
	}

	if err := s.orderRepo.UpdateOrder(order); err != nil {
		logger.Errorf("Error updating order %d: %v", id, err)
		return nil, fmt.Errorf("failed to update order")
	}

	// Get updated order with relations
	updatedOrder, err := s.orderRepo.GetOrderByID(order.ID)
	if err != nil {
		logger.Errorf("Error getting updated order: %v", err)
		return nil, fmt.Errorf("failed to retrieve updated order")
	}

	// Trigger order status updated event if status changed
	if s.eventService != nil && req.Status != nil && *req.Status != oldStatus {
		if err := s.eventService.OnOrderStatusUpdated(updatedOrder, oldStatus, *req.Status); err != nil {
			logger.Errorf("Failed to trigger order status updated event: %v", err)
			// Don't return error, just log it
		}
	}

	return s.toOrderResponse(updatedOrder), nil
}

// DeleteOrder deletes an order
func (s *orderService) DeleteOrder(id uint, userID uint) error {
	order, err := s.orderRepo.GetOrderByID(id)
	if err != nil {
		logger.Errorf("Error getting order by ID %d for deletion: %v", id, err)
		return fmt.Errorf("failed to retrieve order")
	}
	if order == nil {
		return errors.New("order not found")
	}

	// Check if order can be deleted
	if order.IsCompleted() {
		return errors.New("cannot delete completed order")
	}

	if err := s.orderRepo.DeleteOrder(id); err != nil {
		logger.Errorf("Error deleting order %d: %v", id, err)
		return fmt.Errorf("failed to delete order")
	}

	return nil
}

// CancelOrder cancels an order
func (s *orderService) CancelOrder(id uint, userID uint, reason string) error {
	order, err := s.orderRepo.GetOrderByID(id)
	if err != nil {
		logger.Errorf("Error getting order by ID %d for cancellation: %v", id, err)
		return fmt.Errorf("failed to retrieve order")
	}
	if order == nil {
		return errors.New("order not found")
	}

	// Check if order can be cancelled
	if !order.CanBeCancelled() {
		return errors.New("order cannot be cancelled")
	}

	// Update order status
	order.Status = model.OrderStatusCancelled
	order.AdminNotes = fmt.Sprintf("Order cancelled by user %d. Reason: %s", userID, reason)

	if err := s.orderRepo.UpdateOrder(order); err != nil {
		logger.Errorf("Error cancelling order %d: %v", id, err)
		return fmt.Errorf("failed to cancel order")
	}

	// Restore inventory if needed
	if err := s.restoreInventoryForOrder(order); err != nil {
		logger.Warnf("Failed to restore inventory for cancelled order %d: %v", id, err)
	}

	return nil
}

// ConfirmOrder confirms an order
func (s *orderService) ConfirmOrder(id uint, userID uint) error {
	order, err := s.orderRepo.GetOrderByID(id)
	if err != nil {
		logger.Errorf("Error getting order by ID %d for confirmation: %v", id, err)
		return fmt.Errorf("failed to retrieve order")
	}
	if order == nil {
		return errors.New("order not found")
	}

	if order.Status != model.OrderStatusPending {
		return errors.New("order is not in pending status")
	}

	// Reserve inventory
	if err := s.reserveInventoryForOrder(order); err != nil {
		logger.Errorf("Failed to reserve inventory for order %d: %v", id, err)
		return fmt.Errorf("failed to reserve inventory: %v", err)
	}

	// Update order status
	order.Status = model.OrderStatusConfirmed

	if err := s.orderRepo.UpdateOrder(order); err != nil {
		logger.Errorf("Error confirming order %d: %v", id, err)
		return fmt.Errorf("failed to confirm order")
	}

	return nil
}

// ShipOrder ships an order
func (s *orderService) ShipOrder(id uint, userID uint, trackingNumber string) error {
	order, err := s.orderRepo.GetOrderByID(id)
	if err != nil {
		logger.Errorf("Error getting order by ID %d for shipping: %v", id, err)
		return fmt.Errorf("failed to retrieve order")
	}
	if order == nil {
		return errors.New("order not found")
	}

	if !order.CanBeShipped() {
		return errors.New("order cannot be shipped")
	}

	// Update order status
	order.Status = model.OrderStatusShipped
	order.ShippingStatus = model.ShippingStatusInTransit
	order.TrackingNumber = trackingNumber
	now := time.Now()
	order.ShippedAt = &now

	if err := s.orderRepo.UpdateOrder(order); err != nil {
		logger.Errorf("Error shipping order %d: %v", id, err)
		return fmt.Errorf("failed to ship order")
	}

	// Create shipping history entry
	history := &model.ShippingHistory{
		OrderID:     order.ID,
		Status:      model.ShippingStatusInTransit,
		Description: "Order shipped",
		Location:    "Warehouse",
		Notes:       fmt.Sprintf("Tracking number: %s", trackingNumber),
		UpdatedBy:   userID,
	}

	if err := s.orderRepo.CreateShippingHistory(history); err != nil {
		logger.Warnf("Failed to create shipping history for order %d: %v", id, err)
	}

	return nil
}

// DeliverOrder delivers an order
func (s *orderService) DeliverOrder(id uint, userID uint) error {
	order, err := s.orderRepo.GetOrderByID(id)
	if err != nil {
		logger.Errorf("Error getting order by ID %d for delivery: %v", id, err)
		return fmt.Errorf("failed to retrieve order")
	}
	if order == nil {
		return errors.New("order not found")
	}

	if !order.CanBeDelivered() {
		return errors.New("order cannot be delivered")
	}

	// Update order status
	order.Status = model.OrderStatusDelivered
	order.ShippingStatus = model.ShippingStatusDelivered
	now := time.Now()
	order.DeliveredAt = &now

	if err := s.orderRepo.UpdateOrder(order); err != nil {
		logger.Errorf("Error delivering order %d: %v", id, err)
		return fmt.Errorf("failed to deliver order")
	}

	// Create shipping history entry
	history := &model.ShippingHistory{
		OrderID:     order.ID,
		Status:      model.ShippingStatusDelivered,
		Description: "Order delivered successfully",
		Location:    order.ShippingAddress,
		Notes:       "Order delivered to customer",
		UpdatedBy:   userID,
	}

	if err := s.orderRepo.CreateShippingHistory(history); err != nil {
		logger.Warnf("Failed to create shipping history for order %d: %v", id, err)
	}

	return nil
}

// Helper methods

// GenerateOrderNumber generates a unique order number
func (s *orderService) GenerateOrderNumber() string {
	// Format: ORD-YYYYMMDD-XXXXXX
	now := time.Now()
	dateStr := now.Format("20060102")
	randomNum := rand.Intn(999999)
	return fmt.Sprintf("ORD-%s-%06d", dateStr, randomNum)
}

// CalculateOrderTotal calculates total amount for order
func (s *orderService) CalculateOrderTotal(order *model.Order) error {
	// Get order items
	orderItems, err := s.orderRepo.GetOrderItemsByOrder(order.ID)
	if err != nil {
		return err
	}

	// Calculate subtotal
	subTotal := 0.0
	for _, item := range orderItems {
		subTotal += item.TotalPrice
	}

	// Calculate tax (simplified - 10% VAT)
	taxAmount := subTotal * 0.1

	// Update order totals
	order.SubTotal = subTotal
	order.TaxAmount = taxAmount
	order.CalculateTotal()

	return nil
}

// ValidateOrder validates order data
func (s *orderService) ValidateOrder(order *model.Order) error {
	if order.CustomerName == "" {
		return errors.New("customer name is required")
	}
	if order.CustomerEmail == "" {
		return errors.New("customer email is required")
	}
	if order.CustomerPhone == "" {
		return errors.New("customer phone is required")
	}
	if order.ShippingAddress == "" {
		return errors.New("shipping address is required")
	}
	if order.TotalAmount <= 0 {
		return errors.New("total amount must be greater than 0")
	}
	return nil
}

// getVariantName returns variant name string
func (s *orderService) getVariantName(variant *model.ProductVariant) string {
	if variant == nil {
		return ""
	}
	return variant.Name
}

// reserveInventoryForOrder reserves inventory for order items
func (s *orderService) reserveInventoryForOrder(order *model.Order) error {
	orderItems, err := s.orderRepo.GetOrderItemsByOrder(order.ID)
	if err != nil {
		return err
	}

	for _, item := range orderItems {
		// Create inventory movement for reservation
		movement := &model.InventoryMovement{
			ProductID:     item.ProductID,
			Type:          model.MovementTypeOutbound,
			Quantity:      -item.Quantity, // Negative for outbound
			Reference:     order.OrderNumber,
			ReferenceType: "order",
			Status:        model.MovementStatusCompleted,
		}

		if err := s.inventoryRepo.CreateMovement(movement); err != nil {
			return fmt.Errorf("failed to create inventory movement for product %d: %v", item.ProductID, err)
		}
	}

	return nil
}

// restoreInventoryForOrder restores inventory for cancelled order
func (s *orderService) restoreInventoryForOrder(order *model.Order) error {
	orderItems, err := s.orderRepo.GetOrderItemsByOrder(order.ID)
	if err != nil {
		return err
	}

	for _, item := range orderItems {
		// Create inventory movement for restoration
		movement := &model.InventoryMovement{
			ProductID:     item.ProductID,
			Type:          model.MovementTypeReturn,
			Quantity:      item.Quantity, // Positive for return
			Reference:     order.OrderNumber,
			ReferenceType: "order_cancellation",
			Status:        model.MovementStatusCompleted,
		}

		if err := s.inventoryRepo.CreateMovement(movement); err != nil {
			return fmt.Errorf("failed to create inventory movement for product %d: %v", item.ProductID, err)
		}
	}

	return nil
}

// Response conversion methods

func (s *orderService) toOrderResponse(order *model.Order) *model.OrderResponse {
	response := &model.OrderResponse{
		ID:               order.ID,
		OrderNumber:      order.OrderNumber,
		UserID:           order.UserID,
		Status:           order.Status,
		PaymentStatus:    order.PaymentStatus,
		ShippingStatus:   order.ShippingStatus,
		CustomerName:     order.CustomerName,
		CustomerEmail:    order.CustomerEmail,
		CustomerPhone:    order.CustomerPhone,
		ShippingAddress:  order.ShippingAddress,
		BillingAddress:   order.BillingAddress,
		SubTotal:         order.SubTotal,
		TaxAmount:        order.TaxAmount,
		ShippingCost:     order.ShippingCost,
		DiscountAmount:   order.DiscountAmount,
		TotalAmount:      order.TotalAmount,
		PaymentMethod:    order.PaymentMethod,
		PaymentReference: order.PaymentReference,
		PaidAt:           order.PaidAt,
		ShippingMethod:   order.ShippingMethod,
		TrackingNumber:   order.TrackingNumber,
		ShippedAt:        order.ShippedAt,
		DeliveredAt:      order.DeliveredAt,
		Notes:            order.Notes,
		AdminNotes:       order.AdminNotes,
		Tags:             order.Tags,
		CreatedAt:        order.CreatedAt,
		UpdatedAt:        order.UpdatedAt,
	}

	if order.User != nil {
		response.UserName = order.User.Username
	}

	// Convert order items
	if len(order.OrderItems) > 0 {
		var orderItemResponses []model.OrderItemResponse
		for _, item := range order.OrderItems {
			orderItemResponses = append(orderItemResponses, *s.toOrderItemResponse(&item))
		}
		response.OrderItems = orderItemResponses
	}

	// Convert payments
	if len(order.Payments) > 0 {
		var paymentResponses []model.PaymentResponse
		for _, payment := range order.Payments {
			paymentResponses = append(paymentResponses, *s.toPaymentResponse(&payment))
		}
		response.Payments = paymentResponses
	}

	// Convert shipping history
	if len(order.ShippingHistory) > 0 {
		var shippingResponses []model.ShippingHistoryResponse
		for _, history := range order.ShippingHistory {
			shippingResponses = append(shippingResponses, *s.toShippingHistoryResponse(&history))
		}
		response.ShippingHistory = shippingResponses
	}

	return response
}

func (s *orderService) toOrderItemResponse(item *model.OrderItem) *model.OrderItemResponse {
	return &model.OrderItemResponse{
		ID:               item.ID,
		OrderID:          item.OrderID,
		ProductID:        item.ProductID,
		ProductName:      item.ProductName,
		ProductSKU:       item.ProductSKU,
		ProductImage:     item.ProductImage,
		ProductVariantID: item.ProductVariantID,
		VariantName:      item.VariantName,
		UnitPrice:        item.UnitPrice,
		Quantity:         item.Quantity,
		TotalPrice:       item.TotalPrice,
		Weight:           item.Weight,
		Dimensions:       item.Dimensions,
		Notes:            item.Notes,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}
}

func (s *orderService) toPaymentResponse(payment *model.Payment) *model.PaymentResponse {
	return &model.PaymentResponse{
		ID:            payment.ID,
		OrderID:       payment.OrderID,
		UserID:        payment.UserID,
		PaymentMethod: payment.PaymentMethod,
		Status:        payment.Status,
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		TransactionID: payment.TransactionID,
		ReferenceID:   payment.ReferenceID,
		Description:   payment.Description,
		Notes:         payment.Notes,
		ProcessedAt:   payment.ProcessedAt,
		CreatedAt:     payment.CreatedAt,
		UpdatedAt:     payment.UpdatedAt,
	}
}

func (s *orderService) toShippingHistoryResponse(history *model.ShippingHistory) *model.ShippingHistoryResponse {
	response := &model.ShippingHistoryResponse{
		ID:          history.ID,
		OrderID:     history.OrderID,
		Status:      history.Status,
		Description: history.Description,
		Location:    history.Location,
		Notes:       history.Notes,
		UpdatedBy:   history.UpdatedBy,
		CreatedAt:   history.CreatedAt,
	}

	if history.UpdatedByUser != nil {
		response.UpdatedByName = history.UpdatedByUser.Username
	}

	return response
}

// Placeholder methods for remaining interface methods
// These would be implemented similarly to the above methods

func (s *orderService) AddOrderItem(orderID uint, req *model.OrderItemCreateRequest, userID uint) (*model.OrderItemResponse, error) {
	// Implementation would go here
	return nil, errors.New("not implemented")
}

func (s *orderService) UpdateOrderItem(orderID, itemID uint, req *model.OrderItemCreateRequest, userID uint) (*model.OrderItemResponse, error) {
	// Implementation would go here
	return nil, errors.New("not implemented")
}

func (s *orderService) RemoveOrderItem(orderID, itemID uint, userID uint) error {
	// Implementation would go here
	return errors.New("not implemented")
}

func (s *orderService) GetOrderItems(orderID uint) ([]model.OrderItemResponse, error) {
	// Implementation would go here
	return nil, errors.New("not implemented")
}

func (s *orderService) CreateCart(req *model.CartCreateRequest, userID uint) (*model.CartResponse, error) {
	// Implementation would go here
	return nil, errors.New("not implemented")
}

func (s *orderService) GetCart(userID uint, sessionID string) (*model.CartResponse, error) {
	// Implementation would go here
	return nil, errors.New("not implemented")
}

func (s *orderService) UpdateCart(cartID uint, req *model.CartUpdateRequest, userID uint) (*model.CartResponse, error) {
	// Implementation would go here
	return nil, errors.New("not implemented")
}

func (s *orderService) DeleteCart(cartID uint, userID uint) error {
	// Implementation would go here
	return errors.New("not implemented")
}

func (s *orderService) ClearCart(cartID uint, userID uint) error {
	// Implementation would go here
	return errors.New("not implemented")
}

func (s *orderService) AddToCart(cartID uint, req *model.CartItemCreateRequest, userID uint) (*model.CartItemResponse, error) {
	// Implementation would go here
	return nil, errors.New("not implemented")
}

func (s *orderService) UpdateCartItem(cartID, itemID uint, req *model.CartItemCreateRequest, userID uint) (*model.CartItemResponse, error) {
	// Implementation would go here
	return nil, errors.New("not implemented")
}

func (s *orderService) RemoveFromCart(cartID, itemID uint, userID uint) error {
	// Implementation would go here
	return errors.New("not implemented")
}

func (s *orderService) GetCartItems(cartID uint) ([]model.CartItemResponse, error) {
	// Implementation would go here
	return nil, errors.New("not implemented")
}

func (s *orderService) SyncCartWithUser(cartID, userID uint) error {
	// Implementation would go here
	return errors.New("not implemented")
}

func (s *orderService) GetCartStats() (map[string]interface{}, error) {
	// Implementation would go here
	return nil, errors.New("not implemented")
}

func (s *orderService) CreatePayment(req *model.PaymentCreateRequest, userID uint) (*model.PaymentResponse, error) {
	// Implementation would go here
	return nil, errors.New("not implemented")
}

func (s *orderService) ProcessPayment(paymentID uint, userID uint) (*model.PaymentResponse, error) {
	// Implementation would go here
	return nil, errors.New("not implemented")
}

func (s *orderService) RefundPayment(paymentID uint, userID uint, reason string) (*model.PaymentResponse, error) {
	// Implementation would go here
	return nil, errors.New("not implemented")
}

func (s *orderService) GetPaymentsByOrder(orderID uint) ([]model.PaymentResponse, error) {
	// Implementation would go here
	return nil, errors.New("not implemented")
}

func (s *orderService) UpdateShippingStatus(orderID uint, status model.ShippingStatus, userID uint, description, location, notes string) error {
	// Implementation would go here
	return errors.New("not implemented")
}

func (s *orderService) GetShippingHistory(orderID uint) ([]model.ShippingHistoryResponse, error) {
	// Implementation would go here
	return nil, errors.New("not implemented")
}

func (s *orderService) GetOrderStats() (*model.OrderStatsResponse, error) {
	// Implementation would go here
	return nil, errors.New("not implemented")
}

func (s *orderService) GetOrderStatsByUser(userID uint) (map[string]interface{}, error) {
	// Implementation would go here
	return nil, errors.New("not implemented")
}

func (s *orderService) GetRevenueStats() (map[string]interface{}, error) {
	// Implementation would go here
	return nil, errors.New("not implemented")
}

func (s *orderService) ConvertCartToOrder(cartID uint, req *model.OrderCreateRequest, userID uint) (*model.OrderResponse, error) {
	// Implementation would go here
	return nil, errors.New("not implemented")
}
