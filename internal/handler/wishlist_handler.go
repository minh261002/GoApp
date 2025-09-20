package handler

import (
	"net/http"
	"strconv"
	"time"

	"go_app/internal/model"
	"go_app/internal/service"
	"go_app/pkg/response"

	"github.com/gin-gonic/gin"
)

// WishlistHandler handles wishlist-related HTTP requests
type WishlistHandler struct {
	wishlistService service.WishlistService
}

// NewWishlistHandler creates a new WishlistHandler
func NewWishlistHandler() *WishlistHandler {
	return &WishlistHandler{
		wishlistService: service.NewWishlistService(),
	}
}

// Wishlist Management

// CreateWishlist creates a new wishlist
func (h *WishlistHandler) CreateWishlist(c *gin.Context) {
	var req model.WishlistCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	userID := c.GetUint("user_id")
	wishlist, err := h.wishlistService.CreateWishlist(&req, userID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create wishlist", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Wishlist created successfully", wishlist)
}

// GetWishlistByID retrieves a wishlist by ID
func (h *WishlistHandler) GetWishlistByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid wishlist ID", err.Error())
		return
	}

	var userID *uint
	if uid, exists := c.Get("user_id"); exists {
		userIDValue := uid.(uint)
		userID = &userIDValue
	}

	wishlist, err := h.wishlistService.GetWishlistByID(uint(id), userID)
	if err != nil {
		if err.Error() == "wishlist not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Wishlist not found", err.Error())
			return
		}
		if err.Error() == "access denied" {
			response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve wishlist", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Wishlist retrieved successfully", wishlist)
}

// GetWishlistBySlug retrieves a wishlist by slug
func (h *WishlistHandler) GetWishlistBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid slug", "Slug is required")
		return
	}

	var userID *uint
	if uid, exists := c.Get("user_id"); exists {
		userIDValue := uid.(uint)
		userID = &userIDValue
	}

	wishlist, err := h.wishlistService.GetWishlistBySlug(slug, userID)
	if err != nil {
		if err.Error() == "wishlist not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Wishlist not found", err.Error())
			return
		}
		if err.Error() == "access denied" {
			response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve wishlist", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Wishlist retrieved successfully", wishlist)
}

// GetWishlistsByUser retrieves wishlists for a user
func (h *WishlistHandler) GetWishlistsByUser(c *gin.Context) {
	userID := c.GetUint("user_id")

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Parse filters
	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if isPublic := c.Query("is_public"); isPublic != "" {
		if isPublic == "true" {
			filters["is_public"] = true
		} else if isPublic == "false" {
			filters["is_public"] = false
		}
	}
	if isDefault := c.Query("is_default"); isDefault != "" {
		if isDefault == "true" {
			filters["is_default"] = true
		} else if isDefault == "false" {
			filters["is_default"] = false
		}
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	wishlists, total, err := h.wishlistService.GetWishlistsByUser(userID, page, limit, filters)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve wishlists", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Wishlists retrieved successfully", wishlists, page, limit, total)
}

// UpdateWishlist updates a wishlist
func (h *WishlistHandler) UpdateWishlist(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid wishlist ID", err.Error())
		return
	}

	var req model.WishlistUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	userID := c.GetUint("user_id")
	wishlist, err := h.wishlistService.UpdateWishlist(uint(id), &req, userID)
	if err != nil {
		if err.Error() == "wishlist not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Wishlist not found", err.Error())
			return
		}
		if err.Error() == "access denied" {
			response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update wishlist", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Wishlist updated successfully", wishlist)
}

// DeleteWishlist deletes a wishlist
func (h *WishlistHandler) DeleteWishlist(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid wishlist ID", err.Error())
		return
	}

	userID := c.GetUint("user_id")
	err = h.wishlistService.DeleteWishlist(uint(id), userID)
	if err != nil {
		if err.Error() == "wishlist not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Wishlist not found", err.Error())
			return
		}
		if err.Error() == "access denied" {
			response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete wishlist", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Wishlist deleted successfully", nil)
}

// SetDefaultWishlist sets a wishlist as default
func (h *WishlistHandler) SetDefaultWishlist(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid wishlist ID", err.Error())
		return
	}

	userID := c.GetUint("user_id")
	err = h.wishlistService.SetDefaultWishlist(uint(id), userID)
	if err != nil {
		if err.Error() == "wishlist not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Wishlist not found", err.Error())
			return
		}
		if err.Error() == "access denied" {
			response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to set default wishlist", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Default wishlist set successfully", nil)
}

// Public Wishlists

// GetPublicWishlists retrieves public wishlists
func (h *WishlistHandler) GetPublicWishlists(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Parse filters
	filters := make(map[string]interface{})
	if userID := c.Query("user_id"); userID != "" {
		if uid, err := strconv.ParseUint(userID, 10, 32); err == nil {
			filters["user_id"] = uint(uid)
		}
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	wishlists, total, err := h.wishlistService.GetPublicWishlists(page, limit, filters)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve public wishlists", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Public wishlists retrieved successfully", wishlists, page, limit, total)
}

// SearchWishlists searches wishlists
func (h *WishlistHandler) SearchWishlists(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Search query is required", "Query parameter 'q' is required")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	wishlists, total, err := h.wishlistService.SearchWishlists(query, page, limit)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to search wishlists", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Wishlist search completed successfully", wishlists, page, limit, total)
}

// Wishlist Items

// AddItemToWishlist adds an item to a wishlist
func (h *WishlistHandler) AddItemToWishlist(c *gin.Context) {
	var req model.WishlistItemCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	userID := c.GetUint("user_id")
	item, err := h.wishlistService.AddItemToWishlist(&req, userID)
	if err != nil {
		if err.Error() == "wishlist not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Wishlist not found", err.Error())
			return
		}
		if err.Error() == "access denied" {
			response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
			return
		}
		if err.Error() == "product not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Product not found", err.Error())
			return
		}
		if err.Error() == "product already in wishlist" {
			response.ErrorResponse(c, http.StatusConflict, "Product already in wishlist", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to add item to wishlist", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Item added to wishlist successfully", item)
}

// GetWishlistItemByID retrieves a wishlist item by ID
func (h *WishlistHandler) GetWishlistItemByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid item ID", err.Error())
		return
	}

	var userID *uint
	if uid, exists := c.Get("user_id"); exists {
		userIDValue := uid.(uint)
		userID = &userIDValue
	}

	item, err := h.wishlistService.GetWishlistItemByID(uint(id), userID)
	if err != nil {
		if err.Error() == "wishlist item not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Wishlist item not found", err.Error())
			return
		}
		if err.Error() == "access denied" {
			response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve wishlist item", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Wishlist item retrieved successfully", item)
}

// GetWishlistItems retrieves wishlist items
func (h *WishlistHandler) GetWishlistItems(c *gin.Context) {
	wishlistID, err := strconv.ParseUint(c.Param("wishlist_id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid wishlist ID", err.Error())
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Parse filters
	filters := make(map[string]interface{})
	if productID := c.Query("product_id"); productID != "" {
		if pid, err := strconv.ParseUint(productID, 10, 32); err == nil {
			filters["product_id"] = uint(pid)
		}
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if priority := c.Query("priority"); priority != "" {
		if p, err := strconv.Atoi(priority); err == nil {
			filters["priority"] = p
		}
	}
	if priceMin := c.Query("price_min"); priceMin != "" {
		if p, err := strconv.ParseFloat(priceMin, 64); err == nil {
			filters["price_min"] = p
		}
	}
	if priceMax := c.Query("price_max"); priceMax != "" {
		if p, err := strconv.ParseFloat(priceMax, 64); err == nil {
			filters["price_max"] = p
		}
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	var userID *uint
	if uid, exists := c.Get("user_id"); exists {
		userIDValue := uid.(uint)
		userID = &userIDValue
	}

	items, total, err := h.wishlistService.GetWishlistItems(uint(wishlistID), page, limit, filters, userID)
	if err != nil {
		if err.Error() == "wishlist not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Wishlist not found", err.Error())
			return
		}
		if err.Error() == "access denied" {
			response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve wishlist items", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Wishlist items retrieved successfully", items, page, limit, total)
}

// UpdateWishlistItem updates a wishlist item
func (h *WishlistHandler) UpdateWishlistItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid item ID", err.Error())
		return
	}

	var req model.WishlistItemUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	userID := c.GetUint("user_id")
	item, err := h.wishlistService.UpdateWishlistItem(uint(id), &req, userID)
	if err != nil {
		if err.Error() == "wishlist item not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Wishlist item not found", err.Error())
			return
		}
		if err.Error() == "access denied" {
			response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update wishlist item", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Wishlist item updated successfully", item)
}

// DeleteWishlistItem deletes a wishlist item
func (h *WishlistHandler) DeleteWishlistItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid item ID", err.Error())
		return
	}

	userID := c.GetUint("user_id")
	err = h.wishlistService.DeleteWishlistItem(uint(id), userID)
	if err != nil {
		if err.Error() == "wishlist item not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Wishlist item not found", err.Error())
			return
		}
		if err.Error() == "access denied" {
			response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete wishlist item", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Wishlist item deleted successfully", nil)
}

// ReorderWishlistItems reorders wishlist items
func (h *WishlistHandler) ReorderWishlistItems(c *gin.Context) {
	wishlistID, err := strconv.ParseUint(c.Param("wishlist_id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid wishlist ID", err.Error())
		return
	}

	var req struct {
		ItemOrders map[uint]int `json:"item_orders" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	userID := c.GetUint("user_id")
	err = h.wishlistService.ReorderWishlistItems(uint(wishlistID), req.ItemOrders, userID)
	if err != nil {
		if err.Error() == "wishlist not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Wishlist not found", err.Error())
			return
		}
		if err.Error() == "access denied" {
			response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to reorder wishlist items", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Wishlist items reordered successfully", nil)
}

// MoveItemToWishlist moves an item to another wishlist
func (h *WishlistHandler) MoveItemToWishlist(c *gin.Context) {
	itemID, err := strconv.ParseUint(c.Param("item_id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid item ID", err.Error())
		return
	}

	var req struct {
		TargetWishlistID uint `json:"target_wishlist_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	userID := c.GetUint("user_id")
	err = h.wishlistService.MoveItemToWishlist(uint(itemID), req.TargetWishlistID, userID)
	if err != nil {
		if err.Error() == "wishlist item not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Wishlist item not found", err.Error())
			return
		}
		if err.Error() == "access denied" {
			response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
			return
		}
		if err.Error() == "target wishlist not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Target wishlist not found", err.Error())
			return
		}
		if err.Error() == "product already exists in target wishlist" {
			response.ErrorResponse(c, http.StatusConflict, "Product already exists in target wishlist", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to move item to wishlist", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Item moved to wishlist successfully", nil)
}

// Favorites

// AddToFavorites adds a product to favorites
func (h *WishlistHandler) AddToFavorites(c *gin.Context) {
	var req model.FavoriteCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	userID := c.GetUint("user_id")
	favorite, err := h.wishlistService.AddToFavorites(&req, userID)
	if err != nil {
		if err.Error() == "product not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Product not found", err.Error())
			return
		}
		if err.Error() == "product already in favorites" {
			response.ErrorResponse(c, http.StatusConflict, "Product already in favorites", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to add to favorites", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Product added to favorites successfully", favorite)
}

// GetFavoriteByID retrieves a favorite by ID
func (h *WishlistHandler) GetFavoriteByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid favorite ID", err.Error())
		return
	}

	var userID *uint
	if uid, exists := c.Get("user_id"); exists {
		userIDValue := uid.(uint)
		userID = &userIDValue
	}

	favorite, err := h.wishlistService.GetFavoriteByID(uint(id), userID)
	if err != nil {
		if err.Error() == "favorite not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Favorite not found", err.Error())
			return
		}
		if err.Error() == "access denied" {
			response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve favorite", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Favorite retrieved successfully", favorite)
}

// GetFavoritesByUser retrieves favorites for a user
func (h *WishlistHandler) GetFavoritesByUser(c *gin.Context) {
	userID := c.GetUint("user_id")

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Parse filters
	filters := make(map[string]interface{})
	if productID := c.Query("product_id"); productID != "" {
		if pid, err := strconv.ParseUint(productID, 10, 32); err == nil {
			filters["product_id"] = uint(pid)
		}
	}
	if priority := c.Query("priority"); priority != "" {
		if p, err := strconv.Atoi(priority); err == nil {
			filters["priority"] = p
		}
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	favorites, total, err := h.wishlistService.GetFavoritesByUser(userID, page, limit, filters)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve favorites", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Favorites retrieved successfully", favorites, page, limit, total)
}

// UpdateFavorite updates a favorite
func (h *WishlistHandler) UpdateFavorite(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid favorite ID", err.Error())
		return
	}

	var req model.FavoriteUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	userID := c.GetUint("user_id")
	favorite, err := h.wishlistService.UpdateFavorite(uint(id), &req, userID)
	if err != nil {
		if err.Error() == "favorite not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Favorite not found", err.Error())
			return
		}
		if err.Error() == "access denied" {
			response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update favorite", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Favorite updated successfully", favorite)
}

// RemoveFromFavorites removes a favorite
func (h *WishlistHandler) RemoveFromFavorites(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid favorite ID", err.Error())
		return
	}

	userID := c.GetUint("user_id")
	err = h.wishlistService.RemoveFromFavorites(uint(id), userID)
	if err != nil {
		if err.Error() == "favorite not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Favorite not found", err.Error())
			return
		}
		if err.Error() == "access denied" {
			response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove from favorites", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Product removed from favorites successfully", nil)
}

// RemoveFromFavoritesByProduct removes a favorite by product
func (h *WishlistHandler) RemoveFromFavoritesByProduct(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("product_id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", err.Error())
		return
	}

	userID := c.GetUint("user_id")
	err = h.wishlistService.RemoveFromFavoritesByProduct(uint(productID), userID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove from favorites", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Product removed from favorites successfully", nil)
}

// Wishlist Sharing

// ShareWishlist shares a wishlist with another user
func (h *WishlistHandler) ShareWishlist(c *gin.Context) {
	var req model.WishlistShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	userID := c.GetUint("user_id")
	share, err := h.wishlistService.ShareWishlist(&req, userID)
	if err != nil {
		if err.Error() == "wishlist not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Wishlist not found", err.Error())
			return
		}
		if err.Error() == "access denied" {
			response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
			return
		}
		if err.Error() == "shared with user not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Shared with user not found", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to share wishlist", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Wishlist shared successfully", share)
}

// GetWishlistShareByToken retrieves a wishlist share by token
func (h *WishlistHandler) GetWishlistShareByToken(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid token", "Token is required")
		return
	}

	share, err := h.wishlistService.GetWishlistShareByToken(token)
	if err != nil {
		if err.Error() == "wishlist share not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Wishlist share not found", err.Error())
			return
		}
		if err.Error() == "wishlist share has expired" {
			response.ErrorResponse(c, http.StatusGone, "Wishlist share has expired", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve wishlist share", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Wishlist share retrieved successfully", share)
}

// GetWishlistSharesByWishlist retrieves shares for a wishlist
func (h *WishlistHandler) GetWishlistSharesByWishlist(c *gin.Context) {
	wishlistID, err := strconv.ParseUint(c.Param("wishlist_id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid wishlist ID", err.Error())
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	userID := c.GetUint("user_id")
	shares, total, err := h.wishlistService.GetWishlistSharesByWishlist(uint(wishlistID), page, limit, userID)
	if err != nil {
		if err.Error() == "wishlist not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Wishlist not found", err.Error())
			return
		}
		if err.Error() == "access denied" {
			response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve wishlist shares", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Wishlist shares retrieved successfully", shares, page, limit, total)
}

// GetWishlistSharesByUser retrieves shares for a user
func (h *WishlistHandler) GetWishlistSharesByUser(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	userID := c.GetUint("user_id")
	shares, total, err := h.wishlistService.GetWishlistSharesByUser(userID, page, limit)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve wishlist shares", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Wishlist shares retrieved successfully", shares, page, limit, total)
}

// Analytics

// TrackWishlistView tracks a wishlist view
func (h *WishlistHandler) TrackWishlistView(c *gin.Context) {
	wishlistID, err := strconv.ParseUint(c.Param("wishlist_id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid wishlist ID", err.Error())
		return
	}

	var userID *uint
	if uid, exists := c.Get("user_id"); exists {
		userIDValue := uid.(uint)
		userID = &userIDValue
	}

	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	referrer := c.GetHeader("Referer")

	err = h.wishlistService.TrackWishlistView(uint(wishlistID), userID, ip, userAgent, referrer)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to track wishlist view", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Wishlist view tracked successfully", nil)
}

// TrackWishlistItemView tracks a wishlist item view
func (h *WishlistHandler) TrackWishlistItemView(c *gin.Context) {
	itemID, err := strconv.ParseUint(c.Param("item_id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid item ID", err.Error())
		return
	}

	var userID *uint
	if uid, exists := c.Get("user_id"); exists {
		userIDValue := uid.(uint)
		userID = &userIDValue
	}

	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	referrer := c.GetHeader("Referer")

	err = h.wishlistService.TrackWishlistItemView(uint(itemID), userID, ip, userAgent, referrer)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to track wishlist item view", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Wishlist item view tracked successfully", nil)
}

// TrackWishlistItemClick tracks a wishlist item click
func (h *WishlistHandler) TrackWishlistItemClick(c *gin.Context) {
	itemID, err := strconv.ParseUint(c.Param("item_id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid item ID", err.Error())
		return
	}

	var userID *uint
	if uid, exists := c.Get("user_id"); exists {
		userIDValue := uid.(uint)
		userID = &userIDValue
	}

	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	referrer := c.GetHeader("Referer")

	err = h.wishlistService.TrackWishlistItemClick(uint(itemID), userID, ip, userAgent, referrer)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to track wishlist item click", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Wishlist item click tracked successfully", nil)
}

// GetWishlistAnalytics retrieves analytics for a wishlist
func (h *WishlistHandler) GetWishlistAnalytics(c *gin.Context) {
	wishlistID, err := strconv.ParseUint(c.Param("wishlist_id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid wishlist ID", err.Error())
		return
	}

	// Parse date range
	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &t
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &t
		}
	}

	userID := c.GetUint("user_id")
	analytics, err := h.wishlistService.GetWishlistAnalytics(uint(wishlistID), startDate, endDate, userID)
	if err != nil {
		if err.Error() == "wishlist not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Wishlist not found", err.Error())
			return
		}
		if err.Error() == "access denied" {
			response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve wishlist analytics", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Wishlist analytics retrieved successfully", analytics)
}

// GetUserWishlistStats retrieves wishlist statistics for a user
func (h *WishlistHandler) GetUserWishlistStats(c *gin.Context) {
	userID := c.GetUint("user_id")
	stats, err := h.wishlistService.GetUserWishlistStats(userID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve user wishlist statistics", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "User wishlist statistics retrieved successfully", stats)
}

// GetWishlistStats retrieves overall wishlist statistics
func (h *WishlistHandler) GetWishlistStats(c *gin.Context) {
	stats, err := h.wishlistService.GetWishlistStats()
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve wishlist statistics", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Wishlist statistics retrieved successfully", stats)
}

// Price Tracking

// UpdateWishlistItemPrices updates prices for all wishlist items
func (h *WishlistHandler) UpdateWishlistItemPrices(c *gin.Context) {
	err := h.wishlistService.UpdateWishlistItemPrices()
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update wishlist item prices", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Wishlist item prices updated successfully", nil)
}

// GetItemsWithPriceChanges retrieves items with price changes
func (h *WishlistHandler) GetItemsWithPriceChanges(c *gin.Context) {
	items, err := h.wishlistService.GetItemsWithPriceChanges()
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve items with price changes", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Items with price changes retrieved successfully", items)
}

// GetItemsForPriceNotification retrieves items that need price notifications
func (h *WishlistHandler) GetItemsForPriceNotification(c *gin.Context) {
	items, err := h.wishlistService.GetItemsForPriceNotification()
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve items for price notification", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Items for price notification retrieved successfully", items)
}
