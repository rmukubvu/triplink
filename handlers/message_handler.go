package handlers

import (
	"triplink/backend/database"
	"triplink/backend/models"
	"triplink/backend/services"

	"github.com/gofiber/fiber/v2"
)

// CreateMessage @Summary Create a message
// @Description Create a new message between users
// @Tags messages
// @Accept json
// @Produce json
// @Param message body models.Message true "Message data"
// @Success 201 {object} models.Message
// @Router /messages [post]
func CreateMessage(c *fiber.Ctx) error {
	var message models.Message

	if err := c.BodyParser(&message); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// Create message in database
	result := database.DB.Create(&message)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not create message",
		})
	}

	// Trigger notification for new message
	triggerService := services.NewNotificationTriggerService()
	triggerService.MessageReceivedHandler(message.ID)

	return c.Status(201).JSON(message)
}

// GetMessages @Summary Get all messages
// @Description Get all messages
// @Tags messages
// @Produce json
// @Success 200 {array} models.Message
// @Router /messages [get]
func GetMessages(c *fiber.Ctx) error {
	var messages []models.Message

	database.DB.Find(&messages)

	return c.JSON(messages)
}

// GetUserMessages @Summary Get user messages
// @Description Get all messages for a specific user (sent or received)
// @Tags messages
// @Produce json
// @Param user_id path int true "User ID"
// @Success 200 {array} models.Message
// @Router /users/{user_id}/messages [get]
func GetUserMessages(c *fiber.Ctx) error {
	userID := c.Params("user_id")

	var messages []models.Message
	result := database.DB.Where("sender_id = ? OR receiver_id = ?", userID, userID).
		Order("created_at DESC").
		Find(&messages)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not fetch messages",
		})
	}

	return c.JSON(messages)
}

// GetConversation @Summary Get conversation
// @Description Get messages between two users
// @Tags messages
// @Produce json
// @Param user1_id path int true "First User ID"
// @Param user2_id path int true "Second User ID"
// @Success 200 {array} models.Message
// @Router /messages/conversation/{user1_id}/{user2_id} [get]
func GetConversation(c *fiber.Ctx) error {
	user1ID := c.Params("user1_id")
	user2ID := c.Params("user2_id")

	var messages []models.Message
	result := database.DB.Where(
		"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		user1ID, user2ID, user2ID, user1ID,
	).Order("created_at ASC").Find(&messages)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not fetch conversation",
		})
	}

	return c.JSON(messages)
}
