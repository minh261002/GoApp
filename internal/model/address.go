package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// AddressType defines the type of address
type AddressType string

const (
	AddressTypeHome     AddressType = "home"     // Nhà riêng
	AddressTypeOffice   AddressType = "office"   // Văn phòng
	AddressTypeBilling  AddressType = "billing"  // Địa chỉ thanh toán
	AddressTypeShipping AddressType = "shipping" // Địa chỉ giao hàng
	AddressTypeOther    AddressType = "other"    // Khác
)

// Address represents a user's address
type Address struct {
	ID     uint  `json:"id" gorm:"primaryKey"`
	UserID uint  `json:"user_id" gorm:"not null;index"`
	User   *User `json:"user,omitempty" gorm:"foreignKey:UserID"`

	// Address Information
	Type      AddressType `json:"type" gorm:"size:20;default:home;index"` // Loại địa chỉ
	IsDefault bool        `json:"is_default" gorm:"default:false"`        // Địa chỉ mặc định
	IsActive  bool        `json:"is_active" gorm:"default:true"`          // Trạng thái hoạt động

	// Contact Information
	FullName string `json:"full_name" gorm:"size:255;not null"` // Tên đầy đủ người nhận
	Phone    string `json:"phone" gorm:"size:20;not null"`      // Số điện thoại
	Email    string `json:"email" gorm:"size:255"`              // Email (optional)

	// Address Details
	AddressLine1 string `json:"address_line1" gorm:"size:255;not null"`  // Địa chỉ dòng 1
	AddressLine2 string `json:"address_line2" gorm:"size:255"`           // Địa chỉ dòng 2 (optional)
	Ward         string `json:"ward" gorm:"size:100;not null"`           // Phường/Xã
	District     string `json:"district" gorm:"size:100;not null"`       // Quận/Huyện
	City         string `json:"city" gorm:"size:100;not null"`           // Tỉnh/Thành phố
	State        string `json:"state" gorm:"size:100"`                   // Bang/Tỉnh (optional)
	Country      string `json:"country" gorm:"size:100;default:Vietnam"` // Quốc gia
	PostalCode   string `json:"postal_code" gorm:"size:20"`              // Mã bưu điện

	// Geographic Information
	Latitude  *float64 `json:"latitude" gorm:"type:decimal(10,8)"`  // Vĩ độ
	Longitude *float64 `json:"longitude" gorm:"type:decimal(11,8)"` // Kinh độ

	// Additional Information
	Landmark     string `json:"landmark" gorm:"size:255"`      // Địa danh gần đó
	Instructions string `json:"instructions" gorm:"type:text"` // Hướng dẫn giao hàng
	Notes        string `json:"notes" gorm:"type:text"`        // Ghi chú

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// Request/Response structs

// AddressCreateRequest represents the request body for creating an address
type AddressCreateRequest struct {
	Type      AddressType `json:"type" binding:"required,oneof=home office billing shipping other"`
	IsDefault bool        `json:"is_default"`
	IsActive  bool        `json:"is_active"`

	// Contact Information
	FullName string `json:"full_name" binding:"required,min=2,max=255"`
	Phone    string `json:"phone" binding:"required,min=10,max=20"`
	Email    string `json:"email" binding:"omitempty,email"`

	// Address Details
	AddressLine1 string `json:"address_line1" binding:"required,min=5,max=255"`
	AddressLine2 string `json:"address_line2" binding:"omitempty,max=255"`
	Ward         string `json:"ward" binding:"required,min=2,max=100"`
	District     string `json:"district" binding:"required,min=2,max=100"`
	City         string `json:"city" binding:"required,min=2,max=100"`
	State        string `json:"state" binding:"omitempty,max=100"`
	Country      string `json:"country" binding:"omitempty,max=100"`
	PostalCode   string `json:"postal_code" binding:"omitempty,max=20"`

	// Geographic Information
	Latitude  *float64 `json:"latitude" binding:"omitempty,min=-90,max=90"`
	Longitude *float64 `json:"longitude" binding:"omitempty,min=-180,max=180"`

	// Additional Information
	Landmark     string `json:"landmark" binding:"omitempty,max=255"`
	Instructions string `json:"instructions" binding:"omitempty,max=1000"`
	Notes        string `json:"notes" binding:"omitempty,max=1000"`
}

// AddressUpdateRequest represents the request body for updating an address
type AddressUpdateRequest struct {
	Type      *AddressType `json:"type" binding:"omitempty,oneof=home office billing shipping other"`
	IsDefault *bool        `json:"is_default"`
	IsActive  *bool        `json:"is_active"`

	// Contact Information
	FullName string `json:"full_name" binding:"omitempty,min=2,max=255"`
	Phone    string `json:"phone" binding:"omitempty,min=10,max=20"`
	Email    string `json:"email" binding:"omitempty,email"`

	// Address Details
	AddressLine1 string `json:"address_line1" binding:"omitempty,min=5,max=255"`
	AddressLine2 string `json:"address_line2" binding:"omitempty,max=255"`
	Ward         string `json:"ward" binding:"omitempty,min=2,max=100"`
	District     string `json:"district" binding:"omitempty,min=2,max=100"`
	City         string `json:"city" binding:"omitempty,min=2,max=100"`
	State        string `json:"state" binding:"omitempty,max=100"`
	Country      string `json:"country" binding:"omitempty,max=100"`
	PostalCode   string `json:"postal_code" binding:"omitempty,max=20"`

	// Geographic Information
	Latitude  *float64 `json:"latitude" binding:"omitempty,min=-90,max=90"`
	Longitude *float64 `json:"longitude" binding:"omitempty,min=-180,max=180"`

	// Additional Information
	Landmark     string `json:"landmark" binding:"omitempty,max=255"`
	Instructions string `json:"instructions" binding:"omitempty,max=1000"`
	Notes        string `json:"notes" binding:"omitempty,max=1000"`
}

// AddressResponse represents the response body for an address
type AddressResponse struct {
	ID        uint        `json:"id"`
	UserID    uint        `json:"user_id"`
	Type      AddressType `json:"type"`
	IsDefault bool        `json:"is_default"`
	IsActive  bool        `json:"is_active"`

	// Contact Information
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`

	// Address Details
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2"`
	Ward         string `json:"ward"`
	District     string `json:"district"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
	PostalCode   string `json:"postal_code"`

	// Geographic Information
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`

	// Additional Information
	Landmark     string `json:"landmark"`
	Instructions string `json:"instructions"`
	Notes        string `json:"notes"`

	// Computed Fields
	FullAddress  string `json:"full_address"`  // Địa chỉ đầy đủ
	ShortAddress string `json:"short_address"` // Địa chỉ ngắn gọn

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AddressStatsResponse represents address statistics
type AddressStatsResponse struct {
	TotalAddresses   int64                 `json:"total_addresses"`
	ActiveAddresses  int64                 `json:"active_addresses"`
	DefaultAddresses int64                 `json:"default_addresses"`
	AddressesByType  map[AddressType]int64 `json:"addresses_by_type"`
	AddressesByCity  map[string]int64      `json:"addresses_by_city"`
}

// Helper methods

// GetFullAddress returns the complete address string
func (a *Address) GetFullAddress() string {
	address := a.AddressLine1
	if a.AddressLine2 != "" {
		address += ", " + a.AddressLine2
	}
	address += ", " + a.Ward + ", " + a.District + ", " + a.City
	if a.State != "" {
		address += ", " + a.State
	}
	if a.Country != "" && a.Country != "Vietnam" {
		address += ", " + a.Country
	}
	if a.PostalCode != "" {
		address += " " + a.PostalCode
	}
	return address
}

// GetShortAddress returns a short version of the address
func (a *Address) GetShortAddress() string {
	return a.Ward + ", " + a.District + ", " + a.City
}

// IsValidCoordinates checks if latitude and longitude are valid
func (a *Address) IsValidCoordinates() bool {
	if a.Latitude == nil || a.Longitude == nil {
		return false
	}
	return *a.Latitude >= -90 && *a.Latitude <= 90 &&
		*a.Longitude >= -180 && *a.Longitude <= 180
}

// GetTypeDisplayName returns display name for address type
func (a *Address) GetTypeDisplayName() string {
	typeMap := map[AddressType]string{
		AddressTypeHome:     "Nhà riêng",
		AddressTypeOffice:   "Văn phòng",
		AddressTypeBilling:  "Địa chỉ thanh toán",
		AddressTypeShipping: "Địa chỉ giao hàng",
		AddressTypeOther:    "Khác",
	}
	return typeMap[a.Type]
}

// ToResponse converts Address to AddressResponse
func (a *Address) ToResponse() *AddressResponse {
	return &AddressResponse{
		ID:           a.ID,
		UserID:       a.UserID,
		Type:         a.Type,
		IsDefault:    a.IsDefault,
		IsActive:     a.IsActive,
		FullName:     a.FullName,
		Phone:        a.Phone,
		Email:        a.Email,
		AddressLine1: a.AddressLine1,
		AddressLine2: a.AddressLine2,
		Ward:         a.Ward,
		District:     a.District,
		City:         a.City,
		State:        a.State,
		Country:      a.Country,
		PostalCode:   a.PostalCode,
		Latitude:     a.Latitude,
		Longitude:    a.Longitude,
		Landmark:     a.Landmark,
		Instructions: a.Instructions,
		Notes:        a.Notes,
		FullAddress:  a.GetFullAddress(),
		ShortAddress: a.GetShortAddress(),
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}
}

// ValidateAddress validates address data
func (a *Address) ValidateAddress() error {
	if a.FullName == "" {
		return errors.New("full name is required")
	}
	if a.Phone == "" {
		return errors.New("phone is required")
	}
	if a.AddressLine1 == "" {
		return errors.New("address line 1 is required")
	}
	if a.Ward == "" {
		return errors.New("ward is required")
	}
	if a.District == "" {
		return errors.New("district is required")
	}
	if a.City == "" {
		return errors.New("city is required")
	}
	if !a.IsValidCoordinates() && a.Latitude != nil && a.Longitude != nil {
		return errors.New("invalid coordinates")
	}
	return nil
}
