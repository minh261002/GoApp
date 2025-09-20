package repository

import (
	"fmt"
	"go_app/internal/model"
	"go_app/pkg/database"
	"time"

	"gorm.io/gorm"
)

// WishlistRepository defines methods for interacting with wishlist data
type WishlistRepository interface {
	// Wishlist CRUD
	CreateWishlist(wishlist *model.Wishlist) error
	GetWishlistByID(id uint) (*model.Wishlist, error)
	GetWishlistBySlug(slug string) (*model.Wishlist, error)
	GetWishlistsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.Wishlist, int64, error)
	UpdateWishlist(wishlist *model.Wishlist) error
	DeleteWishlist(id uint) error

	// Wishlist Management
	GetPublicWishlists(page, limit int, filters map[string]interface{}) ([]model.Wishlist, int64, error)
	GetDefaultWishlist(userID uint) (*model.Wishlist, error)
	SetDefaultWishlist(userID, wishlistID uint) error
	SearchWishlists(query string, page, limit int) ([]model.Wishlist, int64, error)
	GetWishlistStats() (*model.WishlistStatsResponse, error)

	// Wishlist Items
	CreateWishlistItem(item *model.WishlistItem) error
	GetWishlistItemByID(id uint) (*model.WishlistItem, error)
	GetWishlistItemsByWishlist(wishlistID uint, page, limit int, filters map[string]interface{}) ([]model.WishlistItem, int64, error)
	GetWishlistItemsByProduct(productID uint, page, limit int) ([]model.WishlistItem, int64, error)
	UpdateWishlistItem(item *model.WishlistItem) error
	DeleteWishlistItem(id uint) error
	ReorderWishlistItems(wishlistID uint, itemOrders map[uint]int) error
	MoveItemToWishlist(itemID, targetWishlistID uint) error

	// Favorites
	CreateFavorite(favorite *model.Favorite) error
	GetFavoriteByID(id uint) (*model.Favorite, error)
	GetFavoritesByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.Favorite, int64, error)
	GetFavoriteByUserAndProduct(userID, productID uint) (*model.Favorite, error)
	UpdateFavorite(favorite *model.Favorite) error
	DeleteFavorite(id uint) error
	DeleteFavoriteByUserAndProduct(userID, productID uint) error

	// Wishlist Sharing
	CreateWishlistShare(share *model.WishlistShare) error
	GetWishlistShareByToken(token string) (*model.WishlistShare, error)
	GetWishlistSharesByWishlist(wishlistID uint, page, limit int) ([]model.WishlistShare, int64, error)
	GetWishlistSharesByUser(userID uint, page, limit int) ([]model.WishlistShare, int64, error)
	UpdateWishlistShare(share *model.WishlistShare) error
	DeleteWishlistShare(id uint) error
	DeleteExpiredShares() error

	// Analytics
	TrackWishlistView(wishlistID uint, userID *uint, ip, userAgent, referrer string) error
	TrackWishlistItemView(itemID uint, userID *uint, ip, userAgent, referrer string) error
	TrackWishlistItemClick(itemID uint, userID *uint, ip, userAgent, referrer string) error
	GetWishlistAnalytics(wishlistID uint, startDate, endDate *time.Time) (map[string]interface{}, error)
	GetUserWishlistStats(userID uint) (map[string]interface{}, error)

	// Price Tracking
	UpdateWishlistItemPrices() error
	GetItemsWithPriceChanges() ([]model.WishlistItem, error)
	GetItemsForPriceNotification() ([]model.WishlistItem, error)
}

// wishlistRepository implements WishlistRepository
type wishlistRepository struct {
	db *gorm.DB
}

// NewWishlistRepository creates a new WishlistRepository
func NewWishlistRepository() WishlistRepository {
	return &wishlistRepository{
		db: database.DB,
	}
}

// Wishlist CRUD

// CreateWishlist creates a new wishlist
func (r *wishlistRepository) CreateWishlist(wishlist *model.Wishlist) error {
	return r.db.Create(wishlist).Error
}

// GetWishlistByID retrieves a wishlist by its ID
func (r *wishlistRepository) GetWishlistByID(id uint) (*model.Wishlist, error) {
	var wishlist model.Wishlist
	if err := r.db.Preload("User").Preload("Items.Product").
		First(&wishlist, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &wishlist, nil
}

// GetWishlistBySlug retrieves a wishlist by its slug
func (r *wishlistRepository) GetWishlistBySlug(slug string) (*model.Wishlist, error) {
	var wishlist model.Wishlist
	if err := r.db.Preload("User").Preload("Items.Product").
		Where("slug = ?", slug).First(&wishlist).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &wishlist, nil
}

// GetWishlistsByUser retrieves wishlists for a specific user
func (r *wishlistRepository) GetWishlistsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.Wishlist, int64, error) {
	var wishlists []model.Wishlist
	var total int64
	db := r.db.Model(&model.Wishlist{}).Where("user_id = ?", userID)

	// Apply filters
	for key, value := range filters {
		switch key {
		case "status":
			db = db.Where("status = ?", value)
		case "is_public":
			db = db.Where("is_public = ?", value)
		case "is_default":
			db = db.Where("is_default = ?", value)
		case "search":
			searchTerm := fmt.Sprintf("%%%s%%", value.(string))
			db = db.Where("name LIKE ? OR description LIKE ?", searchTerm, searchTerm)
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
	db = db.Order("sort_order ASC, created_at DESC")

	if err := db.Preload("User").Preload("Items.Product").
		Find(&wishlists).Error; err != nil {
		return nil, 0, err
	}

	return wishlists, total, nil
}

// UpdateWishlist updates an existing wishlist
func (r *wishlistRepository) UpdateWishlist(wishlist *model.Wishlist) error {
	return r.db.Save(wishlist).Error
}

// DeleteWishlist soft deletes a wishlist
func (r *wishlistRepository) DeleteWishlist(id uint) error {
	return r.db.Delete(&model.Wishlist{}, id).Error
}

// Wishlist Management

// GetPublicWishlists retrieves public wishlists
func (r *wishlistRepository) GetPublicWishlists(page, limit int, filters map[string]interface{}) ([]model.Wishlist, int64, error) {
	var wishlists []model.Wishlist
	var total int64
	db := r.db.Model(&model.Wishlist{}).Where("is_public = ? AND status = ?", true, model.WishlistStatusActive)

	// Apply filters
	for key, value := range filters {
		switch key {
		case "user_id":
			db = db.Where("user_id = ?", value)
		case "search":
			searchTerm := fmt.Sprintf("%%%s%%", value.(string))
			db = db.Where("name LIKE ? OR description LIKE ?", searchTerm, searchTerm)
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
	db = db.Order("view_count DESC, created_at DESC")

	if err := db.Preload("User").Preload("Items.Product").
		Find(&wishlists).Error; err != nil {
		return nil, 0, err
	}

	return wishlists, total, nil
}

// GetDefaultWishlist retrieves the default wishlist for a user
func (r *wishlistRepository) GetDefaultWishlist(userID uint) (*model.Wishlist, error) {
	var wishlist model.Wishlist
	if err := r.db.Preload("User").Preload("Items.Product").
		Where("user_id = ? AND is_default = ?", userID, true).First(&wishlist).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &wishlist, nil
}

// SetDefaultWishlist sets a wishlist as default for a user
func (r *wishlistRepository) SetDefaultWishlist(userID, wishlistID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Remove default from all user's wishlists
		if err := tx.Model(&model.Wishlist{}).Where("user_id = ?", userID).
			Update("is_default", false).Error; err != nil {
			return err
		}

		// Set new default
		if err := tx.Model(&model.Wishlist{}).Where("id = ? AND user_id = ?", wishlistID, userID).
			Update("is_default", true).Error; err != nil {
			return err
		}

		return nil
	})
}

// SearchWishlists performs full-text search on wishlists
func (r *wishlistRepository) SearchWishlists(query string, page, limit int) ([]model.Wishlist, int64, error) {
	var wishlists []model.Wishlist
	var total int64

	// Use MATCH AGAINST for full-text search
	db := r.db.Model(&model.Wishlist{}).
		Where("MATCH(name, description) AGAINST(? IN NATURAL LANGUAGE MODE)", query)

	// Count total records
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting by relevance and date
	db = db.Order("MATCH(name, description) AGAINST('" + query + "' IN NATURAL LANGUAGE MODE) DESC, created_at DESC")

	if err := db.Preload("User").Preload("Items.Product").
		Find(&wishlists).Error; err != nil {
		return nil, 0, err
	}

	return wishlists, total, nil
}

// GetWishlistStats retrieves wishlist statistics
func (r *wishlistRepository) GetWishlistStats() (*model.WishlistStatsResponse, error) {
	var stats model.WishlistStatsResponse
	var count int64

	// Total wishlists
	r.db.Model(&model.Wishlist{}).Count(&count)
	stats.TotalWishlists = count

	// Total items
	r.db.Model(&model.WishlistItem{}).Count(&count)
	stats.TotalItems = count

	// Total favorites
	r.db.Model(&model.Favorite{}).Count(&count)
	stats.TotalFavorites = count

	// Public wishlists
	r.db.Model(&model.Wishlist{}).Where("is_public = ?", true).Count(&count)
	stats.PublicWishlists = count

	// Private wishlists
	r.db.Model(&model.Wishlist{}).Where("is_public = ?", false).Count(&count)
	stats.PrivateWishlists = count

	// Shared wishlists
	r.db.Model(&model.WishlistShare{}).Count(&count)
	stats.SharedWishlists = count

	// Most wished products
	var mostWishedProducts []model.Product
	r.db.Model(&model.WishlistItem{}).
		Select("products.*, COUNT(wishlist_items.id) as wish_count").
		Joins("JOIN products ON wishlist_items.product_id = products.id").
		Group("products.id").
		Order("wish_count DESC").
		Limit(10).
		Find(&mostWishedProducts)

	for _, product := range mostWishedProducts {
		stats.MostWishedProducts = append(stats.MostWishedProducts, product.ToResponse())
	}

	// Top wishlists
	var topWishlists []model.Wishlist
	r.db.Model(&model.Wishlist{}).
		Order("view_count DESC").
		Limit(10).
		Preload("User").
		Find(&topWishlists)

	for _, wishlist := range topWishlists {
		stats.TopWishlists = append(stats.TopWishlists, *wishlist.ToResponse())
	}

	return &stats, nil
}

// Wishlist Items

// CreateWishlistItem creates a new wishlist item
func (r *wishlistRepository) CreateWishlistItem(item *model.WishlistItem) error {
	return r.db.Create(item).Error
}

// GetWishlistItemByID retrieves a wishlist item by its ID
func (r *wishlistRepository) GetWishlistItemByID(id uint) (*model.WishlistItem, error) {
	var item model.WishlistItem
	if err := r.db.Preload("Wishlist.User").Preload("Product").
		First(&item, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

// GetWishlistItemsByWishlist retrieves wishlist items for a specific wishlist
func (r *wishlistRepository) GetWishlistItemsByWishlist(wishlistID uint, page, limit int, filters map[string]interface{}) ([]model.WishlistItem, int64, error) {
	var items []model.WishlistItem
	var total int64
	db := r.db.Model(&model.WishlistItem{}).Where("wishlist_id = ?", wishlistID)

	// Apply filters
	for key, value := range filters {
		switch key {
		case "product_id":
			db = db.Where("product_id = ?", value)
		case "status":
			db = db.Where("status = ?", value)
		case "priority":
			db = db.Where("priority = ?", value)
		case "price_min":
			db = db.Where("current_price >= ?", value)
		case "price_max":
			db = db.Where("current_price <= ?", value)
		case "search":
			searchTerm := fmt.Sprintf("%%%s%%", value.(string))
			db = db.Joins("JOIN products ON wishlist_items.product_id = products.id").
				Where("products.name LIKE ? OR products.description LIKE ?", searchTerm, searchTerm)
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
	db = db.Order("sort_order ASC, priority DESC, created_at ASC")

	if err := db.Preload("Wishlist.User").Preload("Product").
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// GetWishlistItemsByProduct retrieves wishlist items for a specific product
func (r *wishlistRepository) GetWishlistItemsByProduct(productID uint, page, limit int) ([]model.WishlistItem, int64, error) {
	var items []model.WishlistItem
	var total int64
	db := r.db.Model(&model.WishlistItem{}).Where("product_id = ?", productID)

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

	if err := db.Preload("Wishlist.User").Preload("Product").
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// UpdateWishlistItem updates an existing wishlist item
func (r *wishlistRepository) UpdateWishlistItem(item *model.WishlistItem) error {
	return r.db.Save(item).Error
}

// DeleteWishlistItem soft deletes a wishlist item
func (r *wishlistRepository) DeleteWishlistItem(id uint) error {
	return r.db.Delete(&model.WishlistItem{}, id).Error
}

// ReorderWishlistItems reorders wishlist items
func (r *wishlistRepository) ReorderWishlistItems(wishlistID uint, itemOrders map[uint]int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for itemID, order := range itemOrders {
			if err := tx.Model(&model.WishlistItem{}).
				Where("id = ? AND wishlist_id = ?", itemID, wishlistID).
				Update("sort_order", order).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// MoveItemToWishlist moves an item to another wishlist
func (r *wishlistRepository) MoveItemToWishlist(itemID, targetWishlistID uint) error {
	return r.db.Model(&model.WishlistItem{}).
		Where("id = ?", itemID).
		Update("wishlist_id", targetWishlistID).Error
}

// Favorites

// CreateFavorite creates a new favorite
func (r *wishlistRepository) CreateFavorite(favorite *model.Favorite) error {
	return r.db.Create(favorite).Error
}

// GetFavoriteByID retrieves a favorite by its ID
func (r *wishlistRepository) GetFavoriteByID(id uint) (*model.Favorite, error) {
	var favorite model.Favorite
	if err := r.db.Preload("User").Preload("Product").
		First(&favorite, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &favorite, nil
}

// GetFavoritesByUser retrieves favorites for a specific user
func (r *wishlistRepository) GetFavoritesByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.Favorite, int64, error) {
	var favorites []model.Favorite
	var total int64
	db := r.db.Model(&model.Favorite{}).Where("user_id = ?", userID)

	// Apply filters
	for key, value := range filters {
		switch key {
		case "product_id":
			db = db.Where("product_id = ?", value)
		case "priority":
			db = db.Where("priority = ?", value)
		case "search":
			searchTerm := fmt.Sprintf("%%%s%%", value.(string))
			db = db.Joins("JOIN products ON favorites.product_id = products.id").
				Where("products.name LIKE ? OR products.description LIKE ?", searchTerm, searchTerm)
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
	db = db.Order("priority DESC, created_at DESC")

	if err := db.Preload("User").Preload("Product").
		Find(&favorites).Error; err != nil {
		return nil, 0, err
	}

	return favorites, total, nil
}

// GetFavoriteByUserAndProduct retrieves a favorite by user and product
func (r *wishlistRepository) GetFavoriteByUserAndProduct(userID, productID uint) (*model.Favorite, error) {
	var favorite model.Favorite
	if err := r.db.Preload("User").Preload("Product").
		Where("user_id = ? AND product_id = ?", userID, productID).First(&favorite).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &favorite, nil
}

// UpdateFavorite updates an existing favorite
func (r *wishlistRepository) UpdateFavorite(favorite *model.Favorite) error {
	return r.db.Save(favorite).Error
}

// DeleteFavorite soft deletes a favorite
func (r *wishlistRepository) DeleteFavorite(id uint) error {
	return r.db.Delete(&model.Favorite{}, id).Error
}

// DeleteFavoriteByUserAndProduct deletes a favorite by user and product
func (r *wishlistRepository) DeleteFavoriteByUserAndProduct(userID, productID uint) error {
	return r.db.Where("user_id = ? AND product_id = ?", userID, productID).
		Delete(&model.Favorite{}).Error
}

// Wishlist Sharing

// CreateWishlistShare creates a new wishlist share
func (r *wishlistRepository) CreateWishlistShare(share *model.WishlistShare) error {
	return r.db.Create(share).Error
}

// GetWishlistShareByToken retrieves a wishlist share by token
func (r *wishlistRepository) GetWishlistShareByToken(token string) (*model.WishlistShare, error) {
	var share model.WishlistShare
	if err := r.db.Preload("Wishlist.User").Preload("SharedByUser").Preload("SharedWithUser").
		Where("token = ? AND is_active = ?", token, true).First(&share).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &share, nil
}

// GetWishlistSharesByWishlist retrieves shares for a specific wishlist
func (r *wishlistRepository) GetWishlistSharesByWishlist(wishlistID uint, page, limit int) ([]model.WishlistShare, int64, error) {
	var shares []model.WishlistShare
	var total int64
	db := r.db.Model(&model.WishlistShare{}).Where("wishlist_id = ?", wishlistID)

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

	if err := db.Preload("Wishlist.User").Preload("SharedByUser").Preload("SharedWithUser").
		Find(&shares).Error; err != nil {
		return nil, 0, err
	}

	return shares, total, nil
}

// GetWishlistSharesByUser retrieves shares for a specific user
func (r *wishlistRepository) GetWishlistSharesByUser(userID uint, page, limit int) ([]model.WishlistShare, int64, error) {
	var shares []model.WishlistShare
	var total int64
	db := r.db.Model(&model.WishlistShare{}).Where("shared_with = ?", userID)

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

	if err := db.Preload("Wishlist.User").Preload("SharedByUser").Preload("SharedWithUser").
		Find(&shares).Error; err != nil {
		return nil, 0, err
	}

	return shares, total, nil
}

// UpdateWishlistShare updates an existing wishlist share
func (r *wishlistRepository) UpdateWishlistShare(share *model.WishlistShare) error {
	return r.db.Save(share).Error
}

// DeleteWishlistShare soft deletes a wishlist share
func (r *wishlistRepository) DeleteWishlistShare(id uint) error {
	return r.db.Delete(&model.WishlistShare{}, id).Error
}

// DeleteExpiredShares deletes expired wishlist shares
func (r *wishlistRepository) DeleteExpiredShares() error {
	now := time.Now()
	return r.db.Where("expires_at < ? AND is_active = ?", now, true).
		Delete(&model.WishlistShare{}).Error
}

// Analytics

// TrackWishlistView tracks a wishlist view
func (r *wishlistRepository) TrackWishlistView(wishlistID uint, userID *uint, ip, userAgent, referrer string) error {
	// Create view record
	view := map[string]interface{}{
		"wishlist_id": wishlistID,
		"user_id":     userID,
		"ip_address":  ip,
		"user_agent":  userAgent,
		"referrer":    referrer,
		"viewed_at":   time.Now(),
	}

	if err := r.db.Table("wishlist_views").Create(view).Error; err != nil {
		return err
	}

	// Update wishlist view count
	return r.db.Model(&model.Wishlist{}).Where("id = ?", wishlistID).
		Update("view_count", gorm.Expr("view_count + 1")).Error
}

// TrackWishlistItemView tracks a wishlist item view
func (r *wishlistRepository) TrackWishlistItemView(itemID uint, userID *uint, ip, userAgent, referrer string) error {
	// Create view record
	view := map[string]interface{}{
		"item_id":    itemID,
		"user_id":    userID,
		"ip_address": ip,
		"user_agent": userAgent,
		"referrer":   referrer,
		"viewed_at":  time.Now(),
	}

	return r.db.Table("wishlist_item_views").Create(view).Error
}

// TrackWishlistItemClick tracks a wishlist item click
func (r *wishlistRepository) TrackWishlistItemClick(itemID uint, userID *uint, ip, userAgent, referrer string) error {
	// Create click record
	click := map[string]interface{}{
		"item_id":    itemID,
		"user_id":    userID,
		"ip_address": ip,
		"user_agent": userAgent,
		"referrer":   referrer,
		"clicked_at": time.Now(),
	}

	return r.db.Table("wishlist_item_clicks").Create(click).Error
}

// GetWishlistAnalytics retrieves analytics for a specific wishlist
func (r *wishlistRepository) GetWishlistAnalytics(wishlistID uint, startDate, endDate *time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// View count
	viewQuery := r.db.Table("wishlist_views").Where("wishlist_id = ?", wishlistID)
	if startDate != nil {
		viewQuery = viewQuery.Where("viewed_at >= ?", *startDate)
	}
	if endDate != nil {
		viewQuery = viewQuery.Where("viewed_at <= ?", *endDate)
	}

	var viewCount int64
	viewQuery.Count(&viewCount)
	stats["views"] = viewCount

	// Item click count
	var clickCount int64
	clickQuery := r.db.Table("wishlist_item_clicks").
		Joins("JOIN wishlist_items ON wishlist_item_clicks.item_id = wishlist_items.id").
		Where("wishlist_items.wishlist_id = ?", wishlistID)
	if startDate != nil {
		clickQuery = clickQuery.Where("wishlist_item_clicks.clicked_at >= ?", *startDate)
	}
	if endDate != nil {
		clickQuery = clickQuery.Where("wishlist_item_clicks.clicked_at <= ?", *endDate)
	}

	clickQuery.Count(&clickCount)
	stats["clicks"] = clickCount

	// Click-through rate
	if viewCount > 0 {
		stats["ctr"] = float64(clickCount) / float64(viewCount) * 100
	} else {
		stats["ctr"] = 0.0
	}

	// Unique users
	var uniqueUsers int64
	viewQuery.Distinct("user_id").Count(&uniqueUsers)
	stats["unique_users"] = uniqueUsers

	return stats, nil
}

// GetUserWishlistStats retrieves wishlist statistics for a specific user
func (r *wishlistRepository) GetUserWishlistStats(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total wishlists
	var totalWishlists int64
	r.db.Model(&model.Wishlist{}).Where("user_id = ?", userID).Count(&totalWishlists)
	stats["total_wishlists"] = totalWishlists

	// Total items
	var totalItems int64
	r.db.Model(&model.WishlistItem{}).
		Joins("JOIN wishlists ON wishlist_items.wishlist_id = wishlists.id").
		Where("wishlists.user_id = ?", userID).Count(&totalItems)
	stats["total_items"] = totalItems

	// Total favorites
	var totalFavorites int64
	r.db.Model(&model.Favorite{}).Where("user_id = ?", userID).Count(&totalFavorites)
	stats["total_favorites"] = totalFavorites

	// Public wishlists
	var publicWishlists int64
	r.db.Model(&model.Wishlist{}).Where("user_id = ? AND is_public = ?", userID, true).Count(&publicWishlists)
	stats["public_wishlists"] = publicWishlists

	// Shared wishlists
	var sharedWishlists int64
	r.db.Model(&model.WishlistShare{}).Where("shared_with = ?", userID).Count(&sharedWishlists)
	stats["shared_wishlists"] = sharedWishlists

	return stats, nil
}

// Price Tracking

// UpdateWishlistItemPrices updates prices for all wishlist items
func (r *wishlistRepository) UpdateWishlistItemPrices() error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var items []model.WishlistItem
		if err := tx.Preload("Product").Find(&items).Error; err != nil {
			return err
		}

		for _, item := range items {
			if item.Product.RegularPrice > 0 {
				item.UpdatePrice(item.Product.RegularPrice)
				if err := tx.Save(&item).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// GetItemsWithPriceChanges retrieves items with price changes
func (r *wishlistRepository) GetItemsWithPriceChanges() ([]model.WishlistItem, error) {
	var items []model.WishlistItem
	err := r.db.Where("price_change != 0").
		Preload("Wishlist.User").Preload("Product").
		Find(&items).Error
	return items, err
}

// GetItemsForPriceNotification retrieves items that need price notifications
func (r *wishlistRepository) GetItemsForPriceNotification() ([]model.WishlistItem, error) {
	var items []model.WishlistItem
	err := r.db.Where("notify_on_price_drop = ? AND price_change < 0", true).
		Preload("Wishlist.User").Preload("Product").
		Find(&items).Error
	return items, err
}
