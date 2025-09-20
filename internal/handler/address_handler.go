package handler

import (
	"net/http"
	"strconv"

	"go_app/internal/model"
	"go_app/internal/service"
	"go_app/pkg/response"

	"github.com/gin-gonic/gin"
)

// AddressHandler handles address-related HTTP requests
type AddressHandler struct {
	addressService service.AddressService
}

// NewAddressHandler creates a new AddressHandler
func NewAddressHandler() *AddressHandler {
	return &AddressHandler{
		addressService: service.NewAddressService(),
	}
}

// Basic CRUD

// CreateAddress creates a new address
func (h *AddressHandler) CreateAddress(c *gin.Context) {
	var req model.AddressCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	address, err := h.addressService.CreateAddress(&req, userID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create address", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Address created successfully", address)
}

// GetAddressByID retrieves an address by its ID
func (h *AddressHandler) GetAddressByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid address ID", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	address, err := h.addressService.GetAddressByID(uint(id), userID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve address", err.Error())
		return
	}

	if address == nil {
		response.ErrorResponse(c, http.StatusNotFound, "Address not found", "address not found")
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Address retrieved successfully", address)
}

// GetAllAddresses retrieves all addresses with pagination and filters
func (h *AddressHandler) GetAllAddresses(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Parse filters
	filters := make(map[string]interface{})
	if userID := c.Query("user_id"); userID != "" {
		if id, err := strconv.ParseUint(userID, 10, 32); err == nil {
			filters["user_id"] = uint(id)
		}
	}
	if addressType := c.Query("type"); addressType != "" {
		filters["type"] = addressType
	}
	if isDefault := c.Query("is_default"); isDefault != "" {
		if isDefaultBool, err := strconv.ParseBool(isDefault); err == nil {
			filters["is_default"] = isDefaultBool
		}
	}
	if isActive := c.Query("is_active"); isActive != "" {
		if isActiveBool, err := strconv.ParseBool(isActive); err == nil {
			filters["is_active"] = isActiveBool
		}
	}
	if city := c.Query("city"); city != "" {
		filters["city"] = city
	}
	if district := c.Query("district"); district != "" {
		filters["district"] = district
	}
	if ward := c.Query("ward"); ward != "" {
		filters["ward"] = ward
	}
	if country := c.Query("country"); country != "" {
		filters["country"] = country
	}
	if postalCode := c.Query("postal_code"); postalCode != "" {
		filters["postal_code"] = postalCode
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	addresses, total, err := h.addressService.GetAllAddresses(page, limit, filters)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve addresses", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Addresses retrieved successfully", addresses, page, limit, total)
}

// GetAddressesByUser retrieves addresses for a specific user
func (h *AddressHandler) GetAddressesByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Parse filters
	filters := make(map[string]interface{})
	if addressType := c.Query("type"); addressType != "" {
		filters["type"] = addressType
	}
	if isDefault := c.Query("is_default"); isDefault != "" {
		if isDefaultBool, err := strconv.ParseBool(isDefault); err == nil {
			filters["is_default"] = isDefaultBool
		}
	}
	if isActive := c.Query("is_active"); isActive != "" {
		if isActiveBool, err := strconv.ParseBool(isActive); err == nil {
			filters["is_active"] = isActiveBool
		}
	}

	addresses, total, err := h.addressService.GetAddressesByUser(uint(userID), page, limit, filters)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve user addresses", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "User addresses retrieved successfully", addresses, page, limit, total)
}

// UpdateAddress updates an existing address
func (h *AddressHandler) UpdateAddress(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid address ID", err.Error())
		return
	}

	var req model.AddressUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	address, err := h.addressService.UpdateAddress(uint(id), &req, userID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update address", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Address updated successfully", address)
}

// DeleteAddress deletes an address
func (h *AddressHandler) DeleteAddress(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid address ID", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	if err := h.addressService.DeleteAddress(uint(id), userID.(uint)); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete address", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Address deleted successfully", nil)
}

// User-specific operations

// GetDefaultAddressByUser retrieves the default address for a user
func (h *AddressHandler) GetDefaultAddressByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	address, err := h.addressService.GetDefaultAddressByUser(uint(userID))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve default address", err.Error())
		return
	}

	if address == nil {
		response.ErrorResponse(c, http.StatusNotFound, "No default address found", "no default address found")
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Default address retrieved successfully", address)
}

// GetAddressesByType retrieves addresses by type for a user
func (h *AddressHandler) GetAddressesByType(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	addressTypeStr := c.Param("type")
	addressType := model.AddressType(addressTypeStr)
	if addressType != model.AddressTypeHome && addressType != model.AddressTypeOffice &&
		addressType != model.AddressTypeBilling && addressType != model.AddressTypeShipping &&
		addressType != model.AddressTypeOther {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid address type", "invalid address type")
		return
	}

	addresses, err := h.addressService.GetAddressesByType(uint(userID), addressType)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve addresses by type", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Addresses by type retrieved successfully", addresses)
}

// GetActiveAddressesByUser retrieves all active addresses for a user
func (h *AddressHandler) GetActiveAddressesByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	addresses, err := h.addressService.GetActiveAddressesByUser(uint(userID))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve active addresses", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Active addresses retrieved successfully", addresses)
}

// SetDefaultAddress sets an address as default for a user
func (h *AddressHandler) SetDefaultAddress(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	addressIDStr := c.Param("address_id")
	addressID, err := strconv.ParseUint(addressIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid address ID", err.Error())
		return
	}

	address, err := h.addressService.SetDefaultAddress(uint(userID), uint(addressID))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to set default address", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Default address set successfully", address)
}

// Geographic operations

// GetAddressesByCity retrieves addresses by city
func (h *AddressHandler) GetAddressesByCity(c *gin.Context) {
	city := c.Param("city")
	if city == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "City parameter is required", "city parameter is required")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	addresses, total, err := h.addressService.GetAddressesByCity(city, page, limit)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve addresses by city", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Addresses by city retrieved successfully", addresses, page, limit, total)
}

// GetAddressesByDistrict retrieves addresses by district
func (h *AddressHandler) GetAddressesByDistrict(c *gin.Context) {
	district := c.Param("district")
	if district == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "District parameter is required", "district parameter is required")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	addresses, total, err := h.addressService.GetAddressesByDistrict(district, page, limit)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve addresses by district", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Addresses by district retrieved successfully", addresses, page, limit, total)
}

// GetAddressesNearby retrieves addresses within a radius
func (h *AddressHandler) GetAddressesNearby(c *gin.Context) {
	// Parse coordinates
	latStr := c.Query("latitude")
	lonStr := c.Query("longitude")
	radiusStr := c.DefaultQuery("radius", "10") // Default 10km radius

	latitude, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid latitude", err.Error())
		return
	}

	longitude, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid longitude", err.Error())
		return
	}

	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid radius", err.Error())
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	addresses, total, err := h.addressService.GetAddressesNearby(latitude, longitude, radius, page, limit)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve nearby addresses", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Nearby addresses retrieved successfully", addresses, page, limit, total)
}

// SearchAddresses performs full-text search on addresses
func (h *AddressHandler) SearchAddresses(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Search query is required", "q parameter is required")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	addresses, total, err := h.addressService.SearchAddresses(query, page, limit)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to search addresses", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Address search completed successfully", addresses, page, limit, total)
}

// Statistics

// GetAddressStats retrieves address statistics
func (h *AddressHandler) GetAddressStats(c *gin.Context) {
	stats, err := h.addressService.GetAddressStats()
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve address statistics", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Address statistics retrieved successfully", stats)
}

// GetAddressStatsByUser retrieves address statistics for a specific user
func (h *AddressHandler) GetAddressStatsByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	stats, err := h.addressService.GetAddressStatsByUser(uint(userID))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve user address statistics", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "User address statistics retrieved successfully", stats)
}

// GetAddressStatsByCity retrieves address statistics by city
func (h *AddressHandler) GetAddressStatsByCity(c *gin.Context) {
	stats, err := h.addressService.GetAddressStatsByCity()
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve city address statistics", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "City address statistics retrieved successfully", stats)
}
