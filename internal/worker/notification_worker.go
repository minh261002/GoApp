package worker

import (
	"go_app/internal/service"
	"go_app/pkg/logger"
	"time"
)

// NotificationWorker handles background notification processing
type NotificationWorker struct {
	notificationService service.NotificationService
	stopChan            chan bool
}

// NewNotificationWorker creates a new NotificationWorker
func NewNotificationWorker(notificationService service.NotificationService) *NotificationWorker {
	return &NotificationWorker{
		notificationService: notificationService,
		stopChan:            make(chan bool),
	}
}

// Start starts the notification worker
func (w *NotificationWorker) Start() {
	logger.Info("Starting notification worker...")

	ticker := time.NewTicker(30 * time.Second) // Process every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Process notification queue
			if err := w.notificationService.ProcessNotificationQueue(); err != nil {
				logger.Errorf("Failed to process notification queue: %v", err)
			}

			// Clean up old notifications
			if err := w.cleanupOldNotifications(); err != nil {
				logger.Errorf("Failed to cleanup old notifications: %v", err)
			}

		case <-w.stopChan:
			logger.Info("Stopping notification worker...")
			return
		}
	}
}

// Stop stops the notification worker
func (w *NotificationWorker) Stop() {
	w.stopChan <- true
}

// cleanupOldNotifications cleans up old notifications
func (w *NotificationWorker) cleanupOldNotifications() error {
	// Delete expired notifications
	if err := w.notificationService.DeleteExpiredNotifications(); err != nil {
		return err
	}

	// Archive notifications older than 30 days
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	if err := w.notificationService.ArchiveOldNotifications(thirtyDaysAgo); err != nil {
		return err
	}

	// Delete notifications older than 90 days
	ninetyDaysAgo := time.Now().AddDate(0, 0, -90)
	if err := w.notificationService.DeleteOldNotifications(ninetyDaysAgo); err != nil {
		return err
	}

	return nil
}
