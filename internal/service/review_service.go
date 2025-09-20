package service

import (
	"errors"
	"fmt"
	"go_app/internal/model"
	"go_app/internal/repository"
	"go_app/pkg/logger"
)

// ReviewService defines methods for review business logic
type ReviewService interface {
	// Basic CRUD
	CreateReview(req *model.ReviewCreateRequest, userID uint) (*model.ReviewResponse, error)
	GetReviewByID(id uint, userID uint) (*model.ReviewResponse, error)
	GetAllReviews(page, limit int, filters map[string]interface{}) ([]model.ReviewResponse, int64, error)
	GetReviewsByProduct(productID uint, page, limit int, filters map[string]interface{}) ([]model.ReviewResponse, int64, error)
	GetReviewsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.ReviewResponse, int64, error)
	UpdateReview(id uint, req *model.ReviewUpdateRequest, userID uint) (*model.ReviewResponse, error)
	DeleteReview(id uint, userID uint) error

	// Review Management
	GetReviewsByStatus(status model.ReviewStatus, page, limit int) ([]model.ReviewResponse, int64, error)
	GetReviewsByType(reviewType model.ReviewType, page, limit int) ([]model.ReviewResponse, int64, error)
	GetVerifiedReviews(productID uint, page, limit int) ([]model.ReviewResponse, int64, error)
	GetRecentReviews(limit int) ([]model.ReviewResponse, error)
	SearchReviews(query string, page, limit int) ([]model.ReviewResponse, int64, error)

	// Rating & Statistics
	GetAverageRating(productID uint) (float64, error)
	GetRatingDistribution(productID uint) (map[int]int64, error)
	GetReviewStats() (*model.ReviewStatsResponse, error)
	GetProductReviewStats(productID uint) (*model.ProductReviewStatsResponse, error)
	GetUserReviewStats(userID uint) (map[string]interface{}, error)

	// Helpful Votes
	CreateHelpfulVote(reviewID uint, userID uint, isHelpful bool) (*model.ReviewHelpfulVoteResponse, error)
	UpdateHelpfulVote(reviewID uint, userID uint, isHelpful bool) (*model.ReviewHelpfulVoteResponse, error)
	DeleteHelpfulVote(reviewID uint, userID uint) error
	GetHelpfulVotesByReview(reviewID uint) ([]model.ReviewHelpfulVoteResponse, error)

	// Review Images
	CreateReviewImage(reviewID uint, imageURL, imagePath, altText string, sortOrder int) (*model.ReviewImageResponse, error)
	GetReviewImagesByReview(reviewID uint) ([]model.ReviewImageResponse, error)
	UpdateReviewImage(imageID uint, imageURL, imagePath, altText string, sortOrder int) (*model.ReviewImageResponse, error)
	DeleteReviewImage(imageID uint) error

	// Moderation
	ModerateReview(reviewID uint, req *model.ReviewModerationRequest, moderatorID uint) (*model.ReviewResponse, error)
	GetPendingReviews(page, limit int) ([]model.ReviewResponse, int64, error)
	GetModeratedReviews(moderatorID uint, page, limit int) ([]model.ReviewResponse, int64, error)

	// Utility
	ValidateReview(review *model.Review) error
	CanUserReviewProduct(userID, productID uint) (bool, error)
	UpdateReviewHelpfulCounts(reviewID uint) error
}

// reviewService implements ReviewService
type reviewService struct {
	reviewRepo  repository.ReviewRepository
	userRepo    repository.UserRepository
	productRepo *repository.ProductRepository
	orderRepo   repository.OrderRepository
}

// NewReviewService creates a new ReviewService
func NewReviewService() ReviewService {
	return &reviewService{
		reviewRepo:  repository.NewReviewRepository(),
		userRepo:    repository.NewUserRepository(),
		productRepo: repository.NewProductRepository(),
		orderRepo:   repository.NewOrderRepository(),
	}
}

// Basic CRUD

// CreateReview creates a new review
func (s *reviewService) CreateReview(req *model.ReviewCreateRequest, userID uint) (*model.ReviewResponse, error) {
	// Get user information
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		logger.Errorf("Error getting user by ID %d: %v", userID, err)
		return nil, fmt.Errorf("failed to retrieve user")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Get product information
	product, err := s.productRepo.GetByID(req.ProductID)
	if err != nil {
		logger.Errorf("Error getting product by ID %d: %v", req.ProductID, err)
		return nil, fmt.Errorf("failed to retrieve product")
	}
	if product == nil {
		return nil, errors.New("product not found")
	}

	// Check if user can review this product
	canReview, err := s.CanUserReviewProduct(userID, req.ProductID)
	if err != nil {
		logger.Errorf("Error checking if user can review product: %v", err)
		return nil, fmt.Errorf("failed to check review eligibility")
	}
	if !canReview {
		return nil, errors.New("you cannot review this product")
	}

	// Create review
	review := &model.Review{
		UserID:      userID,
		ProductID:   req.ProductID,
		OrderID:     req.OrderID,
		Title:       req.Title,
		Content:     req.Content,
		Rating:      req.Rating,
		Type:        req.Type,
		IsAnonymous: req.IsAnonymous,
		Status:      model.ReviewStatusPending, // Default to pending for moderation
	}

	// Set default type if not provided
	if review.Type == "" {
		review.Type = model.ReviewTypeProduct
	}

	// Validate review
	if err := s.ValidateReview(review); err != nil {
		logger.Errorf("Review validation failed: %v", err)
		return nil, err
	}

	// Create review in database
	if err := s.reviewRepo.CreateReview(review); err != nil {
		logger.Errorf("Error creating review: %v", err)
		return nil, fmt.Errorf("failed to create review")
	}

	// Create review images if provided
	if len(req.ImageURLs) > 0 {
		for i, imageURL := range req.ImageURLs {
			altText := ""
			if i < len(req.AltTexts) {
				altText = req.AltTexts[i]
			}

			image := &model.ReviewImage{
				ReviewID:  review.ID,
				ImageURL:  imageURL,
				ImagePath: imageURL, // Assuming imageURL is the path
				AltText:   altText,
				SortOrder: i,
			}

			if err := s.reviewRepo.CreateReviewImage(image); err != nil {
				logger.Warnf("Failed to create review image: %v", err)
			}
		}
	}

	// Get created review with relations
	createdReview, err := s.reviewRepo.GetReviewByID(review.ID)
	if err != nil {
		logger.Errorf("Error getting created review: %v", err)
		return nil, fmt.Errorf("failed to retrieve created review")
	}

	return s.toReviewResponse(createdReview), nil
}

// GetReviewByID retrieves a review by its ID
func (s *reviewService) GetReviewByID(id uint, userID uint) (*model.ReviewResponse, error) {
	review, err := s.reviewRepo.GetReviewByID(id)
	if err != nil {
		logger.Errorf("Error getting review by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve review")
	}
	if review == nil {
		return nil, errors.New("review not found")
	}

	// Check if user can view this review
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		logger.Errorf("Error getting user by ID %d: %v", userID, err)
		return nil, fmt.Errorf("failed to retrieve user")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Allow access if user owns the review, is admin, or review is approved
	if review.UserID != userID && user.Role != "admin" && user.Role != "super_admin" && review.Status != model.ReviewStatusApproved {
		return nil, errors.New("access denied: you can only view your own reviews or approved reviews")
	}

	return s.toReviewResponse(review), nil
}

// GetAllReviews retrieves all reviews with pagination and filters
func (s *reviewService) GetAllReviews(page, limit int, filters map[string]interface{}) ([]model.ReviewResponse, int64, error) {
	reviews, total, err := s.reviewRepo.GetAllReviews(page, limit, filters)
	if err != nil {
		logger.Errorf("Error getting reviews: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve reviews")
	}

	var responses []model.ReviewResponse
	for _, review := range reviews {
		responses = append(responses, *s.toReviewResponse(&review))
	}
	return responses, total, nil
}

// GetReviewsByProduct retrieves reviews for a specific product
func (s *reviewService) GetReviewsByProduct(productID uint, page, limit int, filters map[string]interface{}) ([]model.ReviewResponse, int64, error) {
	reviews, total, err := s.reviewRepo.GetReviewsByProduct(productID, page, limit, filters)
	if err != nil {
		logger.Errorf("Error getting product reviews: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve product reviews")
	}

	var responses []model.ReviewResponse
	for _, review := range reviews {
		responses = append(responses, *s.toReviewResponse(&review))
	}
	return responses, total, nil
}

// GetReviewsByUser retrieves reviews for a specific user
func (s *reviewService) GetReviewsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.ReviewResponse, int64, error) {
	reviews, total, err := s.reviewRepo.GetReviewsByUser(userID, page, limit, filters)
	if err != nil {
		logger.Errorf("Error getting user reviews: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve user reviews")
	}

	var responses []model.ReviewResponse
	for _, review := range reviews {
		responses = append(responses, *s.toReviewResponse(&review))
	}
	return responses, total, nil
}

// UpdateReview updates an existing review
func (s *reviewService) UpdateReview(id uint, req *model.ReviewUpdateRequest, userID uint) (*model.ReviewResponse, error) {
	review, err := s.reviewRepo.GetReviewByID(id)
	if err != nil {
		logger.Errorf("Error getting review by ID %d for update: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve review")
	}
	if review == nil {
		return nil, errors.New("review not found")
	}

	// Check if user can update this review
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		logger.Errorf("Error getting user by ID %d: %v", userID, err)
		return nil, fmt.Errorf("failed to retrieve user")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	if review.UserID != userID && user.Role != "admin" && user.Role != "super_admin" {
		return nil, errors.New("access denied: you can only update your own reviews")
	}

	// Check if review can be updated
	if !review.CanBeEdited() {
		return nil, errors.New("review cannot be updated in current status")
	}

	// Update fields
	if req.Title != "" {
		review.Title = req.Title
	}
	if req.Content != "" {
		review.Content = req.Content
	}
	if req.Rating != nil {
		review.Rating = *req.Rating
	}
	if req.IsAnonymous != nil {
		review.IsAnonymous = *req.IsAnonymous
	}

	// Validate updated review
	if err := s.ValidateReview(review); err != nil {
		logger.Errorf("Review validation failed: %v", err)
		return nil, err
	}

	// Update review images if provided
	if len(req.ImageURLs) > 0 {
		// Delete existing images
		if err := s.reviewRepo.DeleteReviewImagesByReview(review.ID); err != nil {
			logger.Warnf("Failed to delete existing review images: %v", err)
		}

		// Create new images
		for i, imageURL := range req.ImageURLs {
			altText := ""
			if i < len(req.AltTexts) {
				altText = req.AltTexts[i]
			}

			image := &model.ReviewImage{
				ReviewID:  review.ID,
				ImageURL:  imageURL,
				ImagePath: imageURL,
				AltText:   altText,
				SortOrder: i,
			}

			if err := s.reviewRepo.CreateReviewImage(image); err != nil {
				logger.Warnf("Failed to create review image: %v", err)
			}
		}
	}

	if err := s.reviewRepo.UpdateReview(review); err != nil {
		logger.Errorf("Error updating review %d: %v", id, err)
		return nil, fmt.Errorf("failed to update review")
	}

	// Get updated review with relations
	updatedReview, err := s.reviewRepo.GetReviewByID(review.ID)
	if err != nil {
		logger.Errorf("Error getting updated review: %v", err)
		return nil, fmt.Errorf("failed to retrieve updated review")
	}

	return s.toReviewResponse(updatedReview), nil
}

// DeleteReview deletes a review
func (s *reviewService) DeleteReview(id uint, userID uint) error {
	review, err := s.reviewRepo.GetReviewByID(id)
	if err != nil {
		logger.Errorf("Error getting review by ID %d for deletion: %v", id, err)
		return fmt.Errorf("failed to retrieve review")
	}
	if review == nil {
		return errors.New("review not found")
	}

	// Check if user can delete this review
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		logger.Errorf("Error getting user by ID %d: %v", userID, err)
		return fmt.Errorf("failed to retrieve user")
	}
	if user == nil {
		return errors.New("user not found")
	}

	if review.UserID != userID && user.Role != "admin" && user.Role != "super_admin" {
		return errors.New("access denied: you can only delete your own reviews")
	}

	// Check if review can be deleted
	if !review.CanBeDeleted() {
		return errors.New("review cannot be deleted in current status")
	}

	if err := s.reviewRepo.DeleteReview(id); err != nil {
		logger.Errorf("Error deleting review %d: %v", id, err)
		return fmt.Errorf("failed to delete review")
	}

	return nil
}

// Review Management

// GetReviewsByStatus retrieves reviews by status
func (s *reviewService) GetReviewsByStatus(status model.ReviewStatus, page, limit int) ([]model.ReviewResponse, int64, error) {
	reviews, total, err := s.reviewRepo.GetReviewsByStatus(status, page, limit)
	if err != nil {
		logger.Errorf("Error getting reviews by status: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve reviews by status")
	}

	var responses []model.ReviewResponse
	for _, review := range reviews {
		responses = append(responses, *s.toReviewResponse(&review))
	}
	return responses, total, nil
}

// GetReviewsByType retrieves reviews by type
func (s *reviewService) GetReviewsByType(reviewType model.ReviewType, page, limit int) ([]model.ReviewResponse, int64, error) {
	reviews, total, err := s.reviewRepo.GetReviewsByType(reviewType, page, limit)
	if err != nil {
		logger.Errorf("Error getting reviews by type: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve reviews by type")
	}

	var responses []model.ReviewResponse
	for _, review := range reviews {
		responses = append(responses, *s.toReviewResponse(&review))
	}
	return responses, total, nil
}

// GetVerifiedReviews retrieves verified reviews for a product
func (s *reviewService) GetVerifiedReviews(productID uint, page, limit int) ([]model.ReviewResponse, int64, error) {
	reviews, total, err := s.reviewRepo.GetVerifiedReviews(productID, page, limit)
	if err != nil {
		logger.Errorf("Error getting verified reviews: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve verified reviews")
	}

	var responses []model.ReviewResponse
	for _, review := range reviews {
		responses = append(responses, *s.toReviewResponse(&review))
	}
	return responses, total, nil
}

// GetRecentReviews retrieves recent reviews
func (s *reviewService) GetRecentReviews(limit int) ([]model.ReviewResponse, error) {
	reviews, err := s.reviewRepo.GetRecentReviews(limit)
	if err != nil {
		logger.Errorf("Error getting recent reviews: %v", err)
		return nil, fmt.Errorf("failed to retrieve recent reviews")
	}

	var responses []model.ReviewResponse
	for _, review := range reviews {
		responses = append(responses, *s.toReviewResponse(&review))
	}
	return responses, nil
}

// SearchReviews performs full-text search on reviews
func (s *reviewService) SearchReviews(query string, page, limit int) ([]model.ReviewResponse, int64, error) {
	reviews, total, err := s.reviewRepo.SearchReviews(query, page, limit)
	if err != nil {
		logger.Errorf("Error searching reviews: %v", err)
		return nil, 0, fmt.Errorf("failed to search reviews")
	}

	var responses []model.ReviewResponse
	for _, review := range reviews {
		responses = append(responses, *s.toReviewResponse(&review))
	}
	return responses, total, nil
}

// Rating & Statistics

// GetAverageRating calculates average rating for a product
func (s *reviewService) GetAverageRating(productID uint) (float64, error) {
	avg, err := s.reviewRepo.GetAverageRating(productID)
	if err != nil {
		logger.Errorf("Error getting average rating: %v", err)
		return 0, fmt.Errorf("failed to retrieve average rating")
	}
	return avg, nil
}

// GetRatingDistribution gets rating distribution for a product
func (s *reviewService) GetRatingDistribution(productID uint) (map[int]int64, error) {
	distribution, err := s.reviewRepo.GetRatingDistribution(productID)
	if err != nil {
		logger.Errorf("Error getting rating distribution: %v", err)
		return nil, fmt.Errorf("failed to retrieve rating distribution")
	}
	return distribution, nil
}

// GetReviewStats retrieves review statistics
func (s *reviewService) GetReviewStats() (*model.ReviewStatsResponse, error) {
	stats, err := s.reviewRepo.GetReviewStats()
	if err != nil {
		logger.Errorf("Error getting review statistics: %v", err)
		return nil, fmt.Errorf("failed to retrieve review statistics")
	}
	return stats, nil
}

// GetProductReviewStats retrieves product-specific review statistics
func (s *reviewService) GetProductReviewStats(productID uint) (*model.ProductReviewStatsResponse, error) {
	stats, err := s.reviewRepo.GetProductReviewStats(productID)
	if err != nil {
		logger.Errorf("Error getting product review statistics: %v", err)
		return nil, fmt.Errorf("failed to retrieve product review statistics")
	}
	return stats, nil
}

// GetUserReviewStats retrieves user-specific review statistics
func (s *reviewService) GetUserReviewStats(userID uint) (map[string]interface{}, error) {
	stats, err := s.reviewRepo.GetUserReviewStats(userID)
	if err != nil {
		logger.Errorf("Error getting user review statistics: %v", err)
		return nil, fmt.Errorf("failed to retrieve user review statistics")
	}
	return stats, nil
}

// Helpful Votes

// CreateHelpfulVote creates a new helpful vote
func (s *reviewService) CreateHelpfulVote(reviewID uint, userID uint, isHelpful bool) (*model.ReviewHelpfulVoteResponse, error) {
	// Check if vote already exists
	existingVote, err := s.reviewRepo.GetHelpfulVoteByUserAndReview(userID, reviewID)
	if err != nil {
		logger.Errorf("Error checking existing vote: %v", err)
		return nil, fmt.Errorf("failed to check existing vote")
	}
	if existingVote != nil {
		return nil, errors.New("you have already voted on this review")
	}

	// Create vote
	vote := &model.ReviewHelpfulVote{
		ReviewID:  reviewID,
		UserID:    userID,
		IsHelpful: isHelpful,
	}

	if err := s.reviewRepo.CreateHelpfulVote(vote); err != nil {
		logger.Errorf("Error creating helpful vote: %v", err)
		return nil, fmt.Errorf("failed to create helpful vote")
	}

	// Update review helpful counts
	if err := s.UpdateReviewHelpfulCounts(reviewID); err != nil {
		logger.Warnf("Failed to update review helpful counts: %v", err)
	}

	// Get created vote with relations
	createdVote, err := s.reviewRepo.GetHelpfulVoteByUserAndReview(userID, reviewID)
	if err != nil {
		logger.Errorf("Error getting created vote: %v", err)
		return nil, fmt.Errorf("failed to retrieve created vote")
	}

	return s.toHelpfulVoteResponse(createdVote), nil
}

// UpdateHelpfulVote updates an existing helpful vote
func (s *reviewService) UpdateHelpfulVote(reviewID uint, userID uint, isHelpful bool) (*model.ReviewHelpfulVoteResponse, error) {
	// Get existing vote
	vote, err := s.reviewRepo.GetHelpfulVoteByUserAndReview(userID, reviewID)
	if err != nil {
		logger.Errorf("Error getting existing vote: %v", err)
		return nil, fmt.Errorf("failed to retrieve existing vote")
	}
	if vote == nil {
		return nil, errors.New("vote not found")
	}

	// Update vote
	vote.IsHelpful = isHelpful
	if err := s.reviewRepo.UpdateHelpfulVote(vote); err != nil {
		logger.Errorf("Error updating helpful vote: %v", err)
		return nil, fmt.Errorf("failed to update helpful vote")
	}

	// Update review helpful counts
	if err := s.UpdateReviewHelpfulCounts(reviewID); err != nil {
		logger.Warnf("Failed to update review helpful counts: %v", err)
	}

	return s.toHelpfulVoteResponse(vote), nil
}

// DeleteHelpfulVote deletes a helpful vote
func (s *reviewService) DeleteHelpfulVote(reviewID uint, userID uint) error {
	// Get existing vote
	vote, err := s.reviewRepo.GetHelpfulVoteByUserAndReview(userID, reviewID)
	if err != nil {
		logger.Errorf("Error getting existing vote: %v", err)
		return fmt.Errorf("failed to retrieve existing vote")
	}
	if vote == nil {
		return errors.New("vote not found")
	}

	if err := s.reviewRepo.DeleteHelpfulVote(vote.ID); err != nil {
		logger.Errorf("Error deleting helpful vote: %v", err)
		return fmt.Errorf("failed to delete helpful vote")
	}

	// Update review helpful counts
	if err := s.UpdateReviewHelpfulCounts(reviewID); err != nil {
		logger.Warnf("Failed to update review helpful counts: %v", err)
	}

	return nil
}

// GetHelpfulVotesByReview retrieves helpful votes for a review
func (s *reviewService) GetHelpfulVotesByReview(reviewID uint) ([]model.ReviewHelpfulVoteResponse, error) {
	votes, err := s.reviewRepo.GetHelpfulVotesByReview(reviewID)
	if err != nil {
		logger.Errorf("Error getting helpful votes: %v", err)
		return nil, fmt.Errorf("failed to retrieve helpful votes")
	}

	var responses []model.ReviewHelpfulVoteResponse
	for _, vote := range votes {
		responses = append(responses, *s.toHelpfulVoteResponse(&vote))
	}
	return responses, nil
}

// Review Images

// CreateReviewImage creates a new review image
func (s *reviewService) CreateReviewImage(reviewID uint, imageURL, imagePath, altText string, sortOrder int) (*model.ReviewImageResponse, error) {
	image := &model.ReviewImage{
		ReviewID:  reviewID,
		ImageURL:  imageURL,
		ImagePath: imagePath,
		AltText:   altText,
		SortOrder: sortOrder,
	}

	if err := s.reviewRepo.CreateReviewImage(image); err != nil {
		logger.Errorf("Error creating review image: %v", err)
		return nil, fmt.Errorf("failed to create review image")
	}

	return s.toReviewImageResponse(image), nil
}

// GetReviewImagesByReview retrieves images for a review
func (s *reviewService) GetReviewImagesByReview(reviewID uint) ([]model.ReviewImageResponse, error) {
	images, err := s.reviewRepo.GetReviewImagesByReview(reviewID)
	if err != nil {
		logger.Errorf("Error getting review images: %v", err)
		return nil, fmt.Errorf("failed to retrieve review images")
	}

	var responses []model.ReviewImageResponse
	for _, image := range images {
		responses = append(responses, *s.toReviewImageResponse(&image))
	}
	return responses, nil
}

// UpdateReviewImage updates an existing review image
func (s *reviewService) UpdateReviewImage(imageID uint, imageURL, imagePath, altText string, sortOrder int) (*model.ReviewImageResponse, error) {
	// This would require getting the image first, but for simplicity, we'll create a new one
	// In a real implementation, you'd want to get the existing image and update it
	return nil, errors.New("not implemented")
}

// DeleteReviewImage deletes a review image
func (s *reviewService) DeleteReviewImage(imageID uint) error {
	if err := s.reviewRepo.DeleteReviewImage(imageID); err != nil {
		logger.Errorf("Error deleting review image: %v", err)
		return fmt.Errorf("failed to delete review image")
	}
	return nil
}

// Moderation

// ModerateReview moderates a review
func (s *reviewService) ModerateReview(reviewID uint, req *model.ReviewModerationRequest, moderatorID uint) (*model.ReviewResponse, error) {
	// Check if review exists
	review, err := s.reviewRepo.GetReviewByID(reviewID)
	if err != nil {
		logger.Errorf("Error getting review by ID %d: %v", reviewID, err)
		return nil, fmt.Errorf("failed to retrieve review")
	}
	if review == nil {
		return nil, errors.New("review not found")
	}

	// Moderate review
	if err := s.reviewRepo.ModerateReview(reviewID, req.Status, moderatorID, req.ModerationNote); err != nil {
		logger.Errorf("Error moderating review: %v", err)
		return nil, fmt.Errorf("failed to moderate review")
	}

	// Get updated review
	updatedReview, err := s.reviewRepo.GetReviewByID(reviewID)
	if err != nil {
		logger.Errorf("Error getting updated review: %v", err)
		return nil, fmt.Errorf("failed to retrieve updated review")
	}

	return s.toReviewResponse(updatedReview), nil
}

// GetPendingReviews retrieves pending reviews
func (s *reviewService) GetPendingReviews(page, limit int) ([]model.ReviewResponse, int64, error) {
	reviews, total, err := s.reviewRepo.GetPendingReviews(page, limit)
	if err != nil {
		logger.Errorf("Error getting pending reviews: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve pending reviews")
	}

	var responses []model.ReviewResponse
	for _, review := range reviews {
		responses = append(responses, *s.toReviewResponse(&review))
	}
	return responses, total, nil
}

// GetModeratedReviews retrieves reviews moderated by a specific moderator
func (s *reviewService) GetModeratedReviews(moderatorID uint, page, limit int) ([]model.ReviewResponse, int64, error) {
	reviews, total, err := s.reviewRepo.GetModeratedReviews(moderatorID, page, limit)
	if err != nil {
		logger.Errorf("Error getting moderated reviews: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve moderated reviews")
	}

	var responses []model.ReviewResponse
	for _, review := range reviews {
		responses = append(responses, *s.toReviewResponse(&review))
	}
	return responses, total, nil
}

// Utility methods

// ValidateReview validates review data
func (s *reviewService) ValidateReview(review *model.Review) error {
	return review.ValidateReview()
}

// CanUserReviewProduct checks if user can review a product
func (s *reviewService) CanUserReviewProduct(userID, productID uint) (bool, error) {
	// Check if user has already reviewed this product
	reviews, _, err := s.reviewRepo.GetReviewsByUser(userID, 1, 1, map[string]interface{}{
		"product_id": productID,
	})
	if err != nil {
		return false, err
	}

	// If user has already reviewed this product, they cannot review again
	if len(reviews) > 0 {
		return false, nil
	}

	// Additional checks can be added here:
	// - Check if user has purchased the product
	// - Check if product is reviewable
	// - Check if user account is in good standing

	return true, nil
}

// UpdateReviewHelpfulCounts updates helpful counts for a review
func (s *reviewService) UpdateReviewHelpfulCounts(reviewID uint) error {
	// Get all votes for this review
	votes, err := s.reviewRepo.GetHelpfulVotesByReview(reviewID)
	if err != nil {
		return err
	}

	// Count helpful and not helpful votes
	helpfulCount := int64(0)
	notHelpfulCount := int64(0)

	for _, vote := range votes {
		if vote.IsHelpful {
			helpfulCount++
		} else {
			notHelpfulCount++
		}
	}

	// Update review counts
	// This would require a method to update just the counts
	// For now, we'll get the review and update it
	review, err := s.reviewRepo.GetReviewByID(reviewID)
	if err != nil {
		return err
	}
	if review == nil {
		return errors.New("review not found")
	}

	review.IsHelpful = int(helpfulCount)
	review.IsNotHelpful = int(notHelpfulCount)

	return s.reviewRepo.UpdateReview(review)
}

// Helper methods

// toReviewResponse converts Review to ReviewResponse
func (s *reviewService) toReviewResponse(review *model.Review) *model.ReviewResponse {
	return review.ToResponse()
}

// toReviewImageResponse converts ReviewImage to ReviewImageResponse
func (s *reviewService) toReviewImageResponse(image *model.ReviewImage) *model.ReviewImageResponse {
	return &model.ReviewImageResponse{
		ID:        image.ID,
		ReviewID:  image.ReviewID,
		ImageURL:  image.ImageURL,
		ImagePath: image.ImagePath,
		AltText:   image.AltText,
		SortOrder: image.SortOrder,
		CreatedAt: image.CreatedAt,
		UpdatedAt: image.UpdatedAt,
	}
}

// toHelpfulVoteResponse converts ReviewHelpfulVote to ReviewHelpfulVoteResponse
func (s *reviewService) toHelpfulVoteResponse(vote *model.ReviewHelpfulVote) *model.ReviewHelpfulVoteResponse {
	response := &model.ReviewHelpfulVoteResponse{
		ID:        vote.ID,
		ReviewID:  vote.ReviewID,
		UserID:    vote.UserID,
		IsHelpful: vote.IsHelpful,
		CreatedAt: vote.CreatedAt,
		UpdatedAt: vote.UpdatedAt,
	}

	if vote.User != nil {
		response.User = vote.User
	}

	return response
}
