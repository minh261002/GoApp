package model

import (
	"time"
)

// WishlistStatus represents the status of a wishlist
type WishlistStatus string

const (
	WishlistStatusActive   WishlistStatus = "active"
	WishlistStatusInactive WishlistStatus = "inactive"
	WishlistStatusPrivate  WishlistStatus = "private"
	WishlistStatusPublic   WishlistStatus = "public"
)

// WishlistItemStatus represents the status of a wishlist item
type WishlistItemStatus string

const (
	WishlistItemStatusActive    WishlistItemStatus = "active"
	WishlistItemStatusInactive  WishlistItemStatus = "inactive"
	WishlistItemStatusPurchased WishlistItemStatus = "purchased"
	WishlistItemStatusRemoved   WishlistItemStatus = "removed"
)

// Wishlist represents a user's wishlist
type Wishlist struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	UserID       uint           `json:"user_id" gorm:"not null"`
	Name         string         `json:"name" gorm:"size:255;not null"`
	Description  string         `json:"description" gorm:"type:text"`
	Status       WishlistStatus `json:"status" gorm:"size:50;not null;default:'active'"`
	IsDefault    bool           `json:"is_default" gorm:"default:false"`
	IsPublicFlag bool           `json:"is_public" gorm:"default:false"`
	SortOrder    int            `json:"sort_order" gorm:"default:0"`

	// Analytics
	ViewCount  int64 `json:"view_count" gorm:"default:0"`
	ShareCount int64 `json:"share_count" gorm:"default:0"`
	ItemCount  int64 `json:"item_count" gorm:"default:0"`

	// SEO
	Slug            string `json:"slug" gorm:"size:255;uniqueIndex"`
	MetaTitle       string `json:"meta_title" gorm:"size:255"`
	MetaDescription string `json:"meta_description" gorm:"type:text"`

	// Relationships
	User  User           `json:"user" gorm:"foreignKey:UserID"`
	Items []WishlistItem `json:"items" gorm:"foreignKey:WishlistID"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

// WishlistItem represents an item in a wishlist
type WishlistItem struct {
	ID         uint               `json:"id" gorm:"primaryKey"`
	WishlistID uint               `json:"wishlist_id" gorm:"not null"`
	ProductID  uint               `json:"product_id" gorm:"not null"`
	Status     WishlistItemStatus `json:"status" gorm:"size:50;not null;default:'active'"`
	Quantity   int                `json:"quantity" gorm:"default:1"`
	Notes      string             `json:"notes" gorm:"type:text"`
	Priority   int                `json:"priority" gorm:"default:0"` // 0 = normal, 1 = high, 2 = urgent
	SortOrder  int                `json:"sort_order" gorm:"default:0"`

	// Price tracking
	AddedPrice         float64 `json:"added_price" gorm:"type:decimal(10,2)"`         // Price when added
	CurrentPrice       float64 `json:"current_price" gorm:"type:decimal(10,2)"`       // Current price
	PriceChange        float64 `json:"price_change" gorm:"type:decimal(10,2)"`        // Price change amount
	PriceChangePercent float64 `json:"price_change_percent" gorm:"type:decimal(5,2)"` // Price change percentage

	// Notifications
	NotifyOnPriceDrop bool `json:"notify_on_price_drop" gorm:"default:true"`
	NotifyOnStock     bool `json:"notify_on_stock" gorm:"default:true"`
	NotifyOnSale      bool `json:"notify_on_sale" gorm:"default:true"`

	// Relationships
	Wishlist Wishlist `json:"wishlist" gorm:"foreignKey:WishlistID"`
	Product  Product  `json:"product" gorm:"foreignKey:ProductID"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

// Favorite represents a user's favorite product
type Favorite struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	UserID    uint   `json:"user_id" gorm:"not null"`
	ProductID uint   `json:"product_id" gorm:"not null"`
	Notes     string `json:"notes" gorm:"type:text"`
	Priority  int    `json:"priority" gorm:"default:0"`

	// Notifications
	NotifyOnPriceDrop bool `json:"notify_on_price_drop" gorm:"default:true"`
	NotifyOnStock     bool `json:"notify_on_stock" gorm:"default:true"`
	NotifyOnSale      bool `json:"notify_on_sale" gorm:"default:true"`

	// Relationships
	User    User    `json:"user" gorm:"foreignKey:UserID"`
	Product Product `json:"product" gorm:"foreignKey:ProductID"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`

	// Unique constraint
	// UNIQUE(user_id, product_id)
}

// WishlistShare represents a shared wishlist
type WishlistShare struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	WishlistID uint      `json:"wishlist_id" gorm:"not null"`
	SharedBy   uint      `json:"shared_by" gorm:"not null"`
	SharedWith uint      `json:"shared_with" gorm:"not null"`
	Token      string    `json:"token" gorm:"size:255;uniqueIndex"`
	IsActive   bool      `json:"is_active" gorm:"default:true"`
	ExpiresAt  time.Time `json:"expires_at"`

	// Permissions
	CanView   bool `json:"can_view" gorm:"default:true"`
	CanEdit   bool `json:"can_edit" gorm:"default:false"`
	CanDelete bool `json:"can_delete" gorm:"default:false"`

	// Relationships
	Wishlist       Wishlist `json:"wishlist" gorm:"foreignKey:WishlistID"`
	SharedByUser   User     `json:"shared_by_user" gorm:"foreignKey:SharedBy"`
	SharedWithUser User     `json:"shared_with_user" gorm:"foreignKey:SharedWith"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

// Request/Response DTOs

// WishlistCreateRequest represents the request to create a wishlist
type WishlistCreateRequest struct {
	Name            string         `json:"name" binding:"required,min=1,max=255"`
	Description     string         `json:"description" binding:"omitempty,max=1000"`
	Status          WishlistStatus `json:"status" binding:"omitempty,oneof=active inactive private public"`
	IsDefault       bool           `json:"is_default" binding:"omitempty"`
	IsPublic        bool           `json:"is_public" binding:"omitempty"`
	SortOrder       int            `json:"sort_order" binding:"omitempty,min=0"`
	MetaTitle       string         `json:"meta_title" binding:"omitempty,max=255"`
	MetaDescription string         `json:"meta_description" binding:"omitempty,max=500"`
}

// WishlistUpdateRequest represents the request to update a wishlist
type WishlistUpdateRequest struct {
	Name            *string         `json:"name" binding:"omitempty,min=1,max=255"`
	Description     *string         `json:"description" binding:"omitempty,max=1000"`
	Status          *WishlistStatus `json:"status" binding:"omitempty,oneof=active inactive private public"`
	IsDefault       *bool           `json:"is_default" binding:"omitempty"`
	IsPublic        *bool           `json:"is_public" binding:"omitempty"`
	SortOrder       *int            `json:"sort_order" binding:"omitempty,min=0"`
	MetaTitle       *string         `json:"meta_title" binding:"omitempty,max=255"`
	MetaDescription *string         `json:"meta_description" binding:"omitempty,max=500"`
}

// WishlistResponse represents the response for a wishlist
type WishlistResponse struct {
	ID              uint                   `json:"id"`
	UserID          uint                   `json:"user_id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Status          WishlistStatus         `json:"status"`
	IsDefault       bool                   `json:"is_default"`
	IsPublic        bool                   `json:"is_public"`
	SortOrder       int                    `json:"sort_order"`
	ViewCount       int64                  `json:"view_count"`
	ShareCount      int64                  `json:"share_count"`
	ItemCount       int64                  `json:"item_count"`
	Slug            string                 `json:"slug"`
	MetaTitle       string                 `json:"meta_title"`
	MetaDescription string                 `json:"meta_description"`
	User            User                   `json:"user"`
	Items           []WishlistItemResponse `json:"items"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	DeletedAt       *time.Time             `json:"deleted_at"`
}

// WishlistItemCreateRequest represents the request to add an item to wishlist
type WishlistItemCreateRequest struct {
	WishlistID uint   `json:"wishlist_id" binding:"required"`
	ProductID  uint   `json:"product_id" binding:"required"`
	Quantity   int    `json:"quantity" binding:"omitempty,min=1,max=999"`
	Notes      string `json:"notes" binding:"omitempty,max=500"`
	Priority   int    `json:"priority" binding:"omitempty,min=0,max=2"`
	SortOrder  int    `json:"sort_order" binding:"omitempty,min=0"`

	// Notifications
	NotifyOnPriceDrop *bool `json:"notify_on_price_drop" binding:"omitempty"`
	NotifyOnStock     *bool `json:"notify_on_stock" binding:"omitempty"`
	NotifyOnSale      *bool `json:"notify_on_sale" binding:"omitempty"`
}

// WishlistItemUpdateRequest represents the request to update a wishlist item
type WishlistItemUpdateRequest struct {
	Status    *WishlistItemStatus `json:"status" binding:"omitempty,oneof=active inactive purchased removed"`
	Quantity  *int                `json:"quantity" binding:"omitempty,min=1,max=999"`
	Notes     *string             `json:"notes" binding:"omitempty,max=500"`
	Priority  *int                `json:"priority" binding:"omitempty,min=0,max=2"`
	SortOrder *int                `json:"sort_order" binding:"omitempty,min=0"`

	// Notifications
	NotifyOnPriceDrop *bool `json:"notify_on_price_drop" binding:"omitempty"`
	NotifyOnStock     *bool `json:"notify_on_stock" binding:"omitempty"`
	NotifyOnSale      *bool `json:"notify_on_sale" binding:"omitempty"`
}

// WishlistItemResponse represents the response for a wishlist item
type WishlistItemResponse struct {
	ID                 uint               `json:"id"`
	WishlistID         uint               `json:"wishlist_id"`
	ProductID          uint               `json:"product_id"`
	Status             WishlistItemStatus `json:"status"`
	Quantity           int                `json:"quantity"`
	Notes              string             `json:"notes"`
	Priority           int                `json:"priority"`
	SortOrder          int                `json:"sort_order"`
	AddedPrice         float64            `json:"added_price"`
	CurrentPrice       float64            `json:"current_price"`
	PriceChange        float64            `json:"price_change"`
	PriceChangePercent float64            `json:"price_change_percent"`
	NotifyOnPriceDrop  bool               `json:"notify_on_price_drop"`
	NotifyOnStock      bool               `json:"notify_on_stock"`
	NotifyOnSale       bool               `json:"notify_on_sale"`
	Wishlist           WishlistResponse   `json:"wishlist"`
	Product            ProductResponse    `json:"product"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
	DeletedAt          *time.Time         `json:"deleted_at"`
}

// FavoriteCreateRequest represents the request to add a favorite
type FavoriteCreateRequest struct {
	ProductID uint   `json:"product_id" binding:"required"`
	Notes     string `json:"notes" binding:"omitempty,max=500"`
	Priority  int    `json:"priority" binding:"omitempty,min=0,max=2"`

	// Notifications
	NotifyOnPriceDrop *bool `json:"notify_on_price_drop" binding:"omitempty"`
	NotifyOnStock     *bool `json:"notify_on_stock" binding:"omitempty"`
	NotifyOnSale      *bool `json:"notify_on_sale" binding:"omitempty"`
}

// FavoriteUpdateRequest represents the request to update a favorite
type FavoriteUpdateRequest struct {
	Notes    *string `json:"notes" binding:"omitempty,max=500"`
	Priority *int    `json:"priority" binding:"omitempty,min=0,max=2"`

	// Notifications
	NotifyOnPriceDrop *bool `json:"notify_on_price_drop" binding:"omitempty"`
	NotifyOnStock     *bool `json:"notify_on_stock" binding:"omitempty"`
	NotifyOnSale      *bool `json:"notify_on_sale" binding:"omitempty"`
}

// FavoriteResponse represents the response for a favorite
type FavoriteResponse struct {
	ID                uint            `json:"id"`
	UserID            uint            `json:"user_id"`
	ProductID         uint            `json:"product_id"`
	Notes             string          `json:"notes"`
	Priority          int             `json:"priority"`
	NotifyOnPriceDrop bool            `json:"notify_on_price_drop"`
	NotifyOnStock     bool            `json:"notify_on_stock"`
	NotifyOnSale      bool            `json:"notify_on_sale"`
	User              User            `json:"user"`
	Product           ProductResponse `json:"product"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	DeletedAt         *time.Time      `json:"deleted_at"`
}

// WishlistShareRequest represents the request to share a wishlist
type WishlistShareRequest struct {
	WishlistID uint       `json:"wishlist_id" binding:"required"`
	SharedWith uint       `json:"shared_with" binding:"required"`
	CanEdit    bool       `json:"can_edit" binding:"omitempty"`
	CanDelete  bool       `json:"can_delete" binding:"omitempty"`
	ExpiresAt  *time.Time `json:"expires_at" binding:"omitempty"`
}

// WishlistShareResponse represents the response for a shared wishlist
type WishlistShareResponse struct {
	ID             uint             `json:"id"`
	WishlistID     uint             `json:"wishlist_id"`
	SharedBy       uint             `json:"shared_by"`
	SharedWith     uint             `json:"shared_with"`
	Token          string           `json:"token"`
	IsActive       bool             `json:"is_active"`
	ExpiresAt      time.Time        `json:"expires_at"`
	CanView        bool             `json:"can_view"`
	CanEdit        bool             `json:"can_edit"`
	CanDelete      bool             `json:"can_delete"`
	Wishlist       WishlistResponse `json:"wishlist"`
	SharedByUser   User             `json:"shared_by_user"`
	SharedWithUser User             `json:"shared_with_user"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	DeletedAt      *time.Time       `json:"deleted_at"`
}

// WishlistStatsResponse represents wishlist statistics
type WishlistStatsResponse struct {
	TotalWishlists     int64              `json:"total_wishlists"`
	TotalItems         int64              `json:"total_items"`
	TotalFavorites     int64              `json:"total_favorites"`
	PublicWishlists    int64              `json:"public_wishlists"`
	PrivateWishlists   int64              `json:"private_wishlists"`
	SharedWishlists    int64              `json:"shared_wishlists"`
	MostWishedProducts []ProductResponse  `json:"most_wished_products"`
	TopWishlists       []WishlistResponse `json:"top_wishlists"`
}

// WishlistFilterRequest represents wishlist filtering options
type WishlistFilterRequest struct {
	UserID    *uint           `json:"user_id" form:"user_id"`
	Status    *WishlistStatus `json:"status" form:"status"`
	IsPublic  *bool           `json:"is_public" form:"is_public"`
	IsDefault *bool           `json:"is_default" form:"is_default"`
	Search    *string         `json:"search" form:"search"`
	SortBy    *string         `json:"sort_by" form:"sort_by"`
	SortOrder *string         `json:"sort_order" form:"sort_order"`
}

// WishlistItemFilterRequest represents wishlist item filtering options
type WishlistItemFilterRequest struct {
	WishlistID *uint               `json:"wishlist_id" form:"wishlist_id"`
	ProductID  *uint               `json:"product_id" form:"product_id"`
	Status     *WishlistItemStatus `json:"status" form:"status"`
	Priority   *int                `json:"priority" form:"priority"`
	PriceMin   *float64            `json:"price_min" form:"price_min"`
	PriceMax   *float64            `json:"price_max" form:"price_max"`
	Search     *string             `json:"search" form:"search"`
	SortBy     *string             `json:"sort_by" form:"sort_by"`
	SortOrder  *string             `json:"sort_order" form:"sort_order"`
}

// FavoriteFilterRequest represents favorite filtering options
type FavoriteFilterRequest struct {
	UserID    *uint   `json:"user_id" form:"user_id"`
	ProductID *uint   `json:"product_id" form:"product_id"`
	Priority  *int    `json:"priority" form:"priority"`
	Search    *string `json:"search" form:"search"`
	SortBy    *string `json:"sort_by" form:"sort_by"`
	SortOrder *string `json:"sort_order" form:"sort_order"`
}

// Helper methods

// IsActive checks if a wishlist is active
func (w *Wishlist) IsActive() bool {
	return w.Status == WishlistStatusActive
}

// IsPublic checks if a wishlist is public
func (w *Wishlist) IsPublic() bool {
	return w.IsPublicFlag
}

// IsActive checks if a wishlist item is active
func (wi *WishlistItem) IsActive() bool {
	return wi.Status == WishlistItemStatusActive
}

// IsPurchased checks if a wishlist item is purchased
func (wi *WishlistItem) IsPurchased() bool {
	return wi.Status == WishlistItemStatusPurchased
}

// UpdatePrice updates the current price and calculates changes
func (wi *WishlistItem) UpdatePrice(currentPrice float64) {
	wi.CurrentPrice = currentPrice
	wi.PriceChange = currentPrice - wi.AddedPrice
	if wi.AddedPrice > 0 {
		wi.PriceChangePercent = (wi.PriceChange / wi.AddedPrice) * 100
	}
}

// ToResponse converts Wishlist to WishlistResponse
func (w *Wishlist) ToResponse() *WishlistResponse {
	items := make([]WishlistItemResponse, len(w.Items))
	for i, item := range w.Items {
		items[i] = *item.ToResponse()
	}

	return &WishlistResponse{
		ID:              w.ID,
		UserID:          w.UserID,
		Name:            w.Name,
		Description:     w.Description,
		Status:          w.Status,
		IsDefault:       w.IsDefault,
		IsPublic:        w.IsPublicFlag,
		SortOrder:       w.SortOrder,
		ViewCount:       w.ViewCount,
		ShareCount:      w.ShareCount,
		ItemCount:       w.ItemCount,
		Slug:            w.Slug,
		MetaTitle:       w.MetaTitle,
		MetaDescription: w.MetaDescription,
		User:            w.User,
		Items:           items,
		CreatedAt:       w.CreatedAt,
		UpdatedAt:       w.UpdatedAt,
		DeletedAt:       w.DeletedAt,
	}
}

// ToResponse converts WishlistItem to WishlistItemResponse
func (wi *WishlistItem) ToResponse() *WishlistItemResponse {
	return &WishlistItemResponse{
		ID:                 wi.ID,
		WishlistID:         wi.WishlistID,
		ProductID:          wi.ProductID,
		Status:             wi.Status,
		Quantity:           wi.Quantity,
		Notes:              wi.Notes,
		Priority:           wi.Priority,
		SortOrder:          wi.SortOrder,
		AddedPrice:         wi.AddedPrice,
		CurrentPrice:       wi.CurrentPrice,
		PriceChange:        wi.PriceChange,
		PriceChangePercent: wi.PriceChangePercent,
		NotifyOnPriceDrop:  wi.NotifyOnPriceDrop,
		NotifyOnStock:      wi.NotifyOnStock,
		NotifyOnSale:       wi.NotifyOnSale,
		Wishlist:           *wi.Wishlist.ToResponse(),
		Product:            wi.Product.ToResponse(),
		CreatedAt:          wi.CreatedAt,
		UpdatedAt:          wi.UpdatedAt,
		DeletedAt:          wi.DeletedAt,
	}
}

// ToResponse converts Favorite to FavoriteResponse
func (f *Favorite) ToResponse() *FavoriteResponse {
	return &FavoriteResponse{
		ID:                f.ID,
		UserID:            f.UserID,
		ProductID:         f.ProductID,
		Notes:             f.Notes,
		Priority:          f.Priority,
		NotifyOnPriceDrop: f.NotifyOnPriceDrop,
		NotifyOnStock:     f.NotifyOnStock,
		NotifyOnSale:      f.NotifyOnSale,
		User:              f.User,
		Product:           f.Product.ToResponse(),
		CreatedAt:         f.CreatedAt,
		UpdatedAt:         f.UpdatedAt,
		DeletedAt:         f.DeletedAt,
	}
}

// ToResponse converts WishlistShare to WishlistShareResponse
func (ws *WishlistShare) ToResponse() *WishlistShareResponse {
	return &WishlistShareResponse{
		ID:             ws.ID,
		WishlistID:     ws.WishlistID,
		SharedBy:       ws.SharedBy,
		SharedWith:     ws.SharedWith,
		Token:          ws.Token,
		IsActive:       ws.IsActive,
		ExpiresAt:      ws.ExpiresAt,
		CanView:        ws.CanView,
		CanEdit:        ws.CanEdit,
		CanDelete:      ws.CanDelete,
		Wishlist:       *ws.Wishlist.ToResponse(),
		SharedByUser:   ws.SharedByUser,
		SharedWithUser: ws.SharedWithUser,
		CreatedAt:      ws.CreatedAt,
		UpdatedAt:      ws.UpdatedAt,
		DeletedAt:      ws.DeletedAt,
	}
}
