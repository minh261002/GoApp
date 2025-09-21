package handler

import (
	"net/http"
	"strconv"

	"go_app/internal/model"
	"go_app/internal/service"
	"go_app/pkg/response"
	"go_app/pkg/validator"

	"github.com/gin-gonic/gin"
)

type CartAdvancedHandler struct {
	cartAdvancedService *service.CartAdvancedService
}

func NewCartAdvancedHandler() *CartAdvancedHandler {
	return &CartAdvancedHandler{
		cartAdvancedService: service.NewCartAdvancedService(),
	}
}

// ===== CART SHARE ENDPOINTS =====

// CreateCartShare creates a new cart share
// @Summary Create cart share
// @Description Create a shareable link for a cart
// @Tags cart
// @Accept json
// @Produce json
// @Param share body model.CartShareCreateRequest true "Cart share information"
// @Success 201 {object} response.SuccessResponse{data=model.CartShareResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/carts/shares [post]
func (h *CartAdvancedHandler) CreateCartShare(c *gin.Context) {
	var req model.CartShareCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if err := validator.ValidateStruct(req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// Get user ID from context (assuming it's set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", "User ID not found in context")
		return
	}

	share, err := h.cartAdvancedService.CreateCartShare(&req, userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create cart share", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Cart share created successfully", share)
}

// GetCartShareByToken gets a cart share by token
// @Summary Get cart share by token
// @Description Get cart share information by token
// @Tags cart
// @Accept json
// @Produce json
// @Param token path string true "Share token"
// @Success 200 {object} response.SuccessResponse{data=model.CartShareResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/carts/shares/{token} [get]
func (h *CartAdvancedHandler) GetCartShareByToken(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		response.Error(c, http.StatusBadRequest, "Invalid token", "Token is required")
		return
	}

	share, err := h.cartAdvancedService.GetCartShareByToken(token)
	if err != nil {
		if err.Error() == "cart share not found or expired" {
			response.Error(c, http.StatusNotFound, "Cart share not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get cart share", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Cart share retrieved successfully", share)
}

// GetCartSharesByCartID gets all shares for a cart
// @Summary Get cart shares
// @Description Get all shares for a specific cart
// @Tags cart
// @Accept json
// @Produce json
// @Param cart_id path int true "Cart ID"
// @Success 200 {object} response.SuccessResponse{data=[]model.CartShareResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/carts/{cart_id}/shares [get]
func (h *CartAdvancedHandler) GetCartSharesByCartID(c *gin.Context) {
	cartIDStr := c.Param("cart_id")
	cartID, err := strconv.ParseUint(cartIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid cart ID", "Cart ID must be a valid number")
		return
	}

	shares, err := h.cartAdvancedService.GetCartSharesByCartID(uint(cartID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get cart shares", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Cart shares retrieved successfully", shares)
}

// UpdateCartShare updates a cart share
// @Summary Update cart share
// @Description Update cart share settings
// @Tags cart
// @Accept json
// @Produce json
// @Param id path int true "Share ID"
// @Param share body model.CartShareUpdateRequest true "Cart share update information"
// @Success 200 {object} response.SuccessResponse{data=model.CartShareResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/carts/shares/{id} [put]
func (h *CartAdvancedHandler) UpdateCartShare(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid share ID", "Share ID must be a valid number")
		return
	}

	var req model.CartShareUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if err := validator.ValidateStruct(req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", "User ID not found in context")
		return
	}

	share, err := h.cartAdvancedService.UpdateCartShare(uint(id), &req, userID.(uint))
	if err != nil {
		if err.Error() == "cart share not found" {
			response.Error(c, http.StatusNotFound, "Cart share not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update cart share", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Cart share updated successfully", share)
}

// DeleteCartShare deletes a cart share
// @Summary Delete cart share
// @Description Delete a cart share
// @Tags cart
// @Accept json
// @Produce json
// @Param id path int true "Share ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/carts/shares/{id} [delete]
func (h *CartAdvancedHandler) DeleteCartShare(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid share ID", "Share ID must be a valid number")
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", "User ID not found in context")
		return
	}

	err = h.cartAdvancedService.DeleteCartShare(uint(id), userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete cart share", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Cart share deleted successfully", nil)
}

// ===== SAVED FOR LATER ENDPOINTS =====

// SaveItemForLater saves an item for later purchase
// @Summary Save item for later
// @Description Save a product for later purchase
// @Tags cart
// @Accept json
// @Produce json
// @Param item body model.SavedForLaterCreateRequest true "Item information"
// @Success 201 {object} response.SuccessResponse{data=model.SavedForLaterResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/carts/saved-for-later [post]
func (h *CartAdvancedHandler) SaveItemForLater(c *gin.Context) {
	var req model.SavedForLaterCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if err := validator.ValidateStruct(req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", "User ID not found in context")
		return
	}

	item, err := h.cartAdvancedService.SaveItemForLater(&req, userID.(uint))
	if err != nil {
		if err.Error() == "item already saved for later" {
			response.Error(c, http.StatusConflict, "Item already saved for later", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to save item for later", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Item saved for later successfully", item)
}

// GetSavedForLaterByUser gets all saved items for a user
// @Summary Get saved items
// @Description Get all items saved for later by user
// @Tags cart
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.SuccessResponse{data=[]model.SavedForLaterResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/carts/saved-for-later [get]
func (h *CartAdvancedHandler) GetSavedForLaterByUser(c *gin.Context) {
	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", "User ID not found in context")
		return
	}

	items, total, err := h.cartAdvancedService.GetSavedForLaterByUser(userID.(uint), page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get saved items", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Saved items retrieved successfully", gin.H{
		"items": items,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// UpdateSavedForLater updates a saved item
// @Summary Update saved item
// @Description Update a saved for later item
// @Tags cart
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Param item body model.SavedForLaterUpdateRequest true "Item update information"
// @Success 200 {object} response.SuccessResponse{data=model.SavedForLaterResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/carts/saved-for-later/{id} [put]
func (h *CartAdvancedHandler) UpdateSavedForLater(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid item ID", "Item ID must be a valid number")
		return
	}

	var req model.SavedForLaterUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if err := validator.ValidateStruct(req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", "User ID not found in context")
		return
	}

	item, err := h.cartAdvancedService.UpdateSavedForLater(uint(id), &req, userID.(uint))
	if err != nil {
		if err.Error() == "saved for later item not found" {
			response.Error(c, http.StatusNotFound, "Saved item not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update saved item", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Saved item updated successfully", item)
}

// DeleteSavedForLater deletes a saved item
// @Summary Delete saved item
// @Description Delete a saved for later item
// @Tags cart
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/carts/saved-for-later/{id} [delete]
func (h *CartAdvancedHandler) DeleteSavedForLater(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid item ID", "Item ID must be a valid number")
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", "User ID not found in context")
		return
	}

	err = h.cartAdvancedService.DeleteSavedForLater(uint(id), userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete saved item", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Saved item deleted successfully", nil)
}

// MoveToCart moves a saved item to cart
// @Summary Move to cart
// @Description Move a saved item back to cart
// @Tags cart
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Param cart_id query int true "Cart ID"
// @Success 200 {object} response.SuccessResponse{data=model.CartItemResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/carts/saved-for-later/{id}/move-to-cart [post]
func (h *CartAdvancedHandler) MoveToCart(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid item ID", "Item ID must be a valid number")
		return
	}

	cartIDStr := c.Query("cart_id")
	cartID, err := strconv.ParseUint(cartIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid cart ID", "Cart ID is required")
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", "User ID not found in context")
		return
	}

	item, err := h.cartAdvancedService.MoveToCart(uint(id), uint(cartID), userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to move item to cart", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Item moved to cart successfully", item)
}

// ===== BULK ACTIONS ENDPOINTS =====

// BulkCartAction performs bulk actions on cart items
// @Summary Bulk cart action
// @Description Perform bulk actions on cart items
// @Tags cart
// @Accept json
// @Produce json
// @Param action body model.CartBulkActionRequest true "Bulk action information"
// @Success 200 {object} response.SuccessResponse{data=model.CartBulkActionResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/carts/bulk-action [post]
func (h *CartAdvancedHandler) BulkCartAction(c *gin.Context) {
	var req model.CartBulkActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if err := validator.ValidateStruct(req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", "User ID not found in context")
		return
	}

	result, err := h.cartAdvancedService.BulkCartAction(&req, userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to perform bulk action", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Bulk action completed", result)
}
