package service

import (
	"fmt"
	"go_app/internal/model"
	"go_app/internal/repository"
	"go_app/pkg/logger"
	"go_app/pkg/shipping"
	"time"
)

type ShippingService interface {
	// Shipping Providers
	CreateShippingProvider(req *model.ShippingProvider) (*model.ShippingProvider, error)
	GetShippingProviderByID(id uint) (*model.ShippingProvider, error)
	GetShippingProviderByCode(code string) (*model.ShippingProvider, error)
	GetAllShippingProviders() ([]model.ShippingProvider, error)
	GetActiveShippingProviders() ([]model.ShippingProvider, error)
	UpdateShippingProvider(id uint, req *model.ShippingProvider) (*model.ShippingProvider, error)
	DeleteShippingProvider(id uint) error

	// Shipping Rates
	CreateShippingRate(req *model.ShippingRate) (*model.ShippingRate, error)
	GetShippingRateByID(id uint) (*model.ShippingRate, error)
	GetShippingRatesByProvider(providerID uint) ([]model.ShippingRate, error)
	UpdateShippingRate(id uint, req *model.ShippingRate) (*model.ShippingRate, error)
	DeleteShippingRate(id uint) error

	// Shipping Calculation
	CalculateShipping(req *model.CalculateShippingRequest) ([]model.CalculateShippingResponse, error)
	CalculateShippingWithGHTK(req *model.CalculateShippingRequest) (*model.CalculateShippingResponse, error)

	// Shipping Orders
	CreateShippingOrder(req *model.ShippingOrderRequest) (*model.ShippingOrderResponse, error)
	GetShippingOrderByID(id uint) (*model.ShippingOrderResponse, error)
	GetShippingOrderByOrderID(orderID uint) (*model.ShippingOrderResponse, error)
	GetShippingOrderByTrackingCode(trackingCode string) (*model.ShippingOrderResponse, error)
	UpdateShippingOrder(id uint, req *model.ShippingOrder) (*model.ShippingOrderResponse, error)
	CancelShippingOrder(id uint, reason string) error
	GetShippingOrders(page, limit int, filters map[string]interface{}) ([]model.ShippingOrderResponse, int64, error)

	// Tracking
	GetShippingTracking(orderID uint) ([]model.ShippingTracking, error)
	UpdateShippingStatusFromWebhook(webhookData *model.WebhookData) error

	// Statistics
	GetShippingStats() (*model.ShippingStats, error)
	GetShippingStatsByProvider(providerID uint) (*model.ShippingStats, error)
}

type shippingService struct {
	shippingRepo repository.ShippingRepository
	orderRepo    repository.OrderRepository
	ghtkClient   *shipping.GHTKClient
}

func NewShippingService(shippingRepo repository.ShippingRepository, orderRepo repository.OrderRepository, ghtkConfig shipping.GHTKConfig) ShippingService {
	ghtkClient := shipping.NewGHTKClient(ghtkConfig)

	return &shippingService{
		shippingRepo: shippingRepo,
		orderRepo:    orderRepo,
		ghtkClient:   ghtkClient,
	}
}

// Shipping Providers
func (s *shippingService) CreateShippingProvider(req *model.ShippingProvider) (*model.ShippingProvider, error) {
	if err := s.shippingRepo.CreateShippingProvider(req); err != nil {
		logger.Errorf("Failed to create shipping provider: %v", err)
		return nil, fmt.Errorf("failed to create shipping provider")
	}
	return req, nil
}

func (s *shippingService) GetShippingProviderByID(id uint) (*model.ShippingProvider, error) {
	provider, err := s.shippingRepo.GetShippingProviderByID(id)
	if err != nil {
		logger.Errorf("Failed to get shipping provider by ID %d: %v", id, err)
		return nil, fmt.Errorf("shipping provider not found")
	}
	return provider, nil
}

func (s *shippingService) GetShippingProviderByCode(code string) (*model.ShippingProvider, error) {
	provider, err := s.shippingRepo.GetShippingProviderByCode(code)
	if err != nil {
		logger.Errorf("Failed to get shipping provider by code %s: %v", code, err)
		return nil, fmt.Errorf("shipping provider not found")
	}
	return provider, nil
}

func (s *shippingService) GetAllShippingProviders() ([]model.ShippingProvider, error) {
	providers, err := s.shippingRepo.GetAllShippingProviders()
	if err != nil {
		logger.Errorf("Failed to get all shipping providers: %v", err)
		return nil, fmt.Errorf("failed to get shipping providers")
	}
	return providers, nil
}

func (s *shippingService) GetActiveShippingProviders() ([]model.ShippingProvider, error) {
	providers, err := s.shippingRepo.GetActiveShippingProviders()
	if err != nil {
		logger.Errorf("Failed to get active shipping providers: %v", err)
		return nil, fmt.Errorf("failed to get active shipping providers")
	}
	return providers, nil
}

func (s *shippingService) UpdateShippingProvider(id uint, req *model.ShippingProvider) (*model.ShippingProvider, error) {
	req.ID = id
	if err := s.shippingRepo.UpdateShippingProvider(req); err != nil {
		logger.Errorf("Failed to update shipping provider %d: %v", id, err)
		return nil, fmt.Errorf("failed to update shipping provider")
	}
	return req, nil
}

func (s *shippingService) DeleteShippingProvider(id uint) error {
	if err := s.shippingRepo.DeleteShippingProvider(id); err != nil {
		logger.Errorf("Failed to delete shipping provider %d: %v", id, err)
		return fmt.Errorf("failed to delete shipping provider")
	}
	return nil
}

// Shipping Rates
func (s *shippingService) CreateShippingRate(req *model.ShippingRate) (*model.ShippingRate, error) {
	if err := s.shippingRepo.CreateShippingRate(req); err != nil {
		logger.Errorf("Failed to create shipping rate: %v", err)
		return nil, fmt.Errorf("failed to create shipping rate")
	}
	return req, nil
}

func (s *shippingService) GetShippingRateByID(id uint) (*model.ShippingRate, error) {
	rate, err := s.shippingRepo.GetShippingRateByID(id)
	if err != nil {
		logger.Errorf("Failed to get shipping rate by ID %d: %v", id, err)
		return nil, fmt.Errorf("shipping rate not found")
	}
	return rate, nil
}

func (s *shippingService) GetShippingRatesByProvider(providerID uint) ([]model.ShippingRate, error) {
	rates, err := s.shippingRepo.GetShippingRatesByProvider(providerID)
	if err != nil {
		logger.Errorf("Failed to get shipping rates by provider %d: %v", providerID, err)
		return nil, fmt.Errorf("failed to get shipping rates")
	}
	return rates, nil
}

func (s *shippingService) UpdateShippingRate(id uint, req *model.ShippingRate) (*model.ShippingRate, error) {
	req.ID = id
	if err := s.shippingRepo.UpdateShippingRate(req); err != nil {
		logger.Errorf("Failed to update shipping rate %d: %v", id, err)
		return nil, fmt.Errorf("failed to update shipping rate")
	}
	return req, nil
}

func (s *shippingService) DeleteShippingRate(id uint) error {
	if err := s.shippingRepo.DeleteShippingRate(id); err != nil {
		logger.Errorf("Failed to delete shipping rate %d: %v", id, err)
		return fmt.Errorf("failed to delete shipping rate")
	}
	return nil
}

// Shipping Calculation
func (s *shippingService) CalculateShipping(req *model.CalculateShippingRequest) ([]model.CalculateShippingResponse, error) {
	// Get shipping rates for calculation
	rates, err := s.shippingRepo.GetShippingRatesForCalculation(req)
	if err != nil {
		logger.Errorf("Failed to get shipping rates for calculation: %v", err)
		return nil, fmt.Errorf("failed to calculate shipping")
	}

	var responses []model.CalculateShippingResponse
	for _, rate := range rates {
		// Calculate fees
		shippingFee := rate.BaseFee
		if rate.WeightFee > 0 {
			shippingFee += (req.Weight - rate.MinWeight) * rate.WeightFee
		}
		if rate.ValueFee > 0 {
			shippingFee += (req.Value - rate.MinValue) * rate.ValueFee / 1000000 // Convert to VND
		}

		codFee := 0.0
		if req.COD > 0 && rate.COD > 0 {
			codFee = rate.COD
		}

		insuranceFee := 0.0
		if req.Insurance > 0 && rate.Insurance > 0 {
			insuranceFee = rate.Insurance
		}

		totalFee := shippingFee + codFee + insuranceFee

		response := model.CalculateShippingResponse{
			ProviderID:   rate.ProviderID,
			ProviderName: rate.Provider.DisplayName,
			ProviderCode: rate.Provider.Code,
			ShippingFee:  shippingFee,
			COD:          codFee,
			InsuranceFee: insuranceFee,
			TotalFee:     totalFee,
			MinDays:      rate.MinDays,
			MaxDays:      rate.MaxDays,
			IsAvailable:  true,
		}

		responses = append(responses, response)
	}

	return responses, nil
}

func (s *shippingService) CalculateShippingWithGHTK(req *model.CalculateShippingRequest) (*model.CalculateShippingResponse, error) {
	// Convert to GHTK format
	ghtkReq := &shipping.CalculateFeeRequest{
		PickProvince: req.FromProvince,
		PickDistrict: req.FromDistrict,
		PickWard:     "", // Will be filled from address
		Province:     req.ToProvince,
		District:     req.ToDistrict,
		Ward:         "", // Will be filled from address
		Value:        int(req.Value),
		Transport:    "road",                 // Default transport method
		Weight:       int(req.Weight * 1000), // Convert kg to grams
	}

	// Call GHTK API
	ghtkResp, err := s.ghtkClient.CalculateFee(ghtkReq)
	if err != nil {
		logger.Errorf("Failed to calculate GHTK shipping fee: %v", err)
		return nil, fmt.Errorf("failed to calculate shipping fee")
	}

	if !ghtkResp.Success {
		return nil, fmt.Errorf("GHTK API error: %s", ghtkResp.Message)
	}

	// Get GHTK provider
	ghtkProvider, err := s.shippingRepo.GetShippingProviderByCode(model.ProviderCodeGHTK)
	if err != nil {
		logger.Errorf("Failed to get GHTK provider: %v", err)
		return nil, fmt.Errorf("GHTK provider not found")
	}

	response := &model.CalculateShippingResponse{
		ProviderID:   ghtkProvider.ID,
		ProviderName: ghtkProvider.DisplayName,
		ProviderCode: ghtkProvider.Code,
		ShippingFee:  float64(ghtkResp.Fee.Fee),
		COD:          0, // GHTK handles COD separately
		InsuranceFee: float64(ghtkResp.Fee.InsuranceFee),
		TotalFee:     float64(ghtkResp.Fee.TotalFee),
		MinDays:      1, // GHTK typically delivers in 1-2 days
		MaxDays:      2,
		IsAvailable:  true,
	}

	return response, nil
}

// Shipping Orders
func (s *shippingService) CreateShippingOrder(req *model.ShippingOrderRequest) (*model.ShippingOrderResponse, error) {
	// Get order
	order, err := s.orderRepo.GetOrderByID(req.OrderID)
	if err != nil {
		logger.Errorf("Failed to get order %d: %v", req.OrderID, err)
		return nil, fmt.Errorf("order not found")
	}

	// Get provider
	provider, err := s.shippingRepo.GetShippingProviderByID(req.ProviderID)
	if err != nil {
		logger.Errorf("Failed to get provider %d: %v", req.ProviderID, err)
		return nil, fmt.Errorf("shipping provider not found")
	}

	// Create shipping order
	shippingOrder := &model.ShippingOrder{
		OrderID:     req.OrderID,
		ProviderID:  req.ProviderID,
		FromName:    "E-commerce Store",                   // Should be configurable
		FromAddress: "123 Store Street, District 1, HCMC", // Should be configurable
		FromPhone:   "0123456789",                         // Should be configurable
		FromEmail:   "store@example.com",                  // Should be configurable
		ToName:      order.CustomerName,
		ToAddress:   order.ShippingAddress,
		ToPhone:     order.CustomerPhone,
		ToEmail:     order.CustomerEmail,
		Weight:      req.Weight,
		Value:       req.Value,
		COD:         req.COD,
		Insurance:   req.Insurance,
		Status:      model.ShippingOrderStatusPending,
		StatusText:  "Pending",
	}

	// Calculate fees
	calcReq := &model.CalculateShippingRequest{
		FromProvince: "TP. Hồ Chí Minh", // Should be configurable
		FromDistrict: "Quận 1",          // Should be configurable
		ToProvince:   "TP. Hồ Chí Minh", // Should be extracted from address
		ToDistrict:   "Quận 1",          // Should be extracted from address
		Weight:       req.Weight,
		Value:        req.Value,
		ProviderID:   &req.ProviderID,
		COD:          req.COD,
		Insurance:    req.Insurance,
	}

	var calcResp *model.CalculateShippingResponse
	if provider.Code == model.ProviderCodeGHTK {
		calcResp, err = s.CalculateShippingWithGHTK(calcReq)
	} else {
		responses, err := s.CalculateShipping(calcReq)
		if err != nil || len(responses) == 0 {
			return nil, fmt.Errorf("failed to calculate shipping fee")
		}
		calcResp = &responses[0]
	}

	if err != nil {
		logger.Errorf("Failed to calculate shipping fee: %v", err)
		return nil, fmt.Errorf("failed to calculate shipping fee")
	}

	shippingOrder.ShippingFee = calcResp.ShippingFee
	shippingOrder.COD = calcResp.COD
	shippingOrder.InsuranceFee = calcResp.InsuranceFee
	shippingOrder.TotalFee = calcResp.TotalFee

	// Create shipping order in database
	if err := s.shippingRepo.CreateShippingOrder(shippingOrder); err != nil {
		logger.Errorf("Failed to create shipping order: %v", err)
		return nil, fmt.Errorf("failed to create shipping order")
	}

	// Create with external provider if needed
	if provider.Code == model.ProviderCodeGHTK {
		if err := s.createGHTKOrder(shippingOrder, order); err != nil {
			logger.Errorf("Failed to create GHTK order: %v", err)
			// Don't fail the entire operation, just log the error
		}
	}

	return s.toShippingOrderResponse(shippingOrder, provider), nil
}

func (s *shippingService) createGHTKOrder(shippingOrder *model.ShippingOrder, order *model.Order) error {
	// Convert to GHTK format
	ghtkReq := &shipping.CreateOrderRequest{
		Products: []shipping.GHTKProduct{
			{
				Name:        "Order Items",                    // Should be detailed
				Weight:      int(shippingOrder.Weight * 1000), // Convert to grams
				Quantity:    1,                                // Should be calculated from order items
				ProductCode: fmt.Sprintf("ORDER_%d", order.ID),
				Price:       shippingOrder.Value,
			},
		},
		Order: shipping.GHTKOrder{
			ID:             fmt.Sprintf("ORDER_%d", order.ID),
			PickName:       shippingOrder.FromName,
			PickAddress:    shippingOrder.FromAddress,
			PickProvince:   "TP. Hồ Chí Minh",  // Should be configurable
			PickDistrict:   "Quận 1",           // Should be configurable
			PickWard:       "Phường Bến Nghé",  // Should be configurable
			PickStreet:     "123 Store Street", // Should be configurable
			PickTel:        shippingOrder.FromPhone,
			PickEmail:      shippingOrder.FromEmail,
			Name:           shippingOrder.ToName,
			Address:        shippingOrder.ToAddress,
			Province:       "TP. Hồ Chí Minh",       // Should be extracted from address
			District:       "Quận 1",                // Should be extracted from address
			Ward:           "Phường Bến Nghé",       // Should be extracted from address
			Street:         shippingOrder.ToAddress, // Should be extracted
			Tel:            shippingOrder.ToPhone,
			Email:          shippingOrder.ToEmail,
			Note:           "E-commerce order",
			Value:          int(shippingOrder.Value),
			Transport:      "road",
			PickOption:     "cod", // Cash on delivery
			DeliverOption:  "cod",
			PickSession:    2, // Afternoon
			DeliverSession: 2, // Afternoon
		},
	}

	// Call GHTK API
	ghtkResp, err := s.ghtkClient.CreateOrder(ghtkReq)
	if err != nil {
		return fmt.Errorf("failed to create GHTK order: %v", err)
	}

	if !ghtkResp.Success {
		return fmt.Errorf("GHTK API error: %s", ghtkResp.Message)
	}

	// Update shipping order with GHTK response
	shippingOrder.ExternalID = ghtkResp.Order.PartnerID
	shippingOrder.LabelID = ghtkResp.Order.LabelID
	shippingOrder.TrackingCode = ghtkResp.Order.LabelID
	shippingOrder.Status = model.ShippingOrderStatusCreated
	shippingOrder.StatusText = "Created in GHTK"

	// Update in database
	if err := s.shippingRepo.UpdateShippingOrder(shippingOrder); err != nil {
		return fmt.Errorf("failed to update shipping order: %v", err)
	}

	return nil
}

func (s *shippingService) GetShippingOrderByID(id uint) (*model.ShippingOrderResponse, error) {
	shippingOrder, err := s.shippingRepo.GetShippingOrderByID(id)
	if err != nil {
		logger.Errorf("Failed to get shipping order by ID %d: %v", id, err)
		return nil, fmt.Errorf("shipping order not found")
	}

	provider, err := s.shippingRepo.GetShippingProviderByID(shippingOrder.ProviderID)
	if err != nil {
		logger.Errorf("Failed to get provider %d: %v", shippingOrder.ProviderID, err)
		return nil, fmt.Errorf("provider not found")
	}

	return s.toShippingOrderResponse(shippingOrder, provider), nil
}

func (s *shippingService) GetShippingOrderByOrderID(orderID uint) (*model.ShippingOrderResponse, error) {
	shippingOrder, err := s.shippingRepo.GetShippingOrderByOrderID(orderID)
	if err != nil {
		logger.Errorf("Failed to get shipping order by order ID %d: %v", orderID, err)
		return nil, fmt.Errorf("shipping order not found")
	}

	provider, err := s.shippingRepo.GetShippingProviderByID(shippingOrder.ProviderID)
	if err != nil {
		logger.Errorf("Failed to get provider %d: %v", shippingOrder.ProviderID, err)
		return nil, fmt.Errorf("provider not found")
	}

	return s.toShippingOrderResponse(shippingOrder, provider), nil
}

func (s *shippingService) GetShippingOrderByTrackingCode(trackingCode string) (*model.ShippingOrderResponse, error) {
	shippingOrder, err := s.shippingRepo.GetShippingOrderByTrackingCode(trackingCode)
	if err != nil {
		logger.Errorf("Failed to get shipping order by tracking code %s: %v", trackingCode, err)
		return nil, fmt.Errorf("shipping order not found")
	}

	provider, err := s.shippingRepo.GetShippingProviderByID(shippingOrder.ProviderID)
	if err != nil {
		logger.Errorf("Failed to get provider %d: %v", shippingOrder.ProviderID, err)
		return nil, fmt.Errorf("provider not found")
	}

	return s.toShippingOrderResponse(shippingOrder, provider), nil
}

func (s *shippingService) UpdateShippingOrder(id uint, req *model.ShippingOrder) (*model.ShippingOrderResponse, error) {
	req.ID = id
	if err := s.shippingRepo.UpdateShippingOrder(req); err != nil {
		logger.Errorf("Failed to update shipping order %d: %v", id, err)
		return nil, fmt.Errorf("failed to update shipping order")
	}

	provider, err := s.shippingRepo.GetShippingProviderByID(req.ProviderID)
	if err != nil {
		logger.Errorf("Failed to get provider %d: %v", req.ProviderID, err)
		return nil, fmt.Errorf("provider not found")
	}

	return s.toShippingOrderResponse(req, provider), nil
}

func (s *shippingService) CancelShippingOrder(id uint, reason string) error {
	shippingOrder, err := s.shippingRepo.GetShippingOrderByID(id)
	if err != nil {
		logger.Errorf("Failed to get shipping order %d: %v", id, err)
		return fmt.Errorf("shipping order not found")
	}

	provider, err := s.shippingRepo.GetShippingProviderByID(shippingOrder.ProviderID)
	if err != nil {
		logger.Errorf("Failed to get provider %d: %v", shippingOrder.ProviderID, err)
		return fmt.Errorf("provider not found")
	}

	// Cancel with external provider if needed
	if provider.Code == model.ProviderCodeGHTK && shippingOrder.LabelID != "" {
		ghtkReq := &shipping.CancelOrderRequest{
			LabelID: shippingOrder.LabelID,
			Note:    reason,
		}

		ghtkResp, err := s.ghtkClient.CancelOrder(ghtkReq)
		if err != nil {
			logger.Errorf("Failed to cancel GHTK order: %v", err)
			return fmt.Errorf("failed to cancel order with provider")
		}

		if !ghtkResp.Success {
			return fmt.Errorf("provider API error: %s", ghtkResp.Message)
		}
	}

	// Update status
	shippingOrder.Status = model.ShippingOrderStatusCancelled
	shippingOrder.StatusText = "Cancelled: " + reason

	if err := s.shippingRepo.UpdateShippingOrder(shippingOrder); err != nil {
		logger.Errorf("Failed to update shipping order %d: %v", id, err)
		return fmt.Errorf("failed to update shipping order")
	}

	return nil
}

func (s *shippingService) GetShippingOrders(page, limit int, filters map[string]interface{}) ([]model.ShippingOrderResponse, int64, error) {
	orders, total, err := s.shippingRepo.GetShippingOrders(page, limit, filters)
	if err != nil {
		logger.Errorf("Failed to get shipping orders: %v", err)
		return nil, 0, fmt.Errorf("failed to get shipping orders")
	}

	var responses []model.ShippingOrderResponse
	for _, order := range orders {
		provider, err := s.shippingRepo.GetShippingProviderByID(order.ProviderID)
		if err != nil {
			logger.Errorf("Failed to get provider %d: %v", order.ProviderID, err)
			continue
		}
		responses = append(responses, *s.toShippingOrderResponse(&order, provider))
	}

	return responses, total, nil
}

// Tracking
func (s *shippingService) GetShippingTracking(orderID uint) ([]model.ShippingTracking, error) {
	tracking, err := s.shippingRepo.GetShippingTrackingByOrderID(orderID)
	if err != nil {
		logger.Errorf("Failed to get shipping tracking for order %d: %v", orderID, err)
		return nil, fmt.Errorf("failed to get shipping tracking")
	}
	return tracking, nil
}

func (s *shippingService) UpdateShippingStatusFromWebhook(webhookData *model.WebhookData) error {
	// Find shipping order by label ID
	shippingOrder, err := s.shippingRepo.GetShippingOrderByLabelID(webhookData.LabelID)
	if err != nil {
		logger.Errorf("Failed to find shipping order by label ID %s: %v", webhookData.LabelID, err)
		return fmt.Errorf("shipping order not found")
	}

	// Update shipping order status
	shippingOrder.Status = webhookData.Status
	shippingOrder.StatusText = webhookData.StatusText

	if webhookData.DeliverDate != "" {
		if deliveredAt, err := time.Parse("2006-01-02 15:04:05", webhookData.DeliverDate); err == nil {
			shippingOrder.DeliveredAt = &deliveredAt
		}
	}

	if err := s.shippingRepo.UpdateShippingOrder(shippingOrder); err != nil {
		logger.Errorf("Failed to update shipping order: %v", err)
		return fmt.Errorf("failed to update shipping order")
	}

	// Create tracking entry
	tracking := &model.ShippingTracking{
		ShippingOrderID: shippingOrder.ID,
		Status:          webhookData.Status,
		StatusText:      webhookData.StatusText,
		Location:        "", // Will be extracted from timeline
		Note:            "Status updated from webhook",
	}

	if err := s.shippingRepo.CreateShippingTracking(tracking); err != nil {
		logger.Errorf("Failed to create shipping tracking: %v", err)
		// Don't fail the entire operation
	}

	// Update order status if needed
	if webhookData.Status == model.ShippingOrderStatusDelivered {
		// Get order and update status
		order, err := s.orderRepo.GetOrderByID(shippingOrder.OrderID)
		if err == nil && order != nil {
			order.Status = model.OrderStatusDelivered
			order.ShippingStatus = model.ShippingStatusDelivered
			now := time.Now()
			order.DeliveredAt = &now
			if err := s.orderRepo.UpdateOrder(order); err != nil {
				logger.Errorf("Failed to update order status: %v", err)
			}
		}
	}

	return nil
}

// Statistics
func (s *shippingService) GetShippingStats() (*model.ShippingStats, error) {
	return s.shippingRepo.GetShippingStats()
}

func (s *shippingService) GetShippingStatsByProvider(providerID uint) (*model.ShippingStats, error) {
	return s.shippingRepo.GetShippingStatsByProvider(providerID)
}

// Helper methods
func (s *shippingService) toShippingOrderResponse(order *model.ShippingOrder, provider *model.ShippingProvider) *model.ShippingOrderResponse {
	return &model.ShippingOrderResponse{
		ID:           order.ID,
		OrderID:      order.OrderID,
		ProviderID:   order.ProviderID,
		ProviderName: provider.DisplayName,
		ExternalID:   order.ExternalID,
		LabelID:      order.LabelID,
		TrackingCode: order.TrackingCode,
		Status:       order.Status,
		StatusText:   order.StatusText,
		ShippingFee:  order.ShippingFee,
		TotalFee:     order.TotalFee,
		CreatedAt:    order.CreatedAt,
		ShippedAt:    order.ShippedAt,
		DeliveredAt:  order.DeliveredAt,
	}
}
