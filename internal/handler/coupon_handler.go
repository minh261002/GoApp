package handler

import (
	"net/http"
	"strconv"
	"time"

	"go_app/internal/model"
	"go_app/internal/repository"
	"go_app/internal/service"
	"go_app/pkg/response"

	"github.com/gin-gonic/gin"
)

type CouponHandler struct {
	couponService service.CouponService
	pointService  service.PointService
}

func NewCouponHandler() *CouponHandler {
	return &CouponHandler{
		couponService: service.NewCouponService(
			repository.NewCouponRepository(),
			repository.NewUserRepository(),
			repository.NewOrderRepository(),
		),
		pointService: service.NewPointService(
			repository.NewPointRepository(),
			repository.NewUserRepository(),
			repository.NewOrderRepository(),
		),
	}
}

// Coupon Handlers

// CreateCoupon creates a new coupon
func (h *CouponHandler) CreateCoupon(c *gin.Context) {
	var req model.CouponCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	// Get creator ID from context (set by auth middleware)
	creatorID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", "")
		return
	}

	coupon, err := h.couponService.CreateCoupon(&req, creatorID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create coupon", err.Error())
		return
	}

	response.Success(c, *coupon, "Coupon created successfully")
}

// GetCouponByID retrieves a coupon by ID
func (h *CouponHandler) GetCouponByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid coupon ID", err.Error())
		return
	}

	coupon, err := h.couponService.GetCouponByID(uint(id))
	if err != nil {
		if err.Error() == "coupon not found" {
			response.Error(c, http.StatusNotFound, "Coupon not found", "")
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve coupon", err.Error())
		return
	}

	response.Success(c, *coupon, "Coupon retrieved successfully")
}

// GetCouponByCode retrieves a coupon by code
func (h *CouponHandler) GetCouponByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.Error(c, http.StatusBadRequest, "Coupon code is required", "")
		return
	}

	coupon, err := h.couponService.GetCouponByCode(code)
	if err != nil {
		if err.Error() == "coupon not found" {
			response.Error(c, http.StatusNotFound, "Coupon not found", "")
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve coupon", err.Error())
		return
	}

	response.Success(c, *coupon, "Coupon retrieved successfully")
}

// GetAllCoupons retrieves all coupons with pagination and filters
func (h *CouponHandler) GetAllCoupons(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Build filters from query parameters
	filters := make(map[string]interface{})
	if code := c.Query("code"); code != "" {
		filters["code"] = code
	}
	if name := c.Query("name"); name != "" {
		filters["name"] = name
	}
	if couponType := c.Query("type"); couponType != "" {
		filters["type"] = couponType
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if createdBy := c.Query("created_by"); createdBy != "" {
		if createdByID, err := strconv.ParseUint(createdBy, 10, 32); err == nil {
			filters["created_by"] = createdByID
		}
	}
	if isStackable := c.Query("is_stackable"); isStackable != "" {
		if stackable, err := strconv.ParseBool(isStackable); err == nil {
			filters["is_stackable"] = stackable
		}
	}
	if isFirstTimeOnly := c.Query("is_first_time_only"); isFirstTimeOnly != "" {
		if firstTime, err := strconv.ParseBool(isFirstTimeOnly); err == nil {
			filters["is_first_time_only"] = firstTime
		}
	}
	if isNewUserOnly := c.Query("is_new_user_only"); isNewUserOnly != "" {
		if newUser, err := strconv.ParseBool(isNewUserOnly); err == nil {
			filters["is_new_user_only"] = newUser
		}
	}
	if validFrom := c.Query("valid_from"); validFrom != "" {
		if from, err := time.Parse("2006-01-02", validFrom); err == nil {
			filters["valid_from"] = from
		}
	}
	if validTo := c.Query("valid_to"); validTo != "" {
		if to, err := time.Parse("2006-01-02", validTo); err == nil {
			filters["valid_to"] = to
		}
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	coupons, total, err := h.couponService.GetAllCoupons(page, limit, filters)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve coupons", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Coupons retrieved successfully", coupons, page, limit, total)
}

// UpdateCoupon updates an existing coupon
func (h *CouponHandler) UpdateCoupon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid coupon ID", err.Error())
		return
	}

	var req model.CouponUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	coupon, err := h.couponService.UpdateCoupon(uint(id), &req)
	if err != nil {
		if err.Error() == "coupon not found" {
			response.Error(c, http.StatusNotFound, "Coupon not found", "")
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update coupon", err.Error())
		return
	}

	response.Success(c, *coupon, "Coupon updated successfully")
}

// DeleteCoupon deletes a coupon
func (h *CouponHandler) DeleteCoupon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid coupon ID", err.Error())
		return
	}

	err = h.couponService.DeleteCoupon(uint(id))
	if err != nil {
		if err.Error() == "coupon not found" {
			response.Error(c, http.StatusNotFound, "Coupon not found", "")
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to delete coupon", err.Error())
		return
	}

	response.Success(c, nil, "Coupon deleted successfully")
}

// GetActiveCoupons retrieves active coupons
func (h *CouponHandler) GetActiveCoupons(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	coupons, total, err := h.couponService.GetActiveCoupons(page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve active coupons", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Active coupons retrieved successfully", coupons, page, limit, total)
}

// GetExpiredCoupons retrieves expired coupons
func (h *CouponHandler) GetExpiredCoupons(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	coupons, total, err := h.couponService.GetExpiredCoupons(page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve expired coupons", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Expired coupons retrieved successfully", coupons, page, limit, total)
}

// GetCouponsByType retrieves coupons by type
func (h *CouponHandler) GetCouponsByType(c *gin.Context) {
	couponType := c.Param("type")
	if couponType == "" {
		response.Error(c, http.StatusBadRequest, "Coupon type is required", "")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	coupons, total, err := h.couponService.GetCouponsByType(model.CouponType(couponType), page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve coupons by type", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Coupons by type retrieved successfully", coupons, page, limit, total)
}

// GetCouponsByStatus retrieves coupons by status
func (h *CouponHandler) GetCouponsByStatus(c *gin.Context) {
	status := c.Param("status")
	if status == "" {
		response.Error(c, http.StatusBadRequest, "Coupon status is required", "")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	coupons, total, err := h.couponService.GetCouponsByStatus(model.CouponStatus(status), page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve coupons by status", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Coupons by status retrieved successfully", coupons, page, limit, total)
}

// SearchCoupons searches coupons
func (h *CouponHandler) SearchCoupons(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.Error(c, http.StatusBadRequest, "Search query is required", "")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	coupons, total, err := h.couponService.SearchCoupons(query, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to search coupons", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Coupon search completed successfully", coupons, page, limit, total)
}

// ValidateCoupon validates a coupon
func (h *CouponHandler) ValidateCoupon(c *gin.Context) {
	var req model.CouponValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	result, err := h.couponService.ValidateCoupon(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to validate coupon", err.Error())
		return
	}

	response.Success(c, *result, "Coupon validation completed")
}

// UseCoupon uses a coupon
func (h *CouponHandler) UseCoupon(c *gin.Context) {
	var req model.CouponUseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	usage, err := h.couponService.UseCoupon(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to use coupon", err.Error())
		return
	}

	response.Success(c, *usage, "Coupon used successfully")
}

// GetCouponUsagesByCoupon retrieves coupon usages for a specific coupon
func (h *CouponHandler) GetCouponUsagesByCoupon(c *gin.Context) {
	couponIDStr := c.Param("coupon_id")
	couponID, err := strconv.ParseUint(couponIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid coupon ID", err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	usages, total, err := h.couponService.GetCouponUsagesByCoupon(uint(couponID), page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve coupon usages", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Coupon usages retrieved successfully", usages, page, limit, total)
}

// GetCouponUsagesByUser retrieves coupon usages for a specific user
func (h *CouponHandler) GetCouponUsagesByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	usages, total, err := h.couponService.GetCouponUsagesByUser(uint(userID), page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve user coupon usages", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "User coupon usages retrieved successfully", usages, page, limit, total)
}

// GetCouponUsagesByOrder retrieves coupon usages for a specific order
func (h *CouponHandler) GetCouponUsagesByOrder(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid order ID", err.Error())
		return
	}

	usages, err := h.couponService.GetCouponUsagesByOrder(uint(orderID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve order coupon usages", err.Error())
		return
	}

	response.Success(c, usages, "Order coupon usages retrieved successfully")
}

// GetCouponStats retrieves coupon statistics
func (h *CouponHandler) GetCouponStats(c *gin.Context) {
	stats, err := h.couponService.GetCouponStats()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve coupon statistics", err.Error())
		return
	}

	response.Success(c, *stats, "Coupon statistics retrieved successfully")
}

// GetCouponUsageStats retrieves usage statistics for a specific coupon
func (h *CouponHandler) GetCouponUsageStats(c *gin.Context) {
	couponIDStr := c.Param("coupon_id")
	couponID, err := strconv.ParseUint(couponIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid coupon ID", err.Error())
		return
	}

	stats, err := h.couponService.GetCouponUsageStats(uint(couponID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve coupon usage statistics", err.Error())
		return
	}

	response.Success(c, stats, "Coupon usage statistics retrieved successfully")
}

// Point Handlers

// GetPointByUserID retrieves point information for a user
func (h *CouponHandler) GetPointByUserID(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	point, err := h.pointService.GetPointByUserID(uint(userID))
	if err != nil {
		if err.Error() == "user has no point record" {
			response.Error(c, http.StatusNotFound, "User has no point record", "")
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve user points", err.Error())
		return
	}

	response.Success(c, *point, "User points retrieved successfully")
}

// GetUserPointBalance retrieves point balance for a user
func (h *CouponHandler) GetUserPointBalance(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	balance, err := h.pointService.GetUserPointBalance(uint(userID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve user point balance", err.Error())
		return
	}

	response.Success(c, map[string]int{"balance": balance}, "User point balance retrieved successfully")
}

// EarnPoints earns points for a user
func (h *CouponHandler) EarnPoints(c *gin.Context) {
	var req model.PointEarnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	transaction, err := h.pointService.EarnPoints(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to earn points", err.Error())
		return
	}

	response.Success(c, *transaction, "Points earned successfully")
}

// RedeemPoints redeems points for a user
func (h *CouponHandler) RedeemPoints(c *gin.Context) {
	var req model.PointRedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	transaction, err := h.pointService.RedeemPoints(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to redeem points", err.Error())
		return
	}

	response.Success(c, *transaction, "Points redeemed successfully")
}

// RefundPoints refunds points to a user
func (h *CouponHandler) RefundPoints(c *gin.Context) {
	var req model.PointRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	transaction, err := h.pointService.RefundPoints(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to refund points", err.Error())
		return
	}

	response.Success(c, *transaction, "Points refunded successfully")
}

// AdjustPoints adjusts points for a user (admin function)
func (h *CouponHandler) AdjustPoints(c *gin.Context) {
	var req model.PointAdjustRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	transaction, err := h.pointService.AdjustPoints(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to adjust points", err.Error())
		return
	}

	response.Success(c, *transaction, "Points adjusted successfully")
}

// ExpirePoints expires points for a user (admin function)
func (h *CouponHandler) ExpirePoints(c *gin.Context) {
	var req model.PointExpireRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	transaction, err := h.pointService.ExpirePoints(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to expire points", err.Error())
		return
	}

	response.Success(c, *transaction, "Points expired successfully")
}

// GetPointTransactionByID retrieves a point transaction by ID
func (h *CouponHandler) GetPointTransactionByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid transaction ID", err.Error())
		return
	}

	transaction, err := h.pointService.GetPointTransactionByID(uint(id))
	if err != nil {
		if err.Error() == "point transaction not found" {
			response.Error(c, http.StatusNotFound, "Point transaction not found", "")
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve point transaction", err.Error())
		return
	}

	response.Success(c, *transaction, "Point transaction retrieved successfully")
}

// GetPointTransactionsByUser retrieves point transactions for a user
func (h *CouponHandler) GetPointTransactionsByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Build filters from query parameters
	filters := make(map[string]interface{})
	if transactionType := c.Query("type"); transactionType != "" {
		filters["type"] = transactionType
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if referenceType := c.Query("reference_type"); referenceType != "" {
		filters["reference_type"] = referenceType
	}
	if referenceID := c.Query("reference_id"); referenceID != "" {
		if refID, err := strconv.ParseUint(referenceID, 10, 32); err == nil {
			filters["reference_id"] = refID
		}
	}
	if orderID := c.Query("order_id"); orderID != "" {
		if ordID, err := strconv.ParseUint(orderID, 10, 32); err == nil {
			filters["order_id"] = ordID
		}
	}
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if from, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filters["date_from"] = from
		}
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		if to, err := time.Parse("2006-01-02", dateTo); err == nil {
			filters["date_to"] = to
		}
	}

	transactions, total, err := h.pointService.GetPointTransactionsByUser(uint(userID), page, limit, filters)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve point transactions", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Point transactions retrieved successfully", transactions, page, limit, total)
}

// GetPointHistory retrieves point history for a user
func (h *CouponHandler) GetPointHistory(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	transactions, total, err := h.pointService.GetPointHistory(uint(userID), page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve point history", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Point history retrieved successfully", transactions, page, limit, total)
}

// GetExpiredPoints retrieves expired points for a user
func (h *CouponHandler) GetExpiredPoints(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	transactions, err := h.pointService.GetExpiredPoints(uint(userID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve expired points", err.Error())
		return
	}

	response.Success(c, transactions, "Expired points retrieved successfully")
}

// GetExpiringPoints retrieves points expiring within specified days
func (h *CouponHandler) GetExpiringPoints(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))

	transactions, err := h.pointService.GetExpiringPoints(uint(userID), days)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve expiring points", err.Error())
		return
	}

	response.Success(c, transactions, "Expiring points retrieved successfully")
}

// GetPointStats retrieves point statistics
func (h *CouponHandler) GetPointStats(c *gin.Context) {
	stats, err := h.pointService.GetPointStats()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve point statistics", err.Error())
		return
	}

	response.Success(c, *stats, "Point statistics retrieved successfully")
}

// GetUserPointStats retrieves point statistics for a specific user
func (h *CouponHandler) GetUserPointStats(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	stats, err := h.pointService.GetUserPointStats(uint(userID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve user point statistics", err.Error())
		return
	}

	response.Success(c, stats, "User point statistics retrieved successfully")
}

// GetTopEarners retrieves top earning users
func (h *CouponHandler) GetTopEarners(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	points, err := h.pointService.GetTopEarners(limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve top earners", err.Error())
		return
	}

	response.Success(c, points, "Top earners retrieved successfully")
}
