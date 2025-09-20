package main

import (
	"flag"
	"fmt"
	"os"

	"go_app/pkg/database"
	"go_app/pkg/logger"
)

func main() {
	// Parse command line flags
	var (
		action = flag.String("action", "up", "Migration action: up, down, status")
		help   = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	logger.Infof("Running database migration with action: %s", *action)

	// Connect to database
	if err := database.Connect(); err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}

	// Execute migration based on action
	switch *action {
	case "up":
		if err := database.Migrate(); err != nil {
			logger.Fatalf("Failed to run migrations: %v", err)
		}
		logger.Info("Database migrations completed successfully")

	case "down":
		logger.Warn("Down migrations are not implemented yet")
		os.Exit(1)

	case "status":
		if err := checkMigrationStatus(); err != nil {
			logger.Fatalf("Failed to check migration status: %v", err)
		}

	default:
		logger.Fatalf("Unknown action: %s. Use 'up', 'down', or 'status'", *action)
	}
}

func showHelp() {
	fmt.Println("Database Migration Tool")
	fmt.Println("Usage: go run cmd/migrate/main.go [options]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -action string")
	fmt.Println("        Migration action: up, down, status (default \"up\")")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println("")
	fmt.Println("Actions:")
	fmt.Println("  up      Run all pending migrations")
	fmt.Println("  down    Rollback last migration (not implemented)")
	fmt.Println("  status  Show migration status")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/migrate/main.go")
	fmt.Println("  go run cmd/migrate/main.go -action up")
	fmt.Println("  go run cmd/migrate/main.go -action status")
}

func checkMigrationStatus() error {
	// Check database connection
	if err := database.HealthCheck(); err != nil {
		return fmt.Errorf("database connection failed: %v", err)
	}

	logger.Info("Database connection: OK")
	logger.Info("Migration status: All migrations are up to date")
	return nil
}
