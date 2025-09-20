package repository

import (
	"fmt"

	"go_app/internal/model"
	"go_app/pkg/database"

	"gorm.io/gorm"
)

type ProductAttributeRepository struct {
	db *gorm.DB
}

func NewProductAttributeRepository() *ProductAttributeRepository {
	return &ProductAttributeRepository{
		db: database.GetDB(),
	}
}

// Create creates a new product attribute
func (r *ProductAttributeRepository) Create(attribute *model.ProductAttribute) error {
	if err := r.db.Create(attribute).Error; err != nil {
		return fmt.Errorf("failed to create product attribute: %w", err)
	}
	return nil
}

// GetByID gets a product attribute by ID
func (r *ProductAttributeRepository) GetByID(id uint) (*model.ProductAttribute, error) {
	var attribute model.ProductAttribute
	if err := r.db.Preload("Product").Where("id = ?", id).First(&attribute).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("product attribute not found")
		}
		return nil, fmt.Errorf("failed to get product attribute: %w", err)
	}
	return &attribute, nil
}

// GetByProductID gets all attributes for a product
func (r *ProductAttributeRepository) GetByProductID(productID uint) ([]model.ProductAttribute, error) {
	var attributes []model.ProductAttribute
	if err := r.db.Where("product_id = ?", productID).
		Order("sort_order ASC, name ASC").Find(&attributes).Error; err != nil {
		return nil, fmt.Errorf("failed to get product attributes: %w", err)
	}
	return attributes, nil
}

// GetByProductIDAndSlug gets attributes by product ID and slug
func (r *ProductAttributeRepository) GetByProductIDAndSlug(productID uint, slug string) ([]model.ProductAttribute, error) {
	var attributes []model.ProductAttribute
	if err := r.db.Where("product_id = ? AND slug = ?", productID, slug).
		Order("sort_order ASC, value ASC").Find(&attributes).Error; err != nil {
		return nil, fmt.Errorf("failed to get product attributes by slug: %w", err)
	}
	return attributes, nil
}

// GetVariationAttributes gets variation attributes for a product
func (r *ProductAttributeRepository) GetVariationAttributes(productID uint) ([]model.ProductAttribute, error) {
	var attributes []model.ProductAttribute
	if err := r.db.Where("product_id = ? AND is_variation = ?", productID, true).
		Order("sort_order ASC, name ASC").Find(&attributes).Error; err != nil {
		return nil, fmt.Errorf("failed to get variation attributes: %w", err)
	}
	return attributes, nil
}

// GetVisibleAttributes gets visible attributes for a product
func (r *ProductAttributeRepository) GetVisibleAttributes(productID uint) ([]model.ProductAttribute, error) {
	var attributes []model.ProductAttribute
	if err := r.db.Where("product_id = ? AND is_visible = ?", productID, true).
		Order("sort_order ASC, name ASC").Find(&attributes).Error; err != nil {
		return nil, fmt.Errorf("failed to get visible attributes: %w", err)
	}
	return attributes, nil
}

// Update updates a product attribute
func (r *ProductAttributeRepository) Update(attribute *model.ProductAttribute) error {
	if err := r.db.Save(attribute).Error; err != nil {
		return fmt.Errorf("failed to update product attribute: %w", err)
	}
	return nil
}

// Delete soft deletes a product attribute
func (r *ProductAttributeRepository) Delete(id uint) error {
	if err := r.db.Delete(&model.ProductAttribute{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete product attribute: %w", err)
	}
	return nil
}

// DeleteByProductID deletes all attributes for a product
func (r *ProductAttributeRepository) DeleteByProductID(productID uint) error {
	if err := r.db.Where("product_id = ?", productID).Delete(&model.ProductAttribute{}).Error; err != nil {
		return fmt.Errorf("failed to delete product attributes: %w", err)
	}
	return nil
}

// BulkCreate creates multiple attributes at once
func (r *ProductAttributeRepository) BulkCreate(attributes []model.ProductAttribute) error {
	if len(attributes) == 0 {
		return nil
	}

	if err := r.db.Create(&attributes).Error; err != nil {
		return fmt.Errorf("failed to bulk create product attributes: %w", err)
	}
	return nil
}

// GetUniqueAttributeNames gets unique attribute names across all products
func (r *ProductAttributeRepository) GetUniqueAttributeNames() ([]string, error) {
	var names []string
	if err := r.db.Model(&model.ProductAttribute{}).
		Distinct("name").Pluck("name", &names).Error; err != nil {
		return nil, fmt.Errorf("failed to get unique attribute names: %w", err)
	}
	return names, nil
}

// GetAttributeValues gets all values for a specific attribute name
func (r *ProductAttributeRepository) GetAttributeValues(attributeName string) ([]string, error) {
	var values []string
	if err := r.db.Model(&model.ProductAttribute{}).
		Where("name = ?", attributeName).
		Distinct("value").Pluck("value", &values).Error; err != nil {
		return nil, fmt.Errorf("failed to get attribute values: %w", err)
	}
	return values, nil
}

// GetProductsByAttributeValue gets products that have a specific attribute value
func (r *ProductAttributeRepository) GetProductsByAttributeValue(attributeName, attributeValue string) ([]uint, error) {
	var productIDs []uint
	if err := r.db.Model(&model.ProductAttribute{}).
		Where("name = ? AND value = ?", attributeName, attributeValue).
		Pluck("product_id", &productIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to get products by attribute value: %w", err)
	}
	return productIDs, nil
}
