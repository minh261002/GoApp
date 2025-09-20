package database

import (
	"fmt"
	"os"
	"time"

	"go_app/internal/model"
	"go_app/pkg/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

// Config cấu hình database
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	Charset  string
}

// GetConfig lấy cấu hình database từ environment variables
func GetConfig() *Config {
	return &Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "3306"),
		User:     getEnv("DB_USER", "root"),
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", "go_app_db"),
		Charset:  "utf8mb4",
	}
}

// Connect kết nối đến database
func Connect() error {
	config := GetConfig()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
		config.Charset,
	)

	// Cấu hình GORM logger
	var gormLog gormLogger.Interface
	if os.Getenv("GIN_MODE") == "release" {
		gormLog = gormLogger.Default.LogMode(gormLogger.Silent)
	} else {
		gormLog = gormLogger.Default.LogMode(gormLogger.Info)
	}

	// Kết nối database
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLog,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Cấu hình connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// SetMaxIdleConns thiết lập số lượng kết nối idle tối đa
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns thiết lập số lượng kết nối mở tối đa
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime thiết lập thời gian sống tối đa của kết nối
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db
	logger.Info("Database connected successfully")

	return nil
}

// Close đóng kết nối database
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return fmt.Errorf("failed to get underlying sql.DB: %w", err)
		}

		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("failed to close database connection: %w", err)
		}

		logger.Info("Database connection closed")
	}
	return nil
}

// GetDB trả về instance database
func GetDB() *gorm.DB {
	return DB
}

// Migrate runs migration for all models
func Migrate() error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}

	// Migrate all models
	models := []interface{}{
		&model.User{},
		&model.Session{},
		&model.OTP{},
		&model.Brand{},
		&model.Category{},
		&model.Product{},
		&model.ProductVariant{},
		&model.ProductAttribute{},
		&model.InventoryMovement{},
		&model.StockLevel{},
		&model.InventoryAdjustment{},
		&model.Permission{},
		&model.Role{},
		&model.RolePermission{},
		&model.UserPermission{},
		&model.PermissionLog{},
		// Add more models here as needed
	}

	for _, m := range models {
		if err := DB.AutoMigrate(m); err != nil {
			return fmt.Errorf("failed to migrate model %T: %w", m, err)
		}
	}

	logger.Info("Database migration completed successfully")
	return nil
}

// MigrateModels runs migration for specific models
func MigrateModels(models ...interface{}) error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}

	for _, model := range models {
		if err := DB.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate model %T: %w", model, err)
		}
	}

	logger.Info("Database migration completed successfully")
	return nil
}

// Transaction chạy một transaction
func Transaction(fn func(*gorm.DB) error) error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}

	return DB.Transaction(fn)
}

// HealthCheck kiểm tra trạng thái kết nối database
func HealthCheck() error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// getEnv lấy giá trị environment variable với giá trị mặc định
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
