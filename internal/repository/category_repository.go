package repository

import (
	"fmt"
	"strings"

	"go_app/internal/model"
	"go_app/pkg/database"

	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository() *CategoryRepository {
	return &CategoryRepository{
		db: database.GetDB(),
	}
}

// Create creates a new category
func (r *CategoryRepository) Create(category *model.Category) error {
	// Calculate level and path
	if err := r.calculateLevelAndPath(category); err != nil {
		return fmt.Errorf("failed to calculate level and path: %w", err)
	}

	// Check if slug exists
	exists, err := r.ExistsBySlug(category.Slug)
	if err != nil {
		return fmt.Errorf("failed to check slug existence: %w", err)
	}
	if exists {
		return fmt.Errorf("category with slug '%s' already exists", category.Slug)
	}

	if err := r.db.Create(category).Error; err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}

	// Update is_leaf for parent category
	if category.ParentID != nil {
		if err := r.updateParentLeafStatus(*category.ParentID); err != nil {
			return fmt.Errorf("failed to update parent leaf status: %w", err)
		}
	}

	return nil
}

// GetByID gets a category by ID
func (r *CategoryRepository) GetByID(id uint) (*model.Category, error) {
	var category model.Category
	if err := r.db.Where("id = ?", id).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	return &category, nil
}

// GetBySlug gets a category by slug
func (r *CategoryRepository) GetBySlug(slug string) (*model.Category, error) {
	var category model.Category
	if err := r.db.Where("slug = ?", slug).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	return &category, nil
}

// GetWithChildren gets a category with its children
func (r *CategoryRepository) GetWithChildren(id uint) (*model.Category, error) {
	var category model.Category
	if err := r.db.Preload("Children", func(db *gorm.DB) *gorm.DB {
		return db.Where("is_active = ?", true).Order("sort_order ASC, name ASC")
	}).Where("id = ?", id).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category with children: %w", err)
	}
	return &category, nil
}

// GetTree gets the complete category tree
func (r *CategoryRepository) GetTree() ([]model.Category, error) {
	var categories []model.Category
	if err := r.db.Where("parent_id IS NULL AND is_active = ?", true).
		Preload("Children", func(db *gorm.DB) *gorm.DB {
			return r.preloadChildrenRecursively(db)
		}).Order("sort_order ASC, name ASC").Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("failed to get category tree: %w", err)
	}
	return categories, nil
}

// preloadChildrenRecursively recursively preloads children
func (r *CategoryRepository) preloadChildrenRecursively(db *gorm.DB) *gorm.DB {
	return db.Where("is_active = ?", true).Order("sort_order ASC, name ASC").
		Preload("Children", func(db *gorm.DB) *gorm.DB {
			return r.preloadChildrenRecursively(db)
		})
}

// GetByLevel gets categories by level
func (r *CategoryRepository) GetByLevel(level int) ([]model.Category, error) {
	var categories []model.Category
	if err := r.db.Where("level = ? AND is_active = ?", level, true).
		Order("sort_order ASC, name ASC").Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("failed to get categories by level: %w", err)
	}
	return categories, nil
}

// GetByParentID gets categories by parent ID
func (r *CategoryRepository) GetByParentID(parentID *uint) ([]model.Category, error) {
	var categories []model.Category
	query := r.db.Where("is_active = ?", true).Order("sort_order ASC, name ASC")

	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}

	if err := query.Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("failed to get categories by parent ID: %w", err)
	}
	return categories, nil
}

// GetAll gets all categories with pagination and filters
func (r *CategoryRepository) GetAll(page, limit int, search, sortBy, sortOrder string, isActive *bool, level *int) ([]model.Category, int64, error) {
	var categories []model.Category
	var total int64

	query := r.db.Model(&model.Category{})

	// Apply search filter
	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
	}

	// Apply active filter
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	// Apply level filter
	if level != nil {
		query = query.Where("level = ?", *level)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count categories: %w", err)
	}

	// Apply sorting
	if sortBy != "" {
		order := sortBy
		if sortOrder == "desc" {
			order += " DESC"
		} else {
			order += " ASC"
		}
		query = query.Order(order)
	} else {
		query = query.Order("level ASC, sort_order ASC, name ASC")
	}

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		query = query.Offset(offset).Limit(limit)
	}

	// Execute query
	if err := query.Find(&categories).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get categories: %w", err)
	}

	return categories, total, nil
}

// Update updates a category
func (r *CategoryRepository) Update(category *model.Category) error {
	// Calculate level and path
	if err := r.calculateLevelAndPath(category); err != nil {
		return fmt.Errorf("failed to calculate level and path: %w", err)
	}

	// Check if slug exists (excluding current category)
	exists, err := r.ExistsBySlugExcludingID(category.Slug, category.ID)
	if err != nil {
		return fmt.Errorf("failed to check slug existence: %w", err)
	}
	if exists {
		return fmt.Errorf("category with slug '%s' already exists", category.Slug)
	}

	// Update category
	if err := r.db.Save(category).Error; err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	// Update is_leaf for old and new parent categories
	if err := r.updateParentLeafStatuses(category); err != nil {
		return fmt.Errorf("failed to update parent leaf statuses: %w", err)
	}

	return nil
}

// Delete soft deletes a category
func (r *CategoryRepository) Delete(id uint) error {
	// Check if category has children
	var count int64
	if err := r.db.Model(&model.Category{}).Where("parent_id = ? AND deleted_at IS NULL", id).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check children count: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("cannot delete category with children")
	}

	// Get category to update parent's is_leaf status
	category, err := r.GetByID(id)
	if err != nil {
		return err
	}

	// Delete category
	if err := r.db.Delete(&model.Category{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	// Update parent's is_leaf status
	if category.ParentID != nil {
		if err := r.updateParentLeafStatus(*category.ParentID); err != nil {
			return fmt.Errorf("failed to update parent leaf status: %w", err)
		}
	}

	return nil
}

// GetBreadcrumbs gets breadcrumb navigation for a category
func (r *CategoryRepository) GetBreadcrumbs(categoryID uint) ([]model.CategoryBreadcrumb, error) {
	var breadcrumbs []model.CategoryBreadcrumb

	// Get category
	category, err := r.GetByID(categoryID)
	if err != nil {
		return nil, err
	}

	// Parse path to get parent IDs
	if category.Path != "" {
		pathParts := strings.Split(category.Path, "/")
		if len(pathParts) > 1 {
			// Get parent categories
			var parentIDs []uint
			for i := 0; i < len(pathParts)-1; i++ {
				var id uint
				if _, err := fmt.Sscanf(pathParts[i], "%d", &id); err == nil {
					parentIDs = append(parentIDs, id)
				}
			}

			// Get parent categories
			if len(parentIDs) > 0 {
				var parents []model.Category
				if err := r.db.Where("id IN ?", parentIDs).Order("level ASC").Find(&parents).Error; err != nil {
					return nil, fmt.Errorf("failed to get parent categories: %w", err)
				}

				// Add parents to breadcrumbs
				for _, parent := range parents {
					breadcrumbs = append(breadcrumbs, model.CategoryBreadcrumb{
						ID:    parent.ID,
						Name:  parent.Name,
						Slug:  parent.Slug,
						Level: parent.Level,
					})
				}
			}
		}
	}

	// Add current category
	breadcrumbs = append(breadcrumbs, model.CategoryBreadcrumb{
		ID:    category.ID,
		Name:  category.Name,
		Slug:  category.Slug,
		Level: category.Level,
	})

	return breadcrumbs, nil
}

// GetDescendants gets all descendants of a category
func (r *CategoryRepository) GetDescendants(categoryID uint) ([]model.Category, error) {
	var descendants []model.Category

	// Get category to build path pattern
	category, err := r.GetByID(categoryID)
	if err != nil {
		return nil, err
	}

	// Find all categories that have this category in their path
	pathPattern := category.Path + "/%"
	if err := r.db.Where("path LIKE ? AND id != ? AND is_active = ?", pathPattern, categoryID, true).
		Order("level ASC, sort_order ASC, name ASC").Find(&descendants).Error; err != nil {
		return nil, fmt.Errorf("failed to get descendants: %w", err)
	}

	return descendants, nil
}

// GetAncestors gets all ancestors of a category
func (r *CategoryRepository) GetAncestors(categoryID uint) ([]model.Category, error) {
	var ancestors []model.Category

	// Get category to parse path
	category, err := r.GetByID(categoryID)
	if err != nil {
		return nil, err
	}

	// Parse path to get ancestor IDs
	if category.Path != "" {
		pathParts := strings.Split(category.Path, "/")
		if len(pathParts) > 1 {
			var ancestorIDs []uint
			for i := 0; i < len(pathParts)-1; i++ {
				var id uint
				if _, err := fmt.Sscanf(pathParts[i], "%d", &id); err == nil {
					ancestorIDs = append(ancestorIDs, id)
				}
			}

			// Get ancestor categories
			if len(ancestorIDs) > 0 {
				if err := r.db.Where("id IN ? AND is_active = ?", ancestorIDs, true).
					Order("level ASC").Find(&ancestors).Error; err != nil {
					return nil, fmt.Errorf("failed to get ancestors: %w", err)
				}
			}
		}
	}

	return ancestors, nil
}

// ExistsBySlug checks if a category with the given slug exists
func (r *CategoryRepository) ExistsBySlug(slug string) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Category{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check category slug existence: %w", err)
	}
	return count > 0, nil
}

// ExistsBySlugExcludingID checks if a category with the given slug exists excluding specific ID
func (r *CategoryRepository) ExistsBySlugExcludingID(slug string, excludeID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Category{}).Where("slug = ? AND id != ?", slug, excludeID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check category slug existence: %w", err)
	}
	return count > 0, nil
}

// calculateLevelAndPath calculates the level and path for a category
func (r *CategoryRepository) calculateLevelAndPath(category *model.Category) error {
	if category.ParentID == nil {
		// Root category
		category.Level = 0
		category.Path = fmt.Sprintf("%d", category.ID)
	} else {
		// Get parent category
		var parent model.Category
		if err := r.db.Where("id = ?", *category.ParentID).First(&parent).Error; err != nil {
			return fmt.Errorf("parent category not found")
		}

		// Calculate level and path
		category.Level = parent.Level + 1
		category.Path = parent.Path + "/" + fmt.Sprintf("%d", category.ID)
	}
	return nil
}

// updateParentLeafStatus updates the is_leaf status of a parent category
func (r *CategoryRepository) updateParentLeafStatus(parentID uint) error {
	var count int64
	if err := r.db.Model(&model.Category{}).Where("parent_id = ? AND deleted_at IS NULL", parentID).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count children: %w", err)
	}

	isLeaf := count == 0
	if err := r.db.Model(&model.Category{}).Where("id = ?", parentID).Update("is_leaf", isLeaf).Error; err != nil {
		return fmt.Errorf("failed to update parent leaf status: %w", err)
	}

	return nil
}

// updateParentLeafStatuses updates is_leaf status for both old and new parent categories
func (r *CategoryRepository) updateParentLeafStatuses(category *model.Category) error {
	// Get old category data
	var oldCategory model.Category
	if err := r.db.Where("id = ?", category.ID).First(&oldCategory).Error; err != nil {
		return fmt.Errorf("failed to get old category: %w", err)
	}

	// Update old parent's is_leaf status
	if oldCategory.ParentID != nil {
		if err := r.updateParentLeafStatus(*oldCategory.ParentID); err != nil {
			return fmt.Errorf("failed to update old parent leaf status: %w", err)
		}
	}

	// Update new parent's is_leaf status
	if category.ParentID != nil {
		if err := r.updateParentLeafStatus(*category.ParentID); err != nil {
			return fmt.Errorf("failed to update new parent leaf status: %w", err)
		}
	}

	return nil
}

// UpdateSortOrder updates the sort order of a category
func (r *CategoryRepository) UpdateSortOrder(id uint, sortOrder int) error {
	if err := r.db.Model(&model.Category{}).Where("id = ?", id).Update("sort_order", sortOrder).Error; err != nil {
		return fmt.Errorf("failed to update category sort order: %w", err)
	}
	return nil
}

// BulkUpdateStatus updates the status of multiple categories
func (r *CategoryRepository) BulkUpdateStatus(ids []uint, isActive bool) error {
	if err := r.db.Model(&model.Category{}).Where("id IN ?", ids).Update("is_active", isActive).Error; err != nil {
		return fmt.Errorf("failed to bulk update category status: %w", err)
	}
	return nil
}
