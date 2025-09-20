package handler

import (
	"go_app/internal/model"
	"go_app/internal/service"
	"go_app/pkg/logger"
	"go_app/pkg/response"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	searchService *service.SearchService
}

func NewSearchHandler(searchService *service.SearchService) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
	}
}

// SearchProducts handles product search requests
// @Summary Search products
// @Description Search products with filters and pagination
// @Tags Search
// @Accept json
// @Produce json
// @Param query query string false "Search query"
// @Param category_id query int false "Category ID filter"
// @Param brand_id query int false "Brand ID filter"
// @Param min_price query number false "Minimum price filter"
// @Param max_price query number false "Maximum price filter"
// @Param in_stock query boolean false "In stock filter"
// @Param on_sale query boolean false "On sale filter"
// @Param rating query number false "Minimum rating filter"
// @Param sort_by query string false "Sort by field (name, price, rating, created_at)"
// @Param sort_order query string false "Sort order (asc, desc)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.Response{data=model.SearchResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/search/products [get]
func (h *SearchHandler) SearchProducts(c *gin.Context) {
	// Parse query parameters
	req := &model.SearchRequest{
		Query:     c.Query("query"),
		Page:      1,
		Limit:     20,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	// Parse optional parameters
	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		if categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32); err == nil {
			categoryIDUint := uint(categoryID)
			req.CategoryID = &categoryIDUint
		}
	}

	if brandIDStr := c.Query("brand_id"); brandIDStr != "" {
		if brandID, err := strconv.ParseUint(brandIDStr, 10, 32); err == nil {
			brandIDUint := uint(brandID)
			req.BrandID = &brandIDUint
		}
	}

	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			req.MinPrice = &minPrice
		}
	}

	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			req.MaxPrice = &maxPrice
		}
	}

	if inStockStr := c.Query("in_stock"); inStockStr != "" {
		if inStock, err := strconv.ParseBool(inStockStr); err == nil {
			req.InStock = &inStock
		}
	}

	if onSaleStr := c.Query("on_sale"); onSaleStr != "" {
		if onSale, err := strconv.ParseBool(onSaleStr); err == nil {
			req.OnSale = &onSale
		}
	}

	if ratingStr := c.Query("rating"); ratingStr != "" {
		if rating, err := strconv.ParseFloat(ratingStr, 64); err == nil {
			req.Rating = &rating
		}
	}

	if sortBy := c.Query("sort_by"); sortBy != "" {
		req.SortBy = sortBy
	}

	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		req.SortOrder = sortOrder
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			req.Limit = limit
		}
	}

	// Search products
	result, err := h.searchService.SearchProducts(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to search products", "error", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to search products", err)
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Products searched successfully", result)
}

// GetSearchSuggestions handles search suggestions requests
// @Summary Get search suggestions
// @Description Get search suggestions based on query
// @Tags Search
// @Accept json
// @Produce json
// @Param query query string true "Search query"
// @Param limit query int false "Number of suggestions" default(5)
// @Success 200 {object} response.Response{data=[]string}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/search/suggestions [get]
func (h *SearchHandler) GetSearchSuggestions(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Query parameter is required", nil)
		return
	}

	limit := 5
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 20 {
			limit = l
		}
	}

	suggestions, err := h.searchService.GetSearchSuggestions(c.Request.Context(), query, limit)
	if err != nil {
		logger.Error("Failed to get search suggestions", "error", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get search suggestions", err)
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Search suggestions retrieved successfully", suggestions)
}

// GetFilterOptions handles filter options requests
// @Summary Get filter options
// @Description Get available filter options for current search
// @Tags Search
// @Accept json
// @Produce json
// @Param query query string false "Search query"
// @Param category_id query int false "Category ID filter"
// @Param brand_id query int false "Brand ID filter"
// @Param min_price query number false "Minimum price filter"
// @Param max_price query number false "Maximum price filter"
// @Param in_stock query boolean false "In stock filter"
// @Param on_sale query boolean false "On sale filter"
// @Param rating query number false "Minimum rating filter"
// @Success 200 {object} response.Response{data=model.FilterOptions}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/search/filter-options [get]
func (h *SearchHandler) GetFilterOptions(c *gin.Context) {
	// Build search filters from query parameters
	filters := &model.SearchFilters{
		Query: c.Query("query"),
	}

	// Parse optional parameters
	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		if categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32); err == nil {
			categoryIDUint := uint(categoryID)
			filters.CategoryID = &categoryIDUint
		}
	}

	if brandIDStr := c.Query("brand_id"); brandIDStr != "" {
		if brandID, err := strconv.ParseUint(brandIDStr, 10, 32); err == nil {
			brandIDUint := uint(brandID)
			filters.BrandID = &brandIDUint
		}
	}

	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			filters.MinPrice = &minPrice
		}
	}

	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			filters.MaxPrice = &maxPrice
		}
	}

	if inStockStr := c.Query("in_stock"); inStockStr != "" {
		if inStock, err := strconv.ParseBool(inStockStr); err == nil {
			filters.InStock = &inStock
		}
	}

	if onSaleStr := c.Query("on_sale"); onSaleStr != "" {
		if onSale, err := strconv.ParseBool(onSaleStr); err == nil {
			filters.OnSale = &onSale
		}
	}

	if ratingStr := c.Query("rating"); ratingStr != "" {
		if rating, err := strconv.ParseFloat(ratingStr, 64); err == nil {
			filters.Rating = &rating
		}
	}

	// Get filter options
	options, err := h.searchService.GetFilterOptions(c.Request.Context(), filters)
	if err != nil {
		logger.Error("Failed to get filter options", "error", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get filter options", err)
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Filter options retrieved successfully", options)
}

// GetPopularSearches handles popular searches requests
// @Summary Get popular searches
// @Description Get popular search queries
// @Tags Search
// @Accept json
// @Produce json
// @Param limit query int false "Number of popular searches" default(10)
// @Success 200 {object} response.Response{data=[]model.PopularSearch}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/search/popular [get]
func (h *SearchHandler) GetPopularSearches(c *gin.Context) {
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	searches, err := h.searchService.GetPopularSearches(c.Request.Context(), limit)
	if err != nil {
		logger.Error("Failed to get popular searches", "error", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get popular searches", err)
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Popular searches retrieved successfully", searches)
}

// GetSearchTrends handles search trends requests
// @Summary Get search trends
// @Description Get search trends over time
// @Tags Search
// @Accept json
// @Produce json
// @Param days query int false "Number of days" default(7)
// @Success 200 {object} response.Response{data=[]model.SearchTrend}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/search/trends [get]
func (h *SearchHandler) GetSearchTrends(c *gin.Context) {
	days := 7
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	trends, err := h.searchService.GetSearchTrends(c.Request.Context(), days)
	if err != nil {
		logger.Error("Failed to get search trends", "error", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get search trends", err)
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Search trends retrieved successfully", trends)
}

// GetSearchAnalytics handles search analytics requests
// @Summary Get search analytics
// @Description Get search analytics and statistics
// @Tags Search
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param days query int false "Number of days for trends" default(7)
// @Param limit query int false "Number of top queries" default(10)
// @Success 200 {object} response.Response{data=model.SearchAnalyticsResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/search/analytics [get]
func (h *SearchHandler) GetSearchAnalytics(c *gin.Context) {
	// Parse date parameters
	startDate := time.Now().AddDate(0, 0, -30) // Default to last 30 days
	endDate := time.Now()

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed
		}
	}

	days := 7
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	req := &model.SearchAnalyticsRequest{
		StartDate: startDate,
		EndDate:   endDate,
		Days:      days,
		Limit:     limit,
	}

	analytics, err := h.searchService.GetSearchAnalytics(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to get search analytics", "error", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get search analytics", err)
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Search analytics retrieved successfully", analytics)
}

// GetSearchLogs handles search logs requests
// @Summary Get search logs
// @Description Get search logs with pagination
// @Tags Search
// @Accept json
// @Produce json
// @Param query query string false "Search query filter"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.Response{data=[]model.SearchLog}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/search/logs [get]
func (h *SearchHandler) GetSearchLogs(c *gin.Context) {
	// Parse query parameters
	req := &model.SearchLogRequest{
		Query: c.Query("query"),
		Page:  1,
		Limit: 20,
	}

	// Parse optional parameters
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			req.StartDate = &parsed
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			req.EndDate = &parsed
		}
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			req.Limit = limit
		}
	}

	// Get search logs
	logs, total, err := h.searchService.GetSearchLogs(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to get search logs", "error", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get search logs", err)
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Search logs retrieved successfully", logs, req.Page, req.Limit, total)
}

// GetSearchIndexStats handles search index stats requests
// @Summary Get search index stats
// @Description Get search index statistics
// @Tags Search
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=model.SearchIndexStats}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/search/index-stats [get]
func (h *SearchHandler) GetSearchIndexStats(c *gin.Context) {
	stats, err := h.searchService.GetSearchIndexStats(c.Request.Context())
	if err != nil {
		logger.Error("Failed to get search index stats", "error", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get search index stats", err)
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Search index stats retrieved successfully", stats)
}

// CreateSearchIndex handles search index creation requests
// @Summary Create search index
// @Description Create search index for products
// @Tags Search
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/search/index/create [post]
func (h *SearchHandler) CreateSearchIndex(c *gin.Context) {
	err := h.searchService.CreateSearchIndex(c.Request.Context())
	if err != nil {
		logger.Error("Failed to create search index", "error", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create search index", err)
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Search index created successfully", nil)
}

// UpdateSearchIndex handles search index update requests
// @Summary Update search index
// @Description Update search index for a specific product
// @Tags Search
// @Accept json
// @Produce json
// @Param product_id path int true "Product ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/search/index/update/{product_id} [put]
func (h *SearchHandler) UpdateSearchIndex(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", err)
		return
	}

	err = h.searchService.UpdateSearchIndex(c.Request.Context(), uint(productID))
	if err != nil {
		logger.Error("Failed to update search index", "error", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update search index", err)
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Search index updated successfully", nil)
}

// DeleteSearchIndex handles search index deletion requests
// @Summary Delete search index
// @Description Delete search index for a specific product
// @Tags Search
// @Accept json
// @Produce json
// @Param product_id path int true "Product ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/search/index/delete/{product_id} [delete]
func (h *SearchHandler) DeleteSearchIndex(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", err)
		return
	}

	err = h.searchService.DeleteSearchIndex(c.Request.Context(), uint(productID))
	if err != nil {
		logger.Error("Failed to delete search index", "error", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete search index", err)
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Search index deleted successfully", nil)
}

// RebuildSearchIndex handles search index rebuild requests
// @Summary Rebuild search index
// @Description Rebuild the entire search index
// @Tags Search
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/search/index/rebuild [post]
func (h *SearchHandler) RebuildSearchIndex(c *gin.Context) {
	err := h.searchService.RebuildSearchIndex(c.Request.Context())
	if err != nil {
		logger.Error("Failed to rebuild search index", "error", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to rebuild search index", err)
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Search index rebuilt successfully", nil)
}

// DeleteSearchLogs handles search logs deletion requests
// @Summary Delete search logs
// @Description Delete old search logs
// @Tags Search
// @Accept json
// @Produce json
// @Param older_than query string false "Delete logs older than this date (YYYY-MM-DD)"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/search/logs/delete [delete]
func (h *SearchHandler) DeleteSearchLogs(c *gin.Context) {
	olderThan := time.Now().AddDate(0, -6, 0) // Default to 6 months ago

	if olderThanStr := c.Query("older_than"); olderThanStr != "" {
		if parsed, err := time.Parse("2006-01-02", olderThanStr); err == nil {
			olderThan = parsed
		}
	}

	err := h.searchService.DeleteSearchLogs(c.Request.Context(), olderThan)
	if err != nil {
		logger.Error("Failed to delete search logs", "error", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete search logs", err)
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Search logs deleted successfully", nil)
}
