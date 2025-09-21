package repository

import (
	"fmt"
	"go_app/internal/model"
	"go_app/pkg/database"
	"time"

	"gorm.io/gorm"
)

type CartAdvancedRepository struct {
	db *gorm.DB
}

func NewCartAdvancedRepository() *CartAdvancedRepository {
	return &CartAdvancedRepository{
		db: database.GetDB(),
	}
}

// ===== CART SHARE REPOSITORY =====

// CreateCartShare creates a new cart share
func (r *CartAdvancedRepository) CreateCartShare(share *model.CartShare) error {
	if err := r.db.Create(share).Error; err != nil {
		return fmt.Errorf("failed to create cart share: %w", err)
	}
	return nil
}

// GetCartShareByToken gets a cart share by token
func (r *CartAdvancedRepository) GetCartShareByToken(token string) (*model.CartShare, error) {
	var share model.CartShare
	if err := r.db.Preload("Cart").Preload("SharedByUser").Where("token = ? AND is_active = ? AND expires_at > ?", token, true, time.Now()).First(&share).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("cart share not found or expired")
		}
		return nil, fmt.Errorf("failed to get cart share: %w", err)
	}
	return &share, nil
}

// GetCartSharesByCartID gets all shares for a cart
func (r *CartAdvancedRepository) GetCartSharesByCartID(cartID uint) ([]model.CartShare, error) {
	var shares []model.CartShare
	if err := r.db.Preload("SharedByUser").Where("cart_id = ?", cartID).Find(&shares).Error; err != nil {
		return nil, fmt.Errorf("failed to get cart shares: %w", err)
	}
	return shares, nil
}

// GetCartSharesByUser gets all shares created by a user
func (r *CartAdvancedRepository) GetCartSharesByUser(userID uint) ([]model.CartShare, error) {
	var shares []model.CartShare
	if err := r.db.Preload("Cart").Where("shared_by = ?", userID).Find(&shares).Error; err != nil {
		return nil, fmt.Errorf("failed to get user cart shares: %w", err)
	}
	return shares, nil
}

// UpdateCartShare updates a cart share
func (r *CartAdvancedRepository) UpdateCartShare(share *model.CartShare) error {
	if err := r.db.Save(share).Error; err != nil {
		return fmt.Errorf("failed to update cart share: %w", err)
	}
	return nil
}

// DeleteCartShare deletes a cart share
func (r *CartAdvancedRepository) DeleteCartShare(id uint) error {
	if err := r.db.Delete(&model.CartShare{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete cart share: %w", err)
	}
	return nil
}

// IncrementCartShareUsage increments the usage count of a cart share
func (r *CartAdvancedRepository) IncrementCartShareUsage(token string) error {
	if err := r.db.Model(&model.CartShare{}).Where("token = ?", token).Update("used_count", gorm.Expr("used_count + 1")).Error; err != nil {
		return fmt.Errorf("failed to increment cart share usage: %w", err)
	}
	return nil
}

// ===== SAVED FOR LATER REPOSITORY =====

// CreateSavedForLater creates a new saved item
func (r *CartAdvancedRepository) CreateSavedForLater(item *model.SavedForLater) error {
	if err := r.db.Create(item).Error; err != nil {
		return fmt.Errorf("failed to create saved for later item: %w", err)
	}
	return nil
}

// GetSavedForLaterByID gets a saved item by ID
func (r *CartAdvancedRepository) GetSavedForLaterByID(id uint) (*model.SavedForLater, error) {
	var item model.SavedForLater
	if err := r.db.Preload("Product").Preload("ProductVariant").Where("id = ?", id).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("saved for later item not found")
		}
		return nil, fmt.Errorf("failed to get saved for later item: %w", err)
	}
	return &item, nil
}

// GetSavedForLaterByUser gets all saved items for a user
func (r *CartAdvancedRepository) GetSavedForLaterByUser(userID uint, page, limit int) ([]model.SavedForLater, int64, error) {
	var items []model.SavedForLater
	var total int64

	// Count total
	if err := r.db.Model(&model.SavedForLater{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count saved for later items: %w", err)
	}

	// Get items with pagination
	offset := (page - 1) * limit
	if err := r.db.Preload("Product").Preload("ProductVariant").
		Where("user_id = ?", userID).
		Order("priority DESC, created_at DESC").
		Offset(offset).Limit(limit).
		Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get saved for later items: %w", err)
	}

	return items, total, nil
}

// UpdateSavedForLater updates a saved item
func (r *CartAdvancedRepository) UpdateSavedForLater(item *model.SavedForLater) error {
	if err := r.db.Save(item).Error; err != nil {
		return fmt.Errorf("failed to update saved for later item: %w", err)
	}
	return nil
}

// DeleteSavedForLater deletes a saved item
func (r *CartAdvancedRepository) DeleteSavedForLater(id uint) error {
	if err := r.db.Delete(&model.SavedForLater{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete saved for later item: %w", err)
	}
	return nil
}

// GetSavedForLaterByProduct gets saved items by product
func (r *CartAdvancedRepository) GetSavedForLaterByProduct(userID, productID uint, variantID *uint) (*model.SavedForLater, error) {
	var item model.SavedForLater
	query := r.db.Where("user_id = ? AND product_id = ?", userID, productID)
	
	if variantID != nil {
		query = query.Where("product_variant_id = ?", *variantID)
	} else {
		query = query.Where("product_variant_id IS NULL")
	}

	if err := query.Preload("Product").Preload("ProductVariant").First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("saved for later item not found")
		}
		return nil, fmt.Errorf("failed to get saved for later item: %w", err)
	}
	return &item, nil
}

// GetSavedForLaterReminders gets items with reminders due
func (r *CartAdvancedRepository) GetSavedForLaterReminders(before time.Time) ([]model.SavedForLater, error) {
	var items []model.SavedForLater
	if err := r.db.Preload("Product").Preload("ProductVariant").
		Where("remind_at IS NOT NULL AND remind_at <= ?", before).
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to get saved for later reminders: %w", err)
	}
	return items, nil
}

// ===== CART ITEM ADVANCED REPOSITORY =====

// UpdateCartItemAdvanced updates cart item with advanced features
func (r *CartAdvancedRepository) UpdateCartItemAdvanced(item *model.CartItem) error {
	if err := r.db.Save(item).Error; err != nil {
		return fmt.Errorf("failed to update cart item: %w", err)
	}
	return nil
}

// GetCartItemsByCartID gets all cart items for a cart
func (r *CartAdvancedRepository) GetCartItemsByCartID(cartID uint) ([]model.CartItem, error) {
	var items []model.CartItem
	if err := r.db.Preload("Product").Preload("ProductVariant").
		Where("cart_id = ?", cartID).
		Order("priority DESC, created_at ASC").
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to get cart items: %w", err)
	}
	return items, nil
}

// GetCartItemsByIDs gets cart items by IDs
func (r *CartAdvancedRepository) GetCartItemsByIDs(ids []uint) ([]model.CartItem, error) {
	var items []model.CartItem
	if err := r.db.Preload("Product").Preload("ProductVariant").
		Where("id IN ?", ids).
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to get cart items by IDs: %w", err)
	}
	return items, nil
}

// BulkUpdateCartItems updates multiple cart items
func (r *CartAdvancedRepository) BulkUpdateCartItems(items []model.CartItem) error {
	if len(items) == 0 {
		return nil
	}

	if err := r.db.Save(items).Error; err != nil {
		return fmt.Errorf("failed to bulk update cart items: %w", err)
	}
	return nil
}

// BulkDeleteCartItems deletes multiple cart items
func (r *CartAdvancedRepository) BulkDeleteCartItems(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}

	if err := r.db.Delete(&model.CartItem{}, ids).Error; err != nil {
		return fmt.Errorf("failed to bulk delete cart items: %w", err)
	}
	return nil
}

// ===== STATISTICS =====

// GetCartShareStats gets cart share statistics
func (r *CartAdvancedRepository) GetCartShareStats() (map[string]interface{}, error) {
	var stats map[string]interface{} = make(map[string]interface{})

	// Total shares
	var totalShares int64
	if err := r.db.Model(&model.CartShare{}).Count(&totalShares).Error; err != nil {
		return nil, fmt.Errorf("failed to count total shares: %w", err)
	}
	stats["total_shares"] = totalShares

	// Active shares
	var activeShares int64
	if err := r.db.Model(&model.CartShare{}).Where("is_active = ? AND expires_at > ?", true, time.Now()).Count(&activeShares).Error; err != nil {
		return nil, fmt.Errorf("failed to count active shares: %w", err)
	}
	stats["active_shares"] = activeShares

	// Expired shares
	var expiredShares int64
	if err := r.db.Model(&model.CartShare{}).Where("expires_at <= ?", time.Now()).Count(&expiredShares).Error; err != nil {
		return nil, fmt.Errorf("failed to count expired shares: %w", err)
	}
	stats["expired_shares"] = expiredShares

	return stats, nil
}

// GetSavedForLaterStats gets saved for later statistics
func (r *CartAdvancedRepository) GetSavedForLaterStats() (map[string]interface{}, error) {
	var stats map[string]interface{} = make(map[string]interface{})

	// Total saved items
	var totalItems int64
	if err := r.db.Model(&model.SavedForLater{}).Count(&totalItems).Error; err != nil {
		return nil, fmt.Errorf("failed to count total saved items: %w", err)
	}
	stats["total_items"] = totalItems

	// Items with reminders
	var reminderItems int64
	if err := r.db.Model(&model.SavedForLater{}).Where("remind_at IS NOT NULL").Count(&reminderItems).Error; err != nil {
		return nil, fmt.Errorf("failed to count reminder items: %w", err)
	}
	stats["reminder_items"] = reminderItems

	// Items due for reminders
	var dueReminders int64
	if err := r.db.Model(&model.SavedForLater{}).Where("remind_at IS NOT NULL AND remind_at <= ?", time.Now()).Count(&dueReminders).Error; err != nil {
		return nil, fmt.Errorf("failed to count due reminders: %w", err)
	}
	stats["due_reminders"] = dueReminders

	return stats, nil
}
