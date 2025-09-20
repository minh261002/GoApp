package repository

import (
	"encoding/json"
	"fmt"

	"go_app/internal/model"
	"go_app/pkg/database"

	"gorm.io/gorm"
)

type ProductVariantRepository struct {
	db *gorm.DB
}

func NewProductVariantRepository() *ProductVariantRepository {
	return &ProductVariantRepository{
		db: database.GetDB(),
	}
}

// Create creates a new product variant
func (r *ProductVariantRepository) Create(variant *model.ProductVariant) error {
	// Check if SKU exists
	if variant.SKU != "" {
		exists, err := r.ExistsBySKU(variant.SKU)
		if err != nil {
			return fmt.Errorf("failed to check SKU existence: %w", err)
		}
		if exists {
			return fmt.Errorf("variant with SKU '%s' already exists", variant.SKU)
		}
	}

	// Convert attributes to JSON
	if len(variant.Attributes) > 0 {
		attrsJSON, err := json.Marshal(variant.Attributes)
		if err != nil {
			return fmt.Errorf("failed to marshal attributes: %w", err)
		}
		variant.Attributes = string(attrsJSON)
	}

	if err := r.db.Create(variant).Error; err != nil {
		return fmt.Errorf("failed to create product variant: %w", err)
	}

	return nil
}

// GetByID gets a product variant by ID
func (r *ProductVariantRepository) GetByID(id uint) (*model.ProductVariant, error) {
	var variant model.ProductVariant
	if err := r.db.Preload("Product").Where("id = ?", id).First(&variant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("product variant not found")
		}
		return nil, fmt.Errorf("failed to get product variant: %w", err)
	}

	// Parse attributes JSON
	if variant.Attributes != "" {
		var attrs map[string]string
		if err := json.Unmarshal([]byte(variant.Attributes), &attrs); err == nil {
			// Store as JSON string for now, will be handled in service
		}
	}

	return &variant, nil
}

// GetByProductID gets all variants for a product
func (r *ProductVariantRepository) GetByProductID(productID uint) ([]model.ProductVariant, error) {
	var variants []model.ProductVariant
	if err := r.db.Where("product_id = ? AND is_active = ?", productID, true).
		Order("sort_order ASC, name ASC").Find(&variants).Error; err != nil {
		return nil, fmt.Errorf("failed to get product variants: %w", err)
	}

	// Parse attributes JSON for all variants
	for i := range variants {
		if variants[i].Attributes != "" {
			var attrs map[string]string
			if err := json.Unmarshal([]byte(variants[i].Attributes), &attrs); err == nil {
				// Store as JSON string for now, will be handled in service
			}
		}
	}

	return variants, nil
}

// GetBySKU gets a product variant by SKU
func (r *ProductVariantRepository) GetBySKU(sku string) (*model.ProductVariant, error) {
	var variant model.ProductVariant
	if err := r.db.Preload("Product").Where("sku = ?", sku).First(&variant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("product variant not found")
		}
		return nil, fmt.Errorf("failed to get product variant: %w", err)
	}

	return &variant, nil
}

// Update updates a product variant
func (r *ProductVariantRepository) Update(variant *model.ProductVariant) error {
	// Check if SKU exists (excluding current variant)
	if variant.SKU != "" {
		exists, err := r.ExistsBySKUExcludingID(variant.SKU, variant.ID)
		if err != nil {
			return fmt.Errorf("failed to check SKU existence: %w", err)
		}
		if exists {
			return fmt.Errorf("variant with SKU '%s' already exists", variant.SKU)
		}
	}

	// Convert attributes to JSON
	if len(variant.Attributes) > 0 {
		attrsJSON, err := json.Marshal(variant.Attributes)
		if err != nil {
			return fmt.Errorf("failed to marshal attributes: %w", err)
		}
		variant.Attributes = string(attrsJSON)
	}

	if err := r.db.Save(variant).Error; err != nil {
		return fmt.Errorf("failed to update product variant: %w", err)
	}

	return nil
}

// Delete soft deletes a product variant
func (r *ProductVariantRepository) Delete(id uint) error {
	if err := r.db.Delete(&model.ProductVariant{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete product variant: %w", err)
	}
	return nil
}

// DeleteByProductID deletes all variants for a product
func (r *ProductVariantRepository) DeleteByProductID(productID uint) error {
	if err := r.db.Where("product_id = ?", productID).Delete(&model.ProductVariant{}).Error; err != nil {
		return fmt.Errorf("failed to delete product variants: %w", err)
	}
	return nil
}

// UpdateStock updates variant stock quantity
func (r *ProductVariantRepository) UpdateStock(id uint, quantity int) error {
	if err := r.db.Model(&model.ProductVariant{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"stock_quantity": quantity,
			"stock_status": func() string {
				if quantity > 0 {
					return "instock"
				}
				return "outofstock"
			}(),
		}).Error; err != nil {
		return fmt.Errorf("failed to update variant stock: %w", err)
	}
	return nil
}

// ExistsBySKU checks if a variant with the given SKU exists
func (r *ProductVariantRepository) ExistsBySKU(sku string) (bool, error) {
	var count int64
	if err := r.db.Model(&model.ProductVariant{}).Where("sku = ?", sku).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check variant SKU existence: %w", err)
	}
	return count > 0, nil
}

// ExistsBySKUExcludingID checks if a variant with the given SKU exists excluding specific ID
func (r *ProductVariantRepository) ExistsBySKUExcludingID(sku string, excludeID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.ProductVariant{}).Where("sku = ? AND id != ?", sku, excludeID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check variant SKU existence: %w", err)
	}
	return count > 0, nil
}

// BulkUpdateStatus updates the status of multiple variants
func (r *ProductVariantRepository) BulkUpdateStatus(ids []uint, isActive bool) error {
	if err := r.db.Model(&model.ProductVariant{}).Where("id IN ?", ids).Update("is_active", isActive).Error; err != nil {
		return fmt.Errorf("failed to bulk update variant status: %w", err)
	}
	return nil
}
