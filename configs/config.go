package configs

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for our application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Email    EmailConfig
	Upload   UploadConfig
	LogLevel string
	GinMode  string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
	Host string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	Charset  string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey     string
	AccessExpiry  int // in hours
	RefreshExpiry int // in hours
}

// EmailConfig holds email configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

// UploadConfig holds upload configuration
type UploadConfig struct {
	MaxFileSize     int64    // Maximum file size in bytes
	AllowedExts     []string // Allowed file extensions
	UploadPath      string   // Upload directory path
	PublicURL       string   // Public URL for uploaded files
	ImageMaxSize    int64    // Maximum image file size
	DocumentMaxSize int64    // Maximum document file size
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Host: getEnv("HOST", "localhost"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "go_app_db"),
			Charset:  "utf8mb4",
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			SecretKey:     getEnv("JWT_SECRET", "your-secret-key"),
			AccessExpiry:  getEnvAsInt("JWT_ACCESS_EXPIRY", 24),   // 24 hours
			RefreshExpiry: getEnvAsInt("JWT_REFRESH_EXPIRY", 168), // 7 days
		},
		Email: EmailConfig{
			SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
			SMTPUsername: getEnv("SMTP_USERNAME", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			FromEmail:    getEnv("FROM_EMAIL", ""),
			FromName:     getEnv("FROM_NAME", "Go App"),
		},
		Upload: UploadConfig{
			MaxFileSize:     getEnvAsInt64("UPLOAD_MAX_FILE_SIZE", 10*1024*1024), // 10MB
			AllowedExts:     getEnvAsStringSlice("UPLOAD_ALLOWED_EXTS", []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".pdf", ".doc", ".docx", ".txt", ".csv"}),
			UploadPath:      getEnv("UPLOAD_PATH", "uploads"),
			PublicURL:       getEnv("UPLOAD_PUBLIC_URL", "http://localhost:8080"),
			ImageMaxSize:    getEnvAsInt64("UPLOAD_IMAGE_MAX_SIZE", 5*1024*1024),     // 5MB
			DocumentMaxSize: getEnvAsInt64("UPLOAD_DOCUMENT_MAX_SIZE", 20*1024*1024), // 20MB
		},
		LogLevel: getEnv("LOG_LEVEL", "info"),
		GinMode:  getEnv("GIN_MODE", "debug"),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsInt64 gets an environment variable as int64 or returns a default value
func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsStringSlice gets an environment variable as string slice or returns a default value
func getEnvAsStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Split by comma and trim spaces
		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}
