package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"triplink/backend/models"
	"gorm.io/gorm"
)

// ExpoNotificationProvider implements the NotificationProvider interface for Expo Push Notifications
type ExpoNotificationProvider struct {
	ExpoAPIURL string
	HTTPClient *http.Client
	db         *gorm.DB
}

// ExpoNotificationRequest represents a request to the Expo push notification API
type ExpoNotificationRequest struct {
	To      []string               `json:"to"`
	Title   string                 `json:"title"`
	Body    string                 `json:"body"`
	Data    map[string]interface{} `json:"data"`
	Sound   string                 `json:"sound,omitempty"`
	Badge   int                    `json:"badge,omitempty"`
	TTL     int                    `json:"ttl,omitempty"`
	Channel string                 `json:"channelId,omitempty"`
}

// ExpoNotificationResponse represents a response from the Expo push notification API
type ExpoNotificationResponse struct {
	Data []struct {
		Status  string `json:"status"`
		ID      string `json:"id"`
		Message string `json:"message,omitempty"`
	} `json:"data"`
	Errors []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

// NewExpoNotificationProvider creates a new Expo notification provider
func NewExpoNotificationProvider(db *gorm.DB) *ExpoNotificationProvider {
	return &ExpoNotificationProvider{
		ExpoAPIURL: "https://exp.host/--/api/v2/push/send",
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		db: db,
	}
}

// Name returns the name of the provider
func (p *ExpoNotificationProvider) Name() string {
	return "expo"
}

// SendNotification sends a notification to the specified tokens
func (p *ExpoNotificationProvider) SendNotification(tokens []string, notification *models.Notification) error {
	if len(tokens) == 0 {
		return fmt.Errorf("no tokens provided")
	}

	// Get notification service
	notificationService := NewNotificationService(p.db)

	// Prepare notification payload
	payload := notificationService.PrepareNotificationPayload(notification)

	// Create request body
	requestBody := ExpoNotificationRequest{
		To:    tokens,
		Title: notification.Title,
		Body:  notification.Message,
		Data:  payload,
		Sound: "default",
		TTL:   3600, // 1 hour
	}

	// Set appropriate channel based on notification type
	switch notification.Type {
	case "TRIP_DEPARTED", "TRIP_ARRIVED", "TRIP_STATUS_CHANGE":
		requestBody.Channel = "trip-updates"
	case "NEW_MESSAGE":
		requestBody.Channel = "messages"
	case "TRIP_DELAYED", "ETA_UPDATED":
		requestBody.Channel = "alerts"
	default:
		requestBody.Channel = "default"
	}

	// Serialize request body
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to serialize notification request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", p.ExpoAPIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Send request
	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send notification request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var response ExpoNotificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to parse notification response: %w", err)
	}

	// Check for errors
	if len(response.Errors) > 0 {
		return fmt.Errorf("expo API error: %s - %s", response.Errors[0].Code, response.Errors[0].Message)
	}

	// Check response status
	for _, receipt := range response.Data {
		if receipt.Status != "ok" {
			return fmt.Errorf("notification delivery failed: %s", receipt.Message)
		}
	}

	return nil
}

// BatchSendNotifications sends multiple notifications in batch
func (p *ExpoNotificationProvider) BatchSendNotifications(notifications []NotificationBatch) error {
	if len(notifications) == 0 {
		return nil
	}

	// Get notification service
	notificationService := NewNotificationService(p.db)

	// Process each batch
	for _, batch := range notifications {
		// Prepare notification payload
		payload := notificationService.PrepareNotificationPayload(batch.Notification)

		// Create request body
		requestBody := ExpoNotificationRequest{
			To:    batch.Tokens,
			Title: batch.Notification.Title,
			Body:  batch.Notification.Message,
			Data:  payload,
			Sound: "default",
			TTL:   3600, // 1 hour
		}

		// Set appropriate channel based on notification type
		switch batch.Notification.Type {
		case "TRIP_DEPARTED", "TRIP_ARRIVED", "TRIP_STATUS_CHANGE":
			requestBody.Channel = "trip-updates"
		case "NEW_MESSAGE":
			requestBody.Channel = "messages"
		case "TRIP_DELAYED", "ETA_UPDATED":
			requestBody.Channel = "alerts"
		default:
			requestBody.Channel = "default"
		}

		// Serialize request body
		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("failed to serialize notification request: %w", err)
		}

		// Create HTTP request
		req, err := http.NewRequest("POST", p.ExpoAPIURL, bytes.NewBuffer(jsonBody))
		if err != nil {
			return fmt.Errorf("failed to create HTTP request: %w", err)
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		// Send request
		resp, err := p.HTTPClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send notification request: %w", err)
		}
		defer resp.Body.Close()

		// Parse response
		var response ExpoNotificationResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return fmt.Errorf("failed to parse notification response: %w", err)
		}

		// Check for errors
		if len(response.Errors) > 0 {
			return fmt.Errorf("expo API error: %s - %s", response.Errors[0].Code, response.Errors[0].Message)
		}

		// Log success
		log.Printf("Successfully sent batch notification to %d recipients for notification ID %d",
			len(batch.Tokens), batch.Notification.ID)
	}

	return nil
}
