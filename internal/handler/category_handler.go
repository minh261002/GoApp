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

type CategoryHandler struct {
	categoryService *service.CategoryService
}

func NewCategoryHandler() *CategoryHandler {
	return &CategoryHandler{
		categoryService: service.NewCategoryService(),
	}
}

// CreateCategory creates a new category
// @Summary Create a new category
// @Description Create a new category with the provided information
// @Tags categories
// @Accept json
// @Produce json
// @Param category body model.CategoryCreateRequest true "Category information"
// @Success 201 {object} response.SuccessResponse{data=model.CategoryResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/categories [post]
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req model.CategoryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if err := validator.ValidateStruct(req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	category, err := h.categoryService.CreateCategory(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create category", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Category created successfully", category)
}

// GetCategoryByID gets a category by ID
// @Summary Get category by ID
// @Description Get a category by its ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} response.SuccessResponse{data=model.CategoryResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/categories/{id} [get]
func (h *CategoryHandler) GetCategoryByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid category ID", err.Error())
		return
	}

	category, err := h.categoryService.GetCategoryByID(uint(id))
	if err != nil {
		if err.Error() == "category not found" {
			response.Error(c, http.StatusNotFound, "Category not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get category", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Category retrieved successfully", category)
}

// GetCategoryBySlug gets a category by slug
// @Summary Get category by slug
// @Description Get a category by its slug
// @Tags categories
// @Accept json
// @Produce json
// @Param slug path string true "Category slug"
// @Success 200 {object} response.SuccessResponse{data=model.CategoryResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/categories/slug/{slug} [get]
func (h *CategoryHandler) GetCategoryBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		response.Error(c, http.StatusBadRequest, "Invalid slug", "Slug is required")
		return
	}

	category, err := h.categoryService.GetCategoryBySlug(slug)
	if err != nil {
		if err.Error() == "category not found" {
			response.Error(c, http.StatusNotFound, "Category not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get category", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Category retrieved successfully", category)
}

// GetCategoryTree gets the complete category tree
// @Summary Get category tree
// @Description Get the complete hierarchical category tree
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=[]model.CategoryTreeResponse}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/categories/tree [get]
func (h *CategoryHandler) GetCategoryTree(c *gin.Context) {
	tree, err := h.categoryService.GetCategoryTree()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get category tree", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Category tree retrieved successfully", tree)
}

// GetCategoryWithChildren gets a category with its children
// @Summary Get category with children
// @Description Get a category with its direct children
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} response.SuccessResponse{data=model.CategoryResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/categories/{id}/children [get]
func (h *CategoryHandler) GetCategoryWithChildren(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid category ID", err.Error())
		return
	}

	category, err := h.categoryService.GetCategoryWithChildren(uint(id))
	if err != nil {
		if err.Error() == "category not found" {
			response.Error(c, http.StatusNotFound, "Category not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get category with children", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Category with children retrieved successfully", category)
}

// GetAllCategories gets all categories with pagination and filters
// @Summary Get all categories
// @Description Get all categories with pagination and filtering options
// @Tags categories
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search term"
// @Param sort_by query string false "Sort field" default(name)
// @Param sort_order query string false "Sort order" Enums(asc, desc) default(asc)
// @Param is_active query bool false "Filter by active status"
// @Param level query int false "Filter by level"
// @Success 200 {object} response.SuccessResponse{data=[]model.CategoryResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/categories [get]
func (h *CategoryHandler) GetAllCategories(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	sortBy := c.DefaultQuery("sort_by", "name")
	sortOrder := c.DefaultQuery("sort_order", "asc")
	isActiveStr := c.Query("is_active")
	levelStr := c.Query("level")

	var isActive *bool
	if isActiveStr != "" {
		if val, err := strconv.ParseBool(isActiveStr); err == nil {
			isActive = &val
		}
	}

	var level *int
	if levelStr != "" {
		if val, err := strconv.Atoi(levelStr); err == nil {
			level = &val
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

	categories, total, err := h.categoryService.GetAllCategories(page, limit, search, sortBy, sortOrder, isActive, level)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get categories", err.Error())
		return
	}

	response.PaginationResponse(c, categories, page, limit, total)
}

// GetCategoriesByLevel gets categories by level
// @Summary Get categories by level
// @Description Get all categories at a specific level
// @Tags categories
// @Accept json
// @Produce json
// @Param level path int true "Category level"
// @Success 200 {object} response.SuccessResponse{data=[]model.CategoryResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/categories/level/{level} [get]
func (h *CategoryHandler) GetCategoriesByLevel(c *gin.Context) {
	levelStr := c.Param("level")
	level, err := strconv.Atoi(levelStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid level", err.Error())
		return
	}

	categories, err := h.categoryService.GetCategoriesByLevel(level)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get categories by level", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Categories by level retrieved successfully", categories)
}

// GetCategoriesByParent gets categories by parent ID
// @Summary Get categories by parent
// @Description Get all direct children of a parent category
// @Tags categories
// @Accept json
// @Produce json
// @Param parent_id query int false "Parent category ID (null for root categories)"
// @Success 200 {object} response.SuccessResponse{data=[]model.CategoryResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/categories/parent [get]
func (h *CategoryHandler) GetCategoriesByParent(c *gin.Context) {
	parentIDStr := c.Query("parent_id")
	var parentID *uint

	if parentIDStr != "" {
		if id, err := strconv.ParseUint(parentIDStr, 10, 32); err == nil {
			parentIDUint := uint(id)
			parentID = &parentIDUint
		} else {
			response.Error(c, http.StatusBadRequest, "Invalid parent ID", err.Error())
			return
		}
	}

	categories, err := h.categoryService.GetCategoriesByParent(parentID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get categories by parent", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Categories by parent retrieved successfully", categories)
}

// GetRootCategories gets all root categories
// @Summary Get root categories
// @Description Get all root categories (level 0)
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=[]model.CategoryResponse}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/categories/root [get]
func (h *CategoryHandler) GetRootCategories(c *gin.Context) {
	categories, err := h.categoryService.GetRootCategories()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get root categories", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Root categories retrieved successfully", categories)
}

// GetLeafCategories gets all leaf categories
// @Summary Get leaf categories
// @Description Get all leaf categories (categories without children)
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=[]model.CategoryResponse}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/categories/leaf [get]
func (h *CategoryHandler) GetLeafCategories(c *gin.Context) {
	categories, err := h.categoryService.GetLeafCategories()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get leaf categories", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Leaf categories retrieved successfully", categories)
}

// UpdateCategory updates a category
// @Summary Update a category
// @Description Update a category with the provided information
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param category body model.CategoryUpdateRequest true "Category information"
// @Success 200 {object} response.SuccessResponse{data=model.CategoryResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid category ID", err.Error())
		return
	}

	var req model.CategoryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if err := validator.ValidateStruct(req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	category, err := h.categoryService.UpdateCategory(uint(id), &req)
	if err != nil {
		if err.Error() == "category not found" {
			response.Error(c, http.StatusNotFound, "Category not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update category", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Category updated successfully", category)
}

// DeleteCategory deletes a category
// @Summary Delete a category
// @Description Soft delete a category by its ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid category ID", err.Error())
		return
	}

	err = h.categoryService.DeleteCategory(uint(id))
	if err != nil {
		if err.Error() == "category not found" {
			response.Error(c, http.StatusNotFound, "Category not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to delete category", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Category deleted successfully", nil)
}

// GetCategoryBreadcrumbs gets breadcrumb navigation for a category
// @Summary Get category breadcrumbs
// @Description Get breadcrumb navigation for a category
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} response.SuccessResponse{data=[]model.CategoryBreadcrumb}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/categories/{id}/breadcrumbs [get]
func (h *CategoryHandler) GetCategoryBreadcrumbs(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid category ID", err.Error())
		return
	}

	breadcrumbs, err := h.categoryService.GetCategoryBreadcrumbs(uint(id))
	if err != nil {
		if err.Error() == "category not found" {
			response.Error(c, http.StatusNotFound, "Category not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get category breadcrumbs", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Category breadcrumbs retrieved successfully", breadcrumbs)
}

// GetCategoryDescendants gets all descendants of a category
// @Summary Get category descendants
// @Description Get all descendants of a category
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} response.SuccessResponse{data=[]model.CategoryResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/categories/{id}/descendants [get]
func (h *CategoryHandler) GetCategoryDescendants(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid category ID", err.Error())
		return
	}

	descendants, err := h.categoryService.GetCategoryDescendants(uint(id))
	if err != nil {
		if err.Error() == "category not found" {
			response.Error(c, http.StatusNotFound, "Category not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get category descendants", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Category descendants retrieved successfully", descendants)
}

// GetCategoryAncestors gets all ancestors of a category
// @Summary Get category ancestors
// @Description Get all ancestors of a category
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} response.SuccessResponse{data=[]model.CategoryResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/categories/{id}/ancestors [get]
func (h *CategoryHandler) GetCategoryAncestors(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid category ID", err.Error())
		return
	}

	ancestors, err := h.categoryService.GetCategoryAncestors(uint(id))
	if err != nil {
		if err.Error() == "category not found" {
			response.Error(c, http.StatusNotFound, "Category not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get category ancestors", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Category ancestors retrieved successfully", ancestors)
}

// UpdateCategoryStatus updates the status of a category
// @Summary Update category status
// @Description Update the active status of a category
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param status body map[string]bool true "Status object with is_active field"
// @Success 200 {object} response.SuccessResponse{data=model.CategoryResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/categories/{id}/status [patch]
func (h *CategoryHandler) UpdateCategoryStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid category ID", err.Error())
		return
	}

	var req struct {
		IsActive bool `json:"is_active" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	category, err := h.categoryService.UpdateCategoryStatus(uint(id), req.IsActive)
	if err != nil {
		if err.Error() == "category not found" {
			response.Error(c, http.StatusNotFound, "Category not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update category status", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Category status updated successfully", category)
}

// SearchCategories searches categories by query
// @Summary Search categories
// @Description Search categories by query string
// @Tags categories
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Limit results" default(10)
// @Success 200 {object} response.SuccessResponse{data=[]model.CategoryResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/categories/search [get]
func (h *CategoryHandler) SearchCategories(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.Error(c, http.StatusBadRequest, "Search query is required", "Query parameter 'q' is required")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	categories, err := h.categoryService.SearchCategories(query, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to search categories", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Categories search completed", categories)
}

// BulkUpdateCategoryStatus updates the status of multiple categories
// @Summary Bulk update category status
// @Description Update the status of multiple categories at once
// @Tags categories
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Request with ids array and is_active boolean"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/categories/bulk-status [patch]
func (h *CategoryHandler) BulkUpdateCategoryStatus(c *gin.Context) {
	var req struct {
		IDs      []uint `json:"ids" binding:"required"`
		IsActive bool   `json:"is_active" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if len(req.IDs) == 0 {
		response.Error(c, http.StatusBadRequest, "No category IDs provided", "At least one category ID is required")
		return
	}

	err := h.categoryService.BulkUpdateCategoryStatus(req.IDs, req.IsActive)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to bulk update category status", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Category statuses updated successfully", nil)
}
