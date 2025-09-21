package service

import (
	"fmt"
	"go_app/internal/model"
	"go_app/pkg/logger"
	"time"
)

// NotificationServiceInterface defines the interface for notification service
type NotificationServiceInterface interface {
	CreateNotification(req *model.CreateNotificationRequest) (*model.NotificationResponse, error)
}

// EventService handles business events and triggers notifications
type EventService interface {
	// Order events
	OnOrderCreated(order *model.Order) error
	OnOrderStatusUpdated(order *model.Order, oldStatus, newStatus model.OrderStatus) error
	OnOrderShipped(order *model.Order, trackingNumber string) error
	OnOrderDelivered(order *model.Order) error
	OnOrderCancelled(order *model.Order, reason string) error

	// Payment events
	OnPaymentSuccess(order *model.Order, payment *model.Payment) error
	OnPaymentFailed(order *model.Order, payment *model.Payment, errorMsg string) error

	// Product events
	OnProductBackInStock(product *model.Product) error
	OnPriceDrop(product *model.Product, oldPrice, newPrice float64) error

	// Review events
	OnReviewCreated(review *model.Review) error
	OnReviewApproved(review *model.Review) error

	// Wishlist events
	OnWishlistItemOnSale(wishlistItem *model.WishlistItem, oldPrice, newPrice float64) error

	// Inventory events
	OnLowStockAlert(product *model.Product, currentStock, minStock int) error

	// Coupon events
	OnCouponExpiring(coupon *model.Coupon, daysLeft int) error

	// Point events
	OnPointsEarned(userID uint, points int64, source string) error
	OnPointsExpiring(userID uint, points int64, expiryDate time.Time) error
}

// eventService implements EventService
type eventService struct {
	notificationService interface{} // NotificationService interface
	userService         interface{} // Placeholder for UserService
	productService      interface{} // Placeholder for ProductService
}

// NewEventService creates a new EventService
func NewEventService(notificationService interface{}, userService interface{}, productService interface{}) EventService {
	return &eventService{
		notificationService: notificationService,
		userService:         userService,
		productService:      productService,
	}
}

// sendNotification is a helper method to send notifications
func (s *eventService) sendNotification(notification *model.CreateNotificationRequest) error {
	if notifService, ok := s.notificationService.(NotificationServiceInterface); ok {
		_, err := notifService.CreateNotification(notification)
		return err
	}
	return nil
}

// Order events

// OnOrderCreated handles order creation event
func (s *eventService) OnOrderCreated(order *model.Order) error {
	// Create order confirmation notification
	notification := &model.CreateNotificationRequest{
		UserID:   &order.UserID,
		Type:     model.NotificationTypeOrder,
		Priority: model.NotificationPriorityNormal,
		Channel:  model.NotificationChannelEmail,
		Title:    fmt.Sprintf("Order Confirmation - #%s", order.OrderNumber),
		Message:  fmt.Sprintf("Thank you for your order! Your order #%s has been placed successfully and is being processed.", order.OrderNumber),
		Data: map[string]interface{}{
			"order_id":       order.ID,
			"order_number":   order.OrderNumber,
			"total_amount":   order.TotalAmount,
			"item_count":     len(order.OrderItems),
			"order_date":     order.CreatedAt.Format("2006-01-02 15:04:05"),
			"payment_method": order.PaymentMethod,
		},
		ActionURL: fmt.Sprintf("/orders/%d", order.ID),
	}

	// Send email notification
	if err := s.sendNotification(notification); err != nil {
		logger.Errorf("Failed to create order confirmation notification: %v", err)
		return err
	}

	// Also send in-app notification
	notification.Channel = model.NotificationChannelInApp
	notification.Title = "Order Placed"
	notification.Message = fmt.Sprintf("Your order #%s has been placed successfully", order.OrderNumber)

	if err := s.sendNotification(notification); err != nil {
		logger.Errorf("Failed to create in-app order notification: %v", err)
	}

	logger.Infof("Order confirmation notification sent for order #%s", order.OrderNumber)
	return nil
}

// OnOrderStatusUpdated handles order status update event
func (s *eventService) OnOrderStatusUpdated(order *model.Order, oldStatus, newStatus model.OrderStatus) error {
	var title, message string
	var priority model.NotificationPriority = model.NotificationPriorityNormal

	switch newStatus {
	case model.OrderStatusConfirmed:
		title = fmt.Sprintf("Order Confirmed - #%s", order.OrderNumber)
		message = fmt.Sprintf("Your order #%s has been confirmed and is being prepared for shipment.", order.OrderNumber)

	case model.OrderStatusProcessing:
		title = fmt.Sprintf("Order Processing - #%s", order.OrderNumber)
		message = fmt.Sprintf("Your order #%s is being processed and will be shipped soon.", order.OrderNumber)

	case model.OrderStatusShipped:
		title = fmt.Sprintf("Order Shipped - #%s", order.OrderNumber)
		message = fmt.Sprintf("Great news! Your order #%s has been shipped and is on its way to you.", order.OrderNumber)
		priority = model.NotificationPriorityHigh

	case model.OrderStatusDelivered:
		title = fmt.Sprintf("Order Delivered - #%s", order.OrderNumber)
		message = fmt.Sprintf("Your order #%s has been delivered successfully. Thank you for your purchase!", order.OrderNumber)
		priority = model.NotificationPriorityHigh

	case model.OrderStatusCancelled:
		title = fmt.Sprintf("Order Cancelled - #%s", order.OrderNumber)
		message = fmt.Sprintf("Your order #%s has been cancelled. If you have any questions, please contact our support team.", order.OrderNumber)
		priority = model.NotificationPriorityHigh

	case model.OrderStatusReturned:
		title = fmt.Sprintf("Order Returned - #%s", order.OrderNumber)
		message = fmt.Sprintf("Your return for order #%s has been processed. We will process your refund shortly.", order.OrderNumber)

	case model.OrderStatusRefunded:
		title = fmt.Sprintf("Order Refunded - #%s", order.OrderNumber)
		message = fmt.Sprintf("Your refund for order #%s has been processed and will appear in your account within 3-5 business days.", order.OrderNumber)

	default:
		// Generic status update
		title = fmt.Sprintf("Order Update - #%s", order.OrderNumber)
		message = fmt.Sprintf("Your order #%s status has been updated to %s.", order.OrderNumber, string(newStatus))
	}

	notification := &model.CreateNotificationRequest{
		UserID:   &order.UserID,
		Type:     model.NotificationTypeOrder,
		Priority: priority,
		Channel:  model.NotificationChannelEmail,
		Title:    title,
		Message:  message,
		Data: map[string]interface{}{
			"order_id":     order.ID,
			"order_number": order.OrderNumber,
			"old_status":   string(oldStatus),
			"new_status":   string(newStatus),
			"total_amount": order.TotalAmount,
			"updated_at":   time.Now().Format("2006-01-02 15:04:05"),
		},
		ActionURL: fmt.Sprintf("/orders/%d", order.ID),
	}

	if err := s.sendNotification(notification); err != nil {
		logger.Errorf("Failed to create order status notification: %v", err)
		return err
	}

	// Also send in-app notification for important status changes
	if newStatus == model.OrderStatusShipped || newStatus == model.OrderStatusDelivered || newStatus == model.OrderStatusCancelled {
		notification.Channel = model.NotificationChannelInApp
		if err := s.sendNotification(notification); err != nil {
			logger.Errorf("Failed to create in-app order status notification: %v", err)
		}
	}

	logger.Infof("Order status notification sent for order #%s: %s -> %s", order.OrderNumber, oldStatus, newStatus)
	return nil
}

// OnOrderShipped handles order shipped event
func (s *eventService) OnOrderShipped(order *model.Order, trackingNumber string) error {
	notification := &model.CreateNotificationRequest{
		UserID:   &order.UserID,
		Type:     model.NotificationTypeShipping,
		Priority: model.NotificationPriorityHigh,
		Channel:  model.NotificationChannelEmail,
		Title:    fmt.Sprintf("Order Shipped - #%s", order.OrderNumber),
		Message:  fmt.Sprintf("Your order #%s has been shipped! Tracking number: %s. You can track your package using the link below.", order.OrderNumber, trackingNumber),
		Data: map[string]interface{}{
			"order_id":           order.ID,
			"order_number":       order.OrderNumber,
			"tracking_number":    trackingNumber,
			"shipped_at":         time.Now().Format("2006-01-02 15:04:05"),
			"estimated_delivery": "3-5 business days",
		},
		ActionURL: fmt.Sprintf("/orders/%d/tracking", order.ID),
	}

	if err := s.sendNotification(notification); err != nil {
		logger.Errorf("Failed to create order shipped notification: %v", err)
		return err
	}

	// Also send SMS for important shipping updates
	notification.Channel = model.NotificationChannelSMS
	notification.Message = fmt.Sprintf("Your order #%s has been shipped! Track: %s", order.OrderNumber, trackingNumber)

	if err := s.sendNotification(notification); err != nil {
		logger.Errorf("Failed to create SMS shipping notification: %v", err)
	}

	logger.Infof("Order shipped notification sent for order #%s with tracking %s", order.OrderNumber, trackingNumber)
	return nil
}

// OnOrderDelivered handles order delivered event
func (s *eventService) OnOrderDelivered(order *model.Order) error {
	notification := &model.CreateNotificationRequest{
		UserID:   &order.UserID,
		Type:     model.NotificationTypeShipping,
		Priority: model.NotificationPriorityHigh,
		Channel:  model.NotificationChannelEmail,
		Title:    fmt.Sprintf("Order Delivered - #%s", order.OrderNumber),
		Message:  fmt.Sprintf("Your order #%s has been delivered successfully! We hope you enjoy your purchase. Please consider leaving a review.", order.OrderNumber),
		Data: map[string]interface{}{
			"order_id":     order.ID,
			"order_number": order.OrderNumber,
			"delivered_at": time.Now().Format("2006-01-02 15:04:05"),
			"review_url":   fmt.Sprintf("/orders/%d/review", order.ID),
		},
		ActionURL: fmt.Sprintf("/orders/%d", order.ID),
	}

	if err := s.sendNotification(notification); err != nil {
		logger.Errorf("Failed to create order delivered notification: %v", err)
		return err
	}

	// Also send in-app notification
	notification.Channel = model.NotificationChannelInApp
	notification.Title = "Order Delivered"
	notification.Message = fmt.Sprintf("Your order #%s has been delivered!", order.OrderNumber)

	if err := s.sendNotification(notification); err != nil {
		logger.Errorf("Failed to create in-app delivery notification: %v", err)
	}

	logger.Infof("Order delivered notification sent for order #%s", order.OrderNumber)
	return nil
}

// OnOrderCancelled handles order cancelled event
func (s *eventService) OnOrderCancelled(order *model.Order, reason string) error {
	notification := &model.CreateNotificationRequest{
		UserID:   &order.UserID,
		Type:     model.NotificationTypeOrder,
		Priority: model.NotificationPriorityHigh,
		Channel:  model.NotificationChannelEmail,
		Title:    fmt.Sprintf("Order Cancelled - #%s", order.OrderNumber),
		Message:  fmt.Sprintf("Your order #%s has been cancelled. Reason: %s. If you have any questions, please contact our support team.", order.OrderNumber, reason),
		Data: map[string]interface{}{
			"order_id":     order.ID,
			"order_number": order.OrderNumber,
			"cancelled_at": time.Now().Format("2006-01-02 15:04:05"),
			"reason":       reason,
			"refund_info":  "Refund will be processed within 3-5 business days",
		},
		ActionURL: fmt.Sprintf("/orders/%d", order.ID),
	}

	if err := s.sendNotification(notification); err != nil {
		logger.Errorf("Failed to create order cancelled notification: %v", err)
		return err
	}

	logger.Infof("Order cancelled notification sent for order #%s", order.OrderNumber)
	return nil
}

// Payment events

// OnPaymentSuccess handles payment success event
func (s *eventService) OnPaymentSuccess(order *model.Order, payment *model.Payment) error {
	notification := &model.CreateNotificationRequest{
		UserID:   &order.UserID,
		Type:     model.NotificationTypePayment,
		Priority: model.NotificationPriorityNormal,
		Channel:  model.NotificationChannelEmail,
		Title:    fmt.Sprintf("Payment Successful - #%s", order.OrderNumber),
		Message:  fmt.Sprintf("Your payment for order #%s has been processed successfully. Amount: $%.2f", order.OrderNumber, payment.Amount),
		Data: map[string]interface{}{
			"order_id":       order.ID,
			"order_number":   order.OrderNumber,
			"payment_id":     payment.ID,
			"amount":         payment.Amount,
			"payment_method": payment.PaymentMethod,
			"paid_at":        time.Now().Format("2006-01-02 15:04:05"),
		},
		ActionURL: fmt.Sprintf("/orders/%d", order.ID),
	}

	if err := s.sendNotification(notification); err != nil {
		logger.Errorf("Failed to create payment success notification: %v", err)
		return err
	}

	logger.Infof("Payment success notification sent for order #%s", order.OrderNumber)
	return nil
}

// OnPaymentFailed handles payment failed event
func (s *eventService) OnPaymentFailed(order *model.Order, payment *model.Payment, errorMsg string) error {
	notification := &model.CreateNotificationRequest{
		UserID:   &order.UserID,
		Type:     model.NotificationTypePayment,
		Priority: model.NotificationPriorityHigh,
		Channel:  model.NotificationChannelEmail,
		Title:    fmt.Sprintf("Payment Failed - #%s", order.OrderNumber),
		Message:  fmt.Sprintf("Your payment for order #%s has failed. Please try again or use a different payment method. Error: %s", order.OrderNumber, errorMsg),
		Data: map[string]interface{}{
			"order_id":       order.ID,
			"order_number":   order.OrderNumber,
			"payment_id":     payment.ID,
			"amount":         payment.Amount,
			"payment_method": payment.PaymentMethod,
			"error_message":  errorMsg,
			"failed_at":      time.Now().Format("2006-01-02 15:04:05"),
		},
		ActionURL: fmt.Sprintf("/orders/%d/payment", order.ID),
	}

	if err := s.sendNotification(notification); err != nil {
		logger.Errorf("Failed to create payment failed notification: %v", err)
		return err
	}

	logger.Infof("Payment failed notification sent for order #%s", order.OrderNumber)
	return nil
}

// Product events

// OnProductBackInStock handles product back in stock event
func (s *eventService) OnProductBackInStock(product *model.Product) error {
	// Get users who have this product in their wishlist
	// This would require a wishlist service to get users who want this product

	notification := &model.CreateNotificationRequest{
		UserID:   nil, // Will be set for each user
		Type:     model.NotificationTypeProduct,
		Priority: model.NotificationPriorityNormal,
		Channel:  model.NotificationChannelEmail,
		Title:    fmt.Sprintf("Back in Stock - %s", product.Name),
		Message:  fmt.Sprintf("Good news! %s is back in stock and available for purchase.", product.Name),
		Data: map[string]interface{}{
			"product_id":     product.ID,
			"product_name":   product.Name,
			"product_url":    fmt.Sprintf("/products/%d", product.ID),
			"current_price":  product.RegularPrice,
			"stock_quantity": product.StockQuantity,
		},
		ActionURL: fmt.Sprintf("/products/%d", product.ID),
	}

	// TODO: Send to all users who have this product in wishlist
	// For now, just log the event
	logger.Infof("Product back in stock notification prepared for product %s", product.Name)
	_ = notification // Avoid unused variable error
	return nil
}

// OnPriceDrop handles price drop event
func (s *eventService) OnPriceDrop(product *model.Product, oldPrice, newPrice float64) error {
	notification := &model.CreateNotificationRequest{
		UserID:   nil, // Will be set for each user
		Type:     model.NotificationTypeProduct,
		Priority: model.NotificationPriorityHigh,
		Channel:  model.NotificationChannelEmail,
		Title:    fmt.Sprintf("Price Drop Alert - %s", product.Name),
		Message:  fmt.Sprintf("The price of %s has dropped from $%.2f to $%.2f! Save $%.2f now.", product.Name, oldPrice, newPrice, oldPrice-newPrice),
		Data: map[string]interface{}{
			"product_id":          product.ID,
			"product_name":        product.Name,
			"old_price":           oldPrice,
			"new_price":           newPrice,
			"discount_amount":     oldPrice - newPrice,
			"discount_percentage": ((oldPrice - newPrice) / oldPrice) * 100,
			"product_url":         fmt.Sprintf("/products/%d", product.ID),
		},
		ActionURL: fmt.Sprintf("/products/%d", product.ID),
	}

	// TODO: Send to all users who have this product in wishlist
	logger.Infof("Price drop notification prepared for product %s: $%.2f -> $%.2f", product.Name, oldPrice, newPrice)
	_ = notification // Avoid unused variable error
	return nil
}

// Review events

// OnReviewCreated handles review creation event
func (s *eventService) OnReviewCreated(review *model.Review) error {
	// Notify admin about new review for moderation
	notification := &model.CreateNotificationRequest{
		UserID:   nil, // Admin notification
		Type:     model.NotificationTypeReview,
		Priority: model.NotificationPriorityNormal,
		Channel:  model.NotificationChannelInApp,
		Title:    "New Review Submitted",
		Message:  fmt.Sprintf("A new review has been submitted for product ID %d. Please review and approve.", review.ProductID),
		Data: map[string]interface{}{
			"review_id":    review.ID,
			"product_id":   review.ProductID,
			"user_id":      review.UserID,
			"rating":       review.Rating,
			"submitted_at": time.Now().Format("2006-01-02 15:04:05"),
		},
		ActionURL: fmt.Sprintf("/admin/reviews/%d", review.ID),
	}

	// TODO: Send to admin users
	logger.Infof("Review created notification prepared for review ID %d", review.ID)
	_ = notification // Avoid unused variable error
	return nil
}

// OnReviewApproved handles review approval event
func (s *eventService) OnReviewApproved(review *model.Review) error {
	notification := &model.CreateNotificationRequest{
		UserID:   &review.UserID,
		Type:     model.NotificationTypeReview,
		Priority: model.NotificationPriorityNormal,
		Channel:  model.NotificationChannelInApp,
		Title:    "Review Approved",
		Message:  "Your review has been approved and is now visible to other customers. Thank you for your feedback!",
		Data: map[string]interface{}{
			"review_id":   review.ID,
			"product_id":  review.ProductID,
			"approved_at": time.Now().Format("2006-01-02 15:04:05"),
		},
		ActionURL: fmt.Sprintf("/products/%d", review.ProductID),
	}

	if err := s.sendNotification(notification); err != nil {
		logger.Errorf("Failed to create review approved notification: %v", err)
		return err
	}

	logger.Infof("Review approved notification sent for review ID %d", review.ID)
	return nil
}

// Wishlist events

// OnWishlistItemOnSale handles wishlist item on sale event
func (s *eventService) OnWishlistItemOnSale(wishlistItem *model.WishlistItem, oldPrice, newPrice float64) error {
	notification := &model.CreateNotificationRequest{
		UserID:   &wishlistItem.Wishlist.UserID, // TODO: Fix this field access
		Type:     model.NotificationTypeWishlist,
		Priority: model.NotificationPriorityHigh,
		Channel:  model.NotificationChannelEmail,
		Title:    "Wishlist Item on Sale!",
		Message:  fmt.Sprintf("An item from your wishlist is now on sale! Save $%.2f on %s.", oldPrice-newPrice, wishlistItem.Product.Name),
		Data: map[string]interface{}{
			"wishlist_item_id": wishlistItem.ID,
			"product_id":       wishlistItem.ProductID,
			"product_name":     wishlistItem.Product.Name,
			"old_price":        oldPrice,
			"new_price":        newPrice,
			"discount_amount":  oldPrice - newPrice,
			"product_url":      fmt.Sprintf("/products/%d", wishlistItem.ProductID),
		},
		ActionURL: fmt.Sprintf("/products/%d", wishlistItem.ProductID),
	}

	if err := s.sendNotification(notification); err != nil {
		logger.Errorf("Failed to create wishlist sale notification: %v", err)
		return err
	}

	logger.Infof("Wishlist sale notification sent for product %s", wishlistItem.Product.Name)
	return nil
}

// Inventory events

// OnLowStockAlert handles low stock alert event
func (s *eventService) OnLowStockAlert(product *model.Product, currentStock, minStock int) error {
	// Notify admin about low stock
	notification := &model.CreateNotificationRequest{
		UserID:   nil, // Admin notification
		Type:     model.NotificationTypeInventory,
		Priority: model.NotificationPriorityHigh,
		Channel:  model.NotificationChannelInApp,
		Title:    "Low Stock Alert",
		Message:  fmt.Sprintf("Product %s is running low on stock. Current: %d, Minimum: %d", product.Name, currentStock, minStock),
		Data: map[string]interface{}{
			"product_id":         product.ID,
			"product_name":       product.Name,
			"current_stock":      currentStock,
			"minimum_stock":      minStock,
			"alert_triggered_at": time.Now().Format("2006-01-02 15:04:05"),
		},
		ActionURL: fmt.Sprintf("/admin/products/%d", product.ID),
	}

	// TODO: Send to admin users
	logger.Infof("Low stock alert notification prepared for product %s", product.Name)
	_ = notification // Avoid unused variable error
	return nil
}

// Coupon events

// OnCouponExpiring handles coupon expiring event
func (s *eventService) OnCouponExpiring(coupon *model.Coupon, daysLeft int) error {
	// Get users who have used this coupon before
	// This would require a coupon usage service

	notification := &model.CreateNotificationRequest{
		UserID:   nil, // Will be set for each user
		Type:     model.NotificationTypeCoupon,
		Priority: model.NotificationPriorityNormal,
		Channel:  model.NotificationChannelEmail,
		Title:    fmt.Sprintf("Coupon Expiring Soon - %s", coupon.Code),
		Message:  fmt.Sprintf("Your coupon %s expires in %d days. Use it before it's too late!", coupon.Code, daysLeft),
		Data: map[string]interface{}{
			"coupon_id":      coupon.ID,
			"coupon_code":    coupon.Code,
			"discount_type":  "percentage", // TODO: Fix coupon field access
			"discount_value": coupon.DiscountValue,
			"expires_at":     time.Now().AddDate(0, 0, daysLeft).Format("2006-01-02 15:04:05"), // TODO: Fix coupon field access
			"days_left":      daysLeft,
		},
		ActionURL: "/coupons",
	}

	// TODO: Send to users who have used this coupon
	logger.Infof("Coupon expiring notification prepared for coupon %s", coupon.Code)
	_ = notification // Avoid unused variable error
	return nil
}

// Point events

// OnPointsEarned handles points earned event
func (s *eventService) OnPointsEarned(userID uint, points int64, source string) error {
	notification := &model.CreateNotificationRequest{
		UserID:   &userID,
		Type:     model.NotificationTypePoint,
		Priority: model.NotificationPriorityNormal,
		Channel:  model.NotificationChannelInApp,
		Title:    "Points Earned!",
		Message:  fmt.Sprintf("You've earned %d points! Source: %s", points, source),
		Data: map[string]interface{}{
			"points_earned": points,
			"source":        source,
			"earned_at":     time.Now().Format("2006-01-02 15:04:05"),
		},
		ActionURL: "/points",
	}

	if err := s.sendNotification(notification); err != nil {
		logger.Errorf("Failed to create points earned notification: %v", err)
		return err
	}

	logger.Infof("Points earned notification sent for user %d", userID)
	return nil
}

// OnPointsExpiring handles points expiring event
func (s *eventService) OnPointsExpiring(userID uint, points int64, expiryDate time.Time) error {
	notification := &model.CreateNotificationRequest{
		UserID:   &userID,
		Type:     model.NotificationTypePoint,
		Priority: model.NotificationPriorityHigh,
		Channel:  model.NotificationChannelEmail,
		Title:    "Points Expiring Soon!",
		Message:  fmt.Sprintf("You have %d points expiring on %s. Use them before they expire!", points, expiryDate.Format("2006-01-02")),
		Data: map[string]interface{}{
			"points_expiring": points,
			"expiry_date":     expiryDate.Format("2006-01-02 15:04:05"),
			"days_left":       int(time.Until(expiryDate).Hours() / 24),
		},
		ActionURL: "/points",
	}

	if err := s.sendNotification(notification); err != nil {
		logger.Errorf("Failed to create points expiring notification: %v", err)
		return err
	}

	logger.Infof("Points expiring notification sent for user %d", userID)
	return nil
}
