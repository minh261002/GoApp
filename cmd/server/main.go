package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go_app/internal/app"
	"go_app/pkg/logger"
)

func main() {
	// Parse command line flags
	var (
		port = flag.String("port", "8080", "Server port")
		env  = flag.String("env", "development", "Environment (development, production)")
		help = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Set environment
	os.Setenv("GIN_MODE", *env)

	logger.Infof("Starting e-commerce server on port %s in %s mode", *port, *env)

	// Create and run the application
	application := app.New()

	// Set port from command line
	application.SetPort(*port)

	// Setup graceful shutdown
	setupGracefulShutdown(application)

	// Run the application
	if err := application.Run(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func showHelp() {
	fmt.Println("E-commerce Server")
	fmt.Println("Usage: go run cmd/server/main.go [options]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -port string")
	fmt.Println("        Server port (default \"8080\")")
	fmt.Println("  -env string")
	fmt.Println("        Environment: development, production (default \"development\")")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/server/main.go")
	fmt.Println("  go run cmd/server/main.go -port 3000")
	fmt.Println("  go run cmd/server/main.go -env production -port 80")
}

func setupGracefulShutdown(app *app.App) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logger.Info("Shutting down server gracefully...")

		// Perform cleanup here if needed
		// app.Cleanup()

		os.Exit(0)
	}()
}
