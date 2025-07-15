package handlers

import (
	"fmt"
	"time"
	"triplink/backend/database"
	"triplink/backend/models"

	"github.com/gofiber/fiber/v2"
)

// CreateNotification @Summary Create a notification
// @Description Create a new notification for a user
// @Tags notifications
// @Accept json
// @Produce json
// @Param notification body models.Notification true "Notification data"
// @Success 201 {object} models.Notification
// @Router /notifications [post]
func CreateNotification(c *fiber.Ctx) error {
	var notification models.Notification

	if err := c.BodyParser(&notification); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	result := database.DB.Create(&notification)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not create notification",
		})
	}

	return c.Status(201).JSON(notification)
}

// GetUserNotifications @Summary Get user notifications
// @Description Get all notifications for a specific user
// @Tags notifications
// @Produce json
// @Param user_id path int true "User ID"
// @Param unread_only query boolean false "Show only unread notifications"
// @Success 200 {array} models.Notification
// @Router /users/{user_id}/notifications [get]
func GetUserNotifications(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	unreadOnly := c.Query("unread_only") == "true"

	query := database.DB.Where("user_id = ?", userID)
	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	var notifications []models.Notification
	result := query.Order("created_at DESC").Find(&notifications)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not fetch notifications",
		})
	}

	return c.JSON(notifications)
}

// MarkNotificationAsRead @Summary Mark notification as read
// @Description Mark a specific notification as read
// @Tags notifications
// @Param id path int true "Notification ID"
// @Success 200 {object} models.Notification
// @Router /notifications/{id}/read [put]
func MarkNotificationAsRead(c *fiber.Ctx) error {
	id := c.Params("id")
	var notification models.Notification

	if err := database.DB.First(&notification, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Notification not found",
		})
	}

	notification.IsRead = true
	database.DB.Save(&notification)

	return c.JSON(notification)
}

// MarkAllNotificationsAsRead @Summary Mark all notifications as read
// @Description Mark all notifications for a user as read
// @Tags notifications
// @Param user_id path int true "User ID"
// @Success 200 {object} map[string]interface{}
// @Router /users/{user_id}/notifications/read-all [put]
func MarkAllNotificationsAsRead(c *fiber.Ctx) error {
	userID := c.Params("user_id")

	result := database.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not update notifications",
		})
	}

	return c.JSON(fiber.Map{
		"message":       "All notifications marked as read",
		"updated_count": result.RowsAffected,
	})
}

// DeleteNotification @Summary Delete notification
// @Description Delete a specific notification
// @Tags notifications
// @Param id path int true "Notification ID"
// @Success 200 {object} map[string]string
// @Router /notifications/{id} [delete]
func DeleteNotification(c *fiber.Ctx) error {
	id := c.Params("id")
	var notification models.Notification

	if err := database.DB.First(&notification, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Notification not found",
		})
	}

	database.DB.Delete(&notification)
	return c.JSON(fiber.Map{
		"message": "Notification deleted successfully",
	})
}

// GetNotificationCounts @Summary Get notification counts
// @Description Get notification counts for a user (total and unread)
// @Tags notifications
// @Produce json
// @Param user_id path int true "User ID"
// @Success 200 {object} map[string]interface{}
// @Router /users/{user_id}/notifications/count [get]
func GetNotificationCounts(c *fiber.Ctx) error {
	userID := c.Params("user_id")

	var totalCount, unreadCount int64

	// Get total count
	database.DB.Model(&models.Notification{}).
		Where("user_id = ?", userID).
		Count(&totalCount)

	// Get unread count
	database.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&unreadCount)

	return c.JSON(map[string]interface{}{
		"total_count":  totalCount,
		"unread_count": unreadCount,
	})
}

// Helper functions for creating specific notification types

// CreateQuoteNotification creates a notification when a quote is received
func CreateQuoteNotification(shipperID uint, loadID uint, carrierName string) error {
	notification := models.Notification{
		UserID:    shipperID,
		Title:     "New Quote Received",
		Message:   "You have received a new quote from " + carrierName + " for your load",
		Type:      "QUOTE_RECEIVED",
		RelatedID: loadID,
	}
	return database.DB.Create(&notification).Error
}

// CreateBookingNotification creates a notification when a load is booked
func CreateBookingNotification(carrierID uint, loadID uint, shipperName string) error {
	notification := models.Notification{
		UserID:    carrierID,
		Title:     "Load Booked",
		Message:   shipperName + " has accepted your quote and booked the load",
		Type:      "LOAD_BOOKED",
		RelatedID: loadID,
	}
	return database.DB.Create(&notification).Error
}

// CreatePickupNotification creates a notification when pickup is scheduled
func CreatePickupNotification(shipperID uint, loadID uint) error {
	notification := models.Notification{
		UserID:    shipperID,
		Title:     "Pickup Scheduled",
		Message:   "Your load pickup has been scheduled. Check the details for timing.",
		Type:      "PICKUP_SCHEDULED",
		RelatedID: loadID,
	}
	return database.DB.Create(&notification).Error
}

// CreateDeliveryNotification creates a notification when load is delivered
func CreateDeliveryNotification(shipperID uint, loadID uint) error {
	notification := models.Notification{
		UserID:    shipperID,
		Title:     "Load Delivered",
		Message:   "Your load has been successfully delivered. Please review the carrier.",
		Type:      "LOAD_DELIVERED",
		RelatedID: loadID,
	}
	return database.DB.Create(&notification).Error
}

// Tracking-specific notification functions

// CreateTripDepartureNotification creates a notification when a trip departs
func CreateTripDepartureNotification(shipperID uint, tripID uint, carrierName string) error {
	notification := models.Notification{
		UserID:    shipperID,
		Title:     "Trip Departed",
		Message:   "Your shipment with " + carrierName + " has departed and is now in transit.",
		Type:      "TRIP_DEPARTED",
		RelatedID: tripID,
	}
	return database.DB.Create(&notification).Error
}

// CreateTripArrivalNotification creates a notification when a trip arrives at destination
func CreateTripArrivalNotification(shipperID uint, tripID uint, location string) error {
	notification := models.Notification{
		UserID:    shipperID,
		Title:     "Trip Arrived",
		Message:   "Your shipment has arrived at " + location + ". Delivery is being prepared.",
		Type:      "TRIP_ARRIVED",
		RelatedID: tripID,
	}
	return database.DB.Create(&notification).Error
}

// CreateDelayNotification creates a notification when a trip is delayed
func CreateDelayNotification(userID uint, tripID uint, delayMinutes int, reason string) error {
	message := fmt.Sprintf("Your shipment is delayed by %d minutes", delayMinutes)
	if reason != "" {
		message += " due to " + reason
	}

	notification := models.Notification{
		UserID:    userID,
		Title:     "Shipment Delayed",
		Message:   message,
		Type:      "TRIP_DELAYED",
		RelatedID: tripID,
	}
	return database.DB.Create(&notification).Error
}

// CreateETAUpdateNotification creates a notification when ETA is updated
func CreateETAUpdateNotification(shipperID uint, tripID uint, newETA time.Time) error {
	notification := models.Notification{
		UserID:    shipperID,
		Title:     "ETA Updated",
		Message:   "Your shipment's estimated arrival time has been updated to " + newETA.Format("Jan 2, 2006 at 3:04 PM"),
		Type:      "ETA_UPDATED",
		RelatedID: tripID,
	}
	return database.DB.Create(&notification).Error
}

// CreateLoadStatusChangeNotification creates a notification when load status changes
func CreateLoadStatusChangeNotification(shipperID uint, loadID uint, newStatus string, bookingRef string) error {
	statusMessages := map[string]string{
		"PICKUP_SCHEDULED": "Your load pickup has been scheduled",
		"PICKED_UP":        "Your load has been picked up and is now in transit",
		"IN_TRANSIT":       "Your load is currently in transit",
		"OUT_FOR_DELIVERY": "Your load is out for delivery",
		"DELIVERED":        "Your load has been successfully delivered",
		"EXCEPTION":        "There is an issue with your load that requires attention",
	}

	message := statusMessages[newStatus]
	if message == "" {
		message = "Your load status has been updated to " + newStatus
	}

	if bookingRef != "" {
		message += " (Ref: " + bookingRef + ")"
	}

	notification := models.Notification{
		UserID:    shipperID,
		Title:     "Load Status Update",
		Message:   message,
		Type:      "LOAD_STATUS_CHANGED",
		RelatedID: loadID,
	}
	return database.DB.Create(&notification).Error
}

// CreateLocationUpdateNotification creates a notification for significant location updates
func CreateLocationUpdateNotification(shipperID uint, tripID uint, location string) error {
	notification := models.Notification{
		UserID:    shipperID,
		Title:     "Location Update",
		Message:   "Your shipment has reached " + location,
		Type:      "LOCATION_UPDATE",
		RelatedID: tripID,
	}
	return database.DB.Create(&notification).Error
}

// NotifyAllShippersOnTrip sends notifications to all shippers with loads on a trip
func NotifyAllShippersOnTrip(tripID uint, title string, message string, notificationType string) error {
	// Get all loads on this trip
	var loads []models.Load
	if err := database.DB.Where("trip_id = ?", tripID).Find(&loads).Error; err != nil {
		return err
	}

	// Create notifications for each shipper
	for _, load := range loads {
		notification := models.Notification{
			UserID:    load.ShipperID,
			Title:     title,
			Message:   message,
			Type:      notificationType,
			RelatedID: tripID,
		}
		database.DB.Create(&notification)
	}

	return nil
}

// GetTrackingNotificationPreferences gets notification preferences for tracking events
func GetTrackingNotificationPreferences(c *fiber.Ctx) error {
	userID := c.Params("user_id")

	// For now, return default preferences - in a real system, this would be stored in database
	preferences := map[string]interface{}{
		"trip_departure":   true,
		"trip_arrival":     true,
		"delays":           true,
		"eta_updates":      true,
		"load_status":      true,
		"location_updates": false, // Usually disabled by default to avoid spam
		"email_enabled":    true,
		"push_enabled":     true,
	}

	return c.JSON(fiber.Map{
		"user_id":     userID,
		"preferences": preferences,
	})
}

// UpdateTrackingNotificationPreferences updates notification preferences for tracking events
func UpdateTrackingNotificationPreferences(c *fiber.Ctx) error {
	userID := c.Params("user_id")

	var preferences map[string]interface{}
	if err := c.BodyParser(&preferences); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot parse preferences data",
		})
	}

	// In a real system, you would save these preferences to the database
	// For now, just return success
	return c.JSON(fiber.Map{
		"message":     "Notification preferences updated successfully",
		"user_id":     userID,
		"preferences": preferences,
	})
}
