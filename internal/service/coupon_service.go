package service

import (
	"errors"
	"fmt"
	"go_app/internal/model"
	"go_app/internal/repository"
	"go_app/pkg/logger"
	"strconv"
	"strings"
	"time"
)

// CouponService defines the interface for coupon business logic
type CouponService interface {
	// Basic CRUD
	CreateCoupon(req *model.CouponCreateRequest, creatorID uint) (*model.CouponResponse, error)
	GetCouponByID(id uint) (*model.CouponResponse, error)
	GetCouponByCode(code string) (*model.CouponResponse, error)
	GetAllCoupons(page, limit int, filters map[string]interface{}) ([]model.CouponResponse, int64, error)
	UpdateCoupon(id uint, req *model.CouponUpdateRequest) (*model.CouponResponse, error)
	DeleteCoupon(id uint) error

	// Coupon Management
	GetActiveCoupons(page, limit int) ([]model.CouponResponse, int64, error)
	GetExpiredCoupons(page, limit int) ([]model.CouponResponse, int64, error)
	GetCouponsByType(couponType model.CouponType, page, limit int) ([]model.CouponResponse, int64, error)
	GetCouponsByStatus(status model.CouponStatus, page, limit int) ([]model.CouponResponse, int64, error)
	SearchCoupons(query string, page, limit int) ([]model.CouponResponse, int64, error)

	// Coupon Validation and Usage
	ValidateCoupon(req *model.CouponValidateRequest) (*model.CouponValidateResponse, error)
	UseCoupon(req *model.CouponUseRequest) (*model.CouponUsageResponse, error)
	GetCouponUsagesByCoupon(couponID uint, page, limit int) ([]model.CouponUsageResponse, int64, error)
	GetCouponUsagesByUser(userID uint, page, limit int) ([]model.CouponUsageResponse, int64, error)
	GetCouponUsagesByOrder(orderID uint) ([]model.CouponUsageResponse, error)

	// Statistics
	GetCouponStats() (*model.CouponStatsResponse, error)
	GetCouponUsageStats(couponID uint) (map[string]interface{}, error)
}

// PointService defines the interface for point business logic
type PointService interface {
	// Basic CRUD
	GetPointByUserID(userID uint) (*model.PointResponse, error)
	GetUserPointBalance(userID uint) (int, error)

	// Point Operations
	EarnPoints(req *model.PointEarnRequest) (*model.PointTransactionResponse, error)
	RedeemPoints(req *model.PointRedeemRequest) (*model.PointTransactionResponse, error)
	RefundPoints(req *model.PointRefundRequest) (*model.PointTransactionResponse, error)
	AdjustPoints(req *model.PointAdjustRequest) (*model.PointTransactionResponse, error)
	ExpirePoints(req *model.PointExpireRequest) (*model.PointTransactionResponse, error)

	// Point Queries
	GetPointTransactionByID(id uint) (*model.PointTransactionResponse, error)
	GetPointTransactionsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.PointTransactionResponse, int64, error)
	GetPointHistory(userID uint, page, limit int) ([]model.PointTransactionResponse, int64, error)
	GetExpiredPoints(userID uint) ([]model.PointTransactionResponse, error)
	GetExpiringPoints(userID uint, days int) ([]model.PointTransactionResponse, error)

	// Statistics
	GetPointStats() (*model.PointStatsResponse, error)
	GetUserPointStats(userID uint) (map[string]interface{}, error)
	GetTopEarners(limit int) ([]model.PointResponse, error)
}

type couponService struct {
	couponRepo repository.CouponRepository
	userRepo   repository.UserRepository
	orderRepo  repository.OrderRepository
}

type pointService struct {
	pointRepo repository.PointRepository
	userRepo  repository.UserRepository
	orderRepo repository.OrderRepository
}

// NewCouponService creates a new CouponService
func NewCouponService(couponRepo repository.CouponRepository, userRepo repository.UserRepository, orderRepo repository.OrderRepository) CouponService {
	return &couponService{
		couponRepo: couponRepo,
		userRepo:   userRepo,
		orderRepo:  orderRepo,
	}
}

// NewPointService creates a new PointService
func NewPointService(pointRepo repository.PointRepository, userRepo repository.UserRepository, orderRepo repository.OrderRepository) PointService {
	return &pointService{
		pointRepo: pointRepo,
		userRepo:  userRepo,
		orderRepo: orderRepo,
	}
}

// Coupon Service Implementation

func (s *couponService) CreateCoupon(req *model.CouponCreateRequest, creatorID uint) (*model.CouponResponse, error) {
	// Check if coupon code already exists
	existingCoupon, err := s.couponRepo.GetCouponByCode(req.Code)
	if err != nil {
		logger.Errorf("Error checking existing coupon by code %s: %v", req.Code, err)
		return nil, fmt.Errorf("failed to check existing coupon")
	}
	if existingCoupon != nil {
		return nil, errors.New("coupon with this code already exists")
	}

	coupon := &model.Coupon{
		Code:              req.Code,
		Name:              req.Name,
		Description:       req.Description,
		Type:              req.Type,
		DiscountValue:     req.DiscountValue,
		MinOrderAmount:    req.MinOrderAmount,
		MaxDiscountAmount: req.MaxDiscountAmount,
		UsageLimit:        req.UsageLimit,
		UsagePerUser:      req.UsagePerUser,
		ValidFrom:         req.ValidFrom,
		ValidTo:           req.ValidTo,
		TargetType:        req.TargetType,
		TargetIDs:         convertUintSliceToString(req.TargetIDs),
		IsStackable:       req.IsStackable,
		IsFirstTimeOnly:   req.IsFirstTimeOnly,
		IsNewUserOnly:     req.IsNewUserOnly,
		CreatedBy:         creatorID,
		Status:            model.CouponStatusActive,
	}

	if err := s.couponRepo.CreateCoupon(coupon); err != nil {
		logger.Errorf("Error creating coupon: %v", err)
		return nil, fmt.Errorf("failed to create coupon")
	}

	return s.toCouponResponse(coupon), nil
}

func (s *couponService) GetCouponByID(id uint) (*model.CouponResponse, error) {
	coupon, err := s.couponRepo.GetCouponByID(id)
	if err != nil {
		logger.Errorf("Error getting coupon by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve coupon")
	}
	if coupon == nil {
		return nil, errors.New("coupon not found")
	}
	return s.toCouponResponse(coupon), nil
}

func (s *couponService) GetCouponByCode(code string) (*model.CouponResponse, error) {
	coupon, err := s.couponRepo.GetCouponByCode(code)
	if err != nil {
		logger.Errorf("Error getting coupon by code %s: %v", code, err)
		return nil, fmt.Errorf("failed to retrieve coupon")
	}
	if coupon == nil {
		return nil, errors.New("coupon not found")
	}
	return s.toCouponResponse(coupon), nil
}

func (s *couponService) GetAllCoupons(page, limit int, filters map[string]interface{}) ([]model.CouponResponse, int64, error) {
	coupons, total, err := s.couponRepo.GetAllCoupons(page, limit, filters)
	if err != nil {
		logger.Errorf("Error getting all coupons: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve coupons")
	}
	var responses []model.CouponResponse
	for _, coupon := range coupons {
		responses = append(responses, *s.toCouponResponse(&coupon))
	}
	return responses, total, nil
}

func (s *couponService) UpdateCoupon(id uint, req *model.CouponUpdateRequest) (*model.CouponResponse, error) {
	coupon, err := s.couponRepo.GetCouponByID(id)
	if err != nil {
		logger.Errorf("Error getting coupon by ID %d for update: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve coupon")
	}
	if coupon == nil {
		return nil, errors.New("coupon not found")
	}

	// Update fields
	if req.Name != "" {
		coupon.Name = req.Name
	}
	if req.Description != "" {
		coupon.Description = req.Description
	}
	if req.Status != "" {
		coupon.Status = req.Status
	}
	if req.DiscountValue != 0 {
		coupon.DiscountValue = req.DiscountValue
	}
	if req.MinOrderAmount != 0 {
		coupon.MinOrderAmount = req.MinOrderAmount
	}
	if req.MaxDiscountAmount != 0 {
		coupon.MaxDiscountAmount = req.MaxDiscountAmount
	}
	if req.UsageLimit != 0 {
		coupon.UsageLimit = req.UsageLimit
	}
	if req.UsagePerUser != 0 {
		coupon.UsagePerUser = req.UsagePerUser
	}
	if req.ValidFrom != nil {
		coupon.ValidFrom = *req.ValidFrom
	}
	if req.ValidTo != nil {
		coupon.ValidTo = *req.ValidTo
	}
	if req.TargetType != "" {
		coupon.TargetType = req.TargetType
	}
	if req.TargetIDs != nil {
		coupon.TargetIDs = convertUintSliceToString(req.TargetIDs)
	}
	if req.IsStackable != nil {
		coupon.IsStackable = *req.IsStackable
	}
	if req.IsFirstTimeOnly != nil {
		coupon.IsFirstTimeOnly = *req.IsFirstTimeOnly
	}
	if req.IsNewUserOnly != nil {
		coupon.IsNewUserOnly = *req.IsNewUserOnly
	}

	// Update status based on dates if not explicitly set
	if req.Status == "" {
		now := time.Now()
		if now.After(coupon.ValidTo) {
			coupon.Status = model.CouponStatusExpired
		} else if now.Before(coupon.ValidFrom) {
			coupon.Status = model.CouponStatusInactive
		} else {
			coupon.Status = model.CouponStatusActive
		}
	}

	if err := s.couponRepo.UpdateCoupon(coupon); err != nil {
		logger.Errorf("Error updating coupon %d: %v", id, err)
		return nil, fmt.Errorf("failed to update coupon")
	}

	return s.toCouponResponse(coupon), nil
}

func (s *couponService) DeleteCoupon(id uint) error {
	coupon, err := s.couponRepo.GetCouponByID(id)
	if err != nil {
		logger.Errorf("Error getting coupon by ID %d for deletion: %v", id, err)
		return fmt.Errorf("failed to retrieve coupon")
	}
	if coupon == nil {
		return errors.New("coupon not found")
	}

	if err := s.couponRepo.DeleteCoupon(id); err != nil {
		logger.Errorf("Error deleting coupon %d: %v", id, err)
		return fmt.Errorf("failed to delete coupon")
	}
	return nil
}

func (s *couponService) GetActiveCoupons(page, limit int) ([]model.CouponResponse, int64, error) {
	coupons, total, err := s.couponRepo.GetActiveCoupons(page, limit)
	if err != nil {
		logger.Errorf("Error getting active coupons: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve active coupons")
	}
	var responses []model.CouponResponse
	for _, coupon := range coupons {
		responses = append(responses, *s.toCouponResponse(&coupon))
	}
	return responses, total, nil
}

func (s *couponService) GetExpiredCoupons(page, limit int) ([]model.CouponResponse, int64, error) {
	coupons, total, err := s.couponRepo.GetExpiredCoupons(page, limit)
	if err != nil {
		logger.Errorf("Error getting expired coupons: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve expired coupons")
	}
	var responses []model.CouponResponse
	for _, coupon := range coupons {
		responses = append(responses, *s.toCouponResponse(&coupon))
	}
	return responses, total, nil
}

func (s *couponService) GetCouponsByType(couponType model.CouponType, page, limit int) ([]model.CouponResponse, int64, error) {
	coupons, total, err := s.couponRepo.GetCouponsByType(couponType, page, limit)
	if err != nil {
		logger.Errorf("Error getting coupons by type %s: %v", couponType, err)
		return nil, 0, fmt.Errorf("failed to retrieve coupons by type")
	}
	var responses []model.CouponResponse
	for _, coupon := range coupons {
		responses = append(responses, *s.toCouponResponse(&coupon))
	}
	return responses, total, nil
}

func (s *couponService) GetCouponsByStatus(status model.CouponStatus, page, limit int) ([]model.CouponResponse, int64, error) {
	coupons, total, err := s.couponRepo.GetCouponsByStatus(status, page, limit)
	if err != nil {
		logger.Errorf("Error getting coupons by status %s: %v", status, err)
		return nil, 0, fmt.Errorf("failed to retrieve coupons by status")
	}
	var responses []model.CouponResponse
	for _, coupon := range coupons {
		responses = append(responses, *s.toCouponResponse(&coupon))
	}
	return responses, total, nil
}

func (s *couponService) SearchCoupons(query string, page, limit int) ([]model.CouponResponse, int64, error) {
	coupons, total, err := s.couponRepo.SearchCoupons(query, page, limit)
	if err != nil {
		logger.Errorf("Error searching coupons: %v", err)
		return nil, 0, fmt.Errorf("failed to search coupons")
	}
	var responses []model.CouponResponse
	for _, coupon := range coupons {
		responses = append(responses, *s.toCouponResponse(&coupon))
	}
	return responses, total, nil
}

func (s *couponService) ValidateCoupon(req *model.CouponValidateRequest) (*model.CouponValidateResponse, error) {
	response, err := s.couponRepo.ValidateCoupon(req.Code, req.UserID, req.OrderAmount, req.ProductIDs)
	if err != nil {
		logger.Errorf("Error validating coupon %s: %v", req.Code, err)
		return nil, fmt.Errorf("failed to validate coupon")
	}
	return response, nil
}

func (s *couponService) UseCoupon(req *model.CouponUseRequest) (*model.CouponUsageResponse, error) {
	// Validate coupon first
	validateReq := &model.CouponValidateRequest{
		Code:        req.Code,
		UserID:      req.UserID,
		OrderAmount: req.OrderAmount,
		ProductIDs:  req.ProductIDs,
	}

	validateResp, err := s.ValidateCoupon(validateReq)
	if err != nil {
		return nil, err
	}
	if !validateResp.Valid {
		return nil, errors.New(validateResp.Message)
	}

	// Get coupon
	coupon, err := s.couponRepo.GetCouponByCode(req.Code)
	if err != nil {
		logger.Errorf("Error getting coupon by code %s: %v", req.Code, err)
		return nil, fmt.Errorf("failed to retrieve coupon")
	}
	if coupon == nil {
		return nil, errors.New("coupon not found")
	}

	// Create usage record
	usage := &model.CouponUsage{
		CouponID:       coupon.ID,
		UserID:         req.UserID,
		OrderID:        req.OrderID,
		DiscountAmount: validateResp.DiscountAmount,
		UsedAt:         time.Now(),
	}

	if err := s.couponRepo.CreateCouponUsage(usage); err != nil {
		logger.Errorf("Error creating coupon usage: %v", err)
		return nil, fmt.Errorf("failed to use coupon")
	}

	return s.toCouponUsageResponse(usage), nil
}

func (s *couponService) GetCouponUsagesByCoupon(couponID uint, page, limit int) ([]model.CouponUsageResponse, int64, error) {
	usages, total, err := s.couponRepo.GetCouponUsagesByCoupon(couponID, page, limit)
	if err != nil {
		logger.Errorf("Error getting coupon usages for coupon %d: %v", couponID, err)
		return nil, 0, fmt.Errorf("failed to retrieve coupon usages")
	}
	var responses []model.CouponUsageResponse
	for _, usage := range usages {
		responses = append(responses, *s.toCouponUsageResponse(&usage))
	}
	return responses, total, nil
}

func (s *couponService) GetCouponUsagesByUser(userID uint, page, limit int) ([]model.CouponUsageResponse, int64, error) {
	usages, total, err := s.couponRepo.GetCouponUsagesByUser(userID, page, limit)
	if err != nil {
		logger.Errorf("Error getting coupon usages for user %d: %v", userID, err)
		return nil, 0, fmt.Errorf("failed to retrieve user coupon usages")
	}
	var responses []model.CouponUsageResponse
	for _, usage := range usages {
		responses = append(responses, *s.toCouponUsageResponse(&usage))
	}
	return responses, total, nil
}

func (s *couponService) GetCouponUsagesByOrder(orderID uint) ([]model.CouponUsageResponse, error) {
	usages, err := s.couponRepo.GetCouponUsagesByOrder(orderID)
	if err != nil {
		logger.Errorf("Error getting coupon usages for order %d: %v", orderID, err)
		return nil, fmt.Errorf("failed to retrieve order coupon usages")
	}
	var responses []model.CouponUsageResponse
	for _, usage := range usages {
		responses = append(responses, *s.toCouponUsageResponse(&usage))
	}
	return responses, nil
}

func (s *couponService) GetCouponStats() (*model.CouponStatsResponse, error) {
	stats, err := s.couponRepo.GetCouponStats()
	if err != nil {
		logger.Errorf("Error getting coupon stats: %v", err)
		return nil, fmt.Errorf("failed to retrieve coupon statistics")
	}
	return stats, nil
}

func (s *couponService) GetCouponUsageStats(couponID uint) (map[string]interface{}, error) {
	stats, err := s.couponRepo.GetCouponUsageStats(couponID)
	if err != nil {
		logger.Errorf("Error getting coupon usage stats for coupon %d: %v", couponID, err)
		return nil, fmt.Errorf("failed to retrieve coupon usage statistics")
	}
	return stats, nil
}

func (s *couponService) toCouponResponse(coupon *model.Coupon) *model.CouponResponse {
	return &model.CouponResponse{
		ID:                coupon.ID,
		Code:              coupon.Code,
		Name:              coupon.Name,
		Description:       coupon.Description,
		Type:              coupon.Type,
		Status:            coupon.Status,
		DiscountValue:     coupon.DiscountValue,
		MinOrderAmount:    coupon.MinOrderAmount,
		MaxDiscountAmount: coupon.MaxDiscountAmount,
		UsageLimit:        coupon.UsageLimit,
		UsageCount:        coupon.UsageCount,
		UsagePerUser:      coupon.UsagePerUser,
		ValidFrom:         coupon.ValidFrom,
		ValidTo:           coupon.ValidTo,
		TargetType:        coupon.TargetType,
		TargetIDs:         convertStringToUintSlice(coupon.TargetIDs),
		IsStackable:       coupon.IsStackable,
		IsFirstTimeOnly:   coupon.IsFirstTimeOnly,
		IsNewUserOnly:     coupon.IsNewUserOnly,
		CreatedBy:         coupon.CreatedBy,
		Creator:           coupon.Creator,
		CreatedAt:         coupon.CreatedAt,
		UpdatedAt:         coupon.UpdatedAt,
	}
}

func (s *couponService) toCouponUsageResponse(usage *model.CouponUsage) *model.CouponUsageResponse {
	return &model.CouponUsageResponse{
		ID:             usage.ID,
		CouponID:       usage.CouponID,
		UserID:         usage.UserID,
		User:           usage.User,
		OrderID:        usage.OrderID,
		Order:          usage.Order,
		DiscountAmount: usage.DiscountAmount,
		OrderAmount:    usage.OrderAmount,
		UsedAt:         usage.UsedAt,
		CreatedAt:      usage.CreatedAt,
		UpdatedAt:      usage.UpdatedAt,
	}
}

// Point Service Implementation

func (s *pointService) GetPointByUserID(userID uint) (*model.PointResponse, error) {
	point, err := s.pointRepo.GetPointByUserID(userID)
	if err != nil {
		logger.Errorf("Error getting point by user ID %d: %v", userID, err)
		return nil, fmt.Errorf("failed to retrieve user points")
	}
	if point == nil {
		return nil, errors.New("user has no point record")
	}
	return s.toPointResponse(point), nil
}

func (s *pointService) GetUserPointBalance(userID uint) (int, error) {
	balance, err := s.pointRepo.GetUserPointBalance(userID)
	if err != nil {
		logger.Errorf("Error getting point balance for user ID %d: %v", userID, err)
		return 0, fmt.Errorf("failed to retrieve user point balance")
	}
	return balance, nil
}

func (s *pointService) EarnPoints(req *model.PointEarnRequest) (*model.PointTransactionResponse, error) {
	// Check if user exists
	user, err := s.userRepo.GetByID(req.UserID)
	if err != nil {
		logger.Errorf("Error getting user by ID %d: %v", req.UserID, err)
		return nil, fmt.Errorf("failed to retrieve user")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	transaction, err := s.pointRepo.EarnPoints(req.UserID, req.Amount, req.ReferenceType, req.ReferenceID, req.Description, req.ExpiryDays)
	if err != nil {
		logger.Errorf("Error earning points for user %d: %v", req.UserID, err)
		return nil, fmt.Errorf("failed to earn points")
	}

	return s.toPointTransactionResponse(transaction), nil
}

func (s *pointService) RedeemPoints(req *model.PointRedeemRequest) (*model.PointTransactionResponse, error) {
	// Check if user exists
	user, err := s.userRepo.GetByID(req.UserID)
	if err != nil {
		logger.Errorf("Error getting user by ID %d: %v", req.UserID, err)
		return nil, fmt.Errorf("failed to retrieve user")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	transaction, err := s.pointRepo.RedeemPoints(req.UserID, req.Amount, req.ReferenceType, req.ReferenceID, req.Description)
	if err != nil {
		logger.Errorf("Error redeeming points for user %d: %v", req.UserID, err)
		return nil, fmt.Errorf("failed to redeem points")
	}

	return s.toPointTransactionResponse(transaction), nil
}

func (s *pointService) RefundPoints(req *model.PointRefundRequest) (*model.PointTransactionResponse, error) {
	// Check if user exists
	user, err := s.userRepo.GetByID(req.UserID)
	if err != nil {
		logger.Errorf("Error getting user by ID %d: %v", req.UserID, err)
		return nil, fmt.Errorf("failed to retrieve user")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	transaction, err := s.pointRepo.RefundPoints(req.UserID, req.Amount, req.ReferenceType, req.ReferenceID, req.Description)
	if err != nil {
		logger.Errorf("Error refunding points for user %d: %v", req.UserID, err)
		return nil, fmt.Errorf("failed to refund points")
	}

	return s.toPointTransactionResponse(transaction), nil
}

func (s *pointService) AdjustPoints(req *model.PointAdjustRequest) (*model.PointTransactionResponse, error) {
	// Check if user exists
	user, err := s.userRepo.GetByID(req.UserID)
	if err != nil {
		logger.Errorf("Error getting user by ID %d: %v", req.UserID, err)
		return nil, fmt.Errorf("failed to retrieve user")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	transaction, err := s.pointRepo.AdjustPoints(req.UserID, req.Amount, req.Description, req.Notes)
	if err != nil {
		logger.Errorf("Error adjusting points for user %d: %v", req.UserID, err)
		return nil, fmt.Errorf("failed to adjust points")
	}

	return s.toPointTransactionResponse(transaction), nil
}

func (s *pointService) ExpirePoints(req *model.PointExpireRequest) (*model.PointTransactionResponse, error) {
	// Check if user exists
	user, err := s.userRepo.GetByID(req.UserID)
	if err != nil {
		logger.Errorf("Error getting user by ID %d: %v", req.UserID, err)
		return nil, fmt.Errorf("failed to retrieve user")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	transaction, err := s.pointRepo.ExpirePoints(req.UserID, req.Amount, req.Description)
	if err != nil {
		logger.Errorf("Error expiring points for user %d: %v", req.UserID, err)
		return nil, fmt.Errorf("failed to expire points")
	}

	return s.toPointTransactionResponse(transaction), nil
}

func (s *pointService) GetPointTransactionByID(id uint) (*model.PointTransactionResponse, error) {
	transaction, err := s.pointRepo.GetPointTransactionByID(id)
	if err != nil {
		logger.Errorf("Error getting point transaction by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve point transaction")
	}
	if transaction == nil {
		return nil, errors.New("point transaction not found")
	}
	return s.toPointTransactionResponse(transaction), nil
}

func (s *pointService) GetPointTransactionsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.PointTransactionResponse, int64, error) {
	transactions, total, err := s.pointRepo.GetPointTransactionsByUser(userID, page, limit, filters)
	if err != nil {
		logger.Errorf("Error getting point transactions for user %d: %v", userID, err)
		return nil, 0, fmt.Errorf("failed to retrieve point transactions")
	}
	var responses []model.PointTransactionResponse
	for _, transaction := range transactions {
		responses = append(responses, *s.toPointTransactionResponse(&transaction))
	}
	return responses, total, nil
}

func (s *pointService) GetPointHistory(userID uint, page, limit int) ([]model.PointTransactionResponse, int64, error) {
	transactions, total, err := s.pointRepo.GetPointHistory(userID, page, limit)
	if err != nil {
		logger.Errorf("Error getting point history for user %d: %v", userID, err)
		return nil, 0, fmt.Errorf("failed to retrieve point history")
	}
	var responses []model.PointTransactionResponse
	for _, transaction := range transactions {
		responses = append(responses, *s.toPointTransactionResponse(&transaction))
	}
	return responses, total, nil
}

func (s *pointService) GetExpiredPoints(userID uint) ([]model.PointTransactionResponse, error) {
	transactions, err := s.pointRepo.GetExpiredPoints(userID)
	if err != nil {
		logger.Errorf("Error getting expired points for user %d: %v", userID, err)
		return nil, fmt.Errorf("failed to retrieve expired points")
	}
	var responses []model.PointTransactionResponse
	for _, transaction := range transactions {
		responses = append(responses, *s.toPointTransactionResponse(&transaction))
	}
	return responses, nil
}

func (s *pointService) GetExpiringPoints(userID uint, days int) ([]model.PointTransactionResponse, error) {
	transactions, err := s.pointRepo.GetExpiringPoints(userID, days)
	if err != nil {
		logger.Errorf("Error getting expiring points for user %d: %v", userID, err)
		return nil, fmt.Errorf("failed to retrieve expiring points")
	}
	var responses []model.PointTransactionResponse
	for _, transaction := range transactions {
		responses = append(responses, *s.toPointTransactionResponse(&transaction))
	}
	return responses, nil
}

func (s *pointService) GetPointStats() (*model.PointStatsResponse, error) {
	stats, err := s.pointRepo.GetPointStats()
	if err != nil {
		logger.Errorf("Error getting point stats: %v", err)
		return nil, fmt.Errorf("failed to retrieve point statistics")
	}
	return stats, nil
}

func (s *pointService) GetUserPointStats(userID uint) (map[string]interface{}, error) {
	stats, err := s.pointRepo.GetUserPointStats(userID)
	if err != nil {
		logger.Errorf("Error getting user point stats for user %d: %v", userID, err)
		return nil, fmt.Errorf("failed to retrieve user point statistics")
	}
	return stats, nil
}

func (s *pointService) GetTopEarners(limit int) ([]model.PointResponse, error) {
	points, err := s.pointRepo.GetTopEarners(limit)
	if err != nil {
		logger.Errorf("Error getting top earners: %v", err)
		return nil, fmt.Errorf("failed to retrieve top earners")
	}
	var responses []model.PointResponse
	for _, point := range points {
		responses = append(responses, *s.toPointResponse(&point))
	}
	return responses, nil
}

func (s *pointService) toPointResponse(point *model.Point) *model.PointResponse {
	return &model.PointResponse{
		ID:            point.ID,
		UserID:        point.UserID,
		User:          point.User,
		Balance:       point.Balance,
		TotalEarned:   point.TotalEarned,
		TotalRedeemed: point.TotalRedeemed,
		TotalExpired:  point.TotalExpired,
		ExpiryDays:    point.ExpiryDays,
		IsActive:      point.IsActive,
		CreatedAt:     point.CreatedAt,
		UpdatedAt:     point.UpdatedAt,
	}
}

func (s *pointService) toPointTransactionResponse(transaction *model.PointTransaction) *model.PointTransactionResponse {
	return &model.PointTransactionResponse{
		ID:            transaction.ID,
		PointID:       transaction.PointID,
		UserID:        transaction.UserID,
		User:          transaction.User,
		Type:          transaction.Type,
		Status:        transaction.Status,
		Amount:        transaction.Amount,
		Balance:       transaction.Balance,
		ReferenceType: transaction.ReferenceType,
		ReferenceID:   transaction.ReferenceID,
		OrderID:       transaction.OrderID,
		Order:         transaction.Order,
		Description:   transaction.Description,
		Notes:         transaction.Notes,
		ExpiresAt:     transaction.ExpiresAt,
		CreatedBy:     transaction.CreatedBy,
		Creator:       transaction.Creator,
		CreatedAt:     transaction.CreatedAt,
		UpdatedAt:     transaction.UpdatedAt,
	}
}

// Helper function to convert uint slice to JSON string
func convertUintSliceToString(ids []uint) string {
	if len(ids) == 0 {
		return "[]"
	}
	strIds := make([]string, len(ids))
	for i, id := range ids {
		strIds[i] = strconv.FormatUint(uint64(id), 10)
	}
	return "[" + strings.Join(strIds, ",") + "]"
}

// Helper function to convert JSON string back to uint slice
func convertStringToUintSlice(str string) []uint {
	if str == "" || str == "[]" {
		return []uint{}
	}
	// Remove brackets and split by comma
	cleanStr := strings.Trim(str, "[]")
	if cleanStr == "" {
		return []uint{}
	}
	parts := strings.Split(cleanStr, ",")
	ids := make([]uint, 0, len(parts))
	for _, part := range parts {
		if id, err := strconv.ParseUint(strings.TrimSpace(part), 10, 32); err == nil {
			ids = append(ids, uint(id))
		}
	}
	return ids
}
