package repository

import (
	"fmt"
	"go_app/internal/model"
	"go_app/pkg/database"
	"time"

	"gorm.io/gorm"
)

// SearchRepository defines methods for search operations
type SearchRepository interface {
	// Search Operations
	SearchProducts(query string, filters map[string]interface{}, sortBy, sortOrder string, page, limit int) ([]model.SearchResult, int64, error)
	SearchCategories(query string, filters map[string]interface{}, sortBy, sortOrder string, page, limit int) ([]model.SearchResult, int64, error)
	SearchBrands(query string, filters map[string]interface{}, sortBy, sortOrder string, page, limit int) ([]model.SearchResult, int64, error)
	SearchUsers(query string, filters map[string]interface{}, sortBy, sortOrder string, page, limit int) ([]model.SearchResult, int64, error)
	SearchWishlists(query string, filters map[string]interface{}, sortBy, sortOrder string, page, limit int) ([]model.SearchResult, int64, error)
	SearchReviews(query string, filters map[string]interface{}, sortBy, sortOrder string, page, limit int) ([]model.SearchResult, int64, error)
	SearchOrders(query string, filters map[string]interface{}, sortBy, sortOrder string, page, limit int) ([]model.SearchResult, int64, error)
	SearchAll(query string, filters map[string]interface{}, sortBy, sortOrder string, page, limit int) ([]model.SearchResult, int64, error)

	// Search Query Management
	CreateSearchQuery(query *model.SearchQuery) error
	GetSearchQueryByID(id uint) (*model.SearchQuery, error)
	GetSearchQueriesByUser(userID uint, page, limit int) ([]model.SearchQuery, int64, error)
	UpdateSearchQuery(query *model.SearchQuery) error
	DeleteSearchQuery(id uint) error

	// Search History
	CreateSearchHistory(history *model.SearchHistory) error
	GetSearchHistoryByUser(userID uint, page, limit int) ([]model.SearchHistory, int64, error)
	DeleteSearchHistory(id uint) error
	ClearUserSearchHistory(userID uint) error

	// Search Suggestions
	GetSearchSuggestions(query string, searchType model.SearchType, limit int) ([]string, error)
	CreateSearchSuggestion(suggestion *model.SearchSuggestion) error
	UpdateSearchSuggestion(suggestion *model.SearchSuggestion) error
	DeleteSearchSuggestion(id uint) error

	// Search Filters
	GetSearchFilters(searchType model.SearchType) ([]model.SearchFilter, error)
	CreateSearchFilter(filter *model.SearchFilter) error
	UpdateSearchFilter(filter *model.SearchFilter) error
	DeleteSearchFilter(id uint) error

	// Search Index
	CreateSearchIndex(index *model.SearchIndex) error
	UpdateSearchIndex(index *model.SearchIndex) error
	DeleteSearchIndex(id uint) error
	DeleteSearchIndexByEntity(entityType string, entityID uint) error
	GetSearchIndexByEntity(entityType string, entityID uint) (*model.SearchIndex, error)

	// Search Analytics
	GetSearchStats(startDate, endDate *time.Time, searchType *model.SearchType) (*model.SearchStatsResponse, error)
	GetDailySearchStats(startDate, endDate *time.Time, searchType *model.SearchType) ([]model.DailySearchStats, error)
	GetTopQueries(limit int, searchType *model.SearchType) ([]model.QueryStats, error)
	GetZeroResultQueries(limit int, searchType *model.SearchType) ([]string, error)

	// Facets
	GetSearchFacets(searchType model.SearchType, filters map[string]interface{}) (map[string]model.Facet, error)
	UpdateSearchFacets(searchType model.SearchType, facets map[string]model.Facet) error
}

// searchRepository implements SearchRepository
type searchRepository struct {
	db *gorm.DB
}

// NewSearchRepository creates a new SearchRepository
func NewSearchRepository() SearchRepository {
	return &searchRepository{
		db: database.DB,
	}
}

// Search Operations

// SearchProducts searches for products
func (r *searchRepository) SearchProducts(query string, filters map[string]interface{}, sortBy, sortOrder string, page, limit int) ([]model.SearchResult, int64, error) {
	var results []model.SearchResult
	var total int64

	// Build base query
	db := r.db.Model(&model.Product{}).Where("deleted_at IS NULL")

	// Apply text search
	if query != "" {
		db = db.Where("MATCH(name, description, sku) AGAINST(? IN NATURAL LANGUAGE MODE)", query)
	}

	// Apply filters
	for key, value := range filters {
		switch key {
		case "brand_id":
			if ids, ok := value.([]uint); ok {
				db = db.Where("brand_id IN ?", ids)
			}
		case "category_id":
			if ids, ok := value.([]uint); ok {
				db = db.Where("category_id IN ?", ids)
			}
		case "status":
			db = db.Where("status = ?", value)
		case "price_min":
			db = db.Where("regular_price >= ?", value)
		case "price_max":
			db = db.Where("regular_price <= ?", value)
		case "in_stock":
			if value.(bool) {
				db = db.Where("stock_quantity > 0")
			}
		case "on_sale":
			if value.(bool) {
				db = db.Where("sale_price IS NOT NULL AND sale_price > 0")
			}
		case "rating_min":
			db = db.Where("average_rating >= ?", value)
		case "rating_max":
			db = db.Where("average_rating <= ?", value)
		case "created_from":
			db = db.Where("created_at >= ?", value)
		case "created_to":
			db = db.Where("created_at <= ?", value)
		}
	}

	// Count total
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting
	switch sortBy {
	case "name":
		db = db.Order("name " + sortOrder)
	case "price":
		db = db.Order("regular_price " + sortOrder)
	case "rating":
		db = db.Order("average_rating " + sortOrder)
	case "created_at":
		db = db.Order("created_at " + sortOrder)
	case "popularity":
		db = db.Order("view_count " + sortOrder)
	default:
		if query != "" {
			db = db.Order("MATCH(name, description, sku) AGAINST('" + query + "' IN NATURAL LANGUAGE MODE) DESC")
		} else {
			db = db.Order("created_at DESC")
		}
	}

	// Execute query
	var products []model.Product
	if err := db.Preload("Brand").Preload("Category").Find(&products).Error; err != nil {
		return nil, 0, err
	}

	// Convert to search results
	for _, product := range products {
		result := model.SearchResult{
			ID:          product.ID,
			Type:        "product",
			Title:       product.Name,
			Description: product.Description,
			URL:         fmt.Sprintf("/products/%d", product.ID),
			Image:       "", // TODO: Add ImageURL field to Product model
			Price:       &product.RegularPrice,
			Score:       1.0,
			Metadata: map[string]interface{}{
				"brand":    product.Brand.Name,
				"category": product.Category.Name,
				"sku":      product.SKU,
				"status":   product.Status,
				"stock":    product.StockQuantity,
				"rating":   0.0, // TODO: Add AverageRating field to Product model
			},
		}
		results = append(results, result)
	}

	return results, total, nil
}

// SearchCategories searches for categories
func (r *searchRepository) SearchCategories(query string, filters map[string]interface{}, sortBy, sortOrder string, page, limit int) ([]model.SearchResult, int64, error) {
	var results []model.SearchResult
	var total int64

	db := r.db.Model(&model.Category{}).Where("deleted_at IS NULL")

	if query != "" {
		db = db.Where("MATCH(name, description) AGAINST(? IN NATURAL LANGUAGE MODE)", query)
	}

	// Apply filters
	for key, value := range filters {
		switch key {
		case "parent_id":
			db = db.Where("parent_id = ?", value)
		case "level":
			db = db.Where("level = ?", value)
		case "is_active":
			db = db.Where("is_active = ?", value)
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting
	switch sortBy {
	case "name":
		db = db.Order("name " + sortOrder)
	case "level":
		db = db.Order("level " + sortOrder)
	case "created_at":
		db = db.Order("created_at " + sortOrder)
	default:
		if query != "" {
			db = db.Order("MATCH(name, description) AGAINST('" + query + "' IN NATURAL LANGUAGE MODE) DESC")
		} else {
			db = db.Order("name ASC")
		}
	}

	var categories []model.Category
	if err := db.Find(&categories).Error; err != nil {
		return nil, 0, err
	}

	for _, category := range categories {
		result := model.SearchResult{
			ID:          category.ID,
			Type:        "category",
			Title:       category.Name,
			Description: category.Description,
			URL:         fmt.Sprintf("/categories/%d", category.ID),
			Score:       1.0,
			Metadata: map[string]interface{}{
				"level":     category.Level,
				"parent_id": category.ParentID,
				"is_active": category.IsActive,
				"slug":      category.Slug,
			},
		}
		results = append(results, result)
	}

	return results, total, nil
}

// SearchBrands searches for brands
func (r *searchRepository) SearchBrands(query string, filters map[string]interface{}, sortBy, sortOrder string, page, limit int) ([]model.SearchResult, int64, error) {
	var results []model.SearchResult
	var total int64

	db := r.db.Model(&model.Brand{}).Where("deleted_at IS NULL")

	if query != "" {
		db = db.Where("MATCH(name, description) AGAINST(? IN NATURAL LANGUAGE MODE)", query)
	}

	// Apply filters
	for key, value := range filters {
		switch key {
		case "country":
			db = db.Where("country = ?", value)
		case "is_active":
			db = db.Where("is_active = ?", value)
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting
	switch sortBy {
	case "name":
		db = db.Order("name " + sortOrder)
	case "created_at":
		db = db.Order("created_at " + sortOrder)
	default:
		if query != "" {
			db = db.Order("MATCH(name, description) AGAINST('" + query + "' IN NATURAL LANGUAGE MODE) DESC")
		} else {
			db = db.Order("name ASC")
		}
	}

	var brands []model.Brand
	if err := db.Find(&brands).Error; err != nil {
		return nil, 0, err
	}

	for _, brand := range brands {
		result := model.SearchResult{
			ID:          brand.ID,
			Type:        "brand",
			Title:       brand.Name,
			Description: brand.Description,
			URL:         fmt.Sprintf("/brands/%d", brand.ID),
			Image:       "", // TODO: Add LogoURL field to Brand model
			Score:       1.0,
			Metadata: map[string]interface{}{
				"country":   "", // TODO: Add Country field to Brand model
				"is_active": brand.IsActive,
				"slug":      brand.Slug,
			},
		}
		results = append(results, result)
	}

	return results, total, nil
}

// SearchUsers searches for users
func (r *searchRepository) SearchUsers(query string, filters map[string]interface{}, sortBy, sortOrder string, page, limit int) ([]model.SearchResult, int64, error) {
	var results []model.SearchResult
	var total int64

	db := r.db.Model(&model.User{}).Where("deleted_at IS NULL")

	if query != "" {
		db = db.Where("MATCH(username, email, first_name, last_name) AGAINST(? IN NATURAL LANGUAGE MODE)", query)
	}

	// Apply filters
	for key, value := range filters {
		switch key {
		case "role":
			db = db.Where("role = ?", value)
		case "status":
			db = db.Where("status = ?", value)
		case "created_from":
			db = db.Where("created_at >= ?", value)
		case "created_to":
			db = db.Where("created_at <= ?", value)
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting
	switch sortBy {
	case "username":
		db = db.Order("username " + sortOrder)
	case "email":
		db = db.Order("email " + sortOrder)
	case "created_at":
		db = db.Order("created_at " + sortOrder)
	default:
		if query != "" {
			db = db.Order("MATCH(username, email, first_name, last_name) AGAINST('" + query + "' IN NATURAL LANGUAGE MODE) DESC")
		} else {
			db = db.Order("created_at DESC")
		}
	}

	var users []model.User
	if err := db.Find(&users).Error; err != nil {
		return nil, 0, err
	}

	for _, user := range users {
		result := model.SearchResult{
			ID:          user.ID,
			Type:        "user",
			Title:       user.Username,
			Description: user.Email,
			URL:         fmt.Sprintf("/users/%d", user.ID),
			Score:       1.0,
			Metadata: map[string]interface{}{
				"email":      user.Email,
				"role":       user.Role,
				"status":     "active", // TODO: Add Status field to User model
				"first_name": user.FirstName,
				"last_name":  user.LastName,
			},
		}
		results = append(results, result)
	}

	return results, total, nil
}

// SearchWishlists searches for wishlists
func (r *searchRepository) SearchWishlists(query string, filters map[string]interface{}, sortBy, sortOrder string, page, limit int) ([]model.SearchResult, int64, error) {
	var results []model.SearchResult
	var total int64

	db := r.db.Model(&model.Wishlist{}).Where("deleted_at IS NULL")

	if query != "" {
		db = db.Where("MATCH(name, description) AGAINST(? IN NATURAL LANGUAGE MODE)", query)
	}

	// Apply filters
	for key, value := range filters {
		switch key {
		case "user_id":
			db = db.Where("user_id = ?", value)
		case "is_public":
			db = db.Where("is_public = ?", value)
		case "is_default":
			db = db.Where("is_default = ?", value)
		case "status":
			db = db.Where("status = ?", value)
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting
	switch sortBy {
	case "name":
		db = db.Order("name " + sortOrder)
	case "view_count":
		db = db.Order("view_count " + sortOrder)
	case "created_at":
		db = db.Order("created_at " + sortOrder)
	default:
		if query != "" {
			db = db.Order("MATCH(name, description) AGAINST('" + query + "' IN NATURAL LANGUAGE MODE) DESC")
		} else {
			db = db.Order("created_at DESC")
		}
	}

	var wishlists []model.Wishlist
	if err := db.Preload("User").Find(&wishlists).Error; err != nil {
		return nil, 0, err
	}

	for _, wishlist := range wishlists {
		result := model.SearchResult{
			ID:          wishlist.ID,
			Type:        "wishlist",
			Title:       wishlist.Name,
			Description: wishlist.Description,
			URL:         fmt.Sprintf("/wishlists/%d", wishlist.ID),
			Score:       1.0,
			Metadata: map[string]interface{}{
				"user_id":    wishlist.UserID,
				"username":   wishlist.User.Username,
				"is_public":  wishlist.IsPublic(),
				"is_default": wishlist.IsDefault,
				"status":     wishlist.Status,
				"item_count": wishlist.ItemCount,
				"view_count": wishlist.ViewCount,
			},
		}
		results = append(results, result)
	}

	return results, total, nil
}

// SearchReviews searches for reviews
func (r *searchRepository) SearchReviews(query string, filters map[string]interface{}, sortBy, sortOrder string, page, limit int) ([]model.SearchResult, int64, error) {
	var results []model.SearchResult
	var total int64

	db := r.db.Model(&model.Review{}).Where("deleted_at IS NULL")

	if query != "" {
		db = db.Where("MATCH(title, content) AGAINST(? IN NATURAL LANGUAGE MODE)", query)
	}

	// Apply filters
	for key, value := range filters {
		switch key {
		case "product_id":
			db = db.Where("product_id = ?", value)
		case "user_id":
			db = db.Where("user_id = ?", value)
		case "rating":
			db = db.Where("rating = ?", value)
		case "rating_min":
			db = db.Where("rating >= ?", value)
		case "rating_max":
			db = db.Where("rating <= ?", value)
		case "status":
			db = db.Where("status = ?", value)
		case "is_verified":
			db = db.Where("is_verified = ?", value)
		case "created_from":
			db = db.Where("created_at >= ?", value)
		case "created_to":
			db = db.Where("created_at <= ?", value)
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting
	switch sortBy {
	case "rating":
		db = db.Order("rating " + sortOrder)
	case "created_at":
		db = db.Order("created_at " + sortOrder)
	case "helpful_count":
		db = db.Order("helpful_count " + sortOrder)
	default:
		if query != "" {
			db = db.Order("MATCH(title, content) AGAINST('" + query + "' IN NATURAL LANGUAGE MODE) DESC")
		} else {
			db = db.Order("created_at DESC")
		}
	}

	var reviews []model.Review
	if err := db.Preload("User").Preload("Product").Find(&reviews).Error; err != nil {
		return nil, 0, err
	}

	for _, review := range reviews {
		result := model.SearchResult{
			ID:          review.ID,
			Type:        "review",
			Title:       review.Title,
			Description: review.Content,
			URL:         fmt.Sprintf("/reviews/%d", review.ID),
			Score:       1.0,
			Metadata: map[string]interface{}{
				"product_id":    review.ProductID,
				"product_name":  review.Product.Name,
				"user_id":       review.UserID,
				"username":      review.User.Username,
				"rating":        review.Rating,
				"status":        review.Status,
				"is_verified":   review.IsVerified,
				"helpful_count": 0, // TODO: Add HelpfulCount field to Review model
			},
		}
		results = append(results, result)
	}

	return results, total, nil
}

// SearchOrders searches for orders
func (r *searchRepository) SearchOrders(query string, filters map[string]interface{}, sortBy, sortOrder string, page, limit int) ([]model.SearchResult, int64, error) {
	var results []model.SearchResult
	var total int64

	db := r.db.Model(&model.Order{}).Where("deleted_at IS NULL")

	if query != "" {
		db = db.Where("MATCH(order_number, notes) AGAINST(? IN NATURAL LANGUAGE MODE)", query)
	}

	// Apply filters
	for key, value := range filters {
		switch key {
		case "user_id":
			db = db.Where("user_id = ?", value)
		case "status":
			db = db.Where("status = ?", value)
		case "payment_status":
			db = db.Where("payment_status = ?", value)
		case "total_min":
			db = db.Where("total_amount >= ?", value)
		case "total_max":
			db = db.Where("total_amount <= ?", value)
		case "created_from":
			db = db.Where("created_at >= ?", value)
		case "created_to":
			db = db.Where("created_at <= ?", value)
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	// Apply sorting
	switch sortBy {
	case "order_number":
		db = db.Order("order_number " + sortOrder)
	case "total_amount":
		db = db.Order("total_amount " + sortOrder)
	case "created_at":
		db = db.Order("created_at " + sortOrder)
	default:
		if query != "" {
			db = db.Order("MATCH(order_number, notes) AGAINST('" + query + "' IN NATURAL LANGUAGE MODE) DESC")
		} else {
			db = db.Order("created_at DESC")
		}
	}

	var orders []model.Order
	if err := db.Preload("User").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	for _, order := range orders {
		result := model.SearchResult{
			ID:          order.ID,
			Type:        "order",
			Title:       order.OrderNumber,
			Description: order.Notes,
			URL:         fmt.Sprintf("/orders/%d", order.ID),
			Price:       &order.TotalAmount,
			Score:       1.0,
			Metadata: map[string]interface{}{
				"user_id":        order.UserID,
				"username":       order.User.Username,
				"status":         order.Status,
				"payment_status": order.PaymentStatus,
				"total_amount":   order.TotalAmount,
				"item_count":     0, // TODO: Add ItemCount field to Order model
			},
		}
		results = append(results, result)
	}

	return results, total, nil
}

// SearchAll searches across all entity types
func (r *searchRepository) SearchAll(query string, filters map[string]interface{}, sortBy, sortOrder string, page, limit int) ([]model.SearchResult, int64, error) {
	var allResults []model.SearchResult
	var total int64

	// Search products
	products, productTotal, err := r.SearchProducts(query, filters, sortBy, sortOrder, page, limit)
	if err != nil {
		return nil, 0, err
	}
	allResults = append(allResults, products...)
	total += productTotal

	// Search categories
	categories, categoryTotal, err := r.SearchCategories(query, filters, sortBy, sortOrder, page, limit)
	if err != nil {
		return nil, 0, err
	}
	allResults = append(allResults, categories...)
	total += categoryTotal

	// Search brands
	brands, brandTotal, err := r.SearchBrands(query, filters, sortBy, sortOrder, page, limit)
	if err != nil {
		return nil, 0, err
	}
	allResults = append(allResults, brands...)
	total += brandTotal

	return allResults, total, nil
}

// Search Query Management

// CreateSearchQuery creates a new search query
func (r *searchRepository) CreateSearchQuery(query *model.SearchQuery) error {
	return r.db.Create(query).Error
}

// GetSearchQueryByID retrieves a search query by ID
func (r *searchRepository) GetSearchQueryByID(id uint) (*model.SearchQuery, error) {
	var query model.SearchQuery
	if err := r.db.Preload("User").First(&query, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &query, nil
}

// GetSearchQueriesByUser retrieves search queries for a user
func (r *searchRepository) GetSearchQueriesByUser(userID uint, page, limit int) ([]model.SearchQuery, int64, error) {
	var queries []model.SearchQuery
	var total int64
	db := r.db.Model(&model.SearchQuery{}).Where("user_id = ?", userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	db = db.Order("created_at DESC")

	if err := db.Preload("User").Find(&queries).Error; err != nil {
		return nil, 0, err
	}

	return queries, total, nil
}

// UpdateSearchQuery updates a search query
func (r *searchRepository) UpdateSearchQuery(query *model.SearchQuery) error {
	return r.db.Save(query).Error
}

// DeleteSearchQuery deletes a search query
func (r *searchRepository) DeleteSearchQuery(id uint) error {
	return r.db.Delete(&model.SearchQuery{}, id).Error
}

// Search History

// CreateSearchHistory creates a new search history entry
func (r *searchRepository) CreateSearchHistory(history *model.SearchHistory) error {
	return r.db.Create(history).Error
}

// GetSearchHistoryByUser retrieves search history for a user
func (r *searchRepository) GetSearchHistoryByUser(userID uint, page, limit int) ([]model.SearchHistory, int64, error) {
	var history []model.SearchHistory
	var total int64
	db := r.db.Model(&model.SearchHistory{}).Where("user_id = ?", userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	db = db.Order("created_at DESC")

	if err := db.Preload("User").Find(&history).Error; err != nil {
		return nil, 0, err
	}

	return history, total, nil
}

// DeleteSearchHistory deletes a search history entry
func (r *searchRepository) DeleteSearchHistory(id uint) error {
	return r.db.Delete(&model.SearchHistory{}, id).Error
}

// ClearUserSearchHistory clears all search history for a user
func (r *searchRepository) ClearUserSearchHistory(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.SearchHistory{}).Error
}

// Search Suggestions

// GetSearchSuggestions retrieves search suggestions
func (r *searchRepository) GetSearchSuggestions(query string, searchType model.SearchType, limit int) ([]string, error) {
	var suggestions []model.SearchSuggestion
	db := r.db.Model(&model.SearchSuggestion{}).
		Where("query LIKE ? AND search_type = ? AND is_active = ?", "%"+query+"%", searchType, true).
		Order("count DESC")

	if limit > 0 {
		db = db.Limit(limit)
	}

	if err := db.Find(&suggestions).Error; err != nil {
		return nil, err
	}

	var result []string
	for _, suggestion := range suggestions {
		result = append(result, suggestion.Query)
	}

	return result, nil
}

// CreateSearchSuggestion creates a new search suggestion
func (r *searchRepository) CreateSearchSuggestion(suggestion *model.SearchSuggestion) error {
	return r.db.Create(suggestion).Error
}

// UpdateSearchSuggestion updates a search suggestion
func (r *searchRepository) UpdateSearchSuggestion(suggestion *model.SearchSuggestion) error {
	return r.db.Save(suggestion).Error
}

// DeleteSearchSuggestion deletes a search suggestion
func (r *searchRepository) DeleteSearchSuggestion(id uint) error {
	return r.db.Delete(&model.SearchSuggestion{}, id).Error
}

// Search Filters

// GetSearchFilters retrieves search filters for a search type
func (r *searchRepository) GetSearchFilters(searchType model.SearchType) ([]model.SearchFilter, error) {
	var filters []model.SearchFilter
	err := r.db.Where("search_type = ? AND is_active = ?", searchType, true).
		Order("sort_order ASC").
		Find(&filters).Error
	return filters, err
}

// CreateSearchFilter creates a new search filter
func (r *searchRepository) CreateSearchFilter(filter *model.SearchFilter) error {
	return r.db.Create(filter).Error
}

// UpdateSearchFilter updates a search filter
func (r *searchRepository) UpdateSearchFilter(filter *model.SearchFilter) error {
	return r.db.Save(filter).Error
}

// DeleteSearchFilter deletes a search filter
func (r *searchRepository) DeleteSearchFilter(id uint) error {
	return r.db.Delete(&model.SearchFilter{}, id).Error
}

// Search Index

// CreateSearchIndex creates a new search index entry
func (r *searchRepository) CreateSearchIndex(index *model.SearchIndex) error {
	return r.db.Create(index).Error
}

// UpdateSearchIndex updates a search index entry
func (r *searchRepository) UpdateSearchIndex(index *model.SearchIndex) error {
	return r.db.Save(index).Error
}

// DeleteSearchIndex deletes a search index entry
func (r *searchRepository) DeleteSearchIndex(id uint) error {
	return r.db.Delete(&model.SearchIndex{}, id).Error
}

// DeleteSearchIndexByEntity deletes search index entries for an entity
func (r *searchRepository) DeleteSearchIndexByEntity(entityType string, entityID uint) error {
	return r.db.Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Delete(&model.SearchIndex{}).Error
}

// GetSearchIndexByEntity retrieves search index entry for an entity
func (r *searchRepository) GetSearchIndexByEntity(entityType string, entityID uint) (*model.SearchIndex, error) {
	var index model.SearchIndex
	if err := r.db.Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		First(&index).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &index, nil
}

// Search Analytics

// GetSearchStats retrieves search statistics
func (r *searchRepository) GetSearchStats(startDate, endDate *time.Time, searchType *model.SearchType) (*model.SearchStatsResponse, error) {
	var stats model.SearchStatsResponse

	// Build base query
	db := r.db.Model(&model.SearchQuery{}).Where("deleted_at IS NULL")
	if startDate != nil {
		db = db.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		db = db.Where("created_at <= ?", *endDate)
	}
	if searchType != nil {
		db = db.Where("search_type = ?", *searchType)
	}

	// Total searches
	var totalSearches int64
	db.Count(&totalSearches)
	stats.TotalSearches = totalSearches

	// Unique queries
	var uniqueQueries int64
	r.db.Model(&model.SearchQuery{}).Where("deleted_at IS NULL").
		Distinct("query").Count(&uniqueQueries)
	stats.UniqueQueries = uniqueQueries

	// Total results
	var totalResults int64
	r.db.Model(&model.SearchQuery{}).Where("deleted_at IS NULL").
		Select("SUM(results)").Scan(&totalResults)
	stats.TotalResults = totalResults

	// Zero results
	var zeroResults int64
	r.db.Model(&model.SearchQuery{}).Where("deleted_at IS NULL AND results = 0").
		Count(&zeroResults)
	stats.ZeroResults = zeroResults

	// Average results
	var avgResults float64
	r.db.Model(&model.SearchQuery{}).Where("deleted_at IS NULL").
		Select("AVG(results)").Scan(&avgResults)
	stats.AvgResults = avgResults

	// Average duration
	var avgDuration float64
	r.db.Model(&model.SearchQuery{}).Where("deleted_at IS NULL").
		Select("AVG(duration)").Scan(&avgDuration)
	stats.AvgDuration = avgDuration

	// Top queries
	topQueries, err := r.GetTopQueries(10, searchType)
	if err != nil {
		return nil, err
	}
	stats.TopQueries = topQueries

	// Daily stats
	dailyStats, err := r.GetDailySearchStats(startDate, endDate, searchType)
	if err != nil {
		return nil, err
	}
	stats.DailyStats = dailyStats

	return &stats, nil
}

// GetDailySearchStats retrieves daily search statistics
func (r *searchRepository) GetDailySearchStats(startDate, endDate *time.Time, searchType *model.SearchType) ([]model.DailySearchStats, error) {
	var stats []model.DailySearchStats

	query := `
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as searches,
			COUNT(DISTINCT query) as unique_queries,
			AVG(results) as avg_results,
			AVG(duration) as avg_duration
		FROM search_queries 
		WHERE deleted_at IS NULL
	`

	args := []interface{}{}
	if startDate != nil {
		query += " AND created_at >= ?"
		args = append(args, *startDate)
	}
	if endDate != nil {
		query += " AND created_at <= ?"
		args = append(args, *endDate)
	}
	if searchType != nil {
		query += " AND search_type = ?"
		args = append(args, *searchType)
	}

	query += " GROUP BY DATE(created_at) ORDER BY date DESC"

	if err := r.db.Raw(query, args...).Scan(&stats).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// GetTopQueries retrieves top search queries
func (r *searchRepository) GetTopQueries(limit int, searchType *model.SearchType) ([]model.QueryStats, error) {
	var stats []model.QueryStats

	db := r.db.Model(&model.SearchQuery{}).
		Select("query, COUNT(*) as count, search_type as type").
		Where("deleted_at IS NULL").
		Group("query, search_type").
		Order("count DESC")

	if searchType != nil {
		db = db.Where("search_type = ?", *searchType)
	}

	if limit > 0 {
		db = db.Limit(limit)
	}

	if err := db.Scan(&stats).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// GetZeroResultQueries retrieves zero result queries
func (r *searchRepository) GetZeroResultQueries(limit int, searchType *model.SearchType) ([]string, error) {
	var queries []string

	db := r.db.Model(&model.SearchQuery{}).
		Select("query").
		Where("deleted_at IS NULL AND results = 0").
		Group("query").
		Order("COUNT(*) DESC")

	if searchType != nil {
		db = db.Where("search_type = ?", *searchType)
	}

	if limit > 0 {
		db = db.Limit(limit)
	}

	if err := db.Pluck("query", &queries).Error; err != nil {
		return nil, err
	}

	return queries, nil
}

// Facets

// GetSearchFacets retrieves search facets
func (r *searchRepository) GetSearchFacets(searchType model.SearchType, filters map[string]interface{}) (map[string]model.Facet, error) {
	// This would typically involve complex aggregation queries
	// For now, return empty facets
	return make(map[string]model.Facet), nil
}

// UpdateSearchFacets updates search facets cache
func (r *searchRepository) UpdateSearchFacets(searchType model.SearchType, facets map[string]model.Facet) error {
	// This would update the search_facets table
	return nil
}
