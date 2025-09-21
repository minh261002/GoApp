package model

import (
	"time"
)

// SearchType represents the type of search
type SearchType string

const (
	SearchTypeProduct  SearchType = "product"
	SearchTypeCategory SearchType = "category"
	SearchTypeBrand    SearchType = "brand"
	SearchTypeUser     SearchType = "user"
	SearchTypeWishlist SearchType = "wishlist"
	SearchTypeReview   SearchType = "review"
	SearchTypeOrder    SearchType = "order"
	SearchTypeAll      SearchType = "all"
)

// SearchStatus represents the status of a search
type SearchStatus string

const (
	SearchStatusActive   SearchStatus = "active"
	SearchStatusInactive SearchStatus = "inactive"
	SearchStatusPending  SearchStatus = "pending"
	SearchStatusExpired  SearchStatus = "expired"
)

// FilterType represents the type of filter
type FilterType string

const (
	FilterTypeRange       FilterType = "range"
	FilterTypeSelect      FilterType = "select"
	FilterTypeMultiSelect FilterType = "multi_select"
	FilterTypeBoolean     FilterType = "boolean"
	FilterTypeDate        FilterType = "date"
	FilterTypeText        FilterType = "text"
)

// SearchQuery represents a search query
type SearchQuery struct {
	ID         uint         `json:"id" gorm:"primaryKey"`
	UserID     *uint        `json:"user_id" gorm:"index"`
	Query      string       `json:"query" gorm:"size:500;not null"`
	SearchType SearchType   `json:"search_type" gorm:"size:50;not null"`
	Filters    string       `json:"filters" gorm:"type:json"`
	SortBy     string       `json:"sort_by" gorm:"size:100"`
	SortOrder  string       `json:"sort_order" gorm:"size:10;default:'desc'"`
	Page       int          `json:"page" gorm:"default:1"`
	Limit      int          `json:"limit" gorm:"default:10"`
	Results    int          `json:"results" gorm:"default:0"`
	Duration   int64        `json:"duration" gorm:"default:0"` // in milliseconds
	IPAddress  string       `json:"ip_address" gorm:"size:45"`
	UserAgent  string       `json:"user_agent" gorm:"type:text"`
	Referrer   string       `json:"referrer" gorm:"size:500"`
	Status     SearchStatus `json:"status" gorm:"size:50;default:'active'"`

	// Relationships
	User *User `json:"user" gorm:"foreignKey:UserID"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

// SearchFilter represents a search filter configuration
type SearchFilter struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	Name         string     `json:"name" gorm:"size:100;not null"`
	Label        string     `json:"label" gorm:"size:255;not null"`
	Description  string     `json:"description" gorm:"type:text"`
	Type         FilterType `json:"type" gorm:"size:50;not null"`
	Field        string     `json:"field" gorm:"size:100;not null"`
	Options      string     `json:"options" gorm:"type:json"` // For select/multi-select filters
	MinValue     *float64   `json:"min_value" gorm:"type:decimal(10,2)"`
	MaxValue     *float64   `json:"max_value" gorm:"type:decimal(10,2)"`
	DefaultValue string     `json:"default_value" gorm:"size:255"`
	IsRequired   bool       `json:"is_required" gorm:"default:false"`
	IsActive     bool       `json:"is_active" gorm:"default:true"`
	SortOrder    int        `json:"sort_order" gorm:"default:0"`
	SearchType   SearchType `json:"search_type" gorm:"size:50;not null"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

// SearchSuggestion represents a search suggestion
type SearchSuggestion struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Query       string     `json:"query" gorm:"size:255;not null;index"`
	Suggestions string     `json:"suggestions" gorm:"type:json"`
	SearchType  SearchType `json:"search_type" gorm:"size:50;not null"`
	Count       int64      `json:"count" gorm:"default:0"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

// SearchHistory represents user search history
type SearchHistory struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	UserID     uint       `json:"user_id" gorm:"not null;index"`
	Query      string     `json:"query" gorm:"size:500;not null"`
	SearchType SearchType `json:"search_type" gorm:"size:50;not null"`
	Results    int        `json:"results" gorm:"default:0"`
	IPAddress  string     `json:"ip_address" gorm:"size:45"`
	UserAgent  string     `json:"user_agent" gorm:"type:text"`

	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

// SearchAnalytics represents search analytics
type SearchAnalytics struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	Date          time.Time  `json:"date" gorm:"type:date;not null;index"`
	SearchType    SearchType `json:"search_type" gorm:"size:50;not null"`
	TotalSearches int64      `json:"total_searches" gorm:"default:0"`
	UniqueQueries int64      `json:"unique_queries" gorm:"default:0"`
	TotalResults  int64      `json:"total_results" gorm:"default:0"`
	ZeroResults   int64      `json:"zero_results" gorm:"default:0"`
	AvgResults    float64    `json:"avg_results" gorm:"type:decimal(10,2);default:0"`
	AvgDuration   float64    `json:"avg_duration" gorm:"type:decimal(10,2);default:0"`
	TopQueries    string     `json:"top_queries" gorm:"type:json"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

// SearchIndex represents a search index entry
type SearchIndex struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	EntityType string     `json:"entity_type" gorm:"size:50;not null;index"`
	EntityID   uint       `json:"entity_id" gorm:"not null;index"`
	Title      string     `json:"title" gorm:"size:500;not null"`
	Content    string     `json:"content" gorm:"type:text"`
	Keywords   string     `json:"keywords" gorm:"type:text"`
	Tags       string     `json:"tags" gorm:"type:json"`
	Metadata   string     `json:"metadata" gorm:"type:json"`
	Weight     float64    `json:"weight" gorm:"type:decimal(5,2);default:1.0"`
	IsActive   bool       `json:"is_active" gorm:"default:true"`
	IndexedAt  *time.Time `json:"indexed_at"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`

	// Composite index
	// UNIQUE KEY unique_entity (entity_type, entity_id)
}

// Request/Response DTOs

// SearchRequest represents a search request
type SearchRequest struct {
	Query      string                 `json:"query" binding:"required,min=1,max=500"`
	Type       SearchType             `json:"type" binding:"omitempty,oneof=product category brand user wishlist review order all"`
	Filters    map[string]interface{} `json:"filters" binding:"omitempty"`
	SortBy     string                 `json:"sort_by" binding:"omitempty,max=100"`
	SortOrder  string                 `json:"sort_order" binding:"omitempty,oneof=asc desc"`
	Page       int                    `json:"page" binding:"omitempty,min=1,max=1000"`
	Limit      int                    `json:"limit" binding:"omitempty,min=1,max=100"`
	Facets     bool                   `json:"facets" binding:"omitempty"`
	Suggest    bool                   `json:"suggest" binding:"omitempty"`
	CategoryID *uint                  `json:"category_id" binding:"omitempty"`
	BrandID    *uint                  `json:"brand_id" binding:"omitempty"`
	MinPrice   *float64               `json:"min_price" binding:"omitempty,min=0"`
	MaxPrice   *float64               `json:"max_price" binding:"omitempty,min=0"`
	InStock    *bool                  `json:"in_stock" binding:"omitempty"`
	OnSale     *bool                  `json:"on_sale" binding:"omitempty"`
	Rating     *float64               `json:"rating" binding:"omitempty,min=0,max=5"`
}

// SearchResponse represents a search response
type SearchResponse struct {
	Query         string                 `json:"query"`
	Type          SearchType             `json:"type"`
	Results       []SearchResult         `json:"results"`
	Facets        map[string]Facet       `json:"facets,omitempty"`
	Suggestions   []string               `json:"suggestions,omitempty"`
	Pagination    PaginationInfo         `json:"pagination"`
	Duration      int64                  `json:"duration"` // in milliseconds
	Total         int64                  `json:"total"`
	Filters       map[string]interface{} `json:"filters,omitempty"`
	Products      []Product              `json:"products,omitempty"`
	Page          int                    `json:"page"`
	Limit         int                    `json:"limit"`
	TotalPages    int                    `json:"total_pages"`
	HasNext       bool                   `json:"has_next"`
	HasPrev       bool                   `json:"has_prev"`
	FilterOptions *FilterOptions         `json:"filter_options,omitempty"`
	SearchTime    int64                  `json:"search_time"`
}

// SearchResult represents a single search result
type SearchResult struct {
	ID          uint                   `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	URL         string                 `json:"url"`
	Image       string                 `json:"image,omitempty"`
	Price       *float64               `json:"price,omitempty"`
	Score       float64                `json:"score"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Highlights  []string               `json:"highlights,omitempty"`
}

// Facet represents a search facet
type Facet struct {
	Name    string        `json:"name"`
	Label   string        `json:"label"`
	Type    FilterType    `json:"type"`
	Options []FacetOption `json:"options"`
	Min     *float64      `json:"min,omitempty"`
	Max     *float64      `json:"max,omitempty"`
}

// FacetOption represents a facet option
type FacetOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Count int64  `json:"count"`
}

// PaginationInfo represents pagination information
type PaginationInfo struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// FilterRequest represents a filter request
type FilterRequest struct {
	Name         string     `json:"name" binding:"required,min=1,max=100"`
	Label        string     `json:"label" binding:"required,min=1,max=255"`
	Description  string     `json:"description" binding:"omitempty,max=1000"`
	Type         FilterType `json:"type" binding:"required,oneof=range select multi_select boolean date text"`
	Field        string     `json:"field" binding:"required,min=1,max=100"`
	Options      []string   `json:"options" binding:"omitempty"`
	MinValue     *float64   `json:"min_value" binding:"omitempty"`
	MaxValue     *float64   `json:"max_value" binding:"omitempty"`
	DefaultValue string     `json:"default_value" binding:"omitempty,max=255"`
	IsRequired   bool       `json:"is_required" binding:"omitempty"`
	IsActive     bool       `json:"is_active" binding:"omitempty"`
	SortOrder    int        `json:"sort_order" binding:"omitempty,min=0"`
	SearchType   SearchType `json:"search_type" binding:"required,oneof=product category brand user wishlist review order all"`
}

// FilterResponse represents a filter response
type FilterResponse struct {
	ID           uint       `json:"id"`
	Name         string     `json:"name"`
	Label        string     `json:"label"`
	Description  string     `json:"description"`
	Type         FilterType `json:"type"`
	Field        string     `json:"field"`
	Options      []string   `json:"options"`
	MinValue     *float64   `json:"min_value"`
	MaxValue     *float64   `json:"max_value"`
	DefaultValue string     `json:"default_value"`
	IsRequired   bool       `json:"is_required"`
	IsActive     bool       `json:"is_active"`
	SortOrder    int        `json:"sort_order"`
	SearchType   SearchType `json:"search_type"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

// SearchStatsResponse represents search statistics
type SearchStatsResponse struct {
	TotalSearches     int64              `json:"total_searches"`
	UniqueQueries     int64              `json:"unique_queries"`
	TotalResults      int64              `json:"total_results"`
	ZeroResults       int64              `json:"zero_results"`
	AvgResults        float64            `json:"avg_results"`
	AvgDuration       float64            `json:"avg_duration"`
	TopQueries        []QueryStats       `json:"top_queries"`
	SearchTypes       map[string]int64   `json:"search_types"`
	DailyStats        []DailySearchStats `json:"daily_stats"`
	PopularFilters    []FilterStats      `json:"popular_filters"`
	ZeroResultQueries []string           `json:"zero_result_queries"`
}

// QueryStats represents query statistics
type QueryStats struct {
	Query string `json:"query"`
	Count int64  `json:"count"`
	Type  string `json:"type"`
}

// DailySearchStats represents daily search statistics
type DailySearchStats struct {
	Date          string  `json:"date"`
	Searches      int64   `json:"searches"`
	UniqueQueries int64   `json:"unique_queries"`
	AvgResults    float64 `json:"avg_results"`
	AvgDuration   float64 `json:"avg_duration"`
}

// FilterStats represents filter statistics
type FilterStats struct {
	Filter string `json:"filter"`
	Count  int64  `json:"count"`
	Type   string `json:"type"`
}

// SearchSuggestionRequest represents a search suggestion request
type SearchSuggestionRequest struct {
	Query string     `json:"query" binding:"required,min=1,max=255"`
	Type  SearchType `json:"type" binding:"omitempty,oneof=product category brand user wishlist review order all"`
	Limit int        `json:"limit" binding:"omitempty,min=1,max=20"`
}

// SearchSuggestionResponse represents a search suggestion response
type SearchSuggestionResponse struct {
	Query       string     `json:"query"`
	Suggestions []string   `json:"suggestions"`
	Type        SearchType `json:"type"`
}

// SearchIndexRequest represents a search index request
type SearchIndexRequest struct {
	EntityType string                 `json:"entity_type" binding:"required,min=1,max=50"`
	EntityID   uint                   `json:"entity_id" binding:"required"`
	Title      string                 `json:"title" binding:"required,min=1,max=500"`
	Content    string                 `json:"content" binding:"omitempty"`
	Keywords   string                 `json:"keywords" binding:"omitempty"`
	Tags       []string               `json:"tags" binding:"omitempty"`
	Metadata   map[string]interface{} `json:"metadata" binding:"omitempty"`
	Weight     float64                `json:"weight" binding:"omitempty,min=0.1,max=10.0"`
}

// SearchIndexResponse represents a search index response
type SearchIndexResponse struct {
	ID         uint                   `json:"id"`
	EntityType string                 `json:"entity_type"`
	EntityID   uint                   `json:"entity_id"`
	Title      string                 `json:"title"`
	Content    string                 `json:"content"`
	Keywords   string                 `json:"keywords"`
	Tags       []string               `json:"tags"`
	Metadata   map[string]interface{} `json:"metadata"`
	Weight     float64                `json:"weight"`
	IsActive   bool                   `json:"is_active"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	DeletedAt  *time.Time             `json:"deleted_at"`
}

// Helper methods

// ToResponse converts SearchQuery to SearchQueryResponse
func (sq *SearchQuery) ToResponse() *SearchQueryResponse {
	return &SearchQueryResponse{
		ID:         sq.ID,
		UserID:     sq.UserID,
		Query:      sq.Query,
		SearchType: sq.SearchType,
		Filters:    sq.Filters,
		SortBy:     sq.SortBy,
		SortOrder:  sq.SortOrder,
		Page:       sq.Page,
		Limit:      sq.Limit,
		Results:    sq.Results,
		Duration:   sq.Duration,
		IPAddress:  sq.IPAddress,
		UserAgent:  sq.UserAgent,
		Referrer:   sq.Referrer,
		Status:     sq.Status,
		User:       sq.User,
		CreatedAt:  sq.CreatedAt,
		UpdatedAt:  sq.UpdatedAt,
		DeletedAt:  sq.DeletedAt,
	}
}

// ToResponse converts SearchFilter to FilterResponse
func (sf *SearchFilter) ToResponse() *FilterResponse {
	return &FilterResponse{
		ID:           sf.ID,
		Name:         sf.Name,
		Label:        sf.Label,
		Description:  sf.Description,
		Type:         sf.Type,
		Field:        sf.Field,
		Options:      []string{}, // TODO: Parse JSON string to []string
		MinValue:     sf.MinValue,
		MaxValue:     sf.MaxValue,
		DefaultValue: sf.DefaultValue,
		IsRequired:   sf.IsRequired,
		IsActive:     sf.IsActive,
		SortOrder:    sf.SortOrder,
		SearchType:   sf.SearchType,
		CreatedAt:    sf.CreatedAt,
		UpdatedAt:    sf.UpdatedAt,
		DeletedAt:    sf.DeletedAt,
	}
}

// ToResponse converts SearchSuggestion to SearchSuggestionResponse
func (ss *SearchSuggestion) ToResponse() *SearchSuggestionResponse {
	return &SearchSuggestionResponse{
		Query:       ss.Query,
		Suggestions: []string{}, // TODO: Parse JSON string to []string
		Type:        ss.SearchType,
	}
}

// ToResponse converts SearchIndex to SearchIndexResponse
func (si *SearchIndex) ToResponse() *SearchIndexResponse {
	return &SearchIndexResponse{
		ID:         si.ID,
		EntityType: si.EntityType,
		EntityID:   si.EntityID,
		Title:      si.Title,
		Content:    si.Content,
		Keywords:   si.Keywords,
		Tags:       []string{},               // TODO: Parse JSON string to []string
		Metadata:   map[string]interface{}{}, // TODO: Parse JSON string to map
		Weight:     si.Weight,
		IsActive:   si.IsActive,
		CreatedAt:  si.CreatedAt,
		UpdatedAt:  si.UpdatedAt,
		DeletedAt:  si.DeletedAt,
	}
}

// SearchQueryResponse represents a search query response
type SearchQueryResponse struct {
	ID         uint         `json:"id"`
	UserID     *uint        `json:"user_id"`
	Query      string       `json:"query"`
	SearchType SearchType   `json:"search_type"`
	Filters    string       `json:"filters"`
	SortBy     string       `json:"sort_by"`
	SortOrder  string       `json:"sort_order"`
	Page       int          `json:"page"`
	Limit      int          `json:"limit"`
	Results    int          `json:"results"`
	Duration   int64        `json:"duration"`
	IPAddress  string       `json:"ip_address"`
	UserAgent  string       `json:"user_agent"`
	Referrer   string       `json:"referrer"`
	Status     SearchStatus `json:"status"`
	User       *User        `json:"user"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
	DeletedAt  *time.Time   `json:"deleted_at"`
}

// SearchFilters represents search filters
type SearchFilters struct {
	Query      string   `json:"query"`
	CategoryID *uint    `json:"category_id"`
	BrandID    *uint    `json:"brand_id"`
	MinPrice   *float64 `json:"min_price"`
	MaxPrice   *float64 `json:"max_price"`
	InStock    *bool    `json:"in_stock"`
	OnSale     *bool    `json:"on_sale"`
	Rating     *float64 `json:"rating"`
	SortBy     string   `json:"sort_by"`
	SortOrder  string   `json:"sort_order"`
	Page       int      `json:"page"`
	Limit      int      `json:"limit"`
}

// FilterOptions represents available filter options
type FilterOptions struct {
	PriceRange         *PriceRange          `json:"price_range,omitempty"`
	Categories         []Category           `json:"categories,omitempty"`
	Brands             []Brand              `json:"brands,omitempty"`
	RatingDistribution []RatingDistribution `json:"rating_distribution,omitempty"`
	Availability       *AvailabilityOptions `json:"availability,omitempty"`
}

// PriceRange represents price range filter
type PriceRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// RatingDistribution represents rating distribution
type RatingDistribution struct {
	Rating int `json:"rating"`
	Count  int `json:"count"`
}

// AvailabilityOptions represents availability options
type AvailabilityOptions struct {
	InStock    int `json:"in_stock"`
	OutOfStock int `json:"out_of_stock"`
	OnSale     int `json:"on_sale"`
}

// SearchLog represents a search log entry
type SearchLog struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	Query        string     `json:"query" gorm:"size:500;not null"`
	ResultCount  int        `json:"result_count" gorm:"default:0"`
	TotalResults int64      `json:"total_results" gorm:"default:0"`
	UserAgent    string     `json:"user_agent" gorm:"type:text"`
	IPAddress    string     `json:"ip_address" gorm:"size:45"`
	SearchTime   time.Time  `json:"search_time"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at" gorm:"index"`
}

// PopularSearch represents a popular search query
type PopularSearch struct {
	Query string `json:"query"`
	Count int64  `json:"count"`
}

// SearchTrend represents search trend data
type SearchTrend struct {
	Date  time.Time `json:"date"`
	Count int64     `json:"count"`
}

// SearchAnalyticsRequest represents search analytics request
type SearchAnalyticsRequest struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Days      int       `json:"days"`
	Limit     int       `json:"limit"`
}

// SearchAnalyticsResponse represents search analytics response
type SearchAnalyticsResponse struct {
	Stats           *SearchStats     `json:"stats"`
	TopQueries      []*TopQuery      `json:"top_queries"`
	Trends          []*SearchTrend   `json:"trends"`
	NoResultQueries []*NoResultQuery `json:"no_result_queries"`
}

// SearchStats represents search statistics
type SearchStats struct {
	TotalSearches   int64   `json:"total_searches"`
	UniqueQueries   int64   `json:"unique_queries"`
	NoResultQueries int64   `json:"no_result_queries"`
	AverageResults  float64 `json:"average_results"`
	AverageDuration float64 `json:"average_duration"`
}

// TopQuery represents a top search query
type TopQuery struct {
	Query string `json:"query"`
	Count int64  `json:"count"`
}

// NoResultQuery represents a query with no results
type NoResultQuery struct {
	Query string `json:"query"`
	Count int64  `json:"count"`
}

// SearchLogRequest represents search log request
type SearchLogRequest struct {
	Query     string     `json:"query"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	Page      int        `json:"page"`
	Limit     int        `json:"limit"`
}

// SearchIndexStats represents search index statistics
type SearchIndexStats struct {
	TotalIndexes  int64     `json:"total_indexes"`
	ActiveIndexes int64     `json:"active_indexes"`
	LastIndexedAt time.Time `json:"last_indexed_at"`
	IndexSize     int64     `json:"index_size"`
	AverageWeight float64   `json:"average_weight"`
}
