package app

import (
	"math/rand"
	"time"

	"go_app/configs"
	"go_app/internal/repository"
	"go_app/internal/router"
	"go_app/internal/service"
	"go_app/internal/worker"
	"go_app/pkg/database"
	"go_app/pkg/logger"
	"go_app/pkg/redis"

	"github.com/gin-gonic/gin"
)

// App represents the application structure
type App struct {
	Config *configs.Config
	Router *gin.Engine
	Port   string
}

// New creates a new App instance
func New() *App {
	// Load configuration
	config := configs.Load()

	// Initialize random seed for OTP generation
	rand.Seed(time.Now().UnixNano())

	// Connect to database
	if err := database.Connect(); err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := database.Migrate(); err != nil {
		logger.Fatalf("Failed to run migrations: %v", err)
	}

	// Connect to Redis
	if err := redis.Connect(); err != nil {
		logger.Warnf("Failed to connect to Redis: %v", err)
	}

	// Initialize Gin router
	gin.SetMode(config.GinMode)
	r := gin.Default()

	// Setup routes
	router.SetupRoutes(r)

	// Start notification worker
	notificationService := service.NewNotificationService(
		repository.NewNotificationRepository(),
		repository.NewUserRepository(),
	)
	notificationWorker := worker.NewNotificationWorker(notificationService)
	go notificationWorker.Start()

	return &App{
		Config: config,
		Router: r,
		Port:   config.Server.Port,
	}
}

// SetPort sets the server port
func (a *App) SetPort(port string) {
	a.Port = port
}

// Run starts the application server
func (a *App) Run() error {
	port := a.Port
	if port == "" {
		port = a.Config.Server.Port
	}
	if port == "" {
		port = "8080"
	}

	logger.Infof("Starting server on port %s", port)
	return a.Router.Run(":" + port)
}
