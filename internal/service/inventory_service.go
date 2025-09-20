package service

import (
	"errors"
	"fmt"
	"go_app/internal/model"
	"go_app/internal/repository"
	"go_app/pkg/logger"
	"time"
)

// InventoryService defines methods for inventory business logic
type InventoryService interface {
	// Inventory Movements
	CreateMovement(req *model.InventoryMovementCreateRequest, userID uint) (*model.InventoryMovementResponse, error)
	GetMovementByID(id uint) (*model.InventoryMovementResponse, error)
	GetMovements(page, limit int, filters map[string]interface{}) ([]model.InventoryMovementResponse, int64, error)
	UpdateMovement(id uint, req *model.InventoryMovementUpdateRequest, userID uint) (*model.InventoryMovementResponse, error)
	DeleteMovement(id uint) error
	ApproveMovement(id uint, userID uint) (*model.InventoryMovementResponse, error)
	CompleteMovement(id uint) (*model.InventoryMovementResponse, error)
	GetMovementsByProduct(productID uint, variantID *uint) ([]model.InventoryMovementResponse, error)
	GetMovementsByReference(reference string) ([]model.InventoryMovementResponse, error)

	// Stock Levels
	GetStockLevelByProduct(productID uint, variantID *uint) (*model.StockLevelResponse, error)
	GetAllStockLevels(page, limit int, filters map[string]interface{}) ([]model.StockLevelResponse, int64, error)
	UpdateStockLevelSettings(productID uint, variantID *uint, req *model.StockLevelUpdateRequest) (*model.StockLevelResponse, error)
	GetLowStockProducts(threshold int) ([]model.StockLevelResponse, error)
	GetOutOfStockProducts() ([]model.StockLevelResponse, error)

	// Inventory Adjustments
	CreateAdjustment(req *model.InventoryAdjustmentCreateRequest, userID uint) (*model.InventoryAdjustmentResponse, error)
	GetAdjustmentByID(id uint) (*model.InventoryAdjustmentResponse, error)
	GetAdjustments(page, limit int, filters map[string]interface{}) ([]model.InventoryAdjustmentResponse, int64, error)
	GetAdjustmentsByProduct(productID uint, variantID *uint) ([]model.InventoryAdjustmentResponse, error)

	// Statistics and Reports
	GetInventoryStats() (*model.InventoryStatsResponse, error)
	GetLowStockAlerts() ([]model.LowStockAlert, error)
	GetStockValue() (float64, error)
	GetMovementStats(startDate, endDate time.Time) (map[string]interface{}, error)

	// Stock Operations
	ReserveStock(productID uint, variantID *uint, quantity int) error
	ReleaseStock(productID uint, variantID *uint, quantity int) error
	ProcessStockMovement(movement *model.InventoryMovement) error
}

// inventoryService implements InventoryService
type inventoryService struct {
	inventoryRepo repository.InventoryRepository
	productRepo   *repository.ProductRepository
}

// NewInventoryService creates a new InventoryService
func NewInventoryService() InventoryService {
	return &inventoryService{
		inventoryRepo: repository.NewInventoryRepository(),
		productRepo:   repository.NewProductRepository(),
	}
}

// Inventory Movements

// CreateMovement creates a new inventory movement
func (s *inventoryService) CreateMovement(req *model.InventoryMovementCreateRequest, userID uint) (*model.InventoryMovementResponse, error) {
	// Validate product exists
	product, err := s.productRepo.GetByID(req.ProductID)
	if err != nil {
		logger.Errorf("Error getting product by ID %d: %v", req.ProductID, err)
		return nil, fmt.Errorf("failed to retrieve product")
	}
	if product == nil {
		return nil, errors.New("product not found")
	}

	// Validate variant exists if provided
	if req.VariantID != nil {
		// Check if variant exists and belongs to product
		// This would require a variant repository method
		// For now, we'll skip this validation
	}

	// Calculate total cost
	totalCost := float64(req.Quantity) * req.UnitCost

	movement := &model.InventoryMovement{
		ProductID:     req.ProductID,
		VariantID:     req.VariantID,
		Type:          req.Type,
		Status:        model.MovementStatusPending,
		Quantity:      req.Quantity,
		UnitCost:      req.UnitCost,
		TotalCost:     totalCost,
		Reference:     req.Reference,
		ReferenceType: req.ReferenceType,
		Notes:         req.Notes,
		CreatedBy:     userID,
	}

	if err := s.inventoryRepo.CreateMovement(movement); err != nil {
		logger.Errorf("Error creating inventory movement: %v", err)
		return nil, fmt.Errorf("failed to create inventory movement")
	}

	return s.toMovementResponse(movement), nil
}

// GetMovementByID retrieves an inventory movement by its ID
func (s *inventoryService) GetMovementByID(id uint) (*model.InventoryMovementResponse, error) {
	movement, err := s.inventoryRepo.GetMovementByID(id)
	if err != nil {
		logger.Errorf("Error getting movement by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve movement")
	}
	if movement == nil {
		return nil, errors.New("movement not found")
	}
	return s.toMovementResponse(movement), nil
}

// GetMovements retrieves inventory movements with pagination and filters
func (s *inventoryService) GetMovements(page, limit int, filters map[string]interface{}) ([]model.InventoryMovementResponse, int64, error) {
	movements, total, err := s.inventoryRepo.GetMovements(page, limit, filters)
	if err != nil {
		logger.Errorf("Error getting movements: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve movements")
	}

	var responses []model.InventoryMovementResponse
	for _, movement := range movements {
		responses = append(responses, *s.toMovementResponse(&movement))
	}
	return responses, total, nil
}

// UpdateMovement updates an existing inventory movement
func (s *inventoryService) UpdateMovement(id uint, req *model.InventoryMovementUpdateRequest, userID uint) (*model.InventoryMovementResponse, error) {
	movement, err := s.inventoryRepo.GetMovementByID(id)
	if err != nil {
		logger.Errorf("Error getting movement by ID %d for update: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve movement")
	}
	if movement == nil {
		return nil, errors.New("movement not found")
	}

	// Only allow updating pending movements
	if movement.Status != model.MovementStatusPending {
		return nil, errors.New("only pending movements can be updated")
	}

	movement.Status = req.Status
	movement.Notes = req.Notes

	if err := s.inventoryRepo.UpdateMovement(movement); err != nil {
		logger.Errorf("Error updating movement %d: %v", id, err)
		return nil, fmt.Errorf("failed to update movement")
	}

	return s.toMovementResponse(movement), nil
}

// DeleteMovement deletes an inventory movement
func (s *inventoryService) DeleteMovement(id uint) error {
	movement, err := s.inventoryRepo.GetMovementByID(id)
	if err != nil {
		logger.Errorf("Error getting movement by ID %d for deletion: %v", id, err)
		return fmt.Errorf("failed to retrieve movement")
	}
	if movement == nil {
		return errors.New("movement not found")
	}

	// Only allow deleting pending movements
	if movement.Status != model.MovementStatusPending {
		return errors.New("only pending movements can be deleted")
	}

	if err := s.inventoryRepo.DeleteMovement(id); err != nil {
		logger.Errorf("Error deleting movement %d: %v", id, err)
		return fmt.Errorf("failed to delete movement")
	}
	return nil
}

// ApproveMovement approves an inventory movement
func (s *inventoryService) ApproveMovement(id uint, userID uint) (*model.InventoryMovementResponse, error) {
	movement, err := s.inventoryRepo.GetMovementByID(id)
	if err != nil {
		logger.Errorf("Error getting movement by ID %d for approval: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve movement")
	}
	if movement == nil {
		return nil, errors.New("movement not found")
	}

	if movement.Status != model.MovementStatusPending {
		return nil, errors.New("only pending movements can be approved")
	}

	if err := s.inventoryRepo.ApproveMovement(id, userID); err != nil {
		logger.Errorf("Error approving movement %d: %v", id, err)
		return nil, fmt.Errorf("failed to approve movement")
	}

	// Get updated movement
	updatedMovement, err := s.inventoryRepo.GetMovementByID(id)
	if err != nil {
		logger.Errorf("Error getting updated movement %d: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve updated movement")
	}

	return s.toMovementResponse(updatedMovement), nil
}

// CompleteMovement completes an inventory movement
func (s *inventoryService) CompleteMovement(id uint) (*model.InventoryMovementResponse, error) {
	movement, err := s.inventoryRepo.GetMovementByID(id)
	if err != nil {
		logger.Errorf("Error getting movement by ID %d for completion: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve movement")
	}
	if movement == nil {
		return nil, errors.New("movement not found")
	}

	if movement.Status != model.MovementStatusApproved {
		return nil, errors.New("only approved movements can be completed")
	}

	// Process the stock movement
	if err := s.ProcessStockMovement(movement); err != nil {
		logger.Errorf("Error processing stock movement %d: %v", id, err)
		return nil, fmt.Errorf("failed to process stock movement")
	}

	if err := s.inventoryRepo.CompleteMovement(id); err != nil {
		logger.Errorf("Error completing movement %d: %v", id, err)
		return nil, fmt.Errorf("failed to complete movement")
	}

	// Get updated movement
	updatedMovement, err := s.inventoryRepo.GetMovementByID(id)
	if err != nil {
		logger.Errorf("Error getting updated movement %d: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve updated movement")
	}

	return s.toMovementResponse(updatedMovement), nil
}

// GetMovementsByProduct retrieves movements for a specific product/variant
func (s *inventoryService) GetMovementsByProduct(productID uint, variantID *uint) ([]model.InventoryMovementResponse, error) {
	movements, err := s.inventoryRepo.GetMovementsByProduct(productID, variantID)
	if err != nil {
		logger.Errorf("Error getting movements for product %d: %v", productID, err)
		return nil, fmt.Errorf("failed to retrieve movements")
	}

	var responses []model.InventoryMovementResponse
	for _, movement := range movements {
		responses = append(responses, *s.toMovementResponse(&movement))
	}
	return responses, nil
}

// GetMovementsByReference retrieves movements by reference
func (s *inventoryService) GetMovementsByReference(reference string) ([]model.InventoryMovementResponse, error) {
	movements, err := s.inventoryRepo.GetMovementsByReference(reference)
	if err != nil {
		logger.Errorf("Error getting movements by reference %s: %v", reference, err)
		return nil, fmt.Errorf("failed to retrieve movements")
	}

	var responses []model.InventoryMovementResponse
	for _, movement := range movements {
		responses = append(responses, *s.toMovementResponse(&movement))
	}
	return responses, nil
}

// Stock Levels

// GetStockLevelByProduct retrieves stock level for a specific product/variant
func (s *inventoryService) GetStockLevelByProduct(productID uint, variantID *uint) (*model.StockLevelResponse, error) {
	stockLevel, err := s.inventoryRepo.GetStockLevelByProduct(productID, variantID)
	if err != nil {
		logger.Errorf("Error getting stock level for product %d: %v", productID, err)
		return nil, fmt.Errorf("failed to retrieve stock level")
	}
	if stockLevel == nil {
		return nil, errors.New("stock level not found")
	}
	return s.toStockLevelResponse(stockLevel), nil
}

// GetAllStockLevels retrieves all stock levels with pagination and filters
func (s *inventoryService) GetAllStockLevels(page, limit int, filters map[string]interface{}) ([]model.StockLevelResponse, int64, error) {
	stockLevels, total, err := s.inventoryRepo.GetAllStockLevels(page, limit, filters)
	if err != nil {
		logger.Errorf("Error getting stock levels: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve stock levels")
	}

	var responses []model.StockLevelResponse
	for _, stockLevel := range stockLevels {
		responses = append(responses, *s.toStockLevelResponse(&stockLevel))
	}
	return responses, total, nil
}

// UpdateStockLevelSettings updates stock level settings
func (s *inventoryService) UpdateStockLevelSettings(productID uint, variantID *uint, req *model.StockLevelUpdateRequest) (*model.StockLevelResponse, error) {
	stockLevel, err := s.inventoryRepo.GetStockLevelByProduct(productID, variantID)
	if err != nil {
		logger.Errorf("Error getting stock level for product %d: %v", productID, err)
		return nil, fmt.Errorf("failed to retrieve stock level")
	}
	if stockLevel == nil {
		return nil, errors.New("stock level not found")
	}

	stockLevel.MinStockLevel = req.MinStockLevel
	stockLevel.MaxStockLevel = req.MaxStockLevel
	stockLevel.ReorderPoint = req.ReorderPoint

	if err := s.inventoryRepo.UpdateStockLevel(stockLevel); err != nil {
		logger.Errorf("Error updating stock level for product %d: %v", productID, err)
		return nil, fmt.Errorf("failed to update stock level")
	}

	return s.toStockLevelResponse(stockLevel), nil
}

// GetLowStockProducts retrieves products with low stock
func (s *inventoryService) GetLowStockProducts(threshold int) ([]model.StockLevelResponse, error) {
	stockLevels, err := s.inventoryRepo.GetLowStockProducts(threshold)
	if err != nil {
		logger.Errorf("Error getting low stock products: %v", err)
		return nil, fmt.Errorf("failed to retrieve low stock products")
	}

	var responses []model.StockLevelResponse
	for _, stockLevel := range stockLevels {
		responses = append(responses, *s.toStockLevelResponse(&stockLevel))
	}
	return responses, nil
}

// GetOutOfStockProducts retrieves out of stock products
func (s *inventoryService) GetOutOfStockProducts() ([]model.StockLevelResponse, error) {
	stockLevels, err := s.inventoryRepo.GetOutOfStockProducts()
	if err != nil {
		logger.Errorf("Error getting out of stock products: %v", err)
		return nil, fmt.Errorf("failed to retrieve out of stock products")
	}

	var responses []model.StockLevelResponse
	for _, stockLevel := range stockLevels {
		responses = append(responses, *s.toStockLevelResponse(&stockLevel))
	}
	return responses, nil
}

// Inventory Adjustments

// CreateAdjustment creates a new inventory adjustment
func (s *inventoryService) CreateAdjustment(req *model.InventoryAdjustmentCreateRequest, userID uint) (*model.InventoryAdjustmentResponse, error) {
	// Get current stock level
	stockLevel, err := s.inventoryRepo.GetStockLevelByProduct(req.ProductID, req.VariantID)
	if err != nil {
		logger.Errorf("Error getting stock level for product %d: %v", req.ProductID, err)
		return nil, fmt.Errorf("failed to retrieve stock level")
	}
	if stockLevel == nil {
		return nil, errors.New("stock level not found")
	}

	quantityDiff := req.QuantityAfter - stockLevel.AvailableQuantity

	adjustment := &model.InventoryAdjustment{
		ProductID:      req.ProductID,
		VariantID:      req.VariantID,
		Reason:         req.Reason,
		QuantityBefore: stockLevel.AvailableQuantity,
		QuantityAfter:  req.QuantityAfter,
		QuantityDiff:   quantityDiff,
		Notes:          req.Notes,
		CreatedBy:      userID,
	}

	if err := s.inventoryRepo.CreateAdjustment(adjustment); err != nil {
		logger.Errorf("Error creating inventory adjustment: %v", err)
		return nil, fmt.Errorf("failed to create inventory adjustment")
	}

	// Update stock level
	if err := s.inventoryRepo.UpdateStockQuantity(req.ProductID, req.VariantID, req.QuantityAfter); err != nil {
		logger.Errorf("Error updating stock quantity after adjustment: %v", err)
		return nil, fmt.Errorf("failed to update stock quantity")
	}

	return s.toAdjustmentResponse(adjustment), nil
}

// GetAdjustmentByID retrieves an inventory adjustment by its ID
func (s *inventoryService) GetAdjustmentByID(id uint) (*model.InventoryAdjustmentResponse, error) {
	adjustment, err := s.inventoryRepo.GetAdjustmentByID(id)
	if err != nil {
		logger.Errorf("Error getting adjustment by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve adjustment")
	}
	if adjustment == nil {
		return nil, errors.New("adjustment not found")
	}
	return s.toAdjustmentResponse(adjustment), nil
}

// GetAdjustments retrieves inventory adjustments with pagination and filters
func (s *inventoryService) GetAdjustments(page, limit int, filters map[string]interface{}) ([]model.InventoryAdjustmentResponse, int64, error) {
	adjustments, total, err := s.inventoryRepo.GetAdjustments(page, limit, filters)
	if err != nil {
		logger.Errorf("Error getting adjustments: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve adjustments")
	}

	var responses []model.InventoryAdjustmentResponse
	for _, adjustment := range adjustments {
		responses = append(responses, *s.toAdjustmentResponse(&adjustment))
	}
	return responses, total, nil
}

// GetAdjustmentsByProduct retrieves adjustments for a specific product/variant
func (s *inventoryService) GetAdjustmentsByProduct(productID uint, variantID *uint) ([]model.InventoryAdjustmentResponse, error) {
	adjustments, err := s.inventoryRepo.GetAdjustmentsByProduct(productID, variantID)
	if err != nil {
		logger.Errorf("Error getting adjustments for product %d: %v", productID, err)
		return nil, fmt.Errorf("failed to retrieve adjustments")
	}

	var responses []model.InventoryAdjustmentResponse
	for _, adjustment := range adjustments {
		responses = append(responses, *s.toAdjustmentResponse(&adjustment))
	}
	return responses, nil
}

// Statistics and Reports

// GetInventoryStats retrieves inventory statistics
func (s *inventoryService) GetInventoryStats() (*model.InventoryStatsResponse, error) {
	stats, err := s.inventoryRepo.GetInventoryStats()
	if err != nil {
		logger.Errorf("Error getting inventory stats: %v", err)
		return nil, fmt.Errorf("failed to retrieve inventory statistics")
	}
	return stats, nil
}

// GetLowStockAlerts retrieves low stock alerts
func (s *inventoryService) GetLowStockAlerts() ([]model.LowStockAlert, error) {
	alerts, err := s.inventoryRepo.GetLowStockAlerts()
	if err != nil {
		logger.Errorf("Error getting low stock alerts: %v", err)
		return nil, fmt.Errorf("failed to retrieve low stock alerts")
	}
	return alerts, nil
}

// GetStockValue calculates total stock value
func (s *inventoryService) GetStockValue() (float64, error) {
	value, err := s.inventoryRepo.GetStockValue()
	if err != nil {
		logger.Errorf("Error getting stock value: %v", err)
		return 0, fmt.Errorf("failed to calculate stock value")
	}
	return value, nil
}

// GetMovementStats retrieves movement statistics for a date range
func (s *inventoryService) GetMovementStats(startDate, endDate time.Time) (map[string]interface{}, error) {
	stats, err := s.inventoryRepo.GetMovementStats(startDate, endDate)
	if err != nil {
		logger.Errorf("Error getting movement stats: %v", err)
		return nil, fmt.Errorf("failed to retrieve movement statistics")
	}
	return stats, nil
}

// Stock Operations

// ReserveStock reserves stock for a product/variant
func (s *inventoryService) ReserveStock(productID uint, variantID *uint, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	// Check available stock
	stockLevel, err := s.inventoryRepo.GetStockLevelByProduct(productID, variantID)
	if err != nil {
		logger.Errorf("Error getting stock level for product %d: %v", productID, err)
		return fmt.Errorf("failed to retrieve stock level")
	}
	if stockLevel == nil {
		return errors.New("stock level not found")
	}

	if stockLevel.AvailableQuantity < quantity {
		return errors.New("insufficient stock available")
	}

	if err := s.inventoryRepo.ReserveStock(productID, variantID, quantity); err != nil {
		logger.Errorf("Error reserving stock for product %d: %v", productID, err)
		return fmt.Errorf("failed to reserve stock")
	}

	return nil
}

// ReleaseStock releases reserved stock for a product/variant
func (s *inventoryService) ReleaseStock(productID uint, variantID *uint, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	if err := s.inventoryRepo.ReleaseStock(productID, variantID, quantity); err != nil {
		logger.Errorf("Error releasing stock for product %d: %v", productID, err)
		return fmt.Errorf("failed to release stock")
	}

	return nil
}

// ProcessStockMovement processes a stock movement and updates stock levels
func (s *inventoryService) ProcessStockMovement(movement *model.InventoryMovement) error {
	// Get or create stock level
	stockLevel, err := s.inventoryRepo.GetStockLevelByProduct(movement.ProductID, movement.VariantID)
	if err != nil {
		logger.Errorf("Error getting stock level for product %d: %v", movement.ProductID, err)
		return fmt.Errorf("failed to retrieve stock level")
	}

	if stockLevel == nil {
		// Create new stock level
		stockLevel = &model.StockLevel{
			ProductID:         movement.ProductID,
			VariantID:         movement.VariantID,
			AvailableQuantity: 0,
			ReservedQuantity:  0,
			IncomingQuantity:  0,
			TotalQuantity:     0,
		}
		if err := s.inventoryRepo.CreateStockLevel(stockLevel); err != nil {
			logger.Errorf("Error creating stock level: %v", err)
			return fmt.Errorf("failed to create stock level")
		}
	}

	// Update stock based on movement type
	var newQuantity int
	switch movement.Type {
	case model.MovementTypeInbound, model.MovementTypeReturn:
		newQuantity = stockLevel.AvailableQuantity + movement.Quantity
	case model.MovementTypeOutbound, model.MovementTypeTransfer:
		newQuantity = stockLevel.AvailableQuantity - movement.Quantity
		if newQuantity < 0 {
			return errors.New("insufficient stock for outbound movement")
		}
	case model.MovementTypeAdjustment:
		newQuantity = movement.Quantity
	}

	// Update stock level
	stockLevel.AvailableQuantity = newQuantity
	stockLevel.TotalQuantity = newQuantity
	stockLevel.LastMovementAt = &movement.CreatedAt

	if err := s.inventoryRepo.UpdateStockLevel(stockLevel); err != nil {
		logger.Errorf("Error updating stock level: %v", err)
		return fmt.Errorf("failed to update stock level")
	}

	return nil
}

// Helper methods for converting models to responses

func (s *inventoryService) toMovementResponse(movement *model.InventoryMovement) *model.InventoryMovementResponse {
	response := &model.InventoryMovementResponse{
		ID:            movement.ID,
		ProductID:     movement.ProductID,
		VariantID:     movement.VariantID,
		Type:          movement.Type,
		Status:        movement.Status,
		Quantity:      movement.Quantity,
		UnitCost:      movement.UnitCost,
		TotalCost:     movement.TotalCost,
		Reference:     movement.Reference,
		ReferenceType: movement.ReferenceType,
		Notes:         movement.Notes,
		CreatedBy:     movement.CreatedBy,
		ApprovedBy:    movement.ApprovedBy,
		ApprovedAt:    movement.ApprovedAt,
		CompletedAt:   movement.CompletedAt,
		CreatedAt:     movement.CreatedAt,
		UpdatedAt:     movement.UpdatedAt,
	}

	if movement.Product != nil {
		response.ProductName = movement.Product.Name
	}
	if movement.Variant != nil {
		response.VariantName = movement.Variant.Name
	}
	if movement.CreatedByUser != nil {
		response.CreatedByName = fmt.Sprintf("%s %s", movement.CreatedByUser.FirstName, movement.CreatedByUser.LastName)
	}
	if movement.ApprovedByUser != nil {
		response.ApprovedByName = fmt.Sprintf("%s %s", movement.ApprovedByUser.FirstName, movement.ApprovedByUser.LastName)
	}

	return response
}

func (s *inventoryService) toStockLevelResponse(stockLevel *model.StockLevel) *model.StockLevelResponse {
	response := &model.StockLevelResponse{
		ID:                stockLevel.ID,
		ProductID:         stockLevel.ProductID,
		VariantID:         stockLevel.VariantID,
		AvailableQuantity: stockLevel.AvailableQuantity,
		ReservedQuantity:  stockLevel.ReservedQuantity,
		IncomingQuantity:  stockLevel.IncomingQuantity,
		TotalQuantity:     stockLevel.TotalQuantity,
		MinStockLevel:     stockLevel.MinStockLevel,
		MaxStockLevel:     stockLevel.MaxStockLevel,
		ReorderPoint:      stockLevel.ReorderPoint,
		LastMovementAt:    stockLevel.LastMovementAt,
		CreatedAt:         stockLevel.CreatedAt,
		UpdatedAt:         stockLevel.UpdatedAt,
	}

	if stockLevel.Product != nil {
		response.ProductName = stockLevel.Product.Name
	}
	if stockLevel.Variant != nil {
		response.VariantName = stockLevel.Variant.Name
	}

	return response
}

func (s *inventoryService) toAdjustmentResponse(adjustment *model.InventoryAdjustment) *model.InventoryAdjustmentResponse {
	response := &model.InventoryAdjustmentResponse{
		ID:             adjustment.ID,
		ProductID:      adjustment.ProductID,
		VariantID:      adjustment.VariantID,
		Reason:         adjustment.Reason,
		QuantityBefore: adjustment.QuantityBefore,
		QuantityAfter:  adjustment.QuantityAfter,
		QuantityDiff:   adjustment.QuantityDiff,
		Notes:          adjustment.Notes,
		CreatedBy:      adjustment.CreatedBy,
		CreatedAt:      adjustment.CreatedAt,
		UpdatedAt:      adjustment.UpdatedAt,
	}

	if adjustment.Product != nil {
		response.ProductName = adjustment.Product.Name
	}
	if adjustment.Variant != nil {
		response.VariantName = adjustment.Variant.Name
	}
	if adjustment.CreatedByUser != nil {
		response.CreatedByName = fmt.Sprintf("%s %s", adjustment.CreatedByUser.FirstName, adjustment.CreatedByUser.LastName)
	}

	return response
}
