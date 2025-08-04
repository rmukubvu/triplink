package services

import (
	"log"
	"sync"
	"time"
	"triplink/backend/models"
)

// NotificationBatchService handles batching of notifications for efficient delivery
type NotificationBatchService struct {
	notificationService *NotificationService
	batchSize           int
	batchInterval       time.Duration
	batchQueue          []*models.Notification
	mutex               sync.Mutex
	stopChan            chan struct{}
	wg                  sync.WaitGroup
}

// NewNotificationBatchService creates a new notification batch service
func NewNotificationBatchService(notificationService *NotificationService, batchSize int, batchInterval time.Duration) *NotificationBatchService {
	return &NotificationBatchService{
		notificationService: notificationService,
		batchSize:           batchSize,
		batchInterval:       batchInterval,
		batchQueue:          make([]*models.Notification, 0),
		stopChan:            make(chan struct{}),
	}
}

// Start starts the batch processing
func (s *NotificationBatchService) Start() {
	s.wg.Add(1)
	go s.processBatches()
	log.Println("Notification batch service started")
}

// Stop stops the batch processing
func (s *NotificationBatchService) Stop() {
	close(s.stopChan)
	s.wg.Wait()
	log.Println("Notification batch service stopped")
}

// QueueNotification adds a notification to the batch queue
func (s *NotificationBatchService) QueueNotification(notification *models.Notification) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.batchQueue = append(s.batchQueue, notification)

	// If batch size is reached, process immediately
	if len(s.batchQueue) >= s.batchSize {
		go s.processBatch()
	}
}

// processBatches processes batches at regular intervals
func (s *NotificationBatchService) processBatches() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.batchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.processBatch()
		case <-s.stopChan:
			// Process any remaining notifications before stopping
			s.processBatch()
			return
		}
	}
}

// processBatch processes a batch of notifications
func (s *NotificationBatchService) processBatch() {
	s.mutex.Lock()
	if len(s.batchQueue) == 0 {
		s.mutex.Unlock()
		return
	}

	// Take current batch and clear queue
	batch := s.batchQueue
	s.batchQueue = make([]*models.Notification, 0)
	s.mutex.Unlock()

	// Process batch
	results, err := s.notificationService.BatchSendNotifications(batch)
	if err != nil {
		log.Printf("Error sending batch notifications: %v", err)

		// Re-queue failed notifications with exponential backoff
		// In a production system, you would implement retry logic with backoff
		// For now, we'll just log the error
	}

	// Log results
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	log.Printf("Batch processed: %d notifications, %d successful", len(batch), successCount)
}

// Global notification batch service instance
var notificationBatchServiceInstance *NotificationBatchService
var notificationBatchServiceOnce sync.Once

// GetNotificationBatchService returns the singleton notification batch service instance
// GetNotificationBatchService returns the singleton notification batch service instance
// It now requires a *NotificationService instance to be passed.
func GetNotificationBatchService(notificationService *NotificationService) *NotificationBatchService {
	notificationBatchServiceOnce.Do(func() {
		notificationBatchServiceInstance = NewNotificationBatchService(
			notificationService,
			10,            // Batch size
			5*time.Second, // Batch interval
		)
	})
	return notificationBatchServiceInstance
}
