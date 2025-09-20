package service

import (
	"errors"
	"fmt"
	"go_app/internal/model"
	"go_app/internal/repository"
	"go_app/pkg/logger"
)

// AddressService defines methods for address business logic
type AddressService interface {
	// Basic CRUD
	CreateAddress(req *model.AddressCreateRequest, userID uint) (*model.AddressResponse, error)
	GetAddressByID(id uint, userID uint) (*model.AddressResponse, error)
	GetAllAddresses(page, limit int, filters map[string]interface{}) ([]model.AddressResponse, int64, error)
	GetAddressesByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.AddressResponse, int64, error)
	UpdateAddress(id uint, req *model.AddressUpdateRequest, userID uint) (*model.AddressResponse, error)
	DeleteAddress(id uint, userID uint) error

	// User-specific operations
	GetDefaultAddressByUser(userID uint) (*model.AddressResponse, error)
	GetAddressesByType(userID uint, addressType model.AddressType) ([]model.AddressResponse, error)
	GetActiveAddressesByUser(userID uint) ([]model.AddressResponse, error)
	SetDefaultAddress(userID, addressID uint) (*model.AddressResponse, error)

	// Geographic operations
	GetAddressesByCity(city string, page, limit int) ([]model.AddressResponse, int64, error)
	GetAddressesByDistrict(district string, page, limit int) ([]model.AddressResponse, int64, error)
	GetAddressesNearby(latitude, longitude float64, radiusKm float64, page, limit int) ([]model.AddressResponse, int64, error)
	SearchAddresses(query string, page, limit int) ([]model.AddressResponse, int64, error)

	// Statistics
	GetAddressStats() (*model.AddressStatsResponse, error)
	GetAddressStatsByUser(userID uint) (map[string]interface{}, error)
	GetAddressStatsByCity() (map[string]int64, error)

	// Utility
	ValidateAddress(address *model.Address) error
	FormatAddress(address *model.Address) string
	CalculateDistance(lat1, lon1, lat2, lon2 float64) float64
}

// addressService implements AddressService
type addressService struct {
	addressRepo repository.AddressRepository
	userRepo    repository.UserRepository
}

// NewAddressService creates a new AddressService
func NewAddressService() AddressService {
	return &addressService{
		addressRepo: repository.NewAddressRepository(),
		userRepo:    repository.NewUserRepository(),
	}
}

// Basic CRUD

// CreateAddress creates a new address
func (s *addressService) CreateAddress(req *model.AddressCreateRequest, userID uint) (*model.AddressResponse, error) {
	// Get user information
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		logger.Errorf("Error getting user by ID %d: %v", userID, err)
		return nil, fmt.Errorf("failed to retrieve user")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Create address
	address := &model.Address{
		UserID:       userID,
		Type:         req.Type,
		IsDefault:    req.IsDefault,
		IsActive:     req.IsActive,
		FullName:     req.FullName,
		Phone:        req.Phone,
		Email:        req.Email,
		AddressLine1: req.AddressLine1,
		AddressLine2: req.AddressLine2,
		Ward:         req.Ward,
		District:     req.District,
		City:         req.City,
		State:        req.State,
		Country:      req.Country,
		PostalCode:   req.PostalCode,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		Landmark:     req.Landmark,
		Instructions: req.Instructions,
		Notes:        req.Notes,
	}

	// Validate address
	if err := s.ValidateAddress(address); err != nil {
		logger.Errorf("Address validation failed: %v", err)
		return nil, err
	}

	// If this is set as default, remove default flag from other addresses
	if address.IsDefault {
		if err := s.setDefaultAddress(userID, 0); err != nil {
			logger.Warnf("Failed to remove default flag from other addresses: %v", err)
		}
	}

	// Create address in database
	if err := s.addressRepo.CreateAddress(address); err != nil {
		logger.Errorf("Error creating address: %v", err)
		return nil, fmt.Errorf("failed to create address")
	}

	// Get created address with relations
	createdAddress, err := s.addressRepo.GetAddressByID(address.ID)
	if err != nil {
		logger.Errorf("Error getting created address: %v", err)
		return nil, fmt.Errorf("failed to retrieve created address")
	}

	return s.toAddressResponse(createdAddress), nil
}

// GetAddressByID retrieves an address by its ID
func (s *addressService) GetAddressByID(id uint, userID uint) (*model.AddressResponse, error) {
	address, err := s.addressRepo.GetAddressByID(id)
	if err != nil {
		logger.Errorf("Error getting address by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve address")
	}
	if address == nil {
		return nil, errors.New("address not found")
	}

	// Check if user owns this address (unless admin)
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		logger.Errorf("Error getting user by ID %d: %v", userID, err)
		return nil, fmt.Errorf("failed to retrieve user")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Allow access if user owns the address or is admin
	if address.UserID != userID && user.Role != "admin" && user.Role != "super_admin" {
		return nil, errors.New("access denied: you can only view your own addresses")
	}

	return s.toAddressResponse(address), nil
}

// GetAllAddresses retrieves all addresses with pagination and filters
func (s *addressService) GetAllAddresses(page, limit int, filters map[string]interface{}) ([]model.AddressResponse, int64, error) {
	addresses, total, err := s.addressRepo.GetAllAddresses(page, limit, filters)
	if err != nil {
		logger.Errorf("Error getting addresses: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve addresses")
	}

	var responses []model.AddressResponse
	for _, address := range addresses {
		responses = append(responses, *s.toAddressResponse(&address))
	}
	return responses, total, nil
}

// GetAddressesByUser retrieves addresses for a specific user
func (s *addressService) GetAddressesByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.AddressResponse, int64, error) {
	addresses, total, err := s.addressRepo.GetAddressesByUser(userID, page, limit, filters)
	if err != nil {
		logger.Errorf("Error getting user addresses: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve user addresses")
	}

	var responses []model.AddressResponse
	for _, address := range addresses {
		responses = append(responses, *s.toAddressResponse(&address))
	}
	return responses, total, nil
}

// UpdateAddress updates an existing address
func (s *addressService) UpdateAddress(id uint, req *model.AddressUpdateRequest, userID uint) (*model.AddressResponse, error) {
	address, err := s.addressRepo.GetAddressByID(id)
	if err != nil {
		logger.Errorf("Error getting address by ID %d for update: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve address")
	}
	if address == nil {
		return nil, errors.New("address not found")
	}

	// Check if user owns this address (unless admin)
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		logger.Errorf("Error getting user by ID %d: %v", userID, err)
		return nil, fmt.Errorf("failed to retrieve user")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	if address.UserID != userID && user.Role != "admin" && user.Role != "super_admin" {
		return nil, errors.New("access denied: you can only update your own addresses")
	}

	// Update fields
	if req.Type != nil {
		address.Type = *req.Type
	}
	if req.IsDefault != nil {
		address.IsDefault = *req.IsDefault
	}
	if req.IsActive != nil {
		address.IsActive = *req.IsActive
	}
	if req.FullName != "" {
		address.FullName = req.FullName
	}
	if req.Phone != "" {
		address.Phone = req.Phone
	}
	if req.Email != "" {
		address.Email = req.Email
	}
	if req.AddressLine1 != "" {
		address.AddressLine1 = req.AddressLine1
	}
	if req.AddressLine2 != "" {
		address.AddressLine2 = req.AddressLine2
	}
	if req.Ward != "" {
		address.Ward = req.Ward
	}
	if req.District != "" {
		address.District = req.District
	}
	if req.City != "" {
		address.City = req.City
	}
	if req.State != "" {
		address.State = req.State
	}
	if req.Country != "" {
		address.Country = req.Country
	}
	if req.PostalCode != "" {
		address.PostalCode = req.PostalCode
	}
	if req.Latitude != nil {
		address.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		address.Longitude = req.Longitude
	}
	if req.Landmark != "" {
		address.Landmark = req.Landmark
	}
	if req.Instructions != "" {
		address.Instructions = req.Instructions
	}
	if req.Notes != "" {
		address.Notes = req.Notes
	}

	// Validate updated address
	if err := s.ValidateAddress(address); err != nil {
		logger.Errorf("Address validation failed: %v", err)
		return nil, err
	}

	// If this is set as default, remove default flag from other addresses
	if address.IsDefault {
		if err := s.setDefaultAddress(userID, id); err != nil {
			logger.Warnf("Failed to remove default flag from other addresses: %v", err)
		}
	}

	if err := s.addressRepo.UpdateAddress(address); err != nil {
		logger.Errorf("Error updating address %d: %v", id, err)
		return nil, fmt.Errorf("failed to update address")
	}

	// Get updated address with relations
	updatedAddress, err := s.addressRepo.GetAddressByID(address.ID)
	if err != nil {
		logger.Errorf("Error getting updated address: %v", err)
		return nil, fmt.Errorf("failed to retrieve updated address")
	}

	return s.toAddressResponse(updatedAddress), nil
}

// DeleteAddress deletes an address
func (s *addressService) DeleteAddress(id uint, userID uint) error {
	address, err := s.addressRepo.GetAddressByID(id)
	if err != nil {
		logger.Errorf("Error getting address by ID %d for deletion: %v", id, err)
		return fmt.Errorf("failed to retrieve address")
	}
	if address == nil {
		return errors.New("address not found")
	}

	// Check if user owns this address (unless admin)
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		logger.Errorf("Error getting user by ID %d: %v", userID, err)
		return fmt.Errorf("failed to retrieve user")
	}
	if user == nil {
		return errors.New("user not found")
	}

	if address.UserID != userID && user.Role != "admin" && user.Role != "super_admin" {
		return errors.New("access denied: you can only delete your own addresses")
	}

	// Check if this is the default address
	if address.IsDefault {
		return errors.New("cannot delete default address. Please set another address as default first")
	}

	if err := s.addressRepo.DeleteAddress(id); err != nil {
		logger.Errorf("Error deleting address %d: %v", id, err)
		return fmt.Errorf("failed to delete address")
	}

	return nil
}

// User-specific operations

// GetDefaultAddressByUser retrieves the default address for a user
func (s *addressService) GetDefaultAddressByUser(userID uint) (*model.AddressResponse, error) {
	address, err := s.addressRepo.GetDefaultAddressByUser(userID)
	if err != nil {
		logger.Errorf("Error getting default address for user %d: %v", userID, err)
		return nil, fmt.Errorf("failed to retrieve default address")
	}
	if address == nil {
		return nil, errors.New("no default address found")
	}
	return s.toAddressResponse(address), nil
}

// GetAddressesByType retrieves addresses by type for a user
func (s *addressService) GetAddressesByType(userID uint, addressType model.AddressType) ([]model.AddressResponse, error) {
	addresses, err := s.addressRepo.GetAddressesByType(userID, addressType)
	if err != nil {
		logger.Errorf("Error getting addresses by type for user %d: %v", userID, err)
		return nil, fmt.Errorf("failed to retrieve addresses by type")
	}

	var responses []model.AddressResponse
	for _, address := range addresses {
		responses = append(responses, *s.toAddressResponse(&address))
	}
	return responses, nil
}

// GetActiveAddressesByUser retrieves all active addresses for a user
func (s *addressService) GetActiveAddressesByUser(userID uint) ([]model.AddressResponse, error) {
	addresses, err := s.addressRepo.GetActiveAddressesByUser(userID)
	if err != nil {
		logger.Errorf("Error getting active addresses for user %d: %v", userID, err)
		return nil, fmt.Errorf("failed to retrieve active addresses")
	}

	var responses []model.AddressResponse
	for _, address := range addresses {
		responses = append(responses, *s.toAddressResponse(&address))
	}
	return responses, nil
}

// SetDefaultAddress sets an address as default for a user
func (s *addressService) SetDefaultAddress(userID, addressID uint) (*model.AddressResponse, error) {
	// Check if address exists and belongs to user
	address, err := s.addressRepo.GetAddressByID(addressID)
	if err != nil {
		logger.Errorf("Error getting address by ID %d: %v", addressID, err)
		return nil, fmt.Errorf("failed to retrieve address")
	}
	if address == nil {
		return nil, errors.New("address not found")
	}

	if address.UserID != userID {
		return nil, errors.New("access denied: address does not belong to user")
	}

	// Set as default
	if err := s.addressRepo.SetDefaultAddress(userID, addressID); err != nil {
		logger.Errorf("Error setting default address %d for user %d: %v", addressID, userID, err)
		return nil, fmt.Errorf("failed to set default address")
	}

	// Get updated address
	updatedAddress, err := s.addressRepo.GetAddressByID(addressID)
	if err != nil {
		logger.Errorf("Error getting updated address: %v", err)
		return nil, fmt.Errorf("failed to retrieve updated address")
	}

	return s.toAddressResponse(updatedAddress), nil
}

// Geographic operations

// GetAddressesByCity retrieves addresses by city
func (s *addressService) GetAddressesByCity(city string, page, limit int) ([]model.AddressResponse, int64, error) {
	addresses, total, err := s.addressRepo.GetAddressesByCity(city, page, limit)
	if err != nil {
		logger.Errorf("Error getting addresses by city %s: %v", city, err)
		return nil, 0, fmt.Errorf("failed to retrieve addresses by city")
	}

	var responses []model.AddressResponse
	for _, address := range addresses {
		responses = append(responses, *s.toAddressResponse(&address))
	}
	return responses, total, nil
}

// GetAddressesByDistrict retrieves addresses by district
func (s *addressService) GetAddressesByDistrict(district string, page, limit int) ([]model.AddressResponse, int64, error) {
	addresses, total, err := s.addressRepo.GetAddressesByDistrict(district, page, limit)
	if err != nil {
		logger.Errorf("Error getting addresses by district %s: %v", district, err)
		return nil, 0, fmt.Errorf("failed to retrieve addresses by district")
	}

	var responses []model.AddressResponse
	for _, address := range addresses {
		responses = append(responses, *s.toAddressResponse(&address))
	}
	return responses, total, nil
}

// GetAddressesNearby retrieves addresses within a radius
func (s *addressService) GetAddressesNearby(latitude, longitude float64, radiusKm float64, page, limit int) ([]model.AddressResponse, int64, error) {
	addresses, total, err := s.addressRepo.GetAddressesNearby(latitude, longitude, radiusKm, page, limit)
	if err != nil {
		logger.Errorf("Error getting nearby addresses: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve nearby addresses")
	}

	var responses []model.AddressResponse
	for _, address := range addresses {
		responses = append(responses, *s.toAddressResponse(&address))
	}
	return responses, total, nil
}

// SearchAddresses performs full-text search on addresses
func (s *addressService) SearchAddresses(query string, page, limit int) ([]model.AddressResponse, int64, error) {
	addresses, total, err := s.addressRepo.SearchAddresses(query, page, limit)
	if err != nil {
		logger.Errorf("Error searching addresses: %v", err)
		return nil, 0, fmt.Errorf("failed to search addresses")
	}

	var responses []model.AddressResponse
	for _, address := range addresses {
		responses = append(responses, *s.toAddressResponse(&address))
	}
	return responses, total, nil
}

// Statistics

// GetAddressStats retrieves address statistics
func (s *addressService) GetAddressStats() (*model.AddressStatsResponse, error) {
	stats, err := s.addressRepo.GetAddressStats()
	if err != nil {
		logger.Errorf("Error getting address statistics: %v", err)
		return nil, fmt.Errorf("failed to retrieve address statistics")
	}
	return stats, nil
}

// GetAddressStatsByUser retrieves address statistics for a specific user
func (s *addressService) GetAddressStatsByUser(userID uint) (map[string]interface{}, error) {
	stats, err := s.addressRepo.GetAddressStatsByUser(userID)
	if err != nil {
		logger.Errorf("Error getting user address statistics: %v", err)
		return nil, fmt.Errorf("failed to retrieve user address statistics")
	}
	return stats, nil
}

// GetAddressStatsByCity retrieves address statistics by city
func (s *addressService) GetAddressStatsByCity() (map[string]int64, error) {
	stats, err := s.addressRepo.GetAddressStatsByCity()
	if err != nil {
		logger.Errorf("Error getting city address statistics: %v", err)
		return nil, fmt.Errorf("failed to retrieve city address statistics")
	}
	return stats, nil
}

// Utility methods

// ValidateAddress validates address data
func (s *addressService) ValidateAddress(address *model.Address) error {
	return address.ValidateAddress()
}

// FormatAddress formats address as string
func (s *addressService) FormatAddress(address *model.Address) string {
	return address.GetFullAddress()
}

// CalculateDistance calculates distance between two coordinates using Haversine formula
func (s *addressService) CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371 // Earth's radius in kilometers

	// Convert degrees to radians
	lat1Rad := lat1 * 3.14159265359 / 180
	lon1Rad := lon1 * 3.14159265359 / 180
	lat2Rad := lat2 * 3.14159265359 / 180
	lon2Rad := lon2 * 3.14159265359 / 180

	// Haversine formula
	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	a := (1-cos(dLat))/2 + cos(lat1Rad)*cos(lat2Rad)*(1-cos(dLon))/2
	c := 2 * asin(sqrt(a))

	return earthRadius * c
}

// Helper methods

// setDefaultAddress sets an address as default and removes default flag from others
func (s *addressService) setDefaultAddress(userID, addressID uint) error {
	return s.addressRepo.SetDefaultAddress(userID, addressID)
}

// toAddressResponse converts Address to AddressResponse
func (s *addressService) toAddressResponse(address *model.Address) *model.AddressResponse {
	return address.ToResponse()
}

// Helper functions for math calculations
func cos(x float64) float64 {
	// Simple cosine approximation
	return 1.0 - (x*x)/2.0 + (x*x*x*x)/24.0
}

func asin(x float64) float64 {
	// Simple arcsine approximation
	return x + (x*x*x)/6.0 + (3*x*x*x*x*x)/40.0
}

func sqrt(x float64) float64 {
	// Simple square root approximation
	if x == 0 {
		return 0
	}
	if x < 0 {
		return 0
	}

	// Newton's method
	guess := x / 2
	for i := 0; i < 10; i++ {
		guess = (guess + x/guess) / 2
	}
	return guess
}
