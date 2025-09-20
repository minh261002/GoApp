package service

import (
	"fmt"
	"strings"

	"go_app/internal/model"
	"go_app/internal/repository"
	"go_app/pkg/utils"
)

type BrandService struct {
	brandRepo *repository.BrandRepository
}

func NewBrandService() *BrandService {
	return &BrandService{
		brandRepo: repository.NewBrandRepository(),
	}
}

// CreateBrand creates a new brand
func (s *BrandService) CreateBrand(req *model.BrandCreateRequest) (*model.BrandResponse, error) {
	// Generate slug from name
	slug := utils.GenerateSlug(req.Name)

	// Check if brand with same name exists
	exists, err := s.brandRepo.ExistsByName(req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check brand name: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("brand with name '%s' already exists", req.Name)
	}

	// Check if brand with same slug exists
	exists, err = s.brandRepo.ExistsBySlug(slug)
	if err != nil {
		return nil, fmt.Errorf("failed to check brand slug: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("brand with slug '%s' already exists", slug)
	}

	// Set default values
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Create brand
	brand := &model.Brand{
		Name:        strings.TrimSpace(req.Name),
		Slug:        slug,
		Description: strings.TrimSpace(req.Description),
		Logo:        strings.TrimSpace(req.Logo),
		Website:     strings.TrimSpace(req.Website),
		IsActive:    isActive,
		SortOrder:   req.SortOrder,
	}

	if err := s.brandRepo.Create(brand); err != nil {
		return nil, fmt.Errorf("failed to create brand: %w", err)
	}

	response := brand.ToResponse()
	return &response, nil
}

// GetBrandByID gets a brand by ID
func (s *BrandService) GetBrandByID(id uint) (*model.BrandResponse, error) {
	brand, err := s.brandRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := brand.ToResponse()
	return &response, nil
}

// GetBrandBySlug gets a brand by slug
func (s *BrandService) GetBrandBySlug(slug string) (*model.BrandResponse, error) {
	brand, err := s.brandRepo.GetBySlug(slug)
	if err != nil {
		return nil, err
	}

	response := brand.ToResponse()
	return &response, nil
}

// GetAllBrands gets all brands with pagination and filters
func (s *BrandService) GetAllBrands(page, limit int, search, sortBy, sortOrder string, isActive *bool) ([]model.BrandResponse, int64, error) {
	brands, total, err := s.brandRepo.GetAll(page, limit, search, sortBy, sortOrder, isActive)
	if err != nil {
		return nil, 0, err
	}

	// Convert to response format
	responses := make([]model.BrandResponse, len(brands))
	for i, brand := range brands {
		responses[i] = brand.ToResponse()
	}

	return responses, total, nil
}

// UpdateBrand updates a brand
func (s *BrandService) UpdateBrand(id uint, req *model.BrandUpdateRequest) (*model.BrandResponse, error) {
	// Get existing brand
	brand, err := s.brandRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != "" {
		// Check if new name conflicts with existing brands
		exists, err := s.brandRepo.ExistsByName(req.Name, id)
		if err != nil {
			return nil, fmt.Errorf("failed to check brand name: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("brand with name '%s' already exists", req.Name)
		}

		brand.Name = strings.TrimSpace(req.Name)
		brand.Slug = utils.GenerateSlug(req.Name)

		// Check if new slug conflicts with existing brands
		exists, err = s.brandRepo.ExistsBySlug(brand.Slug, id)
		if err != nil {
			return nil, fmt.Errorf("failed to check brand slug: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("brand with slug '%s' already exists", brand.Slug)
		}
	}

	if req.Description != "" {
		brand.Description = strings.TrimSpace(req.Description)
	}

	if req.Logo != "" {
		brand.Logo = strings.TrimSpace(req.Logo)
	}

	if req.Website != "" {
		brand.Website = strings.TrimSpace(req.Website)
	}

	if req.IsActive != nil {
		brand.IsActive = *req.IsActive
	}

	if req.SortOrder != 0 {
		brand.SortOrder = req.SortOrder
	}

	// Update brand
	if err := s.brandRepo.Update(brand); err != nil {
		return nil, fmt.Errorf("failed to update brand: %w", err)
	}

	response := brand.ToResponse()
	return &response, nil
}

// DeleteBrand soft deletes a brand
func (s *BrandService) DeleteBrand(id uint) error {
	// Check if brand exists
	_, err := s.brandRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Delete brand
	if err := s.brandRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete brand: %w", err)
	}

	return nil
}

// GetActiveBrands gets all active brands
func (s *BrandService) GetActiveBrands() ([]model.BrandResponse, error) {
	brands, err := s.brandRepo.GetActiveBrands()
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responses := make([]model.BrandResponse, len(brands))
	for i, brand := range brands {
		responses[i] = brand.ToResponse()
	}

	return responses, nil
}

// UpdateBrandStatus updates the status of a brand
func (s *BrandService) UpdateBrandStatus(id uint, isActive bool) (*model.BrandResponse, error) {
	// Get existing brand
	brand, err := s.brandRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update status
	brand.IsActive = isActive

	// Update brand
	if err := s.brandRepo.Update(brand); err != nil {
		return nil, fmt.Errorf("failed to update brand status: %w", err)
	}

	response := brand.ToResponse()
	return &response, nil
}

// UpdateBrandSortOrder updates the sort order of a brand
func (s *BrandService) UpdateBrandSortOrder(id uint, sortOrder int) (*model.BrandResponse, error) {
	// Get existing brand
	brand, err := s.brandRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update sort order
	brand.SortOrder = sortOrder

	// Update brand
	if err := s.brandRepo.Update(brand); err != nil {
		return nil, fmt.Errorf("failed to update brand sort order: %w", err)
	}

	response := brand.ToResponse()
	return &response, nil
}

// BulkUpdateBrandStatus updates the status of multiple brands
func (s *BrandService) BulkUpdateBrandStatus(ids []uint, isActive bool) error {
	if len(ids) == 0 {
		return fmt.Errorf("no brand IDs provided")
	}

	// Validate all brands exist
	for _, id := range ids {
		_, err := s.brandRepo.GetByID(id)
		if err != nil {
			return fmt.Errorf("brand with ID %d not found: %w", id, err)
		}
	}

	// Bulk update status
	if err := s.brandRepo.BulkUpdateStatus(ids, isActive); err != nil {
		return fmt.Errorf("failed to bulk update brand status: %w", err)
	}

	return nil
}

// SearchBrands searches brands by query
func (s *BrandService) SearchBrands(query string, limit int) ([]model.BrandResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	brands, _, err := s.brandRepo.GetAll(1, limit, query, "name", "asc", nil)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responses := make([]model.BrandResponse, len(brands))
	for i, brand := range brands {
		responses[i] = brand.ToResponse()
	}

	return responses, nil
}
