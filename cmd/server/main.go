package main

import (
	"go_app/internal/app"
)

func main() {
	// Create and run the application
	application := app.New()
	application.Run()
}
