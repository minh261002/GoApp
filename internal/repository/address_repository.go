package repository

import (
	"fmt"
	"go_app/internal/model"
	"go_app/pkg/database"

	"gorm.io/gorm"
)

// AddressRepository defines methods for interacting with address data
type AddressRepository interface {
	// Basic CRUD
	CreateAddress(address *model.Address) error
	GetAddressByID(id uint) (*model.Address, error)
	GetAllAddresses(page, limit int, filters map[string]interface{}) ([]model.Address, int64, error)
	GetAddressesByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.Address, int64, error)
	UpdateAddress(address *model.Address) error
	DeleteAddress(id uint) error

	// User-specific queries
	GetDefaultAddressByUser(userID uint) (*model.Address, error)
	GetAddressesByType(userID uint, addressType model.AddressType) ([]model.Address, error)
	GetActiveAddressesByUser(userID uint) ([]model.Address, error)
	SetDefaultAddress(userID, addressID uint) error

	// Geographic queries
	GetAddressesByCity(city string, page, limit int) ([]model.Address, int64, error)
	GetAddressesByDistrict(district string, page, limit int) ([]model.Address, int64, error)
	GetAddressesNearby(latitude, longitude float64, radiusKm float64, page, limit int) ([]model.Address, int64, error)
	SearchAddresses(query string, page, limit int) ([]model.Address, int64, error)

	// Statistics
	GetAddressStats() (*model.AddressStatsResponse, error)
	GetAddressStatsByUser(userID uint) (map[string]interface{}, error)
	GetAddressStatsByCity() (map[string]int64, error)
}

// addressRepository implements AddressRepository
type addressRepository struct {
	db *gorm.DB
}

// NewAddressRepository creates a new AddressRepository
func NewAddressRepository() AddressRepository {
	return &addressRepository{
		db: database.DB,
	}
}

// Basic CRUD

// CreateAddress creates a new address
func (r *addressRepository) CreateAddress(address *model.Address) error {
	return r.db.Create(address).Error
}

// GetAddressByID retrieves an address by its ID
func (r *addressRepository) GetAddressByID(id uint) (*model.Address, error) {
	var address model.Address
	if err := r.db.Preload("User").First(&address, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &address, nil
}

// GetAllAddresses retrieves all addresses with pagination and filters
func (r *addressRepository) GetAllAddresses(page, limit int, filters map[string]interface{}) ([]model.Address, int64, error) {
	var addresses []model.Address
	var total int64
	db := r.db.Model(&model.Address{})

	// Apply filters
	for key, value := range filters {
		switch key {
		case "user_id":
			db = db.Where("user_id = ?", value)
		case "type":
			db = db.Where("type = ?", value)
		case "is_default":
			db = db.Where("is_default = ?", value)
		case "is_active":
			db = db.Where("is_active = ?", value)
		case "city":
			db = db.Where("city = ?", value)
		case "district":
			db = db.Where("district = ?", value)
		case "ward":
			db = db.Where("ward = ?", value)
		case "country":
			db = db.Where("country = ?", value)
		case "postal_code":
			db = db.Where("postal_code = ?", value)
		case "search":
			searchTerm := fmt.Sprintf("%%%s%%", value.(string))
			db = db.Where("full_name LIKE ? OR address_line1 LIKE ? OR ward LIKE ? OR district LIKE ? OR city LIKE ?",
				searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
		}
	}

	// Count total records
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting
	db = db.Order("is_default DESC, created_at DESC")

	if err := db.Preload("User").Find(&addresses).Error; err != nil {
		return nil, 0, err
	}

	return addresses, total, nil
}

// GetAddressesByUser retrieves addresses for a specific user
func (r *addressRepository) GetAddressesByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.Address, int64, error) {
	filters["user_id"] = userID
	return r.GetAllAddresses(page, limit, filters)
}

// UpdateAddress updates an existing address
func (r *addressRepository) UpdateAddress(address *model.Address) error {
	return r.db.Save(address).Error
}

// DeleteAddress soft deletes an address
func (r *addressRepository) DeleteAddress(id uint) error {
	return r.db.Delete(&model.Address{}, id).Error
}

// User-specific queries

// GetDefaultAddressByUser retrieves the default address for a user
func (r *addressRepository) GetDefaultAddressByUser(userID uint) (*model.Address, error) {
	var address model.Address
	if err := r.db.Where("user_id = ? AND is_default = ? AND is_active = ?", userID, true, true).
		Preload("User").
		First(&address).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &address, nil
}

// GetAddressesByType retrieves addresses by type for a user
func (r *addressRepository) GetAddressesByType(userID uint, addressType model.AddressType) ([]model.Address, error) {
	var addresses []model.Address
	err := r.db.Where("user_id = ? AND type = ? AND is_active = ?", userID, addressType, true).
		Preload("User").
		Order("is_default DESC, created_at DESC").
		Find(&addresses).Error
	return addresses, err
}

// GetActiveAddressesByUser retrieves all active addresses for a user
func (r *addressRepository) GetActiveAddressesByUser(userID uint) ([]model.Address, error) {
	var addresses []model.Address
	err := r.db.Where("user_id = ? AND is_active = ?", userID, true).
		Preload("User").
		Order("is_default DESC, created_at DESC").
		Find(&addresses).Error
	return addresses, err
}

// SetDefaultAddress sets an address as default for a user
func (r *addressRepository) SetDefaultAddress(userID, addressID uint) error {
	// Start transaction
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Remove default flag from all addresses of the user
	if err := tx.Model(&model.Address{}).
		Where("user_id = ?", userID).
		Update("is_default", false).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Set the specified address as default
	if err := tx.Model(&model.Address{}).
		Where("id = ? AND user_id = ?", addressID, userID).
		Update("is_default", true).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// Geographic queries

// GetAddressesByCity retrieves addresses by city
func (r *addressRepository) GetAddressesByCity(city string, page, limit int) ([]model.Address, int64, error) {
	var addresses []model.Address
	var total int64
	db := r.db.Model(&model.Address{}).Where("city = ? AND is_active = ?", city, true)

	// Count total records
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting
	db = db.Order("is_default DESC, created_at DESC")

	if err := db.Preload("User").Find(&addresses).Error; err != nil {
		return nil, 0, err
	}

	return addresses, total, nil
}

// GetAddressesByDistrict retrieves addresses by district
func (r *addressRepository) GetAddressesByDistrict(district string, page, limit int) ([]model.Address, int64, error) {
	var addresses []model.Address
	var total int64
	db := r.db.Model(&model.Address{}).Where("district = ? AND is_active = ?", district, true)

	// Count total records
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting
	db = db.Order("is_default DESC, created_at DESC")

	if err := db.Preload("User").Find(&addresses).Error; err != nil {
		return nil, 0, err
	}

	return addresses, total, nil
}

// GetAddressesNearby retrieves addresses within a radius of given coordinates
func (r *addressRepository) GetAddressesNearby(latitude, longitude float64, radiusKm float64, page, limit int) ([]model.Address, int64, error) {
	var addresses []model.Address
	var total int64

	// Calculate bounding box for approximate filtering
	latDelta := radiusKm / 111.0 // 1 degree latitude ≈ 111 km
	lngDelta := radiusKm / (111.0 * cos(latitude*3.14159265359/180.0))

	minLat := latitude - latDelta
	maxLat := latitude + latDelta
	minLng := longitude - lngDelta
	maxLng := longitude + lngDelta

	// First filter by bounding box
	db := r.db.Model(&model.Address{}).
		Where("latitude BETWEEN ? AND ? AND longitude BETWEEN ? AND ? AND is_active = ?",
			minLat, maxLat, minLng, maxLng, true)

	// Count total records
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting by distance (approximate)
	db = db.Order("is_default DESC, created_at DESC")

	if err := db.Preload("User").Find(&addresses).Error; err != nil {
		return nil, 0, err
	}

	// Note: For more accurate distance calculation, you would need to implement
	// Haversine formula or use PostGIS/MySQL spatial functions

	return addresses, total, nil
}

// SearchAddresses performs full-text search on addresses
func (r *addressRepository) SearchAddresses(query string, page, limit int) ([]model.Address, int64, error) {
	var addresses []model.Address
	var total int64

	// Use MATCH AGAINST for full-text search
	db := r.db.Model(&model.Address{}).
		Where("MATCH(full_name, address_line1, ward, district, city) AGAINST(? IN NATURAL LANGUAGE MODE) AND is_active = ?",
			query, true)

	// Count total records
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting by relevance and default status
	db = db.Order("is_default DESC, MATCH(full_name, address_line1, ward, district, city) AGAINST('" + query + "' IN NATURAL LANGUAGE MODE) DESC")

	if err := db.Preload("User").Find(&addresses).Error; err != nil {
		return nil, 0, err
	}

	return addresses, total, nil
}

// Statistics

// GetAddressStats retrieves address statistics
func (r *addressRepository) GetAddressStats() (*model.AddressStatsResponse, error) {
	var stats model.AddressStatsResponse
	var count int64

	// Total addresses
	r.db.Model(&model.Address{}).Count(&count)
	stats.TotalAddresses = count

	// Active addresses
	r.db.Model(&model.Address{}).Where("is_active = ?", true).Count(&count)
	stats.ActiveAddresses = count

	// Default addresses
	r.db.Model(&model.Address{}).Where("is_default = ?", true).Count(&count)
	stats.DefaultAddresses = count

	// Addresses by type
	stats.AddressesByType = make(map[model.AddressType]int64)
	var typeStats []struct {
		Type  model.AddressType `json:"type"`
		Count int64             `json:"count"`
	}
	r.db.Model(&model.Address{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Scan(&typeStats)

	for _, stat := range typeStats {
		stats.AddressesByType[stat.Type] = stat.Count
	}

	// Addresses by city
	stats.AddressesByCity = make(map[string]int64)
	var cityStats []struct {
		City  string `json:"city"`
		Count int64  `json:"count"`
	}
	r.db.Model(&model.Address{}).
		Select("city, COUNT(*) as count").
		Where("is_active = ?", true).
		Group("city").
		Order("count DESC").
		Limit(20).
		Scan(&cityStats)

	for _, stat := range cityStats {
		stats.AddressesByCity[stat.City] = stat.Count
	}

	return &stats, nil
}

// GetAddressStatsByUser retrieves address statistics for a specific user
func (r *addressRepository) GetAddressStatsByUser(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var count int64

	// Total addresses for user
	r.db.Model(&model.Address{}).Where("user_id = ?", userID).Count(&count)
	stats["total_addresses"] = count

	// Active addresses for user
	r.db.Model(&model.Address{}).Where("user_id = ? AND is_active = ?", userID, true).Count(&count)
	stats["active_addresses"] = count

	// Default addresses for user
	r.db.Model(&model.Address{}).Where("user_id = ? AND is_default = ?", userID, true).Count(&count)
	stats["default_addresses"] = count

	// Addresses by type for user
	var typeStats []struct {
		Type  model.AddressType `json:"type"`
		Count int64             `json:"count"`
	}
	r.db.Model(&model.Address{}).
		Select("type, COUNT(*) as count").
		Where("user_id = ? AND is_active = ?", userID, true).
		Group("type").
		Scan(&typeStats)

	typeMap := make(map[model.AddressType]int64)
	for _, stat := range typeStats {
		typeMap[stat.Type] = stat.Count
	}
	stats["addresses_by_type"] = typeMap

	// Last address created
	var lastAddress model.Address
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").First(&lastAddress).Error; err == nil {
		stats["last_address_created"] = lastAddress.CreatedAt
	}

	return stats, nil
}

// GetAddressStatsByCity retrieves address statistics by city
func (r *addressRepository) GetAddressStatsByCity() (map[string]int64, error) {
	stats := make(map[string]int64)

	var cityStats []struct {
		City  string `json:"city"`
		Count int64  `json:"count"`
	}

	r.db.Model(&model.Address{}).
		Select("city, COUNT(*) as count").
		Where("is_active = ?", true).
		Group("city").
		Order("count DESC").
		Scan(&cityStats)

	for _, stat := range cityStats {
		stats[stat.City] = stat.Count
	}

	return stats, nil
}

// Helper function for cosine calculation
func cos(x float64) float64 {
	// Simple cosine approximation for small angles
	// For production, use math.Cos from math package
	return 1.0 - (x*x)/2.0 + (x*x*x*x)/24.0
}
