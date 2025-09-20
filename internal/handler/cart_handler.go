package handler

import (
	"net/http"
	"strconv"

	"go_app/internal/model"
	"go_app/internal/service"
	"go_app/pkg/logger"
	"go_app/pkg/response"

	"github.com/gin-gonic/gin"
)

// CartHandler handles cart-related HTTP requests
type CartHandler struct {
	orderService service.OrderService
}

// NewCartHandler creates a new CartHandler
func NewCartHandler(orderService service.OrderService) *CartHandler {
	return &CartHandler{
		orderService: orderService,
	}
}

// CreateCart creates a new cart
// @Summary Create a new cart
// @Description Create a new shopping cart for a user or guest
// @Tags carts
// @Accept json
// @Produce json
// @Param user_id query uint false "User ID (for logged-in users)"
// @Param session_id query string false "Session ID (for guest users)"
// @Success 201 {object} response.Response{data=model.CartResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/carts [post]
func (h *CartHandler) CreateCart(c *gin.Context) {
	var userID uint
	var sessionID string

	// Get user ID from query or context
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if id, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			userID = uint(id)
		}
	} else if userIDFromContext, exists := c.Get("user_id"); exists {
		if id, ok := userIDFromContext.(uint); ok {
			userID = id
		}
	}

	// Get session ID from query
	sessionID = c.Query("session_id")

	// At least one of userID or sessionID must be provided
	if userID == 0 && sessionID == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Either user_id or session_id must be provided", nil)
		return
	}

	cart, err := h.orderService.CreateCart(&model.CartCreateRequest{SessionID: sessionID}, userID)
	if err != nil {
		logger.Errorf("Failed to create cart: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create cart", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Cart created successfully", cart)
}

// GetCart gets a cart
// @Summary Get cart details
// @Description Get cart details by user ID or session ID
// @Tags carts
// @Accept json
// @Produce json
// @Param user_id query uint false "User ID (for logged-in users)"
// @Param session_id query string false "Session ID (for guest users)"
// @Success 200 {object} response.Response{data=model.CartResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/carts [get]
func (h *CartHandler) GetCart(c *gin.Context) {
	var userID uint
	var sessionID string

	// Get user ID from query or context
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if id, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			userID = uint(id)
		}
	} else if userIDFromContext, exists := c.Get("user_id"); exists {
		if id, ok := userIDFromContext.(uint); ok {
			userID = id
		}
	}

	// Get session ID from query
	sessionID = c.Query("session_id")

	// At least one of userID or sessionID must be provided
	if userID == 0 && sessionID == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Either user_id or session_id must be provided", nil)
		return
	}

	cart, err := h.orderService.GetCart(userID, sessionID)
	if err != nil {
		if err.Error() == "cart not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Cart not found", nil)
			return
		}
		logger.Errorf("Failed to get cart: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get cart", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Cart retrieved successfully", cart)
}

// UpdateCart updates a cart
// @Summary Update cart details
// @Description Update cart information like shipping address, billing address, and notes
// @Tags carts
// @Accept json
// @Produce json
// @Param id path int true "Cart ID"
// @Param cart body model.CartUpdateRequest true "Cart update data"
// @Success 200 {object} response.Response{data=model.CartResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/carts/{id} [put]
func (h *CartHandler) UpdateCart(c *gin.Context) {
	cartID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid cart ID", err.Error())
		return
	}

	var req model.CartUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	cart, err := h.orderService.UpdateCart(uint(cartID), &req, userID.(uint))
	if err != nil {
		if err.Error() == "cart not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Cart not found", nil)
			return
		}
		if err.Error() == "unauthorized: cart belongs to another user" {
			response.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
			return
		}
		logger.Errorf("Failed to update cart: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update cart", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Cart updated successfully", cart)
}

// DeleteCart deletes a cart
// @Summary Delete cart
// @Description Delete a cart and all its items
// @Tags carts
// @Accept json
// @Produce json
// @Param id path int true "Cart ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/carts/{id} [delete]
func (h *CartHandler) DeleteCart(c *gin.Context) {
	cartID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid cart ID", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	err = h.orderService.DeleteCart(uint(cartID), userID.(uint))
	if err != nil {
		if err.Error() == "cart not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Cart not found", nil)
			return
		}
		if err.Error() == "unauthorized: cart belongs to another user" {
			response.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
			return
		}
		logger.Errorf("Failed to delete cart: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete cart", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Cart deleted successfully", nil)
}

// ClearCart clears all items from a cart
// @Summary Clear cart items
// @Description Remove all items from a cart
// @Tags carts
// @Accept json
// @Produce json
// @Param id path int true "Cart ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/carts/{id}/clear [post]
func (h *CartHandler) ClearCart(c *gin.Context) {
	cartID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid cart ID", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	err = h.orderService.ClearCart(uint(cartID), userID.(uint))
	if err != nil {
		if err.Error() == "cart not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Cart not found", nil)
			return
		}
		if err.Error() == "unauthorized: cart belongs to another user" {
			response.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
			return
		}
		logger.Errorf("Failed to clear cart: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to clear cart", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Cart cleared successfully", nil)
}

// AddToCart adds an item to a cart
// @Summary Add item to cart
// @Description Add a product to the shopping cart
// @Tags carts
// @Accept json
// @Produce json
// @Param id path int true "Cart ID"
// @Param item body model.CartItemCreateRequest true "Cart item data"
// @Success 201 {object} response.Response{data=model.CartItemResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/carts/{id}/items [post]
func (h *CartHandler) AddToCart(c *gin.Context) {
	cartID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid cart ID", err.Error())
		return
	}

	var req model.CartItemCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	cartItem, err := h.orderService.AddToCart(uint(cartID), &req, userID.(uint))
	if err != nil {
		if err.Error() == "cart not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Cart not found", nil)
			return
		}
		if err.Error() == "unauthorized: cart belongs to another user" {
			response.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
			return
		}
		if err.Error() == "product not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Product not found", nil)
			return
		}
		if err.Error() == "product is not available" {
			response.ErrorResponse(c, http.StatusBadRequest, "Product is not available", nil)
			return
		}
		if err.Error() == "insufficient stock" {
			response.ErrorResponse(c, http.StatusBadRequest, "Insufficient stock", nil)
			return
		}
		logger.Errorf("Failed to add item to cart: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to add item to cart", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Item added to cart successfully", cartItem)
}

// UpdateCartItem updates a cart item
// @Summary Update cart item
// @Description Update quantity or other details of a cart item
// @Tags carts
// @Accept json
// @Produce json
// @Param id path int true "Cart ID"
// @Param item_id path int true "Cart Item ID"
// @Param item body model.CartItemCreateRequest true "Cart item update data"
// @Success 200 {object} response.Response{data=model.CartItemResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/carts/{id}/items/{item_id} [put]
func (h *CartHandler) UpdateCartItem(c *gin.Context) {
	cartID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid cart ID", err.Error())
		return
	}

	itemID, err := strconv.ParseUint(c.Param("item_id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid item ID", err.Error())
		return
	}

	var req model.CartItemCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	cartItem, err := h.orderService.UpdateCartItem(uint(cartID), uint(itemID), &req, userID.(uint))
	if err != nil {
		if err.Error() == "cart not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Cart not found", nil)
			return
		}
		if err.Error() == "unauthorized: cart belongs to another user" {
			response.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
			return
		}
		if err.Error() == "item does not belong to this cart" {
			response.ErrorResponse(c, http.StatusBadRequest, "Item does not belong to this cart", nil)
			return
		}
		if err.Error() == "product not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Product not found", nil)
			return
		}
		if err.Error() == "product is not available" {
			response.ErrorResponse(c, http.StatusBadRequest, "Product is not available", nil)
			return
		}
		if err.Error() == "insufficient stock" {
			response.ErrorResponse(c, http.StatusBadRequest, "Insufficient stock", nil)
			return
		}
		logger.Errorf("Failed to update cart item: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update cart item", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Cart item updated successfully", cartItem)
}

// RemoveFromCart removes an item from a cart
// @Summary Remove item from cart
// @Description Remove a product from the shopping cart
// @Tags carts
// @Accept json
// @Produce json
// @Param id path int true "Cart ID"
// @Param item_id path int true "Cart Item ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/carts/{id}/items/{item_id} [delete]
func (h *CartHandler) RemoveFromCart(c *gin.Context) {
	cartID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid cart ID", err.Error())
		return
	}

	itemID, err := strconv.ParseUint(c.Param("item_id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid item ID", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	err = h.orderService.RemoveFromCart(uint(cartID), uint(itemID), userID.(uint))
	if err != nil {
		if err.Error() == "cart not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Cart not found", nil)
			return
		}
		if err.Error() == "unauthorized: cart belongs to another user" {
			response.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
			return
		}
		if err.Error() == "item does not belong to this cart" {
			response.ErrorResponse(c, http.StatusBadRequest, "Item does not belong to this cart", nil)
			return
		}
		logger.Errorf("Failed to remove item from cart: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove item from cart", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Item removed from cart successfully", nil)
}

// GetCartItems gets all items in a cart
// @Summary Get cart items
// @Description Get all items in a cart
// @Tags carts
// @Accept json
// @Produce json
// @Param id path int true "Cart ID"
// @Success 200 {object} response.Response{data=[]model.CartItemResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/carts/{id}/items [get]
func (h *CartHandler) GetCartItems(c *gin.Context) {
	cartID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid cart ID", err.Error())
		return
	}

	cartItems, err := h.orderService.GetCartItems(uint(cartID))
	if err != nil {
		logger.Errorf("Failed to get cart items: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get cart items", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Cart items retrieved successfully", cartItems)
}

// SyncCartWithUser syncs a guest cart with a user account
// @Summary Sync guest cart with user
// @Description Sync a guest cart with a user account after login
// @Tags carts
// @Accept json
// @Produce json
// @Param id path int true "Cart ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/carts/{id}/sync [post]
func (h *CartHandler) SyncCartWithUser(c *gin.Context) {
	cartID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid cart ID", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	err = h.orderService.SyncCartWithUser(uint(cartID), userID.(uint))
	if err != nil {
		if err.Error() == "cart not found" {
			response.ErrorResponse(c, http.StatusNotFound, "Cart not found", nil)
			return
		}
		if err.Error() == "cart is not a guest cart" {
			response.ErrorResponse(c, http.StatusBadRequest, "Cart is not a guest cart", nil)
			return
		}
		logger.Errorf("Failed to sync cart with user: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to sync cart with user", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Cart synced with user successfully", nil)
}

// GetCartStats gets cart statistics
// @Summary Get cart statistics
// @Description Get cart statistics for admin dashboard
// @Tags carts
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Failure 500 {object} response.Response
// @Router /api/v1/carts/stats [get]
func (h *CartHandler) GetCartStats(c *gin.Context) {
	stats, err := h.orderService.GetCartStats()
	if err != nil {
		logger.Errorf("Failed to get cart stats: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get cart stats", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Cart statistics retrieved successfully", stats)
}
