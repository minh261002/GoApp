package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go_app/internal/repository"
	"go_app/internal/service"
	"go_app/internal/worker"
	"go_app/pkg/database"
	"go_app/pkg/logger"
	"go_app/pkg/redis"
)

func main() {
	// Parse command line flags
	var (
		interval = flag.Duration("interval", 30*time.Second, "Worker processing interval")
		help     = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	logger.Infof("Starting notification worker with interval %v", *interval)

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

	// Initialize services
	notificationService := service.NewNotificationService(
		repository.NewNotificationRepository(),
		repository.NewUserRepository(),
	)

	// Create and start worker
	notificationWorker := worker.NewNotificationWorker(notificationService)

	// Setup graceful shutdown
	setupGracefulShutdown(notificationWorker)

	// Start worker
	notificationWorker.Start()
}

func showHelp() {
	fmt.Println("Notification Worker")
	fmt.Println("Usage: go run cmd/worker/main.go [options]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -interval duration")
	fmt.Println("        Worker processing interval (default 30s)")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/worker/main.go")
	fmt.Println("  go run cmd/worker/main.go -interval 1m")
	fmt.Println("  go run cmd/worker/main.go -interval 10s")
}

func setupGracefulShutdown(worker *worker.NotificationWorker) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logger.Info("Shutting down worker gracefully...")
		worker.Stop()
		os.Exit(0)
	}()
}
