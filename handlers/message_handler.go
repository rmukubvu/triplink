package handlers

import (
	"github.com/gofiber/fiber/v2"
	"triplink/backend/database"
	"triplink/backend/models"
)

func CreateMessage(c *fiber.Ctx) error {
	var message models.Message

	if err := c.BodyParser(&message); err != nil {
		return err
	}

	database.DB.Create(&message)

	return c.JSON(message)
}

func GetMessages(c *fiber.Ctx) error {
	var messages []models.Message

	database.DB.Find(&messages)

	return c.JSON(messages)
}
