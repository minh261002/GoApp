package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go_app/pkg/jwt"
	"go_app/pkg/logger"
	"go_app/pkg/redis"
	"go_app/pkg/response"
	"go_app/pkg/validator"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// CORS configures CORS middleware
func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// RequestID thêm request ID vào mỗi request
func RequestID() gin.HandlerFunc {
	return requestid.New()
}

// Logger middleware ghi log cho mỗi request
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.WithFields(map[string]interface{}{
			"status_code": param.StatusCode,
			"latency":     param.Latency,
			"client_ip":   param.ClientIP,
			"method":      param.Method,
			"path":        param.Path,
			"user_agent":  param.Request.UserAgent(),
		}).Info("HTTP Request")
		return ""
	})
}

// Recovery middleware xử lý panic
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.WithField("error", recovered).Error("Panic recovered")
		response.InternalServerError(c, "Internal server error")
		c.Abort()
	})
}

// Auth middleware xác thực JWT token
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header required")
			c.Abort()
			return
		}

		// Kiểm tra format "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		token := parts[1]

		// Validate JWT token
		jwtManager := jwt.NewJWTManager()
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			logger.WithField("error", err.Error()).Warn("Invalid JWT token")
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Lưu thông tin user vào context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("token_claims", claims)

		c.Next()
	}
}

// OptionalAuth middleware xác thực JWT token (không bắt buộc)
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Kiểm tra format "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		token := parts[1]

		// Validate JWT token
		jwtManager := jwt.NewJWTManager()
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			logger.WithField("error", err.Error()).Debug("Invalid JWT token in optional auth")
			c.Next()
			return
		}

		// Lưu thông tin user vào context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("token_claims", claims)

		c.Next()
	}
}

// RateLimitConfig cấu hình rate limiting
type RateLimitConfig struct {
	MaxRequests int                       // Số request tối đa
	Window      time.Duration             // Thời gian window
	KeyFunc     func(*gin.Context) string // Function tạo key
}

// RateLimit middleware limits number of requests using Redis
func RateLimit(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("rate_limit:%s", config.KeyFunc(c))
		ctx := context.Background()

		// Get current count from Redis
		count, err := redis.Increment(ctx, key)
		if err != nil {
			logger.WithField("error", err.Error()).Error("Failed to increment rate limit counter")
			c.Next() // Continue if Redis fails
			return
		}

		// Set expiration on first request
		if count == 1 {
			redis.Expire(ctx, key, config.Window)
		}

		// Check if limit exceeded
		if count > int64(config.MaxRequests) {
			logger.WithFields(map[string]interface{}{
				"key":   key,
				"count": count,
				"limit": config.MaxRequests,
			}).Warn("Rate limit exceeded")

			response.Error(c, http.StatusTooManyRequests, "Rate limit exceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}

// DefaultRateLimit trả về rate limit mặc định
func DefaultRateLimit() gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		MaxRequests: 100,
		Window:      time.Minute,
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
	})
}

// StrictRateLimit trả về rate limit nghiêm ngặt hơn
func StrictRateLimit() gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		MaxRequests: 10,
		Window:      time.Minute,
		KeyFunc: func(c *gin.Context) string {
			// Sử dụng user ID nếu có, nếu không thì dùng IP
			if userID, exists := c.Get("user_id"); exists {
				return fmt.Sprintf("user:%v", userID)
			}
			return c.ClientIP()
		},
	})
}

// RoleMiddleware middleware kiểm tra quyền truy cập dựa trên role
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Lấy thông tin user từ context (được set bởi Auth middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			response.Unauthorized(c, "User not authenticated")
			c.Abort()
			return
		}

		// TODO: Lấy role của user từ database
		// Trong thực tế, bạn cần query database để lấy role của user
		// userRole := getUserRoleFromDB(userID)

		// Tạm thời sử dụng role mặc định hoặc từ JWT claims
		userRole := "user" // Mặc định
		if claims, exists := c.Get("token_claims"); exists {
			if _, ok := claims.(*jwt.Claims); ok {
				// Có thể thêm role vào JWT claims
				// userRole = jwtClaims.Role
			}
		}

		// Kiểm tra role
		hasPermission := false
		for _, role := range allowedRoles {
			if userRole == role {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			logger.WithFields(map[string]interface{}{
				"user_id":        userID,
				"user_role":      userRole,
				"required_roles": allowedRoles,
			}).Warn("Access denied - insufficient permissions")
			response.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminOnly middleware chỉ cho phép admin
func AdminOnly() gin.HandlerFunc {
	return RoleMiddleware("admin")
}

// UserOrAdmin middleware cho phép user và admin
func UserOrAdmin() gin.HandlerFunc {
	return RoleMiddleware("user", "admin")
}

// ValidateJSON middleware validate JSON request
func ValidateJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			if c.GetHeader("Content-Type") != "application/json" {
				response.BadRequest(c, "Content-Type must be application/json")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// ValidateStruct middleware validate struct với validator
func ValidateStruct(model interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bind JSON vào struct
		if err := c.ShouldBindJSON(model); err != nil {
			logger.WithField("error", err.Error()).Warn("Failed to bind JSON")
			response.BadRequest(c, "Invalid JSON format", err.Error())
			c.Abort()
			return
		}

		// Validate struct
		validatorInstance := validator.NewCustomValidator()
		if err := validatorInstance.ValidateStruct(model); err != nil {
			validationErrors := validatorInstance.GetValidationErrors(err)
			logger.WithFields(map[string]interface{}{
				"validation_errors": validationErrors,
				"model":             fmt.Sprintf("%T", model),
			}).Warn("Validation failed")
			response.ValidationError(c, validationErrors)
			c.Abort()
			return
		}

		// Lưu validated model vào context
		c.Set("validated_model", model)
		c.Next()
	}
}

// ValidateQuery middleware validate query parameters
func ValidateQuery(model interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bind query parameters vào struct
		if err := c.ShouldBindQuery(model); err != nil {
			logger.WithField("error", err.Error()).Warn("Failed to bind query parameters")
			response.BadRequest(c, "Invalid query parameters", err.Error())
			c.Abort()
			return
		}

		// Validate struct
		validatorInstance := validator.NewCustomValidator()
		if err := validatorInstance.ValidateStruct(model); err != nil {
			validationErrors := validatorInstance.GetValidationErrors(err)
			logger.WithFields(map[string]interface{}{
				"validation_errors": validationErrors,
				"model":             fmt.Sprintf("%T", model),
			}).Warn("Query validation failed")
			response.ValidationError(c, validationErrors)
			c.Abort()
			return
		}

		// Lưu validated model vào context
		c.Set("validated_query", model)
		c.Next()
	}
}

// ValidateForm middleware validate form data
func ValidateForm(model interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bind form data vào struct
		if err := c.ShouldBind(model); err != nil {
			logger.WithField("error", err.Error()).Warn("Failed to bind form data")
			response.BadRequest(c, "Invalid form data", err.Error())
			c.Abort()
			return
		}

		// Validate struct
		validatorInstance := validator.NewCustomValidator()
		if err := validatorInstance.ValidateStruct(model); err != nil {
			validationErrors := validatorInstance.GetValidationErrors(err)
			logger.WithFields(map[string]interface{}{
				"validation_errors": validationErrors,
				"model":             fmt.Sprintf("%T", model),
			}).Warn("Form validation failed")
			response.ValidationError(c, validationErrors)
			c.Abort()
			return
		}

		// Lưu validated model vào context
		c.Set("validated_form", model)
		c.Next()
	}
}

// SecurityHeaders thêm các security headers
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// HealthCheck middleware cho health check endpoint
func HealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.HealthCheck(c)
	}
}

// NotFoundHandler xử lý 404
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.NotFound(c, "Route not found")
	}
}

// MethodNotAllowedHandler xử lý 405
func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Error(c, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// APIVersioning middleware xử lý API versioning
func APIVersioning() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Lấy version từ header hoặc query parameter
		version := c.GetHeader("API-Version")
		if version == "" {
			version = c.Query("version")
		}
		if version == "" {
			version = "v1" // Mặc định
		}

		// Lưu version vào context
		c.Set("api_version", version)

		// Thêm version vào response header
		c.Header("API-Version", version)

		c.Next()
	}
}

// RequestTimeout middleware thiết lập timeout cho request
func RequestTimeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Tạo context với timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Cập nhật request context
		c.Request = c.Request.WithContext(ctx)

		// Tạo channel để theo dõi completion
		done := make(chan bool, 1)

		// Chạy handler trong goroutine
		go func() {
			c.Next()
			done <- true
		}()

		// Chờ completion hoặc timeout
		select {
		case <-done:
			// Request hoàn thành thành công
			return
		case <-ctx.Done():
			// Request timeout
			logger.WithFields(map[string]interface{}{
				"path":    c.Request.URL.Path,
				"method":  c.Request.Method,
				"timeout": timeout.String(),
			}).Warn("Request timeout")

			response.Error(c, http.StatusRequestTimeout, "Request timeout")
			c.Abort()
			return
		}
	}
}

// DefaultRequestTimeout trả về timeout mặc định (30 giây)
func DefaultRequestTimeout() gin.HandlerFunc {
	return RequestTimeout(30 * time.Second)
}

// LongRequestTimeout trả về timeout dài hơn (5 phút)
func LongRequestTimeout() gin.HandlerFunc {
	return RequestTimeout(5 * time.Minute)
}

// RequestSizeLimit middleware giới hạn kích thước request
func RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Thiết lập MaxBytesReader
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

		c.Next()
	}
}

// DefaultRequestSizeLimit trả về giới hạn kích thước mặc định (10MB)
func DefaultRequestSizeLimit() gin.HandlerFunc {
	return RequestSizeLimit(10 << 20) // 10MB
}

// CacheControl middleware thêm cache control headers
func CacheControl(maxAge time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", int(maxAge.Seconds())))
		c.Next()
	}
}

// NoCache middleware thêm no-cache headers
func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}
