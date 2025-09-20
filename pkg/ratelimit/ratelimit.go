package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	// Redis client for storing rate limit data
	Redis *redis.Client

	// Default rate limit settings
	DefaultRequests int           // Number of requests allowed
	DefaultWindow   time.Duration // Time window for rate limiting

	// Rate limit key prefix
	KeyPrefix string

	// Skip rate limiting for certain conditions
	SkipFunc func(*gin.Context) bool
}

// RateLimitRule represents a specific rate limit rule
type RateLimitRule struct {
	Requests int           // Number of requests allowed
	Window   time.Duration // Time window
	Key      string        // Custom key for this rule
	Message  string        // Custom error message
}

// RateLimitInfo represents rate limit information
type RateLimitInfo struct {
	Limit     int           `json:"limit"`
	Remaining int           `json:"remaining"`
	Reset     time.Time     `json:"reset"`
	Window    time.Duration `json:"window"`
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(config RateLimitConfig) gin.HandlerFunc {
	if config.Redis == nil {
		panic("Redis client is required for rate limiting")
	}

	if config.DefaultRequests <= 0 {
		config.DefaultRequests = 100 // Default: 100 requests
	}

	if config.DefaultWindow <= 0 {
		config.DefaultWindow = time.Hour // Default: 1 hour
	}

	if config.KeyPrefix == "" {
		config.KeyPrefix = "rate_limit"
	}

	return func(c *gin.Context) {
		// Skip rate limiting if skip function returns true
		if config.SkipFunc != nil && config.SkipFunc(c) {
			c.Next()
			return
		}

		// Get client identifier (IP address or user ID)
		clientID := getClientIdentifier(c)

		// Create rate limit key
		key := fmt.Sprintf("%s:%s:%s", config.KeyPrefix, c.Request.Method, clientID)

		// Check rate limit
		info, err := checkRateLimit(config.Redis, key, config.DefaultRequests, config.DefaultWindow)
		if err != nil {
			// If Redis is down, allow the request but log the error
			fmt.Printf("Rate limit check failed: %v\n", err)
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(info.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(info.Reset.Unix(), 10))

		// Check if rate limit exceeded
		if info.Remaining < 0 {
			c.JSON(429, gin.H{
				"error":   "Rate limit exceeded",
				"message": fmt.Sprintf("Too many requests. Limit: %d requests per %v", info.Limit, info.Window),
				"limit":   info.Limit,
				"reset":   info.Reset.Unix(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CustomRateLimitMiddleware creates a rate limiting middleware with custom rules
func CustomRateLimitMiddleware(config RateLimitConfig, rule RateLimitRule) gin.HandlerFunc {
	if config.Redis == nil {
		panic("Redis client is required for rate limiting")
	}

	if rule.Requests <= 0 {
		rule.Requests = config.DefaultRequests
	}

	if rule.Window <= 0 {
		rule.Window = config.DefaultWindow
	}

	if rule.Key == "" {
		rule.Key = fmt.Sprintf("%s:%s", config.KeyPrefix, "custom")
	}

	return func(c *gin.Context) {
		// Skip rate limiting if skip function returns true
		if config.SkipFunc != nil && config.SkipFunc(c) {
			c.Next()
			return
		}

		// Get client identifier
		clientID := getClientIdentifier(c)

		// Create rate limit key
		key := fmt.Sprintf("%s:%s", rule.Key, clientID)

		// Check rate limit
		info, err := checkRateLimit(config.Redis, key, rule.Requests, rule.Window)
		if err != nil {
			// If Redis is down, allow the request but log the error
			fmt.Printf("Rate limit check failed: %v\n", err)
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(info.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(info.Reset.Unix(), 10))

		// Check if rate limit exceeded
		if info.Remaining < 0 {
			message := rule.Message
			if message == "" {
				message = fmt.Sprintf("Too many requests. Limit: %d requests per %v", info.Limit, info.Window)
			}

			c.JSON(429, gin.H{
				"error":   "Rate limit exceeded",
				"message": message,
				"limit":   info.Limit,
				"reset":   info.Reset.Unix(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// IPBasedRateLimit creates IP-based rate limiting
func IPBasedRateLimit(requests int, window time.Duration) gin.HandlerFunc {
	config := RateLimitConfig{
		DefaultRequests: requests,
		DefaultWindow:   window,
		KeyPrefix:       "rate_limit:ip",
	}

	return RateLimitMiddleware(config)
}

// UserBasedRateLimit creates user-based rate limiting
func UserBasedRateLimit(requests int, window time.Duration) gin.HandlerFunc {
	config := RateLimitConfig{
		DefaultRequests: requests,
		DefaultWindow:   window,
		KeyPrefix:       "rate_limit:user",
	}

	return RateLimitMiddleware(config)
}

// APIKeyBasedRateLimit creates API key-based rate limiting
func APIKeyBasedRateLimit(requests int, window time.Duration) gin.HandlerFunc {
	config := RateLimitConfig{
		DefaultRequests: requests,
		DefaultWindow:   window,
		KeyPrefix:       "rate_limit:api_key",
	}

	return RateLimitMiddleware(config)
}

// TieredRateLimit creates tiered rate limiting based on user type
func TieredRateLimit(redis *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user type from context (set by auth middleware)
		userType, exists := c.Get("user_type")
		if !exists {
			userType = "guest"
		}

		var requests int
		var window time.Duration

		// Set rate limits based on user type
		switch userType {
		case "admin":
			requests = 10000 // 10k requests per hour
			window = time.Hour
		case "premium":
			requests = 1000 // 1k requests per hour
			window = time.Hour
		case "user":
			requests = 100 // 100 requests per hour
			window = time.Hour
		case "guest":
			requests = 10 // 10 requests per hour
			window = time.Hour
		default:
			requests = 10
			window = time.Hour
		}

		config := RateLimitConfig{
			Redis:           redis,
			DefaultRequests: requests,
			DefaultWindow:   window,
			KeyPrefix:       fmt.Sprintf("rate_limit:tiered:%s", userType),
		}

		// Apply rate limiting
		RateLimitMiddleware(config)(c)
	}
}

// EndpointSpecificRateLimit creates rate limiting for specific endpoints
func EndpointSpecificRateLimit(redis *redis.Client, endpoint string, requests int, window time.Duration) gin.HandlerFunc {
	rule := RateLimitRule{
		Requests: requests,
		Window:   window,
		Key:      fmt.Sprintf("rate_limit:endpoint:%s", endpoint),
		Message:  fmt.Sprintf("Rate limit exceeded for %s endpoint", endpoint),
	}

	config := RateLimitConfig{
		Redis: redis,
	}

	return CustomRateLimitMiddleware(config, rule)
}

// Helper functions

func getClientIdentifier(c *gin.Context) string {
	// Try to get user ID from context first (if authenticated)
	if userID, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("user:%v", userID)
	}

	// Try to get API key from header
	if apiKey := c.GetHeader("X-API-Key"); apiKey != "" {
		return fmt.Sprintf("api_key:%s", apiKey)
	}

	// Fall back to IP address
	ip := c.ClientIP()
	return fmt.Sprintf("ip:%s", ip)
}

func checkRateLimit(redis *redis.Client, key string, limit int, window time.Duration) (*RateLimitInfo, error) {
	ctx := context.Background()

	// Get current count
	result := redis.Get(ctx, key)
	if result.Err() != nil && result.Err().Error() != "redis: nil" {
		return nil, result.Err()
	}

	var count int
	// If key doesn't exist, create it
	if result.Err() != nil && result.Err().Error() == "redis: nil" {
		count = 0
	} else {
		var err error
		count, err = result.Int()
		if err != nil {
			return nil, err
		}
	}

	if count == 0 {
		// Set key with expiration
		err := redis.SetEx(ctx, key, 1, window).Err()
		if err != nil {
			return nil, err
		}

		return &RateLimitInfo{
			Limit:     limit,
			Remaining: limit - 1,
			Reset:     time.Now().Add(window),
			Window:    window,
		}, nil
	}

	// Check if limit exceeded
	if count >= limit {
		// Get TTL to calculate reset time
		ttl, err := redis.TTL(ctx, key).Result()
		if err != nil {
			return nil, err
		}

		return &RateLimitInfo{
			Limit:     limit,
			Remaining: 0,
			Reset:     time.Now().Add(ttl),
			Window:    window,
		}, nil
	}

	// Increment counter
	err := redis.Incr(ctx, key).Err()
	if err != nil {
		return nil, err
	}

	// Get TTL to calculate reset time
	ttl, err := redis.TTL(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	return &RateLimitInfo{
		Limit:     limit,
		Remaining: limit - count - 1,
		Reset:     time.Now().Add(ttl),
		Window:    window,
	}, nil
}

// RateLimitManager manages multiple rate limit rules
type RateLimitManager struct {
	redis  *redis.Client
	rules  map[string]RateLimitRule
	config RateLimitConfig
}

// NewRateLimitManager creates a new rate limit manager
func NewRateLimitManager(redis *redis.Client) *RateLimitManager {
	return &RateLimitManager{
		redis: redis,
		rules: make(map[string]RateLimitRule),
		config: RateLimitConfig{
			Redis:     redis,
			KeyPrefix: "rate_limit",
		},
	}
}

// AddRule adds a rate limit rule
func (m *RateLimitManager) AddRule(name string, rule RateLimitRule) {
	m.rules[name] = rule
}

// GetMiddleware returns a middleware for a specific rule
func (m *RateLimitManager) GetMiddleware(ruleName string) gin.HandlerFunc {
	rule, exists := m.rules[ruleName]
	if !exists {
		// Return default rate limiting
		return RateLimitMiddleware(m.config)
	}

	return CustomRateLimitMiddleware(m.config, rule)
}

// GetRateLimitInfo gets current rate limit info for a client
func (m *RateLimitManager) GetRateLimitInfo(ruleName, clientID string) (*RateLimitInfo, error) {
	rule, exists := m.rules[ruleName]
	if !exists {
		return nil, fmt.Errorf("rule %s not found", ruleName)
	}

	key := fmt.Sprintf("%s:%s", rule.Key, clientID)
	return checkRateLimit(m.redis, key, rule.Requests, rule.Window)
}

// ClearRateLimit clears rate limit for a client
func (m *RateLimitManager) ClearRateLimit(ruleName, clientID string) error {
	rule, exists := m.rules[ruleName]
	if !exists {
		return fmt.Errorf("rule %s not found", ruleName)
	}

	key := fmt.Sprintf("%s:%s", rule.Key, clientID)
	ctx := context.Background()
	return m.redis.Del(ctx, key).Err()
}
