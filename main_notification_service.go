package main

import (
	"log"
	"time"
	"triplink/backend/services"
)

// initNotificationService initializes the notification service and registers providers
func initNotificationService() {
	// Get notification service instance
	notificationService := services.GetNotificationService()

	// Register Expo notification provider
	expoProvider := services.NewExpoNotificationProvider()
	notificationService.RegisterProvider(expoProvider)

	// Initialize and start notification batch service
	batchService := services.GetNotificationBatchService()
	batchService.Start()

	// Schedule periodic delay checks for notifications
	go scheduleDelayChecks()

	log.Println("Notification service initialized with Expo provider and batch processing")
}

// scheduleDelayChecks schedules periodic checks for delays
func scheduleDelayChecks() {
	triggerService := services.NewNotificationTriggerService()
	ticker := time.NewTicker(15 * time.Minute) // Check every 15 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			triggerService.ScheduleDelayCheck()
		}
	}
}
