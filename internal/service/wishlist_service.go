package service

import (
	"errors"
	"fmt"
	"go_app/internal/model"
	"go_app/internal/repository"
	"go_app/pkg/utils"
	"time"
)

// WishlistService defines methods for wishlist business logic
type WishlistService interface {
	// Wishlist Management
	CreateWishlist(req *model.WishlistCreateRequest, userID uint) (*model.WishlistResponse, error)
	GetWishlistByID(id uint, userID *uint) (*model.WishlistResponse, error)
	GetWishlistBySlug(slug string, userID *uint) (*model.WishlistResponse, error)
	GetWishlistsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.WishlistResponse, int64, error)
	UpdateWishlist(id uint, req *model.WishlistUpdateRequest, userID uint) (*model.WishlistResponse, error)
	DeleteWishlist(id uint, userID uint) error
	SetDefaultWishlist(wishlistID, userID uint) error

	// Public Wishlists
	GetPublicWishlists(page, limit int, filters map[string]interface{}) ([]model.WishlistResponse, int64, error)
	SearchWishlists(query string, page, limit int) ([]model.WishlistResponse, int64, error)

	// Wishlist Items
	AddItemToWishlist(req *model.WishlistItemCreateRequest, userID uint) (*model.WishlistItemResponse, error)
	GetWishlistItemByID(id uint, userID *uint) (*model.WishlistItemResponse, error)
	GetWishlistItems(wishlistID uint, page, limit int, filters map[string]interface{}, userID *uint) ([]model.WishlistItemResponse, int64, error)
	UpdateWishlistItem(id uint, req *model.WishlistItemUpdateRequest, userID uint) (*model.WishlistItemResponse, error)
	DeleteWishlistItem(id uint, userID uint) error
	ReorderWishlistItems(wishlistID uint, itemOrders map[uint]int, userID uint) error
	MoveItemToWishlist(itemID, targetWishlistID, userID uint) error

	// Favorites
	AddToFavorites(req *model.FavoriteCreateRequest, userID uint) (*model.FavoriteResponse, error)
	GetFavoriteByID(id uint, userID *uint) (*model.FavoriteResponse, error)
	GetFavoritesByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.FavoriteResponse, int64, error)
	UpdateFavorite(id uint, req *model.FavoriteUpdateRequest, userID uint) (*model.FavoriteResponse, error)
	RemoveFromFavorites(id uint, userID uint) error
	RemoveFromFavoritesByProduct(productID, userID uint) error

	// Wishlist Sharing
	ShareWishlist(req *model.WishlistShareRequest, userID uint) (*model.WishlistShareResponse, error)
	GetWishlistShareByToken(token string) (*model.WishlistShareResponse, error)
	GetWishlistSharesByWishlist(wishlistID uint, page, limit int, userID uint) ([]model.WishlistShareResponse, int64, error)
	GetWishlistSharesByUser(userID uint, page, limit int) ([]model.WishlistShareResponse, int64, error)
	UpdateWishlistShare(id uint, req *model.WishlistShareRequest, userID uint) (*model.WishlistShareResponse, error)
	DeleteWishlistShare(id uint, userID uint) error

	// Analytics
	TrackWishlistView(wishlistID uint, userID *uint, ip, userAgent, referrer string) error
	TrackWishlistItemView(itemID uint, userID *uint, ip, userAgent, referrer string) error
	TrackWishlistItemClick(itemID uint, userID *uint, ip, userAgent, referrer string) error
	GetWishlistAnalytics(wishlistID uint, startDate, endDate *time.Time, userID uint) (map[string]interface{}, error)
	GetUserWishlistStats(userID uint) (map[string]interface{}, error)
	GetWishlistStats() (*model.WishlistStatsResponse, error)

	// Price Tracking
	UpdateWishlistItemPrices() error
	GetItemsWithPriceChanges() ([]model.WishlistItemResponse, error)
	GetItemsForPriceNotification() ([]model.WishlistItemResponse, error)
}

// wishlistService implements WishlistService
type wishlistService struct {
	wishlistRepo repository.WishlistRepository
	userRepo     repository.UserRepository
	productRepo  repository.ProductRepository
}

// NewWishlistService creates a new WishlistService
func NewWishlistService() WishlistService {
	return &wishlistService{
		wishlistRepo: repository.NewWishlistRepository(),
		userRepo:     repository.NewUserRepository(),
		productRepo:  *repository.NewProductRepository(),
	}
}

// Wishlist Management

// CreateWishlist creates a new wishlist
func (s *wishlistService) CreateWishlist(req *model.WishlistCreateRequest, userID uint) (*model.WishlistResponse, error) {
	// Check if user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Generate slug if not provided
	slug := utils.GenerateSlug(req.Name)
	if slug == "" {
		slug = fmt.Sprintf("wishlist-%d", time.Now().Unix())
	}

	// Check if slug is unique
	existingWishlist, err := s.wishlistRepo.GetWishlistBySlug(slug)
	if err != nil {
		return nil, err
	}
	if existingWishlist != nil {
		slug = fmt.Sprintf("%s-%d", slug, time.Now().Unix())
	}

	// If this is set as default, remove default from other wishlists
	if req.IsDefault {
		if err := s.wishlistRepo.SetDefaultWishlist(userID, 0); err != nil {
			return nil, err
		}
	}

	wishlist := &model.Wishlist{
		UserID:          userID,
		Name:            req.Name,
		Description:     req.Description,
		Status:          req.Status,
		IsDefault:       req.IsDefault,
		IsPublicFlag:    req.IsPublic,
		SortOrder:       req.SortOrder,
		Slug:            slug,
		MetaTitle:       req.MetaTitle,
		MetaDescription: req.MetaDescription,
	}

	if err := s.wishlistRepo.CreateWishlist(wishlist); err != nil {
		return nil, err
	}

	// Load the created wishlist with relationships
	createdWishlist, err := s.wishlistRepo.GetWishlistByID(wishlist.ID)
	if err != nil {
		return nil, err
	}

	return createdWishlist.ToResponse(), nil
}

// GetWishlistByID retrieves a wishlist by ID
func (s *wishlistService) GetWishlistByID(id uint, userID *uint) (*model.WishlistResponse, error) {
	wishlist, err := s.wishlistRepo.GetWishlistByID(id)
	if err != nil {
		return nil, err
	}
	if wishlist == nil {
		return nil, errors.New("wishlist not found")
	}

	// Check if user can view this wishlist
	if !s.canViewWishlist(wishlist, userID) {
		return nil, errors.New("access denied")
	}

	return wishlist.ToResponse(), nil
}

// GetWishlistBySlug retrieves a wishlist by slug
func (s *wishlistService) GetWishlistBySlug(slug string, userID *uint) (*model.WishlistResponse, error) {
	wishlist, err := s.wishlistRepo.GetWishlistBySlug(slug)
	if err != nil {
		return nil, err
	}
	if wishlist == nil {
		return nil, errors.New("wishlist not found")
	}

	// Check if user can view this wishlist
	if !s.canViewWishlist(wishlist, userID) {
		return nil, errors.New("access denied")
	}

	return wishlist.ToResponse(), nil
}

// GetWishlistsByUser retrieves wishlists for a user
func (s *wishlistService) GetWishlistsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.WishlistResponse, int64, error) {
	wishlists, total, err := s.wishlistRepo.GetWishlistsByUser(userID, page, limit, filters)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.WishlistResponse, len(wishlists))
	for i, wishlist := range wishlists {
		responses[i] = *wishlist.ToResponse()
	}

	return responses, total, nil
}

// UpdateWishlist updates a wishlist
func (s *wishlistService) UpdateWishlist(id uint, req *model.WishlistUpdateRequest, userID uint) (*model.WishlistResponse, error) {
	wishlist, err := s.wishlistRepo.GetWishlistByID(id)
	if err != nil {
		return nil, err
	}
	if wishlist == nil {
		return nil, errors.New("wishlist not found")
	}

	// Check if user owns this wishlist
	if wishlist.UserID != userID {
		return nil, errors.New("access denied")
	}

	// Update fields
	if req.Name != nil {
		wishlist.Name = *req.Name
		// Regenerate slug if name changed
		wishlist.Slug = utils.GenerateSlug(*req.Name)
	}
	if req.Description != nil {
		wishlist.Description = *req.Description
	}
	if req.Status != nil {
		wishlist.Status = *req.Status
	}
	if req.IsDefault != nil {
		if *req.IsDefault {
			// Remove default from other wishlists
			if err := s.wishlistRepo.SetDefaultWishlist(userID, 0); err != nil {
				return nil, err
			}
		}
		wishlist.IsDefault = *req.IsDefault
	}
	if req.IsPublic != nil {
		wishlist.IsPublicFlag = *req.IsPublic
	}
	if req.SortOrder != nil {
		wishlist.SortOrder = *req.SortOrder
	}
	if req.MetaTitle != nil {
		wishlist.MetaTitle = *req.MetaTitle
	}
	if req.MetaDescription != nil {
		wishlist.MetaDescription = *req.MetaDescription
	}

	if err := s.wishlistRepo.UpdateWishlist(wishlist); err != nil {
		return nil, err
	}

	return wishlist.ToResponse(), nil
}

// DeleteWishlist deletes a wishlist
func (s *wishlistService) DeleteWishlist(id uint, userID uint) error {
	wishlist, err := s.wishlistRepo.GetWishlistByID(id)
	if err != nil {
		return err
	}
	if wishlist == nil {
		return errors.New("wishlist not found")
	}

	// Check if user owns this wishlist
	if wishlist.UserID != userID {
		return errors.New("access denied")
	}

	return s.wishlistRepo.DeleteWishlist(id)
}

// SetDefaultWishlist sets a wishlist as default
func (s *wishlistService) SetDefaultWishlist(wishlistID, userID uint) error {
	wishlist, err := s.wishlistRepo.GetWishlistByID(wishlistID)
	if err != nil {
		return err
	}
	if wishlist == nil {
		return errors.New("wishlist not found")
	}

	// Check if user owns this wishlist
	if wishlist.UserID != userID {
		return errors.New("access denied")
	}

	return s.wishlistRepo.SetDefaultWishlist(userID, wishlistID)
}

// Public Wishlists

// GetPublicWishlists retrieves public wishlists
func (s *wishlistService) GetPublicWishlists(page, limit int, filters map[string]interface{}) ([]model.WishlistResponse, int64, error) {
	wishlists, total, err := s.wishlistRepo.GetPublicWishlists(page, limit, filters)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.WishlistResponse, len(wishlists))
	for i, wishlist := range wishlists {
		responses[i] = *wishlist.ToResponse()
	}

	return responses, total, nil
}

// SearchWishlists searches wishlists
func (s *wishlistService) SearchWishlists(query string, page, limit int) ([]model.WishlistResponse, int64, error) {
	wishlists, total, err := s.wishlistRepo.SearchWishlists(query, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.WishlistResponse, len(wishlists))
	for i, wishlist := range wishlists {
		responses[i] = *wishlist.ToResponse()
	}

	return responses, total, nil
}

// Wishlist Items

// AddItemToWishlist adds an item to a wishlist
func (s *wishlistService) AddItemToWishlist(req *model.WishlistItemCreateRequest, userID uint) (*model.WishlistItemResponse, error) {
	// Check if wishlist exists and user owns it
	wishlist, err := s.wishlistRepo.GetWishlistByID(req.WishlistID)
	if err != nil {
		return nil, err
	}
	if wishlist == nil {
		return nil, errors.New("wishlist not found")
	}
	if wishlist.UserID != userID {
		return nil, errors.New("access denied")
	}

	// Check if product exists
	product, err := s.productRepo.GetByID(req.ProductID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}

	// Check if item already exists in wishlist
	existingItems, _, err := s.wishlistRepo.GetWishlistItemsByWishlist(req.WishlistID, 1, 1, map[string]interface{}{
		"product_id": req.ProductID,
	})
	if err != nil {
		return nil, err
	}
	if len(existingItems) > 0 {
		return nil, errors.New("product already in wishlist")
	}

	// Set default values
	quantity := req.Quantity
	if quantity <= 0 {
		quantity = 1
	}

	notifyOnPriceDrop := true
	if req.NotifyOnPriceDrop != nil {
		notifyOnPriceDrop = *req.NotifyOnPriceDrop
	}

	notifyOnStock := true
	if req.NotifyOnStock != nil {
		notifyOnStock = *req.NotifyOnStock
	}

	notifyOnSale := true
	if req.NotifyOnSale != nil {
		notifyOnSale = *req.NotifyOnSale
	}

	item := &model.WishlistItem{
		WishlistID:        req.WishlistID,
		ProductID:         req.ProductID,
		Quantity:          quantity,
		Notes:             req.Notes,
		Priority:          req.Priority,
		SortOrder:         req.SortOrder,
		AddedPrice:        product.RegularPrice,
		CurrentPrice:      product.RegularPrice,
		NotifyOnPriceDrop: notifyOnPriceDrop,
		NotifyOnStock:     notifyOnStock,
		NotifyOnSale:      notifyOnSale,
	}

	if err := s.wishlistRepo.CreateWishlistItem(item); err != nil {
		return nil, err
	}

	// Update wishlist item count
	if err := s.updateWishlistItemCount(req.WishlistID); err != nil {
		return nil, err
	}

	// Load the created item with relationships
	createdItem, err := s.wishlistRepo.GetWishlistItemByID(item.ID)
	if err != nil {
		return nil, err
	}

	return createdItem.ToResponse(), nil
}

// GetWishlistItemByID retrieves a wishlist item by ID
func (s *wishlistService) GetWishlistItemByID(id uint, userID *uint) (*model.WishlistItemResponse, error) {
	item, err := s.wishlistRepo.GetWishlistItemByID(id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("wishlist item not found")
	}

	// Check if user can view this item
	if !s.canViewWishlistItem(item, userID) {
		return nil, errors.New("access denied")
	}

	return item.ToResponse(), nil
}

// GetWishlistItems retrieves wishlist items
func (s *wishlistService) GetWishlistItems(wishlistID uint, page, limit int, filters map[string]interface{}, userID *uint) ([]model.WishlistItemResponse, int64, error) {
	// Check if user can view this wishlist
	wishlist, err := s.wishlistRepo.GetWishlistByID(wishlistID)
	if err != nil {
		return nil, 0, err
	}
	if wishlist == nil {
		return nil, 0, errors.New("wishlist not found")
	}
	if !s.canViewWishlist(wishlist, userID) {
		return nil, 0, errors.New("access denied")
	}

	items, total, err := s.wishlistRepo.GetWishlistItemsByWishlist(wishlistID, page, limit, filters)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.WishlistItemResponse, len(items))
	for i, item := range items {
		responses[i] = *item.ToResponse()
	}

	return responses, total, nil
}

// UpdateWishlistItem updates a wishlist item
func (s *wishlistService) UpdateWishlistItem(id uint, req *model.WishlistItemUpdateRequest, userID uint) (*model.WishlistItemResponse, error) {
	item, err := s.wishlistRepo.GetWishlistItemByID(id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("wishlist item not found")
	}

	// Check if user owns this item
	if item.Wishlist.UserID != userID {
		return nil, errors.New("access denied")
	}

	// Update fields
	if req.Status != nil {
		item.Status = *req.Status
	}
	if req.Quantity != nil {
		item.Quantity = *req.Quantity
	}
	if req.Notes != nil {
		item.Notes = *req.Notes
	}
	if req.Priority != nil {
		item.Priority = *req.Priority
	}
	if req.SortOrder != nil {
		item.SortOrder = *req.SortOrder
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

	if err := s.wishlistRepo.UpdateWishlistItem(item); err != nil {
		return nil, err
	}

	return item.ToResponse(), nil
}

// DeleteWishlistItem deletes a wishlist item
func (s *wishlistService) DeleteWishlistItem(id uint, userID uint) error {
	item, err := s.wishlistRepo.GetWishlistItemByID(id)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("wishlist item not found")
	}

	// Check if user owns this item
	if item.Wishlist.UserID != userID {
		return errors.New("access denied")
	}

	if err := s.wishlistRepo.DeleteWishlistItem(id); err != nil {
		return err
	}

	// Update wishlist item count
	return s.updateWishlistItemCount(item.WishlistID)
}

// ReorderWishlistItems reorders wishlist items
func (s *wishlistService) ReorderWishlistItems(wishlistID uint, itemOrders map[uint]int, userID uint) error {
	// Check if user owns this wishlist
	wishlist, err := s.wishlistRepo.GetWishlistByID(wishlistID)
	if err != nil {
		return err
	}
	if wishlist == nil {
		return errors.New("wishlist not found")
	}
	if wishlist.UserID != userID {
		return errors.New("access denied")
	}

	return s.wishlistRepo.ReorderWishlistItems(wishlistID, itemOrders)
}

// MoveItemToWishlist moves an item to another wishlist
func (s *wishlistService) MoveItemToWishlist(itemID, targetWishlistID, userID uint) error {
	// Check if item exists and user owns it
	item, err := s.wishlistRepo.GetWishlistItemByID(itemID)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("wishlist item not found")
	}
	if item.Wishlist.UserID != userID {
		return errors.New("access denied")
	}

	// Check if target wishlist exists and user owns it
	targetWishlist, err := s.wishlistRepo.GetWishlistByID(targetWishlistID)
	if err != nil {
		return err
	}
	if targetWishlist == nil {
		return errors.New("target wishlist not found")
	}
	if targetWishlist.UserID != userID {
		return errors.New("access denied")
	}

	// Check if item already exists in target wishlist
	existingItems, _, err := s.wishlistRepo.GetWishlistItemsByWishlist(targetWishlistID, 1, 1, map[string]interface{}{
		"product_id": item.ProductID,
	})
	if err != nil {
		return err
	}
	if len(existingItems) > 0 {
		return errors.New("product already exists in target wishlist")
	}

	if err := s.wishlistRepo.MoveItemToWishlist(itemID, targetWishlistID); err != nil {
		return err
	}

	// Update item counts for both wishlists
	if err := s.updateWishlistItemCount(item.WishlistID); err != nil {
		return err
	}
	if err := s.updateWishlistItemCount(targetWishlistID); err != nil {
		return err
	}

	return nil
}

// Favorites

// AddToFavorites adds a product to favorites
func (s *wishlistService) AddToFavorites(req *model.FavoriteCreateRequest, userID uint) (*model.FavoriteResponse, error) {
	// Check if product exists
	product, err := s.productRepo.GetByID(req.ProductID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}

	// Check if already in favorites
	existing, err := s.wishlistRepo.GetFavoriteByUserAndProduct(userID, req.ProductID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("product already in favorites")
	}

	// Set default values
	notifyOnPriceDrop := true
	if req.NotifyOnPriceDrop != nil {
		notifyOnPriceDrop = *req.NotifyOnPriceDrop
	}

	notifyOnStock := true
	if req.NotifyOnStock != nil {
		notifyOnStock = *req.NotifyOnStock
	}

	notifyOnSale := true
	if req.NotifyOnSale != nil {
		notifyOnSale = *req.NotifyOnSale
	}

	favorite := &model.Favorite{
		UserID:            userID,
		ProductID:         req.ProductID,
		Notes:             req.Notes,
		Priority:          req.Priority,
		NotifyOnPriceDrop: notifyOnPriceDrop,
		NotifyOnStock:     notifyOnStock,
		NotifyOnSale:      notifyOnSale,
	}

	if err := s.wishlistRepo.CreateFavorite(favorite); err != nil {
		return nil, err
	}

	// Load the created favorite with relationships
	createdFavorite, err := s.wishlistRepo.GetFavoriteByID(favorite.ID)
	if err != nil {
		return nil, err
	}

	return createdFavorite.ToResponse(), nil
}

// GetFavoriteByID retrieves a favorite by ID
func (s *wishlistService) GetFavoriteByID(id uint, userID *uint) (*model.FavoriteResponse, error) {
	favorite, err := s.wishlistRepo.GetFavoriteByID(id)
	if err != nil {
		return nil, err
	}
	if favorite == nil {
		return nil, errors.New("favorite not found")
	}

	// Check if user can view this favorite
	if userID != nil && favorite.UserID != *userID {
		return nil, errors.New("access denied")
	}

	return favorite.ToResponse(), nil
}

// GetFavoritesByUser retrieves favorites for a user
func (s *wishlistService) GetFavoritesByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.FavoriteResponse, int64, error) {
	favorites, total, err := s.wishlistRepo.GetFavoritesByUser(userID, page, limit, filters)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.FavoriteResponse, len(favorites))
	for i, favorite := range favorites {
		responses[i] = *favorite.ToResponse()
	}

	return responses, total, nil
}

// UpdateFavorite updates a favorite
func (s *wishlistService) UpdateFavorite(id uint, req *model.FavoriteUpdateRequest, userID uint) (*model.FavoriteResponse, error) {
	favorite, err := s.wishlistRepo.GetFavoriteByID(id)
	if err != nil {
		return nil, err
	}
	if favorite == nil {
		return nil, errors.New("favorite not found")
	}

	// Check if user owns this favorite
	if favorite.UserID != userID {
		return nil, errors.New("access denied")
	}

	// Update fields
	if req.Notes != nil {
		favorite.Notes = *req.Notes
	}
	if req.Priority != nil {
		favorite.Priority = *req.Priority
	}
	if req.NotifyOnPriceDrop != nil {
		favorite.NotifyOnPriceDrop = *req.NotifyOnPriceDrop
	}
	if req.NotifyOnStock != nil {
		favorite.NotifyOnStock = *req.NotifyOnStock
	}
	if req.NotifyOnSale != nil {
		favorite.NotifyOnSale = *req.NotifyOnSale
	}

	if err := s.wishlistRepo.UpdateFavorite(favorite); err != nil {
		return nil, err
	}

	return favorite.ToResponse(), nil
}

// RemoveFromFavorites removes a favorite
func (s *wishlistService) RemoveFromFavorites(id uint, userID uint) error {
	favorite, err := s.wishlistRepo.GetFavoriteByID(id)
	if err != nil {
		return err
	}
	if favorite == nil {
		return errors.New("favorite not found")
	}

	// Check if user owns this favorite
	if favorite.UserID != userID {
		return errors.New("access denied")
	}

	return s.wishlistRepo.DeleteFavorite(id)
}

// RemoveFromFavoritesByProduct removes a favorite by product
func (s *wishlistService) RemoveFromFavoritesByProduct(productID, userID uint) error {
	return s.wishlistRepo.DeleteFavoriteByUserAndProduct(userID, productID)
}

// Wishlist Sharing

// ShareWishlist shares a wishlist with another user
func (s *wishlistService) ShareWishlist(req *model.WishlistShareRequest, userID uint) (*model.WishlistShareResponse, error) {
	// Check if wishlist exists and user owns it
	wishlist, err := s.wishlistRepo.GetWishlistByID(req.WishlistID)
	if err != nil {
		return nil, err
	}
	if wishlist == nil {
		return nil, errors.New("wishlist not found")
	}
	if wishlist.UserID != userID {
		return nil, errors.New("access denied")
	}

	// Check if shared_with user exists
	sharedWithUser, err := s.userRepo.GetByID(req.SharedWith)
	if err != nil {
		return nil, err
	}
	if sharedWithUser == nil {
		return nil, errors.New("shared with user not found")
	}

	// Generate share token
	token, _ := utils.GenerateRandomString(32)

	// Set expiry date (default 30 days)
	expiresAt := time.Now().AddDate(0, 0, 30)
	if req.ExpiresAt != nil {
		expiresAt = *req.ExpiresAt
	}

	share := &model.WishlistShare{
		WishlistID: req.WishlistID,
		SharedBy:   userID,
		SharedWith: req.SharedWith,
		Token:      token,
		IsActive:   true,
		ExpiresAt:  expiresAt,
		CanView:    true,
		CanEdit:    req.CanEdit,
		CanDelete:  req.CanDelete,
	}

	if err := s.wishlistRepo.CreateWishlistShare(share); err != nil {
		return nil, err
	}

	// Load the created share with relationships
	createdShare, err := s.wishlistRepo.GetWishlistShareByToken(token)
	if err != nil {
		return nil, err
	}

	return createdShare.ToResponse(), nil
}

// GetWishlistShareByToken retrieves a wishlist share by token
func (s *wishlistService) GetWishlistShareByToken(token string) (*model.WishlistShareResponse, error) {
	share, err := s.wishlistRepo.GetWishlistShareByToken(token)
	if err != nil {
		return nil, err
	}
	if share == nil {
		return nil, errors.New("wishlist share not found")
	}

	// Check if share is expired
	if share.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("wishlist share has expired")
	}

	return share.ToResponse(), nil
}

// GetWishlistSharesByWishlist retrieves shares for a wishlist
func (s *wishlistService) GetWishlistSharesByWishlist(wishlistID uint, page, limit int, userID uint) ([]model.WishlistShareResponse, int64, error) {
	// Check if user owns this wishlist
	wishlist, err := s.wishlistRepo.GetWishlistByID(wishlistID)
	if err != nil {
		return nil, 0, err
	}
	if wishlist == nil {
		return nil, 0, errors.New("wishlist not found")
	}
	if wishlist.UserID != userID {
		return nil, 0, errors.New("access denied")
	}

	shares, total, err := s.wishlistRepo.GetWishlistSharesByWishlist(wishlistID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.WishlistShareResponse, len(shares))
	for i, share := range shares {
		responses[i] = *share.ToResponse()
	}

	return responses, total, nil
}

// GetWishlistSharesByUser retrieves shares for a user
func (s *wishlistService) GetWishlistSharesByUser(userID uint, page, limit int) ([]model.WishlistShareResponse, int64, error) {
	shares, total, err := s.wishlistRepo.GetWishlistSharesByUser(userID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.WishlistShareResponse, len(shares))
	for i, share := range shares {
		responses[i] = *share.ToResponse()
	}

	return responses, total, nil
}

// UpdateWishlistShare updates a wishlist share
func (s *wishlistService) UpdateWishlistShare(id uint, req *model.WishlistShareRequest, userID uint) (*model.WishlistShareResponse, error) {
	// This would require a GetWishlistShareByID method in repository
	// For now, return not implemented
	return nil, errors.New("not implemented")
}

// DeleteWishlistShare deletes a wishlist share
func (s *wishlistService) DeleteWishlistShare(id uint, userID uint) error {
	// This would require a GetWishlistShareByID method in repository
	// For now, return not implemented
	return errors.New("not implemented")
}

// Analytics

// TrackWishlistView tracks a wishlist view
func (s *wishlistService) TrackWishlistView(wishlistID uint, userID *uint, ip, userAgent, referrer string) error {
	return s.wishlistRepo.TrackWishlistView(wishlistID, userID, ip, userAgent, referrer)
}

// TrackWishlistItemView tracks a wishlist item view
func (s *wishlistService) TrackWishlistItemView(itemID uint, userID *uint, ip, userAgent, referrer string) error {
	return s.wishlistRepo.TrackWishlistItemView(itemID, userID, ip, userAgent, referrer)
}

// TrackWishlistItemClick tracks a wishlist item click
func (s *wishlistService) TrackWishlistItemClick(itemID uint, userID *uint, ip, userAgent, referrer string) error {
	return s.wishlistRepo.TrackWishlistItemClick(itemID, userID, ip, userAgent, referrer)
}

// GetWishlistAnalytics retrieves analytics for a wishlist
func (s *wishlistService) GetWishlistAnalytics(wishlistID uint, startDate, endDate *time.Time, userID uint) (map[string]interface{}, error) {
	// Check if user owns this wishlist
	wishlist, err := s.wishlistRepo.GetWishlistByID(wishlistID)
	if err != nil {
		return nil, err
	}
	if wishlist == nil {
		return nil, errors.New("wishlist not found")
	}
	if wishlist.UserID != userID {
		return nil, errors.New("access denied")
	}

	return s.wishlistRepo.GetWishlistAnalytics(wishlistID, startDate, endDate)
}

// GetUserWishlistStats retrieves wishlist statistics for a user
func (s *wishlistService) GetUserWishlistStats(userID uint) (map[string]interface{}, error) {
	return s.wishlistRepo.GetUserWishlistStats(userID)
}

// GetWishlistStats retrieves overall wishlist statistics
func (s *wishlistService) GetWishlistStats() (*model.WishlistStatsResponse, error) {
	return s.wishlistRepo.GetWishlistStats()
}

// Price Tracking

// UpdateWishlistItemPrices updates prices for all wishlist items
func (s *wishlistService) UpdateWishlistItemPrices() error {
	return s.wishlistRepo.UpdateWishlistItemPrices()
}

// GetItemsWithPriceChanges retrieves items with price changes
func (s *wishlistService) GetItemsWithPriceChanges() ([]model.WishlistItemResponse, error) {
	items, err := s.wishlistRepo.GetItemsWithPriceChanges()
	if err != nil {
		return nil, err
	}

	responses := make([]model.WishlistItemResponse, len(items))
	for i, item := range items {
		responses[i] = *item.ToResponse()
	}

	return responses, nil
}

// GetItemsForPriceNotification retrieves items that need price notifications
func (s *wishlistService) GetItemsForPriceNotification() ([]model.WishlistItemResponse, error) {
	items, err := s.wishlistRepo.GetItemsForPriceNotification()
	if err != nil {
		return nil, err
	}

	responses := make([]model.WishlistItemResponse, len(items))
	for i, item := range items {
		responses[i] = *item.ToResponse()
	}

	return responses, nil
}

// Helper methods

// canViewWishlist checks if a user can view a wishlist
func (s *wishlistService) canViewWishlist(wishlist *model.Wishlist, userID *uint) bool {
	// Public wishlists can be viewed by anyone
	if wishlist.IsPublic() {
		return true
	}

	// Private wishlists can only be viewed by the owner
	if userID != nil && wishlist.UserID == *userID {
		return true
	}

	return false
}

// canViewWishlistItem checks if a user can view a wishlist item
func (s *wishlistService) canViewWishlistItem(item *model.WishlistItem, userID *uint) bool {
	return s.canViewWishlist(&item.Wishlist, userID)
}

// updateWishlistItemCount updates the item count for a wishlist
func (s *wishlistService) updateWishlistItemCount(wishlistID uint) error {
	_, count, err := s.wishlistRepo.GetWishlistItemsByWishlist(wishlistID, 1, 1, map[string]interface{}{})
	if err != nil {
		return err
	}

	// Update wishlist item count
	return s.wishlistRepo.UpdateWishlist(&model.Wishlist{
		ID:        wishlistID,
		ItemCount: count,
	})
}
