package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// ReviewStatus defines the status of a review
type ReviewStatus string

const (
	ReviewStatusPending  ReviewStatus = "pending"  // Chờ duyệt
	ReviewStatusApproved ReviewStatus = "approved" // Đã duyệt
	ReviewStatusRejected ReviewStatus = "rejected" // Từ chối
	ReviewStatusHidden   ReviewStatus = "hidden"   // Ẩn
)

// ReviewType defines the type of review
type ReviewType string

const (
	ReviewTypeProduct ReviewType = "product" // Đánh giá sản phẩm
	ReviewTypeService ReviewType = "service" // Đánh giá dịch vụ
	ReviewTypeOrder   ReviewType = "order"   // Đánh giá đơn hàng
	ReviewTypeOverall ReviewType = "overall" // Đánh giá tổng thể
)

// Review represents a product review
type Review struct {
	ID        uint     `json:"id" gorm:"primaryKey"`
	UserID    uint     `json:"user_id" gorm:"not null;index"`
	User      *User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	ProductID uint     `json:"product_id" gorm:"not null;index"`
	Product   *Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	OrderID   *uint    `json:"order_id" gorm:"index"`
	Order     *Order   `json:"order,omitempty" gorm:"foreignKey:OrderID"`

	// Review Content
	Title   string       `json:"title" gorm:"size:255;not null"`
	Content string       `json:"content" gorm:"type:text;not null"`
	Rating  int          `json:"rating" gorm:"not null;check:rating >= 1 AND rating <= 5"`
	Status  ReviewStatus `json:"status" gorm:"size:20;default:'pending';index"`
	Type    ReviewType   `json:"type" gorm:"size:20;default:'product';index"`

	// Review Details
	IsVerified   bool `json:"is_verified" gorm:"default:false"`  // Đánh giá đã xác thực
	IsHelpful    int  `json:"is_helpful" gorm:"default:0"`       // Số lượt hữu ích
	IsNotHelpful int  `json:"is_not_helpful" gorm:"default:0"`   // Số lượt không hữu ích
	IsAnonymous  bool `json:"is_anonymous" gorm:"default:false"` // Đánh giá ẩn danh

	// Moderation
	ModeratedBy    *uint      `json:"moderated_by" gorm:"index"`
	Moderator      *User      `json:"moderator,omitempty" gorm:"foreignKey:ModeratedBy"`
	ModeratedAt    *time.Time `json:"moderated_at"`
	ModerationNote string     `json:"moderation_note" gorm:"type:text"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relations
	Images       []ReviewImage       `json:"images,omitempty" gorm:"foreignKey:ReviewID"`
	HelpfulVotes []ReviewHelpfulVote `json:"helpful_votes,omitempty" gorm:"foreignKey:ReviewID"`
}

// ReviewImage represents images attached to a review
type ReviewImage struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	ReviewID  uint           `json:"review_id" gorm:"not null;index"`
	Review    *Review        `json:"review,omitempty" gorm:"foreignKey:ReviewID"`
	ImageURL  string         `json:"image_url" gorm:"size:500;not null"`
	ImagePath string         `json:"image_path" gorm:"size:500;not null"`
	AltText   string         `json:"alt_text" gorm:"size:255"`
	SortOrder int            `json:"sort_order" gorm:"default:0"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// ReviewHelpfulVote represents helpful votes for reviews
type ReviewHelpfulVote struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	ReviewID  uint           `json:"review_id" gorm:"not null;index"`
	Review    *Review        `json:"review,omitempty" gorm:"foreignKey:ReviewID"`
	UserID    uint           `json:"user_id" gorm:"not null;index"`
	User      *User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	IsHelpful bool           `json:"is_helpful" gorm:"not null"` // true = helpful, false = not helpful
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Unique constraint to prevent duplicate votes
	// This will be handled in migration
}

// Request/Response structs

// ReviewCreateRequest represents the request body for creating a review
type ReviewCreateRequest struct {
	ProductID   uint       `json:"product_id" binding:"required"`
	OrderID     *uint      `json:"order_id"` // Optional: if reviewing from order
	Title       string     `json:"title" binding:"required,min=5,max=255"`
	Content     string     `json:"content" binding:"required,min=10,max=2000"`
	Rating      int        `json:"rating" binding:"required,min=1,max=5"`
	Type        ReviewType `json:"type" binding:"omitempty,oneof=product service order overall"`
	IsAnonymous bool       `json:"is_anonymous"`
	ImageURLs   []string   `json:"image_urls" binding:"omitempty,dive,url"`
	AltTexts    []string   `json:"alt_texts" binding:"omitempty,dive,max=255"`
}

// ReviewUpdateRequest represents the request body for updating a review
type ReviewUpdateRequest struct {
	Title       string   `json:"title" binding:"omitempty,min=5,max=255"`
	Content     string   `json:"content" binding:"omitempty,min=10,max=2000"`
	Rating      *int     `json:"rating" binding:"omitempty,min=1,max=5"`
	IsAnonymous *bool    `json:"is_anonymous"`
	ImageURLs   []string `json:"image_urls" binding:"omitempty,dive,url"`
	AltTexts    []string `json:"alt_texts" binding:"omitempty,dive,max=255"`
}

// ReviewModerationRequest represents the request body for moderating a review
type ReviewModerationRequest struct {
	Status         ReviewStatus `json:"status" binding:"required,oneof=pending approved rejected hidden"`
	ModerationNote string       `json:"moderation_note" binding:"omitempty,max=1000"`
}

// ReviewResponse represents the response body for a review
type ReviewResponse struct {
	ID             uint                        `json:"id"`
	UserID         uint                        `json:"user_id"`
	User           *User                       `json:"user,omitempty"`
	ProductID      uint                        `json:"product_id"`
	Product        *ProductResponse            `json:"product,omitempty"`
	OrderID        *uint                       `json:"order_id"`
	Order          *Order                      `json:"order,omitempty"`
	Title          string                      `json:"title"`
	Content        string                      `json:"content"`
	Rating         int                         `json:"rating"`
	Status         ReviewStatus                `json:"status"`
	Type           ReviewType                  `json:"type"`
	IsVerified     bool                        `json:"is_verified"`
	IsHelpful      int                         `json:"is_helpful"`
	IsNotHelpful   int                         `json:"is_not_helpful"`
	IsAnonymous    bool                        `json:"is_anonymous"`
	ModeratedBy    *uint                       `json:"moderated_by"`
	Moderator      *User                       `json:"moderator,omitempty"`
	ModeratedAt    *time.Time                  `json:"moderated_at"`
	ModerationNote string                      `json:"moderation_note"`
	CreatedAt      time.Time                   `json:"created_at"`
	UpdatedAt      time.Time                   `json:"updated_at"`
	Images         []ReviewImageResponse       `json:"images,omitempty"`
	HelpfulVotes   []ReviewHelpfulVoteResponse `json:"helpful_votes,omitempty"`
}

// ReviewImageResponse represents the response body for a review image
type ReviewImageResponse struct {
	ID        uint      `json:"id"`
	ReviewID  uint      `json:"review_id"`
	ImageURL  string    `json:"image_url"`
	ImagePath string    `json:"image_path"`
	AltText   string    `json:"alt_text"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ReviewHelpfulVoteResponse represents the response body for a helpful vote
type ReviewHelpfulVoteResponse struct {
	ID        uint      `json:"id"`
	ReviewID  uint      `json:"review_id"`
	UserID    uint      `json:"user_id"`
	User      *User     `json:"user,omitempty"`
	IsHelpful bool      `json:"is_helpful"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ReviewStatsResponse represents review statistics
type ReviewStatsResponse struct {
	TotalReviews       int64                `json:"total_reviews"`
	ApprovedReviews    int64                `json:"approved_reviews"`
	PendingReviews     int64                `json:"pending_reviews"`
	RejectedReviews    int64                `json:"rejected_reviews"`
	AverageRating      float64              `json:"average_rating"`
	RatingDistribution map[int]int64        `json:"rating_distribution"`
	ReviewsByType      map[ReviewType]int64 `json:"reviews_by_type"`
	RecentReviews      []ReviewResponse     `json:"recent_reviews"`
}

// ProductReviewStatsResponse represents product-specific review statistics
type ProductReviewStatsResponse struct {
	ProductID          uint             `json:"product_id"`
	TotalReviews       int64            `json:"total_reviews"`
	AverageRating      float64          `json:"average_rating"`
	RatingDistribution map[int]int64    `json:"rating_distribution"`
	VerifiedReviews    int64            `json:"verified_reviews"`
	RecentReviews      []ReviewResponse `json:"recent_reviews"`
}

// Helper methods

// IsApproved checks if review is approved
func (r *Review) IsApproved() bool {
	return r.Status == ReviewStatusApproved
}

// IsPending checks if review is pending
func (r *Review) IsPending() bool {
	return r.Status == ReviewStatusPending
}

// IsRejected checks if review is rejected
func (r *Review) IsRejected() bool {
	return r.Status == ReviewStatusRejected
}

// IsHidden checks if review is hidden
func (r *Review) IsHidden() bool {
	return r.Status == ReviewStatusHidden
}

// GetHelpfulScore calculates helpful score
func (r *Review) GetHelpfulScore() float64 {
	total := r.IsHelpful + r.IsNotHelpful
	if total == 0 {
		return 0
	}
	return float64(r.IsHelpful) / float64(total) * 100
}

// CanBeEdited checks if review can be edited
func (r *Review) CanBeEdited() bool {
	return r.Status == ReviewStatusPending || r.Status == ReviewStatusApproved
}

// CanBeDeleted checks if review can be deleted
func (r *Review) CanBeDeleted() bool {
	return r.Status == ReviewStatusPending
}

// ToResponse converts Review to ReviewResponse
func (r *Review) ToResponse() *ReviewResponse {
	response := &ReviewResponse{
		ID:             r.ID,
		UserID:         r.UserID,
		ProductID:      r.ProductID,
		OrderID:        r.OrderID,
		Title:          r.Title,
		Content:        r.Content,
		Rating:         r.Rating,
		Status:         r.Status,
		Type:           r.Type,
		IsVerified:     r.IsVerified,
		IsHelpful:      r.IsHelpful,
		IsNotHelpful:   r.IsNotHelpful,
		IsAnonymous:    r.IsAnonymous,
		ModeratedBy:    r.ModeratedBy,
		ModeratedAt:    r.ModeratedAt,
		ModerationNote: r.ModerationNote,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}

	// Add user information if not anonymous
	if !r.IsAnonymous && r.User != nil {
		response.User = r.User
	}

	// Add product information
	if r.Product != nil {
		productResp := r.Product.ToResponse()
		response.Product = &productResp
	}

	// Add order information
	if r.Order != nil {
		response.Order = r.Order
	}

	// Add moderator information
	if r.Moderator != nil {
		response.Moderator = r.Moderator
	}

	// Add images
	for _, img := range r.Images {
		response.Images = append(response.Images, ReviewImageResponse{
			ID:        img.ID,
			ReviewID:  img.ReviewID,
			ImageURL:  img.ImageURL,
			ImagePath: img.ImagePath,
			AltText:   img.AltText,
			SortOrder: img.SortOrder,
			CreatedAt: img.CreatedAt,
			UpdatedAt: img.UpdatedAt,
		})
	}

	// Add helpful votes
	for _, vote := range r.HelpfulVotes {
		voteResponse := ReviewHelpfulVoteResponse{
			ID:        vote.ID,
			ReviewID:  vote.ReviewID,
			UserID:    vote.UserID,
			IsHelpful: vote.IsHelpful,
			CreatedAt: vote.CreatedAt,
			UpdatedAt: vote.UpdatedAt,
		}
		if vote.User != nil {
			voteResponse.User = vote.User
		}
		response.HelpfulVotes = append(response.HelpfulVotes, voteResponse)
	}

	return response
}

// ValidateReview validates review data
func (r *Review) ValidateReview() error {
	if r.Title == "" {
		return errors.New("title is required")
	}
	if len(r.Title) < 5 || len(r.Title) > 255 {
		return errors.New("title must be between 5 and 255 characters")
	}
	if r.Content == "" {
		return errors.New("content is required")
	}
	if len(r.Content) < 10 || len(r.Content) > 2000 {
		return errors.New("content must be between 10 and 2000 characters")
	}
	if r.Rating < 1 || r.Rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}
	if r.UserID == 0 {
		return errors.New("user_id is required")
	}
	if r.ProductID == 0 {
		return errors.New("product_id is required")
	}
	return nil
}

// GetStatusDisplayName returns display name for review status
func (r *Review) GetStatusDisplayName() string {
	statusMap := map[ReviewStatus]string{
		ReviewStatusPending:  "Chờ duyệt",
		ReviewStatusApproved: "Đã duyệt",
		ReviewStatusRejected: "Từ chối",
		ReviewStatusHidden:   "Ẩn",
	}
	return statusMap[r.Status]
}

// GetTypeDisplayName returns display name for review type
func (r *Review) GetTypeDisplayName() string {
	typeMap := map[ReviewType]string{
		ReviewTypeProduct: "Sản phẩm",
		ReviewTypeService: "Dịch vụ",
		ReviewTypeOrder:   "Đơn hàng",
		ReviewTypeOverall: "Tổng thể",
	}
	return typeMap[r.Type]
}
