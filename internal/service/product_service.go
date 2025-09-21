package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"go_app/internal/model"
	"go_app/internal/repository"
	"go_app/pkg/utils"
)

type ProductService struct {
	productRepo          *repository.ProductRepository
	productVariantRepo   *repository.ProductVariantRepository
	productAttributeRepo *repository.ProductAttributeRepository
}

func NewProductService() *ProductService {
	return &ProductService{
		productRepo:          repository.NewProductRepository(),
		productVariantRepo:   repository.NewProductVariantRepository(),
		productAttributeRepo: repository.NewProductAttributeRepository(),
	}
}

// CreateProduct creates a new product
func (s *ProductService) CreateProduct(req *model.ProductCreateRequest) (*model.ProductResponse, error) {
	// Generate slug from name
	slug := utils.GenerateSlug(req.Name)

	// Check if product with same slug exists
	exists, err := s.productRepo.ExistsBySlug(slug)
	if err != nil {
		return nil, fmt.Errorf("failed to check product slug: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("product with slug '%s' already exists", slug)
	}

	// Check if SKU exists
	if req.SKU != "" {
		exists, err := s.productRepo.ExistsBySKU(req.SKU)
		if err != nil {
			return nil, fmt.Errorf("failed to check product SKU: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("product with SKU '%s' already exists", req.SKU)
		}
	}

	// Set default values
	manageStock := true
	if req.ManageStock != nil {
		manageStock = *req.ManageStock
	}

	isFeatured := false
	if req.IsFeatured != nil {
		isFeatured = *req.IsFeatured
	}

	isDigital := false
	if req.IsDigital != nil {
		isDigital = *req.IsDigital
	}

	requiresShipping := true
	if req.RequiresShipping != nil {
		requiresShipping = *req.RequiresShipping
	}

	isDownloadable := false
	if req.IsDownloadable != nil {
		isDownloadable = *req.IsDownloadable
	}

	// Create product
	product := &model.Product{
		Name:              strings.TrimSpace(req.Name),
		Slug:              slug,
		Description:       strings.TrimSpace(req.Description),
		ShortDescription:  strings.TrimSpace(req.ShortDescription),
		SKU:               strings.TrimSpace(req.SKU),
		Type:              req.Type,
		Status:            req.Status,
		RegularPrice:      req.RegularPrice,
		SalePrice:         req.SalePrice,
		CostPrice:         req.CostPrice,
		ManageStock:       manageStock,
		StockQuantity:     req.StockQuantity,
		LowStockThreshold: req.LowStockThreshold,
		StockStatus:       req.StockStatus,
		Weight:            req.Weight,
		Length:            req.Length,
		Width:             req.Width,
		Height:            req.Height,
		Images:            strings.Join(req.Images, ","), // Will be converted to JSON in repository
		FeaturedImage:     strings.TrimSpace(req.FeaturedImage),
		MetaTitle:         strings.TrimSpace(req.MetaTitle),
		MetaDescription:   strings.TrimSpace(req.MetaDescription),
		MetaKeywords:      strings.TrimSpace(req.MetaKeywords),
		BrandID:           req.BrandID,
		CategoryID:        req.CategoryID,
		IsFeatured:        isFeatured,
		IsDigital:         isDigital,
		RequiresShipping:  requiresShipping,
		IsDownloadable:    isDownloadable,
	}

	// Create product
	if err := s.productRepo.Create(product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	// Create variants if product type is variable
	if product.Type == model.ProductTypeVariable && len(req.Variants) > 0 {
		if err := s.createProductVariants(product.ID, req.Variants); err != nil {
			// Rollback product creation
			s.productRepo.Delete(product.ID)
			return nil, fmt.Errorf("failed to create product variants: %w", err)
		}
	}

	// Create attributes
	if len(req.Attributes) > 0 {
		if err := s.createProductAttributes(product.ID, req.Attributes); err != nil {
			// Rollback product creation
			s.productRepo.Delete(product.ID)
			return nil, fmt.Errorf("failed to create product attributes: %w", err)
		}
	}

	// Get created product with all relations
	createdProduct, err := s.productRepo.GetByID(product.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created product: %w", err)
	}

	response := s.convertToResponse(createdProduct)
	return &response, nil
}

// GetProductByID gets a product by ID
func (s *ProductService) GetProductByID(id uint) (*model.ProductResponse, error) {
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := s.convertToResponse(product)
	return &response, nil
}

// GetProductBySlug gets a product by slug
func (s *ProductService) GetProductBySlug(slug string) (*model.ProductResponse, error) {
	product, err := s.productRepo.GetBySlug(slug)
	if err != nil {
		return nil, err
	}

	response := s.convertToResponse(product)
	return &response, nil
}

// GetProductBySKU gets a product by SKU
func (s *ProductService) GetProductBySKU(sku string) (*model.ProductResponse, error) {
	product, err := s.productRepo.GetBySKU(sku)
	if err != nil {
		return nil, err
	}

	response := s.convertToResponse(product)
	return &response, nil
}

// GetAllProducts gets all products with pagination and filters
func (s *ProductService) GetAllProducts(page, limit int, search, sortBy, sortOrder string,
	status *model.ProductStatus, productType *model.ProductType, brandID *uint, categoryID *uint,
	isFeatured *bool, priceMin, priceMax *float64) ([]model.ProductResponse, int64, error) {

	products, total, err := s.productRepo.GetAll(page, limit, search, sortBy, sortOrder,
		status, productType, brandID, categoryID, isFeatured, priceMin, priceMax)
	if err != nil {
		return nil, 0, err
	}

	// Convert to response format
	responses := make([]model.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = s.convertToResponse(&product)
	}

	return responses, total, nil
}

// UpdateProduct updates a product
func (s *ProductService) UpdateProduct(id uint, req *model.ProductUpdateRequest) (*model.ProductResponse, error) {
	// Get existing product
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != "" {
		// Check if new name conflicts with existing products
		exists, err := s.productRepo.ExistsBySlugExcludingID(utils.GenerateSlug(req.Name), id)
		if err != nil {
			return nil, fmt.Errorf("failed to check product slug: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("product with slug '%s' already exists", utils.GenerateSlug(req.Name))
		}

		product.Name = strings.TrimSpace(req.Name)
		product.Slug = utils.GenerateSlug(req.Name)
	}

	if req.Description != "" {
		product.Description = strings.TrimSpace(req.Description)
	}

	if req.ShortDescription != "" {
		product.ShortDescription = strings.TrimSpace(req.ShortDescription)
	}

	if req.SKU != "" {
		// Check if new SKU conflicts with existing products
		exists, err := s.productRepo.ExistsBySKUExcludingID(req.SKU, id)
		if err != nil {
			return nil, fmt.Errorf("failed to check product SKU: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("product with SKU '%s' already exists", req.SKU)
		}
		product.SKU = strings.TrimSpace(req.SKU)
	}

	if req.Type != "" {
		product.Type = req.Type
	}

	if req.Status != "" {
		product.Status = req.Status
	}

	if req.RegularPrice != nil {
		product.RegularPrice = *req.RegularPrice
	}

	if req.SalePrice != nil {
		product.SalePrice = req.SalePrice
	}

	if req.CostPrice != nil {
		product.CostPrice = req.CostPrice
	}

	if req.ManageStock != nil {
		product.ManageStock = *req.ManageStock
	}

	if req.StockQuantity != nil {
		product.StockQuantity = *req.StockQuantity
	}

	if req.LowStockThreshold != nil {
		product.LowStockThreshold = *req.LowStockThreshold
	}

	if req.StockStatus != "" {
		product.StockStatus = req.StockStatus
	}

	if req.Weight != nil {
		product.Weight = req.Weight
	}

	if req.Length != nil {
		product.Length = req.Length
	}

	if req.Width != nil {
		product.Width = req.Width
	}

	if req.Height != nil {
		product.Height = req.Height
	}

	if len(req.Images) > 0 {
		product.Images = strings.Join(req.Images, ",")
	}

	if req.FeaturedImage != "" {
		product.FeaturedImage = strings.TrimSpace(req.FeaturedImage)
	}

	if req.MetaTitle != "" {
		product.MetaTitle = strings.TrimSpace(req.MetaTitle)
	}

	if req.MetaDescription != "" {
		product.MetaDescription = strings.TrimSpace(req.MetaDescription)
	}

	if req.MetaKeywords != "" {
		product.MetaKeywords = strings.TrimSpace(req.MetaKeywords)
	}

	if req.BrandID != nil {
		product.BrandID = req.BrandID
	}

	if req.CategoryID != nil {
		product.CategoryID = req.CategoryID
	}

	if req.IsFeatured != nil {
		product.IsFeatured = *req.IsFeatured
	}

	if req.IsDigital != nil {
		product.IsDigital = *req.IsDigital
	}

	if req.RequiresShipping != nil {
		product.RequiresShipping = *req.RequiresShipping
	}

	if req.IsDownloadable != nil {
		product.IsDownloadable = *req.IsDownloadable
	}

	// Update product
	if err := s.productRepo.Update(product); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	response := s.convertToResponse(product)
	return &response, nil
}

// DeleteProduct soft deletes a product
func (s *ProductService) DeleteProduct(id uint) error {
	// Check if product exists
	_, err := s.productRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Delete product (variants and attributes will be cascade deleted)
	if err := s.productRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

// GetFeaturedProducts gets featured products
func (s *ProductService) GetFeaturedProducts(limit int) ([]model.ProductResponse, error) {
	products, err := s.productRepo.GetFeatured(limit)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responses := make([]model.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = s.convertToResponse(&product)
	}

	return responses, nil
}

// GetProductsByBrand gets products by brand
func (s *ProductService) GetProductsByBrand(brandID uint, limit int) ([]model.ProductResponse, error) {
	products, err := s.productRepo.GetByBrand(brandID, limit)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responses := make([]model.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = s.convertToResponse(&product)
	}

	return responses, nil
}

// GetProductsByCategory gets products by category
func (s *ProductService) GetProductsByCategory(categoryID uint, limit int) ([]model.ProductResponse, error) {
	products, err := s.productRepo.GetByCategory(categoryID, limit)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responses := make([]model.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = s.convertToResponse(&product)
	}

	return responses, nil
}

// SearchProducts searches products by query
func (s *ProductService) SearchProducts(query string, limit int) ([]model.ProductResponse, error) {
	products, err := s.productRepo.SearchProducts(query, limit)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responses := make([]model.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = s.convertToResponse(&product)
	}

	return responses, nil
}

// GetLowStockProducts gets products with low stock
func (s *ProductService) GetLowStockProducts() ([]model.ProductResponse, error) {
	products, err := s.productRepo.GetLowStockProducts()
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responses := make([]model.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = s.convertToResponse(&product)
	}

	return responses, nil
}

// UpdateProductStock updates product stock quantity
func (s *ProductService) UpdateProductStock(id uint, quantity int) (*model.ProductResponse, error) {
	// Update stock
	if err := s.productRepo.UpdateStock(id, quantity); err != nil {
		return nil, fmt.Errorf("failed to update product stock: %w", err)
	}

	// Get updated product
	updatedProduct, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := s.convertToResponse(updatedProduct)
	return &response, nil
}

// UpdateProductStatus updates the status of a product
func (s *ProductService) UpdateProductStatus(id uint, status model.ProductStatus) (*model.ProductResponse, error) {
	// Get existing product
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update status
	product.Status = status

	// Update product
	if err := s.productRepo.Update(product); err != nil {
		return nil, fmt.Errorf("failed to update product status: %w", err)
	}

	response := s.convertToResponse(product)
	return &response, nil
}

// BulkUpdateProductStatus updates the status of multiple products
func (s *ProductService) BulkUpdateProductStatus(ids []uint, status model.ProductStatus) error {
	if len(ids) == 0 {
		return fmt.Errorf("no product IDs provided")
	}

	// Validate all products exist
	for _, id := range ids {
		_, err := s.productRepo.GetByID(id)
		if err != nil {
			return fmt.Errorf("product with ID %d not found: %w", id, err)
		}
	}

	// Bulk update status
	if err := s.productRepo.BulkUpdateStatus(ids, status); err != nil {
		return fmt.Errorf("failed to bulk update product status: %w", err)
	}

	return nil
}

// GetProductStats gets product statistics
func (s *ProductService) GetProductStats() (map[string]interface{}, error) {
	stats, err := s.productRepo.GetProductStats()
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// createProductVariants creates variants for a product
func (s *ProductService) createProductVariants(productID uint, variantReqs []model.ProductVariantCreateRequest) error {
	for _, variantReq := range variantReqs {
		// Check if variant SKU exists
		if variantReq.SKU != "" {
			exists, err := s.productVariantRepo.ExistsBySKU(variantReq.SKU)
			if err != nil {
				return fmt.Errorf("failed to check variant SKU: %w", err)
			}
			if exists {
				return fmt.Errorf("variant with SKU '%s' already exists", variantReq.SKU)
			}
		}

		// Set default values
		manageStock := true
		if variantReq.ManageStock != nil {
			manageStock = *variantReq.ManageStock
		}

		isActive := true
		if variantReq.IsActive != nil {
			isActive = *variantReq.IsActive
		}

		// Convert attributes to JSON
		attributesJSON := ""
		if len(variantReq.Attributes) > 0 {
			attrsJSON, err := json.Marshal(variantReq.Attributes)
			if err != nil {
				return fmt.Errorf("failed to marshal variant attributes: %w", err)
			}
			attributesJSON = string(attrsJSON)
		}

		variant := &model.ProductVariant{
			ProductID:     productID,
			Name:          strings.TrimSpace(variantReq.Name),
			SKU:           strings.TrimSpace(variantReq.SKU),
			Description:   strings.TrimSpace(variantReq.Description),
			RegularPrice:  variantReq.RegularPrice,
			SalePrice:     variantReq.SalePrice,
			CostPrice:     variantReq.CostPrice,
			StockQuantity: variantReq.StockQuantity,
			StockStatus:   variantReq.StockStatus,
			ManageStock:   manageStock,
			Image:         strings.TrimSpace(variantReq.Image),
			Attributes:    attributesJSON,
			IsActive:      isActive,
			SortOrder:     variantReq.SortOrder,
		}

		if err := s.productVariantRepo.Create(variant); err != nil {
			return fmt.Errorf("failed to create variant: %w", err)
		}
	}

	return nil
}

// createProductAttributes creates attributes for a product
func (s *ProductService) createProductAttributes(productID uint, attributeReqs []model.ProductAttributeCreateRequest) error {
	attributes := make([]model.ProductAttribute, len(attributeReqs))

	for i, attrReq := range attributeReqs {
		// Set default values
		isVisible := true
		if attrReq.IsVisible != nil {
			isVisible = *attrReq.IsVisible
		}

		isVariation := false
		if attrReq.IsVariation != nil {
			isVariation = *attrReq.IsVariation
		}

		attributes[i] = model.ProductAttribute{
			ProductID:   productID,
			Name:        strings.TrimSpace(attrReq.Name),
			Value:       strings.TrimSpace(attrReq.Value),
			Slug:        strings.TrimSpace(attrReq.Slug),
			IsVisible:   isVisible,
			IsVariation: isVariation,
			SortOrder:   attrReq.SortOrder,
		}
	}

	if err := s.productAttributeRepo.BulkCreate(attributes); err != nil {
		return fmt.Errorf("failed to create product attributes: %w", err)
	}

	return nil
}

// convertToResponse converts Product to ProductResponse
func (s *ProductService) convertToResponse(product *model.Product) model.ProductResponse {
	response := product.ToResponse()

	// Parse images JSON
	if product.Images != "" {
		var images []string
		if err := json.Unmarshal([]byte(product.Images), &images); err == nil {
			response.Images = images
		} else {
			// Fallback to comma-separated string
			response.Images = strings.Split(product.Images, ",")
		}
	}

	// Parse variant attributes JSON
	for _, variant := range response.Variants {
		if variant.Attributes != nil {
			// Already parsed in repository
		}
	}

	return response
}

// ===== PRODUCT VARIANTS SERVICE METHODS =====

// GetProductVariants gets all variants for a product
func (s *ProductService) GetProductVariants(productID uint) ([]model.ProductVariantResponse, error) {
	// Check if product exists
	_, err := s.productRepo.GetByID(productID)
	if err != nil {
		return nil, fmt.Errorf("product not found")
	}

	variants, err := s.productVariantRepo.GetByProductID(productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product variants: %w", err)
	}

	responses := make([]model.ProductVariantResponse, len(variants))
	for i, variant := range variants {
		responses[i] = s.convertVariantToResponse(&variant)
	}

	return responses, nil
}

// GetProductVariant gets a specific product variant
func (s *ProductService) GetProductVariant(productID, variantID uint) (*model.ProductVariantResponse, error) {
	// Check if product exists
	_, err := s.productRepo.GetByID(productID)
	if err != nil {
		return nil, fmt.Errorf("product not found")
	}

	variant, err := s.productVariantRepo.GetByID(variantID)
	if err != nil {
		return nil, fmt.Errorf("product variant not found")
	}

	// Verify variant belongs to product
	if variant.ProductID != productID {
		return nil, fmt.Errorf("product variant not found")
	}

	response := s.convertVariantToResponse(variant)
	return &response, nil
}

// CreateProductVariant creates a new product variant
func (s *ProductService) CreateProductVariant(productID uint, req *model.ProductVariantCreateRequest) (*model.ProductVariantResponse, error) {
	// Check if product exists
	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		return nil, fmt.Errorf("product not found")
	}

	// Check if product is variable type
	if product.Type != model.ProductTypeVariable {
		return nil, fmt.Errorf("product must be variable type to add variants")
	}

	// Check if variant SKU exists
	if req.SKU != "" {
		exists, err := s.productVariantRepo.ExistsBySKU(req.SKU)
		if err != nil {
			return nil, fmt.Errorf("failed to check variant SKU: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("variant with SKU '%s' already exists", req.SKU)
		}
	}

	// Set default values
	manageStock := true
	if req.ManageStock != nil {
		manageStock = *req.ManageStock
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Convert attributes to JSON
	attributesJSON := ""
	if len(req.Attributes) > 0 {
		attrsJSON, err := json.Marshal(req.Attributes)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal variant attributes: %w", err)
		}
		attributesJSON = string(attrsJSON)
	}

	variant := &model.ProductVariant{
		ProductID:     productID,
		Name:          strings.TrimSpace(req.Name),
		SKU:           strings.TrimSpace(req.SKU),
		Description:   strings.TrimSpace(req.Description),
		RegularPrice:  req.RegularPrice,
		SalePrice:     req.SalePrice,
		CostPrice:     req.CostPrice,
		StockQuantity: req.StockQuantity,
		StockStatus:   req.StockStatus,
		ManageStock:   manageStock,
		Image:         strings.TrimSpace(req.Image),
		Attributes:    attributesJSON,
		IsActive:      isActive,
		SortOrder:     req.SortOrder,
	}

	if err := s.productVariantRepo.Create(variant); err != nil {
		return nil, fmt.Errorf("failed to create product variant: %w", err)
	}

	response := s.convertVariantToResponse(variant)
	return &response, nil
}

// UpdateProductVariant updates a product variant
func (s *ProductService) UpdateProductVariant(productID, variantID uint, req *model.ProductVariantUpdateRequest) (*model.ProductVariantResponse, error) {
	// Check if product exists
	_, err := s.productRepo.GetByID(productID)
	if err != nil {
		return nil, fmt.Errorf("product not found")
	}

	// Get existing variant
	variant, err := s.productVariantRepo.GetByID(variantID)
	if err != nil {
		return nil, fmt.Errorf("product variant not found")
	}

	// Verify variant belongs to product
	if variant.ProductID != productID {
		return nil, fmt.Errorf("product variant not found")
	}

	// Check if variant SKU exists (if changing SKU)
	if req.SKU != "" && req.SKU != variant.SKU {
		exists, err := s.productVariantRepo.ExistsBySKU(req.SKU)
		if err != nil {
			return nil, fmt.Errorf("failed to check variant SKU: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("variant with SKU '%s' already exists", req.SKU)
		}
	}

	// Update fields
	if req.Name != "" {
		variant.Name = strings.TrimSpace(req.Name)
	}
	if req.SKU != "" {
		variant.SKU = strings.TrimSpace(req.SKU)
	}
	if req.Description != "" {
		variant.Description = strings.TrimSpace(req.Description)
	}
	if req.RegularPrice != nil {
		variant.RegularPrice = *req.RegularPrice
	}
	if req.SalePrice != nil {
		variant.SalePrice = req.SalePrice
	}
	if req.CostPrice != nil {
		variant.CostPrice = req.CostPrice
	}
	if req.StockQuantity != nil {
		variant.StockQuantity = *req.StockQuantity
	}
	if req.StockStatus != "" {
		variant.StockStatus = req.StockStatus
	}
	if req.ManageStock != nil {
		variant.ManageStock = *req.ManageStock
	}
	if req.Image != "" {
		variant.Image = strings.TrimSpace(req.Image)
	}
	if req.Attributes != nil {
		attrsJSON, err := json.Marshal(req.Attributes)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal variant attributes: %w", err)
		}
		variant.Attributes = string(attrsJSON)
	}
	if req.IsActive != nil {
		variant.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		variant.SortOrder = *req.SortOrder
	}

	if err := s.productVariantRepo.Update(variant); err != nil {
		return nil, fmt.Errorf("failed to update product variant: %w", err)
	}

	response := s.convertVariantToResponse(variant)
	return &response, nil
}

// DeleteProductVariant deletes a product variant
func (s *ProductService) DeleteProductVariant(productID, variantID uint) error {
	// Check if product exists
	_, err := s.productRepo.GetByID(productID)
	if err != nil {
		return fmt.Errorf("product not found")
	}

	// Get existing variant
	variant, err := s.productVariantRepo.GetByID(variantID)
	if err != nil {
		return fmt.Errorf("product variant not found")
	}

	// Verify variant belongs to product
	if variant.ProductID != productID {
		return fmt.Errorf("product variant not found")
	}

	if err := s.productVariantRepo.Delete(variantID); err != nil {
		return fmt.Errorf("failed to delete product variant: %w", err)
	}

	return nil
}

// UpdateProductVariantStock updates stock for a product variant
func (s *ProductService) UpdateProductVariantStock(productID, variantID uint, req *model.ProductVariantStockUpdateRequest) (*model.ProductVariantResponse, error) {
	// Check if product exists
	_, err := s.productRepo.GetByID(productID)
	if err != nil {
		return nil, fmt.Errorf("product not found")
	}

	// Get existing variant
	variant, err := s.productVariantRepo.GetByID(variantID)
	if err != nil {
		return nil, fmt.Errorf("product variant not found")
	}

	// Verify variant belongs to product
	if variant.ProductID != productID {
		return nil, fmt.Errorf("product variant not found")
	}

	// Update stock fields
	variant.StockQuantity = req.StockQuantity
	variant.StockStatus = req.StockStatus
	if req.ManageStock != nil {
		variant.ManageStock = *req.ManageStock
	}

	if err := s.productVariantRepo.Update(variant); err != nil {
		return nil, fmt.Errorf("failed to update variant stock: %w", err)
	}

	response := s.convertVariantToResponse(variant)
	return &response, nil
}

// UpdateProductVariantStatus updates status for a product variant
func (s *ProductService) UpdateProductVariantStatus(productID, variantID uint, req *model.ProductVariantStatusUpdateRequest) (*model.ProductVariantResponse, error) {
	// Check if product exists
	_, err := s.productRepo.GetByID(productID)
	if err != nil {
		return nil, fmt.Errorf("product not found")
	}

	// Get existing variant
	variant, err := s.productVariantRepo.GetByID(variantID)
	if err != nil {
		return nil, fmt.Errorf("product variant not found")
	}

	// Verify variant belongs to product
	if variant.ProductID != productID {
		return nil, fmt.Errorf("product variant not found")
	}

	// Update status
	variant.IsActive = *req.IsActive

	if err := s.productVariantRepo.Update(variant); err != nil {
		return nil, fmt.Errorf("failed to update variant status: %w", err)
	}

	response := s.convertVariantToResponse(variant)
	return &response, nil
}

// convertVariantToResponse converts ProductVariant to ProductVariantResponse
func (s *ProductService) convertVariantToResponse(variant *model.ProductVariant) model.ProductVariantResponse {
	response := model.ProductVariantResponse{
		ID:          variant.ID,
		ProductID:   variant.ProductID,
		Name:        variant.Name,
		SKU:         variant.SKU,
		Description: variant.Description,
		RegularPrice: variant.RegularPrice,
		SalePrice:   variant.SalePrice,
		CostPrice:   variant.CostPrice,
		StockQuantity: variant.StockQuantity,
		StockStatus: variant.StockStatus,
		ManageStock: variant.ManageStock,
		Image:       variant.Image,
		IsActive:    variant.IsActive,
		SortOrder:   variant.SortOrder,
		CreatedAt:   variant.CreatedAt,
		UpdatedAt:   variant.UpdatedAt,
	}

	// Parse attributes JSON
	if variant.Attributes != "" {
		var attrs map[string]string
		if err := json.Unmarshal([]byte(variant.Attributes), &attrs); err == nil {
			response.Attributes = attrs
		} else {
			response.Attributes = make(map[string]string)
		}
	} else {
		response.Attributes = make(map[string]string)
	}

	return response
}
