package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"go_app/pkg/logger"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

// Config represents Redis configuration
type Config struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// GetConfig returns Redis configuration from environment variables
func GetConfig() *Config {
	return &Config{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     getEnv("REDIS_PORT", "6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       getEnvAsInt("REDIS_DB", 0),
	}
}

// Connect establishes connection to Redis
func Connect() error {
	config := GetConfig()

	Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	// Test connection
	ctx := context.Background()
	_, err := Client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis connected successfully")
	return nil
}

// Close closes Redis connection
func Close() error {
	if Client != nil {
		if err := Client.Close(); err != nil {
			return fmt.Errorf("failed to close Redis connection: %w", err)
		}
		logger.Info("Redis connection closed")
	}
	return nil
}

// GetClient returns Redis client instance
func GetClient() *redis.Client {
	return Client
}

// Set stores a key-value pair with expiration
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return Client.Set(ctx, key, jsonValue, expiration).Err()
}

// Get retrieves a value by key
func Get(ctx context.Context, key string, dest interface{}) error {
	val, err := Client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found")
		}
		return fmt.Errorf("failed to get value: %w", err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

// Delete removes a key
func Delete(ctx context.Context, key string) error {
	return Client.Del(ctx, key).Err()
}

// Exists checks if key exists
func Exists(ctx context.Context, key string) (bool, error) {
	result, err := Client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// SetNX sets a key only if it doesn't exist
func SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	return Client.SetNX(ctx, key, jsonValue, expiration).Result()
}

// Increment increments a key by 1
func Increment(ctx context.Context, key string) (int64, error) {
	return Client.Incr(ctx, key).Result()
}

// IncrementBy increments a key by specified value
func IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return Client.IncrBy(ctx, key, value).Result()
}

// Expire sets expiration for a key
func Expire(ctx context.Context, key string, expiration time.Duration) error {
	return Client.Expire(ctx, key, expiration).Err()
}

// TTL returns time to live for a key
func TTL(ctx context.Context, key string) (time.Duration, error) {
	return Client.TTL(ctx, key).Result()
}

// HealthCheck checks Redis connection health
func HealthCheck(ctx context.Context) error {
	if Client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	_, err := Client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("Redis ping failed: %w", err)
	}

	return nil
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets environment variable as integer with default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
