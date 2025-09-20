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

type BrandHandler struct {
	brandService *service.BrandService
}

func NewBrandHandler() *BrandHandler {
	return &BrandHandler{
		brandService: service.NewBrandService(),
	}
}

// CreateBrand creates a new brand
// @Summary Create a new brand
// @Description Create a new brand with the provided information
// @Tags brands
// @Accept json
// @Produce json
// @Param brand body model.BrandCreateRequest true "Brand information"
// @Success 201 {object} response.SuccessResponse{data=model.BrandResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/brands [post]
func (h *BrandHandler) CreateBrand(c *gin.Context) {
	var req model.BrandCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if err := validator.ValidateStruct(req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	brand, err := h.brandService.CreateBrand(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create brand", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Brand created successfully", brand)
}

// GetBrandByID gets a brand by ID
// @Summary Get brand by ID
// @Description Get a brand by its ID
// @Tags brands
// @Accept json
// @Produce json
// @Param id path int true "Brand ID"
// @Success 200 {object} response.SuccessResponse{data=model.BrandResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/brands/{id} [get]
func (h *BrandHandler) GetBrandByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid brand ID", err.Error())
		return
	}

	brand, err := h.brandService.GetBrandByID(uint(id))
	if err != nil {
		if err.Error() == "brand not found" {
			response.Error(c, http.StatusNotFound, "Brand not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get brand", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Brand retrieved successfully", brand)
}

// GetBrandBySlug gets a brand by slug
// @Summary Get brand by slug
// @Description Get a brand by its slug
// @Tags brands
// @Accept json
// @Produce json
// @Param slug path string true "Brand slug"
// @Success 200 {object} response.SuccessResponse{data=model.BrandResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/brands/slug/{slug} [get]
func (h *BrandHandler) GetBrandBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		response.Error(c, http.StatusBadRequest, "Invalid slug", "Slug is required")
		return
	}

	brand, err := h.brandService.GetBrandBySlug(slug)
	if err != nil {
		if err.Error() == "brand not found" {
			response.Error(c, http.StatusNotFound, "Brand not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get brand", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Brand retrieved successfully", brand)
}

// GetAllBrands gets all brands with pagination and filters
// @Summary Get all brands
// @Description Get all brands with pagination and filtering options
// @Tags brands
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search term"
// @Param sort_by query string false "Sort field" default(name)
// @Param sort_order query string false "Sort order" Enums(asc, desc) default(asc)
// @Param is_active query bool false "Filter by active status"
// @Success 200 {object} response.SuccessResponse{data=[]model.BrandResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/brands [get]
func (h *BrandHandler) GetAllBrands(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	sortBy := c.DefaultQuery("sort_by", "name")
	sortOrder := c.DefaultQuery("sort_order", "asc")
	isActiveStr := c.Query("is_active")

	var isActive *bool
	if isActiveStr != "" {
		if val, err := strconv.ParseBool(isActiveStr); err == nil {
			isActive = &val
		}
	}

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Validate sort order
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc"
	}

	brands, total, err := h.brandService.GetAllBrands(page, limit, search, sortBy, sortOrder, isActive)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get brands", err.Error())
		return
	}

	response.PaginationResponse(c, brands, page, limit, total)
}

// UpdateBrand updates a brand
// @Summary Update a brand
// @Description Update a brand with the provided information
// @Tags brands
// @Accept json
// @Produce json
// @Param id path int true "Brand ID"
// @Param brand body model.BrandUpdateRequest true "Brand information"
// @Success 200 {object} response.SuccessResponse{data=model.BrandResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/brands/{id} [put]
func (h *BrandHandler) UpdateBrand(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid brand ID", err.Error())
		return
	}

	var req model.BrandUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if err := validator.ValidateStruct(req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	brand, err := h.brandService.UpdateBrand(uint(id), &req)
	if err != nil {
		if err.Error() == "brand not found" {
			response.Error(c, http.StatusNotFound, "Brand not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update brand", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Brand updated successfully", brand)
}

// DeleteBrand deletes a brand
// @Summary Delete a brand
// @Description Soft delete a brand by its ID
// @Tags brands
// @Accept json
// @Produce json
// @Param id path int true "Brand ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/brands/{id} [delete]
func (h *BrandHandler) DeleteBrand(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid brand ID", err.Error())
		return
	}

	err = h.brandService.DeleteBrand(uint(id))
	if err != nil {
		if err.Error() == "brand not found" {
			response.Error(c, http.StatusNotFound, "Brand not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to delete brand", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Brand deleted successfully", nil)
}

// GetActiveBrands gets all active brands
// @Summary Get active brands
// @Description Get all active brands
// @Tags brands
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=[]model.BrandResponse}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/brands/active [get]
func (h *BrandHandler) GetActiveBrands(c *gin.Context) {
	brands, err := h.brandService.GetActiveBrands()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get active brands", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Active brands retrieved successfully", brands)
}

// UpdateBrandStatus updates the status of a brand
// @Summary Update brand status
// @Description Update the active status of a brand
// @Tags brands
// @Accept json
// @Produce json
// @Param id path int true "Brand ID"
// @Param status body map[string]bool true "Status object with is_active field"
// @Success 200 {object} response.SuccessResponse{data=model.BrandResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/brands/{id}/status [patch]
func (h *BrandHandler) UpdateBrandStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid brand ID", err.Error())
		return
	}

	var req struct {
		IsActive bool `json:"is_active" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	brand, err := h.brandService.UpdateBrandStatus(uint(id), req.IsActive)
	if err != nil {
		if err.Error() == "brand not found" {
			response.Error(c, http.StatusNotFound, "Brand not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update brand status", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Brand status updated successfully", brand)
}

// SearchBrands searches brands by query
// @Summary Search brands
// @Description Search brands by query string
// @Tags brands
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Limit results" default(10)
// @Success 200 {object} response.SuccessResponse{data=[]model.BrandResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/brands/search [get]
func (h *BrandHandler) SearchBrands(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.Error(c, http.StatusBadRequest, "Search query is required", "Query parameter 'q' is required")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	brands, err := h.brandService.SearchBrands(query, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to search brands", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Brands search completed", brands)
}

// BulkUpdateBrandStatus updates the status of multiple brands
// @Summary Bulk update brand status
// @Description Update the status of multiple brands at once
// @Tags brands
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Request with ids array and is_active boolean"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/brands/bulk-status [patch]
func (h *BrandHandler) BulkUpdateBrandStatus(c *gin.Context) {
	var req struct {
		IDs      []uint `json:"ids" binding:"required"`
		IsActive bool   `json:"is_active" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if len(req.IDs) == 0 {
		response.Error(c, http.StatusBadRequest, "No brand IDs provided", "At least one brand ID is required")
		return
	}

	err := h.brandService.BulkUpdateBrandStatus(req.IDs, req.IsActive)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to bulk update brand status", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Brand statuses updated successfully", nil)
}
