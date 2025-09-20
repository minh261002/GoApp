package repository

import (
	"fmt"
	"go_app/internal/model"
	"go_app/pkg/database"
	"time"

	"gorm.io/gorm"
)

// ReviewRepository defines methods for interacting with review data
type ReviewRepository interface {
	// Basic CRUD
	CreateReview(review *model.Review) error
	GetReviewByID(id uint) (*model.Review, error)
	GetAllReviews(page, limit int, filters map[string]interface{}) ([]model.Review, int64, error)
	GetReviewsByProduct(productID uint, page, limit int, filters map[string]interface{}) ([]model.Review, int64, error)
	GetReviewsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.Review, int64, error)
	UpdateReview(review *model.Review) error
	DeleteReview(id uint) error

	// Review Management
	GetReviewsByStatus(status model.ReviewStatus, page, limit int) ([]model.Review, int64, error)
	GetReviewsByType(reviewType model.ReviewType, page, limit int) ([]model.Review, int64, error)
	GetVerifiedReviews(productID uint, page, limit int) ([]model.Review, int64, error)
	GetRecentReviews(limit int) ([]model.Review, error)
	SearchReviews(query string, page, limit int) ([]model.Review, int64, error)

	// Rating & Statistics
	GetAverageRating(productID uint) (float64, error)
	GetRatingDistribution(productID uint) (map[int]int64, error)
	GetReviewStats() (*model.ReviewStatsResponse, error)
	GetProductReviewStats(productID uint) (*model.ProductReviewStatsResponse, error)
	GetUserReviewStats(userID uint) (map[string]interface{}, error)

	// Helpful Votes
	CreateHelpfulVote(vote *model.ReviewHelpfulVote) error
	GetHelpfulVoteByUserAndReview(userID, reviewID uint) (*model.ReviewHelpfulVote, error)
	UpdateHelpfulVote(vote *model.ReviewHelpfulVote) error
	DeleteHelpfulVote(id uint) error
	GetHelpfulVotesByReview(reviewID uint) ([]model.ReviewHelpfulVote, error)

	// Review Images
	CreateReviewImage(image *model.ReviewImage) error
	GetReviewImagesByReview(reviewID uint) ([]model.ReviewImage, error)
	UpdateReviewImage(image *model.ReviewImage) error
	DeleteReviewImage(id uint) error
	DeleteReviewImagesByReview(reviewID uint) error

	// Moderation
	ModerateReview(reviewID uint, status model.ReviewStatus, moderatorID uint, note string) error
	GetPendingReviews(page, limit int) ([]model.Review, int64, error)
	GetModeratedReviews(moderatorID uint, page, limit int) ([]model.Review, int64, error)
}

// reviewRepository implements ReviewRepository
type reviewRepository struct {
	db *gorm.DB
}

// NewReviewRepository creates a new ReviewRepository
func NewReviewRepository() ReviewRepository {
	return &reviewRepository{
		db: database.DB,
	}
}

// Basic CRUD

// CreateReview creates a new review
func (r *reviewRepository) CreateReview(review *model.Review) error {
	return r.db.Create(review).Error
}

// GetReviewByID retrieves a review by its ID
func (r *reviewRepository) GetReviewByID(id uint) (*model.Review, error) {
	var review model.Review
	if err := r.db.Preload("User").Preload("Product").Preload("Order").Preload("Moderator").
		Preload("Images").Preload("HelpfulVotes.User").First(&review, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &review, nil
}

// GetAllReviews retrieves all reviews with pagination and filters
func (r *reviewRepository) GetAllReviews(page, limit int, filters map[string]interface{}) ([]model.Review, int64, error) {
	var reviews []model.Review
	var total int64
	db := r.db.Model(&model.Review{})

	// Apply filters
	for key, value := range filters {
		switch key {
		case "user_id":
			db = db.Where("user_id = ?", value)
		case "product_id":
			db = db.Where("product_id = ?", value)
		case "order_id":
			db = db.Where("order_id = ?", value)
		case "status":
			db = db.Where("status = ?", value)
		case "type":
			db = db.Where("type = ?", value)
		case "rating":
			db = db.Where("rating = ?", value)
		case "is_verified":
			db = db.Where("is_verified = ?", value)
		case "is_anonymous":
			db = db.Where("is_anonymous = ?", value)
		case "moderated_by":
			db = db.Where("moderated_by = ?", value)
		case "search":
			searchTerm := fmt.Sprintf("%%%s%%", value.(string))
			db = db.Where("title LIKE ? OR content LIKE ?", searchTerm, searchTerm)
		case "date_from":
			db = db.Where("created_at >= ?", value)
		case "date_to":
			db = db.Where("created_at <= ?", value)
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
	db = db.Order("created_at DESC")

	if err := db.Preload("User").Preload("Product").Preload("Order").Preload("Moderator").
		Preload("Images").Preload("HelpfulVotes.User").Find(&reviews).Error; err != nil {
		return nil, 0, err
	}

	return reviews, total, nil
}

// GetReviewsByProduct retrieves reviews for a specific product
func (r *reviewRepository) GetReviewsByProduct(productID uint, page, limit int, filters map[string]interface{}) ([]model.Review, int64, error) {
	filters["product_id"] = productID
	return r.GetAllReviews(page, limit, filters)
}

// GetReviewsByUser retrieves reviews for a specific user
func (r *reviewRepository) GetReviewsByUser(userID uint, page, limit int, filters map[string]interface{}) ([]model.Review, int64, error) {
	filters["user_id"] = userID
	return r.GetAllReviews(page, limit, filters)
}

// UpdateReview updates an existing review
func (r *reviewRepository) UpdateReview(review *model.Review) error {
	return r.db.Save(review).Error
}

// DeleteReview soft deletes a review
func (r *reviewRepository) DeleteReview(id uint) error {
	return r.db.Delete(&model.Review{}, id).Error
}

// Review Management

// GetReviewsByStatus retrieves reviews by status
func (r *reviewRepository) GetReviewsByStatus(status model.ReviewStatus, page, limit int) ([]model.Review, int64, error) {
	var reviews []model.Review
	var total int64
	db := r.db.Model(&model.Review{}).Where("status = ?", status)

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
	db = db.Order("created_at DESC")

	if err := db.Preload("User").Preload("Product").Preload("Order").Preload("Moderator").
		Preload("Images").Preload("HelpfulVotes.User").Find(&reviews).Error; err != nil {
		return nil, 0, err
	}

	return reviews, total, nil
}

// GetReviewsByType retrieves reviews by type
func (r *reviewRepository) GetReviewsByType(reviewType model.ReviewType, page, limit int) ([]model.Review, int64, error) {
	var reviews []model.Review
	var total int64
	db := r.db.Model(&model.Review{}).Where("type = ?", reviewType)

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
	db = db.Order("created_at DESC")

	if err := db.Preload("User").Preload("Product").Preload("Order").Preload("Moderator").
		Preload("Images").Preload("HelpfulVotes.User").Find(&reviews).Error; err != nil {
		return nil, 0, err
	}

	return reviews, total, nil
}

// GetVerifiedReviews retrieves verified reviews for a product
func (r *reviewRepository) GetVerifiedReviews(productID uint, page, limit int) ([]model.Review, int64, error) {
	var reviews []model.Review
	var total int64
	db := r.db.Model(&model.Review{}).Where("product_id = ? AND is_verified = ? AND status = ?",
		productID, true, model.ReviewStatusApproved)

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
	db = db.Order("created_at DESC")

	if err := db.Preload("User").Preload("Product").Preload("Order").Preload("Moderator").
		Preload("Images").Preload("HelpfulVotes.User").Find(&reviews).Error; err != nil {
		return nil, 0, err
	}

	return reviews, total, nil
}

// GetRecentReviews retrieves recent reviews
func (r *reviewRepository) GetRecentReviews(limit int) ([]model.Review, error) {
	var reviews []model.Review
	err := r.db.Where("status = ?", model.ReviewStatusApproved).
		Preload("User").Preload("Product").Preload("Order").Preload("Moderator").
		Preload("Images").Preload("HelpfulVotes.User").
		Order("created_at DESC").
		Limit(limit).
		Find(&reviews).Error
	return reviews, err
}

// SearchReviews performs full-text search on reviews
func (r *reviewRepository) SearchReviews(query string, page, limit int) ([]model.Review, int64, error) {
	var reviews []model.Review
	var total int64

	// Use MATCH AGAINST for full-text search
	db := r.db.Model(&model.Review{}).
		Where("MATCH(title, content) AGAINST(? IN NATURAL LANGUAGE MODE) AND status = ?",
			query, model.ReviewStatusApproved)

	// Count total records
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting by relevance and date
	db = db.Order("MATCH(title, content) AGAINST(? IN NATURAL LANGUAGE MODE) DESC, created_at DESC")

	if err := db.Preload("User").Preload("Product").Preload("Order").Preload("Moderator").
		Preload("Images").Preload("HelpfulVotes.User").Find(&reviews).Error; err != nil {
		return nil, 0, err
	}

	return reviews, total, nil
}

// Rating & Statistics

// GetAverageRating calculates average rating for a product
func (r *reviewRepository) GetAverageRating(productID uint) (float64, error) {
	var avg float64
	err := r.db.Model(&model.Review{}).
		Where("product_id = ? AND status = ?", productID, model.ReviewStatusApproved).
		Select("AVG(rating)").
		Scan(&avg).Error
	return avg, err
}

// GetRatingDistribution gets rating distribution for a product
func (r *reviewRepository) GetRatingDistribution(productID uint) (map[int]int64, error) {
	distribution := make(map[int]int64)

	var results []struct {
		Rating int   `json:"rating"`
		Count  int64 `json:"count"`
	}

	err := r.db.Model(&model.Review{}).
		Select("rating, COUNT(*) as count").
		Where("product_id = ? AND status = ?", productID, model.ReviewStatusApproved).
		Group("rating").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	for _, result := range results {
		distribution[result.Rating] = result.Count
	}

	return distribution, nil
}

// GetReviewStats retrieves review statistics
func (r *reviewRepository) GetReviewStats() (*model.ReviewStatsResponse, error) {
	var stats model.ReviewStatsResponse
	var count int64

	// Total reviews
	r.db.Model(&model.Review{}).Count(&count)
	stats.TotalReviews = count

	// Approved reviews
	r.db.Model(&model.Review{}).Where("status = ?", model.ReviewStatusApproved).Count(&count)
	stats.ApprovedReviews = count

	// Pending reviews
	r.db.Model(&model.Review{}).Where("status = ?", model.ReviewStatusPending).Count(&count)
	stats.PendingReviews = count

	// Rejected reviews
	r.db.Model(&model.Review{}).Where("status = ?", model.ReviewStatusRejected).Count(&count)
	stats.RejectedReviews = count

	// Average rating
	var avg float64
	r.db.Model(&model.Review{}).
		Where("status = ?", model.ReviewStatusApproved).
		Select("AVG(rating)").
		Scan(&avg)
	stats.AverageRating = avg

	// Rating distribution
	stats.RatingDistribution = make(map[int]int64)
	var ratingStats []struct {
		Rating int   `json:"rating"`
		Count  int64 `json:"count"`
	}
	r.db.Model(&model.Review{}).
		Select("rating, COUNT(*) as count").
		Where("status = ?", model.ReviewStatusApproved).
		Group("rating").
		Scan(&ratingStats)

	for _, stat := range ratingStats {
		stats.RatingDistribution[stat.Rating] = stat.Count
	}

	// Reviews by type
	stats.ReviewsByType = make(map[model.ReviewType]int64)
	var typeStats []struct {
		Type  model.ReviewType `json:"type"`
		Count int64            `json:"count"`
	}
	r.db.Model(&model.Review{}).
		Select("type, COUNT(*) as count").
		Where("status = ?", model.ReviewStatusApproved).
		Group("type").
		Scan(&typeStats)

	for _, stat := range typeStats {
		stats.ReviewsByType[stat.Type] = stat.Count
	}

	// Recent reviews
	recentReviews, err := r.GetRecentReviews(10)
	if err != nil {
		return nil, err
	}

	for _, review := range recentReviews {
		stats.RecentReviews = append(stats.RecentReviews, *review.ToResponse())
	}

	return &stats, nil
}

// GetProductReviewStats retrieves product-specific review statistics
func (r *reviewRepository) GetProductReviewStats(productID uint) (*model.ProductReviewStatsResponse, error) {
	var stats model.ProductReviewStatsResponse
	stats.ProductID = productID

	var count int64

	// Total reviews for product
	r.db.Model(&model.Review{}).Where("product_id = ? AND status = ?", productID, model.ReviewStatusApproved).Count(&count)
	stats.TotalReviews = count

	// Average rating for product
	var avg float64
	r.db.Model(&model.Review{}).
		Where("product_id = ? AND status = ?", productID, model.ReviewStatusApproved).
		Select("AVG(rating)").
		Scan(&avg)
	stats.AverageRating = avg

	// Rating distribution for product
	stats.RatingDistribution, _ = r.GetRatingDistribution(productID)

	// Verified reviews for product
	r.db.Model(&model.Review{}).Where("product_id = ? AND is_verified = ? AND status = ?",
		productID, true, model.ReviewStatusApproved).Count(&count)
	stats.VerifiedReviews = count
	// Recent reviews for product
	recentReviews, _, err := r.GetReviewsByProduct(productID, 1, 10, map[string]interface{}{
		"status": model.ReviewStatusApproved,
	})
	if err != nil {
		return nil, err
	}

	for _, review := range recentReviews {
		stats.RecentReviews = append(stats.RecentReviews, *review.ToResponse())
	}

	return &stats, nil
}

// GetUserReviewStats retrieves user-specific review statistics
func (r *reviewRepository) GetUserReviewStats(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var count int64

	// Total reviews for user
	r.db.Model(&model.Review{}).Where("user_id = ?", userID).Count(&count)
	stats["total_reviews"] = count

	// Approved reviews for user
	r.db.Model(&model.Review{}).Where("user_id = ? AND status = ?", userID, model.ReviewStatusApproved).Count(&count)
	stats["approved_reviews"] = count

	// Pending reviews for user
	r.db.Model(&model.Review{}).Where("user_id = ? AND status = ?", userID, model.ReviewStatusPending).Count(&count)
	stats["pending_reviews"] = count

	// Rejected reviews for user
	r.db.Model(&model.Review{}).Where("user_id = ? AND status = ?", userID, model.ReviewStatusRejected).Count(&count)
	stats["rejected_reviews"] = count

	// Average rating given by user
	var avg float64
	r.db.Model(&model.Review{}).
		Where("user_id = ? AND status = ?", userID, model.ReviewStatusApproved).
		Select("AVG(rating)").
		Scan(&avg)
	stats["average_rating_given"] = avg

	// Total helpful votes received
	var helpfulVotes int64
	r.db.Model(&model.Review{}).
		Where("user_id = ? AND status = ?", userID, model.ReviewStatusApproved).
		Select("SUM(is_helpful)").
		Scan(&helpfulVotes)
	stats["total_helpful_votes"] = helpfulVotes

	// Last review created
	var lastReview model.Review
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").First(&lastReview).Error; err == nil {
		stats["last_review_created"] = lastReview.CreatedAt
	}

	return stats, nil
}

// Helpful Votes

// CreateHelpfulVote creates a new helpful vote
func (r *reviewRepository) CreateHelpfulVote(vote *model.ReviewHelpfulVote) error {
	return r.db.Create(vote).Error
}

// GetHelpfulVoteByUserAndReview retrieves a helpful vote by user and review
func (r *reviewRepository) GetHelpfulVoteByUserAndReview(userID, reviewID uint) (*model.ReviewHelpfulVote, error) {
	var vote model.ReviewHelpfulVote
	if err := r.db.Where("user_id = ? AND review_id = ?", userID, reviewID).
		Preload("User").Preload("Review").First(&vote).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &vote, nil
}

// UpdateHelpfulVote updates an existing helpful vote
func (r *reviewRepository) UpdateHelpfulVote(vote *model.ReviewHelpfulVote) error {
	return r.db.Save(vote).Error
}

// DeleteHelpfulVote deletes a helpful vote
func (r *reviewRepository) DeleteHelpfulVote(id uint) error {
	return r.db.Delete(&model.ReviewHelpfulVote{}, id).Error
}

// GetHelpfulVotesByReview retrieves helpful votes for a review
func (r *reviewRepository) GetHelpfulVotesByReview(reviewID uint) ([]model.ReviewHelpfulVote, error) {
	var votes []model.ReviewHelpfulVote
	err := r.db.Where("review_id = ?", reviewID).
		Preload("User").Preload("Review").
		Order("created_at DESC").
		Find(&votes).Error
	return votes, err
}

// Review Images

// CreateReviewImage creates a new review image
func (r *reviewRepository) CreateReviewImage(image *model.ReviewImage) error {
	return r.db.Create(image).Error
}

// GetReviewImagesByReview retrieves images for a review
func (r *reviewRepository) GetReviewImagesByReview(reviewID uint) ([]model.ReviewImage, error) {
	var images []model.ReviewImage
	err := r.db.Where("review_id = ?", reviewID).
		Order("sort_order ASC, created_at ASC").
		Find(&images).Error
	return images, err
}

// UpdateReviewImage updates an existing review image
func (r *reviewRepository) UpdateReviewImage(image *model.ReviewImage) error {
	return r.db.Save(image).Error
}

// DeleteReviewImage deletes a review image
func (r *reviewRepository) DeleteReviewImage(id uint) error {
	return r.db.Delete(&model.ReviewImage{}, id).Error
}

// DeleteReviewImagesByReview deletes all images for a review
func (r *reviewRepository) DeleteReviewImagesByReview(reviewID uint) error {
	return r.db.Where("review_id = ?", reviewID).Delete(&model.ReviewImage{}).Error
}

// Moderation

// ModerateReview moderates a review
func (r *reviewRepository) ModerateReview(reviewID uint, status model.ReviewStatus, moderatorID uint, note string) error {
	now := time.Now()
	return r.db.Model(&model.Review{}).
		Where("id = ?", reviewID).
		Updates(map[string]interface{}{
			"status":          status,
			"moderated_by":    moderatorID,
			"moderated_at":    &now,
			"moderation_note": note,
		}).Error
}

// GetPendingReviews retrieves pending reviews
func (r *reviewRepository) GetPendingReviews(page, limit int) ([]model.Review, int64, error) {
	return r.GetReviewsByStatus(model.ReviewStatusPending, page, limit)
}

// GetModeratedReviews retrieves reviews moderated by a specific moderator
func (r *reviewRepository) GetModeratedReviews(moderatorID uint, page, limit int) ([]model.Review, int64, error) {
	var reviews []model.Review
	var total int64
	db := r.db.Model(&model.Review{}).Where("moderated_by = ?", moderatorID)

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
	db = db.Order("moderated_at DESC")

	if err := db.Preload("User").Preload("Product").Preload("Order").Preload("Moderator").
		Preload("Images").Preload("HelpfulVotes.User").Find(&reviews).Error; err != nil {
		return nil, 0, err
	}

	return reviews, total, nil
}
