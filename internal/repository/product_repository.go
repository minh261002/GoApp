package repository

import (
	"encoding/json"
	"fmt"
	"strings"

	"go_app/internal/model"
	"go_app/pkg/database"

	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository() *ProductRepository {
	return &ProductRepository{
		db: database.GetDB(),
	}
}

// Create creates a new product
func (r *ProductRepository) Create(product *model.Product) error {
	// Check if SKU exists
	if product.SKU != "" {
		exists, err := r.ExistsBySKU(product.SKU)
		if err != nil {
			return fmt.Errorf("failed to check SKU existence: %w", err)
		}
		if exists {
			return fmt.Errorf("product with SKU '%s' already exists", product.SKU)
		}
	}

	// Check if slug exists
	exists, err := r.ExistsBySlug(product.Slug)
	if err != nil {
		return fmt.Errorf("failed to check slug existence: %w", err)
	}
	if exists {
		return fmt.Errorf("product with slug '%s' already exists", product.Slug)
	}

	// Convert images to JSON
	if len(product.Images) > 0 {
		imagesJSON, err := json.Marshal(product.Images)
		if err != nil {
			return fmt.Errorf("failed to marshal images: %w", err)
		}
		product.Images = string(imagesJSON)
	}

	if err := r.db.Create(product).Error; err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

// GetByID gets a product by ID
func (r *ProductRepository) GetByID(id uint) (*model.Product, error) {
	var product model.Product
	if err := r.db.Preload("Brand").Preload("Category").Preload("Variants").Preload("Attributes").
		Where("id = ?", id).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Parse images JSON
	if product.Images != "" {
		var images []string
		if err := json.Unmarshal([]byte(product.Images), &images); err == nil {
			product.Images = strings.Join(images, ",") // Convert back to string for compatibility
		}
	}

	// Parse variant attributes JSON
	for i := range product.Variants {
		if product.Variants[i].Attributes != "" {
			var attrs map[string]string
			if err := json.Unmarshal([]byte(product.Variants[i].Attributes), &attrs); err == nil {
				// Store as JSON string for now, will be handled in service
			}
		}
	}

	return &product, nil
}

// GetBySlug gets a product by slug
func (r *ProductRepository) GetBySlug(slug string) (*model.Product, error) {
	var product model.Product
	if err := r.db.Preload("Brand").Preload("Category").Preload("Variants").Preload("Attributes").
		Where("slug = ?", slug).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Parse images JSON
	if product.Images != "" {
		var images []string
		if err := json.Unmarshal([]byte(product.Images), &images); err == nil {
			product.Images = strings.Join(images, ",")
		}
	}

	return &product, nil
}

// GetBySKU gets a product by SKU
func (r *ProductRepository) GetBySKU(sku string) (*model.Product, error) {
	var product model.Product
	if err := r.db.Preload("Brand").Preload("Category").Preload("Variants").Preload("Attributes").
		Where("sku = ?", sku).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Parse images JSON
	if product.Images != "" {
		var images []string
		if err := json.Unmarshal([]byte(product.Images), &images); err == nil {
			product.Images = strings.Join(images, ",")
		}
	}

	return &product, nil
}

// GetAll gets all products with pagination and filters
func (r *ProductRepository) GetAll(page, limit int, search, sortBy, sortOrder string,
	status *model.ProductStatus, productType *model.ProductType, brandID *uint, categoryID *uint,
	isFeatured *bool, priceMin, priceMax *float64) ([]model.Product, int64, error) {

	var products []model.Product
	var total int64

	query := r.db.Model(&model.Product{})

	// Apply search filter
	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(sku) LIKE ?",
			searchTerm, searchTerm, searchTerm)
	}

	// Apply status filter
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	// Apply type filter
	if productType != nil {
		query = query.Where("type = ?", *productType)
	}

	// Apply brand filter
	if brandID != nil {
		query = query.Where("brand_id = ?", *brandID)
	}

	// Apply category filter
	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}

	// Apply featured filter
	if isFeatured != nil {
		query = query.Where("is_featured = ?", *isFeatured)
	}

	// Apply price range filter
	if priceMin != nil {
		query = query.Where("regular_price >= ?", *priceMin)
	}
	if priceMax != nil {
		query = query.Where("regular_price <= ?", *priceMax)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
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
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		query = query.Offset(offset).Limit(limit)
	}

	// Execute query with preloads
	if err := query.Preload("Brand").Preload("Category").Preload("Variants").Preload("Attributes").
		Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get products: %w", err)
	}

	// Parse images JSON for all products
	for i := range products {
		if products[i].Images != "" {
			var images []string
			if err := json.Unmarshal([]byte(products[i].Images), &images); err == nil {
				products[i].Images = strings.Join(images, ",")
			}
		}
	}

	return products, total, nil
}

// Update updates a product
func (r *ProductRepository) Update(product *model.Product) error {
	// Check if SKU exists (excluding current product)
	if product.SKU != "" {
		exists, err := r.ExistsBySKUExcludingID(product.SKU, product.ID)
		if err != nil {
			return fmt.Errorf("failed to check SKU existence: %w", err)
		}
		if exists {
			return fmt.Errorf("product with SKU '%s' already exists", product.SKU)
		}
	}

	// Check if slug exists (excluding current product)
	exists, err := r.ExistsBySlugExcludingID(product.Slug, product.ID)
	if err != nil {
		return fmt.Errorf("failed to check slug existence: %w", err)
	}
	if exists {
		return fmt.Errorf("product with slug '%s' already exists", product.Slug)
	}

	// Convert images to JSON
	if len(product.Images) > 0 {
		imagesJSON, err := json.Marshal(product.Images)
		if err != nil {
			return fmt.Errorf("failed to marshal images: %w", err)
		}
		product.Images = string(imagesJSON)
	}

	if err := r.db.Save(product).Error; err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

// Delete soft deletes a product
func (r *ProductRepository) Delete(id uint) error {
	if err := r.db.Delete(&model.Product{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	return nil
}

// GetFeatured gets featured products
func (r *ProductRepository) GetFeatured(limit int) ([]model.Product, error) {
	var products []model.Product
	query := r.db.Where("is_featured = ? AND status = ?", true, model.ProductStatusActive).
		Preload("Brand").Preload("Category").Preload("Variants").Preload("Attributes").
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to get featured products: %w", err)
	}

	// Parse images JSON
	for i := range products {
		if products[i].Images != "" {
			var images []string
			if err := json.Unmarshal([]byte(products[i].Images), &images); err == nil {
				products[i].Images = strings.Join(images, ",")
			}
		}
	}

	return products, nil
}

// GetByBrand gets products by brand
func (r *ProductRepository) GetByBrand(brandID uint, limit int) ([]model.Product, error) {
	var products []model.Product
	query := r.db.Where("brand_id = ? AND status = ?", brandID, model.ProductStatusActive).
		Preload("Brand").Preload("Category").Preload("Variants").Preload("Attributes").
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to get products by brand: %w", err)
	}

	return products, nil
}

// GetByCategory gets products by category
func (r *ProductRepository) GetByCategory(categoryID uint, limit int) ([]model.Product, error) {
	var products []model.Product
	query := r.db.Where("category_id = ? AND status = ?", categoryID, model.ProductStatusActive).
		Preload("Brand").Preload("Category").Preload("Variants").Preload("Attributes").
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to get products by category: %w", err)
	}

	return products, nil
}

// SearchProducts searches products by query
func (r *ProductRepository) SearchProducts(query string, limit int) ([]model.Product, error) {
	var products []model.Product
	searchTerm := "%" + strings.ToLower(query) + "%"

	if err := r.db.Where("(LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(sku) LIKE ?) AND status = ?",
		searchTerm, searchTerm, searchTerm, model.ProductStatusActive).
		Preload("Brand").Preload("Category").Preload("Variants").Preload("Attributes").
		Order("created_at DESC").Limit(limit).Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	return products, nil
}

// GetLowStockProducts gets products with low stock
func (r *ProductRepository) GetLowStockProducts() ([]model.Product, error) {
	var products []model.Product
	if err := r.db.Where("manage_stock = ? AND stock_quantity <= low_stock_threshold AND status = ?",
		true, model.ProductStatusActive).
		Preload("Brand").Preload("Category").
		Order("stock_quantity ASC").Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to get low stock products: %w", err)
	}

	return products, nil
}

// UpdateStock updates product stock quantity
func (r *ProductRepository) UpdateStock(id uint, quantity int) error {
	if err := r.db.Model(&model.Product{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"stock_quantity": quantity,
			"stock_status": func() string {
				if quantity > 0 {
					return "instock"
				}
				return "outofstock"
			}(),
		}).Error; err != nil {
		return fmt.Errorf("failed to update product stock: %w", err)
	}
	return nil
}

// ExistsBySKU checks if a product with the given SKU exists
func (r *ProductRepository) ExistsBySKU(sku string) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Product{}).Where("sku = ?", sku).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check product SKU existence: %w", err)
	}
	return count > 0, nil
}

// ExistsBySKUExcludingID checks if a product with the given SKU exists excluding specific ID
func (r *ProductRepository) ExistsBySKUExcludingID(sku string, excludeID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Product{}).Where("sku = ? AND id != ?", sku, excludeID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check product SKU existence: %w", err)
	}
	return count > 0, nil
}

// ExistsBySlug checks if a product with the given slug exists
func (r *ProductRepository) ExistsBySlug(slug string) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Product{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check product slug existence: %w", err)
	}
	return count > 0, nil
}

// ExistsBySlugExcludingID checks if a product with the given slug exists excluding specific ID
func (r *ProductRepository) ExistsBySlugExcludingID(slug string, excludeID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Product{}).Where("slug = ? AND id != ?", slug, excludeID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check product slug existence: %w", err)
	}
	return count > 0, nil
}

// BulkUpdateStatus updates the status of multiple products
func (r *ProductRepository) BulkUpdateStatus(ids []uint, status model.ProductStatus) error {
	if err := r.db.Model(&model.Product{}).Where("id IN ?", ids).Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to bulk update product status: %w", err)
	}
	return nil
}

// GetProductStats gets product statistics
func (r *ProductRepository) GetProductStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total products
	var totalProducts int64
	if err := r.db.Model(&model.Product{}).Count(&totalProducts).Error; err != nil {
		return nil, fmt.Errorf("failed to count total products: %w", err)
	}
	stats["total_products"] = totalProducts

	// Active products
	var activeProducts int64
	if err := r.db.Model(&model.Product{}).Where("status = ?", model.ProductStatusActive).Count(&activeProducts).Error; err != nil {
		return nil, fmt.Errorf("failed to count active products: %w", err)
	}
	stats["active_products"] = activeProducts

	// Featured products
	var featuredProducts int64
	if err := r.db.Model(&model.Product{}).Where("is_featured = ?", true).Count(&featuredProducts).Error; err != nil {
		return nil, fmt.Errorf("failed to count featured products: %w", err)
	}
	stats["featured_products"] = featuredProducts

	// Low stock products
	var lowStockProducts int64
	if err := r.db.Model(&model.Product{}).Where("manage_stock = ? AND stock_quantity <= low_stock_threshold", true).Count(&lowStockProducts).Error; err != nil {
		return nil, fmt.Errorf("failed to count low stock products: %w", err)
	}
	stats["low_stock_products"] = lowStockProducts

	// Products by type
	var simpleProducts, variableProducts int64
	if err := r.db.Model(&model.Product{}).Where("type = ?", model.ProductTypeSimple).Count(&simpleProducts).Error; err != nil {
		return nil, fmt.Errorf("failed to count simple products: %w", err)
	}
	if err := r.db.Model(&model.Product{}).Where("type = ?", model.ProductTypeVariable).Count(&variableProducts).Error; err != nil {
		return nil, fmt.Errorf("failed to count variable products: %w", err)
	}
	stats["simple_products"] = simpleProducts
	stats["variable_products"] = variableProducts

	return stats, nil
}
