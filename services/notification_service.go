package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
	"triplink/backend/models"
	
	"gorm.io/gorm"
)

// NotificationService handles notification operations
type NotificationService struct {
	// Mutex for thread safety
	mu sync.Mutex
	// Database connection
	db *gorm.DB
	// Map of user IDs to their device tokens
	deviceTokens map[uint][]DeviceToken
	// Delivery providers
	providers []NotificationProvider
}

// DeviceToken represents a user's device token for push notifications
type DeviceToken struct {
	Token      string    `json:"token"`
	DeviceType string    `json:"device_type"` // ios, android
	CreatedAt  time.Time `json:"created_at"`
	LastUsed   time.Time `json:"last_used"`
}

// NotificationProvider interface for different push notification services
type NotificationProvider interface {
	SendNotification(tokens []string, notification *models.Notification) error
	BatchSendNotifications(notifications []NotificationBatch) error
	Name() string
}

// NotificationBatch represents a batch of notifications to be sent
type NotificationBatch struct {
	Tokens       []string             `json:"tokens"`
	Notification *models.Notification `json:"notification"`
}

// NotificationDeliveryResult represents the result of a notification delivery attempt
type NotificationDeliveryResult struct {
	NotificationID uint      `json:"notification_id"`
	UserID         uint      `json:"user_id"`
	Success        bool      `json:"success"`
	Provider       string    `json:"provider"`
	SentAt         time.Time `json:"sent_at"`
	Error          string    `json:"error,omitempty"`
}

// NewNotificationService creates a new NotificationService instance
func NewNotificationService(db *gorm.DB) *NotificationService {
	s := &NotificationService{
		mu:           sync.Mutex{},
		db:           db,
		deviceTokens: make(map[uint][]DeviceToken),
		providers:    []NotificationProvider{},
	}
	// Load device tokens from database
	s.loadDeviceTokens()
	return s
}

// loadDeviceTokens loads all device tokens from the database
func (s *NotificationService) loadDeviceTokens() {
	var tokenRecords []models.NotificationToken
	if err := s.db.Find(&tokenRecords).Error; err != nil {
		log.Printf("Error loading device tokens: %v", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, record := range tokenRecords {
		if _, exists := s.deviceTokens[record.UserID]; !exists {
			s.deviceTokens[record.UserID] = []DeviceToken{}
		}

		s.deviceTokens[record.UserID] = append(s.deviceTokens[record.UserID], DeviceToken{
			Token:      record.Token,
			DeviceType: record.DeviceType,
			CreatedAt:  record.CreatedAt,
			LastUsed:   record.LastUsed,
		})
	}

	log.Printf("Loaded %d device tokens for %d users", len(tokenRecords), len(s.deviceTokens))
}

// RegisterProvider adds a notification provider to the service
func (s *NotificationService) RegisterProvider(provider NotificationProvider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.providers = append(s.providers, provider)
	log.Printf("Registered notification provider: %s", provider.Name())
}

// RegisterDeviceToken registers a device token for a user
func (s *NotificationService) RegisterDeviceToken(userID uint, token string, deviceType string) error {
	if token == "" {
		return errors.New("token cannot be empty")
	}

	// Check if token already exists in database
	var existingToken models.NotificationToken
	result := s.db.Where("token = ? AND user_id = ?", token, userID).First(&existingToken)

	if result.Error == nil {
		// Token exists, update last used
		existingToken.LastUsed = time.Now()
		if err := s.db.Save(&existingToken).Error; err != nil {
			return fmt.Errorf("failed to update existing token: %w", err)
		}
	} else {
		// Token doesn't exist, create new
		newToken := models.NotificationToken{
			UserID:     userID,
			Token:      token,
			DeviceType: deviceType,
			LastUsed:   time.Now(),
		}

		if err := s.db.Create(&newToken).Error; err != nil {
			return fmt.Errorf("failed to create token: %w", err)
		}
	}

	// Update in-memory cache
	s.mu.Lock()
	defer s.mu.Unlock()

	newDeviceToken := DeviceToken{
		Token:      token,
		DeviceType: deviceType,
		CreatedAt:  time.Now(),
		LastUsed:   time.Now(),
	}

	// Check if we already have this token
	for i, dt := range s.deviceTokens[userID] {
		if dt.Token == token {
			// Update existing token
			s.deviceTokens[userID][i].LastUsed = time.Now()
			return nil
		}
	}

	// Add new token
	if _, exists := s.deviceTokens[userID]; !exists {
		s.deviceTokens[userID] = []DeviceToken{}
	}
	s.deviceTokens[userID] = append(s.deviceTokens[userID], newDeviceToken)

	return nil
}

// UnregisterDeviceToken removes a device token for a user
func (s *NotificationService) UnregisterDeviceToken(userID uint, token string) error {
	// Remove from database
	if err := s.db.Where("token = ? AND user_id = ?", token, userID).Delete(&models.NotificationToken{}).Error; err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}

	// Remove from in-memory cache
	s.mu.Lock()
	defer s.mu.Unlock()

	tokens, exists := s.deviceTokens[userID]
	if !exists {
		return nil
	}

	for i, dt := range tokens {
		if dt.Token == token {
			// Remove token
			s.deviceTokens[userID] = append(tokens[:i], tokens[i+1:]...)
			break
		}
	}

	return nil
}

// GetUserDeviceTokens returns all device tokens for a user
func (s *NotificationService) GetUserDeviceTokens(userID uint) []DeviceToken {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokens, exists := s.deviceTokens[userID]
	if !exists {
		return []DeviceToken{}
	}

	// Return a copy to prevent modification of internal state
	result := make([]DeviceToken, len(tokens))
	copy(result, tokens)
	return result
}

// SendNotification sends a notification to a user
func (s *NotificationService) SendNotification(notification *models.Notification) (*NotificationDeliveryResult, error) {
	// Get user's device tokens
	tokens := s.GetUserDeviceTokens(notification.UserID)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("no device tokens found for user %d", notification.UserID)
	}

	// Extract just the token strings
	tokenStrings := make([]string, len(tokens))
	for i, t := range tokens {
		tokenStrings[i] = t.Token
	}

	// Try each provider until one succeeds
	var lastError error
	for _, provider := range s.providers {
		err := provider.SendNotification(tokenStrings, notification)
		if err == nil {
			// Success, record delivery
			result := &NotificationDeliveryResult{
				NotificationID: notification.ID,
				UserID:         notification.UserID,
				Success:        true,
				Provider:       provider.Name(),
				SentAt:         time.Now(),
			}

			// Record delivery in database
			s.recordDelivery(result)

			return result, nil
		}
		lastError = err
	}

	// All providers failed
	result := &NotificationDeliveryResult{
		NotificationID: notification.ID,
		UserID:         notification.UserID,
		Success:        false,
		SentAt:         time.Now(),
		Error:          lastError.Error(),
	}

	// Record failed delivery
	s.recordDelivery(result)

	return result, lastError
}

// BatchSendNotifications sends multiple notifications efficiently
func (s *NotificationService) BatchSendNotifications(notifications []*models.Notification) ([]*NotificationDeliveryResult, error) {
	if len(notifications) == 0 {
		return []*NotificationDeliveryResult{}, nil
	}

	// Group notifications by user
	userNotifications := make(map[uint][]*models.Notification)
	for _, notification := range notifications {
		userNotifications[notification.UserID] = append(userNotifications[notification.UserID], notification)
	}

	// Prepare batches
	var batches []NotificationBatch
	for userID, userNotifs := range userNotifications {
		tokens := s.GetUserDeviceTokens(userID)
		if len(tokens) == 0 {
			continue
		}

		// Extract token strings
		tokenStrings := make([]string, len(tokens))
		for i, t := range tokens {
			tokenStrings[i] = t.Token
		}

		// Create a batch for each notification
		for _, notification := range userNotifs {
			batches = append(batches, NotificationBatch{
				Tokens:       tokenStrings,
				Notification: notification,
			})
		}
	}

	// Send batches through providers
	var results []*NotificationDeliveryResult
	var lastError error

	for _, provider := range s.providers {
		err := provider.BatchSendNotifications(batches)
		if err == nil {
			// Success, record deliveries
			for _, batch := range batches {
				result := &NotificationDeliveryResult{
					NotificationID: batch.Notification.ID,
					UserID:         batch.Notification.UserID,
					Success:        true,
					Provider:       provider.Name(),
					SentAt:         time.Now(),
				}
				results = append(results, result)
				s.recordDelivery(result)
			}
			return results, nil
		}
		lastError = err
	}

	// All providers failed
	for _, batch := range batches {
		result := &NotificationDeliveryResult{
			NotificationID: batch.Notification.ID,
			UserID:         batch.Notification.UserID,
			Success:        false,
			SentAt:         time.Now(),
			Error:          lastError.Error(),
		}
		results = append(results, result)
		s.recordDelivery(result)
	}

	return results, lastError
}

// recordDelivery records a notification delivery attempt in the database
func (s *NotificationService) recordDelivery(result *NotificationDeliveryResult) {
	delivery := models.NotificationDelivery{
		NotificationID: result.NotificationID,
		UserID:         result.UserID,
		Success:        result.Success,
		Provider:       result.Provider,
		SentAt:         result.SentAt,
		Error:          result.Error,
	}

	if err := s.db.Create(&delivery).Error; err != nil {
		log.Printf("Error recording notification delivery: %v", err)
	}
}

// GetUserNotificationPreferences gets a user's notification preferences
func (s *NotificationService) GetUserNotificationPreferences(userID uint) (*models.NotificationPreferences, error) {
	var preferences models.NotificationPreferences
	result := s.db.Where("user_id = ?", userID).First(&preferences)

	if result.Error != nil {
		// If not found, create default preferences
		if result.RowsAffected == 0 {
			preferences = models.NotificationPreferences{
				UserID:          userID,
				TripDeparture:   true,
				TripArrival:     true,
				Delays:          true,
				ETAUpdates:      true,
				LoadStatus:      true,
				LocationUpdates: false,
				EmailEnabled:    true,
				PushEnabled:     true,
			}

			if err := s.db.Create(&preferences).Error; err != nil {
				return nil, fmt.Errorf("failed to create default preferences: %w", err)
			}

			return &preferences, nil
		}

		return nil, fmt.Errorf("failed to get notification preferences: %w", result.Error)
	}

	return &preferences, nil
}

// UpdateUserNotificationPreferences updates a user's notification preferences
func (s *NotificationService) UpdateUserNotificationPreferences(userID uint, preferences *models.NotificationPreferences) error {
	// Ensure the user ID matches
	preferences.UserID = userID

	// Check if preferences exist
	var count int64
	s.db.Model(&models.NotificationPreferences{}).Where("user_id = ?", userID).Count(&count)

	if count == 0 {
		// Create new preferences
		if err := s.db.Create(preferences).Error; err != nil {
			return fmt.Errorf("failed to create preferences: %w", err)
		}
	} else {
		// Update existing preferences
		if err := s.db.Model(&models.NotificationPreferences{}).Where("user_id = ?", userID).Updates(preferences).Error; err != nil {
			return fmt.Errorf("failed to update preferences: %w", err)
		}
	}

	return nil
}

// ShouldSendNotification checks if a notification should be sent based on user preferences
func (s *NotificationService) ShouldSendNotification(userID uint, notificationType string) (bool, error) {
	preferences, err := s.GetUserNotificationPreferences(userID)
	if err != nil {
		return false, err
	}

	// If push notifications are disabled, don't send
	if !preferences.PushEnabled {
		return false, nil
	}

	// Check specific notification type
	switch notificationType {
	case "TRIP_DEPARTED", "TRIP_STATUS_CHANGE":
		return preferences.TripDeparture, nil
	case "TRIP_ARRIVED":
		return preferences.TripArrival, nil
	case "TRIP_DELAYED", "DELAY_ALERT":
		return preferences.Delays, nil
	case "ETA_UPDATED":
		return preferences.ETAUpdates, nil
	case "LOAD_STATUS_CHANGED", "LOAD_BOOKED", "PICKUP_SCHEDULED", "LOAD_DELIVERED":
		return preferences.LoadStatus, nil
	case "LOCATION_UPDATE":
		return preferences.LocationUpdates, nil
	default:
		// For unknown types, default to true
		return true, nil
	}
}

// CreateNotificationWithDelivery creates a notification and delivers it
func (s *NotificationService) CreateNotificationWithDelivery(notification *models.Notification) (*models.Notification, *NotificationDeliveryResult, error) {
	// First check if we should send this notification based on user preferences
	shouldSend, err := s.ShouldSendNotification(notification.UserID, notification.Type)
	if err != nil {
		return nil, nil, fmt.Errorf("error checking notification preferences: %w", err)
	}

	if !shouldSend {
		// Skip this notification based on user preferences
		return notification, nil, nil
	}

	// Create the notification in the database
	if err := s.db.Create(notification).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// Send the notification
	deliveryResult, err := s.SendNotification(notification)
	if err != nil {
		log.Printf("Failed to deliver notification %d: %v", notification.ID, err)
		// We still return the notification even if delivery failed
		return notification, deliveryResult, err
	}

	return notification, deliveryResult, nil
}

// PrepareNotificationPayload prepares the payload for a push notification
func (s *NotificationService) PrepareNotificationPayload(notification *models.Notification) map[string]interface{} {
	// Basic payload with notification content
	payload := map[string]interface{}{
		"id":        notification.ID,
		"userId":    notification.UserID,
		"title":     notification.Title,
		"message":   notification.Message,
		"type":      notification.Type,
		"createdAt": notification.CreatedAt,
	}

	// Add related entity information if available
	if notification.RelatedID > 0 {
		payload["relatedId"] = notification.RelatedID

		// Determine related entity type based on notification type
		switch notification.Type {
		case "TRIP_DEPARTED", "TRIP_ARRIVED", "TRIP_DELAYED", "ETA_UPDATED", "TRIP_STATUS_CHANGE":
			payload["relatedEntityType"] = "trip"
		case "LOAD_STATUS_CHANGED", "LOAD_BOOKED", "PICKUP_SCHEDULED", "LOAD_DELIVERED":
			payload["relatedEntityType"] = "load"
		case "QUOTE_RECEIVED":
			payload["relatedEntityType"] = "quote"
		case "NEW_MESSAGE":
			payload["relatedEntityType"] = "conversation"
		}
	}

	return payload
}

// SerializeNotificationPayload converts a notification payload to JSON
func (s *NotificationService) SerializeNotificationPayload(payload map[string]interface{}) (string, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to serialize notification payload: %w", err)
	}
	return string(jsonData), nil
}
