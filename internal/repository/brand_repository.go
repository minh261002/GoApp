package repository

import (
	"fmt"
	"strings"

	"go_app/internal/model"
	"go_app/pkg/database"

	"gorm.io/gorm"
)

type BrandRepository struct {
	db *gorm.DB
}

func NewBrandRepository() *BrandRepository {
	return &BrandRepository{
		db: database.GetDB(),
	}
}

// Create creates a new brand
func (r *BrandRepository) Create(brand *model.Brand) error {
	if err := r.db.Create(brand).Error; err != nil {
		return fmt.Errorf("failed to create brand: %w", err)
	}
	return nil
}

// GetByID gets a brand by ID
func (r *BrandRepository) GetByID(id uint) (*model.Brand, error) {
	var brand model.Brand
	if err := r.db.Where("id = ?", id).First(&brand).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("brand not found")
		}
		return nil, fmt.Errorf("failed to get brand: %w", err)
	}
	return &brand, nil
}

// GetBySlug gets a brand by slug
func (r *BrandRepository) GetBySlug(slug string) (*model.Brand, error) {
	var brand model.Brand
	if err := r.db.Where("slug = ?", slug).First(&brand).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("brand not found")
		}
		return nil, fmt.Errorf("failed to get brand: %w", err)
	}
	return &brand, nil
}

// GetAll gets all brands with pagination and filters
func (r *BrandRepository) GetAll(page, limit int, search, sortBy, sortOrder string, isActive *bool) ([]model.Brand, int64, error) {
	var brands []model.Brand
	var total int64

	query := r.db.Model(&model.Brand{})

	// Apply search filter
	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
	}

	// Apply active filter
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count brands: %w", err)
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
		query = query.Order("sort_order ASC, name ASC")
	}

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		query = query.Offset(offset).Limit(limit)
	}

	// Execute query
	if err := query.Find(&brands).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get brands: %w", err)
	}

	return brands, total, nil
}

// Update updates a brand
func (r *BrandRepository) Update(brand *model.Brand) error {
	if err := r.db.Save(brand).Error; err != nil {
		return fmt.Errorf("failed to update brand: %w", err)
	}
	return nil
}

// Delete soft deletes a brand
func (r *BrandRepository) Delete(id uint) error {
	if err := r.db.Delete(&model.Brand{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete brand: %w", err)
	}
	return nil
}

// HardDelete permanently deletes a brand
func (r *BrandRepository) HardDelete(id uint) error {
	if err := r.db.Unscoped().Delete(&model.Brand{}, id).Error; err != nil {
		return fmt.Errorf("failed to hard delete brand: %w", err)
	}
	return nil
}

// ExistsByName checks if a brand with the given name exists
func (r *BrandRepository) ExistsByName(name string, excludeID ...uint) (bool, error) {
	query := r.db.Model(&model.Brand{}).Where("name = ?", name)

	if len(excludeID) > 0 {
		query = query.Where("id != ?", excludeID[0])
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check brand existence: %w", err)
	}

	return count > 0, nil
}

// ExistsBySlug checks if a brand with the given slug exists
func (r *BrandRepository) ExistsBySlug(slug string, excludeID ...uint) (bool, error) {
	query := r.db.Model(&model.Brand{}).Where("slug = ?", slug)

	if len(excludeID) > 0 {
		query = query.Where("id != ?", excludeID[0])
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check brand slug existence: %w", err)
	}

	return count > 0, nil
}

// GetActiveBrands gets all active brands
func (r *BrandRepository) GetActiveBrands() ([]model.Brand, error) {
	var brands []model.Brand
	if err := r.db.Where("is_active = ?", true).Order("sort_order ASC, name ASC").Find(&brands).Error; err != nil {
		return nil, fmt.Errorf("failed to get active brands: %w", err)
	}
	return brands, nil
}

// UpdateSortOrder updates the sort order of a brand
func (r *BrandRepository) UpdateSortOrder(id uint, sortOrder int) error {
	if err := r.db.Model(&model.Brand{}).Where("id = ?", id).Update("sort_order", sortOrder).Error; err != nil {
		return fmt.Errorf("failed to update brand sort order: %w", err)
	}
	return nil
}

// BulkUpdateStatus updates the status of multiple brands
func (r *BrandRepository) BulkUpdateStatus(ids []uint, isActive bool) error {
	if err := r.db.Model(&model.Brand{}).Where("id IN ?", ids).Update("is_active", isActive).Error; err != nil {
		return fmt.Errorf("failed to bulk update brand status: %w", err)
	}
	return nil
}
