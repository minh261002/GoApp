package service

import (
	"encoding/json"
	"fmt"
	"time"

	"go_app/internal/model"
	"go_app/pkg/logger"
	"go_app/pkg/payment"
)

// PaymentGatewayService handles payment gateway integrations
type PaymentGatewayService interface {
	CreatePaymentLink(order *model.Order, paymentMethod model.PaymentMethod) (*model.PaymentLinkResponse, error)
	ProcessPayment(orderCode int, paymentMethod model.PaymentMethod) (*model.PaymentInfoResponse, error)
	CancelPayment(orderCode int, paymentMethod model.PaymentMethod, reason string) error
	VerifyWebhook(paymentMethod model.PaymentMethod, signature string, data []byte) bool
	HandleWebhook(paymentMethod model.PaymentMethod, webhookData []byte) (*model.PaymentWebhookResponse, error)
}

type paymentGatewayService struct {
	payOSClient *payment.PayOSClient
}

// NewPaymentGatewayService creates a new PaymentGatewayService
func NewPaymentGatewayService(payOSConfig payment.PayOSConfig) PaymentGatewayService {
	return &paymentGatewayService{
		payOSClient: payment.NewPayOSClient(payOSConfig),
	}
}

// CreatePaymentLink creates a payment link for the specified payment method
func (s *paymentGatewayService) CreatePaymentLink(order *model.Order, paymentMethod model.PaymentMethod) (*model.PaymentLinkResponse, error) {
	switch paymentMethod {
	case model.PaymentMethodVietQR:
		return s.createVietQRPaymentLink(order)
	case model.PaymentMethodCOD:
		return s.createCODPaymentLink(order)
	default:
		return nil, fmt.Errorf("unsupported payment method: %s", paymentMethod)
	}
}

// ProcessPayment processes a payment for the specified payment method
func (s *paymentGatewayService) ProcessPayment(orderCode int, paymentMethod model.PaymentMethod) (*model.PaymentInfoResponse, error) {
	switch paymentMethod {
	case model.PaymentMethodVietQR:
		return s.processVietQRPayment(orderCode)
	case model.PaymentMethodCOD:
		return s.processCODPayment(orderCode)
	default:
		return nil, fmt.Errorf("unsupported payment method: %s", paymentMethod)
	}
}

// CancelPayment cancels a payment for the specified payment method
func (s *paymentGatewayService) CancelPayment(orderCode int, paymentMethod model.PaymentMethod, reason string) error {
	switch paymentMethod {
	case model.PaymentMethodVietQR:
		return s.cancelVietQRPayment(orderCode, reason)
	case model.PaymentMethodCOD:
		return s.cancelCODPayment(orderCode, reason)
	default:
		return fmt.Errorf("unsupported payment method: %s", paymentMethod)
	}
}

// VerifyWebhook verifies webhook signature for the specified payment method
func (s *paymentGatewayService) VerifyWebhook(paymentMethod model.PaymentMethod, signature string, data []byte) bool {
	switch paymentMethod {
	case model.PaymentMethodVietQR:
		return s.payOSClient.VerifyWebhookSignature(signature, data)
	case model.PaymentMethodCOD:
		return true // COD doesn't need webhook verification
	default:
		return false
	}
}

// HandleWebhook handles webhook data for the specified payment method
func (s *paymentGatewayService) HandleWebhook(paymentMethod model.PaymentMethod, webhookData []byte) (*model.PaymentWebhookResponse, error) {
	switch paymentMethod {
	case model.PaymentMethodVietQR:
		return s.handleVietQRWebhook(webhookData)
	case model.PaymentMethodCOD:
		return s.handleCODWebhook(webhookData)
	default:
		return nil, fmt.Errorf("unsupported payment method: %s", paymentMethod)
	}
}

// VietQR (PayOS) specific methods

func (s *paymentGatewayService) createVietQRPaymentLink(order *model.Order) (*model.PaymentLinkResponse, error) {
	// Convert order items to PayOS items
	var items []payment.PayOSItem
	for _, item := range order.OrderItems {
		items = append(items, payment.PayOSItem{
			Name:     item.ProductName,
			Quantity: int(item.Quantity),
			Price:    payment.ConvertVNDToInt(item.UnitPrice),
		})
	}

	// Create PayOS payment data
	payOSData := payment.PayOSPaymentData{
		OrderCode:   payment.GenerateOrderCode(),
		Amount:      payment.ConvertVNDToInt(order.TotalAmount),
		Description: fmt.Sprintf("Thanh toán đơn hàng #%s", order.OrderNumber),
		Items:       items,
		ReturnURL:   fmt.Sprintf("https://your-domain.com/payment/success?order_id=%d", order.ID),
		CancelURL:   fmt.Sprintf("https://your-domain.com/payment/cancel?order_id=%d", order.ID),
		ExpiredAt:   func() *int64 { t := time.Now().Add(24 * time.Hour).Unix(); return &t }(),
	}

	// Create payment link
	response, err := s.payOSClient.CreatePaymentLink(payOSData)
	if err != nil {
		return nil, fmt.Errorf("failed to create VietQR payment link: %v", err)
	}

	// Convert to response format
	return &model.PaymentLinkResponse{
		PaymentURL:    response.Data.CheckoutURL,
		QRCode:        response.Data.QRCode,
		OrderCode:     response.Data.OrderCode,
		Amount:        payment.ConvertIntToVND(response.Data.Amount),
		AccountNumber: response.Data.AccountNumber,
		AccountName:   response.Data.AccountName,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
		PaymentMethod: model.PaymentMethodVietQR,
	}, nil
}

func (s *paymentGatewayService) processVietQRPayment(orderCode int) (*model.PaymentInfoResponse, error) {
	paymentInfo, err := s.payOSClient.GetPaymentInfo(orderCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get VietQR payment info: %v", err)
	}

	return &model.PaymentInfoResponse{
		OrderCode:     paymentInfo.OrderCode,
		Amount:        payment.ConvertIntToVND(paymentInfo.Amount),
		Status:        s.mapPayOSStatusToPaymentStatus(paymentInfo.Code),
		TransactionID: paymentInfo.TransactionID,
		Reference:     paymentInfo.Reference,
		AccountNumber: paymentInfo.AccountNumber,
		Description:   paymentInfo.Description,
		PaymentMethod: model.PaymentMethodVietQR,
	}, nil
}

func (s *paymentGatewayService) cancelVietQRPayment(orderCode int, reason string) error {
	return s.payOSClient.CancelPayment(orderCode, reason)
}

func (s *paymentGatewayService) handleVietQRWebhook(webhookData []byte) (*model.PaymentWebhookResponse, error) {
	var webhook payment.PayOSWebhookData
	if err := json.Unmarshal(webhookData, &webhook); err != nil {
		return nil, fmt.Errorf("failed to unmarshal VietQR webhook data: %v", err)
	}

	return &model.PaymentWebhookResponse{
		OrderCode:     webhook.Data.OrderCode,
		Amount:        payment.ConvertIntToVND(webhook.Data.Amount),
		Status:        s.mapPayOSStatusToPaymentStatus(webhook.Data.Code),
		TransactionID: webhook.Data.TransactionID,
		Reference:     webhook.Data.Reference,
		PaymentMethod: model.PaymentMethodVietQR,
		RawData:       webhookData,
	}, nil
}

// Helper methods

func (s *paymentGatewayService) mapPayOSStatusToPaymentStatus(code string) model.PaymentStatus {
	switch code {
	case "00": // Success
		return model.PaymentStatusPaid
	case "01": // Pending
		return model.PaymentStatusPending
	case "02": // Failed
		return model.PaymentStatusFailed
	case "03": // Cancelled
		return model.PaymentStatusCancelled
	default:
		return model.PaymentStatusPending
	}
}

// COD (Cash on Delivery) specific methods

func (s *paymentGatewayService) createCODPaymentLink(order *model.Order) (*model.PaymentLinkResponse, error) {
	// COD doesn't need a payment link, just return order info
	return &model.PaymentLinkResponse{
		PaymentURL:    "", // No payment URL for COD
		QRCode:        "", // No QR code for COD
		OrderCode:     int(order.ID),
		Amount:        order.TotalAmount,
		AccountNumber: "",
		AccountName:   "",
		ExpiresAt:     time.Now().Add(7 * 24 * time.Hour), // 7 days for COD
		PaymentMethod: model.PaymentMethodCOD,
	}, nil
}

func (s *paymentGatewayService) processCODPayment(orderCode int) (*model.PaymentInfoResponse, error) {
	// COD payment is always pending until delivery
	return &model.PaymentInfoResponse{
		OrderCode:     orderCode,
		Amount:        0, // Will be set when order is delivered
		Status:        model.PaymentStatusPending,
		TransactionID: fmt.Sprintf("COD-%d", orderCode),
		Reference:     fmt.Sprintf("COD-REF-%d", orderCode),
		AccountNumber: "",
		Description:   "Cash on Delivery - Payment pending",
		PaymentMethod: model.PaymentMethodCOD,
	}, nil
}

func (s *paymentGatewayService) cancelCODPayment(orderCode int, reason string) error {
	// COD can be cancelled easily
	logger.Infof("COD payment cancelled: OrderCode=%d, Reason=%s", orderCode, reason)
	return nil
}

func (s *paymentGatewayService) handleCODWebhook(webhookData []byte) (*model.PaymentWebhookResponse, error) {
	// COD doesn't have webhooks, but we can simulate it
	return &model.PaymentWebhookResponse{
		OrderCode:     0,
		Amount:        0,
		Status:        model.PaymentStatusPending,
		TransactionID: "",
		Reference:     "",
		PaymentMethod: model.PaymentMethodCOD,
		RawData:       webhookData,
	}, nil
}
