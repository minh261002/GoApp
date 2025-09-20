package handler

import (
	"net/http"
	"strconv"

	"go_app/internal/model"
	"go_app/internal/service"
	"go_app/pkg/response"

	"github.com/gin-gonic/gin"
)

// ReviewHandler handles review-related HTTP requests
type ReviewHandler struct {
	reviewService service.ReviewService
}

// NewReviewHandler creates a new ReviewHandler
func NewReviewHandler() *ReviewHandler {
	return &ReviewHandler{
		reviewService: service.NewReviewService(),
	}
}

// Basic CRUD

// CreateReview creates a new review
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	var req model.ReviewCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	review, err := h.reviewService.CreateReview(&req, userID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create review", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Review created successfully", review)
}

// GetReviewByID retrieves a review by its ID
func (h *ReviewHandler) GetReviewByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid review ID", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	review, err := h.reviewService.GetReviewByID(uint(id), userID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve review", err.Error())
		return
	}

	if review == nil {
		response.ErrorResponse(c, http.StatusNotFound, "Review not found", "review not found")
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Review retrieved successfully", review)
}

// GetAllReviews retrieves all reviews with pagination and filters
func (h *ReviewHandler) GetAllReviews(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Parse filters
	filters := make(map[string]interface{})
	if userID := c.Query("user_id"); userID != "" {
		if id, err := strconv.ParseUint(userID, 10, 32); err == nil {
			filters["user_id"] = uint(id)
		}
	}
	if productID := c.Query("product_id"); productID != "" {
		if id, err := strconv.ParseUint(productID, 10, 32); err == nil {
			filters["product_id"] = uint(id)
		}
	}
	if orderID := c.Query("order_id"); orderID != "" {
		if id, err := strconv.ParseUint(orderID, 10, 32); err == nil {
			filters["order_id"] = uint(id)
		}
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if reviewType := c.Query("type"); reviewType != "" {
		filters["type"] = reviewType
	}
	if rating := c.Query("rating"); rating != "" {
		if r, err := strconv.Atoi(rating); err == nil {
			filters["rating"] = r
		}
	}
	if isVerified := c.Query("is_verified"); isVerified != "" {
		if verified, err := strconv.ParseBool(isVerified); err == nil {
			filters["is_verified"] = verified
		}
	}
	if isAnonymous := c.Query("is_anonymous"); isAnonymous != "" {
		if anonymous, err := strconv.ParseBool(isAnonymous); err == nil {
			filters["is_anonymous"] = anonymous
		}
	}
	if moderatedBy := c.Query("moderated_by"); moderatedBy != "" {
		if id, err := strconv.ParseUint(moderatedBy, 10, 32); err == nil {
			filters["moderated_by"] = uint(id)
		}
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		filters["date_from"] = dateFrom
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		filters["date_to"] = dateTo
	}

	reviews, total, err := h.reviewService.GetAllReviews(page, limit, filters)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve reviews", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Reviews retrieved successfully", reviews, page, limit, total)
}

// GetReviewsByProduct retrieves reviews for a specific product
func (h *ReviewHandler) GetReviewsByProduct(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", err.Error())
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Parse filters
	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if reviewType := c.Query("type"); reviewType != "" {
		filters["type"] = reviewType
	}
	if rating := c.Query("rating"); rating != "" {
		if r, err := strconv.Atoi(rating); err == nil {
			filters["rating"] = r
		}
	}
	if isVerified := c.Query("is_verified"); isVerified != "" {
		if verified, err := strconv.ParseBool(isVerified); err == nil {
			filters["is_verified"] = verified
		}
	}

	reviews, total, err := h.reviewService.GetReviewsByProduct(uint(productID), page, limit, filters)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve product reviews", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Product reviews retrieved successfully", reviews, page, limit, total)
}

// GetReviewsByUser retrieves reviews for a specific user
func (h *ReviewHandler) GetReviewsByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Parse filters
	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if reviewType := c.Query("type"); reviewType != "" {
		filters["type"] = reviewType
	}
	if rating := c.Query("rating"); rating != "" {
		if r, err := strconv.Atoi(rating); err == nil {
			filters["rating"] = r
		}
	}

	reviews, total, err := h.reviewService.GetReviewsByUser(uint(userID), page, limit, filters)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve user reviews", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "User reviews retrieved successfully", reviews, page, limit, total)
}

// UpdateReview updates an existing review
func (h *ReviewHandler) UpdateReview(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid review ID", err.Error())
		return
	}

	var req model.ReviewUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	review, err := h.reviewService.UpdateReview(uint(id), &req, userID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update review", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Review updated successfully", review)
}

// DeleteReview deletes a review
func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid review ID", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	if err := h.reviewService.DeleteReview(uint(id), userID.(uint)); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete review", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Review deleted successfully", nil)
}

// Review Management

// GetReviewsByStatus retrieves reviews by status
func (h *ReviewHandler) GetReviewsByStatus(c *gin.Context) {
	statusStr := c.Param("status")
	status := model.ReviewStatus(statusStr)
	if status != model.ReviewStatusPending && status != model.ReviewStatusApproved &&
		status != model.ReviewStatusRejected && status != model.ReviewStatusHidden {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid review status", "invalid review status")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	reviews, total, err := h.reviewService.GetReviewsByStatus(status, page, limit)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve reviews by status", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Reviews by status retrieved successfully", reviews, page, limit, total)
}

// GetReviewsByType retrieves reviews by type
func (h *ReviewHandler) GetReviewsByType(c *gin.Context) {
	typeStr := c.Param("type")
	reviewType := model.ReviewType(typeStr)
	if reviewType != model.ReviewTypeProduct && reviewType != model.ReviewTypeService &&
		reviewType != model.ReviewTypeOrder && reviewType != model.ReviewTypeOverall {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid review type", "invalid review type")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	reviews, total, err := h.reviewService.GetReviewsByType(reviewType, page, limit)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve reviews by type", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Reviews by type retrieved successfully", reviews, page, limit, total)
}

// GetVerifiedReviews retrieves verified reviews for a product
func (h *ReviewHandler) GetVerifiedReviews(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", err.Error())
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	reviews, total, err := h.reviewService.GetVerifiedReviews(uint(productID), page, limit)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve verified reviews", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Verified reviews retrieved successfully", reviews, page, limit, total)
}

// GetRecentReviews retrieves recent reviews
func (h *ReviewHandler) GetRecentReviews(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	reviews, err := h.reviewService.GetRecentReviews(limit)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve recent reviews", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Recent reviews retrieved successfully", reviews)
}

// SearchReviews performs full-text search on reviews
func (h *ReviewHandler) SearchReviews(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Search query is required", "q parameter is required")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	reviews, total, err := h.reviewService.SearchReviews(query, page, limit)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to search reviews", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Review search completed successfully", reviews, page, limit, total)
}

// Rating & Statistics

// GetAverageRating retrieves average rating for a product
func (h *ReviewHandler) GetAverageRating(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", err.Error())
		return
	}

	avg, err := h.reviewService.GetAverageRating(uint(productID))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve average rating", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Average rating retrieved successfully", map[string]interface{}{
		"product_id":     productID,
		"average_rating": avg,
	})
}

// GetRatingDistribution retrieves rating distribution for a product
func (h *ReviewHandler) GetRatingDistribution(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", err.Error())
		return
	}

	distribution, err := h.reviewService.GetRatingDistribution(uint(productID))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve rating distribution", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Rating distribution retrieved successfully", map[string]interface{}{
		"product_id":          productID,
		"rating_distribution": distribution,
	})
}

// GetReviewStats retrieves review statistics
func (h *ReviewHandler) GetReviewStats(c *gin.Context) {
	stats, err := h.reviewService.GetReviewStats()
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve review statistics", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Review statistics retrieved successfully", stats)
}

// GetProductReviewStats retrieves product-specific review statistics
func (h *ReviewHandler) GetProductReviewStats(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", err.Error())
		return
	}

	stats, err := h.reviewService.GetProductReviewStats(uint(productID))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve product review statistics", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Product review statistics retrieved successfully", stats)
}

// GetUserReviewStats retrieves user-specific review statistics
func (h *ReviewHandler) GetUserReviewStats(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	stats, err := h.reviewService.GetUserReviewStats(uint(userID))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve user review statistics", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "User review statistics retrieved successfully", stats)
}

// Helpful Votes

// CreateHelpfulVote creates a new helpful vote
func (h *ReviewHandler) CreateHelpfulVote(c *gin.Context) {
	reviewIDStr := c.Param("review_id")
	reviewID, err := strconv.ParseUint(reviewIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid review ID", err.Error())
		return
	}

	var req struct {
		IsHelpful bool `json:"is_helpful" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	vote, err := h.reviewService.CreateHelpfulVote(uint(reviewID), userID.(uint), req.IsHelpful)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create helpful vote", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Helpful vote created successfully", vote)
}

// UpdateHelpfulVote updates an existing helpful vote
func (h *ReviewHandler) UpdateHelpfulVote(c *gin.Context) {
	reviewIDStr := c.Param("review_id")
	reviewID, err := strconv.ParseUint(reviewIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid review ID", err.Error())
		return
	}

	var req struct {
		IsHelpful bool `json:"is_helpful" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	vote, err := h.reviewService.UpdateHelpfulVote(uint(reviewID), userID.(uint), req.IsHelpful)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update helpful vote", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Helpful vote updated successfully", vote)
}

// DeleteHelpfulVote deletes a helpful vote
func (h *ReviewHandler) DeleteHelpfulVote(c *gin.Context) {
	reviewIDStr := c.Param("review_id")
	reviewID, err := strconv.ParseUint(reviewIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid review ID", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	if err := h.reviewService.DeleteHelpfulVote(uint(reviewID), userID.(uint)); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete helpful vote", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Helpful vote deleted successfully", nil)
}

// GetHelpfulVotesByReview retrieves helpful votes for a review
func (h *ReviewHandler) GetHelpfulVotesByReview(c *gin.Context) {
	reviewIDStr := c.Param("review_id")
	reviewID, err := strconv.ParseUint(reviewIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid review ID", err.Error())
		return
	}

	votes, err := h.reviewService.GetHelpfulVotesByReview(uint(reviewID))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve helpful votes", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Helpful votes retrieved successfully", votes)
}

// Review Images

// GetReviewImagesByReview retrieves images for a review
func (h *ReviewHandler) GetReviewImagesByReview(c *gin.Context) {
	reviewIDStr := c.Param("review_id")
	reviewID, err := strconv.ParseUint(reviewIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid review ID", err.Error())
		return
	}

	images, err := h.reviewService.GetReviewImagesByReview(uint(reviewID))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve review images", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Review images retrieved successfully", images)
}

// DeleteReviewImage deletes a review image
func (h *ReviewHandler) DeleteReviewImage(c *gin.Context) {
	imageIDStr := c.Param("image_id")
	imageID, err := strconv.ParseUint(imageIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid image ID", err.Error())
		return
	}

	if err := h.reviewService.DeleteReviewImage(uint(imageID)); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete review image", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Review image deleted successfully", nil)
}

// Moderation

// ModerateReview moderates a review
func (h *ReviewHandler) ModerateReview(c *gin.Context) {
	reviewIDStr := c.Param("review_id")
	reviewID, err := strconv.ParseUint(reviewIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid review ID", err.Error())
		return
	}

	var req model.ReviewModerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get moderator ID from context
	moderatorID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	review, err := h.reviewService.ModerateReview(uint(reviewID), &req, moderatorID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to moderate review", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Review moderated successfully", review)
}

// GetPendingReviews retrieves pending reviews
func (h *ReviewHandler) GetPendingReviews(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	reviews, total, err := h.reviewService.GetPendingReviews(page, limit)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve pending reviews", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Pending reviews retrieved successfully", reviews, page, limit, total)
}

// GetModeratedReviews retrieves reviews moderated by a specific moderator
func (h *ReviewHandler) GetModeratedReviews(c *gin.Context) {
	moderatorIDStr := c.Param("moderator_id")
	moderatorID, err := strconv.ParseUint(moderatorIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid moderator ID", err.Error())
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	reviews, total, err := h.reviewService.GetModeratedReviews(uint(moderatorID), page, limit)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve moderated reviews", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Moderated reviews retrieved successfully", reviews, page, limit, total)
}
