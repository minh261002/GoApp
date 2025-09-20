package model

import (
	"time"

	"gorm.io/gorm"
)

// ProductType represents the type of product
type ProductType string

const (
	ProductTypeSimple   ProductType = "simple"   // Sản phẩm đơn giản
	ProductTypeVariable ProductType = "variable" // Sản phẩm có biến thể
)

// ProductStatus represents the status of product
type ProductStatus string

const (
	ProductStatusDraft    ProductStatus = "draft"    // Bản nháp
	ProductStatusActive   ProductStatus = "active"   // Đang bán
	ProductStatusInactive ProductStatus = "inactive" // Ngừng bán
	ProductStatusArchived ProductStatus = "archived" // Lưu trữ
)

// Product represents a product in the system
type Product struct {
	ID               uint          `json:"id" gorm:"primaryKey"`
	Name             string        `json:"name" gorm:"size:255;not null" validate:"required,min=2,max=255"`
	Slug             string        `json:"slug" gorm:"size:300;not null;uniqueIndex" validate:"required,min=2,max=300"`
	Description      string        `json:"description" gorm:"type:text"`
	ShortDescription string        `json:"short_description" gorm:"size:500"`
	SKU              string        `json:"sku" gorm:"size:100;uniqueIndex"`
	Type             ProductType   `json:"type" gorm:"type:enum('simple','variable');default:'simple'"`
	Status           ProductStatus `json:"status" gorm:"type:enum('draft','active','inactive','archived');default:'draft'"`

	// Pricing
	RegularPrice float64  `json:"regular_price" gorm:"type:decimal(10,2);default:0"`
	SalePrice    *float64 `json:"sale_price" gorm:"type:decimal(10,2)"`
	CostPrice    *float64 `json:"cost_price" gorm:"type:decimal(10,2)"`

	// Inventory
	ManageStock       bool   `json:"manage_stock" gorm:"default:true"`
	StockQuantity     int    `json:"stock_quantity" gorm:"default:0"`
	LowStockThreshold int    `json:"low_stock_threshold" gorm:"default:5"`
	StockStatus       string `json:"stock_status" gorm:"size:20;default:'instock'"` // instock, outofstock, onbackorder

	// Dimensions & Weight
	Weight *float64 `json:"weight" gorm:"type:decimal(8,2)"`
	Length *float64 `json:"length" gorm:"type:decimal(8,2)"`
	Width  *float64 `json:"width" gorm:"type:decimal(8,2)"`
	Height *float64 `json:"height" gorm:"type:decimal(8,2)"`

	// Media
	Images        string `json:"images" gorm:"type:text"` // JSON array of image URLs
	FeaturedImage string `json:"featured_image" gorm:"size:500"`

	// SEO
	MetaTitle       string `json:"meta_title" gorm:"size:255"`
	MetaDescription string `json:"meta_description" gorm:"size:500"`
	MetaKeywords    string `json:"meta_keywords" gorm:"size:500"`

	// Relationships
	BrandID    *uint     `json:"brand_id" gorm:"index"`
	Brand      *Brand    `json:"brand,omitempty" gorm:"foreignKey:BrandID"`
	CategoryID *uint     `json:"category_id" gorm:"index"`
	Category   *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`

	// Variants (for variable products)
	Variants []ProductVariant `json:"variants,omitempty" gorm:"foreignKey:ProductID"`

	// Attributes
	Attributes []ProductAttribute `json:"attributes,omitempty" gorm:"foreignKey:ProductID"`

	// Settings
	IsFeatured       bool `json:"is_featured" gorm:"default:false"`
	IsDigital        bool `json:"is_digital" gorm:"default:false"`
	RequiresShipping bool `json:"requires_shipping" gorm:"default:true"`
	IsDownloadable   bool `json:"is_downloadable" gorm:"default:false"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// ProductVariant represents a variant of a variable product
type ProductVariant struct {
	ID        uint     `json:"id" gorm:"primaryKey"`
	ProductID uint     `json:"product_id" gorm:"not null;index"`
	Product   *Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`

	// Variant Info
	Name        string `json:"name" gorm:"size:255;not null"`
	SKU         string `json:"sku" gorm:"size:100;uniqueIndex"`
	Description string `json:"description" gorm:"type:text"`

	// Pricing
	RegularPrice float64  `json:"regular_price" gorm:"type:decimal(10,2);not null"`
	SalePrice    *float64 `json:"sale_price" gorm:"type:decimal(10,2)"`
	CostPrice    *float64 `json:"cost_price" gorm:"type:decimal(10,2)"`

	// Inventory
	StockQuantity int    `json:"stock_quantity" gorm:"default:0"`
	StockStatus   string `json:"stock_status" gorm:"size:20;default:'instock'"`
	ManageStock   bool   `json:"manage_stock" gorm:"default:true"`

	// Media
	Image string `json:"image" gorm:"size:500"`

	// Attributes (JSON format: {"size": "L", "color": "Red"})
	Attributes string `json:"attributes" gorm:"type:text"`

	// Settings
	IsActive  bool `json:"is_active" gorm:"default:true"`
	SortOrder int  `json:"sort_order" gorm:"default:0"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// ProductAttribute represents an attribute of a product
type ProductAttribute struct {
	ID        uint     `json:"id" gorm:"primaryKey"`
	ProductID uint     `json:"product_id" gorm:"not null;index"`
	Product   *Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`

	// Attribute Info
	Name  string `json:"name" gorm:"size:100;not null"`  // e.g., "Size", "Color"
	Value string `json:"value" gorm:"size:255;not null"` // e.g., "L", "Red"
	Slug  string `json:"slug" gorm:"size:120;not null"`  // e.g., "size", "color"

	// Settings
	IsVisible   bool `json:"is_visible" gorm:"default:true"`
	IsVariation bool `json:"is_variation" gorm:"default:false"` // Used for variations
	SortOrder   int  `json:"sort_order" gorm:"default:0"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// ProductCreateRequest represents the request to create a product
type ProductCreateRequest struct {
	Name             string        `json:"name" validate:"required,min=2,max=255"`
	Description      string        `json:"description"`
	ShortDescription string        `json:"short_description"`
	SKU              string        `json:"sku"`
	Type             ProductType   `json:"type" validate:"oneof=simple variable"`
	Status           ProductStatus `json:"status" validate:"oneof=draft active inactive archived"`

	// Pricing
	RegularPrice float64  `json:"regular_price" validate:"min=0"`
	SalePrice    *float64 `json:"sale_price" validate:"omitempty,min=0"`
	CostPrice    *float64 `json:"cost_price" validate:"omitempty,min=0"`

	// Inventory
	ManageStock       *bool  `json:"manage_stock"`
	StockQuantity     int    `json:"stock_quantity" validate:"min=0"`
	LowStockThreshold int    `json:"low_stock_threshold" validate:"min=0"`
	StockStatus       string `json:"stock_status" validate:"oneof=instock outofstock onbackorder"`

	// Dimensions & Weight
	Weight *float64 `json:"weight" validate:"omitempty,min=0"`
	Length *float64 `json:"length" validate:"omitempty,min=0"`
	Width  *float64 `json:"width" validate:"omitempty,min=0"`
	Height *float64 `json:"height" validate:"omitempty,min=0"`

	// Media
	Images        []string `json:"images"`
	FeaturedImage string   `json:"featured_image"`

	// SEO
	MetaTitle       string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`
	MetaKeywords    string `json:"meta_keywords"`

	// Relationships
	BrandID    *uint `json:"brand_id"`
	CategoryID *uint `json:"category_id"`

	// Variants (for variable products)
	Variants []ProductVariantCreateRequest `json:"variants"`

	// Attributes
	Attributes []ProductAttributeCreateRequest `json:"attributes"`

	// Settings
	IsFeatured       *bool `json:"is_featured"`
	IsDigital        *bool `json:"is_digital"`
	RequiresShipping *bool `json:"requires_shipping"`
	IsDownloadable   *bool `json:"is_downloadable"`
}

// ProductVariantCreateRequest represents the request to create a product variant
type ProductVariantCreateRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	SKU         string `json:"sku"`
	Description string `json:"description"`

	// Pricing
	RegularPrice float64  `json:"regular_price" validate:"min=0"`
	SalePrice    *float64 `json:"sale_price" validate:"omitempty,min=0"`
	CostPrice    *float64 `json:"cost_price" validate:"omitempty,min=0"`

	// Inventory
	StockQuantity int    `json:"stock_quantity" validate:"min=0"`
	StockStatus   string `json:"stock_status" validate:"oneof=instock outofstock onbackorder"`
	ManageStock   *bool  `json:"manage_stock"`

	// Media
	Image string `json:"image"`

	// Attributes (map format: {"size": "L", "color": "Red"})
	Attributes map[string]string `json:"attributes"`

	// Settings
	IsActive  *bool `json:"is_active"`
	SortOrder int   `json:"sort_order"`
}

// ProductAttributeCreateRequest represents the request to create a product attribute
type ProductAttributeCreateRequest struct {
	Name  string `json:"name" validate:"required,min=1,max=100"`
	Value string `json:"value" validate:"required,min=1,max=255"`
	Slug  string `json:"slug" validate:"required,min=1,max=120"`

	// Settings
	IsVisible   *bool `json:"is_visible"`
	IsVariation *bool `json:"is_variation"`
	SortOrder   int   `json:"sort_order"`
}

// ProductUpdateRequest represents the request to update a product
type ProductUpdateRequest struct {
	Name             string        `json:"name" validate:"omitempty,min=2,max=255"`
	Description      string        `json:"description"`
	ShortDescription string        `json:"short_description"`
	SKU              string        `json:"sku"`
	Type             ProductType   `json:"type" validate:"omitempty,oneof=simple variable"`
	Status           ProductStatus `json:"status" validate:"omitempty,oneof=draft active inactive archived"`

	// Pricing
	RegularPrice *float64 `json:"regular_price" validate:"omitempty,min=0"`
	SalePrice    *float64 `json:"sale_price" validate:"omitempty,min=0"`
	CostPrice    *float64 `json:"cost_price" validate:"omitempty,min=0"`

	// Inventory
	ManageStock       *bool  `json:"manage_stock"`
	StockQuantity     *int   `json:"stock_quantity" validate:"omitempty,min=0"`
	LowStockThreshold *int   `json:"low_stock_threshold" validate:"omitempty,min=0"`
	StockStatus       string `json:"stock_status" validate:"omitempty,oneof=instock outofstock onbackorder"`

	// Dimensions & Weight
	Weight *float64 `json:"weight" validate:"omitempty,min=0"`
	Length *float64 `json:"length" validate:"omitempty,min=0"`
	Width  *float64 `json:"width" validate:"omitempty,min=0"`
	Height *float64 `json:"height" validate:"omitempty,min=0"`

	// Media
	Images        []string `json:"images"`
	FeaturedImage string   `json:"featured_image"`

	// SEO
	MetaTitle       string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`
	MetaKeywords    string `json:"meta_keywords"`

	// Relationships
	BrandID    *uint `json:"brand_id"`
	CategoryID *uint `json:"category_id"`

	// Settings
	IsFeatured       *bool `json:"is_featured"`
	IsDigital        *bool `json:"is_digital"`
	RequiresShipping *bool `json:"requires_shipping"`
	IsDownloadable   *bool `json:"is_downloadable"`
}

// ProductResponse represents the response for product data
type ProductResponse struct {
	ID               uint          `json:"id"`
	Name             string        `json:"name"`
	Slug             string        `json:"slug"`
	Description      string        `json:"description"`
	ShortDescription string        `json:"short_description"`
	SKU              string        `json:"sku"`
	Type             ProductType   `json:"type"`
	Status           ProductStatus `json:"status"`

	// Pricing
	RegularPrice float64  `json:"regular_price"`
	SalePrice    *float64 `json:"sale_price"`
	CostPrice    *float64 `json:"cost_price"`

	// Inventory
	ManageStock       bool   `json:"manage_stock"`
	StockQuantity     int    `json:"stock_quantity"`
	LowStockThreshold int    `json:"low_stock_threshold"`
	StockStatus       string `json:"stock_status"`

	// Dimensions & Weight
	Weight *float64 `json:"weight"`
	Length *float64 `json:"length"`
	Width  *float64 `json:"width"`
	Height *float64 `json:"height"`

	// Media
	Images        []string `json:"images"`
	FeaturedImage string   `json:"featured_image"`

	// SEO
	MetaTitle       string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`
	MetaKeywords    string `json:"meta_keywords"`

	// Relationships
	BrandID    *uint             `json:"brand_id"`
	Brand      *BrandResponse    `json:"brand,omitempty"`
	CategoryID *uint             `json:"category_id"`
	Category   *CategoryResponse `json:"category,omitempty"`

	// Variants
	Variants []ProductVariantResponse `json:"variants,omitempty"`

	// Attributes
	Attributes []ProductAttributeResponse `json:"attributes,omitempty"`

	// Settings
	IsFeatured       bool `json:"is_featured"`
	IsDigital        bool `json:"is_digital"`
	RequiresShipping bool `json:"requires_shipping"`
	IsDownloadable   bool `json:"is_downloadable"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProductVariantResponse represents the response for product variant data
type ProductVariantResponse struct {
	ID          uint   `json:"id"`
	ProductID   uint   `json:"product_id"`
	Name        string `json:"name"`
	SKU         string `json:"sku"`
	Description string `json:"description"`

	// Pricing
	RegularPrice float64  `json:"regular_price"`
	SalePrice    *float64 `json:"sale_price"`
	CostPrice    *float64 `json:"cost_price"`

	// Inventory
	StockQuantity int    `json:"stock_quantity"`
	StockStatus   string `json:"stock_status"`
	ManageStock   bool   `json:"manage_stock"`

	// Media
	Image string `json:"image"`

	// Attributes
	Attributes map[string]string `json:"attributes"`

	// Settings
	IsActive  bool `json:"is_active"`
	SortOrder int  `json:"sort_order"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProductAttributeResponse represents the response for product attribute data
type ProductAttributeResponse struct {
	ID        uint   `json:"id"`
	ProductID uint   `json:"product_id"`
	Name      string `json:"name"`
	Value     string `json:"value"`
	Slug      string `json:"slug"`

	// Settings
	IsVisible   bool `json:"is_visible"`
	IsVariation bool `json:"is_variation"`
	SortOrder   int  `json:"sort_order"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts Product to ProductResponse
func (p *Product) ToResponse() ProductResponse {
	response := ProductResponse{
		ID:                p.ID,
		Name:              p.Name,
		Slug:              p.Slug,
		Description:       p.Description,
		ShortDescription:  p.ShortDescription,
		SKU:               p.SKU,
		Type:              p.Type,
		Status:            p.Status,
		RegularPrice:      p.RegularPrice,
		SalePrice:         p.SalePrice,
		CostPrice:         p.CostPrice,
		ManageStock:       p.ManageStock,
		StockQuantity:     p.StockQuantity,
		LowStockThreshold: p.LowStockThreshold,
		StockStatus:       p.StockStatus,
		Weight:            p.Weight,
		Length:            p.Length,
		Width:             p.Width,
		Height:            p.Height,
		FeaturedImage:     p.FeaturedImage,
		MetaTitle:         p.MetaTitle,
		MetaDescription:   p.MetaDescription,
		MetaKeywords:      p.MetaKeywords,
		BrandID:           p.BrandID,
		CategoryID:        p.CategoryID,
		IsFeatured:        p.IsFeatured,
		IsDigital:         p.IsDigital,
		RequiresShipping:  p.RequiresShipping,
		IsDownloadable:    p.IsDownloadable,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}

	// Parse images JSON
	if p.Images != "" {
		// This will be handled in the service layer
		response.Images = []string{}
	}

	// Add brand if exists
	if p.Brand != nil {
		brandResponse := p.Brand.ToResponse()
		response.Brand = &brandResponse
	}

	// Add category if exists
	if p.Category != nil {
		categoryResponse := p.Category.ToResponse()
		response.Category = &categoryResponse
	}

	// Add variants if exists
	if len(p.Variants) > 0 {
		variants := make([]ProductVariantResponse, len(p.Variants))
		for i, variant := range p.Variants {
			variants[i] = variant.ToResponse()
		}
		response.Variants = variants
	}

	// Add attributes if exists
	if len(p.Attributes) > 0 {
		attributes := make([]ProductAttributeResponse, len(p.Attributes))
		for i, attr := range p.Attributes {
			attributes[i] = attr.ToResponse()
		}
		response.Attributes = attributes
	}

	return response
}

// ToResponse converts ProductVariant to ProductVariantResponse
func (pv *ProductVariant) ToResponse() ProductVariantResponse {
	response := ProductVariantResponse{
		ID:            pv.ID,
		ProductID:     pv.ProductID,
		Name:          pv.Name,
		SKU:           pv.SKU,
		Description:   pv.Description,
		RegularPrice:  pv.RegularPrice,
		SalePrice:     pv.SalePrice,
		CostPrice:     pv.CostPrice,
		StockQuantity: pv.StockQuantity,
		StockStatus:   pv.StockStatus,
		ManageStock:   pv.ManageStock,
		Image:         pv.Image,
		IsActive:      pv.IsActive,
		SortOrder:     pv.SortOrder,
		CreatedAt:     pv.CreatedAt,
		UpdatedAt:     pv.UpdatedAt,
	}

	// Parse attributes JSON
	if pv.Attributes != "" {
		// This will be handled in the service layer
		response.Attributes = map[string]string{}
	}

	return response
}

// ToResponse converts ProductAttribute to ProductAttributeResponse
func (pa *ProductAttribute) ToResponse() ProductAttributeResponse {
	return ProductAttributeResponse{
		ID:          pa.ID,
		ProductID:   pa.ProductID,
		Name:        pa.Name,
		Value:       pa.Value,
		Slug:        pa.Slug,
		IsVisible:   pa.IsVisible,
		IsVariation: pa.IsVariation,
		SortOrder:   pa.SortOrder,
		CreatedAt:   pa.CreatedAt,
		UpdatedAt:   pa.UpdatedAt,
	}
}

// TableName returns the table name for Product
func (Product) TableName() string {
	return "products"
}

// TableName returns the table name for ProductVariant
func (ProductVariant) TableName() string {
	return "product_variants"
}

// TableName returns the table name for ProductAttribute
func (ProductAttribute) TableName() string {
	return "product_attributes"
}
