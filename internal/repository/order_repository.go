package repository

import (
	"fmt"
	"go_app/internal/model"
	"go_app/pkg/database"
	"time"

	"gorm.io/gorm"
)

// OrderRepository defines methods for interacting with order data
type OrderRepository interface {
	// Orders
	CreateOrder(order *model.Order) error
	GetOrderByID(id uint) (*model.Order, error)
	GetOrderByOrderNumber(orderNumber string) (*model.Order, error)
	GetAllOrders(page, limit int, filters map[string]interface{}) ([]model.Order, int64, error)
	GetOrdersByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.Order, int64, error)
	UpdateOrder(order *model.Order) error
	DeleteOrder(id uint) error
	GetOrdersByStatus(status model.OrderStatus, page, limit int) ([]model.Order, int64, error)
	GetOrdersByDateRange(startDate, endDate time.Time, page, limit int) ([]model.Order, int64, error)

	// Order Items
	CreateOrderItem(orderItem *model.OrderItem) error
	GetOrderItemsByOrder(orderID uint) ([]model.OrderItem, error)
	UpdateOrderItem(orderItem *model.OrderItem) error
	DeleteOrderItem(id uint) error

	// Cart
	CreateCart(cart *model.Cart) error
	GetCartByID(id uint) (*model.Cart, error)
	GetCartByUser(userID uint) (*model.Cart, error)
	GetCartBySession(sessionID string) (*model.Cart, error)
	UpdateCart(cart *model.Cart) error
	DeleteCart(id uint) error
	ClearCart(cartID uint) error

	// Cart Items
	CreateCartItem(cartItem *model.CartItem) error
	GetCartItemsByCart(cartID uint) ([]model.CartItem, error)
	UpdateCartItem(cartItem *model.CartItem) error
	DeleteCartItem(id uint) error
	DeleteCartItemByProduct(cartID, productID uint, variantID *uint) error

	// Payments
	CreatePayment(payment *model.Payment) error
	GetPaymentByID(id uint) (*model.Payment, error)
	GetPaymentsByOrder(orderID uint) ([]model.Payment, error)
	GetPaymentsByUser(userID uint, page, limit int) ([]model.Payment, int64, error)
	UpdatePayment(payment *model.Payment) error
	DeletePayment(id uint) error

	// Shipping History
	CreateShippingHistory(history *model.ShippingHistory) error
	GetShippingHistoryByOrder(orderID uint) ([]model.ShippingHistory, error)
	UpdateShippingHistory(history *model.ShippingHistory) error

	// Statistics
	GetOrderStats() (*model.OrderStatsResponse, error)
	GetOrderStatsByUser(userID uint) (map[string]interface{}, error)
	GetOrderStatsByDateRange(startDate, endDate time.Time) (map[string]interface{}, error)
	GetRevenueStats() (map[string]interface{}, error)
}

// orderRepository implements OrderRepository
type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new OrderRepository
func NewOrderRepository() OrderRepository {
	return &orderRepository{
		db: database.DB,
	}
}

// Orders

// CreateOrder creates a new order
func (r *orderRepository) CreateOrder(order *model.Order) error {
	return r.db.Create(order).Error
}

// GetOrderByID retrieves an order by its ID
func (r *orderRepository) GetOrderByID(id uint) (*model.Order, error) {
	var order model.Order
	if err := r.db.Preload("User").
		Preload("OrderItems.Product").
		Preload("OrderItems.ProductVariant").
		Preload("Payments").
		Preload("ShippingHistory.UpdatedByUser").
		First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// GetOrderByOrderNumber retrieves an order by its order number
func (r *orderRepository) GetOrderByOrderNumber(orderNumber string) (*model.Order, error) {
	var order model.Order
	if err := r.db.Where("order_number = ?", orderNumber).
		Preload("User").
		Preload("OrderItems.Product").
		Preload("OrderItems.ProductVariant").
		Preload("Payments").
		Preload("ShippingHistory.UpdatedByUser").
		First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// GetAllOrders retrieves all orders with pagination and filters
func (r *orderRepository) GetAllOrders(page, limit int, filters map[string]interface{}) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	db := r.db.Model(&model.Order{})

	// Apply filters
	for key, value := range filters {
		switch key {
		case "status":
			db = db.Where("status = ?", value)
		case "payment_status":
			db = db.Where("payment_status = ?", value)
		case "shipping_status":
			db = db.Where("shipping_status = ?", value)
		case "payment_method":
			db = db.Where("payment_method = ?", value)
		case "user_id":
			db = db.Where("user_id = ?", value)
		case "customer_email":
			db = db.Where("customer_email = ?", value)
		case "customer_phone":
			db = db.Where("customer_phone = ?", value)
		case "order_number":
			db = db.Where("order_number LIKE ?", fmt.Sprintf("%%%s%%", value.(string)))
		case "date_from":
			db = db.Where("created_at >= ?", value)
		case "date_to":
			db = db.Where("created_at <= ?", value)
		case "min_amount":
			db = db.Where("total_amount >= ?", value)
		case "max_amount":
			db = db.Where("total_amount <= ?", value)
		case "search":
			searchTerm := fmt.Sprintf("%%%s%%", value.(string))
			db = db.Where("order_number LIKE ? OR customer_name LIKE ? OR customer_email LIKE ?",
				searchTerm, searchTerm, searchTerm)
		}
	}

	// Count total records
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting
	db = db.Order("created_at DESC")

	if err := db.Preload("User").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// GetOrdersByUser retrieves orders for a specific user
func (r *orderRepository) GetOrdersByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.Order, int64, error) {
	filters["user_id"] = userID
	return r.GetAllOrders(page, limit, filters)
}

// UpdateOrder updates an existing order
func (r *orderRepository) UpdateOrder(order *model.Order) error {
	return r.db.Save(order).Error
}

// DeleteOrder soft deletes an order
func (r *orderRepository) DeleteOrder(id uint) error {
	return r.db.Delete(&model.Order{}, id).Error
}

// GetOrdersByStatus retrieves orders by status
func (r *orderRepository) GetOrdersByStatus(status model.OrderStatus, page, limit int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	db := r.db.Model(&model.Order{}).Where("status = ?", status)

	// Count total records
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting
	db = db.Order("created_at DESC")

	if err := db.Preload("User").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// GetOrdersByDateRange retrieves orders within a date range
func (r *orderRepository) GetOrdersByDateRange(startDate, endDate time.Time, page, limit int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	db := r.db.Model(&model.Order{}).Where("created_at BETWEEN ? AND ?", startDate, endDate)

	// Count total records
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting
	db = db.Order("created_at DESC")

	if err := db.Preload("User").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// Order Items

// CreateOrderItem creates a new order item
func (r *orderRepository) CreateOrderItem(orderItem *model.OrderItem) error {
	return r.db.Create(orderItem).Error
}

// GetOrderItemsByOrder retrieves order items for an order
func (r *orderRepository) GetOrderItemsByOrder(orderID uint) ([]model.OrderItem, error) {
	var orderItems []model.OrderItem
	err := r.db.Where("order_id = ?", orderID).
		Preload("Product").
		Preload("ProductVariant").
		Order("created_at ASC").
		Find(&orderItems).Error
	return orderItems, err
}

// UpdateOrderItem updates an existing order item
func (r *orderRepository) UpdateOrderItem(orderItem *model.OrderItem) error {
	return r.db.Save(orderItem).Error
}

// DeleteOrderItem deletes an order item
func (r *orderRepository) DeleteOrderItem(id uint) error {
	return r.db.Delete(&model.OrderItem{}, id).Error
}

// Cart

// CreateCart creates a new cart
func (r *orderRepository) CreateCart(cart *model.Cart) error {
	return r.db.Create(cart).Error
}

// GetCartByID retrieves a cart by its ID
func (r *orderRepository) GetCartByID(id uint) (*model.Cart, error) {
	var cart model.Cart
	if err := r.db.Preload("User").
		Preload("CartItems.Product").
		Preload("CartItems.ProductVariant").
		First(&cart, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cart, nil
}

// GetCartByUser retrieves a cart by user ID
func (r *orderRepository) GetCartByUser(userID uint) (*model.Cart, error) {
	var cart model.Cart
	if err := r.db.Where("user_id = ?", userID).
		Preload("User").
		Preload("CartItems.Product").
		Preload("CartItems.ProductVariant").
		First(&cart).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cart, nil
}

// GetCartBySession retrieves a cart by session ID
func (r *orderRepository) GetCartBySession(sessionID string) (*model.Cart, error) {
	var cart model.Cart
	if err := r.db.Where("session_id = ?", sessionID).
		Preload("User").
		Preload("CartItems.Product").
		Preload("CartItems.ProductVariant").
		First(&cart).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cart, nil
}

// UpdateCart updates an existing cart
func (r *orderRepository) UpdateCart(cart *model.Cart) error {
	return r.db.Save(cart).Error
}

// DeleteCart deletes a cart
func (r *orderRepository) DeleteCart(id uint) error {
	return r.db.Delete(&model.Cart{}, id).Error
}

// ClearCart clears all items from a cart
func (r *orderRepository) ClearCart(cartID uint) error {
	// Delete all cart items
	if err := r.db.Where("cart_id = ?", cartID).Delete(&model.CartItem{}).Error; err != nil {
		return err
	}

	// Reset cart totals
	cart := &model.Cart{ID: cartID}
	cart.ItemsCount = 0
	cart.ItemsQuantity = 0
	cart.SubTotal = 0
	cart.TaxAmount = 0
	cart.ShippingCost = 0
	cart.DiscountAmount = 0
	cart.TotalAmount = 0

	return r.db.Model(cart).Updates(map[string]interface{}{
		"items_count":     0,
		"items_quantity":  0,
		"sub_total":       0,
		"tax_amount":      0,
		"shipping_cost":   0,
		"discount_amount": 0,
		"total_amount":    0,
	}).Error
}

// Cart Items

// CreateCartItem creates a new cart item
func (r *orderRepository) CreateCartItem(cartItem *model.CartItem) error {
	return r.db.Create(cartItem).Error
}

// GetCartItemsByCart retrieves cart items for a cart
func (r *orderRepository) GetCartItemsByCart(cartID uint) ([]model.CartItem, error) {
	var cartItems []model.CartItem
	err := r.db.Where("cart_id = ?", cartID).
		Preload("Product").
		Preload("ProductVariant").
		Order("created_at ASC").
		Find(&cartItems).Error
	return cartItems, err
}

// UpdateCartItem updates an existing cart item
func (r *orderRepository) UpdateCartItem(cartItem *model.CartItem) error {
	return r.db.Save(cartItem).Error
}

// DeleteCartItem deletes a cart item
func (r *orderRepository) DeleteCartItem(id uint) error {
	return r.db.Delete(&model.CartItem{}, id).Error
}

// DeleteCartItemByProduct deletes a cart item by product and variant
func (r *orderRepository) DeleteCartItemByProduct(cartID, productID uint, variantID *uint) error {
	query := r.db.Where("cart_id = ? AND product_id = ?", cartID, productID)
	if variantID != nil {
		query = query.Where("product_variant_id = ?", *variantID)
	} else {
		query = query.Where("product_variant_id IS NULL")
	}
	return query.Delete(&model.CartItem{}).Error
}

// Payments

// CreatePayment creates a new payment
func (r *orderRepository) CreatePayment(payment *model.Payment) error {
	return r.db.Create(payment).Error
}

// GetPaymentByID retrieves a payment by its ID
func (r *orderRepository) GetPaymentByID(id uint) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.Preload("Order").
		Preload("User").
		First(&payment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

// GetPaymentsByOrder retrieves payments for an order
func (r *orderRepository) GetPaymentsByOrder(orderID uint) ([]model.Payment, error) {
	var payments []model.Payment
	err := r.db.Where("order_id = ?", orderID).
		Preload("Order").
		Preload("User").
		Order("created_at DESC").
		Find(&payments).Error
	return payments, err
}

// GetPaymentsByUser retrieves payments for a user
func (r *orderRepository) GetPaymentsByUser(userID uint, page, limit int) ([]model.Payment, int64, error) {
	var payments []model.Payment
	var total int64
	db := r.db.Model(&model.Payment{}).Where("user_id = ?", userID)

	// Count total records
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting
	db = db.Order("created_at DESC")

	if err := db.Preload("Order").Preload("User").Find(&payments).Error; err != nil {
		return nil, 0, err
	}

	return payments, total, nil
}

// UpdatePayment updates an existing payment
func (r *orderRepository) UpdatePayment(payment *model.Payment) error {
	return r.db.Save(payment).Error
}

// DeletePayment deletes a payment
func (r *orderRepository) DeletePayment(id uint) error {
	return r.db.Delete(&model.Payment{}, id).Error
}

// Shipping History

// CreateShippingHistory creates a new shipping history entry
func (r *orderRepository) CreateShippingHistory(history *model.ShippingHistory) error {
	return r.db.Create(history).Error
}

// GetShippingHistoryByOrder retrieves shipping history for an order
func (r *orderRepository) GetShippingHistoryByOrder(orderID uint) ([]model.ShippingHistory, error) {
	var history []model.ShippingHistory
	err := r.db.Where("order_id = ?", orderID).
		Preload("UpdatedByUser").
		Order("created_at ASC").
		Find(&history).Error
	return history, err
}

// UpdateShippingHistory updates an existing shipping history entry
func (r *orderRepository) UpdateShippingHistory(history *model.ShippingHistory) error {
	return r.db.Save(history).Error
}

// Statistics

// GetOrderStats retrieves order statistics
func (r *orderRepository) GetOrderStats() (*model.OrderStatsResponse, error) {
	var stats model.OrderStatsResponse
	var count int64

	// Total orders
	r.db.Model(&model.Order{}).Count(&count)
	stats.TotalOrders = count

	// Orders by status
	r.db.Model(&model.Order{}).Where("status = ?", model.OrderStatusPending).Count(&count)
	stats.PendingOrders = count

	r.db.Model(&model.Order{}).Where("status = ?", model.OrderStatusConfirmed).Count(&count)
	stats.ConfirmedOrders = count

	r.db.Model(&model.Order{}).Where("status = ?", model.OrderStatusShipped).Count(&count)
	stats.ShippedOrders = count

	r.db.Model(&model.Order{}).Where("status = ?", model.OrderStatusDelivered).Count(&count)
	stats.DeliveredOrders = count

	r.db.Model(&model.Order{}).Where("status = ?", model.OrderStatusCancelled).Count(&count)
	stats.CancelledOrders = count

	// Revenue statistics
	var totalRevenue float64
	r.db.Model(&model.Order{}).Where("status = ?", model.OrderStatusDelivered).Select("SUM(total_amount)").Scan(&totalRevenue)
	stats.TotalRevenue = totalRevenue

	// Average order value
	var avgOrderValue float64
	r.db.Model(&model.Order{}).Where("status = ?", model.OrderStatusDelivered).Select("AVG(total_amount)").Scan(&avgOrderValue)
	stats.AverageOrderValue = avgOrderValue

	// Conversion rate (simplified - would need more complex logic in real app)
	var totalCarts int64
	r.db.Model(&model.Cart{}).Count(&totalCarts)
	if totalCarts > 0 {
		stats.ConversionRate = float64(stats.DeliveredOrders) / float64(totalCarts) * 100
	}

	return &stats, nil
}

// GetOrderStatsByUser retrieves order statistics for a specific user
func (r *orderRepository) GetOrderStatsByUser(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var count int64
	var total float64

	// Total orders for user
	r.db.Model(&model.Order{}).Where("user_id = ?", userID).Count(&count)
	stats["total_orders"] = count

	// Orders by status
	r.db.Model(&model.Order{}).Where("user_id = ? AND status = ?", userID, model.OrderStatusPending).Count(&count)
	stats["pending_orders"] = count

	r.db.Model(&model.Order{}).Where("user_id = ? AND status = ?", userID, model.OrderStatusDelivered).Count(&count)
	stats["delivered_orders"] = count

	r.db.Model(&model.Order{}).Where("user_id = ? AND status = ?", userID, model.OrderStatusCancelled).Count(&count)
	stats["cancelled_orders"] = count

	// Total spent
	r.db.Model(&model.Order{}).Where("user_id = ? AND status = ?", userID, model.OrderStatusDelivered).Select("SUM(total_amount)").Scan(&total)
	stats["total_spent"] = total

	// Average order value
	var avgOrderValue float64
	r.db.Model(&model.Order{}).Where("user_id = ? AND status = ?", userID, model.OrderStatusDelivered).Select("AVG(total_amount)").Scan(&avgOrderValue)
	stats["average_order_value"] = avgOrderValue

	// Last order date
	var lastOrder model.Order
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").First(&lastOrder).Error; err == nil {
		stats["last_order_date"] = lastOrder.CreatedAt
	}

	return stats, nil
}

// GetOrderStatsByDateRange retrieves order statistics for a date range
func (r *orderRepository) GetOrderStatsByDateRange(startDate, endDate time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var count int64
	var total float64

	// Orders in date range
	r.db.Model(&model.Order{}).Where("created_at BETWEEN ? AND ?", startDate, endDate).Count(&count)
	stats["orders_count"] = count

	// Revenue in date range
	r.db.Model(&model.Order{}).Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, model.OrderStatusDelivered).Select("SUM(total_amount)").Scan(&total)
	stats["revenue"] = total

	// Orders by status in date range
	r.db.Model(&model.Order{}).Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, model.OrderStatusPending).Count(&count)
	stats["pending_orders"] = count

	r.db.Model(&model.Order{}).Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, model.OrderStatusDelivered).Count(&count)
	stats["delivered_orders"] = count

	r.db.Model(&model.Order{}).Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, model.OrderStatusCancelled).Count(&count)
	stats["cancelled_orders"] = count

	return stats, nil
}

// GetRevenueStats retrieves revenue statistics
func (r *orderRepository) GetRevenueStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var total float64
	var count int64

	// Total revenue
	r.db.Model(&model.Order{}).Where("status = ?", model.OrderStatusDelivered).Select("SUM(total_amount)").Scan(&total)
	stats["total_revenue"] = total

	// Total orders
	r.db.Model(&model.Order{}).Where("status = ?", model.OrderStatusDelivered).Count(&count)
	stats["total_orders"] = count

	// Average order value
	if count > 0 {
		stats["average_order_value"] = total / float64(count)
	} else {
		stats["average_order_value"] = 0
	}

	// Revenue by payment method
	var paymentStats []map[string]interface{}
	r.db.Model(&model.Order{}).
		Select("payment_method, SUM(total_amount) as revenue, COUNT(*) as count").
		Where("status = ?", model.OrderStatusDelivered).
		Group("payment_method").
		Scan(&paymentStats)
	stats["revenue_by_payment_method"] = paymentStats

	// Monthly revenue (last 12 months)
	var monthlyStats []map[string]interface{}
	r.db.Model(&model.Order{}).
		Select("DATE_FORMAT(created_at, '%Y-%m') as month, SUM(total_amount) as revenue, COUNT(*) as count").
		Where("status = ? AND created_at >= DATE_SUB(NOW(), INTERVAL 12 MONTH)", model.OrderStatusDelivered).
		Group("DATE_FORMAT(created_at, '%Y-%m')").
		Order("month ASC").
		Scan(&monthlyStats)
	stats["monthly_revenue"] = monthlyStats

	return stats, nil
}
