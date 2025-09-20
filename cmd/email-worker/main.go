package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go_app/internal/repository"
	"go_app/internal/service"
	"go_app/pkg/database"
	"go_app/pkg/logger"
)

func main() {
	// Parse command line flags
	var (
		interval = flag.Duration("interval", 30*time.Second, "Email processing interval")
	)
	flag.Parse()

	// Initialize logger (using default logger)

	logger.Infof("Starting email worker with interval: %v", *interval)

	// Connect to database
	db := database.GetDB()

	// Initialize repositories
	emailRepo := repository.NewEmailRepository(db)

	// Initialize services
	emailService := service.NewEmailService(emailRepo)

	// Create email worker
	emailWorker := NewEmailWorker(emailService, *interval)

	// Start worker
	go emailWorker.Start()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down email worker...")

	// Stop worker
	emailWorker.Stop()

	logger.Info("Email worker stopped")
}

// EmailWorker handles email queue processing
type EmailWorker struct {
	emailService service.EmailService
	interval     time.Duration
	stopChan     chan bool
}

// NewEmailWorker creates a new email worker
func NewEmailWorker(emailService service.EmailService, interval time.Duration) *EmailWorker {
	return &EmailWorker{
		emailService: emailService,
		interval:     interval,
		stopChan:     make(chan bool),
	}
}

// Start starts the email worker
func (w *EmailWorker) Start() {
	logger.Info("Email worker started")

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.processEmails()
		case <-w.stopChan:
			logger.Info("Email worker stopping...")
			return
		}
	}
}

// Stop stops the email worker
func (w *EmailWorker) Stop() {
	w.stopChan <- true
}

// processEmails processes pending emails
func (w *EmailWorker) processEmails() {
	logger.Debug("Processing email queue...")

	// Process email queue
	if err := w.emailService.ProcessEmailQueue(); err != nil {
		logger.Errorf("Failed to process email queue: %v", err)
		return
	}

	// Retry failed emails
	if err := w.emailService.RetryFailedEmails(); err != nil {
		logger.Errorf("Failed to retry failed emails: %v", err)
		return
	}

	logger.Debug("Email queue processed successfully")
}
