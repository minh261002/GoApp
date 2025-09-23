package logger

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
}

var defaultLogger *Logger

func init() {
	defaultLogger = NewLogger()
}

// NewLogger tạo logger mới
func NewLogger() *Logger {
	logger := logrus.New()

	// Cấu hình log level
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "warn" // Giảm log mặc định từ info xuống warn
	}

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Cấu hình formatter
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})

	// Cấu hình output
	logFile := os.Getenv("LOG_FILE")
	if logFile != "" {
		// Tạo thư mục logs nếu chưa tồn tại
		if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
			logger.Warnf("Failed to create log directory: %v", err)
		} else {
			// Mở file log
			file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				logger.Warnf("Failed to open log file: %v", err)
			} else {
				// Ghi log vào cả file và console
				multiWriter := io.MultiWriter(os.Stdout, file)
				logger.SetOutput(multiWriter)
			}
		}
	} else {
		logger.SetOutput(os.Stdout)
	}

	return &Logger{logger}
}

// GetLogger trả về logger mặc định
func GetLogger() *Logger {
	return defaultLogger
}

// SetLogger thiết lập logger mặc định
func SetLogger(logger *Logger) {
	defaultLogger = logger
}

// Các method tiện ích
func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

func Warn(args ...interface{}) {
	defaultLogger.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	defaultLogger.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}

func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

// WithField tạo logger với field bổ sung
func WithField(key string, value interface{}) *logrus.Entry {
	return defaultLogger.WithField(key, value)
}

// WithFields tạo logger với nhiều fields
func WithFields(fields map[string]interface{}) *logrus.Entry {
	return defaultLogger.WithFields(fields)
}
