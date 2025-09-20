package model

import (
	"time"

	"gorm.io/gorm"
)

// Brand represents a brand in the shop
type Brand struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"size:100;not null;uniqueIndex" validate:"required,min=2,max=100"`
	Slug        string         `json:"slug" gorm:"size:120;not null;uniqueIndex" validate:"required,min=2,max=120"`
	Description string         `json:"description" gorm:"type:text"`
	Logo        string         `json:"logo" gorm:"size:255"`
	Website     string         `json:"website" gorm:"size:255"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	SortOrder   int            `json:"sort_order" gorm:"default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// BrandCreateRequest represents the request to create a brand
type BrandCreateRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
	Website     string `json:"website" validate:"omitempty,url"`
	IsActive    *bool  `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
}

// BrandUpdateRequest represents the request to update a brand
type BrandUpdateRequest struct {
	Name        string `json:"name" validate:"omitempty,min=2,max=100"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
	Website     string `json:"website" validate:"omitempty,url"`
	IsActive    *bool  `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
}

// BrandResponse represents the response for brand data
type BrandResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Logo        string    `json:"logo"`
	Website     string    `json:"website"`
	IsActive    bool      `json:"is_active"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToResponse converts Brand to BrandResponse
func (b *Brand) ToResponse() BrandResponse {
	return BrandResponse{
		ID:          b.ID,
		Name:        b.Name,
		Slug:        b.Slug,
		Description: b.Description,
		Logo:        b.Logo,
		Website:     b.Website,
		IsActive:    b.IsActive,
		SortOrder:   b.SortOrder,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
}

// TableName returns the table name for Brand
func (Brand) TableName() string {
	return "brands"
}
