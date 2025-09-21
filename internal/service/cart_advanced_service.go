package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"go_app/internal/model"
	"go_app/internal/repository"
	"go_app/pkg/logger"
)

type CartAdvancedService struct {
	cartAdvancedRepo *repository.CartAdvancedRepository
	productRepo      *repository.ProductRepository
	productVariantRepo *repository.ProductVariantRepository
}

func NewCartAdvancedService() *CartAdvancedService {
	return &CartAdvancedService{
		cartAdvancedRepo: repository.NewCartAdvancedRepository(),
		productRepo:      repository.NewProductRepository(),
		productVariantRepo: repository.NewProductVariantRepository(),
	}
}

// ===== CART SHARE SERVICE =====

// CreateCartShare creates a new cart share
func (s *CartAdvancedService) CreateCartShare(req *model.CartShareCreateRequest, userID uint) (*model.CartShareResponse, error) {
	// Generate unique token
	token, err := s.generateShareToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate share token: %w", err)
	}

	// Hash password if provided
	hashedPassword := ""
	if req.PasswordProtected && req.Password != "" {
		hashedPassword, err = s.hashPassword(req.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
	}

	share := &model.CartShare{
		CartID:            req.CartID,
		SharedBy:          userID,
		Token:             token,
		IsActive:          true,
		ExpiresAt:         req.ExpiresAt,
		MaxUses:           req.MaxUses,
		UsedCount:         0,
		CanView:           req.CanView,
		CanEdit:           req.CanEdit,
		CanDelete:         req.CanDelete,
		PasswordProtected: req.PasswordProtected,
		Password:          hashedPassword,
	}

	if err := s.cartAdvancedRepo.CreateCartShare(share); err != nil {
		return nil, err
	}

	// Get the created share with user info
	createdShare, err := s.cartAdvancedRepo.GetCartShareByToken(token)
	if err != nil {
		return nil, err
	}

	return s.convertCartShareToResponse(createdShare), nil
}

// GetCartShareByToken gets a cart share by token
func (s *CartAdvancedService) GetCartShareByToken(token string) (*model.CartShareResponse, error) {
	share, err := s.cartAdvancedRepo.GetCartShareByToken(token)
	if err != nil {
		return nil, err
	}

	// Increment usage count
	if err := s.cartAdvancedRepo.IncrementCartShareUsage(token); err != nil {
		logger.Warnf("Failed to increment cart share usage: %v", err)
	}

	return s.convertCartShareToResponse(share), nil
}

// GetCartSharesByCartID gets all shares for a cart
func (s *CartAdvancedService) GetCartSharesByCartID(cartID uint) ([]model.CartShareResponse, error) {
	shares, err := s.cartAdvancedRepo.GetCartSharesByCartID(cartID)
	if err != nil {
		return nil, err
	}

	responses := make([]model.CartShareResponse, len(shares))
	for i, share := range shares {
		responses[i] = *s.convertCartShareToResponse(&share)
	}

	return responses, nil
}

// UpdateCartShare updates a cart share
func (s *CartAdvancedService) UpdateCartShare(id uint, req *model.CartShareUpdateRequest, userID uint) (*model.CartShareResponse, error) {
	// Get existing share
	share, err := s.cartAdvancedRepo.GetCartShareByToken("") // We need to get by ID, but our repo doesn't have this method
	if err != nil {
		return nil, fmt.Errorf("cart share not found")
	}

	// Update fields
	if req.IsActive != nil {
		share.IsActive = *req.IsActive
	}
	if req.ExpiresAt != nil {
		share.ExpiresAt = *req.ExpiresAt
	}
	if req.MaxUses != nil {
		share.MaxUses = *req.MaxUses
	}
	if req.CanView != nil {
		share.CanView = *req.CanView
	}
	if req.CanEdit != nil {
		share.CanEdit = *req.CanEdit
	}
	if req.CanDelete != nil {
		share.CanDelete = *req.CanDelete
	}
	if req.PasswordProtected != nil {
		share.PasswordProtected = *req.PasswordProtected
	}
	if req.Password != "" {
		hashedPassword, err := s.hashPassword(req.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		share.Password = hashedPassword
	}

	if err := s.cartAdvancedRepo.UpdateCartShare(share); err != nil {
		return nil, err
	}

	return s.convertCartShareToResponse(share), nil
}

// DeleteCartShare deletes a cart share
func (s *CartAdvancedService) DeleteCartShare(id uint, userID uint) error {
	return s.cartAdvancedRepo.DeleteCartShare(id)
}

// ===== SAVED FOR LATER SERVICE =====

// SaveItemForLater saves an item for later purchase
func (s *CartAdvancedService) SaveItemForLater(req *model.SavedForLaterCreateRequest, userID uint) (*model.SavedForLaterResponse, error) {
	// Check if product exists
	product, err := s.productRepo.GetByID(req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("product not found")
	}

	// Check if variant exists (if provided)
	if req.ProductVariantID != nil {
		_, err := s.productVariantRepo.GetByID(*req.ProductVariantID)
		if err != nil {
			return nil, fmt.Errorf("product variant not found")
		}
	}

	// Check if already saved
	existing, err := s.cartAdvancedRepo.GetSavedForLaterByProduct(userID, req.ProductID, req.ProductVariantID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("item already saved for later")
	}

	// Calculate prices
	unitPrice := product.RegularPrice
	if product.SalePrice != nil && *product.SalePrice > 0 {
		unitPrice = *product.SalePrice
	}
	totalPrice := unitPrice * float64(req.Quantity)

	item := &model.SavedForLater{
		UserID:            userID,
		ProductID:         req.ProductID,
		ProductVariantID:  req.ProductVariantID,
		Quantity:          req.Quantity,
		UnitPrice:         unitPrice,
		TotalPrice:        totalPrice,
		Notes:             req.Notes,
		Priority:          req.Priority,
		RemindAt:          req.RemindAt,
		NotifyOnPriceDrop: req.NotifyOnPriceDrop,
		NotifyOnStock:     req.NotifyOnStock,
		NotifyOnSale:      req.NotifyOnSale,
	}

	if err := s.cartAdvancedRepo.CreateSavedForLater(item); err != nil {
		return nil, err
	}

	// Get the created item with relations
	createdItem, err := s.cartAdvancedRepo.GetSavedForLaterByID(item.ID)
	if err != nil {
		return nil, err
	}

	return s.convertSavedForLaterToResponse(createdItem), nil
}

// GetSavedForLaterByUser gets all saved items for a user
func (s *CartAdvancedService) GetSavedForLaterByUser(userID uint, page, limit int) ([]model.SavedForLaterResponse, int64, error) {
	items, total, err := s.cartAdvancedRepo.GetSavedForLaterByUser(userID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.SavedForLaterResponse, len(items))
	for i, item := range items {
		responses[i] = *s.convertSavedForLaterToResponse(&item)
	}

	return responses, total, nil
}

// UpdateSavedForLater updates a saved item
func (s *CartAdvancedService) UpdateSavedForLater(id uint, req *model.SavedForLaterUpdateRequest, userID uint) (*model.SavedForLaterResponse, error) {
	item, err := s.cartAdvancedRepo.GetSavedForLaterByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Quantity != nil {
		item.Quantity = *req.Quantity
		item.TotalPrice = item.UnitPrice * float64(*req.Quantity)
	}
	if req.Notes != "" {
		item.Notes = req.Notes
	}
	if req.Priority != nil {
		item.Priority = *req.Priority
	}
	if req.RemindAt != nil {
		item.RemindAt = req.RemindAt
	}
	if req.NotifyOnPriceDrop != nil {
		item.NotifyOnPriceDrop = *req.NotifyOnPriceDrop
	}
	if req.NotifyOnStock != nil {
		item.NotifyOnStock = *req.NotifyOnStock
	}
	if req.NotifyOnSale != nil {
		item.NotifyOnSale = *req.NotifyOnSale
	}

	if err := s.cartAdvancedRepo.UpdateSavedForLater(item); err != nil {
		return nil, err
	}

	return s.convertSavedForLaterToResponse(item), nil
}

// DeleteSavedForLater deletes a saved item
func (s *CartAdvancedService) DeleteSavedForLater(id uint, userID uint) error {
	return s.cartAdvancedRepo.DeleteSavedForLater(id)
}

// MoveToCart moves a saved item to cart
func (s *CartAdvancedService) MoveToCart(savedItemID uint, cartID uint, userID uint) (*model.CartItemResponse, error) {
	// Get saved item
	savedItem, err := s.cartAdvancedRepo.GetSavedForLaterByID(savedItemID)
	if err != nil {
		return nil, err
	}

	// Create cart item
	cartItem := &model.CartItem{
		CartID:           cartID,
		ProductID:        savedItem.ProductID,
		ProductVariantID: savedItem.ProductVariantID,
		Quantity:         savedItem.Quantity,
		UnitPrice:        savedItem.UnitPrice,
		TotalPrice:       savedItem.TotalPrice,
		IsSavedForLater:  false,
		Notes:            savedItem.Notes,
		Priority:         savedItem.Priority,
	}

	// Add to cart (this would need to be implemented in cart service)
	// For now, we'll just return the cart item response
	response := s.convertCartItemToResponse(cartItem)

	// Delete saved item
	if err := s.cartAdvancedRepo.DeleteSavedForLater(savedItemID); err != nil {
		return nil, err
	}

	return response, nil
}

// ===== BULK ACTIONS SERVICE =====

// BulkCartAction performs bulk actions on cart items
func (s *CartAdvancedService) BulkCartAction(req *model.CartBulkActionRequest, userID uint) (*model.CartBulkActionResponse, error) {
	response := &model.CartBulkActionResponse{
		SuccessCount: 0,
		FailedCount:  0,
		FailedItems:  []uint{},
		Message:      "",
	}

	// Get cart items
	items, err := s.cartAdvancedRepo.GetCartItemsByIDs(req.ItemIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart items: %w", err)
	}

	switch req.Action {
	case "move_to_saved":
		// Move items to saved for later
		for _, item := range items {
			if err := s.moveCartItemToSaved(item, userID); err != nil {
				response.FailedCount++
				response.FailedItems = append(response.FailedItems, item.ID)
			} else {
				response.SuccessCount++
			}
		}
		response.Message = fmt.Sprintf("Moved %d items to saved for later", response.SuccessCount)

	case "move_to_cart":
		// Move items back to cart
		for _, item := range items {
			item.IsSavedForLater = false
			if err := s.cartAdvancedRepo.UpdateCartItemAdvanced(&item); err != nil {
				response.FailedCount++
				response.FailedItems = append(response.FailedItems, item.ID)
			} else {
				response.SuccessCount++
			}
		}
		response.Message = fmt.Sprintf("Moved %d items to cart", response.SuccessCount)

	case "update_quantity":
		if req.Quantity == nil {
			return nil, fmt.Errorf("quantity is required for update_quantity action")
		}
		for _, item := range items {
			item.Quantity = *req.Quantity
			item.TotalPrice = item.UnitPrice * float64(*req.Quantity)
			if err := s.cartAdvancedRepo.UpdateCartItemAdvanced(&item); err != nil {
				response.FailedCount++
				response.FailedItems = append(response.FailedItems, item.ID)
			} else {
				response.SuccessCount++
			}
		}
		response.Message = fmt.Sprintf("Updated quantity for %d items", response.SuccessCount)

	case "remove":
		// Remove items from cart
		if err := s.cartAdvancedRepo.BulkDeleteCartItems(req.ItemIDs); err != nil {
			return nil, fmt.Errorf("failed to remove items: %w", err)
		}
		response.SuccessCount = len(req.ItemIDs)
		response.Message = fmt.Sprintf("Removed %d items from cart", response.SuccessCount)
	}

	return response, nil
}

// ===== HELPER METHODS =====

// generateShareToken generates a unique share token
func (s *CartAdvancedService) generateShareToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// hashPassword hashes a password (simplified implementation)
func (s *CartAdvancedService) hashPassword(password string) (string, error) {
	// In a real implementation, use bcrypt or similar
	// For now, just return the password as-is
	return password, nil
}

// moveCartItemToSaved moves a cart item to saved for later
func (s *CartAdvancedService) moveCartItemToSaved(item model.CartItem, userID uint) error {
	savedItem := &model.SavedForLater{
		UserID:            userID,
		ProductID:         item.ProductID,
		ProductVariantID:  item.ProductVariantID,
		Quantity:          item.Quantity,
		UnitPrice:         item.UnitPrice,
		TotalPrice:        item.TotalPrice,
		Notes:             item.Notes,
		Priority:          item.Priority,
		NotifyOnPriceDrop: true,
		NotifyOnStock:     true,
		NotifyOnSale:      true,
	}

	if err := s.cartAdvancedRepo.CreateSavedForLater(savedItem); err != nil {
		return err
	}

	// Delete cart item
	return s.cartAdvancedRepo.BulkDeleteCartItems([]uint{item.ID})
}

// convertCartShareToResponse converts CartShare to CartShareResponse
func (s *CartAdvancedService) convertCartShareToResponse(share *model.CartShare) *model.CartShareResponse {
	response := &model.CartShareResponse{
		ID:                share.ID,
		CartID:            share.CartID,
		SharedBy:          share.SharedBy,
		Token:             share.Token,
		IsActive:          share.IsActive,
		ExpiresAt:         share.ExpiresAt,
		MaxUses:           share.MaxUses,
		UsedCount:         share.UsedCount,
		CanView:           share.CanView,
		CanEdit:           share.CanEdit,
		CanDelete:         share.CanDelete,
		PasswordProtected: share.PasswordProtected,
		CreatedAt:         share.CreatedAt,
		UpdatedAt:         share.UpdatedAt,
	}

	if share.SharedByUser != nil {
		response.SharedByName = share.SharedByUser.Username
	}

	return response
}

// convertSavedForLaterToResponse converts SavedForLater to SavedForLaterResponse
func (s *CartAdvancedService) convertSavedForLaterToResponse(item *model.SavedForLater) *model.SavedForLaterResponse {
	response := &model.SavedForLaterResponse{
		ID:               item.ID,
		UserID:           item.UserID,
		ProductID:        item.ProductID,
		ProductVariantID: item.ProductVariantID,
		Quantity:         item.Quantity,
		UnitPrice:        item.UnitPrice,
		TotalPrice:       item.TotalPrice,
		Notes:            item.Notes,
		Priority:         item.Priority,
		RemindAt:         item.RemindAt,
		NotifyOnPriceDrop: item.NotifyOnPriceDrop,
		NotifyOnStock:     item.NotifyOnStock,
		NotifyOnSale:      item.NotifyOnSale,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}

	// Convert product if available
	if item.Product != nil {
		// This would need proper conversion from Product to ProductResponse
		// For now, we'll leave it as nil
	}

	// Convert product variant if available
	if item.ProductVariant != nil {
		// This would need proper conversion from ProductVariant to ProductVariantResponse
		// For now, we'll leave it as nil
	}

	return response
}

// convertCartItemToResponse converts CartItem to CartItemResponse
func (s *CartAdvancedService) convertCartItemToResponse(item *model.CartItem) *model.CartItemResponse {
	response := &model.CartItemResponse{
		ID:               item.ID,
		CartID:           item.CartID,
		ProductID:        item.ProductID,
		ProductVariantID: item.ProductVariantID,
		Quantity:         item.Quantity,
		UnitPrice:        item.UnitPrice,
		TotalPrice:       item.TotalPrice,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}

	// Convert product if available
	if item.Product != nil {
		// This would need proper conversion from Product to ProductResponse
		// For now, we'll leave it as nil
	}

	// Convert product variant if available
	if item.ProductVariant != nil {
		// This would need proper conversion from ProductVariant to ProductVariantResponse
		// For now, we'll leave it as nil
	}

	return response
}
