package service

import (
	"fmt"
	"strings"

	"go_app/internal/model"
	"go_app/internal/repository"
	"go_app/pkg/utils"
)

type CategoryService struct {
	categoryRepo *repository.CategoryRepository
}

func NewCategoryService() *CategoryService {
	return &CategoryService{
		categoryRepo: repository.NewCategoryRepository(),
	}
}

// CreateCategory creates a new category
func (s *CategoryService) CreateCategory(req *model.CategoryCreateRequest) (*model.CategoryResponse, error) {
	// Generate slug from name
	slug := utils.GenerateSlug(req.Name)

	// Check if category with same slug exists
	exists, err := s.categoryRepo.ExistsBySlug(slug)
	if err != nil {
		return nil, fmt.Errorf("failed to check category slug: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("category with slug '%s' already exists", slug)
	}

	// Validate parent category if provided
	if req.ParentID != nil {
		parent, err := s.categoryRepo.GetByID(*req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent category not found: %w", err)
		}
		if !parent.IsActive {
			return nil, fmt.Errorf("parent category is not active")
		}
	}

	// Set default values
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Create category
	category := &model.Category{
		Name:        strings.TrimSpace(req.Name),
		Slug:        slug,
		Description: strings.TrimSpace(req.Description),
		Image:       strings.TrimSpace(req.Image),
		Icon:        strings.TrimSpace(req.Icon),
		ParentID:    req.ParentID,
		SortOrder:   req.SortOrder,
		IsActive:    isActive,
		IsLeaf:      true, // Will be updated by repository
	}

	if err := s.categoryRepo.Create(category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	response := category.ToResponse()
	return &response, nil
}

// GetCategoryByID gets a category by ID
func (s *CategoryService) GetCategoryByID(id uint) (*model.CategoryResponse, error) {
	category, err := s.categoryRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := category.ToResponse()
	return &response, nil
}

// GetCategoryBySlug gets a category by slug
func (s *CategoryService) GetCategoryBySlug(slug string) (*model.CategoryResponse, error) {
	category, err := s.categoryRepo.GetBySlug(slug)
	if err != nil {
		return nil, err
	}

	response := category.ToResponse()
	return &response, nil
}

// GetCategoryTree gets the complete category tree
func (s *CategoryService) GetCategoryTree() ([]model.CategoryTreeResponse, error) {
	categories, err := s.categoryRepo.GetTree()
	if err != nil {
		return nil, err
	}

	// Convert to tree response format
	responses := make([]model.CategoryTreeResponse, len(categories))
	for i, category := range categories {
		responses[i] = category.ToTreeResponse()
	}

	return responses, nil
}

// GetCategoryWithChildren gets a category with its children
func (s *CategoryService) GetCategoryWithChildren(id uint) (*model.CategoryResponse, error) {
	category, err := s.categoryRepo.GetWithChildren(id)
	if err != nil {
		return nil, err
	}

	response := category.ToResponse()
	return &response, nil
}

// GetCategoriesByLevel gets categories by level
func (s *CategoryService) GetCategoriesByLevel(level int) ([]model.CategoryResponse, error) {
	categories, err := s.categoryRepo.GetByLevel(level)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responses := make([]model.CategoryResponse, len(categories))
	for i, category := range categories {
		responses[i] = category.ToResponse()
	}

	return responses, nil
}

// GetCategoriesByParent gets categories by parent ID
func (s *CategoryService) GetCategoriesByParent(parentID *uint) ([]model.CategoryResponse, error) {
	categories, err := s.categoryRepo.GetByParentID(parentID)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responses := make([]model.CategoryResponse, len(categories))
	for i, category := range categories {
		responses[i] = category.ToResponse()
	}

	return responses, nil
}

// GetAllCategories gets all categories with pagination and filters
func (s *CategoryService) GetAllCategories(page, limit int, search, sortBy, sortOrder string, isActive *bool, level *int) ([]model.CategoryResponse, int64, error) {
	categories, total, err := s.categoryRepo.GetAll(page, limit, search, sortBy, sortOrder, isActive, level)
	if err != nil {
		return nil, 0, err
	}

	// Convert to response format
	responses := make([]model.CategoryResponse, len(categories))
	for i, category := range categories {
		responses[i] = category.ToResponse()
	}

	return responses, total, nil
}

// UpdateCategory updates a category
func (s *CategoryService) UpdateCategory(id uint, req *model.CategoryUpdateRequest) (*model.CategoryResponse, error) {
	// Get existing category
	category, err := s.categoryRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != "" {
		// Check if new name conflicts with existing categories
		exists, err := s.categoryRepo.ExistsBySlugExcludingID(utils.GenerateSlug(req.Name), id)
		if err != nil {
			return nil, fmt.Errorf("failed to check category slug: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("category with slug '%s' already exists", utils.GenerateSlug(req.Name))
		}

		category.Name = strings.TrimSpace(req.Name)
		category.Slug = utils.GenerateSlug(req.Name)
	}

	if req.Description != "" {
		category.Description = strings.TrimSpace(req.Description)
	}

	if req.Image != "" {
		category.Image = strings.TrimSpace(req.Image)
	}

	if req.Icon != "" {
		category.Icon = strings.TrimSpace(req.Icon)
	}

	if req.ParentID != nil {
		// Validate parent category if provided
		if *req.ParentID != 0 {
			parent, err := s.categoryRepo.GetByID(*req.ParentID)
			if err != nil {
				return nil, fmt.Errorf("parent category not found: %w", err)
			}
			if !parent.IsActive {
				return nil, fmt.Errorf("parent category is not active")
			}
			// Check for circular reference
			if *req.ParentID == id {
				return nil, fmt.Errorf("category cannot be its own parent")
			}
			// Check if parent is a descendant of this category
			if s.isDescendant(*req.ParentID, id) {
				return nil, fmt.Errorf("category cannot be parent of its descendant")
			}
		}
		category.ParentID = req.ParentID
	}

	if req.SortOrder != 0 {
		category.SortOrder = req.SortOrder
	}

	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	// Update category
	if err := s.categoryRepo.Update(category); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	response := category.ToResponse()
	return &response, nil
}

// DeleteCategory soft deletes a category
func (s *CategoryService) DeleteCategory(id uint) error {
	// Check if category exists
	_, err := s.categoryRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Delete category
	if err := s.categoryRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}

// GetCategoryBreadcrumbs gets breadcrumb navigation for a category
func (s *CategoryService) GetCategoryBreadcrumbs(id uint) ([]model.CategoryBreadcrumb, error) {
	breadcrumbs, err := s.categoryRepo.GetBreadcrumbs(id)
	if err != nil {
		return nil, err
	}

	return breadcrumbs, nil
}

// GetCategoryDescendants gets all descendants of a category
func (s *CategoryService) GetCategoryDescendants(id uint) ([]model.CategoryResponse, error) {
	descendants, err := s.categoryRepo.GetDescendants(id)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responses := make([]model.CategoryResponse, len(descendants))
	for i, category := range descendants {
		responses[i] = category.ToResponse()
	}

	return responses, nil
}

// GetCategoryAncestors gets all ancestors of a category
func (s *CategoryService) GetCategoryAncestors(id uint) ([]model.CategoryResponse, error) {
	ancestors, err := s.categoryRepo.GetAncestors(id)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responses := make([]model.CategoryResponse, len(ancestors))
	for i, category := range ancestors {
		responses[i] = category.ToResponse()
	}

	return responses, nil
}

// UpdateCategoryStatus updates the status of a category
func (s *CategoryService) UpdateCategoryStatus(id uint, isActive bool) (*model.CategoryResponse, error) {
	// Get existing category
	category, err := s.categoryRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update status
	category.IsActive = isActive

	// Update category
	if err := s.categoryRepo.Update(category); err != nil {
		return nil, fmt.Errorf("failed to update category status: %w", err)
	}

	response := category.ToResponse()
	return &response, nil
}

// UpdateCategorySortOrder updates the sort order of a category
func (s *CategoryService) UpdateCategorySortOrder(id uint, sortOrder int) (*model.CategoryResponse, error) {
	// Get existing category
	category, err := s.categoryRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update sort order
	category.SortOrder = sortOrder

	// Update category
	if err := s.categoryRepo.Update(category); err != nil {
		return nil, fmt.Errorf("failed to update category sort order: %w", err)
	}

	response := category.ToResponse()
	return &response, nil
}

// BulkUpdateCategoryStatus updates the status of multiple categories
func (s *CategoryService) BulkUpdateCategoryStatus(ids []uint, isActive bool) error {
	if len(ids) == 0 {
		return fmt.Errorf("no category IDs provided")
	}

	// Validate all categories exist
	for _, id := range ids {
		_, err := s.categoryRepo.GetByID(id)
		if err != nil {
			return fmt.Errorf("category with ID %d not found: %w", id, err)
		}
	}

	// Bulk update status
	if err := s.categoryRepo.BulkUpdateStatus(ids, isActive); err != nil {
		return fmt.Errorf("failed to bulk update category status: %w", err)
	}

	return nil
}

// SearchCategories searches categories by query
func (s *CategoryService) SearchCategories(query string, limit int) ([]model.CategoryResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	categories, _, err := s.categoryRepo.GetAll(1, limit, query, "name", "asc", nil, nil)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responses := make([]model.CategoryResponse, len(categories))
	for i, category := range categories {
		responses[i] = category.ToResponse()
	}

	return responses, nil
}

// GetRootCategories gets all root categories (level 0)
func (s *CategoryService) GetRootCategories() ([]model.CategoryResponse, error) {
	return s.GetCategoriesByParent(nil)
}

// GetLeafCategories gets all leaf categories (categories without children)
func (s *CategoryService) GetLeafCategories() ([]model.CategoryResponse, error) {
	categories, _, err := s.categoryRepo.GetAll(0, 0, "", "name", "asc", nil, nil)
	if err != nil {
		return nil, err
	}

	// Filter leaf categories
	var leafCategories []model.Category
	for _, category := range categories {
		if category.IsLeaf {
			leafCategories = append(leafCategories, category)
		}
	}

	// Convert to response format
	responses := make([]model.CategoryResponse, len(leafCategories))
	for i, category := range leafCategories {
		responses[i] = category.ToResponse()
	}

	return responses, nil
}

// isDescendant checks if categoryID is a descendant of ancestorID
func (s *CategoryService) isDescendant(categoryID, ancestorID uint) bool {
	// Get category
	category, err := s.categoryRepo.GetByID(categoryID)
	if err != nil {
		return false
	}

	// Check if ancestorID is in the category's path
	pathParts := strings.Split(category.Path, "/")
	for _, part := range pathParts {
		var id uint
		if _, err := fmt.Sscanf(part, "%d", &id); err == nil {
			if id == ancestorID {
				return true
			}
		}
	}

	return false
}
