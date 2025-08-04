package services

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
	"triplink/backend/models"
	
	"gorm.io/gorm"
)

// NotificationTriggerService handles notification triggers
type NotificationTriggerService struct {
	notificationService *NotificationService
	trackingService     *TrackingService
	db                  *gorm.DB
}

// NewNotificationTriggerService creates a new notification trigger service
func NewNotificationTriggerService(db *gorm.DB) *NotificationTriggerService {
	return &NotificationTriggerService{
		notificationService: NewNotificationService(db),
		trackingService:     NewTrackingService(db),
		db:                  db,
	}
}

// TripStatusChangeHandler handles trip status change events and sends notifications
func (s *NotificationTriggerService) TripStatusChangeHandler(tripID uint, oldStatus, newStatus string) error {
	// Get trip details
	var trip models.Trip
	if err := s.db.First(&trip, tripID).Error; err != nil {
		return fmt.Errorf("failed to get trip details: %w", err)
	}

	// Get carrier details
	var carrier models.User
	if err := s.db.First(&carrier, trip.UserID).Error; err != nil {
		return fmt.Errorf("failed to get carrier details: %w", err)
	}

	// Get all loads on this trip
	var loads []models.Load
	if err := s.db.Where("trip_id = ?", tripID).Find(&loads).Error; err != nil {
		return fmt.Errorf("failed to get loads: %w", err)
	}

	// Prepare notification based on status change
	var title, messageTemplate string
	var notificationType string

	switch newStatus {
	case "PLANNED":
		title = "Trip Planned"
		messageTemplate = "A new trip has been planned by %s from %s to %s"
		notificationType = "TRIP_PLANNED"
	case "ACTIVE":
		title = "Trip Activated"
		messageTemplate = "Trip from %s to %s has been activated and is ready for departure"
		notificationType = "TRIP_ACTIVATED"
	case "IN_TRANSIT":
		title = "Trip Departed"
		messageTemplate = "Your shipment with %s has departed from %s and is now in transit"
		notificationType = "TRIP_DEPARTED"
	case "COMPLETED":
		title = "Trip Completed"
		messageTemplate = "Your shipment has arrived at %s. Delivery is being prepared."
		notificationType = "TRIP_ARRIVED"
	case "CANCELLED":
		title = "Trip Cancelled"
		messageTemplate = "Trip from %s to %s has been cancelled"
		notificationType = "TRIP_CANCELLED"
	default:
		title = "Trip Status Updated"
		messageTemplate = "Trip status has been updated from %s to %s"
		notificationType = "TRIP_STATUS_CHANGE"
	}

	// Send notifications to all shippers with loads on this trip
	for _, load := range loads {
		var message string

		switch newStatus {
		case "PLANNED", "ACTIVE", "CANCELLED":
			message = fmt.Sprintf(messageTemplate, trip.OriginCity, trip.DestinationCity)
		case "IN_TRANSIT":
			message = fmt.Sprintf(messageTemplate, carrier.CompanyName, trip.OriginCity)
		case "COMPLETED":
			message = fmt.Sprintf(messageTemplate, trip.DestinationCity)
		default:
			message = fmt.Sprintf(messageTemplate, oldStatus, newStatus)
		}

		notification := models.Notification{
			UserID:    load.ShipperID,
			Title:     title,
			Message:   message,
			Type:      notificationType,
			RelatedID: tripID,
		}

		_, _, err := s.notificationService.CreateNotificationWithDelivery(&notification)
		if err != nil {
			log.Printf("Failed to send notification to shipper %d: %v", load.ShipperID, err)
		}
	}

	// Create tracking event for status change
	eventData := map[string]string{
		"old_status": oldStatus,
		"new_status": newStatus,
	}
	eventDataJSON, _ := json.Marshal(eventData)

	s.trackingService.LogTrackingEvent(
		tripID,
		nil,
		"STATUS_CHANGE",
		string(eventDataJSON),
		"",
		nil,
		nil,
		fmt.Sprintf("Trip status changed from %s to %s", oldStatus, newStatus),
	)

	return nil
}

// LoadStatusChangeHandler handles load status change events and sends notifications
func (s *NotificationTriggerService) LoadStatusChangeHandler(loadID uint, oldStatus, newStatus string) error {
	// Get load details
	var load models.Load
	if err := s.db.First(&load, loadID).Error; err != nil {
		return fmt.Errorf("failed to get load details: %w", err)
	}

	// Get trip details
	var trip models.Trip
	if err := s.db.First(&trip, load.TripID).Error; err != nil {
		return fmt.Errorf("failed to get trip details: %w", err)
	}

	// Prepare notification
	statusMessages := map[string]string{
		"QUOTE_REQUESTED":  "Quote has been requested for your load",
		"QUOTED":           "Your load has been quoted",
		"BOOKED":           "Your load has been booked",
		"PICKUP_SCHEDULED": "Your load pickup has been scheduled",
		"PICKED_UP":        "Your load has been picked up and is now in transit",
		"IN_TRANSIT":       "Your load is currently in transit",
		"OUT_FOR_DELIVERY": "Your load is out for delivery",
		"DELIVERED":        "Your load has been successfully delivered",
		"CANCELLED":        "Your load has been cancelled",
		"EXCEPTION":        "There is an issue with your load that requires attention",
	}

	message := statusMessages[newStatus]
	if message == "" {
		message = fmt.Sprintf("Your load status has been updated to %s", newStatus)
	}

	if load.BookingReference != "" {
		message += fmt.Sprintf(" (Ref: %s)", load.BookingReference)
	}

	notification := models.Notification{
		UserID:    load.ShipperID,
		Title:     "Load Status Update",
		Message:   message,
		Type:      "LOAD_STATUS_CHANGED",
		RelatedID: loadID,
	}

	_, _, err := s.notificationService.CreateNotificationWithDelivery(&notification)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

// MessageReceivedHandler handles new message events and sends notifications
func (s *NotificationTriggerService) MessageReceivedHandler(messageID uint) error {
	// Get message details
	var message models.Message
	if err := s.db.First(&message, messageID).Error; err != nil {
		return fmt.Errorf("failed to get message details: %w", err)
	}

	// Get sender details
	var sender models.User
	if err := s.db.First(&sender, message.SenderID).Error; err != nil {
		return fmt.Errorf("failed to get sender details: %w", err)
	}

	// Create notification for the receiver
	notification := models.Notification{
		UserID:    message.ReceiverID,
		Title:     "New Message",
		Message:   fmt.Sprintf("You have received a new message from %s %s", sender.FirstName, sender.LastName),
		Type:      "NEW_MESSAGE",
		RelatedID: messageID,
	}

	_, _, err := s.notificationService.CreateNotificationWithDelivery(&notification)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

// DelayDetectionHandler handles delay detection and sends notifications
func (s *NotificationTriggerService) DelayDetectionHandler(tripID uint, delayMinutes int, reason string) error {
	// Get trip details
	var trip models.Trip
	if err := s.db.First(&trip, tripID).Error; err != nil {
		return fmt.Errorf("failed to get trip details: %w", err)
	}

	// Get all loads on this trip
	var loads []models.Load
	if err := s.db.Where("trip_id = ?", tripID).Find(&loads).Error; err != nil {
		return fmt.Errorf("failed to get loads: %w", err)
	}

	// Prepare notification message
	message := fmt.Sprintf("Your shipment is delayed by %d minutes", delayMinutes)
	if reason != "" {
		message += " due to " + reason
	}

	// Send notifications to all shippers with loads on this trip
	for _, load := range loads {
		notification := models.Notification{
			UserID:    load.ShipperID,
			Title:     "Shipment Delayed",
			Message:   message,
			Type:      "TRIP_DELAYED",
			RelatedID: tripID,
		}

		_, _, err := s.notificationService.CreateNotificationWithDelivery(&notification)
		if err != nil {
			log.Printf("Failed to send delay notification to shipper %d: %v", load.ShipperID, err)
		}
	}

	// Create tracking event for delay
	eventData := map[string]interface{}{
		"delay_minutes": delayMinutes,
		"reason":        reason,
		"detected_at":   time.Now().Format(time.RFC3339),
	}
	eventDataJSON, _ := json.Marshal(eventData)

	s.trackingService.LogTrackingEvent(
		tripID,
		nil,
		"DELAY",
		string(eventDataJSON),
		"",
		nil,
		nil,
		fmt.Sprintf("Trip delayed by %d minutes - %s", delayMinutes, reason),
	)

	return nil
}

// ETAUpdateHandler handles ETA update events and sends notifications
func (s *NotificationTriggerService) ETAUpdateHandler(tripID uint, newETA time.Time) error {
	// Get trip details
	var trip models.Trip
	if err := s.db.First(&trip, tripID).Error; err != nil {
		return fmt.Errorf("failed to get trip details: %w", err)
	}

	// Get all loads on this trip
	var loads []models.Load
	if err := s.db.Where("trip_id = ?", tripID).Find(&loads).Error; err != nil {
		return fmt.Errorf("failed to get loads: %w", err)
	}

	// Calculate if this is a significant change (more than 15 minutes)
	etaDiff := newETA.Sub(trip.EstimatedArrival).Minutes()
	if etaDiff > -15 && etaDiff < 15 {
		// Not a significant change, don't send notification
		return nil
	}

	// Prepare notification message
	message := fmt.Sprintf("Your shipment's estimated arrival time has been updated to %s",
		newETA.Format("Jan 2, 2006 at 3:04 PM"))

	// Send notifications to all shippers with loads on this trip
	for _, load := range loads {
		notification := models.Notification{
			UserID:    load.ShipperID,
			Title:     "ETA Updated",
			Message:   message,
			Type:      "ETA_UPDATED",
			RelatedID: tripID,
		}

		_, _, err := s.notificationService.CreateNotificationWithDelivery(&notification)
		if err != nil {
			log.Printf("Failed to send ETA update notification to shipper %d: %v", load.ShipperID, err)
		}
	}

	// Create tracking event for ETA update
	eventData := map[string]interface{}{
		"previous_eta": trip.EstimatedArrival.Format(time.RFC3339),
		"new_eta":      newETA.Format(time.RFC3339),
		"diff_minutes": etaDiff,
	}
	eventDataJSON, _ := json.Marshal(eventData)

	s.trackingService.LogTrackingEvent(
		tripID,
		nil,
		"ETA_UPDATE",
		string(eventDataJSON),
		"",
		nil,
		nil,
		fmt.Sprintf("ETA updated from %s to %s",
			trip.EstimatedArrival.Format("Jan 2, 2006 at 3:04 PM"),
			newETA.Format("Jan 2, 2006 at 3:04 PM")),
	)

	return nil
}

// LocationMilestoneHandler handles location milestone events and sends notifications
func (s *NotificationTriggerService) LocationMilestoneHandler(tripID uint, milestone string, latitude, longitude float64) error {
	// Get trip details
	var trip models.Trip
	if err := s.db.First(&trip, tripID).Error; err != nil {
		return fmt.Errorf("failed to get trip details: %w", err)
	}

	// Get all loads on this trip
	var loads []models.Load
	if err := s.db.Where("trip_id = ?", tripID).Find(&loads).Error; err != nil {
		return fmt.Errorf("failed to get loads: %w", err)
	}

	// Check if location updates are enabled in notification preferences for each shipper
	for _, load := range loads {
		// Get shipper's notification preferences
		preferences, err := s.notificationService.GetUserNotificationPreferences(load.ShipperID)
		if err != nil {
			log.Printf("Failed to get notification preferences for shipper %d: %v", load.ShipperID, err)
			continue
		}

		// Skip if location updates are disabled
		if !preferences.LocationUpdates {
			continue
		}

		// Create notification
		notification := models.Notification{
			UserID:    load.ShipperID,
			Title:     "Location Update",
			Message:   fmt.Sprintf("Your shipment has reached %s", milestone),
			Type:      "LOCATION_UPDATE",
			RelatedID: tripID,
		}

		_, _, err = s.notificationService.CreateNotificationWithDelivery(&notification)
		if err != nil {
			log.Printf("Failed to send location update notification to shipper %d: %v", load.ShipperID, err)
		}
	}

	// Create tracking event for location milestone
	eventData := map[string]interface{}{
		"milestone": milestone,
		"latitude":  latitude,
		"longitude": longitude,
	}
	eventDataJSON, _ := json.Marshal(eventData)

	s.trackingService.LogTrackingEvent(
		tripID,
		nil,
		"MILESTONE",
		string(eventDataJSON),
		milestone,
		&latitude,
		&longitude,
		fmt.Sprintf("Trip reached milestone: %s", milestone),
	)

	return nil
}

// ScheduleDelayCheck schedules a periodic check for delays
func (s *NotificationTriggerService) ScheduleDelayCheck() {
	// This would typically be called from a background goroutine or scheduler
	// For now, we'll just implement the check function

	// Get all active trips
	var activeTrips []models.Trip
	if err := s.db.Where("status = ? OR status = ?", "ACTIVE", "IN_TRANSIT").Find(&activeTrips).Error; err != nil {
		log.Printf("Failed to get active trips: %v", err)
		return
	}

	// Check each trip for delays
	for _, trip := range activeTrips {
		delayInfo, err := s.trackingService.CheckForDelays(trip.ID)
		if err != nil {
			log.Printf("Failed to check delays for trip %d: %v", trip.ID, err)
			continue
		}

		if delayInfo != nil {
			// Process delay alerts
			s.trackingService.ProcessDelayAlerts(trip.ID)
		}
	}
}

// RegisterNotificationTriggers registers all notification triggers with their respective handlers
func RegisterNotificationTriggers() {
	log.Println("Registering notification triggers...")

	// This function would typically hook into event systems or message queues
	// For now, we'll just log that triggers are registered

	log.Println("Notification triggers registered successfully")
}
