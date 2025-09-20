package model

import (
	"time"

	"gorm.io/gorm"
)

// Category represents a category in the hierarchical structure
type Category struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"size:100;not null" validate:"required,min=2,max=100"`
	Slug        string         `json:"slug" gorm:"size:120;not null;uniqueIndex" validate:"required,min=2,max=120"`
	Description string         `json:"description" gorm:"type:text"`
	Image       string         `json:"image" gorm:"size:255"`
	Icon        string         `json:"icon" gorm:"size:100"`
	ParentID    *uint          `json:"parent_id" gorm:"index"`
	Parent      *Category      `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children    []Category     `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Level       int            `json:"level" gorm:"default:0"`
	Path        string         `json:"path" gorm:"size:500;index"` // e.g., "1/2/3" for hierarchical path
	SortOrder   int            `json:"sort_order" gorm:"default:0"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	IsLeaf      bool           `json:"is_leaf" gorm:"default:true"` // true if no children
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// CategoryCreateRequest represents the request to create a category
type CategoryCreateRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Icon        string `json:"icon"`
	ParentID    *uint  `json:"parent_id"`
	SortOrder   int    `json:"sort_order"`
	IsActive    *bool  `json:"is_active"`
}

// CategoryUpdateRequest represents the request to update a category
type CategoryUpdateRequest struct {
	Name        string `json:"name" validate:"omitempty,min=2,max=100"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Icon        string `json:"icon"`
	ParentID    *uint  `json:"parent_id"`
	SortOrder   int    `json:"sort_order"`
	IsActive    *bool  `json:"is_active"`
}

// CategoryResponse represents the response for category data
type CategoryResponse struct {
	ID          uint               `json:"id"`
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	Description string             `json:"description"`
	Image       string             `json:"image"`
	Icon        string             `json:"icon"`
	ParentID    *uint              `json:"parent_id"`
	Parent      *CategoryResponse  `json:"parent,omitempty"`
	Children    []CategoryResponse `json:"children,omitempty"`
	Level       int                `json:"level"`
	Path        string             `json:"path"`
	SortOrder   int                `json:"sort_order"`
	IsActive    bool               `json:"is_active"`
	IsLeaf      bool               `json:"is_leaf"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// CategoryTreeResponse represents a tree structure response
type CategoryTreeResponse struct {
	ID          uint                   `json:"id"`
	Name        string                 `json:"name"`
	Slug        string                 `json:"slug"`
	Description string                 `json:"description"`
	Image       string                 `json:"image"`
	Icon        string                 `json:"icon"`
	Level       int                    `json:"level"`
	Path        string                 `json:"path"`
	SortOrder   int                    `json:"sort_order"`
	IsActive    bool                   `json:"is_active"`
	IsLeaf      bool                   `json:"is_leaf"`
	Children    []CategoryTreeResponse `json:"children,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// CategoryBreadcrumb represents breadcrumb navigation
type CategoryBreadcrumb struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	Level int    `json:"level"`
}

// ToResponse converts Category to CategoryResponse
func (c *Category) ToResponse() CategoryResponse {
	response := CategoryResponse{
		ID:          c.ID,
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		Image:       c.Image,
		Icon:        c.Icon,
		ParentID:    c.ParentID,
		Level:       c.Level,
		Path:        c.Path,
		SortOrder:   c.SortOrder,
		IsActive:    c.IsActive,
		IsLeaf:      c.IsLeaf,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}

	// Add parent if exists
	if c.Parent != nil {
		parentResponse := c.Parent.ToResponse()
		response.Parent = &parentResponse
	}

	// Add children if exists
	if len(c.Children) > 0 {
		children := make([]CategoryResponse, len(c.Children))
		for i, child := range c.Children {
			children[i] = child.ToResponse()
		}
		response.Children = children
	}

	return response
}

// ToTreeResponse converts Category to CategoryTreeResponse
func (c *Category) ToTreeResponse() CategoryTreeResponse {
	response := CategoryTreeResponse{
		ID:          c.ID,
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		Image:       c.Image,
		Icon:        c.Icon,
		Level:       c.Level,
		Path:        c.Path,
		SortOrder:   c.SortOrder,
		IsActive:    c.IsActive,
		IsLeaf:      c.IsLeaf,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}

	// Add children if exists
	if len(c.Children) > 0 {
		children := make([]CategoryTreeResponse, len(c.Children))
		for i, child := range c.Children {
			children[i] = child.ToTreeResponse()
		}
		response.Children = children
	}

	return response
}

// GetBreadcrumbs generates breadcrumb navigation
func (c *Category) GetBreadcrumbs() []CategoryBreadcrumb {
	var breadcrumbs []CategoryBreadcrumb

	// Parse path to get parent IDs
	if c.Path != "" {
		// Path format: "1/2/3" where 3 is current category
		// We need to query parent categories by their IDs
		// This will be implemented in the service layer
	}

	// Add current category
	breadcrumbs = append(breadcrumbs, CategoryBreadcrumb{
		ID:    c.ID,
		Name:  c.Name,
		Slug:  c.Slug,
		Level: c.Level,
	})

	return breadcrumbs
}

// TableName returns the table name for Category
func (Category) TableName() string {
	return "categories"
}
