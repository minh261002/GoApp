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

type ProductHandler struct {
	productService *service.ProductService
}

func NewProductHandler() *ProductHandler {
	return &ProductHandler{
		productService: service.NewProductService(),
	}
}

// CreateProduct creates a new product
// @Summary Create a new product
// @Description Create a new product with the provided information
// @Tags products
// @Accept json
// @Produce json
// @Param product body model.ProductCreateRequest true "Product information"
// @Success 201 {object} response.SuccessResponse{data=model.ProductResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req model.ProductCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if err := validator.ValidateStruct(req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	product, err := h.productService.CreateProduct(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create product", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Product created successfully", product)
}

// GetProductByID gets a product by ID
// @Summary Get product by ID
// @Description Get a product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} response.SuccessResponse{data=model.ProductResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/products/{id} [get]
func (h *ProductHandler) GetProductByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid product ID", err.Error())
		return
	}

	product, err := h.productService.GetProductByID(uint(id))
	if err != nil {
		if err.Error() == "product not found" {
			response.Error(c, http.StatusNotFound, "Product not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get product", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Product retrieved successfully", product)
}

// GetProductBySlug gets a product by slug
// @Summary Get product by slug
// @Description Get a product by its slug
// @Tags products
// @Accept json
// @Produce json
// @Param slug path string true "Product slug"
// @Success 200 {object} response.SuccessResponse{data=model.ProductResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/products/slug/{slug} [get]
func (h *ProductHandler) GetProductBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		response.Error(c, http.StatusBadRequest, "Invalid slug", "Slug is required")
		return
	}

	product, err := h.productService.GetProductBySlug(slug)
	if err != nil {
		if err.Error() == "product not found" {
			response.Error(c, http.StatusNotFound, "Product not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get product", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Product retrieved successfully", product)
}

// GetProductBySKU gets a product by SKU
// @Summary Get product by SKU
// @Description Get a product by its SKU
// @Tags products
// @Accept json
// @Produce json
// @Param sku path string true "Product SKU"
// @Success 200 {object} response.SuccessResponse{data=model.ProductResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/products/sku/{sku} [get]
func (h *ProductHandler) GetProductBySKU(c *gin.Context) {
	sku := c.Param("sku")
	if sku == "" {
		response.Error(c, http.StatusBadRequest, "Invalid SKU", "SKU is required")
		return
	}

	product, err := h.productService.GetProductBySKU(sku)
	if err != nil {
		if err.Error() == "product not found" {
			response.Error(c, http.StatusNotFound, "Product not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get product", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Product retrieved successfully", product)
}

// GetAllProducts gets all products with pagination and filters
// @Summary Get all products
// @Description Get all products with pagination and filtering options
// @Tags products
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search term"
// @Param sort_by query string false "Sort field" default(created_at)
// @Param sort_order query string false "Sort order" Enums(asc, desc) default(desc)
// @Param status query string false "Filter by status" Enums(draft, active, inactive, archived)
// @Param type query string false "Filter by type" Enums(simple, variable)
// @Param brand_id query int false "Filter by brand ID"
// @Param category_id query int false "Filter by category ID"
// @Param is_featured query bool false "Filter by featured status"
// @Param price_min query number false "Minimum price"
// @Param price_max query number false "Maximum price"
// @Success 200 {object} response.SuccessResponse{data=[]model.ProductResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/products [get]
func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	statusStr := c.Query("status")
	typeStr := c.Query("type")
	brandIDStr := c.Query("brand_id")
	categoryIDStr := c.Query("category_id")
	isFeaturedStr := c.Query("is_featured")
	priceMinStr := c.Query("price_min")
	priceMaxStr := c.Query("price_max")

	// Parse filters
	var status *model.ProductStatus
	if statusStr != "" {
		statusVal := model.ProductStatus(statusStr)
		status = &statusVal
	}

	var productType *model.ProductType
	if typeStr != "" {
		typeVal := model.ProductType(typeStr)
		productType = &typeVal
	}

	var brandID *uint
	if brandIDStr != "" {
		if id, err := strconv.ParseUint(brandIDStr, 10, 32); err == nil {
			brandIDUint := uint(id)
			brandID = &brandIDUint
		}
	}

	var categoryID *uint
	if categoryIDStr != "" {
		if id, err := strconv.ParseUint(categoryIDStr, 10, 32); err == nil {
			categoryIDUint := uint(id)
			categoryID = &categoryIDUint
		}
	}

	var isFeatured *bool
	if isFeaturedStr != "" {
		if val, err := strconv.ParseBool(isFeaturedStr); err == nil {
			isFeatured = &val
		}
	}

	var priceMin *float64
	if priceMinStr != "" {
		if val, err := strconv.ParseFloat(priceMinStr, 64); err == nil {
			priceMin = &val
		}
	}

	var priceMax *float64
	if priceMaxStr != "" {
		if val, err := strconv.ParseFloat(priceMaxStr, 64); err == nil {
			priceMax = &val
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
		sortOrder = "desc"
	}

	products, total, err := h.productService.GetAllProducts(page, limit, search, sortBy, sortOrder,
		status, productType, brandID, categoryID, isFeatured, priceMin, priceMax)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get products", err.Error())
		return
	}

	response.PaginationResponse(c, products, page, limit, total)
}

// GetFeaturedProducts gets featured products
// @Summary Get featured products
// @Description Get all featured products
// @Tags products
// @Accept json
// @Produce json
// @Param limit query int false "Limit results" default(10)
// @Success 200 {object} response.SuccessResponse{data=[]model.ProductResponse}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/products/featured [get]
func (h *ProductHandler) GetFeaturedProducts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	products, err := h.productService.GetFeaturedProducts(limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get featured products", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Featured products retrieved successfully", products)
}

// GetProductsByBrand gets products by brand
// @Summary Get products by brand
// @Description Get all products for a specific brand
// @Tags products
// @Accept json
// @Produce json
// @Param brand_id path int true "Brand ID"
// @Param limit query int false "Limit results" default(10)
// @Success 200 {object} response.SuccessResponse{data=[]model.ProductResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/products/brand/{brand_id} [get]
func (h *ProductHandler) GetProductsByBrand(c *gin.Context) {
	brandIDStr := c.Param("brand_id")
	brandID, err := strconv.ParseUint(brandIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid brand ID", err.Error())
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	products, err := h.productService.GetProductsByBrand(uint(brandID), limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get products by brand", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Products by brand retrieved successfully", products)
}

// GetProductsByCategory gets products by category
// @Summary Get products by category
// @Description Get all products for a specific category
// @Tags products
// @Accept json
// @Produce json
// @Param category_id path int true "Category ID"
// @Param limit query int false "Limit results" default(10)
// @Success 200 {object} response.SuccessResponse{data=[]model.ProductResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/products/category/{category_id} [get]
func (h *ProductHandler) GetProductsByCategory(c *gin.Context) {
	categoryIDStr := c.Param("category_id")
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid category ID", err.Error())
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	products, err := h.productService.GetProductsByCategory(uint(categoryID), limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get products by category", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Products by category retrieved successfully", products)
}

// SearchProducts searches products by query
// @Summary Search products
// @Description Search products by query string
// @Tags products
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Limit results" default(10)
// @Success 200 {object} response.SuccessResponse{data=[]model.ProductResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/products/search [get]
func (h *ProductHandler) SearchProducts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.Error(c, http.StatusBadRequest, "Search query is required", "Query parameter 'q' is required")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	products, err := h.productService.SearchProducts(query, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to search products", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Products search completed", products)
}

// GetLowStockProducts gets products with low stock
// @Summary Get low stock products
// @Description Get all products with low stock quantity
// @Tags products
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=[]model.ProductResponse}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/products/low-stock [get]
func (h *ProductHandler) GetLowStockProducts(c *gin.Context) {
	products, err := h.productService.GetLowStockProducts()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get low stock products", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Low stock products retrieved successfully", products)
}

// UpdateProduct updates a product
// @Summary Update a product
// @Description Update a product with the provided information
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param product body model.ProductUpdateRequest true "Product information"
// @Success 200 {object} response.SuccessResponse{data=model.ProductResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid product ID", err.Error())
		return
	}

	var req model.ProductUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if err := validator.ValidateStruct(req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	product, err := h.productService.UpdateProduct(uint(id), &req)
	if err != nil {
		if err.Error() == "product not found" {
			response.Error(c, http.StatusNotFound, "Product not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update product", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Product updated successfully", product)
}

// DeleteProduct deletes a product
// @Summary Delete a product
// @Description Soft delete a product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid product ID", err.Error())
		return
	}

	err = h.productService.DeleteProduct(uint(id))
	if err != nil {
		if err.Error() == "product not found" {
			response.Error(c, http.StatusNotFound, "Product not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to delete product", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Product deleted successfully", nil)
}

// UpdateProductStock updates product stock quantity
// @Summary Update product stock
// @Description Update the stock quantity of a product
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param stock body map[string]int true "Stock object with quantity field"
// @Success 200 {object} response.SuccessResponse{data=model.ProductResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/products/{id}/stock [patch]
func (h *ProductHandler) UpdateProductStock(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid product ID", err.Error())
		return
	}

	var req struct {
		Quantity int `json:"quantity" binding:"required,min=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	product, err := h.productService.UpdateProductStock(uint(id), req.Quantity)
	if err != nil {
		if err.Error() == "product not found" {
			response.Error(c, http.StatusNotFound, "Product not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update product stock", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Product stock updated successfully", product)
}

// UpdateProductStatus updates the status of a product
// @Summary Update product status
// @Description Update the status of a product
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param status body map[string]string true "Status object with status field"
// @Success 200 {object} response.SuccessResponse{data=model.ProductResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/products/{id}/status [patch]
func (h *ProductHandler) UpdateProductStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid product ID", err.Error())
		return
	}

	var req struct {
		Status model.ProductStatus `json:"status" binding:"required,oneof=draft active inactive archived"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	product, err := h.productService.UpdateProductStatus(uint(id), req.Status)
	if err != nil {
		if err.Error() == "product not found" {
			response.Error(c, http.StatusNotFound, "Product not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update product status", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Product status updated successfully", product)
}

// BulkUpdateProductStatus updates the status of multiple products
// @Summary Bulk update product status
// @Description Update the status of multiple products at once
// @Tags products
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Request with ids array and status string"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/products/bulk-status [patch]
func (h *ProductHandler) BulkUpdateProductStatus(c *gin.Context) {
	var req struct {
		IDs    []uint              `json:"ids" binding:"required"`
		Status model.ProductStatus `json:"status" binding:"required,oneof=draft active inactive archived"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if len(req.IDs) == 0 {
		response.Error(c, http.StatusBadRequest, "No product IDs provided", "At least one product ID is required")
		return
	}

	err := h.productService.BulkUpdateProductStatus(req.IDs, req.Status)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to bulk update product status", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Product statuses updated successfully", nil)
}

// GetProductStats gets product statistics
// @Summary Get product statistics
// @Description Get statistics about products
// @Tags products
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=map[string]interface{}}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/products/stats [get]
func (h *ProductHandler) GetProductStats(c *gin.Context) {
	stats, err := h.productService.GetProductStats()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get product statistics", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Product statistics retrieved successfully", stats)
}

// ===== PRODUCT VARIANTS ENDPOINTS =====

// GetProductVariants gets all variants for a product
// @Summary Get product variants
// @Description Get all variants for a specific product
// @Tags products
// @Accept json
// @Produce json
// @Param product_id path int true "Product ID"
// @Success 200 {object} response.SuccessResponse{data=[]model.ProductVariantResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/products/{product_id}/variants [get]
func (h *ProductHandler) GetProductVariants(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid product ID", "Product ID must be a valid number")
		return
	}

	variants, err := h.productService.GetProductVariants(uint(productID))
	if err != nil {
		if err.Error() == "product not found" {
			response.Error(c, http.StatusNotFound, "Product not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get product variants", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Product variants retrieved successfully", variants)
}

// GetProductVariant gets a specific product variant
// @Summary Get product variant
// @Description Get a specific product variant by ID
// @Tags products
// @Accept json
// @Produce json
// @Param product_id path int true "Product ID"
// @Param variant_id path int true "Variant ID"
// @Success 200 {object} response.SuccessResponse{data=model.ProductVariantResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/products/{product_id}/variants/{variant_id} [get]
func (h *ProductHandler) GetProductVariant(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid product ID", "Product ID must be a valid number")
		return
	}

	variantIDStr := c.Param("variant_id")
	variantID, err := strconv.ParseUint(variantIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid variant ID", "Variant ID must be a valid number")
		return
	}

	variant, err := h.productService.GetProductVariant(uint(productID), uint(variantID))
	if err != nil {
		if err.Error() == "product variant not found" {
			response.Error(c, http.StatusNotFound, "Product variant not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get product variant", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Product variant retrieved successfully", variant)
}

// CreateProductVariant creates a new product variant
// @Summary Create product variant
// @Description Create a new variant for a product
// @Tags products
// @Accept json
// @Produce json
// @Param product_id path int true "Product ID"
// @Param variant body model.ProductVariantCreateRequest true "Variant information"
// @Success 201 {object} response.SuccessResponse{data=model.ProductVariantResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/products/{product_id}/variants [post]
func (h *ProductHandler) CreateProductVariant(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid product ID", "Product ID must be a valid number")
		return
	}

	var req model.ProductVariantCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if err := validator.ValidateStruct(req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	variant, err := h.productService.CreateProductVariant(uint(productID), &req)
	if err != nil {
		if err.Error() == "product not found" {
			response.Error(c, http.StatusNotFound, "Product not found", err.Error())
			return
		}
		if err.Error() == "variant with SKU already exists" {
			response.Error(c, http.StatusConflict, "Variant SKU already exists", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to create product variant", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Product variant created successfully", variant)
}

// UpdateProductVariant updates a product variant
// @Summary Update product variant
// @Description Update an existing product variant
// @Tags products
// @Accept json
// @Produce json
// @Param product_id path int true "Product ID"
// @Param variant_id path int true "Variant ID"
// @Param variant body model.ProductVariantUpdateRequest true "Variant information"
// @Success 200 {object} response.SuccessResponse{data=model.ProductVariantResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/products/{product_id}/variants/{variant_id} [put]
func (h *ProductHandler) UpdateProductVariant(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid product ID", "Product ID must be a valid number")
		return
	}

	variantIDStr := c.Param("variant_id")
	variantID, err := strconv.ParseUint(variantIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid variant ID", "Variant ID must be a valid number")
		return
	}

	var req model.ProductVariantUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if err := validator.ValidateStruct(req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	variant, err := h.productService.UpdateProductVariant(uint(productID), uint(variantID), &req)
	if err != nil {
		if err.Error() == "product variant not found" {
			response.Error(c, http.StatusNotFound, "Product variant not found", err.Error())
			return
		}
		if err.Error() == "variant with SKU already exists" {
			response.Error(c, http.StatusConflict, "Variant SKU already exists", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update product variant", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Product variant updated successfully", variant)
}

// DeleteProductVariant deletes a product variant
// @Summary Delete product variant
// @Description Delete a product variant
// @Tags products
// @Accept json
// @Produce json
// @Param product_id path int true "Product ID"
// @Param variant_id path int true "Variant ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/products/{product_id}/variants/{variant_id} [delete]
func (h *ProductHandler) DeleteProductVariant(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid product ID", "Product ID must be a valid number")
		return
	}

	variantIDStr := c.Param("variant_id")
	variantID, err := strconv.ParseUint(variantIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid variant ID", "Variant ID must be a valid number")
		return
	}

	err = h.productService.DeleteProductVariant(uint(productID), uint(variantID))
	if err != nil {
		if err.Error() == "product variant not found" {
			response.Error(c, http.StatusNotFound, "Product variant not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to delete product variant", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Product variant deleted successfully", nil)
}

// UpdateProductVariantStock updates stock for a product variant
// @Summary Update variant stock
// @Description Update stock quantity for a product variant
// @Tags products
// @Accept json
// @Produce json
// @Param product_id path int true "Product ID"
// @Param variant_id path int true "Variant ID"
// @Param stock body model.ProductVariantStockUpdateRequest true "Stock information"
// @Success 200 {object} response.SuccessResponse{data=model.ProductVariantResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/products/{product_id}/variants/{variant_id}/stock [patch]
func (h *ProductHandler) UpdateProductVariantStock(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid product ID", "Product ID must be a valid number")
		return
	}

	variantIDStr := c.Param("variant_id")
	variantID, err := strconv.ParseUint(variantIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid variant ID", "Variant ID must be a valid number")
		return
	}

	var req model.ProductVariantStockUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if err := validator.ValidateStruct(req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	variant, err := h.productService.UpdateProductVariantStock(uint(productID), uint(variantID), &req)
	if err != nil {
		if err.Error() == "product variant not found" {
			response.Error(c, http.StatusNotFound, "Product variant not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update variant stock", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Variant stock updated successfully", variant)
}

// UpdateProductVariantStatus updates status for a product variant
// @Summary Update variant status
// @Description Update active status for a product variant
// @Tags products
// @Accept json
// @Produce json
// @Param product_id path int true "Product ID"
// @Param variant_id path int true "Variant ID"
// @Param status body model.ProductVariantStatusUpdateRequest true "Status information"
// @Success 200 {object} response.SuccessResponse{data=model.ProductVariantResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/products/{product_id}/variants/{variant_id}/status [patch]
func (h *ProductHandler) UpdateProductVariantStatus(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid product ID", "Product ID must be a valid number")
		return
	}

	variantIDStr := c.Param("variant_id")
	variantID, err := strconv.ParseUint(variantIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid variant ID", "Variant ID must be a valid number")
		return
	}

	var req model.ProductVariantStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if err := validator.ValidateStruct(req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	variant, err := h.productService.UpdateProductVariantStatus(uint(productID), uint(variantID), &req)
	if err != nil {
		if err.Error() == "product variant not found" {
			response.Error(c, http.StatusNotFound, "Product variant not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update variant status", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Variant status updated successfully", variant)
}
